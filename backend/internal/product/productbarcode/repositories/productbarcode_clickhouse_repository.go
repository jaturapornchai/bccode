package repositories

import (
	"context"
	"fmt"
	"math"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

type IProductBarcodeClickhouseRepository interface {
	Search(shopID string, pageable micromodels.Pageable) ([]models.ProductBarcodeSearch, common.Pagination, error)
}

type ProductBarcodeClickhouseRepository struct {
	pst microservice.IPersisterClickHouse
}

func NewProductBarcodeClickhouseRepository(pst microservice.IPersisterClickHouse) *ProductBarcodeClickhouseRepository {

	if pst == nil {
		return nil
	}

	insRepo := &ProductBarcodeClickhouseRepository{
		pst: pst,
	}
	return insRepo
}

func (repo ProductBarcodeClickhouseRepository) Search(shopID string, pageable micromodels.Pageable) ([]models.ProductBarcodeSearch, common.Pagination, error) {

	searchInFields := []string{"iccode", "barcode", "unitcode"}

	conn := repo.pst.Conn()

	where := "WHERE shopid = ? "

	whereSerach := ""

	if len(pageable.Query) > 0 {
		for _, field := range searchInFields {
			if whereSerach != "" {
				whereSerach += " OR "
			}
			whereSerach += fmt.Sprintf(" %s LIKE '%%%s%%' ", field, pageable.Query)
		}

		if whereSerach != "" {
			whereSerach += " OR "
		}

		whereSerach += fmt.Sprintf(" arrayJoin(names) LIKE '%%%s%%' ", pageable.Query)
	}

	if whereSerach != "" {
		where += fmt.Sprintf("AND (%s) ", whereSerach)
	}

	offset := pageable.GetOffest()

	results := []models.ProductBarcodeSearch{}
	err := conn.Select(
		context.Background(),
		&results,
		fmt.Sprintf("SELECT iccode, barcode, unitcode, price, names FROM productbarcode %s LIMIT ? OFFSET ?", where),
		shopID,
		uint64(pageable.Limit),
		uint64(offset),
	)

	if err != nil {
		return []models.ProductBarcodeSearch{}, common.Pagination{}, err
	}

	var count uint64

	err = conn.QueryRow(context.Background(), fmt.Sprintf("SELECT count(*) FROM productbarcode %s LIMIT 1", where), shopID).Scan(&count)

	if err != nil {
		return []models.ProductBarcodeSearch{}, common.Pagination{}, err
	}

	totalPage := math.Ceil(float64(count) / float64(pageable.Limit))

	pagination := common.Pagination{
		Total:     int(count),
		Page:      pageable.Page,
		PerPage:   pageable.Limit,
		TotalPage: int(totalPage),
	}

	return results, pagination, err
}
