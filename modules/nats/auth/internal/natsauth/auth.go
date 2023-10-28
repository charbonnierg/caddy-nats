// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsauth

import (
	"errors"
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"go.uber.org/zap"
)

var (
	DEFAULT_AUTH_CALLOUT_SUBJECT = "$SYS.REQ.USER.AUTH"
	DEFAULT_AUTH_CALLOUT_ACCOUNT = "AUTH"
)

// Handler is a function that handles auth callout requests
// It must not sign the response claims, but simply return either user claims or an error
// IMPORTANT: The audience of the user claims MUST be the target account
type Handler = func(req *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error)

// Keystore is an interface for a keystore that can be used to retrieve the auth signing key for an account
type Keystore interface {
	Get(account string) (string, error)
}

// Config is the configuration for an auth service.
type Config struct {
	handler    Handler
	Subject    string
	Account    string
	SigningKey string
	Keystore   Keystore
	Logger     *zap.Logger
}

// NewConfig creates a new config with the given handler.
func NewConfig(handler Handler) *Config {
	return &Config{
		handler: handler,
		Subject: DEFAULT_AUTH_CALLOUT_SUBJECT,
		Account: DEFAULT_AUTH_CALLOUT_ACCOUNT,
	}
}

// Service is a service that handles auth callout requests.
// It listens on a NATS subject and responds with a signed JWT.
// It can be used as a standalone service or embedded in a NATS server.
// It accept a handler function that is called when an auth callout request is received.
// The handler function must return either user claims or an error.
// Decoding the request, signing the user claims and signing the auth response is handled by the service.
// The service can be configured with a signing key or a keystore.
type Service struct {
	logger       *zap.Logger
	sk           nkeys.KeyPair
	pk           string
	subscription *nats.Subscription
	Config       *Config
}

// NewService creates a new auth service with the given config.
// It returns an error when the signing key or keystore is not set.
func NewService(config *Config) (*Service, error) {
	srv := &Service{
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

// Handle handles an incoming authorization request as a NATS message.
// It responds with a NATS message with the authorization response.
// If it fails to respond or to handle the message, it logs an error.
func (s *Service) Handle(msg *nats.Msg) {
	reply := s.handle(msg)
	if reply == nil {
		return
	}
	err := msg.RespondMsg(reply)
	if err != nil {
		s.logger.Error("failed to respond to authorization request", zap.Error(err))
	}
}

// Listen subscribes to the auth callout subject and starts the service.
func (s *Service) Listen(conn *nats.Conn) error {
	sub, err := conn.Subscribe(s.Config.Subject, s.Handle)
	if err != nil {
		return err
	}
	s.subscription = sub
	return nil
}

// Close closes the subscription and effectively stops the service.
func (s *Service) Close() error {
	if s.subscription != nil {
		return s.subscription.Unsubscribe()
	}
	return nil
}

// handle handles an incoming authorization request as a NATS message
// and returns a NATS message with the authorization response.
// It returns nil when the request could not be handled.
// The response is signed with the auth account keypair.
func (s *Service) handle(msg *nats.Msg) *nats.Msg {
	// Decode the request
	request, err := jwt.DecodeAuthorizationRequestClaims(string(msg.Data))
	if err != nil {
		s.logger.Error("failed to decode authorization request", zap.Error(err))
		return nil
	}
	// Get a response
	response, err := s.delegate(request)
	if err != nil {
		s.logger.Error("failed to create authorization response", zap.Error(err))
		return nil
	}
	// Sign the response
	payload, err := s.signAuthResponseClaims(response)
	if err != nil {
		s.logger.Error("failed to sign authorization request", zap.Error(err))
		return nil
	}
	return &nats.Msg{
		Subject: msg.Reply,
		Data:    []byte(payload),
	}
}

// delegate calls the handler and returns either authorization response claims or an error
// when an error is returned, the error is sent to the server and displayed to the user as
// additional detail of the Unauthorized error.
func (s *Service) delegate(request *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error) {
	// Let handler return either user claims or an error
	claims, err := s.Config.handler(request)
	if err != nil {
		return s.createErrorResponse(request, err)
	}
	return s.createSuccessResponse(request, claims)
}

// createErrorResponse creates an authorization response given an error
// it exists so that each handle do not need to create an authorization response
// and sign it. It's an abstraction to make the code more readable.
func (s *Service) createErrorResponse(request *jwt.AuthorizationRequestClaims, err error) (*jwt.AuthorizationResponseClaims, error) {
	response := jwt.NewAuthorizationResponseClaims(request.UserNkey)
	// Authorization response audience is the server ID
	response.Audience = request.Server.ID
	// Error is the error message displayed to the user (i think)
	response.Error = err.Error()
	return response, nil
}

// createSuccessResponse creates an authorization response given some user claims
// it exists so that each handle do not need to create an authorization response
// and sign it. It's an abstraction to make the code more readable.
func (s *Service) createSuccessResponse(request *jwt.AuthorizationRequestClaims, claims *jwt.UserClaims) (*jwt.AuthorizationResponseClaims, error) {
	response := jwt.NewAuthorizationResponseClaims(request.UserNkey)
	// Authorization response audience is the server ID
	response.Audience = request.Server.ID
	// User claims audience MUST be the target account
	userToken, err := s.signUserClaims(claims)
	if err != nil {
		return nil, err
	}
	response.Jwt = userToken
	return response, nil
}

// signUserClaims signs the user claims with the target account keypair when a
// keystore is configured. Otherwise, the auth account keypair from config is used.
// This will work in server mode, but not in operator mode when target account is not the auth account.
func (s *Service) signUserClaims(claims *jwt.UserClaims) (string, error) {
	sk, pk, err := s.getKeyPair(claims.Audience)
	if err != nil {
		return "", err
	}
	claims.Issuer = pk
	return claims.Encode(sk)
}

// signAuthResponseClaims signs the auth response claims with the auth account keypair.
// When a keystore is configured, the auth account keypair is fetched from the keystore.
// Otherwise, the auth account keypair from config is used.
func (s *Service) signAuthResponseClaims(claims *jwt.AuthorizationResponseClaims) (string, error) {
	sk, pk, err := s.getAuthAccountKeyPair()
	if err != nil {
		return "", err
	}
	claims.Issuer = pk
	return claims.Encode(sk)
}

// getKeyPair returns the keypair for the target account when a keystore is configured.
// Otherwise, the auth account keypair is returned.
// This will work in server mode, but not in operator mode when target account is not the auth account.
func (s *Service) getKeyPair(account string) (nkeys.KeyPair, string, error) {
	if s.Config.Keystore != nil {
		key, err := s.Config.Keystore.Get(account)
		if err != nil {
			return nil, "", err
		}
		pk, err := nkeys.FromSeed([]byte(key))
		if err != nil {
			return nil, "", err
		}
		return pk, "", nil
	}
	if s.sk != nil {
		return s.sk, s.pk, nil
	}
	return nil, "", fmt.Errorf("unknown account %s", account)
}

// getAuthAccountKeyPair returns the keypair for the account that is used to sign the auth response
func (s *Service) getAuthAccountKeyPair() (nkeys.KeyPair, string, error) {
	return s.getKeyPair(s.Config.Account)
}
