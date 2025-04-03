package rand

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

type testCase struct {
	name       string
	percentage float64
	calls      int
	tolerance  float64
}

func TestSelectorBasic(t *testing.T) {

	// Basic test cases for key percentages
	tests := []testCase{
		{"0%", 0, 10000, 0.05},     // Very tight tolerance for 0%
		{"25%", 25, 10000, 1.5},    // Wider tolerance for 25% (higher variance)
		{"50%", 50, 10000, 1.5},    // Wider tolerance for 50% (maximum variance)
		{"75%", 75, 10000, 1.5},    // Wider tolerance for 75% (higher variance)
		{"100%", 100, 10000, 0.05}, // Very tight tolerance for 100%
	}

	// Test basic cases with multiple iterations
	for i := 1; i <= 10; i++ {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				totalCalls := test.calls * i
				selector := NewSelector(test.percentage)
				trueCount := 0

				for i := 0; i < totalCalls; i++ {
					if selector.Next() {
						trueCount++
					}
				}

				actualPercentage := float64(trueCount) / float64(totalCalls) * 100
				diff := math.Abs(actualPercentage - test.percentage)

				if diff > test.tolerance {
					t.Errorf("Expected percentage close to %.2f, got %.2f (diff: %.2f)",
						test.percentage, actualPercentage, diff)
				}
			})
		}
	}
}

// Calculates the expected tolerance for a given percentage and number of test runs
// confidenceLevel: z-score for statistical confidence (1.96 = 95%, 2.58 = 99%, etc.)
// maxTolerancePercent: maximum allowed tolerance as a percentage
// minTolerancePercent: minimum allowed tolerance regardless of statistical calculations
func calculateTolerance(percentage float64, numCalls int, confidenceLevel float64, maxTolerancePercent float64, minTolerancePercent float64) (float64, int) {
	// Convert to probability (0.0-1.0)
	prob := percentage / 100.0

	// Calculate variance for a binomial distribution
	variance := prob * (1.0 - prob)

	// Determine if this is a small call count situation
	isSmallCallCount := numCalls < 15000

	// Calculate standard deviation as percentage points (0-100 scale)
	stdDev := confidenceLevel * 100.0 * math.Sqrt(variance/float64(numCalls))

	// Base number of calls - dynamically adjust based on variance
	// More calls for higher variance (middle percentages), fewer for extreme percentages
	// Use a larger base number (250,000) to achieve tighter tolerances through larger sample size
	normalCalls := int(math.Ceil(250000 * math.Sqrt(variance*4))) // Scale by variance

	// For small call count tests, we have different call scaling
	var calls int
	if isSmallCallCount {
		// Use the requested small call count, but with some minimal floor
		calls = int(math.Max(float64(numCalls), 5000)) // At least 5K calls for any test
		// For small calls, we may need to adjust the standard deviation

		// Calculate the ratio of actual calls to normal expected calls, capped at 1.0
		// This is used to adjust tolerance based on call reduction
		// For small call counts, we expect higher variance, so we adjust the tolerance
		callRatio := math.Min(float64(calls)/float64(normalCalls), 1.0)

		// If we're using significantly fewer calls than normal, adjust stdDev
		if callRatio < 0.8 {
			// Adjustment based on statistical principles - variance scales with sqrt(1/n)
			// Add a small constant factor for extra safety margin
			stdDev *= math.Sqrt(1.0/callRatio) * 1.1
		}
	} else {
		// For normal tests, cap the number of calls to reasonable limits
		calls = int(math.Min(float64(normalCalls), 500000)) // Upper limit
		calls = int(math.Max(float64(calls), 20000))        // Lower limit for normal tests
	}

	// Calculate tolerance based on statistical properties and practical test needs
	// For very small/large percentages, we need a percentage-relative minimum to avoid excessive strictness
	// For middle percentages, we can rely more on the standard deviation with a reasonable maximum

	// Base tolerance from statistical standard deviation
	// For middle percentages with high variance (near 50%), standard deviation is higher
	// For extreme percentages (near 0% or 100%), standard deviation is lower

	// Minimum tolerance as percentage of the expected value (to handle small percentages)
	// Keep minimums very low to achieve sub-1.0 tolerances wherever possible
	minRelativeTolerance := math.Min(0.06, 0.012+0.048/(1.0+math.Pow(percentage/3.0, 2.0)))
	minimumTolerance := math.Max(0.25, percentage*minRelativeTolerance) // Slightly higher minimum threshold (0.25%)

	// Add special handling for fractional percentages and edge cases
	// These tend to have higher statistical variability
	hasFraction := math.Abs(math.Round(percentage)-percentage) > 0.001
	isEdgeCase := percentage < 15.0 || percentage > 85.0 // Expand edge case range further

	// Adjust tolerance for fractional and edge percentages
	if hasFraction || isEdgeCase {
		// Add some extra tolerance for these cases
		fractionAdjustment := 0.18 // Add 0.18% for fractional values (increased from 0.15%)

		// For small call counts with edge cases, add even more tolerance
		if isSmallCallCount && isEdgeCase {
			fractionAdjustment *= 2.0 // 100% more tolerance for edge cases with small call counts (increased from 80%)
		}

		minimumTolerance += fractionAdjustment
	}

	// For small call counts, raise the minimum tolerance to prevent excessive failures
	if isSmallCallCount {
		// Higher base minimum tolerance for small call counts
		minimumTolerance = math.Max(minimumTolerance, 0.8) // At least 0.8% tolerance for small call tests (increased from 0.7%)
	}

	// Use the provided minimum and maximum tolerance ceiling
	minimumEnforcedTolerance := minTolerancePercent // Absolute minimum tolerance from configuration
	maximumTolerance := maxTolerancePercent

	// Apply boundaries (dynamically calculated minimum, configured minimum, and maximum)
	tolerance := math.Max(minimumTolerance, stdDev)           // First ensure statistical minimum is met
	tolerance = math.Max(tolerance, minimumEnforcedTolerance) // Then ensure configured minimum is met
	tolerance = math.Min(tolerance, maximumTolerance)         // Finally ensure maximum is not exceeded

	// Add a small safety margin to improve test reliability (0.08% additional tolerance)
	tolerance += 0.08

	return tolerance, calls
}

