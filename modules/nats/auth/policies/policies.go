// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package policies

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/nats"
)

type ConnectionPolicies []*ConnectionPolicy

func (pols ConnectionPolicies) Match(request *jwt.AuthorizationRequestClaims) (*ConnectionPolicy, bool) {
	for _, pol := range pols {
		if pol.Match(request) {
			return pol, true
		}
	}
	return nil, false
}

func (pols *ConnectionPolicies) Provision(app nats.App) error {
	for _, pol := range *pols {
		if err := pol.Provision(app); err != nil {
			return err
		}
	}
	return nil
}

type ConnectionPolicy struct {
	matchers    []Matcher
	handler     nats.AuthCallout
	MatchersRaw []map[string]json.RawMessage `json:"match,omitempty" caddy:"namespace=nats.matchers"`
	HandlerRaw  json.RawMessage              `json:"handler" caddy:"namespace=nats.auth_callout inline_key=module"`
}

func (pol *ConnectionPolicy) SetAccount(account string) error {
	if pol.handler != nil {
		if err := pol.handler.SetAccount(account); err != nil {
			return err
		}
	}
	return nil
}

func (pol *ConnectionPolicy) Match(request *jwt.AuthorizationRequestClaims) bool {
	var matched = false
	for _, m := range pol.matchers {
		if !m.Match(request) {
			return false
		} else {
			matched = true
		}
	}
	return matched
}

func (pol *ConnectionPolicy) Handle(request nats.AuthRequest) (*jwt.UserClaims, error) {
	return pol.handler.Handle(request)
}

func (c *ConnectionPolicy) Provision(app nats.App) error {
	if err := c.loadMatchers(app); err != nil {
		return err
	}
	if err := c.loadHandler(app); err != nil {
		return err
	}
	return nil
}

func (c *ConnectionPolicy) loadMatchers(app nats.App) error {
	unm, err := app.Context().LoadModule(c, "MatchersRaw")
	if err != nil {
		return fmt.Errorf("failed to load matchers: %s", err.Error())
	}
	matchers, ok := unm.([]map[string]interface{})
	if !ok {
		return errors.New("matchers invalid type: must be an array of maps")
	}
	for _, matcher := range matchers {

		for _, m := range matcher {
			matcher, ok := m.(Matcher)
			if !ok {
				return errors.New("matcher invalid type: must be a matcher")
			}
			c.matchers = append(c.matchers, matcher)
		}
	}
	return nil
}

func (c *ConnectionPolicy) loadHandler(app nats.App) error {
	unm, err := app.Context().LoadModule(c, "HandlerRaw")
	if err != nil {
		return fmt.Errorf("failed to load auth callout handler: %s", err.Error())
	}
	handler, ok := unm.(nats.AuthCallout)
	if !ok {
		return errors.New("auth callout handler invalid type")
	}
	if err := handler.Provision(app); err != nil {
		return fmt.Errorf("failed to provision auth callout handler: %s", err.Error())
	}
	c.handler = handler
	return nil
}
