// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	_nats "github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/modules/nats/auth/internal/natsauth"
	"github.com/quara-dev/beyond/pkg/natsutils"
	"go.uber.org/zap"
)

// AuthService is the auth callout service based on policies.
type AuthService struct {
	AuthServiceConfig
	logger         *zap.Logger
	app            nats.App
	ctx            caddy.Context
	service        *natsauth.Service
	defaultHandler nats.AuthCallout
	client         natsutils.Client
}

// handle is the handler for auth requests.
func (s *AuthService) handle(claims *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
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
	if s.ClientRaw != nil {
		client := *s.ClientRaw
		s.client = client
	}
	if s.AuthSigningKey != "" && s.InternalAccount != "" {
		return errors.New("auth signing key and internal account are mutually exclusive")
	}
	if s.AuthSigningKey == "" && s.InternalAccount == "" {
		s.InternalAccount = natsauth.DEFAULT_AUTH_CALLOUT_ACCOUNT
	}
	// Provision subjec to which auth requests will be sent
	cfg := natsauth.NewConfig(s.handle)
	cfg.Logger = s.logger.Named("auth_callout")
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
	// Create auth service
	service, err := natsauth.NewService(cfg)
	if err != nil {
		return err
	}
	s.service = service
	return nil
}

// Client returns the NATS client used by the auth service.
func (s *AuthService) Client() *natsutils.Client {
	return &s.client
}

// Listen will start the auth service.
func (s *AuthService) Listen(conn *_nats.Conn) error {
	return s.service.Listen(conn)
}

// Close will close the auth service.
func (s *AuthService) Close() error {
	return s.service.Close()
}

var (
	_ nats.AuthService = (*AuthService)(nil)
)
