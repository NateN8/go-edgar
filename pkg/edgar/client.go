package edgar

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL   = "https://data.sec.gov"
	userAgent = "Your Company Name yourname@example.com" // Replace with your details
)

// Client represents an EDGAR API client
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new EDGAR API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
		userAgent: userAgent,
	}
}

// makeRequest is a helper function to make HTTP requests with proper headers and gzip handling
func (c *Client) makeRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Host", "data.sec.gov")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() // Ignoring close error

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var reader io.Reader = resp.Body

	// Check if response is gzip compressed
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error creating gzip reader: %w", err)
		}
		defer func() { _ = gzipReader.Close() }() // Ignoring close error
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

// CompanyFacts represents the company facts response
type CompanyFacts struct {
	// Add fields based on the API response structure
	CIK    interface{}            `json:"cik"` // Can be string or number
	Entity string                 `json:"entityName"`
	Facts  map[string]interface{} `json:"facts"`
}

// GetCIKString returns the CIK as a string
func (cf *CompanyFacts) GetCIKString() string {
	switch v := cf.CIK.(type) {
	case string:
		return v
	case json.Number:
		return string(v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// CompanySubmissions represents the company submissions response
type CompanySubmissions struct {
	CIK                      string   `json:"cik"`
	EntityType               string   `json:"entityType"`
	SIC                      string   `json:"sic"`
	SICDesc                  string   `json:"sicDescription"`
	Name                     string   `json:"name"`
	Tickers                  []string `json:"tickers"`
	Exchanges                []string `json:"exchanges"`
	Ein                      string   `json:"ein"`
	Description              string   `json:"description"`
	Website                  string   `json:"website"`
	InvestorWebsite          string   `json:"investorWebsite"`
	Category                 string   `json:"category"`
	FiscalYearEnd            string   `json:"fiscalYearEnd"`
	StateOfIncorporation     string   `json:"stateOfIncorporation"`
	StateOfIncorporationDesc string   `json:"stateOfIncorporationDescription"`
	Addresses                struct {
		Mailing  Address `json:"mailing"`
		Business Address `json:"business"`
	} `json:"addresses"`
	Phone       string `json:"phone"`
	Flags       string `json:"flags"`
	FormerNames []struct {
		Name string `json:"name"`
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"formerNames"`
	Filings struct {
		Recent map[string][]interface{} `json:"recent"`
		Files  []struct {
			Name        string `json:"name"`
			FilingCount int    `json:"filingCount"`
			FilingFrom  string `json:"filingFrom"`
			FilingTo    string `json:"filingTo"`
		} `json:"files"`
	} `json:"filings"`
}

type Address struct {
	Street1            string `json:"street1"`
	Street2            string `json:"street2"`
	City               string `json:"city"`
	StateOrCountry     string `json:"stateOrCountry"`
	ZipCode            string `json:"zipCode"`
	StateOrCountryDesc string `json:"stateOrCountryDescription"`
}

// Filing represents a single filing
type Filing struct {
	AccessionNumber string
	FilingDate      string
	ReportDate      string
	Form            string
	FileNumber      string
	FilmNumber      string
	Items           string
	Size            string
	IsXBRL          string
	IsInlineXBRL    string
	PrimaryDocument string
	PrimaryDocDesc  string
}

// CashFlowMetrics represents the parsed cash flow metrics
type CashFlowMetrics struct {
	CompanyName                    string  `json:"companyName"`
	CIK                            string  `json:"cik"`
	FilingDate                     string  `json:"filingDate"`
	ReportDate                     string  `json:"reportDate"`
	NetCashFromOperatingActivities float64 `json:"netCashFromOperatingActivities"`
	CapitalExpenditures            float64 `json:"capitalExpenditures"`
	FreeCashFlow                   float64 `json:"freeCashFlow"`
	Form                           string  `json:"form"`
	AccessionNumber                string  `json:"accessionNumber"`
}

// QuarterlyCashFlowAnalysis represents cash flow metrics for multiple quarters
type QuarterlyCashFlowAnalysis struct {
	CompanyName string            `json:"companyName"`
	CIK         string            `json:"cik"`
	Quarters    []CashFlowMetrics `json:"quarters"`
}

// EBITDAMetrics represents the calculated EBITDA metrics
type EBITDAMetrics struct {
	CompanyName                 string  `json:"companyName"`
	CIK                         string  `json:"cik"`
	FilingDate                  string  `json:"filingDate"`
	ReportDate                  string  `json:"reportDate"`
	Form                        string  `json:"form"`
	AccessionNumber             string  `json:"accessionNumber"`
	Revenue                     float64 `json:"revenue"`
	NetIncome                   float64 `json:"netIncome"`
	InterestExpense             float64 `json:"interestExpense"`
	IncomeTaxExpense            float64 `json:"incomeTaxExpense"`
	DepreciationAndAmortization float64 `json:"depreciationAndAmortization"`
	EBITDA                      float64 `json:"ebitda"`
	EBITDAMargin                float64 `json:"ebitdaMargin"` // EBITDA / Revenue as percentage
}

// QuarterlyEBITDAAnalysis represents EBITDA metrics for multiple quarters
type QuarterlyEBITDAAnalysis struct {
	CompanyName string          `json:"companyName"`
	CIK         string          `json:"cik"`
	Quarters    []EBITDAMetrics `json:"quarters"`
}

// GetCompanyFacts retrieves company facts for a given CIK
func (c *Client) GetCompanyFacts(cik string) (*CompanyFacts, error) {
	url := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", baseURL, cik)

	body, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var facts CompanyFacts
	if err := json.Unmarshal(body, &facts); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &facts, nil
}

// GetCompanySubmissions retrieves company submissions for a given CIK
func (c *Client) GetCompanySubmissions(cik string) (*CompanySubmissions, error) {
	url := fmt.Sprintf("%s/submissions/CIK%s.json", baseURL, cik)

	body, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var submissions CompanySubmissions
	if err := json.Unmarshal(body, &submissions); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &submissions, nil
}

// GetMostRecent10Q finds the most recent 10-Q filing from company submissions
func (c *Client) GetMostRecent10Q(cik string) (*Filing, error) {
	submissions, err := c.GetCompanySubmissions(cik)
	if err != nil {
		return nil, fmt.Errorf("error getting company submissions: %w", err)
	}

	// Parse recent filings
	filings := c.parseFilings(submissions.Filings.Recent)

	// Filter for 10-Q filings and sort by filing date (most recent first)
	var tenQFilings []Filing
	for _, filing := range filings {
		if filing.Form == "10-Q" {
			tenQFilings = append(tenQFilings, filing)
		}
	}

	if len(tenQFilings) == 0 {
		return nil, fmt.Errorf("no 10-Q filings found for CIK %s", cik)
	}

	// Sort by filing date (most recent first)
	sort.Slice(tenQFilings, func(i, j int) bool {
		return tenQFilings[i].FilingDate > tenQFilings[j].FilingDate
	})

	return &tenQFilings[0], nil
}

// parseFilings converts the submissions recent filings map to Filing structs
func (c *Client) parseFilings(recent map[string][]interface{}) []Filing {
	var filings []Filing

	// Get the length of arrays (should be the same for all)
	if len(recent["accessionNumber"]) == 0 {
		return filings
	}

	count := len(recent["accessionNumber"])
	for i := 0; i < count; i++ {
		filing := Filing{
			AccessionNumber: c.toString(recent["accessionNumber"][i]),
			FilingDate:      c.toString(recent["filingDate"][i]),
			ReportDate:      c.toString(recent["reportDate"][i]),
			Form:            c.toString(recent["form"][i]),
			FileNumber:      c.toString(recent["fileNumber"][i]),
			FilmNumber:      c.toString(recent["filmNumber"][i]),
			Items:           c.toString(recent["items"][i]),
			Size:            c.toString(recent["size"][i]),
			IsXBRL:          c.toString(recent["isXBRL"][i]),
			IsInlineXBRL:    c.toString(recent["isInlineXBRL"][i]),
			PrimaryDocument: c.toString(recent["primaryDocument"][i]),
			PrimaryDocDesc:  c.toString(recent["primaryDocDescription"][i]),
		}
		filings = append(filings, filing)
	}

	return filings
}

// toString safely converts interface{} to string
func (c *Client) toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case float64:
		return strconv.FormatFloat(val, 'f', 0, 64)
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// GetMostRecent4TenQs finds the 4 most recent 10-Q filings from company submissions
func (c *Client) GetMostRecent4TenQs(cik string) ([]Filing, error) {
	submissions, err := c.GetCompanySubmissions(cik)
	if err != nil {
		return nil, fmt.Errorf("error getting company submissions: %w", err)
	}

	// Parse recent filings
	filings := c.parseFilings(submissions.Filings.Recent)

	// Filter for 10-Q filings and sort by filing date (most recent first)
	var tenQFilings []Filing
	for _, filing := range filings {
		if filing.Form == "10-Q" {
			tenQFilings = append(tenQFilings, filing)
		}
	}

	if len(tenQFilings) == 0 {
		return nil, fmt.Errorf("no 10-Q filings found for CIK %s", cik)
	}

	// Sort by filing date (most recent first)
	sort.Slice(tenQFilings, func(i, j int) bool {
		return tenQFilings[i].FilingDate > tenQFilings[j].FilingDate
	})

	// Return up to 4 most recent filings
	count := len(tenQFilings)
	if count > 4 {
		count = 4
	}

	return tenQFilings[:count], nil
}

// GetQuarterlyCashFlowAnalysis retrieves cash flow metrics for the 4 most recent 10-Q filings
func (c *Client) GetQuarterlyCashFlowAnalysis(cik string) (*QuarterlyCashFlowAnalysis, error) {
	// Get the 4 most recent 10-Q filings
	filings, err := c.GetMostRecent4TenQs(cik)
	if err != nil {
		return nil, fmt.Errorf("error getting recent 10-Q filings: %w", err)
	}

	// Get company facts once (we'll reuse this for all quarters)
	facts, err := c.GetCompanyFacts(cik)
	if err != nil {
		return nil, fmt.Errorf("error getting company facts: %w", err)
	}

	analysis := &QuarterlyCashFlowAnalysis{
		CompanyName: facts.Entity,
		CIK:         facts.GetCIKString(),
		Quarters:    make([]CashFlowMetrics, 0, len(filings)),
	}

	// Parse cash flow metrics for each filing
	for _, filing := range filings {
		metrics, err := c.ParseCashFlowMetricsFromFacts(facts, &filing)
		if err != nil {
			log.Printf("Warning: Could not parse cash flow metrics for filing %s: %v", filing.AccessionNumber, err)
			continue
		}
		analysis.Quarters = append(analysis.Quarters, *metrics)
	}

	if len(analysis.Quarters) == 0 {
		return nil, fmt.Errorf("no cash flow metrics could be extracted from any 10-Q filings")
	}

	return analysis, nil
}

// ParseCashFlowMetrics extracts cash flow metrics from a 10-Q filing
func (c *Client) ParseCashFlowMetrics(cik string, filing *Filing) (*CashFlowMetrics, error) {
	// Get company facts which contain the financial data
	facts, err := c.GetCompanyFacts(cik)
	if err != nil {
		return nil, fmt.Errorf("error getting company facts: %w", err)
	}

	metrics := &CashFlowMetrics{
		CompanyName:     facts.Entity,
		CIK:             facts.GetCIKString(),
		FilingDate:      filing.FilingDate,
		ReportDate:      filing.ReportDate,
		Form:            filing.Form,
		AccessionNumber: filing.AccessionNumber,
	}

	// Extract cash flow metrics from facts
	if err := c.extractCashFlowData(facts, metrics, filing.ReportDate); err != nil {
		return nil, fmt.Errorf("error extracting cash flow data: %w", err)
	}

	// Calculate free cash flow
	metrics.FreeCashFlow = metrics.NetCashFromOperatingActivities - metrics.CapitalExpenditures

	return metrics, nil
}

// extractCashFlowData extracts specific cash flow values from company facts
func (c *Client) extractCashFlowData(facts *CompanyFacts, metrics *CashFlowMetrics, reportDate string) error {
	// Navigate through the facts structure to find cash flow data
	factsMap := facts.Facts
	if factsMap == nil {
		return fmt.Errorf("facts data is nil")
	}

	// Look for US-GAAP taxonomy
	usGaap, ok := factsMap["us-gaap"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("us-gaap taxonomy not found")
	}

	// Extract Net Cash from Operating Activities
	if err := c.extractMetric(usGaap, []string{
		"NetCashProvidedByUsedInOperatingActivities",
		"NetCashFromOperatingActivities",
		"CashProvidedByUsedInOperatingActivities",
	}, &metrics.NetCashFromOperatingActivities, reportDate); err != nil {
		log.Printf("Warning: Could not extract operating cash flow: %v", err)
	}

	// Extract Capital Expenditures
	if err := c.extractMetric(usGaap, []string{
		"PaymentsToAcquirePropertyPlantAndEquipment",
		"CapitalExpenditures",
		"PaymentsForPropertyPlantAndEquipment",
		"PaymentsToAcquireProductiveAssets",
	}, &metrics.CapitalExpenditures, reportDate); err != nil {
		log.Printf("Warning: Could not extract capital expenditures: %v", err)
	}

	return nil
}

// extractMetric tries to extract a metric value using multiple possible tag names
func (c *Client) extractMetric(usGaap map[string]interface{}, tagNames []string, result *float64, reportDate string) error {
	for _, tagName := range tagNames {
		if concept, ok := usGaap[tagName].(map[string]interface{}); ok {
			if units, ok := concept["units"].(map[string]interface{}); ok {
				// Try USD first, then other units
				for unitType, unitData := range units {
					if strings.Contains(strings.ToLower(unitType), "usd") {
						if dataArray, ok := unitData.([]interface{}); ok {
							// Find the most recent value for the report date
							value := c.findValueForDate(dataArray, reportDate)
							if value != 0 {
								*result = value
								return nil
							}
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("metric not found with any of the provided tag names: %v", tagNames)
}

// findValueForDate finds the value closest to the given report date
func (c *Client) findValueForDate(dataArray []interface{}, targetDate string) float64 {
	var bestValue float64
	var bestDate string
	var bestScore int // Higher score = better match

	for _, item := range dataArray {
		if dataPoint, ok := item.(map[string]interface{}); ok {
			if date, ok := dataPoint["end"].(string); ok {
				if form, ok := dataPoint["form"].(string); ok {
					// Calculate match score
					score := 0

					// Prefer exact date matches
					if date == targetDate {
						score += 100
					}

					// Prefer 10-Q forms for quarterly analysis
					switch form {
					case "10-Q":
						score += 50
					case "10-K":
						score += 10 // Lower priority for annual forms
					}

					// Prefer more recent dates if no exact match
					if date <= targetDate && date > bestDate {
						score += 25
					}

					// Only update if this is a better match
					if score > bestScore || (score == bestScore && date > bestDate) {
						if val, ok := dataPoint["val"].(float64); ok {
							bestValue = val
							bestDate = date
							bestScore = score
						} else if valStr, ok := dataPoint["val"].(string); ok {
							if val, err := strconv.ParseFloat(valStr, 64); err == nil {
								bestValue = val
								bestDate = date
								bestScore = score
							}
						}
					}
				}
			}
		}
	}

	return bestValue
}

// ParseCashFlowMetricsFromFacts extracts cash flow metrics using pre-fetched company facts
func (c *Client) ParseCashFlowMetricsFromFacts(facts *CompanyFacts, filing *Filing) (*CashFlowMetrics, error) {
	metrics := &CashFlowMetrics{
		CompanyName:     facts.Entity,
		CIK:             facts.GetCIKString(),
		FilingDate:      filing.FilingDate,
		ReportDate:      filing.ReportDate,
		Form:            filing.Form,
		AccessionNumber: filing.AccessionNumber,
	}

	// Extract cash flow metrics from facts
	if err := c.extractCashFlowData(facts, metrics, filing.ReportDate); err != nil {
		return nil, fmt.Errorf("error extracting cash flow data: %w", err)
	}

	// Calculate free cash flow
	metrics.FreeCashFlow = metrics.NetCashFromOperatingActivities - metrics.CapitalExpenditures

	return metrics, nil
}

// CompanyConcept represents a specific concept for a company
type CompanyConcept struct {
	CIK      string `json:"cik"`
	Taxonomy string `json:"taxonomy"`
	Tag      string `json:"tag"`
	Label    string `json:"label"`
	Units    map[string][]struct {
		Form  string  `json:"form"`
		Date  string  `json:"date"`
		Value float64 `json:"val"`
	} `json:"units"`
}

// GetCompanyConcept retrieves a specific concept for a company
func (c *Client) GetCompanyConcept(cik, taxonomy, tag string) (*CompanyConcept, error) {
	url := fmt.Sprintf("%s/api/xbrl/companyconcept/CIK%s/%s/%s.json", baseURL, cik, taxonomy, tag)

	body, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var concept CompanyConcept
	if err := json.Unmarshal(body, &concept); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &concept, nil
}

// ParseEBITDAMetrics extracts EBITDA components from a 10-Q filing
func (c *Client) ParseEBITDAMetrics(cik string, filing *Filing) (*EBITDAMetrics, error) {
	// Get company facts which contain the financial data
	facts, err := c.GetCompanyFacts(cik)
	if err != nil {
		return nil, fmt.Errorf("error getting company facts: %w", err)
	}

	return c.ParseEBITDAMetricsFromFacts(facts, filing)
}

// ParseEBITDAMetricsFromFacts extracts EBITDA components using pre-fetched company facts
func (c *Client) ParseEBITDAMetricsFromFacts(facts *CompanyFacts, filing *Filing) (*EBITDAMetrics, error) {
	metrics := &EBITDAMetrics{
		CompanyName:     facts.Entity,
		CIK:             facts.GetCIKString(),
		FilingDate:      filing.FilingDate,
		ReportDate:      filing.ReportDate,
		Form:            filing.Form,
		AccessionNumber: filing.AccessionNumber,
	}

	// Extract EBITDA components from facts
	if err := c.extractEBITDAData(facts, metrics, filing.ReportDate); err != nil {
		return nil, fmt.Errorf("error extracting EBITDA data: %w", err)
	}

	// Calculate EBITDA
	metrics.EBITDA = metrics.NetIncome + metrics.InterestExpense + metrics.IncomeTaxExpense + metrics.DepreciationAndAmortization

	// Calculate EBITDA Margin (as percentage)
	if metrics.Revenue != 0 {
		metrics.EBITDAMargin = (metrics.EBITDA / metrics.Revenue) * 100
	} else {
		log.Printf("Warning: Revenue is zero, cannot calculate EBITDA margin")
		metrics.EBITDAMargin = 0
	}

	return metrics, nil
}

// extractEBITDAData extracts specific EBITDA components from company facts
func (c *Client) extractEBITDAData(facts *CompanyFacts, metrics *EBITDAMetrics, reportDate string) error {
	// Navigate through the facts structure to find financial data
	factsMap := facts.Facts
	if factsMap == nil {
		return fmt.Errorf("facts data is nil")
	}

	// Look for US-GAAP taxonomy
	usGaap, ok := factsMap["us-gaap"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("us-gaap taxonomy not found")
	}

	// Extract Revenue
	if err := c.extractMetric(usGaap, []string{
		"Revenues",
		"RevenueFromContractWithCustomerExcludingAssessedTax",
		"SalesRevenueNet",
		"RevenueFromContractWithCustomerIncludingAssessedTax",
		"Revenue",
		"SalesRevenueGoodsNet",
		"RevenuesNetOfInterestExpense",
	}, &metrics.Revenue, reportDate); err != nil {
		log.Printf("Warning: Could not extract revenue: %v", err)
	}

	// Extract Net Income
	if err := c.extractMetric(usGaap, []string{
		"NetIncomeLoss",
		"ProfitLoss",
		"NetIncomeLossAvailableToCommonStockholdersBasic",
		"IncomeLossFromContinuingOperations",
	}, &metrics.NetIncome, reportDate); err != nil {
		log.Printf("Warning: Could not extract net income: %v", err)
	}

	// Extract Interest Expense
	if err := c.extractMetric(usGaap, []string{
		"InterestExpense",
		"InterestExpenseDebt",
		"InterestAndDebtExpense",
		"InterestExpenseNet",
	}, &metrics.InterestExpense, reportDate); err != nil {
		log.Printf("Warning: Could not extract interest expense: %v", err)
	}

	// Extract Income Tax Expense
	if err := c.extractMetric(usGaap, []string{
		"IncomeTaxExpenseBenefit",
		"ProvisionForIncomeTaxes",
		"IncomeTaxesPaid",
		"CurrentIncomeTaxExpenseBenefit",
	}, &metrics.IncomeTaxExpense, reportDate); err != nil {
		log.Printf("Warning: Could not extract income tax expense: %v", err)
	}

	// Extract Depreciation and Amortization
	// This is often found in cash flow statement or as a combined figure
	if err := c.extractMetric(usGaap, []string{
		"DepreciationDepletionAndAmortization",
		"Depreciation",
		"DepreciationAndAmortization",
		"AmortizationOfIntangibleAssets",
		"DepreciationAmortizationAndAccretionNet",
	}, &metrics.DepreciationAndAmortization, reportDate); err != nil {
		log.Printf("Warning: Could not extract depreciation and amortization: %v", err)

		// Try to get separate depreciation and amortization figures
		var depreciation, amortization float64
		if err1 := c.extractMetric(usGaap, []string{
			"Depreciation",
			"DepreciationNonproduction",
		}, &depreciation, reportDate); err1 == nil {
			metrics.DepreciationAndAmortization += depreciation
		}

		if err2 := c.extractMetric(usGaap, []string{
			"AmortizationOfIntangibleAssets",
			"Amortization",
		}, &amortization, reportDate); err2 == nil {
			metrics.DepreciationAndAmortization += amortization
		}
	}

	return nil
}

// GetQuarterlyEBITDAAnalysis retrieves EBITDA metrics for the 4 most recent 10-Q filings
func (c *Client) GetQuarterlyEBITDAAnalysis(cik string) (*QuarterlyEBITDAAnalysis, error) {
	// Get the 4 most recent 10-Q filings
	filings, err := c.GetMostRecent4TenQs(cik)
	if err != nil {
		return nil, fmt.Errorf("error getting recent 10-Q filings: %w", err)
	}

	// Get company facts once (we'll reuse this for all quarters)
	facts, err := c.GetCompanyFacts(cik)
	if err != nil {
		return nil, fmt.Errorf("error getting company facts: %w", err)
	}

	analysis := &QuarterlyEBITDAAnalysis{
		CompanyName: facts.Entity,
		CIK:         facts.GetCIKString(),
		Quarters:    make([]EBITDAMetrics, 0, len(filings)),
	}

	// Parse EBITDA metrics for each filing
	for _, filing := range filings {
		metrics, err := c.ParseEBITDAMetricsFromFacts(facts, &filing)
		if err != nil {
			log.Printf("Warning: Could not parse EBITDA metrics for filing %s: %v", filing.AccessionNumber, err)
			continue
		}
		analysis.Quarters = append(analysis.Quarters, *metrics)
	}

	if len(analysis.Quarters) == 0 {
		return nil, fmt.Errorf("no EBITDA metrics could be extracted from any 10-Q filings")
	}

	return analysis, nil
}
