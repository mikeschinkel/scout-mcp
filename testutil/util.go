package testutil

import (
	"testing"
	"time"
)

func WithinTimeframe(t *testing.T, t1, t2 time.Time, d time.Duration) bool {
	t.Helper()
	delta := t1.Sub(t2)
	if delta < 0 {
		delta = -delta
	}
	return delta <= d
}

// ContainsMatch returns true if the item is contained in the list of items per the func f()
func ContainsMatch[T any, U any](item U, items []T, f func(U, T) bool) (contains bool) {
	for _, i := range items {
		if f(item, i) {
			contains = true
			goto end
		}
	}
end:
	return contains
}
