package accountadmin

import (
	"smlcloudplatform/internal/config"
	goMicroModel "smlcloudplatform/internal/models"
	"smlcloudplatform/pkg/microservice"
)

type IAccountAdminHttp interface {
	ListUser(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, perfix string)
}

type AccountAdminHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc IAccountAdminService
}

func NewAccountAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IAccountAdminHttp {

	svc := NewAccountAdminService(microservice.NewPersisterMongo(cfg.MongoPersisterConfig()))
	return &AccountAdminHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (s *AccountAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.GET(prefix+"/accountadmin/listuser", s.ListUser)
}

func (s *AccountAdminHttp) ListUser(ctx microservice.IContext) error {

	userList, err := s.svc.ListUser()

	if err != nil {
		ctx.Response(500, goMicroModel.RestErrorResponse{
			ErrStatus: 500,
			ErrError:  err.Error(),
		})
		return err

	}
	ctx.Response(200, goMicroModel.ApiResponse{
		Success: true,
		Data:    userList,
	})
	return nil
}
