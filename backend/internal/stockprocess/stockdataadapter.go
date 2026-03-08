package stockprocess

import "smlcloudplatform/pkg/microservice"

type IStockDataAdapter interface {
	GetStockDataLength(shopID string, barcode string) int64
}

type StockDataAdapter struct {
	pstRedis microservice.ICacher
	pstPg    microservice.IPersister
}

func newStockDataAdapter(pstRedis microservice.ICacher, pstPg microservice.IPersister) *StockDataAdapter {
	return &StockDataAdapter{
		pstRedis: pstRedis,
		pstPg:    pstPg,
	}
}
