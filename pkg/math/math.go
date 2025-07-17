package math

import "math"

func RoundDown(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Floor(value*factor) / factor
}

func HasMoreThanTwoDecimalPlaces(value float64) bool {
	scaled := value * 100
	return scaled != float64(int64(scaled))
}
