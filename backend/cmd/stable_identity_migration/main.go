package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoURI string
	MongoDB  string
}

// BootstrapJSON defines the structure for bootstrap config
type BootstrapJSON struct {
	Mongodb struct {
		Database string `json:"database"`
		URI      string `json:"uri"`
	} `json:"mongodb"`
}

func main() {
	var configPath string
	var apply bool
	flag.StringVar(&configPath, "config", "bootstrap.json", "path to bootstrap.json config file")
	flag.BoolVar(&apply, "apply", false, "apply modifications to database (otherwise dry-run)")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cfg := loadConfig(configPath)
	if cfg.MongoURI == "" || cfg.MongoDB == "" {
		log.Fatalf("Error: MongoDB connection details are missing")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(cfg.MongoDB)
	fmt.Printf("Connected to Database: %s (Apply changes: %t)\n\n", cfg.MongoDB, apply)

	// Step 1: Backfill users.uid and create unique index
	backfillUsersUID(ctx, db, apply)

	// Step 2: Backfill shopusers.useruid and create index
	backfillShopUsersUID(ctx, db, apply)

	// Step 3: Backfill poapprovalsettings approvers
	backfillPOApprovalSettings(ctx, db, apply)

	// Step 4: Backfill poapprovalstatus history
	backfillPOApprovalHistory(ctx, db, apply)

	printAuditCounts(ctx, db)
	printIndexStatus(ctx, db)

	fmt.Println("\nMigration processing completed.")
}

func loadConfig(path string) Config {
	var cfg Config
	// Try loading from environment first
	cfg.MongoURI = os.Getenv("MONGODB_URI")
	cfg.MongoDB = os.Getenv("MONGO_DB_NAME")

	if cfg.MongoURI != "" && cfg.MongoDB != "" {
		return cfg
	}

	// Try reading bootstrap.json
	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
		bytes, err := io.ReadAll(file)
		if err == nil {
			var boot BootstrapJSON
			if err := json.Unmarshal(bytes, &boot); err == nil {
				if cfg.MongoURI == "" {
					cfg.MongoURI = boot.Mongodb.URI
				}
				if cfg.MongoDB == "" {
					cfg.MongoDB = boot.Mongodb.Database
				}
			}
		}
	}
	return cfg
}

func generateGUID() string {
	// Simple UUID-based GUID generator mimicking utils.NewGUID
	return uuid.New().String()
}

func backfillUsersUID(ctx context.Context, db *mongo.Database, apply bool) {
	fmt.Println("--- 1. Processing 'users' collection ---")
	col := db.Collection("users")

	// Find users without a uid, or with empty uid
	filter := bson.M{
		"$or": []bson.M{
			{"uid": bson.M{"$exists": false}},
			{"uid": nil},
			{"uid": ""},
		},
	}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		log.Printf("Failed to query users: %v\n", err)
		return
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Failed to decode user: %v\n", err)
			continue
		}

		id := doc["_id"]
		username, _ := doc["username"].(string)
		newUID := generateGUID()

		fmt.Printf("  [User] Missing UID for: %s -> generating UID: %s\n", username, newUID)
		if apply {
			_, err := col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"uid": newUID}})
			if err != nil {
				log.Printf("    Error updating user %v: %v\n", id, err)
			}
		}
		count++
	}

	fmt.Printf("  Total users updated/needed backfill: %d\n", count)

	if apply {
		fmt.Println("  Creating unique index 'uxusersuid' on users(uid)...")
		indexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: "uid", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uxusersuid"),
		}
		name, err := col.Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			log.Printf("    Failed to create unique index on users: %v\n", err)
		} else {
			fmt.Printf("    Unique index created successfully: %s\n", name)
		}
	}
}

