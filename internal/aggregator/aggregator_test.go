package aggregator

import (
	"reflect"
	"testing"

	"fv-sec001-go/internal/model"
)

func TestAggregatorAddAggregatesMultipleRowsForSameCampaign(t *testing.T) {
	t.Parallel()

	agg := New()
	rows := []model.AdRow{
		{CampaignID: "alpha", Impressions: 100, Clicks: 10, Spend: 12.50, Conversions: 2},
		{CampaignID: "beta", Impressions: 50, Clicks: 5, Spend: 5.00, Conversions: 1},
		{CampaignID: "alpha", Impressions: 25, Clicks: 2, Spend: 7.25, Conversions: 3},
	}
	for _, row := range rows {
		agg.Add(row)
	}

	campaigns := agg.Campaigns()
	if len(campaigns) != 2 {
		t.Fatalf("campaign count = %d, want 2", len(campaigns))
	}

	got := campaigns[0]
	want := model.CampaignAggregate{
		CampaignID:       "alpha",
		TotalImpressions: 125,
		TotalClicks:      12,
		TotalSpend:       19.75,
		TotalConversions: 5,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("alpha aggregate = %+v, want %+v", got, want)
	}
}

func TestCampaignMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   model.CampaignAggregate
		wantCTR float64
		wantCPA *float64
	}{
		{
			name: "computes CTR and CPA",
			input: model.CampaignAggregate{
				CampaignID:       "normal",
				TotalImpressions: 200,
				TotalClicks:      25,
				TotalSpend:       50.00,
				TotalConversions: 4,
			},
			wantCTR: 0.125,
			wantCPA: ptrFloat64(12.5),
		},
		{
			name: "zero impressions has CTR zero",
			input: model.CampaignAggregate{
				CampaignID:       "zero-impressions",
				TotalImpressions: 0,
				TotalClicks:      5,
				TotalSpend:       10.00,
				TotalConversions: 2,
			},
			wantCTR: 0,
			wantCPA: ptrFloat64(5.0),
		},
		{
			name: "zero conversions has unavailable CPA",
			input: model.CampaignAggregate{
				CampaignID:       "zero-conversions",
				TotalImpressions: 100,
				TotalClicks:      10,
				TotalSpend:       99.00,
				TotalConversions: 0,
			},
			wantCTR: 0.1,
			wantCPA: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.input.CTR(); got != tc.wantCTR {
				t.Fatalf("CTR() = %f, want %f", got, tc.wantCTR)
			}

			gotCPA := tc.input.CPA()
			if tc.wantCPA == nil {
				if gotCPA != nil {
					t.Fatalf("CPA() = %v, want nil", *gotCPA)
				}
				return
			}
			if gotCPA == nil {
				t.Fatalf("CPA() = nil, want %f", *tc.wantCPA)
			}
			if *gotCPA != *tc.wantCPA {
				t.Fatalf("CPA() = %f, want %f", *gotCPA, *tc.wantCPA)
			}
		})
	}
}

func TestTopCTRSetsDeterministicOrderAndIncludesZeroConversions(t *testing.T) {
	t.Parallel()

	campaigns := []model.CampaignAggregate{
		{CampaignID: "zeta", TotalClicks: 0, TotalImpressions: 0, TotalSpend: 10, TotalConversions: 0},     // CTR 0.00, included
		{CampaignID: "bravo", TotalClicks: 20, TotalImpressions: 100, TotalSpend: 30, TotalConversions: 0}, // CTR 0.20, CPA unavailable but included
		{CampaignID: "alpha", TotalClicks: 1, TotalImpressions: 5, TotalSpend: 10, TotalConversions: 1},    // CTR 0.20, tie with bravo
		{CampaignID: "charlie", TotalClicks: 10, TotalImpressions: 100, TotalSpend: 20, TotalConversions: 2},
	}

	got := TopCTR(campaigns, 10)
	assertCampaignOrder(t, got, []string{"alpha", "bravo", "charlie", "zeta"})
}

func TestTopCPASetsDeterministicOrderAndExcludesZeroConversions(t *testing.T) {
	t.Parallel()

	campaigns := []model.CampaignAggregate{
		{CampaignID: "zero", TotalSpend: 0, TotalConversions: 0},     // excluded
		{CampaignID: "delta", TotalSpend: 5, TotalConversions: 1},    // CPA 5.00
		{CampaignID: "bravo", TotalSpend: 10, TotalConversions: 10},  // CPA 1.00
		{CampaignID: "alpha", TotalSpend: 3, TotalConversions: 3},    // CPA 1.00, tie with bravo
		{CampaignID: "charlie", TotalSpend: 10, TotalConversions: 2}, // CPA 5.00, tie with delta
	}

	got := TopCPA(campaigns, 10)
	assertCampaignOrder(t, got, []string{"alpha", "bravo", "charlie", "delta"})
}

func TestTopFunctionsLimitResults(t *testing.T) {
	t.Parallel()

	campaigns := []model.CampaignAggregate{
		{CampaignID: "a", TotalClicks: 1, TotalImpressions: 10, TotalSpend: 30, TotalConversions: 1},
		{CampaignID: "b", TotalClicks: 2, TotalImpressions: 10, TotalSpend: 20, TotalConversions: 1},
		{CampaignID: "c", TotalClicks: 3, TotalImpressions: 10, TotalSpend: 10, TotalConversions: 1},
	}

	assertCampaignOrder(t, TopCTR(campaigns, 2), []string{"c", "b"})
	assertCampaignOrder(t, TopCPA(campaigns, 2), []string{"c", "b"})
}

func TestResultRowsPreserveMetricsAndUnavailableCPA(t *testing.T) {
	t.Parallel()

	rows := ResultRows([]model.CampaignAggregate{
		{CampaignID: "a", TotalImpressions: 100, TotalClicks: 25, TotalSpend: 50, TotalConversions: 0},
		{CampaignID: "b", TotalImpressions: 200, TotalClicks: 20, TotalSpend: 60, TotalConversions: 3},
	})

	if len(rows) != 2 {
		t.Fatalf("len = %d, want 2", len(rows))
	}
	if rows[0].CTR != 0.25 || rows[0].CPA != nil {
		t.Fatalf("row 0 metrics = CTR %f CPA %v, want CTR 0.25 CPA nil", rows[0].CTR, rows[0].CPA)
	}
	if rows[1].CTR != 0.10 || rows[1].CPA == nil || *rows[1].CPA != 20.0 {
		t.Fatalf("row 1 metrics = CTR %f CPA %v, want CTR 0.10 CPA 20.00", rows[1].CTR, rows[1].CPA)
	}
}

func assertCampaignOrder(t *testing.T, got []model.CampaignAggregate, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].CampaignID != want[i] {
			t.Fatalf("rank %d = %s, want %s; full order %+v", i, got[i].CampaignID, want[i], got)
		}
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}
