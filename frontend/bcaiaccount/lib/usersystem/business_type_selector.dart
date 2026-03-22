import 'package:flutter/material.dart';
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// Business Type data พร้อม icon + color
class BusinessTypeInfo {
  final String code;
  final IconData icon;
  final Color color;
  final String emoji;

  const BusinessTypeInfo({required this.code, required this.icon, required this.color, required this.emoji});
}

/// Mapping code → icon + color (8 ประเภทธุรกิจหลัก)
const _businessTypeInfoMap = <String, BusinessTypeInfo>{
  'retail': BusinessTypeInfo(code: 'retail', icon: Icons.shopping_cart_rounded, color: Color(0xFF2196F3), emoji: '🛒'),
  'trading': BusinessTypeInfo(code: 'trading', icon: Icons.shopping_cart_rounded, color: Color(0xFF2196F3), emoji: '🛒'),
  'service': BusinessTypeInfo(code: 'service', icon: Icons.build_rounded, color: Color(0xFF4CAF50), emoji: '🔧'),
  'manufacturing': BusinessTypeInfo(code: 'manufacturing', icon: Icons.factory_rounded, color: Color(0xFF9C27B0), emoji: '🏭'),
  'restaurant': BusinessTypeInfo(code: 'restaurant', icon: Icons.restaurant_rounded, color: Color(0xFFFF9800), emoji: '🍜'),
  'food': BusinessTypeInfo(code: 'food', icon: Icons.restaurant_rounded, color: Color(0xFFFF9800), emoji: '🍜'),
  'import_export': BusinessTypeInfo(code: 'import_export', icon: Icons.directions_boat_rounded, color: Color(0xFF00BCD4), emoji: '🚢'),
  'contractor': BusinessTypeInfo(code: 'contractor', icon: Icons.construction_rounded, color: Color(0xFF795548), emoji: '🏗️'),
  'agriculture': BusinessTypeInfo(code: 'agriculture', icon: Icons.grass_rounded, color: Color(0xFF8BC34A), emoji: '🌾'),
  'rental': BusinessTypeInfo(code: 'rental', icon: Icons.home_work_rounded, color: Color(0xFF607D8B), emoji: '🏠'),
};

/// ค้นหา info จาก code (fallback ถ้าไม่พบ)
BusinessTypeInfo getBusinessTypeInfo(String? code) {
  if (code == null || code.isEmpty) {
    return const BusinessTypeInfo(code: '', icon: Icons.store_rounded, color: Color(0xFF9E9E9E), emoji: '🏪');
  }
  final lower = code.toLowerCase();
  for (final entry in _businessTypeInfoMap.entries) {
    if (lower.contains(entry.key)) return entry.value;
  }
  return BusinessTypeInfo(code: code, icon: Icons.store_rounded, color: const Color(0xFF3949AB), emoji: '🏪');
}

/// Dialog เลือกประเภทธุรกิจ — แบบ card grid สวยๆ
/// รองรับเลือกได้ 1 อัน (single select)
Future<BusinessTypeModel?> showBusinessTypeSelector({
  required BuildContext context,
  required List<BusinessTypeModel> businessTypes,
  BusinessTypeModel? selected,
}) async {
  return showDialog<BusinessTypeModel>(
    context: context,
    builder: (ctx) => _BusinessTypeSelectorDialog(
      businessTypes: businessTypes,
      selected: selected,
    ),
  );
}

class _BusinessTypeSelectorDialog extends StatefulWidget {
  final List<BusinessTypeModel> businessTypes;
  final BusinessTypeModel? selected;

  const _BusinessTypeSelectorDialog({required this.businessTypes, this.selected});

  @override
  State<_BusinessTypeSelectorDialog> createState() => _BusinessTypeSelectorDialogState();
}

class _BusinessTypeSelectorDialogState extends State<_BusinessTypeSelectorDialog> {
  String? _selectedCode;

