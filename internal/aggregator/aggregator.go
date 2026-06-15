package aggregator

import (
	"cmp"
	"slices"
	"strings"

	"fv-sec001-go/internal/model"
)

// Aggregator stores campaign-level aggregates. It does not store raw rows.
type Aggregator struct {
	campaigns map[string]*model.CampaignAggregate
}

// New creates an empty campaign aggregator.
func New() *Aggregator {
	return &Aggregator{campaigns: make(map[string]*model.CampaignAggregate)}
}

// Add applies one validated input row to its campaign aggregate.
func (a *Aggregator) Add(row model.AdRow) {
	agg, ok := a.campaigns[row.CampaignID]
	if !ok {
		// csv.Reader may return field strings backed by a larger record string.
		// Clone only campaign IDs that become map keys so the aggregate map does
		// not retain full CSV lines for each unique campaign.
		campaignID := strings.Clone(row.CampaignID)
		agg = &model.CampaignAggregate{CampaignID: campaignID}
		a.campaigns[campaignID] = agg
	}

	agg.TotalImpressions += row.Impressions
	agg.TotalClicks += row.Clicks
	agg.TotalSpend += row.Spend
	agg.TotalConversions += row.Conversions
}

// Count returns the number of distinct campaigns.
func (a *Aggregator) Count() int {
	return len(a.campaigns)
}

// Aggregates returns aggregate values without imposing an order. Use ranking
// functions for deterministic result-file ordering.
func (a *Aggregator) Aggregates() []model.CampaignAggregate {
	out := make([]model.CampaignAggregate, 0, len(a.campaigns))
	for _, campaign := range a.campaigns {
		out = append(out, *campaign)
	}
	return out
}

// Campaigns returns aggregate values in deterministic campaign_id order.
func (a *Aggregator) Campaigns() []model.CampaignAggregate {
	out := make([]model.CampaignAggregate, 0, len(a.campaigns))
	for _, campaign := range a.campaigns {
		out = append(out, *campaign)
	}
	slices.SortFunc(out, func(x, y model.CampaignAggregate) int {
		return cmp.Compare(x.CampaignID, y.CampaignID)
	})
	return out
}

// TopCTR returns campaigns with the highest CTR. Campaigns with zero impressions
// are included with CTR equal to 0.
//
// Tie-breaking is deterministic and intentionally limited to campaign_id:
// CTR descending, then campaign_id ascending.
func TopCTR(campaigns []model.CampaignAggregate, n int) []model.CampaignAggregate {
	ranked := append([]model.CampaignAggregate(nil), campaigns...)
	slices.SortFunc(ranked, func(a, b model.CampaignAggregate) int {
		aCTR := a.CTR()
		bCTR := b.CTR()
		if aCTR < bCTR {
			return 1
		}
		if aCTR > bCTR {
			return -1
		}
		return cmp.Compare(a.CampaignID, b.CampaignID)
	})
	return limit(ranked, n)
}

// TopCPA returns campaigns with the lowest CPA. Campaigns with zero conversions
// are excluded because CPA is unavailable.
//
// Tie-breaking is deterministic and intentionally limited to campaign_id:
// CPA ascending, then campaign_id ascending.
func TopCPA(campaigns []model.CampaignAggregate, n int) []model.CampaignAggregate {
	ranked := make([]model.CampaignAggregate, 0, len(campaigns))
	for _, campaign := range campaigns {
		if campaign.TotalConversions > 0 {
			ranked = append(ranked, campaign)
		}
	}

	slices.SortFunc(ranked, func(a, b model.CampaignAggregate) int {
		aCPA := *a.CPA()
		bCPA := *b.CPA()
		if aCPA < bCPA {
			return -1
		}
		if aCPA > bCPA {
			return 1
		}
		return cmp.Compare(a.CampaignID, b.CampaignID)
	})
	return limit(ranked, n)
}

// ResultRows converts aggregates into output rows. CTR is always available; CPA is
// nil when total_conversions is zero.
func ResultRows(campaigns []model.CampaignAggregate) []model.ResultRow {
	rows := make([]model.ResultRow, 0, len(campaigns))
	for _, campaign := range campaigns {
		rows = append(rows, model.ResultRow{
			CampaignID:       campaign.CampaignID,
			TotalImpressions: campaign.TotalImpressions,
			TotalClicks:      campaign.TotalClicks,
			TotalSpend:       campaign.TotalSpend,
			TotalConversions: campaign.TotalConversions,
			CTR:              campaign.CTR(),
			CPA:              campaign.CPA(),
		})
	}
	return rows
}

func limit[T any](items []T, n int) []T {
	if n <= 0 || len(items) <= n {
		return items
	}
	return items[:n]
}
