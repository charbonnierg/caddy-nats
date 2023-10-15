// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"errors"

	"github.com/charbonnierg/caddy-nats/embedded/natsoptions"
	"github.com/nats-io/nkeys"
	"go.uber.org/zap"
)

func (s *AuthService) setupInternalAuthAccount() error {
	if s.AuthSigningKey != "" && s.InternalAccount != "" {
		return errors.New("auth signing key and internal account are mutually exclusive")
	}
	if s.InternalUser != "" && s.InternalAccount == "" {
		return errors.New("internal account is required when using internal user")
	}
	if s.InternalAccount != "" && s.app.Options.Authorization != nil {
		return errors.New("internal account is not allowed when custom authorization map is used")
	}
	if s.InternalAccount != "" && s.app.Options.Accounts == nil {
		return errors.New("internal account is not allowed when no accounts are defined")
	}
	if s.InternalAccount != "" {
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
		if s.InternalUser == "" {
			s.InternalUser = pk
		}
		auth := natsoptions.AuthorizationMap{
			AuthCallout: &natsoptions.AuthCalloutMap{
				Issuer:    pk,
				Account:   s.InternalAccount,
				AuthUsers: []string{pk},
			},
		}
		user := natsoptions.User{
			User: pk, Password: string(seed),
		}
		acc := natsoptions.Account{
			Name: s.InternalAccount, Users: []natsoptions.User{user},
		}
		s.AuthSigningKey = string(seed)
		s.app.Options.Authorization = &auth
		s.app.Options.Accounts = append(s.app.Options.Accounts, &acc)
		s.logger.Info("internal auth account created", zap.String("account", s.InternalAccount), zap.String("user", s.InternalUser))
	}
	return nil
}
