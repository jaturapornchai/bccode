package stockbalance_test

import (
	"mime/multipart"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/transaction/transactionconsumer/stockbalance"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

type MockHttpContext struct {
	mock.Mock
}

// Log implementation
func (m *MockHttpContext) Log(message string) {
	m.Called(message)
}

// UserInfo implementation
func (m *MockHttpContext) UserInfo() models.UserInfo {
	args := m.Called()
	return args.Get(0).(models.UserInfo)
}

// Header implementation
func (m *MockHttpContext) Header(attribute string) string {
	args := m.Called(attribute)
	return args.String(0)
}

// RealIp implementation
func (m *MockHttpContext) RealIp() string {
	args := m.Called()
	return args.String(0)
}

// Param implementation
func (m *MockHttpContext) Param(name string) string {
	args := m.Called(name)
	return args.String(0)
}

// QueryParam implementation
func (m *MockHttpContext) QueryParam(name string) string {
	args := m.Called(name)
	return args.String(0)
}

// ReadInput implementation
func (m *MockHttpContext) ReadInput() string {
	args := m.Called()
	return args.String(0)
}

// Response implementation
func (m *MockHttpContext) Response(responseCode int, responseData interface{}) {
	m.Called(responseCode, responseData)
}

// ResponseError implementation
func (m *MockHttpContext) ResponseError(responseCode int, errorMessage string) {
	m.Called(responseCode, errorMessage)
}

// Validate implementation
func (m *MockHttpContext) Validate(model interface{}) error {
	args := m.Called(model)
	return args.Error(0)
}

// FormFile implementation
func (m *MockHttpContext) FormFile(field string) (*multipart.FileHeader, error) {
	args := m.Called(field)
	return args.Get(0).(*multipart.FileHeader), args.Error(1)
}

// FormValue implementation
func (m *MockHttpContext) FormValue(field string) string {
	args := m.Called(field)
	return args.String(0)
}

// ResponseWriter implementation
func (m *MockHttpContext) ResponseWriter() http.ResponseWriter {
	args := m.Called()
	return args.Get(0).(http.ResponseWriter)
}

// Request implementation
func (m *MockHttpContext) Request() *http.Request {
	args := m.Called()
	return args.Get(0).(*http.Request)
}

// Persister implementation
func (m *MockHttpContext) Persister(cfg config.IPersisterConfig) microservice.IPersister {
	args := m.Called(cfg)
	return args.Get(0).(microservice.IPersister)
}

// Cacher implementation
func (m *MockHttpContext) Cacher(cacherConfig config.ICacherConfig) microservice.ICacher {
	args := m.Called(cacherConfig)
	return args.Get(0).(microservice.ICacher)
}

// Producer implementation
func (m *MockHttpContext) Producer(servers config.IMQConfig) microservice.IProducer {
	args := m.Called(servers)
	return args.Get(0).(microservice.IProducer)
}

// MQ implementation
func (m *MockHttpContext) MQ(servers config.IMQConfig) microservice.IMQ {
	args := m.Called(servers)
	return args.Get(0).(microservice.IMQ)
}

// EchoContext implementation
func (m *MockHttpContext) EchoContext() echo.Context {
	args := m.Called()
	return args.Get(0).(echo.Context)
}

func TestStockBalanceConsumerCreate(t *testing.T) {

	giveJsonInput := `{"guid_fixed":"","docno":"IB2025072300002","docdatetime":"2025-07-23T08:45:07.452Z","guid_ref":"3537b981-baa3-4bfc-9873-fa9420a5fef8","shiftdocno":"","devicename":"","guidpos":"","transflag":54,"docreftype":0,"docrefno":"","docrefdate":"2025-07-23T08:45:07.452Z","taxdocdate":"2025-07-23T08:45:07.452Z","taxdocno":"","doc_type":0,"imageurl":"","inquirytype":0,"vat_type":0,"vatrate":7,"custcode":"","custnames":[],"getpoint":0,"usepoint":0,"pointdiscountamount":0,"description":"","discountword":"","totaldiscount":0,"totalvalue":0,"totalexceptvat":0,"totalaftervat":0,"totalbeforevat":0,"totalvatvalue":0,"total_amount":132,"total_cost":0,"posid":"","cashiercode":"","salecode":"","salename":"","membercode":"","iscancel":false,"ismanualamount":false,"status":0,"paymentdetail":{"cashamounttext":"","cashamount":0,"paymentcreditcards":[],"paymenttransfers":[]},"paymentdetailraw":"","paycashamount":0,"paypointamount":0,"branch":{"guid_fixed":"","code":"","names":[]},"billtaxtype":0,"canceldatetime":"","cancelusercode":"","cancelusername":"","canceldescription":"","cancelreason":"","fullvataddress":"","fullvatbranchnumber":"","fullvatname":"","fullvatdocnumber":"","fullvattaxid":"","fullvatprint":false,"isvatregister":false,"printcopybilldatetime":[],"tablenumber":"","tableopendatetime":"","tableclosedatetime":"","mancount":0,"womancount":0,"childcount":0,"istableallacratemode":false,"buffetcode":"","customertelephone":"","totalqty":2,"totaldiscountvatamount":0,"totaldiscountexceptvatamount":0,"cashiername":"","paycashchange":0,"sumqrcode":0,"sumcreditcard":0,"summoneytransfer":0,"sumcheque":0,"sumcoupon":0,"coupons":null,"totalcouponamount":0,"coupondiscountamount":0,"couponcashamount":0,"detaildiscountformula":"","detailtotalamount":0,"detailtotaldiscount":0,"roundamount":0,"totalamountafterdiscount":0,"detailtotalamountbeforediscount":0,"sumcredit":0,"shopid":"2QJPo41eNKMAVXzvS2e55TwyBwN","createdby":"smlsoftdev@gmail.com","created_at":"2025-07-23T08:45:36.232534525Z","updatedby":"","updated_at":"0001-01-01T00:00:00Z","deleted_by":"","deleted_at":"0001-01-01T00:00:00Z","details":[{"inquirytype":0,"line_number":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"IB2507231545E381","docrefdatetime":"2025-07-23T08:45:07.452Z","calcflag":0,"barcode":"885002","itemcode":"","itemnames":[{"code":"th","name":"มาม่า","isauto":false,"isdelete":false}],"unitcode":"PACK","unitnames":null,"item_type":0,"item_guid":"","imageurl":"","description":"","qty":5,"totalqty":0,"price":12,"discount":"","discountamount":0,"totalvaluevat":60,"priceexcludevat":12,"sum_amount":60,"sumamountexcludevat":60,"refguid":"","dividevalue":1,"standvalue":1,"vat_type":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"tax_type":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","group_code":"","group_names":null,"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0},{"inquirytype":0,"line_number":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"IB2507231545E381","docrefdatetime":"2025-07-23T08:45:07.452Z","calcflag":0,"barcode":"885004","itemcode":"","itemnames":[{"code":"th","name":"น้ำเปล่า","isauto":false,"isdelete":false}],"unitcode":"PC","unitnames":null,"item_type":0,"item_guid":"","imageurl":"","description":"","qty":24,"totalqty":0,"price":3,"discount":"","discountamount":0,"totalvaluevat":72,"priceexcludevat":3,"sum_amount":72,"sumamountexcludevat":72,"refguid":"","dividevalue":1,"standvalue":1,"vat_type":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"tax_type":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","group_code":"","group_names":null,"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0}]}`
	giveMockHttpContext := &MockHttpContext{}
	giveMockHttpContext.On("ReadInput").Return(giveJsonInput)

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		t.Error("Failed to create microservice:", err)
	}

	stcokBalanceConsumer := stockbalance.InitStockReceiveTransactionConsumer(ms, cfg)

	stcokBalanceConsumer.ConsumeOnCreateOrUpdate(giveMockHttpContext)

	ms.Cleanup()
}

