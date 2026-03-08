import 'package:flutter/material.dart';
import 'package:smlaicloud/model/bi_report/vat_sale_model.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import '../../../../global.dart' as global;

/// Table view component for displaying VAT Sale report data
/// Shows tax document records with creditor information and VAT calculations
class VatSaleTableView extends StatelessWidget {
  final List<VatSaleModel> data;
  final Function(VatSaleModel)? onRowTap;

  const VatSaleTableView({super.key, required this.data, this.onRowTap});

  @override
  Widget build(BuildContext context) {
    if (data.isEmpty) {
      return Center(
        child: Padding(
          padding: EdgeInsets.all(48.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.info_outline, size: 64, color: Colors.grey.shade400),
              SizedBox(height: 16),
              Text(
                global.language('no_report_data'),
                style: TextStyle(
                  fontSize: 18,
                  color: Colors.grey.shade600,
                  fontWeight: FontWeight.w500,
                ),
              ),
              SizedBox(height: 8),
              Text(
                global.language('please_adjust_search_criteria'),
                style: TextStyle(fontSize: 14, color: Colors.grey.shade500),
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
            color: Colors.grey.shade100,
            border: Border(
              bottom: BorderSide(color: Colors.grey.shade300, width: 1),
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
        Expanded(flex: 2, child: _buildHeaderCell(global.language('date'))),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('vat_num'))),
        Expanded(flex: 3, child: _buildHeaderCell(global.language('customer_name'))),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('tax_id_number'))),
        Expanded(flex: 1, child: _buildHeaderCell(global.language('branch'))),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('tax_base'), TextAlign.right)),
        Expanded(flex: 2, child: _buildHeaderCell(global.language('vat'), TextAlign.right)),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('sale_daily_report_except_vat'), TextAlign.right),
        ),
        Expanded(
          flex: 2,
          child: _buildHeaderCell(global.language('product_amount'), TextAlign.right),
        ),
        Expanded(flex: 1, child: _buildHeaderCell(global.language('status'))),
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

  Widget _buildDataRow(VatSaleModel item, int index) {
    final isEven = index % 2 == 0;
    final backgroundColor = isEven ? Colors.white : Colors.grey.shade50;

    return InkWell(
      onTap: () => onRowTap?.call(item),
      hoverColor: Colors.indigo.shade50,
      highlightColor: Colors.indigo.shade100,
      splashColor: Colors.indigo.shade200,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: backgroundColor,
          border: Border(
            bottom: BorderSide(color: Colors.grey.shade200, width: 0.5),
          ),
        ),
        child: Row(
          children: [
            Expanded(
              flex: 2,
              child: _buildDataCell(_formatDate(item.taxdocdate)),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(item.taxdocno, fontWeight: FontWeight.w600),
            ),
            Expanded(
              flex: 3,
              child: _buildDataCell(_getCreditorName(item.creditorname)),
            ),
            Expanded(flex: 2, child: _buildDataCell(item.taxid)),
            Expanded(flex: 1, child: _buildDataCell(item.branchcode)),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrency(item.totalbeforevat),
                textAlign: TextAlign.right,
              ),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrency(item.totalvatvalue),
                textAlign: TextAlign.right,
                color: Colors.blue.shade700,
                fontWeight: FontWeight.w500,
              ),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrency(item.totalexceptvat),
                textAlign: TextAlign.right,
                color: Colors.orange.shade700,
                fontWeight: FontWeight.w500,
              ),
            ),
            Expanded(
              flex: 2,
              child: _buildDataCell(
                ReportUtils.formatCurrency(item.totalamount),
                textAlign: TextAlign.right,
                fontWeight: FontWeight.bold,
                color: Colors.green.shade700,
              ),
            ),
            Expanded(
              flex: 1,
              child: _buildDataCell(
                item.status.isEmpty ? '-' : item.status,
                color: item.status.isEmpty ? Colors.grey : Colors.red.shade700,
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
      ),
    );
  }

  String _formatDate(String dateString) {
    try {
      final date = DateTime.parse(dateString);
      return DateFormat('dd/MM/yyyy').format(date);
    } catch (e) {
      return dateString;
    }
  }

  String _getCreditorName(List<CreditorName> names) {
    if (names.isEmpty) return '-';

    // Try to get Thai name first
    final thName = names.firstWhere(
      (name) => name.code == 'th',
      orElse: () => names.first,
    );

    return thName.name;
  }
}
