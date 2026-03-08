import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/models/shop_model.dart';
import 'package:bclms/app/providers/service_providers.dart';

class ShopSelectionState {
  final List<ShopModel> shops;
  final bool isLoading;
  final bool autoSelected;
  final bool shopSelected;
  final String? errorMessage;

  const ShopSelectionState({
    this.shops = const [],
    this.isLoading = true,
    this.autoSelected = false,
    this.shopSelected = false,
    this.errorMessage,
  });

  ShopSelectionState copyWith({
    List<ShopModel>? shops,
    bool? isLoading,
    bool? autoSelected,
    bool? shopSelected,
    String? errorMessage,
  }) {
    return ShopSelectionState(
      shops: shops ?? this.shops,
      isLoading: isLoading ?? this.isLoading,
      autoSelected: autoSelected ?? this.autoSelected,
      shopSelected: shopSelected ?? this.shopSelected,
      errorMessage: errorMessage,
    );
  }
}

class ShopSelectionNotifier extends Notifier<ShopSelectionState> {
  @override
  ShopSelectionState build() {
    Future.microtask(() => _initialize());
    return const ShopSelectionState();
  }

  Future<void> _initialize() async {
    state = state.copyWith(isLoading: true);
    try {
      final authService = ref.read(authServiceProvider);
      // เรียก verifyToken + listShops พร้อมกัน
      final results = await Future.wait([
        authService.verifyToken(),
        authService.listShops(),
      ]);
      final result = results[1] as List<ShopModel>;
      state = state.copyWith(shops: result, isLoading: false);

      // ถ้ามี shop เดียว → auto select
      if (result.length == 1) {
        AppLogger.info('มีร้านค้าเดียว — เลือกอัตโนมัติ');
        await _selectShop(result.first);
      }
    } catch (e) {
      AppLogger.error('โหลดรายการร้านค้าล้มเหลว', error: e);
      state = state.copyWith(
        isLoading: false,
        errorMessage: 'ไม่สามารถโหลดรายการร้านค้าได้',
      );
    }
  }

  Future<void> onShopSelected(ShopModel shop) async {
    state = state.copyWith(isLoading: true);
    await _selectShop(shop);
    if (!state.shopSelected) {
      state = state.copyWith(isLoading: false);
    }
  }

  Future<void> _selectShop(ShopModel shop) async {
    final authService = ref.read(authServiceProvider);
    final success = await authService.selectShop(shop);
    if (success) {
      state = state.copyWith(shopSelected: true, isLoading: false);
    } else {
      state = state.copyWith(
        isLoading: false,
        errorMessage: 'ไม่สามารถเลือกร้านค้าได้',
      );
    }
  }

  Future<void> logout() async {
    final authService = ref.read(authServiceProvider);
    await authService.logout();
  }
}

final shopSelectionProvider =
    NotifierProvider<ShopSelectionNotifier, ShopSelectionState>(
  ShopSelectionNotifier.new,
);
