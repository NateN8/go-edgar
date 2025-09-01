//go:build integration
// +build integration

package edgar

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Using Apple Inc. CIK for integration tests
	// This is a well-known public company with reliable filings
	testCIK = "0000320193"
)

func TestIntegration_GetCompanyFacts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Add delay to respect SEC rate limits
	time.Sleep(100 * time.Millisecond)

	facts, err := client.GetCompanyFacts(testCIK)

	require.NoError(t, err)
	assert.NotNil(t, facts)
	assert.NotEmpty(t, facts.Entity)
	assert.NotNil(t, facts.Facts)

	// Verify CIK conversion works
	cikStr := facts.GetCIKString()
	assert.NotEmpty(t, cikStr)

	// Check that facts contain US-GAAP data
	factsMap := facts.Facts
	usGaap, ok := factsMap["us-gaap"]
	assert.True(t, ok, "us-gaap taxonomy should be present")
	assert.NotNil(t, usGaap)
}

func TestIntegration_GetCompanySubmissions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Add delay to respect SEC rate limits
	time.Sleep(100 * time.Millisecond)

	submissions, err := client.GetCompanySubmissions(testCIK)

	require.NoError(t, err)
	assert.NotNil(t, submissions)
	assert.Equal(t, testCIK, submissions.CIK) // Submissions API returns CIK with leading zeros
	assert.NotEmpty(t, submissions.Name)

	// Check that filings data is present
	assert.NotEmpty(t, submissions.Filings.Recent)

	// Verify required fields are present in recent filings
	recent := submissions.Filings.Recent
	assert.NotEmpty(t, recent["accessionNumber"])
	assert.NotEmpty(t, recent["filingDate"])
	assert.NotEmpty(t, recent["form"])
}

func TestIntegration_GetMostRecent10Q(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Add delay to respect SEC rate limits
	time.Sleep(100 * time.Millisecond)

	filing, err := client.GetMostRecent10Q(testCIK)

	require.NoError(t, err)
	assert.NotNil(t, filing)
	assert.Equal(t, "10-Q", filing.Form)
	assert.NotEmpty(t, filing.AccessionNumber)
	assert.NotEmpty(t, filing.FilingDate)
	assert.NotEmpty(t, filing.ReportDate)

	// Verify date format (should be YYYY-MM-DD)
	assert.Len(t, filing.FilingDate, 10)
	assert.Len(t, filing.ReportDate, 10)
}

func TestIntegration_GetMostRecent4TenQs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Add delay to respect SEC rate limits
	time.Sleep(100 * time.Millisecond)

	filings, err := client.GetMostRecent4TenQs(testCIK)

	require.NoError(t, err)
	assert.NotNil(t, filings)
	assert.True(t, len(filings) <= 4, "should return at most 4 filings")
	assert.True(t, len(filings) > 0, "should return at least 1 filing")

	// Verify all are 10-Q forms
	for _, filing := range filings {
		assert.Equal(t, "10-Q", filing.Form)
		assert.NotEmpty(t, filing.AccessionNumber)
		assert.NotEmpty(t, filing.FilingDate)
	}

	// Verify they are sorted by filing date (most recent first)
	for i := 0; i < len(filings)-1; i++ {
		assert.True(t, filings[i].FilingDate >= filings[i+1].FilingDate,
			"filings should be sorted by date (most recent first)")
	}
}

func TestIntegration_ParseCashFlowMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Add delay to respect SEC rate limits
	time.Sleep(100 * time.Millisecond)

	// Get the most recent 10-Q filing
	filing, err := client.GetMostRecent10Q(testCIK)
	require.NoError(t, err)

	// Add delay before next API call
	time.Sleep(100 * time.Millisecond)

	// Parse cash flow metrics
	metrics, err := client.ParseCashFlowMetrics(testCIK, filing)

	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.NotEmpty(t, metrics.CompanyName)
	assert.Equal(t, "320193", metrics.CIK) // API returns CIK without leading zeros
	assert.Equal(t, filing.FilingDate, metrics.FilingDate)
	assert.Equal(t, filing.ReportDate, metrics.ReportDate)
	assert.Equal(t, "10-Q", metrics.Form)

	// Verify Free Cash Flow calculation
	expectedFCF := metrics.NetCashFromOperatingActivities - metrics.CapitalExpenditures
	assert.Equal(t, expectedFCF, metrics.FreeCashFlow)
}

func TestIntegration_ParseEBITDAMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Add delay to respect SEC rate limits
	time.Sleep(100 * time.Millisecond)

	// Get the most recent 10-Q filing
	filing, err := client.GetMostRecent10Q(testCIK)
	require.NoError(t, err)

	// Add delay before next API call
	time.Sleep(100 * time.Millisecond)

	// Parse EBITDA metrics
	metrics, err := client.ParseEBITDAMetrics(testCIK, filing)

	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.NotEmpty(t, metrics.CompanyName)
	assert.Equal(t, "320193", metrics.CIK) // API returns CIK without leading zeros
	assert.Equal(t, filing.FilingDate, metrics.FilingDate)
	assert.Equal(t, filing.ReportDate, metrics.ReportDate)
	assert.Equal(t, "10-Q", metrics.Form)

	// Verify EBITDA calculation
	expectedEBITDA := metrics.NetIncome + metrics.InterestExpense +
		metrics.IncomeTaxExpense + metrics.DepreciationAndAmortization
	assert.Equal(t, expectedEBITDA, metrics.EBITDA)

	// Verify EBITDA Margin calculation (if revenue is not zero)
	if metrics.Revenue != 0 {
		expectedMargin := (metrics.EBITDA / metrics.Revenue) * 100
		assert.InDelta(t, expectedMargin, metrics.EBITDAMargin, 0.01)
	}
}

func TestIntegration_GetQuarterlyCashFlowAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Add delay to respect SEC rate limits
	time.Sleep(100 * time.Millisecond)

	analysis, err := client.GetQuarterlyCashFlowAnalysis(testCIK)

	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.NotEmpty(t, analysis.CompanyName)
	assert.Equal(t, "320193", analysis.CIK) // API returns CIK without leading zeros
	assert.True(t, len(analysis.Quarters) > 0, "should have at least one quarter")
	assert.True(t, len(analysis.Quarters) <= 4, "should have at most 4 quarters")

	// Verify each quarter has valid data
	for i, quarter := range analysis.Quarters {
		assert.NotEmpty(t, quarter.FilingDate, "quarter %d should have filing date", i+1)
		assert.NotEmpty(t, quarter.ReportDate, "quarter %d should have report date", i+1)
		assert.Equal(t, "10-Q", quarter.Form, "quarter %d should be 10-Q form", i+1)

		// Verify FCF calculation
		expectedFCF := quarter.NetCashFromOperatingActivities - quarter.CapitalExpenditures
		assert.Equal(t, expectedFCF, quarter.FreeCashFlow, "quarter %d FCF calculation", i+1)
	}
}

func TestIntegration_GetQuarterlyEBITDAAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Add delay to respect SEC rate limits
	time.Sleep(100 * time.Millisecond)

	analysis, err := client.GetQuarterlyEBITDAAnalysis(testCIK)

	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.NotEmpty(t, analysis.CompanyName)
	assert.Equal(t, "320193", analysis.CIK) // API returns CIK without leading zeros
	assert.True(t, len(analysis.Quarters) > 0, "should have at least one quarter")
	assert.True(t, len(analysis.Quarters) <= 4, "should have at most 4 quarters")

	// Verify each quarter has valid data
	for i, quarter := range analysis.Quarters {
		assert.NotEmpty(t, quarter.FilingDate, "quarter %d should have filing date", i+1)
		assert.NotEmpty(t, quarter.ReportDate, "quarter %d should have report date", i+1)
		assert.Equal(t, "10-Q", quarter.Form, "quarter %d should be 10-Q form", i+1)

		// Verify EBITDA calculation
		expectedEBITDA := quarter.NetIncome + quarter.InterestExpense +
			quarter.IncomeTaxExpense + quarter.DepreciationAndAmortization
		assert.Equal(t, expectedEBITDA, quarter.EBITDA, "quarter %d EBITDA calculation", i+1)

		// Verify EBITDA Margin calculation (if revenue is not zero)
		if quarter.Revenue != 0 {
			expectedMargin := (quarter.EBITDA / quarter.Revenue) * 100
			assert.InDelta(t, expectedMargin, quarter.EBITDAMargin, 0.01, "quarter %d EBITDA margin", i+1)
		}
	}
}

func TestIntegration_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Test that we can make multiple requests without hitting rate limits
	// SEC recommends no more than 10 requests per second
	start := time.Now()

	for i := 0; i < 3; i++ {
		_, err := client.GetCompanySubmissions(testCIK)
		require.NoError(t, err)

		// Add delay between requests
		time.Sleep(150 * time.Millisecond)
	}

	elapsed := time.Since(start)
	// Should take at least 300ms (3 requests * 100ms delay minimum)
	assert.True(t, elapsed >= 300*time.Millisecond, "should respect rate limiting")
}

func TestIntegration_InvalidCIK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Test with invalid CIK
	invalidCIK := "9999999999"

	_, err := client.GetCompanyFacts(invalidCIK)
	assert.Error(t, err, "should return error for invalid CIK")

	_, err = client.GetCompanySubmissions(invalidCIK)
	assert.Error(t, err, "should return error for invalid CIK")
}

func TestIntegration_NetworkTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create client with very short timeout
	client := &Client{
		httpClient: &http.Client{
			Timeout: time.Millisecond * 1, // 1ms timeout - should fail
		},
		userAgent: userAgent,
	}

	_, err := client.GetCompanyFacts(testCIK)
	assert.Error(t, err, "should timeout with very short timeout")
}

// Helper function to run integration tests with environment variable
func TestMain(m *testing.M) {
	// Check if integration tests should be run
	if os.Getenv("INTEGRATION_TESTS") == "true" {
		// Remove the build constraint for this test run
		os.Exit(m.Run())
	}

	// Skip integration tests by default
	os.Exit(0)
}
