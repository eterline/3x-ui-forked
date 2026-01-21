package fixwrite_test

import (
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v2/util/fixwrite"
	"github.com/mhsanaei/3x-ui/v2/util/random"
	"github.com/mhsanaei/3x-ui/v2/util/ringbuffer"
)

func BenchmarkFixedWriter(b *testing.B) {
	text := random.Seq(128)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slc := [512]byte{}
		// 0 B/op	       0 allocs/op
		wr := fixwrite.NewFixedWriter(slc[:])
		_, _ = wr.WriteString(text)
		_ = wr.String()
	}
}

func BenchmarkFixedWriter2(b *testing.B) {
	timeFormat := "2006/01/02 15:04:05"
	level := "ERROR"
	newLog := random.Seq(128)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := [512]byte{}
		wr := fixwrite.NewFixedWriter(buf[:])

		dst := wr.Reserve(64)
		dst = time.Now().AppendFormat(dst[:0], timeFormat)
		wr.SetPos(wr.Pos() - (64 - len(dst)))

		wr.WriteByte(' ')
		wr.WriteString(level)
		wr.WriteString(" - ")
		wr.WriteString(newLog)
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
