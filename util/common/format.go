package common

import (
	"fmt"
)

const (
	step = 1 << 10
)

var (
	units = []string{"B", "KB", "MB", "GB", "TB", "PB"}
)

// FormatTraffic formats traffic bytes into human-readable units (B, KB, MB, GB, TB, PB).
func FormatTraffic(trafficBytes int64) string {
	unitIndex := 0
	size := float64(trafficBytes)

	for size >= step && unitIndex < len(units)-1 {
		size /= step
		unitIndex++
	}
	return fmt.Sprintf("%.2f%s", size, units[unitIndex])
}
