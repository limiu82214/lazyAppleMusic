package util

import "strings"

func ProgressBarUi(percent int, length int) string {
	const (
		progressFullLeft   = ""
		progressFullMid    = ""
		progressFullRight  = ""
		progressEmptyLeft  = ""
		progressEmptyMid   = ""
		progressEmptyRight = ""
	)

	if length < 2 {
		return ""
	}
	filled := percent * length / 100
	if filled > length {
		filled = length
	}
	if filled < 0 {
		filled = 0
	}
	// head
	head := progressEmptyLeft
	if filled > 0 {
		head = progressFullLeft
	}
	// tail
	tail := progressEmptyRight
	if filled == length {
		tail = progressFullRight
	}
	// middle (扣掉頭尾)
	fullMidCount := filled - 1
	if fullMidCount < 0 {
		fullMidCount = 0
	}
	emptyMidCount := length - 2 - fullMidCount
	if emptyMidCount < 0 {
		emptyMidCount = 0
	}
	mid := strings.Repeat(progressFullMid, fullMidCount) +
		strings.Repeat(progressEmptyMid, emptyMidCount)

	return head + mid + tail
}
