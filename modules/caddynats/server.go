// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package caddynats

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nkeys"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/modules/caddynats/natsmetrics"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.uber.org/zap"
)

// Server is the main struct wrapping a NATS server as well
// as NATS client connections.
type Server struct {
	ctx                caddy.Context
	srv                *server.Server
	collector          *natsmetrics.Collector
	authClient         *natsclient.NatsClient
	clients            []*natsclient.NatsClient
	currentConfigIndex int
	dontStart          bool

	*Options
	Disabled bool `json:"disabled,omitempty"`
}

// ProvisionClientConnection adds this server as the InProcessConnProvider for the client connection.
// The connection is NOT started automatically, it is up to the caller to start it
// using the Connect() method. The connection is NOT managed either, it is up to the caller
// to close it using the Close() method.
// A client should not be updated once provisioned.
func (s *Server) ProvisionClientConnection(account string, client *natsclient.NatsClient) error {
	if client == nil {
		return errors.New("client cannot be nil")
	}
	if account != "" {
		return s.provisionInternalClientConnection(account, client)
	}
	if err := client.Provision(s.ctx, s); err != nil {
		return err
	}
	return nil
}

// ProvisionInternalClientConnection adds this server as the InProcessConnProvider for the client connection
// and automatically configure credentials for the client using a pinned token or credentials found in server
// config.
// The connection is NOT started automatically, it is up to the caller to start it
// using the Connect() method. The connection is NOT managed either, it is up to the caller
// to close it using the Close() method.
// A client should not be updated once provisioned.
func (s *Server) provisionInternalClientConnection(account string, client *natsclient.NatsClient) error {
	acc, ok := s.GetAccount(account)
	if !ok {
		return fmt.Errorf("account not found: %s", account)
	}
	if _, err := s.createInternalClientForAccount(acc, client); err != nil {
		return err
	}
	return nil
}

// GetAccount returns the account with the given name.
// An account cannot be updated once the server is started without a config reload.
func (s *Server) GetAccount(name string) (*Account, bool) {
	for _, acc := range s.Accounts {
		if acc.Name == name {
			return acc, true
		}
	}
	return nil, false
}

// GetServer returns the underlying NATS server.
// An error is returned if the server is not created yet or if it is not ready for connections.
// A server cannot be updated once started.
func (s *Server) GetServer() (*server.Server, error) {
	if s.srv == nil {
		return nil, errors.New("server is not created yet")
	}
	if !s.srv.ReadyForConnections(5 * time.Second) {
		return nil, errors.New("server is not ready for connections")
	}
	return s.srv, nil
}

// Provision configures the NATS server without starting it.
// When application is reloaded, a new module is provisioned BEFORE the old one is stopped.
func (s *Server) Provision(ctx caddy.Context) error {
	// Increment last counter
	s.incrementConfigIndex()
	// Save context
	ctx, cancel := caddy.NewContext(ctx)
	s.ctx = ctx
	go func() {
		<-ctx.Done()
		cancel()
	}()
	// Initialize properties which are not valid when empty
	s.clients = []*natsclient.NatsClient{}
	// Parse desired server state
	if s.Disabled {
		s.dontStart = true
		return nil
	}
	// Provision options
	err := s.Options.Provision(s.ctx)
	if err != nil {
		return err
	}
	// this is a hack
	for _, acc := range s.Options.Accounts {
		for _, flow := range acc.Flows {
			if err := flow.Provision(s, acc); err != nil {
				return err
			}
		}
	}
	return nil
}

// Start starts the NATS server.
// When application is reloaded, a new module is started BEFORE the old one is stopped.
func (s *Server) Start() error {
	if len(s.automateSubjects) > 0 {
		if s.tlsApp == nil {
			return errors.New("cannot automate subjects without tls app")
		}
		if err := s.tlsApp.Manage(s.automateSubjects); err != nil {
			return fmt.Errorf("failed to automate tls certificate subjects: %v", err)
		}
	}
	switch {
	// Make sure server and collector are stopped if we should not start
	case s.dontStart:
		stopRunningProcesses()
		return nil
	// Make sure options exist
	case s.serverOpts == nil:
		return errors.New("server options are not provisioned")
	}
	// Make sure auth account and auth user exist
	if s.serverOpts.AuthCallout != nil && s.serverOpts.TrustedOperators == nil {
		if err := s.createClientForAuthService(); err != nil {
			return err
		}
	}
	if serverIsAlreadyRunning() {
		if err := s.reuseExistingServer(); err != nil {
			return err
		}
	} else {
		// Create a new server
		if err := s.createNewServer(); err != nil {
			return err
		}
	}
	// Wait for the server to be ready for connections
	if ok := s.srv.ReadyForConnections(5 * time.Second); !ok {
		s.srv.Shutdown()
		return errors.New("server not ready for connections before timeout")
	}
	// Setup the auth service
	if err := s.setupAuthService(); err != nil {
		return err
	}
	// Setup the accounts
	if s.Options.Accounts != nil {
		if err := s.setupAccounts(); err != nil {
			return err
		}
	}
	s.logger.Info("server is ready for connections")
	// Exit early if there is no need to start metric collector
	if s.collectorOpts == nil {
		return nil
	}
	if collectorIsAlreadyRunning() {
		if err := s.createNewCollector(); err != nil {
			return err
		}
	} else {
		if err := s.reuseExistingCollector(); err != nil {
			return err
		}
	}
	s.logger.Info("metrics collector is running")
	return nil
}

