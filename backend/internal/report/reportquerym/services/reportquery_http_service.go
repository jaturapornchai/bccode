package services

import (
	"context"
	"errors"
	"fmt"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/report/reportquerym/models"
	"smlcloudplatform/internal/report/reportquerym/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IReportQueryHttpService interface {
	PlaygroundReportQuery(shopID string, query models.Query) ([]map[string]interface{}, error)
	ExecuteReportQuery(shopID string, reportCode string, queryParamRequest []models.QueryParamRequest, pageable micromodels.Pageable) ([]map[string]interface{}, common.Pagination, error)

	CreateReportQuery(shopID string, authUsername string, doc models.ReportQuery) (string, error)
	UpdateReportQuery(shopID string, guid string, authUsername string, doc models.ReportQuery) error
	DeleteReportQuery(shopID string, guid string, authUsername string) error
	DeleteReportQueryByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoReportQuery(shopID string, guid string) (models.ReportQueryInfo, error)
	InfoReportQueryByCode(shopID string, code string) (models.ReportQueryInfo, error)
	SearchReportQuery(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReportQueryInfo, mongopagination.PaginationData, error)
	SearchReportQueryStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.ReportQueryInfo, int, error)

	GetModuleName() string
}

type ReportQueryHttpService struct {
	repo           repositories.IReportQueryRepository
	contextTimeout time.Duration
}

func NewReportQueryHttpService(repo repositories.IReportQueryRepository) *ReportQueryHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &ReportQueryHttpService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}

	return insSvc
}

func (svc ReportQueryHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ReportQueryHttpService) CreateReportQuery(shopID string, authUsername string, doc models.ReportQuery) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("Code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.ReportQueryDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.ReportQuery = doc

	docData.IsActived = true
	docData.IsApproved = true

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc ReportQueryHttpService) UpdateReportQuery(shopID string, guid string, authUsername string, doc models.ReportQuery) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	docData := findDoc

	docData.ReportQuery = doc

	docData.IsActived = findDoc.IsActived
	docData.IsApproved = findDoc.IsApproved

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return err
	}

	return nil
}

func (svc ReportQueryHttpService) DeleteReportQuery(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc ReportQueryHttpService) DeleteReportQueryByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc ReportQueryHttpService) InfoReportQuery(shopID string, guid string) (models.ReportQueryInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ReportQueryInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ReportQueryInfo{}, errors.New("document not found")
	}

	return findDoc.ReportQueryInfo, nil
}

func (svc ReportQueryHttpService) InfoReportQueryByCode(shopID string, code string) (models.ReportQueryInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.ReportQueryInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ReportQueryInfo{}, errors.New("document not found")
	}

	return findDoc.ReportQueryInfo, nil
}

func (svc ReportQueryHttpService) SearchReportQuery(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReportQueryInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.ReportQueryInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ReportQueryHttpService) SearchReportQueryStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.ReportQueryInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	if langCode != "" {
		selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
	} else {
		selectFields["names"] = 1
	}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ReportQueryInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ReportQueryHttpService) PlaygroundReportQuery(shopID string, query models.Query) ([]map[string]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	selectFields := bson.M{}

	for _, v := range query.Fields {
		if v == "deletedat" || v == "deletedby" {
			continue
		}
		selectFields[v] = 1
	}

	if len(selectFields) > 0 {
		selectFields["_id"] = 0
	}

	var filterRaw bson.M
	err := bson.UnmarshalExtJSON([]byte(query.Filter), true, &filterRaw)
	if err != nil {

		return []map[string]interface{}{}, err
	}

	params := map[string]interface{}{}

	for _, v := range query.Params {
		params[v.Name] = v.Value
	}

	filters, err := ReplacePlaceholdersInMap(filterRaw, &params)

	filters["deletedat"] = bson.M{"$exists": false}

	if err != nil {
		return []map[string]interface{}{}, err
	}

	result, err := svc.repo.Playground(ctx, query.Collection, selectFields, filters)
	if err != nil {
		return []map[string]interface{}{}, err
	}

	return result, nil
}

func (svc ReportQueryHttpService) ExecuteReportQuery(shopID string, reportCode string, queryParamRequest []models.QueryParamRequest, pageable micromodels.Pageable) ([]map[string]interface{}, common.Pagination, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOneByCode(ctx, reportCode)

	if err != nil {
		return nil, common.Pagination{}, err
	}

	query := models.Query{
		Collection: findDoc.Collection,
		Filter:     findDoc.Filter,
		Fields:     *findDoc.Fields,
	}

	queryParamRequested := map[string]models.QueryParamRequest{}
	queryParamRequired := map[string]models.QueryParam{}

	if findDoc.Params == nil {
		findDoc.Params = &[]models.ReportParam{}
	}

	for _, item := range queryParamRequest {
		queryParamRequested[item.Name] = item
	}

	for _, item := range *findDoc.Params {
		queryParamRequired[item.Name] = models.QueryParam{
			Name: fmt.Sprintf("@%s", item.Name),
			Type: item.Type,
		}

		if _, ok := queryParamRequested[item.Name]; !ok {
			return nil, common.Pagination{}, fmt.Errorf("parameter %s is required", item.Name)
		}
	}

	for _, item := range queryParamRequest {
		if _, ok := queryParamRequired[item.Name]; ok {
			temp := queryParamRequired[item.Name]
			temp.Value = item.Value
			queryParamRequired[item.Name] = temp

		}
	}

	for _, item := range queryParamRequired {
		query.Params = append(query.Params, item)
	}

	selectFields := bson.M{}

	for _, v := range query.Fields {
		if v == "deletedat" || v == "deletedby" {
			continue
		}
		selectFields[v] = 1
	}

	if len(selectFields) > 0 {
		selectFields["_id"] = 0
	}

	var filterRaw bson.M
	err = bson.UnmarshalExtJSON([]byte(query.Filter), true, &filterRaw)
	if err != nil {

		return []map[string]interface{}{}, common.Pagination{}, err
	}

	params := map[string]interface{}{}

	for _, v := range query.Params {
		params[v.Name] = v.Value
	}

	filters, err := ReplacePlaceholdersInMap(filterRaw, &params)

	filters["deletedat"] = bson.M{"$exists": false}

	if err != nil {
		return []map[string]interface{}{}, common.Pagination{}, err
	}

	result, pagination, err := svc.repo.Execute(ctx, query.Collection, selectFields, filters, pageable)
	if err != nil {
		return []map[string]interface{}{}, pagination, err
	}

	return result, pagination, nil
}

func (svc ReportQueryHttpService) GetModuleName() string {
	return "reportQuery"
}
