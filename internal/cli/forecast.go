package commands

import "github.com/spf13/cobra"

func addForecast(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:     "forecast",
		Short:   "Estimate near-term negotiation pressure from stale listings",
		Example: "  cars-il forecast --data-source local --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.formatter().WriteValue(map[string]string{"forecast": "sync more listings for a reliable forecast"})
		},
	}
	root.AddCommand(cmd)
}
