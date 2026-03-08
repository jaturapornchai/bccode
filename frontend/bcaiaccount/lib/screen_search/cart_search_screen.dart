import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/model/cart_model.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/utils/cart_websocket_service.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:intl/intl.dart';

class CartListSearchScreen extends StatefulWidget {
  final Function(List<TransactionDetailModel>, String cartId)? onCartSelected;

  const CartListSearchScreen({
    super.key,
    this.onCartSelected,
  });

  @override
  _CartListSearchScreenState createState() => _CartListSearchScreenState();
}

class _CartListSearchScreenState extends State<CartListSearchScreen> {
  // ใช้บริการ WebSocket แทนการสร้างใหม่
  final CartWebSocketService _cartService = CartWebSocketService();

  // Search functionality variables
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  final _debouncer = global.Debouncer(800);
  String searchText = "";
  bool isSearching = false;

  List<CartModel> carts = [];
  List<CartModel> filteredCarts = []; // เก็บข้อมูลที่ผ่านการค้นหาแล้ว
  bool isLoading = true;
  String? clientId;

  late List<TransactionDetailModel> details = [];
  List<CartItem> cartItems = [];
  WarehouseModel? _currentWarehouse;
  LocationModel? _currentLocation;
  WarehouseModel? _currentDestWarehouse;
  LocationModel? _currentDestLocation;
  String cartName = '';
  String cartId = "";

  ScrollController listScrollController = ScrollController();

  @override
  void initState() {
    super.initState();

    // ตั้งค่า callback สำหรับ CartWebSocketService
    _cartService.initialize(
      onCartListReceived: _handleCartListReceived,
      onCartDetailsReceived: _handleCartDetailsReceived,
      onClientIdReceived: _handleClientIdReceived,
      onConnectionChanged: _handleConnectionChanged,
      onError: _handleError,
    );

    // เชื่อมต่อกับ WebSocket
    _connectToServer();

    // เพิ่ม listener สำหรับ scroll
    listScrollController.addListener(onScrollList);

    // ตั้งค่า search controller และ initial filtered list
    searchController.text = searchText;
    filteredCarts = List.from(carts); // เริ่มต้นด้วยข้อมูลทั้งหมด
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    searchFocusNode.dispose();
    super.dispose();
  }

  // Callbacks สำหรับ CartWebSocketService
  void _handleCartListReceived(List<CartModel> receivedCarts) {
    if (mounted) {
      setState(() {
        carts = receivedCarts;
        isLoading = false;
      });
      _filterCarts(searchText); // ทำการกรองข้อมูลเมื่อได้รับข้อมูลใหม่
    }
  }

  // ฟังก์ชันสำหรับค้นหาและกรองข้อมูล
  void _filterCarts(String query) {
    setState(() {
      searchText = query;

      if (query.isEmpty) {
        filteredCarts = List.from(carts);
      } else {
        filteredCarts = carts.where((cart) {
          final cartName = cart.cartName?.toLowerCase() ?? '';
          final cartId = cart.cartId?.toLowerCase() ?? '';
          final warehouseName =
              global.activeLangName(cart.warehouse?.names ?? []).toLowerCase();
          final locationName =
              global.activeLangName(cart.location?.names ?? []).toLowerCase();
          final destWarehouseName = cart.destWarehouse != null
              ? global.activeLangName(cart.destWarehouse!.names).toLowerCase()
              : '';
          final destLocationName = cart.destLocation != null
              ? global.activeLangName(cart.destLocation!.names).toLowerCase()
              : '';
          final transFlag =
              global.getTransFlagText(cart.cartTransFlag!).toLowerCase();

          final searchQuery = query.toLowerCase();

          return cartName.contains(searchQuery) ||
              cartId.contains(searchQuery) ||
              warehouseName.contains(searchQuery) ||
              locationName.contains(searchQuery) ||
              destWarehouseName.contains(searchQuery) ||
              destLocationName.contains(searchQuery) ||
              transFlag.contains(searchQuery);
        }).toList();
      }
    });
  }

