package commands

import (
	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/markes76/cars-il-pp-cli/internal/output"
	"github.com/spf13/cobra"
)

func addPriceHistory(root *cobra.Command, app *App) {
	var id string
	cmd := &cobra.Command{
		Use:   "price-history",
		Short: "Show price change history for a listing",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return client.InvalidArgs("--id is required")
			}
			history, err := app.Service.DB.GetPriceHistory(id)
			if err != nil {
				return err
			}
			if len(history) == 0 {
				return client.NotFound("no price history found for " + id)
			}
			if app.JSON {
				return app.formatter().WriteValue(history)
			}
			printHuman(app.out, app.Quiet, "Price History: %s\n", id)
			for i, point := range history {
				change := ""
				if i > 0 {
					prev := history[i-1].Price
					if prev > 0 {
						diff := point.Price - prev
						change = " "
						if diff > 0 {
							change += "+"
						}
						change += output.Shekel(diff)
					}
				}
				printHuman(app.out, app.Quiet, "%s   %s%s\n", point.RecordedAt[:10], output.Shekel(point.Price), change)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "listing id with at least one synced price observation")
	_ = cmd.MarkFlagRequired("id")
	root.AddCommand(cmd)
}
