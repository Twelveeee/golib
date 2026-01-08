package utils

func ForEach[T any](data []T, f func(T) error) error {
	for _, item := range data {
		if err := f(item); err != nil {
			return err
		}
	}
	return nil
}

func FindIndex[T any](data []T, f func(T) bool) int {
	for idx, item := range data {
		if f(item) {
			return idx
		}
	}
	return -1
}

func FindItem[T comparable](data []T, target T) int {
	for idx, item := range data {
		if target == item {
			return idx
		}
	}
	return -1
}

func Map[T any, K any](data []T, f func(T) K) []K {
	result := make([]K, 0, len(data))
	for _, item := range data {
		result = append(result, f(item))
	}
	return result
}

func Unique[T comparable](data []T) []T {
	m := make(map[T]struct{})
	for _, item := range data {
		m[item] = struct{}{}
	}
	result := make([]T, 0, len(m))
	for item := range m {
		result = append(result, item)
	}
	return result
}

func InArray[T comparable](target T, data []T) bool {
	for _, item := range data {
		if item == target {
			return true
		}
	}
	return false
}

func Filter[T any](data []T, f func(T) bool) []T {
	result := make([]T, 0, len(data))
	for _, item := range data {
		if f(item) {
			result = append(result, item)
		}
	}
	return result
}

func Chunk[T any](data []T, size int) [][]T {
	if len(data) <= size {
		return [][]T{data}
	}
	result := make([][]T, 0, len(data)/8)
	for i := 0; i < len(data); i += size {
		if i+size <= len(data)-1 {
			result = append(result, data[i:i+size])
		} else {
			result = append(result, data[i:])
		}
	}
	return result
}

// Reverse 反转切片（原地反转）
func Reverse[T any](data []T) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
