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

	giveJsonInput := `{"guidfixed":"","docno":"IB2025072300002","docdatetime":"2025-07-23T08:45:07.452Z","guidref":"3537b981-baa3-4bfc-9873-fa9420a5fef8","shiftdocno":"","devicename":"","guidpos":"","transflag":54,"docreftype":0,"docrefno":"","docrefdate":"2025-07-23T08:45:07.452Z","taxdocdate":"2025-07-23T08:45:07.452Z","taxdocno":"","doctype":0,"imageurl":"","inquirytype":0,"vattype":0,"vatrate":7,"custcode":"","custnames":[],"getpoint":0,"usepoint":0,"pointdiscountamount":0,"description":"","discountword":"","totaldiscount":0,"totalvalue":0,"totalexceptvat":0,"totalaftervat":0,"totalbeforevat":0,"totalvatvalue":0,"totalamount":132,"totalcost":0,"posid":"","cashiercode":"","salecode":"","salename":"","membercode":"","iscancel":false,"ismanualamount":false,"status":0,"paymentdetail":{"cashamounttext":"","cashamount":0,"paymentcreditcards":[],"paymenttransfers":[]},"paymentdetailraw":"","paycashamount":0,"paypointamount":0,"branch":{"guidfixed":"","code":"","names":[]},"billtaxtype":0,"canceldatetime":"","cancelusercode":"","cancelusername":"","canceldescription":"","cancelreason":"","fullvataddress":"","fullvatbranchnumber":"","fullvatname":"","fullvatdocnumber":"","fullvattaxid":"","fullvatprint":false,"isvatregister":false,"printcopybilldatetime":[],"tablenumber":"","tableopendatetime":"","tableclosedatetime":"","mancount":0,"womancount":0,"childcount":0,"istableallacratemode":false,"buffetcode":"","customertelephone":"","totalqty":2,"totaldiscountvatamount":0,"totaldiscountexceptvatamount":0,"cashiername":"","paycashchange":0,"sumqrcode":0,"sumcreditcard":0,"summoneytransfer":0,"sumcheque":0,"sumcoupon":0,"coupons":null,"totalcouponamount":0,"coupondiscountamount":0,"couponcashamount":0,"detaildiscountformula":"","detailtotalamount":0,"detailtotaldiscount":0,"roundamount":0,"totalamountafterdiscount":0,"detailtotalamountbeforediscount":0,"sumcredit":0,"shopid":"2QJPo41eNKMAVXzvS2e55TwyBwN","createdby":"smlsoftdev@gmail.com","createdat":"2025-07-23T08:45:36.232534525Z","updatedby":"","updatedat":"0001-01-01T00:00:00Z","deletedby":"","deletedat":"0001-01-01T00:00:00Z","details":[{"inquirytype":0,"linenumber":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"IB2507231545E381","docrefdatetime":"2025-07-23T08:45:07.452Z","calcflag":0,"barcode":"885002","itemcode":"","itemnames":[{"code":"th","name":"มาม่า","isauto":false,"isdelete":false}],"unitcode":"PACK","unitnames":null,"itemtype":0,"itemguid":"","imageurl":"","description":"","qty":5,"totalqty":0,"price":12,"discount":"","discountamount":0,"totalvaluevat":60,"priceexcludevat":12,"sumamount":60,"sumamountexcludevat":60,"refguid":"","dividevalue":1,"standvalue":1,"vattype":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"taxtype":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","groupcode":"","groupnames":null,"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0},{"inquirytype":0,"linenumber":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"IB2507231545E381","docrefdatetime":"2025-07-23T08:45:07.452Z","calcflag":0,"barcode":"885004","itemcode":"","itemnames":[{"code":"th","name":"น้ำเปล่า","isauto":false,"isdelete":false}],"unitcode":"PC","unitnames":null,"itemtype":0,"itemguid":"","imageurl":"","description":"","qty":24,"totalqty":0,"price":3,"discount":"","discountamount":0,"totalvaluevat":72,"priceexcludevat":3,"sumamount":72,"sumamountexcludevat":72,"refguid":"","dividevalue":1,"standvalue":1,"vattype":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"taxtype":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","groupcode":"","groupnames":null,"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0}]}`
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

	giveJsonInput := `[{"guidfixed":"2dz2dGiKZ1goegJeSTuMlPC4yRx","docno":"IB2024022900001","docdatetime":"2024-02-29T04:50:54.054Z","guidref":"6e89a614-8584-41a3-bfdb-968091242d46","shiftdocno":"","devicename":"","guidpos":"","transflag":54,"docreftype":0,"docrefno":"","docrefdate":"2024-03-21T04:50:47.755Z","taxdocdate":"2024-03-21T04:50:47.755Z","taxdocno":"","doctype":0,"imageurl":"","inquirytype":0,"vattype":0,"vatrate":7,"custcode":"","custnames":[],"getpoint":0,"usepoint":0,"pointdiscountamount":0,"description":"","discountword":"","totaldiscount":0,"totalvalue":0,"totalexceptvat":0,"totalaftervat":0,"totalbeforevat":0,"totalvatvalue":0,"totalamount":1000,"totalcost":0,"posid":"","cashiercode":"","salecode":"","salename":"","membercode":"","iscancel":false,"ismanualamount":false,"status":0,"paymentdetail":{"cashamounttext":"","cashamount":0,"paymentcreditcards":[],"paymenttransfers":[]},"paymentdetailraw":"","paycashamount":0,"paypointamount":0,"branch":{"guidfixed":"","code":"","names":[]},"billtaxtype":0,"canceldatetime":"","cancelusercode":"","cancelusername":"","canceldescription":"","cancelreason":"","fullvataddress":"","fullvatbranchnumber":"","fullvatname":"","fullvatdocnumber":"","fullvattaxid":"","fullvatprint":false,"isvatregister":false,"printcopybilldatetime":[],"tablenumber":"","tableopendatetime":"","tableclosedatetime":"","mancount":0,"womancount":0,"childcount":0,"istableallacratemode":false,"buffetcode":"","customertelephone":"","totalqty":1,"totaldiscountvatamount":0,"totaldiscountexceptvatamount":0,"cashiername":"","paycashchange":0,"sumqrcode":0,"sumcreditcard":0,"summoneytransfer":0,"sumcheque":0,"sumcoupon":0,"coupons":null,"totalcouponamount":0,"coupondiscountamount":0,"couponcashamount":0,"detaildiscountformula":"","detailtotalamount":0,"detailtotaldiscount":0,"roundamount":0,"totalamountafterdiscount":0,"detailtotalamountbeforediscount":0,"sumcredit":0,"shopid":"2QJPo41eNKMAVXzvS2e55TwyBwN","createdby":"","createdat":"0001-01-01T00:00:00Z","updatedby":"","updatedat":"0001-01-01T00:00:00Z","deletedby":"","deletedat":"0001-01-01T00:00:00Z","details":[{"inquirytype":0,"linenumber":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"","docrefdatetime":"0001-01-01T00:00:00Z","calcflag":0,"barcode":"885001","itemcode":"","itemnames":[{"code":"th","name":"มาม่า","isauto":false,"isdelete":false}],"unitcode":"PC","unitnames":[{"code":"th","name":"ชิ้น","isauto":false,"isdelete":false}],"itemtype":0,"itemguid":"2RYEPoU9FGve3ntUYWjlAGiuiwd","imageurl":"","description":"","qty":100,"totalqty":0,"price":10,"discount":"","discountamount":0,"totalvaluevat":0,"priceexcludevat":0,"sumamount":1000,"sumamountexcludevat":0,"refguid":"","dividevalue":1,"standvalue":1,"vattype":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"taxtype":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","groupcode":"","groupnames":[],"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0}]},{"guidfixed":"30GiVsTbYJS4lzJQ0QMTc8fLqtn","docno":"IB2025072300002","docdatetime":"2025-07-23T08:45:07.452Z","guidref":"3537b981-baa3-4bfc-9873-fa9420a5fef8","shiftdocno":"","devicename":"","guidpos":"","transflag":54,"docreftype":0,"docrefno":"","docrefdate":"2025-07-23T08:45:07.452Z","taxdocdate":"2025-07-23T08:45:07.452Z","taxdocno":"","doctype":0,"imageurl":"","inquirytype":0,"vattype":0,"vatrate":7,"custcode":"","custnames":[],"getpoint":0,"usepoint":0,"pointdiscountamount":0,"description":"","discountword":"","totaldiscount":0,"totalvalue":0,"totalexceptvat":0,"totalaftervat":0,"totalbeforevat":0,"totalvatvalue":0,"totalamount":132,"totalcost":0,"posid":"","cashiercode":"","salecode":"","salename":"","membercode":"","iscancel":false,"ismanualamount":false,"status":0,"paymentdetail":{"cashamounttext":"","cashamount":0,"paymentcreditcards":[],"paymenttransfers":[]},"paymentdetailraw":"","paycashamount":0,"paypointamount":0,"branch":{"guidfixed":"","code":"","names":[]},"billtaxtype":0,"canceldatetime":"","cancelusercode":"","cancelusername":"","canceldescription":"","cancelreason":"","fullvataddress":"","fullvatbranchnumber":"","fullvatname":"","fullvatdocnumber":"","fullvattaxid":"","fullvatprint":false,"isvatregister":false,"printcopybilldatetime":[],"tablenumber":"","tableopendatetime":"","tableclosedatetime":"","mancount":0,"womancount":0,"childcount":0,"istableallacratemode":false,"buffetcode":"","customertelephone":"","totalqty":2,"totaldiscountvatamount":0,"totaldiscountexceptvatamount":0,"cashiername":"","paycashchange":0,"sumqrcode":0,"sumcreditcard":0,"summoneytransfer":0,"sumcheque":0,"sumcoupon":0,"coupons":null,"totalcouponamount":0,"coupondiscountamount":0,"couponcashamount":0,"detaildiscountformula":"","detailtotalamount":0,"detailtotaldiscount":0,"roundamount":0,"totalamountafterdiscount":0,"detailtotalamountbeforediscount":0,"sumcredit":0,"shopid":"2QJPo41eNKMAVXzvS2e55TwyBwN","createdby":"","createdat":"0001-01-01T00:00:00Z","updatedby":"","updatedat":"0001-01-01T00:00:00Z","deletedby":"","deletedat":"0001-01-01T00:00:00Z","details":[{"inquirytype":0,"linenumber":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"","docrefdatetime":"0001-01-01T00:00:00Z","calcflag":0,"barcode":"885002","itemcode":"","itemnames":[{"code":"th","name":"มาม่า","isauto":false,"isdelete":false}],"unitcode":"PACK","unitnames":[{"code":"th","name":"แพ็ค","isauto":false,"isdelete":false}],"itemtype":0,"itemguid":"2RYETslkHYro193W34GVeJe3q5i","imageurl":"","description":"","qty":5,"totalqty":0,"price":12,"discount":"","discountamount":0,"totalvaluevat":0,"priceexcludevat":0,"sumamount":60,"sumamountexcludevat":0,"refguid":"","dividevalue":1,"standvalue":1,"vattype":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"taxtype":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","groupcode":"","groupnames":[],"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0},{"inquirytype":0,"linenumber":0,"docdatetime":"0001-01-01T00:00:00Z","docref":"","docrefdatetime":"0001-01-01T00:00:00Z","calcflag":0,"barcode":"885004","itemcode":"","itemnames":[{"code":"th","name":"น้ำเปล่า","isauto":false,"isdelete":false}],"unitcode":"PC","unitnames":[{"code":"th","name":"ชิ้น","isauto":false,"isdelete":false}],"itemtype":0,"itemguid":"2dGOL7I92zVMXYosFMFauTSFN7P","imageurl":"","description":"","qty":24,"totalqty":0,"price":3,"discount":"","discountamount":0,"totalvaluevat":0,"priceexcludevat":0,"sumamount":72,"sumamountexcludevat":0,"refguid":"","dividevalue":1,"standvalue":1,"vattype":0,"remark":"","multiunit":false,"issumpoint":false,"sumofcost":0,"averagecost":0,"foodtype":0,"laststatus":0,"ischoice":0,"ispos":0,"taxtype":0,"vatcal":0,"whcode":"","whnames":null,"shelfcode":"","locationcode":"","locationnames":null,"towhcode":"","towhnames":null,"tolocationcode":"","tolocationnames":null,"sku":"","extrajson":"","groupcode":"","groupnames":[],"manufacturerguid":"","manufacturercode":"","manufacturernames":null,"sumamountchoice":0}]}]`

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
