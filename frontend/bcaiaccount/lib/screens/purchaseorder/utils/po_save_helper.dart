import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Helper class สำหรับเตรียมข้อมูลก่อนบันทึก PO
///
/// รับผิดชอบ:
/// - แปลงวันที่เป็น UTC
/// - ใส่ running number ของบรรทัดสินค้า
/// - ใส่ข้อมูลผู้สร้าง/แก้ไขเอกสาร
class POSaveHelper {
  /// เตรียมข้อมูลก่อนบันทึก
  ///
  /// [screenData] - TransactionModel ที่จะบันทึก
  /// [creatorCode] - รหัสผู้สร้าง
  /// [creatorName] - ชื่อผู้สร้าง
  static void prepareDataForSave(
    TransactionModel screenData, {
    required String creatorCode,
    required String creatorName,
  }) {
    // 0. ตรวจสอบ docdatelocal/doctimelocal/timezone — ต้องมีค่าเสมอ
    // เอกสารเก่าที่โหลดจาก MongoDB อาจไม่มี fields เหล่านี้
    _ensureTimezoneFields(screenData);

    // 1. แปลงวันที่หัวเอกสารเป็น UTC
    _convertHeaderDatesToUtc(screenData);

    // 2. เตรียมข้อมูลรายการสินค้า (linenumber, วันที่, standvalue/dividevalue)
    _prepareDetailsData(screenData);

    // 3. PO ไม่มีเงินมัดจำ — บังคับ sumdeposit = 0
    screenData.sumdeposit = 0;

    // 4. กำหนด created_at/modified_at ให้เป็น UTC ISO string
    // ป้องกัน backend error เมื่อส่ง "" (empty string) ไปให้ Go time.Time
    _setTimestamps(screenData);

    // 5. บันทึกข้อมูลผู้สร้างเอกสาร
    // เซ็ตเฉพาะเอกสารใหม่ (creatorcode ว่าง) — เอกสารเดิมจะคงผู้สร้างเดิมไว้
    // mainapi จะรับผิดชอบเรื่อง creator สำหรับ update (restore ค่าเดิมจาก MongoDB)
    if (screenData.creatorcode == null || screenData.creatorcode!.isEmpty) {
      screenData.creatorcode = creatorCode;
      screenData.creatorname = creatorName;
      AppLogger.debug('[POSaveHelper] ตั้งค่าผู้สร้างเอกสาร: $creatorCode / $creatorName');
    }

    AppLogger.debug('[POSaveHelper] เตรียมข้อมูลก่อนบันทึกเรียบร้อย');
  }

  /// แปลงวันที่หัวเอกสารเป็น UTC
  static void _convertHeaderDatesToUtc(TransactionModel screenData) {
    // docdatetime
    DateTime docDatetimeUtc = _safeParseDatetime(screenData.docdatetime);
    screenData.docdatetime = docDatetimeUtc.toUtc().toIso8601String();

    // docrefdate
    DateTime docRefDatetimeUtc = _safeParseDatetime(screenData.docrefdate);
    screenData.docrefdate = docRefDatetimeUtc.toUtc().toIso8601String();

    // taxdocdate
    DateTime taxDocDatetimeUtc = _safeParseDatetime(screenData.taxdocdate);
    screenData.taxdocdate = taxDocDatetimeUtc.toUtc().toIso8601String();
  }

  /// เตรียมข้อมูลรายการสินค้า
  static void _prepareDetailsData(TransactionModel screenData) {
    if (screenData.details == null) return;

    for (int i = 0; i < screenData.details!.length; i++) {
      final data = screenData.details![i];
      // ใส่ running number
      data.linenumber = i + 1;

      // sync วันที่กับหัวเอกสาร
      data.docdatetime = screenData.docdatetime;

      // แปลงวันที่อ้างอิงเป็น UTC
      DateTime docrefdatetimeDetailUtc = _safeParseDatetime(data.docrefdatetime);
      data.docrefdatetime = docrefdatetimeDetailUtc.toUtc().toIso8601String();

      // ดึงตัวตั้งตัวหารจากบาร์โค้ดอ้างอิง
      if (data.itemtype == 0 && data.refbarcodes != null && data.refbarcodes!.isNotEmpty) {
        for (var element in data.refbarcodes!) {
          data.standvalue = element.standvalue;
          data.dividevalue = element.dividevalue;
        }
      }
    }
  }

  /// กำหนด created_at/modified_at ให้เป็น UTC ISO string
  /// ถ้าเป็นเอกสารใหม่ (createdat ว่าง) → ใส่เวลาปัจจุบัน
  /// ถ้าเป็นเอกสารเดิม → อัพเดท modifiedat เป็นเวลาปัจจุบัน
  static void _setTimestamps(TransactionModel screenData) {
    final nowUtc = DateTime.now().toUtc().toIso8601String();
    if (screenData.createdat == null || screenData.createdat!.isEmpty) {
      screenData.createdat = nowUtc;
    }
    screenData.modifiedat = nowUtc;
  }

  /// ตรวจสอบ docdatelocal/doctimelocal/timezone — ต้องมีค่าเสมอ
  /// เอกสารเก่าที่โหลดจาก MongoDB อาจไม่มี fields เหล่านี้
  /// mainapi ใช้ docdatelocal ในการสร้างเลขที่เอกสาร (เช่น PO20260208xxxxx)
  static void _ensureTimezoneFields(TransactionModel screenData) {
    final needsFix = (screenData.docdatelocal == null || screenData.docdatelocal!.isEmpty) ||
        (screenData.doctimelocal == null || screenData.doctimelocal!.isEmpty) ||
        (screenData.timezone == null || screenData.timezone!.isEmpty);

    if (!needsFix) return;

    final timezoneInfo = global.getLocalTimezoneInfo();
    if (screenData.docdatelocal == null || screenData.docdatelocal!.isEmpty) {
      screenData.docdatelocal = timezoneInfo['docdatelocal'];
    }
    if (screenData.doctimelocal == null || screenData.doctimelocal!.isEmpty) {
      screenData.doctimelocal = timezoneInfo['doctimelocal'];
    }
    if (screenData.timezone == null || screenData.timezone!.isEmpty) {
      screenData.timezone = timezoneInfo['timezone'];
    }
    AppLogger.info('[POSaveHelper] เติม timezone fields: docdatelocal=${screenData.docdatelocal}, timezone=${screenData.timezone}');
  }

  /// Parse date อย่างปลอดภัย
  static DateTime _safeParseDatetime(String? dateString) {
    if (dateString == null || dateString.isEmpty) {
      return DateTime.now();
    }
    try {
      return DateTime.parse(dateString);
    } catch (e) {
      AppLogger.warning('[POSaveHelper] แปลงวันที่ไม่สำเร็จ: "$dateString" → ใช้ DateTime.now() แทน');
      return DateTime.now();
    }
  }
}
