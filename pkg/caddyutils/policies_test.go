// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package caddyutils_test

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/modules/caddytls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/caddyutils"
)

var _ = Describe("GetSubjectsForPolicices", func() {
	Context("for an empty ConnectionPolicies struct", func() {
		policies := caddytls.ConnectionPolicies{}
		It("should return an empty slice", func() {
			Expect(caddyutils.GetSubjectsForPolicices(policies)).To(BeEmpty())
		})
	})
	Context("for a ConnectionPolicies struct with a single policy", func() {
		policies := caddytls.ConnectionPolicies{
			{
				MatchersRaw: map[string]json.RawMessage{
					"sni": json.RawMessage(`["example.com"]`),
				},
			},
		}
		It("should return a slice with the subject", func() {
			Expect(caddyutils.GetSubjectsForPolicices(policies)).To(ConsistOf("example.com"))
		})
	})
	Context("for a ConnectionPolicies struct with several policies", func() {
		policies := caddytls.ConnectionPolicies{
			{
				MatchersRaw: map[string]json.RawMessage{
					"sni": json.RawMessage(`["example.com"]`),
				},
			},
			{
				MatchersRaw: map[string]json.RawMessage{
					"sni": json.RawMessage(`["example.org"]`),
				},
			},
		}
		It("should return a slice with the subjects", func() {
			Expect(caddyutils.GetSubjectsForPolicices(policies)).To(ConsistOf("example.com", "example.org"))
		})
	})
	Context("for a ConnectionPolicies struct with duplicate policies", func() {
		policies := caddytls.ConnectionPolicies{
			{
				MatchersRaw: map[string]json.RawMessage{
					"sni": json.RawMessage(`["example.com"]`),
				},
			},
			{
				MatchersRaw: map[string]json.RawMessage{
					"sni": json.RawMessage(`["example.com"]`),
				},
			},
		}
		It("should return a slice with the subject", func() {
			Expect(caddyutils.GetSubjectsForPolicices(policies)).To(ConsistOf("example.com"))
		})
	})
	Context("for a ConnectionPolicies struct with non-sni policies", func() {
		policies := caddytls.ConnectionPolicies{
			{
				MatchersRaw: map[string]json.RawMessage{
					"protocol": json.RawMessage(`["https"]`),
				},
			},
		}
		It("should return an empty slice", func() {
			Expect(caddyutils.GetSubjectsForPolicices(policies)).To(BeEmpty())
		})
	})
})
