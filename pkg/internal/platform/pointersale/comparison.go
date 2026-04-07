package pointersale

func ToPointer[T any](a T) *T {
	return &a
}

func ComparePointers[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
