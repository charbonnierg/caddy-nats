package natsissuer

import (
	"encoding/json"
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func ParseIssuerOptions(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	app := new(App)
	app.keypairs = make(map[string]nkeys.KeyPair)
	if existingVal != nil {
		var ok bool
		caddyFileApp, ok := existingVal.(httpcaddyfile.App)
		if !ok {
			return nil, d.Errf("existing nats_issuer values of unexpected type: %T", existingVal)
		}
		err := json.Unmarshal(caddyFileApp.Value, app)
		if err != nil {
			return nil, err
		}
	}
	err := app.UnmarshalCaddyfile(d)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return httpcaddyfile.App{
		Name:  "nats.issuer",
		Value: caddyconfig.JSON(app, nil),
	}, err
}

func (app *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "operator":
				operators := d.RemainingArgs()
				app.Operators = append(app.Operators, operators...)
			case "system_account":
				if !d.AllArgs(&app.SystemAccount) {
					return d.ArgErr()
				}
			case "system_user":
				app.SystemUser = new(User)
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "token":
						if !d.AllArgs(&app.SystemUser.JWT) {
							return d.ArgErr()
						}
					case "nkey":
						if !d.AllArgs(&app.SystemUser.NKey) {
							return d.ArgErr()
						}
					default:
						return d.Errf("unrecognized nats_issuer system_user subdirective: %s", d.Val())
					}
				}
			case "accounts":
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					accName := d.Val()
					kp, err := nkeys.CreateAccount()
					if err != nil {
						return err
					}
					sub, err := kp.PublicKey()
					if err != nil {
						return err
					}
					acc := jwt.NewAccountClaims(sub)
					app.ProvisionAccounts = append(app.ProvisionAccounts, acc)
					acc.Name = accName
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "jetstream":
							acc.Limits.MemoryStorage = -1
							acc.Limits.DiskStorage = -1
							acc.Limits.Streams = -1
							acc.Limits.Consumer = -1
						case "max_connections":
							if !d.NextArg() {
								return d.ArgErr()
							}
							conn, err := strconv.Atoi(d.Val())
							if err != nil {
								return err
							}
							acc.Limits.Conn = int64(conn)
						case "max_payload":
							if !d.NextArg() {
								return d.ArgErr()
							}
							payload, err := strconv.Atoi(d.Val())
							if err != nil {
								return err
							}
							acc.Limits.Payload = int64(payload)
						case "role":
							if !d.NextArg() {
								return d.ArgErr()
							}
							roleName := d.Val()
							nk, err := nkeys.CreateAccount()
							if err != nil {
								return err
							}
							pub, err := nk.PublicKey()
							app.keypairs[pub] = nk
							if err != nil {
								return err
							}
							role := jwt.NewUserScope()
							role.Role = roleName
							role.Key = pub
							for nesting := d.Nesting(); d.NextBlock(nesting); {
								switch d.Val() {
								case "limits":
									for nesting := d.Nesting(); d.NextBlock(nesting); {
										switch d.Val() {
										case "max_payload":
											if !d.NextArg() {
												return d.ArgErr()
											}
											payload, err := strconv.Atoi(d.Val())
											if err != nil {
												return err
											}
											role.Template.Payload = int64(payload)
										case "max_data":
											if !d.NextArg() {
												return d.ArgErr()
											}
											conn, err := strconv.Atoi(d.Val())
											if err != nil {
												return err
											}
											role.Template.Data = int64(conn)
										}
									}
								case "publish":
									for nesting := d.Nesting(); d.NextBlock(nesting); {
										switch d.Val() {
										case "allow":
											allowSubjects := d.RemainingArgs()
											role.Template.Permissions.Pub.Allow.Add(allowSubjects...)
										case "deny":
											denySubjects := d.RemainingArgs()
											role.Template.Permissions.Pub.Deny.Add(denySubjects...)
										default:
											return d.Errf("unrecognized nats_issuer role subscribe subdirective: %s", d.Val())
										}
									}
								case "subscribe":
									for nesting := d.Nesting(); d.NextBlock(nesting); {
										switch d.Val() {
										case "allow":
											allowSubjects := d.RemainingArgs()
											role.Template.Permissions.Sub.Allow.Add(allowSubjects...)
										case "deny":
											denySubjects := d.RemainingArgs()
											role.Template.Permissions.Sub.Deny.Add(denySubjects...)
										default:
											return d.Errf("unrecognized nats_issuer role subscribe subdirective: %s", d.Val())
										}
									}
								}
							}
							acc.SigningKeys.AddScopedSigner(role)
						}
					}
				}
			default:
				return d.Errf("unrecognized nats_issuer subdirective: %s", d.Val())
			}
		}
	}
	return nil
}
