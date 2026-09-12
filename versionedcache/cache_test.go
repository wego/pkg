package versionedcache_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wego/pkg/versionedcache"
)

// fakeClock lets a test move time forward without sleeping. Guarded by a mutex
// because the concurrency test reads it from several goroutines under -race.
type fakeClock struct {
	mutex sync.Mutex
	now   time.Time
}

// newFakeClock starts from the real clock, so no test carries a typed-in date.
func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Now()}
}

func (c *fakeClock) Now() time.Time {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.now
}

func (c *fakeClock) Advance(by time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.now = c.now.Add(by)
}

// countingLoader returns a loader that answers with the current value of answer,
// and a pointer to how many times it has been called.
func countingLoader(answer *string) (func(context.Context) (string, error), *int) {
	calls := 0
	return func(context.Context) (string, error) {
		calls++
		return *answer, nil
	}, &calls
}

func newTestClient(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return server, client
}

func TestNoVersionKeyReloadsEveryInterval(t *testing.T) {
	_, client := newTestClient(t)
	clock := newFakeClock()
	answer := "first"
	loadData, loads := countingLoader(&answer)

	cache := versionedcache.New(client, "", loadData, time.Minute, versionedcache.Options{Now: clock.Now})

	data, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", data)
	assert.Equal(t, 1, *loads)

	answer = "second"
	clock.Advance(2 * time.Minute)

	data, err = cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "second", data)
	assert.Equal(t, 2, *loads)
}

func TestDataIsServedFromMemoryInsideTheInterval(t *testing.T) {
	_, client := newTestClient(t)
	clock := newFakeClock()
	answer := "first"
	loadData, loads := countingLoader(&answer)

	cache := versionedcache.New(client, "", loadData, time.Minute, versionedcache.Options{Now: clock.Now})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	answer = "second"
	clock.Advance(30 * time.Second)

	data, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", data, "inside the interval the cache must not reload")
	assert.Equal(t, 1, *loads)
}

func TestRefreshWithoutAVersionKeyIsReported(t *testing.T) {
	_, client := newTestClient(t)
	answer := "first"
	loadData, _ := countingLoader(&answer)

	var outcomes []versionedcache.RefreshOutcome
	cache := versionedcache.New(client, "", loadData, time.Minute, versionedcache.Options{
		WhenRefreshed: func(outcome versionedcache.RefreshOutcome, err error) {
			require.NoError(t, err)
			outcomes = append(outcomes, outcome)
		},
	})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []versionedcache.RefreshOutcome{versionedcache.ReloadedWithoutVersion}, outcomes)
}

const exampleVersionKey = "flight:example:version"

func TestUnchangedVersionDoesNotReadTheData(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	answer := "first"
	loadData, loads := countingLoader(&answer)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	cache := versionedcache.New(client, exampleVersionKey, loadData, time.Minute, versionedcache.Options{Now: clock.Now})

	data, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", data)
	assert.Equal(t, 1, *loads)

	// The data behind the version changed but the version did not: the cache must not see it.
	answer = "second"
	clock.Advance(2 * time.Minute)

	data, err = cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", data)
	assert.Equal(t, 1, *loads, "an unchanged version must not read the data")
}

func TestChangedVersionReloadsTheData(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	answer := "first"
	loadData, loads := countingLoader(&answer)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	cache := versionedcache.New(client, exampleVersionKey, loadData, time.Minute, versionedcache.Options{Now: clock.Now})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	answer = "second"
	require.NoError(t, server.Set(exampleVersionKey, "20260912.2"))
	clock.Advance(2 * time.Minute)

	data, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "second", data)
	assert.Equal(t, 2, *loads)
}

// A version that moves backwards, or to a different shape, still counts as changed:
// the comparison is exact text, never an ordering.
func TestAnyDifferentVersionReloadsTheData(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	answer := "first"
	loadData, loads := countingLoader(&answer)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.9"))

	cache := versionedcache.New(client, exampleVersionKey, loadData, time.Minute, versionedcache.Options{Now: clock.Now})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	answer = "restored from a backup"
	require.NoError(t, server.Set(exampleVersionKey, "20260101.1"))
	clock.Advance(2 * time.Minute)

	data, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "restored from a backup", data)
	assert.Equal(t, 2, *loads)
}

func TestMissingVersionKeyReloadsRatherThanFailing(t *testing.T) {
	_, client := newTestClient(t)
	clock := newFakeClock()
	answer := "first"
	loadData, loads := countingLoader(&answer)

	// The key is configured but nobody has published it yet: the writer has not shipped.
	cache := versionedcache.New(client, exampleVersionKey, loadData, time.Minute, versionedcache.Options{Now: clock.Now})

	data, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", data)

	answer = "second"
	clock.Advance(2 * time.Minute)

	data, err = cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "second", data, "with no version published the cache reloads every interval")
	assert.Equal(t, 2, *loads)
}

func TestOutcomesTellTheThreeCasesApart(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	answer := "first"
	loadData, _ := countingLoader(&answer)

	var outcomes []versionedcache.RefreshOutcome
	cache := versionedcache.New(client, exampleVersionKey, loadData, time.Minute, versionedcache.Options{
		Now: clock.Now,
		WhenRefreshed: func(outcome versionedcache.RefreshOutcome, err error) {
			require.NoError(t, err)
			outcomes = append(outcomes, outcome)
		},
	})

	// No version published yet.
	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	// A version appears.
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))
	clock.Advance(2 * time.Minute)
	_, err = cache.Get(context.Background())
	require.NoError(t, err)

	// The same version again.
	clock.Advance(2 * time.Minute)
	_, err = cache.Get(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []versionedcache.RefreshOutcome{
		versionedcache.ReloadedWithoutVersion,
		versionedcache.ReloadedAfterChange,
		versionedcache.KeptCachedData,
	}, outcomes)
}
