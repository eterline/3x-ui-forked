package fixwrite

import (
	"errors"
	"io"

	"github.com/mhsanaei/3x-ui/v2/util/fastuse"
)

var ErrBufferOverflow = errors.New("buffer overflow")

type FixedWriter struct {
	buf []byte
	pos int
}

func NewFixedWriter(b []byte) *FixedWriter {
	return &FixedWriter{buf: b}
}

func (w *FixedWriter) Write(p []byte) (int, error) {
	if w.pos+len(p) > len(w.buf) {
		n := copy(w.buf[w.pos:], p)
		w.pos += n
		return n, ErrBufferOverflow
	}

	copy(w.buf[w.pos:], p)
	w.pos += len(p)
	return len(p), nil
}

func (w *FixedWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *FixedWriter) WriteByte(b byte) error {
	if w.pos >= len(w.buf) {
		return ErrBufferOverflow
	}

	w.buf[w.pos] = b
	w.pos++
	return nil
}

func (w *FixedWriter) Bytes() []byte {
	return w.buf
}

func (w *FixedWriter) String() string {
	return fastuse.Bytes2String(w.buf)
}

func (w *FixedWriter) WriteTo(wr io.Writer) (n int64, err error) {
	c, err := wr.Write(w.buf)
	return int64(c), err
}
