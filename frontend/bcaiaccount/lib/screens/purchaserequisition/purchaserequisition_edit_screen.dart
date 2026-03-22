// ignore_for_file: deprecated_member_use

import 'dart:convert';
import 'dart:io';
import 'dart:async';
import 'package:smlaicloud/bloc/export_csv/export_csv_bloc.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/services/pdf_service.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/screens/purchaserequisition/purchaserequisition_list_screen.dart';
import 'package:smlaicloud/screens/purchaserequisition/components/pr_header_widget.dart';
import 'package:smlaicloud/screens/transaction/components/document_preview_widget.dart';
import 'package:smlaicloud/screens/transaction/components/document_total_widget.dart';
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
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/widgets/attachment_count_icon_button.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_ui_utils.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_helper.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_product_search_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_approval_status_badge.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_appbar_actions.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_detail_command_handler.dart';
import 'package:smlaicloud/screens/purchaserequisition/utils/pr_form_controller.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_workflow_manager.dart';
import 'package:smlaicloud/screens/purchaserequisition/components/pr_responsive_layout.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_bloc_effect_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_save_helper.dart';
import 'package:smlaicloud/screens/purchaseorder/models/po_warehouse_state.dart';
import 'package:smlaicloud/screens/purchaserequisition/models/pr_model_factory.dart';
import 'package:smlaicloud/repositories/purchase_history_repository.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_currency_utils.dart';

enum PurchaseRequisitionEditScreenModule { header, detail, footer }

class PurchaseRequisitionEditScreen extends StatefulWidget {
  const PurchaseRequisitionEditScreen({super.key});

  @override
  State<PurchaseRequisitionEditScreen> createState() => PurchaseRequisitionEditScreenState();
}

class PurchaseRequisitionEditScreenState extends State<PurchaseRequisitionEditScreen> with TickerProviderStateMixin {
  // PR ใช้ type นี้เสมอ
  final global.TransactionTypeEnum transactionType = global.TransactionTypeEnum.purchaserequisition;
  late SplitViewController splitViewController;
  late TabController tabController;
  late TabController editTabController;
  late TransactionModel screenData;
  late TransactionModel screenDataTemp;
  late PurchaseRequisitionEditScreenModule moduleEdit;
  final ProductBarcodeRepository productBarcodeRepository = ProductBarcodeRepository();

  late final TransactionDialogs _transactionDialogs = TransactionDialogs(this);
  late final TransactionCalculatorBridge _transactionCalculator = TransactionCalculatorBridge(this);
  TransactionCalculator get transactionCalculator => _transactionCalculator;
  late final TransactionSearchHandlers _searchHandlers = TransactionSearchHandlers(this);


  bool get showDocSearchPanel => _showDocSearchPanel;
  set showDocSearchPanel(bool value) => _showDocSearchPanel = value;

  // ==== Form Controllers - ใช้ PRFormController รวมศูนย์ ====
  final PRFormController _formController = PRFormController();

  // ==== Workflow Manager ====
  POWorkflowManager get _workflowManager => POWorkflowManager(
        context: context,
        isMounted: () => mounted,
        getDocNo: () => screenData.docno,
        getCollectionName: _getCollectionNameForType,
        getDocumentTitle: _getDocumentTitle,
        submitPOForApproval: _submitPRForApproval,
        onClearScreen: clearScreenData,
        onShowSaveDialog: (docNo) => showSaveDocNoDialog(context, docNo),
        onRegenerateDocNoAndSave: _regenerateDocNoAndSave,
        editTabController: editTabController,
        checkIsMultiCurrency: () => false, // PR ไม่ใช้ multi-currency
      );

  // ==== Bloc Effect Handler ====
  POBlocEffectHandler get _blocEffectHandler => POBlocEffectHandler(
        isMounted: () => mounted,
        onStateChanged: () => setState(() {}),
        setLoading: (isLoading) => setState(() => _isLoading = isLoading),
        getScreenData: () => screenData,
        setScreenData: (trans) => screenData = trans,
        setScreenDataTemp: (trans) => screenDataTemp = TransactionModel.fromJson(trans.toJson()),
        onClearScreen: clearScreenData,
        onShowSaveDialog: (docNo) => showSaveDocNoDialog(context, docNo),
        onSubmitPOAndHandlePrint: _submitPRAndHandlePrint,
        onSubmitPOAndHandleUpdatePrint: _submitPRAndHandleUpdatePrint,
        onHandleSaveSuccessWithPrintCheck: _handleSaveSuccessWithPrintCheck,
        onHandleUpdateSuccessWithPrintCheck: (transType) => _workflowManager.handleUpdateSuccessWithPrintCheck(transType),
        onRefreshListAfterSave: _refreshListAfterSave,
        onRefreshListAfterDelete: _refreshListAfterDelete,
        onLoadPurchaseTypesAndExisting: _loadPurchaseTypesAndExisting,
        onLoadCurrencies: () {}, // PR ไม่ใช้ currency
        onLoadPOApprovalStatus: _loadPRApprovalStatus,
        onShowDuplicateDocNoDialog: _showDuplicateDocNoDialog,
        onWarehouseListLoaded: (warehouses) => warehouseList = warehouses,
        onSetDefaultWarehouse: _setDefaultWarehouse,
        transactionType: transactionType,
        showDocSearchPanel: _showDocSearchPanel,
      );

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

  void _refreshListAfterDelete(String deletedGuid) {
    if (_showDocSearchPanel && deletedGuid.isNotEmpty && _prListKey.currentState != null) {
      _prListKey.currentState!.refreshListAfterDelete(deletedGuid);
    }
  }

