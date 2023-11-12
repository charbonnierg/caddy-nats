package natsauth

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
)

// AuthorizationRequest is the interface used by the authorization callout.
type AuthorizationRequest interface {
	Claims() *jwt.AuthorizationRequestClaims
	Context() context.Context
	Replacer() *caddy.Replacer
}

type AuthorizationMatcher interface {
	Provision(ctx caddy.Context) error
	Match(request AuthorizationRequest) (bool, error)
}

type AuthorizationCallout interface {
	Provision(ctx caddy.Context, account string) error
	Handle(request AuthorizationRequest) (*jwt.UserClaims, error)
}

// AuthorizationPolicy is a struct holding the configuration for an authorization policy.
// It defines a list of matchers and a callout. Matchers are used to decide whether this policy
// should be applied to a given request. The callout is used to authenticate the request and
// generate the user claims.
type AuthorizationPolicy struct {
	callout  AuthorizationCallout
	matchers []AuthorizationMatcher

	CalloutRaw  json.RawMessage            `json:"callout,omitempty" caddy:"namespace=nats_server.callouts inline_key=module"`
	MatchersRaw map[string]json.RawMessage `json:"match,omitempty"`
}

// Provision loads the callout and matchers modules.
func (p *AuthorizationPolicy) Provision(ctx caddy.Context, account string) error {
	unm, err := ctx.LoadModule(p, "CalloutRaw")
	if err != nil {
		return err
	}
	callout, ok := unm.(AuthorizationCallout)
	if !ok {
		return errors.New("callout module is not a callout")
	}
	if err := callout.Provision(ctx, account); err != nil {
		return err
	}
	p.callout = callout
	for matcher, matcherRaw := range p.MatchersRaw {
		unm, err := ctx.LoadModuleByID("nats_server.matchers."+matcher, matcherRaw)
		if err != nil {
			return err
		}
		matcher, ok := unm.(AuthorizationMatcher)
		if !ok {
			return errors.New("matcher module is not a matcher")
		}
		if err := matcher.Provision(ctx); err != nil {
			return err
		}
		p.matchers = append(p.matchers, matcher)
	}
	return nil
}

// Match returns true if the request matches all the matchers of the policy.
// If any of the matchers returns an error, error is returned without continuing.
// If any of the matchers returns false, false is returned without continuing.
func (p *AuthorizationPolicy) Match(request AuthorizationRequest) (bool, error) {
	for _, matcher := range p.matchers {
		matched, err := matcher.Match(request)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

// Authorize calls the callout to authenticate the request and generate the user claims.
func (p *AuthorizationPolicy) Authorize(request AuthorizationRequest) (*jwt.UserClaims, error) {
	return p.callout.Handle(request)
}

// AuthorizationPolicies is a list of authorization policies.
// It defines the helper method MatchAndAuthorize which can be used
// to match a request against the policies and authorize it.
type AuthorizationPolicies []*AuthorizationPolicy

func (p AuthorizationPolicies) MatchAndAuthorize(ctx context.Context, claims *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	request := &authorizationRequest{
		claims:  claims,
		context: ctx,
	}
	pol, matched := p.match(request)
	if !matched {
		return nil, errors.New("no matching policy")
	}
	return pol.Authorize(request)
}

func (p AuthorizationPolicies) match(request AuthorizationRequest) (*AuthorizationPolicy, bool) {
	for _, pol := range p {
		matched, err := pol.Match(request)
		if err != nil {
			return nil, false
		}
		if matched {
			return pol, true
		}
	}
	return nil, false
}

type replacerCtxKey struct{}

// AuthorizationRequest is the authorization request.
type authorizationRequest struct {
	claims  *jwt.AuthorizationRequestClaims
	context context.Context
}

// Claims returns the claims of the authorization request.
func (r *authorizationRequest) Claims() *jwt.AuthorizationRequestClaims {
	return r.claims
}

// Context returns the context of the authorization request.
func (r *authorizationRequest) Context() context.Context {
	return r.context
}

// Replacer returns the caddy replacer of the authorization request.
func (r *authorizationRequest) Replacer() *caddy.Replacer {
	return r.getReplacer()
}

func (r *authorizationRequest) setReplacer() *caddy.Replacer {
	repl := caddy.NewReplacer()
	addSecretsVarsToReplacer(repl)
	addAuthRequestVarsToReplacer(repl, r.claims)
	r.context = context.WithValue(r.context, replacerCtxKey{}, repl)
	return repl
}

func (r *authorizationRequest) getReplacer() *caddy.Replacer {
	raw := r.context.Value(replacerCtxKey{})
	if raw == nil {
		return r.setReplacer()
	}
	repl, ok := raw.(*caddy.Replacer)
	if !ok {
		return r.setReplacer()
	}
	return repl
}

func addSecretsVarsToReplacer(repl *caddy.Replacer) {
	secretVars := func(key string) (any, bool) {
		filePrefix := "file."
		if strings.HasPrefix(key, filePrefix) {
			filename := strings.TrimPrefix(key, filePrefix)
			content, err := os.ReadFile(filename)
			if err != nil {
				return nil, false
			}
			return string(content), true
		}
		return nil, false
	}
	repl.Map(secretVars)
}

func addAuthRequestVarsToReplacer(repl *caddy.Replacer, req *jwt.AuthorizationRequestClaims) {
	natsVars := func(key string) (any, bool) {
		if req == nil {
			return nil, false
		}
		switch key {
		case "connect_opts.username":
			return req.ConnectOptions.Username, true
		case "connect_opts.password":
			return req.ConnectOptions.Password, true
		case "connect_opts.lang":
			return req.ConnectOptions.Lang, true
		case "connect_opts.version":
			return req.ConnectOptions.Version, true
		case "connect_opts.protocol":
			return req.ConnectOptions.Protocol, true
		case "client_info.id":
			return req.ClientInformation.ID, true
		case "client_info.name":
			return req.ClientInformation.Name, true
		case "client_info.host":
			return req.ClientInformation.Host, true
		case "client_info.user":
			return req.ClientInformation.User, true
		case "client_info.kind":
			return req.ClientInformation.Kind, true
		case "client_info.type":
			return req.ClientInformation.Type, true
		case "client_info.mqtt":
			return req.ClientInformation.MQTT, true
		case "user_nkey":
			return req.UserNkey, true
		default:
			return nil, false
		}
	}
	repl.Map(natsVars)
}
