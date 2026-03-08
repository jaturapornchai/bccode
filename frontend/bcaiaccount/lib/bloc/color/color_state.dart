part of 'color_bloc.dart';

abstract class ColorState extends Equatable {
  const ColorState();

  @override
  List<Object> get props => [];
}

class ColorInitial extends ColorState {}

class ColorInProgress extends ColorState {}

class ColorLoadSuccess extends ColorState {
  final List<ColorModel> colors;

  const ColorLoadSuccess({required this.colors});

  ColorLoadSuccess copyWith({
    List<ColorModel>? colors,
  }) =>
      ColorLoadSuccess(colors: colors ?? this.colors);

  @override
  List<Object> get props => [colors];
}

class ColorLoadFailed extends ColorState {
  final String message;

  const ColorLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ColorSaveInitial extends ColorState {}

class ColorSaveInProgress extends ColorState {}

class ColorSaveSuccess extends ColorState {}

class ColorSaveFailed extends ColorState {
  final String message;

  const ColorSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ColorDeleteInProgress extends ColorState {}

class ColorDeleteSuccess extends ColorState {}

class ColorDeleteFailed extends ColorState {}

class ColorDeleteManyInProgress extends ColorState {}

class ColorDeleteManySuccess extends ColorState {}

class ColorDeleteManyFailed extends ColorState {}

class ColorGetInProgress extends ColorState {}

class ColorGetSuccess extends ColorState {
  final ColorModel color;

  const ColorGetSuccess({required this.color});

  ColorGetSuccess copyWith({
    ColorModel? color,
  }) =>
      ColorGetSuccess(color: color ?? this.color);

  @override
  List<Object> get props => [color];
}

class ColorGetFailed extends ColorState {
  final String message;

  const ColorGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ColorUpdateInitial extends ColorState {}

class ColorUpdateInProgress extends ColorState {}

class ColorUpdateSuccess extends ColorState {}

class ColorUpdateFailed extends ColorState {
  final String message;

  const ColorUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
