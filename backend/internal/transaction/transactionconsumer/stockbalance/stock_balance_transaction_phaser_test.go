package stockbalance_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/stockbalance"
	"testing"

	"github.com/tj/assert"
)

func TestPhaserSingDoc(t *testing.T) {

	giveJson := `{"guidfixed":"guidfixed001","docno":"IB2024012500008","docdatetime":"2024-01-25T11:38:48.048Z","guid_ref":"","transflag":0,"docreftype":0,"docrefno":"","docrefdate":"0001-01-01T00:00:00Z","branch":{"code":"branch01","names":[{"code":"th","name":"สาขาที่ 1","isauto":false,"isdelete":false}]},"taxdocdate":"0001-01-01T00:00:00Z","taxdocno":"","doc_type":0,"inquirytype":0,"vat_type":0,"vatrate":0,"custcode":"","custnames":null,"description":"","discountword":"","totaldiscount":0,"totalvalue":0,"totalexceptvat":0,"totalaftervat":0,"totalbeforevat":0,"totalvatvalue":0,"total_amount":0,"total_cost":0,"posid":"","cashiercode":"","salecode":"","salename":"","membercode":"","iscancel":false,"ismanualamount":false,"status":0,"paymentdetail":{"cashamounttext":"","cashamount":0,"paymentcreditcards":null,"paymenttransfers":null},"paymentdetailraw":"","paycashamount":0,"billtaxtype":0,"canceldatetime":"","cancelusercode":"","cancelusername":"","canceldescription":"","cancelreason":"","fullvataddress":"","fullvatbranchnumber":"","fullvatname":"","fullvatdocnumber":"","fullvattaxid":"","fullvatprint":false,"isvatregister":false,"printcopybilldatetime":null,"tablenumber":"","tableopendatetime":"","tableclosedatetime":"","mancount":0,"womancount":0,"childcount":0,"istableallacratemode":false,"buffetcode":"","customertelephone":"","totalqty":0,"totaldiscountvatamount":0,"totaldiscountexceptvatamount":0,"cashiername":"","paycashchange":0,"sumqrcode":0,"sumcreditcard":0,"summoneytransfer":0,"sumcheque":0,"sumcoupon":0,"detaildiscountformula":"","detailtotalamount":0,"detailtotaldiscount":0,"roundamount":0,"totalamountafterdiscount":0,"detailtotalamountbeforediscount":0,"sumcredit":0,"holdingcode":"2aWWshrYXunh7L4VZXmVQmOPXO5","createdby":"tester","createdat":"2024-01-25T11:40:48.210967+07:00","updatedby":"","updatedat":"0001-01-01T00:00:00Z","deletedby":"","deletedat":"0001-01-01T00:00:00Z","details":[{"inquirytype":0,"line_number":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"","docrefdatetime":"2024-01-25T11:38:48.048Z","calcflag":0,"barcode":"BAR0001","itemcode":"","unitcode":"PCS","item_type":0,"item_guid":"","qty":24,"totalqty":0,"price":10,"discount":"","discountamount":0,"totalvaluevat":240,"priceexcludevat":10,"sum_amount":240,"sumamountexcludevat":240,"dividevalue":0,"standvalue":0,"vat_type":0,"remark":"","multiunit":false,"sumofcost":0,"averagecost":0,"laststatus":0,"ispos":0,"tax_type":0,"vatcal":0,"whcode":"00000","shelfcode":"","locationcode":"00000","towhcode":"","tolocationcode":"","itemnames":[{"code":"TH","name":"สินค้า 2","isauto":false,"isdelete":false}],"unitnames":null,"whnames":null,"locationnames":null,"towhnames":null,"tolocationnames":null,"sku":"","extrajson":""},{"inquirytype":0,"line_number":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"","docrefdatetime":"2024-01-25T11:38:48.048Z","calcflag":0,"barcode":"BAR0002","itemcode":"","unitcode":"PCS","item_type":0,"item_guid":"","qty":144,"totalqty":0,"price":20,"discount":"","discountamount":0,"totalvaluevat":2880,"priceexcludevat":20,"sum_amount":2880,"sumamountexcludevat":2880,"dividevalue":0,"standvalue":0,"vat_type":0,"remark":"","multiunit":false,"sumofcost":0,"averagecost":0,"laststatus":0,"ispos":0,"tax_type":0,"vatcal":0,"whcode":"00000","shelfcode":"","locationcode":"00000","towhcode":"","tolocationcode":"","itemnames":[{"code":"TH","name":"สินค้า 2","isauto":false,"isdelete":false}],"unitnames":null,"whnames":null,"locationnames":null,"towhnames":null,"tolocationnames":null,"sku":"","extrajson":""}]}`

	phaser := stockbalance.StockBalanceTransactionPhaser{}

	got, err := phaser.PhaseSingleDoc(giveJson)

	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	assert.Equal(t, "2aWWshrYXunh7L4VZXmVQmOPXO5", got.HoldingCode)
	assert.Equal(t, "IB2024012500008", got.DocNo)

	assert.Equal(t, 2, len(*got.Items))

	branchNames := []pkgModels.NameX{
		*pkgModels.NewNameXWithCodeName("th", "สาขาที่ 1"),
	}

	// branch
	assert.Equal(t, "branch01", got.BranchCode)
	assert.Equal(t, pkgModels.JSONB(branchNames), got.BranchNames)

	assert.Equal(t, "guidfixed001", (*got.Items)[0].GuidFixed)
}
