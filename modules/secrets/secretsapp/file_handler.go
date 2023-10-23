package secretsapp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/caddyserver/caddy/v2"
)

func init() {
	caddy.RegisterModule(FileHandler{})
}

// FileHandler is a handler that saves the secret vlaue to a file.
type FileHandler struct {
	// This is the path to the file to save the secret to
	File           string      `json:"file,omitempty"`
	FilePerm       fs.FileMode `json:"file_perm,omitempty"`
	NoCreate       bool        `json:"no_create,omitempty"`
	NoCreateParent bool        `json:"no_create_parent,omitempty"`
	ParentPerm     fs.FileMode `json:"parent_perm,omitempty"`
}

func (FileHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.handlers.file",
		New: func() caddy.Module { return new(FileHandler) },
	}
}

func (h *FileHandler) Provision(automation *Automation) error {
	parent := filepath.Dir(h.File)
	stat, err := os.Stat(parent)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to stat parent directory: %w", err)
	}
	if err != nil && (h.NoCreateParent || h.NoCreate) {
		return fmt.Errorf("parent directory does not exist: %s", parent)
	}
	if err != nil {
		if h.ParentPerm == 0 {
			h.ParentPerm = 0600
		}
		err := os.Mkdir(parent, h.ParentPerm)
		if err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		return nil
	}
	if !stat.IsDir() {
		return fmt.Errorf("parent directory is not a directory: %s", parent)
	}
	if !h.NoCreate {
		return nil
	}
	_, err = os.Stat(h.File)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if err != nil {
		return fmt.Errorf("file does not exist: %s", h.File)
	}
	return nil
}

func (h *FileHandler) Handle(value string) (string, error) {
	if h.FilePerm == 0 {
		h.FilePerm = 0600
	}
	err := os.WriteFile(h.File, []byte(value), h.FilePerm)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}
	return h.File, nil
}
