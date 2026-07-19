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
	"smlcloudplatform/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoProductProjection struct {
	Code  string `bson:"code"`
	Names []struct {
		Name string `bson:"name"`
	} `bson:"names"`
}

// ProcessProductRebuildAll reconciles the PostgreSQL product projection with
// active MongoDB product masters. Barcode-only records intentionally remain
// outside this table until they are linked to a product master by code.
func ProcessProductRebuildAll(holdingCode string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	mongoClient := myglobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return 0, fmt.Errorf("connect MongoDB for product projection")
	}

	collection := mongoClient.
		Database(config.NewServiceConfig().MongodbDatabaseName()).
		Collection("products")
	cursor, err := collection.Find(
		ctx,
		bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}},
		options.Find().
			SetProjection(bson.M{"code": 1, "names": 1}).
			SetSort(bson.D{{Key: "code", Value: 1}, {Key: "_id", Value: 1}}),
	)
	if err != nil {
		return 0, fmt.Errorf("read MongoDB products: %w", err)
	}
	defer cursor.Close(ctx)

	productNames := make(map[string]string)
	for cursor.Next(ctx) {
		var product mongoProductProjection
		if err := cursor.Decode(&product); err != nil {
			return 0, fmt.Errorf("decode MongoDB product: %w", err)
		}
		code := utils.NormalizeBusinessCode(product.Code)
		if code == "" {
			return 0, fmt.Errorf("active MongoDB product has an empty code")
		}
		if _, exists := productNames[code]; exists {
			return 0, fmt.Errorf("duplicate active MongoDB product code %q", code)
		}
		name := code
		for _, item := range product.Names {
			if item.Name != "" {
				name = item.Name
				break
			}
		}
		productNames[code] = name
	}
	if err := cursor.Err(); err != nil {
		return 0, fmt.Errorf("iterate MongoDB products: %w", err)
	}

	pgDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return 0, fmt.Errorf("connect PostgreSQL product projection: %w", err)
	}
	if err := TableProductCreate(pgDB); err != nil {
		return 0, err
	}
	codes := make([]string, 0, len(productNames))
	for code := range productNames {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	tx, err := pgDB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin product projection rebuild: %w", err)
	}
	defer tx.Rollback()

	if len(codes) == 0 {
		if _, err := tx.ExecContext(ctx, "DELETE FROM product"); err != nil {
			return 0, fmt.Errorf("clear empty PostgreSQL product projection: %w", err)
		}
	} else {
		placeholders := make([]string, len(codes))
		args := make([]interface{}, len(codes))
		for index, code := range codes {
			placeholders[index] = fmt.Sprintf("$%d", index+1)
			args[index] = code
		}
		query := fmt.Sprintf("DELETE FROM product WHERE itemcode NOT IN (%s)", strings.Join(placeholders, ","))
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return 0, fmt.Errorf("remove stale PostgreSQL products: %w", err)
		}
	}
	upsert, err := tx.PrepareContext(ctx, `
		INSERT INTO product (itemcode, name0, unitcode, unitname)
		VALUES ($1, $2, '', '')
		ON CONFLICT ON CONSTRAINT product_itemcode_unique
		DO UPDATE SET
			name0 = EXCLUDED.name0,
			unitcode = '',
			unitname = ''`)
	if err != nil {
		return 0, fmt.Errorf("prepare PostgreSQL product projection upsert: %w", err)
	}
	defer upsert.Close()

	for _, code := range codes {
		if _, err := upsert.ExecContext(ctx, code, productNames[code]); err != nil {
			return 0, fmt.Errorf("upsert PostgreSQL product %s: %w", code, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE product p
		SET unitcode = units.unitcode,
			unitname = units.unitname
		FROM (
			SELECT DISTINCT ON (itemcode)
				itemcode,
				COALESCE(unitcode, '') AS unitcode,
				COALESCE(unitname, '') AS unitname
			FROM productbarcode
			WHERE itemcode IS NOT NULL AND itemcode <> ''
			ORDER BY itemcode,
				CASE WHEN barcoderefunitstand = 1 AND barcoderefunitdivide = 1 THEN 0 ELSE 1 END,
				barcode
		) units
		WHERE units.itemcode = p.itemcode`); err != nil {
		return 0, fmt.Errorf("enrich PostgreSQL product units: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit product projection rebuild: %w", err)
	}
	logger.Info("Rebuilt %d PostgreSQL products from MongoDB for shop %s", len(codes), holdingCode)
	return len(codes), nil
}
