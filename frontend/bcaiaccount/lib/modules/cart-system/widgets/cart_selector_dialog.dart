import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import '../models/cart_model.dart';
import '../models/cart_system_type.dart';
import '../models/cart_type.dart';
import '../services/mongodb_cart_service.dart';
import '../../../utils/logger/app_logger.dart';

/// Dialog สำหรับเลือกตะกร้าและแก้ไขข้อมูลตะกร้า
///
/// ตาม systemguide.md:
/// - กดที่ตระกร้าเพื่อเลือกใช้งาน เมื่อกดให้ Show Dialog
/// - ปุ่ม ยืนยันการเปลี่ยนตะกร้า
/// - ปุ่ม แก้ไขชื่อตะกร้า ประเภทตะกร้า (Show Dialog)
class CartSelectorDialog extends StatefulWidget {
  final List<CartModel> allCarts;
  final CartModel? currentCart;
  final CartSystemType systemType;
  final Function(CartModel) onCartSelected;
  final Function(CartModel) onCartUpdated;

  const CartSelectorDialog({
    super.key,
    required this.allCarts,
    required this.currentCart,
    required this.systemType,
    required this.onCartSelected,
    required this.onCartUpdated,
  });

  @override
  State<CartSelectorDialog> createState() => _CartSelectorDialogState();
}

class _CartSelectorDialogState extends State<CartSelectorDialog> with global.ThemeRefreshMixin {
  CartModel? _selectedCart;

  @override
  void initState() {
    super.initState();
    _selectedCart = widget.currentCart;
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500, maxHeight: 600),
        padding: EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // หัวข้อ
            Row(
              children: [
                Icon(
                  widget.systemType == CartSystemType.sales
                      ? Icons.shopping_cart
                      : widget.systemType == CartSystemType.purchase
                      ? Icons.add_shopping_cart
                      : Icons.inventory_2,
                  color: global.theme.infoHighlightTextColor,
                  size: 28,
                ),
                SizedBox(width: 12),
                Expanded(
                  child: Text(
                    '${global.language("select_cart")} ${widget.systemType.displayNameThai}',
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.pop(context),
                  icon: Icon(Icons.close),
                ),
              ],
            ),
            const Divider(height: 24),

            // รายการตะกร้า
            Expanded(
              child: widget.allCarts.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.shopping_cart_outlined,
                            size: 64,
                            color: global.theme.iconSecondaryColor,
                          ),
                          SizedBox(height: 16),
                          Text(
                            global.language("no_cart_yet"),
                            style: TextStyle(
                              fontSize: 16,
                              color: global.theme.iconSecondaryColor,
                            ),
                          ),
                        ],
                      ),
                    )
                  : ListView.builder(
                      shrinkWrap: true,
                      itemCount: widget.allCarts.length,
                      itemBuilder: (context, index) {
                        final cart = widget.allCarts[index];
                        final isSelected = _selectedCart?.cartId == cart.cartId;
                        final isCurrent =
                            widget.currentCart?.cartId == cart.cartId;

                        return Card(
                          elevation: isSelected ? 4 : 1,
                          color: isSelected
                              ? global.theme.infoHighlightColor
                              : isCurrent
                              ? global.theme.positiveHighlightColor
                              : global.theme.cardColor,
                          margin: const EdgeInsets.only(bottom: 12),
                          child: InkWell(
                            onTap: () {
                              setState(() {
                                _selectedCart = cart;
                              });
                            },
                            borderRadius: BorderRadius.circular(8),
                            child: Padding(
                              padding: const EdgeInsets.all(16),
                              child: Row(
                                children: [
                                  // Radio button
                                  Radio<String>(
                                    value: cart.cartId,
                                    groupValue: _selectedCart?.cartId,
                                    onChanged: (value) {
                                      setState(() {
                                        _selectedCart = cart;
                                      });
                                    },
                                  ),
                                  const SizedBox(width: 12),

                                  // ข้อมูลตะกร้า
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Row(
                                          children: [
                                            Expanded(
                                              child: Text(
                                                cart.cartName,
                                                style: TextStyle(
                                                  fontSize: 16,
                                                  fontWeight:
                                                      isSelected || isCurrent
                                                      ? FontWeight.bold
                                                      : FontWeight.normal,
                                                ),
                                              ),
                                            ),
                                            if (isCurrent)
                                              Container(
                                                padding:
                                                    const EdgeInsets.symmetric(
                                                      horizontal: 8,
                                                      vertical: 4,
                                                    ),
                                                decoration: BoxDecoration(
                                                  color: global.theme.positiveHighlightTextColor,
                                                  borderRadius:
                                                      BorderRadius.circular(4),
                                                ),
                                                child: Text(
                                                  global.language("in_use"),
                                                  style: TextStyle(
                                                    color: global.theme.onPrimaryColor,
                                                    fontSize: 12,
                                                  ),
                                                ),
                                              ),
                                          ],
                                        ),
                                        SizedBox(height: 4),
                                        Text(
                                          '${global.language("cart_type_label")}: ${cart.cartType.displayNameThai}',
                                          style: TextStyle(
                                            fontSize: 14,
                                            color: global.theme.iconSecondaryColor,
                                          ),
                                        ),
                                        Text(
                                          '${global.language("product")} ${cart.itemCount} ${global.language("items")} • ${global.formatPrice(cart.totalAmount)}',
                                          style: TextStyle(
                                            fontSize: 14,
                                            color: global.theme.iconSecondaryColor,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),

                                  // ปุ่มแก้ไข
                                  IconButton(
                                    onPressed: () => _showCartEditDialog(cart),
                                    icon: Icon(Icons.edit),
                                    tooltip: global.language("edit_cart_name_and_type"),
                                    color: global.theme.infoHighlightTextColor,
                                  ),
                                ],
                              ),
                            ),
                          ),
                        );
                      },
                    ),
            ),

            const Divider(height: 24),

            // ปุ่มด้านล่าง
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: Text(global.language("cancel")),
                ),
                SizedBox(width: 12),
                ElevatedButton(
                  onPressed: _selectedCart == null
                      ? null
                      : () {
                          if (_selectedCart != null) {
                            widget.onCartSelected(_selectedCart!);
                            Navigator.pop(context);
                          }
                        },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: global.theme.infoHighlightTextColor,
                    foregroundColor: global.theme.onPrimaryColor,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 24,
                      vertical: 12,
                    ),
                  ),
                  child: Text(global.language("confirm_change_cart")),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  /// แสดง Dialog สำหรับแก้ไขชื่อและประเภทตะกร้า
  Future<void> _showCartEditDialog(CartModel cart) async {
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) =>
          _CartEditDialog(cart: cart, systemType: widget.systemType),
    );

    if (result != null) {
      final newName = result['name'] as String;
      final newType = result['type'] as CartType;

      AppLogger.info(
        '📝 กำลังอัพเดทตะกร้า: ${cart.cartName} → $newName, ประเภท: ${cart.cartType.displayNameThai} → ${newType.displayNameThai}',
      );

      try {
        // อัพเดทตะกร้าใน MongoDB
        final updatedCart = cart.updateNameAndType(newName, newType);
        final cartService = MongoDBCartService();
        await cartService.updateCart(updatedCart);

        AppLogger.info(
          '✅ อัพเดทตะกร้าสำเร็จ: $newName (${newType.displayNameThai})',
        );

        // แจ้งผู้ใช้ว่าอัพเดทสำเร็จ
        if (mounted) {
          global.showSuccessSnackBar(context, '${global.language("update_cart_success")} "$newName" (${newType.displayNameThai})');

          // อัพเดท state ภายใน dialog เพื่อให้แสดงข้อมูลใหม่
          setState(() {
            // อัพเดทตะกร้าในลิสต์
            final index = widget.allCarts.indexWhere(
              (c) => c.cartId == cart.cartId,
            );
            if (index != -1) {
              widget.allCarts[index] = updatedCart;
            }

            // อัพเดท selected cart ถ้าเป็นตัวเดียวกัน
            if (_selectedCart?.cartId == cart.cartId) {
              _selectedCart = updatedCart;
            }
          });

          // แจ้ง parent ว่าตะกร้าถูกอัพเดท (สำคัญมาก!)
          widget.onCartUpdated(updatedCart);

          AppLogger.info('📢 แจ้ง parent widget ให้รีเฟรชข้อมูล');
        }
      } catch (e) {
        AppLogger.error('❌ Error updating cart: $e', error: e);
        if (mounted) {
          global.showErrorSnackBar(context, '${global.language("cannot_update_cart")}: $e');
        }
      }
    } else {
      AppLogger.info('⚠️ ยกเลิกการแก้ไขตะกร้า');
    }
  }
}

