import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// ข้อมูลประเภทธุรกิจแต่ละตัว
class _BizType {
  final String key;
  final IconData icon;
  final Color color;
  final String labelKey;
  final String descKey;
  final bool value;
  final ValueChanged<bool> onChanged;

  const _BizType({
    required this.key,
    required this.icon,
    required this.color,
    required this.labelKey,
    required this.descKey,
    required this.value,
    required this.onChanged,
  });
}

class BusinessPropertyWidget extends StatelessWidget {
  final bool isRestaurant;
  final bool isTire;
  final bool isAgriculture;
  final bool isPharmacy;
  final bool isService;
  final bool isManufacturing;
  final bool isImportExport;
  final bool isContractor;
  final bool isRental;
  final bool isWholesale;
  final bool isEcommerce;
  final bool isLogistics;
  final bool isEducation;
  final bool isHotel;
  final bool isBeauty;
  final bool isGoldShop;
  final bool isAccountingFirm;
  final bool isRetail;
  final bool isConstruction;
  final bool isElectronics;
  final bool isMobileShop;
  final ValueChanged<bool> onRestaurantChanged;
  final ValueChanged<bool> onTireChanged;
  final ValueChanged<bool> onAgricultureChanged;
  final ValueChanged<bool> onPharmacyChanged;
  final ValueChanged<bool> onServiceChanged;
  final ValueChanged<bool> onManufacturingChanged;
  final ValueChanged<bool> onImportExportChanged;
  final ValueChanged<bool> onContractorChanged;
  final ValueChanged<bool> onRentalChanged;
  final ValueChanged<bool> onWholesaleChanged;
  final ValueChanged<bool> onEcommerceChanged;
  final ValueChanged<bool> onLogisticsChanged;
  final ValueChanged<bool> onEducationChanged;
  final ValueChanged<bool> onHotelChanged;
  final ValueChanged<bool> onBeautyChanged;
  final ValueChanged<bool> onGoldShopChanged;
  final ValueChanged<bool> onAccountingFirmChanged;
  final ValueChanged<bool> onRetailChanged;
  final ValueChanged<bool> onConstructionChanged;
  final ValueChanged<bool> onElectronicsChanged;
  final ValueChanged<bool> onMobileShopChanged;

  const BusinessPropertyWidget({
    super.key,
    required this.isRestaurant,
    this.isTire = false,
    this.isAgriculture = false,
    this.isPharmacy = false,
    this.isService = false,
    this.isManufacturing = false,
    this.isImportExport = false,
    this.isContractor = false,
    this.isRental = false,
    this.isWholesale = false,
    this.isEcommerce = false,
    this.isLogistics = false,
    this.isEducation = false,
    this.isHotel = false,
    this.isBeauty = false,
    this.isGoldShop = false,
    this.isAccountingFirm = false,
    this.isRetail = false,
    this.isConstruction = false,
    this.isElectronics = false,
    this.isMobileShop = false,
    required this.onRestaurantChanged,
    ValueChanged<bool>? onTireChanged,
    ValueChanged<bool>? onAgricultureChanged,
    ValueChanged<bool>? onPharmacyChanged,
    ValueChanged<bool>? onServiceChanged,
    ValueChanged<bool>? onManufacturingChanged,
    ValueChanged<bool>? onImportExportChanged,
    ValueChanged<bool>? onContractorChanged,
    ValueChanged<bool>? onRentalChanged,
    ValueChanged<bool>? onWholesaleChanged,
    ValueChanged<bool>? onEcommerceChanged,
    ValueChanged<bool>? onLogisticsChanged,
    ValueChanged<bool>? onEducationChanged,
    ValueChanged<bool>? onHotelChanged,
    ValueChanged<bool>? onBeautyChanged,
    ValueChanged<bool>? onGoldShopChanged,
    ValueChanged<bool>? onAccountingFirmChanged,
    ValueChanged<bool>? onRetailChanged,
    ValueChanged<bool>? onConstructionChanged,
    ValueChanged<bool>? onElectronicsChanged,
    ValueChanged<bool>? onMobileShopChanged,
  })  : onTireChanged = onTireChanged ?? _noop,
        onAgricultureChanged = onAgricultureChanged ?? _noop,
        onPharmacyChanged = onPharmacyChanged ?? _noop,
        onServiceChanged = onServiceChanged ?? _noop,
        onManufacturingChanged = onManufacturingChanged ?? _noop,
        onImportExportChanged = onImportExportChanged ?? _noop,
        onContractorChanged = onContractorChanged ?? _noop,
        onRentalChanged = onRentalChanged ?? _noop,
        onWholesaleChanged = onWholesaleChanged ?? _noop,
        onEcommerceChanged = onEcommerceChanged ?? _noop,
        onLogisticsChanged = onLogisticsChanged ?? _noop,
        onEducationChanged = onEducationChanged ?? _noop,
        onHotelChanged = onHotelChanged ?? _noop,
        onBeautyChanged = onBeautyChanged ?? _noop,
        onGoldShopChanged = onGoldShopChanged ?? _noop,
        onAccountingFirmChanged = onAccountingFirmChanged ?? _noop,
        onRetailChanged = onRetailChanged ?? _noop,
        onConstructionChanged = onConstructionChanged ?? _noop,
        onElectronicsChanged = onElectronicsChanged ?? _noop,
        onMobileShopChanged = onMobileShopChanged ?? _noop;

