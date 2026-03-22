import 'package:flutter/material.dart';
import 'package:smlaicloud/modules/cart-system/cartsystem.dart';
import 'package:smlaicloud/screens/config/company_screen.dart';
import 'package:smlaicloud/screens/config/product_screen.dart';
import 'package:smlaicloud/screens/config/product_barcode_screen.dart';
import 'package:smlaicloud/screens/config/product_unit_screen.dart';
import 'package:smlaicloud/screens/config/product_group_screen.dart';
import 'package:smlaicloud/screens/config/product_category_list_screen.dart';
import 'package:smlaicloud/screens/config/company_branch_screen.dart';
import 'package:smlaicloud/screens/config/department_screen.dart';
import 'package:smlaicloud/screens/config/cost_center_screen.dart';
import 'package:smlaicloud/screens/config/job_project_screen.dart';
import 'package:smlaicloud/screens/config/employee_screen.dart';
import 'package:smlaicloud/screens/formdesign/form_design.dart';
import 'package:smlaicloud/screens/config/user_screen.dart';
import 'package:smlaicloud/screens/config/creditor_screen.dart';
import 'package:smlaicloud/screens/config/creditor_group_screen.dart';
import 'package:smlaicloud/screens/config/debtor_screen.dart';
import 'package:smlaicloud/screens/config/debtor_group_screen.dart';
import 'package:smlaicloud/screens/config/bank_screen.dart';
import 'package:smlaicloud/screens/currency/currency_list_screen.dart';
import 'package:smlaicloud/screens/currency/exchange_rate_screen.dart';
import 'package:smlaicloud/screens/config/pos_setting_screen.dart';
import 'package:smlaicloud/screens/config/order_setting_screen.dart';
import 'package:smlaicloud/screens/config/order_setting_teamplate_screen.dart';
import 'package:smlaicloud/screens/config/pos_media_screen.dart';
import 'package:smlaicloud/screens/config/doc_format_screen.dart';
import 'package:smlaicloud/screens/config/bill_design_screen.dart';
import 'package:smlaicloud/screens/config/qr_screen.dart';
import 'package:smlaicloud/screens/config/config_screen.dart';
import 'package:smlaicloud/screens/config/product_dimension_screen.dart';
import 'package:smlaicloud/screens/config/product_barcode_bom_screen.dart';
import 'package:smlaicloud/screens/config/promotion_screen.dart';
import 'package:smlaicloud/screens/config/product_barcode_shelf.dart';
import 'package:smlaicloud/screens/config/price_history/price_history_screen.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_gross_profit_by_document.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_gross_profit_by_product.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_vatbuy.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_vatsale.dart';
import 'package:smlaicloud/screens/report/reportmovement_screen.dart';
import 'package:smlaicloud/screens/report/pdf_report_main_screen.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_sales.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_sale_daily.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_stock_movement_screen.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_payment_daily.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_sale_return.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_stock_balance.dart';
import 'package:smlaicloud/screens/report/dedebi/report_dedebi_purchase_partial.dart';
import 'package:smlaicloud/screens/report/dedebi/report_sale_profit_report_by_document.dart';
import 'package:smlaicloud/screens/report/report_product_balance.dart';
import 'package:smlaicloud/screens/import/import_product_screen.dart';
import 'package:smlaicloud/screens/import/import_product_from_file_screen.dart';
import 'package:smlaicloud/screens/import/import_product_image_screen.dart';
import 'package:smlaicloud/screens/transaction/transaction_stock_balance.dart';
import 'package:smlaicloud/screens/transaction/transaction_edit.dart';
import 'package:smlaicloud/screens/purchaseorder/purchaseorder_edit_screen.dart';
import 'package:smlaicloud/screens/purchaserequisition/purchaserequisition_edit_screen.dart';
import 'package:smlaicloud/screens/rfq/rfq_edit_screen.dart';
import 'package:smlaicloud/screens/purchaseorder/purchaseorder_list_screen.dart';
import 'package:smlaicloud/screens/procurement_dashboard/procurement_dashboard_screen.dart';
import 'package:smlaicloud/screens/transaction/transaction_paid.dart';
import 'package:smlaicloud/screens/gl/gl_process_screen.dart';
import 'package:smlaicloud/screens/check_daily/daily_info_screen.dart';
import 'package:smlaicloud/screens/config/group_number_select_screen.dart';
import 'package:smlaicloud/screens/config/line_notify_screen.dart';
import 'package:smlaicloud/screens/config/reminder_screen.dart';
import 'package:smlaicloud/screens/lineoa/lineoa_config_screen.dart';
import 'package:smlaicloud/screens/enhanced_cash_drawer_screen.dart';
import 'package:smlaicloud/screens/config/point_setting_screen.dart';
import 'package:smlaicloud/screens/coupon/coupon_screen.dart';
import 'package:smlaicloud/screens/master/master_brand_screen.dart';
import 'package:smlaicloud/screens/master/master_category_screen.dart';
import 'package:smlaicloud/screens/master/master_class_screen.dart';
import 'package:smlaicloud/screens/master/master_design_screen.dart';
import 'package:smlaicloud/screens/master/master_grade_screen.dart';
import 'package:smlaicloud/menu_screen.dart';
import 'package:smlaicloud/screens/master/master_model_screen.dart';
import 'package:smlaicloud/screens/master/master_pattern_screen.dart';
import 'package:smlaicloud/screens/master/master_group_screen.dart';
import 'package:smlaicloud/screens/master/master_group_sub1_screen.dart';
import 'package:smlaicloud/screens/master/master_group_sub2_screen.dart';
import 'package:smlaicloud/usersystem/login_shop_screen.dart';
import 'package:smlaicloud/usersystem/flavor_login_selector.dart';
import 'package:smlaicloud/screens/config/work_day_screen.dart';
import 'package:smlaicloud/screens/config/holiday_screen.dart';
import 'package:smlaicloud/screens/config/color_screen.dart';
import 'package:smlaicloud/screens/config/business_type_screen.dart';
import 'package:smlaicloud/screens/config/book_bank_screen.dart';
import 'package:smlaicloud/screens/config/sale_channel.dart';
import 'package:smlaicloud/screens/config/transport_channel.dart';
import 'package:smlaicloud/screens/config/permission_definition_screen.dart';
import 'package:smlaicloud/screens/config/permission_link_screen.dart';
import 'package:smlaicloud/screens/config/mcp_apikey_screen.dart';
import 'package:smlaicloud/screens/config/copy_uat_to_dev_screen.dart';
import 'package:smlaicloud/screens/config/product_warehouse.dart';
import 'package:smlaicloud/screens/config/product_location.dart';
import 'package:smlaicloud/screens/config/order_type_screen.dart';
import 'package:smlaicloud/screens/config/product_type_screen.dart';
import 'package:smlaicloud/screens/config/add_product_to_branch_screen.dart';
import 'package:smlaicloud/screens/config/add_product_to_department_screen.dart';
import 'package:smlaicloud/screens/config/add_product_to_kitchen_screen.dart';
import 'package:smlaicloud/screens/report/stock/report_stock_balance_item_warehouse_location.dart';
import 'package:smlaicloud/screens/report/stock/report_stock_balance_warehouse_barcode.dart';
import 'package:smlaicloud/screens/report/stock/report_stock_balance_location_barcode.dart';
import 'package:smlaicloud/screens/report/stock/report_stock_movement_cost.dart';
import 'package:smlaicloud/screens/audit_screen.dart';
import 'package:smlaicloud/screens/knowledge_base/knowledge_base_screen.dart';
import 'package:smlaicloud/screens/alert_agent/alert_agent_screen.dart';
import 'package:smlaicloud/screens/config/rebuild_stock_screen.dart';
import 'package:smlaicloud/screens/config/rebuild_products_screen.dart';
import 'package:smlaicloud/screens/config/rebuild_product_balance_screen.dart';
import 'package:smlaicloud/screens/slip/slip_list_screen.dart';
import 'package:smlaicloud/model/slip_image_model.dart';
import 'package:smlaicloud/screens/config/approval/purchase_type_screen.dart';
import 'package:smlaicloud/screens/config/approval/po_approval_setting_screen.dart';
import 'package:smlaicloud/screens/config/approval/quotation_type_screen.dart';
import 'package:smlaicloud/screens/config/approval/qt_approval_setting_screen.dart';
import 'package:smlaicloud/screens/config/approval/sale_order_type_screen.dart';
import 'package:smlaicloud/screens/config/approval/so_approval_setting_screen.dart';
import 'package:smlaicloud/screens/quotation/quotation_edit_screen.dart';
import 'package:smlaicloud/screens/quotation/quotation_list_screen.dart';
import 'global.dart' as global;

