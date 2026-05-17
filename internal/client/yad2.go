package client

import (
	"encoding/json"
	"html"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	yad2BaseURL       = "https://www.yad2.co.il"
	yad2VehiclesURL   = "https://www.yad2.co.il/vehicles/cars"
	yad2VehicleDetail = "https://www.yad2.co.il/vehicles/item/"
)

var nextDataRE = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

type Yad2Client struct {
	HTTPClient *http.Client
	Cookie     string
	UserAgent  string
	Delay      time.Duration
	Jitter     float64
}

func NewYad2Client() *Yad2Client {
	return &Yad2Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Cookie:     os.Getenv("CARS_IL_YAD2_COOKIE"),
		UserAgent:  randomUserAgent(),
		Delay:      0,
		Jitter:     0.2,
	}
}

func (c *Yad2Client) Search(params SearchParams) ([]Listing, PaginationState, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	out := make([]Listing, 0, limit)
	total := 0
	hasNext := false
	startPage := params.Page
	if startPage <= 0 {
		startPage = 1
	}
	maxPages := 5
	for page := startPage; page < startPage+maxPages && len(out) < limit; page++ {
		pageParams := params
		pageParams.Page = page
		u, err := c.searchURL(pageParams)
		if err != nil {
			return nil, PaginationState{}, err
		}
		body, err := c.get(u)
		if err != nil {
			return nil, PaginationState{}, err
		}
		data, err := extractNextData(body)
		if err != nil {
			return nil, PaginationState{}, err
		}
		rawListings, pageTotal := extractYad2Feed(data)
		if pageTotal > total {
			total = pageTotal
		}
		if len(rawListings) == 0 {
			break
		}
		for _, raw := range rawListings {
			listing, err := c.NormalizeToUnifiedSchema(raw)
			if err != nil {
				continue
			}
			if !matchesParams(listing, params) {
				continue
			}
			if len(out) < limit {
				if detail, err := c.GetListing(strings.TrimPrefix(listing.ID, "yad2-")); err == nil {
					listing = detail
				}
				if matchesParams(listing, params) {
					out = append(out, listing)
				}
			}
		}
		hasNext = total > page*len(rawListings)
	}
	return out, PaginationState{Page: startPage, HasNext: hasNext, Total: total}, nil
}

func (c *Yad2Client) GetListing(id string) (Listing, error) {
	token := strings.TrimPrefix(id, "yad2-")
	if token == "" {
		return Listing{}, InvalidArgs("missing Yad2 listing id")
	}
	body, err := c.get(yad2VehicleDetail + url.PathEscape(token))
	if err != nil {
		return Listing{}, err
	}
	data, err := extractNextData(body)
	if err != nil {
		return Listing{}, err
	}
	raw, ok := firstYad2Detail(data)
	if !ok {
		return Listing{}, NotFound("listing not found in Yad2 page payload")
	}
	return c.NormalizeToUnifiedSchema(raw)
}

func (c *Yad2Client) NormalizeToUnifiedSchema(raw interface{}) (Listing, error) {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return Listing{}, APIError("unexpected Yad2 listing payload")
	}
	token := text(m, "token")
	if token == "" {
		return Listing{}, APIError("Yad2 listing missing token")
	}
	sourceID := "yad2-" + token
	price := intValue(m["price"])
	created := parseYad2Time(nestedText(m, "dates", "createdAt"))
	updated := parseYad2Time(nestedText(m, "dates", "updatedAt"))
	if created == "" {
		created = NowISO()
	}
	if updated == "" {
		updated = NowISO()
	}
	imageURLs := stringSlice(nested(m, "metaData", "images"))
	testExpiry := parseYad2Date(nestedText(m, "vehicleDates", "testDate"))
	hand := intValue(nested(m, "hand", "id"))
	makeName := nestedText(m, "manufacturer", "text")
	if makeName == "" {
		makeName = nestedText(m, "manufacturer", "textEng")
	}
	listing := Listing{
		ID:               sourceID,
		Source:           SourceYad2,
		Make:             makeName,
		Model:            nestedText(m, "model", "text"),
		Year:             intValue(nested(m, "vehicleDates", "yearOfProduction")),
		Mileage:          intValue(m["km"]),
		Price:            price,
		City:             nestedText(m, "address", "city", "text"),
		Region:           firstNonEmpty(nestedText(m, "address", "topArea", "text"), nestedText(m, "address", "area", "text")),
		FuelType:         nestedText(m, "engineType", "text"),
		GearType:         nestedText(m, "gearBox", "text"),
		Color:            nestedText(m, "color", "text"),
		Hand:             hand,
		IsDealer:         text(m, "adType") != "private" || nestedText(m, "customer", "agencyName") != "",
		TestExpiry:       testExpiry,
		Description:      nestedText(m, "metaData", "description"),
		ImageURLs:        imageURLs,
		URL:              yad2VehicleDetail + token,
		FirstSeenAt:      created,
		LastSeenAt:       updated,
		PriceAtFirstSeen: price,
		DaysOnMarket:     DaysBetween(created, updated),
	}
	if listing.City == "" {
		listing.City = nestedText(m, "address", "area", "text")
	}
	return listing, nil
}

