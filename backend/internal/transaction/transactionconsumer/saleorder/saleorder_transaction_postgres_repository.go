package saleorder

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type ISaleOrderTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.SaleOrderPG, error)
	Create(doc models.SaleOrderPG) error
	Update(holdingCode string, docNo string, doc models.SaleOrderPG) error
	Delete(holdingCode string, docNo string, doc models.SaleOrderPG) error
	DeleteData(holdingCode string, docNo string, doc models.SaleOrderPG) error
}

type SaleOrderTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.SaleOrderPG]
}

func NewSaleOrderTransactionPGRepository(pst microservice.IPersister) ISaleOrderTransactionPGRepository {

	repo := &SaleOrderTransactionPGRepository{
		pst: pst,
	}
	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.SaleOrderPG](pst)
	return repo
}

func (repo SaleOrderTransactionPGRepository) Create(doc models.SaleOrderPG) error {
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo SaleOrderTransactionPGRepository) Update(holdingCode string, docNo string, doc models.SaleOrderPG) error {

	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *SaleOrderTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.SaleOrderPG) error {

	var details *[]models.SaleOrderDetailPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.SaleOrderDetailPG{}).Where(" holding_code=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		// mark delete
		err := tx.Delete(&models.SaleOrderDetailPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(&models.SaleOrderPG{}, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
