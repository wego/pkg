package versionedcache_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/maypok86/otter/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wego/pkg/versionedcache"
)

// fakeClock lets a test move time forward without sleeping. Guarded by a mutex
// because the cache reads it from its refresh goroutines under -race.
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

// answers is the data the loader hands back, and a count of how often it was asked. Guarded:
// the loader runs on a refresh goroutine while the test changes what it should answer.
type answers struct {
	mutex sync.Mutex
	value string
	reads atomic.Int64
}

func newAnswers(value string) *answers {
	return &answers{value: value}
}

func (a *answers) set(value string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.value = value
}

func (a *answers) load(context.Context) (string, error) {
	a.reads.Add(1)
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.value, nil
}

func (a *answers) timesRead() int {
	return int(a.reads.Load())
}

// cycles records what every refresh cycle reported and lets a test wait for one. Every cycle
// after the first load runs in the background, so a test waits for it rather than guessing
// how long it takes.
type cycles struct {
	t        *testing.T
	mutex    sync.Mutex
	outcomes []versionedcache.RefreshOutcome
	errs     []error
}

func newCycles(t *testing.T) *cycles {
	return &cycles{t: t}
}

func (c *cycles) record(outcome versionedcache.RefreshOutcome, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.outcomes = append(c.outcomes, outcome)
	c.errs = append(c.errs, err)
}

// waitUntil blocks until this many cycles have finished reporting. Counted rather than awaited
// one at a time, so the cycles a test has already seen cannot be mistaken for the one it awaits.
func (c *cycles) waitUntil(count int) {
	c.t.Helper()
	require.Eventually(c.t, func() bool { return c.count() >= count },
		5*time.Second, 5*time.Millisecond, "waited for %d refresh cycles", count)
}

func (c *cycles) count() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.outcomes)
}

func (c *cycles) seen() []versionedcache.RefreshOutcome {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return slices.Clone(c.outcomes)
}

func (c *cycles) lastError() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.errs) == 0 {
		return nil
	}
	return c.errs[len(c.errs)-1]
}

func (c *cycles) failures() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	failed := 0
	for _, err := range c.errs {
		if err != nil {
			failed++
		}
	}
	return failed
}

// refreshCycle asks for the data and waits for the background cycle that the elapsed interval
// starts. The call itself is answered with whatever the cache already holds.
func refreshCycle(t *testing.T, cache *versionedcache.Cache[string], reported *cycles) {
	t.Helper()
	sofar := reported.count()
	_, err := cache.Get(context.Background())
	require.NoError(t, err)
	reported.waitUntil(sofar + 1)
}

// assertServes waits for the cache to be answering with want. A cycle reports from inside the
// load, a moment before the value it loaded is stored, so a read straight after the cycle can
// still be answered with the old one. Read the counts a cycle changed before calling this: a
// read landing in that moment starts a further cycle of its own.
func assertServes(t *testing.T, cache *versionedcache.Cache[string], want string) {
	t.Helper()
	require.Eventually(t, func() bool {
		served, err := cache.Get(context.Background())
		return err == nil && served == want
	}, 2*time.Second, 20*time.Millisecond, "the cache never served %q", want)
}

func newTestClient(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return server, client
}

const exampleVersionKey = "example:version"

const (
	exampleVersionHashKey = "example:versions"
	exampleVersionField   = "example"
)

// versionPlace is one of the two places a version can live. A behaviour that has to hold
// wherever it lives is written once here and run for each place, so the two cannot drift apart.
type versionPlace struct {
	name    string
	publish func(*testing.T, *miniredis.Miniredis, string)
	build   func(*redis.Client, func(context.Context) (string, error), versionedcache.Options) *versionedcache.Cache[string]
}

func versionPlaces() []versionPlace {
	return []versionPlace{
		{
			name: "a string key of its own",
			publish: func(t *testing.T, server *miniredis.Miniredis, version string) {
				require.NoError(t, server.Set(exampleVersionKey, version))
			},
			build: func(client *redis.Client, loadData func(context.Context) (string, error),
				options versionedcache.Options) *versionedcache.Cache[string] {
				return versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute, options)
			},
		},
		{
			name: "one field of a shared hash",
			publish: func(_ *testing.T, server *miniredis.Miniredis, version string) {
				server.HSet(exampleVersionHashKey, exampleVersionField, version)
			},
			build: func(client *redis.Client, loadData func(context.Context) (string, error),
				options versionedcache.Options) *versionedcache.Cache[string] {
				return versionedcache.New(versionedcache.VersionInHashField(client,
					exampleVersionHashKey, exampleVersionField), loadData, time.Minute, options)
			},
		},
	}
}

