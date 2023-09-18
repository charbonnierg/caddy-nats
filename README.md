# caddy-nats

> Run `nats-server` as a [caddy app](https://caddyserver.com/docs/extending-caddy#app-modules).

## Example usage

```nginx
{
	log file {
		output file caddy.log {
			roll_size 100mb
			roll_keep_for 2160h # 90 days
		}
		format json
	}
	cert_issuer acme {
		disable_http_challenge
		disable_tlsalpn_challenge
		dir https://acme-staging-v02.api.letsencrypt.org/directory
		dns digitalocean {$DO_API_TOKEN}
	}
	servers {
		metrics
		protocols h1 h2
	}
	nats_issuer {
		operator eyJhbGciOiJlZDI1NTE5LW5rZXkiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoidGVzdC1vcGVyYXRvciIsInN1YiI6Ik9CM0lOWTRRTTJVRzRJWTVIVjJST0JDSkdGUVFEMzZSQVBLTFBENzdaTUhJQ09DQ1hBTU1QR1VDIiwiaXNzIjoiT0IzSU5ZNFFNMlVHNElZNUhWMlJPQkNKR0ZRUUQzNlJBUEtMUEQ3N1pNSElDT0NDWEFNTVBHVUMiLCJqdGkiOiJJRFY0U1RWSDJNNEszUFlPRjM1UElPVkhINk5VNURWTjdZRVJUTlZSRzNaUDZKRFJZM1JBIiwiaWF0IjoxNjkzODE4ODk4LCJuYXRzIjp7InR5cGUiOiJvcGVyYXRvciIsInZlcnNpb24iOjJ9fQ.9srzttM1i9X-TxUVZuScKwmL36LdBmYWRQji4tVfzagwxoeERIYm-UJDo65DoCXLVLsNLrgUK9gXP0IIwQKJAw
		system_account eyJhbGciOiJlZDI1NTE5LW5rZXkiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiU1lTIiwic3ViIjoiQUFZUldFMjJBQlBVNE1JQVRVV1pLQkNJVVk3R01BTldQTlZOUVZNWE9EMlNZTEhDTk5IM040VDUiLCJpc3MiOiJPQjNJTlk0UU0yVUc0SVk1SFYyUk9CQ0pHRlFRRDM2UkFQS0xQRDc3Wk1ISUNPQ0NYQU1NUEdVQyIsImp0aSI6IklEVjRTVFU2UktQS0RHRE9HNFhLQVNESUxRSlJNM1NVWTJUR1RTWE5YR0tINjRUU1RHV0EiLCJpYXQiOjE2OTM4MTg4OTgsIm5hdHMiOnsidHlwZSI6ImFjY291bnQiLCJ2ZXJzaW9uIjoyLCJzaWduaW5nX2tleXMiOlt7ImtleSI6IkFDMkFOMlVPNllTU1pQSU1TMjJQRjVKV1ZYMklORjVTWFVGUjQ0M1pCUVdWSk9YSzVWUEtBS0VVIiwicm9sZSI6Im1vbml0b3IiLCJ0ZW1wbGF0ZSI6eyJwdWIiOnsiYWxsb3ciOlsiJFNZUy5SRVEuQUNDT1VOVC4qLioiLCIkU1lTLlJFUS5TRVJWRVIuKi4qIl19LCJzdWIiOnsiYWxsb3ciOlsiJFNZUy5BQ0NPVU5ULiouPiIsIiRTWVMuU0VSVkVSLiouPiJdfSwic3VicyI6LTEsInBheWxvYWQiOjEwNDg1NzYsImFsbG93ZWRfY29ubmVjdGlvbl90eXBlcyI6WyJTVEFOREFSRCIsIldFQlNPQ0tFVCJdfSwia2luZCI6InVzZXJfc2NvcGUifSx7ImtleSI6IkFBRlIyTDNHQjVQUFFNV0s2WE5QREhXUTVLTzVWMkVISlBSQVpYR1Q2RUpPUDdaWEdIWE9OQkYzIiwicm9sZSI6Imlzc3VlciIsInRlbXBsYXRlIjp7InB1YiI6eyJhbGxvdyI6WyIkU1lTLlJFUS5DTEFJTVMuKi4iLCIkU1lTLlJFUS5BQ0NPVU5ULiouQ0xBSU1TLioiXX0sInN1YnMiOi0xLCJwYXlsb2FkIjoxMDQ4NTc2LCJhbGxvd2VkX2Nvbm5lY3Rpb25fdHlwZXMiOlsiU1RBTkRBUkQiLCJXRUJTT0NLRVQiXX0sImtpbmQiOiJ1c2VyX3Njb3BlIn0seyJrZXkiOiJBQ0pMNVlYN0tVU1ZMRjNMSTVCVFVRRk1JSk42VFhUUkk2TVBFS1ZLWkhRNFlWTlZITDVQWkJQUCIsInJvbGUiOiJhZG1pbmlzdHJhdG9yIiwidGVtcGxhdGUiOnsicHViIjp7ImFsbG93IjpbIj4iXX0sInN1YiI6eyJhbGxvdyI6WyI-Il19LCJzdWJzIjotMSwicGF5bG9hZCI6MTA0ODU3NiwiYWxsb3dlZF9jb25uZWN0aW9uX3R5cGVzIjpbIlNUQU5EQVJEIiwiV0VCU09DS0VUIl19LCJraW5kIjoidXNlcl9zY29wZSJ9LHsia2V5IjoiQUE3T0RYSEVXMk1OS0dLWUtMRUdNNjRUR0FPWFpNQlg0V09XUlFPUFFBS0w2QUZYSVJQNUdRTU0iLCJyb2xlIjoibGVhZm5vZGUiLCJ0ZW1wbGF0ZSI6eyJwdWIiOnsiYWxsb3ciOlsiPiJdfSwic3ViIjp7ImFsbG93IjpbIj4iXX0sInN1YnMiOi0xLCJkYXRhIjotMSwicGF5bG9hZCI6MTA0ODU3NiwiYWxsb3dlZF9jb25uZWN0aW9uX3R5cGVzIjpbIkxFQUZOT0RFIiwiTEVBRk5PREVfV1MiXX0sImtpbmQiOiJ1c2VyX3Njb3BlIn1dLCJleHBvcnRzIjpbeyJkZXNjcmlwdGlvbiI6IkFjY291bnQgc3BlY2lmaWMgbW9uaXRvcmluZyBzdHJlYW0iLCJpbmZvX3VybCI6Imh0dHBzOi8vZG9jcy5uYXRzLmlvL25hdHMtc2VydmVyL2NvbmZpZ3VyYXRpb24vc3lzX2FjY291bnRzIiwibmFtZSI6ImFjY291bnQtbW9uaXRvcmluZy1zdHJlYW1zIiwic3ViamVjdCI6IiRTWVMuQUNDT1VOVC4qLj4iLCJ0eXBlIjoic3RyZWFtIiwiYWNjb3VudF90b2tlbl9wb3NpdGlvbiI6M30seyJkZXNjcmlwdGlvbiI6IlJlcXVlc3QgYWNjb3VudCBzcGVjaWZpYyBtb25pdG9yaW5nIHNlcnZpY2VzIGZvcjogU1VCU1osIENPTk5aLCBMRUFGWiwgSlNaIGFuZCBJTkZPIiwiaW5mb191cmwiOiJodHRwczovL2RvY3MubmF0cy5pby9uYXRzLXNlcnZlci9jb25maWd1cmF0aW9uL3N5c19hY2NvdW50cyIsIm5hbWUiOiJhY2NvdW50LW1vbml0b3Jpbmctc2VydmljZXMiLCJzdWJqZWN0IjoiJFNZUy5SRVEuQUNDT1VOVC4qLioiLCJ0eXBlIjoic2VydmljZSIsInJlc3BvbnNlX3R5cGUiOiJTdHJlYW0iLCJhY2NvdW50X3Rva2VuX3Bvc2l0aW9uIjo0fSx7ImRlc2NyaXB0aW9uIjoiUmVxdWVzdCBhY2NvdW50IEpXVCIsImluZm9fdXJsIjoiaHR0cHM6Ly9kb2NzLm5hdHMuaW8vbmF0cy1zZXJ2ZXIvY29uZmlndXJhdGlvbi9zeXNfYWNjb3VudHMiLCJuYW1lIjoiYWNjb3VudC1sb29rdXAtc2VydmljZSIsInN1YmplY3QiOiIkU1lTLlJFUS5BQ0NPVU5ULiouQ0xBSU1TLkxPT0tVUCIsInR5cGUiOiJzZXJ2aWNlIiwicmVzcG9uc2VfdHlwZSI6IlN0cmVhbSIsImFjY291bnRfdG9rZW5fcG9zaXRpb24iOjR9LHsiZGVzY3JpcHRpb24iOiJSZXF1ZXN0IGFsbCBzZXJ2ZXJzIGhlYWx0aCIsImluZm9fdXJsIjoiaHR0cHM6Ly9kb2NzLm5hdHMuaW8vbmF0cy1zZXJ2ZXIvY29uZmlndXJhdGlvbi9zeXNfYWNjb3VudHMiLCJuYW1lIjoic2VydmVyLWhlYWx0aC1zZXJ2aWNlIiwic3ViamVjdCI6IiRTWVMuUkVRLlNFUlZFUi4qLkhFQUxUSFoiLCJ0eXBlIjoic2VydmljZSIsInJlc3BvbnNlX3R5cGUiOiJTdHJlYW0ifV0sImxpbWl0cyI6eyJpbXBvcnRzIjoxMCwiZXhwb3J0cyI6NSwid2lsZGNhcmRzIjp0cnVlLCJjb25uIjoxMCwibGVhZiI6MTAsInN1YnMiOjEwMDAsInBheWxvYWQiOjIwOTcxNTJ9fX0.U18Lo7yf5uLUMUSWGKJOv8Iz_rfpOJEoT3GTL846OpvKnZALKlOIpUkrLL69vQiP7v5llLmdrRZ9gLMGhlMNDw
		accounts {
			test {
				jetstream
				max_connections 100
				role default {
					publish {
						allow >
					}
					subscribe {
						allow >
					}
					limits {
						max_payload 1073741824
					}
				}
			}
		}
	}
	nats_server {
		sni local.quara.cloud
		server_name server-01
		host 127.0.0.1
		port 4222
		http_port 8222
		client_advertise local.quara.cloud:4222
		leafnodes {
			port 7422
			advertise leaf.local.quara.cloud:7422
			sni leaf.local.quara.cloud
		}
		jetstream {
			domain server-01
			max_memory 1073741824
			max_file 1073741824
		}
		websocket {
			port 10443
			advertise ws.local.quara.cloud:10443
			sni ws.local.quara.cloud
		}
		mqtt {
			port 8883
			jetstream_domain server-01
			sni mqtt.local.quara.cloud
		}
		metrics {
			host 127.0.0.1
			port 2020
			base_path /metrics
		}
		resolver full {
			path ./sanbox/jwt
		}
	}
}

app.local.quara-dev.com {
	handle /account {
		route {
			nats_issuer test {
				role default
			}
		}
	}
	handle /ws {
		reverse_proxy {
			to localhost:10443
			transport http {
				tls
				tls_server_name ws.local.quara.cloud
			}
		}
	}
	root * /testing/www
	file_server
}

metrics.local.quara-dev.com {
	basicauth /* {
		Bob $2a$14$Zkx19XLiW6VYouLHR5NmfOFU0z2GTNmpkT/5qqR7hx4IjWJPDhjvG
	}
	handle_path /caddy {
		metrics
	}
	handle_path /nats {
		rewrite * /metrics
		reverse_proxy {
			to localhost:2020
		}
	}
}
```