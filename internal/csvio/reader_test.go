package csvio

import (
	"strings"
	"testing"

	"fv-sec001-go/internal/model"
)

func TestProcessReaderAggregatesAndReturnsSummary(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(`campaign_id,date,impressions,clicks,spend,conversions
alpha,2024-01-01,100,10,12.50,2
beta,2024-01-01,0,0,3.25,0
alpha,2024-01-02,50,5,7.50,1
`)

	campaigns, summary, err := ProcessReader(input, false)
	if err != nil {
		t.Fatalf("ProcessReader returned error: %v", err)
	}
	if summary.ProcessedRows != 3 || summary.ValidRows != 3 || summary.InvalidRows != 0 || summary.Campaigns != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(campaigns) != 2 {
		t.Fatalf("campaign count = %d, want 2", len(campaigns))
	}

	byID := campaignsByID(campaigns)
	alpha := byID["alpha"]
	if alpha.CampaignID != "alpha" || alpha.TotalImpressions != 150 || alpha.TotalClicks != 15 || alpha.TotalSpend != 20.0 || alpha.TotalConversions != 3 {
		t.Fatalf("unexpected alpha aggregate: %+v", alpha)
	}
	beta := byID["beta"]
	if beta.CampaignID != "beta" || beta.TotalImpressions != 0 || beta.CTR() != 0 || beta.CPA() != nil {
		t.Fatalf("unexpected beta aggregate/metrics: %+v CTR=%f CPA=%v", beta, beta.CTR(), beta.CPA())
	}
}

func TestProcessReaderSkipsMalformedRowsInNonStrictMode(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(`campaign_id,date,impressions,clicks,spend,conversions
ok,2024-01-01,100,10,12.50,2
bad_negative_int,2024-01-01,-1,10,1.00,1
bad_negative_spend,2024-01-01,10,1,-2.00,1
bad_missing_columns,2024-01-01,1
,2024-01-01,1,1,1.00,1
ok,2024-01-02,50,5,7.50,1
`)

	campaigns, summary, err := ProcessReader(input, false)
	if err != nil {
		t.Fatalf("ProcessReader returned error: %v", err)
	}
	if summary.ProcessedRows != 6 || summary.ValidRows != 2 || summary.InvalidRows != 4 || summary.Campaigns != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(campaigns) != 1 || campaigns[0].CampaignID != "ok" {
		t.Fatalf("unexpected campaigns: %+v", campaigns)
	}
}

func TestProcessReaderFailsOnMalformedRowsInStrictMode(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(`campaign_id,date,impressions,clicks,spend,conversions
ok,2024-01-01,100,10,12.50,2
bad,2024-01-01,100,10,-1.00,1
ok,2024-01-02,50,5,7.50,1
`)

	_, summary, err := ProcessReader(input, true)
	if err == nil {
		t.Fatal("ProcessReader returned nil error, want strict-mode error")
	}
	if summary.ProcessedRows != 2 || summary.ValidRows != 1 || summary.InvalidRows != 1 {
		t.Fatalf("unexpected summary after strict failure: %+v", summary)
	}
}

func TestProcessReaderRejectsInvalidHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "missing date column",
			input: "campaign_id,impressions,clicks,spend,conversions\n",
		},
		{
			name:  "reordered columns",
			input: "campaign_id,date,clicks,impressions,spend,conversions\n",
		},
		{
			name:  "extra column",
			input: "campaign_id,date,impressions,clicks,spend,conversions,extra\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := ProcessReader(strings.NewReader(tc.input), false)
			if err == nil {
				t.Fatal("ProcessReader returned nil error, want header error")
			}
		})
	}
}

func TestParseRecordMalformedRowsTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		record []string
	}{
		{name: "empty campaign", record: []string{"", "2024-01-01", "1", "1", "1.00", "1"}},
		{name: "negative impressions", record: []string{"bad", "2024-01-01", "-1", "1", "1.00", "1"}},
		{name: "negative clicks", record: []string{"bad", "2024-01-01", "1", "-1", "1.00", "1"}},
		{name: "negative spend", record: []string{"bad", "2024-01-01", "1", "1", "-1.00", "1"}},
		{name: "negative conversions", record: []string{"bad", "2024-01-01", "1", "1", "1.00", "-1"}},
		{name: "invalid impressions", record: []string{"bad", "2024-01-01", "nope", "1", "1.00", "1"}},
		{name: "invalid spend", record: []string{"bad", "2024-01-01", "1", "1", "nope", "1"}},
		{name: "missing fields", record: []string{"bad", "2024-01-01", "1"}},
		{name: "extra fields", record: []string{"bad", "2024-01-01", "1", "1", "1.00", "1", "extra"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := parseRecord(tc.record); err == nil {
				t.Fatal("parseRecord returned nil error, want validation error")
			}
		})
	}
}

func TestProcessReaderFailsWhenNoValidRowsRemain(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(`campaign_id,date,impressions,clicks,spend,conversions
bad,2024-01-01,-1,1,1.00,1
also_bad,2024-01-02,1,1,-2.00,1
`)

	campaigns, summary, err := ProcessReader(input, false)
	if err == nil {
		t.Fatal("ProcessReader returned nil error, want no-valid-rows error")
	}
	if len(campaigns) != 0 {
		t.Fatalf("campaigns = %+v, want none", campaigns)
	}
	if summary.ProcessedRows != 2 || summary.ValidRows != 0 || summary.InvalidRows != 2 || summary.Campaigns != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(summary.InvalidSamples) != 2 {
		t.Fatalf("invalid samples len = %d, want 2: %+v", len(summary.InvalidSamples), summary.InvalidSamples)
	}
}

func campaignsByID(campaigns []model.CampaignAggregate) map[string]model.CampaignAggregate {
	byID := make(map[string]model.CampaignAggregate, len(campaigns))
	for _, campaign := range campaigns {
		byID[campaign.CampaignID] = campaign
	}
	return byID
}