func (c *Yad2Client) searchURL(params SearchParams) (string, error) {
	values := url.Values{}
	if params.YearMin > 0 || params.YearMax > 0 {
		minYear := params.YearMin
		maxYear := params.YearMax
		if minYear == 0 {
			minYear = 1950
		}
		if maxYear == 0 {
			maxYear = time.Now().Year() + 1
		}
		values.Set("year", strconv.Itoa(minYear)+"-"+strconv.Itoa(maxYear))
	}
	if id := yad2ManufacturerID(params.Make); id != "" {
		values.Set("manufacturer", id)
	} else if params.Make != "" {
		values.Set("text", params.Make)
	}
	if params.Model != "" {
		existing := values.Get("text")
		model := NormalizeModel(params.Model)
		if existing != "" {
			values.Set("text", strings.TrimSpace(existing+" "+model))
		} else {
			values.Set("text", model)
		}
	}
	if area := yad2AreaID(firstNonEmpty(params.Region, params.City)); area != "" {
		values.Set("area", area)
	}
	if params.Page > 1 {
		values.Set("page", strconv.Itoa(params.Page))
	}
	if values.Encode() == "" {
		return yad2VehiclesURL, nil
	}
	return yad2VehiclesURL + "?" + values.Encode(), nil
}

func (c *Yad2Client) get(rawURL string) ([]byte, error) {
	if body, ok := readCache(rawURL, 5*time.Minute); ok {
		return body, nil
	}
	if c.Delay > 0 {
		sleep := c.Delay
		if c.Jitter > 0 {
			delta := int64(float64(c.Delay) * c.Jitter)
			if delta > 0 {
				sleep = c.Delay + time.Duration(rand.Int63n(delta*2)-delta)
			}
		}
		time.Sleep(sleep)
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,application/json;q=0.8,*/*;q=0.7")
	req.Header.Set("Accept-Language", "he-IL,he;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Referer", yad2VehiclesURL)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-CH-UA", `"Chromium";v="125", "Google Chrome";v="125", "Not-A.Brand";v="99"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"macOS"`)
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.Cookie)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, APIError(err.Error())
	}
	if resp.StatusCode == http.StatusBadRequest && c.Cookie != "" {
		_ = resp.Body.Close()
		noCookie := *c
		noCookie.Cookie = ""
		return noCookie.get(rawURL)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, AuthFailure("Yad2 rejected the request; provide CARS_IL_YAD2_COOKIE or refresh the browser session")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		delay := retryAfterDelay(resp)
		if delay == 0 {
			delay = backoff(1)
		}
		return nil, RateLimited("Yad2 rate limited the request; retry after " + delay.String())
	}
	if resp.StatusCode >= 500 {
		return nil, APIError("Yad2 upstream returned " + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if debugPath := os.Getenv("CARS_IL_DEBUG_HTML"); debugPath != "" {
		_ = os.WriteFile(debugPath, body, 0o600)
	}
	if strings.Contains(string(body), "ShieldSquare Captcha") || strings.Contains(string(body), "Are you for real") {
		return nil, AuthFailure("Yad2 returned a bot-check challenge; use Playwright capture or browser cookies")
	}
	_ = writeCache(rawURL, body)
	return body, nil
}

func extractNextData(body []byte) (map[string]interface{}, error) {
	bodyText := string(body)
	if strings.Contains(bodyText, "ShieldSquare Captcha") || strings.Contains(bodyText, "hcaptcha") || strings.Contains(bodyText, "Are you for real") {
		return nil, AuthFailure("Yad2 returned a bot-check challenge; provide CARS_IL_YAD2_COOKIE from a browser session")
	}
	match := nextDataRE.FindSubmatch(body)
	if len(match) < 2 {
		return nil, AuthFailure("Yad2 page did not include listing data; provide CARS_IL_YAD2_COOKIE from a browser session")
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(html.UnescapeString(string(match[1]))), &data); err != nil {
		return nil, APIError("could not parse __NEXT_DATA__: " + err.Error())
	}
	return data, nil
}

func extractYad2Feed(next map[string]interface{}) ([]interface{}, int) {
	var out []interface{}
	total := 0
	for _, q := range yad2Queries(next) {
		query, _ := q.(map[string]interface{})
		data, _ := nested(query, "state", "data").(map[string]interface{})
		if data == nil {
			continue
		}
		for _, key := range []string{"platinum", "boost", "solo", "commercial", "private"} {
			if rows, ok := data[key].([]interface{}); ok {
				out = append(out, rows...)
			}
		}
		if pagination, ok := data["pagination"].(map[string]interface{}); ok {
			total = intValue(pagination["total"])
		}
		if len(out) > 0 {
			return out, total
		}
	}
	return out, total
}

func firstYad2Detail(next map[string]interface{}) (map[string]interface{}, bool) {
	for _, q := range yad2Queries(next) {
		query, _ := q.(map[string]interface{})
		data, _ := nested(query, "state", "data").(map[string]interface{})
		if data == nil {
			continue
		}
		if text(data, "token") != "" {
			return data, true
		}
	}
	return nil, false
}

func yad2Queries(next map[string]interface{}) []interface{} {
	queries, _ := nested(next, "props", "pageProps", "dehydratedState", "queries").([]interface{})
	return queries
}

func yad2ManufacturerID(make string) string {
	switch NormalizeMake(make) {
	case "טויוטה":
		return "19"
	case "מאזדה":
		return "27"
	case "קיה":
		return "48"
	case "אופל":
		return "2"
	case "סובארו":
		return "35"
	case "סוזוקי":
		return "36"
	case "פורד":
		return "43"
	default:
		return ""
	}
}

func yad2AreaID(region string) string {
	switch strings.TrimSpace(region) {
	case "תל אביב", "תל אביב יפו", "Tel Aviv", "tel aviv", "אזור תל אביב יפו":
		return "1"
	case "רעננה", "כפר סבא", "שרון":
		return "42"
	case "חיפה", "Haifa":
		return "15"
	default:
		return ""
	}
}

func matchesParams(l Listing, p SearchParams) bool {
	if p.Make != "" {
		want := NormalizeMake(p.Make)
		got := NormalizeMake(l.Make)
		if !strings.Contains(strings.ToLower(got), strings.ToLower(want)) && !strings.Contains(strings.ToLower(l.Make), strings.ToLower(p.Make)) {
			return false
		}
	}
	if p.Model != "" {
		ok := false
		for _, alt := range ModelAlternates(p.Model) {
			if containsFold(l.Model, alt) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if (p.YearMin > 0 || p.YearMax > 0) && l.Year == 0 {
		return false
	}
	if p.YearMin > 0 && l.Year < p.YearMin {
		return false
	}
	if p.YearMax > 0 && l.Year > p.YearMax {
		return false
	}
	if (p.PriceMin > 0 || p.PriceMax > 0) && l.Price == 0 {
		return false
	}
	if p.PriceMin > 0 && l.Price < p.PriceMin {
		return false
	}
	if p.PriceMax > 0 && l.Price > p.PriceMax {
		return false
	}
	if p.MileageMax > 0 && l.Mileage == 0 {
		return false
	}
	if p.MileageMax > 0 && l.Mileage > p.MileageMax {
		return false
	}
	if p.City != "" && l.City != "" && !containsFold(l.City, p.City) {
		return false
	}
	if p.Region != "" && l.Region != "" && !containsFold(l.Region, p.Region) {
		return false
	}
	if p.Fuel != "" && p.Fuel != "all" && l.FuelType != "" && !fuelMatches(l.FuelType, p.Fuel) {
		return false
	}
	if p.Gear != "" && p.Gear != "all" && l.GearType != "" && !gearMatches(l.GearType, p.Gear) {
		return false
	}
	if p.HandMax > 0 && l.Hand > 0 && l.Hand > p.HandMax {
		return false
	}
	if p.PrivateOnly && l.IsDealer {
		return false
	}
	if p.DealerOnly && !l.IsDealer {
		return false
	}
	return true
}

func fuelMatches(hebrew, flag string) bool {
	h := strings.ToLower(hebrew)
	switch flag {
	case "petrol":
		return strings.Contains(h, "בנזין")
	case "diesel":
		return strings.Contains(h, "דיזל")
	case "hybrid":
		return strings.Contains(h, "היברידי")
	case "electric":
		return strings.Contains(h, "חשמלי")
	default:
		return containsFold(hebrew, flag)
	}
}

func gearMatches(hebrew, flag string) bool {
	h := strings.ToLower(hebrew)
	switch flag {
	case "auto":
		return strings.Contains(h, "אוט")
	case "manual":
		return strings.Contains(h, "ידני")
	default:
		return containsFold(hebrew, flag)
	}
}

func containsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func parseYad2Time(value string) string {
	if value == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	if t, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	return value
}

func parseYad2Date(value string) string {
	if value == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.Format("2006-01-02")
	}
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}

func randomUserAgent() string {
	agents := []string{
		"Mozilla/5.0 (Macintosh; ARM Mac OS X 15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; ARM Mac OS X 15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	}
	return agents[rand.Intn(len(agents))]
}

func nested(m map[string]interface{}, path ...string) interface{} {
	var cur interface{} = m
	for _, key := range path {
		asMap, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur = asMap[key]
	}
	return cur
}

func text(m map[string]interface{}, key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}

func nestedText(m map[string]interface{}, path ...string) string {
	if value, ok := nested(m, path...).(string); ok {
		return value
	}
	return ""
}

func intValue(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	case string:
		clean := strings.NewReplacer(",", "", "₪", "", " ", "").Replace(v)
		i, _ := strconv.Atoi(clean)
		return i
	default:
		return 0
	}
}

func stringSlice(value interface{}) []string {
	rows, ok := value.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if s, ok := row.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
