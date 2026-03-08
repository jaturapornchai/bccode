import 'package:flutter/material.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/utils/image_tooltip.dart';
import 'package:smlaicloud/global.dart' as global;

class ProductListItem extends StatelessWidget {
  final int index;
  final ProductBarcodeModel product;
  final String selectedGuid;
  final bool showImage;
  final bool showBranches;
  final GlobalKey<ImageTooltipState> imageTooltipKey;
  final VoidCallback onTap;

  const ProductListItem({
    super.key,
    required this.index,
    required this.product,
    required this.selectedGuid,
    required this.showImage,
    required this.showBranches,
    required this.imageTooltipKey,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final textStyle = (selectedGuid == product.guidfixed)
        ? TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.bold)
        : TextStyle(fontSize: global.deviceConfig.listDataFontSize);

    return GestureDetector(
      onTap: onTap,
      child: Container(
        color: _getContainerColor(selectedGuid, product.guidfixed, index),
        padding: EdgeInsets.only(
          left: 10,
          right: 10,
          top: global.deviceConfig.listDataLineSpace,
          bottom: global.deviceConfig.listDataLineSpace,
        ),
        child: Column(
          children: [
            _buildRow(product, textStyle, showImage, imageTooltipKey),
            if (showBranches &&
                product.branches != null &&
                product.branches!.isNotEmpty)
              SizedBox(
                width: double.infinity,
                child: _buildChipWrap(product.branches!),
              ),
          ],
        ),
      ),
    );
  }

  Color? _getContainerColor(String selectedGuid, String guidfixed, int index) {
    return (selectedGuid == guidfixed)
        ? Colors.cyan[100]
        : (index % 2 == 0)
            ? global.theme.columnAlternateEvenColor
            : global.theme.columnAlternateOddColor;
  }

  Row _buildRow(ProductBarcodeModel value, TextStyle textStyle, bool showImage,
      GlobalKey<ImageTooltipState> key) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
            flex: 6,
            child: Text(value.barcode!,
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis)),
        Expanded(
          flex: 10,
          child: Text(
            global.activeLangName(value.names!),
            style: textStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        Expanded(
            flex: 2,
            child: Text(global.activeLangName(value.itemunitnames!),
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis)),
        Expanded(
            flex: 4,
            child: Text(value.itemcode!,
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis)),
        Expanded(
            flex: 4,
            child: Text(global.activeLangName(value.groupnames!),
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis)),
        if (showImage) _buildImageWidget(value, key),
      ],
    );
  }

  Widget _buildImageWidget(
      ProductBarcodeModel value, GlobalKey<ImageTooltipState> key) {
    return Expanded(
      flex: 1,
      child: value.imageuri!.isNotEmpty
          ? ImageTooltip(
              key: key,
              image: Image.network(value.imageuri!),
              child: _buildImageContainer(value.imageuri!),
            )
          : Container(),
    );
  }

  Container _buildImageContainer(String imageUri) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(2),
        border: Border.all(color: Colors.grey, width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withValues(alpha: 0.5),
            spreadRadius: 1,
            blurRadius: 1,
            offset: const Offset(1, 1),
          ),
        ],
      ),
      padding: const EdgeInsets.all(1),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(2),
        child: Image.network(
          imageUri,
          fit: BoxFit.fitHeight,
          height: 20,
          errorBuilder:
              (BuildContext context, Object exception, StackTrace? stackTrace) {
            return const Icon(Icons.image_not_supported);
          },
        ),
      ),
    );
  }

  Wrap _buildChipWrap(List<CompanyBranchModel> branches) {
    return Wrap(
      spacing: 2.0,
      runSpacing: 2.0,
      children: branches.map((branch) => _buildChip(branch)).toList(),
    );
  }

  Chip _buildChip(CompanyBranchModel branch) {
    return Chip(
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
      padding: const EdgeInsets.all(0),
      label: Text(
        '${branch.code} - ${global.activeLangName(branch.names)}',
        style: const TextStyle(fontSize: 12),
      ),
    );
  }
}
