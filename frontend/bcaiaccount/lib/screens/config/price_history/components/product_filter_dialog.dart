import 'package:flutter/material.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;

class ProductFilterDialog extends StatefulWidget {
  final FiltterBarcodeModel initialFilter;

  const ProductFilterDialog({
    super.key,
    required this.initialFilter,
  });

  @override
  State<ProductFilterDialog> createState() => _ProductFilterDialogState();
}

class _ProductFilterDialogState extends State<ProductFilterDialog>
    with global.ThemeRefreshMixin {
  late FiltterBarcodeModel filterBarcode;

  @override
  void initState() {
    super.initState();
    filterBarcode = FiltterBarcodeModel(branch: widget.initialFilter.branch);
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(global.language("filter_product")),
      content: SizedBox(
        width: 600.0,
        height: 600.0,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Divider(),
            RadioListTile(
              title: Text(global.language("branch_selected")),
              value: true,
              groupValue: filterBarcode.branch,
              onChanged: (value) {
                setState(() {
                  filterBarcode.branch = value!;
                });
              },
            ),
            RadioListTile(
              title: Text(global.language("all")),
              value: false,
              groupValue: filterBarcode.branch,
              onChanged: (value) {
                setState(() {
                  filterBarcode.branch = value!;
                });
              },
            ),
          ],
        ),
      ),
      actions: [
        ElevatedButton(
          onPressed: () {
            Navigator.pop(context, filterBarcode);
          },
          child: Text(global.language("confirm")),
        ),
        ElevatedButton(
          style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
          onPressed: () {
            Navigator.pop(context, null);
          },
          child: Text(global.language("clear")),
        ),
      ],
    );
  }
}
