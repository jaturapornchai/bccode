package services_test

import (
	"encoding/json"
	pkg_models "smlcloudplatform/internal/models"
	trans_models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/saleinvoice/models"
	"smlcloudplatform/internal/transaction/saleinvoice/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCSV(t *testing.T) {

	svcExport := services.SaleInvocieExport{}

	giveRaw := `{
		"docNo": "SI0001",
		"docDatetime": "2021-01-01T00:00:00Z",
		"details": [
			{
				"barcode": "B0001",
				"itemcode": "I0001",
				"itemnames": [{
					"code": "en",
					"name": "Item 1"
				}],
				"itemtype": 1,
				"unitcode": "U0001",
				"unitnames": [
					{
						"code": "en",
						"name": "Unit 1"
					}
				],
				"qty": 10.00,
				"price": 100.00,
				"discountamount": 10.00,
				"sumamount": 900.00
			}
		]
	}`

	saleInvoiceInfo := models.SaleInvoiceInfo{}
	err := json.Unmarshal([]byte(giveRaw), &saleInvoiceInfo)
	require.NoError(t, err)

	result := svcExport.ParseCSV("en", saleInvoiceInfo)

	require.Equal(t, 1, len(result))
	require.Equal(t, 10, len(result[0]))
	assert.Equal(t, "2021-01-01", result[0][0])
	assert.Equal(t, "SI0001", result[0][1])
	assert.Equal(t, "B0001", result[0][2])
	assert.Equal(t, "Item 1", result[0][3])
	assert.Equal(t, "U0001", result[0][4])
	assert.Equal(t, "Unit 1", result[0][5])
	assert.Equal(t, "10.00", result[0][6])
	assert.Equal(t, "100.00", result[0][7])
	assert.Equal(t, "10.00", result[0][8])
	assert.Equal(t, "900.00", result[0][9])

}

func TestGetName(t *testing.T) {

	svcExport := services.SaleInvocieExport{}

	giveRaw := `[
		{
			"code": "en",
			"name": "Item 1"
		},
		{
			"code": "th",
			"name": "รายการ 1"
		}
	]`

	names := []pkg_models.NameX{}

	err := json.Unmarshal([]byte(giveRaw), &names)
	require.NoError(t, err)

	result := svcExport.GetName(&names, "en")
	assert.Equal(t, "Item 1", result)

	result = svcExport.GetName(&names, "th")
	assert.Equal(t, "รายการ 1", result)

	result = svcExport.GetName(&names, "jp")
	assert.Equal(t, "", result)

}

func TestParseDetailString(t *testing.T) {

	svcExport := services.SaleInvocieExport{}

	giveRaw := `{
		"docNo": "SI0001",
		"docDatetime": "2021-01-01T00:00:00Z",
		"barcode": "B0001",
		"itemcode": "I0001",
		"itemnames": [{
			"code": "en",
			"name": "Item 1"
		}],
		"itemtype": 1,
		"unitcode": "U0001",
		"unitnames": [
			{
				"code": "en",
				"name": "Unit 1"
			}
		],
		"qty": 10.00,
		"price": 100.00,
		"discountamount": 10.00,
		"sumamount": 900.00
	}`
	detail := trans_models.Detail{}
	err := json.Unmarshal([]byte(giveRaw), &detail)
	require.NoError(t, err)

	result := svcExport.ParseDetailString("en", "SI0001", time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC), detail)

	require.Equal(t, 10, len(result))
	assert.Equal(t, "2021-01-01", result[0])
	assert.Equal(t, "SI0001", result[1])
	assert.Equal(t, "B0001", result[2])
	assert.Equal(t, "Item 1", result[3])
	assert.Equal(t, "U0001", result[4])
	assert.Equal(t, "Unit 1", result[5])
	assert.Equal(t, "10.00", result[6])
	assert.Equal(t, "100.00", result[7])
	assert.Equal(t, "10.00", result[8])
	assert.Equal(t, "900.00", result[9])

}
