// ignore_for_file: deprecated_member_use

import 'dart:io';
import 'dart:async';
import 'package:smlaicloud/bloc/export_csv/export_csv_bloc.dart';
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/book_bank_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/services/pdf_service.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/screen_search/bookbank_select_screen.dart';
import 'package:smlaicloud/screens/purchaseorder/purchaseorder_list_screen.dart';
import 'package:smlaicloud/screens/transaction/components/document_header_widget.dart';
import 'package:smlaicloud/screens/transaction/components/document_preview_widget.dart';
import 'package:smlaicloud/screens/transaction/components/document_product_list_widget.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_dialogs.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_calculator.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_calculator_bridge.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_search_handlers.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/modules/cart-system/models/cart_transfer_model.dart' show CartTransferModel, CartTransferProvider;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/model/approval_model.dart';
// data_history_dialog ถูกย้ายไป po_appbar_actions.dart
import 'package:smlaicloud/model/currency_model.dart';
// CurrencyApiService ถูกย้ายไป po_currency_handler.dart
import 'package:smlaicloud/screens/transaction/utils/transaction_currency_utils.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_ui_utils.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_helper.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_product_search_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_currency_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_approval_status_badge.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_appbar_actions.dart';
// ==== PO Refactored Components ====
import 'package:smlaicloud/screens/purchaseorder/utils/po_detail_command_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_form_controller.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_workflow_manager.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_responsive_layout.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_bloc_effect_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_cart_integration_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_save_helper.dart';
import 'package:smlaicloud/screens/purchaseorder/models/po_warehouse_state.dart';
import 'package:smlaicloud/screens/purchaseorder/models/po_model_factory.dart';

enum PurchaseOrderEditScreenModule { header, detail, footer }

class PurchaseOrderEditScreen extends StatefulWidget {
  const PurchaseOrderEditScreen({super.key});

  @override
  State<PurchaseOrderEditScreen> createState() => PurchaseOrderEditScreenState();
}

class PurchaseOrderEditScreenState extends State<PurchaseOrderEditScreen> with TickerProviderStateMixin {
  // PO ใช้ type นี้เสมอ
  final global.TransactionTypeEnum transactionType = global.TransactionTypeEnum.purchaseorder;
  late SplitViewController splitViewController;
  late TabController tabController;
  late TabController editTabController;
  late TransactionModel screenData;
  late TransactionModel screenDataTemp;
  late PurchaseOrderEditScreenModule moduleEdit;
  final ProductBarcodeRepository productBarcodeRepository = ProductBarcodeRepository();

  /// Helper instance สำหรับ dialog functions
  late final TransactionDialogs _transactionDialogs = TransactionDialogs(this);

  /// Helper instance สำหรับ calculator functions
  late final TransactionCalculatorBridge _transactionCalculator = TransactionCalculatorBridge(this);

  /// Public getter สำหรับให้ helper classes เข้าถึง calculator
  TransactionCalculator get transactionCalculator => _transactionCalculator;

  /// Helper instance สำหรับ search handler functions
  late final TransactionSearchHandlers _searchHandlers = TransactionSearchHandlers(this);

  /// Public getter/setter สำหรับ showDocSearchPanel (ให้ helper class เข้าถึงได้)
  bool get showDocSearchPanel => _showDocSearchPanel;
  set showDocSearchPanel(bool value) => _showDocSearchPanel = value;

  // ==== Form Controllers - ใช้ POFormController รวมศูนย์ ====
  final POFormController _formController = POFormController();

  // ==== Workflow Manager - จัดการ Save/Print/Approval flow ====
  POWorkflowManager get _workflowManager => POWorkflowManager(
        context: context,
        isMounted: () => mounted,
        getDocNo: () => screenData.docno,
        getCollectionName: _getCollectionNameForType,
        getDocumentTitle: _getDocumentTitle,
        submitPOForApproval: _submitPOForApproval,
        onClearScreen: clearScreenData,
        onShowSaveDialog: (docNo) => showSaveDocNoDialog(context, docNo),
        onRegenerateDocNoAndSave: _regenerateDocNoAndSave,
        editTabController: editTabController,
        checkIsMultiCurrency: _checkIsMultiCurrency,
      );

  // ==== Bloc Effect Handler - จัดการ Bloc states (Trans, Warehouse, ExportCsv) ====
  POBlocEffectHandler get _blocEffectHandler => POBlocEffectHandler(
        isMounted: () => mounted,
        onStateChanged: () => setState(() {}),
        setLoading: (isLoading) => setState(() => _isLoading = isLoading),
        getScreenData: () => screenData,
        setScreenData: (trans) => screenData = trans,
        setScreenDataTemp: (trans) => screenDataTemp = TransactionModel.fromJson(trans.toJson()),
        onClearScreen: clearScreenData,
        onShowSaveDialog: (docNo) => showSaveDocNoDialog(context, docNo),
        onSubmitPOAndHandlePrint: _submitPOAndHandlePrint,
        onSubmitPOAndHandleUpdatePrint: _submitPOAndHandleUpdatePrint,
        onHandleSaveSuccessWithPrintCheck: _handleSaveSuccessWithPrintCheck,
        onHandleUpdateSuccessWithPrintCheck: (transType) => _workflowManager.handleUpdateSuccessWithPrintCheck(transType),
        onRefreshListAfterSave: _refreshListAfterSave,
        onRefreshListAfterDelete: _refreshListAfterDelete,
        onLoadPurchaseTypesAndExisting: _loadPurchaseTypesAndExisting,
        onLoadCurrencies: _loadCurrencies,
        onLoadPOApprovalStatus: _loadPOApprovalStatus,
        onShowDuplicateDocNoDialog: _showDuplicateDocNoDialog,
        onWarehouseListLoaded: (warehouses) => warehouseList = warehouses,
        onSetDefaultWarehouse: _setDefaultWarehouse,
        transactionType: transactionType,
        showDocSearchPanel: _showDocSearchPanel,
      );

  /// Helper สำหรับตั้งค่า warehouse defaults (ใช้โดย _blocEffectHandler)
  void _setDefaultWarehouse({
    required String defaultWarehouse,
    required List<LanguageDataModel> defaultWarehouseNames,
    required String defaultLocation,
    required List<LanguageDataModel> defaultLocationNames,
    required String defaultToWarehouse,
    required List<LanguageDataModel> defaultToWarehouseNames,
    required String defaultToLocation,
    required List<LanguageDataModel> defaultToLocationNames,
  }) {
    setState(() {
      _warehouseState.update(
        warehouse: defaultWarehouse,
        warehouseNames: defaultWarehouseNames,
        location: defaultLocation,
        locationNames: defaultLocationNames,
        toWarehouse: defaultToWarehouse,
        toWarehouseNames: defaultToWarehouseNames,
        toLocation: defaultToLocation,
        toLocationNames: defaultToLocationNames,
      );
    });
  }

