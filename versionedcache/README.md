# versionedcache

Holds a dataset from Redis in memory and reloads it only when a small version key next to
that dataset says the data changed. A cycle that finds the version unchanged costs one `GET`
of a short string instead of reading the whole dataset.

The dataset's owner publishes the version key, writing it after the data, so a value is
never advertised for data that has not landed. This package only reads it.

## Use

```go
providerConfigs := versionedcache.New(
	staticRedis,
	"flight:breakerConfig:version",
	func(ctx context.Context) (map[string]Config, error) {
		return readTheWholeHash(ctx, staticRedis)
	},
	30*time.Second,
	versionedcache.Options{
		WhenRefreshed: func(outcome versionedcache.RefreshOutcome, err error) {
			metrics.Count("config.refresh", "outcome:"+string(outcome))
		},
	},
)

configs, err := providerConfigs.Get(ctx)
```

Pass an empty version key for a dataset with no version published: the data then reloads
every interval, which is what every caller did before this package existed.

## Two things that surprise people

**Get returns an error only when it has never loaded.** Once a load has succeeded, a failure
to read either the version or the data keeps the previous value and `Get` returns it with a
nil error. The failure reaches you through `Options.WhenRefreshed`, which is where your log
line and metric belong. A Redis outage therefore does not empty a warm cache.

**The version is compared as exact text, never as an ordering.** Any different value counts
as changed — a newer one, an older one restored from a backup, or a different shape
entirely. That means the publisher only has to make the value change; it does not have to
make it increase.
