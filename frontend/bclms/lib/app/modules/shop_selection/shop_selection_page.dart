import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/models/shop_model.dart';
import 'package:bclms/app/modules/shop_selection/shop_selection_controller.dart';

class ShopSelectionPage extends ConsumerWidget {
  const ShopSelectionPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(shopSelectionProvider);
    final notifier = ref.read(shopSelectionProvider.notifier);

    // Listen for navigation triggers
    ref.listen(shopSelectionProvider, (prev, next) {
      if (next.shopSelected && prev?.shopSelected != true) {
        context.go('/branch-selection');
      }
      if (next.errorMessage != null && next.errorMessage != prev?.errorMessage) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(next.errorMessage!),
            backgroundColor: Colors.red.shade700,
            behavior: SnackBarBehavior.floating,
            margin: const EdgeInsets.all(16),
          ),
        );
      }
    });

    return Scaffold(
      appBar: AppBar(
        title: const Text('เลือกร้านค้า'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: 'ออกจากระบบ',
            onPressed: () async {
              await notifier.logout();
              if (context.mounted) context.go('/login');
            },
          ),
        ],
      ),
      body: _buildBody(state, notifier),
    );
  }

  Widget _buildBody(ShopSelectionState state, ShopSelectionNotifier notifier) {
    if (state.isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: AppTheme.terracotta),
      );
    }

    if (state.shops.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.store_outlined, size: 64, color: AppTheme.terracotta.withValues(alpha: 0.4)),
            const SizedBox(height: 16),
            Text(
              'ไม่พบร้านค้า',
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
                  'กรุณาเลือกร้านค้าที่ต้องการ',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                    color: AppTheme.earthBrown.withValues(alpha: 0.8),
                  ),
                ),
              ),
              Expanded(
                child: ListView.separated(
                  itemCount: state.shops.length,
                  separatorBuilder: (context, index) => const SizedBox(height: 10),
                  itemBuilder: (context, index) {
                    final shop = state.shops[index];
                    return _ShopCard(
                      shop: shop,
                      onTap: () => notifier.onShopSelected(shop),
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

class _ShopCard extends StatelessWidget {
  final ShopModel shop;
  final VoidCallback onTap;

  const _ShopCard({required this.shop, required this.onTap});

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
                  color: AppTheme.terracotta.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(Icons.store, color: AppTheme.terracotta),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      shop.shopName.isEmpty ? shop.shopId : shop.shopName,
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: AppTheme.earthBrown,
                      ),
                    ),
                    if (shop.shopName.isNotEmpty) ...[
                      const SizedBox(height: 2),
                      Text(
                        shop.shopId,
                        style: TextStyle(
                          fontSize: 12,
                          color: AppTheme.earthBrown.withValues(alpha: 0.5),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 5),
                decoration: BoxDecoration(
                  color: shop.role >= 1
                      ? AppTheme.terracotta.withValues(alpha: 0.1)
                      : AppTheme.warmBeige,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  shop.roleName,
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    color: shop.role >= 1
                        ? AppTheme.terracotta
                        : AppTheme.earthBrown.withValues(alpha: 0.6),
                  ),
                ),
              ),
              const SizedBox(width: 8),
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
