package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/natedogg/edgar/pkg/edgar"
)

func main() {
	// Define command line flags
	var cik string
	var quarterly bool
	var ebitda bool
	var ebitdaQuarterly bool
	flag.StringVar(&cik, "cik", "", "Company CIK (Central Index Key) - required")
	flag.BoolVar(&quarterly, "quarterly", false, "Get 4 most recent 10-Q filings and their cash flow metrics")
	flag.BoolVar(&ebitda, "ebitda", false, "Calculate EBITDA for the most recent 10-Q filing")
	flag.BoolVar(&ebitdaQuarterly, "ebitda-quarterly", false, "Calculate EBITDA for the 4 most recent 10-Q filings")
	flag.Parse()

	// Validate required flag
	if cik == "" {
		fmt.Fprintf(os.Stderr, "Error: CIK is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -cik <CIK> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -quarterly          Get 4 most recent 10-Q cash flow metrics\n")
		fmt.Fprintf(os.Stderr, "  -ebitda            Calculate EBITDA for most recent 10-Q\n")
		fmt.Fprintf(os.Stderr, "  -ebitda-quarterly  Calculate EBITDA for 4 most recent 10-Q filings\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -cik 0000320193\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -cik 0000320193 -quarterly\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -cik 0000320193 -ebitda\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -cik 0000320193 -ebitda-quarterly\n", os.Args[0])
		os.Exit(1)
	}

	// Pad CIK with leading zeros if needed (SEC expects 10 digits)
	if len(cik) < 10 {
		cik = fmt.Sprintf("%010s", cik)
	}

	client := edgar.NewClient()

	if ebitdaQuarterly {
		// Get quarterly EBITDA analysis for 4 most recent 10-Q filings
		fmt.Printf("Fetching 4 most recent 10-Q filings and EBITDA metrics for CIK: %s\n", cik)

		analysis, err := client.GetQuarterlyEBITDAAnalysis(cik)
		if err != nil {
			log.Fatalf("Error getting quarterly EBITDA analysis: %v", err)
		}

		// Display the results
		fmt.Printf("\nQuarterly EBITDA Analysis for %s\n", analysis.CompanyName)
		fmt.Printf("==========================================\n")
		fmt.Printf("CIK: %s\n", analysis.CIK)
		fmt.Printf("Number of quarters analyzed: %d\n\n", len(analysis.Quarters))

		for i, quarter := range analysis.Quarters {
			fmt.Printf("Quarter %d:\n", i+1)
			fmt.Printf("----------\n")
			fmt.Printf("  Filing Date: %s\n", quarter.FilingDate)
			fmt.Printf("  Report Date: %s\n", quarter.ReportDate)
			fmt.Printf("  Accession Number: %s\n", quarter.AccessionNumber)
			fmt.Printf("  Revenue: $%.2f\n", quarter.Revenue)
			fmt.Printf("  Net Income: $%.2f\n", quarter.NetIncome)
			fmt.Printf("  Interest Expense: $%.2f\n", quarter.InterestExpense)
			fmt.Printf("  Income Tax Expense: $%.2f\n", quarter.IncomeTaxExpense)
			fmt.Printf("  Depreciation & Amortization: $%.2f\n", quarter.DepreciationAndAmortization)
			fmt.Printf("  EBITDA: $%.2f\n", quarter.EBITDA)
			fmt.Printf("  EBITDA Margin: %.2f%%\n", quarter.EBITDAMargin)
			fmt.Println()
		}

		// Calculate and display trends
		if len(analysis.Quarters) > 1 {
			fmt.Printf("Trends (Quarter 1 vs Quarter %d):\n", len(analysis.Quarters))
			fmt.Printf("----------------------------------\n")
			latest := analysis.Quarters[0]
			oldest := analysis.Quarters[len(analysis.Quarters)-1]

			ebitdaChange := latest.EBITDA - oldest.EBITDA
			ebitdaChangePercent := (ebitdaChange / oldest.EBITDA) * 100

			fmt.Printf("  EBITDA Change: $%.2f (%.2f%%)\n", ebitdaChange, ebitdaChangePercent)

			netIncomeChange := latest.NetIncome - oldest.NetIncome
			netIncomeChangePercent := (netIncomeChange / oldest.NetIncome) * 100

			fmt.Printf("  Net Income Change: $%.2f (%.2f%%)\n", netIncomeChange, netIncomeChangePercent)

			revenueChange := latest.Revenue - oldest.Revenue
			revenueChangePercent := (revenueChange / oldest.Revenue) * 100

			fmt.Printf("  Revenue Change: $%.2f (%.2f%%)\n", revenueChange, revenueChangePercent)

			marginChange := latest.EBITDAMargin - oldest.EBITDAMargin

			fmt.Printf("  EBITDA Margin Change: %.2f%% to %.2f%% (%.2f percentage points)\n",
				oldest.EBITDAMargin, latest.EBITDAMargin, marginChange)
			fmt.Println()
		}

		// Also output as JSON for programmatic use
		fmt.Println("JSON Output:")
		fmt.Println("============")
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(analysis); err != nil {
			log.Fatalf("Error encoding JSON response: %v", err)
		}

	} else if ebitda {
		// Single EBITDA analysis
		fmt.Printf("Fetching most recent 10-Q filing and calculating EBITDA for CIK: %s\n", cik)

		// Get the most recent 10-Q filing
		filing, err := client.GetMostRecent10Q(cik)
		if err != nil {
			log.Fatalf("Error getting most recent 10-Q filing: %v", err)
		}

		fmt.Printf("Found 10-Q filing:\n")
		fmt.Printf("  Accession Number: %s\n", filing.AccessionNumber)
		fmt.Printf("  Filing Date: %s\n", filing.FilingDate)
		fmt.Printf("  Report Date: %s\n", filing.ReportDate)
		fmt.Printf("  Primary Document: %s\n", filing.PrimaryDocument)
		fmt.Println()

		// Parse EBITDA metrics from the filing
		fmt.Println("Calculating EBITDA...")
		metrics, err := client.ParseEBITDAMetrics(cik, filing)
		if err != nil {
			log.Fatalf("Error parsing EBITDA metrics: %v", err)
		}

		// Display the results
		fmt.Printf("EBITDA Analysis for %s\n", metrics.CompanyName)
		fmt.Printf("=====================================\n")
		fmt.Printf("CIK: %s\n", metrics.CIK)
		fmt.Printf("Form: %s\n", metrics.Form)
		fmt.Printf("Filing Date: %s\n", metrics.FilingDate)
		fmt.Printf("Report Date: %s\n", metrics.ReportDate)
		fmt.Printf("Accession Number: %s\n", metrics.AccessionNumber)
		fmt.Println()

		fmt.Printf("EBITDA Components:\n")
		fmt.Printf("------------------\n")
		fmt.Printf("Revenue: $%.2f\n", metrics.Revenue)
		fmt.Printf("Net Income: $%.2f\n", metrics.NetIncome)
		fmt.Printf("Interest Expense: $%.2f\n", metrics.InterestExpense)
		fmt.Printf("Income Tax Expense: $%.2f\n", metrics.IncomeTaxExpense)
		fmt.Printf("Depreciation & Amortization: $%.2f\n", metrics.DepreciationAndAmortization)
		fmt.Printf("EBITDA: $%.2f\n", metrics.EBITDA)
		fmt.Printf("EBITDA Margin: %.2f%%\n", metrics.EBITDAMargin)
		fmt.Println()

		// Also output as JSON for programmatic use
		fmt.Println("JSON Output:")
		fmt.Println("============")
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(metrics); err != nil {
			log.Fatalf("Error encoding JSON response: %v", err)
		}

	} else if quarterly {
		// Get quarterly analysis for 4 most recent 10-Q filings
		fmt.Printf("Fetching 4 most recent 10-Q filings and cash flow metrics for CIK: %s\n", cik)

		analysis, err := client.GetQuarterlyCashFlowAnalysis(cik)
		if err != nil {
			log.Fatalf("Error getting quarterly cash flow analysis: %v", err)
		}

		// Display the results
		fmt.Printf("\nQuarterly Cash Flow Analysis for %s\n", analysis.CompanyName)
		fmt.Printf("===============================================\n")
		fmt.Printf("CIK: %s\n", analysis.CIK)
		fmt.Printf("Number of quarters analyzed: %d\n\n", len(analysis.Quarters))

		for i, quarter := range analysis.Quarters {
			fmt.Printf("Quarter %d:\n", i+1)
			fmt.Printf("----------\n")
			fmt.Printf("  Filing Date: %s\n", quarter.FilingDate)
			fmt.Printf("  Report Date: %s\n", quarter.ReportDate)
			fmt.Printf("  Accession Number: %s\n", quarter.AccessionNumber)
			fmt.Printf("  Net Cash from Operating Activities: $%.2f\n", quarter.NetCashFromOperatingActivities)
			fmt.Printf("  Capital Expenditures: $%.2f\n", quarter.CapitalExpenditures)
			fmt.Printf("  Free Cash Flow (FCF): $%.2f\n", quarter.FreeCashFlow)
			fmt.Println()
		}

		// Calculate and display trends
		if len(analysis.Quarters) > 1 {
			fmt.Printf("Trends (Quarter 1 vs Quarter %d):\n", len(analysis.Quarters))
			fmt.Printf("----------------------------------\n")
			latest := analysis.Quarters[0]
			oldest := analysis.Quarters[len(analysis.Quarters)-1]

			fcfChange := latest.FreeCashFlow - oldest.FreeCashFlow
			fcfChangePercent := (fcfChange / oldest.FreeCashFlow) * 100

			fmt.Printf("  Free Cash Flow Change: $%.2f (%.2f%%)\n", fcfChange, fcfChangePercent)

			opCashChange := latest.NetCashFromOperatingActivities - oldest.NetCashFromOperatingActivities
			opCashChangePercent := (opCashChange / oldest.NetCashFromOperatingActivities) * 100

			fmt.Printf("  Operating Cash Flow Change: $%.2f (%.2f%%)\n", opCashChange, opCashChangePercent)

			capexChange := latest.CapitalExpenditures - oldest.CapitalExpenditures
			capexChangePercent := (capexChange / oldest.CapitalExpenditures) * 100

			fmt.Printf("  Capital Expenditures Change: $%.2f (%.2f%%)\n", capexChange, capexChangePercent)
			fmt.Println()
		}

		// Also output as JSON for programmatic use
		fmt.Println("JSON Output:")
		fmt.Println("============")
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(analysis); err != nil {
			log.Fatalf("Error encoding JSON response: %v", err)
		}

	} else {
		// Original single 10-Q analysis
		fmt.Printf("Fetching most recent 10-Q filing for CIK: %s\n", cik)

		// Get the most recent 10-Q filing
		filing, err := client.GetMostRecent10Q(cik)
		if err != nil {
			log.Fatalf("Error getting most recent 10-Q filing: %v", err)
		}

		fmt.Printf("Found 10-Q filing:\n")
		fmt.Printf("  Accession Number: %s\n", filing.AccessionNumber)
		fmt.Printf("  Filing Date: %s\n", filing.FilingDate)
		fmt.Printf("  Report Date: %s\n", filing.ReportDate)
		fmt.Printf("  Primary Document: %s\n", filing.PrimaryDocument)
		fmt.Println()

		// Parse cash flow metrics from the filing
		fmt.Println("Parsing cash flow metrics...")
		metrics, err := client.ParseCashFlowMetrics(cik, filing)
		if err != nil {
			log.Fatalf("Error parsing cash flow metrics: %v", err)
		}

		// Display the results
		fmt.Printf("Cash Flow Analysis for %s\n", metrics.CompanyName)
		fmt.Printf("=====================================\n")
		fmt.Printf("CIK: %s\n", metrics.CIK)
		fmt.Printf("Form: %s\n", metrics.Form)
		fmt.Printf("Filing Date: %s\n", metrics.FilingDate)
		fmt.Printf("Report Date: %s\n", metrics.ReportDate)
		fmt.Printf("Accession Number: %s\n", metrics.AccessionNumber)
		fmt.Println()

		fmt.Printf("Cash Flow Metrics:\n")
		fmt.Printf("------------------\n")
		fmt.Printf("Net Cash from Operating Activities: $%.2f\n", metrics.NetCashFromOperatingActivities)
		fmt.Printf("Capital Expenditures: $%.2f\n", metrics.CapitalExpenditures)
		fmt.Printf("Free Cash Flow (FCF): $%.2f\n", metrics.FreeCashFlow)
		fmt.Println()

		// Also output as JSON for programmatic use
		fmt.Println("JSON Output:")
		fmt.Println("============")
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(metrics); err != nil {
			log.Fatalf("Error encoding JSON response: %v", err)
		}
	}
}
