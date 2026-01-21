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

func (w *FixedWriter) Pos() int {
	return w.pos
}

func (w *FixedWriter) Reserve(n int) []byte {
	start := w.pos
	if w.pos+n > len(w.buf) {
		w.pos = len(w.buf)
		return w.buf[start:]
	}
	w.pos += n
	return w.buf[start:w.pos]
}

func (w *FixedWriter) SetPos(n int) {
	if n >= 0 {
		w.pos = n
	}
}

func (w *FixedWriter) Skip(n int) (int, error) {
	if n <= 0 {
		return 0, nil
	}

	remain := len(w.buf) - w.pos
	if n > remain {
		w.pos += remain
		return remain, ErrBufferOverflow
	}

	w.pos += n
	return n, nil
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
