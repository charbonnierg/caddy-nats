package natsmagic

import (
	"crypto/tls"
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

func NewServer(opts *Options) *NatsMagic {
	return &NatsMagic{Options: opts.GetServerOptions(), MagicOptions: opts}
}

type NatsMagic struct {
	undoMaxProcs *func()
	logger       *zap.Logger
	ns           *server.Server
	Options      *server.Options
	MagicOptions *Options
	TLSConfig    *tls.Config
}

func (o *NatsMagic) SetLogger(logger *zap.Logger) {
	o.logger = logger
}

func (o *NatsMagic) SetTLSConfig(tlsConfig *tls.Config) {
	o.TLSConfig = tlsConfig
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
	return nil
}

func (o *NatsMagic) Stop() error {
	undo := *o.undoMaxProcs
	if undo != nil {
		defer undo()
	}
	if o.ns == nil {
		o.ns.Shutdown()
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
	tlsConfig := o.TLSConfig.Clone()
	if tlsConfig == nil {
		return
	}
	opts.LeafNode.TLSConfig = tlsConfig.Clone()
	if !opts.Websocket.NoTLS && opts.Websocket.TLSConfig == nil {
		opts.Websocket.TLSConfig = tlsConfig.Clone()
		opts.Websocket.TLSConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			config, err := tlsConfig.GetConfigForClient(hello)
			if err != nil {
				return nil, err
			}
			return config.GetCertificate(hello)
		}
		opts.Websocket.NoTLS = false
	}
	if o.MagicOptions.MQTT != nil && !o.MagicOptions.MQTT.NoTLS && opts.MQTT.TLSConfig == nil {
		opts.MQTT.TLSConfig = tlsConfig.Clone()
	}
	opts.TLSConfig = tlsConfig.Clone()
	opts.TLS = true
	opts.TLSVerify = false
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
