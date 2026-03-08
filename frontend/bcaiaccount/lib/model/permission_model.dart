// ignore_for_file: non_constant_identifier_names

import 'package:json_annotation/json_annotation.dart';
import '../global.dart' as global;

part 'permission_model.g.dart';

/// สิทธิ์การทำงานในแต่ละหน้าจอ
@JsonSerializable(explicitToJson: true)
class ScreenPermissionModel {
  /// รหัสหน้าจอ เช่น purchaseorder, sale, purchase
  String screenCode;

  /// ชื่อหน้าจอ
  String screenName;

  /// สิทธิ์ดู (เข้าหน้าจอได้)
  bool canView;

  /// สิทธิ์เพิ่ม
  bool canAdd;

  /// สิทธิ์แก้ไขของตัวเองเท่านั้น
  bool canEditOwn;

  /// สิทธิ์แก้ไขได้ทั้งหมด
  bool canEditAll;

  /// สิทธิ์ลบของตัวเองเท่านั้น
  bool canDeleteOwn;

  /// สิทธิ์ลบได้ทั้งหมด
  bool canDeleteAll;

  /// สิทธิ์พิมพ์
  bool canPrint;

  ScreenPermissionModel({
    required this.screenCode,
    this.screenName = '',
    this.canView = true,
    this.canAdd = false,
    this.canEditOwn = false,
    this.canEditAll = false,
    this.canDeleteOwn = false,
    this.canDeleteAll = false,
    this.canPrint = false,
  });

  factory ScreenPermissionModel.fromJson(Map<String, dynamic> json) =>
      _$ScreenPermissionModelFromJson(json);

  Map<String, dynamic> toJson() => _$ScreenPermissionModelToJson(this);

  /// ตรวจสอบว่าสามารถแก้ไขได้หรือไม่ (ของตัวเองหรือทั้งหมด)
  bool get canEdit => canEditOwn || canEditAll;

  /// ตรวจสอบว่าสามารถลบได้หรือไม่ (ของตัวเองหรือทั้งหมด)
  bool get canDelete => canDeleteOwn || canDeleteAll;

  ScreenPermissionModel copyWith({
    String? screenCode,
    String? screenName,
    bool? canView,
    bool? canAdd,
    bool? canEditOwn,
    bool? canEditAll,
    bool? canDeleteOwn,
    bool? canDeleteAll,
    bool? canPrint,
  }) {
    return ScreenPermissionModel(
      screenCode: screenCode ?? this.screenCode,
      screenName: screenName ?? this.screenName,
      canView: canView ?? this.canView,
      canAdd: canAdd ?? this.canAdd,
      canEditOwn: canEditOwn ?? this.canEditOwn,
      canEditAll: canEditAll ?? this.canEditAll,
      canDeleteOwn: canDeleteOwn ?? this.canDeleteOwn,
      canDeleteAll: canDeleteAll ?? this.canDeleteAll,
      canPrint: canPrint ?? this.canPrint,
    );
  }
}

/// สิทธิ์การเข้าถึงสาขา
@JsonSerializable(explicitToJson: true)
class BranchPermissionModel {
  /// รหัสสาขา
  String branchCode;

  /// ชื่อสาขา
  String branchName;

  /// รายการสิทธิ์หน้าจอในสาขานี้
  List<ScreenPermissionModel> screens;

  BranchPermissionModel({
    required this.branchCode,
    this.branchName = '',
    List<ScreenPermissionModel>? screens,
  }) : screens = screens ?? [];

  factory BranchPermissionModel.fromJson(Map<String, dynamic> json) =>
      _$BranchPermissionModelFromJson(json);

  Map<String, dynamic> toJson() => _$BranchPermissionModelToJson(this);

  BranchPermissionModel copyWith({
    String? branchCode,
    String? branchName,
    List<ScreenPermissionModel>? screens,
  }) {
    return BranchPermissionModel(
      branchCode: branchCode ?? this.branchCode,
      branchName: branchName ?? this.branchName,
      screens: screens ?? this.screens.map((e) => e.copyWith()).toList(),
    );
  }
}

