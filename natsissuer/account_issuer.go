package natsissuer

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type AccountIssuer struct {
	app           *App
	operator      *jwt.OperatorClaims
	operatorJWT   string
	operatorPub   string
	keypair       nkeys.KeyPair
	sub           string
	seed          string
	OperatorRaw   string `json:"operator,omitempty"`
	SigningKeyRaw string `json:"signing_key,omitempty"`
}

func (a *AccountIssuer) Provision(ctx caddy.Context) error {
	if a.OperatorRaw == "internal" {
		operatorClaims, err := jwt.DecodeOperatorClaims(a.app.internalOperator)
		if err != nil {
			return err
		}
		seed, err := a.app.internalOperatorKeypair.Seed()
		if err != nil {
			return err
		}
		a.keypair = a.app.internalOperatorKeypair
		a.operator = operatorClaims
		a.operatorJWT = a.app.internalOperator
		a.seed = string(seed)
		a.sub, err = a.keypair.PublicKey()
		if err != nil {
			return err
		}
		return nil
	}
	operatorClaims, err := jwt.DecodeOperatorClaims(a.OperatorRaw)
	if err != nil {
		return err
	}
	nk, err := nkeys.FromSeed([]byte(a.SigningKeyRaw))
	if err != nil {
		return err
	}
	_, err = nk.PublicKey()
	if err != nil {
		return err
	}
	a.keypair = nk
	a.operator = operatorClaims
	a.operatorJWT = a.OperatorRaw
	a.operatorPub = a.operator.Subject
	a.sub, err = a.keypair.PublicKey()
	if err != nil {
		return err
	}
	return nil
}

func (a *AccountIssuer) CreateAccount(name string) (string, error) {
	keys, err := nkeys.CreateAccount()
	if err != nil {
		return "", err
	}
	sub, err := keys.PublicKey()
	if err != nil {
		return "", err
	}
	claims := jwt.NewAccountClaims(sub)
	claims.Issuer = a.sub
	claims.Name = name
	claims.Limits.Data = -1
	claims.Limits.MemoryStorage = -1
	claims.Limits.DiskStorage = -1
	claims.Limits.Subs = -1
	token, err := claims.Encode(a.keypair)
	if err != nil {
		return "", err
	}
	return token, nil
}
