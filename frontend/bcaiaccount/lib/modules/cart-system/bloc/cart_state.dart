import 'package:equatable/equatable.dart';
import '../models/cart_model.dart';

/// Base class สำหรับ Cart States
abstract class CartState extends Equatable {
  const CartState();

  @override
  List<Object?> get props => [];
}

/// สถานะเริ่มต้น
class CartInitial extends CartState {}

/// กำลังโหลด
class CartLoading extends CartState {}

/// โหลดสำเร็จ
class CartLoaded extends CartState {
  final List<CartModel> carts;
  final CartModel? selectedCart;

  const CartLoaded({required this.carts, this.selectedCart});

  @override
  List<Object?> get props => [carts, selectedCart];

  // Helper methods
  int get cartCount => carts.length;
  int get activeCartCount => carts.where((c) => c.status == 'active').length;
  List<CartModel> get activeCarts =>
      carts.where((c) => c.status == 'active').toList();

  // Copy with new values
  CartLoaded copyWith({List<CartModel>? carts, CartModel? selectedCart}) {
    return CartLoaded(
      carts: carts ?? this.carts,
      selectedCart: selectedCart ?? this.selectedCart,
    );
  }
}

/// สร้างตะกร้าสำเร็จ
class CartCreated extends CartState {
  final CartModel cart;

  const CartCreated({required this.cart});

  @override
  List<Object?> get props => [cart];
}

/// อัพเดทสำเร็จ
class CartUpdated extends CartState {
  final CartModel cart;
  final String message;

  const CartUpdated({
    required this.cart,
    this.message = 'Cart updated successfully',
  });

  @override
  List<Object?> get props => [cart, message];
}

/// ลบสำเร็จ
class CartDeleted extends CartState {
  final String message;

  const CartDeleted({this.message = 'Cart deleted successfully'});

  @override
  List<Object?> get props => [message];
}

/// เกิดข้อผิดพลาด
class CartError extends CartState {
  final String error;

  const CartError({required this.error});

  @override
  List<Object?> get props => [error];
}

/// กำลังอัพเดท (แสดง loading indicator แบบย่อย)
class CartUpdating extends CartState {
  final List<CartModel> carts;
  final CartModel? selectedCart;

  const CartUpdating({required this.carts, this.selectedCart});

  @override
  List<Object?> get props => [carts, selectedCart];
}
