package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	migrationAPI "smlcloudplatform/cmd/migrationapi/api"
	"smlcloudplatform/docs"
	"smlcloudplatform/internal/apikeyservice"
	"smlcloudplatform/internal/authentication"
	"smlcloudplatform/internal/channel/salechannel"
	"smlcloudplatform/internal/channel/transportchannel"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/coupon"
	coupon_database "smlcloudplatform/internal/coupon/database"
	coupon_repositories "smlcloudplatform/internal/coupon/repositories"
	coupon_services "smlcloudplatform/internal/coupon/services"
	"smlcloudplatform/internal/currency"
	"smlcloudplatform/internal/debtaccount/creditor"
	"smlcloudplatform/internal/debtaccount/creditorgroup"
	"smlcloudplatform/internal/debtaccount/customer"
	"smlcloudplatform/internal/debtaccount/customergroup"
	"smlcloudplatform/internal/debtaccount/debtor"
	"smlcloudplatform/internal/debtaccount/debtorgroup"
	"smlcloudplatform/internal/dimension"
	"smlcloudplatform/internal/documentwarehouse/documentimage"
	"smlcloudplatform/internal/filestatus"
	"smlcloudplatform/internal/form/formtemplate"
	"smlcloudplatform/internal/images"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/masterexpense"
	"smlcloudplatform/internal/masterincome"
	"smlcloudplatform/internal/mastersync"
	"smlcloudplatform/internal/media"
	"smlcloudplatform/internal/member"
	"smlcloudplatform/internal/migration"
	"smlcloudplatform/internal/notify"
	"smlcloudplatform/internal/ocr"
	order_device "smlcloudplatform/internal/order/device"
	order_setting "smlcloudplatform/internal/order/setting"
	"smlcloudplatform/internal/organization/branch"
	"smlcloudplatform/internal/organization/businesstype"
	"smlcloudplatform/internal/organization/company"
	"smlcloudplatform/internal/organization/costcenter"
	"smlcloudplatform/internal/organization/department"
	"smlcloudplatform/internal/organization/jobproject"
	"smlcloudplatform/internal/payment/bankmaster"
	"smlcloudplatform/internal/payment/bookbank"
	"smlcloudplatform/internal/payment/qrpayment"
	"smlcloudplatform/internal/paymentmaster"
	pickandpackdevice "smlcloudplatform/internal/pickandpack"
	pos_media "smlcloudplatform/internal/pos/media"
	pos_setting "smlcloudplatform/internal/pos/setting"
	"smlcloudplatform/internal/pos/shift"
	"smlcloudplatform/internal/pos/temp"
	"smlcloudplatform/internal/product/bom"
	"smlcloudplatform/internal/product/color"
	"smlcloudplatform/internal/product/eorder"
	"smlcloudplatform/internal/product/option"
	"smlcloudplatform/internal/product/optionpattern"
	"smlcloudplatform/internal/product/ordertype"
	products "smlcloudplatform/internal/product/product"
	"smlcloudplatform/internal/product/productbarcode"
	"smlcloudplatform/internal/product/productcategory"
	"smlcloudplatform/internal/product/productgroup"
	"smlcloudplatform/internal/product/producttype"
	"smlcloudplatform/internal/product/promotion"
	"smlcloudplatform/internal/product/unit"
	"smlcloudplatform/internal/productimport"
	"smlcloudplatform/internal/productsection/sectionbranch"
	"smlcloudplatform/internal/productsection/sectionbusinesstype"
	"smlcloudplatform/internal/productsection/sectiondepartment"
	"smlcloudplatform/internal/purchasetype"
	"smlcloudplatform/internal/restaurant/device"
	"smlcloudplatform/internal/restaurant/kitchen"
	"smlcloudplatform/internal/restaurant/printer"
	"smlcloudplatform/internal/restaurant/settings"
	"smlcloudplatform/internal/restaurant/staff"
	"smlcloudplatform/internal/restaurant/table"
	"smlcloudplatform/internal/restaurant/zone"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/shop/employee"
	"smlcloudplatform/internal/shopdesign/zonedesign"
	"smlcloudplatform/internal/slipimage"
	"smlcloudplatform/internal/smlaiproduct/brandproduct"
	"smlcloudplatform/internal/smlaiproduct/categoryproduct"
	"smlcloudplatform/internal/smlaiproduct/classproduct"
	"smlcloudplatform/internal/smlaiproduct/designproduct"
	"smlcloudplatform/internal/smlaiproduct/gradeproduct"
	"smlcloudplatform/internal/smlaiproduct/groupproduct"
	"smlcloudplatform/internal/smlaiproduct/groupsuboneproduct"
	"smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct"
	"smlcloudplatform/internal/smlaiproduct/modelproduct"
	"smlcloudplatform/internal/smlaiproduct/patternproduct"
	"smlcloudplatform/internal/smsreceive/smstransaction"
	"smlcloudplatform/internal/stockbalanceimport"
	"smlcloudplatform/internal/stockprocess"
	"smlcloudplatform/internal/systemadmin"
	"smlcloudplatform/internal/task"

	"smlcloudplatform/internal/transaction/accrualreceive"
	"smlcloudplatform/internal/transaction/advancepayment"
	"smlcloudplatform/internal/transaction/advancepaymentrefund"
	"smlcloudplatform/internal/transaction/banktransferrecord"
	"smlcloudplatform/internal/transaction/chequechange"
	"smlcloudplatform/internal/transaction/chequedeposit"
	"smlcloudplatform/internal/transaction/chequedisqualified"
	"smlcloudplatform/internal/transaction/chequepass"
	"smlcloudplatform/internal/transaction/chequepaymentchange"
	"smlcloudplatform/internal/transaction/chequepaymentdeposit"
	"smlcloudplatform/internal/transaction/chequepaymentdisqualified"
	"smlcloudplatform/internal/transaction/chequepaymentreturn"
	"smlcloudplatform/internal/transaction/chequerenew"
	"smlcloudplatform/internal/transaction/chequereturn"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal"
	"smlcloudplatform/internal/transaction/deposit"
	"smlcloudplatform/internal/transaction/depositrecord"
	"smlcloudplatform/internal/transaction/depositrefund"
	"smlcloudplatform/internal/transaction/documentformate"
	"smlcloudplatform/internal/transaction/paid"
	"smlcloudplatform/internal/transaction/paidadvance"
	"smlcloudplatform/internal/transaction/paidadvancerefund"
	"smlcloudplatform/internal/transaction/pay"
	"smlcloudplatform/internal/transaction/payment"
	"smlcloudplatform/internal/transaction/paymentdetail"
	"smlcloudplatform/internal/transaction/pickandpack"
	"smlcloudplatform/internal/transaction/purchase"
	"smlcloudplatform/internal/transaction/purchaseorder"
	"smlcloudplatform/internal/transaction/purchasepartial"
	"smlcloudplatform/internal/transaction/purchaserequisition"
	"smlcloudplatform/internal/transaction/purchasereturn"
	"smlcloudplatform/internal/transaction/quotation"
	"smlcloudplatform/internal/transaction/receivedeposit"
	"smlcloudplatform/internal/transaction/receivedepositrefund"
	"smlcloudplatform/internal/transaction/rfq"
	"smlcloudplatform/internal/transaction/saleinvoice"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice"
	"smlcloudplatform/internal/transaction/saleinvoicereturn"
	"smlcloudplatform/internal/transaction/saleorder"
	"smlcloudplatform/internal/transaction/smltransaction"
	"smlcloudplatform/internal/transaction/stockadjustment"
	"smlcloudplatform/internal/transaction/stockbalance"
	"smlcloudplatform/internal/transaction/stockbalancedetail"
	"smlcloudplatform/internal/transaction/stockpickupproduct"
	"smlcloudplatform/internal/transaction/stockreceiveproduct"
	"smlcloudplatform/internal/transaction/stockreturnproduct"
	"smlcloudplatform/internal/transaction/stocktransfer"
	"smlcloudplatform/internal/transaction/withdrawalrecord"
	"smlcloudplatform/internal/vfgl/accountgroup"
	"smlcloudplatform/internal/vfgl/accountperiodmaster"
	"smlcloudplatform/internal/vfgl/chartofaccount"
	"smlcloudplatform/internal/vfgl/journal"
	"smlcloudplatform/internal/vfgl/journalbook"
	"smlcloudplatform/internal/vfgl/journalreport"
	"smlcloudplatform/internal/warehouse"
	"smlcloudplatform/pkg/microservice"
	"time"

	"smlcloudplatform/internal/transaction/transactionconsumer"

	paid_consumer "smlcloudplatform/internal/transaction/transactionconsumer/paid"
	pay_consumer "smlcloudplatform/internal/transaction/transactionconsumer/pay"

	appurchasereceive_consumer "smlcloudplatform/internal/transaction/transactionconsumer/appurchasereceive"
	purchase_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchase"
	purchasedebitnote_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchasedebitnote"
	purchaseorder_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchaseorder"
	purchasereceive_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchasereceive"
	purchaserequisition_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchaserequisition"
	purchasereturn_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchasereturn"
	rfq_consumer "smlcloudplatform/internal/transaction/transactionconsumer/rfq"

	saledebitnote_consumer "smlcloudplatform/internal/transaction/transactionconsumer/saledebitnote"
	saleinvoice_consumer "smlcloudplatform/internal/transaction/transactionconsumer/saleinvoice"
	saleinvoicereutrn_consumer "smlcloudplatform/internal/transaction/transactionconsumer/saleinvoicereturn"
	saleorder_consumer "smlcloudplatform/internal/transaction/transactionconsumer/saleorder"

	stockadjustment_consumer "smlcloudplatform/internal/transaction/transactionconsumer/stockadjustment"
	stockpickupproduct_consumer "smlcloudplatform/internal/transaction/transactionconsumer/stockpickupproduct"
	stockreceiveproduct_consumer "smlcloudplatform/internal/transaction/transactionconsumer/stockreceiveproduct"
	stockreturnproduct_consumer "smlcloudplatform/internal/transaction/transactionconsumer/stockreturnproduct"

	stockbalance_consumer "smlcloudplatform/internal/transaction/transactionconsumer/stockbalance"
	stocktranferproduct_consumer "smlcloudplatform/internal/transaction/transactionconsumer/stocktransfer"

	apadvancepayment_consumer "smlcloudplatform/internal/transaction/transactionconsumer/apadvancepayment"
	apadvancepaymentrefund_consumer "smlcloudplatform/internal/transaction/transactionconsumer/apadvancepaymentrefund"
	apdepositpayment_consumer "smlcloudplatform/internal/transaction/transactionconsumer/apdepositpayment"
	apdepositpaymentrefund_consumer "smlcloudplatform/internal/transaction/transactionconsumer/apdepositpaymentrefund"
	creditorpayment_consumer "smlcloudplatform/internal/transaction/transactionconsumer/creditorpayment"

	aradvancepayment_consumer "smlcloudplatform/internal/transaction/transactionconsumer/aradvancepayment"
	aradvancepaymentrefund_consumer "smlcloudplatform/internal/transaction/transactionconsumer/aradvancepaymentrefund"
	ardepositpayment_consumer "smlcloudplatform/internal/transaction/transactionconsumer/ardepositpayment"
	ardepositpaymentrefund_consumer "smlcloudplatform/internal/transaction/transactionconsumer/ardepositpaymentrefund"
	debtorpayment_consumer "smlcloudplatform/internal/transaction/transactionconsumer/debtorpayment"

	"smlcloudplatform/internal/setupconfig"

	goapi "smlcloudplatform/internal/goapi"
	goapi_handlers "smlcloudplatform/internal/goapi/handlers"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func init() {
	// ไม่โหลด .env อีกต่อไป — config ทั้งหมดมาจาก bootstrap.json + MongoDB system_config
	time.Local = time.UTC
}