func TestAnUnchangedVersionDoesNotReadTheData(t *testing.T) {
	for _, place := range versionPlaces() {
		t.Run(place.name, func(t *testing.T) {
			server, client := newTestClient(t)
			clock := newFakeClock()
			data := newAnswers("first")
			reported := newCycles(t)
			place.publish(t, server, "20260912.1")

			cache := place.build(client, data.load,
				versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

			served, err := cache.Get(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "first", served)
			assert.Equal(t, 1, data.timesRead())

			// The data behind the version changed but the version did not: the cache must not see it.
			data.set("second")
			clock.Advance(2 * time.Minute)

			refreshCycle(t, cache, reported)

			served, err = cache.Get(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "first", served)
			assert.Equal(t, 1, data.timesRead(), "an unchanged version must not read the data")
			assert.Equal(t, versionedcache.VersionUnchanged, reported.seen()[1])
		})
	}
}

func TestADifferentVersionReloadsTheData(t *testing.T) {
	// The comparison is exact text and never an ordering, so a version that moves backwards
	// counts as changed exactly as much as one that moves forwards.
	changes := []struct{ name, from, to string }{
		{"a newer version", "20260912.1", "20260912.2"},
		{"an older version restored from a backup", "20260912.9", "20260101.1"},
	}
	for _, place := range versionPlaces() {
		for _, change := range changes {
			t.Run(place.name+", "+change.name, func(t *testing.T) {
				server, client := newTestClient(t)
				clock := newFakeClock()
				data := newAnswers("first")
				reported := newCycles(t)
				place.publish(t, server, change.from)

				cache := place.build(client, data.load,
					versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

				_, err := cache.Get(context.Background())
				require.NoError(t, err)

				data.set("second")
				place.publish(t, server, change.to)
				clock.Advance(2 * time.Minute)

				refreshCycle(t, cache, reported)
				assert.Equal(t, 2, data.timesRead())
				assertServes(t, cache, "second")
			})
		}
	}
}

// Every one of these reads the same as a version nobody publishes, which is the behaviour every
// caller had before this package existed. Deleting the version is an off switch needing no deploy.
func TestWithNothingToCompareTheDataReloadsEveryInterval(t *testing.T) {
	inAHash := func(client *redis.Client, loadData func(context.Context) (string, error),
		options versionedcache.Options) *versionedcache.Cache[string] {
		return versionedcache.New(versionedcache.VersionInHashField(client, exampleVersionHashKey,
			exampleVersionField), loadData, time.Minute, options)
	}
	cases := []struct {
		name  string
		seed  func(*miniredis.Miniredis)
		build func(*redis.Client, func(context.Context) (string, error), versionedcache.Options) *versionedcache.Cache[string]
	}{
		{
			name: "no version key was configured",
			build: func(_ *redis.Client, loadData func(context.Context) (string, error),
				options versionedcache.Options) *versionedcache.Cache[string] {
				return versionedcache.New(nil, loadData, time.Minute, options)
			},
		},
		{
			name: "the key is configured but nobody has published it",
			build: func(client *redis.Client, loadData func(context.Context) (string, error),
				options versionedcache.Options) *versionedcache.Cache[string] {
				return versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute, options)
			},
		},
		{
			name: "the hash carries somebody else's field but not this one",
			seed: func(server *miniredis.Miniredis) {
				server.HSet(exampleVersionHashKey, "someoneElse", "1789380316123456789")
			},
			build: inAHash,
		},
		{
			name:  "the hash is not there at all",
			build: inAHash,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			server, client := newTestClient(t)
			clock := newFakeClock()
			data := newAnswers("first")
			if testCase.seed != nil {
				testCase.seed(server)
			}
			reported := newCycles(t)

			cache := testCase.build(client, data.load,
				versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

			served, err := cache.Get(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "first", served)

			data.set("second")
			clock.Advance(2 * time.Minute)

			refreshCycle(t, cache, reported)
			assert.Equal(t, 2, data.timesRead(), "with no version published the cache reloads every interval")
			assert.Equal(t, []versionedcache.RefreshOutcome{
				versionedcache.ReloadedWithoutVersion, versionedcache.ReloadedWithoutVersion,
			}, reported.seen())
			assertServes(t, cache, "second")
		})
	}
}

func TestDataIsServedFromMemoryInsideTheInterval(t *testing.T) {
	// A cache with no version to check needs no redis at all.
	clock := newFakeClock()
	data := newAnswers("first")

	cache := versionedcache.New(nil, data.load, time.Minute, versionedcache.Options{Now: clock.Now})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	data.set("second")
	clock.Advance(30 * time.Second)

	served, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", served, "inside the interval the cache must not reload")
	assert.Equal(t, 1, data.timesRead())
}

func TestOutcomesTellTheThreeCasesApart(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	data := newAnswers("first")
	reported := newCycles(t)

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), data.load, time.Minute, versionedcache.Options{
		Now:           clock.Now,
		WhenRefreshed: reported.record,
	})

	// No version published yet.
	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	// A version appears.
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))
	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)

	// The same version again.
	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)

	assert.Equal(t, []versionedcache.RefreshOutcome{
		versionedcache.ReloadedWithoutVersion,
		versionedcache.ReloadedAfterChange,
		versionedcache.VersionUnchanged,
	}, reported.seen())
}

