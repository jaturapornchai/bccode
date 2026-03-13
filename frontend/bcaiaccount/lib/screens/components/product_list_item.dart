import 'package:flutter/material.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/global.dart' as global;

class ProductListItem extends StatelessWidget {
  final ProductBarcodeModel product;
  final int index;
  final bool isSelected;
  final VoidCallback onAddProduct;

  const ProductListItem({
    super.key,
    required this.product,
    required this.index,
    required this.isSelected,
    required this.onAddProduct,
  });

  @override
  Widget build(BuildContext context) {
    final bool isEvenRow = index % 2 == 0;

    return RepaintBoundary(
      child: Container(
        color: isEvenRow ? global.theme.columnAlternateEvenColor : global.theme.columnAlternateOddColor,
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Expanded(
              flex: 4,
              child: Text(
                product.barcode!,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 6,
              child: Text(
                global.activeLangName(product.names!),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 3,
              child: Text(
                global.activeLangName(product.itemunitnames!),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 3,
              child: Text(
                product.itemcode!,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 4,
              child: Text(
                global.activeLangName(product.groupnames!),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            IconButton(
              onPressed: isSelected ? null : onAddProduct,
              icon: Icon(
                Icons.add_circle,
                color: isSelected ? global.theme.iconSecondaryColor : global.theme.positiveHighlightTextColor,
                size: 28,
              ),
              tooltip: isSelected ? global.language('product_list_item_selected') : global.language('product_list_item_select'),
            ),
          ],
        ),
      ),
    );
  }
}
