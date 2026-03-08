import 'package:flutter/material.dart';
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart';
import 'package:smlaicloud/model/bi_report/entity_selection_model.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart'; // เพิ่ม import
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import '../../../../global.dart' as global;

class FilterPanel extends StatelessWidget {
  final DateTime? fromDate;
  final DateTime? toDate;
  final bool? showDetails;
  final String? showCancelledDocuments;
  final String? saleType;
  final String? posType;
  final BranchSelectionModel? selectedBranches;
  final EntitySelectionModel? selectedDebtors;
  final EntitySelectionModel? selectedSalespersons;
  final EntitySelectionModel? selectedBarcodes; // เพิ่ม parameter
  final BiReportType reportType; // เพิ่ม parameter
  final VoidCallback? onShowConditionDialog;
  final VoidCallback? onRefresh;

  const FilterPanel({
    super.key,
    this.fromDate,
    this.toDate,
    this.showDetails,
    this.showCancelledDocuments,
    this.saleType,
    this.posType,
    this.selectedBranches,
    this.selectedDebtors,
    this.selectedSalespersons,
    this.selectedBarcodes,
    required this.reportType,
    this.onShowConditionDialog,
    this.onRefresh,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.search, color: Colors.indigo.shade600, size: 18),
              SizedBox(width: 6),
              Text(
                global.language('search_condition'),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Colors.indigo,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Date Range Display (แสดงเสมอ)
          _buildConditionRow(
            reportType == BiReportType.stockBalance ? global.language('as_date') : global.language('date'),
            reportType == BiReportType.stockBalance
                ? (toDate != null
                      ? ReportUtils.formatDate(toDate.toString())
                      : global.language('not_specified'))
                : (fromDate != null && toDate != null
                      ? '${ReportUtils.formatDate(fromDate.toString())} - ${ReportUtils.formatDate(toDate.toString())}'
                      : global.language('not_specified')),
            Icons.date_range,
            Colors.blue.shade600,
          ),

          // สำหรับ Stock Balance แสดงเฉพาะบาร์โค้ด
          if (reportType == BiReportType.stockBalance) ...[
            _buildConditionRow(
              global.language('barcode'),
              _getBarcodeDisplayText(),
              Icons.qr_code,
              Colors.purple.shade600,
            ),
          ],

          // สำหรับ Payment Daily แสดงเฉพาะสาขา
          if (reportType == BiReportType.paymentDaily) ...[
            _buildConditionRow(
              global.language('branch'),
              _getBranchDisplayText(),
              Icons.business,
              Colors.orange.shade600,
            ),
          ],

          // สำหรับ Stock Movement แสดงเฉพาะบาร์โค้ด
          if (reportType == BiReportType.stockMovement) ...[
            _buildConditionRow(
              global.language('barcode'),
              _getBarcodeDisplayText(),
              Icons.qr_code,
              Colors.purple.shade600,
            ),
          ],

          // สำหรับ Sale Return แสดงเฉพาะวันที่และสาขา
          if (reportType == BiReportType.saleReturn) ...[
            // Branch Condition
            _buildConditionRow(
              global.language('branch'),
              _getBranchDisplayText(),
              Icons.business,
              Colors.orange.shade600,
            ),
          ],

          // สำหรับ Sale และ Sale Daily แสดงเงื่อนไขตามประเภท
          if (reportType == BiReportType.sale ||
              reportType == BiReportType.saleDaily) ...[
            // Branch Condition
            _buildConditionRow(
              global.language('branch'),
              _getBranchDisplayText(),
              Icons.business,
              Colors.orange.shade600,
            ),

            // Sale Type Condition
            _buildConditionRow(
              global.language('sale_type'),
              saleType ?? global.language('not_specified'),
              Icons.point_of_sale,
              Colors.green.shade600,
            ),

            // POS Type Condition
            _buildConditionRow(
              global.language('system_sales'),
              posType ?? global.language('not_specified'),
              Icons.computer,
              Colors.purple.shade600,
            ),

            // Details Display Condition
            _buildConditionRow(
              global.language('product_detail'),
              _getShowDetailsText(),
              Icons.list_alt,
              _getShowDetailsColor(),
            ),

            // Cancelled Documents Condition
            _buildConditionRow(
              global.language('show_cancelled_documents'),
              showCancelledDocuments ?? global.language('not_specified'),
              Icons.cancel,
              Colors.red.shade600,
            ),

            // Creditors และ Salespersons เฉพาะ Sale Report
            if (reportType == BiReportType.sale) ...[
              // Creditors Condition (only show if has data)
              if (_hasCreditors())
                _buildConditionRow(
                  global.language('debtor'),
                  _getCreditorsDisplayText(),
                  Icons.people,
                  Colors.blue.shade600,
                ),

              // Salespersons Condition (only show if has data)
              if (_hasSalespersons())
                _buildConditionRow(
                  global.language('sale_person'),
                  _getSalespersonsDisplayText(),
                  Icons.person_pin,
                  Colors.green.shade600,
                ),
            ],
          ],

          if (reportType == BiReportType.purchasepartial) ...[
            // Branch Condition
            _buildConditionRow(
              global.language('branch'),
              _getBranchDisplayText(),
              Icons.business,
              Colors.orange.shade600,
            ),

            // Cancelled Documents Condition
            _buildConditionRow(
              global.language('show_cancelled_documents'),
              showCancelledDocuments ?? global.language('not_specified'),
              Icons.cancel,
              Colors.red.shade600,
            ),

            if (_hasCreditors())
              _buildConditionRow(
                global.language('creditor'),
                _getCreditorsDisplayText(),
                Icons.people,
                Colors.blue.shade600,
              ),
          ],

          // สำหรับ Gross Profit By Document Report
          if (reportType == BiReportType.grossProfitByDocument) ...[
            // Branch Condition
            _buildConditionRow(
              global.language('branch'),
              _getBranchDisplayText(),
              Icons.business,
              Colors.orange.shade600,
            ),

            // Creditor Condition
            if (_hasCreditors())
              _buildConditionRow(
                global.language('creditor'),
                _getCreditorsDisplayText(),
                Icons.people,
                Colors.blue.shade600,
              ),

            // Cancelled Documents Condition
            _buildConditionRow(
              global.language('show_cancelled_documents'),
              showCancelledDocuments ?? global.language('not_specified'),
              Icons.cancel,
              Colors.red.shade600,
            ),
          ],

          // สำหรับ Gross Profit By Product Report
          if (reportType == BiReportType.grossProfitByProduct) ...[
            // Barcode Condition
            _buildConditionRow(
              global.language('barcode'),
              _getBarcodeDisplayText(),
              Icons.qr_code,
              Colors.purple.shade600,
            ),

            // Branch Condition
            _buildConditionRow(
              global.language('branch'),
              _getBranchDisplayText(),
              Icons.business,
              Colors.orange.shade600,
            ),
          ],

          // สำหรับ VAT Sale Report
          if (reportType == BiReportType.vatSale) ...[
            // Branch Condition
            _buildConditionRow(
              global.language('branch'),
              _getBranchDisplayText(),
              Icons.business,
              Colors.orange.shade600,
            ),
          ],

          // สำหรับ VAT Buy Report
          if (reportType == BiReportType.vatBuy) ...[
            // Branch Condition
            _buildConditionRow(
              global.language('branch'),
              _getBranchDisplayText(),
              Icons.business,
              Colors.orange.shade600,
            ),
          ],

          const SizedBox(height: 12),

          // Action Buttons
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: onShowConditionDialog,
                  icon: Icon(Icons.edit, size: 16),
                  label: Text(global.language('edit'), style: TextStyle(fontSize: 12)),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: Colors.indigo.shade600,
                    side: BorderSide(color: Colors.indigo.shade600),
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(6),
                    ),
                  ),
                ),
              ),
              SizedBox(width: 8),
              Expanded(
                child: ElevatedButton.icon(
                  onPressed: onRefresh,
                  icon: Icon(Icons.refresh, size: 16),
                  label: Text(global.language('database_master_info.refresh'), style: TextStyle(fontSize: 12)),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.indigo.shade600,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(6),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // เพิ่ม method สำหรับแสดงบาร์โค้ด
  String _getBarcodeDisplayText() {
    if (selectedBarcodes == null ||
        selectedBarcodes!.selectedEntities.isEmpty) {
      // สำหรับ Stock Balance แสดงว่าเป็นสินค้าทั้งหมด
      if (reportType == BiReportType.stockBalance) {
        return global.language('all_products');
      }
      // สำหรับ Gross Profit By Product แสดงว่าเป็นสินค้าทั้งหมด
      if (reportType == BiReportType.grossProfitByProduct) {
        return global.language('all_products');
      }
      return global.language('not_selected');
    }

    final barcode = selectedBarcodes!.selectedEntities.first;
    return '${barcode.code} : ${ReportUtils.getDisplayNameSafe(barcode.names)}';
  }

  // Helper methods for null-safe text generation
  String _getBranchDisplayText() {
    if (selectedBranches == null ||
        selectedBranches!.selectedBranches.isEmpty) {
      return global.language('kb_all_branches');
    }

    if (selectedBranches!.selectedBranches.length == 1) {
      return ReportUtils.getDisplayNameSafe(
        selectedBranches!.selectedBranches.first.names,
      );
    }

    return '${selectedBranches!.selectedBranches.length} สาขา';
  }

  String _getShowDetailsText() {
    return showDetails == true ? global.language('display') : global.language('not_display');
  }

  Color _getShowDetailsColor() {
    return showDetails == true ? Colors.green.shade600 : Colors.grey.shade600;
  }

  bool _hasCreditors() {
    return selectedDebtors != null &&
        selectedDebtors!.selectedEntities.isNotEmpty;
  }

  String _getCreditorsDisplayText() {
    if (!_hasCreditors()) return '';

    if (selectedDebtors!.selectedEntities.length == 1) {
      return ReportUtils.getDisplayNameSafe(
        selectedDebtors!.selectedEntities.first.names,
      );
    }

    return '${selectedDebtors!.selectedEntities.length} รายการ';
  }

  bool _hasSalespersons() {
    return selectedSalespersons != null &&
        selectedSalespersons!.selectedEntities.isNotEmpty;
  }

  String _getSalespersonsDisplayText() {
    if (!_hasSalespersons()) return '';

    if (selectedSalespersons!.selectedEntities.length == 1) {
      return ReportUtils.getDisplayNameSafe(
        selectedSalespersons!.selectedEntities.first.names,
      );
    }

    return '${selectedSalespersons!.selectedEntities.length} คน';
  }

  Widget _buildConditionRow(
    String label,
    String value,
    IconData icon,
    Color iconColor,
  ) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(4),
            decoration: BoxDecoration(
              color: iconColor.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Icon(icon, size: 14, color: iconColor),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: RichText(
              text: TextSpan(
                style: const TextStyle(fontSize: 12, color: Colors.black87),
                children: [
                  TextSpan(
                    text: '$label: ',
                    style: const TextStyle(
                      fontWeight: FontWeight.w500,
                      color: Colors.grey,
                    ),
                  ),
                  TextSpan(
                    text: value,
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