func TestFailedVersionReadKeepsTheCachedData(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	data := newAnswers("first")
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), data.load, time.Minute, versionedcache.Options{
		Now:           clock.Now,
		WhenRefreshed: reported.record,
	})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	// Redis goes away.
	server.Close()
	clock.Advance(2 * time.Minute)

	refreshCycle(t, cache, reported)

	served, err := cache.Get(context.Background())
	require.NoError(t, err, "a warm cache keeps serving through a Redis outage")
	assert.Equal(t, "first", served)
	assert.Equal(t, 1, data.timesRead(), "the data must not be read when the version could not be")
	assert.ErrorContains(t, reported.lastError(), exampleVersionKey,
		"the failure is reported even though Get succeeded, and names the key it could not read")
}

func TestFailedDataReadKeepsTheCachedData(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	var failNextLoad atomic.Bool
	loadData := func(context.Context) (string, error) {
		if failNextLoad.Load() {
			return "", errors.New("the hash could not be read")
		}
		return "first", nil
	}

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute, versionedcache.Options{
		Now:           clock.Now,
		WhenRefreshed: reported.record,
	})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	failNextLoad.Store(true)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.2"))
	clock.Advance(2 * time.Minute)

	refreshCycle(t, cache, reported)

	served, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", served)
	assert.Equal(t, versionedcache.RefreshFailed, reported.seen()[1])
	assert.Error(t, reported.lastError())
}

// The version is read before the data, so a redis that is down fails the first load before the
// loader is ever called. The caller gets the error and runs on its own defaults.
func TestAFirstLoadThatCannotReadTheVersionReturnsTheError(t *testing.T) {
	server, client := newTestClient(t)
	reported := newCycles(t)
	var loads atomic.Int64
	loadData := func(context.Context) (string, error) {
		loads.Add(1)
		return "never reached", nil
	}

	server.Close()

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute,
		versionedcache.Options{WhenRefreshed: reported.record})

	data, err := cache.Get(context.Background())
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Equal(t, int64(0), loads.Load(), "the data is not read when the version could not be")
	assert.Equal(t, []versionedcache.RefreshOutcome{versionedcache.FirstLoadFailed}, reported.seen())
}

func TestFirstLoadFailureIsReturnedToTheCaller(t *testing.T) {
	_, client := newTestClient(t)
	loadData := func(context.Context) (string, error) {
		return "", errors.New("the hash could not be read")
	}

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute, versionedcache.Options{})

	data, err := cache.Get(context.Background())
	assert.Error(t, err, "with nothing cached there is nothing to fall back to")
	assert.Empty(t, data)
}

