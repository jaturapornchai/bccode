// ignore_for_file: non_constant_identifier_names

import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/wht_model.dart';
import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;
part 'transaction_model.g.dart';

@JsonSerializable(explicitToJson: true)
class TransactionModel {
  String? guidfixed;
  String? shopid;
  late String guidref;
  late String docno;
  late String docdatetime;

  /// วันที่ local ของ frontend สำหรับสร้าง docno (เช่น "20260107" หรือ "2026-01-07")
  @JsonKey(name: 'docdatelocal')
  String? docdatelocal;

  /// เวลา local ของ frontend (เช่น "08:30:00")
  @JsonKey(name: 'doctimelocal')
  String? doctimelocal;

  /// timezone ของ frontend (เช่น "Asia/Bangkok")
  @JsonKey(name: 'timezone')
  String? timezone;

  late String docrefno;
  late String docrefdate;
  late int docreftype;
  late int doctype;

  /// ประเภทภาษี 0 = ราคาไม่รวมภาษี , 1 = ราคารวมภาษี , 2 = ภาษีอัตราศูนย์ , 3 = ไม่กระทบภาษี
  late int vattype;

  /// inquirytype ประเภทการขาย
  /// 0 = เครดิต , 1 = เงินสด
  /// inquirytype transflag 56=เบิก
  /// 0 = เบิกผลิต , 1 = เบิกใช้เอง , เบิกของเสียหาย , 9 = อื่น ๆ
  late int inquirytype;
  late int transflag; // 12=ซื้อ
  late String custcode;
  late List<LanguageDataModel>? custnames;
  late String salecode;
  late String salename;
  late String discountword;
  late double totalcost;
  late double totalvalue; //มูลค่ารวมสินค้าที่ลดรายตัวแล้ว
  late double totaldiscount; //ส่วนลดท้ายบิล
  late double totalvatvalue; //มูลค่าภาษี
  late double totalbeforevat; //มูลค่าก่อนหักภาษี
  late double totalaftervat; //มูลค่าหลังหักภาษี
  late double totalexceptvat; //มูลค่ายกเว้นภาษีเทียบitemtype
  late double totalamount; //มูลค่ารวม
  late String cashiercode;
  late String posid;
  late String membercode;
  bool? isref;
  bool? islocked;
  bool? isclosed;
  int? iscomparedsuccess;
  bool? isclosedmanual;
  String? closedmanualByCode;
  String? closedmanualByName;
  String? closedmanualAt;
  String? closedmanualReason;
  int? detailcount;

  /// บันทึกยอดเงินเอง
  late bool ismanualamount;

  ///เลขที่ใบกำกับภาษี
  late String taxdocno;

  ///วันที่ใบกำกับภาษี
  late String taxdocdate;
  late double vatrate;
  late bool iscancel;
  late int status;
  List<TransactionDetailModel>? details;
  TransactionPayModel? paymentdetail;
  String? paymentdetailraw;
  List<BillPayStruct>? paymentStructs;
  double? paycashamount;

  /// หมายเหตุ
  String? description;

  //0=บิลทั่วไปไม่มีภาษี,1=ใบเสร็จรับเงิน/ใบกำกับภาษีอย่างย่อ,2=ใบเสร็จรับเงิน/ใบกำกับภาษีอย่างเต็ม
  int? billtaxtype;
  String? canceltime;
  String? canceldatetime;
  String? cancelusercode;
  String? cancelusername;
  String? canceldescription;
  String? cancelreason;

  /// รหัสผู้สร้างเอกสาร
  @JsonKey(name: 'creator_code')
  String? creatorcode;

  /// ชื่อผู้สร้างเอกสาร
  @JsonKey(name: 'creator_name')
  String? creatorname;

  /// วันเวลาที่สร้างเอกสาร (จาก pgsql)
  @JsonKey(name: 'created_at')
  String? createdat;

  /// รหัสผู้แก้ไขเอกสารล่าสุด
  @JsonKey(name: 'modifier_code')
  String? modifiercode;

  /// ชื่อผู้แก้ไขเอกสารล่าสุด
  @JsonKey(name: 'modifier_name')
  String? modifiername;

  /// วันเวลาที่แก้ไขเอกสารล่าสุด (จาก pgsql)
  @JsonKey(name: 'modified_at')
  String? modifiedat;

  String? fullvataddress;
  //เลขสาขาใบกำกับภาษีแบบเต็ม
  String? fullvatbranchnumber;
  //ชื่อลูกค้าใบกำกับภาษีแบบเต็ม
  String? fullvatname;
  //เลขที่ใบกำกับภาษีแบบเต็ม
  String? fullvatdocnumber;
  //ชื่อลูกค้าใบกำกับภาษีแบบเต็ม
  String? fullvattaxid;
  //พิมพ์ใบกำกับภาษีแบบเต็มแล้ว
  bool? fullvatprint;

  bool? isvatregister;

  /// วันที่พิมพ์ใบเสร็จ (สำเนา)
  List<String>? printcopybilldatetime;

  // หมายเลขโต๊ะ
  String? tablenumber;

  String? tableopendatetime;
  String? tableclosedatetime;

  /// จำนวนคน ชาย
  int? mancount;

  /// จำนวนคน หญิง
  int? womancount;

  /// จำนวนเด็ก
  int? childcount;

  //False=สั่งแบบอลาคาร์ทไม่ได้,True=สั่งแบบอลาคาร์ทได้
  bool? tableallacratemode;

  String? buffetcode;

  /// เบอร์โทรลูกค้า (สะสมแต้ม)
  String? customertelephone;

  /// จำนวนชิ้น
  double? totalqty;

  /// ส่วนลดสินค้ามีภาษี
  double? totaldiscountvatamount;

  /// ส่วนลดสินค้ายกเว้นภาษี
  double? totaldiscountexceptvatamount;

  /// ชื่อพนักงาน Cashier
  String? cashiername;

  /// เงินทอน
  double? paycashchange;

  /// ชำระเงินโดย QR Code
  double? sumqrcode;

  /// ชำระเงินโดย Credit Card
  double? sumcreditcard;

  /// ชำระเงินโดยเงินโอน
  double? summoneytransfer;

  /// ชำระเงินโดยเช็ค
  double? sumcheque;

  /// ชำระเงินโดย Coupon
  double? sumcoupon;

  /// สูตรส่วนลดรายการสินค้า (ก่อนคิดเงิน)
  String? detaildiscountformula;
  double? detailtotalamount;
  double? detailtotaldiscount;

  /// ยอดปัดเศษ
  double? roundamount;

  /// ยอดรวมหลังหักส่วนลดท้ายบิล
  double? totalamountafterdiscount;

  // ยอดรวมสินค้าก่อนหักส่วนลดสินค้า
  double? detailtotalamountbeforediscount;

  double? sumcredit;

  /// ช่องทางการขาย
  String? salechannelcode;
  double? salechannelgp;
  int? salechannelgptype;

  /// กลับบ้าน
  int? takeaway;

  /// สาขา
  BranchModel? branch;

  /// มูลค่าตามใบกำกับเดิม
  double? reftotaloriginal;

  /// มูลค่าที่ถูกต้อง
  double? reftotalcorrect;

  /// ผลต่าง
  double? reftotaldiff;

  /// รูป slip
  String? slipurl;

  bool? ispos;

  bool? isbom;

  /// isdelivery
  bool? isdelivery;

  /// มูลค่าขาย delivery
  double? deliveryamount;

  /// ช่องทางการจัดส่ง
  bool? istransport;

  /// รหัสช่องทางการจัดส่ง
  String? transportcode;

  /// มูลค่า ช่องทางการจัดส่ง
  double? transportamount;

  /// แต้มที่ได้
  int? getpoint;

  /// ใช้แต้ม
  int? usepoint;

  /// มูลค่า point ที่เป็นส่วนลด
  int? pointdiscountamount;

  /// จ่ายเป็นแต้ม
  int? paypointamount;

  /// รหัสใช้แต้ม
  String? pointscode;

  /// คูปอง
  List<CouponModel>? coupons;

  /// รวมจำนวนเงินคูปอง
  double? totalcouponamount;

  /// จำนวนเงินส่วนลดคูปอง
  double? coupondiscountamount;

  /// จำนวนเงินสดคูปอง
  double? couponcashamount;

  /// รายการเอกสารอ้างอิง (optional field)
  @JsonKey(includeIfNull: false)
  List<DocReferenceModel>? docreferences;

