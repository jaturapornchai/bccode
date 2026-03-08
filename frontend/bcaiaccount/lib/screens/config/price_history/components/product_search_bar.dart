import 'package:flutter/material.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;

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
  final Function() onLineSpaceChange;

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
    required this.onLineSpaceChange,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(5),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(2),
      ),
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
              decoration: InputDecoration(
                isDense: true,
                contentPadding:
                    const EdgeInsets.only(top: 0, bottom: 0, left: 0, right: 0),
                border: InputBorder.none,
                hintText: global.language('search'),
              ),
            ),
          ),
          IconButton(
            onPressed: onFilterPressed,
            icon: Icon(
              (filterBarcode.branch == false)
                  ? Icons.filter_alt_off
                  : Icons.filter_alt,
              color:
                  (filterBarcode.branch == false) ? Colors.black : Colors.blue,
            ),
          ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            icon: Icon((showImage) ? Icons.image_not_supported : Icons.image),
            onPressed: onImageToggle,
          ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            icon: const FaIcon(FontAwesomeIcons.font),
            onPressed: onFontSizeChange,
          ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            icon: const Icon(Icons.line_weight),
            onPressed: onLineSpaceChange,
          ),
        ],
      ),
    );
  }
}
