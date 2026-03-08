import 'package:uuid/uuid.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// Factory class สำหรับสร้าง TransactionModel เปล่าสำหรับ PO
///
/// แยกออกมาจาก clearScreenData() เพื่อลดความซับซ้อนของหน้าจอหลัก
class POModelFactory {
  /// สร้าง TransactionModel เปล่าสำหรับ PO ใหม่
  ///
  /// [transflag] - flag บอกประเภท transaction (PO = 6)
  /// [calcflag] - flag สำหรับคำนวณ (PO = 1)
  /// [vattype] - ประเภทภาษี (จาก config)
  /// [inquirytype] - ประเภทการสอบถาม (จาก config)
  /// [timezoneInfo] - ข้อมูล timezone จาก global.getLocalTimezoneInfo()
  static TransactionModel createDefault({
    required int transflag,
    required int vattype,
    required int inquirytype,
    required Map<String, dynamic> timezoneInfo,
  }) {
    const uuid = Uuid();

    return TransactionModel(
      shopid: global.apiShopCode,
      guidref: uuid.v4(),
      docno: '', // mainapi จะสร้าง docno เอง
      description: '',
      docdatetime: DateTime.now().toUtc().toIso8601String(),
      docdatelocal: timezoneInfo['docdatelocal'],
      doctimelocal: timezoneInfo['doctimelocal'],
      timezone: timezoneInfo['timezone'],
      docrefno: '',
      docrefdate: DateTime.now().toUtc().toIso8601String(),
      docreftype: 0,
      doctype: 0,
      vattype: vattype,
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
      inquirytype: inquirytype,
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
      branch: BranchModel(
        guidfixed: global.companyBranchSelectData.guidfixed,
        code: global.companyBranchSelectData.code,
        names: global.companyBranchSelectData.names,
      ),
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
      creditdays: 0,
      duedate: '',
    );
  }

  /// สร้าง TransactionModel สำหรับ PO พร้อมค่า default ของ PO
  static TransactionModel createForPO() {
    return createDefault(
      transflag: 6, // PO transflag
      vattype: global.config.vattypepurchase,
      inquirytype: global.config.inquirytypepurchase,
      timezoneInfo: global.getLocalTimezoneInfo(),
    );
  }
}
