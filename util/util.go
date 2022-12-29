package util

// fills a slice with a value (returns the same slice)
func Fill[T any](sl []T, val T) []T {
	for i := range sl {
		sl[i] = val
	}

	return sl
}
