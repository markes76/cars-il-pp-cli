package commands

import (
	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/spf13/cobra"
)

func addGet(root *cobra.Command, app *App) {
	listingsCmd := &cobra.Command{Use: "listings", Short: "Listing operations"}
	var id string
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a full listing detail view",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return client.InvalidArgs("--id is required")
			}
			listing, err := app.Service.Get(id, app.DataSource)
			if err != nil {
				return err
			}
			return app.formatter().WriteValue(listing)
		},
	}
	getCmd.Flags().StringVar(&id, "id", "", "listing id such as yad2-1234 or autotrader-9012")
	_ = getCmd.MarkFlagRequired("id")
	listingsCmd.AddCommand(getCmd)
	root.AddCommand(listingsCmd)
}
