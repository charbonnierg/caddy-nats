// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsoptions

import (
	"github.com/quara-dev/beyond/modules/nats/internal/natslogger"
	"github.com/quara-dev/beyond/modules/nats/internal/natsmetrics"

	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"
)

func (o *Options) Server(logger *zap.Logger) (*server.Server, error) {
	opts, err := o.GetServerOptions()
	if err != nil {
		return nil, err
	}
	srv, err := server.NewServer(opts)
	if err != nil {
		return nil, err
	}
	if o.NoLog {
		natslogger.NewNop().Attach(srv)
	} else if logger == nil {
		natslogger.NewDevelopment(opts).Attach(srv)
	} else {
		natslogger.New(logger, opts).Attach(srv)
	}
	return srv, nil
}

func (o *Options) Collector() (*natsmetrics.Collector, error) {
	if o.Metrics == nil {
		return nil, nil
	}
	opts, err := o.GetExporterOptions()
	if err != nil {
		return nil, err
	}
	if opts == nil {
		return nil, nil
	}
	return natsmetrics.NewCollector(*opts), nil
}
