package fastuse

import (
	"errors"
	"io"
	"sync"

	"golang.org/x/exp/mmap"
)

type MMapReadCloser struct {
	r   *mmap.ReaderAt
	pos int64
	mu  sync.Mutex
}

func OpenMMapReadCloser(path string) (*MMapReadCloser, error) {
	r, err := mmap.Open(path)
	if err != nil {
		return nil, err
	}

	return &MMapReadCloser{
		r: r,
	}, nil
}

func (m *MMapReadCloser) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	n, err := m.r.ReadAt(p, m.pos)
	m.pos += int64(n)

	if err == io.EOF && n > 0 {
		return n, nil
	}

	return n, err
}

func (m *MMapReadCloser) Seek(offset int64, whence int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var newPos int64

	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = m.pos + offset
	case io.SeekEnd:
		newPos = int64(m.r.Len()) + offset
	default:
		return 0, errors.New("invalid whence")
	}

	if newPos < 0 {
		return 0, io.ErrUnexpectedEOF
	}

	m.pos = newPos
	return m.pos, nil
}

func (m *MMapReadCloser) Pos() int64 {
	return m.pos
}

func (m *MMapReadCloser) Close() error {
	err := m.r.Close()
	*m = MMapReadCloser{}
	return err
}
