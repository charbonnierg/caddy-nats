// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package embedded

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/server/certstore"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

type TLSMapType int

var (
	STANDARD_TLS_MAP  TLSMapType = 0
	WEBSOCKET_TLS_MAP TLSMapType = 1
	LEAFNODE_TLS_MAP  TLSMapType = 2
)

func (o *TLSMap) setTLSOpts(tlsMap TLSMapType, opts *server.Options) error {
	if o == nil {
		return nil
	}
	var setConfig func(cfg *tls.Config)
	switch tlsMap {
	case STANDARD_TLS_MAP:
		// Set tls global options (not sure this is useful)
		opts.TLSVerify = o.Verify
		opts.TLSMap = o.Map
		if len(o.PinnedCerts) > 0 {
			certs := map[string]struct{}{}
			for _, cert := range o.PinnedCerts {
				certs[cert] = struct{}{}
			}
			opts.TLSPinnedCerts = certs
		}
		opts.AllowNonTLS = o.AllowNonTLS
		opts.TLSRateLimit = o.RateLimit
		opts.TLSTimeout = o.Timeout.Seconds()
		// If TLS is managed, use the provided tls.Config
		if o.IsManaged() {
			opts.TLSConfig = o.config.Clone()
			return nil
		}
		setConfig = func(cfg *tls.Config) {
			opts.TLSConfig = cfg
		}
	case WEBSOCKET_TLS_MAP:
		// Set tls global options (not sure this is useful)
		opts.Websocket.TLSMap = o.Map
		if len(o.PinnedCerts) > 0 {
			certs := map[string]struct{}{}
			for _, cert := range o.PinnedCerts {
				certs[cert] = struct{}{}
			}
			opts.Websocket.TLSPinnedCerts = certs
		}
		// If TLS is managed, use the provided tls.Config
		if o.config != nil {
			opts.Websocket.TLSConfig = o.config.Clone()
			return nil
		}
		setConfig = func(cfg *tls.Config) {
			opts.Websocket.TLSConfig = cfg
		}
	case LEAFNODE_TLS_MAP:
		// Set tls global options (not sure this is useful)
		opts.LeafNode.TLSMap = o.Map
		if len(o.PinnedCerts) > 0 {
			certs := map[string]struct{}{}
			for _, cert := range o.PinnedCerts {
				certs[cert] = struct{}{}
			}
			opts.LeafNode.TLSPinnedCerts = certs
		}
		opts.LeafNode.TLSTimeout = o.Timeout.Seconds()
		// If TLS is managed, use the provided tls.Config
		if o.config != nil {
			opts.LeafNode.TLSConfig = o.config.Clone()
			return nil
		}
		setConfig = func(cfg *tls.Config) {
			opts.LeafNode.TLSConfig = cfg
		}
	default:
		return fmt.Errorf("invalid TLSMapType: %d", tlsMap)
	}

	// Parse ciphers
	var ciphers = []uint16{}
	for _, name := range o.Ciphers {
		cipher, err := parseutils.ParseCipherFromName(name)
		if err != nil {
			return err
		}
		ciphers = append(ciphers, cipher)
	}
	// Use default cipher suites if none specified
	if len(ciphers) == 0 {
		ciphers = parseutils.DefaultCipherSuites()
	}
	// Parse curves preferences
	var curves = []tls.CurveID{}
	for _, curve := range o.CurvePreferences {
		if curve == "" {
			continue
		}
		curveID, ok := parseutils.CurvePreferenceMap[curve]
		if !ok {
			return fmt.Errorf("invalid tls curve preference: %s", curve)
		}
		curves = append(curves, curveID)
	}
	// Use default curve preferences if none specified
	if len(curves) == 0 {
		curves = parseutils.DefaultCurvePreferences()
	}
	// Create the tls.Config from our options before including the certs.
	// It will determine the cipher suites that we prefer.
	config := tls.Config{
		MinVersion:       tls.VersionTLS12,
		CipherSuites:     ciphers,
		CurvePreferences: curves,
	}
	switch {
	case o.CertFile != "" && o.CertStore != "":
		return certstore.ErrConflictCertFileAndStore
	case o.CertFile != "" && o.KeyFile == "":
		return fmt.Errorf("missing 'key_file' in TLS configuration")
	case o.CertFile == "" && o.KeyFile != "":
		return fmt.Errorf("missing 'cert_file' in TLS configuration")
	case o.CertFile != "" && o.KeyFile != "":
		// Now load in cert and private key
		cert, err := tls.LoadX509KeyPair(o.CertFile, o.KeyFile)
		if err != nil {
			return fmt.Errorf("error parsing X509 certificate/key pair: %v", err)
		}
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return fmt.Errorf("error parsing certificate: %v", err)
		}
		config.Certificates = []tls.Certificate{cert}
	case o.CertStore != "":
		store, err := certstore.ParseCertStore(o.CertStore)
		if err != nil {
			return fmt.Errorf("invalid tls.cert_store option: %s", err.Error())
		}
		matchBy, err := certstore.ParseCertMatchBy(o.CertMatchBy)
		if err != nil {
			return fmt.Errorf("invalid tls.cert_match_by option: %s", err.Error())
		}
		err = certstore.TLSConfig(store, matchBy, o.CertMatch, &config)
		if err != nil {
			return fmt.Errorf("error generating tls config using cert_store: %s", err.Error())
		}
	}
	// Require client certificates as needed
	if o.Verify {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}
	// Add in CAs if applicable.
	if o.CaFile != "" {
		rootPEM, err := os.ReadFile(o.CaFile)
		if err != nil || rootPEM == nil {
			return fmt.Errorf("error reading tls root ca certificate: %s", err.Error())
		}
		pool := x509.NewCertPool()
		ok := pool.AppendCertsFromPEM(rootPEM)
		if !ok {
			return fmt.Errorf("error parsing tls root ca certificate")
		}
		config.ClientCAs = pool
	}
	// Set the config
	setConfig(&config)
	return nil
}
