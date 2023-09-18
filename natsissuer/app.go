package natsissuer

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"go.uber.org/zap"
)

// Register caddy module when file is imported
func init() {
	caddy.RegisterModule(App{})
	httpcaddyfile.RegisterGlobalOption("nats_issuer", ParseIssuerOptions)
}

type User struct {
	JWT  string `json:"token,omitempty"`
	NKey string `json:"nkey,omitempty"`
}

type App struct {
	ctx                     caddy.Context
	keypairs                map[string]nkeys.KeyPair
	internalOperator        string
	internalOperatorKeypair nkeys.KeyPair
	internalAccountIssuer   *AccountIssuer
	logger                  *zap.Logger
	Operators               []string             `json:"operators,omitempty"`
	SystemAccount           string               `json:"system_account,omitempty"`
	SystemUser              *User                `json:"system_user,omitempty"`
	Accounts                []string             `json:"accounts,omitempty"`
	ProvisionAccounts       []*jwt.AccountClaims `json:"provision_accounts,omitempty"`
	Issuers                 []*UserIssuer        `json:"issuers,omitempty"`
}

func (a *App) GetAccount(account string) (string, error) {
	for _, acc := range a.Accounts {
		accClaims, err := jwt.DecodeAccountClaims(acc)
		if err != nil {
			return "", err
		}
		if accClaims.Subject == account {
			return acc, nil
		}
		if accClaims.Name == account {
			return acc, nil
		}
	}
	return "", errors.New("account not found")
}

func (a *App) GetIssuer(account string, role string) (*UserIssuer, error) {

	for _, iss := range a.Issuers {
		if (iss.account.Subject == account || iss.account.Name == account || iss.accountJWT == account) && iss.RoleRaw == role {
			return iss, nil
		}
	}
	return nil, fmt.Errorf("issuer not found: %s/%s", account, role)
}

func (a *App) SetInternalOperator() error {
	kp, err := nkeys.CreateOperator()
	if err != nil {
		return err
	}
	sub, err := kp.PublicKey()
	if err != nil {
		return err
	}
	op := jwt.NewOperatorClaims(sub)
	op.Name = "internal"
	sys, err := a.GetSystemAccountPublicKey()
	if err != nil {
		return err
	}
	op.SystemAccount = sys
	token, err := op.Encode(kp)
	if err != nil {
		return err
	}
	a.internalOperator = token
	a.Operators = append(a.Operators, token)
	a.internalOperatorKeypair = kp
	a.internalAccountIssuer = &AccountIssuer{app: a, OperatorRaw: "internal"}
	if err := a.internalAccountIssuer.Provision(a.ctx); err != nil {
		return fmt.Errorf("provisioning internal account issuer: %s", err.Error())
	}
	return nil
}

func (a *App) GetSystemAccountPublicKey() (string, error) {
	claims, err := jwt.DecodeAccountClaims(a.SystemAccount)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.issuer",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()
	a.logger.Info("Provisioning NATS issuer")
	err := a.SetInternalOperator()
	if err != nil {
		return err
	}
	for _, acc := range a.ProvisionAccounts {
		accJWT, err := acc.Encode(a.internalOperatorKeypair)
		if err != nil {
			return err
		}
		a.Accounts = append(a.Accounts, accJWT)
		for _, scope := range acc.SigningKeys {
			if reflect.TypeOf(scope) == reflect.TypeOf(jwt.UserScope{}) {
				sc := scope.(jwt.UserScope)
				kp, ok := a.keypairs[sc.Key]
				if !ok {
					return errors.New("keypair not found")
				}
				seed, err := kp.Seed()
				if err != nil {
					return err
				}
				a.Issuers = append(a.Issuers, &UserIssuer{
					AccountRaw:    accJWT,
					SigningKeyRaw: string(seed),
					RoleRaw:       sc.Role,
				})
			}
		}
	}
	for _, iss := range a.Issuers {
		if err := iss.Provision(a.ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Validate() error {
	return nil
}

func (a *App) Start() error {
	return nil
}

func (a *App) Stop() error {
	return nil
}

// Interface guards
var (
	_ caddy.Module      = (*App)(nil)
	_ caddy.Provisioner = (*App)(nil)
	_ caddy.App         = (*App)(nil)
	_ caddy.Validator   = (*App)(nil)
)
