package ringbuffer

import "io"

type ByteRing struct {
	buffer   []byte
	size     int64
	writePos int64
	headPos  int64
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

	length := int64(len(p))

	wp := r.writePos
	head := r.headPos

	if wp < head && head-wp <= length {
		return 0, ErrRingBufferOverflow
	}

	if wp >= head && head+r.size-wp <= length {
		return 0, ErrRingBufferOverflow
	}

	if wp < head || r.size-wp >= length { // no split needed

		copy(r.buffer[wp:], p)
		r.writePos = (wp + length) % r.size

		return int(length), nil

	} else { // split needed not enough contiguous memory

		split := r.size - wp
		left := length - split

		copy(r.buffer[wp:], p[:split])
		copy(r.buffer, p[split:])

		r.writePos = left
		return int(length), nil
	}
}

// Read - io.Reader implementation byte steram reading
func (r *ByteRing) Read(p []byte) (n int, err error) {

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

func (r *ByteRing) Bytes2() (a, b []byte) {
	head := r.headPos
	write := r.writePos

	if head == write {
		return nil, nil
	}

	if head < write {
		return r.buffer[head:write], nil
	}

	return r.buffer[head:], r.buffer[:write]
}