/// Model สำหรับกำหนดสิทธิ์ (Permission Definition)
/// Collection: permission_definitions
@JsonSerializable(explicitToJson: true)
class PermissionDefinitionModel {
  /// ID ของเอกสาร (MongoDB _id)
  /// ไม่รวม _id ถ้าเป็น null เพื่อให้ MongoDB สร้างให้อัตโนมัติ
  @JsonKey(name: '_id', includeIfNull: false)
  String? id;

  /// รหัสร้านค้า
  String shopid;

  /// รหัสสิทธิ์ (ไม่ซ้ำ)
  String permissionCode;

  /// ชื่อสิทธิ์
  String permissionName;

  /// คำอธิบาย
  String description;

  /// รายการสิทธิ์ตามสาขา
  List<BranchPermissionModel> branches;

  /// สถานะใช้งาน
  bool isActive;

  /// วันที่สร้าง
  String createdAt;

  /// วันที่แก้ไขล่าสุด
  String updatedAt;

  /// ผู้สร้าง
  String createdBy;

  /// ผู้แก้ไขล่าสุด
  String updatedBy;

  PermissionDefinitionModel({
    this.id,
    required this.shopid,
    required this.permissionCode,
    this.permissionName = '',
    this.description = '',
    List<BranchPermissionModel>? branches,
    this.isActive = true,
    String? createdAt,
    String? updatedAt,
    this.createdBy = '',
    this.updatedBy = '',
  })  : branches = branches ?? [],
        createdAt = createdAt ?? DateTime.now().toIso8601String(),
        updatedAt = updatedAt ?? DateTime.now().toIso8601String();

  factory PermissionDefinitionModel.fromJson(Map<String, dynamic> json) =>
      _$PermissionDefinitionModelFromJson(json);

  Map<String, dynamic> toJson() => _$PermissionDefinitionModelToJson(this);

  PermissionDefinitionModel copyWith({
    String? id,
    String? shopid,
    String? permissionCode,
    String? permissionName,
    String? description,
    List<BranchPermissionModel>? branches,
    bool? isActive,
    String? createdAt,
    String? updatedAt,
    String? createdBy,
    String? updatedBy,
  }) {
    return PermissionDefinitionModel(
      id: id ?? this.id,
      shopid: shopid ?? this.shopid,
      permissionCode: permissionCode ?? this.permissionCode,
      permissionName: permissionName ?? this.permissionName,
      description: description ?? this.description,
      branches: branches ?? this.branches.map((e) => e.copyWith()).toList(),
      isActive: isActive ?? this.isActive,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      createdBy: createdBy ?? this.createdBy,
      updatedBy: updatedBy ?? this.updatedBy,
    );
  }
}

/// Model สำหรับเชื่อมพนักงานกับสิทธิ์ (Employee Permission Link)
/// Collection: employee_permissions
@JsonSerializable(explicitToJson: true)
class EmployeePermissionModel {
  /// ID ของเอกสาร (MongoDB _id)
  /// ไม่รวม _id ถ้าเป็น null เพื่อให้ MongoDB สร้างให้อัตโนมัติ
  @JsonKey(name: '_id', includeIfNull: false)
  String? id;

  /// รหัสร้านค้า
  String shopid;

  /// รหัสพนักงาน
  String employeeCode;

  /// ชื่อพนักงาน (สำหรับแสดงผล)
  String employeeName;

  /// รายการรหัสสิทธิ์ที่เชื่อม (สามารถมีได้หลายตัว - จะ OR กัน)
  List<String> permissionCodes;

  /// วันที่สร้าง
  String createdAt;

  /// วันที่แก้ไขล่าสุด
  String updatedAt;

  /// ผู้สร้าง
  String createdBy;

