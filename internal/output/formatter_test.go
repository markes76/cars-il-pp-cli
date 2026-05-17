package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mvanhorn/cars-il-pp-cli/internal/client"
)

func TestFormatterPreservesHebrewAndJSON(t *testing.T) {
	var out bytes.Buffer
	formatter := Formatter{Format: "json", Writer: &out}
	err := formatter.WriteListings([]client.Listing{{ID: "yad2-1", Make: "טויוטה", Model: "קורולה", City: "תל אביב", Price: 89000}})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(out.String(), "תל אביב") {
		t.Fatalf("hebrew missing from output: %s", out.String())
	}
	var decoded []client.Listing
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
}

func TestCompactTableUsesHighGravityFields(t *testing.T) {
	var out bytes.Buffer
	formatter := Formatter{Format: "table", Compact: true, Writer: &out}
	err := formatter.WriteListings([]client.Listing{{ID: "yad2-1", Make: "Toyota", Model: "Corolla", Year: 2020, Price: 89000, City: "תל אביב", Mileage: 65000}})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "ID") || !strings.Contains(text, "Price") {
		t.Fatalf("missing compact headers: %s", text)
	}
	if strings.Contains(text, "Mileage") {
		t.Fatalf("compact output should omit mileage: %s", text)
	}
}
