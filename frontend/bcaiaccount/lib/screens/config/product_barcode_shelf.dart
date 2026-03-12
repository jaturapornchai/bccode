import 'package:flutter/foundation.dart';
import 'dart:math';

import 'package:smlaicloud/bloc/selected_products/selected_products_bloc.dart';
import 'package:smlaicloud/bloc/product_list/product_list_bloc.dart';
import 'package:smlaicloud/bloc/shelf_product_selector/shelf_product_selector_bloc.dart';
import 'package:smlaicloud/bloc/product/product_bloc.dart';
import 'package:smlaicloud/repositories/product_repository.dart';
import 'package:smlaicloud/repositories/product_section_repository.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/repositories/warehouse_repository.dart';
import 'package:smlaicloud/screens/components/shelf_product_selector.dart';
import 'package:smlaicloud/screens/components/print_label_dialog.dart';
import 'package:smlaicloud/screens/components/selected_product_grid_item.dart';
import 'package:smlaicloud/screens/components/product_list_item.dart';
import 'package:smlaicloud/components/product_label_print_shelf.dart';
import 'package:smlaicloud/components/product_label_print_a4_shelf.dart';
import 'package:smlaicloud/components/product_label_print_a4_shelf_medium.dart';
import 'package:smlaicloud/components/product_label_print_a4_shelf_large.dart';
import 'package:smlaicloud/components/product_label_print_a4_shelf_xlarge.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:pdf/pdf.dart';

import 'package:split_view/split_view.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

// Import ProductWithCopies from component file

class ProductBarcodeShelf extends StatefulWidget {
  const ProductBarcodeShelf({super.key});

  @override
  State<ProductBarcodeShelf> createState() => ProductBarcodeShelfState();
}

