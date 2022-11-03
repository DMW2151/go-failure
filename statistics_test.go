package failure

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

// Test_IntervalStatistics_Phi - test rolling stats -> phi
func Test_IntervalStatistics_Phi(t *testing.T) {

	var testDeltaTolerance float64 = 0.0001
	var lastArrivalTime time.Time = time.Now()

	var phiCalcTestScenarios = []struct {
		name          string
		currentTime   time.Time
		rSum          float64
		rSumSquares   float64
		windowSize    int
		nTotalSamples int
		soln          float64
	}{
		{
			name:          "at-exactly-expectation",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 100),
			rSum:          100_000,
			rSumSquares:   math.Pow(10, 7) * 1.04,
			nTotalSamples: 1_000,
			windowSize:    1_000,
			soln:          0.3010299956639812,
		},
		{
			name:          "only-one-sample-so-far",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 100),
			rSum:          100,
			rSumSquares:   math.Pow(100, 2),
			nTotalSamples: 1,
			windowSize:    1_000,
			soln:          math.NaN(),
		},
		{
			name:          "only-two-samples-so-far",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 100),
			rSum:          98.8 + 100.1,
			rSumSquares:   math.Pow(98.8, 2) + math.Pow(100.1, 2),
			nTotalSamples: 2,
			windowSize:    1_000,
			soln:          0.7017290002925862,
		},
		{
			name:          "no-samples-so-far",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 100),
			rSum:          0,
			rSumSquares:   0,
			nTotalSamples: 0,
			windowSize:    1_000,
			soln:          math.NaN(),
		},
		{
			name:          "at-exactly-expectation-huge-variance",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 100),
			rSum:          100_000,
			rSumSquares:   math.Pow(10, 7) * 2.00,
			nTotalSamples: 1_000,
			windowSize:    1_000,
			soln:          0.3010299956639812,
		},
		{
			name:          "after-expectation",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 105),
			rSum:          100_000,
			rSumSquares:   math.Pow(10, 7) * 1.04,
			nTotalSamples: 1_000,
			windowSize:    1_000,
			soln:          0.3965376860943206,
		},
		{
			name:          "after-expectation-huge-variance",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 100),
			rSum:          100_000,
			rSumSquares:   math.Pow(10, 7) * 2.00,
			nTotalSamples: 1_000,
			windowSize:    1_000,
			soln:          0.3010299956639812,
		},
		{
			name:          "after-expectation-low-variance",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 105),
			rSum:          100_000,
			rSumSquares:   math.Pow(10, 7) * 1.0001,
			nTotalSamples: 1_000,
			windowSize:    1_000,
			soln:          6.54264567249168,
		},
		{
			name:          "no-variance",
			currentTime:   lastArrivalTime.Add(time.Millisecond * 105),
			rSum:          100_000,
			rSumSquares:   math.Pow(10, 7),
			nTotalSamples: 1_000,
			windowSize:    1_000,
			soln:          math.Inf(1),
		},
	}

	for _, sc := range phiCalcTestScenarios {
		t.Run(sc.name, func(t *testing.T) {

			s := IntervalStatistics{
				rSumSquares:   sc.rSumSquares,
				rSum:          sc.rSum,
				nTotalSamples: sc.nTotalSamples,
				windowSize:    sc.windowSize,
			}

			phi := s.Phi(lastArrivalTime, sc.currentTime)

			if !math.IsNaN(sc.soln) {
				assert.InDelta(t, sc.soln, phi, testDeltaTolerance)
				return
			}
			assert.True(t, math.IsNaN(phi))
		})
	}
}
