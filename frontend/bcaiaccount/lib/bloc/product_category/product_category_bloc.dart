import 'dart:io';
import 'dart:typed_data';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/product_category_repository.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/model/product_category_model.dart';

part 'product_category_event.dart';
part 'product_category_state.dart';

class ProductCategoryBloc
    extends Bloc<ProductCategoryEvent, ProductCategoryState> {
  final ProductCategoryRepository _productCategoryRepository;

  ProductCategoryBloc(
      {required ProductCategoryRepository productCategoryRepository})
      : _productCategoryRepository = productCategoryRepository,
        super(ProductCategoryInitial()) {
    on<ProductCategoryLoadList>(onProductCategoryLoad);
    on<ProductCategorySave>(onProductCategorySave);
    on<ProductCategoryWithImageSave>(onProductCategoryWithImageSave);
    on<ProductCategoryUpdate>(onProductCategoryUpdate);
    on<ProductCategoryWithImageUpdate>(onProductCategoryWithImageUpdate);
    on<ProductCategoryDelete>(productCategoryDelete);
    on<ProductCategoryDeleteMany>(productCategoryDeleteMany);
    on<ProductCategoryGet>(onProductCategoryGet);
    on<ProductCategoryUpdateXOrder>(onProductCategoryUpdateXOrder);
  }

  void onProductCategoryLoad(
      ProductCategoryLoadList event, Emitter<ProductCategoryState> emit) async {
    emit(ProductCategoryInProgress());
    int offset = event.offset ?? 0;
    int limit = event.limit ?? 10000;
    String search = event.search ?? "";
    int groupNumber = event.groupNumber ?? 0;

    try {
      final results = await _productCategoryRepository.getCategoryList(
          offset: offset,
          limit: limit,
          search: search,
          groupNumber: groupNumber);

      if (results.success) {
        List<ProductCategoryModel> productCategorys = (results.data as List)
            .map((category) => ProductCategoryModel.fromJson(category))
            .toList();
        emit(ProductCategoryLoadSuccess(productCategorys: productCategorys));
      } else {
        emit(const ProductCategoryLoadFailed(message: 'Category Not Found'));
      }
    } catch (e) {
      emit(ProductCategoryLoadFailed(message: e.toString()));
    }
  }

  void productCategoryDelete(
      ProductCategoryDelete event, Emitter<ProductCategoryState> emit) async {
    emit(ProductCategoryDeleteInProgress());
    try {
      await _productCategoryRepository.deleteCategory(event.guid);

      emit(ProductCategoryDeleteSuccess());
    } catch (e) {
      // emit(CategoryDeleteFailure(message: e.toString()));
    }
  }

  void productCategoryDeleteMany(ProductCategoryDeleteMany event,
      Emitter<ProductCategoryState> emit) async {
    emit(ProductCategoryDeleteManyInProgress());
    try {
      await _productCategoryRepository.deleteCategoryMany(event.guid);

      emit(ProductCategoryDeleteManySuccess());
    } catch (e) {
      // emit(CategoryDeleteFailure(message: e.toString()));
    }
  }

  void onProductCategoryWithImageSave(ProductCategoryWithImageSave event,
      Emitter<ProductCategoryState> emit) async {
    emit(ProductCategorySaveInProgress());
    try {
      ApiResponse result = await _productCategoryRepository.uploadImage(
          event.imageFile, event.imageWeb!);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        ProductCategoryModel category = event.category;
        category.imageuri = uploadImage.uri;
        await _productCategoryRepository.saveCategory(category);
        emit(ProductCategorySaveSuccess());
      } else {
        emit(ProductCategorySaveFailed(message: result.message));
      }
    } catch (e) {
      emit(ProductCategorySaveFailed(message: e.toString()));
    }
  }

  void onProductCategorySave(
      ProductCategorySave event, Emitter<ProductCategoryState> emit) async {
    emit(ProductCategorySaveInProgress());
    try {
      await _productCategoryRepository.saveCategory(event.category);
      emit(ProductCategorySaveSuccess());
    } catch (e) {
      emit(ProductCategorySaveFailed(message: e.toString()));
    }
  }

  void onProductCategoryUpdateXOrder(ProductCategoryUpdateXOrder event,
      Emitter<ProductCategoryState> emit) async {
    try {
      await _productCategoryRepository.updateCategoryXOrder(event.orderLists);
    } catch (_) {}
  }

  void onProductCategoryUpdate(
      ProductCategoryUpdate event, Emitter<ProductCategoryState> emit) async {
    emit(ProductCategoryUpdateInProgress());
    try {
      await _productCategoryRepository.updateCategory(
          event.guid, event.category);
      emit(ProductCategoryUpdateSuccess());
    } catch (e) {
      emit(ProductCategoryUpdateFailed(message: e.toString()));
    }
  }

  void onProductCategoryWithImageUpdate(ProductCategoryWithImageUpdate event,
      Emitter<ProductCategoryState> emit) async {
    emit(ProductCategoryUpdateInProgress());
    try {
      ApiResponse result = await _productCategoryRepository.uploadImage(
          event.imageFile, event.imageWeb);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        ProductCategoryModel category = ProductCategoryModel(
          guidfixed: event.category.guidfixed,
          parentguid: event.category.parentguid,
          parentguidall: event.category.parentguidall,
          imageuri: uploadImage.uri,
          childcount: event.category.childcount,
          colorselect: event.category.colorselect,
          colorselecthex: event.category.colorselecthex,
          useimageorcolor: event.category.useimageorcolor,
          isdisabled: event.category.isdisabled,
          names: event.category.names,
          xsorts: event.category.xsorts,
          codelist: event.category.codelist,
          timeforsales: event.category.timeforsales,
        );
        await _productCategoryRepository.updateCategory(event.guid, category);
        emit(ProductCategoryUpdateSuccess());
      } else {
        emit(ProductCategoryUpdateFailed(message: result.message));
      }
    } catch (e) {
      emit(ProductCategoryUpdateFailed(message: e.toString()));
    }
  }

  void onProductCategoryGet(
      ProductCategoryGet event, Emitter<ProductCategoryState> emit) async {
    emit(ProductCategoryGetInProgress());
    try {
      final result = await _productCategoryRepository.getCategory(event.guid);
      if (result.success) {
        ProductCategoryModel category =
            ProductCategoryModel.fromJson(result.data);
        emit(ProductCategoryGetSuccess(category: category));
      } else {
        emit(const ProductCategoryGetFailed(message: 'Category Not Found'));
      }
    } catch (e) {
      // emit(CategoryDeleteFailure(message: e.toString()));
    }
  }
}
