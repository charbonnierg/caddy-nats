// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package caddyutils

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/quara-dev/beyond/pkg/datatypes"
)

// GetSubjectsForPolicices returns a slice of all subjects that are
// matched by the given policies.
func GetSubjectsForPolicices(policies caddytls.ConnectionPolicies) []string {
	subjects := datatypes.Set[string]{}
	// Iterate over all policies, but we're only interested in the SNI matchers
	for _, policy := range policies {
		subs := []string{}
		v, ok := policy.MatchersRaw["sni"]
		// If there is no SNI matcher, we can skip this policy
		if !ok {
			continue
		}
		json.Unmarshal(v, &subs)
		if len(subs) == 0 {
			continue
		}
		// Add all subjects to the set
		subjects.Add(subs[0], subs[1:]...)
	}
	return subjects.Slice()
}
