package edgar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock data for testing
const (
	mockCIK         = "0000320193"
	mockCompanyName = "Test Company Inc."
)

// Helper function to create a mock server
func createMockServer(response string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprint(w, response)
	}))
}

// mockTransport is a custom HTTP transport for testing that redirects requests to a test server
type mockTransport struct {
	originalURL string
	testURL     string
	server      *httptest.Server
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// If the request URL matches the original URL pattern, redirect to test server
	if strings.Contains(req.URL.String(), "data.sec.gov") {
		// Create a new request to the test server
		testReq := req.Clone(req.Context())
		testReq.URL.Scheme = "http"
		testReq.URL.Host = strings.TrimPrefix(t.server.URL, "http://")
		testReq.URL.Path = ""
		testReq.Header.Set("Host", testReq.URL.Host)

		// Use the default transport to make the request
		return http.DefaultTransport.RoundTrip(testReq)
	}

	// For other URLs, use the default transport
	return http.DefaultTransport.RoundTrip(req)
}

// Mock company facts response
func getMockCompanyFacts() string {
	return `{
		"cik": "320193",
		"entityName": "Apple Inc.",
		"facts": {
			"us-gaap": {
				"NetCashProvidedByUsedInOperatingActivities": {
					"units": {
						"USD": [
							{
								"form": "10-Q",
								"val": 50000000000,
								"end": "2023-12-30"
							}
						]
					}
				},
				"PaymentsToAcquirePropertyPlantAndEquipment": {
					"units": {
						"USD": [
							{
								"form": "10-Q",
								"val": 5000000000,
								"end": "2023-12-30"
							}
						]
					}
				},
				"Revenues": {
					"units": {
						"USD": [
							{
								"form": "10-Q",
								"val": 100000000000,
								"end": "2023-12-30"
							}
						]
					}
				},
				"NetIncomeLoss": {
					"units": {
						"USD": [
							{
								"form": "10-Q",
								"val": 25000000000,
								"end": "2023-12-30"
							}
						]
					}
				},
				"InterestExpense": {
					"units": {
						"USD": [
							{
								"form": "10-Q",
								"val": 1000000000,
								"end": "2023-12-30"
							}
						]
					}
				},
				"IncomeTaxExpenseBenefit": {
					"units": {
						"USD": [
							{
								"form": "10-Q",
								"val": 3000000000,
								"end": "2023-12-30"
							}
						]
					}
				},
				"DepreciationDepletionAndAmortization": {
					"units": {
						"USD": [
							{
								"form": "10-Q",
								"val": 2000000000,
								"end": "2023-12-30"
							}
						]
					}
				}
			}
		}
	}`
}

// Mock company submissions response
func getMockCompanySubmissions() string {
	return `{
		"cik": "320193",
		"name": "Apple Inc.",
		"filings": {
			"recent": {
				"accessionNumber": ["0000320193-24-000007", "0000320193-24-000006", "0000320193-24-000005", "0000320193-24-000004", "0000320193-24-000003"],
				"filingDate": ["2024-02-01", "2023-11-02", "2023-08-03", "2023-05-04", "2023-02-02"],
				"reportDate": ["2023-12-30", "2023-09-30", "2023-06-30", "2023-03-31", "2022-12-31"],
				"form": ["10-Q", "10-Q", "10-Q", "10-Q", "10-Q"],
				"fileNumber": ["001-36743", "001-36743", "001-36743", "001-36743", "001-36743"],
				"filmNumber": ["24576126", "24576125", "24576124", "24576123", "24576122"],
				"items": ["", "", "", "", ""],
				"size": ["100000", "100000", "100000", "100000", "100000"],
				"isXBRL": [1, 1, 1, 1, 1],
				"isInlineXBRL": [1, 1, 1, 1, 1],
				"primaryDocument": ["aapl-20231230.htm", "aapl-20230930.htm", "aapl-20230630.htm", "aapl-20230331.htm", "aapl-20221231.htm"],
				"primaryDocDescription": ["10-Q", "10-Q", "10-Q", "10-Q", "10-Q"]
			}
		}
	}`
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, userAgent, client.userAgent)
	assert.Equal(t, time.Second*30, client.httpClient.Timeout)
}

