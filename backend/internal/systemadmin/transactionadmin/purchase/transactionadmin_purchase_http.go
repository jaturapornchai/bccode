package purchase

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseTransactionAdminHttp interface {
	RegisterHttp(ms *microservice.Microservice, prefix string)
	ReSyncPurchaseTransaction(ms microservice.IContext) error
	ReSyncPurchaseDeleteTransaction(ms microservice.IContext) error
}

type PurchaseTransactionAdminHttp struct {
	svc IPurchaseTransactionAdminService
}

func NewPurchaseTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IPurchaseTransactionAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewPurchaseTransactionAdminService(mongoPersister, producer)

	return &PurchaseTransactionAdminHttp{
		svc: svc,
	}
}

func (s *PurchaseTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/purchase/resynctransaction", s.ReSyncPurchaseTransaction)
	ms.POST(prefix+"/transactionadmin/purchase/resyncdeletetransaction", s.ReSyncPurchaseDeleteTransaction)
}

func (s *PurchaseTransactionAdminHttp) ReSyncPurchaseTransaction(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncPurchaseDoc(req.HoldingCode)
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

func (s *PurchaseTransactionAdminHttp) ReSyncPurchaseDeleteTransaction(ctx microservice.IContext) error {
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

	err = s.svc.ReSyncPurchaseDeleteDoc(req.HoldingCode)
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
