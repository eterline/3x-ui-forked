package ringbuffer

import "fmt"

type sErr string

func (e sErr) Error() string { return string(e) }

func newSErr(s string, p ...any) sErr {
	return sErr(fmt.Sprintf(s, p...))
}

// =============== Ring buffer errors ===============
const (
	ErrRingBufferInvalidSize sErr = "ring buffer: size must be above 0"
	ErrRingBufferOverflow    sErr = "ring buffer: overflow"
)
