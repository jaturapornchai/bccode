import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/models/branch_model.dart';
import 'package:bclms/app/providers/service_providers.dart';

class BranchSelectionState {
  final List<BranchModel> branches;
  final bool isLoading;
  final bool shouldNavigateHome;

  const BranchSelectionState({
    this.branches = const [],
    this.isLoading = true,
    this.shouldNavigateHome = false,
  });

  BranchSelectionState copyWith({
    List<BranchModel>? branches,
    bool? isLoading,
    bool? shouldNavigateHome,
  }) {
    return BranchSelectionState(
      branches: branches ?? this.branches,
      isLoading: isLoading ?? this.isLoading,
      shouldNavigateHome: shouldNavigateHome ?? this.shouldNavigateHome,
    );
  }
}

class BranchSelectionNotifier extends Notifier<BranchSelectionState> {
  @override
  BranchSelectionState build() {
    Future.microtask(() => _loadBranches());
    return const BranchSelectionState();
  }

  Future<void> _loadBranches() async {
    state = state.copyWith(isLoading: true);
    try {
      final authService = ref.read(authServiceProvider);
      final result = await authService.listBranches();

      if (result.isEmpty) {
        AppLogger.info('ไม่พบสาขา — ข้ามไปหน้า Home');
        state = state.copyWith(isLoading: false, shouldNavigateHome: true);
        return;
      }

      if (result.length == 1) {
        AppLogger.info('มีสาขาเดียว — เลือกอัตโนมัติ');
        await _selectBranch(result.first);
        return;
      }

      state = state.copyWith(branches: result, isLoading: false);
    } catch (e) {
      AppLogger.warning('โหลดรายการสาขาล้มเหลว — ข้ามไปหน้า Home');
      state = state.copyWith(isLoading: false, shouldNavigateHome: true);
    }
  }

  Future<void> onBranchSelected(BranchModel branch) async {
    state = state.copyWith(isLoading: true);
    await _selectBranch(branch);
  }

  Future<void> _selectBranch(BranchModel branch) async {
    final storage = ref.read(storageServiceProvider);
    await storage.saveBranchGuidFixed(branch.guidFixed);
    AppLogger.info('เลือกสาขา: ${branch.getDisplayName()}');
    state = state.copyWith(isLoading: false, shouldNavigateHome: true);
  }
}

final branchSelectionProvider =
    NotifierProvider<BranchSelectionNotifier, BranchSelectionState>(
  BranchSelectionNotifier.new,
);
