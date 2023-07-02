package utils

import "math"

func AlmostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0000000000001
}
