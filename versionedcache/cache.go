// Package versionedcache holds a dataset in memory and reloads it only when a small version
// published beside that dataset says the data changed. A cycle that finds the version
// unchanged costs one read of a short string instead of reading the whole dataset.
//
// Where the version lives is the caller's choice, handed in as a ReadVersion. VersionInKey and
// VersionInHashField cover the two Redis shapes. A nil ReadVersion means the dataset carries no
// version at all, so the data reloads every interval.
package versionedcache

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/redis/go-redis/v9"
)

// ReadVersion reports the version published for a dataset right now. published is false when
// there is nothing to compare against, which tells the cache to reload rather than to fail.
type ReadVersion func(ctx context.Context) (version string, published bool, err error)

// VersionInKey reads a version that has a Redis string key to itself, with GET.
func VersionInKey(client redis.UniversalClient, key string) ReadVersion {
	return func(ctx context.Context) (string, bool, error) {
		version, err := client.Get(ctx, key).Result()
		return publishedVersion(version, err, key)
	}
}

// VersionInHashField reads one field of a Redis hash, with HGET. Use it when a publisher keeps
// the versions of several datasets in one hash; each cache still reads its own field.
func VersionInHashField(client redis.UniversalClient, hashKey, field string) ReadVersion {
	location := hashKey + " field " + field
	return func(ctx context.Context) (string, bool, error) {
		version, err := client.HGet(ctx, hashKey, field).Result()
		return publishedVersion(version, err, location)
	}
}

// publishedVersion turns one Redis reply into the answer ReadVersion owes the cache. A key or
// field that is not there is not a failure: it means nobody publishes one, so the data reloads.
func publishedVersion(version string, err error, location string) (string, bool, error) {
	switch {
	case errors.Is(err, redis.Nil):
		return "", false, nil
	case err != nil:
		return "", false, fmt.Errorf("read version from %s: %w", location, err)
	}
	return version, true, nil
}

// Options are the parts of a cache a caller may leave out. The zero value is fine: nothing
// is reported and the real clock is used.
type Options struct {
	// WhenRefreshed is called once per refresh cycle with what the cycle did and any error
	// that stopped it. This is where a caller emits its own metric and log line.
	//
	// It runs inside the cycle, so calling Get on this cache from here during the first load
	// waits on the load it is itself part of and never returns. On a later cycle that Get is
	// answered from memory instead. A panic here is caught and the report dropped.
	WhenRefreshed func(RefreshOutcome, error)

	// MaxStaleness is how long the data may be kept on an unchanged version before it is read
	// again anyway, at the next cycle. Zero means DefaultMaxStaleness. It is the way out when a
	// version stops moving for some reason other than the data being unchanged, such as a
	// writer that changed rows and died before publishing.
	MaxStaleness time.Duration

	// Now replaces the clock. Tests set it; leave it nil in production.
	Now func() time.Time
}

// DefaultMaxStaleness is the limit a cache runs on when Options.MaxStaleness is left zero.
const DefaultMaxStaleness = 6 * time.Hour

// theDataset is the key the one dataset is stored under. A cache holds a single dataset, so
// the key never varies; the store underneath is keyed and needs one.
const theDataset = "dataset"

// cacheEntry is one loaded dataset and the version it came with. They are kept together so a
// cycle that finds the version unchanged can prove the data it keeps was loaded under it.
type cacheEntry[T any] struct {
	data    T
	version string
	// published says a version was actually read for this entry. Without it the empty string is
	// both "nobody publishes one" and a legitimate published value, and the two compare equal.
	published bool
	// loadedAt is when the data was last read. A cycle that keeps the data on an unchanged
	// version leaves it alone, so the staleness limit counts from the read, not the check.
	loadedAt time.Time
}

// Cache holds one dataset together with the version it was loaded under, so a cycle that skips
// the reload can never serve a version whose data the cache does not hold. Holding the value,
// timing the interval and electing one refresher are otter's; the version comparison and what
// each cycle reports are here.
type Cache[T any] struct {
	cached *otter.Cache[string, cacheEntry[T]]
	cycle  *refreshCycle[T]
}

