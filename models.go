// Copyright 2011 Arne Roomann-Kurrik
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
	"strconv"
)

type User map[string]interface{}

func (u User) Id() uint64 {
	id, _ := strconv.ParseUint(u["id_str"].(string), 10, 64)
	return id
}

func (u User) IdStr() string {
	return u["id_str"].(string)
}

func (u User) Name() string {
	return u["name"].(string)
}
