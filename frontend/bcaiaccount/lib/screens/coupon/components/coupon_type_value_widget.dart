import 'package:flutter/material.dart';
import '../../../global.dart' as global;

/// Widget สำหรับเลือกประเภทคูปองและกำหนดค่าส่วนลด
class CouponTypeValueWidget extends StatelessWidget {
  final int couponType;
  final TextEditingController couponValueController;
  final FocusNode couponValueFocusNode;
  final Map<String, String> fieldErrors;
  final Function(int) onTypeChanged;
  final Function(String) onValueChanged;

  const CouponTypeValueWidget({
    super.key,
    required this.couponType,
    required this.couponValueController,
    required this.couponValueFocusNode,
    required this.fieldErrors,
    required this.onTypeChanged,
    required this.onValueChanged,
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

  Widget _buildRequiredField({
    required TextEditingController controller,
    required FocusNode focusNode,
    required String label,
    required String hint,
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
                color: hasError ? global.theme.negativeHighlightTextColor : global.theme.textColor,
              ),
            ),
            Text(
              ' *',
              style: TextStyle(
                color: global.theme.negativeHighlightTextColor,
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
              color: hasError ? global.theme.negativeHighlightTextColor : global.theme.dividerBorderColor,
              width: hasError ? 1.5 : 1,
            ),
          ),
          child: TextFormField(
            controller: controller,
            focusNode: focusNode,
            keyboardType: keyboardType,
            onChanged: onChanged,
            style: TextStyle(fontSize: 16),
            decoration: InputDecoration(
              hintText: hint,
              hintStyle: TextStyle(color: global.theme.formHintColor, fontSize: 15),
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
              color: global.theme.negativeHighlightTextColor,
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
        SizedBox(height: hasError ? 12 : 16),
      ],
    );
  }

  Widget _buildCouponTypeSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          global.language("coupon_type_label"),
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w500,
            color: global.theme.textColor,
          ),
        ),
        SizedBox(height: 6),
        Row(
          children: [
            Expanded(
              child: _buildTypeOption(
                type: 0,
                icon: Icons.money_off,
                title: global.language("coupon_type_fixed_amount"),
                description: global.language("coupon_type_fixed_amount_desc"),
              ),
            ),
            SizedBox(width: 6),
            Expanded(
              child: _buildTypeOption(
                type: 1,
                icon: Icons.percent,
                title: global.language("coupon_type_percentage"),
                description: global.language("coupon_type_percentage_desc"),
              ),
            ),
            SizedBox(width: 6),
            Expanded(
              child: _buildTypeOption(
                type: 2,
                icon: Icons.local_atm,
                title: global.language("coupon_type_cash"),
                description: global.language("coupon_type_cash_desc"),
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
      ],
    );
  }

  Widget _buildTypeOption({
    required int type,
    required IconData icon,
    required String title,
    required String description,
  }) {
    final isSelected = couponType == type;
    return GestureDetector(
      onTap: () => onTypeChanged(type),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isSelected
              ? global.theme.buttonColor.withValues(alpha: 0.1)
              : global.theme.onPrimaryColor,
          border: Border.all(
            color: isSelected ? global.theme.buttonColor : global.theme.dividerBorderColor,
            width: isSelected ? 2 : 1,
          ),
          borderRadius: BorderRadius.circular(6),
        ),
        child: Column(
          children: [
            Icon(
              icon,
              size: 28,
              color: isSelected ? global.theme.buttonColor : global.theme.textSecondaryColor,
            ),
            const SizedBox(height: 6),
            Text(
              title,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: isSelected ? global.theme.buttonColor : global.theme.textSecondaryColor,
              ),
            ),
            Text(
              description,
              style: TextStyle(
                fontSize: 11,
                color: isSelected ? global.theme.buttonColor : global.theme.textSecondaryColor,
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _getCouponValueLabel() {
    switch (couponType) {
      case 0:
        return global.language("coupon_value_fixed_amount");
      case 1:
        return global.language("coupon_value_percentage");
      case 2:
        return global.language("coupon_value_cash_back");
      default:
        return global.language("coupon_value_default");
    }
  }

  String _getCouponValueHint() {
    switch (couponType) {
      case 0:
        return global.language("coupon_hint_fixed_amount");
      case 1:
        return global.language("coupon_hint_percentage");
      case 2:
        return global.language("coupon_hint_cash_back");
      default:
        return global.language("coupon_hint_value_default");
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),

        // Section Header
        _buildSectionHeader(
          global.language("coupon_section_type_value"),
          Icons.local_offer,
        ),
        const SizedBox(height: 16),

        // เลือกประเภทคูปอง
        _buildCouponTypeSelector(),

        // ค่าส่วนลด (จำเป็น)
        _buildRequiredField(
          controller: couponValueController,
          focusNode: couponValueFocusNode,
          label: _getCouponValueLabel(),
          hint: _getCouponValueHint(),
          keyboardType: TextInputType.number,
          onChanged: onValueChanged,
          errorKey: 'coupon_value',
        ),
      ],
    );
  }
}
