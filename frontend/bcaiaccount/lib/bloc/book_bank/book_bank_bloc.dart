import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/book_bank_repository.dart';
import 'package:smlaicloud/model/book_bank_model.dart';

part 'book_bank_event.dart';
part 'book_bank_state.dart';

class BookBankBloc extends Bloc<BookBankEvent, BookBankState> {
  final BookBankRepository _bookBankRepository;
  final JsonRepository _jsonRepository;
  BookBankBloc(
      {required BookBankRepository bookBankRepository,
      required JsonRepository jsonRepository})
      : _bookBankRepository = bookBankRepository,
        _jsonRepository = jsonRepository,
        super(BookBankInitial()) {
    on<BookBankLoadList>(onBookBankLoad);
    on<BookBankSave>(onBookBankSave);
    on<BookBankWithImageSave>(onBookBankWithImageSave);
    on<BookBankUpdate>(onBookBankUpdate);
    on<BookBankWithImageUpdate>(onBookBankWithImageUpdate);
    on<BookBankDelete>(bookBookBankDelete);
    on<BookBankDeleteMany>(bookBankDeleteMany);
    on<BookBankGet>(onBookBankGet);
  }

  void onBookBankLoad(
      BookBankLoadList event, Emitter<BookBankState> emit) async {
    emit(BookBankInProgress());

    try {
      final results = await _bookBankRepository.getBookBankList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<BookBankModel> bookBanks = (results.data as List)
            .map((bookBookBank) => BookBankModel.fromJson(bookBookBank))
            .toList();
        emit(BookBankLoadSuccess(bookBanks: bookBanks));
      } else {
        emit(const BookBankLoadFailed(message: 'BookBank Not Found'));
      }
    } catch (e) {
      emit(BookBankLoadFailed(message: e.toString()));
    }
  }

  void bookBookBankDelete(
      BookBankDelete event, Emitter<BookBankState> emit) async {
    emit(BookBankDeleteInProgress());
    try {
      await _bookBankRepository.deleteBookBank(event.guid);

      emit(BookBankDeleteSuccess());
    } catch (e) {
      // emit(BookBankDeleteFailure(message: e.toString()));
    }
  }

  void bookBankDeleteMany(
      BookBankDeleteMany event, Emitter<BookBankState> emit) async {
    emit(BookBankDeleteManyInProgress());
    try {
      await _bookBankRepository.deleteBookBankMany(event.guid);

      emit(BookBankDeleteManySuccess());
    } catch (e) {
      // emit(BookBankDeleteFailure(message: e.toString()));
    }
  }

  void onBookBankSave(BookBankSave event, Emitter<BookBankState> emit) async {
    emit(BookBankSaveInProgress());
    try {
      await _bookBankRepository.saveBookBank(event.bookBank);
      emit(BookBankSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(BookBankSaveFailed(message: error['message']));
    }
  }

  void onBookBankUpdate(
      BookBankUpdate event, Emitter<BookBankState> emit) async {
    emit(BookBankUpdateInProgress());
    try {
      await _bookBankRepository.updateBookBank(event.guid, event.bookBank);
      emit(BookBankUpdateSuccess());
    } catch (e) {
      emit(BookBankUpdateFailed(message: e.toString()));
    }
  }

  void onBookBankGet(BookBankGet event, Emitter<BookBankState> emit) async {
    emit(BookBankGetInProgress());
    try {
      final result = await _bookBankRepository.getBookBank(event.guid);
      if (result.success) {
        BookBankModel bookBank = BookBankModel.fromJson(result.data);
        emit(BookBankGetSuccess(bookBank: bookBank));
      } else {
        emit(const BookBankGetFailed(message: 'Book Bank Not Found'));
      }
    } catch (e) {
      // emit(BookBankDeleteFailure(message: e.toString()));
    }
  }

  void onBookBankWithImageSave(
      BookBankWithImageSave event, Emitter<BookBankState> emit) async {
    emit(BookBankSaveInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imageFile.isNotEmpty) {
        for (int i = 0; i < event.imageFile.length; i++) {
          if (event.imageFile[i].uri.toString() != '') {
            ApiResponse result = await _jsonRepository.uploadImage(
                event.imageFile[i], event.imageWeb[i]);
            if (result.success) {
              UploadImageModel uploadImage =
                  UploadImageModel.fromJson(result.data);
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(BookBankSaveFailed(message: result.message));
            }
          }
        }

        if (images.length == event.imageFile.length) {
          BookBankModel bookBank = event.bookBank;
          bookBank.images = images;

          await _bookBankRepository.saveBookBank(bookBank);
          emit(BookBankSaveSuccess());
        } else {
          emit(const BookBankSaveFailed(message: 'image upload failed'));
        }
      } else {
        emit(const BookBankSaveFailed(message: 'no image found'));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(BookBankSaveFailed(message: error['message']));
    }
  }

  void onBookBankWithImageUpdate(
      BookBankWithImageUpdate event, Emitter<BookBankState> emit) async {
    emit(BookBankUpdateInProgress());
    try {
      List<ImagesModel> imagesList = [];
      if (event.imagesUris.isNotEmpty) {
        for (int i = 0; i < event.imagesUris.length; i++) {
          if (event.imageWeb[i].isNotEmpty) {
            ApiResponse result = await _jsonRepository.uploadImage(
                event.imageFiles[i], event.imageWeb[i]);
            if (result.success) {
              UploadImageModel uploadImage =
                  UploadImageModel.fromJson(result.data);
              imagesList.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(BookBankUpdateFailed(message: result.message));
            }
          } else if (event.imagesUris[i].uri != '') {
            imagesList
                .add(ImagesModel(uri: event.imagesUris[i].uri, xorder: i));
          }
        }

        if (imagesList.isNotEmpty) {
          BookBankModel bookBankModel = event.bookBank;
          bookBankModel.images = imagesList;

          await _bookBankRepository.updateBookBank(event.guid, bookBankModel);
          emit(BookBankUpdateSuccess());
        } else {
          emit(const BookBankUpdateFailed(message: 'image upload failed'));
        }
      } else {
        emit(const BookBankUpdateFailed(message: 'no image found'));
      }
    } catch (e) {
      emit(BookBankUpdateFailed(message: e.toString()));
    }
  }
}
