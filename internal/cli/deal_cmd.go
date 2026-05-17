package commands

import (
	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/markes76/cars-il-pp-cli/internal/output"
	"github.com/spf13/cobra"
)

func addDeal(root *cobra.Command, app *App) {
	var id string
	cmd := &cobra.Command{
		Use:     "deal",
		Short:   "Score a listing against its local market cohort",
		Example: "  cars-il deal --id yad2-1234 --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return client.InvalidArgs("--id is required")
			}
			result, err := app.Service.Deal(id)
			if err != nil {
				return err
			}
			if app.JSON {
				return app.formatter().WriteValue(result)
			}
			l := result.Listing
			printHuman(app.out, app.Quiet, "Listing: %d %s %s, %d km, %s (%s)\n", l.Year, l.Make, l.Model, l.Mileage, output.Shekel(l.Price), l.City)
			printHuman(app.out, app.Quiet, "Market median:          %s\n", output.Shekel(result.MarketMedian))
			printHuman(app.out, app.Quiet, "Price vs median:        %+.1f%%\n", result.PriceVsMedianPct)
			printHuman(app.out, app.Quiet, "Mileage vs avg:         %+.1f%%\n", result.MileageVsAvgPct)
			printHuman(app.out, app.Quiet, "Deal Score:             %d / 100\n", result.Score)
			printHuman(app.out, app.Quiet, "Verdict:                %s\n", result.Verdict)
			printHuman(app.out, app.Quiet, "Negotiation range:      %s-%s\n", output.Shekel(result.NegotiationLow), output.Shekel(result.NegotiationHigh))
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "listing id to score against the synced local market cohort")
	_ = cmd.MarkFlagRequired("id")
	root.AddCommand(cmd)
}
