package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	coreconfig "smlcloudplatform/internal/config"
	"sort"
	"strings"
	"time"
	"unicode"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type renameItem struct {
	Target  string
	OldName string
	NewName string
	Status  string
	Reason  string
}

type config struct {
	target             string
	apply              bool
	timeout            time.Duration
	mongoURI           string
	mongoDatabase      string
	postgresDSN        string
	postgresSchema     string
	clickHouseURL      string
	clickHouseDatabase string
	clickHouseUser     string
	clickHousePassword string
}

var legacyStorageNameAliases = map[string]string{
	"advancepayment":                          "advancepayment",
	"advancepaymentrefund":                    "advancepaymentrefund",
	"apadvancepaymentrefundtransaction":       "apadvancepaymentrefundtransaction",
	"apadvancepaymentrefundtransactiondetail": "apadvancepaymentrefundtransactiondetail",
	"apadvancepaymenttransaction":             "apadvancepaymenttransaction",
	"apadvancepaymenttransactiondetail":       "apadvancepaymenttransactiondetail",
	"apdepositpaymentrefundtransaction":       "apdepositpaymentrefundtransaction",
	"apdepositpaymentrefundtransactiondetail": "apdepositpaymentrefundtransactiondetail",
	"apdepositpaymenttransaction":             "apdepositpaymenttransaction",
	"apdepositpaymenttransactiondetail":       "apdepositpaymenttransactiondetail",
	"appurchasereceivetransaction":            "appurchasereceivetransaction",
	"appurchasereceivetransactiondetail":      "appurchasereceivetransactiondetail",
	"aradvancepaymentrefundtransaction":       "aradvancepaymentrefundtransaction",
	"aradvancepaymentrefundtransactiondetail": "aradvancepaymentrefundtransactiondetail",
	"aradvancepaymenttransaction":             "aradvancepaymenttransaction",
	"aradvancepaymenttransactiondetail":       "aradvancepaymenttransactiondetail",
	"ardepositpaymentrefundtransaction":       "ardepositpaymentrefundtransaction",
	"ardepositpaymentrefundtransactiondetail": "ardepositpaymentrefundtransactiondetail",
	"ardepositpaymenttransaction":             "ardepositpaymenttransaction",
	"ardepositpaymenttransactiondetail":       "ardepositpaymenttransactiondetail",
	"bankmaster":                              "bankmaster",
	"banktransferrecord":                      "banktransferrecord",
	"bookbank":                                "bookbank",
	"cartorder":                               "cartorder",
	"cartorderdetail":                         "cartorderdetail",
	"chartofaccounts":                         "chartofaccounts",
	"chequechange":                            "chequechange",
	"chequedeposit":                           "chequedeposit",
	"chequedisqualified":                      "chequedisqualified",
	"chequepass":                              "chequepass",
	"chequepaymentchange":                     "chequepaymentchange",
	"chequepaymentdeposit":                    "chequepaymentdeposit",
	"chequepaymentdisqualified":               "chequepaymentdisqualified",
	"chequepaymentreturn":                     "chequepaymentreturn",
	"chequerenew":                             "chequerenew",
	"chequereturn":                            "chequereturn",
	"costcenter":                              "costcenter",
	"creditcardwithdrawal":                    "creditcardwithdrawal",
	"depositrecord":                           "depositrecord",
	"depositrefund":                           "depositrefund",
	"docdetail":                               "docdetail",
	"docdetailupdated":                        "docdetailupdated",
	"docpayment":                              "docpayment",
	"docref":                                  "docref",
	"docwaitprocess":                          "docwaitprocess",
	"inventoryoptions":                        "inventoryoptions",
	"jobproject":                              "jobproject",
	"journalvatsdetails":                      "journalvatsdetails",
	"journaltaxesdetails":                     "journaltaxesdetails",
	"ordertype":                               "ordertype",
	"paymentmaster":                           "paymentmaster",
	"paidadvance":                             "paidadvance",
	"paidadvancerefund":                       "paidadvancerefund",
	"pickandpack":                             "pickandpack",
	"processstock":                            "processstock",
	"processstockcost":                        "processstockcost",
	"processstockdetail":                      "processstockdetail",
	"processstocklot":                         "processstocklot",
	"productbarcode":                          "productbarcodes",
	"productbarcodedict":                      "productbarcodedict",
	"productbarcodeboms":                      "productbarcodeboms",
	"productbarcodeimport":                    "productbarcodeimport",
	"productbarcodeprocess":                   "productbarcodeprocess",
	"productbarcoderef":                       "productbarcoderef",
	"productcategory":                         "productcategory",
	"producttype":                             "producttype",
	"productunit":                             "productunit",
	"purchaseorder":                           "purchaseorder",
	"purchaseordertransaction":                "purchaseordertransaction",
	"purchaseordertransactiondetail":          "purchaseordertransactiondetail",
	"purchasepartial":                         "purchasepartial",
	"purchasereceivetransaction":              "purchasereceivetransaction",
	"purchasereceivetransactiondetail":        "purchasereceivetransactiondetail",
	"purchaserequisition":                     "purchaserequisition",
	"purchasedebitnote":                       "purchasedebitnote",
	"receivedeposit":                          "receivedeposit",
	"receivedepositrefund":                    "receivedepositrefund",
	"resultfordashboard":                      "resultfordashboard",
	"qrpayment":                               "qrpayment",
	"salechannel":                             "salechannel",
	"saledebitnote":                           "saledebitnote",
	"saledebitnotetransaction":                "saledebitnotetransaction",
	"saledebitnotetransactiondetail":          "saledebitnotetransactiondetail",
	"saleinvoice":                             "saleinvoice",
	"saleinvoicereturntransaction":            "saleinvoicereturntransaction",
	"saleinvoicereturntransactiondetail":      "saleinvoicereturntransactiondetail",
	"saleinvoicetransaction":                  "saleinvoicetransaction",
	"saleinvoicetransactiondetail":            "saleinvoicetransactiondetail",
	"saleorder":                               "saleorder",
	"saleordertransaction":                    "saleordertransaction",
	"saleordertransactiondetail":              "saleordertransactiondetail",
	"stockadjustment":                         "stockadjustment",
	"stockbalanceimport":                      "stockbalanceimport",
	"stockwaitprocess":                        "stockwaitprocess",
	"userlogin":                               "userlogin",
	"withdrawalrecord":                        "withdrawalrecord",
}

