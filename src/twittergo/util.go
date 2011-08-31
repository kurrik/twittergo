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
	"fmt"
	"io"
	"os"
)

// Wraps an existing ReadCloser and logs all output passed through it
// to stdout.
type LoggingReader struct {
	reader io.ReadCloser
}

// Wrap an existing ReadCloser with the logging one.
func NewLoggingReader(reader io.ReadCloser) *LoggingReader {
	return &LoggingReader{reader: reader}
}

// Read the requested bytes from the wrapped ReadCloser, logging them to stdout.
func (lr *LoggingReader) Read(p []byte) (int, os.Error) {
	n, err := lr.reader.Read(p)
	fmt.Println(string(p))
	return n, err
}

// Closes the wrapped ReadCloser.
func (lr *LoggingReader) Close() os.Error {
	return lr.reader.Close()
}
