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

import "strconv"

func arrayValue(m map[string]interface{}, key string) []interface{} {
	v, exists := m[key]
	if exists {
		return v.([]interface{})
	} else {
		return []interface{}{}
	}
}

func boolValue(m map[string]interface{}, key string) bool {
	v, exists := m[key]
	if exists {
		switch value := v.(type) {
		case bool:
			return value
		default:
			return false
		}
	} else {
		return false
	}
}

func int32Value(m map[string]interface{}, key string) int32 {
	v, exists := m[key]
	if exists {
		switch value := v.(type) {
		case int32:
			return value
		case int64:
			return -1 // TODO: Should allow lib to control preference here.
		case float64:
			return -1 // TODO: Should allow lib to control preference here.
		default:
			return 0
		}
	} else {
		return 0
	}
}

func int64Value(m map[string]interface{}, key string) int64 {
	v, exists := m[key]
	if exists {
		switch value := v.(type) {
		case int64:
			return value
		case float64:
			return int64(value) // TODO: Should allow lib to control preference here.
		default:
			return 0
		}
	} else {
		return 0
	}
}

func float64Value(m map[string]interface{}, key string) float64 {
	v, exists := m[key]
	if exists {
		switch value := v.(type) {
		case float64:
			return value
		case int64:
			return float64(value) // TODO: Should allow lib to control preference here.
		default:
			return 0.0
		}
	} else {
		return 0.0
	}
}

func mapValue(m map[string]interface{}, key string) map[string]interface{} {
	v, exists := m[key]
	if exists {
		return v.(map[string]interface{})
	} else {
		return map[string]interface{}{}
	}
}

func stringValue(m map[string]interface{}, key string) string {
	v, exists := m[key]
	if exists {
		switch value := v.(type) {
		case string:
			return value
		case int64:
			return strconv.FormatInt(value, 10)
		case float64:
			return strconv.FormatFloat(value, 'G', -1, 64)
		case bool:
			return strconv.FormatBool(value)
		default:
			return ""
		}
	} else {
		return ""
	}
}
