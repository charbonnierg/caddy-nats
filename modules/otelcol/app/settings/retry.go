package settings

import "time"

// Config defines configuration for retrying batches in case of receiving a retryable error from a downstream
// consumer. If the retryable error doesn't provide a delay, exponential backoff is applied.
type RetryConfig struct {
	// Enabled indicates whether to not retry sending logs in case of receiving a retryable error from a downstream
	// consumer. Default is false.
	Enabled bool `json:"enabled,omitempty"`
	// InitialInterval the time to wait after the first failure before retrying. Default value is 1 second.
	InitialInterval time.Duration `json:"initial_interval,omitempty"`
	// MaxInterval is the upper bound on backoff interval. Once this value is reached the delay between
	// consecutive retries will always be `MaxInterval`. Default value is 30 seconds.
	MaxInterval time.Duration `json:"max_interval,omitempty"`
	// MaxElapsedTime is the maximum amount of time (including retries) spent trying to send a logs batch to
	// a downstream consumer. Once this value is reached, the data is discarded. It never stops if MaxElapsedTime == 0.
	// Default value is 5 minutes.
	MaxElapsedTime time.Duration `json:"max_elapsed_time,omitempty"`
}
