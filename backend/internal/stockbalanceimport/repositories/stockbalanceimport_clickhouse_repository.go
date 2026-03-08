package repositories

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/stockbalanceimport/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/fatih/structs"
)

type IStockBalanceImportClickHouseRepository interface {
	All(ctx context.Context, shopID string, taskID string) ([]models.StockBalanceImportDoc, error)
	List(ctx context.Context, shopID string, taskID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceImportDoc, models.PaginationData, error)
	Create(ctx context.Context, doc models.StockBalanceImportDoc) error
	CreateInBatch(ctx context.Context, docs []models.StockBalanceImportDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StockBalanceImportRaw) error
	DeleteByGUID(ctx context.Context, shopID string, guid string) error
	DeleteByTaskID(ctx context.Context, shopID string, taskID string) error
	Meta(ctx context.Context, shopID string, taskID string) (models.StockBalanceImportMeta, error)
	FindOne(ctx context.Context, shopID string, taskID string, sorts []micromodels.KeyInt) (models.StockBalanceImportDoc, error)
	UpdateExist(ctx context.Context, shopID string, taskID string, isExist bool, barcodes []string) error
	CountExist(ctx context.Context, shopID string, taskID string, isExist bool) (int, error)
}

type StockBalanceImportClickHouseRepository struct {
	pst          microservice.IPersisterClickHouse
	structFileds map[string]struct{}
}

func NewStockBalanceImportClickHouseRepository(pst microservice.IPersisterClickHouse) StockBalanceImportClickHouseRepository {

	structFileds := make(map[string]struct{})
	fields := structs.Fields(models.StockBalanceImport{})

	for _, field := range fields {
		tag := field.Tag("ch")
		structFileds[tag] = struct{}{}
	}

	return StockBalanceImportClickHouseRepository{
		pst: pst,
	}
}

func (repo StockBalanceImportClickHouseRepository) All(ctx context.Context, shopID string, taskID string) ([]models.StockBalanceImportDoc, error) {

	results := []models.StockBalanceImportDoc{}

	sqlExpr := "SELECT * FROM stockbalanceimport WHERE shopid = ? AND taskid = ?"
	err := repo.pst.Select(ctx, &results, sqlExpr, shopID, taskID)

	if err != nil {
		return results, err
	}

	return results, nil
}

func (repo StockBalanceImportClickHouseRepository) Meta(ctx context.Context, shopID string, taskID string) (models.StockBalanceImportMeta, error) {

	results := []models.StockBalanceImportMeta{}

	sqlExpr := "SELECT COUNT(*) totalitem, SUM(sumamount) totalamount FROM stockbalanceimport WHERE shopid = ? AND taskid = ?"
	err := repo.pst.Select(ctx, &results, sqlExpr, shopID, taskID)

	if err != nil {
		return models.StockBalanceImportMeta{}, err
	}

	return results[0], nil
}

func (repo StockBalanceImportClickHouseRepository) FindOne(ctx context.Context, shopID string, taskID string, sorts []micromodels.KeyInt) (models.StockBalanceImportDoc, error) {

	orderExpr := ""
	if len(sorts) > 0 {
		for _, sort := range sorts {
			if orderExpr != "" {
				orderExpr += ", "
			}

			orderTxt := "ASC"
			if sort.Value == -1 {
				orderTxt = "DESC"
			}

			if _, ok := repo.structFileds[sort.Key]; !ok {
				orderExpr += fmt.Sprintf("%s %s", sort.Key, orderTxt)
			}

		}
	}

	if orderExpr != "" {
		orderExpr = fmt.Sprintf("ORDER BY %s", orderExpr)
	}

	results := []models.StockBalanceImportDoc{}

	sqlExpr := fmt.Sprintf("SELECT * FROM stockbalanceimport WHERE shopid = ? AND taskid = ? %s LIMIT 1 OFFSET 0", orderExpr)
	err := repo.pst.Select(ctx, &results, sqlExpr, shopID, taskID)

	if err != nil {
		return models.StockBalanceImportDoc{}, err
	}

	if len(results) == 0 {
		return models.StockBalanceImportDoc{}, nil
	}

	return results[0], nil
}