// Nothing has ever loaded, so a caller arriving while the first load is running has no previous
// value to fall back to. It waits for that load rather than being turned away, and however many
// callers arrive the data is read once.
func TestCallersArrivingDuringTheFirstLoadAllGetIt(t *testing.T) {
	server, client := newTestClient(t)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	firstLoadStarted := make(chan struct{})
	releaseFirstLoad := make(chan struct{})
	var loads atomic.Int64
	loadData := func(context.Context) (string, error) {
		if loads.Add(1) == 1 {
			close(firstLoadStarted)
			<-releaseFirstLoad
		}
		return "first", nil
	}

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute, versionedcache.Options{})

	var callers sync.WaitGroup
	for i := 0; i < 4; i++ {
		callers.Add(1)
		go func() {
			defer callers.Done()
			data, err := cache.Get(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "first", data)
		}()
	}

	<-firstLoadStarted
	close(releaseFirstLoad)
	callers.Wait()

	assert.Equal(t, int64(1), loads.Load(), "however many callers arrive, the data is read once")
}

// A failed cycle waits out the interval like a successful one, so an unreachable Redis is tried
// once a cycle rather than on every call.
func TestAFailedCycleRestartsTheInterval(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	data := newAnswers("first")
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), data.load, time.Minute, versionedcache.Options{
		Now:           clock.Now,
		WhenRefreshed: reported.record,
	})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	server.Close()
	clock.Advance(2 * time.Minute)

	refreshCycle(t, cache, reported)
	require.Equal(t, 1, reported.failures())

	served, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", served)

	for i := 0; i < 5; i++ {
		served, err = cache.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "first", served)
	}
	assert.Equal(t, 1, reported.failures(), "calls inside the interval must not reach for redis again")

	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)
	assert.Equal(t, 2, reported.failures(), "the next interval tries again")
	assert.Equal(t, 1, data.timesRead(), "the data was never re-read while the version could not be")
}

// A failed cycle is put back on the clock, so calls inside the interval start no further refresh.
// The store's own calculator leaves a failed reload still due, which would send every call that
// follows straight back to a source already known to be down. The loader here fails at once, so an
// extra cycle would report well inside the window below.
func TestAFailedCycleIsPutBackOnTheClock(t *testing.T) {
	clock := newFakeClock()
	reported := newCycles(t)
	var failNow atomic.Bool
	loadData := func(context.Context) (string, error) {
		if failNow.Load() {
			return "", errors.New("the dataset could not be read")
		}
		return "first", nil
	}

	cache := versionedcache.New(nil, loadData, time.Minute,
		versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	failNow.Store(true)
	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)
	require.Equal(t, 1, reported.failures())

	// The clock does not move again, so every call below is inside the interval the failed cycle
	// restarted. None of them may start a cycle of its own.
	assert.Never(t, func() bool {
		served, getErr := cache.Get(context.Background())
		assert.NoError(t, getErr)
		assert.Equal(t, "first", served)
		return reported.failures() > 1
	}, 300*time.Millisecond, 10*time.Millisecond, "a failed cycle must wait out the interval like a good one")
}

// The cold case is the exception: with nothing to serve there is no interval to wait out.
func TestACacheThatHasNeverLoadedRetriesEveryCall(t *testing.T) {
	_, client := newTestClient(t)
	clock := newFakeClock()
	var attempts atomic.Int64
	loadData := func(context.Context) (string, error) {
		attempts.Add(1)
		return "", errors.New("the hash could not be read")
	}

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute,
		versionedcache.Options{Now: clock.Now})

	// The clock never moves, so an interval that applied here would stop the later calls.
	for call := 1; call <= 3; call++ {
		_, err := cache.Get(context.Background())
		require.Error(t, err)
		assert.Equal(t, int64(call), attempts.Load())
	}
}

// A cycle that fails must leave the newest value in place rather than writing back whatever the
// cache was holding when that cycle started.
func TestAFailedCycleLeavesTheNewestValueInPlace(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	data := newAnswers("v1data")
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "v1"))

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), data.load, time.Minute, versionedcache.Options{
		Now:           clock.Now,
		WhenRefreshed: reported.record,
	})

	warm, err := cache.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "v1data", warm)

	data.set("v2data")
	require.NoError(t, server.Set(exampleVersionKey, "v2"))
	clock.Advance(2 * time.Minute)

	refreshCycle(t, cache, reported)
	assertServes(t, cache, "v2data")

	// Redis goes away, so the next cycle fails with v2 already held.
	server.Close()
	clock.Advance(2 * time.Minute)

	refreshCycle(t, cache, reported)
	require.Equal(t, 1, reported.failures())

	served, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v2data", served, "a failed cycle must not write an older snapshot back")
}

