part of 'apikey_bloc.dart';

abstract class ApiKeyState extends Equatable {
  const ApiKeyState();

  @override
  List<Object> get props => [];
}

/// ApiKeyInitial
class ApiKeyInitial extends ApiKeyState {}

class GetApiKeyInitial extends ApiKeyState {}

class GetApiKeyInProgress extends ApiKeyState {}

class GetApiKeySuccess extends ApiKeyState {
  final bool success;
  final String token;

  const GetApiKeySuccess({
    required this.success,
    required this.token,
  });

  @override
  List<Object> get props => [success, token];
}

class GetApiKeyFailed extends ApiKeyState {
  final String message;

  const GetApiKeyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

/// delete apikey
class DeleteApikeyInitial extends ApiKeyState {}

class DeleteApikeyInProgress extends ApiKeyState {}

class DeleteApikeySuccess extends ApiKeyState {}

class DeleteApikeyFailed extends ApiKeyState {
  final String message;

  const DeleteApikeyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
