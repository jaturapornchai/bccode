// ignore_for_file: depend_on_referenced_packages

import 'dart:async';

import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/report/report_bloc.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/employee_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/report_list_model.dart';
import 'package:smlaicloud/model/select_colums_csv_model.dart';
import 'package:smlaicloud/repositories/report_repository.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';
import 'package:smlaicloud/screen_search/customer_search_screen.dart';
import 'package:smlaicloud/screen_search/employee_search_screen.dart';
import 'package:smlaicloud/screen_search/product_group_search_screen.dart';
import 'package:smlaicloud/screens/report/file_download.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:split_view/split_view.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/columns_csv_list.dart' as columns_csv;
import 'package:smlaicloud/utils/logger/app_logger.dart';

class PdfReportMainScreen extends StatefulWidget {
  const PdfReportMainScreen({super.key});

  @override
  State<PdfReportMainScreen> createState() => PdfReportMainScreenState();
}

class PdfReportMainScreenState extends State<PdfReportMainScreen>
    with SingleTickerProviderStateMixin {
  ReportRepository reportRepository = ReportRepository();

  late TabController tabController;
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];

  String selectCode = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  List<ReportListModel> reportData = [];
  List<ReportListModel> reportDataTeamp = [];
  ReportListModel selectedReport = ReportListModel(
    code: '',
    group: '',
    names: [],
    type: null,
  );
  String modulename = "";

  List<LogDownloadParthModel> logDownloadParth = [];
  late Timer _timer;
  final List<String> reportGroup = [
    'report_filter_all',
    'report_group_product',
    'report_group_debtor',
    'report_group_creditor',
    'report_group_bank',
    'CSV',
  ];
  String selectedFilter = 'report_filter_all';
  List<CompanyBranchModel> companyBranchListData = [];
  final TextEditingController fromCustCodeController = TextEditingController();
  final TextEditingController toCustCodeController = TextEditingController();
  final TextEditingController fromDate = TextEditingController();
  final TextEditingController toDate = TextEditingController();
  final TextEditingController search = TextEditingController();
  final TextEditingController year = TextEditingController();
  final TextEditingController month = TextEditingController();
  final TextEditingController fromsaleController = TextEditingController();
  final TextEditingController tosaleController = TextEditingController();
  final TextEditingController barcodeController = TextEditingController();
  final TextEditingController fromBarcodeController = TextEditingController();
  final TextEditingController toBarcodeController = TextEditingController();
  final TextEditingController fromGroupController = TextEditingController();
  final TextEditingController toGroupController = TextEditingController();

  String fromCustCode = "";
  String toCustCode = "";
  String fromsalecode = "";
  String tosalecode = "";
  String barcode = "";
  String fromBarcode = "";
  String toBarcode = "";
  String fromGroup = "";
  String toGroup = "";

  String selectedBranch = "";
  bool showDetail = false;
  bool showSumByDate = true;
  bool showIscancel = false;
  bool showIscost = false;
  bool ispdf = true;
  String typefile = "pdf";

  List<String> inquirytype = ["inquiry_all", "inquiry_credit", "inquiry_cash"];
  String selectedInquiryType = "inquiry_all";

  List<String> isposType = ["sale_all", "sale_pos", "sale_merchant"];
  String selectedIsposType = "sale_all";

  List<ListColumsCsvModel> selectColumsCsv = [];

  @override
  void initState() {
    selectColumsCsv.addAll(columns_csv.listColumns);
    reportDataTeamp.addAll([
      ReportListModel(
        code: 'DEBTOR-001',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการขาย')],
        type: global.ReportEnum.saleinvoice,
      ),
      ReportListModel(
        code: 'DEBTOR-002',
        group: global.language('debtor'),
        names: [
          LanguageDataModel(code: 'th', name: 'รายงานการขาย แสดงรายละเอียด'),
        ],
        type: global.ReportEnum.saleinvoicedetail,
      ),
      ReportListModel(
        code: 'DEBTOR-003',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการขาย ตามสินค้า')],
        type: global.ReportEnum.salebyproduct,
      ),
      ReportListModel(
        code: 'DEBTOR-004',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการขาย ตามวัน')],
        type: global.ReportEnum.salebydate,
      ),
      ReportListModel(
        code: 'DEBTOR-005',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานรับเงินตามวัน')],
        type: global.ReportEnum.receivemoney,
      ),
      ReportListModel(
        code: 'DEBTOR-006',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานรับคืน/ลดหนี้')],
        type: global.ReportEnum.saleinvoicereturn,
      ),
      ReportListModel(
        code: 'DEBTOR-007',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการรับเงิน')],
        type: global.ReportEnum.paid,
      ),
      ReportListModel(
        code: 'DEBTOR-008',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการรับชำระ')],
        type: global.ReportEnum.getpaid,
      ),
      ReportListModel(
        code: 'DEBTOR-009',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานภาษีขาย')],
        type: global.ReportEnum.vatsale,
      ),
      ReportListModel(
        code: 'DEBTOR-010',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานซื้อสะสมตามสมาชิก')],
        type: global.ReportEnum.salebydebtor,
      ),
      ReportListModel(
        code: 'DEBTOR-011',
        group: global.language('debtor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานลูกหนี้')],
        type: global.ReportEnum.debtor,
      ),
      ReportListModel(
        code: 'CREDITOR-001',
        group: global.language('creditor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานเจ้าหนี้')],
        type: global.ReportEnum.creditor,
      ),
      ReportListModel(
        code: 'CREDITOR-002',
        group: global.language('creditor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการซื้อสินค้า')],
        type: global.ReportEnum.purchase,
      ),
      ReportListModel(
        code: 'CREDITOR-003',
        group: global.language('creditor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานส่งคืน/ลดหนี้')],
        type: global.ReportEnum.purchasereturn,
      ),
      ReportListModel(
        code: 'CREDITOR-004',
        group: global.language('creditor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการจ่ายเงิน')],
        type: global.ReportEnum.pay,
      ),
      ReportListModel(
        code: 'CREDITOR-005',
        group: global.language('creditor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการจ่ายชำระ')],
        type: global.ReportEnum.getpay,
      ),
      ReportListModel(
        code: 'CREDITOR-006',
        group: global.language('creditor'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานภาษีซื้อ')],
        type: global.ReportEnum.vatpurchase,
      ),
      ReportListModel(
        code: 'PRODUCT-001',
        group: global.language('product'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานสินค้า')],
        type: global.ReportEnum.product,
      ),
      ReportListModel(
        code: 'PRODUCT-002',
        group: global.language('product'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการโอนสินค้า')],
        type: global.ReportEnum.transfer,
      ),
      ReportListModel(
        code: 'PRODUCT-003',
        group: global.language('product'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการรับสินค้า')],
        type: global.ReportEnum.receive,
      ),
      ReportListModel(
        code: 'PRODUCT-004',
        group: global.language('product'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการเบิกสินค้า')],
        type: global.ReportEnum.pickup,
      ),
      ReportListModel(
        code: 'PRODUCT-005',
        group: global.language('product'),
        names: [
          LanguageDataModel(code: 'th', name: 'รายงานรับคืนจากเบิกสินค้า'),
        ],
        type: global.ReportEnum.returnproduct,
      ),
      ReportListModel(
        code: 'PRODUCT-006',
        group: global.language('product'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานการปรับปรุงสต็อก')],
        type: global.ReportEnum.stockadjustment,
      ),
      ReportListModel(
        code: 'PRODUCT-007',
        group: global.language('product'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานสินค้าคงเหลือ')],
        type: global.ReportEnum.stockbalance,
      ),
      ReportListModel(
        code: 'PRODUCT-008',
        group: global.language('product'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานเคลือนไหวสินค้า')],
        type: global.ReportEnum.productmovement,
      ),
      ReportListModel(
        code: 'PRODUCT-009',
        group: global.language('product'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานบัญชีคุมพิเศษ')],
        type: global.ReportEnum.stockcard,
      ),
      ReportListModel(
        code: 'BANK-001',
        group: global.language('bank'),
        names: [LanguageDataModel(code: 'th', name: 'รายงานสมุดบัญชีธนาคาร')],
        type: global.ReportEnum.bookbank,
      ),
      ReportListModel(
        code: 'CSV-001',
        group: 'CSV',
        names: [LanguageDataModel(code: 'th', name: 'ข้อมูลการขาย')],
        type: global.ReportEnum.csvsaledetail,
      ),
    ]);

    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );

    listScrollController.addListener(onScrollList);

    loadDataList("");
    loadDataBranch();

    super.initState();
  }

  /// start timer
  Future<void> startTimer() async {
    _timer = Timer.periodic(
      const Duration(seconds: 3),
      (Timer t) => checkFileStatus(),
    );
  }

  void loadDataList(String search) {
    setState(() {
      currentListIndex = -1;
      if (search.isNotEmpty) {
        reportData = reportDataTeamp
            .where(
              (element) =>
                  element.code.toLowerCase().contains(search.toLowerCase()) ||
                  global
                      .activeLangName(element.names)
                      .toLowerCase()
                      .contains(search.toLowerCase()),
            )
            .toList();
      } else {
        reportData = reportDataTeamp;
      }
      selectCode = reportData.first.code;
      getData(reportData.first.code);
    });
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchController.text);
    }
  }

  void loadDataBranch() {
    context.read<CompanyBranchBloc>().add(
      const CompanyBranchLoadList(offset: 0, limit: 100, search: ""),
    );
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    searchController.dispose();
    _timer.cancel();

    super.dispose();
  }

  void getData(String guid) async {
    await startTimer();

    setState(() {
      fromDate.text = DateFormat('dd/MM/yyyy').format(DateTime.now());
      toDate.text = DateFormat('dd/MM/yyyy').format(DateTime.now());
      search.text = "";
      year.text = DateTime.now().year.toString();
      month.text = DateTime.now().month.toString();

      selectedReport = reportDataTeamp.firstWhere(
        (element) => element.code == guid,
      );

      fromCustCode = "";
      toCustCode = "";
      selectedBranch = "";
      fromCustCodeController.text = "";
      toCustCodeController.text = "";
      fromsaleController.text = "";
      tosaleController.text = "";
      fromsalecode = "";
      tosalecode = "";
      barcodeController.text = "";
      fromBarcodeController.text = "";
      toBarcodeController.text = "";
      barcode = "";
      fromBarcode = "";
      toBarcode = "";
      fromGroupController.text = "";
      toGroupController.text = "";
      fromGroup = "";
      toGroup = "";

      if (selectedReport.type == global.ReportEnum.saleinvoice) {
        /// รายงานกาาขาย
        modulename = "saleinvoice";
        showDetail = false;
      } else if (selectedReport.type == global.ReportEnum.saleinvoicedetail) {
        /// รายงานการขาย แสดงรายละเอียด
        modulename = "saleinvoicedetail";
        showDetail = true;
      } else if (selectedReport.type == global.ReportEnum.salebydate) {
        /// รายงานยอดขายตามวัน
        modulename = "salebydate";
      } else if (selectedReport.type == global.ReportEnum.receivemoney) {
        /// รายงานรับเงิน
        modulename = "receivemoney";
      } else if (selectedReport.type == global.ReportEnum.product) {
        /// รายงานสินค้า
        modulename = "product";
      } else if (selectedReport.type == global.ReportEnum.debtor) {
        /// รายงานลูกหนี้
        modulename = "debtor";
      } else if (selectedReport.type == global.ReportEnum.creditor) {
        /// รายงานเจ้าหนี้
        modulename = "creditor";
      } else if (selectedReport.type == global.ReportEnum.bookbank) {
        /// รายงานสมุดบัญชีธนาคาร
        modulename = "bookbank";
      } else if (selectedReport.type == global.ReportEnum.purchase) {
        /// รายงานการซื้อสินค้า
        modulename = "purchase";
      } else if (selectedReport.type == global.ReportEnum.purchasereturn) {
        /// รายงานส่งคืน/ลดหนี้
        modulename = "purchasereturn";
      } else if (selectedReport.type == global.ReportEnum.saleinvoicereturn) {
        /// รายงานรับคืน/ลดหนี้
        modulename = "saleinvoicereturn";
      } else if (selectedReport.type == global.ReportEnum.transfer) {
        /// รายงานการโอนสินค้า
        modulename = "transfer";
      } else if (selectedReport.type == global.ReportEnum.receive) {
        /// รายงานการรับสินค้า
        modulename = "receive";
      } else if (selectedReport.type == global.ReportEnum.pickup) {
        /// รายงานการเบิกสินค้า
        modulename = "pickup";
      } else if (selectedReport.type == global.ReportEnum.returnproduct) {
        /// รายงานรับคืนจากเบิกสินค้า
        modulename = "returnproduct";
      } else if (selectedReport.type == global.ReportEnum.stockadjustment) {
        /// รายงานการปรับปรุงสต็อก
        modulename = "stockadjustment";
      } else if (selectedReport.type == global.ReportEnum.paid) {
        /// รายงานการรับเงิน
        modulename = "paid";
      } else if (selectedReport.type == global.ReportEnum.pay) {
        /// รายงานการจ่ายเงิน
        modulename = "pay";
      } else if (selectedReport.type == global.ReportEnum.getpaid) {
        /// รายงานการรับชำระ
        modulename = "getpaid";
      } else if (selectedReport.type == global.ReportEnum.getpay) {
        /// รายงานการจ่ายชำระ
        modulename = "getpay";
      } else if (selectedReport.type == global.ReportEnum.vatsale) {
        /// รายงานภาษีขาย
        modulename = "vatsale";
      } else if (selectedReport.type == global.ReportEnum.vatpurchase) {
        /// รายงานภาษีซื้อ
        modulename = "vatpurchase";
      } else if (selectedReport.type == global.ReportEnum.salebydebtor) {
        /// รายงานซื้อสะสมตามสมาชิก
        modulename = "salebydebtor";
      } else if (selectedReport.type == global.ReportEnum.salebyproduct) {
        /// รายงานการขาย ตามสินค้า
        modulename = "salebyproduct";
      } else if (selectedReport.type == global.ReportEnum.productmovement) {
        /// รายงานเคลือนไหวสินค้า
        modulename = "productmovement";
      } else if (selectedReport.type == global.ReportEnum.stockcard) {
        /// รายงานบัญชีคุมพิเศษ
        modulename = "stockcard";
      } else if (selectedReport.type == global.ReportEnum.csvsaledetail) {
        /// รายงานข้อมูลการขาย
        modulename = "csvsaledetail";
      }
      // _loadDownloadPath();

      getLogDownloadPath();
    });

    if (currentListIndex >= 0) {
      RenderBox? boxHeader =
          headerKey.currentContext?.findRenderObject() as RenderBox?;
      Offset? positionHeader = boxHeader?.localToGlobal(Offset.zero);
      RenderBox? box =
          listKeys[currentListIndex].currentContext?.findRenderObject()
              as RenderBox?;
      Offset? position = box?.localToGlobal(Offset.zero);
      if (position != null &&
          positionHeader != null &&
          boxHeader != null &&
          box != null) {
        // Scroll Up
        if (isKeyUp &&
            position.dy <=
                (positionHeader.dy +
                    (boxHeader.size.height + (box.size.height * 2)))) {
          setState(() {
            listScrollController.animateTo(
              listScrollController.offset -
                  (boxHeader.size.height + box.size.height),
              duration: const Duration(milliseconds: 100),
              curve: Curves.ease,
            );
            isKeyUp = false;
          });
        }
        // Scroll Down
        if (isKeyDown && position.dy > (queryData.size.height - 100)) {
          setState(() {
            listScrollController.animateTo(
              listScrollController.offset +
                  (position.dy - (queryData.size.height - 100)),
              duration: const Duration(milliseconds: 100),
              curve: Curves.easeOut,
            );
            isKeyDown = false;
          });
        }
      }
    }
  }

  Future<void> getLogDownloadPath() async {
    context.read<ReportBloc>().add(
      FileStatusGetList(
        offset: 0,
        limit: 100,
        menu: selectedReport.type.toString(),
      ),
    );
  }

  void updateLogDownloadParth(LogDownloadParthModel logDownloadParthModel) {
    context.read<ReportBloc>().add(
      FileStatusUpdate(
        logDownloadParthModel: logDownloadParthModel,
        guid: logDownloadParthModel.guidfixed!,
      ),
    );
  }

  void removeAllLogDownloadParth() {
    context.read<ReportBloc>().add(
      FileStatusDeleteByMenu(menu: selectedReport.type.toString()),
    );
  }

  Future<void> removeLogDownloadParth(String guid) async {
    context.read<ReportBloc>().add(FileStatusDeleteById(guid: guid));
  }

  void selectFromDate(BuildContext context) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: DateTime.now(),
      firstDate: DateTime(2020),
      lastDate: DateTime.now(),
    );

    if (pickedDate != null) {
      setState(() {
        fromDate.text = DateFormat('dd/MM/yyyy').format(pickedDate);
      });
    }
  }

  void selectToDate(BuildContext context) async {
    DateTime? pickedDate;

    if (selectedReport.type == global.ReportEnum.saleinvoicedetail &&
        fromDate.text.isNotEmpty) {
      final DateTime fromDateValue = DateFormat(
        'dd/MM/yyyy',
      ).parse(fromDate.text);
      final DateTime startDate = fromDateValue;
      final DateTime endDate = fromDateValue.add(const Duration(days: 30));

      pickedDate = await showDatePicker(
        context: context,
        initialDate: fromDateValue,
        firstDate: startDate,
        lastDate: endDate,
      );
    } else {
      pickedDate = await showDatePicker(
        context: context,
        initialDate: DateTime.now(),
        firstDate: DateTime(2020),
        lastDate: DateTime.now(),
      );
    }

    if (pickedDate != null) {
      setState(() {
        toDate.text = DateFormat('dd/MM/yyyy').format(pickedDate!);
      });
    }
  }

  void searchCustomer(String type) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const CustomerSearchScreen(word: ''),
      ),
    ).then((value) {
      global.SearchDebtorModel result = value;
      if (result.code.trim().isNotEmpty) {
        setState(() {
          if (type == "fromCustCode") {
            fromCustCodeController.text =
                "${result.code} ~ ${global.activeLangName(result.names)}";
            fromCustCode = result.code;
          } else {
            toCustCodeController.text =
                "${result.code} ~ ${global.activeLangName(result.names)}";
            toCustCode = result.code;
          }
        });
      }
    });
  }

  void searchBarcode(String type) {
    String fromscreen = "";
    if (kDebugMode) {
      AppLogger.debug(selectedReport.type);
    }
    if (selectedReport.type == global.ReportEnum.stockbalance ||
        selectedReport.type == global.ReportEnum.stockcard) {
      if (type == "barcode" || type == "fromBarcode" || type == "toBarcode") {
        /// ค้นหาสินค้าที่ไม่มี refbarcode และสินค้าชุด
        fromscreen = "stockbalance";
      }
    }
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BarcodeSearchScreen(word: '', screen: fromscreen),
      ),
    ).then((value) {
      ProductBarcodeModel result = value;
      if (result.barcode!.trim().isNotEmpty) {
        setState(() {
          if (type == "fromBarcode") {
            fromBarcodeController.text =
                "${result.barcode} ~ ${global.activeLangName(result.names!)}";
            fromBarcode = result.barcode!;
          } else if (type == "toBarcode") {
            toBarcodeController.text =
                "${result.barcode} ~ ${global.activeLangName(result.names!)}";
            toBarcode = result.barcode!;
          } else if (type == "barcode") {
            barcodeController.text =
                "${result.barcode} ~ ${global.activeLangName(result.names!)}";
            barcode = result.barcode!;
          }
        });
      }
    });
  }

  void onClearSearchField(String type) {
    if (type == "fromBarcode") {
      fromBarcodeController.text = "";
      fromBarcode = "";
    } else if (type == "toBarcode") {
      toBarcodeController.text = "";
      toBarcode = "";
    } else if (type == "barcode") {
      barcodeController.text = "";
      barcode = "";
    } else if (type == "tosale") {
      tosaleController.text = "";
      tosalecode = "";
    } else if (type == "fromsale") {
      fromsaleController.text = "";
      fromsalecode = "";
    } else if (type == "fromCustCode") {
      fromCustCodeController.text = "";
      fromCustCode = "";
    } else if (type == "toCustCode") {
      toCustCodeController.text = "";
      toCustCode = "";
    } else if (type == "fromGroup") {
      fromGroupController.text = "";
      fromGroup = "";
    } else if (type == "toGroup") {
      toGroupController.text = "";
      toGroup = "";
    }
  }

  void searchGroupProduct(String type) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const ProductGroupSearchScreen(word: ''),
      ),
    ).then((value) {
      global.SearchCodeNameModel result = value;
      if (result.code.trim().isNotEmpty) {
        setState(() {
          if (type == "fromGroup") {
            fromGroupController.text =
                "${result.code} ~ ${global.activeLangName(result.names)}";
            fromGroup = result.code;
          } else {
            toGroupController.text =
                "${result.code} ~ ${global.activeLangName(result.names)}";
            toGroup = result.code;
          }
        });
      }
    });
  }

  void searchSale(String type) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const EmployeeSearchScreen(word: ''),
      ),
    ).then((value) {
      EmployeeModel result = value;
      if (result.code.trim().isNotEmpty) {
        setState(() {
          if (type == "fromsale") {
            fromsaleController.text = "${result.code} ~ ${result.name}";
            fromsalecode = result.code;
          } else {
            tosaleController.text = "${result.code} ~ ${result.name}";
            tosalecode = result.code;
          }
        });
      }
    });
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('report')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            _timer.cancel();
            global.gotoMainMenu(context);
          },
        ),
      ),
      body: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(5),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(2),
            ),
            child: Row(
              children: [
                Expanded(
                  child: TextFormField(
                    onFieldSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      setState(() {
                        reportData = [];
                      });
                      loadDataList(value);
                    },
                    autofocus: false,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.only(
                        top: 0,
                        bottom: 0,
                        left: 0,
                        right: 0,
                      ),
                      border: InputBorder.none,
                      hintText: global.language('search'),
                    ),
                  ),
                ),
                IconButton(
                  focusNode: FocusNode(skipTraversal: true),
                  icon: const FaIcon(FontAwesomeIcons.font),
                  onPressed: () async {
                    setState(() {
                      global.listDataFontSizeChange();
                    });
                  },
                ),
                IconButton(
                  focusNode: FocusNode(skipTraversal: true),
                  icon: const Icon(Icons.line_weight),
                  onPressed: () async {
                    setState(() {
                      global.listDataLineSpaceChange();
                    });
                  },
                ),
              ],
            ),
          ),
          Container(color: global.theme.appBarColor, height: 6),
          Center(
            child: Wrap(
              spacing: 5.0,
              children: reportGroup.map((String report) {
                return Padding(
                  padding: EdgeInsets.all(5.0),
                  child: FilterChip(
                    label: Text(global.language(report)),
                    selected: selectedFilter == report,
                    onSelected: (bool selected) {
                      setState(() {
                        if (selected) {
                          selectedFilter = report;

                          if (report == 'report_filter_all') {
                            reportData = reportDataTeamp;
                          } else {
                            reportData = reportDataTeamp
                                .where(
                                  (element) =>
                                      element.group == global.language(report),
                                )
                                .toList();
                          }
                        } else if (selectedFilter == report) {
                          selectedFilter = 'report_filter_all';
                          reportData = reportDataTeamp;
                        }
                      });
                    },
                  ),
                );
              }).toList(),
            ),
          ),
          Container(
            color: global.theme.columnHeaderColor,
            key: headerKey,
            padding: const EdgeInsets.only(
              left: 10,
              right: 10,
              top: 5,
              bottom: 5,
            ),
            child: Row(
              children: [
                Expanded(
                  flex: 1,
                  child: Text(
                    global.language("report_code"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Text(
                    global.language("report_name"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Text(
                    global.language("report_group"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: RawKeyboardListener(
              autofocus: true,
              focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
              child: ListView(
                controller: listScrollController,
                children: reportData
                    .map(
                      (value) => listObject(reportData.indexOf(value), value),
                    )
                    .toList(),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget buildWigetReport(ReportListModel value) {
    switch (value.type) {
      case global.ReportEnum.saleinvoice:
        return buildSaleInvoiceForm();
      case global.ReportEnum.saleinvoicedetail:
        return buildSaleInvoiceDetailForm();
      case global.ReportEnum.salebydate:
        return buildSaleByDateForm();
      case global.ReportEnum.receivemoney:
        return buildReceiveMoneyForm();
      case global.ReportEnum.product:
        return buildProductForm();
      case global.ReportEnum.debtor:
        return buildDebtorForm();
      case global.ReportEnum.creditor:
        return buildCreditorForm();
      case global.ReportEnum.bookbank:
        return buildBookBankForm();
      case global.ReportEnum.purchase:
        return buildPurchaseForm();
      case global.ReportEnum.purchasereturn:
        return buildPurchaseReturnForm();
      case global.ReportEnum.saleinvoicereturn:
        return buildSaleInvoiceReturnForm();
      case global.ReportEnum.transfer:
        return buildTransferForm();
      case global.ReportEnum.receive:
        return buildReceiveForm();
      case global.ReportEnum.pickup:
        return buildPickupForm();
      case global.ReportEnum.returnproduct:
        return buildReturnProductForm();
      case global.ReportEnum.stockadjustment:
        return buildStockAdjustmentForm();
      case global.ReportEnum.paid:
        return buildPaidForm();
      case global.ReportEnum.pay:
        return buildPayForm();
      case global.ReportEnum.getpaid:
        return buildGetPaidForm();
      case global.ReportEnum.getpay:
        return buildGetPayForm();
      case global.ReportEnum.vatsale:
        return buildVatSaleForm();
      case global.ReportEnum.vatpurchase:
        return buildVatPurchaseForm();
      case global.ReportEnum.salebydebtor:
        return buildSaleByDebtorForm();
      case global.ReportEnum.salebyproduct:
        return buildsalebyproductForm();
      case global.ReportEnum.productmovement:
        return buildProductMovementForm();
      case global.ReportEnum.stockbalance:
        return buildStockBalanceForm();
      case global.ReportEnum.stockcard:
        return buildStockCardForm();
      case global.ReportEnum.csvsaledetail:
        return buildCsvSaleDetailForm();
      default:
        return Container();
    }
  }

  Widget buildVatSaleForm() {
    List<dynamic> monthsThai = [
      {"code": 1, "name": "month_january"},
      {"code": 2, "name": "month_february"},
      {"code": 3, "name": "month_march"},
      {"code": 4, "name": "month_april"},
      {"code": 5, "name": "month_may"},
      {"code": 6, "name": "month_june"},
      {"code": 7, "name": "month_july"},
      {"code": 8, "name": "month_august"},
      {"code": 9, "name": "month_september"},
      {"code": 10, "name": "month_october"},
      {"code": 11, "name": "month_november"},
      {"code": 12, "name": "month_december"},
    ];

    /// list year now + 10 and -10
    List<int> years = [];
    for (int i = DateTime.now().year - 10; i <= DateTime.now().year + 10; i++) {
      years.add(i);
    }

    return Column(
      children: [
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              /// select year
              Expanded(
                child: DropdownButtonFormField<int>(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("vat_year"),
                  ),
                  initialValue: DateTime.now().year,
                  items: years.map((item) {
                    return DropdownMenuItem<int>(
                      value: item,
                      child: Text((item + 543).toString()),
                    );
                  }).toList(),
                  onChanged: (value) {
                    setState(() {
                      year.text = value.toString();
                    });
                  },
                ),
              ),

              const SizedBox(width: 5),

              /// select month
              Expanded(
                child: DropdownButtonFormField<int>(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("vat_month"),
                  ),
                  initialValue: DateTime.now().month,
                  items: monthsThai.map((item) {
                    return DropdownMenuItem<int>(
                      value: item["code"],
                      child: Text(global.language(item["name"])),
                    );
                  }).toList(),
                  onChanged: (value) {
                    setState(() {
                      month.text = value.toString();
                    });
                  },
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget buildVatPurchaseForm() {
    List<dynamic> monthsThai = [
      {"code": 1, "name": "month_january"},
      {"code": 2, "name": "month_february"},
      {"code": 3, "name": "month_march"},
      {"code": 4, "name": "month_april"},
      {"code": 5, "name": "month_may"},
      {"code": 6, "name": "month_june"},
      {"code": 7, "name": "month_july"},
      {"code": 8, "name": "month_august"},
      {"code": 9, "name": "month_september"},
      {"code": 10, "name": "month_october"},
      {"code": 11, "name": "month_november"},
      {"code": 12, "name": "month_december"},
    ];

    /// list year now + 10 and -10
    List<int> years = [];
    for (int i = DateTime.now().year - 10; i <= DateTime.now().year + 10; i++) {
      years.add(i);
    }

    return Column(
      children: [
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              /// select year
              Expanded(
                child: DropdownButtonFormField<int>(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("vat_year"),
                  ),
                  initialValue: DateTime.now().year,
                  items: years.map((item) {
                    return DropdownMenuItem<int>(
                      value: item,
                      child: Text((item + 543).toString()),
                    );
                  }).toList(),
                  onChanged: (value) {
                    setState(() {
                      selectedReport.code =
                          "VATPURCHASE-${DateTime.now().month}-$value";
                    });
                  },
                ),
              ),

              const SizedBox(width: 5),

              /// select month
              Expanded(
                child: DropdownButtonFormField<int>(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("vat_month"),
                  ),
                  initialValue: int.parse(selectedReport.code.split("-")[1]),
                  items: monthsThai.map((item) {
                    return DropdownMenuItem<int>(
                      value: item["code"],
                      child: Text(global.language(item["name"])),
                    );
                  }).toList(),
                  onChanged: (value) {
                    setState(() {
                      selectedReport.code =
                          "VATPURCHASE-$value-${DateTime.now().year}";
                    });
                  },
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget buildSaleByDebtorForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("debtor_code"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildGetPayForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildGetPaidForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildPayForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
      ],
    );
  }

  Widget buildPaidForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Row(
          children: [
            Expanded(
              child: buildSearchField(
                "from_customer_code",
                fromCustCodeController,
                (context) => searchCustomer("fromCustCode"),
                (context) => onClearSearchField("fromCustCode"),
              ),
            ),
            const SizedBox(width: 5),
            Expanded(
              child: buildSearchField(
                "to_customer_code",
                toCustCodeController,
                (context) => searchCustomer("toCustCode"),
                (context) => onClearSearchField("toCustCode"),
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        buildBranchSelectionWidget(
          companyBranchListData:
              companyBranchListData, // Replace with your actual data
          selectedBranch: selectedBranch, // The currently selected branch code
          onBranchSelected: (String code) {
            setState(() {
              if (selectedBranch != code) {
                selectedBranch = code;
              } else {
                selectedBranch = '';
              }
            });
          },
        ),
        buildCheckboxListTile("show_sum_by_date", showSumByDate, (value) {
          setState(() => showSumByDate = value!);
        }),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildStockAdjustmentForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildReturnProductForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildPickupForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildReceiveForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildTransferForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildSaleInvoiceReturnForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildPurchaseReturnForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("docno"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildPurchaseForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_supplier_code",
                  fromCustCodeController,
                  (context) => searchCustomer("fromCustCode"),
                  (context) => onClearSearchField("fromCustCode"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_supplier_code",
                  toCustCodeController,
                  (context) => searchCustomer("toCustCode"),
                  (context) => onClearSearchField("toCustCode"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_sale",
                  fromsaleController,
                  (context) => searchSale("fromsale"),
                  (context) => onClearSearchField("fromsale"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_sale",
                  tosaleController,
                  (context) => searchSale("tosale"),
                  (context) => onClearSearchField("tosale"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDropDownField(
                  labelfield: "inquiry_type",
                  value: selectedInquiryType,
                  items: inquirytype,
                  onChanged: (value) {
                    setState(() {
                      selectedInquiryType = value;
                    });
                  },
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        buildBranchSelectionWidget(
          companyBranchListData: companyBranchListData,
          selectedBranch: selectedBranch,
          onBranchSelected: (String code) {
            setState(() {
              if (selectedBranch != code) {
                selectedBranch = code;
              } else {
                selectedBranch = '';
              }
            });
          },
        ),
        buildCheckboxListTile("show_sum_by_date", showSumByDate, (value) {
          setState(() => showSumByDate = value!);
        }),
        buildCheckboxListTile("show_detail", showDetail, (value) {
          setState(() => showDetail = value!);
        }),
        buildCheckboxListTile("show_iscancel", showIscancel, (value) {
          setState(() => showIscancel = value!);
        }),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildBookBankForm() {
    return Column(
      children: [
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("bookbank_code"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildCreditorForm() {
    return Column(
      children: [
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("creditor_code"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildDebtorForm() {
    return Column(
      children: [
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("debtor_code"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildProductForm() {
    return Column(
      children: [
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            decoration: InputDecoration(
              border: OutlineInputBorder(),
              labelText: global.language("barcode"),
            ),
            controller: search,
            onChanged: (value) {},
          ),
        ),
      ],
    );
  }

  Widget buildReceiveMoneyForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        const SizedBox(height: 10),
        buildBranchSelectionWidget(
          companyBranchListData: companyBranchListData,
          selectedBranch: selectedBranch,
          onBranchSelected: (String code) {
            setState(() {
              if (selectedBranch != code) {
                selectedBranch = code;
              } else {
                selectedBranch = '';
              }
            });
          },
        ),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildSaleByDateForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDropDownField(
                  labelfield: "inquiry_type",
                  value: selectedInquiryType,
                  items: inquirytype,
                  onChanged: (value) {
                    setState(() {
                      selectedInquiryType = value;
                    });
                  },
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildDropDownField(
                  labelfield: "sale_type",
                  value: selectedIsposType,
                  items: isposType,
                  onChanged: (value) {
                    setState(() {
                      selectedIsposType = value;
                    });
                  },
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        buildBranchSelectionWidget(
          companyBranchListData: companyBranchListData,
          selectedBranch: selectedBranch,
          onBranchSelected: (String code) {
            setState(() {
              if (selectedBranch != code) {
                selectedBranch = code;
              } else {
                selectedBranch = '';
              }
            });
          },
        ),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildsalebyproductForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_barcode",
                  fromBarcodeController,
                  (context) => searchBarcode("fromBarcode"),
                  (context) => onClearSearchField("fromBarcode"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_barcode",
                  toBarcodeController,
                  (context) => searchBarcode("toBarcode"),
                  (context) => onClearSearchField("toBarcode"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_group",
                  fromGroupController,
                  (context) => searchGroupProduct("fromGroup"),
                  (context) => onClearSearchField("fromGroup"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_group",
                  toGroupController,
                  (context) => searchGroupProduct("toGroup"),
                  (context) => onClearSearchField("toGroup"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        buildBranchSelectionWidget(
          companyBranchListData: companyBranchListData,
          selectedBranch: selectedBranch,
          onBranchSelected: (String code) {
            setState(() {
              if (selectedBranch != code) {
                selectedBranch = code;
              } else {
                selectedBranch = '';
              }
            });
          },
        ),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildSaleInvoiceForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_customer_code",
                  fromCustCodeController,
                  (context) => searchCustomer("fromCustCode"),
                  (context) => onClearSearchField("fromCustCode"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_customer_code",
                  toCustCodeController,
                  (context) => searchCustomer("toCustCode"),
                  (context) => onClearSearchField("toCustCode"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_sale",
                  fromsaleController,
                  (context) => searchSale("fromsale"),
                  (context) => onClearSearchField("fromsale"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_sale",
                  tosaleController,
                  (context) => searchSale("tosale"),
                  (context) => onClearSearchField("tosale"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDropDownField(
                  labelfield: "inquiry_type",
                  value: selectedInquiryType,
                  items: inquirytype,
                  onChanged: (value) {
                    setState(() {
                      selectedInquiryType = value;
                    });
                  },
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildDropDownField(
                  labelfield: "sale_type",
                  value: selectedIsposType,
                  items: isposType,
                  onChanged: (value) {
                    setState(() {
                      selectedIsposType = value;
                    });
                  },
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        buildBranchSelectionWidget(
          companyBranchListData: companyBranchListData,
          selectedBranch: selectedBranch,
          onBranchSelected: (String code) {
            setState(() {
              if (selectedBranch != code) {
                selectedBranch = code;
              } else {
                selectedBranch = '';
              }
            });
          },
        ),
        buildCheckboxListTile("show_sum_by_date", showSumByDate, (value) {
          setState(() => showSumByDate = value!);
        }),
        // buildCheckboxListTile("show_detail", showDetail, (value) {
        //   setState(() => showDetail = value!);
        // }),
        buildCheckboxListTile("show_iscancel", showIscancel, (value) {
          setState(() => showIscancel = value!);
        }),
        Container(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: Row(
            children: [
              Switch(
                value: ispdf,
                activeThumbColor: Colors.red,
                inactiveThumbColor: Colors.green,
                onChanged: (bool value) {
                  setState(() {
                    ispdf = value;
                    if (ispdf) {
                      typefile = "pdf";
                    } else {
                      typefile = "excel";
                    }
                  });
                },
              ),
              Text(
                global.language(
                  (ispdf) ? "file_download_pdf" : "file_download_excel",
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget buildSaleInvoiceDetailForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_customer_code",
                  fromCustCodeController,
                  (context) => searchCustomer("fromCustCode"),
                  (context) => onClearSearchField("fromCustCode"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_customer_code",
                  toCustCodeController,
                  (context) => searchCustomer("toCustCode"),
                  (context) => onClearSearchField("toCustCode"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_sale",
                  fromsaleController,
                  (context) => searchSale("fromsale"),
                  (context) => onClearSearchField("fromsale"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_sale",
                  tosaleController,
                  (context) => searchSale("tosale"),
                  (context) => onClearSearchField("tosale"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDropDownField(
                  labelfield: "inquiry_type",
                  value: selectedInquiryType,
                  items: inquirytype,
                  onChanged: (value) {
                    setState(() {
                      selectedInquiryType = value;
                    });
                  },
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildDropDownField(
                  labelfield: "sale_type",
                  value: selectedIsposType,
                  items: isposType,
                  onChanged: (value) {
                    setState(() {
                      selectedIsposType = value;
                    });
                  },
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        buildBranchSelectionWidget(
          companyBranchListData: companyBranchListData,
          selectedBranch: selectedBranch,
          onBranchSelected: (String code) {
            setState(() {
              if (selectedBranch != code) {
                selectedBranch = code;
              } else {
                selectedBranch = '';
              }
            });
          },
        ),
        buildCheckboxListTile("show_sum_by_date", showSumByDate, (value) {
          setState(() => showSumByDate = value!);
        }),
        // buildCheckboxListTile("show_detail", showDetail, (value) {
        //   setState(() => showDetail = value!);
        // }),
        buildCheckboxListTile("show_iscancel", showIscancel, (value) {
          setState(() => showIscancel = value!);
        }),
        Container(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: Row(
            children: [
              Switch(
                value: ispdf,
                activeThumbColor: Colors.red,
                inactiveThumbColor: Colors.green,
                onChanged: (bool value) {
                  setState(() {
                    ispdf = value;
                    if (ispdf) {
                      typefile = "pdf";
                    } else {
                      typefile = "excel";
                    }
                  });
                },
              ),
              Text(
                global.language(
                  (ispdf) ? "file_download_pdf" : "file_download_excel",
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget buildStockBalanceForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              // Expanded(child: buildDateField("from_date", fromDate, selectFromDate)),
              // const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildSearchField(
                  "from_barcode",
                  fromBarcodeController,
                  (context) => searchBarcode("fromBarcode"),
                  (context) => onClearSearchField("fromBarcode"),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: buildSearchField(
                  "to_barcode",
                  toBarcodeController,
                  (context) => searchBarcode("toBarcode"),
                  (context) => onClearSearchField("toBarcode"),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        buildCheckboxListTile("show_iscost", showIscost, (value) {
          setState(() => showIscost = value!);
        }),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildProductMovementForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: buildSearchField(
            "barcode",
            barcodeController,
            (context) => searchBarcode("barcode"),
            (context) => onClearSearchField("barcode"),
          ),
        ),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildStockCardForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: buildSearchField(
            "barcode",
            barcodeController,
            (context) => searchBarcode("barcode"),
            (context) => onClearSearchField("barcode"),
          ),
        ),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildCsvSaleDetailForm() {
    return Column(
      children: [
        Container(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: buildDateField("from_date", fromDate, selectFromDate),
              ),
              const SizedBox(width: 5),
              Expanded(child: buildDateField("to_date", toDate, selectToDate)),
            ],
          ),
        ),
        SizedBox(height: 10),
        Align(
          alignment: Alignment.centerRight,
          child: TextButton.icon(
            label: Text(global.language("advance_setting")),
            icon: const Icon(Icons.settings_sharp),
            onPressed: () {
              /// show dialog selected listColumns for export csv
              showDialog(
                context: context,
                builder: (context) {
                  return AlertDialog(
                    title: Text(global.language("select_columns")),
                    content: StatefulBuilder(
                      builder: (context, setState) {
                        return Column(
                          children: [
                            CheckboxListTile(
                              title: Text(global.language("select_all")),
                              value:
                                  selectColumsCsv.length ==
                                  columns_csv.listColumns.length,
                              onChanged: (value) {
                                setState(() {
                                  if (value!) {
                                    selectColumsCsv.clear();
                                    selectColumsCsv.addAll(
                                      columns_csv.listColumns,
                                    );
                                  } else {
                                    selectColumsCsv.clear();
                                  }
                                });
                              },
                            ),
                            const Divider(height: 0),
                            Expanded(
                              child: SingleChildScrollView(
                                child: SizedBox(
                                  width: 600,
                                  child: Column(
                                    children: [
                                      for (
                                        int i = 0;
                                        i < columns_csv.listColumns.length;
                                        i++
                                      )
                                        Column(
                                          children: [
                                            CheckboxListTile(
                                              title: Text(
                                                "${i + 1}. ${columns_csv.listColumns[i].name}",
                                              ),
                                              value: selectColumsCsv.contains(
                                                columns_csv.listColumns[i],
                                              ),
                                              onChanged: (value) {
                                                setState(() {
                                                  if (value!) {
                                                    selectColumsCsv.add(
                                                      columns_csv
                                                          .listColumns[i],
                                                    );
                                                  } else {
                                                    selectColumsCsv.remove(
                                                      columns_csv
                                                          .listColumns[i],
                                                    );
                                                  }
                                                });
                                              },
                                            ),
                                            const Divider(height: 0),
                                          ],
                                        ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ],
                        );
                      },
                    ),
                    actions: [
                      TextButton(
                        onPressed: () {
                          selectColumsCsv.clear();
                          Navigator.pop(context);
                        },
                        child: Text(global.language("cancel")),
                      ),

                      /// ElevatedButton save
                      ElevatedButton(
                        onPressed: () {
                          Navigator.pop(context);
                        },
                        child: Text(global.language("save")),
                      ),
                    ],
                  );
                },
              );
            },
          ),
        ),
        const SizedBox(height: 10),
      ],
    );
  }

  Widget buildBranchSelectionWidget({
    required List<CompanyBranchModel> companyBranchListData,
    required String selectedBranch,
    required Function(String) onBranchSelected,
  }) {
    return Column(
      children: [
        Text(global.language("select_branch")),
        Center(
          child: Wrap(
            spacing: 5.0,
            children: companyBranchListData.map((value) {
              return Padding(
                padding: const EdgeInsets.all(5.0),
                child: FilterChip(
                  selectedColor: Colors.blue.shade300,
                  backgroundColor: Colors.grey.shade400,
                  label: Text(value.names.first.name),
                  selected: selectedBranch == value.code,
                  onSelected: (bool selected) {
                    onBranchSelected(value.code);
                  },
                ),
              );
            }).toList(),
          ),
        ),
      ],
    );
  }

  Widget buildDropDownField({
    required String labelfield,
    required String value,
    required List<String> items,
    required Function(String) onChanged,
  }) {
    return DropdownButtonFormField<String>(
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        labelText: global.language(labelfield),
      ),
      initialValue: value,
      items: items.map((item) {
        return DropdownMenuItem<String>(
          value: item,
          child: Text(global.language(item)),
        );
      }).toList(),
      onChanged: (value) {
        onChanged(value!);
      },
    );
  }

  Widget buildSearchField(
    String labelKey,
    TextEditingController controller,
    Function(BuildContext) onTap,
    Function(BuildContext) onClear,
  ) {
    return TextField(
      readOnly: true,
      textAlign: TextAlign.left,
      controller: controller,
      decoration: InputDecoration(
        floatingLabelBehavior: FloatingLabelBehavior.always,
        border: OutlineInputBorder(),
        labelText: global.language(labelKey),
        suffixIcon: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween, // added line
          mainAxisSize: MainAxisSize.min,
          children: [
            (controller.text.isNotEmpty)
                ? IconButton(
                    onPressed: () => onClear(context),
                    // Icon to clear text
                    icon: const Icon(Icons.clear),
                  )
                : const SizedBox(),
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              icon: Icon(
                (labelKey == "from_barcode" ||
                        labelKey == "to_barcode" ||
                        labelKey == "from_group" ||
                        labelKey == "to_group")
                    ? Icons.search
                    : (labelKey == "barcode")
                    ? Icons.search
                    : Icons.person_search,
              ),
              onPressed: () => onTap(context),
            ),
          ],
        ),
      ),
    );
  }

  Widget buildDateField(
    String labelKey,
    TextEditingController controller,
    Function(BuildContext) onSelect,
  ) {
    return TextField(
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        labelText: global.language(labelKey),
        suffixIcon: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.calendar_month),
          onPressed: () => onSelect(context),
        ),
      ),
      controller: controller,
      onChanged: (value) {},
    );
  }

  Widget buildCheckboxListTile(
    String titleKey,
    bool? value,
    Function(bool?) onChanged,
  ) {
    return CheckboxListTile(
      title: Text(global.language(titleKey)),
      value: value,
      onChanged: onChanged,
      controlAffinity: ListTileControlAffinity.leading,
    );
  }

  Widget listObject(int index, ReportListModel value) {
    bool selected = selectCode == value.code;
    TextStyle textStyle = TextStyle(
      fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
      fontSize: (selected)
          ? global.deviceConfig.listDataFontSize + 2.0
          : global.deviceConfig.listDataFontSize,
    );
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    return GestureDetector(
      onTap: () {
        setState(() {
          _timer.cancel();
          selectCode = value.code;
          logDownloadParth.clear();
          getData(selectCode);
          WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
            tabController.animateTo(1);
          });
        });
      },
      child: Container(
        key: index < listKeys.length ? listKeys[index] : null,
        decoration: BoxDecoration(
          color: (selectCode == value.code)
              ? Colors.cyan[100]
              : (index % 2 == 0)
              ? global.theme.columnAlternateEvenColor
              : global.theme.columnAlternateOddColor,
        ),
        padding: EdgeInsets.only(
          left: 10,
          right: 10,
          top: global.deviceConfig.listDataLineSpace,
          bottom: global.deviceConfig.listDataLineSpace,
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              flex: 1,
              child: Text(
                value.code,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 2,
              child: Text(
                global.packName(value.names),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 1,
              child: Text(
                value.group,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
          ],
        ),
      ),
    );
  }

  bool verifyData() {
    List<String> errorList = [];
    if (selectedReport.type == global.ReportEnum.productmovement ||
        selectedReport.type == global.ReportEnum.stockcard) {
      if (barcode.isEmpty) {
        errorList.add(global.language("must") + global.language("barcode"));
      }
    }

    if (errorList.isNotEmpty) {
      global.showSnackBar(
        context,
        Icon(Icons.save, color: Colors.white),
        "${global.language("not_success_process_report")} : ${errorList.join(",")}",
        Colors.red,
      );
      return false;
    } else {
      return true;
    }
  }

  void pdfDownload() {
    if (verifyData()) {
      context.read<ReportBloc>().add(
        DownloadReport(
          xorder: logDownloadParth.length + 1,
          type: selectedReport.type!,
          fromdate: fromDate.text,
          todate: toDate.text,
          showDetail: showDetail ? 1 : 0,
          showSumByDate: showSumByDate ? 1 : 0,
          search: search.text,
          yearnum: year.text,
          monthnum: month.text,
          fromcustcode: fromCustCode,
          tocustcode: toCustCode,
          branch: selectedBranch,
          iscancel: showIscancel ? 1 : 0,
          iscost: showIscost ? 1 : 0,
          fromsalecode: fromsalecode,
          tosalecode: tosalecode,
          inquirytype: selectedInquiryType,
          ispos: selectedIsposType,
          barcode: barcode,
          frombarcode: fromBarcode,
          tobarcode: toBarcode,
          fromgroup: fromGroup,
          togroup: toGroup,
          typefile: typefile,
          listcolumscsv: selectColumsCsv,
        ),
      );
    }
  }

  void checkFileStatus() async {
    await getLogDownloadPath();

    List<LogDownloadParthModel> logDownloadParthTeamp = [];

    if (logDownloadParth.isNotEmpty) {
      for (int i = 0; i < logDownloadParth.length; i++) {
        if (logDownloadParth[i].path!.isNotEmpty &&
            logDownloadParth[i].status == "processing") {
          await reportRepository
              .checkFileStatus(
                selectedReport.type!,
                logDownloadParth[i].path!,
                logDownloadParth[i].jobid!,
              )
              .then((value) {
                if (value.success) {
                  logDownloadParth[i].status = "success";
                  updateLogDownloadParth(logDownloadParth[i]);
                  setState(() {});
                } else {
                  if (value.message == "regenerated") {
                    logDownloadParthTeamp.add(logDownloadParth[i]);
                  }
                }
              })
              .onError((error, stackTrace) {
                if (kDebugMode) {
                  AppLogger.error(error.toString());
                }
              });
        }
      }

      if (logDownloadParthTeamp.isNotEmpty) {
        for (int i = 0; i < logDownloadParthTeamp.length; i++) {
          await pdfDownloadCache(logDownloadParthTeamp[i]);
          await removeLogDownloadParth(logDownloadParthTeamp[i].guidfixed!);
        }
      }
      if (logDownloadParth.every((element) => element.status != "processing")) {
        _timer.cancel();
      }
    } else {
      _timer.cancel();
    }
  }

  Future<void> pdfDownloadCache(LogDownloadParthModel? logTeampFalse) async {
    if (logTeampFalse != null) {
      if (mounted) {
        if (logTeampFalse.filter!.listcolumscsv!.isEmpty) {
          logTeampFalse.filter!.listcolumscsv = selectColumsCsv;
        }
        context.read<ReportBloc>().add(
          DownloadReport(
            xorder: logDownloadParth.length + 1,
            type: selectedReport.type!,
            fromdate: logTeampFalse.filter!.fromdate!,
            todate: logTeampFalse.filter!.todate!,
            showDetail: logTeampFalse.filter!.showdetail,
            showSumByDate: logTeampFalse.filter!.showsumbydate,
            search: logTeampFalse.filter!.search,
            yearnum: logTeampFalse.filter!.yearnum,
            monthnum: logTeampFalse.filter!.monthnum,
            fromcustcode: logTeampFalse.filter!.fromcustcode,
            tocustcode: logTeampFalse.filter!.tocustcode,
            branch: logTeampFalse.filter!.branch,
            iscancel: logTeampFalse.filter!.iscancel,
            fromsalecode: logTeampFalse.filter!.fromsalecode,
            tosalecode: logTeampFalse.filter!.tosalecode,
            inquirytype: logTeampFalse.filter!.inquirytype,
            ispos: logTeampFalse.filter!.ispos,
            frombarcode: logTeampFalse.filter!.frombarcode,
            tobarcode: logTeampFalse.filter!.tobarcode,
            fromgroup: logTeampFalse.filter!.fromgroup,
            togroup: logTeampFalse.filter!.togroup,
            barcode: logTeampFalse.filter!.barcode,
            typefile: logTeampFalse.filter!.typefile,
            listcolumscsv: logTeampFalse.filter!.listcolumscsv,
          ),
        );
      }
    }
  }

  String generateSubtitle(int index) {
    String fromDate = "";
    String toDate = "";
    String custcodeText = "";
    String branchText = "";
    String inquirytypeText = "";
    String saleTypeText = "";
    String barcodeText = "";
    String groupText = "";

    if (logDownloadParth[index].filter!.fromdate!.isNotEmpty) {
      fromDate = logDownloadParth[index].filter!.fromdate!;
    }

    if (logDownloadParth[index].filter!.todate!.isNotEmpty) {
      toDate = logDownloadParth[index].filter!.todate!;
    }

    if (logDownloadParth[index].filter!.fromcustcode!.isNotEmpty &&
        logDownloadParth[index].filter!.tocustcode!.isNotEmpty) {
      custcodeText =
          " , ${global.language("customer")} : ${logDownloadParth[index].filter!.fromcustcode} ${global.language("to")} ${logDownloadParth[index].filter!.tocustcode}";
    } else if (logDownloadParth[index].filter!.fromcustcode!.isNotEmpty) {
      custcodeText =
          " , ${global.language("customer")} : ${logDownloadParth[index].filter!.fromcustcode}";
    } else if (logDownloadParth[index].filter!.tocustcode!.isNotEmpty) {
      custcodeText =
          " , ${global.language("customer")} : ${logDownloadParth[index].filter!.tocustcode}";
    }

    if (logDownloadParth[index].filter!.branch!.isNotEmpty) {
      branchText =
          " , ${global.language("branch")} : ${logDownloadParth[index].filter!.branch}";
    }

    if (logDownloadParth[index].filter!.inquirytype!.isNotEmpty) {
      if (logDownloadParth[index].filter!.inquirytype! == "inquiry_all") {
        inquirytypeText =
            " , ${global.language("inquiry_type")} : ${global.language("inquiry_all")}";
      } else if (logDownloadParth[index].filter!.inquirytype! ==
          "inquiry_credit") {
        inquirytypeText =
            " , ${global.language("inquiry_type")} : ${global.language("inquiry_credit")}";
      } else if (logDownloadParth[index].filter!.inquirytype! ==
          "inquiry_cash") {
        inquirytypeText =
            " , ${global.language("inquiry_type")} : ${global.language("inquiry_cash")}";
      }
    }

    if (logDownloadParth[index].filter!.ispos!.isNotEmpty) {
      if (logDownloadParth[index].filter!.ispos! == "sale_all") {
        saleTypeText =
            " , ${global.language("sale_type")} : ${global.language("sale_all")}";
      } else if (logDownloadParth[index].filter!.ispos! == "sale_pos") {
        saleTypeText =
            " , ${global.language("sale_type")} : ${global.language("sale_pos")}";
      } else if (logDownloadParth[index].filter!.ispos! == "sale_merechant") {
        saleTypeText =
            " , ${global.language("sale_type")} : ${global.language("sale_merechant")}";
      }
    }

    if (logDownloadParth[index].filter!.barcode!.isNotEmpty) {
      barcodeText =
          " , ${global.language("barcode")} : ${logDownloadParth[index].filter!.barcode}";
    }

    if (logDownloadParth[index].filter!.frombarcode!.isNotEmpty &&
        logDownloadParth[index].filter!.tobarcode!.isNotEmpty) {
      barcodeText =
          " , ${global.language("barcode")} : ${logDownloadParth[index].filter!.frombarcode} ${global.language("to")} ${logDownloadParth[index].filter!.tobarcode}";
    } else if (logDownloadParth[index].filter!.frombarcode!.isNotEmpty) {
      barcodeText =
          " , ${global.language("barcode")} : ${logDownloadParth[index].filter!.frombarcode}";
    } else if (logDownloadParth[index].filter!.tobarcode!.isNotEmpty) {
      barcodeText =
          " , ${global.language("barcode")} : ${logDownloadParth[index].filter!.tobarcode}";
    }

    if (logDownloadParth[index].filter!.fromgroup!.isNotEmpty &&
        logDownloadParth[index].filter!.togroup!.isNotEmpty) {
      groupText =
          " , ${global.language("group")} : ${logDownloadParth[index].filter!.fromgroup} ${global.language("to")} ${logDownloadParth[index].filter!.togroup}";
    }

    if (selectedReport.type == global.ReportEnum.saleinvoice ||
        selectedReport.type == global.ReportEnum.saleinvoicedetail) {
      return "${global.language("from_date")} : $fromDate ${global.language("to_date")} $toDate$custcodeText$branchText$inquirytypeText$saleTypeText";
    } else if (selectedReport.type == global.ReportEnum.purchase) {
      return "${global.language("from_date")} : $fromDate ${global.language("to_date")} $toDate$custcodeText$branchText$inquirytypeText";
    } else if (selectedReport.type == global.ReportEnum.productmovement ||
        selectedReport.type == global.ReportEnum.stockcard) {
      return "${global.language("from_date")} : $fromDate ${global.language("to_date")} $toDate$barcodeText";
    } else if (selectedReport.type == global.ReportEnum.salebyproduct) {
      return "${global.language("from_date")} : $fromDate ${global.language("to_date")} $toDate$barcodeText$groupText";
    } else if (selectedReport.type == global.ReportEnum.stockbalance) {
      return "${global.language("to_date")} $toDate$barcodeText";
    } else {
      return "${global.language("from_date")} : $fromDate ${global.language("to_date")} $toDate";
    }
  }

  Widget editScreen({mobileScreen}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        leading: mobileScreen
            ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: const Icon(Icons.arrow_back),
                onPressed: () async {
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                },
              )
            : null,
        title: Text(global.packName(selectedReport.names)),
        actions: const [],
      ),
      body: MultiBlocListener(
        listeners: [
          BlocListener<ReportBloc, ReportState>(
            listener: (context, state) async {
              if (state is DownloadReportSuccess) {
                _timer.cancel();
                await getLogDownloadPath();
                startTimer();
              } else if (state is DownloadReportFailed) {
                global.showSnackBar(
                  context,
                  Icon(Icons.save, color: Colors.white),
                  "${global.language("not_success_process_report")} : ${state.message}",
                  Colors.red,
                );
              }

              if (state is FileStatusGetSuccess) {
                setState(() {
                  /// order by desc logDownloadParth
                  logDownloadParth = state.logDownloadParthModels.reversed
                      .toList();
                });
              } else if (state is FileStatusGetFailed) {
                if (mounted) {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    "${global.language("not_success_process_report")} : ${state.message}",
                    Colors.red,
                  );
                }
              }

              if (state is FileStatusUpdateSuccess) {
                getLogDownloadPath();
              } else if (state is FileStatusUpdateFailed) {
                if (mounted) {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    "${global.language("not_success_process_report")} : ${state.message}",
                    Colors.red,
                  );
                }
              }

              if (state is FileStatusDeleteByIdDeleteSuccess) {
                getLogDownloadPath();
              } else if (state is FileStatusDeleteByIdDeleteFailed) {
                if (mounted) {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    "${global.language("not_success_process_report")} : ${state.message}",
                    Colors.red,
                  );
                }
              }

              if (state is FileStatusDeleteByMenuDeleteSuccess) {
                getLogDownloadPath();
              } else if (state is FileStatusDeleteByMenuDeleteFailed) {
                if (mounted) {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    "${global.language("not_success_process_report")} : ${state.message}",
                    Colors.red,
                  );
                }
              }
            },
          ),
          BlocListener<CompanyBranchBloc, CompanyBranchState>(
            listener: (context, state) {
              if (state is CompanyBranchLoadSuccess) {
                setState(() {
                  companyBranchListData = state.companyBranch;
                });
              }
            },
          ),
        ],
        child: Padding(
          padding: EdgeInsets.all(8.0),
          child: Column(
            children: [
              // _downloadView(),
              Padding(
                padding: const EdgeInsets.only(top: 10.0),
                child: buildWigetReport(selectedReport),
              ),
              (selectedReport.code.isNotEmpty)
                  ? SizedBox(
                      height: 50,
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        icon: const Icon(Icons.play_circle_outline),
                        onPressed: () {
                          pdfDownload();
                        },
                        label: Text(global.language("process")),
                      ),
                    )
                  : Container(),
              const SizedBox(height: 10),
              (logDownloadParth.isNotEmpty)
                  ? Padding(
                      padding: EdgeInsets.only(right: 10.0),
                      child: Align(
                        alignment: Alignment.centerRight,
                        child: TextButton(
                          style: TextButton.styleFrom(
                            foregroundColor: Colors.red,
                          ),
                          onPressed: () {
                            setState(() {
                              removeAllLogDownloadParth();
                            });
                          },
                          child: Text(global.language("clear_all")),
                        ),
                      ),
                    )
                  : const SizedBox(),

              Expanded(
                /// list view logDownloadParth
                child: ListView.builder(
                  itemCount: logDownloadParth.length,
                  itemBuilder: (context, index) {
                    return ListTile(
                      leading: CircleAvatar(child: Text('${index + 1}')),
                      title: Row(
                        children: [
                          Text(
                            "${global.language("status")} : ${(logDownloadParth[index].status! == "success" ? global.language("success") : global.language("processing"))}",
                            style: TextStyle(
                              color:
                                  logDownloadParth[index].status! == "success"
                                  ? Colors.green
                                  : Colors.blue,
                            ),
                          ),
                          const SizedBox(width: 5),
                          (selectedReport.type ==
                                      global.ReportEnum.saleinvoice ||
                                  selectedReport.type ==
                                      global.ReportEnum.saleinvoicedetail)
                              ? Center(
                                  child: SvgPicture.asset(
                                    (logDownloadParth[index].filter!.typefile ==
                                            "pdf")
                                        ? 'assets/icons/file-type-pdf.svg'
                                        : 'assets/icons/file-type-excel.svg',
                                    width: 20,
                                    height: 20,
                                  ),
                                )
                              : const SizedBox(),
                        ],
                      ),
                      subtitle: Text(generateSubtitle(index)),
                      trailing: Wrap(
                        spacing: 12, // space between two icons
                        children: <Widget>[
                          IconButton(
                            disabledColor: Colors.grey,
                            icon: const Icon(Icons.download),
                            color: Colors.blue,
                            onPressed:
                                (logDownloadParth[index].status == "processing")
                                ? null
                                : () {
                                    downloadFile(
                                      logDownloadParth[index].path!,
                                      logDownloadParth[index].path!
                                          .split('/')
                                          .last,
                                    );
                                    Future.delayed(
                                      const Duration(seconds: 1),
                                      () {
                                        setState(() {
                                          _timer.cancel();
                                        });
                                      },
                                    );
                                  },
                          ),
                          IconButton(
                            disabledColor: Colors.grey,
                            icon: const Icon(Icons.delete),
                            color: Colors.red,
                            onPressed:
                                (logDownloadParth[index].status == "processing")
                                ? null
                                : () {
                                    setState(() {
                                      removeLogDownloadParth(
                                        logDownloadParth[index].guidfixed!,
                                      );
                                    });
                                  },
                          ),
                          (logDownloadParth[index].status == "processing")
                              ? const SizedBox(
                                  height: 40.0,
                                  width: 40.0,
                                  child: Center(
                                    child: CircularProgressIndicator(),
                                  ),
                                )
                              : const SizedBox(),
                        ],
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();

    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return (constraints.maxWidth > 800)
              ? SplitView(
                  controller: splitViewController,
                  gripSize: 14,
                  gripColor: global.theme.appBarColor,
                  gripColorActive: Colors.blue,
                  viewMode: SplitViewMode.Horizontal,
                  indicator: const SplitIndicator(
                    viewMode: SplitViewMode.Horizontal,
                  ),
                  activeIndicator: const SplitIndicator(
                    viewMode: SplitViewMode.Horizontal,
                    isActive: true,
                  ),
                  children: [
                    listScreen(mobileScreen: false),
                    editScreen(mobileScreen: false),
                  ],
                )
              : TabBarView(
                  physics: const NeverScrollableScrollPhysics(),
                  controller: tabController,
                  children: [
                    listScreen(mobileScreen: true),
                    editScreen(mobileScreen: true),
                  ],
                );
        },
      ),
    );
  }
}
