import 'package:flutter/foundation.dart';
import 'package:smlaicloud/model/shop_model.dart';
import 'package:smlaicloud/repositories/shop_repository.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'shop_event.dart';
part 'shop_state.dart';

class ShopBloc extends Bloc<ShopEvent, ShopState> {
  final ShopRepository _shopRepository;

  ShopBloc({required ShopRepository shopRepository})
      : _shopRepository = shopRepository,
        super(ShopInitial()) {
    on<GetShopInfo>(_onGetShopInfo);
    on<GetMainShopCenterTypes>(_onGetMainShopCenterTypes);
    on<ShopUpdate>(_onShopUpdate);
  }
  void _onGetShopInfo(GetShopInfo event, Emitter<ShopState> emit) async {
    emit(GetShopInfoInProgress());
    try {
      final result = await _shopRepository.loadShopInfo(event.shopid);

      if (result.success) {
        ShopModel shopmodel = ShopModel.fromJson(result.data);
        emit(GetShopInfoSuccess(shop: shopmodel));
      } else {
        emit(const GetShopInfoFailed(message: 'Shop Not Found'));
      }
    } catch (e) {
      emit(GetShopInfoFailed(message: e.toString()));
    }
  }

  void _onGetMainShopCenterTypes(
      GetMainShopCenterTypes event, Emitter<ShopState> emit) async {
    if (kDebugMode) {
      AppLogger.debug(
          "_onGetMainShopCenterTypes called with mainShopId: ${event.mainShopId}");
    }
    emit(GetMainShopCenterTypesInProgress());
    try {
      final result = await _shopRepository.loadShopInfo(event.mainShopId);
      if (kDebugMode) {
        AppLogger.info("Repository result - success: ${result.success}");
      }

      if (result.success) {
        ShopModel mainShopModel = ShopModel.fromJson(result.data);
        int productCenterType = mainShopModel.productcentertype ?? 0;
        int debtorCenterType = mainShopModel.debtorcentertype ?? 0;
        int posProductCenterType = mainShopModel.posproductcentertype ?? 0;

        if (kDebugMode) {
          AppLogger.debug(
              "Main shop - productCenterType: $productCenterType, debtorCenterType: $debtorCenterType, posProductCenterType: $posProductCenterType");
        }

        emit(GetMainShopCenterTypesSuccess(
          productCenterType: productCenterType,
          debtorCenterType: debtorCenterType,
          posProductCenterType: posProductCenterType,
        ));
      } else {
        if (kDebugMode) {
          AppLogger.debug("Main Shop Not Found");
        }
        emit(
            const GetMainShopCenterTypesFailed(message: 'Main Shop Not Found'));
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error("Exception in _onGetMainShopCenterTypes: $e");
      }
      emit(GetMainShopCenterTypesFailed(message: e.toString()));
    }
  }

  void _onShopUpdate(ShopUpdate event, Emitter<ShopState> emit) async {
    emit(ShopUpdateInProgress());
    try {
      final result =
          await _shopRepository.updateShop(event.shopid, event.shopdata);
      if (result.success) {
        emit(ShopUpdateSuccess());
      } else {
        emit(ShopUpdateFailed(
            message: 'Shop Update Failed :  ${result.message}'));
      }
    } catch (e) {
      emit(ShopUpdateFailed(message: e.toString()));
    }
  }
}
