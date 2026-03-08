part of 'promotion_bloc.dart';

abstract class PromotionEvent extends Equatable {
  const PromotionEvent();

  @override
  List<Object> get props => [];
}

class PromotionGet extends PromotionEvent {
  final String guid;

  const PromotionGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class PromotionLoadList extends PromotionEvent {
  final int limit;
  final int offset;
  final String search;

  const PromotionLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class PromotionDelete extends PromotionEvent {
  final String guid;

  const PromotionDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class PromotionDeleteMany extends PromotionEvent {
  final List<String> guid;

  const PromotionDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class PromotionSave extends PromotionEvent {
  final PromotionModel promotionModel;

  const PromotionSave({
    required this.promotionModel,
  });

  @override
  List<Object> get props => [promotionModel];
}

class PromotionUpdate extends PromotionEvent {
  final String guid;
  final PromotionModel promotionModel;

  const PromotionUpdate({
    required this.guid,
    required this.promotionModel,
  });

  @override
  List<Object> get props => [promotionModel];
}
