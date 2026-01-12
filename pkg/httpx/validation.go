package httpx

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidateStruct validates struct fields based on binding tags
// This is a simple validator that checks for required fields
func ValidateStruct(s any) map[string]any {
	errors := make(map[string]any)

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get binding tag
		bindingTag := field.Tag.Get("binding")
		if bindingTag == "" {
			continue
		}

		// Parse binding rules
		rules := strings.Split(bindingTag, ",")

		for _, rule := range rules {
			rule = strings.TrimSpace(rule)

			// Check required
			if rule == "required" {
				if isZeroValue(value) {
					jsonTag := field.Tag.Get("json")
					fieldName := strings.Split(jsonTag, ",")[0]
					if fieldName == "" {
						fieldName = field.Name
					}
					errors[fieldName] = fmt.Sprintf("El campo %s es requerido", fieldName)
				}
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

// isZeroValue checks if a value is the zero value for its type
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	}
	return false
}
