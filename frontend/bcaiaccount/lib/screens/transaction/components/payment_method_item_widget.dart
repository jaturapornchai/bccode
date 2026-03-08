// ignore_for_file: deprecated_member_use

import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/date_picker.dart';

/// ประเภทของการชำระเงิน
enum PaymentMethodType {
  creditCard, // บัตรเครดิต
  transfer, // โอนเงิน
  cheque, // เช็ค
  coupon, // คูปอง
  qrCode, // QR Code
}

/// Widget สำหรับแสดงรายการการชำระเงินแต่ละประเภท
/// รองรับทุกประเภทการชำระเงิน โดยปรับฟิลด์ตามประเภท
class PaymentMethodItemWidget extends StatelessWidget {
  final PaymentMethodType paymentType;
  final int itemIndex;
  final BillPayStruct paymentData;
  final TextEditingController amountController;
  final VoidCallback onDelete;
  final Function(double) onAmountChanged;
  final Function()? onBankSearch;
  final Function(DateTime?)? onDateSelected;
  final Function(DateTime?)? onDueDateSelected; // สำหรับ cheque

  const PaymentMethodItemWidget({
    super.key,
    required this.paymentType,
    required this.itemIndex,
    required this.paymentData,
    required this.amountController,
    required this.onDelete,
    required this.onAmountChanged,
    this.onBankSearch,
    this.onDateSelected,
    this.onDueDateSelected,
  });

  /// ดึงชื่อการชำระเงินตาม type
  String _getPaymentMethodName() {
    switch (paymentType) {
      case PaymentMethodType.creditCard:
        return global.language('list_creditcard');
      case PaymentMethodType.transfer:
        return global.language('list_transfer');
      case PaymentMethodType.cheque:
        return global.language('list_cheque');
      case PaymentMethodType.coupon:
        return global.language('list_coupon');
      case PaymentMethodType.qrCode:
        return global.language('list_qr');
    }
  }

