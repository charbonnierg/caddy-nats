// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream_fs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JetStreamFS{})
}

// CaddyModule implements the caddy.Module interface. It returns
// informations about the caddy module.
func (JetStreamFS) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.fs.jetstream",
		New: func() caddy.Module { return new(JetStreamFS) },
	}
}

// JetStreamFS implements the fs.FS interface.
// It can be used to serve files from a JetStream Object Store
// bucket. This is very experimental, and may change in the future.
type JetStreamFS struct {
	ctx     caddy.Context
	logger  *zap.Logger
	mutex   *sync.Mutex
	started bool

	store   nats.ObjectStore
	Store   string                 `json:"store,omitempty"`
	Account string                 `json:"account,omitempty"`
	Client  *natsclient.NatsClient `json:"client,omitempty"`
	SyncDir string                 `json:"sync_dir,omitempty"`
}

// Provision implements the caddy.Provisioner interface.
// It is called when module is loaded (on caddy startup or
// or config reload) and is responsible for registering the
// nats client connection used to communicate with JetStream
// engine.
func (f *JetStreamFS) Provision(ctx caddy.Context) error {
	f.ctx = ctx
	f.logger = ctx.Logger()
	f.mutex = &sync.Mutex{}
	if f.Client == nil {
		f.Client = &natsclient.NatsClient{Internal: true}
	}
	if err := caddynats.ProvisionClientConnection(ctx, f.Account, f.Client); err != nil {
		return err
	}
	return nil
}

// Open implements the fs.FS interface. It returns a fs.File interface
// for given file name.
// If file has already been downloaded and exists within the
// sync directory, file will be opened from the sync directory directly.
// If file does not exist in sync directory, this module will attempt
// to download file using nats client connection, and write the file
// under the sync directory, before opening the file from the sync directory.
func (f *JetStreamFS) Open(name string) (fs.File, error) {
	if !f.started {
		go f.startWatcher()
	}
	// Rewrite filename for root path
	if name == "." || name == "" || name == "/" {
		name = "index.html"
	}
	f.logger.Info("Checking if file exists on disk", zap.String("file", name))
	// Check if file exists on disk
	file, err := f.readFromCache(name)
	switch {
	// When no error, return opened file
	case err == nil:
		return file, nil
	// When error is something else that os.ErrNotExists
	// return error
	case err != os.ErrNotExist:
		return nil, err
	}
	// Download file from JetStream object store
	return f.download(name)
}

func (f *JetStreamFS) getStore() (nats.ObjectStore, error) {
	if f.store != nil {
		return f.store, nil
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.store != nil {
		return f.store, nil
	}
	f.logger.Info("Connecting to JetStream", zap.String("store", f.Store))
	js, err := f.Client.JetStreamContext()
	if err != nil {
		f.logger.Error("Failed to connect to JetStream", zap.String("store", f.Store), zap.Error(err))
		return nil, err
	}
	f.logger.Info("JetStream connection established", zap.String("store", f.Store))
	syncDir, err := os.Stat(f.SyncDir)
	if err != nil {
		f.logger.Info("Creating sync dir for jetstream_fs", zap.String("dir", f.SyncDir))
		if err := os.MkdirAll(f.SyncDir, 0755); err != nil {
			return nil, err
		}
	} else if !syncDir.IsDir() {
		return nil, errors.New("sync_dir must be a directory")
	}
	store, err := js.ObjectStore(f.Store)
	if err != nil {
		f.logger.Error("Failed to get object store", zap.String("store", f.Store), zap.Error(err))
		return nil, err
	}
	f.store = store
	return f.store, nil
}

func (f *JetStreamFS) startWatcher() {
	store, err := f.getStore()
	if f.started {
		return
	}
	f.started = true
	if err != nil {
		f.logger.Error("Failed to start JetStream watcher", zap.String("store", f.Store), zap.Error(err))
		return
	}
	f.logger.Info("Starting JetStream watcher", zap.String("store", f.Store))
	// Watch for changes on the store
	sub, err := store.Watch(nats.Context(f.ctx))
	if err != nil {
		f.logger.Error("Failed to start JetStream watcher", zap.String("store", f.Store), zap.Error(err))
		return
	}
	updates := sub.Updates()
	// Close subscription on exit
	defer sub.Stop()
	// Watch for changes
	for {
		select {
		case <-f.ctx.Done():
			f.logger.Warn("object store watcher stopped", zap.String("store", f.Store))
			return
		case update := <-updates:
			if update == nil {
				f.logger.Warn("object store finished rewinding most recent data", zap.String("store", f.Store))
				continue
			}
			switch update.Deleted {
			case true:
				f.logger.Info("File deleted", zap.String("store", f.Store), zap.String("file", update.Name))
				target := filepath.Join(f.SyncDir, update.Name)
				if err := os.Remove(target); err != nil {
					f.logger.Error("Failed to delete file", zap.String("store", f.Store), zap.String("file", update.Name), zap.Error(err))
					continue
				}
			case false:
				f.logger.Info("File updated", zap.String("store", f.Store), zap.String("file", update.Name))
				if _, err := f.download(update.Name); err != nil {
					f.logger.Error("Failed to download file", zap.String("store", f.Store), zap.String("file", update.Name), zap.Error(err))
					continue
				}
			}
		}
	}
}

func (f *JetStreamFS) readFromCache(name string) (fs.File, error) {
	f.logger.Info("reading from cache", zap.String("file", name))
	target := filepath.Join(f.SyncDir, name)
	_, err := os.Stat(target)
	if err == nil {
		return os.Open(target)
	}
	return nil, os.ErrNotExist
}

func (f *JetStreamFS) download(name string) (fs.File, error) {
	// Make sure store exists
	store, err := f.getStore()
	if err != nil {
		return nil, err
	}
	// Expect file to be new.
	target := filepath.Join(f.SyncDir, name)
	name = filepath.ToSlash(name)
	parent := filepath.Dir(target)
	f.logger.Info("Downloading file from store", zap.String("store", f.Store), zap.String("file", name), zap.String("destination", target))
	_, e := os.Stat(parent)
	// Create parent if it does not exist
	if e != nil {
		if err := os.MkdirAll(parent, 0700); err != nil {
			return nil, fmt.Errorf("failed to create parent directory for %s", parent)
		}
	}
	// Create file if it does not exist
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %s", target, err.Error())
	}
	// Fetch object from JetStream store
	result, err := store.Get(name)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, fmt.Errorf("failed to download file %s: %s", name, err.Error())
	}
	// Close object on exit
	defer result.Close()
	// Stream copy content to the opened file
	size, err := io.Copy(file, result)
	if err != nil {
		file.Close()
		f.logger.Error("failed to copy content to file", zap.String("file", file.Name()), zap.Error(err))
		return nil, err
	}
	f.logger.Info("Wrote to file", zap.String("file", file.Name()), zap.Int64("size", size))
	// Seed beginning of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	return file, nil
}

var (
	_ fs.FS                 = (*JetStreamFS)(nil)
	_ caddy.Module          = (*JetStreamFS)(nil)
	_ caddy.Provisioner     = (*JetStreamFS)(nil)
	_ caddyfile.Unmarshaler = (*JetStreamFS)(nil)
)
