package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/natedogg/edgar/pkg/edgar"
)

// MockClient provides a mock implementation of the EDGAR client for testing
type MockClient struct {
	CompanyFactsResponse   *edgar.CompanyFacts
	CompanySubmissionsResp *edgar.CompanySubmissions
	CashFlowMetricsResp    *edgar.CashFlowMetrics
	EBITDAMetricsResp      *edgar.EBITDAMetrics
	QuarterlyCashFlowResp  *edgar.QuarterlyCashFlowAnalysis
	QuarterlyEBITDAResp    *edgar.QuarterlyEBITDAAnalysis
	FilingsResp            []edgar.Filing
	ErrorToReturn          error
}

// GetCompanyFacts returns the mocked company facts response
func (m *MockClient) GetCompanyFacts(cik string) (*edgar.CompanyFacts, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.CompanyFactsResponse, nil
}

// GetCompanySubmissions returns the mocked company submissions response
func (m *MockClient) GetCompanySubmissions(cik string) (*edgar.CompanySubmissions, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.CompanySubmissionsResp, nil
}

// GetMostRecent10Q returns the first filing from the mocked filings response
func (m *MockClient) GetMostRecent10Q(cik string) (*edgar.Filing, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	if len(m.FilingsResp) == 0 {
		return nil, fmt.Errorf("no 10-Q filings found")
	}
	return &m.FilingsResp[0], nil
}

// GetMostRecent4TenQs returns up to 4 filings from the mocked filings response
func (m *MockClient) GetMostRecent4TenQs(cik string) ([]edgar.Filing, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}

	count := len(m.FilingsResp)
	if count > 4 {
		count = 4
	}

	return m.FilingsResp[:count], nil
}

// ParseCashFlowMetrics returns the mocked cash flow metrics response
func (m *MockClient) ParseCashFlowMetrics(cik string, filing *edgar.Filing) (*edgar.CashFlowMetrics, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.CashFlowMetricsResp, nil
}

// ParseEBITDAMetrics returns the mocked EBITDA metrics response
func (m *MockClient) ParseEBITDAMetrics(cik string, filing *edgar.Filing) (*edgar.EBITDAMetrics, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.EBITDAMetricsResp, nil
}

// GetQuarterlyCashFlowAnalysis returns the mocked quarterly cash flow analysis
func (m *MockClient) GetQuarterlyCashFlowAnalysis(cik string) (*edgar.QuarterlyCashFlowAnalysis, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.QuarterlyCashFlowResp, nil
}

// GetQuarterlyEBITDAAnalysis returns the mocked quarterly EBITDA analysis
func (m *MockClient) GetQuarterlyEBITDAAnalysis(cik string) (*edgar.QuarterlyEBITDAAnalysis, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.QuarterlyEBITDAResp, nil
}

// TestDataProvider provides commonly used test data
type TestDataProvider struct{}

// NewTestDataProvider creates a new test data provider
func NewTestDataProvider() *TestDataProvider {
	return &TestDataProvider{}
}

// GetMockCompanyFacts returns mock company facts data
func (p *TestDataProvider) GetMockCompanyFacts() *edgar.CompanyFacts {
	return &edgar.CompanyFacts{
		CIK:    "0000320193",
		Entity: "Apple Inc.",
		Facts: map[string]interface{}{
			"us-gaap": map[string]interface{}{
				"NetCashProvidedByUsedInOperatingActivities": map[string]interface{}{
					"units": map[string]interface{}{
						"USD": []interface{}{
							map[string]interface{}{
								"form": "10-Q",
								"val":  50000000000.0,
								"end":  "2023-12-30",
							},
						},
					},
				},
				"PaymentsToAcquirePropertyPlantAndEquipment": map[string]interface{}{
					"units": map[string]interface{}{
						"USD": []interface{}{
							map[string]interface{}{
								"form": "10-Q",
								"val":  5000000000.0,
								"end":  "2023-12-30",
							},
						},
					},
				},
				"Revenues": map[string]interface{}{
					"units": map[string]interface{}{
						"USD": []interface{}{
							map[string]interface{}{
								"form": "10-Q",
								"val":  100000000000.0,
								"end":  "2023-12-30",
							},
						},
					},
				},
				"NetIncomeLoss": map[string]interface{}{
					"units": map[string]interface{}{
						"USD": []interface{}{
							map[string]interface{}{
								"form": "10-Q",
								"val":  25000000000.0,
								"end":  "2023-12-30",
							},
						},
					},
				},
				"InterestExpense": map[string]interface{}{
					"units": map[string]interface{}{
						"USD": []interface{}{
							map[string]interface{}{
								"form": "10-Q",
								"val":  1000000000.0,
								"end":  "2023-12-30",
							},
						},
					},
				},
				"IncomeTaxExpenseBenefit": map[string]interface{}{
					"units": map[string]interface{}{
						"USD": []interface{}{
							map[string]interface{}{
								"form": "10-Q",
								"val":  3000000000.0,
								"end":  "2023-12-30",
							},
						},
					},
				},
				"DepreciationDepletionAndAmortization": map[string]interface{}{
					"units": map[string]interface{}{
						"USD": []interface{}{
							map[string]interface{}{
								"form": "10-Q",
								"val":  2000000000.0,
								"end":  "2023-12-30",
							},
						},
					},
				},
			},
		},
	}
}

