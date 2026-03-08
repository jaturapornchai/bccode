package utils

import (
	"smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
)

const RequestSearchMaxPage = 2_000_000_000 // can maximum 2_147_483_647
const RequestSearchMinPage = 1
const RequestSearchMaxLimit = 100_000
const RequestSearchMinLimit = 1

const RequestSearchDefaultOffset = 0
const RequestSearchDefaultLimit = 20

const RequestMaxOffset = 2_000_000_000 // can maximum 2_147_483_647
const RequestMinOffset = 0

func GetSearchQueryParam(fnGetParam func(string) string) string {
	q := strings.Trim(fnGetParam("q"), " ")
	return q
}

func GetPageParam(fnGetParam func(string) string) (int, int) {

	pageRawText := fnGetParam("page")
	limitRawText := fnGetParam("limit")

	page, err := strconv.Atoi(pageRawText) //strconv.ParseUint(pageRawText, 10, 32)
	if err != nil {
		page = RequestSearchDefaultOffset
	}

	limit, err := strconv.Atoi(limitRawText) //strconv.ParseUint(limitRawText, 10, 32)

	if err != nil {
		limit = RequestSearchDefaultLimit
	}

	if page < 1 {
		page = RequestSearchMinPage
	}

	if page > RequestSearchMaxPage {
		page = RequestSearchMaxPage
	}

	if limit < RequestSearchMinLimit {
		limit = RequestSearchMinLimit
	}

	if limit > RequestSearchMaxLimit {
		limit = RequestSearchMaxLimit
	}

	return page, limit
}

func GetStepParam(fnGetParam func(string) string) (int, int) {

	offsetRawText := fnGetParam("offset")
	limitRawText := fnGetParam("limit")

	offset, err := strconv.Atoi(offsetRawText) //strconv.ParseUint(pageRawText, 10, 32)
	if err != nil {
		offset = RequestSearchDefaultOffset
	}

	limit, err := strconv.Atoi(limitRawText) //strconv.ParseUint(limitRawText, 10, 32)

	if err != nil {
		limit = RequestSearchDefaultLimit
	}

	if offset < 0 {
		offset = RequestMinOffset
	}

	if offset > RequestSearchMaxPage {
		offset = RequestMaxOffset
	}

	if limit < RequestSearchMinLimit {
		limit = RequestSearchMinLimit
	}

	if limit > RequestSearchMaxLimit {
		limit = RequestSearchMaxLimit
	}

	return offset, limit
}

func GetSortParam(fnGetParam func(string) string) []models.KeyInt {
	tempSort := []models.KeyInt{}

	sortRawText := strings.Trim(fnGetParam("sort"), " ")

	if sortRawText == "" {
		return tempSort
	}

	sortSplitArr := strings.Split(sortRawText, ",")

	for _, sortFieldRaw := range sortSplitArr {
		sortFieldRawArr := strings.Split(sortFieldRaw, ":")
		if len(sortFieldRawArr) != 2 {
			continue
		}

		sortKey := sortFieldRawArr[0]
		sortValRaw := sortFieldRawArr[1]

		sortVal, err := strconv.Atoi(sortValRaw)
		if err != nil {
			sortVal = 1
		}

		tempSort = append(tempSort, models.KeyInt{
			Key:   sortKey,
			Value: int8(sortVal),
		})
	}

	return tempSort
}

func GetPageable(fnGetParam func(string) string) models.Pageable {
	q := GetSearchQueryParam(fnGetParam)
	page, limit := GetPageParam(fnGetParam)
	sorts := GetSortParam(fnGetParam)

	return models.Pageable{
		Query: q,
		Page:  page,
		Limit: limit,
		Sorts: sorts,
	}
}

func GetPageableStep(fnGetParam func(string) string) models.PageableStep {
	q := GetSearchQueryParam(fnGetParam)
	skip, limit := GetStepParam(fnGetParam)
	sorts := GetSortParam(fnGetParam)

	return models.PageableStep{
		Query: q,
		Skip:  skip,
		Limit: limit,
		Sorts: sorts,
	}
}
