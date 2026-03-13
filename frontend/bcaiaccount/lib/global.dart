import 'dart:io';
import 'dart:math';
import 'package:thai_buddhist_date/thai_buddhist_date.dart' as tbd;
import 'package:decimal/decimal.dart';
import 'package:smlaicloud/api/app_const.dart';
import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/model/bank_model.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/profile_model.dart';
import 'package:smlaicloud/model/shop_model.dart';
import 'dart:async';
import 'dart:convert';
import 'package:smlaicloud/theme/app_theme.dart';
export 'package:smlaicloud/theme/app_theme.dart';
import 'package:smlaicloud/model/timezones_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/user_login_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/company_branch_repository.dart';
import 'package:smlaicloud/repositories/shop_repository.dart';
import 'package:device_info_plus/device_info_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_rounded_date_picker/flutter_rounded_date_picker.dart';
import 'package:smlaicloud/model/config_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/public_color_model.dart';
import 'package:smlaicloud/model/public_name_model.dart';
import 'package:shared_preferences/shared_preferences.dart';
export 'components/numpad.dart';
import 'package:translator/translator.dart';
import 'dart:math' as math;
import 'package:intl/intl.dart' as intl;
import 'package:http/http.dart' as http;
import 'package:archive/archive.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
export 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/services/notification_queue_service.dart';
import 'package:smlaicloud/services/branch_config_watcher_service.dart';
import 'package:smlaicloud/services/po_approval_polling_service.dart';

// Global Notification Queue Service instance
final notificationQueue = NotificationQueueService();

// Global PO Approval Polling Service instance
final poApprovalPolling = POApprovalPollingService();

typedef GuidCallback = void Function({String guid, bool isEdit});

class FieldFocusModel {
  FocusNode focusNode;
  bool isReadOnly;

  FieldFocusModel({required this.focusNode, this.isReadOnly = false});
}

DeviceConfigModel deviceConfig = DeviceConfigModel(listDataFontSize: 12, listDataLineSpace: 0, itemDisplayPrice: true, itemDisplaySku: true);

bool allowManualCostTestHostFlag = false;

class SearchCodeNameModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names;
  bool isCancel;
  SearchCodeNameModel({required this.guidfixed, required this.code, required this.names, this.isCancel = false});
}

class SearchDebtorModel {
  String guid;
  String code;
  List<LanguageDataModel> names;
  bool ismember;
  String pricelevel;
  String pointscode;
  SearchDebtorModel({required this.guid, required this.code, required this.names, this.ismember = false, required this.pricelevel, this.pointscode = ''});
}

enum DateTimeFormatEnum { fullDate, date, dateTime, dateTimeDay, time, dateDay }

enum PrintColumnAlign { left, right, center }

enum TransactionTypeEnum {
  purchase,
  purchasereturn,
  sale,
  salereturn,
  stocktransfer,
  stockreceiveproduct,
  stockpickupproduct,
  stockreturnproduct,
  adjust,
  paid,
  pay,
  stockbalance,
  saleorder,
  quotation,
  purchaseorder,
  //ทยอยรับ
  purchasepartial,
  //ตั้งหนี้จากการทยอยรับ
  accrualreceive,
  // จ่ายเงินล่วงหน้า
  advancePayment,
  // รับคืนเงินล่วงหน้า
  advancePaymentRefund,
  // จ่ายเงินมัดจำ
  deposit,
  // รับคืนเงินมัดจำ
  depositRefund,
  // รับเงินล่วงหน้า
  paidAdvance,
  // คืนเงินล่วงหน้า
  paidAdvanceRefund,
  // รับเงินมัดจำ
  receiveDeposit,
  // คืนเงินมัดจำ
  receiveDepositRefund,
}

enum SelectGroupNumberEnum { category, zone, table, tableOrder, kitchen, genQrcode }

enum ReportEnum {
  product,
  saleinvoice,
  saleinvoicedetail,
  debtor,
  creditor,
  bookbank,
  purchase,
  purchasereturn,
  saleinvoicereturn,
  transfer,
  receive,
  pickup,
  returnproduct,
  stockadjustment,

  ///รับชำระ
  paid,

  /// จ่ายชำระ
  pay,

  /// รับเงิน
  getpaid,

  /// จ่ายเงิน
  getpay,
  vatsale,
  vatpurchase,
  salebydebtor,

  /// new
  salebydate,
  receivemoney,
  salebyproduct,
  productmovement,

  /// สินค้าคงเหลือ
  stockbalance,

  ///บัญชีคุมพิเศษ
  stockcard,

  /// CSV
  csvsaledetail,
}

enum PosVersionEnum { pos, restaurant }

enum LoginEnum { none, google, facebook, apple, password }

enum ScreenEventEnum { add, edit, list, display }

enum DeviceModeEnum { none, iphone, ipad, windowsDesktop, macosDesktop, linuxDesktop, androidPhone, androidTablet }

// โหมด พัฒนาโปรแกรม
bool developerMode = true;

/// sent isdev mode to  clickhouse
/// 0 = dev , 1 = prod , 2 = uat
int isdevPin = 0;

UserLoginModel userLoginData = UserLoginModel(token: "", name: "", email: "", photourl: "", code: "", refreshtoken: "");
late SharedPreferences prefs;
List<String> googleLanguageCode = [];
late CompanyModel companyData;
String userLanguage = "";
bool apiConnected = false;
String apiToken = "";
TransactionPayModel payScreenData = TransactionPayModel();
String apiUserName = "maxkorn"; //maxkorn
String apiUserPassword = "maxkorn"; //maxkorn
String apiShopCode = "2Eh6e3pfWvXTp0yV3CyFEhKPjdI"; // "27dcEdktOoaSBYFmnN6G6ett4Jb";
bool isLoginProcess = false;
late SharedPreferences appConfig;
String systemLanguage = "th";
String loginName = "";
String loginEmail = "";
LoginEnum currentLoginMethod = LoginEnum.none;
String loginPhotoUrl = "";
int activeIndexMenu = 0;
DeviceModeEnum deviceMode = DeviceModeEnum.androidPhone;
List<PublicColorModel> publicColors = [];
late List<LanguageSystemModel> languageSystemData;
Map<String, String> _languageMap = {}; // HashMap สำหรับ O(1) lookup
int loadDataPerPage = 100;
late PosVersionEnum posVersion;
ConfigModel config = ConfigModel();
String goApiVersion = "";
ShopModel shopSelectData = ShopModel();
CompanyBranchModel companyBranchSelectData = CompanyBranchModel(guidfixed: '', code: '', businesstype: null);
final translator = GoogleTranslator();
AppConfigClass myAppConfig = AppConfigClass();

List<int> groupNumber = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20];

List<BankModel> bankTempListDatas = [];

/// โหลดโหมดธีมจาก SharedPreferences
void loadThemeSettings() {
  displayThemeMode = appConfig.getString('display_theme_mode') ?? 'light';
}

/// บันทึกโหมดธีมลง SharedPreferences
Future<void> saveThemeSettings() async {
  await appConfig.setString('display_theme_mode', displayThemeMode);
  applyThemeMode();
  displaySettingsNotifier.value++;
}

/// โหลดการตั้งค่าการแสดงผลจาก SharedPreferences
Future<void> loadDisplaySettings() async {
  displayFontFamily = appConfig.getString('display_font_family') ?? 'Sarabun';
  displayZoomLevel = appConfig.getDouble('display_zoom_level') ?? 100.0;

  // ตรวจสอบค่า zoom ให้อยู่ในช่วงที่กำหนด
  if (displayZoomLevel < 50.0) displayZoomLevel = 50.0;
  if (displayZoomLevel > 300.0) displayZoomLevel = 300.0;

  // โหลดโหมดธีม
  loadThemeSettings();
}

/// บันทึกการตั้งค่าการแสดงผลลง SharedPreferences
Future<void> saveDisplaySettings() async {
  await appConfig.setString('display_font_family', displayFontFamily);
  await appConfig.setDouble('display_zoom_level', displayZoomLevel);

  // แจ้งเตือนทุกจอให้ rebuild
  displaySettingsNotifier.value++;
}

/// รายการโปรโมชั่น
late List<PromotionListModel> promotionList;

SlipListModel summaryOrderBill = SlipListModel(
  code: "SALE01",
  names: [
    LanguageDataModel(code: "th", name: "ใบสรุปรายการก่อนบันทึก"),
    LanguageDataModel(code: "en", name: "Pre-Posting Summary"),
  ],
);

SlipListModel taxInvoice = SlipListModel(
  code: "SALE02",
  names: [
    LanguageDataModel(code: "th", name: "ใบกำกับภาษีอย่างย่อ"),
    LanguageDataModel(code: "en", name: "Pre-Posting Summary"),
  ],
);

SlipListModel taxInvoiceFull = SlipListModel(
  code: "SALE03",
  names: [
    LanguageDataModel(code: "th", name: "ใบกำกับภาษีอย่างเต็ม"),
    LanguageDataModel(code: "en", name: "Tax Invoice Full"),
  ],
);

SlipListModel slipReceipt = SlipListModel(
  code: "SALE04",
  names: [
    LanguageDataModel(code: "th", name: "ใบเสร็จรับเงิน"),
    LanguageDataModel(code: "en", name: "Receipt Slip"),
  ],
);

SlipListModel slipReturn = SlipListModel(
  code: "CN01",
  names: [
    LanguageDataModel(code: "th", name: "ใบรับคืนอย่างย่อ"),
    LanguageDataModel(code: "en", name: "Return Slip"),
  ],
);

SlipListModel slipReturnFull = SlipListModel(
  code: "CN02",
  names: [
    LanguageDataModel(code: "th", name: "ใบรับคืนอย่างเต็ม"),
    LanguageDataModel(code: "en", name: "Return Slip Full"),
  ],
);

/// slip pos ขาย
List<SlipListModel> slipPosSaleList = [summaryOrderBill, taxInvoice, taxInvoiceFull, slipReceipt, slipReturn, slipReturnFull];

/// timezone list
List<TimezonesModel> timezonesListData = [];

ProfileModel profileData = ProfileModel();

Locale local = const Locale('en', 'US');
EraMode eraMode = EraMode.CHRIST_YEAR;

final moneyFormatAndDot = intl.NumberFormat("##,##0.00");
final moneyFormat = intl.NumberFormat("##,##0.##");
// วิธีการปัดเศษเงินยอดรวม 0=ไม่ปัดเศษ,1=ปัดเศษตามกฏหมาย,2=ปัดเศษขึ้นเป็นจำนวนเต็ม,3=ปัดเศษลงเป็นจำนวนเต็ม
int payTotalMoneyRoundType = 1;
// Step การปัดเศษ ค่าว่าง=จำนวนเต็มอัตโนมัติ,0.25,0.5,0.75
List<MoneyRoundPayModel> payTotalMoneyRoundStep = [
  MoneyRoundPayModel(begin: 0.01, end: 0.12, value: 0),
  MoneyRoundPayModel(begin: 0.13, end: 0.37, value: 0.25),
  MoneyRoundPayModel(begin: 0.38, end: 0.62, value: 0.5),
  MoneyRoundPayModel(begin: 0.63, end: 0.87, value: 0.75),
  MoneyRoundPayModel(begin: 0.88, end: 0.99, value: 1.0),
];

double roundDouble(double value, int places) {
  num mod = pow(10.0, places);
  return ((value * mod).round().toDouble() / mod);
}

int getPageNumber(int itemIndex, int limit) {
  return (itemIndex ~/ limit) + 1;
}

double roundDoubleDown(double value, int precision) {
  final divideBy = pow(10, precision);
  return ((value * divideBy).floorToDouble() / divideBy);
}

double roundMoneyForPay(double value) {
  double result = value;
  if (payTotalMoneyRoundStep.isEmpty) {
    // ถ้าเป็นค่าว่าง ให้ปัดจำนวนเต็มอัตโนมัติ
    result = roundDouble(value, 0);
  } else {
    value = roundDouble(value, 2);
    double calcRound = roundDouble(value - value.floorToDouble(), 2);
    for (int index = 0; index < payTotalMoneyRoundStep.length; index++) {
      if (calcRound >= payTotalMoneyRoundStep[index].begin && calcRound <= payTotalMoneyRoundStep[index].end) {
        result = roundDoubleDown(value, 0) + payTotalMoneyRoundStep[index].value;
        break;
      }
    }
  }
  return result;
}

bool isMobileScreen(BuildContext context) {
  return MediaQuery.of(context).size.width < 600;
}

