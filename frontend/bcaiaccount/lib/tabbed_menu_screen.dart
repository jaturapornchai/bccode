import 'package:flutter/material.dart';
import 'package:smlaicloud/menu_screen.dart';
import 'package:smlaicloud/app_routes.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';
import 'package:smlaicloud/screens/dashboard/dashboard_screen.dart';
import 'package:smlaicloud/screens/chatbot/chatbot_overlay.dart';
import 'package:smlaicloud/screens/transaction/ai_analysis_screen.dart';
import 'package:smlaicloud/widgets/display_settings_dialog.dart';
import 'package:smlaicloud/modules/cart-system/cartsystem.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/flavors.dart';

/// Tabbed Menu Screen - จัดการหลาย tab ของ MenuScreen
/// - เริ่มต้นด้วย 1 tab
/// - สามารถเพิ่ม tab ใหม่ได้ด้วยปุ่ม +
/// - แต่ละ tab มีปุ่มปิดได้ ยกเว้น tab แรก
class TabbedMenuScreen extends StatefulWidget {
  const TabbedMenuScreen({super.key});

  /// GlobalKey สำหรับเข้าถึง TabbedMenuScreenState จากที่ไหนก็ได้
  static final GlobalKey<TabbedMenuScreenState> globalKey =
      GlobalKey<TabbedMenuScreenState>();

  @override
  State<TabbedMenuScreen> createState() => TabbedMenuScreenState();
}

