import 'package:flutter/material.dart';
import '../../../global.dart' as global;
import '../../../model/coupon_model.dart';

/// Widget สำหรับแสดงรายการคูปองแต่ละรายการ
class CouponListItemWidget extends StatelessWidget {
  final CouponModel coupon;
  final bool isSelected;
  final double fontSize;
  final double lineSpace;
  final VoidCallback onTap;
  final VoidCallback onDoubleTap;

  const CouponListItemWidget({
    super.key,
    required this.coupon,
    required this.isSelected,
    required this.fontSize,
    required this.lineSpace,
    required this.onTap,
    required this.onDoubleTap,
  });

  String _getCouponTypeName() {
    switch (coupon.coupontype) {
      case 0:
        return global.language("fixed_amount");
      case 1:
        return global.language("percentage");
      case 2:
        return global.language("cash_discount");
      default:
        return global.language("fixed_amount");
    }
  }

  String _getStatusName() {
    return (coupon.status == 0)
        ? global.language("active")
        : global.language("inactive");
  }

  @override
  Widget build(BuildContext context) {
    TextStyle textStyle = TextStyle(
      fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
      fontSize: isSelected ? fontSize + 2.0 : fontSize,
    );

    return GestureDetector(
      onTap: onTap,
      onDoubleTap: onDoubleTap,
      child: Container(
        decoration: BoxDecoration(
          color: isSelected ? Colors.cyan[100] : Colors.white,
        ),
        padding: EdgeInsets.only(
          left: 10,
          right: 10,
          top: lineSpace,
          bottom: lineSpace,
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              flex: 4,
              child: Text(
                coupon.couponcode ?? "",
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 6,
              child: Text(
                global.packName(coupon.names ?? []),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 3,
              child: Text(
                _getCouponTypeName(),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 2,
              child: Text(
                _getStatusName(),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
