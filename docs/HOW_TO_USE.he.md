# מדריך שימוש ב-cars-il-pp-cli

המדריך הזה מסביר איך להשתמש בכלי מהתחלה: בנייה, חיפוש חי ביד2, סנכרון למסד נתונים מקומי, וניתוח עסקאות.

## 1. בניית הקבצים

```bash
cd cars-il-pp-cli
go build -o cars-il-pp-cli ./cmd/cars-il-pp-cli
go build -o cars-il-pp-mcp ./cmd/cars-il-mcp
```

בדיקה:

```bash
./cars-il-pp-cli --help
./cars-il-pp-cli doctor --json
```

## 2. חיפוש חי ביד2

אפשר להתחיל בלי cookies:

```bash
./cars-il-pp-cli search \
  --make Toyota \
  --model Corolla \
  --limit 5 \
  --source yad2 \
  --data-source live \
  --json
```

אפשר גם בעברית:

```bash
./cars-il-pp-cli search \
  --make "טויוטה" \
  --model "קורולה" \
  --limit 5 \
  --source yad2 \
  --data-source live \
  --json
```

חשוב לשים מירכאות סביב ערכים עם רווחים:

```bash
./cars-il-pp-cli search --city "תל אביב"
```

## 3. שימוש ב-cookie במקרה הצורך

ברוב הבדיקות האחרונות יד2 עבד בלי cookie. אם מתקבלת שגיאת `AUTH_FAILURE`, אפשר להעביר cookie מהדפדפן:

1. פתחו Chrome.
2. היכנסו ל-`https://www.yad2.co.il/vehicles/cars`.
3. חכו שהעמוד ייטען.
4. פתחו DevTools עם `Option + Command + I` במק.
5. עברו ללשונית Network.
6. לחצו על בקשת המסמך הראשית של `/vehicles/cars`.
7. פתחו Headers.
8. תחת Request Headers העתיקו רק את הערך שאחרי `Cookie:`.
9. בטרמינל:

```bash
export CARS_IL_YAD2_COOKIE='כאן-מדביקים-את-ה-cookie'
```

לא להדביק cookies ב-GitHub, ב-README, בצילומי מסך, או בשיחות עם כלי AI. cookie הוא פרטי.

## 4. חיפוש לרכב אמיתי

דוגמה: היברידי סביב 50,000 ש"ח עם קילומטראז' נמוך:

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

כדאי לחפש גם לפי דגמים רלוונטיים:

```bash
./cars-il-pp-cli search --make Toyota --model Yaris --fuel hybrid --price-max 65000 --source yad2 --data-source live --json
./cars-il-pp-cli search --make Toyota --model Prius --fuel hybrid --price-max 65000 --source yad2 --data-source live --json
./cars-il-pp-cli search --make Hyundai --model Ioniq --fuel hybrid --price-max 70000 --source yad2 --data-source live --json
```

## 5. סנכרון לפני ניתוח שוק

פקודות ניתוח כמו `market`, `deal`, `stale`, ו-`market-heat` עובדות הכי טוב אחרי סנכרון למסד SQLite מקומי.

```bash
DB=/tmp/cars-il.db

./cars-il-pp-cli --db "$DB" sync \
  --make Toyota \
  --model Corolla \
  --limit 50 \
  --source yad2
```

לאחר מכן:

```bash
./cars-il-pp-cli --db "$DB" search --make Toyota --model Corolla --data-source local --compact
./cars-il-pp-cli --db "$DB" market --make Toyota --model Corolla --json
./cars-il-pp-cli --db "$DB" market-heat --json
```

## 6. בדיקת ציון עסקה

אחרי סנכרון, בוחרים מזהה מודעה:

```bash
./cars-il-pp-cli --db "$DB" deal --id yad2-1234 --json
```

הציון משלב:

- מחיר ביחס לחציון השוק
- קילומטראז' ביחס לממוצע
- מספר ימים שהמודעה באוויר
- מספר בעלים

## 7. השוואת מודעות

```bash
./cars-il-pp-cli --db "$DB" compare --ids yad2-1234,yad2-5678 --data-source local
```

## 8. תקלות נפוצות

מסד הנתונים ריק:

```bash
./cars-il-pp-cli sync --make Toyota --model Corolla --limit 25 --source yad2
```

שגיאת הרשאה ביד2:

```bash
export CARS_IL_YAD2_COOKIE='cookie-חדש-מהדפדפן'
```

פלט JSON לסוכן או סקריפט:

```bash
./cars-il-pp-cli search --make Toyota --limit 5 --json
```

מסד בדיקה זמני:

```bash
DB=/tmp/cars-il-test.db
rm -f "$DB" "$DB-shm" "$DB-wal"
./cars-il-pp-cli --db "$DB" doctor --json
```

## 9. שימוש כשרת MCP עבור Claude

בנייה, התקנה, או שימוש בבינארי שכבר נבנה.

אם Go מותקן:

```bash
go install github.com/markes76/cars-il-pp-cli/cmd/cars-il-mcp@latest
which cars-il-mcp
```

אם מתקבלת השגיאה `go: command not found` במחשב הזה, משתמשים בבינארי הקיים של Printing Press:

```bash
/Users/mark.s/printing-press/library/cars-il/cars-il-pp-mcp
```

ב-macOS קובץ ההגדרות של Claude Desktop נמצא בדרך כלל כאן:

```bash
$HOME/Library/Application Support/Claude/claude_desktop_config.json
```

להוסיף:

```json
{
  "mcpServers": {
    "cars-il": {
      "type": "stdio",
      "command": "/Users/YOUR_USER/go/bin/cars-il-mcp",
      "args": [],
      "env": {
        "CARS_IL_YAD2_COOKIE": ""
      }
    }
  }
}
```

לאחר מכן מפעילים מחדש את Claude Desktop ושואלים:

```text
Use the cars-il MCP. Call context and doctor, then search Yad2 live for 5 Toyota Corolla listings.
```

עבור Claude Code:

```bash
claude mcp add-json cars-il '{
  "type": "stdio",
  "command": "/Users/YOUR_USER/go/bin/cars-il-mcp",
  "args": [],
  "env": {
    "CARS_IL_YAD2_COOKIE": ""
  }
}'
```

בטיחות: שרת ה-MCP מבצע קריאות קריאה בלבד מול יד2. הפקודה `sync` כותבת רק למסד SQLite מקומי אצלכם במחשב.
