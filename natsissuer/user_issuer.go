package natsissuer

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type UserIssuer struct {
	account       *jwt.AccountClaims
	accountJWT    string
	accountPub    string
	keypair       nkeys.KeyPair
	sub           string
	role          string
	AccountRaw    string `json:"account,omitempty"`
	SigningKeyRaw string `json:"signing_key,omitempty"`
	RoleRaw       string `json:"role,omitempty"`
}

func (u *UserIssuer) Provision(ctx caddy.Context) error {
	acc, err := jwt.DecodeAccountClaims(u.AccountRaw)
	if err != nil {
		return err
	}
	u.account = acc
	u.accountJWT = u.AccountRaw
	u.accountPub = acc.Subject
	u.sub = acc.Subject
	u.role = u.RoleRaw
	kp, err := nkeys.FromSeed([]byte(u.SigningKeyRaw))
	if err != nil {
		return err
	}
	u.keypair = kp
	return nil
}
