// Copyright 2021 ByteDance Inc.
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

package logger

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultLoggerCaller(t *testing.T) {
	previousLogger, previousLevel := defaultLogger, level
	defer func() { defaultLogger, level = previousLogger, previousLevel }()
	SetLevel(LevelTrace)
	var output bytes.Buffer
	SetDefaultLogger(&localLogger{logger: log.New(&output, "", log.Lshortfile)})
	cases := []struct {
		name string
		call func() int
	}{
		{"plain", func() int {
			_, _, line, _ := runtime.Caller(0)
			Info("message")
			return line + 1
		}},
		{"formatted", func() int {
			_, _, line, _ := runtime.Caller(0)
			Infof("%s", "message")
			return line + 1
		}},
		{"context", func() int {
			_, _, line, _ := runtime.Caller(0)
			CtxInfof(context.Background(), "%s", "message")
			return line + 1
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output.Reset()
			line := tc.call()
			expected := fmt.Sprintf("default_test.go:%d: ", line)
			if !strings.HasPrefix(output.String(), expected) {
				t.Fatalf("expected caller %q, got %q", expected, output.String())
			}
		})
	}
}
