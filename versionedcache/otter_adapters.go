package versionedcache

import (
	"time"

	"github.com/maypok86/otter/v2"
)

// refreshEvery puts the dataset back on the clock after every cycle, a failed one included.
// otter's own RefreshWriting leaves a failed reload still due, which would send the very next
// call back to a source already known to be down.
type refreshEvery[T any] struct{ interval time.Duration }

func (r refreshEvery[T]) RefreshAfterCreate(otter.Entry[string, cacheEntry[T]]) time.Duration {
	return r.interval
}

func (r refreshEvery[T]) RefreshAfterUpdate(otter.Entry[string, cacheEntry[T]], cacheEntry[T]) time.Duration {
	return r.interval
}

func (r refreshEvery[T]) RefreshAfterReload(otter.Entry[string, cacheEntry[T]], cacheEntry[T]) time.Duration {
	return r.interval
}

func (r refreshEvery[T]) RefreshAfterReloadFailure(otter.Entry[string, cacheEntry[T]], error) time.Duration {
	return r.interval
}

// clockFrom lets Options.Now drive the interval. Tick is only asked for by a cache that expires
// entries early, which this one never configures, so the real ticker is right there.
type clockFrom struct{ now func() time.Time }

func (c *clockFrom) NowNano() int64 { return c.now().UnixNano() }

func (c *clockFrom) Tick(every time.Duration) <-chan time.Time { return time.Tick(every) }
