package settings

import "time"

type TLSSetting struct {
	// Path to the CA cert. For a client this verifies the server certificate.
	// For a server this verifies client certificates. If empty uses system root CA.
	// (optional)
	CAFile string `json:"ca_file,omitempty"`

	// In memory PEM encoded cert. (optional)
	CAPem string `json:"ca_pem,omitempty"`

	// Path to the TLS cert to use for TLS required connections. (optional)
	CertFile string `json:"cert_file,omitempty"`

	// In memory PEM encoded TLS cert to use for TLS required connections. (optional)
	CertPem string `json:"cert_pem,omitempty"`

	// Path to the TLS key to use for TLS required connections. (optional)
	KeyFile string `json:"key_file,omitempty"`

	// In memory PEM encoded TLS key to use for TLS required connections. (optional)
	KeyPem string `json:"key_pem,omitempty"`

	// MinVersion sets the minimum TLS version that is acceptable.
	// If not set, TLS 1.2 will be used. (optional)
	MinVersion string `json:"min_version,omitempty"`

	// MaxVersion sets the maximum TLS version that is acceptable.
	// If not set, refer to crypto/tls for defaults. (optional)
	MaxVersion string `json:"max_version,omitempty"`

	// ReloadInterval specifies the duration after which the certificate will be reloaded
	// If not set, it will never be reloaded (optional)
	ReloadInterval time.Duration `json:"reload_interval,omitempty"`
}

// TLSServerSetting contains TLS configurations that are specific to server
// connections in addition to the common configurations. This should be used by
// components configuring TLS server connections.
type TLSServerSetting struct {
	TLSSetting

	// These are config options specific to server connections.

	// Path to the TLS cert to use by the server to verify a client certificate. (optional)
	// This sets the ClientCAs and ClientAuth to RequireAndVerifyClientCert in the TLSConfig. Please refer to
	// https://godoc.org/crypto/tls#Config for more information. (optional)
	ClientCAFile string `json:"client_ca_file,omitempty"`

	// Reload the ClientCAs file when it is modified
	// (optional, default false)
	ReloadClientCAFile bool `json:"client_ca_file_reload,omitempty"`
}

// TLSClientSetting contains TLS configurations that are specific to client
// connections in addition to the common configurations. This should be used by
// components configuring TLS client connections.
type TLSClientSetting struct {
	TLSSetting

	// These are config options specific to client connections.

	// In gRPC when set to true, this is used to disable the client transport security.
	// See https://godoc.org/google.golang.org/grpc#WithInsecure.
	// In HTTP, this disables verifying the server's certificate chain and host name
	// (InsecureSkipVerify in the tls Config). Please refer to
	// https://godoc.org/crypto/tls#Config for more information.
	// (optional, default false)
	Insecure bool `json:"insecure"`
	// InsecureSkipVerify will enable TLS but not verify the certificate.
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
	// ServerName requested by client for virtual hosting.
	// This sets the ServerName in the TLSConfig. Please refer to
	// https://godoc.org/crypto/tls#Config for more information. (optional)
	ServerName string `json:"server_name_override"`
}
