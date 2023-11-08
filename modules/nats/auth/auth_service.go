// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nkeys"
	"github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/modules/nats/auth/policies"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/fnutils"
	runner "github.com/quara-dev/beyond/pkg/natsutils/embedded"
	"go.uber.org/zap"
)

var (
	DEFAULT_AUTH_CALLOUT_SUBJECT = "$SYS.REQ.USER.AUTH"
	DEFAULT_AUTH_CALLOUT_ACCOUNT = "AUTH"
)

// AuthService is the auth callout service based on policies.
type AuthService struct {
	Name              string                      `json:"name,omitempty"`
	Connection        *client.Connection          `json:"connection,omitempty"`
	AuthPublicKey     string                      `json:"auth_public_key,omitempty"`
	AuthSigningKey    string                      `json:"auth_signing_key,omitempty"`
	SigningKeyStore   json.RawMessage             `json:"signing_key_store,omitempty" caddy:"namespace=secrets.stores inline_key=module"`
	Subject           string                      `json:"subject,omitempty"`
	Policies          policies.ConnectionPolicies `json:"policies,omitempty"`
	DefaultHandlerRaw json.RawMessage             `json:"handler,omitempty" caddy:"namespace=nats.auth_callout inline_key=module"`
	QueueGroup        string                      `json:"queue_group,omitempty"`
	Description       string                      `json:"description,omitempty"`
	Version           string                      `json:"version,omitempty"`
	Metadata          map[string]string           `json:"metadata,omitempty"`

	logger         *zap.Logger
	app            nats.App
	ctx            caddy.Context
	defaultHandler nats.AuthCallout
	keystore       nats.Keystore
	service        micro.Service
}

// Handle is the auth callout handler
func (s *AuthService) Handle(claims *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	s.logger.Info("auth callout request", zap.Any("client_infos", claims.ClientInformation))
	req := &AuthorizationRequest{
		claims:  claims,
		context: context.TODO(),
	}
	var handler nats.AuthCallout
	// Match handler for this request
	matchedHandler, ok := s.Policies.Match(claims)
	if !ok {
		s.logger.Info("using default handler", zap.Any("client_infos", claims.ClientInformation))
		handler = s.defaultHandler
	} else {
		s.logger.Info("handler policy matcher", zap.String("handler", string(matchedHandler.HandlerRaw)), zap.Any("client_infos", claims.ClientInformation))
		handler = matchedHandler
	}
	if handler == nil {
		s.logger.Info("no matching policy", zap.Any("client_infos", claims.ClientInformation))
		return nil, errors.New("no handler")
	}
	// Let handler handle the request
	return handler.Handle(req)
}

// Provision will provision the auth callout service.
// It implements the caddy.Provisioner interface.
// It will load and validate the auth callout handler module.
// It will load and validate the auth signing key.
func (s *AuthService) Provision(app nats.App) error {
	s.app = app
	s.ctx = app.Context()
	s.logger = app.Logger().Named("auth_service")
	// Validate configuration
	if s.Connection == nil {
		s.Connection = &client.Connection{}
	}
	if s.Connection.Name == "" {
		s.Connection.Name = "auth_callout"
	}
	if s.Connection.Account == "" {
		s.Connection.Account = DEFAULT_AUTH_CALLOUT_ACCOUNT
	}
	if s.AuthPublicKey != "" && s.AuthSigningKey != "" {
		return errors.New("auth signing key and auth public key are mutually exclusive")
	}
	if s.AuthPublicKey != "" && s.SigningKeyStore == nil {
		return errors.New("auth public key is set but no keystore is defined")
	}
	if err := s.Connection.Provision(app); err != nil {
		return err
	}
	// Generate an NATS server account if needed
	// This account will be used to authenticate the auth callout
	// A single user will be created in this account, password will
	// be the auth signing key.
	if err := s.setupInternalAuthAccount(); err != nil {
		return err
	}
	if s.AuthSigningKey == "" && s.SigningKeyStore == nil {
		return errors.New("auth signing key or keystore must be set")
	}
	// Provision default handler
	if s.DefaultHandlerRaw != nil {
		unm, err := s.ctx.LoadModule(s, "DefaultHandlerRaw")
		if err != nil {
			return fmt.Errorf("failed to load default handler: %s", err.Error())
		}
		handler, ok := unm.(nats.AuthCallout)
		if !ok {
			return errors.New("default handler invalid type")
		}
		if err := handler.Provision(app); err != nil {
			return fmt.Errorf("failed to provision default handler: %s", err.Error())
		}
		s.defaultHandler = handler
	}
	// Provision policies
	if err := s.Policies.Provision(app); err != nil {
		return err
	}
	// s.cfg = cfg
	return nil
}

