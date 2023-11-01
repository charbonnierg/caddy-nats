package auth

import (
	"encoding/json"

	"github.com/quara-dev/beyond/modules/nats/auth/policies"
	"github.com/quara-dev/beyond/pkg/natsutils"
)

// AuthServiceConfig is the configuration for the auth callout service.
type AuthServiceConfig struct {
	ClientRaw         *natsutils.Client           `json:"client,omitempty"`
	InternalAccount   string                      `json:"internal_account,omitempty"`
	InternalUser      string                      `json:"internal_user,omitempty"`
	AuthAccount       string                      `json:"auth_account,omitempty"`
	AuthSigningKey    string                      `json:"auth_signing_key,omitempty"`
	SubjectRaw        string                      `json:"subject,omitempty"`
	Policies          policies.ConnectionPolicies `json:"policies,omitempty"`
	DefaultHandlerRaw json.RawMessage             `json:"handler,omitempty" caddy:"namespace=nats.auth_callout inline_key=module"`
}

func (a *AuthServiceConfig) Zero() bool {
	if a == nil {
		return true
	}
	return a.ClientRaw == nil &&
		a.InternalAccount == "" &&
		a.InternalUser == "" &&
		a.AuthAccount == "" &&
		a.AuthSigningKey == "" &&
		a.SubjectRaw == "" &&
		a.Policies == nil &&
		a.DefaultHandlerRaw == nil
}
