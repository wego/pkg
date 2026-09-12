// Package versionedcache holds a dataset from Redis in memory and reloads it only when a
// small version key next to that dataset says the data changed. A cycle that finds the
// version unchanged costs one GET of a short string instead of reading the whole dataset.
package versionedcache

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrFirstLoadInProgress is returned when there is nothing cached to serve yet and another
// caller is already doing the first load. Callers fall back to their own defaults; the next
// call after that load finishes gets the data.
var ErrFirstLoadInProgress = errors.New("versionedcache: first load is in progress")

// Options are the parts of a cache a caller may leave out. The zero value is fine: nothing
// is reported and the real clock is used.
type Options struct {
	// WhenRefreshed is called once per refresh cycle with what the cycle did and any error
	// that stopped it. This is where a caller emits its own metric and log line.
	WhenRefreshed func(RefreshOutcome, error)

	// Now replaces the clock. Tests set it; leave it nil in production.
	Now func() time.Time
}

// Cache holds one dataset together with the version it was loaded under, so a cycle that
// skips the reload can never serve a version whose data the cache does not hold.
type Cache[T any] struct {
	client     redis.UniversalClient
	versionKey string
	loadData   func(context.Context) (T, error)
	checkEvery time.Duration

	whenRefreshed func(RefreshOutcome, error)
	now           func() time.Time

	held       atomic.Pointer[heldData[T]]
	refreshing atomic.Bool
}

// heldData is one loaded dataset and the version it came with. It is replaced whole rather
// than edited in place, so readers holding the old pointer are never written underneath.
type heldData[T any] struct {
	data      T
	version   string
	checkedAt time.Time
}

// New builds a cache that reloads its data only when versionKey has changed.
//
// An empty versionKey means there is no version to check, so the data reloads every
// checkEvery. loadData must not be nil. client may be nil only when versionKey is empty.
func New[T any](
	client redis.UniversalClient,
	versionKey string,
	loadData func(context.Context) (T, error),
	checkEvery time.Duration,
	options Options,
) *Cache[T] {
	cache := &Cache[T]{
		client:        client,
		versionKey:    versionKey,
		loadData:      loadData,
		checkEvery:    checkEvery,
		whenRefreshed: options.WhenRefreshed,
		now:           options.Now,
	}
	if cache.now == nil {
		cache.now = time.Now
	}
	return cache
}

// Get returns the cached data, checking the version first when the interval has elapsed.
// It returns an error only when there is nothing cached to serve; once a load has
// succeeded, a later failure keeps the previous value and is reported through
// Options.WhenRefreshed instead.
func (c *Cache[T]) Get(ctx context.Context) (T, error) {
	held := c.held.Load()
	if held != nil && c.now().Sub(held.checkedAt) < c.checkEvery {
		return held.data, nil
	}

	// One refresher at a time. Losers serve the last good value rather than queue behind it.
	if !c.refreshing.CompareAndSwap(false, true) {
		if held != nil {
			return held.data, nil
		}
		var nothing T
		return nothing, ErrFirstLoadInProgress
	}
	defer c.refreshing.Store(false)

	return c.refresh(ctx, held)
}

// refresh reads the version, then the data if it has to. Only the elected refresher runs it.
func (c *Cache[T]) refresh(ctx context.Context, held *heldData[T]) (T, error) {
	// The version is read before the data on purpose. A cycle that overlaps a writer is then
	// stamped with the older version and reloads again next time, rather than advertising
	// data it does not hold.
	version, published, err := c.readVersion(ctx)
	if err != nil {
		return c.keepCachedData(held, err)
	}

	if published && held != nil && held.version == version {
		c.hold(held.data, version)
		c.report(KeptCachedData, nil)
		return held.data, nil
	}

	data, err := c.loadData(ctx)
	if err != nil {
		return c.keepCachedData(held, err)
	}

	c.hold(data, version)
	if published {
		c.report(ReloadedAfterChange, nil)
	} else {
		c.report(ReloadedWithoutVersion, nil)
	}
	return data, nil
}

// readVersion returns the published version. published is false when there is nothing to
// compare against — no version key was configured, or the key is not in Redis — which is
// the signal to reload unconditionally rather than an error.
func (c *Cache[T]) readVersion(ctx context.Context) (version string, published bool, err error) {
	if c.versionKey == "" {
		return "", false, nil
	}

	version, err = c.client.Get(ctx, c.versionKey).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", false, nil
	case err != nil:
		return "", false, fmt.Errorf("read version key %s: %w", c.versionKey, err)
	}
	return version, true, nil
}

// keepCachedData serves the last good value after a failed cycle, or gives up when there
// has never been one.
func (c *Cache[T]) keepCachedData(held *heldData[T], err error) (T, error) {
	c.report(KeptCachedData, err)
	if held != nil {
		return held.data, nil
	}
	var nothing T
	return nothing, err
}

// hold replaces what the cache serves, and restarts the interval.
func (c *Cache[T]) hold(data T, version string) {
	c.held.Store(&heldData[T]{data: data, version: version, checkedAt: c.now()})
}

func (c *Cache[T]) report(outcome RefreshOutcome, err error) {
	if c.whenRefreshed != nil {
		c.whenRefreshed(outcome, err)
	}
}
