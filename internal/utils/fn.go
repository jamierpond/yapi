package utils

// Map transforms a slice of T to a slice of U.
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i, t := range ts {
		us[i] = f(t)
	}
	return us
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
