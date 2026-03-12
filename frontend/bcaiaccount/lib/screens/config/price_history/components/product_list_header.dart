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
      padding: EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
      color: global.theme.columnHeaderColor,
      child: Row(
        children: [
          Expanded(
            flex: 6,
            child: Text(
              global.language("barcode"),
              style: TextStyle(
                  color: global.theme.textColor, fontWeight: FontWeight.bold),
            ),
          ),
          Expanded(
            flex: 10,
            child: Text(
              global.language("product_name"),
              style: TextStyle(
                  color: global.theme.textColor, fontWeight: FontWeight.bold),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          Expanded(
            flex: 2,
            child: Text(
              global.language("unit"),
              style: TextStyle(
                  color: global.theme.textColor, fontWeight: FontWeight.bold),
            ),
          ),
          Expanded(
            flex: 4,
            child: Text(
              global.language("item_code"),
              style: TextStyle(
                  color: global.theme.textColor, fontWeight: FontWeight.bold),
            ),
          ),
          Expanded(
            flex: 4,
            child: Text(
              global.language("product_group"),
              style: TextStyle(
                  color: global.theme.textColor, fontWeight: FontWeight.bold),
            ),
          ),
          if (showImage)
            Expanded(
              flex: 1,
              child: Icon(Icons.image, color: global.theme.textColor, size: 12),
            ),
        ],
      ),
    );
  }
}
