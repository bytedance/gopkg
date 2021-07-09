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

	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	is := assert.New(t)

	is.False(IsNumeric(""))
	is.False(IsNumeric("  "))
	is.False(IsNumeric(" bob "))
	is.True(IsNumeric("123"))

	is.False(IsAlpha(""))
	is.False(IsAlpha(" "))
	is.False(IsAlpha(" Voa "))
	is.False(IsAlpha("123"))
	is.True(IsAlpha("Voa"))
	is.True(IsAlpha("bròwn"))

	is.False(IsAlphanumeric(""))
	is.False(IsAlphanumeric(" "))
	is.False(IsAlphanumeric(" Voa "))
	is.True(IsAlphanumeric("Voa"))
	is.True(IsAlphanumeric("123"))
	is.True(IsAlphanumeric("v123oa"))
	is.False(IsAlphanumeric("v123oa,"))
}
