package otlp

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/otelcol/app/settings"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.opentelemetry.io/collector/component"
)

func (r *OtlpReceiver) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := caddyutils.ExpectString(d, "otlp"); err != nil {
		return err
	}
	if d.CountRemainingArgs() > 0 {
		proto := d.Val()
		switch proto {
		case "grpc":
			r.Protocols = fnutils.DefaultIfNil(r.Protocols, &Protocols{})
			r.Protocols.GRPC = fnutils.DefaultIfNil(r.Protocols.GRPC, &settings.GRPCServerSettings{})
			if err := caddyutils.ParseString(d, &r.Protocols.GRPC.Endpoint); err != nil {
				return err
			}
		case "http":
			r.Protocols = fnutils.DefaultIfNil(r.Protocols, &Protocols{})
			r.Protocols.HTTP = fnutils.DefaultIfNil(r.Protocols.HTTP, &HTTPConfig{
				HTTPServerSettings: &settings.HTTPServerSettings{},
			})
			if err := caddyutils.ParseString(d, &r.Protocols.HTTP.Endpoint); err != nil {
				return err
			}
		default:
			return d.Errf("unexpected argument %s", d.Val())
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "grpc":
				r.Protocols = fnutils.DefaultIfNil(r.Protocols, &Protocols{})
				r.Protocols.GRPC = fnutils.DefaultIfNil(r.Protocols.GRPC, &settings.GRPCServerSettings{})
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ParseString(d, &r.Protocols.GRPC.Endpoint); err != nil {
						return err
					}
				} else {
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "endpoint":
							if err := caddyutils.ParseString(d, &r.Protocols.GRPC.Endpoint); err != nil {
								return err
							}
						case "max_recv_msg_size":
							if err := caddyutils.ParseByteSizeMiB(d, &r.Protocols.GRPC.MaxRecvMsgSizeMiB); err != nil {
								return err
							}
						case "max_concurrent_streams":
							if err := caddyutils.ParseUInt32(d, &r.Protocols.GRPC.MaxConcurrentStreams); err != nil {
								return err
							}
						case "read_buffer_size":
							if err := caddyutils.ParseByteSize(d, &r.Protocols.GRPC.ReadBufferSize); err != nil {
								return err
							}
						case "write_buffer_size":
							if err := caddyutils.ParseByteSize(d, &r.Protocols.GRPC.WriteBufferSize); err != nil {
								return err
							}
						case "transport":
							if err := caddyutils.ParseString(d, &r.Protocols.GRPC.Transport); err != nil {
								return err
							}
						case "auth", "authenticator":
							r.Protocols.GRPC.Auth = fnutils.DefaultIfNil(r.Protocols.GRPC.Auth, &settings.Authentication{})
							var name string
							if err := caddyutils.ParseString(d, &name); err != nil {
								return err
							}
							id := component.ID{}
							if err := id.UnmarshalText([]byte(name)); err != nil {
								return err
							}
							r.Protocols.GRPC.Auth.AuthenticatorID = id
						default:
							return d.Errf("unrecognized subdirective %s", d.Val())
						}
					}
				}
			case "http":
				r.Protocols = fnutils.DefaultIfNil(r.Protocols, &Protocols{})
				r.Protocols.HTTP = fnutils.DefaultIfNil(r.Protocols.HTTP, &HTTPConfig{
					HTTPServerSettings: &settings.HTTPServerSettings{},
				})
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ParseString(d, &r.Protocols.HTTP.Endpoint); err != nil {
						return err
					}
				} else {
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "endpoint":
							if err := caddyutils.ParseString(d, &r.Protocols.HTTP.Endpoint); err != nil {
								return err
							}
						case "traces_url_path":
							if err := caddyutils.ParseString(d, &r.Protocols.HTTP.TracesURLPath); err != nil {
								return err
							}
						case "metrics_url_path":
							if err := caddyutils.ParseString(d, &r.Protocols.HTTP.MetricsURLPath); err != nil {
								return err
							}
						case "logs_url_path":
							if err := caddyutils.ParseString(d, &r.Protocols.HTTP.LogsURLPath); err != nil {
								return err
							}
						case "auth", "authenticator":
							r.Protocols.HTTP.Auth = fnutils.DefaultIfNil(r.Protocols.HTTP.Auth, &settings.Authentication{})
							var name string
							if err := caddyutils.ParseString(d, &name); err != nil {
								return err
							}
							id := component.ID{}
							if err := id.UnmarshalText([]byte(name)); err != nil {
								return err
							}
							r.Protocols.HTTP.Auth.AuthenticatorID = id
						case "max_request_body_size":
							if err := caddyutils.ParseByteSizeI64(d, &r.Protocols.HTTP.MaxRequestBodySize); err != nil {
								return err
							}
						case "response_header":
							r.Protocols.HTTP.ResponseHeaders = fnutils.DefaultIfEmptyMap(r.Protocols.HTTP.ResponseHeaders, map[string]string{})
							if d.CountRemainingArgs() > 0 {
								var key string
								var value string
								if err := caddyutils.ParseString(d, &key); err != nil {
									return err
								}
								if err := caddyutils.ParseString(d, &value); err != nil {
									return err
								}
								r.Protocols.HTTP.ResponseHeaders[key] = value
							} else {
								if err := caddyutils.ParseStringMap(d, &r.Protocols.HTTP.ResponseHeaders); err != nil {
									return err
								}
							}

						default:
							return d.Errf("unrecognized subdirective %s", d.Val())
						}
					}
				}
			default:
				return d.Errf("unrecognized subdirective %s", d.Val())
			}
		}
	}
	return nil
}
