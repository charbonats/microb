package utils

import "strings"

// Get a union of two maps.
// Items present both in map1 and map2 will be overwritten by map2.
func Union(map1 map[string]string, map2 map[string]string) map[string]string {
	m := make(map[string]string)

	for key, value := range map1 {
		m[key] = value
	}

	for key, value := range map2 {
		m[key] = value
	}

	return m
}

// Filter filters the given map to only include keys that start with the
// given prefix
func Filter(mapping map[string]string, prefix string) map[string]string {
	filtered := map[string]string{}

	for k, v := range mapping {
		if strings.HasPrefix(k, prefix) {
			filtered[strings.TrimPrefix(k, prefix)] = v
		}
	}

	return filtered
}

// Unique returns a new slice containing only the unique elements of the given slice
func Unique(slice []string) []string {
	keys := make(map[string]struct{})
	for _, entry := range slice {
		keys[entry] = struct{}{}
	}
	uniqueKeys := make([]string, 0, len(keys))
	for key := range keys {
		uniqueKeys = append(uniqueKeys, key)
	}
	return uniqueKeys
}