  // ฟังก์ชันสำหรับ clear search
  void _clearSearch() {
    searchController.clear();
    _filterCarts('');
    setState(() {
      isSearching = false;
    });
  }

  void _handleCartDetailsReceived(
      String receivedCartName,
      String receivedCartId,
      Map<String, dynamic> itemDetailsMap,
      WarehouseModel currentWarehouse,
      LocationModel currentLocation,
      WarehouseModel currentDestWarehouse,
      LocationModel currentDestLocation) {
    if (mounted) {
      setState(() {
        cartName = receivedCartName;
        cartId = receivedCartId;
        _currentWarehouse = currentWarehouse;
        _currentLocation = currentLocation;
        _currentDestWarehouse = currentDestWarehouse;
        _currentDestLocation = currentDestLocation;

        // แปลง map เป็น list ของ cart items
        cartItems = itemDetailsMap.values
            .map((json) => CartItem.fromJson(json as Map<String, dynamic>))
            .toList();
      });

      // สร้าง list barcode จากข้อมูล cart details
      List<String> barcodes = itemDetailsMap.values
          .map((item) =>
              (item as Map<String, dynamic>)['barcode']?.toString() ?? '')
          .where((barcode) => barcode.isNotEmpty)
          .toList();

      // ส่ง event ไปยัง bloc
      context
          .read<ProductBarcodeBloc>()
          .add(ProductBarcodeGetByBarcodeList(barcodes: barcodes));
    }
  }

  void _handleClientIdReceived(String receivedClientId) {
    if (mounted) {
      setState(() {
        clientId = receivedClientId;
      });
    }
  }

  void _handleConnectionChanged(bool connected) {
    if (mounted) {
      setState(() {
        isLoading = !connected;
      });
    }
  }

  void _handleError(String errorMessage) {
    if (mounted) {
      global.showErrorSnackBar(context, errorMessage);
    }
  }

