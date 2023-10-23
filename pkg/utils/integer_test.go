// Copyright 2023 Guillaume Charbonnier
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils_test

import (
	"testing"

	"github.com/quara-dev/beyond/pkg/utils"
)

func TestInt32(t *testing.T) {
	tests := []struct {
		value     int
		wantInt32 int32
		wantErr   bool
	}{
		{
			value:     2147483647,
			wantInt32: 2147483647,
			wantErr:   false,
		},
		{
			value:     -2147483648,
			wantInt32: -2147483648,
			wantErr:   false,
		},
		{
			value:   2147483648,
			wantErr: true,
		},
		{
			value:   -2147483649,
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := utils.Int32(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantInt32 {
			t.Errorf("value is wrong, value=%v, wantInt32=%v, got=%v", test.value, test.wantInt32, got)
		}
	}
}

func TestInt16(t *testing.T) {
	tests := []struct {
		value     int
		wantInt16 int16
		wantErr   bool
	}{
		{
			value:     32767,
			wantInt16: 32767,
			wantErr:   false,
		},
		{
			value:     -32768,
			wantInt16: -32768,
			wantErr:   false,
		},
		{
			value:   32768,
			wantErr: true,
		},
		{
			value:   -32769,
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := utils.Int16(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantInt16 {
			t.Errorf("value is wrong, value=%v, wantInt16=%v, got=%v", test.value, test.wantInt16, got)
		}
	}
}

func TestInt8(t *testing.T) {
	tests := []struct {
		value    int
		wantInt8 int8
		wantErr  bool
	}{
		{
			value:    127,
			wantInt8: 127,
			wantErr:  false,
		},
		{
			value:    -128,
			wantInt8: -128,
			wantErr:  false,
		},
		{
			value:   128,
			wantErr: true,
		},
		{
			value:   -129,
			wantErr: true,
		},
	}

	for _, test := range tests {
		got, err := utils.Int8(test.value)
		if (test.wantErr && err == nil) || !test.wantErr && err != nil {
			t.Errorf("err is wrong, value=%v, wantErr=%v, err=%v", test.value, test.wantErr, err)
			continue
		}
		if got != test.wantInt8 {
			t.Errorf("value is wrong, value=%v, wantInt8=%v, got=%v", test.value, test.wantInt8, got)
		}
	}
}
