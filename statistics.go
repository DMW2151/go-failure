package failure

import (
	"math"
	"time"
)

// IntervalStatistics - collection of stats for calculating phi (phi-accrual only, others tbd)
type IntervalStatistics struct {
	rSumSquares   float64
	rSum          float64
	nTotalSamples int
	windowSize    int
}

// Phi - calculate phi (suspicion level ) from IntervalStatistics
func (s *IntervalStatistics) Phi(lastT time.Time, currentT time.Time) float64 {

	var (
		nSamp     float64 = math.Min(float64(s.windowSize), float64(s.nTotalSamples))
		rAvg      float64 = (s.rSum / nSamp)
		rVar      float64 = (s.rSumSquares / nSamp) - math.Pow(rAvg, 2)
		timeDelta float64 = float64(currentT.Sub(lastT) / time.Microsecond)
	)

	// use the def'n straight from the book -> https://en.wikipedia.org/wiki/Normal_distribution
	F := 0.5 * (1 + math.Erf((timeDelta-rAvg)/(math.Pow(rVar, 0.5)*math.Pow(2, 0.5))))
	return -math.Log10(1 - F)
}
