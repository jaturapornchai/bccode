import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_widget_from_html/flutter_widget_from_html.dart';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/components/create_sub_shop_dialog.dart';
import 'package:smlaicloud/model/shop_model.dart';
import 'package:flutter/material.dart';
import 'package:auto_size_text/auto_size_text.dart';
import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';
import 'package:smlaicloud/bloc/profile/profile_bloc.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_rounded_date_picker/flutter_rounded_date_picker.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'flavors.dart';
import 'package:smlaicloud/services/permission_service.dart';
import 'package:smlaicloud/services/branch_config_watcher_service.dart';
import 'package:smlaicloud/repositories/profile_repository.dart';
import 'package:smlaicloud/widgets/line_login_qr_dialog.dart';
import 'package:smlaicloud/widgets/manual_button.dart';

class MenuScreen extends StatefulWidget {
  const MenuScreen({super.key});

  @override
  State<MenuScreen> createState() => MenuScreenState();
}

class MenuScreenState extends State<MenuScreen> with TickerProviderStateMixin {
  late TabController mainTabController = TabController(length: 4, vsync: this, initialIndex: global.activeIndexMenu);
  List<Widget> masterMenuList = [];
  List<Widget> masterProductMenuList = [];
  List<Widget> transactionPurchaseMenuList = [];
  List<Widget> transactionSaleMenuList = [];
  List<Widget> transactionStockMenuList = [];
  List<Widget> transactionPaidMenuList = [];
  List<Widget> reportMenuListOld = [];
  List<Widget> reportMenuProductList = [];
  List<Widget> reportMenuSaleList = []; // เพิ่มรายงานการขาย
  // reportMenuList ถูกเปลี่ยนเป็น getter ด้านล่าง เพื่อให้ rebuild ทุกครั้ง
  List<Widget> customerMenuList = [];
  List<Widget> newMenuList = [];
  List<Widget> configMenuList = [];
  List<Widget> restaurantMenuList = [];
  List<Widget> glMenuList = [];
  List<Widget> configCompanyMenuList = [];
  List<Widget> masterDataMenuList = [];
  List<Widget> transactionAdvanceAndDeposit = [];
  List<Widget> approvalMenuList = []; // เมนูระบบอนุมัติ
  List<Widget> productSettingMenuList = []; // ตั้งค่าสินค้า
  List<Widget> importExportMenuList = [];

  // HTML content from web
  String htmlContent = '';
  int htmlContentVersion = 0; // ใช้เป็น key เพื่อบังคับ rebuild

  // Build info variables
  String buildVersion = '';
  String buildTime = '';
  String buildGitHash = '';
  String buildGitBranch = '';

  // สีจากตารางสี - Modern Theme (ใช้สีจาก Flavor)
  Color get primaryColor => F.primaryGradientColor;
  Color get primaryDarkColor => F.secondaryGradientColor;
  final Color primaryLightColor = const Color(0xFFa5b4fc);
  final Color secondaryColor = const Color(0xFF06b6d4);
  final Color secondaryDarkColor = const Color(0xFF0891b2);
  final Color secondaryLightColor = const Color(0xFF67e8f9);
  final Color backgroundColor = const Color(0xFFF8FAFC);
  final Color surfaceColor = const Color(0xFFFFFFFF);
  final Color errorColor = const Color(0xFFef4444);
  final Color successColor = const Color(0xFF22c55e);
  final Color warningColor = const Color(0xFFf59e0b);
  final Color infoColor = const Color(0xFF3b82f6);

  // สีสำหรับปุ่มและการแสดงข้อมูลทางการเงิน
  final Color positiveColor = const Color(0xFF0B8043);
  final Color negativeColor = const Color(0xFFD50000);
  final Color pendingColor = const Color(0xFFF9A825);
  final Color verifiedColor = const Color(0xFF039BE5);
  final Color taxColor = const Color(0xFF7B1FA2);
  final Color assetColor = const Color(0xFF00897B);
  final Color liabilityColor = const Color(0xFFC2185B);

  // สีสำหรับชาร์ตและการวิเคราะห์
  final Map<String, Color> chartColors = {
    'income': const Color(0xFF2E7D32),
    'expense': const Color(0xFFC62828),
    'assets': const Color(0xFF1565C0),
    'liabilities': const Color(0xFFD81B60),
    'equity': const Color(0xFF6A1B9A),
    'series1': const Color(0xFF5899DA),
    'series2': const Color(0xFFE8743B),
    'series3': const Color(0xFF19A979),
    'series4': const Color(0xFFED4A7B),
    'series5': const Color(0xFF945ECF),
    'series6': const Color(0xFF13A4B4),
    'series7': const Color(0xFF525DF4),
    'series8': const Color(0xFFBF399E),
  };

  // ปรับปรุงสีใหม่สำหรับปุ่มเมนู (Modern Gradient Theme - ใช้สีจาก Flavor)
  Map<String, List<Color>> get menuButtonColors => {
    'config': [F.primaryGradientColor, F.secondaryGradientColor],
    'customer': [const Color(0xFF06b6d4), const Color(0xFF0891b2)],
    'master': [const Color(0xFF10b981), const Color(0xFF059669)],
    'transaction': [const Color(0xFF22c55e), const Color(0xFF16a34a)],
    'report': [const Color(0xFFf59e0b), const Color(0xFFd97706)],
    'restaurant': [const Color(0xFF8b5cf6), const Color(0xFF7c3aed)],
    'gl': [const Color(0xFF06b6d4), const Color(0xFF0e7490)],
    'finance': [const Color(0xFFec4899), const Color(0xFFdb2777)],
    'system': [F.primaryGradientColor, F.secondaryGradientColor],
    'advance_payment': [const Color(0xFF14b8a6), const Color(0xFF0d9488)],
  };

  List<Widget> get reportMenuAuditList => [
    menuWidget(label: global.language("test_reprocess"), category: 'report', icon: Icons.build, routeName: '/rebuild_stock_screen'),
    menuWidget(label: global.language("audit_data"), category: 'report', icon: Icons.inventory_2, routeName: '/audit_screen'),
    menuWidget(label: global.language("rebuild_products"), category: 'report', icon: Icons.sync, routeName: '/rebuild_products_screen'),
    menuWidget(label: global.language("rebuild_product_balance"), category: 'report', icon: Icons.calculate, routeName: '/rebuild_product_balance_screen'),
  ];

  // ✅ เพิ่ม getter สำหรับ reportMenuList เพื่อให้ rebuild ทุกครั้ง
  Widget get reportMenuList {
    return SingleChildScrollView(
      child: Padding(
        padding: EdgeInsets.all(10),
        child: Column(
          children: [
            buildMenuCategoryContainer(title: global.language("inventory_report"), menuItems: reportMenuProductList),
            SizedBox(height: 15),
            buildMenuCategoryContainer(title: global.language("sale_report"), menuItems: reportMenuSaleList),
            SizedBox(height: 15),
            buildMenuCategoryContainer(title: global.language("other_reports"), menuItems: reportMenuListOld),
            const SizedBox(height: 15),
            htmlView(),
          ],
        ),
      ),
    );
  }