var legacyStorageFieldNameAliases = map[string]string{
	"branchcode":     "branchcode",
	"custcode":       "custcode",
	"departmentcode": "departmentcode",
	"docdate":        "docdate",
	"docdatetime":    "docdatetime",
	"docno":          "docno",
	"itemcode":       "itemcode",
	"locationcode":   "locationcode",
	"holdingcode":    "holdingcode",
	"transflag":      "transflag",
	"unitcode":       "unitcode",
	"warehousecode":  "warehousecode",
	"whcode":         "whcode",
}

func main() {
	cfg := parseConfig()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	var items []renameItem
	var err error
	switch cfg.target {
	case "all":
		items, err = runAll(ctx, cfg)
	case "mongo":
		items, err = runMongo(ctx, cfg)
	case "postgres":
		items, err = runPostgres(ctx, cfg)
	case "clickhouse":
		items, err = runClickHouse(ctx, cfg)
	default:
		err = fmt.Errorf("unknown target %q, use all, mongo, postgres, or clickhouse", cfg.target)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "storage name migration failed: %v\n", err)
		os.Exit(1)
	}

	printItems(items)
}

func parseConfig() config {
	var cfg config
	var timeout string
	flag.StringVar(&cfg.target, "target", "all", "target database: all, mongo, postgres, clickhouse")
	flag.BoolVar(&cfg.apply, "apply", false, "execute renames; default is dry-run")
	flag.StringVar(&timeout, "timeout", "30s", "operation timeout")
	flag.StringVar(&cfg.mongoURI, "mongo-uri", coreconfig.MongoURIForCurrentEnvironment(), "MongoDB URI")
	flag.StringVar(&cfg.mongoDatabase, "mongo-db", coreconfig.MongoDatabaseForCurrentEnvironment(""), "MongoDB database name")
	flag.StringVar(&cfg.postgresDSN, "postgres-dsn", os.Getenv("POSTGRESDSN"), "PostgreSQL DSN")
	flag.StringVar(&cfg.postgresSchema, "postgres-schema", getenvDefault("POSTGRESSCHEMA", "public"), "PostgreSQL schema name")
	flag.StringVar(&cfg.clickHouseURL, "clickhouse-url", getenvDefault("CLICKHOUSEHTTPURL", "http://localhost:8123"), "ClickHouse HTTP URL")
	flag.StringVar(&cfg.clickHouseDatabase, "clickhouse-db", os.Getenv("CLICKHOUSEDATABASE"), "ClickHouse database name")
	flag.StringVar(&cfg.clickHouseUser, "clickhouse-user", os.Getenv("CLICKHOUSEUSER"), "ClickHouse username")
	flag.StringVar(&cfg.clickHousePassword, "clickhouse-password", os.Getenv("CLICKHOUSEPASSWORD"), "ClickHouse password")
	flag.Parse()

	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid timeout %q: %v\n", timeout, err)
		os.Exit(2)
	}
	cfg.timeout = parsedTimeout
	return cfg
}