String transactionName(TransactionTypeEnum transactionType) {
  switch (transactionType) {
    case TransactionTypeEnum.purchase:
      return language("transaction_purchase");
    case TransactionTypeEnum.purchasereturn:
      return language("transaction_purchasereturn");
    case TransactionTypeEnum.sale:
      return language("transaction_sale");
    case TransactionTypeEnum.salereturn:
      return language("transaction_salereturn");
    case TransactionTypeEnum.stockpickupproduct:
      return language("transaction_stock_pick_up_product");
    case TransactionTypeEnum.stockreceiveproduct:
      return language("transaction_stock_receive_product");
    case TransactionTypeEnum.stockreturnproduct:
      return language("transaction_stock_return_product");
    case TransactionTypeEnum.stocktransfer:
      return language("transaction_stock_transfer");
    case TransactionTypeEnum.adjust:
      return language("transaction_adjust");
    case TransactionTypeEnum.paid:
      return language("transaction_paid");
    case TransactionTypeEnum.pay:
      return language("transaction_pay");
    case TransactionTypeEnum.stockbalance:
      return language("transaction_stock_balance");
    case TransactionTypeEnum.saleorder:
      return language("transaction_sale_order");
    case TransactionTypeEnum.quotation:
      return language("transaction_quotation");
    case TransactionTypeEnum.purchaseorder:
      return language("transaction_purchase_order");
    case TransactionTypeEnum.purchasepartial:
      return language("transaction_purchase_partial");
    case TransactionTypeEnum.accrualreceive:
      return language("transaction_accrual_receive");
    case TransactionTypeEnum.advancePayment:
      return language("transaction_advance_payment");
    case TransactionTypeEnum.advancePaymentRefund:
      return language("transaction_advance_payment_refund");
    case TransactionTypeEnum.deposit:
      return language("transaction_deposit");
    case TransactionTypeEnum.depositRefund:
      return language("transaction_deposit_refund");
    case TransactionTypeEnum.paidAdvance:
      return language("transaction_paid_advance");
    case TransactionTypeEnum.paidAdvanceRefund:
      return language("transaction_paid_advance_refund");
    case TransactionTypeEnum.receiveDeposit:
      return language("transaction_receive_deposit");
    case TransactionTypeEnum.receiveDepositRefund:
      return language("transaction_receive_deposit_refund");
  }
}

String getreportName(ReportEnum reportType) {
  switch (reportType) {
    case ReportEnum.product:
      return language("report_product");
    case ReportEnum.saleinvoice:
      return language("report_saleinvoice");
    case ReportEnum.saleinvoicedetail:
      return language("report_saleinvoicedetail");
    case ReportEnum.debtor:
      return language("report_debtor");
    case ReportEnum.creditor:
      return language("report_creditor");
    case ReportEnum.bookbank:
      return language("report_bookbank");
    case ReportEnum.purchase:
      return language("report_purchase");
    case ReportEnum.purchasereturn:
      return language("report_purchasereturn");
    case ReportEnum.saleinvoicereturn:
      return language("report_saleinvoicereturn");
    case ReportEnum.transfer:
      return language("report_transfer");
    case ReportEnum.receive:
      return language("report_receive");
    case ReportEnum.pickup:
      return language("report_pickup");
    case ReportEnum.returnproduct:
      return language("report_returnproduct");
    case ReportEnum.stockadjustment:
      return language("report_stockadjustment");
    case ReportEnum.paid:
      return language("report_paid");
    case ReportEnum.pay:
      return language("report_pay");
    case ReportEnum.getpaid:
      return language("report_getpaid");
    case ReportEnum.getpay:
      return language("report_getpay");
    case ReportEnum.vatsale:
      return language("report_vatsale");
    case ReportEnum.vatpurchase:
      return language("report_vatpurchase");
    case ReportEnum.salebydebtor:
      return language("report_salebydebtor");
    case ReportEnum.salebydate:
      return language("report_salebydate");
    case ReportEnum.receivemoney:
      return language("report_receivemoney");
    case ReportEnum.salebyproduct:
      return language("report_salebyproduct");
    case ReportEnum.productmovement:
      return language("report_productmovement");
    case ReportEnum.stockbalance:
      return language("report_stockbalance");
    case ReportEnum.stockcard:
      return language("report_stockcard");
    case ReportEnum.csvsaledetail:
      return language("csv_saledetail");
  }
}

String dateTimeBuddhist(DateTime dateTime, {DateTimeFormatEnum format = DateTimeFormatEnum.fullDate}) {
  int day = dateTime.day;
  int month = dateTime.month;
  String dayOfWeek = intl.DateFormat.EEEE('th_TH').format(dateTime);
  String monthYear = intl.DateFormat.MMMM('th_TH').format(dateTime);
  String yearStr = tbd.format(dateTime, pattern: 'yyyy', era: tbd.Era.be);
  switch (format) {
    case DateTimeFormatEnum.fullDate:
      return '$dayOfWeek ที่ $day $monthYear $yearStr';
    case DateTimeFormatEnum.date:
      return "${day.toString().padLeft(2, '0')}/${month.toString().padLeft(2, '0')}/$yearStr";
    case DateTimeFormatEnum.time:
      return "${dateTime.hour.toString().padLeft(2, '0')}:${dateTime.minute.toString().padLeft(2, '0')}";
    case DateTimeFormatEnum.dateTime:
      if (dateTime.hour == 0 && dateTime.minute == 0) {
        return "${day.toString().padLeft(2, '0')}/${month.toString().padLeft(2, '0')}/$yearStr";
      } else {
        return "${day.toString().padLeft(2, '0')}/${month.toString().padLeft(2, '0')}/$yearStr ${dateTime.hour.toString().padLeft(2, '0')}:${dateTime.minute.toString().padLeft(2, '0')}";
      }
    case DateTimeFormatEnum.dateTimeDay:
      if (dateTime.hour == 0 && dateTime.minute == 0) {
        return "${day.toString().padLeft(2, '0')}/${month.toString().padLeft(2, '0')}/$yearStr ($dayOfWeek)";
      } else {
        return "${day.toString().padLeft(2, '0')}/${month.toString().padLeft(2, '0')}/$yearStr ${dateTime.hour.toString().padLeft(2, '0')}:${dateTime.minute.toString().padLeft(2, '0')} ($dayOfWeek)";
      }
    case DateTimeFormatEnum.dateDay:
      return "${day.toString().padLeft(2, '0')}/${month.toString().padLeft(2, '0')}/$yearStr ($dayOfWeek)";
  }
}

void getTimezones() {
  const githubRawUrl = 'https://raw.githubusercontent.com/smlsoft/dedepos_template/main/timezones.json';
  readFileFromGithub(githubRawUrl)
      .then((fileContent) {
        List<TimezonesModel> timezones = (json.decode(fileContent) as List).map((timezone) => TimezonesModel.fromJson(timezone)).toList();

        for (var data in timezones) {
          timezonesListData.add(data);
        }
      })
      .catchError((error) {
        // ignore: avoid_print
        if (kDebugMode) {
          AppLogger.error('Error reading file: $error');
        }
      });
}

TimeOfDay getTimeOfDayFromString(String timeString) {
  // Split the time string into hours and minutes
  List<String> timeParts = timeString.split(":");
  int hours = int.parse(timeParts[0]);
  int minutes = int.parse(timeParts[1]);

  // Create the TimeOfDay object
  TimeOfDay time = TimeOfDay(hour: hours, minute: minutes);

  return time;
}

// randomDocNo ถูกลบออกแล้ว - mainapi จะสร้าง docno เองโดยใช้ DocDateLocal จาก frontend
// ดู mainapi/internal/transaction/purchaseorder/services/purchaseorder_http_service.go

/// ดึงข้อมูล timezone ของ frontend สำหรับส่งไป mainapi
/// คืนค่า Map ที่มี docdatelocal, doctimelocal, timezone
Map<String, String> getLocalTimezoneInfo([DateTime? dateTime]) {
  final now = dateTime ?? DateTime.now();
  final localTime = now.toLocal();

  // วันที่ local format YYYYMMDD
  final docdatelocal =
      "${localTime.year.toString().padLeft(4, '0')}"
      "${localTime.month.toString().padLeft(2, '0')}"
      "${localTime.day.toString().padLeft(2, '0')}";

  // เวลา local format HH:mm:ss
  final doctimelocal =
      "${localTime.hour.toString().padLeft(2, '0')}:"
      "${localTime.minute.toString().padLeft(2, '0')}:"
      "${localTime.second.toString().padLeft(2, '0')}";

  // timezone name - ใช้ offset เพราะ Dart ไม่มี IANA timezone name โดยตรง
  final offset = localTime.timeZoneOffset;
  final hours = offset.inHours.abs().toString().padLeft(2, '0');
  final minutes = (offset.inMinutes.abs() % 60).toString().padLeft(2, '0');
  final sign = offset.isNegative ? '-' : '+';
  final timezone = 'UTC$sign$hours:$minutes'; // เช่น "UTC+07:00"

  return {'docdatelocal': docdatelocal, 'doctimelocal': doctimelocal, 'timezone': timezone};
}

String packName(List<LanguageDataModel> names) {
  String result = "";
  for (int i = 0; i < names.length; i++) {
    if (names[i].name.isNotEmpty) {
      if (i > 0) {
        if (names[i].name != '') {
          result += ",";
        }
      }
      result += names[i].name;
    }
  }
  return result;
}

class Debouncer {
  final int? milliseconds;
  VoidCallback? action;
  Timer? _timer;

  Debouncer(this.milliseconds);

  void run(VoidCallback action) {
    if (_timer != null) {
      _timer!.cancel();
    }

    _timer = Timer(Duration(milliseconds: milliseconds!), action);
  }

  /// Cancel any pending timer to prevent memory leaks
  void cancel() {
    _timer?.cancel();
    _timer = null;
  }

  /// Dispose the debouncer - alias for cancel for consistency
  void dispose() {
    cancel();
  }
}

Future<void> loadShopProfile() async {}

