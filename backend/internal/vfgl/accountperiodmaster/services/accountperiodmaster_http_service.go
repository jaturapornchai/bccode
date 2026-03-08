package services

import (
	"context"
	"errors"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/vfgl/accountperiodmaster/models"
	"smlcloudplatform/internal/vfgl/accountperiodmaster/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IAccountPeriodMasterHttpService interface {
	CreateAccountPeriodMaster(shopID string, authUsername string, doc models.AccountPeriodMaster) (string, error)
	UpdateAccountPeriodMaster(shopID string, guid string, authUsername string, doc models.AccountPeriodMaster) error
	DeleteAccountPeriodMaster(shopID string, guid string, authUsername string) error
	DeleteAccountPeriodMasterByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoAccountPeriodMaster(shopID string, guid string) (models.AccountPeriodMasterInfo, error)
	InfoAccountPeriodMasterByDate(shopID string, findDate time.Time) (models.AccountPeriodMasterInfo, error)
	InfoAccountPeriodMasterByDateList(shopID string, findDateList []time.Time) ([]models.MapDateAccountPeriodMasterInfo, error)
	SearchAccountPeriodMaster(shopID string, pageable micromodels.Pageable) ([]models.AccountPeriodMasterInfo, mongopagination.PaginationData, error)
	SearchAccountPeriodMasterStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.AccountPeriodMasterInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.AccountPeriodMaster) error
}

type AccountPeriodMasterHttpService struct {
	repo           repositories.IAccountPeriodMasterRepository
	contextTimeout time.Duration
}

func NewAccountPeriodMasterHttpService(repo repositories.IAccountPeriodMasterRepository) *AccountPeriodMasterHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return &AccountPeriodMasterHttpService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}
}

func (svc AccountPeriodMasterHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc AccountPeriodMasterHttpService) CreateAccountPeriodMaster(shopID string, authUsername string, doc models.AccountPeriodMaster) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByPeriod(ctx, shopID, doc.Period)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("period already exists")
	}

	findDocExists, err := svc.repo.FindByDateRange(ctx, shopID, doc.StartDate, doc.EndDate)

	if err != nil {
		return "", err
	}

	if len(findDocExists.GuidFixed) > 0 {
		return "", errors.New("date range already exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.AccountPeriodMasterDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.AccountPeriodMaster = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc AccountPeriodMasterHttpService) UpdateAccountPeriodMaster(shopID string, guid string, authUsername string, doc models.AccountPeriodMaster) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDocPeriod, err := svc.repo.FindByPeriod(ctx, shopID, doc.Period)

	if err != nil {
		return err
	}

	if len(findDocPeriod.GuidFixed) > 0 && findDocPeriod.GuidFixed != guid {
		return errors.New("period already exists")
	}

	findDocExists, err := svc.repo.FindByDateRange(ctx, shopID, doc.StartDate, doc.EndDate)

	if err != nil {
		return err
	}

	if len(findDocExists.GuidFixed) > 0 && findDocExists.GuidFixed != guid {
		return errors.New("date range already exists")
	}

	findDoc.AccountPeriodMaster = doc
	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc AccountPeriodMasterHttpService) DeleteAccountPeriodMaster(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc AccountPeriodMasterHttpService) DeleteAccountPeriodMasterByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc AccountPeriodMasterHttpService) InfoAccountPeriodMaster(shopID string, guid string) (models.AccountPeriodMasterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.AccountPeriodMasterInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.AccountPeriodMasterInfo{}, errors.New("document not found")
	}

	return findDoc.AccountPeriodMasterInfo, nil

}

func (svc AccountPeriodMasterHttpService) InfoAccountPeriodMasterByDate(shopID string, findDate time.Time) (models.AccountPeriodMasterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDateRange(ctx, shopID, findDate, findDate)

	if err != nil {
		return models.AccountPeriodMasterInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.AccountPeriodMasterInfo{}, errors.New("document not found")
	}

	return findDoc.AccountPeriodMasterInfo, nil

}

func (svc AccountPeriodMasterHttpService) SearchAccountPeriodMaster(shopID string, pageable micromodels.Pageable) ([]models.AccountPeriodMasterInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"docno",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return []models.AccountPeriodMasterInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc AccountPeriodMasterHttpService) SearchAccountPeriodMasterStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.AccountPeriodMasterInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"docno",
	}

	selectFields := map[string]interface{}{
		"guidfixed": 1,
		"docno":     1,
	}

	if langCode != "" {
		selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
	} else {
		selectFields["names"] = 1
	}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.AccountPeriodMasterInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc AccountPeriodMasterHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.AccountPeriodMaster) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.ValidatePeriod(dataList)
	if err != nil {
		return err
	}

	docList := []models.AccountPeriodMasterDoc{}

	for _, doc := range dataList {

		findDoc, err := svc.repo.FindByPeriod(ctx, shopID, doc.Period)

		if err != nil {
			return err
		}

		if len(findDoc.GuidFixed) > 0 {
			return errors.New("period already exists")
		}

		findDocExists, err := svc.repo.FindByDateRange(ctx, shopID, doc.StartDate, doc.EndDate)

		if err != nil {
			return err
		}

		if len(findDocExists.GuidFixed) > 0 {
			return errors.New("date range already exists")
		}

		newGuidFixed := utils.NewGUID()

		docData := models.AccountPeriodMasterDoc{}
		docData.ShopID = shopID
		docData.GuidFixed = newGuidFixed
		docData.AccountPeriodMaster = doc

		docData.CreatedBy = authUsername
		docData.CreatedAt = time.Now()

		docList = append(docList, docData)
	}

	err = svc.repo.CreateInBatch(ctx, docList)

	if err != nil {
		return err
	}

	return nil
}

