package commands

import (
	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/spf13/cobra"
)

func addProfile(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Show the active local buyer profile for agent workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.formatter().WriteValue(map[string]interface{}{"profile": "default", "currency": "ILS", "source": app.Source, "limit": app.Limit})
		},
	}
	cmd.Flags().String("name", "default", "saved profile name")
	root.AddCommand(cmd)
	_ = client.SourceAll
}
