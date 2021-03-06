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

package xxhash3

import "unsafe"

const (
	_stripe = 64
	_block  = 1024

	prime32_1 = 2654435761
	prime32_2 = 2246822519
	prime32_3 = 3266489917

	prime64_1 = 11400714785074694791
	prime64_2 = 14029467366897019727
	prime64_3 = 1609587929392839161
	prime64_4 = 9650029242287828579
	prime64_5 = 2870177450012600261

	xsecret_000 uint64 = 0xbe4ba423396cfeb8
	xsecret_008 uint64 = 0x1cad21f72c81017c
	xsecret_016 uint64 = 0xdb979083e96dd4de
	xsecret_024 uint64 = 0x1f67b3b7a4a44072
	xsecret_032 uint64 = 0x78e5c0cc4ee679cb
	xsecret_040 uint64 = 0x2172ffcc7dd05a82
	xsecret_048 uint64 = 0x8e2443f7744608b8
	xsecret_056 uint64 = 0x4c263a81e69035e0
	xsecret_064 uint64 = 0xcb00c391bb52283c
	xsecret_072 uint64 = 0xa32e531b8b65d088
	xsecret_080 uint64 = 0x4ef90da297486471
	xsecret_088 uint64 = 0xd8acdea946ef1938
	xsecret_096 uint64 = 0x3f349ce33f76faa8
	xsecret_104 uint64 = 0x1d4f0bc7c7bbdcf9
	xsecret_112 uint64 = 0x3159b4cd4be0518a
	xsecret_120 uint64 = 0x647378d9c97e9fc8
	xsecret_128 uint64 = 0xc3ebd33483acc5ea
	xsecret_136 uint64 = 0xeb6313faffa081c5
	xsecret_144 uint64 = 0x49daf0b751dd0d17
	xsecret_152 uint64 = 0x9e68d429265516d3
	xsecret_160 uint64 = 0xfca1477d58be162b
	xsecret_168 uint64 = 0xce31d07ad1b8f88f
	xsecret_176 uint64 = 0x280416958f3acb45
	xsecret_184 uint64 = 0x7e404bbbcafbd7af

	xsecret_011 uint64 = 0x6dd4de1cad21f72c
	xsecret_019 uint64 = 0xa44072db979083e9
	xsecret_027 uint64 = 0xe679cb1f67b3b7a4
	xsecret_035 uint64 = 0xd05a8278e5c0cc4e
	xsecret_043 uint64 = 0x4608b82172ffcc7d
	xsecret_051 uint64 = 0x9035e08e2443f774
	xsecret_059 uint64 = 0x52283c4c263a81e6
	xsecret_067 uint64 = 0x65d088cb00c391bb

	xsecret_103 uint64 = 0x4f0bc7c7bbdcf93f
	xsecret_111 uint64 = 0x59b4cd4be0518a1d
	xsecret_119 uint64 = 0x7378d9c97e9fc831
	xsecret_127 uint64 = 0xebd33483acc5ea64

	xsecret_117 uint64 = 0xd9c97e9fc83159b4
	xsecret_125 uint64 = 0x3483acc5ea647378
	xsecret_133 uint64 = 0xfaffa081c5c3ebd3
	xsecret_141 uint64 = 0xb751dd0d17eb6313
	xsecret_149 uint64 = 0x29265516d349daf0
	xsecret_157 uint64 = 0x7d58be162b9e68d4
	xsecret_165 uint64 = 0x7ad1b8f88ffca147
	xsecret_173 uint64 = 0x958f3acb45ce31d0

	xsecret32_000 uint32 = 0xbe4ba423
	xsecret32_004 uint32 = 0x396cfeb8
	xsecret32_008 uint32 = 0x1cad21f7
	xsecret32_012 uint32 = 0x2c81017c
)

var xsecret = unsafe.Pointer(&[192]uint8{
	/* 	0 	*/ 0xb8, 0xfe, 0x6c, 0x39, 0x23, 0xa4, 0x4b, 0xbe, 0x7c, 0x01, 0x81, 0x2c, 0xf7, 0x21, 0xad, 0x1c,
	/* 	16 	*/ 0xde, 0xd4, 0x6d, 0xe9, 0x83, 0x90, 0x97, 0xdb, 0x72, 0x40, 0xa4, 0xa4, 0xb7, 0xb3, 0x67, 0x1f,
	/* 	32 	*/ 0xcb, 0x79, 0xe6, 0x4e, 0xcc, 0xc0, 0xe5, 0x78, 0x82, 0x5a, 0xd0, 0x7d, 0xcc, 0xff, 0x72, 0x21,
	/* 	48 	*/ 0xb8, 0x08, 0x46, 0x74, 0xf7, 0x43, 0x24, 0x8e, 0xe0, 0x35, 0x90, 0xe6, 0x81, 0x3a, 0x26, 0x4c,
	/* 	64 	*/ 0x3c, 0x28, 0x52, 0xbb, 0x91, 0xc3, 0x00, 0xcb, 0x88, 0xd0, 0x65, 0x8b, 0x1b, 0x53, 0x2e, 0xa3,
	/* 	80 	*/ 0x71, 0x64, 0x48, 0x97, 0xa2, 0x0d, 0xf9, 0x4e, 0x38, 0x19, 0xef, 0x46, 0xa9, 0xde, 0xac, 0xd8,
	/* 	96 	*/ 0xa8, 0xfa, 0x76, 0x3f, 0xe3, 0x9c, 0x34, 0x3f, 0xf9, 0xdc, 0xbb, 0xc7, 0xc7, 0x0b, 0x4f, 0x1d,
	/* 	112	*/ 0x8a, 0x51, 0xe0, 0x4b, 0xcd, 0xb4, 0x59, 0x31, 0xc8, 0x9f, 0x7e, 0xc9, 0xd9, 0x78, 0x73, 0x64,
	/* 	128	*/ 0xea, 0xc5, 0xac, 0x83, 0x34, 0xd3, 0xeb, 0xc3, 0xc5, 0x81, 0xa0, 0xff, 0xfa, 0x13, 0x63, 0xeb,
	/* 	144	*/ 0x17, 0x0d, 0xdd, 0x51, 0xb7, 0xf0, 0xda, 0x49, 0xd3, 0x16, 0x55, 0x26, 0x29, 0xd4, 0x68, 0x9e,
	/* 	160	*/ 0x2b, 0x16, 0xbe, 0x58, 0x7d, 0x47, 0xa1, 0xfc, 0x8f, 0xf8, 0xb8, 0xd1, 0x7a, 0xd0, 0x31, 0xce,
	/* 	176	*/ 0x45, 0xcb, 0x3a, 0x8f, 0x95, 0x16, 0x04, 0x28, 0xaf, 0xd7, 0xfb, 0xca, 0xbb, 0x4b, 0x40, 0x7e,
})
