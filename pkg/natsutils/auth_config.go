// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsutils

import (
	"go.uber.org/zap"
)

var (
	DEFAULT_AUTH_CALLOUT_SUBJECT = "$SYS.REQ.USER.AUTH"
	DEFAULT_AUTH_CALLOUT_ACCOUNT = "AUTH"
)

// Config is the configuration for an auth service.
type AuthServiceConfig struct {
	handler     authServiceHandlerFunc
	Subject     string
	Account     string
	SigningKey  string
	QueueGroup  string
	Name        string
	Version     string
	Description string
	Metadata    map[string]string
	Keystore    Keystore
	Logger      *zap.Logger
}

// NewConfig creates a new config with the given handler.
func NewAuthServiceConfig(handler authServiceHandlerFunc) *AuthServiceConfig {
	return &AuthServiceConfig{
		handler: handler,
		Subject: DEFAULT_AUTH_CALLOUT_SUBJECT,
		Account: DEFAULT_AUTH_CALLOUT_ACCOUNT,
	}
}
