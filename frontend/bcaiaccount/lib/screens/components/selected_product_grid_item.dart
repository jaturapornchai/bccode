import 'package:flutter/material.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;

// รูปแบบข้อมูลสินค้าพร้อมจำนวนที่ต้องการพิมพ์
class ProductWithCopies {
  final SearchCodeAndNameAndUnitModel product;
  int copies;
  final List<LanguageDataModel>? unitname;

  ProductWithCopies({
    required this.product,
    this.copies = 1,
    this.unitname,
  });
}

class SelectedProductGridItem extends StatelessWidget {
  final ProductWithCopies product;
  final VoidCallback onCopiesChanged;
  final VoidCallback onRemove;
  final VoidCallback onEditCopies;

  const SelectedProductGridItem({
    super.key,
    required this.product,
    required this.onCopiesChanged,
    required this.onRemove,
    required this.onEditCopies,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      shadowColor: Colors.black.withOpacity(0.1),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
      ),
      child: Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          color: Colors.white,
        ),
        child: Column(
          children: [
            // Header section with barcode prominently displayed
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              decoration: BoxDecoration(
                color: Colors.indigo.shade400,
                borderRadius: const BorderRadius.only(
                  topLeft: Radius.circular(12),
                  topRight: Radius.circular(12),
                ),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      product.product.barcode,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 11,
                        fontWeight: FontWeight.w600,
                        fontFamily: 'Monospace',
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  // Delete button integrated in header
                  InkWell(
                    onTap: onRemove,
                    borderRadius: BorderRadius.circular(15),
                    child: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(15),
                      ),
                      child: Icon(
                        Icons.close,
                        size: 16,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ],
              ),
            ),

            // Main content area
            Expanded(
              child: Padding(
                padding: const EdgeInsets.all(8), // ลดจาก 12 เป็น 8
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Product name - ลดขนาดและลดระยะห่าง
                    Text(
                      global.activeLangName(product.product.name),
                      style: TextStyle(
                        fontSize: 11, // ลดจาก 12 เป็น 11
                        fontWeight: FontWeight.w600,
                        color: Colors.grey.shade800,
                        height: 1.0, // ลดจาก 1.1 เป็น 1.0
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),

                    const SizedBox(height: 4), // ลดจาก 6 เป็น 4

                    // Unit and Shelf info - ลดขนาดและระยะห่าง
                    Column(
                      children: [
                        // Unit row - ลดขนาด
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(2), // ลดจาก 3 เป็น 2
                              decoration: BoxDecoration(
                                color: Colors.blue.shade50,
                                borderRadius: BorderRadius.circular(3), // ลดจาก 4 เป็น 3
                              ),
                              child: Icon(
                                Icons.straighten,
                                size: 10, // ลดจาก 12 เป็น 10
                                color: Colors.blue.shade600,
                              ),
                            ),
                            const SizedBox(width: 4), // ลดจาก 6 เป็น 4
                            Expanded(
                              child: Text(
                                global.activeLangName(product.product.unitname),
                                style: TextStyle(
                                  fontSize: 9, // ลดจาก 10 เป็น 9
                                  fontWeight: FontWeight.w500,
                                  color: Colors.blue.shade700,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                        ),

                        const SizedBox(height: 3), // ลดจาก 4 เป็น 3

                        // Shelf row (if available) - ลดขนาด
                        if (product.product.shelfCode != null && product.product.shelfName != null)
                          Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.all(2), // ลดจาก 3 เป็น 2
                                decoration: BoxDecoration(
                                  color: Colors.orange.shade50,
                                  borderRadius: BorderRadius.circular(3), // ลดจาก 4 เป็น 3
                                ),
                                child: Icon(
                                  Icons.inventory_2,
                                  size: 10, // ลดจาก 12 เป็น 10
                                  color: Colors.orange.shade600,
                                ),
                              ),
                              const SizedBox(width: 4), // ลดจาก 6 เป็น 4
                              Expanded(
                                child: Text(
                                  '${product.product.shelfCode} - ${product.product.shelfName}',
                                  style: TextStyle(
                                    fontSize: 9, // ลดจาก 10 เป็น 9
                                    fontWeight: FontWeight.w500,
                                    color: Colors.orange.shade700,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                            ],
                          ),
                      ],
                    ),

                    const Spacer(),

                    // Quantity section - ปรับให้สวยกว่าเดิม
                    Container(
                      decoration: BoxDecoration(
                        color: Colors.blue.shade50,
                        borderRadius: BorderRadius.circular(6),
                        border: Border.all(color: Colors.blue.shade200, width: 1),
                      ),
                      child: Row(
                        children: [
                          // Decrease button
                          Expanded(
                            child: InkWell(
                              onTap: product.copies > 1
                                  ? () {
                                      product.copies--;
                                      onCopiesChanged();
                                    }
                                  : null,
                              borderRadius: const BorderRadius.only(
                                topLeft: Radius.circular(6),
                                bottomLeft: Radius.circular(6),
                              ),
                              child: Container(
                                height: 26,
                                decoration: BoxDecoration(
                                  color: product.copies > 1 ? Colors.red.shade100.withOpacity(0.7) : Colors.grey.shade100,
                                  borderRadius: const BorderRadius.only(
                                    topLeft: Radius.circular(6),
                                    bottomLeft: Radius.circular(6),
                                  ),
                                ),
                                child: Icon(
                                  Icons.remove,
                                  size: 16,
                                  color: product.copies > 1 ? Colors.red.shade600 : Colors.grey.shade400,
                                ),
                              ),
                            ),
                          ),

                          // Quantity display - ตรงกลาง
                          Container(
                            width: 50,
                            height: 26,
                            decoration: BoxDecoration(
                              color: Colors.white,
                              border: Border.symmetric(
                                vertical: BorderSide(color: Colors.blue.shade200, width: 1),
                              ),
                            ),
                            child: InkWell(
                              onTap: onEditCopies,
                              child: Center(
                                child: Text(
                                  '${product.copies}',
                                  style: TextStyle(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.blue.shade700,
                                  ),
                                ),
                              ),
                            ),
                          ),

                          // Increase button
                          Expanded(
                            child: InkWell(
                              onTap: product.copies < 99
                                  ? () {
                                      product.copies++;
                                      onCopiesChanged();
                                    }
                                  : null,
                              borderRadius: const BorderRadius.only(
                                topRight: Radius.circular(6),
                                bottomRight: Radius.circular(6),
                              ),
                              child: Container(
                                height: 26,
                                decoration: BoxDecoration(
                                  color: product.copies < 99 ? Colors.green.shade100.withOpacity(0.7) : Colors.grey.shade100,
                                  borderRadius: const BorderRadius.only(
                                    topRight: Radius.circular(6),
                                    bottomRight: Radius.circular(6),
                                  ),
                                ),
                                child: Icon(
                                  Icons.add,
                                  size: 16,
                                  color: product.copies < 99 ? Colors.green.shade600 : Colors.grey.shade400,
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
