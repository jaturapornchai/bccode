package services

import (
	"context"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"sync"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ActivityService[TCU any, TDEL any] struct {
	repo repositories.IActivityRepository[TCU, TDEL]
}

func NewActivityService[TCU any, TDEL any](repo repositories.IActivityRepository[TCU, TDEL]) ActivityService[TCU, TDEL] {
	return ActivityService[TCU, TDEL]{
		repo: repo,
	}
}

func (svc *ActivityService[TCU, TDEL]) InitialActivityService(pst microservice.IPersisterMongo, repo repositories.IActivityRepository[TCU, TDEL]) {
	// repo.InitialActivityRepository(pst)
	svc.repo = repo
}

func (svc ActivityService[TCU, TDEL]) LastActivity(shopID string, action string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) (common.LastActivity, mongopagination.PaginationData, error) {

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Duration(15)*time.Second)
	defer ctxCancel()

	lastActivity := common.LastActivity{}
	var wg sync.WaitGroup

	isActionRemove := len(action) == 0 || action == "all" || action == "remove"
	isActionNew := len(action) == 0 || action == "all" || action == "new"

	var deleteDocList []TDEL
	pagination1 := mongopagination.PaginationData{}
	var errFindDel error

	if isActionRemove {
		wg.Add(1)
		go func() {
			deleteDocList, pagination1, errFindDel = svc.repo.FindDeletedPage(ctx, shopID, lastUpdatedDate, filters, pageable)
			wg.Done()
		}()
	}

	var createAndUpdateDocList []TCU
	pagination2 := mongopagination.PaginationData{}
	var errFindCreateUpdate error

	if isActionNew {
		wg.Add(1)
		go func() {
			createAndUpdateDocList, pagination2, errFindCreateUpdate = svc.repo.FindCreatedOrUpdatedPage(ctx, shopID, lastUpdatedDate, filters, pageable)
			wg.Done()
		}()
	}

	wg.Wait()

	if isActionRemove {
		if errFindDel != nil {
			return common.LastActivity{}, pagination1, errFindDel
		}
		lastActivity.Remove = &deleteDocList
	}

	if isActionNew {
		if errFindCreateUpdate != nil {
			return common.LastActivity{}, pagination2, errFindCreateUpdate
		}
		lastActivity.New = &createAndUpdateDocList
	}

	pagination := pagination1

	if pagination.Total < pagination2.Total {
		pagination = pagination2
	}

	return lastActivity, pagination, nil
}

func (svc ActivityService[TCU, TDEL]) LastActivityStep(shopID string, action string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) (common.LastActivity, error) {

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Duration(15)*time.Second)
	defer ctxCancel()

	lastActivity := common.LastActivity{}

	var wg sync.WaitGroup

	isActionRemove := len(action) == 0 || action == "all" || action == "remove"
	isActionNew := len(action) == 0 || action == "all" || action == "new"

	var errFindDel error
	var deleteDocList []TDEL
	if isActionRemove {
		wg.Add(1)
		go func() {
			deleteDocList, errFindDel = svc.repo.FindDeletedStep(ctx, shopID, lastUpdatedDate, filters, pageableStep)
			wg.Done()
		}()
	}

	var createAndUpdateDocList []TCU
	var errFindCreateUpdate error
	if isActionNew {
		wg.Add(1)
		go func() {
			createAndUpdateDocList, errFindCreateUpdate = svc.repo.FindCreatedOrUpdatedStep(ctx, shopID, lastUpdatedDate, filters, pageableStep)
			wg.Done()
		}()
	}

	wg.Wait()

	if isActionRemove {
		if errFindDel != nil {
			return common.LastActivity{}, errFindDel
		}
		lastActivity.Remove = &deleteDocList
	}

	if isActionNew {
		if errFindCreateUpdate != nil {
			return common.LastActivity{}, errFindCreateUpdate
		}
		lastActivity.New = &createAndUpdateDocList
	}

	return lastActivity, nil
}
