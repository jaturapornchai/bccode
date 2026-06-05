package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"smlcloudplatform/internal/organization/branch/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const maxExampleRows = 50

type branchRecord struct {
	ID              primitive.ObjectID `bson:"id"`
	HoldingCode     string             `bson:"holdingcode"`
	GuidFixed       string             `bson:"guidfixed"`
	Code            any                `bson:"code"`
	IsVatRegistered bool               `bson:"isvatregistered"`
	POS             struct {
		TaxID string `bson:"taxid"`
	} `bson:"pos"`
}

type auditIssue struct {
	HoldingCode    string `json:"holdingcode"`
	GuidFixed      string `json:"guidfixed,omitempty"`
	Code           string `json:"code,omitempty"`
	NormalizedCode string `json:"normalizedcode,omitempty"`
	Reason         string `json:"reason"`
}

type auditSummary struct {
	Database                 string       `json:"database"`
	Collection               string       `json:"collection"`
	TotalBranches            int          `json:"totalbranches"`
	TotalCompanies           int          `json:"totalcompanies"`
	InvalidCodeCount         int          `json:"invalidcodecount"`
	DuplicateNormalizedCount int          `json:"duplicatenormalizedcount"`
	MissingHeadOfficeCount   int          `json:"missingheadofficecount"`
	VatMissingTaxIDCount     int          `json:"vatmissingtaxidcount"`
	Examples                 []auditIssue `json:"examples"`
}

type discoverSummary struct {
	Success   bool           `json:"success"`
	Summaries []auditSummary `json:"summaries"`
}

type companyAudit struct {
	hasHeadOffice bool
	branches      map[string][]branchRecord
}

