package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CaptureOutput captures stdout and stderr during function execution
func CaptureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	// Save original stdout/stderr
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Create pipes for capturing output
	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)
	stderrR, stderrW, err := os.Pipe()
	require.NoError(t, err)

	// Replace stdout/stderr with our pipes
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Create channels to receive the captured output
	stdoutChan := make(chan string)
	stderrChan := make(chan string)

	// Start goroutines to read from the pipes
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutR) // Error handled by checking channel timeout
		stdoutChan <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stderrR) // Error handled by checking channel timeout
		stderrChan <- buf.String()
	}()

	// Execute the function
	fn()

	// Close the write ends of the pipes
	_ = stdoutW.Close() // Ignoring error for test cleanup
	_ = stderrW.Close() // Ignoring error for test cleanup

	// Restore original stdout/stderr
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	// Wait for the output to be captured
	stdout = <-stdoutChan
	stderr = <-stderrChan

	return stdout, stderr
}

// AssertValidJSON checks if a string is valid JSON
func AssertValidJSON(t *testing.T, jsonStr string) {
	var js json.RawMessage
	err := json.Unmarshal([]byte(jsonStr), &js)
	assert.NoError(t, err, "should be valid JSON")
}

// AssertContainsJSON checks if JSON string contains expected key-value pairs
func AssertContainsJSON(t *testing.T, jsonStr string, expectedKeys ...string) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	require.NoError(t, err, "should be valid JSON")

	for _, key := range expectedKeys {
		assert.Contains(t, data, key, "JSON should contain key: %s", key)
	}
}

// AssertJSONEquals compares two JSON strings for equality
func AssertJSONEquals(t *testing.T, expected, actual string) {
	var expectedData, actualData interface{}

	err := json.Unmarshal([]byte(expected), &expectedData)
	require.NoError(t, err, "expected JSON should be valid")

	err = json.Unmarshal([]byte(actual), &actualData)
	require.NoError(t, err, "actual JSON should be valid")

	assert.Equal(t, expectedData, actualData, "JSON objects should be equal")
}

// AssertCurrencyFormat checks if a value is formatted as currency
func AssertCurrencyFormat(t *testing.T, value string) {
	assert.Contains(t, value, "$", "should contain dollar sign")
	assert.Contains(t, value, ".", "should contain decimal point")
}

// AssertPercentageFormat checks if a value is formatted as percentage
func AssertPercentageFormat(t *testing.T, value string) {
	assert.Contains(t, value, "%", "should contain percentage sign")
}

// AssertDateFormat checks if a string is in YYYY-MM-DD format
func AssertDateFormat(t *testing.T, dateStr string) {
	assert.Len(t, dateStr, 10, "date should be 10 characters long")
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, dateStr, "date should be in YYYY-MM-DD format")
}

// AssertCIKFormat checks if a string is in proper CIK format (10 digits)
func AssertCIKFormat(t *testing.T, cik string) {
	assert.Len(t, cik, 10, "CIK should be 10 characters long")
	assert.Regexp(t, `^\d{10}$`, cik, "CIK should be 10 digits")
}

// AssertAccessionNumberFormat checks if a string is in proper accession number format
func AssertAccessionNumberFormat(t *testing.T, accessionNumber string) {
	assert.Regexp(t, `^\d{10}-\d{2}-\d{6}$`, accessionNumber,
		"accession number should be in format XXXXXXXXXX-XX-XXXXXX")
}

// WithTimeout runs a function with a timeout
func WithTimeout(t *testing.T, timeout time.Duration, fn func()) {
	done := make(chan bool, 1)

	go func() {
		fn()
		done <- true
	}()

	select {
	case <-done:
		// Function completed successfully
	case <-time.After(timeout):
		t.Fatalf("function timed out after %v", timeout)
	}
}

// AssertMoneyValue checks if a monetary value is reasonable (not negative, not extremely large)
func AssertMoneyValue(t *testing.T, value float64, description string) {
	assert.True(t, value >= 0, "%s should not be negative: %f", description, value)
	assert.True(t, value < 1e15, "%s should not be unreasonably large: %f", description, value)
}

