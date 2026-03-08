import 'dart:convert';

import 'package:smlaicloud/model/product_type_model.dart';
import 'package:smlaicloud/repositories/product_type_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'product_type_event.dart';
part 'product_type_state.dart';

class ProductTypeBloc extends Bloc<ProductTypeEvent, ProductTypeState> {
  final ProductTypeRepository _productTypeRepository;

  ProductTypeBloc({required ProductTypeRepository productTypeRepository})
      : _productTypeRepository = productTypeRepository,
        super(ProductTypeInitial()) {
    on<ProductTypeLoadList>(onProductTypeLoad);
    on<ProductTypeSave>(onProductTypeSave);
    on<ProductTypeUpdate>(onProductTypeUpdate);
    on<ProductTypeDelete>(onProductTypeDelete);
    on<ProductTypeDeleteMany>(onProductTypeDeleteMany);
    on<ProductTypeGet>(onProductTypeGet);
  }

  void onProductTypeLoad(
      ProductTypeLoadList event, Emitter<ProductTypeState> emit) async {
    emit(ProductTypeInProgress());
    try {
      final results = await _productTypeRepository.getProductTypeList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<ProductTypeModel> productType = (results.data as List)
            .map((productType) => ProductTypeModel.fromJson(productType))
            .toList();
        // print(productType.length);
        emit(ProductTypeLoadSuccess(productType: productType));
      } else {
        emit(const ProductTypeLoadFailed(message: 'Product Type Not Found'));
      }
    } catch (e) {
      emit(ProductTypeLoadFailed(message: e.toString()));
    }
  }

  void onProductTypeDelete(
      ProductTypeDelete event, Emitter<ProductTypeState> emit) async {
    emit(ProductTypeDeleteInProgress());
    try {
      await _productTypeRepository.deleteProductType(event.guid);

      emit(ProductTypeDeleteSuccess());
    } catch (e) {
      emit(ProductTypeDeleteFailed(message: e.toString()));
    }
  }

  void onProductTypeDeleteMany(
      ProductTypeDeleteMany event, Emitter<ProductTypeState> emit) async {
    emit(ProductTypeDeleteManyInProgress());
    try {
      await _productTypeRepository.deleteProductTypeMany(event.guid);

      emit(ProductTypeDeleteManySuccess());
    } catch (e) {
      emit(ProductTypeDeleteManyFailed(message: e.toString()));
    }
  }

  void onProductTypeSave(
      ProductTypeSave event, Emitter<ProductTypeState> emit) async {
    emit(ProductTypeSaveInProgress());
    try {
      await _productTypeRepository.saveProductType(event.productTypemodel);
      emit(ProductTypeSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(ProductTypeSaveFailed(message: error['message']));
    }
  }

  void onProductTypeUpdate(
      ProductTypeUpdate event, Emitter<ProductTypeState> emit) async {
    emit(ProductTypeUpdateInProgress());
    try {
      await _productTypeRepository.updateProductType(
          event.guid, event.productTypemodel);
      emit(ProductTypeUpdateSuccess());
    } catch (e) {
      emit(ProductTypeUpdateFailed(message: e.toString()));
    }
  }

  void onProductTypeGet(
      ProductTypeGet event, Emitter<ProductTypeState> emit) async {
    emit(ProductTypeGetInProgress());
    try {
      final result = await _productTypeRepository.getProductType(event.guid);
      if (result.success) {
        ProductTypeModel productTypeModel =
            ProductTypeModel.fromJson(result.data);
        emit(ProductTypeGetSuccess(productType: productTypeModel));
      } else {
        emit(const ProductTypeGetFailed(message: 'ProductType Not Found'));
      }
    } catch (e) {
      emit(ProductTypeGetFailed(message: e.toString()));
    }
  }
}
