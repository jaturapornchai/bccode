import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class AddProductButtom extends StatelessWidget {
  final VoidCallback onSearchPressed;
  final VoidCallback onBarcodePressed;
  final VoidCallback onAddPressed;

  const AddProductButtom({
    super.key,
    required this.onSearchPressed,
    required this.onBarcodePressed,
    required this.onAddPressed,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.all(8.0),
      child: Wrap(
        spacing: 10,
        runSpacing: 10,
        children: [
          ElevatedButton(
            onPressed: onSearchPressed,
            child: Text(global.language('add_new_line_by_search')),
          ),
          ElevatedButton(
            onPressed: onBarcodePressed,
            child: Text(global.language('add_new_line_by_barcode')),
          ),
          ElevatedButton(
            onPressed: onAddPressed,
            child: Text(global.language('add_new_line')),
          ),
        ],
      ),
    );
  }
}