Future<void> setSystemLanguage(BuildContext context) async {
  List<LanguageModel> defaultLanguages = [
    LanguageModel(code: "th", codeTranslator: "th", name: "Thai", isuse: false),
    LanguageModel(code: "en", codeTranslator: "en", name: "English", isuse: false),
    LanguageModel(code: "cn", codeTranslator: "cn", name: "Chinese", isuse: false),
    LanguageModel(code: "ja", codeTranslator: "ja", name: "Japanese", isuse: false),
    LanguageModel(code: "ko", codeTranslator: "ko", name: "Korean", isuse: false),
    LanguageModel(code: "lo", codeTranslator: "lo", name: "Lao", isuse: false),
    LanguageModel(code: "my", codeTranslator: "my", name: "Burmese", isuse: false),
    LanguageModel(code: "ms", codeTranslator: "ms", name: "Malaysian", isuse: false),
    LanguageModel(code: "vi", codeTranslator: "vi", name: "Vietnamese", isuse: false),
    LanguageModel(code: "km", codeTranslator: "km", name: "Khmer", isuse: false),
  ];
  config.languages = [];
  ShopRepository shopRepository = ShopRepository();
  // JsonRepository jsonRepository = JsonRepository();
  CompanyBranchRepository companyBranchRepository = CompanyBranchRepository();

  try {
    ApiResponse result = await shopRepository.loadShopInfo(getShopId());
    if (result.success) {
      ShopModel shopmodel = ShopModel.fromJson(result.data);
      shopSelectData = shopmodel;

      config.vatrate = shopmodel.settings!.vatrate!;
      config.vattypesale = shopmodel.settings!.vattypesale!;
      config.vattypepurchase = shopmodel.settings!.vattypepurchase!;
      config.inquirytypesale = shopmodel.settings!.inquirytypesale!;
      config.inquirytypepurchase = shopmodel.settings!.inquirytypepurchase!;

      for (var data in shopmodel.settings!.languageconfigs!) {
        if (data.code == "th") {
          config.languages.add(LanguageModel(code: "th", codeTranslator: "th", name: "Thai", isuse: true));
        } else if (data.code == "en") {
          config.languages.add(LanguageModel(code: "en", codeTranslator: "en", name: "English", isuse: true));
        } else if (data.code == "cn") {
          config.languages.add(LanguageModel(code: "cn", codeTranslator: "cn", name: "Chinese", isuse: true));
        } else if (data.code == "ja") {
          config.languages.add(LanguageModel(code: "ja", codeTranslator: "ja", name: "Japanese", isuse: true));
        } else if (data.code == "ko") {
          config.languages.add(LanguageModel(code: "ko", codeTranslator: "ko", name: "Korean", isuse: true));
        } else if (data.code == "lo") {
          config.languages.add(LanguageModel(code: "lo", codeTranslator: "lo", name: "Lao", isuse: true));
        } else if (data.code == "my") {
          config.languages.add(LanguageModel(code: "my", codeTranslator: "my", name: "Burmese", isuse: true));
        } else if (data.code == "ms") {
          config.languages.add(LanguageModel(code: "ms", codeTranslator: "ms", name: "Malaysian", isuse: true));
        } else if (data.code == "vi") {
          config.languages.add(LanguageModel(code: "vi", codeTranslator: "vi", name: "Vietnamese", isuse: true));
        } else if (data.code == "km") {
          config.languages.add(LanguageModel(code: "km", codeTranslator: "km", name: "Khmer", isuse: true));
        }
      }

      try {
        ApiResponse resultBranch = await companyBranchRepository.getBranch(getBranchGuidFixed());

        if (resultBranch.success) {
          CompanyBranchModel companyBranchModel = CompanyBranchModel.fromJson(resultBranch.data);
          companyBranchSelectData = companyBranchModel;

          AppLogger.info('[Branch] ค่าทศนิยมจากสาขา: quantity=${companyBranchModel.decimalQuantity}, price=${companyBranchModel.decimalPrice}, document=${companyBranchModel.decimalDocument}');

          // เริ่มติดตามการเปลี่ยนแปลง config ของสาขา
          branchConfigWatcher.startWatching(
            branchGuidFixed: companyBranchModel.guidfixed,
            intervalSeconds: 60, // ตรวจสอบทุก 60 วินาที
          );
        }
      } catch (ex) {
        if (kDebugMode) {
          AppLogger.debug(ex);
        }
      }
    }

    // ApiResponse result = await jsonRepository.getSetting("ConfigSystem", "");
    // if (result.success) {
    //   if (result.data.length > 0) {
    //     ConfigSystemModel configSystem = ConfigSystemModel.fromJson(json.decode(result.data[0]['body']));
    //     config.vatrate = configSystem.vatrate!;
    //     config.vattypesale = configSystem.vattypesale!;
    //     config.vattypepurchase = configSystem.vattypepurchase!;
    //     config.inquirytypesale = configSystem.inquirytypesale!;
    //     config.inquirytypepurchase = configSystem.inquirytypepurchase!;

    //     for (var data in configSystem.languageList) {
    //       if (data == "th") {
    //         config.languages.add(
    //           LanguageModel(code: "th", codeTranslator: "th", name: "Thai", isuse: true),
    //         );
    //       } else if (data == "en") {
    //         config.languages.add(
    //           LanguageModel(code: "en", codeTranslator: "en", name: "English", isuse: true),
    //         );
    //       } else if (data == "cn") {
    //         config.languages.add(
    //           LanguageModel(code: "cn", codeTranslator: "cn", name: "Chinese", isuse: true),
    //         );
    //       } else if (data == "ja") {
    //         config.languages.add(
    //           LanguageModel(code: "ja", codeTranslator: "ja", name: "Japanese", isuse: true),
    //         );
    //       } else if (data == "ko") {
    //         config.languages.add(
    //           LanguageModel(code: "ko", codeTranslator: "ko", name: "Korean", isuse: true),
    //         );
    //       } else if (data == "lo") {
    //         config.languages.add(
    //           LanguageModel(code: "lo", codeTranslator: "lo", name: "Lao", isuse: true),
    //         );
    //       } else if (data == "my") {
    //         config.languages.add(
    //           LanguageModel(code: "my", codeTranslator: "my", name: "Burmese", isuse: true),
    //         );
    //       } else if (data == "ms") {
    //         config.languages.add(
    //           LanguageModel(code: "ms", codeTranslator: "ms", name: "Malaysian", isuse: true),
    //         );
    //       } else if (data == "vi") {
    //         config.languages.add(
    //           LanguageModel(code: "vi", codeTranslator: "vi", name: "Vietnamese", isuse: true),
    //         );
    //       } else if (data == "km") {
    //         config.languages.add(
    //           LanguageModel(code: "km", codeTranslator: "km", name: "Khmer", isuse: true),
    //         );
    //       }
    //     }

    //     for (var defaultLang in defaultLanguages) {
    //       LanguageModel result =
    //           config.languages.firstWhere((data) => data.code == defaultLang.code, orElse: () => LanguageModel(code: '', codeTranslator: '', name: '', isuse: false));
    //       if (result.code == '') {
    //         config.languages.add(defaultLang);
    //       }
    //     }
    //   } else {
    //     config.languages = defaultLanguages;
    //     config.languages[0].use = true;
    //     config.vatrate = 7.0;
    //     config.vattypesale = 0;
    //     config.vattypepurchase = 0;
    //     config.inquirytypesale = 0;
    //     config.inquirytypepurchase = 0;
    //   }
    // }
  } catch (ex) {
    config.languages = defaultLanguages;
    config.languages[0].isuse = true;
    config.vatrate = 7.0;
    config.vattypesale = 0;
    config.vattypepurchase = 0;
    config.inquirytypesale = 0;
    config.inquirytypepurchase = 0;
  }

  List<PriceModel> allPrices = [
    PriceModel(
      keyNumber: 1,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: language("price_retail"), isuse: true)],
    ),
    PriceModel(
      keyNumber: 2,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: language("price_member"), isuse: true)],
    ),
    PriceModel(
      keyNumber: 3,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: language("product_price_delivery"), isuse: true)],
    ),
    PriceModel(
      keyNumber: 4,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: language("price_channel_1"), isuse: true)],
    ),
    PriceModel(
      keyNumber: 5,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: language("price_channel_2"), isuse: true)],
    ),
    PriceModel(
      keyNumber: 6,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: language("price_channel_3"), isuse: true)],
    ),
    PriceModel(
      keyNumber: 7,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: language("price_channel_4"), isuse: true)],
    ),
    PriceModel(
      keyNumber: 8,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: language("price_channel_5"), isuse: true)],
    ),
    PriceModel(
      keyNumber: 9,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 1", isuse: true)],
    ),
    PriceModel(
      keyNumber: 10,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 2", isuse: true)],
    ),
    PriceModel(
      keyNumber: 11,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 3", isuse: true)],
    ),
    PriceModel(
      keyNumber: 12,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 4", isuse: true)],
    ),
    PriceModel(
      keyNumber: 13,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 5", isuse: true)],
    ),
    PriceModel(
      keyNumber: 14,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 6", isuse: true)],
    ),
    PriceModel(
      keyNumber: 15,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 7", isuse: true)],
    ),
    PriceModel(
      keyNumber: 16,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 8", isuse: true)],
    ),
    PriceModel(
      keyNumber: 17,
      isUse: true,
      names: [LanguageModel(code: "th", codeTranslator: "th", name: "ราคาลู่ที่ 9", isuse: true)],
    ),
  ];

  // กรองราคาตาม posVersion
  if (posVersion == PosVersionEnum.pos) {
    // POS: แสดง keyNumber 1-2 และ 9-17
    config.prices = allPrices.where((price) => (price.keyNumber >= 1 && price.keyNumber <= 2) || (price.keyNumber >= 9 && price.keyNumber <= 17)).toList();
  } else if (posVersion == PosVersionEnum.restaurant) {
    // Restaurant: แสดง keyNumber 1-8
    config.prices = allPrices.where((price) => price.keyNumber >= 1 && price.keyNumber <= 8).toList();
  }

  publicColors = [
    PublicColorModel()
      ..code = "white"
      ..names = [
        PublicNameModel(languageCode: "th", name: "ขาว"),
        PublicNameModel(languageCode: "en", name: "White"),
        PublicNameModel(languageCode: "cn", name: "白色"),
        PublicNameModel(languageCode: "jp", name: "白"),
        PublicNameModel(languageCode: "kr", name: "하얀색"),
        PublicNameModel(languageCode: "lo", name: "ຊາຍ"),
        PublicNameModel(languageCode: "mr", name: "အဖြူ"),
        PublicNameModel(languageCode: "my", name: "Putih"),
        PublicNameModel(languageCode: "vi", name: "Trắng"),
        PublicNameModel(languageCode: "km", name: "ស"),
      ]
      ..color = "#FFFFFF",
    PublicColorModel()
      ..code = "black"
      ..names = [
        PublicNameModel(languageCode: "th", name: "ดำ"),
        PublicNameModel(languageCode: "en", name: "Black"),
        PublicNameModel(languageCode: "cn", name: "黑色"),
        PublicNameModel(languageCode: "jp", name: "黒"),
        PublicNameModel(languageCode: "kr", name: "검은색"),
        PublicNameModel(languageCode: "lo", name: "ດັງ"),
        PublicNameModel(languageCode: "mr", name: "အနောက်"),
        PublicNameModel(languageCode: "my", name: "Hitam"),
        PublicNameModel(languageCode: "vi", name: "Đen"),
        PublicNameModel(languageCode: "km", name: "ខ្មៅ"),
      ]
      ..color = "#000000",
    PublicColorModel()
      ..code = "red"
      ..names = [
        PublicNameModel(languageCode: "th", name: "แดง"),
        PublicNameModel(languageCode: "en", name: "Red"),
        PublicNameModel(languageCode: "cn", name: "红色"),
        PublicNameModel(languageCode: "jp", name: "赤"),
        PublicNameModel(languageCode: "kr", name: "빨간색"),
        PublicNameModel(languageCode: "lo", name: "ເທື່ອ"),
        PublicNameModel(languageCode: "mr", name: "အန္တရာယ်"),
        PublicNameModel(languageCode: "my", name: "Merah"),
        PublicNameModel(languageCode: "vi", name: "Đỏ"),
        PublicNameModel(languageCode: "km", name: "ក្រហម"),
      ]
      ..color = "#FF0000",
    PublicColorModel()
      ..code = "green"
      ..names = [
        PublicNameModel(languageCode: "th", name: "เขียว"),
        PublicNameModel(languageCode: "en", name: "Green"),
        PublicNameModel(languageCode: "cn", name: "绿色"),
        PublicNameModel(languageCode: "jp", name: "緑"),
        PublicNameModel(languageCode: "kr", name: "녹색"),
        PublicNameModel(languageCode: "lo", name: "ສີຂາວ"),
        PublicNameModel(languageCode: "mr", name: "အဖြူ"),
        PublicNameModel(languageCode: "my", name: "Hijau"),
        PublicNameModel(languageCode: "vi", name: "Xanh lá"),
        PublicNameModel(languageCode: "km", name: "បៃតង"),
      ]
      ..color = "#008000",
    PublicColorModel()
      ..code = "blue"
      ..names = [
        PublicNameModel(languageCode: "th", name: "น้ำเงิน"),
        PublicNameModel(languageCode: "en", name: "Blue"),
        PublicNameModel(languageCode: "cn", name: "蓝色"),
        PublicNameModel(languageCode: "jp", name: "青"),
        PublicNameModel(languageCode: "kr", name: "파란색"),
        PublicNameModel(languageCode: "lo", name: "ສີບາດ"),
        PublicNameModel(languageCode: "mr", name: "အဖြူ"),
        PublicNameModel(languageCode: "my", name: "Biru"),
        PublicNameModel(languageCode: "vi", name: "Xanh da trời"),
        PublicNameModel(languageCode: "km", name: "ខៀវ"),
      ]
      ..color = "#0000FF",
    PublicColorModel()
      ..code = "yellow"
      ..names = [
        PublicNameModel(languageCode: "th", name: "เหลือง"),
        PublicNameModel(languageCode: "en", name: "Yellow"),
        PublicNameModel(languageCode: "cn", name: "黄色"),
        PublicNameModel(languageCode: "jp", name: "黄"),
        PublicNameModel(languageCode: "kr", name: "노란색"),
        PublicNameModel(languageCode: "lo", name: "ສີເຫຼືອ"),
        PublicNameModel(languageCode: "mr", name: "အရောင်"),
        PublicNameModel(languageCode: "my", name: "Kuning"),
        PublicNameModel(languageCode: "vi", name: "Vàng"),
        PublicNameModel(languageCode: "km", name: "លឿង"),
      ]
      ..color = "#FFFF00",
    PublicColorModel()
      ..code = "orange"
      ..names = [
        PublicNameModel(languageCode: "th", name: "ส้ม"),
        PublicNameModel(languageCode: "en", name: "Orange"),
        PublicNameModel(languageCode: "cn", name: "橙色"),
        PublicNameModel(languageCode: "jp", name: "オレンジ"),
        PublicNameModel(languageCode: "kr", name: "주황색"),
        PublicNameModel(languageCode: "lo", name: "ສີເຫຼືອ"),
        PublicNameModel(languageCode: "mr", name: "အရောင်"),
        PublicNameModel(languageCode: "my", name: "Kuning"),
        PublicNameModel(languageCode: "vi", name: "Cam"),
        PublicNameModel(languageCode: "km", name: "លឿង"),
      ]
      ..color = "#FFA500",
    PublicColorModel()
      ..code = "purple"
      ..names = [
        PublicNameModel(languageCode: "th", name: "ม่วง"),
        PublicNameModel(languageCode: "en", name: "Purple"),
        PublicNameModel(languageCode: "cn", name: "紫色"),
        PublicNameModel(languageCode: "jp", name: "紫"),
        PublicNameModel(languageCode: "kr", name: "보라색"),
        PublicNameModel(languageCode: "lo", name: "ສີມາດ"),
        PublicNameModel(languageCode: "mr", name: "အမျိုးသား"),
        PublicNameModel(languageCode: "my", name: "Ungu"),
        PublicNameModel(languageCode: "vi", name: "Tím"),
        PublicNameModel(languageCode: "km", name: "ស្វាយ"),
      ]
      ..color = "#800080",
    PublicColorModel()
      ..code = "brown"
      ..names = [
        PublicNameModel(languageCode: "th", name: "น้ำตาล"),
        PublicNameModel(languageCode: "en", name: "Brown"),
        PublicNameModel(languageCode: "cn", name: "棕色"),
        PublicNameModel(languageCode: "jp", name: "茶色"),
        PublicNameModel(languageCode: "kr", name: "갈색"),
        PublicNameModel(languageCode: "lo", name: "ສີບາດ"),
        PublicNameModel(languageCode: "mr", name: "အဖြူ"),
        PublicNameModel(languageCode: "my", name: "Biru"),
        PublicNameModel(languageCode: "vi", name: "Xanh da trời"),
        PublicNameModel(languageCode: "km", name: "ខៀវ"),
      ]
      ..color = "#A52A2A",
    PublicColorModel()
      ..code = "pink"
      ..names = [
        PublicNameModel(languageCode: "th", name: "ชมพู"),
        PublicNameModel(languageCode: "en", name: "Pink"),
        PublicNameModel(languageCode: "cn", name: "粉色"),
        PublicNameModel(languageCode: "jp", name: "ピンク"),
        PublicNameModel(languageCode: "kr", name: "분홍색"),
        PublicNameModel(languageCode: "lo", name: "ສີມາດ"),
        PublicNameModel(languageCode: "mr", name: "အမျိုးသား"),
        PublicNameModel(languageCode: "my", name: "Ungu"),
        PublicNameModel(languageCode: "vi", name: "Tím"),
        PublicNameModel(languageCode: "km", name: "ស្វាយ"),
      ]
      ..color = "#FFC0CB",
    PublicColorModel()
      ..code = "gray"
      ..names = [
        PublicNameModel(languageCode: "th", name: "เทา"),
        PublicNameModel(languageCode: "en", name: "Gray"),
        PublicNameModel(languageCode: "cn", name: "灰色"),
        PublicNameModel(languageCode: "jp", name: "灰色"),
        PublicNameModel(languageCode: "kr", name: "회색"),
        PublicNameModel(languageCode: "lo", name: "ສີເຫຼືອ"),
        PublicNameModel(languageCode: "mr", name: "အရောင်"),
        PublicNameModel(languageCode: "my", name: "Kuning"),
        PublicNameModel(languageCode: "vi", name: "Vàng"),
        PublicNameModel(languageCode: "km", name: "លឿង"),
      ]
      ..color = "#808080",
  ];

  for (var i = 0; i < publicColors.length; i++) {
    for (var j = 0; j < publicColors[i].names.length; j++) {
      publicColors[i].name = "";
      if (publicColors[i].names[j].languageCode == systemLanguage) {
        publicColors[i].name = publicColors[i].names[j].name;
        break;
      }
    }
  }
}

