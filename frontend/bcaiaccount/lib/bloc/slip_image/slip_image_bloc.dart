import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/slip_image_model.dart';
import 'package:smlaicloud/repositories/slip_image_repository.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import '../../global.dart' as global;

part 'slip_image_event.dart';
part 'slip_image_state.dart';

/// Singleton สำหรับเก็บรูปภาพ SLIP ใน memory
/// เพื่อให้ข้อมูลไม่หายเมื่อเปลี่ยนหน้า
class SlipImageStorage {
  static final SlipImageStorage _instance = SlipImageStorage._internal();
  factory SlipImageStorage() => _instance;
  SlipImageStorage._internal();

  final Map<SlipType, List<SlipImageModel>> _images = {};

  /// Cache สำหรับเก็บ base64 data ที่โหลดแล้ว (key: filename)
  final Map<String, String> _base64Cache = {};

  /// เพิ่มรูปภาพ
  void addImage(SlipType slipType, SlipImageModel image) {
    _images[slipType] ??= [];
    _images[slipType]!.insert(0, image);
    debugPrint('[SLIP STORAGE] Added image to ${slipType.value}, total: ${_images[slipType]!.length}');
  }

  /// ลบรูปภาพ
  void removeImage(SlipType slipType, String filename) {
    _images[slipType]?.removeWhere((img) => img.filename == filename);
    _base64Cache.remove(filename);
    debugPrint('[SLIP STORAGE] Removed image $filename from ${slipType.value}');
  }

  /// ดึงรายการรูปภาพ
  List<SlipImageModel> getImages(SlipType slipType) {
    return _images[slipType] ?? [];
  }

  /// ตั้งค่ารายการรูปภาพ (แทนที่ทั้งหมด)
  void setImages(SlipType slipType, List<SlipImageModel> images) {
    _images[slipType] = List.from(images);
    debugPrint('[SLIP STORAGE] Set ${images.length} images for ${slipType.value}');
  }

  /// เก็บ base64 data ลง cache
  void cacheBase64(String filename, String base64Data) {
    _base64Cache[filename] = base64Data;
    debugPrint('[SLIP CACHE] Cached base64 for $filename (${base64Data.length} chars)');
  }

  /// ดึง base64 data จาก cache
  String? getCachedBase64(String filename) {
    return _base64Cache[filename];
  }

  /// ตรวจสอบว่ามี cache หรือไม่
  bool hasCachedBase64(String filename) {
    return _base64Cache.containsKey(filename);
  }

  /// ล้างข้อมูลทั้งหมด
  void clear() {
    _images.clear();
    _base64Cache.clear();
    debugPrint('[SLIP STORAGE] Cleared all data');
  }
}

/// BLoC สำหรับจัดการ SLIP เงินเข้า/ออก
class SlipImageBloc extends Bloc<SlipImageEvent, SlipImageState> {
  final SlipImageRepository _slipImageRepository;
  final SlipImageStorage _storage = SlipImageStorage();

  SlipImageBloc({required SlipImageRepository slipImageRepository})
      : _slipImageRepository = slipImageRepository,
        super(SlipImageInitial()) {
    on<LoadSlipImages>(_onLoadSlipImages);
    on<UploadSlipImage>(_onUploadSlipImage);
    on<DeleteSlipImage>(_onDeleteSlipImage);
    on<VerifySlipImage>(_onVerifySlipImage);
    on<ResetSlipImageState>(_onResetState);
  }

