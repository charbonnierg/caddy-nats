// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"go.uber.org/zap"
)

var (
	DEFAULT_AUTH_CALLOUT_SUBJECT = "$SYS.REQ.USER.AUTH"
)

// AuthCallout is the auth callout handler interface.
// It is used to handle authorization requests.
// Modules in the "nats.auth_callout" namespace must implement this interface.
type AuthCallout interface {
	Provision(app *App) error
	Handle(request *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error)
}

// AuthService is the auth callout service.
// Both an handler and a signing key must be provided.
// The signing key must be the private key for the authorization callout issuer
// or the account in operator mode.
// The credentials are the credentials for the authorization callout issuer when
// using operator mode. Credentials are not used in server mode.
type AuthService struct {
	logger            *zap.Logger
	pk                string
	sk                nkeys.KeyPair
	app               *App
	conn              *nats.Conn
	subject           string
	subscription      *nats.Subscription
	defaultHandler    AuthCallout
	InternalAccount   string             `json:"internal_account,omitempty"`
	InternalUser      string             `json:"internal_user,omitempty"`
	AuthSigningKey    string             `json:"auth_signing_key"`
	SubjectRaw        string             `json:"subject,omitempty"`
	Credentials       string             `json:"credentials,omitempty"`
	Policies          ConnectionPolicies `json:"policies,omitempty"`
	DefaultHandlerRaw json.RawMessage    `json:"handler,omitempty" caddy:"namespace=nats.auth_callout inline_key=module"`
}

// Provision will provision the auth callout service.
// It implements the caddy.Provisioner interface.
// It will load and validate the auth callout handler module.
// It will load and validate the auth signing key.
func (s *AuthService) Provision(app *App) error {
	s.app = app
	s.logger = app.logger.Named("auth_callout")
	// Provision subjec to which auth requests will be sent
	if s.SubjectRaw == "" {
		s.subject = DEFAULT_AUTH_CALLOUT_SUBJECT
	} else {
		s.subject = s.SubjectRaw
	}
	// Generate an NATS server account if needed
	// This account will be used to authenticate the auth callout
	// A single user will be created in this account, password will
	// be the auth signing key.
	if s.AuthSigningKey == "" {
		if err := s.setupInternalAuthAccount(); err != nil {
			return err
		}
	}
	// At this point, either a signing key was provided in configuration
	// or an internal account was created and the signing key is set
	if s.AuthSigningKey == "" {
		return errors.New("internal error: auth signing key is not set but should be")
	}
	// Provision auth signing key
	sk, err := nkeys.FromSeed([]byte(s.AuthSigningKey))
	if err != nil {
		return errors.New("failed to decode auth signing key")
	}
	s.sk = sk
	// Provision auth public key
	pk, err := sk.PublicKey()
	if err != nil {
		return errors.New("failed to get auth signing key public key")
	}
	s.pk = pk
	// Provision default handler
	if s.DefaultHandlerRaw != nil {
		unm, err := app.ctx.LoadModule(s, "DefaultHandlerRaw")
		if err != nil {
			return fmt.Errorf("failed to load default handler: %s", err.Error())
		}
		handler, ok := unm.(AuthCallout)
		if !ok {
			return errors.New("default handler invalid type")
		}
		s.defaultHandler = handler
	}
	// Provision policies
	if err := s.Policies.Provision(app); err != nil {
		return err
	}
	return nil
}

// Start will start the auth callout service.
// It will subscribe to the auth callout subject.
func (s *AuthService) Start(server *server.Server) error {
	s.logger.Info("Starting auth callout service")
	// Get default options
	opts := nats.GetDefaultOptions()
	// Set in process server option
	if err := nats.InProcessServer(server)(&opts); err != nil {
		return err
	}
	if s.Credentials != "" {
		if err := nats.UserCredentials(s.Credentials)(&opts); err != nil {
			return err
		}
	} else {
		// Set password if any
		s.setPassword(&opts)
	}
	// Create connection
	conn, err := opts.Connect()
	if err != nil {
		return err
	}
	s.conn = conn
	// Subscribe to auth callout subject
	sub, err := s.conn.Subscribe(s.subject, s.handleMsg)
	if err != nil {
		return err
	}
	s.subscription = sub
	return nil
}

