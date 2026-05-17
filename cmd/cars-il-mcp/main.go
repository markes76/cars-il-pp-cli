package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/markes76/cars-il-pp-cli/internal/cli"
	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/markes76/cars-il-pp-cli/internal/mcp"
	"github.com/markes76/cars-il-pp-cli/internal/store"
)

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var version = "1.0.0"

func main() {
	db, err := store.Open("")
	if err != nil {
		write(response{JSONRPC: "2.0", Error: map[string]interface{}{"code": -32000, "message": err.Error()}})
		os.Exit(1)
	}
	defer db.Close()
	service := commands.NewService(db)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			write(response{JSONRPC: "2.0", Error: rpcError(-32700, "parse error")})
			continue
		}
		if req.ID == nil && strings.HasPrefix(req.Method, "notifications/") {
			continue
		}
		result, callErr := handle(service, req.Method, req.Params)
		resp := response{JSONRPC: "2.0", ID: req.ID}
		if callErr != nil {
			resp.Error = rpcError(-32603, callErr.Error())
		} else {
			resp.Result = result
		}
		write(resp)
	}
}

func handle(service commands.Service, method string, params json.RawMessage) (interface{}, error) {
	switch method {
	case "initialize":
		return map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
			"serverInfo":      map[string]string{"name": "cars-il-pp-mcp", "version": version},
		}, nil
	case "ping":
		return map[string]interface{}{}, nil
	case "tools/list":
		return map[string]interface{}{"tools": mcp.RegisterTools()}, nil
	case "tools/call":
		var call struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(params, &call); err != nil {
			return nil, err
		}
		value, err := callTool(service, call.Name, call.Arguments)
		if err != nil {
			return mcpText(err.Error(), true), nil
		}
		return mcpJSON(value, false)
	default:
		return nil, fmt.Errorf("unsupported MCP method: %s", method)
	}
}

func callTool(service commands.Service, name string, raw json.RawMessage) (interface{}, error) {
	var params client.SearchParams
	_ = json.Unmarshal(raw, &params)
	switch name {
	case "context":
		return map[string]interface{}{
			"name":       "cars-il-pp-mcp",
			"version":    version,
			"sources":    []string{client.SourceYad2, client.SourceAutoTrader},
			"language":   "he",
			"currency":   "ILS",
			"read_only":  true,
			"safety":     []string{"GET-only remote access", "no phone-number extraction", "no account scraping", "cookies only via environment variables"},
			"autotrader": "Current public autotrader.co.il is a WordPress import/services site, not a public used-car listing catalogue.",
			"usage_hint": "Use search live for discovery, then sync a small cohort before market/deal/price_history analytics.",
		}, nil
	case "search":
		return service.Search(params)
	case "listings_get":
		var in struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(raw, &in)
		return service.Get(in.ID, params.DataSource)
	case "sync":
		return service.Sync(params)
	case "market":
		return service.Market(params)
	case "deal":
		var in struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(raw, &in)
		return service.Deal(in.ID)
	case "price_history":
		var in struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(raw, &in)
		return service.DB.GetPriceHistory(in.ID)
	case "stale":
		var in struct {
			Days int `json:"days"`
		}
		_ = json.Unmarshal(raw, &in)
		if in.Days <= 0 {
			in.Days = 30
		}
		params.DataSource = "local"
		rows, err := service.DB.SearchListings(params)
		if err != nil {
			return nil, err
		}
		var out []client.Listing
		for _, row := range rows {
			if row.DaysOnMarket >= in.Days {
				out = append(out, row)
			}
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].DaysOnMarket > out[j].DaysOnMarket })
		return out, nil
	case "compare":
		ids, err := parseCompareIDs(raw)
		if err != nil {
			return nil, err
		}
		if len(ids) > 5 {
			return nil, client.InvalidArgs("compare supports up to 5 listings")
		}
		var rows []client.Listing
		for _, id := range ids {
			row, err := service.Get(id, params.DataSource)
			if err != nil {
				return nil, err
			}
			rows = append(rows, row)
		}
		return rows, nil
	case "depreciation":
		rows, err := service.DB.SearchListings(client.SearchParams{Make: params.Make, Model: params.Model, Limit: 200})
		if err != nil {
			return nil, err
		}
		type depRow struct {
			Year       int `json:"year"`
			AvgPrice   int `json:"avg_price"`
			Listings   int `json:"listings"`
			AvgMileage int `json:"avg_mileage"`
		}
		group := map[int][]client.Listing{}
		for _, row := range rows {
			if row.Year > 0 {
				group[row.Year] = append(group[row.Year], row)
			}
		}
		var out []depRow
		for year, items := range group {
			var priceSum, priceCount, mileageSum, mileageCount int
			for _, item := range items {
				if item.Price > 0 {
					priceSum += item.Price
					priceCount++
				}
				if item.Mileage > 0 {
					mileageSum += item.Mileage
					mileageCount++
				}
			}
			row := depRow{Year: year, Listings: len(items)}
			if priceCount > 0 {
				row.AvgPrice = priceSum / priceCount
			}
			if mileageCount > 0 {
				row.AvgMileage = mileageSum / mileageCount
			}
			out = append(out, row)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Year < out[j].Year })
		return out, nil
	case "market_heat":
		rows, err := service.DB.SearchListings(client.SearchParams{Limit: 200})
		if err != nil {
			return nil, err
		}
		type cohort struct {
			Key         string  `json:"make_model"`
			AvgDays     float64 `json:"avg_days"`
			Listings    int     `json:"listings"`
			MedianPrice int     `json:"median_price"`
		}
		group := map[string][]client.Listing{}
		for _, row := range rows {
			key := strings.TrimSpace(row.Make + " " + row.Model)
			group[key] = append(group[key], row)
		}
		var out []cohort
		for key, listings := range group {
			stats := commands.ComputeMarketStats(listings, client.SearchParams{})
			out = append(out, cohort{Key: key, AvgDays: stats.AvgDaysOnMarket, Listings: len(listings), MedianPrice: stats.MedianPrice})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].AvgDays < out[j].AvgDays })
		return out, nil
	case "doctor":
		return doctor(service)
	case "watch":
		return nil, client.InvalidArgs("watch is intentionally not run inside MCP; use the CLI for long-running polling")
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func write(resp response) {
	_ = json.NewEncoder(os.Stdout).Encode(resp)
}

