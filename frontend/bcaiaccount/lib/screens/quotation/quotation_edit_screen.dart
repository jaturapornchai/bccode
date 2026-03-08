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
import 'package:smlaicloud/screens/quotation/quotation_list_screen.dart';
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
import 'package:smlaicloud/modules/cart-system/models/cart_transfer_model.dart' show CartTransferModel;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_currency_utils.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_ui_utils.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_helper.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_approval_status_badge.dart';
// ==== QT Refactored Components ====
import 'package:smlaicloud/screens/quotation/utils/qt_form_controller.dart';
import 'package:smlaicloud/screens/quotation/utils/qt_workflow_manager.dart';
import 'package:smlaicloud/screens/quotation/utils/qt_bloc_effect_handler.dart';
import 'package:smlaicloud/screens/quotation/utils/qt_detail_command_handler.dart';
import 'package:smlaicloud/screens/quotation/utils/qt_product_search_handler.dart';
import 'package:smlaicloud/screens/quotation/utils/qt_currency_handler.dart';
import 'package:smlaicloud/screens/quotation/utils/qt_cart_integration_handler.dart';
import 'package:smlaicloud/screens/quotation/utils/qt_save_helper.dart';
import 'package:smlaicloud/screens/quotation/components/qt_responsive_layout.dart';
import 'package:smlaicloud/screens/quotation/components/qt_appbar_actions.dart';
import 'package:smlaicloud/screens/quotation/models/qt_model_factory.dart';
// Reuse PO approval handler + warehouse state
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/models/po_warehouse_state.dart';

enum QuotationEditScreenModule { header, detail, footer }

class QuotationEditScreen extends StatefulWidget {
  const QuotationEditScreen({super.key});

  @override
  State<QuotationEditScreen> createState() => QuotationEditScreenState();
}

class QuotationEditScreenState extends State<QuotationEditScreen> with TickerProviderStateMixin {
  // QT ใช้ type นี้เสมอ
  final global.TransactionTypeEnum transactionType = global.TransactionTypeEnum.quotation;
  late SplitViewController splitViewController;
  late TabController tabController;
  late TabController editTabController;
  late TransactionModel screenData;
  late TransactionModel screenDataTemp;
  late QuotationEditScreenModule moduleEdit;
  final ProductBarcodeRepository productBarcodeRepository = ProductBarcodeRepository();

  /// Helper instance สำหรับ dialog functions
  late final TransactionDialogs _transactionDialogs = TransactionDialogs(this);

  /// Helper instance สำหรับ calculator functions
  late final TransactionCalculatorBridge _transactionCalculator = TransactionCalculatorBridge(this);

  /// Public getter สำหรับให้ helper classes เข้าถึง calculator
  TransactionCalculator get transactionCalculator => _transactionCalculator;

  /// Helper instance สำหรับ search handler functions
  late final TransactionSearchHandlers _searchHandlers = TransactionSearchHandlers(this);

  /// Public getter/setter สำหรับ showDocSearchPanel
  bool get showDocSearchPanel => _showDocSearchPanel;
  set showDocSearchPanel(bool value) => _showDocSearchPanel = value;

  // ==== Form Controllers - ใช้ QTFormController รวมศูนย์ ====
  final QTFormController _formController = QTFormController();

  // ==== Workflow Manager - จัดการ Save/Print/Approval flow ====
  QTWorkflowManager get _workflowManager => QTWorkflowManager(
        context: context,
        isMounted: () => mounted,
        getDocNo: () => screenData.docno,
        getCollectionName: _getCollectionNameForType,
        getDocumentTitle: _getDocumentTitle,
        submitQTForApproval: _submitQTForApproval,
        onClearScreen: clearScreenData,
        onShowSaveDialog: (docNo) => showSaveDocNoDialog(context, docNo),
        onRegenerateDocNoAndSave: _regenerateDocNoAndSave,
        editTabController: editTabController,
        checkIsMultiCurrency: _checkIsMultiCurrency,
      );

