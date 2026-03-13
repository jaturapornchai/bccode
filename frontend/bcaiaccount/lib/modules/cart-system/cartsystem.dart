import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'bloc/cart_cubit.dart';
import 'bloc/product_search_cubit.dart';
import 'services/mongodb_cart_service.dart';
import 'services/pgsql_product_service.dart';
import 'services/product_search_api.dart';
import 'models/cart_model.dart';
import 'models/cart_system_type.dart';
import 'models/cart_type.dart';
import 'models/debtor_model.dart';
import 'models/creditor_model.dart';
import 'widgets/product_search_widget.dart';
import 'widgets/cart_list_widget.dart';
import 'widgets/debtor_selector_dialog.dart';
import 'widgets/creditor_selector_dialog.dart';
import 'widgets/warehouse_selector_dialog.dart';
import 'widgets/cart_detail_screen.dart';
import 'widgets/cart_selector_dialog.dart';
import '../../global.dart' as global;
import '../../utils/logger/app_logger.dart';

/// Cart System Screen ใหม่ - แบบ Tab 3 ระบบ (ขาย, ซื้อ, คงคลัง)
class CartSystemScreen extends StatefulWidget {
  const CartSystemScreen({super.key});

  @override
  State<CartSystemScreen> createState() => _CartSystemScreenState();
}

class _CartSystemScreenState extends State<CartSystemScreen> with global.ThemeRefreshMixin {
  String? _userEmail;
  String? _shopId;

  // สถานะการแสดงผล (toggle ระหว่างค้นหาสินค้า กับ จัดการตะกร้า)
  bool _showingCartList = false;

  // ขนาดการ์ดสินค้า (1=เล็ก, 2=กลาง, 3=ใหญ่)
  int _cardZoomLevel = 2;

  // ระบบที่กำลังแสดงอยู่ (Default: ขาย)
  CartSystemType _selectedSystemType = CartSystemType.sales;

  // ตะกร้าทั้งหมดในแต่ละระบบ (สามารถมีหลายตะกร้า)
  List<CartModel> _salesCarts = []; // ตะกร้าขายทั้งหมด
  List<CartModel> _purchaseCarts = []; // ตะกร้าซื้อทั้งหมด
  List<CartModel> _inventoryCarts = []; // ตะกร้าคงคลังทั้งหมด

  // ตะกร้าที่กำลัง active ในแต่ละระบบ
  CartModel? _activeSalesCart;
  CartModel? _activePurchaseCart;
  CartModel? _activeInventoryCart;

  // ข้อมูลที่เลือกสำหรับแต่ละระบบ
  DebtorModel? _selectedDebtor;
  CreditorModel? _selectedCreditor;

  // ข้อมูลคลัง (ใช้ร่วมกันทุกระบบ)
  String _selectedWarehouseId = '';
  String _selectedWarehouseName = ''; // Default - will be set in initState
  String? _selectedLocationId;
  String? _selectedLocationName;