  /// รายการเอกสารเงินจ่ายล่วงหน้า
  List<AdvancePaymentDocModel>? advancepaymentdocs;

  /// ยอดรวมเงินจ่ายล่วงหน้า
  double? sumadvancepayment;

  /// ยอดรวมเงินมัดจำ
  double? sumdeposit;

  /// รายการเอกสารเงินมัดจำ
  List<DepositDocModel>? depositdocs;

  /// ประเภทการจัดซื้อ (สำหรับ PO Approval)
  String? purchasetypecode;

  /// ชื่อประเภทการจัดซื้อ (หลายภาษา)
  List<LanguageDataModel>? purchasetypenames;

  /// สกุลเงินของเอกสาร (Current Currency) - สกุลเงินที่ใช้บันทึกเอกสาร
  /// เช่น USD, EUR, JPY (สำหรับเอกสารต่างประเทศ)
  String? currency;

  /// สัญลักษณ์สกุลเงินเอกสาร (เช่น $, €, ¥)
  String? currencysymbol;

  /// อัตราแลกเปลี่ยน (1 สกุลเงินเอกสาร = ? สกุลเงินหลัก)
  /// ใช้แปลงยอดเงินเป็นสกุลเงินหลักสำหรับลงบัญชี
  double? exchangerate;

  /// ยอดรวมในสกุลเงินเอกสาร (สำหรับ Multi-Currency)
  /// ใช้สำหรับคำนวณ WHT และภาษีในกรณีที่ใช้สกุลเงินต่างประเทศ
  @JsonKey(name: 'totalamount_doc')
  double? totalamountDoc;

  /// มูลค่ารวมสินค้า (Document Currency)
  @JsonKey(name: 'totalvalue_doc')
  double? totalvalueDoc;

  /// ส่วนลดท้ายบิล (Document Currency)
  @JsonKey(name: 'totaldiscount_doc')
  double? totaldiscountDoc;

  /// มูลค่าภาษี (Document Currency)
  @JsonKey(name: 'totalvatvalue_doc')
  double? totalvatvalueDoc;

  /// มูลค่าก่อนหักภาษี (Document Currency)
  @JsonKey(name: 'totalbeforevat_doc')
  double? totalbeforevatDoc;

  /// มูลค่าหลังหักภาษี (Document Currency)
  @JsonKey(name: 'totalaftervat_doc')
  double? totalaftervatDoc;

  /// สกุลเงินเอกสาร (Document Currency) - สกุลเงินที่ใช้บันทึกเอกสาร
  /// ต่างจากสกุลเงินหลัก (currency) ที่ใช้ลงบัญชี
  @JsonKey(name: 'doc_currency')
  String? docCurrency;

  /// สัญลักษณ์สกุลเงินเอกสาร
  @JsonKey(name: 'doc_currencysymbol')
  String? docCurrencySymbol;

  /// รายการภาษีหัก ณ ที่จ่าย (WHT Entries)
  @JsonKey(name: 'wht_entries')
  List<WHTEntryModel>? whtEntries;

  /// เครดิตเทอม (จำนวนวัน) — 0 = ชำระทันที, >0 = เครดิต N วัน
  int? creditdays;

  /// วันครบกำหนดชำระ (YYYY-MM-DD)
  String? duedate;

  /// สถานะลบแบบ Soft Delete (true = ถูกลบแล้ว)
  bool? isdelete;

  // === PR-specific transient fields (ไม่ serialize — ใช้ extraFields pattern ตอน save) ===
  // ไม่ต้อง build_runner เพราะ .g.dart ไม่รู้จัก fields เหล่านี้
  String? prDepartmentCode;
  String? prDepartmentName;
  String? prJobCode;
  String? prJobName;
  String? prShipToAddress;
  String? prShipToName;
  String? prShipToLat;
  String? prShipToLng;
  String? prCostCenterCode;
  String? prCostCenterName;
  String? prPaymentCondition;
  int? prUrgency;

  // Estimated Landed Cost
  String? prEstimatedUnitCost;
  String? prEstimatedFreight;
  String? prEstimatedDuty;
  String? prEstimatedLandedCost;

  // Vendor Preferences
  String? prPreferredVendor;
  String? prAlternativeVendor;
  String? prReasonPreferred;

  // Approval Enhancement
  String? prApprovalDeadline;

  // PR เพิ่มเติม (session 6)
  String? prInternalNote;
  String? prPurchasingGroup;
  double? prOverDeliveryTolerance;
  double? prUnderDeliveryTolerance;
  String? prBudgetCode;
  double? prBudgetAmount;
  String? prVendorNote;

  // Tracking (read-only from backend)
  String? prConversionStatus;
  String? prRefRfqDocNo;
  String? prRefPoDocNo;