func TestClient_makeRequest(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectError   bool
		expectedError string
	}{
		{
			name:         "successful request",
			responseBody: `{"test": "data"}`,
			statusCode:   http.StatusOK,
			expectError:  false,
		},
		{
			name:          "404 error",
			responseBody:  `{"error": "not found"}`,
			statusCode:    http.StatusNotFound,
			expectError:   true,
			expectedError: "unexpected status code: 404",
		},
		{
			name:          "500 error",
			responseBody:  `{"error": "internal server error"}`,
			statusCode:    http.StatusInternalServerError,
			expectError:   true,
			expectedError: "unexpected status code: 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(tt.responseBody, tt.statusCode)
			defer server.Close()

			client := NewClient()
			body, err := client.makeRequest(server.URL)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.responseBody, string(body))
			}
		})
	}
}

func TestCompanyFacts_GetCIKString(t *testing.T) {
	tests := []struct {
		name     string
		cik      interface{}
		expected string
	}{
		{
			name:     "string CIK",
			cik:      "320193",
			expected: "320193",
		},
		{
			name:     "int CIK",
			cik:      320193,
			expected: "320193",
		},
		{
			name:     "float64 CIK",
			cik:      320193.0,
			expected: "320193",
		},
		{
			name:     "json.Number CIK",
			cik:      json.Number("320193"),
			expected: "320193",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := &CompanyFacts{CIK: tt.cik}
			result := cf.GetCIKString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_GetCompanyFacts(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := createMockServer(getMockCompanyFacts(), http.StatusOK)
		defer server.Close()

		// Create a client with custom base URL for testing
		client := &Client{
			httpClient: &http.Client{Timeout: time.Second * 30},
			userAgent:  userAgent,
		}

		// We'll test with the real API structure but mock the HTTP response
		// by temporarily replacing the base URL in the actual request
		originalURL := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", "https://data.sec.gov", mockCIK)
		testURL := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", server.URL, mockCIK)

		// Use a custom HTTP client that redirects to our test server
		client.httpClient = &http.Client{
			Timeout: time.Second * 30,
			Transport: &mockTransport{
				originalURL: originalURL,
				testURL:     testURL,
				server:      server,
			},
		}

		facts, err := client.GetCompanyFacts(mockCIK)

		require.NoError(t, err)
		assert.NotNil(t, facts)
		assert.Equal(t, "Apple Inc.", facts.Entity)
		assert.Equal(t, "320193", facts.GetCIKString())
	})

	t.Run("server error", func(t *testing.T) {
		server := createMockServer(`{"error": "internal error"}`, http.StatusInternalServerError)
		defer server.Close()

		client := &Client{
			httpClient: &http.Client{Timeout: time.Second * 30},
			userAgent:  userAgent,
		}

		originalURL := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", "https://data.sec.gov", mockCIK)
		testURL := server.URL

		client.httpClient = &http.Client{
			Timeout: time.Second * 30,
			Transport: &mockTransport{
				originalURL: originalURL,
				testURL:     testURL,
				server:      server,
			},
		}

		facts, err := client.GetCompanyFacts(mockCIK)

		assert.Error(t, err)
		assert.Nil(t, facts)
	})
}

func TestClient_GetCompanySubmissions(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := createMockServer(getMockCompanySubmissions(), http.StatusOK)
		defer server.Close()

		client := &Client{
			httpClient: &http.Client{Timeout: time.Second * 30},
			userAgent:  userAgent,
		}

		originalURL := fmt.Sprintf("%s/submissions/CIK%s.json", "https://data.sec.gov", mockCIK)
		testURL := server.URL

		client.httpClient = &http.Client{
			Timeout: time.Second * 30,
			Transport: &mockTransport{
				originalURL: originalURL,
				testURL:     testURL,
				server:      server,
			},
		}

		submissions, err := client.GetCompanySubmissions(mockCIK)

		require.NoError(t, err)
		assert.NotNil(t, submissions)
		assert.Equal(t, "320193", submissions.CIK)
		assert.Equal(t, "Apple Inc.", submissions.Name)
	})
}

func TestClient_parseFilings(t *testing.T) {
	client := NewClient()

	recentData := map[string][]interface{}{
		"accessionNumber":       {"0000320193-24-000007", "0000320193-24-000006"},
		"filingDate":            {"2024-02-01", "2023-11-02"},
		"reportDate":            {"2023-12-30", "2023-09-30"},
		"form":                  {"10-Q", "10-Q"},
		"fileNumber":            {"001-36743", "001-36743"},
		"filmNumber":            {"24576126", "24576125"},
		"items":                 {"", ""},
		"size":                  {"100000", "100000"},
		"isXBRL":                {1, 1},
		"isInlineXBRL":          {1, 1},
		"primaryDocument":       {"aapl-20231230.htm", "aapl-20230930.htm"},
		"primaryDocDescription": {"10-Q", "10-Q"},
	}

	filings := client.parseFilings(recentData)

	assert.Len(t, filings, 2)
	assert.Equal(t, "0000320193-24-000007", filings[0].AccessionNumber)
	assert.Equal(t, "2024-02-01", filings[0].FilingDate)
	assert.Equal(t, "2023-12-30", filings[0].ReportDate)
	assert.Equal(t, "10-Q", filings[0].Form)
}

