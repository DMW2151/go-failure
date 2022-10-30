package failure

import (
	"math"
	"math/rand"
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

// Test_Detector_PhiCalculation
func Test_Detector_PhiCalculation(t *testing.T) {

	var (
		lastArrivalTime         = time.Now()
		standardPhiDetector     = setupPhiAccDetector(testHeartbeatArrivalMean, testHeartbeatArrivalStdev, testWindowSize, testRandSeed)
		noisyNetworkPhiDetector = setupPhiAccDetector(testHeartbeatArrivalMean, testHeartbeatArrivalStdev*2, testWindowSize, testRandSeed)
	)

	var phiCalcTestScenarios = []struct {
		name      string
		detectors []Detector
		offset    int
		expected  float64
	}{
		{
			name:      "a-little-early",
			detectors: []Detector{standardPhiDetector},
			offset:    int(testHeartbeatArrivalMean - 1*testHeartbeatArrivalStdev),
			expected:  0.0783268,
		},
		{
			name:      "about-expected-arrival",
			detectors: []Detector{standardPhiDetector},
			offset:    int(testHeartbeatArrivalMean + 0*testHeartbeatArrivalStdev),
			expected:  0.3017074,
		},
		{
			name:      "a-little-late",
			detectors: []Detector{standardPhiDetector},
			offset:    int(testHeartbeatArrivalMean + 1*testHeartbeatArrivalStdev),
			expected:  0.7850040,
		},
		{
			name:      "right-after-last-heartbeat",
			detectors: []Detector{standardPhiDetector, noisyNetworkPhiDetector},
			offset:    0,
			expected:  math.Pow(0.01, 2), // expect this to be very near 0.0
		},
		{
			name:      "negative-sign-on-duration",
			detectors: []Detector{standardPhiDetector, noisyNetworkPhiDetector},
			offset:    -int(testHeartbeatArrivalMean),
			expected:  math.Pow(0.01, 2), // expect this to be very near 0.0
		},
		{
			// note :: can get about +/- 8 STDEV before get unstable results and cave to +INF
			name:      "very-high-no-inf",
			detectors: []Detector{standardPhiDetector},
			offset:    int(testHeartbeatArrivalMean + 6*testHeartbeatArrivalStdev),
			expected:  8.62966,
		},
		{
			// note: here is where we want to see the collapse, see: `very-high-no-inf`
			name:      "very-high-expect-inf",
			detectors: []Detector{standardPhiDetector, noisyNetworkPhiDetector},
			offset:    int(testHeartbeatArrivalMean + 255*testHeartbeatArrivalStdev),
			expected:  math.Inf(1),
		},
	}

	for _, sc := range phiCalcTestScenarios {
		t.Run(sc.name, func(t *testing.T) {

			assert := assert.New(t)
			var currentTime time.Time = lastArrivalTime.Add(time.Microsecond * time.Duration(sc.offset))

			// calculate phi ...
			for _, detector := range sc.detectors {
				phi := detector.Suspicion(lastArrivalTime, currentTime)
				assert.InDelta(sc.expected, phi, testDeltaTolerance)
			}

		})
	}
}
