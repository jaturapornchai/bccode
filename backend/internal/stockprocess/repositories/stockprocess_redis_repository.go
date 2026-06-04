package repositories

import (
	stockProcessModels "smlcloudplatform/internal/stockprocess/models"
	"smlcloudplatform/pkg/microservice"
)

const STOCK_PROCESS_KEY = "STKPROCESS"

type IStockProcessRedisRepository interface {
	AddStockData(holdingCode string, barcode string, stockData stockProcessModels.StockData) error
	BulkAddStockData(holdingCode string, barcode string, stockDatas []stockProcessModels.StockData) error
	GetStockDataLength(holdingCode string, barcode string) (int64, error)
	// GetStockDataLine(holdingCode string, barcode string, line int64) (stockProcessModels.StockData, error)
	// PutStockData(holdingCode string, barcode string, line int64, stockData stockProcessModels.StockData) error
	// GetStockDataList(holdingCode string, barcode string) ([]stockProcessModels.StockData, error)
	FindStockMovement(holdingCode string, barcode string, stockData stockProcessModels.StockData) (int64, error)
}

type StockProcessRedisRepository struct {
	pst microservice.ICacher
}

func NewStockProcessRedisRepository(pst microservice.ICacher) IStockProcessRedisRepository {
	return &StockProcessRedisRepository{
		pst: pst,
	}
}

func (repo StockProcessRedisRepository) AddStockData(holdingCode string, barcode string, stockData stockProcessModels.StockData) error {
	redisKey := STOCK_PROCESS_KEY + "::" + holdingCode + "::" + barcode

	err := repo.pst.RPush(redisKey, stockData)
	if err != nil {
		return err
	}

	return nil
}

func (repo StockProcessRedisRepository) BulkAddStockData(holdingCode string, barcode string, stockDatas []stockProcessModels.StockData) error {

	redisKey := STOCK_PROCESS_KEY + "::" + holdingCode + "::" + barcode

	for _, data := range stockDatas {
		err := repo.pst.RPush(redisKey, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo StockProcessRedisRepository) GetStockDataLength(holdingCode string, barcode string) (int64, error) {

	redisKey := STOCK_PROCESS_KEY + "::" + holdingCode + "::" + barcode

	size, err := repo.pst.LLen(redisKey)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (repo StockProcessRedisRepository) FindStockMovement(holdingCode string, barcode string, stockData stockProcessModels.StockData) (int64, error) {

	redisKey := STOCK_PROCESS_KEY + "::" + holdingCode + "::" + barcode

	pos, err := repo.pst.LPos(redisKey, stockData)
	if err != nil {
		return 0, err
	}

	return pos, nil
}
