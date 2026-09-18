package versionedcache

// RefreshOutcome says what one refresh cycle did. Callers turn it into a metric tag;
// this package deliberately depends on no metrics library of its own.
type RefreshOutcome string

const (
	// VersionUnchanged means the published version matched the one the cached data was
	// loaded under, so the data was not read. This is the cycle the package exists for.
	VersionUnchanged RefreshOutcome = "version_unchanged"

	// ReloadedAfterChange means the version differed from the one the cached data was
	// loaded under, so the data was read again. A first load with a version published
	// reports this too, because there is nothing cached to compare against.
	ReloadedAfterChange RefreshOutcome = "reloaded_after_change"

	// ReloadedWithoutVersion means there was no version to check — either the caller gave
	// no ReadVersion, or nothing is published — so the data was read again.
	ReloadedWithoutVersion RefreshOutcome = "reloaded_without_version"

	// ReloadedAtMaxStaleness means the version had not moved but the data had been kept for
	// Options.MaxStaleness, so it was read again in case the version stopped moving for some
	// reason other than the data being unchanged.
	ReloadedAtMaxStaleness RefreshOutcome = "reloaded_at_max_staleness"

	// RefreshFailed means reading the version or the data failed and the previous value is
	// still being served. The cache is warm, so the caller is running on slightly old data.
	RefreshFailed RefreshOutcome = "refresh_failed"

	// FirstLoadFailed means that same read failed with nothing ever loaded. There is
	// nothing to serve, Get returned the error, and the caller is on its own defaults.
	FirstLoadFailed RefreshOutcome = "first_load_failed"
)
