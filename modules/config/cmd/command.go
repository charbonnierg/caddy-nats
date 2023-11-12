// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/spf13/cobra"
)

func init() {
	caddycmd.RegisterCommand(caddycmd.Command{
		Name:  "config",
		Usage: "[--help] [--host <url>]",
		Short: "Interact with config.",
		Long:  `A collection of command line tools to work with config.`,
		CobraFunc: func(cmd *cobra.Command) {
			// Add global  options
			cmd.PersistentFlags().StringP("host", "H", "http://localhost:2019", "Host to connect to")
			// Add config get subcommand using inline function
			cmd.AddCommand(func() *cobra.Command {
				// Create subcommand
				child := cobra.Command{
					Use:     "get [key]",
					Short:   "Get a config value",
					Example: `caddy config get`,
					RunE:    caddycmd.WrapCommandFuncForCobra(getCmd),
				}
				return &child
			}())
			// Add config set subcommand using inline function
			cmd.AddCommand(func() *cobra.Command {
				// Create subcommand
				child := cobra.Command{
					Use:   "set [key] [value]",
					Short: "Set a config value",
					RunE:  caddycmd.WrapCommandFuncForCobra(setCmd),
				}
				return &child
			}())
			// Add config set subcommand using inline function
			cmd.AddCommand(func() *cobra.Command {
				// Create subcommand
				child := cobra.Command{
					Use:   "update [key] [value]",
					Short: "Update a config value",
					RunE:  caddycmd.WrapCommandFuncForCobra(updateCmd),
				}
				return &child
			}())
		},
	})
}
