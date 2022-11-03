package failure

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testWindowSize              int           = 1_000
	testRandSeed                int64         = 2151
	testHeartbeatArrivalMean    float64       = 100.0
	testHeartbeatArrivalStdev   float64       = 10.0
	testDeltaTolerance          float64       = 0.0001
	testUnits                   time.Duration = time.Millisecond
	testReapInterval            time.Duration = time.Second * 10
	testReapIntervalGracePeriod time.Duration = time.Second * 10
)

// setupPhiAccDetector
func setupPhiAccDetector(arrivalTime time.Time, mean, stdev float64, windowSize int, seed int64) *PhiAccrualDetector {

	var (
		nOpts = NodeOptions{
			EstimationWindowSize: testWindowSize,
			ReapInterval:         testReapInterval,
			PurgeGracePeriod:     testReapIntervalGracePeriod,
		}

		phiD            *PhiAccrualDetector = NewPhiAccrualDetector(arrivalTime, &nOpts, &NodeMetadata{})
		lastArrivalTime time.Time
	)

	rand.Seed(seed)

	// add forward windowSize intervals
	ctx := context.Background()

	for i := 0; i < windowSize; i++ {
		lastArrivalTime = phiD.lastHeartbeat
		arrivalTime = lastArrivalTime.Add(
			time.Duration(rand.NormFloat64()*stdev+mean) * testUnits,
		)

		phiD.AddValue(ctx, arrivalTime)
	}
	return phiD
}

// Test_Detector_PhiCalculation
func Test_Detector_PhiCalculation(t *testing.T) {

	var (
		lastArrivalTime = time.Now()

		standardPhiDetector = setupPhiAccDetector(
			lastArrivalTime,
			testHeartbeatArrivalMean,
			testHeartbeatArrivalStdev,
			testWindowSize,
			testRandSeed,
		)
	)

	var phiCalcTestScenarios = []struct {
		name     string
		detector *PhiAccrualDetector
		offset   int
		expected float64
	}{
		{
			name:     "about-expected-arrival",
			detector: standardPhiDetector,
			offset:   int(testHeartbeatArrivalMean),
			expected: 0.319290,
		},
		{
			name:     "a-little-early",
			detector: standardPhiDetector,
			offset:   int(testHeartbeatArrivalMean - 1*testHeartbeatArrivalStdev),
			expected: 0.085151,
		},
		{
			name:     "a-little-late",
			detector: standardPhiDetector,
			offset:   int(testHeartbeatArrivalMean + 1*testHeartbeatArrivalStdev),
			expected: 0.816980,
		},
		{
			name:     "right-after-last-heartbeat",
			detector: standardPhiDetector,
			offset:   0,
			expected: math.Pow(0.01, 3), // expect this to be very near 0.0
		},
		{
			name:     "negative-sign-on-duration",
			detector: standardPhiDetector,
			offset:   -int(testHeartbeatArrivalMean),
			expected: math.Pow(0.01, 3), // expect this to be very near 0.0
		},
		{
			// note :: can get about +/- 8 STDEV before get unstable results and cave to +INF
			name:     "very-high-no-inf",
			detector: standardPhiDetector,
			offset:   int(testHeartbeatArrivalMean + 6*testHeartbeatArrivalStdev),
			expected: 8.736880,
		},
		{
			// note: here is where we want to see the collapse, see: `very-high-no-inf`
			name:     "very-high-expect-inf",
			detector: standardPhiDetector,
			offset:   int(testHeartbeatArrivalMean + 255*testHeartbeatArrivalStdev),
			expected: math.Inf(1),
		},
	}

	for _, sc := range phiCalcTestScenarios {
		t.Run(sc.name, func(t *testing.T) {

			assert := assert.New(t)

			// calculate phi ...
			var testTimestamp time.Time = sc.detector.lastHeartbeat.Add(
				testUnits * time.Duration(sc.offset),
			)

			phi := sc.detector.Suspicion(testTimestamp)
			assert.InDelta(sc.expected, phi, testDeltaTolerance)
		})
	}
}
