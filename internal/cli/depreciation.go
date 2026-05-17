package commands

import (
	"sort"

	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/markes76/cars-il-pp-cli/internal/output"
	"github.com/spf13/cobra"
)

func addDepreciation(root *cobra.Command, app *App) {
	var makeName, modelName string
	cmd := &cobra.Command{
		Use:   "depreciation",
		Short: "Compute price-by-year curve for a model",
		RunE: func(cmd *cobra.Command, args []string) error {
			if makeName == "" || modelName == "" {
				return client.InvalidArgs("--make and --model are required")
			}
			listings, err := app.Service.DB.SearchListings(client.SearchParams{Make: makeName, Model: modelName, Limit: 200})
			if err != nil {
				return err
			}
			if len(listings) < 1 {
				return client.NotFound("not enough local listings for depreciation curve")
			}
			type row struct {
				Year       int `json:"year"`
				AvgPrice   int `json:"avg_price"`
				Listings   int `json:"listings"`
				AvgMileage int `json:"avg_mileage"`
			}
			group := map[int][]client.Listing{}
			for _, listing := range listings {
				if listing.Year > 0 {
					group[listing.Year] = append(group[listing.Year], listing)
				}
			}
			var rows []row
			for year, items := range group {
				var priceSum, priceCount, mileageSum, mileageCount int
				for _, item := range items {
					if item.Price > 0 {
						priceSum += item.Price
						priceCount++
					}
					if item.Mileage > 0 {
						mileageSum += item.Mileage
						mileageCount++
					}
				}
				r := row{Year: year, Listings: len(items)}
				if priceCount > 0 {
					r.AvgPrice = priceSum / priceCount
				}
				if mileageCount > 0 {
					r.AvgMileage = mileageSum / mileageCount
				}
				rows = append(rows, r)
			}
			sort.Slice(rows, func(i, j int) bool { return rows[i].Year < rows[j].Year })
			if app.JSON {
				return app.formatter().WriteValue(rows)
			}
			printHuman(app.out, app.Quiet, "%s %s - Market Price by Year\n", makeName, modelName)
			printHuman(app.out, app.Quiet, "Year | Avg Price | Listings | Avg Mileage\n")
			for _, r := range rows {
				printHuman(app.out, app.Quiet, "%d | %s | %d | %d\n", r.Year, output.Shekel(r.AvgPrice), r.Listings, r.AvgMileage)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&makeName, "make", "", "make")
	cmd.Flags().StringVar(&modelName, "model", "", "model")
	root.AddCommand(cmd)
}
