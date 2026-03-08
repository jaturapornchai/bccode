import 'package:smlaicloud/model/shelf_product_model.dart';
import 'package:smlaicloud/repositories/shelf_product_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'shelf_product_event.dart';
part 'shelf_product_state.dart';

class ShelfProductBloc extends Bloc<ShelfProductEvent, ShelfProductState> {
  final ShelfProductRepository _shelfProductRepository;

  ShelfProductBloc({required ShelfProductRepository shelfProductRepository})
      : _shelfProductRepository = shelfProductRepository,
        super(ShelfProductInitial()) {
    on<AddProductsToShelf>(onAddProductsToShelf);
    on<RemoveProductsFromShelf>(onRemoveProductsFromShelf);
    on<LoadProductsInShelf>(onLoadProductsInShelf);
  }

  /// จัดการการเพิ่มสินค้าเข้าชั้นวาง
  void onAddProductsToShelf(
      AddProductsToShelf event, Emitter<ShelfProductState> emit) async {
    emit(ShelfProductAddInProgress());

    try {
      final result = await _shelfProductRepository.addProductsToShelf(
        warehouseCode: event.warehouseCode,
        locationCode: event.locationCode,
        shelfCode: event.shelfCode,
        products: event.products,
      );

      if (result.success) {
        emit(const ShelfProductAddSuccess());
      } else {
        emit(ShelfProductAddFailed(message: result.message));
      }
    } catch (e) {
      emit(ShelfProductAddFailed(message: e.toString()));
    }
  }

  /// จัดการการลบสินค้าออกจากชั้นวาง
  void onRemoveProductsFromShelf(
      RemoveProductsFromShelf event, Emitter<ShelfProductState> emit) async {
    emit(ShelfProductRemoveInProgress());

    try {
      final result = await _shelfProductRepository.removeProductsFromShelf(
        warehouseCode: event.warehouseCode,
        locationCode: event.locationCode,
        shelfCode: event.shelfCode,
        productGuids: event.productGuids,
      );

      if (result.success) {
        emit(const ShelfProductRemoveSuccess());
      } else {
        emit(ShelfProductRemoveFailed(message: result.message));
      }
    } catch (e) {
      emit(ShelfProductRemoveFailed(message: e.toString()));
    }
  }

  /// จัดการการดึงรายการสินค้าที่มีอยู่ในชั้นวาง
  void onLoadProductsInShelf(
      LoadProductsInShelf event, Emitter<ShelfProductState> emit) async {
    emit(ShelfProductLoadInProgress());

    try {
      final result = await _shelfProductRepository.getProductsInShelf(
        warehouseCode: event.warehouseCode,
        locationCode: event.locationCode,
        shelfCode: event.shelfCode,
        limit: event.limit,
        offset: event.offset,
        search: event.search,
      );

      if (result.success) {
        List<ShelfProductDisplayModel> products = (result.data as List)
            .map((product) => ShelfProductDisplayModel.fromJson(product))
            .toList();
        emit(ShelfProductLoadSuccess(products: products));
      } else {
        emit(const ShelfProductLoadFailed(message: 'Products not found'));
      }
    } catch (e) {
      emit(ShelfProductLoadFailed(message: e.toString()));
    }
  }
}
