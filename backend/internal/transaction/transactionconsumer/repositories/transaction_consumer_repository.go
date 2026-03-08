package repositories

import (
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm/clause"
)

type ITransactionConsumerRepository[T any] interface {
	Get(shopID string, docNo string) (*T, error)
	Create(doc T) error
	Update(shopID string, docNo string, doc T) error
	Delete(shopID string, docNo string, doc T) error
}

type TransactionConsumerRepository[T any] struct {
	pst microservice.IPersister
}

func NewTransactionConsumerRepository[T any](pst microservice.IPersister) ITransactionConsumerRepository[T] {
	return &TransactionConsumerRepository[T]{
		pst: pst,
	}
}

func (repo *TransactionConsumerRepository[T]) Get(shopID string, docNo string) (*T, error) {

	var data T
	err := repo.pst.DBClient().Preload(clause.Associations).
		Where("shopid=? AND docno=?", shopID, docNo).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repo *TransactionConsumerRepository[T]) Create(doc T) error {
	err := repo.pst.Create(doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *TransactionConsumerRepository[T]) Update(shopID string, docNo string, doc T) error {
	err := repo.pst.Update(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	})

	if err != nil {
		return err
	}
	return nil
}

func (repo *TransactionConsumerRepository[T]) Delete(shopID string, docNo string, doc T) error {

	tx := repo.pst.DBClient().Begin()

	err := tx.Delete(&doc, map[string]interface{}{
		"shopid": shopID,
		"docno":  docNo,
	}).Error

	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
