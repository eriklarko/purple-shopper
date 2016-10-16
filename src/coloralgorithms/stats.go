package coloralgorithms

/**
	Shamelessly stolen from github.com/GaryBoone/GoStats
 */

func PopulationVariance(data []float64) float64 {
	n := float64(len(data))
	ssd := sumSquaredDeltas(data)
	return ssd / n
}

func sumSquaredDeltas(data []float64) (ssd float64) {
	mean := Mean(data)
	for _, v := range data {
		delta := v - mean
		ssd += delta * delta
	}
	return
}

func Mean(data []float64) float64 {
	return Sum(data) / float64(len(data))
}

func Sum(data []float64) (sum float64) {
	for _, v := range data {
		sum += v
	}
	return
}

func LaMaximum(data []Quality, leastAmountOfPoints int) (int, float64) {
	maxIndex := -1
	max := -1.0

	for i, quality := range data {
		if quality.dataPoints < leastAmountOfPoints {
			continue
		}

		if quality.rootSquaredError > max {
			max = quality.rootSquaredError
			maxIndex = i
		}
	}

	return maxIndex, max
}
