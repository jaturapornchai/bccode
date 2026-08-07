package repositories

import (
	"fmt"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IProductBarcodePGRepository interface {
	FindByBarcode(holdingCode string, barcode string) (*models.ProductBarcodePg, error)
	FindByBarcodes(holdingCode string, barcodes []string) ([]models.ProductBarcodePg, error)
	Get(holdingCode string, barcode string) (*models.ProductBarcodePg, error)
	GetByKey(holdingCode string, itemCode string, barcode string) (*models.ProductBarcodePg, error)
	Create(doc *models.ProductBarcodePg) error
	Update(holdingCode string, itemCode string, barcode string, doc *models.ProductBarcodePg) error
	Delete(holdingCode string, itemCode string, barcode string) error
}

// ICompanyProductBarcodePGRepository is the metadata projection contract.
// It stays separate from the legacy stock repository contract so stock code
// cannot accidentally resolve a duplicate barcode across companies.
type ICompanyProductBarcodePGRepository interface {
	GetByCompanyBarcode(holdingCode string, businessCode string, barcode string) (*models.ProductBarcodePg, error)
	Create(doc *models.ProductBarcodePg) error
	UpdateInCompany(holdingCode string, businessCode string, barcode string, doc *models.ProductBarcodePg) error
	DeleteInCompany(holdingCode string, businessCode string, barcode string) error
}

type ProductBarcodePGRepository struct {
	pst microservice.IPersister
}

func NewProductBarcodePGRepository(pst microservice.IPersister) *ProductBarcodePGRepository {
	return &ProductBarcodePGRepository{
		pst: pst,
	}
}

func (repo *ProductBarcodePGRepository) Get(holdingCode string, barcode string) (*models.ProductBarcodePg, error) {
	return repo.FindByBarcode(holdingCode, barcode)
}

func (repo *ProductBarcodePGRepository) GetByKey(holdingCode string, itemCode string, barcode string) (*models.ProductBarcodePg, error) {
	var result models.ProductBarcodePg
	_, err := repo.pst.First(&result, "holding_code=? AND itemcode=? AND barcode=?", holdingCode, itemCode, barcode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (repo *ProductBarcodePGRepository) FindByBarcode(holdingCode string, barcode string) (*models.ProductBarcodePg, error) {
	var results []models.ProductBarcodePg
	_, err := repo.pst.Where(&results, "holding_code=? AND barcode = ?", holdingCode, barcode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if len(results) == 0 {
		return nil, nil
	}
	if len(results) > 1 {
		return nil, fmt.Errorf("barcode %s matches multiple products; itemcode is required", barcode)
	}
	return &results[0], nil
}

func (repo *ProductBarcodePGRepository) FindByBarcodes(holdingCode string, barcodes []string) ([]models.ProductBarcodePg, error) {
	var results []models.ProductBarcodePg
	_, err := repo.pst.Where(&results, "holding_code=? AND barcode IN ?", holdingCode, barcodes)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return results, nil
}

func (repo *ProductBarcodePGRepository) Create(doc *models.ProductBarcodePg) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *ProductBarcodePGRepository) Update(holdingCode string, itemCode string, barcode string, doc *models.ProductBarcodePg) error {
	err := repo.pst.Update(doc, map[string]interface{}{
		"holding_code": holdingCode,
		"itemcode":     itemCode,
		"barcode":      barcode,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *ProductBarcodePGRepository) Delete(holdingCode string, itemCode string, barcode string) error {

	err := repo.pst.Delete(&models.ProductBarcodePg{}, map[string]interface{}{
		"holding_code": holdingCode,
		"itemcode":     itemCode,
		"barcode":      barcode,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *ProductBarcodePGRepository) GetByCompanyBarcode(holdingCode string, businessCode string, barcode string) (*models.ProductBarcodePg, error) {
	var result models.ProductBarcodePg
	_, err := repo.pst.First(
		&result,
		"holding_code=? AND businesscode=? AND barcode=?",
		holdingCode,
		businessCode,
		barcode,
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (repo *ProductBarcodePGRepository) UpdateInCompany(holdingCode string, businessCode string, barcode string, doc *models.ProductBarcodePg) error {
	return repo.pst.Update(doc, map[string]interface{}{
		"holding_code": holdingCode,
		"businesscode": businessCode,
		"barcode":      barcode,
	})
}

func (repo *ProductBarcodePGRepository) DeleteInCompany(holdingCode string, businessCode string, barcode string) error {
	return repo.pst.Delete(&models.ProductBarcodePg{}, map[string]interface{}{
		"holding_code": holdingCode,
		"businesscode": businessCode,
		"barcode":      barcode,
	})
}
