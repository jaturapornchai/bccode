import 'dart:convert';

import 'package:smlaicloud/model/wallet_model.dart';
import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'wallet_pay_event.dart';
part 'wallet_pay_state.dart';

class WalletPayBloc extends Bloc<WalletPayEvent, WalletPayState> {
  final JsonRepository _jsonRepository;

  WalletPayBloc({required JsonRepository jsonRepository})
      : _jsonRepository = jsonRepository,
        super(WalletPayInitial()) {
    on<WalletPayLoadList>(onWalletPayLoad);
    on<WalletPaySave>(onWalletPaySave);
    on<WalletPayUpdate>(onWalletPayUpdate);
    on<WalletPayDelete>(onWalletPayDelete);
    on<WalletPayDeleteMany>(onWalletPayDeleteMany);
  }

  void onWalletPayLoad(
      WalletPayLoadList event, Emitter<WalletPayState> emit) async {
    emit(WalletPayInProgress());

    try {
      final results = await _jsonRepository.getSetting('wallet', event.search);

      if (results.success) {
        if (results.data.length > 0) {
          List<WalletModel> walletPayList = [];

          for (int i = 0; i < results.data.length; i++) {
            WalletModel walletModel =
                WalletModel.fromJson(json.decode(results.data[i]['body']));
            walletModel.guidfixed = results.data[i]['guidfixed'];
            walletPayList.add(walletModel);
          }

          emit(WalletPayLoadSuccess(walletPays: walletPayList));
        } else {
          emit(const WalletPayLoadFailed(message: 'WalletPay No Data Found'));
        }
      } else {
        emit(const WalletPayLoadFailed(message: 'WalletPay Not Found'));
      }
    } catch (e) {
      emit(WalletPayLoadFailed(message: e.toString()));
    }
  }

  void onWalletPaySave(
      WalletPaySave event, Emitter<WalletPayState> emit) async {
    emit(WalletPaySaveInProgress());
    try {
      final postData = {
        "code": 'wallet',
        "body": jsonEncode(event.walletModel)
      };
      await _jsonRepository.saveSetting(postData);
      emit(WalletPaySaveSuccess());
    } catch (e) {
      emit(WalletPaySaveFailed(message: e.toString()));
    }
  }

  void onWalletPayUpdate(
      WalletPayUpdate event, Emitter<WalletPayState> emit) async {
    emit(WalletPayUpdateInProgress());
    try {
      final postData = {
        "code": 'wallet',
        "body": jsonEncode(event.walletModel)
      };

      await _jsonRepository.updateSetting(event.guid, postData);
      emit(WalletPayUpdateSuccess());
    } catch (e) {
      emit(WalletPayUpdateFailed(message: e.toString()));
    }
  }

  void onWalletPayDelete(
      WalletPayDelete event, Emitter<WalletPayState> emit) async {
    emit(WalletPayDeleteInProgress());
    try {
      await _jsonRepository.deleteSetting(event.guid);

      emit(WalletPayDeleteSuccess());
    } catch (e) {
      emit(WalletPayDeleteFailed());
    }
  }

  void onWalletPayDeleteMany(
      WalletPayDeleteMany event, Emitter<WalletPayState> emit) async {
    emit(WalletPayDeleteManyInProgress());
    try {
      await _jsonRepository.deleteManySetting(event.guid);

      emit(WalletPayDeleteManySuccess());
    } catch (e) {
      emit(WalletPayDeleteManyFailed());
    }
  }
}