/// Helper class สำหรับจัดการ routes ภายใน tab
class AppRoutes {
  /// สร้าง route สำหรับใช้ใน Navigator แต่ละ tab
  static Route<dynamic>? onGenerateRoute(RouteSettings settings) {
    switch (settings.name) {
      case '/cart-system':
        return MaterialPageRoute(builder: (_) => const CartSystemScreen(), settings: settings);
      case '/company':
        return MaterialPageRoute(builder: (_) => const CompanyScreen(), settings: settings);
      case '/product':
        return MaterialPageRoute(builder: (_) => const ProductScreen(), settings: settings);
      case '/product_barcode':
        return MaterialPageRoute(builder: (_) => const ProductBarcodeScreen(), settings: settings);
      case '/productunit':
        return MaterialPageRoute(builder: (_) => const ProductUnitScreen(), settings: settings);
      case '/productgroup':
        return MaterialPageRoute(builder: (_) => const ProductGroupScreen(), settings: settings);
      case '/productcategorylist':
        return MaterialPageRoute(builder: (_) => const ProductCategoryListScreen(), settings: settings);
      case '/branch':
        return MaterialPageRoute(builder: (_) => const CompanyBranchScreen(), settings: settings);
      case '/business_type_screen':
        return MaterialPageRoute(builder: (_) => const BusinessTypeScreen(), settings: settings);
      case '/department':
        return MaterialPageRoute(builder: (_) => const DepartmentScreen(), settings: settings);
      case '/costcenter':
        return MaterialPageRoute(builder: (_) => const CostCenterScreen(), settings: settings);
      case '/jobproject':
        return MaterialPageRoute(builder: (_) => const JobProjectScreen(), settings: settings);
      case '/project':
        return MaterialPageRoute(builder: (_) => const JobProjectScreen(isProjectMode: true), settings: settings);
      case '/job':
        return MaterialPageRoute(builder: (_) => const JobProjectScreen(isProjectMode: false), settings: settings);
      case '/employee':
        return MaterialPageRoute(builder: (_) => const EmployeeScreen(), settings: settings);
      case '/formdesign':
        return MaterialPageRoute(builder: (_) => const FormDesignScreen(), settings: settings);
      case '/user':
        return MaterialPageRoute(builder: (_) => const UserScreen(), settings: settings);
      case '/permission_definition':
        return MaterialPageRoute(builder: (_) => const PermissionDefinitionScreen(), settings: settings);
      case '/permission_link':
        return MaterialPageRoute(builder: (_) => const PermissionLinkScreen(), settings: settings);
      case '/mcp_apikey':
        return MaterialPageRoute(builder: (_) => const MCPAPIKeyScreen(), settings: settings);
      case '/copy_uat_to_dev':
        return MaterialPageRoute(builder: (_) => const CopyUatToDevScreen(), settings: settings);
      case '/creditor':
        return MaterialPageRoute(builder: (_) => const CreditorScreen(), settings: settings);
      case '/creditorgroup':
        return MaterialPageRoute(builder: (_) => const CreditorGroupScreen(), settings: settings);
      case '/debtor':
        return MaterialPageRoute(builder: (_) => const DebtorScreen(), settings: settings);
      case '/debtorgroup':
        return MaterialPageRoute(builder: (_) => const DebtorGroupScreen(), settings: settings);
      case '/bank':
        return MaterialPageRoute(builder: (_) => const BankScreen(), settings: settings);
      case '/currency':
        return MaterialPageRoute(builder: (_) => const CurrencyListScreen(), settings: settings);
      case '/exchange_rate':
        return MaterialPageRoute(builder: (_) => const ExchangeRateScreen(), settings: settings);
      case '/possetting':
        return MaterialPageRoute(builder: (_) => const PosSettingScreen(), settings: settings);
      case '/ordersetting':
        return MaterialPageRoute(builder: (_) => const OrderSettingScreen(), settings: settings);
      case '/ordertemplatsetting':
        return MaterialPageRoute(builder: (_) => const OrderTemplateSettingScreen(), settings: settings);
      case '/posmedia':
        return MaterialPageRoute(builder: (_) => const PosMediaScreen(), settings: settings);
      case '/docformat':
        return MaterialPageRoute(builder: (_) => const DocFormatScreen(), settings: settings);
      case '/billdesign':
        return MaterialPageRoute(builder: (_) => const BillDesignScreen(), settings: settings);
      case '/qrprovider':
        return MaterialPageRoute(builder: (_) => const QrScreen(), settings: settings);
      case '/config_system':
        return MaterialPageRoute(builder: (_) => const ConfigScreen(), settings: settings);
      case '/product_dimension':
        return MaterialPageRoute(builder: (_) => const ProductDimensionScreen(), settings: settings);
      case '/product_bom':
        return MaterialPageRoute(builder: (_) => const ProductBarcodeBomScreen(), settings: settings);
      case '/promotion_screen':
        return MaterialPageRoute(builder: (_) => const PromotionScreen(), settings: settings);
      case '/product_barcode_shelf':
        return MaterialPageRoute(builder: (_) => const ProductBarcodeShelf(), settings: settings);
      case '/price_history':
        return MaterialPageRoute(builder: (_) => const PriceHistoryScreen(), settings: settings);
      case '/report/movement':
        return MaterialPageRoute(builder: (_) => const ReportMovementScreen(), settings: settings);
      case '/report/pdfreport':
        return MaterialPageRoute(builder: (_) => const PdfReportMainScreen(), settings: settings);
      case '/report/report_dedebi_sales':
        return MaterialPageRoute(builder: (_) => const ReportDedebiSalesScreen(), settings: settings);
      case '/report/report_dedebi_sales_daily':
        return MaterialPageRoute(builder: (_) => const ReportDedebiSaleDailyScreen(), settings: settings);
      case '/report/report_stock_movement':
        return MaterialPageRoute(builder: (_) => const ReportStockMovementScreen(), settings: settings);
      case '/report/report_dedebi_payment_daily':
        return MaterialPageRoute(builder: (_) => const ReportPaymentDailyScreen(), settings: settings);
      case '/report/report_dedebi_sale_return':
        return MaterialPageRoute(builder: (_) => const ReportDedebiSaleReturnScreen(), settings: settings);
      case '/report/report_dedebi_stock_balance':
        return MaterialPageRoute(builder: (_) => const ReportStockBalanceScreen(), settings: settings);
      case '/report/report_dedebi_purchase_partial':
        return MaterialPageRoute(builder: (_) => const ReportDedebiPurchasePartialScreen(), settings: settings);
      case '/report/sales_report_by_document':
        return MaterialPageRoute(builder: (_) => const ReportSaleProfitReportByDocumentScreen(), settings: settings);
      case '/report/report_gross_profit_by_document':
        return MaterialPageRoute(builder: (_) => const ReportGrossProfitByDocumentScreen(), settings: settings);
      case '/report/report_gross_profit_by_product':
        return MaterialPageRoute(builder: (_) => const ReportGrossProfitByProductScreen(), settings: settings);
      case '/report/report_vat_sale':
        return MaterialPageRoute(builder: (_) => const ReportVatSaleScreen(), settings: settings);
      case '/report/report_vat_buy':
        return MaterialPageRoute(builder: (_) => const ReportVatBuyScreen(), settings: settings);
      case '/report/product_balance':
        return MaterialPageRoute(builder: (_) => const ReportProductBalanceScreen(), settings: settings);
      case '/importproduct':
        return MaterialPageRoute(builder: (_) => const ImportProductScreen(), settings: settings);
      case '/importproductfromfile':
        return MaterialPageRoute(builder: (_) => const ImportProductFromFileScreen(), settings: settings);
      case '/importproductimage':
        return MaterialPageRoute(builder: (_) => const ImportProductImageScreen(), settings: settings);
      case '/transaction/stockbalance':
        return MaterialPageRoute(
          builder: (_) => const TransactionStockBalaceScreen(type: global.TransactionTypeEnum.stockbalance),
          settings: settings,
        );
      case '/transaction/purchase':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.purchase),
          settings: settings,
        );
      case '/transaction/purchasereturn':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.purchasereturn),
          settings: settings,
        );
      case '/gl/gl_process':
        return MaterialPageRoute(builder: (_) => const GlProcessScreen(), settings: settings);
      case '/check_daily/daily_info_screen':
        return MaterialPageRoute(builder: (_) => const DailyInfoScreen(), settings: settings);
      case '/product_category_group_select_screen':
        return MaterialPageRoute(
          builder: (_) => const GroupNumberSelectScreen(type: global.SelectGroupNumberEnum.category),
          settings: settings,
        );
      case '/zone_group_select_screen':
        return MaterialPageRoute(
          builder: (_) => const GroupNumberSelectScreen(type: global.SelectGroupNumberEnum.zone),
          settings: settings,
        );
      case '/table_group_select_screen':
        return MaterialPageRoute(
          builder: (_) => const GroupNumberSelectScreen(type: global.SelectGroupNumberEnum.table),
          settings: settings,
        );
      case '/table_map_group_select_screen':
        return MaterialPageRoute(
          builder: (_) => const GroupNumberSelectScreen(type: global.SelectGroupNumberEnum.tableOrder),
          settings: settings,
        );
      case '/kitchen_group_select_screen':
        return MaterialPageRoute(
          builder: (_) => const GroupNumberSelectScreen(type: global.SelectGroupNumberEnum.kitchen),
          settings: settings,
        );
      case '/qrcodeorder_group_select_screen':
        return MaterialPageRoute(
          builder: (_) => const GroupNumberSelectScreen(type: global.SelectGroupNumberEnum.genQrcode),
          settings: settings,
        );
      case '/line_notify':
        return MaterialPageRoute(builder: (_) => const LineNotifyScreen(), settings: settings);
      case '/lineoa_config':
        return MaterialPageRoute(builder: (_) => const LineOAConfigScreen(), settings: settings);
      case '/reminder':
        return MaterialPageRoute(builder: (_) => const ReminderScreen(), settings: settings);
      case '/cashing_in_the_drawer':
        return MaterialPageRoute(builder: (_) => const EnhancedCashInDrawerScreen(), settings: settings);
      case '/point_setting':
        return MaterialPageRoute(builder: (_) => const PointSettingScreen(), settings: settings);
      case '/coupon_setting':
        return MaterialPageRoute(builder: (_) => const CouponScreen(), settings: settings);
      case '/master_brand_screen':
        return MaterialPageRoute(builder: (_) => const MasterBrandScreen(), settings: settings);
      case '/master_category_screen':
        return MaterialPageRoute(builder: (_) => const MasterCategoryScreen(), settings: settings);
      case '/master_class_screen':
        return MaterialPageRoute(builder: (_) => const MasterClassScreen(), settings: settings);
      case '/master_design_screen':
        return MaterialPageRoute(builder: (_) => const MasterDesignScreen(), settings: settings);
      case '/master_grade_screen':
        return MaterialPageRoute(builder: (_) => const MasterGradeScreen(), settings: settings);
      case '/master_model_screen':
        return MaterialPageRoute(builder: (_) => const MasterModelScreen(), settings: settings);
      case '/master_pattern_screen':
        return MaterialPageRoute(builder: (_) => const MasterPatternScreen(), settings: settings);
      case '/master_group_screen':
        return MaterialPageRoute(builder: (_) => const MasterGroupScreen(), settings: settings);
      case '/master_group_sub1_screen':
        return MaterialPageRoute(builder: (_) => const MasterGroupSub1Screen(), settings: settings);
      case '/master_group_sub2_screen':
        return MaterialPageRoute(builder: (_) => const MasterGroupSub2Screen(), settings: settings);
      case '/select_shop_screen':
        return MaterialPageRoute(builder: (_) => const LoginShopScreen(), settings: settings);
      case '/work_day_screen':
        return MaterialPageRoute(builder: (_) => const WorkDayScreen(), settings: settings);
      case '/holiday_screen':
        return MaterialPageRoute(builder: (_) => const HolidayScreen(), settings: settings);
      case '/color_screen':
        return MaterialPageRoute(builder: (_) => const ColorScreen(), settings: settings);
      case '/bussiness_type_screen':
        return MaterialPageRoute(builder: (_) => const BusinessTypeScreen(), settings: settings);
      case '/book_bank_screen':
        return MaterialPageRoute(builder: (_) => const BookBankScreen(), settings: settings);
      case '/sale_channel_screen':
        return MaterialPageRoute(builder: (_) => const SaleChannelScreen(), settings: settings);
      case '/transport_channel_screen':
        return MaterialPageRoute(builder: (_) => const TransportChannelScreen(), settings: settings);
      case '/product_warehouse_screen':
        return MaterialPageRoute(builder: (_) => const ProductWarehouseScreen(), settings: settings);
      case '/product_location_screen':
        return MaterialPageRoute(builder: (_) => const ProductLocaltionScreen(), settings: settings);
      case '/order_type_screen':
        return MaterialPageRoute(builder: (_) => const OrderTypeScreen(), settings: settings);
      case '/product_type_screen':
        return MaterialPageRoute(builder: (_) => const ProductTypeScreen(), settings: settings);
      case '/add_product_to_branch_screen':
        return MaterialPageRoute(builder: (_) => const AddProductToBranchScreen(), settings: settings);
      case '/add_product_to_department_screen':
        return MaterialPageRoute(builder: (_) => const AddProductToDepartmentScreen(), settings: settings);
      case '/add_product_to_kitchen_screen':
        return MaterialPageRoute(builder: (_) => const AddProductToKitchenScreen(), settings: settings);
      case '/transaction/purchaserequisition':
        return MaterialPageRoute(builder: (_) => const PurchaseRequisitionEditScreen(), settings: settings);
      case '/transaction/rfq':
        return MaterialPageRoute(builder: (_) => const RFQEditScreen(), settings: settings);
      case '/transaction/purchaseorder':
        return MaterialPageRoute(
          builder: (_) => const PurchaseOrderEditScreen(),
          settings: settings,
        );
      case '/transaction/purchaseorder_list':
        return MaterialPageRoute(
          builder: (_) => const PurchaseOrderListScreen(),
          settings: settings,
        );
      case '/procurement/dashboard':
        return MaterialPageRoute(
          builder: (_) => const ProcurementDashboardScreen(),
          settings: settings,
        );
      case '/transaction/advancepayment':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.advancePayment),
          settings: settings,
        );
      case '/transaction/advancepaymentrefund':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.advancePaymentRefund),
          settings: settings,
        );
      case '/transaction/deposit':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.deposit),
          settings: settings,
        );
      case '/transaction/depositrefund':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.depositRefund),
          settings: settings,
        );
      case '/transaction/purchasepartial':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.purchasepartial),
          settings: settings,
        );
      case '/transaction/accrualreceive':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.accrualreceive),
          settings: settings,
        );
      case '/transaction/saleorder':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.saleorder),
          settings: settings,
        );
      case '/transaction/quotation':
        return MaterialPageRoute(
          builder: (_) => const QuotationEditScreen(),
          settings: settings,
        );
      case '/transaction/quotation_list':
        return MaterialPageRoute(
          builder: (_) => const QuotationListScreen(),
          settings: settings,
        );
      case '/transaction/paidadvance':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.paidAdvance),
          settings: settings,
        );
      case '/transaction/paidadvancerefund':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.paidAdvanceRefund),
          settings: settings,
        );
      case '/transaction/receivedeposit':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.receiveDeposit),
          settings: settings,
        );
      case '/transaction/receivedepositrefund':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.receiveDepositRefund),
          settings: settings,
        );
      case '/transaction/sale':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.sale),
          settings: settings,
        );
      case '/transaction/salereturn':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.salereturn),
          settings: settings,
        );
      case '/transaction/stocktransfer':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.stocktransfer),
          settings: settings,
        );
      case '/transaction/stockreceiveproduct':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.stockreceiveproduct),
          settings: settings,
        );
      case '/transaction/stockpickupproduct':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.stockpickupproduct),
          settings: settings,
        );
      case '/transaction/stockreturnproduct':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.stockreturnproduct),
          settings: settings,
        );
      case '/transaction/adjust':
        return MaterialPageRoute(
          builder: (_) => const TransactionEditScreen(type: global.TransactionTypeEnum.adjust),
          settings: settings,
        );
      case '/transaction/paid':
        return MaterialPageRoute(
          builder: (_) => const TransactionPaidScreen(type: global.TransactionTypeEnum.paid),
          settings: settings,
        );
      case '/transaction/pay':
        return MaterialPageRoute(
          builder: (_) => const TransactionPaidScreen(type: global.TransactionTypeEnum.pay),
          settings: settings,
        );
      case '/report/stock_balance_item':
        return MaterialPageRoute(builder: (_) => ReportStockBalanceItemWareHouseLocation(), settings: settings);
      case '/report/stock_balance_warehouse':
        return MaterialPageRoute(builder: (_) => ReportStockBalanceWhBarcode(), settings: settings);
      case '/report/stock_balance_location':
        return MaterialPageRoute(builder: (_) => ReportStockBalanceLocationBarcode(), settings: settings);
      case '/report/stock_movement_cost':
        return MaterialPageRoute(builder: (_) => ReportStockMovementCost(), settings: settings);
      case '/audit_screen':
        return MaterialPageRoute(builder: (_) => const AuditScreen(), settings: settings);
      case '/rebuild_stock_screen':
        return MaterialPageRoute(builder: (_) => const RebuildStockScreen(), settings: settings);
      case '/rebuild_products_screen':
        return MaterialPageRoute(builder: (_) => const RebuildProductsScreen(), settings: settings);
      case '/rebuild_product_balance_screen':
        return MaterialPageRoute(builder: (_) => const RebuildProductBalanceScreen(), settings: settings);
      case '/knowledge_base_screen':
        return MaterialPageRoute(builder: (_) => const KnowledgeBaseScreen(), settings: settings);
      case '/alert_agent_screen':
        return MaterialPageRoute(builder: (_) => const AlertAgentScreen(), settings: settings);
      case '/menu':
        return MaterialPageRoute(builder: (_) => const MenuScreen(), settings: settings);
      // Login screens
      case '/login_screen':
        return MaterialPageRoute(builder: (_) => const FlavorLoginSelector(), settings: settings);
      case '/login_screen_shop':
        return MaterialPageRoute(builder: (_) => const LoginShopScreen(), settings: settings);
      // SLIP เงินเข้า/ออก
      case '/slip_money_in':
        return MaterialPageRoute(builder: (_) => const SlipListScreen(slipType: SlipType.moneyIn), settings: settings);
      case '/slip_money_out':
        return MaterialPageRoute(builder: (_) => const SlipListScreen(slipType: SlipType.moneyOut), settings: settings);
      // ระบบอนุมัติ (Approval System)
      case '/purchase_type_screen':
        return MaterialPageRoute(builder: (_) => const PurchaseTypeScreen(), settings: settings);
      case '/po_approval_setting_screen':
        return MaterialPageRoute(builder: (_) => const POApprovalSettingScreen(), settings: settings);
      case '/quotation_type_screen':
        return MaterialPageRoute(builder: (_) => const QuotationTypeScreen(), settings: settings);
      case '/qt_approval_setting_screen':
        return MaterialPageRoute(builder: (_) => const QTApprovalSettingScreen(), settings: settings);
      case '/sale_order_type_screen':
        return MaterialPageRoute(builder: (_) => const SaleOrderTypeScreen(), settings: settings);
      case '/so_approval_setting_screen':
        return MaterialPageRoute(builder: (_) => const SOApprovalSettingScreen(), settings: settings);
      default:
        return null;
    }
  }
}