  static void _noop(bool _) {}

  List<_BizType> _buildItems() => [
        // กลุ่ม: ค้าขาย
        _BizType(key: 'retail', icon: Icons.storefront_rounded, color: const Color(0xFF43A047), labelKey: 'business_retail', descKey: 'business_retail_desc', value: isRetail, onChanged: onRetailChanged),
        _BizType(key: 'wholesale', icon: Icons.shopping_cart_rounded, color: const Color(0xFF2196F3), labelKey: 'business_wholesale', descKey: 'business_wholesale_desc', value: isWholesale, onChanged: onWholesaleChanged),
        _BizType(key: 'ecommerce', icon: Icons.language_rounded, color: const Color(0xFF3F51B5), labelKey: 'business_ecommerce', descKey: 'business_ecommerce_desc', value: isEcommerce, onChanged: onEcommerceChanged),
        // กลุ่ม: อาหาร & บริการ
        _BizType(key: 'restaurant', icon: Icons.restaurant_rounded, color: const Color(0xFFFF9800), labelKey: 'business_restaurant', descKey: 'business_restaurant_desc', value: isRestaurant, onChanged: onRestaurantChanged),
        _BizType(key: 'service', icon: Icons.build_rounded, color: const Color(0xFF4CAF50), labelKey: 'business_service', descKey: 'business_service_desc', value: isService, onChanged: onServiceChanged),
        _BizType(key: 'hotel', icon: Icons.hotel_rounded, color: const Color(0xFF009688), labelKey: 'business_hotel', descKey: 'business_hotel_desc', value: isHotel, onChanged: onHotelChanged),
        _BizType(key: 'beauty', icon: Icons.spa_rounded, color: const Color(0xFFE91E63), labelKey: 'business_beauty', descKey: 'business_beauty_desc', value: isBeauty, onChanged: onBeautyChanged),
        // กลุ่ม: ผลิต & อุตสาหกรรม
        _BizType(key: 'manufacturing', icon: Icons.factory_rounded, color: const Color(0xFF9C27B0), labelKey: 'business_manufacturing', descKey: 'business_manufacturing_desc', value: isManufacturing, onChanged: onManufacturingChanged),
        _BizType(key: 'contractor', icon: Icons.construction_rounded, color: const Color(0xFF795548), labelKey: 'business_contractor', descKey: 'business_contractor_desc', value: isContractor, onChanged: onContractorChanged),
        // กลุ่ม: นำเข้า-ส่งออก & ขนส่ง
        _BizType(key: 'import_export', icon: Icons.directions_boat_rounded, color: const Color(0xFF00BCD4), labelKey: 'business_import_export', descKey: 'business_import_export_desc', value: isImportExport, onChanged: onImportExportChanged),
        _BizType(key: 'logistics', icon: Icons.local_shipping_rounded, color: const Color(0xFF5C6BC0), labelKey: 'business_logistics', descKey: 'business_logistics_desc', value: isLogistics, onChanged: onLogisticsChanged),
        // กลุ่ม: อสังหาฯ & เกษตร & เฉพาะทาง
        _BizType(key: 'rental', icon: Icons.home_work_rounded, color: const Color(0xFF607D8B), labelKey: 'business_rental', descKey: 'business_rental_desc', value: isRental, onChanged: onRentalChanged),
        _BizType(key: 'agriculture', icon: Icons.grass_rounded, color: const Color(0xFF8BC34A), labelKey: 'business_agriculture', descKey: 'business_agriculture_desc', value: isAgriculture, onChanged: onAgricultureChanged),
        _BizType(key: 'education', icon: Icons.school_rounded, color: const Color(0xFFFF5722), labelKey: 'business_education', descKey: 'business_education_desc', value: isEducation, onChanged: onEducationChanged),
        _BizType(key: 'pharmacy', icon: Icons.local_pharmacy_rounded, color: const Color(0xFFE91E63), labelKey: 'business_pharmacy', descKey: 'business_pharmacy_desc', value: isPharmacy, onChanged: onPharmacyChanged),
        _BizType(key: 'tire', icon: Icons.tire_repair_rounded, color: const Color(0xFF455A64), labelKey: 'business_tire', descKey: 'business_tire_desc', value: isTire, onChanged: onTireChanged),
        _BizType(key: 'gold_shop', icon: Icons.diamond_rounded, color: const Color(0xFFFFB300), labelKey: 'business_gold_shop', descKey: 'business_gold_shop_desc', value: isGoldShop, onChanged: onGoldShopChanged),
        _BizType(key: 'accounting_firm', icon: Icons.account_balance_rounded, color: const Color(0xFF37474F), labelKey: 'business_accounting_firm', descKey: 'business_accounting_firm_desc', value: isAccountingFirm, onChanged: onAccountingFirmChanged),
        // กลุ่ม: เฉพาะทาง (จาก bcaccount.com)
        _BizType(key: 'construction', icon: Icons.hardware_rounded, color: const Color(0xFFEF6C00), labelKey: 'business_construction', descKey: 'business_construction_desc', value: isConstruction, onChanged: onConstructionChanged),
        _BizType(key: 'electronics', icon: Icons.electrical_services_rounded, color: const Color(0xFF1565C0), labelKey: 'business_electronics', descKey: 'business_electronics_desc', value: isElectronics, onChanged: onElectronicsChanged),
        _BizType(key: 'mobile_shop', icon: Icons.phone_android_rounded, color: const Color(0xFF00897B), labelKey: 'business_mobile_shop', descKey: 'business_mobile_shop_desc', value: isMobileShop, onChanged: onMobileShopChanged),
      ];