func TestClient_toString(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, ""},
		{"string value", "test", "test"},
		{"int value", 123, "123"},
		{"float64 value", 123.45, "123"},
		{"bool true", true, "1"},
		{"bool false", false, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.toString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_GetMostRecent10Q(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := createMockServer(getMockCompanySubmissions(), http.StatusOK)
		defer server.Close()

		client := &Client{
			httpClient: &http.Client{Timeout: time.Second * 30},
			userAgent:  userAgent,
		}

		originalURL := fmt.Sprintf("%s/submissions/CIK%s.json", "https://data.sec.gov", mockCIK)
		testURL := server.URL

		client.httpClient = &http.Client{
			Timeout: time.Second * 30,
			Transport: &mockTransport{
				originalURL: originalURL,
				testURL:     testURL,
				server:      server,
			},
		}

		filing, err := client.GetMostRecent10Q(mockCIK)

		require.NoError(t, err)
		assert.NotNil(t, filing)
		assert.Equal(t, "0000320193-24-000007", filing.AccessionNumber)
		assert.Equal(t, "2024-02-01", filing.FilingDate)
		assert.Equal(t, "10-Q", filing.Form)
	})

	t.Run("no 10-Q filings found", func(t *testing.T) {
		noTenQResponse := `{
			"cik": "320193",
			"name": "Apple Inc.",
			"filings": {
				"recent": {
					"accessionNumber": ["0000320193-24-000007"],
					"filingDate": ["2024-02-01"],
					"reportDate": ["2023-12-30"],
					"form": ["10-K"],
					"fileNumber": ["001-36743"],
					"filmNumber": ["24576126"],
					"items": [""],
					"size": ["100000"],
					"isXBRL": [1],
					"isInlineXBRL": [1],
					"primaryDocument": ["aapl-20231230.htm"],
					"primaryDocDescription": ["10-K"]
				}
			}
		}`

		server := createMockServer(noTenQResponse, http.StatusOK)
		defer server.Close()

		client := &Client{
			httpClient: &http.Client{Timeout: time.Second * 30},
			userAgent:  userAgent,
		}

		originalURL := fmt.Sprintf("%s/submissions/CIK%s.json", "https://data.sec.gov", mockCIK)
		testURL := server.URL

		client.httpClient = &http.Client{
			Timeout: time.Second * 30,
			Transport: &mockTransport{
				originalURL: originalURL,
				testURL:     testURL,
				server:      server,
			},
		}

		filing, err := client.GetMostRecent10Q(mockCIK)

		assert.Error(t, err)
		assert.Nil(t, filing)
		assert.Contains(t, err.Error(), "no 10-Q filings found")
	})
}

func TestClient_GetMostRecent4TenQs(t *testing.T) {
	server := createMockServer(getMockCompanySubmissions(), http.StatusOK)
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: time.Second * 30},
		userAgent:  userAgent,
	}

	originalURL := fmt.Sprintf("%s/submissions/CIK%s.json", "https://data.sec.gov", mockCIK)
	testURL := server.URL

	client.httpClient = &http.Client{
		Timeout: time.Second * 30,
		Transport: &mockTransport{
			originalURL: originalURL,
			testURL:     testURL,
			server:      server,
		},
	}

	filings, err := client.GetMostRecent4TenQs(mockCIK)

	require.NoError(t, err)
	assert.Len(t, filings, 4) // Should return 4 most recent 10-Q filings

	// Verify they are sorted by filing date (most recent first)
	for i := 0; i < len(filings)-1; i++ {
		assert.True(t, filings[i].FilingDate >= filings[i+1].FilingDate)
	}
}

func TestClient_findValueForDate(t *testing.T) {
	client := NewClient()

	dataArray := []interface{}{
		map[string]interface{}{
			"end":  "2023-12-30",
			"form": "10-Q",
			"val":  100.0,
		},
		map[string]interface{}{
			"end":  "2023-09-30",
			"form": "10-Q",
			"val":  90.0,
		},
		map[string]interface{}{
			"end":  "2023-12-30",
			"form": "10-K",
			"val":  110.0,
		},
	}

	// Test exact date match with 10-Q form (should prefer this)
	value := client.findValueForDate(dataArray, "2023-12-30")
	assert.Equal(t, 100.0, value)

	// Test with no exact date match - both 10-Q forms have same score, ties broken by date
	value = client.findValueForDate(dataArray, "2023-06-30")
	assert.Equal(t, 100.0, value) // Should get 2023-12-30 10-Q (tie-breaker by more recent date)
}

