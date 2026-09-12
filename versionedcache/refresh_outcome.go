package versionedcache

// RefreshOutcome says what one refresh cycle did. Callers turn it into a metric tag;
// this package deliberately depends on no metrics library of its own.
type RefreshOutcome string

const (
	// KeptCachedData means the cycle did not reload: either the version was unchanged,
	// or reading the version or the data failed and the previous value was kept. The
	// error passed alongside is what separates those two.
	KeptCachedData RefreshOutcome = "kept_cached_data"

	// ReloadedAfterChange means the version differed from the one the held data was
	// loaded under, so the data was read again. A first load with a version key present
	// reports this too, because there is nothing held to compare against.
	ReloadedAfterChange RefreshOutcome = "reloaded_after_change"

	// ReloadedWithoutVersion means there was no version to check — either the caller
	// gave no version key, or the key is not in Redis — so the data was read again.
	ReloadedWithoutVersion RefreshOutcome = "reloaded_without_version"
)
