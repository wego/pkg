package rand_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestInt63(t *testing.T) {
	total := 100 * 10000
	unique := make(map[int64]bool, total)
	for i := 0; i < total; i++ {
		v := rand.Int63()
		_, ok := unique[v]
		assert.False(t, ok)
		unique[v] = true
	}
}

func TestInt64Parallel(t *testing.T) {
	total := 100 * 10000
	concurrent := runtime.NumCPU()
	if concurrent < 4 {
		concurrent = 4
	}
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
	total := 100 * 10000
	unique := make(map[uint64]bool, total)
	for i := 0; i < total; i++ {
		v := rand.Uint64()
		_, ok := unique[v]
		assert.False(t, ok)
		unique[v] = true
	}
}

func TestUint64Parallel(t *testing.T) {
	total := 100 * 10000
	concurrent := runtime.NumCPU()
	if concurrent < 4 {
		concurrent = 4
	}
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
	total := 100 * 10000
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
	total := 100 * 10000
	concurrent := runtime.NumCPU()
	if concurrent < 4 {
		concurrent = 4
	}
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

func TestStringWithOptionInvalidOption(t *testing.T) {
	str, err := rand.StringWithOption(10, -1)
	assert.Empty(t, str)
	assert.Contains(t, err.Error(), "option -1 is not supported")
}

func TestStringWithOption(t *testing.T) {
	total := 100 * 10000
	uniqueMap := make(map[string]bool, total)
	for i := 0; i < total; i++ {
		s, _ := rand.StringWithOption(10, rand.Numbers|rand.Lower|rand.Letters)
		if _, ok := uniqueMap[s]; ok {
			t.Errorf("String(10, Numbers |Lower | Letters) produced a duplicate: %s", s)
		}
		uniqueMap[s] = true
	}
}

func TestStringWithOptionParallel(t *testing.T) {
	total := 100 * 10000
	concurrent := runtime.NumCPU()
	if concurrent < 10 {
		concurrent = 10
	}
	var wg sync.WaitGroup
	wg.Add(concurrent)
	var uniqueMap sync.Map
	for i := 0; i < concurrent; i++ {
		go func() {
			for i := 0; i < total; i++ {
				s, _ := rand.StringWithOption(10, rand.Numbers|rand.Lower|rand.Letters)
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
		_, _ = rand.StringWithOption(10, rand.Numbers|rand.Letters|rand.Upper)
	}
}

func BenchmarkStringWithOptionParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = rand.StringWithOption(10, rand.Numbers|rand.Letters|rand.Upper)
		}
	})
}