Future<String> readFileFromGithub(String url) async {
  try {
    final response = await http.get(Uri.parse(url));

    if (response.statusCode == 200) {
      final fileContent = response.body;
      return fileContent;
    } else {
      throw Exception('Failed to load file');
    }
  } catch (e) {
    rethrow;
  }
}

void showSnackBar(BuildContext context, Icon icon, String message, Color color) {
  ScaffoldMessenger.of(context).clearSnackBars();
  ScaffoldMessenger.of(context).showSnackBar(
    SnackBar(
      backgroundColor: color,
      duration: const Duration(seconds: 3),
      behavior: SnackBarBehavior.floating,
      margin: const EdgeInsets.only(bottom: 20, left: 20, right: 20),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      content: Row(
        children: [
          icon,
          const SizedBox(width: 10),
          Flexible(child: Text(message, overflow: TextOverflow.ellipsis)),
        ],
      ),
    ),
  );
}

/// แสดง SnackBar สำเร็จ (สีเขียว, 3 วินาที, ปิดอัตโนมัติ)
void showSuccessSnackBar(BuildContext context, String message) {
  showSnackBar(context, const Icon(Icons.check_circle, color: Colors.white), message, Colors.green);
}

/// แสดง SnackBar แจ้งข้อผิดพลาด (สีแดง, 3 วินาที, ปิดอัตโนมัติ)
void showErrorSnackBar(BuildContext context, String message) {
  showSnackBar(context, const Icon(Icons.error, color: Colors.white), message, Colors.red);
}

/// แสดง SnackBar แจ้งเตือน (สีส้ม, 3 วินาที, ปิดอัตโนมัติ)
void showWarningSnackBar(BuildContext context, String message) {
  showSnackBar(context, const Icon(Icons.warning, color: Colors.white), message, Colors.orange);
}

/// แสดง SnackBar ข้อมูลทั่วไป (สีน้ำเงิน, 3 วินาที, ปิดอัตโนมัติ)
void showInfoSnackBar(BuildContext context, String message) {
  showSnackBar(context, const Icon(Icons.info, color: Colors.white), message, Colors.blue);
}

/// แปลงข้อความจาก language code เป็นข้อความภาษาที่เลือก
/// ใช้ HashMap O(1) lookup แทน Linear Search O(n) เพื่อความเร็ว
String language(String code) {
  code = code.trim().toLowerCase();

  // ค้นหาจาก HashMap O(1) - เร็วกว่า loop มาก
  final result = _languageMap[code];

  // ถ้าไม่เจอ ให้คืนค่า code เดิม
  if (result == null || result.trim().isEmpty) {
    return code;
  }

  return result;
}

/// โหลดข้อมูลภาษา: asset baseline ก่อน แล้ว merge จาก API
/// 1. โหลด asset → clear + populate baseline (ได้ keys ครบ)
/// โหลดภาษาจาก API อย่างเดียว — ถ้าล้มเหลว throw error ทันที (ไม่มี fallback)
Future<void> loadLanguageFile(String languageCode) async {
  String apiPath;
  String apiPort;

  // ใช้ goApiUrl เพราะ language API อยู่ใน goapi ไม่ใช่ mainapi
  if (myAppConfig.goApiUrl.isNotEmpty) {
    apiPath = myAppConfig.goApiUrl;
    apiPort = myAppConfig.goApiPort;
  } else {
    apiPath = myAppConfig.serviceApi;
    apiPort = myAppConfig.servicePort;
  }

  if (apiPath.isEmpty) {
    throw Exception('ไม่พบ Backend URL — กรุณาตั้งค่า Server ก่อนใช้งาน');
  }

  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }
  if (apiPath.endsWith('/')) {
    apiPath = apiPath.substring(0, apiPath.length - 1);
  }
  String urlString = (apiPort.isNotEmpty) ? "$apiPath:$apiPort/api/language/$languageCode" : "$apiPath/api/language/$languageCode";
  final url = Uri.parse(urlString);

  if (kDebugMode) {
    AppLogger.debug("[Language] กำลังเรียก API: $urlString");
  }

  final response = await http.get(url).timeout(const Duration(seconds: 10));

  if (response.statusCode != 200) {
    throw Exception('โหลดภาษาจาก API ไม่สำเร็จ (status: ${response.statusCode})');
  }

  _applyLanguageData(json.decode(response.body), languageCode, 'API');
}

/// Apply ข้อมูลภาษาที่โหลดมาจาก API
void _applyLanguageData(Map<String, dynamic> jsonData, String languageCode, String source) {
  _languageMap.clear();

  jsonData.forEach((key, value) {
    final code = key.trim().toLowerCase();
    final text = value?.toString() ?? '';
    _languageMap[code] = text;
  });

  // rebuild languageSystemData จาก _languageMap ทั้งหมด
  languageSystemData = _languageMap.entries.map((e) => LanguageSystemModel(code: e.key, text: e.value)).toList();

  if (kDebugMode) {
    AppLogger.debug("[Language] Loaded from $source: $languageCode (${_languageMap.length} keys)");
  }
}

/// เลือกภาษาและโหลดไฟล์ภาษาที่ต้องการ
Future<void> languageSelect(String languageCode) async {
  userLanguage = languageCode;
  await loadLanguageFile(languageCode);
}

void deviceConfigLoad() {
  try {
    dynamic json = appConfig.getString("device");
    if (json != null) {
      deviceConfig = DeviceConfigModel.fromJson(jsonDecode(json));
    }
  } catch (e) {
    if (kDebugMode) {
      AppLogger.debug("deviceConfigLoad : $e");
    }
  }
}

Future<List<TimezonesModel>> getTimezonesList(filter) async {
  return timezonesListData;
}

String getLangName(String code, List<LanguageModel> languageList) {
  LanguageModel name = languageList.firstWhere(
    (element) => element.code == code,
    orElse: () => LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
  );
  return name.name!;
}

Future<void> deviceConfigSaveJson() async {
  try {
    await appConfig.setString("device", jsonEncode(deviceConfig.toJson()));
  } catch (e) {
    AppLogger.debug("deviceConfigSaveJson : $e");
  }
}

void listDataFontSizeChange() {
  // เก็บขนาดตัวอักษร (หน้าจอ List Data ทุกจอ)
  if (deviceConfig.listDataFontSize > 24) {
    deviceConfig.listDataFontSize = 12;
  } else {
    deviceConfig.listDataFontSize = deviceConfig.listDataFontSize + 2;
  }
  deviceConfigSaveJson();
}

/// ลดขนาด font ใน data list 1 px (min 8)
void listDataFontSizeDecrease() {
  if (deviceConfig.listDataFontSize > 8) {
    deviceConfig.listDataFontSize -= 1;
    deviceConfigSaveJson();
  }
}

/// เพิ่มขนาด font ใน data list 1 px (max 24)
void listDataFontSizeIncrease() {
  if (deviceConfig.listDataFontSize < 24) {
    deviceConfig.listDataFontSize += 1;
    deviceConfigSaveJson();
  }
}

/// ลด font scale หน้าจอ edit 5% (min 50%)
void editFontScaleDecrease() {
  if (deviceConfig.editFontScale > 50) {
    deviceConfig.editFontScale -= 5;
    deviceConfigSaveJson();
  }
}

/// เพิ่ม font scale หน้าจอ edit 5% (max 300%)
void editFontScaleIncrease() {
  if (deviceConfig.editFontScale < 300) {
    deviceConfig.editFontScale += 5;
    deviceConfigSaveJson();
  }
}

/// คืนค่า scale factor (เช่น 100 → 1.0, 80 → 0.8)
double get editFontScaleFactor => deviceConfig.editFontScale / 100;

void listDataLineSpaceChange() {
  // ขนาดช่องว่างข้อมูล
  if (deviceConfig.listDataLineSpace > 5) {
    deviceConfig.listDataLineSpace = 0;
  } else {
    deviceConfig.listDataLineSpace = deviceConfig.listDataLineSpace + 1;
  }
  deviceConfigSaveJson();
}

class NumberInputFormatter extends TextInputFormatter {
  final int maximumFractionDigits;
  NumberInputFormatter({this.maximumFractionDigits = 2}) : assert(maximumFractionDigits >= 0);
  @override
  TextEditingValue formatEditUpdate(TextEditingValue oldValue, TextEditingValue newValue) {
    var newText = newValue.text;
    var selectionOffset = newValue.selection.extent.offset;
    bool isNegative = false;
    if (newText.startsWith('-')) {
      newText = newText.substring(1);
      isNegative = true;
    }
    if (newText.isEmpty) {
      return newValue;
    }
    if (newText.indexOf('.') != newText.lastIndexOf('.')) {
      // inputted more than one dot.
      return oldValue;
    }
    if (newText.startsWith('.') && maximumFractionDigits > 0) {
      newText = '0$newText';
      selectionOffset += 1;
    }
    while (newText.length > 1 && !newText.startsWith('0.') && newText.startsWith('0')) {
      newText = newText.substring(1);
      selectionOffset -= 1;
    }
    if (decimalDigitsOf(newText) > maximumFractionDigits) {
      // delete the extra digits.
      newText = newText.substring(0, newText.indexOf('.') + 1 + maximumFractionDigits);
    }
    if (newValue.text.length == oldValue.text.length - 1 && oldValue.text.substring(newValue.selection.extentOffset, newValue.selection.extentOffset + 1) == ',') {
      // in this case, user deleted the thousands separator, we should delete the digit number before the cursor.
      newText = newText.replaceRange(newValue.selection.extentOffset - 1, newValue.selection.extentOffset, '');
      selectionOffset -= 1;
    }
    if (newText.endsWith('.')) {
      // in order to calculate the selection offset correctly, we delete the last decimal point first.
      newText = newText.replaceRange(newText.length - 1, newText.length, '');
    }
    int lengthBeforeFormat = newText.length;
    newText = _removeComma(newText);
    if (double.tryParse(newText) == null) {
      // invalid decimal number
      return oldValue;
    }
    newText = _addComma(newText);
    selectionOffset += newText.length - lengthBeforeFormat; // thousands separator newly added
    if (maximumFractionDigits > 0 && newValue.text.endsWith('.')) {
      // decimal point is at the last digit, we need to append it back.
      newText = '$newText.';
    }
    if (isNegative) {
      newText = '-$newText';
    }
    try {
      return TextEditingValue(
        text: newText,
        selection: TextSelection.collapsed(offset: math.min(selectionOffset, newText.length)),
      );
    } catch (e) {
      return oldValue;
    }
  }

