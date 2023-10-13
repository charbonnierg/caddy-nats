package oauth2

import (
	"errors"
	"net/http"
)

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
