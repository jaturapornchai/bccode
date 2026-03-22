// ignore_for_file: deprecated_member_use

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
import 'package:smlaicloud/screens/rfq/rfq_list_screen.dart';
import 'package:smlaicloud/screens/rfq/components/rfq_vendor_widget.dart';
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
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/widgets/attachment_count_icon_button.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_ui_utils.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_helper.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_product_search_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_approval_status_badge.dart';
import 'package:smlaicloud/screens/purchaseorder/components/po_appbar_actions.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_detail_command_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_workflow_manager.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_bloc_effect_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_approval_handler.dart';
import 'package:smlaicloud/screens/purchaseorder/utils/po_save_helper.dart';
import 'package:smlaicloud/screens/purchaseorder/models/po_warehouse_state.dart';
import 'package:smlaicloud/screens/purchaserequisition/models/pr_model_factory.dart';

enum RFQEditScreenModule { header, detail, vendor }

class RFQEditScreen extends StatefulWidget {
  const RFQEditScreen({super.key});

  @override
  State<RFQEditScreen> createState() => RFQEditScreenState();
}

class RFQEditScreenState extends State<RFQEditScreen> with TickerProviderStateMixin {
  final global.TransactionTypeEnum transactionType = global.TransactionTypeEnum.rfq;
  late SplitViewController splitViewController;
  late TabController tabController;
  late TabController editTabController;
  late TransactionModel screenData;
  late TransactionModel screenDataTemp;
  late RFQEditScreenModule moduleEdit;
  final ProductBarcodeRepository productBarcodeRepository = ProductBarcodeRepository();

  late final TransactionDialogs _transactionDialogs = TransactionDialogs(this);
  late final TransactionCalculatorBridge _transactionCalculator = TransactionCalculatorBridge(this);
  TransactionCalculator get transactionCalculator => _transactionCalculator;
  late final TransactionSearchHandlers _searchHandlers = TransactionSearchHandlers(this);

  bool get showDocSearchPanel => _showDocSearchPanel;
  set showDocSearchPanel(bool value) => _showDocSearchPanel = value;

  // ==== Workflow Manager ====
  POWorkflowManager get _workflowManager => POWorkflowManager(
        context: context,
        isMounted: () => mounted,
        getDocNo: () => screenData.docno,
        getCollectionName: _getCollectionNameForType,
        getDocumentTitle: _getDocumentTitle,
        submitPOForApproval: _submitRFQForApproval,
        onClearScreen: clearScreenData,
        onShowSaveDialog: (docNo) => showSaveDocNoDialog(context, docNo),
        onRegenerateDocNoAndSave: _regenerateDocNoAndSave,
        editTabController: editTabController,
        checkIsMultiCurrency: () => false,
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
        onSubmitPOAndHandlePrint: _submitRFQAndHandlePrint,
        onSubmitPOAndHandleUpdatePrint: _submitRFQAndHandleUpdatePrint,
        onHandleSaveSuccessWithPrintCheck: _handleSaveSuccessWithPrintCheck,
        onHandleUpdateSuccessWithPrintCheck: (transType) => _workflowManager.handleUpdateSuccessWithPrintCheck(transType),
        onRefreshListAfterSave: _refreshListAfterSave,
        onRefreshListAfterDelete: _refreshListAfterDelete,
        onLoadPurchaseTypesAndExisting: _loadPurchaseTypesAndExisting,
        onLoadCurrencies: () {},
        onLoadPOApprovalStatus: _loadRFQApprovalStatus,
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
    if (_showDocSearchPanel && deletedGuid.isNotEmpty && _rfqListKey.currentState != null) {
      _rfqListKey.currentState!.refreshListAfterDelete(deletedGuid);
    }
  }

  void _refreshListAfterSave() {
    if (!_showDocSearchPanel) return;
    final guidToSelect = screenData.guidfixed;
    if (guidToSelect != null && _rfqListKey.currentState != null) {
      _rfqListKey.currentState!.setCardLoadingState(guidToSelect);
    }
    Future.delayed(const Duration(milliseconds: 2000), () {
      if (mounted && _rfqListKey.currentState != null) {
        _rfqListKey.currentState!.refreshListAndMaintainPosition(
          selectDocGuid: guidToSelect,
        );
      }
    });
  }

