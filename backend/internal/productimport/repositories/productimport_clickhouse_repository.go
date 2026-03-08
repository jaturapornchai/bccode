package repositories

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/productimport/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/fatih/structs"
)

type IProductImportClickHouseRepository interface {
	All(ctx context.Context, shopID string, taskID string) ([]models.ProductImportDoc, error)
	List(ctx context.Context, shopID string, taskID string, pageable micromodels.Pageable) ([]models.ProductImportDoc, models.PaginationData, error)
	Create(ctx context.Context, doc models.ProductImportDoc) error
	CreateInBatch(ctx context.Context, docs []models.ProductImportDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ProductImportRaw) error
	DeleteByGUID(ctx context.Context, shopID string, guid string) error
	DeleteByTaskID(ctx context.Context, shopID string, taskID string) error
	FindOne(ctx context.Context, shopID string, taskID string, sorts []micromodels.KeyInt) (models.ProductImportDoc, error)

	UpdateDuplicate(ctx context.Context, shopID string, taskID string, isDuplicate bool, barcodes []string) error
	UpdateExist(ctx context.Context, shopID string, taskID string, isExist bool, barcodes []string) error

	CountExist(ctx context.Context, shopID string, taskID string, isExist bool) (int, error)
	CountDuplicate(ctx context.Context, shopID string, taskID string, isDuplicate bool) (int, error)

	CountUnitExist(ctx context.Context, shopID string, taskID string, isExist bool) (int, error)
	UpdateUnitExist(ctx context.Context, shopID string, taskID string, isExist bool, unitCodes []string) error
}

type ProductImportClickHouseRepository struct {
	pst          microservice.IPersisterClickHouse
	structFileds map[string]struct{}
}

func NewProductImportClickHouseRepository(pst microservice.IPersisterClickHouse) ProductImportClickHouseRepository {

	structFileds := make(map[string]struct{})
	fields := structs.Fields(models.ProductImport{})

	for _, field := range fields {
		tag := field.Tag("ch")
		structFileds[tag] = struct{}{}
	}

	return ProductImportClickHouseRepository{
		pst: pst,
	}
}

func (repo ProductImportClickHouseRepository) All(ctx context.Context, shopID string, taskID string) ([]models.ProductImportDoc, error) {

	results := []models.ProductImportDoc{}

	sqlExpr := "SELECT * FROM productbarcodeimport WHERE shopid = ? AND taskid = ?"
	err := repo.pst.Select(ctx, &results, sqlExpr, shopID, taskID)

	if err != nil {
		return results, err
	}

	return results, nil
}

func (repo ProductImportClickHouseRepository) FindOne(ctx context.Context, shopID string, taskID string, sorts []micromodels.KeyInt) (models.ProductImportDoc, error) {

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

	results := []models.ProductImportDoc{}

	sqlExpr := fmt.Sprintf("SELECT * FROM productbarcodeimport WHERE shopid = ? AND taskid = ? %s LIMIT 1 OFFSET 0", orderExpr)
	err := repo.pst.Select(ctx, &results, sqlExpr, shopID, taskID)

	if err != nil {
		return models.ProductImportDoc{}, err
	}

	if len(results) == 0 {
		return models.ProductImportDoc{}, nil
	}

	return results[0], nil
}

func (repo ProductImportClickHouseRepository) List(ctx context.Context, shopID string, taskID string, pageable micromodels.Pageable) ([]models.ProductImportDoc, models.PaginationData, error) {

	offset := (pageable.Page - 1) * pageable.Limit

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

			if _, ok := repo.structFileds[sort.Key]; !ok {
				orderExpr += fmt.Sprintf("%s %s", sort.Key, orderTxt)
			}

		}
	}

	if orderExpr != "" {
		orderExpr = fmt.Sprintf("ORDER BY %s", orderExpr)
	}

	exprSeach := ""
	searchArgs := []interface{}{}
	if pageable.Query != "" {
		exprSeach = "AND (barcode LIKE ? OR name LIKE ? OR unitcode LIKE ?)"
		searchTxt := fmt.Sprintf("%s%%", pageable.Query)
		searchArgs = append(searchArgs, searchTxt, searchTxt, searchTxt)
	}

	paginationData := models.PaginationData{}
	results := []models.ProductImportDoc{}

	args := []interface{}{}
	args = append(args, shopID, taskID)
	args = append(args, searchArgs...)
	args = append(args, pageable.Limit, offset)

	sqlExpr := fmt.Sprintf("SELECT * FROM productbarcodeimport WHERE shopid = ? AND taskid = ? %s %s LIMIT ? OFFSET ?", exprSeach, orderExpr)
	err := repo.pst.Select(ctx, &results, sqlExpr, args...)

	if err != nil {
		return results, paginationData, err
	}

	countArgs := []interface{}{}
	countArgs = append(countArgs, shopID, taskID)
	countArgs = append(countArgs, searchArgs...)

	exprCount := fmt.Sprintf("shopid = ? AND taskid = ? %s", exprSeach)
	count, err := repo.pst.Count(ctx, &models.ProductImportDoc{}, exprCount, countArgs...)

	if err != nil {
		return results, paginationData, err
	}

	paginationData.PerPage = int64(pageable.Limit)
	paginationData.Page = int64(pageable.Page)
	paginationData.Total = int64(count)

	paginationData.Build()
	return results, paginationData, nil
}

