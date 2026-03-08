package saleinvoicereturn

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type ISaleInvoiceReturnTransactionAdminHttp interface {
	ReSyncSaleInvoiceReturnTransaction(ctx microservice.IContext) error
	ReSyncSaleInvoiceReturnDeleteTransaction(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type SaleInvoiceReturnTransactionAdminHttp struct {
	svc ISaleInvoiceReturnTransactionAdminService
}

func NewSaleInvoiceReturnTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) ISaleInvoiceReturnTransactionAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewSaleInvoiceReturnTransactionAdminService(mongoPersister, producer)

	return &SaleInvoiceReturnTransactionAdminHttp{
		svc: svc,
	}
}

func (s *SaleInvoiceReturnTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/saleinvoicereturn/resynctransaction", s.ReSyncSaleInvoiceReturnTransaction)
	ms.POST(prefix+"/transactionadmin/saleinvoicereturn/resyncdeletetransaction", s.ReSyncSaleInvoiceReturnDeleteTransaction)
}

func (s *SaleInvoiceReturnTransactionAdminHttp) ReSyncSaleInvoiceReturnTransaction(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncSaleInvoiceReturnDoc(req.ShopID)
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

func (s *SaleInvoiceReturnTransactionAdminHttp) ReSyncSaleInvoiceReturnDeleteTransaction(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncSaleInvoiceReturnDeleteDoc(req.ShopID)
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
