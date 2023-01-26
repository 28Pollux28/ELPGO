package utils

func Clamp(value, low, high int) int {
	if value > high {
		return high
	} else if value < low {
		return low
	} else {
		return value
	}
}

func Mean(numbers []float64) float64 {
	var sum float64
	for _, number := range numbers {
		sum += number
	}
	return sum / float64(len(numbers))
}

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
