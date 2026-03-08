import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/global.dart' as global;

class AdvancePaymentWidget extends StatefulWidget {
  final TransactionModel screenData;
  final Function(int index) deleteItemDetail;
  final Function() calTotalValue;
  final Function(void Function()) setState;

  const AdvancePaymentWidget({
    super.key,
    required this.screenData,
    required this.deleteItemDetail,
    required this.calTotalValue,
    required this.setState,
  });

  @override
  State<AdvancePaymentWidget> createState() => _AdvancePaymentWidgetState();
}

class _AdvancePaymentWidgetState extends State<AdvancePaymentWidget> {
  @override
  void initState() {
    super.initState();
  }

  @override
  void dispose() {
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        _buildAdvancePaymentHeader(context),
        Expanded(child: _buildAdvancePaymentDetail(context)),
      ],
    );
  }

  Widget _buildAdvancePaymentHeader(BuildContext context) {
    // ใช้ MediaQuery เพียงครั้งเดียว
    final isDesktop = MediaQuery.of(context).size.width > 799;

    if (!isDesktop) {
      return const SizedBox.shrink(); // มีประสิทธิภาพกว่า Container()
    }

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.grey[100],
        border: Border.all(color: Colors.grey[300]!),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Table(
        columnWidths: const {
          0: FixedColumnWidth(50.0), // ลบ
          1: FixedColumnWidth(70.0), // ลำดับ
          2: FlexColumnWidth(3.0), // รายละเอียด
          3: FixedColumnWidth(150.0), // จำนวนเงิน
        },
        children: [
          TableRow(
            decoration: BoxDecoration(
              color: const Color(0xFF2A6F97),
              borderRadius: BorderRadius.circular(4),
            ),
            children: [
              _buildHeaderCell(global.language('delete'), '${global.language('delete')} header'),
              _buildHeaderCell(global.language('item_order'), '${global.language('item_order')} header'),
              _buildHeaderCell(global.language('description'), '${global.language('description')} header'),
              _buildHeaderCell(global.language('amount'), '${global.language('amount')} header'),
            ],
          ),
        ],
      ),
    );
  }

  // Helper method สำหรับสร้าง header cells
  Widget _buildHeaderCell(String text, String semanticLabel) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFF2A6F97),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Semantics(
        label: semanticLabel,
        header: true,
        child: Text(
          text,
          textAlign: TextAlign.center,
          style: const TextStyle(
            fontSize: 14,
            // fontWeight: FontWeight.w600,
            color: Colors.white,
          ),
          overflow: TextOverflow.ellipsis,
        ),
      ),
    );
  }

  Widget _buildAdvancePaymentDetail(BuildContext context) {
    final isDesktop = MediaQuery.of(context).size.width > 799;
    final itemCount = widget.screenData.details?.length ?? 0;

    // แสดง empty state ถ้าไม่มีข้อมูล
    if (itemCount == 0) {
      return _buildEmptyState();
    }

    if (isDesktop) {
      return _buildDesktopView(itemCount);
    } else {
      return _buildMobileView(itemCount);
    }
  }

  Widget _buildEmptyState() {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 8),
      padding: const EdgeInsets.all(32),
      decoration: BoxDecoration(
        border: Border.all(color: Colors.grey[300]!),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.payment_outlined, size: 48, color: Colors.grey[400]),
            SizedBox(height: 16),
            Text(
              global.language('no_items_yet'),
              style: TextStyle(
                fontSize: 16,
                color: Colors.grey[600],
                fontWeight: FontWeight.w500,
              ),
            ),
            SizedBox(height: 8),
            Text(
              global.language('press_add_item_to_start'),
              style: TextStyle(fontSize: 14, color: Colors.grey[500]),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDesktopView(int itemCount) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 8),
      decoration: BoxDecoration(
        border: Border.all(color: Colors.grey[300]!),
        borderRadius: BorderRadius.circular(4),
      ),
      child: ListView.builder(
        itemCount: itemCount,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        itemBuilder: (context, index) => _buildDesktopTableRow(index),
      ),
    );
  }

  Widget _buildMobileView(int itemCount) {
    return SingleChildScrollView(
      child: ListView.builder(
        itemCount: itemCount,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        itemBuilder: (context, index) => _buildMobileCard(index),
      ),
    );
  }

  Widget _buildDesktopTableRow(int index) {
    final isEven = index % 2 == 0;

    return RepaintBoundary(
      key: ValueKey('advance_payment_row_$index'),
      child: Container(
        decoration: BoxDecoration(
          color: isEven ? Colors.white : Colors.grey[50],
          border: Border(
            bottom: BorderSide(color: Colors.grey[200]!, width: 0.5),
          ),
        ),
        child: Table(
          columnWidths: const {
            0: FixedColumnWidth(50.0), // ลบ
            1: FixedColumnWidth(70.0), // ลำดับ
            2: FlexColumnWidth(3.0), // รายละเอียด
            3: FixedColumnWidth(150.0), // จำนวนเงิน
          },
          children: [
            TableRow(
              children: [
                _buildDeleteButton(index),
                _buildOrderNumber(index),
                _buildDescriptionField(index),
                _buildAmountField(index),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDeleteButton(int index) {
    return Container(
      height: 50,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
      alignment: Alignment.center,
      child: Container(
        width: 32,
        height: 32,
        decoration: BoxDecoration(
          color: Colors.red[50],
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: Colors.red[200]!),
          boxShadow: [
            BoxShadow(
              color: Colors.red.withValues(alpha: 0.1),
              blurRadius: 2,
              offset: const Offset(0, 1),
            ),
          ],
        ),
        child: IconButton(
          padding: EdgeInsets.zero,
          onPressed: () => widget.deleteItemDetail(index),
          icon: Icon(Icons.delete_outline, color: Colors.red[600], size: 16),
          tooltip: global.language('delete_item'),
          hoverColor: Colors.red[100],
          splashRadius: 16,
        ),
      ),
    );
  }

  Widget _buildOrderNumber(int index) {
    return Container(
      height: 50,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
      alignment: Alignment.center,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        child: Text(
          '${index + 1}',
          textAlign: TextAlign.center,
          style: TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.bold,
            color: const Color(0xFF2A6F97),
          ),
        ),
      ),
    );
  }

  Widget _buildDescriptionField(int index) {
    final detail = widget.screenData.details![index];
    final description = detail.description ?? '';

    return Container(
      height: 50,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
      alignment: Alignment.center,
      child: TextFormField(
        key: ValueKey('description_${index}_$description'),
        initialValue: description,
        style: const TextStyle(fontSize: 12),
        textAlignVertical: TextAlignVertical.center,
        decoration: _buildInputDecoration(
          hintText: global.language('specify_details'),
          prefixIcon: Icons.description_outlined,
          iconColor: const Color(0xFF2A6F97),
          focusColor: const Color(0xFF2A6F97).withOpacity(0.7),
        ),
        onChanged: (value) {
          widget.screenData.details![index].description = value;
        },
      ),
    );
  }

  Widget _buildAmountField(int index) {
    final detail = widget.screenData.details![index];
    final amount = _formatAmount(detail.sumamount);

    return Container(
      height: 50,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
      alignment: Alignment.center,
      child: TextFormField(
        key: ValueKey('amount_${index}_${detail.sumamount}'),
        initialValue: amount,
        style: const TextStyle(fontSize: 12),
        textAlign: TextAlign.right,
        textAlignVertical: TextAlignVertical.center,
        keyboardType: const TextInputType.numberWithOptions(decimal: true),
        inputFormatters: [_NumberTextInputFormatter()],
        decoration: _buildInputDecoration(
          hintText: '0.00',
          prefixIcon: Icons.attach_money,
          iconColor: Colors.green[600]!,
          focusColor: Colors.green[400]!,
        ),
        onChanged: (value) {
          _handleAmountChange(index, value);
        },
      ),
    );
  }

  InputDecoration _buildInputDecoration({
    required String hintText,
    required IconData prefixIcon,
    required Color iconColor,
    required Color focusColor,
  }) {
    return InputDecoration(
      isDense: true,
      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(6),
        borderSide: BorderSide(color: Colors.grey[300]!),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(6),
        borderSide: BorderSide(color: Colors.grey[300]!),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(6),
        borderSide: BorderSide(color: focusColor, width: 2),
      ),
      hintText: hintText,
      hintStyle: TextStyle(color: Colors.grey[400]),
      filled: true,
      fillColor: Colors.white,
      prefixIcon: Icon(prefixIcon, size: 16, color: iconColor),
    );
  }

  void _handleAmountChange(int index, String value) {
    final cleanValue = value.replaceAll(',', '');
    final amount = double.tryParse(cleanValue) ?? 0.0;

    widget.screenData.details![index].qty = 1;
    widget.screenData.details![index].price = amount;
    widget.screenData.details![index].sumamount = amount;

    widget.calTotalValue();
    // ลบ setState() เพื่อไม่ให้ rebuild ขณะที่ user กำลังพิมพ์
    // widget.setState(() {});
  }

  Widget _buildMobileCard(int index) {
    return RepaintBoundary(
      key: ValueKey('mobile_advance_payment_card_$index'),
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.grey[300]!),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.05),
              blurRadius: 4,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildMobileHeader(index),
              const SizedBox(height: 12),
              _buildMobileDescriptionField(index),
              const SizedBox(height: 12),
              _buildMobileAmountField(index),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildMobileHeader(int index) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Container(
          padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          child: Text(
            '${global.language('item_order')} ${index + 1}',
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.bold,
              color: const Color(0xFF2A6F97),
            ),
          ),
        ),
        Container(
          width: 32,
          height: 32,
          decoration: BoxDecoration(
            color: Colors.red[50],
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: Colors.red[200]!),
          ),
          child: IconButton(
            padding: EdgeInsets.zero,
            onPressed: () => widget.deleteItemDetail(index),
            icon: Icon(Icons.delete_outline, color: Colors.red[600], size: 16),
            tooltip: global.language('delete_item'),
          ),
        ),
      ],
    );
  }

  Widget _buildMobileDescriptionField(int index) {
    final detail = widget.screenData.details![index];
    final description = detail.description ?? '';

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          global.language('description'),
          style: TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.w600,
            color: Colors.grey[700],
          ),
        ),
        SizedBox(height: 6),
        TextFormField(
          key: ValueKey('mobile_description_${index}_$description'),
          initialValue: description,
          style: TextStyle(fontSize: 14),
          decoration: _buildMobileInputDecoration(
            hintText: global.language('specify_details'),
            prefixIcon: Icons.description_outlined,
            iconColor: const Color(0xFF2A6F97),
            focusColor: const Color(0xFF2A6F97).withOpacity(0.7),
          ),
          onChanged: (value) {
            widget.screenData.details![index].description = value;
          },
        ),
      ],
    );
  }

  Widget _buildMobileAmountField(int index) {
    final detail = widget.screenData.details![index];
    final amount = _formatAmount(detail.sumamount);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          global.language('amount'),
          style: TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.w600,
            color: Colors.grey[700],
          ),
        ),
        const SizedBox(height: 6),
        TextFormField(
          key: ValueKey('mobile_amount_${index}_${detail.sumamount}'),
          initialValue: amount,
          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
          textAlign: TextAlign.right,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          inputFormatters: [_NumberTextInputFormatter()],
          decoration: _buildMobileInputDecoration(
            hintText: '0.00',
            prefixIcon: Icons.attach_money,
            iconColor: Colors.green[600]!,
            focusColor: Colors.green[400]!,
          ),
          onChanged: (value) {
            _handleAmountChange(index, value);
          },
        ),
      ],
    );
  }

  InputDecoration _buildMobileInputDecoration({
    required String hintText,
    required IconData prefixIcon,
    required Color iconColor,
    required Color focusColor,
  }) {
    return InputDecoration(
      isDense: true,
      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(6),
        borderSide: BorderSide(color: Colors.grey[300]!),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(6),
        borderSide: BorderSide(color: Colors.grey[300]!),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(6),
        borderSide: BorderSide(color: focusColor, width: 2),
      ),
      hintText: hintText,
      hintStyle: TextStyle(color: Colors.grey[400]),
      filled: true,
      fillColor: Colors.grey[50],
      prefixIcon: Icon(prefixIcon, size: 18, color: iconColor),
    );
  }

  // Helper method สำหรับจัดรูปแบบจำนวนเงิน - optimized version
  String _formatAmount(double amount) {
    if (amount == 0) return '';
    return NumberFormat('#,##0.00').format(amount);
  }
}

