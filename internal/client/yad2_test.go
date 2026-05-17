package client

import (
	"strings"
	"testing"
)

func TestYad2SearchURLNormalizesModelAndAddsPage(t *testing.T) {
	u, err := NewYad2Client().searchURL(SearchParams{Make: "Toyota", Model: "Corolla", Page: 2})
	if err != nil {
		t.Fatalf("search url: %v", err)
	}
	if !strings.Contains(u, "manufacturer=19") {
		t.Fatalf("expected Toyota manufacturer id in %s", u)
	}
	if !strings.Contains(u, "%D7%A7%D7%95%D7%A8%D7%95%D7%9C%D7%94") {
		t.Fatalf("expected Hebrew Corolla query in %s", u)
	}
	if !strings.Contains(u, "page=2") {
		t.Fatalf("expected page=2 in %s", u)
	}
}
