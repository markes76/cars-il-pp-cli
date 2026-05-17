package store

import "github.com/mvanhorn/cars-il-pp-cli/internal/client"

func (db *DB) RecordPrice(listingID string, price int, recordedAt string) error {
	if recordedAt == "" {
		recordedAt = client.NowISO()
	}
	_, err := db.Exec(`INSERT OR IGNORE INTO price_history(listing_id, recorded_at, price) VALUES (?, ?, ?)`, listingID, recordedAt, price)
	return err
}

func (db *DB) GetPriceHistory(listingID string) ([]client.PricePoint, error) {
	rows, err := db.Query(`SELECT listing_id, recorded_at, price FROM price_history WHERE listing_id = ? ORDER BY recorded_at ASC`, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []client.PricePoint
	for rows.Next() {
		var point client.PricePoint
		if err := rows.Scan(&point.ListingID, &point.RecordedAt, &point.Price); err != nil {
			return nil, err
		}
		out = append(out, point)
	}
	return out, rows.Err()
}
