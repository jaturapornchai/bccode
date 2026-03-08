import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/modules/home/models/menu_item_model.dart';
import 'package:bclms/app/routes/app_routes.dart';
import 'package:bclms/app/providers/service_providers.dart';

class HomeState {
  final int currentIndex;
  final String shopId;
  final String shopName;
  final String roleName;
  final String userName;
  final String userFullName;
  final String branchGuidFixed;
  final String apiUrl;
  final bool loggedOut;

  const HomeState({
    this.currentIndex = 0,
    this.shopId = '',
    this.shopName = '',
    this.roleName = '',
    this.userName = '',
    this.userFullName = '',
    this.branchGuidFixed = '',
    this.apiUrl = '',
    this.loggedOut = false,
  });

  HomeState copyWith({
    int? currentIndex,
    String? shopId,
    String? shopName,
    String? roleName,
    String? userName,
    String? userFullName,
    String? branchGuidFixed,
    String? apiUrl,
    bool? loggedOut,
  }) {
    return HomeState(
      currentIndex: currentIndex ?? this.currentIndex,
      shopId: shopId ?? this.shopId,
      shopName: shopName ?? this.shopName,
      roleName: roleName ?? this.roleName,
      userName: userName ?? this.userName,
      userFullName: userFullName ?? this.userFullName,
      branchGuidFixed: branchGuidFixed ?? this.branchGuidFixed,
      apiUrl: apiUrl ?? this.apiUrl,
      loggedOut: loggedOut ?? this.loggedOut,
    );
  }
}

class HomeNotifier extends Notifier<HomeState> {
  // === เมนูย่อยสำหรับแต่ละ Tab ===
  // อ้างอิงจาก: Flash Express, Kerry, J&T, Shippop, MOLOG WMS, Packhai,
  // ZORT, SKYFROG TMS, กฎหมาย COD ไทย (พ.ร.บ. คุ้มครองผู้บริโภค 2567)

  /// Tab 1: ปฏิบัติการ — 10 เมนู (เรียงตาม workflow ขนส่ง)
  /// รับออเดอร์ → Pick&Pack → พิมพ์ฉลาก → จัดเที่ยว → Dispatch → ติดตาม → POD → คืน → COD → ขนส่งหลายรูปแบบ
  static const operationMenus = <MenuItemModel>[
    MenuItemModel(
      id: 'order_intake',
      label: 'รับออเดอร์',
      icon: Icons.assignment_rounded,
      route: AppRoutes.orderIntake,
      color: AppTheme.terracotta,
    ),
    MenuItemModel(
      id: 'pick_pack',
      label: 'Pick & Pack',
      icon: Icons.inventory_2_rounded,
      route: AppRoutes.pickPack,
      color: Color(0xFF2E7D32),
    ),
    MenuItemModel(
      id: 'barcode_label',
      label: 'บาร์โค้ด/ฉลาก',
      icon: Icons.qr_code_scanner_rounded,
      route: AppRoutes.barcodeLabel,
      color: Color(0xFF4E342E),
    ),
    MenuItemModel(
      id: 'route_plan',
      label: 'จัดเที่ยวรถ',
      icon: Icons.route_rounded,
      route: AppRoutes.routePlan,
      color: Color(0xFF1565C0),
    ),
    MenuItemModel(
      id: 'dispatch',
      label: 'Dispatch',
      icon: Icons.local_shipping_rounded,
      route: AppRoutes.dispatch,
      color: Color(0xFFE65100),
    ),
    MenuItemModel(
      id: 'tracking',
      label: 'ติดตาม GPS',
      icon: Icons.gps_fixed_rounded,
      route: AppRoutes.tracking,
      color: Color(0xFF00838F),
    ),
    MenuItemModel(
      id: 'pod',
      label: 'ยืนยันส่ง (POD)',
      icon: Icons.verified_rounded,
      route: AppRoutes.pod,
      color: Color(0xFF388E3C),
    ),
    MenuItemModel(
      id: 'returns',
      label: 'คืนสินค้า (RMA)',
      icon: Icons.assignment_return_rounded,
      route: AppRoutes.returns,
      color: Color(0xFF6A1B9A),
    ),
    MenuItemModel(
      id: 'cod_management',
      label: 'จัดการ COD',
      icon: Icons.account_balance_wallet_rounded,
      route: AppRoutes.codManagement,
      color: Color(0xFFFF6F00),
    ),
    MenuItemModel(
      id: 'multimodal',
      label: 'ขนส่งหลายรูปแบบ',
      icon: Icons.hub_rounded,
      route: AppRoutes.multimodal,
      color: Color(0xFF37474F),
    ),
  ];

