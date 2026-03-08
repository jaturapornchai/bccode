import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../bloc/cart_cubit.dart';
import '../bloc/cart_state.dart';
import '../models/cart_model.dart';
import '../models/cart_system_type.dart';
import 'cart_detail_screen.dart';
import '../../../global.dart' as global;

/// Widget แสดงรายการตะกร้าทั้งหมด
class CartListWidget extends StatelessWidget {
  final String shopId;
  final String email;
  final String? activeCartId; // Cart ID ที่กำลังใช้งาน
  final Function(CartModel)? onCartTap;
  final Function(CartModel)? onCartDelete;
  final VoidCallback? onCartUpdated; // Callback เมื่อมีการแก้ไขตะกร้า

  const CartListWidget({
    super.key,
    required this.shopId,
    required this.email,
    this.activeCartId,
    this.onCartTap,
    this.onCartDelete,
    this.onCartUpdated,
  });

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<CartCubit, CartState>(
      builder: (context, state) {
        if (state is CartLoading) {
          return const Center(child: CircularProgressIndicator());
        }

        if (state is CartError) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 64, color: Colors.red),
                const SizedBox(height: 16),
                Text(
                  'Error: ${state.error}',
                  style: const TextStyle(color: Colors.red),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () {
                    context.read<CartCubit>().loadCarts(shopId, email);
                  },
                  child: const Text('Retry'),
                ),
              ],
            ),
          );
        }

        if (state is CartLoaded) {
          final carts = state.carts;

          if (carts.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(
                    Icons.shopping_cart_outlined,
                    size: 64,
                    color: Colors.grey,
                  ),
                  const SizedBox(height: 16),
                  const Text(
                    'No carts yet',
                    style: TextStyle(fontSize: 18, color: Colors.grey),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Create a new cart to get started',
                    style: TextStyle(color: Colors.grey),
                  ),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () async {
              await context.read<CartCubit>().refreshCarts(shopId, email);
            },
            child: ListView.builder(
              itemCount: carts.length,
              itemBuilder: (context, index) {
                final cart = carts[index];
                final isActive = cart.cartId == activeCartId;
                return _CartCard(
                  cart: cart,
                  isActive: isActive,
                  onTap: () async {
                    // เปิดหน้ารายละเอียดตะกร้า
                    await Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => CartDetailScreen(
                          cart: cart,
                          onCartUpdated: onCartUpdated,
                          onCartDeleted: () {
                            if (onCartDelete != null) {
                              onCartDelete!(cart);
                            }
                          },
                        ),
                      ),
                    );
                    // Refresh หลังกลับมา
                    onCartUpdated?.call();
                  },
                  onDelete: onCartDelete != null
                      ? () => onCartDelete!(cart)
                      : null,
                );
              },
            ),
          );
        }

        return const Center(child: Text('Unknown state'));
      },
    );
  }
}

/// Card แสดงข้อมูลตะกร้า
class _CartCard extends StatelessWidget {
  final CartModel cart;
  final bool isActive; // ตะกร้านี้เป็น active cart หรือไม่
  final VoidCallback? onTap;
  final VoidCallback? onDelete;

  const _CartCard({
    required this.cart,
    this.isActive = false,
    this.onTap,
    this.onDelete,
  });

  Color _getStatusColor() {
    switch (cart.status) {
      case 'active':
        return Colors.green;
      case 'ordered':
        return Colors.blue;
      case 'completed':
        return Colors.grey;
      case 'cancelled':
        return Colors.red;
      default:
        return Colors.grey;
    }
  }