  TransactionModel({
    String? shopid,
    required this.guidref,
    required this.docno,
    required this.docdatetime,
    this.docdatelocal,
    this.doctimelocal,
    this.timezone,
    required this.docrefno,
    required this.docrefdate,
    required this.docreftype,
    required this.doctype,
    required this.vattype,
    required this.custcode,
    required this.salecode,
    required this.salename,
    required this.discountword,
    required this.totalcost,
    required this.totalvalue,
    required this.totaldiscount,
    required this.totalvatvalue,
    required this.totalbeforevat,
    required this.totalaftervat,
    required this.totalexceptvat,
    required this.totalamount,
    required this.cashiercode,
    required this.posid,
    required this.membercode,
    required this.vatrate,
    required this.status,
    this.isref,
    this.islocked,
    this.isclosed,
    this.iscomparedsuccess,
    this.isclosedmanual,
    this.closedmanualByCode,
    this.closedmanualByName,
    this.closedmanualAt,
    this.closedmanualReason,
    this.detailcount,
    List<TransactionDetailModel>? details,
    required this.inquirytype,
    required this.taxdocno,
    required this.taxdocdate,
    required this.transflag,
    required this.iscancel,
    required this.ismanualamount,
    TransactionPayModel? paymentdetail,
    List<LanguageDataModel>? custnames,
    String? guidfixed,
    String? description,
    String? paymentdetailraw,
    List<BillPayStruct>? paymentStructs,
    double? paycashamount,
    int? billtaxtype,
    String? canceltime,
    String? canceldatetime,
    String? cancelusercode,
    String? cancelusername,
    String? canceldescription,
    String? cancelreason,
    String? creatorcode,
    String? creatorname,
    String? createdat,
    String? modifiercode,
    String? modifiername,
    String? modifiedat,
    String? fullvataddress,
    String? fullvatbranchnumber,
    String? fullvatname,
    String? fullvatdocnumber,
    String? fullvattaxid,
    bool? fullvatprint,
    bool? isvatregister,
    List<String>? printcopybilldatetime,
    String? tablenumber,
    String? tableopendatetime,
    String? tableclosedatetime,
    int? mancount,
    int? womancount,
    int? childcount,
    bool? tableallacratemode,
    String? buffetcode,
    String? customertelephone,
    double? totalqty,
    double? totaldiscountvatamount,
    double? totaldiscountexceptvatamount,
    String? cashiername,
    double? paycashchange,
    double? sumqrcode,
    double? sumcreditcard,
    double? summoneytransfer,
    double? sumcheque,
    double? sumcoupon,
    String? detaildiscountformula,
    double? detailtotalamount,
    double? detailtotaldiscount,
    double? roundamount,
    double? totalamountafterdiscount,
    double? detailtotalamountbeforediscount,
    double? sumcredit,
    String? salechannelcode,
    double? salechannelgp,
    int? salechannelgptype,
    int? takeaway,
    BranchModel? branch,
    double? reftotaloriginal,
    double? reftotalcorrect,
    double? reftotaldiff,
    String? slipurl,
    bool? ispos,
    bool? isbom,
    bool? isdelivery,
    double? deliveryamount,
    bool? istransport,
    String? transportcode,
    double? transportamount,
    int? getpoint,
    int? usepoint,
    int? pointdiscountamount,
    int? paypointamount,
    String? pointscode,
    List<CouponModel>? coupons,
    double? totalcouponamount,
    double? coupondiscountamount,
    double? couponcashamount,
    List<DocReferenceModel>? docreferences,
    List<AdvancePaymentDocModel>? advancepaymentdocs,
    double? sumadvancepayment,
    double? sumdeposit,
    List<DepositDocModel>? depositdocs,
    String? purchasetypecode,
    List<LanguageDataModel>? purchasetypenames,
    int? creditdays,
    String? duedate,
  }) : shopid = shopid ?? "",
       paymentdetail = paymentdetail ?? TransactionPayModel(),
       custnames = custnames ?? <LanguageDataModel>[],
       guidfixed = guidfixed ?? "",
       description = description ?? "",
       paymentdetailraw = paymentdetailraw ?? "",
       paymentStructs = paymentStructs ?? <BillPayStruct>[],
       paycashamount = paycashamount ?? 0,
       billtaxtype = billtaxtype ?? 0,
       canceltime = canceltime,
       canceldatetime = canceldatetime ?? "",
       cancelusercode = cancelusercode ?? "",
       cancelusername = cancelusername ?? "",
       canceldescription = canceldescription ?? "",
       cancelreason = cancelreason ?? "",
       creatorcode = creatorcode ?? "",
       creatorname = creatorname ?? "",
       createdat = createdat ?? "",
       modifiercode = modifiercode ?? "",
       modifiername = modifiername ?? "",
       modifiedat = modifiedat ?? "",
       fullvataddress = fullvataddress ?? "",
       fullvatbranchnumber = fullvatbranchnumber ?? "",
       fullvatname = fullvatname ?? "",
       fullvatdocnumber = fullvatdocnumber ?? "",
       fullvattaxid = fullvattaxid ?? "",
       fullvatprint = fullvatprint ?? false,
       isvatregister = isvatregister ?? false,
       printcopybilldatetime = printcopybilldatetime ?? <String>[],
       tablenumber = tablenumber ?? "",
       tableopendatetime = tableopendatetime ?? "",
       tableclosedatetime = tableclosedatetime ?? "",
       mancount = mancount ?? 0,
       womancount = womancount ?? 0,
       childcount = childcount ?? 0,
       tableallacratemode = tableallacratemode ?? false,
       buffetcode = buffetcode ?? "",
       customertelephone = customertelephone ?? "",
       totalqty = totalqty ?? 0,
       totaldiscountvatamount = totaldiscountvatamount ?? 0,
       totaldiscountexceptvatamount = totaldiscountexceptvatamount ?? 0,
       cashiername = cashiername ?? "",
       paycashchange = paycashchange ?? 0,
       sumqrcode = sumqrcode ?? 0,
       sumcreditcard = sumcreditcard ?? 0,
       summoneytransfer = summoneytransfer ?? 0,
       sumcheque = sumcheque ?? 0,
       sumcoupon = sumcoupon ?? 0,
       detaildiscountformula = detaildiscountformula ?? "",
       detailtotalamount = detailtotalamount ?? 0,
       detailtotaldiscount = detailtotaldiscount ?? 0,
       roundamount = roundamount ?? 0,
       totalamountafterdiscount = totalamountafterdiscount ?? 0,
       detailtotalamountbeforediscount = detailtotalamountbeforediscount ?? 0,
       sumcredit = sumcredit ?? 0,
       salechannelcode = salechannelcode ?? "",
       salechannelgp = salechannelgp ?? 0,
       salechannelgptype = salechannelgptype ?? 0,
       takeaway = takeaway ?? 0,
       details = details ?? <TransactionDetailModel>[],
       branch = branch ?? BranchModel(),
       reftotaloriginal = reftotaloriginal ?? 0,
       reftotalcorrect = reftotalcorrect ?? 0,
       reftotaldiff = reftotaldiff ?? 0,
       slipurl = slipurl ?? "",
       ispos = ispos ?? false,
       isbom = isbom ?? false,
       isdelivery = isdelivery ?? false,
       deliveryamount = deliveryamount ?? 0,
       istransport = istransport ?? false,
       transportcode = transportcode ?? "",
       transportamount = transportamount ?? 0,
       getpoint = getpoint ?? 0,
       usepoint = usepoint ?? 0,
       pointdiscountamount = pointdiscountamount ?? 0,
       paypointamount = paypointamount ?? 0,
       pointscode = pointscode ?? "",
       coupons = coupons ?? <CouponModel>[],
       totalcouponamount = totalcouponamount ?? 0,
       coupondiscountamount = coupondiscountamount ?? 0,
       couponcashamount = couponcashamount ?? 0,
       docreferences = docreferences ?? <DocReferenceModel>[],
       advancepaymentdocs = advancepaymentdocs ?? <AdvancePaymentDocModel>[],
       sumadvancepayment = sumadvancepayment ?? 0,
       sumdeposit = sumdeposit ?? 0,
       depositdocs = depositdocs ?? <DepositDocModel>[],
       purchasetypecode = purchasetypecode ?? "",
       purchasetypenames = purchasetypenames ?? <LanguageDataModel>[],
       creditdays = creditdays ?? 0,
       duedate = duedate ?? "";

  factory TransactionModel.fromJson(Map<String, dynamic> json) =>
      _$TransactionModelFromJson(json);

  Map<String, dynamic> toJson() => _$TransactionModelToJson(this);

