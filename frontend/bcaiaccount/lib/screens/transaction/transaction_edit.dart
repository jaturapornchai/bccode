// ignore_for_file: deprecated_member_use

import 'dart:convert';
import 'dart:io';
import 'dart:async';
import 'package:smlaicloud/bloc/export_csv/export_csv_bloc.dart';
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/book_bank_model.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/pdfgen/pdfpreview.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';
import 'package:smlaicloud/screen_search/bookbank_select_screen.dart';
import 'package:smlaicloud/screen_search/transaction_search_screen.dart';

import 'package:smlaicloud/components/numpad.dart';

import 'package:smlaicloud/screens/transaction/components/document_header_widget.dart';
import 'package:smlaicloud/screens/transaction/components/document_preview_widget.dart';
import 'package:smlaicloud/screens/transaction/components/document_product_list_widget.dart';
import 'package:smlaicloud/screens/transaction/widgets/payment_widgets.dart';
import 'package:smlaicloud/screens/transaction/widgets/summary_widgets.dart';
import 'package:smlaicloud/screens/transaction/widgets/mobile_product_list_widgets.dart';
import 'package:smlaicloud/screens/transaction/widgets/desktop_product_widgets.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_dialogs.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_calculator.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_calculator_bridge.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_search_handlers.dart';
import 'package:smlaicloud/services/pdf_service.dart';
import 'package:smlaicloud/utils/cart_websocket_service.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:split_view/split_view.dart';
import 'package:uuid/uuid.dart';
import 'package:smlaicloud/global.dart' as global;
// ignore: library_prefixes
import 'package:smlaicloud/calamount.dart' as calAmount;
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/modules/cart-system/models/cart_transfer_model.dart' show CartTransferModel, CartTransferProvider;
import 'package:smlaicloud/modules/cart-system/models/cart_system_type.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/widgets/po_approval_history_dialog.dart';
import 'package:smlaicloud/widgets/data_history_dialog.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_currency_utils.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_ui_utils.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';

enum TransactionEditScreenModule { header, detail, footer }

class TransactionEditScreen extends StatefulWidget {
  final global.TransactionTypeEnum type;

  const TransactionEditScreen({super.key, required this.type});

  @override
  State<TransactionEditScreen> createState() => TransactionEditScreenState();
}

class TransactionEditScreenState extends State<TransactionEditScreen> with TickerProviderStateMixin {
  late SplitViewController splitViewController;
  late TabController tabController;
  late TabController editTabController;
  late TransactionModel screenData;
  late TransactionModel screenDataTemp;
  late TransactionEditScreenModule moduleEdit;
  final ProductBarcodeRepository productBarcodeRepository = ProductBarcodeRepository();

  /// Helper instance สำหรับ payment widgets
  late final PaymentWidgets _paymentWidgets = PaymentWidgets(this);

  /// Helper instance สำหรับ dialog functions
  late final TransactionDialogs _transactionDialogs = TransactionDialogs(this);

  /// Helper instance สำหรับ calculator functions
  late final TransactionCalculatorBridge _transactionCalculator = TransactionCalculatorBridge(this);

  /// Public getter สำหรับให้ helper classes เข้าถึง calculator
  TransactionCalculator get transactionCalculator => _transactionCalculator;

  /// Helper instance สำหรับ search handler functions
  late final TransactionSearchHandlers _searchHandlers = TransactionSearchHandlers(this);

  /// Helper instance สำหรับ summary widgets
  late final SummaryWidgets _summaryWidgets = SummaryWidgets(this);

  /// Helper instance สำหรับ mobile product list widgets
  late final MobileProductListWidgets _mobileProductListWidgets = MobileProductListWidgets(this);

  /// Helper instance สำหรับ desktop product widgets
  late final DesktopProductWidgets _desktopProductWidgets = DesktopProductWidgets(this);

  /// Public getter สำหรับให้ helper classes เข้าถึง paymentWidgets
  PaymentWidgets get paymentWidgets => _paymentWidgets;

  /// Public getter/setter สำหรับ showDocSearchPanel (ให้ helper class เข้าถึงได้)
  bool get showDocSearchPanel => _showDocSearchPanel;
  set showDocSearchPanel(bool value) => _showDocSearchPanel = value;

  TextEditingController docNumberController = TextEditingController();
  TextEditingController docDateController = TextEditingController();
  TextEditingController docTimeController = TextEditingController();
  TextEditingController docRefNumberController = TextEditingController();
  TextEditingController docRefDateController = TextEditingController();
  TextEditingController taxDocNoController = TextEditingController();
  TextEditingController taxDocDateController = TextEditingController();
  TextEditingController docRefTypeController = TextEditingController();
  TextEditingController docTypeController = TextEditingController();
  TextEditingController vatTypeController = TextEditingController();
  TextEditingController custCodeController = TextEditingController();
  TextEditingController custnamesController = TextEditingController();
  TextEditingController saleCodeController = TextEditingController();
  TextEditingController saleNameController = TextEditingController();
  TextEditingController discountWordController = TextEditingController();
  TextEditingController totalCostController = TextEditingController();
  TextEditingController totalValueController = TextEditingController();
  TextEditingController totalDiscountController = TextEditingController();
  TextEditingController totalVatValueController = TextEditingController();
  TextEditingController totalAfterVatController = TextEditingController();
  TextEditingController totalExceptVatController = TextEditingController();
  TextEditingController totalAmountController = TextEditingController();
  TextEditingController cashierCodeController = TextEditingController();
  TextEditingController posIdController = TextEditingController();
  TextEditingController memberCodeController = TextEditingController();
  TextEditingController vatRateController = TextEditingController();
  TextEditingController statusController = TextEditingController();
  TextEditingController docDiscountWordController = TextEditingController();
  TextEditingController descriptionController = TextEditingController();
  TextEditingController detailTotalDiscount = TextEditingController();
  TextEditingController totalDiscountVatAmount = TextEditingController();
  TextEditingController totalDiscountExceptVatAmount = TextEditingController();
  TextEditingController roundAmount = TextEditingController();
  TextEditingController totalAmountAfterDiscount = TextEditingController();
  TextEditingController detailDiscountFormula = TextEditingController();
  TextEditingController detailTotalAmount = TextEditingController();

  TextEditingController payCashAmountController = TextEditingController();
  TextEditingController roundAmountController = TextEditingController();
  TextEditingController detaildiscountformulaController = TextEditingController();
  TextEditingController payDeliveryCashAmountController = TextEditingController();

  TextEditingController transportAmountController = TextEditingController();

  List<TextEditingController> payTransferAmountController = [];
  List<TextEditingController> payCreditCardAmountController = [];
  List<TextEditingController> payChequeAmountController = [];
  List<TextEditingController> payCouponAmountController = [];
  List<TextEditingController> payQrAmountController = [];

  List<TextEditingController> creditCardDateController = [];
  List<TextEditingController> transferDateController = [];
  List<TextEditingController> chequeDateController = [];
  List<TextEditingController> chequeDueDateDateController = [];
  List<TextEditingController> couponDateController = [];
  List<TextEditingController> qrDateController = [];

  late DateTime selectedDate = DateTime.now();
  late List<WarehouseModel> warehouseList = [];
  String defaultWarehouse = "";
  List<LanguageDataModel> defaultWarehouseNames = [];
  String defaultLocation = "";
  List<LanguageDataModel> defaultLocationNames = [];
  String defaultToWarehouse = "";
  List<LanguageDataModel> defaultToWarehouseNames = [];
  String defaultToLocation = "";
  List<LanguageDataModel> defaultToLocationNames = [];
  List<Widget> tab = [];
  List<Widget> childrens = [];
  List<String> groupedItems = [];
  int transflag = 0;
  int calcflag = 0;
  bool docDateTimeValidated = false;
  late String fieldNameCustCode;
  late String fieldNameCustName;
  final _debouncer = global.Debouncer(1000);
  int showPayDetail = 0;
  String docnoFormat = "";
  bool isMember = false;
  int configvattype = 0;
  int configinquirytype = 0;
  double payTotalBill = 0;
  String enumtype = '';
  bool _isLoading = false;
  bool _showPreview = false; // ค่าเริ่มต้นไม่ต้องแสดง
  bool _hasCheckedScreenSize = false; // ตรวจสอบว่าได้ตรวจสอบขนาดหน้าจอแล้วหรือยัง
  bool _showDocSearchPanel = false; // แสดง side panel ค้นหาเอกสาร (จอใหญ่)
  double _docSearchPanelWidth = 450.0; // ความกว้าง side panel (ปรับขนาดได้)

  // GlobalKey สำหรับเข้าถึง TransSearchScreen state เพื่อ refresh list หลังจาก save/delete
  final GlobalKey<TransSearchScreenState> _transSearchKey = GlobalKey<TransSearchScreenState>();

  // Key สำหรับ SharedPreferences เก็บค่า printAfterSave (ใช้ SharedPreferences แทน instance variable เพื่อป้องกัน Hot Reload reset)
  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  //  pricelevel ของลูกค้า
  int customerPriceLevel = 1; // ค่าเริ่มต้นเป็น 1

  final CartWebSocketService _cartService = CartWebSocketService();
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

  List<BillPayStruct> payTransfer = [];
  List<BillPayStruct> payCreditCard = [];
  List<BillPayStruct> payCheque = [];
  List<BillPayStruct> payCoupon = [];
  List<BillPayStruct> payQr = [];

  /// model ใหม่จาก pos
  List<CouponModel>? couPons = [];

  List<dynamic> docrefs = [];
  List<String> cartList = [];

  // ประเภทการจัดซื้อ (สำหรับ PO)
  List<PurchaseTypeModel> _purchaseTypes = [];
  String? _selectedPurchaseTypeCode;
  String? _selectedPurchaseTypeName;

  // สกุลเงิน (สำหรับ PO Multi-Currency)
  List<CurrencyModel> _currencies = [];
  String? _selectedDocCurrencyCode;   // สกุลเงินเอกสาร (Document Currency)
  String? _selectedDocCurrencySymbol; // สัญลักษณ์สกุลเงินเอกสาร
  double? _selectedExchangeRate;      // อัตราแลกเปลี่ยน
  String? _baseCurrency;              // สกุลเงินหลัก (Base) สำหรับลงบัญชี - เก็บใน field currency เดิม
  String? _baseCurrencySymbol;        // สัญลักษณ์สกุลเงินหลัก

  // สถานะการอนุมัติ PO
  POApprovalStatusModel? _poApprovalStatus;
  bool _isLoadingApprovalStatus = false;

  void setSystemLanguageList() async {
    clearScreenData();
    await global.setSystemLanguage(context);

    headerTableDetail();
  }

