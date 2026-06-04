package mastersync

import (
	"smlcloudplatform/internal/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ActivityModuleManager struct {
	pst                microservice.IPersisterMongo
	activityModuleList map[string]ActivityModule
}

func NewActivityModuleManager(pst microservice.IPersisterMongo) *ActivityModuleManager {
	return &ActivityModuleManager{
		pst:                pst,
		activityModuleList: map[string]ActivityModule{},
	}
}

func (m *ActivityModuleManager) Add(activityModule ActivityModule) *ActivityModuleManager {
	m.activityModuleList[activityModule.GetModuleName()] = activityModule
	return m
}

func (m ActivityModuleManager) GetList() map[string]ActivityModule {
	return m.activityModuleList
}

func (m ActivityModuleManager) GetModules() []string {
	modules := []string{}
	for module := range m.activityModuleList {
		modules = append(modules, module)
	}
	return modules
}

func (m ActivityModuleManager) GetPage(moduleSelectList map[string]struct{}, activityParam ActivityParamPage) (map[string]interface{}, mongopagination.PaginationData, error) {
	moduleList := map[string]ActivityModule{}

	for _, activityModule := range m.activityModuleList {
		moduleList[activityModule.GetModuleName()] = activityModule
	}

	return listDataModulePage(moduleList, moduleSelectList, activityParam)
}

type ActivityModule interface {
	LastActivity(holdingCode string, action string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) (models.LastActivity, mongopagination.PaginationData, error)
	LastActivityStep(holdingCode string, action string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) (models.LastActivity, error)
	GetModuleName() string
}

type ActivityParamPage struct {
	HoldingCode string
	Action      string
	LastUpdate  time.Time
	Filters     string
	Pageable    micromodels.Pageable
}

type ActivityParamOffset struct {
	HoldingCode  string
	Action       string
	LastUpdate   time.Time
	Filters      string
	PageableStep micromodels.PageableStep
	ShopsID      []string
}

func listDataModulePage(appModules map[string]ActivityModule, moduleSelectList map[string]struct{}, param ActivityParamPage) (map[string]interface{}, mongopagination.PaginationData, error) {

	result := map[string]interface{}{}

	resultPagination := mongopagination.PaginationData{}
	for moduleName, appModule := range appModules {
		if len(moduleSelectList) == 0 || isSelectModule(moduleSelectList, moduleName) {
			filters := filterRawTextToMap(param.Filters)
			docList, pagination, err := appModule.LastActivity(param.HoldingCode, param.Action, param.LastUpdate, filters, param.Pageable)

			if err != nil {
				return map[string]interface{}{}, mongopagination.PaginationData{}, err
			}

			result[moduleName] = docList

			if pagination.Total > resultPagination.Total {
				resultPagination = pagination
			}
		}
	}

	return result, resultPagination, nil
}

func listDataModuleOffset(appModules map[string]ActivityModule, moduleSelectList map[string]struct{}, param ActivityParamOffset) (map[string]interface{}, error) {

	result := map[string]interface{}{}

	for moduleName, appModule := range appModules {
		if len(moduleSelectList) == 0 || isSelectModule(moduleSelectList, moduleName) {

			filters := filterRawTextToMap(param.Filters)

			// Use ShopsID for productbarcode module if provided
			holdingCode := param.HoldingCode
			if moduleName == "productbarcode" && len(param.ShopsID) > 0 {
				// For productbarcode module, use the first holding Code from ShopsID list
				// or pass the entire list to the module for IN query handling
				holdingCode = strings.Join(param.ShopsID, ",")
			}

			docList, err := appModule.LastActivityStep(holdingCode, param.Action, param.LastUpdate, filters, param.PageableStep)

			if err != nil {
				return map[string]interface{}{}, err
			}

			result[moduleName] = docList

		}
	}

	return result, nil
}

func filterRawTextToMap(rawText string) map[string]interface{} {
	filters := map[string]interface{}{}

	splitText := strings.Split(rawText, ",")
	for _, text := range splitText {
		splitText := strings.Split(text, ":")
		if len(splitText) == 2 {
			filters[splitText[0]] = splitText[1]
		}
	}

	return filters
}

func isSelectModule(moduleList map[string]struct{}, moduleKey string) bool {
	_, ok := moduleList[moduleKey]
	return ok
}