func backfillShopUsersUID(ctx context.Context, db *mongo.Database, apply bool) {
	fmt.Println("--- 2. Processing 'shopusers' collection ---")
	col := db.Collection("shopusers")
	usersCol := db.Collection("users")

	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to query shopusers: %v\n", err)
		return
	}
	defer cursor.Close(ctx)

	// Keep map of username -> uid in memory for speed
	userUIDMap := make(map[string]string)

	count := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Failed to decode shopUser: %v\n", err)
			continue
		}

		id := doc["_id"]
		username, _ := doc["username"].(string)
		holdingcode, _ := doc["holdingcode"].(string)
		existingUID, _ := doc["useruid"].(string)

		if existingUID != "" {
			continue // Already backfilled
		}

		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}

		// Find user UID
		uid, ok := userUIDMap[username]
		if !ok {
			var uDoc bson.M
			err := usersCol.FindOne(ctx, bson.M{"username": username}).Decode(&uDoc)
			if err == nil {
				uid, _ = uDoc["uid"].(string)
				userUIDMap[username] = uid
			}
		}

		if uid == "" {
			log.Printf("  [Warning] UID not found in users collection for username: %s\n", username)
			continue
		}

		fmt.Printf("  [ShopUser] Mapping username: %s in shop: %s -> useruid: %s\n", username, holdingcode, uid)
		if apply {
			_, err := col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"useruid": uid}})
			if err != nil {
				log.Printf("    Error updating shopUser %v: %v\n", id, err)
			}
		}
		count++
	}

	fmt.Printf("  Total shopusers mapped/needed backfill: %d\n", count)

	if apply {
		fmt.Println("  Creating index 'ix_shopusers_holdingcode_useruid' on shopusers(holdingcode, useruid)...")
		indexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: "holdingcode", Value: 1}, {Key: "useruid", Value: 1}},
			Options: options.Index().SetName("ix_shopusers_holdingcode_useruid"),
		}
		name, err := col.Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			if strings.Contains(err.Error(), "IndexOptionsConflict") {
				legacyName := "ix_shopusers_holdingcode_useruid"
				fmt.Printf("    Found same-key legacy index conflict. Dropping '%s' before recreating canonical name...\n", legacyName)
				if _, dropErr := col.Indexes().DropOne(ctx, legacyName); dropErr != nil {
					log.Printf("    Failed to drop legacy index on shopusers: %v\n", dropErr)
				} else if name, err = col.Indexes().CreateOne(ctx, indexModel); err != nil {
					log.Printf("    Failed to create canonical index on shopusers: %v\n", err)
				} else {
					fmt.Printf("    Index created successfully: %s\n", name)
				}
			} else {
				log.Printf("    Failed to create index on shopusers: %v\n", err)
			}
		} else {
			fmt.Printf("    Index created successfully: %s\n", name)
		}
	}
}

func backfillPOApprovalSettings(ctx context.Context, db *mongo.Database, apply bool) {
	fmt.Println("--- 3. Processing 'poapprovalsettings' collection ---")
	col := db.Collection("poapprovalsettings")
	usersCol := db.Collection("users")

	cursor, err := col.Find(ctx, bson.M{"rules.approvers.usercode": bson.M{"$exists": true}})
	if err != nil {
		log.Printf("Failed to query poapprovalsettings: %v\n", err)
		return
	}
	defer cursor.Close(ctx)

	userUIDMap := make(map[string]string)
	count := 0

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Failed to decode poapprovalsettings: %v\n", err)
			continue
		}

		id := doc["_id"]
		purchaseTypeCode, _ := doc["purchasetypecode"].(string)
		var rules []bson.M
		bytes, err := json.Marshal(doc["rules"])
		if err == nil {
			_ = json.Unmarshal(bytes, &rules)
		}
		if len(rules) == 0 {
			continue
		}

		modified := false
		for ruleIdx := range rules {
			var approvers []bson.M
			bytes, err := json.Marshal(rules[ruleIdx]["approvers"])
			if err == nil {
				_ = json.Unmarshal(bytes, &approvers)
			}
			if len(approvers) == 0 {
				continue
			}

			for approverIdx := range approvers {
				userCode, _ := approvers[approverIdx]["usercode"].(string)
				existingUID, _ := approvers[approverIdx]["approver_useruid"].(string)
				userCode = strings.TrimSpace(userCode)
				if existingUID != "" || userCode == "" {
					continue
				}

				uid, ok := userUIDMap[userCode]
				if !ok {
					var uDoc bson.M
					if err := usersCol.FindOne(ctx, bson.M{"username": userCode}).Decode(&uDoc); err == nil {
						uid, _ = uDoc["uid"].(string)
						userUIDMap[userCode] = uid
					}
				}
				if uid == "" {
					log.Printf("  [Warning] UID not found in users collection for approval usercode: %s\n", userCode)
					continue
				}

				approvers[approverIdx]["approver_useruid"] = uid
				modified = true
				fmt.Printf("  [PO Settings] PurchaseType: %s -> Mapping UserCode: %s -> approver_useruid: %s\n", purchaseTypeCode, userCode, uid)
			}
			rules[ruleIdx]["approvers"] = approvers
		}

		if modified && apply {
			if _, err := col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"rules": rules}}); err != nil {
				log.Printf("    Error updating PO approval setting %v: %v\n", id, err)
			}
			count++
		} else if modified {
			count++
		}
	}

	fmt.Printf("  Total PO approval settings backfilled/needed: %d\n", count)
}

