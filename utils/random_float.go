package utils

import "math/rand"

// RandomFloat64 returns a random float64 between min and max
func RandomFloat64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
