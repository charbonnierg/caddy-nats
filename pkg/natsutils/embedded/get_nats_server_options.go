// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package embedded

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
)

// GetServerOptions returns the options for the NATS server.
func (o *Options) GetServerOptions() (*server.Options, error) {
	// Initialize options
	// DisableJetStreamBanner is always set to true
	// NoSigs is always set to true
	serverOpts := server.Options{
		DisableJetStreamBanner: true,
		NoSigs:                 true,
	}
	// Verify and set global options
	if err := o.setGlobalOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Static token auth
	if err := o.setTokenAuth(&serverOpts); err != nil {
		return nil, err
	}
	// Static user/password auth
	if err := o.setUserPasswordAuth(&serverOpts); err != nil {
		return nil, err
	}
	// Static multi-users auth
	if err := o.setUsersAuth(&serverOpts); err != nil {
		return nil, err
	}
	// Static accounts auth
	if err := o.setAccountsAuth(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set no auth user
	if err := o.setNoAuthUser(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set auth callout
	if err := o.setAuthCallout(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set monitoring options
	if err := o.setMonitoringOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Veirfy and set cluster options
	if err := o.setClusterOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set jetstream options
	if err := o.setJetStreamOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set leafnode options
	if err := o.setLeafnodeOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set websocket options
	if err := o.setWebsocketOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set mqtt options
	if err := o.setMqttOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set system account
	if err := o.setSystemAccountOpt(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set operator mode options
	if err := o.setOperatorModeOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set resolver options
	if err := o.setResolverOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Verify and set standard tls options
	if err := o.setTLSOpts(&serverOpts); err != nil {
		return nil, err
	}
	// Gather options
	return &serverOpts, nil
}

func (o *Options) setGlobalOpts(opts *server.Options) error {
	opts.ServerName = o.ServerName
	tags := []string{}
	for key, value := range o.ServerTags {
		tag := strings.Join([]string{key, value}, ":")
		tags = append(tags, tag)
	}
	opts.Tags = tags
	opts.Host = o.Host
	opts.Port = o.Port
	opts.ClientAdvertise = o.Advertise
	opts.Debug = o.Debug
	opts.Trace = o.Trace
	opts.TraceVerbose = o.TraceVerbose
	opts.NoLog = o.NoLog
	opts.NoSublistCache = o.NoSublistCache
	opts.MaxConn = o.MaxConn
	opts.MaxSubs = o.MaxSubs
	opts.MaxPayload = o.MaxPayload
	opts.MaxPending = o.MaxPending
	opts.MaxClosedClients = o.MaxClosedClients
	opts.MaxControlLine = o.MaxControlLine
	opts.MaxPingsOut = o.MaxPingsOut
	opts.MaxSubTokens = o.MaxSubsTokens
	opts.PingInterval = o.PingInterval
	opts.MaxTracedMsgLen = o.MaxTracedMsgLen
	opts.WriteDeadline = o.WriteDeadline
	return nil
}

func (o *Options) setTLSOpts(opts *server.Options) error {
	if o.TLS != nil {
		// Verify and set standard tls options
		if err := o.TLS.setTLSOpts(STANDARD_TLS_MAP, opts); err != nil {
			return err
		}
	}
	if o.Websocket != nil && o.Websocket.TLS != nil {
		// Verify and set websocket tls options
		if err := o.Websocket.TLS.setTLSOpts(WEBSOCKET_TLS_MAP, opts); err != nil {
			return err
		}
	}
	if o.Leafnode != nil && o.Leafnode.TLS != nil {
		// Verify and set leafnode tls options
		if err := o.Leafnode.TLS.setTLSOpts(LEAFNODE_TLS_MAP, opts); err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) setResolverOpts(opts *server.Options) error {
	if o.FullResolver == nil && o.CacheResolver == nil && o.MemoryResolver == nil {
		if o.Operators != nil {
			return errors.New("operators are set but resolver is not configured")
		}
		return nil
	}
	if o.FullResolver != nil {
		if o.MemoryResolver != nil || o.CacheResolver != nil {
			return errors.New("full_resolver and memory_resolver/cache_resolver cannot be set at the same time")
		}
		var deleteType = server.NoDelete
		if o.FullResolver.AllowDelete && o.FullResolver.HardDelete {
			deleteType = server.HardDelete
		} else if o.FullResolver.AllowDelete {
			deleteType = server.RenameDeleted
		}

		resolver, err := server.NewDirAccResolver(o.FullResolver.Path, o.FullResolver.Limit, o.FullResolver.SyncInterval, deleteType)
		if err != nil {
			return fmt.Errorf("invalid full resolver: %s", err.Error())
		}
		if o.SystemAccount != "" && o.systemAccount != nil {
			if err := resolver.Store(o.systemAccount.Subject, o.SystemAccount); err != nil {
				return fmt.Errorf("invalid system account: %s", err.Error())
			}
		}
		for _, entry := range o.FullResolver.Preload {
			claims, err := jwt.DecodeAccountClaims(entry)
			if err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
			if err := resolver.Store(claims.Subject, entry); err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
		}
		opts.AccountResolver = resolver
	}
	if o.MemoryResolver != nil {
		if o.CacheResolver != nil || o.FullResolver != nil {
			return errors.New("memory_resolver and cache_resolver/full_resolver cannot be set at the same time")
		}
		resolver := server.MemAccResolver{}
		if o.SystemAccount != "" && o.systemAccount != nil {
			if err := resolver.Store(o.systemAccount.Subject, o.SystemAccount); err != nil {
				return fmt.Errorf("invalid system account: %s", err.Error())
			}
		}
		for _, entry := range o.MemoryResolver.Preload {
			claims, err := jwt.DecodeAccountClaims(entry)
			if err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
			if err := resolver.Store(claims.Subject, entry); err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
		}
		opts.AccountResolver = &resolver
	}
	if o.CacheResolver != nil {
		if o.MemoryResolver != nil || o.FullResolver != nil {
			return errors.New("cache_resolver and memory_resolver/full_resolver cannot be set at the same time")
		}
		resolver, err := server.NewCacheDirAccResolver(
			o.CacheResolver.Path,
			int64(o.CacheResolver.Limit),
			o.CacheResolver.TTL,
		)
		if err != nil {
			return fmt.Errorf("invalid cache resolver: %s", err.Error())
		}
		if o.SystemAccount != "" && o.systemAccount != nil {
			if err := resolver.Store(o.systemAccount.Subject, o.SystemAccount); err != nil {
				return fmt.Errorf("invalid system account: %s", err.Error())
			}
		}
		for _, entry := range o.CacheResolver.Preload {
			claims, err := jwt.DecodeAccountClaims(entry)
			if err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
			if err := resolver.Store(claims.Subject, entry); err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
		}
		opts.AccountResolver = resolver
	}
	return nil
}

func (o *Options) setSystemAccountOpt(opts *server.Options) error {
	if o.Operators != nil {
		// Parse system account jwt
		claims, err := jwt.DecodeAccountClaims(o.SystemAccount)
		if err != nil {
			return fmt.Errorf("invalid system account: %s", err.Error())
		}
		opts.SystemAccount = claims.Subject
		o.systemAccount = claims
		return nil
	}
	// Don't attempt to parse, system account may be a simple name, maybe it's empty
	opts.SystemAccount = o.SystemAccount
	// Check is system account must be created
	if o.SystemAccount == "" && o.systemAccount == nil && o.Accounts != nil {
		// We have accounts, but we don't have a system account.
		// Let's create one named "SYS"
		o.SystemAccount = "SYS"
		// If this account already exists, raise an error, because we don't know
		// if administrator is aware that this will be the system account or not
		for _, account := range o.Accounts {
			if account.Name == o.SystemAccount {
				return errors.New("system account must be explicitely specified when an account named SYS is used")
			}
		}
		o.addAccount(opts, &Account{Name: o.SystemAccount})
	}
	return nil
}

func (o *Options) setOperatorModeOpts(opts *server.Options) error {
	if o.Operators == nil {
		return nil
	}
	operators := []*jwt.OperatorClaims{}
	for _, token := range o.Operators {
		claims, err := jwt.DecodeOperatorClaims(token)
		if err != nil {
			return fmt.Errorf("invalid operator token: %s", err.Error())
		}
		// Set default system account
		if opts.SystemAccount == "" && claims.SystemAccount != "" {
			opts.SystemAccount = claims.SystemAccount
		}
		operators = append(operators, claims)
	}
	opts.TrustedOperators = operators
	return nil
}

func (o *Options) setTokenAuth(opts *server.Options) error {
	if o.Authorization == nil {
		return nil
	}
	if o.Authorization.Token == "" {
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.token and operators cannot be set at the same time")
	}
	if o.Authorization.Users != nil {
		return errors.New("authorization.token and authorization.users cannot be set at the same time")
	}
	if o.Authorization.User != "" {
		return errors.New("authorization.token and authorization.user cannot be set at the same time")
	}
	opts.Authorization = o.Authorization.Token
	return nil
}

func (o *Options) setUserPasswordAuth(opts *server.Options) error {
	if o.Authorization == nil {
		return nil
	}
	if o.Authorization.User == "" {
		if o.Authorization.Password != "" {
			return errors.New("authorization.password cannot be set without authorization.user")
		}
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.user and operators cannot be set at the same time")
	}
	if o.Authorization.Users != nil {
		return errors.New("authorization.user and authorization.users cannot be set at the same time")
	}
	opts.Username = o.Authorization.User
	opts.Password = o.Authorization.Password
	return nil
}

func (o *Options) addUser(opts *server.Options, user *User) error {
	if user.User == "" {
		return errors.New("cannot add user without a name")
	}
	if user.Password == "" {
		return errors.New("cannot add user without a password")
	}
	allowedConnTypes, err := validateConnectionTypes(user.AllowedConnectionTypes)
	if err != nil {
		return err
	}
	opts.Users = append(opts.Users, &server.User{
		Username:               user.User,
		Password:               user.Password,
		Permissions:            user.Permissions,
		AllowedConnectionTypes: allowedConnTypes,
	})
	return nil
}

func (o *Options) setUsersAuth(opts *server.Options) error {
	if o.Authorization == nil {
		return nil
	}
	if o.Authorization.Users == nil {
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.users and operators cannot be set at the same time")
	}
	if o.Authorization.Users == nil {
		return nil
	}
	if len(o.Authorization.Users) == 0 {
		return errors.New("authorization.users must either be omitted or set with at least one user")
	}
	opts.Users = []*server.User{}
	for _, user := range o.Authorization.Users {
		if err := o.addUser(opts, &user); err != nil {
			return fmt.Errorf("invalid user: %s", err.Error())
		}
	}
	return nil
}

func (o *Options) addAccountUser(opts *server.Options, account *server.Account, user *User) error {
	if user.User == "" {
		return errors.New("cannot add an account user without a name")
	}
	if user.Password == "" {
		return errors.New("cannot add an account user without a password")
	}
	allowedConnTypes, err := validateConnectionTypes(user.AllowedConnectionTypes)
	if err != nil {
		return err
	}
	accUser := server.User{
		Username:               user.User,
		Password:               user.Password,
		Permissions:            user.Permissions,
		AllowedConnectionTypes: allowedConnTypes,
		Account:                account,
	}
	opts.Users = append(opts.Users, &accUser)
	return nil
}

func (o *Options) addAccount(opts *server.Options, account *Account) error {
	if account.Name == "" {
		return errors.New("authorization.accounts.name cannot be empty")
	}
	acc := server.NewAccount(account.Name)
	// Add mappings
	for _, mapping := range account.Mappings {
		if err := acc.AddWeightedMappings(mapping.Subject, mapping.MapDest...); err != nil {
			return fmt.Errorf("invalid account subject mapping: %s", err.Error())
		}
	}
	// Add users
	for _, user := range account.Users {
		if err := o.addAccountUser(opts, acc, &user); err != nil {
			return fmt.Errorf("invalid user: %s", err.Error())
		}
	}
	opts.Accounts = append(opts.Accounts, acc)
	return nil
}

func (o *Options) setAuthCallout(opts *server.Options) error {
	if o.Authorization == nil {
		return nil
	}
	if o.Authorization.AuthCallout == nil {
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.auth_callout and operators cannot be set at the same time")
	}
	acc := o.Authorization.AuthCallout.Account
	if acc == "" {
		acc = "$G"
	}
	opts.AuthCallout = &server.AuthCallout{
		Account:   acc,
		Issuer:    o.Authorization.AuthCallout.Issuer,
		AuthUsers: o.Authorization.AuthCallout.AuthUsers,
		XKey:      o.Authorization.AuthCallout.XKey,
	}
	return nil
}

func (o *Options) setAccountsAuth(opts *server.Options) error {
	if o.Accounts == nil {
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.accounts and operators cannot be set at the same time")
	}
	if o.Authorization != nil && o.Authorization.Users != nil {
		return errors.New("authorization.accounts and authorization.users cannot be set at the same time")
	}
	if o.Authorization != nil && o.Authorization.User != "" {
		return errors.New("authorization.accounts and authorization.user cannot be set at the same time")
	}
	if o.Authorization != nil && o.Authorization.Token != "" {
		return errors.New("authorization.accounts and authorization.token cannot be set at the same time")
	}
	if len(o.Accounts) == 0 {
		return errors.New("authorization.accounts must either be omitted or set with at least one account")
	}
	opts.Accounts = []*server.Account{}
	for _, account := range o.Accounts {
		if err := o.addAccount(opts, account); err != nil {
			return fmt.Errorf("invalid account: %s", err.Error())
		}
	}
	return nil
}

func (o *Options) setNoAuthUser(opts *server.Options) error {
	if o.NoAuthUser != "" {
		if o.Operators != nil {
			return errors.New("no_auth_user and operators cannot be set at the same time")
		}
		opts.NoAuthUser = o.NoAuthUser
	}
	return nil
}

func (o *Options) setMonitoringOpts(opts *server.Options) error {
	// Verify that one of http_port or https_port may be defined but not both
	if o.HTTPPort != 0 && o.HTTPSPort != 0 {
		return errors.New("metrics.http_port and metrics.https_port cannot be set at the same time")
	}
	opts.HTTPPort = o.HTTPPort
	opts.HTTPSPort = o.HTTPSPort
	opts.HTTPHost = o.HTTPHost
	opts.HTTPBasePath = o.HTTPBasePath
	return nil
}

func (o *Options) setJetStreamOpts(opts *server.Options) error {
	if o.JetStream == nil {
		return nil
	}
	opts.JetStream = true
	opts.JetStreamMaxMemory = o.JetStream.MaxMemory
	opts.JetStreamMaxStore = o.JetStream.MaxFile
	opts.StoreDir = o.JetStream.StoreDir
	opts.JetStreamDomain = o.JetStream.Domain
	opts.JetStreamUniqueTag = o.JetStream.UniqueTag
	return nil
}

func (o *Options) setMqttOpts(opts *server.Options) error {
	if o.Mqtt == nil {
		return nil
	}
	if o.JetStream == nil {
		return errors.New("mqtt cannot be enabled without jetstream")
	}
	if o.ServerName == "" {
		return errors.New("mqtt cannot be enabled without server name")
	}
	// Set default port if none specified
	var port = o.Mqtt.Port
	if port == 0 {
		if o.Mqtt.TLS != nil {
			port = 8883
		} else {
			port = 1883
		}
	}
	opts.MQTT.Host = o.Mqtt.Host
	opts.MQTT.Port = port
	opts.MQTT.Username = o.Mqtt.Username
	opts.MQTT.Password = o.Mqtt.Password
	opts.MQTT.AuthTimeout = o.Mqtt.AuthTimeout
	opts.MQTT.StreamReplicas = o.Mqtt.StreamReplicas
	opts.MQTT.NoAuthUser = o.Mqtt.NoAuthUser
	return nil
}

func (o *Options) setWebsocketOpts(opts *server.Options) error {
	if o.Websocket == nil {
		return nil
	}
	var port = o.Websocket.Port
	if port == 0 {
		if o.Websocket.TLS != nil {
			port = 10443
		} else {
			port = 10080
		}
	}
	opts.Websocket.Host = o.Websocket.Host
	opts.Websocket.Port = port
	opts.Websocket.Advertise = o.Websocket.Advertise
	opts.Websocket.NoTLS = o.Websocket.NoTLS
	opts.Websocket.Username = o.Websocket.Username
	opts.Websocket.Password = o.Websocket.Password
	opts.Websocket.NoAuthUser = o.Websocket.NoAuthUser
	opts.Websocket.Compression = o.Websocket.Compression
	opts.Websocket.SameOrigin = o.Websocket.SameOrigin
	opts.Websocket.AllowedOrigins = o.Websocket.AllowedOrigins
	opts.Websocket.JWTCookie = o.Websocket.JWTCookie
	return nil
}

func (o *Options) setLeafnodeOpts(opts *server.Options) error {
	if o.Leafnode == nil {
		return nil
	}
	port := o.Leafnode.Port
	// Set default listening port when no remotes are defined
	if port == 0 && len(o.Leafnode.Remotes) == 0 {
		port = 7422
	}
	opts.LeafNode.Host = o.Leafnode.Host
	opts.LeafNode.Port = port
	opts.LeafNode.Advertise = o.Leafnode.Advertise
	if len(o.Leafnode.Remotes) == 0 {
		return nil
	}
	opts.LeafNode.Remotes = make([]*server.RemoteLeafOpts, len(o.Leafnode.Remotes))
	for i, remote := range o.Leafnode.Remotes {
		if remote.Url == "" && len(remote.Urls) == 0 {
			return errors.New("leafnode.remotes.url or leafnode.remotes.urls must be set")
		}
		if remote.Url != "" && len(remote.Urls) > 0 {
			return errors.New("leafnode.remotes.url and leafnode.remotes.urls cannot be set at the same time")
		}
		urls := []*url.URL{}
		if remote.Url != "" {
			remote.Urls = append(remote.Urls, remote.Url)
		}
		for _, r := range remote.Urls {
			remoteUrl, err := url.Parse(r)
			if err != nil {
				return fmt.Errorf("invalid remote leafnode url: %s", err.Error())
			}
			urls = append(urls, remoteUrl)
		}
		opts.LeafNode.Remotes[i] = &server.RemoteLeafOpts{
			URLs:         urls,
			NoRandomize:  remote.NoRandomize,
			LocalAccount: remote.Account,
			Hub:          remote.Hub,
			Credentials:  remote.Credentials,
			DenyImports:  remote.DenyImports,
			DenyExports:  remote.DenyExports,
			Websocket: struct {
				Compression bool `json:"-"`
				NoMasking   bool `json:"-"`
			}(remote.Websocket),
		}
	}
	return nil
}

func (o *Options) setClusterOpts(opts *server.Options) error {
	if o.Cluster == nil {
		return nil
	}
	if o.Cluster.Name == "" {
		return errors.New("cluster.name cannot be empty")
	}
	port := o.Cluster.Port
	if port == 0 {
		port = 6222
	}
	opts.Cluster.Name = o.Cluster.Name
	opts.Cluster.Host = o.Cluster.Host
	opts.Cluster.Port = port
	opts.Cluster.Advertise = o.Cluster.Advertise
	opts.Cluster.NoAdvertise = o.Cluster.NoAdvertise
	opts.Cluster.ConnectRetries = o.Cluster.ConnectRetries
	opts.Cluster.PoolSize = o.Cluster.PoolSize
	if o.Cluster.Compression != nil {
		opts.Cluster.Compression = server.CompressionOpts{Mode: o.Cluster.Compression.Mode, RTTThresholds: o.Cluster.Compression.RTTThresholds}
	}
	if len(o.Cluster.Routes) == 0 {
		opts.Routes = make([]*url.URL, 1)
		routeUrl, err := url.Parse(fmt.Sprintf("nats-route://localhost:%d", port))
		if err != nil {
			return fmt.Errorf("invalid cluster route url: %s", err.Error())
		}
		opts.Routes[0] = routeUrl
	} else {
		opts.Routes = make([]*url.URL, len(o.Cluster.Routes))
		for i, route := range o.Cluster.Routes {
			routeUrl, err := url.Parse(route)
			if err != nil {
				return fmt.Errorf("invalid cluster route url: %s", err.Error())
			}
			opts.Routes[i] = routeUrl
		}
	}
	return nil
}

func validateConnectionTypes(allowedConnectionTypes []string) (map[string]struct{}, error) {
	allowed := map[string]struct{}{}
	for _, connType := range allowedConnectionTypes {
		typ := strings.ToUpper(connType)
		if typ != jwt.ConnectionTypeStandard &&
			typ != jwt.ConnectionTypeWebsocket &&
			typ != jwt.ConnectionTypeMqtt &&
			typ != jwt.ConnectionTypeMqttWS &&
			typ != jwt.ConnectionTypeLeafnode &&
			typ != jwt.ConnectionTypeLeafnodeWS {
			return nil, fmt.Errorf("invalid connection type: %q", connType)
		}
		allowed[connType] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, nil
	}
	return allowed, nil
}
