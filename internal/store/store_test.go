package store

import (
	"path/filepath"
	"testing"

	"github.com/markes76/cars-il-pp-cli/internal/client"
)

func TestUpsertAndSearchHebrewListing(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "cars.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	listing := client.Listing{
		ID:               "yad2-test",
		Source:           "yad2",
		Make:             "טויוטה",
		Model:            "קורולה",
		Year:             2020,
		Mileage:          65000,
		Price:            89000,
		City:             "תל אביב",
		Region:           "מרכז",
		FuelType:         "בנזין",
		GearType:         "אוטומטי",
		Hand:             2,
		Description:      "רכב שמור במצב מצוין",
		URL:              "https://www.yad2.co.il/vehicles/item/test",
		PriceAtFirstSeen: 89000,
	}
	if err := db.UpsertListing(listing); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.SearchListings(client.SearchParams{Make: "Toyota", Model: "קורולה", City: "תל אביב", Limit: 10})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected one listing, got %d", len(got))
	}
	if got[0].City != "תל אביב" || got[0].Make != "טויוטה" {
		t.Fatalf("hebrew text was not preserved: %#v", got[0])
	}

	englishModel, err := db.SearchListings(client.SearchParams{Make: "Toyota", Model: "Corolla", Limit: 10})
	if err != nil {
		t.Fatalf("english model search: %v", err)
	}
	if len(englishModel) != 1 {
		t.Fatalf("expected English Corolla alias to find one listing, got %d", len(englishModel))
	}

	fts, err := db.SearchListings(client.SearchParams{Query: "מצוין", Limit: 10})
	if err != nil {
		t.Fatalf("fts search: %v", err)
	}
	if len(fts) != 1 {
		t.Fatalf("expected Hebrew FTS to find one listing, got %d", len(fts))
	}
}

func TestRecordPriceOnlyOnChange(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "cars.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	listing := client.Listing{ID: "yad2-price", Source: "yad2", Make: "Mazda", Model: "3", Year: 2021, Price: 90000}
	if err := db.UpsertListing(listing); err != nil {
		t.Fatalf("upsert initial: %v", err)
	}
	listing.Price = 87000
	if err := db.UpsertListing(listing); err != nil {
		t.Fatalf("upsert changed: %v", err)
	}
	history, err := db.GetPriceHistory("yad2-price")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(history) < 2 {
		t.Fatalf("expected at least two price points, got %d", len(history))
	}
}

func TestSyncStateRoundTrip(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "cars.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.SaveSyncState("yad2", "toyota-corolla", "page=2"); err != nil {
		t.Fatalf("save sync state: %v", err)
	}
	cursor, err := db.GetSyncState("yad2", "toyota-corolla")
	if err != nil {
		t.Fatalf("get sync state: %v", err)
	}
	if cursor != "page=2" {
		t.Fatalf("expected page=2, got %q", cursor)
	}
}