func main() {
	uri := strings.TrimSpace(firstEnv("MONGODB_DEV_URI", "MONGODB_URI"))
	dbName := strings.TrimSpace(firstEnv("MONGODB_DEV_DB", "MONGODB_DB"))
	if dbName == "" {
		dbName = "dev"
	}
	if uri == "" {
		fatalJSON("MONGODB_DEV_URI or MONGODB_URI is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fatalJSON(err.Error())
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	if err := client.Ping(ctx, nil); err != nil {
		fatalJSON(err.Error())
	}

	if isTruthy(os.Getenv("BRANCH_AUDIT_DISCOVER")) {
		summaries, err := discoverAndAudit(ctx, client, dbName)
		if err != nil {
			fatalJSON(err.Error())
		}
		printJSON(discoverSummary{Success: true, Summaries: summaries})
		return
	}

	collectionName := strings.TrimSpace(firstEnv("BRANCH_AUDIT_COLLECTION"))
	if collectionName == "" {
		collectionName = "organizationBranches"
	}
	summary, err := auditCollection(ctx, client, dbName, collectionName)
	if err != nil {
		fatalJSON(err.Error())
	}
	printJSON(summary)
}

func discoverAndAudit(ctx context.Context, client *mongo.Client, fallbackDBName string) ([]auditSummary, error) {
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if len(databases) == 0 && fallbackDBName != "" {
		databases = []string{fallbackDBName}
	}

	candidateCollections := map[string]bool{
		"organizationBranches": true,
		"branch":               true,
	}
	summaries := []auditSummary{}
	for _, databaseName := range databases {
		collections, err := client.Database(databaseName).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			continue
		}
		for _, collectionName := range collections {
			if !candidateCollections[collectionName] {
				continue
			}
			summary, err := auditCollection(ctx, client, databaseName, collectionName)
			if err != nil {
				return nil, err
			}
			if summary.TotalBranches > 0 {
				summaries = append(summaries, summary)
			}
		}
	}
	return summaries, nil
}

func auditCollection(ctx context.Context, client *mongo.Client, dbName string, collectionName string) (auditSummary, error) {
	collection := client.Database(dbName).Collection(collectionName)
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{
		"holdingcode":       1,
		"guidfixed":         1,
		"code":              1,
		"is_vat_registered": 1,
		"pos.tax_id":        1,
	}))
	if err != nil {
		return auditSummary{}, err
	}
	defer cursor.Close(ctx)

	companies := map[string]*companyAudit{}
	summary := auditSummary{
		Database:   dbName,
		Collection: collectionName,
		Examples:   []auditIssue{},
	}

	for cursor.Next(ctx) {
		var record branchRecord
		if err := cursor.Decode(&record); err != nil {
			return auditSummary{}, err
		}
		summary.TotalBranches++

		holdingCode := strings.TrimSpace(record.HoldingCode)
		if holdingCode == "" {
			holdingCode = "(missing)"
		}
		if companies[holdingCode] == nil {
			companies[holdingCode] = &companyAudit{branches: map[string][]branchRecord{}}
		}

		code := branchCodeString(record.Code)
		normalizedCode, err := models.NormalizeThaiTaxBranchCode(code)
		if err != nil {
			summary.InvalidCodeCount++
			addExample(&summary, auditIssue{HoldingCode: holdingCode, GuidFixed: record.GuidFixed, Code: code, Reason: err.Error()})
			continue
		}

		if normalizedCode == models.ThaiHeadOfficeBranchCode {
			companies[holdingCode].hasHeadOffice = true
		}
		companies[holdingCode].branches[normalizedCode] = append(companies[holdingCode].branches[normalizedCode], record)

		if record.IsVatRegistered && strings.TrimSpace(record.POS.TaxID) == "" {
			summary.VatMissingTaxIDCount++
			addExample(&summary, auditIssue{HoldingCode: holdingCode, GuidFixed: record.GuidFixed, Code: code, NormalizedCode: normalizedCode, Reason: "VAT registered branch has empty pos.tax_id"})
		}
	}
	if err := cursor.Err(); err != nil {
		return auditSummary{}, err
	}

	summary.TotalCompanies = len(companies)
	for holdingCode, company := range companies {
		if !company.hasHeadOffice {
			summary.MissingHeadOfficeCount++
			addExample(&summary, auditIssue{HoldingCode: holdingCode, NormalizedCode: models.ThaiHeadOfficeBranchCode, Reason: "company has no head-office branch 00000"})
		}
		for normalizedCode, records := range company.branches {
			if len(records) <= 1 {
				continue
			}
			summary.DuplicateNormalizedCount++
			addExample(&summary, auditIssue{HoldingCode: holdingCode, NormalizedCode: normalizedCode, Reason: fmt.Sprintf("%d branches normalize to the same tax branch code", len(records))})
		}
	}

	sort.Slice(summary.Examples, func(i, j int) bool {
		if summary.Examples[i].HoldingCode == summary.Examples[j].HoldingCode {
			return summary.Examples[i].Reason < summary.Examples[j].Reason
		}
		return summary.Examples[i].HoldingCode < summary.Examples[j].HoldingCode
	})

	return summary, nil
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func branchCodeString(value any) string {
	switch code := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(code)
	case int32:
		return fmt.Sprintf("%d", code)
	case int64:
		return fmt.Sprintf("%d", code)
	case int:
		return fmt.Sprintf("%d", code)
	case float64:
		if code == float64(int64(code)) {
			return fmt.Sprintf("%d", int64(code))
		}
		return fmt.Sprintf("%v", code)
	default:
		return fmt.Sprintf("%v", code)
	}
}

func addExample(summary *auditSummary, issue auditIssue) {
	if len(summary.Examples) >= maxExampleRows {
		return
	}
	summary.Examples = append(summary.Examples, issue)
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func printJSON(value any) {
	output, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fatalJSON(err.Error())
	}
	fmt.Println(string(output))
}

func fatalJSON(message string) {
	output, _ := json.MarshalIndent(map[string]any{
		"success": false,
		"message": message,
	}, "", "  ")
	fmt.Println(string(output))
	os.Exit(1)
}
