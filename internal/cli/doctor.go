package commands

import (
	"net/http"
	"os"
	"time"

	"github.com/mvanhorn/cars-il-pp-cli/internal/client"
	"github.com/spf13/cobra"
)

type DoctorResult struct {
	Yad2Reachable       bool   `json:"yad2_reachable"`
	AutoTraderReachable bool   `json:"autotrader_reachable"`
	StorePath           string `json:"store_path"`
	StoreReadable       bool   `json:"store_readable"`
	ListingCount        int    `json:"listing_count"`
	LastSyncedAt        string `json:"last_synced_at,omitempty"`
	Version             string `json:"version"`
	Yad2AuthConfigured  bool   `json:"yad2_auth_configured"`
	AutoAuthConfigured  bool   `json:"autotrader_auth_configured"`
	Warning             string `json:"warning,omitempty"`
}

func addDoctor(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Connectivity and local store health check",
		RunE: func(cmd *cobra.Command, args []string) error {
			result := DoctorResult{StorePath: app.Service.DB.Path(), StoreReadable: true, Version: "1.0.0"}
			result.Yad2Reachable = canReach("https://www.yad2.co.il/vehicles/cars")
			result.AutoTraderReachable = canReach("https://autotrader.co.il/")
			result.Yad2AuthConfigured = os.Getenv("CARS_IL_YAD2_COOKIE") != ""
			result.AutoAuthConfigured = os.Getenv("CARS_IL_AUTOTRADER_COOKIE") != ""
			count, err := app.Service.DB.CountListings()
			if err != nil {
				result.StoreReadable = false
			}
			result.ListingCount = count
			last, _ := app.Service.DB.LastSyncedAt()
			result.LastSyncedAt = last
			_ = collectCacheReport()
			if count == 0 {
				result.Warning = "store is empty: run cars-il sync first"
			}
			if app.JSON {
				return app.formatter().WriteValue(result)
			}
			printHuman(app.out, app.Quiet, "yad2.co.il reachable:       %v\n", result.Yad2Reachable)
			printHuman(app.out, app.Quiet, "autotrader.co.il reachable: %v\n", result.AutoTraderReachable)
			printHuman(app.out, app.Quiet, "store:                      %s\n", result.StorePath)
			printHuman(app.out, app.Quiet, "config/auth:                yad2=%v autotrader=%v\n", result.Yad2AuthConfigured, result.AutoAuthConfigured)
			printHuman(app.out, app.Quiet, "version:                    %s\n", result.Version)
			printHuman(app.out, app.Quiet, "listings:                   %d\n", result.ListingCount)
			if result.Warning != "" {
				printHuman(app.out, app.Quiet, "warning:                    %s\n", result.Warning)
			}
			if !result.Yad2Reachable || !result.AutoTraderReachable {
				return client.APIError("one or more sources are unreachable")
			}
			return nil
		},
	}
	root.AddCommand(cmd)
}

func collectCacheReport() map[string]string {
	return map[string]string{"freshness": "run sync to refresh local listings"}
}

func renderCacheReport() string {
	return "cache freshness is available through cars-il doctor"
}

func canReach(rawURL string) bool {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 cars-il doctor")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}
