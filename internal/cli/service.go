package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/markes76/cars-il-pp-cli/internal/store"
)

type Service struct {
	DB         *store.DB
	Dispatcher client.Dispatcher
}

func NewService(db *store.DB) Service {
	return Service{
		DB: db,
		Dispatcher: client.Dispatcher{
			Yad2:       client.NewYad2Client(),
			AutoTrader: client.NewAutoTraderClient(),
		},
	}
}

func (s Service) Search(params client.SearchParams) ([]client.Listing, error) {
	switch params.DataSource {
	case "local":
		return searchDomainListings(s.DB, params)
	case "live":
		listings, _, err := s.Dispatcher.Search(params)
		return listings, err
	default:
		local, err := searchDomainListings(s.DB, params)
		if err == nil && len(local) > 0 {
			return local, nil
		}
		listings, _, liveErr := s.Dispatcher.Search(params)
		if liveErr != nil && len(local) == 0 {
			return nil, liveErr
		}
		return listings, nil
	}
}

func (s Service) Get(id string, dataSource string) (client.Listing, error) {
	if dataSource != "live" {
		listing, err := s.DB.GetListing(id)
		if err == nil {
			return listing, nil
		}
		if dataSource == "local" {
			return client.Listing{}, err
		}
	}
	if strings.HasPrefix(id, "autotrader-") {
		return s.Dispatcher.AutoTrader.GetListing(id)
	}
	return s.Dispatcher.Yad2.GetListing(id)
}

func (s Service) Sync(params client.SearchParams) (SyncResult, error) {
	liveParams := params
	liveParams.DataSource = "live"
	scope := syncScope(params)
	if cursor, _ := s.DB.GetSyncState(params.Source, scope); cursor != "" && liveParams.Page == 0 {
		fmt.Sscanf(cursor, "page=%d", &liveParams.Page)
	}
	if liveParams.Page <= 0 {
		liveParams.Page = 1
	}
	limit := liveParams.Limit
	if limit <= 0 {
		limit = 20
	}
	result := SyncResult{}
	for page := liveParams.Page; page < liveParams.Page+5 && result.Total < limit; page++ {
		liveParams.Page = page
		liveParams.Limit = limit - result.Total
		listings, pagination, err := s.Dispatcher.Search(liveParams)
		if err != nil {
			return result, err
		}
		if len(listings) == 0 {
			break
		}
		for _, listing := range listings {
			_, existingErr := s.DB.GetListing(listing.ID)
			if existingErr != nil {
				result.New++
			}
			if err := syncDomainUpsertListing(s.DB, listing); err != nil {
				return result, err
			}
			result.Total++
		}
		nextCursor := fmt.Sprintf("page=%d", page+1)
		if err := s.DB.SaveSyncState(params.Source, scope, nextCursor); err != nil {
			return result, err
		}
		if !pagination.HasNext {
			break
		}
	}
	return result, nil
}

func syncScope(params client.SearchParams) string {
	return strings.Join([]string{
		params.Make,
		params.Model,
		fmt.Sprint(params.YearMin),
		fmt.Sprint(params.YearMax),
		fmt.Sprint(params.PriceMin),
		fmt.Sprint(params.PriceMax),
		fmt.Sprint(params.MileageMax),
		params.Fuel,
		params.Gear,
	}, "|")
}

type SyncResult struct {
	Total        int `json:"total"`
	New          int `json:"new"`
	PriceChanges int `json:"price_changes"`
	Removed      int `json:"removed"`
}

