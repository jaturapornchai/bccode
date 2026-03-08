part of 'debtor_filter_cubit.dart';

/// State สำหรับ DebtorFilterCubit
/// เก็บรายชื่อลูกหนี้ที่ค้นหาได้และลูกหนี้ที่เลือกอยู่
class DebtorFilterState {
  /// รายชื่อลูกหนี้ที่ค้นหาได้จาก API
  final List<DebtorModel> debtors;

  /// กำลังโหลดข้อมูลอยู่หรือไม่
  final bool isLoading;

  /// รหัสลูกหนี้ที่เลือกอยู่ (null = ไม่ได้เลือก)
  final String? selectedCode;

  /// ข้อความ error (ถ้ามี)
  final String? errorMessage;

  /// keyword ที่ใช้ค้นหาล่าสุด
  final String searchText;

  const DebtorFilterState({
    this.debtors = const [],
    this.isLoading = false,
    this.selectedCode,
    this.errorMessage,
    this.searchText = "",
  });

  /// สร้าง state ใหม่โดยอัพเดทบาง field
  DebtorFilterState copyWith({
    List<DebtorModel>? debtors,
    bool? isLoading,
    String? selectedCode,
    String? errorMessage,
    String? searchText,
    bool clearSelectedCode = false,
  }) {
    return DebtorFilterState(
      debtors: debtors ?? this.debtors,
      isLoading: isLoading ?? this.isLoading,
      selectedCode: clearSelectedCode ? null : (selectedCode ?? this.selectedCode),
      errorMessage: errorMessage,
      searchText: searchText ?? this.searchText,
    );
  }

  /// ดึงชื่อลูกหนี้จากรหัส (สำหรับแสดงผล)
  String getDebtorName(String code) {
    try {
      final debtor = debtors.firstWhere((d) => d.code == code);
      if (debtor.names.isNotEmpty) {
        return debtor.names.first.name;
      }
    } catch (_) {
      // ไม่พบลูกหนี้ในรายการ — return code เดิม
    }
    return code;
  }

  /// ดึง Map ของรหัส -> ชื่อ สำหรับแสดง dropdown
  List<Map<String, String>> get debtorOptions {
    return debtors.map((d) {
      final name = d.names.isNotEmpty ? d.names.first.name : '';
      return {
        'code': d.code,
        'name': '$name (${d.code})',
      };
    }).toList();
  }
}
