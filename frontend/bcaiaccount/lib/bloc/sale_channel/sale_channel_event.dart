part of 'sale_channel_bloc.dart';

abstract class SaleChannelEvent extends Equatable {
  const SaleChannelEvent();

  @override
  List<Object> get props => [];
}

class SaleChannelGet extends SaleChannelEvent {
  final String guid;

  const SaleChannelGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class SaleChannelLoadList extends SaleChannelEvent {
  final int limit;
  final int offset;
  final String search;

  const SaleChannelLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class SaleChannelDelete extends SaleChannelEvent {
  final String guid;

  const SaleChannelDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class SaleChannelDeleteMany extends SaleChannelEvent {
  final List<String> guid;

  const SaleChannelDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class SaleChannelSave extends SaleChannelEvent {
  final SaleChannelModel salechannelmodel;

  const SaleChannelSave({
    required this.salechannelmodel,
  });

  @override
  List<Object> get props => [salechannelmodel];
}

class SaleChannelUpdate extends SaleChannelEvent {
  final String guid;
  final SaleChannelModel salechannelmodel;

  const SaleChannelUpdate({
    required this.guid,
    required this.salechannelmodel,
  });

  @override
  List<Object> get props => [salechannelmodel];
}

class SaleChannelWithImageSave extends SaleChannelEvent {
  final File imageFile;
  final SaleChannelModel salechannel;
  final Uint8List? imageWeb;
  const SaleChannelWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.salechannel,
  });

  @override
  List<Object> get props => [salechannel, imageFile];
}

class SaleChannelWithImageUpdate extends SaleChannelEvent {
  final String guid;
  final SaleChannelModel salechannel;
  final File imageFile;
  final Uint8List imageWeb;
  const SaleChannelWithImageUpdate({
    required this.guid,
    required this.imageFile,
    required this.imageWeb,
    required this.salechannel,
  });

  @override
  List<Object> get props => [salechannel, imageWeb];
}

class SaleChannelSaveBulk extends SaleChannelEvent {
  final List<SaleChannelModel> salechannels;

  const SaleChannelSaveBulk({
    required this.salechannels,
  });

  @override
  List<Object> get props => [salechannels];
}
