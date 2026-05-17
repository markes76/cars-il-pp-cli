package commands

import "github.com/spf13/cobra"

func addJobs(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List background search and watch jobs for agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.formatter().WriteValue([]map[string]string{})
		},
	}
	cmd.Flags().Bool("wait", false, "wait for a background job to finish")
	root.AddCommand(cmd)
}
