import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/model/creditor_model.dart';
import 'package:smlaicloud/repositories/creditor_repository.dart';

part 'creditor_filter_state.dart';

/// Cubit สำหรับจัดการรายชื่อเจ้าหนี้เพื่อใช้ใน filter
/// ค้นหาเจ้าหนี้จาก API (PostgreSQL) โดยตรง
class CreditorFilterCubit extends Cubit<CreditorFilterState> {
  final CreditorRepository _creditorRepository;

  CreditorFilterCubit({required CreditorRepository creditorRepository})
      : _creditorRepository = creditorRepository,
        super(const CreditorFilterState());

  /// ค้นหาเจ้าหนี้จาก API (PostgreSQL) โดยตรง
  /// [searchText] คือ keyword ที่ต้องการค้นหา (รหัส หรือ ชื่อ)
  Future<void> searchCreditors(String searchText) async {
    // ถ้ากำลัง loading อยู่ ไม่ต้อง search ซ้ำ
    if (state.isLoading) return;

    emit(state.copyWith(isLoading: true, searchText: searchText));

    try {
      // ค้นหาเจ้าหนี้จาก API โดยส่ง search parameter
      final result = await _creditorRepository.getCreditorList(
        limit: 50, // จำกัดผลลัพธ์ให้เหมาะสมกับการแสดงผล
        offset: 0,
        search: searchText,
      );

      if (result.success) {
        final creditors = (result.data as List)
            .map((json) => CreditorModel.fromJson(json))
            .toList();

        emit(state.copyWith(
          creditors: creditors,
          isLoading: false,
          errorMessage: null,
        ));
      } else {
        emit(state.copyWith(
          isLoading: false,
          errorMessage: 'ไม่สามารถค้นหาเจ้าหนี้ได้',
        ));
      }
    } catch (e) {
      emit(state.copyWith(
        isLoading: false,
        errorMessage: e.toString(),
      ));
    }
  }

  /// โหลดรายชื่อเจ้าหนี้เริ่มต้น (ไม่มี search)
  Future<void> loadCreditors() async {
    await searchCreditors("");
  }

  /// รีเฟรชรายชื่อเจ้าหนี้ (บังคับ load ใหม่)
  Future<void> refreshCreditors() async {
    emit(state.copyWith(isLoading: false)); // reset loading state
    await searchCreditors(state.searchText);
  }

  /// เลือกเจ้าหนี้สำหรับ filter
  void selectCreditor(String? code) {
    emit(state.copyWith(
      selectedCode: code,
      clearSelectedCode: code == null,
    ));
  }

  /// ล้างการเลือกเจ้าหนี้
  void clearSelection() {
    emit(state.copyWith(clearSelectedCode: true));
  }
}
