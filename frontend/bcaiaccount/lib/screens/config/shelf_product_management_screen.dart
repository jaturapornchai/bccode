// ignore_for_file: deprecated_member_use

import 'dart:math';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/bloc/shelf_product/shelf_product_bloc.dart';
import 'package:smlaicloud/bloc/warehouse/warehose_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/shelf_model.dart';
import 'package:smlaicloud/model/shelf_product_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/global.dart' as global;

/// รูปแบบข้อมูลสินค้าพร้อมจำนวนที่ต้องการเพิ่มเข้าชั้นวาง
class ProductForShelf {
  final SearchCodeAndNameAndUnitModel product;

  ProductForShelf({
    required this.product,
  });
}

class ShelfProductManagementScreen extends StatefulWidget {
  final String warehouseCode;
  final String locationCode;
  final String shelfCode;
  final String warehouseName;
  final String locationName;
  final String shelfName;
  final WarehouseModel screenData; // เพิ่ม parameter screenData

  const ShelfProductManagementScreen({
    super.key,
    required this.warehouseCode,
    required this.locationCode,
    required this.shelfCode,
    required this.warehouseName,
    required this.locationName,
    required this.shelfName,
    required this.screenData, // เพิ่ม parameter screenData
  });

  @override
  State<ShelfProductManagementScreen> createState() =>
      ShelfProductManagementScreenState();
}

