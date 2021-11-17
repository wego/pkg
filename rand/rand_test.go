package rand_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

var (
	total = 10 * 10000
)

func TestInt63(t *testing.T) {
	unique := make(map[int64]bool, total)
	for i := 0; i < total; i++ {
		v := rand.Int63()
		_, ok := unique[v]
		assert.False(t, ok)
		unique[v] = true
	}
}

func TestInt64Parallel(t *testing.T) {
	concurrent := runtime.NumCPU()
	if concurrent < 4 {
		concurrent = 4
	}
	t.Logf("running test in parallel with %v goroutines", concurrent)
	var uniqueMap sync.Map
	var wg sync.WaitGroup
	wg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			for i := 0; i < total; i++ {
				v := rand.Int63()
				_, ok := uniqueMap.Load(v)
				assert.False(t, ok)
				uniqueMap.Store(v, true)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestUint64(t *testing.T) {
	unique := make(map[uint64]bool, total)
	for i := 0; i < total; i++ {
		v := rand.Uint64()
		_, ok := unique[v]
		assert.False(t, ok)
		unique[v] = true
	}
}

func TestUint64Parallel(t *testing.T) {
	concurrent := runtime.NumCPU()
	if concurrent < 4 {
		concurrent = 4
	}
	t.Logf("running test in parallel with %v goroutines", concurrent)
	var uniqueMap sync.Map
	var wg sync.WaitGroup
	wg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			for i := 0; i < total; i++ {
				v := rand.Uint64()
				_, ok := uniqueMap.Load(v)
				assert.False(t, ok)
				uniqueMap.Store(v, true)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestString(t *testing.T) {
	uniqueMap := make(map[string]bool, total)
	for i := 0; i < total; i++ {
		s := rand.String(10)
		if _, ok := uniqueMap[s]; ok {
			t.Errorf("String(10) produced a duplicate: %s", s)
		}
		uniqueMap[s] = true
	}
}

func TestStringFor15s(t *testing.T) {
	start := time.Now()
	var uniqueMap sync.Map
	total, duplicates := 0, 0
	for time.Now().Sub(start) < 15*time.Second {
		for i := 0; i < 10000; i++ {
			s := rand.String(10)
			if _, ok := uniqueMap.Load(s); ok {
				duplicates++
			}
			total++
			uniqueMap.Store(s, true)
		}
	}
	t.Logf("Generated %d strings, %d duplicates, duplicate ratio = %f", total, duplicates, float64(duplicates)/float64(total))
}

func TestStringParallel(t *testing.T) {
	concurrent := runtime.NumCPU()
	if concurrent < 4 {
		concurrent = 4
	}
	t.Logf("running test in parallel with %v goroutines", concurrent)
	var wg sync.WaitGroup
	wg.Add(concurrent)
	var uniqueMap sync.Map
	for i := 0; i < concurrent; i++ {
		go func() {
			for i := 0; i < total; i++ {
				s := rand.String(10)
				if _, ok := uniqueMap.Load(s); ok {
					t.Errorf("String(10) produced a duplicate: %s", s)
				}
				uniqueMap.Store(s, true)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestStringWithPrefixAndSuffix(t *testing.T) {
	s := rand.StringWithOption(10, rand.Numbers|rand.Letters, "prefix", "suffix")
	t.Logf("%s", s)
	assert.Equal(t, 22, len(s))
	assert.Equal(t, "prefix", s[:6])
	assert.Equal(t, "suffix", s[len(s)-6:])
}

func TestStringWithOption(t *testing.T) {
	uniqueMap := make(map[string]bool, total)
	for i := 0; i < total; i++ {
		s := rand.StringWithOption(10, rand.Numbers|rand.Letters|rand.Upper, "prefix", "suffix")
		if _, ok := uniqueMap[s]; ok {
			t.Errorf("String(10, Numbers |Lower | Letters) produced a duplicate: %s", s)
		}
		uniqueMap[s] = true
	}
}

func TestStringWithOptionParallel(t *testing.T) {
	concurrent := runtime.NumCPU()
	if concurrent < 4 {
		concurrent = 4
	}
	t.Logf("running test in parallel with %v goroutines", concurrent)
	var wg sync.WaitGroup
	wg.Add(concurrent)
	var uniqueMap sync.Map
	for i := 0; i < concurrent; i++ {
		go func() {
			for i := 0; i < total; i++ {
				s := rand.StringWithOption(10, rand.Numbers|rand.Letters|rand.Upper, "prefix", "suffix")
				if _, ok := uniqueMap.Load(s); ok {
					t.Errorf("String(10, Numbers |Lower | Letters) produced a duplicate: %s", s)
				}
				uniqueMap.Store(s, true)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rand.String(10)
	}
}

func BenchmarkStringParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rand.String(10)
		}
	})
}

func BenchmarkStringWithOption(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rand.StringWithOption(10, rand.Numbers|rand.Letters|rand.Upper, "prefix", "suffix")
	}
}

func BenchmarkStringWithOptionParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rand.StringWithOption(10, rand.Numbers|rand.Letters|rand.Upper, "prefix", "suffix")
		}
	})
}

func TestCheckOption(t *testing.T) {
	err := rand.CheckOption(rand.Numbers, 2, 30)
	assert.Nil(t, err)
	err = rand.CheckOption(-1, 2, 30)
	assert.Contains(t, err.Error(), "invalid option: -1")
	err = rand.CheckOption(rand.Numbers, 1, 11)
	assert.Contains(t, err.Error(), "can not generate 11 Numbers codes with length 1, minimal length should be 2")
	err = rand.CheckOption(rand.Letters, 1, 30)
	assert.Contains(t, err.Error(), "can not generate 30 Letters codes with length 1, minimal length should be 2")
	err = rand.CheckOption(rand.Numbers|rand.Letters, 2, 5000)
	assert.Contains(t, err.Error(), "can not generate 5000 NumbersAndLetters codes with length 2, minimal length should be 3")
	err = rand.CheckOption(rand.Numbers|rand.Upper, 2, 1297)
	assert.Contains(t, err.Error(), "can not generate 1297 NumbersAndUpperLetters codes with length 2, minimal length should be 3")
}