  // Add the copyWith method here
  TransactionModel copyWith({
    String? guidfixed,
    String? shopid,
    String? guidref,
    String? docno,
    String? docdatetime,
    String? docrefno,
    String? docrefdate,
    int? docreftype,
    int? doctype,
    int? vattype,
    int? inquirytype,
    int? transflag,
    String? custcode,
    List<LanguageDataModel>? custnames,
    String? salecode,
    String? salename,
    String? discountword,
    double? totalcost,
    double? totalvalue,
    double? totaldiscount,
    double? totalvatvalue,
    double? totalbeforevat,
    double? totalaftervat,
    double? totalexceptvat,
    double? totalamount,
    String? cashiercode,
    String? posid,
    String? membercode,
    bool? isref,
    bool? islocked,
    bool? isclosed,
    int? iscomparedsuccess,
    bool? isclosedmanual,
    String? closedmanualByCode,
    String? closedmanualByName,
    String? closedmanualAt,
    String? closedmanualReason,
    int? detailcount,
    bool? ismanualamount,
    String? taxdocno,
    String? taxdocdate,
    double? vatrate,
    bool? iscancel,
    int? status,
    List<TransactionDetailModel>? details,
    TransactionPayModel? paymentdetail,
    String? paymentdetailraw,
    List<BillPayStruct>? paymentStructs,
    double? paycashamount,
    String? description,
    int? billtaxtype,
    String? canceltime,
    String? canceldatetime,
    String? cancelusercode,
    String? cancelusername,
    String? canceldescription,
    String? cancelreason,
    String? creatorcode,
    String? creatorname,
    String? createdat,
    String? modifiercode,
    String? modifiername,
    String? modifiedat,
    String? fullvataddress,
    String? fullvatbranchnumber,
    String? fullvatname,
    String? fullvatdocnumber,
    String? fullvattaxid,
    bool? fullvatprint,
    bool? isvatregister,
    List<String>? printcopybilldatetime,
    String? tablenumber,
    String? tableopendatetime,
    String? tableclosedatetime,
    int? mancount,
    int? womancount,
    int? childcount,
    bool? tableallacratemode,
    String? buffetcode,
    String? customertelephone,
    double? totalqty,
    double? totaldiscountvatamount,
    double? totaldiscountexceptvatamount,
    String? cashiername,
    double? paycashchange,
    double? sumqrcode,
    double? sumcreditcard,
    double? summoneytransfer,
    double? sumcheque,
    double? sumcoupon,
    String? detaildiscountformula,
    double? detailtotalamount,
    double? detailtotaldiscount,
    double? roundamount,
    double? totalamountafterdiscount,
    double? detailtotalamountbeforediscount,
    double? sumcredit,
    String? salechannelcode,
    double? salechannelgp,
    int? salechannelgptype,
    int? takeaway,
    BranchModel? branch,
    double? reftotaloriginal,
    double? reftotalcorrect,
    double? reftotaldiff,
    String? slipurl,
    bool? ispos,
    bool? isbom,
    bool? isdelivery,
    double? deliveryamount,
    bool? istransport,
    String? transportcode,
    double? transportamount,
    int? getpoint,
    int? usepoint,
    int? pointdiscountamount,
    int? paypointamount,
    String? pointscode,
    List<CouponModel>? coupons,
    double? totalcouponamount,
    double? coupondiscountamount,
    double? couponcashamount,
    List<DocReferenceModel>? docreferences,
    List<AdvancePaymentDocModel>? advancepaymentdocs,
    double? sumadvancepayment,
    double? sumdeposit,
    List<DepositDocModel>? depositdocs,
  }) {
    return TransactionModel(
      guidfixed: guidfixed ?? this.guidfixed,
      shopid: shopid ?? this.shopid,
      guidref: guidref ?? this.guidref,
      docno: docno ?? this.docno,
      docdatetime: docdatetime ?? this.docdatetime,
      docrefno: docrefno ?? this.docrefno,
      docrefdate: docrefdate ?? this.docrefdate,
      docreftype: docreftype ?? this.docreftype,
      doctype: doctype ?? this.doctype,
      vattype: vattype ?? this.vattype,
      inquirytype: inquirytype ?? this.inquirytype,
      transflag: transflag ?? this.transflag,
      custcode: custcode ?? this.custcode,
      custnames: custnames ?? this.custnames,
      salecode: salecode ?? this.salecode,
      salename: salename ?? this.salename,
      discountword: discountword ?? this.discountword,
      totalcost: totalcost ?? this.totalcost,
      totalvalue: totalvalue ?? this.totalvalue,
      totaldiscount: totaldiscount ?? this.totaldiscount,
      totalvatvalue: totalvatvalue ?? this.totalvatvalue,
      totalbeforevat: totalbeforevat ?? this.totalbeforevat,
      totalaftervat: totalaftervat ?? this.totalaftervat,
      totalexceptvat: totalexceptvat ?? this.totalexceptvat,
      totalamount: totalamount ?? this.totalamount,
      cashiercode: cashiercode ?? this.cashiercode,
      posid: posid ?? this.posid,
      membercode: membercode ?? this.membercode,
      ismanualamount: ismanualamount ?? this.ismanualamount,
      taxdocno: taxdocno ?? this.taxdocno,
      taxdocdate: taxdocdate ?? this.taxdocdate,
      vatrate: vatrate ?? this.vatrate,
      iscancel: iscancel ?? this.iscancel,
      isref: isref ?? this.isref,
      islocked: islocked ?? this.islocked,
      isclosed: isclosed ?? this.isclosed,
      iscomparedsuccess: iscomparedsuccess ?? this.iscomparedsuccess,
      isclosedmanual: isclosedmanual ?? this.isclosedmanual,
      closedmanualByCode: closedmanualByCode ?? this.closedmanualByCode,
      closedmanualByName: closedmanualByName ?? this.closedmanualByName,
      closedmanualAt: closedmanualAt ?? this.closedmanualAt,
      closedmanualReason: closedmanualReason ?? this.closedmanualReason,
      detailcount: detailcount ?? this.detailcount,
      status: status ?? this.status,
      details: details ?? this.details,
      paymentdetail: paymentdetail ?? this.paymentdetail,
      paymentdetailraw: paymentdetailraw ?? this.paymentdetailraw,
      paymentStructs:
          paymentStructs ?? this.paymentStructs,
      paycashamount: paycashamount ?? this.paycashamount,
      description: description ?? this.description,
      billtaxtype: billtaxtype ?? this.billtaxtype,
      canceltime: canceltime ?? this.canceltime,
      canceldatetime: canceldatetime ?? this.canceldatetime,
      cancelusercode: cancelusercode ?? this.cancelusercode,
      cancelusername: cancelusername ?? this.cancelusername,
      canceldescription: canceldescription ?? this.canceldescription,
      cancelreason: cancelreason ?? this.cancelreason,
      creatorcode: creatorcode ?? this.creatorcode,
      creatorname: creatorname ?? this.creatorname,
      createdat: createdat ?? this.createdat,
      modifiercode: modifiercode ?? this.modifiercode,
      modifiername: modifiername ?? this.modifiername,
      modifiedat: modifiedat ?? this.modifiedat,
      fullvataddress: fullvataddress ?? this.fullvataddress,
      fullvatbranchnumber: fullvatbranchnumber ?? this.fullvatbranchnumber,
      fullvatname: fullvatname ?? this.fullvatname,
      fullvatdocnumber: fullvatdocnumber ?? this.fullvatdocnumber,
      fullvattaxid: fullvattaxid ?? this.fullvattaxid,
      fullvatprint: fullvatprint ?? this.fullvatprint,
      isvatregister: isvatregister ?? this.isvatregister,
      printcopybilldatetime:
          printcopybilldatetime ?? this.printcopybilldatetime,
      tablenumber: tablenumber ?? this.tablenumber,
      tableopendatetime: tableopendatetime ?? this.tableopendatetime,
      tableclosedatetime: tableclosedatetime ?? this.tableclosedatetime,
      mancount: mancount ?? this.mancount,
      womancount: womancount ?? this.womancount,
      childcount: childcount ?? this.childcount,
      tableallacratemode: tableallacratemode ?? this.tableallacratemode,
      buffetcode: buffetcode ?? this.buffetcode,
      customertelephone: customertelephone ?? this.customertelephone,
      totalqty: totalqty ?? this.totalqty,
      totaldiscountvatamount:
          totaldiscountvatamount ?? this.totaldiscountvatamount,
      totaldiscountexceptvatamount:
          totaldiscountexceptvatamount ?? this.totaldiscountexceptvatamount,
      cashiername: cashiername ?? this.cashiername,
      paycashchange: paycashchange ?? this.paycashchange,
      sumqrcode: sumqrcode ?? this.sumqrcode,
      sumcreditcard: sumcreditcard ?? this.sumcreditcard,
      summoneytransfer: summoneytransfer ?? this.summoneytransfer,
      sumcheque: sumcheque ?? this.sumcheque,
      sumcoupon: sumcoupon ?? this.sumcoupon,
      detaildiscountformula:
          detaildiscountformula ?? this.detaildiscountformula,
      detailtotalamount: detailtotalamount ?? this.detailtotalamount,
      detailtotaldiscount: detailtotaldiscount ?? this.detailtotaldiscount,
      roundamount: roundamount ?? this.roundamount,
      totalamountafterdiscount:
          totalamountafterdiscount ?? this.totalamountafterdiscount,
      detailtotalamountbeforediscount:
          detailtotalamountbeforediscount ??
          this.detailtotalamountbeforediscount,
      sumcredit: sumcredit ?? this.sumcredit,
      salechannelcode: salechannelcode ?? this.salechannelcode,
      salechannelgp: salechannelgp ?? this.salechannelgp,
      salechannelgptype: salechannelgptype ?? this.salechannelgptype,
      takeaway: takeaway ?? this.takeaway,
      branch: branch ?? this.branch,
      reftotaloriginal: reftotaloriginal ?? this.reftotaloriginal,
      reftotalcorrect: reftotalcorrect ?? this.reftotalcorrect,
      reftotaldiff: reftotaldiff ?? this.reftotaldiff,
      slipurl: slipurl ?? this.slipurl,
      ispos: ispos ?? this.ispos,
      isbom: isbom ?? this.isbom,
      isdelivery: isdelivery ?? this.isdelivery,
      deliveryamount: deliveryamount ?? this.deliveryamount,
      istransport: istransport ?? this.istransport,
      transportcode: transportcode ?? this.transportcode,
      transportamount: transportamount ?? 0,
      getpoint: getpoint ?? 0,
      usepoint: usepoint ?? 0,
      pointdiscountamount: pointdiscountamount ?? 0,
      paypointamount: paypointamount ?? 0,
      pointscode: pointscode ?? "",
      coupons: coupons ?? this.coupons,
      totalcouponamount: totalcouponamount ?? this.totalcouponamount,
      coupondiscountamount: coupondiscountamount ?? this.coupondiscountamount,
      couponcashamount: couponcashamount ?? this.couponcashamount,
      docreferences: docreferences ?? this.docreferences,
      advancepaymentdocs: advancepaymentdocs ?? this.advancepaymentdocs,
      sumadvancepayment: sumadvancepayment ?? this.sumadvancepayment,
      sumdeposit: sumdeposit ?? this.sumdeposit,
      depositdocs: depositdocs ?? this.depositdocs,
    );
  }
}

@JsonSerializable(explicitToJson: true)
class TransactionDetailModel {
  late String docdatetime;
  late String itemguid;
  late String barcode;
  late String itemcode;
  String? docref;
  String? docrefdatetime;
  List<LanguageDataModel>? itemnames;
  late String unitcode;
  late double qty;
  late double price;
  late String discount;
  late double sumofcost;
  late double sumamount; //รวมหักส่วนลด
  late String remark;
  late int linenumber;
  late String whcode;
  List<LanguageDataModel>? whnames;
  late String locationcode;
  List<LanguageDataModel>? locationnames;
  late String? towhcode;
  List<LanguageDataModel>? towhnames;
  late String? tolocationcode;
  List<LanguageDataModel>? tolocationnames;
  late String shelfcode;
  late double totalqty;
  late double discountamount;
  late double averagecost;