  /// ผู้แก้ไขล่าสุด
  String updatedBy;

  EmployeePermissionModel({
    this.id,
    required this.shopid,
    required this.employeeCode,
    this.employeeName = '',
    List<String>? permissionCodes,
    String? createdAt,
    String? updatedAt,
    this.createdBy = '',
    this.updatedBy = '',
  })  : permissionCodes = permissionCodes ?? [],
        createdAt = createdAt ?? DateTime.now().toIso8601String(),
        updatedAt = updatedAt ?? DateTime.now().toIso8601String();

  factory EmployeePermissionModel.fromJson(Map<String, dynamic> json) =>
      _$EmployeePermissionModelFromJson(json);

  Map<String, dynamic> toJson() => _$EmployeePermissionModelToJson(this);

  EmployeePermissionModel copyWith({
    String? id,
    String? shopid,
    String? employeeCode,
    String? employeeName,
    List<String>? permissionCodes,
    String? createdAt,
    String? updatedAt,
    String? createdBy,
    String? updatedBy,
  }) {
    return EmployeePermissionModel(
      id: id ?? this.id,
      shopid: shopid ?? this.shopid,
      employeeCode: employeeCode ?? this.employeeCode,
      employeeName: employeeName ?? this.employeeName,
      permissionCodes: permissionCodes ?? List.from(this.permissionCodes),
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      createdBy: createdBy ?? this.createdBy,
      updatedBy: updatedBy ?? this.updatedBy,
    );
  }
}

/// Model สำหรับเก็บสิทธิ์รวมของผู้ใช้ (หลังจาก OR สิทธิ์ทั้งหมดแล้ว)
/// ใช้ใน runtime เท่านั้น ไม่บันทึกลง database
class UserEffectivePermissionModel {
  /// รหัสพนักงาน
  final String employeeCode;

  /// Map ของสิทธิ์: branchCode -> screenCode -> ScreenPermissionModel
  final Map<String, Map<String, ScreenPermissionModel>> permissions;

  /// เวลาที่โหลดสิทธิ์ล่าสุด
  final DateTime loadedAt;

  UserEffectivePermissionModel({
    required this.employeeCode,
    required this.permissions,
    DateTime? loadedAt,
  }) : loadedAt = loadedAt ?? DateTime.now();

  /// ตรวจสอบว่ามีสิทธิ์เข้าสาขานี้หรือไม่
  bool canAccessBranch(String branchCode) {
    return permissions.containsKey(branchCode);
  }

  /// ตรวจสอบว่ามีสิทธิ์เข้าหน้าจอนี้ในสาขานี้หรือไม่
  bool canAccessScreen(String branchCode, String screenCode) {
    if (!permissions.containsKey(branchCode)) return false;
    final branchPermissions = permissions[branchCode]!;
    if (!branchPermissions.containsKey(screenCode)) return false;
    return branchPermissions[screenCode]!.canView;
  }

  /// ดึงสิทธิ์หน้าจอ
  ScreenPermissionModel? getScreenPermission(String branchCode, String screenCode) {
    if (!permissions.containsKey(branchCode)) return null;
    return permissions[branchCode]?[screenCode];
  }

  /// ตรวจสอบสิทธิ์เพิ่ม
  bool canAdd(String branchCode, String screenCode) {
    return getScreenPermission(branchCode, screenCode)?.canAdd ?? false;
  }

  /// ตรวจสอบสิทธิ์แก้ไขตัวเอง
  bool canEditOwn(String branchCode, String screenCode) {
    return getScreenPermission(branchCode, screenCode)?.canEditOwn ?? false;
  }

  /// ตรวจสอบสิทธิ์แก้ไขทั้งหมด
  bool canEditAll(String branchCode, String screenCode) {
    return getScreenPermission(branchCode, screenCode)?.canEditAll ?? false;
  }

