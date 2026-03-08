import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class CouponUsageTypeWidget extends StatelessWidget {
  final int? couponUseType;
  final Function(int?) onChanged;

  const CouponUsageTypeWidget({
    super.key,
    required this.couponUseType,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            decoration: BoxDecoration(
              border: Border.all(color: Colors.blue),
              borderRadius: BorderRadius.circular(5),
              color: Colors.white,
            ),
            child: Column(
              children: [
                Padding(
                  padding: EdgeInsets.only(top: 8.0, left: 8.0),
                  child: Text(
                    global.language("coupon_usage_type"),
                    style: const TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                RadioListTile<int>(
                  title: Text(global.language("coupon_multiple_use")),
                  subtitle: Text(
                    global.language("coupon_can_use_multiple"),
                    style: const TextStyle(fontSize: 12, color: Colors.grey),
                  ),
                  value: 0,
                  groupValue: couponUseType ?? 0,
                  onChanged: onChanged,
                ),
                Divider(height: 1),
                RadioListTile<int>(
                  title: Text(global.language("coupon_single_use")),
                  subtitle: Text(
                    global.language("coupon_can_use_once"),
                    style: const TextStyle(fontSize: 12, color: Colors.grey),
                  ),
                  value: 1,
                  groupValue: couponUseType ?? 0,
                  onChanged: onChanged,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
