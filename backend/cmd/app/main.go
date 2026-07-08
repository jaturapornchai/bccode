package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"smlcloudplatform/internal/apikeyservice"
	"smlcloudplatform/internal/authentication"
	"smlcloudplatform/internal/channel/salechannel"
	"smlcloudplatform/internal/channel/transportchannel"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/creditor"
	"smlcloudplatform/internal/debtaccount/creditorgroup"
	"smlcloudplatform/internal/debtaccount/customer"
	"smlcloudplatform/internal/debtaccount/customergroup"
	"smlcloudplatform/internal/debtaccount/debtor"
	"smlcloudplatform/internal/debtaccount/debtorgroup"
	"smlcloudplatform/internal/dimension"
	"smlcloudplatform/internal/documentwarehouse/documentimage"
	"smlcloudplatform/internal/filestatus"
	"smlcloudplatform/internal/images"
	"smlcloudplatform/internal/masterexpense"
	"smlcloudplatform/internal/masterincome"
	"smlcloudplatform/internal/mastersync"
	"smlcloudplatform/internal/member"
	"smlcloudplatform/internal/notify"
	"smlcloudplatform/internal/ocr"
	"smlcloudplatform/internal/organization/branch"
	"smlcloudplatform/internal/organization/businesstype"
	"smlcloudplatform/internal/organization/company"
	"smlcloudplatform/internal/organization/department"
	"smlcloudplatform/internal/payment/bankmaster"
	"smlcloudplatform/internal/payment/bookbank"
	"smlcloudplatform/internal/payment/qrpayment"
	"smlcloudplatform/internal/paymentmaster"
	"smlcloudplatform/internal/product/bom"
	"smlcloudplatform/internal/product/color"
	"smlcloudplatform/internal/product/eorder"
	"smlcloudplatform/internal/product/optionpattern"
	"smlcloudplatform/internal/product/ordertype"
	"smlcloudplatform/internal/product/productbarcode"
	"smlcloudplatform/internal/product/productcategory"
	"smlcloudplatform/internal/product/productgroup"
	"smlcloudplatform/internal/product/promotion"
	"smlcloudplatform/internal/product/unit"
	"smlcloudplatform/internal/productimport"
	"smlcloudplatform/internal/productsection/sectionbranch"
	"smlcloudplatform/internal/productsection/sectionbusinesstype"
	"smlcloudplatform/internal/productsection/sectiondepartment"
	"smlcloudplatform/internal/report/reportquerym"
	"smlcloudplatform/internal/slipimage"
	"smlcloudplatform/internal/stockbalanceimport"
	"smlcloudplatform/pkg/microservice"
	"time"

	// "smlcloudplatform/internal/report/reportquery"
	"smlcloudplatform/internal/restaurant/device"
	"smlcloudplatform/internal/restaurant/kitchen"
	"smlcloudplatform/internal/restaurant/notifier"
	"smlcloudplatform/internal/restaurant/notifierdevice"
	"smlcloudplatform/internal/restaurant/printer"
	"smlcloudplatform/internal/restaurant/settings"
	"smlcloudplatform/internal/restaurant/staff"
	"smlcloudplatform/internal/restaurant/table"
	"smlcloudplatform/internal/restaurant/zone"
	"smlcloudplatform/internal/shop"

	// "smlcloudplatform/internal/shop/branch"
	order_device "smlcloudplatform/internal/order/device"
	order_setting "smlcloudplatform/internal/order/setting"
	"smlcloudplatform/internal/pos/media"
	pos_setting "smlcloudplatform/internal/pos/setting"
	"smlcloudplatform/internal/pos/shift"
	"smlcloudplatform/internal/pos/temp"
	"smlcloudplatform/internal/shop/employee"
	"smlcloudplatform/internal/shopdesign/zonedesign"
	"smlcloudplatform/internal/smsreceive/smspatterns"
	"smlcloudplatform/internal/smsreceive/smspaymentsettings"
	"smlcloudplatform/internal/smsreceive/smstransaction"
	"smlcloudplatform/internal/storefront"
	"smlcloudplatform/internal/task"
	"smlcloudplatform/internal/tools"
	"smlcloudplatform/internal/transaction/documentformate"
	"smlcloudplatform/internal/transaction/paid"
	"smlcloudplatform/internal/transaction/pay"
	"smlcloudplatform/internal/transaction/purchase"
	"smlcloudplatform/internal/transaction/purchaseorder"
	"smlcloudplatform/internal/transaction/purchaserequisition"
	"smlcloudplatform/internal/transaction/purchasereturn"
	"smlcloudplatform/internal/transaction/rfq"
	"smlcloudplatform/internal/transaction/saleinvoice"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice"
	"smlcloudplatform/internal/transaction/saleinvoicereturn"
	"smlcloudplatform/internal/transaction/smltransaction"
	"smlcloudplatform/internal/transaction/stockadjustment"
	"smlcloudplatform/internal/transaction/stockbalance"
	"smlcloudplatform/internal/transaction/stockbalancedetail"
	"smlcloudplatform/internal/transaction/stockpickupproduct"
	"smlcloudplatform/internal/transaction/stockreceiveproduct"
	"smlcloudplatform/internal/transaction/stockreturnproduct"
	"smlcloudplatform/internal/transaction/stocktransfer"
	"smlcloudplatform/internal/vfgl/accountgroup"
	"smlcloudplatform/internal/vfgl/accountperiodmaster"
	"smlcloudplatform/internal/vfgl/chartofaccount"
	"smlcloudplatform/internal/vfgl/journal"
	"smlcloudplatform/internal/vfgl/journalbook"
	"smlcloudplatform/internal/vfgl/journalreport"
	"smlcloudplatform/internal/warehouse"

	"smlcloudplatform/internal/logistics/vehicle"

	purchase_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchase"
	saleinvoice_consumer "smlcloudplatform/internal/transaction/transactionconsumer/saleinvoice"

	_ "net/http/pprof"

	"github.com/labstack/echo/v4"
)

