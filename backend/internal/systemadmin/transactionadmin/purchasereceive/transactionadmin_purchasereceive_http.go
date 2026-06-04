package purchasereceive

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseReceiveTransactionAdminHttp interface {
	RegisterHttp(ms *microservice.Microservice, prefix string)
	ReSyncPurchaseReceiveTransaction(ms microservice.IContext) error
	ReSyncPurchaseReceiveDeleteTransaction(ms microservice.IContext) error
}

type PurchaseReceiveTransactionAdminHttp struct {
	svc IPurchaseReceiveTransactionAdminService
}

func NewPurchaseReceiveTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IPurchaseReceiveTransactionAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewPurchaseReceiveTransactionAdminService(mongoPersister, producer)

	return &PurchaseReceiveTransactionAdminHttp{
		svc: svc,
	}
}

func (s *PurchaseReceiveTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/purchasereceive/resynctransaction", s.ReSyncPurchaseReceiveTransaction)
	ms.POST(prefix+"/transactionadmin/purchasereceive/resyncdeletetransaction", s.ReSyncPurchaseReceiveDeleteTransaction)
}

func (s *PurchaseReceiveTransactionAdminHttp) ReSyncPurchaseReceiveTransaction(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncPurchaseReceiveDoc(req.HoldingCode)
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

func (s *PurchaseReceiveTransactionAdminHttp) ReSyncPurchaseReceiveDeleteTransaction(ctx microservice.IContext) error {
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

	err = s.svc.ReSyncPurchaseReceiveDeleteDoc(req.HoldingCode)
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
