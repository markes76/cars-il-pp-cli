package commands

import (
	"strconv"
	"strings"

	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/markes76/cars-il-pp-cli/internal/output"
	"github.com/spf13/cobra"
)

func addCompare(root *cobra.Command, app *App) {
	var ids string
	cmd := &cobra.Command{
		Use:     "compare",
		Short:   "Compare up to five listings side by side",
		Example: "  cars-il compare --ids yad2-1234,yad2-5678,autotrader-9012 --data-source local",
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := splitCSV(ids)
			if len(parts) == 0 {
				return client.InvalidArgs("--ids is required")
			}
			if len(parts) > 5 {
				return client.InvalidArgs("compare supports up to 5 listings")
			}
			var listings []client.Listing
			for _, id := range parts {
				listing, err := app.Service.Get(id, app.DataSource)
				if err != nil {
					return err
				}
				listings = append(listings, listing)
			}
			if app.JSON || app.CSV {
				return app.formatter().WriteListings(listings)
			}
			fields := []string{"Make/Model", "Year", "Mileage", "Price", "City", "Hand (יד)", "Seller type", "Fuel", "Gear", "Test expiry", "Days listed"}
			rows := [][]string{{"Field"}}
			for _, listing := range listings {
				rows[0] = append(rows[0], listing.ID)
			}
			for _, field := range fields {
				row := []string{field}
				for _, l := range listings {
					row = append(row, compareValue(field, l))
				}
				rows = append(rows, row)
			}
			for _, row := range rows {
				printHuman(app.out, app.Quiet, "%s\n", strings.Join(row, " | "))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&ids, "ids", "", "comma-separated listing ids to compare side by side")
	_ = cmd.MarkFlagRequired("ids")
	root.AddCommand(cmd)
}

func compareValue(field string, l client.Listing) string {
	switch field {
	case "Make/Model":
		return strings.TrimSpace(l.Make + " " + l.Model)
	case "Year":
		return intText(l.Year)
	case "Mileage":
		return intText(l.Mileage) + " km"
	case "Price":
		return output.Shekel(l.Price)
	case "City":
		return l.City
	case "Hand (יד)":
		return intText(l.Hand)
	case "Seller type":
		if l.IsDealer {
			return "Dealer"
		}
		return "Private"
	case "Fuel":
		return l.FuelType
	case "Gear":
		return l.GearType
	case "Test expiry":
		if len(l.TestExpiry) >= 7 {
			return l.TestExpiry[:7]
		}
		return l.TestExpiry
	case "Days listed":
		return intText(l.DaysOnMarket)
	default:
		return ""
	}
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func intText(value int) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}
