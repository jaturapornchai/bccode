import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/bi_report/stock_balance_model.dart';

class StockBalanceTableView extends StatefulWidget {
  final List<StockBalanceModel> stockBalances;
  final String Function(double) formatCurrency;

  const StockBalanceTableView({
    super.key,
    required this.stockBalances,
    required this.formatCurrency,
  });

  @override
  State<StockBalanceTableView> createState() => _StockBalanceTableViewState();
}

class _StockBalanceTableViewState extends State<StockBalanceTableView> with global.ThemeRefreshMixin {
  int? _hoveredIndex;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Data table
        Expanded(child: _buildDataTable()),
      ],
    );
  }

  Widget _buildDataTable() {
    if (widget.stockBalances.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.inventory_2_outlined,
              size: 64,
              color: global.theme.iconSecondaryColor,
            ),
            const SizedBox(height: 16),
            Text(
              'ไม่พบข้อมูลยอดคงเหลือสินค้า',
              style: TextStyle(
                fontSize: 16,
                color: global.theme.textSecondaryColor,
                fontWeight: FontWeight.w500,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'ลองเปลี่ยนเงื่อนไขการค้นหาใหม่',
              style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
            ),
          ],
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: global.theme.dividerBorderColor),
        boxShadow: [
          BoxShadow(
            color: global.theme.surfaceColor,
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          // Table header
          _buildTableHeader(),
          // Table data
          Expanded(
            child: ListView.builder(
              itemCount: widget.stockBalances.length,
              itemBuilder: (context, index) {
                return _buildTableRow(widget.stockBalances[index], index);
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTableHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.purple.shade50,
        borderRadius: const BorderRadius.only(
          topLeft: Radius.circular(12),
          topRight: Radius.circular(12),
        ),
        border: Border(bottom: BorderSide(color: Colors.purple.shade200)),
      ),
      child: Row(
        children: [
          Expanded(
            flex: 2,
            child: Text(
              global.language('barcode'),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.purple,
              ),
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('product_code'),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.purple,
              ),
            ),
          ),
          Expanded(
            flex: 3,
            child: Text(
              global.language('import_product_detail.product_name'),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.purple,
              ),
            ),
          ),
          Expanded(
            child: Text(
              global.language('unit'),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.purple,
              ),
              textAlign: TextAlign.center,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('balance'),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.purple,
              ),
              textAlign: TextAlign.right,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('average_cost'),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.purple,
              ),
              textAlign: TextAlign.right,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('balance_amount'),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.purple,
              ),
              textAlign: TextAlign.right,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTableRow(StockBalanceModel stockBalance, int index) {
    final isNegative = stockBalance.balanceQty < 0;
    final isHovered = _hoveredIndex == index;

    // Determine row color based on hover state and index
    Color rowColor;
    if (isHovered) {
      rowColor = Colors.purple.shade100;
    } else {
      rowColor = index.isEven ? global.theme.cardColor : global.theme.backgroundColor;
    }

    return MouseRegion(
      onEnter: (_) => setState(() => _hoveredIndex = index),
      onExit: (_) => setState(() => _hoveredIndex = null),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: rowColor,
          border:
              index < widget.stockBalances.length - 1
                  ? Border(bottom: BorderSide(color: global.theme.surfaceColor))
                  : null,
        ),
        child: Row(
          children: [
            // บาร์โค้ด
            Expanded(
              flex: 2,
              child: Text(
                stockBalance.barcode,
                style: TextStyle(
                  fontSize: 11,
                  color:
                      isHovered
                          ? Colors.purple.shade800
                          : Colors.purple.shade700,
                  fontWeight: FontWeight.w600,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),

            // รหัสสินค้า
            Expanded(
              flex: 2,
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 8),
                child: Text(
                  (stockBalance.itemcode!.isNotEmpty ||
                          stockBalance.itemcode != null)
                      ? stockBalance.itemcode!
                      : '-',
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: isHovered ? global.theme.textColor : global.theme.textSecondaryColor,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ),

            // ชื่อสินค้า
            Expanded(
              flex: 3,
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 8),
                child: Text(
                  global.activeLangName(stockBalance.names),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: isHovered ? global.theme.textColor : global.theme.textSecondaryColor,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ),

            // หน่วยนับ
            Expanded(
              child: Text(
                global.activeLangName(stockBalance.unitNames),
                style: TextStyle(
                  fontSize: 11,
                  color:
                      isHovered ? global.theme.iconColor : global.theme.textSecondaryColor,
                  fontWeight: FontWeight.w500,
                ),
                textAlign: TextAlign.center,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),

            // จำนวนคงเหลือ
            Expanded(
              flex: 2,
              child: Text(
                stockBalance.balanceQty.toStringAsFixed(2),
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color:
                      isNegative
                          ? (isHovered
                              ? global.theme.negativeHighlightTextColor
                              : global.theme.negativeHighlightTextColor)
                          : (isHovered
                              ? global.theme.positiveHighlightTextColor
                              : global.theme.positiveHighlightTextColor),
                ),
                textAlign: TextAlign.right,
              ),
            ),

            // ต้นทุนเฉลี่ย
            Expanded(
              flex: 2,
              child: Text(
                stockBalance.averageCost != null
                    ? widget.formatCurrency(stockBalance.averageCost!)
                    : '-',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w500,
                  color:
                      isHovered
                          ? global.theme.warningHighlightTextColor
                          : global.theme.warningHighlightTextColor,
                ),
                textAlign: TextAlign.right,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),

            // มูลค่าคงเหลือ
            Expanded(
              flex: 2,
              child: Text(
                stockBalance.balanceAmount != null
                    ? widget.formatCurrency(stockBalance.balanceAmount!)
                    : '-',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color:
                      stockBalance.balanceAmount != null &&
                              stockBalance.balanceAmount! < 0
                          ? (isHovered
                              ? global.theme.negativeHighlightTextColor
                              : global.theme.negativeHighlightTextColor)
                          : (isHovered
                              ? Colors.purple.shade800
                              : Colors.purple.shade700),
                ),
                textAlign: TextAlign.right,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