  static int decimalDigitsOf(String text) {
    var index = text.indexOf('.');
    return index == -1 ? 0 : text.length - index - 1;
  }

  static String _addComma(String text) {
    StringBuffer sb = StringBuffer();
    var pointIndex = text.indexOf('.');
    String integerPart;
    String decimalPart;
    if (pointIndex >= 0) {
      integerPart = text.substring(0, pointIndex);
      decimalPart = text.substring(pointIndex);
    } else {
      integerPart = text;
      decimalPart = '';
    }
    List<String> parts = [];
    while (integerPart.length > 3) {
      parts.add(integerPart.substring(integerPart.length - 3));
      integerPart = integerPart.substring(0, integerPart.length - 3);
    }
    parts.add(integerPart);
    sb.writeAll(parts.reversed, ',');
    sb.write(decimalPart);
    return sb.toString();
  }

  static String _removeComma(String text) {
    return text.replaceAll(',', '');
  }
}

String getVatName(int number) {
  String returnValue = "";
  if (number == 0) {
    returnValue = language("vat_exclude");
  } else if (number == 1) {
    returnValue = language("vat_include");
  } else if (number == 2) {
    returnValue = language("vat_zero");
  } else if (number == 3) {
    returnValue = language("vat_none");
  }
  return returnValue;
}

String getInquiryName(int number) {
  String returnValue = "";
  if (number == 0) {
    returnValue = language("credit");
  } else if (number == 1) {
    returnValue = language("cash");
  }
  return returnValue;
}

bool isWideScreen() {
  return (deviceMode == DeviceModeEnum.androidTablet ||
      deviceMode == DeviceModeEnum.ipad ||
      deviceMode == DeviceModeEnum.macosDesktop ||
      deviceMode == DeviceModeEnum.linuxDesktop ||
      deviceMode == DeviceModeEnum.windowsDesktop);
}

Future<void> getDeviceModel(BuildContext context) async {
  final deviceInfo = DeviceInfoPlugin();

  String model = '';

  if (kIsWeb) {
    deviceMode = DeviceModeEnum.windowsDesktop;
    return;
  }

  if (Platform.isAndroid) {
    final androidInfo = await deviceInfo.androidInfo;

    model = androidInfo.model;

    // ignore: use_build_context_synchronously
    var shortestSide = MediaQuery.of(context).size.shortestSide;

    if (shortestSide > 600) {
      deviceMode = DeviceModeEnum.androidTablet;
    } else {
      deviceMode = DeviceModeEnum.androidPhone;
    }
  } else if (Platform.isIOS) {
    final iosInfo = await deviceInfo.iosInfo;

    model = iosInfo.model;

    model = model.toLowerCase();

    if (model.contains("iphone")) {
      deviceMode = DeviceModeEnum.iphone;
    } else if (model.contains("ipad")) {
      deviceMode = DeviceModeEnum.ipad;
    }
  } else if (Platform.isMacOS) {
    deviceMode = DeviceModeEnum.macosDesktop;
  } else if (Platform.isLinux) {
    deviceMode = DeviceModeEnum.linuxDesktop;
  } else if (Platform.isWindows) {
    deviceMode = DeviceModeEnum.windowsDesktop;
  }
}

/// ดึงค่าทศนิยมจำนวนสินค้าจากสาขา (default=2)
/// หมายเหตุ: mainapi ส่ง 0 เมื่อยังไม่ได้ตั้งค่า (Go zero value) จึงต้องตรวจ < 2
int getDecimalQuantity() {
  final val = companyBranchSelectData.decimalQuantity;
  return (val != null && val >= 2) ? val : 2;
}

/// ดึงค่าทศนิยมมูลค่าสินค้าจากสาขา (default=2)
int getDecimalPrice() {
  final val = companyBranchSelectData.decimalPrice;
  return (val != null && val >= 2) ? val : 2;
}

/// ดึงค่าทศนิยมมูลค่าเอกสารจากสาขา (default=2)
int getDecimalDocument() {
  final val = companyBranchSelectData.decimalDocument;
  return (val != null && val >= 2) ? val : 2;
}

/// ดึง locale ตาม base currency ที่กำหนดในสาขา
/// ใช้สำหรับ NumberFormat เพื่อแสดงตัวเลขตามรูปแบบของสกุลเงินนั้นๆ
String getLocaleForCurrency() {
  final currency = getBaseCurrency();
  const currencyLocaleMap = {
    'THB': 'th_TH',
    'USD': 'en_US',
    'EUR': 'de_DE',
    'GBP': 'en_GB',
    'JPY': 'ja_JP',
    'CNY': 'zh_CN',
    'KRW': 'ko_KR',
    'VND': 'vi_VN',
    'LAK': 'en_US',
    'KHR': 'en_US',
    'MMK': 'en_US',
    'SGD': 'en_SG',
    'MYR': 'ms_MY',
    'IDR': 'id_ID',
    'PHP': 'en_PH',
    'INR': 'en_IN',
    'AUD': 'en_AU',
    'NZD': 'en_NZ',
    'HKD': 'zh_HK',
    'TWD': 'zh_TW',
  };
  return currencyLocaleMap[currency] ?? 'en_US';
}

/// format ตัวเลขพร้อม comma separator
/// decimalPlaces: ถ้าไม่ระบุจะใช้ค่า decimalDocument จากสาขา
/// locale: ใช้ตาม base currency ที่กำหนดในสาขา
String formatNumber(double val, {int? decimalPlaces}) {
  if (val == 0) {
    return "";
  }
  final dp = decimalPlaces ?? getDecimalDocument();
  final pattern = dp > 0 ? '#,##0.${'0' * dp}' : '#,##0';
  return intl.NumberFormat(pattern, getLocaleForCurrency()).format(val);
}

/// format จำนวนสินค้า (ใช้ decimalQuantity จากสาขา)
String formatQuantity(double val) {
  if (val == 0) return "";
  final dp = getDecimalQuantity();
  final pattern = dp > 0 ? '#,##0.${'0' * dp}' : '#,##0';
  return intl.NumberFormat(pattern, getLocaleForCurrency()).format(val);
}

/// format ราคาสินค้าตามทศนิยมสาขา (ใช้ decimalPrice จากสาขา)
String formatUnitPrice(double val) {
  if (val == 0) return "";
  final dp = getDecimalPrice();
  final pattern = dp > 0 ? '#,##0.${'0' * dp}' : '#,##0';
  return intl.NumberFormat(pattern, getLocaleForCurrency()).format(val);
}

/// Helper function สำหรับ parse date อย่างปลอดภัย
/// ถ้า parse ไม่ได้จะ return DateTime.now() แทนที่จะ throw error
DateTime safeParseDatetime(String? dateString) {
  if (dateString == null || dateString.isEmpty) {
    return DateTime.now();
  }
  try {
    return DateTime.parse(dateString);
  } catch (e) {
    return DateTime.now();
  }
}

/// Helper function สำหรับแปลง dynamic value เป็น double อย่างปลอดภัย
/// รองรับ: int, double, num, String, และ null
/// ใช้เพื่อป้องกัน error: "type 'int' is not a subtype of type 'double'"
double safeToDouble(dynamic value, [double defaultValue = 0.0]) {
  if (value == null) return defaultValue;
  if (value is double) return value;
  if (value is int) return value.toDouble();
  if (value is num) return value.toDouble();
  if (value is String) {
    final parsed = double.tryParse(value);
    return parsed ?? defaultValue;
  }
  return defaultValue;
}

/// Helper function สำหรับแปลง dynamic value เป็น int อย่างปลอดภัย
/// รองรับ: int, double, num, String, และ null
int safeToInt(dynamic value, [int defaultValue = 0]) {
  if (value == null) return defaultValue;
  if (value is int) return value;
  if (value is double) return value.toInt();
  if (value is num) return value.toInt();
  if (value is String) {
    final parsed = int.tryParse(value);
    return parsed ?? defaultValue;
  }
  return defaultValue;
}

String formatThaiDateTime({required DateTime dateTime, bool showTime = false}) {
  String result = "";
  int day = dateTime.day;
  String month = getThaiMonthNameShort(dateTime.month);
  int year = (dateTime.year + 543) % 100; // แปลงเป็นปีพ.ศ. และเอาแค่ 2 หลักสุดท้าย
  result = '$day $month $year';
  if (showTime) {
    String time = intl.DateFormat('HH:mm').format(dateTime);
    result += ' $time';
  }
  return result;
}

/// Format DateTime เป็น string (alias ของ formatThaiDateTime)
String formatDateTime({required DateTime dateTime, bool showTime = false}) {
  return formatThaiDateTime(dateTime: dateTime, showTime: showTime);
}

/// ชื่อ ClickHouse database (dev: bcbidev, prod: bcbi)
/// ใช้ใน SQL queries ที่ส่งไป ClickHouse API
const String clickHouseDatabaseName = 'bcbidev';

/// ดึง base currency code
/// ลำดับความสำคัญ: 1. สาขา 2. ร้านค้า 3. Default THB
String getBaseCurrency() {
  try {
    // 1. ตรวจสอบจากสาขาก่อน
    if (companyBranchSelectData.baseCurrency != null && companyBranchSelectData.baseCurrency!.isNotEmpty) {
      return companyBranchSelectData.baseCurrency!.toUpperCase();
    }
    // 2. ถ้าสาขาไม่มี ใช้จากร้านค้า
    if (shopSelectData.settings != null && shopSelectData.settings!.baseCurrency != null) {
      return shopSelectData.settings!.baseCurrency!.toUpperCase();
    }
  } catch (e) {
    // Ignore error and return default
  }
  return 'THB';
}

/// ดึง yeartype (ประเภทปี)
/// ลำดับความสำคัญ: 1. สาขา 2. โปรไฟล์ 3. Default buddhist (พ.ศ.)
String getYearType() {
  try {
    // 1. ตรวจสอบจากสาขาก่อน
    if (companyBranchSelectData.yeartype != null && companyBranchSelectData.yeartype!.isNotEmpty) {
      return companyBranchSelectData.yeartype!;
    }
    // 2. ถ้าสาขาไม่มี ใช้จากโปรไฟล์
    if (profileData.yeartype != null && profileData.yeartype!.isNotEmpty) {
      return profileData.yeartype!;
    }
  } catch (e) {
    // Ignore error and return default
  }
  return 'buddhist'; // Default เป็น พ.ศ.
}

/// ดึง timezone
/// ลำดับความสำคัญ: 1. สาขา 2. Default Asia/Bangkok
String getBranchTimezone() {
  try {
    if (companyBranchSelectData.timezone != null && companyBranchSelectData.timezone!.isNotEmpty) {
      return companyBranchSelectData.timezone!;
    }
  } catch (e) {
    // Ignore error and return default
  }
  return 'Asia/Bangkok';
}

/// ดึง timezone offset
/// ลำดับความสำคัญ: 1. สาขา 2. Default +07:00
String getBranchTimezoneOffset() {
  try {
    if (companyBranchSelectData.timezoneoffset != null && companyBranchSelectData.timezoneoffset!.isNotEmpty) {
      return companyBranchSelectData.timezoneoffset!;
    }
  } catch (e) {
    // Ignore error and return default
  }
  return '+07:00';
}

/// ดึงภาษาของสาขา
/// ลำดับความสำคัญ: 1. สาขา 2. Default th
String getBranchLanguage() {
  try {
    if (companyBranchSelectData.language != null && companyBranchSelectData.language!.isNotEmpty) {
      return companyBranchSelectData.language!;
    }
  } catch (e) {
    // Ignore error and return default
  }
  return 'th';
}

/// ดึง symbol ของ base currency
/// ใช้ mapping ที่รู้จัก, default เป็น currency code
String getBaseCurrencySymbol() {
  final currencyCode = getBaseCurrency();
  const currencySymbols = {
    'THB': '฿',
    'USD': '\$',
    'EUR': '€',
    'GBP': '£',
    'JPY': '¥',
    'CNY': '¥',
    'KRW': '₩',
    'VND': '₫',
    'LAK': '₭',
    'KHR': '៛',
    'MMK': 'K',
    'SGD': 'S\$',
    'MYR': 'RM',
    'IDR': 'Rp',
    'PHP': '₱',
    'INR': '₹',
    'AUD': 'A\$',
    'NZD': 'NZ\$',
    'HKD': 'HK\$',
    'TWD': 'NT\$',
  };
  return currencySymbols[currencyCode] ?? currencyCode;
}

