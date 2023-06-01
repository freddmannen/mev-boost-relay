package api

import (
	"sync"

	"go.uber.org/atomic"
)

type bidStatsEntry struct {
	slot        uint64
	payloadSize uint64
}

func newBidStatsEntry(slot, payloadSize uint64) *bidStatsEntry {
	return &bidStatsEntry{
		slot:        slot,
		payloadSize: payloadSize,
	}
}

type BidStats struct {
	lock            sync.RWMutex
	entries         []*bidStatsEntry
	entriesForTally uint64

	payloadCnt     uint64
	payloadSizeAvg atomic.Float64 // running tally
}

func NewBidStats(entriesForTally uint64) *BidStats {
	return &BidStats{ //nolint:exhaustruct
		entries:         make([]*bidStatsEntry, 0),
		entriesForTally: entriesForTally,
	}
}

func (b *BidStats) AddEntry(slot uint64, payloadSize int) {
	b.lock.Lock()
	// add to entries
	b.entries = append(b.entries, newBidStatsEntry(slot, uint64(payloadSize)))

	// truncate if needed
	if len(b.entries) > int(b.entriesForTally) {
		b.entries = b.entries[1:]
	}
	b.lock.Unlock()

	// recalc stats
	_, _, avg := b.GetPayloadSizeStatsN(b.entriesForTally)
	b.payloadSizeAvg.Store(avg)
	b.payloadCnt++
}

// GetPayloadSizeStatsN returns stats about n entries, going backwards starting at slot n
func (b *BidStats) GetPayloadSizeStatsN(maxEntries uint64) (total, count uint64, avg float64) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if len(b.entries) == 0 {
		return 0, 0, 0
	}

	var startIdx int
	if len(b.entries) > int(maxEntries) {
		startIdx = len(b.entries) - int(maxEntries)
	}

	for i := startIdx; i < len(b.entries); i++ {
		total += b.entries[i].payloadSize
		count++
	}

	avg = float64(total) / float64(count)
	return total, count, avg
}

// PayloadSizeDeviation returns the payloadSize deviation from the running average
// - returns -1..1 (i.e. 50% below average is -0.5)
// - returns 0 until `entriesForTally` entries are added
func (b *BidStats) PayloadSizeDeviation(payloadSize uint64) (deviationPercent float64) {
	prevAvg := b.payloadSizeAvg.Load()
	if prevAvg == 0 || b.payloadCnt < b.entriesForTally {
		return 0
	}
	return (float64(payloadSize) - prevAvg) / prevAvg
}
