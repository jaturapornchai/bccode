import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/shelf_product_selector/shelf_product_selector_bloc.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/model/warehouse_location_model.dart';
import 'package:smlaicloud/model/shelf_model.dart';
import 'package:smlaicloud/model/shelf_product_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import '../../global.dart' as global;

class ShelfProductSelector extends StatefulWidget {
  final Function(
    List<ShelfProductItemModel> products, {
    String? shelfCode,
    String? shelfName,
  })
  onProductsSelected;

  const ShelfProductSelector({super.key, required this.onProductsSelected});

  @override
  State<ShelfProductSelector> createState() => _ShelfProductSelectorState();
}

class _ShelfProductSelectorState extends State<ShelfProductSelector> {
  // Track deleted products
  Set<String> deletedProducts = {};

  @override
  void initState() {
    super.initState();
    // Reset selection first, then load warehouses after a microtask delay
    context.read<ShelfProductSelectorBloc>().add(const ResetSelection());
    Future.microtask(() {
      if (mounted) {
        context.read<ShelfProductSelectorBloc>().add(const LoadWarehouses());
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 500,
      height: 600,
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                global.language('select_product_from_shelf'),
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              IconButton(
                icon: const Icon(Icons.close),
                onPressed: () => Navigator.of(context).pop(),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Selection dropdowns
          Expanded(
            child:
                BlocBuilder<
                  ShelfProductSelectorBloc,
                  ShelfProductSelectorState
                >(
                  builder: (context, state) {
                    return Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Warehouse dropdown
                        _buildWarehouseDropdown(state),
                        const SizedBox(height: 12),

                        // Location dropdown
                        _buildLocationDropdown(state),
                        const SizedBox(height: 12),

                        // Shelf dropdown
                        _buildShelfDropdown(state),
                        const SizedBox(height: 16),

                        // Products list
                        Expanded(child: _buildProductsList(state)),

                        const SizedBox(height: 16),

                        // Action buttons
                        _buildActionButtons(state),
                      ],
                    );
                  },
                ),
          ),
        ],
      ),
    );
  }

  Widget _buildWarehouseDropdown(ShelfProductSelectorState state) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          global.language('select_warehouse'),
          style: Theme.of(
            context,
          ).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 8),
        Container(
          width: double.infinity,
          decoration: BoxDecoration(
            border: Border.all(color: Colors.grey.shade100),
            borderRadius: BorderRadius.circular(4),
          ),
          child: state.warehouseLoadStatus == LoadingStatus.loading
              ? Container(
                  height: 48,
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  child: Row(
                    children: [
                      SizedBox(
                        width: 20,
                        height: 20,
                        child: LoadingAnimationWidget.staggeredDotsWave(
                          color: Theme.of(context).primaryColor,
                          size: 20,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(global.language('loading_warehouse')),
                    ],
                  ),
                )
              : DropdownButtonHideUnderline(
                  child: DropdownButton<WarehouseModel>(
                    isExpanded: true,
                    value: state.selectedWarehouse != null
                        ? state.warehouses.firstWhere(
                            (w) =>
                                w.guidfixed ==
                                state.selectedWarehouse!.guidfixed,
                            orElse: () => state.selectedWarehouse!,
                          )
                        : null,
                    hint: Padding(
                      padding: EdgeInsets.symmetric(horizontal: 12),
                      child: Text(global.language('select_warehouse')),
                    ),
                    items: state.warehouses.map((warehouse) {
                      String warehouseName = _getLanguageName(warehouse.names);
                      return DropdownMenuItem(
                        value: warehouse,
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 12),
                          child: Text('${warehouse.code} - $warehouseName'),
                        ),
                      );
                    }).toList(),
                    onChanged: (warehouse) {
                      if (warehouse != null) {
                        context.read<ShelfProductSelectorBloc>().add(
                          SelectWarehouse(warehouse: warehouse),
                        );
                      }
                    },
                  ),
                ),
        ),
      ],
    );
  }

  Widget _buildLocationDropdown(ShelfProductSelectorState state) {
    bool isEnabled = state.selectedWarehouse != null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          global.language('select_storage_location'),
          style: Theme.of(context).textTheme.titleSmall?.copyWith(
            fontWeight: FontWeight.w600,
            color: isEnabled ? null : Colors.grey,
          ),
        ),
        const SizedBox(height: 8),
        Container(
          width: double.infinity,
          decoration: BoxDecoration(
            border: Border.all(
              color: isEnabled ? Colors.grey.shade300 : Colors.grey.shade200,
            ),
            borderRadius: BorderRadius.circular(4),
            color: isEnabled ? null : Colors.grey.shade50,
          ),
          child: !isEnabled
              ? Container(
                  height: 48,
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      global.language('please_select_warehouse'),
                      style: TextStyle(color: Colors.grey),
                    ),
                  ),
                )
              : state.locationLoadStatus == LoadingStatus.loading
              ? Container(
                  height: 48,
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  child: Row(
                    children: [
                      SizedBox(
                        width: 20,
                        height: 20,
                        child: LoadingAnimationWidget.staggeredDotsWave(
                          color: Theme.of(context).primaryColor,
                          size: 20,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(global.language('loading_locations')),
                    ],
                  ),
                )
              : DropdownButtonHideUnderline(
                  child: DropdownButton<WarehouseLocationModel>(
                    isExpanded: true,
                    value: state.selectedLocation != null
                        ? state.locations.firstWhere(
                            (l) =>
                                l.guidfixed ==
                                state.selectedLocation!.guidfixed,
                            orElse: () => state.selectedLocation!,
                          )
                        : null,
                    hint: Padding(
                      padding: EdgeInsets.symmetric(horizontal: 12),
                      child: Text(global.language('select_storage_location')),
                    ),
                    items: state.locations.map((location) {
                      String locationName = _getLanguageName(
                        location.locationnames,
                      );
                      return DropdownMenuItem(
                        value: location,
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 12),
                          child: Text(
                            '${location.locationcode} - $locationName',
                          ),
                        ),
                      );
                    }).toList(),
                    onChanged: (location) {
                      if (location != null) {
                        context.read<ShelfProductSelectorBloc>().add(
                          SelectLocation(location: location),
                        );
                      }
                    },
                  ),
                ),
        ),
      ],
    );
  }

  Widget _buildShelfDropdown(ShelfProductSelectorState state) {
    bool isEnabled = state.selectedLocation != null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          global.language('select_shelf'),
          style: Theme.of(context).textTheme.titleSmall?.copyWith(
            fontWeight: FontWeight.w600,
            color: isEnabled ? null : Colors.grey,
          ),
        ),
        const SizedBox(height: 8),
        Container(
          width: double.infinity,
          decoration: BoxDecoration(
            border: Border.all(
              color: isEnabled ? Colors.grey.shade300 : Colors.grey.shade200,
            ),
            borderRadius: BorderRadius.circular(4),
            color: isEnabled ? null : Colors.grey.shade50,
          ),
          child: !isEnabled
              ? Container(
                  height: 48,
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      global.language('please_select_location'),
                      style: TextStyle(color: Colors.grey),
                    ),
                  ),
                )
              : DropdownButtonHideUnderline(
                  child: DropdownButton<ShelfModel>(
                    isExpanded: true,
                    value: state.selectedShelf != null
                        ? state.shelves.firstWhere(
                            (s) => s.code == state.selectedShelf!.code,
                            orElse: () => state.selectedShelf!,
                          )
                        : null,
                    hint: Padding(
                      padding: EdgeInsets.symmetric(horizontal: 12),
                      child: Text(global.language('select_shelf')),
                    ),
                    items: state.shelves.map((shelf) {
                      int productCount = shelf.productitems?.length ?? 0;
                      return DropdownMenuItem(
                        value: shelf,
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 12),
                          child: Row(
                            children: [
                              Text('${shelf.code} - ${shelf.name}'),
                              const Spacer(),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 2,
                                ),
                                decoration: BoxDecoration(
                                  color: productCount > 0
                                      ? Colors.green.shade100
                                      : Colors.grey.shade100,
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Text(
                                  '$productCount ${global.language("items")}',
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: productCount > 0
                                        ? Colors.green.shade700
                                        : Colors.grey.shade600,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    }).toList(),
                    onChanged: (shelf) {
                      if (shelf != null) {
                        context.read<ShelfProductSelectorBloc>().add(
                          SelectShelf(shelf: shelf),
                        );
                      }
                    },
                  ),
                ),
        ),
      ],
    );
  }

  Widget _buildProductsList(ShelfProductSelectorState state) {
    if (state.selectedShelf == null) {
      return Container(
        width: double.infinity,
        decoration: BoxDecoration(
          border: Border.all(color: Colors.grey.shade200),
          borderRadius: BorderRadius.circular(8),
          color: Colors.grey.shade50,
        ),
        child: Center(
          child: Padding(
            padding: const EdgeInsets.all(32),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.inventory_2_outlined, size: 48, color: Colors.grey),
                const SizedBox(height: 16),
                Text(
                  global.language('please_select_shelf'),
                  style: TextStyle(color: Colors.grey, fontSize: 16),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
        ),
      );
    }

    if (state.productLoadStatus == LoadingStatus.loading) {
      return Container(
        width: double.infinity,
        decoration: BoxDecoration(
          border: Border.all(color: Colors.grey.shade200),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              LoadingAnimationWidget.staggeredDotsWave(
                color: Theme.of(context).primaryColor,
                size: 32,
              ),
              const SizedBox(height: 16),
              Text(global.language('loading_products')),
            ],
          ),
        ),
      );
    }

    if (state.shelfProducts.isEmpty) {
      return Container(
        width: double.infinity,
        decoration: BoxDecoration(
          border: Border.all(color: Colors.grey.shade200),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Center(
          child: Padding(
            padding: EdgeInsets.all(32),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.inbox_outlined, size: 48, color: Colors.grey),
                SizedBox(height: 16),
                Text(
                  global.language('no_products_in_shelf'),
                  style: TextStyle(color: Colors.grey, fontSize: 16),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.grey.shade200),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Products header
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.grey.shade50,
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(8),
                topRight: Radius.circular(8),
              ),
            ),
            child: Row(
              children: [
                const Icon(Icons.inventory_2, size: 20),
                const SizedBox(width: 8),
                Text(
                  '${global.language("products_in_shelf")} (${state.shelfProducts.where((p) => !deletedProducts.contains(p.barcode)).length} ${global.language("items")})',
                  style: const TextStyle(fontWeight: FontWeight.w600),
                ),
              ],
            ),
          ),

          // Products list
          Expanded(
            child: ListView.builder(
              itemCount: state.shelfProducts
                  .where((p) => !deletedProducts.contains(p.barcode))
                  .length,
              itemBuilder: (context, index) {
                final filteredProducts = state.shelfProducts
                    .where((p) => !deletedProducts.contains(p.barcode))
                    .toList();
                final product = filteredProducts[index];
                final productName = _getLanguageName(product.names);

                return ListTile(
                  leading: Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: Colors.blue.shade50,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      Icons.inventory,
                      color: Colors.blue,
                      size: 20,
                    ),
                  ),
                  title: Text(
                    productName.isNotEmpty ? productName : global.language('product_name_not_specified'),
                    style: const TextStyle(fontWeight: FontWeight.w500),
                  ),
                  subtitle: Text(
                    '${global.language("barcode_label")}: ${product.barcode}',
                    style: TextStyle(color: Colors.grey.shade600),
                  ),
                  trailing: IconButton(
                    icon: Icon(
                      Icons.delete_outline,
                      color: Colors.red.shade600,
                      size: 20,
                    ),
                    onPressed: () => _deleteSingleProduct(product),
                    tooltip: global.language('delete_this_product'),
                  ),
                  dense: true,
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildActionButtons(ShelfProductSelectorState state) {
    // Filter out deleted products
    final availableProducts = state.shelfProducts
        .where((product) => !deletedProducts.contains(product.barcode))
        .toList();

    return Row(
      children: [
        Expanded(
          child: OutlinedButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(global.language('cancel')),
          ),
        ),
        SizedBox(width: 12),
        Expanded(
          child: ElevatedButton(
            onPressed:
                availableProducts.isNotEmpty && state.selectedShelf != null
                ? () {
                    // ส่งข้อมูลสินค้าพร้อมกับข้อมูล shelf
                    _selectProductsWithShelfInfo(
                      availableProducts,
                      state.selectedShelf!,
                    );
                  }
                : null,
            child: Text(
              availableProducts.isNotEmpty
                  ? '${global.language("add_all_products")} (${availableProducts.length})'
                  : global.language('add_product'),
            ),
          ),
        ),
      ],
    );
  }

  void _selectProductsWithShelfInfo(
    List<ShelfProductItemModel> products,
    ShelfModel selectedShelf,
  ) {
    // ส่งข้อมูลสินค้าพร้อมกับข้อมูล shelf code และ name
    widget.onProductsSelected(
      products,
      shelfCode: selectedShelf.code,
      shelfName: selectedShelf.name,
    );
    Navigator.of(context).pop();
  }

  // Delete single product
  void _deleteSingleProduct(ShelfProductItemModel product) {
    // Mark product as deleted in local state immediately
    setState(() {
      deletedProducts.add(product.barcode);
    });
  }

  String _getLanguageName(List<LanguageDataModel> names) {
    if (names.isEmpty) return '';

    // Try to find Thai name first
    final thaiName = names.firstWhere(
      (name) => name.code == 'th',
      orElse: () => names.first,
    );

    return thaiName.name;
  }
}

// Dialog function to show the shelf product selector
Future<void> showShelfProductSelectorDialog(
  BuildContext context,
  Function(
    List<ShelfProductItemModel> products, {
    String? shelfCode,
    String? shelfName,
  })
  onProductsSelected,
) {
  return showDialog(
    context: context,
    barrierDismissible: false,
    builder: (context) => Dialog(
      child: ShelfProductSelector(onProductsSelected: onProductsSelected),
    ),
  );
}
