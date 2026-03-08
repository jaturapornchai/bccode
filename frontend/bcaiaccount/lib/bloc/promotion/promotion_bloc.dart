import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/promotion_model.dart';
import 'package:smlaicloud/repositories/promotion_repository.dart';

part 'promotion_event.dart';
part 'promotion_state.dart';

class PromotionBloc extends Bloc<PromotionEvent, PromotionState> {
  final PromotionRepository _promotionRepository;

  PromotionBloc({required PromotionRepository promotionRepository})
      : _promotionRepository = promotionRepository,
        super(PromotionInitial()) {
    on<PromotionLoadList>(onPromotionLoad);
    on<PromotionSave>(onPromotionSave);
    on<PromotionUpdate>(onPromotionUpdate);
    on<PromotionDelete>(onPromotionDelete);
    on<PromotionDeleteMany>(onPromotionDeleteMany);
    on<PromotionGet>(onPromotionGet);
  }

  void onPromotionLoad(
      PromotionLoadList event, Emitter<PromotionState> emit) async {
    emit(PromotionInProgress());
    try {
      final results = await _promotionRepository.getPromotionList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<PromotionModel> promotions = (results.data as List)
            .map((promotion) => PromotionModel.fromJson(promotion))
            .toList();
        // print(promotions.length);
        emit(PromotionLoadSuccess(promotions: promotions));
      } else {
        emit(const PromotionLoadFailed(message: 'Promotion Not Found'));
      }
    } catch (e) {
      emit(PromotionLoadFailed(message: e.toString()));
    }
  }

  void onPromotionDelete(
      PromotionDelete event, Emitter<PromotionState> emit) async {
    emit(PromotionDeleteInProgress());
    try {
      await _promotionRepository.deletePromotion(event.guid);

      emit(PromotionDeleteSuccess());
    } catch (e) {
      emit(PromotionDeleteFailed(message: e.toString()));
    }
  }

  void onPromotionDeleteMany(
      PromotionDeleteMany event, Emitter<PromotionState> emit) async {
    emit(PromotionDeleteManyInProgress());
    try {
      await _promotionRepository.deletePromotionMany(event.guid);

      emit(PromotionDeleteManySuccess());
    } catch (e) {
      emit(PromotionDeleteManyFailed(message: e.toString()));
    }
  }

  void onPromotionSave(
      PromotionSave event, Emitter<PromotionState> emit) async {
    emit(PromotionSaveInProgress());
    try {
      await _promotionRepository.savePromotion(event.promotionModel);
      emit(PromotionSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(PromotionSaveFailed(message: error['message']));
    }
  }

  void onPromotionUpdate(
      PromotionUpdate event, Emitter<PromotionState> emit) async {
    emit(PromotionUpdateInProgress());
    try {
      await _promotionRepository.updatePromotion(
          event.guid, event.promotionModel);
      emit(PromotionUpdateSuccess());
    } catch (e) {
      emit(PromotionUpdateFailed(message: e.toString()));
    }
  }

  void onPromotionGet(PromotionGet event, Emitter<PromotionState> emit) async {
    emit(PromotionGetInProgress());
    try {
      final result = await _promotionRepository.getPromotion(event.guid);
      if (result.success) {
        PromotionModel promotion = PromotionModel.fromJson(result.data);
        emit(PromotionGetSuccess(promotions: promotion));
      } else {
        emit(const PromotionGetFailed(message: 'Promotion Not Found'));
      }
    } catch (e) {
      emit(PromotionGetFailed(message: e.toString()));
    }
  }
}
