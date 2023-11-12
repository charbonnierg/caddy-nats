// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func WrapLogger(logger *zap.Logger) (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(logger.Level())
	return logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core { return core })), logs
}

func NewLogger(options ...zap.Option) (*zap.Logger, *observer.ObservedLogs) {
	return WrapLogger(zap.NewExample(options...))
}
