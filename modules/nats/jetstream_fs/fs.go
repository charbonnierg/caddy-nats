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
	caddynats "github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/modules/nats/client"
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
	logger     *zap.Logger
	mutex      *sync.Mutex
	store      nats.ObjectStore
	Store      string             `json:"store,omitempty"`
	Connection *client.Connection `json:"connection,omitempty"`
	SyncDir    string             `json:"sync_dir,omitempty"`
}

// Open implements the fs.FS interface. It returns a fs.File interface
// for given file name.
// If file has already been downloaded and exists within the
// sync directory, file will be opened from the sync directory directly.
// If file does not exist in sync directory, this module will attempt
// to download file using nats client connection, and write the file
// under the sync directory, before opening the file from the sync directory.
func (f *JetStreamFS) Open(name string) (fs.File, error) {
	// Rewrite filename for root path
	if name == "." || name == "" || name == "/" {
		name = "index.html"
	}
	f.logger.Info("Opening file", zap.String("file", name))
	// Make sure store exists
	f.setup()
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

func (f *JetStreamFS) setup() error {
	if f.store != nil {
		return nil
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.store != nil {
		return nil
	}
	f.logger.Info("Connecting to JetStream", zap.String("store", f.Store))
	js, err := f.Connection.JetStream()
	if err != nil {
		f.logger.Error("Failed to connect to JetStream", zap.String("store", f.Store), zap.Error(err))
		return err
	}
	f.logger.Info("JetStream connection established", zap.String("store", f.Store))
	syncDir, err := os.Stat(f.SyncDir)
	if err != nil {
		f.logger.Info("Creating sync dir for jetstream_fs", zap.String("dir", f.SyncDir))
		if err := os.MkdirAll(f.SyncDir, 0755); err != nil {
			return err
		}
	} else if !syncDir.IsDir() {
		return errors.New("sync_dir must be a directory")
	}
	store, err := js.ObjectStore(f.Store)
	if err != nil {
		f.logger.Error("Failed to get object store", zap.String("store", f.Store), zap.Error(err))
		return err
	}
	f.store = store
	return nil
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
	f.logger.Info("downloading", zap.String("file", name))
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
	result, err := f.store.Get(name)
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

// Provision implements the caddy.Provisioner interface.
// It is called when module is loaded (on caddy startup or
// or config reload) and is responsible for registering the
// nats client connection used to communicate with JetStream
// engine.
func (f *JetStreamFS) Provision(ctx caddy.Context) error {
	f.mutex = &sync.Mutex{}
	app, err := caddynats.Load(ctx)
	if err != nil {
		return err
	}
	f.logger = ctx.Logger()
	f.logger.Info("Provisioning jetstream_fs connection")
	if err := f.Connection.Provision(app); err != nil {
		return err
	}
	return nil
}

var (
	_ fs.FS                 = (*JetStreamFS)(nil)
	_ caddy.Module          = (*JetStreamFS)(nil)
	_ caddy.Provisioner     = (*JetStreamFS)(nil)
	_ caddyfile.Unmarshaler = (*JetStreamFS)(nil)
)