  Widget menuWidget({
    required String label,
    required String category,
    IconData? icon,
    Function? callback,
    String? routeName, // ✅ เพิ่ม routeName สำหรับ auto-navigation พร้อม title
    bool isAllowed = true, // ถ้า false จะแสดงเมนูแต่ block การเข้าถึง
  }) {
    // ถ้าไม่มีสิทธิ์ → แสดง dialog แจ้งเตือนแทน navigate
    if (!isAllowed) {
      return GestureDetector(
        onTap: () {
          AppLogger.warning('[Menu] ไม่มีสิทธิ์เข้าถึงเมนู "$label"');
          showDialog(
            context: context,
            builder: (ctx) => AlertDialog(
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
              title: Row(
                children: [
                  Icon(Icons.lock_outline, color: Colors.red[400], size: 24),
                  SizedBox(width: 8),
                  Text(global.language('no_permission_title'), style: TextStyle(fontSize: 18)),
                ],
              ),
              content: Text(global.language('no_permission_message')),
              actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: Text(global.language('ok')))],
            ),
          );
        },
        child: Opacity(opacity: 0.5, child: _buildMenuContainer(label, category, icon)),
      );
    }

    // ถ้ามี routeName ให้ใช้ auto callback ที่ส่ง label เป็น title
    final effectiveCallback = routeName != null
        ? () => navigateInTab(routeName, title: label)
        : (callback ??
              () {
                // Log สำหรับ callback ที่ไม่มี routeName
                AppLogger.info('🔄 MENU CLICK: กดเมนู "$label" (category: $category)');
              });

    return _buildMenuContainer(label, category, icon, onTap: () => effectiveCallback());
  }

  /// สร้าง UI container ของปุ่มเมนู (ใช้ร่วมกันทั้ง isAllowed=true และ false)
  Widget _buildMenuContainer(String label, String category, IconData? icon, {VoidCallback? onTap}) {
    List<Color> gradientColors = menuButtonColors[category] ?? [primaryDarkColor, primaryColor];

    Widget textWidget = Center(
      child: AutoSizeText(
        label,
        maxLines: 3,
        textAlign: TextAlign.center,
        style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Colors.white, letterSpacing: 0.3),
      ),
    );

    return Container(
      decoration: BoxDecoration(
        boxShadow: [BoxShadow(color: gradientColors[0].withOpacity(0.35), spreadRadius: 0, blurRadius: 12, offset: const Offset(0, 6))],
        borderRadius: BorderRadius.circular(20),
        gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: gradientColors),
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(20),
          child: Container(
            padding: const EdgeInsets.all(12),
            child: Stack(
              children: [
                Positioned(
                  top: -15,
                  right: -15,
                  child: Container(
                    width: 50,
                    height: 50,
                    decoration: BoxDecoration(shape: BoxShape.circle, color: Colors.white.withOpacity(0.1)),
                  ),
                ),
                Center(
                  child: FittedBox(
                    fit: BoxFit.scaleDown,
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        if (icon != null) ...[
                          Container(
                            padding: const EdgeInsets.all(10),
                            decoration: BoxDecoration(color: Colors.white.withOpacity(0.2), borderRadius: BorderRadius.circular(12)),
                            child: Icon(icon, size: 26, color: Colors.white),
                          ),
                          const SizedBox(height: 10),
                        ],
                        textWidget,
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  /// Helper function สำหรับ navigate ภายใน tab
  /// ใช้แทน Navigator.pushReplacementNamed เพื่อให้แต่ละ tab แยกกัน
  /// [title] - ชื่อที่จะแสดงบน tab (ถ้าไม่ระบุจะใช้ route mapping)
  void navigateInTab(String routeName, {String? title}) {
    // 📝 Log เพื่อติดตามว่าผู้ใช้เข้าหน้าไหนบ้าง
    AppLogger.info('🔄 NAVIGATION: กำลังเปิดหน้า "$title" (route: $routeName)');

    // ใช้ Navigator.of(context, rootNavigator: false) เพื่อให้ใช้ Navigator ใกล้ที่สุด (ใน tab)
    Navigator.of(context, rootNavigator: false).pushNamed(routeName, arguments: title != null ? {'title': title} : null);
  }

  void buildMenu() {
    // เช็คค่า ismainshop จาก shop_info
    bool showCreateSubShop = false;
    String? shopInfoJson = global.appConfig.getString("shop_info");

    if (shopInfoJson != null && shopInfoJson.isNotEmpty) {
      try {
        ShopModel shopModel = ShopModel.fromJson(jsonDecode(shopInfoJson));
        // String mainShopId = shopModel.mainshopid ?? ''; // Not used
        bool ismainShopid = shopModel.ismainshop ?? false;
        if (ismainShopid) {
          showCreateSubShop = true;
        }
      } catch (e) {
        // ถ้า error ในการ parse ให้ไม่แสดงเมนู
        showCreateSubShop = false;
      }
    }

    // เช็ค role จาก storage
    String userRole = global.appConfig.getString("role") ?? "0";
    bool isUser = userRole == "0";
    bool isAdmin = userRole == "1";
    bool isSuperAdmin = userRole == "2";

    configMenuList = [];

    // แสดงทุกเมนู — ถ้าไม่มีสิทธิ์จะ block ตอนคลิก (isAllowed: false)
    configMenuList.add(menuWidget(label: global.language("currency"), category: 'config', icon: Icons.currency_exchange, routeName: '/currency', isAllowed: isSuperAdmin));
    configMenuList.add(menuWidget(label: global.language("company_type"), category: 'config', icon: Icons.category, routeName: '/business_type_screen', isAllowed: isSuperAdmin));
    configMenuList.add(menuWidget(label: global.language("company"), category: 'config', icon: Icons.business, routeName: '/company', isAllowed: isSuperAdmin));
    configMenuList.add(
      menuWidget(
        label: global.language("company_branch"),
        category: 'config',
        icon: Icons.account_tree_outlined,
        isAllowed: isSuperAdmin,
        callback: () async {
          final label = global.language("company_branch");
          AppLogger.info('🔄 NAVIGATION: กำลังเปิดหน้า "$label" (route: /branch)');
          await Navigator.of(context, rootNavigator: false).pushNamed('/branch', arguments: {'title': label});
          await _refreshBranchSettings();
        },
      ),
    );
    configMenuList.add(menuWidget(label: global.language("company_department"), category: 'config', icon: Icons.corporate_fare, routeName: '/department', isAllowed: isSuperAdmin));
    if (global.posVersion == global.PosVersionEnum.restaurant) {
      configMenuList.add(menuWidget(label: global.language("workday"), category: 'config', icon: Icons.work_history, routeName: '/work_day_screen', isAllowed: isSuperAdmin));
      configMenuList.add(menuWidget(label: global.language("holidays"), category: 'config', icon: Icons.event_available, routeName: '/holiday_screen', isAllowed: isSuperAdmin));
    }

    // เมนู employee และ user
    configMenuList.add(menuWidget(label: global.language("employee"), category: 'system', icon: Icons.badge, routeName: '/employee'));
    configMenuList.add(menuWidget(label: global.language("user"), category: 'system', icon: Icons.manage_accounts, routeName: '/user', isAllowed: isAdmin || isSuperAdmin));
    configMenuList.add(menuWidget(label: global.language("design_form"), category: 'system', icon: Icons.manage_accounts, routeName: '/formdesign'));
    if (showCreateSubShop) {
      configMenuList.add(
        menuWidget(
          label: global.language("create_sub_shop"),
          category: 'system',
          icon: Icons.store,
          callback: () {
            CreateSubShopDialog.show(context);
          },
        ),
      );
    }
    if (global.posVersion == global.PosVersionEnum.restaurant) {
      configMenuList.add(menuWidget(label: global.language("line_notify"), category: 'system', icon: FontAwesomeIcons.bell, routeName: '/line_notify', isAllowed: isSuperAdmin));
    }

    configMenuList.add(menuWidget(label: global.language("permission_definition"), category: 'system', icon: Icons.security, routeName: '/permission_definition', isAllowed: isSuperAdmin));
    configMenuList.add(menuWidget(label: global.language("permission_link"), category: 'system', icon: Icons.link, routeName: '/permission_link', isAllowed: isSuperAdmin));
    configMenuList.add(menuWidget(label: global.language('mcp_token'), category: 'system', icon: Icons.vpn_key, routeName: '/mcp_apikey', isAllowed: isSuperAdmin));
    configMenuList.add(menuWidget(label: global.language("transfer_data"), category: 'system', icon: Icons.cloud_download, routeName: '/copy_uat_to_dev', isAllowed: isSuperAdmin));

    customerMenuList = [];
    // สำหรับ target flavor ให้เช็ค role, สำหรับ flavor อื่น ๆ ให้แสดงปกติ
    if (isUser || isAdmin || isSuperAdmin) {
      customerMenuList.add(menuWidget(label: global.language("creditor"), category: 'customer', icon: Icons.person_outline, routeName: '/creditor'));
      customerMenuList.add(menuWidget(label: global.language("creditor_group"), category: 'customer', icon: Icons.groups_outlined, routeName: '/creditorgroup'));
      customerMenuList.add(menuWidget(label: global.language("debtor"), category: 'customer', icon: Icons.person, routeName: '/debtor'));
      customerMenuList.add(menuWidget(label: global.language("debtor_group"), category: 'customer', icon: Icons.groups, routeName: '/debtorgroup'));
    }

    newMenuList = [];
    if (global.posVersion == global.PosVersionEnum.restaurant && isSuperAdmin) {
      newMenuList.add(menuWidget(label: global.language("color"), category: 'system', icon: Icons.palette, routeName: '/color_screen'));
    }

    // สำหรับ target flavor เฉพาะ admin และ superadmin เท่านั้น
    if (isAdmin || isSuperAdmin) {
      newMenuList.add(menuWidget(label: global.language("qr_provider"), category: 'system', icon: Icons.qr_code, routeName: '/qrprovider'));
      newMenuList.add(menuWidget(label: global.language("bank"), category: 'system', icon: Icons.account_balance, routeName: '/bank'));
      newMenuList.add(menuWidget(label: global.language("book_bank"), category: 'system', icon: Icons.account_balance_wallet, routeName: '/book_bank_screen'));
    }

    if (global.posVersion == global.PosVersionEnum.restaurant && isSuperAdmin) {
      newMenuList.add(menuWidget(label: global.language("sale_channel"), category: 'master', icon: Icons.shopping_basket, routeName: '/sale_channel_screen'));
      newMenuList.add(menuWidget(label: global.language("transport_channel"), category: 'master', icon: Icons.local_shipping, routeName: '/transport_channel_screen'));
    }

    // สำหรับ target flavor เฉพาะ admin และ superadmin เท่านั้น
    if (isAdmin || isSuperAdmin) {
      newMenuList.add(menuWidget(label: global.language("pos_setting"), category: 'system', icon: Icons.point_of_sale, routeName: '/possetting'));
    }

    // เมนูกำหนดอัตราแลกเปลี่ยน - แสดงสำหรับ admin และ superadmin
    if (isAdmin || isSuperAdmin) {
      newMenuList.add(menuWidget(label: global.language("set_exchange_rate"), category: 'finance', icon: Icons.currency_exchange, routeName: '/exchange_rate'));
    }

    if (isSuperAdmin) {
      newMenuList.add(menuWidget(label: global.language("pos_media"), category: 'system', icon: Icons.perm_media, routeName: '/posmedia'));
      newMenuList.add(menuWidget(label: global.language("set_loyalty_points"), category: 'system', icon: Icons.card_membership, routeName: '/point_setting'));
      newMenuList.add(menuWidget(label: global.language("set_coupons"), category: 'system', icon: Icons.card_membership, routeName: '/coupon_setting'));
      if (global.posVersion == global.PosVersionEnum.restaurant) {
        newMenuList.add(menuWidget(label: global.language("doc_format"), category: 'system', icon: Icons.description, routeName: '/docformat'));
        newMenuList.add(menuWidget(label: global.language("bill_design"), category: 'system', icon: Icons.receipt_long, routeName: '/billdesign'));
      }
    }
    masterMenuList = [];

    // เมนูที่ user, admin และ superadmin เห็นได้ทั้งหมด
    if (isUser || isAdmin || isSuperAdmin) {
      masterMenuList.add(menuWidget(label: global.language("barcode"), category: 'master', icon: Icons.qr_code, routeName: '/product_barcode'));
      masterMenuList.add(menuWidget(label: global.language("product_unit"), category: 'master', icon: Icons.straighten, routeName: '/productunit'));
      masterMenuList.add(menuWidget(label: global.language("print_product_label"), category: 'master', icon: Icons.print, routeName: '/product_barcode_shelf'));
    }

    // เมนูที่เฉพาะ admin และ superadmin เท่านั้น สำหรับ target flavor
    if (isAdmin || isSuperAdmin) {
      masterMenuList.add(menuWidget(label: global.language("promotion"), category: 'master', icon: Icons.card_giftcard, routeName: '/promotion_screen'));
      masterMenuList.add(menuWidget(label: global.language("product_price_edit_history"), category: 'master', icon: Icons.money, routeName: '/price_history'));
    }

    /// ตั้งค่าสินค้า (แยกจากจัดการสินค้า — ตั้งค่าครั้งเดียว ไม่ค่อยแก้)
    productSettingMenuList = [];
    if (isSuperAdmin) {
      productSettingMenuList.add(menuWidget(label: global.language("product_category"), category: 'master', icon: Icons.folder, routeName: '/product_category_group_select_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("product_category_list"), category: 'master', icon: Icons.list, routeName: '/productcategorylist'));
      productSettingMenuList.add(menuWidget(label: global.language("product_group"), category: 'master', icon: Icons.workspaces, routeName: '/productgroup'));
      productSettingMenuList.add(menuWidget(label: global.language("product_warehouse"), category: 'master', icon: Icons.warehouse, routeName: '/product_warehouse_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("product_location"), category: 'master', icon: Icons.location_on, routeName: '/product_location_screen'));
      if (global.posVersion == global.PosVersionEnum.restaurant) {
        productSettingMenuList.add(menuWidget(label: global.language("order_type"), category: 'master', icon: Icons.restaurant_menu, routeName: '/order_type_screen'));
      }
      productSettingMenuList.add(menuWidget(label: global.language("product_type"), category: 'master', icon: Icons.style, routeName: '/product_type_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("product_dimension"), category: 'master', icon: Icons.architecture, routeName: '/product_dimension'));
      if (global.posVersion == global.PosVersionEnum.restaurant) {
        productSettingMenuList.add(menuWidget(label: global.language("product_bom"), category: 'master', icon: Icons.ballot, routeName: '/product_bom'));
      }
      // ข้อมูลหลักสินค้า (ยี่ห้อ, ระดับ, รูปทรง, เกรด, รุ่น, รูปแบบ, กลุ่มย่อย)
      productSettingMenuList.add(menuWidget(label: global.language("brand"), category: 'master', icon: Icons.branding_watermark, routeName: '/master_brand_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("category"), category: 'master', icon: Icons.category, routeName: '/master_category_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("class"), category: 'master', icon: Icons.class_, routeName: '/master_class_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("design"), category: 'master', icon: Icons.design_services, routeName: '/master_design_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("grade"), category: 'master', icon: Icons.grade, routeName: '/master_grade_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("model"), category: 'master', icon: Icons.model_training, routeName: '/master_model_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("pattern"), category: 'master', icon: Icons.pattern, routeName: '/master_pattern_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("group_main"), category: 'master', icon: Icons.account_tree, routeName: '/master_group_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("group_sub1"), category: 'master', icon: Icons.subdirectory_arrow_right, routeName: '/master_group_sub1_screen'));
      productSettingMenuList.add(menuWidget(label: global.language("group_sub2"), category: 'master', icon: Icons.subdirectory_arrow_right, routeName: '/master_group_sub2_screen'));
    }
    masterProductMenuList = [];
    // สำหรับ target flavor เฉพาะ admin และ superadmin เท่านั้น
    if (isAdmin || isSuperAdmin) {
      if (global.posVersion == global.PosVersionEnum.restaurant) {
        masterProductMenuList.add(menuWidget(label: global.language("add_product_to_branch"), category: 'master', icon: Icons.add_business, routeName: '/add_product_to_branch_screen'));
        masterProductMenuList.add(menuWidget(label: global.language("add_product_to_department"), category: 'master', icon: Icons.add_business, routeName: '/add_product_to_department_screen'));
      }

    }
    transactionPurchaseMenuList = [];
    transactionPurchaseMenuList.add(
      menuWidget(label: global.language("transaction_purchase_order"), category: 'transaction', icon: Icons.shopping_cart_outlined, routeName: '/transaction/purchaseorder'),
    );
    // จ่ายเงินล่วงหน้า
    transactionPurchaseMenuList.add(menuWidget(label: global.language("transaction_advance_payment"), category: 'advance_payment', icon: Icons.payments, routeName: '/transaction/advancepayment'));
    // รับคืนเงินล่วงหน้า
    transactionPurchaseMenuList.add(
      menuWidget(label: global.language("receive_back_advance_payment"), category: 'advance_payment', icon: Icons.attach_money, routeName: '/transaction/advancepaymentrefund'),
    );
    // จ่ายเงินมัดจำ
    transactionPurchaseMenuList.add(menuWidget(label: global.language("pay_deposit"), category: 'advance_payment', icon: Icons.attach_money, routeName: '/transaction/deposit'));
    // รับคืนเงินมัดจำ
    transactionPurchaseMenuList.add(menuWidget(label: global.language("receive_back_deposit"), category: 'advance_payment', icon: Icons.attach_money, routeName: '/transaction/depositrefund'));
    transactionPurchaseMenuList.add(menuWidget(label: global.language("transaction_purchase"), category: 'transaction', icon: Icons.inventory, routeName: '/transaction/purchase'));
    transactionPurchaseMenuList.add(menuWidget(label: global.language("transaction_purchase_return"), category: 'transaction', icon: Icons.keyboard_return, routeName: '/transaction/purchasereturn'));
    transactionPurchaseMenuList.add(menuWidget(label: global.language("gradual_product_receipt"), category: 'transaction', icon: Icons.add_shopping_cart, routeName: '/transaction/purchasepartial'));
    transactionPurchaseMenuList.add(
      menuWidget(label: global.language("set_debt_from_gradual_receipt"), category: 'transaction', icon: Icons.add_shopping_cart, routeName: '/transaction/accrualreceive'),
    );

    transactionSaleMenuList = [];
    // ใบเสนอราคา
    transactionSaleMenuList.add(menuWidget(label: global.language("transaction_quotation"), category: 'transaction', icon: Icons.request_quote, routeName: '/transaction/quotation'));
    // ใบสั่งขาย
    transactionSaleMenuList.add(menuWidget(label: global.language("transaction_sale_order"), category: 'transaction', icon: Icons.shopping_cart, routeName: '/transaction/saleorder'));

    // รับเงินล่วงหน้า
    transactionSaleMenuList.add(menuWidget(label: global.language("receive_advance_payment"), category: 'advance_payment', icon: Icons.attach_money, routeName: '/transaction/paidadvance'));
    // คืนเงินล่วงหน้า
    transactionSaleMenuList.add(menuWidget(label: global.language("return_advance_payment"), category: 'advance_payment', icon: Icons.attach_money, routeName: '/transaction/paidadvancerefund'));
    // รับเงินมัดจำ
    transactionSaleMenuList.add(menuWidget(label: global.language("receive_deposit"), category: 'advance_payment', icon: Icons.attach_money, routeName: '/transaction/receivedeposit'));
    // คืนเงินมัดจำ
    transactionSaleMenuList.add(menuWidget(label: global.language("return_deposit"), category: 'advance_payment', icon: Icons.attach_money, routeName: '/transaction/receivedepositrefund'));

    transactionSaleMenuList.add(menuWidget(label: global.language("transaction_sale"), category: 'transaction', icon: Icons.point_of_sale, routeName: '/transaction/sale'));
    transactionSaleMenuList.add(menuWidget(label: global.language("transaction_sale_return"), category: 'transaction', icon: Icons.assignment_return, routeName: '/transaction/salereturn'));

    transactionStockMenuList = [];
    transactionStockMenuList.add(menuWidget(label: global.language("transaction_stock_transfer"), category: 'transaction', icon: Icons.compare_arrows, routeName: '/transaction/stocktransfer'));
    transactionStockMenuList.add(menuWidget(label: global.language("transaction_stock_receive_product"), category: 'transaction', icon: Icons.inbox, routeName: '/transaction/stockreceiveproduct'));

    transactionStockMenuList.add(menuWidget(label: global.language("transaction_stock_pick_up_product"), category: 'transaction', icon: Icons.outbox, routeName: '/transaction/stockpickupproduct'));
    transactionStockMenuList.add(
      menuWidget(label: global.language("transaction_stock_return_product"), category: 'transaction', icon: Icons.assignment_return, routeName: '/transaction/stockreturnproduct'),
    );
    transactionStockMenuList.add(menuWidget(label: global.language("transaction_adjust"), category: 'transaction', icon: Icons.balance, routeName: '/transaction/adjust'));
    transactionStockMenuList.add(menuWidget(label: global.language("transaction_stock_balance"), category: 'transaction', icon: Icons.inventory, routeName: '/transaction/stockbalance'));
    transactionPaidMenuList = [];
    transactionPaidMenuList.addAll([
      menuWidget(label: global.language("transaction_paid"), category: 'finance', icon: Icons.payments, routeName: '/transaction/paid'),
      menuWidget(label: global.language("transaction_pay"), category: 'finance', icon: Icons.credit_card, routeName: '/transaction/pay'),
      // SLIP เงินเข้า/ออก
      menuWidget(label: global.language("slip_money_in"), category: 'finance', icon: Icons.add_photo_alternate, routeName: '/slip_money_in'),
      menuWidget(label: global.language("slip_money_out"), category: 'finance', icon: Icons.photo_library, routeName: '/slip_money_out'),
    ]);
    reportMenuListOld = [];
    reportMenuListOld.addAll([
      // menuWidget(
      //   label: global.language("report"),
      //   category: 'report',
      //   icon: Icons.assessment,
      //   callback: () {
      //     Navigator.pushNamed(context, '/report/pdfreport');
      //   },
      // ),
      menuWidget(label: global.language("sales_report"), category: 'report', icon: Icons.assessment, routeName: '/report/report_dedebi_sales'),
      menuWidget(label: global.language("sales_report_by_date"), category: 'report', icon: Icons.assessment, routeName: '/report/report_dedebi_sales_daily'),
      menuWidget(label: global.language("product_movement"), category: 'report', icon: Icons.assessment, routeName: '/report/report_stock_movement'),
      menuWidget(label: global.language("payment_received_report_by_date"), category: 'report', icon: Icons.assessment, routeName: '/report/report_dedebi_payment_daily'),
      menuWidget(label: global.language("debt_reduction_return_report"), category: 'report', icon: Icons.assessment, routeName: '/report/report_dedebi_sale_return'),
      menuWidget(label: global.language("inventory_report"), category: 'report', icon: Icons.assessment, routeName: '/report/report_dedebi_stock_balance'),
      menuWidget(label: global.language("gradual_receipt_report"), category: 'report', icon: Icons.assessment, routeName: '/report/report_dedebi_purchase_partial'),

      /// รายงาน กำไรขั้นต้นตามเอกสาร
      menuWidget(label: global.language("gross_profit_by_document_report"), category: 'report', icon: Icons.assessment, routeName: '/report/report_gross_profit_by_document'),

      /// รายงาน กำไรขั้นต้นตามสินค้า
      menuWidget(label: global.language("gross_profit_by_product_report"), category: 'report', icon: Icons.assessment, routeName: '/report/report_gross_profit_by_product'),

      /// รายงาน ภาษีขาย report_vat_sale
      menuWidget(label: global.language("report_vat_sale"), category: 'report', icon: Icons.assessment, routeName: '/report/report_vat_sale'),

      /// รายงาน ภาษีซื้อ report_vat_buy
      menuWidget(label: global.language("report_vat_buy"), category: 'report', icon: Icons.assessment, routeName: '/report/report_vat_buy'),
    ]);

    reportMenuProductList = [
      menuWidget(label: global.language("stock_balance_by_item"), category: 'report', icon: Icons.description, routeName: '/report/stock_balance_item'),
      menuWidget(label: global.language("stock_balance_by_warehouse"), category: 'report', icon: Icons.summarize, routeName: '/report/stock_balance_warehouse'),
      menuWidget(label: global.language("stock_balance_by_location"), category: 'report', icon: Icons.insert_chart, routeName: '/report/stock_balance_location'),
      menuWidget(label: global.language("stock_movement_with_cost"), category: 'report', icon: Icons.insert_chart, routeName: '/report/stock_movement_cost'),
    ];
    reportMenuSaleList = [];
    reportMenuSaleList.addAll([menuWidget(label: global.language("gross_profit_report_by_document"), category: 'report', icon: Icons.bar_chart, routeName: '/report/sales_report_by_document')]);
    // reportMenuList ถูกย้ายไปเป็น getter function แล้ว
    restaurantMenuList = [];
    if (global.posVersion == global.PosVersionEnum.restaurant && isSuperAdmin) {
      restaurantMenuList.add(menuWidget(label: global.language("zone"), category: 'restaurant', icon: Icons.grid_on, routeName: '/zone_group_select_screen'));
      restaurantMenuList.add(menuWidget(label: global.language("table"), category: 'restaurant', icon: Icons.table_bar, routeName: '/table_group_select_screen'));
      restaurantMenuList.add(menuWidget(label: global.language("table_map"), category: 'restaurant', icon: Icons.maps_home_work, routeName: '/table_map_group_select_screen'));
      restaurantMenuList.add(menuWidget(label: global.language("kitchen"), category: 'restaurant', icon: Icons.kitchen, routeName: '/kitchen_group_select_screen'));
      restaurantMenuList.add(menuWidget(label: global.language("add_product_to_kitchen"), category: 'restaurant', icon: Icons.restaurant, routeName: '/add_product_to_kitchen_screen'));
      restaurantMenuList.add(menuWidget(label: global.language("qr_code_order"), category: 'restaurant', icon: Icons.qr_code_2, routeName: '/qrcodeorder_group_select_screen'));
      restaurantMenuList.add(menuWidget(label: global.language("order_template_setting"), category: 'restaurant', icon: Icons.receipt, routeName: '/ordertemplatsetting'));
      restaurantMenuList.add(menuWidget(label: global.language("order_setting"), category: 'restaurant', icon: Icons.settings_applications, routeName: '/ordersetting'));
    }
    glMenuList = [];
    if (isSuperAdmin || isAdmin || isUser) {
      glMenuList.addAll([
        // menuWidget(
        //     label: global.language("gl_process"),
        //     category: 'gl',
        //     icon: Icons.account_balance,
        //     callback: () {
        //       navigateInTab( '/gl/gl_process');
        //     }),
        menuWidget(label: global.language("check_daily"), category: 'gl', icon: Icons.fact_check, routeName: '/check_daily/daily_info_screen'),
        menuWidget(label: global.language("receive_send_money_pos"), category: 'gl', icon: Icons.fact_check, routeName: '/cashing_in_the_drawer'),
      ]);
    }

    /// ข้อมูลหลัก (ย้าย master data ไปรวมกับ ตั้งค่าสินค้า แล้ว — เหลือแค่เมนูตรวจสอบ)
    masterDataMenuList = [];
    if (isSuperAdmin) {
      // เมนูตรวจสอบข้อมูล — ย้ายมาจากแท็บรายงาน
      masterDataMenuList.addAll(reportMenuAuditList);
    }

    /// ระบบอนุมัติ (Approval System)
    approvalMenuList = [];
    if (isSuperAdmin || isAdmin) {
      approvalMenuList.add(menuWidget(label: global.language("purchase_type"), category: 'approval', icon: Icons.category_outlined, routeName: '/purchase_type_screen'));
      approvalMenuList.add(menuWidget(label: global.language("approve_purchase_order"), category: 'approval', icon: Icons.fact_check, routeName: '/po_approval_setting_screen'));
      approvalMenuList.add(menuWidget(label: global.language("quotation_type"), category: 'approval', icon: Icons.request_quote_outlined, routeName: '/quotation_type_screen'));
      approvalMenuList.add(menuWidget(label: global.language("approve_quotation"), category: 'approval', icon: Icons.verified_outlined, routeName: '/qt_approval_setting_screen'));
      approvalMenuList.add(menuWidget(label: global.language("sale_order_type"), category: 'approval', icon: Icons.shopping_cart_outlined, routeName: '/sale_order_type_screen'));
      approvalMenuList.add(menuWidget(label: global.language("approve_sale_order"), category: 'approval', icon: Icons.shopping_cart_checkout, routeName: '/so_approval_setting_screen'));
    }

    /// Import/Export ข้อมูล
    importExportMenuList = [];
    // สำหรับ target flavor เฉพาะ admin และ superadmin เท่านั้น
    if (isAdmin || isSuperAdmin) {
      importExportMenuList.add(menuWidget(label: global.language("import_product"), category: 'master', icon: Icons.upload_file, routeName: '/importproduct'));
      importExportMenuList.add(menuWidget(label: global.language("import_product_data"), category: 'master', icon: Icons.upload_file, routeName: '/importproductfromfile'));
      importExportMenuList.add(menuWidget(label: global.language("manage_knowledge_base"), category: 'master', icon: Icons.upload_file, routeName: '/knowledge_base_screen'));
      // BC Alert Agent - จัดการการแจ้งเตือนอัตโนมัติ
      importExportMenuList.add(menuWidget(label: global.language('bc_alert_agent'), category: 'master', icon: Icons.notifications_active, routeName: '/alert_agent_screen'));
    }
    // เฉพาะ superadmin เท่านั้น
    if (isSuperAdmin) {
      importExportMenuList.add(menuWidget(label: global.language("import_product_image"), category: 'master', icon: Icons.photo_library, routeName: '/importproductimage'));
    }
  }

  // วิดเจ็ตสำหรับแสดงส่วนหัวของกลุ่มเมนู
  Widget buildMenuCategoryContainer({required String title, required List<Widget> menuItems}) {
    // ถ้าไม่มีเมนูให้ return SizedBox.shrink() เพื่อซ่อน
    if (menuItems.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: primaryColor.withOpacity(0.3), width: 1),
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.05), spreadRadius: 1, blurRadius: 3, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.only(left: 5, bottom: 8),
            child: Text(
              title,
              style: TextStyle(color: primaryDarkColor, fontWeight: FontWeight.bold, fontSize: 16),
            ),
          ),
          GridView.builder(
            padding: const EdgeInsets.all(0),
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(maxCrossAxisExtent: 150, crossAxisSpacing: 10, mainAxisSpacing: 10),
            itemCount: menuItems.length,
            itemBuilder: (BuildContext ctx, index) {
              return menuItems[index];
            },
          ),
        ],
      ),
    );
  }

  Future<void> loadBuildInfo() async {
    if (kIsWeb) {
      try {
        final response = await http.get(Uri.parse('/build-info.json?t=${DateTime.now().millisecondsSinceEpoch}'));

        if (response.statusCode == 200) {
          final data = jsonDecode(response.body);
          if (mounted) {
            setState(() {
              buildVersion = '${data['version']}+${data['buildNumber']}';
              buildTime = data['buildTime'] ?? '';
              buildGitHash = data['gitHash'] ?? '';
              buildGitBranch = data['gitBranch'] ?? '';
            });
          }
          if (kDebugMode) {
            AppLogger.info('[MenuScreen] Build info loaded: $buildVersion');
          }
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('[MenuScreen] Error loading build info: $e');
        }
      }
    }
  }

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);
    mainTabController.addListener(() {
      global.activeIndexMenu = mainTabController.index;
    });
    buildMenu();
    loadBuildInfo(); // Load build info
    if (mounted) {
      setState(() {});
    }
  }

  Future<void> loadHtmlContent(int mode) async {
    final response = await global.getApiDataInfo("main_menu", mode);
    if (mounted) {
      setState(() {
        htmlContent = response.toString();
        htmlContentVersion++;
      });
    }
  }

  @override
  void initState() {
    super.initState();
    // Log เมื่อเข้าหน้าเมนู
    AppLogger.info('🏠 MENU SCREEN: เข้าสู่หน้าเมนูหลัก (initial tab index: ${global.activeIndexMenu})');

    mainTabController = TabController(length: 4, vsync: this, initialIndex: global.activeIndexMenu);
    setSystemLanguageList();
    global.getApiServiceVersion().then((value) {
      global.goApiVersion = value.toString();
      AppLogger.info('🌐 API VERSION: $value');
      if (mounted) {
        setState(() {});
      }
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // โหลด HTML ทุกครั้งที่เข้าหน้าจอนี้
    loadHtmlContent(1);
    // โหลดสิทธิ์ทุกครั้งที่กลับมาหน้าเมนู
    _loadUserPermissions();
    // Refresh branch settings จาก MongoDB ทุกครั้งที่เข้าเมนู
    // เพื่อให้แน่ใจว่าได้การตั้งค่าล่าสุด (ภาษา, timezone, ประเภทปี)
    _refreshBranchSettings();
  }

  /// Refresh การตั้งค่าสาขาจาก MongoDB
  /// เรียกทุกครั้งที่เข้าเมนูหลัก เพื่อให้แน่ใจว่าได้การตั้งค่าล่าสุด
  /// (ภาษา, timezone, ประเภทปี) กรณีมีคนอื่นแก้ไข
  Future<void> _refreshBranchSettings() async {
    try {
      AppLogger.info('[MenuScreen] Refreshing branch settings จาก MongoDB');
      await branchConfigWatcher.forceReload();
      if (mounted) setState(() {});
    } catch (e) {
      AppLogger.error('[MenuScreen] Error refreshing branch settings: $e');
    }
  }

  /// ชื่อที่แสดงใน menu bar — แสดง "รหัส ชื่อ" (เช่น "00001 สมชาย")
  String _getDisplayName() {
    if (global.loginName.isNotEmpty) return global.loginName;
    final username = global.profileData.username ?? global.userLoginData.name;
    final profileName = global.profileData.name ?? '';
    if (profileName.isNotEmpty && profileName != username) {
      return '$username $profileName';
    }
    return username;
  }

  /// Dialog แก้ไขชื่อ (เฉพาะ user/password login)
  void _showEditNameDialog() {
    final profileName = global.profileData.name ?? '';
    final username = global.profileData.username ?? global.userLoginData.name;
    // pre-fill เฉพาะชื่อ (ไม่รวมรหัส)
    final controller = TextEditingController(text: (profileName.isNotEmpty && profileName != username) ? profileName : '');
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(global.language("edit_name")),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // แสดงรหัสพนักงาน (แก้ไม่ได้)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              decoration: BoxDecoration(
                color: Colors.grey[100],
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.grey[300]!),
              ),
              child: Row(
                children: [
                  Icon(Icons.badge_outlined, color: Colors.grey[600], size: 20),
                  SizedBox(width: 8),
                  Text(
                    '${global.language("employee_code")} : $username',
                    style: TextStyle(fontSize: 14, color: Colors.grey[700]),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            // ช่องแก้ไขชื่อ
            TextField(
              controller: controller,
              autofocus: true,
              decoration: InputDecoration(
                labelText: global.language("name"),
                border: const OutlineInputBorder(),
                prefixIcon: const Icon(Icons.person_outline),
              ),
              onSubmitted: (_) => _saveNameFromDialog(ctx, controller),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text(global.language("cancel"))),
          ElevatedButton(onPressed: () => _saveNameFromDialog(ctx, controller), child: Text(global.language("save"))),
        ],
      ),
    );
  }

  /// บันทึกชื่อใหม่ลง MongoDB ผ่าน PUT /profile
  Future<void> _saveNameFromDialog(BuildContext ctx, TextEditingController controller) async {
    final newName = controller.text.trim();
    if (newName.isEmpty) return;

    Navigator.pop(ctx);
    try {
      final profileRepo = ProfileRepository();
      final profile = global.profileData;
      profile.name = newName;
      final result = await profileRepo.updateProfile(profile);
      if (result.success) {
        setState(() {
          global.profileData.name = newName;
        });
        if (mounted) {
          showSuccessSnackBar(context, global.language("success_save"));
        }
      } else {
        if (mounted) {
          showErrorSnackBar(context, global.language("not_success_save"));
        }
      }
    } catch (e) {
      AppLogger.error('[MenuScreen] Error updating name: $e');
      if (mounted) {
        showErrorSnackBar(context, '${global.language("not_success_save")}: $e');
      }
    }
  }

  /// โหลดสิทธิ์ของผู้ใช้ปัจจุบัน
  Future<void> _loadUserPermissions() async {
    try {
      final username = global.profileData.username;
      if (username != null && username.isNotEmpty) {
        await permissionService.loadUserPermissions(username);
        // เริ่ม auto refresh ถ้ายังไม่ได้เริ่ม
        permissionService.startAutoRefresh();
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error loading user permissions: $e');
      }
    }
  }

  @override
  void dispose() {
    mainTabController.dispose();
    super.dispose();
  }

  Widget htmlView() {
    return Container(
      key: ValueKey('html_$htmlContentVersion'), // ✅ เพิ่ม key เพื่อบังคับ rebuild
      width: double.infinity,
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: primaryColor.withOpacity(0.3), width: 1),
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.05), spreadRadius: 1, blurRadius: 3, offset: const Offset(0, 2))],
      ),
      child: SelectionArea(
        // ✅ เพิ่ม SelectionArea เพื่อให้เลือกข้อความได้
        child: HtmlWidget(
          htmlContent,
          key: ValueKey('html_widget_$htmlContentVersion'), // ✅ เพิ่ม key ให้ HtmlWidget ด้วย
          textStyle: const TextStyle(fontSize: 14, color: Colors.black87),
          // เพิ่ม renderMode เพื่อให้ support selection ได้ดีขึ้น
          renderMode: RenderMode.column,
        ),
      ),
    );
  }

  String getHeader() {
    String headTitle = "";
    global.activeLangName(global.companyBranchSelectData.names);
    if (mounted) {
      setState(() {
        switch (global.activeIndexMenu) {
          case 0:
            headTitle = global.language("menu_transaction");
            break;
          case 1:
            headTitle = global.language("menu_report");
            break;
          case 2:
            headTitle = global.language("menu_master");
            break;
          case 3:
            headTitle = global.language("menu_setup");
            break;
        }
      });
    }
    return headTitle;
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: [
        BlocListener<LoginBloc, LoginState>(
          listener: (context, state) {
            if (state is LogoutSuccess) {
              Navigator.pushNamedAndRemoveUntil(context, '/login_screen', (route) => false);
            }
          },
        ),
        BlocListener<ProfileBloc, ProfileState>(
          listener: (context, state) {
            if (state is UpdateProfileSuccess) {
              setState(() {
                global.showSnackBar(context, Icon(Icons.edit, color: Colors.white), global.language("edit_success"), Colors.blue);
              });
              context.read<ProfileBloc>().add(const GetProfile());
            }
            if (state is GetProfileSuccess) {
              setState(() {
                global.profileData = state.profile;
                if (global.profileData.yeartype == "buddhist") {
                  global.local = const Locale('th', 'TH');
                  global.eraMode = EraMode.BUDDHIST_YEAR;
                }
              });
            }
          },
        ),
      ],
      child: Scaffold(
        resizeToAvoidBottomInset: true,
        appBar: AppBar(
          automaticallyImplyLeading: false,
          title: Row(
            children: [
              Text(
                getHeader(),
                style: TextStyle(
                  fontSize: 18,
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  letterSpacing: 0.5,
                  shadows: [Shadow(offset: const Offset(1.5, 1.5), blurRadius: 3.0, color: Colors.black.withOpacity(0.4))],
                ),
                overflow: TextOverflow.ellipsis,
              ),
              const Text(' : ', style: TextStyle(fontSize: 14, color: Colors.white54)),
              Expanded(
                child: Text(
                  global.goApiUrlPath('').replaceAll(RegExp(r'/+$'), ''),
                  style: const TextStyle(fontSize: 12, color: Colors.white70),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          backgroundColor: const Color(0xFF0A3880),
          actions: const [
            ManualButton(),
            SizedBox(width: 8),
          ],
        ),
        body: Container(
          width: double.infinity,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topRight,
              end: Alignment.bottomLeft,
              colors: [Colors.white, primaryLightColor.withOpacity(0.3), Colors.white.withOpacity(0.9)],
              stops: const [0.0, 0.5, 1.0],
            ),
          ),
          child: Column(
            children: [
              Container(
                width: double.infinity,
                padding: const EdgeInsets.only(left: 10, right: 10, top: 4, bottom: 4),
                color: primaryDarkColor.withOpacity(0.95),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Wrap(
                      crossAxisAlignment: WrapCrossAlignment.center,
                      children: [
                        Text(
                          (appConfig.getString("name") != "") ? (appConfig.getString("name") ?? "") : global.activeLangName(global.shopSelectData.names!),
                          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
                        ),
                        SizedBox(width: 10),
                        Text(" (${global.activeLangName(global.companyBranchSelectData.names)})", style: const TextStyle(fontSize: 12, color: Colors.white)),
                        /*if (appConfig.getInt("branch_total") != 1)
                      ElevatedButton(
                        onPressed: (appConfig.getInt("branch_total") != 1)
                            ? () {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                      builder: (context) =>
                                          const SelectBranchScreen()),
                                );
                              }
                            : null,
                        child: null,
                      ),*/
                        SizedBox(width: 10),
                        Text(
                          "SHOP ID : ${global.getShopId()} : Api (${global.goApiVersion})",
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(color: Colors.white, fontSize: 10),
                        ),
                        SizedBox(width: 8),
                        Text(
                          global.goApiUrlPath('').replaceAll(RegExp(r'/+$'), ''),
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(color: Colors.white70, fontSize: 10),
                        ),
                      ],
                    ),
                    if (buildVersion.isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.only(top: 3),
                        child: Wrap(
                          spacing: 8,
                          children: [
                            Text("Version: $buildVersion", style: const TextStyle(color: Colors.white70, fontSize: 9)),
                            if (buildGitBranch.isNotEmpty) Text("Branch: $buildGitBranch", style: const TextStyle(color: Colors.white70, fontSize: 9)),
                            if (buildGitHash.isNotEmpty) Text("Hash: $buildGitHash", style: const TextStyle(color: Colors.white70, fontSize: 9)),
                            if (buildTime.isNotEmpty) Text("Build: $buildTime", style: const TextStyle(color: Colors.white70, fontSize: 9)),
                          ],
                        ),
                      ),
                    // แสดง username + ข้อมูลสาขา
                    Padding(
                      padding: const EdgeInsets.only(top: 5, bottom: 5),
                      child: Wrap(
                        crossAxisAlignment: WrapCrossAlignment.center,
                        runSpacing: 4,
                        children: [
                          // Username + ชื่อ (กดเปลี่ยนชื่อได้ เฉพาะ user/password login)
                          GestureDetector(
                            onTap: global.loginName.isEmpty ? () => _showEditNameDialog() : null,
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Text(
                                  _getDisplayName(),
                                  style: const TextStyle(
                                    fontSize: 14,
                                    color: Colors.white,
                                    fontWeight: FontWeight.bold,
                                    shadows: [Shadow(offset: Offset(1.0, 1.0), blurRadius: 3.0, color: Color.fromARGB(128, 0, 0, 0))],
                                  ),
                                ),
                                const SizedBox(width: 5),
                                Icon(
                                  color: Colors.white,
                                  global.loginName.isEmpty ? Icons.edit : Icons.person,
                                  size: 16,
                                  shadows: const [Shadow(offset: Offset(1.0, 1.0), blurRadius: 3.0, color: Colors.black54)],
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(width: 10),
                          // Divider
                          Container(width: 1, height: 16, color: Colors.white54),
                          const SizedBox(width: 10),
                          // Branch Code & Name
                          const Icon(Icons.store, color: Colors.white70, size: 14),
                          const SizedBox(width: 4),
                          Text(
                            "${global.companyBranchSelectData.code} - ${global.activeLangName(global.companyBranchSelectData.names)}",
                            style: const TextStyle(
                              fontSize: 12,
                              color: Colors.white,
                              shadows: [Shadow(offset: Offset(1.0, 1.0), blurRadius: 3.0, color: Color.fromARGB(128, 0, 0, 0))],
                            ),
                          ),
                          const SizedBox(width: 10),
                          // Currency
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(color: Colors.amber.shade700, borderRadius: BorderRadius.circular(4)),
                            child: Text(
                              global.companyBranchSelectData.baseCurrency ?? 'THB',
                              style: const TextStyle(fontSize: 10, color: Colors.white, fontWeight: FontWeight.bold),
                            ),
                          ),
                          const SizedBox(width: 6),
                          // Year Type (พ.ศ. / ค.ศ.)
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(
                              color: global.companyBranchSelectData.yeartype == 'buddhist' ? Colors.purple.shade600 : Colors.blue.shade600,
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Text(
                              global.companyBranchSelectData.yeartype == 'buddhist' ? global.language("buddhist_era") : global.language("christian_era"),
                              style: const TextStyle(fontSize: 10, color: Colors.white, fontWeight: FontWeight.bold),
                            ),
                          ),
                          const SizedBox(width: 6),
                          // Language
                          Builder(
                            builder: (context) {
                              final languageKey = global.language("languages");
                              final rawLanguage = global.companyBranchSelectData.language;

                              if (rawLanguage == null || rawLanguage.isEmpty) {
                                return Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                  decoration: BoxDecoration(color: Colors.red.shade600, borderRadius: BorderRadius.circular(4)),
                                  child: Text(
                                    '$languageKey: ${global.language("not_configured")}',
                                    style: const TextStyle(fontSize: 11, color: Colors.white, fontWeight: FontWeight.bold),
                                  ),
                                );
                              }

                              return Container(
                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                decoration: BoxDecoration(color: Colors.green.shade600, borderRadius: BorderRadius.circular(4)),
                                child: Text(
                                  '$languageKey: ${rawLanguage.toUpperCase()}',
                                  style: const TextStyle(fontSize: 11, color: Colors.white, fontWeight: FontWeight.bold),
                                ),
                              );
                            },
                          ),
                          const SizedBox(width: 6),
                          // Timezone
                          Builder(
                            builder: (context) {
                              final timezoneKey = global.language("timezone");

                              String? timezoneValue;
                              if (global.companyBranchSelectData.timezonelabel != null && global.companyBranchSelectData.timezonelabel!.isNotEmpty) {
                                timezoneValue = global.companyBranchSelectData.timezonelabel!;
                              } else if (global.companyBranchSelectData.timezone != null && global.companyBranchSelectData.timezone!.isNotEmpty) {
                                timezoneValue = global.companyBranchSelectData.timezone!;
                              }

                              if (timezoneValue == null) {
                                return Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                  decoration: BoxDecoration(color: Colors.red.shade600, borderRadius: BorderRadius.circular(4)),
                                  child: Text(
                                    '$timezoneKey: ไม่ได้ตั้งค่า',
                                    style: const TextStyle(fontSize: 11, color: Colors.white, fontWeight: FontWeight.bold),
                                  ),
                                );
                              }

                              return Container(
                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                decoration: BoxDecoration(color: Colors.teal.shade600, borderRadius: BorderRadius.circular(4)),
                                child: Text(
                                  '$timezoneKey: $timezoneValue',
                                  style: const TextStyle(fontSize: 11, color: Colors.white, fontWeight: FontWeight.bold),
                                ),
                              );
                            },
                          ),
                          // LINE Profile Badge
                          Builder(
                            builder: (context) {
                              final lineUserId = global.profileData.lineUserId ?? '';
                              final lineDisplayName = global.profileData.lineDisplayName ?? '';
                              final isLinked = lineUserId.isNotEmpty;

                              return ElevatedButton.icon(
                                onPressed: () async {
                                  if (isLinked) {
                                    // แสดง dialog ยืนยันยกเลิกการเชื่อมต่อ
                                    final confirm = await showDialog<bool>(
                                      context: context,
                                      builder: (ctx) => AlertDialog(
                                        title: Text(global.language("unlink_line")),
                                        content: Text('${global.language("confirm_unlink_line")} ($lineDisplayName)'),
                                        actions: [
                                          TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(global.language("cancel"))),
                                          TextButton(
                                            onPressed: () => Navigator.pop(ctx, true),
                                            child: Text(global.language("confirm"), style: TextStyle(color: Colors.red)),
                                          ),
                                        ],
                                      ),
                                    );
                                    if (confirm == true && context.mounted) {
                                      try {
                                        await ProfileRepository().unlinkLine();
                                        if (context.mounted) {
                                          context.read<ProfileBloc>().add( GetProfile());
                                          global.showSuccessSnackBar(context, global.language("unlink_line_success"));
                                        }
                                      } catch (e) {
                                        if (context.mounted) {
                                          global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
                                        }
                                      }
                                    }
                                  } else {
                                    // เปิด LINE QR dialog เพื่อเชื่อมต่อ
                                    final result = await showLineLoginQRDialog(context: context, liffId: '2008792333-tziDpgnJ');
                                    if (result != null && context.mounted) {
                                      try {
                                        await ProfileRepository().linkLine(lineUserId: result.userId, lineDisplayName: result.displayName, linePictureUrl: result.pictureUrl ?? '');
                                        if (context.mounted) {
                                          context.read<ProfileBloc>().add( GetProfile());
                                          global.showSuccessSnackBar(context, global.language("link_line_success"));
                                        }
                                      } catch (e) {
                                        if (context.mounted) {
                                          global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
                                        }
                                      }
                                    }
                                  }
                                },
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: isLinked ? const Color(0xFF06C755) : Colors.grey.shade600,
                                  foregroundColor: Colors.white,
                                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 0),
                                  minimumSize: const Size(0, 28),
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                                  elevation: 2,
                                ),
                                icon: Icon(Icons.chat_bubble, size: 13),
                                label: Text(isLinked ? 'LINE: $lineDisplayName' : global.language("connect_line"), style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold)),
                              );
                            },
                          ),
                          const SizedBox(width: 6),
                          // Google Profile
                          Builder(
                            builder: (context) {
                              // ใช้ข้อมูลจาก userLoginData (persist ใน storage) หรือ global vars
                              final photoUrl = global.loginPhotoUrl.isNotEmpty ? global.loginPhotoUrl : global.userLoginData.photourl;
                              final email = global.loginEmail.isNotEmpty ? global.loginEmail : global.userLoginData.email;
                              final displayName = global.loginName.isNotEmpty ? global.loginName : global.userLoginData.name;

                              if (photoUrl.isEmpty && email.isEmpty) {
                                return const SizedBox.shrink();
                              }

                              return Container(
                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                decoration: BoxDecoration(color: Colors.white.withOpacity(0.15), borderRadius: BorderRadius.circular(20)),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    // Google profile info
                                    Column(
                                      crossAxisAlignment: CrossAxisAlignment.end,
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        if (displayName.isNotEmpty)
                                          Text(
                                            displayName,
                                            style: const TextStyle(fontSize: 11, color: Colors.white, fontWeight: FontWeight.bold),
                                          ),
                                        if (email.isNotEmpty) Text(email, style: TextStyle(fontSize: 10, color: Colors.white.withOpacity(0.8))),
                                      ],
                                    ),
                                    const SizedBox(width: 8),
                                    // Google profile avatar
                                    CircleAvatar(
                                      radius: 14,
                                      backgroundColor: Colors.white24,
                                      backgroundImage: photoUrl.isNotEmpty ? NetworkImage(photoUrl) : null,
                                      child: photoUrl.isEmpty ? const Icon(Icons.account_circle, size: 24, color: Colors.white70) : null,
                                    ),
                                  ],
                                ),
                              );
                            },
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: TabBarView(
                  physics: const NeverScrollableScrollPhysics(),
                  controller: mainTabController,
                  children: [
                    SingleChildScrollView(
                      child: Padding(
                        padding: EdgeInsets.all(10),
                        child: Column(
                          children: [
                            buildMenuCategoryContainer(title: global.language("purchase_list"), menuItems: transactionPurchaseMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("sales_list"), menuItems: transactionSaleMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("stock_list"), menuItems: transactionStockMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("payment_list"), menuItems: transactionPaidMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("accounting_list"), menuItems: glMenuList),
                            const SizedBox(height: 15),
                            htmlView(),
                          ],
                        ),
                      ),
                    ),
                    reportMenuList,
                    SingleChildScrollView(
                      child: Padding(
                        padding: EdgeInsets.all(10),
                        child: Column(
                          children: [
                            buildMenuCategoryContainer(title: global.language("manage_products"), menuItems: masterMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("product_settings"), menuItems: productSettingMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("manage_database"), menuItems: masterProductMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("manage_customers_creditors"), menuItems: customerMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("sales_settings"), menuItems: newMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("processing"), menuItems: masterDataMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("approval"), menuItems: approvalMenuList),
                            SizedBox(height: 15),
                            buildMenuCategoryContainer(title: global.language("import_export_data"), menuItems: importExportMenuList),
                            SizedBox(height: 15),
                            if (global.posVersion == global.PosVersionEnum.restaurant) buildMenuCategoryContainer(title: global.language("restaurant_cafe"), menuItems: restaurantMenuList),
                            const SizedBox(height: 15),
                            htmlView(),
                          ],
                        ),
                      ),
                    ),
                    SingleChildScrollView(
                      child: Padding(
                        padding: EdgeInsets.all(10),
                        child: Column(
                          children: [
                            buildMenuCategoryContainer(title: global.language("system_settings"), menuItems: configMenuList),
                            const SizedBox(height: 15),
                            htmlView(),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        bottomNavigationBar: BottomNavigationBar(
          type: BottomNavigationBarType.fixed,
          elevation: 10.0,
          currentIndex: global.activeIndexMenu,
          backgroundColor: primaryDarkColor.withOpacity(0.95),
          selectedItemColor: Colors.white,
          unselectedItemColor: Colors.white.withOpacity(.60),
          selectedLabelStyle: const TextStyle(
            fontWeight: FontWeight.bold,
            fontSize: 14,
            letterSpacing: 0.5,
            shadows: [Shadow(offset: Offset(1.0, 1.0), blurRadius: 3.0, color: Color.fromARGB(150, 0, 0, 0))],
          ),
          unselectedLabelStyle: const TextStyle(fontWeight: FontWeight.normal, fontSize: 12),
          onTap: (value) async {
            // Log เพื่อติดตามการเปลี่ยน tab
            String tabName = '';
            switch (value) {
              case 0:
                tabName = global.language('transaction_list');
                break;
              case 1:
                tabName = global.language('report');
                break;
              case 2:
                tabName = global.language('menu_master');
                break;
              case 3:
                tabName = global.language('setting');
                break;
            }
            AppLogger.info('📑 TAB CHANGE: เปลี่ยนไปยังแท็บ "$tabName" (index: $value)');

            loadHtmlContent(value);
            setState(() {
              global.activeIndexMenu = value;
              mainTabController.animateTo(value);
            });
          },
          items: [
            BottomNavigationBarItem(
              label: global.language("menu_transaction"),
              icon: const Icon(
                Icons.transcribe,
                shadows: [Shadow(offset: Offset(1.0, 1.0), blurRadius: 3.0, color: Color.fromARGB(150, 0, 0, 0))],
              ),
            ),
            BottomNavigationBarItem(
              label: global.language("menu_report"),
              icon: const Icon(
                Icons.assessment,
                shadows: [Shadow(offset: Offset(1.0, 1.0), blurRadius: 3.0, color: Color.fromARGB(150, 0, 0, 0))],
              ),
            ),
            BottomNavigationBarItem(
              label: global.language("menu_master"),
              icon: const Icon(
                Icons.inventory_2,
                shadows: [Shadow(offset: Offset(1.0, 1.0), blurRadius: 3.0, color: Color.fromARGB(150, 0, 0, 0))],
              ),
            ),
            BottomNavigationBarItem(
              label: global.language("menu_setup"),
              icon: const Icon(
                Icons.settings,
                shadows: [Shadow(offset: Offset(1.0, 1.0), blurRadius: 3.0, color: Color.fromARGB(150, 0, 0, 0))],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