  @override
  void initState() {
    super.initState();
    // ตั้งค่า callbacks สำหรับ CartWebSocketService
    _cartService.initialize(
      onClientIdReceived: (receivedClientId) {
        if (mounted) {
          setState(() {
            clientId = receivedClientId;
          });
        }
      },
      onCartDeleted: (success, message) {
        if (kDebugMode) {
          AppLogger.debug('ผลการลบตะกร้า: ${success ? "สำเร็จ" : "ไม่สำเร็จ"} - $message');
        }
      },
      // callback อื่นๆ สามารถเพิ่มตามต้องการ
    );

    // เชื่อมต่อกับ WebSocket
    _cartService.connect(context);

    global.getDeviceModel(context);
    switch (widget.type) {
      case global.TransactionTypeEnum.purchase:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.purchaseorder:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.purchasepartial:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.purchasereturn:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.accrualreceive:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.sale:
        fieldNameCustCode = "doc_customer_code";
        fieldNameCustName = "doc_customer_name";
        break;
      case global.TransactionTypeEnum.saleorder:
        fieldNameCustCode = "doc_customer_code";
        fieldNameCustName = "doc_customer_name";
        break;
      case global.TransactionTypeEnum.salereturn:
        fieldNameCustCode = "doc_customer_code";
        fieldNameCustName = "doc_customer_name";
        break;
      case global.TransactionTypeEnum.advancePayment:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.advancePaymentRefund:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.deposit:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.depositRefund:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      case global.TransactionTypeEnum.receiveDepositRefund:
        fieldNameCustCode = "doc_supplier_code";
        fieldNameCustName = "doc_supplier_name";
        break;
      default:
        fieldNameCustCode = "";
        fieldNameCustName = "";
        break;
    }

    setSystemLanguageList();
    splitViewController = SplitViewController(limits: [null, WeightLimit(min: 0.1, max: 0.9)]);
    splitViewController.weights = [0.30, 0.70];
    tabController = TabController(length: 2, vsync: this);

    if (widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.saleorder &&
        widget.type != global.TransactionTypeEnum.purchaseorder &&
        widget.type != global.TransactionTypeEnum.quotation &&
        widget.type != global.TransactionTypeEnum.purchasepartial) {
      editTabController = TabController(length: 3, vsync: this);
    } else {
      editTabController = TabController(length: 2, vsync: this);
    }

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
  void _populateFromCartTransfer(CartTransferModel cartTransfer) {
    try {
      // อัพเดทข้อมูลหัวเอกสาร (header) ตามประเภทระบบ
      // ระบบขาย (sale, salesOrder, quotation) ใช้ข้อมูล debtor (ลูกค้า)
      // ระบบซื้อ (purchaseOrder) ใช้ข้อมูล creditor (ผู้ขาย)
      // ระบบสต๊อก (withdraw, transfer) ไม่ต้องการข้อมูล customer

      String customerCode = '';
      String customerName = '';

      switch (cartTransfer.systemType) {
        case CartSystemType.sales:
          // ระบบขาย: ใช้ข้อมูล debtor
          customerCode = cartTransfer.debtorCode ?? '';
          customerName = cartTransfer.debtorName ?? '';
          break;
        case CartSystemType.purchase:
          // ระบบซื้อ: ใช้ข้อมูล creditor
          customerCode = cartTransfer.creditorCode ?? '';
          customerName = cartTransfer.creditorName ?? '';
          break;
        case CartSystemType.inventory:
          // ระบบสต๊อก: ไม่ต้องการข้อมูล customer
          customerCode = '';
          customerName = '';
          break;
      }

      screenData.custcode = customerCode;
      screenData.custnames = customerName.isNotEmpty ? [LanguageDataModel(code: 'th', name: customerName)] : [];

      // แปลงรายการสินค้าจาก cart เป็น transaction details
      screenData.details = cartTransfer.items.map((cartItem) {
        return TransactionDetailModel(
          docdatetime: DateTime.now().toUtc().toIso8601String(), // แปลงเป็น UTC ตามกฎ CLAUDE.md
          itemguid: const Uuid().v4(),
          barcode: cartItem.barcode,
          itemcode: cartItem.itemCode,
          itemnames: [LanguageDataModel(code: 'th', name: cartItem.productName)],
          unitcode: cartItem.unitCode,
          qty: cartItem.quantity,
          price: cartItem.price,
          discount: cartItem.discount.toString(),
          discountamount: cartItem.discountAmount,
          sumofcost: 0,
          sumamount: cartItem.totalPrice,
          remark: cartItem.remark ?? '',
          linenumber: 0,
          whcode: cartItem.warehouseId ?? '',
          whnames: cartItem.warehouseName != null ? [LanguageDataModel(code: 'th', name: cartItem.warehouseName!)] : [],
          locationcode: cartItem.locationId ?? '',
          locationnames: cartItem.locationName != null ? [LanguageDataModel(code: 'th', name: cartItem.locationName!)] : [],
          towhcode: '',
          towhnames: [],
          tolocationcode: '',
          tolocationnames: [],
          shelfcode: '',
          totalqty: cartItem.quantity,
          averagecost: 0,
          totalvaluevat: 0,
          sumamountexcludevat: 0,
          priceexcludevat: 0,
          standvalue: cartItem.unitStand.toDouble(),
          dividevalue: cartItem.unitDive.toDouble(),
          calcflag: calcflag,
          vattype: screenData.vattype,
          ispos: 0,
          laststatus: 0,
          itemtype: 0,
          taxtype: 0,
          inquirytype: screenData.inquirytype,
          multiunit: false,
          unitnames: [LanguageDataModel(code: 'th', name: cartItem.unitName)],
          refbarcodes: [],
          imageuri: cartItem.imageuri,
        );
      }).toList();

      // คำนวณยอดรวม
      screenData.totalvalue = cartTransfer.totalAmount;
      screenData.totalamount = cartTransfer.totalAmount;

      // คำนวณยอดรวมใหม่ตามประเภทภาษี
      screenData = calAmount.calTotalValue(screenData);

      // โหลดข้อมูลไปยังหน้าจอ
      setState(() {
        loadDataToScreen();
      });

      AppLogger.info('✅ โหลดข้อมูลจากตระกร้าสำเร็จ: ${cartTransfer.items.length} รายการ');
    } catch (e, stackTrace) {
      AppLogger.error('❌ Error populating from cart transfer: $e\n$stackTrace');
      if (mounted) {
        context.ui.showError('${global.language('error_loading_cart_data')}: $e');
      }
    }
  }

  @override
  void dispose() {
    // เคลียร์ข้อมูลที่อาจทำให้ memory leak
    docrefs.clear();
    cartList.clear();
    _cartService.disconnect();
    splitViewController.dispose();
    tabController.dispose();
    editTabController.dispose();
    docNumberController.dispose();
    docDateController.dispose();
    docTimeController.dispose();
    docRefNumberController.dispose();
    docRefDateController.dispose();
    docRefTypeController.dispose();
    docTypeController.dispose();
    vatTypeController.dispose();
    custCodeController.dispose();
    custnamesController.dispose();
    saleCodeController.dispose();
    saleNameController.dispose();
    docDiscountWordController.dispose();
    totalCostController.dispose();
    totalValueController.dispose();
    totalDiscountController.dispose();
    totalVatValueController.dispose();
    totalAfterVatController.dispose();
    totalExceptVatController.dispose();
    totalAmountController.dispose();
    cashierCodeController.dispose();
    posIdController.dispose();
    memberCodeController.dispose();
    vatRateController.dispose();
    statusController.dispose();
    descriptionController.dispose();
    payCashAmountController.dispose();
    payDeliveryCashAmountController.dispose();
    roundAmountController.dispose();
    detaildiscountformulaController.dispose();
    detailTotalDiscount.dispose();
    totalDiscountVatAmount.dispose();
    totalDiscountExceptVatAmount.dispose();
    roundAmount.dispose();
    totalAmountAfterDiscount.dispose();
    detailDiscountFormula.dispose();
    detailTotalAmount.dispose();

    for (var element in payTransferAmountController) {
      element.dispose();
    }

    for (var element in payCreditCardAmountController) {
      element.dispose();
    }

    for (var element in payChequeAmountController) {
      element.dispose();
    }

    for (var element in payCouponAmountController) {
      element.dispose();
    }

    for (var element in payQrAmountController) {
      element.dispose();
    }

    for (var element in creditCardDateController) {
      element.dispose();
    }

    for (var element in transferDateController) {
      element.dispose();
    }

    for (var element in chequeDateController) {
      element.dispose();
    }

    for (var element in chequeDueDateDateController) {
      element.dispose();
    }

    for (var element in couponDateController) {
      element.dispose();
    }

    for (var element in qrDateController) {
      element.dispose();
    }

    super.dispose();
  }

  void headerTableDetail() {
    if (widget.type == global.TransactionTypeEnum.stocktransfer) {
      headers = [
        global.DataTableHeader(code: "delete", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 10, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 20),
        global.DataTableHeader(code: "item_code", label: global.language('item_code'), width: 15), // เพิ่มคอลัมน์ itemcode
        global.DataTableHeader(code: "product_name", label: global.language('product_name'), width: 40),
        global.DataTableHeader(code: "product_unit", label: global.language('product_unit'), width: 10),
      ];

      headers.add(global.DataTableHeader(code: "product_ware_house", label: global.language('product_ware_house'), width: 10));
      headers.add(global.DataTableHeader(code: "product_location", label: global.language('product_location'), width: 10));
      headers.add(global.DataTableHeader(code: "product_to_ware_house", label: global.language('product_to_ware_house'), width: 10));
      headers.add(global.DataTableHeader(code: "product_to_location", label: global.language('product_to_location'), width: 10));

      headers.addAll([global.DataTableHeader(code: "product_qty", label: global.language('product_qty'), width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right)]);
    } else if (widget.type == global.TransactionTypeEnum.stockreceiveproduct) {
      headers = [
        global.DataTableHeader(code: "delete", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 10, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 20),
        global.DataTableHeader(code: "item_code", label: global.language('item_code'), width: 15), // เพิ่มคอลัมน์ itemcode
        global.DataTableHeader(code: "product_name", label: global.language('product_name'), width: 40),
        global.DataTableHeader(code: "product_unit", label: global.language('product_unit'), width: 10),
      ];

      headers.add(global.DataTableHeader(code: "product_ware_house", label: global.language('product_ware_house'), width: 10));
      headers.add(global.DataTableHeader(code: "product_location", label: global.language('product_location'), width: 10));
      headers.addAll([
        global.DataTableHeader(code: "product_qty", label: global.language('product_qty'), width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right),
        global.DataTableHeader(code: "product_price", label: global.language('product_cost'), width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right),
        global.DataTableHeader(code: "product_amount", label: global.language('product_amount'), alignment: Alignment.centerRight, width: 30, textAlign: TextAlign.right),
      ]);
    } else if (widget.type == global.TransactionTypeEnum.stockpickupproduct || widget.type == global.TransactionTypeEnum.stockreturnproduct || widget.type == global.TransactionTypeEnum.adjust) {
      headers = [
        global.DataTableHeader(code: "delete", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 10, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 20),
        global.DataTableHeader(code: "item_code", label: global.language('item_code'), width: 15), // เพิ่มคอลัมน์ itemcode
        global.DataTableHeader(code: "product_name", label: global.language('product_name'), width: 40),
        global.DataTableHeader(code: "product_unit", label: global.language('product_unit'), width: 10),
      ];

      headers.add(global.DataTableHeader(code: "product_ware_house", label: global.language('product_ware_house'), width: 10));
      headers.add(global.DataTableHeader(code: "product_location", label: global.language('product_location'), width: 10));
      headers.add(
        global.DataTableHeader(
          code: _getColumnCodeByTransFlag(screenData.transflag),
          label: global.language(_getColumnCodeByTransFlag(screenData.transflag)),
          width: 10,
          alignment: Alignment.centerRight,
          textAlign: TextAlign.right,
        ),
      );
    } else if (widget.type == global.TransactionTypeEnum.purchase ||
        widget.type == global.TransactionTypeEnum.purchasereturn ||
        widget.type == global.TransactionTypeEnum.purchasepartial ||
        widget.type == global.TransactionTypeEnum.accrualreceive ||
        widget.type == global.TransactionTypeEnum.sale ||
        widget.type == global.TransactionTypeEnum.salereturn) {
      headers = [
        global.DataTableHeader(code: "delete", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 10, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 15),
        global.DataTableHeader(code: "item_code", label: global.language('item_code'), width: 12), // เพิ่มคอลัมน์ itemcode
        global.DataTableHeader(code: "product_name", label: global.language('product_name'), width: 25),
        global.DataTableHeader(code: "docref", label: global.language('doc_reference'), width: 12),
        global.DataTableHeader(code: "docrefdatetime", label: global.language('reference_date'), width: 10),
        // global.DataTableHeader(code: "eventqty", label: "จำนวนเหตุการณ์", width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right),
      ];

      headers.add(global.DataTableHeader(code: "product_ware_house", label: global.language('product_ware_house'), width: 10));
      headers.add(global.DataTableHeader(code: "product_location", label: global.language('product_location'), width: 8));
      headers.addAll([
        global.DataTableHeader(code: "product_unit", label: global.language('product_unit'), width: 8),
        global.DataTableHeader(code: "product_qty", label: global.language('product_qty'), width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right),
        global.DataTableHeader(code: "product_price", label: global.language('product_price'), alignment: Alignment.centerRight, width: 12, textAlign: TextAlign.right),
        global.DataTableHeader(code: "product_discount", label: global.language('product_discount'), width: 10, alignment: Alignment.centerRight, textAlign: TextAlign.right),
        global.DataTableHeader(code: "product_amount", label: global.language('product_amount'), alignment: Alignment.centerRight, width: 12, textAlign: TextAlign.right),
      ]);
    } else {
      headers = [
        global.DataTableHeader(code: "delete", label: "", width: 5, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "line_number", label: global.language('line_number'), width: 10, textAlign: TextAlign.center, alignment: Alignment.center),
        global.DataTableHeader(code: "barcode", label: global.language('barcode'), width: 20),
        global.DataTableHeader(code: "item_code", label: global.language('item_code'), width: 15), // เพิ่มคอลัมน์ itemcode
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
  }

  String _getColumnCodeByTransFlag(int transflag) {
    /// 66 = ปรับปรุง stock เพิ่ม
    /// 68 = ปรับปรุง stock ลด
    /// 866 = ปรับปรุงมูลค่าเพิ่ม
    /// 868 = ปรับปรุงมูลค่าลด
    /// 966 = ปรับปรุงต้นทุน
    if (transflag == 66 || transflag == 68) {
      return "product_qty";
    } else if (transflag == 866 || transflag == 868) {
      return "product_amount";
    } else if (transflag == 966) {
      return "product_price_adjust";
    } else {
      return "product_qty"; // default
    }
  }

  // Helper function สำหรับ parse date อย่างปลอดภัย
  DateTime _safeParseDatetime(String? dateString) {
    if (dateString == null || dateString.isEmpty) {
      return DateTime.now();
    }
    try {
      return DateTime.parse(dateString);
    } catch (e) {
      return DateTime.now();
    }
  }

  void loadDataToScreen() {
    headerTableDetail();
    DateTime docDateTimeFormat = _safeParseDatetime(screenData.docdatetime);
    DateTime docRefDateTimeFormat = _safeParseDatetime(screenData.docrefdate);
    DateTime taxDocDateTimeFormat = _safeParseDatetime(screenData.taxdocdate);

    docNumberController.text = screenData.docno;
    docDateController.text = DateFormat('dd/MM/yyyy').format(docDateTimeFormat.toLocal());
    docTimeController.text = DateFormat('HH:mm').format(docDateTimeFormat.toLocal());
    docRefNumberController.text = screenData.docrefno;
    docRefDateController.text = DateFormat('dd/MM/yyyy').format(docRefDateTimeFormat.toLocal());
    taxDocDateController.text = DateFormat('dd/MM/yyyy').format(taxDocDateTimeFormat.toLocal());
    taxDocNoController.text = screenData.taxdocno;
    docRefTypeController.text = screenData.docreftype.toString();
    docTypeController.text = screenData.doctype.toString();
    vatTypeController.text = screenData.vattype.toString();

    /// แก้ไขปัญหา ข้อมูลเดิม ประเภทภาษีผิด ต้องเป็น 3
    if (widget.type == global.TransactionTypeEnum.stockpickupproduct) {
      screenData.vattype = 3;
      vatTypeController.text = "3";
    } else if (widget.type == global.TransactionTypeEnum.stockreceiveproduct) {
      screenData.vattype = 3;
      vatTypeController.text = "3";
    } else if (widget.type == global.TransactionTypeEnum.stockreturnproduct) {
      screenData.vattype = 3;
      vatTypeController.text = "3";
    } else if (widget.type == global.TransactionTypeEnum.stocktransfer) {
      screenData.vattype = 3;
      vatTypeController.text = "3";
    } else if (widget.type == global.TransactionTypeEnum.adjust) {
      screenData.vattype = 3;
      vatTypeController.text = "3";
    }
    custCodeController.text = screenData.custcode;
    custnamesController.text = global.activeLangName(screenData.custnames ?? []);
    saleCodeController.text = screenData.salecode;
    saleNameController.text = screenData.salename;
    docDiscountWordController.text = screenData.discountword;
    totalCostController.text = screenData.totalcost.toString();
    totalValueController.text = screenData.totalvalue.toString();
    totalDiscountController.text = screenData.totaldiscount.toString();
    totalVatValueController.text = screenData.totalvatvalue.toString();
    totalAfterVatController.text = screenData.totalaftervat.toString();
    totalExceptVatController.text = screenData.totalexceptvat.toString();
    totalAmountController.text = screenData.totalamount.toString();

    totalDiscountVatAmount.text = screenData.totaldiscountvatamount.toString();
    totalDiscountExceptVatAmount.text = screenData.totaldiscountexceptvatamount.toString();
    totalAmountAfterDiscount.text = screenData.totalamountafterdiscount.toString();
    detailDiscountFormula.text = screenData.detaildiscountformula.toString();
    detailTotalAmount.text = screenData.detailtotalamount.toString();

    cashierCodeController.text = screenData.cashiercode;
    posIdController.text = screenData.posid;
    memberCodeController.text = screenData.membercode;
    vatRateController.text = screenData.vatrate.toString();
    statusController.text = screenData.status.toString();
    descriptionController.text = screenData.description.toString();

    transportAmountController.text = NumberFormat('#,##0.00', 'en_US').format(screenData.transportamount);

    if (global.profileData.yeartype == "buddhist") {
      docDateController.text = global.dateTimeBuddhist(docDateTimeFormat, format: global.DateTimeFormatEnum.dateDay);
      docRefDateController.text = global.dateTimeBuddhist(docRefDateTimeFormat, format: global.DateTimeFormatEnum.dateDay);
      taxDocDateController.text = global.dateTimeBuddhist(taxDocDateTimeFormat, format: global.DateTimeFormatEnum.dateDay);
    } else {
      docDateController.text = DateFormat('dd/MM/yyyy').format(docDateTimeFormat);
      docRefDateController.text = DateFormat('dd/MM/yyyy').format(docRefDateTimeFormat);
      taxDocDateController.text = DateFormat('dd/MM/yyyy').format(taxDocDateTimeFormat);
    }

    double totalPayCreditCard = 0;
    double totalPayTranfer = 0;
    double totalPayCheque = 0;
    double totalPayCoupon = 0;
    double totalPayQr = 0;

    payTransfer = [];
    payCreditCard = [];
    payCheque = [];
    payCoupon = [];
    payQr = [];

    /// model ใหม่จาก pos
    couPons = [];

    creditCardDateController = [];
    transferDateController = [];
    chequeDateController = [];
    chequeDueDateDateController = [];
    couponDateController = [];
    qrDateController = [];

    payTransferAmountController = [];
    payCreditCardAmountController = [];
    payChequeAmountController = [];
    payCouponAmountController = [];
    payQrAmountController = [];

    couPons = screenData.coupons ?? [];

    // Handle null paymentStructs สำหรับเอกสารเก่าที่ไม่มีข้อมูลการชำระเงิน
    final billPayList = screenData.paymentStructs ?? [];
    for (int i = 0; i < billPayList.length; i++) {
      /// บัตรเครดิต
      if (billPayList[i].trans_flag == 1) {
        payCreditCard.add(billPayList[i]);
        totalPayCreditCard += billPayList[i].amount ?? 0;
        creditCardDateController.add(TextEditingController());
        payCreditCardAmountController.add(TextEditingController());

        for (int f = 0; f < creditCardDateController.length; f++) {
          if (global.profileData.yeartype == "buddhist") {
            creditCardDateController[f].text = global.dateTimeBuddhist(billPayList[i].doc_date_time ?? DateTime.now(), format: global.DateTimeFormatEnum.dateDay);
          } else {
            creditCardDateController[f].text = DateFormat('dd/MM/yyyy').format(billPayList[i].doc_date_time ?? DateTime.now());
          }
          payCreditCardAmountController[f].text = (billPayList[i].amount ?? 0).toString();
        }
      }

      /// เงินโอน
      if (billPayList[i].trans_flag == 2) {
        payTransfer.add(billPayList[i]);
        totalPayTranfer += billPayList[i].amount ?? 0;
        transferDateController.add(TextEditingController());
        payTransferAmountController.add(TextEditingController());

        for (int f = 0; f < transferDateController.length; f++) {
          if (global.profileData.yeartype == "buddhist") {
            transferDateController[f].text = global.dateTimeBuddhist(billPayList[i].doc_date_time ?? DateTime.now(), format: global.DateTimeFormatEnum.dateDay);
          } else {
            transferDateController[f].text = DateFormat('dd/MM/yyyy').format(billPayList[i].doc_date_time ?? DateTime.now());
          }
          payTransferAmountController[f].text = (billPayList[i].amount ?? 0).toString();
        }
      }

      /// เช็ค
      if (billPayList[i].trans_flag == 3) {
        payCheque.add(billPayList[i]);
        totalPayCheque += billPayList[i].amount ?? 0;
        chequeDateController.add(TextEditingController());
        chequeDueDateDateController.add(TextEditingController());
        payChequeAmountController.add(TextEditingController());

        for (int f = 0; f < transferDateController.length; f++) {
          if (global.profileData.yeartype == "buddhist") {
            chequeDateController[f].text = global.dateTimeBuddhist(billPayList[i].doc_date_time ?? DateTime.now(), format: global.DateTimeFormatEnum.dateDay);
            chequeDueDateDateController[f].text = global.dateTimeBuddhist(billPayList[i].due_date ?? DateTime.now(), format: global.DateTimeFormatEnum.dateDay);
          } else {
            chequeDateController[f].text = DateFormat('dd/MM/yyyy').format(billPayList[i].doc_date_time ?? DateTime.now());
            chequeDueDateDateController[f].text = DateFormat('dd/MM/yyyy').format(billPayList[i].due_date ?? DateTime.now());
          }
          payChequeAmountController[f].text = (billPayList[i].amount ?? 0).toString();
        }
      }

      /// คูปอง
      if (billPayList[i].trans_flag == 4) {
        payCoupon.add(billPayList[i]);
        totalPayCoupon += billPayList[i].amount ?? 0;
        couponDateController.add(TextEditingController());
        payCouponAmountController.add(TextEditingController());

        for (int f = 0; f < couponDateController.length; f++) {
          if (global.profileData.yeartype == "buddhist") {
            couponDateController[f].text = global.dateTimeBuddhist(billPayList[i].doc_date_time ?? DateTime.now(), format: global.DateTimeFormatEnum.dateDay);
          } else {
            couponDateController[f].text = DateFormat('dd/MM/yyyy').format(billPayList[i].doc_date_time ?? DateTime.now());
          }
          payCouponAmountController[f].text = (billPayList[i].amount ?? 0).toString();
        }
      }

      /// QR
      if (billPayList[i].trans_flag == 5) {
        payQr.add(billPayList[i]);
        totalPayQr += billPayList[i].amount ?? 0;
        qrDateController.add(TextEditingController());
        payQrAmountController.add(TextEditingController());

        for (int f = 0; f < qrDateController.length; f++) {
          if (global.profileData.yeartype == "buddhist") {
            qrDateController[f].text = global.dateTimeBuddhist(billPayList[i].doc_date_time ?? DateTime.now(), format: global.DateTimeFormatEnum.dateDay);
          } else {
            qrDateController[f].text = DateFormat('dd/MM/yyyy').format(billPayList[i].doc_date_time ?? DateTime.now());
          }
          payQrAmountController[f].text = (billPayList[i].amount ?? 0).toString();
        }
      }
    }

    screenData.sumcreditcard = totalPayCreditCard;
    screenData.summoneytransfer = totalPayTranfer;
    screenData.sumcheque = totalPayCheque;
    screenData.sumcoupon = totalPayCoupon;
    screenData.sumqrcode = totalPayQr;
    // payCashAmountController.text = (screenData.paycashamount! - screenData.paycashchange!).toString();
    payCashAmountController.text = NumberFormat('#,##0.00', 'en_US').format((screenData.paycashamount!));
    payDeliveryCashAmountController.text = NumberFormat('#,##0.00', 'en_US').format(screenData.deliveryamount);
    roundAmountController.text = (screenData.roundamount != 0) ? (screenData.roundamount!).toString() : "0";
    detaildiscountformulaController.text = (screenData.detaildiscountformula!.isNotEmpty) ? screenData.detaildiscountformula.toString() : "0";
    discountWordController.text = (screenData.discountword.isNotEmpty) ? screenData.discountword.toString() : "0";
    _transactionCalculator.calPayTotal();
  }

  /// Helper method สำหรับดึงชื่อ collection ตามประเภทเอกสาร (สำหรับ PDF API)
  String _getCollectionNameForType(global.TransactionTypeEnum type) {
    switch (type) {
      case global.TransactionTypeEnum.sale:
        return "transactionSale";
      case global.TransactionTypeEnum.salereturn:
        return "transactionSaleReturn";
      case global.TransactionTypeEnum.saleorder:
        return "transactionSaleOrder";
      case global.TransactionTypeEnum.purchase:
        return "transactionPurchase";
      case global.TransactionTypeEnum.purchasereturn:
        return "transactionPurchaseReturn";
      case global.TransactionTypeEnum.purchaseorder:
        return "transactionPurchaseOrder";
      case global.TransactionTypeEnum.purchasepartial:
        return "transactionPurchasePartial";
      case global.TransactionTypeEnum.stocktransfer:
        return "transactionStockTransfer";
      case global.TransactionTypeEnum.stockreceiveproduct:
        return "transactionStockReceive";
      case global.TransactionTypeEnum.stockpickupproduct:
        return "transactionStockPickup";
      case global.TransactionTypeEnum.stockreturnproduct:
        return "transactionStockReturn";
      case global.TransactionTypeEnum.adjust:
        return "transactionAdjust";
      case global.TransactionTypeEnum.advancePayment:
        return "transactionAdvancePayment";
      case global.TransactionTypeEnum.accrualreceive:
        return "transactionAccrualReceive";
      default:
        return "";
    }
  }

  /// Helper method สำหรับดึงชื่อเอกสารตามประเภท (สำหรับ PDF title)
  String _getDocumentTitle(global.TransactionTypeEnum type) {
    switch (type) {
      case global.TransactionTypeEnum.sale:
        return global.language("sale_invoice");
      case global.TransactionTypeEnum.salereturn:
        return global.language("sale_return");
      case global.TransactionTypeEnum.saleorder:
        return global.language("sale_order");
      case global.TransactionTypeEnum.purchase:
        return global.language("purchase_invoice");
      case global.TransactionTypeEnum.purchasereturn:
        return global.language("purchase_return");
      case global.TransactionTypeEnum.purchaseorder:
        return global.language("purchase_order");
      case global.TransactionTypeEnum.purchasepartial:
        return global.language("purchase_partial");
      case global.TransactionTypeEnum.stocktransfer:
        return global.language("stock_transfer");
      case global.TransactionTypeEnum.stockreceiveproduct:
        return global.language("stock_receive");
      case global.TransactionTypeEnum.stockpickupproduct:
        return global.language("stock_pickup");
      case global.TransactionTypeEnum.stockreturnproduct:
        return global.language("stock_return");
      case global.TransactionTypeEnum.adjust:
        return global.language("stock_adjust");
      case global.TransactionTypeEnum.advancePayment:
        return global.language("advance_payment");
      case global.TransactionTypeEnum.accrualreceive:
        return global.language("accrual_receive");
      default:
        return global.language("document");
    }
  }

  /// ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ
  /// - ถ้า auto_approved → พิมพ์อัตโนมัติทันที
  /// - ถ้าไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox จาก SharedPreferences)
  Future<void> _submitPOAndHandlePrint(String docNo, global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [_submitPOAndHandlePrint] START - docNo: $docNo');

    // ส่ง PO เข้าระบบอนุมัติและรอผลลัพธ์
    final isAutoApproved = await _submitPOForApproval();
    AppLogger.debug('🖨️ [_submitPOAndHandlePrint] isAutoApproved: $isAutoApproved');

    if (isAutoApproved && mounted) {
      // กรณี auto_approved → พิมพ์อัตโนมัติทันที
      AppLogger.debug('🖨️ [_submitPOAndHandlePrint] Auto-approved, opening PDF...');

      // Clear SharedPreferences flag (ถ้ามี)
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(_printAfterSaveKey, false);

      // แสดงข้อความแจ้งเตือนสำเร็จ
      context.ui.showSuccess('${global.language("save_success")} - ${global.language("docno")}: $docNo');

      // เปิด PDF จาก backend (พิมพ์อัตโนมัติ)
      _openPdfAfterSave(docNo, transType);
    } else {
      // กรณีไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox)
      _handleSaveSuccessWithPrintCheck(docNo, transType);
    }
  }

  /// ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ (สำหรับ Update)
  /// - ถ้า auto_approved → พิมพ์อัตโนมัติทันที
  /// - ถ้าไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox จาก SharedPreferences)
  Future<void> _submitPOAndHandleUpdatePrint(global.TransactionTypeEnum transType) async {
    final docNo = screenData.docno;
    AppLogger.debug('🖨️ [_submitPOAndHandleUpdatePrint] START - docNo: $docNo');

    // ส่ง PO เข้าระบบอนุมัติและรอผลลัพธ์
    final isAutoApproved = await _submitPOForApproval();
    AppLogger.debug('🖨️ [_submitPOAndHandleUpdatePrint] isAutoApproved: $isAutoApproved');

    if (isAutoApproved && mounted) {
      // กรณี auto_approved → พิมพ์อัตโนมัติทันที
      AppLogger.debug('🖨️ [_submitPOAndHandleUpdatePrint] Auto-approved, opening PDF...');

      // Clear SharedPreferences flag (ถ้ามี)
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(_printAfterSaveKey, false);

      // แสดงข้อความแจ้งเตือนสำเร็จ
      if (mounted) {
        context.ui.showSuccess('${global.language("update_success")} - ${global.language("docno")}: $docNo');
      }

      // เปิด PDF จาก backend (พิมพ์อัตโนมัติ)
      _openPdfAfterSave(docNo, transType);
    } else {
      // กรณีไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox)
      _handleUpdateSuccessWithPrintCheck(transType);
    }
  }

  /// จัดการหลังบันทึกสำเร็จ - อ่านค่าจาก SharedPreferences และตัดสินใจว่าจะพิมพ์หรือไม่
  Future<void> _handleSaveSuccessWithPrintCheck(String docNo, global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [_handleSaveSuccessWithPrintCheck] START - docNo: $docNo');

    // อ่านค่าจาก SharedPreferences (ป้องกัน Hot Reload reset)
    final prefs = await SharedPreferences.getInstance();
    final printAfterSave = prefs.getBool(_printAfterSaveKey) ?? false;
    AppLogger.debug('🖨️ [_handleSaveSuccessWithPrintCheck] printAfterSave from prefs: $printAfterSave');

    // reset flag หลังจากอ่านค่าแล้ว
    await prefs.setBool(_printAfterSaveKey, false);

    if (printAfterSave && mounted) {
      AppLogger.debug('🖨️ [_handleSaveSuccessWithPrintCheck] จะเปิด PDF...');

      // แสดงข้อความแจ้งเตือนสำเร็จ
      context.ui.showSuccess('${global.language("save_success")} - ${global.language("docno")}: $docNo');

      // เปิด PDF จาก backend
      _openPdfAfterSave(docNo, transType);
    } else {
      AppLogger.debug('🖨️ [_handleSaveSuccessWithPrintCheck] แสดง dialog ปกติ');
      // แสดง dialog ปกติ (กรณีไม่เลือกพิมพ์)
      if (mounted) {
        showSaveDocNoDialog(context, docNo);
      }
    }
  }

  /// จัดการหลัง Update สำเร็จ - อ่านค่าจาก SharedPreferences และตัดสินใจว่าจะพิมพ์หรือไม่
  Future<void> _handleUpdateSuccessWithPrintCheck(global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [_handleUpdateSuccessWithPrintCheck] START');

    // อ่านค่าจาก SharedPreferences (ป้องกัน Hot Reload reset)
    final prefs = await SharedPreferences.getInstance();
    final printAfterSave = prefs.getBool(_printAfterSaveKey) ?? false;
    AppLogger.debug('🖨️ [_handleUpdateSuccessWithPrintCheck] printAfterSave from prefs: $printAfterSave');

    // reset flag หลังจากอ่านค่าแล้ว
    await prefs.setBool(_printAfterSaveKey, false);

    if (printAfterSave && mounted) {
      AppLogger.debug('🖨️ [_handleUpdateSuccessWithPrintCheck] จะเปิด PDF...');
      final docNo = screenData.docno;

      // แสดงข้อความแจ้งเตือนสำเร็จ
      context.ui.showSuccess('${global.language("edit_success")} - ${global.language("docno")}: $docNo');

      // เปิด PDF จาก backend
      _openPdfAfterSave(docNo, transType);
    } else {
      AppLogger.debug('🖨️ [_handleUpdateSuccessWithPrintCheck] ไม่พิมพ์ - clear screen');
      // ไม่พิมพ์ ให้ clear screen ปกติ
      if (mounted) {
        clearScreenData();
        context.ui.showInfo(global.language("edit_success"));
      }
    }
  }

  /// เปิด PDF หลังจากบันทึกเอกสารสำเร็จ (เรียกจาก BlocListener)
  Future<void> _openPdfAfterSave(String docNo, global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [_openPdfAfterSave] START - docNo: $docNo, transType: $transType');

    if (!mounted) {
      AppLogger.debug('🖨️ [_openPdfAfterSave] Widget not mounted, returning');
      return;
    }

    PdfService pdfService = PdfService();
    String collection = _getCollectionNameForType(transType);
    AppLogger.debug('🖨️ [_openPdfAfterSave] collection: $collection');

    if (collection.isNotEmpty) {
      AppLogger.debug('🖨️ [_openPdfAfterSave] Calling generateAndOpenPdf...');
      try {
        await pdfService.generateAndOpenPdf(collection: collection, docNo: docNo, title: _getDocumentTitle(transType), context: context);
        AppLogger.debug('🖨️ [_openPdfAfterSave] generateAndOpenPdf completed');
      } catch (e) {
        AppLogger.error('🖨️ [_openPdfAfterSave] ERROR: $e');
      }
    } else {
      AppLogger.debug('🖨️ [_openPdfAfterSave] collection is empty, skipping PDF');
    }

    // หลังจากปิด PDF แล้ว ให้ clear หน้าจอ
    if (mounted) {
      AppLogger.debug('🖨️ [_openPdfAfterSave] Clearing screen data...');
      clearScreenData();
      editTabController.animateTo(0);
    }
    AppLogger.debug('🖨️ [_openPdfAfterSave] END');
  }

  /// แสดง dialog เมื่อ docno ซ้ำ และให้ผู้ใช้เลือกสร้าง docno ใหม่แล้วบันทึกอีกครั้ง
  void _showDuplicateDocNoDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return AlertDialog(
          title: Row(
            children: [
              Icon(Icons.warning_amber_rounded, color: global.theme.warningHighlightTextColor, size: 28),
              SizedBox(width: 8),
              Text(global.language('duplicate_docno')),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('${global.language('docno')} "${screenData.docno}" ${global.language('already_exists_in_system')}'),
              SizedBox(height: 12),
              Text(global.language('create_new_docno_confirm')),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.of(dialogContext).pop();
              },
              child: Text(global.language('cancel')),
            ),
            ElevatedButton(
              onPressed: () {
                Navigator.of(dialogContext).pop();
                _regenerateDocNoAndSave();
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.infoHighlightTextColor,
                foregroundColor: global.theme.onPrimaryColor,
              ),
              child: Text(global.language('create_new_docno_and_save')),
            ),
          ],
        );
      },
    );
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
    if (widget.type == global.TransactionTypeEnum.purchase) {
      transflag = 12;
      calcflag = 1;
      docnoFormat = "PU";
      configvattype = global.config.vattypepurchase;
      configinquirytype = global.config.inquirytypepurchase;
    } else if (widget.type == global.TransactionTypeEnum.purchasereturn) {
      transflag = 16;
      calcflag = -1;
      docnoFormat = "PT";
      configvattype = global.config.vattypepurchase;
    } else if (widget.type == global.TransactionTypeEnum.sale) {
      transflag = 44;
      calcflag = -1;
      docnoFormat = "SI";
      configvattype = global.config.vattypesale;
      configinquirytype = global.config.inquirytypesale;
    } else if (widget.type == global.TransactionTypeEnum.salereturn) {
      transflag = 48;
      calcflag = 1;
      docnoFormat = "ST";
      configvattype = global.config.vattypesale;
    } else if (widget.type == global.TransactionTypeEnum.stockpickupproduct) {
      transflag = 56;
      calcflag = -1;
      docnoFormat = "IO";
      configvattype = 3;
    } else if (widget.type == global.TransactionTypeEnum.stockreceiveproduct) {
      transflag = 60;
      calcflag = 1;
      docnoFormat = "IF";
      configvattype = 3;
    } else if (widget.type == global.TransactionTypeEnum.stockreturnproduct) {
      transflag = 58;
      calcflag = 1;
      docnoFormat = "IR";
      configvattype = 3;
    } else if (widget.type == global.TransactionTypeEnum.stocktransfer) {
      transflag = 72;
      calcflag = -1;
      docnoFormat = "IF";
      configvattype = 3;
    } else if (widget.type == global.TransactionTypeEnum.adjust) {
      transflag = 66;
      calcflag = 1;
      docnoFormat = "AJ";
      configvattype = 3;
    } else if (widget.type == global.TransactionTypeEnum.quotation) {
      transflag = 30;
      calcflag = -1;
      docnoFormat = "QT";
      configvattype = global.config.vattypesale;
    } else if (widget.type == global.TransactionTypeEnum.purchaseorder) {
      transflag = 6;
      calcflag = 1;
      docnoFormat = "PO";
      configvattype = global.config.vattypepurchase;
      configinquirytype = global.config.inquirytypepurchase;
    } else if (widget.type == global.TransactionTypeEnum.saleorder) {
      transflag = 36;
      calcflag = -1;
      docnoFormat = "SO";
      configvattype = global.config.vattypesale;
      configinquirytype = global.config.inquirytypesale;
    } else if (widget.type == global.TransactionTypeEnum.purchasepartial) {
      transflag = 310;
      calcflag = 1;
      docnoFormat = "PI";
      configvattype = global.config.vattypesale;
      configinquirytype = global.config.inquirytypesale;
    } else if (widget.type == global.TransactionTypeEnum.accrualreceive) {
      transflag = 315;
      calcflag = 1;
      docnoFormat = "PIU";
      configvattype = global.config.vattypepurchase;
    } else if (widget.type == global.TransactionTypeEnum.advancePayment) {
      /// จ่ายเงินล่วงหน้า
      transflag = 10;
      calcflag = -1;
      docnoFormat = "PD";
      configinquirytype = 1;
      configvattype = 3;
      global.config.vatrate = 0;
    } else if (widget.type == global.TransactionTypeEnum.advancePaymentRefund) {
      /// รับคืนเงินล่วงหน้า
      transflag = 20;
      calcflag = 1;
      docnoFormat = "PDR";
      configinquirytype = 1;
      configvattype = 3;
      global.config.vatrate = 0;
    } else if (widget.type == global.TransactionTypeEnum.deposit) {
      /// จ่ายเงินมัดจำ
      transflag = 11;
      calcflag = -1;
      docnoFormat = "PC";
      configinquirytype = 1;
      configvattype = 0;
    } else if (widget.type == global.TransactionTypeEnum.depositRefund) {
      /// รับคืนเงินมัดจำ
      transflag = 25;
      calcflag = 1;
      docnoFormat = "PCR";
      configinquirytype = 3;
      configvattype = 0;
      global.config.vatrate = 7;
    } else if (widget.type == global.TransactionTypeEnum.paidAdvance) {
      /// รับเงินล่วงหน้า
      transflag = 40;
      calcflag = 1;
      docnoFormat = "SD";
      configinquirytype = 1;

      /// ประเภทภาษี 0 = ราคาไม่รวมภาษี , 1 = ราคารวมภาษี , 2 = ภาษีอัตราศูนย์ , 3 = ไม่กระทบภาษี
      configvattype = 3;
      global.config.vatrate = 0;
    } else if (widget.type == global.TransactionTypeEnum.paidAdvanceRefund) {
      /// คืนเงินล่วงหน้า
      transflag = 42;
      calcflag = -1;
      docnoFormat = "SDR";
      configinquirytype = 3;

      /// ประเภทภาษี 0 = ราคาไม่รวมภาษี , 1 = ราคารวมภาษี , 2 = ภาษีอัตราศูนย์ , 3 = ไม่กระทบภาษี
      configvattype = 3;
      global.config.vatrate = 0;
    } else if (widget.type == global.TransactionTypeEnum.receiveDeposit) {
      /// รับเงินมัดจำ
      transflag = 110;
      calcflag = 1;
      docnoFormat = "SRV";
      configinquirytype = 1;

      /// ประเภทภาษี 0 = ราคาไม่รวมภาษี , 1 = ราคารวมภาษี , 2 = ภาษีอัตราศูนย์ , 3 = ไม่กระทบภาษี
      configvattype = 0;
      global.config.vatrate = 7;
    } else if (widget.type == global.TransactionTypeEnum.receiveDepositRefund) {
      /// คืนเงินมัดจำ
      transflag = 112;
      calcflag = -1;
      docnoFormat = "SRT";
      configinquirytype = 3;

      /// ประเภทภาษี 0 = ราคาไม่รวมภาษี , 1 = ราคารวมภาษี , 2 = ภาษีอัตราศูนย์ , 3 = ไม่กระทบภาษี
      configvattype = 0;
      global.config.vatrate = 7;
    }

