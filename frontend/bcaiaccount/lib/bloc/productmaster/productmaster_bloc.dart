import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/repositories/product_master_repository.dart';

part 'productmaster_event.dart';
part 'productmaster_state.dart';

class ProductMasterBloc extends Bloc<ProductMasterEvent, ProductMasterState> {
  final ProductMasterRepository _productMasterRepository;

  ProductMasterBloc({required ProductMasterRepository productMasterRepository})
      : _productMasterRepository = productMasterRepository,
        super(ProductMasterInitial()) {
    on<ProductMasterLoadList>(onProductMasterLoad);
    on<ProductMasterSave>(onProductMasterSave);
    on<ProductMasterUpdate>(onProductMasterUpdate);
    on<ProductMasterDelete>(productMasterDelete);
    on<ProductMasterDeleteMany>(productMasterDeleteMany);
    on<ProductMasterGet>(onProductMasterGet);
    on<ProductMasterSaveBulk>(onProductMasterSaveBulk);
  }

  void onProductMasterLoad(
      ProductMasterLoadList event, Emitter<ProductMasterState> emit) async {
    emit(ProductMasterInProgress());

    try {
      final results = await _productMasterRepository.getProductList(
          page: event.page, limit: event.limit, search: event.search);

      if (results.success) {
        List<ProductMasterModel> productMasters = (results.data as List)
            .map((productMaster) => ProductMasterModel.fromJson(productMaster))
            .toList();
        // print(productMasters.length);
        emit(ProductMasterLoadSuccess(productMasters: productMasters));
      } else {
        emit(const ProductMasterLoadFailed(message: 'ProductMaster Not Found'));
      }
    } catch (e) {
      emit(ProductMasterLoadFailed(message: e.toString()));
    }
  }

  void productMasterDelete(
      ProductMasterDelete event, Emitter<ProductMasterState> emit) async {
    emit(ProductMasterDeleteInProgress());
    try {
      await _productMasterRepository.deleteProduct(event.guid);

      emit(ProductMasterDeleteSuccess());
    } catch (e) {
      // emit(ProductMasterDeleteFailure(message: e.toString()));
    }
  }

  void productMasterDeleteMany(
      ProductMasterDeleteMany event, Emitter<ProductMasterState> emit) async {
    emit(ProductMasterDeleteManyInProgress());
    try {
      await _productMasterRepository.deleteProductMany(event.guid);

      emit(ProductMasterDeleteManySuccess());
    } catch (e) {
      // emit(ProductMasterDeleteFailure(message: e.toString()));
    }
  }

  void onProductMasterSave(
      ProductMasterSave event, Emitter<ProductMasterState> emit) async {
    emit(ProductMasterSaveInProgress());
    try {
      await _productMasterRepository.saveProduct(event.productMasterModel);
      emit(ProductMasterSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(ProductMasterSaveFailed(message: error['message']));
    }
  }

  void onProductMasterSaveBulk(
      ProductMasterSaveBulk event, Emitter<ProductMasterState> emit) async {
    emit(ProductMasterSaveInProgress());
    try {
      await _productMasterRepository.saveProductBulk(event.productMasters);
      emit(ProductMasterSaveSuccess());
    } catch (e) {
      emit(ProductMasterSaveFailed(message: e.toString()));
    }
  }

  void onProductMasterUpdate(
      ProductMasterUpdate event, Emitter<ProductMasterState> emit) async {
    emit(ProductMasterUpdateInProgress());
    try {
      await _productMasterRepository.updateProduct(
          event.guid, event.productMasterModel);
      emit(ProductMasterUpdateSuccess());
    } catch (e) {
      emit(ProductMasterUpdateFailed(message: e.toString()));
    }
  }

  void onProductMasterGet(
      ProductMasterGet event, Emitter<ProductMasterState> emit) async {
    emit(ProductMasterGetInProgress());
    try {
      final result = await _productMasterRepository.getProduct(event.guid);
      if (result.success) {
        ProductMasterModel productMaster =
            ProductMasterModel.fromJson(result.data);
        emit(ProductMasterGetSuccess(productMaster: productMaster));
      } else {
        emit(const ProductMasterGetFailed(message: 'ProductMaster Not Found'));
      }
    } catch (e) {
      // emit(ProductMasterDeleteFailure(message: e.toString()));
    }
  }
}
