package model

// AdRow is a validated input CSV row.
type AdRow struct {
	CampaignID  string
	Impressions uint64
	Clicks      uint64
	Spend       float64
	Conversions uint64
}

// CampaignAggregate contains campaign-level totals derived from valid rows only.
type CampaignAggregate struct {
	CampaignID       string
	TotalImpressions uint64
	TotalClicks      uint64
	TotalSpend       float64
	TotalConversions uint64
}

// CTR returns total_clicks / total_impressions. If impressions are zero, CTR is 0.
func (c CampaignAggregate) CTR() float64 {
	if c.TotalImpressions == 0 {
		return 0
	}
	return float64(c.TotalClicks) / float64(c.TotalImpressions)
}

// CPA returns total_spend / total_conversions. If conversions are zero, CPA is unavailable.
func (c CampaignAggregate) CPA() *float64 {
	if c.TotalConversions == 0 {
		return nil
	}
	cpa := c.TotalSpend / float64(c.TotalConversions)
	return &cpa
}

// ResultRow is a row ready to be written to an output CSV.
type ResultRow struct {
	CampaignID       string
	TotalImpressions uint64
	TotalClicks      uint64
	TotalSpend       float64
	TotalConversions uint64
	CTR              float64
	CPA              *float64
}
