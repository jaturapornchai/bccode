package accountadmin

import (
	"context"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IAccountAdminService interface {
	ListUser() ([]Account, error)
}

type AccountAdminService struct {
	repo           IAccountAdminRepository
	serviceTimeout time.Duration
}

func NewAccountAdminService(pst microservice.IPersisterMongo) IAccountAdminService {
	repo := NewAccountAdminRepository(pst)
	return &AccountAdminService{
		repo:           repo,
		serviceTimeout: time.Duration(30) * time.Second,
	}
}

func (s *AccountAdminService) ListUser() ([]Account, error) {

	ctx, ctxCancel := context.WithTimeout(context.Background(), s.serviceTimeout)
	defer ctxCancel()

	return s.repo.ListAllUser(ctx)
}
