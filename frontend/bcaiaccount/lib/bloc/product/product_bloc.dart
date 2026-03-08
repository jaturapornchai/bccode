import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';

import 'package:smlaicloud/model/product_branch_model.dart';
import 'package:smlaicloud/model/product_department_model.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/repositories/product_section_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/product_repository.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'product_event.dart';
part 'product_state.dart';

class ProductBloc extends Bloc<ProductEvent, ProductState> {
  final ProductRepository _productRepository;
  final ProductSectionRepository _productSectionRepository;
  final ProductBarcodeRepository _productBarcodeRepository;

  ProductBloc({
    required ProductRepository productRepository,
    required ProductSectionRepository productSectionRepository,
    required ProductBarcodeRepository productBarcodeRepository,
  }) : _productRepository = productRepository,
       _productSectionRepository = productSectionRepository,
       _productBarcodeRepository = productBarcodeRepository,
       super(ProductInitial()) {
    on<ProductLoadList>(onProductLoad);
    on<ProductSave>(onProductSave);
    on<ProductUpdate>(onProductUpdate);
    on<ProductDelete>(onProductDelete);
    on<ProductDeleteMany>(onProductDeleteMany);
    on<ProductGet>(onProductGet);
    on<ProductWithImageSave>(onProductWithImageSave);
    on<ProductWithImageUpdate>(onProductWithImageUpdate);
    on<ProductBranchGet>(onProductBranchGet);
    on<ProductBranchSave>(onProductBranchSave);
    on<ProductDepartmentGet>(onProductDepartmentGet);
    on<ProductDepartmentSave>(onProductDepartmentSave);
    on<ProductGetByBarcodes>(onGetProductByBarcodes);
  }

  void onProductLoad(ProductLoadList event, Emitter<ProductState> emit) async {
    emit(ProductInProgress());

    try {
      final results = await _productRepository.getProductList(
        offset: event.offset,
        limit: event.limit,
        search: event.search,
      );

      if (results.success) {
        List<ProductModel> products = (results.data as List)
            .map((product) => ProductModel.fromJson(product))
            .toList();
        emit(ProductLoadSuccess(products: products));
      } else {
        emit(const ProductLoadFailed(message: 'Product Not Found'));
      }
    } catch (e) {
      // print(e);
      emit(ProductLoadFailed(message: e.toString()));
    }
  }

  void onProductDelete(ProductDelete event, Emitter<ProductState> emit) async {
    emit(ProductDeleteInProgress());
    try {
      await _productRepository.deleteProduct(event.guid);

      emit(ProductDeleteSuccess());
    } catch (e) {
      // emit(ProductDeleteFailure(message: e.toString()));
    }
  }

  void onProductDeleteMany(
    ProductDeleteMany event,
    Emitter<ProductState> emit,
  ) async {
    emit(ProductDeleteManyInProgress());
    try {
      await _productRepository.deleteProductMany(event.guid);

      emit(ProductDeleteManySuccess());
    } catch (e) {
      emit(ProductDeleteManyFailed(message: e.toString()));
    }
  }