  late double totalvaluevat; //มูลค่าภาษี
  late double sumamountexcludevat; //รวมมูลค่าก่อนvat
  late double priceexcludevat; //ราคาก่อนภาษี

  late double standvalue;
  late double dividevalue;
  late int calcflag; //ขาออก -1 เขาเข้า 1
  late int vattype; //
  late int ispos;
  late int laststatus;
  late int itemtype; //สินค้าชุด
  late int taxtype; //ประเภทภาษี
  int? vatcal; //ประเภทภาษี
  late int inquirytype;
  bool? multiunit;
  List<LanguageDataModel>? unitnames;

  /// บาร์โค้ดอ้างอิง
  List<BarCodeSubModel>? refbarcodes;

  /// SKU สินค้า
  String? sku;
  String? extrajson;
  List<ExtraJsonListModel>? extrajsonlist;

  /// ผู้ผลิต
  String? manufacturerguid;

  String? description;
  String? imageuri;

  /// เฉพาะสำหรับ purchase - จำนวนเหตุการณ์
  double? eventqty;

  /// ราคาในสกุลเงินเอกสาร (Document Currency)
  /// ใช้เมื่อ multi-currency — ผู้ใช้ป้อนราคาเป็นสกุลเงินเอกสาร
  /// null = เอกสารเก่า หรือไม่ใช่ multi-currency (ใช้ price ตรงๆ)
  @JsonKey(name: 'price_doc')
  double? priceDoc;

  /// ยอดรวมในสกุลเงินเอกสาร (Document Currency)
  /// คำนวณจาก (qty × priceDoc) - discountAmountDoc
  @JsonKey(name: 'sumamount_doc')
  double? sumAmountDoc;

  /// ยอดส่วนลดในสกุลเงินเอกสาร
  @JsonKey(name: 'discountamount_doc')
  double? discountamountDoc;

  /// ราคาก่อน VAT ในสกุลเงินเอกสาร
  @JsonKey(name: 'priceexcludevat_doc')
  double? priceexcludevatDoc;

  /// ยอดรวมก่อน VAT ในสกุลเงินเอกสาร
  @JsonKey(name: 'sumamountexcludevat_doc')
  double? sumamountexcludevatDoc;

  /// มูลค่าภาษีในสกุลเงินเอกสาร
  @JsonKey(name: 'totalvaluevat_doc')
  double? totalvaluevatDoc;

  TransactionDetailModel({
    required this.docdatetime,
    required this.itemguid,
    required this.barcode,
    required this.itemcode,
    List<LanguageDataModel>? itemnames,
    required this.unitcode,
    required this.qty,
    required this.price,
    required this.discount,
    required this.sumofcost,
    required this.remark,
    required this.linenumber,
    required this.whcode,
    required this.shelfcode,
    required this.sumamount,
    required this.locationcode,
    required this.totalvaluevat,
    required this.totalqty,
    required this.sumamountexcludevat,
    required this.priceexcludevat,
    required this.discountamount,
    required this.standvalue,
    required this.dividevalue,
    required this.calcflag,
    required this.vattype,
    required this.averagecost,
    required this.ispos,
    required this.laststatus,
    required this.itemtype,
    required this.taxtype,
    required this.inquirytype,
    bool? multiunit,
    List<LanguageDataModel>? unitnames,
    List<LanguageDataModel>? whnames,
    List<LanguageDataModel>? locationnames,
    List<LanguageDataModel>? towhnames,
    List<LanguageDataModel>? tolocationnames,
    String? towhcode,
    String? tolocationcode,
    String? docref,
    String? docrefdatetime,
    int? vatcal,

    /// บาร์โค้ดอ้างอิง
    List<BarCodeSubModel>? refbarcodes,
    String? sku,
    String? extrajson,
    List<ExtraJsonListModel>? extrajsonlist,
    String? manufacturerguid,
    String? description,
    String? imageuri,
    double? eventqty,
    this.priceDoc,
    this.sumAmountDoc,
    this.discountamountDoc,
    this.priceexcludevatDoc,
    this.sumamountexcludevatDoc,
    this.totalvaluevatDoc,
  }) : multiunit = multiunit ?? false,
       unitnames = unitnames ?? <LanguageDataModel>[],
       whnames = whnames ?? <LanguageDataModel>[],
       locationnames = locationnames ?? <LanguageDataModel>[],
       towhnames = towhnames ?? <LanguageDataModel>[],
       tolocationnames = tolocationnames ?? <LanguageDataModel>[],
       refbarcodes = refbarcodes ?? <BarCodeSubModel>[],
       towhcode = towhcode ?? "",
       tolocationcode = tolocationcode ?? "",
       itemnames = itemnames ?? <LanguageDataModel>[],
       vatcal = vatcal ?? 0,
       docref = docref ?? "",
       docrefdatetime =
           docrefdatetime ?? DateTime.now().toLocal().toIso8601String(),
       sku = sku ?? "",
       extrajson = extrajson ?? "",
       extrajsonlist = extrajsonlist ?? <ExtraJsonListModel>[],
       manufacturerguid = manufacturerguid ?? "",
       description = description ?? "",
       imageuri = imageuri ?? "",
       eventqty = eventqty;

  factory TransactionDetailModel.fromJson(Map<String, dynamic> json) =>
      _$TransactionDetailModelFromJson(json);

  Map<String, dynamic> toJson() => _$TransactionDetailModelToJson(this);

  TransactionDetailModel copyWith({
    String? docdatetime,
    String? itemguid,
    String? barcode,
    String? itemcode,
    String? docref,
    String? docrefdatetime,
    List<LanguageDataModel>? itemnames,
    String? unitcode,
    double? qty,
    double? price,
    String? discount,
    double? sumofcost,
    double? sumamount,
    String? remark,
    int? linenumber,
    String? whcode,
    List<LanguageDataModel>? whnames,
    String? locationcode,
    List<LanguageDataModel>? locationnames,
    String? towhcode,
    List<LanguageDataModel>? towhnames,
    String? tolocationcode,
    List<LanguageDataModel>? tolocationnames,
    String? shelfcode,
    double? totalqty,
    double? discountamount,
    double? averagecost,
    double? totalvaluevat,
    double? sumamountexcludevat,
    double? priceexcludevat,
    double? standvalue,
    double? dividevalue,
    int? calcflag,
    int? vattype,
    int? ispos,
    int? laststatus,
    int? itemtype,
    int? taxtype,
    int? vatcal,
    int? inquirytype,
    bool? multiunit,
    List<LanguageDataModel>? unitnames,
    List<BarCodeSubModel>? refbarcodes,
    String? sku,
    String? extrajson,
    List<ExtraJsonListModel>? extrajsonlist,
    String? manufacturerguid,
    String? description,
    String? imageuri,
    double? eventqty,
    double? priceDoc,
    double? sumAmountDoc,
    double? discountamountDoc,
    double? priceexcludevatDoc,
    double? sumamountexcludevatDoc,
    double? totalvaluevatDoc,
  }) {
    return TransactionDetailModel(
      docdatetime: docdatetime ?? this.docdatetime,
      itemguid: itemguid ?? this.itemguid,
      barcode: barcode ?? this.barcode,
      itemcode: itemcode ?? this.itemcode,
      itemnames: itemnames ?? this.itemnames,
      unitcode: unitcode ?? this.unitcode,
      qty: qty ?? this.qty,
      price: price ?? this.price,
      discount: discount ?? this.discount,
      sumofcost: sumofcost ?? this.sumofcost,
      remark: remark ?? this.remark,
      linenumber: linenumber ?? this.linenumber,
      whcode: whcode ?? this.whcode,
      shelfcode: shelfcode ?? this.shelfcode,
      sumamount: sumamount ?? this.sumamount,
      locationcode: locationcode ?? this.locationcode,
      totalvaluevat: totalvaluevat ?? this.totalvaluevat,
      totalqty: totalqty ?? this.totalqty,
      sumamountexcludevat: sumamountexcludevat ?? this.sumamountexcludevat,
      priceexcludevat: priceexcludevat ?? this.priceexcludevat,
      discountamount: discountamount ?? this.discountamount,
      standvalue: standvalue ?? this.standvalue,
      dividevalue: dividevalue ?? this.dividevalue,
      calcflag: calcflag ?? this.calcflag,
      vattype: vattype ?? this.vattype,
      averagecost: averagecost ?? this.averagecost,
      ispos: ispos ?? this.ispos,
      laststatus: laststatus ?? this.laststatus,
      itemtype: itemtype ?? this.itemtype,
      taxtype: taxtype ?? this.taxtype,
      inquirytype: inquirytype ?? this.inquirytype,
      multiunit: multiunit ?? this.multiunit,
      unitnames: unitnames ?? this.unitnames,
      whnames: whnames ?? this.whnames,
      locationnames: locationnames ?? this.locationnames,
      towhnames: towhnames ?? <LanguageDataModel>[],
      tolocationnames: tolocationnames ?? this.tolocationnames,
      towhcode: towhcode ?? this.towhcode,
      tolocationcode: tolocationcode ?? this.tolocationcode,
      docref: docref ?? this.docref,
      docrefdatetime: docrefdatetime ?? this.docrefdatetime,
      vatcal: vatcal ?? this.vatcal,
      refbarcodes: refbarcodes ?? this.refbarcodes,
      sku: sku ?? this.sku,
      extrajson: extrajson ?? this.extrajson,
      extrajsonlist: extrajsonlist ?? this.extrajsonlist,
      manufacturerguid: manufacturerguid ?? this.manufacturerguid,
      description: description ?? this.description,
      imageuri: imageuri ?? this.imageuri,
      eventqty: eventqty ?? this.eventqty,
      priceDoc: priceDoc ?? this.priceDoc,
      sumAmountDoc: sumAmountDoc ?? this.sumAmountDoc,
      discountamountDoc: discountamountDoc ?? this.discountamountDoc,
      priceexcludevatDoc: priceexcludevatDoc ?? this.priceexcludevatDoc,
      sumamountexcludevatDoc: sumamountexcludevatDoc ?? this.sumamountexcludevatDoc,
      totalvaluevatDoc: totalvaluevatDoc ?? this.totalvaluevatDoc,
    );
  }
}

