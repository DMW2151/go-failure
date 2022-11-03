package failure

import (
	"context"
	"math"
	"sync"
	"time"
)

// PhiAccrualDetector - detector for each client pinging the server
type PhiAccrualDetector struct {
	stats          *IntervalStatistics
	metadata       *NodeMetadata
	window         []windowElement
	expiringSample *windowElement
	lastHeartbeat  time.Time
	mu             sync.Mutex
}

// windowElement - node in ring
type windowElement struct {
	value float64
	next  *windowElement
}

// NewPhiAccrualDetector -
func NewPhiAccrualDetector(hbTime time.Time, nOpts *NodeOptions, metadata *NodeMetadata) *PhiAccrualDetector {

	var windowSize int = nOpts.EstimationWindowSize

	// init phi-acc detector - create a very simple ring to iterate through
	window := make([]windowElement, windowSize)
	for i := 0; i < windowSize-1; i++ {
		window[i].next = &window[i+1]
	}
	window[windowSize-1].next = &window[0]

	return &PhiAccrualDetector{
		lastHeartbeat: hbTime,
		stats: &IntervalStatistics{
			rSum:          0.0,
			rSumSquares:   0.0,
			nTotalSamples: 0,
			windowSize:    windowSize,
		},
		window:         window,
		expiringSample: &window[0],
		metadata:       metadata,
	}
}

// AddValue - tack on a value to the statistics
func (phiD *PhiAccrualDetector) AddValue(ctx context.Context, arrivalTime time.Time) error {

	// safety first; note :: idk if we run into locking issues, only one goroutine should
	// ever be touching these values, but good to be safe...
	phiD.mu.Lock()
	defer phiD.mu.Unlock()

	timeDelta := float64(arrivalTime.Sub(phiD.lastHeartbeat) / time.Millisecond)
	var expTimeDelta float64 = phiD.expiringSample.value

	phiD.lastHeartbeat = arrivalTime

	// update value at the current ptr
	phiD.expiringSample.value = timeDelta
	phiD.expiringSample = phiD.expiringSample.next

	// upate the nTotalSamples, rSum, and rSumSquares for current collection
	phiD.stats.rSum += (timeDelta - expTimeDelta)
	phiD.stats.rSumSquares += (math.Pow(timeDelta, 2) - math.Pow(expTimeDelta, 2))
	phiD.stats.nTotalSamples++
	return nil
}

// Suspicion - rm
func (phiD *PhiAccrualDetector) Suspicion(ctime time.Time) float64 {
	return phiD.stats.Phi(phiD.lastHeartbeat, ctime)
}
