import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';

/// Widget สำหรับจัดการ Vendor Entries ใน RFQ (ใบสืบราคา)
/// แสดงรายการ vendor แต่ละราย พร้อมรายการสินค้าและราคา
/// รองรับเพิ่ม/ลบ vendor, เลือก vendor ที่ดีที่สุด
class RFQVendorWidget extends StatefulWidget {
  const RFQVendorWidget({
    super.key,
    required this.vendorEntries,
    required this.onVendorEntriesChanged,
    required this.isEditable,
    required this.details,
  });

  final List<Map<String, dynamic>> vendorEntries;
  final void Function(List<Map<String, dynamic>> entries) onVendorEntriesChanged;
  final bool isEditable;
  final List<TransactionDetailModel> details;

  @override
  State<RFQVendorWidget> createState() => _RFQVendorWidgetState();
}

class _RFQVendorWidgetState extends State<RFQVendorWidget> with global.ThemeRefreshMixin {
  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildVendorHeader(),
        Expanded(
          child: widget.vendorEntries.isEmpty
              ? _buildEmptyState()
              : _buildVendorList(),
        ),
      ],
    );
  }

  Widget _buildVendorHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor)),
      ),
      child: Row(
        children: [
          Icon(Icons.store, size: 18, color: global.theme.infoHighlightTextColor),
          const SizedBox(width: 8),
          Text(
            '${global.language("vendor")} (${widget.vendorEntries.length})',
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.bold,
              color: global.theme.textColor,
            ),
          ),
          const Spacer(),
          if (widget.isEditable)
            TextButton.icon(
              icon: Icon(Icons.add, size: 18),
              label: Text(global.language('add')),
              onPressed: _addVendor,
            ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.store_outlined, size: 48, color: global.theme.iconSecondaryColor),
          const SizedBox(height: 8),
          Text(
            global.language('no_data'),
            style: TextStyle(color: global.theme.iconSecondaryColor),
          ),
          if (widget.isEditable) ...[
            const SizedBox(height: 12),
            ElevatedButton.icon(
              icon: const Icon(Icons.add),
              label: Text(global.language('add_vendor')),
              onPressed: _addVendor,
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildVendorList() {
    return ListView.builder(
      padding: const EdgeInsets.all(8),
      itemCount: widget.vendorEntries.length,
      itemBuilder: (context, index) {
        return _buildVendorCard(index);
      },
    );
  }

  Widget _buildVendorCard(int index) {
    final vendor = widget.vendorEntries[index];
    final vendorCode = vendor['vendorcode'] as String? ?? '';
    final vendorName = _getVendorName(vendor);
    final isSelected = vendor['isselected'] == true;
    final totalAmount = (vendor['totalamount'] as num?)?.toDouble() ?? 0;
    final creditDays = vendor['creditdays'] as int? ?? 0;
    final deliveryDays = vendor['deliverydays'] as int? ?? 0;
    final quotationDocNo = vendor['quotationdocno'] as String? ?? '';

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      elevation: isSelected ? 3 : 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
        side: isSelected
            ? BorderSide(color: global.theme.positiveHighlightTextColor, width: 2)
            : BorderSide.none,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Vendor header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: isSelected
                  ? global.theme.positiveHighlightTextColor.withValues(alpha: 0.1)
                  : global.theme.cardColor,
              borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
            ),
            child: Row(
              children: [
                Icon(
                  isSelected ? Icons.check_circle : Icons.store,
                  size: 20,
                  color: isSelected
                      ? global.theme.positiveHighlightTextColor
                      : global.theme.iconSecondaryColor,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        vendorName.isNotEmpty ? vendorName : '${global.language("vendor")} ${index + 1}',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          color: global.theme.textColor,
                        ),
                      ),
                      if (vendorCode.isNotEmpty)
                        Text(
                          vendorCode,
                          style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor),
                        ),
                    ],
                  ),
                ),
                if (isSelected)
                  Chip(
                    label: Text(
                      global.language('selected'),
                      style: TextStyle(fontSize: 11, color: global.theme.positiveHighlightTextColor),
                    ),
                    backgroundColor: global.theme.positiveHighlightTextColor.withValues(alpha: 0.15),
                    side: BorderSide.none,
                    padding: EdgeInsets.zero,
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
                if (widget.isEditable) ...[
                  if (!isSelected)
                    IconButton(
                      icon: Icon(Icons.check_circle_outline, size: 20, color: global.theme.iconSecondaryColor),
                      tooltip: global.language('select'),
                      onPressed: () => _selectVendor(index),
                    ),
                  IconButton(
                    icon: Icon(Icons.delete_outline, size: 20, color: global.theme.negativeHighlightTextColor),
                    tooltip: global.language('delete'),
                    onPressed: () => _removeVendor(index),
                  ),
                ],
              ],
            ),
          ),
          // Vendor details
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Quotation info row
                Wrap(
                  spacing: 16,
                  runSpacing: 4,
                  children: [
                    if (quotationDocNo.isNotEmpty)
                      _buildInfoChip(Icons.receipt, '${global.language("docno")}: $quotationDocNo'),
                    if (creditDays > 0)
                      _buildInfoChip(Icons.schedule, '${global.language("credit_days")}: $creditDays'),
                    if (deliveryDays > 0)
                      _buildInfoChip(Icons.local_shipping, '${global.language("delivery")}: $deliveryDays ${global.language("day")}'),
                  ],
                ),
                const SizedBox(height: 8),
                // Total amount
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    Text(
                      '${global.language("total")}: ',
                      style: TextStyle(fontSize: 13, color: global.theme.textSecondaryColor),
                    ),
                    Text(
                      global.moneyFormat.format(totalAmount),
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: isSelected
                            ? global.theme.positiveHighlightTextColor
                            : global.theme.textColor,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInfoChip(IconData icon, String text) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: global.theme.iconSecondaryColor),
        const SizedBox(width: 4),
        Text(text, style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
      ],
    );
  }

  String _getVendorName(Map<String, dynamic> vendor) {
    final names = vendor['vendornames'];
    if (names is List && names.isNotEmpty) {
      return global.activeLangName(
        names.map((n) => LanguageDataModel.fromJson(n is Map<String, dynamic> ? n : {})).toList(),
      );
    }
    return '';
  }

  void _addVendor() {
    final entries = List<Map<String, dynamic>>.from(widget.vendorEntries);
    entries.add({
      'vendorcode': '',
      'vendornames': <Map<String, dynamic>>[],
      'quotationdocno': '',
      'quotationdate': '',
      'creditdays': 0,
      'deliverydays': 0,
      'deliveryterms': '',
      'qualitynotes': '',
      'totalamount': 0.0,
      'isselected': false,
      'items': <Map<String, dynamic>>[],
    });
    widget.onVendorEntriesChanged(entries);
  }

  void _removeVendor(int index) {
    final entries = List<Map<String, dynamic>>.from(widget.vendorEntries);
    entries.removeAt(index);
    widget.onVendorEntriesChanged(entries);
  }

  void _selectVendor(int index) {
    final entries = List<Map<String, dynamic>>.from(widget.vendorEntries);
    for (int i = 0; i < entries.length; i++) {
      entries[i] = Map<String, dynamic>.from(entries[i]);
      entries[i]['isselected'] = (i == index);
    }
    widget.onVendorEntriesChanged(entries);
  }
}
