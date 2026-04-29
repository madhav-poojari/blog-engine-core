package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringSlice is a PostgreSQL text[] compatible type for sqlx.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "{}", nil
	}
	// Format as PostgreSQL array literal: {"a","b","c"}
	arr := "{"
	for i, v := range s {
		if i > 0 {
			arr += ","
		}
		b, _ := json.Marshal(v) // quotes & escapes the string
		arr += string(b)
	}
	arr += "}"
	return arr, nil
}

func (s *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*s = StringSlice{}
		return nil
	}

	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("StringSlice.Scan: unsupported type %T", src)
	}

	// Parse PostgreSQL array literal: {val1,val2,...}
	if len(raw) < 2 || raw[0] != '{' || raw[len(raw)-1] != '}' {
		*s = StringSlice{}
		return nil
	}

	inner := raw[1 : len(raw)-1]
	if inner == "" {
		*s = StringSlice{}
		return nil
	}

	// Parse CSV, respecting double-quoted values
	var result []string
	var current []byte
	inQuotes := false
	escaped := false

	for i := 0; i < len(inner); i++ {
		ch := inner[i]
		if escaped {
			current = append(current, ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if ch == ',' && !inQuotes {
			result = append(result, string(current))
			current = current[:0]
			continue
		}
		current = append(current, ch)
	}
	result = append(result, string(current))

	*s = result
	return nil
}
