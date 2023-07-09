package utils

// DivideSlice divides a slice into n parts
func DivideSlice[K any](slice []K, n int) [][]K {
	if n <= 0 {
		return nil
	}

	length := len(slice)
	partSize := (length + n - 1) / n // Calculate the size of each part, rounding up

	result := make([][]K, n)
	for i := 0; i < n; i++ {
		start := i * partSize
		end := (i + 1) * partSize
		if end > length {
			end = length
		}
		result[i] = slice[start:end]
	}

	return result
}