// Stop stops the NATS server.
// When application is reloaded, old module is stopped AFTER the new one is started.
func (s *Server) Stop() error {
	s.logger.Warn("preparing to shutdown server")
	// Do nothing if there is no server
	if s.srv == nil {
		return nil
	}
	// Stop data flows
	for _, acc := range s.Options.Accounts {
		for _, flow := range acc.Flows {
			s.logger.Warn("stopping data flow", zap.String("source", flow.source.CaddyModule().String()), zap.String("destination", flow.destination.CaddyModule().String()), zap.String("account", acc.Name))
			if err := flow.Stop(); err != nil {
				s.logger.Warn("failed to stop data flow", zap.String("source", flow.source.CaddyModule().String()), zap.String("destination", flow.destination.CaddyModule().String()), zap.String("account", acc.Name), zap.Error(err))
			}
		}
	}
	// Stop client connections
	for _, client := range s.clients {
		s.logger.Warn("stopping client", zap.String("name", client.Name))
		if err := client.Close(); err != nil {
			s.logger.Warn("failed to stop client", zap.Error(err))
		}
	}
	// Stop the auth service if it exists
	if s.authClient != nil {
		s.logger.Warn("stopping auth service")
		if err := s.authClient.Close(); err != nil {
			s.logger.Warn("failed to stop auth service", zap.Error(err))
		}
	}
	// Don't stop resources if we are reloading
	if s.currentConfigIndex != __last__ {
		return nil
	}
	// Stop the collector if it exists
	if s.collector != nil {
		s.logger.Warn("stopping metrics collector")
		s.collector.Stop()
	}
	s.logger.Warn("stopping server")
	// Shutdown the server
	if s.srv.Running() {
		s.srv.Shutdown()
	}
	return nil
}

// helper method used to keep track of configuration reloads
func (s *Server) incrementConfigIndex() {
	__last__++
	s.currentConfigIndex = __last__
}

// helper method used to create a new NATS server
func (s *Server) createNewServer() error {
	s.logger.Info("creating nats server", zap.String("name", s.serverOpts.ServerName), zap.Int("port", s.serverOpts.Port))
	srv, err := s.Options.CreateServer()
	if err != nil {
		return err
	}
	s.logger.Info("starting nats server", zap.String("name", s.serverOpts.ServerName), zap.Int("port", s.serverOpts.Port))
	srv.Start()
	__server__ = srv
	s.srv = srv
	return nil
}

// helper method used to reload an existing NATS server
func (s *Server) reuseExistingServer() error {
	s.srv = __server__
	// TODO: We should reload only if there is a change in the config
	// and the ReloadOptions from nats is not really usable. For example, it completely overrides
	// logging if logging level changes (which is not what we want, we want to keep the logging format).
	// We should detect changes in the config and apply changes accordingly
	s.logger.Info("reloading nats server", zap.String("name", s.serverOpts.ServerName), zap.Int("port", s.serverOpts.Port))
	if err := s.srv.ReloadOptions(s.serverOpts); err != nil {
		if !strings.Contains(err.Error(), "config reload not supported for") {
			s.logger.Error("failed to reload nats server", zap.Error(err))
			return err
		}
		s.logger.Warn(err.Error())
		s.logger.Info("restarting nats server", zap.String("name", s.serverOpts.ServerName), zap.Int("port", s.serverOpts.Port))
		stopRunningProcesses()
		if err := s.createNewServer(); err != nil {
			return err
		}
	}
	s.logger.Info("reloading nats auth service policies", zap.String("name", s.serverOpts.ServerName), zap.Int("port", s.serverOpts.Port))
	s.authService.policies = natsauth.AuthorizationPolicies{}
	for _, acc := range s.Options.Accounts {
		for _, pol := range acc.AuthorizationPolicies {
			s.authService.policies = append(s.authService.policies, pol)
		}
	}
	return nil
}