  /// Refresh list หลัง delete (ใช้โดย _blocEffectHandler)
  void _refreshListAfterDelete(String deletedGuid) {
    if (_showDocSearchPanel && deletedGuid.isNotEmpty && _poListKey.currentState != null) {
      _poListKey.currentState!.refreshListAfterDelete(deletedGuid);
    }
  }

  /// Refresh list หลัง save/update (ใช้โดย _blocEffectHandler)
  void _refreshListAfterSave() {
    if (!_showDocSearchPanel) return;

    final guidToSelect = screenData.guidfixed;
    if (guidToSelect != null && _poListKey.currentState != null) {
      _poListKey.currentState!.setCardLoadingState(guidToSelect);
    }
    Future.delayed(const Duration(milliseconds: 2000), () {
      if (mounted && _poListKey.currentState != null) {
        _poListKey.currentState!.refreshListAndMaintainPosition(
          selectDocGuid: guidToSelect,
        );
      }
    });
  }

  // Getters สำหรับเข้าถึง controllers (backward compatibility)
  TextEditingController get docNumberController => _formController.docNumberController;
  TextEditingController get docDateController => _formController.docDateController;
  TextEditingController get docTimeController => _formController.docTimeController;
  TextEditingController get docRefNumberController => _formController.docRefNumberController;
  TextEditingController get docRefDateController => _formController.docRefDateController;
  TextEditingController get taxDocNoController => _formController.taxDocNoController;
  TextEditingController get taxDocDateController => _formController.taxDocDateController;
  TextEditingController get docRefTypeController => _formController.docRefTypeController;
  TextEditingController get docTypeController => _formController.docTypeController;
  TextEditingController get vatTypeController => _formController.vatTypeController;
  TextEditingController get custCodeController => _formController.custCodeController;
  TextEditingController get custnamesController => _formController.custnamesController;
  TextEditingController get saleCodeController => _formController.saleCodeController;
  TextEditingController get saleNameController => _formController.saleNameController;
  TextEditingController get discountWordController => _formController.discountWordController;
  TextEditingController get totalCostController => _formController.totalCostController;
  TextEditingController get totalValueController => _formController.totalValueController;
  TextEditingController get totalDiscountController => _formController.totalDiscountController;
  TextEditingController get totalVatValueController => _formController.totalVatValueController;
  TextEditingController get totalAfterVatController => _formController.totalAfterVatController;
  TextEditingController get totalExceptVatController => _formController.totalExceptVatController;
  TextEditingController get totalAmountController => _formController.totalAmountController;
  TextEditingController get vatRateController => _formController.vatRateController;
  TextEditingController get statusController => _formController.statusController;
  TextEditingController get docDiscountWordController => _formController.docDiscountWordController;
  TextEditingController get descriptionController => _formController.descriptionController;
  TextEditingController get detailTotalDiscount => _formController.detailTotalDiscount;
  TextEditingController get totalDiscountVatAmount => _formController.totalDiscountVatAmount;
  TextEditingController get totalDiscountExceptVatAmount => _formController.totalDiscountExceptVatAmount;
  TextEditingController get roundAmount => _formController.roundAmount;
  TextEditingController get totalAmountAfterDiscount => _formController.totalAmountAfterDiscount;
  TextEditingController get detailDiscountFormula => _formController.detailDiscountFormula;
  TextEditingController get detailTotalAmount => _formController.detailTotalAmount;
  TextEditingController get roundAmountController => _formController.roundAmountController;
  TextEditingController get detaildiscountformulaController => _formController.detaildiscountformulaController;
  TextEditingController get transportAmountController => _formController.transportAmountController;

  late DateTime selectedDate = DateTime.now();
  late List<WarehouseModel> warehouseList = [];

  // ==== Warehouse State - รวม 8 ตัวแปร warehouse/location ====
  final POWarehouseState _warehouseState = POWarehouseState();

  // Getters สำหรับ backward compatibility (warehouse)
  String get defaultWarehouse => _warehouseState.defaultWarehouse;
  List<LanguageDataModel> get defaultWarehouseNames => _warehouseState.defaultWarehouseNames;
  String get defaultLocation => _warehouseState.defaultLocation;
  List<LanguageDataModel> get defaultLocationNames => _warehouseState.defaultLocationNames;
  String get defaultToWarehouse => _warehouseState.defaultToWarehouse;
  List<LanguageDataModel> get defaultToWarehouseNames => _warehouseState.defaultToWarehouseNames;
  String get defaultToLocation => _warehouseState.defaultToLocation;
  List<LanguageDataModel> get defaultToLocationNames => _warehouseState.defaultToLocationNames;
  List<Widget> tab = [];
  List<Widget> childrens = [];
  List<String> groupedItems = [];
  int transflag = 0;
  int calcflag = 0;
  bool docDateTimeValidated = false;
  late String fieldNameCustCode;
  late String fieldNameCustName;
  final _debouncer = global.Debouncer(1000);
  String docnoFormat = "";
  int configvattype = 0;
  int configinquirytype = 0;
  String enumtype = '';
  bool _isLoading = false;
  bool _showPreview = false; // ค่าเริ่มต้นไม่ต้องแสดง
  bool _hasCheckedScreenSize = false; // ตรวจสอบว่าได้ตรวจสอบขนาดหน้าจอแล้วหรือยัง
  bool _showDocSearchPanel = false; // แสดง side panel ค้นหาเอกสาร (จอใหญ่)
  double _docSearchPanelWidth = 550.0; // ความกว้าง side panel (ปรับขนาดได้) - เพิ่มขึ้นเพื่อแสดงข้อมูลได้มากขึ้น

  // GlobalKey สำหรับเข้าถึง PurchaseOrderListScreen state เพื่อ refresh list หลังจาก save/delete
  final GlobalKey<PurchaseOrderListScreenState> _poListKey = GlobalKey<PurchaseOrderListScreenState>();

  // Key สำหรับ SharedPreferences เก็บค่า printAfterSave (ใช้ SharedPreferences แทน instance variable เพื่อป้องกัน Hot Reload reset)
  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  //  pricelevel ของลูกค้า
  int customerPriceLevel = 1; // ค่าเริ่มต้นเป็น 1

  String? clientId;

