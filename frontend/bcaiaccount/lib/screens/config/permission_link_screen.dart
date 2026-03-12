import 'package:flutter/material.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/user/user_bloc.dart';
import 'package:smlaicloud/model/user_model.dart';
import 'package:smlaicloud/model/permission_model.dart';
import 'package:smlaicloud/services/permission_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// หน้าจอเชื่อมสิทธิ์ - เชื่อมผู้ใช้กับรหัสสิทธิ์
class PermissionLinkScreen extends StatefulWidget {
  const PermissionLinkScreen({super.key});

  @override
  State<PermissionLinkScreen> createState() => _PermissionLinkScreenState();
}

class _PermissionLinkScreenState extends State<PermissionLinkScreen>
    with global.ThemeRefreshMixin {
  List<UserModel> _userList = [];
  List<PermissionDefinitionModel> _permissionDefList = [];
  List<EmployeePermissionModel> _userPermissionList = [];
  bool _isLoading = true;
  String _searchText = '';

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    try {
      // โหลดรายการผู้ใช้งาน
      context.read<UserBloc>().add(const UserLoadList(offset: 0, limit: 1000, search: ""));

      // โหลดรายการกำหนดสิทธิ์ที่ใช้งาน
      final permDefs = await permissionService.getPermissionDefinitions(global.getShopId());
      final activePermDefs = permDefs.where((p) => p.isActive).toList();

      // โหลดรายการเชื่อมสิทธิ์ทั้งหมด
      final userPerms = await permissionService.getEmployeePermissions(global.getShopId());

      setState(() {
        _permissionDefList = activePermDefs;
        _userPermissionList = userPerms;
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    }
  }

  List<UserModel> get _filteredUserList {
    if (_searchText.isEmpty) return _userList;
    return _userList.where((u) {
      return u.username.toLowerCase().contains(_searchText.toLowerCase()) ||
          (u.email?.toLowerCase().contains(_searchText.toLowerCase()) ?? false);
    }).toList();
  }

  /// ดึงข้อมูลเชื่อมสิทธิ์ของผู้ใช้
  EmployeePermissionModel? _getUserPermission(String username) {
    try {
      return _userPermissionList.firstWhere((p) => p.employeeCode == username);
    } catch (_) {
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<UserBloc, UserState>(
      listener: (context, state) {
        if (state is UserLoadSuccess) {
          setState(() {
            _userList = state.users;
          });
        }
      },
      child: Scaffold(
        appBar: AppBar(
          title: Text(global.language('link_permission')),
          actions: [
            IconButton(
              icon: Icon(Icons.refresh),
              onPressed: _loadData,
              tooltip: global.language('database_master_info.refresh'),
            ),
            const ManualButton(path: 'settings-permission-link'),
          ],
        ),
        body: Column(
          children: [
            // Search bar
            Padding(
              padding: EdgeInsets.all(8.0),
              child: TextField(
                decoration: InputDecoration(
                  hintText: global.language('search_user_hint'),
                  prefixIcon: Icon(Icons.search),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                  filled: true,
                  fillColor: global.theme.formFillColor,
                ),
                onChanged: (value) => setState(() => _searchText = value),
              ),
            ),

            // Info banner
            if (_permissionDefList.isEmpty)
              Container(
                margin: const EdgeInsets.symmetric(horizontal: 8),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.orange[50],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.orange[200]!),
                ),
                child: Row(
                  children: [
                    Icon(Icons.warning_amber, color: Colors.orange[700]),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        global.language('no_available_permissions_hint'),
                        style: TextStyle(fontSize: 13),
                      ),
                    ),
                  ],
                ),
              ),

            // User list
            _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _filteredUserList.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.people_outline, size: 64, color: global.theme.iconSecondaryColor),
                            const SizedBox(height: 16),
                            Text(
                              global.language('no_users_found'),
                              style: TextStyle(fontSize: 16, color: global.theme.textSecondaryColor),
                            ),
                          ],
                        ),
                      )
                    : ListView.builder(
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        itemCount: _filteredUserList.length,
                        itemBuilder: (context, index) {
                          final user = _filteredUserList[index];
                          return _buildUserCard(user);
                        },
                      ),

          ],
        ),
      ),
    );
  }

  Widget _buildUserCard(UserModel user) {
    final userPerm = _getUserPermission(user.username);
    final permCount = userPerm?.permissionCodes.length ?? 0;
    final roleText = _getRoleText(user.role);

    return Card(
      margin: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: _getRoleColor(user.role),
          child: Icon(
            Icons.person,
            color: global.theme.onPrimaryColor,
            size: 20,
          ),
        ),
        title: Text(
          user.username,
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (user.email != null && user.email!.isNotEmpty)
              Text(user.email!, style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: _getRoleColor(user.role).withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    roleText,
                    style: TextStyle(fontSize: 10, color: _getRoleColor(user.role)),
                  ),
                ),
                const SizedBox(width: 8),
                if (permCount > 0)
                  Container(
                    padding: EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: global.theme.infoHighlightColor,
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      '$permCount ${global.language("permission")}',
                      style: TextStyle(fontSize: 10, color: Colors.blue[700]),
                    ),
                  ),
              ],
            ),
          ],
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            // ปุ่มตรวจสอบสิทธิ์ - แสดงเฉพาะเมื่อมีสิทธิ์เชื่อมอยู่
            if (permCount > 0)
              IconButton(
                icon: Icon(Icons.visibility, color: global.theme.positiveHighlightTextColor),
                onPressed: () => _showPermissionPreviewDialog(user, userPerm!),
                tooltip: global.language('check_permission'),
              ),
            // ปุ่มแก้ไขสิทธิ์
            IconButton(
              icon: Icon(Icons.edit, color: global.theme.infoHighlightTextColor),
              onPressed: () => _showEditDialog(user, userPerm),
              tooltip: global.language('edit_permission'),
            ),
          ],
        ),
        onTap: () => _showEditDialog(user, userPerm),
      ),
    );
  }

  String _getRoleText(int role) {
    switch (role) {
      case 0:
        return global.language('role_general_user');
      case 1:
        return global.language('role_admin');
      case 2:
        return global.language('role_super_admin');
      default:
        return global.language('alert_not_specified');
    }
  }

  Color _getRoleColor(int role) {
    switch (role) {
      case 0:
        return global.theme.iconSecondaryColor;
      case 1:
        return global.theme.infoHighlightTextColor;
      case 2:
        return Colors.purple;
      default:
        return global.theme.iconSecondaryColor;
    }
  }

  Future<void> _showEditDialog(UserModel user, EmployeePermissionModel? existingPerm) async {
    // สร้าง or ใช้ permission ที่มีอยู่
    final userPerm = existingPerm?.copyWith() ??
        EmployeePermissionModel(
          shopid: global.getShopId(),
          employeeCode: user.username,
          employeeName: user.username,
          createdBy: global.profileData.username ?? '',
        );

    final result = await showDialog<EmployeePermissionModel>(
      context: context,
      barrierDismissible: false,
      builder: (context) => _PermissionLinkEditDialog(
        userPermission: userPerm,
        permissionDefinitions: _permissionDefList,
        userName: user.username,
      ),
    );

    if (result != null) {
      setState(() => _isLoading = true);

      bool success;
      // ถ้าไม่มีสิทธิ์เลือก ให้ลบ record ออก
      if (result.permissionCodes.isEmpty) {
        success = await permissionService.deleteEmployeePermission(
          global.getShopId(),
          user.username,
        );
      } else {
        success = await permissionService.saveEmployeePermission(result);
      }

      if (success) {
        await _loadData();
        if (mounted) {
          global.showSuccessSnackBar(context, global.language('save_permission_success'));
        }
      } else {
        setState(() => _isLoading = false);
        if (mounted) {
          global.showErrorSnackBar(context, global.language('error_saving'));
        }
      }
    }
  }

  /// แสดง popup ตรวจสอบสิทธิ์รวมของผู้ใช้
  Future<void> _showPermissionPreviewDialog(UserModel user, EmployeePermissionModel userPerm) async {
    // รวมสิทธิ์ทั้งหมด (OR) จาก permission codes ที่เชื่อมไว้
    final Map<String, Map<String, ScreenPermissionModel>> combinedPermissions = {};

    for (final permCode in userPerm.permissionCodes) {
      // หา definition ของแต่ละ permission code
      final permDef = _permissionDefList.firstWhere(
        (p) => p.permissionCode == permCode,
        orElse: () => PermissionDefinitionModel(shopid: '', permissionCode: ''),
      );

      if (permDef.permissionCode.isEmpty) continue;

      // OR สิทธิ์เข้าด้วยกัน
      for (final branch in permDef.branches) {
        combinedPermissions.putIfAbsent(branch.branchCode, () => {});

        for (final screen in branch.screens) {
          final existing = combinedPermissions[branch.branchCode]![screen.screenCode];

          if (existing == null) {
            combinedPermissions[branch.branchCode]![screen.screenCode] = screen.copyWith();
          } else {
            // OR สิทธิ์แต่ละตัว
            combinedPermissions[branch.branchCode]![screen.screenCode] = ScreenPermissionModel(
              screenCode: screen.screenCode,
              screenName: screen.screenName.isNotEmpty ? screen.screenName : existing.screenName,
              canView: existing.canView || screen.canView,
              canAdd: existing.canAdd || screen.canAdd,
              canEditOwn: existing.canEditOwn || screen.canEditOwn,
              canEditAll: existing.canEditAll || screen.canEditAll,
              canDeleteOwn: existing.canDeleteOwn || screen.canDeleteOwn,
              canDeleteAll: existing.canDeleteAll || screen.canDeleteAll,
              canPrint: existing.canPrint || screen.canPrint,
            );
          }
        }
      }
    }

    await showDialog(
      context: context,
      builder: (context) => _PermissionPreviewDialog(
        userName: user.username,
        permissionCodes: userPerm.permissionCodes,
        combinedPermissions: combinedPermissions,
        permissionDefList: _permissionDefList,
      ),
    );
  }
}

