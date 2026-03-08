import 'dart:io';
import 'dart:typed_data';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/repositories/product_group_repository.dart';
import 'package:smlaicloud/model/product_group_model.dart';

part 'product_group_event.dart';
part 'product_group_state.dart';

class ProductGroupBloc extends Bloc<ProductGroupEvent, ProductGroupState> {
  final ProductGroupRepository _productGroupRepository;

  ProductGroupBloc({required ProductGroupRepository productGroupRepository})
      : _productGroupRepository = productGroupRepository,
        super(ProductGroupInitial()) {
    on<ProductGroupLoadList>(onProductGroupLoad);
    on<ProductGroupSave>(onProductGroupSave);
    on<ProductGroupUpdate>(onProductGroupUpdate);
    on<ProductGroupDelete>(productGroupDelete);
    on<ProductGroupDeleteMany>(productGroupDeleteMany);
    on<ProductGroupGet>(onProductGroupGet);
    on<ProductGroupUpdateXOrder>(onProductGroupUpdateXOrder);
  }

  void onProductGroupLoad(
      ProductGroupLoadList event, Emitter<ProductGroupState> emit) async {
    emit(ProductGroupInProgress());

    try {
      final results = await _productGroupRepository.getProductGroupList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<ProductGroupModel> productGroups = (results.data as List)
            .map((productGroup) => ProductGroupModel.fromJson(productGroup))
            .toList();
        emit(ProductGroupLoadSuccess(productGroups: productGroups));
      } else {
        emit(const ProductGroupLoadFailed(message: 'ProductGroup Not Found'));
      }
    } catch (e) {
      emit(ProductGroupLoadFailed(message: e.toString()));
    }
  }

  void productGroupDelete(
      ProductGroupDelete event, Emitter<ProductGroupState> emit) async {
    emit(ProductGroupDeleteInProgress());
    try {
      await _productGroupRepository.deleteProductGroup(event.guid);

      emit(ProductGroupDeleteSuccess());
    } catch (e) {
      // emit(ProductGroupDeleteFailure(message: e.toString()));
    }
  }

  void productGroupDeleteMany(
      ProductGroupDeleteMany event, Emitter<ProductGroupState> emit) async {
    emit(ProductGroupDeleteManyInProgress());
    try {
      await _productGroupRepository.deleteProductGroupMany(event.guid);

      emit(ProductGroupDeleteManySuccess());
    } catch (e) {
      // emit(ProductGroupDeleteFailure(message: e.toString()));
    }
  }

  void onProductGroupSave(
      ProductGroupSave event, Emitter<ProductGroupState> emit) async {
    emit(ProductGroupSaveInProgress());
    try {
      await _productGroupRepository.saveProductGroup(event.productGroup);
      emit(ProductGroupSaveSuccess());
    } catch (e) {
      emit(ProductGroupSaveFailed(message: e.toString()));
    }
  }

  void onProductGroupUpdateXOrder(
      ProductGroupUpdateXOrder event, Emitter<ProductGroupState> emit) async {
    try {
      await _productGroupRepository.updateProductGroupXOrder(event.orderLists);
    } catch (_) {}
  }

  void onProductGroupUpdate(
      ProductGroupUpdate event, Emitter<ProductGroupState> emit) async {
    emit(ProductGroupUpdateInProgress());
    try {
      await _productGroupRepository.updateProductGroup(
          event.guid, event.productGroup);
      emit(ProductGroupUpdateSuccess());
    } catch (e) {
      emit(ProductGroupUpdateFailed(message: e.toString()));
    }
  }

  void onProductGroupGet(
      ProductGroupGet event, Emitter<ProductGroupState> emit) async {
    emit(ProductGroupGetInProgress());
    try {
      final result = await _productGroupRepository.getProductGroup(event.guid);
      if (result.success) {
        ProductGroupModel productGroup =
            ProductGroupModel.fromJson(result.data);
        emit(ProductGroupGetSuccess(productGroup: productGroup));
      } else {
        emit(const ProductGroupGetFailed(message: 'Product Group Not Found'));
      }
    } catch (e) {
      // emit(ProductGroupDeleteFailure(message: e.toString()));
    }
  }
}
