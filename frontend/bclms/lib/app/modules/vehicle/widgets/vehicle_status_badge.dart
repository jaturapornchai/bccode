import 'package:flutter/material.dart';
import 'package:bclms/app/core/utils/global.dart';

/// Badge แสดงสถานะยานพาหนะ (0=ไม่ใช้งาน, 1=ใช้งาน, 2=ซ่อมบำรุง)
class VehicleStatusBadge extends StatelessWidget {
  final int status;

  const VehicleStatusBadge({super.key, required this.status});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: _color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _color.withValues(alpha: 0.3)),
      ),
      child: Text(
        _label,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: _color,
        ),
      ),
    );
  }

  String get _label {
    switch (status) {
      case 1:
        return 'ใช้งาน';
      case 2:
        return 'ซ่อมบำรุง';
      case 0:
        return 'ไม่ใช้งาน';
      default:
        return 'ไม่ระบุ';
    }
  }

  Color get _color {
    switch (status) {
      case 1:
        return const Color(0xFF2E7D32);
      case 2:
        return const Color(0xFFE65100);
      case 0:
        return const Color(0xFFD32F2F);
      default:
        return Colors.grey;
    }
  }
}

/// Badge แสดงวันหมดอายุเอกสาร (ประกัน, ทะเบียน)
class ExpiryBadge extends StatelessWidget {
  final String label;
  final DateTime? expiryDate;

  const ExpiryBadge({
    super.key,
    required this.label,
    required this.expiryDate,
  });

  @override
  Widget build(BuildContext context) {
    final now = DateTime.now();
    final Color color;
    final String text;

    if (expiryDate == null) {
      color = Colors.grey;
      text = '$label: ไม่ระบุ';
    } else {
      final dateStr = dateShortBuddhist(expiryDate!);
      final daysLeft = expiryDate!.difference(now).inDays;
      if (daysLeft < 0) {
        color = const Color(0xFFD32F2F);
        text = '$label: หมดอายุ ($dateStr)';
      } else if (daysLeft <= 30) {
        color = const Color(0xFFE65100);
        text = '$label: เหลือ $daysLeft วัน ($dateStr)';
      } else {
        color = const Color(0xFF2E7D32);
        text = '$label: เหลือ $daysLeft วัน ($dateStr)';
      }
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w500,
          color: color,
        ),
      ),
    );
  }
}
