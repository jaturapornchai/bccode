package stocktransfer

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockTransferTransactionAdminHttp interface {
	ReSyncStockTransferTransaction(ctx microservice.IContext) error
	ReSyncStockTransferDeleteTransaction(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type StockTransferTransactionAdminHttp struct {
	svc IStockTransferTransactionAdminService
}

func NewStockTransferTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IStockTransferTransactionAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewStockTransferTransactionAdminService(mongoPersister, producer)

	return &StockTransferTransactionAdminHttp{
		svc: svc,
	}
}

func (s *StockTransferTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/stocktransfer/resynctransaction", s.ReSyncStockTransferTransaction)
	ms.POST(prefix+"/transactionadmin/stocktransfer/resyncdeletetransaction", s.ReSyncStockTransferDeleteTransaction)
}

func (s *StockTransferTransactionAdminHttp) ReSyncStockTransferTransaction(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncStockTransferDoc(req.ShopID)
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

func (s *StockTransferTransactionAdminHttp) ReSyncStockTransferDeleteTransaction(ctx microservice.IContext) error {
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

	err = s.svc.ReSyncStockTransferDeleteDoc(req.ShopID)
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