@JsonSerializable(explicitToJson: true)
class MongoTransactionModel {
  late String collection;
  late String keyid;
  late TransactionModel body;

  MongoTransactionModel({
    required this.collection,
    required this.keyid,
    required this.body,
  });

  factory MongoTransactionModel.fromJson(Map<String, dynamic> json) =>
      _$MongoTransactionModelFromJson(json);

  Map<String, dynamic> toJson() => _$MongoTransactionModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class TransactionPayModel {
  String? cashamounttext;
  double? cashamount; // ยอดชำระเงินสด
  String? discountformula; // สูตรส่วนลด
  double? discountamount; // ยอดส่วนลด
  double? totalafterdiscount; // ยอดรวมหลังหักส่วนลด
  double? roundamount; // ยอดปัดเศษ
  double? totalafterround; // ยอดรวมหลังหักส่วนลดและปัดเศษ
  double? creditamount; // ยอดเงินเชื่อม
  List<PayCreditCardModel>? paymentcreditcards; // บัตรเครดิต
  List<PayTransferModel>? paymenttransfers; // เงินโอน

  TransactionPayModel({
    String? cashamounttext,
    double? cashamount, // ยอดชำระเงินสด
    String? discountformula, // สูตรส่วนลด
    double? discountamount, // ยอดส่วนลด
    double? totalafterdiscount, // ยอดรวมหลังหักส่วนลด
    double? roundamount, // ยอดปัดเศษ
    double? totalafterround, // ยอดรวมหลังหักส่วนลดและปัดเศษ
    double? creditamount, // ยอดเงินเชื่อม
    List<PayCreditCardModel>? paymentcreditcards = const [],
    List<PayTransferModel>? paymenttransfers = const [],
  }) : cashamounttext = cashamounttext ?? '',
       cashamount = cashamount ?? 0,
       discountformula = discountformula ?? '',
       discountamount = discountamount ?? 0,
       totalafterdiscount = totalafterdiscount ?? 0,
       roundamount = roundamount ?? 0,
       totalafterround = totalafterround ?? 0,
       creditamount = creditamount ?? 0,
       paymentcreditcards = paymentcreditcards ?? <PayCreditCardModel>[],
       paymenttransfers = paymenttransfers ?? <PayTransferModel>[];

  factory TransactionPayModel.fromJson(Map<String, dynamic> json) =>
      _$TransactionPayModelFromJson(json);
  Map<String, dynamic> toJson() => _$TransactionPayModelToJson(this);

  TransactionPayModel copyWith({
    String? cashamounttext,
    double? cashamount,
    String? discountformula,
    double? discountamount,
    double? totalafterdiscount,
    double? roundamount,
    double? totalafterround,
    double? creditamount,
    List<PayCreditCardModel>? paymentcreditcards,
    List<PayTransferModel>? paymenttransfers,
  }) {
    return TransactionPayModel(
      cashamounttext: cashamounttext ?? this.cashamounttext,
      cashamount: cashamount ?? this.cashamount,
      discountformula: discountformula ?? this.discountformula,
      discountamount: discountamount ?? this.discountamount,
      totalafterdiscount: totalafterdiscount ?? this.totalafterdiscount,
      roundamount: roundamount ?? this.roundamount,
      totalafterround: totalafterround ?? this.totalafterround,
      creditamount: creditamount ?? this.creditamount,
      paymentcreditcards: paymentcreditcards ?? this.paymentcreditcards,
      paymenttransfers: paymenttransfers ?? this.paymenttransfers,
    );
  }
}

@JsonSerializable(explicitToJson: true)
class PayCouponModel {
  String number; // เลขที่คูปอง
  String description; // รายละเอียด
  double amount; // จำนวนเงิน

  PayCouponModel({
    required this.number,
    required this.description,
    required this.amount,
  });

