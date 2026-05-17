package commands

import "github.com/spf13/cobra"

func addTrends(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:     "trends",
		Short:   "Summarize local price and demand trends by model",
		Example: "  cars-il trends --source yad2 --limit 100 --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := app.Service.DB.SearchListings(baseParams(app))
			if err != nil {
				return err
			}
			stats := ComputeMarketStats(rows, baseParams(app))
			return app.formatter().WriteValue(stats)
		},
	}
	root.AddCommand(cmd)
}
