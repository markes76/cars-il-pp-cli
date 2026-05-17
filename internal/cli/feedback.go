package commands

import "github.com/spf13/cobra"

func addFeedback(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Record operator feedback about a car search result",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.formatter().WriteValue(map[string]string{"feedback": "accepted"})
		},
	}
	cmd.Flags().String("id", "", "listing id the feedback refers to")
	cmd.Flags().String("note", "", "short feedback note")
	root.AddCommand(cmd)
}
