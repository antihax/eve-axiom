package dogma

func avg(vals []float64) float64 {
	var avg float64
	for _, v := range vals {
		avg += v
	}
	return avg / float64(len(vals))
}

func min(vals []float64) float64 {
	min := vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
	}
	return min
}

func max(vals []float64) float64 {
	max := vals[0]
	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	return max
}