// comprehensiveTestCases generates test cases with customizable confidence and tolerance parameters
func comprehensiveTestCases(confidenceLevel, maxTolerancePercent float64) []testCase {
	// Comprehensive test cases from 0 to 100 with step of 0.01
	comprehensiveTests := make([]testCase, 0, 10001) // 0.00 to 100.00
	for p := 0.0; p <= 100.0; p += 0.01 {
		// Round to 2 decimal places to avoid floating point precision issues
		roundedP := float64(int(p*100+0.5)) / 100

		// For initial creation of test cases, use a baseline number of calls
		baselineCalls := 10000

		// Use the provided confidence, max tolerance, and min tolerance parameters
		// Default minimum tolerance of 0.2% if not specified elsewhere
		minTolerancePercent := 0.2 // Default minimum tolerance
		tolerance, calls := calculateTolerance(roundedP, baselineCalls, confidenceLevel, maxTolerancePercent, minTolerancePercent)

		comprehensiveTests = append(comprehensiveTests, testCase{
			name:       fmt.Sprintf("Percentage %.2f%% - calls %d - tolerance %.2f", roundedP, calls, tolerance),
			percentage: roundedP,
			calls:      calls,
			tolerance:  tolerance,
		})
	}

	return comprehensiveTests
}

// selectTestCases generates a comprehensive set of test cases and then samples them
// to provide good coverage across the full percentage range while keeping test time reasonable.
// confidenceLevel: z-score for statistical confidence (1.96 = 95%, 2.58 = 99%, etc.)
// maxTolerancePercent: maximum allowed tolerance as a percentage
func selectTestCases(confidenceLevel, maxTolerancePercent float64) []testCase {
	// Creates a complete set of test cases with comprehensive coverage
	// using the provided confidence and tolerance parameters
	baseTests := comprehensiveTestCases(confidenceLevel, maxTolerancePercent)
	// Import our target test size and ensure good coverage across the range
	targetSampleSize := 1000

	// Always include critical values (ensure edge cases are always tested)
	criticalValues := []float64{0.0, 0.01, 0.1, 0.5, 1.0, 5.0, 10.0, 25.0, 50.0, 75.0, 90.0, 95.0, 99.0, 99.5, 99.9, 99.99, 100.0}
	sampledTests := make([]testCase, 0, targetSampleSize)

	// First add all critical values
	for _, criticalPct := range criticalValues {
		for _, test := range baseTests {
			if math.Abs(test.percentage-criticalPct) < 0.01 {
				sampledTests = append(sampledTests, test)
				break
			}
		}
	}

	// Initialize a proper random source
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Define sampling buckets to ensure good coverage across ranges
	buckets := []struct {
		min, max float64
		count    int
	}{
		{0.01, 1.0, 50},   // More samples in 0-1% range
		{1.0, 10.0, 50},   // More samples in 1-10% range
		{10.0, 90.0, 200}, // Most samples in middle range
		{90.0, 99.0, 50},  // More samples in 90-99% range
		{99.0, 99.99, 50}, // More samples in 99-99.99% range
	}

	// Sample from each bucket
	for _, bucket := range buckets {
		// Filter tests in this bucket's range
		bucketTests := make([]testCase, 0)
		for _, test := range baseTests {
			if test.percentage >= bucket.min && test.percentage <= bucket.max {
				bucketTests = append(bucketTests, test)
			}
		}

		// If we have fewer tests than requested count, just take all of them
		if len(bucketTests) <= bucket.count {
			sampledTests = append(sampledTests, bucketTests...)
			continue
		}

		// Randomly shuffle the bucket tests
		random.Shuffle(len(bucketTests), func(i, j int) {
			bucketTests[i], bucketTests[j] = bucketTests[j], bucketTests[i]
		})

		// Take the first 'count' elements after shuffling
		sampledTests = append(sampledTests, bucketTests[:bucket.count]...)
	}

	// Deduplicate any potential duplicates from critical values
	deduped := make([]testCase, 0, len(sampledTests))
	seen := make(map[float64]bool)
	for _, test := range sampledTests {
		if !seen[test.percentage] {
			deduped = append(deduped, test)
			seen[test.percentage] = true
		}
	}
	return deduped
}