/// Dialog สำหรับแก้ไขชื่อและประเภทตะกร้า
class _CartEditDialog extends StatefulWidget {
  final CartModel cart;
  final CartSystemType systemType;

  const _CartEditDialog({required this.cart, required this.systemType});

  @override
  State<_CartEditDialog> createState() => _CartEditDialogState();
}

class _CartEditDialogState extends State<_CartEditDialog> with global.ThemeRefreshMixin {
  late TextEditingController _nameController;
  late CartType _selectedCartType;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.cart.cartName);
    _selectedCartType = widget.cart.cartType;
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // รายการ CartType ที่ใช้ได้กับระบบนี้
    final availableCartTypes = getCartTypesForSystem(widget.systemType);

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Container(
        constraints: const BoxConstraints(maxWidth: 400),
        padding: EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // หัวข้อ
            Row(
              children: [
                Icon(Icons.edit, color: global.theme.infoHighlightTextColor, size: 28),
                SizedBox(width: 12),
                Expanded(
                  child: Text(
                    global.language("edit_cart_name_and_type"),
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.pop(context),
                  icon: Icon(Icons.close),
                ),
              ],
            ),
            const Divider(height: 24),

            // ชื่อตะกร้า
            Text(
              global.language("cart_name"),
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 8),
            TextField(
              controller: _nameController,
              decoration: InputDecoration(
                border: OutlineInputBorder(),
                hintText: global.language("enter_cart_name"),
              ),
              autofocus: true,
            ),
            const SizedBox(height: 20),

            // ประเภทตะกร้า
            Text(
              global.language("cart_type_label"),
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),

            // Radio buttons สำหรับเลือกประเภท
            ...availableCartTypes.map((cartType) {
              return RadioListTile<CartType>(
                value: cartType,
                groupValue: _selectedCartType,
                onChanged: (value) {
                  if (value != null) {
                    setState(() {
                      _selectedCartType = value;
                    });
                  }
                },
                title: Text(cartType.displayNameThai),
                dense: true,
                contentPadding: EdgeInsets.zero,
              );
            }),

            const SizedBox(height: 24),

            // ปุ่มด้านล่าง
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: Text(global.language("cancel")),
                ),
                SizedBox(width: 12),
                ElevatedButton(
                  onPressed: () {
                    final newName = _nameController.text.trim();
                    if (newName.isEmpty) {
                      global.showWarningSnackBar(context, global.language("please_enter_cart_name"));
                      return;
                    }

                    Navigator.pop(context, {
                      'name': newName,
                      'type': _selectedCartType,
                    });
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: global.theme.infoHighlightTextColor,
                    foregroundColor: global.theme.onPrimaryColor,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 24,
                      vertical: 12,
                    ),
                  ),
                  child: Text(global.language("confirm_changes")),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