/// Format number with currency symbol
/// ใช้สำหรับแสดงตัวเลขพร้อมสัญลักษณ์สกุลเงิน
///
/// Examples:
/// - formatNumberWithCurrency(1234.56, currency: 'USD', currencySymbol: '$') → "$1,234.56"
/// - formatNumberWithCurrency(1234.56, currency: 'EUR', currencySymbol: '€') → "€1,234.56"
/// - formatNumberWithCurrency(1234.56) → "1,234.56" (no symbol)
String formatNumberWithCurrency(double value, {String? currency, String? currencySymbol}) {
  // ถ้าไม่มีสกุลเงิน หรือเป็นสกุลเงินหลัก → แสดงแค่ตัวเลข
  if (currency == null || currency.isEmpty || currency.toUpperCase() == getBaseCurrency()) {
    return formatNumber(value);
  }

  // ถ้ามี symbol → แสดง symbol + number
  if (currencySymbol != null && currencySymbol.isNotEmpty) {
    final formatter = intl.NumberFormat.currency(locale: 'en_US', symbol: currencySymbol, decimalDigits: 2);
    return formatter.format(value);
  }

  // ถ้าไม่มี symbol → แสดง currency code + number
  return '${currency.toUpperCase()} ${formatNumber(value)}';
}

/// Format price - ไม่แสดง symbol ถ้าเป็น base currency
/// ใช้แทน '฿${formatNumberRemoveRightZero(price)}'
///
/// Examples:
/// - Base currency = THB → "1,234.56" (ไม่มี ฿)
/// - Base currency = USD → "฿1,234.56" (แสดง ฿ เพราะราคาเป็น THB)
String formatPrice(double value) {
  final baseCurrency = getBaseCurrency();
  // ถ้า base currency เป็น THB → ไม่ต้องแสดง ฿
  if (baseCurrency == 'THB') {
    return formatNumberRemoveRightZero(value);
  }
  // ถ้า base currency ไม่ใช่ THB → แสดง ฿ เพื่อบอกว่าราคานี้เป็น THB
  return '฿${formatNumberRemoveRightZero(value)}';
}

String getThaiMonthNameFull(int month) {
  List<String> monthNames = ['มกราคม', 'กุมภาพันธ์', 'มีนาคม', 'เมษายน', 'พฤษภาคม', 'มิถุนายน', 'กรกฎาคม', 'สิงหาคม', 'กันยายน', 'ตุลาคม', 'พฤศจิกายน', 'ธันวาคม'];
  return monthNames[month - 1];
}

String getThaiMonthNameShort(int month) {
  List<String> monthNames = ['ม.ค.', 'ก.พ.', 'มี.ค.', 'เม.ย.', 'พ.ค.', 'มิ.ย.', 'ก.ค.', 'ส.ค.', 'ก.ย.', 'ต.ค.', 'พ.ย.', 'ธ.ค.'];
  return monthNames[month - 1];
}

String getEnglishMonthNameFull(int month) {
  List<String> monthNames = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
  return monthNames[month - 1];
}

String getEnglishMonthNameShort(int month) {
  List<String> monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  return monthNames[month - 1];
}

String formatname(String text) {
  if (text == "sale") {
    return text = "ชื่อลูกค้า   ";
  } else if (text == "purchase") {
    return text = "ชื่อเจ้าหนี้  ";
  }
  return text;
}

String activeLangName(List<LanguageDataModel> names) {
  String res = "";
  for (int i = 0; i < names.length; i++) {
    if (userLanguage == names[i].code) {
      res = names[i].name;
    }
  }
  return res;
}

String branchName() {
  // ชื่อสาขา
  return language("head_office");
}

String shopType() {
  // ประเภทร้าน
  return "ร้านโซลาว";
}

bool isValidDate(String dateString) {
  try {
    intl.DateFormat('yyyy-MM-dd').parseStrict(dateString);
    return true;
  } catch (e) {
    return false;
  }
}

double calcTextToNumber(String text) {
  double result = 0;
  String textTrim = text.trim();
  while (textTrim.contains(" ")) {
    textTrim = textTrim.replaceAll(" ", "");
  }
  if (textTrim.isNotEmpty) {
    textTrim = textTrim.replaceAll("X", "").replaceAll("x", "").replaceAll("+", "").replaceAll("-", "");
    result = double.parse(textTrim);
  }
  return result;
}

/// แปลงข้อความจาก TextField ที่มี comma เป็น double
/// รองรับรูปแบบ: "1,234.56", "1234.56", "1,234", etc.
double parseTextToNumber(String text) {
  if (text.isEmpty) return 0.0;
  // ลบ comma และช่องว่างทั้งหมด
  String cleanText = text.replaceAll(',', '').replaceAll(' ', '').trim();
  return double.tryParse(cleanText) ?? 0.0;
}

/// อัพเดทค่า TextEditingController ด้วยตัวเลขที่จัดรูปแบบแล้ว
/// ใช้สำหรับแสดงตัวเลขในรูปแบบ #,##0.00
void updateControllerWithFormattedNumber(TextEditingController controller, double value) {
  controller.text = formatNumber(value);
}

/// ตรวจสอบความถูกต้องของยอดเงิน
/// คืนค่า true ถ้ายอดเงินมากกว่า 0
bool isValidAmount(String text) {
  if (text.isEmpty) return false;
  double amount = parseTextToNumber(text);
  return amount > 0;
}

class DataTableHeader {
  String label;
  String code;
  double width;
  TextAlign textAlign;
  Alignment alignment;

  DataTableHeader({required this.label, required this.code, required this.width, this.textAlign = TextAlign.left, this.alignment = Alignment.centerLeft});
}

class TimeInputFormatter extends TextInputFormatter {
  @override
  TextEditingValue formatEditUpdate(TextEditingValue oldValue, TextEditingValue newValue) {
    final regExp = RegExp(r'^([01]?[0-9]|2[0-3]):[0-5][0-9]$');
    if (regExp.hasMatch(newValue.text)) {
      return newValue;
    }
    return oldValue;
  }
}

double calcDiscountFormula({required double totalAmount, required String discountText}) {
  double sumDiscount = 0.0;
  List<String> split = discountText.trim().replaceAll(" ", "").replaceAll(" ", "").split(",");
  for (int index = 0; index < split.length; index++) {
    String discount = split[index];
    double result = 0.0;
    if (discount.contains("%")) {
      // ลด %
      double? percent = double.tryParse(discount.replaceAll("%", ""));
      if (percent != null) {
        result = totalAmount * (percent / 100);
        sumDiscount += result;
        totalAmount -= result;
      }
    } else {
      // ลด จำนวนเงิน
      double? discountMoney = double.tryParse(discount);
      if (discountMoney != null) {
        sumDiscount += discountMoney;
        totalAmount -= discountMoney;
      }
    }
  }
  return sumDiscount;
}

Future<List<LanguageDataModel>> translateNames({required List<LanguageDataModel> namesData}) async {
  List<LanguageDataModel> result = [];
  for (int i = 0; i < config.languages.length; i++) {
    var translation = await translator.translate(namesData[0].name, to: config.languages[i].codeTranslator!);
    result.add(LanguageDataModel(code: config.languages[i].code!, name: translation.text));
  }
  return result;
}

String getDocType(int transflag) {
  String result = "";
  if (transflag == 12) {
    result = "transaction_purchase";
  } else if (transflag == 16) {
    result = "transaction_purchase_return";
  } else if (transflag == 44) {
    result = "transaction_sale";
  } else if (transflag == 48) {
    result = "transaction_sale_return";
  } else if (transflag == 56) {
    result = "transaction_stock_pick_up_product";
  } else if (transflag == 60) {
    result = "transaction_stock_receive_product";
  } else if (transflag == 58) {
    result = "transaction_returnproduct";
  } else if (transflag == 72) {
    result = "transaction_stock_transfer";
  } else if (transflag == 66) {
    result = "stock adjustment";
  } else {
    result = "Trans falg not found";
  }

  return result;
}

Future<Uint8List> fetchNetworkImage(String imageUrl) async {
  final response = await http.get(Uri.parse(imageUrl));
  if (response.statusCode == 200) {
    return response.bodyBytes;
  } else {
    throw Exception('Failed to load network image');
  }
}

class NumberToWordThai {
  static const String _baht = 'บาท';
  static const String _stang = 'สตางค์';
  static const String _zero = 'ศูนย์'; //0
  static const String _ed = 'เอ็ด'; //1
  static const String _hundred = 'ร้อย'; //100
  static const String _thousand = 'พัน'; //1000
  static const String _tenthousand = 'หมื่น'; //10 000
  static const String _hundredthousand = 'แสน'; //100 000
  static const String _million = 'ล้าน'; //1000 000
  static const String _billion = 'พันล้าน'; //1000 000 000

  ///numNames
  static const List<String> _numNames = [
    '',
    'หนึ่ง',
    'สอง',
    'สาม',
    'สี่',
    'ห้า',
    'หก',
    'เจ็ด',
    'แปด',
    'เก้า',
    'สิบ',
    'สิบเอ็ด',
    'สิบสอง',
    'สิบสาม',
    'สิบสี่',
    'สิบห้า',
    'สิบหก',
    'สิบเจ็ด',
    'สิบแปด',
    'สิบเก้า',
    'ยี่สิบ',
  ];

  ///tensNames
  static const List<String> _tensNames = ['', 'สิบ', 'ยี่สิบ', 'สามสิบ', 'สี่สิบ', 'ห้าสิบ', 'หกสิบ', 'เจ็ดสิบ', 'แปดสิบ', 'เก้าสิบ'];

  /// convertLessThanOneThousand
  static String _convertLessThanOneThousand(int number, [bool isLastThreeDigits = true]) {
    String soFar = '';

    if (number % 100 <= 20) {
      soFar = _numNames[number % 100];
      number = (number ~/ 100).toInt();
    } else {
      soFar = _numNames[number % 10];
      if (soFar == _numNames[1]) {
        soFar = _ed;
      }
      number = (number ~/ 10).toInt();
      soFar = _tensNames[number % 10] + soFar;
      number = (number ~/ 10).toInt();
    }
    if (number == 0) {
      return soFar;
    }
    return _numNames[number] + _hundred + soFar;
  }

  ///handle converter
  static String convert(double number) {
    Decimal decimalNumber = Decimal.parse(number.toString());
    String formattedNumber = decimalNumber.toStringAsFixed(2);
    if (number == 0) {
      return _zero;
    }

    final String strNumber = formattedNumber.split(".")[0].padLeft(12, '0');

    // XXXnnnnnnnnn
    final int billions = int.parse(strNumber.substring(0, 3));
    // nnnXXXnnnnnn
    final int millions = int.parse(strNumber.substring(3, 6));
    // nnnnnnXXXnnn
    final int hundredThousands = int.parse(strNumber.substring(6, 7));

    final int tenThousands = int.parse(strNumber.substring(7, 8));

    // nnnnnnnnnXXX
    final int thousands = int.parse(strNumber.substring(8, 9));

    // nnnnnnnnnXXX
    final int hundred = int.parse(strNumber.substring(9, 12));

    final int stang = int.parse(formattedNumber.toString().split(".")[1]);

    final String tradBillions = _getBillions(billions);
    String result = tradBillions + ((tradBillions != "") ? _baht : "");

    final String tradMillions = _getMillions(millions);
    result = result + tradMillions + ((tradMillions != "") ? _baht : "");

    final String tradHundredThousands = _getHundredThousands(hundredThousands);
    result = result + tradHundredThousands + ((tradHundredThousands != "") ? _baht : "");

    final String tradTenThousands = _getTenThousands(tenThousands);
    result = result + tradTenThousands + ((tradTenThousands != "") ? _baht : "");

    final String tradThousands = _getThousands(thousands);
    result = result + tradThousands + ((tradThousands != "") ? _baht : "");

    String tradThousand;
    tradThousand = _convertLessThanOneThousand(hundred);
    result = result + tradThousand + ((tradThousand != "") ? _baht : "");

    String tradStang;
    tradStang = _convertLessThanOneThousand(stang, false);
    result = result + tradStang + ((tradStang != "") ? _stang : "ถ้วน");

    // remove extra spaces!
    result = result.replaceAll(RegExp('\\s+'), ' ').replaceAll('\\b\\s{2,}\\b', ' ');
    return result.trim();
  }

  ///get Billions
  static String _getBillions(int billions) {
    String tradBillions;
    switch (billions) {
      case 0:
        tradBillions = '';
        break;
      case 1:
        tradBillions = _convertLessThanOneThousand(billions) + _billion;
        break;
      default:
        tradBillions = _convertLessThanOneThousand(billions) + _billion;
    }
    return tradBillions;
  }

  ///get Millions
  static String _getMillions(int millions) {
    String tradMillions;
    switch (millions) {
      case 0:
        tradMillions = '';
        break;
      case 1:
        tradMillions = _convertLessThanOneThousand(millions) + _million;
        break;
      default:
        tradMillions = _convertLessThanOneThousand(millions) + _million;
    }
    return tradMillions;
  }

  ///get Hundred Thousands
  static String _getHundredThousands(int hundredThousands) {
    String tradHundredThousands;
    switch (hundredThousands) {
      case 0:
        tradHundredThousands = '';
        break;
      case 1:
        tradHundredThousands = _convertLessThanOneThousand(hundredThousands) + _hundredthousand;
        break;
      default:
        tradHundredThousands = _convertLessThanOneThousand(hundredThousands) + _hundredthousand;
    }

    return tradHundredThousands;
  }

