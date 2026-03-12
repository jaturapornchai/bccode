import 'package:flutter/material.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/list_font_size_control.dart';

class ProductSearchBar extends StatelessWidget {
  final TextEditingController searchController;
  final FocusNode searchFocusNode;
  final FocusNode listFocusNode;
  final FiltterBarcodeModel filterBarcode;
  final bool showImage;
  final Function(String) onSearchChanged;
  final Function() onFilterPressed;
  final Function() onImageToggle;
  final Function() onFontSizeChange;

  const ProductSearchBar({
    super.key,
    required this.searchController,
    required this.searchFocusNode,
    required this.listFocusNode,
    required this.filterBarcode,
    required this.showImage,
    required this.onSearchChanged,
    required this.onFilterPressed,
    required this.onImageToggle,
    required this.onFontSizeChange,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      color: global.theme.searchBarColor,
      child: Row(
        children: [
          Expanded(
            child: TextField(
              onSubmitted: (value) {
                listFocusNode.requestFocus();
              },
              onChanged: onSearchChanged,
              autofocus: false,
              focusNode: searchFocusNode,
              controller: searchController,
              style: TextStyle(color: global.theme.formTextColor),
              decoration: InputDecoration(
                isDense: true,
                contentPadding:
                    EdgeInsets.symmetric(vertical: 8, horizontal: 4),
                border: InputBorder.none,
                filled: false,
                hintText: global.language('search'),
                hintStyle: TextStyle(color: global.theme.formHintColor),
                prefixIcon: Icon(Icons.search,
                    size: 20, color: global.theme.formHintColor),
                prefixIconConstraints:
                    const BoxConstraints(minWidth: 32, minHeight: 32),
              ),
            ),
          ),
          IconButton(
            onPressed: onFilterPressed,
            icon: Icon(
              (filterBarcode.branch == false)
                  ? Icons.filter_alt_off
                  : Icons.filter_alt,
              color: (filterBarcode.branch == false)
                  ? global.theme.iconSecondaryColor
                  : global.theme.primaryColor,
            ),
          ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            icon: Icon(
              (showImage) ? Icons.image_not_supported : Icons.image,
              color: global.theme.iconSecondaryColor,
            ),
            onPressed: onImageToggle,
          ),
          ListFontSizeControl(onChanged: onFontSizeChange),
        ],
      ),
    );
  }
}
