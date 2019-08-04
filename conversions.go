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
		return v.(bool)
	} else {
		return false
	}
}

func int32Value(m map[string]interface{}, key string) int32 {
	v, exists := m[key]
	if exists {
		return v.(int32)
	} else {
		return 0
	}
}

func int64Value(m map[string]interface{}, key string) int64 {
	v, exists := m[key]
	if exists {
		return v.(int64)
	} else {
		return 0
	}
}

func float64Value(m map[string]interface{}, key string) float64 {
	v, exists := m[key]
	if exists {
		return v.(float64)
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
		return v.(string)
	} else {
		return ""
	}
}
