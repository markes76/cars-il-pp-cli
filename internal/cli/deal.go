package commands

import (
	"fmt"
	"math"

	"github.com/markes76/cars-il-pp-cli/internal/client"
)

type DealResult struct {
	Listing          client.Listing `json:"listing"`
	MarketMedian     int            `json:"market_median"`
	PriceVsMedianPct float64        `json:"price_vs_median_pct"`
	MileageVsAvgPct  float64        `json:"mileage_vs_avg_pct"`
	Score            int            `json:"score"`
	Verdict          string         `json:"verdict"`
	NegotiationLow   int            `json:"negotiation_low"`
	NegotiationHigh  int            `json:"negotiation_high"`
}

func (s Service) Deal(id string) (DealResult, error) {
	listing, err := s.Get(id, "auto")
	if err != nil {
		return DealResult{}, err
	}
	params := client.SearchParams{Make: listing.Make, Model: listing.Model, YearMin: listing.Year, YearMax: listing.Year, Limit: 200, DataSource: "local"}
	stats, err := s.Market(params)
	if err != nil {
		return DealResult{}, err
	}
	result := ComputeDealScore(listing, stats)
	result.Listing = listing
	return result, nil
}

func ComputeDealScore(listing client.Listing, market MarketStats) DealResult {
	priceScore := 50.0
	if market.MedianPrice > 0 && listing.Price > 0 {
		diff := float64(market.MedianPrice-listing.Price) / float64(market.MedianPrice)
		priceScore = clamp(70+diff*600, 0, 100)
	}
	mileageScore := 50.0
	if market.AvgMileage > 0 && listing.Mileage > 0 {
		diff := float64(market.AvgMileage-listing.Mileage) / float64(market.AvgMileage)
		mileageScore = clamp(65+diff*250, 0, 100)
	}
	domScore := clamp(50+float64(listing.DaysOnMarket)-market.AvgDaysOnMarket, 0, 100)
	handScore := 60.0
	if market.AvgHand > 0 && listing.Hand > 0 {
		handScore = clamp(65+(market.AvgHand-float64(listing.Hand))*20, 0, 100)
	}
	score := int(math.Round(priceScore*0.40 + mileageScore*0.30 + domScore*0.15 + handScore*0.15))
	verdict := "FAIR DEAL"
	if score >= 82 {
		verdict = "STRONG DEAL"
	} else if score < 55 {
		verdict = "WEAK DEAL"
	}
	if market.MedianPrice > 0 && listing.Price > 0 {
		verdict = fmt.Sprintf("%s - priced %.1f%% vs market median", verdict, (float64(listing.Price-market.MedianPrice)/float64(market.MedianPrice))*100)
	}
	low := int(float64(listing.Price) * 0.955)
	high := int(float64(listing.Price) * 0.978)
	result := DealResult{
		MarketMedian:     market.MedianPrice,
		PriceVsMedianPct: pctDelta(listing.Price, market.MedianPrice),
		MileageVsAvgPct:  pctDelta(listing.Mileage, market.AvgMileage),
		Score:            score,
		Verdict:          verdict,
		NegotiationLow:   low,
		NegotiationHigh:  high,
	}
	return result
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func pctDelta(value, baseline int) float64 {
	if value == 0 || baseline == 0 {
		return 0
	}
	return (float64(value-baseline) / float64(baseline)) * 100
}
