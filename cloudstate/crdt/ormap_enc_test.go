package crdt

import (
	"testing"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
)

func BenchmarkORMapEncoding(b *testing.B) {
	b.Run("add and append decoded value", func(b *testing.B) {
		b.ReportAllocs()
		m := NewORMap()
		m.Set(encoding.Int64(int64(1)), NewLWWRegister(encoding.String("one")))
		m.Set(encoding.Int64(int64(2)), NewLWWRegister(encoding.String("two")))
		m.Set(encoding.Int64(int64(3)), NewLWWRegister(encoding.String("three")))
		m.Set(encoding.Int64(int64(4)), NewLWWRegister(encoding.String("four")))
		v := make([]string, 0, m.Size()*b.N)
		for i := 0; i < b.N; i++ {
			for _, state := range m.Entries() {
				v = append(v, encoding.DecodeString(state.Value.(*LWWRegister).Value()))
			}
		}
	})
	b.Run("add and get value", func(b *testing.B) {
		b.ReportAllocs()
		m := NewORMap()
		m.Set(encoding.Int64(int64(1)), NewLWWRegister(encoding.String("one")))
		m.Set(encoding.Int64(int64(2)), NewLWWRegister(encoding.String("two")))
		m.Set(encoding.Int64(int64(3)), NewLWWRegister(encoding.String("three")))
		m.Set(encoding.Int64(int64(4)), NewLWWRegister(encoding.String("four")))
		s0 := ""
		for i := 0; i < b.N; i++ {
			for _, state := range m.Entries() {
				// m.LWWRegister(state.Key)
				s0 = encoding.DecodeString(state.Value.(*LWWRegister).Value())
			}
		}
		_ = s0 == "" // use any0
	})
	b.Run("set and delete unknown", func(b *testing.B) {
		b.ReportAllocs()
		m := NewORMap()
		m.Set(encoding.Int64(int64(1)), NewLWWRegister(encoding.String("one")))
		m.Set(encoding.Int64(int64(2)), NewLWWRegister(encoding.String("two")))
		m.Set(encoding.Int64(int64(3)), NewLWWRegister(encoding.String("three")))
		m.Set(encoding.Int64(int64(4)), NewLWWRegister(encoding.String("four")))
		s0 := ""
		for i := 0; i < b.N; i++ {
			m.Set(encoding.Int64(int64(5)), NewLWWRegister(encoding.String("five")))
			m.Delete(encoding.Int64(int64(5)))
			m.Entries()
		}
		_ = s0 == "" // use any0
	})
	b.Run("set and delete known", func(b *testing.B) {
		b.ReportAllocs()
		m := NewORMap()
		m.Set(encoding.Int64(int64(1)), NewLWWRegister(encoding.String("one")))
		m.Set(encoding.Int64(int64(2)), NewLWWRegister(encoding.String("two")))
		m.Set(encoding.Int64(int64(3)), NewLWWRegister(encoding.String("three")))
		m.Set(encoding.Int64(int64(4)), NewLWWRegister(encoding.String("four")))
		s0 := ""
		for i := 0; i < b.N; i++ {
			m.Set(encoding.Int64(int64(1)), NewLWWRegister(encoding.String("one")))
			m.Delete(encoding.Int64(int64(1)))
			m.Entries()
		}
		_ = s0 == "" // use any0
	})
	b.Run("set and delete pre-encoded", func(b *testing.B) {
		b.ReportAllocs()
		m := NewORMap()
		m.Set(encoding.Int64(int64(1)), NewLWWRegister(encoding.String("one")))
		m.Set(encoding.Int64(int64(2)), NewLWWRegister(encoding.String("two")))
		m.Set(encoding.Int64(int64(3)), NewLWWRegister(encoding.String("three")))
		m.Set(encoding.Int64(int64(4)), NewLWWRegister(encoding.String("four")))
		s0 := ""
		one := encoding.String("one")
		oneInt := encoding.Int64(int64(1))
		for i := 0; i < b.N; i++ {
			m.Set(oneInt, NewLWWRegister(one))
			m.Delete(oneInt)
			m.Entries()
		}
		_ = s0 == "" // use any0
	})
}