  static String _getTenThousands(int tenThousands) {
    String tradTenThousands;
    switch (tenThousands) {
      case 0:
        tradTenThousands = '';
        break;
      case 1:
        tradTenThousands = _convertLessThanOneThousand(tenThousands) + _tenthousand;
        break;
      default:
        tradTenThousands = _convertLessThanOneThousand(tenThousands) + _tenthousand;
    }

    return tradTenThousands;
  }

  static String _getThousands(int hundredThousands) {
    String tradHundredThousands;
    switch (hundredThousands) {
      case 0:
        tradHundredThousands = '';
        break;
      case 1:
        tradHundredThousands = _convertLessThanOneThousand(hundredThousands) + _thousand;
        break;
      default:
        tradHundredThousands = _convertLessThanOneThousand(hundredThousands) + _thousand;
    }

    return tradHundredThousands;
  }
}

/// แปลงตัวเลขเป็นตัวอักษรภาษาอังกฤษ (สกุลเงินบาท/สตางค์)
class NumberToWordEnglish {
  static const String _baht = 'Baht';
  static const String _satang = 'Satang';
  static const String _only = 'Only';

  static const List<String> _ones = [
    '',
    'One',
    'Two',
    'Three',
    'Four',
    'Five',
    'Six',
    'Seven',
    'Eight',
    'Nine',
    'Ten',
    'Eleven',
    'Twelve',
    'Thirteen',
    'Fourteen',
    'Fifteen',
    'Sixteen',
    'Seventeen',
    'Eighteen',
    'Nineteen',
  ];

  static const List<String> _tens = ['', '', 'Twenty', 'Thirty', 'Forty', 'Fifty', 'Sixty', 'Seventy', 'Eighty', 'Ninety'];

  static String _convertHundreds(int n) {
    if (n == 0) return '';
    String result = '';
    if (n >= 100) {
      result += '${_ones[n ~/ 100]} Hundred';
      n %= 100;
      if (n > 0) result += ' ';
    }
    if (n >= 20) {
      result += _tens[n ~/ 10];
      if (n % 10 > 0) result += ' ${_ones[n % 10]}';
    } else if (n > 0) {
      result += _ones[n];
    }
    return result;
  }

  static String _convertWhole(int n) {
    if (n == 0) return 'Zero';
    final List<String> parts = [];
    if (n >= 1000000000) {
      parts.add('${_convertHundreds(n ~/ 1000000000)} Billion');
      n %= 1000000000;
    }
    if (n >= 1000000) {
      parts.add('${_convertHundreds(n ~/ 1000000)} Million');
      n %= 1000000;
    }
    if (n >= 1000) {
      parts.add('${_convertHundreds(n ~/ 1000)} Thousand');
      n %= 1000;
    }
    if (n > 0) {
      parts.add(_convertHundreds(n));
    }
    return parts.join(' ');
  }

  static String convert(double number) {
    if (number == 0) return 'Zero $_baht $_only';
    final Decimal d = Decimal.parse(number.toStringAsFixed(2));
    final int intPart = d.truncate().toBigInt().toInt();
    final int satang = ((d - d.truncate()) * Decimal.fromInt(100)).round().toBigInt().toInt();
    String result = '${_convertWhole(intPart)} $_baht';
    if (satang > 0) {
      result += ' and ${_convertHundreds(satang)} $_satang';
    } else {
      result += ' $_only';
    }
    return result;
  }
}

/// Dispatcher: แปลงตัวเลขเป็นตัวอักษรตามภาษาที่ user เลือก
String numberToWordByLang(double amount) {
  switch (userLanguage) {
    case 'th':
      return NumberToWordThai.convert(amount);
    case 'en':
    default:
      return NumberToWordEnglish.convert(amount);
  }
}

int checkDeveloperMode(String environment) {
  switch (environment) {
    case Environment.DEV:
      return 1;
    case Environment.PROD:
      return 0;
    case Environment.UAT:
      return 0;
    default:
      return 1;
  }
}

String docDateTimeFormateDDMMYYY(String docdatetime) {
  DateTime docDateTimeFormatDetail = DateTime.parse(docdatetime);
  // ใช้ getYearType() เพื่อตรวจสอบจากสาขาก่อน
  if (getYearType() == "buddhist") {
    return dateTimeBuddhist(docDateTimeFormatDetail, format: DateTimeFormatEnum.date);
  } else {
    return intl.DateFormat('dd/MM/yyyy').format(docDateTimeFormatDetail);
  }
}

enum MainMenuEnum {
  /// Dashboard/Home page
  home,

  /// Customer Relationship Management
  crm,

  /// Sales module
  sell,

  /// Purchase module
  buy,

  /// Inventory/Warehouse management
  inventory,

  /// Delivery/Shipping management
  delivery,

  /// Accounts payable
  payable,

  /// Accounts receivable
  receivable,

  /// Financial management
  finance,

  /// Asset and depreciation management
  asset,

  /// General ledger
  ledger,

  /// Master data management
  master,

  /// System settings
  settings,

  /// Device permissions and management
  permissionDevices,
}

List<LanguageDataModel> convertToLanguageDataList(List<dynamic> names) {
  return names.map((item) {
    if (item is Map<String, dynamic>) {
      return LanguageDataModel(code: item['code'] ?? '', name: item['name'] ?? '');
    }
    return LanguageDataModel(code: '', name: '');
  }).toList();
}

String getTransFlagText(int flag) {
  switch (flag) {
    case 6:
      return transactionName(TransactionTypeEnum.purchaseorder);
    case 12:
      return transactionName(TransactionTypeEnum.purchase);
    case 16:
      return transactionName(TransactionTypeEnum.purchasereturn);
    case 36:
      return transactionName(TransactionTypeEnum.quotation);
    case 44:
      return transactionName(TransactionTypeEnum.sale);
    case 48:
      return transactionName(TransactionTypeEnum.salereturn);
    case 54:
      return transactionName(TransactionTypeEnum.stockbalance);
    case 56:
      return transactionName(TransactionTypeEnum.stockpickupproduct);
    case 58:
      return transactionName(TransactionTypeEnum.stockreturnproduct);
    case 60:
      return transactionName(TransactionTypeEnum.stockreceiveproduct);
    case 66:
      return transactionName(TransactionTypeEnum.adjust);
    case 68:
      return language("transaction_stock_count");
    case 72:
      return transactionName(TransactionTypeEnum.stocktransfer);
    case 310:
      return transactionName(TransactionTypeEnum.purchasepartial);
    case 866:
      return language("transaction_sale_cancel");
    default:
      return language("unknown");
  }
}

IconData getTransFlagIcon(int flag) {
  switch (flag) {
    case 56:
      return Icons.shopping_cart_checkout;
    case 58:
      return Icons.assignment_return;
    case 66:
      return Icons.inventory;
    case 72:
      return Icons.swap_horiz;
    default:
      return Icons.add_shopping_cart;
  }
}

Future<Map<String, dynamic>> clickhouseSelect(String query) async {
  var httpClient = http.Client();
  // ปรับให้ใช้ absolute URL เสมอ (ห้ามใช้ localhost/127.0.0.1 บน web)
  String apiPath = myAppConfig.serviceClickhouse;
  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }
  var urlString = "$apiPath/select";
  var url = Uri.parse(urlString);

  if (kDebugMode) {
    AppLogger.debug("query : $query");
  }
  try {
    var response = await httpClient.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode({"query": query}));

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      return {"status": "error", "code": response.statusCode, "message": "เกิดข้อผิดพลาดในการเรียก API: ${response.reasonPhrase}"};
    }
  } catch (e) {
    return {"status": "error", "code": 500, "message": "${language("error_connecting")}: $e"};
  } finally {
    httpClient.close();
  }
}

Future<Map<String, dynamic>> reportServicePost(String jsonCommand) async {
  var httpClient = http.Client();
  // ใช้ goapi เพราะ /reportpost endpoint อยู่บน goapi ไม่ใช่ mainapi
  String apiPath = myAppConfig.goApiUrl;
  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }
  var urlString = (myAppConfig.goApiPort.isNotEmpty) ? "$apiPath:${myAppConfig.goApiPort}/reportpost" : "$apiPath/reportpost";
  var url = Uri.parse(urlString);

  if (kDebugMode) {
    AppLogger.debug("reportServicePost (goapi) : $urlString");
  }

  // เพิ่ม Authorization header ถ้ามี token
  Map<String, String> headers = {'Content-Type': 'application/json'};
  String token = appConfig.getString("token") ?? '';
  if (token.isNotEmpty) {
    headers['Authorization'] = 'Bearer $token';
  }

  try {
    var response = await httpClient.post(url, headers: headers, body: jsonEncode(jsonCommand)).timeout(const Duration(minutes: 5));

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      return {"status": "error", "code": response.statusCode, "message": "เกิดข้อผิดพลาดในการเรียก API: ${response.reasonPhrase}"};
    }
  } catch (e) {
    return {"status": "error", "code": 500, "message": "${language("error_connecting")}: $e"};
  } finally {
    httpClient.close();
  }
}

/// เชื่อมต่อ SSE stream สำหรับ rebuild progress
/// คืนค่า Stream ของ Map<String, dynamic> ที่มี step, total_steps, step_name, status, progress
Stream<Map<String, dynamic>> connectRebuildSSE(String jobId) async* {
  String apiPath = myAppConfig.goApiUrl;
  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }
  var urlString = (myAppConfig.goApiPort.isNotEmpty) ? "$apiPath:${myAppConfig.goApiPort}/rebuild/progress/$jobId" : "$apiPath/rebuild/progress/$jobId";

  if (kDebugMode) {
    AppLogger.debug("connectRebuildSSE: $urlString");
  }

  final client = http.Client();
  try {
    final request = http.Request('GET', Uri.parse(urlString));

    // เพิ่ม Authorization header ถ้ามี token
    String token = appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      request.headers['Authorization'] = 'Bearer $token';
    }
    request.headers['Accept'] = 'text/event-stream';
    request.headers['Cache-Control'] = 'no-cache';

    final response = await client.send(request);

    if (response.statusCode != 200) {
      yield {"status": "error", "message": "เชื่อมต่อ SSE ล้มเหลว: HTTP ${response.statusCode}"};
      return;
    }

    // อ่าน SSE stream — แต่ละ event อยู่ในรูป "data: {...}\n\n"
    String buffer = '';
    await for (var chunk in response.stream.transform(const Utf8Decoder())) {
      buffer += chunk;
      // แยก events ที่สมบูรณ์ (จบด้วย \n\n)
      while (buffer.contains('\n\n')) {
        final eventEnd = buffer.indexOf('\n\n');
        final eventStr = buffer.substring(0, eventEnd);
        buffer = buffer.substring(eventEnd + 2);

        // Parse SSE lines
        for (var line in eventStr.split('\n')) {
          if (line.startsWith('data: ')) {
            try {
              final jsonData = jsonDecode(line.substring(6));
              if (jsonData is Map<String, dynamic>) {
                yield jsonData;
                // ถ้า status เป็น completed หรือ error ให้หยุด
                if (jsonData['status'] == 'completed' || jsonData['status'] == 'error') {
                  return;
                }
              }
            } catch (e) {
              AppLogger.err('SSE parse error: $e');
            }
          }
        }
      }
    }
  } catch (e) {
    yield {"status": "error", "message": "SSE connection error: $e"};
  } finally {
    client.close();
  }
}

Future<Map<String, dynamic>> reportServiceGet(Map<String, String> jsonCommand) async {
  var httpClient = http.Client();
  // ปรับให้ใช้ absolute URL เสมอ (ห้ามใช้ localhost/127.0.0.1 บน web)
  String apiPath = myAppConfig.reportApiPath;
  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }
  var urlString = (myAppConfig.reportApiPort.isNotEmpty) ? "$apiPath:${myAppConfig.reportApiPort}/reportget" : "$apiPath/reportget";

  // แปลง jsonCommand เป็น query parameters
  var url = Uri.parse(urlString).replace(queryParameters: jsonCommand);

  if (kDebugMode) {
    AppLogger.debug("reportServiceCaller : $url");
  }

  try {
    var response = await httpClient.get(
      url,
      headers: {'Content-Type': 'application/json'},
      // ไม่ต้องใส่ body สำหรับ GET
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      return {"status": "error", "code": response.statusCode, "message": "เกิดข้อผิดพลาดในการเรียก API: ${response.reasonPhrase}"};
    }
  } catch (e) {
    return {"status": "error", "code": 500, "message": "${language("error_connecting")}: $e"};
  } finally {
    httpClient.close();
  }
}

List<int> decodeResponse(Uint8List responseBodyBytes) {
  var decodedBytes = GZipDecoder().decodeBytes(responseBodyBytes);
  return decodedBytes;
}

