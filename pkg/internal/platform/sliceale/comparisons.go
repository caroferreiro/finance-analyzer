package sliceale

// IsOneOf returns true iff the element is in the slice.
func IsOneOf[T comparable](s []T, elem T) bool {
	for _, e := range s {
		if e == elem {
			return true
		}
	}
	return false
}
