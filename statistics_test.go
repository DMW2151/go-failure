package phiacc

import (
	"log"
	"math"
	"testing"
	"time"
)

// Test_IntervalStatistics_Phi -
func Test_IntervalStatistics_Phi(t *testing.T) {

	var lastArrivalTime time.Time = time.Now()

	var phiCalcTestScenarios = []struct {
		name        string
		currentTime time.Time
	}{
		{
			name:        "a-little-early",
			currentTime: lastArrivalTime.Add(time.Millisecond * 100),
		},
	}

	for _, sc := range phiCalcTestScenarios {
		t.Run(sc.name, func(t *testing.T) {

			s := IntervalStatistics{
				rSumSquares:   math.Pow(10, 7) * 1.04,
				rSum:          100_000,
				nTotalSamples: 1_000,
				windowSize:    1_000,
			}

			phi := s.Phi(lastArrivalTime, sc.currentTime)
			log.Println(phi)
		})
	}
}