  factory PayCouponModel.fromJson(Map<String, dynamic> json) =>
      _$PayCouponModelFromJson(json);
  Map<String, dynamic> toJson() => _$PayCouponModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PayCashModel {
  String walletid; // รหัส
  String amount; // จำนวนเงิน

  PayCashModel({required this.walletid, required this.amount});

  factory PayCashModel.fromJson(Map<String, dynamic> json) =>
      _$PayCashModelFromJson(json);
  Map<String, dynamic> toJson() => _$PayCashModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PayCreditCardModel {
  double? amount; // จำนวนเงิน
  String? cardnumber; // เลขที่บัตรเครดิต
  double? chargevalue; // ยอด ค่าธรรมเนียม
  String? chargeword; // ค่าธรรมเนียม
  String? docdatetime; // วันที่
  double? totalnetworth; // ยอดเงินสุทธิ

  PayCreditCardModel({
    double? amount,
    String? cardnumber,
    double? chargevalue,
    String? chargeword,
    String? docdatetime,
    double? totalnetworth,
  }) : cardnumber = cardnumber ?? "",
       amount = amount ?? 0,
       chargevalue = chargevalue ?? 0,
       chargeword = chargeword ?? "",
       docdatetime = docdatetime ?? "",
       totalnetworth = totalnetworth ?? 0;

  factory PayCreditCardModel.fromJson(Map<String, dynamic> json) =>
      _$PayCreditCardModelFromJson(json);
  Map<String, dynamic> toJson() => _$PayCreditCardModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PayTransferModel {
  String? docdatetime;
  String? bankcode; // รหัสธนาคาร
  String? bookbankcode; // รหัสบัญชี
  List<LanguageDataModel>? banknames; // ธนาคาร
  String? accountnumber; // เลขที่บัญชี
  double? amount; // จำนวนเงิน
  String? frombankcode; // จากธนาคาร

  PayTransferModel({
    String? docdatetime,
    String? bankcode,
    String? bookbankcode,
    List<LanguageDataModel>? banknames,
    double? amount,
    String? accountnumber,
    String? frombankcode,
  }) : docdatetime = docdatetime ?? "",
       bankcode = bankcode ?? "",
       bookbankcode = bookbankcode ?? "",
       accountnumber = accountnumber ?? "",
       amount = amount ?? 0,
       banknames = banknames ?? <LanguageDataModel>[],
       frombankcode = frombankcode ?? "";

  factory PayTransferModel.fromJson(Map<String, dynamic> json) =>
      _$PayTransferModelFromJson(json);
  Map<String, dynamic> toJson() => _$PayTransferModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PayChequeModel {
  DateTime duedate; // วันที่สั่งจ่ายบนเช็ค
  String bankcode; // รหัสธนาคาร
  String bankname; // ธนาคาร
  String branchnumber; // สาขาธนาคาร
  String chequenumber; // เลขที่เช็ค
  double amount; // จำนวนเงิน

  PayChequeModel({
    required this.duedate,
    required this.bankcode,
    required this.bankname,
    required this.branchnumber,
    required this.chequenumber,
    required this.amount,
  });

  factory PayChequeModel.fromJson(Map<String, dynamic> json) =>
      _$PayChequeModelFromJson(json);
  Map<String, dynamic> toJson() => _$PayChequeModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PayDiscountModel {
  String code; // รหัสส่วนลด
  String description; // รายละเอียด (เพิ่มเติม)
  String formula; // สูตร
  double amount; // มูลค่าส่วนลด

  PayDiscountModel({
    required this.code,
    required this.description,
    required this.formula,
    required this.amount,
  });

  factory PayDiscountModel.fromJson(Map<String, dynamic> json) =>
      _$PayDiscountModelFromJson(json);
  Map<String, dynamic> toJson() => _$PayDiscountModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PayQrModel {
  String providercode; // รหัสกระเป๋า เจ้าของเงิน (Provider)
  String providername; // เจ้าของเงิน (Provider)
  String description; // รายละเอียด (อื่นๆ)
  double amount; // จำนวนเงิน

  PayQrModel({
    this.providercode = "",
    this.providername = "",
    this.description = "",
    required this.amount,
  });

  factory PayQrModel.fromJson(Map<String, dynamic> json) =>
      _$PayQrModelFromJson(json);
  Map<String, dynamic> toJson() => _$PayQrModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class TransactionPaidPayModel {
  String? guidfixed;
  String? shopid;
  String docno;
  String docdatetime;
  int doctype;
  int transflag;
  String custcode;
  List<LanguageDataModel>? custnames;
  String salecode;
  String salename;
  double totalpaymentamount;
  double totalamount;
  double totalbalance;
  double totalvalue;
  List<TransactionPaidPayDetailModel>? details;
  String paymentdetailraw;
  List<BillPayStruct>? paymentStructs;
  double? paycashamount;
  double? sumqrcode;
  double? sumcreditcard;
  double? summoneytransfer;
  double? sumcheque;
  double? sumcoupon;
  double? sumcredit;
  double? roundamount;
  double? paycashchange;

  /// สาขา
  BranchModel? branch;

  TransactionPaidPayModel({
    String? shopid,
    required this.docno,
    required this.docdatetime,
    required this.doctype,
    required this.custcode,
    required this.salecode,
    required this.salename,
    required this.totalpaymentamount,
    required this.totalamount,
    required this.totalvalue,
    required this.totalbalance,
    List<TransactionPaidPayDetailModel>? details,
    required this.transflag,
    List<LanguageDataModel>? custnames,
    String? guidfixed,
    String? paymentdetailraw,
    List<BillPayStruct>? paymentStructs,
    double? paycashamount,
    double? sumqrcode,
    double? sumcreditcard,
    double? summoneytransfer,
    double? sumcheque,
    double? sumcoupon,
    double? sumcredit,
    double? roundamount,
    double? paycashchange,
    BranchModel? branch,
  }) : custnames = custnames ?? <LanguageDataModel>[],
       guidfixed = guidfixed ?? '',
       shopid = shopid ?? global.apiShopCode,
       paymentdetailraw = paymentdetailraw ?? "",
       paymentStructs = paymentStructs ?? <BillPayStruct>[],
       paycashamount = paycashamount ?? 0,
       sumqrcode = sumqrcode ?? 0,
       sumcreditcard = sumcreditcard ?? 0,
       summoneytransfer = summoneytransfer ?? 0,
       sumcheque = sumcheque ?? 0,
       sumcoupon = sumcoupon ?? 0,
       sumcredit = sumcredit ?? 0,
       roundamount = roundamount ?? 0,
       paycashchange = paycashchange ?? 0,
       branch = branch ?? BranchModel(),
       details = details ?? <TransactionPaidPayDetailModel>[];

  factory TransactionPaidPayModel.fromJson(Map<String, dynamic> json) =>
      _$TransactionPaidPayModelFromJson(json);

  Map<String, dynamic> toJson() => _$TransactionPaidPayModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class TransactionPaidPayDetailModel {
  bool? selected;
  String docno;
  String docdatetime;
  int transflag;
  double value;
  double balance;
  double paymentamount;
  TransactionPaidPayDetailModel({
    bool? selected = false,
    required this.docno,
    required this.docdatetime,
    required this.transflag,
    required this.value,
    required this.balance,
    required this.paymentamount,
  }) : selected = selected ?? false;

  factory TransactionPaidPayDetailModel.fromJson(Map<String, dynamic> json) =>
      _$TransactionPaidPayDetailModelFromJson(json);
  Map<String, dynamic> toJson() => _$TransactionPaidPayDetailModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class BillPayStruct {
  /// ประเภทเอกสาร (1 = ขาย, 2 = คืน)
  int? doc_mode;

  /// 1=บัตรเครดิต,2=เงินโอน,3=เช็ค,4=คูปอง,5=QR
  int? trans_flag;

  /// รหัสธนาคาร
  String? bank_code;

  /// ชื่อธนาคาร (อื่นๆ)
  String? bank_name;

  /// เลขที่บัญชี (เงินเข้า)
  String? book_bank_code;

  /// เลขที่บัตรเครดิต
  String? card_number;

  /// รหัสอนุมัติ
  String? approved_code;

  /// วันที่โอนเงิน
  DateTime? doc_date_time;

  /// สาขาธนาคาร
  String? branch_number;

  /// รหัสอ้างอิงธนาคาร
  String? bank_reference;

  /// วันที่สั่งจ่ายบนเช็ค
  DateTime? due_date;

  /// เลขที่เช็ค
  String? cheque_number;

  /// รหัสส่วนลด
  String? code;

  /// รายละเอียด (เพิ่มเติม)
  String? description;

  /// เลขคูปอง
  String? number;

  /// อ้างอิง 1
  String? reference_one;

  /// อ้างอิง 2
  String? reference_two;

  /// รหัสกระเป๋า เจ้าของเงิน (Provider)
  String? provider_code;

  /// เจ้าของเงิน (Provider)
  String? provider_name;

  /// จำนวนเงิน
  double? amount;

  BillPayStruct({
    int? doc_mode,
    int? trans_flag,
    String? bank_code,
    String? bank_name,
    String? book_bank_code,
    String? card_number,
    String? approved_code,
    DateTime? doc_date_time,
    String? branch_number,
    String? bank_reference,
    DateTime? due_date,
    String? cheque_number,
    String? code,
    String? description,
    String? number,
    String? reference_one,
    String? reference_two,
    String? provider_code,
    String? provider_name,
    double? amount,
  }) : doc_mode = doc_mode ?? 0,
       trans_flag = trans_flag ?? 0,
       bank_code = bank_code ?? "",
       bank_name = bank_name ?? "",
       book_bank_code = book_bank_code ?? "",
       card_number = card_number ?? "",
       approved_code = approved_code ?? "",
       doc_date_time = doc_date_time ?? DateTime.now(),
       branch_number = branch_number ?? "",
       bank_reference = bank_reference ?? "",
       due_date = due_date ?? DateTime.now(),
       cheque_number = cheque_number ?? "",
       code = code ?? "",
       description = description ?? "",
       number = number ?? "",
       reference_one = reference_one ?? "",
       reference_two = reference_two ?? "",
       provider_code = provider_code ?? "",
       provider_name = provider_name ?? "",
       amount = amount ?? 0;

  factory BillPayStruct.fromJson(Map<String, dynamic> json) =>
      _$BillPayStructFromJson(json);
  Map<String, dynamic> toJson() => _$BillPayStructToJson(this);

  BillPayStruct copyWith({
    int? doc_mode,
    int? trans_flag,
    String? bank_code,
    String? bank_name,
    String? book_bank_code,
    String? card_number,
    String? approved_code,
    DateTime? doc_date_time,
    String? branch_number,
    String? bank_reference,
    DateTime? due_date,
    String? cheque_number,
    String? code,
    String? description,
    String? number,
    String? reference_one,
    String? reference_two,
    String? provider_code,
    String? provider_name,
    double? amount,
  }) {
    return BillPayStruct(
      doc_mode: doc_mode ?? this.doc_mode,
      trans_flag: trans_flag ?? this.trans_flag,
      bank_code: bank_code ?? this.bank_code,
      bank_name: bank_name ?? this.bank_name,
      book_bank_code: book_bank_code ?? this.book_bank_code,
      card_number: card_number ?? this.card_number,
      approved_code: approved_code ?? this.approved_code,
      doc_date_time: doc_date_time ?? this.doc_date_time,
      branch_number: branch_number ?? this.branch_number,
      bank_reference: bank_reference ?? this.bank_reference,
      due_date: due_date ?? this.due_date,
      cheque_number: cheque_number ?? this.cheque_number,
      code: code ?? this.code,
      description: description ?? this.description,
      number: number ?? this.number,
      reference_one: reference_one ?? this.reference_one,
      reference_two: reference_two ?? this.reference_two,
      provider_code: provider_code ?? this.provider_code,
      provider_name: provider_name ?? this.provider_name,
      amount: amount ?? this.amount,
    );
  }
}

@JsonSerializable()
class ExtraJsonListModel {
  String? barcode;
  String? item_code;
  String? unit_code;
  String? unit_name;
  double? qty;
  double? price;
  double? total_amount;
  bool? is_except_vat;
  int? vat_type;
  double? price_exclude_vat;
  String? item_name;
  List<LanguageDataModel>? itemnames;

  ExtraJsonListModel({
    String? barcode,
    String? item_code,
    String? unit_code,
    String? unit_name,
    double? qty,
    double? price,
    double? total_amount,
    bool? is_except_vat,
    int? vat_type,
    double? price_exclude_vat,
    String? item_name,
    List<LanguageDataModel>? itemnames,
  }) : barcode = barcode ?? "",
       item_code = item_code ?? "",
       unit_code = unit_code ?? "",
       unit_name = unit_name ?? "",
       qty = qty ?? 0,
       price = price ?? 0,
       total_amount = total_amount ?? 0,
       is_except_vat = is_except_vat ?? false,
       vat_type = vat_type ?? 0,
       price_exclude_vat = price_exclude_vat ?? 0,
       itemnames = itemnames ?? <LanguageDataModel>[],
       item_name = item_name ?? "";

  factory ExtraJsonListModel.fromJson(Map<String, dynamic> json) =>
      _$ExtraJsonListModelFromJson(json);
  Map<String, dynamic> toJson() => _$ExtraJsonListModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class GetCustcodeTransationModel {
  String? shopid;
  String? parid;
  String? docno;
  DateTime? docdate;
  String? creditorcode;
  String? inquirytype;
  int? transflag8;
  String? totalvalue;
  String? totalbeforevat;
  String? totalvatvalue;
  String? totalexceptvat;
  String? totalaftervat;
  String? totalamount;
  String? paidamount;
  String? balanceamount;
  int? status;
  bool? iscancel;
  String? guidfixed;
  GetCustcodeTransationModel({
    String? shopid,
    String? parid,
    String? docno,
    DateTime? docdate,
    String? creditorcode,
    String? inquirytype,
    int? transflag8,
    String? totalvalue,
    String? totalbeforevat,
    String? totalvatvalue,
    String? totalexceptvat,
    String? totalaftervat,
    String? totalamount,
    String? paidamount,
    String? balanceamount,
    int? status,
    bool? iscancel,
    String? guidfixed,
  }) : shopid = shopid ?? "",
       parid = parid ?? "",
       docno = docno ?? "",
       docdate = docdate ?? DateTime.now(),
       creditorcode = creditorcode ?? "",
       inquirytype = inquirytype ?? "",
       transflag8 = transflag8 ?? 0,
       totalvalue = totalvalue ?? "",
       totalbeforevat = totalbeforevat ?? "",
       totalvatvalue = totalvatvalue ?? "",
       totalexceptvat = totalexceptvat ?? "",
       totalaftervat = totalaftervat ?? "",
       totalamount = totalamount ?? "",
       paidamount = paidamount ?? "",
       balanceamount = balanceamount ?? "",
       status = status ?? 0,
       iscancel = iscancel ?? false,
       guidfixed = guidfixed ?? "";

  factory GetCustcodeTransationModel.fromJson(Map<String, dynamic> json) =>
      _$GetCustcodeTransationModelFromJson(json);
  Map<String, dynamic> toJson() => _$GetCustcodeTransationModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class CouponModel {
  String? couponno;
  double? couponamount;
  String? coupondescription;
  String? coupontype;

  CouponModel({
    String? couponno,
    double? couponamount,
    String? coupondescription,
    String? coupontype,
  }) : couponno = couponno ?? "",
       couponamount = couponamount ?? 0,
       coupondescription = coupondescription ?? "",
       coupontype = coupontype ?? "";

  factory CouponModel.fromJson(Map<String, dynamic> json) =>
      _$CouponModelFromJson(json);
  Map<String, dynamic> toJson() => _$CouponModelToJson(this);

  CouponModel copyWith({
    String? couponno,
    double? couponamount,
    String? coupondescription,
    String? coupontype,
  }) {
    return CouponModel(
      couponno: couponno ?? this.couponno,
      couponamount: couponamount ?? this.couponamount,
      coupondescription: coupondescription ?? this.coupondescription,
      coupontype: coupontype ?? this.coupontype,
    );
  }
}

@JsonSerializable(explicitToJson: true)
class DocReferenceModel {
  String docdatetime;
  String docno;
  String guidfixed;

  DocReferenceModel({
    required this.docdatetime,
    required this.docno,
    required this.guidfixed,
  });

  factory DocReferenceModel.fromJson(Map<String, dynamic> json) =>
      _$DocReferenceModelFromJson(json);

  Map<String, dynamic> toJson() => _$DocReferenceModelToJson(this);

  DocReferenceModel copyWith({
    String? docdatetime,
    String? docno,
    String? guidfixed,
  }) {
    return DocReferenceModel(
      docdatetime: docdatetime ?? this.docdatetime,
      docno: docno ?? this.docno,
      guidfixed: guidfixed ?? this.guidfixed,
    );
  }
}

@JsonSerializable(explicitToJson: true)
class AdvancePaymentDocModel {
  double balance;
  String docdatetime;
  String docno;
  String guidfixed;
  String remark;
  double totalamount;
  double useamount;

  AdvancePaymentDocModel({
    required this.balance,
    required this.docdatetime,
    required this.docno,
    required this.guidfixed,
    required this.remark,
    required this.totalamount,
    required this.useamount,
  });

  factory AdvancePaymentDocModel.fromJson(Map<String, dynamic> json) =>
      _$AdvancePaymentDocModelFromJson(json);

  Map<String, dynamic> toJson() => _$AdvancePaymentDocModelToJson(this);

  AdvancePaymentDocModel copyWith({
    double? balance,
    String? docdatetime,
    String? docno,
    String? guidfixed,
    String? remark,
    double? totalamount,
    double? useamount,
  }) {
    return AdvancePaymentDocModel(
      balance: balance ?? this.balance,
      docdatetime: docdatetime ?? this.docdatetime,
      docno: docno ?? this.docno,
      guidfixed: guidfixed ?? this.guidfixed,
      remark: remark ?? this.remark,
      totalamount: totalamount ?? this.totalamount,
      useamount: useamount ?? this.useamount,
    );
  }
}

@JsonSerializable(explicitToJson: true)
class DepositDocModel {
  double balance;
  String docdatetime;
  String docno;
  String guidfixed;
  String remark;
  double totalamount;
  double useamount;

  DepositDocModel({
    required this.balance,
    required this.docdatetime,
    required this.docno,
    required this.guidfixed,
    required this.remark,
    required this.totalamount,
    required this.useamount,
  });

  factory DepositDocModel.fromJson(Map<String, dynamic> json) =>
      _$DepositDocModelFromJson(json);

  Map<String, dynamic> toJson() => _$DepositDocModelToJson(this);

  DepositDocModel copyWith({
    double? balance,
    String? docdatetime,
    String? docno,
    String? guidfixed,
    String? remark,
    double? totalamount,
    double? useamount,
  }) {
    return DepositDocModel(
      balance: balance ?? this.balance,
      docdatetime: docdatetime ?? this.docdatetime,
      docno: docno ?? this.docno,
      guidfixed: guidfixed ?? this.guidfixed,
      remark: remark ?? this.remark,
      totalamount: totalamount ?? this.totalamount,
      useamount: useamount ?? this.useamount,
    );
  }
}
