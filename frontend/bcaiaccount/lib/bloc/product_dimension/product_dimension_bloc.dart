import 'dart:convert';

import 'package:smlaicloud/model/dimension_model.dart';
import 'package:smlaicloud/repositories/product_dimension_reporsitory.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'product_dimension_event.dart';
part 'product_dimension_state.dart';

class ProductDimensionBloc
    extends Bloc<ProductDimensionEvent, ProductDimensionState> {
  final ProductDimensionRepository _productDimensionRepository;

  ProductDimensionBloc(
      {required ProductDimensionRepository productDimensionRepository})
      : _productDimensionRepository = productDimensionRepository,
        super(ProductDimensionInitial()) {
    on<ProductDimensionLoadList>(onProductDimensionLoad);
    on<ProductDimensionSave>(onProductDimensionSave);
    on<ProductDimensionUpdate>(onProductDimensionUpdate);
    on<ProductDimensionDelete>(onProductDimensionDelete);
    on<ProductDimensionDeleteMany>(onProductDimensionDeleteMany);
    on<ProductDimensionGet>(onProductDimensionGet);
  }

  void onProductDimensionLoad(ProductDimensionLoadList event,
      Emitter<ProductDimensionState> emit) async {
    emit(ProductDimensionInProgress());
    try {
      final results = await _productDimensionRepository.getProductDimensionList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<DimensionModel> productDimension = (results.data as List)
            .map(
                (productDimension) => DimensionModel.fromJson(productDimension))
            .toList();
        // print(productDimension.length);
        emit(ProductDimensionLoadSuccess(productDimension: productDimension));
      } else {
        emit(const ProductDimensionLoadFailed(
            message: 'Product Dimension Not Found'));
      }
    } catch (e) {
      emit(ProductDimensionLoadFailed(message: e.toString()));
    }
  }

  void onProductDimensionDelete(
      ProductDimensionDelete event, Emitter<ProductDimensionState> emit) async {
    emit(ProductDimensionDeleteInProgress());
    try {
      await _productDimensionRepository.deleteProductDimension(event.guid);

      emit(ProductDimensionDeleteSuccess());
    } catch (e) {
      emit(ProductDimensionDeleteFailed(message: e.toString()));
    }
  }

  void onProductDimensionDeleteMany(ProductDimensionDeleteMany event,
      Emitter<ProductDimensionState> emit) async {
    emit(ProductDimensionDeleteManyInProgress());
    try {
      await _productDimensionRepository.deleteProductDimensionMany(event.guid);

      emit(ProductDimensionDeleteManySuccess());
    } catch (e) {
      emit(ProductDimensionDeleteManyFailed(message: e.toString()));
    }
  }

  void onProductDimensionSave(
      ProductDimensionSave event, Emitter<ProductDimensionState> emit) async {
    emit(ProductDimensionSaveInProgress());
    try {
      await _productDimensionRepository
          .saveProductDimension(event.productDimensionmodel);
      emit(ProductDimensionSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(ProductDimensionSaveFailed(message: error['message']));
    }
  }

  void onProductDimensionUpdate(
      ProductDimensionUpdate event, Emitter<ProductDimensionState> emit) async {
    emit(ProductDimensionUpdateInProgress());
    try {
      await _productDimensionRepository.updateProductDimension(
          event.guid, event.productDimensionmodel);
      emit(ProductDimensionUpdateSuccess());
    } catch (e) {
      emit(ProductDimensionUpdateFailed(message: e.toString()));
    }
  }

  void onProductDimensionGet(
      ProductDimensionGet event, Emitter<ProductDimensionState> emit) async {
    emit(ProductDimensionGetInProgress());
    try {
      final result =
          await _productDimensionRepository.getProductDimension(event.guid);
      if (result.success) {
        DimensionModel productDimensionModel =
            DimensionModel.fromJson(result.data);
        emit(ProductDimensionGetSuccess(
            productDimension: productDimensionModel));
      } else {
        emit(const ProductDimensionGetFailed(
            message: 'ProductDimension Not Found'));
      }
    } catch (e) {
      emit(ProductDimensionGetFailed(message: e.toString()));
    }
  }
}