  Future<void> _connectToServer() async {
    bool connected = await _cartService.connect(context);
    if (connected) {
      _requestCartList();
    }
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      // เพิ่มฟังก์ชันโหลดข้อมูลเพิ่มเติมหากต้องการ pagination
    }
  }

  void _requestCartList() {
    setState(() => isLoading = true);
    _cartService.requestCartList(global.getShopId());
  }

  Widget _buildCartItem(CartModel cart) {
    final totalQuantity = cart.statistics!.totalQuantity;
    final uniqueItems = cart.statistics!.uniqueItems;
    final cartTransFlag = cart.cartTransFlag;
    final cartName = cart.cartName;
    final createAt =
        DateFormat('dd/MM/yyyy').format(DateTime.parse(cart.createdAt!));

    return GestureDetector(
      onTap: () async {
        if (widget.onCartSelected != null) {
          _cartService.requestCartDetails(cart.cartId!, global.getShopId());
        }
      },
      child: Container(
        decoration: BoxDecoration(
          border: const Border(
            bottom: BorderSide(width: 1.0, color: Colors.grey),
          ),
        ),
        padding: const EdgeInsets.only(top: 8, left: 10, right: 10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Row 1 - Main information (single line per column)
            Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                // Cart Name Column
                Expanded(
                  flex: 5,
                  child: Text(
                    cartName!,
                    style: const TextStyle(fontWeight: FontWeight.w500),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),

                // Type Column
                Expanded(
                  flex: 4,
                  child: Text(
                    global.getTransFlagText(cartTransFlag!),
                    style: TextStyle(
                      fontSize: 13,
                      color: _getTransFlagColor(cartTransFlag),
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),

                // Origin Location Column
                Expanded(
                  flex: 7,
                  child: Text(
                    '${global.language("warehouse")}: ${global.activeLangName(cart.warehouse?.names ?? [])}',
                    style: const TextStyle(fontSize: 13),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),

                // Destination Location Column
                Expanded(
                  flex: 7,
                  child: cartTransFlag == 72 && cart.destWarehouse != null
                      ? Text(
                          '${global.language("warehouse")}: ${global.activeLangName(cart.destWarehouse?.names ?? [])}',
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.blue[700],
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        )
                      : const SizedBox(),
                ),

                // Quantity Column
                Expanded(
                  flex: 3,
                  child: Text(
                    '$uniqueItems รายการ',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey[600],
                    ),
                    maxLines: 1,
                  ),
                ),
              ],
            ),

            // Row 2 - Additional information (separate columns)
            Padding(
              padding: const EdgeInsets.only(bottom: 8, top: 4),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  // Cart Date Column
                  Expanded(
                    flex: 5,
                    child: Text(
                      createAt,
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey[600],
                      ),
                    ),
                  ),

                  // Cart ID Column
                  Expanded(
                    flex: 4,
                    child: Text(
                      cart.cartId!,
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey[600],
                      ),
                      maxLines: 1,
                    ),
                  ),

                  // Origin Location Detail
                  Expanded(
                    flex: 7,
                    child: cart.location?.names != null
                        ? Text(
                            '${global.language("location")}: ${global.activeLangName(cart.location?.names ?? [])}',
                            style: const TextStyle(fontSize: 12),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          )
                        : const SizedBox(),
                  ),

                  // Destination Location Detail
                  Expanded(
                    flex: 7,
                    child: (cartTransFlag == 72 &&
                            cart.destLocation?.names != null)
                        ? Text(
                            '${global.language("location")}: ${global.activeLangName(cart.destLocation?.names ?? [])}',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.blue[700],
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          )
                        : const SizedBox(),
                  ),

                  // Total Pieces Column
                  Expanded(
                    flex: 3,
                    child: Text(
                      '$totalQuantity ${global.language("pieces")}',
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey[600],
                      ),
                      maxLines: 1,
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

  Color _getTransFlagColor(int cartTransFlag) {
    switch (cartTransFlag) {
      case 56:
        return Colors.orange;
      case 58:
        return Colors.purple;
      case 66:
        return Colors.green;
      case 72:
        return Colors.blue;
      default:
        return Colors.grey;
    }
  }

  Widget _buildDisconnectedScreen() {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Text(global.language("cart_system"),
            style: const TextStyle(fontWeight: FontWeight.bold)),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error, size: 48, color: Colors.red),
            SizedBox(height: 16),
            Text(global.language("cannot_connect_websocket"),
                style: const TextStyle(fontSize: 18)),
            SizedBox(height: 16),
            ElevatedButton(
              onPressed: _connectToServer,
              child: Text(global.language("try_reconnect")),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (!_cartService.isSocketConnected()) {
      return _buildDisconnectedScreen();
    }

    return BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
      listener: (context, state) {
        if (state is ProductBarcodeGetByBarcodeListSuccess) {
          final formattedDate = DateTime.now().toUtc().toIso8601String();
          details = []; // เคลียร์ details เก่า

          // ใช้ cartItems เป็นหลักในการวนลูป
          for (var cartItem in cartItems) {
            // หา product barcode ที่ตรงกับ cartItem
            final productBarcode = state.productBarcodes.firstWhere(
              (pb) => pb.barcode == cartItem.barcode,
            );

            details.add(
              TransactionDetailModel(
                docref: cartName,
                docdatetime: formattedDate,
                docrefdatetime: formattedDate,
                itemguid: productBarcode.guidfixed,
                barcode: cartItem.barcode,
                itemcode: productBarcode.itemcode ?? "",
                itemnames: productBarcode.names,
                unitcode: cartItem.itemunitcode,
                qty: cartItem.quantity,
                price: 0,
                discount: '',
                sumofcost: 0,
                sumamount: 0,
                remark: '',
                linenumber: 0,
                whcode: _currentWarehouse!.code,
                whnames: _currentWarehouse!.names,
                shelfcode: '',
                locationcode: _currentLocation!.code,
                locationnames: _currentLocation!.names,
                totalvaluevat: 0,
                totalqty: 0,
                standvalue: productBarcode.standvalue!,
                dividevalue: productBarcode.dividevalue!,
                multiunit: true,
                unitnames: cartItem.itemunitnames,
                calcflag: 1,
                vattype: 0,
                averagecost: 0,
                sumamountexcludevat: 0,
                discountamount: 0,
                ispos: 0,
                laststatus: 0,
                itemtype: 0,
                inquirytype: 0,
                priceexcludevat: 0,
                taxtype: productBarcode.taxtype!,
                vatcal: productBarcode.vatcal,
                towhcode: _currentDestWarehouse!.code,
                towhnames: _currentDestWarehouse!.names,
                tolocationcode: _currentDestLocation!.code,
                tolocationnames: _currentDestLocation!.names,
                refbarcodes: productBarcode.refbarcodes,
                manufacturerguid: productBarcode.manufacturerguid,
                description: cartItem.description,
                imageuri: cartItem.imageuri,
              ),
            );
          }

          if (widget.onCartSelected != null) {
            widget.onCartSelected!(details, cartId);
          }
          Navigator.of(context).pop();
        } else if (state is ProductBarcodeGetByBarcodeListFailed) {
          setState(() => isLoading = false);
          global.showErrorSnackBar(context, '${global.language("cannot_fetch_data")} ${state.message}');
        }
      },
      child: Scaffold(
        appBar: AppBar(
          backgroundColor: global.theme.appBarColor,
          title: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(global.language("shopping_cart"),
                  style: const TextStyle(fontWeight: FontWeight.bold)),
            ],
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: _requestCartList,
              tooltip: global.language("refresh_data"),
            ),
          ],
        ),
        body: Column(
          children: [
            // Search container
            Container(
              padding: const EdgeInsets.fromLTRB(10, 8, 10, 8),
              color: global.theme.appBarColor,
              child: Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(8),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.grey.withValues(alpha: 0.3),
                      spreadRadius: 1,
                      blurRadius: 4,
                      offset: const Offset(0, 2),
                    ),
                  ],
                  border: Border.all(
                    color: Colors.grey.withValues(alpha: 0.3),
                    width: 1,
                  ),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    Expanded(
                      child: TextFormField(
                        controller: searchController,
                        focusNode: searchFocusNode,
                        autofocus: false,
                        onChanged: (value) {
                          final trimmedValue = value.trim();

                          setState(() {
                            isSearching = trimmedValue.isNotEmpty;
                          });

                          _debouncer.run(() {
                            _filterCarts(trimmedValue);
                            if (mounted) {
                              setState(() {
                                isSearching = false;
                              });
                            }
                          });
                        },
                        onFieldSubmitted: (value) {
                          final trimmedValue = value.trim();
                          _filterCarts(trimmedValue);
                          setState(() {
                            isSearching = false;
                          });
                          searchFocusNode.requestFocus();
                        },
                        decoration: InputDecoration(
                          isDense: true,
                          contentPadding: const EdgeInsets.symmetric(
                              horizontal: 8, vertical: 12),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(6),
                            borderSide: BorderSide.none,
                          ),
                          filled: true,
                          fillColor: Colors.grey.shade50,
                          hintText:
                              global.language("search_cart_hint"),
                          hintStyle: TextStyle(
                            color: Colors.grey.shade500,
                            fontSize: 14,
                          ),
                          prefixIcon: Icon(
                            Icons.search,
                            color: Colors.grey.shade600,
                            size: 20,
                          ),
                          suffixIcon: isSearching
                              ? const Padding(
                                  padding: EdgeInsets.all(8.0),
                                  child: SizedBox(
                                    width: 20,
                                    height: 20,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                    ),
                                  ),
                                )
                              : searchController.text.isNotEmpty
                                  ? IconButton(
                                      icon: const Icon(Icons.clear, size: 18),
                                      onPressed: _clearSearch,
                                    )
                                  : null,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),

            // Table header
            Container(
              padding:
                  const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
              decoration: BoxDecoration(
                  color: global.theme.columnHeaderColor,
                  border: const Border(
                    bottom: BorderSide(width: 1.0, color: Colors.grey),
                  )),
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("cart_name"),
                      style: const TextStyle(
                          color: Colors.black, fontWeight: FontWeight.bold),
                    ),
                  ),
                  Expanded(
                    flex: 4,
                    child: Text(
                      global.language("type"),
                      style: const TextStyle(
                          color: Colors.black, fontWeight: FontWeight.bold),
                    ),
                  ),
                  Expanded(
                    flex: 7,
                    child: Text(
                      global.language("origin"),
                      style: const TextStyle(
                          color: Colors.black, fontWeight: FontWeight.bold),
                    ),
                  ),
                  Expanded(
                    flex: 7,
                    child: Text(
                      global.language("destination"),
                      style: const TextStyle(
                          color: Colors.black, fontWeight: FontWeight.bold),
                    ),
                  ),
                  Expanded(
                    flex: 3,
                    child: Text(
                      global.language("quantity"),
                      style: const TextStyle(
                          color: Colors.black, fontWeight: FontWeight.bold),
                    ),
                  ),
                ],
              ),
            ),

            // List or loading indicator
            Expanded(
              child: isLoading
                  ? const Center(
                      child: CircularProgressIndicator(),
                    )
                  : RefreshIndicator(
                      color: Colors.blue,
                      backgroundColor: Colors.white,
                      strokeWidth: 3.0,
                      onRefresh: () async {
                        _requestCartList();
                      },
                      child: filteredCarts.isEmpty
                          ? _buildEmptyState()
                          : ListView.builder(
                              padding: EdgeInsets.zero,
                              itemCount: filteredCarts.length,
                              controller: listScrollController,
                              itemBuilder: (context, index) {
                                return _buildCartItem(filteredCarts[index]);
                              },
                            ),
                    ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    final bool hasSearchText = searchText.isNotEmpty;
    final bool hasOriginalData = carts.isNotEmpty;

    return CustomScrollView(
      slivers: [
        SliverFillRemaining(
          child: Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  hasSearchText
                      ? Icons.search_off
                      : Icons.shopping_cart_outlined,
                  size: 48,
                  color: Colors.grey[400],
                ),
                SizedBox(height: 16),
                Text(
                  hasSearchText && hasOriginalData
                      ? global.language("cart_not_found_search")
                      : global.language("cart_not_found"),
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                    color: Colors.grey[600],
                  ),
                ),
                const SizedBox(height: 8),
                if (hasSearchText && hasOriginalData) ...[
                  Text(
                    global.language("try_different_search"),
                    style: TextStyle(
                      fontSize: 14,
                      color: Colors.grey[500],
                    ),
                  ),
                  SizedBox(height: 16),
                  OutlinedButton.icon(
                    onPressed: _clearSearch,
                    icon: const Icon(
                      Icons.clear,
                      color: Colors.orange,
                      size: 16,
                    ),
                    label: Text(
                      global.language("clear_search"),
                      style: const TextStyle(
                        color: Colors.orange,
                        fontSize: 14,
                      ),
                    ),
                    style: OutlinedButton.styleFrom(
                      side: const BorderSide(color: Colors.orange),
                    ),
                  ),
                ] else ...[
                  OutlinedButton.icon(
                    onPressed: _requestCartList,
                    icon: const Icon(
                      Icons.refresh,
                      color: Colors.blue,
                      size: 16,
                    ),
                    label: Text(
                      global.language("refresh_data"),
                      style: const TextStyle(
                        color: Colors.blue,
                        fontSize: 14,
                      ),
                    ),
                    style: OutlinedButton.styleFrom(
                      side: const BorderSide(color: Colors.blue),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ],
    );
  }
}
