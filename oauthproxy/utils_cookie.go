// SPDX-License-Identifier: Apache-2.0

package oauthproxy

import (
	"errors"
	"fmt"
	"net/http"
)

// joinCookies takes a slice of cookies from the request and reconstructs the
// full session cookie
func joinCookies(cookies []*http.Cookie, cookieName string) (*http.Cookie, error) {
	if len(cookies) == 0 {
		return nil, fmt.Errorf("list of cookies must be > 0")
	}
	if len(cookies) == 1 {
		return cookies[0], nil
	}
	c := copyCookie(cookies[0])
	for i := 1; i < len(cookies); i++ {
		c.Value += cookies[i].Value
	}
	c.Name = cookieName
	return c, nil
}

// copyCookie copies a cookie
func copyCookie(c *http.Cookie) *http.Cookie {
	return &http.Cookie{
		Name:       c.Name,
		Value:      c.Value,
		Path:       c.Path,
		Domain:     c.Domain,
		Expires:    c.Expires,
		RawExpires: c.RawExpires,
		MaxAge:     c.MaxAge,
		Secure:     c.Secure,
		HttpOnly:   c.HttpOnly,
		Raw:        c.Raw,
		Unparsed:   c.Unparsed,
		SameSite:   c.SameSite,
	}
}

// parseCookies parses a cookie string into a slice of cookies
func parseCookies(value string) ([]*http.Cookie, error) {
	header := http.Header{}
	header.Add("Cookie", value)
	request := http.Request{Header: header}
	cookies := request.Cookies()
	if len(cookies) == 0 {
		return nil, errors.New("list of cookies must be > 0")
	}
	return cookies, nil
}