type MarketStats struct {
	Make             string  `json:"make"`
	Model            string  `json:"model"`
	YearMin          int     `json:"year_min"`
	YearMax          int     `json:"year_max"`
	ActiveListings   int     `json:"active_listings"`
	AvgPrice         int     `json:"avg_price"`
	MedianPrice      int     `json:"median_price"`
	MinPrice         int     `json:"min_price"`
	MaxPrice         int     `json:"max_price"`
	AvgMileage       int     `json:"avg_mileage"`
	AvgDaysOnMarket  float64 `json:"avg_days_on_market"`
	PrivatePercent   float64 `json:"private_percent"`
	DealerPercent    float64 `json:"dealer_percent"`
	PrivateAvgPrice  int     `json:"private_avg_price"`
	DealerAvgPrice   int     `json:"dealer_avg_price"`
	DealerPremiumPct float64 `json:"dealer_premium_pct"`
	MostCommonCity   string  `json:"most_common_city"`
	AvgHand          float64 `json:"avg_hand"`
}

func (s Service) Market(params client.SearchParams) (MarketStats, error) {
	params.DataSource = "local"
	params.Limit = 200
	listings, err := s.DB.SearchListings(params)
	if err != nil {
		return MarketStats{}, err
	}
	if len(listings) == 0 {
		return MarketStats{}, client.NotFound("no local listings match market query; run cars-il sync first")
	}
	return ComputeMarketStats(listings, params), nil
}

func ComputeMarketStats(listings []client.Listing, params client.SearchParams) MarketStats {
	stats := MarketStats{Make: params.Make, Model: params.Model, YearMin: params.YearMin, YearMax: params.YearMax, ActiveListings: len(listings)}
	prices := make([]int, 0, len(listings))
	cityCounts := map[string]int{}
	var priceSum, mileageSum, mileageCount, daysSum, privateCount, dealerCount, privatePrice, dealerPrice, handSum, handCount int
	for _, listing := range listings {
		if listing.Price > 0 {
			prices = append(prices, listing.Price)
			priceSum += listing.Price
			if stats.MinPrice == 0 || listing.Price < stats.MinPrice {
				stats.MinPrice = listing.Price
			}
			if listing.Price > stats.MaxPrice {
				stats.MaxPrice = listing.Price
			}
			if listing.IsDealer {
				dealerPrice += listing.Price
			} else {
				privatePrice += listing.Price
			}
		}
		if listing.Mileage > 0 {
			mileageSum += listing.Mileage
			mileageCount++
		}
		daysSum += listing.DaysOnMarket
		if listing.IsDealer {
			dealerCount++
		} else {
			privateCount++
		}
		if listing.City != "" {
			cityCounts[listing.City]++
		}
		if listing.Hand > 0 {
			handSum += listing.Hand
			handCount++
		}
	}
	sort.Ints(prices)
	if len(prices) > 0 {
		stats.AvgPrice = priceSum / len(prices)
		stats.MedianPrice = median(prices)
	}
	if mileageCount > 0 {
		stats.AvgMileage = mileageSum / mileageCount
	}
	if len(listings) > 0 {
		stats.AvgDaysOnMarket = float64(daysSum) / float64(len(listings))
		stats.PrivatePercent = float64(privateCount) * 100 / float64(len(listings))
		stats.DealerPercent = float64(dealerCount) * 100 / float64(len(listings))
	}
	if privateCount > 0 {
		stats.PrivateAvgPrice = privatePrice / privateCount
	}
	if dealerCount > 0 {
		stats.DealerAvgPrice = dealerPrice / dealerCount
	}
	if stats.PrivateAvgPrice > 0 && stats.DealerAvgPrice > 0 {
		stats.DealerPremiumPct = (float64(stats.DealerAvgPrice-stats.PrivateAvgPrice) / float64(stats.PrivateAvgPrice)) * 100
	}
	if handCount > 0 {
		stats.AvgHand = float64(handSum) / float64(handCount)
	}
	for city, count := range cityCounts {
		if count > cityCounts[stats.MostCommonCity] {
			stats.MostCommonCity = city
		}
	}
	return stats
}

func median(values []int) int {
	if len(values) == 0 {
		return 0
	}
	mid := len(values) / 2
	if len(values)%2 == 1 {
		return values[mid]
	}
	return (values[mid-1] + values[mid]) / 2
}
