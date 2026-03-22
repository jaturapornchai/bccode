import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Helper class สำหรับเตรียมข้อมูลก่อนบันทึก PR (ใบขอซื้อ)
///
/// รับผิดชอบ:
/// - แปลงวันที่เป็น UTC
/// - ใส่ running number ของบรรทัดสินค้า
/// - ใส่ข้อมูลผู้สร้าง/แก้ไขเอกสาร
class PRSaveHelper {
  /// เตรียมข้อมูลก่อนบันทึก
  static void prepareDataForSave(
    TransactionModel screenData, {
    required String creatorCode,
    required String creatorName,
  }) {
    // 0. ตรวจสอบ docdatelocal/doctimelocal/timezone — ต้องมีค่าเสมอ
    _ensureTimezoneFields(screenData);

    // 1. แปลงวันที่หัวเอกสารเป็น UTC
    _convertHeaderDatesToUtc(screenData);

    // 2. เตรียมข้อมูลรายการสินค้า (linenumber, วันที่, standvalue/dividevalue)
    _prepareDetailsData(screenData);

    // 3. PR ไม่มีเงินมัดจำ — บังคับ sumdeposit = 0
    screenData.sumdeposit = 0;

    // 4. กำหนด created_at/modified_at
    _setTimestamps(screenData);

    // 5. บันทึกข้อมูลผู้สร้างเอกสาร
    if (screenData.creatorcode == null || screenData.creatorcode!.isEmpty) {
      screenData.creatorcode = creatorCode;
      screenData.creatorname = creatorName;
      AppLogger.debug('[PRSaveHelper] ตั้งค่าผู้สร้างเอกสาร: $creatorCode / $creatorName');
    }

    AppLogger.debug('[PRSaveHelper] เตรียมข้อมูลก่อนบันทึกเรียบร้อย');
  }

  /// แปลงวันที่หัวเอกสารเป็น UTC
  static void _convertHeaderDatesToUtc(TransactionModel screenData) {
    DateTime docDatetimeUtc = _safeParseDatetime(screenData.docdatetime);
    screenData.docdatetime = docDatetimeUtc.toUtc().toIso8601String();

    DateTime docRefDatetimeUtc = _safeParseDatetime(screenData.docrefdate);
    screenData.docrefdate = docRefDatetimeUtc.toUtc().toIso8601String();

    DateTime taxDocDatetimeUtc = _safeParseDatetime(screenData.taxdocdate);
    screenData.taxdocdate = taxDocDatetimeUtc.toUtc().toIso8601String();
  }

  /// เตรียมข้อมูลรายการสินค้า
  static void _prepareDetailsData(TransactionModel screenData) {
    if (screenData.details == null) return;

    for (int i = 0; i < screenData.details!.length; i++) {
      final data = screenData.details![i];
      data.linenumber = i + 1;
      data.docdatetime = screenData.docdatetime;

      DateTime docrefdatetimeDetailUtc = _safeParseDatetime(data.docrefdatetime);
      data.docrefdatetime = docrefdatetimeDetailUtc.toUtc().toIso8601String();

      if (data.itemtype == 0 && data.refbarcodes != null && data.refbarcodes!.isNotEmpty) {
        for (var element in data.refbarcodes!) {
          data.standvalue = element.standvalue;
          data.dividevalue = element.dividevalue;
        }
      }
    }
  }

  /// กำหนด created_at/modified_at
  static void _setTimestamps(TransactionModel screenData) {
    final nowUtc = DateTime.now().toUtc().toIso8601String();
    if (screenData.createdat == null || screenData.createdat!.isEmpty) {
      screenData.createdat = nowUtc;
    }
    screenData.modifiedat = nowUtc;
  }

  /// ตรวจสอบ timezone fields
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
    AppLogger.info('[PRSaveHelper] เติม timezone fields: docdatelocal=${screenData.docdatelocal}, timezone=${screenData.timezone}');
  }

  /// Parse date อย่างปลอดภัย
  static DateTime _safeParseDatetime(String? dateString) {
    if (dateString == null || dateString.isEmpty) {
      return DateTime.now();
    }
    try {
      return DateTime.parse(dateString);
    } catch (e) {
      AppLogger.warning('[PRSaveHelper] แปลงวันที่ไม่สำเร็จ: "$dateString" → ใช้ DateTime.now() แทน');
      return DateTime.now();
    }
  }
}
