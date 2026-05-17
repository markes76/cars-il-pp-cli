package commands

import (
	"github.com/mvanhorn/cars-il-pp-cli/internal/client"
	"github.com/mvanhorn/cars-il-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

var defaultSyncResources = []string{"listings", "price_history", "market_snapshots", "sync_state"}

func addSync(root *cobra.Command, app *App) {
	var params client.SearchParams
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Pull listings into local SQLite",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := baseParams(app)
			params.Source, params.Limit, params.DataSource = base.Source, base.Limit, "live"
			if app.DryRun {
				yad2Count := params.Limit
				autoCount := int(float64(params.Limit) * 0.6)
				printHuman(app.out, app.Quiet, "Would sync %d listings from yad2, %d from autotrader\n", yad2Count, autoCount)
				app.ExitCode = client.ExitDryRun
				return nil
			}
			result, err := app.Service.Sync(params)
			if err != nil {
				return err
			}
			printHuman(app.out, app.Quiet, "Synced %d listings (%d new, %d price changes, %d removed)\n", result.Total, result.New, result.PriceChanges, result.Removed)
			if app.JSON {
				return app.formatter().WriteValue(result)
			}
			return nil
		},
	}
	bindSearchFlags(cmd, &params)
	cmd.Flags().IntVar(&params.MaxDaily, "max-daily", 500, "daily listing fetch cap per source")
	root.AddCommand(cmd)
}

func syncDomainUpsertListing(db *store.DB, listing client.Listing) error {
	return db.UpsertListing(listing)
}
