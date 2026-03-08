import 'package:flutter/material.dart';
import '../../../global.dart' as global;
import '../../../widgets/date_picker_widget.dart';

/// Widget สำหรับจัดการเงื่อนไขการใช้คูปอง
/// - วันที่ออกและหมดอายุ
/// - จำนวนการใช้งานทั้งหมด
/// - จำนวนการใช้งานต่อลูกค้า
class CouponConditionsWidget extends StatelessWidget {
  final DateTime? issuedDate;
  final DateTime? expiryDate;
  final TextEditingController maxUsageCountController;
  final TextEditingController maxUsageCountPerCustomerController;
  final FocusNode maxUsageCountFocusNode;
  final FocusNode maxUsageCountPerCustomerFocusNode;
  final Map<String, String> fieldErrors;
  final bool isEditMode;
  final Function(DateTime?) onIssuedDateChanged;
  final Function(DateTime?) onExpiryDateChanged;
  final Function(String) onMaxUsageChanged;
  final Function(String) onMaxUsagePerCustomerChanged;

  const CouponConditionsWidget({
    super.key,
    required this.issuedDate,
    required this.expiryDate,
    required this.maxUsageCountController,
    required this.maxUsageCountPerCustomerController,
    required this.maxUsageCountFocusNode,
    required this.maxUsageCountPerCustomerFocusNode,
    required this.fieldErrors,
    required this.isEditMode,
    required this.onIssuedDateChanged,
    required this.onExpiryDateChanged,
    required this.onMaxUsageChanged,
    required this.onMaxUsagePerCustomerChanged,
  });

  Widget _buildSectionHeader(String title, IconData icon) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 16),
      decoration: BoxDecoration(
        color: global.theme.appBarColor.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: global.theme.appBarColor.withValues(alpha: 0.2)),
      ),
      child: Row(
        children: [
          Icon(icon, size: 20, color: global.theme.appBarColor),
          const SizedBox(width: 10),
          Text(
            title,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: global.theme.appBarColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDateField(
    String label,
    DateTime? selectedDate,
    Function(DateTime?) onDateSelected, {
    bool isRequired = false,
    String? errorKey,
  }) {
    final hasError = errorKey != null && fieldErrors.containsKey(errorKey);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        DatePickerWidget(
          label: label,
          selectedDate: selectedDate,
          onDateSelected: onDateSelected,
          isEnabled: isEditMode,
          isRequired: isRequired,
          hasError: hasError,
        ),
        if (hasError) ...[
          Text(
            fieldErrors[errorKey]!,
            style: TextStyle(
              color: Colors.red[600],
              fontSize: 11,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 5),
        ],
      ],
    );
  }

  Widget _buildRequiredField({
    required TextEditingController controller,
    required FocusNode focusNode,
    required String label,
    required String hint,
    bool isRequired = true,
    TextInputType? keyboardType,
    Function(String)? onChanged,
    String? errorKey,
  }) {
    final hasError = errorKey != null && fieldErrors.containsKey(errorKey);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              label,
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w500,
                color: hasError ? Colors.red[700] : Colors.grey[700],
              ),
            ),
            if (isRequired)
              Text(
                ' *',
                style: TextStyle(
                  color: Colors.red[600],
                  fontSize: 15,
                  fontWeight: FontWeight.bold,
                ),
              ),
          ],
        ),
        const SizedBox(height: 8),
        Container(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(6),
            border: Border.all(
              color: hasError ? Colors.red[400]! : Colors.grey[300]!,
              width: hasError ? 1.5 : 1,
            ),
          ),
          child: TextFormField(
            controller: controller,
            focusNode: focusNode,
            keyboardType: keyboardType,
            onChanged: onChanged,
            style: const TextStyle(fontSize: 16),
            decoration: InputDecoration(
              hintText: hint,
              hintStyle: TextStyle(color: Colors.grey[400], fontSize: 15),
              border: InputBorder.none,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 14,
                vertical: 14,
              ),
              isDense: true,
            ),
          ),
        ),
        if (hasError) ...[
          const SizedBox(height: 6),
          Text(
            fieldErrors[errorKey]!,
            style: TextStyle(
              color: Colors.red[600],
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
        SizedBox(height: hasError ? 12 : 16),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),

        // Section Header
        _buildSectionHeader(
          global.language("coupon_section_conditions"),
          Icons.rule,
        ),
        const SizedBox(height: 16),

        // วันที่ออกและหมดอายุ
        Row(
          children: [
            Expanded(
              child: _buildDateField(
                global.language("coupon_issued_date"),
                issuedDate,
                onIssuedDateChanged,
                isRequired: true,
                errorKey: 'issued_date',
              ),
            ),
            SizedBox(width: 16),
            Expanded(
              child: _buildDateField(
                global.language("coupon_expiry_date"),
                expiryDate,
                onExpiryDateChanged,
                isRequired: true,
                errorKey: 'expiry_date',
              ),
            ),
          ],
        ),

        const SizedBox(height: 16),

        // จำนวนการใช้งาน
        Row(
          children: [
            Expanded(
              child: _buildRequiredField(
                controller: maxUsageCountController,
                focusNode: maxUsageCountFocusNode,
                label: global.language("coupon_max_usage_total"),
                hint: global.language("coupon_hint_usage_example"),
                isRequired: true,
                keyboardType: TextInputType.number,
                onChanged: onMaxUsageChanged,
                errorKey: 'max_usage_count',
              ),
            ),
            SizedBox(width: 12),
            Expanded(
              child: _buildRequiredField(
                controller: maxUsageCountPerCustomerController,
                focusNode: maxUsageCountPerCustomerFocusNode,
                label: global.language("coupon_max_usage_per_customer"),
                hint: global.language("coupon_hint_unlimited"),
                isRequired: false,
                keyboardType: TextInputType.number,
                onChanged: onMaxUsagePerCustomerChanged,
                errorKey: null,
              ),
            ),
          ],
        ),
      ],
    );
  }
}
