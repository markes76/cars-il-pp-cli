# cars-il-pp-cli

Read-only Israeli used-car market intelligence CLI and MCP server for Yad2 Cars, with a guarded AutoTrader IL adapter.

`cars-il-pp-cli` is built around one idea: used-car listings are market signals. Prices, mileage, ownership count, listing age, and Hebrew seller text can show buyer demand and seller urgency before official market data catches up.

> Current status: Printing Press scorecard Grade A, 88/100. Yad2 live search and local sync were smoke-tested successfully on 2026-05-17. AutoTrader IL is reachable, but the current public `autotrader.co.il` site is a WordPress import/services site, not a public classifieds catalogue, so the adapter returns `SOURCE_UNAVAILABLE` instead of fabricating listings.

## עברית

`cars-il-pp-cli` הוא כלי CLI ושרת MCP לקריאת מודעות רכב משומש בישראל בצורה בטוחה וקריאה בלבד. הכלי עובד היום מול יד2, שומר נתונים בעברית ב-SQLite, ומחשב תובנות כמו מחיר שוק, ציון עסקה, מודעות ישנות, היסטוריית מחיר וחום שוק.

סטטוס נוכחי: הכלי קיבל Grade A בציון Printing Press עם 88/100. חיפוש חי ביד2 וסנכרון מקומי נבדקו בהצלחה ב-17 במאי 2026. אתר AutoTrader IL הנוכחי אינו מציג קטלוג מודעות ציבורי אלא אתר שירותי יבוא/טרייד, ולכן הכלי מחזיר `SOURCE_UNAVAILABLE` עבור המקור הזה.

## What You Can Do

- Search live Yad2 listings from the terminal.
- Store Hebrew listings locally in SQLite with UTF-8 and FTS5 `unicode61`.
- Search with English or Hebrew make/model names, for example `Toyota` or `טויוטה`.
- Sync listings into a local database.
- Compare listings side by side.
- Score a listing against local market data.
- Inspect stale listings that may indicate motivated sellers.
- Calculate market summaries and market heat.
- Use the MCP server from agent tools that speak JSON-RPC over stdio.

## Install

From source:

```bash
git clone https://github.com/markes76/cars-il-pp-cli.git
cd cars-il-pp-cli
go build -o cars-il-pp-cli ./cmd/cars-il-pp-cli
go build -o cars-il-pp-mcp ./cmd/cars-il-mcp
```

Local build paths from this Printing Press run:

```bash
/Users/mark.s/printing-press/library/cars-il/cars-il-pp-cli
/Users/mark.s/printing-press/library/cars-il/cars-il-pp-mcp
```

## Quick Start

```bash
./cars-il-pp-cli doctor --json
./cars-il-pp-cli search --make Toyota --model Corolla --limit 5 --source yad2 --data-source live --json
./cars-il-pp-cli search --make "טויוטה" --model "קורולה" --city "תל אביב" --compact
./cars-il-pp-cli sync --make Toyota --model Corolla --limit 25 --source yad2
./cars-il-pp-cli market --make Toyota --model Corolla --data-source local --json
```

Hebrew values with spaces should be quoted:

```bash
./cars-il-pp-cli search --city "תל אביב"  # correct
./cars-il-pp-cli search --city תל אביב    # usually parsed as two arguments
```

## התחלה מהירה בעברית

```bash
./cars-il-pp-cli doctor --json
./cars-il-pp-cli search --make "טויוטה" --model "קורולה" --limit 5 --source yad2 --data-source live --json
./cars-il-pp-cli sync --make "טויוטה" --model "קורולה" --limit 25 --source yad2
./cars-il-pp-cli market --make "טויוטה" --model "קורולה" --data-source local --json
```

## Common Workflows

Find a hybrid around ₪50,000 with lower mileage:

```bash
./cars-il-pp-cli search \
  --fuel hybrid \
  --price-max 60000 \
  --mileage-max 140000 \
  --source yad2 \
  --data-source live \
  --limit 20 \
  --json
```

Find a 2020 Toyota Corolla under ₪95,000 in Tel Aviv:

```bash
./cars-il-pp-cli search \
  --make Toyota \
  --model Corolla \
  --year-min 2020 \
  --year-max 2020 \
  --price-max 95000 \
  --city "תל אביב" \
  --sort price-asc \
  --compact
```

Sync data before analytics:

```bash
./cars-il-pp-cli sync --make Toyota --model Corolla --limit 50 --source yad2
./cars-il-pp-cli market --make Toyota --model Corolla --json
```

Score a listing:

```bash
./cars-il-pp-cli deal --id yad2-1234 --json
```

Compare listings:

```bash
./cars-il-pp-cli compare --ids yad2-1234,yad2-5678 --data-source local
```

Show stale listings:

