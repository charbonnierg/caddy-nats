package nats

import (
	"os"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
)

type ReplacerCtxKey struct{}

type Template jwt.User

func (t *Template) Render(request *AuthorizationRequest, user *jwt.UserClaims) {
	repl := request.GetReplacer()
	// Copy the template
	if t.Payload != 0 {
		user.NatsLimits.Payload = t.Payload
	}
	if t.Subs != 0 {
		user.NatsLimits.Subs = t.Subs
	}
	if t.Data != 0 {
		user.NatsLimits.Data = t.Payload
	}
	if t.Permissions.Pub.Allow != nil {
		user.Permissions.Pub.Allow = jwt.StringList{}
		for _, allow := range t.Permissions.Pub.Allow {
			user.Permissions.Pub.Allow.Add(repl.ReplaceKnown(allow, ""))
		}
	}
	if t.Permissions.Pub.Deny != nil {
		user.Permissions.Pub.Deny = jwt.StringList{}
		for _, deny := range t.Permissions.Pub.Deny {
			user.Permissions.Pub.Deny.Add(repl.ReplaceKnown(deny, ""))
		}
	}
	if t.Permissions.Sub.Allow != nil {
		user.Permissions.Sub.Allow = jwt.StringList{}
		for _, allow := range t.Permissions.Sub.Allow {
			user.Permissions.Sub.Allow.Add(repl.ReplaceKnown(allow, ""))
		}
	}
	if t.Permissions.Sub.Deny != nil {
		user.Permissions.Sub.Deny = jwt.StringList{}
		for _, deny := range t.Permissions.Sub.Deny {
			user.Permissions.Sub.Deny.Add(repl.ReplaceKnown(deny, ""))
		}
	}
	if t.Permissions.Resp != nil {
		user.Permissions.Resp = t.Permissions.Resp
	}
	if t.Times != nil {
		user.UserLimits.Times = t.Times
	}
	if t.Src != nil {
		user.UserLimits.Src = jwt.CIDRList{}
		for _, src := range t.Src {
			user.UserLimits.Src.Add(repl.ReplaceKnown(src, ""))
		}
	}
	if t.Locale != "" {
		user.UserLimits.Locale = repl.ReplaceKnown(t.Locale, "")
	}
	if t.AllowedConnectionTypes != nil {
		user.AllowedConnectionTypes = jwt.StringList{}
		for _, connType := range t.AllowedConnectionTypes {
			user.AllowedConnectionTypes.Add(repl.ReplaceKnown(connType, ""))
		}
	}
	if t.BearerToken {
		user.BearerToken = true
	}
}

func AddSecretsVarsToReplacer(repl *caddy.Replacer) {
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

func AddAuthRequestVarsToReplacer(repl *caddy.Replacer, req *jwt.AuthorizationRequestClaims) {
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
