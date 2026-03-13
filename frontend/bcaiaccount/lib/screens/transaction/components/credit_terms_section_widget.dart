import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/date_picker.dart';

/// Credit Terms Section Widget - สำหรับจัดการเครดิตเทอม (เตือนความจำ)
/// ใช้ใน PO เพื่อแสดงจำนวนวันเครดิตและวันครบกำหนดชำระ
/// แก้ช่องใดช่องหนึ่ง → คำนวณอีกช่องใหม่อัตโนมัติ
class CreditTermsSectionWidget extends StatefulWidget {
  final TransactionModel screenData;
  final Function(void Function()) parentSetState;

  const CreditTermsSectionWidget({
    super.key,
    required this.screenData,
    required this.parentSetState,
  });

  @override
  State<CreditTermsSectionWidget> createState() => _CreditTermsSectionWidgetState();
}

class _CreditTermsSectionWidgetState extends State<CreditTermsSectionWidget>
    with global.ThemeRefreshMixin {
  static Color get _primaryColor => global.theme.primaryColor;
  late TextEditingController _creditDaysController;

  TransactionModel get screenData => widget.screenData;

  @override
  void initState() {
    super.initState();
    final days = screenData.creditdays ?? 0;
    _creditDaysController = TextEditingController(
      text: days > 0 ? days.toString() : '',
    );
  }

  @override
  void didUpdateWidget(covariant CreditTermsSectionWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    // ซิงค์ controller เมื่อ screenData เปลี่ยนจากภายนอก
    final days = screenData.creditdays ?? 0;
    final currentText = _creditDaysController.text;
    final expectedText = days > 0 ? days.toString() : '';
    if (currentText != expectedText) {
      _creditDaysController.text = expectedText;
    }
  }

  @override
  void dispose() {
    _creditDaysController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Row(
            children: [
              Icon(Icons.calendar_today_outlined, color: _primaryColor, size: 20),
              const SizedBox(width: 8),
              Text(
                _getLabel("credit_terms"),
                style: TextStyle(
                  color: global.theme.textColor,
                  fontWeight: FontWeight.bold,
                  fontSize: 14,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Credit Days Input + Due Date
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // จำนวนวันเครดิต
              Expanded(
                flex: 2,
                child: TextFormField(
                  controller: _creditDaysController,
                  decoration: InputDecoration(
                    labelText: _getLabel("credit_days"),
                    hintText: '0',
                    prefixIcon: Icon(Icons.timer_outlined, size: 20),
                    suffixText: _getLabel("days"),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
                  ),
                  keyboardType: TextInputType.number,
                  inputFormatters: [
                    FilteringTextInputFormatter.digitsOnly,
                  ],
                  onChanged: (value) {
                    final days = int.tryParse(value) ?? 0;
                    // แก้จำนวนวัน → คำนวณวันครบกำหนดใหม่
                    widget.parentSetState(() {
                      screenData.creditdays = days;
                      if (days > 0) {
                        final docDate = _parseDocDate();
                        if (docDate != null) {
                          final dueDate = docDate.add(Duration(days: days));
                          screenData.duedate = DateFormat('yyyy-MM-dd').format(dueDate);
                        }
                      } else {
                        screenData.duedate = '';
                      }
                    });
                  },
                ),
              ),
              const SizedBox(width: 12),

              // วันครบกำหนดชำระ (คำนวณอัตโนมัติ แต่แก้ไขได้)
              Expanded(
                flex: 3,
                child: InkWell(
                  onTap: () => _selectDueDate(context),
                  child: InputDecorator(
                    decoration: InputDecoration(
                      labelText: _getLabel("due_date"),
                      prefixIcon: Icon(Icons.event_outlined, size: 20),
                      suffixIcon: Icon(Icons.edit_calendar, size: 18, color: global.theme.iconSecondaryColor),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
                    ),
                    child: Text(
                      _formatDueDate(),
                      style: TextStyle(
                        fontSize: 14,
                        color: (screenData.duedate ?? '').isNotEmpty
                            ? global.theme.textColor
                            : global.theme.textSecondaryColor,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),

          // แสดงสรุปเครดิตเทอม
          if ((screenData.creditdays ?? 0) > 0) ...[
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: _primaryColor.withValues(alpha: 0.05),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: _primaryColor.withValues(alpha: 0.2)),
              ),
              child: Row(
                children: [
                  Icon(Icons.info_outline, color: _primaryColor, size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '${_getLabel("credit_terms")}: ${screenData.creditdays} ${_getLabel("days")}${_formatDueDateSummary()}',
                      style: TextStyle(
                        fontSize: 13,
                        color: _primaryColor,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  /// แปลง docdate จาก screenData เป็น DateTime
  DateTime? _parseDocDate() {
    try {
      if (screenData.docdatetime.isNotEmpty) {
        return DateTime.parse(screenData.docdatetime);
      }
    } catch (_) {}
    return null;
  }

  /// แสดงวันครบกำหนดชำระในรูปแบบที่อ่านง่าย
  String _formatDueDate() {
    if ((screenData.duedate ?? '').isEmpty) {
      return _getLabel("no_due_date");
    }
    try {
      final date = DateTime.parse(screenData.duedate!);
      if (global.profileData.yeartype == "buddhist") {
        return global.dateTimeBuddhist(date, format: global.DateTimeFormatEnum.dateDay);
      }
      return DateFormat('dd/MM/yyyy').format(date);
    } catch (_) {
      return screenData.duedate ?? '';
    }
  }

  /// สรุปวันครบกำหนดสำหรับ info box
  String _formatDueDateSummary() {
    if ((screenData.duedate ?? '').isEmpty) return '';
    try {
      final date = DateTime.parse(screenData.duedate!);
      String formattedDate;
      if (global.profileData.yeartype == "buddhist") {
        formattedDate = global.dateTimeBuddhist(date, format: global.DateTimeFormatEnum.dateDay);
      } else {
        formattedDate = DateFormat('dd/MM/yyyy').format(date);
      }
      return ' — ${_getLabel("due_date")}: $formattedDate';
    } catch (_) {
      return '';
    }
  }

  /// เปิด Date Picker สำหรับเลือกวันครบกำหนด (ใช้ CustomDatePicker)
  Future<void> _selectDueDate(BuildContext context) async {
    final initialDate = (screenData.duedate ?? '').isNotEmpty
        ? DateTime.tryParse(screenData.duedate!) ?? DateTime.now()
        : DateTime.now();

    final pickedDate = await showDialog<DateTime>(
      context: context,
      barrierColor: Colors.black.withValues(alpha: 0.4),
      builder: (BuildContext context) {
        return Dialog(
          backgroundColor: Colors.transparent,
          child: CustomDatePickerDialog(
            initialDate: initialDate,
            firstDate: DateTime(2020),
            lastDate: DateTime(2099),
            // อ่านจาก global.getYearType() อัตโนมัติ (ไม่ต้องส่ง parameter)
          ),
        );
      },
    );

    if (pickedDate != null) {
      // เลือกวันครบกำหนด → คำนวณจำนวนวันเครดิตใหม่
      widget.parentSetState(() {
        screenData.duedate = DateFormat('yyyy-MM-dd').format(pickedDate);
        final docDate = _parseDocDate();
        if (docDate != null) {
          final diff = pickedDate.difference(DateTime(docDate.year, docDate.month, docDate.day)).inDays;
          screenData.creditdays = diff > 0 ? diff : 0;
        }
      });
      // อัพเดต TextEditingController ให้แสดงค่าใหม่
      final days = screenData.creditdays ?? 0;
      _creditDaysController.text = days > 0 ? days.toString() : '';
    }
  }

  /// Helper method สำหรับดึงข้อความ label พร้อม fallback ภาษาไทย
  String _getLabel(String key) {
    Map<String, String> fallbackLabels = {
      'credit_terms': global.language('credit_terms'),
      'credit_days': global.language('credit_day'),
      'days': 'วัน',
      'due_date': 'วันครบกำหนดชำระ',
      'no_due_date': global.language('no_due_date'),
    };

    final translated = global.language(key);
    if (translated == key) {
      return fallbackLabels[key] ?? key;
    }
    return translated;
  }
}
