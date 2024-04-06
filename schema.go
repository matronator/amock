package main

import "strings"

type ValidationResult struct {
	Valid  bool
	Errors []string
}

func ValidateField(field *Field, value any, key string, table *Table) *ValidationResult {
	if value == nil {
		if !field.Nullable {
			return &ValidationResult{false, []string{"Field is not nullable: " + key}}
		}
		return &ValidationResult{true, nil}
	}

	if field.Type == "enum" {
		params := strings.Split(strings.TrimPrefix(field.Params, ":"), ",")
		for _, param := range params {
			if value == param {
				return &ValidationResult{true, nil}
			}
		}
		return &ValidationResult{false, []string{"Value doesn't match any of the enum values for field: " + key}}
	}

	if field.Type == "id" && field.Subtype == "uuid" {
		if len(value.(string)) == 36 {
			return &ValidationResult{true, nil}
		}
		return &ValidationResult{false, []string{"Invalid UUID format for field: " + key}}
	} else if field.Type == "id" && field.Subtype != "uuid" {
		entities, err := ReadTable(table)
		if err != nil {
			return &ValidationResult{false, []string{err.Error()}}
		}
		idExists := false
		for _, entity := range entities {
			if entity[key] == value {
				idExists = true
			}
		}
		if !idExists {
			return &ValidationResult{true, nil}
		} else {
			return &ValidationResult{false, []string{"Duplicate ID for field: " + key}}
		}
	}

	switch value.(type) {
	case bool:
		if field.Type == "bool" {
			return &ValidationResult{true, nil}
		}
		return &ValidationResult{false, []string{"Invalid value for field: " + key}}
	case string:
		if field.Type == "string" || (field.Type == "date" && field.Subtype != "timestamp") {
			return &ValidationResult{true, nil}
		}
		return &ValidationResult{false, []string{"Invalid value for field: " + key}}
	case float32:
	case float64:
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
		if field.Type == "number" {
			return &ValidationResult{true, nil}
		}

		if field.Type == "date" && field.Subtype == "timestamp" {
			return &ValidationResult{true, nil}
		}

		return &ValidationResult{false, []string{"Invalid value for field: " + key}}
	default:
		return &ValidationResult{false, []string{"Invalid value for field: " + key}}
	}

	return &ValidationResult{true, nil}
}
