import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../utils/logger/app_logger.dart';
import '../models/product_search_model.dart';
import '../models/product_search_response.dart';
import '../services/pgsql_product_service.dart';
import '../services/product_search_api.dart';
import 'product_search_state.dart';

/// Cubit สำหรับค้นหาสินค้า
/// รองรับทั้ง API ใหม่ (ProductSearchApi) และ API เก่า (PgSQLProductService)
class ProductSearchCubit extends Cubit<ProductSearchState> {
  final PgSQLProductService? productService;
  final ProductSearchApi? searchApi;

  /// ใช้ API ใหม่ (Backend Unified Search) หรือไม่
  final bool useUnifiedApi;

  /// Version counter สำหรับ discard stale results
  int _searchVersion = 0;

  ProductSearchCubit({
    this.productService,
    this.searchApi,
    this.useUnifiedApi = true, // Default ใช้ API ใหม่
  }) : super(ProductSearchInitial()) {
    // ตรวจสอบว่ามี service ที่ใช้งานได้
    assert(
      productService != null || searchApi != null,
      'Must provide either productService or searchApi',
    );
  }

  /// ค้นหาสินค้าด้วยชื่อ
  Future<void> searchByName({
    required String shopId,
    required String keyword,
    int limit = 50,
    bool includeBalance = true,
  }) async {
    if (keyword.isEmpty) {
      emit(ProductSearchInitial());
      return;
    }

    // Increment version — ถ้ามี search ใหม่มา ผลเก่าจะถูก discard
    final currentVersion = ++_searchVersion;

    emit(ProductSearchLoading());
    try {
      if (useUnifiedApi && searchApi != null) {
        final swCubit = Stopwatch()..start();

        // ใช้ API ใหม่ (Backend Unified Search)
        final response = await searchApi!.search(
          shopId: shopId,
          keyword: keyword,
          limit: limit,
          includeBalance: includeBalance,
        );
        final apiMs = swCubit.elapsedMilliseconds;

        // Discard stale result — user พิมพ์คำใหม่แล้ว
        if (currentVersion != _searchVersion) return;

        if (response.isSuccess && response.hasProducts) {
          // แปลง ProductItem → ProductSearchModel เพื่อ backward compatibility
          final products = _convertToProductSearchModels(response.products, shopId);

          // สร้าง maps สำหรับ units และ balances จาก API response
          final unitsMap = <String, List<ProductSearchModel>>{};
          final balancesMap = <String, ProductBalanceModel>{};
          final balanceFormattedMap = <String, String>{};

          for (final item in response.products) {
            // สร้าง units list สำหรับ itemcode นี้
            final unitsList = item.units
                .map((unit) => ProductSearchModel(
                      shopId: shopId,
                      itemCode: item.itemCode,
                      barcode: unit.barcode,
                      name0: item.name0,
                      unitCode: unit.unitCode,
                      unitName: unit.unitName,
                      price1: unit.price1,
                      priceRetail: unit.priceRetail,
                      unitStand: unit.unitStand,
                      unitDive: unit.unitDivide,
                    ))
                .toList();
            unitsMap[item.itemCode] = unitsList;

            // สร้าง balance model สำหรับ itemcode นี้ (รองรับ hierarchical warehouse/location)
            if (item.balance != null) {
              // แปลง warehouses จาก API response
              final warehouseModels = item.balance!.warehouses.map((wh) {
                return WarehouseBalanceModel(
                  warehouseCode: wh.warehouseCode,
                  balanceQty: wh.balanceQty,
                  balanceWord: wh.balanceWord,
                  locations: wh.locations.map((loc) {
                    return LocationBalanceModel(
                      locationCode: loc.locationCode,
                      balanceQty: loc.balanceQty,
                      balanceWord: loc.balanceWord,
                    );
                  }).toList(),
                );
              }).toList();

              balancesMap[item.itemCode] = ProductBalanceModel(
                itemCode: item.itemCode,
                totalBalance: item.balance!.total,
                balanceByShop: {shopId: item.balance!.total},
                formattedBalance: item.balance!.formatted,
                warehouses: warehouseModels,
                pendingRecvQty: item.balance!.pendingRecvQty,
                pendingRecvWord: item.balance!.pendingRecvWord,
                pendingSendQty: item.balance!.pendingSendQty,
                pendingSendWord: item.balance!.pendingSendWord,
              );
              balanceFormattedMap[item.itemCode] = item.balance!.formatted;
            }
          }

          final processMs = swCubit.elapsedMilliseconds - apiMs;
          AppLogger.info(
            '⏱️ [Cubit] API=${apiMs}ms, dataProcess=${processMs}ms, total=${swCubit.elapsedMilliseconds}ms (${response.products.length} items)',
          );

          emit(ProductSearchLoaded(
            products: products,
            searchKeyword: keyword,
            unitsMap: unitsMap,
            balancesMap: balancesMap,
            balanceFormattedMap: balanceFormattedMap,
            total: response.total,
            hasMore: response.hasMore,
            currentOffset: 0,
          ));
        } else {
          emit(ProductNotFound(searchKeyword: keyword));
        }
      } else if (productService != null) {
        // ใช้ API เก่า (PostgreSQL direct)
        final result = await productService!.universalSearch(keyword, shopId);

        // Discard stale result
        if (currentVersion != _searchVersion) return;

        if (result.isSuccess && result.hasProducts) {
          emit(ProductSearchLoaded(products: result.products, searchKeyword: keyword));
        } else {
          emit(ProductNotFound(searchKeyword: keyword));
        }
      } else {
        emit(ProductSearchError(error: 'No search service available'));
      }
    } catch (e) {
      emit(ProductSearchError(error: e.toString()));
    }
  }

