package math

import (
	"math"
	"strconv"
	"strings"
)

func RoundDown(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Floor(value*factor) / factor
}

func HasMoreThanTwoDecimalPlaces(value float64) bool {
	s := strconv.FormatFloat(value, 'f', -1, 64)
	parts := strings.Split(s, ".")
	return len(parts) == 2 && len(parts[1]) > 2
}