    const uuid = Uuid();
    moduleEdit = TransactionEditScreenModule.header;
    // ดึงข้อมูล timezone สำหรับส่งไป mainapi
    final timezoneInfo = global.getLocalTimezoneInfo();
    screenData = TransactionModel(
      shopid: global.apiShopCode,
      guidref: uuid.v4(),
      docno: '', // mainapi จะสร้าง docno เอง
      description: '',
      docdatetime: DateTime.now().toUtc().toIso8601String(), // แปลงเป็น UTC ตามกฎ CLAUDE.md
      docdatelocal: timezoneInfo['docdatelocal'], // วันที่ local สำหรับ docno
      doctimelocal: timezoneInfo['doctimelocal'], // เวลา local
      timezone: timezoneInfo['timezone'], // timezone ของ frontend
      docrefno: '',
      docrefdate: DateTime.now().toUtc().toIso8601String(), // แปลงเป็น UTC ตามกฎ CLAUDE.md
      docreftype: 0,
      doctype: 0,
      vattype: configvattype,
      custcode: '',
      custnames: [],
      salecode: '',
      salename: '',
      discountword: '0',
      totalcost: 0,
      totalvalue: 0,
      totaldiscount: 0,
      totalvatvalue: 0,
      totalaftervat: 0,
      totalexceptvat: 0,
      totalamount: 0,
      cashiercode: '',
      posid: '',
      membercode: '',
      vatrate: global.config.vatrate,
      status: 0,
      inquirytype: configinquirytype,
      taxdocdate: DateTime.now().toLocal().toIso8601String(),
      taxdocno: '',
      totalbeforevat: 0,
      transflag: transflag,
      details: <TransactionDetailModel>[],
      iscancel: false,
      ismanualamount: false,
      paymentdetailraw: "",
      paymentStructs: [],
      sumcreditcard: 0,
      summoneytransfer: 0,
      sumcheque: 0,
      sumcoupon: 0,
      sumqrcode: 0,
      paycashamount: 0,
      paycashchange: 0,
      roundamount: 0,
      detaildiscountformula: '0',
      totaldiscountvatamount: 0,
      totaldiscountexceptvatamount: 0,
      totalamountafterdiscount: 0,
      detailtotalamount: 0,
      branch: BranchModel(guidfixed: global.companyBranchSelectData.guidfixed, code: global.companyBranchSelectData.code, names: global.companyBranchSelectData.names),
      reftotaloriginal: 0,
      reftotalcorrect: 0,
      reftotaldiff: 0,
      isdelivery: false,
      deliveryamount: 0,
      istransport: false,
      transportcode: '',
      transportamount: 0,
      docreferences: <DocReferenceModel>[],
      isclosed: false,
      isref: false,
      islocked: false,
      detailcount: 0,
    );

    payTransfer = [];
    payCreditCard = [];
    payCheque = [];
    payCoupon = [];
    payQr = [];
    couPons = [];

    creditCardDateController = [];
    transferDateController = [];
    chequeDateController = [];
    chequeDueDateDateController = [];
    couponDateController = [];
    qrDateController = [];

    payTransferAmountController = [];
    payCreditCardAmountController = [];
    payChequeAmountController = [];
    payCouponAmountController = [];
    payQrAmountController = [];

    // รีเซ็ต customerPriceLevel เป็นค่าเริ่มต้น
    customerPriceLevel = 1;

    // รีเซ็ตประเภทการจัดซื้อ (สำหรับ PO)
    _selectedPurchaseTypeCode = null;
    _selectedPurchaseTypeName = null;

    // รีเซ็ตสถานะการอนุมัติ PO (เอกสารใหม่ยังไม่มีสถานะ)
    _poApprovalStatus = null;
    _isLoadingApprovalStatus = false;

    setState(() {
      loadDataToScreen();
      docDateTimeValidated = true;
    });

