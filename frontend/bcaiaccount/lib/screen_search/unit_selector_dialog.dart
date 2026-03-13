import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/modules/cart-system/models/product_search_model.dart';

/// Dialog สำหรับเลือกหน่วยนับเมื่อสินค้ามีหลายหน่วย
/// ใช้ใน PO search - เมื่อกดเลือกสินค้าที่มีหลายหน่วยนับ
class UnitSelectorDialog extends StatelessWidget {
  final String itemCode;
  final String productName;
  final List<ProductSearchModel> units; // List ของข้อมูลหน่วยนับจาก database

  const UnitSelectorDialog({
    super.key,
    required this.itemCode,
    required this.productName,
    required this.units,
  });

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
      ),
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500, maxHeight: 600),
        padding: EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Row(
              children: [
                Icon(Icons.inventory_2, color: global.theme.primaryColor, size: 24),
                SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        global.language("select_unit"),
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: global.theme.primaryColor,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        productName,
                        style: TextStyle(
                          fontSize: 14,
                          color: global.theme.textColor,
                        ),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                      Text(
                        itemCode,
                        style: TextStyle(
                          fontSize: 12,
                          color: global.theme.iconSecondaryColor,
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                  tooltip: global.language("close"),
                ),
              ],
            ),
            const Divider(height: 30),

            // Unit List
            Flexible(
              child: ListView.separated(
                shrinkWrap: true,
                itemCount: units.length,
                separatorBuilder: (context, index) => const Divider(height: 1),
                itemBuilder: (context, index) {
                  final unit = units[index];

                  // จัดรูปแบบอัตราส่วนเป็นจำนวนเต็ม
                  final ratioStand = unit.unitStand.toInt();
                  final ratioDivide = unit.unitDive.toInt();
                  final ratioText = '$ratioStand:$ratioDivide';

                  return InkWell(
                    onTap: () {
                      // Return ProductSearchModel ทั้งก้อน (มี barcode อยู่แล้ว)
                      Navigator.pop(context, unit);
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      child: Row(
                        children: [
                          // Unit Icon
                          Container(
                            width: 44,
                            height: 44,
                            decoration: BoxDecoration(
                              color: global.theme.primaryColor,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Icon(
                              Icons.category,
                              color: global.theme.primaryColor,
                              size: 22,
                            ),
                          ),
                          const SizedBox(width: 14),

                          // Unit Info - แสดง ชื่อหน่วยนับ, barcode, รหัส, อัตราส่วน
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                // ชื่อหน่วยนับ
                                Text(
                                  unit.unitName,
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                    color: global.theme.textColor,
                                  ),
                                ),
                                const SizedBox(height: 6),

                                // Barcode
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 3,
                                  ),
                                  decoration: BoxDecoration(
                                    color: global.theme.primaryColor,
                                    borderRadius: BorderRadius.circular(4),
                                    border: Border.all(color: global.theme.primaryColor),
                                  ),
                                  child: Text(
                                    unit.barcode,
                                    style: TextStyle(
                                      fontSize: 13,
                                      color: global.theme.primaryColor,
                                      fontWeight: FontWeight.w600,
                                      letterSpacing: 0.5,
                                    ),
                                  ),
                                ),
                                const SizedBox(height: 6),

                                // รหัสหน่วยนับ และ อัตราส่วน
                                Row(
                                  children: [
                                    // รหัสหน่วยนับ
                                    Container(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 6,
                                        vertical: 2,
                                      ),
                                      decoration: BoxDecoration(
                                        color: global.theme.surfaceColor,
                                        borderRadius: BorderRadius.circular(4),
                                        border: Border.all(color: global.theme.dividerBorderColor),
                                      ),
                                      child: Text(
                                        unit.unitCode,
                                        style: TextStyle(
                                          fontSize: 11,
                                          color: global.theme.textColor,
                                          fontWeight: FontWeight.w500,
                                        ),
                                      ),
                                    ),
                                    const SizedBox(width: 8),

                                    // อัตราส่วน
                                    Text(
                                      '${global.language("ratio")}: $ratioText',
                                      style: TextStyle(
                                        fontSize: 13,
                                        color: global.theme.iconSecondaryColor,
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ),

                          // Arrow Icon
                          Icon(
                            Icons.arrow_forward_ios,
                            size: 16,
                            color: global.theme.iconSecondaryColor,
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),

            const SizedBox(height: 16),

            // Cancel Button
            SizedBox(
              width: double.infinity,
              child: OutlinedButton(
                onPressed: () => Navigator.pop(context),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                ),
                child: Text(global.language("cancel")),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
