import 'package:smlaicloud/model/kitchen_product_model.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/kitchen_model.dart';
import 'package:smlaicloud/repositories/kitchen_repository.dart';

part 'kitchen_event.dart';
part 'kitchen_state.dart';

class KitchenBloc extends Bloc<KitchenEvent, KitchenState> {
  final KitchenRepository _kitchenRepository;
  final ProductBarcodeRepository _productBarcodeRepository;

  KitchenBloc({
    required KitchenRepository kitchenRepository,
    required ProductBarcodeRepository productBarcodeRepository,
  })  : _kitchenRepository = kitchenRepository,
        _productBarcodeRepository = productBarcodeRepository,
        super(KitchenInitial()) {
    on<KitchenLoadList>(onKitchenLoad);
    on<KitchenSave>(onKitchenSave);
    on<KitchenUpdate>(onKitchenUpdate);
    on<KitchenDelete>(kitchenDelete);
    on<KitchenDeleteMany>(kitchenDeleteMany);
    on<KitchenGet>(onKitchenGet);
    on<KitchenDetailBarcodeGet>(onKitchenDetailBarcodeGet);
    on<KitchenProductsLoadList>(onKitchenProductsLoad);
  }

  void onKitchenLoad(KitchenLoadList event, Emitter<KitchenState> emit) async {
    emit(KitchenInProgress());

    try {
      final results = await _kitchenRepository.getKitchenList(
          offset: event.offset,
          limit: event.limit,
          search: event.search,
          groupNumber: event.groupnumber);

      if (results.success) {
        List<KitchenModel> kitchens = (results.data as List)
            .map((kitchen) => KitchenModel.fromJson(kitchen))
            .toList();
        emit(KitchenLoadSuccess(kitchens: kitchens));
      } else {
        emit(const KitchenLoadFailed(message: 'Kitchen Group Not Found'));
      }
    } catch (e) {
      emit(KitchenLoadFailed(message: e.toString()));
    }
  }

  void onKitchenProductsLoad(
      KitchenProductsLoadList event, Emitter<KitchenState> emit) async {
    emit(KitchenProductsInProgress());

    try {
      final results = await _kitchenRepository.getKitchenProductsList();

      if (results.success) {
        List<ProductInKitchenModel> kitchens = (results.data as List)
            .map((kitchen) => ProductInKitchenModel.fromJson(kitchen))
            .toList();
        emit(KitchenProductsLoadSuccess(kitchens: kitchens));
      } else {
        emit(const KitchenProductsLoadFailed(message: 'Products Not Found'));
      }
    } catch (e) {
      emit(KitchenProductsLoadFailed(message: e.toString()));
    }
  }

  void kitchenDelete(KitchenDelete event, Emitter<KitchenState> emit) async {
    emit(KitchenDeleteInProgress());
    try {
      await _kitchenRepository.deleteKitchen(event.guid);

      emit(KitchenDeleteSuccess());
    } catch (e) {
      // emit(KitchenDeleteFailure(message: e.toString()));
    }
  }

  void kitchenDeleteMany(
      KitchenDeleteMany event, Emitter<KitchenState> emit) async {
    emit(KitchenDeleteManyInProgress());
    try {
      await _kitchenRepository.deleteKitchenMany(event.guid);

      emit(KitchenDeleteManySuccess());
    } catch (e) {
      // emit(KitchenDeleteFailure(message: e.toString()));
    }
  }

  void onKitchenSave(KitchenSave event, Emitter<KitchenState> emit) async {
    emit(KitchenSaveInProgress());
    try {
      await _kitchenRepository.saveKitchen(event.kitchenModel);
      emit(KitchenSaveSuccess());
    } catch (e) {
      emit(KitchenSaveFailed(message: e.toString()));
    }
  }

  void onKitchenUpdate(KitchenUpdate event, Emitter<KitchenState> emit) async {
    emit(KitchenUpdateInProgress());
    try {
      await _kitchenRepository.updateKitchen(event.guid, event.kitchenModel);
      emit(KitchenUpdateSuccess());
    } catch (e) {
      emit(KitchenUpdateFailed(message: e.toString()));
    }
  }

  void onKitchenGet(KitchenGet event, Emitter<KitchenState> emit) async {
    emit(KitchenGetInProgress());
    try {
      final result = await _kitchenRepository.getKitchen(event.guid);
      if (result.success) {
        KitchenModel kitchen = KitchenModel.fromJson(result.data);

        emit(KitchenGetSuccess(kitchen: kitchen));
      } else {
        emit(const KitchenGetFailed(message: 'Kitchen Not Found'));
      }
    } catch (e) {
      emit(KitchenDeleteFailed(message: e.toString()));
    }
  }

  void onKitchenDetailBarcodeGet(
      KitchenDetailBarcodeGet event, Emitter<KitchenState> emit) async {
    emit(KitchenDetailBarcodeGetInProgress());
    try {
      final result = await _kitchenRepository.getKitchen(event.guid);
      if (result.success) {
        KitchenModel kitchen = KitchenModel.fromJson(result.data);

        final getResultName = await _productBarcodeRepository
            .getProductBarcodeByBarcode(kitchen.products!);

        if (getResultName.success) {
          GetKitchenModel getKitchenModel = GetKitchenModel(
            guidfixed: kitchen.guidfixed!,
            code: kitchen.code,
            names: kitchen.names,
            printers: kitchen.printers,
            zones: kitchen.zones,
          );

          List<dynamic> resultData = [];
          for (int i = 0; i < getResultName.data.length; i++) {
            if (getResultName.data[i] != null) {
              resultData.add(getResultName.data[i]);
            }
          }

          List<KitchenProductModel> products = (resultData)
              .map((product) => KitchenProductModel.fromJson(product))
              .toList();

          if (products.isNotEmpty) {
            getKitchenModel.products = products;
          } else {
            getKitchenModel.products = <KitchenProductModel>[];
          }

          emit(KitchenDetailBarcodeGetSuccess(kitchen: getKitchenModel));
        }
      } else {
        emit(const KitchenDetailBarcodeGetFailed(message: 'Kitchen Not Found'));
      }
    } catch (e) {
      emit(KitchenDeleteFailed(message: e.toString()));
    }
  }
}