func backfillPOApprovalHistory(ctx context.Context, db *mongo.Database, apply bool) {
	fmt.Println("--- 4. Processing 'poapprovalstatus' collection ---")
	col := db.Collection("poapprovalstatus")
	usersCol := db.Collection("users")

	cursor, err := col.Find(ctx, bson.M{"history": bson.M{"$exists": true}})
	if err != nil {
		log.Printf("Failed to query poapprovalstatus: %v\n", err)
		return
	}
	defer cursor.Close(ctx)

	userUIDMap := make(map[string]string)
	count := 0

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Failed to decode poapprovalstatus: %v\n", err)
			continue
		}

		id := doc["_id"]
		docno, _ := doc["docno"].(string)
		var history []bson.M

		// Workaround bson.M array conversion safely
		bytes, err := json.Marshal(doc["history"])
		if err == nil {
			json.Unmarshal(bytes, &history)
		}

		if len(history) == 0 {
			continue
		}

		modified := false
		for i, h := range history {
			actionBy, _ := h["actionby"].(string)
			existingUID, _ := h["approver_useruid"].(string)

			if existingUID != "" || actionBy == "" {
				continue
			}

			uid, ok := userUIDMap[actionBy]
			if !ok {
				var uDoc bson.M
				err := usersCol.FindOne(ctx, bson.M{"username": actionBy}).Decode(&uDoc)
				if err == nil {
					uid, _ = uDoc["uid"].(string)
					userUIDMap[actionBy] = uid
				}
			}

			if uid != "" {
				history[i]["approver_useruid"] = uid
				modified = true
				fmt.Printf("  [PO History] Document: %s -> Mapping ActionBy: %s -> approver_useruid: %s\n", docno, actionBy, uid)
			}
		}

		if modified && apply {
			_, err := col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"history": history}})
			if err != nil {
				log.Printf("    Error updating PO approval status %v: %v\n", id, err)
			}
			count++
		} else if modified {
			count++
		}
	}

	fmt.Printf("  Total PO approval documents backfilled/needed: %d\n", count)
}

func printAuditCounts(ctx context.Context, db *mongo.Database) {
	fmt.Println("\n--- Audit counts after migration pass ---")
	printCount := func(label string, collection string, filter bson.M) {
		count, err := db.Collection(collection).CountDocuments(ctx, filter)
		if err != nil {
			log.Printf("  %s: count failed: %v\n", label, err)
			return
		}
		fmt.Printf("  %s: %d\n", label, count)
	}

	missingUIDFilter := bson.M{"$or": []bson.M{{"uid": bson.M{"$exists": false}}, {"uid": nil}, {"uid": ""}}}
	missingUserUIDFilter := bson.M{"$or": []bson.M{{"useruid": bson.M{"$exists": false}}, {"useruid": nil}, {"useruid": ""}}}
	missingApproverUIDFilter := bson.M{
		"rules.approvers": bson.M{"$elemMatch": bson.M{
			"usercode": bson.M{"$exists": true, "$ne": ""},
			"$or": []bson.M{
				{"approver_useruid": bson.M{"$exists": false}},
				{"approver_useruid": nil},
				{"approver_useruid": ""},
			},
		}},
	}
	missingHistoryUIDFilter := bson.M{
		"history": bson.M{"$elemMatch": bson.M{
			"actionby": bson.M{"$exists": true, "$ne": ""},
			"$or": []bson.M{
				{"approver_useruid": bson.M{"$exists": false}},
				{"approver_useruid": nil},
				{"approver_useruid": ""},
			},
		}},
	}

	printCount("users missing uid", "users", missingUIDFilter)
	printCount("shopusers missing useruid", "shopusers", missingUserUIDFilter)
	printCount("poapprovalsettings approvers missing approver_useruid", "poapprovalsettings", missingApproverUIDFilter)
	printCount("poapprovalstatus history missing approver_useruid", "poapprovalstatus", missingHistoryUIDFilter)
}

func printIndexStatus(ctx context.Context, db *mongo.Database) {
	fmt.Println("\n--- Index status ---")
	printIndexExists := func(collection string, indexName string) {
		cursor, err := db.Collection(collection).Indexes().List(ctx)
		if err != nil {
			log.Printf("  %s.%s: list indexes failed: %v\n", collection, indexName, err)
			return
		}
		defer cursor.Close(ctx)

		found := false
		for cursor.Next(ctx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				continue
			}
			if name, _ := doc["name"].(string); name == indexName {
				found = true
				break
			}
		}
		fmt.Printf("  %s.%s exists: %t\n", collection, indexName, found)
	}

	printIndexExists("users", "uxusersuid")
	printIndexExists("shopusers", "ix_shopusers_holdingcode_useruid")
}
