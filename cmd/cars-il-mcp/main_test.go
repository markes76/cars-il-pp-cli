package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseCompareIDs(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "ids array", raw: `{"ids":["yad2-a","yad2-b"]}`, want: []string{"yad2-a", "yad2-b"}},
		{name: "listing ids array", raw: `{"listing_ids":["yad2-a","yad2-b"]}`, want: []string{"yad2-a", "yad2-b"}},
		{name: "ids csv", raw: `{"ids_csv":"yad2-a, yad2-b"}`, want: []string{"yad2-a", "yad2-b"}},
		{name: "ids string csv", raw: `{"ids":"yad2-a, yad2-b"}`, want: []string{"yad2-a", "yad2-b"}},
		{name: "nested input", raw: `{"input":{"ids":["yad2-a","yad2-b"]}}`, want: []string{"yad2-a", "yad2-b"}},
		{name: "single id", raw: `{"id":"yad2-a"}`, want: []string{"yad2-a"}},
		{name: "indexed ids", raw: `{"ids[1]":"yad2-b","ids[0]":"yad2-a"}`, want: []string{"yad2-a", "yad2-b"}},
		{name: "indexed listing ids", raw: `{"listing_ids.0":"yad2-a","listing_ids.1":"yad2-b"}`, want: []string{"yad2-a", "yad2-b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCompareIDs(json.RawMessage(tt.raw))
			if err != nil {
				t.Fatalf("parseCompareIDs() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("parseCompareIDs() len = %d, want %d: %v", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("parseCompareIDs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseCompareIDsErrorIncludesReceivedKeys(t *testing.T) {
	_, err := parseCompareIDs(json.RawMessage(`{"listing_ids_wrong":["yad2-a"],"input":{"other":true}}`))
	if err == nil {
		t.Fatal("parseCompareIDs() expected error")
	}
	message := err.Error()
	for _, want := range []string{"compare requires ids", "listing_ids_wrong", "input.other"} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q does not include %q", message, want)
		}
	}
}
