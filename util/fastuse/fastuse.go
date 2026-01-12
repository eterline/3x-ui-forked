package fastuse

import "unsafe"

func Bytes2String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func TrimZeros(b []byte) []byte {
	l := 0
	for l < len(b) && b[l] == 0 {
		l++
	}

	r := len(b)
	for r > l && b[r-1] == 0 {
		r--
	}

	return b[l:r]
}
