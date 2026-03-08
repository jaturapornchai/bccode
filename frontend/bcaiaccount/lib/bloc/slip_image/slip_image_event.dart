part of 'slip_image_bloc.dart';

/// Events สำหรับ SlipImageBloc
abstract class SlipImageEvent extends Equatable {
  const SlipImageEvent();

  @override
  List<Object?> get props => [];
}

/// โหลดรายการ SLIP
class LoadSlipImages extends SlipImageEvent {
  final SlipType slipType;
  final String? fromDate;
  final String? toDate;

  const LoadSlipImages({
    required this.slipType,
    this.fromDate,
    this.toDate,
  });

  @override
  List<Object?> get props => [slipType, fromDate, toDate];
}

/// อัพโหลดรูปภาพ SLIP
class UploadSlipImage extends SlipImageEvent {
  final Uint8List imageBytes;
  final String fileName;
  final SlipType slipType;

  const UploadSlipImage({
    required this.imageBytes,
    required this.fileName,
    required this.slipType,
  });

  @override
  List<Object?> get props => [imageBytes, fileName, slipType];
}

/// ลบรูปภาพ SLIP
class DeleteSlipImage extends SlipImageEvent {
  final String filename;
  final SlipType slipType;

  const DeleteSlipImage({
    required this.filename,
    required this.slipType,
  });

  @override
  List<Object?> get props => [filename, slipType];
}

/// รีเซ็ต state
class ResetSlipImageState extends SlipImageEvent {
  const ResetSlipImageState();
}

/// ตรวจสอบสลิป PromptPay
class VerifySlipImage extends SlipImageEvent {
  final String imageId;
  final String filename;
  final SlipType slipType;
  final bool checkDuplicate;
  final String type; // "bank" หรือ "truewallet"

  const VerifySlipImage({
    required this.imageId,
    required this.filename,
    required this.slipType,
    this.checkDuplicate = true,
    this.type = 'bank',
  });

  @override
  List<Object?> get props => [imageId, filename, slipType, checkDuplicate, type];
}