// Connect
func (s *AuthService) Connect() error {
	var pk = s.AuthPublicKey
	var sk nkeys.KeyPair
	if s.AuthSigningKey != "" {
		signingKey, err := nkeys.FromSeed([]byte(s.AuthSigningKey))
		if err != nil {
			return errors.New("failed to decode auth issuer signing key")
		}
		publicKey, err := signingKey.PublicKey()
		if err != nil {
			return errors.New("failed to get auth issuer public key")
		}
		sk = signingKey
		pk = publicKey
	}
	subject := fnutils.DefaultIfEmptyString(s.Subject, DEFAULT_AUTH_CALLOUT_SUBJECT)
	s.logger.Info("auth callout service connected", zap.String("subject", subject))
	u, p, err := s.getAuthUserPass()
	if err == nil {
		s.Connection.Username = u
		s.Connection.Password = p
	}
	// Create service
	handler := &authServiceHandler{
		signingKey: sk,
		publicKey:  pk,
		logger:     s.logger,
		handle:     s.Handle,
		keystore:   s.keystore,
	}
	definition := &client.ServiceDefinition{
		QueueGroup: s.QueueGroup,
		Endpoints: []*client.EndpointDefinition{
			{
				Name:    fnutils.DefaultIfEmptyString(s.Name, "auth-service"),
				Subject: subject,
				Handler: handler,
			},
		},
		Version:     fnutils.DefaultIfEmptyString(s.Version, "0.0.1"),
		Description: s.Description,
		Metadata:    s.Metadata,
	}
	nc, err := s.Connection.Conn()
	if err != nil {
		return err
	}
	service, err := definition.Start(nc)
	if err != nil {
		return err
	}
	s.service = service
	return nil
}

func (s *AuthService) Close() error {
	if s.service != nil {
		return s.service.Stop()
	}
	return nil
}

func (a *AuthService) Zero() bool {
	if a == nil {
		return true
	}
	return a.Connection == nil &&
		a.AuthSigningKey == "" &&
		a.Subject == "" &&
		a.Policies == nil &&
		a.DefaultHandlerRaw == nil
}

// setupInternalAuthAccount sets up the internal auth account in embedded server options.
func (s *AuthService) setupInternalAuthAccount() error {
	if s.AuthSigningKey != "" {
		return nil
	}
	opts := s.app.GetOptions()
	if s.Connection.Account != "" && opts.Authorization != nil {
		return errors.New("internal account is not allowed when custom authorization map is used")
	}
	if s.Connection.Account != "" && opts.Accounts == nil {
		return errors.New("internal account is not allowed when no accounts are defined")
	}
	if s.Connection.Account != "" {
		sk, err := nkeys.CreateAccount()
		if err != nil {
			return errors.New("failed to create internal auth account")
		}
		seed, err := sk.Seed()
		if err != nil {
			return errors.New("failed to get internal auth account seed")
		}
		pk, err := sk.PublicKey()
		if err != nil {
			return errors.New("failed to get internal auth account public key")
		}
		auth := runner.AuthorizationMap{
			AuthCallout: &runner.AuthCalloutMap{
				Issuer:    pk,
				Account:   s.Connection.Account,
				AuthUsers: []string{pk},
			},
		}
		user := runner.User{
			User: pk, Password: string(seed),
		}
		acc := runner.Account{
			Name: s.Connection.Account, Users: []runner.User{user},
		}
		s.AuthSigningKey = string(seed)
		opts.Authorization = &auth
		opts.Accounts = append(opts.Accounts, &acc)
	}
	return nil
}

// getAuthUserPass will return the user and password to use for the auth service
// according to server configuration.
func (a *AuthService) getAuthUserPass() (string, string, error) {
	opts := a.app.GetOptions()
	// The goal is to "guess" the user and password to use for the auth callout
	if opts.Authorization != nil {
		auth := opts.Authorization
		accs := opts.Accounts
		config := auth.AuthCallout
		if config != nil && config.AuthUsers != nil {
			if auth.Users != nil {
				for _, user := range auth.Users {
					for _, authUser := range config.AuthUsers {
						if user.User == authUser {
							return user.User, user.Password, nil
						}
					}
				}
			} else {
				for _, acc := range accs {
					if acc.Name == config.Account {
						for _, user := range acc.Users {
							for _, authUser := range config.AuthUsers {
								if user.User == authUser {
									return user.User, user.Password, nil
								}
							}
						}
					}
				}
			}
		}
	}
	return "", "", fmt.Errorf("user not found")
}
