part of 'color_bloc.dart';

abstract class ColorEvent extends Equatable {
  const ColorEvent();

  @override
  List<Object> get props => [];
}

class ColorGet extends ColorEvent {
  final String guid;

  const ColorGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ColorLoadList extends ColorEvent {
  final int limit;
  final int offset;
  final String search;

  const ColorLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class ColorDelete extends ColorEvent {
  final String guid;

  const ColorDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ColorDeleteMany extends ColorEvent {
  final List<String> guid;

  const ColorDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ColorSave extends ColorEvent {
  final ColorModel color;

  const ColorSave({
    required this.color,
  });

  @override
  List<Object> get props => [color];
}

class ColorUpdate extends ColorEvent {
  final String guid;
  final ColorModel color;

  const ColorUpdate({
    required this.guid,
    required this.color,
  });

  @override
  List<Object> get props => [color];
}
