package otlphttp

import (
	"strings"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/otelcol/app/settings"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.opentelemetry.io/collector/component"
)

func (r *OtlpHttpExporter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := caddyutils.ExpectString(d, "otlphttp"); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "endpoint":
			if err := caddyutils.ParseString(d, &r.Endpoint); err != nil {
				return err
			}
		case "auth", "authenticator":
			r.Auth = fnutils.DefaultIfNil(r.Auth, &settings.Authentication{})
			var id string
			if err := caddyutils.ParseString(d, &id); err != nil {
				return err
			}
			parts := strings.Split(id, "/")
			if len(parts) == 1 {
				r.Auth.AuthenticatorID = component.NewID(component.Type(parts[0]))
			} else {
				r.Auth.AuthenticatorID = component.NewIDWithName(component.Type(parts[0]), parts[1])
			}

		case "trace_endpoint", "traces_endpoint":
			if err := caddyutils.ParseString(d, &r.TracesEndpoint); err != nil {
				return err
			}
		case "metric_endpoint", "metrics_endpoint":
			if err := caddyutils.ParseString(d, &r.MetricsEndpoint); err != nil {
				return err
			}
		case "log_endpoint", "logs_endpoint":
			if err := caddyutils.ParseString(d, &r.LogsEndpoint); err != nil {
				return err
			}
		case "read_buffer_size":
			if err := caddyutils.ParseByteSize(d, &r.ReadBufferSize); err != nil {
				return err
			}
		case "write_buffer_size":
			if err := caddyutils.ParseByteSize(d, &r.WriteBufferSize); err != nil {
				return err
			}
		case "timeout":
			if err := caddyutils.ParseDuration(d, &r.Timeout); err != nil {
				return err
			}
		case "compression":
			if err := caddyutils.ParseString(d, &r.Compression); err != nil {
				return err
			}
		case "idle_conn_timeout":
			if err := caddyutils.ParseDuration(d, r.IdleConnTimeout); err != nil {
				return err
			}
		case "max_idle_conns_per_host":
			if err := caddyutils.ParseInt(d, r.MaxIdleConnsPerHost); err != nil {
				return err
			}
		case "max_idle_conns":
			if err := caddyutils.ParseInt(d, r.MaxIdleConns); err != nil {
				return err
			}
		case "headers":
			if err := caddyutils.ParseStringMap(d, &r.Headers); err != nil {
				return err
			}
		case "sending_queue":
			r.SendingQueue = fnutils.DefaultIfNil(r.SendingQueue, &QueueSettings{Enabled: true})
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "enabled":
					if err := caddyutils.ParseBool(d, &r.SendingQueue.Enabled); err != nil {
						return err
					}
				case "num_consumers":
					if err := caddyutils.ParseInt(d, &r.SendingQueue.NumConsumers); err != nil {
						return err
					}
				case "queue_size":
					if err := caddyutils.ParseInt(d, &r.SendingQueue.QueueSize); err != nil {
						return err
					}
				case "storage_id":
					if err := caddyutils.ParseString(d, r.SendingQueue.StorageID); err != nil {
						return err
					}
				}
			}
		case "retry_settings":
			r.RetrySettings = fnutils.DefaultIfNil(r.RetrySettings, &RetrySettings{Enabled: true})
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "enabled":
					if err := caddyutils.ParseBool(d, &r.RetrySettings.Enabled); err != nil {
						return err
					}
				case "initial_interval":
					if err := caddyutils.ParseDuration(d, &r.RetrySettings.InitialInterval); err != nil {
						return err
					}
				case "randomization_factor":
					if err := caddyutils.ParseFloat64(d, &r.RetrySettings.RandomizationFactor); err != nil {
						return err
					}
				case "multiplier":
					if err := caddyutils.ParseFloat64(d, &r.RetrySettings.Multiplier); err != nil {
						return err
					}
				case "max_interval":
					if err := caddyutils.ParseDuration(d, &r.RetrySettings.MaxInterval); err != nil {
						return err
					}
				case "max_elapsed_time":
					if err := caddyutils.ParseDuration(d, &r.RetrySettings.MaxElapsedTime); err != nil {
						return err
					}
				}
			}
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}
