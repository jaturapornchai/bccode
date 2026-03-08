part of 'slip_image_bloc.dart';

/// States สำหรับ SlipImageBloc
abstract class SlipImageState extends Equatable {
  const SlipImageState();

  @override
  List<Object?> get props => [];
}

/// สถานะเริ่มต้น
class SlipImageInitial extends SlipImageState {}

/// กำลังโหลดรายการ
class SlipImageLoading extends SlipImageState {}

/// โหลดรายการสำเร็จ
class SlipImageLoaded extends SlipImageState {
  final List<SlipImageModel> images;
  final SlipType slipType;

  const SlipImageLoaded({
    required this.images,
    required this.slipType,
  });

  @override
  List<Object?> get props => [images, slipType];
}

/// โหลดรายการล้มเหลว
class SlipImageLoadFailed extends SlipImageState {
  final String message;

  const SlipImageLoadFailed({required this.message});

  @override
  List<Object?> get props => [message];
}

/// กำลังอัพโหลด
class SlipImageUploading extends SlipImageState {}

/// อัพโหลดสำเร็จ
class SlipImageUploadSuccess extends SlipImageState {
  final String message;

  const SlipImageUploadSuccess({this.message = 'อัพโหลดสำเร็จ'});

  @override
  List<Object?> get props => [message];
}

/// อัพโหลดล้มเหลว
class SlipImageUploadFailed extends SlipImageState {
  final String message;

  const SlipImageUploadFailed({required this.message});

  @override
  List<Object?> get props => [message];
}

/// กำลังลบ
class SlipImageDeleting extends SlipImageState {}

/// ลบสำเร็จ
class SlipImageDeleteSuccess extends SlipImageState {
  final String message;

  SlipImageDeleteSuccess({String? message}) : message = message ?? global.language('delete_success');

  @override
  List<Object?> get props => [message];
}

/// ลบล้มเหลว
class SlipImageDeleteFailed extends SlipImageState {
  final String message;

  const SlipImageDeleteFailed({required this.message});

  @override
  List<Object?> get props => [message];
}

/// กำลังตรวจสอบสลิป
class SlipImageVerifying extends SlipImageState {
  final String filename; // ชื่อไฟล์ที่กำลังตรวจสอบ

  const SlipImageVerifying({required this.filename});

  @override
  List<Object?> get props => [filename];
}

/// ตรวจสอบสลิปสำเร็จ
class SlipImageVerifySuccess extends SlipImageState {
  final SlipImageModel updatedImage; // รูปที่อัพเดตแล้ว
  final SlipVerifyResponse verifyResult; // ผลการตรวจสอบ

  const SlipImageVerifySuccess({
    required this.updatedImage,
    required this.verifyResult,
  });

  @override
  List<Object?> get props => [updatedImage, verifyResult];
}

/// ตรวจสอบสลิปล้มเหลว
class SlipImageVerifyFailed extends SlipImageState {
  final String filename;
  final String message;

  const SlipImageVerifyFailed({
    required this.filename,
    required this.message,
  });

  @override
  List<Object?> get props => [filename, message];
}