/// Dialog แสดงสิทธิ์รวมของผู้ใช้
class _PermissionPreviewDialog extends StatelessWidget {
  final String userName;
  final List<String> permissionCodes;
  final Map<String, Map<String, ScreenPermissionModel>> combinedPermissions;
  final List<PermissionDefinitionModel> permissionDefList;

  const _PermissionPreviewDialog({
    required this.userName,
    required this.permissionCodes,
    required this.combinedPermissions,
    required this.permissionDefList,
  });

  /// ชื่อหมวดหมู่เป็นภาษาไทย และไอคอน
  static final Map<String, Map<String, dynamic>> _categoryInfo = {
    'system': {'name': global.language('system_setting'), 'icon': Icons.settings, 'color': Colors.blueGrey},
    'user': {'name': global.language('users_and_employees'), 'icon': Icons.people, 'color': Colors.indigo},
    'partner': {'name': global.language('customers_and_sellers'), 'icon': Icons.handshake, 'color': Colors.teal},
    'sale_setting': {'name': global.language('sale_settings'), 'icon': Icons.point_of_sale, 'color': Colors.amber},
    'product': {'name': global.language('product'), 'icon': Icons.inventory_2, 'color': Colors.deepPurple},
    'purchase': {'name': global.language('purchase_transactions'), 'icon': Icons.shopping_cart, 'color': global.theme.warningHighlightTextColor},
    'sale': {'name': global.language('sale_transactions'), 'icon': Icons.storefront, 'color': global.theme.positiveHighlightTextColor},
    'stock': {'name': global.language('stock'), 'icon': Icons.warehouse, 'color': Colors.brown},
    'finance': {'name': global.language('finance'), 'icon': Icons.account_balance, 'color': global.theme.infoHighlightTextColor},
    'report': {'name': global.language('report'), 'icon': Icons.bar_chart, 'color': Colors.cyan},
    'restaurant': {'name': global.language('restaurant'), 'icon': Icons.restaurant, 'color': global.theme.negativeHighlightTextColor},
    'check': {'name': global.language('audit'), 'icon': Icons.fact_check, 'color': Colors.pink},
    'master': {'name': global.language('menu_master'), 'icon': Icons.dataset, 'color': Colors.lime},
    'import_export': {'name': global.language('import_export'), 'icon': Icons.import_export, 'color': Colors.deepOrange},
    'ai': {'name': global.language('ai_smart_system'), 'icon': Icons.smart_toy, 'color': Colors.purple},
  };

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Container(
        width: MediaQuery.of(context).size.width * 0.85,
        height: MediaQuery.of(context).size.height * 0.8,
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Row(
              children: [
                Icon(Icons.visibility, color: global.theme.positiveHighlightTextColor),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        global.language('check_combined_permission'),
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                      ),
                      Text(
                        '${global.language("user_label")}: $userName',
                        style: TextStyle(fontSize: 13, color: global.theme.textSecondaryColor),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                ),
              ],
            ),
            const Divider(),

