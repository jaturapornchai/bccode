import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/modules/home/home_controller.dart';
import 'package:bclms/app/modules/home/home_page.dart';
import 'package:bclms/app/routes/app_routes.dart';

/// Tabbed Home Screen — จัดการหลาย tab เหมือน browser
/// - เริ่มต้นด้วย 1 tab (เมนูหลัก)
/// - สามารถเพิ่ม tab ใหม่ได้ด้วยปุ่ม +
/// - แต่ละ tab มี Navigator แยกกัน เก็บ state อิสระ
/// - แต่ละ tab มีปุ่มปิดได้ ยกเว้น tab แรก
class TabbedHomePage extends ConsumerStatefulWidget {
  const TabbedHomePage({super.key});

  @override
  ConsumerState<TabbedHomePage> createState() => TabbedHomePageState();
}

class TabbedHomePageState extends ConsumerState<TabbedHomePage>
    with TickerProviderStateMixin {
  late TabController _tabController;
  final List<_HomeTab> _tabs = [];
  final Map<int, GlobalKey<NavigatorState>> _navigatorKeys = {};
  final Map<int, _TabNavigatorObserver> _navigatorObservers = {};
  int _tabIdCounter = 0;

  @override
  void initState() {
    super.initState();
    _addNewTab(isInitial: true);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  /// เพิ่ม tab ใหม่ — เรียกจากภายนอกได้
  void addNewTab({String? title, Widget? customContent}) {
    _addNewTab(title: title, customContent: customContent);
  }

  void _addNewTab({
    bool isInitial = false,
    String? title,
    Widget? customContent,
  }) {
    setState(() {
      final tabId = _tabIdCounter++;
      _tabs.add(_HomeTab(
        id: tabId,
        title: title ?? 'เมนูหลัก',
        isClosable: !isInitial,
        customContent: customContent,
      ));
      _navigatorKeys[tabId] = GlobalKey<NavigatorState>();
      _navigatorObservers[tabId] = _TabNavigatorObserver(
        onRouteChanged: (route) => _updateTabTitle(tabId, route),
      );

      final oldController = isInitial ? null : _tabController;
      _tabController = TabController(
        length: _tabs.length,
        vsync: this,
        initialIndex: _tabs.length - 1,
      );
      _tabController.addListener(() {
        if (!_tabController.indexIsChanging) setState(() {});
      });
      oldController?.dispose();
    });
  }

  /// อัพเดทชื่อและ Icon ของ Tab ตาม Route ปัจจุบัน
  void _updateTabTitle(int tabId, Route<dynamic>? route) {
    final tabIndex = _tabs.indexWhere((tab) => tab.id == tabId);
    if (tabIndex == -1) return;

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      setState(() {
        final name = route?.settings.name;
        if (name == '/' || name == null || name.isEmpty) {
          _tabs[tabIndex].title = 'เมนูหลัก';
          _tabs[tabIndex].icon = Icons.grid_view_rounded;
        } else {
          final args = route?.settings.arguments;
          if (args is Map<String, dynamic> && args.containsKey('title')) {
            _tabs[tabIndex].title = args['title'] as String;
          } else {
            _tabs[tabIndex].title = 'หน้างาน';
          }
          _tabs[tabIndex].icon = Icons.description_rounded;
        }
      });
    });
  }

  /// ปิด tab
  Future<void> _closeTab(int index) async {
    if (_tabs.length <= 1 || index == 0) return;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Row(
          children: [
            Icon(Icons.warning_amber_rounded, color: Colors.orange),
            SizedBox(width: 8),
            Text('ปิด Tab'),
          ],
        ),
        content: const Text('ข้อมูลที่ยังไม่ได้บันทึกจะหายไป\nต้องการปิด Tab นี้หรือไม่?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('ยกเลิก'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('ปิด'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    setState(() {
      final tabId = _tabs[index].id;
      _tabs.removeAt(index);
      _navigatorKeys.remove(tabId);
      _navigatorObservers.remove(tabId);

      int newIndex = _tabController.index;
      if (newIndex >= _tabs.length) newIndex = _tabs.length - 1;

      final oldController = _tabController;
      _tabController = TabController(
        length: _tabs.length,
        vsync: this,
        initialIndex: newIndex,
      );
      _tabController.addListener(() {
        if (!_tabController.indexIsChanging) setState(() {});
      });
      oldController.dispose();
    });
  }

  /// ย้ายตำแหน่ง tab (drag & drop)
  void _reorderTab(int oldIndex, int newIndex) {
    if (oldIndex == newIndex) return;

    setState(() {
      if (newIndex > oldIndex) newIndex -= 1;

      final tab = _tabs.removeAt(oldIndex);
      _tabs.insert(newIndex, tab);

      final currentIndex = _tabController.index;
      final oldController = _tabController;

      int adjustedIndex = currentIndex;
      if (currentIndex == oldIndex) {
        adjustedIndex = newIndex;
      } else if (oldIndex < currentIndex && newIndex >= currentIndex) {
        adjustedIndex = currentIndex - 1;
      } else if (oldIndex > currentIndex && newIndex <= currentIndex) {
        adjustedIndex = currentIndex + 1;
      }

      _tabController = TabController(
        length: _tabs.length,
        vsync: this,
        initialIndex: adjustedIndex,
      );
      _tabController.addListener(() {
        if (!_tabController.indexIsChanging) setState(() {});
      });
      oldController.dispose();
    });
  }

  // === Build Methods ===

  Widget _buildTab(_HomeTab tab, int index) {
    return DragTarget<int>(
      onWillAcceptWithDetails: (details) => details.data != index,
      onAcceptWithDetails: (details) => _reorderTab(details.data, index),
      builder: (context, candidateData, rejectedData) {
        final isHovering = candidateData.isNotEmpty;

        return LongPressDraggable<int>(
          data: index,
          feedback: Material(
            elevation: 8,
            borderRadius: BorderRadius.circular(8),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: AppTheme.terracottaDark,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(tab.icon, size: 16, color: Colors.white),
                  const SizedBox(width: 6),
                  Text(
                    tab.title,
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ],
              ),
            ),
          ),
          childWhenDragging: Opacity(
            opacity: 0.3,
            child: _buildTabContent(tab, index),
          ),
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            decoration: BoxDecoration(
              border: isHovering
                  ? Border.all(color: Colors.white, width: 2)
                  : null,
              borderRadius: BorderRadius.circular(8),
            ),
            child: _buildTabContent(tab, index),
          ),
        );
      },
    );
  }

  Widget _buildTabContent(_HomeTab tab, int index) {
    final isActive = _tabController.index == index;

    return Tab(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: isActive ? AppTheme.terracottaDark : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
          boxShadow: isActive
              ? [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.2),
                    blurRadius: 6,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              tab.icon,
              size: 16,
              color: isActive ? Colors.white : Colors.white70,
            ),
            const SizedBox(width: 6),
            Flexible(
              child: Text(
                tab.title,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: isActive ? FontWeight.w700 : FontWeight.normal,
                  color: isActive ? Colors.white : Colors.white70,
                ),
              ),
            ),
            if (tab.isClosable) ...[
              const SizedBox(width: 6),
              InkWell(
                onTap: () => _closeTab(index),
                borderRadius: BorderRadius.circular(10),
                child: Container(
                  padding: const EdgeInsets.all(3),
                  decoration: BoxDecoration(
                    color: isActive
                        ? Colors.white.withValues(alpha: 0.25)
                        : Colors.white.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Icon(
                    Icons.close,
                    size: 14,
                    color: isActive ? Colors.white : Colors.white70,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final homeState = ref.watch(homeProvider);

    // Listen for logout → navigate to login
    ref.listen(homeProvider, (prev, next) {
      if (next.loggedOut && prev?.loggedOut != true) {
        context.go('/login');
      }
    });

    return Scaffold(
      backgroundColor: AppTheme.cream,
      appBar: PreferredSize(
        preferredSize: const Size.fromHeight(kToolbarHeight),
        child: Container(
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: [AppTheme.terracotta, AppTheme.terracottaDark],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.15),
                blurRadius: 6,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: SafeArea(
            child: Row(
              children: [
                // Tab Bar (scrollable)
                Expanded(
                  child: TabBar(
                    controller: _tabController,
                    isScrollable: true,
                    indicator: const BoxDecoration(),
                    indicatorColor: Colors.transparent,
                    labelColor: Colors.white,
                    unselectedLabelColor: Colors.white70,
                    dividerColor: Colors.transparent,
                    tabAlignment: TabAlignment.start,
                    tabs: _tabs
                        .asMap()
                        .entries
                        .map((e) => _buildTab(e.value, e.key))
                        .toList(),
                  ),
                ),
                // Action buttons
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    // ปุ่มเพิ่ม tab
                    Tooltip(
                      message: 'เปิดหน้าใหม่',
                      child: InkWell(
                        onTap: () => _addNewTab(),
                        borderRadius: BorderRadius.circular(20),
                        child: Container(
                          margin: const EdgeInsets.symmetric(horizontal: 4),
                          padding: const EdgeInsets.all(6),
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.25),
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: const Icon(
                            Icons.add,
                            size: 20,
                            color: Colors.white,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 4),
                    // Shop name
                    if (homeState.shopName.isNotEmpty)
                      _AppBarChip(
                        icon: Icons.store,
                        label: homeState.shopName,
                      ),
                    // User name + Role
                    if (homeState.userFullName.isNotEmpty)
                      _AppBarChip(
                        icon: Icons.person,
                        label: homeState.userFullName,
                        badge: homeState.roleName.isNotEmpty
                            ? homeState.roleName
                            : null,
                        badgeColor: _getRoleBadgeColor(homeState.roleName),
                      ),
                    // Logout
                    IconButton(
                      icon: const Icon(Icons.logout, color: Colors.white, size: 20),
                      tooltip: 'ออกจากระบบ',
                      onPressed: () => ref.read(homeProvider.notifier).logout(),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        physics: const NeverScrollableScrollPhysics(),
        children: _tabs.map((tab) {
          return _TabContent(
            key: ValueKey('tab_${tab.id}'),
            tabId: tab.id,
            navigatorKey: _navigatorKeys[tab.id]!,
            navigatorObserver: _navigatorObservers[tab.id]!,
            customContent: tab.customContent,
          );
        }).toList(),
      ),
    );
  }

  Color _getRoleBadgeColor(String roleName) {
    switch (roleName) {
      case 'Super Admin':
        return const Color(0xFFD32F2F);
      case 'Admin':
        return const Color(0xFFE65100);
      default:
        return Colors.grey.shade600;
    }
  }
}

// === Private Widgets ===

/// Chip ใน AppBar แสดง icon + label + optional badge
class _AppBarChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final String? badge;
  final Color? badgeColor;

  const _AppBarChip({
    required this.icon,
    required this.label,
    this.badge,
    this.badgeColor,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 3),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha: 0.2),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 14, color: Colors.white),
            const SizedBox(width: 4),
            Text(
              label,
              style: const TextStyle(
                fontSize: 12,
                color: Colors.white,
                fontWeight: FontWeight.w500,
              ),
            ),
            if (badge != null) ...[
              const SizedBox(width: 6),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: badgeColor ?? Colors.grey,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  badge!,
                  style: const TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.w600,
                    color: Colors.white,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

/// Navigator content สำหรับแต่ละ Tab
/// ใช้ AutomaticKeepAliveClientMixin เพื่อเก็บ state เมื่อสลับ tab
class _TabContent extends StatefulWidget {
  final int tabId;
  final GlobalKey<NavigatorState> navigatorKey;
  final _TabNavigatorObserver navigatorObserver;
  final Widget? customContent;

  const _TabContent({
    super.key,
    required this.tabId,
    required this.navigatorKey,
    required this.navigatorObserver,
    this.customContent,
  });

  @override
  State<_TabContent> createState() => _TabContentState();
}

class _TabContentState extends State<_TabContent>
    with AutomaticKeepAliveClientMixin {
  @override
  bool get wantKeepAlive => true;

  @override
  Widget build(BuildContext context) {
    super.build(context);

    if (widget.customContent != null) {
      return widget.customContent!;
    }

    return Navigator(
      key: widget.navigatorKey,
      observers: [widget.navigatorObserver],
      onGenerateRoute: (settings) {
        // หน้าแรก — แสดง HomePage (เมนู Grid + Bottom Nav)
        if (settings.name == '/' ||
            settings.name == null ||
            settings.name!.isEmpty) {
          return MaterialPageRoute(
            builder: (_) => const HomePage(),
            settings: settings,
          );
        }
        // หน้าอื่นๆ — ใช้ AppRoutes
        return AppRoutes.onGenerateRoute(settings);
      },
    );
  }
}

/// NavigatorObserver สำหรับติดตามการเปลี่ยนหน้าจอใน Tab
class _TabNavigatorObserver extends NavigatorObserver {
  final Function(Route<dynamic>?) onRouteChanged;

  _TabNavigatorObserver({required this.onRouteChanged});

  @override
  void didPush(Route<dynamic> route, Route<dynamic>? previousRoute) {
    onRouteChanged(route);
  }

  @override
  void didPop(Route<dynamic> route, Route<dynamic>? previousRoute) {
    onRouteChanged(previousRoute);
  }

  @override
  void didReplace({Route<dynamic>? newRoute, Route<dynamic>? oldRoute}) {
    onRouteChanged(newRoute);
  }
}

/// Model สำหรับ Tab
class _HomeTab {
  final int id;
  String title;
  IconData icon;
  final bool isClosable;
  final Widget? customContent;

  _HomeTab({
    required this.id,
    required this.title,
    this.icon = Icons.tab,
    required this.isClosable,
    this.customContent,
  });
}