  /// ตรวจสอบสิทธิ์แก้ไข (ตัวเองหรือทั้งหมด)
  bool canEdit(String branchCode, String screenCode) {
    final perm = getScreenPermission(branchCode, screenCode);
    return perm?.canEditOwn == true || perm?.canEditAll == true;
  }

  /// ตรวจสอบสิทธิ์ลบตัวเอง
  bool canDeleteOwn(String branchCode, String screenCode) {
    return getScreenPermission(branchCode, screenCode)?.canDeleteOwn ?? false;
  }

  /// ตรวจสอบสิทธิ์ลบทั้งหมด
  bool canDeleteAll(String branchCode, String screenCode) {
    return getScreenPermission(branchCode, screenCode)?.canDeleteAll ?? false;
  }

  /// ตรวจสอบสิทธิ์ลบ (ตัวเองหรือทั้งหมด)
  bool canDelete(String branchCode, String screenCode) {
    final perm = getScreenPermission(branchCode, screenCode);
    return perm?.canDeleteOwn == true || perm?.canDeleteAll == true;
  }

  /// ตรวจสอบสิทธิ์พิมพ์
  bool canPrint(String branchCode, String screenCode) {
    return getScreenPermission(branchCode, screenCode)?.canPrint ?? false;
  }

  /// ตรวจสอบว่าสามารถแก้ไขเอกสารนี้ได้หรือไม่ (โดยดูจากผู้สร้าง)
  bool canEditDocument(String branchCode, String screenCode, String createdBy, String currentUser) {
    final perm = getScreenPermission(branchCode, screenCode);
    if (perm == null) return false;
    if (perm.canEditAll) return true;
    if (perm.canEditOwn && createdBy == currentUser) return true;
    return false;
  }

  /// ตรวจสอบว่าสามารถลบเอกสารนี้ได้หรือไม่ (โดยดูจากผู้สร้าง)
  bool canDeleteDocument(String branchCode, String screenCode, String createdBy, String currentUser) {
    final perm = getScreenPermission(branchCode, screenCode);
    if (perm == null) return false;
    if (perm.canDeleteAll) return true;
    if (perm.canDeleteOwn && createdBy == currentUser) return true;
    return false;
  }

  /// ตรวจสอบว่าสิทธิ์หมดอายุหรือยัง (เกิน 5 นาที)
  bool isExpired() {
    return DateTime.now().difference(loadedAt).inMinutes >= 5;
  }
}

/// รายการหน้าจอทั้งหมดในระบบ (สำหรับใช้ในหน้ากำหนดสิทธิ์)
class ScreenDefinition {
  final String code;
  final String name;
  final String category;

  const ScreenDefinition({
    required this.code,
    required this.name,
    required this.category,
  });
}

