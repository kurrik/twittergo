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

// This file provides a custom JSON parser suitable for processing Twitter data.
// Some differences from the standard Golang json package:
//   * Does not use reflection, parses into standard map/slice/value structs.
//   * Parses numbers into int64 where possible, float64 otherwise.
//   * Faster!  Probably due to no reflection:
//      BenchmarkParseTweet	   10000	    155714 ns/op
//      BenchmarkCustomJSON	   20000	     85822 ns/op

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

const (
	STRING = iota
	NUMBER
	MAP
	ARRAY
	ENDARRAY
	ESCAPE
	BOOL
	NULL
)

type Event struct {
	Type  int
	Index int
}

type State struct {
	data   []byte
	i      int
	v      interface{}
	events []Event
}

func (s *State) Read() (err error) {
	var t int = s.nextType()
	switch t {
	case STRING:
		err = s.readString()
	case NUMBER:
		err = s.readNumber()
	case MAP:
		err = s.readMap()
	case ARRAY:
		err = s.readArray()
	case ENDARRAY:
		s.i++
		err = EndArray{}
	case BOOL:
		err = s.readBool()
	case NULL:
		err = s.readNull()
	case ESCAPE:
		err = fmt.Errorf("JSON should not start with escape")
	default:
		b := string(s.data[s.i-10 : s.i])
		c := string(s.data[s.i : s.i+1])
		e := string(s.data[s.i+1 : s.i+10])
		err = fmt.Errorf("Unrecognized type in %v -->%v<-- %v", b, c, e)
	}
	return
}

func (s *State) nextType() int {
	c := s.data[s.i]
	switch {
	case c == '"':
		return STRING
	case '0' <= c && c <= '9' || c == '-':
		return NUMBER
	case c == '[':
		return ARRAY
	case c == ']':
		return ENDARRAY
	case c == '{':
		return MAP
	case c == 't' || c == 'T' || c == 'f' || c == 'F':
		return BOOL
	case c == 'n':
		return NULL
	}
	return -1
}

func (s *State) readString() error {
	var c byte
	c = s.data[s.i]
	switch {
	case c == '"':
		break
	case c == '}':
		s.i++
		return EndMap{}
	case c == ']':
		s.i++
		return EndArray{}
	}
	s.i++
	var start int = s.i
	var buf = new(bytes.Buffer)
	var more = true
	for more {
		c = s.data[s.i]
		switch {
		case c == '\\':
			buf.Write(s.data[start:s.i])
			s.i++
			start = s.i
		case c == '"':
			more = false
		case s.i >= len(s.data)-1:
			return fmt.Errorf("No string terminator")
		}
		s.i++
	}
	buf.Write(s.data[start : s.i-1])
	s.v = buf.String()
	return nil
}

func (s *State) readNumber() (err error) {
	var c byte
	var val int64 = 0
	var valf float64 = 0
	var mult int64 = 1
	if s.data[s.i] == '-' {
		mult = -1
		s.i++
	}
	var more = true
	var places int = 0
	for more {
		c = s.data[s.i]
		switch {
		case '0' <= c && c <= '9':
			if places != 0 {
				places *= 10
			}
			val = val*10 + int64(c-'0')
		case '}' == c:
			err = EndMap{}
			more = false
		case ']' == c:
			err = EndArray{}
			more = false
		case ',' == c:
			s.i--
			more = false
		case ' ' == c || '\t' == c:
			more = false
		case '.' == c:
			valf = float64(val)
			val = 0
			places = 1
		default:
			return fmt.Errorf("Bad num char: %v", string([]byte{c}))
		}
		if s.i >= len(s.data)-1 {
			more = false
		}
		s.i++
	}
	if places > 0 {
		s.v = valf + (float64(val)/float64(places))*float64(mult)
	} else {
		s.v = val * mult
	}
	return
}

type EndMap struct{}

func (e EndMap) Error() string {
	return "End of map structure encountered."
}

type EndArray struct{}

func (e EndArray) Error() string {
	return "End of array structure encountered."
}

func (s *State) readComma() (err error) {
	var more = true
	for more {
		switch {
		case s.data[s.i] == ',':
			more = false
		case s.data[s.i] == '}':
			s.i++
			return EndMap{}
		case s.data[s.i] == ']':
			s.i++
			return EndArray{}
		case s.i >= len(s.data)-1:
			return fmt.Errorf("No comma")
		}
		s.i++
	}
	return nil
}

func (s *State) readColon() (err error) {
	var more = true
	for more {
		switch {
		case s.data[s.i] == ':':
			more = false
		case s.i >= len(s.data)-1:
			return fmt.Errorf("No colon")
		}
		s.i++
	}
	return nil
}

func (s *State) readMap() (err error) {
	s.i++
	var (
		m   map[string]interface{}
		key string
	)
	m = make(map[string]interface{})
	for {
		if err = s.readString(); err != nil {
			return
		}
		key = s.v.(string)
		if err = s.readColon(); err != nil {
			return
		}
		if err = s.Read(); err != nil {
			if _, ok := err.(EndMap); !ok {
				return
			}
		}
		m[key] = s.v
		if _, ok := err.(EndMap); ok {
			break
		}
		if err = s.readComma(); err != nil {
			if _, ok := err.(EndMap); ok {
				break
			}
			return
		}
	}
	s.v = m
	return nil
}

func (s *State) readArray() (err error) {
	s.i++
	var (
		a []interface{}
	)
	a = make([]interface{}, 0, 10)
	for {
		if err = s.Read(); err != nil {
			if _, ok := err.(EndArray); !ok {
				return
			}
		}
		a = append(a, s.v)
		if _, ok := err.(EndArray); ok {
			break
		}
		if err = s.readComma(); err != nil {
			if _, ok := err.(EndArray); ok {
				break
			}
			return
		}
	}
	s.v = a
	return nil
}

func (s *State) readBool() (err error) {
	if strings.ToLower(string(s.data[s.i:s.i+4])) == "true" {
		s.i += 4
		s.v = true
	} else if strings.ToLower(string(s.data[s.i:s.i+5])) == "false" {
		s.i += 5
		s.v = false
	} else {
		err = fmt.Errorf("Could not parse boolean")
	}
	return
}

func (s *State) readNull() (err error) {
	if strings.ToLower(string(s.data[s.i:s.i+4])) == "null" {
		s.i += 4
		s.v = nil
	} else {
		err = fmt.Errorf("Could not parse null")
	}
	return
}

func Unmarshal(data []byte, v interface{}) error {
	state := &State{data, 0, v, make([]Event, 0, 10)}
	if err := state.Read(); err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("Need a pointer, got %v", reflect.TypeOf(v))
	}
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	sv := reflect.ValueOf(state.v)
	for sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}
	rv.Set(sv)
	return nil
}
