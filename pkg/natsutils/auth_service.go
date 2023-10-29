// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsutils

import (
	"errors"

	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nkeys"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.uber.org/zap"
)

// NewFactory creates a new auth service with the given config.
// It returns an error when the signing key or keystore is not set.
func NewAuthServiceFactory(config *AuthServiceConfig) (*AuthServiceFactory, error) {
	srv := &AuthServiceFactory{
		Config: config,
		logger: config.Logger,
	}
	if config.SigningKey == "" && config.Keystore == nil {
		return nil, errors.New("auth signing key or keystore must be set")
	}
	if config.SigningKey != "" {
		sk, err := nkeys.FromSeed([]byte(config.SigningKey))
		if err != nil {
			return nil, errors.New("failed to decode auth issuer signing key")
		}
		pk, err := sk.PublicKey()
		if err != nil {
			return nil, errors.New("failed to get auth issuer public key")
		}
		srv.sk = sk
		srv.pk = pk
	}
	return srv, nil
}

// Factory is a service that handles auth callout requests.
// It listens on a NATS subject and responds with a signed JWT.
// It can be used as a standalone service or embedded in a NATS server.
// It accept a handler function that is called when an auth callout request is received.
// The handler function must return either user claims or an error.
// Decoding the request, signing the user claims and signing the auth response is handled by the service.
// The service can be configured with a signing key or a keystore.
type AuthServiceFactory struct {
	Config *AuthServiceConfig

	logger *zap.Logger
	sk     nkeys.KeyPair
	pk     string
}

// Handler returns the handler for the auth service.
func (s *AuthServiceFactory) Handler() micro.Handler {
	return &authServiceHandler{
		account:  s.Config.Account,
		sk:       s.sk,
		pk:       s.pk,
		logger:   s.logger,
		handle:   s.Config.handler,
		keystore: s.Config.Keystore,
	}
}

// NewService returns a new nats service definition.
func (s *AuthServiceFactory) NewService() *ServiceDefinition {
	return &ServiceDefinition{
		QueueGroup: s.Config.QueueGroup,
		Endpoints: []*EndpointConfig{
			{
				Name:    fnutils.DefaultIfEmptyString(s.Config.Name, "auth-service"),
				Subject: s.Config.Subject,
				Handler: s.Handler(),
			},
		},
		Version:     fnutils.DefaultIfEmptyString(s.Config.Version, "0.0.1"),
		Description: s.Config.Description,
		Metadata:    s.Config.Metadata,
	}
}
