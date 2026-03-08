package stockpickupproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockPickupTransactionAdminHttp interface {
	ReSyncStockPickupTransaction(ctx microservice.IContext) error
	ReSyncStockPickupDeleteTransaction(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type StockPickupTransactionAdminHttp struct {
	svc IStockPickupTransactionAdminService
}

func NewStockPickupTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IStockPickupTransactionAdminHttp {
	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewStockPickupTransactionAdminService(mongoPersister, producer)
	return &StockPickupTransactionAdminHttp{
		svc: svc,
	}
}

func (s *StockPickupTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/stockpickup/resynctransaction", s.ReSyncStockPickupTransaction)
	ms.POST(prefix+"/transactionadmin/stockpickup/resyncdeletetransaction", s.ReSyncStockPickupDeleteTransaction)
}

func (s *StockPickupTransactionAdminHttp) ReSyncStockPickupTransaction(ctx microservice.IContext) error {
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

	err = s.svc.ReSyncStockPickupTransaction(req.ShopID)
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

func (s *StockPickupTransactionAdminHttp) ReSyncStockPickupDeleteTransaction(ctx microservice.IContext) error {
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

	err = s.svc.ReSyncStockPickupDeleteTransaction(req.ShopID)
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
