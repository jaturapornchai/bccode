package services_test

import (
	"fmt"
	"smlcloudplatform/internal/vfgl/accountperiodmaster/models"
	"smlcloudplatform/internal/vfgl/accountperiodmaster/services"
	"testing"
	"time"

	"github.com/tj/assert"
)

func TestIsInDateSpan(t *testing.T) {

	dateFormat := "2006-01-02"
	svc := services.AccountPeriodMasterHttpService{}

	fromDate, _ := time.Parse(dateFormat, "2020-04-01")
	toDate, _ := time.Parse(dateFormat, "2020-04-30")

	// checkFromDate, _ := time.Parse(dateFormat, "2020-03-01")
	// checkToDate, _ := time.Parse(dateFormat, "2020-04-01")

	gives := []struct {
		fromDate string
		toDate   string
		want     bool
	}{
		{"2020-03-01", "2020-03-31", false},
		{"2020-03-01", "2020-04-01", true},
		{"2020-03-31", "2020-04-01", true},
		{"2020-03-31", "2020-04-05", true},
		{"2020-03-10", "2020-04-20", true},
		{"2020-04-30", "2020-05-01", true},
		{"2020-05-01", "2020-05-10", false},
	}

	for _, give := range gives {
		checkFromDate, _ := time.Parse(dateFormat, give.fromDate)
		checkToDate, _ := time.Parse(dateFormat, give.toDate)
		result := svc.IsInDateTimeSpan(fromDate, toDate, checkFromDate, checkToDate)
		fmt.Printf("%s %s %t %t\n", give.fromDate, give.toDate, give.want, result)
		assert.Equal(t, give.want, result)
	}
}

func TestIsInDateRange(t *testing.T) {

	dateFormat := "2006-01-02"

	svc := services.AccountPeriodMasterHttpService{}

	fromDate, _ := time.Parse(dateFormat, "2020-04-01")
	toDate, _ := time.Parse(dateFormat, "2020-04-30")

	gives := []struct {
		checkDate string
		want      bool
	}{
		{"2020-03-01", false},
		{"2020-03-31", false},
		{"2020-04-01", true},
		{"2020-04-15", true},
		{"2020-04-30", true},
		{"2020-05-01", false},
		{"2021-04-01", false},
	}

	for _, give := range gives {
		checkDate, _ := time.Parse(dateFormat, give.checkDate)
		result := svc.IsInDateRange(fromDate, toDate, checkDate)
		fmt.Printf("%s %t %t\n", give.checkDate, give.want, result)
		assert.Equal(t, give.want, result)
	}
}

func TestValidatePeriod(t *testing.T) {

	dateFormat := "2006-01-02"
	svc := services.AccountPeriodMasterHttpService{}

	ts1, _ := time.Parse(dateFormat, "2020-01-01")
	te1, _ := time.Parse(dateFormat, "2020-01-31")

	ts2, _ := time.Parse(dateFormat, "2020-03-01")
	te2, _ := time.Parse(dateFormat, "2020-03-31")

	ts3, _ := time.Parse(dateFormat, "2020-03-01")
	te3, _ := time.Parse(dateFormat, "2020-03-31")

	docList := []models.AccountPeriodMaster{
		{
			Period:    1,
			StartDate: ts1,
			EndDate:   te1,
		},
		{
			Period:    2,
			StartDate: ts2,
			EndDate:   te2,
		},
		{
			Period:    3,
			StartDate: ts3,
			EndDate:   te3,
		},
	}

	err := svc.ValidatePeriod(docList)

	if err != nil {
		fmt.Println(err)
	}
}

func TestMapDatePeriod(t *testing.T) {

	dateFormat := "2006-01-02"

	svc := services.AccountPeriodMasterHttpService{}

	givePeriodList := buildGivePeriodList()

	testCases := []struct {
		checkDate string
		want      int
	}{
		{"2019-01-01", 0},
		{"2020-01-01", 1},
		{"2020-03-31", 2},
		{"2020-04-01", 3},
		{"2020-04-15", 3},
		{"2020-04-30", 3},
		{"2020-05-01", 0},
	}

	for _, tc := range testCases {
		checkDate, _ := time.Parse(dateFormat, tc.checkDate)
		datePeriodMap := svc.MapDateToAccountPeriod(givePeriodList, checkDate)
		fmt.Printf("%s  : %d %d\n", tc.checkDate, tc.want, datePeriodMap.PeriodData.AccountPeriodMaster.Period)
		assert.Equal(t, tc.want, datePeriodMap.PeriodData.AccountPeriodMaster.Period)
	}

	// fmt.Println(datePeriodMap)
}

func buildGivePeriodList() []models.AccountPeriodMasterDoc {
	dateFormat := "2006-01-02"
	gives := []struct {
		period   int
		fromDate string
		toDate   string
	}{
		{1, "2020-01-01", "2020-01-31"},
		{2, "2020-03-01", "2020-03-31"},
		{3, "2020-04-01", "2020-03-30"},
	}

	givePeriodList := []models.AccountPeriodMasterDoc{}

	for _, give := range gives {
		doc := models.AccountPeriodMasterDoc{}
		doc.Period = give.period
		doc.StartDate, _ = time.Parse(dateFormat, give.fromDate)
		doc.EndDate, _ = time.Parse(dateFormat, give.toDate)
		givePeriodList = append(givePeriodList, doc)
	}

	return givePeriodList
}
