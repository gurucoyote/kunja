package api

import "sync/atomic"

var lastTotal int64

// setLastTotal stores the total number of items returned by the last API
// request that provided the X-Total header.
func setLastTotal(n int) {
	atomic.StoreInt64(&lastTotal, int64(n))
}

// GetLastTotal returns the last total value recorded from an API response.
// It is safe for concurrent reads.
func GetLastTotal() int {
	return int(atomic.LoadInt64(&lastTotal))
}