```bash
./cars-il-pp-cli stale --days 30 --make Toyota --model Corolla --data-source local
```

## תהליכי עבודה נפוצים

חיפוש רכב היברידי סביב 50,000 ש"ח עם קילומטראז' נמוך:

```bash
./cars-il-pp-cli search \
  --fuel hybrid \
  --price-max 60000 \
  --mileage-max 140000 \
  --source yad2 \
  --data-source live \
  --limit 20 \
  --json
```

בדיקת מחיר שוק אחרי סנכרון:

```bash
./cars-il-pp-cli sync --make "טויוטה" --model "קורולה" --limit 50 --source yad2
./cars-il-pp-cli market --make "טויוטה" --model "קורולה" --json
```

בדיקת ציון עסקה:

```bash
./cars-il-pp-cli deal --id yad2-1234 --json
```

השוואת מודעות:

```bash
./cars-il-pp-cli compare --ids yad2-1234,yad2-5678 --data-source local
```

## Commands

```bash
cars-il-pp-cli search
cars-il-pp-cli listings get --id <listing-id>
cars-il-pp-cli sync
cars-il-pp-cli watch
cars-il-pp-cli compare --ids <id1,id2,id3>
cars-il-pp-cli deal --id <listing-id>
cars-il-pp-cli price-history --id <listing-id>
cars-il-pp-cli stale --days 30
cars-il-pp-cli market
cars-il-pp-cli market-heat
cars-il-pp-cli depreciation --make <make> --model <model>
cars-il-pp-cli doctor
cars-il-pp-cli agent-context
```

Global flags:

```text
--json
--csv
--compact
--select FIELDS
--quiet
--dry-run
--yes
--no-input
--no-color
--data-source auto|local|live
--source yad2|autotrader|all
--limit N
--output-file PATH
--db PATH
--currency ILS|USD
--romanize
--agent
```

When stdout is not a terminal, output defaults to JSON. `--json` and `--csv` override that behavior.

## Authentication And Cookies

Most Yad2 live searches currently work without cookies because the client reads public HTML payloads with browser-like headers. If Yad2 returns `AUTH_FAILURE`, use a fresh browser session cookie:

```bash
export CARS_IL_YAD2_COOKIE='paste-cookie-here'
./cars-il-pp-cli search --make Toyota --limit 5 --source yad2 --data-source live --json
```

Do not commit cookies. Do not paste cookies into issues, pull requests, or chat logs. Cookies are session credentials.

AutoTrader cookie support exists as `CARS_IL_AUTOTRADER_COOKIE`, but the current public site does not expose a classifieds listing surface.

## Database

Default database path is under the user cache directory. For testing, pass `--db`:

```bash
DB=/tmp/cars-il-test.db
./cars-il-pp-cli --db "$DB" sync --make Toyota --model Corolla --limit 10 --source yad2
./cars-il-pp-cli --db "$DB" search --make Toyota --data-source local --json
```

SQLite uses:

```sql
PRAGMA encoding = 'UTF-8';
CREATE VIRTUAL TABLE listings_fts USING fts5(
  id UNINDEXED,
  make,
  model,
  description,
  tokenize='unicode61'
);
```

Hebrew text is stored as original UTF-8, not transliterated.

## MCP Server

Build:

```bash
go build -o cars-il-pp-mcp ./cmd/cars-il-mcp
```

The MCP server exposes the same operation names as the CLI over JSON-RPC stdio, including:

```text
search
listings_get
sync
market
deal
price_history
stale
compare
depreciation
market_heat
doctor
```

The CLI remains the better interface for long-running watch workflows.

## Exit Codes

```text
0 success
2 invalid arguments
3 not found or source unavailable
4 auth failure
5 upstream API error
7 rate limited
9 dry-run completed
```

Errors are JSON on stderr:

```json
{"error":{"code":"SOURCE_UNAVAILABLE","message":"AutoTrader IL currently exposes WordPress service pages, not a public used-car listing catalogue"}}
```

## Safety And Ethics

- GET-only against target sites.
- No account scraping.
- No phone-number extraction.
- No hardcoded secrets.
- 10-second request timeout.
- Conservative request cadence and backoff on rate limits.
- Users are responsible for complying with each site's terms of service.

## Current Limitations

- AutoTrader IL currently does not expose a public used-car listing catalogue at `autotrader.co.il`.
- Yad2 has no public API; the adapter depends on public browser-rendered HTML payloads and may need updates if Yad2 changes its Next.js data shape.
- Some live filters are applied client-side after fetching public search pages. For serious analysis, sync a wider cohort and run local filters.
- Price history becomes useful only after repeated sync runs.

## Detailed Guides

- [English how-to guide](docs/HOW_TO_USE.en.md)
- [מדריך שימוש בעברית](docs/HOW_TO_USE.he.md)

