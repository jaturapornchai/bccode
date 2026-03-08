import 'package:equatable/equatable.dart';
import '../../model/coupon_model.dart';
import '../../model/coupon_import_model.dart';

abstract class CouponState extends Equatable {
  const CouponState();

  @override
  List<Object> get props => [];
}

class CouponInitial extends CouponState {
  const CouponInitial();
}

class CouponLoading extends CouponState {
  const CouponLoading();
}

class CouponsLoaded extends CouponState {
  final List<CouponModel> coupons;

  const CouponsLoaded({required this.coupons});

  @override
  List<Object> get props => [coupons];
}

class CouponLoaded extends CouponState {
  final CouponModel coupon;

  const CouponLoaded({required this.coupon});

  @override
  List<Object> get props => [coupon];
}

class CouponOperationSuccess extends CouponState {
  final String message;

  const CouponOperationSuccess({required this.message});

  @override
  List<Object> get props => [message];
}

class CouponError extends CouponState {
  final String message;

  const CouponError({required this.message});

  @override
  List<Object> get props => [message];
}

class CouponImportValidating extends CouponState {
  const CouponImportValidating();
}

class CouponImportValidated extends CouponState {
  final CouponImportUploadData validationResult;

  const CouponImportValidated({required this.validationResult});

  @override
  List<Object> get props => [validationResult];
}

class CouponImportUploading extends CouponState {
  const CouponImportUploading();
}

class CouponImportStatusChecking extends CouponState {
  final CouponImportStatusData statusData;

  const CouponImportStatusChecking({required this.statusData});

  @override
  List<Object> get props => [statusData];
}

class CouponImportCompleted extends CouponState {
  final CouponImportStatusData finalStatus;

  const CouponImportCompleted({required this.finalStatus});

  @override
  List<Object> get props => [finalStatus];
}

class CouponImportFailed extends CouponState {
  final String message;
  final List<CouponImportError>? errors;

  const CouponImportFailed({required this.message, this.errors});

  @override
  List<Object> get props => [message, errors ?? []];
}
