// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(FileHandler{})
}

// FileHandler is a handler that saves the secret vlaue to a file.
type FileHandler struct {
	notifyHandlers []Handler
	logger         *zap.Logger
	// This is the path to the file to save the secret to
	File           string            `json:"file,omitempty"`
	FilePerm       fs.FileMode       `json:"file_perm,omitempty"`
	NoCreate       bool              `json:"no_create,omitempty"`
	NoCreateParent bool              `json:"no_create_parent,omitempty"`
	ParentPerm     fs.FileMode       `json:"parent_perm,omitempty"`
	Notify         []json.RawMessage `json:"notify,omitempty" caddy:"namespace=secrets.handlers inline_key=type"`
}

func (FileHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.handlers.file",
		New: func() caddy.Module { return new(FileHandler) },
	}
}

func (h *FileHandler) Provision(automation *Automation) error {
	h.logger = automation.ctx.Logger().Named("file_handler")
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
	} else if !stat.IsDir() {
		return fmt.Errorf("parent directory is not a directory: %s", parent)
	}
	unm, err := automation.ctx.LoadModule(h, "Notify")
	if err != nil {
		return err
	}
	for _, raw := range unm.([]interface{}) {
		notify, ok := raw.(Handler)
		if !ok {
			return fmt.Errorf("invalid notify handler type")
		}
		if err := notify.Provision(automation); err != nil {
			return err
		}
		h.notifyHandlers = append(h.notifyHandlers, notify)
	}
	if h.NoCreate {
		_, err = os.Stat(h.File)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to stat file: %w", err)
		}
		if err != nil {
			return fmt.Errorf("file does not exist: %s", h.File)
		}
	}
	return nil
}

func (h *FileHandler) Handle(value string) (string, error) {
	var err error
	if h.FilePerm == 0 {
		h.FilePerm = 0600
	}
	currentValue, err := os.ReadFile(h.File)
	// Exit early if the file already contains the secret value
	if err == nil && string(currentValue) == value {
		h.logger.Info("skipping write to file because secret value did not change")
		return h.File, nil
	}
	// Write the secret value to the file
	err = os.WriteFile(h.File, []byte(value), h.FilePerm)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}
	// Notify handlers
	for _, notifier := range h.notifyHandlers {
		h.logger.Info("notifying handler", zap.String("handler", notifier.CaddyModule().ID.Name()))
		if _, err := notifier.Handle(h.File); err != nil {
			return "", fmt.Errorf("failed to notify handler: %w", err)
		}
	}
	return h.File, nil
}
