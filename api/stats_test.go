package api

import (
	log "github.com/sirupsen/logrus"
)

func Mean(numbers []float64) float64 {
	sum := float64(0)
	for i := range numbers {
		sum += numbers[i]
	}
	return sum / float64(len(numbers))
}

func DSquared(numbers []float64) float64 {
	mean := Mean(numbers)
	sum := float64(0)
	for i := range numbers {
		sum += (numbers[i] - mean) * (numbers[i] - mean)
	}
	return sum / float64(len(numbers))
}

func (suite *APISuite) TestStats() {
	r := suite.Require()

	numbers := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	r.Equal(5.5, Mean(numbers))
	r.Equal(8.25, DSquared(numbers))

	s := NewStats()
	i := 1
	for i <= len(numbers) {
		s.Add(numbers[i-1], 0)
		subNumbers := numbers[0:i]
		log.Infof("Numbers %+v. Mean: %f %f, DSquared: %f %f.",
			subNumbers, Mean(subNumbers), s.Mean, DSquared(subNumbers), s.DSquared)
		r.Equal(Mean(subNumbers), s.Mean)
		r.Equal(DSquared(subNumbers), s.DSquared)
		i++
	}

	i = 0
	for i < len(numbers)-1 {
		s.Remove(numbers[i], 0)
		subNumbers := numbers[i+1:]
		log.Infof("Numbers %+v. Mean: %f %f, DSquared: %f %f.",
			subNumbers, Mean(subNumbers), s.Mean, DSquared(subNumbers), s.DSquared)
		r.Equal(Mean(subNumbers), s.Mean)
		r.Equal(DSquared(subNumbers), s.DSquared)
		i++
	}

	// Empty list should actually be NaN, but due to serialization we
	// will work with 0.0
	s.Remove(numbers[len(numbers)-1], 0)
	r.Equal(0.0, s.Mean)
	r.Equal(0.0, s.DSquared)
}
