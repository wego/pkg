package snowflake

import (
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/common"
	"github.com/wego/pkg/host"
	"runtime"
	"testing"
	"time"
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
	err := g.setResolver(nil)
	assert.Error(err)
}

func Test_SetResolver_Ok(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{}
	err := g.setResolver(AtomicResolver)
	assert.NoError(err)
	err = g.setResolver(AtomicResolver)
	assert.NoError(err)
}

func Test_SetNodeIDProvider_Nil(t *testing.T) {
	assert := assert.New(t)
	g := &Generator{
		Settings: Settings{
			Epoch:            time.Now(),
			SequenceResolver: AtomicResolver,
		},
	}
	err := g.setNodeIDProvider(nil)
	assert.Error(err)
}

func Test_SetNodeIDProvider_Ok(t *testing.T) {
	g := &Generator{
		Settings: Settings{
			Epoch:            time.Now(),
			SequenceResolver: AtomicResolver,
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
	sequenceBits := 8
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

		actualMSB := parts.MSB
		if actualMSB != 0 {
			t.Errorf("unexpected msb: %d", actualMSB)
		}

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
	consumer := make(chan uint64, numID)

	generate := func() {
		for i := 0; i < numID; i++ {
			consumer <- newID(t)
		}
	}

	const numGenerator = 10
	for i := 0; i < numGenerator; i++ {
		go generate()
	}

	set := make(map[uint64]bool)
	for i := 0; i < numID*numGenerator; i++ {
		id := <-consumer
		if set[id] {
			decomposed := Decompose(id)
			t.Fatalf("duplicated ID: %v = %v", decomposed.Timestamp, decomposed.Sequence)
		}
		set[id] = true
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