// helper method used to create a new metrics collector
func (s *Server) createNewCollector() error {
	// Create the metrics collector
	collector, err := s.Options.CreateCollector()
	if err != nil {
		return err
	}
	s.logger.Info("starting metrics collector")
	if err := collector.Start(); err != nil {
		return err
	}
	// Save reference to the collector
	__collector__ = collector
	s.collector = collector
	return nil
}

// helper method used to reload an existing metrics collector
func (s *Server) reuseExistingCollector() error {
	s.collector = __collector__
	return nil
}

// Helper method to setup auth service
func (s *Server) setupAuthService() error {
	if s.authService == nil {
		return nil
	}
	s.authService.policies = natsauth.AuthorizationPolicies{}
	for _, acc := range s.Options.Accounts {
		for _, pol := range acc.AuthorizationPolicies {
			s.authService.policies = append(s.authService.policies, pol)
		}
	}
	s.logger.Info("starting auth service", zap.String("account", s.authService.account))
	err := s.authClient.Connect()
	if err != nil {
		return err
	}
	_, err = s.authClient.ConfigureService(s.authService)
	if err != nil {
		return err
	}
	return nil
}

// Helper method to setup server accounts.
func (s *Server) setupAccounts() error {
	if len(s.Accounts) > 0 {
		s.logger.Info("setting up nats accounts")
	}
	// Lookup and enable jetstream for accounts
	for _, acc := range s.Options.Accounts {
		account, err := s.srv.LookupAccount(acc.Name)
		if err != nil {
			return fmt.Errorf("account was not initialized: %s", err.Error())
		}
		if acc.JetStream {
			// Enable jetstream
			if !account.JetStreamEnabled() {
				s.logger.Info("Enabling jetstream for account", zap.String("name", acc.Name))
				err = account.EnableJetStream(nil)
				if err != nil {
					return fmt.Errorf("failed to enabled jetstream for account: %s", err.Error())
				}
			}
		}
	}
	// Look once again to add exports
	for _, acc := range s.Options.Accounts {
		account, err := s.srv.LookupAccount(acc.Name)
		if err != nil {
			return fmt.Errorf("account was not initialized: %s", err.Error())
		}
		// Enable exports
		if acc.Exports != nil {
			for _, export := range acc.Exports.Services {
				var targets []*server.Account
				if export.To != nil {
					for _, target := range export.To {
						targetAccount, err := s.srv.LookupAccount(target)
						if err != nil {
							return fmt.Errorf("account was not initialized: %s", err.Error())
						}
						targets = append(targets, targetAccount)
					}
				}
				err := account.AddServiceExportWithResponse(export.Subject, server.Singleton, targets)
				if err != nil {
					return fmt.Errorf("failed to add service export: %s", err.Error())
				}
				s.logger.Info("Added service export", zap.String("subject", export.Subject), zap.Strings("targets", export.To))
			}
			for _, export := range acc.Exports.Streams {
				var targets []*server.Account
				if export.To != nil {
					for _, target := range export.To {
						targetAccount, err := s.srv.LookupAccount(target)
						if err != nil {
							return fmt.Errorf("account was not initialized: %s", err.Error())
						}
						targets = append(targets, targetAccount)
					}
				}
				err := account.AddStreamExport(export.Subject, targets)
				if err != nil {
					return fmt.Errorf("failed to add stream export: %s", err.Error())
				} else {
					s.logger.Info("Added stream export", zap.String("subject", export.Subject), zap.Strings("targets", export.To))
				}
			}
		}
	}
	// Look once again to add imports
	for _, acc := range s.Options.Accounts {
		account, err := s.srv.LookupAccount(acc.Name)
		if err != nil {
			return fmt.Errorf("account was not initialized: %s", err.Error())
		}
		if acc.Imports != nil {
			for _, import_ := range acc.Imports.Services {
				remoteAcc, err := s.srv.LookupAccount(import_.Account)
				if err != nil {
					return fmt.Errorf("account was not initialized: %s", err.Error())
				}
				to := fnutils.DefaultIfEmptyString(import_.To, import_.Subject)
				err = account.AddServiceImport(remoteAcc, import_.Subject, to)
				if err != nil {
					return fmt.Errorf("failed to add service import from %s on %s: %s", import_.Account, import_.Subject, err.Error())
				} else {
					s.logger.Info("Added service import", zap.String("subject", import_.Subject), zap.String("from", import_.Account), zap.String("to", to))
				}
			}
			for _, import_ := range acc.Imports.Streams {
				remoteAcc, err := s.srv.LookupAccount(import_.Account)
				if err != nil {
					return fmt.Errorf("account was not initialized: %s", err.Error())
				}
				err = account.AddStreamImport(remoteAcc, import_.Subject, import_.To)
				if err != nil {
					return fmt.Errorf("failed to add stream import: %s", err.Error())
				}
			}
		}
		if acc.Mappings != nil {
			for _, mapping := range acc.Mappings {
				err := account.AddWeightedMappings(mapping.Subject, mapping.MapDest...)
				if err != nil {
					return fmt.Errorf("failed to add weighted mapping: %s", err.Error())
				}
			}
		}
	}
	// Look once again and make sure all resources exist
	go func() {
		for _, acc := range s.Options.Accounts {
			if len(acc.Streams) == 0 &&
				len(acc.Consumers) == 0 &&
				len(acc.KeyValueStores) == 0 &&
				len(acc.ObjectStores) == 0 &&
				len(acc.services) == 0 &&
				len(acc.Flows) == 0 {
				continue
			}
			// Create a new client for this account
			client, err := s.createInternalClientForAccount(acc, &natsclient.NatsClient{Internal: true})
			if err != nil {
				s.logger.Error("failed to create internal client", zap.String("account", acc.Name), zap.Error(err))
				continue
			}
			if err := client.Connect(); err != nil {
				s.logger.Error("failed to create internal client", zap.String("account", acc.Name), zap.Error(err))
				continue
			}
			// Keep track of the client to close it later
			s.clients = append(s.clients, client)
			// Make sure all streams exist
			for _, stream := range acc.Streams {
				ctx, cancel := context.WithTimeout(s.ctx, 3*time.Second)
				s.logger.Info("configuring stream", zap.String("name", stream.Name), zap.String("account", acc.Name))
				err := client.ConfigureStream(ctx, stream)
				cancel()
				if err != nil {
					s.logger.Error("failed to configure stream", zap.String("name", stream.Name), zap.String("account", acc.Name), zap.Error(err))
				} else {
					s.logger.Info("stream is configured", zap.String("name", stream.Name), zap.String("account", acc.Name))
				}
			}
			// Make sure all consumers exist
			for _, consumer := range acc.Consumers {
				ctx, cancel := context.WithTimeout(s.ctx, 3*time.Second)
				s.logger.Info("configuring consumer", zap.String("name", consumer.Name), zap.String("account", acc.Name))
				err := client.ConfigureConsumer(ctx, consumer)
				cancel()
				if err != nil {
					s.logger.Error("failed to configure consumer", zap.String("name", consumer.Name), zap.String("account", acc.Name), zap.Error(err))
				}
			}
			// Make sure all key value stores exist
			for _, store := range acc.KeyValueStores {
				ctx, cancel := context.WithTimeout(s.ctx, 3*time.Second)
				s.logger.Info("configuring key value store", zap.String("bucket", store.Bucket), zap.String("account", acc.Name))
				err := client.ConfigureKeyValueStore(ctx, store)
				cancel()
				if err != nil {
					s.logger.Error("failed to configure key value store", zap.String("account", acc.Name), zap.Error(err))
				}
			}
			// Make sure all object stores exist
			for _, store := range acc.ObjectStores {
				ctx, cancel := context.WithTimeout(s.ctx, 3*time.Second)
				s.logger.Info("configuring object store", zap.String("bucket", store.Bucket), zap.String("account", acc.Name))
				err := client.ConfigureObjectStore(ctx, store)
				cancel()
				if err != nil {
					s.logger.Error("failed to configure object store", zap.String("account", acc.Name), zap.Error(err))
				}
			}
			// Make sure all services exist
			for _, service := range acc.services {
				_, err := client.ConfigureService(service)
				if err != nil {
					s.logger.Error("failed to start service", zap.String("account", acc.Name), zap.Error(err))
				}
			}
			// Make sure all data flows are started
			for _, flow := range acc.Flows {
				s.logger.Info("starting data flow", zap.String("source", flow.source.CaddyModule().String()), zap.String("destination", flow.destination.CaddyModule().String()), zap.String("account", acc.Name))
				err := flow.Start()
				if err != nil {
					s.logger.Error("failed to start flow", zap.String("source", flow.source.CaddyModule().String()), zap.String("destination", flow.destination.CaddyModule().String()), zap.String("account", acc.Name), zap.Error(err))
				}
			}
		}
	}()

	return nil
}

