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

	"github.com/stretchr/testify/assert"
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

	is := assert.New(t)
	for _, testCase := range testCases {
		is.Equal(testCase.leftExpected, PadLeftChar(testCase.input, testCase.size, testCase.padChar))
		is.Equal(testCase.leftExpectedSpace, PadLeftSpace(testCase.input, testCase.size))

		is.Equal(testCase.rightExpected, PadRightChar(testCase.input, testCase.size, testCase.padChar))
		is.Equal(testCase.rightExpectedSpace, PadRightSpace(testCase.input, testCase.size))

		is.Equal(testCase.centerExpected, PadCenterChar(testCase.input, testCase.size, testCase.padChar))
		is.Equal(testCase.centerExpectedSpace, PadCenterSpace(testCase.input, testCase.size))
	}
}

func TestRemove(t *testing.T) {
	is := assert.New(t)
	is.Equal("", RemoveChar("", 'h'))
	is.Equal("z??????un???", RemoveChar("zh??????hunh???", 'h'))
	is.Equal("zh???hun???", RemoveChar("zh??????hun??????", '???'))

	is.Equal("", RemoveString("", "???hun"))
	is.Equal("zh??????hun???", RemoveString("zh??????hun???", ""))
	is.Equal("zh??????", RemoveString("zh??????hun???", "???hun"))
	is.Equal("zh??????hun???", RemoveString("zh??????hun???", ""))
}

func TestRepeat(t *testing.T) {
	is := assert.New(t)
	is.Equal("", RepeatChar('-', 0))
	is.Equal("----", RepeatChar('-', 4))
	is.Equal("   ", RepeatChar(' ', 3))
}

func TestRotate(t *testing.T) {
	is := assert.New(t)

	is.Equal("", Rotate("", 2))

	is.Equal("abc", Rotate("abc", 0))
	is.Equal("abc", Rotate("abc", 3))
	is.Equal("abc", Rotate("abc", 6))

	is.Equal("cab", Rotate("abc", 1))
	is.Equal("bca", Rotate("abc", -1))
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abc", "cba"},
		{"a", "a"},
		{"????nar", "ran????"},
		{"    ya??mur", "rum??ay    "},
		{"??????????????????????????", "??????????????????????????"},
	}
	for _, test := range tests {
		output := MustReverse(test.input)
		assert.Equalf(t, test.expected, output, "Test case %s is not successful\n", test.input)
	}

	assert.Equal(t, MustReverse(""), "")
	assert.Equal(t, MustReverse("X"), "X")
	assert.Equal(t, MustReverse("\u0301b"), "b\u0301")
	assert.Equal(t, MustReverse("???????"), "???????")
	assert.Equal(t, MustReverse("Les Mise\u0301rables"), "selbar\u0301esiM seL")
	assert.Equal(t, MustReverse("ab\u0301cde"), "edc\u0301ba")
	assert.Equal(t, MustReverse("The quick br??wn ??? jumped over the lazy ???"), "??? yzal eht revo depmuj ??? nw??rb kciuq ehT")
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

	newTestCase := func(intput string, start, end int, expected string) testData {
		return testData{
			input:    intput,
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
		newTestCase("zh??????hun???", 2, 5, "??????h"),
		newTestCase("zh??????hun???", 2, -1, "??????hun"),
		newTestCase("zh??????hun???", -100, -1, "zh??????hun"),
		newTestCase("zh??????hun???", -100, -90, ""),
		newTestCase("zh??????hun???", -10, -90, ""),
	}

	is := assert.New(t)
	for _, testCase := range testCases {
		is.Equal(testCase.expected, Sub(testCase.input, testCase.start, testCase.end))
	}
}

func TestContainsAnySubstrings(t *testing.T) {
	is := assert.New(t)
	is.True(ContainsAnySubstrings("abcdefg", []string{"a", "b"}))
	is.True(ContainsAnySubstrings("abcdefg", []string{"a", "z"}))
	is.False(ContainsAnySubstrings("abcdefg", []string{"ac", "z"}))
	is.False(ContainsAnySubstrings("abcdefg", []string{"x", "z"}))
}

func TestShuffle(t *testing.T) {
	is := assert.New(t)

	shuffleAndSort := func(str string) string {
		s := Shuffle(str)
		slice := sort.StringSlice(strings.Split(s, ""))
		slice.Sort()
		return strings.Join(slice, "")
	}

	strMap := map[string]string{
		"":            "",
		"facgbheidjk": "abcdefghijk",
		"????????????":        "????????????",
		"zh??????hun???":    "hhnuz?????????",
	}
	for input, expected := range strMap {
		actual := shuffleAndSort(input)
		is.Equal(expected, actual)
	}
}
