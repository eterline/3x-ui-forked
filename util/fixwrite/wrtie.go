package fixwrite

import "errors"

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

func (w *FixedWriter) WriteByte(b byte) error {
	if w.pos >= len(w.buf) {
		return ErrBufferOverflow
	}

	w.buf[w.pos] = b
	w.pos++
	return nil
}
