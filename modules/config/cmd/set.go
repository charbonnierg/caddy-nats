package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
)

func setCmd(fs caddycmd.Flags) (int, error) {
	host := fs.String("host")
	key := fs.Arg(0)
	value := fs.Arg(1)
	err := setConfig(host, key, value)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func setConfig(host string, key string, value string) error {
	requestUrl := fmt.Sprintf("%s/config/%s", host, key)
	body := []byte(value)
	bodyReader := bytes.NewReader(body)
	response, err := http.Post(requestUrl, "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("error setting config: %s", string(responseBody))
	}
	return nil
}
