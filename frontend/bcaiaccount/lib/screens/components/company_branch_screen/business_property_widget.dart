import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class BusinessPropertyWidget extends StatelessWidget {
  final bool isRestaurant;
  final bool isTire;
  final bool isAgriculture;
  final bool isPharmacy;
  final ValueChanged<bool> onRestaurantChanged;
  final ValueChanged<bool> onTireChanged;
  final ValueChanged<bool> onAgricultureChanged;
  final ValueChanged<bool> onPharmacyChanged;

  const BusinessPropertyWidget({
    super.key,
    required this.isRestaurant,
    required this.isTire,
    required this.isAgriculture,
    required this.isPharmacy,
    required this.onRestaurantChanged,
    required this.onTireChanged,
    required this.onAgricultureChanged,
    required this.onPharmacyChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            global.language("business_property"),
            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
          ),
          const SizedBox(height: 4),
          _buildCheckbox(
            label: global.language("business_restaurant"),
            value: isRestaurant,
            onChanged: onRestaurantChanged,
          ),
          _buildCheckbox(
            label: global.language("business_tire"),
            value: isTire,
            onChanged: onTireChanged,
          ),
          _buildCheckbox(
            label: global.language("business_agriculture"),
            value: isAgriculture,
            onChanged: onAgricultureChanged,
          ),
          _buildCheckbox(
            label: global.language("business_pharmacy"),
            value: isPharmacy,
            onChanged: onPharmacyChanged,
          ),
        ],
      ),
    );
  }

  Widget _buildCheckbox({
    required String label,
    required bool value,
    required ValueChanged<bool> onChanged,
  }) {
    return InkWell(
      onTap: () => onChanged(!value),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Checkbox(
            value: value,
            onChanged: (v) => onChanged(v ?? false),
            activeColor: global.theme.primaryColor,
          ),
          Text(label, style: TextStyle(fontSize: 14)),
        ],
      ),
    );
  }
}
