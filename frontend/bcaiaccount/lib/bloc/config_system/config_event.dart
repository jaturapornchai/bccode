part of 'config_bloc.dart';

abstract class ConfigSystemEvent extends Equatable {
  const ConfigSystemEvent();

  @override
  List<Object> get props => [];
}

class ConfigSystemGet extends ConfigSystemEvent {
  final String guid;

  const ConfigSystemGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ConfigSystemLoad extends ConfigSystemEvent {
  const ConfigSystemLoad();

  @override
  List<Object> get props => [];
}

class ConfigSystemDelete extends ConfigSystemEvent {
  final String guid;

  const ConfigSystemDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ConfigSystemDeleteMany extends ConfigSystemEvent {
  final List<String> guid;

  const ConfigSystemDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ConfigSystemSave extends ConfigSystemEvent {
  final ConfigSystemModel data;

  const ConfigSystemSave({
    required this.data,
  });

  @override
  List<Object> get props => [ConfigSystemModel];
}

class ConfigSystemUpdate extends ConfigSystemEvent {
  final String guid;
  final ConfigSystemModel data;

  const ConfigSystemUpdate({
    required this.guid,
    required this.data,
  });

  @override
  List<Object> get props => [ConfigSystemModel];
}
