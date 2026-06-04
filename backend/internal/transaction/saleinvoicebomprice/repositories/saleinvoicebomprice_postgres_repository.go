package repositories

import (
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type ISaleInvoiceBomPricePostgresRepository interface {
	Get(holdingCode string, docNo string) (*models.SaleInvoiceBomPricePg, error)
	Create(doc models.SaleInvoiceBomPricePg) error
	Update(holdingCode string, docNo string, doc models.SaleInvoiceBomPricePg) error
	Delete(holdingCode string, docNo string) error
}

type SaleInvoiceBomPricePostgresRepository struct {
	pst microservice.IPersister
}

func NewSaleInvoiceBomPricePostgresRepository(pst microservice.IPersister) ISaleInvoiceBomPricePostgresRepository {
	return &SaleInvoiceBomPricePostgresRepository{
		pst: pst,
	}
}

func (repo *SaleInvoiceBomPricePostgresRepository) Get(holdingCode string, docNo string) (*models.SaleInvoiceBomPricePg, error) {
	var result models.SaleInvoiceBomPricePg
	_, err := repo.pst.First(&result, "holding_code=? AND docno=?", holdingCode, docNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (repo *SaleInvoiceBomPricePostgresRepository) Create(doc models.SaleInvoiceBomPricePg) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *SaleInvoiceBomPricePostgresRepository) Update(holdingCode string, docNo string, doc models.SaleInvoiceBomPricePg) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *SaleInvoiceBomPricePostgresRepository) Delete(holdingCode string, docNo string) error {
	err := repo.pst.Delete(&models.SaleInvoiceBomPricePg{}, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	})

	if err != nil {
		return err
	}
	return nil
}