  // ==== Bloc Effect Handler - จัดการ Bloc states (Trans, Warehouse, ExportCsv) ====
  QTBlocEffectHandler get _blocEffectHandler => QTBlocEffectHandler(
        isMounted: () => mounted,
        onStateChanged: () => setState(() {}),
        setLoading: (isLoading) => setState(() => _isLoading = isLoading),
        getScreenData: () => screenData,
        setScreenData: (trans) => screenData = trans,
        setScreenDataTemp: (trans) => screenDataTemp = TransactionModel.fromJson(trans.toJson()),
        onClearScreen: clearScreenData,
        onShowSaveDialog: (docNo) => showSaveDocNoDialog(context, docNo),
        onSubmitQTAndHandlePrint: _submitQTAndHandlePrint,
        onSubmitQTAndHandleUpdatePrint: _submitQTAndHandleUpdatePrint,
        onHandleSaveSuccessWithPrintCheck: _handleSaveSuccessWithPrintCheck,
        onHandleUpdateSuccessWithPrintCheck: (transType) => _workflowManager.handleUpdateSuccessWithPrintCheck(transType),
        onRefreshListAfterSave: _refreshListAfterSave,
        onRefreshListAfterDelete: _refreshListAfterDelete,
        onLoadPurchaseTypesAndExisting: _loadPurchaseTypesAndExisting,
        onLoadCurrencies: _loadCurrencies,
        onLoadQTApprovalStatus: _loadQTApprovalStatus,
        onShowDuplicateDocNoDialog: _showDuplicateDocNoDialog,
        onWarehouseListLoaded: (warehouses) => warehouseList = warehouses,
        onSetDefaultWarehouse: _setDefaultWarehouse,
        transactionType: transactionType,
        showDocSearchPanel: _showDocSearchPanel,
      );

  /// Helper สำหรับตั้งค่า warehouse defaults
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

  /// Refresh list หลัง delete
  void _refreshListAfterDelete(String deletedGuid) {
    if (_showDocSearchPanel && deletedGuid.isNotEmpty && _qtListKey.currentState != null) {
      _qtListKey.currentState!.refreshListAfterDelete(deletedGuid);
    }
  }

  /// Refresh list หลัง save/update
  void _refreshListAfterSave() {
    if (!_showDocSearchPanel) return;

    final guidToSelect = screenData.guidfixed;
    if (guidToSelect != null && _qtListKey.currentState != null) {
      _qtListKey.currentState!.setCardLoadingState(guidToSelect);
    }
    Future.delayed(const Duration(milliseconds: 2000), () {
      if (mounted && _qtListKey.currentState != null) {
        _qtListKey.currentState!.refreshListAndMaintainPosition(
          selectDocGuid: guidToSelect,
        );
      }
    });
  }

  // Getters สำหรับเข้าถึง controllers
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

  // ==== Warehouse State ====
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
  bool _showPreview = false;
  bool _hasCheckedScreenSize = false;
  bool _showDocSearchPanel = false;
  double _docSearchPanelWidth = 550.0;

  final GlobalKey<QuotationListScreenState> _qtListKey = GlobalKey<QuotationListScreenState>();

  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  // pricelevel ของลูกค้า — Quotation ดึงราคาขายปลีก (keynumber=1)
  int customerPriceLevel = 1;

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

  // ==== Approval Handler - จัดการ purchase types และ approval status (Reuse PO) ====
  late final POApprovalHandler _approvalHandler = POApprovalHandler(
    onStateChanged: () => setState(() {}),
    onShowSuccess: (msg) => context.ui.showSuccess(msg),
    onShowInfo: (msg) => context.ui.showInfo(msg),
    onShowError: (msg) => context.ui.showError(msg),
  );

