package utils

func Map[T, S any](f func(S) T, ts []S) []T {
	mapped := make([]T, len(ts))
	for i, v := range ts {
		mapped[i] = f(v)
	}

	return mapped
}

// MapToWithError transforms a slice given the mapping function 'f'.
// It will return the first error when the result from 'f' returns an error.
// Otherwise returns a new slice with the result of the function calls.
func MapToWithError[T, S any](f func(S) (T, error), ts []S) ([]T, error) {
	mapped := make([]T, len(ts))

	for i, v := range ts {
		result, err := f(v)
		if err != nil {
			return nil, err
		}

		mapped[i] = result
	}

	return mapped, nil
}
