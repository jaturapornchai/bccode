import 'dart:convert';

import 'package:smlaicloud/repositories/apikey_repository.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

part 'apikey_event.dart';
part 'apikey_state.dart';

class ApiKeyBloc extends Bloc<ApiKeyEvent, ApiKeyState> {
  final ApiKeyRepository _apiKeyRepository;

  ApiKeyBloc({required ApiKeyRepository apiKeyRepository})
      : _apiKeyRepository = apiKeyRepository,
        super(ApiKeyInitial()) {
    on<GetApiKey>(onGetApikey);
    on<DeleteApikey>(onDeleteApikey);
  }

  void onGetApikey(GetApiKey event, Emitter<ApiKeyState> emit) async {
    emit(GetApiKeyInProgress());
    try {
      final result = await _apiKeyRepository.apiKeyService();
      if (result.success) {
        emit(GetApiKeySuccess(success: true, token: result.token!));
      } else {
        emit(const GetApiKeyFailed(message: 'Api key Not Found'));
      }
    } catch (e) {
      emit(GetApiKeyFailed(message: e.toString()));
    }
  }

  void onDeleteApikey(DeleteApikey event, Emitter<ApiKeyState> emit) async {
    emit(DeleteApikeyInProgress());
    try {
      await _apiKeyRepository.deleteApiKeyService(event.apikey);
      emit(DeleteApikeySuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DeleteApikeyFailed(message: error['message']));
    }
  }
}
