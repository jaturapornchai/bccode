import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;
import '../models/cart_item_model.dart';

/// Widget แสดงรายการสินค้าในตะกร้า
class CartItemWidget extends StatelessWidget {
  final CartItemModel item;
  final Function(double)? onQuantityChanged;
  final VoidCallback? onDelete;
  final bool isEditable;

  const CartItemWidget({
    super.key,
    required this.item,
    this.onQuantityChanged,
    this.onDelete,
    this.isEditable = true,
  });

  /// สร้าง Widget แสดงรูปภาพสินค้า
  Widget _buildProductImage() {
    // ถ้าไม่มีรูป ไม่แสดงอะไร (ประหยัดพื้นที่)
    if (item.imageuri == null || item.imageuri!.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      width: 60,
      height: 60,
      margin: const EdgeInsets.only(right: 12),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.dividerBorderColor),
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
              color: global.theme.dividerBorderColor,
              child: Icon(
                Icons.image_not_supported,
                color: global.theme.iconSecondaryColor,
                size: 24,
              ),
            );
          },
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Product name and delete button
            Row(
              children: [
                // รูปภาพสินค้า (แสดงเฉพาะถ้ามี)
                _buildProductImage(),

                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        item.productName,
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'Code: ${item.itemCode}',
                        style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
                      ),
                      if (item.barcode.isNotEmpty)
                        Text(
                          'Barcode: ${item.barcode}',
                          style: TextStyle(
                            fontSize: 12,
                            color: global.theme.iconSecondaryColor,
                          ),
                        ),
                    ],
                  ),
                ),
                if (isEditable && onDelete != null)
                  IconButton(
                    icon: Icon(Icons.delete_outline, color: global.theme.negativeHighlightTextColor),
                    onPressed: onDelete,
                  ),
              ],
            ),

            const Divider(height: 16),

            // Price and quantity controls
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                // Price per unit
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Price/Unit',
                      style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      global.formatPrice(item.price),
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: global.theme.infoHighlightTextColor,
                      ),
                    ),
                  ],
                ),

                // Quantity controls
                if (isEditable && onQuantityChanged != null)
                  _QuantityControls(
                    quantity: item.quantity,
                    unitName: item.unitName,
                    onChanged: onQuantityChanged!,
                  )
                else
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        'Quantity',
                        style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '${item.quantity.toStringAsFixed(2)} ${item.unitName}',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
              ],
            ),

            const SizedBox(height: 12),

            // Total price
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: global.theme.positiveHighlightColor,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: global.theme.positiveHighlightTextColor, width: 1),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Total',
                    style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                  ),
                  Text(
                    global.formatPrice(item.totalPrice),
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: global.theme.positiveHighlightTextColor,
                    ),
                  ),
                ],
              ),
            ),

            // Unit conversion info
            if (item.unitStand > 1)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(
                  'Unit conversion: ${item.unitStand}:${item.unitDive}',
                  style: TextStyle(
                    fontSize: 11,
                    color: global.theme.iconSecondaryColor,
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

/// Quantity controls widget
class _QuantityControls extends StatefulWidget {
  final double quantity;
  final String unitName;
  final Function(double) onChanged;

  const _QuantityControls({
    required this.quantity,
    required this.unitName,
    required this.onChanged,
  });

  @override
  State<_QuantityControls> createState() => _QuantityControlsState();
}

class _QuantityControlsState extends State<_QuantityControls> with global.ThemeRefreshMixin {
  late TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(
      text: widget.quantity.toStringAsFixed(2),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        // Decrease button
        IconButton(
          icon: Icon(Icons.remove_circle_outline),
          onPressed: () {
            final newQty = widget.quantity - 1;
            if (newQty > 0) {
              _controller.text = newQty.toStringAsFixed(2);
              widget.onChanged(newQty);
            }
          },
        ),

        // Quantity input
        SizedBox(
          width: 100,
          child: TextField(
            controller: _controller,
            textAlign: TextAlign.center,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            inputFormatters: [
              FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
            ],
            decoration: InputDecoration(
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 8,
                vertical: 8,
              ),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
              ),
              suffix: Text(
                widget.unitName,
                style: TextStyle(fontSize: 12),
              ),
            ),
            onSubmitted: (value) {
              final newQty = double.tryParse(value) ?? widget.quantity;
              if (newQty > 0) {
                _controller.text = newQty.toStringAsFixed(2);
                widget.onChanged(newQty);
              } else {
                _controller.text = widget.quantity.toStringAsFixed(2);
              }
            },
          ),
        ),

        // Increase button
        IconButton(
          icon: Icon(Icons.add_circle_outline),
          onPressed: () {
            final newQty = widget.quantity + 1;
            _controller.text = newQty.toStringAsFixed(2);
            widget.onChanged(newQty);
          },
        ),
      ],
    );
  }
}