  IconData _getCartTypeIcon() {
    switch (cart.systemType) {
      case CartSystemType.purchase:
        return Icons.add_shopping_cart;
      case CartSystemType.sales:
        return Icons.shopping_cart;
      case CartSystemType.inventory:
        return Icons.inventory_2;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      elevation: isActive ? 4 : 1, // Active cart ให้ elevation สูงขึ้น
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: isActive
            ? const BorderSide(color: Color(0xFFEE4D2D), width: 2)
            : BorderSide.none,
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header
              Row(
                children: [
                  Icon(_getCartTypeIcon(), size: 24),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            // แสดงไอคอน active
                            if (isActive) ...[
                              const Icon(
                                Icons.check_circle,
                                color: Color(0xFFEE4D2D),
                                size: 16,
                              ),
                              const SizedBox(width: 4),
                            ],
                            Flexible(
                              child: Text(
                                cart.cartName,
                                style: const TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 4),
                        Text(
                          cart.systemType.displayNameThai,
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ),
                  ),
                  // Status badge
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: _getStatusColor().withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: _getStatusColor(), width: 1),
                    ),
                    child: Text(
                      cart.status.toUpperCase(),
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        color: _getStatusColor(),
                      ),
                    ),
                  ),
                  // Delete button
                  if (onDelete != null)
                    IconButton(
                      icon: const Icon(Icons.delete_outline, color: Colors.red),
                      onPressed: onDelete,
                    ),
                ],
              ),

              const Divider(height: 24),

              // แสดงรายการสินค้าทั้งหมด
              if (cart.items.isNotEmpty) ...[
                Text(
                  '${global.language("product_list")}:',
                  style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 8),
                ...cart.items.map(
                  (item) => Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: Colors.grey[50],
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.grey[200]!),
                      ),
                      child: Row(
                        children: [
                          // รูปภาพสินค้า (แสดงเฉพาะถ้ามี)
                          if (item.imageuri != null && item.imageuri!.isNotEmpty)
                            Container(
                              width: 50,
                              height: 50,
                              margin: const EdgeInsets.only(right: 8),
                              decoration: BoxDecoration(
                                borderRadius: BorderRadius.circular(6),
                                border: Border.all(color: Colors.grey[300]!),
                              ),
                              child: ClipRRect(
                                borderRadius: BorderRadius.circular(6),
                                child: Image.network(
                                  item.imageuri!,
                                  fit: BoxFit.cover,
                                  loadingBuilder: (context, child,
                                      loadingProgress) {
                                    if (loadingProgress == null) return child;
                                    return Center(
                                      child: CircularProgressIndicator(
                                        value: loadingProgress
                                                    .expectedTotalBytes !=
                                                null
                                            ? loadingProgress
                                                    .cumulativeBytesLoaded /
                                                loadingProgress
                                                    .expectedTotalBytes!
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
                                        size: 20,
                                      ),
                                    );
                                  },
                                ),
                              ),
                            ),

                          // Product info
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  item.productName,
                                  style: const TextStyle(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w600,
                                  ),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  '${item.itemCode} | ${item.barcode}',
                                  style: TextStyle(
                                    fontSize: 11,
                                    color: Colors.grey[600],
                                  ),
                                ),
                                if (item.remark != null &&
                                    item.remark!.isNotEmpty)
                                  Padding(
                                    padding: const EdgeInsets.only(top: 2),
                                    child: Row(
                                      children: [
                                        Icon(
                                          Icons.note,
                                          size: 11,
                                          color: Colors.amber[700],
                                        ),
                                        const SizedBox(width: 4),
                                        Flexible(
                                          child: Text(
                                            item.remark!,
                                            style: TextStyle(
                                              fontSize: 11,
                                              color: Colors.amber[900],
                                            ),
                                            maxLines: 1,
                                            overflow: TextOverflow.ellipsis,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                              ],
                            ),
                          ),
                          const SizedBox(width: 8),
                          // Quantity and Price
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.blue[50],
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                child: Text(
                                  '${global.formatNumberRemoveRightZero(item.quantity)} ${item.unitName}',
                                  style: const TextStyle(
                                    fontSize: 12,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.blue,
                                  ),
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                global.formatPrice(item.totalPrice),
                                style: const TextStyle(
                                  fontSize: 14,
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFFEE4D2D),
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 12),
              ],

              // Cart summary
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFEE4D2D).withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${cart.items.length} รายการ | ${cart.totalQuantity.toStringAsFixed(0)} ชิ้น',
                          style: const TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        Text(
                          'Updated: ${_formatDate(cart.updatedAt)}',
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ),
                    Text(
                      global.formatPrice(cart.totalAmount),
                      style: const TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFFEE4D2D),
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 8),

              // Action button
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  icon: Icon(Icons.edit, size: 16),
                  label: Text(global.language("edit_items")),
                  onPressed: onTap,
                  style: OutlinedButton.styleFrom(
                    foregroundColor: const Color(0xFFEE4D2D),
                    side: const BorderSide(color: Color(0xFFEE4D2D)),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inMinutes < 1) {
      return 'Just now';
    } else if (diff.inHours < 1) {
      return '${diff.inMinutes}m ago';
    } else if (diff.inDays < 1) {
      return '${diff.inHours}h ago';
    } else if (diff.inDays < 7) {
      return '${diff.inDays}d ago';
    } else {
      return '${date.day}/${date.month}/${date.year}';
    }
  }
}
