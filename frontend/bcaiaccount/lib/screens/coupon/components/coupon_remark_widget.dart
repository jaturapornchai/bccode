import 'package:flutter/material.dart';
import '../../../global.dart' as global;

/// Widget สำหรับกรอกหมายเหตุของคูปอง
class CouponRemarkWidget extends StatelessWidget {
  final TextEditingController remarkController;
  final FocusNode remarkFocusNode;
  final Function(String) onChanged;

  const CouponRemarkWidget({
    super.key,
    required this.remarkController,
    required this.remarkFocusNode,
    required this.onChanged,
  });

  Widget _buildSectionHeader(String title, IconData icon) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 16),
      decoration: BoxDecoration(
        color: global.theme.appBarColor.withOpacity(0.08),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: global.theme.appBarColor.withOpacity(0.2)),
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

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),

        // Section Header
        _buildSectionHeader(
          global.language("coupon_section_remark"),
          Icons.note,
        ),
        const SizedBox(height: 16),

        // Remark Field
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              global.language("coupon_remark_label"),
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w500,
                color: Colors.grey[700],
              ),
            ),
            SizedBox(height: 8),
            Container(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(6),
                border: Border.all(color: Colors.grey[300]!),
              ),
              child: TextFormField(
                controller: remarkController,
                focusNode: remarkFocusNode,
                maxLines: 3,
                onChanged: onChanged,
                style: TextStyle(fontSize: 16),
                decoration: InputDecoration(
                  hintText: global.language("coupon_hint_remark"),
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
            const SizedBox(height: 16),
          ],
        ),
      ],
    );
  }
}
