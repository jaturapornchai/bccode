part of 'image_upload_bloc.dart';

abstract class ImageUploadEvent extends Equatable {
  const ImageUploadEvent();

  @override
  List<Object> get props => [];
}

class ImageUploadSaved extends ImageUploadEvent {
  final ImageUpload imageUpload;

  const ImageUploadSaved({
    required this.imageUpload,
  });

  @override
  List<Object> get props => [imageUpload];
}

class ImageUploadFileSaved extends ImageUploadEvent {
  final List<File> imageFiles;
  final List<Uint8List> imageWeb;

  const ImageUploadFileSaved({
    required this.imageWeb,
    required this.imageFiles,
  });

  @override
  List<Object> get props => [imageFiles];
}

class LogoUploadFileSaved extends ImageUploadEvent {
  final List<File> imageFiles;
  final List<Uint8List> imageWeb;

  const LogoUploadFileSaved({
    required this.imageWeb,
    required this.imageFiles,
  });

  @override
  List<Object> get props => [imageFiles, imageWeb];
}

class ImportImageProduct extends ImageUploadEvent {
  final File imageFiles;
  final Uint8List imageWeb;
  final int index;
  final String imageName;

  const ImportImageProduct({
    required this.imageWeb,
    required this.imageFiles,
    required this.index,
    required this.imageName,
  });

  @override
  List<Object> get props => [imageFiles, imageWeb, index, imageName];
}

class ImageUploadResposneUri extends ImageUploadEvent {
  final File imageFiles;
  final Uint8List imageWeb;
  final int index;
  final String imageName;

  const ImageUploadResposneUri({
    required this.imageWeb,
    required this.imageFiles,
    required this.index,
    required this.imageName,
  });

  @override
  List<Object> get props => [imageWeb, imageFiles, imageName, index];
}

class VideoUploadResposneUri extends ImageUploadEvent {
  final File videoFiles;
  final Uint8List videoWeb;
  final int index;
  final String videoName;

  const VideoUploadResposneUri({
    required this.videoFiles,
    required this.videoWeb,
    required this.index,
    required this.videoName,
  });

  @override
  List<Object> get props => [videoFiles, videoWeb, videoName, index];
}
