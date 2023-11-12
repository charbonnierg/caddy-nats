package natstls_test

import (
	"crypto/tls"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quara-dev/beyond/modules/caddynats/natstls"
)

func TestParseCipherFromNameError(t *testing.T) {
	cases := []string{
		"",
		"abc",
	}
	for _, case_ := range cases {
		t.Run(case_, func(t *testing.T) {
			_, err := natstls.ParseCipherFromName(case_)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestParseCurvePreference(t *testing.T) {
	for key, value := range natstls.CurvePreferenceMap {
		result, err := natstls.ParseCurvePreferenceFromName(key)
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
			_, err := natstls.ParseCurvePreferenceFromName(case_)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestDefaultCurvePreferences(t *testing.T) {
	defaultCurvePreferences := natstls.DefaultCurvePreferences()
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
	defaultCipherSuites := natstls.DefaultCipherSuites()
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
