// Package mocktest provides small helpers for use with
// [github.com/stretchr/testify/mock] assertion helpers.
package mocktest

// SilentT is a [github.com/stretchr/testify/mock.TestingT] implementation that
// discards every call.
//
// Use it as the TestingT argument to mock assertion helpers
// (AssertNumberOfCalls, AssertCalled, AssertNotCalled, AssertExpectations) when
// polling inside [github.com/stretchr/testify/assert.Eventually] or
// [github.com/stretchr/testify/require.Eventually], so the per-poll failure
// does not fail the outer test:
//
//	assert.Eventually(t, func() bool {
//	    return m.AssertNumberOfCalls(mocktest.Silent, "Foo", 1)
//	}, waitFor, tick)
//
// Background: testify v1.11.0 changed Eventually to evaluate the condition at
// t=0 before the first tick (stretchr/testify#1424). Passing the real
// *testing.T (or suite.T()) to a mock assertion inside Eventually causes the
// initial poll — which runs before any async goroutine has had a chance to
// fire — to call t.Errorf and permanently mark the test failed even when the
// condition becomes true on a later tick.
type SilentT struct{}

// Logf implements [github.com/stretchr/testify/mock.TestingT] and discards the call.
func (SilentT) Logf(string, ...any) {}

// Errorf implements [github.com/stretchr/testify/mock.TestingT] and discards the call.
func (SilentT) Errorf(string, ...any) {}

// FailNow implements [github.com/stretchr/testify/mock.TestingT] and discards the call.
func (SilentT) FailNow() {}

// Silent is the shared [SilentT] value. Use it as the TestingT argument to
// mock assertion helpers inside Eventually conditions.
var Silent SilentT
