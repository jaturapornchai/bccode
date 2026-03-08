import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/repositories/image_upload_repository.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:equatable/equatable.dart';

part 'image_upload_event.dart';
part 'image_upload_state.dart';

class ImageUploadBloc extends Bloc<ImageUploadEvent, ImageUploadState> {
  final ImageUploadRepository _imageUploadRepository;

  ImageUploadBloc({required ImageUploadRepository imageUploadRepository})
      : _imageUploadRepository = imageUploadRepository,
        super(ImageUploadInitial()) {
    on<ImageUploadSaved>(_onImageUpload);
    on<ImageUploadFileSaved>(_onImageUploadFile);
    on<LogoUploadFileSaved>(_onLogoUploadFile);
    on<ImportImageProduct>(_onImportImageProduct);
    on<ImageUploadResposneUri>(_onImageUploadResposneUri);
    on<VideoUploadResposneUri>(_onVideoUploadResposneUri);
  }
  void _onImageUpload(
      ImageUploadSaved event, Emitter<ImageUploadState> emit) async {
    emit(ImageUploadSaveInProgress());
    try {
      final result =
          await _imageUploadRepository.uploadImage(event.imageUpload);

      if (result.success) {
        ImageUpload imageUpload = ImageUpload.fromJson(result.data);
        emit(ImageUploadSaveSuccess(imageUpload: imageUpload));
      } else {
        emit(const ImageUploadSaveFailure(message: 'ImageUpload Not Found'));
      }
    } catch (e) {
      emit(ImageUploadSaveFailure(message: e.toString()));
    }
  }

  void _onImageUploadFile(
      ImageUploadFileSaved event, Emitter<ImageUploadState> emit) async {
    emit(ImageUploadSaveInProgress());
    try {
      final result = await _imageUploadRepository.uploadImageFile(
          event.imageFiles[0], event.imageWeb[0]);

      if (result.success) {
        ImageUpload imageUpload = ImageUpload.fromJson(result.data);
        emit(ImageUploadSaveSuccess(imageUpload: imageUpload));
      } else {
        emit(const ImageUploadSaveFailure(message: 'ImageUpload Not Found'));
      }
    } catch (e) {
      emit(ImageUploadSaveFailure(message: e.toString()));
    }
  }

  void _onLogoUploadFile(
      LogoUploadFileSaved event, Emitter<ImageUploadState> emit) async {
    emit(ImageUploadSaveInProgress());
    try {
      final result = await _imageUploadRepository.uploadImageFile(
          event.imageFiles[0], event.imageWeb[0]);

      if (result.success) {
        ImageUpload imageUpload = ImageUpload.fromJson(result.data);
        emit(LogoUploadSaveSuccess(imageUpload: imageUpload));
      } else {
        emit(const ImageUploadSaveFailure(message: 'ImageUpload Not Found'));
      }
    } catch (e) {
      emit(ImageUploadSaveFailure(message: e.toString()));
    }
  }

  void _onImportImageProduct(
      ImportImageProduct event, Emitter<ImageUploadState> emit) async {
    emit(ImportImageProductInProgress(index: event.index));
    try {
      final result = await _imageUploadRepository.importImageProduct(
          event.imageFiles, event.imageWeb, event.imageName);

      if (result.success) {
        emit(ImportImageProductSuccess(
            success: result.success, index: event.index));
      } else {
        emit(ImportImageProductFailure(
            message: 'Image Upload Not Found', index: event.index));
      }
    } catch (e) {
      emit(
          ImportImageProductFailure(message: e.toString(), index: event.index));
    }
  }

  void _onImageUploadResposneUri(
      ImageUploadResposneUri event, Emitter<ImageUploadState> emit) async {
    emit(ImageUploadResposneUriInProgress());
    try {
      final result = await _imageUploadRepository.imageUploadResposneUri(
          event.imageFiles, event.imageWeb, event.imageName);

      if (result.success) {
        ImageUpload imageUpload = ImageUpload.fromJson(result.data);
        emit(ImageUploadResposneUriSaveSuccess(
            uri: imageUpload.uri, index: event.index));
      } else {
        emit(const ImageUploadResposneUriSaveFailure(
            message: 'Image Upload Not Found'));
      }
    } catch (e) {
      emit(ImageUploadResposneUriSaveFailure(message: e.toString()));
    }
  }

  void _onVideoUploadResposneUri(
      VideoUploadResposneUri event, Emitter<ImageUploadState> emit) async {
    emit(VideoUploadResposneUriInProgress(index: event.index));
    try {
      final result = await _imageUploadRepository.videoUploadResponseUri(
          file: event.videoFiles,
          videoWeb: event.videoWeb,
          filename: event.videoName);

      if (result.success) {
        String videoUri = result.uri!;
        emit(VideoUploadResposneUriSaveSuccess(
            uri: videoUri, index: event.index));
      } else {
        emit(VideoUploadResposneUriSaveFailure(
            message: 'Video Upload Not Found', index: event.index));
      }
    } catch (e) {
      emit(VideoUploadResposneUriSaveFailure(
          message: e.toString(), index: event.index));
    }
  }
}
