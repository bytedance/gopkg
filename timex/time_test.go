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

package timex

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNowStd(t *testing.T) {
	tStd := time.Now()
	tNew := Now()
	t.Log(tNew, tStd)
	t.Log(tStd.Sub(tNew))
	assert.True(t, math.Abs(tNew.Sub(tStd).Seconds()) < 1)
}

func TestNow(t *testing.T) {
	tOld := Now()
	time.Sleep(4 * time.Millisecond)
	tNew := Now()
	t.Log(tOld, tNew)
	t.Log(tNew.Sub(tOld))
	assert.True(t, tNew.After(tOld))
}

func BenchmarkNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Now()
	}
}

func BenchmarkStdNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Now()
	}
}
