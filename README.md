# SEC EDGAR Cash Flow Analysis Tool

A Go-based command-line tool that fetches the most recent 10-Q filing for a company from the SEC's EDGAR database and calculates Free Cash Flow (FCF) and EBITDA metrics.

## Features

- Fetches the most recent 10-Q filing for any public company using their CIK
- **NEW**: Quarterly analysis - Get the 4 most recent 10-Q filings and their cash flow metrics
- **NEW**: EBITDA calculation - Calculate EBITDA from income statement components
- Extracts cash flow metrics from SEC filings:
  - Net Cash from Operating Activities
  - Capital Expenditures
  - Free Cash Flow (calculated as Operating Cash Flow - Capital Expenditures)
- **NEW**: Extracts EBITDA components from SEC filings:
  - Net Income
  - Interest Expense
  - Income Tax Expense
  - Depreciation & Amortization
  - EBITDA (calculated as Net Income + Interest + Taxes + Depreciation & Amortization)
  - **EBITDA Margin** (calculated as EBITDA / Revenue × 100)
- Outputs results in both human-readable format and JSON
- **NEW**: Trend analysis showing changes between quarters

## Usage

```bash
# Build the program
go build -o bin/edgar ./cmd/edgar

# Get most recent 10-Q filing (cash flow analysis)
./bin/edgar -cik <CIK>

# Get 4 most recent 10-Q filings with quarterly cash flow analysis
./bin/edgar -cik <CIK> -quarterly

# Calculate EBITDA for most recent 10-Q filing
./bin/edgar -cik <CIK> -ebitda

# Calculate EBITDA for 4 most recent 10-Q filings with quarterly analysis
./bin/edgar -cik <CIK> -ebitda-quarterly

# Examples:
./bin/edgar -cik 320193                    # Apple Inc. - single quarter cash flow
./bin/edgar -cik 320193 -quarterly         # Apple Inc. - 4 quarters cash flow
./bin/edgar -cik 320193 -ebitda            # Apple Inc. - single quarter EBITDA
./bin/edgar -cik 320193 -ebitda-quarterly  # Apple Inc. - 4 quarters EBITDA
./bin/edgar -cik 789019 -ebitda-quarterly  # Microsoft Corporation - 4 quarters EBITDA
```

## Command Line Options

- `-cik <CIK>`: Company CIK (Central Index Key) - **required**
- `-quarterly`: Get 4 most recent 10-Q filings and their cash flow metrics (optional)
- `-ebitda`: Calculate EBITDA for the most recent 10-Q filing (optional)
- `-ebitda-quarterly`: Calculate EBITDA for the 4 most recent 10-Q filings (optional)

## How to Find a Company's CIK

You can find a company's CIK (Central Index Key) by:
1. Going to the SEC's EDGAR database: https://www.sec.gov/edgar/searchedgar/companysearch.html
2. Searching for the company name
3. The CIK will be displayed in the search results

## Example Output

### Single Quarter Cash Flow Analysis
```
Cash Flow Analysis for Apple Inc.
=====================================
CIK: 320193
Form: 10-Q
Filing Date: 2025-05-02
Report Date: 2025-03-29
Accession Number: 0000320193-25-000057

Cash Flow Metrics:
------------------
Net Cash from Operating Activities: $53,887,000,000.00
Capital Expenditures: $6,011,000,000.00
Free Cash Flow (FCF): $47,876,000,000.00
```

### Single Quarter EBITDA Analysis
```
EBITDA Analysis for Apple Inc.
=====================================
CIK: 320193
Form: 10-Q
Filing Date: 2025-05-02
Report Date: 2025-03-29
Accession Number: 0000320193-25-000057

EBITDA Components:
------------------
Revenue: $265,595,000,000.00
Net Income: $61,110,000,000.00
Interest Expense: $2,931,000,000.00
Income Tax Expense: $10,784,000,000.00
Depreciation & Amortization: $5,741,000,000.00
EBITDA: $80,566,000,000.00
EBITDA Margin: 30.33%
```

### Quarterly EBITDA Analysis (4 Quarters)
```
Quarterly EBITDA Analysis for MICROSOFT CORPORATION
==========================================
CIK: 789019
Number of quarters analyzed: 4

Quarter 1:
----------
  Filing Date: 2025-04-30
  Report Date: 2025-03-31
  Revenue: $61,900,000,000.00
  Net Income: $21,900,000,000.00
  Interest Expense: $700,000,000.00
  Income Tax Expense: $4,300,000,000.00
  Depreciation & Amortization: $3,800,000,000.00
  EBITDA: $30,700,000,000.00
  EBITDA Margin: 49.60%

[... additional quarters ...]

Trends (Quarter 1 vs Quarter 4):
----------------------------------
  EBITDA Change: $2,500,000,000.00 (8.86%)
  Net Income Change: $1,200,000,000.00 (5.80%)
  Revenue Change: $3,200,000,000.00 (5.45%)
  EBITDA Margin Change: 47.20% to 49.60% (2.40 percentage points)
```

## EBITDA Calculation Method

The tool calculates EBITDA using the standard formula:
**EBITDA = Net Income + Interest Expense + Income Tax Expense + Depreciation & Amortization**

**EBITDA Margin** is calculated as:
**EBITDA Margin = (EBITDA / Revenue) × 100**

EBITDA Margin represents the proportion of a company's revenue that remains after accounting for operating expenses (excluding interest, taxes, depreciation, and amortization). It's a key profitability metric that shows how efficiently a company generates earnings from its operations.

### EBITDA Margin Interpretation:
- **High EBITDA Margin (>20%)**: Indicates strong operational efficiency and pricing power
- **Moderate EBITDA Margin (10-20%)**: Typical for many industries, shows decent profitability
- **Low EBITDA Margin (<10%)**: May indicate competitive pressure or operational challenges
- **Industry Comparison**: EBITDA margins vary significantly by industry (tech companies typically have higher margins than retail or manufacturing)

The tool automatically searches for these components in the SEC filings using multiple possible GAAP tag names to ensure maximum compatibility across different companies and reporting formats.

### EBITDA Components Extracted:
- **Revenue**: Revenues, RevenueFromContractWithCustomerExcludingAssessedTax, SalesRevenueNet
- **Net Income**: NetIncomeLoss, ProfitLoss, NetIncomeLossAvailableToCommonStockholdersBasic
- **Interest Expense**: InterestExpense, InterestExpenseDebt, InterestAndDebtExpense
- **Income Tax Expense**: IncomeTaxExpenseBenefit, ProvisionForIncomeTaxes, IncomeTaxesPaid
- **Depreciation & Amortization**: DepreciationDepletionAndAmortization, DepreciationAndAmortization, AmortizationOfIntangibleAssets

## Requirements

- Go 1.23.5 or later
- Internet connection to access SEC EDGAR API

## API Compliance

This tool uses the SEC's official EDGAR API and follows their guidelines:
- Includes proper User-Agent headers
- Handles gzip compression
- Respects rate limits

## License

This project is open source and available under the MIT License. 