// GetMockCompanySubmissions returns mock company submissions data
func (p *TestDataProvider) GetMockCompanySubmissions() *edgar.CompanySubmissions {
	return &edgar.CompanySubmissions{
		CIK:  "0000320193",
		Name: "Apple Inc.",
		Filings: struct {
			Recent map[string][]interface{} `json:"recent"`
			Files  []struct {
				Name        string `json:"name"`
				FilingCount int    `json:"filingCount"`
				FilingFrom  string `json:"filingFrom"`
				FilingTo    string `json:"filingTo"`
			} `json:"files"`
		}{
			Recent: map[string][]interface{}{
				"accessionNumber":       {"0000320193-24-000007", "0000320193-24-000006", "0000320193-24-000005", "0000320193-24-000004"},
				"filingDate":            {"2024-02-01", "2023-11-02", "2023-08-03", "2023-05-04"},
				"reportDate":            {"2023-12-30", "2023-09-30", "2023-06-30", "2023-03-31"},
				"form":                  {"10-Q", "10-Q", "10-Q", "10-Q"},
				"fileNumber":            {"001-36743", "001-36743", "001-36743", "001-36743"},
				"filmNumber":            {"24576126", "24576125", "24576124", "24576123"},
				"items":                 {"", "", "", ""},
				"size":                  {"100000", "100000", "100000", "100000"},
				"isXBRL":                {1, 1, 1, 1},
				"isInlineXBRL":          {1, 1, 1, 1},
				"primaryDocument":       {"aapl-20231230.htm", "aapl-20230930.htm", "aapl-20230630.htm", "aapl-20230331.htm"},
				"primaryDocDescription": {"10-Q", "10-Q", "10-Q", "10-Q"},
			},
		},
	}
}

// GetMockFilings returns mock filing data
func (p *TestDataProvider) GetMockFilings() []edgar.Filing {
	return []edgar.Filing{
		{
			AccessionNumber: "0000320193-24-000007",
			FilingDate:      "2024-02-01",
			ReportDate:      "2023-12-30",
			Form:            "10-Q",
			FileNumber:      "001-36743",
			FilmNumber:      "24576126",
			Items:           "",
			Size:            "100000",
			IsXBRL:          "1",
			IsInlineXBRL:    "1",
			PrimaryDocument: "aapl-20231230.htm",
			PrimaryDocDesc:  "10-Q",
		},
		{
			AccessionNumber: "0000320193-24-000006",
			FilingDate:      "2023-11-02",
			ReportDate:      "2023-09-30",
			Form:            "10-Q",
			FileNumber:      "001-36743",
			FilmNumber:      "24576125",
			Items:           "",
			Size:            "100000",
			IsXBRL:          "1",
			IsInlineXBRL:    "1",
			PrimaryDocument: "aapl-20230930.htm",
			PrimaryDocDesc:  "10-Q",
		},
		{
			AccessionNumber: "0000320193-24-000005",
			FilingDate:      "2023-08-03",
			ReportDate:      "2023-06-30",
			Form:            "10-Q",
			FileNumber:      "001-36743",
			FilmNumber:      "24576124",
			Items:           "",
			Size:            "100000",
			IsXBRL:          "1",
			IsInlineXBRL:    "1",
			PrimaryDocument: "aapl-20230630.htm",
			PrimaryDocDesc:  "10-Q",
		},
		{
			AccessionNumber: "0000320193-24-000004",
			FilingDate:      "2023-05-04",
			ReportDate:      "2023-03-31",
			Form:            "10-Q",
			FileNumber:      "001-36743",
			FilmNumber:      "24576123",
			Items:           "",
			Size:            "100000",
			IsXBRL:          "1",
			IsInlineXBRL:    "1",
			PrimaryDocument: "aapl-20230331.htm",
			PrimaryDocDesc:  "10-Q",
		},
	}
}

// GetMockCashFlowMetrics returns mock cash flow metrics
func (p *TestDataProvider) GetMockCashFlowMetrics() *edgar.CashFlowMetrics {
	return &edgar.CashFlowMetrics{
		CompanyName:                    "Apple Inc.",
		CIK:                            "0000320193",
		FilingDate:                     "2024-02-01",
		ReportDate:                     "2023-12-30",
		Form:                           "10-Q",
		AccessionNumber:                "0000320193-24-000007",
		NetCashFromOperatingActivities: 50000000000,
		CapitalExpenditures:            5000000000,
		FreeCashFlow:                   45000000000,
	}
}

