import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:smlaicloud/model/product_bom_model.dart';
import 'package:smlaicloud/model/price_history_model.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'product_barcode_event.dart';
part 'product_barcode_state.dart';

class ProductBarcodeBloc
    extends Bloc<ProductBarcodeEvent, ProductBarcodeState> {
  final ProductBarcodeRepository _productBarcodeRepository;

  ProductBarcodeBloc({
    required ProductBarcodeRepository productBarcodeRepository,
  }) : _productBarcodeRepository = productBarcodeRepository,
       super(ProductBarcodeInitial()) {
    on<ProductBarcodeLoadList>(onProductBarcodeLoad);
    on<ProductBarcodeSave>(onProductBarcodeSave);
    on<ProductBarcodeUpdate>(onProductBarcodeUpdate);
    on<ProductBarcodeDelete>(productbarcodeDelete);
    on<ProductBarcodeDeleteMany>(productbarcodeDeleteMany);
    on<ProductBarcodeGet>(onProductBarcodeGet);
    on<ProductBarcodeGetRef>(onProductBarcodeGetRef);
    on<ProductBarcodeWithImageSave>(onProductBarcodeWithImageSave);
    on<ProductBarcodeWithImageUpdate>(onProductBarcodeWithImageUpdate);
    on<ProductBarcodeGetBom>(onProductBarcodeGetBom);
    on<ProductBarcodeLoadListSearch>(onProductBarcodeLoadSearch);
    on<ProductBarcodeGetByBarcodeList>(onProductBarcodeGetByBarcodeList);
    on<ProductBarcodeGetPriceHistory>(onProductBarcodeGetPriceHistory);
    on<ProductBarcodeLoadListPg>(onProductBarcodeLoadListPg);
    on<ProductBarcodeGetByBarcode>(onProductBarcodeGetByBarcode);
  }

  void onProductBarcodeLoad(
    ProductBarcodeLoadList event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeInProgress());

    try {
      final results = await _productBarcodeRepository.getProductBarcodeList(
        offset: event.offset,
        limit: event.limit,
        search: event.search,
        itemtype: event.itemtype!,
        branchcode: event.branchcode,
        businesstypecode: event.businesstypecode,
        isbom: event.isbom!,
        isusesubbarcodes: event.isusesubbarcodes!,
        shopsid: event.shopsid!,
      );

      AppLogger.debug('🔍 BLOC: results.success = ${results.success}');
      AppLogger.debug(
        '🔍 BLOC: results.data type = ${results.data.runtimeType}',
      );
      AppLogger.debug('🔍 BLOC: results.data = ${results.data}');

      if (results.success) {
        if (results.data == null) {
          AppLogger.warning('⚠️ BLOC: results.data is null!');
          emit(ProductBarcodeLoadSuccess(productBarcodes: []));
          return;
        }

        AppLogger.debug(
          '🔍 BLOC: กำลัง parse data เป็น ProductBarcodeModel...',
        );
        List<ProductBarcodeModel> productbarcodes = (results.data as List).map((
          productbarcode,
        ) {
          final barcode = productbarcode['barcode'];
          AppLogger.debug('🔍 BLOC: Parsing item: $barcode');

          // Log ข้อมูลดิบของ ITEMTEST002 เพื่อดูว่าฟิลด์ไหนเป็น null
          if (barcode == 'ITEMTEST002') {
            AppLogger.debug('🔍 ITEMTEST002 RAW DATA:');
            AppLogger.debug('  - groupnames: ${productbarcode['groupnames']}');
            AppLogger.debug('  - brandnames: ${productbarcode['brandnames']}');
            AppLogger.debug(
              '  - categorynames: ${productbarcode['categorynames']}',
            );
            AppLogger.debug('  - classnames: ${productbarcode['classnames']}');
            AppLogger.debug(
              '  - designnames: ${productbarcode['designnames']}',
            );
            AppLogger.debug('  - gradenames: ${productbarcode['gradenames']}');
            AppLogger.debug('  - modelnames: ${productbarcode['modelnames']}');
            AppLogger.debug(
              '  - patternnames: ${productbarcode['patternnames']}',
            );
            AppLogger.debug(
              '  - groupsubonenames: ${productbarcode['groupsubonenames']}',
            );
            AppLogger.debug(
              '  - groupsubtwonames: ${productbarcode['groupsubtwonames']}',
            );
            AppLogger.debug(
              '  - itemunitnames: ${productbarcode['itemunitnames']}',
            );
            AppLogger.debug(
              '  - manufacturernames: ${productbarcode['manufacturernames']}',
            );
            AppLogger.debug('  - dimensions: ${productbarcode['dimensions']}');
            AppLogger.debug(
              '  - businesstypes: ${productbarcode['businesstypes']}',
            );
            AppLogger.debug(
              '  - ignorebranches: ${productbarcode['ignorebranches']}',
            );
            AppLogger.debug('  - branches: ${productbarcode['branches']}');
            AppLogger.debug('  - categorys: ${productbarcode['categorys']}');
            AppLogger.debug(
              '  - timeforsales: ${productbarcode['timeforsales']}',
            );
            AppLogger.debug('  - fixedcost: ${productbarcode['fixedcost']}');
            AppLogger.debug('  - options: ${productbarcode['options']}');
            AppLogger.debug('  - prices: ${productbarcode['prices']}');
            AppLogger.debug('  - names: ${productbarcode['names']}');
            AppLogger.debug('  - barcodes: ${productbarcode['barcodes']}');
            AppLogger.debug(
              '  - refbarcodes: ${productbarcode['refbarcodes']}',
            );
            AppLogger.debug('  - bom: ${productbarcode['bom']}');
            AppLogger.debug('  - ordertypes: ${productbarcode['ordertypes']}');
            AppLogger.debug('  - xsorts: ${productbarcode['xsorts']}');
          }

          return ProductBarcodeModel.fromJson(productbarcode);
        }).toList();
        AppLogger.info('✅ BLOC: Parse สำเร็จ ${productbarcodes.length} รายการ');
        emit(ProductBarcodeLoadSuccess(productBarcodes: productbarcodes));
      } else {
        emit(
          const ProductBarcodeLoadFailed(message: 'ProductBarcode Not Found'),
        );
      }
    } catch (e) {
      AppLogger.error('❌ BLOC ERROR: ${e.toString()}');
      if (e is TypeError) {
        AppLogger.error('❌ BLOC TypeError: ${e.toString()}');
        AppLogger.error('❌ BLOC StackTrace: ${StackTrace.current}');
      }
      emit(ProductBarcodeLoadFailed(message: e.toString()));
    }
  }

  void onProductBarcodeLoadSearch(
    ProductBarcodeLoadListSearch event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeSearchInProgress());

    try {
      final results = await _productBarcodeRepository.getProductBarcodeList(
        offset: event.offset,
        limit: event.limit,
        search: event.search,
        itemtype: event.itemtype!,
        branchcode: event.branchcode,
        businesstypecode: event.businesstypecode,
        isbom: event.isbom!,
        isusesubbarcodes: event.isusesubbarcodes!,
        shopsid: event.shopsid!,
      );

      if (results.success) {
        if (results.data == null) {
          AppLogger.warning('⚠️ BLOC SEARCH: results.data is null!');
          emit(ProductBarcodeLoadSearchSuccess(productBarcodes: []));
          return;
        }

        List<ProductBarcodeModel> productbarcodes = (results.data as List)
            .map(
              (productbarcode) => ProductBarcodeModel.fromJson(productbarcode),
            )
            .toList();
        emit(ProductBarcodeLoadSearchSuccess(productBarcodes: productbarcodes));
      } else {
        emit(
          const ProductBarcodeLoadSearchFailed(
            message: 'ProductBarcode Not Found',
          ),
        );
      }
    } catch (e) {
      AppLogger.error('❌ BLOC SEARCH ERROR: ${e.toString()}');
      emit(ProductBarcodeLoadSearchFailed(message: e.toString()));
    }
  }

  void productbarcodeDelete(
    ProductBarcodeDelete event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeDeleteInProgress());
    try {
      await _productBarcodeRepository.deleteProductBarcode(event.guid);

      emit(ProductBarcodeDeleteSuccess());
    } catch (e) {
      // emit(ProductBarcodeDeleteFailure(message: e.toString()));
    }
  }

  void productbarcodeDeleteMany(
    ProductBarcodeDeleteMany event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeDeleteManyInProgress());
    try {
      await _productBarcodeRepository.deleteProductBarcodeMany(event.guid);

      emit(ProductBarcodeDeleteManySuccess());
    } catch (e) {
      emit(ProductBarcodeDeleteManyFailed(message: e.toString()));
    }
  }

  void onProductBarcodeSave(
    ProductBarcodeSave event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeSaveInProgress());
    try {
      await _productBarcodeRepository.saveProductBarcode(event.productBarcode);
      emit(ProductBarcodeSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(ProductBarcodeSaveFailed(message: error['message']));
    }
  }

  void onProductBarcodeUpdate(
    ProductBarcodeUpdate event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeUpdateInProgress());
    try {
      await _productBarcodeRepository.updateProductBarcode(
        event.guid,
        event.productBarcode,
      );
      emit(ProductBarcodeUpdateSuccess());
    } catch (e) {
      emit(ProductBarcodeUpdateFailed(message: e.toString()));
    }
  }

  void onProductBarcodeGet(
    ProductBarcodeGet event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    AppLogger.info(
      '🔍 BLOC GET: กำลังโหลดข้อมูลสินค้าจาก API (guid: ${event.guid})',
    );
    emit(ProductBarcodeGetInProgress());

    try {
      final result = await _productBarcodeRepository.getProductBarcode(
        event.guid,
      );

      AppLogger.debug('🔍 BLOC GET: API response - success: ${result.success}');

      if (result.success) {
        AppLogger.debug('🔍 BLOC GET: Raw data from API: ${result.data}');

        try {
          ProductBarcodeModel productBarcode = ProductBarcodeModel.fromJson(
            result.data,
          );

          AppLogger.info(
            '✅ BLOC GET SUCCESS: Parse สำเร็จ - barcode: ${productBarcode.barcode}',
          );
          emit(ProductBarcodeGetSuccess(productBarcode: productBarcode));
        } catch (parseError, parseStackTrace) {
          AppLogger.error('❌ PARSE ERROR: ไม่สามารถ parse JSON ได้');
          AppLogger.error('Parse Error: $parseError');

          // Log ข้อมูลแต่ละ field เพื่อหา field ที่มีปัญหา
          if (result.data is Map) {
            final data = result.data as Map<String, dynamic>;
            AppLogger.error('🔍 Checking each field:');

            // Check list fields ที่น่าจะมีปัญหา
            final listFields = [
              'itemunitnames',
              'groupnames',
              'names',
              'prices',
              'options',
              'barcodes',
              'refbarcodes',
              'bom',
              'ordertypes',
              'manufacturernames',
              'dimensions',
              'businesstypes',
              'ignorebranches',
              'branches',
              'categorys',
              'timeforsales',
              'fixedcost',
              'brandnames',
              'categorynames',
              'classnames',
              'designnames',
              'gradenames',
              'modelnames',
              'groupsubonenames',
              'groupsubtwonames',
              'patternnames',
              'xsorts',
            ];

            for (final field in listFields) {
              if (data.containsKey(field)) {
                final value = data[field];
                final type = value.runtimeType;
                if (value == null) {
                  AppLogger.error('  ❌ $field: null (should be List)');
                } else if (value is! List) {
                  AppLogger.error(
                    '  ❌ $field: $type (should be List) - value: $value',
                  );
                } else {
                  AppLogger.debug('  ✅ $field: List (${value.length} items)');
                }
              }
            }
          }

          AppLogger.error('Parse Stack trace: $parseStackTrace');
          emit(ProductBarcodeGetFailed(message: 'Parse Error: $parseError'));
        }
      } else {
        AppLogger.error(
          '❌ BLOC GET FAILED: ProductBarcode Not Found (guid: ${event.guid})',
        );
        emit(
          const ProductBarcodeGetFailed(message: 'ProductBarcode Not Found'),
        );
      }
    } catch (e, stackTrace) {
      AppLogger.error('❌ BLOC GET ERROR: $e');
      AppLogger.error('Stack trace: $stackTrace');
      emit(ProductBarcodeGetFailed(message: e.toString()));
    }
  }

  void onProductBarcodeGetRef(
    ProductBarcodeGetRef event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeGetRefInProgress());
    // print(event.guid);
    try {
      final result = await _productBarcodeRepository.getProductBarcodeRef(
        event.guid,
      );
      if (result.success) {
        List<ProductBarcodeModel> productbarcodes = (result.data as List)
            .map(
              (productbarcode) => ProductBarcodeModel.fromJson(productbarcode),
            )
            .toList();
        emit(ProductBarcodeGetRefSuccess(productBarcodes: productbarcodes));
      } else {
        emit(
          const ProductBarcodeGetRefFailed(
            message: 'ProductBarcodeRef Not Found',
          ),
        );
      }
    } catch (e) {
      emit(ProductBarcodeGetRefFailed(message: e.toString()));
    }
  }

  void onProductBarcodeWithImageSave(
    ProductBarcodeWithImageSave event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeSaveInProgress());
    try {
      ApiResponse result = await _productBarcodeRepository
          .uploadFileNoLimitSize(event.imageFile, event.imageWeb!);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        ProductBarcodeModel productBarcodeModel = event.productBarcode;
        productBarcodeModel.imageuri = uploadImage.uri;
        await _productBarcodeRepository.saveProductBarcode(productBarcodeModel);
        emit(ProductBarcodeSaveSuccess());
      } else {
        emit(ProductBarcodeSaveFailed(message: result.message));
      }
    } catch (e) {
      emit(ProductBarcodeSaveFailed(message: e.toString()));
    }
  }

  void onProductBarcodeWithImageUpdate(
    ProductBarcodeWithImageUpdate event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeUpdateInProgress());
    try {
      ApiResponse result = await _productBarcodeRepository
          .uploadFileNoLimitSize(event.imageFile, event.imageWeb);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        ProductBarcodeModel productBarcodeModel = event.productBarcode;
        productBarcodeModel.imageuri = uploadImage.uri;
        await _productBarcodeRepository.updateProductBarcode(
          event.guid,
          productBarcodeModel,
        );
        emit(ProductBarcodeUpdateSuccess());
      } else {
        emit(ProductBarcodeUpdateFailed(message: result.message));
      }
    } catch (e) {
      emit(ProductBarcodeUpdateFailed(message: e.toString()));
    }
  }

  void onProductBarcodeGetBom(
    ProductBarcodeGetBom event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeGetBomInProgress());
    try {
      final result = await _productBarcodeRepository.getProductBarcodeBom(
        event.barcode,
      );
      if (result.success) {
        ProductBomModel productbom = ProductBomModel.fromJson(result.data);
        if (productbom.bom!.isNotEmpty) {}

        emit(ProductBarcodeGetBomSuccess(productBom: productbom));
      } else {
        emit(
          const ProductBarcodeGetBomFailed(
            message: 'ProductBarcodeBom Not Found',
          ),
        );
      }
    } catch (e) {
      emit(ProductBarcodeGetBomFailed(message: e.toString()));
    }
  }

  void onProductBarcodeGetByBarcodeList(
    ProductBarcodeGetByBarcodeList event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeGetByBarcodeListInProgress());
    try {
      final result = await _productBarcodeRepository.getProductBarcodeByBarcode(
        event.barcodes,
      );
      if (result.success) {
        List<ProductBarcodeModel> productbarcodes = (result.data as List)
            .map(
              (productbarcode) => ProductBarcodeModel.fromJson(productbarcode),
            )
            .toList();
        emit(
          ProductBarcodeGetByBarcodeListSuccess(
            productBarcodes: productbarcodes,
          ),
        );
      } else {
        emit(
          const ProductBarcodeGetByBarcodeListFailed(
            message: 'ProductBarcode Not Found',
          ),
        );
      }
    } catch (e) {
      emit(ProductBarcodeGetByBarcodeListFailed(message: e.toString()));
    }
  }

  void onProductBarcodeLoadListPg(
    ProductBarcodeLoadListPg event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeLoadPgInProgress());

    try {
      final result = await _productBarcodeRepository.searchBarcodeListPg(
        keyword: event.keyword,
        groupCode: event.groupCode,
        brandCode: event.brandCode,
        categoryCode: event.categoryCode,
        classCode: event.classCode,
        designCode: event.designCode,
        gradeCode: event.gradeCode,
        modelCode: event.modelCode,
        patternCode: event.patternCode,
        priceMin: event.priceMin,
        priceMax: event.priceMax,
        limit: event.limit,
        offset: event.offset,
        sortField: event.sortField,
        sortOrder: event.sortOrder,
      );

      if (result['success'] == true) {
        final data = result['data'];
        final total = result['total'] ?? 0;

        if (data == null || (data is List && data.isEmpty)) {
          emit(ProductBarcodeLoadPgSuccess(productBarcodes: [], total: total is int ? total : 0));
          return;
        }

        List<ProductBarcodeModel> productbarcodes = (data as List)
            .map((item) => ProductBarcodeModel.fromJson(item))
            .toList();

        emit(ProductBarcodeLoadPgSuccess(
          productBarcodes: productbarcodes,
          total: total is int ? total : 0,
        ));
      } else {
        emit(ProductBarcodeLoadPgFailed(
          message: result['message'] ?? 'Failed to load barcode list',
        ));
      }
    } catch (e) {
      AppLogger.error('onProductBarcodeLoadListPg error: $e');
      emit(ProductBarcodeLoadPgFailed(message: e.toString()));
    }
  }

  void onProductBarcodeGetByBarcode(
    ProductBarcodeGetByBarcode event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    AppLogger.info('🔍 BLOC GET BY BARCODE: ค้นหาสินค้าจาก MongoDB ด้วย barcode: ${event.barcode}');
    emit(ProductBarcodeGetInProgress());

    try {
      // ค้นหาจาก MongoDB list API ด้วย barcode เป็น keyword
      final result = await _productBarcodeRepository.getProductBarcodeList(
        offset: 0,
        limit: 5,
        search: event.barcode,
      );

      if (result.success && result.data != null) {
        final dataList = result.data is List ? result.data as List : [result.data];

        // หา item ที่ barcode ตรงกัน
        for (final item in dataList) {
          try {
            final productBarcode = ProductBarcodeModel.fromJson(
              item is Map<String, dynamic> ? item : {},
            );
            if (productBarcode.barcode == event.barcode) {
              AppLogger.info('✅ BLOC GET BY BARCODE: พบสินค้า barcode=${productBarcode.barcode}, guidfixed=${productBarcode.guidfixed}');
              emit(ProductBarcodeGetSuccess(productBarcode: productBarcode));
              return;
            }
          } catch (e) {
            AppLogger.error('❌ BLOC GET BY BARCODE: parse item error: $e');
          }
        }

        // ถ้าไม่เจอ exact match ใช้ตัวแรก
        if (dataList.isNotEmpty) {
          try {
            final productBarcode = ProductBarcodeModel.fromJson(
              dataList[0] is Map<String, dynamic> ? dataList[0] : {},
            );
            AppLogger.info('✅ BLOC GET BY BARCODE: ใช้ผลลัพธ์แรก barcode=${productBarcode.barcode}');
            emit(ProductBarcodeGetSuccess(productBarcode: productBarcode));
            return;
          } catch (e) {
            AppLogger.error('❌ BLOC GET BY BARCODE: parse first item error: $e');
          }
        }

        emit(const ProductBarcodeGetFailed(message: 'ไม่พบสินค้าจาก barcode'));
      } else {
        emit(const ProductBarcodeGetFailed(message: 'ไม่พบสินค้าจาก barcode'));
      }
    } catch (e) {
      AppLogger.error('onProductBarcodeGetByBarcode error: $e');
      emit(ProductBarcodeGetFailed(message: e.toString()));
    }
  }

  void onProductBarcodeGetPriceHistory(
    ProductBarcodeGetPriceHistory event,
    Emitter<ProductBarcodeState> emit,
  ) async {
    emit(ProductBarcodeGetPriceHistoryInProgress());
    try {
      final result = await _productBarcodeRepository
          .getProductBarcodePriceHistory(
            barcode: event.barcode,
            page: event.page,
            limit: event.limit,
          );

      if (result.success == true && result.data != null) {
        emit(
          ProductBarcodeGetPriceHistorySuccess(
            priceHistory: result.data!,
            pagination: result.pagination,
          ),
        );
      } else {
        emit(
          const ProductBarcodeGetPriceHistoryFailed(
            message: 'Failed to load price history',
          ),
        );
      }
    } catch (e) {
      emit(ProductBarcodeGetPriceHistoryFailed(message: e.toString()));
    }
  }
}
