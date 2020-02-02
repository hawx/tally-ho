package mfutil

import "strings"

func Get(value interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if v, ok := get(value, key); ok {
			return v
		}
	}

	return nil
}

func get(value interface{}, key string) (interface{}, bool) {
	// if an array get the first value
	if typed, ok := value.([]interface{}); ok {
		if len(typed) > 0 {
			return get(typed[0], key)
		}

		return nil, false
	}

	// if no key then this must be what we were looking for
	if key == "" {
		return value, true
	}

	parts := strings.SplitN(key, ".", 2)

	if typed, ok := value.(map[string][]interface{}); ok {
		next, ok := typed[parts[0]]

		if !ok || len(next) == 0 {
			return nil, false
		}

		if len(parts) == 2 {
			return get(next[0], parts[1])
		}

		return get(next[0], "")
	}

	if typed, ok := value.(map[string]interface{}); ok {
		next, ok := typed[parts[0]]

		if !ok {
			return nil, ok
		}

		if len(parts) == 2 {
			return get(next, parts[1])
		}

		return get(next, "")
	}

	return nil, false
}
