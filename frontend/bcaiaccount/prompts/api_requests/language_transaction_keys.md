# API Request: เพิ่ม Language Keys สำหรับ Transaction Types ใหม่

## สิ่งที่ต้องการ
เพิ่ม language keys 7 รายการลง `assets/language/languages.json` สำหรับ transaction types ที่เพิ่งเพิ่มเข้ามาใหม่ (advance payment refund, deposit, paid advance, receive deposit)

## ไฟล์ที่ต้องแก้
`d:\bcdev\backend\assets\language\languages.json`

## Keys ที่ต้องเพิ่ม

เพิ่มหลัง key `"transaction_advance_payment"` (บรรทัด ~33012):

```json
  "transaction_advance_payment_refund": {
    "cn": "Advance Payment Refund",
    "en": "Advance Payment Refund",
    "ja": "前払い返金",
    "km": "Advance Payment Refund",
    "ko": "Advance Payment Refund",
    "lo": "ຄືນເງິນລ່ວງໜ້າ",
    "my": "ကြိုတင်ငွေပေးချေမှု ပြန်အမ်း",
    "th": "รับคืนเงินล่วงหน้า",
    "vi": "Hoàn tiền đặt cọc trước"
  },
  "transaction_deposit": {
    "cn": "Deposit",
    "en": "Deposit",
    "ja": "保証金支払い",
    "km": "Deposit",
    "ko": "보증금",
    "lo": "ຈ່າຍເງິນມັດຈໍາ",
    "my": "အပ်ငွေ",
    "th": "จ่ายเงินมัดจำ",
    "vi": "Đặt cọc"
  },
  "transaction_deposit_refund": {
    "cn": "Deposit Refund",
    "en": "Deposit Refund",
    "ja": "保証金返金",
    "km": "Deposit Refund",
    "ko": "보증금 환불",
    "lo": "ຄືນເງິນມັດຈໍາ",
    "my": "အပ်ငွေ ပြန်အမ်း",
    "th": "รับคืนเงินมัดจำ",
    "vi": "Hoàn tiền đặt cọc"
  },
  "transaction_paid_advance": {
    "cn": "Receive Advance Payment",
    "en": "Receive Advance Payment",
    "ja": "前払い受取",
    "km": "Receive Advance Payment",
    "ko": "선불금 수령",
    "lo": "ຮັບເງິນລ່ວງໜ້າ",
    "my": "ကြိုတင်ငွေ လက်ခံ",
    "th": "รับเงินล่วงหน้า",
    "vi": "Nhận tiền đặt trước"
  },
  "transaction_paid_advance_refund": {
    "cn": "Advance Payment Refund",
    "en": "Advance Payment Refund",
    "ja": "前払い返金",
    "km": "Advance Payment Refund",
    "ko": "선불금 환불",
    "lo": "ຄືນເງິນລ່ວງໜ້າ",
    "my": "ကြိုတင်ငွေ ပြန်အမ်း",
    "th": "คืนเงินล่วงหน้า",
    "vi": "Hoàn tiền trả trước"
  },
  "transaction_receive_deposit": {
    "cn": "Receive Deposit",
    "en": "Receive Deposit",
    "ja": "保証金受取",
    "km": "Receive Deposit",
    "ko": "보증금 수령",
    "lo": "ຮັບເງິນມັດຈໍາ",
    "my": "အပ်ငွေ လက်ခံ",
    "th": "รับเงินมัดจำ",
    "vi": "Nhận tiền đặt cọc"
  },
  "transaction_receive_deposit_refund": {
    "cn": "Receive Deposit Refund",
    "en": "Receive Deposit Refund",
    "ja": "保証金返金受取",
    "km": "Receive Deposit Refund",
    "ko": "보증금 환불 수령",
    "lo": "ຄືນເງິນມັດຈໍາ",
    "my": "အပ်ငွေ ပြန်အမ်း လက်ခံ",
    "th": "คืนเงินมัดจำ",
    "vi": "Hoàn tiền nhận đặt cọc"
  },
```

## หมายเหตุ
- ภาษาที่สำคัญที่สุดคือ `th` (Thai) และ `en` (English)
- ภาษาอื่น (cn, ja, km, ko, lo, my, vi) สามารถใช้ทับศัพท์ภาษาอังกฤษได้ก่อน แล้วปรับเพิ่มเติมภายหลัง
- Flutter app ใน `lib/global.dart` ได้ update ให้ใช้ language keys เหล่านี้แล้ว
- ถ้า key ไม่มีใน language map จะแสดง key name แทน (เช่น `transaction_deposit`) ซึ่งดูไม่สวย

## Flutter Code ที่เปลี่ยนแล้ว (global.dart)
```dart
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
```
