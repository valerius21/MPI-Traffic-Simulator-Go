package utils

func MapFun[T any](arr []T, fn func(v T) T) []T {
	var result []T
	for _, v := range arr {
		result = append(result, fn(v))
	}
	return result
}
