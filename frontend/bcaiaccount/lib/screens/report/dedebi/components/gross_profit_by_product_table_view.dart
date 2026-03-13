import 'package:flutter/material.dart';
import 'package:smlaicloud/model/bi_report/gross_profit_by_product_model.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import '../../../../global.dart' as global;

/// Table view component for displaying Gross Profit By Product report data
/// Shows profit analysis per product/item with multi-language support
class GrossProfitByProductTableView extends StatelessWidget {
  final List<GrossProfitByProductModel> data;
  final Function(GrossProfitByProductModel)? onRowTap;

  const GrossProfitByProductTableView({
    super.key,
    required this.data,
    this.onRowTap,
  });

  @override
  Widget build(BuildContext context) {
    if (data.isEmpty) {
      return Center(
        child: Padding(
          padding: EdgeInsets.all(48.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.info_outline, size: 64, color: global.theme.iconSecondaryColor),
              SizedBox(height: 16),
              Text(
                global.language('no_report_data'),
                style: TextStyle(
                  fontSize: 18,
                  color: global.theme.textSecondaryColor,
                  fontWeight: FontWeight.w500,
                ),
              ),
              SizedBox(height: 8),
              Text(
                global.language('please_adjust_search_criteria'),
                style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
              ),
            ],
          ),
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
        Expanded(flex: 2, child: _buildHeaderCell(global.language('product_barcode'))),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('product_code'))),
        Expanded(flex: 3, child: _buildHeaderCell(global.language('import_product_detail.product_name'))),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('unit'))),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('sales_quantity'), TextAlign.right)),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_value'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('cost_of_sales'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('profit_loss'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('gross_profit_pct'), TextAlign.right),
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

  Widget _buildDataRow(GrossProfitByProductModel item, int index) {
    final isEven = index % 2 == 0;
    final backgroundColor = isEven ? global.theme.columnAlternateEvenColor : global.theme.columnAlternateOddColor;

    return InkWell(
      onTap: () => onRowTap?.call(item),
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
            Expanded(flex: 2, child: _buildDataCell(item.mainbarcoderef!)),
            Expanded(flex: 2, child: _buildDataCell(item.itemcode!)),
            Expanded(
              flex: 3,
              child: _buildDataCell(
                _getProductName(item.names!),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(_getUnitName(item.unitnames!)),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                item.qty!.toStringAsFixed(2),
                textAlign: TextAlign.right,
              ),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrency(item.sumamountexcludevat!),
                textAlign: TextAlign.right,
                fontWeight: FontWeight.w600,
              ),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrency(item.totalcost!),
                textAlign: TextAlign.right,
                color: global.theme.negativeHighlightTextColor,
                fontWeight: FontWeight.w500,
              ),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrency(item.pal!),
                textAlign: TextAlign.right,
                fontWeight: FontWeight.bold,
                color: item.pal! >= 0
                    ? global.theme.positiveHighlightTextColor
                    : global.theme.negativeHighlightTextColor,
              ),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                '${item.perPal!.toStringAsFixed(2)}%',
                textAlign: TextAlign.right,
                fontWeight: FontWeight.bold,
                color: item.perPal! >= 0
                    ? global.theme.positiveHighlightTextColor
                    : global.theme.negativeHighlightTextColor,
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
    TextDecoration? decoration,
    TextOverflow? overflow,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Text(
        text,
        style: TextStyle(
          fontSize: fontSize ?? 11,
          fontWeight: fontWeight,
          color: color,
          decoration: decoration,
        ),
        textAlign: textAlign,
        overflow: overflow,
        maxLines: 2,
      ),
    );
  }

  String _getProductName(List<ProductName> names) {
    if (names.isEmpty) return '-';

    // Try to get Thai name first
    final thName = names.firstWhere(
      (name) => name.code == 'th',
      orElse: () => names.first,
    );

    return thName.name!;
  }

  String _getUnitName(List<UnitName> names) {
    if (names.isEmpty) return '-';

    // Try to get Thai name first
    final thName = names.firstWhere(
      (name) => name.code == 'th',
      orElse: () => names.first,
    );

    return thName.name!;
  }
}
