package mfutil

import (
	"strings"
)

func Get(value any, keys ...string) any {
	for _, key := range keys {
		if v, ok := SafeGet(value, key); ok {
			return v
		}
	}

	return nil
}

func GetAll(value any, keys ...string) []any {
	for _, key := range keys {
		if v, ok := SafeGetAll(value, key); ok {
			return v
		}
	}

	return nil
}

func Has(value any, key string) bool {
	_, ok := SafeGet(value, key)

	return ok
}

func SafeGet(value any, key string) (any, bool) {
	// if an array get the first value
	if typed, ok := value.([]any); ok {
		if len(typed) > 0 {
			return SafeGet(typed[0], key)
		}

		return nil, false
	}

	// if no key then this must be what we were looking for
	if key == "" {
		return value, true
	}

	parts := strings.SplitN(key, ".", 2)

	if typed, ok := value.(map[string][]any); ok {
		next, ok := typed[parts[0]]

		if !ok || len(next) == 0 {
			return nil, false
		}

		if len(parts) == 2 {
			return SafeGet(next[0], parts[1])
		}

		return SafeGet(next[0], "")
	}

	if typed, ok := value.(map[string]any); ok {
		next, ok := typed[parts[0]]

		if !ok {
			return nil, ok
		}

		if len(parts) == 2 {
			return SafeGet(next, parts[1])
		}

		return SafeGet(next, "")
	}

	return nil, false
}

func SafeGetAll(value any, key string) ([]any, bool) {
	if typed, ok := value.([]any); ok {
		if key == "" {
			return typed, true
		}

		return SafeGetAll(typed[0], key)
	}

	if key == "" {
		return nil, false
	}

	parts := strings.SplitN(key, ".", 2)

	if typed, ok := value.(map[string][]any); ok {
		next, ok := typed[parts[0]]

		if !ok || len(next) == 0 {
			return nil, false
		}

		if len(parts) == 2 {
			return SafeGetAll(next, parts[1])
		}

		return SafeGetAll(next, "")
	}

	if typed, ok := value.(map[string]any); ok {
		next, ok := typed[parts[0]]

		if !ok {
			return nil, ok
		}

		if len(parts) == 2 {
			return SafeGetAll(next, parts[1])
		}

		return SafeGetAll(next, "")
	}

	return nil, false
}