func (repo ProductImportClickHouseRepository) Create(ctx context.Context, doc models.ProductImportDoc) error {
	return repo.pst.Create(ctx, &doc)
}

func (repo ProductImportClickHouseRepository) CreateInBatch(ctx context.Context, docs []models.ProductImportDoc) error {
	tempDocs := make([]interface{}, len(docs))
	for i := range docs {
		tempDocs[i] = &docs[i]
	}

	return repo.pst.CreateInBatch(ctx, tempDocs)
}

func (repo ProductImportClickHouseRepository) Update(ctx context.Context, shopID string, guid string, doc models.ProductImportRaw) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE productbarcodeimport UPDATE barcode = ?, name = ?, unitcode = ?,  price = ? , pricemember = ? , code = ? WHERE shopid = ? AND guidfixed = ?",
		doc.Barcode, doc.Name, doc.UnitCode, doc.Price, doc.PriceMember, doc.Code, shopID, guid)
}

func (repo ProductImportClickHouseRepository) DeleteByGUID(ctx context.Context, shopID string, guid string) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE productbarcodeimport DELETE WHERE shopid = ? AND guidfixed = ?",
		shopID, guid)
}

func (repo ProductImportClickHouseRepository) DeleteByTaskID(ctx context.Context, shopID string, taskID string) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE productbarcodeimport DELETE WHERE shopid = ? AND taskid = ?",
		shopID, taskID)
}

func (repo ProductImportClickHouseRepository) UpdateDuplicate(ctx context.Context, shopID string, taskID string, isDuplicate bool, barcodes []string) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE productbarcodeimport UPDATE isduplicate = ? WHERE shopid = ? AND taskid = ? AND barcode IN (?)", isDuplicate, shopID, taskID, barcodes)
}

func (repo ProductImportClickHouseRepository) UpdateExist(ctx context.Context, shopID string, taskID string, isExist bool, barcodes []string) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE productbarcodeimport UPDATE isexist = ? WHERE shopid = ? AND taskid = ? AND barcode IN (?)", isExist, shopID, taskID, barcodes)
}

func (repo ProductImportClickHouseRepository) UpdateUnitExist(ctx context.Context, shopID string, taskID string, isExist bool, unitCodes []string) error {
	return repo.pst.Exec(ctx,
		"ALTER TABLE productbarcodeimport UPDATE isunitnotexist = ? WHERE shopid = ? AND taskid = ? AND unitcode IN (?)", isExist, shopID, taskID, unitCodes)
}

func (repo ProductImportClickHouseRepository) CountExist(ctx context.Context, shopID string, taskID string, isExist bool) (int, error) {

	countArgs := []interface{}{}
	countArgs = append(countArgs, shopID, taskID, isExist)

	exprCount := "shopid = ? AND taskid = ? AND isexist = ?"
	count, err := repo.pst.Count(ctx, &models.ProductImportDoc{}, exprCount, countArgs...)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo ProductImportClickHouseRepository) CountDuplicate(ctx context.Context, shopID string, taskID string, isDuplicate bool) (int, error) {

	countArgs := []interface{}{}
	countArgs = append(countArgs, shopID, taskID, isDuplicate)

	exprCount := "shopid = ? AND taskid = ? AND isduplicate = ?"
	count, err := repo.pst.Count(ctx, &models.ProductImportDoc{}, exprCount, countArgs...)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo ProductImportClickHouseRepository) CountUnitExist(ctx context.Context, shopID string, taskID string, isNotExist bool) (int, error) {

	countArgs := []interface{}{}
	countArgs = append(countArgs, shopID, taskID, isNotExist)

	exprCount := "shopid = ? AND taskid = ? AND isunitnotexist = ?"
	count, err := repo.pst.Count(ctx, &models.ProductImportDoc{}, exprCount, countArgs...)

	if err != nil {
		return 0, err
	}

	return count, nil
}