Future<Map<String, dynamic>> reportServiceGetBinary(Map<String, String> jsonCommand) async {
  var httpClient = http.Client();
  // ปรับให้ใช้ absolute URL เสมอ (ห้ามใช้ localhost/127.0.0.1 บน web)
  String apiPath = myAppConfig.reportApiPath;
  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }
  var urlString = (myAppConfig.reportApiPort.isNotEmpty) ? "$apiPath:${myAppConfig.reportApiPort}/reportget" : "$apiPath/reportget";

  // แปลง jsonCommand เป็น query parameters
  var url = Uri.parse(urlString).replace(queryParameters: jsonCommand);

  if (kDebugMode) {
    AppLogger.debug("reportServiceGetBin : $url");
  }

  try {
    var response = await httpClient.get(
      url,
      headers: {
        'Accept': 'application/octet-stream', // ระบุว่าเราต้องการรับข้อมูลไบนารี่
      },
    );

    if (response.statusCode == 200) {
      // ส่งคืนข้อมูลไบนารี่เป็น Uint8List และข้อมูลเพิ่มเติม
      // unzip
      var bytes = decodeResponse(response.bodyBytes);

      return {"status": "success", "code": 200, "binaryData": bytes, "contentType": response.headers["content-type"] ?? "application/octet-stream", "contentLength": response.contentLength};
    } else {
      // กรณีเกิดข้อผิดพลาด
      try {
        // ลองแปลงข้อความข้อผิดพลาดเป็น JSON (ถ้ามี)
        var errorResponse = String.fromCharCodes(response.bodyBytes);
        var errorJson = jsonDecode(errorResponse);
        return {"status": "error", "code": response.statusCode, "message": errorJson["message"] ?? language("error.api_call_failed"), "error": errorJson};
      } catch (e) {
        // ถ้าไม่สามารถแปลงเป็น JSON ได้ ให้ส่งข้อความผิดพลาดปกติ
        return {"status": "error", "code": response.statusCode, "message": "เกิดข้อผิดพลาดในการเรียก API: ${response.reasonPhrase}"};
      }
    }
  } catch (e) {
    return {"status": "error", "code": 500, "message": "${language("error_connecting")}: $e"};
  } finally {
    httpClient.close();
  }
}

String googleSignInClientId = "183871505957-7co3akl7smnhbpr4009vgmd7ht1iesfn.apps.googleusercontent.com";
String googleSignInClientSecret = "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";

// Google OAuth สำหรับ Windows Desktop เท่านั้น (test06 project)
String googleSignInClientIdWindows = "381991785720-gji10vhk3hubarpbr4jlm9rtfig7oqhc.apps.googleusercontent.com";
String googleSignInClientSecretWindows = "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";

String getShopId() {
  String shopid = appConfig.getString('shopid') ?? "";
  return shopid;
}

String getBranchGuidFixed() {
  String branchGuid = appConfig.getString('branch_guidfixed') ?? "";
  return branchGuid;
}

void setShopName(List<LanguageDataModel> languageData) {
  String jsonString = jsonEncode(languageData);
  appConfig.setString('shopname', jsonString);
}

String getShopName() {
  List<LanguageDataModel> languageData = [];
  String shopNameJson = appConfig.getString('shopname') ?? "";
  if (shopNameJson.isNotEmpty) {
    languageData = convertToLanguageDataList(jsonDecode(shopNameJson));
  } else {
    if (kDebugMode) {
      AppLogger.debug("shopname is empty, please set it in appConfig");
    }
  }
  String shopName = activeLangName(languageData);
  return shopName;
}

/// คืน GoAPI base URL พร้อม /goapi prefix (ไม่มี trailing slash)
/// ใช้แทน _baseUrl ที่สร้างเองในแต่ละ service
String get goApiBaseUrl {
  String apiPath;
  String apiPort;

  if (myAppConfig.goApiUrl.isNotEmpty) {
    apiPath = myAppConfig.goApiUrl;
    apiPort = myAppConfig.goApiPort;
  } else {
    apiPath = myAppConfig.reportApiPath;
    apiPort = myAppConfig.reportApiPort;
  }

  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }

  String baseUrl;
  if (apiPort.isNotEmpty) {
    baseUrl = "$apiPath:$apiPort";
  } else {
    baseUrl = apiPath;
  }

  if (!baseUrl.contains('/goapi')) {
    baseUrl = baseUrl.endsWith('/') ? '${baseUrl}goapi' : '$baseUrl/goapi';
  }
  if (baseUrl.endsWith('/')) {
    baseUrl = baseUrl.substring(0, baseUrl.length - 1);
  }

  return baseUrl;
}

String goApiUrlPath(String endpoint) {
  String apiPath;
  String apiPort;

  // ใช้ goApiUrl ถ้ามี (goapi รวมเข้า mainapi แล้ว — path /goapi)
  if (myAppConfig.goApiUrl.isNotEmpty) {
    apiPath = myAppConfig.goApiUrl;
    apiPort = myAppConfig.goApiPort;
  } else {
    apiPath = myAppConfig.reportApiPath;
    apiPort = myAppConfig.reportApiPort;
  }

  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }

  // สร้าง base URL พร้อม port
  String baseUrl;
  if (apiPort.isNotEmpty) {
    baseUrl = "$apiPath:$apiPort";
  } else {
    baseUrl = apiPath;
  }

  // GoAPI routes ต้องผ่าน /goapi prefix (Caddy route หรือ MainAPI integration)
  // ถ้า URL ไม่มี /goapi → เติมให้อัตโนมัติ
  if (!baseUrl.contains('/goapi')) {
    baseUrl = baseUrl.endsWith('/') ? '${baseUrl}goapi' : '$baseUrl/goapi';
  }

  return '$baseUrl/$endpoint';
}

/// แปลง file URL จาก API response เป็น full URL
/// - ถ้าเป็น Docker internal URL (seaweedfs-filer) → rewrite ผ่าน goapi S3 proxy
/// - ถ้าเป็น full URL (http/https) → ใช้ตรงๆ (backward compatible)
/// - ถ้าเป็น relative path (เช่น /s3/file/...) → ต่อ goapi base URL ข้างหน้า
String resolveFileUrl(String url) {
  if (url.isEmpty) return url;

  // Docker internal hostname → rewrite ผ่าน goapi S3 proxy
  if (url.contains('seaweedfs-filer')) {
    final uri = Uri.parse(url);
    final pathSegments = uri.pathSegments;
    // path = /bucket/key... → ข้ามชื่อ bucket (segment แรก) ส่งเฉพาะ key ผ่าน proxy
    if (pathSegments.length > 1) {
      final key = pathSegments.sublist(1).join('/');
      return goApiUrlPath('s3/file/$key');
    }
    return url;
  }

  if (url.startsWith('http://') || url.startsWith('https://')) return url;

  // สร้าง goapi base URL
  String apiPath = myAppConfig.goApiUrl;
  String apiPort = myAppConfig.goApiPort;

  if (apiPath.isEmpty) {
    apiPath = myAppConfig.reportApiPath;
    apiPort = myAppConfig.reportApiPort;
  }

  if (!apiPath.startsWith('http')) {
    apiPath = 'https://$apiPath';
  }

  String baseUrl = apiPort.isNotEmpty ? "$apiPath:$apiPort" : apiPath;

  if (!url.startsWith('/')) {
    url = '/$url';
  }

  return '$baseUrl$url';
}

/// Persistent HTTP client สำหรับ goApi (reuse TCP connection)
final http.Client _goApiHttpClient = http.Client();

Future<Object> goApiPost(String url, Map<String, dynamic> payLoad) async {
  var uri = Uri.parse(url);
  try {
    // สร้าง headers พร้อม Authorization token
    Map<String, String> headers = {'Content-Type': 'application/json'};
    String token = appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }

    var response = await _goApiHttpClient.post(uri, headers: headers, body: jsonEncode(payLoad));

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      return {"status": "error", "code": response.statusCode, "message": "เกิดข้อผิดพลาดในการเรียก API: ${response.reasonPhrase}"};
    }
  } catch (e) {
    return {"status": "error", "code": 500, "message": "${language("error_connecting")}: $e"};
  }
}

Future<Object> getApiServiceVersion() async {
  var httpClient = http.Client();
  // แปลง jsonCommand เป็น query parameters
  // ส่ง shopid ไปด้วย
  var url = Uri.parse(goApiUrlPath("version")).replace(queryParameters: {"shopid": getShopId(), "refresh": DateTime.now().millisecondsSinceEpoch.toString()});

  try {
    // สร้าง headers พร้อม Authorization token
    Map<String, String> headers = {'Content-Type': 'application/json'};
    String token = appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }

    var response = await httpClient.get(url, headers: headers);

    if (response.statusCode == 200) {
      return response.body.toString();
    } else {
      return {"status": "error", "code": response.statusCode, "message": "เกิดข้อผิดพลาดในการเรียก API: ${response.reasonPhrase}"};
    }
  } catch (e) {
    return {"status": "error", "code": 500, "message": "${language("error_connecting")}: $e"};
  } finally {
    httpClient.close();
  }
}

String formatNumberRemoveRightZero(double val, {int? decimalPlaces}) {
  if (val == 0) {
    return "";
  }
  // ใช้ '#' เพื่อแสดงทศนิยมเฉพาะเมื่อจำเป็น (ไม่แสดงเลข 0 ที่ไม่จำเป็น)
  final dp = decimalPlaces ?? getDecimalDocument();
  final pattern = '#,##0.${'#' * dp}';
  return intl.NumberFormat(pattern, getLocaleForCurrency()).format(val);
}

enum ProcessState { idle, processing, success, error }

// เพิ่มฟังก์ชันสำหรับอ่าน shop_info จาก local storage
String getShopsIdFromLocalStorage() {
  try {
    String? shopInfoString = appConfig.getString("shop_info");

    if (shopInfoString != null && shopInfoString.isNotEmpty) {
      Map<String, dynamic> shopInfoJson = jsonDecode(shopInfoString);
      ShopModel shopModel = ShopModel.fromJson(shopInfoJson);

      // ถ้า posproductcentertype = 1 ให้ส่งค่า guidfixed และ mainshopid
      if (shopModel.posproductcentertype == 1 && shopModel.guidfixed?.isNotEmpty == true && shopModel.mainshopid?.isNotEmpty == true) {
        return Uri.encodeComponent('${shopModel.guidfixed},${shopModel.mainshopid}');
      }
    }
  } catch (e) {
    if (kDebugMode) {
      AppLogger.error('Error reading shop_info: $e');
    }
  }

  return '';
}

/// Parse และประมวลผล JSON data ของ Transaction Model
/// ใช้สำหรับ parse paymentdetailraw และ extrajson ของ transaction details
void parseTransactionJsonData(TransactionModel transaction) {
  // Parse paymentdetailraw
  if (transaction.paymentdetailraw != null && transaction.paymentdetailraw!.isNotEmpty) {
    try {
      final List<dynamic> jsonStr = jsonDecode(transaction.paymentdetailraw!);
      transaction.paymentStructs = jsonStr.map((e) => BillPayStruct.fromJson(e)).toList();
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error parsing paymentdetailraw JSON: $e');
      }
    }
  }

  // Parse extrajson ของแต่ละ detail
  if (transaction.details != null) {
    for (var detail in transaction.details!) {
      if (detail.extrajson != null && detail.extrajson! != "[]" && detail.extrajson!.isNotEmpty) {
        try {
          List<dynamic> jsonData = jsonDecode(detail.extrajson!);
          detail.extrajsonlist = jsonData.map((e) => ExtraJsonListModel.fromJson(e)).toList();

          // Parse item_name ของแต่ละ extra
          for (var extra in detail.extrajsonlist!) {
            if (extra.item_name != null && extra.item_name! != "[]" && extra.item_name!.isNotEmpty) {
              try {
                List<dynamic> itemNameData = jsonDecode(extra.item_name!);
                extra.itemnames = itemNameData.map((e) => LanguageDataModel.fromJson(e)).toList();
              } catch (e) {
                if (kDebugMode) {
                  AppLogger.error('Error parsing item_name JSON: $e');
                }
              }
            }
          }
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error parsing extrajson JSON: $e');
          }
        }
      }
    }
  }
}

void gotoMainMenu(BuildContext context) {
  // ปิดหน้าต่างทั้งหมดที่เปิดอยู่และกลับไปหน้าแรก (Main Menu)
  Navigator.of(context).popUntil((route) => route.isFirst);
}

String getThaiDayNameFull(int weekday) {
  List<String> dayNames = [
    'จันทร์', // Monday
    'อังคาร', // Tuesday
    'พุธ', // Wednesday
    'พฤหัสบดี', // Thursday
    'ศุกร์', // Friday
    'เสาร์', // Saturday
    'อาทิตย์', // Sunday
  ];
  return dayNames[weekday - 1];
}

String getThaiDayNameShort(int weekday) {
  List<String> dayNames = [
    'จ', // Monday
    'อ', // Tuesday
    'พ', // Wednesday
    'พฤ', // Thursday
    'ศ', // Friday
    'ส', // Saturday
    'อา', // Sunday
  ];
  return dayNames[weekday - 1];
}

String getEnglishDayNameFull(int weekday) {
  List<String> dayNames = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
  return dayNames[weekday - 1];
}

String getEnglishDayNameShort(int weekday) {
  List<String> dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  return dayNames[weekday - 1];
}
