package utils

func TryMap[T, U any](slice []T, f func(T) (U, error)) ([]U, error) {
	if len(slice) == 0 {
		return nil, nil
	}
	result := make([]U, len(slice))
	for i, value := range slice {
		mapped, err := f(value)
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}