// New builds a cache that reloads its data only when readVersion reports a different version,
// or when the data has been kept past Options.MaxStaleness on a version that has not moved.
//
// A nil readVersion means the dataset carries no version, so the data reloads every checkEvery,
// which is what every caller did before this package existed.
//
// New panics on a nil loadData, a checkEvery that is not positive, or an Options.MaxStaleness
// shorter than checkEvery — including a negative one, and including the DefaultMaxStaleness a
// zero takes. Each of those builds a cache that compiles and then never refreshes properly, so
// it is refused here rather than at the first cycle.
func New[T any](
	readVersion ReadVersion,
	loadData func(context.Context) (T, error),
	checkEvery time.Duration,
	options Options,
) *Cache[T] {
	if loadData == nil {
		panic("versionedcache: loadData must not be nil")
	}
	// An interval that is not positive never refreshes at all.
	if checkEvery <= 0 {
		panic("versionedcache: checkEvery must be greater than zero")
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	maxStaleness := options.MaxStaleness
	if maxStaleness == 0 {
		maxStaleness = DefaultMaxStaleness
	}
	// A limit the interval always clears puts every cycle past it, so the whole dataset is read
	// every time and the version decides nothing. Checked after the default is filled in, because
	// a zero left with a check interval longer than six hours lands in the same place.
	if maxStaleness < checkEvery {
		panic(fmt.Sprintf(
			"versionedcache: Options.MaxStaleness (%s) must not be shorter than checkEvery (%s)",
			maxStaleness, checkEvery))
	}
	store := &otter.Options[string, cacheEntry[T]]{
		RefreshCalculator: refreshEvery[T]{interval: checkEvery},
		// Every failure is already reported through WhenRefreshed, where the caller's log line
		// and metric live. otter's own logger writes to slog, which is not this service's file.
		Logger: &otter.NoopLogger{},
	}
	if options.Now != nil {
		store.Clock = &clockFrom{now: options.Now}
	}
	return &Cache[T]{
		cached: otter.Must(store),
		cycle: &refreshCycle[T]{
			readVersion:   readVersion,
			loadData:      loadData,
			interval:      checkEvery,
			maxStaleness:  maxStaleness,
			whenRefreshed: options.WhenRefreshed,
			now:           now,
		},
	}
}

// Get returns the cached data, and starts a refresh once the interval has elapsed. The refresh
// runs in the background: this call is answered with the data the cache already holds, and a
// changed dataset reaches a later call. It returns an error only when nothing has ever loaded.
func (c *Cache[T]) Get(ctx context.Context) (T, error) {
	entry, err := c.cached.Get(ctx, theDataset, c.cycle)
	if err != nil {
		var nothing T
		return nothing, err
	}
	return entry.data, nil
}

// refreshCycle is one cache's refresh: the first read, and every cycle after it. The store
// calls Load with nothing cached and Reload with the data the cache is already holding.
type refreshCycle[T any] struct {
	readVersion   ReadVersion
	loadData      func(context.Context) (T, error)
	interval      time.Duration
	maxStaleness  time.Duration
	whenRefreshed func(RefreshOutcome, error)
	now           func() time.Time

	// startedAt is when the last cycle began, nil before the first one. Kept as a time.Time so the
	// comparison below counts elapsed time the way the store's own clock does, and a clock set
	// backwards does not stall every cycle after it.
	startedAt atomic.Pointer[time.Time]

	// newest is what the last finished cycle produced. The store hands Reload the value the
	// cache held when that call's Get ran, which is out of date for a cycle that was queued
	// behind another one, and comparing against an out-of-date version re-reads a dataset the
	// cycle in front has just read.
	newest atomic.Pointer[cacheEntry[T]]
}

// Load is the first read. A failure has no previous value to fall back to, so the error travels
// back to Get, which is how a caller learns it is running on its own defaults.
func (c *refreshCycle[T]) Load(ctx context.Context, _ string) (cacheEntry[T], error) {
	version, published, err := c.currentVersion(ctx)
	if err != nil {
		return c.failed(FirstLoadFailed, err)
	}

	data, err := c.loadDataNow(ctx)
	if err != nil {
		return c.failed(FirstLoadFailed, err)
	}

	c.report(reloadReason(published, false), nil)
	return c.loadedNow(data, version, published), nil
}

// Reload is every cycle after the first. A cycle that cannot read reports the failure and hands
// the error back: the cache keeps the data it already holds, and refreshEvery below waits out
// the interval before the next attempt.
func (c *refreshCycle[T]) Reload(ctx context.Context, _ string, cached cacheEntry[T]) (cacheEntry[T], error) {
	if newest := c.newest.Load(); newest != nil {
		cached = *newest
	}

	if !c.startCycle() {
		return cached, nil
	}

	// The version is read before the data on purpose. A cycle that overlaps a writer is then
	// stamped with the older version and reloads again next time, rather than advertising
	// data it does not hold.
	version, published, err := c.currentVersion(ctx)
	if err != nil {
		return c.failed(RefreshFailed, err)
	}

	versionUnchanged := published && cached.published && cached.version == version
	if versionUnchanged && !c.keptPastMaxStaleness(cached) {
		c.report(VersionUnchanged, nil)
		return cached, nil
	}

	data, err := c.loadDataNow(ctx)
	if err != nil {
		return c.failed(RefreshFailed, err)
	}

	c.report(reloadReason(published, versionUnchanged), nil)
	return c.loadedNow(data, version, published), nil
}

// startCycle claims the interval, so only the first caller that finds the data due goes on to read
// it. The store puts the data back on the clock when a cycle finishes, not when it starts, so a
// slow failing read leaves it due meanwhile and every call arriving starts its own cycle — reading
// a source already known to be down over and over.
//
// A cycle that loses the claim reports nothing, because none ran.
func (c *refreshCycle[T]) startCycle() bool {
	for {
		started := c.startedAt.Load()
		now := c.now()
		// A negative gap means the clock was set backwards. Run the cycle rather than wait for the
		// clock to reach a time it has already been. Reading the clock inside the loop is what
		// makes that safe: a caller that loses the swap reads it again, so a negative gap is only
		// ever a clock that really moved, not a caller that sampled early and lost.
		if started != nil {
			if since := now.Sub(*started); since >= 0 && since < c.interval {
				return false
			}
		}
		if c.startedAt.CompareAndSwap(started, &now) {
			return true
		}
	}
}

// failed reports a cycle that could not read and hands the error back. A refresh keeps whatever
// the cache already holds; a first load has nothing to keep, so the error reaches Get instead.
func (c *refreshCycle[T]) failed(outcome RefreshOutcome, err error) (cacheEntry[T], error) {
	c.report(outcome, err)
	var nothing cacheEntry[T]
	if errors.Is(err, otter.ErrNotFound) {
		// The store reads that error as the dataset being gone and throws away what it holds. A
		// loader reporting it means the read failed, so flatten it rather than pass it on.
		return nothing, fmt.Errorf("versionedcache: %s", err)
	}
	return nothing, err
}

// currentVersion asks the caller's ReadVersion, treating a cache that has none as a dataset
// nobody publishes a version for: nothing to compare, so reload.
func (c *refreshCycle[T]) currentVersion(ctx context.Context) (version string, published bool, err error) {
	if c.readVersion == nil {
		return "", false, nil
	}
	defer func() {
		if r := recover(); r != nil {
			version, published, err = "", false, recovered("ReadVersion", r)
		}
	}()
	return c.readVersion(ctx)
}

// loadDataNow reads the dataset, turning a panic into an error. The store runs a background reload
// on a goroutine of its own and re-raises whatever the loader panicked with, so a panic left alone
// there ends the process instead of failing one cycle.
func (c *refreshCycle[T]) loadDataNow(ctx context.Context) (data T, err error) {
	defer func() {
		if r := recover(); r != nil {
			var nothing T
			data, err = nothing, recovered("loadData", r)
		}
	}()
	return c.loadData(ctx)
}

// recovered carries the stack into the error, because nothing else will print it: the cycle
// reports through WhenRefreshed and the goroutine it ran on ends quietly.
func recovered(what string, panicked any) error {
	return fmt.Errorf("versionedcache: %s panicked: %v\n%s", what, panicked, debug.Stack())
}

// keptPastMaxStaleness says whether the data has been kept for MaxStaleness or longer since it
// was last read. It is asked only on a cycle, so the read lands on the first cycle past the limit.
func (c *refreshCycle[T]) keptPastMaxStaleness(cached cacheEntry[T]) bool {
	return c.now().Sub(cached.loadedAt) >= c.maxStaleness
}

// loadedNow is the entry for data read this cycle, with the staleness clock started again.
func (c *refreshCycle[T]) loadedNow(data T, version string, published bool) cacheEntry[T] {
	loaded := cacheEntry[T]{data: data, version: version, published: published, loadedAt: c.now()}
	c.newest.Store(&loaded)
	return loaded
}

// reloadReason names why a cycle read the data: nothing to compare against, a version that
// moved, or one that stayed put for longer than the data may be kept.
func reloadReason(published, versionUnchanged bool) RefreshOutcome {
	switch {
	case !published:
		return ReloadedWithoutVersion
	case versionUnchanged:
		return ReloadedAtMaxStaleness
	default:
		return ReloadedAfterChange
	}
}

// report hands the cycle's outcome to the caller. It runs on the store's refresh goroutine, like
// the two reads above, so a panic here would end the process. Nothing can report a broken
// reporter, so the report is dropped and the next cycle still runs.
func (c *refreshCycle[T]) report(outcome RefreshOutcome, err error) {
	if c.whenRefreshed == nil {
		return
	}
	defer func() { _ = recover() }()
	c.whenRefreshed(outcome, err)
}
