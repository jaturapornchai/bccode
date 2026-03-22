package rfq

import (
	"errors"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IRFQTransactionPGRepository interface {
	Get(shopID string, docNo string) (*models.RFQTransactionPG, error)
	Create(doc models.RFQTransactionPG) error
	Update(shopID string, docNo string, doc models.RFQTransactionPG) error
	DeleteData(shopID string, docNo string, doc models.RFQTransactionPG) error
}

type RFQTransactionPGRepository struct {
	pst microservice.IPersister
	repositories.ITransactionConsumerRepository[models.RFQTransactionPG]
}

func NewRFQTransactionRepository(pst microservice.IPersister) IRFQTransactionPGRepository {
	repo := &RFQTransactionPGRepository{
		pst: pst,
	}
	repo.ITransactionConsumerRepository = repositories.NewTransactionConsumerRepository[models.RFQTransactionPG](pst)
	return repo
}

func (repo RFQTransactionPGRepository) Create(doc models.RFQTransactionPG) error {
	if doc.DocNo == "" {
		return errors.New("DocNo cannot be empty")
	}
	err := repo.pst.Create(&doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo RFQTransactionPGRepository) Update(shopID string, docNo string, doc models.RFQTransactionPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *RFQTransactionPGRepository) DeleteData(shopID string, docNo string, doc models.RFQTransactionPG) error {
	var details *[]models.RFQDetailTransactionPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.RFQDetailTransactionPG{}).Where(" shopid=? AND docno=?", shopID, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.RFQDetailTransactionPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(models.RFQTransactionPG{}, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
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
