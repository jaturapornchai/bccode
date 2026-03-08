package importdata

func FilterDuplicate[TDATA any](docList []TDATA, fnGetID func(TDATA) string) (itemFiltered []TDATA, itemDuplicate []TDATA) {
	tempFilterDict := map[string]TDATA{}
	for _, doc := range docList {
		idKey := fnGetID(doc)
		if _, ok := tempFilterDict[idKey]; ok {
			itemDuplicate = append(itemDuplicate, doc)

		}
		tempFilterDict[idKey] = doc
	}

	for _, doc := range tempFilterDict {
		itemFiltered = append(itemFiltered, doc)
	}

	return itemFiltered, itemDuplicate
}

func PreparePayloadData[TDATA any, TDOC any](
	shopID string,
	authUsername string,
	itemGuidList []string,
	payloadCategoryList []TDATA,
	fnGetID func(TDATA) string,
	fnPrepareData func(string, string, TDATA) TDOC,
) ([]TDATA, []TDOC) {

	tempItemGuidDict := make(map[string]bool)
	duplicateDataList := []TDATA{}
	createDataList := []TDOC{}

	for _, itemGuid := range itemGuidList {
		tempItemGuidDict[itemGuid] = true
	}

	for _, doc := range payloadCategoryList {
		idKey := fnGetID(doc)
		if _, ok := tempItemGuidDict[idKey]; ok {
			duplicateDataList = append(duplicateDataList, doc)
		} else {
			dataDoc := fnPrepareData(shopID, authUsername, doc)
			createDataList = append(createDataList, dataDoc)
		}
	}
	return duplicateDataList, createDataList
}

func UpdateOnDuplicate[TDATA any, TDOC any](
	shopID string,
	authUsername string,
	duplicateDataList []TDATA,
	fnGetID func(TDATA) string,
	fnFindGuid func(string, string) (TDOC, error),
	fnCheckExistsDoc func(TDOC) bool,
	fnUptdateDoc func(string, string, TDATA, TDOC) error,
) ([]TDOC, []TDATA) {

	updateSuccessDataList := []TDOC{}
	updateFailDataList := []TDATA{}

	for _, doc := range duplicateDataList {
		idKey := fnGetID(doc)
		findDoc, err := fnFindGuid(shopID, idKey)

		if !(fnCheckExistsDoc(findDoc)) {
			updateFailDataList = append(updateFailDataList, doc)
			continue
		}

		err = fnUptdateDoc(shopID, authUsername, doc, findDoc)

		if err != nil {
			updateFailDataList = append(updateFailDataList, doc)
			continue
		}

		updateSuccessDataList = append(updateSuccessDataList, findDoc)
	}
	return updateSuccessDataList, updateFailDataList
}
