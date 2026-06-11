package csvio

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"fv-sec001-go/internal/model"
)

func TestFormatRowNumericFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		row  model.ResultRow
		want []string
	}{
		{
			name: "rounds spend CTR and CPA to required precision",
			row: model.ResultRow{
				CampaignID:       "campaign-a",
				TotalImpressions: 1234,
				TotalClicks:      123,
				TotalSpend:       45.678,
				TotalConversions: 4,
				CTR:              0.099675,
				CPA:              ptrFloat64(12.345),
			},
			want: []string{"campaign-a", "1234", "123", "45.68", "4", "0.0997", "12.35"},
		},
		{
			name: "pads whole-number decimals",
			row: model.ResultRow{
				CampaignID:       "campaign-b",
				TotalImpressions: 100,
				TotalClicks:      10,
				TotalSpend:       25,
				TotalConversions: 5,
				CTR:              0.1,
				CPA:              ptrFloat64(5),
			},
			want: []string{"campaign-b", "100", "10", "25.00", "5", "0.1000", "5.00"},
		},
		{
			name: "leaves CPA empty when unavailable",
			row: model.ResultRow{
				CampaignID:       "campaign-zero-conversions",
				TotalImpressions: 100,
				TotalClicks:      10,
				TotalSpend:       25,
				TotalConversions: 0,
				CTR:              0.1,
				CPA:              nil,
			},
			want: []string{"campaign-zero-conversions", "100", "10", "25.00", "0", "0.1000", ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := formatRow(tc.row)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("formatRow = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestWriteResultsCreatesRequiredFilesWithExactHeaderAndFormatting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ctrRows := []model.ResultRow{
		{CampaignID: "a", TotalImpressions: 100, TotalClicks: 10, TotalSpend: 7, TotalConversions: 0, CTR: 0.1, CPA: nil},
	}
	cpaRows := []model.ResultRow{
		{CampaignID: "b", TotalImpressions: 200, TotalClicks: 20, TotalSpend: 4, TotalConversions: 4, CTR: 0.1, CPA: ptrFloat64(1)},
	}

	if err := WriteResults(dir, ctrRows, cpaRows); err != nil {
		t.Fatalf("WriteResults returned error: %v", err)
	}

	ctrRecords := readCSVRecords(t, filepath.Join(dir, "top10_ctr.csv"))
	cpaRecords := readCSVRecords(t, filepath.Join(dir, "top10_cpa.csv"))

	wantHeader := []string{"campaign_id", "total_impressions", "total_clicks", "total_spend", "total_conversions", "CTR", "CPA"}
	if !reflect.DeepEqual(ctrRecords[0], wantHeader) {
		t.Fatalf("CTR header = %#v, want %#v", ctrRecords[0], wantHeader)
	}
	if !reflect.DeepEqual(cpaRecords[0], wantHeader) {
		t.Fatalf("CPA header = %#v, want %#v", cpaRecords[0], wantHeader)
	}

	if want := []string{"a", "100", "10", "7.00", "0", "0.1000", ""}; !reflect.DeepEqual(ctrRecords[1], want) {
		t.Fatalf("CTR data row = %#v, want %#v", ctrRecords[1], want)
	}
	if want := []string{"b", "200", "20", "4.00", "4", "0.1000", "1.00"}; !reflect.DeepEqual(cpaRecords[1], want) {
		t.Fatalf("CPA data row = %#v, want %#v", cpaRecords[1], want)
	}
}

func TestWriteResultsCreatesOutputDirectory(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "nested", "results")
	if err := WriteResults(dir, nil, nil); err != nil {
		t.Fatalf("WriteResults returned error: %v", err)
	}
	for _, name := range []string{"top10_ctr.csv", "top10_cpa.csv"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
}

func TestWriteResultsHeaderLineIsExact(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := WriteResults(dir, nil, nil); err != nil {
		t.Fatalf("WriteResults returned error: %v", err)
	}

	content := readFile(t, filepath.Join(dir, "top10_ctr.csv"))
	want := "campaign_id,total_impressions,total_clicks,total_spend,total_conversions,CTR,CPA\n"
	if content != want {
		t.Fatalf("content = %q, want exact header-only file %q", content, want)
	}
	if !strings.HasPrefix(content, want) {
		t.Fatalf("header prefix mismatch: %q", content)
	}
}

func readCSVRecords(t *testing.T, path string) [][]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read csv %s: %v", path, err)
	}
	return records
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func ptrFloat64(v float64) *float64 {
	return &v
}