func (svc AccountPeriodMasterHttpService) InfoAccountPeriodMasterByDateList(shopID string, findDateList []time.Time) ([]models.MapDateAccountPeriodMasterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	periodData, err := svc.repo.FindAll(ctx, shopID)

	if err != nil {
		return []models.MapDateAccountPeriodMasterInfo{}, err
	}

	if len(periodData) == 0 {
		return []models.MapDateAccountPeriodMasterInfo{}, errors.New("document not found")
	}

	mapAccountPeriodMaster := []models.MapDateAccountPeriodMasterInfo{}

	for _, findDate := range findDateList {
		tempMapeAccountPeriodMaster := svc.MapDateToAccountPeriod(periodData, findDate)
		mapAccountPeriodMaster = append(mapAccountPeriodMaster, tempMapeAccountPeriodMaster)
	}

	return mapAccountPeriodMaster, nil

}

func (svc AccountPeriodMasterHttpService) MapDateToAccountPeriod(periodData []models.AccountPeriodMasterDoc, findDate time.Time) models.MapDateAccountPeriodMasterInfo {
	tempMapeAccountPeriodMaster := models.MapDateAccountPeriodMasterInfo{}
	tempMapeAccountPeriodMaster.Date = findDate.Format("2006-01-02")
	tempMapeAccountPeriodMaster.PeriodData = models.AccountPeriodMasterInfo{}

	for _, doc := range periodData {
		if svc.IsInDateRange(doc.StartDate, doc.EndDate, findDate) {
			tempMapeAccountPeriodMaster.PeriodData = doc.AccountPeriodMasterInfo
			break
		}
	}
	return tempMapeAccountPeriodMaster
}

func (svc AccountPeriodMasterHttpService) findMinAndMaxTimes(times []time.Time) (min, max time.Time) {
	if len(times) == 0 {
		return time.Time{}, time.Time{}
	}

	min, max = times[0], times[0]

	for _, t := range times[1:] {
		if t.Before(min) {
			min = t
		}
		if t.After(max) {
			max = t
		}
	}

	return min, max
}

// func ValidatePeriod(dataList []models.AccountPeriodMaster) error {
// 	periodData := map[int]models.AccountPeriodMaster{}

// 	for _, doc := range dataList {
// 		if _, ok := periodData[doc.Period]; ok {
// 			return errors.New("period is duplicate")
// 		}
// 		periodData[doc.Period] = doc
// 	}

// 	for periodKey, doc := range periodData {
// 		for periodKeyCheck, docCheck := range periodData {
// 			if periodKey == periodKeyCheck {
// 				continue
// 			}

// 			if InDateTimeSpan(docCheck.StartDate, docCheck.EndDate, doc.StartDate, doc.EndDate) {
// 				return errors.New("date range invalid")
// 			}
// 		}
// 	}
// 	return nil
// }

func (svc AccountPeriodMasterHttpService) ValidatePeriod(periodList []models.AccountPeriodMaster) error {
	periodMap, duplicatePeriodErr := svc.buildPeriodMap(periodList)
	if duplicatePeriodErr != nil {
		return duplicatePeriodErr
	}

	if err := svc.checkForInvalidDateRanges(periodMap); err != nil {
		return err
	}

	return nil
}

func (svc AccountPeriodMasterHttpService) buildPeriodMap(periodList []models.AccountPeriodMaster) (map[int]models.AccountPeriodMaster, error) {
	periodMap := make(map[int]models.AccountPeriodMaster)

	for _, period := range periodList {
		if _, ok := periodMap[period.Period]; ok {
			return nil, errors.New("period is duplicate")
		}
		periodMap[period.Period] = period
	}

	return periodMap, nil
}

func (svc AccountPeriodMasterHttpService) checkForInvalidDateRanges(periodMap map[int]models.AccountPeriodMaster) error {
	for _, currentPeriod := range periodMap {
		for _, comparisonPeriod := range periodMap {
			if currentPeriod.Period == comparisonPeriod.Period {
				continue
			}

			if svc.IsInDateTimeSpan(comparisonPeriod.StartDate, comparisonPeriod.EndDate, currentPeriod.StartDate, currentPeriod.EndDate) {
				return errors.New("date range invalid")
			}
		}
	}
	return nil
}

func (svc AccountPeriodMasterHttpService) IsInDateTimeSpan(fromDate, toDate, checkFromDate, checkToDate time.Time) bool {
	toDate = toDate.AddDate(0, 0, 1)

	isCheckFromDateInRange := checkFromDate.Equal(fromDate) || checkFromDate.After(fromDate)
	isCheckToDateInRange := checkToDate.Equal(fromDate) || checkToDate.After(fromDate)
	rangeFrom := isCheckFromDateInRange || isCheckToDateInRange

	isCheckFromDateBeforeToDate := checkFromDate.Before(toDate)
	isCheckToDateBeforeToDate := checkToDate.Before(toDate)
	rangeTo := isCheckFromDateBeforeToDate || isCheckToDateBeforeToDate

	return rangeFrom && rangeTo
}

func (svc AccountPeriodMasterHttpService) IsInDateRange(fromDate, toDate, checkDate time.Time) bool {
	toDate = toDate.AddDate(0, 0, 1)

	rangeFrom := checkDate.Equal(fromDate) || checkDate.After(fromDate)
	rangeTo := checkDate.Before(toDate)

	return rangeFrom && rangeTo
}
