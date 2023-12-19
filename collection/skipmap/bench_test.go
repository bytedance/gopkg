package skipmap

import (
	"testing"

	"github.com/bytedance/gopkg/lang/fastrand"
)

func BenchmarkLoadOrStoreExist(b *testing.B) {
	m := NewInt()
	m.Store(1, 1)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			m.LoadOrStore(1, 1)
		}
	})
}

func BenchmarkLoadOrStoreLazyExist(b *testing.B) {
	m := NewInt()
	m.Store(1, 1)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			m.LoadOrStoreLazy(1, func() interface{} { return 1 })
		}
	})
}

func BenchmarkLoadOrStoreExistSingle(b *testing.B) {
	m := NewInt()
	m.Store(1, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.LoadOrStore(1, 1)
	}
}

func BenchmarkLoadOrStoreLazyExistSingle(b *testing.B) {
	m := NewInt()
	m.Store(1, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.LoadOrStoreLazy(1, func() interface{} { return 1 })
	}
}

func BenchmarkLoadOrStoreRandom(b *testing.B) {
	m := NewInt()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			m.LoadOrStore(fastrand.Int(), 1)
		}
	})
}

func BenchmarkLoadOrStoreLazyRandom(b *testing.B) {
	m := NewInt()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			m.LoadOrStoreLazy(fastrand.Int(), func() interface{} { return 1 })
		}
	})
}

func BenchmarkLoadOrStoreRandomSingle(b *testing.B) {
	m := NewInt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.LoadOrStore(fastrand.Int(), 1)
	}
}

func BenchmarkLoadOrStoreLazyRandomSingle(b *testing.B) {
	m := NewInt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.LoadOrStoreLazy(fastrand.Int(), func() interface{} { return 1 })
	}
}
