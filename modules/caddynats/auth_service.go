package caddynats

import (
	"context"
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nkeys"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.uber.org/zap"
)

type authService struct {
	ctx        context.Context
	logger     *zap.Logger
	queueGroup string

	account    string
	issuer     string
	signingKey nkeys.KeyPair
	keystore   KeyStore
	policies   natsauth.AuthorizationPolicies
}

func (a *authService) Definition() (*natsclient.ServiceDefinition, error) {
	var group = a.queueGroup
	if group == "" {
		group = "auth_callout"
	}
	// Create a service definition
	definition := natsclient.ServiceDefinition{
		QueueGroup: a.queueGroup,
		Endpoints: []*natsclient.EndpointDefinition{
			{
				Name:    "auth-service",
				Subject: "$SYS.REQ.USER.AUTH",
				Handler: a,
			},
		},
		Version: "0.0.1",
	}
	return &definition, nil
}

// Handle handles an incoming authorization request as a NATS message.
// It responds with a NATS message with the authorization response.
// If it fails to respond or to handle the message, it logs an error.
func (s *authService) Handle(req micro.Request) {
	// Decode the request
	request, err := jwt.DecodeAuthorizationRequestClaims(string(req.Data()))
	if err != nil {
		s.logger.Error("failed to decode authorization request", zap.Error(err))
		return
	}
	account, claims, ok := __pinnedtokens__.Lookup(request.ConnectOptions.Token)
	var response *jwt.AuthorizationResponseClaims
	var authError error
	if request.ConnectOptions.Token != "" && ok {
		// Check if the request has a pinned token
		s.logger.Info("using pinned token", zap.Any("client_infos", request.ClientInformation))
		claims.Subject = request.UserNkey
		claims.Audience = account
		response, authError = s.createSuccessResponse(request, claims)
	} else {
		// Or delegate to authorization policies
		response, authError = s.delegate(request)
	}
	// Get a response
	if authError != nil {
		s.logger.Error("failed to create authorization response", zap.Error(authError))
		return
	}
	// Sign the response
	payload, encodeError := s.signAuthResponseClaims(response)
	if encodeError != nil {
		s.logger.Error("failed to sign authorization request", zap.Error(encodeError))
		return
	}
	respondError := req.Respond([]byte(payload))
	if respondError != nil {
		s.logger.Error("failed to respond to authorization request", zap.Error(respondError))
	}
}

// delegate calls the handler and returns either authorization response claims or an error
// when an error is returned, the error is sent to the server and displayed to the user as
// additional detail of the Unauthorized error.
func (s *authService) delegate(claims *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error) {
	s.logger.Info("auth callout request", zap.Any("client_infos", claims.ClientInformation), zap.Any("connect_opts", claims.ConnectOptions))
	// Let handler return either user claims or an error
	userClaims, err := s.policies.MatchAndAuthorize(s.ctx, claims)
	if err != nil {
		return s.createErrorResponse(claims, err)
	}
	return s.createSuccessResponse(claims, userClaims)
}

// createErrorResponse creates an authorization response given an error
// it exists so that each handle do not need to create an authorization response
// and sign it. It's an abstraction to make the code more readable.
func (s *authService) createErrorResponse(request *jwt.AuthorizationRequestClaims, err error) (*jwt.AuthorizationResponseClaims, error) {
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
func (s *authService) createSuccessResponse(request *jwt.AuthorizationRequestClaims, claims *jwt.UserClaims) (*jwt.AuthorizationResponseClaims, error) {
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
func (s *authService) signUserClaims(claims *jwt.UserClaims) (string, error) {
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
func (s *authService) signAuthResponseClaims(claims *jwt.AuthorizationResponseClaims) (string, error) {
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
func (s *authService) getKeyPair(account string) (nkeys.KeyPair, string, error) {
	if s.keystore != nil {
		key, err := s.keystore.GetKey(account)
		if err != nil {
			return nil, "", err
		}
		pk, err := nkeys.FromSeed([]byte(key))
		if err != nil {
			return nil, "", err
		}
		return pk, "", nil
	}
	if s.signingKey != nil {
		return s.signingKey, s.issuer, nil
	}
	return nil, "", fmt.Errorf("unknown account %s", account)
}

// getAuthAccountKeyPair returns the keypair for the account that is used to sign the auth response
func (s *authService) getAuthAccountKeyPair() (nkeys.KeyPair, string, error) {
	return s.getKeyPair(s.issuer)
}
