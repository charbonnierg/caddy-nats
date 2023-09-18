package natsmagic

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

type AppPolicies struct {
	StandardPolicies  caddytls.ConnectionPolicies `json:"standard,omitempty"`
	WebsocketPolicies caddytls.ConnectionPolicies `json:"websocket,omitempty"`
	LeafnodePolicies  caddytls.ConnectionPolicies `json:"leafnode,omitempty"`
	MQTTPolicies      caddytls.ConnectionPolicies `json:"mqtt,omitempty"`
}

func (p *AppPolicies) Subjects() []string {
	var subjects []string
	for _, policy := range p.StandardPolicies {
		subs := []string{}
		v, ok := policy.MatchersRaw["sni"]
		if !ok {
			continue
		}
		json.Unmarshal(v, &subs)
		subjects = append(subjects, subs...)
	}
	for _, policy := range p.WebsocketPolicies {
		subs := []string{}
		v, ok := policy.MatchersRaw["sni"]
		if !ok {
			continue
		}
		json.Unmarshal(v, &subs)
		subjects = append(subjects, subs...)
	}
	for _, policy := range p.LeafnodePolicies {
		subs := []string{}
		v, ok := policy.MatchersRaw["sni"]
		if !ok {
			continue
		}
		json.Unmarshal(v, &subs)
		subjects = append(subjects, subs...)
	}
	for _, policy := range p.MQTTPolicies {
		subs := []string{}
		v, ok := policy.MatchersRaw["sni"]
		if !ok {
			continue
		}
		json.Unmarshal(v, &subs)
		subjects = append(subjects, subs...)
	}
	return subjects
}
