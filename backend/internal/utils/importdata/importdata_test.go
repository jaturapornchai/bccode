package importdata_test

import (
	"fmt"
	"smlcloudplatform/internal/utils/importdata"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockUser struct {
	Code string
	Name string
}

func getMockUserID(u MockUser) string {
	return u.Code
}

func TestFilterDuplicate(t *testing.T) {

	userMockList := []MockUser{}

	for i := 1; i <= 5; i++ {
		code := fmt.Sprintf("code%03d", i)
		name := fmt.Sprintf("name %d", i)

		userMockList = append(userMockList, MockUser{
			Code: code,
			Name: name,
		})
	}

	userFiltered, userDuplicate := importdata.FilterDuplicate[MockUser](userMockList, getMockUserID)

	require.Equal(t, len(userFiltered), 5, "case 1: filtered data invalid")
	require.Equal(t, len(userDuplicate), 0, "case 1: duplicate data invalid")

	userMockList = append(userMockList, MockUser{
		Code: "code001",
		Name: "name 1",
	})

	userFiltered, userDuplicate = importdata.FilterDuplicate[MockUser](userMockList, getMockUserID)

	require.Equal(t, len(userFiltered), 5, "case 2: filtered data  invalid")
	require.Equal(t, len(userDuplicate), 1, "case 2: duplicate data invalid")
}

func TestPreparePayloadData(t *testing.T) {
	userMockList := []MockUser{}
	userMockGuidList := []string{}

	for i := 1; i <= 5; i++ {
		code := fmt.Sprintf("code%03d", i)
		name := fmt.Sprintf("name %d", i)

		userMockList = append(userMockList, MockUser{
			Code: code,
			Name: name,
		})

	}

	duplicateDataList, userCreateDataList := importdata.PreparePayloadData[MockUser, MockUser]("shopID", "authUser", userMockGuidList, userMockList, getMockUserID, prepareData)

	require.Equal(t, len(userCreateDataList), 5, "case 1: filtered data  invalid")
	require.Equal(t, len(duplicateDataList), 0, "case 1: duplicate data invalid")

	userMockGuidList = append(userMockGuidList, "code001")

	duplicateDataList, userCreateDataList = importdata.PreparePayloadData[MockUser, MockUser]("shopID", "authUser", userMockGuidList, userMockList, getMockUserID, prepareData)

	require.Equal(t, len(userCreateDataList), 4, "case 2: filtered data  invalid")
	require.Equal(t, len(duplicateDataList), 1, "case 2: duplicate data invalid")
}

func prepareData(shopID string, authUser string, user MockUser) MockUser {
	return user
}

func TestUpdateOnDuplicate(t *testing.T) {
	userMockList := []MockUser{}

	for i := 1; i <= 5; i++ {
		code := fmt.Sprintf("code%03d", i)
		name := fmt.Sprintf("name %d", i)

		userMockList = append(userMockList, MockUser{
			Code: code,
			Name: name,
		})
	}

	updateSuccess, updateFail := importdata.UpdateOnDuplicate[MockUser]("shopID", "authUser", userMockList, getMockUserID, findGuid, checkExistDoc, updateDoc)

	require.Equal(t, len(updateSuccess), 5, "case 1: filtered data  invalid")
	require.Equal(t, len(updateFail), 0, "case 1: duplicate data invalid")
}

func findGuid(shopID string, authUser string) (MockUser, error) {
	return MockUser{
		Code: shopID,
		Name: authUser,
	}, nil
}

func checkExistDoc(u MockUser) bool {
	return true
}

func updateDoc(shopID string, authUser string, u1 MockUser, u2 MockUser) error {
	return nil
}
