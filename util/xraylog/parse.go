package xraylog

import (
	"bytes"
	"strings"
	"time"
)

var (
	apiSubBytes      = []byte("api -> api")
	fromSubBytes     = []byte("from ")
	acceptedSubBytes = []byte("accepted ")
	emailSubBytes    = []byte("email: ")
)

func RecordRoute(record []byte) (inbound, outbound string, ok bool) {
	routeStart := bytes.LastIndexByte(record, '[')
	if routeStart == -1 {
		return "", "", false
	}

	routeEnd := bytes.LastIndexByte(record[routeStart:], ']')
	if routeEnd == -1 {
		return "", "", false
	}

	routeFields := bytes.Fields(record[routeStart+1 : routeStart+routeEnd])
	if len(routeFields) != 3 {
		return "", "", false
	}

	inbound = string(routeFields[0])
	outbound = string(routeFields[2])

	return inbound, outbound, true
}

func RecordTimestamp(record []byte) (time.Time, bool) {
	recordParts := bytes.Fields(record)
	if len(recordParts) < 2 {
		return time.Time{}, false
	}

	partDate := recordParts[0]
	partTime := recordParts[1]

	var b strings.Builder
	b.Grow(len(partDate) + 1 + len(partTime))
	b.Write(partDate)
	b.WriteByte(' ')
	b.Write(partTime)
	dateTimeStr := b.String()

	dateTime, err := time.ParseInLocation("2006/01/02 15:04:05.999999", dateTimeStr, time.Local)
	if err != nil {
		return time.Time{}, false
	}

	return dateTime.UTC(), true
}

func RecordAddrs(record []byte) (from, to string, ok bool) {
	var fromOk, toOk bool

	if lookIdx := bytes.Index(record, fromSubBytes); lookIdx != -1 {
		fromStart := lookIdx + len(fromSubBytes)
		fromEnd := bytes.IndexByte(record[fromStart:], ' ')
		if fromEnd != -1 {
			from = string(record[fromStart : fromStart+fromEnd])
			fromOk = true
		}
	}

	if accIdx := bytes.Index(record, acceptedSubBytes); accIdx != -1 {
		accStart := accIdx + len(acceptedSubBytes)
		accEnd := bytes.IndexByte(record[accStart:], ' ')
		if accEnd == -1 {
			accEnd = len(record) - accStart
		}
		to = string(record[accStart : accStart+accEnd])
		toOk = true
	}

	ok = fromOk && toOk
	return
}

func RecordEmail(record []byte) (email string, ok bool) {
	if lookIdx := bytes.Index(record, emailSubBytes); lookIdx != -1 {
		emailStart := lookIdx + len(emailSubBytes)
		return string(record[emailStart:]), true
	}
	return "", false
}

func RecordIsApiCallOrEmpty(record []byte) bool {
	return len(record) == 0 || bytes.Contains(record, apiSubBytes)
}

func FoundInRecordByFilter(record, fliter []byte) bool {
	fliter = bytes.TrimSpace(fliter)
	if len(fliter) == 0 {
		return true
	}
	return bytes.Contains(record, fliter)
}

func FoundInRecordByFilterString(record []byte, fliter string) bool {
	return FoundInRecordByFilter(record, []byte(fliter))
}

func RecordContains(line []byte, suffixes []string) bool {
	for _, sfx := range suffixes {
		if bytes.Contains(line, []byte(sfx)) {
			return true
		}
	}
	return false
}
