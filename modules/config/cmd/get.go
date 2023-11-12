// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"net/http"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
)

func getCmd(fs caddycmd.Flags) (int, error) {
	host := fs.String("host")
	key := fs.Arg(0)
	cfg, err := getConfigPath(host, key)
	if err != nil {
		return 1, err
	}
	fmt.Println(cfg)
	return 0, nil
}

func getConfigPath(host string, key string) (string, error) {
	requestUrl := fmt.Sprintf("%s/config/%s", host, key)
	response, err := http.Get(requestUrl)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("error getting config: %s", string(body))
	}
	return string(body), nil
}
