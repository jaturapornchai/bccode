import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;

class CouponWidget extends StatefulWidget {
  final List<CouponModel> couPons;
  final Function(int) onCouponDeleted;
  final Function() onAmountChanged;
  final bool isReadOnly;

  const CouponWidget({
    super.key,
    required this.couPons,
    required this.onCouponDeleted,
    required this.onAmountChanged,
    this.isReadOnly = false,
  });

  @override
  State<CouponWidget> createState() => _CouponWidgetState();
}

class _CouponWidgetState extends State<CouponWidget>
    with global.ThemeRefreshMixin {
  // Helper function to get appropriate label for coupon amount field
  String _getCouponAmountLabel(String couponType) {
    switch (couponType) {
      case "0":
        return global.language("discount_value");
      case "1":
        return global.language("discount_percent");
      case "2":
        return global.language("coupon_value");
      default:
        return global.language("quantity");
    }
  }

  // Helper function to get coupon type color
  Color _getCouponTypeColor(String couponType) {
    switch (couponType) {
      case "0":
        return global.theme.infoHighlightTextColor;
      case "1":
        return global.theme.warningHighlightTextColor;
      case "2":
        return global.theme.positiveHighlightTextColor;
      default:
        return global.theme.iconSecondaryColor;
    }
  }

  // Helper function to get coupon type text
  String _getCouponTypeText(String couponType) {
    switch (couponType) {
      case "0":
        return global.language("reduce_by_value");
      case "1":
        return global.language("reduce_by_percent");
      case "2":
        return global.language("cash_equivalent");
      default:
        return global.language("not_specified");
    }
  }

  // Helper function to format coupon amount display
  String _formatCouponAmount(CouponModel coupon) {
    if (coupon.couponamount == null) return "0";

    switch (coupon.coupontype) {
      case "0":
        return "${coupon.couponamount!.toStringAsFixed(2)} ${global.language('baht')}";
      case "1":
        return "${coupon.couponamount!.toStringAsFixed(2)} ${global.language('baht')}";
      case "2":
        return "${coupon.couponamount!.toStringAsFixed(2)} ${global.language('baht')}";
      default:
        return "${coupon.couponamount!.toStringAsFixed(2)} ${global.language('baht')}";
    }
  }

  // Helper function to calculate coupon discount amount
  double _calculateCouponDiscount(CouponModel coupon, double totalAmount) {
    switch (coupon.coupontype) {
      case "0": // ลดตามมูลค่า
        return coupon.couponamount ?? 0;
      case "1": // ลดตามเปอร์เซ็นต์
        double percentage = coupon.couponamount ?? 0;
        return (totalAmount * percentage) / 100;
      case "2": // คูปองแทนเงินสด
        return coupon.couponamount ?? 0;
      default:
        return 0;
    }
  }

  // Helper function to build info rows
  Widget _buildInfoRow({
    required IconData icon,
    required String label,
    required String value,
    required Color color,
    bool isDescription = false,
    bool isAmount = false,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(4),
          ),
          child: Icon(icon, size: 16, color: color),
        ),
        const SizedBox(width: 8),
        Expanded(
          flex: 2,
          child: Text(
            label,
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w500,
              color: global.theme.textSecondaryColor,
            ),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          flex: 3,
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: isAmount ? global.theme.negativeHighlightColor : global.theme.cardColor,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(
                color: isAmount ? global.theme.negativeHighlightColor : global.theme.dividerBorderColor,
              ),
            ),
            child: Text(
              value,
              style: TextStyle(
                fontSize: isAmount ? 13 : 12,
                fontWeight: isAmount ? FontWeight.bold : FontWeight.normal,
                color: isAmount ? global.theme.negativeHighlightTextColor : global.theme.textColor,
              ),
              maxLines: isDescription ? 2 : 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ),
      ],
    );
  }

  // Build editable coupon card for editing mode
  Widget _buildEditableCouponCard(CouponModel coupon, int index) {
    return Container(child: Text(global.language("waiting_for_api")));
  }

  // Build read-only coupon card for display mode
  Widget _buildReadOnlyCouponCard(CouponModel coupon, int index) {
    return Card(
      elevation: 2,
      margin: const EdgeInsets.only(bottom: 12),
      child: Container(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            // Coupon Icon
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: _getCouponTypeColor(
                  coupon.coupontype ?? "0",
                ).withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                Icons.card_giftcard,
                size: 28,
                color: _getCouponTypeColor(coupon.coupontype ?? "0"),
              ),
            ),

            const SizedBox(width: 16),

            // Coupon Info
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Text(
                        coupon.couponno ?? '-',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: _getCouponTypeColor(coupon.coupontype ?? "0"),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text(
                          _getCouponTypeText(coupon.coupontype ?? "0"),
                          style: TextStyle(
                            fontSize: 12,
                            color: global.theme.onPrimaryColor,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  if (coupon.coupondescription != null &&
                      coupon.coupondescription!.isNotEmpty)
                    Text(
                      coupon.coupondescription!,
                      style: TextStyle(
                        fontSize: 14,
                        color: global.theme.textSecondaryColor,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                ],
              ),
            ),

            const SizedBox(width: 12),

            // Amount Display
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
              decoration: BoxDecoration(
                color: global.theme.positiveHighlightColor,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: global.theme.positiveHighlightColor),
              ),
              child: Text(
                _formatCouponAmount(coupon),
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.bold,
                  color: global.theme.positiveHighlightTextColor,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (widget.couPons.isEmpty) {
      return Container();
    }

    return Column(
      children: widget.couPons.asMap().entries.map((entry) {
        int index = entry.key;
        CouponModel coupon = entry.value;

        return widget.isReadOnly
            ? _buildReadOnlyCouponCard(coupon, index)
            : _buildEditableCouponCard(coupon, index);
      }).toList(),
    );
  }
}