// GetMockEBITDAMetrics returns mock EBITDA metrics
func (p *TestDataProvider) GetMockEBITDAMetrics() *edgar.EBITDAMetrics {
	return &edgar.EBITDAMetrics{
		CompanyName:                 "Apple Inc.",
		CIK:                         "0000320193",
		FilingDate:                  "2024-02-01",
		ReportDate:                  "2023-12-30",
		Form:                        "10-Q",
		AccessionNumber:             "0000320193-24-000007",
		Revenue:                     100000000000,
		NetIncome:                   25000000000,
		InterestExpense:             1000000000,
		IncomeTaxExpense:            3000000000,
		DepreciationAndAmortization: 2000000000,
		EBITDA:                      31000000000,
		EBITDAMargin:                31.0,
	}
}

// GetMockQuarterlyCashFlowAnalysis returns mock quarterly cash flow analysis
func (p *TestDataProvider) GetMockQuarterlyCashFlowAnalysis() *edgar.QuarterlyCashFlowAnalysis {
	filings := p.GetMockFilings()
	quarters := make([]edgar.CashFlowMetrics, len(filings))

	for i, filing := range filings {
		quarters[i] = edgar.CashFlowMetrics{
			CompanyName:                    "Apple Inc.",
			CIK:                            "0000320193",
			FilingDate:                     filing.FilingDate,
			ReportDate:                     filing.ReportDate,
			Form:                           "10-Q",
			AccessionNumber:                filing.AccessionNumber,
			NetCashFromOperatingActivities: 50000000000 - float64(i)*2000000000, // Decreasing trend
			CapitalExpenditures:            5000000000 + float64(i)*500000000,   // Increasing trend
			FreeCashFlow:                   45000000000 - float64(i)*2500000000, // Decreasing trend
		}
	}

	return &edgar.QuarterlyCashFlowAnalysis{
		CompanyName: "Apple Inc.",
		CIK:         "0000320193",
		Quarters:    quarters,
	}
}

// GetMockQuarterlyEBITDAAnalysis returns mock quarterly EBITDA analysis
func (p *TestDataProvider) GetMockQuarterlyEBITDAAnalysis() *edgar.QuarterlyEBITDAAnalysis {
	filings := p.GetMockFilings()
	quarters := make([]edgar.EBITDAMetrics, len(filings))

	for i, filing := range filings {
		revenue := 100000000000.0 + float64(i)*2000000000  // Increasing trend
		netIncome := 25000000000.0 + float64(i)*1000000000 // Increasing trend
		ebitda := 31000000000.0 + float64(i)*1500000000    // Increasing trend

		quarters[i] = edgar.EBITDAMetrics{
			CompanyName:                 "Apple Inc.",
			CIK:                         "0000320193",
			FilingDate:                  filing.FilingDate,
			ReportDate:                  filing.ReportDate,
			Form:                        "10-Q",
			AccessionNumber:             filing.AccessionNumber,
			Revenue:                     revenue,
			NetIncome:                   netIncome,
			InterestExpense:             1000000000,
			IncomeTaxExpense:            3000000000,
			DepreciationAndAmortization: 2000000000,
			EBITDA:                      ebitda,
			EBITDAMargin:                (ebitda / revenue) * 100,
		}
	}

	return &edgar.QuarterlyEBITDAAnalysis{
		CompanyName: "Apple Inc.",
		CIK:         "0000320193",
		Quarters:    quarters,
	}
}

// CreateMockServer creates an HTTP test server with predefined responses
func CreateMockServer(responses map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Find matching response
		for pattern, response := range responses {
			if contains(path, pattern) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, response)
				return
			}
		}

		// Default 404 response
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error": "not found"}`)
	}))
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetMockCompanyFactsJSON returns JSON representation of mock company facts
func (p *TestDataProvider) GetMockCompanyFactsJSON() string {
	facts := p.GetMockCompanyFacts()
	data, _ := json.MarshalIndent(facts, "", "  ")
	return string(data)
}

// GetMockCompanySubmissionsJSON returns JSON representation of mock company submissions
func (p *TestDataProvider) GetMockCompanySubmissionsJSON() string {
	submissions := p.GetMockCompanySubmissions()
	data, _ := json.MarshalIndent(submissions, "", "  ")
	return string(data)
}

// SetupMockClient creates a fully configured mock client for testing
func SetupMockClient() *MockClient {
	provider := NewTestDataProvider()

	return &MockClient{
		CompanyFactsResponse:   provider.GetMockCompanyFacts(),
		CompanySubmissionsResp: provider.GetMockCompanySubmissions(),
		CashFlowMetricsResp:    provider.GetMockCashFlowMetrics(),
		EBITDAMetricsResp:      provider.GetMockEBITDAMetrics(),
		QuarterlyCashFlowResp:  provider.GetMockQuarterlyCashFlowAnalysis(),
		QuarterlyEBITDAResp:    provider.GetMockQuarterlyEBITDAAnalysis(),
		FilingsResp:            provider.GetMockFilings(),
		ErrorToReturn:          nil,
	}
}
