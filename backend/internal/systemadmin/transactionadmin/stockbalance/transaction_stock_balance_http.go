package stockbalance

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockBalanceTransactionAdminHttp interface {
	ReSyncStockBalanceProductTransaction(ctx microservice.IContext) error
	ReSyncStockBalanceProductDeleteTransaction(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type StockBalanceTransactionAdminHttp struct {
	svc IStockBalanceProductTransactionAdminService
}

func NewStockBalanceTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IStockBalanceTransactionAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewStockBalanceProductTransactionAdminService(mongoPersister, producer)
	return &StockBalanceTransactionAdminHttp{
		svc: svc,
	}
}

func (s *StockBalanceTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/stockbalance/resynctransaction", s.ReSyncStockBalanceProductTransaction)
	ms.POST(prefix+"/transactionadmin/stockbalance/resyncdeletetransaction", s.ReSyncStockBalanceProductDeleteTransaction)
}

func (s *StockBalanceTransactionAdminHttp) ReSyncStockBalanceProductTransaction(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	var req adminModels.RequestReSyncTenant

	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	err = s.svc.ReSyncStockBalanceProductDoc(req.ShopID)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	return nil
}

func (s *StockBalanceTransactionAdminHttp) ReSyncStockBalanceProductDeleteTransaction(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	var req adminModels.RequestReSyncTenant

	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	err = s.svc.ReSyncStockBalanceProductDeleteDoc(req.ShopID)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	return nil
}
