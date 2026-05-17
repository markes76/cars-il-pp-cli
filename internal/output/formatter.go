package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/markes76/cars-il-pp-cli/internal/client"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

type Formatter struct {
	Format  string
	Compact bool
	Select  []string
	Quiet   bool
	Writer  io.Writer
}

func AutoFormat(forceJSON, forceCSV bool) string {
	outputFormat := "table"
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		outputFormat = "json"
	}
	if forceJSON {
		outputFormat = "json"
	}
	if forceCSV {
		outputFormat = "csv"
	}
	return outputFormat
}

func (f Formatter) writer() io.Writer {
	if f.Writer != nil {
		return f.Writer
	}
	return os.Stdout
}

func (f Formatter) WriteListings(listings []client.Listing) error {
	switch f.Format {
	case "json":
		enc := json.NewEncoder(f.writer())
		enc.SetIndent("", "  ")
		return enc.Encode(listings)
	case "csv":
		return f.writeCSV(listings)
	default:
		return f.writeTable(listings)
	}
}

func (f Formatter) WriteValue(value interface{}) error {
	if f.Format == "json" {
		enc := json.NewEncoder(f.writer())
		enc.SetIndent("", "  ")
		return enc.Encode(value)
	}
	_, err := fmt.Fprintln(f.writer(), value)
	return err
}

func (f Formatter) writeCSV(listings []client.Listing) error {
	w := csv.NewWriter(f.writer())
	headers := f.fields()
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, listing := range listings {
		row := make([]string, 0, len(headers))
		for _, field := range headers {
			row = append(row, fieldValue(listing, field))
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func (f Formatter) writeTable(listings []client.Listing) error {
	fields := f.fields()
	rows := make([][]string, 0, len(listings)+1)
	rows = append(rows, fields)
	for _, listing := range listings {
		row := make([]string, 0, len(fields))
		for _, field := range fields {
			row = append(row, fieldValue(listing, field))
		}
		rows = append(rows, row)
	}
	widths := make([]int, len(fields))
	for _, row := range rows {
		for i, cell := range row {
			if w := runewidth.StringWidth(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}
	for idx, row := range rows {
		for i, cell := range row {
			if i > 0 {
				if _, err := fmt.Fprint(f.writer(), " | "); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(f.writer(), pad(cell, widths[i])); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(f.writer()); err != nil {
			return err
		}
		if idx == 0 {
			total := 0
			for _, w := range widths {
				total += w
			}
			total += (len(widths) - 1) * 3
			if _, err := fmt.Fprintln(f.writer(), strings.Repeat("-", total)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f Formatter) fields() []string {
	if len(f.Select) > 0 {
		return f.Select
	}
	if f.Compact {
		return []string{"ID", "Make", "Model", "Year", "Price", "City"}
	}
	return []string{"ID", "Make", "Model", "Year", "Mileage", "Price", "City", "Hand", "Days"}
}

func fieldValue(listing client.Listing, field string) string {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "id":
		return listing.ID
	case "source":
		return listing.Source
	case "make":
		return listing.Make
	case "model":
		return listing.Model
	case "year":
		return intString(listing.Year)
	case "mileage":
		if listing.Mileage == 0 {
			return ""
		}
		return comma(listing.Mileage)
	case "price":
		return Shekel(listing.Price)
	case "city":
		return listing.City
	case "region":
		return listing.Region
	case "fuel", "fuel_type":
		return listing.FuelType
	case "gear", "gear_type":
		return listing.GearType
	case "hand":
		return intString(listing.Hand)
	case "days", "days_on_market":
		return intString(listing.DaysOnMarket)
	case "url":
		return listing.URL
	default:
		return ""
	}
}

func Shekel(value int) string {
	if value == 0 {
		return "לא צוין מחיר"
	}
	return "₪" + comma(value)
}

func comma(value int) string {
	s := strconv.Itoa(value)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	pre := len(s) % 3
	if pre == 0 {
		pre = 3
	}
	out = append(out, s[:pre]...)
	for i := pre; i < len(s); i += 3 {
		out = append(out, ',')
		out = append(out, s[i:i+3]...)
	}
	return string(out)
}

func intString(value int) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func pad(value string, width int) string {
	diff := width - runewidth.StringWidth(value)
	if diff <= 0 {
		return value
	}
	return value + strings.Repeat(" ", diff)
}
