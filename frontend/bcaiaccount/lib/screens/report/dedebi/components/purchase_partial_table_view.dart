import 'package:flutter/material.dart';
import 'package:smlaicloud/model/bi_report/purchase_partial_model.dart';
import 'package:smlaicloud/global.dart' as global;

class PurchasePartialTableView extends StatefulWidget {
  final List<PurchasePartialModel> data;
  final String Function(double) formatCurrency;
  final Function(PurchasePartialModel)? onRowTap;

  const PurchasePartialTableView({
    super.key,
    required this.data,
    required this.formatCurrency,
    this.onRowTap,
  });

  @override
  State<PurchasePartialTableView> createState() =>
      _PurchasePartialTableViewState();
}

class _PurchasePartialTableViewState extends State<PurchasePartialTableView> {
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
    if (widget.data.isEmpty) {
      return Center(
        child: Padding(
          padding: EdgeInsets.all(32.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.inventory_outlined, size: 64, color: Colors.grey),
              SizedBox(height: 16),
              Text(
                global.language('no_purchase_partial_report_data'),
                style: const TextStyle(
                  fontSize: 16,
                  color: Colors.grey,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.shade100,
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
              itemCount: widget.data.length,
              itemBuilder: (context, index) {
                return _buildTableRow(widget.data[index], index);
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
        color: Colors.indigo.shade50,
        borderRadius: const BorderRadius.only(
          topLeft: Radius.circular(12),
          topRight: Radius.circular(12),
        ),
        border: Border(bottom: BorderSide(color: Colors.indigo.shade200)),
      ),
      child: Row(
        children: [
          Expanded(
            flex: 2,
            child: Text(
              global.language('date'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
            ),
          ),
          Expanded(
            flex: 3,
            child: Text(
              global.language('document_number'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('reference_document'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('creditor_code'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
            ),
          ),
          Expanded(
            flex: 3,
            child: Text(
              global.language('creditor_name'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('product_value'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
              textAlign: TextAlign.right,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('tax_exempt'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
              textAlign: TextAlign.right,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('before_tax'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
              textAlign: TextAlign.right,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('tax_value'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
              textAlign: TextAlign.right,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language('net_value'),
              style: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.indigo,
              ),
              textAlign: TextAlign.right,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTableRow(PurchasePartialModel purchasePartial, int index) {
    final isHovered = _hoveredIndex == index;

    // Determine row color based on hover state and index
    Color rowColor;
    if (isHovered) {
      rowColor = Colors.indigo.shade100;
    } else {
      rowColor = index.isEven ? Colors.white : Colors.grey.shade50;
    }

    return MouseRegion(
      onEnter: (_) => setState(() => _hoveredIndex = index),
      onExit: (_) => setState(() => _hoveredIndex = null),
      child: InkWell(
        onTap: widget.onRowTap != null
            ? () => widget.onRowTap!(purchasePartial)
            : null,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 150),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: rowColor,
            border: index < widget.data.length - 1
                ? Border(bottom: BorderSide(color: Colors.grey.shade100))
                : null,
          ),
          child: Row(
            children: [
              // วันที่เอกสาร
              Expanded(
                flex: 2,
                child: Text(
                  _formatDate(purchasePartial.docdate),
                  style: TextStyle(
                    fontSize: 11,
                    color: isHovered
                        ? Colors.indigo.shade800
                        : Colors.indigo.shade700,
                    fontWeight: FontWeight.w600,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),

              // เลขที่เอกสาร
              Expanded(
                flex: 3,
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  child: Text(
                    purchasePartial.docno,
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: isHovered
                          ? Colors.blue.shade800
                          : Colors.blue.shade700,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ),

              // เลขที่เอกสารอ้างอิง
              Expanded(
                flex: 2,
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  child: Text(
                    purchasePartial.docrefno.isNotEmpty
                        ? purchasePartial.docrefno
                        : '-',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: purchasePartial.docrefno.isNotEmpty
                          ? (isHovered ? Colors.black87 : Colors.black)
                          : (isHovered
                                ? Colors.grey.shade600
                                : Colors.grey.shade500),
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ),

              // รหัสเจ้าหนี้
              Expanded(
                flex: 2,
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  child: Text(
                    purchasePartial.creditorcode,
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: isHovered ? Colors.black87 : Colors.black,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ),

              // ชื่อเจ้าหนี้
              Expanded(
                flex: 3,
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  child: Text(
                    purchasePartial.creditorNameTh,
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: isHovered ? Colors.black87 : Colors.black,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ),

              // มูลค่าสินค้า
              Expanded(
                flex: 2,
                child: Text(
                  widget.formatCurrency(purchasePartial.totalvalue),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: isHovered
                        ? Colors.green.shade800
                        : Colors.green.shade700,
                  ),
                  textAlign: TextAlign.right,
                ),
              ),

              // มูลค่ายกเว้นภาษี
              Expanded(
                flex: 2,
                child: Text(
                  widget.formatCurrency(purchasePartial.totalexceptvat),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                    color: isHovered
                        ? Colors.orange.shade800
                        : Colors.orange.shade700,
                  ),
                  textAlign: TextAlign.right,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),

              // มูลค่าก่อนภาษี
              Expanded(
                flex: 2,
                child: Text(
                  widget.formatCurrency(purchasePartial.totalbeforevat),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                    color: isHovered
                        ? Colors.cyan.shade800
                        : Colors.cyan.shade700,
                  ),
                  textAlign: TextAlign.right,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),

              // มูลค่าภาษี
              Expanded(
                flex: 2,
                child: Text(
                  widget.formatCurrency(purchasePartial.totalvatvalue),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                    color: isHovered
                        ? Colors.red.shade800
                        : Colors.red.shade700,
                  ),
                  textAlign: TextAlign.right,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),

              // มูลค่าสุทธิ
              Expanded(
                flex: 2,
                child: Text(
                  widget.formatCurrency(purchasePartial.totalamount),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: isHovered
                        ? Colors.indigo.shade800
                        : Colors.indigo.shade700,
                  ),
                  textAlign: TextAlign.right,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatDate(String dateStr) {
    try {
      final date = DateTime.parse(dateStr);
      return '${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year + 543}';
    } catch (e) {
      return dateStr;
    }
  }
}
