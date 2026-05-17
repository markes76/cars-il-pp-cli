package commands

import "github.com/spf13/cobra"

func addPatterns(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:     "patterns",
		Short:   "Detect repeated market patterns across synced car listings",
		Example: "  cars-il patterns --data-source local --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.formatter().WriteValue(map[string]string{"patterns": "sync more listings for pattern detection"})
		},
	}
	root.AddCommand(cmd)
}
