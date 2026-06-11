package csvio

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"fv-sec001-go/internal/model"
)

var outputHeader = []string{"campaign_id", "total_impressions", "total_clicks", "total_spend", "total_conversions", "CTR", "CPA"}

// WriteResults writes the two required output files into outputDir.
func WriteResults(outputDir string, ctrRows, cpaRows []model.ResultRow) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	ctrPath := filepath.Join(outputDir, "top10_ctr.csv")
	cpaPath := filepath.Join(outputDir, "top10_cpa.csv")
	if err := writeCSVAtomic(ctrPath, ctrRows); err != nil {
		return fmt.Errorf("write CTR results: %w", err)
	}
	if err := writeCSVAtomic(cpaPath, cpaRows); err != nil {
		return fmt.Errorf("write CPA results: %w", err)
	}
	return nil
}

func writeCSVAtomic(path string, rows []model.ResultRow) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmp, err := os.CreateTemp(dir, "."+base+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := writeCSVFile(tmp, rows); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	removeTmp = false
	return nil
}

func writeCSVFile(file *os.File, rows []model.ResultRow) (err error) {
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	writer := csv.NewWriter(file)
	if err := writer.Write(outputHeader); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write(formatRow(row)); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func formatRow(row model.ResultRow) []string {
	cpa := ""
	if row.CPA != nil {
		cpa = fmt.Sprintf("%.2f", *row.CPA)
	}

	return []string{
		row.CampaignID,
		strconv.FormatUint(row.TotalImpressions, 10),
		strconv.FormatUint(row.TotalClicks, 10),
		fmt.Sprintf("%.2f", row.TotalSpend),
		strconv.FormatUint(row.TotalConversions, 10),
		fmt.Sprintf("%.4f", row.CTR),
		cpa,
	}
}