  void _refreshListAfterSave() {
    if (!_showDocSearchPanel) return;
    final guidToSelect = screenData.guidfixed;
    if (guidToSelect != null && _prListKey.currentState != null) {
      _prListKey.currentState!.setCardLoadingState(guidToSelect);
    }
    Future.delayed(const Duration(milliseconds: 2000), () {
      if (mounted && _prListKey.currentState != null) {
        _prListKey.currentState!.refreshListAndMaintainPosition(
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

  final POWarehouseState _warehouseState = POWarehouseState();

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

  final GlobalKey<PurchaseRequisitionListScreenState> _prListKey = GlobalKey<PurchaseRequisitionListScreenState>();

  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  int customerPriceLevel = 1;
  String? clientId;

  // === Multi-Currency ===
  List<CurrencyModel> _currencies = [];
  String? _selectedDocCurrencyCode;
  String? _selectedDocCurrencySymbol;
  double? _selectedExchangeRate;
  String? _baseCurrency;
  String? _baseCurrencySymbol;

  List<global.DataTableHeader> headers = [
    global.DataTableHeader(code: "delete", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
    global.DataTableHeader(code: "reorder", label: "", width: 4, textAlign: TextAlign.center, alignment: Alignment.center),
    global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 10, textAlign: TextAlign.center, alignment: Alignment.center),
    global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 25),
    global.DataTableHeader(code: "product_name", label: global.language('product_name'), width: 50),
    global.DataTableHeader(code: "product_unit", label: global.language('product_unit'), width: 15),
    global.DataTableHeader(code: "product_qty", label: global.language('product_qty'), width: 15, alignment: Alignment.centerRight, textAlign: TextAlign.right),
  ];

  List<dynamic> docrefs = [];

  // ==== PR Approval Handler ====
  late final POApprovalHandler _approvalHandler = POApprovalHandler(
    onStateChanged: () => setState(() {}),
    onShowSuccess: (msg) => context.ui.showSuccess(msg),
    onShowInfo: (msg) => context.ui.showInfo(msg),
    onShowError: (msg) => context.ui.showError(msg),
  );

  List<PurchaseTypeModel> get _purchaseTypes => _approvalHandler.purchaseTypes;
  String? get _selectedPurchaseTypeCode => _approvalHandler.selectedPurchaseTypeCode;
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
    // PR ใช้ supplier (ผู้จำหน่าย) เหมือน PO
    fieldNameCustCode = "doc_supplier_code";
    fieldNameCustName = "doc_supplier_name";

    // Initialize screenData ก่อน build แรก เพื่อป้องกัน LateInitializationError
    screenData = PRModelFactory.createForPR();
    screenDataTemp = PRModelFactory.createForPR();

    splitViewController = SplitViewController(limits: [null, WeightLimit(min: 0.1, max: 0.9)]);
    splitViewController.weights = [0.30, 0.70];
    tabController = TabController(length: 2, vsync: this);
    editTabController = TabController(length: 2, vsync: this);

    context.read<WarehouseBloc>().add(const WarehouseLoadList(offset: 0, limit: 100, search: ''));

    WidgetsBinding.instance.addPostFrameCallback((_) {
      setSystemLanguageList();
      _loadCurrencies();
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
      global.DataTableHeader(code: "reorder", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
      global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 8, textAlign: TextAlign.center, alignment: Alignment.center),
      global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 18),
      global.DataTableHeader(code: "item_code", label: global.language('item_code'), width: 14),
      global.DataTableHeader(code: "product_name", label: global.language('product_name'), width: 30),
      global.DataTableHeader(code: "product_unit", label: global.language('product_unit'), width: 12),
      global.DataTableHeader(code: "product_qty", label: global.language('product_qty'), width: 12, alignment: Alignment.centerRight, textAlign: TextAlign.right),
      global.DataTableHeader(code: "price", label: global.language('price'), width: 14, alignment: Alignment.centerRight, textAlign: TextAlign.right),
      global.DataTableHeader(code: "discount", label: global.language('discount'), width: 12, alignment: Alignment.centerRight, textAlign: TextAlign.right),
      global.DataTableHeader(code: "sum_amount", label: global.language('sum_amount'), width: 14, alignment: Alignment.centerRight, textAlign: TextAlign.right),
      global.DataTableHeader(code: "serial_no", label: global.language('serial_no'), width: 14),
      global.DataTableHeader(code: "lot_no", label: global.language('lot_no'), width: 14),
      global.DataTableHeader(code: "expiry_date", label: global.language('expiry_date'), width: 14),
    ];
  }

  DateTime _safeParseDatetime(String? dateString) {
    if (dateString == null || dateString.isEmpty) return DateTime.now();
    try {
      return DateTime.parse(dateString);
    } catch (e) {
      AppLogger.warning('[PR] แปลงวันที่ไม่สำเร็จ: "$dateString" → ใช้ DateTime.now() แทน');
      return DateTime.now();
    }
  }

  void loadDataToScreen() {
    headerTableDetail();
    _formController.loadDataFromModel(screenData);
    _transactionCalculator.calPayTotal();
  }

  String _getCollectionNameForType(global.TransactionTypeEnum type) {
    return "transactionPurchaseRequisition";
  }

  String _getDocumentTitle(global.TransactionTypeEnum type) {
    return global.language("purchase_requisition");
  }

  Future<void> _submitPRAndHandlePrint(String docNo, global.TransactionTypeEnum transType) async {
    await _workflowManager.submitAndHandlePrint(docNo, transType);
  }

  Future<void> _submitPRAndHandleUpdatePrint(global.TransactionTypeEnum transType) async {
    await _workflowManager.submitAndHandleUpdatePrint(transType);
  }

  Future<void> _handleSaveSuccessWithPrintCheck(String docNo, global.TransactionTypeEnum transType) async {
    await _workflowManager.handleSaveSuccessWithPrintCheck(docNo, transType);
  }

  void _showDuplicateDocNoDialog() {
    _workflowManager.showDuplicateDocNoDialog(screenData.docno);
  }

  void _regenerateDocNoAndSave() {
    setState(() {
      screenData.docno = '';
      final timezoneInfo = global.getLocalTimezoneInfo();
      screenData.docdatelocal = timezoneInfo['docdatelocal'];
      screenData.doctimelocal = timezoneInfo['doctimelocal'];
      screenData.timezone = timezoneInfo['timezone'];
      AppLogger.info('[PR] ล้าง docno เพื่อให้ mainapi สร้างใหม่');
    });
    saveOrUpdateData(taxInvoice: false);
  }

  void clearScreenData() {
    // PR: ค่าคงที่สำหรับ Purchase Requisition
    transflag = 21;
    calcflag = 1;
    docnoFormat = "PR";
    configvattype = global.config.vattypepurchase;
    configinquirytype = global.config.inquirytypepurchase;
    moduleEdit = PurchaseRequisitionEditScreenModule.header;

    screenData = PRModelFactory.createForPR();
    screenDataTemp = PRModelFactory.createForPR();

    customerPriceLevel = 1;

    _approvalHandler.resetPurchaseType();
    _approvalHandler.resetApprovalStatus();

    setState(() {
      loadDataToScreen();
      docDateTimeValidated = true;
    });

    _loadPurchaseTypesForPR();
  }

  // ===== Multi-Currency =====

  Future<void> _loadCurrencies() async {
    try {
      final apiService = CurrencyApiService();
      final shopId = global.prefs.getString("shopid") ?? "";
      final response = await apiService.getCurrencies(shopId: shopId, page: 1, limit: 1000);

      if (response.success && response.data != null) {
        final currencies = (response.data as List)
            .map((e) => CurrencyModel.fromJson(e))
            .where((c) => !c.isdisabled)
            .toList();

        final baseCurrencyFromApi = TransactionCurrencyUtils.findBaseCurrency(currencies);
        _baseCurrency = baseCurrencyFromApi.code;
        _baseCurrencySymbol = baseCurrencyFromApi.symbol;

        if (currencies.length <= 1) {
          _handleSingleCurrency(currencies);
          return;
        }

        setState(() {
          _currencies = currencies;
          _initializeCurrencySelection();
        });
      }
    } catch (e) {
      AppLogger.error('[PR] โหลดรายการสกุลเงินผิดพลาด: $e');
      _baseCurrency = 'THB';
      _baseCurrencySymbol = '฿';
    }
  }

  void _handleSingleCurrency(List<CurrencyModel> currencies) {
    if (screenData.isMultiCurrency) {
      _selectedDocCurrencyCode = screenData.docCurrency;
      _selectedDocCurrencySymbol = screenData.docCurrencySymbol;
      _selectedExchangeRate = screenData.safeExchangeRate;
      setState(() {
        _currencies = TransactionCurrencyUtils.createCurrencyListForExistingDoc(
          _baseCurrency!, _baseCurrencySymbol!, screenData.docCurrency!, screenData.docCurrencySymbol,
        );
      });
      return;
    }
    _initializeWithBaseCurrency();
  }

  void _initializeCurrencySelection() {
    if (screenData.docCurrency != null && screenData.docCurrency!.isNotEmpty) {
      _selectedDocCurrencyCode = screenData.docCurrency;
      _selectedDocCurrencySymbol = screenData.docCurrencySymbol;
      _selectedExchangeRate = screenData.safeExchangeRate;
    } else {
      _initializeWithBaseCurrency();
    }
  }

  void _initializeWithBaseCurrency() {
    _selectedDocCurrencyCode = _baseCurrency;
    _selectedDocCurrencySymbol = _baseCurrencySymbol;
    _selectedExchangeRate = 1.0;
    TransactionCurrencyUtils.initializeWithBaseCurrency(screenData, _baseCurrency!, _baseCurrencySymbol!);
  }

  void _saveCurrencyData() {
    if (_selectedDocCurrencyCode == null || _selectedDocCurrencyCode!.isEmpty) return;
    TransactionCurrencyUtils.initializeWithBaseCurrency(screenData, _baseCurrency!, _baseCurrencySymbol!);
    if (_selectedDocCurrencyCode != _baseCurrency) {
      screenData.docCurrency = _selectedDocCurrencyCode;
      screenData.docCurrencySymbol = _selectedDocCurrencySymbol;
      screenData.exchangerate = _selectedExchangeRate ?? 1.0;
    }
  }

  void _onCurrencyChanged(String? code, String? symbol, double? exchangeRate) async {
    if (code == _baseCurrency) {
      setState(() {
        _selectedDocCurrencyCode = code;
        _selectedDocCurrencySymbol = symbol;
        _selectedExchangeRate = 1.0;
        screenData.currency = _baseCurrency;
        screenData.currencysymbol = _baseCurrencySymbol;
        screenData.docCurrency = code;
        screenData.docCurrencySymbol = symbol;
        screenData.exchangerate = 1.0;
      });
      _transactionCalculator.calTotalValue();
      return;
    }

    double rate = exchangeRate ?? 1.0;
    if (code != null) {
      try {
        final today = DateTime.now();
        final dateStr = '${today.year}-${today.month.toString().padLeft(2, '0')}-${today.day.toString().padLeft(2, '0')}';
        final latestRate = await CurrencyApiService().getLatestExchangeRate(currency: code, date: dateStr);
        if (latestRate != null && latestRate.rate > 0) {
          rate = latestRate.rate;
        }
      } catch (_) {}
    }

    final validation = TransactionCurrencyUtils.validateExchangeRate(rate, code ?? '', _baseCurrency!);
    if (!validation.isValid) rate = validation.correctedRate;
    if (validation.hasWarning && validation.message != null) {
      context.ui.showWarning(validation.message!);
    }

    setState(() {
      _selectedDocCurrencyCode = code;
      _selectedDocCurrencySymbol = symbol;
      _selectedExchangeRate = rate;
      screenData.currency = _baseCurrency;
      screenData.currencysymbol = _baseCurrencySymbol;
      screenData.docCurrency = code;
      screenData.docCurrencySymbol = symbol;
      screenData.exchangerate = rate;
    });
    _transactionCalculator.calTotalValue();
  }

  void _onExchangeRateChanged(double? exchangeRate) {
    double safeRate = exchangeRate ?? 1.0;
    if (_selectedDocCurrencyCode != null && _selectedDocCurrencyCode != _baseCurrency) {
      final validation = TransactionCurrencyUtils.validateExchangeRate(safeRate, _selectedDocCurrencyCode!, _baseCurrency!);
      safeRate = validation.correctedRate;
      if (validation.hasWarning && validation.message != null) {
        context.ui.showWarning(validation.message!);
      }
    } else {
      if (safeRate != 1.0) safeRate = 1.0;
    }
    setState(() {
      _selectedExchangeRate = safeRate;
      screenData.exchangerate = _selectedExchangeRate;
    });
    _transactionCalculator.calTotalValue();
  }

  Future<void> _loadPurchaseTypesForPR() async {
    if (!mounted) return;
    await _approvalHandler.loadPurchaseTypesForNewPO();
  }

  Future<void> _loadPurchaseTypesAndExisting(String docNo) async {
    if (!mounted) return;
    await _approvalHandler.loadPurchaseTypesAndExisting(screenData);
  }

  void _savePRPurchaseType() {
    _approvalHandler.savePurchaseTypeToModel(screenData);
  }

  void _onPurchaseTypeChanged(String? code, String? name) {
    _approvalHandler.onPurchaseTypeChanged(code, name);
  }

  Future<void> _loadPRApprovalStatus(String docNo) async {
    if (!mounted) return;
    await _approvalHandler.loadApprovalStatus(docNo);
  }

  Future<bool> _submitPRForApproval({bool isModified = false}) async {
    screenData.description = descriptionController.text;
    return await _approvalHandler.submitForApproval(
      screenData: screenData,
      description: screenData.description ?? '',
      isModified: isModified,
    );
  }

  Future<bool> _withdrawPRApproval() async {
    return await _approvalHandler.withdrawApproval(
      docNo: screenData.docno,
    );
  }

  Future<bool> _approvePR(String comment) async {
    return await _approvalHandler.approveApproval(
      docNo: screenData.docno,
      comment: comment,
    );
  }

  Future<bool> _rejectPR(String comment) async {
    return await _approvalHandler.rejectApproval(
      docNo: screenData.docno,
      comment: comment,
    );
  }

  // ===== Approval Helper Methods =====

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

  bool _canPrintPR() {
    return POApprovalHelper.canPrint(_poApprovalStatus);
  }

  bool _canEditPR() {
    if (screenData.docno.isEmpty) return true;
    return POApprovalHelper.canEdit(
      docNo: screenData.docno,
      approvalStatus: _poApprovalStatus,
      isref: screenData.isref ?? false,
      creatorCode: screenData.creatorcode,
      currentUserCode: global.profileData.username,
    );
  }

  String _getPREditDisabledReason() {
    if (screenData.docno.isEmpty) return '';
    return POApprovalHelper.getEditDisabledReason(
      docNo: screenData.docno,
      approvalStatus: _poApprovalStatus,
      isref: screenData.isref ?? false,
      creatorCode: screenData.creatorcode,
      currentUserCode: global.profileData.username,
    );
  }

  // ===== Save / Update / Delete =====

  /// สร้าง extraFields map จาก PR form controllers
  /// สำหรับ fields ที่ไม่มีใน TransactionModel (departmentcode, jobcode, shiptoaddress ฯลฯ)
  Map<String, dynamic> _buildPRExtraFields() {
    final extra = <String, dynamic>{};

    final dc = _formController.departmentCodeController.text;
    if (dc.isNotEmpty) {
      extra['departmentcode'] = dc;
      final dn = _formController.departmentNameController.text;
      if (dn.isNotEmpty) extra['departmentnames'] = [{'code': 'th', 'name': dn}];
    }

    final jc = _formController.jobCodeController.text;
    if (jc.isNotEmpty) {
      extra['jobcode'] = jc;
      final jn = _formController.jobNameController.text;
      if (jn.isNotEmpty) extra['jobnames'] = [{'code': 'th', 'name': jn}];
    }

    final sa = _formController.shipToAddressController.text;
    if (sa.isNotEmpty) extra['shiptoaddress'] = sa;
    final lat = _formController.shipToLatController.text;
    if (lat.isNotEmpty) extra['shiptolat'] = lat;
    final lng = _formController.shipToLngController.text;
    if (lng.isNotEmpty) extra['shiptolng'] = lng;

    final cc = _formController.costCenterCodeController.text;
    if (cc.isNotEmpty) {
      extra['costcentercode'] = cc;
      final cn = _formController.costCenterNameController.text;
      if (cn.isNotEmpty) extra['costcenternames'] = [{'code': 'th', 'name': cn}];
    }

    // Estimated Landed Cost
    final euc = _formController.estimatedUnitCostController.text;
    if (euc.isNotEmpty) extra['estimatedunitcost'] = euc;
    final ef = _formController.estimatedFreightController.text;
    if (ef.isNotEmpty) extra['estimatedfreight'] = ef;
    final ed = _formController.estimatedDutyController.text;
    if (ed.isNotEmpty) extra['estimatedduty'] = ed;
    final elc = _formController.estimatedLandedCostController.text;
    if (elc.isNotEmpty) extra['estimatedlandedcost'] = elc;

    // Vendor Preferences (JSON array — auto-synced จาก PRHeaderWidget)
    final vpJson = screenData.prPreferredVendor ?? '';
    if (vpJson.isNotEmpty) extra['preferredvendor'] = vpJson;

    // Approval Enhancement
    final ad = _formController.approvalDeadlineController.text;
    if (ad.isNotEmpty) extra['approvaldeadline'] = ad;

    // PR เพิ่มเติม (session 6)
    final internalNote = _formController.internalNoteController.text;
    if (internalNote.isNotEmpty) extra['internalnote'] = internalNote;

    final purchasingGroup = _formController.purchasingGroupController.text;
    if (purchasingGroup.isNotEmpty) extra['purchasinggroup'] = purchasingGroup;

    final overTol = double.tryParse(_formController.overDeliveryToleranceController.text) ?? 0;
    if (overTol > 0) extra['overdeliverytolerance'] = overTol;

    final underTol = double.tryParse(_formController.underDeliveryToleranceController.text) ?? 0;
    if (underTol > 0) extra['underdeliverytolerance'] = underTol;

    final budgetCode = _formController.budgetCodeController.text;
    if (budgetCode.isNotEmpty) extra['budgetcode'] = budgetCode;

    final budgetAmount = double.tryParse(_formController.budgetAmountController.text) ?? 0;
    if (budgetAmount > 0) extra['budgetamount'] = budgetAmount;

    return extra;
  }

  void saveOrUpdateData({bool taxInvoice = false}) {
    _transactionCalculator.calTotalValue();

    if (!_transactionCalculator.verifyPayment()) return;

    POSaveHelper.prepareDataForSave(
      screenData,
      creatorCode: global.profileData.username ?? '',
      creatorName: global.profileData.name ?? '',
    );

    // Set fields ที่มีใน TransactionModel ตรงๆ
    screenData.creditdays = int.tryParse(_formController.creditDaysController.text) ?? 0;
    screenData.docrefno = _formController.docRefNumberController.text;
    // Sync purpose → description (TransactionModel มี description ที่ repository จะ map เป็น purpose)
    screenData.description = _formController.purposeController.text;

    _savePRPurchaseType();
    _saveCurrencyData();

    final extraFields = _buildPRExtraFields();

    if (screenData.guidfixed?.isNotEmpty ?? false) {
      context.read<TransBloc>().add(TransUpdate(
        guid: screenData.guidfixed!,
        trans: screenData,
        type: transactionType,
        extraFields: extraFields,
      ));
    } else {
      context.read<TransBloc>().add(TransSave(
        trans: screenData,
        type: transactionType,
        extraFields: extraFields,
      ));
    }
  }

  void deleteDoc() {
    if (screenData.guidfixed == null || screenData.guidfixed!.isEmpty) return;
    context.read<TransBloc>().add(TransDelete(guid: screenData.guidfixed!, type: transactionType));
  }

  void showSaveDocNoDialog(BuildContext context, String docno) async {
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
                    collection: 'transactionPurchaseRequisition',
                    docNo: screenData.docno,
                    title: global.transactionName(transactionType),
                    context: this.context,
                    isMultiCurrency: (_selectedDocCurrencyCode ?? '').isNotEmpty && _selectedDocCurrencyCode != _baseCurrency && _selectedExchangeRate != 1.0,
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

  bool get showPreview => _showPreview;
  set showPreview(bool value) {
    setState(() {
      _showPreview = value;
    });
  }

  void togglePreview() {
    setState(() {
      _showPreview = !_showPreview;
      tabController.dispose();
      tabController = TabController(length: _showPreview ? 2 : 1, vsync: this);
    });
  }

  double getPrice(List<PriceDataModel>? result) {
    // PR: ราคาจะถูกกรอกโดยผู้ใช้ ไม่ดึงจากฐานข้อมูล
    return 0;
  }

  /// Auto-fill ราคาล่าสุดจาก purchase history เมื่อเพิ่มสินค้า
  Future<void> _autoFillLastPrice(TransactionDetailModel detail) async {
    final barcode = detail.barcode;
    if (barcode.isEmpty) return;

    try {
      final result = await PurchaseHistoryRepository().getHistory(barcode, months: 12);
      final stats = result.stats;
      if (stats != null && stats.lastPrice > 0 && detail.price == 0) {
        setState(() {
          detail.price = stats.lastPrice;
          detail.sumamount = detail.qty * detail.price;
        });
        _transactionCalculator.calTotalValue();
        AppLogger.info('[PR] Auto-fill ราคา ${global.formatNumber(stats.lastPrice)} จาก ${stats.lastVendor} (${stats.lastDate})');
      }
    } catch (_) {
      // ไม่มี purchase history — ไม่ต้องทำอะไร
    }
  }

  void showDialogCommand(String cmd, int index, TransactionDetailModel details) async {
    // PR-specific commands
    if (cmd == 'purchase_history') {
      await _showPurchaseHistoryDialog(details);
      return;
    }
    if (cmd == 'serial_no' || cmd == 'lot_no' || cmd == 'expiry_date') {
      await _showSerialLotEditDialog(index, cmd);
      return;
    }
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

  /// แสดงประวัติการซื้อสินค้า
  Future<void> _showPurchaseHistoryDialog(TransactionDetailModel details) async {
    final barcode = details.barcode;
    if (barcode.isEmpty) return;

    // แสดง loading dialog
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (_) => const Center(child: CircularProgressIndicator()),
    );

    final result = await PurchaseHistoryRepository().getHistory(barcode, months: 6);
    if (!mounted) return;
    Navigator.pop(context); // ปิด loading

    final items = result.items;
    final stats = result.stats;
    final productName = global.activeLangName(details.itemnames ?? []);

    await showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.history, size: 18, color: global.theme.primaryColor),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                '${global.language("purchase_history")}: $productName',
                style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold),
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
        contentPadding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
        content: SizedBox(
          width: 420,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Stats bar
              if (stats != null) ...[
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: global.theme.primaryColor.withValues(alpha: 0.08),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceAround,
                    children: [
                      _historyStatCell(global.language('avg_price'), global.formatUnitPrice(stats.avgPrice)),
                      _historyStatCell(global.language('min_price'), global.formatUnitPrice(stats.minPrice)),
                      _historyStatCell(global.language('max_price'), global.formatUnitPrice(stats.maxPrice)),
                      _historyStatCell(global.language('last_price'), global.formatUnitPrice(stats.lastPrice)),
                    ],
                  ),
                ),
                const SizedBox(height: 10),
              ],
              // List
              ConstrainedBox(
                constraints: const BoxConstraints(maxHeight: 300),
                child: items.isEmpty
                    ? Center(
                        child: Padding(
                          padding: const EdgeInsets.all(24),
                          child: Text(global.language('no_data'), style: TextStyle(color: global.theme.textSecondaryColor)),
                        ),
                      )
                    : ListView.separated(
                        shrinkWrap: true,
                        itemCount: items.length,
                        separatorBuilder: (context, index) => Divider(height: 1, color: global.theme.dividerBorderColor),
                        itemBuilder: (_, i) {
                          final item = items[i];
                          return Padding(
                            padding: const EdgeInsets.symmetric(vertical: 6),
                            child: Row(
                              children: [
                                Expanded(
                                  flex: 3,
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(item.docNo, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: global.theme.textColor)),
                                      Text(item.docDate, style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor)),
                                    ],
                                  ),
                                ),
                                Expanded(
                                  flex: 2,
                                  child: Text(item.custCode, style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor)),
                                ),
                                Expanded(
                                  flex: 2,
                                  child: Text(
                                    '${global.formatQuantity(item.totalQty)} ${item.unitCode}',
                                    textAlign: TextAlign.right,
                                    style: TextStyle(fontSize: 11, color: global.theme.textColor),
                                  ),
                                ),
                                Expanded(
                                  flex: 2,
                                  child: Text(
                                    global.formatUnitPrice(item.price),
                                    textAlign: TextAlign.right,
                                    style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: global.theme.positiveHighlightTextColor),
                                  ),
                                ),
                              ],
                            ),
                          );
                        },
                      ),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text(global.language('close')),
          ),
        ],
      ),
    );
  }

  /// แก้ไข Serial No. หรือ Lot No. ของ detail row
  Future<void> _showSerialLotEditDialog(int index, String fieldCode) async {
    final detail = screenData.details![index];
    final label = global.language(fieldCode);
    final productName = global.activeLangName(detail.itemnames ?? []);

    // Parse existing extrajson
    Map<String, dynamic> extra = {};
    if (detail.extrajson != null && detail.extrajson!.isNotEmpty) {
      try {
        extra = jsonDecode(detail.extrajson!) as Map<String, dynamic>;
      } catch (_) {}
    }

    final controller = TextEditingController(text: (extra[fieldCode] ?? '').toString());

    final result = await showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('$label: $productName', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold), overflow: TextOverflow.ellipsis),
        content: SizedBox(
          width: 300,
          child: TextField(
            controller: controller,
            autofocus: true,
            decoration: InputDecoration(
              labelText: label,
              border: const OutlineInputBorder(),
              isDense: true,
            ),
            onSubmitted: (v) => Navigator.pop(ctx, v),
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text(global.language('cancel'))),
          FilledButton(onPressed: () => Navigator.pop(ctx, controller.text), child: Text(global.language('save'))),
        ],
      ),
    );

    if (result == null || !mounted) return;

    // Save back to extrajson
    if (result.isEmpty) {
      extra.remove(fieldCode);
    } else {
      extra[fieldCode] = result;
    }
    detail.extrajson = extra.isEmpty ? null : jsonEncode(extra);
    setState(() {});
  }

  Widget _historyStatCell(String label, String value) {
    return Column(
      children: [
        Text(value, style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: global.theme.primaryColor)),
        Text(label, style: TextStyle(fontSize: 10, color: global.theme.textSecondaryColor)),
      ],
    );
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
    if (mounted) setState(() {});
    context.ui.showSuccess('${global.language("delete_reference_and_products")} $docno');
  }

  // ===== Copy Document =====

  /// คัดลอก PR เดิมเป็นเอกสารใหม่
  void _copyDocument() async {
    if (screenData.docno.isEmpty) return;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.copy, size: 18, color: global.theme.primaryColor),
            const SizedBox(width: 8),
            Text(global.language('copy_document'), style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold)),
          ],
        ),
        content: Text(global.language('copy_document_confirm')),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(global.language('cancel'))),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: Text(global.language('copy'))),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    // สร้าง PR ใหม่จากข้อมูลเดิม
    final oldDetails = screenData.details ?? [];
    final oldDescription = _formController.purposeController.text;
    final oldCustCode = _formController.custCodeController.text;
    final oldCustNames = _formController.custnamesController.text;
    final oldCreditDays = _formController.creditDaysController.text;
    final oldDeptCode = _formController.departmentCodeController.text;
    final oldDeptName = _formController.departmentNameController.text;
    final oldJobCode = _formController.jobCodeController.text;
    final oldJobName = _formController.jobNameController.text;
    final oldCostCenterCode = _formController.costCenterCodeController.text;
    final oldCostCenterName = _formController.costCenterNameController.text;
    final oldShipTo = _formController.shipToAddressController.text;
    final oldDocRefNo = _formController.docRefNumberController.text;
    final oldVatType = screenData.vattype;
    final oldVatRate = screenData.vatrate;

    // Clear screen → สร้าง PR ใหม่
    clearScreenData();

    // คัดลอกข้อมูลกลับ
    _formController.purposeController.text = oldDescription;
    _formController.custCodeController.text = oldCustCode;
    _formController.custnamesController.text = oldCustNames;
    _formController.creditDaysController.text = oldCreditDays;
    _formController.departmentCodeController.text = oldDeptCode;
    _formController.departmentNameController.text = oldDeptName;
    _formController.jobCodeController.text = oldJobCode;
    _formController.jobNameController.text = oldJobName;
    _formController.costCenterCodeController.text = oldCostCenterCode;
    _formController.costCenterNameController.text = oldCostCenterName;
    _formController.shipToAddressController.text = oldShipTo;
    _formController.docRefNumberController.text = oldDocRefNo;
    screenData.vattype = oldVatType;
    screenData.vatrate = oldVatRate;

    // คัดลอก details (สร้าง copy ใหม่ ไม่อ้างอิง object เดิม)
    for (final d in oldDetails) {
      screenData.details!.add(TransactionDetailModel(
        docdatetime: DateTime.now().toUtc().toIso8601String(),
        itemguid: d.itemguid,
        barcode: d.barcode,
        itemcode: d.itemcode,
        itemnames: d.itemnames,
        unitcode: d.unitcode,
        qty: d.qty,
        price: d.price,
        discount: d.discount,
        sumofcost: d.sumofcost,
        remark: d.remark,
        linenumber: d.linenumber,
        whcode: d.whcode,
        whnames: d.whnames,
        locationcode: d.locationcode,
        locationnames: d.locationnames,
        towhcode: d.towhcode,
        towhnames: d.towhnames,
        tolocationcode: d.tolocationcode,
        tolocationnames: d.tolocationnames,
        shelfcode: d.shelfcode,
        totalqty: d.totalqty,
        discountamount: d.discountamount,
        averagecost: d.averagecost,
        totalvaluevat: d.totalvaluevat,
        sumamountexcludevat: d.sumamountexcludevat,
        priceexcludevat: d.priceexcludevat,
        standvalue: d.standvalue,
        dividevalue: d.dividevalue,
        calcflag: d.calcflag,
        vattype: d.vattype,
        ispos: d.ispos,
        laststatus: d.laststatus,
        itemtype: d.itemtype,
        taxtype: d.taxtype,
        inquirytype: d.inquirytype,
        sumamount: d.sumamount,
        unitnames: d.unitnames,
      ));
    }

    _transactionCalculator.calTotalValue();
    setState(() {});

    if (mounted) {
      context.ui.showSuccess(global.language('copy_document_success'));
    }
  }

  // ===== Conversion Buttons (PR → RFQ / PO) =====

  /// ตรวจสอบว่า PR อนุมัติแล้วหรือไม่
  bool _isPRApproved() {
    return _poApprovalStatus?.isCompleted == true &&
        (screenData.guidfixed ?? '').isNotEmpty;
  }

  /// สร้าง RFQ จาก PR ที่อนุมัติแล้ว
  void _convertToRFQ() {
    if (!_isPRApproved()) return;
    Navigator.pushNamed(
      context,
      '/transaction/rfq',
      arguments: _buildConversionArgs('pr'),
    );
  }

  /// สร้าง PO จาก PR ที่อนุมัติแล้ว
  void _convertToPO() {
    if (!_isPRApproved()) return;
    Navigator.pushNamed(
      context,
      '/transaction/purchaseorder',
      arguments: _buildConversionArgs('pr'),
    );
  }

  /// สร้าง arguments สำหรับส่งไปหน้าเอกสารปลายทาง
  Map<String, dynamic> _buildConversionArgs(String sourceType) {
    return {
      'source': sourceType,
      'sourceDocNo': screenData.docno,
      'sourceGuid': screenData.guidfixed,
      'custcode': screenData.custcode,
      'custnames': screenData.custnames,
      'details': screenData.details?.map((d) => d.toJson()).toList(),
      'description': screenData.description,
      'creditdays': screenData.creditdays,
    };
  }

  // ===== Widget Builders (เหมือน PO) =====

  /// Sync PR form controller values → screenData transient fields ก่อนแสดง preview
  void _syncPRFieldsToScreenData() {
    screenData.prDepartmentCode = _formController.departmentCodeController.text;
    screenData.prDepartmentName = _formController.departmentNameController.text;
    screenData.prJobCode = _formController.jobCodeController.text;
    screenData.prJobName = _formController.jobNameController.text;
    screenData.prCostCenterCode = _formController.costCenterCodeController.text;
    screenData.prCostCenterName = _formController.costCenterNameController.text;
    screenData.prShipToAddress = _formController.shipToAddressController.text;
    screenData.prEstimatedUnitCost = _formController.estimatedUnitCostController.text;
    screenData.prEstimatedFreight = _formController.estimatedFreightController.text;
    screenData.prEstimatedDuty = _formController.estimatedDutyController.text;
    // prPreferredVendor auto-synced จาก PRHeaderWidget (JSON array)
    screenData.prApprovalDeadline = _formController.approvalDeadlineController.text;
    screenData.prInternalNote = _formController.internalNoteController.text;
    screenData.prPurchasingGroup = _formController.purchasingGroupController.text;
    screenData.prOverDeliveryTolerance = double.tryParse(_formController.overDeliveryToleranceController.text);
    screenData.prUnderDeliveryTolerance = double.tryParse(_formController.underDeliveryToleranceController.text);
    screenData.prBudgetCode = _formController.budgetCodeController.text;
    screenData.prBudgetAmount = double.tryParse(_formController.budgetAmountController.text);
    screenData.description = _formController.purposeController.text;
    screenData.creatorname = _formController.requesterNameController.text;
    screenData.prUrgency = screenData.inquirytype;
  }

  Widget docPreview() {
    // sync ใน addPostFrameCallback เพื่อไม่ให้ setState ระหว่าง build
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _syncPRFieldsToScreenData();
    });
    return DocumentPreviewWidget(
      screenData: screenData,
      transactionType: transactionType,
      docDateTimeValidated: docDateTimeValidated,
      fieldNameCustCode: fieldNameCustCode,
      fieldNameCustName: fieldNameCustName,
    );
  }

  Widget previewWidget() {
    return Container(
      margin: const EdgeInsets.all(5),
      padding: const EdgeInsets.all(10),
      height: 100,
      width: double.infinity,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(10),
        boxShadow: [
          BoxShadow(
            color: global.theme.dividerBorderColor.withValues(alpha: 0.5),
            spreadRadius: 5,
            blurRadius: 7,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: docPreview(),
    );
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
      showLocation: false,
      canViewPurchaseHistory: true,
      onDetailAdded: _autoFillLastPrice,
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
      onReorderItem: (oldIndex, newIndex) {
        if (screenData.details == null || screenData.details!.isEmpty) return;
        setState(() {
          final item = screenData.details!.removeAt(oldIndex);
          screenData.details!.insert(newIndex, item);
          for (int i = 0; i < screenData.details!.length; i++) {
            screenData.details![i].linenumber = i + 1;
          }
        });
      },
      persistentSearch: true,
    );
  }

  /// สรุปยอดรวม PR — ครบทุกข้อมูลเพื่อให้คนอนุมัติเห็นชัดเจน
  Widget _buildSummarySection() {
    final itemCount = screenData.details?.length ?? 0;
    final totalQty = screenData.totalqty ?? 0.0;
    final estUnit = _formController.estimatedUnitCostController.text;
    final estFreight = _formController.estimatedFreightController.text;
    final estDuty = _formController.estimatedDutyController.text;
    final hasEstimate = estUnit.isNotEmpty || estFreight.isNotEmpty || estDuty.isNotEmpty;
    final estUnitVal = double.tryParse(estUnit.replaceAll(',', '')) ?? 0.0;
    final estFreightVal = double.tryParse(estFreight.replaceAll(',', '')) ?? 0.0;
    final estDutyVal = double.tryParse(estDuty.replaceAll(',', '')) ?? 0.0;
    final otherExpenses = estUnitVal + estFreightVal + estDutyVal;
    final urgency = screenData.prUrgency ?? 1;
    final purpose = screenData.description ?? '';
    final supplier = screenData.custnames?.isNotEmpty == true ? screenData.custnames!.first.name : '';
    final costCenter = screenData.prCostCenterName ?? '';
    final jobName = screenData.prJobName ?? '';
    final approvalDeadline = screenData.prApprovalDeadline ?? '';

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // หัวข้อ
          Row(
            children: [
              Icon(Icons.receipt_long, size: 16, color: global.theme.primaryColor),
              const SizedBox(width: 6),
              Text(global.language('total'), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: global.theme.primaryColor)),
            ],
          ),
          const SizedBox(height: 8),

          // === ข้อมูลเอกสาร (จอใหญ่ = 2 คอลัมน์) ===
          LayoutBuilder(
            builder: (context, constraints) {
              final isWide = constraints.maxWidth > 500;

              final infoItems = <Widget>[
                _summaryInfoChip(icon: Icons.list_alt, label: global.language('line_number'), value: '$itemCount'),
                _summaryInfoChip(icon: Icons.inventory_2_outlined, label: global.language('product_qty'), value: global.formatNumber(totalQty)),
                if (supplier.isNotEmpty)
                  _summaryInfoChip(icon: Icons.store, label: global.language('supplier'), value: supplier),
                _summaryInfoChip(
                  icon: urgency == 3 ? Icons.warning : (urgency == 2 ? Icons.speed : Icons.check_circle_outline),
                  label: global.language('urgency'),
                  value: urgency == 3 ? global.language('urgency_critical') : (urgency == 2 ? global.language('urgency_urgent') : global.language('urgency_normal')),
                ),
                if (costCenter.isNotEmpty)
                  _summaryInfoChip(icon: Icons.account_balance, label: global.language('cost_center'), value: costCenter),
                if (jobName.isNotEmpty)
                  _summaryInfoChip(icon: Icons.work_outline, label: global.language('job'), value: jobName),
              ];

              if (isWide) {
                final rows = <Widget>[];
                for (var i = 0; i < infoItems.length; i += 2) {
                  rows.add(Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Row(children: [
                      Expanded(child: infoItems[i]),
                      const SizedBox(width: 8),
                      if (i + 1 < infoItems.length) Expanded(child: infoItems[i + 1]) else const Expanded(child: SizedBox()),
                    ]),
                  ));
                }
                return Column(children: rows);
              }
              return Column(children: infoItems.map((w) => Padding(padding: const EdgeInsets.only(bottom: 6), child: w)).toList());
            },
          ),

          // วัตถุประสงค์
          if (purpose.isNotEmpty) ...[
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
              decoration: BoxDecoration(color: global.theme.surfaceColor, borderRadius: BorderRadius.circular(8), border: Border.all(color: global.theme.dividerBorderColor)),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(Icons.description_outlined, size: 14, color: global.theme.iconSecondaryColor),
                  const SizedBox(width: 6),
                  Expanded(child: Text(purpose, style: TextStyle(fontSize: 12, color: global.theme.textColor), maxLines: 2, overflow: TextOverflow.ellipsis)),
                ],
              ),
            ),
            const SizedBox(height: 6),
          ],

          // กำหนดวันอนุมัติ
          if (approvalDeadline.isNotEmpty) ...[
            _summaryInfoChip(icon: Icons.event, label: global.language('approval_deadline'), value: approvalDeadline),
            const SizedBox(height: 6),
          ],

          const Divider(height: 8),

          // === ส่วนลดท้ายบิล ===
          TextFormField(
            controller: _formController.docDiscountWordController,
            decoration: InputDecoration(
              labelText: global.language('discount'),
              prefixIcon: Icon(Icons.discount_outlined, size: 18, color: global.theme.iconColor),
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              isDense: true,
            ),
            style: const TextStyle(fontSize: 14),
            onChanged: (_) {
              screenData.discountword = _formController.docDiscountWordController.text;
              _transactionCalculator.calTotalValue();
              setState(() {});
            },
          ),
          const SizedBox(height: 8),

          // === ยอดเงิน — ใช้ DocumentTotalWidget เหมือน PO ===
          DocumentTotalWidget(screenData: screenData),

          // === ประมาณการต้นทุน ===
          if (hasEstimate) ...[
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: global.theme.surfaceColor,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: global.theme.dividerBorderColor),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(global.language('estimated_landed_cost'), style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: global.theme.primaryColor)),
                  const SizedBox(height: 4),
                  if (estUnitVal > 0)
                    _summaryCell(global.language('estimated_unit_cost'), global.formatNumber(estUnitVal)),
                  if (estFreightVal > 0)
                    _summaryCell(global.language('estimated_freight'), global.formatNumber(estFreightVal)),
                  if (estDutyVal > 0)
                    _summaryCell(global.language('estimated_duty'), global.formatNumber(estDutyVal)),
                  if (otherExpenses > 0) ...[
                    const Divider(height: 8),
                    _summaryRow(global.language('estimated_landed_cost'), global.formatNumber(otherExpenses), isBold: true),
                  ],
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  /// Cell สรุปยอด — label ซ้าย value ขวา (compact สำหรับ 2-col layout)
  Widget _summaryCell(String label, String value, {bool isNegative = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
          Text(
            isNegative && value != '0' && value != '0.00' ? '-$value' : value,
            style: TextStyle(fontSize: 13, fontWeight: FontWeight.w500, color: isNegative ? global.theme.negativeHighlightTextColor : global.theme.textColor),
          ),
        ],
      ),
    );
  }

  /// Chip แสดงข้อมูลสรุป (จำนวนรายการ, จำนวนสินค้า)
  Widget _summaryInfoChip({required IconData icon, required String label, required String value}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Row(
        children: [
          Icon(icon, size: 16, color: global.theme.iconSecondaryColor),
          const SizedBox(width: 6),
          Text('$label: ', style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
          Text(value, style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: global.theme.textColor)),
        ],
      ),
    );
  }

  Widget _summaryRow(String label, String value, {bool isNegative = false, bool isBold = false, bool isHighlight = false}) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: TextStyle(
          fontSize: isBold ? 16 : 14,
          fontWeight: isBold ? FontWeight.bold : FontWeight.normal,
          color: global.theme.textColor,
        )),
        Text(
          isNegative && value != '0' ? '-$value' : value,
          style: TextStyle(
            fontSize: isBold ? 18 : 14,
            fontWeight: isBold ? FontWeight.bold : FontWeight.w500,
            color: isHighlight ? global.theme.primaryColor : (isNegative ? global.theme.negativeHighlightTextColor : global.theme.textColor),
          ),
        ),
      ],
    );
  }

  Widget editDocumentWidget() {
    return PRHeaderWidget(
      screenData: screenData,
      formController: _formController,
      setState: setState,
      parentContext: context,
      docDateTimeValidated: docDateTimeValidated,
      searchSupplier: _searchHandlers.searchSupplier,
      currencies: _currencies,
      selectedDocCurrencyCode: _selectedDocCurrencyCode,
      selectedDocCurrencySymbol: _selectedDocCurrencySymbol,
      selectedExchangeRate: _selectedExchangeRate,
      baseCurrency: _baseCurrency,
      baseCurrencySymbol: _baseCurrencySymbol,
      onCurrencyChanged: _onCurrencyChanged,
      onExchangeRateChanged: _onExchangeRateChanged,
    );
  }

  Widget editWidget() {
    // PR: 2 tabs — เอกสาร (header + summary ล่าง) + รายการสินค้า
    Widget documentHeader = SingleChildScrollView(
      child: Column(
        children: [
          editDocumentWidget(),
          // สรุปยอดรวม — เลื่อนไปด้วยกับ header
          _buildSummarySection(),
          POApprovalStatusBadge(
            screenData: screenData,
            approvalStatus: _poApprovalStatus,
            isLoading: _isLoadingApprovalStatus,
            onWithdrawApproval: _withdrawPRApproval,
            onApprove: _approvePR,
            onReject: _rejectPR,
          ),
        ],
      ),
    );
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
                  if (!_canEditPR()) {
                    context.ui.showWarning(_getPREditDisabledReason());
                    return;
                  }
                  final result = await showEnhancedSaveConfirmDialog(
                    context,
                    screenData.guidfixed ?? '',
                    showPrintOption: _canPrintPR(),
                  );
                  if (result != null && result.confirmed) {
                    final prefs = await SharedPreferences.getInstance();
                    await prefs.setBool(_printAfterSaveKey, result.printAfterSave);
                    saveOrUpdateData(taxInvoice: false);
                  }
                }
              }
            }
          }
        },
        child: Focus(
          skipTraversal: true,
          child: Builder(
            builder: (context) {
              final scaleFactor = global.editFontScaleFactor;
              return MediaQuery(
                data: MediaQuery.of(context).copyWith(
                  textScaler: TextScaler.linear(scaleFactor),
                ),
                child: Theme(
                  data: Theme.of(context).copyWith(
                    inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
                      contentPadding: EdgeInsets.symmetric(
                        horizontal: 12 * scaleFactor,
                        vertical: 10 * scaleFactor,
                      ),
                    ),
                    iconTheme: IconThemeData(size: 24 * scaleFactor),
                  ),
                  child: TabBarView(
                    controller: editTabController,
                    physics: const NeverScrollableScrollPhysics(),
                    children: childrenList,
                  ),
                ),
              );
            },
          ),
        ),
      ),
    );
  }

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
        _loadPRApprovalStatus(document.docno);
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
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            global.gotoMainMenu(context);
          },
        ),
        backgroundColor: global.theme.appBarColor,
        title: Text(global.transactionName(transactionType)),
        actions: [
          // === Font Size (ซ้ายสุด — ใช้น้อย) ===
          EditFontSizeControl(onChanged: () => setState(() {})),
          // === ไฟล์แนบ ===
          if ((screenData.guidfixed ?? '').isNotEmpty)
            AttachmentCountIconButton(
              docNo: screenData.docno,
              guidFixed: screenData.guidfixed ?? '',
              screenType: 'purchaserequisition',
            ),
          // === คัดลอกเอกสาร ===
          if ((screenData.guidfixed ?? '').isNotEmpty)
            Tooltip(
              message: global.language('copy_document'),
              child: IconButton(
                icon: const Icon(Icons.copy),
                onPressed: _copyDocument,
              ),
            ),
          // === แปลงเอกสาร (PR → RFQ / PO) ===
          if (_isPRApproved()) ...[
            Tooltip(
              message: global.language('convert_to_rfq'),
              child: IconButton(
                icon: Icon(Icons.request_quote_outlined),
                onPressed: _convertToRFQ,
              ),
            ),
            Tooltip(
              message: global.language('convert_to_po'),
              child: IconButton(
                icon: Icon(Icons.shopping_cart_outlined),
                onPressed: _convertToPO,
              ),
            ),
          ],
          // === POAppBarActions: ลบ ยกเลิก ปิด | ประวัติ x3 | เพิ่มใหม่ Preview ค้นหา บันทึก ===
          ...POAppBarActions.build(
            context: context,
            screenData: screenData,
            screenDataTemp: screenDataTemp,
            transactionType: transactionType,
            isLoading: _isLoading,
            showPreview: _showPreview,
            canEditPO: _canEditPR(),
            canPrintPO: _canPrintPR(),
            editDisabledReason: _getPREditDisabledReason(),
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
        ],
      ),
      body: MultiBlocListener(
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
        child: PRResponsiveLayout(
          splitViewController: splitViewController,
          tabController: tabController,
          showPreview: _showPreview,
          showDocSearchPanel: _showDocSearchPanel,
          docSearchPanelWidth: _docSearchPanelWidth,
          prListKey: _prListKey,
          buildEditWidget: editWidget,
          buildPreviewWidget: previewWidget,
          onDocumentSelected: _onDocumentSelectedFromPanel,
          onClosePanel: () => setState(() => _showDocSearchPanel = false),
          onPanelWidthChanged: (width) => setState(() => _docSearchPanelWidth = width),
        ),
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
