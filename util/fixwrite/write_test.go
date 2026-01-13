package fixwrite_test

import (
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v2/util/fixwrite"
	"github.com/mhsanaei/3x-ui/v2/util/ringbuffer"
)

func BenchmarkFixedWriter(b *testing.B) {
	text := []byte("ldp[oewp9orfker9okfoerkgf0[ermrg]]")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slc := [512]byte{}
		// 0 B/op	       0 allocs/op
		wr := fixwrite.NewFixedWriter(slc[:])
		_, _ = wr.Write(text)
	}
}

func BenchmarkFixedWriter2(b *testing.B) {
	timeFormat := "2006/01/02 15:04:05"
	level := "ERROR"
	newLog := "ldp[oewp9orfker9okfoerkgf0[ermrg]]"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := [512]byte{}
		wr := fixwrite.NewFixedWriter(buf[:])

		wr.Write([]byte(time.Now().Format(timeFormat))) // allocs there
		wr.WriteByte(' ')
		wr.Write([]byte(level))
		wr.Write([]byte(" - "))
		wr.Write([]byte(newLog))
	}
}

func BenchmarkFixedWriterWithRingbuf(b *testing.B) {
	timeFormat := "2006/01/02 15:04:05"
	level := "ERROR"
	newLog := "ldp[oewp9orfker9okfoerkgf0[ermrg]]"
	buf := ringbuffer.NewByteRing(1 << 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		in := [512]byte{}
		wr := fixwrite.NewFixedWriter(in[:])

		wr.WriteString(time.Now().Format(timeFormat)) // allocs there
		wr.WriteString(" ")
		wr.WriteString(level)
		wr.WriteString(" - ")
		wr.WriteString(newLog)

		wr.WriteTo(buf)
	}
}
