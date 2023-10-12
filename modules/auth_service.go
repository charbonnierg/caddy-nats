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
	Handle(request *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error)
}

// AuthService is the auth callout service.
// Both an handler and a signing key must be provided.
// The signing key must be the private key for the authorization callout issuer
// or the account in operator mode.
// The credentials are the credentials for the authorization callout issuer when
// using operator mode. Credentials are not used in server mode.
type AuthService struct {
	logger         *zap.Logger
	app            *App
	conn           *nats.Conn
	sub            *nats.Subscription
	sk             nkeys.KeyPair
	handler        AuthCallout
	subject        string
	HandlerRaw     json.RawMessage `json:"handler" caddy:"namespace=nats.auth_callout inline_key=module"`
	AuthSigningKey string          `json:"auth_signing_key"`
	SubjectRaw     string          `json:"subject,omitempty"`
	Credentials    string          `json:"credentials,omitempty"`
}

// Provision will provision the auth callout service.
// It implements the caddy.Provisioner interface.
// It will load and validate the auth callout handler module.
// It will load and validate the auth signing key.
func (s *AuthService) Provision(app *App) error {
	s.logger = app.logger.Named("auth_callout")
	s.app = app
	sk, err := nkeys.FromSeed([]byte(s.AuthSigningKey))
	if err != nil {
		return errors.New("failed to decode auth signing key")
	}
	s.sk = sk
	unm, err := app.ctx.LoadModule(s, "HandlerRaw")
	if err != nil {
		return fmt.Errorf("failed to load auth callout handler: %s", err.Error())
	}
	handler, ok := unm.(AuthCallout)
	if !ok {
		return errors.New("auth callout handler invalid type")
	}
	s.handler = handler

	if s.SubjectRaw == "" {
		s.subject = DEFAULT_AUTH_CALLOUT_SUBJECT
	} else {
		s.subject = s.SubjectRaw
	}
	return nil
}

// Validate will validate the auth callout service.
// It implements the caddy.Validator interface.
func (s *AuthService) Validate() error {
	if s.AuthSigningKey == "" {
		return errors.New("auth signing key is required")
	}
	if s.handler == nil {
		return errors.New("auth callout handler is required")
	}
	return nil
}

// Start will start the auth callout service.
// It will subscribe to the auth callout subject.
func (s *AuthService) Start(server *server.Server) error {
	s.logger.Info("Starting auth callout service")
	opts := nats.GetDefaultOptions()
	// Set in process server option
	if err := nats.InProcessServer(server)(&opts); err != nil {
		return err
	}
	if s.Credentials != "" {
		if err := nats.UserCredentials(s.Credentials)(&opts); err != nil {
			return err
		}
	}
	// TODO: Refactor this big mess
	// The goal is to "guess" the user and password to use for the auth callout
	if s.app.Options != nil && s.app.Options.Authorization != nil {
		auth := s.app.Options.Authorization
		accs := s.app.Options.Accounts
		config := auth.AuthCallout
		if config != nil {
			if config.AuthUsers != nil {
				if auth.Users != nil {
					var found = false
					for _, user := range auth.Users {
						if found {
							break
						}
						for _, authUser := range config.AuthUsers {
							if user.User == authUser {
								opts.User = user.User
								opts.Password = user.Password
								found = true
								break
							}
						}
					}
				} else {
					for _, acc := range accs {
						if acc.Name == config.Account {
							var found = false
							for _, user := range acc.Users {
								if found {
									break
								}
								for _, authUser := range config.AuthUsers {
									if user.User == authUser {
										opts.User = user.User
										opts.Password = user.Password
										found = true
										break
									}
								}
							}
						}
					}
				}
			}
		}
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
	s.sub = sub
	return nil
}

// Stop will stop the auth callout service.
// It will unsubscribe from the auth callout subject.
func (s *AuthService) Stop() error {
	s.logger.Info("Stopping auth callout service")
	if s.sub != nil {
		return s.sub.Unsubscribe()
	}
	return nil
}

func (s *AuthService) handleMsg(msg *nats.Msg) {
	s.logger.Debug("Received authorization request", zap.ByteString("payload", msg.Data))
	// Decode the request
	request, err := jwt.DecodeAuthorizationRequestClaims(string(msg.Data))
	if err != nil {
		s.handleError(msg, request, err)
		return
	}
	// Let module handle the request
	response, err := s.handler.Handle(request)
	// Either handle success or failure
	if err != nil {
		s.handleError(msg, request, err)
	} else {
		s.handleSuccess(msg, request, response)
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
	payload, err := resp.Encode(s.sk)
	if err != nil {
		s.logger.Error("failed to encode authorization response", zap.Error(err))
		return
	}
	if err := msg.Respond([]byte(payload)); err != nil {
		s.logger.Error("failed to respond to authorization request", zap.Error(err))
	}
}
