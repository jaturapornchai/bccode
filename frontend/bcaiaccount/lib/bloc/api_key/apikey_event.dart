part of 'apikey_bloc.dart';

abstract class ApiKeyEvent extends Equatable {
  const ApiKeyEvent();

  @override
  List<Object> get props => [];
}

class GetApiKey extends ApiKeyEvent {
  const GetApiKey();

  @override
  List<Object> get props => [];
}

class DeleteApikey extends ApiKeyEvent {
  final String apikey;

  const DeleteApikey({
    required this.apikey,
  });

  @override
  List<Object> get props => [apikey];
}
