// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package natsmetrics is a Prometheus collector for NATS.
package natsmetrics

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/prometheus-nats-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	modeStarted uint8 = iota + 1
	modeStopped
)

// CollectorOptions are options to configure the NATS collector
type CollectorOptions struct {
	GetHealthz             bool
	GetConnz               bool
	GetConnzDetailed       bool
	GetVarz                bool
	GetSubz                bool
	GetRoutez              bool
	GetGatewayz            bool
	GetLeafz               bool
	GetReplicatorVarz      bool
	GetJszFilter           string
	RetryInterval          time.Duration
	NATSServerURL          string
	NATSServerLabel        string
	NATSMonitoringUser     string
	NATSMonitoringPassword string
	NATSMonitoringBasePath string
}

// Collector collects NATS metrics
type Collector struct {
	sync.Mutex
	opts       *CollectorOptions
	doneWg     sync.WaitGroup
	Collectors []prometheus.Collector
	servers    []*collector.CollectedServer
	mode       uint8
}

// Defaults
var (
	DefaultMonitorURL        = "http://localhost:8222"
	DefaultRetryIntervalSecs = 30
)

// NewCollector creates a new NATS exporter
func NewCollector(opts CollectorOptions) *Collector {
	o := opts
	collector.RemoveLogger()
	ne := &Collector{
		opts: &o,
	}
	if o.NATSServerURL != "" {
		_ = ne.AddServer(o.NATSServerLabel, o.NATSServerURL)
	}
	return ne
}

func (ne *Collector) createCollector(system, endpoint string) {
	ne.registerCollector(system, endpoint,
		collector.NewCollector(system, endpoint,
			ne.opts.NATSMonitoringBasePath,
			ne.servers))
}

func (ne *Collector) registerCollector(system, endpoint string, nc prometheus.Collector) {
	if err := prometheus.Register(nc); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			collector.Errorf("A collector for this server's metrics has already been registered.")
		} else {
			collector.Debugf("Unable to register collector %s (%v), Retrying.", endpoint, err)
			time.AfterFunc(ne.opts.RetryInterval, func() {
				collector.Debugf("Creating a collector for endpoint: %s", endpoint)
				ne.Lock()
				ne.createCollector(system, endpoint)
				ne.Unlock()
			})
		}
	} else {
		collector.Debugf("Registered collector for system %s, endpoint: %s", system, endpoint)
		ne.Collectors = append(ne.Collectors, nc)
	}
}

// AddServer is an advanced API; normally the NATS server should be set
// through the options.  Adding more than one server will
// violate Prometheus.io guidelines.
func (ne *Collector) AddServer(id, url string) error {
	ne.Lock()
	defer ne.Unlock()
	if ne.mode == modeStarted {
		return fmt.Errorf("servers cannot be added after the exporter is started")
	}
	cs := &collector.CollectedServer{ID: id, URL: url}
	if ne.servers == nil {
		ne.servers = make([]*collector.CollectedServer, 0)
	}
	ne.servers = append(ne.servers, cs)
	return nil
}

// InitializeCollectors initializes the Collectors for the exporter.
// Caller must lock
func (ne *Collector) InitializeCollectors() error {
	opts := ne.opts

	if len(ne.servers) == 0 {
		return fmt.Errorf("no servers configured to obtain metrics")
	}

	getJsz := opts.GetJszFilter != ""
	if !opts.GetHealthz && !opts.GetConnz && !opts.GetConnzDetailed && !opts.GetRoutez &&
		!opts.GetSubz && !opts.GetVarz && !opts.GetGatewayz && !opts.GetLeafz &&
		!opts.GetReplicatorVarz && !getJsz {
		return fmt.Errorf("no Collectors specfied")
	}
	if opts.GetReplicatorVarz && opts.GetVarz {
		return fmt.Errorf("replicatorVarz cannot be used with varz")
	}
	if opts.GetSubz {
		ne.createCollector(collector.CoreSystem, "subsz")
	}
	if opts.GetVarz {
		ne.createCollector(collector.CoreSystem, "varz")
	}
	if opts.GetHealthz {
		ne.createCollector(collector.CoreSystem, "healthz")
	}
	if opts.GetConnzDetailed {
		ne.createCollector(collector.CoreSystem, "connz_detailed")
	} else if opts.GetConnz {
		ne.createCollector(collector.CoreSystem, "connz")
	}
	if opts.GetGatewayz {
		ne.createCollector(collector.CoreSystem, "gatewayz")
	}
	if opts.GetLeafz {
		ne.createCollector(collector.CoreSystem, "leafz")
	}
	if opts.GetRoutez {
		ne.createCollector(collector.CoreSystem, "routez")
	}
	if opts.GetReplicatorVarz {
		ne.createCollector(collector.ReplicatorSystem, "varz")
	}
	if getJsz {
		switch strings.ToLower(opts.GetJszFilter) {
		case "account", "accounts", "consumer", "consumers", "all", "stream", "streams":
		default:
			return fmt.Errorf("invalid jsz filter %q", opts.GetJszFilter)
		}
		ne.createCollector(collector.JetStreamSystem, opts.GetJszFilter)
	}

	return nil
}

// ClearCollectors unregisters the collectors
// caller must lock
func (ne *Collector) ClearCollectors() {
	if ne.Collectors != nil {
		for _, c := range ne.Collectors {
			prometheus.Unregister(c)
		}
		ne.Collectors = nil
	}
}

// Start runs the exporter process.
func (ne *Collector) Start() error {
	ne.Lock()
	defer ne.Unlock()
	if ne.mode == modeStarted {
		return nil
	}

	if err := ne.InitializeCollectors(); err != nil {
		ne.ClearCollectors()
		return err
	}
	ne.doneWg.Add(1)
	ne.mode = modeStarted
	return nil
}

// WaitUntilDone blocks until the collector is stopped.
func (ne *Collector) WaitUntilDone() {
	ne.Lock()
	wg := &ne.doneWg
	ne.Unlock()
	wg.Wait()
}

// Stop stops the collector.
func (ne *Collector) Stop() {
	ne.Lock()
	defer ne.Unlock()

	if ne.mode == modeStopped {
		return
	}
	ne.ClearCollectors()
	ne.doneWg.Done()
	ne.mode = modeStopped
}