  /// Tab 2: รายงาน — 8 เมนู
  static const reportMenus = <MenuItemModel>[
    MenuItemModel(
      id: 'dashboard',
      label: 'Dashboard',
      icon: Icons.dashboard_rounded,
      route: AppRoutes.dashboard,
      color: AppTheme.terracotta,
    ),
    MenuItemModel(
      id: 'fleet_map',
      label: 'แผนที่กองรถ',
      icon: Icons.map_rounded,
      route: AppRoutes.fleetMap,
      color: Color(0xFF1565C0),
    ),
    MenuItemModel(
      id: 'transport_report',
      label: 'รายงานขนส่ง',
      icon: Icons.analytics_rounded,
      route: AppRoutes.transportReport,
      color: Color(0xFF00838F),
    ),
    MenuItemModel(
      id: 'cod_report',
      label: 'รายงาน COD',
      icon: Icons.account_balance_rounded,
      route: AppRoutes.codReport,
      color: Color(0xFFFF6F00),
    ),
    MenuItemModel(
      id: 'pickpack_report',
      label: 'รายงาน Pick&Pack',
      icon: Icons.fact_check_rounded,
      route: AppRoutes.pickpackReport,
      color: Color(0xFF2E7D32),
    ),
    MenuItemModel(
      id: 'inventory_report',
      label: 'รายงานสต๊อก',
      icon: Icons.inventory_rounded,
      route: AppRoutes.inventoryReport,
      color: Color(0xFF795548),
    ),
    MenuItemModel(
      id: 'vehicle_report',
      label: 'รายงานยานพาหนะ',
      icon: Icons.directions_car_rounded,
      route: AppRoutes.vehicleReport,
      color: Color(0xFFE65100),
    ),
    MenuItemModel(
      id: 'driver_report',
      label: 'รายงานคนขับ',
      icon: Icons.person_pin_rounded,
      route: AppRoutes.driverReport,
      color: Color(0xFF6A1B9A),
    ),
  ];

  /// Tab 3: ข้อมูลหลัก — 10 เมนู
  static const masterMenus = <MenuItemModel>[
    MenuItemModel(
      id: 'vehicle',
      label: 'ยานพาหนะ',
      icon: Icons.local_shipping_rounded,
      route: AppRoutes.vehicleList,
      color: AppTheme.terracotta,
    ),
    MenuItemModel(
      id: 'driver',
      label: 'คนขับ',
      icon: Icons.badge_rounded,
      route: AppRoutes.driverList,
      color: Color(0xFF1565C0),
    ),
    MenuItemModel(
      id: 'route_master',
      label: 'เส้นทาง',
      icon: Icons.alt_route_rounded,
      route: AppRoutes.routeMaster,
      color: Color(0xFF2E7D32),
    ),
    MenuItemModel(
      id: 'delivery_zone',
      label: 'โซนจัดส่ง',
      icon: Icons.location_on_rounded,
      route: AppRoutes.deliveryZone,
      color: Color(0xFFE65100),
    ),
    MenuItemModel(
      id: 'warehouse',
      label: 'คลังสินค้า',
      icon: Icons.warehouse_rounded,
      route: AppRoutes.warehouse,
      color: Color(0xFF00838F),
    ),
    MenuItemModel(
      id: 'storage_location',
      label: 'ที่เก็บสินค้า',
      icon: Icons.shelves,
      route: AppRoutes.storageLocation,
      color: Color(0xFF00695C),
    ),
    MenuItemModel(
      id: 'inventory',
      label: 'สินค้า/สต๊อก',
      icon: Icons.category_rounded,
      route: AppRoutes.inventory,
      color: Color(0xFF795548),
    ),
    MenuItemModel(
      id: 'customer_location',
      label: 'ลูกค้า/จุดส่ง',
      icon: Icons.people_alt_rounded,
      route: AppRoutes.customerLocation,
      color: Color(0xFF6A1B9A),
    ),
    MenuItemModel(
      id: 'carrier',
      label: 'ผู้ให้บริการขนส่ง',
      icon: Icons.business_rounded,
      route: AppRoutes.carrier,
      color: Color(0xFF37474F),
    ),
    MenuItemModel(
      id: 'rate_table',
      label: 'ค่าขนส่ง',
      icon: Icons.payments_rounded,
      route: AppRoutes.rateTable,
      color: Color(0xFF4E342E),
    ),
    MenuItemModel(
      id: 'marketplace',
      label: 'Marketplace',
      icon: Icons.storefront_rounded,
      route: AppRoutes.marketplace,
      color: Color(0xFFAD1457),
    ),
  ];

