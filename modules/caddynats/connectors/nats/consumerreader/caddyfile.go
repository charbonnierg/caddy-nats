// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package consumerreader

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
)

func (r *StreamConsumerReader) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if r.Consumer == nil {
		r.Consumer = new(natsclient.Consumer)
	}
	if err := r.Consumer.UnmarshalCaddyfile(d); err != nil {
		return err
	}
	return nil
}
