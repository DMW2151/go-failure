package phiacc

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testWindowSize            int     = 1_000
	testRandSeed              int64   = 2151
	testHeartbeatArrivalMean  float64 = 100.0
	testHeartbeatArrivalStdev float64 = 10.0
	testDeltaTolerance        float64 = 0.0001
)

// setupPhiAccDetector
func setupPhiAccDetector(mean, stdev float64, windowSize int, seed int64) *PhiAccrualDetector {
	rand.Seed(seed)
	var phiD *PhiAccrualDetector = NewPhiAccrualDetector(windowSize)
	for i := 0; i < windowSize; i++ {
		phiD.AddValue(rand.NormFloat64()*stdev + mean)
	}
	return phiD
}

// Test_WindowCollector_PhiCalculation
func Test_WindowCollector_PhiCalculation(t *testing.T) {

	var (
		lastArrivalTime   time.Time           = time.Now()
		hbmean            float64             = testHeartbeatArrivalMean
		hbstdv            float64             = testHeartbeatArrivalStdev
		standardCollector *PhiAccrualDetector = setupPhiAccDetector(
			hbmean, hbstdv, testWindowSize, testRandSeed,
		)
		noVarianceCollector *PhiAccrualDetector = setupPhiAccDetector(
			hbmean, hbstdv*0, testWindowSize, testRandSeed,
		)
		unstableNetworkCollector *PhiAccrualDetector = setupPhiAccDetector(
			hbmean, hbstdv*2, testWindowSize, testRandSeed,
		)
		emptyCollector *PhiAccrualDetector = NewPhiAccrualDetector(testWindowSize)
	)

	var phiCalcTestScenarios = []struct {
		name      string
		detectors []Detector
		offset    int
		suspicion float64
	}{
		{
			name:      "a-little-early",
			detectors: []Detector{standardCollector},
			offset:    int(hbmean - 1*hbstdv),
			suspicion: 0.0783268,
		},
		{
			name:      "about-expected-arrival",
			detectors: []Detector{standardCollector},
			offset:    int(hbmean + 0*hbstdv),
			suspicion: 0.3017074,
		},
		{
			name:      "a-little-late",
			detectors: []Detector{standardCollector},
			offset:    int(hbmean + 1*hbstdv),
			suspicion: 0.7850040,
		},
		{
			// note :: on `unstableNetworkCollector` -> cannot have stdev too high or else suspicion is never 0
			name:      "right-after-last-heartbeat",
			detectors: []Detector{standardCollector, unstableNetworkCollector},
			offset:    0,
			suspicion: math.Pow(0.01, 2),
		},
		{
			name:      "negative-sign-on-duration",
			detectors: []Detector{standardCollector, unstableNetworkCollector},
			offset:    -int(hbmean),
			suspicion: math.Pow(0.01, 2),
		},
		{
			// note :: can get about +/- 8 STDEV before get unstable results and cave to 8
			name:      "very-high-no-inf",
			detectors: []Detector{standardCollector},
			offset:    int(hbmean + 6*hbstdv),
			suspicion: 8.62966,
		},
		{
			name:      "very-high-expect-inf",
			detectors: []Detector{standardCollector, unstableNetworkCollector},
			offset:    int(hbmean + 255*hbstdv),
			suspicion: math.Inf(1),
		},
		{
			name:      "no-variance-collector",
			detectors: []Detector{noVarianceCollector},
			offset:    int(hbmean),
			suspicion: math.NaN(),
		},
		{
			name:      "empty-collector",
			detectors: []Detector{emptyCollector},
			offset:    int(hbmean),
			suspicion: math.NaN(),
		},
	}

	for _, sc := range phiCalcTestScenarios {
		t.Run(sc.name, func(t *testing.T) {

			// load seed data ...
			assert := assert.New(t)
			var currentTime time.Time = lastArrivalTime.Add(time.Millisecond * time.Duration(sc.offset))

			// calculate phi ...
			for _, detector := range sc.detectors {
				phi := detector.Suspicion(lastArrivalTime, currentTime)
				assert.InDelta(
					sc.suspicion, phi, testDeltaTolerance,
					fmt.Sprintf(
						"detector (%s); got phi (%f) outside of expected range (%f, %f)",
						reflect.TypeOf(detector),
						phi,
						sc.suspicion-testDeltaTolerance,
						sc.suspicion+testDeltaTolerance,
					),
				)
			}

		})
	}
}
