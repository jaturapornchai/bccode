import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screens/purchaseorder/models/po_model_factory.dart';
import 'package:smlaicloud/global.dart' as global;

/// Factory class สำหรับสร้าง TransactionModel เปล่าสำหรับใบเสนอราคา (Quotation)
///
/// Reuse POModelFactory.createDefault() โดยส่งค่าเฉพาะของ Quotation
class QTModelFactory {
  /// สร้าง TransactionModel สำหรับ Quotation ใหม่
  ///
  /// ค่าเฉพาะ Quotation:
  /// - transflag = 30
  /// - vattype = vattypesale (ภาษีขาย)
  /// - inquirytype = 0 (ไม่มี)
  static TransactionModel createForQuotation() {
    return POModelFactory.createDefault(
      transflag: 30, // Quotation transflag
      vattype: global.config.vattypesale,
      inquirytype: 0, // Quotation ไม่มี inquirytype
      timezoneInfo: global.getLocalTimezoneInfo(),
    );
  }
}
