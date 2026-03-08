package chartofaccountadmin

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	adminModels "smlcloudplatform/internal/systemadmin/models"
	"smlcloudplatform/pkg/microservice"
)

type IChartOfAccountAdminHttp interface {
	RegisterHttp(ms *microservice.Microservice, prefix string)
	ReSyncChartOfAccount(ms microservice.IContext) error
}

type ChartOfAccountAdminHttp struct {
	svc IChartOfAccountAdminService
}

func NewChartOfAccountAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IChartOfAccountAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewChartOfAccountAdminService(mongoPersister, producer)
	return &ChartOfAccountAdminHttp{
		svc: svc,
	}
}

func (s *ChartOfAccountAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/chartofaccountadmin/resyncchartofaccount", s.ReSyncChartOfAccount)
}

func (s *ChartOfAccountAdminHttp) ReSyncChartOfAccount(ctx microservice.IContext) error {

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

	err = s.svc.ReSyncChartOfAccountDoc(req.ShopID)
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
