// Copyright 2012 Arne Roomann-Kurrik
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
	"fmt"
	"testing"
)

func TestParseString(t *testing.T) {
	var (
		gold    = "Hello world"
		encoded = []byte(fmt.Sprintf("\"%v\"", gold))
		parsed  string
	)
	if err := Unmarshal(encoded, &parsed); err != nil {
		t.Fatalf("%v", err)
	}

	if gold != parsed {
		t.Fatalf("%v != %v", gold, parsed)
	}
}

func TestParseNumber(t *testing.T) {
	var (
		gold    int64 = 1234567
		encoded       = []byte(fmt.Sprintf("%v", gold))
		parsed  int64
	)
	if err := Unmarshal(encoded, &parsed); err != nil {
		t.Fatalf("%v", err)
	}
	if gold != parsed {
		t.Fatalf("%v != %v", gold, parsed)
	}
}

func TestParseNegativeNumber(t *testing.T) {
	var (
		gold    int64 = -1234567
		encoded       = []byte(fmt.Sprintf("%v", gold))
		parsed  int64
	)
	if err := Unmarshal(encoded, &parsed); err != nil {
		t.Fatalf("%v", err)
	}
	if gold != parsed {
		t.Fatalf("%v != %v", gold, parsed)
	}
}

func TestParseFloat(t *testing.T) {
	var (
		gold    float64 = 1234567.89
		encoded         = []byte("1234567.89")
		parsed  float64
	)
	if err := Unmarshal(encoded, &parsed); err != nil {
		t.Fatalf("%v", err)
	}
	if gold != parsed {
		t.Fatalf("%v != %v", gold, parsed)
	}
}

func TestParseMap(t *testing.T) {
	var (
		gold = map[string]interface{}{
			"foo": "Bar",
			"baz": 1234,
		}
		encoded = []byte("{\"foo\":\"Bar\",\"baz\":1234}")
		parsed  map[string]interface{}
	)
	if err := Unmarshal(encoded, &parsed); err != nil {
		t.Fatalf("%v", err)
	}
	if len(parsed) != len(gold) {
		t.Fatalf("Parsed len %v != gold len %v", len(parsed), len(gold))
	}
	for i, v := range parsed {
		if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", gold[i]) {
			t.Errorf("%v: %v != %v", i, v, gold[i])
		}
	}
}

func TestParseArray(t *testing.T) {
	var (
		gold    = []interface{}{1234, "Foo", 5678}
		encoded = []byte("[1234,\"Foo\",5678]")
		parsed  []interface{}
	)
	if err := Unmarshal(encoded, &parsed); err != nil {
		t.Fatalf("%v", err)
	}
	if len(parsed) != len(gold) {
		t.Fatalf("Parsed len %v != gold len %v", len(parsed), len(gold))
	}
	for i, v := range parsed {
		if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", gold[i]) {
			t.Errorf("%v: %v != %v", i, v, gold[i])
		}
	}
}
