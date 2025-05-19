package utils

import (
	"math/rand"
)

// randInt возвращает случайное число в диапазоне [min, max)
func RandInt(min, max int) int {
	return min + rand.Intn(max-min)
}