  /// โหลดรายการ SLIP
  void _onLoadSlipImages(
    LoadSlipImages event,
    Emitter<SlipImageState> emit,
  ) async {
    debugPrint('[SLIP BLOC] _onLoadSlipImages called');
    debugPrint('[SLIP BLOC] slipType: ${event.slipType.value}');
    emit(SlipImageLoading());

    // ดึงข้อมูลจาก memory
    final localList = _storage.getImages(event.slipType);
    debugPrint('[SLIP BLOC] Local images count: ${localList.length}');

    try {
      // เรียก API /image/list พร้อม Presigned URL
      // API จะส่ง url field มา ใช้ Image.network() แสดงผล (เร็วกว่า base64)
      final result = await _slipImageRepository.getSlipImageList(
        slipType: event.slipType,
        includeData: false, // ไม่ต้องส่ง base64 - ใช้ Presigned URL แทน
      );

      debugPrint('[SLIP BLOC] Load result - success: ${result.success}, total: ${result.total}');

      if (result.success && result.data.isNotEmpty) {
        debugPrint('[SLIP BLOC] Loaded ${result.data.length} SLIP images from API');
        // Debug: ตรวจสอบว่า API ส่งข้อมูล verified กลับมาหรือไม่
        for (final img in result.data) {
          debugPrint('[SLIP BLOC] Image: ${img.filename}, verified: ${img.verified}, status: ${img.verifyStatus}, isDuplicate: ${img.isDuplicate}');
        }
        // รวมรูปจาก API กับรูปที่เก็บไว้ใน storage (ไม่รวมซ้ำ)
        // และรักษาข้อมูลการตรวจสอบจาก storage ไว้ (กรณี API ยังไม่ส่งกลับมา)
        final List<SlipImageModel> mergedList = [];

        for (final apiImage in result.data) {
          // หารูปใน storage ที่มี filename เดียวกัน
          final localImage = localList.firstWhere(
            (img) => img.filename == apiImage.filename,
            orElse: () => apiImage,
          );

          // ถ้า API ส่ง verified มา ใช้ค่าจาก API
          // ถ้าไม่มี ใช้ค่าจาก storage (ที่เคยตรวจสอบไว้)
          if (apiImage.verified != null) {
            mergedList.add(apiImage);
          } else if (localImage.verified != null) {
            // ใช้ข้อมูลจาก API แต่เอาข้อมูลการตรวจสอบจาก storage
            mergedList.add(apiImage.copyWithVerification(
              verified: localImage.verified,
              verifyStatus: localImage.verifyStatus,
              isDuplicate: localImage.isDuplicate,
              transRef: localImage.transRef,
              transAmount: localImage.transAmount,
              senderName: localImage.senderName,
              receiverName: localImage.receiverName,
            ));
          } else {
            mergedList.add(apiImage);
          }
        }

        // เพิ่มรูปจาก storage ที่ไม่มีใน API (รูปที่เพิ่งอัพโหลด)
        final apiFilenames = result.data.map((img) => img.filename).toSet();
        final filteredLocalList = localList.where((img) => !apiFilenames.contains(img.filename)).toList();
        final combinedList = [...mergedList, ...filteredLocalList];

        // เก็บรูปทั้งหมดลง storage เพื่อใช้ตอนอัพโหลดเสร็จ
        _storage.setImages(event.slipType, combinedList);

        emit(SlipImageLoaded(images: combinedList, slipType: event.slipType));
      } else {
        debugPrint('[SLIP BLOC] No data from API, showing local images');
        // แสดงรูปที่เก็บไว้ใน storage
        emit(SlipImageLoaded(images: localList, slipType: event.slipType));
      }
    } on Exception catch (e) {
      debugPrint('[SLIP BLOC] Exception: $e');
      // ถ้า API ยังไม่พร้อม ให้แสดงรายการจาก storage
      debugPrint('[SLIP BLOC] Showing ${localList.length} local images');
      emit(SlipImageLoaded(images: localList, slipType: event.slipType));
    } catch (e) {
      debugPrint('[SLIP BLOC] Unknown error: $e');
      emit(SlipImageLoaded(images: localList, slipType: event.slipType));
    }
  }

  /// อัพโหลดรูปภาพ SLIP
  void _onUploadSlipImage(
    UploadSlipImage event,
    Emitter<SlipImageState> emit,
  ) async {
    debugPrint('[SLIP BLOC] _onUploadSlipImage called');
    debugPrint('[SLIP BLOC] fileName: ${event.fileName}, slipType: ${event.slipType.value}');
    emit(SlipImageUploading());

    try {
      debugPrint('[SLIP BLOC] Calling repository.uploadSlipImage...');
      final result = await _slipImageRepository.uploadSlipImage(
        imageBytes: event.imageBytes,
        fileName: event.fileName,
        slipType: event.slipType,
      );

      debugPrint('[SLIP BLOC] Upload result - success: ${result.success}, uri: ${result.uri}');

      if (result.success) {
        if (kDebugMode) {
          AppLogger.debug('Upload SLIP success: ${result.uri}');
        }

        // ใช้ filename จาก API response (backend สร้างชื่อใหม่)
        // ถ้าไม่มีให้ใช้ชื่อเดิม
        final actualFilename = result.filename ?? event.fileName;
        debugPrint('[SLIP BLOC] Actual filename from API: $actualFilename');

        // เก็บรูปที่อัพโหลดไว้ใน local เพื่อแสดงผลทันที
        final newImage = SlipImageModel(
          id: result.id ?? DateTime.now().millisecondsSinceEpoch.toString(),
          shopId: '',
          filename: actualFilename,
          originalName: event.fileName,
          contentType: 'image/jpeg',
          size: event.imageBytes.length,
          category: event.slipType.value,
          createdAt: DateTime.now(),
          imageBytes: event.imageBytes, // เก็บ bytes ไว้สำหรับแสดงผล
        );

        // เพิ่มเข้า storage (memory)
        _storage.addImage(event.slipType, newImage);

        // emit SlipImageLoaded พร้อมรูปใหม่ทันที (ไม่ต้อง reload จาก API)
        final allImages = _storage.getImages(event.slipType);
        emit(SlipImageLoaded(images: allImages, slipType: event.slipType));
      } else {
        debugPrint('[SLIP BLOC] Upload failed: ${result.message}');
        emit(SlipImageUploadFailed(message: result.message ?? 'อัพโหลดล้มเหลว'));
      }
    } on Exception catch (e) {
      debugPrint('[SLIP BLOC] Exception: $e');
      emit(SlipImageUploadFailed(message: 'อัพโหลดล้มเหลว: $e'));
    } catch (e) {
      debugPrint('[SLIP BLOC] Unknown error: $e');
      emit(SlipImageUploadFailed(message: 'อัพโหลดล้มเหลว: $e'));
    }
  }

