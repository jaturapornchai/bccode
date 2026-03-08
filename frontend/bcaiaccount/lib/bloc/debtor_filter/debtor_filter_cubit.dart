import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/model/debtor_model.dart';
import 'package:smlaicloud/repositories/debtor_repository.dart';

part 'debtor_filter_state.dart';

/// Cubit สำหรับจัดการรายชื่อลูกหนี้เพื่อใช้ใน filter
/// ค้นหาลูกหนี้จาก API (mainapi) โดยตรง
class DebtorFilterCubit extends Cubit<DebtorFilterState> {
  final DebtorRepository _debtorRepository;

  DebtorFilterCubit({required DebtorRepository debtorRepository})
      : _debtorRepository = debtorRepository,
        super(const DebtorFilterState());

  /// ค้นหาลูกหนี้จาก API โดยตรง
  /// [searchText] คือ keyword ที่ต้องการค้นหา (รหัส หรือ ชื่อ)
  Future<void> searchDebtors(String searchText) async {
    // ถ้ากำลัง loading อยู่ ไม่ต้อง search ซ้ำ
    if (state.isLoading) return;

    emit(state.copyWith(isLoading: true, searchText: searchText));

    try {
      // ค้นหาลูกหนี้จาก API โดยส่ง search parameter
      final result = await _debtorRepository.getDebtorList(
        limit: 50, // จำกัดผลลัพธ์ให้เหมาะสมกับการแสดงผล
        offset: 0,
        search: searchText,
      );

      if (result.success) {
        final debtors = (result.data as List)
            .map((json) => DebtorModel.fromJson(json))
            .toList();

        emit(state.copyWith(
          debtors: debtors,
          isLoading: false,
          errorMessage: null,
        ));
      } else {
        emit(state.copyWith(
          isLoading: false,
          errorMessage: 'ไม่สามารถค้นหาลูกหนี้ได้',
        ));
      }
    } catch (e) {
      emit(state.copyWith(
        isLoading: false,
        errorMessage: e.toString(),
      ));
    }
  }

  /// โหลดรายชื่อลูกหนี้เริ่มต้น (ไม่มี search)
  Future<void> loadDebtors() async {
    await searchDebtors("");
  }

  /// รีเฟรชรายชื่อลูกหนี้ (บังคับ load ใหม่)
  Future<void> refreshDebtors() async {
    emit(state.copyWith(isLoading: false)); // reset loading state
    await searchDebtors(state.searchText);
  }

  /// เลือกลูกหนี้สำหรับ filter
  void selectDebtor(String? code) {
    emit(state.copyWith(
      selectedCode: code,
      clearSelectedCode: code == null,
    ));
  }

  /// ล้างการเลือกลูกหนี้
  void clearSelection() {
    emit(state.copyWith(clearSelectedCode: true));
  }
}
