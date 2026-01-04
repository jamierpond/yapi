package vars

import (
	"reflect"
)

// ExpandAll recursively expands environment variables in all string fields and map values
// of a struct using the provided resolver. This eliminates the need for manual field-by-field
// expansion when adding new config fields.
func ExpandAll(obj any, resolver Resolver) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			if expanded, err := ExpandString(field.String(), resolver); err == nil {
				field.SetString(expanded)
			}
		case reflect.Map:
			if field.Type().Elem().Kind() == reflect.String {
				// Handle map[string]string
				for _, key := range field.MapKeys() {
					val := field.MapIndex(key)
					if val.Kind() == reflect.String {
						if exp, err := ExpandString(val.String(), resolver); err == nil {
							field.SetMapIndex(key, reflect.ValueOf(exp))
						}
					}
				}
			} else if field.Type().Elem().Kind() == reflect.Interface {
				// Handle map[string]any - recursively expand values
				if !field.IsNil() {
					expandedMap := expandMapAny(field.Interface(), resolver)
					field.Set(reflect.ValueOf(expandedMap))
				}
			}
		case reflect.Struct:
			ExpandAll(field.Addr().Interface(), resolver)
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				ExpandAll(field.Interface(), resolver)
			}
		}
	}
}

// expandMapAny recursively expands variables in a map[string]any
func expandMapAny(m any, resolver Resolver) map[string]any {
	mapVal, ok := m.(map[string]any)
	if !ok {
		return nil
	}

	result := make(map[string]any, len(mapVal))
	for k, v := range mapVal {
		result[k] = expandValue(v, resolver)
	}
	return result
}

// expandValue recursively expands variables in any value type
func expandValue(v any, resolver Resolver) any {
	switch val := v.(type) {
	case string:
		// Expand variables in string values
		if expanded, err := ExpandString(val, resolver); err == nil {
			return expanded
		}
		return val
	case map[string]any:
		// Recursively expand nested maps
		return expandMapAny(val, resolver)
	case []any:
		// Recursively expand array elements
		result := make([]any, len(val))
		for i, elem := range val {
			result[i] = expandValue(elem, resolver)
		}
		return result
	default:
		// Return other types unchanged (numbers, bools, nil, etc.)
		return val
	}
}
