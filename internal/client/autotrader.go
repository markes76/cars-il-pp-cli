package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var autoTraderBaseURL = "https://autotrader.co.il"

type AutoTraderClient struct {
	HTTPClient *http.Client
	Cookie     string
	UserAgent  string
}

func NewAutoTraderClient() *AutoTraderClient {
	return &AutoTraderClient{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Cookie:     os.Getenv("CARS_IL_AUTOTRADER_COOKIE"),
		UserAgent:  randomUserAgent(),
	}
}

func (c *AutoTraderClient) Search(params SearchParams) ([]Listing, PaginationState, error) {
	values := url.Values{}
	query := strings.TrimSpace(params.Make + " " + params.Model)
	if query == "" {
		query = "car"
	}
	values.Set("search", query)
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	values.Set("per_page", strconv.Itoa(limit))
	rawURL := autoTraderBaseURL + "/wp-json/wp/v2/search?" + values.Encode()
	body, err := c.get(rawURL)
	if err != nil {
		return nil, PaginationState{}, err
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, PaginationState{}, APIError("AutoTrader IL returned non-JSON search response")
	}
	listings := make([]Listing, 0)
	for _, row := range rows {
		listing, err := c.NormalizeToUnifiedSchema(row)
		if err == nil && listing.Make != "" && matchesParams(listing, params) {
			listings = append(listings, listing)
		}
	}
	if len(listings) == 0 {
		return nil, PaginationState{Page: 1, HasNext: false, Total: 0}, SourceUnavailable("AutoTrader IL currently exposes WordPress service pages, not a public used-car listing catalogue")
	}
	return listings, PaginationState{Page: 1, HasNext: false, Total: len(listings)}, nil
}

func (c *AutoTraderClient) GetListing(id string) (Listing, error) {
	wpID := strings.TrimPrefix(id, "autotrader-")
	if wpID == "" {
		return Listing{}, InvalidArgs("missing AutoTrader id")
	}
	rawURL := autoTraderBaseURL + "/wp-json/wp/v2/pages/" + url.PathEscape(wpID)
	body, err := c.get(rawURL)
	if err != nil {
		return Listing{}, err
	}
	var row map[string]interface{}
	if err := json.Unmarshal(body, &row); err != nil {
		return Listing{}, APIError("AutoTrader IL page response was not JSON")
	}
	return c.NormalizeToUnifiedSchema(row)
}

func (c *AutoTraderClient) NormalizeToUnifiedSchema(raw interface{}) (Listing, error) {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return Listing{}, APIError("unexpected AutoTrader payload")
	}
	id := intValue(m["id"])
	if id == 0 {
		return Listing{}, APIError("AutoTrader payload missing id")
	}
	title := ""
	if rendered, ok := nested(m, "title", "rendered").(string); ok {
		title = rendered
	} else if s, ok := m["title"].(string); ok {
		title = s
	}
	link := text(m, "url")
	if link == "" {
		link = text(m, "link")
	}
	// The live capture shows autotrader.co.il is a WordPress import/services site,
	// not a listing marketplace. Only normalize pages that visibly contain listing-like
	// data; service pages are intentionally ignored by returning a sparse record.
	listing := Listing{
		ID:          "autotrader-" + strconv.Itoa(id),
		Source:      SourceAutoTrader,
		Description: stripHTML(nestedText(m, "content", "rendered")),
		URL:         link,
		FirstSeenAt: NowISO(),
		LastSeenAt:  NowISO(),
	}
	parts := strings.Fields(title)
	if len(parts) >= 3 {
		if year, err := strconv.Atoi(parts[len(parts)-1]); err == nil && year > 1980 && year < 2100 {
			listing.Year = year
			listing.Model = parts[len(parts)-2]
			listing.Make = strings.Join(parts[:len(parts)-2], " ")
		}
	}
	return listing, nil
}

func (c *AutoTraderClient) get(rawURL string) ([]byte, error) {
	if body, ok := readCache(rawURL, 5*time.Minute); ok {
		return body, nil
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json,text/html;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "he-IL,he;q=0.9,en-US;q=0.8,en;q=0.7")
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.Cookie)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, APIError(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, AuthFailure("AutoTrader IL rejected the request")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		delay := retryAfterDelay(resp)
		if delay == 0 {
			delay = backoff(1)
		}
		return nil, RateLimited("AutoTrader IL rate limited the request; retry after " + delay.String())
	}
	if resp.StatusCode >= 500 {
		return nil, APIError("AutoTrader IL upstream returned " + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = writeCache(rawURL, body)
	return body, nil
}

func stripHTML(input string) string {
	var out strings.Builder
	inTag := false
	for _, r := range input {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				out.WriteRune(r)
			}
		}
	}
	return strings.Join(strings.Fields(out.String()), " ")
}
