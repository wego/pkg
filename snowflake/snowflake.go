package snowflake

import (
	"sync"
	"time"

	"github.com/wego/pkg/errors"
)

// A Snowflake ID is composed of
//
//	39 bits for time in units of 10 msec (about 174 years)
//	16 bits for a node ID
//	 9 bits for a sequence number
const (
	timestampBits = 39 // nodeBits holds the number of bits to use for timestamp (in units of 10 msec)
	nodeBits      = 16 // nodeBits holds the number of bits to use for Node
	sequenceBits  = 9  // sequenceBits holds the number of bits to use for sequence
	maxTimestamp  = 1<<timestampBits - 1
	maxSequence   = 1<<sequenceBits - 1
	timeShift     = nodeBits + sequenceBits
	nodeShift     = sequenceBits
	timeUnit      = 10 // milliseconds, 10 msec

	maskSequence = uint64(1<<sequenceBits - 1)
	maskNodeID   = uint64((1<<nodeBits - 1) << nodeShift)
)

// Settings snowflake generate settings
type Settings struct {
	// Epoch base time for the timestamp Ref: https://en.wikipedia.org/wiki/Epoch_(computing)
	Epoch             time.Time
	SequenceGenerator SequenceGenerator
	NodeIDProvider    NodeIDProvider
}

// Generator a snowflakeID generator
type Generator struct {
	Settings
	epochGuard          sync.Once
	generatorGuard      sync.Once
	nodeIDProviderGuard sync.Once
	nodeID              uint16
}

var (
	defaultSettings = Settings{
		Epoch:             time.Date(2020, 0, 0, 0, 0, 0, 0, time.UTC),
		SequenceGenerator: AtomicGenerator,
		NodeIDProvider:    privateIP,
	}
	defaultGenerator *Generator
	defaultGuard     sync.Once
)

func stockGenerator() *Generator {
	defaultGuard.Do(func() {
		var err error
		defaultGenerator, err = NewGenerator(defaultSettings)
		if err != nil {
			panic(err)
		}
	})
	return defaultGenerator
}

