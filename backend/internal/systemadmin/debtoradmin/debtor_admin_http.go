package debtoradmin

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorAdminHttp interface {
	ReSyncDebtor(ctx microservice.IContext) error
	ReCalcDebtorBalance(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type DebtorAdminHttp struct {
	svc IDebtorAdminService
}

func NewDebtorAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IDebtorAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewDebtorAdminService(mongoPersister, producer)

	return &DebtorAdminHttp{
		svc: svc,
	}
}

func (s *DebtorAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/debtoradmin/resyncdebtor", s.ReSyncDebtor)
	ms.POST(prefix+"/debtoradmin/recalcbalance", s.ReCalcDebtorBalance)
}

func (s *DebtorAdminHttp) ReSyncDebtor(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncDebtor(req.HoldingCode)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	return nil
}

func (s *DebtorAdminHttp) ReCalcDebtorBalance(ctx microservice.IContext) error {

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

	err = s.svc.ReCalcDebtorBalance(req.HoldingCode)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	return nil
}
