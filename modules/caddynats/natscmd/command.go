package cmd

import (
	"fmt"
	"math"
	"os"
	"strings"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/spf13/cobra"
)

func init() {
	// Register the "caddy nats" subcommand
	caddycmd.RegisterCommand(caddycmd.Command{
		Name:  "nats",
		Usage: "[--help] [--server <url>] [--creds <creds>]",
		Short: "Interact with NATS servers.",
		Long:  `A collection of command line tools to work with NATS.`,
		CobraFunc: func(cmd *cobra.Command) {
			// Add global  options
			cmd.PersistentFlags().StringP("server", "s", "nats://localhost:4222", "Server to connect to")
			cmd.PersistentFlags().StringP("name", "n", "", "Client name used connect to")
			cmd.PersistentFlags().StringP("user", "u", "", "User name used connect to")
			cmd.PersistentFlags().StringP("password", "p", "", "Password used connect to")
			cmd.PersistentFlags().StringP("creds", "c", "", "Path to credential file used to connect")
			cmd.PersistentFlags().StringP("token", "t", "", "Token used to connect")
			cmd.PersistentFlags().StringP("seed", "S", "", "Nkey seed used to connect")
			cmd.PersistentFlags().StringP("js-domain", "D", "", "JetStream domain used to connect")
			cmd.PersistentFlags().String("js-prefix", "", "JetStream prefix used to connect")
			cmd.PersistentFlags().StringP("jwt", "j", "", "JWT used to connect")
			cmd.PersistentFlags().StringP("inbox-prefix", "I", "", "Inbox prefix for the client")
			// Add nats pub subcommand using inline function
			cmd.AddCommand(func() *cobra.Command {
				// Create subcommand
				child := cobra.Command{
					Use:   "pub <subject> [payload] [--header key=value]",
					Short: "Publish a message on NATS",
					Example: `
caddy nats pub some.subject
caddy nats pub some.subject "some message"
caddy nats pub some.subject "some message" -H key=value
caddy nats pub some.subject "some message" -H key=value --count 1000
`,
					RunE: caddycmd.WrapCommandFuncForCobra(pubCmd),
				}
				// Register subcommand flags
				child.Flags().StringArrayP("header", "H", nil, "Headers formatted as key=value pair to send with message")
				child.Flags().String("reply", "", "Sets a custom reply to subject")
				child.Flags().Int("count", 1, "Publish multiple messages")
				child.Flags().Duration("sleep", 0, "When publishing multiple messages, sleep between publishes")
				return &child
			}())
			// Add nats sub subcommand using inline function
			cmd.AddCommand(func() *cobra.Command {
				// Create subcommand
				child := cobra.Command{
					Use:     "sub <subject> [--queue <queue>]",
					Short:   "Subscribe to a NATS subject",
					Example: "caddy nats sub some.subject",
					RunE:    caddycmd.WrapCommandFuncForCobra(subCmd),
				}
				// Register subcommand flags
				child.Flags().String("queue", "", "Queue group to join")
				return &child
			}())
			// Add nats watch subcommand using inline function
			cmd.AddCommand(func() *cobra.Command {
				// Create subcommand
				child := cobra.Command{
					Use:     "watch <stream> [--filter <filter>]",
					Short:   "Watch a JetStream stream",
					Example: "caddy nats watch some.stream",
					RunE:    caddycmd.WrapCommandFuncForCobra(watchCmd),
				}
				// Register subcommand flags
				child.Flags().StringArray("filter", []string{}, "Filter subjects to watch")
				return &child
			}())
			// Add nats upload-dir subcommand using inline function
			cmd.AddCommand(func() *cobra.Command {
				// Create subcommand
				child := cobra.Command{
					Use:     "upload-dir --store=<store> <directory>",
					Short:   "Upload directory content to JetStream object store",
					Example: "upload-dir --store docs ./build",
					RunE:    caddycmd.WrapCommandFuncForCobra(uploadDirCmd),
				}
				// Register subcommand flag
				child.Flags().String("store", "", "JetStream store to upload files to")
				child.Flags().Bool("create", false, "Create bucket in JetStream object store if it does not exist yet")
				child.Flags().String("prefix", "", "A prefix to add to filenames")
				child.Flags().Bool("windows-path", false, `Convert filenames to Windows path (using "\" instead of "/")`)
				return &child
			}())
		},
	})
}

func connect(fs caddycmd.Flags) (*natsclient.NatsClient, error) {
	var servers []string
	if server := fs.String("server"); server != "" {
		servers = strings.Split(server, ",")
	}
	client := natsclient.NatsClient{
		Name:        fs.String("name"),
		Username:    fs.String("user"),
		Password:    fs.String("password"),
		Token:       fs.String("token"),
		Credentials: fs.String("creds"),
		Jwt:         fs.String("jwt"),
		Seed:        fs.String("seed"),
		JSDomain:    fs.String("js-domain"),
		JSPrefix:    fs.String("js-prefix"),
		InboxPrefix: fs.String("inbox-prefix"),
		Servers:     servers,
	}
	err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %s", err.Error())
	}
	return &client, nil
}

func parseCount(fs caddycmd.Flags) int {
	count := fs.Int("count")
	if count < 0 {
		count = math.MaxInt32
	}
	return count
}

func parsePayload(fs caddycmd.Flags, pos int) ([]byte, error) {
	value := fs.Arg(pos)
	switch {
	case value == "":
		return []byte{}, nil
	case value[0] == "@"[0]:
		path := value[1:]
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read payload from file %s: %s", path, err.Error())
		}
		return content, nil
	default:
		return []byte(value), nil
	}
}

func parseHeaders(fs caddycmd.Flags) (nats.Header, error) {
	rawHeaders, err := fs.GetStringArray("header")
	if err != nil {
		return nil, fmt.Errorf("failed to parse header: %s", err.Error())
	}
	natsHeaders := make(nats.Header)
	for _, v := range rawHeaders {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid header format %s: (expected 2 parts, got %d)", v, len(kv))
		}
		_, exists := natsHeaders[kv[0]]
		if !exists {
			natsHeaders[kv[0]] = []string{}
		}
		natsHeaders[kv[0]] = append(natsHeaders[kv[0]], kv[1])
	}
	return natsHeaders, nil
}
