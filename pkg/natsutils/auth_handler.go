// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsutils

import (
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nkeys"
	"go.uber.org/zap"
)

// handlerFunc is a function that handles auth callout requests
// It must not sign the response claims, but simply return either user claims or an error
// IMPORTANT: The audience of the user claims MUST be the target account
type authServiceHandlerFunc = func(req *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error)

type authServiceHandler struct {
	account  string
	sk       nkeys.KeyPair
	pk       string
	logger   *zap.Logger
	handle   authServiceHandlerFunc
	keystore Keystore
}

// Handle handles an incoming authorization request as a NATS message.
// It responds with a NATS message with the authorization response.
// If it fails to respond or to handle the message, it logs an error.
func (s *authServiceHandler) Handle(req micro.Request) {
	// Decode the request
	request, err := jwt.DecodeAuthorizationRequestClaims(string(req.Data()))
	if err != nil {
		s.logger.Error("failed to decode authorization request", zap.Error(err))
		return
	}
	// Get a response
	response, err := s.delegate(request)
	if err != nil {
		s.logger.Error("failed to create authorization response", zap.Error(err))
		return
	}
	// Sign the response
	payload, err := s.signAuthResponseClaims(response)
	if err != nil {
		s.logger.Error("failed to sign authorization request", zap.Error(err))
		return
	}
	err = req.Respond([]byte(payload))
	if err != nil {
		s.logger.Error("failed to respond to authorization request", zap.Error(err))
	}
}

// delegate calls the handler and returns either authorization response claims or an error
// when an error is returned, the error is sent to the server and displayed to the user as
// additional detail of the Unauthorized error.
func (s *authServiceHandler) delegate(request *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error) {
	// Let handler return either user claims or an error
	claims, err := s.handle(request)
	if err != nil {
		return s.createErrorResponse(request, err)
	}
	return s.createSuccessResponse(request, claims)
}

// createErrorResponse creates an authorization response given an error
// it exists so that each handle do not need to create an authorization response
// and sign it. It's an abstraction to make the code more readable.
func (s *authServiceHandler) createErrorResponse(request *jwt.AuthorizationRequestClaims, err error) (*jwt.AuthorizationResponseClaims, error) {
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
func (s *authServiceHandler) createSuccessResponse(request *jwt.AuthorizationRequestClaims, claims *jwt.UserClaims) (*jwt.AuthorizationResponseClaims, error) {
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
func (s *authServiceHandler) signUserClaims(claims *jwt.UserClaims) (string, error) {
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
func (s *authServiceHandler) signAuthResponseClaims(claims *jwt.AuthorizationResponseClaims) (string, error) {
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
func (s *authServiceHandler) getKeyPair(account string) (nkeys.KeyPair, string, error) {
	if s.keystore != nil {
		key, err := s.keystore.Get(account)
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
func (s *authServiceHandler) getAuthAccountKeyPair() (nkeys.KeyPair, string, error) {
	return s.getKeyPair(s.account)
}
