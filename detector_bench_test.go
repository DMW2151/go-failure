package phiacc

import (
	"math/rand"
	"testing"
)

// Benchmark_PhiAccDetector_Add
// note :: rough approx. -> 40ns/ operation until window full; then 60ns / opearation after
func Benchmark_PhiAccDetector_Add(b *testing.B) {

	var collectorAddBenchScenarios = []struct {
		windowSize    int
		nTotalSamples int
		name          string
	}{
		{
			name:          "small-window",
			windowSize:    10,
			nTotalSamples: 100_000,
		},
		{
			name:          "medium-window",
			windowSize:    10_000,
			nTotalSamples: 100_000,
		},
		{
			name:          "large-window",
			windowSize:    100_000,
			nTotalSamples: 100_000,
		},
	}

	for _, sc := range collectorAddBenchScenarios {
		b.Run(sc.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {

				//
				b.StopTimer()
				var phiD *PhiAccrualDetector = NewPhiAccrualDetector(sc.windowSize)
				b.StartTimer()

				//
				for i := 0; i < sc.nTotalSamples; i++ {
					phiD.AddValue(rand.Float64() * 100)
				}
			}
		})
	}
}