func mcpJSON(value interface{}, isError bool) (toolResult, error) {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return toolResult{}, err
	}
	return mcpText(string(body), isError), nil
}

func mcpText(text string, isError bool) toolResult {
	return toolResult{Content: []toolContent{{Type: "text", Text: text}}, IsError: isError}
}

func rpcError(code int, message string) map[string]interface{} {
	return map[string]interface{}{"code": code, "message": message}
}

func parseCompareIDs(raw json.RawMessage) ([]string, error) {
	var root interface{}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, client.InvalidArgs("compare requires ids. Accepted inputs: ids or listing_ids as an array, ids_csv as a comma-separated string")
	}
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, client.InvalidArgs("compare arguments must be a JSON object")
	}
	ids := extractIDs(root, 0)
	if len(ids) == 0 {
		return nil, client.InvalidArgs(fmt.Sprintf("compare requires ids. Accepted inputs: ids/listing_ids as an array, ids_csv as a comma-separated string, or id/listing_id for one listing. Received keys: %s", receivedKeys(root)))
	}
	return uniqueStrings(ids), nil
}

func extractIDs(value interface{}, depth int) []string {
	if depth > 3 {
		return nil
	}
	switch v := value.(type) {
	case map[string]interface{}:
		for _, key := range []string{"ids", "listing_ids", "listingIds", "ids_csv", "id", "listing_id", "listingId"} {
			if candidate, ok := v[key]; ok {
				if ids := extractIDValues(candidate); len(ids) > 0 {
					return ids
				}
			}
		}
		for _, key := range []string{"input", "arguments", "params", "payload"} {
			if nested, ok := v[key]; ok {
				if ids := extractIDs(nested, depth+1); len(ids) > 0 {
					return ids
				}
			}
		}
	}
	return nil
}

func extractIDValues(value interface{}) []string {
	switch v := value.(type) {
	case string:
		return splitIDs(v)
	case []interface{}:
		var ids []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				ids = append(ids, splitIDs(s)...)
			}
		}
		return ids
	default:
		return nil
	}
}

func splitIDs(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "[") {
		var array []string
		if err := json.Unmarshal([]byte(trimmed), &array); err == nil {
			return array
		}
	}
	parts := strings.Split(trimmed, ",")
	var ids []string
	for _, part := range parts {
		if id := strings.TrimSpace(part); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func receivedKeys(value interface{}) string {
	keys := collectKeys(value, 0)
	if len(keys) == 0 {
		return "<none>"
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func collectKeys(value interface{}, depth int) []string {
	if depth > 2 {
		return nil
	}
	m, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	var keys []string
	for key, nested := range m {
		keys = append(keys, key)
		for _, nestedKey := range collectKeys(nested, depth+1) {
			keys = append(keys, key+"."+nestedKey)
		}
	}
	return keys
}

func doctor(service commands.Service) (map[string]interface{}, error) {
	checkURL := func(rawURL string) bool {
		client := http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			return false
		}
		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode > 0 && resp.StatusCode < 500
	}
	count, err := service.DB.CountListings()
	if err != nil {
		return nil, err
	}
	lastSynced, _ := service.DB.LastSyncedAt()
	return map[string]interface{}{
		"yad2_reachable":             checkURL("https://www.yad2.co.il/vehicles/cars"),
		"autotrader_reachable":       checkURL("https://autotrader.co.il/"),
		"listing_count":              count,
		"last_synced_at":             lastSynced,
		"yad2_auth_configured":       os.Getenv("CARS_IL_YAD2_COOKIE") != "",
		"autotrader_auth_configured": os.Getenv("CARS_IL_AUTOTRADER_COOKIE") != "",
		"read_only":                  true,
	}, nil
}
