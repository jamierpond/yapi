package vars

import "reflect"

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
				for _, key := range field.MapKeys() {
					val := field.MapIndex(key)
					if val.Kind() == reflect.String {
						if exp, err := ExpandString(val.String(), resolver); err == nil {
							field.SetMapIndex(key, reflect.ValueOf(exp))
						}
					}
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