    // โหลดประเภทการจัดซื้อสำหรับ PO
    if (widget.type == global.TransactionTypeEnum.purchaseorder) {
      _loadPurchaseTypesForPO();
      _loadCurrencies(); // โหลดรายการสกุลเงิน
    }
  }

  /// โหลดรายการประเภทการจัดซื้อสำหรับ PO
  Future<void> _loadPurchaseTypesForPO() async {
    try {
      final result = await ApprovalApiService.getPurchaseTypes();
      if (result.isSuccess && mounted) {
        final currentLang = global.appConfig.getString("language") ?? 'th';
        setState(() {
          _purchaseTypes = result.items;
          // ถ้ามีประเภทการจัดซื้อ ให้เลือกตัวแรกเป็นค่าเริ่มต้น
          if (_purchaseTypes.isNotEmpty && _selectedPurchaseTypeCode == null) {
            _selectedPurchaseTypeCode = _purchaseTypes.first.code;
            _selectedPurchaseTypeName = _purchaseTypes.first.getName(currentLang);
          }
        });
        AppLogger.info('[PO] โหลดประเภทการจัดซื้อสำเร็จ: ${_purchaseTypes.length} รายการ');
      }
    } catch (e) {
      AppLogger.error('[PO] โหลดประเภทการจัดซื้อผิดพลาด: $e');
    }
  }

  /// โหลด list ประเภทการจัดซื้อ และค่าที่บันทึกไว้ (สำหรับโหลดเอกสารที่มีอยู่)
  /// NOTE: purchase_type_code และ purchase_type_names ถูกย้ายไปเก็บใน TransactionModel แล้ว
  /// ไม่ต้องเรียก API แยกอีกต่อไป - อ่านค่าจาก screenData โดยตรง
  Future<void> _loadPurchaseTypesAndExisting(String docNo) async {
    try {
      // โหลด list ประเภทการจัดซื้อก่อน
      final typesResult = await ApprovalApiService.getPurchaseTypes();
      if (typesResult.isSuccess && mounted) {
        setState(() {
          _purchaseTypes = typesResult.items;
        });
        AppLogger.info('[PO] โหลดประเภทการจัดซื้อสำเร็จ: ${_purchaseTypes.length} รายการ');
      }

      // อ่านค่าที่บันทึกไว้จาก TransactionModel โดยตรง (ไม่ต้องเรียก API แยก)
      if (mounted) {
        final existingCode = screenData.purchasetypecode;
        final existingNames = screenData.purchasetypenames;

        if (existingCode != null && existingCode.isNotEmpty) {
          // ดึงชื่อจาก purchasetypenames ตามภาษาปัจจุบัน
          String? purchaseTypeName;
          if (existingNames != null && existingNames.isNotEmpty) {
            final currentLang = global.appConfig.getString("language") ?? 'th';
            final nameData = existingNames.firstWhere(
              (n) => n.code == currentLang,
              orElse: () => existingNames.first,
            );
            purchaseTypeName = nameData.name;
          }

          setState(() {
            _selectedPurchaseTypeCode = existingCode;
            _selectedPurchaseTypeName = purchaseTypeName ?? '';
          });
          AppLogger.info('[PO] โหลดประเภทการจัดซื้อของ PO สำเร็จ: $purchaseTypeName');
        } else if (_purchaseTypes.isNotEmpty) {
          // ถ้าไม่พบค่าที่บันทึกไว้ ให้ใช้ค่าแรกใน list
          final currentLang = global.appConfig.getString("language") ?? 'th';
          setState(() {
            _selectedPurchaseTypeCode = _purchaseTypes.first.code;
            _selectedPurchaseTypeName = _purchaseTypes.first.getName(currentLang);
          });
        }
      }
    } catch (e) {
      AppLogger.error('[PO] โหลดประเภทการจัดซื้อผิดพลาด: $e');
    }
  }

  /// บันทึกประเภทการจัดซื้อของ PO ลงใน TransactionModel
  /// NOTE: purchase_type_code และ purchase_type_names ถูกย้ายไปเก็บใน TransactionModel แล้ว
  /// ค่าจะถูกบันทึกไปพร้อมกับเอกสารหลัก ไม่ต้องเรียก API แยก
  void _savePOPurchaseType() {
    if (widget.type != global.TransactionTypeEnum.purchaseorder) return;
    if (_selectedPurchaseTypeCode == null || _selectedPurchaseTypeCode!.isEmpty) return;

    // บันทึกค่าลงใน screenData (จะถูกบันทึกไปพร้อมกับเอกสารหลัก)
    screenData.purchasetypecode = _selectedPurchaseTypeCode;

    // สร้าง purchasetypenames (หลายภาษา) จากชื่อที่เลือก
    if (_selectedPurchaseTypeName != null && _selectedPurchaseTypeName!.isNotEmpty) {
      final currentLang = global.appConfig.getString("language") ?? 'th';
      screenData.purchasetypenames = [
        LanguageDataModel(code: currentLang, name: _selectedPurchaseTypeName!),
      ];
    }

    AppLogger.info('[PO] บันทึกประเภทการจัดซื้อลงใน TransactionModel: $_selectedPurchaseTypeCode');
    
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
      screenData.exchangerate = _selectedExchangeRate ?? 1.0;
    }
    
    AppLogger.info('[PO] บันทึกสกุลเงิน: doc=${screenData.docCurrency}, base=${screenData.currency}, อัตรา=${screenData.exchangerate}');
  }

  /// เปลี่ยนประเภทการจัดซื้อที่เลือก
  void _onPurchaseTypeChanged(String? code, String? name) {
    setState(() {
      _selectedPurchaseTypeCode = code;
      _selectedPurchaseTypeName = name;
    });
  }

  /// โหลดรายการสกุลเงิน (สำหรับ PO Multi-Currency)
  /// 
  /// รองรับกรณีต่างๆ:
  /// 1. เอกสารใหม่ - ใช้ Base Currency เป็นค่าเริ่มต้น
  /// 2. Legacy Document (มีแค่ currency ไม่มี docCurrency) - copy ค่า
  /// 3. Multi-Currency Document (มี docCurrency) - ใช้ค่าที่มี
  Future<void> _loadCurrencies() async {
    if (widget.type != global.TransactionTypeEnum.purchaseorder) return;

    try {
      final apiService = CurrencyApiService();
      final shopId = global.prefs.getString("shopid") ?? "";
      
      final response = await apiService.getCurrencies(
        shopId: shopId,
        page: 1,
        limit: 1000,
      );

      if (response.success && response.data != null) {
        final currencies = (response.data as List)
            .map((e) => CurrencyModel.fromJson(e))
            .where((c) => !c.isdisabled) // กรองเฉพาะสกุลเงินที่ไม่ได้ปิดการใช้งาน
            .toList();

        // ============================================
        // STEP 1: กำหนด Base Currency (สกุลเงินหลัก)
        // ============================================
        // ใช้ TransactionCurrencyUtils เพื่อความ consistency
        final baseCurrencyFromApi = TransactionCurrencyUtils.findBaseCurrency(currencies);
        
        _baseCurrency = baseCurrencyFromApi.code;
        _baseCurrencySymbol = baseCurrencyFromApi.symbol;
        
        // STEP 2: (ยกเลิกแล้ว) ไม่ migrate legacy documents — เดินหน้าอย่างเดียว
        
        // ============================================
        // STEP 3: ตรวจสอบจำนวนสกุลเงิน
        // ============================================
        if (currencies.length <= 1) {
          _handleSingleCurrency(currencies);
          return;
        }
        
        // ============================================
        // STEP 4: มีหลายสกุลเงิน - ตั้งค่าเริ่มต้น
        // ============================================
        setState(() {
          _currencies = currencies;
          _initializeCurrencySelection();
        });
        
        AppLogger.info('[PO] โหลดรายการสกุลเงินสำเร็จ: ${_currencies.length} รายการ, base=$_baseCurrency');
      }
    } catch (e) {
      AppLogger.error('[PO] โหลดรายการสกุลเงินผิดพลาด: $e');
      // ใช้ค่าเริ่มต้นแม้โหลดไม่สำเร็จ
      _baseCurrency = 'THB';
      _baseCurrencySymbol = '฿';
    }
  }

  /// จัดการกรณีมีสกุลเงินเดียว
  void _handleSingleCurrency(List<CurrencyModel> currencies) {
    // ตรวจสอบว่าเอกสารมี multi-currency อยู่แล้วหรือไม่
    if (screenData.isMultiCurrency) {
      // เอกสารเก่าที่มี multi-currency → เก็บค่าเดิมไว้
      _selectedDocCurrencyCode = screenData.docCurrency;
      _selectedDocCurrencySymbol = screenData.docCurrencySymbol;
      _selectedExchangeRate = screenData.safeExchangeRate;

      // ต้องแสดง section สกุลเงินเพื่อให้ user เห็นว่าเอกสารนี้มี multi-currency
      setState(() {
        _currencies = TransactionCurrencyUtils.createCurrencyListForExistingDoc(
          _baseCurrency!,
          _baseCurrencySymbol!,
          screenData.docCurrency!,
          screenData.docCurrencySymbol,
        );
      });

      AppLogger.info('[PO] เอกสารเก่ามี multi-currency: ${screenData.docCurrency} (rate: $_selectedExchangeRate)');
      return;
    }

    // เอกสารใหม่หรือไม่มี multi-currency → ใช้ Base Currency
    _initializeWithBaseCurrency();

    // ไม่ต้อง setState _currencies เพื่อไม่ให้แสดง section สกุลเงิน
    AppLogger.info('[PO] มีสกุลเงินแค่ ${_currencies.length} สกุล ไม่แสดง section สกุลเงิน');
  }

  /// ตั้งค่าเริ่มต้นสำหรับการเลือกสกุลเงิน
  void _initializeCurrencySelection() {
    if (screenData.docCurrency != null && screenData.docCurrency!.isNotEmpty) {
      // มี docCurrency อยู่แล้ว → ใช้ค่านั้น
      _selectedDocCurrencyCode = screenData.docCurrency;
      _selectedDocCurrencySymbol = screenData.docCurrencySymbol;
      _selectedExchangeRate = screenData.safeExchangeRate;
    } else {
      // ไม่มี → ใช้ Base Currency
      _initializeWithBaseCurrency();
    }
  }

  /// ตั้งค่าเริ่มต้นด้วย Base Currency
  void _initializeWithBaseCurrency() {
    _selectedDocCurrencyCode = _baseCurrency;
    _selectedDocCurrencySymbol = _baseCurrencySymbol;
    _selectedExchangeRate = 1.0;

    // ใช้ TransactionCurrencyUtils สำหรับ initialize
    TransactionCurrencyUtils.initializeWithBaseCurrency(
      screenData,
      _baseCurrency!,
      _baseCurrencySymbol!,
    );
  }

  /// เปลี่ยนสกุลเงินที่เลือก
  /// 
  /// รองรับการตรวจสอบและแจ้งเตือนเมื่อมีปัญหากับ exchange rate
  void _onCurrencyChanged(String? code, String? symbol, double? exchangeRate) async {
    // ถ้าเป็นสกุลเงินหลัก ใช้อัตรา 1.0
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
      
      // Recalculate totals
      _transactionCalculator.calTotalValue();
      
      AppLogger.info('[PO] เปลี่ยนสกุลเงิน: $code (base currency), อัตรา: 1.0');
      return;
    }

    // ดึงอัตราแลกเปลี่ยนล่าสุดจาก API
    double rate = exchangeRate ?? 1.0;
    bool isRateFromApi = false;
    
    if (code != null) {
      try {
        final today = DateTime.now();
        final dateStr = '${today.year}-${today.month.toString().padLeft(2, '0')}-${today.day.toString().padLeft(2, '0')}';
        final apiService = CurrencyApiService();
        final latestRate = await apiService.getLatestExchangeRate(
          currency: code,
          date: dateStr,
        );
        if (latestRate != null && latestRate.rate > 0) {
          rate = latestRate.rate;
          isRateFromApi = true;
          AppLogger.info('[PO] ดึงอัตราแลกเปลี่ยนล่าสุดจาก API: $code = $rate');
        } else {
          // ไม่มีข้อมูลจาก API
          _showExchangeRateWarning('${global.language('exchange_rate_not_found_for')} $code ${global.language('please_check_and_enter_manually')}');
        }
      } catch (e) {
        AppLogger.error('[PO] ดึงอัตราแลกเปลี่ยนล่าสุดไม่สำเร็จ: $e');
        _showExchangeRateWarning('${global.language('fetch_exchange_rate_failed')} ${global.language('please_check_and_enter_manually')}');
      }
    }

    // Validation: ใช้ TransactionCurrencyUtils
    final validation = TransactionCurrencyUtils.validateExchangeRate(
      rate,
      code ?? '',
      _baseCurrency!,
    );
    if (!validation.isValid) {
      rate = validation.correctedRate;
    }
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

    // Recalculate totals หลังเปลี่ยนสกุลเงิน
    _transactionCalculator.calTotalValue();

    AppLogger.info('[PO] เปลี่ยนสกุลเงิน: doc=$code, base=$_baseCurrency, อัตรา=$rate, fromApi=$isRateFromApi');
  }

  /// แสดง warning เมื่อมีปัญหากับ exchange rate
  /// 
  /// Deprecated: ใช้ context.ui.showWarning() แทน
  void _showExchangeRateWarning(String message) {
    context.ui.showWarning(message);
  }
  /// 
  /// Returns: Record {isValid, correctedRate, hasWarning}
  /// แก้ไขอัตราแลกเปลี่ยนเอง
  /// 
  /// มีการตรวจสอบค่าที่ผู้ใช้กรอกและแสดง warning ถ้าผิดปกติ
  void _onExchangeRateChanged(double? exchangeRate) {
    double safeRate = exchangeRate ?? 1.0;
    
    // ตรวจสอบค่าที่กรอก ใช้ TransactionCurrencyUtils
    if (_selectedDocCurrencyCode != null && _selectedDocCurrencyCode != _baseCurrency) {
      final validation = TransactionCurrencyUtils.validateExchangeRate(
        safeRate,
        _selectedDocCurrencyCode!,
        _baseCurrency!,
      );
      safeRate = validation.correctedRate;
      if (validation.hasWarning && validation.message != null) {
        context.ui.showWarning(validation.message!);
      }
    } else {
      // ถ้าเป็น base currency บังคับให้เป็น 1.0
      if (safeRate != 1.0) {
        safeRate = 1.0;
        context.ui.showWarning(global.language('base_currency_exchange_rate_must_be_one'));
      }
    }

    setState(() {
      _selectedExchangeRate = safeRate;
      screenData.exchangerate = _selectedExchangeRate;
    });

    // Recalculate totals หลังเปลี่ยนอัตราแลกเปลี่ยน
    _transactionCalculator.calTotalValue();
    
    AppLogger.info('[PO] แก้ไขอัตราแลกเปลี่ยน: $safeRate');
  }

  /// โหลดสถานะการอนุมัติของ PO
  Future<void> _loadPOApprovalStatus(String docNo) async {
    if (widget.type != global.TransactionTypeEnum.purchaseorder) return;

    AppLogger.debug('[PO] _loadPOApprovalStatus เริ่มโหลดสถานะสำหรับ docNo: $docNo');
    setState(() => _isLoadingApprovalStatus = true);

    try {
      final result = await ApprovalApiService.getPOApprovalStatus(docNo);
      AppLogger.debug('[PO] _loadPOApprovalStatus result - isSuccess: ${result.isSuccess}, isFound: ${result.isFound}');
      if (result.isSuccess && result.isFound && mounted) {
        setState(() {
          _poApprovalStatus = result.data;
          _isLoadingApprovalStatus = false;
        });
        AppLogger.info('[PO] โหลดสถานะการอนุมัติสำเร็จ: ${result.data?.status}, _poApprovalStatus set to: ${_poApprovalStatus != null}');
      } else {
        setState(() {
          _poApprovalStatus = null;
          _isLoadingApprovalStatus = false;
        });
        AppLogger.debug('[PO] ไม่พบสถานะการอนุมัติสำหรับ docNo: $docNo');
      }
    } catch (e) {
      AppLogger.error('[PO] โหลดสถานะการอนุมัติผิดพลาด: $e');
      if (mounted) {
        setState(() => _isLoadingApprovalStatus = false);
      }
    }
  }

  /// ส่ง PO เข้าระบบอนุมัติ
  /// [isModified] - true เมื่อเป็นการแก้ไขเอกสารที่เคยส่งอนุมัติแล้ว
  /// คืนค่า true ถ้าเป็น auto_approved (สามารถพิมพ์ได้ทันที)
  Future<bool> _submitPOForApproval({bool isModified = false}) async {
    if (widget.type != global.TransactionTypeEnum.purchaseorder) return false;
    if (_selectedPurchaseTypeCode == null || _selectedPurchaseTypeCode!.isEmpty) return false;

    try {
      // คำนวณยอดรวม
      final totalAmount = screenData.totalaftervat;

      // ดึงข้อมูล user ปัจจุบัน
      final userCode = global.appConfig.getString("user") ?? '';
      final userName = userCode; // ใช้ username เป็นชื่อ (สามารถปรับปรุงภายหลัง)
      const approvalLevel = 0; // ระดับสิทธิ์อนุมัติ (สามารถปรับปรุงจาก user profile ภายหลัง)

      // ถ้ามี approval status อยู่แล้ว ถือว่าเป็นการแก้ไขเอกสาร
      final effectiveIsModified = isModified || (_poApprovalStatus != null);

      AppLogger.debug('[PO] Submit approval - isModified: $isModified, hasExistingStatus: ${_poApprovalStatus != null}, effectiveIsModified: $effectiveIsModified');

      // Sync หมายเหตุจาก controller ไปยัง screenData (user อาจพิมพ์ค่าใหม่)
      screenData.description = descriptionController.text;
      AppLogger.debug('[PO] Description (หมายเหตุ): ${screenData.description}');

      // Build items list for LIFF display
      final List<Map<String, dynamic>> itemsList = [];
      if (screenData.details != null) {
        for (int i = 0; i < screenData.details!.length; i++) {
          final detail = screenData.details![i];
          itemsList.add({
            'linenumber': i + 1,
            'itemcode': detail.itemcode,
            'itemname': global.activeLangName(detail.itemnames ?? []),
            'qty': detail.qty,
            'unitname': global.activeLangName(detail.unitnames ?? []),
            'price': detail.price,
            'sumamount': detail.sumamount,
          });
        }
      }

      final result = await ApprovalApiService.submitPOApproval(
        docNo: screenData.docno ?? '',
        guidFixed: screenData.guidfixed ?? '',
        purchaseTypeCode: _selectedPurchaseTypeCode!,
        purchaseTypeName: _selectedPurchaseTypeName ?? '',
        totalAmount: totalAmount,
        actionBy: userCode,
        actionByName: userName,
        approvalLevel: approvalLevel,
        isModified: effectiveIsModified,
        // หมายเหตุจากใบสั่งซื้อ
        comment: (screenData.description?.isNotEmpty ?? false) ? screenData.description : null,
        // Transaction details for LIFF display
        docDatetime: screenData.docdatetime,
        custCode: screenData.custcode,
        custName: global.activeLangName(screenData.custnames ?? []),
        items: itemsList,
      );

      if (result.isSuccess && mounted) {
        setState(() {
          _poApprovalStatus = result.data;
        });

        // ตรวจสอบว่าเป็น auto_approved หรือไม่
        final isAutoApproved = result.data?.status == POApprovalStatus.autoApproved;

        // แสดงข้อความแจ้งเตือน
        final message = result.message ?? global.language('submit_approval_success');
        if (isAutoApproved) {
          context.ui.showSuccess(message);
        } else {
          context.ui.showInfo(message);
        }

        if (result.notificationsSent > 0) {
          AppLogger.info('[PO] ส่งแจ้งเตือนผู้อนุมัติแล้ว ${result.notificationsSent} คน');
        }

        // คืนค่า true ถ้า auto_approved (สามารถพิมพ์ได้ทันที)
        return isAutoApproved;
      } else if (mounted) {
        context.ui.showError('${global.language('submit_approval_failed')}: ${result.errorMessage}');
      }
      return false;
    } catch (e) {
      AppLogger.error('[PO] ส่งเข้าระบบอนุมัติผิดพลาด: $e');
      if (mounted) {
        context.ui.showError('${global.language('submit_approval_error')}: $e');
      }
      return false;
    }
  }

  /// Widget แสดงสถานะการอนุมัติ PO - แสดงเป็น banner ด้านบนของเอกสาร
  Widget _buildApprovalStatusBadge() {
    if (widget.type != global.TransactionTypeEnum.purchaseorder) {
      return const SizedBox.shrink();
    }

    // ตรวจสอบสถานะยกเลิกก่อน - แสดงทันทีไม่ต้องรอ approval status
    if (screenData.iscancel == true) {
      return Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        decoration: BoxDecoration(
          color: global.theme.dividerBorderColor,
          border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1)),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.block, size: 20, color: global.theme.iconSecondaryColor),
                SizedBox(width: 8),
                Text(
                  '${global.language('status')}: ${global.language('cancelled')}',
                  style: TextStyle(
                    color: global.theme.textColor,
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
              ],
            ),
            // แสดงเหตุผลการยกเลิก (ถ้ามี)
            if ((screenData.cancelreason ?? '').isNotEmpty) ...[
              SizedBox(height: 4),
              Text(
                '${global.language('cancel_reason')}: ${screenData.cancelreason}',
                style: TextStyle(
                  color: global.theme.textSecondaryColor,
                  fontSize: 12,
                ),
                textAlign: TextAlign.center,
              ),
            ],
            // แสดงข้อมูลผู้ยกเลิกและเวลา (ถ้ามี)
            if ((screenData.cancelusercode ?? '').isNotEmpty || (screenData.canceldatetime ?? '').isNotEmpty) ...[
              SizedBox(height: 2),
              Text(
                '${global.language('by')}: ${(screenData.cancelusername ?? '').isNotEmpty ? screenData.cancelusername : screenData.cancelusercode}',
                style: TextStyle(
                  color: global.theme.textSecondaryColor,
                  fontSize: 11,
                ),
              ),
            ],
          ],
        ),
      );
    }

    if (_isLoadingApprovalStatus) {
      return Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        color: global.theme.surfaceColor,
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            SizedBox(width: 8),
            Text('${global.language('loading_approval_status')}...'),
          ],
        ),
      );
    }

    if (_poApprovalStatus == null) {
      return const SizedBox.shrink();
    }

    final status = _poApprovalStatus!.status;
    Color bgColor;
    Color textColor;
    Color borderColor;
    String statusText;
    IconData icon;

    switch (status) {
      case POApprovalStatus.pending:
        bgColor = global.theme.warningHighlightColor;
        textColor = global.theme.warningHighlightTextColor;
        borderColor = global.theme.warningHighlightTextColor;
        statusText = global.language('pending_approval');
        icon = Icons.hourglass_empty;
        break;
      case POApprovalStatus.approved:
        bgColor = global.theme.positiveHighlightColor;
        textColor = global.theme.positiveHighlightTextColor;
        borderColor = global.theme.positiveHighlightTextColor;
        statusText = global.language('approved');
        icon = Icons.check_circle;
        break;
      case POApprovalStatus.rejected:
        bgColor = global.theme.negativeHighlightColor;
        textColor = global.theme.negativeHighlightTextColor;
        borderColor = global.theme.negativeHighlightTextColor;
        statusText = global.language('rejected');
        icon = Icons.cancel;
        break;
      case POApprovalStatus.autoApproved:
        bgColor = global.theme.infoHighlightColor;
        textColor = global.theme.infoHighlightTextColor;
        borderColor = global.theme.infoHighlightTextColor;
        statusText = global.language('auto_approved');
        icon = Icons.verified;
        break;
      default:
        bgColor = global.theme.surfaceColor;
        textColor = global.theme.textColor;
        borderColor = global.theme.dividerBorderColor;
        statusText = global.language('unknown_status');
        icon = Icons.help_outline;
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        color: bgColor,
        border: Border(bottom: BorderSide(color: borderColor, width: 1)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, size: 20, color: textColor),
          SizedBox(width: 8),
          Text(
            '${global.language('approval_status')}: $statusText',
            style: TextStyle(
              color: textColor,
              fontWeight: FontWeight.bold,
              fontSize: 14,
            ),
          ),
          const SizedBox(width: 12),
          // ปุ่มดูประวัติการอนุมัติ
          InkWell(
            onTap: () {
              final docNo = screenData.docno;
              if (docNo.isNotEmpty) {
                POApprovalHistoryDialog.show(context, docNo);
              }
            },
            borderRadius: BorderRadius.circular(16),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              decoration: BoxDecoration(
                color: textColor.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: textColor.withValues(alpha: 0.3)),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.history, size: 16, color: textColor),
                  SizedBox(width: 4),
                  Text(
                    global.language('view_history'),
                    style: TextStyle(
                      color: textColor,
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),
          ),
          // ปุ่มส่งคำเตือนผู้อนุมัติ (แสดงเฉพาะสถานะ pending)
          if (status == POApprovalStatus.pending) ...[
            const SizedBox(width: 8),
            InkWell(
              onTap: () async {
                final docNo = screenData.docno;
                if (docNo.isEmpty) return;

                // แสดง loading dialog พร้อมเก็บ context ของ dialog ไว้ปิด
                // ใช้ useRootNavigator: true เพื่อให้ dialog แยกจาก screen navigator
                showDialog(
                  context: context,
                  barrierDismissible: false,
                  useRootNavigator: true,
                  builder: (dialogContext) => const Center(
                    child: CircularProgressIndicator(),
                  ),
                );

                try {
                  // เรียก API ส่งคำเตือน
                  final result = await ApprovalApiService.resendApprovalNotification(docNo: docNo);

                  // ปิด loading dialog - ใช้ rootNavigator เพื่อ pop dialog ไม่ใช่ screen
                  if (mounted) {
                    Navigator.of(context, rootNavigator: true).pop();
                  }

                  // แสดง dialog ผลลัพธ์
                  if (mounted) {
                    final isSuccess = result.isSuccess && result.sent > 0;
                    await showDialog(
                      context: context,
                      useRootNavigator: true,
                      builder: (ctx) => AlertDialog(
                        title: Row(
                          children: [
                            Icon(
                              isSuccess ? Icons.check_circle : Icons.error,
                              color: isSuccess ? global.theme.positiveHighlightTextColor : global.theme.negativeHighlightTextColor,
                            ),
                            SizedBox(width: 8),
                            Text(isSuccess ? global.language('send_success') : global.language('send_failed')),
                          ],
                        ),
                        content: Column(
                          mainAxisSize: MainAxisSize.min,
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(result.message ?? result.errorMessage ?? (isSuccess ? global.language('send_notification_success') : global.language('send_notification_failed'))),
                            if (isSuccess) ...[
                              SizedBox(height: 8),
                              Text('Email: ${result.emailSent} ${global.language('times')}', style: TextStyle(fontSize: 14, color: global.theme.iconSecondaryColor)),
                              Text('LINE: ${result.lineSent} ${global.language('times')}', style: TextStyle(fontSize: 14, color: global.theme.iconSecondaryColor)),
                            ],
                          ],
                        ),
                        actions: [
                          TextButton(
                            onPressed: () => Navigator.of(ctx).pop(),
                            child: Text(global.language('ok')),
                          ),
                        ],
                      ),
                    );
                  }
                } catch (e) {
                  // ปิด loading dialog
                  if (mounted) {
                    Navigator.of(context, rootNavigator: true).pop();
                  }
                  // แสดง error dialog
                  if (mounted) {
                    await showDialog(
                      context: context,
                      useRootNavigator: true,
                      builder: (ctx) => AlertDialog(
                        title: Row(
                          children: [
                            Icon(Icons.error, color: global.theme.negativeHighlightTextColor),
                            SizedBox(width: 8),
                            Text(global.language('error_occurred')),
                          ],
                        ),
                        content: Text('$e'),
                        actions: [
                          TextButton(
                            onPressed: () => Navigator.of(ctx).pop(),
                            child: Text(global.language('ok')),
                          ),
                        ],
                      ),
                    );
                  }
                }
              },
              borderRadius: BorderRadius.circular(16),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                decoration: BoxDecoration(
                  color: global.theme.infoHighlightColor,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: global.theme.infoHighlightTextColor),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.notifications_active, size: 16, color: global.theme.infoHighlightTextColor),
                    SizedBox(width: 4),
                    Text(
                      global.language('send_reminder'),
                      style: TextStyle(
                        color: global.theme.infoHighlightTextColor,
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  // =====================================================
  // Helper methods สำหรับตรวจสอบสิทธิ์พิมพ์/แก้ไข PO ตามสถานะการอนุมัติ
  // =====================================================

  /// ตรวจสอบว่า PO สามารถพิมพ์ได้หรือไม่
  /// - อนุมัติอัตโนมัติ (auto_approved): พิมพ์ได้ทันทีหลัง save
  /// - อนุมัติแล้ว (approved): พิมพ์ได้
  /// - รออนุมัติ (pending): พิมพ์ไม่ได้
  /// - ถูกปฏิเสธ (rejected): พิมพ์ไม่ได้
  /// - ยังไม่ส่งอนุมัติ: พิมพ์ไม่ได้
  bool _canPrintPO() {
    // ถ้าไม่ใช่ PO ให้พิมพ์ได้ตามปกติ
    if (widget.type != global.TransactionTypeEnum.purchaseorder) {
      return true;
    }

    // ถ้าไม่มีสถานะการอนุมัติ = ยังไม่ส่งอนุมัติ = พิมพ์ไม่ได้
    if (_poApprovalStatus == null) {
      return false;
    }

    // พิมพ์ได้เฉพาะเมื่ออนุมัติแล้ว หรืออนุมัติอัตโนมัติ
    return _poApprovalStatus!.status == POApprovalStatus.approved ||
           _poApprovalStatus!.status == POApprovalStatus.autoApproved;
  }

  /// ตรวจสอบว่า PO สามารถแก้ไขได้หรือไม่
  /// - ถ้ามีการอนุมัติแล้ว (ไม่ว่าขั้นตอนใด) = ห้ามแก้ไข
  /// - อนุมัติอัตโนมัติ (auto_approved): ห้ามแก้ไข
  /// - อนุมัติแล้ว (approved): ห้ามแก้ไข
  /// - รออนุมัติ (pending): ห้ามแก้ไข (มีคนอนุมัติไปบางส่วนแล้ว หรือรอดำเนินการ)
  /// - ถูกปฏิเสธ (rejected): แก้ไขได้ (เพื่อแก้ไขแล้วส่งอนุมัติใหม่)
  /// - ยังไม่ส่งอนุมัติ: แก้ไขได้
  bool _canEditPO() {
    // ถ้าไม่ใช่ PO ให้แก้ไขได้ตามปกติ
    if (widget.type != global.TransactionTypeEnum.purchaseorder) {
      return true;
    }

    // ถ้าไม่มีสถานะการอนุมัติ = ยังไม่ส่งอนุมัติ = แก้ไขได้
    if (_poApprovalStatus == null) {
      return true;
    }

    // ตรวจสอบว่ามีการอนุมัติบางส่วนไปแล้วหรือไม่ (current_approved_level > 0)
    // ถ้ามี = ห้ามแก้ไข
    if (_poApprovalStatus!.currentApprovedLevel > 0) {
      return false;
    }

    // ถูกปฏิเสธ = แก้ไขได้ (เพื่อส่งอนุมัติใหม่)
    if (_poApprovalStatus!.status == POApprovalStatus.rejected) {
      return true;
    }

    // รออนุมัติ (pending) แต่ยังไม่มีใครอนุมัติ = ห้ามแก้ไข (อยู่ในระหว่างรออนุมัติ)
    if (_poApprovalStatus!.status == POApprovalStatus.pending) {
      return false;
    }

    // อนุมัติอัตโนมัติ หรือ อนุมัติแล้ว = ห้ามแก้ไข
    if (_poApprovalStatus!.status == POApprovalStatus.autoApproved ||
        _poApprovalStatus!.status == POApprovalStatus.approved) {
      return false;
    }

    // กรณีอื่นๆ (draft) = แก้ไขได้
    return true;
  }

  /// ข้อความแสดงเหตุผลที่ไม่สามารถพิมพ์ได้
  String _getPOPrintDisabledReason() {
    if (_poApprovalStatus == null) {
      return global.language('not_submitted_for_approval');
    }

    switch (_poApprovalStatus!.status) {
      case POApprovalStatus.pending:
        return '${global.language('pending_approval')} - ${global.language('cannot_print_yet')}';
      case POApprovalStatus.rejected:
        return '${global.language('rejected')} - ${global.language('cannot_print')}';
      case POApprovalStatus.draft:
        return global.language('not_submitted_for_approval');
      default:
        return '';
    }
  }

  /// ข้อความแสดงเหตุผลที่ไม่สามารถแก้ไขได้
  String _getPOEditDisabledReason() {
    if (_poApprovalStatus == null) {
      return '';
    }

    if (_poApprovalStatus!.currentApprovedLevel > 0) {
      return '${global.language('already_approved')} - ${global.language('cannot_edit_document')}';
    }

    switch (_poApprovalStatus!.status) {
      case POApprovalStatus.pending:
        return '${global.language('pending_approval')} - ${global.language('cannot_edit_document')}';
      case POApprovalStatus.approved:
        return '${global.language('approved')} - ${global.language('cannot_edit_document')}';
      case POApprovalStatus.autoApproved:
        return '${global.language('auto_approved')} - ${global.language('cannot_edit_document')}';
      default:
        return '';
    }
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
      transactionType: widget.type,
      docDateTimeValidated: docDateTimeValidated,
      fieldNameCustCode: fieldNameCustCode,
      fieldNameCustName: fieldNameCustName,
    );
  }

  double getPrice(List<PriceDataModel>? result) {
    double price = 0;

    if (widget.type == global.TransactionTypeEnum.purchaseorder ||
        widget.type == global.TransactionTypeEnum.purchasepartial ||
        widget.type == global.TransactionTypeEnum.purchasereturn ||
        widget.type == global.TransactionTypeEnum.stockreceiveproduct ||
        widget.type == global.TransactionTypeEnum.purchase ||
        widget.type == global.TransactionTypeEnum.accrualreceive) {
      price = 0;
    } else if (widget.type == global.TransactionTypeEnum.sale) {
      if (result != null && result.isNotEmpty) {
        // หาราคาตาม pricelevel ของลูกค้า - ใช้ keynumber แทน priceNumber
        PriceDataModel? targetPrice = result.where((p) => p.keynumber == customerPriceLevel).firstOrNull;

        if (targetPrice != null) {
          price = targetPrice.price;
        }

        // ถ้าราคาเป็น 0 หรือไม่มีราคาตาม pricelevel ให้ใช้ keynumber = 1
        if (price == 0) {
          PriceDataModel? defaultPrice = result.where((p) => p.keynumber == 1).firstOrNull;
          price = defaultPrice?.price ?? 0;
        }
      }
    } else if (widget.type == global.TransactionTypeEnum.saleorder) {
      if (result != null && result.isNotEmpty) {
        // หาราคาตาม pricelevel ของลูกค้า - ใช้ keynumber แทน priceNumber
        PriceDataModel? targetPrice = result.where((p) => p.keynumber == customerPriceLevel).firstOrNull;

        if (targetPrice != null) {
          price = targetPrice.price;
        }

        // ถ้าราคาเป็น 0 หรือไม่มีราคาตาม pricelevel ให้ใช้ keynumber = 1
        if (price == 0) {
          PriceDataModel? defaultPrice = result.where((p) => p.keynumber == 1).firstOrNull;
          price = defaultPrice?.price ?? 0;
        }
      }
    } else {
      if (result!.isNotEmpty) {
        // หาราคาตาม pricelevel ของลูกค้า - ใช้ keynumber แทน priceNumber
        PriceDataModel? targetPrice = result.where((p) => p.keynumber == customerPriceLevel).firstOrNull;

        if (targetPrice != null) {
          price = targetPrice.price;
        }

        // ถ้าราคาเป็น 0 หรือไม่มีราคาตาม pricelevel ให้ใช้ keynumber = 1
        if (price == 0) {
          PriceDataModel? defaultPrice = result.where((p) => p.keynumber == 1).firstOrNull;
          price = defaultPrice?.price ?? 0;
        }
      } else {
        price = 0;
      }
    }

    return price;
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

  /// Widget สำหรับแสดงรายการสินค้าแบบ Receive บน Mobile
  /// (ย้าย logic ไปที่ MobileProductListWidgets class)
  Widget editProductListReceiveDetailMobileWidget() {
    return _mobileProductListWidgets.buildReceiveDetailMobileWidget();
  }

  /// Widget สำหรับแสดงรายการสินค้าแบบ PickUp บน Mobile
  /// (ย้าย logic ไปที่ MobileProductListWidgets class)
  Widget editProductListPickUpDetailMobileWidget() {
    return _mobileProductListWidgets.buildPickUpDetailMobileWidget();
  }

  /// Widget สำหรับแสดงรายการสินค้าทั่วไปบน Mobile
  /// (ย้าย logic ไปที่ MobileProductListWidgets class)
  Widget editProductListDetailMobileWidget() {
    return _mobileProductListWidgets.buildDetailMobileWidget();
  }

  /// Widget สำหรับแสดงรายการสินค้าแบบ Transfer บน Mobile
  /// (ย้าย logic ไปที่ MobileProductListWidgets class)
  Widget editProductTransferListDetailMobileWidget() {
    return _mobileProductListWidgets.buildTransferDetailMobileWidget();
  }

  Widget editProductListDetailWidget(double maxWidth, double sumWidth) {
    List<Widget> widgets = [];

    String groupName = "";
    for (var index = 0; index < screenData.details!.length; index++) {
      if (widget.type == global.TransactionTypeEnum.purchasereturn ||
          widget.type == global.TransactionTypeEnum.salereturn ||
          widget.type == global.TransactionTypeEnum.stockreturnproduct ||
          widget.type == global.TransactionTypeEnum.stockreceiveproduct) {
        if (screenData.details![index].docref! != groupName) {
          String dateTime = screenData.details![index].docrefdatetime.toString();
          String tolocaldateTime = _safeParseDatetime(dateTime).toLocal().toIso8601String();
          if (screenData.details![index].docref != "") {
            widgets.add(
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  (widget.type != global.TransactionTypeEnum.stockreceiveproduct)
                      ? Row(
                          children: [
                            Text(screenData.details![index].docref!, style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
                            const SizedBox(width: 10),
                            Text(DateFormat('dd/MM/yyyy').format(_safeParseDatetime(tolocaldateTime)), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
                          ],
                        )
                      : Row(
                          children: [Text('${global.language("cart_name_label")} : ${screenData.details![index].docref!}', style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold))],
                        ),
                ],
              ),
            );
          }
          groupName = screenData.details![index].docref!;
        }
      }
      List<Widget> dataWidgets = [];

      for (var loop = 0; loop < headers.length; loop++) {
        String dataText = '';

        switch (headers[loop].code) {
          case "delete":
            dataText = "";
            break;
          case "line_number":
            dataText = (index + 1).toString();
            break;
          case "barcode":
            dataText = screenData.details![index].barcode;
            break;
          case "product_name":
            dataText = global.activeLangName(screenData.details![index].itemnames!);
            break;
          case "product_ware_house":
            dataText = global.activeLangName(screenData.details![index].whnames!);
            break;
          case "product_location":
            dataText = global.activeLangName(screenData.details![index].locationnames!);
            break;
          case "product_to_ware_house":
            dataText = global.activeLangName(screenData.details![index].towhnames!);
            break;
          case "product_to_location":
            dataText = global.activeLangName(screenData.details![index].tolocationnames!);
            break;
          case "product_unit":
            dataText = global.activeLangName(screenData.details![index].unitnames!);
            break;
          case "product_qty":
            dataText = global.formatNumber(screenData.details![index].qty);
            break;
          case "product_price_adjust":
            dataText = global.formatNumber(screenData.details![index].price);
            break;
          case "product_price":
            dataText = global.formatNumber(screenData.details![index].price);
            break;
          case "product_discount":
            dataText = screenData.details![index].discount;
            break;
          case "product_amount":
            dataText = global.formatNumber(screenData.details![index].sumamount);

            break;
        }

        if (headers[loop].code == 'delete') {
          dataWidgets.add(
            SizedBox(
              width: maxWidth * headers[loop].width / sumWidth,
              child: IconButton(
                onPressed: () {
                  deleteItemDetail(index);
                },
                icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
              ),
            ),
          );
        } else {
          dataWidgets.add(
            Container(
              padding: (loop == 0)
                  ? const EdgeInsets.only(left: 0)
                  : (loop == headers.length - 1)
                  ? const EdgeInsets.only(right: 4, left: 4)
                  : const EdgeInsets.only(left: 4),
              width: maxWidth * headers[loop].width / sumWidth,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.all(2),
                  alignment: headers[loop].alignment,
                  foregroundColor: global.theme.textColor,
                  backgroundColor: global.theme.cardColor,
                  shape: const RoundedRectangleBorder(borderRadius: BorderRadius.all(Radius.circular(0))),
                ),
                onPressed: () {
                  showDialogCommand(headers[loop].code, index, screenData.details![index]);
                },
                child: Text(dataText, textAlign: headers[loop].textAlign, style: TextStyle(fontSize: 12)),
              ),
            ),
          );
        }
      }

      widgets.add(
        Column(
          children: [
            Row(children: dataWidgets),
            productDescriptionWidget(maxWidth, sumWidth, screenData.details![index].description!),
            productOptionsListDetailWidget(maxWidth, sumWidth, screenData.details![index].extrajsonlist!),
          ],
        ),
      );
    }

    return Column(children: widgets);
  }

  /// Widget สำหรับแสดงคำอธิบายสินค้า
  /// (ย้าย logic ไปที่ DesktopProductWidgets class)
  Widget productDescriptionWidget(double maxWidth, double sumWidth, String description) {
    return _desktopProductWidgets.buildProductDescriptionWidget(maxWidth, sumWidth, description);
  }

  /// Widget สำหรับแสดงรายการ options ของสินค้า
  /// (ย้าย logic ไปที่ DesktopProductWidgets class)
  Widget productOptionsListDetailWidget(double maxWidth, double sumWidth, List<ExtraJsonListModel> opstions) {
    return _desktopProductWidgets.buildProductOptionsListDetailWidget(maxWidth, sumWidth, opstions);
  }

  /// Widget สำหรับแสดงสรุปยอดและการชำระเงิน
  /// (ย้าย logic ไปที่ SummaryWidgets class)
  Widget editSummeryWidget() {
    return _summaryWidgets.buildEditSummaryWidget();
  }

  Widget editProductListWidget() {
    return DocumentProductListWidget(
      screenData: screenData,
      transactionType: widget.type,
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
      cartList: cartList,
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
              defaultWarehouse = newDefualtwarehouse;
              defaultWarehouseNames = newDefualtwarehousenames;
              defaultLocation = newDefualtlocation;
              defaultLocationNames = newDefualtlocationnames;
              defaultToWarehouse = newDefualttowarehouse;
              defaultToWarehouseNames = newDefualttowarehousenames;
              defaultToLocation = newDefualttolocation;
              defaultToLocationNames = newDefualttolocationnames;
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
      transactionType: widget.type,
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
    // สำหรับ PO: รวม approval status badge กับ document header
    Widget documentHeader = editDocumentWidget();
    if (widget.type == global.TransactionTypeEnum.purchaseorder) {
      documentHeader = Column(
        children: [
          Expanded(child: editDocumentWidget()),
          // แสดง badge สถานะอนุมัติด้านล่างของเอกสาร
          _buildApprovalStatusBadge(),
        ],
      );
    }
    List<Widget> childrenList = [documentHeader, editProductListWidget()];
    List<Widget> tabx = [Tab(text: global.language("doc_header")), Tab(text: global.language("doc_details"))];
    if (widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.saleorder &&
        widget.type != global.TransactionTypeEnum.purchaseorder &&
        widget.type != global.TransactionTypeEnum.purchasepartial &&
        widget.type != global.TransactionTypeEnum.quotation) {
      tabx.add(Tab(text: global.language("total")));
    }
    if (widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.saleorder &&
        widget.type != global.TransactionTypeEnum.purchaseorder &&
        widget.type != global.TransactionTypeEnum.purchasepartial &&
        widget.type != global.TransactionTypeEnum.quotation) {
      childrenList.add(editSummeryWidget());
    }
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
                  child: TabBarView(controller: editTabController, physics: const NeverScrollableScrollPhysics(), children: childrenList),
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  void deleteDoc() {
    context.read<TransBloc>().add(TransDelete(guid: screenData.guidfixed!, type: widget.type));
  }

  void saveOrUpdateData({required bool taxInvoice}) {
    AppLogger.info('🔵 [saveOrUpdateData] START - taxInvoice: $taxInvoice, type: ${widget.type}');
    AppLogger.info('🔵 [saveOrUpdateData] guidfixed: ${screenData.guidfixed}, docno: ${screenData.docno}');
    AppLogger.info('🔵 [saveOrUpdateData] details count: ${screenData.details?.length ?? 0}');

    // คำนวณใหม่
    if (widget.type != global.TransactionTypeEnum.adjust) {
      _transactionCalculator.calTotalValue();
    }

    final verifyResult = _transactionCalculator.verifyPayment();
    AppLogger.info('🔵 [saveOrUpdateData] verifyPayment result: $verifyResult');

    if (verifyResult) {
      DateTime docDatetimeUtc = _safeParseDatetime(screenData.docdatetime);
      screenData.docdatetime = docDatetimeUtc.toUtc().toIso8601String();
      DateTime docRefDatetimeUtc = _safeParseDatetime(screenData.docrefdate);
      screenData.docrefdate = docRefDatetimeUtc.toUtc().toIso8601String();
      DateTime taxDocDatetimeUtc = _safeParseDatetime(screenData.taxdocdate);
      screenData.taxdocdate = taxDocDatetimeUtc.toUtc().toIso8601String();

      for (var data in screenData.details!) {
        data.linenumber = screenData.details!.indexOf(data) + 1;

        data.docdatetime = screenData.docdatetime;

        DateTime docrefdatetimeDetailUtc = _safeParseDatetime(data.docrefdatetime);
        data.docrefdatetime = docrefdatetimeDetailUtc.toUtc().toIso8601String();

        ///   ดึงตัวตั้งตัวหารจาก บาร์โค้ดอ้างอิง
        if (data.itemtype == 0 && data.refbarcodes!.isNotEmpty) {
          for (var element in data.refbarcodes!) {
            data.standvalue = element.standvalue;
            data.dividevalue = element.dividevalue;
          }
        }
      }

      screenData.paymentStructs = [];

      if (payTransfer.isNotEmpty) {
        screenData.paymentStructs!.addAll(payTransfer);
      }

      if (payCheque.isNotEmpty) {
        screenData.paymentStructs!.addAll(payCheque);
      }

      if (payCoupon.isNotEmpty) {
        screenData.paymentStructs!.addAll(payCoupon);
      }

      if (payQr.isNotEmpty) {
        screenData.paymentStructs!.addAll(payQr);
      }

      if (payCreditCard.isNotEmpty) {
        screenData.paymentStructs!.addAll(payCreditCard);
      }

      screenData.paymentdetailraw = jsonEncode(screenData.paymentStructs);

      if (screenData.ispos! && screenData.iscancel) {
        screenData = screenDataTemp;
      }

      // บันทึกประเภทการจัดซื้อลง screenData ก่อน save (สำหรับ PO)
      _savePOPurchaseType();

      // บันทึกข้อมูลผู้สร้าง/แก้ไขเอกสาร (สำหรับ Data History)
      screenData.creatorcode = global.userLoginData.code;
      screenData.creatorname = global.userLoginData.name;

      /// ออกใบกำกับภาษีแบบเต็ม
      if (taxInvoice == true) {
        // ใช้ Event ใหม่สำหรับการสร้างใบกำกับภาษีแบบเต็ม
        AppLogger.info('🔵 [saveOrUpdateData] Sending TransCreateFullInvoice event');
        context.read<TransBloc>().add(TransCreateFullInvoice(guid: screenData.guidfixed!, trans: screenData, type: widget.type));
      } else {
        // ตรวจสอบว่า guidfixed มีค่าจริงๆ (ไม่ใช่ null หรือ empty) ถึงจะ update
        if (screenData.guidfixed?.isNotEmpty ?? false) {
          AppLogger.info('🔵 [saveOrUpdateData] Sending TransUpdate event - guidfixed: ${screenData.guidfixed}');
          context.read<TransBloc>().add(TransUpdate(guid: screenData.guidfixed!, trans: screenData, type: widget.type));
        } else {
          AppLogger.info('🔵 [saveOrUpdateData] Sending TransSave event (new document)');
          context.read<TransBloc>().add(TransSave(trans: screenData, type: widget.type));
        }
      }
    } else {
      AppLogger.warn('🔴 [saveOrUpdateData] verifyPayment() returned FALSE - save blocked');
    }
  }

  void showDialogCommand(String cmd, int index, TransactionDetailModel details) async {
    if (cmd == 'barcode') {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => BarcodeSearchScreen(
            word: details.barcode,
            screen: (widget.type == global.TransactionTypeEnum.sale || widget.type == global.TransactionTypeEnum.saleorder || widget.type == global.TransactionTypeEnum.salereturn)
                ? 'not_material'
                : 'material',
          ),
        ),
      ).then((value) {
        ProductBarcodeModel result = value;
        if (result.barcode!.trim().isNotEmpty) {
          setState(() {
            screenData.details![index].itemguid = result.guidfixed;
            screenData.details![index].barcode = result.barcode!;
            screenData.details![index].itemcode = result.itemcode ?? "";
            screenData.details![index].itemnames = result.names;
            screenData.details![index].multiunit = true;
            screenData.details![index].unitcode = result.itemunitcode;
            screenData.details![index].dividevalue = result.dividevalue!;
            screenData.details![index].standvalue = result.standvalue!;
            screenData.details![index].unitnames = result.itemunitnames;
            screenData.details![index].whcode = defaultWarehouse;
            screenData.details![index].whnames = defaultWarehouseNames;
            screenData.details![index].locationcode = defaultLocation;
            screenData.details![index].locationnames = defaultLocationNames;
            screenData.details![index].towhcode = defaultToWarehouse;
            screenData.details![index].towhnames = defaultToWarehouseNames;
            screenData.details![index].tolocationcode = defaultToLocation;
            screenData.details![index].tolocationnames = defaultToLocationNames;
            screenData.details![index].taxtype = result.taxtype!;
            screenData.details![index].qty = 1;
            screenData.details![index].vatcal = result.vatcal;
            screenData.details![index].price = getPrice(result.prices);
            screenData.details![index].manufacturerguid = result.manufacturerguid;
          });

          _transactionCalculator.calTotalValue();
        }
      });
    } else if (cmd == 'item_code') {
      // เพิ่ม case สำหรับ item_code
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => BarcodeSearchScreen(
            word: details.itemcode, // ส่ง itemcode แทน barcode
            screen: (widget.type == global.TransactionTypeEnum.sale || widget.type == global.TransactionTypeEnum.saleorder || widget.type == global.TransactionTypeEnum.salereturn)
                ? 'not_material'
                : 'material',
          ),
        ),
      ).then((value) {
        ProductBarcodeModel result = value;
        if (result.barcode!.trim().isNotEmpty) {
          setState(() {
            screenData.details![index].itemguid = result.guidfixed;
            screenData.details![index].barcode = result.barcode!;
            screenData.details![index].itemcode = result.itemcode ?? "";
            screenData.details![index].itemnames = result.names;
            screenData.details![index].multiunit = true;
            screenData.details![index].unitcode = result.itemunitcode;
            screenData.details![index].dividevalue = result.dividevalue!;
            screenData.details![index].standvalue = result.standvalue!;
            screenData.details![index].unitnames = result.itemunitnames;
            screenData.details![index].whcode = defaultWarehouse;
            screenData.details![index].whnames = defaultWarehouseNames;
            screenData.details![index].locationcode = defaultLocation;
            screenData.details![index].locationnames = defaultLocationNames;
            screenData.details![index].towhcode = defaultToWarehouse;
            screenData.details![index].towhnames = defaultToWarehouseNames;
            screenData.details![index].tolocationcode = defaultToLocation;
            screenData.details![index].tolocationnames = defaultToLocationNames;
            screenData.details![index].taxtype = result.taxtype!;
            screenData.details![index].qty = 1;
            screenData.details![index].vatcal = result.vatcal;
            screenData.details![index].price = getPrice(result.prices);
            screenData.details![index].manufacturerguid = result.manufacturerguid;
          });

          _transactionCalculator.calTotalValue();
        }
      });
    } else if (cmd == 'product_name') {
      // แสดง dialog สำหรับแก้ไขชื่อสินค้าและหมายเหตุ
      String currentProductName = global.activeLangName(details.itemnames!);
      String currentDescription = details.description ?? "";

      TextEditingController productNameController = TextEditingController(text: currentProductName);
      TextEditingController descriptionController = TextEditingController(text: currentDescription);

      await showDialog(
        context: context,
        builder: (BuildContext dialogContext) {
          return Dialog(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            child: SizedBox(
              width: (global.isMobileScreen(context)) ? 350 : 420,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Header
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: global.theme.primaryColor,
                      borderRadius: const BorderRadius.only(topLeft: Radius.circular(16), topRight: Radius.circular(16)),
                    ),
                    child: Row(
                      children: [
                        Icon(Icons.edit_note, color: global.theme.onPrimaryColor, size: 24),
                        SizedBox(width: 12),
                        Expanded(
                          child: Text(
                            global.language("edit_product_info"),
                            style: TextStyle(color: global.theme.onPrimaryColor, fontSize: 16, fontWeight: FontWeight.bold),
                          ),
                        ),
                      ],
                    ),
                  ),
                  // Content
                  Padding(
                    padding: EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // ฟิลด์แก้ไขชื่อสินค้า
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(6),
                              decoration: BoxDecoration(color: global.theme.primaryColor.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(6)),
                              child: Icon(Icons.inventory_2_outlined, color: global.theme.primaryColor, size: 18),
                            ),
                            SizedBox(width: 8),
                            Text(
                              global.language("product_name"),
                              style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: global.theme.primaryColor),
                            ),
                          ],
                        ),
                        SizedBox(height: 8),
                        TextField(
                          controller: productNameController,
                          decoration: InputDecoration(
                            hintText: global.language("enter_product_name"),
                            hintStyle: TextStyle(color: global.theme.formHintColor),
                            border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                            focusedBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(10),
                              borderSide: BorderSide(color: global.theme.primaryColor, width: 2),
                            ),
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                          ),
                          maxLines: 2,
                        ),
                        const SizedBox(height: 16),
                        // ฟิลด์แก้ไขหมายเหตุ
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(6),
                              decoration: BoxDecoration(color: global.theme.primaryColor.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(6)),
                              child: Icon(Icons.description_outlined, color: global.theme.primaryColor, size: 18),
                            ),
                            SizedBox(width: 8),
                            Text(
                              global.language("description"),
                              style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: global.theme.primaryColor),
                            ),
                          ],
                        ),
                        SizedBox(height: 8),
                        TextField(
                          controller: descriptionController,
                          decoration: InputDecoration(
                            hintText: global.language("enter_description"),
                            hintStyle: TextStyle(color: global.theme.formHintColor),
                            border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                            focusedBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(10),
                              borderSide: BorderSide(color: global.theme.primaryColor, width: 2),
                            ),
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                          ),
                          maxLines: 3,
                        ),
                      ],
                    ),
                  ),
                  // Footer
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      border: Border(top: BorderSide(color: global.theme.dividerBorderColor)),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        TextButton(
                          onPressed: () => Navigator.of(dialogContext).pop(),
                          style: TextButton.styleFrom(padding: EdgeInsets.symmetric(horizontal: 20, vertical: 10)),
                          child: Text(global.language("cancel"), style: TextStyle(color: global.theme.textSecondaryColor)),
                        ),
                        SizedBox(width: 8),
                        ElevatedButton(
                          onPressed: () {
                            // อัพเดทชื่อสินค้าตามภาษาที่ใช้งาน
                            for (int i = 0; i < screenData.details![index].itemnames!.length; i++) {
                              if (screenData.details![index].itemnames![i].code == global.userLanguage) {
                                screenData.details![index].itemnames![i].name = productNameController.text;
                                break;
                              }
                            }

                            // อัพเดทหมายเหตุ
                            screenData.details![index].description = descriptionController.text;

                            Navigator.of(dialogContext).pop();
                            setState(() {});
                          },
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.primaryColor,
                            foregroundColor: global.theme.onPrimaryColor,
                            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 10),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                          ),
                          child: Text(global.language("confirm")),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      );

      // dispose controllers หลังจาก dialog ปิดแล้ว
      productNameController.dispose();
      descriptionController.dispose();
    } else if (cmd == 'product_ware_house') {
      if (details.itemguid.trim().isNotEmpty) {
        WarehouseModel? result = await showWareHouseDialog(context, details.itemguid, warehouseList) ?? WarehouseModel(guidfixed: '', code: '');
        if (result.code.isNotEmpty) {
          setState(() {
            screenData.details![index].whcode = result.code;
            screenData.details![index].whnames = result.names;

            if (result.location.isNotEmpty) {
              screenData.details![index].locationcode = result.location[0].code;
              screenData.details![index].locationnames = result.location[0].names;
            } else {
              screenData.details![index].locationcode = "";
              screenData.details![index].locationnames = [];
            }
          });
        }
      }
    } else if (cmd == 'product_location') {
      if (details.itemguid.trim().isNotEmpty && details.whcode.trim().isNotEmpty) {
        LocationModel? result = await showWareHouseLocationDialog(context, details.itemguid, details.whcode) ?? LocationModel(code: '');
        if (result.code.isNotEmpty) {
          setState(() {
            screenData.details![index].locationcode = result.code;
            screenData.details![index].locationnames = result.names;
          });
        }
      }
    } else if (cmd == 'product_to_ware_house') {
      if (details.itemguid.trim().isNotEmpty) {
        WarehouseModel? result = await showWareHouseDialog(context, details.itemguid, warehouseList) ?? WarehouseModel(guidfixed: '', code: '');
        if (result.code.isNotEmpty) {
          setState(() {
            screenData.details![index].towhcode = result.code;
            screenData.details![index].towhnames = result.names;

            if (result.location.isNotEmpty) {
              screenData.details![index].tolocationcode = result.location[0].code;
              screenData.details![index].tolocationnames = result.location[0].names;
            } else {
              screenData.details![index].tolocationcode = "";
              screenData.details![index].tolocationnames = [];
            }
          });
        }
      }
    } else if (cmd == 'product_to_location') {
      if (details.itemguid.trim().isNotEmpty && details.towhcode!.trim().isNotEmpty) {
        LocationModel? result = await showWareHouseLocationDialog(context, details.itemguid, details.towhcode!) ?? LocationModel(code: '');
        if (result.code.isNotEmpty) {
          setState(() {
            screenData.details![index].tolocationcode = result.code;
            screenData.details![index].tolocationnames = result.names;
          });
        }
      }
    } else if (cmd == 'product_unit') {
      if (details.itemguid.trim().isNotEmpty && details.multiunit == true) {
        ProductBarcodeModel? result = await showUnitsDialog(context, details.itemguid) ?? ProductBarcodeModel(guidfixed: '');

        if (result.guidfixed.isNotEmpty) {
          screenData.details![index].docdatetime = DateTime.now().toUtc().toIso8601String(); // แปลงเป็น UTC ตามกฎ CLAUDE.md
          screenData.details![index].itemguid = result.guidfixed;
          screenData.details![index].barcode = result.barcode!;
          screenData.details![index].itemcode = result.itemcode ?? "";
          screenData.details![index].itemnames = result.names;
          screenData.details![index].unitcode = result.itemunitcode;
          screenData.details![index].qty = 1;
          screenData.details![index].price = getPrice(result.prices);
          screenData.details![index].discount = '';
          screenData.details![index].sumofcost = 0;
          screenData.details![index].sumamount = 0;
          screenData.details![index].remark = '';
          screenData.details![index].linenumber = 0;
          screenData.details![index].shelfcode = '';
          screenData.details![index].totalvaluevat = 0;
          screenData.details![index].totalqty = 0;
          screenData.details![index].standvalue = result.standvalue!;
          screenData.details![index].dividevalue = result.dividevalue!;
          screenData.details![index].multiunit = true;
          screenData.details![index].unitnames = result.itemunitnames;
          screenData.details![index].calcflag = calcflag;
          screenData.details![index].vattype = 0;
          screenData.details![index].averagecost = 0;
          screenData.details![index].sumamountexcludevat = 0;
          screenData.details![index].discountamount = 0;
          screenData.details![index].ispos = 0;
          screenData.details![index].laststatus = 0;
          screenData.details![index].itemtype = 0;
          screenData.details![index].inquirytype = 0;
          screenData.details![index].priceexcludevat = 0;
          screenData.details![index].taxtype = result.taxtype!;
          screenData.details![index].vatcal = result.vatcal;
          screenData.details![index].towhcode = defaultToWarehouse;
          screenData.details![index].towhnames = defaultToWarehouseNames;
          screenData.details![index].tolocationcode = defaultToLocation;
          screenData.details![index].tolocationnames = defaultLocationNames;

          _transactionCalculator.calTotalValue();
          setState(() {});
        }
      }
    } else if (cmd == 'product_qty') {
      String? result = await showQtyDialog(context, details.qty, global.activeLangName(details.itemnames!)) ?? screenData.details![index].qty.toString();
      final numbercheck = result.replaceAll(',', '');
      setState(() {
        screenData.details![index].qty = double.parse(numbercheck);
        _transactionCalculator.calTotalValue();
      });
    } else if (cmd == 'product_price' || cmd == 'product_price_adjust') {
      String? result = await showPriceDialog(context, details.price, global.activeLangName(details.itemnames!)) ?? screenData.details![index].price.toString();
      final numbercheck = result.replaceAll(',', '');
      setState(() {
        screenData.details![index].price = double.parse(numbercheck);
        _transactionCalculator.calTotalValue();
      });
    } else if (cmd == 'product_discount') {
      String? result = await showDiscountDialog(context, details.discount, global.activeLangName(details.itemnames!)) ?? screenData.details![index].discount.toString();

      String numbercheck = result.replaceAll(',', '');

      screenData.details![index].discount = numbercheck;

      setState(() {
        _transactionCalculator.calTotalValue();
      });
    } else if (cmd == 'product_amount') {
      // เพิ่ม case สำหรับ product_amount โดยเฉพาะสำหรับ stockreceiveproduct
      if (widget.type == global.TransactionTypeEnum.stockreceiveproduct ||
          widget.type == global.TransactionTypeEnum.purchase ||
          widget.type == global.TransactionTypeEnum.purchaseorder ||
          widget.type == global.TransactionTypeEnum.purchasepartial ||
          widget.type == global.TransactionTypeEnum.accrualreceive ||
          widget.type == global.TransactionTypeEnum.adjust) {
        String? result =
            await showNumericInputDialog(context, title: '${global.activeLangName(details.itemnames!)} ${global.language("amount")}', initialValue: details.sumamount.toString(), decimalPlaces: 2) ??
            screenData.details![index].sumamount.toString();
        final numbercheck = result.replaceAll(',', '');
        setState(() {
          screenData.details![index].sumamount = double.parse(numbercheck);
        });
      }
    } else if (cmd == 'docref') {
      if (widget.type == global.TransactionTypeEnum.purchase ||
          widget.type == global.TransactionTypeEnum.purchasereturn ||
          widget.type == global.TransactionTypeEnum.purchasepartial ||
          widget.type == global.TransactionTypeEnum.accrualreceive ||
          widget.type == global.TransactionTypeEnum.sale ||
          widget.type == global.TransactionTypeEnum.salereturn) {
        _searchHandlers.selectDocRefForItem(itemIndex: index);
      }
    } else if (cmd == 'docrefdatetime') {
      if (widget.type == global.TransactionTypeEnum.purchase ||
          widget.type == global.TransactionTypeEnum.purchasereturn ||
          widget.type == global.TransactionTypeEnum.purchasepartial ||
          widget.type == global.TransactionTypeEnum.accrualreceive ||
          widget.type == global.TransactionTypeEnum.sale ||
          widget.type == global.TransactionTypeEnum.salereturn) {
        _searchHandlers.selectDocRefForItem(itemIndex: index);
      }
    } else if (cmd == 'eventqty') {
      // สำหรับ purchase - แก้ไขจำนวนเหตุการณ์
      if (widget.type == global.TransactionTypeEnum.purchase) {
        String? result = await showQtyDialog(context, details.eventqty ?? 0, global.language('event_quantity')) ?? (screenData.details![index].eventqty ?? 0).toString();
        final numbercheck = result.replaceAll(',', '');
        setState(() {
          screenData.details![index].eventqty = double.parse(numbercheck);
        });
      }
    } else if (cmd == 'description') {
      // แก้ไขหมายเหตุสินค้า
      String? result = await showDescriptionDialog(context, details.description ?? '', global.activeLangName(details.itemnames!)) ?? screenData.details![index].description;
      setState(() {
        screenData.details![index].description = result;
      });
    } else if (cmd == 'options') {
      // แก้ไขตัวเลือกเพิ่มเติม - เปิดหน้าจัดการตัวเลือก
      showDialog(
        context: context,
        builder: (BuildContext context) {
          return AlertDialog(
            title: Text('${global.language("additional_options")}: ${global.activeLangName(details.itemnames!)}'),
            content: SizedBox(
              width: (global.isMobileScreen(context)) ? 350 : 400,
              height: 300,
              child: details.extrajsonlist != null && details.extrajsonlist!.isNotEmpty
                  ? Column(
                      children: [
                        Text('${global.language("options_found")} ${details.extrajsonlist!.length} ${global.language("items_text")}'),
                        SizedBox(height: 10),
                        Expanded(
                          child: ListView.builder(
                            itemCount: details.extrajsonlist!.length,
                            itemBuilder: (context, optionIndex) {
                              final option = details.extrajsonlist![optionIndex];
                              return Card(
                                child: ListTile(
                                  title: Text(global.activeLangName(option.itemnames!)),
                                  subtitle: Text('${global.language("price_colon")}: ${global.formatNumber(option.price!)} | ${global.language("amount")}: ${global.formatNumber(option.qty!)}'),
                                  trailing: IconButton(
                                    icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                                    onPressed: () {
                                      setState(() {
                                        screenData.details![index].extrajsonlist!.removeAt(optionIndex);
                                      });
                                      Navigator.pop(context);
                                    },
                                  ),
                                ),
                              );
                            },
                          ),
                        ),
                      ],
                    )
                  : Center(child: Text(global.language('no_additional_options'))),
            ),
            actions: <Widget>[
              TextButton(
                child: Text(global.language("close")),
                onPressed: () {
                  Navigator.pop(context);
                },
              ),
            ],
          );
        },
      );
    } else if (cmd == 'product_name') {
      // แก้ไขชื่อสินค้า - นำไปสู่หน้าค้นหาสินค้า
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => BarcodeSearchScreen(
            word: details.barcode,
            screen: (widget.type == global.TransactionTypeEnum.sale || widget.type == global.TransactionTypeEnum.saleorder || widget.type == global.TransactionTypeEnum.salereturn)
                ? 'not_material'
                : 'material',
          ),
        ),
      ).then((value) {
        ProductBarcodeModel result = value;
        if (result.barcode!.trim().isNotEmpty) {
          setState(() {
            screenData.details![index].itemguid = result.guidfixed;
            screenData.details![index].barcode = result.barcode!;
            screenData.details![index].itemcode = result.itemcode ?? "";
            screenData.details![index].itemnames = result.names;
            screenData.details![index].multiunit = true;
            screenData.details![index].unitcode = result.itemunitcode;
            screenData.details![index].dividevalue = result.dividevalue!;
            screenData.details![index].standvalue = result.standvalue!;
            screenData.details![index].unitnames = result.itemunitnames;
            screenData.details![index].taxtype = result.taxtype!;
            screenData.details![index].vatcal = result.vatcal;
            screenData.details![index].price = getPrice(result.prices);
            screenData.details![index].manufacturerguid = result.manufacturerguid;
          });

          _transactionCalculator.calTotalValue();
        }
      });
    } else if (cmd == 'qty') {
      // Alias สำหรับ product_qty
      String? result = await showQtyDialog(context, details.qty, global.activeLangName(details.itemnames!)) ?? screenData.details![index].qty.toString();
      final numbercheck = result.replaceAll(',', '');
      setState(() {
        screenData.details![index].qty = double.parse(numbercheck);
        _transactionCalculator.calTotalValue();
      });
    } else if (cmd == 'price') {
      // Alias สำหรับ product_price
      String? result = await showPriceDialog(context, details.price, global.activeLangName(details.itemnames!)) ?? screenData.details![index].price.toString();
      final numbercheck = result.replaceAll(',', '');
      setState(() {
        screenData.details![index].price = double.parse(numbercheck);
        _transactionCalculator.calTotalValue();
      });
    } else if (cmd == 'warehouse') {
      // Alias สำหรับ product_ware_house
      if (details.itemguid.trim().isNotEmpty) {
        WarehouseModel? result = await showWareHouseDialog(context, details.itemguid, warehouseList) ?? WarehouseModel(guidfixed: '', code: '');
        if (result.code.isNotEmpty) {
          setState(() {
            screenData.details![index].whcode = result.code;
            screenData.details![index].whnames = result.names;

            if (result.location.isNotEmpty) {
              screenData.details![index].locationcode = result.location[0].code;
              screenData.details![index].locationnames = result.location[0].names;
            } else {
              screenData.details![index].locationcode = "";
              screenData.details![index].locationnames = [];
            }
          });
        }
      }
    } else if (cmd == 'location') {
      // Alias สำหรับ product_location
      if (details.itemguid.trim().isNotEmpty && details.whcode.trim().isNotEmpty) {
        LocationModel? result = await showWareHouseLocationDialog(context, details.itemguid, details.whcode) ?? LocationModel(code: '');
        if (result.code.isNotEmpty) {
          setState(() {
            screenData.details![index].locationcode = result.code;
            screenData.details![index].locationnames = result.names;
          });
        }
      }
    }
  }

  void deleteItemDetail(int index) {
    setState(() {
      screenData.details!.removeAt(index);
    });

    // ลบ มูลค่าเดิม , ผลต่าง , ผลต่างที่ถูกต้อง ตามเอกสารอ้างอิงที่มีใน Details
    if (widget.type == global.TransactionTypeEnum.salereturn ||
        widget.type == global.TransactionTypeEnum.purchasereturn ||
        widget.type == global.TransactionTypeEnum.sale ||
        widget.type == global.TransactionTypeEnum.purchase ||
        widget.type == global.TransactionTypeEnum.purchasepartial ||
        widget.type == global.TransactionTypeEnum.accrualreceive) {
      List<String> docrefsdetail = [];
      for (var val in screenData.details!) {
        if (val.docref!.isNotEmpty) {
          addIfNotExists(docrefsdetail, val.docref!);
        }
      }
      var docrefsremove = docrefs.where((element) => !docrefsdetail.contains(element['docref'])).toList();
      docrefs = docrefs.where((element) => docrefsdetail.contains(element['docref'])).toList();

      if (docrefsremove.isNotEmpty) {
        for (var val in docrefsremove) {
          screenData.reftotaloriginal = (screenData.reftotaloriginal! - val['totaloriginal']);
        }
        if (screenData.details!.isEmpty) {
          screenData.reftotaldiff = 0;
          screenData.reftotalcorrect = 0;
        }
      }
    }

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
                onPressed: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => PdfPreviewPage(screenData: screenData, type: widget.type),
                    ),
                  ).then((value) {
                    Navigator.of(context).pop();
                    clearScreenData();
                    editTabController.animateTo(0);
                  });
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

  /// สร้าง main content รองรับ side panel ค้นหาเอกสาร (จอใหญ่)
  Widget _buildMainContent(BoxConstraints constraints) {
    final isLargeScreen = constraints.maxWidth > 700;

    // สร้าง content หลัก
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
      mainContent = TabBarView(physics: const NeverScrollableScrollPhysics(), controller: tabController, children: _showPreview ? [editWidget(), previewWidget()] : [editWidget()]);
    }

    // ถ้าเปิด side panel ค้นหาเอกสาร (เฉพาะจอใหญ่)
    if (_showDocSearchPanel && isLargeScreen) {
      // จำกัดขนาด panel ไม่ให้เล็กหรือใหญ่เกินไป
      final minWidth = 300.0;
      final maxWidth = constraints.maxWidth * 0.6; // ไม่เกิน 60% ของหน้าจอ

      return Row(
        children: [
          // Side panel ค้นหาเอกสาร (ด้านซ้าย)
          SizedBox(
            width: _docSearchPanelWidth.clamp(minWidth, maxWidth),
            child: TransSearchScreen(
              key: _transSearchKey,
              type: widget.type,
              custcode: '',
              isEmbedded: true,
              onDocumentSelected: (document) {
                // เมื่อเลือกเอกสาร
                _onDocumentSelectedFromPanel(document);
              },
              onClose: () {
                // เมื่อกดปิด panel
                setState(() {
                  _showDocSearchPanel = false;
                });
              },
            ),
          ),
          // Handle สำหรับปรับขนาด panel (ลากแนวนอน)
          MouseRegion(
            cursor: SystemMouseCursors.resizeColumn,
            child: GestureDetector(
              onHorizontalDragUpdate: (details) {
                setState(() {
                  _docSearchPanelWidth += details.delta.dx;
                  _docSearchPanelWidth = _docSearchPanelWidth.clamp(minWidth, maxWidth);
                });
              },
              child: Container(
                width: 8,
                color: Colors.transparent,
                child: Center(
                  child: Container(
                    width: 4,
                    height: 40,
                    decoration: BoxDecoration(color: global.theme.iconSecondaryColor, borderRadius: BorderRadius.circular(2)),
                  ),
                ),
              ),
            ),
          ),
          // เนื้อหาหลัก
          Expanded(child: mainContent),
        ],
      );
    }

    return mainContent;
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

      /// sort linenumber (ถ้ามี details)
      if (screenData.details != null && screenData.details!.isNotEmpty) {
        screenData.details!.sort((a, b) => a.linenumber.compareTo(b.linenumber));
      }
      screenData.docdatetime = localDateTime.toUtc().toIso8601String(); // แปลงเป็น UTC ตามกฎ CLAUDE.md
      showPayDetail == 0;

      loadDataToScreen();
      // คำนวณยอดรวมใหม่
      _transactionCalculator.calTotalValue();

      // โหลดประเภทการจัดซื้อสำหรับ PO
      if (widget.type == global.TransactionTypeEnum.purchaseorder && document.docno.isNotEmpty) {
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
    global.TransactionTypeEnum transactionType = widget.type;
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
        title: Text(global.transactionName(widget.type)),
        actions: <Widget>[
          EditFontSizeControl(onChanged: () => setState(() {})),
          // Loading indicator บน AppBar - แสดงเมื่อกำลังโหลด/บันทึก/ลบข้อมูล
          if (_isLoading)
            Padding(
              padding: const EdgeInsets.only(right: 12.0),
              child: Center(
                child: SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    valueColor: AlwaysStoppedAnimation<Color>(global.theme.onPrimaryColor),
                  ),
                ),
              ),
            ),
          /// ออกใบกำกับแบบเต็ม
          (screenData.ispos == true && screenData.iscancel == false)
              ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  /// ยืนยันการ ออกใบกำกับเต็ม
                  final result = await showAlertConfirmFullInvoiceDialog(context, screenData.docno);
                  if (result != null && result) {
                    saveOrUpdateData(taxInvoice: true);
                  }
                },
                icon: Icon(Icons.receipt_long_outlined, size: 26.0),
              )
              : Container(),
          (screenData.slipurl != '')
              ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  /// show dialog preview image
                  showDialog(
                    context: context,
                    builder: (BuildContext context) {
                      return AlertDialog(content: Image.network(screenData.slipurl!));
                    },
                  );
                },
                icon: Icon(Icons.image_outlined, size: 26.0),
              )
              : Container(),

          /// export csv
          (transactionType == global.TransactionTypeEnum.sale && (screenData.guidfixed?.isNotEmpty ?? false))
              ? IconButton(
                onPressed: () {
                  /// dialog select language code
                  showDialog<String>(
                    context: context,
                    builder: (BuildContext context) {
                      // Initialize the selected language outside of StatefulBuilder
                      LanguageModel? selectedLanguage = global.config.languages[0];
                      return AlertDialog(
                        title: Text(global.language('select_language')),
                        content: StatefulBuilder(
                          builder: (BuildContext context, StateSetter setState) {
                            return InputDecorator(
                              decoration: const InputDecoration(border: OutlineInputBorder()),
                              child: DropdownButtonHideUnderline(
                                child: DropdownButton<LanguageModel>(
                                  value: selectedLanguage,
                                  icon: Icon(Icons.arrow_drop_down),
                                  style: TextStyle(color: Colors.deepPurple),
                                  underline: Container(color: Colors.deepPurpleAccent),
                                  onChanged: (LanguageModel? value) {
                                    setState(() {
                                      selectedLanguage = value!;
                                    });
                                  },
                                  isDense: true,
                                  isExpanded: true,
                                  items: global.config.languages.map<DropdownMenuItem<LanguageModel>>((LanguageModel value) {
                                    return DropdownMenuItem<LanguageModel>(
                                      value: value,
                                      child: Row(
                                        children: <Widget>[
                                          Image.asset(
                                            'assets/flags/${value.code}.png', // Ensure the image path is correct
                                            width: 30,
                                            height: 30,
                                            errorBuilder: (context, error, stackTrace) {
                                              return Icon(Icons.error); // Error icon if the image fails to load
                                            },
                                          ),
                                          const SizedBox(width: 10), // Spacing between the image and text
                                          Text(value.name!),
                                        ],
                                      ),
                                    );
                                  }).toList(),
                                ),
                              ),
                            );
                          },
                        ),
                        actions: <Widget>[
                          // Text button cancel
                          ElevatedButton(
                            style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
                            onPressed: () {
                              Navigator.pop(context);
                            },
                            child: Text(global.language('cancel')),
                          ),
                          // Export button
                          ElevatedButton(
                            style: ElevatedButton.styleFrom(backgroundColor: global.theme.infoHighlightTextColor),
                            onPressed: () {
                              context.read<ExportCsvBloc>().add(SaleInvoiceExport(languageCode: selectedLanguage!.code!));
                              Navigator.pop(context);
                            },
                            child: Text(global.language('export')),
                          ),
                        ],
                      );
                    },
                  );
                },
                icon: Icon(Icons.download),
              )
              : Container(),
          ((screenData.guidfixed?.isNotEmpty ?? false) && screenData.iscancel == false)
              ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  final result = await showAlertConfirmDeleteDialog(context, screenData.docno);
                  if (result != null && result) {
                    deleteDoc();
                  }
                },
                icon: Icon(Icons.delete, size: 26.0),
              )
              : Container(),

          (!screenData.iscancel && (screenData.guidfixed?.isNotEmpty ?? false))
              ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  if (screenData.details!.isNotEmpty) {
                    final result = await showAlertConfirmCancelDialog(context);
                    if (result != null && result.isNotEmpty) {
                      screenData.iscancel = true;
                      screenData.cancelreason = result;
                      screenData.cancelusercode = global.profileData.username;
                      screenData.cancelusername = global.profileData.name;
                      screenData.canceldatetime = DateTime.now().toLocal().toIso8601String();
                      screenData.canceltime = DateTime.now().toLocal().toIso8601String();
                      screenDataTemp.iscancel = true;
                      screenDataTemp.cancelreason = result;
                      screenDataTemp.cancelusercode = global.profileData.username;
                      screenDataTemp.cancelusername = global.profileData.name;
                      screenDataTemp.canceldatetime = DateTime.now().toLocal().toIso8601String();
                      screenDataTemp.canceltime = DateTime.now().toLocal().toIso8601String();
                      saveOrUpdateData(taxInvoice: false);
                    }
                  }
                },
                icon: Icon(Icons.cancel_rounded, size: 26.0),
              )
              : Container(),

          /// ปุ่มดูประวัติการพิมพ์ - แสดงเมื่อเอกสารถูกบันทึกแล้ว
          /// ใช้ PrintHistoryIconButton เพื่อแสดง badge จำนวนครั้งที่พิมพ์
          ((screenData.guidfixed?.isNotEmpty ?? false) && screenData.ispos == false)
              ? PrintHistoryIconButton(collection: _getCollectionNameForType(widget.type), docNo: screenData.docno, title: _getDocumentTitle(widget.type), iconSize: 26.0)
              : Container(),

          /// ปุ่มดูประวัติการแก้ไข - แสดงเฉพาะใบสั่งซื้อ (PO) เมื่อเอกสารถูกบันทึกแล้ว (guidfixed ต้องมีค่าจริงๆ)
          (widget.type == global.TransactionTypeEnum.purchaseorder && (screenData.guidfixed?.isNotEmpty ?? false) && screenData.docno.isNotEmpty)
              ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showDataHistoryDialog(
                    context,
                    docNo: screenData.docno,
                    title: global.language('edit_history'),
                    // ส่งข้อมูลผู้สร้างจาก screenData (modifier อยู่ใน datahistory)
                    creatorCode: screenData.creatorcode,
                    creatorName: screenData.creatorname,
                    createdAt: screenData.createdat,
                  );
                },
                icon: Icon(Icons.history, size: 26.0),
                tooltip: global.language('view_edit_history'),
              )
              : Container(),

          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () async {
              clearScreenData();
            },
            icon: Icon(Icons.add, size: 26.0),
          ),
          // ปุ่ม Toggle Preview
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              togglePreview();
            },
            icon: Icon(_showPreview ? Icons.visibility_off : Icons.visibility, size: 26.0),
            tooltip: _showPreview ? global.language('hide_preview') : global.language('show_preview'),
          ),

          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () async {
              _searchHandlers.searchTrans();
            },
            icon: Icon(Icons.list_alt, size: 26.0),
          ),
          (MediaQuery.of(context).size.width < 800 && _showPreview)
              ? (tabController.index == 0)
                    ? IconButton(
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () {
                        setState(() {
                          tabController.animateTo(1);
                        });
                      },
                      icon: Icon(Icons.file_open, size: 26.0),
                    )
                    : IconButton(
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () {
                        setState(() {
                          tabController.animateTo(0);
                        });
                      },
                      icon: Icon(Icons.text_fields, size: 26.0),
                    )
              : Container(),
          (screenData.posid.isEmpty && screenData.iscancel == false)
              ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
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
                      AppLogger.debug('🖨️ [Save Button] printAfterSave saved to prefs: ${result.printAfterSave}');
                      saveOrUpdateData(taxInvoice: false);
                    }
                  }
                },
                icon: Icon(Icons.save, size: 26.0),
              )
              : Container(),
        ],
      ),
      body: Stack(
        children: [
          LayoutBuilder(
            builder: (context, constraints) {
              return MultiBlocListener(
                listeners: [
                  BlocListener<TransBloc, TransState>(
                    listener: (context, state) {
                      // จัดการ Loading States - ต้องใช้ setState เพื่อให้ UI update
                      if (state is TransInProgress) {
                        if (mounted) setState(() => _isLoading = true);
                      } else if (state is TransSaveInProgress) {
                        AppLogger.debug('🖨️ [TransSaveInProgress] กำลังบันทึก...');
                        if (mounted) setState(() => _isLoading = true);
                      } else if (state is TransUpdateInProgress) {
                        if (mounted) setState(() => _isLoading = true);
                      } else if (state is TransDeleteInProgress) {
                        if (mounted) setState(() => _isLoading = true);
                      } else if (state is TransDeleteManyInProgress) {
                        if (mounted) setState(() => _isLoading = true);
                      } else if (state is TransGetInProgress) {
                        if (mounted) setState(() => _isLoading = true);
                      } else if (state is TransFullInvoiceInProgress) {
                        if (mounted) setState(() => _isLoading = true);
                      }
                      // จัดการ Success/Failed States
                      else if (state is TransLoadSuccess) {
                        if (mounted) {
                          setState(() {
                            _isLoading = false;
                          });
                        }
                        // Handle load success
                      } else if (state is TransLoadFailed) {
                        if (mounted) {
                          setState(() {
                            _isLoading = false;
                          });
                        }
                        context.ui.showError(state.message);
                      } else if (state is TransSaveSuccess) {
                        AppLogger.debug('🖨️ [TransSaveSuccess] ENTERED - docno: ${state.docno}');
                        _isLoading = false;

                        screenData.docno = state.docno;

                        // ส่ง PO เข้าระบบอนุมัติและตรวจสอบว่าเป็น auto_approved หรือไม่
                        // หมายเหตุ: purchasetypecode ถูก save ลง transactionheader ก่อนแล้วใน saveOrUpdateData()
                        _submitPOAndHandlePrint(state.docno, widget.type);

                        // ทำการลบตะกร้าที่ใช้ไปแล้วผ่าน CartWebSocketService
                        if (cartList.isNotEmpty) {
                          // ตรวจสอบว่า service เชื่อมต่ออยู่หรือไม่
                          if (_cartService.isSocketConnected()) {
                            for (String cartId in cartList) {
                              try {
                                // ใช้ฟังก์ชัน deleteCart ของ CartWebSocketService
                                _cartService.deleteCart(cartId, global.getShopId());

                                if (kDebugMode) {
                                  AppLogger.debug('ลบตะกร้า ID: $cartId ผ่าน CartWebSocketService');
                                }
                              } catch (e) {
                                if (kDebugMode) {
                                  AppLogger.debug('เกิดข้อผิดพลาดในการลบตะกร้า: $e');
                                }
                              }
                            }
                          } else {
                            // กรณี service ไม่ได้เชื่อมต่อ ให้ทำการเชื่อมต่อก่อน
                            _cartService.connect(context).then((connected) {
                              if (connected) {
                                for (String cartId in cartList) {
                                  _cartService.deleteCart(cartId, global.getShopId());
                                }
                              }
                            });
                          }

                          // เคลียร์รายการตะกร้าหลังจากลบเสร็จ
                          cartList.clear();
                        }
                        if (mounted) {
                          setState(() {});
                        }

                        // ใช้ async helper เพื่ออ่านค่าจาก SharedPreferences และจัดการพิมพ์
                        _handleSaveSuccessWithPrintCheck(state.docno, widget.type);

                        // Refresh list ด้านซ้ายหลังจาก save สำเร็จ โดยรักษาตำแหน่งเดิม
                        // Backend อาจใช้เวลาประมวลผลนาน (อัพเดท isref ฯลฯ)
                        // ใช้ delay สั้นๆ ก่อน แล้วปล่อยให้ auto-refresh (ทุก 5 วินาที) จัดการต่อ
                        if (_showDocSearchPanel) {
                          final guidToSelect = screenData.guidfixed;
                          // แสดง loading animation บน card ทันที
                          if (guidToSelect != null && _transSearchKey.currentState != null) {
                            _transSearchKey.currentState!.setCardLoadingState(guidToSelect);
                          }
                          Future.delayed(const Duration(milliseconds: 2000), () {
                            if (mounted && _transSearchKey.currentState != null) {
                              _transSearchKey.currentState!.refreshListAndMaintainPosition(
                                selectDocGuid: guidToSelect,
                              );
                            }
                          });
                        }
                      } else if (state is TransSaveFailed) {
                        if (mounted) {
                          setState(() {
                            _isLoading = false;
                          });
                        }
                        // ตรวจสอบว่าเป็น duplicate key error หรือไม่ (docno ซ้ำ)
                        final isDuplicateKeyError = state.message.toLowerCase().contains('duplicate key') ||
                            state.message.contains('E11000') ||
                            state.message.toLowerCase().contains('docno') && state.message.toLowerCase().contains(global.language('duplicate_short'));

                        if (isDuplicateKeyError) {
                          // แสดง dialog ให้ผู้ใช้เลือกสร้าง docno ใหม่
                          _showDuplicateDocNoDialog();
                        } else {
                          context.ui.showError(state.message);
                        }
                      } else if (state is TransUpdateSuccess) {
                        AppLogger.debug('🖨️ [TransUpdateSuccess] ENTERED');
                        setState(() {
                          _isLoading = false;
                        });

                        // ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ
                        // หมายเหตุ: purchasetypecode ถูก save ลง transactionheader ก่อนแล้วใน saveOrUpdateData()
                        _submitPOAndHandleUpdatePrint(widget.type);

                        // Refresh list ด้านซ้ายหลังจาก update สำเร็จ โดยรักษาตำแหน่งเดิม
                        // Backend อาจใช้เวลาประมวลผลนาน (อัพเดท isref ฯลฯ)
                        // ใช้ delay สั้นๆ ก่อน แล้วปล่อยให้ auto-refresh (ทุก 5 วินาที) จัดการต่อ
                        if (_showDocSearchPanel) {
                          final guidToSelect = screenData.guidfixed;
                          // แสดง loading animation บน card ทันที
                          if (guidToSelect != null && _transSearchKey.currentState != null) {
                            _transSearchKey.currentState!.setCardLoadingState(guidToSelect);
                          }
                          Future.delayed(const Duration(milliseconds: 2000), () {
                            if (mounted && _transSearchKey.currentState != null) {
                              _transSearchKey.currentState!.refreshListAndMaintainPosition(
                                selectDocGuid: guidToSelect,
                              );
                            }
                          });
                        }
                      } else if (state is TransUpdateFailed) {
                        setState(() {
                          _isLoading = false;
                        });
                        context.ui.showError(state.message);
                      } else if (state is TransDeleteSuccess) {
                        // เก็บ guid ของเอกสารที่ลบไว้ก่อน clear screen
                        String deletedGuid = screenData.guidfixed ?? '';

                        clearScreenData();
                        setState(() {
                          _isLoading = false;
                        });
                        context.ui.showSuccess(global.language("delete_success"));

                        // Refresh list ด้านซ้ายหลังจาก delete สำเร็จ โดยเลือก item ถัดไป
                        // ใช้ Future.delayed เพื่อให้ delete flow เสร็จสมบูรณ์ก่อน
                        if (_showDocSearchPanel && deletedGuid.isNotEmpty) {
                          Future.delayed(const Duration(milliseconds: 2000), () {
                            if (mounted && _transSearchKey.currentState != null) {
                              _transSearchKey.currentState!.refreshListAfterDelete(deletedGuid);
                            }
                          });
                        }
                      } else if (state is TransDeleteFailed) {
                        setState(() {
                          _isLoading = false;
                        });
                        context.ui.showError(state.message);
                      } else if (state is TransDeleteManySuccess) {
                        setState(() {
                          _isLoading = false;
                        });
                        context.ui.showSuccess(global.language('delete_multiple_success'));
                      } else if (state is TransDeleteManyFailed) {
                        setState(() {
                          _isLoading = false;
                        });
                        context.ui.showError(global.language('delete_multiple_failed'));
                      } else if (state is TransGetSuccess) {
                        setState(() {
                          _isLoading = false;
                        });
                        // Handle get success
                        screenData = state.trans;
                        screenDataTemp = TransactionModel.fromJson(state.trans.toJson());

                        // Log ข้อมูลสกุลเงินที่ได้จาก API (ก่อนเรียก _loadCurrencies)
                        AppLogger.info('[PO] TransGetSuccess - docCurrency: ${screenData.docCurrency}, '
                            'docCurrencySymbol: ${screenData.docCurrencySymbol}, '
                            'currency: ${screenData.currency}, '
                            'exchangerate: ${screenData.exchangerate}');

                        // โหลดประเภทการจัดซื้อสำหรับ PO ที่มีอยู่แล้ว
                        if (widget.type == global.TransactionTypeEnum.purchaseorder && screenData.docno.isNotEmpty) {
                          // โหลด list ประเภทการจัดซื้อก่อน แล้วค่อยโหลดค่าที่บันทึกไว้
                          _loadPurchaseTypesAndExisting(screenData.docno);
                          _loadCurrencies(); // โหลดรายการสกุลเงิน
                          // โหลดสถานะการอนุมัติของ PO
                          _loadPOApprovalStatus(screenData.docno);
                        }
                      } else if (state is TransGetFailed) {
                        setState(() {
                          _isLoading = false;
                        });
                        context.ui.showError(state.message);
                      } else if (state is TransFullInvoiceSuccess) {
                        if (mounted) {
                          setState(() {
                            _isLoading = false;
                          });
                        }
                        showSaveDocNoDialog(context, state.docno);
                        context.ui.showSuccess(global.language('create_full_tax_invoice_success'));
                      } else if (state is TransFullInvoiceFailed) {
                        setState(() {
                          _isLoading = false;
                        });
                        context.ui.showError('${global.language("create_full_tax_invoice_failed")}: ${state.message}');
                      }
                      // Reset loading สำหรับ states อื่นๆ ที่ไม่ได้จัดการ
                      else {
                        setState(() {
                          _isLoading = false;
                        });
                      }
                    },
                  ),
                  BlocListener<WarehouseBloc, WarehouseState>(
                    listener: (context, state) {
                      if (state is WarehouseLoadSuccess) {
                        warehouseList = state.warehouses;
                        if (mounted) {
                          if (warehouseList.isNotEmpty) {
                            setState(() {
                              if (warehouseList.isNotEmpty) {
                                /// จากคลัง
                                defaultWarehouse = warehouseList[0].code;
                                defaultWarehouseNames = warehouseList[0].names;

                                /// ไปคลัง
                                defaultToWarehouse = warehouseList[0].code;
                                defaultToWarehouseNames = warehouseList[0].names;

                                if (warehouseList[0].location.isNotEmpty) {
                                  /// จากที่เก็บ
                                  defaultLocation = warehouseList[0].location[0].code;
                                  defaultLocationNames = warehouseList[0].location[0].names;

                                  ///
                                  defaultToLocation = warehouseList[0].location[0].code;
                                  defaultToLocationNames = warehouseList[0].location[0].names;
                                }
                              }
                            });
                          }
                        }
                      }
                    },
                  ),
                  BlocListener<ExportCsvBloc, ExportCsvState>(
                    listener: (context, state) {
                      /// download csv
                      if (state is SaleInvoiceExportInProgress) {
                        // Show a loading indicator if desired
                      } else if (state is SaleInvoiceExportSuccess) {
                        global.showSnackBar(context, Icon(Icons.save, color: global.theme.onPrimaryColor), global.language("export_success"), global.theme.infoHighlightTextColor);
                      } else if (state is SaleInvoiceExportFailed) {
                        // Show an error message
                        global.showSnackBar(context, Icon(Icons.save, color: global.theme.onPrimaryColor), "${global.language("not_export_success")} : ${state.message}", global.theme.negativeHighlightTextColor);
                      }
                    },
                  ),
                ],
                child: _buildMainContent(constraints),
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
