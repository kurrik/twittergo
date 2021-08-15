// Copyright 2019 Arne Roomann-Kurrik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package twittergo

import (
	"testing"
)

func TestConversions(t *testing.T) {
	m := map[string]interface{}{
		"arrayKey":   []interface{}{"foo", "bar", "baz"},
		"boolKey":    true,
		"int32Key":   int32(1234),
		"int64Key":   int64(1002011200236892166),
		"float64Key": float64(1002011200236892166.1234),
		"mapKey":     map[string]interface{}{"foo": "bar"},
	}
	if len(arrayValue(m, "arrayKey")) != 3 {
		t.Errorf("arrayValue did not produce correct result for valid key")
	}
	if len(arrayValue(m, "badKey")) != 0 {
		t.Errorf("arrayValue did not product correct result for invalid key")
	}
	if boolValue(m, "boolKey") != true {
		t.Errorf("boolValue did not produce correct result for valid key")
	}
	if boolValue(m, "badKey") != false {
		t.Errorf("boolValue did not product correct result for invalid key")
	}
	if int32Value(m, "int32Key") != 1234 {
		t.Errorf("int32Value did not produce correct result for valid key")
	}
	if int32Value(m, "badKey") != 0 {
		t.Errorf("int32Value did not product correct result for invalid key")
	}
	if int64Value(m, "int64Key") != 1002011200236892166 {
		t.Errorf("int64Value did not produce correct result for valid key")
	}
	if int64Value(m, "badKey") != 0 {
		t.Errorf("int64Value did not product correct result for invalid key")
	}
	if float64Value(m, "float64Key") != 1002011200236892166.1234 {
		t.Errorf("float64Value did not produce correct result for valid key")
	}
	if float64Value(m, "badKey") != 0 {
		t.Errorf("float64Value did not product correct result for invalid key")
	}
	if len(mapValue(m, "mapKey")) != 1 {
		t.Errorf("mapValue did not produce correct result for valid key")
	}
	if len(mapValue(m, "badKey")) != 0 {
		t.Errorf("mapValue did not product correct result for invalid key")
	}

	// Test conversions
	if int32Value(m, "int64Key") != -1 {
		t.Errorf("int32Value did not product correct result for an int64 key")
	}
	if int32Value(m, "float64Key") != -1 {
		t.Errorf("int32Value did not product correct result for a float64 key")
	}
	// This is dangerous - note the discrepancy in returned value: 1002011200236892160 vs 1002011200236892166.1234
	if int64Value(m, "float64Key") != 1002011200236892160 {
		t.Errorf("int64Value did not product correct result for a float64 key")
	}
	if float64Value(m, "int64Key") != 1002011200236892166.0 {
		t.Errorf("float64Value did not product correct result for an int64 key")
	}
}
