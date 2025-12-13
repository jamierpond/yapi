// Package utils provides generic utility functions.
package utils

// Map transforms a slice of T to a slice of U.
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i, t := range ts {
		us[i] = f(t)
	}
	return us
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

// DeepCloneMap creates a deep copy of a map[string]any.
func DeepCloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case map[string]any:
			dst[k] = DeepCloneMap(val)
		case []any:
			dst[k] = DeepCloneSlice(val)
		default:
			dst[k] = v
		}
	}
	return dst
}

// DeepCloneSlice creates a deep copy of a slice of interfaces.
func DeepCloneSlice(src []any) []any {
	if src == nil {
		return nil
	}
	dst := make([]any, len(src))
	for i, v := range src {
		switch val := v.(type) {
		case map[string]any:
			dst[i] = DeepCloneMap(val)
		case []any:
			dst[i] = DeepCloneSlice(val)
		default:
			dst[i] = v
		}
	}
	return dst
}
