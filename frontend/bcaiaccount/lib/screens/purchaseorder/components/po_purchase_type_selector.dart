import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/approval_model.dart';

/// Widget สำหรับเลือกประเภทการจัดซื้อ (Purchase Type)
class POPurchaseTypeSelector extends StatelessWidget {
  const POPurchaseTypeSelector({
    super.key,
    required this.purchaseTypes,
    required this.selectedCode,
    required this.onChanged,
    this.isEnabled = true,
    this.isLoading = false,
    this.labelText,
    this.hintText,
  });

  final List<PurchaseTypeModel> purchaseTypes;
  final String? selectedCode;
  final void Function(String code, String name) onChanged;
  final bool isEnabled;
  final bool isLoading;
  final String? labelText;
  final String? hintText;

  @override
  Widget build(BuildContext context) {
    final currentLang = global.appConfig.getString("langcode") ?? "th";

    if (isLoading) {
      return _buildLoadingState();
    }

    if (purchaseTypes.isEmpty) {
      return _buildEmptyState();
    }

    return DropdownButtonFormField<String>(
      initialValue: selectedCode,
      decoration: InputDecoration(
        labelText: labelText ?? global.language('purchase_type'),
        hintText: hintText ?? global.language('select_purchase_type'),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        border: const OutlineInputBorder(),
        enabled: isEnabled,
      ),
      items: purchaseTypes.map((type) {
        return DropdownMenuItem<String>(
          value: type.code,
          child: Text(type.getName(currentLang)),
        );
      }).toList(),
      onChanged: isEnabled
          ? (code) {
              if (code != null) {
                final selectedType =
                    purchaseTypes.firstWhere((t) => t.code == code);
                onChanged(code, selectedType.getName(currentLang));
              }
            }
          : null,
    );
  }

  Widget _buildLoadingState() {
    return InputDecorator(
      decoration: InputDecoration(
        labelText: labelText ?? global.language('purchase_type'),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        border: const OutlineInputBorder(),
      ),
      child: Row(
        children: [
          SizedBox(
            width: 16,
            height: 16,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
          SizedBox(width: 8),
          Text(global.language('loading')),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return InputDecorator(
      decoration: InputDecoration(
        labelText: labelText ?? global.language('purchase_type'),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        border: const OutlineInputBorder(),
      ),
      child: Text(
        global.language('no_purchase_type_data'),
        style: TextStyle(color: Colors.grey[600]),
      ),
    );
  }
}

/// Badge แสดงประเภทการจัดซื้อที่เลือก (สำหรับแสดงใน list หรือ card)
class POPurchaseTypeBadge extends StatelessWidget {
  const POPurchaseTypeBadge({
    super.key,
    required this.purchaseTypeCode,
    required this.purchaseTypeName,
    this.fontSize = 12,
  });

  final String purchaseTypeCode;
  final String purchaseTypeName;
  final double fontSize;

  @override
  Widget build(BuildContext context) {
    if (purchaseTypeCode.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.purple[50],
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: Colors.purple[200]!),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.category, size: 14, color: Colors.purple[700]),
          const SizedBox(width: 4),
          Text(
            purchaseTypeName.isNotEmpty ? purchaseTypeName : purchaseTypeCode,
            style: TextStyle(
              color: Colors.purple[700],
              fontSize: fontSize,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }
}
