import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../global.dart' as global;
import '../../../model/global_model.dart';

/// Widget สำหรับแสดงและจัดการข้อมูลพื้นฐานของคูปอง
/// - รหัสคูปอง
/// - ชื่อคูปองในหลายภาษา
class CouponBasicInfoWidget extends StatelessWidget {
  final List<TextEditingController> fieldTextController;
  final List<global.FieldFocusModel> fieldFocusNodes;
  final List<LanguageModel> languageList;
  final Map<String, String> fieldErrors;
  final bool isLoadTranslation;
  final Function(String) onCodeChanged;
  final Function(String) onNameChanged;

  const CouponBasicInfoWidget({
    super.key,
    required this.fieldTextController,
    required this.fieldFocusNodes,
    required this.languageList,
    required this.fieldErrors,
    required this.isLoadTranslation,
    required this.onCodeChanged,
    required this.onNameChanged,
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
    bool isRequired = true,
    TextInputType? keyboardType,
    int maxLines = 1,
    List<TextInputFormatter>? inputFormatters,
    TextCapitalization textCapitalization = TextCapitalization.none,
    Widget? suffixIcon,
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
            if (isRequired)
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
            maxLines: maxLines,
            inputFormatters: inputFormatters,
            textCapitalization: textCapitalization,
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
              suffixIcon: suffixIcon,
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

  Widget _buildLoadingIcon() {
    return Container(
      width: 20,
      height: 20,
      margin: const EdgeInsets.all(12),
      child: CircularProgressIndicator(
        strokeWidth: 2,
        valueColor: AlwaysStoppedAnimation<Color>(global.theme.buttonColor),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Section Header
        _buildSectionHeader(
          global.language("coupon_section_basic_info"),
          Icons.receipt,
        ),
        const SizedBox(height: 16),

        // รหัสคูปอง (จำเป็น)
        _buildRequiredField(
          controller: fieldTextController[0],
          focusNode: fieldFocusNodes[0].focusNode,
          label: global.language("coupon_code"),
          hint: global.language("coupon_hint_code_example"),
          isRequired: true,
          textCapitalization: TextCapitalization.characters,
          inputFormatters: [
            FilteringTextInputFormatter.allow(RegExp('[a-zA-Z0-9-]')),
            FilteringTextInputFormatter.deny(' '),
          ],
          onChanged: (value) {
            onCodeChanged(value.toUpperCase());
          },
          errorKey: 'coupon_code',
        ),

        // ชื่อคูปองแต่ละภาษา (จำเป็นอย่างน้อย 1 ภาษา)
        for (int i = 0; i < languageList.length; i++)
          _buildRequiredField(
            controller: fieldTextController[i + 1],
            focusNode: fieldFocusNodes[i + 1].focusNode,
            label:
                "${global.language("coupon_name")} (${languageList[i].name})",
            hint: global.language("coupon_hint_name"),
            isRequired: i == 0, // ภาษาแรกเป็น required
            onChanged: (value) => onNameChanged(value),
            suffixIcon: isLoadTranslation ? _buildLoadingIcon() : null,
            errorKey: i == 0 ? 'coupon_name' : null,
          ),
      ],
    );
  }
}
