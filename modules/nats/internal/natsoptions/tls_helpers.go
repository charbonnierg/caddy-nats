// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsoptions

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/server/certstore"
)

// Where we maintain all of the available ciphers
var cipherMap = map[string]uint16{
	"TLS_RSA_WITH_RC4_128_SHA":                tls.TLS_RSA_WITH_RC4_128_SHA,
	"TLS_RSA_WITH_3DES_EDE_CBC_SHA":           tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA":            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA256":         tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_RSA_WITH_AES_256_CBC_SHA":            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":         tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":        tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_RC4_128_SHA":          tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":     tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	"TLS_AES_128_GCM_SHA256":                  tls.TLS_AES_128_GCM_SHA256,
	"TLS_AES_256_GCM_SHA384":                  tls.TLS_AES_256_GCM_SHA384,
	"TLS_CHACHA20_POLY1305_SHA256":            tls.TLS_CHACHA20_POLY1305_SHA256,
}

// Used to verify that cipher exist when an integer is provided
var cipherMapByID = map[uint16]string{
	tls.TLS_RSA_WITH_RC4_128_SHA:                "TLS_RSA_WITH_RC4_128_SHA",
	tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:           "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
	tls.TLS_RSA_WITH_AES_128_CBC_SHA:            "TLS_RSA_WITH_AES_128_CBC_SHA",
	tls.TLS_RSA_WITH_AES_128_CBC_SHA256:         "TLS_RSA_WITH_AES_128_CBC_SHA256",
	tls.TLS_RSA_WITH_AES_256_CBC_SHA:            "TLS_RSA_WITH_AES_256_CBC_SHA",
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384:         "TLS_RSA_WITH_AES_256_GCM_SHA384",
	tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:        "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
	tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:          "TLS_ECDHE_RSA_WITH_RC4_128_SHA",
	tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:     "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	tls.TLS_AES_128_GCM_SHA256:                  "TLS_AES_128_GCM_SHA256",
	tls.TLS_AES_256_GCM_SHA384:                  "TLS_AES_256_GCM_SHA384",
	tls.TLS_CHACHA20_POLY1305_SHA256:            "TLS_CHACHA20_POLY1305_SHA256",
}

// Where we maintain available curve preferences
var curvePreferenceMap = map[string]tls.CurveID{
	"X25519":    tls.X25519,
	"CurveP256": tls.CurveP256,
	"CurveP384": tls.CurveP384,
	"CurveP521": tls.CurveP521,
}

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
	for _, cipher := range o.Ciphers {
		if cipher == "" {
			continue
		}
		cipherInt, err := strconv.Atoi(cipher)
		if err != nil {
			return fmt.Errorf("invalid tls cipher: %s", err.Error())
		}
		cipherUInt16, err := ParseCipherFromUInt16(uint16(cipherInt))
		if err != nil {
			return err
		}
		ciphers = append(ciphers, cipherUInt16)
	}
	// Use default cipher suites if none specified
	if len(ciphers) == 0 {
		ciphers = DefaultCipherSuites()
	}
	// Parse curves preferences
	var curves = []tls.CurveID{}
	for _, curve := range o.CurvePreferences {
		if curve == "" {
			continue
		}
		curveID, ok := curvePreferenceMap[curve]
		if !ok {
			return fmt.Errorf("invalid tls curve preference: %s", curve)
		}
		curves = append(curves, curveID)
	}
	// Use default curve preferences if none specified
	if len(curves) == 0 {
		curves = DefaultCurvePreferences()
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

// ParseCipherFromInt parses a cipher as uint16 from an integer
func ParseCipherFromUInt16(cipherKey uint16) (uint16, error) {
	_, ok := cipherMapByID[cipherKey]
	if !ok {
		return 0, fmt.Errorf("unknown cipher key: %d", cipherKey)
	}
	return cipherKey, nil
}

// ParseCipherFromName parses a cipher as a uint16 from a string
func ParseCipherFromName(cipherName string) (uint16, error) {
	cipher, exists := cipherMap[cipherName]
	if !exists {
		return 0, fmt.Errorf("unrecognized cipher %s", cipherName)
	}

	return cipher, nil
}

// ParseCurvePreferenceFromName parses curve preference as a [tls.CurveID] from a string
func ParseCurvePreferenceFromName(curveName string) (tls.CurveID, error) {
	curve, exists := curvePreferenceMap[curveName]
	if !exists {
		return 0, fmt.Errorf("unrecognized curve preference %s", curveName)
	}
	return curve, nil
}

func DefaultCurvePreferences() []tls.CurveID {
	return []tls.CurveID{
		tls.X25519, // faster than P256, arguably more secure
		tls.CurveP256,
		tls.CurveP384,
		tls.CurveP521,
	}
}

func DefaultCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	}
}
