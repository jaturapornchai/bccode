import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:intl/intl.dart';
import 'package:smlaicloud/bloc/trans/trans_bloc.dart';
import 'package:smlaicloud/bloc/creditor_filter/creditor_filter_cubit.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/doc_payload_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/creditor_repository.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/services/pdf_service.dart';

class TransSearchScreen extends StatefulWidget {
  const TransSearchScreen({super.key, required this.type, required this.custcode, this.onLoadingStateChanged, this.isEmbedded = false, this.onDocumentSelected, this.onClose});
  final TransactionTypeEnum type;
  final String custcode;
  final Function(bool isLoading)? onLoadingStateChanged; // เพิ่ม callback สำหรับแจ้ง loading state

  /// ถ้า true จะแสดงเป็น embedded widget (ไม่ใช้ Scaffold)
  /// ใช้สำหรับแสดงเป็น side panel ในจอใหญ่
  final bool isEmbedded;

  /// callback เมื่อเลือกเอกสาร (ใช้ในโหมด embedded)
  final Function(TransactionModel document)? onDocumentSelected;

  /// callback เมื่อกดปิด panel (ใช้ในโหมด embedded)
  final VoidCallback? onClose;

  @override
  State<TransSearchScreen> createState() => TransSearchScreenState();
}

class TransSearchScreenState extends State<TransSearchScreen> with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();

  // Responsive utility functions
  bool isMobile(double width) => width < 600;
  bool isTablet(double width) => width >= 600 && width < 1024;
  bool isDesktop(double width) => width >= 1024;

  double getResponsiveFontSize(double width, {double mobile = 12, double tablet = 14, double desktop = 16}) {
    if (isMobile(width)) return mobile;
    if (isTablet(width)) return tablet;
    return desktop;
  }

  double getResponsivePadding(double width, {double mobile = 8, double tablet = 12, double desktop = 16}) {
    if (isMobile(width)) return mobile;
    if (isTablet(width)) return tablet;
    return desktop;
  }

  int getTableFlexRatio(double width, int defaultFlex) {
    if (isMobile(width)) return (defaultFlex / 2).ceil();
    return defaultFlex;
  }

  String searchText = "";
  List<TransactionModel> transListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  final int _hoverIndex = -1;
  int currentListIndex = 0;
  final _debouncer = global.Debouncer(1000);
  int filterTransaction = 1;
  bool loadingData = false;
  bool isSearching = false; // เพิ่ม flag สำหรับตรวจสอบว่ากำลัง search อยู่หรือไม่
  String pendingSearchText = ""; // เก็บ search text ที่กำลังรอ
  String currentSearchRequest = ""; // เก็บ search text ที่ส่ง request ไปแล้ว
  int totalRecord = 0;
  int dateOrderByType = 1; // 1=วันที่ล่าสุด, 2=วันที่เก่าสุด

  // === ตัวแปรสำหรับการกรองข้อมูล (client-side) ===
  DateTime? _filterFromDate;
  DateTime? _filterToDate;
  double? _filterMinAmount;
  double? _filterMaxAmount;
  List<String> _filterCreditorCodes = []; // กรองตามเจ้าหนี้ (multi-select)
  bool _showAdvancedFilter = false; // แสดง/ซ่อน panel กรองขั้นสูง

  // Cubit สำหรับจัดการรายชื่อเจ้าหนี้ (ใช้ร่วมกับ BLoC หลักเพื่อเพิ่มประสิทธิภาพ)
  late CreditorFilterCubit _creditorFilterCubit;

  // Controller สำหรับค้นหาเจ้าหนี้
  final TextEditingController _creditorSearchController = TextEditingController();
  Timer? _creditorSearchDebounce; // debounce timer สำหรับค้นหาเจ้าหนี้

  // Controllers สำหรับ TextField ยอดเงิน
  final TextEditingController _minAmountController = TextEditingController();
  final TextEditingController _maxAmountController = TextEditingController();

  // === ขนาด font ปรับได้ ===
  double _customFontSize = 12.0;
  static const double _minFontSize = 10.0;
  static const double _maxFontSize = 16.0;

  // Auto-refresh variables สำหรับ sync ข้อมูลจากเครื่องอื่น
  Timer? _autoRefreshTimer;
  static const int _baseRefreshInterval = 5; // เริ่มต้น 5 วินาที
  int _currentRefreshInterval = _baseRefreshInterval; // interval ปัจจุบัน
  bool _isAutoRefreshing = false; // flag บอกว่ากำลัง auto-refresh อยู่
  String _lastDataHash = ''; // hash ของข้อมูลล่าสุดเพื่อเปรียบเทียบ

  // Loading state for card animation during save/refresh
  String _pendingLoadingGuid = ''; // guid ของ card ที่กำลังรอ refresh

  // สถานะการอนุมัติ PO (เฉพาะ purchaseorder)
  Map<String, POApprovalStatusInfo> _approvalStatusMap = {};

  // Timer สำหรับ polling สถานะอนุมัติทุก 5 วินาที (เฉพาะ purchaseorder)
  Timer? _approvalStatusPollingTimer;
  static const int _approvalPollingInterval = 5; // วินาที

  TransactionModel emptyResult = TransactionModel(
    shopid: global.apiShopCode,
    guidref: '',
    docno: '',
    docdatetime: '',
    docrefno: '',
    docrefdate: '',
    docreftype: 0,
    doctype: 0,
    vattype: 0,
    custcode: '',
    custnames: [],
    salecode: '',
    salename: '',
    discountword: '',
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
    vatrate: 0,
    status: 0,
    inquirytype: 0,
    taxdocdate: '',
    taxdocno: '',
    totalbeforevat: 0,
    transflag: 0,
    isref: false,
    islocked: false,
    isclosed: false,
    iscancel: false,
    detailcount: 0,
    ismanualamount: false,
    description: '',
    details: <TransactionDetailModel>[],
    paymentdetailraw: "",
    paymentStructs: [],
  );

  List<Widget> tableHeader = [];

  void reloadDataFromMongoDbAndReturnSelected(TransactionModel selected) async {
    // Update selectGuid เพื่อให้ highlight item ที่เลือก
    setState(() {
      selectGuid = selected.guidfixed ?? '';
    });

    // ดึง collection name ตาม transaction type
    String collection = _getCollectionNameForType(widget.type);

    // ถ้าไม่มี collection name ให้ใช้ข้อมูลเดิม
    if (collection.isEmpty) {
      if (widget.isEmbedded && widget.onDocumentSelected != null) {
        widget.onDocumentSelected!(selected);
      } else if (mounted) {
        Navigator.pop(context, selected);
      }
      return;
    }

    // Query ข้อมูลเต็มจาก MongoDB
    try {
      MongoGetDataPayloadModel docPayLoad = MongoGetDataPayloadModel(shopid: global.getShopId(), collection: collection, guidfixed: selected.guidfixed!);
      final response = await global.goApiPost(global.goApiUrlPath("mongogetdata"), docPayLoad.toJson());

      String rawData = json.encoder.convert(response);
      GoApiQueryOneResultResponse data = GoApiQueryOneResultResponse.fromMap(json.decode(rawData));

      // แปลง data.result (Map<String, dynamic>) เป็น TransactionModel
      TransactionModel trans = TransactionModel.fromJson(data.result);

      // Parse JSON data ของ transaction
      global.parseTransactionJsonData(trans);

      if (widget.isEmbedded && widget.onDocumentSelected != null) {
        widget.onDocumentSelected!(trans);
      } else if (mounted) {
        Navigator.pop(context, trans);
      }
    } catch (ex) {
      // ถ้า query ไม่สำเร็จ ใช้ข้อมูลเดิม
      if (widget.isEmbedded && widget.onDocumentSelected != null) {
        widget.onDocumentSelected!(selected);
      } else if (mounted) {
        Navigator.pop(context, selected);
      }
    }
  }

  /// นำทางไปหน้าพิมพ์ PO (ใช้ PdfService ที่มี theme selection)
  /// ใช้สำหรับกรณี PO ที่อนุมัติแล้ว สามารถพิมพ์ได้จากหน้ารายการ
  Future<void> _navigateToPrintPO(TransactionModel selected) async {
    if (!mounted) return;

    // ใช้ PdfService เพื่อเปิดหน้าพิมพ์ที่มีการเลือกธีม
    const String collection = "transactionPurchaseOrder";
    final String docNo = selected.docno;
    final String title = global.language("purchase_order");

    try {
      PdfService pdfService = PdfService();
      await pdfService.generateAndOpenPdf(
        context: context,
        collection: collection,
        docNo: docNo,
        title: title,
        // ตรวจ 2 ระดับ: 1) docCurrency ต่างจาก baseCurrency 2) exchangerate != 1.0
        isMultiCurrency: (selected.docCurrency != null &&
                selected.docCurrency!.isNotEmpty &&
                selected.docCurrency!.toUpperCase() != global.getBaseCurrency()) ||
            (selected.exchangerate != null &&
                selected.exchangerate != 0 &&
                selected.exchangerate != 1.0),
      );
    } catch (ex) {
      // แสดง error (PdfService จัดการ loading เอง)
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("cannot_open_print_page")}: $ex');
      }
    }
  }

  /// Helper method สำหรับดึงชื่อ collection ตามประเภทเอกสาร
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
      case global.TransactionTypeEnum.advancePaymentRefund:
        return "transactionAdvancePaymentRefund";
      case global.TransactionTypeEnum.deposit:
        return "transactionDeposit";
      case global.TransactionTypeEnum.depositRefund:
        return "transactionDepositRefund";
      case global.TransactionTypeEnum.paidAdvance:
        return "transactionPaidAdvance";
      case global.TransactionTypeEnum.paidAdvanceRefund:
        return "transactionPaidAdvanceRefund";
      case global.TransactionTypeEnum.receiveDeposit:
        return "transactionReceiveDeposit";
      case global.TransactionTypeEnum.receiveDepositRefund:
        return "transactionReceiveDepositRefund";
      case global.TransactionTypeEnum.accrualreceive:
        return "transactionAccrualReceive";
      default:
        return "";
    }
  }

  List<Widget> buildTableHeader(double screenWidth) {
    final headerFontSize = getResponsiveFontSize(screenWidth, mobile: 11, tablet: 13, desktop: 14);
    final headerPadding = getResponsivePadding(screenWidth, mobile: 5, tablet: 8, desktop: 10);
    List<Widget> headers = [];

    headers.add(
      Expanded(
        flex: getTableFlexRatio(screenWidth, 1),
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
          child: Text(
            global.language("doc_date"),
            style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
          ),
        ),
      ),
    );

    headers.add(
      Expanded(
        flex: getTableFlexRatio(screenWidth, 1),
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
          child: Text(
            global.language("docno"),
            style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
          ),
        ),
      ),
    );

    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.stockbalance) {
      headers.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 3),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
            child: Text(
              (widget.type == TransactionTypeEnum.purchase ||
                      widget.type == TransactionTypeEnum.purchasereturn ||
                      widget.type == TransactionTypeEnum.accrualreceive ||
                      widget.type == TransactionTypeEnum.purchasepartial ||
                      widget.type == TransactionTypeEnum.purchaseorder ||
                      widget.type == TransactionTypeEnum.advancePayment)
                  ? global.language("supplier")
                  : global.language("customer"),
              style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
              textAlign: TextAlign.left,
            ),
          ),
        ),
      );
    }

    // เพิ่ม headers อื่นๆ ต่อไป...
    if (widget.type == global.TransactionTypeEnum.purchaseorder) {
      headers.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
            child: Text(
              textAlign: TextAlign.left,
              global.language("status"),
              style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
            ),
          ),
        ),
      );
    } else {
      if (widget.type != global.TransactionTypeEnum.stockbalance) {
        headers.add(
          Expanded(
            flex: getTableFlexRatio(screenWidth, 1),
            child: Padding(
              padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
              child: Text(
                textAlign: TextAlign.left,
                global.language("sale_name"),
                style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
              ),
            ),
          ),
        );
      }
    }

    // เพิ่ม headers ส่วนที่เหลือ...
    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.stockbalance) {
      headers.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
            child: Text(
              textAlign: TextAlign.left,
              global.language("inquiry_type"),
              style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
            ),
          ),
        ),
      );
    }

    if (widget.type == global.TransactionTypeEnum.adjust) {
      headers.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
            child: Text(
              textAlign: TextAlign.left,
              global.language("adjust_type"),
              style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
            ),
          ),
        ),
      );
    }

    headers.add(
      Expanded(
        flex: getTableFlexRatio(screenWidth, 1),
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
          child: Text(
            textAlign: TextAlign.right,
            global.language("product_list"),
            style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
          ),
        ),
      ),
    );

    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct) {
      headers.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: headerPadding / 2),
            child: Text(
              textAlign: TextAlign.right,
              global.language("total_value"),
              style: TextStyle(color: global.theme.textColor, fontWeight: FontWeight.bold, fontSize: headerFontSize),
            ),
          ),
        ),
      );
    }

    return headers;
  }

  Widget buildMobileCard(TransactionModel value, int index, DateTime docDateTime, double fontSize, ValueNotifier<bool> isHovered) {
    // กำหนดสีพื้นหลัง Card
    Color getCardColor() {
      if (selectGuid == value.guidfixed) {
        return global.theme.rowSelectedColor; // สีเมื่อ selected
      } else if (index % 2 == 0) {
        return global.theme.columnAlternateEvenColor; // บรรทัดคู่
      } else {
        return global.theme.columnAlternateOddColor; // บรรทัดคี่
      }
    }

    return ValueListenableBuilder<bool>(
      valueListenable: isHovered,
      builder: (context, hovered, child) {
        return MouseRegion(
          onEnter: (_) => isHovered.value = true,
          onExit: (_) => isHovered.value = false,
          child: GestureDetector(
            onTap: () {
              reloadDataFromMongoDbAndReturnSelected(value);
            },
            child: Container(
              margin: EdgeInsets.symmetric(horizontal: 4, vertical: 2),
              child: Stack(
                children: [
                  Card(
                    elevation: hovered ? 2 : 1,
                    shadowColor: global.theme.dividerBorderColor.withValues(alpha: 0.2),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(6),
                      side: BorderSide(color: selectGuid == value.guidfixed ? global.theme.infoHighlightTextColor : Colors.transparent, width: 1.5),
                    ),
                    color: getCardColor(),
                    child: Padding(
                      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                      child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      // Row 1: วันที่ + เลขที่เอกสาร + ปุ่มพิมพ์ (สำหรับ PO ที่อนุมัติแล้ว)
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          // ปุ่มพิมพ์สำหรับ PO ที่อนุมัติแล้ว (แสดงทางซ้ายสุด)
                          if (widget.type == global.TransactionTypeEnum.purchaseorder)
                            Builder(
                              builder: (context) {
                                final approvalStatus = _approvalStatusMap[value.docno];
                                final canPrint = approvalStatus != null &&
                                    (approvalStatus.status == 'approved' || approvalStatus.status == 'auto_approved');
                                if (canPrint) {
                                  return Padding(
                                    padding: const EdgeInsets.only(right: 8),
                                    child: InkWell(
                                      onTap: () async {
                                        // โหลดข้อมูลเต็มจาก MongoDB แล้วไปหน้าพิมพ์
                                        await _navigateToPrintPO(value);
                                      },
                                      borderRadius: BorderRadius.circular(4),
                                      child: Container(
                                        padding: const EdgeInsets.all(4),
                                        decoration: BoxDecoration(
                                          color: global.theme.positiveHighlightColor,
                                          borderRadius: BorderRadius.circular(4),
                                          border: Border.all(color: global.theme.positiveHighlightTextColor),
                                        ),
                                        child: Icon(
                                          Icons.print,
                                          size: 16,
                                          color: global.theme.positiveHighlightTextColor,
                                        ),
                                      ),
                                    ),
                                  );
                                }
                                return const SizedBox.shrink();
                              },
                            ),
                          Expanded(
                            child: Text(
                              global.formatThaiDateTime(dateTime: docDateTime.toLocal(), showTime: true),
                              style: TextStyle(fontSize: fontSize - 1, fontWeight: FontWeight.w600, color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.infoHighlightTextColor),
                            ),
                          ),
                          Text(
                            value.docno,
                            style: TextStyle(fontSize: fontSize - 1, fontWeight: FontWeight.bold, color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor),
                          ),
                        ],
                      ),

                      // Row 2: ชื่อลูกค้า (แสดงเฉพาะบาง transaction type)
                      if (widget.type != global.TransactionTypeEnum.adjust &&
                          widget.type != global.TransactionTypeEnum.stocktransfer &&
                          widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
                          widget.type != global.TransactionTypeEnum.stockpickupproduct &&
                          widget.type != global.TransactionTypeEnum.stockreturnproduct &&
                          widget.type != global.TransactionTypeEnum.stockbalance)
                        Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Text(
                            "${global.activeLangName(value.custnames ?? [])} (${value.custcode})",
                            style: TextStyle(fontSize: fontSize - 1, color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),

                      // Row 3: ประเภท + จำนวนรายการ + ยอดรวม
                      Padding(
                        padding: EdgeInsets.only(top: 3),
                        child: Row(
                          children: [
                            // ประเภท (เครดิต/เงินสด)
                            if (widget.type != global.TransactionTypeEnum.adjust &&
                                widget.type != global.TransactionTypeEnum.stocktransfer &&
                                widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
                                widget.type != global.TransactionTypeEnum.stockpickupproduct &&
                                widget.type != global.TransactionTypeEnum.stockreturnproduct &&
                                widget.type != global.TransactionTypeEnum.stockbalance)
                              Container(
                                padding: EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                                margin: EdgeInsets.only(right: 6),
                                decoration: BoxDecoration(color: value.inquirytype == 0 ? global.theme.warningHighlightColor : global.theme.infoHighlightColor, borderRadius: BorderRadius.circular(3)),
                                child: Text(
                                  value.inquirytype == 0 ? global.language('credit') : global.language('cash'),
                                  style: TextStyle(fontSize: fontSize - 2, color: value.inquirytype == 0 ? global.theme.warningHighlightTextColor : global.theme.infoHighlightTextColor),
                                ),
                              ),

                            // จำนวนรายการ
                            Icon(Icons.list_alt, size: 12, color: global.theme.iconSecondaryColor),
                            SizedBox(width: 2),
                            Text(
                              widget.type == global.TransactionTypeEnum.purchaseorder
                                  ? value.detailcount.toString()
                                  : widget.type == global.TransactionTypeEnum.stockbalance
                                  ? global.formatNumber(value.totalqty!)
                                  : (value.details?.length ?? value.detailcount).toString(),
                              style: TextStyle(fontSize: fontSize - 2, color: global.theme.iconSecondaryColor),
                            ),

                            Spacer(),

                            // ยอดรวม (รวม roundamount ให้ตรงกับยอดรวมทั้งสิ้นของเอกสาร)
                            if (widget.type != global.TransactionTypeEnum.adjust &&
                                widget.type != global.TransactionTypeEnum.stocktransfer &&
                                widget.type != global.TransactionTypeEnum.stockpickupproduct &&
                                widget.type != global.TransactionTypeEnum.stockreturnproduct)
                              Text(
                                global.formatNumber(value.totalamount != 0 ? value.totalamount + (value.roundamount ?? 0) : value.totalvalue.toDouble()),
                                style: TextStyle(fontSize: fontSize, fontWeight: FontWeight.bold, color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.positiveHighlightTextColor),
                              ),
                          ],
                        ),
                      ),

                      // Row 4: แสดงผู้สร้างเอกสาร (เฉพาะ PO ที่มีข้อมูล) พร้อมวันที่เอกสาร
                      if (widget.type == global.TransactionTypeEnum.purchaseorder &&
                          (value.creatorname?.isNotEmpty ?? false))
                        Padding(
                          padding: const EdgeInsets.only(top: 4),
                          child: Row(
                            children: [
                              Icon(Icons.person_outline,
                                  size: fontSize - 2, color: global.theme.iconSecondaryColor),
                              const SizedBox(width: 4),
                              Text(
                                'สร้างโดย: ${value.creatorname}',
                                style: TextStyle(
                                  color: global.theme.iconSecondaryColor,
                                  fontSize: fontSize - 2,
                                  fontStyle: FontStyle.italic,
                                ),
                              ),
                              const SizedBox(width: 6),
                              // แสดงวันที่เอกสาร (docdatetime) ไม่ใช่วันที่สร้าง (created_at)
                              Text(
                                global.formatThaiDateTime(dateTime: docDateTime.toLocal(), showTime: true),
                                style: TextStyle(
                                  color: global.theme.iconSecondaryColor,
                                  fontSize: fontSize - 2,
                                ),
                              ),
                            ],
                          ),
                        ),

                      // แสดงสถานะการอนุมัติสำหรับ PO (Mobile view)
                      if (widget.type == global.TransactionTypeEnum.purchaseorder)
                        Builder(
                          builder: (context) {
                            final approvalStatus = _approvalStatusMap[value.docno];
                            // มีสถานะอนุมัติ - แสดงสถานะจริง
                            if (approvalStatus != null && approvalStatus.statusText.isNotEmpty) {
                              return Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Container(
                                    margin: const EdgeInsets.only(top: 6),
                                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                    decoration: BoxDecoration(
                                      color: Color(approvalStatus.statusColorValue).withValues(alpha: 0.15),
                                      borderRadius: BorderRadius.circular(4),
                                      border: Border.all(color: Color(approvalStatus.statusColorValue), width: 0.5),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(
                                          approvalStatus.status == 'approved' ? Icons.check_circle :
                                          approvalStatus.status == 'rejected' ? Icons.cancel :
                                          approvalStatus.status == 'pending' ? Icons.hourglass_empty :
                                          Icons.auto_awesome, // auto_approved
                                          size: 14,
                                          color: Color(approvalStatus.statusColorValue),
                                        ),
                                        const SizedBox(width: 4),
                                        Text(
                                          approvalStatus.statusText,
                                          style: TextStyle(
                                            color: Color(approvalStatus.statusColorValue),
                                            fontWeight: FontWeight.bold,
                                            fontSize: fontSize - 1,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  // แสดงสถานะล่าสุด (latest action)
                                  if (approvalStatus.latestActionText != null && approvalStatus.latestActionText!.isNotEmpty)
                                    Padding(
                                      padding: const EdgeInsets.only(top: 4, left: 2),
                                      child: Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          Icon(
                                            _getLatestActionIcon(approvalStatus.latestActionText!),
                                            size: 12,
                                            color: global.theme.textSecondaryColor,
                                          ),
                                          const SizedBox(width: 4),
                                          Text(
                                            _formatLatestAction(approvalStatus),
                                            style: TextStyle(
                                              color: global.theme.textSecondaryColor,
                                              fontSize: fontSize - 2,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                ],
                              );
                            }
                            // ไม่มีสถานะอนุมัติและยังไม่ถูกยกเลิก - แสดง "ยังไม่ส่งขออนุมัติ"
                            if (!value.iscancel) {
                              return Container(
                                margin: const EdgeInsets.only(top: 6),
                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                decoration: BoxDecoration(
                                  color: global.theme.dividerBorderColor,
                                  borderRadius: BorderRadius.circular(4),
                                  border: Border.all(color: global.theme.iconSecondaryColor, width: 0.5),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Icon(Icons.schedule, size: 14, color: global.theme.textSecondaryColor),
                                    SizedBox(width: 4),
                                    Text(
                                      global.language('not_yet_submitted'),
                                      style: TextStyle(
                                        color: global.theme.textSecondaryColor,
                                        fontWeight: FontWeight.bold,
                                        fontSize: fontSize - 1,
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            }
                            return const SizedBox.shrink();
                          },
                        ),

                      // Status for Purchase Order
                      if (widget.type == global.TransactionTypeEnum.purchaseorder && (value.isref! || value.iscancel))
                        Container(
                          margin: EdgeInsets.only(top: 8),
                          padding: EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: global.theme.rowHoverColor,
                            borderRadius: BorderRadius.circular(6),
                            border: Border.all(color: global.theme.primaryColor, width: 1),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              if (value.isref!) ...[
                                Row(
                                  children: [
                                    Icon(Icons.link, size: 16, color: global.theme.primaryColor),
                                    SizedBox(width: 4),
                                    Text(
                                      global.language("referenced"),
                                      style: TextStyle(color: global.theme.primaryColor, fontWeight: FontWeight.bold, fontSize: fontSize - 1),
                                    ),
                                  ],
                                ),
                                Row(
                                  children: [
                                    Icon(
                                      value.isclosed! ? Icons.check_circle : Icons.pending,
                                      size: 16,
                                      color: value.isclosed! ? global.theme.positiveHighlightTextColor : global.theme.warningHighlightTextColor,
                                    ),
                                    SizedBox(width: 4),
                                    Text(
                                      value.isclosed! ? global.language("fully_received") : global.language("partially_received"),
                                      style: TextStyle(
                                        color: value.isclosed! ? global.theme.positiveHighlightTextColor : global.theme.warningHighlightTextColor,
                                        fontWeight: FontWeight.bold,
                                        fontSize: fontSize - 1,
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                              if (value.iscancel)
                                Row(
                                  children: [
                                    Icon(Icons.cancel, size: 16, color: global.theme.negativeHighlightTextColor ?? global.theme.negativeHighlightTextColor),
                                    SizedBox(width: 4),
                                    Text(
                                      global.language("cancelled"),
                                      style: TextStyle(color: global.theme.negativeHighlightTextColor ?? global.theme.negativeHighlightTextColor, fontWeight: FontWeight.bold, fontSize: fontSize - 1),
                                    ),
                                  ],
                                ),
                            ],
                          ),
                        ),

                      // Adjust Type
                      if (widget.type == global.TransactionTypeEnum.adjust)
                        Container(
                          margin: EdgeInsets.only(top: 8),
                          padding: EdgeInsets.symmetric(vertical: 4, horizontal: 8),
                          decoration: BoxDecoration(
                            color: global.theme.warningHighlightColor ?? global.theme.warningHighlightColor,
                            borderRadius: BorderRadius.circular(6),
                            border: Border.all(color: global.theme.warningHighlightColor ?? global.theme.warningHighlightColor, width: 1),
                          ),
                          child: Text(
                            value.transflag == 66
                                ? global.language('increase')
                                : value.transflag == 68
                                ? global.language('decrease')
                                : global.language('adjust_cost'),
                            style: TextStyle(color: global.theme.warningHighlightTextColor ?? global.theme.warningHighlightTextColor, fontWeight: FontWeight.bold, fontSize: fontSize),
                          ),
                        ),
                    ],
                  ),
                ),
              ),
                  // Loading overlay animation
                  if (_pendingLoadingGuid == value.guidfixed)
                    Positioned.fill(
                      child: Card(
                        elevation: 0,
                        margin: EdgeInsets.zero,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(6),
                        ),
                        color: Colors.transparent,
                        child: Container(
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(6),
                            color: global.theme.infoHighlightColor,
                          ),
                          child: Center(
                            child: SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                valueColor: AlwaysStoppedAnimation<Color>(global.theme.infoHighlightTextColor),
                              ),
                            ),
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  void setSystemLanguageList() async {
    // เพิ่มการ clear transListData
    transListData.clear();
    await global.setSystemLanguage(context);
    loadDataList(searchText, filterTransaction);
  }

  @override
  void initState() {
    super.initState();

    // สร้าง Cubit สำหรับจัดการรายชื่อเจ้าหนี้
    _creditorFilterCubit = CreditorFilterCubit(
      creditorRepository: CreditorRepository(),
    );
    // โหลดรายชื่อเจ้าหนี้ทันทีเพื่อให้พร้อมใช้งาน
    _creditorFilterCubit.loadCreditors();

    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    searchText = '';
    pendingSearchText = ''; // initialize pending search
    currentSearchRequest = ''; // initialize current request
    searchController.text = searchText;

    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          global.language("doc_date"),
          style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
        ),
      ),
    );
    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          global.language("docno"),
          style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
        ),
      ),
    );
    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.stockbalance) {
      tableHeader.add(
        Expanded(
          flex: 3,
          child: Text(
            (widget.type == TransactionTypeEnum.purchase ||
                    widget.type == TransactionTypeEnum.purchasereturn ||
                    widget.type == TransactionTypeEnum.accrualreceive ||
                    widget.type == TransactionTypeEnum.purchasepartial ||
                    widget.type == TransactionTypeEnum.purchaseorder ||
                    widget.type == TransactionTypeEnum.advancePayment)
                ? global.language("supplier")
                : global.language("customer"),
            style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
            textAlign: TextAlign.left,
          ),
        ),
      );
    }
    if (widget.type == global.TransactionTypeEnum.purchaseorder) {
      tableHeader.add(
        Expanded(
          flex: 1,
          child: Text(
            textAlign: TextAlign.left,
            global.language("status"),
            style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
          ),
        ),
      );
    } else {
      if (widget.type != global.TransactionTypeEnum.stockbalance) {
        tableHeader.add(
          Expanded(
            flex: 1,
            child: Text(
              textAlign: TextAlign.left,
              global.language("sale_name"),
              style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
            ),
          ),
        );
      }
    }
    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.stockbalance) {
      tableHeader.add(
        Expanded(
          flex: 1,
          child: Text(
            textAlign: TextAlign.left,
            global.language("inquiry_type"),
            style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
          ),
        ),
      );
    }
    if (widget.type == global.TransactionTypeEnum.adjust) {
      tableHeader.add(
        Expanded(
          flex: 1,
          child: Text(
            textAlign: TextAlign.left,
            global.language("adjust_type"),
            style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
          ),
        ),
      );
    }
    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          textAlign: TextAlign.right,
          global.language("product_list"),
          style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
        ),
      ),
    );
    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct) {
      tableHeader.add(
        Expanded(
          flex: 1,
          child: Text(
            textAlign: TextAlign.right,
            global.language("total_value"),
            style: TextStyle(color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
          ),
        ),
      );
    }
  }

  void loadDataList(String search, filterTransaction) {
    // ตรวจสอบว่า widget ยังคง mounted หรือไม่
    if (!mounted) return;

    // ป้องกัน race condition - ถ้ากำลัง search อยู่แล้วและ search text เหมือนกัน ไม่ต้องทำอะไร
    if (loadingData && searchText == search) {
      return;
    }

    // เก็บ request ที่กำลังจะส่ง
    currentSearchRequest = search;

    setState(() {
      loadingData = true;
      // เพิ่มการ clear transListData เมื่อเป็นการค้นหาใหม่
      if (searchText != search) {
        transListData.clear();
      }
    });

    // แจ้ง parent widget ว่าเริ่ม loading แล้ว
    if (widget.onLoadingStateChanged != null) {
      widget.onLoadingStateChanged!(true);
    }

    searchText = search;

    context.read<TransBloc>().add(
      TransLoad(
        offset: (transListData.isEmpty) ? 0 : transListData.length,
        limit: 100,
        search: search,
        type: widget.type,
        custcode: widget.custcode,
        dateorder: dateOrderByType,
        ispos: (filterTransaction == 2)
            ? "true"
            : (filterTransaction == 3)
            ? "false"
            : "null",
      ),
    );
  }

  void _moveCursorToTextEnd() {
    if (!mounted) return;
    final textLength = searchController.text.length;
    searchController.selection = TextSelection.fromPosition(TextPosition(offset: textLength));
  }

  void onScrollList() {
    if (listScrollController.offset >= listScrollController.position.maxScrollExtent && !listScrollController.position.outOfRange) {
      loadDataList(searchText, filterTransaction);
    }
  }

  // ตัวแปรสำหรับเก็บตำแหน่งเพื่อ restore หลังจาก refresh
  String _pendingRestoreGuid = '';
  double _pendingRestoreScrollOffset = 0;
  bool _isRefreshing = false;

  /// Public method สำหรับเซ็ต loading state บน card ก่อนที่จะ refresh
  /// เรียกใช้จาก parent widget ผ่าน GlobalKey เพื่อแสดง animation ระหว่างรอ
  void setCardLoadingState(String guid) {
    if (!mounted) return;
    setState(() {
      _pendingLoadingGuid = guid;
    });
  }

  /// Public method สำหรับเคลียร์ loading state
  void clearCardLoadingState() {
    if (!mounted) return;
    setState(() {
      _pendingLoadingGuid = '';
    });
  }

  /// Public method สำหรับ refresh list หลังจาก save โดยรักษาตำแหน่งเดิม
  /// เรียกใช้จาก parent widget ผ่าน GlobalKey
  void refreshListAndMaintainPosition({String? selectDocGuid}) {
    if (!mounted) return;

    // Reset auto-refresh interval กลับไป 5 วินาที (เพราะมีการ save/update)
    _currentRefreshInterval = _baseRefreshInterval;

    // เก็บตำแหน่ง scroll และ guid ที่เลือกอยู่
    _pendingRestoreScrollOffset = listScrollController.hasClients
        ? listScrollController.offset
        : 0;
    _pendingRestoreGuid = selectDocGuid ?? selectGuid;
    _isRefreshing = true;

    // ไม่ clear list เพื่อให้ smooth - โหลดข้อมูลใหม่มาก่อนแล้วค่อย replace
    // เรียก _loadDataListForRefresh เพื่อโหลดข้อมูลใหม่ตั้งแต่ต้น
    _loadDataListForRefresh();
  }

  /// Helper method สำหรับโหลดข้อมูลใหม่ตั้งแต่ต้น (สำหรับ refresh)
  /// ไม่ clear list ก่อน เพื่อให้ UI smooth ไม่กระพริบ
  void _loadDataListForRefresh() {
    if (!mounted) return;

    setState(() {
      loadingData = true;
    });

    // โหลดข้อมูลใหม่ตั้งแต่ต้น (offset = 0)
    context.read<TransBloc>().add(
      TransLoad(
        offset: 0,
        limit: 100,
        search: searchText,
        type: widget.type,
        custcode: widget.custcode,
        dateorder: dateOrderByType,
        ispos: (filterTransaction == 2)
            ? "true"
            : (filterTransaction == 3)
            ? "false"
            : "null",
      ),
    );
  }

  /// Public method สำหรับ refresh list หลังจาก delete
  /// จะเลือก item ถัดไปหรือก่อนหน้าหลังจากลบ
  void refreshListAfterDelete(String deletedGuid) {
    if (!mounted) return;

    // Reset auto-refresh interval กลับไป 5 วินาที (เพราะมีการ delete)
    _currentRefreshInterval = _baseRefreshInterval;

    // หา index ของ item ที่ถูกลบ
    int deletedIndex = transListData.indexWhere((item) => item.guidfixed == deletedGuid);

    // เก็บตำแหน่ง scroll
    _pendingRestoreScrollOffset = listScrollController.hasClients
        ? listScrollController.offset
        : 0;

    // กำหนด guid ที่จะเลือกหลังจาก refresh (item ถัดไปหรือก่อนหน้า)
    if (deletedIndex >= 0 && transListData.length > 1) {
      // ถ้าลบ item สุดท้าย ให้เลือก item ก่อนหน้า
      // ถ้าลบ item อื่น ให้เลือก item ถัดไป (ซึ่งจะขยับมาอยู่ที่ index เดิม)
      int nextIndex = deletedIndex >= transListData.length - 1
          ? deletedIndex - 1
          : deletedIndex;
      if (nextIndex >= 0 && nextIndex < transListData.length) {
        // ถ้าเป็น index เดิม ต้องเลือก item ถัดไป (หลังลบ item นี้จะขยับมา)
        if (nextIndex == deletedIndex && deletedIndex + 1 < transListData.length) {
          _pendingRestoreGuid = transListData[deletedIndex + 1].guidfixed ?? '';
        } else {
          _pendingRestoreGuid = transListData[nextIndex].guidfixed ?? '';
        }
      } else {
        _pendingRestoreGuid = '';
      }
    } else {
      _pendingRestoreGuid = '';
    }

    _isRefreshing = true;

    // ไม่ clear list เพื่อให้ smooth - โหลดข้อมูลใหม่มาก่อนแล้วค่อย replace
    _loadDataListForRefresh();
  }

  /// Helper method เพื่อ restore scroll position และ selection หลังจากโหลดเสร็จ
  void _restorePositionAfterRefresh() {
    if (!_isRefreshing) return;

    _isRefreshing = false;

    // Restore scroll position
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted && listScrollController.hasClients) {
        // ตรวจสอบว่า offset ไม่เกิน maxScrollExtent
        double targetOffset = _pendingRestoreScrollOffset;
        if (listScrollController.position.maxScrollExtent > 0) {
          targetOffset = targetOffset.clamp(0, listScrollController.position.maxScrollExtent);
        }
        listScrollController.jumpTo(targetOffset);
      }

      // Restore selection
      if (mounted && _pendingRestoreGuid.isNotEmpty) {
        int foundIndex = transListData.indexWhere((item) => item.guidfixed == _pendingRestoreGuid);
        if (foundIndex >= 0) {
          setState(() {
            selectGuid = _pendingRestoreGuid;
          });
        }
      }
    });
  }

  // =====================================================
  // Auto-Refresh Methods สำหรับ sync ข้อมูลจากเครื่องอื่น
  // =====================================================

  /// คำนวณ hash ของข้อมูลเพื่อเปรียบเทียบว่ามีการเปลี่ยนแปลงหรือไม่
  String _calculateDataHash() {
    if (transListData.isEmpty) return '';
    // ใช้ข้อมูลที่สำคัญในการสร้าง hash: guidfixed, docno, totalamount, iscancel
    final buffer = StringBuffer();
    for (var item in transListData) {
      buffer.write('${item.guidfixed}|${item.docno}|${item.totalamount}|${item.iscancel}|');
    }
    return buffer.toString().hashCode.toString();
  }

  /// เริ่ม auto-refresh timer
  void _startAutoRefresh() {
    AppLogger.debug('🔄 [AutoRefresh] _startAutoRefresh called');
    // ยกเลิก timer เดิมก่อน (ถ้ามี)
    _stopAutoRefresh();

    // Reset interval กลับไป 5 วินาที
    _currentRefreshInterval = _baseRefreshInterval;

    // บันทึก hash ปัจจุบัน
    _lastDataHash = _calculateDataHash();
    AppLogger.debug('🔄 [AutoRefresh] Initial hash: $_lastDataHash');

    // เริ่ม timer ใหม่
    _scheduleNextAutoRefresh();
  }

  /// หยุด auto-refresh timer
  void _stopAutoRefresh() {
    _autoRefreshTimer?.cancel();
    _autoRefreshTimer = null;
    _isAutoRefreshing = false;
  }

  /// ตั้งเวลา auto-refresh ครั้งถัดไป
  void _scheduleNextAutoRefresh() {
    if (!mounted) return;

    AppLogger.debug('🔄 [AutoRefresh] Scheduling next refresh in $_currentRefreshInterval seconds');
    _autoRefreshTimer?.cancel();
    _autoRefreshTimer = Timer(Duration(seconds: _currentRefreshInterval), () {
      AppLogger.debug('🔄 [AutoRefresh] Timer fired! mounted=$mounted, loadingData=$loadingData, _isRefreshing=$_isRefreshing');
      if (mounted && !loadingData && !_isRefreshing) {
        _performAutoRefresh();
      } else {
        // ถ้ากำลัง loading อยู่ ให้รอแล้วลองใหม่
        AppLogger.debug('🔄 [AutoRefresh] Busy, rescheduling...');
        _scheduleNextAutoRefresh();
      }
    });
  }

  /// ทำการ auto-refresh และเปรียบเทียบข้อมูล
  void _performAutoRefresh() {
    AppLogger.debug('🔄 [AutoRefresh] _performAutoRefresh called');
    if (!mounted || loadingData || _isRefreshing) {
      AppLogger.debug('🔄 [AutoRefresh] Skipped - mounted=$mounted, loadingData=$loadingData, _isRefreshing=$_isRefreshing');
      return;
    }

    _isAutoRefreshing = true;
    AppLogger.debug('🔄 [AutoRefresh] Loading data silently...');

    // โหลดข้อมูลใหม่โดยไม่แสดง loading indicator (silent refresh)
    _loadDataListForAutoRefresh();
  }

  /// โหลดข้อมูลสำหรับ auto-refresh (ไม่แสดง loading indicator)
  void _loadDataListForAutoRefresh() {
    if (!mounted) return;

    // ไม่ต้อง setState loadingData = true เพื่อไม่ให้แสดง loading indicator

    context.read<TransBloc>().add(
      TransLoad(
        offset: 0,
        limit: 100,
        search: searchText,
        type: widget.type,
        custcode: widget.custcode,
        dateorder: dateOrderByType,
        ispos: (filterTransaction == 2)
            ? "true"
            : (filterTransaction == 3)
            ? "false"
            : "null",
      ),
    );
  }

  /// Public method สำหรับ reset auto-refresh interval (เรียกจาก parent เมื่อ save/update/delete)
  void resetAutoRefreshInterval() {
    _currentRefreshInterval = _baseRefreshInterval;
    _startAutoRefresh();
  }

  /// เพิ่ม interval ทีละ 5 วินาที (เมื่อข้อมูลไม่เปลี่ยน)
  void _increaseAutoRefreshInterval() {
    _currentRefreshInterval += _baseRefreshInterval;
    // จำกัดไม่ให้เกิน 60 วินาที
    if (_currentRefreshInterval > 60) {
      _currentRefreshInterval = 60;
    }
  }

  /// จัดการผลลัพธ์จาก auto-refresh
  void _handleAutoRefreshResult(List<TransactionModel> newData) {
    AppLogger.debug('🔄 [AutoRefresh] _handleAutoRefreshResult called, newData.length=${newData.length}');
    if (!mounted) return;

    // คำนวณ hash ของข้อมูลใหม่
    final newHash = _calculateNewDataHash(newData);
    AppLogger.debug('🔄 [AutoRefresh] Old hash: $_lastDataHash, New hash: $newHash');

    if (newHash != _lastDataHash) {
      // ข้อมูลเปลี่ยน -> update UI
      AppLogger.debug('🔄 [AutoRefresh] Data CHANGED! Updating UI...');
      setState(() {
        transListData = List.from(newData);
      });
      _lastDataHash = newHash;
      // Reset interval กลับไป 5 วินาที
      _currentRefreshInterval = _baseRefreshInterval;
    } else {
      // ข้อมูลไม่เปลี่ยน -> เพิ่ม interval
      AppLogger.debug('🔄 [AutoRefresh] Data unchanged, increasing interval to ${_currentRefreshInterval + _baseRefreshInterval}s');
      _increaseAutoRefreshInterval();
    }

    _isAutoRefreshing = false;

    // ตั้งเวลา refresh ครั้งถัดไป
    _scheduleNextAutoRefresh();
  }

  /// คำนวณ hash จากข้อมูลใหม่
  /// รวม isref, isclosed, islocked เพื่อให้ตรวจจับการเปลี่ยนแปลงสถานะจาก backend
  String _calculateNewDataHash(List<TransactionModel> data) {
    if (data.isEmpty) return '';
    final buffer = StringBuffer();
    for (var item in data) {
      buffer.write('${item.guidfixed}|${item.docno}|${item.totalamount}|${item.iscancel}|${item.isref}|${item.isclosed}|${item.islocked}|');
    }
    return buffer.toString().hashCode.toString();
  }

  // โหลดสถานะการอนุมัติสำหรับเอกสาร PO ในรายการ
  Future<void> _loadApprovalStatuses() async {
    // โหลดเฉพาะกรณีเป็น purchaseorder
    if (widget.type != global.TransactionTypeEnum.purchaseorder) return;
    if (transListData.isEmpty) return;

    // ดึง docno ทั้งหมดในรายการ
    final docNos = transListData
        .map((t) => t.docno)
        .where((d) => d.isNotEmpty)
        .toList();

    if (docNos.isEmpty) return;

    // เรียก batch API
    final statusMap = await ApprovalApiService.getBatchPOApprovalStatus(docNos);

    if (mounted) {
      setState(() {
        _approvalStatusMap = statusMap;
      });
    }
  }

  /// เริ่ม polling สถานะอนุมัติ PO ทุก 5 วินาที
  void _startApprovalStatusPolling() {
    // เฉพาะ purchaseorder เท่านั้น
    if (widget.type != global.TransactionTypeEnum.purchaseorder) return;

    _stopApprovalStatusPolling();
    _approvalStatusPollingTimer = Timer.periodic(
      Duration(seconds: _approvalPollingInterval),
      (_) => _loadApprovalStatuses(),
    );
    AppLogger.debug('🔄 [ApprovalPolling] Started polling every $_approvalPollingInterval seconds');
  }

  /// หยุด polling สถานะอนุมัติ PO
  void _stopApprovalStatusPolling() {
    _approvalStatusPollingTimer?.cancel();
    _approvalStatusPollingTimer = null;
  }

  /// Helper: เลือก icon ตาม latest action text
  IconData _getLatestActionIcon(String actionText) {
    if (actionText.contains('ส่งการแจ้งเตือน')) return Icons.notifications_active;
    if (actionText.contains(global.language('open_read'))) return Icons.visibility;
    if (actionText.contains(global.language('approve'))) return Icons.check;
    if (actionText.contains(global.language('reject'))) return Icons.close;
    if (actionText.contains(global.language('submit_for_approval')) || actionText.contains('ส่งใหม่')) return Icons.send;
    return Icons.info_outline;
  }

  /// Helper: จัดรูปแบบข้อความ latest action พร้อมวันที่เวลา
  String _formatLatestAction(POApprovalStatusInfo status) {
    final actionText = status.latestActionText ?? '';
    final actionAt = status.latestActionAt;
    final actionBy = status.latestActionBy ?? '';

    String result = actionText;
    if (actionAt != null) {
      // แปลงเวลา UTC เป็น local time
      final localTime = actionAt.toLocal();
      final dateFormat = DateFormat('d/M/yy HH:mm');
      result += ' ${dateFormat.format(localTime)}';
    }
    if (actionBy.isNotEmpty && !actionText.contains(actionBy)) {
      result += ' ($actionBy)';
    }
    return result;
  }

  // ฟังก์ชันสำหรับยกเลิกการโหลดและรีเซ็ต state ทั้งหมด
  void cancelAllOperationsAndReset() {
    // แจ้ง parent widget ว่าหยุด loading แล้ว
    if (widget.onLoadingStateChanged != null) {
      widget.onLoadingStateChanged!(false);
    }

    // รีเซ็ต loading states
    if (mounted) {
      setState(() {
        loadingData = false;
        isSearching = false;

        // รีเซ็ต search variables
        pendingSearchText = "";
        currentSearchRequest = "";
      });
    }
  }

  @override
  void dispose() {
    _stopAutoRefresh(); // หยุด auto-refresh timer
    _stopApprovalStatusPolling(); // หยุด polling สถานะอนุมัติ
    listScrollController.dispose();
    searchController.dispose();
    _minAmountController.dispose();
    _maxAmountController.dispose();
    _creditorSearchDebounce?.cancel(); // ยกเลิก debounce timer
    _creditorSearchController.dispose();
    _creditorFilterCubit.close(); // ปิด Cubit เมื่อ widget ถูกทำลาย
    super.dispose();
  }

  // === Helper methods สำหรับการกรองข้อมูล (client-side) ===

  /// กรองรายการตามเงื่อนไขที่ตั้งไว้
  List<TransactionModel> _applyClientFilters(List<TransactionModel> data) {
    var results = data;

    // กรองตามช่วงวันที่
    if (_filterFromDate != null) {
      results = results.where((item) {
        final docDate = DateTime.tryParse(item.docdatetime);
        if (docDate == null) return true;
        return docDate.isAfter(_filterFromDate!) ||
            docDate.isAtSameMomentAs(_filterFromDate!);
      }).toList();
    }

    if (_filterToDate != null) {
      results = results.where((item) {
        final docDate = DateTime.tryParse(item.docdatetime);
        if (docDate == null) return true;
        return docDate.isBefore(_filterToDate!.add(const Duration(days: 1))) ||
            docDate.isAtSameMomentAs(_filterToDate!);
      }).toList();
    }

    // กรองตามช่วงจำนวนเงิน
    if (_filterMinAmount != null) {
      results = results.where((item) {
        return item.totalamount >= _filterMinAmount!;
      }).toList();
    }

    if (_filterMaxAmount != null) {
      results = results.where((item) {
        return item.totalamount <= _filterMaxAmount!;
      }).toList();
    }

    // กรองตามเจ้าหนี้ที่เลือก (multi-select)
    if (_filterCreditorCodes.isNotEmpty) {
      results = results.where((item) {
        return _filterCreditorCodes.contains(item.custcode);
      }).toList();
    }

    return results;
  }

  /// ตรวจสอบว่ามีเงื่อนไขการกรองหรือไม่
  bool get _hasActiveFilters =>
      _filterFromDate != null ||
      _filterToDate != null ||
      _filterMinAmount != null ||
      _filterMaxAmount != null ||
      _filterCreditorCodes.isNotEmpty;

  /// ล้างเงื่อนไขการกรองทั้งหมด
  void _clearAllFilters() {
    setState(() {
      _filterFromDate = null;
      _filterToDate = null;
      _filterMinAmount = null;
      _filterMaxAmount = null;
      _filterCreditorCodes = []; // ล้างเจ้าหนี้ที่เลือก
      // ล้าง text ใน TextField ด้วย
      _minAmountController.clear();
      _maxAmountController.clear();
    });
  }

  /// เพิ่มเจ้าหนี้ในรายการ filter
  void _addCreditorFilter(String code) {
    if (!_filterCreditorCodes.contains(code)) {
      setState(() {
        _filterCreditorCodes = [..._filterCreditorCodes, code];
      });
    }
  }

  /// ลบเจ้าหนี้ออกจากรายการ filter
  void _removeCreditorFilter(String code) {
    setState(() {
      _filterCreditorCodes = _filterCreditorCodes.where((c) => c != code).toList();
    });
  }

  /// เพิ่มขนาด font
  void _increaseFontSize() {
    setState(() {
      if (_customFontSize < _maxFontSize) {
        _customFontSize += 1.0;
      }
    });
  }

  /// ลดขนาด font
  void _decreaseFontSize() {
    setState(() {
      if (_customFontSize > _minFontSize) {
        _customFontSize -= 1.0;
      }
    });
  }

  /// สร้าง widget สำหรับ Advanced Filter Panel - Modern Design
  Widget _buildAdvancedFilterPanel(double screenWidth) {
    final isSmallScreen = isMobile(screenWidth);
    final fontSize = isSmallScreen ? 11.0 : 12.0;

    if (!_showAdvancedFilter) return const SizedBox.shrink();

    return Container(
      margin: EdgeInsets.symmetric(horizontal: isSmallScreen ? 4 : 8, vertical: 4),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [global.theme.infoHighlightColor, global.theme.positiveHighlightColor],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: global.theme.infoHighlightColor),
        boxShadow: [
          BoxShadow(
            color: global.theme.infoHighlightTextColor.withValues(alpha: 0.1),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Padding(
        padding: EdgeInsets.all(isSmallScreen ? 8 : 12),
        child: Column(
          children: [
            // แถวบน: ช่วงวันที่
            Row(
              children: [
                // Icon วันที่
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: global.theme.columnHeaderColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.date_range, size: 16, color: global.theme.infoHighlightTextColor),
                ),
                const SizedBox(width: 8),
                // ปุ่มเลือกวันที่จาก
                Expanded(
                  child: _buildDateButton(
                    label: global.language('from'),
                    date: _filterFromDate,
                    isFromDate: true,
                    color: global.theme.primaryColor,
                    fontSize: fontSize,
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: Icon(Icons.arrow_forward, size: 14, color: global.theme.iconSecondaryColor),
                ),
                // ปุ่มเลือกวันที่ถึง
                Expanded(
                  child: _buildDateButton(
                    label: global.language('kb_to'),
                    date: _filterToDate,
                    isFromDate: false,
                    color: global.theme.primaryColor,
                    fontSize: fontSize,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            // แถวล่าง: ช่วงจำนวนเงิน
            Row(
              children: [
                // Icon ยอดเงิน
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: global.theme.positiveHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.attach_money, size: 16, color: global.theme.positiveHighlightTextColor),
                ),
                const SizedBox(width: 8),
                // TextField ยอดเงินต่ำสุด
                Expanded(
                  child: _buildAmountField(
                    controller: _minAmountController,
                    hint: global.language('lower_bound'),
                    fontSize: fontSize,
                    onChanged: (value) {
                      setState(() => _filterMinAmount = double.tryParse(value));
                    },
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: Icon(Icons.arrow_forward, size: 14, color: global.theme.iconSecondaryColor),
                ),
                // TextField ยอดเงินสูงสุด
                Expanded(
                  child: _buildAmountField(
                    controller: _maxAmountController,
                    hint: global.language('max'),
                    fontSize: fontSize,
                    onChanged: (value) {
                      setState(() => _filterMaxAmount = double.tryParse(value));
                    },
                  ),
                ),
                // ปุ่มล้างเงื่อนไข
                const SizedBox(width: 8),
                if (_hasActiveFilters)
                  Material(
                    color: global.theme.negativeHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                    child: InkWell(
                      onTap: _clearAllFilters,
                      borderRadius: BorderRadius.circular(8),
                      child: Container(
                        padding: const EdgeInsets.all(6),
                        child: Icon(Icons.clear_all, size: 18, color: global.theme.negativeHighlightTextColor),
                      ),
                    ),
                  ),
              ],
            ),
            // แถวเจ้าหนี้: แสดงเฉพาะประเภทที่เกี่ยวข้องกับเจ้าหนี้
            if (_isPurchaseType()) ...[
              const SizedBox(height: 10),
              _buildCreditorFilterSection(fontSize),
            ],
          ],
        ),
      ),
    );
  }

  /// ตรวจสอบว่าเป็นประเภทเอกสารที่เกี่ยวกับเจ้าหนี้หรือไม่
  bool _isPurchaseType() {
    return widget.type == TransactionTypeEnum.purchase ||
        widget.type == TransactionTypeEnum.purchasereturn ||
        widget.type == TransactionTypeEnum.purchaseorder ||
        widget.type == TransactionTypeEnum.purchasepartial ||
        widget.type == TransactionTypeEnum.accrualreceive ||
        widget.type == TransactionTypeEnum.advancePayment;
  }

  /// สร้าง section สำหรับเลือกเจ้าหนี้
  Widget _buildCreditorFilterSection(double fontSize) {
    return BlocBuilder<CreditorFilterCubit, CreditorFilterState>(
      bloc: _creditorFilterCubit,
      builder: (context, state) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // แถวเลือกเจ้าหนี้
            Row(
              children: [
                // Icon เจ้าหนี้
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: global.theme.infoHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.business, size: 16, color: global.theme.warningHighlightTextColor),
                ),
                const SizedBox(width: 8),
                // TextField ค้นหาเจ้าหนี้ (ค้นหาจาก API โดยตรง)
                Expanded(
                  child: _buildCreditorDropdown(state, fontSize),
                ),
              ],
            ),
            // แสดงเจ้าหนี้ที่เลือกแบบ Wrap
            if (_filterCreditorCodes.isNotEmpty) ...[
              const SizedBox(height: 8),
              _buildSelectedCreditorChips(state, fontSize),
            ],
          ],
        );
      },
    );
  }

  /// สร้าง TextField ค้นหาเจ้าหนี้แบบ Autocomplete (ค้นหาจาก API โดยตรง)
  Widget _buildCreditorDropdown(CreditorFilterState state, double fontSize) {
    // กรองเฉพาะที่ยังไม่ได้เลือก (จากผลลัพธ์ที่ค้นหาจาก API)
    final availableCreditors = state.creditors
        .where((c) => !_filterCreditorCodes.contains(c.code))
        .toList();

    // ใช้ searchText จาก state (ค้นหาจาก API แล้ว)
    final searchText = state.searchText;
    final hasSearchText = searchText.isNotEmpty;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // TextField สำหรับค้นหา (style เหมือน amount field)
        SizedBox(
          height: 36,
          child: TextField(
            controller: _creditorSearchController,
            style: TextStyle(fontSize: fontSize, fontWeight: FontWeight.w500),
            decoration: InputDecoration(
              hintText: state.isLoading ? global.language('searching') : language('type_creditor_code_name_to_search'),
              hintStyle: TextStyle(fontSize: fontSize, color: global.theme.formHintColor),
              prefixIcon: state.isLoading
                  ? SizedBox(
                      width: 16,
                      height: 16,
                      child: Padding(
                        padding: const EdgeInsets.all(10),
                        child: CircularProgressIndicator(strokeWidth: 2, color: global.theme.warningHighlightTextColor),
                      ),
                    )
                  : Icon(Icons.search, size: 16, color: global.theme.warningHighlightTextColor),
              suffixIcon: hasSearchText
                  ? IconButton(
                      icon: Icon(Icons.clear, size: 14, color: global.theme.iconSecondaryColor),
                      onPressed: () {
                        _creditorSearchController.clear();
                        // ล้าง search และ load ข้อมูลเริ่มต้น
                        _creditorFilterCubit.searchCreditors("");
                      },
                      padding: EdgeInsets.zero,
                      constraints: const BoxConstraints(),
                    )
                  : null,
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
              filled: true,
              fillColor: global.theme.formFillColor,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: global.theme.formBorderColor),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: global.theme.formBorderColor),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: global.theme.warningHighlightTextColor, width: 1.5),
              ),
            ),
            onChanged: (value) {
              // ใช้ debounce เพื่อไม่ให้เรียก API บ่อยเกินไป
              _creditorSearchDebounce?.cancel();
              _creditorSearchDebounce = Timer(const Duration(milliseconds: 300), () {
                _creditorFilterCubit.searchCreditors(value);
              });
            },
          ),
        ),
        // รายการ suggestions (แสดงเมื่อมีการพิมพ์ค้นหา)
        if (hasSearchText && availableCreditors.isNotEmpty && !state.isLoading)
          Container(
            margin: const EdgeInsets.only(top: 4),
            constraints: const BoxConstraints(maxHeight: 150),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: global.theme.warningHighlightColor),
              boxShadow: [
                BoxShadow(
                  color: global.theme.textColor.withValues(alpha: 0.1),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: ListView.builder(
              shrinkWrap: true,
              padding: EdgeInsets.zero,
              itemCount: availableCreditors.length > 10 ? 10 : availableCreditors.length,
              itemBuilder: (context, index) {
                final creditor = availableCreditors[index];
                final name = creditor.names.isNotEmpty ? creditor.names.first.name : '';
                return InkWell(
                  onTap: () {
                    _addCreditorFilter(creditor.code);
                    // ล้าง search หลังเลือก
                    _creditorSearchController.clear();
                    _creditorFilterCubit.searchCreditors("");
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                    decoration: BoxDecoration(
                      border: index < (availableCreditors.length > 10 ? 9 : availableCreditors.length - 1)
                          ? Border(bottom: BorderSide(color: global.theme.dividerBorderColor))
                          : null,
                    ),
                    child: Row(
                      children: [
                        Icon(Icons.business, size: 14, color: global.theme.warningHighlightTextColor),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                name,
                                style: TextStyle(
                                  fontSize: fontSize,
                                  fontWeight: FontWeight.w500,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                              Text(
                                'รหัส: ${creditor.code}',
                                style: TextStyle(
                                  fontSize: fontSize - 1,
                                  color: global.theme.textSecondaryColor,
                                ),
                              ),
                            ],
                          ),
                        ),
                        Icon(Icons.add_circle_outline, size: 16, color: global.theme.positiveHighlightTextColor),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
        // แสดงข้อความถ้าไม่พบผลลัพธ์
        if (hasSearchText && availableCreditors.isEmpty && !state.isLoading)
          Container(
            margin: const EdgeInsets.only(top: 4),
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: global.theme.dividerBorderColor,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Icon(Icons.info_outline, size: 14, color: global.theme.iconSecondaryColor),
                const SizedBox(width: 6),
                Text(
                  'ไม่พบเจ้าหนี้ที่ตรงกับ "$searchText"',
                  style: TextStyle(fontSize: fontSize - 1, color: global.theme.textSecondaryColor),
                ),
              ],
            ),
          ),
      ],
    );
  }

  /// สร้าง Chips แสดงเจ้าหนี้ที่เลือก
  Widget _buildSelectedCreditorChips(CreditorFilterState state, double fontSize) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(6),
      decoration: BoxDecoration(
        color: global.theme.warningHighlightColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.warningHighlightColor),
      ),
      child: Wrap(
        spacing: 6,
        runSpacing: 4,
        children: _filterCreditorCodes.map((code) {
          // หาชื่อเจ้าหนี้จาก state (ถ้าไม่เจอใช้ code แทน)
          String name = code;
          try {
            final creditor = state.creditors.firstWhere((c) => c.code == code);
            if (creditor.names.isNotEmpty) {
              name = creditor.names.first.name;
            }
          } catch (_) {
            // ใช้ code เป็นชื่อถ้าหาไม่เจอ
          }

          return Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: global.theme.warningHighlightTextColor),
              boxShadow: [
                BoxShadow(
                  color: global.theme.warningHighlightColor,
                  blurRadius: 2,
                  offset: const Offset(0, 1),
                ),
              ],
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.business, size: 12, color: global.theme.warningHighlightTextColor),
                const SizedBox(width: 4),
                Text(
                  name.length > 15 ? '${name.substring(0, 15)}...' : name,
                  style: TextStyle(
                    fontSize: fontSize - 1,
                    color: global.theme.warningHighlightTextColor,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(width: 4),
                // ปุ่มลบ
                InkWell(
                  onTap: () => _removeCreditorFilter(code),
                  child: Container(
                    padding: const EdgeInsets.all(2),
                    decoration: BoxDecoration(
                      color: global.theme.negativeHighlightColor,
                      shape: BoxShape.circle,
                    ),
                    child: Icon(Icons.close, size: 12, color: global.theme.negativeHighlightTextColor),
                  ),
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }

  /// สร้างปุ่มเลือกวันที่แบบสวย (แสดง popup ใกล้ปุ่ม)
  Widget _buildDateButton({
    required String label,
    required DateTime? date,
    required bool isFromDate,
    required Color color,
    required double fontSize,
  }) {
    final hasDate = date != null;
    return Builder(
      builder: (buttonContext) {
        return Material(
          color: hasDate ? color.withValues(alpha: 0.1) : global.theme.cardColor,
          borderRadius: BorderRadius.circular(8),
          child: InkWell(
            onTap: () => _selectDate(isFromDate, buttonContext),
            borderRadius: BorderRadius.circular(8),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: hasDate ? color : global.theme.dividerBorderColor,
                  width: hasDate ? 1.5 : 1,
                ),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.calendar_today,
                    size: 12,
                    color: hasDate ? color : global.theme.iconSecondaryColor,
                  ),
                  const SizedBox(width: 6),
                  Flexible(
                    child: Text(
                      hasDate
                          ? global.formatThaiDateTime(dateTime: date, showTime: false)
                          : label,
                      style: TextStyle(
                        fontSize: fontSize,
                        color: hasDate ? global.theme.infoHighlightTextColor : global.theme.iconSecondaryColor,
                        fontWeight: hasDate ? FontWeight.w600 : FontWeight.normal,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  /// สร้างปุ่มยอดเงินที่กดแล้วแสดง NumPad Dialog ใกล้ปุ่ม
  Widget _buildAmountField({
    required TextEditingController controller,
    required String hint,
    required double fontSize,
    required ValueChanged<String> onChanged,
  }) {
    final hasValue = controller.text.isNotEmpty;
    return Builder(
      builder: (buttonContext) {
        return InkWell(
          onTap: () async {
            // หาตำแหน่งของปุ่ม
            final RenderBox button = buttonContext.findRenderObject() as RenderBox;
            final Offset buttonPosition = button.localToGlobal(Offset.zero);
            final Size buttonSize = button.size;
            final screenSize = MediaQuery.of(context).size;

            // คำนวณตำแหน่ง popup (แสดงใต้ปุ่ม หรือบนปุ่มถ้าไม่มีที่)
            const dialogWidth = 320.0;
            const dialogHeight = 560.0;

            // คำนวณ left: ให้อยู่ตรงกลางของปุ่ม หรือปรับให้อยู่ในจอ
            double left = buttonPosition.dx + (buttonSize.width / 2) - (dialogWidth / 2);
            if (left < 10) left = 10;
            if (left + dialogWidth > screenSize.width - 10) {
              left = screenSize.width - dialogWidth - 10;
            }

            // คำนวณ top: แสดงใต้ปุ่ม ถ้าไม่พอให้แสดงบนปุ่ม
            double top = buttonPosition.dy + buttonSize.height + 8;
            if (top + dialogHeight > screenSize.height - 20) {
              top = buttonPosition.dy - dialogHeight - 8;
              if (top < 20) top = 20;
            }

            // แสดง NumPad Dialog ใกล้ปุ่ม
            final result = await showDialog<String>(
              context: context,
              barrierColor: global.theme.textColor.withValues(alpha: 0.3),
              builder: (BuildContext dialogContext) {
                return Stack(
                  children: [
                    Positioned(
                      left: left,
                      top: top,
                      child: Material(
                        color: Colors.transparent,
                        child: global.NumPadDialog(
                          title: hint,
                          initialValue: controller.text.isEmpty ? null : controller.text,
                          decimalPlaces: 2,
                        ),
                      ),
                    ),
                  ],
                );
              },
            );
            if (result != null) {
              controller.text = result;
              onChanged(result);
            }
          },
          child: Container(
            height: 36,
            padding: const EdgeInsets.symmetric(horizontal: 8),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: hasValue ? global.theme.positiveHighlightTextColor : global.theme.dividerBorderColor,
                width: hasValue ? 1.5 : 1,
              ),
            ),
            child: Center(
              child: Text(
                hasValue ? controller.text : hint,
                style: TextStyle(
                  fontSize: fontSize,
                  fontWeight: hasValue ? FontWeight.w500 : FontWeight.normal,
                  color: hasValue ? global.theme.textColor : global.theme.iconSecondaryColor,
                ),
                textAlign: TextAlign.center,
              ),
            ),
          ),
        );
      },
    );
  }

  /// เลือกวันที่ ใช้ CustomDatePickerDialog แบบ พ.ศ. (แสดง popup ใกล้ปุ่ม)
  Future<void> _selectDate(bool isFromDate, BuildContext buttonContext) async {
    DateTime initialDate = isFromDate
        ? (_filterFromDate ?? DateTime.now())
        : (_filterToDate ?? DateTime.now());
    DateTime now = DateTime.now();
    DateTime firstDate = DateTime(now.year - 10, 1, 1);
    DateTime lastDate = DateTime(now.year + 10, 12, 31);

    if (initialDate.isBefore(firstDate)) {
      initialDate = firstDate;
    } else if (initialDate.isAfter(lastDate)) {
      initialDate = lastDate;
    }

    // หาตำแหน่งของปุ่ม
    final RenderBox button = buttonContext.findRenderObject() as RenderBox;
    final Offset buttonPosition = button.localToGlobal(Offset.zero);
    final Size buttonSize = button.size;
    final screenSize = MediaQuery.of(context).size;

    const dialogWidth = 360.0;
    const dialogHeight = 450.0;

    // คำนวณ left: ให้อยู่ตรงกลางของปุ่ม หรือปรับให้อยู่ในจอ
    double left = buttonPosition.dx + (buttonSize.width / 2) - (dialogWidth / 2);
    if (left < 10) left = 10;
    if (left + dialogWidth > screenSize.width - 10) {
      left = screenSize.width - dialogWidth - 10;
    }

    // คำนวณ top: แสดงใต้ปุ่ม ถ้าไม่พอให้แสดงบนปุ่ม
    double top = buttonPosition.dy + buttonSize.height + 8;
    if (top + dialogHeight > screenSize.height - 20) {
      top = buttonPosition.dy - dialogHeight - 8;
      if (top < 20) top = 20;
    }

    DateTime? pickedDate = await showDialog<DateTime>(
      context: context,
      barrierColor: global.theme.textColor.withValues(alpha: 0.3),
      builder: (BuildContext dialogContext) {
        return Stack(
          children: [
            Positioned(
              left: left,
              top: top,
              child: Material(
                color: Colors.transparent,
                child: CustomDatePickerDialog(
                  initialDate: initialDate,
                  firstDate: firstDate,
                  lastDate: lastDate,
                  // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                ),
              ),
            ),
          ],
        );
      },
    );

    if (pickedDate != null) {
      setState(() {
        if (isFromDate) {
          _filterFromDate = pickedDate;
        } else {
          _filterToDate = pickedDate;
        }
      });
    }
  }

  Widget listScreen({required double screenWidth}) {
    final isSmallScreen = isMobile(screenWidth);
    final titleFontSize = getResponsiveFontSize(screenWidth, mobile: 14, tablet: 16, desktop: 18);
    final buttonTextSize = getResponsiveFontSize(screenWidth, mobile: 12, tablet: 14, desktop: 14);
    final spacing = getResponsivePadding(screenWidth, mobile: 4, tablet: 8, desktop: 8);

    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        elevation: 1,
        titleSpacing: spacing,
        title: Row(
          children: [
            // Title - ปรับขนาดตามหน้าจอ
            Expanded(
              flex: isSmallScreen ? 2 : 3,
              child: Text(
                global.transactionName(widget.type),
                style: TextStyle(fontSize: titleFontSize),
                overflow: TextOverflow.ellipsis,
              ),
            ),

            // Spacer สำหรับดันไปขวา
            Expanded(child: SizedBox()),

            // Date Sort Button - ปรับขนาดตามหน้าจอ
            ElevatedButton(
              onPressed: () {
                setState(() {
                  if (dateOrderByType == 1) {
                    dateOrderByType = 2;
                  } else {
                    dateOrderByType = 1;
                  }
                });
                transListData = [];
                loadDataList(searchText, filterTransaction);
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.white.withValues(alpha: 0.2),
                foregroundColor: global.theme.onPrimaryColor,
                elevation: isSmallScreen ? 2 : 3,
                padding: EdgeInsets.symmetric(horizontal: isSmallScreen ? 8 : 12, vertical: isSmallScreen ? 4 : 6),
                minimumSize: Size(isSmallScreen ? 36 : 48, isSmallScreen ? 28 : 32),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(isSmallScreen ? 6 : 8)),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(dateOrderByType == 1 ? Icons.arrow_downward : Icons.arrow_upward, size: isSmallScreen ? 12 : 16, color: global.theme.onPrimaryColor),
                  SizedBox(width: isSmallScreen ? 2 : 4),
                  if (!isSmallScreen) // แสดงข้อความเต็มสำหรับจอใหญ่
                    Text(
                      dateOrderByType == 1 ? global.language("latest_date") : "วันที่เก่าสุด",
                      style: TextStyle(fontSize: buttonTextSize, color: global.theme.onPrimaryColor),
                    ),
                ],
              ),
            ),

            SizedBox(width: isSmallScreen ? 4 : 8),

            // Record Count - ปรับขนาดและตำแหน่งตามหน้าจอ
            Text(
              isSmallScreen ? "$totalRecord" : "จำนวน $totalRecord รายการ",
              style: TextStyle(fontSize: buttonTextSize),
              textAlign: TextAlign.right,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          // ถ้าเป็น embedded mode ใช้ปุ่มปิด ถ้าไม่ใช่ใช้ปุ่มกลับ
          icon: Icon(widget.isEmbedded ? Icons.close : Icons.arrow_back),
          onPressed: () {
            cancelAllOperationsAndReset();
            if (widget.isEmbedded && widget.onClose != null) {
              widget.onClose!();
            } else {
              Navigator.pop(context, emptyResult);
            }
          },
        ),
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb || Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.escape) {
                cancelAllOperationsAndReset();
                Navigator.pop(context, emptyResult);
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  cancelAllOperationsAndReset();
                  Navigator.pop(context, emptyResult);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = transListData.indexOf(transListData.firstWhere((element) => element.guidfixed == selectGuid));
                if (index > 0) {
                  selectGuid = transListData[index - 1].guidfixed!;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = transListData.indexOf(transListData.firstWhere((element) => element.guidfixed == selectGuid));
                if (index < transListData.length - 1) {
                  selectGuid = transListData[index + 1].guidfixed!;
                }
                isKeyDown = true;
                setState(() {});
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            Container(
              padding: EdgeInsets.fromLTRB(
                getResponsivePadding(screenWidth, mobile: 8, tablet: 10, desktop: 10),
                getResponsivePadding(screenWidth, mobile: 6, tablet: 8, desktop: 8),
                getResponsivePadding(screenWidth, mobile: 8, tablet: 10, desktop: 10),
                getResponsivePadding(screenWidth, mobile: 6, tablet: 8, desktop: 8),
              ),
              color: global.theme.appBarColor,
              child: Container(
                padding: EdgeInsets.symmetric(horizontal: isSmallScreen ? 0.5 : 1, vertical: isSmallScreen ? 0.5 : 1),
                decoration: BoxDecoration(
                  color: global.theme.cardColor,
                  borderRadius: BorderRadius.circular(isSmallScreen ? 6 : 8),
                  boxShadow: [BoxShadow(color: global.theme.dividerBorderColor.withValues(alpha: 0.3), spreadRadius: 1, blurRadius: 4, offset: const Offset(0, 2))],
                  border: Border.all(color: global.theme.dividerBorderColor.withValues(alpha: 0.3), width: 1),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    Expanded(
                      child: TextFormField(
                        onFieldSubmitted: (value) {
                          // ถ้ากำลัง loading อยู่ ไม่ทำการค้นหา
                          if (loadingData) {
                            return;
                          }

                          // ค้นหาทันทีเมื่อกด Enter
                          final trimmedValue = value.trim();
                          pendingSearchText = trimmedValue; // sync pending search

                          if (trimmedValue != currentSearchRequest) {
                            setState(() {
                              transListData = [];
                              isSearching = false;
                            });
                            loadDataList(trimmedValue, filterTransaction);
                          }
                          searchFocusNode.requestFocus();
                        },
                        onChanged: (value) {
                          // ถ้ากำลัง loading อยู่ ไม่ทำการค้นหา
                          if (loadingData) {
                            return;
                          }

                          final trimmedValue = value.trim();

                          // อัพเดท pending search text
                          pendingSearchText = trimmedValue;

                          // ถ้า search text เหมือนกับที่กำลัง request อยู่ ไม่ต้องทำอะไร
                          if (trimmedValue == currentSearchRequest) {
                            return;
                          }

                          setState(() {
                            isSearching = true;
                          });

                          _debouncer.run(() {
                            try {
                              // ตรวจสอบว่า widget ยังคง mounted หรือไม่
                              if (!mounted) return;

                              // ตรวจสอบอีกครั้งว่ากำลัง loading หรือไม่
                              if (loadingData) {
                                setState(() {
                                  isSearching = false;
                                });
                                return;
                              }

                              // ตรวจสอบว่า pending search text ยังตรงกับที่ user พิมพ์ล่าสุดหรือไม่
                              final currentPending = pendingSearchText;

                              // ถ้า pending text เปลี่ยนไประหว่างที่รอ debounce ก็ไม่ search
                              if (currentPending != pendingSearchText) {
                                setState(() {
                                  isSearching = false;
                                });
                                return;
                              }

                              // ถ้าเหมือนกับที่กำลัง request อยู่ ก็ไม่ search
                              if (currentPending == currentSearchRequest) {
                                setState(() {
                                  isSearching = false;
                                });
                                return;
                              }

                              // ทำการ search
                              setState(() {
                                transListData = [];
                                isSearching = false;
                              });
                              loadDataList(currentPending, filterTransaction);
                            } catch (_) {
                              if (mounted) {
                                setState(() {
                                  isSearching = false;
                                });
                              }
                            }
                          });
                        },
                        autofocus: true,
                        focusNode: searchFocusNode,
                        controller: searchController,
                        decoration: InputDecoration(
                          isDense: true,
                          contentPadding: EdgeInsets.symmetric(
                            horizontal: getResponsivePadding(screenWidth, mobile: 6, tablet: 8, desktop: 8),
                            vertical: getResponsivePadding(screenWidth, mobile: 10, tablet: 12, desktop: 12),
                          ),
                          border: OutlineInputBorder(borderRadius: BorderRadius.circular(isSmallScreen ? 4 : 6), borderSide: BorderSide.none),
                          filled: true,
                          fillColor: global.theme.formFillColor,
                          hintText: loadingData ? global.language('searching') : global.language('search'),
                          hintStyle: TextStyle(color: global.theme.formHintColor, fontSize: getResponsiveFontSize(screenWidth, mobile: 12, tablet: 14, desktop: 14)),
                          prefixIcon: Icon(Icons.search, color: loadingData ? global.theme.iconSecondaryColor : global.theme.textSecondaryColor, size: isSmallScreen ? 18 : 20),
                          suffixIcon: (isSearching || loadingData)
                              ? const Padding(
                                  padding: EdgeInsets.all(8.0),
                                  child: SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)),
                                )
                              : searchController.text.isNotEmpty
                              ? IconButton(
                                  icon: Icon(Icons.clear, size: 18),
                                  onPressed: loadingData
                                      ? null
                                      : () {
                                          searchController.clear();
                                          pendingSearchText = '';
                                          setState(() {
                                            transListData = [];
                                            isSearching = false;
                                          });
                                          loadDataList('', filterTransaction);
                                        },
                                )
                              : null,
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    (widget.type == global.TransactionTypeEnum.sale)
                        ? Container(
                            decoration: BoxDecoration(
                              color: loadingData ? global.theme.dividerBorderColor : global.theme.cardColor,
                              borderRadius: BorderRadius.circular(2),
                              border: Border.all(color: global.theme.dividerBorderColor, width: 1),
                            ),
                            child: IconButton(
                              constraints: BoxConstraints(minWidth: isSmallScreen ? 36 : 40, minHeight: isSmallScreen ? 36 : 40),
                              onPressed: loadingData
                                  ? null
                                  : () async {
                                      filterTransaction = await filterBox(filterTransaction);
                                      transListData.clear();
                                      loadDataList(searchText, filterTransaction);
                                      setState(() {});
                                    },
                              icon: Icon(
                                (filterTransaction == 1) ? Icons.filter_alt_off : Icons.filter_alt,
                                color: loadingData
                                    ? global.theme.iconSecondaryColor
                                    : (filterTransaction == 1)
                                    ? global.theme.warningHighlightTextColor
                                    : global.theme.infoHighlightTextColor,
                                size: isSmallScreen ? 18 : 20,
                              ),
                              tooltip: filterTransaction == 1 ? global.language('remove_filter') : global.language('add_filter'),
                            ),
                          )
                        : Container(),
                    // ปุ่ม Advanced Filter
                    SizedBox(width: 4),
                    Container(
                      decoration: BoxDecoration(
                        color: _showAdvancedFilter ? global.theme.rowHoverColor : global.theme.cardColor,
                        borderRadius: BorderRadius.circular(2),
                        border: Border.all(color: _hasActiveFilters ? global.theme.infoHighlightTextColor : global.theme.dividerBorderColor, width: 1),
                      ),
                      child: IconButton(
                        constraints: BoxConstraints(minWidth: isSmallScreen ? 36 : 40, minHeight: isSmallScreen ? 36 : 40),
                        onPressed: () {
                          setState(() {
                            _showAdvancedFilter = !_showAdvancedFilter;
                          });
                        },
                        icon: Icon(
                          _hasActiveFilters ? Icons.filter_alt : Icons.tune,
                          color: _hasActiveFilters ? global.theme.infoHighlightTextColor : global.theme.textSecondaryColor,
                          size: isSmallScreen ? 18 : 20,
                        ),
                        tooltip: global.language('advanced_filter'),
                      ),
                    ),
                    // ปุ่มปรับขนาด font
                    SizedBox(width: 4),
                    Container(
                      decoration: BoxDecoration(
                        color: global.theme.cardColor,
                        borderRadius: BorderRadius.circular(2),
                        border: Border.all(color: global.theme.dividerBorderColor, width: 1),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          IconButton(
                            constraints: BoxConstraints(minWidth: isSmallScreen ? 28 : 32, minHeight: isSmallScreen ? 36 : 40),
                            padding: EdgeInsets.zero,
                            onPressed: _customFontSize > _minFontSize ? _decreaseFontSize : null,
                            icon: Icon(Icons.text_decrease, size: isSmallScreen ? 14 : 16, color: _customFontSize > _minFontSize ? global.theme.textColor : global.theme.iconSecondaryColor),
                            tooltip: global.language('decrease_font_size'),
                          ),
                          Text('${_customFontSize.toInt()}', style: TextStyle(fontSize: 10, color: global.theme.textSecondaryColor)),
                          IconButton(
                            constraints: BoxConstraints(minWidth: isSmallScreen ? 28 : 32, minHeight: isSmallScreen ? 36 : 40),
                            padding: EdgeInsets.zero,
                            onPressed: _customFontSize < _maxFontSize ? _increaseFontSize : null,
                            icon: Icon(Icons.text_increase, size: isSmallScreen ? 14 : 16, color: _customFontSize < _maxFontSize ? global.theme.textColor : global.theme.iconSecondaryColor),
                            tooltip: global.language('increase_font_size'),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            // Advanced Filter Panel
            _buildAdvancedFilterPanel(screenWidth),
            // ซ่อน Table Header สำหรับมือถือ
            if (!isMobile(screenWidth))
              Container(
                padding: EdgeInsets.only(
                  left: getResponsivePadding(screenWidth, mobile: 8, tablet: 10, desktop: 10),
                  right: getResponsivePadding(screenWidth, mobile: 8, tablet: 10, desktop: 10),
                  top: getResponsivePadding(screenWidth, mobile: 4, tablet: 5, desktop: 5),
                  bottom: getResponsivePadding(screenWidth, mobile: 4, tablet: 5, desktop: 5),
                ),
                decoration: BoxDecoration(
                  color: global.theme.columnHeaderColor,
                  border: Border(bottom: BorderSide(width: 0.1, color: global.theme.dividerBorderColor)),
                ),
                child: Row(children: buildTableHeader(screenWidth)),
              ),
            Expanded(
              child: Stack(
                children: [
                  // List content with fade animation (ใช้ข้อมูลที่กรองแล้ว)
                  AnimatedOpacity(
                    duration: const Duration(milliseconds: 200),
                    opacity: _isRefreshing ? 0.5 : 1.0,
                    child: SingleChildScrollView(
                      controller: listScrollController,
                      child: Column(children: _applyClientFilters(transListData).asMap().entries.map((entry) => listObject(entry.value, entry.key, screenWidth)).toList()),
                    ),
                  ),
                  // Loading indicator overlay when refreshing
                  if (_isRefreshing)
                    Positioned.fill(
                      child: Container(
                        color: global.theme.cardColor.withValues(alpha: 0.3),
                        child: const Center(
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                      ),
                    ),
                ],
              ),
            ),
            if (loadingData)
              const Center(
                child: Padding(padding: EdgeInsets.all(16.0), child: CircularProgressIndicator()),
              ),
          ],
        ),
      ),
    );
  }

  Future<int> filterBox(int filterTransaction) async {
    int selectedOption = filterTransaction;
    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (context, setState) {
            return AlertDialog(
              title: Text(global.language("filter")),
              content: SizedBox(
                width: 600.0,
                height: 600.0,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Divider(),
                    Text(global.language("filter_transaction")),
                    RadioListTile(
                      title: Text(global.language("all")),
                      value: 1,
                      groupValue: selectedOption,
                      onChanged: (value) {
                        setState(() {
                          selectedOption = value!;
                        });
                      },
                    ),
                    RadioListTile(
                      title: Text(global.language("pos")),
                      value: 2,
                      groupValue: selectedOption,
                      onChanged: (value) {
                        setState(() {
                          selectedOption = value!;
                        });
                      },
                    ),
                    RadioListTile(
                      title: Text(global.language("merchant")),
                      value: 3,
                      groupValue: selectedOption,
                      onChanged: (value) {
                        setState(() {
                          selectedOption = value!;
                        });
                      },
                    ),
                  ],
                ),
              ),
              actions: [
                ElevatedButton(
                  onPressed: () {
                    Navigator.pop(context);
                  },
                  child: Text(global.language("confirm")),
                ),
                ElevatedButton(
                  onPressed: () {
                    Navigator.pop(context);
                  },
                  child: Text(global.language("cancel")),
                ),
              ],
            );
          },
        );
      },
    );

    return selectedOption;
  }


  Color _getContainerColor(String itemGuid, int index) {
    if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
      return global.theme.rowSelectedColor;
    }
    if (_hoverIndex == index) {
      return global.theme.rowHoverColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  Widget listObject(TransactionModel value, int index, double screenWidth) {
    final isSelected = selectGuid == value.guidfixed;
    TextStyle textStyle = isSelected
        ? TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w700, color: global.theme.textColor)
        : TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w400, color: global.theme.textSecondaryColor);
    DateTime docDateTime = DateTime.parse(value.docdatetime);
    ValueNotifier<bool> isHovered = ValueNotifier<bool>(false);

    final isSmallScreen = isMobile(screenWidth);
    // ใช้ _customFontSize แทน responsive font size
    final rowFontSize = _customFontSize;
    final rowPadding = getResponsivePadding(screenWidth, mobile: 4, tablet: 6, desktop: 8);

    // สำหรับมือถือ ใช้ Card layout
    if (isSmallScreen) {
      return buildMobileCard(value, index, docDateTime, rowFontSize, isHovered);
    }

    // สำหรับ Tablet และ Desktop ใช้ Table layout
    List<Widget> tableDetails = [];

    tableDetails.add(
      Expanded(
        flex: getTableFlexRatio(screenWidth, 1),
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
          child: Text(
            global.formatThaiDateTime(dateTime: docDateTime.toLocal(), showTime: true),
            style: TextStyle(color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor, fontSize: rowFontSize),
          ),
        ),
      ),
    );
    tableDetails.add(
      Expanded(
        flex: getTableFlexRatio(screenWidth, 1),
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
          child: Text(
            value.docno,
            style: TextStyle(color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor, fontSize: rowFontSize),
          ),
        ),
      ),
    );
    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.stockbalance) {
      tableDetails.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 3),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
            child: Text(
              isSmallScreen ? global.activeLangName(value.custnames ?? []) : "${global.activeLangName(value.custnames ?? [])} (${value.custcode})",
              textAlign: TextAlign.left,
              style: TextStyle(color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor, fontSize: rowFontSize),
            ),
          ),
        ),
      );
    }
    if (widget.type == global.TransactionTypeEnum.purchaseorder) {
      Column statusColumn = Column(mainAxisAlignment: MainAxisAlignment.center, crossAxisAlignment: CrossAxisAlignment.start, children: []);

      // แสดงสถานะการอนุมัติ
      final approvalStatus = _approvalStatusMap[value.docno];
      if (approvalStatus != null && approvalStatus.statusText.isNotEmpty) {
        // มีสถานะอนุมัติ - แสดงสถานะจริง
        statusColumn.children.add(
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            margin: const EdgeInsets.only(bottom: 2),
            decoration: BoxDecoration(
              color: Color(approvalStatus.statusColorValue).withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(4),
              border: Border.all(color: Color(approvalStatus.statusColorValue), width: 0.5),
            ),
            child: Text(
              approvalStatus.statusText,
              style: TextStyle(
                color: Color(approvalStatus.statusColorValue),
                fontWeight: FontWeight.bold,
                fontSize: rowFontSize - 1,
              ),
            ),
          ),
        );
      } else if (!value.iscancel) {
        // ไม่มีสถานะอนุมัติและยังไม่ถูกยกเลิก - แสดง "ยังไม่ส่งขออนุมัติ"
        statusColumn.children.add(
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            margin: const EdgeInsets.only(bottom: 2),
            decoration: BoxDecoration(
              color: global.theme.dividerBorderColor,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(color: global.theme.iconSecondaryColor, width: 0.5),
            ),
            child: Text(
              global.language('not_yet_submitted'),
              style: TextStyle(
                color: global.theme.textSecondaryColor,
                fontWeight: FontWeight.bold,
                fontSize: rowFontSize - 1,
              ),
            ),
          ),
        );
      }

      if (value.isref!) {
        statusColumn.children.add(
          Text(
            global.language("referenced"),
            style: TextStyle(color: global.theme.primaryColor, fontWeight: FontWeight.bold, fontSize: rowFontSize),
          ),
        );
        statusColumn.children.add(
          Text(
            (value.isclosed!) ? global.language("fully_received") : global.language("partially_received"),
            style: TextStyle(color: global.theme.positiveHighlightTextColor, fontWeight: FontWeight.bold, fontSize: rowFontSize),
          ),
        );
      }
      if (value.iscancel) {
        statusColumn.children.add(
          Text(
            global.language("cancelled"),
            style: TextStyle(color: global.theme.negativeHighlightTextColor, fontWeight: FontWeight.bold, fontSize: rowFontSize),
          ),
        );
      }
      tableDetails.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
            child: statusColumn,
          ),
        ),
      );
    } else {
      if (widget.type != global.TransactionTypeEnum.stockbalance) {
        tableDetails.add(
          Expanded(
            flex: getTableFlexRatio(screenWidth, 1),
            child: Padding(
              padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
              child: Text(
                value.salename,
                textAlign: TextAlign.left,
                style: TextStyle(color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor, fontSize: rowFontSize),
              ),
            ),
          ),
        );
      }
    }

    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockreceiveproduct &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct &&
        widget.type != global.TransactionTypeEnum.stockbalance) {
      tableDetails.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
            child: Text(
              value.inquirytype == 0 ? global.language('credit') : global.language('cash'),
              textAlign: TextAlign.left,
              style: TextStyle(color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor, fontSize: rowFontSize),
            ),
          ),
        ),
      );
    }
    if (widget.type == global.TransactionTypeEnum.adjust) {
      tableDetails.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
            child: Text(
              value.transflag == 66
                  ? global.language('increase')
                  : value.transflag == 68
                  ? global.language('decrease')
                  : global.language('adjust_cost'),
              textAlign: TextAlign.center,
              style: TextStyle(fontSize: rowFontSize, color: global.theme.textColor),
            ),
          ),
        ),
      );
    }

    if (widget.type == global.TransactionTypeEnum.purchaseorder) {
      tableDetails.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
            child: Text(
              value.detailcount.toString(),
              textAlign: TextAlign.right,
              style: TextStyle(fontSize: rowFontSize, color: global.theme.textColor),
            ),
          ),
        ),
      );
    } else {
      if (widget.type != global.TransactionTypeEnum.stockbalance) {
        tableDetails.add(
          Expanded(
            flex: getTableFlexRatio(screenWidth, 1),
            child: Padding(
              padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
              child: Text(
                value.details!.length.toString(),
                textAlign: TextAlign.center,
                style: TextStyle(color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor, fontSize: rowFontSize),
              ),
            ),
          ),
        );
      } else {
        tableDetails.add(
          Expanded(
            flex: getTableFlexRatio(screenWidth, 1),
            child: Padding(
              padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
              child: Text(
                global.formatNumber(value.totalqty!),
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: rowFontSize, color: global.theme.textColor),
              ),
            ),
          ),
        );
      }
    }

    if (widget.type != global.TransactionTypeEnum.adjust &&
        widget.type != global.TransactionTypeEnum.stocktransfer &&
        widget.type != global.TransactionTypeEnum.stockpickupproduct &&
        widget.type != global.TransactionTypeEnum.stockreturnproduct) {
      tableDetails.add(
        Expanded(
          flex: getTableFlexRatio(screenWidth, 1),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: rowPadding / 2),
            child: Text(
              global.formatNumber(value.totalamount),
              textAlign: TextAlign.right,
              style: TextStyle(color: (value.iscancel == true) ? global.theme.negativeHighlightTextColor : global.theme.textColor, fontSize: rowFontSize),
            ),
          ),
        ),
      );
    }

    return ValueListenableBuilder<bool>(
      valueListenable: isHovered,
      builder: (context, hovered, child) {
        // กำหนดสีพื้นหลังตาม odd/even rows
        Color getBackgroundColor() {
          if (selectGuid == value.guidfixed) {
            return global.theme.rowSelectedColor; // สีเมื่อ selected (ความสำคัญสูงสุด)
          } else if (hovered) {
            return global.theme.rowHoverColor; // สีเมื่อ hover
          } else if (index % 2 == 0) {
            return global.theme.columnAlternateEvenColor; // บรรทัดคู่ (0, 2, 4, ...) - พื้นขาว
          } else {
            return global.theme.columnAlternateOddColor; // บรรทัดคี่ (1, 3, 5, ...) - พื้นเทาอ่อน
          }
        }

        return MouseRegion(
          onEnter: (_) => isHovered.value = true,
          onExit: (_) => isHovered.value = false,
          child: GestureDetector(
            onTap: () {
              reloadDataFromMongoDbAndReturnSelected(value);
            },
            child: Container(
              decoration: BoxDecoration(
                color: getBackgroundColor(),
                border: Border(bottom: BorderSide(width: 1.0, color: global.theme.dividerBorderColor)),
              ),
              padding: EdgeInsets.only(
                left: getResponsivePadding(screenWidth, mobile: 8, tablet: 10, desktop: 10),
                right: getResponsivePadding(screenWidth, mobile: 8, tablet: 10, desktop: 10),
                top: getResponsivePadding(screenWidth, mobile: 8, tablet: 6, desktop: 4),
                bottom: getResponsivePadding(screenWidth, mobile: 8, tablet: 6, desktop: 4),
              ),
              child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: tableDetails),
            ),
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    for (int i = 0; i < transListData.length; i++) {
      if (transListData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }

    // สร้าง content widget ที่ใช้ร่วมกัน
    Widget contentWidget = LayoutBuilder(
      builder: (context, constraints) {
        return BlocListener<TransBloc, TransState>(
          listener: (context, state) {
            // ตรวจสอบว่า widget ยังคง mounted หรือไม่ก่อนทำ setState
            if (!mounted) return;

            // Load Success
            if (state is TransLoadSuccess) {
              AppLogger.debug('🔄 [BlocListener] TransLoadSuccess received, _isAutoRefreshing=$_isAutoRefreshing');
              // ถ้ากำลัง auto-refresh อยู่ ให้ใช้ handler แยก
              if (_isAutoRefreshing) {
                AppLogger.debug('🔄 [BlocListener] Delegating to _handleAutoRefreshResult');
                _handleAutoRefreshResult(state.trans);
                return;
              }

              setState(() {
                loadingData = false;
                isSearching = false;
                totalRecord = state.totalRecord;
                if (state.trans.isNotEmpty) {
                  // ถ้ากำลัง refresh อยู่ ให้ replace list แทน append เพื่อให้ smooth
                  if (_isRefreshing) {
                    transListData = List.from(state.trans);
                  } else {
                    transListData.addAll(state.trans);
                    if (transListData.isNotEmpty) {
                      selectGuid = transListData[0].guidfixed!;
                    } else {
                      selectGuid = "";
                    }
                  }
                } else if (_isRefreshing) {
                  // ถ้า refresh แล้วไม่มีข้อมูล ก็ต้อง clear list
                  transListData.clear();
                }
              });

              // ใช้ postFrameCallback เพื่อให้แน่ใจว่า UI rebuild เสร็จก่อนตั้งค่า cursor
              WidgetsBinding.instance.addPostFrameCallback((_) {
                if (mounted) {
                  _moveCursorToTextEnd();
                }
              });

              // ถ้ากำลัง refresh อยู่ ให้ restore position และ selection
              if (_isRefreshing) {
                _restorePositionAfterRefresh();
                // เคลียร์ loading animation หลังจาก refresh เสร็จ
                if (_pendingLoadingGuid.isNotEmpty) {
                  setState(() {
                    _pendingLoadingGuid = '';
                  });
                }
              }

              // แจ้ง parent widget ว่าหยุด loading แล้ว
              if (widget.onLoadingStateChanged != null) {
                widget.onLoadingStateChanged!(false);
              }

              // เริ่ม auto-refresh หลังจากโหลดข้อมูลเสร็จครั้งแรก (ถ้ายังไม่ได้เริ่ม)
              if (_autoRefreshTimer == null) {
                AppLogger.debug('🔄 [AutoRefresh] Starting auto-refresh timer...');
                _startAutoRefresh();
              }

              // โหลดสถานะการอนุมัติสำหรับ PO และเริ่ม polling
              _loadApprovalStatuses();
              // เริ่ม polling สถานะอนุมัติทุก 5 วินาที (ถ้ายังไม่ได้เริ่ม)
              if (_approvalStatusPollingTimer == null) {
                _startApprovalStatusPolling();
              }

              // หลังจาก response กลับมา ถ้ามี pending search รออยู่ให้ทำการค้นหาใหม่
              Future.microtask(() {
                if (mounted && pendingSearchText != currentSearchRequest && pendingSearchText.isNotEmpty) {
                  loadDataList(pendingSearchText, filterTransaction);
                }
              });
            }

            // Handle error state
            if (state is TransLoadFailed) {
              setState(() {
                loadingData = false;
                isSearching = false;
              });

              _moveCursorToTextEnd();

              // แจ้ง parent widget ว่าหยุด loading แล้ว
              if (widget.onLoadingStateChanged != null) {
                widget.onLoadingStateChanged!(false);
              }

              // เหมือนกับ success case สำหรับ pending search
              Future.microtask(() {
                if (mounted && pendingSearchText != currentSearchRequest && pendingSearchText.isNotEmpty) {
                  loadDataList(pendingSearchText, filterTransaction);
                }
              });
            }
          },
          child: listScreen(screenWidth: constraints.maxWidth),
        );
      },
    );

    // ถ้าเป็น embedded mode ให้ return widget โดยไม่ต้องมี header "select_document"
    if (widget.isEmbedded) {
      return Container(
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          border: Border(left: BorderSide(color: global.theme.dividerBorderColor, width: 1)),
          boxShadow: [BoxShadow(color: global.theme.textColor.withValues(alpha: 0.1), blurRadius: 8, offset: const Offset(-2, 0))],
        ),
        child: contentWidget,
      );
    }

    // โหมดปกติ ใช้ Scaffold
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: PopScope(
        canPop: true,
        onPopInvoked: (didPop) {
          if (didPop) {
            cancelAllOperationsAndReset();
          }
        },
        child: contentWidget,
      ),
    );
  }
}
