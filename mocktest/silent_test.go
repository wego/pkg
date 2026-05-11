package mocktest_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wego/pkg/mocktest"
)

// SilentT must satisfy mock.TestingT at compile time.
var _ mock.TestingT = mocktest.Silent

type fakeMock struct {
	mock.Mock
}

func (m *fakeMock) Foo() { m.Called() }

func TestSilent_DoesNotMarkOuterTestFailed(t *testing.T) {
	m := &fakeMock{}
	m.On("Foo").Return()

	// Expectation is 1 but no calls have happened — would call t.Errorf
	// if we passed the real t. With Silent, it is swallowed.
	ok := m.AssertNumberOfCalls(mocktest.Silent, "Foo", 1)
	require.False(t, ok, "expected mismatch")
	require.False(t, t.Failed(), "outer test must not be marked failed")
}

func TestSilent_WorksInsideEventually(t *testing.T) {
	m := &fakeMock{}
	m.On("Foo").Return()

	// Fire Foo asynchronously after a short delay so the t=0 evaluation of
	// Eventually sees 0 calls. The poll must NOT fail the outer test.
	go func() {
		time.Sleep(20 * time.Millisecond)
		m.Foo()
	}()

	assert.Eventually(t, func() bool {
		return m.AssertNumberOfCalls(mocktest.Silent, "Foo", 1)
	}, time.Second, 5*time.Millisecond)
	require.False(t, t.Failed())
}

func TestSilent_WorksWithAssertCalled(t *testing.T) {
	m := &fakeMock{}
	m.On("Foo").Return()

	// Before the goroutine fires, AssertCalled returns false but must not
	// mark the outer test failed.
	ok := m.AssertCalled(mocktest.Silent, "Foo")
	require.False(t, ok)
	require.False(t, t.Failed())

	go func() {
		time.Sleep(20 * time.Millisecond)
		m.Foo()
	}()

	assert.Eventually(t, func() bool {
		return m.AssertCalled(mocktest.Silent, "Foo")
	}, time.Second, 5*time.Millisecond)
	require.False(t, t.Failed())
}

func TestSilent_WorksWithAssertNotCalled(t *testing.T) {
	m := &fakeMock{}
	m.On("Foo").Return()

	// AssertNotCalled succeeds while no call has happened.
	require.True(t, m.AssertNotCalled(mocktest.Silent, "Foo"))

	// After the call, AssertNotCalled returns false but must not mark the
	// outer test failed.
	m.Foo()
	require.False(t, m.AssertNotCalled(mocktest.Silent, "Foo"))
	require.False(t, t.Failed())
}

func TestSilent_WorksWithAssertExpectations(t *testing.T) {
	m := &fakeMock{}
	m.On("Foo").Return().Once()

	var done atomic.Bool
	go func() {
		time.Sleep(20 * time.Millisecond)
		m.Foo()
		done.Store(true)
	}()

	assert.Eventually(t, func() bool {
		if !done.Load() {
			return false
		}
		return m.AssertExpectations(mocktest.Silent)
	}, time.Second, 5*time.Millisecond)
	require.False(t, t.Failed())
}
