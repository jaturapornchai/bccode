package stockadjustment

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockAdjustmentTransactionAdminHttp interface {
	ReSyncStockAdjustmentTransaction(ctx microservice.IContext) error
	ReSyncStockAdjustmentDeleteTransaction(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type StockAdjustmentTransactionAdminHttp struct {
	svc IStockAdjustmentTransactionAdminService
}

func NewStockAdjustmentTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IStockAdjustmentTransactionAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewStockAdjustmentTransactionAdminService(mongoPersister, producer)

	return &StockAdjustmentTransactionAdminHttp{
		svc: svc,
	}
}

func (s *StockAdjustmentTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/stockadjustment/resynctransaction", s.ReSyncStockAdjustmentTransaction)
	ms.POST(prefix+"/transactionadmin/stockadjustment/resyncdeletetransaction", s.ReSyncStockAdjustmentDeleteTransaction)
}

func (s *StockAdjustmentTransactionAdminHttp) ReSyncStockAdjustmentTransaction(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncStockAdjustmentDoc(req.HoldingCode)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "success",
	})

	return nil
}

func (s *StockAdjustmentTransactionAdminHttp) ReSyncStockAdjustmentDeleteTransaction(ctx microservice.IContext) error {
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

	err = s.svc.ReSyncStockAdjustmentDeleteDoc(req.HoldingCode)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "success",
	})

	return nil
}
