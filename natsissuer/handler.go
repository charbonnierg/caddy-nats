package natsissuer

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Handler{})
	httpcaddyfile.RegisterHandlerDirective("nats_issuer", parseIssuerHandler)
}

type Handler struct {
	app        *App
	logger     *zap.Logger
	accountJWT string
	Account    string `json:"account,omitempty"`
	Role       string `json:"role,omitempty"`
}

func (Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.nats_issuer",
		New: func() caddy.Module { return new(Handler) },
	}
}

func (h *Handler) Provision(ctx caddy.Context) error {
	h.logger = ctx.Logger()
	issuerAppIface, err := ctx.App("nats.issuer")
	if err != nil {
		return fmt.Errorf("getting nats.issuer app: %v. Make sure nats issuer is configured in global options", err)
	}

	h.app = issuerAppIface.(*App)
	h.accountJWT, err = h.app.GetAccount(h.Account)
	if err != nil {
		return fmt.Errorf("getting account: %v", err)
	}
	return nil
}

func (h Handler) HandleGetJWT(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(h.accountJWT))
}

func (h Handler) HandlerPostUserCreds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	issuer, err := h.app.GetIssuer(h.Account, h.Role)
	if err != nil {
		h.logger.Error("error getting issuer", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	nk, err := nkeys.CreateUser()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	seed, err := nk.Seed()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	subject, err := nk.PublicKey()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user := jwt.NewUserClaims(subject)
	user.Name = "test"
	user.IssuerAccount = issuer.accountPub
	token, err := user.Encode(issuer.keypair)
	creds, err := jwt.FormatUserConfig(token, seed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(creds)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if r.Method == http.MethodGet {
		h.HandleGetJWT(w, r)
	} else if r.Method == http.MethodPost {
		h.HandlerPostUserCreds(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
	return next.ServeHTTP(w, r)
}

func parseIssuerHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	p := Handler{}
	err := p.UnmarshalCaddyfile(h.Dispenser)
	return p, err
}

func (h *Handler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		account := d.Val()
		h.Account = account
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "role":
				if !d.AllArgs(&h.Role) {
					return d.ArgErr()
				}
			default:
				return d.Errf("unrecognized subdirective: %s", d.Val())
			}
		}
	}

	return nil
}

var (
	_ caddyhttp.MiddlewareHandler = (*Handler)(nil)
	_ caddy.Provisioner           = (*Handler)(nil)
	_ caddyfile.Unmarshaler       = (*Handler)(nil)
)