class ProductBarcodeShelfState extends State<ProductBarcodeShelf>
    with global.ThemeRefreshMixin {
  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocusNode = FocusNode(skipTraversal: true);
  final ScrollController _listScrollController = ScrollController();
  late SplitViewController
  _splitViewController; // Changed to late initialization
  final ScrollController _selectedProductsScrollController = ScrollController();
  final _debouncer = global.Debouncer(1000);
  String _searchText = "";
  String? _currentA4FormType; // เก็บข้อมูลประเภท A4 ฟอร์มที่เลือก

  // Store BLoC references for safe disposal
  SelectedProductsBloc? _selectedProductsBloc;
  ProductListBloc? _productListBloc;

  @override
  void initState() {
    super.initState();
    _setupControllers();
    _loadInitialData();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Store BLoC references for safe disposal
    _selectedProductsBloc = context.read<SelectedProductsBloc>();
    _productListBloc = context.read<ProductListBloc>();
  }

  void _setupControllers() {
    _listScrollController.addListener(_onScrollList);
    _splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
  }

  void _loadInitialData() {
    context.read<ProductListBloc>().add(
      LoadProducts(
        search: "",
        branchcode: global.companyBranchSelectData.code,
        businesstypecode: global.companyBranchSelectData.businesstype!.code!,
      ),
    );
  }

  void _onScrollList() {
    final state = context.read<ProductListBloc>().state;
    if (state is ProductListLoaded &&
        _listScrollController.position.pixels >=
            _listScrollController.position.maxScrollExtent - 200 &&
        !state.isLoadingMore &&
        state.hasMore) {
      context.read<ProductListBloc>().add(
        LoadMoreProducts(
          search: _searchText,
          branchcode: global.companyBranchSelectData.code,
          businesstypecode: global.companyBranchSelectData.businesstype!.code!,
        ),
      );
    }
  }

  @override
  void dispose() {
    // Dispose controllers
    _searchController.dispose();
    _searchFocusNode.dispose();
    _listScrollController.dispose();
    _selectedProductsScrollController.dispose();

    // Clean up BLoC states using stored references
    _selectedProductsBloc?.add(const ClearAllProducts());
    _productListBloc?.add(const ClearProductList());

    super.dispose();
  }

  // Helper methods for BLoC interactions
  void _addProductToSelection(ProductBarcodeModel product) {
    context.read<SelectedProductsBloc>().add(AddProductToSelection(product));
  }

  void _addAllProducts(ProductListLoaded state) {
    context.read<SelectedProductsBloc>().add(AddAllProducts(state.products));
    global.showSuccessSnackBar(context, global.language('products_added_all_success'));
  }

  void _removeProduct(String barcode) {
    context.read<SelectedProductsBloc>().add(
      RemoveProductFromSelection(barcode),
    );
  }

  void _updateProductCopies(String barcode, int copies) {
    context.read<SelectedProductsBloc>().add(
      UpdateProductCopies(barcode, copies),
    );
  }

  // เพิ่ม methods ที่ขาดหายไป
  void _openEditCopiesDialog(ProductWithCopies product) {
    openEditCopiesDialog(product);
  }

  bool isProductSelected(String barcode) {
    final currentState = context.read<SelectedProductsBloc>().state;
    if (currentState is SelectedProductsLoaded) {
      return currentState.selectedProducts.any(
        (item) => item.product.barcode == barcode,
      );
    }
    return false;
  }

  // สร้างหน้าจอแสดงรายการสินค้า
  Widget buildProductListScreen() {
    return Scaffold(
      body: Column(
        children: [
          // ช่องค้นหาและตัวกรอง
          Container(
            padding: EdgeInsets.all(5),
            color: global.theme.searchBarColor,
            child: TextField(
              onSubmitted: (value) {
                _searchFocusNode.requestFocus();
              },
              onChanged: (value) {
                _debouncer.run(() {
                  _searchText = value;
                  context.read<ProductListBloc>().add(
                    SearchProducts(
                      search: value,
                      branchcode: global.companyBranchSelectData.code,
                      businesstypecode:
                          global.companyBranchSelectData.businesstype!.code!,
                    ),
                  );
                });
              },
              autofocus: false,
              focusNode: _searchFocusNode,
              controller: _searchController,
              decoration: InputDecoration(
                isDense: true,
                contentPadding: EdgeInsets.symmetric(
                  vertical: 12,
                  horizontal: 10,
                ),
                border: InputBorder.none,
                hintText: global.language('search'),
              ),
            ),
          ),
          Container(color: global.theme.appBarColor, height: 6),

          // ส่วนหัวของตาราง
          BlocBuilder<ProductListBloc, ProductListState>(
            builder: (context, state) {
              final hasProducts =
                  state is ProductListLoaded && state.products.isNotEmpty;

              return Container(
                padding: EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 12,
                ),
                color: global.theme.columnHeaderColor,
                child: Row(
                  children: [
                    Expanded(
                      flex: 4,
                      child: Text(
                        global.language("barcode"),
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                    ),
                    Expanded(
                      flex: 6,
                      child: Text(
                        global.language("product_name"),
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                    ),
                    Expanded(
                      flex: 3,
                      child: Text(
                        global.language("unit"),
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                    ),
                    Expanded(
                      flex: 3,
                      child: Text(
                        global.language("item_code"),
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                    ),
                    Expanded(
                      flex: 4,
                      child: Text(
                        global.language("product_group"),
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                    ),

                    // ปุ่มเพิ่มทั้งหมด
                    if (hasProducts)
                      IconButton(
                        icon: Icon(
                          Icons.playlist_add,
                          color: global.theme.positiveHighlightTextColor,
                        ),
                        onPressed: () => _addAllProducts(state),
                        tooltip: global.language('add_all_products'),
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                    const SizedBox(width: 8),
                  ],
                ),
              );
            },
          ),

          // รายการสินค้า
          Expanded(
            child: BlocBuilder<ProductListBloc, ProductListState>(
              builder: (context, productState) {
                return BlocBuilder<SelectedProductsBloc, SelectedProductsState>(
                  builder: (context, selectedState) {
                    if (productState is ProductListLoading) {
                      return const Center(child: CircularProgressIndicator());
                    }

                    if (productState is ProductListError) {
                      return Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.error,
                              size: 64,
                              color: global.theme.negativeHighlightTextColor,
                            ),
                            SizedBox(height: 16),
                            Text(
                              '${global.language('error_occurred')}: ${productState.message}',
                            ),
                            SizedBox(height: 16),
                            ElevatedButton(
                              onPressed: () =>
                                  context.read<ProductListBloc>().add(
                                    RefreshProducts(
                                      search: _searchText,
                                      branchcode:
                                          global.companyBranchSelectData.code,
                                      businesstypecode: global
                                          .companyBranchSelectData
                                          .businesstype!
                                          .code!,
                                    ),
                                  ),
                              child: Text(global.language('try_again')),
                            ),
                          ],
                        ),
                      );
                    }

                    if (productState is ProductListLoaded) {
                      if (productState.products.isEmpty) {
                        return Center(
                          child: Text(global.language('no_products_found')),
                        );
                      }

                      return ListView.builder(
                        controller: _listScrollController,
                        itemCount:
                            productState.products.length +
                            (productState.isLoadingMore ? 1 : 0),
                        itemBuilder: (context, index) {
                          if (index >= productState.products.length) {
                            // Loading indicator for pagination
                            return const Center(
                              child: Padding(
                                padding: EdgeInsets.all(16.0),
                                child: CircularProgressIndicator(),
                              ),
                            );
                          }

                          final product = productState.products[index];
                          final isSelected =
                              selectedState is SelectedProductsLoaded &&
                              selectedState.isProductSelected(product.barcode!);

                          return ProductListItem(
                            product: product,
                            index: index,
                            isSelected: isSelected,
                            onAddProduct: () => _addProductToSelection(product),
                          );
                        },
                      );
                    }

                    return Center(
                      child: Text(global.language('start_search_products')),
                    );
                  },
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  // สร้างส่วนแสดงรายการสินค้าที่เลือก (GridView)
  Widget buildSelectedProductsList() {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(8),
      ),
      child: BlocBuilder<SelectedProductsBloc, SelectedProductsState>(
        builder: (context, state) {
          return Column(
            children: [
              // ส่วนหัวของรายการ - แสดงเสมอ
              buildListHeader(state),

              // รายการสินค้าที่เลือก
              Expanded(
                child:
                    state is SelectedProductsLoaded &&
                        state.selectedProducts.isNotEmpty
                    ? LayoutBuilder(
                        builder: (context, constraints) {
                          final double cardWidth = 150.0;
                          int crossAxisCount = max(
                            1,
                            (constraints.maxWidth / cardWidth).floor(),
                          );

                          return AnimatedContainer(
                            duration: const Duration(milliseconds: 300),
                            padding: const EdgeInsets.all(8),
                            child: GridView.builder(
                              controller: _selectedProductsScrollController,
                              gridDelegate:
                                  SliverGridDelegateWithFixedCrossAxisCount(
                                    crossAxisCount: crossAxisCount,
                                    childAspectRatio: 0.8,
                                    crossAxisSpacing: 8,
                                    mainAxisSpacing: 8,
                                  ),
                              itemCount: state.selectedProducts.length,
                              itemBuilder: (context, index) =>
                                  AnimatedContainer(
                                    duration: const Duration(milliseconds: 300),
                                    curve: Curves.easeInOut,
                                    child: SelectedProductGridItem(
                                      product: state.selectedProducts[index],
                                      onCopiesChanged: () =>
                                          _updateProductCopies(
                                            state
                                                .selectedProducts[index]
                                                .product
                                                .barcode,
                                            state
                                                .selectedProducts[index]
                                                .copies,
                                          ),
                                      onRemove: () => _removeProduct(
                                        state
                                            .selectedProducts[index]
                                            .product
                                            .barcode,
                                      ),
                                      onEditCopies: () => _openEditCopiesDialog(
                                        state.selectedProducts[index],
                                      ),
                                    ),
                                  ),
                            ),
                          );
                        },
                      )
                    : buildEmptySelectionView(),
              ),

              // Bottom action bar
              if (state is SelectedProductsLoaded &&
                  state.selectedProducts.isNotEmpty)
                buildBottomActionBar(state),
            ],
          );
        },
      ),
    );
  }

  // เพิ่มสินค้าจากชั้นวาง
  void addProductsFromShelf() {
    showShelfProductSelectorDialog(context, (
      selectedShelfProducts, {
      String? shelfCode,
      String? shelfName,
    }) {
      // แปลงสินค้าจากชั้นวางเป็น ProductWithCopies
      for (var shelfProduct in selectedShelfProducts) {
        // ตรวจสอบว่าสินค้านี้ถูกเลือกไปแล้วหรือไม่
        if (!isProductSelected(shelfProduct.barcode)) {
          final product = ProductBarcodeModel(
            guidfixed: DateTime.now().millisecondsSinceEpoch.toString(),
            barcode: shelfProduct.barcode,
            names: shelfProduct.names,
            itemunitcode: shelfProduct.unitcode ?? '',
            itemunitnames: shelfProduct.unitnames ?? [],
            shelfCode: shelfCode, // เพิ่มข้อมูลรหัสชั้นวาง
            shelfName: shelfName, // เพิ่มข้อมูลชื่อชั้นวาง
          );
          context.read<SelectedProductsBloc>().add(
            AddProductToSelection(product),
          );
        }
      }

      // อัปเดต UI
      setState(() {});
    });
  }

  // สร้างส่วนหัวของรายการ - ออกแบบเรียบง่าย
  Widget buildListHeader(SelectedProductsState state) {
    return Container(
      height: 60, // กำหนดความสูงคงที่
      padding: EdgeInsets.symmetric(horizontal: 12, vertical: 12),
      margin: EdgeInsets.only(bottom: 8),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border.all(color: global.theme.dividerBorderColor, width: 1),
      ),
      child: Row(
        children: [
          // ไอคอนและข้อมูลจำนวน
          Icon(
            Icons.shopping_cart_outlined,
            color: global.theme.appBarColor,
            size: 20,
          ),
          const SizedBox(width: 12),

          // จำนวนรายการและชิ้น
          Expanded(
            child: BlocBuilder<SelectedProductsBloc, SelectedProductsState>(
              builder: (context, state) {
                if (state is SelectedProductsLoaded) {
                  return RichText(
                    text: TextSpan(
                      style: TextStyle(
                        fontSize: 16,
                        color: global.theme.textColor,
                      ),
                      children: [
                        TextSpan(
                          text: '${state.selectedProducts.length}',
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            color: global.theme.appBarColor,
                            fontSize: 18,
                          ),
                        ),
                        TextSpan(text: ' ${global.language('items')} '),
                        TextSpan(
                          text:
                              '(${state.selectedProducts.fold(0, (sum, item) => sum + item.copies)} ${global.language('pieces')})',
                          style: TextStyle(
                            color: global.theme.textSecondaryColor,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  );
                }
                return Text('0 ${global.language('items')}');
              },
            ),
          ),

          // ปุ่มเพิ่มจากชั้นวาง
          TextButton.icon(
            onPressed: addProductsFromShelf,
            icon: Icon(
              Icons.add_shopping_cart,
              size: 18,
              color: global.theme.infoHighlightTextColor,
            ),
            label: Text(
              global.language('add_from_shelf'),
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w500,
                color: global.theme.infoHighlightTextColor,
              ),
            ),
            style: TextButton.styleFrom(
              padding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(6),
                side: BorderSide(color: global.theme.infoHighlightColor),
              ),
            ),
          ),

          // ปุ่มล้างรายการ - ใช้ Visibility แทน BlocBuilder
          SizedBox(width: 8), // spacing คงที่
          BlocBuilder<SelectedProductsBloc, SelectedProductsState>(
            builder: (context, state) {
              final hasProducts =
                  state is SelectedProductsLoaded &&
                  state.selectedProducts.isNotEmpty;
              return SizedBox(
                width: 44, // กำหนดความกว้างคงที่
                height: 36, // กำหนดความสูงคงที่
                child: Visibility(
                  visible: hasProducts,
                  maintainSize: true, // รักษาขนาดแม้ซ่อน
                  maintainAnimation: true,
                  maintainState: true,
                  child: IconButton(
                    onPressed: clearAllProducts,
                    icon: Icon(
                      Icons.clear_all,
                      size: 18,
                      color: global.theme.negativeHighlightTextColor,
                    ),
                    tooltip: global.language("clear_all"),
                    style: IconButton.styleFrom(
                      padding: const EdgeInsets.all(8),
                      minimumSize: const Size(36, 36),
                      tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    ),
                  ),
                ),
              );
            },
          ),
        ],
      ),
    );
  } // สร้างส่วนแสดงเมื่อไม่มีสินค้าที่เลือก

  Widget buildEmptySelectionView() {
    return Center(
      child: Container(
        padding: EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: global.theme.infoHighlightColor,
                shape: BoxShape.circle,
              ),
              child: Icon(
                Icons.shopping_cart_outlined,
                size: 68,
                color: global.theme.appBarColor,
              ),
            ),
            SizedBox(height: 24),
            Text(
              global.language('no_products_selected'),
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: global.theme.textColor,
              ),
            ),
            SizedBox(height: 12),
            Text(
              global.language('select_products_from_left'),
              textAlign: TextAlign.center,
              style: TextStyle(fontSize: 16, color: global.theme.textSecondaryColor),
            ),
            const SizedBox(height: 32),
            Icon(Icons.arrow_back, size: 32, color: global.theme.infoHighlightTextColor),
          ],
        ),
      ),
    );
  }

  // สร้างรายการสินค้าที่เลือก (ListView Item) - Space-efficient version
  Widget buildSelectedProductItem(ProductWithCopies product) {
    return Card(
      elevation: 1,
      margin: const EdgeInsets.symmetric(vertical: 4),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            // Left side - Product info
            Expanded(
              flex: 7,
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Barcode badge - vertical orientation
                  Container(
                    width: 32,
                    padding: EdgeInsets.symmetric(
                      vertical: 6,
                      horizontal: 2,
                    ),
                    decoration: BoxDecoration(
                      color: global.theme.surfaceColor,
                      borderRadius: BorderRadius.circular(4),
                      border: Border.all(
                        color: global.theme.dividerBorderColor,
                        width: 0.5,
                      ),
                    ),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          Icons.qr_code,
                          size: 14,
                          color: global.theme.textColor,
                        ),
                        const SizedBox(height: 2),
                        RotatedBox(
                          quarterTurns: 1,
                          child: Text(
                            product.product.barcode.substring(
                              product.product.barcode.length - 4,
                            ),
                            style: TextStyle(
                              fontSize: 10,
                              fontFamily: 'Monospace',
                              color: global.theme.textColor,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(width: 8),

                  // Product info
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        // Product name
                        Text(
                          global.activeLangName(product.product.name),
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),

                        const SizedBox(height: 2),

                        // Barcode row
                        Text(
                          product.product.barcode,
                          style: TextStyle(
                            fontSize: 11,
                            fontFamily: 'Monospace',
                            color: global.theme.textColor,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),

                        const SizedBox(height: 4),

                        // Unit and Shelf chips in Wrap
                        Wrap(
                          spacing: 4.0,
                          runSpacing: 2.0,
                          children: [
                            // Unit chip
                            Chip(
                              label: Text(
                                global.activeLangName(product.product.unitname),
                                style: TextStyle(
                                  fontSize: 9,
                                  color: global.theme.infoHighlightTextColor,
                                ),
                              ),
                              backgroundColor: global.theme.infoHighlightColor,
                              side: BorderSide(
                                color: global.theme.columnHeaderColor,
                                width: 0.5,
                              ),
                              padding: EdgeInsets.zero,
                              materialTapTargetSize:
                                  MaterialTapTargetSize.shrinkWrap,
                              visualDensity: VisualDensity.compact,
                            ),

                            // Shelf chip (if available)
                            if (product.product.shelfCode != null &&
                                product.product.shelfName != null)
                              Chip(
                                label: Text(
                                  '${product.product.shelfCode} - ${product.product.shelfName}',
                                  style: TextStyle(
                                    fontSize: 9,
                                    color: global.theme.warningHighlightTextColor,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                                backgroundColor: global.theme.warningHighlightColor,
                                side: BorderSide(
                                  color: global.theme.rowEditColor,
                                  width: 0.5,
                                ),
                                padding: EdgeInsets.zero,
                                materialTapTargetSize:
                                    MaterialTapTargetSize.shrinkWrap,
                                visualDensity: VisualDensity.compact,
                              ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),

            // Right side - Controls
            Expanded(
              flex: 5,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  // Compact quantity selector
                  Container(
                    height: 32,
                    decoration: BoxDecoration(
                      color: global.theme.surfaceColor,
                      borderRadius: BorderRadius.circular(4),
                      border: Border.all(
                        color: global.theme.dividerBorderColor,
                        width: 0.5,
                      ),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        // Decrease button
                        InkWell(
                          onTap: product.copies > 1
                              ? () => setState(() => product.copies--)
                              : null,
                          child: Container(
                            width: 28,
                            height: 32,
                            decoration: BoxDecoration(
                              color: product.copies > 1
                                  ? global.theme.negativeHighlightColor
                                  : global.theme.surfaceColor,
                              borderRadius: const BorderRadius.only(
                                topLeft: Radius.circular(4),
                                bottomLeft: Radius.circular(4),
                              ),
                            ),
                            alignment: Alignment.center,
                            child: Icon(
                              Icons.remove,
                              size: 14,
                              color: product.copies > 1
                                  ? global.theme.negativeHighlightTextColor
                                  : global.theme.iconSecondaryColor,
                            ),
                          ),
                        ),

                        // Quantity display
                        InkWell(
                          onTap: () => openEditCopiesDialog(product),
                          child: Container(
                            width: 32,
                            height: 32,
                            alignment: Alignment.center,
                            child: Text(
                              '${product.copies}',
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.bold,
                                color: global.theme.infoHighlightTextColor,
                              ),
                            ),
                          ),
                        ),

                        // Increase button
                        InkWell(
                          onTap: product.copies < 99
                              ? () => setState(() => product.copies++)
                              : null,
                          child: Container(
                            width: 28,
                            height: 32,
                            decoration: BoxDecoration(
                              color: product.copies < 99
                                  ? global.theme.positiveHighlightColor
                                  : global.theme.surfaceColor,
                              borderRadius: const BorderRadius.only(
                                topRight: Radius.circular(4),
                                bottomRight: Radius.circular(4),
                              ),
                            ),
                            alignment: Alignment.center,
                            child: Icon(
                              Icons.add,
                              size: 14,
                              color: product.copies < 99
                                  ? global.theme.positiveHighlightTextColor
                                  : global.theme.iconSecondaryColor,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(width: 6),

                  // Delete button
                  InkWell(
                    onTap: () {
                      context.read<SelectedProductsBloc>().add(
                        RemoveProductFromSelection(product.product.barcode),
                      );
                    },
                    borderRadius: BorderRadius.circular(4),
                    child: Container(
                      padding: EdgeInsets.all(6),
                      decoration: BoxDecoration(
                        color: global.theme.negativeHighlightColor,
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(
                          color: global.theme.negativeHighlightColor,
                          width: 0.5,
                        ),
                      ),
                      child: Icon(
                        Icons.delete_outline,
                        size: 16,
                        color: global.theme.negativeHighlightTextColor,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  // Bottom action bar with buttons
  Widget buildBottomActionBar(SelectedProductsState state) {
    return Container(
      padding: EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        boxShadow: [
          BoxShadow(
            color: global.theme.dividerBorderColor,
            blurRadius: 6,
            offset: const Offset(0, -3),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: ElevatedButton.icon(
              onPressed: openCopiesDialog,
              icon: Icon(Icons.edit),
              label: Text(global.language('edit_all_quantities')),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.amber.shade600,
                foregroundColor: global.theme.onPrimaryColor,
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
            ),
          ),
          SizedBox(width: 16),
          Expanded(
            child: ElevatedButton.icon(
              onPressed: printLabels,
              icon: Icon(Icons.print),
              label: Text(global.language('print_product_label')),
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.infoHighlightTextColor,
                foregroundColor: global.theme.onPrimaryColor,
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
                elevation: 2,
              ),
            ),
          ),
        ],
      ),
    );
  }

  // เปิดหน้าต่างแก้ไขจำนวนพร้อมกันทั้งหมด
  void openCopiesDialog() {
    final TextEditingController copiesController = TextEditingController(
      text: "1",
    );

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('edit_all_quantities')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: copiesController,
              decoration: InputDecoration(
                labelText: global.language('quantity_to_print'),
                border: const OutlineInputBorder(),
              ),
              keyboardType: TextInputType.number,
              inputFormatters: [
                FilteringTextInputFormatter.digitsOnly,
                LengthLimitingTextInputFormatter(2),
              ],
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language("cancel")),
          ),
          ElevatedButton(
            onPressed: () {
              final int copies = int.tryParse(copiesController.text) ?? 1;
              if (copies > 0 && copies <= 99) {
                setState(() {
                  context.read<SelectedProductsBloc>().add(
                    UpdateAllProductsCopies(copies),
                  );
                });
                Navigator.pop(context);
              }
            },
            child: Text(global.language("save")),
          ),
        ],
      ),
    );
  }

  // พิมพ์ฉลากสินค้า
  void printLabels() {
    final currentState = context.read<SelectedProductsBloc>().state;
    if (currentState is SelectedProductsLoaded) {
      if (kDebugMode) {
        AppLogger.debug(
          'printLabels called, selectedProducts count: ${currentState.selectedProducts.length}',
        ); // Debug log
      }
      PrintLabelDialog.show(
        context: context,
        products: currentState.selectedProducts,
        onRegularLabelSelected: _handleRegularLabelPrinting,
        onA4ShelfSmallSelected: _handleA4ShelfSmallPrinting,
        onA4ShelfMediumSelected: _handleA4ShelfMediumPrinting,
        onA4ShelfLargeSelected: _handleA4ShelfLargePrinting,
        onA4ShelfXLargeSelected: _handleA4ShelfXLargePrinting,
      );
    }
  }

  // จัดการการพิมพ์ regular label
  void _handleRegularLabelPrinting() {
    final currentState = context.read<SelectedProductsBloc>().state;
    if (currentState is SelectedProductsLoaded) {
      final List<String> barcodes = currentState.selectedProducts
          .map((p) => p.product.barcode)
          .toList();
      if (kDebugMode) {
        AppLogger.debug(
          'ProductBarcodeShelf: Sending ProductGetByBarcodes with barcodes: $barcodes',
        );
      }

      // ลองเรียก ProductLabelPrintShelf โดยตรงก่อนเพื่อทดสอบ
      if (kDebugMode) {
        AppLogger.debug(
          'ProductBarcodeShelf: Testing direct ProductLabelPrintShelf call',
        );
      }

      // สร้าง ProductBarcodeModel จากข้อมูลที่มี
      List<ProductBarcodeModel> testProducts = [];
      List<int> testCopies = [];

      for (var productWithCopies in currentState.selectedProducts) {
        final product = ProductBarcodeModel(
          guidfixed: DateTime.now().millisecondsSinceEpoch.toString(),
          barcode: productWithCopies.product.barcode,
          names: productWithCopies.product.name,
          itemunitcode: productWithCopies.product.unitcode,
          itemunitnames: productWithCopies.product.unitname,
          prices: [], // ราคาเริ่มต้น
        );
        testProducts.add(product);
        testCopies.add(productWithCopies.copies);
      }

      if (kDebugMode) {
        AppLogger.debug(
          'ProductBarcodeShelf: Calling ProductLabelPrintShelf.showPdfPreview directly with ${testProducts.length} products',
        );
      }

      // เรียก ProductLabelPrintShelf โดยตรง
      ProductLabelPrintShelf.showPdfPreview(
        context,
        testProducts,
        testCopies,
        showImages: true,
      );

      // ยังคงส่ง event ไปยัง ProductBloc ด้วย (สำหรับทดสอบ)
      context.read<ProductBloc>().add(ProductGetByBarcodes(barcodes: barcodes));
      if (kDebugMode) {
        AppLogger.debug(
          'ProductBarcodeShelf: ProductGetByBarcodes event sent successfully',
        );
      }
    }
  }

  // จัดการการพิมพ์ A4 Shelf Small
  void _handleA4ShelfSmallPrinting(PdfColor priceTextColor, bool showBorder) {
    if (kDebugMode) {
      AppLogger.debug(
        'ProductBarcodeShelf: _handleA4ShelfSmallPrinting called',
      );
      AppLogger.debug(
        'ProductBarcodeShelf: priceTextColor = ${priceTextColor == PdfColors.red ? "Red" : "Black"}, showBorder = $showBorder',
      );
    }
    _currentA4FormType = 'small';
    _handleA4FormPrintingDirect('small', priceTextColor, showBorder);
  }

  // จัดการการพิมพ์ A4 Shelf Medium
  void _handleA4ShelfMediumPrinting(PdfColor priceTextColor, bool showBorder) {
    if (kDebugMode) {
      AppLogger.debug(
        'ProductBarcodeShelf: _handleA4ShelfMediumPrinting called',
      );
      AppLogger.debug(
        'ProductBarcodeShelf: priceTextColor = ${priceTextColor == PdfColors.red ? "Red" : "Black"}, showBorder = $showBorder',
      );
    }
    _currentA4FormType = 'medium';
    _handleA4FormPrintingDirect('medium', priceTextColor, showBorder);
  }

  // จัดการการพิมพ์ A4 Shelf Large
  void _handleA4ShelfLargePrinting(PdfColor priceTextColor, bool showBorder) {
    if (kDebugMode) {
      AppLogger.debug(
        'ProductBarcodeShelf: _handleA4ShelfLargePrinting called',
      );
      AppLogger.debug(
        'ProductBarcodeShelf: priceTextColor = ${priceTextColor == PdfColors.red ? "Red" : "Black"}, showBorder = $showBorder',
      );
    }
    _currentA4FormType = 'large';
    _handleA4FormPrintingDirect('large', priceTextColor, showBorder);
  }

  // จัดการการพิมพ์ A4 Shelf XLarge
  void _handleA4ShelfXLargePrinting(PdfColor priceTextColor, bool showBorder) {
    if (kDebugMode) {
      AppLogger.debug(
        'ProductBarcodeShelf: _handleA4ShelfXLargePrinting called',
      );
      AppLogger.debug(
        'ProductBarcodeShelf: priceTextColor = ${priceTextColor == PdfColors.red ? "Red" : "Black"}, showBorder = $showBorder',
      );
    }
    _currentA4FormType = 'xlarge';
    _handleA4FormPrintingDirect('xlarge', priceTextColor, showBorder);
  }

  // ฟังก์ชันใหม่สำหรับเรียก API และแสดง PDF โดยตรง
  void _handleA4FormPrintingDirect(
    String formType,
    PdfColor priceTextColor,
    bool showBorder,
  ) async {
    final currentState = context.read<SelectedProductsBloc>().state;
    if (currentState is SelectedProductsLoaded) {
      final List<String> barcodes = currentState.selectedProducts
          .map((p) => p.product.barcode)
          .toList();
      if (kDebugMode) {
        AppLogger.debug(
          'ProductBarcodeShelf: Getting products for A4 $formType with barcodes: $barcodes',
        );
      }

      try {
        // เรียก API โดยตรงแทนการใช้ BlocConsumer
        final productBloc = context.read<ProductBloc>();
        productBloc.add(ProductGetByBarcodes(barcodes: barcodes));

        // รอ state เปลี่ยน
        await for (final state in productBloc.stream) {
          if (state is ProductGetByBarcodesSuccess) {
            if (kDebugMode) {
              AppLogger.debug(
                'ProductBarcodeShelf: Got products directly, showing A4 $formType',
              );
              AppLogger.debug('Products received: ${state.products.length}');
            }

            // *** แก้ไข: Merge ข้อมูล shelf จาก selectedProducts กลับไปยังข้อมูลที่ได้จาก API ***
            final List<ProductBarcodeModel> productsWithShelf = state.products
                .map((apiProduct) {
                  // หาสินค้าตัวเดียวกันใน selectedProducts
                  final selectedProduct = currentState.selectedProducts
                      .firstWhere(
                        (sp) => sp.product.barcode == apiProduct.barcode,
                        orElse: () => currentState.selectedProducts.first,
                      );

                  // Copy ข้อมูล shelf จาก selectedProduct ไปยัง apiProduct
                  return ProductBarcodeModel(
                    guidfixed: apiProduct.guidfixed,
                    barcode: apiProduct.barcode,
                    names: apiProduct.names,
                    itemunitcode: apiProduct.itemunitcode,
                    itemunitnames: apiProduct.itemunitnames,
                    prices: apiProduct.prices,
                    itemcode: apiProduct.itemcode,
                    groupcode: apiProduct.groupcode,
                    groupnames: apiProduct.groupnames,
                    imageuri: apiProduct.imageuri,
                    // *** เพิ่มข้อมูล shelf จาก selectedProduct ***
                    shelfCode: selectedProduct.product.shelfCode,
                    shelfName: selectedProduct.product.shelfName,
                  );
                })
                .toList();

            final List<int> copies = currentState.selectedProducts
                .map((p) => p.copies)
                .toList();
            _showA4FormByType(
              productsWithShelf, // ใช้ productsWithShelf แทน state.products
              copies,
              priceTextColor,
              showBorder,
            );
            break;
          } else if (state is ProductGetByBarcodesFailed) {
            if (kDebugMode) {
              AppLogger.debug(
                'ProductBarcodeShelf: Failed to get products: ${state.message}',
              );
            }
            global.showErrorSnackBar(context, '${global.language('error_occurred')}: ${state.message}');
            break;
          }
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('ProductBarcodeShelf: Error in direct API call: $e');
        }
      } finally {
        // รีเซ็ต A4 form type
        _currentA4FormType = null;
      }
    }
  }

  // แสดง A4 ฟอร์มตามประเภทที่เลือก
  void _showA4FormByType(
    List<ProductBarcodeModel> products,
    List<int> copies,
    PdfColor priceTextColor,
    bool showBorder,
  ) {
    if (kDebugMode) {
      AppLogger.debug(
        'ProductBarcodeShelf: Showing A4 form type: $_currentA4FormType',
      );
      AppLogger.debug(
        'ProductBarcodeShelf: Using priceTextColor = ${priceTextColor == PdfColors.red ? "Red" : "Black"}, showBorder = $showBorder',
      );
    }

    switch (_currentA4FormType) {
      case 'small':
        if (kDebugMode) {
          AppLogger.debug(
            'ProductBarcodeShelf: Showing ProductLabelPrintA4Shelf',
          );
        }
        ProductLabelPrintA4Shelf.showPdfPreview(
          context,
          products,
          copies,
          showImages: false,
          priceTextColor: priceTextColor,
          showBorder: showBorder,
        );
        break;
      case 'medium':
        if (kDebugMode) {
          AppLogger.debug(
            'ProductBarcodeShelf: Showing ProductLabelPrintA4ShelfMedium',
          );
        }
        ProductLabelPrintA4ShelfMedium.showPdfPreview(
          context,
          products,
          copies,
          priceTextColor: priceTextColor,
          showBorder: showBorder,
        );
        break;
      case 'large':
        if (kDebugMode) {
          AppLogger.debug(
            'ProductBarcodeShelf: Showing ProductLabelPrintA4ShelfLarge',
          );
        }
        ProductLabelPrintA4ShelfLarge.showPdfPreview(
          context,
          products,
          copies,
          priceTextColor: priceTextColor,
          showBorder: showBorder,
        );
        break;
      case 'xlarge':
        if (kDebugMode) {
          AppLogger.debug(
            'ProductBarcodeShelf: Showing ProductLabelPrintA4ShelfXLarge',
          );
        }
        ProductLabelPrintA4ShelfXLarge.showPdfPreview(
          context,
          products,
          copies,
          priceTextColor: priceTextColor,
          showBorder: showBorder,
        );
        break;
      default:
        if (kDebugMode) {
          AppLogger.debug(
            'ProductBarcodeShelf: Unknown A4 form type: $_currentA4FormType, using default',
          );
        }
        ProductLabelPrintA4Shelf.showPdfPreview(
          context,
          products,
          copies,
          showImages: false,
          priceTextColor: priceTextColor,
          showBorder: showBorder,
        );
        break;
    }
  }

  // ล้างรายการสินค้าทั้งหมด
  void clearAllProducts() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('confirm_clear_products')),
        content: Text(global.language('confirm_clear_products_message')),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language("cancel")),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
            onPressed: () {
              context.read<SelectedProductsBloc>().add(
                const ClearAllProducts(),
              );
              Navigator.pop(context);
            },
            child: Text(global.language("clear_all")),
          ),
        ],
      ),
    );
  }

  // เปิดหน้าต่างแก้ไขจำนวนเฉพาะสินค้า
  void openEditCopiesDialog(ProductWithCopies product) {
    final TextEditingController copiesController = TextEditingController(
      text: product.copies.toString(),
    );

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('edit_quantity')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              global.packName(product.product.name),
              style: TextStyle(fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 16),
            TextField(
              controller: copiesController,
              decoration: InputDecoration(
                labelText: global.language('quantity_to_print'),
                border: const OutlineInputBorder(),
              ),
              keyboardType: TextInputType.number,
              inputFormatters: [
                FilteringTextInputFormatter.digitsOnly,
                LengthLimitingTextInputFormatter(2),
              ],
              autofocus: true,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language("cancel")),
          ),
          ElevatedButton(
            onPressed: () {
              final int copies = int.tryParse(copiesController.text) ?? 1;
              if (copies > 0 && copies <= 99) {
                setState(() {
                  product.copies = copies;
                });
                Navigator.pop(context);
              }
            },
            child: Text(global.language("save")),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocProvider(
      providers: [
        BlocProvider(
          create: (context) => ShelfProductSelectorBloc(
            warehouseRepository: WarehouseRepository(),
          ),
        ),
        BlocProvider(
          create: (context) => ProductBloc(
            productRepository: ProductRepository(),
            productSectionRepository: ProductSectionRepository(),
            productBarcodeRepository: ProductBarcodeRepository(),
          ),
        ),
      ],
      child: BlocConsumer<ProductBloc, ProductState>(
        listener: (context, state) {
          if (kDebugMode) {
            AppLogger.debug('=== ProductBarcodeShelf BlocListener START ===');
            AppLogger.debug(
              'ProductBarcodeShelf BlocListener received state: ${state.runtimeType}',
            );
            AppLogger.debug(
              'Current A4 form type at listener start: $_currentA4FormType',
            );
          }
          if (state is ProductGetByBarcodesSuccess) {
            if (kDebugMode) {
              AppLogger.debug(
                'ProductGetByBarcodesSuccess received, showing ProductLabelPrint...',
              );
              AppLogger.debug('Products received: ${state.products.length}');
              AppLogger.debug('Current A4 form type: $_currentA4FormType');
            }

            // แสดง ProductLabelPrint ตามประเภทที่เลือก
            final currentSelectedState = context
                .read<SelectedProductsBloc>()
                .state;
            if (currentSelectedState is SelectedProductsLoaded) {
              final List<int> copies = currentSelectedState.selectedProducts
                  .map((p) => p.copies)
                  .toList();

              // *** แก้ไข: Merge ข้อมูล shelf จาก selectedProducts ***
              final List<ProductBarcodeModel> productsWithShelf = state.products
                  .map((apiProduct) {
                    // หาสินค้าตัวเดียวกันใน selectedProducts
                    final selectedProduct = currentSelectedState
                        .selectedProducts
                        .firstWhere(
                          (sp) => sp.product.barcode == apiProduct.barcode,
                          orElse: () =>
                              currentSelectedState.selectedProducts.first,
                        );

                    // Copy ข้อมูล shelf จาก selectedProduct ไปยัง apiProduct
                    return ProductBarcodeModel(
                      guidfixed: apiProduct.guidfixed,
                      barcode: apiProduct.barcode,
                      names: apiProduct.names,
                      itemunitcode: apiProduct.itemunitcode,
                      itemunitnames: apiProduct.itemunitnames,
                      prices: apiProduct.prices,
                      itemcode: apiProduct.itemcode,
                      groupcode: apiProduct.groupcode,
                      groupnames: apiProduct.groupnames,
                      imageuri: apiProduct.imageuri,
                      // *** เพิ่มข้อมูล shelf จาก selectedProduct ***
                      shelfCode: selectedProduct.product.shelfCode,
                      shelfName: selectedProduct.product.shelfName,
                    );
                  })
                  .toList();

              if (_currentA4FormType == null) {
                // ถ้าไม่มีการเลือก A4 ฟอร์ม แสดงว่าเป็น regular label
                ProductLabelPrintShelf.showPdfPreview(
                  context,
                  productsWithShelf, // ใช้ productsWithShelf แทน state.products
                  copies,
                  showImages: true,
                );
              } else {
                // แสดง A4 ฟอร์มตามประเภทที่เลือก (โค้ดเก่า - ใช้ค่า default)
                _showA4FormByType(
                  productsWithShelf, // ใช้ productsWithShelf แทน state.products
                  copies,
                  PdfColors.black,
                  true,
                );
              }

              // รีเซ็ต A4 form type หลังใช้งาน
              _currentA4FormType = null;
            }
          } else if (state is ProductGetByBarcodesFailed) {
            if (kDebugMode) {
              AppLogger.error('ProductGetByBarcodesFailed: ${state.message}');
            }
            global.showErrorSnackBar(context, '${global.language('error_occurred')}: ${state.message}');
            // รีเซ็ต A4 form type เมื่อเกิดข้อผิดพลาด
            _currentA4FormType = null;
          }
        },
        builder: (context, state) {
          // แสดง debug info เฉพาะ ProductBloc states
          if (state is ProductGetByBarcodesInProgress) {
            if (kDebugMode) {
              AppLogger.debug(
                'ProductBarcodeShelf: ProductGetByBarcodesInProgress detected in builder',
              );
            }
          } else if (state is ProductGetByBarcodesSuccess) {
            if (kDebugMode) {
              AppLogger.debug(
                'ProductBarcodeShelf: ProductGetByBarcodesSuccess detected in builder with ${state.products.length} products',
              );
            }
          }

          // Return the normal UI (ไม่เปลี่ยนแปลง UI ตาม ProductBloc state)
          return LayoutBuilder(
            builder: (context, constraints) {
              return Scaffold(
                appBar: AppBar(
                  backgroundColor: global.theme.appBarColor,
                  title: Text(global.language('print_product_label')),
                  leading: IconButton(
                    icon: Icon(Icons.arrow_back),
                    onPressed: () => Navigator.pop(context),
                  ),
                  actions: [
                    BlocBuilder<SelectedProductsBloc, SelectedProductsState>(
                      builder: (context, state) {
                        final hasProducts =
                            state is SelectedProductsLoaded &&
                            state.selectedProducts.isNotEmpty;
                        return hasProducts
                            ? IconButton(
                                onPressed: printLabels,
                                icon: Icon(Icons.print),
                                tooltip: global.language('print_product_label'),
                              )
                            : const SizedBox.shrink();
                      },
                    ),
                  ],
                ),
                body: (constraints.maxWidth < 800.0)
                    ? SplitView(
                        controller: _splitViewController,
                        gripSize: 8,
                        gripColor: global.theme.appBarColor,
                        gripColorActive: global.theme.infoHighlightTextColor,
                        viewMode: SplitViewMode.Vertical,
                        indicator: const SplitIndicator(
                          viewMode: SplitViewMode.Vertical,
                        ),
                        activeIndicator: const SplitIndicator(
                          viewMode: SplitViewMode.Vertical,
                          isActive: true,
                        ),
                        children: [
                          buildProductListScreen(),
                          buildSelectedProductsList(),
                        ],
                      )
                    : SplitView(
                        controller: _splitViewController,
                        gripSize: 8,
                        gripColor: global.theme.appBarColor,
                        gripColorActive: global.theme.infoHighlightTextColor,
                        viewMode: SplitViewMode.Horizontal,
                        indicator: const SplitIndicator(
                          viewMode: SplitViewMode.Horizontal,
                        ),
                        activeIndicator: const SplitIndicator(
                          viewMode: SplitViewMode.Horizontal,
                          isActive: true,
                        ),
                        children: [
                          buildProductListScreen(),
                          buildSelectedProductsList(),
                        ],
                      ),
              );
            },
          );
        },
      ),
    );
  }
}
