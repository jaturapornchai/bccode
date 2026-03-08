part of 'creditor_filter_cubit.dart';

/// State สำหรับ CreditorFilterCubit
/// เก็บรายชื่อเจ้าหนี้ที่ค้นหาได้และเจ้าหนี้ที่เลือกอยู่
class CreditorFilterState {
  /// รายชื่อเจ้าหนี้ที่ค้นหาได้จาก API
  final List<CreditorModel> creditors;

  /// กำลังโหลดข้อมูลอยู่หรือไม่
  final bool isLoading;

  /// รหัสเจ้าหนี้ที่เลือกอยู่ (null = ไม่ได้เลือก)
  final String? selectedCode;

  /// ข้อความ error (ถ้ามี)
  final String? errorMessage;

  /// keyword ที่ใช้ค้นหาล่าสุด
  final String searchText;

  const CreditorFilterState({
    this.creditors = const [],
    this.isLoading = false,
    this.selectedCode,
    this.errorMessage,
    this.searchText = "",
  });

  /// สร้าง state ใหม่โดยอัพเดทบาง field
  CreditorFilterState copyWith({
    List<CreditorModel>? creditors,
    bool? isLoading,
    String? selectedCode,
    String? errorMessage,
    String? searchText,
    bool clearSelectedCode = false,
  }) {
    return CreditorFilterState(
      creditors: creditors ?? this.creditors,
      isLoading: isLoading ?? this.isLoading,
      selectedCode: clearSelectedCode ? null : (selectedCode ?? this.selectedCode),
      errorMessage: errorMessage,
      searchText: searchText ?? this.searchText,
    );
  }

  /// ดึงชื่อเจ้าหนี้จากรหัส (สำหรับแสดงผล)
  String getCreditorName(String code) {
    final creditor = creditors.firstWhere(
      (c) => c.code == code,
      orElse: () => CreditorModel(),
    );
    if (creditor.names.isNotEmpty) {
      return creditor.names.first.name;
    }
    return code;
  }

  /// ดึง Map ของรหัส -> ชื่อ สำหรับแสดง dropdown
  List<Map<String, String>> get creditorOptions {
    return creditors.map((c) {
      final name = c.names.isNotEmpty ? c.names.first.name : '';
      return {
        'code': c.code,
        'name': '$name (${c.code})',
      };
    }).toList();
  }
}
