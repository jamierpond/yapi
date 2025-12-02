package utils

// Filter returns a new slice containing only elements where f(t) is true.
func Filter[T any](ts []T, f func(T) bool) []T {
	var matches []T
	for _, t := range ts {
		if f(t) {
			matches = append(matches, t)
		}
	}
	return matches
}

// ContainsFunc returns true if any element satisfies f.
func ContainsFunc[T any](ts []T, f func(T) bool) bool {
	for _, t := range ts {
		if f(t) {
			return true
		}
	}
	return false
}
