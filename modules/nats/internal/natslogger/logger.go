// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package natslogger is a logger for nats-server based on zap sugared logger.
package natslogger

import (
	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"
)

// An implementation of server.Logger based on zap sugared logger
type Logger struct {
	*zap.SugaredLogger
	debug        bool
	trace        bool
	traceVerbose bool
}

// Tracef logs a trace (debug) statement
func (l Logger) Tracef(template string, args ...interface{}) {
	l.Debugf(template, args...)
}

// Noticef logs a notice (info) statement
func (l Logger) Noticef(template string, args ...interface{}) {
	l.Infof(template, args...)
}

// Attach attaches the logger to the given server
func (l Logger) Attach(srv *server.Server) {
	// No need to set flags if zap logger is not debug
	if l.Level() != zap.DebugLevel {
		l.debug = false
		l.trace = false
		l.traceVerbose = false
	}
	srv.SetLoggerV2(l, l.debug, l.trace, l.traceVerbose)
}

// New returns a new logger based from provided parent logger.
func New(parent *zap.Logger, opts *server.Options) *Logger {
	var debug = false
	var trace = false
	var traceVerbose = false
	switch {
	case opts.TraceVerbose:
		debug = true
		trace = true
		traceVerbose = true
	case opts.Trace:
		trace = true
		debug = true
	case opts.Debug:
		debug = true
	}
	return &Logger{parent.Sugar().WithOptions(zap.AddCallerSkip(4), zap.AddStacktrace(zap.DPanicLevel)), debug, trace, traceVerbose}
}

// NewLogger returns a new no-op logger
func NewNop() *Logger {
	return &Logger{zap.NewNop().Sugar(), false, false, false}
}

// NewDevelopment returns a new logger based from provided parent logger
func NewDevelopment(opts *server.Options, zapOption ...zap.Option) *Logger {
	logger, err := zap.NewDevelopment(zapOption...)
	if err != nil {
		panic(err)
	}
	return New(logger, opts)
}

// NewProduction returns a new logger based from provided parent logger
func NewProduction(opts *server.Options, zapOption ...zap.Option) *Logger {
	logger, err := zap.NewProduction(zapOption...)
	if err != nil {
		panic(err)
	}
	return New(logger, opts)
}