  /// ลบรูปภาพ SLIP
  void _onDeleteSlipImage(
    DeleteSlipImage event,
    Emitter<SlipImageState> emit,
  ) async {
    debugPrint('[SLIP BLOC] _onDeleteSlipImage called: ${event.filename}');

    // ลบจาก memory ก่อนทันที (ไม่ต้องรอ API)
    _storage.removeImage(event.slipType, event.filename);

    // emit SlipImageLoaded ทันที (UI อัพเดตเลย)
    final remainingImages = _storage.getImages(event.slipType);
    emit(SlipImageLoaded(images: remainingImages, slipType: event.slipType));

    // เรียก API ลบใน background (ไม่ block UI)
    _deleteFromApiInBackground(event.filename);
  }

  /// ลบรูปจาก API ใน background (ไม่ block UI)
  void _deleteFromApiInBackground(String filename) async {
    try {
      debugPrint('[SLIP BLOC] Deleting from API in background: $filename');
      final result = await _slipImageRepository.deleteSlipImage(
        filename: filename,
      );

      if (result.success) {
        debugPrint('[SLIP BLOC] API delete success: $filename');
      } else {
        debugPrint('[SLIP BLOC] API delete failed: ${result.message}');
      }
    } catch (e) {
      // API อาจมีปัญหา แต่ลบจาก local แล้ว
      debugPrint('[SLIP BLOC] API delete error: $e');
    }
  }

  /// ตรวจสอบสลิป PromptPay
  void _onVerifySlipImage(
    VerifySlipImage event,
    Emitter<SlipImageState> emit,
  ) async {
    debugPrint('[SLIP BLOC] _onVerifySlipImage called: ${event.filename}');
    emit(SlipImageVerifying(filename: event.filename));

    try {
      debugPrint('[SLIP BLOC] Calling repository.verifySlip...');
      final result = await _slipImageRepository.verifySlip(
        imageId: event.imageId,
        filename: event.filename,
        checkDuplicate: event.checkDuplicate,
        type: event.type,
      );

      debugPrint('[SLIP BLOC] Verify result - success: ${result.success}, verified: ${result.verified}');

      if (result.success) {
        // หารูปใน storage แล้วอัพเดตสถานะ
        final images = _storage.getImages(event.slipType);
        final imageIndex = images.indexWhere((img) => img.filename == event.filename);

        if (imageIndex != -1) {
          // อัพเดตรูปด้วยข้อมูลการตรวจสอบ
          final updatedImage = images[imageIndex].copyWithVerification(
            verified: result.verified,
            verifyStatus: result.verifyStatus,
            isDuplicate: result.isDuplicate,
            transRef: result.transRef,
            transAmount: result.transAmount,
            senderName: result.senderName,
            receiverName: result.receiverName,
          );

          // อัพเดตใน storage
          images[imageIndex] = updatedImage;
          _storage.setImages(event.slipType, images);

          emit(SlipImageVerifySuccess(
            updatedImage: updatedImage,
            verifyResult: result,
          ));

          // emit SlipImageLoaded เพื่ออัพเดต UI
          emit(SlipImageLoaded(images: images, slipType: event.slipType));
        } else {
          debugPrint('[SLIP BLOC] Image not found in storage: ${event.filename}');
          emit(SlipImageVerifyFailed(
            filename: event.filename,
            message: 'ไม่พบรูปภาพใน storage',
          ));
        }
      } else {
        debugPrint('[SLIP BLOC] Verify failed: ${result.message}');
        emit(SlipImageVerifyFailed(
          filename: event.filename,
          message: result.message ?? 'ตรวจสอบล้มเหลว',
        ));
      }
    } on Exception catch (e) {
      debugPrint('[SLIP BLOC] Exception: $e');
      emit(SlipImageVerifyFailed(
        filename: event.filename,
        message: 'ตรวจสอบล้มเหลว: $e',
      ));
    } catch (e) {
      debugPrint('[SLIP BLOC] Unknown error: $e');
      emit(SlipImageVerifyFailed(
        filename: event.filename,
        message: 'ตรวจสอบล้มเหลว: $e',
      ));
    }
  }

  /// รีเซ็ต state
  void _onResetState(
    ResetSlipImageState event,
    Emitter<SlipImageState> emit,
  ) {
    emit(SlipImageInitial());
  }
}