  /// ค้นหาแบบ Universal
  Future<void> universalSearch({
    required String shopId,
    required String keyword,
    int limit = 50,
    bool includeBalance = true,
  }) async {
    // เรียก searchByName เพราะใช้ logic เดียวกัน
    await searchByName(
      shopId: shopId,
      keyword: keyword,
      limit: limit,
      includeBalance: includeBalance,
    );
  }

  /// โหลดสินค้าเพิ่ม (infinite scroll)
  Future<void> loadMore({
    required String shopId,
    int limit = 200,
    bool includeBalance = true,
  }) async {
    // ตรวจสอบว่า state ปัจจุบันเป็น ProductSearchLoaded และยังมีข้อมูลเพิ่ม
    final currentState = state;
    if (currentState is! ProductSearchLoaded || !currentState.hasMore) {
      return;
    }

    // ป้องกันการเรียกซ้ำขณะกำลังโหลด
    if (state is ProductSearchLoadingMore) {
      return;
    }

    final keyword = currentState.searchKeyword;
    // Fix: ใช้จำนวน products ที่โหลดแล้วเป็น offset (แทนที่จะบวกซ้ำ)
    final newOffset = currentState.products.length;

    // emit loading more state (เก็บ products เดิมไว้แสดง)
    emit(ProductSearchLoadingMore(
      currentProducts: currentState.products,
      searchKeyword: keyword,
      unitsMap: currentState.unitsMap,
      balancesMap: currentState.balancesMap,
      balanceFormattedMap: currentState.balanceFormattedMap,
      total: currentState.total,
      currentOffset: newOffset,
    ));

    try {
      if (useUnifiedApi && searchApi != null) {
        final response = await searchApi!.search(
          shopId: shopId,
          keyword: keyword,
          limit: limit,
          offset: newOffset,
          includeBalance: includeBalance,
        );

        if (response.isSuccess && response.hasProducts) {
          // แปลง ProductItem → ProductSearchModel
          final newProducts = _convertToProductSearchModels(response.products, shopId);

          // Merge units, balances, และ balanceFormatted maps
          final mergedUnitsMap = Map<String, List<ProductSearchModel>>.from(currentState.unitsMap ?? {});
          final mergedBalancesMap = Map<String, ProductBalanceModel>.from(currentState.balancesMap ?? {});
          final mergedBalanceFormattedMap = Map<String, String>.from(currentState.balanceFormattedMap ?? {});

          for (final item in response.products) {
            final unitsList = item.units
                .map((unit) => ProductSearchModel(
                      shopId: shopId,
                      itemCode: item.itemCode,
                      barcode: unit.barcode,
                      name0: item.name0,
                      unitCode: unit.unitCode,
                      unitName: unit.unitName,
                      price1: unit.price1,
                      priceRetail: unit.priceRetail,
                      unitStand: unit.unitStand,
                      unitDive: unit.unitDivide,
                    ))
                .toList();
            mergedUnitsMap[item.itemCode] = unitsList;

            if (item.balance != null) {
              final warehouseModels = item.balance!.warehouses.map((wh) {
                return WarehouseBalanceModel(
                  warehouseCode: wh.warehouseCode,
                  balanceQty: wh.balanceQty,
                  balanceWord: wh.balanceWord,
                  locations: wh.locations.map((loc) {
                    return LocationBalanceModel(
                      locationCode: loc.locationCode,
                      balanceQty: loc.balanceQty,
                      balanceWord: loc.balanceWord,
                    );
                  }).toList(),
                );
              }).toList();

              mergedBalancesMap[item.itemCode] = ProductBalanceModel(
                itemCode: item.itemCode,
                totalBalance: item.balance!.total,
                balanceByShop: {shopId: item.balance!.total},
                formattedBalance: item.balance!.formatted,
                warehouses: warehouseModels,
                pendingRecvQty: item.balance!.pendingRecvQty,
                pendingRecvWord: item.balance!.pendingRecvWord,
                pendingSendQty: item.balance!.pendingSendQty,
                pendingSendWord: item.balance!.pendingSendWord,
              );
              mergedBalanceFormattedMap[item.itemCode] = item.balance!.formatted;
            }
          }

          // รวม products เดิมกับใหม่
          final allProducts = [...currentState.products, ...newProducts];

          emit(ProductSearchLoaded(
            products: allProducts,
            searchKeyword: keyword,
            unitsMap: mergedUnitsMap,
            balancesMap: mergedBalancesMap,
            balanceFormattedMap: mergedBalanceFormattedMap,
            total: response.total,
            hasMore: response.hasMore,
            currentOffset: newOffset,
          ));
        } else {
          // ไม่มีข้อมูลเพิ่ม กลับไป state เดิมแต่ hasMore = false
          emit(currentState.copyWith(hasMore: false));
        }
      } else {
        // API เก่าไม่รองรับ pagination
        emit(currentState.copyWith(hasMore: false));
      }
    } catch (e) {
      // เกิด error กลับไป state เดิม
      emit(currentState);
    }
  }