// Stop will stop the auth callout service.
// It will unsubscribe from the auth callout subject.
func (s *AuthService) Stop() error {
	s.logger.Info("Stopping auth callout service")
	if s.subscription != nil {
		return s.subscription.Unsubscribe()
	}
	return nil
}

// SignUserClaims will sign the user claims using the auth signing key.
func (s *AuthService) SignUserClaims(userClaims *jwt.UserClaims) (string, error) {
	return userClaims.Encode(s.sk)
}

func (s *AuthService) handleMsg(msg *nats.Msg) {
	s.logger.Debug("Received authorization request", zap.ByteString("payload", msg.Data))
	// Decode the request
	request, err := jwt.DecodeAuthorizationRequestClaims(string(msg.Data))
	if err != nil {
		s.handleError(msg, request, err)
		return
	}
	var handler AuthCallout
	// Match handler for this request
	matchedHandler, ok := s.Policies.Match(request)
	// Fail if no policy matched and there is no default handler
	if !ok && s.defaultHandler == nil {
		s.handleError(msg, request, fmt.Errorf("no matching policy"))
		return
	}
	// Use default handler if no policy matched
	if !ok {
		handler = s.defaultHandler
	} else {
		handler = matchedHandler
	}
	// Let handler handle the request
	response, err := handler.Handle(request)
	// Either handle success or failure
	switch err {
	case nil:
		s.handleSuccess(msg, request, response)
		return
	default:
		s.handleError(msg, request, err)
		return
	}
}

func (s *AuthService) handleError(msg *nats.Msg, request *jwt.AuthorizationRequestClaims, err error) {
	response := jwt.NewAuthorizationResponseClaims(request.UserNkey)
	response.Error = err.Error()
	response.Audience = request.Server.ID
	response.Subject = request.UserNkey
	payload, err := response.Encode(s.sk)
	if err != nil {
		s.handleError(msg, request, err)
		return
	}
	if err := msg.Respond([]byte(payload)); err != nil {
		s.logger.Error("failed to respond to authorization request", zap.Error(err))
	}
}

func (s *AuthService) handleSuccess(msg *nats.Msg, request *jwt.AuthorizationRequestClaims, resp *jwt.AuthorizationResponseClaims) {
	resp.Subject = request.UserNkey
	resp.Audience = request.Server.ID
	resp.IssuerAccount = s.pk
	payload, err := resp.Encode(s.sk)
	if err != nil {
		s.logger.Error("failed to encode authorization response", zap.Error(err))
		return
	}
	if err := msg.Respond([]byte(payload)); err != nil {
		s.logger.Error("failed to respond to authorization request", zap.Error(err))
	}
}

func (s *AuthService) setPassword(opts *nats.Options) {
	// The goal is to "guess" the user and password to use for the auth callout
	if s.app.Options != nil && s.app.Options.Authorization != nil {
		auth := s.app.Options.Authorization
		accs := s.app.Options.Accounts
		config := auth.AuthCallout
		if config != nil && config.AuthUsers != nil {
			if auth.Users != nil {
				for _, user := range auth.Users {
					for _, authUser := range config.AuthUsers {
						if user.User == authUser {
							opts.User = user.User
							opts.Password = user.Password
							return
						}
					}
				}
			} else {
				for _, acc := range accs {
					if acc.Name == config.Account {
						for _, user := range acc.Users {
							for _, authUser := range config.AuthUsers {
								if user.User == authUser {
									opts.User = user.User
									opts.Password = user.Password
									return
								}
							}
						}
					}
				}
			}
		}
	}
}
