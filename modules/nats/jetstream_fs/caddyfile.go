package jetstream_fs

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	natscaddyfile "github.com/quara-dev/beyond/modules/nats/caddyfile"
)

// UnmarshalCaddyfile parses the "jetstream_publish" directive from
// a Caddyfile dispenser.
// Syntax:
//
//	  fs jetstream {
//		connection <connection>
//	  	store <name>
//		sync_dir <path>
//	  }
//
// Example usage with a file_server:
//
//	  file_server {
//		   fs jetstream {
//		  	 connection local
//		  	 store docs
//			 sync_dir ./www/spa
//	    }
//	  }
func (f *JetStreamFS) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	if d.Val() != "jetstream" {
		return d.Errf("Expected 'jetstream' directive, got '%s'", d.Val())
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "store":
			store := ""
			if !d.AllArgs(&store) {
				return d.Err("invalid store")
			}
			f.Store = store
		case "sync_dir":
			syncDir := ""
			if !d.AllArgs(&syncDir) {
				return d.Err("invalid sync_dir")
			}
			f.SyncDir = syncDir
		case "connection":
			if err := natscaddyfile.ParseConnection(d, &f.Connection); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
