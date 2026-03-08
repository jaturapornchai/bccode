import 'package:flutter/material.dart';
import '../models/cart_model.dart';
import '../models/cart_item_model.dart';
import '../models/cart_type.dart';
import '../models/cart_transfer_model.dart';
import '../services/mongodb_cart_service.dart';
import '../../../utils/logger/app_logger.dart';
import '../../../global.dart' as global;
import '../../../tabbed_menu_screen.dart';

/// หน้าจอแสดงรายละเอียดสินค้าในตะกร้า
class CartDetailScreen extends StatefulWidget {
  final CartModel cart;
  final VoidCallback? onCartUpdated;
  final VoidCallback? onCartDeleted;

  const CartDetailScreen({
    super.key,
    required this.cart,
    this.onCartUpdated,
    this.onCartDeleted,
  });

  @override
  State<CartDetailScreen> createState() => _CartDetailScreenState();
}

class _CartDetailScreenState extends State<CartDetailScreen> {
  final _cartService = MongoDBCartService();
  late CartModel _currentCart;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _currentCart = widget.cart;
  }

  /// ลบรายการสินค้า
  Future<void> _removeItem(String itemId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language("delete_item")),
        content: Text(global.language("confirm_delete_item_from_cart")),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language("cancel")),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: Text(global.language("delete")),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      setState(() => _isLoading = true);
      try {
        final updatedCart = _currentCart.removeItem(itemId);
        await _cartService.updateCart(updatedCart);
        setState(() {
          _currentCart = updatedCart;
        });
        widget.onCartUpdated?.call();
        if (mounted) {
          global.showSuccessSnackBar(context, global.language("delete_item_success"));
        }
      } catch (e) {
        AppLogger.error('Failed to remove item: $e');
        if (mounted) {
          global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
        }
      } finally {
        setState(() => _isLoading = false);
      }
    }
  }

  /// แก้ไขจำนวนสินค้า
  Future<void> _updateQuantity(CartItemModel item) async {
    final controller = TextEditingController(
      text: global.formatNumberRemoveRightZero(item.quantity),
    );

    final newQuantity = await showDialog<double>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('${global.language("edit_quantity")}: ${item.productName}'),
        content: TextField(
          controller: controller,
          keyboardType: TextInputType.numberWithOptions(decimal: true),
          decoration: InputDecoration(
            labelText: '${global.language("qty")} (${item.unitName})',
            border: const OutlineInputBorder(),
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language("cancel")),
          ),
          ElevatedButton(
            onPressed: () {
              final qty = double.tryParse(controller.text);
              if (qty != null && qty > 0) {
                Navigator.pop(context, qty);
              }
            },
            child: Text(global.language("save")),
          ),
        ],
      ),
    );

    if (newQuantity != null && newQuantity != item.quantity) {
      setState(() => _isLoading = true);
      try {
        final updatedCart = _currentCart.updateItemQuantity(
          item.itemId,
          newQuantity,
        );
        await _cartService.updateCart(updatedCart);
        setState(() {
          _currentCart = updatedCart;
        });
        widget.onCartUpdated?.call();
        if (mounted) {
          global.showSuccessSnackBar(context, global.language("edit_quantity_success"));
        }
      } catch (e) {
        AppLogger.error('Failed to update quantity: $e');
        if (mounted) {
          global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
        }
      } finally {
        setState(() => _isLoading = false);
      }
    }
  }

  /// ส่งข้อมูลไปยังระบบปลายทาง
  Future<void> _sendToSystem() async {
    try {
      // สร้าง CartTransferModel จาก cart ปัจจุบัน
      final transferModel = CartTransferModel.fromCartModel(_currentCart);
      final targetName = transferModel.getTargetSystemName();

      AppLogger.info('🚀 ส่งข้อมูลไป $targetName');

      // ปิดหน้า detail screen ก่อน
      Navigator.pop(context);

      // เข้าถึง TabbedMenuScreen state ผ่าน GlobalKey
      final tabbedMenuState = TabbedMenuScreen.globalKey.currentState;

      if (tabbedMenuState != null) {
        // เปิด tab ใหม่พร้อม TransactionEditScreen
        tabbedMenuState.addTransactionTab(
          title: targetName,
          screen: transferModel.createTransactionScreen(),
        );

        // แสดงข้อความแจ้งเตือน (ใช้ context ก่อน pop ถ้ายัง mounted)
        if (mounted) {
          global.showSuccessSnackBar(context, '${global.language("opened_system")} $targetName ${global.language("in_new_tab")}');
        }
      } else {
        throw Exception('TabbedMenuScreen not found');
      }
    } catch (e) {
      AppLogger.error('❌ Error sending to system: $e');
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    }
  }

  /// ลบตะกร้า
  Future<void> _deleteCart() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language("delete_cart")),
        content: Text(
          '${global.language("confirm_delete_cart")} "${_currentCart.cartName}"?\n${global.language("all_items_will_be_deleted")}',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language("cancel")),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: Text(global.language("delete")),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      setState(() => _isLoading = true);
      try {
        await _cartService.deleteCart(
          _currentCart.shopId,
          _currentCart.email,
          _currentCart.cartId,
        );
        widget.onCartDeleted?.call();
        if (mounted) {
          Navigator.pop(context);
          global.showSuccessSnackBar(context, global.language("delete_cart_success"));
        }
      } catch (e) {
        AppLogger.error('Failed to delete cart: $e');
        if (mounted) {
          global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
        }
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        elevation: 0,
        backgroundColor: const Color(0xFFEE4D2D),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(_currentCart.cartName, style: const TextStyle(fontSize: 16)),
            Text(
              '${_currentCart.items.length} รายการ | ${global.formatPrice(_currentCart.totalAmount)}',
              style: const TextStyle(fontSize: 12, color: Colors.white70),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.delete_forever),
            onPressed: _deleteCart,
            tooltip: global.language("delete_cart"),
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _currentCart.items.isEmpty
          ? Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.shopping_cart_outlined,
                    size: 80,
                    color: Colors.grey[400],
                  ),
                  SizedBox(height: 16),
                  Text(
                    global.language("no_items_in_cart"),
                    style: TextStyle(fontSize: 18, color: Colors.grey[600]),
                  ),
                ],
              ),
            )
          : ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: _currentCart.items.length,
              itemBuilder: (context, index) {
                final item = _currentCart.items[index];
                return _buildCartItemCard(item);
              },
            ),
      bottomNavigationBar: _currentCart.items.isNotEmpty
          ? Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.05),
                    blurRadius: 10,
                    offset: const Offset(0, -5),
                  ),
                ],
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        global.language("grand_total"),
                        style: const TextStyle(fontSize: 14, color: Colors.grey),
                      ),
                      Text(
                        global.formatPrice(_currentCart.recalculatedTotal),
                        style: const TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFFEE4D2D),
                        ),
                      ),
                    ],
                  ),
                  ElevatedButton.icon(
                    icon: Icon(Icons.send, size: 20),
                    label: Text(
                      '${global.language("send_to_system")} ${_currentCart.cartType.displayNameThai}',
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    onPressed: _sendToSystem,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFFEE4D2D),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 32,
                        vertical: 16,
                      ),
                    ),
                  ),
                ],
              ),
            )
          : null,
    );
  }

  /// สร้าง Widget แสดงรูปภาพสินค้า
  Widget _buildProductImage(CartItemModel item) {
    // ถ้าไม่มีรูป ไม่แสดงอะไร (ประหยัดพื้นที่)
    if (item.imageuri == null || item.imageuri!.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      width: 80,
      height: 80,
      margin: const EdgeInsets.only(right: 12, bottom: 8),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey[300]!),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(8),
        child: Image.network(
          item.imageuri!,
          fit: BoxFit.cover,
          loadingBuilder: (context, child, loadingProgress) {
            if (loadingProgress == null) return child;
            return Center(
              child: CircularProgressIndicator(
                value: loadingProgress.expectedTotalBytes != null
                    ? loadingProgress.cumulativeBytesLoaded /
                          loadingProgress.expectedTotalBytes!
                    : null,
                strokeWidth: 2,
              ),
            );
          },
          errorBuilder: (context, error, stackTrace) {
            return Container(
              color: Colors.grey[200],
              child: Icon(
                Icons.image_not_supported,
                color: Colors.grey[400],
                size: 32,
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildCartItemCard(CartItemModel item) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // รูปภาพสินค้า (แสดงเฉพาะถ้ามี)
                _buildProductImage(item),

                // Product Info
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        item.productName,
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      SizedBox(height: 4),
                      Text(
                        '${global.language("code")}: ${item.itemCode}',
                        style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                      ),
                      Text(
                        '${global.language("barcode")}: ${item.barcode}',
                        style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                      ),
                    ],
                  ),
                ),
                // Actions
                Row(
                  children: [
                    IconButton(
                      icon: const Icon(Icons.edit, size: 20),
                      onPressed: () => _updateQuantity(item),
                      tooltip: global.language("edit_quantity"),
                      color: Colors.blue,
                    ),
                    IconButton(
                      icon: const Icon(Icons.delete, size: 20),
                      onPressed: () => _removeItem(item.itemId),
                      tooltip: global.language("delete_item"),
                      color: Colors.red,
                    ),
                  ],
                ),
              ],
            ),
            const Divider(height: 16),
            // Quantity and Price
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                // Quantity
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.blue[50],
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    '${global.formatNumberRemoveRightZero(item.quantity)} ${item.unitName}',
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      color: Colors.blue,
                    ),
                  ),
                ),
                // Price
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(
                      '${global.formatPrice(item.price)} / ${item.unitName}',
                      style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                    ),
                    if (item.discount > 0)
                      Text(
                        '${global.language("discount")}: ${global.formatPrice(item.discount)}',
                        style: const TextStyle(fontSize: 12, color: Colors.red),
                      ),
                    const SizedBox(height: 4),
                    Text(
                      global.formatPrice(item.totalPrice),
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFFEE4D2D),
                      ),
                    ),
                  ],
                ),
              ],
            ),
            // Remark
            if (item.remark != null && item.remark!.isNotEmpty) ...[
              const SizedBox(height: 8),
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.amber[50],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.amber[200]!),
                ),
                child: Row(
                  children: [
                    Icon(Icons.note, size: 16, color: Colors.amber[700]),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        item.remark!,
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.amber[900],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
