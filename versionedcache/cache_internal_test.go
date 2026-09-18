package versionedcache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// otter hands Reload the value the cache held when the refreshing Get ran, captured when the
// refresh was scheduled rather than when it runs. A cycle that starts while another is still
// finishing therefore sees an out-of-date entry, and comparing against its version re-reads the
// dataset the cycle in front has just read. The cycle is driven directly here because that
// window is a few instructions wide inside the library and cannot be forced from outside.
func TestAReloadHandedAnOutOfDateEntryComparesAgainstTheLastCycleInstead(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	reads := 0

	cycle := &refreshCycle[string]{
		readVersion:  func(context.Context) (string, bool, error) { return "v2", true, nil },
		loadData:     func(context.Context) (string, error) { reads++; return "data for v2", nil },
		maxStaleness: time.Hour,
		now:          func() time.Time { return now },
	}

	outOfDate := cacheEntry[string]{data: "data for v1", version: "v1", published: true, loadedAt: now}

	// The version has moved, so this cycle reads the dataset and records what it loaded.
	loaded, err := cycle.Reload(ctx, theDataset, outOfDate)
	require.NoError(t, err)
	require.Equal(t, "v2", loaded.version)
	require.Equal(t, 1, reads)

	// The next cycle is handed the same out-of-date entry, which is what a queued one gets.
	kept, err := cycle.Reload(ctx, theDataset, outOfDate)
	require.NoError(t, err)
	assert.Equal(t, 1, reads, "a cycle must compare against what the last finished cycle loaded")
	assert.Equal(t, "data for v2", kept.data)
}

// A caller's clock with no monotonic reading, which is the case the guard exists for: time.Now
// carries one and Sub counts on it whatever the wall clock does, so only a clock without one can
// report time going backwards. Driven directly because Options.Now is the store's clock too, so a
// test cannot move just one of them.
func TestCyclesAreNotHeldBackByAClockWithNoMonotonicReading(t *testing.T) {
	// Off the real clock, as every fixture here is, with Round(0) dropping the monotonic reading.
	now := time.Now().Round(0)
	cycle := &refreshCycle[string]{interval: time.Minute, now: func() time.Time { return now }}

	require.True(t, cycle.startCycle(), "the first cycle has nothing to wait behind")
	require.False(t, cycle.startCycle(), "a second inside the interval waits for it")

	now = now.Add(-time.Hour)
	assert.True(t, cycle.startCycle(), "a clock that moved back must not hold the cycle back with it")
}

// startCycle's contract is that only one cycle holds the interval. A caller that loses the swap
// re-reads the clock, so the gap it compares is against a start that really is in the past.
func TestOnlyOneCycleHoldsTheIntervalWhenAnotherClaimsItMidSwap(t *testing.T) {
	base := time.Now().Round(0)
	cycle := &refreshCycle[string]{interval: time.Minute}

	reads := 0
	cycle.now = func() time.Time {
		reads++
		if reads == 1 {
			// Another cycle reads the clock a second later and wins the claim, which is the
			// window this caller's swap has to survive.
			won := base.Add(2 * time.Second)
			cycle.startedAt.Store(&won)
			return base.Add(1 * time.Second)
		}
		return base.Add(3 * time.Second)
	}

	assert.False(t, cycle.startCycle(),
		"a cycle that lost the swap must re-read the clock, not read its own stale sample as a clock that moved backwards")
}
