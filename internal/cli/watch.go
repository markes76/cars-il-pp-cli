package commands

import (
	"bytes"
	"net/http"
	"time"

	"github.com/mvanhorn/cars-il-pp-cli/internal/client"
	"github.com/spf13/cobra"
)

func addWatch(root *cobra.Command, app *App) {
	var params client.SearchParams
	var interval int
	var alert, webhookURL string
	var priceDropOnly bool
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Poll for new listings or price drops",
		RunE: func(cmd *cobra.Command, args []string) error {
			if interval < 15 {
				return client.InvalidArgs("--interval must be at least 15 minutes")
			}
			base := baseParams(app)
			params.Source, params.Limit, params.DataSource = base.Source, base.Limit, "live"
			listings, err := app.Service.Search(params)
			if err != nil {
				return err
			}
			for _, listing := range listings {
				if priceDropOnly {
					continue
				}
				line := "NEW LISTING: " + listing.ID + "\n" + listing.URL + "\n"
				if alert == "webhook" && webhookURL != "" {
					req, _ := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBufferString(line))
					req.Header.Set("Content-Type", "text/plain; charset=utf-8")
					resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
					if err != nil {
						return client.APIError(err.Error())
					}
					if resp != nil {
						_ = resp.Body.Close()
						if resp.StatusCode == http.StatusTooManyRequests {
							return client.RateLimited("webhook endpoint returned 429")
						}
						if resp.StatusCode >= 400 {
							return client.APIError("webhook endpoint returned " + resp.Status)
						}
					}
				} else {
					printHuman(app.out, app.Quiet, "%s", line)
				}
			}
			if app.DryRun {
				app.ExitCode = client.ExitDryRun
				return nil
			}
			_ = time.Duration(interval) * time.Minute
			return nil
		},
	}
	bindSearchFlags(cmd, &params)
	cmd.Flags().IntVar(&interval, "interval", 60, "poll interval in minutes")
	cmd.Flags().StringVar(&alert, "alert", "stdout", "[stdout|slack|webhook]")
	cmd.Flags().StringVar(&webhookURL, "webhook-url", "", "webhook URL")
	cmd.Flags().BoolVar(&priceDropOnly, "price-drop-only", false, "only alert on price drops")
	root.AddCommand(cmd)
}
