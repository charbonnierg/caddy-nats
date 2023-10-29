// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package httputils

import (
	"net/http"
)

// JoinCookies takes a slice of cookies from the request and reconstructs the
// full session cookie
func JoinCookies(cookies []*http.Cookie, cookieName string) *http.Cookie {
	if len(cookies) == 0 {
		return &http.Cookie{}
	}
	if len(cookies) == 1 {
		return cookies[0]
	}
	c := CopyCookie(cookies[0])
	for i := 1; i < len(cookies); i++ {
		c.Value += cookies[i].Value
	}
	c.Name = cookieName
	return c
}

// copyCookie copies a cookie
func CopyCookie(c *http.Cookie) *http.Cookie {
	if c == nil {
		return nil
	}
	cp := *c
	return &cp
}

// ParseCookies parses a cookie string into a slice of cookies
func ParseCookies(value string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", value)
	request := http.Request{Header: header}
	return request.Cookies()
}
