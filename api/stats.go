package api

import (
	"math"

	log "github.com/Sirupsen/logrus"
)

type Stats struct {
	Mean                float64 `json:"mean"`
	DSquared            float64 `json:"dsquared"`
	Length              float64 `json:"length"`
	MaxAddedTimestamp   int64   `json:"max_added_timestamp"`
	MaxRemovedTimestamp int64   `json:"max_removed_timestamp"`
	NumAdds             int64   `json:"num_adds"`
	NumRemoves          int64   `json:"num_removes"`
	NumEmptyRemoves     int64   `json:"num_empty_removes"`
}

func NewStats() *Stats {
	return &Stats{}
}

func (s *Stats) Add(value float64, timestamp int64) {
	if math.IsNaN(value) || math.IsInf(value, 1) || math.IsInf(value, -1) {
		log.Infof("Add %+v BAD VALUE", value)
		return
	}
	s.NumAdds++
	if timestamp > s.MaxAddedTimestamp {
		s.MaxAddedTimestamp = timestamp
	} else {
		log.Warnf("Expecting to add only new values, old timestamp: %d found, max %d.", timestamp, s.MaxAddedTimestamp)
	}
	s.Length++

	meanIncrement := (value - s.Mean) / s.Length
	newMean := s.Mean + meanIncrement

	dSquaredIncrement := (value - newMean) * (value - s.Mean)
	newDSquared := (s.DSquared*(s.Length-1) + dSquaredIncrement) / s.Length
	if newDSquared < 0 {
		// Correcting float inaccuracy.
		if newDSquared < -0.00001 {
			log.Warnf("Add: newDSquared negative: %+v. Setting to 0. %f, %d %+v", newDSquared, value, timestamp, s)
		}
		newDSquared = 0
	}

	s.Mean = newMean
	s.DSquared = newDSquared
}

func (s *Stats) Remove(value float64, timestamp int64) {
	if math.IsNaN(value) || math.IsInf(value, 1) || math.IsInf(value, -1) {
		log.Infof("Remove %+v BAD VALUE", value)
		return
	}
	if timestamp > s.MaxRemovedTimestamp {
		s.MaxRemovedTimestamp = timestamp
	} else {
		log.Warnf("Expecting to remove only new values, old timestamp: %d found, max %d.", timestamp, s.MaxRemovedTimestamp)
	}
	if s.Length <= 1 {
		if s.Length == 1 {
			s.NumRemoves++
		} else {
			s.NumEmptyRemoves++
		}
		log.Warnf("Empty stats (%f, %d, %+v).", value, timestamp, s)
		s.Mean = 0
		s.DSquared = 0
		s.Length = 0
		return
	}
	s.NumRemoves++
	s.Length--

	meanIncrement := (s.Mean - value) / s.Length
	newMean := s.Mean + meanIncrement

	dSquaredIncrement := ((newMean - value) * (value - s.Mean))
	newDSquared := (s.DSquared*(s.Length+1) + dSquaredIncrement) / s.Length
	if newDSquared < 0 {
		// Correcting float inaccuracy.
		if newDSquared < -0.00001 {
			log.Warnf("Remove: newDSquared negative: %+v. Setting to 0. %f, %d %+v", newDSquared, value, timestamp, s)
		}
		newDSquared = 0
	}

	s.Mean = newMean
	s.DSquared = newDSquared
}
