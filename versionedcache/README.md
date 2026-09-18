# versionedcache

Holds a dataset in memory and reloads it only when a small version published
beside that dataset says the data changed. A cycle that finds the version unchanged costs one
read of a short string instead of reading the whole dataset.

Holding the value, timing the interval and running one refresh at a time are
[otter](https://github.com/maypok86/otter)'s, configured here as refresh-after-write. What this
package adds is the version comparison that decides whether a cycle reads the data at all, and
the outcome it reports for each cycle.

The dataset's owner publishes the version, writing it after the data, so a value is never
advertised for data that has not landed. This package only reads it.

## Use

`New` takes a `ReadVersion`, which says where the version lives. Two are provided.

**One field of a hash**, read with `HGET`. Use this when the publisher keeps the versions of
several datasets in a single hash. Each cache reads its own field with its own `HGET`, so the
shared hash is the publisher's layout rather than a batched read.

```go
settings := versionedcache.New(
	versionedcache.VersionInHashField(rdb, "app:versions", "settings"),
	func(ctx context.Context) (map[string]Config, error) {
		return readTheWholeHash(ctx, rdb)
	},
	30*time.Second,
	versionedcache.Options{
		WhenRefreshed: func(outcome versionedcache.RefreshOutcome, err error) {
			metrics.Count("config.refresh", "outcome:"+string(outcome))
		},
	},
)

configs, err := settings.Get(ctx)
```

`New` refuses a cache that would be quietly wrong for the life of the process, and panics on a
nil loader, a `checkEvery` that is not positive, or a negative `Options.MaxStaleness`. There is no
error to return, and each of those builds a cache that compiles and then never refreshes properly.

**A string key of its own**, read with `GET`. Use this when the version has no hash to sit in.

```go
versionedcache.New(
	versionedcache.VersionInKey(rdb, "app:settings:version"),
	loadData, 30*time.Second, versionedcache.Options{},
)
```

**No version at all**: pass `nil`. The data then reloads every interval, which is what every
caller did before this package existed. A dataset small enough not to be worth gating wants
this, and a cache built that way needs no Redis behind it at all.

A version that is configured but not published reads the same way. A key or field that is not
there means nobody publishes one, so deleting it is an off switch that needs no deploy.

**A version that stops moving** is not proof that the data is unchanged: a writer can change
rows and die before it publishes, or a bulk import can publish once at the very end. So data
kept on an unchanged version for `Options.MaxStaleness` is read again at the next cycle anyway,
and that read restarts the clock; a cycle that skipped does not. Zero means
`DefaultMaxStaleness`, six hours. Set it against how often the writer runs, and keep it longer
than the check interval, or every cycle reloads.

`Get` returns `T` by value, so it copies whatever you store on every call. Hold a map, a slice
or a pointer when the dataset is large.

## What each cycle reports

`Options.WhenRefreshed` is called once per cycle, from inside the load and a moment before the
value that cycle produced is stored. The outcome carries the whole classification, so a caller's
metric tag is `string(outcome)` with nothing to branch on.

| Outcome | Meaning |
| --- | --- |
| `version_unchanged` | the version matched, so nothing was read |
| `reloaded_after_change` | the version moved, so the data was read again |
| `reloaded_without_version` | there was no version to compare, so the data was read again |
| `reloaded_at_max_staleness` | the version had not moved but the data had been kept for `MaxStaleness`, so it was read again |
| `refresh_failed` | the read failed and the previous value is still being served |
| `first_load_failed` | the read failed with nothing ever loaded, so `Get` returned the error |

The `error` argument carries detail for the log line, never the classification.

## Three things that surprise people

**A refresh runs in the background.** The call that finds the interval elapsed is answered with
the data the cache already holds, and starts the refresh; a dataset that changed reaches a later
call. Nothing waits on Redis once the cache is warm, and the worst-case staleness is the interval
plus one refresh. The refresh also runs with the caller's cancellation stripped, so a deadline on the
`Get` that started it does not cut it short; the Redis client's own timeouts bound it instead.

**Get returns an error only when it has never loaded.** Once a load has succeeded, a failure
to read either the version or the data keeps the previous value and `Get` returns it with a
nil error. The failure reaches you through `Options.WhenRefreshed`, which is where your log
line and metric belong. A Redis outage therefore does not empty a warm cache, and while one lasts
the source is still read only once an interval however much traffic arrives.

The first load is the exception on both counts: it runs on the calling goroutine, and callers
arriving while it is still running wait for it and are handed what it produced. Only when that
load fails does `Get` return an error, and the next call tries again straight away.

**The version is compared as exact text, never as an ordering.** Any different value counts
as changed — a newer one, an older one restored from a backup, or a different shape
entirely. That means the publisher only has to make the value change; it does not have to
make it increase.