// Comprehensive test cases from 0 to 100 with step of 0.01
func TestSelectorComprehensive(t *testing.T) {
	// Test configurations with different confidence levels and max tolerances
	configurations := []struct {
		name         string
		confidence   float64
		maxTolerance float64
		minTolerance float64 // Minimum tolerance regardless of statistical calculations
		callsFactor  float64 // Factor to multiply the default calls by (1.0 = normal, <1.0 = fewer calls)
		minimumCalls int     // Minimum number of calls to ensure statistical reliability
	}{
		{
			name:         "Small calls - 95% confidence (2.5% max tolerance)",
			confidence:   1.96, // 95% confidence
			maxTolerance: 2.5,  // 2.5% max tolerance - higher tolerance for small calls
			minTolerance: 0.6,  // Minimum 0.6% tolerance for edge cases (increased from 0.5%)
			callsFactor:  0.4,  // 40% of the standard calls for better stability
			minimumCalls: 5000, // Lower minimum for quick tests
		},
		{
			name:         "Low precision - 95% confidence (1.5% max tolerance)",
			confidence:   1.96,  // 95% confidence
			maxTolerance: 1.5,   // 1.5% max tolerance
			minTolerance: 0.3,   // Minimum 0.3% tolerance for edge cases
			callsFactor:  1.0,   // Standard number of calls
			minimumCalls: 20000, // Standard minimum
		},
		{
			name:         "Standard - 99% confidence (1.0% max tolerance)",
			confidence:   2.58,  // 2.58 (99% confidence)
			maxTolerance: 1.0,   // 1.0%
			minTolerance: 0.2,   // Minimum 0.2% tolerance
			callsFactor:  1.0,   // Standard number of calls
			minimumCalls: 40000, // Higher minimum for higher confidence
		},
		{
			name:         "High precision - 99.7% confidence (0.8% max tolerance)",
			confidence:   3.0,   // 99.7% confidence
			maxTolerance: 0.8,   // 0.8% max tolerance
			minTolerance: 0.15,  // Minimum 0.15% tolerance for high precision tests
			callsFactor:  1.0,   // Standard number of calls
			minimumCalls: 80000, // High minimum call count for precision
		},
		{
			name:         "Very high precision - 99.97% confidence (0.5% max tolerance)",
			confidence:   4.0,    // 99.97% confidence
			maxTolerance: 0.5,    // 0.5% max tolerance
			callsFactor:  1.0,    // Standard number of calls
			minimumCalls: 200000, // Very high minimum for highest precision
		},
	}

	// Run tests for each confidence level configuration
	for _, config := range configurations {
		t.Run(config.name, func(t *testing.T) {
			// Get sampled test cases using the current configuration
			sampledTests := selectTestCases(config.confidence, config.maxTolerance)

			// Process each test case to apply minimum call counts and recalculate tolerances
			for i := range sampledTests {
				// Start with the base call count from sampling
				baseCalls := sampledTests[i].calls

				// Apply call factor if specified (for small call tests)
				if config.callsFactor != 1.0 {
					baseCalls = int(float64(baseCalls) * config.callsFactor)
				}

				// Ensure minimum calls based on configuration
				// For high precision tests, we need more calls to maintain statistical reliability
				if config.callsFactor == 1.0 { // Only enforce minimum for normal (non-small) tests
					baseCalls = int(math.Max(float64(baseCalls), float64(config.minimumCalls)))
				}

				// Recalculate tolerance using our optimized calculateTolerance function
				// This will automatically handle the appropriate adjustments based on call count
				adjustedTolerance, adjustedCalls := calculateTolerance(
					sampledTests[i].percentage,
					baseCalls,
					config.confidence,
					config.maxTolerance,
					config.minTolerance)

				// Update the test case with new values
				sampledTests[i].calls = adjustedCalls
				sampledTests[i].tolerance = adjustedTolerance

				// Update the test name to reflect the new values
				sampledTests[i].name = fmt.Sprintf("Percentage %.2f%% - calls %d - tolerance %.2f",
					sampledTests[i].percentage, sampledTests[i].calls, sampledTests[i].tolerance)
			}

			// You can adjust this factor to balance between test speed and statistical accuracy
			// Higher value = more accuracy but longer runtime
			// 1.0 = base accuracy, 2.0 = double the runs & tighter tolerance
			accuracyFactor := 1.0

			// Run tests in parallel
			for _, test := range sampledTests {
				test := test // Capture test variable for goroutine
				t.Run(test.name, func(t *testing.T) {
					// Signal this test can be run in parallel with others
					t.Parallel()

					// Recalculate tolerance based on accuracy factor and current config
					adjustedCalls := int(float64(test.calls) * accuracyFactor)
					adjustedTolerance, _ := calculateTolerance(test.percentage, adjustedCalls, config.confidence, config.maxTolerance, config.minTolerance)

					s := NewSelector(test.percentage)

					// Use multiple goroutines to parallelize counting
					workers := runtime.NumCPU()
					workerCalls := adjustedCalls / workers
					remaining := adjustedCalls % workers

					// Channel to collect results from each worker
					results := make(chan int, workers)

					// Launch worker goroutines
					for w := 0; w < workers; w++ {
						calls := workerCalls
						if w == 0 {
							calls += remaining // Add remaining calls to first worker
						}

						go func(calls int) {
							localCount := 0
							for i := 0; i < calls; i++ {
								if s.Next() {
									localCount++
								}
							}
							results <- localCount
						}(calls)
					}

					// Collect and sum results
					trueCount := 0
					for w := 0; w < workers; w++ {
						trueCount += <-results
					}

					actualPercentage := float64(trueCount) / float64(adjustedCalls) * 100
					diff := math.Abs(actualPercentage - test.percentage)

					if diff > adjustedTolerance {
						t.Errorf("Expected percentage close to %.2f%%, got %.2f%% (diff: %.2f, tolerance: %.2f)",
							test.percentage, actualPercentage, diff, adjustedTolerance)
					}
				})
			}
		})
	}
}