  @override
  void initState() {
    super.initState();
    _initializeUserData();
    _selectedWarehouseId = _shopId ?? ''; // ใช้ shopId เป็น warehouse default
    _selectedWarehouseName = global.language("default_warehouse");

    // โหลดตะกร้าทั้งหมดหลังจาก build เสร็จ
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _loadAllCarts();
    });
  }

  void _initializeUserData() {
    _shopId = global.getShopId();
    _userEmail = _getUserEmailWithFallback();
    AppLogger.info('🛒 Cart System - ShopId: $_shopId, Email: $_userEmail');
  }

  String _getUserEmailWithFallback() {
    if (global.userLoginData.email.isNotEmpty) {
      return global.userLoginData.email;
    }
    if (global.loginEmail.isNotEmpty) {
      return global.loginEmail;
    }
    if (global.userLoginData.name.isNotEmpty) {
      return global.userLoginData.name;
    }
    if (global.loginName.isNotEmpty) {
      return global.loginName;
    }
    if (global.userLoginData.code.isNotEmpty) {
      return global.userLoginData.code;
    }
    return 'guest';
  }

  /// โหลดตะกร้าทั้งหมดจาก MongoDB และจัดกลุ่มตามระบบ
  Future<void> _loadAllCarts() async {
    try {
      if (_shopId == null || _userEmail == null) {
        AppLogger.warning('⚠️ ไม่มี shopId หรือ userEmail, ข้าม load carts');
        return;
      }

      AppLogger.info(
        '🔄 กำลังโหลดตะกร้าจาก MongoDB... (shopId: $_shopId, email: $_userEmail)',
      );

      final cartService = MongoDBCartService();

      // โหลดตะกร้าแยกตามระบบ (จะได้เฉพาะ active status)
      final salesCarts = await cartService.getCartsByType(
        _shopId!,
        _userEmail!,
        CartSystemType.sales.toJson(),
      );

      final purchaseCarts = await cartService.getCartsByType(
        _shopId!,
        _userEmail!,
        CartSystemType.purchase.toJson(),
      );

      final inventoryCarts = await cartService.getCartsByType(
        _shopId!,
        _userEmail!,
        CartSystemType.inventory.toJson(),
      );

      AppLogger.info(
        '📦 พบตะกร้า: ขาย ${salesCarts.length}, ซื้อ ${purchaseCarts.length}, คงคลัง ${inventoryCarts.length}',
      );

      // Log ชื่อตะกร้าทั้งหมดเพื่อตรวจสอบ
      if (salesCarts.isNotEmpty) {
        AppLogger.debug(
          '📝 ตะกร้าขาย: ${salesCarts.map((c) => '${c.cartName} (${c.cartType.displayNameThai})').join(', ')}',
        );
      }
      if (purchaseCarts.isNotEmpty) {
        AppLogger.debug(
          '📝 ตะกร้าซื้อ: ${purchaseCarts.map((c) => '${c.cartName} (${c.cartType.displayNameThai})').join(', ')}',
        );
      }
      if (inventoryCarts.isNotEmpty) {
        AppLogger.debug(
          '📝 ตะกร้าคงคลัง: ${inventoryCarts.map((c) => '${c.cartName} (${c.cartType.displayNameThai})').join(', ')}',
        );
      }

      if (mounted) {
        setState(() {
          _salesCarts = salesCarts;
          _purchaseCarts = purchaseCarts;
          _inventoryCarts = inventoryCarts;

          // เลือก active cart (ถ้ามี) ให้ใช้ตัวแรก หรือคงค่าเดิมถ้ายังมีอยู่
          _activeSalesCart =
              _activeSalesCart != null &&
                  salesCarts.any((c) => c.cartId == _activeSalesCart!.cartId)
              ? salesCarts.firstWhere(
                  (c) => c.cartId == _activeSalesCart!.cartId,
                )
              : (salesCarts.isNotEmpty ? salesCarts.first : null);

          _activePurchaseCart =
              _activePurchaseCart != null &&
                  purchaseCarts.any(
                    (c) => c.cartId == _activePurchaseCart!.cartId,
                  )
              ? purchaseCarts.firstWhere(
                  (c) => c.cartId == _activePurchaseCart!.cartId,
                )
              : (purchaseCarts.isNotEmpty ? purchaseCarts.first : null);

          _activeInventoryCart =
              _activeInventoryCart != null &&
                  inventoryCarts.any(
                    (c) => c.cartId == _activeInventoryCart!.cartId,
                  )
              ? inventoryCarts.firstWhere(
                  (c) => c.cartId == _activeInventoryCart!.cartId,
                )
              : (inventoryCarts.isNotEmpty ? inventoryCarts.first : null);
        });

        AppLogger.info('✅ โหลดตะกร้าสำเร็จและเลือก active cart แล้ว');
        if (_activeSalesCart != null) {
          AppLogger.debug(
            '🎯 Active Sales Cart: ${_activeSalesCart!.cartName} (${_activeSalesCart!.cartType.displayNameThai})',
          );
        }
        if (_activePurchaseCart != null) {
          AppLogger.debug(
            '🎯 Active Purchase Cart: ${_activePurchaseCart!.cartName} (${_activePurchaseCart!.cartType.displayNameThai})',
          );
        }
        if (_activeInventoryCart != null) {
          AppLogger.debug(
            '🎯 Active Inventory Cart: ${_activeInventoryCart!.cartName} (${_activeInventoryCart!.cartType.displayNameThai})',
          );
        }
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ โหลดตะกร้าล้มเหลว: $e',
        error: e,
        stackTrace: stackTrace,
      );
    }
  }

  /// รีเฟรชตะกร้าเฉพาะระบบ
  Future<void> _refreshCart(CartModel cart) async {
    try {
      final cartService = MongoDBCartService();
      final updatedCart = await cartService.getCartById(
        cart.shopId,
        cart.email,
        cart.cartId,
      );

      if (updatedCart != null && mounted) {
        setState(() {
          // อัพเดทตะกร้าในลิสต์
          switch (cart.systemType) {
            case CartSystemType.sales:
              final index = _salesCarts.indexWhere(
                (c) => c.cartId == cart.cartId,
              );
              if (index != -1) {
                _salesCarts[index] = updatedCart;
                if (_activeSalesCart?.cartId == cart.cartId) {
                  _activeSalesCart = updatedCart;
                }
              }
              break;
            case CartSystemType.purchase:
              final index = _purchaseCarts.indexWhere(
                (c) => c.cartId == cart.cartId,
              );
              if (index != -1) {
                _purchaseCarts[index] = updatedCart;
                if (_activePurchaseCart?.cartId == cart.cartId) {
                  _activePurchaseCart = updatedCart;
                }
              }
              break;
            case CartSystemType.inventory:
              final index = _inventoryCarts.indexWhere(
                (c) => c.cartId == cart.cartId,
              );
              if (index != -1) {
                _inventoryCarts[index] = updatedCart;
                if (_activeInventoryCart?.cartId == cart.cartId) {
                  _activeInventoryCart = updatedCart;
                }
              }
              break;
          }
        });
      }
    } catch (e) {
      AppLogger.error('❌ รีเฟรชตะกร้าล้มเหลว: $e');
    }
  }

  /// เลือกตะกร้าที่จะใช้งาน
  void _setActiveCart(CartSystemType systemType, CartModel cart) {
    setState(() {
      switch (systemType) {
        case CartSystemType.sales:
          _activeSalesCart = cart;
          break;
        case CartSystemType.purchase:
          _activePurchaseCart = cart;
          break;
        case CartSystemType.inventory:
          _activeInventoryCart = cart;
          break;
      }
    });
    AppLogger.info(
      '✅ เปลี่ยนตะกร้า: ${cart.cartName} [${cart.cartType.displayNameThai}]',
    );
  }

  /// เปิดหน้าดูรายละเอียดสินค้าในตะกร้า
  Future<void> _openCartDetail(CartModel cart) async {
    await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => CartDetailScreen(
          cart: cart,
          onCartUpdated: () {
            // Refresh cart data หลังจากแก้ไข
            _refreshCart(cart);
          },
          onCartDeleted: () {
            // ถ้าลบตะกร้านี้ ให้โหลดใหม่ทั้งหมด
            _loadAllCarts();
          },
        ),
      ),
    );
  }

  /// สร้างตะกร้าใหม่
  Future<void> _createNewCart(CartSystemType systemType) async {
    try {
      // ไม่บังคับให้เลือกลูกหนี้ เจ้าหนี้ หรือคลังสินค้า (เป็นส่วนเสริม)

      // สร้างชื่อตะกร้า
      final cartName = _generateCartName(systemType);

      // สร้าง cart model
      final cart = CartModel(
        shopId: _shopId!,
        email: _userEmail!,
        cartName: cartName,
        systemType: systemType,
        debtorId: _selectedDebtor?.guidFixed,
        debtorCode: _selectedDebtor?.code,
        debtorName: _selectedDebtor?.fullName,
        creditorId: _selectedCreditor?.guidFixed,
        creditorCode: _selectedCreditor?.code,
        creditorName: _selectedCreditor?.fullName,
        warehouseId: _selectedWarehouseId.isNotEmpty
            ? _selectedWarehouseId
            : null,
        warehouseName: _selectedWarehouseId.isNotEmpty
            ? _selectedWarehouseName
            : null,
        locationId: _selectedLocationId,
        locationName: _selectedLocationName,
      );

      // บันทึกลง MongoDB
      final cartService = MongoDBCartService();
      await cartService.createCart(cart);

      // เพิ่มเข้าลิสต์และตั้งเป็น active
      if (mounted) {
        setState(() {
          switch (systemType) {
            case CartSystemType.sales:
              _salesCarts.add(cart);
              _activeSalesCart = cart;
              break;
            case CartSystemType.purchase:
              _purchaseCarts.add(cart);
              _activePurchaseCart = cart;
              break;
            case CartSystemType.inventory:
              _inventoryCarts.add(cart);
              _activeInventoryCart = cart;
              break;
          }
        });
      }

      AppLogger.info(
        '✅ สร้างตะกร้าใหม่สำเร็จ: ${cart.cartName} (ID: ${cart.cartId})',
      );

      if (mounted) {
        global.showSuccessSnackBar(context, '${global.language("create_cart_success")} "${cart.cartName}"');
      }
    } catch (e) {
      AppLogger.error('❌ Error creating cart: $e');
      _showErrorDialog('${global.language("cannot_create_cart")}: $e');
    }
  }

  String _generateCartName(CartSystemType systemType) {
    final timestamp = DateTime.now();
    final dateStr = '${timestamp.day}/${timestamp.month}/${timestamp.year}';
    final timeStr = '${timestamp.hour}:${timestamp.minute}';

    switch (systemType) {
      case CartSystemType.sales:
        return '${global.language("sales")}-${_selectedDebtor?.code ?? 'XXXX'}-$dateStr $timeStr';
      case CartSystemType.purchase:
        return '${global.language("purchase")}-${_selectedCreditor?.code ?? 'XXXX'}-$dateStr $timeStr';
      case CartSystemType.inventory:
        return '${global.language("inventory")}-$dateStr $timeStr';
    }
  }

  void _showErrorDialog(String message) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language("error_occurred")),
        content: Text(message),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language("ok")),
          ),
        ],
      ),
    );
  }

  /// เปิด dialog เลือกลูกหนี้
  Future<void> _selectDebtor() async {
    final result = await showDialog<DebtorModel>(
      context: context,
      builder: (context) => DebtorSelectorDialog(shopId: _shopId!),
    );

    if (result != null && mounted) {
      setState(() {
        _selectedDebtor = result;
      });
      AppLogger.info('✅ เลือกลูกหนี้: ${result.displayName}');
    }
  }

  /// เปิด dialog เลือกเจ้าหนี้
  Future<void> _selectCreditor() async {
    final result = await showDialog<CreditorModel>(
      context: context,
      builder: (context) => CreditorSelectorDialog(shopId: _shopId!),
    );

    if (result != null && mounted) {
      setState(() {
        _selectedCreditor = result;
      });
      AppLogger.info('✅ เลือกเจ้าหนี้: ${result.displayName}');
    }
  }

  /// เปิด dialog เลือกคลังและที่เก็บ
  Future<void> _selectWarehouse() async {
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) => WarehouseSelectorDialog(
        shopId: _shopId!,
        currentWarehouseId: _selectedWarehouseId,
        currentLocationId: _selectedLocationId,
      ),
    );

    if (result != null && mounted) {
      final warehouse = result['warehouse'];
      final location = result['location'];

      setState(() {
        _selectedWarehouseId = warehouse.guidFixed;
        _selectedWarehouseName = warehouse.displayName;
        _selectedLocationId = location?.guidFixed;
        _selectedLocationName = location?.displayName;
      });

      AppLogger.info(
        '✅ เลือกคลัง: $_selectedWarehouseName${location != null ? ', ที่เก็บ: ${location.displayName}' : ''}',
      );
    }
  }

  /// เปลี่ยนระบบที่กำลังแสดง
  void _switchSystemType(CartSystemType newSystemType) {
    setState(() {
      _selectedSystemType = newSystemType;
    });
    AppLogger.info('🔄 เปลี่ยนระบบเป็น: ${newSystemType.displayNameThai}');
  }

  /// สร้างปุ่มเลือกระบบใน AppBar
  Widget _buildSystemButton({
    required CartSystemType systemType,
    required IconData icon,
    required String label,
  }) {
    final isSelected = _selectedSystemType == systemType;

    return Container(
      margin: const EdgeInsets.all(4),
      child: ElevatedButton(
        onPressed: () => _switchSystemType(systemType),
        style: ElevatedButton.styleFrom(
          backgroundColor: isSelected ? global.theme.infoHighlightTextColor : null,
          foregroundColor: isSelected ? global.theme.onPrimaryColor : null,
        ),
        child: Text(label),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocProvider(
      providers: [
        BlocProvider(
          create: (_) => CartCubit(cartService: MongoDBCartService()),
        ),
        BlocProvider(
          create: (_) => ProductSearchCubit(
            searchApi: ProductSearchApi(), // ใช้ API ใหม่ (Backend Unified Search)
            productService: PgSQLProductService(), // Fallback สำหรับ Popular Products
            useUnifiedApi: true,
          ),
        ),
      ],
      child: Scaffold(
        appBar: AppBar(
          title: Text(global.language("cart_system")),
          actions: [
            // ปุ่มเลือกระบบขาย
            _buildSystemButton(
              systemType: CartSystemType.sales,
              icon: Icons.shopping_cart,
              label: global.language("sales"),
            ),
            const SizedBox(width: 8),
            // ปุ่มเลือกระบบซื้อ
            _buildSystemButton(
              systemType: CartSystemType.purchase,
              icon: Icons.add_shopping_cart,
              label: global.language("purchase"),
            ),
            const SizedBox(width: 8),
            // ปุ่มเลือกระบบคงคลัง
            _buildSystemButton(
              systemType: CartSystemType.inventory,
              icon: Icons.inventory_2,
              label: global.language("inventory"),
            ),
            const SizedBox(width: 16),

            // ปุ่ม Toggle ระหว่างค้นหาสินค้า กับ จัดการตะกร้า
            IconButton(
              icon: Icon(
                _showingCartList ? Icons.store : Icons.shopping_cart,
                color: global.theme.onPrimaryColor,
              ),
              onPressed: () =>
                  setState(() => _showingCartList = !_showingCartList),
              tooltip: _showingCartList ? global.language("search_product") : global.language("manage_cart"),
            ),

            // Zoom Controls (สำหรับหน้า Product Search เท่านั้น)
            if (!_showingCartList) ...[
              // ปุ่ม Zoom Out (-)
              IconButton(
                icon: Icon(Icons.remove, color: global.theme.onPrimaryColor),
                onPressed: _cardZoomLevel > 1
                    ? () => setState(() => _cardZoomLevel--)
                    : null,
                tooltip: global.language("reduce_card_size"),
              ),
              // ปุ่ม Reset (=)
              IconButton(
                icon: Icon(Icons.crop_square, color: global.theme.onPrimaryColor),
                onPressed: _cardZoomLevel != 2
                    ? () => setState(() => _cardZoomLevel = 2)
                    : null,
                tooltip: global.language("reset_card_size"),
              ),
              // ปุ่ม Zoom In (+)
              IconButton(
                icon: Icon(Icons.add, color: global.theme.onPrimaryColor),
                onPressed: _cardZoomLevel < 4
                    ? () => setState(() => _cardZoomLevel++)
                    : null,
                tooltip: global.language("enlarge_card_size"),
              ),
            ],
            const SizedBox(width: 8),
          ],
        ),
        body: AnimatedSwitcher(
          duration: const Duration(milliseconds: 300),
          child: _showingCartList
              ? _buildCartListPage()
              : LayoutBuilder(
                  builder: (context, constraints) {
                    // Responsive: ถ้าหน้าจอกว้างมาก (> 900) แสดงแบบ side-by-side
                    final isWideScreen = constraints.maxWidth > 900;

                    // แสดงระบบตามที่เลือก
                    return _buildSystemTab(_selectedSystemType, isWideScreen);
                  },
                ),
        ),
      ),
    );
  }

  Widget _buildSystemTab(CartSystemType systemType, bool isWideScreen) {
    // ดึง active cart ของแต่ละระบบ
    final activeCart = systemType == CartSystemType.sales
        ? _activeSalesCart
        : systemType == CartSystemType.purchase
        ? _activePurchaseCart
        : _activeInventoryCart;

    // ดึงรายการตะกร้าทั้งหมดของระบบนี้
    final allCarts = systemType == CartSystemType.sales
        ? _salesCarts
        : systemType == CartSystemType.purchase
        ? _purchaseCarts
        : _inventoryCarts;

    // แสดง Panel กระทัดรัดอยู่ด้านบนเสมอ พร้อม search box
    return Column(
      children: [
        // บน: ตัวเลือก, search box และปุ่มต่างๆ (กระทัดรัด)
        Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [global.theme.primaryColor, global.theme.primaryLightColor],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.1),
                blurRadius: 4,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: _buildCompactPanel(systemType, activeCart, allCarts),
        ),
        Expanded(child: _buildProductSearchArea(systemType, activeCart)),
      ],
    );
  }

  /// สร้าง Panel กระทัดรัดพร้อม Wrap เลือกตะกร้า (เต็มความกว้าง)
  Widget _buildCompactPanel(
    CartSystemType systemType,
    CartModel? activeCart,
    List<CartModel> allCarts,
  ) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // แถวที่ 1: Wrap แสดงตะกร้าทั้งหมด (ซ้ายสุด, เต็มความกว้าง)
          Wrap(
            spacing: 8,
            runSpacing: 8,
            alignment: WrapAlignment.start,
            children: [],
          ),

          // แถวที่ 2: Chip ตัวเลือกอื่นๆ (ลูกหนี้, เจ้าหนี้, คลัง)
          Wrap(
            spacing: 8,
            runSpacing: 8,
            alignment: WrapAlignment.start,
            children: [
              // เลือกลูกหนี้ (สำหรับระบบขาย)
              if (systemType == CartSystemType.sales) _buildCompactDebtorChip(),

              // เลือกเจ้าหนี้ (สำหรับระบบซื้อ)
              if (systemType == CartSystemType.purchase)
                _buildCompactCreditorChip(),

              // เลือกคลัง
              _buildCompactWarehouseChip(),
              // รายการตะกร้าทั้งหมด
              ...allCarts.map(
                (cart) => _buildCartChip(
                  systemType,
                  cart,
                  isActive: activeCart?.cartId == cart.cartId,
                ),
              ),

              // ปุ่มดูสินค้าในตะกร้า (แสดงเฉพาะเมื่อมี active cart)
              if (activeCart != null)
                ElevatedButton.icon(
                  onPressed: () => _openCartDetail(activeCart),
                  icon: Icon(Icons.list_alt, size: 18),
                  label: Text(
                    '${activeCart.cartName} [${activeCart.cartType.displayNameThai}] (${activeCart.itemCount})',
                    style: TextStyle(fontSize: 13),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: global.theme.infoHighlightTextColor,
                    foregroundColor: global.theme.onPrimaryColor,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 8,
                    ),
                    minimumSize: const Size(0, 36),
                  ),
                ),

              // ปุ่มสร้างตะกร้าใหม่ (กระทัดรัด)
              ElevatedButton.icon(
                onPressed: () => _createNewCart(systemType),
                icon: Icon(Icons.add, size: 18),
                label: Text(
                  global.language("create_cart"),
                  style: TextStyle(fontSize: 13),
                ),
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.positiveHighlightTextColor,
                  foregroundColor: global.theme.onPrimaryColor,
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 8,
                  ),
                  minimumSize: const Size(0, 36),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// เปิด Dialog เลือกตะกร้า (ตาม systemguide.md)
  Future<void> _openCartSelectorDialog(CartSystemType systemType) async {
    // ดึงรายการตะกร้าทั้งหมดของระบบนี้
    final allCarts = systemType == CartSystemType.sales
        ? _salesCarts
        : systemType == CartSystemType.purchase
        ? _purchaseCarts
        : _inventoryCarts;

    // ดึง active cart ปัจจุบัน
    final currentCart = systemType == CartSystemType.sales
        ? _activeSalesCart
        : systemType == CartSystemType.purchase
        ? _activePurchaseCart
        : _activeInventoryCart;

    await showDialog(
      context: context,
      builder: (context) => CartSelectorDialog(
        allCarts: allCarts,
        currentCart: currentCart,
        systemType: systemType,
        onCartSelected: (selectedCart) {
          AppLogger.info(
            '✅ เลือกตะกร้า: ${selectedCart.cartName} [${selectedCart.cartType.displayNameThai}]',
          );
          _setActiveCart(systemType, selectedCart);
        },
        onCartUpdated: (updatedCart) {
          AppLogger.info(
            '🔄 ได้รับการแจ้งเตือนว่าตะกร้าถูกอัพเดท: ${updatedCart.cartName} [${updatedCart.cartType.displayNameThai}]',
          );
          // โหลดตะกร้าใหม่ทั้งหมดเพื่อรับข้อมูลล่าสุด
          _loadAllCarts();
          AppLogger.info('📥 กำลังโหลดตะกร้าทั้งหมดใหม่...');
        },
      ),
    );
  }

  /// Chip สำหรับแต่ละตะกร้า (กดเพื่อเปิด Dialog เลือกตะกร้า)
  Widget _buildCartChip(
    CartSystemType systemType,
    CartModel cart, {
    required bool isActive,
  }) {
    return ActionChip(
      avatar: Icon(
        Icons.shopping_cart,
        size: 18,
        color: isActive ? global.theme.onPrimaryColor : global.theme.primaryColor,
      ),
      label: Text(
        '${cart.cartName} [${cart.cartType.displayNameThai}] (${cart.itemCount})',
        style: TextStyle(
          fontSize: 13,
          color: isActive ? global.theme.onPrimaryColor : global.theme.primaryColor,
          fontWeight: isActive ? FontWeight.bold : FontWeight.normal,
        ),
      ),
      backgroundColor: isActive ? global.theme.primaryColor : global.theme.infoHighlightColor,
      side: BorderSide(color: global.theme.primaryColor, width: isActive ? 2 : 1),
      onPressed: () => _openCartSelectorDialog(systemType),
    );
  }

  /// Chip กระทัดรัดสำหรับเลือกลูกหนี้
  Widget _buildCompactDebtorChip() {
    return ActionChip(
      avatar: Icon(Icons.person, size: 18, color: global.theme.onPrimaryColor),
      label: Text(
        _selectedDebtor?.displayName ?? '${global.language("debtor")}: ${global.language("not_selected")}',
        style: TextStyle(fontSize: 13, color: global.theme.onPrimaryColor),
      ),
      backgroundColor: global.theme.infoHighlightTextColor,
      onPressed: _selectDebtor,
    );
  }

  /// Chip กระทัดรัดสำหรับเลือกเจ้าหนี้
  Widget _buildCompactCreditorChip() {
    return ActionChip(
      avatar: Icon(Icons.business, size: 18, color: global.theme.onPrimaryColor),
      label: Text(
        _selectedCreditor?.displayName ?? '${global.language("creditor")}: ${global.language("not_selected")}',
        style: TextStyle(fontSize: 13, color: global.theme.onPrimaryColor),
      ),
      backgroundColor: global.theme.warningHighlightTextColor,
      onPressed: _selectCreditor,
    );
  }

  /// Chip กระทัดรัดสำหรับเลือกคลัง
  Widget _buildCompactWarehouseChip() {
    final displayText = _selectedLocationName != null
        ? '$_selectedWarehouseName / $_selectedLocationName'
        : '${global.language("warehouse")}: $_selectedWarehouseName';

    return ActionChip(
      avatar: Icon(Icons.warehouse, size: 18, color: global.theme.onPrimaryColor),
      label: Text(
        displayText,
        style: TextStyle(fontSize: 13, color: global.theme.onPrimaryColor),
      ),
      backgroundColor: global.theme.primaryColor,
      onPressed: _selectWarehouse,
    );
  }

  Widget _buildProductSearchArea(
    CartSystemType systemType,
    CartModel? activeCart,
  ) {
    if (activeCart == null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.shopping_cart_outlined,
              size: 80,
              color: global.theme.iconSecondaryColor,
            ),
            SizedBox(height: 16),
            Text(
              global.language("please_create_cart_first"),
              style: TextStyle(fontSize: 16, color: global.theme.iconSecondaryColor),
            ),
          ],
        ),
      );
    }

    return ProductSearchWidget(
      shopId: _shopId!,
      showPopularProducts: true,
      cardZoomLevel: _cardZoomLevel,
      activeCartId: activeCart.cartId,
      onCartUpdated: () => _refreshCart(activeCart),
    );
  }

  /// หน้าจัดการตะกร้า (แสดง CartListWidget)
  Widget _buildCartListPage() {
    // ดึง active cart ของระบบที่เลือก
    final activeCart = _selectedSystemType == CartSystemType.sales
        ? _activeSalesCart
        : _selectedSystemType == CartSystemType.purchase
        ? _activePurchaseCart
        : _activeInventoryCart;

    return Builder(
      builder: (context) {
        // โหลดข้อมูล cart ใหม่เมื่อแสดงหน้านี้
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) {
            context.read<CartCubit>().loadCarts(_shopId!, _userEmail!);
          }
        });

        return CartListWidget(
          shopId: _shopId!,
          email: _userEmail!,
          activeCartId: activeCart?.cartId,
          onCartUpdated: _loadAllCarts, // Callback เมื่อแก้ไขตะกร้า
          onCartTap: (cart) {
            // เลือกตะกร้าเป็น active cart และกลับไปหน้าค้นหาสินค้า
            _setActiveCart(_selectedSystemType, cart);
            setState(() => _showingCartList = false);
            global.showSuccessSnackBar(context, '${global.language("selected_cart")}: ${cart.cartName}');
          },
          onCartDelete: (cart) async {
            final confirmed = await showDialog<bool>(
              context: context,
              builder: (context) => AlertDialog(
                title: Text(global.language("delete_cart")),
                content: Text(
                  '${global.language("confirm_delete_cart")} "${cart.cartName}"?',
                ),
                actions: [
                  TextButton(
                    onPressed: () => Navigator.pop(context, false),
                    child: Text(global.language("cancel")),
                  ),
                  ElevatedButton(
                    onPressed: () => Navigator.pop(context, true),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.negativeHighlightTextColor,
                      foregroundColor: global.theme.onPrimaryColor,
                    ),
                    child: Text(global.language("delete")),
                  ),
                ],
              ),
            );

            if (confirmed == true) {
              try {
                final cartService = MongoDBCartService();
                await cartService.deleteCart(
                  cart.shopId,
                  cart.email,
                  cart.cartId,
                );

                // โหลดตะกร้าใหม่
                await _loadAllCarts();

                if (mounted) {
                  global.showSuccessSnackBar(context, '${global.language("delete_cart_success")} "${cart.cartName}"');
                }
              } catch (e) {
                AppLogger.error('❌ Error deleting cart: $e');
                if (mounted) {
                  global.showErrorSnackBar(context, '${global.language("cannot_delete_cart")}: $e');
                }
              }
            }
          },
        );
      },
    );
  }
}
