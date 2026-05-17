package commands

import (
	"testing"

	"github.com/markes76/cars-il-pp-cli/internal/client"
)

func TestComputeDealScoreRewardsBelowMedianLowMileage(t *testing.T) {
	listing := client.Listing{Price: 89000, Mileage: 65000, Hand: 2, DaysOnMarket: 14}
	market := MarketStats{MedianPrice: 91500, AvgMileage: 74200, AvgDaysOnMarket: 18.4, AvgHand: 1.8}

	score := ComputeDealScore(listing, market)
	if score.Score < 70 || score.Score > 80 {
		t.Fatalf("expected fair deal score around 73, got %d (%+v)", score.Score, score)
	}
	if score.Verdict == "" {
		t.Fatalf("expected verdict")
	}
}