  // Getters สำหรับ backward compatibility (purchase types)
  List<PurchaseTypeModel> get _purchaseTypes => _approvalHandler.purchaseTypes;
  String? get _selectedPurchaseTypeCode => _approvalHandler.selectedPurchaseTypeCode;

  // สกุลเงิน - ใช้ QTCurrencyHandler
  late final QTCurrencyHandler _currencyHandler = QTCurrencyHandler(
    onStateChanged: () => setState(() {}),
    onWarning: (msg) => context.ui.showWarning(msg),
    onRecalculate: () => _transactionCalculator.calTotalValue(),
  );

  // Getters สำหรับเข้าถึงข้อมูลจาก handler
  List<CurrencyModel> get _currencies => _currencyHandler.currencies;
  String? get _selectedDocCurrencyCode => _currencyHandler.docCurrencyCode;
  String? get _selectedDocCurrencySymbol => _currencyHandler.docCurrencySymbol;
  double? get _selectedExchangeRate => _currencyHandler.exchangeRate;
  String? get _baseCurrency => _currencyHandler.baseCurrency;
  String? get _baseCurrencySymbol => _currencyHandler.baseCurrencySymbol;

  // Getters สำหรับ approval status
  POApprovalStatusModel? get _qtApprovalStatus => _approvalHandler.approvalStatus;
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
    // QT ใช้ลูกค้า (ลูกหนี้)
    fieldNameCustCode = "doc_customer_code";
    fieldNameCustName = "doc_customer_name";

    setSystemLanguageList();
    splitViewController = SplitViewController(limits: [null, WeightLimit(min: 0.1, max: 0.9)]);
    splitViewController.weights = [0.30, 0.70];
    tabController = TabController(length: 2, vsync: this);

    // QT ใช้ 2 tabs
    editTabController = TabController(length: 2, vsync: this);

