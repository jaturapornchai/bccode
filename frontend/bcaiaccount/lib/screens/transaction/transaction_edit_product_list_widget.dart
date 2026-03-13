import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;

class TransactionEditProductListWidget extends StatelessWidget {
  final TransactionModel screenData;

  const TransactionEditProductListWidget({super.key, required this.screenData});

  @override
  Widget build(BuildContext context) {
    if (screenData.details == null || screenData.details!.isEmpty) {
      return const SizedBox();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 10),
        _buildHeader(),
        _buildTable(),
        _buildFooter(), // แถวยอดรวมด้านล่างตาราง
      ],
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [global.theme.infoHighlightTextColor, global.theme.infoHighlightTextColor],
        ),
        borderRadius: const BorderRadius.only(
          topLeft: Radius.circular(12),
          topRight: Radius.circular(12),
        ),
      ),
      child: Row(
        children: [
          Icon(Icons.list_alt, color: global.theme.onPrimaryColor, size: 20),
          SizedBox(width: 8),
          Text(
            global.language('product_list'),
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: global.theme.onPrimaryColor,
            ),
          ),
          Spacer(),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: global.theme.cardColor.withValues(alpha: 0.5),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Text(
              '${screenData.details!.length} ${global.language('items')}',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: global.theme.onPrimaryColor,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTable() {
    return Container(
      decoration: BoxDecoration(
        borderRadius: const BorderRadius.only(
          bottomLeft: Radius.circular(12),
          bottomRight: Radius.circular(12),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Table(
        columnWidths: const {
          0: FlexColumnWidth(3), // Line
          1: FlexColumnWidth(6), // Barcode
          2: FlexColumnWidth(6), // Item Code
          3: FlexColumnWidth(14), // Name (เพิ่มความกว้างเพราะรวม remark)
          4: FlexColumnWidth(8), // Warehouse
          5: FlexColumnWidth(6), // Location
          6: FlexColumnWidth(4), // Unit
          7: FlexColumnWidth(5), // Qty
          8: FlexColumnWidth(6), // Price
          9: FlexColumnWidth(5), // Discount
          10: FlexColumnWidth(6), // Discount Amount
          11: FlexColumnWidth(6), // Amount
        },
        children: [
          _buildTableHeader(),
          ...screenData.details!.asMap().entries.map((entry) {
            return _buildTableRow(entry.key, entry.value);
          }),
        ],
      ),
    );
  }

  TableRow _buildTableHeader() {
    return TableRow(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [global.theme.primaryColor, global.theme.primaryColor.withValues(alpha: 0.7)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
      ),
      children: [
        _headerCell(global.language('line_number'), alignment: TextAlign.right),
        _headerCell(global.language('barcode')),
        _headerCell(global.language('item_code')),
        _headerCell(global.language('name')),
        _headerCell(global.language('warehouse')),
        _headerCell(global.language('location')),
        _headerCell(global.language('unit')),
        _headerCell(global.language('qty'), alignment: TextAlign.right),
        _headerCell(global.language('price'), alignment: TextAlign.right),
        _headerCell(global.language('discount'), alignment: TextAlign.right),
        _headerCell(
          global.language('discount_amount'),
          alignment: TextAlign.right,
        ),
        _headerCell(global.language('amount'), alignment: TextAlign.right),
      ],
    );
  }

  Widget _headerCell(String text, {TextAlign alignment = TextAlign.left}) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8.0, vertical: 4.0),
      child: Text(
        text,
        style: TextStyle(
          color: global.theme.onPrimaryColor,
          fontWeight: FontWeight.w600,
          fontSize: 11,
        ),
        textAlign: alignment,
        overflow: TextOverflow.ellipsis,
      ),
    );
  }

  TableRow _buildTableRow(int index, TransactionDetailModel item) {
    final isEven = index % 2 == 0;
    final itemName = _getName(item.itemnames);
    final whName = _getName(item.whnames);
    final locationName = _getName(item.locationnames);
    final sku = item.itemcode;
    final discount = item.discount.isNotEmpty ? item.discount : '-';
    final discountAmount = item.discountamount > 0
        ? global.formatNumber(item.discountamount)
        : '-';
    // ตรวจสอบว่ามีรายละเอียดเพิ่มเติมหรือไม่ (description หรือ remark)
    final hasDescription = item.description != null && item.description!.isNotEmpty;
    final hasRemark = item.remark.isNotEmpty;
    final additionalInfo = hasDescription ? item.description : (hasRemark ? item.remark : null);

    final isMulti = _isMultiCurrency();
    final rate = screenData.exchangerate ?? 1.0;

    // ราคาในสกุลเงินเอกสาร (doc currency)
    final docPrice = isMulti ? (item.priceDoc ?? (rate > 0 ? item.price / rate : 0)) : item.price;
    final docAmount = isMulti ? (item.sumAmountDoc ?? (rate > 0 ? item.sumamount / rate : 0)) : item.sumamount;

    return TableRow(
      decoration: BoxDecoration(
        color: isEven ? global.theme.surfaceColor : global.theme.onPrimaryColor,
        border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1)),
      ),
      children: [
        _dataCell(index + 1, alignment: TextAlign.right),
        _dataCell(item.barcode),
        _dataCell(sku),
        // ชื่อสินค้า พร้อมรายละเอียดเพิ่มเติม (ถ้ามี)
        _buildNameWithRemarkCell(itemName, additionalInfo),
        _dataCell(whName),
        _dataCell(locationName),
        _dataCell(item.unitcode),
        _dataCell(
          global.formatQuantity(item.qty),
          alignment: TextAlign.right,
          fontWeight: FontWeight.w500,
        ),
        // ราคา — แสดง 2 สกุลเงิน ถ้า multi-currency
        isMulti
            ? _buildDualCurrencyCell(docPrice, item.price, formatter: global.formatUnitPrice)
            : _dataCell(global.formatUnitPrice(item.price), alignment: TextAlign.right),
        _dataCell(
          discount,
          alignment: TextAlign.right,
          color: discount != '-' ? global.theme.negativeHighlightTextColor : null,
          fontWeight: discount != '-' ? FontWeight.w500 : null,
        ),
        _dataCell(
          discountAmount,
          alignment: TextAlign.right,
          color: discountAmount != '-' ? global.theme.negativeHighlightTextColor : null,
        ),
        // ยอดรวม — แสดง 2 สกุลเงิน ถ้า multi-currency
        isMulti
            ? _buildDualCurrencyCell(docAmount, item.sumamount, isBold: true)
            : _dataCell(
                global.formatNumber(item.sumamount),
                alignment: TextAlign.right,
                fontWeight: FontWeight.w600,
                color: global.theme.positiveHighlightTextColor,
              ),
      ],
    );
  }

  /// สร้าง cell แสดง 2 สกุลเงิน: doc currency (น้ำเงิน) + base currency (เทา)
  Widget _buildDualCurrencyCell(double docValue, double baseValue, {bool isBold = false, String Function(double)? formatter}) {
    final fmt = formatter ?? global.formatNumber;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 3),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.end,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            '${_getDocCurrencySymbol()} ${fmt(docValue)}',
            style: TextStyle(
              fontSize: 11,
              fontWeight: isBold ? FontWeight.w600 : FontWeight.normal,
              color: global.theme.infoHighlightTextColor,
            ),
            textAlign: TextAlign.right,
          ),
          const SizedBox(height: 1),
          Text(
            '${_getBaseCurrencySymbol()} ${fmt(baseValue)}',
            style: TextStyle(
              fontSize: 9,
              color: global.theme.textSecondaryColor,
              fontStyle: FontStyle.italic,
            ),
            textAlign: TextAlign.right,
          ),
        ],
      ),
    );
  }

  /// สร้าง cell สำหรับชื่อสินค้าพร้อมรายละเอียดเพิ่มเติม (ถ้ามี)
  Widget _buildNameWithRemarkCell(String itemName, String? remark) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 3),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          // ชื่อสินค้า
          Text(
            itemName,
            style: TextStyle(
              color: global.theme.textColor,
              fontWeight: FontWeight.w500,
              fontSize: 11,
            ),
            overflow: TextOverflow.ellipsis,
            maxLines: 2,
          ),
          // รายละเอียดเพิ่มเติม (ถ้ามี)
          if (remark != null && remark.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 2),
              child: Text(
                '📋 $remark',
                style: TextStyle(
                  color: global.theme.infoHighlightTextColor,
                  fontSize: 10,
                  fontStyle: FontStyle.italic,
                ),
                overflow: TextOverflow.ellipsis,
                maxLines: 2,
              ),
            ),
        ],
      ),
    );
  }

  Widget _dataCell(
    dynamic content, {
    TextAlign alignment = TextAlign.left,
    Color? color,
    FontWeight? fontWeight,
    double? fontSize,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 3),
      child: content is Widget
          ? content
          : Text(
              content.toString(),
              style: TextStyle(
                color: color ?? global.theme.textColor,
                fontWeight: fontWeight,
                fontSize: fontSize ?? 11,
              ),
              textAlign: alignment,
              overflow: TextOverflow.ellipsis,
              maxLines: 2,
            ),
    );
  }

  String _getName(List<LanguageDataModel>? names) {
    if (names == null || names.isEmpty) return '';
    final userLang = global.userLanguage;
    final nameObj = names.firstWhere(
      (element) => element.code == userLang,
      orElse: () => names.first,
    );
    return nameObj.name;
  }

  /// ตรวจสอบว่าเป็นเอกสาร Multi-Currency หรือไม่
  bool _isMultiCurrency() {
    return screenData.docCurrency != null &&
        screenData.docCurrency!.isNotEmpty &&
        screenData.currency != null &&
        screenData.currency!.isNotEmpty &&
        screenData.docCurrency != screenData.currency &&
        (screenData.exchangerate ?? 1.0) != 1.0;
  }

  /// ดึงสัญลักษณ์สกุลเงินเอกสาร
  String _getDocCurrencySymbol() {
    return screenData.docCurrencySymbol ?? screenData.docCurrency ?? '';
  }

  /// ดึงสัญลักษณ์สกุลเงินหลัก
  String _getBaseCurrencySymbol() {
    return screenData.currencysymbol ?? screenData.currency ?? '฿';
  }

  /// สร้างแถวยอดรวมด้านล่างตาราง
  Widget _buildFooter() {
    // คำนวณยอดรวม
    double totalQty = 0;
    double totalAmount = 0;
    double totalDiscountAmount = 0;
    double totalAmountDoc = 0;

    final isMulti = _isMultiCurrency();
    final rate = screenData.exchangerate ?? 1.0;

    for (var item in screenData.details!) {
      totalQty += item.qty;
      totalAmount += item.sumamount;
      totalDiscountAmount += item.discountamount;
      if (isMulti) {
        totalAmountDoc += (item.sumAmountDoc ?? (rate > 0 ? item.sumamount / rate : 0));
      }
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [global.theme.infoHighlightTextColor, global.theme.infoHighlightTextColor.withValues(alpha: 0.8)],
        ),
        borderRadius: const BorderRadius.only(
          bottomLeft: Radius.circular(12),
          bottomRight: Radius.circular(12),
        ),
        boxShadow: [
          BoxShadow(
            color: global.theme.infoHighlightTextColor.withValues(alpha: 0.3),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          // จำนวนรายการ
          Icon(Icons.inventory_2, color: global.theme.onPrimaryColor, size: 18),
          SizedBox(width: 6),
          Text(
            '${global.language('total')} ${screenData.details!.length} ${global.language('items')}',
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.bold,
              color: global.theme.onPrimaryColor,
            ),
          ),
          const Spacer(),
          // รวมจำนวน
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: global.theme.cardColor.withValues(alpha: 0.2),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(
              '${global.language("qty")}: ${global.formatNumber(totalQty)}',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: global.theme.onPrimaryColor,
              ),
            ),
          ),
          const SizedBox(width: 10),
          // ส่วนลด (ถ้ามี)
          if (totalDiscountAmount > 0) ...[
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: global.theme.negativeHighlightTextColor,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                '${global.language("discount")}: ${global.formatNumber(totalDiscountAmount)}',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: global.theme.onPrimaryColor,
                ),
              ),
            ),
            const SizedBox(width: 10),
          ],
          // ยอดรวม — แสดง 2 สกุลเงิน ถ้าเป็น multi-currency
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
            ),
            child: isMulti
                ? Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        '${_getDocCurrencySymbol()} ${global.formatNumber(totalAmountDoc)}',
                        style: TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                          color: global.theme.infoHighlightTextColor,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        '${_getBaseCurrencySymbol()} ${global.formatNumber(totalAmount)}',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: global.theme.infoHighlightTextColor,
                        ),
                      ),
                    ],
                  )
                : Text(
                    '${global.language("total_amount")}: ${global.formatNumber(totalAmount)}',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.bold,
                      color: global.theme.infoHighlightTextColor,
                    ),
                  ),
          ),
        ],
      ),
    );
  }
}