// Optimized TextInputFormatter สำหรับจัดรูปแบบจำนวนเงิน
class _NumberTextInputFormatter extends TextInputFormatter {
  static final RegExp _numberRegex = RegExp(r'[^0-9.]');
  static final NumberFormat _formatter = NumberFormat('#,##0');

  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    // อนุญาตเฉพาะตัวเลขและจุดทศนิยม
    String newText = newValue.text.replaceAll(_numberRegex, '');

    // ตรวจสอบจุดทศนิยม (อนุญาตเฉพาะ 1 จุด)
    final dotCount = '.'.allMatches(newText).length;
    if (dotCount > 1) {
      final firstDotIndex = newText.indexOf('.');
      newText =
          newText.substring(0, firstDotIndex + 1) +
          newText.substring(firstDotIndex + 1).replaceAll('.', '');
    }

    // จำกัดทศนิยมไม่เกิน 2 ตำแหน่ง
    final parts = newText.split('.');
    if (parts.length == 2 && parts[1].length > 2) {
      newText = '${parts[0]}.${parts[1].substring(0, 2)}';
    }

    // จัดรูปแบบด้วยเครื่องหมายจุลภาค
    final formattedText = _formatNumber(newText);

    return TextEditingValue(
      text: formattedText,
      selection: TextSelection.collapsed(offset: formattedText.length),
    );
  }

  String _formatNumber(String text) {
    if (text.isEmpty) return text;

    // แยกส่วนจำนวนเต็มและทศนิยม
    final parts = text.split('.');
    final integerPart = parts[0];
    final decimalPart = parts.length > 1 ? parts[1] : '';

    // จัดรูปแบบส่วนจำนวนเต็มด้วยเครื่องหมายจุลภาค
    String formattedInteger = '';
    if (integerPart.isNotEmpty) {
      final number = int.tryParse(integerPart);
      if (number != null) {
        formattedInteger = _formatter.format(number);
      } else {
        formattedInteger = integerPart;
      }
    }

    // รวมส่วนจำนวนเต็มและทศนิยม
    if (decimalPart.isNotEmpty) {
      return '$formattedInteger.$decimalPart';
    } else if (text.endsWith('.')) {
      return '$formattedInteger.';
    } else {
      return formattedInteger;
    }
  }
}