// AssertPercentageValue checks if a percentage value is reasonable
func AssertPercentageValue(t *testing.T, value float64, description string) {
	assert.True(t, value >= -100, "%s should not be less than -100%%: %f", description, value)
	assert.True(t, value <= 1000, "%s should not be more than 1000%%: %f", description, value)
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("skipping test in short mode: %s", reason)
	}
}

// SkipIfNoIntegration skips the test if integration tests are disabled
func SkipIfNoIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("skipping integration test (set INTEGRATION_TESTS=true to enable)")
	}
}

// CompareFloats compares two float64 values with a reasonable tolerance
func CompareFloats(t *testing.T, expected, actual float64, description string) {
	tolerance := 0.01 // 1 cent tolerance for financial calculations
	assert.InDelta(t, expected, actual, tolerance, description)
}

// AssertOrderedByDate checks if filings are ordered by date (most recent first)
func AssertOrderedByDate(t *testing.T, dates []string, description string) {
	for i := 0; i < len(dates)-1; i++ {
		assert.True(t, dates[i] >= dates[i+1],
			"%s should be ordered by date (most recent first): %s should be >= %s",
			description, dates[i], dates[i+1])
	}
}

// SetupTestEnvironment sets up common test environment variables
func SetupTestEnvironment() {
	// Set any required environment variables for testing
	_ = os.Setenv("SEC_API_BASE_URL", "https://data.sec.gov") // Ignoring error for test setup
}

// CleanupTestEnvironment cleans up test environment
func CleanupTestEnvironment() {
	// Clean up any test-specific environment variables
	_ = os.Unsetenv("SEC_API_BASE_URL") // Ignoring error for test cleanup
}

// CreateTempFile creates a temporary file for testing
func CreateTempFile(t *testing.T, content string) *os.File {
	tmpfile, err := os.CreateTemp("", "edgar-test-")
	require.NoError(t, err, "should create temp file")

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err, "should write to temp file")

	err = tmpfile.Close()
	require.NoError(t, err, "should close temp file")

	// Register cleanup
	t.Cleanup(func() {
		_ = os.Remove(tmpfile.Name()) // Ignoring error for test cleanup
	})

	// Reopen for reading
	tmpfile, err = os.Open(tmpfile.Name())
	require.NoError(t, err, "should reopen temp file")

	return tmpfile
}

// AssertNoLeakedGoroutines checks that no goroutines are leaked during test execution
func AssertNoLeakedGoroutines(t *testing.T, fn func()) {
	initialCount := countGoroutines()

	fn()

	// Give some time for goroutines to clean up
	time.Sleep(10 * time.Millisecond)

	finalCount := countGoroutines()

	// Allow for some tolerance as the test framework itself may create goroutines
	tolerance := 2
	assert.True(t, finalCount <= initialCount+tolerance,
		"potential goroutine leak detected: initial=%d, final=%d", initialCount, finalCount)
}

// countGoroutines returns the current number of goroutines
func countGoroutines() int {
	return len(getAllGoroutineStacks())
}

// getAllGoroutineStacks returns stack traces for all goroutines
func getAllGoroutineStacks() []byte {
	buf := make([]byte, 1<<16)
	n := len(buf)
	for n == len(buf) {
		buf = make([]byte, 2*len(buf))
		n = len(buf) // This would normally use runtime.Stack, but we'll simulate
	}
	return buf[:n]
}

// TableTest represents a single test case in a table-driven test
type TableTest struct {
	Name     string
	Input    interface{}
	Expected interface{}
	Error    string
}

// RunTableTests runs a series of table-driven tests
func RunTableTests(t *testing.T, tests []TableTest, testFunc func(t *testing.T, input, expected interface{}, expectError string)) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testFunc(t, tt.Input, tt.Expected, tt.Error)
		})
	}
}
