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
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/bytedance/gopkg/internal/hack"
	"github.com/bytedance/gopkg/lang/dirtmake"
	"github.com/bytedance/gopkg/lang/fastrand"
)

// Error pre define
var (
	ErrDecodeRune = errors.New("error occurred on rune decoding")
)

// PadLeftChar left pad a string with a specified character in a larger string (specified size).
// if the size is less than the param string, the param string is returned.
// note: size is unicode size.
func PadLeftChar(s string, size int, ch rune) string {
	return padCharLeftOrRight(s, size, ch, true)
}

// PadLeftSpace left pad a string with space character(' ') in a larger string(specified size).
// if the size is less than the param string, the param string is returned.
// note: size is unicode size.
func PadLeftSpace(s string, size int) string {
	return PadLeftChar(s, size, ' ')
}

// PadRightChar right pad a string with a specified character in a larger string(specified size).
// if the size is less than the param string, the param string is returned.
// note: size is unicode size.
func PadRightChar(s string, size int, ch rune) string {
	return padCharLeftOrRight(s, size, ch, false)
}

// PadRightSpace right pad a string with space character(' ') in a large string(specified size).
// if the size is less than the param string, the param string is returned.
// note: size is unicode size.
func PadRightSpace(s string, size int) string {
	return PadRightChar(s, size, ' ')
}

// PadCenterChar center pad a string with a specified character in a larger string(specified size).
// if the size is less than the param string, the param string is returned.
// note: size is unicode size.
func PadCenterChar(s string, size int, ch rune) string {
	if size <= 0 {
		return s
	}
	length := utf8.RuneCountInString(s)
	pads := size - length
	if pads <= 0 {
		return s
	}

	// pad left
	leftPads := pads / 2
	if leftPads > 0 {
		s = padRawLeftChar(s, ch, leftPads)
	}
	// pad right
	rightPads := size - leftPads - length
	if rightPads > 0 {
		s = padRawRightChar(s, ch, rightPads)
	}
	return s
}

// PadCenterSpace center pad a string with space character(' ') in a larger string(specified size).
// if the size is less than the param string, the param string is returned.
// note: size is unicode size.
func PadCenterSpace(s string, size int) string {
	return PadCenterChar(s, size, ' ')
}

func padCharLeftOrRight(s string, size int, ch rune, isLeft bool) string {
	if size <= 0 {
		return s
	}
	pads := size - utf8.RuneCountInString(s)
	if pads <= 0 {
		return s
	}
	if isLeft {
		return padRawLeftChar(s, ch, pads)
	}
	return padRawRightChar(s, ch, pads)
}

func padRawLeftChar(s string, ch rune, padSize int) string {
	return RepeatChar(ch, padSize) + s
}

func padRawRightChar(s string, ch rune, padSize int) string {
	return s + RepeatChar(ch, padSize)
}

// RepeatChar returns padding using the specified delimiter repeated to a given length.
func RepeatChar(ch rune, repeat int) string {
	if repeat <= 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.Grow(repeat)
	for i := 0; i < repeat; i++ {
		sb.WriteRune(ch)
	}
	return sb.String()
}

// RemoveChar removes all occurrences of a specified character from the string.
func RemoveChar(s string, rmVal rune) string {
	if s == "" {
		return s
	}
	sb := strings.Builder{}
	sb.Grow(len(s) / 2)

	for _, v := range s {
		if v != rmVal {
			sb.WriteRune(v)
		}
	}
	return sb.String()
}

// RemoveString removes all occurrences of a substring from the string.
func RemoveString(s, rmStr string) string {
	if s == "" || rmStr == "" {
		return s
	}
	return strings.ReplaceAll(s, rmStr, "")
}

// Rotate rotates(circular shift) a string of shift characters.
func Rotate(s string, shift int) string {
	if shift == 0 {
		return s
	}
	sLen := len(s)
	if sLen == 0 {
		return s
	}

	shiftMod := shift % sLen
	if shiftMod == 0 {
		return s
	}

	offset := -(shiftMod)
	sb := strings.Builder{}
	sb.Grow(sLen)
	_, _ = sb.WriteString(SubStart(s, offset))
	_, _ = sb.WriteString(Sub(s, 0, offset))
	return sb.String()
}

// Sub returns substring from specified string avoiding panics with index start and end.
// start, end are based on unicode(utf8) count.
func Sub(s string, start, end int) string {
	return sub(s, start, end)
}

// SubStart returns substring from specified string avoiding panics with start.
// start, end are based on unicode(utf8) count.
func SubStart(s string, start int) string {
	return sub(s, start, math.MaxInt64)
}

func sub(s string, start, end int) string {
	if s == "" {
		return ""
	}

	unicodeLen := utf8.RuneCountInString(s)
	// end
	if end < 0 {
		end += unicodeLen
	}
	if end > unicodeLen {
		end = unicodeLen
	}
	// start
	if start < 0 {
		start += unicodeLen
	}
	if start > end {
		return ""
	}

	// start <= end
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start == 0 && end == unicodeLen {
		return s
	}

	sb := strings.Builder{}
	sb.Grow(end - start)
	runeIndex := 0
	for _, v := range s {
		if runeIndex >= end {
			break
		}
		if runeIndex >= start {
			sb.WriteRune(v)
		}
		runeIndex++
	}
	return sb.String()
}

// MustReverse reverses a string, panics when error happens.
func MustReverse(s string) string {
	result, err := Reverse(s)
	if err != nil {
		panic(err)
	}
	return result
}

// Reverse reverses a string with error status returned.
func Reverse(s string) (string, error) {
	if s == "" {
		return s, nil
	}
	src := hack.StringToBytes(s)
	dst := dirtmake.Bytes(len(s), len(s))
	srcIndex := len(s)
	dstIndex := 0
	for srcIndex > 0 {
		r, n := utf8.DecodeLastRune(src[:srcIndex])
		if r == utf8.RuneError {
			return hack.BytesToString(dst), ErrDecodeRune
		}
		utf8.EncodeRune(dst[dstIndex:], r)
		srcIndex -= n
		dstIndex += n
	}
	return hack.BytesToString(dst), nil
}

// Shuffle shuffles runes in a string and returns.
func Shuffle(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	index := 0
	for i := len(runes) - 1; i > 0; i-- {
		index = fastrand.Intn(i + 1)
		if i != index {
			runes[i], runes[index] = runes[index], runes[i]
		}
	}
	return string(runes)
}

// ContainsAnySubstrings returns whether s contains any of substring in slice.
func ContainsAnySubstrings(s string, subs []string) bool {
	for _, v := range subs {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}
