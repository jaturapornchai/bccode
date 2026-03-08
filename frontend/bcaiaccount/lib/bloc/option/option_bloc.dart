import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/option_repository.dart';
import 'package:smlaicloud/model/option_model.dart';
import 'package:equatable/equatable.dart';

part 'option_event.dart';
part 'option_state.dart';

class OptionBloc extends Bloc<OptionEvent, OptionState> {
  final OptionRepository _optionRepository;

  OptionBloc({required OptionRepository optionRepository})
    : _optionRepository = optionRepository,
      super(OptionInitial()) {
    on<ListOptionLoad>(_onOptionLoad);
    on<ListOptionLoadById>(_onGetOptionId);
    on<ListOptionLoadSelect>(_onOptionLoadSelect);
    on<OptionSaved>(_onOptionSaved);
    on<OptionUpdate>(_onOptionUpdate);
    on<OptionDelete>(_onOptionDelete);
  }

  void _onOptionLoadSelect(
    ListOptionLoadSelect event,
    Emitter<OptionState> emit,
  ) async {
    emit(OptionLoadSelectInProgress());
    try {
      final result = await _optionRepository.getOptionList();

      List<OptionModel> option = (result.data as List)
          .map((option) => OptionModel.fromJson(option))
          .toList();
      emit(OptionLoadSelectSuccess(option));
    } catch (e) {
      emit(OptionLoadSelectFailure(message: e.toString()));
    }
  }

  void _onOptionLoad(ListOptionLoad event, Emitter<OptionState> emit) async {
    OptionLoadSuccess optionLoadSuccess;
    List<OptionModel> previousOption = [];
    if (state is OptionLoadSuccess) {
      optionLoadSuccess = (state as OptionLoadSuccess).copyWith();
      previousOption = optionLoadSuccess.option;
    }
    emit(OptionInProgress());

    try {
      final result = await _optionRepository.getOptionList(
        perPage: event.perPage,
        page: event.page,
        search: event.search,
      );

      if (result.success) {
        if (event.nextPage) {
          List<OptionModel> option = (result.data as List)
              .map((option) => OptionModel.fromJson(option))
              .toList();
          // print(_option);
          emit(OptionLoadSuccess(option: option, page: result.page));
        } else {
          List<OptionModel> option = (result.data as List)
              .map((option) => OptionModel.fromJson(option))
              .toList();
          // print(_option);
          previousOption.addAll(option);
          emit(OptionLoadSuccess(option: previousOption, page: result.page));
        }
      } else {
        emit(OptionLoadFailed(message: 'Option Not Found'));
      }
    } catch (e) {
      emit(OptionLoadFailed(message: e.toString()));
    }
  }

  void _onGetOptionId(
    ListOptionLoadById event,
    Emitter<OptionState> emit,
  ) async {
    emit(OptionLoadByIdInProgress());
    try {
      final result = await _optionRepository.getOptionId(event.id);

      if (result.success) {
        OptionModel option = OptionModel.fromJson(result.data);
        // // print(_option);
        emit(OptionLoadByIdLoadSuccess(option: option));
      } else {
        emit(OptionLoadByIdLoadFailed(message: 'Option Not Found'));
      }
    } catch (e) {
      emit(OptionLoadByIdLoadFailed(message: e.toString()));
    }
  }

  void _onOptionSaved(OptionSaved event, Emitter<OptionState> emit) async {
    emit(OptionSaveInProgress());
    try {
      // // print(event.inventory.toString());

      await _optionRepository.saveOption(event.option);
      // // print('Success');
      emit(OptionSaveSuccess());
    } catch (e) {
      emit(OptionSaveFailure(message: e.toString()));
    }
  }

  void _onOptionUpdate(OptionUpdate event, Emitter<OptionState> emit) async {
    emit(OptionUpdateInProgress());
    try {
      // // print(event.inventory.toString());

      await _optionRepository.updateOption(event.option);

      emit(OptionUpdateSuccess());
    } catch (e) {
      emit(OptionUpdateFailure(message: e.toString()));
    }
  }

  void _onOptionDelete(OptionDelete event, Emitter<OptionState> emit) async {
    emit(OptionDeleteInProgress());
    try {
      // // print(event.inventory.toString());

      await _optionRepository.deleteOption(event.id);

      emit(OptionDeleteSuccess());
    } catch (e) {
      emit(OptionDeleteFailure(message: e.toString()));
    }
  }
}
