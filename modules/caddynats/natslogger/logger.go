// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package natslogger is a logger for nats-server based on zap sugared logger.
package natslogger

import (
	"fmt"

	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// An implementation of server.Logger based on zap sugared logger
type Logger struct {
	*zap.Logger
	debug        bool
	trace        bool
	traceVerbose bool
}

// Attach attaches the logger to the given server
func (l Logger) Attach(srv *server.Server) {
	// No need to enable debug/trace/traceVerbose if the log level is not debug
	if l.Level() != zapcore.DebugLevel {
		l.debug = false
		l.trace = false
		l.traceVerbose = false
	}
	srv.SetLoggerV2(l, l.debug, l.trace, l.traceVerbose)
}

func (l Logger) Debugf(template string, args ...interface{}) {
	l.Debug(fmt.Sprintf(template, args...))
}

func (l Logger) Warnf(template string, args ...interface{}) {
	l.Warn(fmt.Sprintf(template, args...))
}

func (l Logger) Errorf(template string, args ...interface{}) {
	l.Error(fmt.Sprintf(template, args...))
}

func (l Logger) Fatalf(template string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(template, args...))
}

// Tracef logs a trace (debug) statement
func (l Logger) Tracef(template string, args ...interface{}) {
	l.Debug(fmt.Sprintf(template, args...))
}

// Noticef logs a notice (info) statement
func (l Logger) Noticef(template string, args ...interface{}) {
	l.Info(fmt.Sprintf(template, args...))
}

// New returns a new logger based from provided parent logger.
func New(parent *zap.Logger, opts *server.Options) *Logger {
	if opts.NoLog {
		return &Logger{zap.NewNop(), false, false, false}
	}
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
	return &Logger{parent.WithOptions(zap.AddCallerSkip(4), zap.AddStacktrace(zap.DPanicLevel)), debug, trace, traceVerbose}
}
