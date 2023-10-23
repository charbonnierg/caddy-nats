// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsrunner

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quara-dev/beyond/modules/nats/internal/natsmetrics"
	"github.com/quara-dev/beyond/modules/nats/internal/natsoptions"

	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"
)

// Run a NATS server with given logger
func Run(opts *natsoptions.Options, logger *zap.Logger, deadline time.Duration) error {
	runner, err := New().
		WithOptions(opts).
		WithLogger(logger).
		WithReadyTimeout(deadline).
		Build()
	if err != nil {
		return err
	}
	runner.RunForever()
	return nil
}

// Create a NATS server runner
func New() *Runner {
	runner := Runner{}
	return &runner
}

// Runner is a NATS server runner.
// It can be used to build, start, stop and wait for a NATS server.
// It can also be used to run a NATS server forever until an operating system
// signal is received.
type Runner struct {
	done          chan bool
	collector     *natsmetrics.Collector
	server        *server.Server
	logger        *zap.Logger
	Options       *natsoptions.Options
	ReadyDeadline time.Duration
}

// WithServerOptions will attach the given options to the NATS server.
func (r *Runner) WithOptions(options *natsoptions.Options) *Runner {
	r.Options = options
	return r
}

// WithLogger will attach the given logger to the NATS server
// unless the NoLog option is set to true.
func (r *Runner) WithLogger(logger *zap.Logger) *Runner {
	if r.Options != nil && r.Options.NoLog {
		return r
	}
	r.logger = logger
	return r
}

// WithReadyTimeout will set the deadline for the server to be ready for
// connections. If the server is not ready for connections before the deadline,
// an error is returned when calling Start().
func (r *Runner) WithReadyTimeout(deadline time.Duration) *Runner {
	r.ReadyDeadline = deadline
	return r
}

// Build will create the NATS server with the given options and logger.
// If the logger is nil, no logger will be attached to the server.
// If the server cannot be created, an error is returned.
func (r *Runner) Build() (*Runner, error) {
	// Use default options if none are provided
	if r.Options == nil {
		r.Options = natsoptions.New()
	}
	// Create NATS server
	srv, err := r.Options.Server(r.logger)
	if err != nil {
		return nil, err
	}
	r.server = srv
	// Create NATS collector
	collector, err := r.Options.Collector()
	if err != nil {
		return nil, err
	}
	r.collector = collector
	// Return runner
	return r, nil
}

// Server returns the NATS server.
func (r *Runner) Server() *server.Server {
	return r.server
}

// Collector returns the NATS collector.
func (r *Runner) Collector() *natsmetrics.Collector {
	return r.collector
}

// Running returns true if the NATS server is running.
func (r *Runner) Running() bool {
	if r.server == nil {
		return false
	}
	return r.server.Running()
}

// Start will start the NATS server and wait for it to be ready for connections.
// If the server is not ready for connections before the deadline, an error is
// returned.
func (r *Runner) Start() error {
	// Start the server
	r.server.Start()
	// Lookup and enable jetstream for accounts
	for _, acc := range r.Options.Accounts {
		if acc.JetStream {
			account, err := r.server.LookupAccount(acc.Name)
			if err != nil {
				return fmt.Errorf("account was not initialized: %s", err.Error())
			}
			// Enable jetstream
			err = account.EnableJetStream(nil)
			if err != nil {
				return fmt.Errorf("failed to enabled jetstream for account: %s", err.Error())
			}
		}
	}
	// Wait for server to be ready for connections
	if r.ReadyDeadline != 0 {
		if ok := r.server.ReadyForConnections(r.ReadyDeadline); !ok {
			r.server.Shutdown()
			return errors.New("server not ready for connections before deadline")
		}
		r.server.Noticef("server is waiting for connections")
	}
	// Kick-off a goroutine to track when we're done with this server
	r.done = make(chan bool, 1)
	go func() {
		r.server.WaitForShutdown()
		r.done <- true
	}()
	if r.collector != nil {
		err := r.collector.Start()
		if err != nil {
			r.server.Shutdown()
			return err
		}
	}
	return nil
}

// Stop will stop the NATS server.
func (r *Runner) Stop() error {
	// Only stop the server if it is running
	if r.server.Running() {
		r.server.Shutdown()
	}
	if r.collector != nil {
		r.collector.Stop()
	}
	if r.done != nil {
		<-r.done
	}
	r.server.Noticef("server is stopped")
	return nil
}

// Reload will reload the NATS server.
func (r *Runner) Reload() error {
	opts, err := r.Options.GetServerOptions()
	if err != nil {
		return err
	}
	// Only reload the server if it is running
	if r.server.Running() {
		r.server.ReloadOptions(opts)
	} else {
		return errors.New("server is not running")
	}
	return nil
}

// Wait will wait for the NATS server to be down or for an operating system
// signal to be received. If the server is down before operating system signal
// is received, an error is returned.
func (r *Runner) Wait() {
	// Wait for an OS signal to be received
	signalReceived := make(chan os.Signal, 1)
	signal.Notify(signalReceived, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	// Exit as soon as one of the two waiters is finished
	for {
		select {
		// Exit on OS signal (interrupt or sigterm)
		case s := <-signalReceived:
			r.server.Noticef("received signal %s", s)
			switch s {
			case syscall.SIGHUP:
				r.server.Noticef("reloading server")
				r.Reload()
				continue
			default:
				r.server.Noticef("stopping server")
				r.Stop()
				os.Exit(0)
			}
		// Return on server shutdown
		case <-r.done:
			return
		}
	}
}

// Run the NATS server forever.
// If an operating system signal is received, the server will be stopped,
// and the program will exit with code 0.
func (r *Runner) RunForever() error {
	// Start the server
	if err := r.Start(); err != nil {
		return err
	}
	// Wait for NATS server to be down
	r.Wait()
	// This will be a no-op if server is already stopped
	return r.Stop()
}