  void onProductSave(ProductSave event, Emitter<ProductState> emit) async {
    emit(ProductSaveInProgress());
    try {
      await _productRepository.saveProduct(event.product);
      emit(ProductSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(ProductSaveFailed(message: error['message']));
    }
  }

  void onProductUpdate(ProductUpdate event, Emitter<ProductState> emit) async {
    emit(ProductUpdateInProgress());
    try {
      await _productRepository.updateProduct(event.guid, event.product);
      emit(ProductUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(ProductUpdateFailed(message: error['message']));
    }
  }

  void onProductGet(ProductGet event, Emitter<ProductState> emit) async {
    emit(ProductGetInProgress());
    try {
      final result = await _productRepository.getProduct(event.guid);
      if (result.success) {
        ProductModel product = ProductModel.fromJson(result.data);
        emit(ProductGetSuccess(product: product));
      } else {
        emit(const ProductGetFailed(message: 'Product Not Found'));
      }
    } catch (e) {
      // emit(ProductDeleteFailure(message: e.toString()));
    }
  }

  void onProductWithImageSave(
    ProductWithImageSave event,
    Emitter<ProductState> emit,
  ) async {
    emit(ProductSaveInProgress());
    try {
      List<ImagesModel> imagesList = [];
      if (event.imageWeb.isNotEmpty) {
        for (int i = 0; i < event.imageWeb.length; i++) {
          if (event.imageWeb[i].isNotEmpty) {
            ApiResponse result = await _productRepository.uploadImage(
              event.imageFile[i],
              event.imageWeb[i],
            );
            if (result.success) {
              UploadImageModel uploadImage = UploadImageModel.fromJson(
                result.data,
              );
              imagesList.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(ProductSaveFailed(message: result.message));
            }
          }
        }

        if (imagesList.length == event.imageFile.length) {
          ProductModel product = event.product;
          product.images = imagesList;

          await _productRepository.saveProduct(product);
          emit(ProductSaveSuccess());
        } else {
          emit(const ProductSaveFailed(message: 'image upload failed'));
        }
      } else {
        emit(const ProductSaveFailed(message: 'no image found'));
      }
    } catch (e) {
      emit(ProductSaveFailed(message: e.toString()));
    }
  }

  void onProductWithImageUpdate(
    ProductWithImageUpdate event,
    Emitter<ProductState> emit,
  ) async {
    emit(ProductUpdateInProgress());
    try {
      List<ImagesModel> imagesList = [];
      if (event.imagesUri.isNotEmpty) {
        for (int i = 0; i < event.imagesUri.length; i++) {
          if (event.imageWeb[i].isNotEmpty) {
            ApiResponse result = await _productRepository.uploadImage(
              event.imageFile[i],
              event.imageWeb[i],
            );
            if (result.success) {
              UploadImageModel uploadImage = UploadImageModel.fromJson(
                result.data,
              );
              imagesList.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(ProductUpdateFailed(message: result.message));
            }
          } else if (event.imagesUri[i].uri != '') {
            imagesList.add(ImagesModel(uri: event.imagesUri[i].uri, xorder: i));
          }
        }

        if (imagesList.isNotEmpty) {
          ProductModel productModel = event.product;
          productModel.images = imagesList;

          await _productRepository.updateProduct(event.guid, event.product);
          emit(ProductUpdateSuccess());
        } else {
          emit(const ProductUpdateFailed(message: 'image upload failed'));
        }
      } else {
        emit(const ProductUpdateFailed(message: 'no image found'));
      }
    } catch (e) {
      emit(ProductUpdateFailed(message: e.toString()));
    }
  }

  /// ดึงสินค้าตามสาขา
  void onProductBranchGet(
    ProductBranchGet event,
    Emitter<ProductState> emit,
  ) async {
    emit(ProductBranchGetInProgress());
    try {
      final result = await _productSectionRepository.getProductBranch(
        event.branchcode,
      );
      if (result.success) {
        ProductBranchModel productBranch = ProductBranchModel.fromJson(
          result.data,
        );

        final getResultName = await _productBarcodeRepository
            .getProductBarcodeByBarcode(productBranch.productcodes);

        if (getResultName.success) {
          List<ProductBarcodeModel> products = (getResultName.data as List)
              .map((product) => ProductBarcodeModel.fromJson(product))
              .toList();

          emit(ProductBranchGetSuccess(products: products));
        }
      } else {
        emit(const ProductBranchGetFailed(message: 'Product branch Not Found'));
      }
    } catch (e) {
      emit(ProductBranchGetFailed(message: e.toString()));
    }
  }

  /// save สินค้าตามสาขา
  void onProductBranchSave(
    ProductBranchSave event,
    Emitter<ProductState> emit,
  ) async {
    emit(ProductBranchupdateInProgress());
    try {
      await _productSectionRepository.updateBarcodeInBranch(
        event.productBranchModel,
      );
      emit(ProductBranchUpdateSuccess());
    } catch (e) {
      emit(ProductBranchUpdateFailed(message: e.toString()));
    }
  }

  /// ดึงสินค้าตามแผนก
  void onProductDepartmentGet(
    ProductDepartmentGet event,
    Emitter<ProductState> emit,
  ) async {
    emit(ProductDepartmentGetInProgress());
    try {
      final result = await _productSectionRepository.getProductDepartment(
        event.branchcode,
        event.departmentcode,
      );
      if (result.success) {
        ProductDepartmentModel productDepartment =
            ProductDepartmentModel.fromJson(result.data);

        final getResultName = await _productBarcodeRepository
            .getProductBarcodeByBarcode(productDepartment.productcodes);

        if (getResultName.success) {
          List<ProductBarcodeModel> products = (getResultName.data as List)
              .map((product) => ProductBarcodeModel.fromJson(product))
              .toList();

          emit(ProductDepartmentGetSuccess(products: products));
        }
      } else {
        emit(
          const ProductDepartmentGetFailed(message: 'Product branch Not Found'),
        );
      }
    } catch (e) {
      emit(ProductDepartmentGetFailed(message: e.toString()));
    }
  }

  /// save สินค้าตามแผนก
  void onProductDepartmentSave(
    ProductDepartmentSave event,
    Emitter<ProductState> emit,
  ) async {
    emit(ProductDepartmentupdateInProgress());
    try {
      await _productSectionRepository.updateBarcodeInDepartment(
        event.productDepartmentModel,
      );
      emit(ProductDepartmentUpdateSuccess());
    } catch (e) {
      emit(ProductDepartmentUpdateFailed(message: e.toString()));
    }
  }

  void onGetProductByBarcodes(
    ProductGetByBarcodes event,
    Emitter<ProductState> emit,
  ) async {
    if (kDebugMode) {
      AppLogger.debug(
        'ProductBloc: onGetProductByBarcodes called with barcodes: ${event.barcodes}',
      );
    }
    emit(ProductGetByBarcodesInProgress());
    try {
      if (kDebugMode) {
        AppLogger.debug(
          'ProductBloc: Calling repository with barcodes: ${event.barcodes}',
        );
      }
      final result = await _productBarcodeRepository.getProductBarcodeByBarcode(
        event.barcodes,
      );

      if (result.success) {
        List<ProductBarcodeModel> products = (result.data as List)
            .map((product) => ProductBarcodeModel.fromJson(product))
            .toList();
        if (kDebugMode) {
          AppLogger.debug(
            'ProductBloc: Successfully fetched ${products.length} products',
          );
          AppLogger.info(
            'ProductBloc: About to emit ProductGetByBarcodesSuccess',
          );
        }
        emit(
          ProductGetByBarcodesSuccess(products: products, source: event.source),
        );
        if (kDebugMode) {
          AppLogger.debug(
            'ProductBloc: ProductGetByBarcodesSuccess emitted successfully',
          );
        }
      } else {
        if (kDebugMode) {
          AppLogger.debug(
            'ProductBloc: Repository returned failure: ${result.message}',
          );
        }
        emit(const ProductGetByBarcodesFailed(message: 'Products Not Found'));
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('ProductBloc: Exception occurred: $e');
      }
      emit(ProductGetByBarcodesFailed(message: e.toString()));
    }
  }
}
