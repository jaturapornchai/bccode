package repositories

import (
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type ISaleInvoiceBomPricePostgresRepository interface {
	Get(shopID string, docNo string) (*models.SaleInvoiceBomPricePg, error)
	Create(doc models.SaleInvoiceBomPricePg) error
	Update(shopID string, docNo string, doc models.SaleInvoiceBomPricePg) error
	Delete(shopID string, docNo string) error
}

type SaleInvoiceBomPricePostgresRepository struct {
	pst microservice.IPersister
}

func NewSaleInvoiceBomPricePostgresRepository(pst microservice.IPersister) ISaleInvoiceBomPricePostgresRepository {
	return &SaleInvoiceBomPricePostgresRepository{
		pst: pst,
	}
}

func (repo *SaleInvoiceBomPricePostgresRepository) Get(shopID string, docNo string) (*models.SaleInvoiceBomPricePg, error) {
	var result models.SaleInvoiceBomPricePg
	_, err := repo.pst.First(&result, "shopid=? AND docno=?", shopID, docNo)
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

func (repo *SaleInvoiceBomPricePostgresRepository) Update(shopID string, docNo string, doc models.SaleInvoiceBomPricePg) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *SaleInvoiceBomPricePostgresRepository) Delete(shopID string, docNo string) error {
	err := repo.pst.Delete(&models.SaleInvoiceBomPricePg{}, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}
