import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../model/coupon_model.dart';
import '../../repositories/coupon_repository.dart';
import 'coupon_event.dart';
import 'coupon_state.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class CouponBloc extends Bloc<CouponEvent, CouponState> {
  final CouponRepository _couponRepository;

  CouponBloc({required CouponRepository couponRepository})
    : _couponRepository = couponRepository,
      super(const CouponInitial()) {
    on<GetCoupons>(_onGetCoupons);
    on<GetCouponById>(_onGetCouponById);
    on<CreateCoupon>(_onCreateCoupon);
    on<UpdateCoupon>(_onUpdateCoupon);
    on<DeleteCoupon>(_onDeleteCoupon);
    on<UploadCouponExcel>(_onUploadCouponExcel);
    on<CheckImportStatus>(_onCheckImportStatus);
  }

  Future<void> _onGetCoupons(
    GetCoupons event,
    Emitter<CouponState> emit,
  ) async {
    emit(const CouponLoading());
    try {
      final couponResponse = await _couponRepository.getCoupons();
      if (couponResponse.success == true && couponResponse.data != null) {
        emit(CouponsLoaded(coupons: couponResponse.data!));
      } else {
        emit(const CouponError(message: 'Failed to load coupons'));
      }
    } catch (e) {
      emit(CouponError(message: e.toString()));
    }
  }

  Future<void> _onGetCouponById(
    GetCouponById event,
    Emitter<CouponState> emit,
  ) async {
    emit(const CouponLoading());
    try {
      if (kDebugMode) {
        AppLogger.debug('🔍 Getting coupon by ID: ${event.id}');
      }

      final coupon = await _couponRepository.getCouponById(event.id);

      if (kDebugMode) {
        AppLogger.debug('✅ Repository returned coupon: ${coupon.toJson()}');
        AppLogger.debug('📋 Coupon guidfixed: ${coupon.guidfixed}');
        AppLogger.debug('📋 Coupon couponcode: ${coupon.couponcode}');
      }

      emit(CouponLoaded(coupon: coupon));

      if (kDebugMode) {
        AppLogger.info('🎯 Emitted CouponLoaded state');
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error in _onGetCouponById: $e');
      }
      emit(CouponError(message: e.toString()));
    }
  }

  Future<void> _onCreateCoupon(
    CreateCoupon event,
    Emitter<CouponState> emit,
  ) async {
    emit(const CouponLoading());
    try {
      final couponModel = CouponModel.fromJson(event.couponData);
      final response = await _couponRepository.createCoupon(couponModel);
      if (response.success) {
        emit(
          const CouponOperationSuccess(message: 'Coupon created successfully'),
        );
      } else {
        emit(const CouponError(message: 'Failed to create coupon'));
      }
    } catch (e) {
      emit(CouponError(message: e.toString()));
    }
  }

  Future<void> _onUpdateCoupon(
    UpdateCoupon event,
    Emitter<CouponState> emit,
  ) async {
    emit(const CouponLoading());
    try {
      final couponModel = CouponModel.fromJson(event.couponData);
      final response = await _couponRepository.updateCoupon(
        event.id,
        couponModel,
      );
      if (response.success) {
        emit(
          const CouponOperationSuccess(message: 'Coupon updated successfully'),
        );
      } else {
        emit(const CouponError(message: 'Failed to update coupon'));
      }
    } catch (e) {
      emit(CouponError(message: e.toString()));
    }
  }

  Future<void> _onDeleteCoupon(
    DeleteCoupon event,
    Emitter<CouponState> emit,
  ) async {
    emit(const CouponLoading());
    try {
      final response = await _couponRepository.deleteCoupon(event.id);
      if (response.success) {
        emit(
          const CouponOperationSuccess(message: 'Coupon deleted successfully'),
        );
      } else {
        emit(const CouponError(message: 'Failed to delete coupon'));
      }
    } catch (e) {
      emit(CouponError(message: e.toString()));
    }
  }

  Future<void> _onUploadCouponExcel(
    UploadCouponExcel event,
    Emitter<CouponState> emit,
  ) async {
    try {
      if (event.validateOnly) {
        emit(const CouponImportValidating());
        if (kDebugMode) {
          AppLogger.debug('🔍 Validating coupon Excel file...');
        }
      } else {
        emit(const CouponImportUploading());
        if (kDebugMode) {
          AppLogger.debug('📤 Uploading coupon Excel file...');
        }
      }

      final response = await _couponRepository.uploadExcelFile(
        file: event.file,
        filename: event.filename,
        validateOnly: event.validateOnly,
      );

      if (response.success && response.data.success) {
        if (event.validateOnly) {
          // Validation mode - emit validation result
          if (kDebugMode) {
            AppLogger.debug(
              '✅ Validation completed: ${response.data.totalRows} rows, ${response.data.errorCount} errors',
            );
          }
          emit(CouponImportValidated(validationResult: response.data));
        } else {
          // Import mode - start checking status
          if (kDebugMode) {
            AppLogger.debug(
              '✅ Upload completed, batch_id: ${response.data.batchId}',
            );
          }
          // Automatically check status after upload
          add(CheckImportStatus(batchId: response.data.batchId));
        }
      } else {
        emit(
          CouponImportFailed(
            message: response.data.message,
            errors: response.data.errors,
          ),
        );
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error uploading Excel: $e');
      }
      emit(CouponImportFailed(message: e.toString()));
    }
  }

  Future<void> _onCheckImportStatus(
    CheckImportStatus event,
    Emitter<CouponState> emit,
  ) async {
    try {
      if (kDebugMode) {
        AppLogger.debug('🔄 Checking import status for ${event.batchId}...');
      }

      final response = await _couponRepository.getImportStatus(event.batchId);

      if (response.success) {
        final statusData = response.data;

        if (kDebugMode) {
          AppLogger.debug(
            '📊 Status: ${statusData.status}, Progress: ${statusData.progress}%',
          );
        }

        if (statusData.isCompleted) {
          // Import completed successfully
          if (kDebugMode) {
            AppLogger.info(
              '✅ Import completed: ${statusData.successCount} success, ${statusData.errorCount} errors',
            );
          }
          emit(CouponImportCompleted(finalStatus: statusData));
        } else if (statusData.isFailed) {
          // Import failed
          if (kDebugMode) {
            AppLogger.error('❌ Import failed: ${statusData.message}');
          }
          emit(
            CouponImportFailed(
              message: statusData.message,
              errors: statusData.errors,
            ),
          );
        } else {
          // Still processing - emit status for progress update
          emit(CouponImportStatusChecking(statusData: statusData));
        }
      } else {
        emit(const CouponImportFailed(message: 'Failed to get import status'));
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error checking import status: $e');
      }
      emit(CouponImportFailed(message: e.toString()));
    }
  }
}