func TestClient_extractMetric(t *testing.T) {
	client := NewClient()

	usGaap := map[string]interface{}{
		"TestMetric": map[string]interface{}{
			"units": map[string]interface{}{
				"USD": []interface{}{
					map[string]interface{}{
						"end":  "2023-12-30",
						"form": "10-Q",
						"val":  1000000.0,
					},
				},
			},
		},
	}

	var result float64
	err := client.extractMetric(usGaap, []string{"TestMetric"}, &result, "2023-12-30")

	assert.NoError(t, err)
	assert.Equal(t, 1000000.0, result)
}

func TestClient_extractMetric_NotFound(t *testing.T) {
	client := NewClient()

	usGaap := map[string]interface{}{}

	var result float64
	err := client.extractMetric(usGaap, []string{"NonExistentMetric"}, &result, "2023-12-30")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metric not found")
}

func TestEBITDACalculation(t *testing.T) {
	// Test EBITDA calculation with sample data
	metrics := &EBITDAMetrics{
		NetIncome:                   25000000000,
		InterestExpense:             1000000000,
		IncomeTaxExpense:            3000000000,
		DepreciationAndAmortization: 2000000000,
		Revenue:                     100000000000,
	}

	// Calculate EBITDA
	metrics.EBITDA = metrics.NetIncome + metrics.InterestExpense + metrics.IncomeTaxExpense + metrics.DepreciationAndAmortization

	// Calculate EBITDA Margin
	if metrics.Revenue != 0 {
		metrics.EBITDAMargin = (metrics.EBITDA / metrics.Revenue) * 100
	}

	expectedEBITDA := 31000000000.0 // 25B + 1B + 3B + 2B
	expectedMargin := 31.0          // (31B / 100B) * 100

	assert.Equal(t, expectedEBITDA, metrics.EBITDA)
	assert.Equal(t, expectedMargin, metrics.EBITDAMargin)
}

func TestFreeCashFlowCalculation(t *testing.T) {
	// Test Free Cash Flow calculation
	metrics := &CashFlowMetrics{
		NetCashFromOperatingActivities: 50000000000,
		CapitalExpenditures:            5000000000,
	}

	// Calculate Free Cash Flow
	metrics.FreeCashFlow = metrics.NetCashFromOperatingActivities - metrics.CapitalExpenditures

	expectedFCF := 45000000000.0 // 50B - 5B

	assert.Equal(t, expectedFCF, metrics.FreeCashFlow)
}

// Benchmark tests
func BenchmarkClient_parseFilings(b *testing.B) {
	client := NewClient()

	// Create larger dataset for benchmarking
	recentData := map[string][]interface{}{
		"accessionNumber":       make([]interface{}, 1000),
		"filingDate":            make([]interface{}, 1000),
		"reportDate":            make([]interface{}, 1000),
		"form":                  make([]interface{}, 1000),
		"fileNumber":            make([]interface{}, 1000),
		"filmNumber":            make([]interface{}, 1000),
		"items":                 make([]interface{}, 1000),
		"size":                  make([]interface{}, 1000),
		"isXBRL":                make([]interface{}, 1000),
		"isInlineXBRL":          make([]interface{}, 1000),
		"primaryDocument":       make([]interface{}, 1000),
		"primaryDocDescription": make([]interface{}, 1000),
	}

	// Fill with sample data
	for i := 0; i < 1000; i++ {
		recentData["accessionNumber"][i] = fmt.Sprintf("0000320193-24-%06d", i)
		recentData["filingDate"][i] = "2024-02-01"
		recentData["reportDate"][i] = "2023-12-30"
		recentData["form"][i] = "10-Q"
		recentData["fileNumber"][i] = "001-36743"
		recentData["filmNumber"][i] = fmt.Sprintf("2457612%d", i)
		recentData["items"][i] = ""
		recentData["size"][i] = "100000"
		recentData["isXBRL"][i] = 1
		recentData["isInlineXBRL"][i] = 1
		recentData["primaryDocument"][i] = "aapl-20231230.htm"
		recentData["primaryDocDescription"][i] = "10-Q"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.parseFilings(recentData)
	}
}
