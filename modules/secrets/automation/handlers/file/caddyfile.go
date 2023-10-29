package file

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
)

func (h *FileHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		notify := []json.RawMessage{}
		if d.NextArg() {
			h.File = d.Val()
			switch d.CountRemainingArgs() {
			case 1:
				if err := caddyutils.ParsePermissions(d, &h.FilePerm); err != nil {
					return err
				}
			default:
				return d.Err("too many arguments")
			}
		} else {
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "path":
					if err := caddyutils.ParseString(d, &h.File); err != nil {
						return err
					}
				case "no_create":
					if err := caddyutils.ParseBool(d, &h.NoCreate); err != nil {
						return err
					}
				case "no_create_parent":
					if err := caddyutils.ParseBool(d, &h.NoCreateParent); err != nil {
						return err
					}
				case "chmod", "file_perm":
					if err := caddyutils.ParsePermissions(d, &h.FilePerm); err != nil {
						return err
					}
				case "parent_chmod", "parent_perm":
					if err := caddyutils.ParsePermissions(d, &h.ParentPerm); err != nil {
						return err
					}
				case "notify":
					if !d.NextArg() {
						return d.Err("expected a notify type")
					}
					handlerType := d.Val()
					mod, err := caddyfile.UnmarshalModule(d, "secrets.handlers."+handlerType)
					if err != nil {
						return d.Errf("failed to unmarshal module 'secrets.handlers.%s': %v", handlerType, err)
					}
					notify = append(notify, caddyconfig.JSONModuleObject(mod, "type", handlerType, nil))
				default:
					return d.Errf("unknown file handler property '%s'", d.Val())
				}
				if len(notify) > 0 {
					if h.Notify == nil {
						h.Notify = []json.RawMessage{}
					}
					h.Notify = append(h.Notify, notify...)
				}
			}
		}

	}
	return nil
}
