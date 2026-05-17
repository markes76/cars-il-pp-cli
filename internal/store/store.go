package store

import (
	"database/sql"

	"github.com/markes76/cars-il-pp-cli/internal/client"
)

const StoreSchemaVersion = 1
const StoreDriver = "sqlite"

var migrationStatements = []string{
	`PRAGMA encoding = 'UTF-8'`,
	`PRAGMA busy_timeout = 5000`,
	`PRAGMA foreign_keys = ON`,
	`PRAGMA journal_mode = WAL`,
	`PRAGMA user_version = 1`,
	`CREATE TABLE IF NOT EXISTS listings (
		id TEXT PRIMARY KEY,
		source TEXT,
		make TEXT,
		model TEXT,
		year INTEGER,
		mileage INTEGER,
		price INTEGER,
		city TEXT,
		region TEXT,
		fuel_type TEXT,
		gear_type TEXT,
		color TEXT,
		hand INTEGER,
		is_dealer INTEGER,
		test_expiry TEXT,
		description TEXT,
		image_urls TEXT,
		url TEXT,
		first_seen_at TEXT,
		last_seen_at TEXT,
		price_at_first_seen INTEGER,
		days_on_market INTEGER
	);`,
	`CREATE TABLE IF NOT EXISTS price_history (
		listing_id TEXT,
		recorded_at TEXT,
		price INTEGER,
		PRIMARY KEY (listing_id, recorded_at)
	);`,
	`CREATE TABLE IF NOT EXISTS market_snapshots (
		snapshot_at TEXT,
		make TEXT,
		model TEXT,
		year_min INTEGER,
		year_max INTEGER,
		avg_price INTEGER,
		median_price INTEGER,
		listing_count INTEGER,
		avg_days_on_market REAL
	);`,
	`CREATE TABLE IF NOT EXISTS sync_state (
		source TEXT,
		scope TEXT,
		cursor TEXT,
		updated_at TEXT,
		PRIMARY KEY (source, scope)
	);`,
	`CREATE VIRTUAL TABLE IF NOT EXISTS listings_fts USING fts5(
		id UNINDEXED,
		make,
		model,
		description,
		tokenize='unicode61'
	)`,
}

func (db *DB) SaveSyncState(source, scope, cursor string) error {
	_, err := db.Exec(`INSERT INTO sync_state(source, scope, cursor, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(source, scope) DO UPDATE SET
			cursor=excluded.cursor,
			updated_at=excluded.updated_at`,
		client.NormalizeSource(source), scope, cursor, client.NowISO())
	return err
}

func (db *DB) GetSyncState(source, scope string) (string, error) {
	var cursor sql.NullString
	err := db.QueryRow(`SELECT cursor FROM sync_state WHERE source = ? AND scope = ?`,
		client.NormalizeSource(source), scope).Scan(&cursor)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return cursor.String, nil
}

func (db *DB) ResolveByName(makeName, modelName string) ([]client.Listing, error) {
	return db.SearchListings(client.SearchParams{Make: makeName, Model: modelName, Limit: 20})
}
