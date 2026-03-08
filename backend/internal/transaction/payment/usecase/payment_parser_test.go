package usecase_test

import (
	"encoding/json"
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"
	payment_models "smlcloudplatform/internal/transaction/payment/models"
	"smlcloudplatform/internal/transaction/payment/usecase"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTransactionToPayment(t *testing.T) {

	langCode := "en"
	langName := "BranchTest"

	branch := []models.NameX{
		{
			Code: &langCode,
			Name: &langName,
		},
	}

	testCases := []struct {
		name     string
		doc      transmodels.TransactionMessageQueue
		expected payment_models.TransactionPayment
	}{
		{
			name: "Parse Transaction to Payment",
			doc: transmodels.TransactionMessageQueue{
				Transaction: transmodels.Transaction{
					TransactionHeader: transmodels.TransactionHeader{
						DocNo:            "doc1",
						PaymentDetailRaw: "{\"cashamount\":1.0,\"cashamounttext\":\"\",\"paymentcreditcards\":[],\"paymenttransfers\":[]}",
						TransFlag:        50,
						Branch:           transmodels.TransactionBranch{Code: "0001", Names: &branch},
						CustCode:         "c01",
						CustNames: &[]models.NameX{
							*models.NewNameXWithCodeName("th", "ลูกค้าทดสอบ"),
						},
					},
				},
			},
			expected: payment_models.TransactionPayment{
				BranchCode:  "0001",
				BranchNames: branch,
				TransFlag:   50,
				CustCode:    "c01",
				CustNames: []models.NameX{
					*models.NewNameXWithCodeName("th", "ลูกค้าทดสอบ"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentDoc, err := usecase.ParseTransactionToPayment(tc.doc)

			assert.NoError(t, err, "Error should be nil")
			assert.Equal(t, tc.expected.BranchCode, paymentDoc.BranchCode, "Branch code invalid")
			assert.Equal(t, tc.expected.BranchNames, paymentDoc.BranchNames, "Branch name invalid")
			assert.Equal(t, "en", *paymentDoc.BranchNames[0].Code, "Branch name code should be en")
			assert.Equal(t, int8(50), paymentDoc.TransFlag, "Trans Flag should be 50")
			assert.Equal(t, "c01", paymentDoc.CustCode, "Cust code should be c01")
			assert.Equal(t, "ลูกค้าทดสอบ", *paymentDoc.CustNames[0].Name, "Cust name should be ลูกค้าทดสอบ")
		})
	}

}

func TestParseSaleInvoice(t *testing.T) {

	rawData := saleInvoiceRawData()
	transMQDoc := transmodels.TransactionMessageQueue{}

	err := json.Unmarshal([]byte(rawData), &transMQDoc)

	require.NoError(t, err, "Error should be nil")

	paymentDoc, err := usecase.ParseTransactionToPayment(transMQDoc)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "00000", paymentDoc.BranchCode, "Branch code invalid")
	assert.Equal(t, "สำนักงานใหญ่", *paymentDoc.BranchNames[0].Name, "Branch name invalid")
	assert.Equal(t, "th", *paymentDoc.BranchNames[0].Code, "Branch name code should be en")
	assert.Equal(t, int8(1), paymentDoc.DocType, "Doc type should be 1")

	assert.Equal(t, int8(44), paymentDoc.TransFlag)
	assert.Equal(t, "c1", paymentDoc.CustCode)

}

func saleInvoiceRawData() string {
	return `{"shopid":"2PrIIqTWxoBXv16K310sNwfHmfY","guidfixed":"2c4AhuOukBF3VCHdTxCVRO47zbw","parid":"","docno":"SI2024020800001","docdatetime":"2024-02-08T02:49:28.012Z","guidref":"4d4b5035-8ffb-4813-8c50-7c1da8ca6005","transflag":44,"docreftype":0,"docrefno":"","docrefdate":"2024-02-08T02:49:28.012Z","taxdocdate":"2024-02-08T02:49:28.012Z","taxdocno":"SI2024020800001","doctype":1,"inquirytype":1,"vattype":1,"vatrate":7,"custcode":"c1","custnames":[{"code":"th","name": "ลูกค้า 1"}],"description":"","discountword":"15","totaldiscount":15,"totalvalue":170,"totalexceptvat":47.06,"totalaftervat":112.94,"totalbeforevat":105.55140186915888,"totalvatvalue":7.38859813084112,"totalamount":145,"totalcost":0,"posid":"","cashiercode":"","salecode":"SALE001","salename":"นาย ขายดี","membercode":"","iscancel":false,"ismanualamount":false,"status":0,"paymentdetail":{"cashamounttext":"","cashamount":0,"paymentcreditcards":[],"paymenttransfers":[]},"paymentdetailraw":"[]","paycashamount":145,"branch":{"guidfixed":"2Prp2MbDKqpDBAgSYBtqbVXODwT","code":"00000","names":[{"code":"th","name":"สำนักงานใหญ่","isauto":false,"isdelete":false},{"code":"en","name":"","isauto":false,"isdelete":false},{"code":"lo","name":"","isauto":false,"isdelete":false},{"code":"ja","name":"","isauto":false,"isdelete":false}]},"billtaxtype":0,"canceldatetime":"","cancelusercode":"","cancelusername":"","canceldescription":"","cancelreason":"","fullvataddress":"","fullvatbranchnumber":"","fullvatname":"","fullvatdocnumber":"","fullvattaxid":"","fullvatprint":false,"isvatregister":false,"printcopybilldatetime":[],"tablenumber":"","tableopendatetime":"","tableclosedatetime":"","mancount":0,"womancount":0,"childcount":0,"istableallacratemode":false,"buffetcode":"","customertelephone":"","totalqty":15,"totaldiscountvatamount":7.06,"totaldiscountexceptvatamount":2.94,"cashiername":"","paycashchange":0,"sumqrcode":0,"sumcreditcard":0,"summoneytransfer":0,"sumcheque":0,"sumcoupon":0,"detaildiscountformula":"10","detailtotalamount":160,"detailtotaldiscount":10,"roundamount":0,"totalamountafterdiscount":145,"detailtotalamountbeforediscount":0,"sumcredit":0,"details":[{"manufacturerguid":"2QK6szWFzrb43310Dq80vGcum7q","inquirytype":1,"linenumber":0,"docdatetime":"2024-02-08T02:49:36.564Z","docref":"","docrefdatetime":"2024-02-08T02:49:36.564Z","calcflag":-1,"barcode":"BARCODE001","itemcode":"ITEM001","unitcode":"ENV","itemtype":0,"itemguid":"2PrfDoufKF7KF0Ua2V6sbHBlm2R","qty":10,"totalqty":10,"price":12,"discount":"0","discountamount":0,"totalvaluevat":7.850467289719626,"priceexcludevat":11.214953271028037,"sumamount":120,"sumamountexcludevat":112.14953271028037,"dividevalue":1,"standvalue":1,"vattype":1,"remark":"","multiunit":true,"sumofcost":0,"averagecost":0,"laststatus":0,"ispos":0,"taxtype":0,"vatcal":0,"whcode":"00000","shelfcode":"","locationcode":"","towhcode":"00000","tolocationcode":"","itemnames":[{"code":"th","name":"มาม่าา","isauto":false,"isdelete":false},{"code":"en","name":"mama","isauto":false,"isdelete":false},{"code":"ja","name":"","isauto":false,"isdelete":false},{"code":"ko","name":"","isauto":false,"isdelete":false},{"code":"lo","name":"","isauto":false,"isdelete":false}],"unitnames":[{"code":"th","name":"ซอง","isauto":false,"isdelete":false},{"code":"en","name":"Envelope","isauto":false,"isdelete":false},{"code":"ja","name":"","isauto":false,"isdelete":false}],"whnames":[{"code":"th","name":"คลังสำนักงานใหญ่","isauto":false,"isdelete":false},{"code":"en","name":"Headquarters","isauto":false,"isdelete":false},{"code":"lo","name":"ສໍາ​ນັກ​ງານ​ໃຫຍ່","isauto":false,"isdelete":false},{"code":"ja","name":"本部","isauto":false,"isdelete":false}],"locationnames":[],"towhnames":[{"code":"th","name":"คลังสำนักงานใหญ่","isauto":false,"isdelete":false},{"code":"en","name":"Headquarters","isauto":false,"isdelete":false},{"code":"lo","name":"ສໍາ​ນັກ​ງານ​ໃຫຍ່","isauto":false,"isdelete":false},{"code":"ja","name":"本部","isauto":false,"isdelete":false}],"tolocationnames":[],"sku":"","extrajson":""},{"manufacturerguid":"2QK6szWFzrb43310Dq80vGcum7q","inquirytype":1,"linenumber":0,"docdatetime":"2024-02-08T02:49:39.814Z","docref":"","docrefdatetime":"2024-02-08T02:49:39.815Z","calcflag":-1,"barcode":"BARCODE008","itemcode":"ITEM003","unitcode":"KG","itemtype":0,"itemguid":"2PrsLY9QvcbPEEUX2qQKvQboK7w","qty":5,"totalqty":5,"price":10,"discount":"","discountamount":0,"totalvaluevat":0,"priceexcludevat":10,"sumamount":50,"sumamountexcludevat":50,"dividevalue":1,"standvalue":1,"vattype":1,"remark":"","multiunit":true,"sumofcost":0,"averagecost":0,"laststatus":0,"ispos":0,"taxtype":0,"vatcal":1,"whcode":"00000","shelfcode":"","locationcode":"","towhcode":"00000","tolocationcode":"","itemnames":[{"code":"th","name":"ข้าวสาร","isauto":false,"isdelete":false},{"code":"en","name":"","isauto":false,"isdelete":false},{"code":"ja","name":"","isauto":false,"isdelete":false}],"unitnames":[{"code":"th","name":"กิโลกรัม","isauto":false,"isdelete":false},{"code":"en","name":"Kilogram","isauto":false,"isdelete":false},{"code":"ja","name":"","isauto":false,"isdelete":false}],"whnames":[{"code":"th","name":"คลังสำนักงานใหญ่","isauto":false,"isdelete":false},{"code":"en","name":"Headquarters","isauto":false,"isdelete":false},{"code":"lo","name":"ສໍາ​ນັກ​ງານ​ໃຫຍ່","isauto":false,"isdelete":false},{"code":"ja","name":"本部","isauto":false,"isdelete":false}],"locationnames":[],"towhnames":[{"code":"th","name":"คลังสำนักงานใหญ่","isauto":false,"isdelete":false},{"code":"en","name":"Headquarters","isauto":false,"isdelete":false},{"code":"lo","name":"ສໍາ​ນັກ​ງານ​ໃຫຍ່","isauto":false,"isdelete":false},{"code":"ja","name":"本部","isauto":false,"isdelete":false}],"tolocationnames":[],"sku":"","extrajson":""}],"ispos":false,"couponno":"","couponamount":0,"coupondescription":"","qrcode":"","qrcodeamount":0,"chequeno":"","chequebooknumber":"","chequebookcode":"","chequeduedate":"","chequeamount":0,"salechannelcode":"","salechannelgp":0,"salechannelgptype":0,"takeaway":0,"createdby":"smlsoftdev@gmail.com","createdat":"2024-02-08T02:49:59.197Z","updatedat":"2024-02-08T03:10:53.061Z","updatedby":"smlsoftdev@gmail.com"}`
}