// @title           BC Ai Account API
// @version         1.0
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @securityDefinitions.apikey  AccessToken
// @in                          header
// @name                        Authorization

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @schemes http https
func main() {

	// โหลด config ทั้งหมดจาก bootstrap.json (ไม่ใช้ .env อีกต่อไป)
	// ทำก่อน config.NewConfig() เพื่อให้ env vars พร้อมใช้งาน
	setupconfig.LoadBootstrapConfig()

	devApiMode := os.Getenv("DEV_API_MODE")
	host := os.Getenv("HOST_API")
	if host != "" {
		fmt.Printf("Host: %v\n", host)
		docs.SwaggerInfo.Host = host
	}

	// config ทั้งหมดมาจาก bootstrap.json เท่านั้น (โหลดไว้แล้วตอนต้น main())
	// ไม่ใช้ MongoDB system_config อีกต่อไป

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		panic(err)
	}

	ms.HttpUsePrometheus()

	if devApiMode == "" || devApiMode == "2" {

		logger.GetLogger().Info("Starting API...")

		ms.Echo().GET("/swagger/*", echoSwagger.WrapHandler)

		cacher := ms.Cacher(cfg.CacherConfig())
		authService := microservice.NewAuthService(cacher, 24*3*time.Hour, 24*30*time.Hour)
		publicPath := []string{
			"/migrationtools/",
			"/swagger/*",

			"/tokenlogin",
			"/googlelogin",

			"/login",
			"/poslogin",
			"/login/email",
			"/login/phone-number",
			"/login/line",
			"/linelogin",
			"/register",
			"/register-username",
			"/refresh",
			"/register-phonenumber",
			"/register/exists-username",
			"/register/exists-phonenumber",
			"/send-phonenumber-otp",

			"/employee/login",

			"/images*",
			"/productimage/*",

			"/healthz",
			"/ws",
			"/metrics",
			"/e-order/product",
			"/e-order/category",
			"/e-order/product-barcode",
			"/e-order/shop-info",
			"/e-order/shop-info/v1.1",
			"/e-order/restaurant/zone",
			"/e-order/restaurant/kitchen",
			"/e-order/restaurant/table",
			"/e-order/sale-invoice/last-pos-docno",
			"/e-order/notify",
			"/line-notify",
			"/line-sync",
			"/reload-config",

			"/goapi/*",        // GoAPI routes — bypass auth (goapi มี auth ของตัวเอง)
			"/api/language/*", // Language — public (pre-login language loading)
		}

		exceptShopPath := []string{
			"/holding",
			"/shop",
			"/profile",
			"/profile/disable-user",
			"/list-holding",
			"/list-shop",
			"/select-holding",
			"/select-shop",
			"/create-holding",
			"/create-shop",
			"/favorite-holding",
			"/favorite-shop",
		}

		// Reload config endpoint — goapi เรียกหลัง save config เพื่อให้ mainapi ใช้ config ใหม่
		ms.Echo().POST("/reload-config", func(c echo.Context) error {
			// ตรวจสอบ shared secret
			secret := c.Request().Header.Get("X-Reload-Secret")
			expectedSecret := os.Getenv("RELOAD_CONFIG_SECRET")
			if expectedSecret != "" && secret != expectedSecret {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			log.Println("[ReloadConfig] ได้รับ request reload config จาก goapi")

			go func() {
				setupconfig.ReloadConfig()
			}()

			return c.JSON(http.StatusOK, map[string]string{
				"message": "กำลัง reload config...",
			})
		})

		ms.HttpMiddleware(authService.MWFuncWithRedisMixShop(cacher, exceptShopPath, publicPath...))
		ms.RegisterLivenessProbeEndpoint("/healthz")
		ms.HttpUseCors()
		ms.HttpPreRemoveTrailingSlash()
		// ms.Echo().GET("/healthz", func(c echo.Context) error {
		// 	return c.String(http.StatusOK, "ok")
		// })
		azureFileBlob := microservice.NewFilePersister()
		imagePersister := microservice.NewPersisterImage(azureFileBlob)

		// Initialize dynamic DB hook and register tenant models
		if err := migration.StartMigrateModel(ms, cfg); err != nil {
			log.Printf("ERROR: Failed to initialize migrate model: %v", err)
		}

		httpServices := []HttpRegister{

			apikeyservice.NewApiKeyServiceHttp(ms, cfg),
			authentication.NewAuthenticationHttp(ms, cfg),
			apikeyservice.NewApiKeyServiceHttp(ms, cfg),
			shop.NewShopHttp(ms, cfg),

			shop.NewShopMemberHttp(ms, cfg),
			employee.NewEmployeeHttp(ms, cfg), member.NewMemberHttp(ms, cfg),

			option.NewOptionHttp(ms, cfg),
			unit.NewUnitHttp(ms, cfg),
			optionpattern.NewOptionPatternHttp(ms, cfg),
			color.NewColorHttp(ms, cfg),

			//product
			products.NewProductHttp(ms, cfg),
			productcategory.NewProductCategoryHttp(ms, cfg),
			productbarcode.NewProductBarcodeHttp(ms, cfg),
			// product.NewProductHttp(ms, cfg),
			formtemplate.NewFormTemplateHttp(ms, cfg),
			productgroup.NewProductGroupHttp(ms, cfg),
			producttype.NewProductTypeHttp(ms, cfg),

			images.NewImagesHttp(ms, cfg, imagePersister),

			// restaurant
			zone.NewZoneHttp(ms, cfg),
			table.NewTableHttp(ms, cfg),
			printer.NewPrinterHttp(ms, cfg),
			kitchen.NewKitchenHttp(ms, cfg),
			zonedesign.NewZoneDesignHttp(ms, cfg),
			settings.NewRestaurantSettingsHttp(ms, cfg),
			device.NewDeviceHttp(ms, cfg),
			staff.NewStaffHttp(ms, cfg),

			chartofaccount.NewChartOfAccountHttp(ms, cfg),
			journal.NewJournalHttp(ms, cfg),
			journal.NewJournalWs(ms, cfg),
			journalreport.NewJournalReportHttp(ms, cfg),
			accountgroup.NewAccountGroupHttp(ms, cfg),
			journalbook.NewJournalBookHttp(ms, cfg),

			documentimage.NewDocumentImageHttp(ms, cfg),
			mastersync.NewMasterSyncHttp(ms, cfg),
			smstransaction.NewSmsTransactionHttp(ms, cfg),
			paymentmaster.NewPaymentMasterHttp(ms, cfg),
			warehouse.NewWarehouseHttp(ms, cfg),

			accountperiodmaster.NewAccountPeriodMasterHttp(ms, cfg),

			bankmaster.NewBankMasterHttp(ms, cfg),
			bookbank.NewBookBankHttp(ms, cfg),
			qrpayment.NewQrPaymentHttp(ms, cfg),

			task.NewTaskHttp(ms, cfg),
			smltransaction.NewSMLTransactionHttp(ms, cfg),

			// debt account
			creditor.NewCreditorHttp(ms, cfg),
			creditorgroup.NewCreditorGroupHttp(ms, cfg),
			debtor.NewDebtorHttp(ms, cfg),
			debtorgroup.NewDebtorGroupHttp(ms, cfg),

			customer.NewCustomerHttp(ms, cfg),
			customergroup.NewCustomerGroupHttp(ms, cfg),

			company.NewCompanyHttp(ms, cfg),
			branch.NewBranchHttp(ms, cfg),
			department.NewDepartmentHttp(ms, cfg),
			businesstype.NewBusinessTypeHttp(ms, cfg),
			costcenter.NewCostCenterHttp(ms, cfg),
			jobproject.NewJobProjectHttp(ms, cfg),

			//transaction
			purchase.NewPurchaseHttp(ms, cfg),
			purchasereturn.NewPurchaseReturnHttp(ms, cfg),
			saleinvoice.NewSaleInvoiceHttp(ms, cfg),
			saleinvoicereturn.NewSaleInvoiceReturnHttp(ms, cfg),
			stocktransfer.NewStockTransferHttp(ms, cfg),
			stockreceiveproduct.NewStockReceiveProductHttp(ms, cfg),
			stockreturnproduct.NewStockReturnProductHttp(ms, cfg),
			stockpickupproduct.NewStockPickupProductHttp(ms, cfg),
			stockadjustment.NewStockAdjustmentHttp(ms, cfg),
			paid.NewPaidHttp(ms, cfg),
			pay.NewPayHttp(ms, cfg),
			stockbalance.NewStockBalanceHttp(ms, cfg),
			stockbalancedetail.NewStockBalanceDetailHttp(ms, cfg),
			purchaseorder.NewPurchaseOrderHttp(ms, cfg),
			purchaserequisition.NewPurchaseRequisitionHttp(ms, cfg),
			rfq.NewRFQHttp(ms, cfg),
			quotation.NewQuotationHttp(ms, cfg),
			saleorder.NewSaleOrderHttp(ms, cfg),
			purchasepartial.NewPurchasepartialHttp(ms, cfg),
			accrualreceive.NewAccrualreceiveHttp(ms, cfg),
			pickandpack.NewPickandpackHttp(ms, cfg),
			pickandpackdevice.NewDeviceHttp(ms, cfg),
			advancepayment.NewAdvancePaymentHttp(ms, cfg),
			advancepaymentrefund.NewAdvancePaymentRefundHttp(ms, cfg),
			deposit.NewDepositHttp(ms, cfg),
			depositrefund.NewDepositRefundHttp(ms, cfg),
			paidadvance.NewPaidAdvanceHttp(ms, cfg),
			paidadvancerefund.NewPaidAdvanceRefundHttp(ms, cfg),
			receivedeposit.NewReceiveDepositHttp(ms, cfg),
			receivedepositrefund.NewReceiveDepositRefundHttp(ms, cfg),
			depositrecord.NewDepositRecordHttp(ms, cfg),
			withdrawalrecord.NewWithdrawalRecordHttp(ms, cfg),
			banktransferrecord.NewBankTransferRecordHttp(ms, cfg),

			creditcardwithdrawal.NewCreditCardWithdrawalHttp(ms, cfg),

			chequedeposit.NewChequeDepositHttp(ms, cfg),
			chequepass.NewChequePassHttp(ms, cfg),
			chequechange.NewChequeChangeHttp(ms, cfg),
			chequedisqualified.NewChequeDisqualifiedHttp(ms, cfg),
			chequerenew.NewChequeRenewHttp(ms, cfg),
			chequereturn.NewChequeReturnHttp(ms, cfg),

			chequepaymentdeposit.NewChequePaymentDepositHttp(ms, cfg),
			chequepaymentchange.NewChequePaymentChangeHttp(ms, cfg),
			chequepaymentdisqualified.NewChequePaymentDisqualifiedHttp(ms, cfg),
			chequepaymentreturn.NewChequePaymentReturnHttp(ms, cfg),

			//product section
			sectionbranch.NewSectionBranchHttp(ms, cfg),
			sectiondepartment.NewSectionDepartmentHttp(ms, cfg),
			sectionbusinesstype.NewSectionBusinessTypeHttp(ms, cfg),

			//smlai product
			brandproduct.NewBrandProductHttp(ms, cfg),
			designproduct.NewDesignProductHttp(ms, cfg),
			modelproduct.NewModelProductHttp(ms, cfg),
			groupproduct.NewGroupProductHttp(ms, cfg),
			gradeproduct.NewGradeProductHttp(ms, cfg),
			patternproduct.NewPatternProductHttp(ms, cfg),
			categoryproduct.NewCategoryProductHttp(ms, cfg),
			classproduct.NewClassProductHttp(ms, cfg),
			groupsuboneproduct.NewGroupsuboneProductHttp(ms, cfg),
			groupsubtwoproduct.NewGroupsubtwoProductHttp(ms, cfg),

			//channel
			salechannel.NewSaleChannelHttp(ms, cfg),
			transportchannel.NewTransportChannelHttp(ms, cfg),

			// e-order
			eorder.NewEOrderHttp(ms, cfg),

			// promiotions
			promotion.NewPromotionHttp(ms, cfg),

			ordertype.NewOrderTypeHttp(ms, cfg),

			pos_setting.NewSettingHttp(ms, cfg),
			pos_media.NewMediaHttp(ms, cfg),
			shift.NewShiftHttp(ms, cfg),

			order_setting.NewSettingHttp(ms, cfg),
			order_device.NewDeviceHttp(ms, cfg),

			documentformate.NewDocumentFormateHttp(ms, cfg),
			ocr.NewOcrHttp(ms, cfg),

			notify.NewNotifyHttp(ms, cfg),
			slipimage.NewSlipImageHttp(ms, cfg),

			// import

			stockbalanceimport.NewStockBalanceImportHttp(ms, cfg),
			productimport.NewProductImportHttp(ms, cfg),

			dimension.NewDimensionHttp(ms, cfg),
			currency.NewCurrencyHttp(ms, cfg),

			// master
			masterexpense.NewMasterExpenseHttp(ms, cfg),
			masterincome.NewMasterIncomeHttp(ms, cfg),
			purchasetype.NewPurchaseTypeHttp(ms, cfg),

			// coupon
			coupon.NewCouponHttp(ms, cfg),

			// system admin
			systemadmin.NewSystemAdmin(ms, cfg),

			temp.NewPOSTempHttp(ms, cfg),
			filestatus.NewFileStatusHttp(ms, cfg),

			// member
			member.NewMemberHttp(ms, cfg),

			// BOM
			bom.NewBOMHttp(ms, cfg),
			saleinvoicebomprice.NewSaleInvoiceBomPriceHttp(ms, cfg),
		}

		serviceStartHttp(ms, httpServices...)

		ms.RegisterHttp(migrationAPI.NewMigrationAPI(ms, cfg))
		ms.RegisterHttp(media.InitMediaUploadHttp(ms, cfg))

		// เริ่ม cleanup scheduler สำหรับ expired coupon reservations
		pst := ms.MongoPersister(cfg.MongoPersisterConfig())

		// สร้าง indexes สำหรับ coupon reservations
		err := coupon_database.CreateCouponReservationIndexes(pst)
		if err != nil {
			fmt.Printf("Error creating coupon reservation indexes: %v\n", err)
		}

		reservationRepo := coupon_repositories.NewCouponReservationRepository(pst)
		cleanupService := coupon_services.NewCouponCleanupService(reservationRepo)

		// เริ่ม background scheduler ที่ตรวจสอบทุก 5 นาที
		// ใช้ empty holdingCode เพื่อ cleanup ทุกร้าน (จะปรับปรุงในอนาคตให้สำหรับแต่ละร้าน)
		go cleanupService.StartCleanupScheduler(context.Background(), 5*time.Minute, "")

		// === GoAPI Routes (BI/Analytics) ===
		goapiServer := goapi.New()
		if err := goapiServer.Init(); err != nil {
			log.Printf("GoAPI init failed: %v (goapi routes disabled)", err)
		} else {
			goapiGroup := ms.Echo().Group("/goapi")
			goapiServer.RegisterMiddleware(goapiGroup)
			goapiServer.RegisterRoutes(goapiGroup, "/goapi")
			defer goapiServer.Shutdown()
			log.Println("GoAPI routes registered under /goapi/*")

			// Public route alias — /api/language/:lang (ไม่ต้อง auth, ไม่ต้องผ่าน /goapi prefix)
			// สำหรับ frontend pre-login language loading
			ms.Echo().GET("/api/language/:lang", goapi_handlers.GetLanguageHandler)
		}
	}

	// Migration
	if devApiMode == "3" {

		logger.GetLogger().Info("Starting Migration...")

		// migration db only
		journal.MigrationJournalTable(ms, cfg)
		chartofaccount.MigrationChartOfAccountTable(ms, cfg)
		productbarcode.MigrationDatabase(ms, cfg)
		// transactionconsumer.MigrationDatabase(ms, cfg)
		// payment migration
		payment.MigrationDatabase(ms, cfg)
		paymentdetail.MigrationDatabase(ms, cfg)
		pay_consumer.MigrationDatabase(ms, cfg)
		paid_consumer.MigrationDatabase(ms, cfg)
		transactionconsumer.MigrationDatabase(ms, cfg)

		// purchase
		purchaseorder_consumer.MigrationDatabase(ms, cfg)       // สั่งซื้อสินค้า
		purchaserequisition_consumer.MigrationDatabase(ms, cfg) // ใบขอซื้อ
		rfq_consumer.MigrationDatabase(ms, cfg)                 // สืบราคา
		purchase_consumer.MigrationDatabase(ms, cfg)            // ซื้อสินค้า
		purchasereturn_consumer.MigrationDatabase(ms, cfg)      // ส่งคืนสินค้า
		purchasedebitnote_consumer.MigrationDatabase(ms, cfg)   // เพิ่มหนี้ซื้อสินค้า
		purchasereceive_consumer.MigrationDatabase(ms, cfg)     // รับสินค้า
		appurchasereceive_consumer.MigrationDatabase(ms, cfg)   // ตั้งหนี้จากการรับ

		// sales
		saleorder_consumer.MigrationDatabase(ms, cfg)         // สั่งขายสินค้า
		saleinvoice_consumer.MigrationDatabase(ms, cfg)       // ขายสินค้า
		saleinvoicereutrn_consumer.MigrationDatabase(ms, cfg) // ส่งคืนสินค้า
		saledebitnote_consumer.MigrationDatabase(ms, cfg)     // เพิ่มหนี้ขายสินค้า

		// stock
		stockreceiveproduct_consumer.MigrationDatabase(ms, cfg)
		stockpickupproduct_consumer.MigrationDatabase(ms, cfg)
		stockreturnproduct_consumer.MigrationDatabase(ms, cfg)
		stockadjustment_consumer.MigrationDatabase(ms, cfg)
		stocktranferproduct_consumer.MigrationDatabase(ms, cfg)

		// ap
		creditorpayment_consumer.MigrationDatabase(ms, cfg)
		apadvancepayment_consumer.MigrationDatabase(ms, cfg)       // จ่ายล่วงหน้าเจ้าหนี้
		apadvancepaymentrefund_consumer.MigrationDatabase(ms, cfg) // คืนเงินจ่ายล่วงหน้าเจ้าหนี้
		apdepositpayment_consumer.MigrationDatabase(ms, cfg)       // ฝากเงินล่วงหน้าเจ้าหนี้
		apdepositpaymentrefund_consumer.MigrationDatabase(ms, cfg) // คืนเงินฝากล่วงหน้าเจ้าหนี้

		// ar
		debtorpayment_consumer.MigrationDatabase(ms, cfg)
		aradvancepayment_consumer.MigrationDatabase(ms, cfg) // รับล่วงหน้าลูกหนี้
		aradvancepaymentrefund_consumer.MigrationDatabase(ms, cfg)
		ardepositpayment_consumer.MigrationDatabase(ms, cfg)       // ฝากเงินล่วงหน้าลูกหนี้
		ardepositpaymentrefund_consumer.MigrationDatabase(ms, cfg) // คืนเงินฝากล่วงหน้าลูกหนี้

		warehouse.MigrationDatabase(ms, cfg)

		stockbalance_consumer.MigrationDatabase(ms, cfg)

		// debt account
		creditor.MigrationDatabase(ms, cfg)
		debtor.MigrationDatabase(ms, cfg)
		shift.MigrationDatabase(ms, cfg)

		// BOM
		bom.MigrationDatabase(ms, cfg)
		saleinvoicebomprice.MigrationDatabase(ms, cfg)

		return
	}

	if devApiMode == "" || devApiMode == "1" {

		logger.GetLogger().Info("Starting Consumers...")

		ms.RegisterLivenessProbeEndpoint("/healthz")

		consumerGroupName := os.Getenv("CONSUMER_GROUP_NAME")
		if consumerGroupName == "" {
			consumerGroupName = "03"
		}

		ms.RegisterConsumer(journal.InitJournalTransactionConsumer(ms, cfg))

		chartofaccount.StartChartOfAccountConsumerCreated(ms, cfg, consumerGroupName)
		chartofaccount.StartChartOfAccountConsumerUpdated(ms, cfg, consumerGroupName)
		chartofaccount.StartChartOfAccountConsumerDeleted(ms, cfg, consumerGroupName)
		chartofaccount.StartChartOfAccountConsumerBlukCreated(ms, cfg, consumerGroupName)

		// Transaction
		ms.RegisterConsumer(stockprocess.NewStockProcessConsumer(ms, cfg))

		// stock
		ms.RegisterConsumer(stockreceiveproduct_consumer.InitStockReceiveTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(stockpickupproduct_consumer.InitStockReceiveTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(stockreturnproduct_consumer.InitStockReturnTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(stockadjustment_consumer.InitStockAdjustmentTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(stocktranferproduct_consumer.InitStockTransferTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(stockbalance_consumer.InitStockReceiveTransactionConsumer(ms, cfg))

		// purchase
		ms.RegisterConsumer(purchaseorder_consumer.InitPurchaseOrderTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(purchaserequisition_consumer.InitPurchaseRequisitionTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(rfq_consumer.InitRFQTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(purchase_consumer.InitPurchaseTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(purchasereturn_consumer.InitPurchaseReturnTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(purchasereceive_consumer.InitPurchaseReceiveTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(appurchasereceive_consumer.InitAPPurchaseReceiveTransactionConsumer(ms, cfg))

		// sale
		ms.RegisterConsumer(saleinvoice_consumer.InitSaleInvoiceTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(saleinvoicereutrn_consumer.InitSaleInvoiceReturnTransactionConsumer(ms, cfg))

		// ap
		ms.RegisterConsumer(creditorpayment_consumer.InitCreditorPaymentTransactionConsumer(ms, cfg))

		// ar
		ms.RegisterConsumer(debtorpayment_consumer.InitDebtorPaymentTransactionConsumer(ms, cfg))

		// Warehouse
		ms.RegisterConsumer(warehouse.InitWarehouseConsumer(ms, cfg))

		// Debt Account
		ms.RegisterConsumer(creditor.InitCreditorConsumer(ms, cfg))
		ms.RegisterConsumer(debtor.InitDebtorConsumer(ms, cfg))

		// Shift
		ms.RegisterConsumer(shift.InitShiftConsumer(ms, cfg))

		// การรับเงิน จ่ายเงิน
		ms.RegisterConsumer(pay_consumer.InitPayTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(paid_consumer.InitPaidTransactionConsumer(ms, cfg))

		// BOM
		ms.RegisterConsumer(bom.InitBOMConsumer(ms, cfg))
		ms.RegisterConsumer(saleinvoicebomprice.InitSaleInvoiceBomPriceConsumer(ms, cfg))

		consumerServices := []ConsumerRegister{
			task.NewTaskConsumer(ms, cfg),
			productbarcode.NewProductBarcodeConsumer(ms, cfg),
		}

		serviceStartConsumer(ms, consumerServices...)
	}

	ms.Start()
}

type HttpRegister interface {
	RegisterHttp()
}

func serviceStartHttp(ms *microservice.Microservice, services ...HttpRegister) {
	for _, service := range services {
		ms.RegisterHttp(service)
	}
}

type ConsumerRegister interface {
	RegisterConsumer()
}

func serviceStartConsumer(ms *microservice.Microservice, services ...ConsumerRegister) {
	for _, service := range services {
		service.RegisterConsumer()
	}
}
