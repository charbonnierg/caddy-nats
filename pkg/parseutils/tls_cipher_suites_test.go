// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parseutils_test

import (
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

func TestParseCipherFromInt(t *testing.T) {
	for key, value := range parseutils.CipherMap {
		result, err := parseutils.ParseCipherFromInt(int(value))
		if err != nil {
			t.Errorf("ParseCipherFromInt(%v) error: %v", value, err)
		}
		if result != value {
			t.Errorf("ParseCipherFromInt(%v) != %v", value, key)
		}
		backward, err := parseutils.ParseCipherFromName(key)
		if err != nil {
			t.Errorf("ParseCipherFromName(%v) error: %v", key, err)
		}
		if backward != value {
			t.Errorf("ParseCipherFromName(%v) != %v", key, value)
		}
	}
}

func TestParseCipherFromIntError(t *testing.T) {
	cases := []int{
		-1,
		0,
		1,
		2,
		3,
		4,
	}
	for _, case_ := range cases {
		t.Run(fmt.Sprintf("%d", case_), func(t *testing.T) {
			_, err := parseutils.ParseCipherFromInt(case_)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestParseCipherFromNameError(t *testing.T) {
	cases := []string{
		"",
		"abc",
	}
	for _, case_ := range cases {
		t.Run(case_, func(t *testing.T) {
			_, err := parseutils.ParseCipherFromName(case_)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestParseCurvePreference(t *testing.T) {
	for key, value := range parseutils.CurvePreferenceMap {
		result, err := parseutils.ParseCurvePreferenceFromName(key)
		if err != nil {
			t.Errorf("ParseCurvePreference(%v) error: %v", key, err)
		}
		if result != value {
			t.Errorf("ParseCurvePreference(%v) != %v", key, value)
		}
	}
}

func TestParseCurvePreferenceError(t *testing.T) {
	cases := []string{
		"",
		"abc",
	}
	for _, case_ := range cases {
		t.Run(case_, func(t *testing.T) {
			_, err := parseutils.ParseCurvePreferenceFromName(case_)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestDefaultCurvePreferences(t *testing.T) {
	defaultCurvePreferences := parseutils.DefaultCurvePreferences()
	if !cmp.Equal(defaultCurvePreferences, []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
		tls.CurveP384,
		tls.CurveP521,
	}) {
		t.Errorf("DefaultCurvePreferences() != %v", defaultCurvePreferences)
	}
}

func TestDefaultCipherSuites(t *testing.T) {
	defaultCipherSuites := parseutils.DefaultCipherSuites()
	if !cmp.Equal(defaultCipherSuites, []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	}) {
		t.Errorf("DefaultCipherSuites() != %v", defaultCipherSuites)
	}
}
