import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/models/branch_model.dart';
import 'package:bclms/app/modules/branch_selection/branch_selection_controller.dart';

class BranchSelectionPage extends ConsumerWidget {
  const BranchSelectionPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(branchSelectionProvider);
    final notifier = ref.read(branchSelectionProvider.notifier);

    // Listen for auto-navigate to home
    ref.listen(branchSelectionProvider, (prev, next) {
      if (next.shouldNavigateHome && prev?.shouldNavigateHome != true) {
        context.go('/home');
      }
    });

    if (state.isLoading) {
      return const Scaffold(
        body: Center(
          child: CircularProgressIndicator(color: AppTheme.terracotta),
        ),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('เลือกสาขา'),
      ),
      body: _buildBody(state, notifier),
    );
  }

  Widget _buildBody(BranchSelectionState state, BranchSelectionNotifier notifier) {
    if (state.branches.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.business_outlined, size: 64, color: AppTheme.terracotta.withValues(alpha: 0.4)),
            const SizedBox(height: 16),
            Text(
              'ไม่พบสาขา',
              style: TextStyle(
                fontSize: 16,
                color: AppTheme.earthBrown.withValues(alpha: 0.6),
              ),
            ),
          ],
        ),
      );
    }

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 600),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: const EdgeInsets.only(bottom: 16, left: 4),
                child: Text(
                  'กรุณาเลือกสาขาที่ต้องการ',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                    color: AppTheme.earthBrown.withValues(alpha: 0.8),
                  ),
                ),
              ),
              Expanded(
                child: ListView.separated(
                  itemCount: state.branches.length,
                  separatorBuilder: (context, index) => const SizedBox(height: 10),
                  itemBuilder: (context, index) {
                    final branch = state.branches[index];
                    return _BranchCard(
                      branch: branch,
                      onTap: () => notifier.onBranchSelected(branch),
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _BranchCard extends StatelessWidget {
  final BranchModel branch;
  final VoidCallback onTap;

  const _BranchCard({required this.branch, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        splashColor: AppTheme.terracotta.withValues(alpha: 0.1),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: AppTheme.cloudBlue.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(Icons.business, color: AppTheme.cloudBlue),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Text(
                  branch.getDisplayName(),
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: AppTheme.earthBrown,
                  ),
                ),
              ),
              Icon(
                Icons.chevron_right,
                color: AppTheme.terracotta.withValues(alpha: 0.4),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
