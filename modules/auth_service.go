package modules

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/charbonnierg/caddy-nats/embedded/natsauth"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type AuthCallout interface {
	Handle(request *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error)
	Provision(app *App) error
}

type AuthService2 struct {
	app               *App
	conn              *nats.Conn
	service           *natsauth.Service
	defaultHandler    AuthCallout
	InternalAccount   string             `json:"internal_account,omitempty"`
	InternalUser      string             `json:"internal_user,omitempty"`
	AuthAccount       string             `json:"auth_account,omitempty"`
	AuthSigningKey    string             `json:"auth_signing_key"`
	SubjectRaw        string             `json:"subject,omitempty"`
	Credentials       string             `json:"credentials,omitempty"`
	Policies          ConnectionPolicies `json:"policies,omitempty"`
	DefaultHandlerRaw json.RawMessage    `json:"handler,omitempty" caddy:"namespace=nats.auth_callout inline_key=module"`
}

func (s *AuthService2) Handle(request *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	var handler AuthCallout
	// Match handler for this request
	matchedHandler, ok := s.Policies.Match(request)
	// Fail if no policy matched and there is no default handler
	if !ok && s.defaultHandler == nil {
		return nil, errors.New("no matching policy")
	}
	// Use default handler if no policy matched
	if !ok {
		handler = s.defaultHandler
	} else {
		handler = matchedHandler
	}
	// Let handler handle the request
	return handler.Handle(request)
}

// Provision will provision the auth callout service.
// It implements the caddy.Provisioner interface.
// It will load and validate the auth callout handler module.
// It will load and validate the auth signing key.
func (s *AuthService2) Provision(app *App) error {
	s.app = app
	// Validate configuration
	if s.AuthSigningKey != "" && s.InternalAccount != "" {
		return errors.New("auth signing key and internal account are mutually exclusive")
	}
	if s.AuthSigningKey == "" && s.InternalAccount == "" {
		s.InternalAccount = natsauth.DEFAULT_AUTH_CALLOUT_ACCOUNT
	}
	// Provision subjec to which auth requests will be sent
	cfg := natsauth.NewConfig(s.Handle)
	cfg.Logger = app.logger.Named("auth_callout")
	if s.SubjectRaw != "" {
		cfg.Subject = s.SubjectRaw
	}
	// Generate an NATS server account if needed
	// This account will be used to authenticate the auth callout
	// A single user will be created in this account, password will
	// be the auth signing key.
	if err := s.setupInternalAuthAccount(); err != nil {
		return err
	}
	// At this point, either a signing key was provided in configuration
	// or an internal account was created and the signing key is set
	if s.AuthSigningKey == "" {
		return errors.New("internal error: auth signing key is not set but should be")
	}
	cfg.SigningKey = s.AuthSigningKey
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
	// Create auth service
	service, err := natsauth.NewService(cfg)
	if err != nil {
		return err
	}
	s.service = service
	return nil
}

func (s *AuthService2) Start(server *server.Server) error {
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
	return s.service.Listen(conn)
}

func (s *AuthService2) Stop() error {
	if s.conn != nil {
		s.conn.Close()
	}
	return nil
}

func (s *AuthService2) setPassword(opts *nats.Options) {
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