func TestStockBalanceConsumerBulkCreate(t *testing.T) {

	giveJsonInput := `[{"guid_fixed":"2dz2dGiKZ1goegJeSTuMlPC4yRx","docno":"IB2024022900001","docdatetime":"2024-02-29T04:50:54.054Z","guid_ref":"6e89a614-8584-41a3-bfdb-968091242d46","shiftdocno":"","devicename":"","guidpos":"","transflag":54,"docreftype":0,"docrefno":"","docrefdate":"2024-03-21T04:50:47.755Z","taxdocdate":"2024-03-21T04:50:47.755Z","taxdocno":"","doc_type":0,"imageurl":"","inquirytype":0,"vat_type":0,"vatrate":7,"custcode":"","custnames":[],"getpoint":0,"usepoint":0,"pointdiscountamount":0,"description":"","discountword":"","totaldiscount":0,"totalvalue":0,"totalexceptvat":0,"totalaftervat":0,"totalbeforevat":0,"totalvatvalue":0,"total_amount":1000,"total_cost":0,"posid":"","cashiercode":"","salecode":"","salename":"","membercode":"","iscancel":false,"ismanualamount":false,"status":0,"paymentdetail":{"cashamounttext":"","cashamount":0,"paymentcreditcards":[],"paymenttransfers":[]},"paymentdetailraw":"","paycashamount":0,"paypointamount":0,"branch":{"guid_fixed":"","code":"","names":[]},"billtaxtype":0,"canceldatetime":"","cancelusercode":"","cancelusername":"","canceldescription":"","cancelreason":"","fullvataddress":"","fullvatbranchnumber":"","fullvatname":"","fullvatdocnumber":"","fullvattaxid":"","fullvatprint":false,"isvatregister":false,"printcopybilldatetime":[],"tablenumber":"","tableopendatetime":"","tableclosedatetime":"","mancount":0,"womancount":0,"childcount":0,"istableallacratemode":false,"buffetcode":"","customertelephone":"","totalqty":1,"totaldiscountvatamount":0,"totaldiscountexceptvatamount":0,"cashiername":"","paycashchange":0,"sumqrcode":0,"sumcreditcard":0,"summoneytransfer":0,"sumcheque":0,"sumcoupon":0,"coupons":null,"totalcouponamount":0,"coupondiscountamount":0,"couponcashamount":0,"detaildiscountformula":"","detailtotalamount":0,"detailtotaldiscount":0,"roundamount":0,"totalamountafterdiscount":0,"detailtotalamountbeforediscount":0,"sumcredit":0,"shopid":"2QJPo41eNKMAVXzvS2e55TwyBwN","createdby":"","created_at":"0001-01-01T00:00:00Z","updatedby":"","updated_at":"0001-01-01T00:00:00Z","deleted_by":"","deleted_at":"0001-01-01T00:00:00Z","details":[{"inquirytype":0,"line_number":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"","docrefdatetime":"0001-01-01T00:00:00Z","calcflag":0,"barcode":"885001","itemcode":"","itemnames":[{"code":"th","name":"มาม่า","isauto":false,"isdelete":false}],"unitcode":"PC","unitnames":[{"code":"th","name":"ชิ้น","isauto":false,"isdelete":false}],"item_type":0,"item_guid":"2RYEPoU9FGve3ntUYWjlAGiuiwd","imageurl":"","description":"","qty":100,"totalqty":0,"price":10,"discount":"","discountamount":0,"totalvaluevat":0,"priceexcludevat":0,"sum_amount":1000,"sumamountexcludevat":0,"refguid":"","dividevalue":1,"standvalue":1,"vat_type":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"tax_type":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","group_code":"","group_names":[],"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0}]},{"guid_fixed":"30GiVsTbYJS4lzJQ0QMTc8fLqtn","docno":"IB2025072300002","docdatetime":"2025-07-23T08:45:07.452Z","guid_ref":"3537b981-baa3-4bfc-9873-fa9420a5fef8","shiftdocno":"","devicename":"","guidpos":"","transflag":54,"docreftype":0,"docrefno":"","docrefdate":"2025-07-23T08:45:07.452Z","taxdocdate":"2025-07-23T08:45:07.452Z","taxdocno":"","doc_type":0,"imageurl":"","inquirytype":0,"vat_type":0,"vatrate":7,"custcode":"","custnames":[],"getpoint":0,"usepoint":0,"pointdiscountamount":0,"description":"","discountword":"","totaldiscount":0,"totalvalue":0,"totalexceptvat":0,"totalaftervat":0,"totalbeforevat":0,"totalvatvalue":0,"total_amount":132,"total_cost":0,"posid":"","cashiercode":"","salecode":"","salename":"","membercode":"","iscancel":false,"ismanualamount":false,"status":0,"paymentdetail":{"cashamounttext":"","cashamount":0,"paymentcreditcards":[],"paymenttransfers":[]},"paymentdetailraw":"","paycashamount":0,"paypointamount":0,"branch":{"guid_fixed":"","code":"","names":[]},"billtaxtype":0,"canceldatetime":"","cancelusercode":"","cancelusername":"","canceldescription":"","cancelreason":"","fullvataddress":"","fullvatbranchnumber":"","fullvatname":"","fullvatdocnumber":"","fullvattaxid":"","fullvatprint":false,"isvatregister":false,"printcopybilldatetime":[],"tablenumber":"","tableopendatetime":"","tableclosedatetime":"","mancount":0,"womancount":0,"childcount":0,"istableallacratemode":false,"buffetcode":"","customertelephone":"","totalqty":2,"totaldiscountvatamount":0,"totaldiscountexceptvatamount":0,"cashiername":"","paycashchange":0,"sumqrcode":0,"sumcreditcard":0,"summoneytransfer":0,"sumcheque":0,"sumcoupon":0,"coupons":null,"totalcouponamount":0,"coupondiscountamount":0,"couponcashamount":0,"detaildiscountformula":"","detailtotalamount":0,"detailtotaldiscount":0,"roundamount":0,"totalamountafterdiscount":0,"detailtotalamountbeforediscount":0,"sumcredit":0,"shopid":"2QJPo41eNKMAVXzvS2e55TwyBwN","createdby":"","created_at":"0001-01-01T00:00:00Z","updatedby":"","updated_at":"0001-01-01T00:00:00Z","deleted_by":"","deleted_at":"0001-01-01T00:00:00Z","details":[{"inquirytype":0,"line_number":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"","docrefdatetime":"0001-01-01T00:00:00Z","calcflag":0,"barcode":"885002","itemcode":"","itemnames":[{"code":"th","name":"มาม่า","isauto":false,"isdelete":false}],"unitcode":"PACK","unitnames":[{"code":"th","name":"แพ็ค","isauto":false,"isdelete":false}],"item_type":0,"item_guid":"2RYETslkHYro193W34GVeJe3q5i","imageurl":"","description":"","qty":5,"totalqty":0,"price":12,"discount":"","discountamount":0,"totalvaluevat":0,"priceexcludevat":0,"sum_amount":60,"sumamountexcludevat":0,"refguid":"","dividevalue":1,"standvalue":1,"vat_type":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"tax_type":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","group_code":"","group_names":[],"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0},{"inquirytype":0,"line_number":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"","docrefdatetime":"0001-01-01T00:00:00Z","calcflag":0,"barcode":"885004","itemcode":"","itemnames":[{"code":"th","name":"น้ำเปล่า","isauto":false,"isdelete":false}],"unitcode":"PC","unitnames":[{"code":"th","name":"ชิ้น","isauto":false,"isdelete":false}],"item_type":0,"item_guid":"2dGOL7I92zVMXYosFMFauTSFN7P","imageurl":"","description":"","qty":24,"totalqty":0,"price":3,"discount":"","discountamount":0,"totalvaluevat":0,"priceexcludevat":0,"sum_amount":72,"sumamountexcludevat":0,"refguid":"","dividevalue":1,"standvalue":1,"vat_type":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"tax_type":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","group_code":"","group_names":[],"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0}]}]`

	giveMockHttpContext := &MockHttpContext{}
	giveMockHttpContext.On("ReadInput").Return(giveJsonInput)

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		t.Error("Failed to create microservice:", err)
	}

	stcokBalanceConsumer := stockbalance.InitStockReceiveTransactionConsumer(ms, cfg)

	stcokBalanceConsumer.ConsumeOnBulkCreateOrUpdate(giveMockHttpContext)

	ms.Cleanup()
}
