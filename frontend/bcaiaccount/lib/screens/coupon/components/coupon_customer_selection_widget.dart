import 'package:flutter/material.dart';
import '../../../global.dart' as global;

/// Widget สำหรับจัดการการเลือกลูกค้าที่สามารถใช้คูปอง
class CouponCustomerSelectionWidget extends StatelessWidget {
  final List<String> customerCodes;
  final Function() onAddCustomer;
  final Function(String) onRemoveCustomer;

  const CouponCustomerSelectionWidget({
    super.key,
    required this.customerCodes,
    required this.onAddCustomer,
    required this.onRemoveCustomer,
  });

  Widget _buildSectionHeader(String title, IconData icon) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 16),
      decoration: BoxDecoration(
        color: global.theme.appBarColor.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: global.theme.appBarColor.withValues(alpha: 0.2)),
      ),
      child: Row(
        children: [
          Icon(icon, size: 20, color: global.theme.appBarColor),
          const SizedBox(width: 10),
          Text(
            title,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: global.theme.appBarColor,
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),

        // Section Header
        _buildSectionHeader(
          global.language("coupon_section_customers"),
          Icons.people,
        ),
        const SizedBox(height: 16),

        // แสดงรายการลูกค้าที่เลือก
        if (customerCodes.isNotEmpty)
          Column(
            children: [
              Container(
                decoration: BoxDecoration(
                  color: Colors.grey[50],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.grey[200]!),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Header
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.blue[600],
                        borderRadius: const BorderRadius.only(
                          topLeft: Radius.circular(8),
                          topRight: Radius.circular(8),
                        ),
                      ),
                      child: Row(
                        children: [
                          const Icon(
                            Icons.people,
                            color: Colors.white,
                            size: 18,
                          ),
                          SizedBox(width: 8),
                          Text(
                            "${global.language("coupon_selected_customers")}: ${customerCodes.length}",
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                            ),
                          ),

                          /// between --- IGNORE --
                          const Spacer(),

                          /// ปุ่มเพิ่มลูกค้า
                          ElevatedButton.icon(
                            style: ElevatedButton.styleFrom(
                              backgroundColor: global.theme.buttonColor,
                              foregroundColor: Colors.white,
                              padding: const EdgeInsets.symmetric(
                                vertical: 8,
                                horizontal: 12,
                              ),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(6),
                              ),
                              elevation: 2,
                            ),
                            onPressed: onAddCustomer,
                            icon: Icon(Icons.person_add, size: 16),
                            label: Text(
                              global.language("coupon_button_add_customer"),
                              style: const TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    // Customer List
                    Container(
                      constraints: const BoxConstraints(maxHeight: 200),
                      child: SingleChildScrollView(
                        child: Column(
                          children: customerCodes.map((customerCode) {
                            return Container(
                              decoration: BoxDecoration(
                                border: Border(
                                  bottom: BorderSide(
                                    color: Colors.grey[200]!,
                                    width: 1,
                                  ),
                                ),
                              ),
                              child: ListTile(
                                dense: true,
                                leading: CircleAvatar(
                                  radius: 16,
                                  backgroundColor: Colors.blue[100],
                                  child: Icon(
                                    Icons.person,
                                    size: 16,
                                    color: Colors.blue[700],
                                  ),
                                ),
                                title: Text(
                                  customerCode,
                                  style: const TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                                trailing: IconButton(
                                  icon: Icon(
                                    Icons.close,
                                    color: Colors.red[400],
                                    size: 18,
                                  ),
                                  onPressed: () =>
                                      onRemoveCustomer(customerCode),
                                  tooltip: global.language(
                                    "coupon_remove_customer",
                                  ),
                                ),
                              ),
                            );
                          }).toList(),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
            ],
          ),

        // // ปุ่มเพิ่มลูกค้า
        // SizedBox(
        //   width: double.infinity,
        //   child: ElevatedButton.icon(
        //     style: ElevatedButton.styleFrom(
        //       backgroundColor: global.theme.buttonColor,
        //       foregroundColor: Colors.white,
        //       padding: const EdgeInsets.symmetric(vertical: 14),
        //       shape: RoundedRectangleBorder(
        //         borderRadius: BorderRadius.circular(8),
        //       ),
        //       elevation: 2,
        //     ),
        //     onPressed: onAddCustomer,
        // icon: Icon(Icons.person_add, size: 18),
        //     label: Text(
        //       customerCodes.isEmpty
        //           ? global.language("coupon_button_select_customer")
        //           : global.language("coupon_button_add_customer"),
        //       style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
        //     ),
        //   ),
        // ),

        // แสดงข้อความช่วยเหลือ
        if (customerCodes.isEmpty)
          Padding(
            padding: EdgeInsets.only(bottom: 10),
            child: Row(
              children: [
                Icon(Icons.info_outline, size: 16, color: Colors.grey[500]),
                SizedBox(width: 6),
                Expanded(
                  child: Text(
                    global.language("coupon_customer_empty_hint"),
                    style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                  ),
                ),
              ],
            ),
          ),
      ],
    );
  }
}