class ShelfProductManagementScreenState
    extends State<ShelfProductManagementScreen>
    with global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  List<ProductBarcodeModel> listData = [];
  List<ShelfProductDisplayModel> productsInShelf = [];
  bool loadingData = false;
  bool loadingShelfProducts = false;
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  final _debouncer = global.Debouncer(1000);
  ScrollController listScrollController = ScrollController();
  ScrollController selectedProductsScrollController = ScrollController();
  ScrollController shelfProductsScrollController = ScrollController();

  // รายการสินค้าที่เลือกสำหรับเพิ่มเข้าชั้นวาง
  List<ProductForShelf> selectedProducts = [];

  String searchText = "";
  late SplitViewController splitViewController;

  @override
  void initState() {
    loadProducts("");
    loadProductsInShelf("");
    listScrollController.addListener(onScrollList);
    splitViewController =
        SplitViewController(limits: [null, WeightLimit(min: 0.2, max: 0.8)]);
    super.initState();
  }

  @override
  void dispose() {
    searchController.dispose();
    searchFocusNode.dispose();
    listScrollController.dispose();
    selectedProductsScrollController.dispose();
    shelfProductsScrollController.dispose();
    super.dispose();
  }

  /// โหลดข้อมูลสินค้าที่สามารถเพิ่มได้
  void loadProducts(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<ProductBarcodeBloc>().add(ProductBarcodeLoadList(
          offset: (listData.isEmpty) ? 0 : listData.length,
          limit: global.loadDataPerPage,
          search: search,
          branchcode: global.companyBranchSelectData.code,
          businesstypecode: global.companyBranchSelectData.businesstype!.code!,
        ));
  }

  /// โหลดรายการสินค้าที่มีอยู่ในชั้นวางแล้ว
  void loadProductsInShelf(String search) {
    setState(() {
      loadingShelfProducts = true;
    });
    context.read<ShelfProductBloc>().add(LoadProductsInShelf(
          warehouseCode: widget.warehouseCode,
          locationCode: widget.locationCode,
          shelfCode: widget.shelfCode,
          limit: global.loadDataPerPage,
          offset: (productsInShelf.isEmpty) ? 0 : productsInShelf.length,
          search: search,
        ));
  }

  /// จัดการกับการเลื่อนรายการสินค้า
  void onScrollList() {
    if (listScrollController.position.pixels >=
            listScrollController.position.maxScrollExtent - 200 &&
        !loadingData) {
      loadProducts(searchText);
    }
  }

  /// ตรวจสอบว่าสินค้าถูกเลือกไปแล้วหรือไม่
  bool isProductSelected(String barcode) {
    return selectedProducts.any((item) => item.product.barcode == barcode);
  }

  /// เพิ่มสินค้าเข้าไปในรายการที่เลือก
  void addProductToSelection(ProductBarcodeModel product) {
    if (!isProductSelected(product.barcode!)) {
      final productForShelf = ProductForShelf(
        product: SearchCodeAndNameAndUnitModel(
          guid: product.guidfixed,
          barcode: product.barcode!,
          code: product.itemcode ?? "",
          name: product.names ?? [],
          unitcode: product.itemunitcode,
          unitname: product.itemunitnames ?? [],
        ),
      );

      setState(() {
        selectedProducts.add(productForShelf);
      });
    }
  }

  /// เพิ่มสินค้าทั้งหมดจากรายการที่แสดงอยู่
  void addAllProducts() {
    if (listData.isEmpty) {
      global.showWarningSnackBar(context, global.language('no_products_to_add'));
      return;
    }

    // ตรวจสอบแต่ละรายการสินค้า
    for (var product in listData) {
      // ตรวจสอบว่าสินค้านี้ถูกเลือกไปแล้วหรือไม่
      if (!isProductSelected(product.barcode!)) {
        // ถ้ายังไม่ได้เลือก ให้เพิ่มเข้าไปในรายการ
        selectedProducts.add(
          ProductForShelf(
            product: SearchCodeAndNameAndUnitModel(
              guid: product.guidfixed,
              barcode: product.barcode!,
              code: product.itemcode ?? "",
              name: product.names ?? [],
              unitcode: product.itemunitcode,
              unitname: product.itemunitnames ?? [],
            ),
          ),
        );
      }
    }

    // อัปเดต UI
    setState(() {});
  }

  /// ลบสินค้าออกจากรายการที่เลือก
  void removeProductFromSelection(ProductForShelf product) {
    setState(() {
      selectedProducts.removeWhere(
          (item) => item.product.barcode == product.product.barcode);
    });

    // Force rebuild ด้วยการ setState อีกครั้งหลังจาก frame เสร็จ
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        setState(() {
          // Force rebuild UI
        });
      }
    });
  }

  /// ล้างรายการสินค้าที่เลือกทั้งหมด
  void clearAllProducts() {
    if (selectedProducts.isEmpty) return;

    showDialog(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: Text(global.language('clear_list')),
        content: Text(global.language('confirm_clear_selected_products')),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
            onPressed: () {
              Navigator.pop(context);
              setState(() {
                selectedProducts.clear();
              });
            },
            child: Text(global.language('confirm')),
          ),
        ],
      ),
    );
  }

  /// สร้าง Widget สำหรับรายการสินค้าที่สามารถเพิ่มได้ (ใช้แบบ ProductBarcodeShelf)
  Widget buildProductListItem(int index, ProductBarcodeModel product) {
    final bool isEvenRow = index % 2 == 0;
    final bool productAlreadySelected = isProductSelected(product.barcode!);

    return Container(
      color: isEvenRow
          ? global.theme.columnAlternateEvenColor
          : global.theme.columnAlternateOddColor,
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          Expanded(
              flex: 4,
              child: Text(product.barcode!,
                  maxLines: 1, overflow: TextOverflow.ellipsis)),
          Expanded(
              flex: 6,
              child: Text(global.activeLangName(product.names!),
                  maxLines: 2, overflow: TextOverflow.ellipsis)),
          Expanded(
              flex: 3,
              child: Text(global.activeLangName(product.itemunitnames!),
                  maxLines: 1, overflow: TextOverflow.ellipsis)),
          Expanded(
              flex: 3,
              child: Text(product.itemcode!,
                  maxLines: 1, overflow: TextOverflow.ellipsis)),
          Expanded(
              flex: 4,
              child: Text(global.activeLangName(product.groupnames!),
                  maxLines: 1, overflow: TextOverflow.ellipsis)),

          // ปุ่มเพิ่มสินค้า - ปิดเฉพาะเมื่อถูกเลือกแล้ว
          IconButton(
            onPressed: productAlreadySelected
                ? null
                : () => addProductToSelection(product),
            icon: Icon(
              Icons.add_circle,
              color: productAlreadySelected ? global.theme.iconSecondaryColor : global.theme.positiveHighlightTextColor,
              size: 28,
            ),
            tooltip: productAlreadySelected
                ? global.language('product_already_selected')
                : global.language('select_this_product'),
          ),
        ],
      ),
    );
  }

  /// สร้างหน้าจอแสดงรายการสินค้า (ใช้แบบ ProductBarcodeShelf)
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
                searchFocusNode.requestFocus();
              },
              onChanged: (value) {
                _debouncer.run(() {
                  setState(() {
                    listData = [];
                  });
                  loadProducts(value);
                });
              },
              autofocus: false,
              focusNode: searchFocusNode,
              controller: searchController,
              decoration: InputDecoration(
                isDense: true,
                contentPadding:
                    EdgeInsets.symmetric(vertical: 12, horizontal: 10),
                border: InputBorder.none,
                hintText: global.language('search'),
              ),
            ),
          ),
          Container(
            color: global.theme.appBarColor,
            height: 6,
          ),

          // ส่วนหัวของตาราง
          Container(
            padding: EdgeInsets.symmetric(horizontal: 10, vertical: 12),
            color: global.theme.columnHeaderColor,
            child: Row(
              children: [
                Expanded(
                    flex: 4,
                    child: Text(global.language("barcode"),
                        style: TextStyle(fontWeight: FontWeight.bold))),
                Expanded(
                    flex: 6,
                    child: Text(global.language("product_name"),
                        style: TextStyle(fontWeight: FontWeight.bold))),
                Expanded(
                    flex: 3,
                    child: Text(global.language("unit"),
                        style: TextStyle(fontWeight: FontWeight.bold))),
                Expanded(
                    flex: 3,
                    child: Text(global.language("item_code"),
                        style: TextStyle(fontWeight: FontWeight.bold))),
                Expanded(
                    flex: 4,
                    child: Text(global.language("product_group"),
                        style: TextStyle(fontWeight: FontWeight.bold))),
                // ปุ่มเพิ่มทั้งหมด
                if (listData.isNotEmpty)
                  IconButton(
                    icon: Icon(Icons.playlist_add, color: global.theme.positiveHighlightTextColor),
                    onPressed: addAllProducts,
                    tooltip: global.language('add_all_products'),
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                  ),
                const SizedBox(width: 8),
              ],
            ),
          ),

          // รายการสินค้า
          Expanded(
            child: ListView.builder(
              controller: listScrollController,
              itemCount: listData.length,
              itemBuilder: (context, index) =>
                  buildProductListItem(index, listData[index]),
            ),
          ),

          // ตัวโหลดข้อมูล
          if (loadingData)
            Center(
              child: Padding(
                padding: EdgeInsets.all(16.0),
                child: LoadingAnimationWidget.staggeredDotsWave(
                  color: global.theme.infoHighlightTextColor,
                  size: 40,
                ),
              ),
            )
        ],
      ),
    );
  }

  /// สร้างส่วนแสดงรายการสินค้าที่เลือก
  Widget buildSelectedProductsList() {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          // ส่วนหัวของรายการ
          buildListHeader(),

          // Line separator
          Divider(
            height: 1,
            thickness: 1,
            color: global.theme.dividerBorderColor,
          ),

          // รายการสินค้าที่เลือก
          Expanded(
            child: selectedProducts.isEmpty
                ? buildEmptySelectionView()
                : LayoutBuilder(builder: (context, constraints) {
                    // Calculate number of columns based on available width
                    // Adjust the divisor (150.0) to control card width
                    final double cardWidth = 150.0;
                    int crossAxisCount =
                        max(1, (constraints.maxWidth / cardWidth).floor());

                    return AnimatedContainer(
                      duration: const Duration(milliseconds: 300),
                      padding: const EdgeInsets.all(8),
                      child: GridView.builder(
                        controller: selectedProductsScrollController,
                        gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                          crossAxisCount: crossAxisCount,
                          childAspectRatio: 1, // Square cards
                          crossAxisSpacing: 8,
                          mainAxisSpacing: 8,
                        ),
                        itemCount: selectedProducts.length,
                        itemBuilder: (context, index) => AnimatedContainer(
                          duration: const Duration(milliseconds: 300),
                          curve: Curves.easeInOut,
                          child: buildSelectedProductGridItem(
                              selectedProducts[index]),
                        ),
                      ),
                    );
                  }),
          ),

          // Bottom action bar
          if (selectedProducts.isNotEmpty) buildBottomActionBar(),
        ],
      ),
    );
  }

  /// สร้างส่วนหัวของรายการ
  Widget buildListHeader() {
    return Container(
      padding: EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        boxShadow: [
          BoxShadow(
            color: global.theme.dividerBorderColor,
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          Icon(
            Icons.shopping_cart,
            color: global.theme.appBarColor,
            size: 24,
          ),
          SizedBox(width: 12),
          Expanded(
            child: Text(
              '${global.language('selected_products')} (${selectedProducts.length})',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: global.theme.appBarColor,
              ),
            ),
          ),
          if (selectedProducts.isNotEmpty)
            IconButton(
              onPressed: clearAllProducts,
              icon: Icon(Icons.delete_sweep),
              tooltip: global.language("clear_all"),
              color: global.theme.negativeHighlightTextColor,
            ),
        ],
      ),
    );
  }

  /// สร้างส่วนแสดงเมื่อไม่มีสินค้าที่เลือก
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
              style: TextStyle(
                fontSize: 16,
                color: global.theme.textSecondaryColor,
              ),
            ),
            const SizedBox(height: 32),
            Icon(
              Icons.arrow_back,
              size: 32,
              color: global.theme.infoHighlightTextColor,
            ),
          ],
        ),
      ),
    );
  }

  /// สร้างรายการสินค้าที่เลือก (GridView Item)
  Widget buildSelectedProductGridItem(ProductForShelf product) {
    return Card(
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
        side: BorderSide(
          color: global.theme.dividerBorderColor,
          width: 1,
        ),
      ),
      child: Stack(
        children: [
          // Main content
          Padding(
            padding: EdgeInsets.all(8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Barcode badge
                Container(
                  padding:
                      EdgeInsets.symmetric(horizontal: 6, vertical: 3),
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
                      Icon(
                        Icons.qr_code,
                        size: 12,
                        color: global.theme.textSecondaryColor,
                      ),
                      const SizedBox(width: 4),
                      Flexible(
                        child: Text(
                          product.product.barcode,
                          style: TextStyle(
                            fontSize: 10,
                            fontFamily: 'Monospace',
                            color: global.theme.textColor,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 8),

                // Product name
                Text(
                  global.activeLangName(product.product.name),
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),

                const SizedBox(height: 4),

                Container(
                  padding:
                      EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: global.theme.infoHighlightColor,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    global.activeLangName(product.product.unitname),
                    style: TextStyle(
                      fontSize: 10,
                      color: global.theme.infoHighlightTextColor,
                    ),
                  ),
                ),
              ],
            ),
          ),

          // Delete button in top right corner
          Positioned(
            top: 0,
            right: 0,
            child: InkWell(
              onTap: () => removeProductFromSelection(product),
              child: Container(
                padding: EdgeInsets.all(4),
                decoration: BoxDecoration(
                  color: global.theme.negativeHighlightColor,
                  borderRadius: BorderRadius.only(
                    topRight: Radius.circular(8),
                    bottomLeft: Radius.circular(8),
                  ),
                ),
                child: Icon(
                  Icons.close,
                  size: 16,
                  color: global.theme.negativeHighlightTextColor,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// สร้าง Bottom Action Bar
  Widget buildBottomActionBar() {
    return Container(
      padding: EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border(top: BorderSide(color: global.theme.dividerBorderColor)),
        boxShadow: [
          BoxShadow(
            color: global.theme.dividerBorderColor,
            blurRadius: 4,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              '${global.language('total')} ${selectedProducts.length} ${global.language('items')}',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          SizedBox(width: 16),
          ElevatedButton(
            onPressed: addProductsToShelf,
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.positiveHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.add_to_photos, size: 20),
                SizedBox(width: 8),
                Text(global.language('add_to_shelf')),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// เพิ่มสินค้าเข้าชั้นวาง - ใช้ WarehouseBloc แทน ShelfProductBloc
  void addProductsToShelf() {
    showDialog(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: Text(global.language('confirm_add_products')),
        content: Text(
            '${global.language('want_to_add_products')} ${selectedProducts.length} ${global.language('items_to_shelf')} "${widget.shelfName}" ${global.language('question_mark')}'),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: global.theme.positiveHighlightTextColor),
            onPressed: () {
              Navigator.pop(context);
              _updateShelfWithProducts();
            },
            child: Text(global.language('confirm')),
          ),
        ],
      ),
    );
  }

  /// อัปเดตข้อมูลชั้นวางด้วยสินค้าใหม่ (Optimized - ส่งเฉพาะ selectedProducts)
  void _updateShelfWithProducts() {
    // แปลง selectedProducts เป็น ShelfProductItemModel
    List<ShelfProductItemModel> newProducts =
        selectedProducts.map((selectedProduct) {
      return ShelfProductItemModel(
        barcode: selectedProduct.product.barcode,
        guidfixed: selectedProduct.product.guid,
        names: selectedProduct.product.name,
        unitcode: selectedProduct.product.unitcode,
        unitnames: selectedProduct.product.unitname,
      );
    }).toList();

    // สร้าง WarehouseModel ใหม่ที่เรียบง่าย
    WarehouseModel updatedWarehouse = WarehouseModel(
      guidfixed: "",
      code: widget.screenData.code,
      names: widget.screenData.names,
      location: [
        LocationModel(
          code: widget.locationCode,
          names: [
            LanguageDataModel(code: "th", name: widget.locationName),
          ],
          shelf: [
            ShelfModel(
              code: widget.shelfCode,
              name: widget.shelfName,
              productitems: newProducts,
            ),
          ],
        ),
      ],
    );

    // ส่งข้อมูลไปยัง WarehouseBloc
    context.read<WarehouseBloc>().add(
          WarehouseUpdate(
            guid: widget.screenData.guidfixed,
            warehouseModel: updatedWarehouse,
          ),
        );
  }

  /// สร้าง Widget สำหรับรายการสินค้าที่สามารถเพิ่มได้
  Widget buildAvailableProductItem(int index, ProductBarcodeModel product) {
    final bool isEvenRow = index % 2 == 0;
    final bool productAlreadySelected = isProductSelected(product.barcode!);

    return Container(
      color: productAlreadySelected
          ? Colors.green.withValues(alpha: 0.2)
          : (isEvenRow ? global.theme.surfaceColor : global.theme.cardColor),
      child: ListTile(
        dense: true,
        leading: Container(
          width: 50,
          height: 50,
          decoration: BoxDecoration(
            color: productAlreadySelected ? global.theme.positiveHighlightTextColor : global.theme.dividerBorderColor,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(
            productAlreadySelected ? Icons.check : Icons.add,
            color: productAlreadySelected ? global.theme.cardColor : global.theme.textSecondaryColor,
          ),
        ),
        title: Text(
          product.barcode!,
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(global.packName(product.names ?? [])),
            // ลบส่วนแสดงราคา เนื่องจาก ProductBarcodeModel ไม่มี sellprice
          ],
        ),
        trailing: productAlreadySelected
            ? Icon(Icons.check_circle, color: global.theme.positiveHighlightTextColor)
            : Icon(Icons.add_circle_outline, color: global.theme.iconSecondaryColor),
        onTap: () => addProductToSelection(product),
      ),
    );
  }

  /// สร้าง Widget สำหรับสินค้าที่มีอยู่ในชั้นวางแล้ว
  Widget buildShelfProductItem(ShelfProductDisplayModel product) {
    return Container(
      color: product.isSelected ? Colors.red.withValues(alpha: 0.2) : null,
      child: CheckboxListTile(
        dense: true,
        value: product.isSelected,
        onChanged: (bool? value) {
          setState(() {
            int index = productsInShelf.indexOf(product);
            if (index >= 0) {
              productsInShelf[index] =
                  product.copyWith(isSelected: value ?? false);
            }
          });
        },
        title: Text(
          product.barcode,
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Text(global.packName(product.names)),
        secondary: Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: product.isSelected ? global.theme.negativeHighlightTextColor : global.theme.positiveHighlightTextColor,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(
            product.isSelected ? Icons.remove : Icons.inventory,
            color: global.theme.onPrimaryColor,
            size: 20,
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Text(
          '${global.language('warehouse')} : ${widget.warehouseName} > ${global.language('location')} : ${widget.locationName} > ${global.language('shelf')} : ${widget.shelfName}',
        ),
        actions: [
          /// save
          IconButton(
            icon: Icon(Icons.save),
            onPressed: () {
              addProductsToShelf();
            },
          ),
          // IconButton(
          //   icon: Icon(Icons.refresh),
          //   onPressed: () {
          //     loadProducts(searchText);
          //     loadProductsInShelf("");
          //   },
          // ),
        ],
      ),
      body: MultiBlocListener(
        listeners: [
          BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
            listener: (context, state) {
              if (state is ProductBarcodeLoadSuccess) {
                setState(() {
                  loadingData = false;
                  if (state.productBarcodes.isNotEmpty) {
                    listData.addAll(state.productBarcodes);
                  }
                });
              } else if (state is ProductBarcodeLoadFailed) {
                setState(() {
                  loadingData = false;
                });
                global.showSnackBar(
                  context,
                  Icon(Icons.error, color: global.theme.onPrimaryColor),
                  state.message,
                  global.theme.negativeHighlightTextColor,
                );
              }
            },
          ),
          BlocListener<WarehouseBloc, WarehouseState>(
            listener: (context, state) {
              if (state is WarehouseUpdateSuccess) {
                global.showSnackBar(
                  context,
                  Icon(Icons.check, color: global.theme.onPrimaryColor),
                  global.language('shelf_update_success'),
                  global.theme.positiveHighlightTextColor,
                );
                setState(() {
                  selectedProducts.clear();
                });
                loadProductsInShelf("");
                // ปิด dialog loading หากมี
              } else if (state is WarehouseUpdateFailed) {
                global.showSnackBar(
                  context,
                  Icon(Icons.error, color: global.theme.onPrimaryColor),
                  '${global.language('shelf_update_failed')}: ${state.message}',
                  global.theme.negativeHighlightTextColor,
                );
              }
            },
          ),
          BlocListener<ShelfProductBloc, ShelfProductState>(
            listener: (context, state) {
              if (state is ShelfProductLoadSuccess) {
                setState(() {
                  loadingShelfProducts = false;
                  productsInShelf = state.products;
                });
                // เพิ่มสินค้าที่มีอยู่ในชั้นวางเข้า selectedProducts อัตโนมัติ
                selectedProducts.clear();
                selectedProducts.addAll(state.products.map((shelfProduct) {
                  return ProductForShelf(
                    product: SearchCodeAndNameAndUnitModel(
                      guid: shelfProduct.guidfixed,
                      barcode: shelfProduct.barcode,
                      code: "",
                      name: shelfProduct.names,
                      unitcode: shelfProduct.unitcode,
                      unitname: shelfProduct.unitnames,
                    ),
                  );
                }).toList());
              } else if (state is ShelfProductLoadFailed) {
                setState(() {
                  loadingShelfProducts = false;
                });
              }
            },
          ),
        ],
        child: LayoutBuilder(
          builder: (context, constraints) {
            return (constraints.maxWidth > 800)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 14,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: global.theme.infoHighlightTextColor,
                    viewMode: SplitViewMode.Horizontal,
                    children: [
                      buildProductListScreen(),
                      buildSelectedProductsList(),
                    ],
                  )
                : DefaultTabController(
                    length: 2,
                    child: Column(
                      children: [
                        TabBar(
                          tabs: [
                            Tab(text: global.language('select_products_tab')),
                            Tab(text: global.language('products_in_shelf_tab')),
                          ],
                        ),
                        Expanded(
                          child: TabBarView(
                            children: [
                              buildProductListScreen(),
                              _buildShelfManagementPanel(),
                            ],
                          ),
                        ),
                      ],
                    ),
                  );
          },
        ),
      ),
    );
  }

  /// Panel สำหรับจัดการสินค้าในชั้นวาง
  Widget _buildShelfManagementPanel() {
    int selectedCount =
        productsInShelf.where((product) => product.isSelected).length;

    return Column(
      children: [
        // Header with actions
        Container(
          color: global.theme.surfaceColor,
          padding: EdgeInsets.all(8),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  '${global.language('products_in_shelf')} (${productsInShelf.length} ${global.language('items')})',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
              if (selectedCount > 0) ...[
                Text(
                  '${global.language('selected_count')}: $selectedCount ${global.language('items')}',
                  style: TextStyle(
                      color: global.theme.negativeHighlightTextColor, fontWeight: FontWeight.bold),
                ),
              ],
            ],
          ),
        ),

        // Products in shelf list
        Expanded(
          child: loadingShelfProducts
              ? Center(
                  child: LoadingAnimationWidget.staggeredDotsWave(
                    color: global.theme.infoHighlightTextColor,
                    size: 50,
                  ),
                )
              : productsInShelf.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.inventory_2_outlined,
                              size: 64, color: global.theme.iconSecondaryColor),
                          SizedBox(height: 16),
                          Text(
                            global.language('no_products_in_shelf'),
                            style: TextStyle(color: global.theme.iconSecondaryColor, fontSize: 16),
                          ),
                        ],
                      ),
                    )
                  : ListView.builder(
                      controller: shelfProductsScrollController,
                      itemCount: productsInShelf.length,
                      itemBuilder: (context, index) {
                        return buildShelfProductItem(productsInShelf[index]);
                      },
                    ),
        ),
      ],
    );
  }
}
