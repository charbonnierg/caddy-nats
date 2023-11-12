package natsauth

import (
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func NewPinnedTokens() *PinnedTokens {
	return &PinnedTokens{
		tokens:          make(map[string]*jwt.UserClaims),
		tokensByAccount: make(map[string]string),
		accountsByToken: make(map[string]string),
	}
}

type PinnedTokens struct {
	tokens          map[string]*jwt.UserClaims
	tokensByAccount map[string]string
	accountsByToken map[string]string
}

func (p *PinnedTokens) Add(account string, claims *jwt.UserClaims) (string, error) {
	if claims == nil {
		claims = &jwt.UserClaims{}
		claims.Audience = account
		claims.Limits = jwt.Limits{
			UserLimits: jwt.UserLimits{
				Src:    jwt.CIDRList{},
				Times:  nil,
				Locale: "",
			},
			NatsLimits: jwt.NatsLimits{
				Subs:    jwt.NoLimit,
				Data:    jwt.NoLimit,
				Payload: jwt.NoLimit,
			},
		}
	}
	sk, err := nkeys.CreateUser()
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %s", err.Error())
	}
	seed, err := sk.Seed()
	if err != nil {
		return "", err
	}
	token := string(seed)
	p.Remove(account)
	p.tokens[token] = claims
	p.tokensByAccount[account] = token
	p.accountsByToken[token] = account
	return token, nil
}

func (p *PinnedTokens) Remove(account string) {
	token, ok := p.tokensByAccount[account]
	if !ok {
		return
	}
	delete(p.tokens, token)
	delete(p.tokensByAccount, account)
	delete(p.accountsByToken, token)
}

func (p *PinnedTokens) Get(account string) (string, *jwt.UserClaims, bool) {
	token, ok := p.tokensByAccount[account]
	if !ok {
		return "", nil, false
	}
	claims, ok := p.tokens[token]
	if !ok {
		return "", nil, false
	}
	return token, claims, true
}

func (p *PinnedTokens) Lookup(token string) (string, *jwt.UserClaims, bool) {
	claims, ok := p.tokens[token]
	if !ok {
		return "", nil, false
	}
	account, ok := p.accountsByToken[token]
	if !ok {
		return "", nil, false
	}
	return account, claims, true
}
