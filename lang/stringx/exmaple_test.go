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
	"fmt"
	"unicode/utf8"
)

func Example_sub() {
	fmt.Printf("Sub-[0:100]=%s\n", Sub("", 0, 100))
	fmt.Printf("Sub-facgbheidjk[3:9]=%s\n", Sub("facgbheidjk", 3, 9))
	fmt.Printf("Sub-facgbheidjk[-50:100]=%s\n", Sub("facgbheidjk", -50, 100))
	fmt.Printf("Sub-facgbheidjk[-3:length]=%s\n", Sub("facgbheidjk", -3, utf8.RuneCountInString("facgbheidjk")))
	fmt.Printf("Sub-facgbheidjk[-3:-1]=%s\n", Sub("facgbheidjk", -3, -1))
	fmt.Printf("Sub-zh英文hun排[2:5]=%s\n", Sub("zh英文hun排", 2, 5))
	fmt.Printf("Sub-zh英文hun排[2:-1]=%s\n", Sub("zh英文hun排", 2, -1))
	fmt.Printf("Sub-zh英文hun排[-100:-1]=%s\n", Sub("zh英文hun排", -100, -1))
	fmt.Printf("Sub-zh英文hun排[-100:-90]=%s\n", Sub("zh英文hun排", -100, -90))
	fmt.Printf("Sub-zh英文hun排[-10:-90]=%s\n", Sub("zh英文hun排", -10, -90))

	// Output:
	// Sub-[0:100]=
	// Sub-facgbheidjk[3:9]=gbheid
	// Sub-facgbheidjk[-50:100]=facgbheidjk
	// Sub-facgbheidjk[-3:length]=djk
	// Sub-facgbheidjk[-3:-1]=dj
	// Sub-zh英文hun排[2:5]=英文h
	// Sub-zh英文hun排[2:-1]=英文hun
	// Sub-zh英文hun排[-100:-1]=zh英文hun
	// Sub-zh英文hun排[-100:-90]=
	// Sub-zh英文hun排[-10:-90]=
}

func Example_substart() {
	fmt.Printf("SubStart-[0:]=%s\n", SubStart("", 0))
	fmt.Printf("SubStart-[2:]=%s\n", SubStart("", 2))
	fmt.Printf("SubStart-facgbheidjk[3:]=%s\n", SubStart("facgbheidjk", 3))
	fmt.Printf("SubStart-facgbheidjk[-50:]=%s\n", SubStart("facgbheidjk", -50))
	fmt.Printf("SubStart-facgbheidjk[-3:]=%s\n", SubStart("facgbheidjk", -3))
	fmt.Printf("SubStart-zh英文hun排[3:]=%s\n", SubStart("zh英文hun排", 3))

	// Output:
	// SubStart-[0:]=
	// SubStart-[2:]=
	// SubStart-facgbheidjk[3:]=gbheidjk
	// SubStart-facgbheidjk[-50:]=facgbheidjk
	// SubStart-facgbheidjk[-3:]=djk
	// SubStart-zh英文hun排[3:]=文hun排
}

func Example_pad() {

	fmt.Printf("PadLeft=[%s]\n", PadLeftSpace("abc", 7))
	fmt.Printf("PadLeft=[%s]\n", PadLeftChar("abc", 7, '-'))
	fmt.Printf("PadCenter=[%s]\n", PadCenterChar("abc", 7, '-'))
	fmt.Printf("PadCenter=[%s]\n", PadCenterChar("abcd", 7, '-'))

	// Output:
	// PadLeft=[    abc]
	// PadLeft=[----abc]
	// PadCenter=[--abc--]
	// PadCenter=[-abcd--]
}
