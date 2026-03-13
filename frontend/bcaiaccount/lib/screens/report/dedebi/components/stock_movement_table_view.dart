import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/bi_report/stock_movment_model.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';

class StockMovementTableView extends StatelessWidget {
  final List<StockMovementModel> data;
  final Function(StockMovementModel)? onRowTap;

  const StockMovementTableView({
    super.key,
    required this.data,
    this.onRowTap,
  });

  @override
  Widget build(BuildContext context) {
    if (data.isEmpty) {
      return Center(
        child: Text(
          global.language('stock_movement_no_data'),
          style: TextStyle(fontSize: 16, color: global.theme.textSecondaryColor),
        ),
      );
    }

    return Column(
      children: [
        // Header Row
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            border: Border(
              bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1),
            ),
          ),
          child: _buildHeaderRow(),
        ),
        // Data Rows
        Expanded(
          child: ListView.builder(
            itemCount: data.length,
            itemBuilder: (context, index) => _buildDataRow(data[index], index),
          ),
        ),
      ],
    );
  }

  Widget _buildHeaderRow() {
    return Row(
      children: [
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_date')),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_time')),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_transaction_type'), TextAlign.center),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_document_no')),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_warehouse')),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_location')),
        ),
        Expanded(
          flex: 1,
          child: _buildHeaderCell(global.language('stock_movement_unit'), TextAlign.center),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_qty_in'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_cost_in'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_value_in'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_qty_out'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_cost_out'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_value_out'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_balance_qty'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_average_cost'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('stock_movement_balance_amount'), TextAlign.right),
        ),
      ],
    );
  }

  Widget _buildHeaderCell(String text, [TextAlign? textAlign]) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.bold,
          color: Colors.indigo.shade700,
        ),
        textAlign: textAlign ?? TextAlign.left,
      ),
    );
  }

  Widget _buildDataRow(StockMovementModel stockMovement, int index) {
    final isEven = index % 2 == 0;
    final backgroundColor = isEven ? global.theme.columnAlternateEvenColor : global.theme.columnAlternateOddColor;

    return InkWell(
      onTap: null,
      hoverColor: global.theme.rowHoverColor,
      highlightColor: Colors.indigo.shade100,
      splashColor: Colors.indigo.shade200,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: backgroundColor,
          border: Border(
            bottom: BorderSide(color: global.theme.dividerBorderColor, width: 0.5),
          ),
        ),
        child: Row(
          children: [
            // วันที่
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatDate(stockMovement.docdate),
                fontSize: 11,
              ),
            ),
            // เวลา
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.doctime.isNotEmpty ? stockMovement.doctime : '-',
                fontSize: 11,
                color: global.theme.textSecondaryColor,
              ),
            ),
            // ประเภทรายการ
            Expanded(
              flex: 2,
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 4),
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: _getTransactionFlagColor(stockMovement.transFlag),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    _getTransactionFlagText(stockMovement.transFlag),
                    style: TextStyle(
                      fontSize: 9,
                      fontWeight: FontWeight.w600,
                      color:
                          _getTransactionFlagTextColor(stockMovement.transFlag),
                    ),
                    textAlign: TextAlign.center,
                  ),
                ),
              ),
            ),
            // เลขที่เอกสาร
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.docno,
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: Colors.indigo.shade700,
              ),
            ),
            // คลัง
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.whCode.isNotEmpty ? stockMovement.whCode : '-',
                fontSize: 11,
                color: global.theme.infoHighlightTextColor,
              ),
            ),
            // ที่เก็บ
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.locationCode.isNotEmpty
                    ? stockMovement.locationCode
                    : '-',
                fontSize: 11,
                color: Colors.teal.shade700,
              ),
            ),
            // หน่วยนับ
            Expanded(
              flex: 1,
              child: _buildDataCell(
                global.activeLangName(stockMovement.mainunitnames),
                fontSize: 10,
                textAlign: TextAlign.center,
                color: Colors.purple.shade600,
              ),
            ),
            // จำนวนเพิ่ม
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.qtyIn > 0
                    ? stockMovement.qtyIn.toStringAsFixed(2)
                    : '-',
                textAlign: TextAlign.right,
                color: stockMovement.qtyIn > 0
                    ? global.theme.positiveHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight: stockMovement.qtyIn > 0
                    ? FontWeight.w600
                    : FontWeight.normal,
              ),
            ),
            // ต้นทุนเพิ่ม
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.averageCostIn > 0
                    ? ReportUtils.formatCurrencySafe(
                        stockMovement.averageCostIn)
                    : '-',
                textAlign: TextAlign.right,
                color: stockMovement.averageCostIn > 0
                    ? global.theme.positiveHighlightTextColor
                    : global.theme.textSecondaryColor,
              ),
            ),
            // มูลค่าเพิ่ม
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.balanceIn > 0
                    ? ReportUtils.formatCurrencySafe(stockMovement.balanceIn)
                    : '-',
                textAlign: TextAlign.right,
                color: stockMovement.balanceIn > 0
                    ? global.theme.positiveHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight: stockMovement.balanceIn > 0
                    ? FontWeight.w600
                    : FontWeight.normal,
              ),
            ),
            // จำนวนลด
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.qtyOut > 0
                    ? stockMovement.qtyOut.toStringAsFixed(2)
                    : '-',
                textAlign: TextAlign.right,
                color: stockMovement.qtyOut > 0
                    ? global.theme.negativeHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight: stockMovement.qtyOut > 0
                    ? FontWeight.w600
                    : FontWeight.normal,
              ),
            ),
            // ต้นทุนลด
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.averageCostOut > 0
                    ? ReportUtils.formatCurrencySafe(
                        stockMovement.averageCostOut)
                    : '-',
                textAlign: TextAlign.right,
                color: stockMovement.averageCostOut > 0
                    ? global.theme.negativeHighlightTextColor
                    : global.theme.textSecondaryColor,
              ),
            ),
            // มูลค่าลด
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.balanceOut > 0
                    ? ReportUtils.formatCurrencySafe(stockMovement.balanceOut)
                    : '-',
                textAlign: TextAlign.right,
                color: stockMovement.balanceOut > 0
                    ? global.theme.negativeHighlightTextColor
                    : global.theme.textSecondaryColor,
                fontWeight: stockMovement.balanceOut > 0
                    ? FontWeight.w600
                    : FontWeight.normal,
              ),
            ),
            // จำนวนคงเหลือ
            Expanded(
              flex: 2,
              child: _buildDataCell(
                stockMovement.balanceQty.toStringAsFixed(2),
                textAlign: TextAlign.right,
                fontWeight: FontWeight.w600,
                color: Colors.indigo.shade700,
              ),
            ),
            // ต้นทุนเฉลี่ย
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(stockMovement.averageCost),
                textAlign: TextAlign.right,
                color: global.theme.warningHighlightTextColor,
                fontWeight: FontWeight.w500,
              ),
            ),
            // มูลค่าคงเหลือ
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrencySafe(stockMovement.balanceAmount),
                textAlign: TextAlign.right,
                fontWeight: FontWeight.bold,
                color: Colors.purple.shade700,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDataCell(
    String text, {
    TextAlign? textAlign,
    double? fontSize,
    FontWeight? fontWeight,
    Color? color,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Text(
        text,
        style: TextStyle(
          fontSize: fontSize ?? 11,
          fontWeight: fontWeight,
          color: color,
        ),
        textAlign: textAlign,
      ),
    );
  }

  // Helper methods for Stock Movement
  Color _getTransactionFlagColor(String transFlag) {
    final greenFlags = {
      global.language('transaction_purchase'),
      global.language('stock_movement_receive_partial'),
      global.language('good_return'),
      global.language('stock_movement_transfer_in'),
      global.language('stock_movement_opening_balance'),
      global.language('stock return product'),
      global.language('transaction_stock_receive_product'),
      global.language('stock_movement_adjust_increase'),
    };
    final redFlags = {
      global.language('pdf_return_goods'),
      global.language('transaction_sale'),
      global.language('stock_movement_transfer_out'),
      global.language('transaction_stock_pick_up_product'),
      global.language('stock_movement_adjust_decrease'),
    };
    if (greenFlags.contains(transFlag)) return global.theme.positiveHighlightColor;
    if (redFlags.contains(transFlag)) return global.theme.negativeHighlightColor;
    return global.theme.surfaceColor;
  }

  Color _getTransactionFlagTextColor(String transFlag) {
    final greenFlags = {
      global.language('transaction_purchase'),
      global.language('stock_movement_receive_partial'),
      global.language('good_return'),
      global.language('stock_movement_transfer_in'),
      global.language('stock_movement_opening_balance'),
      global.language('stock return product'),
      global.language('transaction_stock_receive_product'),
      global.language('stock_movement_adjust_increase'),
    };
    final redFlags = {
      global.language('pdf_return_goods'),
      global.language('transaction_sale'),
      global.language('stock_movement_transfer_out'),
      global.language('transaction_stock_pick_up_product'),
      global.language('stock_movement_adjust_decrease'),
    };
    if (greenFlags.contains(transFlag)) return global.theme.positiveHighlightTextColor;
    if (redFlags.contains(transFlag)) return global.theme.negativeHighlightTextColor;
    return global.theme.iconColor;
  }

  String _getTransactionFlagText(String transFlag) {
    // transFlag ที่ส่งมาจากระบบจะเป็นข้อความภาษาไทยแล้ว
    // ตามเงื่อนไข SQL case when ที่กำหนด
    return transFlag;
  }
}
