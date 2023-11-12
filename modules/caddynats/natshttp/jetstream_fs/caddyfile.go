// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream_fs

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/pkg/fnutils"
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
		case "account":
			account := ""
			if !d.AllArgs(&account) {
				return d.Err("invalid account")
			}
			f.Account = account
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
		case "client":
			f.Client = fnutils.DefaultIfNil(f.Client, &natsclient.NatsClient{})
			if err := f.Client.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
