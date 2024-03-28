package dirtmake

import (
	"fmt"
	"testing"
)

var data []byte

const block1kb = 1024

func TestDirtBytes(b *testing.T) {
	bs := Bytes(block1kb*block1kb, block1kb*block1kb)
	_ = bs[block1kb*block1kb-1]
}

func BenchmarkDirtBytes(b *testing.B) {
	for size := block1kb; size < block1kb*20; size += block1kb * 2 {
		b.Run(fmt.Sprintf("size=%dkb", size/block1kb), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				data = Bytes(size, size)
			}
		})
	}
}

func BenchmarkOriginBytes(b *testing.B) {
	for size := block1kb; size < block1kb*20; size += block1kb * 2 {
		b.Run(fmt.Sprintf("size=%dkb", size/block1kb), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				data = make([]byte, size)
			}
		})
	}
}
