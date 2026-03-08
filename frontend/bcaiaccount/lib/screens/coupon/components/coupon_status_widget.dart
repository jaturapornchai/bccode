import 'package:flutter/material.dart';
import '../../../global.dart' as global;

/// Widget สำหรับจัดการสถานะและตัวเลือกการใช้งานคูปอง
/// - สถานะเปิด/ปิด
/// - ใช้ได้ครั้งเดียว/หลายครั้ง
class CouponStatusWidget extends StatelessWidget {
  final int status;
  final bool isOneTimeUse;
  final Function(int) onStatusChanged;
  final Function(bool) onOneTimeUseChanged;

  const CouponStatusWidget({
    super.key,
    required this.status,
    required this.isOneTimeUse,
    required this.onStatusChanged,
    required this.onOneTimeUseChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 16),
        // สถานะการใช้งาน
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.grey[50],
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: Colors.grey[200]!),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language("coupon_status_options"),
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w500,
                  color: Colors.grey[700],
                ),
              ),
              SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: Row(
                      children: [
                        Switch(
                          value: status == 0,
                          onChanged: (value) => onStatusChanged(value ? 0 : 1),
                          activeThumbColor: Colors.green,
                        ),
                        SizedBox(width: 6),
                        Text(
                          status == 0
                              ? global.language("coupon_status_active")
                              : global.language("coupon_status_inactive"),
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            color: status == 0 ? Colors.green : Colors.grey,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: Row(
                      children: [
                        Switch(
                          value: isOneTimeUse,
                          onChanged: onOneTimeUseChanged,
                          activeThumbColor: Colors.orange,
                        ),
                        SizedBox(width: 6),
                        Expanded(
                          child: Text(
                            isOneTimeUse
                                ? global.language("coupon_one_time_use")
                                : global.language("coupon_multiple_use"),
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w500,
                              color: isOneTimeUse ? Colors.orange : Colors.grey,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }
}
