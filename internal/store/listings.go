package store

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/markes76/cars-il-pp-cli/internal/client"
)

func (db *DB) UpsertListing(listing client.Listing) error {
	now := client.NowISO()
	if listing.FirstSeenAt == "" {
		listing.FirstSeenAt = now
	}
	if listing.LastSeenAt == "" {
		listing.LastSeenAt = now
	}
	if listing.PriceAtFirstSeen == 0 {
		listing.PriceAtFirstSeen = listing.Price
	}
	var existingPrice int
	err := db.QueryRow(`SELECT price FROM listings WHERE id = ?`, listing.ID).Scan(&existingPrice)
	if err == sql.ErrNoRows {
		if err := db.RecordPrice(listing.ID, listing.Price, listing.FirstSeenAt); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if existingPrice != listing.Price {
		if err := db.RecordPrice(listing.ID, listing.Price, now); err != nil {
			return err
		}
	}
	images, err := json.Marshal(listing.ImageURLs)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO listings (
		id, source, make, model, year, mileage, price, city, region, fuel_type, gear_type, color,
		hand, is_dealer, test_expiry, description, image_urls, url, first_seen_at, last_seen_at,
		price_at_first_seen, days_on_market
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		source=excluded.source,
		make=excluded.make,
		model=excluded.model,
		year=excluded.year,
		mileage=excluded.mileage,
		price=excluded.price,
		city=excluded.city,
		region=excluded.region,
		fuel_type=excluded.fuel_type,
		gear_type=excluded.gear_type,
		color=excluded.color,
		hand=excluded.hand,
		is_dealer=excluded.is_dealer,
		test_expiry=excluded.test_expiry,
		description=excluded.description,
		image_urls=excluded.image_urls,
		url=excluded.url,
		last_seen_at=excluded.last_seen_at,
		days_on_market=excluded.days_on_market`,
		listing.ID, listing.Source, listing.Make, listing.Model, listing.Year, listing.Mileage, listing.Price,
		listing.City, listing.Region, listing.FuelType, listing.GearType, listing.Color, listing.Hand,
		boolInt(listing.IsDealer), listing.TestExpiry, listing.Description, string(images), listing.URL,
		listing.FirstSeenAt, listing.LastSeenAt, listing.PriceAtFirstSeen, listing.DaysOnMarket,
	)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO listings_fts(id, make, model, description)
		VALUES (?, ?, ?, ?)
		ON CONFLICT DO UPDATE SET make=excluded.make, model=excluded.model, description=excluded.description`,
		listing.ID, listing.Make, listing.Model, listing.Description)
	if err != nil {
		_, _ = db.Exec(`DELETE FROM listings_fts WHERE id = ?`, listing.ID)
		_, err = db.Exec(`INSERT INTO listings_fts(id, make, model, description) VALUES (?, ?, ?, ?)`, listing.ID, listing.Make, listing.Model, listing.Description)
	}
	return err
}

func (db *DB) SearchListings(params client.SearchParams) ([]client.Listing, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	where := []string{"1=1"}
	args := []interface{}{}
	addLike := func(column, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		where = append(where, column+" LIKE ?")
		args = append(args, "%"+value+"%")
	}
	if params.Source != "" && params.Source != client.SourceAll {
		where = append(where, "source = ?")
		args = append(args, client.NormalizeSource(params.Source))
	}
	if params.Make != "" {
		alts := client.MakeAlternates(params.Make)
		if len(alts) == 0 {
			addLike("make", params.Make)
		} else {
			parts := make([]string, 0, len(alts))
			for _, alt := range alts {
				parts = append(parts, "make LIKE ?")
				args = append(args, "%"+alt+"%")
			}
			where = append(where, "("+strings.Join(parts, " OR ")+")")
		}
	}
	if strings.TrimSpace(params.Query) != "" {
		where = append(where, "id IN (SELECT id FROM listings_fts WHERE listings_fts MATCH ?)")
		args = append(args, ftsQuery(params.Query))
	}
	if params.Model != "" {
		alts := client.ModelAlternates(params.Model)
		if len(alts) == 0 {
			addLike("model", params.Model)
		} else {
			parts := make([]string, 0, len(alts))
			for _, alt := range alts {
				parts = append(parts, "model LIKE ?")
				args = append(args, "%"+alt+"%")
			}
			where = append(where, "("+strings.Join(parts, " OR ")+")")
		}
	}
	addLike("city", params.City)
	addLike("region", params.Region)
	if params.YearMin > 0 {
		where = append(where, "year >= ?")
		args = append(args, params.YearMin)
	}
	if params.YearMax > 0 {
		where = append(where, "year <= ?")
		args = append(args, params.YearMax)
	}
	if params.PriceMin > 0 {
		where = append(where, "price >= ?")
		args = append(args, params.PriceMin)
	}
	if params.PriceMax > 0 {
		where = append(where, "price <= ?")
		args = append(args, params.PriceMax)
	}
	if params.MileageMax > 0 {
		where = append(where, "mileage > 0 AND mileage <= ?")
		args = append(args, params.MileageMax)
	}
	if params.HandMax > 0 {
		where = append(where, "hand > 0 AND hand <= ?")
		args = append(args, params.HandMax)
	}
	if params.PrivateOnly {
		where = append(where, "is_dealer = 0")
	}
	if params.DealerOnly {
		where = append(where, "is_dealer = 1")
	}
	if params.Fuel != "" && params.Fuel != "all" {
		addLike("fuel_type", mapFuel(params.Fuel))
	}
	if params.Gear != "" && params.Gear != "all" {
		addLike("gear_type", mapGear(params.Gear))
	}
	order := sortSQL(params.Sort)
	query := `SELECT id, source, make, model, year, mileage, price, city, region, fuel_type, gear_type, color,
		hand, is_dealer, test_expiry, description, image_urls, url, first_seen_at, last_seen_at,
		price_at_first_seen, days_on_market
		FROM listings WHERE ` + strings.Join(where, " AND ") + " " + order + " LIMIT ?"
	args = append(args, limit)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListings(rows)
}

func (db *DB) GetListing(id string) (client.Listing, error) {
	rows, err := db.Query(`SELECT id, source, make, model, year, mileage, price, city, region, fuel_type, gear_type, color,
		hand, is_dealer, test_expiry, description, image_urls, url, first_seen_at, last_seen_at,
		price_at_first_seen, days_on_market
		FROM listings WHERE id = ? LIMIT 1`, id)
	if err != nil {
		return client.Listing{}, err
	}
	defer rows.Close()
	listings, err := scanListings(rows)
	if err != nil {
		return client.Listing{}, err
	}
	if len(listings) == 0 {
		return client.Listing{}, client.NotFound("listing not found: " + id)
	}
	return listings[0], nil
}

func (db *DB) CountListings() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM listings`).Scan(&count)
	return count, err
}

func (db *DB) LastSyncedAt() (string, error) {
	var value sql.NullString
	err := db.QueryRow(`SELECT MAX(last_seen_at) FROM listings`).Scan(&value)
	if err != nil || !value.Valid {
		return "", err
	}
	return value.String, nil
}

func scanListings(rows *sql.Rows) ([]client.Listing, error) {
	var out []client.Listing
	for rows.Next() {
		var listing client.Listing
		var images string
		var isDealer int
		err := rows.Scan(&listing.ID, &listing.Source, &listing.Make, &listing.Model, &listing.Year, &listing.Mileage,
			&listing.Price, &listing.City, &listing.Region, &listing.FuelType, &listing.GearType, &listing.Color,
			&listing.Hand, &isDealer, &listing.TestExpiry, &listing.Description, &images, &listing.URL,
			&listing.FirstSeenAt, &listing.LastSeenAt, &listing.PriceAtFirstSeen, &listing.DaysOnMarket)
		if err != nil {
			return nil, err
		}
		listing.IsDealer = isDealer == 1
		_ = json.Unmarshal([]byte(images), &listing.ImageURLs)
		out = append(out, listing)
	}
	return out, rows.Err()
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func sortSQL(sortBy string) string {
	switch sortBy {
	case "price-asc":
		return "ORDER BY price ASC"
	case "price-desc":
		return "ORDER BY price DESC"
	case "date-old":
		return "ORDER BY first_seen_at ASC"
	case "mileage-asc":
		return "ORDER BY mileage ASC"
	case "days-asc":
		return "ORDER BY days_on_market ASC"
	default:
		return "ORDER BY last_seen_at DESC"
	}
}

func mapFuel(flag string) string {
	switch flag {
	case "petrol":
		return "בנזין"
	case "diesel":
		return "דיזל"
	case "hybrid":
		return "היברידי"
	case "electric":
		return "חשמלי"
	default:
		return flag
	}
}

func mapGear(flag string) string {
	switch flag {
	case "auto":
		return "אוט"
	case "manual":
		return "ידני"
	default:
		return flag
	}
}

func ftsQuery(value string) string {
	terms := strings.Fields(value)
	for i, term := range terms {
		terms[i] = `"` + strings.ReplaceAll(term, `"`, `""`) + `"`
	}
	return strings.Join(terms, " ")
}