  /// Tab 4: ตั้งค่า — 8 เมนู
  static const settingMenus = <MenuItemModel>[
    MenuItemModel(
      id: 'general_settings',
      label: 'ตั้งค่าทั่วไป',
      icon: Icons.tune_rounded,
      route: AppRoutes.generalSettings,
      color: AppTheme.terracotta,
    ),
    MenuItemModel(
      id: 'user_roles',
      label: 'ผู้ใช้งาน & สิทธิ์',
      icon: Icons.admin_panel_settings_rounded,
      route: AppRoutes.userRoles,
      color: Color(0xFF1565C0),
    ),
    MenuItemModel(
      id: 'notification_settings',
      label: 'แจ้งเตือนลูกค้า',
      icon: Icons.notifications_active_rounded,
      route: AppRoutes.notificationSettings,
      color: Color(0xFFE65100),
    ),
    MenuItemModel(
      id: 'gps_iot',
      label: 'GPS & IoT',
      icon: Icons.satellite_alt_rounded,
      route: AppRoutes.gpsIotSettings,
      color: Color(0xFF00838F),
    ),
    MenuItemModel(
      id: 'print_template',
      label: 'พิมพ์เอกสาร',
      icon: Icons.print_rounded,
      route: AppRoutes.printTemplate,
      color: Color(0xFF6A1B9A),
    ),
    MenuItemModel(
      id: 'integration',
      label: 'เชื่อมต่อภายนอก',
      icon: Icons.extension_rounded,
      route: AppRoutes.integrationSettings,
      color: Color(0xFF37474F),
    ),
    MenuItemModel(
      id: 'fleet_maintenance',
      label: 'บำรุงรักษา',
      icon: Icons.build_rounded,
      route: AppRoutes.fleetMaintenance,
      color: Color(0xFF4E342E),
    ),
    MenuItemModel(
      id: 'cod_settings',
      label: 'ตั้งค่า COD',
      icon: Icons.money_rounded,
      route: AppRoutes.codSettings,
      color: Color(0xFFFF6F00),
    ),
  ];

  @override
  HomeState build() {
    Future.microtask(() => _loadUserData());
    final api = ref.read(apiServiceProvider);
    return HomeState(apiUrl: api.baseUrl);
  }

  Future<void> _loadUserData() async {
    final storage = ref.read(storageServiceProvider);
    final shopId = await storage.getShopId() ?? '';
    final shopName = await storage.getShopName() ?? '';
    final userName = await storage.getUsername() ?? '';
    final userFullName = await storage.getName() ?? '';
    final branchGuidFixed = await storage.getBranchGuidFixed() ?? '';

    final role = await storage.getRole() ?? 0;
    final roleName = switch (role) {
      2 => 'Super Admin',
      1 => 'Admin',
      _ => 'User',
    };

    state = state.copyWith(
      shopId: shopId,
      shopName: shopName,
      userName: userName,
      userFullName: userFullName,
      branchGuidFixed: branchGuidFixed,
      roleName: roleName,
    );

    AppLogger.info('[Home] โหลดข้อมูลร้าน: $shopName ($shopId), user: $userName, role: $roleName');
  }

  void changeTab(int index) {
    state = state.copyWith(currentIndex: index);
  }

  Future<void> logout() async {
    final authService = ref.read(authServiceProvider);
    await authService.logout();
    state = state.copyWith(loggedOut: true);
  }
}

final homeProvider = NotifierProvider<HomeNotifier, HomeState>(
  HomeNotifier.new,
);
