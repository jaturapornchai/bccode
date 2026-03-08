part of 'transport_channel_bloc.dart';

abstract class TransportChannelEvent extends Equatable {
  const TransportChannelEvent();

  @override
  List<Object> get props => [];
}

class TransportChannelGet extends TransportChannelEvent {
  final String guid;

  const TransportChannelGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class TransportChannelLoadList extends TransportChannelEvent {
  final int limit;
  final int offset;
  final String search;

  const TransportChannelLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class TransportChannelDelete extends TransportChannelEvent {
  final String guid;

  const TransportChannelDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class TransportChannelDeleteMany extends TransportChannelEvent {
  final List<String> guid;

  const TransportChannelDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class TransportChannelSave extends TransportChannelEvent {
  final TransportChannelModel transportchannelmodel;

  const TransportChannelSave({
    required this.transportchannelmodel,
  });

  @override
  List<Object> get props => [transportchannelmodel];
}

class TransportChannelSaveBulk extends TransportChannelEvent {
  final List<TransportChannelModel> transportchannels;

  const TransportChannelSaveBulk({
    required this.transportchannels,
  });

  @override
  List<Object> get props => [transportchannels];
}

class TransportChannelUpdate extends TransportChannelEvent {
  final String guid;
  final TransportChannelModel transportchannelmodel;

  const TransportChannelUpdate({
    required this.guid,
    required this.transportchannelmodel,
  });

  @override
  List<Object> get props => [transportchannelmodel];
}

class TransportChannelWithImageSave extends TransportChannelEvent {
  final File imageFile;
  final TransportChannelModel transportchannel;
  final Uint8List? imageWeb;
  const TransportChannelWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.transportchannel,
  });

  @override
  List<Object> get props => [transportchannel, imageFile];
}

class TransportChannelWithImageUpdate extends TransportChannelEvent {
  final String guid;
  final TransportChannelModel transportchannel;
  final File imageFile;
  final Uint8List imageWeb;
  const TransportChannelWithImageUpdate({
    required this.guid,
    required this.imageFile,
    required this.imageWeb,
    required this.transportchannel,
  });

  @override
  List<Object> get props => [transportchannel, imageWeb];
}
