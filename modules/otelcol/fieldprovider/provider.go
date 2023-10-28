// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package fieldprovider implements a confmap.Provider that reads the
// configuration from a struct pointer holding a json.RawMessage.
// It is used to load the configuration of the OpenTelemetry Collector from
// the caddy module configuration.
package fieldprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"go.opentelemetry.io/collector/confmap"
)

const (
	schemeName = "field"
)

// StructProvider is a confmap.Provider that reads the configuration from a
// struct pointer holding a json.RawMessage.
type StructProvider struct {
	structPointer interface{}
}

// New returns a new confmap.Provider that reads the configuration from a
// struct pointer holding a json.RawMessage.
func NewProvider(structPointer interface{}) confmap.Provider {
	return &StructProvider{structPointer: structPointer}
}

// Retrieve implements confmap.Provider.
// The watcher function is ignored.
func (fmp *StructProvider) Retrieve(ctx context.Context, structField string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	data, err := fmp.loadFieldByName(structField)
	if err != nil {
		return nil, err
	}
	return confmap.NewRetrieved(data)
}

// loadFieldByName loads the field with the given name from the struct pointer
// It uses the reflect package to access the field from its name.
func (s *StructProvider) loadFieldByName(field string) (map[string]interface{}, error) {
	if len(field) < len(schemeName) {
		return nil, fmt.Errorf("field %q not found", field)
	}
	field = field[len(schemeName)+1:]
	data := map[string]interface{}{}
	value := reflect.ValueOf(s.structPointer).Elem().FieldByName(field)
	if !value.IsValid() {
		return nil, fmt.Errorf("field %q not found", field)
	}
	slice := value.Bytes()
	if slice == nil {
		return nil, fmt.Errorf("field %q is nil", field)
	}
	if err := json.Unmarshal(slice, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// Scheme implements confmap.Provider.
// It returns "field".
func (*StructProvider) Scheme() string {
	return schemeName
}

// Shutdown implements confmap.Provider.
// It is a no-op.
func (*StructProvider) Shutdown(context.Context) error {
	return nil
}