func runAll(ctx context.Context, cfg config) ([]renameItem, error) {
	var items []renameItem
	for _, runner := range []func(context.Context, config) ([]renameItem, error){runMongo, runPostgres, runClickHouse} {
		next, err := runner(ctx, cfg)
		if err != nil {
			return items, err
		}
		items = append(items, next...)
	}
	return items, nil
}

func runMongo(ctx context.Context, cfg config) ([]renameItem, error) {
	if cfg.mongoURI == "" || cfg.mongoDatabase == "" {
		return []renameItem{{Target: "mongo", Status: "skipped", Reason: "missing --mongo-uri or --mongo-db"}}, nil
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.mongoURI))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.mongoDatabase)
	names, err := db.ListCollectionNames(ctx, map[string]any{})
	if err != nil {
		return nil, err
	}

	existing := toSet(names)
	var items []renameItem
	for _, oldName := range names {
		newName := normalizeStorageName(oldName)
		if oldName == newName {
			continue
		}
		item := renameItem{Target: "mongo", OldName: oldName, NewName: newName}
		if existing[newName] {
			item.Status = "conflict"
			item.Reason = "target collection already exists"
			items = append(items, item)
			continue
		}
		if cfg.apply {
			command := bson.D{
				{Key: "renameCollection", Value: cfg.mongoDatabase + "." + oldName},
				{Key: "to", Value: cfg.mongoDatabase + "." + newName},
				{Key: "dropTarget", Value: false},
			}
			if err := client.Database("admin").RunCommand(ctx, command).Err(); err != nil {
				return items, err
			}
			item.Status = "renamed"
		} else {
			item.Status = "dry-run"
		}
		delete(existing, oldName)
		existing[newName] = true
		items = append(items, item)
	}
	return items, nil
}