  @override
  void initState() {
    super.initState();
    _selectedCode = widget.selected?.code;
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final dialogWidth = screenWidth > 700 ? 600.0 : screenWidth * 0.9;

    return Dialog(
      backgroundColor: global.theme.cardColor,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: dialogWidth, maxHeight: 560),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Header
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: global.theme.primaryColor.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(Icons.category_rounded, color: global.theme.primaryColor, size: 24),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          global.language('select_business_type'),
                          style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: global.theme.textColor),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          global.language('select_business_type_desc').isEmpty
                              ? 'เลือกประเภทที่ตรงกับธุรกิจของคุณ'
                              : global.language('select_business_type_desc'),
                          style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor),
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    icon: Icon(Icons.close_rounded, color: global.theme.iconSecondaryColor),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
              const SizedBox(height: 20),
              // Grid
              Flexible(
                child: LayoutBuilder(
                  builder: (context, constraints) {
                    final crossAxisCount = constraints.maxWidth > 400 ? 3 : 2;
                    return GridView.builder(
                      shrinkWrap: true,
                      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: crossAxisCount,
                        crossAxisSpacing: 12,
                        mainAxisSpacing: 12,
                        childAspectRatio: 1.0,
                      ),
                      itemCount: widget.businessTypes.length,
                      itemBuilder: (context, index) {
                        final bt = widget.businessTypes[index];
                        final info = getBusinessTypeInfo(bt.code);
                        final isSelected = _selectedCode == bt.code;
                        final name = global.activeLangName(bt.names!);

                        return _BusinessTypeCard(
                          name: name,
                          info: info,
                          isSelected: isSelected,
                          onTap: () {
                            setState(() => _selectedCode = bt.code);
                            // Auto confirm หลังเลือก (delay นิดนึงให้เห็น animation)
                            Future.delayed(const Duration(milliseconds: 250), () {
                              if (mounted) Navigator.pop(context, bt);
                            });
                          },
                        );
                      },
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Card สำหรับแต่ละประเภทธุรกิจ — hover + press animation
class _BusinessTypeCard extends StatefulWidget {
  final String name;
  final BusinessTypeInfo info;
  final bool isSelected;
  final VoidCallback onTap;

  const _BusinessTypeCard({required this.name, required this.info, required this.isSelected, required this.onTap});

  @override
  State<_BusinessTypeCard> createState() => _BusinessTypeCardState();
}

class _BusinessTypeCardState extends State<_BusinessTypeCard> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final color = widget.info.color;
    final isSelected = widget.isSelected;

    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: widget.onTap,
        child: AnimatedScale(
          scale: isSelected ? 0.95 : (_isHovered ? 1.04 : 1.0),
          duration: const Duration(milliseconds: 150),
          curve: Curves.easeOutCubic,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            decoration: BoxDecoration(
              color: isSelected ? color.withValues(alpha: 0.12) : (_isHovered ? global.theme.surfaceColor : global.theme.cardColor),
              borderRadius: BorderRadius.circular(18),
              border: Border.all(
                color: isSelected ? color : (_isHovered ? color.withValues(alpha: 0.3) : global.theme.dividerBorderColor),
                width: isSelected ? 2.5 : 1,
              ),
              boxShadow: [
                if (isSelected || _isHovered)
                  BoxShadow(color: color.withValues(alpha: isSelected ? 0.2 : 0.1), blurRadius: 12, offset: const Offset(0, 4)),
              ],
            ),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                // Icon container
                AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  width: isSelected ? 52 : 48,
                  height: isSelected ? 52 : 48,
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: isSelected ? [color, color.withValues(alpha: 0.7)] : [color.withValues(alpha: 0.15), color.withValues(alpha: 0.08)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    borderRadius: BorderRadius.circular(14),
                    boxShadow: isSelected ? [BoxShadow(color: color.withValues(alpha: 0.3), blurRadius: 8, offset: const Offset(0, 3))] : [],
                  ),
                  child: Icon(
                    widget.info.icon,
                    size: 24,
                    color: isSelected ? Colors.white : color,
                  ),
                ),
                const SizedBox(height: 10),
                // Name
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: Text(
                    widget.name,
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                      color: isSelected ? color : global.theme.textColor,
                      height: 1.2,
                    ),
                    textAlign: TextAlign.center,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                // Checkmark
                if (isSelected) ...[
                  const SizedBox(height: 6),
                  Icon(Icons.check_circle_rounded, size: 18, color: color),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
