package csvio

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"fv-sec001-go/internal/aggregator"
	"fv-sec001-go/internal/model"
)

var expectedHeader = []string{"campaign_id", "date", "impressions", "clicks", "spend", "conversions"}

const maxInvalidSamples = 10

// ReadSummary describes the streaming import result.
type ReadSummary struct {
	ProcessedRows  uint64
	ValidRows      uint64
	InvalidRows    uint64
	Campaigns      int
	InvalidSamples []string
}

// MalformedRows is kept for compatibility with older callers.
func (s ReadSummary) MalformedRows() uint64 {
	return s.InvalidRows
}

// ProcessFile streams a CSV file, validates rows, and returns campaign aggregates.
func ProcessFile(path string, strict bool) ([]model.CampaignAggregate, ReadSummary, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, ReadSummary{}, fmt.Errorf("open input file: %w", err)
	}
	defer file.Close()

	return ProcessReader(file, strict)
}

// ProcessReader streams CSV records from r. It reads and validates the header,
// processes one data record at a time, and stores only per-campaign aggregates.
func ProcessReader(r io.Reader, strict bool) ([]model.CampaignAggregate, ReadSummary, error) {
	reader := csv.NewReader(bufio.NewReaderSize(r, 1024*1024))
	reader.FieldsPerRecord = -1
	reader.ReuseRecord = true

	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, ReadSummary{}, errors.New("input CSV is empty")
		}
		return nil, ReadSummary{}, fmt.Errorf("read header: %w", err)
	}
	if err := validateHeader(header); err != nil {
		return nil, ReadSummary{}, err
	}

	agg := aggregator.New()
	summary := ReadSummary{}

	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			summary.ProcessedRows++
			rowErr := fmt.Errorf("malformed CSV: %w", err)
			if handleInvalidRow(&summary, strict, rowErr) {
				return nil, summary, fmt.Errorf("row %d: %w", summary.ProcessedRows+1, rowErr)
			}
			continue
		}

		summary.ProcessedRows++
		row, err := parseRecord(record)
		if err != nil {
			if handleInvalidRow(&summary, strict, err) {
				return nil, summary, fmt.Errorf("row %d: %w", summary.ProcessedRows+1, err)
			}
			continue
		}

		agg.Add(row)
		summary.ValidRows++
	}

	campaigns := agg.Aggregates()
	summary.Campaigns = len(campaigns)
	if summary.ValidRows == 0 {
		return nil, summary, errors.New("no valid data rows processed")
	}
	return campaigns, summary, nil
}

func handleInvalidRow(summary *ReadSummary, strict bool, err error) bool {
	summary.InvalidRows++
	if len(summary.InvalidSamples) < maxInvalidSamples {
		summary.InvalidSamples = append(summary.InvalidSamples, fmt.Sprintf("row %d: %v", summary.ProcessedRows+1, err))
	}
	return strict
}

func validateHeader(header []string) error {
	if len(header) != len(expectedHeader) {
		return fmt.Errorf("invalid header: got %d columns, want %d", len(header), len(expectedHeader))
	}
	for i, got := range header {
		if strings.TrimSpace(got) != expectedHeader[i] {
			return fmt.Errorf("invalid header column %d: got %q, want %q", i+1, got, expectedHeader[i])
		}
	}
	return nil
}

func parseRecord(record []string) (model.AdRow, error) {
	if len(record) != len(expectedHeader) {
		return model.AdRow{}, fmt.Errorf("got %d columns, want %d", len(record), len(expectedHeader))
	}

	campaignID := strings.TrimSpace(record[0])
	if campaignID == "" {
		return model.AdRow{}, errors.New("campaign_id is required")
	}

	impressions, err := parseUintField("impressions", record[2])
	if err != nil {
		return model.AdRow{}, err
	}
	clicks, err := parseUintField("clicks", record[3])
	if err != nil {
		return model.AdRow{}, err
	}
	spend, err := parseSpend(record[4])
	if err != nil {
		return model.AdRow{}, err
	}
	conversions, err := parseUintField("conversions", record[5])
	if err != nil {
		return model.AdRow{}, err
	}

	return model.AdRow{
		CampaignID:  campaignID,
		Impressions: impressions,
		Clicks:      clicks,
		Spend:       spend,
		Conversions: conversions,
	}, nil
}

func parseUintField(name, value string) (uint64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	if strings.HasPrefix(trimmed, "-") {
		return 0, fmt.Errorf("%s must be non-negative", name)
	}
	parsed, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", name, value)
	}
	return parsed, nil
}

func parseSpend(value string) (float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, errors.New("spend is required")
	}
	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, fmt.Errorf("invalid spend %q", value)
	}
	if parsed < 0 {
		return 0, errors.New("spend must be non-negative")
	}
	return parsed, nil
}