func TestConcurrentGetsCauseOneDataRead(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	var loads atomic.Int64
	refreshIsRunning := make(chan struct{})
	releaseRefresh := make(chan struct{})
	loadData := func(context.Context) (string, error) {
		if loads.Add(1) == 2 {
			// Hold the refresh open so the other callers arrive while it is still running.
			close(refreshIsRunning)
			<-releaseRefresh
			return "second", nil
		}
		return "first", nil
	}

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute, versionedcache.Options{
		Now:           clock.Now,
		WhenRefreshed: reported.record,
	})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	require.NoError(t, server.Set(exampleVersionKey, "20260912.2"))
	clock.Advance(2 * time.Minute)

	// The call that trips the interval starts the refresh and is answered from memory.
	served, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", served)
	<-refreshIsRunning

	for i := 0; i < 4; i++ {
		served, err = cache.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "first", served, "callers arriving mid-refresh are served the held value")
	}

	close(releaseRefresh)
	reported.waitUntil(2)
	assert.Equal(t, int64(2), loads.Load(), "one warm-up load plus one refresh, not six")
	assertServes(t, cache, "second")
}

// the whole point of one hash for every dataset: a version moving for somebody else's namespace
// must not send this reader off to read its own dataset again
func TestAVersionMovingInAnotherFieldIsIgnored(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	data := newAnswers("first")
	reported := newCycles(t)
	server.HSet(exampleVersionHashKey, exampleVersionField, "1789380316123456789")

	cache := versionedcache.New(
		versionedcache.VersionInHashField(client, exampleVersionHashKey, exampleVersionField),
		data.load, time.Minute, versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	server.HSet(exampleVersionHashKey, "someoneElse", "1789380341772904118")
	clock.Advance(2 * time.Minute)

	refreshCycle(t, cache, reported)
	assert.Equal(t, 1, data.timesRead())
}

// A version that stops moving is not proof the data is unchanged: a writer can change rows and die
// before publishing. So data kept on an unchanged version for MaxStaleness is read again anyway, and
// only a read restarts that clock, never a check that skipped.
func TestDataKeptPastMaxStalenessIsReadAgainOnAnUnchangedVersion(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	data := newAnswers("first")
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), data.load, time.Minute,
		versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record, MaxStaleness: 10 * time.Minute})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	// The rows change under a version that never moves.
	data.set("second")

	clock.Advance(6 * time.Minute)
	refreshCycle(t, cache, reported)
	assert.Equal(t, 1, data.timesRead(), "inside the limit an unchanged version still skips the read")

	// Eleven minutes since the read, five since the check that skipped: only the read counts.
	clock.Advance(5 * time.Minute)
	refreshCycle(t, cache, reported)
	assert.Equal(t, 2, data.timesRead(), "past the limit the data is read though the version has not moved")
	assert.Equal(t, versionedcache.ReloadedAtMaxStaleness, reported.seen()[2])
	assertServes(t, cache, "second")

	// That read restarted the clock, so the next cycle skips again.
	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)
	assert.Equal(t, 2, data.timesRead(), "a read restarts the staleness clock")
	assert.Equal(t, versionedcache.VersionUnchanged, reported.seen()[3])
}

func TestMaxStalenessDefaultsToSixHours(t *testing.T) {
	assert.Equal(t, 6*time.Hour, versionedcache.DefaultMaxStaleness)

	server, client := newTestClient(t)
	clock := newFakeClock()
	data := newAnswers("first")
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), data.load, time.Minute,
		versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	clock.Advance(6*time.Hour - time.Minute)
	refreshCycle(t, cache, reported)
	assert.Equal(t, 1, data.timesRead(), "a minute short of six hours is still inside the default")

	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)
	assert.Equal(t, 2, data.timesRead(), "six hours on an unchanged version reads the data again")
	assert.Equal(t, versionedcache.ReloadedAtMaxStaleness, reported.seen()[2])
}

// A panic in the loader has to become a failed cycle. otter runs a background reload on a
// goroutine it starts itself and re-panics whatever the loader panicked with, so without a
// recover the process ends on the refresh timer instead of reporting one bad cycle.
func TestALoaderThatPanicsIsReportedAsAFailedCycle(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "v1"))

	var reads atomic.Int64
	loadData := func(context.Context) (string, error) {
		if reads.Add(1) > 1 {
			panic("boom from loadData")
		}
		return "first", nil
	}

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), loadData, time.Minute,
		versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	require.NoError(t, server.Set(exampleVersionKey, "v2"))
	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)

	assert.Equal(t, versionedcache.RefreshFailed, reported.seen()[1])
	require.Error(t, reported.lastError())
	assert.Contains(t, reported.lastError().Error(), "loadData panicked")

	served, err := cache.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "first", served, "a panicking reload keeps the data the cache already held")
}

