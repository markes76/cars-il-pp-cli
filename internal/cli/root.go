package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mvanhorn/cars-il-pp-cli/internal/client"
	"github.com/mvanhorn/cars-il-pp-cli/internal/cliutil"
	"github.com/mvanhorn/cars-il-pp-cli/internal/output"
	"github.com/mvanhorn/cars-il-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

type App struct {
	DBPath     string
	ExitCode   int
	JSON       bool
	CSV        bool
	Compact    bool
	Plain      bool
	Select     string
	Quiet      bool
	DryRun     bool
	Yes        bool
	NoInput    bool
	NoColor    bool
	DataSource string
	Source     string
	Limit      int
	OutputFile string
	Currency   string
	Romanize   bool
	Agent      bool
	Service    Service
	out        io.Writer
}

func NewRootCommand() (*cobra.Command, *App) {
	app := &App{ExitCode: client.ExitSuccess, DataSource: "auto", Source: client.SourceAll, Limit: 20, Currency: "ILS"}
	root := &cobra.Command{
		Use:           "cars-il",
		Short:         "Israeli used-car market intelligence for Yad2 and AutoTrader IL",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cliutil.EnsureFresh() && autoRefreshIfStale() && !app.Quiet {
				_, _ = fmt.Fprintln(os.Stderr, "cache refresh recommended")
			}
			db, err := store.Open(app.DBPath)
			if err != nil {
				return err
			}
			app.Service = NewService(db)
			if app.OutputFile != "" {
				f, err := os.Create(app.OutputFile)
				if err != nil {
					return err
				}
				app.out = f
			} else {
				app.out = os.Stdout
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if app.Service.DB != nil {
				_ = app.Service.DB.Close()
			}
			if closer, ok := app.out.(io.Closer); ok && closer != os.Stdout {
				_ = closer.Close()
			}
		},
	}
	root.PersistentFlags().BoolVar(&app.JSON, "json", false, "force JSON output")
	root.PersistentFlags().BoolVar(&app.CSV, "csv", false, "CSV output")
	root.PersistentFlags().BoolVar(&app.Compact, "compact", false, "high-gravity fields only")
	root.PersistentFlags().BoolVar(&app.Plain, "plain", false, "plain table output")
	root.PersistentFlags().StringVar(&app.Select, "select", "", "comma-separated field filter")
	root.PersistentFlags().BoolVar(&app.Quiet, "quiet", false, "suppress non-data output")
	root.PersistentFlags().BoolVar(&app.DryRun, "dry-run", false, "print what would happen, no changes")
	root.PersistentFlags().BoolVar(&app.Yes, "yes", false, "skip confirmation prompts")
	root.PersistentFlags().BoolVar(&app.NoInput, "no-input", false, "never prompt for input")
	root.PersistentFlags().BoolVar(&app.NoColor, "no-color", false, "disable ANSI")
	root.PersistentFlags().StringVar(&app.DataSource, "data-source", "auto", "[auto|local|live]")
	root.PersistentFlags().StringVar(&app.Source, "source", client.SourceAll, "[yad2|autotrader|all]")
	root.PersistentFlags().IntVar(&app.Limit, "limit", 20, "max results")
	root.PersistentFlags().StringVar(&app.OutputFile, "output-file", "", "write output to file")
	root.PersistentFlags().StringVar(&app.DBPath, "db", "", "SQLite store path")
	root.PersistentFlags().StringVar(&app.Currency, "currency", "ILS", "display currency [ILS|USD]")
	root.PersistentFlags().BoolVar(&app.Romanize, "romanize", false, "transliterate Hebrew city/region names when available")
	root.PersistentFlags().BoolVar(&app.Agent, "agent", false, "optimize output for agent workflows")

	addSearch(root, app)
	addGet(root, app)
	addSync(root, app)
	addWatch(root, app)
	addCompare(root, app)
	addDeal(root, app)
	addMarket(root, app)
	addDepreciation(root, app)
	addStale(root, app)
	addDoctor(root, app)
	addPriceHistory(root, app)
	addProfile(root, app)
	addDeliver(root, app)
	addFeedback(root, app)
	addJobs(root, app)
	addTrends(root, app)
	addGaps(root, app)
	addForecast(root, app)
	addPatterns(root, app)
	addAgentContext(root)
	return root, app
}

func (app *App) formatter() output.Formatter {
	format := output.AutoFormat(app.JSON, app.CSV)
	fields := []string(nil)
	if app.Select != "" {
		for _, field := range strings.Split(app.Select, ",") {
			if trimmed := strings.TrimSpace(field); trimmed != "" {
				fields = append(fields, trimmed)
			}
		}
	}
	return output.Formatter{Format: format, Compact: app.Compact, Select: fields, Quiet: app.Quiet, Writer: app.out}
}

func baseParams(app *App) client.SearchParams {
	limit := app.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	return client.SearchParams{Source: app.Source, Limit: limit, DataSource: app.DataSource}
}

func Execute() int {
	root, app := NewRootCommand()
	if err := root.Execute(); err != nil {
		WriteError(err)
		return ExitCode(err)
	}
	return app.ExitCode
}

func WriteError(err error) {
	var appErr client.AppError
	if errors.As(err, &appErr) {
		_ = json.NewEncoder(os.Stderr).Encode(map[string]interface{}{"error": map[string]string{"code": appErr.Code, "message": appErr.Message}})
		return
	}
	_ = json.NewEncoder(os.Stderr).Encode(map[string]interface{}{"error": map[string]string{"code": "ERROR", "message": err.Error()}})
}

func ExitCode(err error) int {
	var appErr client.AppError
	if errors.As(err, &appErr) {
		return appErr.ExitCode
	}
	return client.ExitAPIError
}

func printHuman(w io.Writer, quiet bool, format string, args ...interface{}) {
	if quiet {
		return
	}
	_, _ = fmt.Fprintf(w, format, args...)
}
