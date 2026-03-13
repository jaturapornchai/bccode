import 'package:flutter/material.dart';
import '../utils/date_picker.dart';
import 'package:smlaicloud/global.dart' as global;

/// DatePickerWidget - ส่วนประกอบสำหรับเลือกวันที่ที่ใช้ได้ทั่วไปในระบบ
/// รองรับการแสดงปี พ.ศ. และมี UI ที่สอดคล้องกับ _buildOptionalField และ _buildRequiredField
/// ใช้ CustomDatePicker สำหรับการแสดงปฏิทินที่สวยงามและรองรับภาษาไทย
class DatePickerWidget extends StatefulWidget {
  final String label;
  final DateTime? selectedDate;
  final Function(DateTime?)? onDateSelected;
  final bool isEnabled;
  final bool isRequired;
  final bool hasError;

  const DatePickerWidget({
    super.key,
    required this.label,
    required this.selectedDate,
    required this.onDateSelected,
    this.isEnabled = true,
    this.isRequired = false,
    this.hasError = false,
  });

  @override
  State<DatePickerWidget> createState() => _DatePickerWidgetState();
}

class _DatePickerWidgetState extends State<DatePickerWidget> with global.ThemeRefreshMixin {
  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Label with required indicator (เหมือนกับ _buildRequiredField)
        Row(
          children: [
            Text(
              widget.label,
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w500,
                color: widget.hasError ? global.theme.negativeHighlightTextColor : global.theme.textColor,
              ),
            ),
            if (widget.isRequired)
              Text(
                ' *',
                style: TextStyle(
                  color: global.theme.negativeHighlightTextColor,
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                ),
              ),
          ],
        ),
        const SizedBox(height: 6),

        // ใช้ CustomDatePicker
        CustomDatePicker(
          labelText: null, // ไม่ใช้ label ของ CustomDatePicker เพราะแสดงด้านบนแล้ว
          initialDate: widget.selectedDate,
          // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ (global.getYearType(), global.getBranchLanguage())
          useIconSelectDate: true, // แสดงไอคอนปฏิทิน
          firstDate: DateTime(2000),
          lastDate: DateTime(2100),
          onDateSelected: widget.isEnabled ? widget.onDateSelected : null,
          decoration: InputDecoration(
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 12,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: widget.hasError ? global.theme.negativeHighlightTextColor : global.theme.dividerBorderColor,
                width: widget.hasError ? 1.5 : 1,
              ),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: widget.hasError ? global.theme.negativeHighlightTextColor : global.theme.dividerBorderColor,
                width: widget.hasError ? 1.5 : 1,
              ),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: widget.hasError ? global.theme.negativeHighlightTextColor : global.theme.dividerBorderColor,
                width: widget.hasError ? 1.5 : 1,
              ),
            ),
            errorBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: global.theme.negativeHighlightTextColor,
                width: 1.5,
              ),
            ),
          ),
        ),

        // Bottom spacing เหมือนกับ _buildRequiredField/_buildOptionalField
        const SizedBox(height: 12),
      ],
    );
  }
}
