import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class TaxIdWidget extends StatelessWidget {
  final String? taxId;
  final bool Function() disableBranch;
  final bool isEditMode;
  final ValueChanged<String> onChanged;
  final VoidCallback onSubmitted;
  final FocusNode focusNode;

  const TaxIdWidget({
    super.key,
    required this.taxId,
    required this.disableBranch,
    required this.isEditMode,
    required this.onChanged,
    required this.onSubmitted,
    required this.focusNode,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
      child: TextField(
        enabled: disableBranch(),
        readOnly: !isEditMode,
        onChanged: onChanged,
        onSubmitted: (value) => onSubmitted(),
        focusNode: focusNode,
        textAlign: TextAlign.left,
        controller: TextEditingController(text: taxId),
        decoration: InputDecoration(
          floatingLabelBehavior: FloatingLabelBehavior.always,
          border: OutlineInputBorder(),
          labelText: global.language("company_tax_id"),
        ),
      ),
    );
  }
}