    context.read<WarehouseBloc>().add(const WarehouseLoadList(offset: 0, limit: 100, search: ''));

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_hasCheckedScreenSize) {
        final screenWidth = MediaQuery.of(context).size.width;
        if (screenWidth > 600) {
          setState(() {
            tabController.dispose();
            tabController = TabController(length: _showPreview ? 2 : 1, vsync: this);
          });
        }
        _hasCheckedScreenSize = true;
      }
    });
  }

  bool _hasProcessedCartTransfer = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();

    if (!_hasProcessedCartTransfer) {
      CartTransferModel? cartTransfer = QTCartIntegrationHandler.getCartTransfer(context);

      if (cartTransfer != null) {
        AppLogger.info('[QT] รับข้อมูลจาก Cart System: ${cartTransfer.cartName}');
        _populateFromCartTransfer(cartTransfer);
        _hasProcessedCartTransfer = true;
      }
    }
  }

  /// แปลงข้อมูลจาก CartTransferModel มาเป็น TransactionModel
  void _populateFromCartTransfer(CartTransferModel cartTransfer) {
    try {
      screenData = QTCartIntegrationHandler.populateFromCart(
        screenData: screenData,
        cartTransfer: cartTransfer,
        calcflag: calcflag,
      );

      setState(() {
        loadDataToScreen();
      });
    } catch (e, stackTrace) {
      AppLogger.error('[QT] Error populating from cart transfer: $e\n$stackTrace');
      if (mounted) {
        context.ui.showError('${global.language("error_loading_data_from_cart")}: $e');
      }
    }
  }

  @override
  void dispose() {
    _debouncer.cancel();
    docrefs.clear();
    splitViewController.dispose();
    tabController.dispose();
    editTabController.dispose();
    _formController.dispose();

    super.dispose();
  }

  void headerTableDetail() {
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

  DateTime _safeParseDatetime(String? dateString) {
    if (dateString == null || dateString.isEmpty) {
      return DateTime.now();
    }
    try {
      return DateTime.parse(dateString);
    } catch (e) {
      AppLogger.warning('[QT] แปลงวันที่ไม่สำเร็จ: "$dateString" → ใช้ DateTime.now() แทน');
      return DateTime.now();
    }
  }

  void loadDataToScreen() {
    headerTableDetail();
    _formController.loadDataFromModel(screenData);
    _transactionCalculator.calPayTotal();
  }

  /// Helper method สำหรับดึงชื่อ collection (สำหรับ PDF API)
  String _getCollectionNameForType(global.TransactionTypeEnum type) {
    return "transactionQuotation";
  }

  /// Helper method สำหรับดึงชื่อเอกสาร (สำหรับ PDF title)
  String _getDocumentTitle(global.TransactionTypeEnum type) {
    return global.language("quotation");
  }

  /// ส่ง QT เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ (Save)
  Future<void> _submitQTAndHandlePrint(String docNo, global.TransactionTypeEnum transType) async {
    await _workflowManager.submitAndHandlePrint(docNo, transType);
  }

  /// ส่ง QT เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ (Update)
  Future<void> _submitQTAndHandleUpdatePrint(global.TransactionTypeEnum transType) async {
    await _workflowManager.submitAndHandleUpdatePrint(transType);
  }

  /// จัดการหลังบันทึกสำเร็จ
  Future<void> _handleSaveSuccessWithPrintCheck(String docNo, global.TransactionTypeEnum transType) async {
    await _workflowManager.handleSaveSuccessWithPrintCheck(docNo, transType);
  }

  /// แสดง dialog เมื่อ docno ซ้ำ
  void _showDuplicateDocNoDialog() {
    _workflowManager.showDuplicateDocNoDialog(screenData.docno);
  }

  /// บันทึกอีกครั้งโดยให้ mainapi สร้าง docno ใหม่
  void _regenerateDocNoAndSave() {
    setState(() {
      screenData.docno = '';
      final timezoneInfo = global.getLocalTimezoneInfo();
      screenData.docdatelocal = timezoneInfo['docdatelocal'];
      screenData.doctimelocal = timezoneInfo['doctimelocal'];
      screenData.timezone = timezoneInfo['timezone'];
      AppLogger.info('[QT] ล้าง docno เพื่อให้ mainapi สร้างใหม่ (docdatelocal: ${screenData.docdatelocal})');
    });

    saveOrUpdateData(taxInvoice: false);
  }

  void clearScreenData() {
    // QT: ค่าคงที่สำหรับ Quotation
    transflag = 30;
    calcflag = -1;
    docnoFormat = "QT";
    configvattype = global.config.vattypesale;
    configinquirytype = 0;
    moduleEdit = QuotationEditScreenModule.header;

    // สร้าง TransactionModel เปล่าผ่าน Factory
    screenData = QTModelFactory.createForQuotation();
    screenDataTemp = QTModelFactory.createForQuotation();

    // Quotation ดึงราคาขายปลีก (keynumber=1)
    customerPriceLevel = 1;

    // รีเซ็ตประเภทและสถานะการอนุมัติ
    _approvalHandler.resetPurchaseType();
    _approvalHandler.resetApprovalStatus();

    setState(() {
      loadDataToScreen();
      docDateTimeValidated = true;
    });

    // โหลดประเภทใบเสนอราคาและสกุลเงิน
    _loadPurchaseTypesForQT();
    _loadCurrencies();
  }

  /// โหลดรายการประเภทใบเสนอราคา
  Future<void> _loadPurchaseTypesForQT() async {
    if (!mounted) return;
    await _approvalHandler.loadPurchaseTypesForNewPO();
  }

  /// โหลดประเภทใบเสนอราคาและค่าที่บันทึกไว้
  Future<void> _loadPurchaseTypesAndExisting(String docNo) async {
    if (!mounted) return;
    await _approvalHandler.loadPurchaseTypesAndExisting(screenData);
  }

  /// บันทึกประเภทใบเสนอราคาลงใน TransactionModel
  void _saveQTPurchaseType() {
    _approvalHandler.savePurchaseTypeToModel(screenData);
    _saveCurrencyData();
  }

  /// บันทึกข้อมูลสกุลเงินลงใน TransactionModel
  void _saveCurrencyData() {
    if (_selectedDocCurrencyCode == null || _selectedDocCurrencyCode!.isEmpty) return;

    TransactionCurrencyUtils.initializeWithBaseCurrency(
      screenData,
      _baseCurrency!,
      _baseCurrencySymbol!,
    );

    if (_selectedDocCurrencyCode != _baseCurrency) {
      screenData.docCurrency = _selectedDocCurrencyCode;
      screenData.docCurrencySymbol = _selectedDocCurrencySymbol;
      double rate = _selectedExchangeRate ?? 1.0;
      if (rate <= 0 || rate.isNaN || rate.isInfinite) {
        AppLogger.warning('[QT] Exchange rate ไม่ถูกต้อง: $_selectedExchangeRate → ใช้ 1.0 แทน');
        rate = 1.0;
      }
      screenData.exchangerate = rate;
    }

    AppLogger.info('[QT] บันทึกสกุลเงิน: doc=${screenData.docCurrency}, base=${screenData.currency}, อัตรา=${screenData.exchangerate}');
  }

  /// เปลี่ยนประเภทใบเสนอราคาที่เลือก
  void _onPurchaseTypeChanged(String? code, String? name) {
    _approvalHandler.onPurchaseTypeChanged(code, name);
  }

  /// โหลดรายการสกุลเงิน
  bool _checkIsMultiCurrency() {
    return _currencyHandler.isMultiCurrency ||
        (screenData.docCurrency != null &&
         screenData.docCurrency!.isNotEmpty &&
         screenData.docCurrency!.toUpperCase() != global.getBaseCurrency()) ||
        (screenData.exchangerate != null &&
         screenData.exchangerate != 0 &&
         screenData.exchangerate != 1.0);
  }

  Future<void> _loadCurrencies() async {
    await _currencyHandler.loadCurrencies(screenData);
  }

  void _onCurrencyChanged(String? code, String? symbol, double? exchangeRate) async {
    await _currencyHandler.onCurrencyChanged(code, symbol, exchangeRate, screenData);
  }

  void _onExchangeRateChanged(double? exchangeRate) {
    _currencyHandler.onExchangeRateChanged(exchangeRate, screenData);
  }

  /// โหลดสถานะการอนุมัติ
  Future<void> _loadQTApprovalStatus(String docNo) async {
    if (!mounted) return;
    await _approvalHandler.loadApprovalStatus(docNo);
  }

  /// ส่ง QT เข้าระบบอนุมัติ
  Future<bool> _submitQTForApproval({bool isModified = false}) async {
    screenData.description = descriptionController.text;
    return await _approvalHandler.submitForApproval(
      screenData: screenData,
      description: screenData.description ?? '',
      isModified: isModified,
    );
  }

  /// ถอนการส่งอนุมัติ QT
  Future<bool> _withdrawQTApproval() async {
    return await _approvalHandler.withdrawApproval(
      docNo: screenData.docno,
    );
  }

  // =====================================================
  // Helper methods สำหรับตรวจสอบสิทธิ์พิมพ์/แก้ไข QT
  // =====================================================

  /// ดึงข้อมูลคลังสินค้าเริ่มต้น
  QTWarehouseInfo _getWarehouseInfo() {
    return QTWarehouseInfo(
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

  /// ตรวจสอบว่า QT สามารถพิมพ์ได้หรือไม่
  bool _canPrintQT() {
    return POApprovalHelper.canPrint(_qtApprovalStatus);
  }

  /// ตรวจสอบว่า QT สามารถแก้ไขได้หรือไม่
  bool _canEditQT() {
    if (screenData.docno.isEmpty) return true;
    return POApprovalHelper.canEdit(
      docNo: screenData.docno,
      approvalStatus: _qtApprovalStatus,
      isref: screenData.isref ?? false,
      creatorCode: screenData.creatorcode,
      currentUserCode: global.profileData.username,
    );
  }

  /// ข้อความแสดงเหตุผลที่ไม่สามารถแก้ไขได้
  String _getQTEditDisabledReason() {
    if (screenData.docno.isEmpty) return '';
    return POApprovalHelper.getEditDisabledReason(
      docNo: screenData.docno,
      approvalStatus: _qtApprovalStatus,
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
    // Quotation: ดึงราคาขายปลีก (keynumber=1)
    if (result != null && result.isNotEmpty) {
      try {
        final retailPrice = result.firstWhere((p) => p.keynumber == customerPriceLevel);
        return retailPrice.price;
      } catch (_) {
        // ถ้าไม่มี keynumber ที่ต้องการ ใช้ราคาแรก
        return result.first.price;
      }
    }
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
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: docPreview(),
    );
  }

  void togglePreview() {
    setState(() {
      _showPreview = !_showPreview;
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
      cartList: [],
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
      // ประเภทใบเสนอราคา
      purchaseTypes: _purchaseTypes,
      selectedPurchaseTypeCode: _selectedPurchaseTypeCode,
      onPurchaseTypeChanged: _onPurchaseTypeChanged,
      // สกุลเงิน (Multi-Currency)
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
    // QT: รวม approval status badge กับ document header
    Widget documentHeader = Column(
      children: [
        Expanded(child: editDocumentWidget()),
        // แสดง badge สถานะอนุมัติด้านล่างของเอกสาร
        POApprovalStatusBadge(
          screenData: screenData,
          approvalStatus: _qtApprovalStatus,
          isLoading: _isLoadingApprovalStatus,
          onWithdrawApproval: _withdrawQTApproval,
        ),
      ],
    );
    // QT: ใช้ 2 tabs (doc_header, doc_details)
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
                  if (!_canEditQT()) {
                    context.ui.showWarning(_getQTEditDisabledReason());
                    return;
                  }
                  final result = await showEnhancedSaveConfirmDialog(
                    context,
                    screenData.guidfixed ?? '',
                    showPrintOption: _canPrintQT(),
                  );
                  if (result != null && result.confirmed) {
                    final prefs = await SharedPreferences.getInstance();
                    await prefs.setBool(_printAfterSaveKey, result.printAfterSave);
                    AppLogger.debug('[QT Save F10] printAfterSave saved to prefs: ${result.printAfterSave}');
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
    AppLogger.info('[QT saveOrUpdateData] START - taxInvoice: $taxInvoice, type: $transactionType');

    _transactionCalculator.calTotalValue();

    if (!_transactionCalculator.verifyPayment()) {
      AppLogger.warn('[QT saveOrUpdateData] verifyPayment() returned FALSE - save blocked');
      return;
    }

    // เตรียมข้อมูลก่อนบันทึก
    QTSaveHelper.prepareDataForSave(
      screenData,
      creatorCode: global.profileData.username ?? '',
      creatorName: global.profileData.name ?? '',
    );

    // บันทึกประเภทใบเสนอราคาและสกุลเงิน
    _saveQTPurchaseType();

    // ส่ง Bloc Event
    if (taxInvoice) {
      AppLogger.info('[QT saveOrUpdateData] Sending TransCreateFullInvoice event');
      context.read<TransBloc>().add(TransCreateFullInvoice(guid: screenData.guidfixed!, trans: screenData, type: transactionType));
    } else if (screenData.guidfixed?.isNotEmpty ?? false) {
      AppLogger.info('[QT saveOrUpdateData] Sending TransUpdate event');
      context.read<TransBloc>().add(TransUpdate(guid: screenData.guidfixed!, trans: screenData, type: transactionType));
    } else {
      AppLogger.info('[QT saveOrUpdateData] Sending TransSave event (new document)');
      context.read<TransBloc>().add(TransSave(trans: screenData, type: transactionType));
    }
  }

  /// จัดการคำสั่งจากตารางรายการสินค้า
  void showDialogCommand(String cmd, int index, TransactionDetailModel details) async {
    final handler = QTDetailCommandHandler(
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
    screenData.docreferences!.removeAt(index);
    screenData.details!.removeWhere((detail) => detail.docref == docno);
    docrefs.removeWhere((docref) => docref['docref'] == docno);

    _transactionCalculator.calTotalValue();
    if (mounted) {
      setState(() {});
    }

    context.ui.showSuccess('${global.language("delete_reference_and_products")} $docno');
  }

  Future<void> showSaveDocNoDialog(BuildContext context, String docno) async {
    FocusNode focusNode = FocusNode();

    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        focusNode.requestFocus();
        return RawKeyboardListener(
          focusNode: focusNode,
          onKey: (event) {
            if (event is RawKeyDownEvent && event.logicalKey == LogicalKeyboardKey.escape) {
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
                    collection: 'transactionQuotation',
                    docNo: screenData.docno,
                    title: global.transactionName(transactionType),
                    context: this.context,
                    isMultiCurrency: _checkIsMultiCurrency(),
                  );
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
    });
  }

  /// เมื่อเลือกเอกสารจาก side panel
  void _onDocumentSelectedFromPanel(TransactionModel document) {
    if (document.docno.isNotEmpty) {
      DateTime localDateTime = _safeParseDatetime(document.docdatetime);
      screenData = document;
      screenDataTemp = document.copyWith(
        details: document.details?.map((detail) => detail.copyWith()).toList(),
        paymentdetail: document.paymentdetail?.copyWith(),
        paymentStructs: document.paymentStructs?.map((struct) => struct.copyWith()).toList(),
      );
      screenDataTemp.isdelete = document.isdelete;

      if (screenData.details != null && screenData.details!.isNotEmpty) {
        screenData.details!.sort((a, b) => a.linenumber.compareTo(b.linenumber));
      }
      screenData.docdatetime = localDateTime.toUtc().toIso8601String();

      loadDataToScreen();
      _transactionCalculator.calTotalValue();

      if (document.docno.isNotEmpty) {
        _loadPurchaseTypesAndExisting(document.docno);
        _loadCurrencies();
        _loadQTApprovalStatus(document.docno);
      }

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
        actions: QTAppBarActions.build(
          context: context,
          screenData: screenData,
          screenDataTemp: screenDataTemp,
          transactionType: transactionType,
          isLoading: _isLoading,
          showPreview: _showPreview,
          canEditQT: _canEditQT(),
          canPrintQT: _canPrintQT(),
          editDisabledReason: _getQTEditDisabledReason(),
          getCollectionName: _getCollectionNameForType,
          getDocumentTitle: _getDocumentTitle,
          tabController: tabController,
          onClearScreen: clearScreenData,
          onTogglePreview: togglePreview,
          onSearchTrans: _searchHandlers.searchTrans,
          onDeleteDoc: deleteDoc,
          onSaveOrUpdate: saveOrUpdateData,
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
                child: QTResponsiveLayout(
                  splitViewController: splitViewController,
                  tabController: tabController,
                  showPreview: _showPreview,
                  showDocSearchPanel: _showDocSearchPanel,
                  docSearchPanelWidth: _docSearchPanelWidth,
                  qtListKey: _qtListKey,
                  buildEditWidget: editWidget,
                  buildPreviewWidget: previewWidget,
                  onDocumentSelected: _onDocumentSelectedFromPanel,
                  onClosePanel: () => setState(() => _showDocSearchPanel = false),
                  onPanelWidthChanged: (width) => setState(() => _docSearchPanelWidth = width),
                ),
              );
            },
          ),
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
