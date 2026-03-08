import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class ProductListHeader extends StatelessWidget {
  final bool showImage;

  const ProductListHeader({
    super.key,
    required this.showImage,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
      color: global.theme.columnHeaderColor,
      child: Row(
        children: [
          Expanded(
            flex: 6,
            child: Text(
              global.language("barcode"),
              style: const TextStyle(
                  color: Colors.black, fontWeight: FontWeight.bold),
            ),
          ),
          Expanded(
            flex: 10,
            child: Text(
              global.language("product_name"),
              style: const TextStyle(
                  color: Colors.black, fontWeight: FontWeight.bold),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language("unit"),
              style: const TextStyle(
                  color: Colors.black, fontWeight: FontWeight.bold),
            ),
          ),
          Expanded(
            flex: 4,
            child: Text(
              global.language("item_code"),
              style: const TextStyle(
                  color: Colors.black, fontWeight: FontWeight.bold),
            ),
          ),
          Expanded(
            flex: 4,
            child: Text(
              global.language("product_group"),
              style: const TextStyle(
                  color: Colors.black, fontWeight: FontWeight.bold),
            ),
          ),
          if (showImage)
            const Expanded(
              flex: 1,
              child: Icon(Icons.image, color: Colors.black, size: 12),
            ),
        ],
      ),
    );
  }
}
