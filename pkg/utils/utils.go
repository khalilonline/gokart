package utils

func SafeDeref[T any](v *T) T {
	if v != nil {
		return *v
	}
	var zero T
	return zero
}

func SafeDerefOrDefault[T any](v *T, defaultValue T) T {
	if v != nil {
		return *v
	}
	return defaultValue
}

func AsPtrFromAny[T any](value any) *T {
	if value == nil {
		return nil
	}
	assertedValue := value.(T)
	return &assertedValue
}

func AsValueFromAny[T any](value any) (T, bool) {
	if value == nil {
		var zero T
		return zero, false
	}
	assertedValue, ok := value.(T)
	return assertedValue, ok
}

func ToStrings[S []E, E ~string](slice S) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = string(v)
	}

	return result
}