  /// ดึงสินค้ายอดนิยม
  Future<void> loadPopularProducts({
    required String shopId,
    int limit = 200,
  }) async {
    emit(ProductSearchLoading());
    try {
      if (productService != null) {
        // Popular products ใช้ API เก่าเท่านั้น (ยังไม่มีใน unified API)
        final result = await productService!.getPopularProducts(shopId, limit: limit);

        if (result.isSuccess && result.hasProducts) {
          emit(ProductSearchLoaded(products: result.products, searchKeyword: 'Popular Products'));
        } else {
          emit(ProductNotFound(searchKeyword: 'Popular Products'));
        }
      } else {
        emit(ProductSearchError(error: 'Popular products not available with unified API'));
      }
    } catch (e) {
      emit(ProductSearchError(error: e.toString()));
    }
  }

  /// Clear search results
  void clearSearch() {
    emit(ProductSearchInitial());
  }

  /// Reset to initial state
  void reset() {
    emit(ProductSearchInitial());
  }

  /// แปลง ProductItem (API ใหม่) → ProductSearchModel (เดิม)
  /// เพื่อ backward compatibility กับ widget
  List<ProductSearchModel> _convertToProductSearchModels(
    List<ProductItem> items,
    String shopId,
  ) {
    final List<ProductSearchModel> result = [];

    for (final item in items) {
      // แต่ละ unit ของ item จะเป็น ProductSearchModel แยก
      for (final unit in item.units) {
        result.add(ProductSearchModel(
          shopId: shopId,
          itemCode: item.itemCode,
          barcode: unit.barcode,
          name0: item.name0,
          unitCode: unit.unitCode,
          unitName: unit.unitName,
          price1: unit.price1,
          priceRetail: unit.priceRetail,
          unitStand: unit.unitStand,
          unitDive: unit.unitDivide,
          // Balance: ใช้ formatted string จาก backend (ถ้ามี)
          // Note: ProductSearchModel ไม่มี field สำหรับ balance formatted
          // จะต้องแก้ไข widget ให้แสดง balance แยกต่างหาก
        ));
      }

      // ถ้าไม่มี units ให้สร้าง default entry
      if (item.units.isEmpty) {
        result.add(ProductSearchModel(
          shopId: shopId,
          itemCode: item.itemCode,
          barcode: item.itemCode, // ใช้ itemCode เป็น barcode
          name0: item.name0,
          unitCode: '',
          unitName: '',
          price1: 0,
          priceRetail: 0,
          unitStand: 1,
          unitDive: 1,
        ));
      }
    }

    return result;
  }
}
