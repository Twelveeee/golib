package utils

func MapByKey[T any, K comparable](base []T, keyFunc func(T) K) map[K]T {
	result := make(map[K]T)
	for _, v := range base {
		result[keyFunc(v)] = v
	}
	return result
}

func MapColumn[T any, U any](slice []T, extractor func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = extractor(v)
	}
	return result
}

func ArrayKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func ArrayValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
