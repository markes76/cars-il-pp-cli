package store

type MarketSnapshot struct {
	SnapshotAt      string  `json:"snapshot_at"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	YearMin         int     `json:"year_min"`
	YearMax         int     `json:"year_max"`
	AvgPrice        int     `json:"avg_price"`
	MedianPrice     int     `json:"median_price"`
	ListingCount    int     `json:"listing_count"`
	AvgDaysOnMarket float64 `json:"avg_days_on_market"`
}

func (db *DB) UpsertSnapshot(snapshot MarketSnapshot) error {
	_, err := db.Exec(`INSERT INTO market_snapshots(snapshot_at, make, model, year_min, year_max, avg_price, median_price, listing_count, avg_days_on_market)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		snapshot.SnapshotAt, snapshot.Make, snapshot.Model, snapshot.YearMin, snapshot.YearMax, snapshot.AvgPrice,
		snapshot.MedianPrice, snapshot.ListingCount, snapshot.AvgDaysOnMarket)
	return err
}

func (db *DB) GetMarketSnapshot(make, model string, yearMin, yearMax int) (MarketSnapshot, error) {
	var snapshot MarketSnapshot
	err := db.QueryRow(`SELECT snapshot_at, make, model, year_min, year_max, avg_price, median_price, listing_count, avg_days_on_market
		FROM market_snapshots WHERE make LIKE ? AND model LIKE ? AND year_min = ? AND year_max = ?
		ORDER BY snapshot_at DESC LIMIT 1`, "%"+make+"%", "%"+model+"%", yearMin, yearMax).
		Scan(&snapshot.SnapshotAt, &snapshot.Make, &snapshot.Model, &snapshot.YearMin, &snapshot.YearMax,
			&snapshot.AvgPrice, &snapshot.MedianPrice, &snapshot.ListingCount, &snapshot.AvgDaysOnMarket)
	return snapshot, err
}