// NewGenerator create a generator with custom settings
func NewGenerator(settings Settings) (*Generator, error) {
	var g Generator
	err := g.init(settings)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// NewID generate a unique snowflake ID with default settings, use lower 16 bits of current private IP as nodeID
// which is out-of-box for AWS private networks and also containers, will ignore the errors when generating
// unique ID
func NewID() uint64 {
	sid, _ := NextID()
	return sid
}

// NextID generate a unique snowflake ID with default settings, use lower 16 bits of current private IP as nodeID
// which is out-of-box for AWS private networks and also containers, return the errors when generating the unique ID
func NextID() (uint64, error) {
	return stockGenerator().NextID()
}

// Decompose returns a struct of snowflake ID with the default generator
func Decompose(id uint64) ID {
	return stockGenerator().Decompose(id)
}

// SequenceGenerator the snowflake sequence generator.
// When use the snowflake algorithm to generate unique ID, make sure:
//
//	The sequence-number generated in the same 10 milliseconds of the same node is unique.
//
// Based on this, we create this interface provides following Generator:
//
//	AtomicGenerator : base sync/atomic (by default).
type SequenceGenerator func(current int64) (uint16, error)

// NodeIDProvider the snowflake Node Generator provider.
type NodeIDProvider func() (uint16, error)

// setEpoch set the start time for snowflake algorithm.
// It will panic when:
//   - t IsZero
//   - t > current millisecond
//   - current millisecond - t > 2^39( * 10ms).
//
// NOTE: This function is thread-unsafe, call before init
func (g *Generator) setEpoch(epoch time.Time) (err error) {
	const op errors.Op = "snowflakeGenerator.setEpoch"
	epoch = epoch.UTC()

	if epoch.IsZero() {
		err = errors.New(nil, op, "Epoch cannot be a zero value")
		return
	}

	if epoch.After(time.Now()) {
		err = errors.New(nil, op, "Epoch cannot be in the future")
		return
	}

	// Because s must after now, so the `df` not < 0.
	if since(epoch) > maxTimestamp {
		err = errors.New(nil, op, "The maximum life cycle of the snowflake algorithm is 179 years(2^39-10ms)")
		return
	}
	g.epochGuard.Do(func() {
		g.Epoch = epoch
	})
	return
}

// setSequenceGenerator set the custom sequence generator
func (g *Generator) setSequenceGenerator(sequenceGenerator SequenceGenerator) (err error) {
	const op errors.Op = "snowflakeGenerator.setSequenceGenerator"
	if sequenceGenerator == nil {
		err = errors.New(nil, op, "SequenceGenerator cannot be nil")
		return
	}
	g.generatorGuard.Do(func() {
		g.SequenceGenerator = sequenceGenerator
	})
	return
}

// setNodeIDProvider set the sequence NodeID provider
func (g *Generator) setNodeIDProvider(nodeIDProvider NodeIDProvider) (err error) {
	const op errors.Op = "snowflakeGenerator.setNodeIDProvider"
	if nodeIDProvider == nil {
		err = errors.New(nil, op, "NodeIDProvider cannot be nil")
		return
	}
	g.nodeIDProviderGuard.Do(func() {
		g.NodeIDProvider = nodeIDProvider
		var nodeID uint16
		nodeID, err = g.NodeIDProvider()
		if err != nil {
			err = errors.New(nil, op, "error generating nodeID", err)
			return
		}
		g.nodeID = nodeID
	})
	return
}

// Decompose returns a map of snowflake id parts.
func (g *Generator) Decompose(sid uint64) (id ID) {
	timestamp := sid >> timeShift
	id.Timestamp = time.UnixMilli(g.Epoch.UnixMilli() + int64(timestamp*timeUnit))
	id.Sequence = uint16(sid & maskSequence)
	id.NodeID = uint16(sid & maskNodeID >> nodeShift)
	return
}

// NextID generate a new unique snowflake ID
func (g *Generator) NextID() (sid uint64, err error) {
	var current int64
	var seq uint16
	current, err = g.currentTimestamp()
	if err != nil {
		return
	}

	seq, err = g.SequenceGenerator(current)
	if err != nil {
		return
	}

	for seq >= maxSequence {
		current, err = g.waitForNext10Millis(current)
		if err != nil {
			return
		}

		seq, err = g.SequenceGenerator(current)
		if err != nil {
			return
		}
	}
	sid = uint64(current)<<timeShift | uint64(g.nodeID)<<nodeShift | uint64(seq)
	return
}

func (g *Generator) currentTimestamp() (current int64, err error) {
	const op errors.Op = "snowflakeGenerator.currentTimestamp"
	current = g.currentTimeSlot()
	if current < 0 {
		err = errors.New(nil, op, "current time can not be negative, please make sure the epoch is not in the future")
	} else if current > maxTimestamp {
		err = errors.New(nil, op, "timestamp exceeds max time(2^39-1 * 10ms), please check the epoch settings")
	}
	return
}

func (g *Generator) currentTimeSlot() int64 {
	return since(g.Epoch)
}

func since(t time.Time) int64 {
	return time.Since(t).Milliseconds() / timeUnit
}

func (g *Generator) waitForNext10Millis(last int64) (int64, error) {
	current, err := g.currentTimestamp()
	if err != nil {
		return last, err
	}
	for current == last {
		current, err = g.currentTimestamp()
		if err != nil {
			return last, err
		}
	}
	return current, nil
}

func (g *Generator) init(settings Settings) (err error) {
	err = g.setEpoch(settings.Epoch)
	if err != nil {
		return
	}
	err = g.setSequenceGenerator(settings.SequenceGenerator)
	if err != nil {
		return
	}
	err = g.setNodeIDProvider(settings.NodeIDProvider)
	if err != nil {
		return
	}
	return
}
