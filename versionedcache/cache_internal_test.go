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

// The store's clock never goes backwards, so it keeps asking for cycles however the wall clock
// moves. Counting the interval on wall time would stall them for the whole of a clock correction.
// Driven directly because Options.Now is the store's clock too, so a test cannot move just one.
func TestCyclesAreNotHeldBackByAClockThatMovesBackwards(t *testing.T) {
	now := time.Now()
	cycle := &refreshCycle[string]{interval: time.Minute, now: func() time.Time { return now }}

	require.True(t, cycle.startCycle(), "the first cycle has nothing to wait behind")
	require.False(t, cycle.startCycle(), "a second inside the interval waits for it")

	now = now.Add(-time.Hour)
	assert.True(t, cycle.startCycle(), "a clock that moved back must not hold the cycle back with it")
}
