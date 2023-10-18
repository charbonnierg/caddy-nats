package oauth2

import (
	"net/http"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/providers/util"
)

type OAuth2App interface {
	GetOrAddEndpoint(endpoint OAuth2Endpoint) (OAuth2Endpoint, error)
	GetEndpoint(name string) (OAuth2Endpoint, error)
}

type OAuth2Endpoint interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error
	DecodeSessionState(session []*http.Cookie) (*sessions.SessionState, error)
	DecodeSessionStateFromString(session string) (*sessions.SessionState, error)
	GetOidcSessionClaimExtractor(session *sessions.SessionState) (util.ClaimExtractor, error)
}
