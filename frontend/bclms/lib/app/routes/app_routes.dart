import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:bclms/app/modules/login/login_page.dart';
import 'package:bclms/app/modules/shop_selection/shop_selection_page.dart';
import 'package:bclms/app/modules/branch_selection/branch_selection_page.dart';
import 'package:bclms/app/modules/home/tabbed_home_page.dart';
import 'package:bclms/app/modules/placeholder/placeholder_page.dart';
import 'package:bclms/app/modules/vehicle/vehicle_list_page.dart';
import 'package:bclms/app/modules/warehouse/warehouse_list_page.dart';

class AppRoutes {
  // === หน้าหลัก ===
  static const login = '/login';
  static const shopSelection = '/shop-selection';
  static const branchSelection = '/branch-selection';
  static const home = '/home';

  // === Tab 1: ปฏิบัติการ (Operations — workflow order) ===
  static const orderIntake = '/order-intake'; // รับออเดอร์ + Marketplace
  static const pickPack = '/pick-pack'; // Pick & Pack
  static const barcodeLabel = '/barcode-label'; // บาร์โค้ด/ฉลาก
  static const routePlan = '/route-plan'; // จัดเที่ยวรถ
  static const dispatch = '/dispatch'; // Dispatch
  static const tracking = '/tracking'; // ติดตาม GPS
  static const pod = '/pod'; // ยืนยันส่ง (POD)
  static const returns = '/returns'; // คืนสินค้า (RMA)
  static const codManagement = '/cod-management'; // จัดการ COD
  static const multimodal = '/multimodal'; // ขนส่งหลายรูปแบบ

  // === Tab 2: รายงาน (Reports & Analytics) ===
  static const dashboard = '/dashboard';
  static const fleetMap = '/fleet-map'; // แผนที่กองรถ
  static const transportReport = '/transport-report';
  static const codReport = '/cod-report'; // รายงาน COD
  static const pickpackReport = '/pickpack-report';
  static const inventoryReport = '/inventory-report'; // รายงานสต๊อก
  static const vehicleReport = '/vehicle-report';
  static const driverReport = '/driver-report';

  // === Tab 3: ข้อมูลหลัก (Master Data) ===
  static const vehicleList = '/vehicle-list';
  static const driverList = '/driver-list';
  static const routeMaster = '/route-master';
  static const deliveryZone = '/delivery-zone';
  static const warehouse = '/warehouse';
  static const storageLocation = '/storage-location'; // ที่เก็บสินค้า (Zone/Rack/Shelf/Bin)
  static const inventory = '/inventory'; // สินค้า/สต๊อก
  static const customerLocation = '/customer-location';
  static const carrier = '/carrier';
  static const rateTable = '/rate-table';
  static const marketplace = '/marketplace'; // Marketplace (Shopee/Lazada/TikTok)

  // === Tab 4: ตั้งค่า (Settings & System) ===
  static const generalSettings = '/general-settings';
  static const userRoles = '/user-roles';
  static const notificationSettings = '/notification-settings'; // LINE/SMS
  static const gpsIotSettings = '/gps-iot-settings';
  static const printTemplate = '/print-template';
  static const integrationSettings = '/integration-settings';
  static const fleetMaintenance = '/fleet-maintenance'; // บำรุงรักษา
  static const codSettings = '/cod-settings'; // ตั้งค่า COD

  /// Route generation สำหรับ nested Navigator ใน Tab
  /// route ที่พัฒนาแล้วจะไปหน้าจริง — ที่เหลือไปที่ PlaceholderPage
  static Route<dynamic>? onGenerateRoute(RouteSettings settings) {
    Widget page;
    switch (settings.name) {
      case vehicleList:
        page = const VehicleListPage();
        break;
      case warehouse:
      case storageLocation:
        page = const WarehouseListPage();
        break;
      default:
        page = const PlaceholderPage();
    }
    return MaterialPageRoute(
      builder: (_) => page,
      settings: settings,
    );
  }
}

/// GoRouter provider — root navigation (login flow → home)
final goRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: AppRoutes.login,
    routes: [
      GoRoute(
        path: AppRoutes.login,
        builder: (_, _) => const LoginPage(),
      ),
      GoRoute(
        path: AppRoutes.shopSelection,
        builder: (_, _) => const ShopSelectionPage(),
      ),
      GoRoute(
        path: AppRoutes.branchSelection,
        builder: (_, _) => const BranchSelectionPage(),
      ),
      GoRoute(
        path: AppRoutes.home,
        builder: (_, _) => const TabbedHomePage(),
      ),
    ],
  );
});
