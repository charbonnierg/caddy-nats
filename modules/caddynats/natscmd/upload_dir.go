// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/nats-io/nats.go"
)

func uploadDirCmd(fs caddycmd.Flags) (int, error) {
	create := fs.Bool("create")
	directory := fs.Arg(0)
	if directory == "" {
		return 1, errors.New("missing directory postitional argument")
	}
	store := fs.String("store")
	if store == "" {
		return 1, errors.New("missing store option")
	}
	client, err := connect(fs)
	if err != nil {
		return 1, err
	}
	js, _ := client.JetStreamContext()
	nc, _ := client.Nats()
	defer nc.Close()
	objects, err := js.ObjectStore(store)
	if err != nil {
		switch {
		case err == nats.ErrStreamNotFound && create:
			objects, err = js.CreateObjectStore(&nats.ObjectStoreConfig{Bucket: store})
			if err != nil {
				return 1, fmt.Errorf("failed to create store %s: %s", store, err.Error())
			}
		default:
			return 1, fmt.Errorf("store does not exist: %s", store)
		}
	}
	totalBytes := uint64(0)
	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		size, err := processFile(objects, directory, path, info)
		if err == nil {
			totalBytes = totalBytes + size
		}
		return err
	}); err != nil {
		return 1, err
	}
	fmt.Printf("total uploaded size: %d\n", totalBytes)
	return 0, nil
}

func processFile(store nats.ObjectStore, root string, path string, info os.FileInfo) (uint64, error) {
	if info.IsDir() {
		return 0, nil
	}
	sanitized, err := filepath.Rel(root, path)
	if err != nil {
		return 0, err
	}
	sanitized = strings.Replace(sanitized, `\`, "/", -1)
	fmt.Printf("uploading file: %s (size=%d)\n", sanitized, info.Size())
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	_info, err := store.Put(&nats.ObjectMeta{Name: sanitized}, f)
	if err != nil {
		return 0, err
	}
	return _info.Size, err
}