func main() {

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		log.Printf("ERROR: Failed to initialize microservice: %v", err)
		os.Exit(1)
	}

	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	cacher := ms.Cacher(cfg.CacherConfig())
	// jwtService := microservice.NewJwtService(cacher, cfg.JwtSecretKey(), 24*3)
	authService := microservice.NewAuthService(cacher, 24*3*time.Hour, 24*30*time.Hour)

	publicPath := []string{
		"/slip/*",
		"/swagger",
		"/login",
		"/poslogin",
		"/login/phone-number",
		"/register",
		"/refresh",
		"/register-phonenumber",
		"/register/exists-username",
		"/register/exists-phonenumber",
		"/send-phonenumber-otp",

		"/employee/login",

		"/images*",
		"/productimage",

		"/healthz",
		"/ws",
		"/metrics",

		"/e-order/product",
		"/e-order/category",
		"/e-order/product-barcode",
		"/e-order/shop-info",
		"/e-order/shop-info/v1.1",
		"/e-order/sale-invoice/last-pos-docno",

		"/e-order/restaurant/zone",
		"/e-order/restaurant/kitchen",
		"/e-order/restaurant/table",
		"/e-order/notify",
		"/line-notify",

		"/restaurant/notifier-device/ref-confirm",
	}

	exceptShopPath := []string{
		"/holding",
		"/shop",
		"/list-holding",
		"/list-shop",
		"/select-holding",
		"/select-shop",
		"/create-holding",
		"/create-shop",
		"/favorite-holding",
		"/verify-token",
	}

	ms.HttpPreRemoveTrailingSlash()
	// ms.HttpUsePrometheus()
	// ms.HttpUseJaeger()

	// ms.HttpMiddleware(authService.MWFuncWithRedis(cacher, publicPath...))
	ms.HttpMiddleware(authService.MWFuncWithRedisMixShop(cacher, exceptShopPath, publicPath...))

	ms.RegisterLivenessProbeEndpoint("/healthz")

	httpServices := []HttpRegister{
		authentication.NewAuthenticationHttp(ms, cfg),
		shop.NewShopHttp(ms, cfg),
		shop.NewShopMemberHttp(ms, cfg),
		member.NewMemberHttp(ms, cfg),
		employee.NewEmployeeHttp(ms, cfg),

		//restaurants
		zone.NewZoneHttp(ms, cfg),
		table.NewTableHttp(ms, cfg),
		printer.NewPrinterHttp(ms, cfg),
		kitchen.NewKitchenHttp(ms, cfg),
		settings.NewRestaurantSettingsHttp(ms, cfg),
		device.NewDeviceHttp(ms, cfg),
		staff.NewStaffHttp(ms, cfg),
		notifier.NewNotifierHttp(ms, cfg),
		notifierdevice.NewNotifierDeviceHttp(ms, cfg),

		//Journal
		journal.NewJournalHttp(ms, cfg),
		journal.NewJournalWs(ms, cfg),
		accountgroup.NewAccountGroupHttp(ms, cfg),
		journalbook.NewJournalBookHttp(ms, cfg),
		zonedesign.NewZoneDesignHttp(ms, cfg),

		mastersync.NewMasterSyncHttp(ms, cfg),

		documentimage.NewDocumentImageHttp(ms, cfg),
		chartofaccount.NewChartOfAccountHttp(ms, cfg),
		//new

		paymentmaster.NewPaymentMasterHttp(ms, cfg),
		apikeyservice.NewApiKeyServiceHttp(ms, cfg),

		smstransaction.NewSmsTransactionHttp(ms, cfg),
		smspatterns.NewSmsPatternsHttp(ms, cfg),
		smspaymentsettings.NewSmsPaymentSettingsHttp(ms, cfg),

		warehouse.NewWarehouseHttp(ms, cfg),
		storefront.NewStorefrontHttp(ms, cfg),

		unit.NewUnitHttp(ms, cfg),
		journalreport.NewJournalReportHttp(ms, cfg),

		optionpattern.NewOptionPatternHttp(ms, cfg),
		color.NewColorHttp(ms, cfg),
		productcategory.NewProductCategoryHttp(ms, cfg),
		productbarcode.NewProductBarcodeHttp(ms, cfg),
		productgroup.NewProductGroupHttp(ms, cfg),

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

		department.NewDepartmentHttp(ms, cfg),
		businesstype.NewBusinessTypeHttp(ms, cfg),
		company.NewCompanyHttp(ms, cfg),
		branch.NewBranchHttp(ms, cfg),

		//transaction
		purchase.NewPurchaseHttp(ms, cfg),
		purchasereturn.NewPurchaseReturnHttp(ms, cfg),
		saleinvoice.NewSaleInvoiceHttp(ms, cfg),
		saleinvoicereturn.NewSaleInvoiceReturnHttp(ms, cfg),
		stockreceiveproduct.NewStockReceiveProductHttp(ms, cfg),
		stockreturnproduct.NewStockReturnProductHttp(ms, cfg),
		stockpickupproduct.NewStockPickupProductHttp(ms, cfg),
		stockadjustment.NewStockAdjustmentHttp(ms, cfg),
		stocktransfer.NewStockTransferHttp(ms, cfg),
		stockbalance.NewStockBalanceHttp(ms, cfg),
		stockbalancedetail.NewStockBalanceDetailHttp(ms, cfg),
		purchaseorder.NewPurchaseOrderHttp(ms, cfg),
		purchaserequisition.NewPurchaseRequisitionHttp(ms, cfg),
		rfq.NewRFQHttp(ms, cfg),

		//product section
		sectionbranch.NewSectionBranchHttp(ms, cfg),
		sectiondepartment.NewSectionDepartmentHttp(ms, cfg),
		sectionbusinesstype.NewSectionBusinessTypeHttp(ms, cfg),

		//channel
		salechannel.NewSaleChannelHttp(ms, cfg),
		transportchannel.NewTransportChannelHttp(ms, cfg),

		paid.NewPaidHttp(ms, cfg),
		pay.NewPayHttp(ms, cfg),

		promotion.NewPromotionHttp(ms, cfg),

		eorder.NewEOrderHttp(ms, cfg),

		ordertype.NewOrderTypeHttp(ms, cfg),

		// POS
		pos_setting.NewSettingHttp(ms, cfg),
		shift.NewShiftHttp(ms, cfg),
		order_setting.NewSettingHttp(ms, cfg),
		order_device.NewDeviceHttp(ms, cfg),

		documentformate.NewDocumentFormateHttp(ms, cfg),

		//reportquery.NewReportQueryHttp(ms, cfg),
		reportquerym.NewReportQueryHttp(ms, cfg),

		media.NewMediaHttp(ms, cfg),

		ocr.NewOcrHttp(ms, cfg),
		stockbalanceimport.NewStockBalanceImportHttp(ms, cfg),
		productbarcode.NewProductBarcodeHttp(ms, cfg),
		notify.NewNotifyHttp(ms, cfg),
		slipimage.NewSlipImageHttp(ms, cfg),
		productimport.NewProductImportHttp(ms, cfg),

		dimension.NewDimensionHttp(ms, cfg),

		// master
		masterexpense.NewMasterExpenseHttp(ms, cfg),
		masterincome.NewMasterIncomeHttp(ms, cfg),

		temp.NewPOSTempHttp(ms, cfg),
		filestatus.NewFileStatusHttp(ms, cfg),

		member.NewMemberHttp(ms, cfg),

		bom.NewBOMHttp(ms, cfg),
		saleinvoicebomprice.NewSaleInvoiceBomPriceHttp(ms, cfg),

		// logistics
		vehicle.NewVehicleHttp(ms, cfg),
	}

	azureFileBlob := microservice.NewFilePersister()
	imagePersister := microservice.NewPersisterImage(azureFileBlob)

	ms.RegisterHttp(images.NewImagesHttp(ms, cfg, imagePersister))

	serviceStartHttp(ms, httpServices...)

	// journal.MigrationJournalTable(ms, cfg)

	// warehouse.MigrationDatabase(ms, cfg)
	// creditor.MigrationDatabase(ms, cfg)
	// ms.RegisterConsumer(creditor.InitCreditorConsumer(ms, cfg))

	debtor.MigrationDatabase(ms, cfg)
	ms.RegisterConsumer(debtor.InitDebtorConsumer(ms, cfg))
	// warehouse.MigrationDatabase(ms, cfg)

	// inventory.StartInventoryAsync(ms, cfg)
	// inventory.StartInventoryComsumeCreated(ms, cfg)

	// transactionconsumer.MigrationDatabase(ms, cfg)

	// payment.MigrationDatabase(ms, cfg)
	// paymentdetail.MigrationDatabase(ms, cfg)

	// purchase_consumer.MigrationDatabase(ms, cfg)
	// saleinvoice_consumer.MigrationDatabase(ms, cfg)

	// productbarcode.MigrationDatabase(ms, cfg)

	ms.RegisterConsumer(purchase_consumer.InitPurchaseTransactionConsumer(ms, cfg))

	ms.RegisterConsumer(saleinvoice_consumer.InitSaleInvoiceTransactionConsumer(ms, cfg))

	ms.RegisterConsumer(journal.InitJournalTransactionConsumer(ms, cfg))

	ms.RegisterConsumer(warehouse.InitWarehouseConsumer(ms, cfg))

	consumeServices := []ConsumerRegister{}

	task.NewTaskConsumer(ms, cfg).RegisterConsumer()
	productbarcode.NewProductBarcodeConsumer(ms, cfg).RegisterConsumer()

	serviceStartConsumer(ms, consumeServices...)

	toolSvc := tools.NewToolsService(ms, cfg)

	toolSvc.RegisterHttp()

	ms.Echo().GET("/routes", func(ctx echo.Context) error {
		data, err := json.MarshalIndent(ms.Echo().Routes(), "", "  ")

		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		os.WriteFile("routes.json", data, 0644)

		// ctx.JSON(http.StatusOK, data)
		ctx.JSON(http.StatusOK, map[string]interface{}{"success": true, "data": ms.Echo().Routes()})

		return nil
	})

	//ms.Echo().GET("/swagger/*", echoSwagger.WrapHandler)

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
	RegisterConsumer(*microservice.Microservice)
}

func serviceStartConsumer(ms *microservice.Microservice, services ...ConsumerRegister) {
	for _, service := range services {
		ms.RegisterConsumer(service)
	}
}
