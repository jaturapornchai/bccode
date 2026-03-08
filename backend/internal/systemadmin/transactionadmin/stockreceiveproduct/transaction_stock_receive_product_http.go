package stockreceiveproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockReceiveTransactionAdminHttp interface {
	ReSyncStockReceiveProductTransaction(ctx microservice.IContext) error
	ReSyncStockReceiveProductDeleteTransaction(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type StockReceiveTransactionAdminHttp struct {
	svc IStockReceiveProductTransactionAdminService
}

func NewStockReceiveTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IStockReceiveTransactionAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewStockReceiveProductTransactionAdminService(mongoPersister, producer)
	return &StockReceiveTransactionAdminHttp{
		svc: svc,
	}
}

func (s *StockReceiveTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/stockreceiveproduct/resynctransaction", s.ReSyncStockReceiveProductTransaction)
	ms.POST(prefix+"/transactionadmin/stockreceiveproduct/resyncdeletetransaction", s.ReSyncStockReceiveProductDeleteTransaction)
}

func (s *StockReceiveTransactionAdminHttp) ReSyncStockReceiveProductTransaction(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncStockReceiveProductDoc(req.ShopID)
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

func (s *StockReceiveTransactionAdminHttp) ReSyncStockReceiveProductDeleteTransaction(ctx microservice.IContext) error {
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

	err = s.svc.ReSyncStockReceiveProductDeleteDoc(req.ShopID)
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