// Same for the version read, which runs on the same background goroutine.
func TestAVersionReadThatPanicsIsReportedAsAFailedCycle(t *testing.T) {
	clock := newFakeClock()
	reported := newCycles(t)

	var checks atomic.Int64
	readVersion := func(context.Context) (string, bool, error) {
		if checks.Add(1) > 1 {
			panic("boom from ReadVersion")
		}
		return "v1", true, nil
	}

	cache := versionedcache.New(readVersion, newAnswers("first").load, time.Minute,
		versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)

	assert.Equal(t, versionedcache.RefreshFailed, reported.seen()[1])
	require.Error(t, reported.lastError())
	assert.Contains(t, reported.lastError().Error(), "ReadVersion panicked")
}

// Each of these builds a cache that compiles, loads once and is quietly wrong for the rest of
// the process. New has no error return, so refusing them has to happen before the first tag.
func TestNewRejectsSettingsThatWouldNeverRefresh(t *testing.T) {
	_, client := newTestClient(t)
	readVersion := versionedcache.VersionInKey(client, exampleVersionKey)
	loadData := newAnswers("first").load

	cases := []struct {
		name  string
		build func()
	}{
		{"an interval of zero", func() {
			versionedcache.New(readVersion, loadData, 0, versionedcache.Options{})
		}},
		{"a negative interval", func() {
			versionedcache.New(readVersion, loadData, -time.Second, versionedcache.Options{})
		}},
		{"a negative staleness limit", func() {
			versionedcache.New(readVersion, loadData, time.Minute, versionedcache.Options{MaxStaleness: -time.Second})
		}},
		{"a staleness limit shorter than the interval", func() {
			versionedcache.New(readVersion, loadData, time.Minute, versionedcache.Options{MaxStaleness: 30 * time.Second})
		}},
		{"an interval longer than the staleness limit a zero takes", func() {
			versionedcache.New(readVersion, loadData, versionedcache.DefaultMaxStaleness+time.Minute, versionedcache.Options{})
		}},
		{"no loader", func() {
			versionedcache.New[string](readVersion, nil, time.Minute, versionedcache.Options{})
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Panics(t, c.build)
		})
	}
}

// A consumer tags its metric with string(outcome), so these six literals are the contract and
// renaming one silently renames someone's dashboard series. Nothing else in the suite asserts a
// literal value, so they are pinned here.
func TestTheOutcomeStringsAreTheMetricContract(t *testing.T) {
	assert.Equal(t, "version_unchanged", string(versionedcache.VersionUnchanged))
	assert.Equal(t, "reloaded_after_change", string(versionedcache.ReloadedAfterChange))
	assert.Equal(t, "reloaded_without_version", string(versionedcache.ReloadedWithoutVersion))
	assert.Equal(t, "reloaded_at_max_staleness", string(versionedcache.ReloadedAtMaxStaleness))
	assert.Equal(t, "refresh_failed", string(versionedcache.RefreshFailed))
	assert.Equal(t, "first_load_failed", string(versionedcache.FirstLoadFailed))
}

// The empty string is what a missing key reads as, and it is also a value a publisher can write.
// An entry loaded while nothing was published must not compare equal to a published empty one.
func TestAPublishedEmptyVersionIsNotMistakenForNothingPublished(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	data := newAnswers("first")
	reported := newCycles(t)

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey), data.load, time.Minute,
		versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, versionedcache.ReloadedWithoutVersion, reported.seen()[0])

	// The publisher starts publishing, with an empty value, and changes the data with it.
	data.set("second")
	require.NoError(t, server.Set(exampleVersionKey, ""))
	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)

	assert.Equal(t, 2, data.timesRead(), "an empty published version is a version, not the absence of one")
	assertServes(t, cache, "second")
}

