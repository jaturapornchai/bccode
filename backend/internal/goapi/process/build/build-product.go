package build

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/product/projection"
	"smlcloudplatform/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongoProductProjection struct {
	BusinessCode string `bson:"businesscode"`
	Code         string `bson:"code"`
	UnitCode     string `bson:"unitcode"`
	Names        []struct {
		Name string `bson:"name"`
	} `bson:"names"`
	UnitNames []struct {
		Name string `bson:"name"`
	} `bson:"unitnames"`
}

type productProjectionValue struct {
	Name     string
	UnitCode string
	UnitName string
}

// ProcessProductRebuildAll reconciles the PostgreSQL product projection with
// active MongoDB product masters. Barcode-only records intentionally remain
// outside this table until they are linked to a product master by code.
func ProcessProductRebuildAll(holdingCode string) (int, error) {
	return 0, fmt.Errorf("company-scoped product rebuild requires businesscode")
}

// ProcessProductRebuildCompany reconciles only one company's product metadata.
// Stock columns are deliberately left untouched until stock projections carry
// BusinessCode too.
func ProcessProductRebuildCompany(holdingCode string, businessCode string) (int, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if holdingCode == "" || businessCode == "" {
		return 0, fmt.Errorf("holdingcode and businesscode are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	mongoClient := myglobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return 0, fmt.Errorf("connect MongoDB for product projection")
	}

	pgDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return 0, fmt.Errorf("connect PostgreSQL product projection: %w", err)
	}
	if err := TableProductCreate(pgDB); err != nil {
		return 0, err
	}
	tx, err := pgDB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin product projection rebuild: %w", err)
	}
	defer tx.Rollback()

	// Exclude event reconciliation before taking the MongoDB snapshot.
	if _, err := tx.ExecContext(ctx, projection.ExclusiveSQL, projection.LockKey("product", holdingCode, businessCode, "")); err != nil {
		return 0, err
	}

	collection := mongoClient.
		Database(config.NewServiceConfig().MongodbDatabaseName()).
		Collection("products", options.Collection().SetReadPreference(readpref.Primary()))
	cursor, err := collection.Find(
		ctx,
		bson.M{
			"holdingcode":  holdingCode,
			"businesscode": businessCode,
			"deletedat":    bson.M{"$exists": false},
		},
		options.Find().
			SetProjection(bson.M{"businesscode": 1, "code": 1, "names": 1, "unitcode": 1, "unitnames": 1}).
			SetSort(bson.D{{Key: "code", Value: 1}, {Key: "_id", Value: 1}}),
	)
	if err != nil {
		return 0, fmt.Errorf("read MongoDB products: %w", err)
	}
	defer cursor.Close(ctx)

	products := make(map[string]productProjectionValue)
	for cursor.Next(ctx) {
		var product mongoProductProjection
		if err := cursor.Decode(&product); err != nil {
			return 0, fmt.Errorf("decode MongoDB product: %w", err)
		}
		if utils.NormalizeBusinessCode(product.BusinessCode) != businessCode {
			return 0, fmt.Errorf("MongoDB product belongs to a different company")
		}
		code := utils.NormalizeBusinessCode(product.Code)
		if code == "" {
			return 0, fmt.Errorf("active MongoDB product has an empty code")
		}
		if _, exists := products[code]; exists {
			return 0, fmt.Errorf("duplicate active MongoDB product code %q", code)
		}
		unitCode := utils.NormalizeBusinessCode(product.UnitCode)
		if unitCode == "" {
			return 0, fmt.Errorf("active MongoDB product %q has an empty base unit", code)
		}
		name := code
		for _, item := range product.Names {
			if item.Name != "" {
				name = item.Name
				break
			}
		}
		unitName := unitCode
		for _, item := range product.UnitNames {
			if item.Name != "" {
				unitName = item.Name
				break
			}
		}
		products[code] = productProjectionValue{Name: name, UnitCode: unitCode, UnitName: unitName}
	}
	if err := cursor.Err(); err != nil {
		return 0, fmt.Errorf("iterate MongoDB products: %w", err)
	}

	codes := make([]string, 0, len(products))
	for code := range products {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	if len(codes) == 0 {
		if _, err := tx.ExecContext(
			ctx,
			"DELETE FROM product WHERE holding_code = $1 AND businesscode = $2",
			holdingCode,
			businessCode,
		); err != nil {
			return 0, fmt.Errorf("clear empty PostgreSQL product projection: %w", err)
		}
	} else {
		placeholders := make([]string, len(codes))
		args := make([]interface{}, 0, len(codes)+2)
		args = append(args, holdingCode, businessCode)
		for index, code := range codes {
			placeholders[index] = fmt.Sprintf("$%d", index+3)
			args = append(args, code)
		}
		query := fmt.Sprintf(
			"DELETE FROM product WHERE holding_code = $1 AND businesscode = $2 AND itemcode NOT IN (%s)",
			strings.Join(placeholders, ","),
		)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return 0, fmt.Errorf("remove stale PostgreSQL products: %w", err)
		}
	}
	upsert, err := tx.PrepareContext(ctx, `
		INSERT INTO product (holding_code, businesscode, itemcode, name0, unitcode, unitname)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT ON CONSTRAINT product_company_itemcode_unique
		DO UPDATE SET
			name0 = EXCLUDED.name0,
			unitcode = EXCLUDED.unitcode,
			unitname = EXCLUDED.unitname`)
	if err != nil {
		return 0, fmt.Errorf("prepare PostgreSQL product projection upsert: %w", err)
	}
	defer upsert.Close()

	for _, code := range codes {
		product := products[code]
		if _, err := upsert.ExecContext(ctx, holdingCode, businessCode, code, product.Name, product.UnitCode, product.UnitName); err != nil {
			return 0, fmt.Errorf("upsert PostgreSQL product %s: %w", code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit product projection rebuild: %w", err)
	}
	logger.Info("Rebuilt %d PostgreSQL products from MongoDB for holding %s company %s", len(codes), holdingCode, businessCode)
	return len(codes), nil
}
