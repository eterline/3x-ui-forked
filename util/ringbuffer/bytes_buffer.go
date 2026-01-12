package ringbuffer

import (
	"io"
	"sync"
)

type ByteRing struct {
	buffer   []byte
	size     int64
	writePos int64
	headPos  int64
	mu       sync.RWMutex
}

// NewByteRing - creates io.ReadWriter byte buffer for stream usage
func NewByteRing(size int) *ByteRing {
	if size < 1 {
		panic(ErrRingBufferInvalidSize)
	}

	return &ByteRing{
		buffer: make([]byte, size),
		size:   int64(size),
	}
}

// Len - returns buffer len
func (r *ByteRing) Len() int {
	if r.writePos < r.headPos {
		return int(r.size - r.headPos + r.writePos)
	}

	return int(r.writePos - r.headPos)
}

// Write - io.Writer implementation byte steram writing
func (r *ByteRing) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	length := int64(len(p))
	size := r.size

	if length >= size {
		p = p[length-size:]
		length = size

		copy(r.buffer, p)
		r.headPos = 0
		r.writePos = 0

		return int(length), nil
	}

	wp := r.writePos
	head := r.headPos

	var used int64
	if wp >= head {
		used = wp - head
	} else {
		used = size - head + wp
	}

	free := size - used

	if length > free {
		overwrite := length - free
		r.headPos = (head + overwrite) % size
	}

	if wp+length <= size {
		copy(r.buffer[wp:], p)
		r.writePos = (wp + length) % size
	} else {
		split := size - wp
		copy(r.buffer[wp:], p[:split])
		copy(r.buffer, p[split:])
		r.writePos = length - split
	}

	return int(length), nil
}

// Read - io.Reader implementation byte steram reading
func (r *ByteRing) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	length := int64(len(p))

	head := r.headPos
	write := r.writePos

	if head == write {
		return 0, io.EOF
	}

	var size int64
	if head < write {

		size = write - head
		if length < size {
			size = length
		}
		copy(p, r.buffer[head:head+size])

	} else if r.size-head >= length {

		size = length
		copy(p, r.buffer[head:head+size])

	} else { // split needed

		size = r.size + write - head
		if length < size {
			size = length
		}

		right := r.size - head
		left := size - right

		copy(p, r.buffer[head:])
		copy(p[right:], r.buffer[:left])
	}

	r.headPos = (head + size) % r.size

	return int(size), nil
}

func (r *ByteRing) Bytes() []byte {
	r.mu.RLock()
	defer r.mu.RUnlock()

	head := r.headPos
	write := r.writePos

	return r.buffer[head:write]
}