/// รายการหน้าจอทั้งหมดที่สามารถกำหนดสิทธิ์ได้
final List<ScreenDefinition> allScreenDefinitions = [
  // ═══════════════════════════════════════════════════════════════
  // ตั้งค่าระบบ (System Config)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'company', name: global.language('company_data'), category: 'system'),
  ScreenDefinition(code: 'branch', name: global.language('branch'), category: 'system'),
  ScreenDefinition(code: 'department', name: global.language('department'), category: 'system'),
  ScreenDefinition(code: 'business_type', name: global.language('business_type'), category: 'system'),
  ScreenDefinition(code: 'work_day', name: global.language('work_day'), category: 'system'),
  ScreenDefinition(code: 'holiday', name: global.language('holiday'), category: 'system'),
  ScreenDefinition(code: 'formdesign', name: global.language('form_design'), category: 'system'),
  ScreenDefinition(code: 'line_notify', name: 'Line Notify', category: 'system'),

  // ═══════════════════════════════════════════════════════════════
  // ผู้ใช้และพนักงาน (Users & Employees)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'employee', name: global.language('employee'), category: 'user'),
  ScreenDefinition(code: 'user', name: global.language('user'), category: 'user'),

  // ═══════════════════════════════════════════════════════════════
  // ลูกค้าและผู้ขาย (Customers & Vendors)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'creditor', name: global.language('creditor'), category: 'partner'),
  ScreenDefinition(code: 'creditorgroup', name: global.language('creditor_group'), category: 'partner'),
  ScreenDefinition(code: 'debtor', name: global.language('debtor'), category: 'partner'),
  ScreenDefinition(code: 'debtorgroup', name: global.language('debtor_group'), category: 'partner'),

  // ═══════════════════════════════════════════════════════════════
  // ตั้งค่าขาย (Sale Settings)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'color', name: global.language('color'), category: 'sale_setting'),
  ScreenDefinition(code: 'qrprovider', name: 'QR Provider', category: 'sale_setting'),
  ScreenDefinition(code: 'bank', name: global.language('bank'), category: 'sale_setting'),
  ScreenDefinition(code: 'book_bank', name: global.language('book_bank'), category: 'sale_setting'),
  ScreenDefinition(code: 'sale_channel', name: global.language('sale_channel'), category: 'sale_setting'),
  ScreenDefinition(code: 'transport_channel', name: global.language('transport_channel'), category: 'sale_setting'),
  ScreenDefinition(code: 'possetting', name: global.language('pos_setting'), category: 'sale_setting'),
  ScreenDefinition(code: 'posmedia', name: global.language('pos_media'), category: 'sale_setting'),
  ScreenDefinition(code: 'point_setting', name: global.language('point_setting'), category: 'sale_setting'),
  ScreenDefinition(code: 'coupon_setting', name: 'ตั้งค่าคูปอง', category: 'sale_setting'),
  ScreenDefinition(code: 'docformat', name: global.language('doc_format'), category: 'sale_setting'),
  ScreenDefinition(code: 'billdesign', name: global.language('bill_design'), category: 'sale_setting'),

  // ═══════════════════════════════════════════════════════════════
  // สินค้า (Products)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'product_barcode', name: 'สินค้า/บาร์โค้ด', category: 'product'),
  ScreenDefinition(code: 'productunit', name: global.language('product_unit'), category: 'product'),
  ScreenDefinition(code: 'product_barcode_shelf', name: global.language('product_shelf'), category: 'product'),
  ScreenDefinition(code: 'product_category', name: global.language('product_category_list'), category: 'product'),
  ScreenDefinition(code: 'productcategorylist', name: 'รายการหมวดหมู่', category: 'product'),
  ScreenDefinition(code: 'product_warehouse', name: global.language('warehouse'), category: 'product'),
  ScreenDefinition(code: 'product_location', name: global.language('product_location'), category: 'product'),
  ScreenDefinition(code: 'order_type', name: 'ประเภทออเดอร์', category: 'product'),
  ScreenDefinition(code: 'product_type', name: global.language('product_type'), category: 'product'),
  ScreenDefinition(code: 'product_dimension', name: global.language('product_dimension'), category: 'product'),
  ScreenDefinition(code: 'product_bom', name: 'BOM', category: 'product'),
  ScreenDefinition(code: 'promotion', name: global.language('promotion'), category: 'product'),
  ScreenDefinition(code: 'price_history', name: global.language('price_history'), category: 'product'),
  ScreenDefinition(code: 'importproduct', name: global.language('import_product'), category: 'product'),
  ScreenDefinition(code: 'importproductimage', name: global.language('import_product_image'), category: 'product'),

  // ═══════════════════════════════════════════════════════════════
  // ธุรกรรมซื้อ (Purchase Transactions)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'purchaseorder', name: global.language('transaction_purchase_order'), category: 'purchase'),
  ScreenDefinition(code: 'advancepayment', name: global.language('pay_advance'), category: 'purchase'),
  ScreenDefinition(code: 'advancepaymentrefund', name: 'คืนเงินจ่ายล่วงหน้า', category: 'purchase'),
  ScreenDefinition(code: 'deposit', name: 'มัดจำจ่าย', category: 'purchase'),
  ScreenDefinition(code: 'depositrefund', name: 'คืนเงินมัดจำจ่าย', category: 'purchase'),
  ScreenDefinition(code: 'purchase', name: global.language('transaction_purchase'), category: 'purchase'),
  ScreenDefinition(code: 'purchasereturn', name: 'ส่งคืน', category: 'purchase'),
  ScreenDefinition(code: 'purchasepartial', name: global.language('transaction_purchase_partial'), category: 'purchase'),
  ScreenDefinition(code: 'accrualreceive', name: 'รับสินค้าค้างรับ', category: 'purchase'),

  // ═══════════════════════════════════════════════════════════════
  // ธุรกรรมขาย (Sale Transactions)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'quotation', name: global.language('transaction_quotation'), category: 'sale'),
  ScreenDefinition(code: 'saleorder', name: global.language('transaction_sale_order'), category: 'sale'),
  ScreenDefinition(code: 'paidadvance', name: global.language('receive_advance'), category: 'sale'),
  ScreenDefinition(code: 'paidadvancerefund', name: 'คืนเงินรับล่วงหน้า', category: 'sale'),
  ScreenDefinition(code: 'receivedeposit', name: global.language('receive_deposit'), category: 'sale'),
  ScreenDefinition(code: 'receivedepositrefund', name: 'คืนเงินรับมัดจำ', category: 'sale'),
  ScreenDefinition(code: 'sale', name: global.language('transaction_sale'), category: 'sale'),
  ScreenDefinition(code: 'salereturn', name: global.language('return_goods'), category: 'sale'),

  // ═══════════════════════════════════════════════════════════════
  // สต๊อก (Stock)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'stocktransfer', name: global.language('transaction_stock_transfer'), category: 'stock'),
  ScreenDefinition(code: 'stockreceiveproduct', name: global.language('transaction_stock_receive_product'), category: 'stock'),
  ScreenDefinition(code: 'stockpickupproduct', name: global.language('transaction_stock_pick_up_product'), category: 'stock'),
  ScreenDefinition(code: 'stockreturnproduct', name: global.language('transaction_stock_return_product'), category: 'stock'),
  ScreenDefinition(code: 'adjust', name: global.language('transaction_adjust'), category: 'stock'),
  ScreenDefinition(code: 'stockbalance', name: global.language('stock_balance'), category: 'stock'),

  // ═══════════════════════════════════════════════════════════════
  // การเงิน (Finance)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'paid', name: global.language('paid'), category: 'finance'),
  ScreenDefinition(code: 'pay', name: global.language('pay'), category: 'finance'),
  ScreenDefinition(code: 'slip_money_in', name: 'รับเงินสด', category: 'finance'),
  ScreenDefinition(code: 'slip_money_out', name: 'จ่ายเงินสด', category: 'finance'),

  // ═══════════════════════════════════════════════════════════════
  // รายงาน (Reports)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'report_dedebi_sales', name: 'รายงานขาย', category: 'report'),
  ScreenDefinition(code: 'report_dedebi_sales_daily', name: 'รายงานขายรายวัน', category: 'report'),
  ScreenDefinition(code: 'report_stock_movement', name: 'รายงานเคลื่อนไหวสต๊อก', category: 'report'),
  ScreenDefinition(code: 'report_dedebi_payment_daily', name: 'รายงานชำระรายวัน', category: 'report'),
  ScreenDefinition(code: 'report_dedebi_sale_return', name: 'รายงานรับคืน', category: 'report'),
  ScreenDefinition(code: 'report_dedebi_stock_balance', name: 'รายงานยอดคงเหลือ', category: 'report'),
  ScreenDefinition(code: 'report_dedebi_purchase_partial', name: 'รายงานซื้อบางส่วน', category: 'report'),
  ScreenDefinition(code: 'report_gross_profit_by_document', name: global.language('report_gross_profit_by_doc'), category: 'report'),
  ScreenDefinition(code: 'report_gross_profit_by_product', name: global.language('report_gross_profit_by_product'), category: 'report'),
  ScreenDefinition(code: 'report_vat_sale', name: global.language('report_vat_sale'), category: 'report'),
  ScreenDefinition(code: 'report_vat_buy', name: global.language('report_vat_buy'), category: 'report'),
  ScreenDefinition(code: 'report_stock_balance_item', name: global.language('balance_by_product'), category: 'report'),
  ScreenDefinition(code: 'report_stock_balance_warehouse', name: global.language('balance_by_warehouse'), category: 'report'),
  ScreenDefinition(code: 'report_stock_balance_location', name: 'ยอดคงเหลือตามตำแหน่ง', category: 'report'),
  ScreenDefinition(code: 'report_stock_movement_cost', name: 'เคลื่อนไหวสต๊อกต้นทุน', category: 'report'),
  ScreenDefinition(code: 'report_sales_by_document', name: 'รายงานขายตามเอกสาร', category: 'report'),

  // ═══════════════════════════════════════════════════════════════
  // ร้านอาหาร (Restaurant)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'zone', name: global.language('zone'), category: 'restaurant'),
  ScreenDefinition(code: 'table', name: global.language('table'), category: 'restaurant'),
  ScreenDefinition(code: 'table_map', name: global.language('table_map'), category: 'restaurant'),
  ScreenDefinition(code: 'kitchen', name: 'ครัว', category: 'restaurant'),
  ScreenDefinition(code: 'add_product_to_kitchen', name: 'เพิ่มสินค้าเข้าครัว', category: 'restaurant'),
  ScreenDefinition(code: 'qrcodeorder', name: 'QR Code Order', category: 'restaurant'),
  ScreenDefinition(code: 'ordertemplatsetting', name: 'แม่แบบออเดอร์', category: 'restaurant'),
  ScreenDefinition(code: 'ordersetting', name: 'ตั้งค่าออเดอร์', category: 'restaurant'),

  // ═══════════════════════════════════════════════════════════════
  // ตรวจสอบ (Check)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'daily_info', name: global.language('daily_info'), category: 'check'),
  ScreenDefinition(code: 'cashing_drawer', name: 'ลิ้นชักเงินสด', category: 'check'),

  // ═══════════════════════════════════════════════════════════════
  // ข้อมูลหลัก (Master Data)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'master_brand', name: global.language('brand'), category: 'master'),
  ScreenDefinition(code: 'master_category', name: global.language('product_category'), category: 'master'),
  ScreenDefinition(code: 'master_class', name: 'คลาส', category: 'master'),
  ScreenDefinition(code: 'master_design', name: global.language('design'), category: 'master'),
  ScreenDefinition(code: 'master_grade', name: global.language('grade'), category: 'master'),
  ScreenDefinition(code: 'master_model', name: global.language('model'), category: 'master'),
  ScreenDefinition(code: 'master_pattern', name: global.language('pattern'), category: 'master'),
  ScreenDefinition(code: 'master_group', name: global.language('group'), category: 'master'),
  ScreenDefinition(code: 'master_group_sub1', name: global.language('sub_group_1'), category: 'master'),
  ScreenDefinition(code: 'master_group_sub2', name: global.language('sub_group_2'), category: 'master'),

  // ═══════════════════════════════════════════════════════════════
  // นำเข้า/ส่งออก (Import/Export)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'importproductfromfile', name: global.language('import_from_file'), category: 'import_export'),

  // ═══════════════════════════════════════════════════════════════
  // AI & ระบบอัจฉริยะ (AI & Smart)
  // ═══════════════════════════════════════════════════════════════
  ScreenDefinition(code: 'knowledge_base', name: 'ฐานความรู้', category: 'ai'),
  ScreenDefinition(code: 'alert_agent', name: 'แจ้งเตือนอัจฉริยะ', category: 'ai'),
];
