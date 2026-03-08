import 'package:flutter_bloc/flutter_bloc.dart';
import '../models/cart_model.dart';
import '../models/cart_item_model.dart';
import '../models/cart_system_type.dart';
import '../services/mongodb_cart_service.dart';
import 'cart_state.dart';

/// Cubit สำหรับจัดการ Cart
class CartCubit extends Cubit<CartState> {
  final MongoDBCartService cartService;

  CartCubit({required this.cartService}) : super(CartInitial());

  /// โหลดตะกร้าทั้งหมด
  Future<void> loadCarts(String shopId, String email) async {
    emit(CartLoading());
    try {
      final carts = await cartService.getCarts(shopId, email);
      emit(CartLoaded(carts: carts));
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// โหลดตะกร้าเฉพาะ ID
  Future<void> loadCartById(String shopId, String email, String cartId) async {
    emit(CartLoading());
    try {
      final cart = await cartService.getCartById(shopId, email, cartId);

      if (cart != null) {
        emit(CartLoaded(carts: [cart], selectedCart: cart));
      } else {
        emit(const CartError(error: 'Cart not found'));
      }
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// สร้างตะกร้าใหม่ (รองรับ CartModel ใหม่)
  Future<void> createCart({
    required String shopId,
    required String email,
    required String cartName,
    required CartSystemType systemType,
    required String warehouseId,
    required String warehouseName,
    String? debtorId,
    String? debtorCode,
    String? debtorName,
    String? creditorId,
    String? creditorCode,
    String? creditorName,
    String? locationId,
    String? locationName,
  }) async {
    emit(CartLoading());
    try {
      final newCart = CartModel(
        shopId: shopId,
        email: email,
        cartName: cartName,
        systemType: systemType,
        warehouseId: warehouseId,
        warehouseName: warehouseName,
        debtorId: debtorId,
        debtorCode: debtorCode,
        debtorName: debtorName,
        creditorId: creditorId,
        creditorCode: creditorCode,
        creditorName: creditorName,
        locationId: locationId,
        locationName: locationName,
      );

      final createdCart = await cartService.createCart(newCart);
      emit(CartCreated(cart: createdCart));

      // โหลดตะกร้าทั้งหมดใหม่
      await loadCarts(shopId, email);
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// เพิ่มสินค้าเข้าตะกร้า
  Future<void> addItemToCart({
    required CartModel cart,
    required String itemCode,
    required String barcode,
    required String productName,
    required String unitCode,
    required String unitName,
    required double quantity,
    required double price,
    int unitStand = 1,
    int unitDive = 1,
  }) async {
    if (state is CartLoaded) {
      final currentState = state as CartLoaded;
      emit(
        CartUpdating(
          carts: currentState.carts,
          selectedCart: currentState.selectedCart,
        ),
      );
    }

    try {
      final newItem = CartItemModel(
        itemCode: itemCode,
        barcode: barcode,
        productName: productName,
        unitCode: unitCode,
        unitName: unitName,
        quantity: quantity,
        price: price,
        unitStand: unitStand,
        unitDive: unitDive,
      );

      final updatedCart = cart.addItem(newItem);
      final savedCart = await cartService.updateCart(updatedCart);

      emit(CartUpdated(cart: savedCart, message: 'Item added successfully'));

      // โหลดตะกร้าทั้งหมดใหม่
      await loadCarts(cart.shopId, cart.email);
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// อัพเดทจำนวนสินค้า
  Future<void> updateItemQuantity({
    required CartModel cart,
    required String itemId,
    required double newQuantity,
  }) async {
    if (state is CartLoaded) {
      final currentState = state as CartLoaded;
      emit(
        CartUpdating(
          carts: currentState.carts,
          selectedCart: currentState.selectedCart,
        ),
      );
    }

    try {
      final updatedCart = cart.updateItemQuantity(itemId, newQuantity);
      final savedCart = await cartService.updateCart(updatedCart);

      emit(CartUpdated(cart: savedCart, message: 'Quantity updated'));

      // โหลดตะกร้าทั้งหมดใหม่
      await loadCarts(cart.shopId, cart.email);
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// ลบสินค้าออกจากตะกร้า
  Future<void> removeItemFromCart({
    required CartModel cart,
    required String itemId,
  }) async {
    if (state is CartLoaded) {
      final currentState = state as CartLoaded;
      emit(
        CartUpdating(
          carts: currentState.carts,
          selectedCart: currentState.selectedCart,
        ),
      );
    }

    try {
      final updatedCart = cart.removeItem(itemId);
      final savedCart = await cartService.updateCart(updatedCart);

      emit(CartUpdated(cart: savedCart, message: 'Item removed'));

      // โหลดตะกร้าทั้งหมดใหม่
      await loadCarts(cart.shopId, cart.email);
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// อัพเดทชื่อตะกร้า
  Future<void> updateCartName({
    required CartModel cart,
    required String newName,
  }) async {
    if (state is CartLoaded) {
      final currentState = state as CartLoaded;
      emit(
        CartUpdating(
          carts: currentState.carts,
          selectedCart: currentState.selectedCart,
        ),
      );
    }

    try {
      final updatedCart = cart.updateName(newName);
      final savedCart = await cartService.updateCart(updatedCart);

      emit(CartUpdated(cart: savedCart, message: 'Cart name updated'));

      // โหลดตะกร้าทั้งหมดใหม่
      await loadCarts(cart.shopId, cart.email);
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// เปลี่ยนสถานะตะกร้า
  Future<void> updateCartStatus({
    required CartModel cart,
    required String newStatus,
  }) async {
    if (state is CartLoaded) {
      final currentState = state as CartLoaded;
      emit(
        CartUpdating(
          carts: currentState.carts,
          selectedCart: currentState.selectedCart,
        ),
      );
    }

    try {
      final updatedCart = cart.updateStatus(newStatus);
      final savedCart = await cartService.updateCart(updatedCart);

      emit(
        CartUpdated(cart: savedCart, message: 'Status updated to $newStatus'),
      );

      // โหลดตะกร้าทั้งหมดใหม่
      await loadCarts(cart.shopId, cart.email);
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// ลบตะกร้า
  Future<void> deleteCart({
    required String shopId,
    required String email,
    required String cartId,
  }) async {
    if (state is CartLoaded) {
      final currentState = state as CartLoaded;
      emit(
        CartUpdating(
          carts: currentState.carts,
          selectedCart: currentState.selectedCart,
        ),
      );
    }

    try {
      await cartService.deleteCart(shopId, email, cartId);
      emit(const CartDeleted());

      // โหลดตะกร้าทั้งหมดใหม่
      await loadCarts(shopId, email);
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// โหลดตะกร้าตามประเภท
  Future<void> loadCartsByType({
    required String shopId,
    required String email,
    required String cartType,
  }) async {
    emit(CartLoading());
    try {
      final carts = await cartService.getCartsByType(shopId, email, cartType);
      emit(CartLoaded(carts: carts));
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// Refresh ตะกร้า
  Future<void> refreshCarts(String shopId, String email) async {
    try {
      final carts = await cartService.getCarts(shopId, email);
      emit(CartLoaded(carts: carts));
    } catch (e) {
      emit(CartError(error: e.toString()));
    }
  }

  /// Select cart
  void selectCart(CartModel cart) {
    if (state is CartLoaded) {
      final currentState = state as CartLoaded;
      emit(currentState.copyWith(selectedCart: cart));
    }
  }

  /// Clear selection
  void clearSelection() {
    if (state is CartLoaded) {
      final currentState = state as CartLoaded;
      emit(currentState.copyWith(selectedCart: null));
    }
  }
}