  List<global.DataTableHeader> headers = [
    global.DataTableHeader(code: "delete", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
    global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 10, textAlign: TextAlign.center, alignment: Alignment.center),
    global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 20),
    global.DataTableHeader(code: "product_name", label: global.language('product_name'), width: 40),
    global.DataTableHeader(code: "product_ware_house", label: global.language('product_ware_house'), width: 15),
    global.DataTableHeader(code: "product_location", label: global.language('product_location'), width: 10),
    global.DataTableHeader(code: "product_unit", label: global.language('product_unit'), width: 10),
    global.DataTableHeader(code: "product_qty", label: global.language('product_qty'), width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right),
    global.DataTableHeader(code: "product_price", label: global.language('product_price'), alignment: Alignment.centerRight, width: 20, textAlign: TextAlign.right),
    global.DataTableHeader(code: "product_discount", label: global.language('product_discount'), width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right),
    global.DataTableHeader(code: "product_amount", label: global.language('product_amount'), alignment: Alignment.centerRight, width: 30, textAlign: TextAlign.right),
  ];

  List<dynamic> docrefs = [];

  // ==== PO Approval Handler - จัดการ purchase types และ approval status ====
  late final POApprovalHandler _approvalHandler = POApprovalHandler(
    onStateChanged: () => setState(() {}),
    onShowSuccess: (msg) => context.ui.showSuccess(msg),
    onShowInfo: (msg) => context.ui.showInfo(msg),
    onShowError: (msg) => context.ui.showError(msg),
  );

  // Getters สำหรับ backward compatibility (purchase types)
  List<PurchaseTypeModel> get _purchaseTypes => _approvalHandler.purchaseTypes;
  String? get _selectedPurchaseTypeCode => _approvalHandler.selectedPurchaseTypeCode;

  // สกุลเงิน (สำหรับ PO Multi-Currency) - ใช้ POCurrencyHandler
  late final POCurrencyHandler _currencyHandler = POCurrencyHandler(
    onStateChanged: () => setState(() {}),
    onWarning: (msg) => context.ui.showWarning(msg),
    onRecalculate: () => _transactionCalculator.calTotalValue(),
  );

  // Getters สำหรับเข้าถึงข้อมูลจาก handler (backward compatibility)
  List<CurrencyModel> get _currencies => _currencyHandler.currencies;
  String? get _selectedDocCurrencyCode => _currencyHandler.docCurrencyCode;
  String? get _selectedDocCurrencySymbol => _currencyHandler.docCurrencySymbol;
  double? get _selectedExchangeRate => _currencyHandler.exchangeRate;
  String? get _baseCurrency => _currencyHandler.baseCurrency;
  String? get _baseCurrencySymbol => _currencyHandler.baseCurrencySymbol;

  // Getters สำหรับ backward compatibility (approval status)
  POApprovalStatusModel? get _poApprovalStatus => _approvalHandler.approvalStatus;
  bool get _isLoadingApprovalStatus => _approvalHandler.isLoadingApprovalStatus;

  void setSystemLanguageList() async {
    clearScreenData();
    await global.setSystemLanguage(context);

    headerTableDetail();
  }

  @override
  void initState() {
    super.initState();

    global.getDeviceModel(context);
    // PO ใช้ supplier (ผู้จำหน่าย)
    fieldNameCustCode = "doc_supplier_code";
    fieldNameCustName = "doc_supplier_name";

    setSystemLanguageList();
    splitViewController = SplitViewController(limits: [null, WeightLimit(min: 0.1, max: 0.9)]);
    splitViewController.weights = [0.30, 0.70];
    tabController = TabController(length: 2, vsync: this);

    // PO ใช้ 2 tabs
    editTabController = TabController(length: 2, vsync: this);

    context.read<WarehouseBloc>().add(const WarehouseLoadList(offset: 0, limit: 100, search: ''));

    // ตรวจสอบขนาดหน้าจอเพื่อตั้งค่า _showPreview (ครั้งเดียวเท่านั้น)
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_hasCheckedScreenSize) {
        final screenWidth = MediaQuery.of(context).size.width;
        // ถ้าเป็นจอ tablet หรือ desktop (มากกว่า 600px) ให้แสดง preview
        if (screenWidth > 600) {
          setState(() {
            // ปรับ tab controller length สำหรับ mobile mode
            tabController.dispose();
            tabController = TabController(length: _showPreview ? 2 : 1, vsync: this);
          });
        }
        _hasCheckedScreenSize = true; // ทำเครื่องหมายว่าได้ตรวจสอบแล้ว
      }
    });
  }

  bool _hasProcessedCartTransfer = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();

    // ตรวจสอบและประมวลผล CartTransferModel (เพียงครั้งเดียว)
    if (!_hasProcessedCartTransfer) {
      CartTransferModel? cartTransfer;

      // 1. ตรวจสอบจาก CartTransferProvider (สำหรับ tab ใหม่)
      cartTransfer = CartTransferProvider.of(context);

      // 2. ถ้าไม่มี ลองตรวจสอบจาก route arguments (สำหรับ Navigator.pushNamed)
      if (cartTransfer == null) {
        final args = ModalRoute.of(context)?.settings.arguments;
        if (args is CartTransferModel) {
          cartTransfer = args;
        }
      }

      // ถ้าพบข้อมูล ให้โหลดเข้าระบบ
      if (cartTransfer != null) {
        AppLogger.info('📦 รับข้อมูลจาก Cart System: ${cartTransfer.cartName}');
        _populateFromCartTransfer(cartTransfer);
        _hasProcessedCartTransfer = true;
      }
    }
  }

  /// แปลงข้อมูลจาก CartTransferModel มาเป็น TransactionModel
  /// Delegate to POCartIntegrationHandler
  void _populateFromCartTransfer(CartTransferModel cartTransfer) {
    try {
      screenData = POCartIntegrationHandler.populateFromCart(
        screenData: screenData,
        cartTransfer: cartTransfer,
        calcflag: calcflag,
      );

      // โหลดข้อมูลไปยังหน้าจอ
      setState(() {
        loadDataToScreen();
      });
    } catch (e, stackTrace) {
      AppLogger.error('❌ Error populating from cart transfer: $e\n$stackTrace');
      if (mounted) {
        context.ui.showError('เกิดข้อผิดพลาดในการโหลดข้อมูลจากตระกร้า: $e');
      }
    }
  }

  @override
  void dispose() {
    // เคลียร์ข้อมูลที่อาจทำให้ memory leak
    _debouncer.cancel();
    docrefs.clear();
    splitViewController.dispose();
    tabController.dispose();
    editTabController.dispose();
    // Dispose form controllers ผ่าน POFormController
    _formController.dispose();

    super.dispose();
  }

  void headerTableDetail() {
    // PO screen - ใช้ headers สำหรับเอกสารจัดซื้อ
    headers = [
      global.DataTableHeader(code: "delete", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
      global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 10, textAlign: TextAlign.center, alignment: Alignment.center),
      global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 20),
      global.DataTableHeader(code: "item_code", label: global.language('item_code'), width: 15),
      global.DataTableHeader(code: "product_name", label: global.language('product_name'), width: 40),
    ];

    headers.add(global.DataTableHeader(code: "product_ware_house", label: global.language('product_ware_house'), width: 10));
    headers.add(global.DataTableHeader(code: "product_location", label: global.language('product_location'), width: 10));
    headers.addAll([
      global.DataTableHeader(code: "product_unit", label: global.language('product_unit'), width: 10),
      global.DataTableHeader(code: "product_qty", label: global.language('product_qty'), width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right),
      global.DataTableHeader(code: "product_price", label: global.language('product_price'), alignment: Alignment.centerRight, width: 15, textAlign: TextAlign.right),
      global.DataTableHeader(code: "product_discount", label: global.language('product_discount'), width: 15, alignment: Alignment.centerRight, textAlign: TextAlign.right),
      global.DataTableHeader(code: "product_amount", label: global.language('product_amount'), alignment: Alignment.centerRight, width: 15, textAlign: TextAlign.right),
    ]);
  }

  // Helper function สำหรับ parse date อย่างปลอดภัย
  DateTime _safeParseDatetime(String? dateString) {
    if (dateString == null || dateString.isEmpty) {
      return DateTime.now();
    }
    try {
      return DateTime.parse(dateString);
    } catch (e) {
      AppLogger.warning('[PO] แปลงวันที่ไม่สำเร็จ: "$dateString" → ใช้ DateTime.now() แทน');
      return DateTime.now();
    }
  }

  void loadDataToScreen() {
    headerTableDetail();
    // ใช้ POFormController เพื่อโหลดข้อมูลจาก screenData ไปยัง controllers
    _formController.loadDataFromModel(screenData);
    _transactionCalculator.calPayTotal();
  }

  /// Helper method สำหรับดึงชื่อ collection (สำหรับ PDF API)
  /// PO screen ใช้ค่าคงที่ "transactionPurchaseOrder"
  String _getCollectionNameForType(global.TransactionTypeEnum type) {
    return "transactionPurchaseOrder";
  }

  /// Helper method สำหรับดึงชื่อเอกสาร (สำหรับ PDF title)
  /// PO screen ใช้ค่าคงที่ "purchase_order"
  String _getDocumentTitle(global.TransactionTypeEnum type) {
    return global.language("purchase_order");
  }

  /// ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ
  /// Delegate to POWorkflowManager - ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ (Save)
  Future<void> _submitPOAndHandlePrint(String docNo, global.TransactionTypeEnum transType) async {
    await _workflowManager.submitAndHandlePrint(docNo, transType);
  }

  /// Delegate to POWorkflowManager - ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ (Update)
  Future<void> _submitPOAndHandleUpdatePrint(global.TransactionTypeEnum transType) async {
    await _workflowManager.submitAndHandleUpdatePrint(transType);
  }

  /// Delegate to POWorkflowManager - จัดการหลังบันทึกสำเร็จ
  Future<void> _handleSaveSuccessWithPrintCheck(String docNo, global.TransactionTypeEnum transType) async {
    await _workflowManager.handleSaveSuccessWithPrintCheck(docNo, transType);
  }

  /// Delegate to POWorkflowManager - แสดง dialog เมื่อ docno ซ้ำ
  void _showDuplicateDocNoDialog() {
    _workflowManager.showDuplicateDocNoDialog(screenData.docno);
  }

  /// บันทึกอีกครั้งโดยให้ mainapi สร้าง docno ใหม่
  void _regenerateDocNoAndSave() {
    setState(() {
      // ล้าง docno เพื่อให้ mainapi สร้างใหม่
      screenData.docno = '';
      // อัพเดท timezone info ใหม่
      final timezoneInfo = global.getLocalTimezoneInfo();
      screenData.docdatelocal = timezoneInfo['docdatelocal'];
      screenData.doctimelocal = timezoneInfo['doctimelocal'];
      screenData.timezone = timezoneInfo['timezone'];
      AppLogger.info('🔄 ล้าง docno เพื่อให้ mainapi สร้างใหม่ (docdatelocal: ${screenData.docdatelocal})');
    });

    // บันทึกอีกครั้ง - mainapi จะสร้าง docno ใหม่เอง
    saveOrUpdateData(taxInvoice: false);
  }

  void clearScreenData() {
    // PO: ค่าคงที่สำหรับ Purchase Order
    transflag = 6;
    calcflag = 1;
    docnoFormat = "PO";
    configvattype = global.config.vattypepurchase;
    configinquirytype = global.config.inquirytypepurchase;
    moduleEdit = PurchaseOrderEditScreenModule.header;

    // สร้าง TransactionModel เปล่าผ่าน Factory
    screenData = POModelFactory.createForPO();
    screenDataTemp = POModelFactory.createForPO();

    // รีเซ็ต customerPriceLevel เป็นค่าเริ่มต้น
    customerPriceLevel = 1;

    // รีเซ็ตประเภทการจัดซื้อและสถานะการอนุมัติ (สำหรับ PO) - delegate to handler
    _approvalHandler.resetPurchaseType();
    _approvalHandler.resetApprovalStatus();

    setState(() {
      loadDataToScreen();
      docDateTimeValidated = true;
    });

    // โหลดประเภทการจัดซื้อและสกุลเงินสำหรับ PO
    _loadPurchaseTypesForPO();
    _loadCurrencies();
  }

  /// โหลดรายการประเภทการจัดซื้อสำหรับ PO - delegate to POApprovalHandler
  Future<void> _loadPurchaseTypesForPO() async {
    if (!mounted) return;
    await _approvalHandler.loadPurchaseTypesForNewPO();
  }

  /// โหลด list ประเภทการจัดซื้อ และค่าที่บันทึกไว้ (สำหรับโหลดเอกสารที่มีอยู่)
  /// Delegate to POApprovalHandler
  Future<void> _loadPurchaseTypesAndExisting(String docNo) async {
    if (!mounted) return;
    await _approvalHandler.loadPurchaseTypesAndExisting(screenData);
  }

  /// บันทึกประเภทการจัดซื้อของ PO ลงใน TransactionModel - delegate to POApprovalHandler
  void _savePOPurchaseType() {
    _approvalHandler.savePurchaseTypeToModel(screenData);
    // บันทึกสกุลเงินลงใน TransactionModel
    _saveCurrencyData();
  }

  /// บันทึกข้อมูลสกุลเงินลงใน TransactionModel
  void _saveCurrencyData() {
    if (_selectedDocCurrencyCode == null || _selectedDocCurrencyCode!.isEmpty) return;
    
    // ใช้ TransactionCurrencyUtils สำหรับ initialize
    TransactionCurrencyUtils.initializeWithBaseCurrency(
      screenData,
      _baseCurrency!,
      _baseCurrencySymbol!,
    );
    
    // ถ้าไม่ใช่ base currency ตั้งค่า doc currency
    if (_selectedDocCurrencyCode != _baseCurrency) {
      screenData.docCurrency = _selectedDocCurrencyCode;
      screenData.docCurrencySymbol = _selectedDocCurrencySymbol;
      double rate = _selectedExchangeRate ?? 1.0;
      if (rate <= 0 || rate.isNaN || rate.isInfinite) {
        AppLogger.warning('[PO] Exchange rate ไม่ถูกต้อง: $_selectedExchangeRate → ใช้ 1.0 แทน');
        rate = 1.0;
      }
      screenData.exchangerate = rate;
    }
    
    AppLogger.info('[PO] บันทึกสกุลเงิน: doc=${screenData.docCurrency}, base=${screenData.currency}, อัตรา=${screenData.exchangerate}');
  }

  /// เปลี่ยนประเภทการจัดซื้อที่เลือก - delegate to POApprovalHandler
  void _onPurchaseTypeChanged(String? code, String? name) {
    _approvalHandler.onPurchaseTypeChanged(code, name);
  }

  /// โหลดรายการสกุลเงิน (สำหรับ PO Multi-Currency)
  /// 
  /// รองรับกรณีต่างๆ:
  /// 1. เอกสารใหม่ - ใช้ Base Currency เป็นค่าเริ่มต้น
  /// 2. Legacy Document (มีแค่ currency ไม่มี docCurrency) - copy ค่า
  /// 3. Multi-Currency Document (มี docCurrency) - ใช้ค่าที่มี
  /// ตรวจสอบว่าเอกสารปัจจุบันเป็น Multi-Currency หรือไม่ (3 ระดับ)
  bool _checkIsMultiCurrency() {
    return _currencyHandler.isMultiCurrency ||
        (screenData.docCurrency != null &&
         screenData.docCurrency!.isNotEmpty &&
         screenData.docCurrency!.toUpperCase() != global.getBaseCurrency()) ||
        (screenData.exchangerate != null &&
         screenData.exchangerate != 0 &&
         screenData.exchangerate != 1.0);
  }

  /// โหลดรายการสกุลเงิน - delegate ไปยัง POCurrencyHandler
  Future<void> _loadCurrencies() async {
    await _currencyHandler.loadCurrencies(screenData);
  }

  /// เปลี่ยนสกุลเงินที่เลือก - delegate ไปยัง POCurrencyHandler
  void _onCurrencyChanged(String? code, String? symbol, double? exchangeRate) async {
    await _currencyHandler.onCurrencyChanged(code, symbol, exchangeRate, screenData);
  }

  /// แก้ไขอัตราแลกเปลี่ยน - delegate ไปยัง POCurrencyHandler
  void _onExchangeRateChanged(double? exchangeRate) {
    _currencyHandler.onExchangeRateChanged(exchangeRate, screenData);
  }

  /// โหลดสถานะการอนุมัติของ PO - delegate to POApprovalHandler
  Future<void> _loadPOApprovalStatus(String docNo) async {
    if (!mounted) return;
    await _approvalHandler.loadApprovalStatus(docNo);
  }

  /// ส่ง PO เข้าระบบอนุมัติ - delegate to POApprovalHandler
  /// [isModified] - true เมื่อเป็นการแก้ไขเอกสารที่เคยส่งอนุมัติแล้ว
  /// คืนค่า true ถ้าเป็น auto_approved (สามารถพิมพ์ได้ทันที)
  Future<bool> _submitPOForApproval({bool isModified = false}) async {
    // Sync หมายเหตุจาก controller ไปยัง screenData (user อาจพิมพ์ค่าใหม่)
    screenData.description = descriptionController.text;
    return await _approvalHandler.submitForApproval(
      screenData: screenData,
      description: screenData.description ?? '',
      isModified: isModified,
    );
  }

  /// ถอนการส่งอนุมัติ PO - delegate to POApprovalHandler
  Future<bool> _withdrawPOApproval() async {
    return await _approvalHandler.withdrawApproval(
      docNo: screenData.docno,
    );
  }

  // =====================================================
  // Helper methods สำหรับตรวจสอบสิทธิ์พิมพ์/แก้ไข PO ตามสถานะการอนุมัติ
  // =====================================================

  /// ดึงข้อมูลคลังสินค้าเริ่มต้นสำหรับ POProductSearchHandler
  POWarehouseInfo _getWarehouseInfo() {
    return POWarehouseInfo(
      defaultWarehouse: defaultWarehouse,
      defaultWarehouseNames: defaultWarehouseNames,
      defaultLocation: defaultLocation,
      defaultLocationNames: defaultLocationNames,
      defaultToWarehouse: defaultToWarehouse,
      defaultToWarehouseNames: defaultToWarehouseNames,
      defaultToLocation: defaultToLocation,
      defaultToLocationNames: defaultToLocationNames,
    );
  }

  /// ตรวจสอบว่า PO สามารถพิมพ์ได้หรือไม่
  /// ใช้ POApprovalHelper เพื่อ reuse logic
  bool _canPrintPO() {
    return POApprovalHelper.canPrint(_poApprovalStatus);
  }

  /// ตรวจสอบว่า PO สามารถแก้ไขได้หรือไม่
  /// IMPORTANT: เอกสารใหม่ (docno ว่าง) ต้องแก้ไขได้เสมอ
  bool _canEditPO() {
    // Defensive check: เอกสารใหม่ที่ยังไม่มี docno = แก้ไขได้เสมอ
    if (screenData.docno.isEmpty) {
      return true;
    }
    return POApprovalHelper.canEdit(
      docNo: screenData.docno,
      approvalStatus: _poApprovalStatus,
      isref: screenData.isref ?? false,
      creatorCode: screenData.creatorcode,
      currentUserCode: global.profileData.username,
    );
  }

  /// ข้อความแสดงเหตุผลที่ไม่สามารถแก้ไขได้
  /// IMPORTANT: เอกสารใหม่ (docno ว่าง) ไม่มีเหตุผลที่จะห้ามแก้ไข
  String _getPOEditDisabledReason() {
    // Defensive check: เอกสารใหม่ที่ยังไม่มี docno = ไม่มีเหตุผลที่จะห้าม
    if (screenData.docno.isEmpty) {
      return '';
    }
    return POApprovalHelper.getEditDisabledReason(
      docNo: screenData.docno,
      approvalStatus: _poApprovalStatus,
      isref: screenData.isref ?? false,
      creatorCode: screenData.creatorcode,
      currentUserCode: global.profileData.username,
    );
  }

  Future<BookBankModel?> bookBankSearch() async {
    Completer<BookBankModel?> completer = Completer<BookBankModel?>();

    Navigator.push(context, MaterialPageRoute(builder: (context) => const BookBankSelectScreen())).then((value) {
      completer.complete(value);
    });

    return completer.future;
  }

  // void _calTotalCharge(int index) {
  //   if (screenData.paymentdetail!.paymentcreditcards![index].chargeword != '') {
  //     if (screenData.paymentdetail!.paymentcreditcards![index].chargeword!.contains('%')) {
  //       double chargeValue = screenData.paymentdetail!.paymentcreditcards![index].amount! *
  //           (double.parse(screenData.paymentdetail!.paymentcreditcards![index].chargeword!.replaceAll('%', '')) / 100);
  //       screenData.paymentdetail!.paymentcreditcards![index].chargevalue = chargeValue;
  //       screenData.paymentdetail!.paymentcreditcards![index].totalnetworth = screenData.paymentdetail!.paymentcreditcards![index].amount! + chargeValue;
  //     } else {
  //       screenData.paymentdetail!.paymentcreditcards![index].chargevalue = double.parse(screenData.paymentdetail!.paymentcreditcards![index].chargeword!);
  //       screenData.paymentdetail!.paymentcreditcards![index].totalnetworth =
  //           screenData.paymentdetail!.paymentcreditcards![index].amount! + double.parse(screenData.paymentdetail!.paymentcreditcards![index].chargeword!);
  //     }
  //   } else {
  //     screenData.paymentdetail!.paymentcreditcards![index].chargeword = '0';
  //     screenData.paymentdetail!.paymentcreditcards![index].chargevalue = 0;
  //     screenData.paymentdetail!.paymentcreditcards![index].totalnetworth = screenData.paymentdetail!.paymentcreditcards![index].amount!;
  //   }
  //   setState(() {});
  // }

  Widget docPreview() {
    return DocumentPreviewWidget(
      screenData: screenData,
      transactionType: transactionType,
      docDateTimeValidated: docDateTimeValidated,
      fieldNameCustCode: fieldNameCustCode,
      fieldNameCustName: fieldNameCustName,
    );
  }

  double getPrice(List<PriceDataModel>? result) {
    // PO: ราคาจะถูกกรอกโดยผู้ใช้ ไม่ดึงจากฐานข้อมูล
    return 0;
  }

  Widget previewWidget() {
    return Container(
      margin: const EdgeInsets.all(5),
      padding: const EdgeInsets.all(10),
      height: 100,
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withValues(alpha: 0.5),
            spreadRadius: 5,
            blurRadius: 7,
            offset: const Offset(0, 3), // changes position of shadow
          ),
        ],
      ),
      child: docPreview(),
    );
  }

  void togglePreview() {
    setState(() {
      _showPreview = !_showPreview;
      // ปรับ tab controller length สำหรับ mobile mode
      tabController.dispose();
      tabController = TabController(length: _showPreview ? 2 : 1, vsync: this);
    });
  }

  Widget editProductListWidget() {
    return DocumentProductListWidget(
      screenData: screenData,
      transactionType: transactionType,
      headers: headers,
      warehouseList: warehouseList,
      defaultWarehouse: defaultWarehouse,
      defaultWarehouseNames: defaultWarehouseNames,
      defaultLocation: defaultLocation,
      defaultLocationNames: defaultLocationNames,
      defaultToWarehouse: defaultToWarehouse,
      defaultToWarehouseNames: defaultToWarehouseNames,
      defaultToLocation: defaultToLocation,
      defaultToLocationNames: defaultToLocationNames,
      calcflag: calcflag,
      cartList: [], // PO ไม่ใช้ cart system
      setState: setState,
      context: context,
      docrefs: docrefs,
      showWareHouseDefualtDialog: showWareHouseDefualtDialog,
      showWareHouseLocationDefualtDialog: showWareHouseLocationDefualtDialog,
      showDialogCommand: showDialogCommand,
      deleteItemDetail: deleteItemDetail,
      calTotalValue: _transactionCalculator.calTotalValue,
      getPrice: getPrice,
      showBarcodeDialog: () => _transactionDialogs.showBarcodeDialog(context),
      onWarehouseChanged:
          (
            String newDefualtwarehouse,
            List<LanguageDataModel> newDefualtwarehousenames,
            String newDefualtlocation,
            List<LanguageDataModel> newDefualtlocationnames,
            String newDefualttowarehouse,
            List<LanguageDataModel> newDefualttowarehousenames,
            String newDefualttolocation,
            List<LanguageDataModel> newDefualttolocationnames,
          ) {
            setState(() {
              _warehouseState.update(
                warehouse: newDefualtwarehouse,
                warehouseNames: newDefualtwarehousenames,
                location: newDefualtlocation,
                locationNames: newDefualtlocationnames,
                toWarehouse: newDefualttowarehouse,
                toWarehouseNames: newDefualttowarehousenames,
                toLocation: newDefualttolocation,
                toLocationNames: newDefualttolocationnames,
              );
            });
          },
    );
  }

  void addIfNotExists(List list, var item) {
    if (!list.contains(item)) {
      list.add(item);
    }
  }

  Widget editDocumentWidget() {
    return DocumentHeaderWidget(
      screenData: screenData,
      transactionType: transactionType,
      setState: setState,
      context: context,
      custCodeController: custCodeController,
      custnamesController: custnamesController,
      saleCodeController: saleCodeController,
      saleNameController: saleNameController,
      docRefNumberController: docRefNumberController,
      taxDocNoController: taxDocNoController,
      vatRateController: vatRateController,
      descriptionController: descriptionController,
      transportAmountController: transportAmountController,
      docDateController: docDateController,
      docTimeController: docTimeController,
      docDateTimeValidated: docDateTimeValidated,
      debouncer: _debouncer,
      calTotalValue: _transactionCalculator.calTotalValue,
      headerTableDetail: headerTableDetail,
      searchCustomer: _searchHandlers.searchCustomer,
      searchSupplier: _searchHandlers.searchSupplier,
      searchSale: _searchHandlers.searchSale,
      searchSaleChannel: _searchHandlers.searchSaleChannel,
      searchDocRef: _searchHandlers.searchDocRef,
      onDeleteDocReference: deleteDocReference,
      // ประเภทการจัดซื้อ (สำหรับ PO)
      purchaseTypes: _purchaseTypes,
      selectedPurchaseTypeCode: _selectedPurchaseTypeCode,
      onPurchaseTypeChanged: _onPurchaseTypeChanged,
      // สกุลเงิน (สำหรับ PO Multi-Currency)
      currencies: _currencies,
      selectedDocCurrencyCode: _selectedDocCurrencyCode,
      selectedDocCurrencySymbol: _selectedDocCurrencySymbol,
      selectedExchangeRate: _selectedExchangeRate,
      onCurrencyChanged: _onCurrencyChanged,
      onExchangeRateChanged: _onExchangeRateChanged,
      baseCurrency: _baseCurrency,
      baseCurrencySymbol: _baseCurrencySymbol,
    );
  }

  Widget editWidget() {
    // PO: รวม approval status badge กับ document header
    Widget documentHeader = Column(
      children: [
        Expanded(child: editDocumentWidget()),
        // แสดง badge สถานะอนุมัติด้านล่างของเอกสาร
        POApprovalStatusBadge(
          screenData: screenData,
          approvalStatus: _poApprovalStatus,
          isLoading: _isLoadingApprovalStatus,
          onWithdrawApproval: _withdrawPOApproval,
        ),
      ],
    );
    // PO: ใช้ 2 tabs (doc_header, doc_details) - ไม่มี tab total
    List<Widget> childrenList = [documentHeader, editProductListWidget()];
    List<Widget> tabx = [Tab(text: global.language("doc_header")), Tab(text: global.language("doc_details"))];
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Focus(
          skipTraversal: true,
          canRequestFocus: false,
          child: TabBar(controller: editTabController, tabs: tabx),
        ),
      ),
      body: RawKeyboardListener(
        focusNode: FocusNode(skipTraversal: true),
        onKey: (event) async {
          if (kIsWeb || Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
            if (event is RawKeyUpEvent) {
              if (event.logicalKey == LogicalKeyboardKey.f10) {
                if (screenData.details!.isNotEmpty) {
                  // ตรวจสอบว่า PO สามารถแก้ไขได้หรือไม่
                  if (!_canEditPO()) {
                    context.ui.showWarning(_getPOEditDisabledReason());
                    return;
                  }
                  // ใช้ dialog แบบใหม่พร้อม checkbox เลือกพิมพ์
                  // ซ่อนตัวเลือกพิมพ์ถ้า PO ยังไม่ผ่านการอนุมัติ
                  final result = await showEnhancedSaveConfirmDialog(
                    context,
                    screenData.guidfixed ?? '',
                    showPrintOption: _canPrintPO(),
                  );
                  if (result != null && result.confirmed) {
                    // บันทึก flag ลง SharedPreferences (ป้องกัน Hot Reload reset)
                    final prefs = await SharedPreferences.getInstance();
                    await prefs.setBool(_printAfterSaveKey, result.printAfterSave);
                    AppLogger.debug('🖨️ [Save F10] printAfterSave saved to prefs: ${result.printAfterSave}');
                    saveOrUpdateData(taxInvoice: false);
                  }
                }
              }
            }
          }
        },
        child: Focus(
          skipTraversal: true,
          child: TabBarView(controller: editTabController, physics: const NeverScrollableScrollPhysics(), children: childrenList),
        ),
      ),
    );
  }

  void deleteDoc() {
    context.read<TransBloc>().add(TransDelete(guid: screenData.guidfixed!, type: transactionType));
  }

  void saveOrUpdateData({required bool taxInvoice}) {
    AppLogger.info('🔵 [saveOrUpdateData] START - taxInvoice: $taxInvoice, type: $transactionType');

    // คำนวณยอดรวมใหม่
    _transactionCalculator.calTotalValue();

    // ตรวจสอบความถูกต้องของการชำระเงิน
    if (!_transactionCalculator.verifyPayment()) {
      AppLogger.warn('🔴 [saveOrUpdateData] verifyPayment() returned FALSE - save blocked');
      return;
    }

    // เตรียมข้อมูลก่อนบันทึก (UTC, Line Numbers, Creator Info)
    // ใช้ profileData ซึ่งมี username/name จริงของผู้ login (userLoginData.code ว่างเสมอ)
    POSaveHelper.prepareDataForSave(
      screenData,
      creatorCode: global.profileData.username ?? '',
      creatorName: global.profileData.name ?? '',
    );

    // บันทึกประเภทการจัดซื้อและสกุลเงิน
    _savePOPurchaseType();

    // ส่ง Bloc Event ตามประเภทการบันทึก
    if (taxInvoice) {
      AppLogger.info('🔵 [saveOrUpdateData] Sending TransCreateFullInvoice event');
      context.read<TransBloc>().add(TransCreateFullInvoice(guid: screenData.guidfixed!, trans: screenData, type: transactionType));
    } else if (screenData.guidfixed?.isNotEmpty ?? false) {
      AppLogger.info('🔵 [saveOrUpdateData] Sending TransUpdate event');
      context.read<TransBloc>().add(TransUpdate(guid: screenData.guidfixed!, trans: screenData, type: transactionType));
    } else {
      AppLogger.info('🔵 [saveOrUpdateData] Sending TransSave event (new document)');
      context.read<TransBloc>().add(TransSave(trans: screenData, type: transactionType));
    }
  }

  /// จัดการคำสั่งจากตารางรายการสินค้า - delegate ไปยัง PODetailCommandHandler
  void showDialogCommand(String cmd, int index, TransactionDetailModel details) async {
    final handler = PODetailCommandHandler(
      context: context,
      screenData: screenData,
      warehouseList: warehouseList,
      warehouseInfo: _getWarehouseInfo(),
      getPrice: getPrice,
      onStateChanged: () => setState(() {}),
      onRecalculate: () => _transactionCalculator.calTotalValue(),
      calcflag: calcflag,
      defaultToWarehouse: defaultToWarehouse,
      defaultToWarehouseNames: defaultToWarehouseNames,
      defaultToLocation: defaultToLocation,
      defaultToLocationNames: defaultToLocationNames,
    );
    await handler.handleCommand(cmd, index, details);
  }

  void deleteItemDetail(int index) {
    setState(() {
      screenData.details!.removeAt(index);
    });

    Future.delayed(const Duration(milliseconds: 200), () {
      _transactionCalculator.calTotalValue();
    });
  }

  void deleteDocReference(int index, String docno) {
    // ลบเอกสารอ้างอิงจาก array
    screenData.docreferences!.removeAt(index);

    // ลบรายการสินค้าที่เกี่ยวข้องกับเอกสารอ้างอิงนี้
    screenData.details!.removeWhere((detail) => detail.docref == docno);

    // ลบจาก docrefs array
    docrefs.removeWhere((docref) => docref['docref'] == docno);

    // คำนวณยอดรวมใหม่
    _transactionCalculator.calTotalValue();
    if (mounted) {
      setState(() {});
    }

    // แสดงข้อความแจ้งเตือน
    context.ui.showSuccess('${global.language("delete_reference_and_products")} $docno');
  }

  Future<void> showSaveDocNoDialog(BuildContext context, String docno) async {
    FocusNode focusNode = FocusNode();

    // Show the dialog
    await showDialog(
      context: context,
      barrierDismissible: false, // Allows the dialog to be dismissed by tapping outside.
      builder: (BuildContext context) {
        // Listen for raw keyboard events
        focusNode.requestFocus();
        return RawKeyboardListener(
          focusNode: focusNode,
          onKey: (event) {
            if (event is RawKeyDownEvent && event.logicalKey == LogicalKeyboardKey.escape) {
              // ESC key is pressed
              Navigator.of(context).pop();
              clearScreenData();
              editTabController.animateTo(0);
            }
          },
          child: AlertDialog(
            title: Text(global.language("save_success")),
            content: Text("${global.language("docno")} : $docno  ${global.language("how_to_print")}"),
            actions: [
              TextButton(
                child: Text(global.language("cancel")),
                onPressed: () {
                  Navigator.of(context).pop();
                  clearScreenData();
                  editTabController.animateTo(0);
                },
              ),
              TextButton.icon(
                icon: Icon(Icons.print),
                label: Text(global.language("confirm")),
                onPressed: () async {
                  final dialogNavigator = Navigator.of(context);
                  final pdfService = PdfService();
                  await pdfService.generateAndOpenPdf(
                    collection: 'transactionPurchaseOrder',
                    docNo: screenData.docno,
                    title: global.transactionName(transactionType),
                    context: this.context,
                    // ตรวจ 3 ระดับ: 1) currencyHandler 2) docCurrency field 3) exchangerate != 1.0
                    isMultiCurrency: _checkIsMultiCurrency(),
                  );
                  // ปิด dialog หลังจากปิด PDF viewer แล้ว
                  if (mounted) {
                    dialogNavigator.pop();
                    clearScreenData();
                    editTabController.animateTo(0);
                  }
                },
              ),
            ],
          ),
        );
      },
    ).then((value) {
      focusNode.dispose();
    }); // Handle dialog dismiss by tapping outside or back button on Android
  }

  /// เมื่อเลือกเอกสารจาก side panel
  void _onDocumentSelectedFromPanel(TransactionModel document) {
    if (document.docno.isNotEmpty) {
      DateTime localDateTime = _safeParseDatetime(document.docdatetime);
      screenData = document;
      // Store the initial copy in screenDataTemp to keep it unchanged
      screenDataTemp = document.copyWith(
        details: document.details?.map((detail) => detail.copyWith()).toList(),
        paymentdetail: document.paymentdetail?.copyWith(),
        paymentStructs: document.paymentStructs?.map((struct) => struct.copyWith()).toList(),
      );
      // copyWith ไม่ copy isdelete (เป็น field นอก constructor) ต้อง set แยก
      screenDataTemp.isdelete = document.isdelete;

      /// sort linenumber (ถ้ามี details)
      if (screenData.details != null && screenData.details!.isNotEmpty) {
        screenData.details!.sort((a, b) => a.linenumber.compareTo(b.linenumber));
      }
      screenData.docdatetime = localDateTime.toUtc().toIso8601String(); // แปลงเป็น UTC ตามกฎ CLAUDE.md

      loadDataToScreen();
      // คำนวณยอดรวมใหม่
      _transactionCalculator.calTotalValue();

      // โหลดประเภทการจัดซื้อสำหรับ PO
      if (document.docno.isNotEmpty) {
        _loadPurchaseTypesAndExisting(document.docno);
        _loadCurrencies(); // โหลดรายการสกุลเงิน
        // โหลดสถานะการอนุมัติ PO สำหรับตรวจสอบ isModified
        _loadPOApprovalStatus(document.docno);
      }

      // ไม่ปิด panel เพื่อให้สามารถตรวจสอบเอกสารอื่นๆ ต่อได้
      setState(() {});
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            global.gotoMainMenu(context);
          },
        ),
        backgroundColor: global.theme.appBarColor,
        title: Text(global.transactionName(transactionType)),
        actions: POAppBarActions.build(
          context: context,
          screenData: screenData,
          screenDataTemp: screenDataTemp,
          transactionType: transactionType,
          isLoading: _isLoading,
          showPreview: _showPreview,
          canEditPO: _canEditPO(),
          canPrintPO: _canPrintPO(),
          editDisabledReason: _getPOEditDisabledReason(),
          getCollectionName: _getCollectionNameForType,
          getDocumentTitle: _getDocumentTitle,
          tabController: tabController,
          onClearScreen: clearScreenData,
          onTogglePreview: togglePreview,
          onSearchTrans: _searchHandlers.searchTrans,
          onDeleteDoc: deleteDoc,
          onSaveOrUpdate: saveOrUpdateData,
          onManualCloseToggle: () {
            setState(() {});
            _refreshListAfterSave();
          },
        ),
      ),
      body: Stack(
        children: [
          LayoutBuilder(
            builder: (context, constraints) {
              return MultiBlocListener(
                listeners: [
                  BlocListener<TransBloc, TransState>(
                    listener: _blocEffectHandler.handleTransState,
                  ),
                  BlocListener<WarehouseBloc, WarehouseState>(
                    listener: _blocEffectHandler.handleWarehouseState,
                  ),
                  BlocListener<ExportCsvBloc, ExportCsvState>(
                    listener: _blocEffectHandler.handleExportCsvState,
                  ),
                ],
                child: POResponsiveLayout(
                  splitViewController: splitViewController,
                  tabController: tabController,
                  showPreview: _showPreview,
                  showDocSearchPanel: _showDocSearchPanel,
                  docSearchPanelWidth: _docSearchPanelWidth,
                  poListKey: _poListKey,
                  buildEditWidget: editWidget,
                  buildPreviewWidget: previewWidget,
                  onDocumentSelected: _onDocumentSelectedFromPanel,
                  onClosePanel: () => setState(() => _showDocSearchPanel = false),
                  onPanelWidthChanged: (width) => setState(() => _docSearchPanelWidth = width),
                ),
              );
            },
          ),

          // Loading Overlay - ปิดไว้เพื่อไม่ให้จอกระพริบ
          // if (_isLoading) ... (removed)
        ],
      ),
    );
  }

  void setFocusNode(FocusNode focus) {
    focus.unfocus();
    Future.delayed(const Duration(milliseconds: 2000), () {
      focus.requestFocus();
    });
  }
}