func (repo StockBalanceImportClickHouseRepository) List(ctx context.Context, shopID string, taskID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceImportDoc, models.PaginationData, error) {

	offset := (pageable.Page - 1) * pageable.Limit

	filterExpr := ""
	filterArgs := []interface{}{}
	if len(filters) > 0 {
		for key, value := range filters {
			if filterExpr != "" {
				filterExpr += " AND "
			}

			filterExpr += fmt.Sprintf("%s = ?", key)
			filterArgs = append(filterArgs, value)
		}
	}

	if filterExpr != "" {
		filterExpr = fmt.Sprintf("AND (%s)", filterExpr)
	}

	orderExpr := ""
	if len(pageable.Sorts) > 0 {
		for _, sort := range pageable.Sorts {
			if orderExpr != "" {
				orderExpr += ", "
			}

			orderTxt := "ASC"
			if sort.Value == -1 {
				orderTxt = "DESC"
			}

			if colName, ok := repo.structFileds[sort.Key]; !ok {
				orderExpr += fmt.Sprintf("%s %s", colName, orderTxt)
			}

		}
	}

	if orderExpr != "" {
		orderExpr = fmt.Sprintf("ORDER BY %s", orderExpr)
	}

	searchExpr := ""
	searchArgs := []interface{}{}
	if pageable.Query != "" {
		searchExpr = "AND (barcode LIKE ? OR name LIKE ? OR unitcode LIKE ?)"
		searchTxt := fmt.Sprintf("%s%%", pageable.Query)
		searchArgs = append(searchArgs, searchTxt, searchTxt, searchTxt)
	}

	paginationData := models.PaginationData{}
	results := []models.StockBalanceImportDoc{}

	args := []interface{}{}
	args = append(args, shopID, taskID)
	args = append(args, filterArgs...)
	args = append(args, searchArgs...)
	args = append(args, pageable.Limit, offset)

	sqlExpr := fmt.Sprintf("SELECT * FROM stockbalanceimport WHERE shopid = ? AND taskid = ? %s %s %s LIMIT ? OFFSET ?", filterExpr, searchExpr, orderExpr)
	err := repo.pst.Select(ctx, &results, sqlExpr, args...)

	if err != nil {
		return results, paginationData, err
	}

	countArgs := []interface{}{}
	countArgs = append(countArgs, shopID, taskID)
	countArgs = append(countArgs, searchArgs...)

	exprCount := fmt.Sprintf("shopid = ? AND taskid = ? %s", searchExpr)
	count, err := repo.pst.Count(ctx, &models.StockBalanceImportDoc{}, exprCount, countArgs...)

	if err != nil {
		return results, paginationData, err
	}

	paginationData.PerPage = int64(pageable.Limit)
	paginationData.Page = int64(pageable.Page)
	paginationData.Total = int64(count)

	paginationData.Build()
	return results, paginationData, nil
}

func (repo StockBalanceImportClickHouseRepository) Create(ctx context.Context, doc models.StockBalanceImportDoc) error {
	return repo.pst.Create(ctx, &doc)
}

func (repo StockBalanceImportClickHouseRepository) CreateInBatch(ctx context.Context, docs []models.StockBalanceImportDoc) error {
	tempDocs := make([]interface{}, len(docs))
	for i := range docs {
		tempDocs[i] = &docs[i]
	}

	return repo.pst.CreateInBatch(ctx, tempDocs)
}

func (repo StockBalanceImportClickHouseRepository) Update(ctx context.Context, shopID string, guid string, doc models.StockBalanceImportRaw) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE stockbalanceimport UPDATE barcode = ?, name = ?, unitcode = ?, qty = ?, price = ? , sumamount = ? WHERE shopid = ? AND guidfixed = ?",
		doc.Barcode, doc.Name, doc.UnitCode, doc.Qty, doc.Price, doc.SumAmount, shopID, guid)
}

func (repo StockBalanceImportClickHouseRepository) DeleteByGUID(ctx context.Context, shopID string, guid string) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE stockbalanceimport DELETE WHERE shopid = ? AND guidfixed = ?",
		shopID, guid)
}

func (repo StockBalanceImportClickHouseRepository) DeleteByTaskID(ctx context.Context, shopID string, taskID string) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE stockbalanceimport DELETE WHERE shopid = ? AND taskid = ?",
		shopID, taskID)
}

func (repo StockBalanceImportClickHouseRepository) UpdateExist(ctx context.Context, shopID string, taskID string, isExist bool, barcodes []string) error {

	isNotExist := !isExist
	return repo.pst.Exec(ctx,
		"ALTER TABLE stockbalanceimport UPDATE isnotexist = ? WHERE shopid = ? AND taskid = ? AND barcode IN (?)", isNotExist, shopID, taskID, barcodes)
}

func (repo StockBalanceImportClickHouseRepository) CountExist(ctx context.Context, shopID string, taskID string, isExist bool) (int, error) {

	isNotExist := !isExist

	countArgs := []interface{}{}
	countArgs = append(countArgs, shopID, taskID, isNotExist)

	exprCount := "shopid = ? AND taskid = ? AND isnotexist = ?"
	count, err := repo.pst.Count(ctx, &models.StockBalanceImportDoc{}, exprCount, countArgs...)

	if err != nil {
		return 0, err
	}

	return count, nil
}
