package snowflake

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/common"
	"github.com/wego/pkg/host"
)

var (
	nodeID = uint16(24)
)

func customProvider() (uint16, error) {
	return nodeID, nil
}

func customProviderPlus() (uint16, error) {
	return nodeID + 1, nil
}

func Test_SetEpoch_Zero(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{}
	err := g.setEpoch(time.Time{})
	assert.Error(err)
	assert.Contains(err.Error(), "Epoch cannot be a zero value")
}

func Test_SetEpoch_InTheFuture(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{}
	now := common.CurrentUTCTime()
	err := g.setEpoch(now.Add(1 * time.Hour))
	assert.Error(err)
	assert.Contains(err.Error(), "Epoch cannot be in the future")
}

func Test_SetEpoch_MaxTimeExceeded(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{}
	err := g.setEpoch(time.Date(1751, 1, 1, 1, 0, 0, 0, time.UTC))
	assert.Error(err)
	assert.Contains(err.Error(), "The maximum life cycle of the snowflake algorithm is 179 years")
}

func Test_SetResolver_Nil(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{}
	err := g.setSequenceGenerator(nil)
	assert.Error(err)
}

func Test_SetResolver_Ok(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{}
	err := g.setSequenceGenerator(AtomicGenerator)
	assert.NoError(err)
	err = g.setSequenceGenerator(AtomicGenerator)
	assert.NoError(err)
}

func Test_SetNodeIDProvider_Nil(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{
		Settings: Settings{
			Epoch:             time.Now(),
			SequenceGenerator: AtomicGenerator,
		},
	}
	err := g.setNodeIDProvider(nil)
	assert.Error(err)
}

func Test_SetNodeIDProvider_Error(t *testing.T) {
	assert := assert.New(t)
	errorMsg := "some error"
	g := &Generator{
		Settings: Settings{
			Epoch:             time.Now(),
			SequenceGenerator: AtomicGenerator,
		},
	}
	err := g.setNodeIDProvider(func() (uint16, error) {
		return nodeID, errors.New(errorMsg)
	})
	assert.Error(err)
	assert.Contains(err.Error(), errorMsg)
}

func Test_SetNodeIDProvider_Ok(t *testing.T) {
	g := &Generator{
		Settings: Settings{
			Epoch:             time.Now(),
			SequenceGenerator: AtomicGenerator,
		},
	}

	assert := assert.New(t)
	err := g.setNodeIDProvider(customProvider)
	assert.NoError(err)

	generated, err := g.NextID()
	assert.NoError(err)
	decomposed := g.Decompose(generated)
	assert.Equal(nodeID, decomposed.NodeID)

	err = g.setNodeIDProvider(customProviderPlus)
	assert.NoError(err)
	generated, err = g.NextID()
	assert.NoError(err)
	decomposed = g.Decompose(generated)
	assert.Equal(nodeID, decomposed.NodeID)
}

func Test_CurrentTimestamp_Ok(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{
		Settings: Settings{
			Epoch:             time.Now().Add(-24 * time.Hour),
			SequenceGenerator: AtomicGenerator,
		},
	}
	v, e := g.currentTimestamp()
	assert.GreaterOrEqual(v, int64(0))
	assert.NoError(e)
}

func Test_CurrentTimestamp_ExceedsMaxTime(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{
		Settings: Settings{
			Epoch:             time.Date(1751, 1, 1, 1, 0, 0, 0, time.UTC),
			SequenceGenerator: AtomicGenerator,
		},
	}
	v, e := g.currentTimestamp()
	assert.GreaterOrEqual(v, int64(0))
	assert.Error(e)
	assert.Contains(e.Error(), "timestamp exceeds max time(2^39-1 * 10ms), please check the epoch settings")
}

func Test_CurrentTimestamp_EpochInTheFuture(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{
		Settings: Settings{
			Epoch:             time.Now().Add(24 * time.Hour),
			SequenceGenerator: AtomicGenerator,
		},
	}
	v, e := g.currentTimestamp()
	assert.Less(v, int64(0))
	assert.Error(e)
	assert.Contains(e.Error(), "current time can not be negative, please make sure the epoch is not in the future")
}