  // ==== RFQ Vendor Entries ====
  List<Map<String, dynamic>> _vendorEntries = [];

  // ==== Common State ====
  bool _showDocSearchPanel = false;
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
  double _docSearchPanelWidth = 550.0;

  final GlobalKey<RFQListScreenState> _rfqListKey = GlobalKey<RFQListScreenState>();

  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  int customerPriceLevel = 1;
  String? clientId;

  List<global.DataTableHeader> headers = [];

  List<dynamic> docrefs = [];

  // ==== Approval Handler ====
  late final POApprovalHandler _approvalHandler = POApprovalHandler(
    onStateChanged: () => setState(() {}),
    onShowSuccess: (msg) => context.ui.showSuccess(msg),
    onShowInfo: (msg) => context.ui.showInfo(msg),
    onShowError: (msg) => context.ui.showError(msg),
  );

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
    fieldNameCustCode = "doc_supplier_code";
    fieldNameCustName = "doc_supplier_name";

    setSystemLanguageList();
    splitViewController = SplitViewController(limits: [null, WeightLimit(min: 0.1, max: 0.9)]);
    splitViewController.weights = [0.30, 0.70];
    tabController = TabController(length: 2, vsync: this);
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

  bool _hasProcessedConversion = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();

    // ตรวจสอบ PR → RFQ conversion arguments
    if (!_hasProcessedConversion) {
      final args = ModalRoute.of(context)?.settings.arguments;
      if (args is Map<String, dynamic> && args.containsKey('source')) {
        _hasProcessedConversion = true;
        WidgetsBinding.instance.addPostFrameCallback((_) {
          _populateFromConversion(args);
        });
      }
    }
  }

  /// Pre-fill RFQ จาก PR ที่อนุมัติแล้ว
  void _populateFromConversion(Map<String, dynamic> args) {
    try {
      final source = args['source'] as String;
      final sourceDocNo = args['sourceDocNo'] as String? ?? '';
      final custcode = args['custcode'] as String? ?? '';
      final custnames = args['custnames'];
      final description = args['description'] as String? ?? '';
      final creditdays = args['creditdays'];
      final detailsJson = args['details'] as List?;

      // ตั้งค่า supplier
      screenData.custcode = custcode;
      if (custnames is List) {
        screenData.custnames = custnames.cast<LanguageDataModel>();
      }

      // ตั้งค่า description + docrefno (อ้างอิง PR)
      screenData.description = description;
      screenData.docrefno = sourceDocNo;

      // ตั้งค่า credit days
      if (creditdays is int && creditdays > 0) {
        screenData.creditdays = creditdays;
      }

      // โหลด details (รายการสินค้า)
      if (detailsJson != null && detailsJson.isNotEmpty) {
        screenData.details = detailsJson
            .map((d) => TransactionDetailModel.fromJson(d as Map<String, dynamic>))
            .toList();
        for (int i = 0; i < screenData.details!.length; i++) {
          screenData.details![i].linenumber = i + 1;
          screenData.details![i].docref = '';
        }
      }

      setState(() {
        loadDataToScreen();
        _transactionCalculator.calTotalValue();
      });

      AppLogger.info('[RFQ] Pre-fill จาก $source: $sourceDocNo (${screenData.details?.length ?? 0} รายการ)');
    } catch (e, stackTrace) {
      AppLogger.error('[RFQ] Error populating from conversion: $e\n$stackTrace');
    }
  }

  @override
  void dispose() {
    _debouncer.cancel();
    docrefs.clear();
    splitViewController.dispose();
    tabController.dispose();
    editTabController.dispose();
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
    ];
  }

  DateTime _safeParseDatetime(String? dateString) {
    if (dateString == null || dateString.isEmpty) return DateTime.now();
    try {
      return DateTime.parse(dateString);
    } catch (e) {
      AppLogger.warning('[RFQ] แปลงวันที่ไม่สำเร็จ: "$dateString" → ใช้ DateTime.now() แทน');
      return DateTime.now();
    }
  }

  void loadDataToScreen() {
    headerTableDetail();
    _transactionCalculator.calPayTotal();
  }

  String _getCollectionNameForType(global.TransactionTypeEnum type) {
    return "transactionRFQ";
  }

  String _getDocumentTitle(global.TransactionTypeEnum type) {
    return global.language("request_for_quotation");
  }

  Future<void> _submitRFQAndHandlePrint(String docNo, global.TransactionTypeEnum transType) async {
    await _workflowManager.submitAndHandlePrint(docNo, transType);
  }

  Future<void> _submitRFQAndHandleUpdatePrint(global.TransactionTypeEnum transType) async {
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
      AppLogger.info('[RFQ] ล้าง docno เพื่อให้ mainapi สร้างใหม่');
    });
    saveOrUpdateData(taxInvoice: false);
  }

  void clearScreenData() {
    transflag = 22;
    calcflag = 1;
    docnoFormat = "RFQ";
    configvattype = global.config.vattypepurchase;
    configinquirytype = global.config.inquirytypepurchase;
    moduleEdit = RFQEditScreenModule.header;

    screenData = PRModelFactory.createDefault(
      transflag: 22,
      vattype: configvattype,
      inquirytype: configinquirytype,
      timezoneInfo: global.getLocalTimezoneInfo(),
    );
    screenDataTemp = PRModelFactory.createDefault(
      transflag: 22,
      vattype: configvattype,
      inquirytype: configinquirytype,
      timezoneInfo: global.getLocalTimezoneInfo(),
    );

    customerPriceLevel = 1;
    _vendorEntries = [];

    _approvalHandler.resetPurchaseType();
    _approvalHandler.resetApprovalStatus();

    setState(() {
      loadDataToScreen();
      docDateTimeValidated = true;
    });

    _loadPurchaseTypesForRFQ();
  }

  Future<void> _loadPurchaseTypesForRFQ() async {
    if (!mounted) return;
    await _approvalHandler.loadPurchaseTypesForNewPO();
  }

  Future<void> _loadPurchaseTypesAndExisting(String docNo) async {
    if (!mounted) return;
    await _approvalHandler.loadPurchaseTypesAndExisting(screenData);
  }

  void _saveRFQPurchaseType() {
    _approvalHandler.savePurchaseTypeToModel(screenData);
  }

  Future<void> _loadRFQApprovalStatus(String docNo) async {
    if (!mounted) return;
    await _approvalHandler.loadApprovalStatus(docNo);
  }

  Future<bool> _submitRFQForApproval({bool isModified = false}) async {
    return await _approvalHandler.submitForApproval(
      screenData: screenData,
      description: screenData.description ?? '',
      isModified: isModified,
    );
  }

  Future<bool> _withdrawRFQApproval() async {
    return await _approvalHandler.withdrawApproval(
      docNo: screenData.docno,
    );
  }

  Future<bool> _approveRFQ(String comment) async {
    return await _approvalHandler.approveApproval(
      docNo: screenData.docno,
      comment: comment,
    );
  }

  Future<bool> _rejectRFQ(String comment) async {
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

  bool _canPrintRFQ() {
    return POApprovalHelper.canPrint(_poApprovalStatus);
  }

  bool _canEditRFQ() {
    if (screenData.docno.isEmpty) return true;
    return POApprovalHelper.canEdit(
      docNo: screenData.docno,
      approvalStatus: _poApprovalStatus,
      isref: screenData.isref ?? false,
      creatorCode: screenData.creatorcode,
      currentUserCode: global.profileData.username,
    );
  }

  String _getRFQEditDisabledReason() {
    if (screenData.docno.isEmpty) return '';
    return POApprovalHelper.getEditDisabledReason(
      docNo: screenData.docno,
      approvalStatus: _poApprovalStatus,
      isref: screenData.isref ?? false,
      creatorCode: screenData.creatorcode,
      currentUserCode: global.profileData.username,
    );
  }

  // ===== Conversion Button (RFQ → PO) =====

  /// ตรวจสอบว่า RFQ อนุมัติแล้วหรือไม่
  bool _isRFQApproved() {
    return _poApprovalStatus?.isCompleted == true &&
        (screenData.guidfixed ?? '').isNotEmpty;
  }

  /// สร้าง PO จาก RFQ ที่อนุมัติแล้ว
  void _convertToPO() {
    if (!_isRFQApproved()) return;
    // หา vendor ที่ถูกเลือก
    final selectedVendor = _vendorEntries.firstWhere(
      (v) => v['isselected'] == true,
      orElse: () => <String, dynamic>{},
    );
    Navigator.pushNamed(
      context,
      '/transaction/purchaseorder',
      arguments: {
        'source': 'rfq',
        'sourceDocNo': screenData.docno,
        'sourceGuid': screenData.guidfixed,
        'custcode': selectedVendor.isNotEmpty ? (selectedVendor['vendorcode'] ?? screenData.custcode) : screenData.custcode,
        'custnames': selectedVendor.isNotEmpty ? (selectedVendor['vendornames'] ?? screenData.custnames) : screenData.custnames,
        'details': screenData.details?.map((d) => d.toJson()).toList(),
        'description': screenData.description,
        'creditdays': selectedVendor.isNotEmpty ? (selectedVendor['creditdays'] ?? screenData.creditdays) : screenData.creditdays,
      },
    );
  }

  // ===== Save / Update / Delete =====

  void saveOrUpdateData({bool taxInvoice = false}) {
    _transactionCalculator.calTotalValue();

    if (!_transactionCalculator.verifyPayment()) return;

    POSaveHelper.prepareDataForSave(
      screenData,
      creatorCode: global.profileData.username ?? '',
      creatorName: global.profileData.name ?? '',
    );

    _saveRFQPurchaseType();

    final extraFields = <String, dynamic>{};
    // RFQ-specific: vendorentries
    if (_vendorEntries.isNotEmpty) {
      extraFields['vendorentries'] = _vendorEntries;
    }

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
                    collection: 'transactionRFQ',
                    docNo: screenData.docno,
                    title: global.transactionName(transactionType),
                    context: this.context,
                    isMultiCurrency: false,
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
    return 0;
  }

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
    screenData.docreferences!.removeAt(index);
    screenData.details!.removeWhere((detail) => detail.docref == docno);
    docrefs.removeWhere((docref) => docref['docref'] == docno);
    _transactionCalculator.calTotalValue();
    if (mounted) setState(() {});
    context.ui.showSuccess('${global.language("delete_reference_and_products")} $docno');
  }

  // ===== Widget Builders =====

  Widget docPreview() {
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
    );
  }

  Widget editWidget() {
    return RawKeyboardListener(
      focusNode: FocusNode(skipTraversal: true),
      onKey: (event) async {
        if (kIsWeb || Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
          if (event is RawKeyUpEvent) {
            if (event.logicalKey == LogicalKeyboardKey.f10) {
              if (screenData.details!.isNotEmpty) {
                if (!_canEditRFQ()) {
                  context.ui.showWarning(_getRFQEditDisabledReason());
                  return;
                }
                final result = await showEnhancedSaveConfirmDialog(
                  context,
                  screenData.guidfixed ?? '',
                  showPrintOption: _canPrintRFQ(),
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
                child: Column(
                  children: [
                    // === Approval Status ===
                    POApprovalStatusBadge(
                      screenData: screenData,
                      approvalStatus: _poApprovalStatus,
                      isLoading: _isLoadingApprovalStatus,
                      onWithdrawApproval: _withdrawRFQApproval,
                      onApprove: _approveRFQ,
                      onReject: _rejectRFQ,
                    ),
                    // === Tab: รายการสินค้า + Vendor Entries ===
                    Expanded(
                      child: DefaultTabController(
                        length: 2,
                        child: Column(
                          children: [
                            TabBar(
                              labelColor: global.theme.textColor,
                              unselectedLabelColor: global.theme.textSecondaryColor,
                              indicatorColor: global.theme.infoHighlightTextColor,
                              tabs: [
                                Tab(text: global.language('product_list')),
                                Tab(
                                  child: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Text(global.language('vendor')),
                                      if (_vendorEntries.isNotEmpty) ...[
                                        const SizedBox(width: 6),
                                        Container(
                                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                          decoration: BoxDecoration(
                                            color: global.theme.infoHighlightTextColor,
                                            borderRadius: BorderRadius.circular(10),
                                          ),
                                          child: Text(
                                            '${_vendorEntries.length}',
                                            style: const TextStyle(fontSize: 11, color: Colors.white),
                                          ),
                                        ),
                                      ],
                                    ],
                                  ),
                                ),
                              ],
                            ),
                            Expanded(
                              child: TabBarView(
                                children: [
                                  editProductListWidget(),
                                  RFQVendorWidget(
                                    vendorEntries: _vendorEntries,
                                    onVendorEntriesChanged: (entries) {
                                      setState(() => _vendorEntries = entries);
                                    },
                                    isEditable: _canEditRFQ(),
                                    details: screenData.details ?? [],
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
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
        _loadRFQApprovalStatus(document.docno);
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
          if ((screenData.guidfixed ?? '').isNotEmpty)
            AttachmentCountIconButton(
              docNo: screenData.docno,
              guidFixed: screenData.guidfixed ?? '',
              screenType: 'rfq',
            ),
          // === Conversion Button (RFQ → PO) ===
          if (_isRFQApproved())
            Tooltip(
              message: global.language('convert_to_po'),
              child: IconButton(
                icon: Icon(Icons.shopping_cart_outlined),
                onPressed: _convertToPO,
              ),
            ),
          EditFontSizeControl(onChanged: () => setState(() {})),
          ...POAppBarActions.build(
            context: context,
            screenData: screenData,
            screenDataTemp: screenDataTemp,
            transactionType: transactionType,
            isLoading: _isLoading,
            showPreview: _showPreview,
            canEditPO: _canEditRFQ(),
            canPrintPO: _canPrintRFQ(),
            editDisabledReason: _getRFQEditDisabledReason(),
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
                child: _buildResponsiveLayout(constraints),
              );
            },
          ),
        ],
      ),
    );
  }

  Widget _buildResponsiveLayout(BoxConstraints constraints) {
    final isLargeScreen = constraints.maxWidth > 700;

    Widget mainContent;
    if (isLargeScreen) {
      mainContent = _showPreview
          ? SplitView(
              controller: splitViewController,
              gripSize: 8,
              gripColor: global.theme.appBarColor,
              gripColorActive: global.theme.infoHighlightTextColor,
              viewMode: SplitViewMode.Horizontal,
              indicator: const SplitIndicator(viewMode: SplitViewMode.Horizontal),
              activeIndicator: const SplitIndicator(viewMode: SplitViewMode.Horizontal, isActive: true),
              children: [previewWidget(), editWidget()],
            )
          : editWidget();
    } else {
      mainContent = TabBarView(
        physics: const NeverScrollableScrollPhysics(),
        controller: tabController,
        children: _showPreview ? [editWidget(), previewWidget()] : [editWidget()],
      );
    }

    // Side panel ค้นหาเอกสาร
    if (_showDocSearchPanel && isLargeScreen) {
      const minWidth = 450.0;
      final maxWidth = constraints.maxWidth * 0.75;

      return Row(
        children: [
          SizedBox(
            width: _docSearchPanelWidth.clamp(minWidth, maxWidth),
            child: RFQListScreen(
              key: _rfqListKey,
              custcode: '',
              isEmbedded: true,
              onDocumentSelected: _onDocumentSelectedFromPanel,
              onClose: () => setState(() => _showDocSearchPanel = false),
            ),
          ),
          _buildResizeHandle(minWidth, maxWidth),
          Expanded(child: mainContent),
        ],
      );
    }

    return mainContent;
  }

  Widget _buildResizeHandle(double minWidth, double maxWidth) {
    return MouseRegion(
      cursor: SystemMouseCursors.resizeColumn,
      child: GestureDetector(
        onHorizontalDragUpdate: (details) {
          double newWidth = _docSearchPanelWidth + details.delta.dx;
          newWidth = newWidth.clamp(minWidth, maxWidth);
          setState(() => _docSearchPanelWidth = newWidth);
        },
        child: Container(
          width: 8,
          color: Colors.transparent,
          child: Center(
            child: Container(
              width: 4,
              height: 40,
              decoration: BoxDecoration(
                color: global.theme.dividerBorderColor,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),
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
