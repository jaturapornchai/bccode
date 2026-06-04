package rfq

import (
	"errors"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/repositories"
	"smlcloudplatform/pkg/microservice"
)

type IRFQTransactionPGRepository interface {
	Get(holdingCode string, docNo string) (*models.RFQTransactionPG, error)
	Create(doc models.RFQTransactionPG) error
	Update(holdingCode string, docNo string, doc models.RFQTransactionPG) error
	DeleteData(holdingCode string, docNo string, doc models.RFQTransactionPG) error
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

func (repo RFQTransactionPGRepository) Update(holdingCode string, docNo string, doc models.RFQTransactionPG) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"holding_code": holdingCode,
		"docno":        docNo,
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *RFQTransactionPGRepository) DeleteData(holdingCode string, docNo string, doc models.RFQTransactionPG) error {
	var details *[]models.RFQDetailTransactionPG
	tx := repo.pst.DBClient().Begin()

	tx.Model(&models.RFQDetailTransactionPG{}).Where(" holding_code=? AND docno=?", holdingCode, docNo).Find(&details)
	for _, tmp := range *details {
		err := tx.Delete(&models.RFQDetailTransactionPG{}, tmp.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err := tx.Delete(models.RFQTransactionPG{}, map[string]interface{}{
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
