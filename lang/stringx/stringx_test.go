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
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/bytedance/gopkg/internal/assert"
)

func TestPad(t *testing.T) {
	type testData struct {
		input             string
		padChar           rune
		size              int
		leftExpected      string
		leftExpectedSpace string

		rightExpected      string
		rightExpectedSpace string

		centerExpected      string
		centerExpectedSpace string
	}

	testCases := []testData{
		{
			"", '-', 4,
			"----", "    ",
			"----", "    ",
			"----", "    ",
		},
		{
			"abc", '-', 0,
			"abc", "abc",
			"abc", "abc",
			"abc", "abc",
		},
		{
			"abc", '-', 2,
			"abc", "abc",
			"abc", "abc",
			"abc", "abc",
		},
		{
			"abc", '-', 4,
			"-abc", " abc",
			"abc-", "abc ",
			"abc-", "abc ",
		},
		{
			"abc", '-', 5,
			"--abc", "  abc",
			"abc--", "abc  ",
			"-abc-", " abc ",
		},
		{
			"abc", '-', 6,
			"---abc", "   abc",
			"abc---", "abc   ",
			"-abc--", " abc  ",
		},
		{
			"abc", '-', 7,
			"----abc", "    abc",
			"abc----", "abc    ",
			"--abc--", "  abc  ",
		},

		{
			"abcd", '-', 7,
			"---abcd", "   abcd",
			"abcd---", "abcd   ",
			"-abcd--", " abcd  ",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.leftExpected, PadLeftChar(testCase.input, testCase.size, testCase.padChar))
		assert.Equal(t, testCase.leftExpectedSpace, PadLeftSpace(testCase.input, testCase.size))

		assert.Equal(t, testCase.rightExpected, PadRightChar(testCase.input, testCase.size, testCase.padChar))
		assert.Equal(t, testCase.rightExpectedSpace, PadRightSpace(testCase.input, testCase.size))

		assert.Equal(t, testCase.centerExpected, PadCenterChar(testCase.input, testCase.size, testCase.padChar))
		assert.Equal(t, testCase.centerExpectedSpace, PadCenterSpace(testCase.input, testCase.size))
	}
}

func TestRemove(t *testing.T) {
	assert.Equal(t, "", RemoveChar("", 'h'))
	assert.Equal(t, "z英文un排", RemoveChar("zh英文hunh排", 'h'))
	assert.Equal(t, "zh英hun排", RemoveChar("zh英文hun文排", '文'))

	assert.Equal(t, "", RemoveString("", "文hun"))
	assert.Equal(t, "zh英文hun排", RemoveString("zh英文hun排", ""))
	assert.Equal(t, "zh英排", RemoveString("zh英文hun排", "文hun"))
	assert.Equal(t, "zh英文hun排", RemoveString("zh英文hun排", ""))
}

func TestRepeat(t *testing.T) {
	assert.Equal(t, "", RepeatChar('-', 0))
	assert.Equal(t, "----", RepeatChar('-', 4))
	assert.Equal(t, "   ", RepeatChar(' ', 3))
}

func TestRotate(t *testing.T) {
	assert.Equal(t, "", Rotate("", 2))

	assert.Equal(t, "abc", Rotate("abc", 0))
	assert.Equal(t, "abc", Rotate("abc", 3))
	assert.Equal(t, "abc", Rotate("abc", 6))

	assert.Equal(t, "cab", Rotate("abc", 1))
	assert.Equal(t, "bca", Rotate("abc", -1))
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abc", "cba"},
		{"a", "a"},
		{"çınar", "ranıç"},
		{"    yağmur", "rumğay    "},
		{"επαγγελματίες", "ςείταμλεγγαπε"},
	}
	for _, test := range tests {
		output := MustReverse(test.input)
		if test.expected != output {
			t.Fatalf("test case %s is not successful: expected %#v, got %#v", test.input, test.expected, output)
		}
	}

	assert.Equal(t, MustReverse(""), "")
	assert.Equal(t, MustReverse("X"), "X")
	assert.Equal(t, MustReverse("\u0301b"), "b\u0301")
	assert.Equal(t, MustReverse("😎⚽"), "⚽😎")
	assert.Equal(t, MustReverse("Les Mise\u0301rables"), "selbar\u0301esiM seL")
	assert.Equal(t, MustReverse("ab\u0301cde"), "edc\u0301ba")
	assert.Equal(t, MustReverse("The quick bròwn 狐 jumped over the lazy 犬"), "犬 yzal eht revo depmuj 狐 nwòrb kciuq ehT")
	_, err := Reverse(string([]byte{128, 128, 128, 128, 0}))
	assert.Equal(t, ErrDecodeRune, err)
}

func TestSub(t *testing.T) {
	type testData struct {
		input    string
		start    int
		end      int
		expected string
	}

	newTestCase := func(input string, start, end int, expected string) testData {
		return testData{
			input:    input,
			start:    start,
			end:      end,
			expected: expected,
		}
	}

	testCases := []testData{
		newTestCase("", 0, 100, ""),
		newTestCase("facgbheidjk", 3, 9, "gbheid"),
		newTestCase("facgbheidjk", -50, 100, "facgbheidjk"),
		newTestCase("facgbheidjk", -3, utf8.RuneCountInString("facgbheidjk"), "djk"),
		newTestCase("facgbheidjk", -3, -1, "dj"),
		newTestCase("zh英文hun排", 2, 5, "英文h"),
		newTestCase("zh英文hun排", 2, -1, "英文hun"),
		newTestCase("zh英文hun排", -100, -1, "zh英文hun"),
		newTestCase("zh英文hun排", -100, -90, ""),
		newTestCase("zh英文hun排", -10, -90, ""),
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expected, Sub(testCase.input, testCase.start, testCase.end))
	}
}

func TestContainsAnySubstrings(t *testing.T) {
	assert.True(t, ContainsAnySubstrings("abcdefg", []string{"a", "b"}))
	assert.True(t, ContainsAnySubstrings("abcdefg", []string{"a", "z"}))
	assert.False(t, ContainsAnySubstrings("abcdefg", []string{"ac", "z"}))
	assert.False(t, ContainsAnySubstrings("abcdefg", []string{"x", "z"}))
}

func TestShuffle(t *testing.T) {
	shuffleAndSort := func(str string) string {
		s := Shuffle(str)
		slice := sort.StringSlice(strings.Split(s, ""))
		slice.Sort()
		return strings.Join(slice, "")
	}

	strMap := map[string]string{
		"":            "",
		"facgbheidjk": "abcdefghijk",
		"尝试中文":        "中尝文试",
		"zh英文hun排":    "hhnuz排文英",
	}
	for input, expected := range strMap {
		actual := shuffleAndSort(input)
		assert.Equal(t, expected, actual)
	}
}
