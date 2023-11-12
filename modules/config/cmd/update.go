// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
)

func updateCmd(fs caddycmd.Flags) (int, error) {
	host := fs.String("host")
	key := fs.Arg(0)
	value := fs.Arg(1)
	err := updateConfig(host, key, value)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func updateConfig(host string, key string, value string) error {
	requestUrl := fmt.Sprintf("%s/config/%s", host, key)
	body := []byte(value)
	bodyReader := bytes.NewReader(body)
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", requestUrl, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("error updating config: %s", string(responseBody))
	}
	return nil
}