  @override
  Widget build(BuildContext context) {
    final items = _buildItems();
    final selectedItems = items.where((i) => i.value).toList();

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header + count
          Row(
            children: [
              Text(
                global.language("business_property"),
                style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: global.theme.formLabelColor),
              ),
              if (selectedItems.isNotEmpty) ...[
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(color: global.theme.primaryColor.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(10)),
                  child: Text(
                    '${selectedItems.length} ${global.language("selected")}',
                    style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: global.theme.primaryColor),
                  ),
                ),
              ],
            ],
          ),
          const SizedBox(height: 10),

          // สรุปที่เลือก (แสดงเมื่อเลือก > 0)
          if (selectedItems.isNotEmpty) ...[
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: global.theme.positiveHighlightColor,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Wrap(
                spacing: 6,
                runSpacing: 6,
                children: selectedItems.map((item) {
                  final label = global.language(item.labelKey);
                  return Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                    decoration: BoxDecoration(
                      color: item.color.withValues(alpha: 0.15),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: item.color.withValues(alpha: 0.3)),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(item.icon, size: 14, color: item.color),
                        const SizedBox(width: 4),
                        Text(label, style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: item.color)),
                      ],
                    ),
                  );
                }).toList(),
              ),
            ),
            const SizedBox(height: 12),
          ],

          // Grid ทั้งหมด
          ...items.map((item) => _BusinessPropertyTile(item: item)),
        ],
      ),
    );
  }
}

/// Tile สำหรับแต่ละประเภท — แสดง icon + ชื่อ + คำอธิบาย + toggle
class _BusinessPropertyTile extends StatefulWidget {
  final _BizType item;
  const _BusinessPropertyTile({required this.item});

  @override
  State<_BusinessPropertyTile> createState() => _BusinessPropertyTileState();
}

class _BusinessPropertyTileState extends State<_BusinessPropertyTile> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final item = widget.item;
    final isActive = item.value;
    final label = global.language(item.labelKey);
    final desc = global.language(item.descKey);

    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: () => item.onChanged(!isActive),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 180),
          margin: const EdgeInsets.only(bottom: 6),
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: isActive ? item.color.withValues(alpha: 0.06) : (_isHovered ? global.theme.surfaceColor : Colors.transparent),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: isActive ? item.color.withValues(alpha: 0.3) : (_isHovered ? global.theme.dividerBorderColor : Colors.transparent),
            ),
          ),
          child: Row(
            children: [
              // Icon
              AnimatedContainer(
                duration: const Duration(milliseconds: 180),
                width: 36,
                height: 36,
                decoration: BoxDecoration(
                  color: isActive ? item.color : item.color.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Icon(item.icon, size: 18, color: isActive ? Colors.white : item.color),
              ),
              const SizedBox(width: 12),
              // Label + desc
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      label,
                      style: TextStyle(fontSize: 13, fontWeight: isActive ? FontWeight.w600 : FontWeight.w500, color: isActive ? item.color : global.theme.textColor),
                    ),
                    if (desc.isNotEmpty && desc != item.descKey)
                      Text(
                        desc,
                        style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor, height: 1.3),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                  ],
                ),
              ),
              // Toggle
              AnimatedContainer(
                duration: const Duration(milliseconds: 180),
                width: 22,
                height: 22,
                decoration: BoxDecoration(
                  color: isActive ? item.color : Colors.transparent,
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(color: isActive ? item.color : global.theme.formBorderColor, width: isActive ? 0 : 1.5),
                ),
                child: isActive ? const Icon(Icons.check_rounded, size: 16, color: Colors.white) : null,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