func runPostgres(ctx context.Context, cfg config) ([]renameItem, error) {
	if cfg.postgresDSN == "" {
		return []renameItem{{Target: "postgres", Status: "skipped", Reason: "missing --postgres-dsn"}}, nil
	}
	if !isSafeIdentifier(cfg.postgresSchema) {
		return nil, fmt.Errorf("unsafe postgres schema name %q", cfg.postgresSchema)
	}

	db, err := sql.Open("postgres", cfg.postgresDSN)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name`, cfg.postgresSchema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return renamePostgresTables(ctx, db, cfg, names)
}

func renamePostgresTables(ctx context.Context, db *sql.DB, cfg config, names []string) ([]renameItem, error) {
	existing := toSet(names)
	var tx *sql.Tx
	var err error
	if cfg.apply {
		tx, err = db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()
	}

	var items []renameItem
	for _, oldName := range names {
		newName := normalizeStorageName(oldName)
		if oldName == newName {
			continue
		}
		item := renameItem{Target: "postgres", OldName: oldName, NewName: newName}
		if existing[newName] {
			item.Status = "conflict"
			item.Reason = "target table already exists"
			items = append(items, item)
			continue
		}
		if !isSafeIdentifier(oldName) || !isSafeIdentifier(newName) {
			item.Status = "skipped"
			item.Reason = "unsafe identifier"
			items = append(items, item)
			continue
		}
		if cfg.apply {
			stmt := fmt.Sprintf("ALTER TABLE %s.%s RENAME TO %s", quotePGIdent(cfg.postgresSchema), quotePGIdent(oldName), quotePGIdent(newName))
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return items, err
			}
			item.Status = "renamed"
		} else {
			item.Status = "dry-run"
		}
		delete(existing, oldName)
		existing[newName] = true
		items = append(items, item)
	}

	if cfg.apply {
		if err := tx.Commit(); err != nil {
			return items, err
		}
	}
	return items, nil
}

func runClickHouse(ctx context.Context, cfg config) ([]renameItem, error) {
	if cfg.clickHouseDatabase == "" {
		return []renameItem{{Target: "clickhouse", Status: "skipped", Reason: "missing --clickhouse-db"}}, nil
	}
	if !isSafeIdentifier(cfg.clickHouseDatabase) {
		return nil, fmt.Errorf("unsafe clickhouse database name %q", cfg.clickHouseDatabase)
	}

	namesText, err := clickHouseQuery(ctx, cfg, "SHOW TABLES FROM "+quoteClickHouseIdent(cfg.clickHouseDatabase))
	if err != nil {
		return nil, err
	}

	names := splitLines(namesText)
	existing := toSet(names)
	var items []renameItem
	for _, oldName := range names {
		newName := normalizeStorageName(oldName)
		if oldName == newName {
			continue
		}
		item := renameItem{Target: "clickhouse", OldName: oldName, NewName: newName}
		if existing[newName] {
			item.Status = "conflict"
			item.Reason = "target table already exists"
			items = append(items, item)
			continue
		}
		if !isSafeIdentifier(oldName) || !isSafeIdentifier(newName) {
			item.Status = "skipped"
			item.Reason = "unsafe identifier"
			items = append(items, item)
			continue
		}
		if cfg.apply {
			query := fmt.Sprintf(
				"RENAME TABLE %s.%s TO %s.%s",
				quoteClickHouseIdent(cfg.clickHouseDatabase),
				quoteClickHouseIdent(oldName),
				quoteClickHouseIdent(cfg.clickHouseDatabase),
				quoteClickHouseIdent(newName),
			)
			if _, err := clickHouseQuery(ctx, cfg, query); err != nil {
				return items, err
			}
			item.Status = "renamed"
		} else {
			item.Status = "dry-run"
		}
		delete(existing, oldName)
		existing[newName] = true
		items = append(items, item)
	}
	return items, nil
}

func clickHouseQuery(ctx context.Context, cfg config, query string) (string, error) {
	endpoint, err := url.Parse(cfg.clickHouseURL)
	if err != nil {
		return "", err
	}
	if endpoint.Scheme == "" {
		endpoint.Scheme = "http"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewBufferString(query))
	if err != nil {
		return "", err
	}
	if cfg.clickHouseUser != "" {
		req.SetBasicAuth(cfg.clickHouseUser, cfg.clickHousePassword)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("clickhouse status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return string(body), nil
}

func normalizeStorageName(name string) string {
	return normalizeStorageIdentifier(name, legacyStorageNameAliases)
}

func normalizeStorageFieldName(name string) string {
	return normalizeStorageIdentifier(name, legacyStorageFieldNameAliases)
}

func normalizeStorageIdentifier(name string, aliases map[string]string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	var builder strings.Builder
	for _, current := range name {
		if current == '_' || current == '-' || unicode.IsSpace(current) {
			continue
		}

		if unicode.IsUpper(current) {
			builder.WriteRune(unicode.ToLower(current))
			continue
		}

		builder.WriteRune(unicode.ToLower(current))
	}

	normalized := builder.String()
	if alias, ok := aliases[normalized]; ok {
		return alias
	}
	return normalized
}

func printItems(items []renameItem) {
	if len(items) == 0 {
		fmt.Println("storage name migration: no rename needed")
		return
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Target == items[j].Target {
			return items[i].OldName < items[j].OldName
		}
		return items[i].Target < items[j].Target
	})

	for _, item := range items {
		if item.OldName == "" {
			fmt.Printf("[%s] %s: %s\n", item.Target, item.Status, item.Reason)
			continue
		}
		if item.Reason != "" {
			fmt.Printf("[%s] %s %s -> %s (%s)\n", item.Target, item.Status, item.OldName, item.NewName, item.Reason)
			continue
		}
		fmt.Printf("[%s] %s %s -> %s\n", item.Target, item.Status, item.OldName, item.NewName)
	}
}

func splitLines(text string) []string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func toSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

func isSafeIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '_' {
			continue
		}
		return false
	}
	return true
}

func quotePGIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func quoteClickHouseIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func getenvDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
