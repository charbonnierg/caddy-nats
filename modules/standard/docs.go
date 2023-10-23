// SPDX-License-Identifier: Apache-2.0

// Package standard contains the standard Beyond modules.
// It can be imported by main packages to build custom Caddy binaries.
// Example usage:
//
//	  package main
//
//	  import (
//		  caddycmd "github.com/caddyserver/caddy/v2/cmd"
//		  // Standard caddy plugins
//		  _ "github.com/caddyserver/caddy/v2/modules/standard"
//		  // Standard beyond plugins
//		  _ "github.com/quara-dev/beyond/modules/standard"
//		  // plug in additional Caddy modules here
//		  // ...
//	  )
//
//	  func main() {
//		  caddycmd.Main()
//	  }
package standard
