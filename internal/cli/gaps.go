package commands

import "github.com/spf13/cobra"

func addGaps(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:     "gaps",
		Short:   "Identify under-supplied local car cohorts from synced listings",
		Example: "  cars-il gaps --data-source local --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.formatter().WriteValue(map[string]string{"status": "requires larger synced dataset"})
		},
	}
	root.AddCommand(cmd)
}
