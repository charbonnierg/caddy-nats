// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package otelcol

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

type StructProvider struct {
	structPointer interface{}
}

// New returns a new confmap.Provider that reads the configuration from a
// struct pointer holding a json.RawMessage.
func NewProvider(structPointer interface{}) confmap.Provider {
	return &StructProvider{structPointer: structPointer}
}

func (fmp *StructProvider) Retrieve(ctx context.Context, structField string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	if len(structField) < len(schemeName) {
		return nil, fmt.Errorf("field %q not found", structField)
	}
	field := structField[len(schemeName)+1:]
	data := map[string]interface{}{}
	value := reflect.ValueOf(fmp.structPointer).Elem().FieldByName(field)
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
	return confmap.NewRetrieved(data)
}

func (*StructProvider) Scheme() string {
	return schemeName
}

func (*StructProvider) Shutdown(context.Context) error {
	return nil
}
