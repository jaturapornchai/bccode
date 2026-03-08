import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/color_repository.dart';
import 'package:smlaicloud/model/color_model.dart';

part 'color_event.dart';
part 'color_state.dart';

class ColorBloc extends Bloc<ColorEvent, ColorState> {
  final ColorRepository _colorRepository;

  ColorBloc({required ColorRepository colorRepository})
      : _colorRepository = colorRepository,
        super(ColorInitial()) {
    on<ColorLoadList>(onColorLoad);
    on<ColorSave>(onColorSave);
    on<ColorUpdate>(onColorUpdate);
    on<ColorDelete>(colorDelete);
    on<ColorDeleteMany>(colorDeleteMany);
    on<ColorGet>(onColorGet);
  }

  void onColorLoad(ColorLoadList event, Emitter<ColorState> emit) async {
    emit(ColorInProgress());

    try {
      final results = await _colorRepository.getColorList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<ColorModel> colors = (results.data as List)
            .map((color) => ColorModel.fromJson(color))
            .toList();
        emit(ColorLoadSuccess(colors: colors));
      } else {
        emit(const ColorLoadFailed(message: 'Color Not Found'));
      }
    } catch (e) {
      emit(ColorLoadFailed(message: e.toString()));
    }
  }

  void colorDelete(ColorDelete event, Emitter<ColorState> emit) async {
    emit(ColorDeleteInProgress());
    try {
      await _colorRepository.deleteColor(event.guid);

      emit(ColorDeleteSuccess());
    } catch (e) {
      // emit(ColorDeleteFailure(message: e.toString()));
    }
  }

  void colorDeleteMany(ColorDeleteMany event, Emitter<ColorState> emit) async {
    emit(ColorDeleteManyInProgress());
    try {
      await _colorRepository.deleteColorMany(event.guid);

      emit(ColorDeleteManySuccess());
    } catch (e) {
      // emit(ColorDeleteFailure(message: e.toString()));
    }
  }

  void onColorSave(ColorSave event, Emitter<ColorState> emit) async {
    emit(ColorSaveInProgress());
    try {
      await _colorRepository.saveColor(event.color);
      emit(ColorSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(ColorSaveFailed(message: error['message']));
    }
  }

  void onColorUpdate(ColorUpdate event, Emitter<ColorState> emit) async {
    emit(ColorUpdateInProgress());
    try {
      await _colorRepository.updateColor(event.guid, event.color);
      emit(ColorUpdateSuccess());
    } catch (e) {
      emit(ColorUpdateFailed(message: e.toString()));
    }
  }

  void onColorGet(ColorGet event, Emitter<ColorState> emit) async {
    emit(ColorGetInProgress());
    try {
      final result = await _colorRepository.getColor(event.guid);
      if (result.success) {
        ColorModel color = ColorModel.fromJson(result.data);
        emit(ColorGetSuccess(color: color));
      } else {
        emit(const ColorGetFailed(message: 'Color Not Found'));
      }
    } catch (e) {
      // emit(ColorDeleteFailure(message: e.toString()));
    }
  }
}