// A first load with a version already published has nothing to compare against, so it counts as
// a change. This is a live dashboard tag value, so the outcome is asserted rather than assumed.
func TestAFirstLoadWithAVersionPublishedReportsReloadedAfterChange(t *testing.T) {
	server, client := newTestClient(t)
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "v1"))

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey),
		newAnswers("first").load, time.Minute, versionedcache.Options{WhenRefreshed: reported.record})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	require.Len(t, reported.seen(), 1)
	assert.Equal(t, versionedcache.ReloadedAfterChange, reported.seen()[0])
}

// A caller's reporter is the third callback that runs on the store's refresh goroutine, so a
// panic in it would end the process the same way a panicking loader would.
func TestAReporterThatPanicsDoesNotEndTheProcess(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	var reports atomic.Int64
	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey),
		newAnswers("first").load, time.Minute,
		versionedcache.Options{Now: clock.Now, WhenRefreshed: func(versionedcache.RefreshOutcome, error) {
			if reports.Add(1) == 2 {
				panic("the caller's reporter panicked")
			}
		}})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	clock.Advance(2 * time.Minute)
	_, err = cache.Get(context.Background())
	require.NoError(t, err)
	require.Eventually(t, func() bool { return reports.Load() >= 2 },
		5*time.Second, 5*time.Millisecond, "waited for the cycle that panics")

	clock.Advance(2 * time.Minute)
	_, err = cache.Get(context.Background())
	require.NoError(t, err)
	assert.Eventually(t, func() bool { return reports.Load() >= 3 },
		5*time.Second, 5*time.Millisecond, "the cycle after the panicking one must still run")
}

// The interval is what keeps a source that is down from being read on every call. The store puts
// the data back on the clock only once a cycle finishes, so without a claim of our own every call
// arriving while a slow failure is in flight starts a cycle of its own.
func TestAFailingSourceIsReadOncePerIntervalNotOncePerGet(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	var versionReads atomic.Int64
	var failing atomic.Bool
	published := versionedcache.VersionInKey(client, exampleVersionKey)
	counted := func(ctx context.Context) (string, bool, error) {
		versionReads.Add(1)
		if failing.Load() {
			// A real outage times out rather than answering at once, and the window that failure
			// leaves open is the whole point of this test.
			time.Sleep(2 * time.Millisecond)
			return "", false, errors.New("the source is down")
		}
		return published(ctx)
	}

	cache := versionedcache.New(counted, newAnswers("first").load, time.Minute,
		versionedcache.Options{Now: clock.Now})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	failing.Store(true)
	readsBefore := versionReads.Load()
	clock.Advance(2 * time.Minute)

	// Live traffic arriving while the source is down, with the clock held still. It has to keep
	// arriving: the calls that pile up behind one failing read are folded into it, and the cycle
	// that should not happen is the one a later call starts once that read has failed.
	stop := make(chan struct{})
	var callers sync.WaitGroup
	for range 50 {
		callers.Add(1)
		go func() {
			defer callers.Done()
			for {
				select {
				case <-stop:
					return
				default:
					served, err := cache.Get(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, "first", served)
				}
			}
		}()
	}
	time.Sleep(300 * time.Millisecond)
	close(stop)
	callers.Wait()

	assert.Equal(t, int64(1), versionReads.Load()-readsBefore,
		"one interval elapsed, so the source is read once however long the traffic keeps arriving")
}

// otter reads its own ErrNotFound as the dataset being gone and drops what it holds. A loader that
// wraps that sentinel would therefore empty a warm cache on one failed read.
func TestAFailedReadReportingNotFoundKeepsTheWarmData(t *testing.T) {
	server, client := newTestClient(t)
	clock := newFakeClock()
	reported := newCycles(t)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.1"))

	var failing atomic.Bool
	loadData := func(context.Context) (string, error) {
		if failing.Load() {
			return "", fmt.Errorf("read the dataset: %w", otter.ErrNotFound)
		}
		return "first", nil
	}

	cache := versionedcache.New(versionedcache.VersionInKey(client, exampleVersionKey),
		loadData, time.Minute, versionedcache.Options{Now: clock.Now, WhenRefreshed: reported.record})

	_, err := cache.Get(context.Background())
	require.NoError(t, err)

	failing.Store(true)
	require.NoError(t, server.Set(exampleVersionKey, "20260912.2"))
	clock.Advance(2 * time.Minute)
	refreshCycle(t, cache, reported)

	assert.Equal(t, versionedcache.RefreshFailed, reported.seen()[len(reported.seen())-1])
	assertServes(t, cache, "first")
}