func Test_Init_EpochError(t *testing.T) {
	assert := assert.New(t)
	s := Settings{
		Epoch:             time.Now().Add(24 * time.Hour),
		SequenceGenerator: AtomicGenerator,
	}
	g := &Generator{}
	err := g.init(s)
	assert.Error(err)
	assert.Contains(err.Error(), "Epoch cannot be in the future")
}

func Test_Init_SequenceGeneratorError(t *testing.T) {
	assert := assert.New(t)
	s := Settings{
		Epoch:             time.Now().Add(-24 * time.Hour),
		SequenceGenerator: nil,
	}
	g := &Generator{}
	err := g.init(s)
	assert.Error(err)
	assert.Contains(err.Error(), "SequenceGenerator cannot be nil")
}

func Test_Init_NodeIDProviderError(t *testing.T) {
	assert := assert.New(t)
	s := Settings{
		Epoch:             time.Now().Add(-24 * time.Hour),
		SequenceGenerator: AtomicGenerator,
		NodeIDProvider:    nil,
	}
	g := &Generator{}
	err := g.init(s)
	assert.Error(err)
	assert.Contains(err.Error(), "NodeIDProvider cannot be nil")
}

func newID(t *testing.T) uint64 {
	assert := assert.New(t)

	id, err := NextID()
	if err != nil {
		t.Fatal("ID not generated")
	}
	assert.NotZero(id)
	return id
}

func currentTime() time.Time {
	return common.CurrentUTCTime()
}

func Test_NextIDFor10Sec(t *testing.T) {
	assert := assert.New(t)
	var numID uint32
	var lastID uint64
	var maxSequence uint16
	nodeID, err := host.Lower16BitPrivateIP()
	assert.NoError(err)

	initial := currentTime()
	current := initial
	for current.Sub(initial) < 10*time.Second {
		id := newID(t)
		parts := Decompose(id)
		numID++

		if id <= lastID {
			t.Fatal("duplicated ID")
		}
		lastID = id

		current = currentTime()

		actualTime := parts.Timestamp
		overtime := actualTime.Sub(current)
		if overtime > 0 {
			t.Errorf("unexpected overtime: %d", overtime)
		}

		actualSequence := parts.Sequence
		if maxSequence < actualSequence {
			maxSequence = actualSequence
		}

		actualNodeID := parts.NodeID
		if actualNodeID != nodeID {
			t.Errorf("unexpected nodeID: %d", actualNodeID)
		}
	}

	if maxSequence > 1<<sequenceBits-1 {
		t.Errorf("unexpected max sequence: %d", maxSequence)
	}
}

func Test_NewIDInParallel(t *testing.T) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	const numID = 10000
	numGenerator := 1
	if numCPU/2 > numGenerator {
		numGenerator = numCPU / 2
	}

	generate := func(wg *sync.WaitGroup, partition []uint64) {
		defer wg.Done()
		start := time.Now()
		for i := 0; i < numID; i++ {
			partition[i] = NewID()

		}
		elapsed := time.Since(start)
		_, _ = fmt.Printf("%v IDs generaged, tooks %v s\n", numID, elapsed)
	}
	wg := sync.WaitGroup{}
	start := time.Now()
	ids := make(map[int][]uint64, numGenerator)
	for i := 0; i < numGenerator; i++ {
		ids[i] = make([]uint64, numID)
		wg.Add(1)
		go generate(&wg, ids[i])
	}
	wg.Wait()
	elapsed := time.Since(start)
	_, _ = fmt.Printf("%v IDs generaged, tooks %v, speed: %v IDs/sec\n", numID*numGenerator,
		elapsed, float64(numID*numGenerator)/elapsed.Seconds())

	set := make(map[uint64]int)
	for i, partition := range ids {
		for j := 0; j < numID; j++ {
			id := partition[j]
			if actualIdx, ok := set[id]; ok {
				decomposed := Decompose(id)
				t.Fatalf("duplicated ID: ids[%v][%v]=%v(%v = %v), i = %v", i, j, id,
					decomposed.Timestamp, decomposed.Sequence, actualIdx)
			}
			set[id] = i*numID + j
		}

	}
}

func BenchmarkNewParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewID()
		}
	})
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewID()
	}
}
