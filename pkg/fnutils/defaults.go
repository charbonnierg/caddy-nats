package fnutils

func DefaultIfNil[T any](value *T, defaultValue *T) *T {
	if value == nil {
		return defaultValue
	}
	return value
}

func DefaultIfEmpty[T any](value []T, defaultValue []T) []T {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func DefaultIfEmptyString(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
