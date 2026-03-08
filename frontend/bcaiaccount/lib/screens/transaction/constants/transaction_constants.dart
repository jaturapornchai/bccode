// Constants สำหรับ Transaction
// เก็บค่าคงที่ต่างๆ ที่ใช้ในระบบ Transaction เพื่อหลีกเลี่ยง magic numbers

class TransactionConstants {
  // ห้ามสร้าง instance
  TransactionConstants._();

  /// Payment Method Transaction Flags
  static const int paymentTransferFlag = 2;
  static const int paymentCreditCardFlag = 1;
  static const int paymentChequeFlag = 3;
  static const int paymentCouponFlag = 4;
  static const int paymentQrCodeFlag = 5;

  /// UI Spacing
  static const double defaultPadding = 10.0;
  static const double defaultSpacing = 5.0;
  static const double cardPadding = 10.0;

  /// Text Sizes
  static const double titleFontSize = 16.0;
  static const double normalFontSize = 14.0;

  /// Number Format
  static const int maximumFractionDigits = 2;
  static const String currencySymbol = '฿';
  static const String numberFormatPattern = '#,##0.00';

  /// Date Format
  static const int buddhistYearOffset = 543;
  static const String defaultDateFormat = 'dd/MM/yyyy';
  static const String defaultTimeFormat = 'HH:mm';

  /// Validation
  static const double minimumAmount = 0.0;
  static const int minimumQuantity = 0;

  /// Colors (ถ้าต้องการกำหนดสีเฉพาะ Transaction)
  // static const Color deleteButtonColor = Colors.red;
  // static const Color primaryActionColor = Colors.blue;
}

/// Payment Method Types Mapping
class PaymentMethodFlags {
  // ห้ามสร้าง instance
  PaymentMethodFlags._();

  /// แปลง trans_flag เป็นชื่อประเภท
  static String getFlagName(int flag) {
    switch (flag) {
      case TransactionConstants.paymentCreditCardFlag:
        return 'Credit Card';
      case TransactionConstants.paymentTransferFlag:
        return 'Transfer';
      case TransactionConstants.paymentChequeFlag:
        return 'Cheque';
      case TransactionConstants.paymentCouponFlag:
        return 'Coupon';
      case TransactionConstants.paymentQrCodeFlag:
        return 'QR Code';
      default:
        return 'Unknown';
    }
  }

  /// ตรวจสอบว่า flag นี้ต้องการข้อมูลธนาคารหรือไม่
  static bool requiresBankInfo(int flag) {
    return flag == TransactionConstants.paymentCreditCardFlag ||
        flag == TransactionConstants.paymentTransferFlag ||
        flag == TransactionConstants.paymentChequeFlag;
  }

  /// ตรวจสอบว่า flag นี้ต้องการวันครบกำหนดหรือไม่
  static bool requiresDueDate(int flag) {
    return flag == TransactionConstants.paymentChequeFlag;
  }
}
