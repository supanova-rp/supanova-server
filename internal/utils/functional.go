package utils

func Map[S, T any](items []S, f func(S) T) []T {
	mapped := make([]T, len(items))
	for i, item := range items {
		mapped[i] = f(item)
	}

	return mapped
}

// MapToWithError transforms a slice given the mapping function 'f'.
// It will return the first error when the result from 'f' returns an error.
// Otherwise returns a new slice with the result of the function calls.
func MapToWithError[S, T any](items []S, f func(S) (T, error)) ([]T, error) {
	mapped := make([]T, len(items))

	for i, item := range items {
		result, err := f(item)
		if err != nil {
			return nil, err
		}

		mapped[i] = result
	}

	return mapped, nil
}
