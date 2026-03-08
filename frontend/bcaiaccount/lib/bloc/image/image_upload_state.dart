part of 'image_upload_bloc.dart';

abstract class ImageUploadState extends Equatable {
  const ImageUploadState();

  @override
  List<Object> get props => [];
}

class ImageUploadInitial extends ImageUploadState {}

class ImageUploadSaveInProgress extends ImageUploadState {}

// ignore: must_be_immutable
class ImageUploadSaveSuccess extends ImageUploadState {
  ImageUpload imageUpload;

  ImageUploadSaveSuccess({
    required this.imageUpload,
  });

  @override
  List<Object> get props => [imageUpload];
}

// ignore: must_be_immutable
class LogoUploadSaveSuccess extends ImageUploadState {
  ImageUpload imageUpload;

  LogoUploadSaveSuccess({
    required this.imageUpload,
  });

  @override
  List<Object> get props => [imageUpload];
}

class ImageUploadSaveFailure extends ImageUploadState {
  final String message;
  const ImageUploadSaveFailure({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ImportImageProductInProgress extends ImageUploadState {
  final int index;

  const ImportImageProductInProgress({
    required this.index,
  });

  @override
  List<Object> get props => [index];
}

// ignore: must_be_immutable
class ImportImageProductSuccess extends ImageUploadState {
  bool success;
  int index;

  ImportImageProductSuccess({
    required this.success,
    required this.index,
  });

  @override
  List<Object> get props => [success, index];
}

class ImportImageProductFailure extends ImageUploadState {
  final String message;
  final int index;
  const ImportImageProductFailure({
    required this.message,
    required this.index,
  });

  @override
  List<Object> get props => [message, index];
}

class ImageUploadResposneUriInProgress extends ImageUploadState {}

class ImageUploadResposneUriSaveSuccess extends ImageUploadState {
  final String uri;
  final int index;

  const ImageUploadResposneUriSaveSuccess({
    required this.uri,
    required this.index,
  });

  @override
  List<Object> get props => [uri, index];
}

class ImageUploadResposneUriSaveFailure extends ImageUploadState {
  final String message;
  const ImageUploadResposneUriSaveFailure({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class VideoUploadResposneUriInProgress extends ImageUploadState {
  final int index;

  const VideoUploadResposneUriInProgress({
    required this.index,
  });

  @override
  List<Object> get props => [index];
}

class VideoUploadResposneUriSaveSuccess extends ImageUploadState {
  final String uri;
  final int index;

  const VideoUploadResposneUriSaveSuccess({
    required this.uri,
    required this.index,
  });

  @override
  List<Object> get props => [uri, index];
}

class VideoUploadResposneUriSaveFailure extends ImageUploadState {
  final int index;
  final String message;
  const VideoUploadResposneUriSaveFailure({
    required this.index,
    required this.message,
  });

  @override
  List<Object> get props => [message, index];
}
