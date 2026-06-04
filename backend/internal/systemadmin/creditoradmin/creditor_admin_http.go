package creditoradmin

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type ICreditorAdminHttp interface {
	ReSyncCreditor(ctx microservice.IContext) error
	ReCalcCreditorBalance(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type CreditorAdminHttp struct {
	svc ICreditorAdminService
}

func NewCreditorAdminHttp(ms *microservice.Microservice, cfg config.IConfig) ICreditorAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewCreditorAdminService(mongoPersister, producer)

	return &CreditorAdminHttp{
		svc: svc,
	}
}

func (s *CreditorAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/creditoradmin/resynccreditor", s.ReSyncCreditor)
	ms.POST(prefix+"/creditoradmin/recalcbalance", s.ReCalcCreditorBalance)
}

func (s *CreditorAdminHttp) ReSyncCreditor(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncCreditor(req.HoldingCode)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	return nil
}

func (s *CreditorAdminHttp) ReCalcCreditorBalance(ctx microservice.IContext) error {

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

	err = s.svc.ReCalcCreditorBalance(req.HoldingCode)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	return nil
}
