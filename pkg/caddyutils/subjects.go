// Copyright 2023 Guillaume Charbonnier
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package caddyutils

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

// GetSubjectsForPolicices returns a slice of all subjects that are
// matched by the given policies.
func GetSubjectsForPolicices(policies caddytls.ConnectionPolicies) []string {
	subjects := SubjectSet{}
	for _, policy := range policies {
		subs := []string{}
		v, ok := policy.MatchersRaw["sni"]
		if !ok {
			continue
		}
		json.Unmarshal(v, &subs)
		for _, sub := range subs {
			subjects.Add(sub)
		}
	}
	return subjects.GetAll()
}

type SubjectSet map[string]struct{}

// Add adds one or several subjects to the set
func (s SubjectSet) Add(subject ...string) {
	for _, sub := range subject {
		s[sub] = struct{}{}
	}
}

// GetAll returns the list of subjects in the set
func (s SubjectSet) GetAll() []string {
	subjects := make([]string, 0, len(s))
	for subject := range s {
		subjects = append(subjects, subject)
	}
	return subjects
}
