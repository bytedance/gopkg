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
	"unicode"
)

// IsAlpha checks if the string contains only unicode letters.
func IsAlpha(s string) bool {
	if s == "" {
		return false
	}
	for _, v := range s {
		if !unicode.IsLetter(v) {
			return false
		}
	}
	return true
}

// IsAlphanumeric checks if the string contains only Unicode letters or digits.
func IsAlphanumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, v := range s {
		if !isAlphanumeric(v) {
			return false
		}
	}
	return true
}

// IsNumeric Checks if the string contains only digits. A decimal point is not a digit and returns false.
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, v := range s {
		if !unicode.IsDigit(v) {
			return false
		}
	}
	return true
}

func isAlphanumeric(v rune) bool {
	return unicode.IsDigit(v) || unicode.IsLetter(v)
}
