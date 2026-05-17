package commands

import (
	"sort"

	"github.com/mvanhorn/cars-il-pp-cli/internal/client"
	"github.com/spf13/cobra"
)

func addStale(root *cobra.Command, app *App) {
	var params client.SearchParams
	var days int
	cmd := &cobra.Command{
		Use:     "stale",
		Short:   "List listings unchanged for N+ days",
		Example: "  cars-il stale --days 30 --make Toyota --model Corolla --compact",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := baseParams(app)
			params.Source, params.Limit, params.DataSource = base.Source, base.Limit, "local"
			listings, err := app.Service.DB.SearchListings(params)
			if err != nil {
				return err
			}
			var stale []client.Listing
			for _, listing := range listings {
				if listing.DaysOnMarket >= days {
					stale = append(stale, listing)
				}
			}
			sort.SliceStable(stale, func(i, j int) bool { return stale[i].DaysOnMarket > stale[j].DaysOnMarket })
			if len(stale) == 0 {
				return client.NotFound("no stale listings found")
			}
			if err := app.formatter().WriteListings(stale); err != nil {
				return err
			}
			printHuman(app.out, app.Quiet, "Listings stale %d+ days. Sellers likely motivated. Use `cars-il deal --id <id>` before negotiating.\n", days)
			return nil
		},
	}
	bindSearchFlags(cmd, &params)
	cmd.Flags().IntVar(&days, "days", 30, "minimum days unchanged")
	root.AddCommand(cmd)
}
