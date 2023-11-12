// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package settings

import "go.opentelemetry.io/collector/component"

// Authentication defines the auth settings for the receiver.
type Authentication struct {
	// AuthenticatorID specifies the name of the extension to use in order to authenticate the incoming data point.
	AuthenticatorID component.ID `json:"authenticator"`
}
