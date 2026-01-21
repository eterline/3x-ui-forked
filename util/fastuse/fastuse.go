package fastuse

import "unsafe"

func Bytes2String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func TrimZerosLeft(b []byte) []byte {
	l := 0
	for l < len(b) && b[l] == 0 {
		l++
	}

	return b[l:]
}

func TrimZerosRight(b []byte) []byte {
	r := len(b)
	for r > 0 && b[r-1] == 0 {
		r--
	}

	return b[:r]
}

func TrimZeros(b []byte) []byte {
	b = TrimZerosLeft(b)
	b = TrimZerosRight(b)
	return b
}

func ReverseSlice[T any](a []T) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}
