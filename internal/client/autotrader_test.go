package client

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAutoTraderSearchReportsUnavailableWhenNoListingSurfaceExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1537,"title":"Projects","url":"https://autotrader.co.il/projects/","type":"post","subtype":"page"}]`))
	}))
	defer server.Close()

	oldBase := autoTraderBaseURL
	autoTraderBaseURL = server.URL
	defer func() { autoTraderBaseURL = oldBase }()

	_, _, err := NewAutoTraderClient().Search(SearchParams{Make: "Honda", Limit: 5})
	if err == nil {
		t.Fatalf("expected source unavailable error")
	}
	var appErr AppError
	if !errors.As(err, &appErr) || appErr.Code != "SOURCE_UNAVAILABLE" {
		t.Fatalf("expected SOURCE_UNAVAILABLE, got %#v", err)
	}
}
