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

package stringx

import (
	"testing"

	"github.com/bytedance/gopkg/internal/assert"
)

func TestIs(t *testing.T) {
	assert.False(t, IsNumeric(""))
	assert.False(t, IsNumeric("  "))
	assert.False(t, IsNumeric(" bob "))
	assert.True(t, IsNumeric("123"))

	assert.False(t, IsAlpha(""))
	assert.False(t, IsAlpha(" "))
	assert.False(t, IsAlpha(" Voa "))
	assert.False(t, IsAlpha("123"))
	assert.True(t, IsAlpha("Voa"))
	assert.True(t, IsAlpha("bròwn"))

	assert.False(t, IsAlphanumeric(""))
	assert.False(t, IsAlphanumeric(" "))
	assert.False(t, IsAlphanumeric(" Voa "))
	assert.True(t, IsAlphanumeric("Voa"))
	assert.True(t, IsAlphanumeric("123"))
	assert.True(t, IsAlphanumeric("v123oa"))
	assert.False(t, IsAlphanumeric("v123oa,"))
}
