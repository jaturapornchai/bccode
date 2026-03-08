part of 'pos_setting_bloc.dart';

abstract class PosSettingEvent extends Equatable {
  const PosSettingEvent();

  @override
  List<Object> get props => [];
}

class PosSettingGet extends PosSettingEvent {
  final String guid;

  const PosSettingGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class PosSettingLoadList extends PosSettingEvent {
  final int limit;
  final int offset;
  final String search;

  const PosSettingLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class PosSettingDelete extends PosSettingEvent {
  final String guid;

  const PosSettingDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class PosSettingDeleteMany extends PosSettingEvent {
  final List<String> guid;

  const PosSettingDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class PosSettingSave extends PosSettingEvent {
  final PosSettingModel posSetting;

  const PosSettingSave({
    required this.posSetting,
  });

  @override
  List<Object> get props => [posSetting];
}

class PosSettingUpdate extends PosSettingEvent {
  final String guid;
  final PosSettingModel posSetting;

  const PosSettingUpdate({
    required this.guid,
    required this.posSetting,
  });

  @override
  List<Object> get props => [posSetting];
}

class PosSettingWithImageSave extends PosSettingEvent {
  final File imageFile;
  final PosSettingModel posSetting;
  final Uint8List? imageWeb;
  const PosSettingWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.posSetting,
  });

  @override
  List<Object> get props => [posSetting, imageFile];
}

class PosSettingWithImageUpdate extends PosSettingEvent {
  final String guid;
  final PosSettingModel posSetting;
  final File imageFile;
  final Uint8List imageWeb;
  const PosSettingWithImageUpdate({
    required this.guid,
    required this.imageFile,
    required this.imageWeb,
    required this.posSetting,
  });

  @override
  List<Object> get props => [posSetting, imageWeb];
}

class GetApiKey extends PosSettingEvent {
  const GetApiKey();

  @override
  List<Object> get props => [];
}

class DeleteApikey extends PosSettingEvent {
  final String apikey;

  const DeleteApikey({
    required this.apikey,
  });

  @override
  List<Object> get props => [apikey];
}
