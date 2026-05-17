package commands

import (
	"github.com/mvanhorn/cars-il-pp-cli/internal/client"
	"github.com/mvanhorn/cars-il-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func addSearch(root *cobra.Command, app *App) {
	var params client.SearchParams
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Yad2 and AutoTrader IL car listings",
		RunE: func(cmd *cobra.Command, args []string) error {
			base := baseParams(app)
			params.Source, params.Limit, params.DataSource = base.Source, base.Limit, base.DataSource
			listings, err := app.Service.Search(params)
			if err != nil {
				return err
			}
			if len(listings) == 0 {
				return client.NotFound("no listings found")
			}
			return app.formatter().WriteListings(listings)
		},
	}
	bindSearchFlags(cmd, &params)
	root.AddCommand(cmd)
}

func bindSearchFlags(cmd *cobra.Command, params *client.SearchParams) {
	cmd.Flags().StringVar(&params.Make, "make", "", "vehicle make, English or Hebrew")
	cmd.Flags().StringVar(&params.Model, "model", "", "vehicle model, English or Hebrew")
	cmd.Flags().StringVar(&params.Query, "query", "", "full-text search over make, model, and Hebrew description")
	cmd.Flags().IntVar(&params.YearMin, "year-min", 0, "minimum year")
	cmd.Flags().IntVar(&params.YearMax, "year-max", 0, "maximum year")
	cmd.Flags().IntVar(&params.PriceMin, "price-min", 0, "minimum price in ILS")
	cmd.Flags().IntVar(&params.PriceMax, "price-max", 0, "maximum price in ILS")
	cmd.Flags().IntVar(&params.MileageMax, "mileage-max", 0, "maximum mileage in km")
	cmd.Flags().StringVar(&params.City, "city", "", "city")
	cmd.Flags().StringVar(&params.Region, "region", "", "region")
	cmd.Flags().StringVar(&params.Fuel, "fuel", "all", "[petrol|diesel|hybrid|electric|all]")
	cmd.Flags().StringVar(&params.Gear, "gear", "all", "[auto|manual|all]")
	cmd.Flags().IntVar(&params.HandMax, "hand-max", 0, "max ownership count")
	cmd.Flags().BoolVar(&params.PrivateOnly, "private-only", false, "exclude dealer listings")
	cmd.Flags().BoolVar(&params.DealerOnly, "dealer-only", false, "exclude private listings")
	cmd.Flags().StringVar(&params.Sort, "sort", "date-new", "[price-asc|price-desc|date-new|date-old|mileage-asc|days-asc]")
}

func searchDomainListings(db *store.DB, params client.SearchParams) ([]client.Listing, error) {
	return db.SearchListings(params)
}
