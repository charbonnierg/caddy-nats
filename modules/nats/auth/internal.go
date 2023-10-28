// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"errors"

	"github.com/nats-io/nkeys"
	runner "github.com/quara-dev/beyond/modules/nats/embedded"
)

// setupInternalAuthAccount sets up the internal auth account in embedded server options.
func (s *AuthService) setupInternalAuthAccount() error {
	if s.AuthSigningKey != "" {
		return nil
	}

	if s.InternalUser != "" && s.InternalAccount == "" {
		return errors.New("internal account is required when using internal user")
	}
	opts := s.app.Options()
	if s.InternalAccount != "" && opts.Authorization != nil {
		return errors.New("internal account is not allowed when custom authorization map is used")
	}
	if s.InternalAccount != "" && opts.Accounts == nil {
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
		auth := runner.AuthorizationMap{
			AuthCallout: &runner.AuthCalloutMap{
				Issuer:    pk,
				Account:   s.InternalAccount,
				AuthUsers: []string{pk},
			},
		}
		user := runner.User{
			User: pk, Password: string(seed),
		}
		acc := runner.Account{
			Name: s.InternalAccount, Users: []runner.User{user},
		}
		s.AuthSigningKey = string(seed)
		opts.Authorization = &auth
		opts.Accounts = append(opts.Accounts, &acc)
	}
	return nil
}
