package fastuse

import "unsafe"

func Bytes2String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
