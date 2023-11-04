package parser_test

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	. "github.com/onsi/gomega"
)

func NewTestDispenser(input string) *caddyfile.Dispenser {
	tokens, err := caddyfile.Tokenize([]byte(input), "test")
	Expect(err).NotTo(HaveOccurred())
	return caddyfile.NewDispenser(tokens)
}
