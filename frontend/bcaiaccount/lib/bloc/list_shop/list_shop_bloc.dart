import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:smlaicloud/model/shop_list_model.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/user_repository.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'list_shop_event.dart';
part 'list_shop_state.dart';

class ListShopBloc extends Bloc<ListShopEvent, ListShopState> {
  final UserRepository _userRepository;

  ListShopBloc({required UserRepository userRepository}) : _userRepository = userRepository, super(ListShopInitial()) {
    on<ListShopLoad>(_onListShopLoad);
  }

  void _onListShopLoad(ListShopLoad event, Emitter<ListShopState> emit) async {
    emit(ListShopInProgress());
    try {
      if (kDebugMode) {
        AppLogger.info('[ListShopBloc] 🔄 Starting to fetch shop list...');
      }

      final result = await _userRepository.getShopList();

      if (result.success) {
        List<ShopListModel> shop0 = [];
        final totalShops = (result.data as List).length;

        if (kDebugMode) {
          AppLogger.info('[ListShopBloc] 📦 Received $totalShops shops from API');
        }

        // วนทีละตัวเพื่อจัดการ error ในแต่ละ item
        for (int i = 0; i < totalShops; i++) {
          var shopData = (result.data as List)[i];
          try {
            ShopListModel shop = ShopListModel.fromJson(shopData);
            shop0.add(shop);
          } catch (e, stackTrace) {
            if (kDebugMode) {
              AppLogger.debug('═══════════════════════════════════════');
              AppLogger.error('[ListShopBloc] ❌ ERROR parsing shop at index $i');
              AppLogger.error('[ListShopBloc] Error: $e');
              AppLogger.debug('[ListShopBloc] Stack trace: $stackTrace');
              AppLogger.debug('[ListShopBloc] Raw JSON data:');

              // แสดง JSON ในรูปแบบที่อ่านง่าย
              try {
                String prettyJson = const JsonEncoder.withIndent('  ').convert(shopData);
                AppLogger.debug(prettyJson);
              } catch (jsonError) {
                AppLogger.debug('[ListShopBloc] Cannot format as JSON: $shopData');
              }

              // แสดงรายละเอียดแต่ละ field
              if (shopData is Map) {
                AppLogger.debug('[ListShopBloc] Field details:');
                shopData.forEach((key, value) {
                  AppLogger.debug('  - $key: $value (${value.runtimeType})');
                });
              }

              AppLogger.debug('═══════════════════════════════════════');
            }
            // ข้าม item ที่มีปัญหาและดำเนินการต่อ
            continue;
          }
        }

        if (kDebugMode) {
          AppLogger.info('[ListShopBloc] ✅ Successfully parsed ${shop0.length}/$totalShops shops');
          if (shop0.length < totalShops) {
            AppLogger.warn('[ListShopBloc] ⚠️  Skipped ${totalShops - shop0.length} shops due to parsing errors');
          }
        }

        emit(ListShopLoadSuccess(shop: shop0));
      } else {
        if (kDebugMode) {
          AppLogger.error('[ListShopBloc] ❌ API returned failure: Shop Not Found');
        }
        emit(ListShopLoadFailed(message: 'Shop Not Found'));
      }
    } catch (e, stackTrace) {
      if (kDebugMode) {
        AppLogger.error('[ListShopBloc] ❌ Fatal error in _onListShopLoad: $e');
        AppLogger.debug('[ListShopBloc] Stack trace: $stackTrace');
      }
      emit(ListShopLoadFailed(message: e.toString()));
    }
  }
}