// Helper method to create a client for the auth service.
func (s *Server) createClientForAuthService() error {
	if s.authService == nil {
		return nil
	}
	// Make sure we've got a client to run the auth service
	if s.AuthCallout.Client != nil {
		client := s.AuthCallout.Client
		s.logger.Info("creating client for auth service", zap.String("account", s.authService.account))
		if err := client.Provision(s.ctx, s); err != nil {
			return err
		}
		s.authClient = client
		return nil
	}
	var user *User
	if __authuser__ != nil {
		user = __authuser__
		// Override auth callout config
		s.serverOpts.AuthCallout = __options__.AuthCallout
	} else {
		// Create the user
		sk, err := nkeys.CreateUser()
		if err != nil {
			return err
		}
		pk, err := sk.PublicKey()
		if err != nil {
			return err
		}
		seed, err := sk.Seed()
		if err != nil {
			return err
		}
		// Create the account and the user
		user = &User{User: pk, Password: string(seed)}
		// Override auth callout config
		s.serverOpts.AuthCallout.AuthUsers = []string{pk}
	}
	// Create new account
	acc := Account{
		Name: s.authService.account, Users: []*User{user},
	}
	// Add the account to the server
	s.logger.Info("adding auth account", zap.String("name", s.authService.account))
	if err := s.Options.addAccount(s.serverOpts, &acc); err != nil {
		return err
	}
	s.logger.Info("creating in-process client for auth service", zap.String("account", s.authService.account))
	// Provision nats client
	s.authClient = &natsclient.NatsClient{Internal: true, Username: user.User, Password: user.Password}
	if err := s.authClient.Provision(s.ctx, s); err != nil {
		return err
	}
	// Set global variables
	__authuser__ = user
	__options__ = s.serverOpts
	return nil
}

