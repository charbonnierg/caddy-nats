// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parseutils_test

import (
	"testing"

	"github.com/quara-dev/beyond/pkg/parseutils"
)

func TestParseInt32(t *testing.T) {
	tests := []struct {
		value     string
		wantInt32 int32
		wantErr   bool
	}{
		{
			value:     "2147483647",
			wantInt32: 2147483647,
			wantErr:   false,
		},
		{
			value:     "-2147483648",
			wantInt32: -2147483648,
			wantErr:   false,
		},
		{
			value:   "2147483648",
			wantErr: true,
		},
		{
			value:   "-2147483649",
			wantErr: true,
		},
		{
			value:   "test",
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := parseutils.ParseInt32(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantInt32 {
			t.Errorf("value is wrong, value=%v, wantInt32=%v, got=%v", test.value, test.wantInt32, got)
		}
	}
}

func TestParseUInt32(t *testing.T) {
	tests := []struct {
		value      string
		wantUInt32 uint32
		wantErr    bool
	}{
		{
			value:      "4294967295",
			wantUInt32: 4294967295,
			wantErr:    false,
		},
		{
			value:      "0",
			wantUInt32: 0,
			wantErr:    false,
		},
		{
			value:   "4294967296",
			wantErr: true,
		},
		{
			value:   "-1",
			wantErr: true,
		},
		{
			value:   "test",
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := parseutils.ParseUInt32(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantUInt32 {
			t.Errorf("value is wrong, value=%v, wantInt32=%v, got=%v", test.value, test.wantUInt32, got)
		}
	}
}

func TestParseInt16(t *testing.T) {
	tests := []struct {
		value     string
		wantInt16 int16
		wantErr   bool
	}{
		{
			value:     "32767",
			wantInt16: 32767,
			wantErr:   false,
		},
		{
			value:     "-32768",
			wantInt16: -32768,
			wantErr:   false,
		},
		{
			value:   "32768",
			wantErr: true,
		},
		{
			value:   "-32769",
			wantErr: true,
		},
		{
			value:   "test",
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := parseutils.ParseInt16(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantInt16 {
			t.Errorf("value is wrong, value=%v, wantInt16=%v, got=%v", test.value, test.wantInt16, got)
		}
	}
}

func TestParseUInt16(t *testing.T) {
	tests := []struct {
		value      string
		wantUInt16 uint16
		wantErr    bool
	}{
		{
			value:      "65535",
			wantUInt16: 65535,
			wantErr:    false,
		},
		{
			value:      "0",
			wantUInt16: 0,
			wantErr:    false,
		},
		{
			value:   "65536",
			wantErr: true,
		},
		{
			value:   "-1",
			wantErr: true,
		},
		{
			value:   "test",
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := parseutils.ParseUInt16(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantUInt16 {
			t.Errorf("value is wrong, value=%v, wantInt16=%v, got=%v", test.value, test.wantUInt16, got)
		}
	}
}

func TestParseInt8(t *testing.T) {
	tests := []struct {
		value    string
		wantInt8 int8
		wantErr  bool
	}{
		{
			value:    "127",
			wantInt8: 127,
			wantErr:  false,
		},
		{
			value:    "-128",
			wantInt8: -128,
			wantErr:  false,
		},
		{
			value:   "128",
			wantErr: true,
		},
		{
			value:   "-129",
			wantErr: true,
		},
		{
			value:   "test",
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := parseutils.ParseInt8(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantInt8 {
			t.Errorf("value is wrong, value=%v, wantInt8=%v, got=%v", test.value, test.wantInt8, got)
		}
	}
}

func TestParseUInt8(t *testing.T) {
	tests := []struct {
		value     string
		wantUInt8 uint8
		wantErr   bool
	}{
		{
			value:     "255",
			wantUInt8: 255,
			wantErr:   false,
		},
		{
			value:     "0",
			wantUInt8: 0,
			wantErr:   false,
		},
		{
			value:   "256",
			wantErr: true,
		},
		{
			value:   "-1",
			wantErr: true,
		},
		{
			value:   "test",
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := parseutils.ParseUInt8(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantUInt8 {
			t.Errorf("value is wrong, value=%v, wantInt16=%v, got=%v", test.value, test.wantUInt8, got)
		}
	}
}
