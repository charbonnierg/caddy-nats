// SPDX-License-Identifier: Apache-2.0

package natsoptions

import (
	"errors"
	"fmt"

	"github.com/charbonnierg/caddy-nats/embedded/natsmetrics"
)

func (o *Options) GetExporterOptions() (*natsmetrics.CollectorOptions, error) {
	if o.Metrics == nil {
		return nil, nil
	}
	// Initialize exporter options
	exporterOpts := natsmetrics.CollectorOptions{}
	// Verify that monitoring is enabled
	if err := o.verifyMonitoringEnabled(); err != nil {
		return nil, err
	}
	// Verify that the server label is set
	if err := o.setServerLabel(&exporterOpts); err != nil {
		return nil, err
	}
	// Verify that the server URL is set
	if err := o.setServerUrl(&exporterOpts); err != nil {
		return nil, err
	}
	// Enable flags
	if err := o.setCollectorFlags(&exporterOpts); err != nil {
		return nil, err
	}
	// Return options
	return &exporterOpts, nil
}

func (o *Options) verifyMonitoringEnabled() error {
	if o.HTTPPort == 0 && o.HTTPSPort == 0 {
		return errors.New("metrics.http_port or metrics.https_port must be set when metrics are enabled")
	}
	return nil
}

func (o *Options) setServerLabel(opts *natsmetrics.CollectorOptions) error {
	// Verify that the server label is set
	if o.Metrics.ServerLabel != "" {
		opts.NATSServerLabel = o.Metrics.ServerLabel
	} else if o.ServerName != "" {
		opts.NATSServerLabel = o.ServerName
	} else {
		return errors.New("metrics.server_label or server_name must be set when metrics are enabled")
	}
	return nil
}

func (o *Options) setServerUrl(opts *natsmetrics.CollectorOptions) error {
	// Verify that the server URL is set
	if o.Metrics.ServerUrl != "" {
		opts.NATSServerURL = o.Metrics.ServerUrl
	} else if o.HTTPPort != 0 {
		opts.NATSServerURL = fmt.Sprintf("http://localhost:%d", o.HTTPPort)
	} else if o.HTTPSPort != 0 {
		return errors.New("metrics.server_url must be enabled when metrics.https_port is set")
	} else {
		return errors.New("either http_port or metrics.server_url and https_port must be set when metrics are enabled")
	}
	return nil
}

func (o *Options) setCollectorFlags(opts *natsmetrics.CollectorOptions) error {
	// GetVarz is always enabled
	opts.GetVarz = true
	opts.GetHealthz = o.Metrics.Healthz
	opts.GetConnz = o.Metrics.Connz
	opts.GetConnzDetailed = o.Metrics.ConnzDetailed
	opts.GetSubz = o.Metrics.Subz
	opts.GetRoutez = o.Metrics.Routez
	opts.GetGatewayz = o.Metrics.Gatewayz
	opts.GetLeafz = o.Metrics.Leafz
	opts.GetJszFilter = o.Metrics.JszFilter
	opts.GetReplicatorVarz = o.Metrics.ReplicatorVarz
	return nil
}