  /// ตรวจสอบว่าต้องแสดงส่วนของธนาคารหรือไม่
  bool _requiresBankSection() {
    return paymentType == PaymentMethodType.creditCard ||
        paymentType == PaymentMethodType.transfer ||
        paymentType == PaymentMethodType.cheque;
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Container(
        padding: const EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildHeader(),
            const SizedBox(height: 10),
            _buildDateSection(),
            if (_requiresBankSection()) ...[
              const SizedBox(height: 10),
              _buildBankSection(),
            ],
            if (paymentType == PaymentMethodType.creditCard) ...[
              const SizedBox(height: 10),
              _buildCardNumberField(),
            ],
            if (paymentType == PaymentMethodType.cheque) ...[
              const SizedBox(height: 10),
              _buildChequeFields(),
            ],
            if (paymentType == PaymentMethodType.coupon) ...[
              const SizedBox(height: 10),
              _buildCouponFields(),
            ],
            if (paymentType == PaymentMethodType.qrCode) ...[
              const SizedBox(height: 10),
              _buildQrCodeFields(),
            ],
            const SizedBox(height: 10),
            _buildAmountSection(),
          ],
        ),
      ),
    );
  }

  /// สร้าง Header พร้อมปุ่มลบ
  Widget _buildHeader() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          '${_getPaymentMethodName()} ${itemIndex + 1}',
          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
        ),
        IconButton(
          onPressed: onDelete,
          icon: Icon(Icons.delete, color: Colors.red),
          tooltip: global.language('delete'),
        ),
      ],
    );
  }

  /// สร้างส่วนเลือกวันที่
  Widget _buildDateSection() {
    return Row(
      children: [
        Expanded(
          child: CustomDatePicker(
            key: ValueKey(paymentData.doc_date_time),
            labelText: global.language('doc_date'),
            initialDate: paymentData.doc_date_time ?? DateTime.now(),
            // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
            onDateSelected: onDateSelected,
            decoration: const InputDecoration(
              labelText: '',
              border: OutlineInputBorder(),
            ),
          ),
        ),
      ],
    );
  }

  /// สร้างส่วนธนาคาร (สำหรับ Credit Card, Transfer, Cheque)
  Widget _buildBankSection() {
    return Row(
      children: [
        Expanded(
          child: TextField(
            readOnly: true,
            decoration: InputDecoration(
              labelText: global.language('book_bank'),
              border: const OutlineInputBorder(),
              suffixIcon: IconButton(
                icon: const Icon(Icons.search),
                onPressed: onBankSearch,
              ),
            ),
            controller: TextEditingController(
              text: paymentData.book_bank_code ?? '',
            ),
          ),
        ),
        SizedBox(width: 10),
        Expanded(
          child: TextField(
            readOnly: true,
            decoration: InputDecoration(
              labelText: global.language('bank'),
              border: const OutlineInputBorder(),
            ),
            controller: TextEditingController(
              text: paymentData.bank_name ?? '',
            ),
          ),
        ),
      ],
    );
  }

  /// สร้างฟิลด์หมายเลขบัตร (สำหรับ Credit Card)
  Widget _buildCardNumberField() {
    return TextField(
      decoration: InputDecoration(
        labelText: global.language('creditnumber'),
        border: const OutlineInputBorder(),
      ),
      controller: TextEditingController(text: paymentData.card_number ?? ''),
      onChanged: (value) {
        paymentData.card_number = value;
      },
    );
  }

  /// สร้างฟิลด์สำหรับเช็ค (Cheque Number + Due Date)
  Widget _buildChequeFields() {
    return Column(
      children: [
        TextField(
          decoration: InputDecoration(
            labelText: global.language('cheque_number'),
            border: const OutlineInputBorder(),
          ),
          controller: TextEditingController(
            text: paymentData.cheque_number ?? '',
          ),
          onChanged: (value) {
            paymentData.cheque_number = value;
          },
        ),
        SizedBox(height: 10),
        CustomDatePicker(
          key: ValueKey(paymentData.due_date),
          labelText: global.language('due_date'),
          initialDate: paymentData.due_date ?? DateTime.now(),
          // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
          onDateSelected: onDueDateSelected,
          decoration: const InputDecoration(
            labelText: '',
            border: OutlineInputBorder(),
          ),
        ),
      ],
    );
  }

  /// สร้างฟิลด์สำหรับคูปอง (Coupon Number + Description)
  Widget _buildCouponFields() {
    return Column(
      children: [
        TextField(
          decoration: InputDecoration(
            labelText: global.language('coupon_number'),
            border: const OutlineInputBorder(),
          ),
          controller: TextEditingController(text: paymentData.number ?? ''),
          onChanged: (value) {
            paymentData.number = value;
          },
        ),
        SizedBox(height: 10),
        TextField(
          decoration: InputDecoration(
            labelText: global.language('description'),
            border: const OutlineInputBorder(),
          ),
          controller: TextEditingController(
            text: paymentData.description ?? '',
          ),
          onChanged: (value) {
            paymentData.description = value;
          },
          maxLines: 2,
        ),
      ],
    );
  }

  /// สร้างฟิลด์สำหรับ QR Code (Provider Code + Provider Name)
  Widget _buildQrCodeFields() {
    return Row(
      children: [
        Expanded(
          child: TextField(
            decoration: InputDecoration(
              labelText: global.language('provider_code'),
              border: const OutlineInputBorder(),
            ),
            controller: TextEditingController(
              text: paymentData.provider_code ?? '',
            ),
            onChanged: (value) {
              paymentData.provider_code = value;
            },
          ),
        ),
        SizedBox(width: 10),
        Expanded(
          child: TextField(
            decoration: InputDecoration(
              labelText: global.language('provider_name'),
              border: const OutlineInputBorder(),
            ),
            controller: TextEditingController(
              text: paymentData.provider_name ?? '',
            ),
            onChanged: (value) {
              paymentData.provider_name = value;
            },
          ),
        ),
      ],
    );
  }

  /// สร้างส่วนใส่จำนวนเงิน
  Widget _buildAmountSection() {
    return TextFormField(
      controller: amountController,
      decoration: InputDecoration(
        labelText: global.language('amount'),
        border: const OutlineInputBorder(),
        prefixText: '฿ ',
      ),
      keyboardType: TextInputType.number,
      inputFormatters: [global.NumberInputFormatter()],
      onChanged: (value) {
        final amount = global.parseTextToNumber(value);
        paymentData.amount = amount;
        onAmountChanged(amount);
      },
    );
  }
}
