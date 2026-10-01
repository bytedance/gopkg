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

package wyhash

import "testing"

// wyhashV43Vectors are reference values produced by the reference C
// implementation (wangyi-fudan/wyhash, final version 4.3, little-endian
// flavor, WYHASH_CONDOM=1, default secret parameters) for the input
// `data[i] = byte(i % 256)`.
var wyhashV43Vectors = []struct {
	length int
	seed   uint64
	want   uint64
}{
	{0, 0x0000000000000000, 0x93228a4de0eec5a2},
	{1, 0x0000000000000000, 0x8e6d4af7d310c8c4},
	{2, 0x0000000000000000, 0x5121ba5bc9a828b5},
	{3, 0x0000000000000000, 0x78c4aa0c972a522d},
	{4, 0x0000000000000000, 0xe08aeeb68058fb32},
	{5, 0x0000000000000000, 0x845a2c5da2318785},
	{7, 0x0000000000000000, 0x094e98feb6055cc6},
	{8, 0x0000000000000000, 0xb4d6ac74d009e1d4},
	{9, 0x0000000000000000, 0xb42922e019b409be},
	{15, 0x0000000000000000, 0x87edaf96d89a08ef},
	{16, 0x0000000000000000, 0x305fdea0ed4a2619},
	{17, 0x0000000000000000, 0xd29ffdd201a46f9a},
	{23, 0x0000000000000000, 0x735a0e86527147f0},
	{31, 0x0000000000000000, 0xebc13906e5018315},
	{32, 0x0000000000000000, 0x5b00c06ef7540f8f},
	{33, 0x0000000000000000, 0x5e1a2536ff90cc32},
	{47, 0x0000000000000000, 0xe2cb58f6ab8e4419},
	{48, 0x0000000000000000, 0xecbfb7ff9e3d9a97},
	{49, 0x0000000000000000, 0x0691f11bac523a91},
	{63, 0x0000000000000000, 0x907220c8cff2e2c7},
	{64, 0x0000000000000000, 0xe0fe4c75f61d710d},
	{65, 0x0000000000000000, 0x74602394786a8035},
	{95, 0x0000000000000000, 0xa39f0af73a3eee99},
	{96, 0x0000000000000000, 0x948137d69794b570},
	{97, 0x0000000000000000, 0x2501575738d109be},
	{111, 0x0000000000000000, 0x869baaa70c652386},
	{127, 0x0000000000000000, 0x766e08dbb996f344},
	{128, 0x0000000000000000, 0x693d6d731c86b2ee},
	{129, 0x0000000000000000, 0x681978fad3f15d26},
	{191, 0x0000000000000000, 0x8884338d94de9161},
	{255, 0x0000000000000000, 0x530ea0c8a69bb188},
	{256, 0x0000000000000000, 0x139c96a974ad43cb},
	{511, 0x0000000000000000, 0x8924f6a37b55f704},
	{512, 0x0000000000000000, 0x07ee347c58decea0},
	{1024, 0x0000000000000000, 0x804dd1e6f38e5361},
	{0, 0x0000000000000007, 0x9411771484003547},
	{1, 0x0000000000000007, 0x0a393ec43f5b7188},
	{2, 0x0000000000000007, 0x599184b98ebd5fb2},
	{3, 0x0000000000000007, 0x1ce7ae7523ce9850},
	{4, 0x0000000000000007, 0x1bfbdc59121a5a5e},
	{5, 0x0000000000000007, 0x9a8ae5e04e535a5e},
	{7, 0x0000000000000007, 0xdbe2b92c7869d657},
	{8, 0x0000000000000007, 0xcf0e797cc1ba614f},
	{9, 0x0000000000000007, 0xabe1c1dd83c19301},
	{15, 0x0000000000000007, 0x4024b8cb2a1c907a},
	{16, 0x0000000000000007, 0xc4f3d94eba74b559},
	{17, 0x0000000000000007, 0x0caa8f7bd43c90b1},
	{23, 0x0000000000000007, 0xee8c173c4a1bee14},
	{31, 0x0000000000000007, 0x6243a9afcd2ad4b2},
	{32, 0x0000000000000007, 0x5b00f3162d5e24d2},
	{33, 0x0000000000000007, 0x3ac7d57c4d0203a4},
	{47, 0x0000000000000007, 0x3399dd337401228b},
	{48, 0x0000000000000007, 0x2d369a1639889127},
	{49, 0x0000000000000007, 0xd1fb644ed2bc0827},
	{63, 0x0000000000000007, 0xdf113507740e0ef6},
	{64, 0x0000000000000007, 0x076896c79ef3720f},
	{65, 0x0000000000000007, 0xd520e004fd27b57c},
	{95, 0x0000000000000007, 0xc365518fb3a9471b},
	{96, 0x0000000000000007, 0xf009852225d574e9},
	{97, 0x0000000000000007, 0x4efeb6ed5bdd7410},
	{111, 0x0000000000000007, 0x692489e75e4f0e1e},
	{127, 0x0000000000000007, 0xfbbc7131e283f541},
	{128, 0x0000000000000007, 0x93d412921e9d7169},
	{129, 0x0000000000000007, 0xd4c87f242b6c22bd},
	{191, 0x0000000000000007, 0x0ffdfb1d364ae528},
	{255, 0x0000000000000007, 0xfeb2464e2110f40d},
	{256, 0x0000000000000007, 0x098f80584458632c},
	{511, 0x0000000000000007, 0x7b10c8626afc8684},
	{512, 0x0000000000000007, 0x5d4ac09a024ed018},
	{1024, 0x0000000000000007, 0x542171872b2ca1c0},
	{0, 0xa0761d6478bd642f, 0xe4917dbfbc583a95},
	{1, 0xa0761d6478bd642f, 0x1ca2eb6cbfb870a3},
	{2, 0xa0761d6478bd642f, 0xde2d9b1e717acdce},
	{3, 0xa0761d6478bd642f, 0x2b176c4624aab54b},
	{4, 0xa0761d6478bd642f, 0xb9a3ac0519f2ded8},
	{5, 0xa0761d6478bd642f, 0x10a423505aab54c5},
	{7, 0xa0761d6478bd642f, 0x91229848b7bba473},
	{8, 0xa0761d6478bd642f, 0x20c1f8c95b3bf383},
	{9, 0xa0761d6478bd642f, 0x506ba1d0b049361d},
	{15, 0xa0761d6478bd642f, 0xacda1e5de65568c6},
	{16, 0xa0761d6478bd642f, 0x9f569f9b1f73ca48},
	{17, 0xa0761d6478bd642f, 0x8e71b8fd15f5d603},
	{23, 0xa0761d6478bd642f, 0x0eba40b34ec5e753},
	{31, 0xa0761d6478bd642f, 0xfb3bd33706ee62e9},
	{32, 0xa0761d6478bd642f, 0x8b8c451f37b90b17},
	{33, 0xa0761d6478bd642f, 0x9f57ced65d276185},
	{47, 0xa0761d6478bd642f, 0xfffbec2a51516439},
	{48, 0xa0761d6478bd642f, 0x14c44e7db1022b7c},
	{49, 0xa0761d6478bd642f, 0xcb9f289b75440af7},
	{63, 0xa0761d6478bd642f, 0x789ee1fbc5c48c24},
	{64, 0xa0761d6478bd642f, 0xe7f0f1a6b1074e1b},
	{65, 0xa0761d6478bd642f, 0x78960b76f8bef87b},
	{95, 0xa0761d6478bd642f, 0x4b0becc511ab4e48},
	{96, 0xa0761d6478bd642f, 0x9027125b0f663f12},
	{97, 0xa0761d6478bd642f, 0x701b984fa7753d45},
	{111, 0xa0761d6478bd642f, 0xf32b4215717e363a},
	{127, 0xa0761d6478bd642f, 0xcd9f6795d9fcc55b},
	{128, 0xa0761d6478bd642f, 0x7f3557c82a657b26},
	{129, 0xa0761d6478bd642f, 0xcebe70719a37efb8},
	{191, 0xa0761d6478bd642f, 0xa1e9cf60dc346e8c},
	{255, 0xa0761d6478bd642f, 0xda75cfa0f3ec4f48},
	{256, 0xa0761d6478bd642f, 0xeaaf92f386f07556},
	{511, 0xa0761d6478bd642f, 0x8754d4bb8142f5db},
	{512, 0xa0761d6478bd642f, 0x4e63a875b1ae5e36},
	{1024, 0xa0761d6478bd642f, 0x64a205fc777a1c95},
}

func TestWyhashV43Vectors(t *testing.T) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	for _, tc := range wyhashV43Vectors {
		got := Sum64StringWithSeed(string(data[:tc.length]), tc.seed)
		if got != tc.want {
			t.Fatalf("Sum64StringWithSeed(len=%d, seed=%#x) = %#x, want %#x", tc.length, tc.seed, got, tc.want)
		}
		// The streaming digest must agree for any write pattern.
		d := New(tc.seed)
		half := tc.length / 2
		d.Write(data[:half])
		d.Write(data[half:tc.length])
		if got := d.Sum64(); got != tc.want {
			t.Fatalf("Digest(len=%d, seed=%#x, split write) = %#x, want %#x", tc.length, tc.seed, got, tc.want)
		}
		d = New(tc.seed)
		for i := 0; i < tc.length; i += 7 {
			j := i + 7
			if j > tc.length {
				j = tc.length
			}
			d.Write(data[i:j])
		}
		if got := d.Sum64(); got != tc.want {
			t.Fatalf("Digest(len=%d, seed=%#x, 7-byte writes) = %#x, want %#x", tc.length, tc.seed, got, tc.want)
		}
	}
}