            // สิทธิ์ที่เชื่อม
            Container(
              padding: EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: global.theme.infoHighlightColor,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                children: [
                  Icon(Icons.link, color: Colors.blue[700], size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Wrap(
                      spacing: 4,
                      runSpacing: 4,
                      children: [
                        Text('${global.language("linked_permissions")}: ', style: TextStyle(fontSize: 12)),
                        ...permissionCodes.map((code) {
                          final permDef = permissionDefList.firstWhere(
                            (p) => p.permissionCode == code,
                            orElse: () => PermissionDefinitionModel(shopid: '', permissionCode: code),
                          );
                          return Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(
                              color: Colors.blue[100],
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Text(
                              permDef.permissionName.isNotEmpty ? permDef.permissionName : code,
                              style: TextStyle(fontSize: 11, color: Colors.blue[800]),
                            ),
                          );
                        }),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),

            // รายการสิทธิ์รวม
            Expanded(
              child: combinedPermissions.isEmpty
                  ? Center(child: Text(global.language('no_permission_title')))
                  : SingleChildScrollView(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: combinedPermissions.entries.map((branchEntry) {
                          return _buildBranchSection(branchEntry.key, branchEntry.value);
                        }).toList(),
                      ),
                    ),
            ),

            // Actions
            Divider(),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                ElevatedButton(
                  onPressed: () => Navigator.pop(context),
                  child: Text(global.language('close')),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBranchSection(String branchCode, Map<String, ScreenPermissionModel> screens) {
    // จัดกลุ่มหน้าจอตามหมวดหมู่
    final Map<String, List<ScreenPermissionModel>> screensByCategory = {};

    for (final screen in screens.values) {
      // หา category จาก allScreenDefinitions
      String category = 'other';
      for (final def in allScreenDefinitions) {
        if (def.code == screen.screenCode) {
          category = def.category;
          break;
        }
      }
      screensByCategory.putIfAbsent(category, () => []).add(screen);
    }

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Branch header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.indigo[50],
              borderRadius: const BorderRadius.vertical(top: Radius.circular(4)),
            ),
            child: Row(
              children: [
                Icon(Icons.store, size: 18, color: Colors.indigo),
                const SizedBox(width: 8),
                Text(
                  '${global.language("branch")}: $branchCode',
                  style: TextStyle(fontWeight: FontWeight.bold, color: Colors.indigo),
                ),
                const Spacer(),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.indigo[100],
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Text(
                    '${screens.length} ${global.language("screen")}',
                    style: TextStyle(fontSize: 11, color: Colors.indigo[800]),
                  ),
                ),
              ],
            ),
          ),
          // Screen permissions table grouped by category
          Padding(
            padding: EdgeInsets.all(8),
            child: Column(
              children: [
                // Header row
                Container(
                  padding: EdgeInsets.symmetric(vertical: 6, horizontal: 4),
                  decoration: BoxDecoration(
                    color: global.theme.surfaceColor,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Row(
                    children: [
                      Expanded(flex: 3, child: Text(global.language('screen'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11))),
                      Expanded(child: Center(child: Text(global.language('view'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10)))),
                      Expanded(child: Center(child: Text(global.language('add'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10)))),
                      Expanded(child: Center(child: Text(global.language('edit_own'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 9)))),
                      Expanded(child: Center(child: Text(global.language('edit_all'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 9)))),
                      Expanded(child: Center(child: Text(global.language('delete_own'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 9)))),
                      Expanded(child: Center(child: Text(global.language('clear_all'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 9)))),
                      Expanded(child: Center(child: Text(global.language('print'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10)))),
                    ],
                  ),
                ),
                // Category sections
                ...screensByCategory.entries.map((entry) {
                  final categoryCode = entry.key;
                  final categoryScreens = entry.value;
                  final categoryInfo = _categoryInfo[categoryCode] ?? {
                    'name': categoryCode,
                    'icon': Icons.folder,
                    'color': global.theme.iconSecondaryColor,
                  };
                  return _buildCategorySection(
                    categoryCode: categoryCode,
                    categoryName: categoryInfo['name'] as String,
                    categoryIcon: categoryInfo['icon'] as IconData,
                    categoryColor: categoryInfo['color'] as Color,
                    screens: categoryScreens,
                  );
                }),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// สร้าง section สำหรับแต่ละหมวดหมู่
  Widget _buildCategorySection({
    required String categoryCode,
    required String categoryName,
    required IconData categoryIcon,
    required Color categoryColor,
    required List<ScreenPermissionModel> screens,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // หัวข้อหมวดหมู่
        Container(
          margin: const EdgeInsets.only(top: 8, bottom: 4),
          padding: const EdgeInsets.symmetric(vertical: 6, horizontal: 10),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [
                categoryColor.withValues(alpha: 0.2),
                categoryColor.withValues(alpha: 0.05),
              ],
            ),
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: categoryColor.withValues(alpha: 0.3)),
          ),
          child: Row(
            children: [
              Icon(categoryIcon, color: categoryColor, size: 16),
              const SizedBox(width: 6),
              Text(
                categoryName,
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: categoryColor.withValues(alpha: 0.9),
                  fontSize: 12,
                ),
              ),
              const SizedBox(width: 6),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                decoration: BoxDecoration(
                  color: categoryColor.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  '${screens.length}',
                  style: TextStyle(
                    fontSize: 10,
                    color: categoryColor,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
        ),
        // รายการหน้าจอในหมวดหมู่
        ...screens.map((screen) => _buildScreenRow(screen)),
      ],
    );
  }

  Widget _buildScreenRow(ScreenPermissionModel screen) {
    return Container(
      padding: EdgeInsets.symmetric(vertical: 4, horizontal: 4),
      decoration: BoxDecoration(
        border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor)),
      ),
      child: Row(
        children: [
          Expanded(
            flex: 3,
            child: Text(
              screen.screenName.isNotEmpty ? screen.screenName : screen.screenCode,
              style: TextStyle(fontSize: 12),
            ),
          ),
          Expanded(child: _buildPermissionIcon(screen.canView)),
          Expanded(child: _buildPermissionIcon(screen.canAdd)),
          Expanded(child: _buildPermissionIcon(screen.canEditOwn)),
          Expanded(child: _buildPermissionIcon(screen.canEditAll)),
          Expanded(child: _buildPermissionIcon(screen.canDeleteOwn)),
          Expanded(child: _buildPermissionIcon(screen.canDeleteAll)),
          Expanded(child: _buildPermissionIcon(screen.canPrint)),
        ],
      ),
    );
  }

  Widget _buildPermissionIcon(bool hasPermission) {
    return Center(
      child: Icon(
        hasPermission ? Icons.check_circle : Icons.cancel,
        color: hasPermission ? global.theme.positiveHighlightTextColor : global.theme.dividerBorderColor,
        size: 18,
      ),
    );
  }
}

/// Dialog สำหรับแก้ไขการเชื่อมสิทธิ์ของผู้ใช้
class _PermissionLinkEditDialog extends StatefulWidget {
  final EmployeePermissionModel userPermission;
  final List<PermissionDefinitionModel> permissionDefinitions;
  final String userName;

  const _PermissionLinkEditDialog({
    required this.userPermission,
    required this.permissionDefinitions,
    required this.userName,
  });

  @override
  State<_PermissionLinkEditDialog> createState() => _PermissionLinkEditDialogState();
}

class _PermissionLinkEditDialogState extends State<_PermissionLinkEditDialog>
    with global.ThemeRefreshMixin {
  late EmployeePermissionModel _userPermission;
  late Set<String> _selectedCodes;

  @override
  void initState() {
    super.initState();
    _userPermission = widget.userPermission.copyWith();
    _selectedCodes = Set.from(_userPermission.permissionCodes);
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Container(
        width: MediaQuery.of(context).size.width * 0.6,
        constraints: const BoxConstraints(maxWidth: 600, maxHeight: 500),
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            // Header
            Row(
              children: [
                Icon(Icons.link, color: global.theme.infoHighlightTextColor),
                SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        global.language('link_permission'),
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                      ),
                      Text(
                        '${global.language("user_label")}: ${widget.userName}',
                        style: TextStyle(fontSize: 13, color: global.theme.textSecondaryColor),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                ),
              ],
            ),
            const Divider(),

            // Info
            Container(
              padding: EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: global.theme.infoHighlightColor,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                children: [
                  Icon(Icons.info_outline, color: Colors.blue[700], size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      global.language('select_permission_info'),
                      style: TextStyle(fontSize: 12),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),

            // Permission list
            Expanded(
              child: widget.permissionDefinitions.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.warning_amber, size: 48, color: global.theme.warningHighlightTextColor),
                          const SizedBox(height: 8),
                          Text(global.language('no_available_permissions')),
                          Text(
                            global.language('go_to_permission_definition'),
                            style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
                          ),
                        ],
                      ),
                    )
                  : ListView.builder(
                      itemCount: widget.permissionDefinitions.length,
                      itemBuilder: (context, index) {
                        final perm = widget.permissionDefinitions[index];
                        final isSelected = _selectedCodes.contains(perm.permissionCode);
                        return _buildPermissionItem(perm, isSelected);
                      },
                    ),
            ),

            // Actions
            Divider(),
            Row(
              children: [
                // เลือกทั้งหมด / ยกเลิกทั้งหมด
                TextButton.icon(
                  onPressed: () {
                    setState(() {
                      if (_selectedCodes.length == widget.permissionDefinitions.length) {
                        _selectedCodes.clear();
                      } else {
                        _selectedCodes = Set.from(
                          widget.permissionDefinitions.map((p) => p.permissionCode),
                        );
                      }
                    });
                  },
                  icon: Icon(
                    _selectedCodes.length == widget.permissionDefinitions.length
                        ? Icons.deselect
                        : Icons.select_all,
                    size: 18,
                  ),
                  label: Text(
                    _selectedCodes.length == widget.permissionDefinitions.length
                        ? global.language('cancel_all')
                        : global.language('select_all'),
                  ),
                ),
                Spacer(),
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: Text(global.language('cancel')),
                ),
                const SizedBox(width: 8),
                ElevatedButton.icon(
                  onPressed: () {
                    _userPermission.permissionCodes = _selectedCodes.toList();
                    Navigator.pop(context, _userPermission);
                  },
                  icon: Icon(Icons.save),
                  label: Text('${global.language("save")} (${_selectedCodes.length} ${global.language("permission")})'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPermissionItem(PermissionDefinitionModel perm, bool isSelected) {
    // นับจำนวนสาขาและหน้าจอ
    final branchCount = perm.branches.length;
    final screenCount = perm.branches.fold<int>(
      0,
      (sum, branch) => sum + branch.screens.length,
    );

    return Card(
      margin: EdgeInsets.only(bottom: 4),
      color: isSelected ? global.theme.infoHighlightColor : null,
      child: CheckboxListTile(
        value: isSelected,
        onChanged: (value) {
          setState(() {
            if (value == true) {
              _selectedCodes.add(perm.permissionCode);
            } else {
              _selectedCodes.remove(perm.permissionCode);
            }
          });
        },
        title: Text(
          perm.permissionCode,
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (perm.permissionName.isNotEmpty)
              Text(perm.permissionName, style: TextStyle(fontSize: 13)),
            Text(
              '$branchCount ${global.language("branch")}, $screenCount ${global.language("screen")}',
              style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor),
            ),
          ],
        ),
        secondary: Container(
          padding: EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: isSelected ? global.theme.infoHighlightTextColor : global.theme.dividerBorderColor,
            shape: BoxShape.circle,
          ),
          child: Icon(
            Icons.security,
            color: isSelected ? global.theme.onPrimaryColor : global.theme.textSecondaryColor,
            size: 20,
          ),
        ),
        controlAffinity: ListTileControlAffinity.leading,
      ),
    );
  }
}
