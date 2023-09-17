package natsmagic

import (
	"crypto/tls"
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/prometheus-nats-exporter/collector"
	"github.com/nats-io/prometheus-nats-exporter/exporter"
	"go.uber.org/automaxprocs/maxprocs"

	"go.uber.org/zap"
)

func NewServer(opts *NatsConfig) *NatsMagic {
	return &NatsMagic{Options: opts.GetServerOptions(), MagicOptions: opts}
}

type NatsMagic struct {
	undoMaxProcs       *func()
	logger             *zap.Logger
	ns                 *server.Server
	exporter           *exporter.NATSExporter
	Options            *server.Options
	MagicOptions       *NatsConfig
	StandardTLSConfig  *tls.Config
	WebsocketTLSConfig *tls.Config
	LeafnodeTLSConfig  *tls.Config
	MQTTTLSConfig      *tls.Config
}

func (o *NatsMagic) SetLogger(logger *zap.Logger) {
	o.logger = logger
}

func (o *NatsMagic) SetStandardTLSConfig(tlsConfig *tls.Config) {
	o.StandardTLSConfig = tlsConfig
}

func (o *NatsMagic) SetWebsocketTLSConfig(tlsConfig *tls.Config) {
	o.WebsocketTLSConfig = tlsConfig
}

func (o *NatsMagic) SetLeafnodeTLSConfig(tlsConfig *tls.Config) {
	o.LeafnodeTLSConfig = tlsConfig
}

func (o *NatsMagic) SetMQTTTLSConfig(tlsConfig *tls.Config) {
	o.MQTTTLSConfig = tlsConfig
}

func (o *NatsMagic) Start() error {
	if err := o.createServer(); err != nil {
		return err
	}
	// Start things up. Block here until done.
	if err := server.Run(o.ns); err != nil {
		return err
	}
	// Adjust MAXPROCS if running under linux/cgroups quotas.
	undo, err := maxprocs.Set(maxprocs.Logger(o.ns.Debugf))
	if err != nil {
		o.ns.Warnf("Failed to set GOMAXPROCS: %v", err)
	}
	o.undoMaxProcs = &undo
	if o.MagicOptions.Metrics == nil {
		return nil
	}
	// Create the exporter
	exporter := exporter.NewExporter(o.MagicOptions.GetMetricsOptions())
	collector.RemoveLogger()
	if err := exporter.Start(); err != nil {
		return err
	}
	o.logger.Info(
		fmt.Sprintf("Listening for Prometheus scrapes on %s", o.MagicOptions.Metrics.GetUrl()),
	)
	o.exporter = exporter
	return nil
}

func (o *NatsMagic) Stop() error {
	undo := *o.undoMaxProcs
	defer func() {
		if undo != nil {
			undo()
		}
	}()
	defer func() {
		if o.ns == nil {
			o.ns.Shutdown()
		}
	}()
	if o.exporter != nil {
		o.exporter.Stop()
	}
	return nil
}

func (o *NatsMagic) createServer() error {
	// Clone options
	opts := o.Options.Clone()
	// Process config file when not empty
	if opts.ConfigFile != "" {
		if err := opts.ProcessConfigFile(opts.ConfigFile); err != nil {
			return err
		}
	}
	// Configure TLS
	o.configureTLS(opts)
	// Create the server with appropriate options.
	ns, err := server.NewServer(opts)
	if err != nil {
		return err
	}
	// Configure logger
	ns.SetLoggerV2(
		newNatsLogger(o.logger),
		opts.Debug,
		opts.Trace,
		opts.TraceVerbose,
	)
	for _, v := range o.MagicOptions.ResolverPreload {
		acc, err := jwt.DecodeAccountClaims(v)
		sub := acc.Subject
		if err != nil {
			return fmt.Errorf("preload account error for %s: %s", sub, err.Error())
		}
		err = opts.AccountResolver.Store(sub, v)
		if err != nil {
			return fmt.Errorf("preload account error for %s: %s", sub, err.Error())
		}
	}
	o.ns = ns
	return nil
}

func (o *NatsMagic) configureTLS(opts *server.Options) {
	if !o.MagicOptions.NoTLS && opts.TLSConfig == nil {
		cfg := o.StandardTLSConfig.Clone()
		opts.TLSConfig = cfg.Clone()
		opts.TLSConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			config, err := cfg.GetConfigForClient(hello)
			if err != nil {
				return nil, err
			}
			return config.GetCertificate(hello)
		}
		opts.TLS = true
		opts.TLSVerify = false
	}
	if !o.MagicOptions.Websocket.NoTLS && opts.Websocket.TLSConfig == nil {
		cfg := o.WebsocketTLSConfig.Clone()
		opts.Websocket.TLSConfig = cfg.Clone()
		opts.Websocket.TLSConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			config, err := cfg.GetConfigForClient(hello)
			if err != nil {
				return nil, err
			}
			return config.GetCertificate(hello)
		}
		opts.Websocket.NoTLS = false
	}
	if !o.MagicOptions.LeafNode.NoTLS && opts.LeafNode.TLSConfig == nil {
		opts.LeafNode.TLSConfig = o.LeafnodeTLSConfig.Clone()
	}
	if !o.MagicOptions.MQTT.NoTLS && opts.MQTT.TLSConfig == nil {
		opts.MQTT.TLSConfig = o.MQTTTLSConfig.Clone()
	}
}

type natsLogger struct {
	*zap.SugaredLogger
}

func (l *natsLogger) Noticef(format string, v ...interface{}) {
	l.Infof(format, v...)
}

func (l *natsLogger) Tracef(format string, v ...interface{}) {
	l.Debugf(format, v...)
}

func newNatsLogger(logger *zap.Logger) *natsLogger {
	return &natsLogger{logger.WithOptions(zap.AddCallerSkip(4)).Sugar()}
}