// Helper method to create a client for any account.
func (s *Server) createInternalClientForAccount(account *Account, client *natsclient.NatsClient) (*natsclient.NatsClient, error) {
	if s.Options == nil {
		return nil, errors.New("server options are not provisioned")
	}
	switch {
	case s.Operators != nil:
		return nil, errors.New("cannot create internal client in operator mode")
	case s.Authorization != nil && s.Authorization.Users != nil:
		return nil, errors.New("cannot create internal client when multiple users are configured")
	case s.Authorization != nil && s.Authorization.Token != "":
		client.Token = s.Authorization.Token
	case s.Authorization != nil && s.Authorization.User != "":
		client.Username = s.Authorization.User
		client.Password = s.Authorization.Password
	case s.authService != nil:
		token, _, ok := __pinnedtokens__.Get(account.Name)
		if ok {
			client.Token = token
		} else {
			token, err := __pinnedtokens__.Add(account.Name, nil)
			if err != nil {
				return nil, err
			}
			client.Token = string(token)
		}
	}
	// Provision new client
	if err := client.Provision(s.ctx, s); err != nil {
		return nil, err
	}
	return client, nil
}

// helper function to check if a server is already running
func serverIsAlreadyRunning() bool { return __server__ != nil }

// helper function to check if a collector is already running
func collectorIsAlreadyRunning() bool { return __collector__ != nil }

// helper function to stop running processes (server and collector)
func stopRunningProcesses() {
	if __collector__ != nil {
		__collector__.Stop()
		__collector__ = nil
	}
	if __server__ != nil && __server__.Running() {
		__server__.Shutdown()
		__server__ = nil
	}
}

// Those variables are required to support reloading at the moment
// I'd like to have a "cleaner" way to support server reloading.
var (
	// The server options are used to create the server
	__options__ *server.Options
	// The NATS server
	__server__ *server.Server
	// The collector is used to collect metrics from the server
	__collector__ *natsmetrics.Collector
	// The auth service is used by auth service to sign jwt responses
	__authsigningkey__ nkeys.KeyPair
	// The auth public key is also called the "issuer" in the auth callout config
	__authpublickey__ string
	// The auth user is used by auth service to connect to in-process server
	__authuser__ *User
	// Pinned tokens are used to store tokens for internal clients
	// Since those tokens must persist accross reload, we store them in memory
	__pinnedtokens__ = natsauth.NewPinnedTokens()
	// Counter to detect reloads
	__last__ = 0
)
