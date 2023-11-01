package basicauth

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
)

func init() {
	caddy.RegisterModule(BasicAuthExtension{})
}

type Users map[string]string

func (u Users) inline() string {
	elems := make([]string, 0, len(u))
	for k, v := range u {
		elems = append(elems, fmt.Sprintf("%s:%s", k, v))
	}
	return strings.Join(elems, "\n")
}

type htpasswd struct {
	File   string `json:"file,omitempty"`
	Inline string `json:"inline,omitempty"`
}

type ServerAuth struct {
	File  string `json:"file,omitempty"`
	Users Users  `json:"users,omitempty"`
}

func (s *ServerAuth) MarshalJSON() ([]byte, error) {
	if s.File != "" && s.Users != nil {
		return nil, fmt.Errorf("cannot use both file and users")
	}
	return json.Marshal(htpasswd{
		File:   s.File,
		Inline: s.Users.inline(),
	})
}

type ClientAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type BasicAuthExtension struct {
	ClientAuth *ClientAuth `json:"client_auth,omitempty"`
	Htpasswd   *ServerAuth `json:"htpasswd,omitempty"`
}

func (b *BasicAuthExtension) Clone() *BasicAuthExtension {
	other := &BasicAuthExtension{}
	if b.ClientAuth != nil {
		otherAuth := *b.ClientAuth
		other.ClientAuth = &otherAuth
	}
	if b.Htpasswd != nil {
		otherHtpasswd := *b.Htpasswd
		other.Htpasswd = &otherHtpasswd
	}
	return other
}

func (BasicAuthExtension) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.extensions.basicauth",
		New: func() caddy.Module { return new(BasicAuthExtension) },
	}
}

func (e *BasicAuthExtension) Marshal(repl *caddy.Replacer) ([]byte, error) {
	ext := e.Clone()
	if ext.ClientAuth != nil {
		pwd, err := repl.ReplaceOrErr(ext.ClientAuth.Password, true, true)
		if err != nil {
			return nil, err
		}
		ext.ClientAuth.Password = pwd
		usr, err := repl.ReplaceOrErr(ext.ClientAuth.Username, true, true)
		if err != nil {
			return nil, err
		}
		ext.ClientAuth.Username = usr
	}
	return json.Marshal(ext)
}

// Interface guards
var (
	_ config.Extension = (*BasicAuthExtension)(nil)
)
