package client

import (
	"errors"
	"strings"
	"time"
)

const (
	SourceYad2       = "yad2"
	SourceAutoTrader = "autotrader"
	SourceAll        = "all"
)

type Listing struct {
	ID               string   `json:"id"`
	Source           string   `json:"source"`
	Make             string   `json:"make"`
	Model            string   `json:"model"`
	Year             VehicleYear `json:"year"`
	Mileage          MileageKM   `json:"mileage"`
	Price            PriceILS    `json:"price"`
	City             string   `json:"city"`
	Region           string   `json:"region"`
	FuelType         string   `json:"fuel_type"`
	GearType         string   `json:"gear_type"`
	Color            string   `json:"color"`
	Hand             OwnerHand `json:"hand"`
	IsDealer         bool     `json:"is_dealer"`
	TestExpiry       string   `json:"test_expiry"`
	Description      string   `json:"description"`
	ImageURLs        []string `json:"image_urls"`
	URL              string   `json:"url"`
	FirstSeenAt      string   `json:"first_seen_at"`
	LastSeenAt       string   `json:"last_seen_at"`
	PriceAtFirstSeen PriceILS `json:"price_at_first_seen"`
	DaysOnMarket     int      `json:"days_on_market"`
}

type PricePoint struct {
	ListingID  string `json:"listing_id"`
	RecordedAt string `json:"recorded_at"`
	Price      int    `json:"price"`
}

type SearchParams struct {
	Make          string `json:"make"`
	Model         string `json:"model"`
	YearMin       int    `json:"year_min"`
	YearMax       int    `json:"year_max"`
	PriceMin      int    `json:"price_min"`
	PriceMax      int    `json:"price_max"`
	MileageMax    int    `json:"mileage_max"`
	City          string `json:"city"`
	Region        string `json:"region"`
	Fuel          string `json:"fuel"`
	Gear          string `json:"gear"`
	HandMax       int    `json:"hand_max"`
	PrivateOnly   bool   `json:"private_only"`
	DealerOnly    bool   `json:"dealer_only"`
	Sort          string `json:"sort"`
	Source        string `json:"source"`
	Limit         int    `json:"limit"`
	Page          int    `json:"page"`
	DataSource    string `json:"data_source"`
	Query         string `json:"query"`
	MaxDaily      int    `json:"max_daily"`
	NoRobotsFetch bool   `json:"-"`
}

type PaginationState struct {
	Page    int    `json:"page"`
	HasNext bool   `json:"has_next"`
	NextURL string `json:"next_url,omitempty"`
	Total   int    `json:"total,omitempty"`
}

type CarSource interface {
	Search(params SearchParams) ([]Listing, PaginationState, error)
	GetListing(id string) (Listing, error)
	NormalizeToUnifiedSchema(raw interface{}) (Listing, error)
}

type Dispatcher struct {
	Yad2       CarSource
	AutoTrader CarSource
}

func (d Dispatcher) Search(params SearchParams) ([]Listing, PaginationState, error) {
	source := NormalizeSource(params.Source)
	var out []Listing
	var state PaginationState
	var errs []error
	if source == SourceYad2 || source == SourceAll {
		listings, pagination, err := d.Yad2.Search(params)
		if err != nil {
			errs = append(errs, err)
		}
		out = append(out, listings...)
		if pagination.Total > state.Total {
			state = pagination
		}
	}
	if source == SourceAutoTrader || source == SourceAll {
		listings, pagination, err := d.AutoTrader.Search(params)
		if err != nil {
			errs = append(errs, err)
		}
		out = append(out, listings...)
		if pagination.Total > state.Total {
			state = pagination
		}
	}
	if len(out) == 0 && len(errs) > 0 {
		return nil, state, errors.Join(errs...)
	}
	if params.Limit > 0 && len(out) > params.Limit {
		out = out[:params.Limit]
	}
	return out, state, nil
}

func NormalizeSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "", "all", "auto":
		return SourceAll
	case "yad2", "יד2":
		return SourceYad2
	case "autotrader", "auto-trader":
		return SourceAutoTrader
	default:
		return source
	}
}

var makeAliases = map[string]string{
	"toyota": "טויוטה",
	"טויוטה": "טויוטה",
	"honda":  "הונדה",
	"הונדה":  "הונדה",
	"mazda":  "מאזדה",
	"מזדה":   "מאזדה",
	"מאזדה":  "מאזדה",
}

var modelAliases = map[string]string{
	"corolla": "קורולה",
	"קורולה": "קורולה",
	"civic":   "סיוויק",
	"סיוויק":  "סיוויק",
	"cx-5":    "CX-5",
	"cx5":     "CX-5",
	"rav4":    "RAV4",
	"rav 4":   "RAV4",
}

func NormalizeMake(make string) string {
	key := strings.ToLower(strings.TrimSpace(make))
	if value, ok := makeAliases[key]; ok {
		return value
	}
	return strings.TrimSpace(make)
}

func NormalizeModel(model string) string {
	key := strings.ToLower(strings.TrimSpace(model))
	if value, ok := modelAliases[key]; ok {
		return value
	}
	return strings.TrimSpace(model)
}

func MakeAlternates(make string) []string {
	normalized := NormalizeMake(make)
	if normalized == "" {
		return nil
	}
	out := []string{normalized}
	for k, v := range makeAliases {
		if v == normalized && k != normalized {
			out = append(out, k)
		}
	}
	return out
}

func ModelAlternates(model string) []string {
	normalized := NormalizeModel(model)
	if normalized == "" {
		return nil
	}
	out := []string{normalized}
	for k, v := range modelAliases {
		if v == normalized && k != normalized {
			out = append(out, k)
		}
	}
	return out
}

func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func DaysBetween(start, end string) int {
	a, errA := time.Parse(time.RFC3339, start)
	b, errB := time.Parse(time.RFC3339, end)
	if errA != nil || errB != nil || b.Before(a) {
		return 0
	}
	return int(b.Sub(a).Hours() / 24)
}
