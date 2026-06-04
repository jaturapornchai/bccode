package debtorpayment

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorPaymentTransactionAdminHttp interface {
	RegisterHttp(ms *microservice.Microservice, prefix string)
	ReSyncCreditorPaymentTransaction(ms microservice.IContext) error
}

type DebtorPaymentTransactionAdminHttp struct {
	svc IDebtorPaymentTransactionAdminService
}

func NewDebtorPaymentTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IDebtorPaymentTransactionAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewDebtorPaymentTransactionAdminService(mongoPersister, producer)

	return &DebtorPaymentTransactionAdminHttp{
		svc: svc,
	}
}

func (s *DebtorPaymentTransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/transactionadmin/debtorpayment/resynctransaction", s.ReSyncCreditorPaymentTransaction)
}

func (s *DebtorPaymentTransactionAdminHttp) ReSyncCreditorPaymentTransaction(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncDebtorPaymentDoc(req.HoldingCode)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	return nil
}
