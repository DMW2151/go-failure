package failure

import (
	"math"
	"sync"
	"time"
)

// Detector - Any object that resolves (timestamp, timestamp) -> suspicionlevel
//
// note/warn :: not useful as an interface right now, only have the phiAccrualDetector, but could
// easily add in binaryFailureDetector, chenFailureDetector, or bertierFailureDetector
type Detector interface {
	Suspicion(time.Time, time.Time) float64
}

// PhiAccrualDetector - detector for each process tracked by the server
type PhiAccrualDetector struct {
	stats          *IntervalStatistics
	window         []windowElement
	expiringSample *windowElement
	lastHeartbeat  time.Time
	mu             sync.Mutex
	Tags           map[string]string
}

// windowElement - node in ring
type windowElement struct {
	value float64
	next  *windowElement
}

// NewPhiAccrualDetector -
func NewPhiAccrualDetector(windowSize int) *PhiAccrualDetector {

	// init phi-acc detector - create a very simple ring to iterate through
	window := make([]windowElement, windowSize)
	for i := 0; i < windowSize-1; i++ {
		window[i].next = &window[i+1]
	}
	window[windowSize-1].next = &window[0]

	return &PhiAccrualDetector{
		stats: &IntervalStatistics{
			rSum:          0.0,
			rSumSquares:   0.0,
			nTotalSamples: 0,
			windowSize:    windowSize,
		},
		window:         window,
		expiringSample: &window[0],
	}
}

// AddValue - tack on a value to the statistics
func (phiD *PhiAccrualDetector) AddValue(newVal float64) error {

	// safety first; note :: idk if we run into locking issues, only one goroutine should  
	// ever be toouching these values, but good to be safe...
	phiD.mu.Lock()
	defer phiD.mu.Unlock()

	var expVal float64 = phiD.expiringSample.value

	// update value at the current ptr
	phiD.expiringSample.value = newVal
	phiD.expiringSample = phiD.expiringSample.next

	// upate the nTotalSamples, rSum, and rSumSquares for current collection
	phiD.stats.rSum += (newVal - expVal)
	phiD.stats.rSumSquares += (math.Pow(newVal, 2) - math.Pow(expVal, 2))
	phiD.stats.nTotalSamples++
	return nil
}

// Parameters - kinda useless, calculates the intermediates for phiD.Suspicion or IntervalStatistics.Phi
func (phiD *PhiAccrualDetector) Parameters() (float64, float64) {
	var (
		nSamp float64 = math.Min(float64(phiD.stats.windowSize), float64(phiD.stats.nTotalSamples))
		rAvg  float64 = (phiD.stats.rSum / nSamp)
		rVar  float64 = (phiD.stats.rSumSquares / nSamp) - math.Pow(rAvg, 2)
	)
	return rAvg, math.Pow(rVar, 0.5)
}

// Suspicion
func (phiD PhiAccrualDetector) Suspicion(mtime time.Time, ctime time.Time) float64 {
	return phiD.stats.Phi(mtime, ctime)
}
