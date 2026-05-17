package mcp

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

func RegisterTools() []Tool {
	return ToolList()
}

func ToolList() []Tool {
	return []Tool{
		{Name: "context", Description: "Returns object describing the Israeli used-car market intelligence context, supported sources, Hebrew handling, and when to call sync before analytics. Call this first when planning an agent workflow.", InputSchema: objectSchema()},
		{Name: "search", Description: "Returns array of unified car listings from local cache or live Yad2/AutoTrader adapters. Supports Hebrew and English make/model aliases, price, mileage, fuel, gear, city, region, and source filters.", InputSchema: objectSchema()},
		{Name: "sync", Description: "Returns object with sync counts after fetching read-only live listings into SQLite. Requires conservative limits and stores price history for later market and deal tools.", InputSchema: objectSchema()},
		{Name: "listings_get", Description: "Returns object with one full listing, including Hebrew description, image URLs, test expiry, owner count, and source URL.", InputSchema: objectSchema()},
		{Name: "market", Description: "Returns object with local aggregate market stats. Requires sync first so median price, dealer premium, mileage averages, and days-on-market are computed from SQLite.", InputSchema: objectSchema()},
		{Name: "deal", Description: "Returns object scoring a listing against local market medians using price, mileage, days-on-market, and owner count. Requires sync first for useful comparison.", InputSchema: objectSchema()},
		{Name: "price_history", Description: "Returns array of observed price points for one listing. Requires at least one sync and becomes more useful after repeated syncs.", InputSchema: objectSchema()},
		{Name: "stale", Description: "Returns array of local listings whose days-on-market suggests seller motivation. Requires sync first and supports the same search filters.", InputSchema: objectSchema()},
		{Name: "compare", Description: "Returns array of up to five listings for side-by-side comparison by an agent or UI client.", InputSchema: objectSchema()},
		{Name: "depreciation", Description: "Returns array describing average price by model year for a local make/model cohort. Requires enough synced listings to produce a meaningful curve.", InputSchema: objectSchema()},
		{Name: "market_heat", Description: "Returns array of make/model cohorts sorted by average days-on-market so agents can identify fast-moving Israeli car segments.", InputSchema: objectSchema()},
		{Name: "doctor", Description: "Returns object with connectivity, auth/config, version, and SQLite store health for troubleshooting before search or sync.", InputSchema: objectSchema()},
	}
}

func objectSchema() map[string]interface{} {
	return map[string]interface{}{"type": "object", "additionalProperties": true}
}

func handleContext() map[string]interface{} {
	return map[string]interface{}{
		"sources": []string{"yad2", "autotrader"},
		"language": "he",
		"currency": "ILS",
		"note": "AutoTrader IL currently exposes a WordPress services site, not a listing catalogue.",
	}
}

func handleSearch()  {}
func handleSync()    {}
func handleSQL()     {}

var ToolHandlers = map[string]interface{}{
	"context": handleContext,
	"search":  handleSearch,
	"sync":    handleSync,
	"sql":     handleSQL,
}
