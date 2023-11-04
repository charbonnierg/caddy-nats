package otlp

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/otelcol/app/settings"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.opentelemetry.io/collector/component"
)

func (r *OtlpReceiver) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := parser.ExpectString(d, parser.Match("otlp")); err != nil {
		return err
	}
	if d.CountRemainingArgs() > 0 {
		proto := d.Val()
		switch proto {
		case "grpc":
			r.Protocols = fnutils.DefaultIfNil(r.Protocols, &Protocols{})
			r.Protocols.GRPC = fnutils.DefaultIfNil(r.Protocols.GRPC, &settings.GRPCServerSettings{})
			if err := parser.ParseString(d, &r.Protocols.GRPC.Endpoint); err != nil {
				return err
			}
		case "http":
			r.Protocols = fnutils.DefaultIfNil(r.Protocols, &Protocols{})
			r.Protocols.HTTP = fnutils.DefaultIfNil(r.Protocols.HTTP, &HTTPConfig{
				HTTPServerSettings: &settings.HTTPServerSettings{},
			})
			if err := parser.ParseString(d, &r.Protocols.HTTP.Endpoint); err != nil {
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
					if err := parser.ParseString(d, &r.Protocols.GRPC.Endpoint); err != nil {
						return err
					}
				} else {
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "endpoint":
							if err := parser.ParseString(d, &r.Protocols.GRPC.Endpoint); err != nil {
								return err
							}
						case "max_recv_msg_size":
							var size int
							if err := parser.ParseIntByteSize(d, &size); err != nil {
								return err
							}
							r.Protocols.GRPC.MaxRecvMsgSizeMiB = uint64(size / 1024 / 1024)
							if size != 0 && r.Protocols.GRPC.MaxRecvMsgSizeMiB == 0 {
								r.Protocols.GRPC.MaxRecvMsgSizeMiB = 1
							}
						case "max_concurrent_streams":
							if err := parser.ParseUint32(d, &r.Protocols.GRPC.MaxConcurrentStreams); err != nil {
								return err
							}
						case "read_buffer_size":
							if err := parser.ParseIntByteSize(d, &r.Protocols.GRPC.ReadBufferSize); err != nil {
								return err
							}
						case "write_buffer_size":
							if err := parser.ParseIntByteSize(d, &r.Protocols.GRPC.WriteBufferSize); err != nil {
								return err
							}
						case "transport":
							if err := parser.ParseString(d, &r.Protocols.GRPC.Transport); err != nil {
								return err
							}
						case "auth", "authenticator":
							r.Protocols.GRPC.Auth = fnutils.DefaultIfNil(r.Protocols.GRPC.Auth, &settings.Authentication{})
							var name string
							if err := parser.ParseString(d, &name); err != nil {
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
					if err := parser.ParseString(d, &r.Protocols.HTTP.Endpoint); err != nil {
						return err
					}
				} else {
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "endpoint":
							if err := parser.ParseString(d, &r.Protocols.HTTP.Endpoint); err != nil {
								return err
							}
						case "traces_url_path":
							if err := parser.ParseString(d, &r.Protocols.HTTP.TracesURLPath); err != nil {
								return err
							}
						case "metrics_url_path":
							if err := parser.ParseString(d, &r.Protocols.HTTP.MetricsURLPath); err != nil {
								return err
							}
						case "logs_url_path":
							if err := parser.ParseString(d, &r.Protocols.HTTP.LogsURLPath); err != nil {
								return err
							}
						case "auth", "authenticator":
							r.Protocols.HTTP.Auth = fnutils.DefaultIfNil(r.Protocols.HTTP.Auth, &settings.Authentication{})
							var name string
							if err := parser.ParseString(d, &name); err != nil {
								return err
							}
							id := component.ID{}
							if err := id.UnmarshalText([]byte(name)); err != nil {
								return err
							}
							r.Protocols.HTTP.Auth.AuthenticatorID = id
						case "max_request_body_size":
							if err := parser.ParseInt64ByteSize(d, &r.Protocols.HTTP.MaxRequestBodySize); err != nil {
								return err
							}
						case "response_header":
							r.Protocols.HTTP.ResponseHeaders = fnutils.DefaultIfEmptyMap(r.Protocols.HTTP.ResponseHeaders, map[string]string{})
							if d.CountRemainingArgs() > 0 {
								var key string
								var value string
								if err := parser.ParseString(d, &key); err != nil {
									return err
								}
								if err := parser.ParseString(d, &value); err != nil {
									return err
								}
								r.Protocols.HTTP.ResponseHeaders[key] = value
							} else {
								if err := parser.ParseStringMap(d, &r.Protocols.HTTP.ResponseHeaders); err != nil {
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