class TabbedMenuScreenState extends State<TabbedMenuScreen>
    with TickerProviderStateMixin {
  late TabController _tabController;
  final List<MenuTab> _tabs = [];
  final Map<int, GlobalKey<NavigatorState>> _navigatorKeys = {};
  final Map<int, TabNavigatorObserver> _navigatorObservers = {};
  int _tabIdCounter = 0;

  // สีสำหรับ Tab Bar - ใช้สีจาก Flavor
  Color get primaryColor => F.primaryGradientColor;
  Color get primaryDarkColor => F.secondaryGradientColor;
  final Color surfaceColor = const Color(0xFFFFFFFF);
  final Color backgroundColor = const Color(0xFFF5F5F5);

  // Chatbot overlay state
  bool _showChatbot = false;
  double _chatbotWidth = 400; // ความกว้าง chatbot panel (ปรับได้)

  /// อัพเดทชื่อและ Icon ของ Tab ตาม Route ปัจจุบัน
  void _updateTabTitle(int tabId, Route<dynamic>? route) {
    final tabIndex = _tabs.indexWhere((tab) => tab.id == tabId);
    if (tabIndex != -1) {
      // ใช้ addPostFrameCallback เพื่อไม่ให้เรียก setState ระหว่าง build
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          setState(() {
            _tabs[tabIndex].title = getRouteTitle(route);
            _tabs[tabIndex].icon = getRouteIcon(route);
          });
        }
      });
    }
  }

  @override
  void initState() {
    super.initState();
    // สร้าง tab แรก
    _addNewTab(isInitial: true);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  /// เพิ่ม Transaction tab ใหม่ (เรียกจาก cart system)
  void addTransactionTab({required String title, required Widget screen}) {
    _addNewTab(title: title, customContent: screen);
  }

  /// เพิ่ม tab ใหม่
  void _addNewTab({
    bool isInitial = false,
    String? title,
    Widget? customContent,
  }) {
    setState(() {
      final tabId = _tabIdCounter;
      final newTab = MenuTab(
        id: tabId,
        title: title ?? global.language("main_menu"),
        isClosable: !isInitial,
        customContent: customContent,
      );
      _tabs.add(newTab);
      _navigatorKeys[tabId] = GlobalKey<NavigatorState>();

      // สร้าง observer สำหรับ tab นี้
      _navigatorObservers[tabId] = TabNavigatorObserver(
        onRouteChanged: (route) => _updateTabTitle(tabId, route),
      );

      _tabIdCounter++;

      // สร้าง TabController ใหม่
      TabController? oldController;
      if (!isInitial) {
        oldController = _tabController;
      }

      _tabController = TabController(
        length: _tabs.length,
        vsync: this,
        initialIndex: _tabs.length - 1, // ไปที่ tab ใหม่
      );

      // เพิ่ม listener เพื่อ rebuild UI เมื่อเปลี่ยน tab
      _tabController.addListener(() {
        if (!_tabController.indexIsChanging) {
          setState(() {});
        }
      });

      // Dispose controller เดิม (ยกเว้นครั้งแรก)
      if (oldController != null) {
        oldController.dispose();
      }
    });
  }

  /// ปิด tab
  Future<void> _closeTab(int index) async {
    if (_tabs.length <= 1 || index == 0) {
      // ไม่ให้ปิด tab แรก หรือถ้าเหลือ tab เดียว
      return;
    }

    // แสดง dialog ยืนยัน
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(global.language('confirm_close_tab_title')),
          content: Text(global.language('confirm_close_tab_message')),
          actions: <Widget>[
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: Text(global.language('cancel')),
            ),
            TextButton(
              onPressed: () => Navigator.of(context).pop(true),
              style: TextButton.styleFrom(foregroundColor: Colors.red),
              child: Text(global.language('confirm')),
            ),
          ],
        );
      },
    );

    // ถ้ายืนยัน ให้ปิด tab
    if (confirmed != true) {
      return;
    }

    setState(() {
      final tabId = _tabs[index].id;
      _tabs.removeAt(index);
      _navigatorKeys.remove(tabId);

      // ปรับ index ของ tab ที่จะเปิด
      int newIndex = _tabController.index;
      if (newIndex >= _tabs.length) {
        newIndex = _tabs.length - 1;
      }

      // สร้าง TabController ใหม่
      final oldController = _tabController;
      _tabController = TabController(
        length: _tabs.length,
        vsync: this,
        initialIndex: newIndex,
      );

      // เพิ่ม listener เพื่อ rebuild UI เมื่อเปลี่ยน tab
      _tabController.addListener(() {
        if (!_tabController.indexIsChanging) {
          setState(() {});
        }
      });

      oldController.dispose();
    });
  }

  /// ย้ายตำแหน่ง tab
  void _reorderTab(int oldIndex, int newIndex) {
    if (oldIndex == newIndex) return;

    setState(() {
      // ถ้า newIndex มากกว่า ให้ลบ 1 (เพราะตัวที่ลบออกจะทำให้ index เลื่อน)
      if (newIndex > oldIndex) {
        newIndex -= 1;
      }

      // เก็บ tab และ navigator key ที่จะย้าย
      final tab = _tabs.removeAt(oldIndex);
      _tabs.insert(newIndex, tab);

      // อัปเดต TabController
      final currentIndex = _tabController.index;
      final oldController = _tabController;

      // คำนวณ index ใหม่
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

      // เพิ่ม listener เพื่อ rebuild UI เมื่อเปลี่ยน tab
      _tabController.addListener(() {
        if (!_tabController.indexIsChanging) {
          setState(() {});
        }
      });

      oldController.dispose();
    });
  }

  /// สร้าง Tab widget
  Widget _buildTab(MenuTab tab, int index) {
    return DragTarget<int>(
      onWillAcceptWithDetails: (details) {
        return details.data != index;
      },
      onAcceptWithDetails: (details) {
        _reorderTab(details.data, index);
      },
      builder: (context, candidateData, rejectedData) {
        final bool isHovering = candidateData.isNotEmpty;

        return LongPressDraggable<int>(
          data: index,
          feedback: Material(
            elevation: 8,
            borderRadius: BorderRadius.circular(8),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: primaryColor.withOpacity(0.9),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(tab.icon, size: 16, color: Colors.white),
                  const SizedBox(width: 8),
                  Text(
                    tab.title,
                    style: const TextStyle(
                      fontSize: 14,
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
                  ? Border.all(color: primaryColor, width: 2)
                  : null,
              borderRadius: BorderRadius.circular(8),
            ),
            child: _buildTabContent(tab, index),
          ),
        );
      },
    );
  }

  /// สร้างเนื้อหาของ Tab
  Widget _buildTabContent(MenuTab tab, int index) {
    final bool isActive = _tabController.index == index;

    return Tab(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        decoration: BoxDecoration(
          color: isActive ? primaryColor : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
          boxShadow: isActive
              ? [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.2),
                    blurRadius: 8,
                    offset: const Offset(0, 3),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Icon สำหรับ tab - เปลี่ยนตาม route
            Icon(
              tab.icon,
              size: 18,
              color: isActive ? Colors.white : Colors.grey.shade600,
            ),
            const SizedBox(width: 8),
            // ชื่อ tab
            Flexible(
              child: Text(
                tab.title,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: isActive ? FontWeight.w800 : FontWeight.normal,
                  color: isActive ? Colors.white : Colors.grey.shade700,
                  letterSpacing: 0.3,
                ),
              ),
            ),
            // ปุ่มปิด (ถ้าไม่ใช่ tab แรก)
            if (tab.isClosable) ...[
              const SizedBox(width: 8),
              InkWell(
                onTap: () => _closeTab(index),
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  padding: const EdgeInsets.all(4),
                  decoration: BoxDecoration(
                    color: isActive
                        ? Colors.white.withOpacity(0.25)
                        : Colors.grey.shade200,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(
                    Icons.close,
                    size: 16,
                    color: isActive ? Colors.white : Colors.grey.shade700,
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
    return BlocListener<LoginBloc, LoginState>(
      listener: (context, state) {
        if (state is LogoutSuccess) {
          Navigator.pushNamedAndRemoveUntil(
            context,
            '/login_screen',
            (route) => false,
          );
        }
      },
      child: Scaffold(
        backgroundColor: backgroundColor,
        appBar: PreferredSize(
          preferredSize: const Size.fromHeight(kToolbarHeight),
          child: Container(
            decoration: BoxDecoration(
              color: Colors.grey.shade100,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.1),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: SafeArea(
              child: Row(
                children: [
                  // Tab Bar
                  Expanded(
                    child: TabBar(
                      controller: _tabController,
                      isScrollable: true,
                      indicator: const BoxDecoration(), // ปิด indicator
                      indicatorColor: Colors.transparent,
                      labelColor: Colors.white, // สีตัวอักษร tab ที่ active
                      unselectedLabelColor:
                          Colors.grey.shade700, // สีตัวอักษร tab ที่ไม่ active
                      dividerColor: Colors.transparent,
                      tabAlignment: TabAlignment.start,
                      tabs: _tabs
                          .asMap()
                          .entries
                          .map((entry) => _buildTab(entry.value, entry.key))
                          .toList(),
                    ),
                  ),
                  // ปุ่มด้านขวา (Dashboard, Switch Shop, Logout, Language, Add Tab)
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      // ปุ่มเพิ่ม tab
                      Container(
                        margin: EdgeInsets.only(right: 8),
                        child: Tooltip(
                          message: global.language("new_work_page"),
                          child: InkWell(
                            onTap: () => _addNewTab(),
                            borderRadius: BorderRadius.circular(20),
                            child: Container(
                              padding: const EdgeInsets.all(8),
                              decoration: BoxDecoration(
                                color: primaryColor,
                                borderRadius: BorderRadius.circular(20),
                                boxShadow: [
                                  BoxShadow(
                                    color: primaryColor.withOpacity(0.3),
                                    blurRadius: 4,
                                    offset: const Offset(0, 2),
                                  ),
                                ],
                              ),
                              child: const Icon(
                                Icons.add,
                                size: 20,
                                color: Colors.white,
                              ),
                            ),
                          ),
                        ),
                      ),
                      SizedBox(width: 8),
                      // ปุ่ม Cart System
                      IconButton(
                        icon: const Icon(
                          Icons.shopping_cart,
                          color: Colors.orange,
                          shadows: [
                            Shadow(
                              offset: Offset(1.0, 1.0),
                              blurRadius: 3.0,
                              color: Color.fromARGB(128, 0, 0, 0),
                            ),
                          ],
                        ),
                        onPressed: () {
                          _addNewTab(
                            title: global.language("cart_system"),
                            customContent: const CartSystemScreen(),
                          );
                        },
                        tooltip: global.language("cart_system"),
                      ),
                      SizedBox(width: 8),
                      // ปุ่ม AI Chat
                      IconButton(
                        icon: const Icon(
                          Icons.smart_toy, // AI icon
                          shadows: [
                            Shadow(
                              offset: Offset(1.0, 1.0),
                              blurRadius: 3.0,
                              color: Color.fromARGB(128, 0, 0, 0),
                            ),
                          ],
                        ),
                        onPressed: () {
                          // Toggle chatbot overlay
                          setState(() {
                            _showChatbot = !_showChatbot;
                          });
                        },
                        tooltip: global.language('ai_chat_assistant'),
                      ),
                      SizedBox(width: 8),
                      // ปุ่ม AI Analysis - วิเคราะห์เอกสารด้วย AI
                      IconButton(
                        icon: const Icon(
                          Icons.auto_awesome, // AI Analysis icon
                          color: Colors.amber,
                          shadows: [
                            Shadow(
                              offset: Offset(1.0, 1.0),
                              blurRadius: 3.0,
                              color: Color.fromARGB(128, 0, 0, 0),
                            ),
                          ],
                        ),
                        onPressed: () {
                          // เปิด AI Analysis เป็น tab ใหม่
                          _addNewTab(
                            title: global.language('ai_analysis'),
                            customContent: const AiAnalysisScreen(),
                          );
                        },
                        tooltip: global.language('ai_analysis'),
                      ),
                      SizedBox(width: 8),
                      // ปุ่ม Display Settings - ตั้งค่า Font และ Zoom
                      IconButton(
                        icon: const Icon(
                          Icons.settings,
                          shadows: [
                            Shadow(
                              offset: Offset(1.0, 1.0),
                              blurRadius: 3.0,
                              color: Color.fromARGB(128, 0, 0, 0),
                            ),
                          ],
                        ),
                        onPressed: () {
                          showDisplaySettingsDialog(
                            context,
                            onSettingsChanged: () {
                              // Rebuild UI เมื่อตั้งค่าเปลี่ยน
                              setState(() {});
                            },
                          );
                        },
                        tooltip: global.language('display_settings'),
                      ),
                      SizedBox(width: 8),
                      // ปุ่ม Dashboard
                      IconButton(
                        icon: const Icon(
                          Icons.dashboard,
                          shadows: [
                            Shadow(
                              offset: Offset(1.0, 1.0),
                              blurRadius: 3.0,
                              color: Color.fromARGB(128, 0, 0, 0),
                            ),
                          ],
                        ),
                        onPressed: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) => const DashBoardScreen(),
                            ),
                          );
                        },
                        tooltip: 'Dashboard',
                      ),
                      // ปุ่มสลับร้าน
                      IconButton(
                        icon: const Icon(
                          Icons.swap_vert,
                          shadows: [
                            Shadow(
                              offset: Offset(1.0, 1.0),
                              blurRadius: 3.0,
                              color: Color.fromARGB(128, 0, 0, 0),
                            ),
                          ],
                        ),
                        onPressed: () async {
                          // แสดง confirmation dialog
                          final confirmed = await showDialog<bool>(
                            context: context,
                            builder: (BuildContext context) {
                              return AlertDialog(
                                title: Row(
                                  children: [
                                    const Icon(
                                      Icons.warning,
                                      color: Colors.orange,
                                    ),
                                    SizedBox(width: 10),
                                    Text(
                                      global.language("switch_shop_confirm"),
                                    ),
                                  ],
                                ),
                                content: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      global.language("warning"),
                                      style: const TextStyle(
                                        fontSize: 16,
                                        fontWeight: FontWeight.bold,
                                        color: Colors.red,
                                      ),
                                    ),
                                    SizedBox(height: 12),
                                    Text(
                                      global.language("close_all_tabs"),
                                      style: TextStyle(
                                        fontSize: 14,
                                        color: Colors.grey[700],
                                      ),
                                    ),
                                    SizedBox(height: 8),
                                    Text(
                                      global.language("work_data_lost"),
                                      style: TextStyle(
                                        fontSize: 14,
                                        color: Colors.grey[700],
                                      ),
                                    ),
                                    SizedBox(height: 16),
                                    Text(
                                      global.language("confirm_switch_shop"),
                                      style: const TextStyle(
                                        fontSize: 14,
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                  ],
                                ),
                                actions: [
                                  TextButton(
                                    onPressed: () =>
                                        Navigator.of(context).pop(false),
                                    child: Text(global.language("cancel")),
                                  ),
                                  ElevatedButton(
                                    onPressed: () =>
                                        Navigator.of(context).pop(true),
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: Colors.orange,
                                      foregroundColor: Colors.white,
                                    ),
                                    child: Text(global.language("switch_shop")),
                                  ),
                                ],
                              );
                            },
                          );

                          // ถ้ายืนยัน ให้ปิด tab ทั้งหมดและไปหน้าสลับร้าน
                          if (confirmed == true && mounted) {
                            // Navigate ไปหน้า select_shop_screen แบบ replace ทั้งหมด
                            Navigator.of(
                              context,
                              rootNavigator: true,
                            ).pushReplacementNamed('/select_shop_screen');
                          }
                        },
                        tooltip: global.language("switch_shop"),
                      ),
                      // ปุ่ม Logout
                      IconButton(
                        icon: const Icon(
                          Icons.logout,
                          shadows: [
                            Shadow(
                              offset: Offset(1.0, 1.0),
                              blurRadius: 3.0,
                              color: Color.fromARGB(128, 0, 0, 0),
                            ),
                          ],
                        ),
                        onPressed: () async {
                          final confirmed = await showDialog<bool>(
                            context: context,
                            builder: (BuildContext context) {
                              return AlertDialog(
                                title: Row(
                                  children: [
                                    const Icon(
                                      Icons.logout,
                                      color: Colors.orange,
                                    ),
                                    SizedBox(width: 10),
                                    Text(global.language("logout_confirm")),
                                  ],
                                ),
                                content: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      global.language("warning"),
                                      style: const TextStyle(
                                        fontSize: 16,
                                        fontWeight: FontWeight.bold,
                                        color: Colors.red,
                                      ),
                                    ),
                                    SizedBox(height: 12),
                                    Text(
                                      global.language("close_all_tabs"),
                                      style: TextStyle(
                                        fontSize: 14,
                                        color: Colors.grey[700],
                                      ),
                                    ),
                                    SizedBox(height: 8),
                                    Text(
                                      global.language("work_data_lost"),
                                      style: TextStyle(
                                        fontSize: 14,
                                        color: Colors.grey[700],
                                      ),
                                    ),
                                    SizedBox(height: 16),
                                    Text(
                                      global.language("confirm_logout"),
                                      style: const TextStyle(
                                        fontSize: 14,
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                  ],
                                ),
                                actions: [
                                  TextButton(
                                    onPressed: () =>
                                        Navigator.of(context).pop(false),
                                    child: Text(global.language("cancel")),
                                  ),
                                  ElevatedButton(
                                    onPressed: () =>
                                        Navigator.of(context).pop(true),
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: Colors.red,
                                      foregroundColor: Colors.white,
                                    ),
                                    child: Text(global.language("logout")),
                                  ),
                                ],
                              );
                            },
                          );

                          if (confirmed == true && mounted) {
                            // Logout - BlocListener จะจัดการ navigation ให้
                            context.read<LoginBloc>().add(const Logout());
                          }
                        },
                        tooltip: global.language("logout"),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
        body: Row(
          children: [
            // ซ้าย: เนื้อหาหลัก
            Expanded(
              child: TabBarView(
                controller: _tabController,
                physics: const NeverScrollableScrollPhysics(),
                children: _tabs.map((tab) {
                  return _TabNavigator(
                    key: ValueKey('tab_navigator_${tab.id}'),
                    tabId: tab.id,
                    navigatorKey: _navigatorKeys[tab.id]!,
                    navigatorObserver: _navigatorObservers[tab.id]!,
                    customContent: tab.customContent,
                  );
                }).toList(),
              ),
            ),
            // Divider ลากได้ (แสดงเมื่อเปิด chatbot)
            if (_showChatbot)
              GestureDetector(
                behavior: HitTestBehavior.opaque,
                onPanUpdate: (details) {
                  setState(() {
                    final screenWidth = MediaQuery.of(context).size.width;
                    _chatbotWidth = (_chatbotWidth - details.delta.dx)
                        .clamp(280.0, screenWidth * 0.7);
                  });
                },
                child: MouseRegion(
                  cursor: SystemMouseCursors.resizeLeftRight,
                  child: SizedBox(
                    width: 8,
                    child: Container(
                      color: Colors.grey.shade200,
                      child: Center(
                        child: Container(
                          width: 4,
                          height: 48,
                          decoration: BoxDecoration(
                            color: Colors.grey.shade400,
                            borderRadius: BorderRadius.circular(2),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            // ขวา: Chat panel
            if (_showChatbot)
              SizedBox(
                width: _chatbotWidth,
                child: ChatbotOverlay(
                  isEmbedded: true,
                  onClose: () {
                    setState(() {
                      _showChatbot = false;
                    });
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }
}

/// Widget สำหรับแต่ละ Tab ที่มี Navigator แยกกัน
/// ใช้ AutomaticKeepAliveClientMixin เพื่อเก็บ state ไว้เมื่อสลับ tab
class _TabNavigator extends StatefulWidget {
  final int tabId;
  final GlobalKey<NavigatorState> navigatorKey;
  final TabNavigatorObserver navigatorObserver;
  final Widget? customContent;

  const _TabNavigator({
    super.key,
    required this.tabId,
    required this.navigatorKey,
    required this.navigatorObserver,
    this.customContent,
  });

  @override
  State<_TabNavigator> createState() => _TabNavigatorState();
}

class _TabNavigatorState extends State<_TabNavigator>
    with AutomaticKeepAliveClientMixin {
  @override
  bool get wantKeepAlive => true; // เก็บ state ไว้เมื่อสลับ tab

  @override
  Widget build(BuildContext context) {
    super.build(context); // ต้องเรียก super.build เพื่อให้ KeepAlive ทำงาน

    // ถ้ามี customContent ให้แสดงตรงๆ ไม่ต้องใช้ Navigator
    if (widget.customContent != null) {
      return widget.customContent!;
    }

    return Navigator(
      key: widget.navigatorKey,
      observers: [widget.navigatorObserver],
      onGenerateRoute: (settings) {
        // ถ้าเป็นหน้าแรก (/) ให้แสดง MenuScreen
        if (settings.name == '/' ||
            settings.name == null ||
            settings.name!.isEmpty) {
          return MaterialPageRoute(
            builder: (context) =>
                MenuScreen(key: ValueKey('menu_${widget.tabId}')),
            settings: settings,
          );
        }

        // หน้าอื่นๆ ใช้ AppRoutes
        return AppRoutes.onGenerateRoute(settings);
      },
    );
  }
}

/// NavigatorObserver สำหรับติดตามการเปลี่ยนหน้าจอ
class TabNavigatorObserver extends NavigatorObserver {
  final Function(Route<dynamic>?) onRouteChanged;

  TabNavigatorObserver({required this.onRouteChanged});

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

/// แปลง route name เป็นชื่อตามภาษา (fallback สำหรับกรณีที่ไม่มี title จาก navigation arguments)
String getRouteTitle(Route<dynamic>? route) {
  if (route == null) return global.language("work_page");

  // ✅ ตรวจสอบ arguments ก่อน เพื่อใช้ชื่อจาก menu label
  if (route.settings.arguments != null) {
    if (route.settings.arguments is Map) {
      final args = route.settings.arguments as Map;
      if (args.containsKey('title') && args['title'] is String) {
        return args['title'] as String;
      }
    }
  }

  final routeName = route.settings.name ?? '/';

  // Map route names เป็นชื่อภาษาไทย (fallback)
  final Map<String, String> routeTitles = {
    '/': global.language('main_menu'),
    '/menu': global.language('main_menu'),
    '/login_screen': global.language('login'),
    '/login_screen_shop': global.language('select_shop'),
    '/company': global.language('company'),
    '/product': global.language('product'),
    '/product_barcode': global.language('product_barcode'),
    '/productunit': global.language('product_unit'),
    '/productgroup': global.language('product_group'),
    '/productcategorylist': global.language('product_category_list'),
    '/branch': global.language('branch'),
    '/department': global.language('department'),
    '/employee': global.language('employee'),
    '/formdesign': global.language('form_design'),
    '/user': global.language('user'),
    '/creditor': global.language('creditor'),
    '/creditorgroup': global.language('creditor_group'),
    '/debtor': global.language('debtor'),
    '/debtorgroup': global.language('debtor_group'),
    '/bank': global.language('bank'),
    '/possetting': global.language('pos_setting'),
    '/ordersetting': global.language('order_setting'),
    '/ordertemplatsetting': global.language('order_template_setting'),
    '/posmedia': global.language('pos_media'),
    '/docformat': global.language('doc_format'),
    '/billdesign': global.language('bill_design'),
    '/qrprovider': global.language('qr_payment'),
    '/config_system': global.language('system_setting'),
    '/product_dimension': global.language('product_dimension'),
    '/product_bom': global.language('product_bom'),
    '/promotion_screen': global.language('promotion'),
    '/product_barcode_shelf': global.language('barcode_shelf'),
    '/price_history': global.language('price_history'),
    '/importproduct': global.language('import_product'),
    '/importproductfromfile': global.language('import_from_file'),
    '/importproductimage': global.language('import_product_image'),
    '/product_category_group_select_screen': global.language('select_category_group'),
    '/zone_group_select_screen': global.language('select_zone'),
    '/table_group_select_screen': global.language('select_table'),
    '/table_map_group_select_screen': global.language('table_map'),
    '/kitchen_group_select_screen': global.language('select_kitchen'),
    '/qrcodeorder_group_select_screen': global.language('create_qr_order'),
    '/transaction/stockbalance': global.language('stock_balance'),
    '/transaction/purchase': global.language('transaction_purchase'),
    '/transaction/purchasereturn': global.language('transaction_purchasereturn'),
    '/transaction/purchaseorder': global.language('transaction_purchase_order'),
    '/transaction/advancepayment': global.language('transaction_advance_payment'),
    '/transaction/advancepaymentrefund': global.language('transaction_advance_payment_refund'),
    '/transaction/deposit': global.language('transaction_deposit'),
    '/transaction/depositrefund': global.language('transaction_deposit_refund'),
    '/transaction/purchasepartial': global.language('transaction_purchase_partial'),
    '/transaction/accrualreceive': global.language('transaction_accrual_receive'),
    '/transaction/sale': global.language('transaction_sale'),
    '/transaction/salereturn': global.language('transaction_salereturn'),
    '/transaction/saleorder': global.language('transaction_sale_order'),
    '/transaction/paidadvance': global.language('transaction_paid_advance'),
    '/transaction/paidadvancerefund': global.language('transaction_paid_advance_refund'),
    '/transaction/receivedeposit': global.language('transaction_receive_deposit'),
    '/transaction/receivedepositrefund': global.language('transaction_receive_deposit_refund'),
    '/transaction/stocktransfer': global.language('transaction_stock_transfer'),
    '/transaction/stockreceiveproduct': global.language('transaction_stock_receive_product'),
    '/transaction/stockpickupproduct': global.language('transaction_stock_pick_up_product'),
    '/transaction/stockreturnproduct': global.language('transaction_stock_return_product'),
    '/transaction/adjust': global.language('transaction_adjust'),
    '/transaction/paid': global.language('transaction_paid'),
    '/transaction/pay': global.language('transaction_pay'),
    '/report/movement': global.language('report_movement'),
    '/report/pdfreport': global.language('report_pdf'),
    '/report/product_balance': global.language('report_stock_balance'),
    '/report/report_dedebi_sales': 'รายงานขาย DedeBi',
    '/report/report_dedebi_sales_daily': global.language('report_daily_sales'),
    '/report/report_stock_movement': global.language('report_stock_movement'),
    '/report/report_dedebi_payment_daily': global.language('report_daily_payment'),
    '/report/report_dedebi_sale_return': 'คืนสินค้า DedeBi',
    '/report/report_dedebi_stock_balance': 'คงเหลือ DedeBi',
    '/report/report_dedebi_purchase_partial': global.language('transaction_purchase_partial'),
    '/report/report_gross_profit_by_document': global.language('report_gross_profit_by_doc'),
    '/report/report_gross_profit_by_product': global.language('report_gross_profit_by_product'),
    '/report/report_vat_sale': global.language('report_vat_sale'),
    '/report/report_vat_buy': global.language('report_vat_buy'),
    '/report/stock_balance_item': global.language('report_stock_by_warehouse'),
    '/report/stock_balance_warehouse': global.language('report_warehouse_balance'),
    '/report/stock_balance_location': global.language('report_stock_by_location'),
    '/report/stock_movement_cost': global.language('report_stock_movement_cost'),
    '/report/audit_product': global.language('audit_product'),
    '/master_brand_screen': global.language('brand'),
    '/master_category_screen': global.language('product_category'),
    '/master_class_screen': global.language('class'),
    '/master_design_screen': global.language('design'),
    '/master_grade_screen': global.language('grade'),
    '/master_model_screen': global.language('model'),
    '/master_pattern_screen': global.language('pattern'),
    '/master_group_screen': global.language('group'),
    '/master_group_sub1_screen': global.language('sub_group_1'),
    '/master_group_sub2_screen': global.language('sub_group_2'),
    '/select_shop_screen': global.language('select_shop'),
    '/work_day_screen': global.language('work_day'),
    '/holiday_screen': global.language('holiday'),
    '/color_screen': global.language('color'),
    '/bussiness_type_screen': global.language('business_type'),
    '/book_bank_screen': global.language('book_bank'),
    '/sale_channel_screen': global.language('sale_channel'),
    '/transport_channel_screen': global.language('transport_channel'),
    '/product_warehouse_screen': global.language('warehouse'),
    '/product_location_screen': global.language('product_location'),
    '/order_type_screen': global.language('order_type'),
    '/product_type_screen': global.language('product_type'),
    '/add_product_to_branch_screen': global.language('add_product_to_branch'),
    '/add_product_to_department_screen': global.language('add_product_to_department'),
    '/add_product_to_kitchen_screen': global.language('add_product_to_kitchen'),
    '/gl/gl_process': global.language('gl_process'),
    '/check_daily/daily_info_screen': global.language('daily_info'),
    '/line_notify': global.language('line_notify_menu'),
    '/lineoa_config': global.language('line_oa_menu'),
    '/reminder': global.language('notification'),
    '/cashing_in_the_drawer': global.language('cash'),
    '/point_setting': global.language('point_setting'),
    '/coupon_setting': global.language('coupon'),
    // ระบบอนุมัติ
    '/purchase_type_screen': global.language('purchase_type'),
    '/po_approval_setting_screen': global.language('po_approval_setting'),
  };

  return routeTitles[routeName] ?? global.language('workspace');
}

/// ดึง Icon ตาม route name - ใช้ icon เดียวกับ menu_screen.dart
IconData getRouteIcon(Route<dynamic>? route) {
  if (route == null) return Icons.grid_view_rounded;

  final routeName = route.settings.name ?? '/';

  // Map route names เป็น Icons - ต้องตรงกับ menu_screen.dart
  final Map<String, IconData> routeIcons = {
    // === เมนูหลัก ===
    '/': Icons.grid_view_rounded,
    '/menu': Icons.grid_view_rounded,

    // === ตั้งค่าบริษัท (configMenuList) ===
    '/company': Icons.business, // บริษัท
    '/branch': Icons.account_tree_outlined, // สาขา
    '/department': Icons.corporate_fare, // แผนก
    '/business_type_screen': Icons.category, // ประเภทธุรกิจ
    '/work_day_screen': Icons.work_history, // วันทำงาน
    '/holiday_screen': Icons.event_available, // วันหยุด
    '/employee': Icons.badge, // พนักงาน (สำหรับ flavor อื่น)
    '/user': Icons.person_add, // ผู้ใช้งาน
    '/formdesign': Icons.manage_accounts, // ออกแบบฟอร์ม
    '/line_notify': Icons.notifications, // Line Notify (FontAwesome bell)
    '/lineoa_config': Icons.chat, // Line OA

    // === ลูกค้า/เจ้าหนี้ (customerMenuList) ===
    '/creditor': Icons.person_outline, // เจ้าหนี้
    '/creditorgroup': Icons.groups_outlined, // กลุ่มเจ้าหนี้
    '/debtor': Icons.person, // ลูกหนี้
    '/debtorgroup': Icons.groups, // กลุ่มลูกหนี้

    // === ตั้งค่าการขาย (newMenuList) ===
    '/color_screen': Icons.palette, // สี
    '/qrprovider': Icons.qr_code, // QR Provider
    '/bank': Icons.account_balance, // ธนาคาร
    '/book_bank_screen': Icons.account_balance_wallet, // สมุดบัญชี
    '/sale_channel_screen': Icons.shopping_basket, // ช่องทางการขาย
    '/transport_channel_screen': Icons.local_shipping, // ช่องทางขนส่ง
    '/possetting': Icons.point_of_sale, // ตั้งค่า POS
    '/posmedia': Icons.perm_media, // สื่อ POS
    '/point_setting': Icons.card_membership, // ตั้งค่าแต้ม
    '/coupon_setting': Icons.card_membership, // คูปอง
    '/docformat': Icons.description, // รูปแบบเอกสาร
    '/billdesign': Icons.receipt_long, // ออกแบบบิล

    // === ข้อมูลสินค้า (masterMenuList) ===
    '/product_barcode': Icons.qr_code, // บาร์โค้ด
    '/productunit': Icons.straighten, // หน่วยสินค้า
    '/product_barcode_shelf': Icons.print, // พิมพ์บาร์โค้ด
    '/product_category_group_select_screen': Icons.folder, // หมวดหมู่สินค้า
    '/productcategorylist': Icons.list, // รายการหมวดหมู่
    '/product_warehouse_screen': Icons.warehouse, // คลังสินค้า
    '/product_location_screen': Icons.location_on, // ตำแหน่งสินค้า
    '/order_type_screen': Icons.restaurant_menu, // ประเภทคำสั่ง
    '/product_type_screen': Icons.style, // ประเภทสินค้า
    '/product_dimension': Icons.architecture, // มิติสินค้า
    '/product_bom': Icons.ballot, // BOM
    '/promotion_screen': Icons.card_giftcard, // โปรโมชั่น
    '/price_history': Icons.money, // ประวัติราคา

    // === นำเข้าข้อมูลสินค้า (masterProductMenuList) ===
    '/product_category_list_screen': Icons.cloud_sync, // เพิ่มจาก Data Center
    '/add_product_to_branch_screen': Icons.add_business, // เพิ่มสินค้าสาขา
    '/add_product_to_department_screen': Icons.add_business, // เพิ่มสินค้าแผนก
    '/importproduct': Icons.upload_file, // นำเข้าสินค้า
    '/importproductimage': Icons.photo_library, // นำเข้ารูปสินค้า

    // === รายการซื้อ (transactionPurchaseMenuList) ===
    '/transaction/purchaseorder': Icons.shopping_cart_outlined, // ใบสั่งซื้อ
    '/transaction/advancepayment': Icons.payments, // จ่ายเงินล่วงหน้า
    '/transaction/advancepaymentrefund': Icons.attach_money, // รับคืนเงินล่วงหน้า
    '/transaction/deposit': Icons.attach_money, // จ่ายเงินมัดจำ
    '/transaction/depositrefund': Icons.attach_money, // รับคืนเงินมัดจำ
    '/transaction/purchase': Icons.inventory, // ซื้อสินค้า
    '/transaction/purchasereturn': Icons.keyboard_return, // คืนสินค้าซื้อ
    '/transaction/purchasepartial': Icons.add_shopping_cart, // รับสินค้าทยอย
    '/transaction/accrualreceive': Icons.add_shopping_cart, // ตั้งหนี้จากรับทยอย

    // === รายการขาย (transactionSaleMenuList) ===
    '/transaction/quotation': Icons.request_quote, // ใบเสนอราคา
    '/transaction/saleorder': Icons.shopping_cart, // ใบสั่งขาย
    '/transaction/paidadvance': Icons.attach_money, // รับเงินล่วงหน้า
    '/transaction/paidadvancerefund': Icons.attach_money, // คืนเงินล่วงหน้า
    '/transaction/receivedeposit': Icons.attach_money, // รับเงินมัดจำ
    '/transaction/receivedepositrefund': Icons.attach_money, // คืนเงินมัดจำ
    '/transaction/sale': Icons.point_of_sale, // ขายสินค้า
    '/transaction/salereturn': Icons.assignment_return, // คืนสินค้าขาย

    // === รายการสต๊อก (transactionStockMenuList) ===
    '/transaction/stocktransfer': Icons.compare_arrows, // โอนสินค้า
    '/transaction/stockreceiveproduct': Icons.inbox, // รับสินค้า
    '/transaction/stockpickupproduct': Icons.outbox, // เบิกสินค้า
    '/transaction/stockreturnproduct': Icons.assignment_return, // คืนสินค้า
    '/transaction/adjust': Icons.balance, // ปรับสต๊อก
    '/transaction/stockbalance': Icons.inventory, // ยอดคงเหลือ

    // === รายการชำระเงิน (transactionPaidMenuList) ===
    '/transaction/paid': Icons.payments, // รับชำระ
    '/transaction/pay': Icons.credit_card, // จ่ายชำระ
    '/slip_money_in': Icons.add_photo_alternate, // SLIP เงินเข้า
    '/slip_money_out': Icons.photo_library, // SLIP เงินออก

    // === รายงาน (reportMenuListOld) ===
    '/report/report_dedebi_sales': Icons.assessment, // รายงานขาย
    '/report/report_dedebi_sales_daily': Icons.assessment, // รายงานขายรายวัน
    '/report/report_stock_movement': Icons.assessment, // เคลื่อนไหวสินค้า
    '/report/report_dedebi_payment_daily': Icons.assessment, // รายงานรับจ่ายรายวัน
    '/report/report_dedebi_sale_return': Icons.assessment, // รายงานคืนสินค้า
    '/report/report_dedebi_stock_balance': Icons.assessment, // รายงานคงเหลือ
    '/report/report_dedebi_purchase_partial': Icons.assessment, // รายงานรับทยอย
    '/report/report_gross_profit_by_document': Icons.assessment, // กำไรขั้นต้นตามเอกสาร
    '/report/report_gross_profit_by_product': Icons.assessment, // กำไรขั้นต้นตามสินค้า
    '/report/report_vat_sale': Icons.assessment, // ภาษีขาย
    '/report/report_vat_buy': Icons.assessment, // ภาษีซื้อ

    // === รายงานสินค้า (reportMenuProductList) ===
    '/report/stock_balance_item': Icons.description, // คงเหลือตามสินค้า
    '/report/stock_balance_warehouse': Icons.summarize, // คงเหลือตามคลัง
    '/report/stock_balance_location': Icons.insert_chart, // คงเหลือตามที่เก็บ
    '/report/stock_movement_cost': Icons.insert_chart, // เคลื่อนไหวพร้อมต้นทุน

    // === รายงานการขาย (reportMenuSaleList) ===
    '/report/sales_report_by_document': Icons.bar_chart, // กำไรขั้นต้นตามเอกสาร

    // === ประมวลผล (reportMenuAuditList) ===
    '/rebuild_stock_screen': Icons.build, // ทดสอบประมวลผลใหม่
    '/audit_screen': Icons.inventory_2, // ตรวจสอบข้อมูล
    '/rebuild_product_balance_screen': Icons.calculate, // ประมวลผลยอดคงเหลือ

    // === ร้านอาหาร (restaurantMenuList) ===
    '/zone_group_select_screen': Icons.grid_on, // โซน
    '/table_group_select_screen': Icons.table_bar, // โต๊ะ
    '/table_map_group_select_screen': Icons.maps_home_work, // แผนผังโต๊ะ
    '/kitchen_group_select_screen': Icons.kitchen, // ครัว
    '/add_product_to_kitchen_screen': Icons.restaurant, // เพิ่มสินค้าครัว
    '/qrcodeorder_group_select_screen': Icons.qr_code_2, // QR สั่งอาหาร
    '/ordertemplatsetting': Icons.receipt, // เทมเพลตคำสั่งซื้อ
    '/ordersetting': Icons.settings_applications, // ตั้งค่าคำสั่งซื้อ

    // === บัญชี (glMenuList) ===
    '/check_daily/daily_info_screen': Icons.fact_check, // ตรวจสอบรายวัน
    '/cashing_in_the_drawer': Icons.fact_check, // รับ-ส่งเงิน POS

    // === ข้อมูลหลัก (masterDataMenuList) ===
    '/master_brand_screen': Icons.settings, // ยี่ห้อ
    '/master_category_screen': Icons.category, // หมวดหมู่
    '/master_class_screen': Icons.class_, // ชั้น
    '/master_design_screen': Icons.design_services, // ดีไซน์
    '/master_grade_screen': Icons.grade, // เกรด
    '/master_model_screen': Icons.model_training, // รุ่น
    '/master_pattern_screen': Icons.pattern, // ลาย
    '/master_group_screen': Icons.pattern, // กลุ่มหลัก
    '/master_group_sub1_screen': Icons.pattern, // กลุ่มย่อย 1
    '/master_group_sub2_screen': Icons.pattern, // กลุ่มย่อย 2

    // === Import/Export (importExportMenuList) ===
    '/importproductfromfile': Icons.upload_file, // นำเข้าข้อมูลสินค้า
    '/knowledge_base_screen': Icons.upload_file, // จัดการ Knowledge Base
    '/alert_agent_screen': Icons.notifications_active, // BC Alert Agent

    // === ระบบอนุมัติ (approvalMenuList) ===
    '/purchase_type_screen': Icons.category_outlined, // ประเภทการจัดซื้อ
    '/po_approval_setting_screen': Icons.fact_check, // อนุมัติใบสั่งซื้อ

    // === อื่นๆ ===
    '/cart-system': Icons.shopping_cart_checkout,
    '/select_shop_screen': Icons.store,
  };

  return routeIcons[routeName] ?? Icons.grid_view_rounded;
}

/// Model สำหรับ Tab
class MenuTab {
  final int id;
  String title; // เปลี่ยนจาก final เพื่อให้สามารถเปลี่ยนชื่อได้
  IconData icon; // icon ของ tab ที่เปลี่ยนได้ตาม route
  final bool isClosable;
  final Widget? customContent; // สำหรับ custom content เช่น CartSystemScreen

  MenuTab({
    required this.id,
    required this.title,
    this.icon = Icons.grid_view_rounded, // default icon
    required this.isClosable,
    this.customContent,
  });
}
