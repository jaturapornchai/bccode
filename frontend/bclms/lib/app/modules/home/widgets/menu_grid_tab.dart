import 'package:flutter/material.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/home/models/menu_item_model.dart';

/// Grid เมนูย่อยสำหรับแต่ละ Tab — แสดง icon + label เป็นตาราง
class MenuGridTab extends StatelessWidget {
  final List<MenuItemModel> items;
  final String? headerTitle;

  const MenuGridTab({
    super.key,
    required this.items,
    this.headerTitle,
  });

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (headerTitle != null) ...[
            Padding(
              padding: const EdgeInsets.only(left: 4, bottom: 12),
              child: Text(
                headerTitle!,
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: AppTheme.earthBrown,
                ),
              ),
            ),
          ],
          LayoutBuilder(
            builder: (context, constraints) {
              // คำนวณจำนวนคอลัมน์ตามขนาดหน้าจอ
              final crossAxisCount = constraints.maxWidth > 600 ? 4 : 3;
              return GridView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: crossAxisCount,
                  mainAxisSpacing: 12,
                  crossAxisSpacing: 12,
                  childAspectRatio: 0.95,
                ),
                itemCount: items.length,
                itemBuilder: (context, index) {
                  return _MenuGridItem(item: items[index]);
                },
              );
            },
          ),
        ],
      ),
    );
  }
}

/// แต่ละ item ใน Grid — icon กลม + label ด้านล่าง
class _MenuGridItem extends StatelessWidget {
  final MenuItemModel item;

  const _MenuGridItem({required this.item});

  @override
  Widget build(BuildContext context) {
    final color = item.isEnabled ? item.color : Colors.grey.shade400;
    final labelColor = item.isEnabled
        ? AppTheme.earthBrown
        : Colors.grey.shade400;

    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: item.isEnabled
            ? () => Navigator.of(context).pushNamed(
                  item.route,
                  arguments: {'title': item.label},
                )
            : null,
        borderRadius: BorderRadius.circular(16),
        child: Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: item.isEnabled
                  ? AppTheme.warmBeige
                  : Colors.grey.shade200,
            ),
            boxShadow: item.isEnabled
                ? [
                    BoxShadow(
                      color: color.withValues(alpha: 0.08),
                      blurRadius: 8,
                      offset: const Offset(0, 2),
                    ),
                  ]
                : null,
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // วงกลม icon
              Container(
                width: 52,
                height: 52,
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.12),
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  item.icon,
                  size: 26,
                  color: color,
                ),
              ),
              const SizedBox(height: 8),
              // label
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 4),
                child: Text(
                  item.label,
                  textAlign: TextAlign.center,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                    color: labelColor,
                    height: 1.2,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
