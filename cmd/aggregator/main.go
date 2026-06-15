package main

import (
	"flag"
	"fmt"
	"os"

	"fv-sec001-go/internal/aggregator"
	"fv-sec001-go/internal/csvio"
)

func main() {
	input := flag.String("input", "", "path to input CSV file; required")
	output := flag.String("output", "results", "output directory")
	strict := flag.Bool("strict", false, "fail on the first malformed data row")
	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "error: --input is required")
		flag.Usage()
		os.Exit(2)
	}

	campaigns, summary, err := csvio.ProcessFile(*input, *strict)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		printInvalidSamples(summary)
		os.Exit(1)
	}

	ctr := aggregator.ResultRows(aggregator.TopCTR(campaigns, 10))
	cpa := aggregator.ResultRows(aggregator.TopCPA(campaigns, 10))
	if err := csvio.WriteResults(*output, ctr, cpa); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "processed_rows=%d valid_rows=%d invalid_rows=%d campaigns=%d\n", summary.ProcessedRows, summary.ValidRows, summary.InvalidRows, summary.Campaigns)
	printInvalidSamples(summary)
}

func printInvalidSamples(summary csvio.ReadSummary) {
	if len(summary.InvalidSamples) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "invalid_row_samples:")
	for _, sample := range summary.InvalidSamples {
		fmt.Fprintf(os.Stderr, "  %s\n", sample)
	}
}
