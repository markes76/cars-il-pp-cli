# How To Use cars-il-pp-cli

This guide walks through the normal flow for using the CLI from a fresh terminal.

## 1. Build The Binaries

```bash
cd cars-il-pp-cli
go build -o cars-il-pp-cli ./cmd/cars-il-pp-cli
go build -o cars-il-pp-mcp ./cmd/cars-il-mcp
```

Check the binary:

```bash
./cars-il-pp-cli --help
./cars-il-pp-cli doctor --json
```

## 2. Run A Live Yad2 Search

Start without cookies:

```bash
./cars-il-pp-cli search \
  --make Toyota \
  --model Corolla \
  --limit 5 \
  --source yad2 \
  --data-source live \
  --json
```

Hebrew input works:

```bash
./cars-il-pp-cli search \
  --make "טויוטה" \
  --model "קורולה" \
  --limit 5 \
  --source yad2 \
  --data-source live \
  --json
```

## 3. Optional Cookie Setup

Only do this if Yad2 returns `AUTH_FAILURE`.

1. Open Chrome.
2. Go to `https://www.yad2.co.il/vehicles/cars`.
3. Let the page fully load.
4. Open DevTools: `Option + Command + I` on macOS.
5. Open the Network tab.
6. Click the main document request for `/vehicles/cars`.
7. Open Headers.
8. Under Request Headers, copy only the value after `Cookie:`.
9. In Terminal, run:

```bash
export CARS_IL_YAD2_COOKIE='paste-cookie-value-here'
```

Do not paste cookies into GitHub issues, README examples, or AI chats. Cookies are private session credentials.

## 4. Search For A Real Buying Target

Example: hybrid, around ₪50,000, lower mileage:

```bash
./cars-il-pp-cli search \
  --fuel hybrid \
  --price-max 60000 \
  --mileage-max 140000 \
  --source yad2 \
  --data-source live \
  --limit 20 \
  --sort mileage-asc \
  --json
```

For better results, search by likely models:

```bash
./cars-il-pp-cli search --make Toyota --model Yaris --fuel hybrid --price-max 65000 --source yad2 --data-source live --json
./cars-il-pp-cli search --make Toyota --model Prius --fuel hybrid --price-max 65000 --source yad2 --data-source live --json
./cars-il-pp-cli search --make Hyundai --model Ioniq --fuel hybrid --price-max 70000 --source yad2 --data-source live --json
```

## 5. Sync Before Analytics

Analytics commands use local SQLite data.

```bash
DB=/tmp/cars-il.db

./cars-il-pp-cli --db "$DB" sync \
  --make Toyota \
  --model Corolla \
  --limit 50 \
  --source yad2
```

Then query locally:

```bash
./cars-il-pp-cli --db "$DB" search --make Toyota --model Corolla --data-source local --compact
./cars-il-pp-cli --db "$DB" market --make Toyota --model Corolla --json
./cars-il-pp-cli --db "$DB" market-heat --json
```

## 6. Score A Listing

After syncing, choose an ID and score it:

```bash
./cars-il-pp-cli --db "$DB" deal --id yad2-1234 --json
```

The score combines:

- price versus cohort median
- mileage versus cohort average
- days on market
- ownership count

## 7. Compare Listings

```bash
./cars-il-pp-cli --db "$DB" compare --ids yad2-1234,yad2-5678 --data-source local
```

## 8. AutoTrader IL Behavior

As of 2026-05-17, `https://autotrader.co.il/` is a WordPress import/services site. It does not expose a used-car listings catalogue. The CLI therefore returns:

```json
{"error":{"code":"SOURCE_UNAVAILABLE","message":"AutoTrader IL currently exposes WordPress service pages, not a public used-car listing catalogue"}}
```

That is expected behavior, not a crash.

## 9. Troubleshooting

Store is empty:

```bash
./cars-il-pp-cli sync --make Toyota --model Corolla --limit 25 --source yad2
```

Yad2 auth failure:

```bash
export CARS_IL_YAD2_COOKIE='fresh-cookie-from-browser'
```

Need machine-readable output:

```bash
./cars-il-pp-cli search --make Toyota --limit 5 --json
```

Need a disposable test DB:

```bash
DB=/tmp/cars-il-test.db
rm -f "$DB" "$DB-shm" "$DB-wal"
./cars-il-pp-cli --db "$DB" doctor --json
```

## 10. Use It As A Claude MCP Server

Build or install:

```bash
go install github.com/markes76/cars-il-pp-cli/cmd/cars-il-mcp@latest
which cars-il-mcp
```

Claude Desktop on macOS reads local MCP subprocess configuration from:

```bash
$HOME/Library/Application Support/Claude/claude_desktop_config.json
```

Add:

```json
{
  "mcpServers": {
    "cars-il": {
      "type": "stdio",
      "command": "/Users/YOUR_USER/go/bin/cars-il-mcp",
      "args": [],
      "env": {
        "CARS_IL_YAD2_COOKIE": "",
        "CARS_IL_AUTOTRADER_COOKIE": ""
      }
    }
  }
}
```

Restart Claude Desktop. Then ask:

```text
Use the cars-il MCP. Call context and doctor, then search Yad2 live for 5 Toyota Corolla listings.
```

For Claude Code:

```bash
claude mcp add-json cars-il '{
  "type": "stdio",
  "command": "/Users/YOUR_USER/go/bin/cars-il-mcp",
  "args": [],
  "env": {
    "CARS_IL_YAD2_COOKIE": "",
    "CARS_IL_AUTOTRADER_COOKIE": ""
  }
}'
```

Safety: the MCP server is read-only against remote sites. `sync` writes only to the local SQLite cache.
