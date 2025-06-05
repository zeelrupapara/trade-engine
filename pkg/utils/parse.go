package utils

import "strconv"

func ParseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
