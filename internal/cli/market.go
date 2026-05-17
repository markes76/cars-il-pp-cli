package commands

import (
	"fmt"

	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/markes76/cars-il-pp-cli/internal/output"
	"github.com/spf13/cobra"
)

func addMarket(root *cobra.Command, app *App) {
	var params client.SearchParams
	cmd := &cobra.Command{
		Use:     "market",
		Short:   "Aggregate local market intelligence",
		Example: "  cars-il market --make Toyota --model Corolla --year-min 2019 --year-max 2022 --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Limit = 200
			stats, err := app.Service.Market(params)
			if err != nil {
				return err
			}
			if app.JSON {
				return app.formatter().WriteValue(stats)
			}
			printHuman(app.out, app.Quiet, "Market: %s %s %d-%d\n", params.Make, params.Model, params.YearMin, params.YearMax)
			printHuman(app.out, app.Quiet, "Active listings:       %d\n", stats.ActiveListings)
			printHuman(app.out, app.Quiet, "Avg price:             %s\n", output.Shekel(stats.AvgPrice))
			printHuman(app.out, app.Quiet, "Median price:          %s\n", output.Shekel(stats.MedianPrice))
			printHuman(app.out, app.Quiet, "Price range:           %s - %s\n", output.Shekel(stats.MinPrice), output.Shekel(stats.MaxPrice))
			printHuman(app.out, app.Quiet, "Avg mileage:           %d km\n", stats.AvgMileage)
			printHuman(app.out, app.Quiet, "Avg days on market:    %.1f days\n", stats.AvgDaysOnMarket)
			printHuman(app.out, app.Quiet, "Private vs dealer:     %.0f%% private / %.0f%% dealer\n", stats.PrivatePercent, stats.DealerPercent)
			printHuman(app.out, app.Quiet, "Private avg price:     %s\n", output.Shekel(stats.PrivateAvgPrice))
			printHuman(app.out, app.Quiet, "Dealer avg price:      %s\n", output.Shekel(stats.DealerAvgPrice))
			printHuman(app.out, app.Quiet, "Dealer premium:        %+.1f%%\n", stats.DealerPremiumPct)
			printHuman(app.out, app.Quiet, "Most common city:      %s\n", stats.MostCommonCity)
			return nil
		},
	}
	cmd.Flags().StringVar(&params.Make, "make", "", "make")
	cmd.Flags().StringVar(&params.Model, "model", "", "model")
	cmd.Flags().IntVar(&params.YearMin, "year-min", 0, "minimum year")
	cmd.Flags().IntVar(&params.YearMax, "year-max", 0, "maximum year")
	cmd.Flags().StringVar(&params.Fuel, "fuel", "all", "fuel")
	cmd.Flags().StringVar(&params.Gear, "gear", "all", "gear")
	root.AddCommand(cmd)

	heat := &cobra.Command{
		Use:     "market-heat",
		Short:   "Show fastest-moving make/model cohorts",
		Example: "  cars-il market-heat --data-source local --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := app.Service.DB.SearchListings(client.SearchParams{Limit: 200})
			if err != nil {
				return err
			}
			if len(rows) == 0 {
				return client.NotFound("store is empty; run cars-il sync first")
			}
			type cohort struct {
				Key         string  `json:"make_model"`
				AvgDays     float64 `json:"avg_days"`
				Listings    int     `json:"listings"`
				MedianPrice int     `json:"median_price"`
			}
			group := map[string][]client.Listing{}
			for _, row := range rows {
				key := fmt.Sprintf("%s %s", row.Make, row.Model)
				group[key] = append(group[key], row)
			}
			var cohorts []cohort
			for key, listings := range group {
				stats := ComputeMarketStats(listings, client.SearchParams{})
				cohorts = append(cohorts, cohort{Key: key, AvgDays: stats.AvgDaysOnMarket, Listings: len(listings), MedianPrice: stats.MedianPrice})
			}
			return app.formatter().WriteValue(cohorts)
		},
	}
	root.AddCommand(heat)
}
