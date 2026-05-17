package commands

import "github.com/spf13/cobra"

func addDeliver(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:     "deliver",
		Short:   "Format the latest result for downstream agent delivery",
		Example: "  cars-il deliver --to stdout --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.formatter().WriteValue(map[string]string{"delivery": "stdout", "status": "ready"})
		},
	}
	cmd.Flags().String("to", "stdout", "delivery target")
	root.AddCommand(cmd)
}
