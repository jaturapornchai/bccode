// ignore_for_file: depend_on_referenced_packages

import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// Widget สำหรับแสดง TextField ที่จัดรูปแบบตัวเลข
/// ใช้สำหรับแสดงยอดรวม, ราคา, จำนวน ฯลฯ ในหน้า Transaction
class TotalTextController extends StatelessWidget {
  const TotalTextController({
    super.key,
    required this.readOnly,
    required this.title,
    required this.data,
    this.icon,
    required this.onChanged,
    this.useColor,
  });

  final dynamic data;
  final String title;
  final Icon? icon;
  final Function(String) onChanged;
  final bool readOnly;
  final bool? useColor;

  @override
  Widget build(BuildContext context) {
    String initialText = '';
    if (data != null) {
      if (data is String) {
        try {
          // Attempt to parse the string as a double
          final double parsedValue = double.parse(data);
          initialText = global.formatNumber(parsedValue);
        } catch (e) {
          // Handle the case where parsing fails (e.g., if the string is not a valid number)
          initialText =
              ''; // Set to an empty string or any other default value you prefer
        }
      } else if (data is double) {
        // If data is already a double, just format it
        initialText = global.formatNumber(data);
      } else if (data is int) {
        // Handle int type as well
        initialText = global.formatNumber(data.toDouble());
      }
    }
    return TextField(
      readOnly: readOnly,
      decoration: InputDecoration(
        border: const OutlineInputBorder(),
        labelText: title,
        floatingLabelBehavior: FloatingLabelBehavior.always,
        suffixIcon: icon,
        filled: useColor,
        fillColor: Colors.yellow[100],
      ),
      controller: TextEditingController(text: initialText),
      onChanged: onChanged,
      keyboardType: const TextInputType.numberWithOptions(decimal: true),
      inputFormatters: [global.NumberInputFormatter()],
    );
  }
}
