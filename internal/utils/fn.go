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

// Coalesce returns the first non-zero value.
func Coalesce[T comparable](vals ...T) T {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}
	return zero
}

// MergeMaps merges src into dst. Keys in src overwrite dst. Returns new map.
func MergeMaps[K comparable, V any](dst, src map[K]V) map[K]V {
	out := make(map[K]V, len(dst)+len(src))
	for k, v := range dst {
		out[k] = v
	}
	for k, v := range src {
		out[k] = v
	}
	return out
}

// DeepCloneMap creates a deep copy of a map[string]interface{}.
func DeepCloneMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[k] = DeepCloneMap(val)
		case []interface{}:
			dst[k] = DeepCloneSlice(val)
		default:
			dst[k] = v
		}
	}
	return dst
}

// DeepCloneSlice creates a deep copy of a slice of interfaces.
func DeepCloneSlice(src []interface{}) []interface{} {
	if src == nil {
		return nil
	}
	dst := make([]interface{}, len(src))
	for i, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[i] = DeepCloneMap(val)
		case []interface{}:
			dst[i] = DeepCloneSlice(val)
		default:
			dst[i] = v
		}
	}
	return dst
}
