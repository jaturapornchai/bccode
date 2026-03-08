import 'package:flutter/material.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/permission_model.dart';
import 'package:smlaicloud/services/permission_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// หน้าจอกำหนดสิทธิ์ - สร้างและแก้ไขรหัสสิทธิ์
class PermissionDefinitionScreen extends StatefulWidget {
  const PermissionDefinitionScreen({super.key});

  @override
  State<PermissionDefinitionScreen> createState() => _PermissionDefinitionScreenState();
}

class _PermissionDefinitionScreenState extends State<PermissionDefinitionScreen> {
  List<PermissionDefinitionModel> _permissionList = [];
  List<CompanyBranchModel> _branchList = [];
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
      // โหลดรายการสิทธิ์
      final permissions = await permissionService.getPermissionDefinitions(global.getShopId());

      // โหลดรายการสาขา
      context.read<CompanyBranchBloc>().add(const CompanyBranchLoadList(offset: 0, limit: 100, search: ""));

      setState(() {
        _permissionList = permissions;
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    }
  }

  List<PermissionDefinitionModel> get _filteredList {
    if (_searchText.isEmpty) return _permissionList;
    return _permissionList.where((p) {
      return p.permissionCode.toLowerCase().contains(_searchText.toLowerCase()) ||
          p.permissionName.toLowerCase().contains(_searchText.toLowerCase());
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<CompanyBranchBloc, CompanyBranchState>(
      listener: (context, state) {
        if (state is CompanyBranchLoadSuccess) {
          setState(() {
            _branchList = state.companyBranch;
          });
        }
      },
      child: Scaffold(
        appBar: AppBar(
          title: Text(global.language('permission_definition')),
          actions: [
            const ManualButton(path: 'settings-permissions'),
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: _loadData,
              tooltip: global.language('database_master_info.refresh'),
            ),
            IconButton(
              icon: const Icon(Icons.add),
              onPressed: () => _showEditDialog(null),
              tooltip: global.language('add_new_permission'),
            ),
          ],
        ),
        body: Column(
          children: [
            // Search bar
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: TextField(
                decoration: InputDecoration(
                  hintText: global.language('search_permission_hint'),
                  prefixIcon: const Icon(Icons.search),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                  filled: true,
                  fillColor: Colors.grey[100],
                ),
                onChanged: (value) => setState(() => _searchText = value),
              ),
            ),

              // List
              _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _filteredList.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.security, size: 64, color: Colors.grey[400]),
                              const SizedBox(height: 16),
                              Text(
                                global.language('no_permission_list'),
                                style: TextStyle(fontSize: 16, color: Colors.grey[600]),
                              ),
                              SizedBox(height: 8),
                              ElevatedButton.icon(
                                onPressed: () => _showEditDialog(null),
                                icon: Icon(Icons.add),
                                label: Text(global.language('add_new_permission')),
                              ),
                            ],
                          ),
                        )
                      : ListView.builder(
                          shrinkWrap: true,
                          physics: const NeverScrollableScrollPhysics(),
                          itemCount: _filteredList.length,
                          itemBuilder: (context, index) {
                            final permission = _filteredList[index];
                            return _buildPermissionCard(permission);
                          },
                        ),

          ],
        ),
      ),
    );
  }

  Widget _buildPermissionCard(PermissionDefinitionModel permission) {
    final branchCount = permission.branches.length;
    final screenCount = permission.branches.fold<int>(
      0,
      (sum, branch) => sum + branch.screens.length,
    );

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: permission.isActive ? Colors.green : Colors.grey,
          child: Icon(
            Icons.security,
            color: Colors.white,
            size: 20,
          ),
        ),
        title: Text(
          permission.permissionCode,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (permission.permissionName.isNotEmpty)
              Text(permission.permissionName),
            Text(
              '$branchCount ${global.language("branch")} $screenCount ${global.language("screen")}',
              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
            ),
          ],
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            // สถานะ
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: permission.isActive ? Colors.green[100] : Colors.grey[200],
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                permission.isActive ? global.language('active') : global.language('alert_disabled'),
                style: TextStyle(
                  fontSize: 11,
                  color: permission.isActive ? Colors.green[800] : Colors.grey[600],
                ),
              ),
            ),
            const SizedBox(width: 8),
            // คัดลอก
            IconButton(
              icon: const Icon(Icons.copy, color: Colors.orange),
              onPressed: () => _showCopyDialog(permission),
              tooltip: global.language('copy_permission'),
            ),
            // แก้ไข
            IconButton(
              icon: const Icon(Icons.edit, color: Colors.blue),
              onPressed: () => _showEditDialog(permission),
              tooltip: global.language('edit'),
            ),
            // ลบ
            IconButton(
              icon: const Icon(Icons.delete, color: Colors.red),
              onPressed: () => _confirmDelete(permission),
              tooltip: global.language('delete'),
            ),
          ],
        ),
        onTap: () => _showEditDialog(permission),
      ),
    );
  }

  Future<void> _showEditDialog(PermissionDefinitionModel? permission) async {
    final isNew = permission == null;
    final editPermission = permission?.copyWith() ??
        PermissionDefinitionModel(
          shopid: global.getShopId(),
          permissionCode: '',
          permissionName: '',
          createdBy: global.profileData.username ?? '',
        );

    final result = await showDialog<PermissionDefinitionModel>(
      context: context,
      barrierDismissible: false,
      builder: (context) => _PermissionEditDialog(
        permission: editPermission,
        branchList: _branchList,
        isNew: isNew,
        existingCodes: _permissionList.map((p) => p.permissionCode).toList(),
      ),
    );

    if (result != null) {
      setState(() => _isLoading = true);

      final success = await permissionService.savePermissionDefinition(result);

      if (success) {
        await _loadData();
        if (mounted) {
          global.showSuccessSnackBar(context, isNew ? global.language('add_permission_success') : global.language('edit_permission_success'));
        }
      } else {
        setState(() => _isLoading = false);
        if (mounted) {
          global.showErrorSnackBar(context, global.language('error_saving'));
        }
      }
    }
  }

  Future<void> _confirmDelete(PermissionDefinitionModel permission) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('alert_confirm_delete')),
        content: Text('${global.language("confirm_delete_permission")} "${permission.permissionCode}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: Text(global.language('delete')),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      setState(() => _isLoading = true);

      final success = await permissionService.deletePermissionDefinition(
        global.getShopId(),
        permission.permissionCode,
      );

      if (success) {
        await _loadData();
        if (mounted) {
          global.showSuccessSnackBar(context, global.language('delete_permission_success'));
        }
      } else {
        setState(() => _isLoading = false);
        if (mounted) {
          global.showErrorSnackBar(context, global.language('error_deleting'));
        }
      }
    }
  }

  Future<void> _showCopyDialog(PermissionDefinitionModel source) async {
    final codeController = TextEditingController();
    final nameController = TextEditingController();
    final formKey = GlobalKey<FormState>();
    final existingCodes = _permissionList.map((p) => p.permissionCode).toSet();

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.copy, color: Colors.orange, size: 24),
            const SizedBox(width: 8),
            Text('${global.language("copy_permission_from")} "${source.permissionCode}"'),
          ],
        ),
        content: Form(
          key: formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(
                controller: codeController,
                decoration: InputDecoration(
                  labelText: '${global.language("new_permission_code")} *',
                  hintText: 'เช่น SALE_MANAGER',
                  border: OutlineInputBorder(),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return global.language('please_enter_permission_code');
                  }
                  if (existingCodes.contains(value.trim())) {
                    return global.language('permission_code_exists');
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: nameController,
                decoration: InputDecoration(
                  labelText: '${global.language("new_permission_name")} *',
                  hintText: 'เช่น ผู้จัดการฝ่ายขาย',
                  border: OutlineInputBorder(),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return global.language('please_enter_permission_name');
                  }
                  return null;
                },
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton.icon(
            onPressed: () {
              if (formKey.currentState!.validate()) {
                Navigator.pop(context, true);
              }
            },
            icon: Icon(Icons.copy),
            label: Text(global.language('chatbot_copy')),
            style: ElevatedButton.styleFrom(backgroundColor: Colors.orange),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await _copyPermission(
        source,
        codeController.text.trim(),
        nameController.text.trim(),
      );
    }

    codeController.dispose();
    nameController.dispose();
  }

  Future<void> _copyPermission(
    PermissionDefinitionModel source,
    String newCode,
    String newName,
  ) async {
    setState(() => _isLoading = true);

    try {
      final success = await permissionService.copyPermissionDefinition(
        global.getShopId(),
        source.permissionCode,
        newCode,
        newName,
      );

      if (success) {
        await _loadData();
        if (mounted) {
          global.showSuccessSnackBar(context, '${global.language("copy_permission_success")} "${source.permissionCode}" → "$newCode"');
        }
      } else {
        setState(() => _isLoading = false);
        if (mounted) {
          global.showErrorSnackBar(context, global.language('error_copying'));
        }
      }
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    }
  }
}

/// Dialog สำหรับแก้ไขสิทธิ์
class _PermissionEditDialog extends StatefulWidget {
  final PermissionDefinitionModel permission;
  final List<CompanyBranchModel> branchList;
  final bool isNew;
  final List<String> existingCodes;

  const _PermissionEditDialog({
    required this.permission,
    required this.branchList,
    required this.isNew,
    required this.existingCodes,
  });

  @override
  State<_PermissionEditDialog> createState() => _PermissionEditDialogState();
}

class _PermissionEditDialogState extends State<_PermissionEditDialog> {
  late PermissionDefinitionModel _permission;
  final _codeController = TextEditingController();
  final _nameController = TextEditingController();
  final _descController = TextEditingController();
  String? _codeError;

  @override
  void initState() {
    super.initState();
    _permission = widget.permission.copyWith();
    _codeController.text = _permission.permissionCode;
    _nameController.text = _permission.permissionName;
    _descController.text = _permission.description;
  }

  @override
  void dispose() {
    _codeController.dispose();
    _nameController.dispose();
    _descController.dispose();
    super.dispose();
  }

  void _validateCode(String value) {
    setState(() {
      if (value.isEmpty) {
        _codeError = global.language('please_enter_permission_code');
      } else if (widget.isNew && widget.existingCodes.contains(value)) {
        _codeError = global.language('permission_code_exists');
      } else {
        _codeError = null;
      }
    });
  }

  /// สร้างรหัสสิทธิ์อัตโนมัติ (ใช้ GUID format)
  String _generateAutoCode() {
    final now = DateTime.now();
    // สร้างรหัสจากวันเวลา + random เพื่อให้ไม่ซ้ำ
    final code = 'PERM_${now.year}${now.month.toString().padLeft(2, '0')}${now.day.toString().padLeft(2, '0')}_${now.millisecondsSinceEpoch.toString().substring(7)}';
    return code.toUpperCase();
  }

  /// สร้างปุ่มลัดสำหรับกำหนดสิทธิ์
  Widget _buildQuickButton({
    required IconData icon,
    required String label,
    required Color color,
    required VoidCallback onPressed,
  }) {
    return Material(
      color: color.withValues(alpha: 0.1),
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: color.withValues(alpha: 0.3)),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(icon, size: 14, color: color),
              const SizedBox(width: 4),
              Text(
                label,
                style: TextStyle(fontSize: 11, color: color, fontWeight: FontWeight.w500),
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// เลือกหน้าจอทั้งหมดในสาขา
  void _selectAllScreens(BranchPermissionModel branchPerm) {
    setState(() {
      for (var screenDef in allScreenDefinitions) {
        if (!branchPerm.screens.any((s) => s.screenCode == screenDef.code)) {
          branchPerm.screens.add(ScreenPermissionModel(
            screenCode: screenDef.code,
            screenName: screenDef.name,
            canView: true,
          ));
        }
      }
    });
  }

  /// ยกเลิกเลือกหน้าจอทั้งหมดในสาขา
  void _deselectAllScreens(BranchPermissionModel branchPerm) {
    setState(() {
      branchPerm.screens.clear();
    });
  }

  /// กำหนดสิทธิ์ทั้งหมดตามประเภท
  void _setAllPermission(BranchPermissionModel branchPerm, String permType, bool value) {
    setState(() {
      for (var screen in branchPerm.screens) {
        switch (permType) {
          case 'view':
            screen.canView = value;
            break;
          case 'add':
            screen.canAdd = value;
            break;
          case 'editOwn':
            screen.canEditOwn = value;
            break;
          case 'editAll':
            screen.canEditAll = value;
            break;
          case 'deleteOwn':
            screen.canDeleteOwn = value;
            break;
          case 'deleteAll':
            screen.canDeleteAll = value;
            break;
          case 'print':
            screen.canPrint = value;
            break;
        }
      }
    });
  }

  /// เปิดทุกสิทธิ์ (Full Access)
  void _setFullAccess(BranchPermissionModel branchPerm) {
    setState(() {
      // เลือกทุกหน้าจอก่อน
      for (var screenDef in allScreenDefinitions) {
        if (!branchPerm.screens.any((s) => s.screenCode == screenDef.code)) {
          branchPerm.screens.add(ScreenPermissionModel(
            screenCode: screenDef.code,
            screenName: screenDef.name,
          ));
        }
      }
      // เปิดทุกสิทธิ์
      for (var screen in branchPerm.screens) {
        screen.canView = true;
        screen.canAdd = true;
        screen.canEditOwn = true;
        screen.canEditAll = true;
        screen.canDeleteOwn = true;
        screen.canDeleteAll = true;
        screen.canPrint = true;
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Container(
        width: MediaQuery.of(context).size.width * 0.9,
        height: MediaQuery.of(context).size.height * 0.85,
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Row(
              children: [
                const Icon(Icons.security, color: Colors.blue),
                SizedBox(width: 8),
                Text(
                  widget.isNew ? global.language('add_new_permission') : global.language('edit_permission'),
                  style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                ),
                const Spacer(),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                ),
              ],
            ),
            const Divider(),

            // Form fields
            Expanded(
              child: SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // รหัสสิทธิ์
                    Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _codeController,
                            enabled: widget.isNew,
                            decoration: InputDecoration(
                              labelText: '${global.language("permission_code_required")} *',
                              hintText: 'เช่น ADMIN_FULL, SALE_ONLY',
                              errorText: _codeError,
                              border: const OutlineInputBorder(),
                            ),
                            onChanged: (value) {
                              _permission.permissionCode = value.toUpperCase();
                              _validateCode(value);
                            },
                          ),
                        ),
                        // ปุ่มสร้างรหัสอัตโนมัติ
                        if (widget.isNew) ...[
                          const SizedBox(width: 8),
                          Tooltip(
                            message: global.language('auto_generate_code'),
                            child: ElevatedButton.icon(
                              onPressed: () {
                                final autoCode = _generateAutoCode();
                                setState(() {
                                  _codeController.text = autoCode;
                                  _permission.permissionCode = autoCode;
                                  _validateCode(autoCode);
                                });
                              },
                              icon: const Icon(Icons.auto_awesome, size: 18),
                              label: Text(global.language('auto')),
                              style: ElevatedButton.styleFrom(
                                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 16),
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                    const SizedBox(height: 12),

                    // ชื่อสิทธิ์
                    TextField(
                      controller: _nameController,
                      decoration: InputDecoration(
                        labelText: global.language('permission_name_label'),
                        hintText: 'เช่น แอดมินเต็มรูปแบบ',
                        border: OutlineInputBorder(),
                      ),
                      onChanged: (value) => _permission.permissionName = value,
                    ),
                    const SizedBox(height: 12),

                    // คำอธิบาย
                    TextField(
                      controller: _descController,
                      maxLines: 2,
                      decoration: InputDecoration(
                        labelText: global.language('form_design_description'),
                        border: OutlineInputBorder(),
                      ),
                      onChanged: (value) => _permission.description = value,
                    ),
                    const SizedBox(height: 12),

                    // สถานะใช้งาน
                    SwitchListTile(
                      title: Text(global.language('active_status')),
                      value: _permission.isActive,
                      onChanged: (value) => setState(() => _permission.isActive = value),
                    ),
                    const Divider(),

                    // สาขาและหน้าจอ
                    Text(
                      global.language('set_permissions_by_branch'),
                      style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 8),

                    // Branch list
                    ...widget.branchList.map((branch) {
                      final branchPerm = _permission.branches.firstWhere(
                        (b) => b.branchCode == branch.code,
                        orElse: () => BranchPermissionModel(
                          branchCode: branch.code,
                          branchName: global.activeLangName(branch.names),
                        ),
                      );
                      final isSelected = _permission.branches.any((b) => b.branchCode == branch.code);

                      return _buildBranchSection(branch, branchPerm, isSelected);
                    }),
                  ],
                ),
              ),
            ),

            // Actions
            Divider(),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: Text(global.language('cancel')),
                ),
                SizedBox(width: 8),
                ElevatedButton.icon(
                  onPressed: _codeError == null && _codeController.text.isNotEmpty
                      ? () {
                          _permission.permissionCode = _codeController.text.toUpperCase();
                          _permission.permissionName = _nameController.text;
                          _permission.description = _descController.text;
                          Navigator.pop(context, _permission);
                        }
                      : null,
                  icon: Icon(Icons.save),
                  label: Text(global.language('form_design_save')),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBranchSection(
    CompanyBranchModel branch,
    BranchPermissionModel branchPerm,
    bool isSelected,
  ) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ExpansionTile(
        leading: Checkbox(
          value: isSelected,
          onChanged: (value) {
            setState(() {
              if (value == true) {
                if (!_permission.branches.any((b) => b.branchCode == branch.code)) {
                  _permission.branches.add(BranchPermissionModel(
                    branchCode: branch.code,
                    branchName: global.activeLangName(branch.names),
                  ));
                }
              } else {
                _permission.branches.removeWhere((b) => b.branchCode == branch.code);
              }
            });
          },
        ),
        title: Text(
          global.activeLangName(branch.names),
          style: const TextStyle(fontWeight: FontWeight.w500),
        ),
        subtitle: Text('${global.language("code_label")}: ${branch.code}'),
        children: isSelected
            ? [
                Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: _buildScreenPermissionsTable(branchPerm),
                ),
              ]
            : [],
      ),
    );
  }

  /// ชื่อหมวดหมู่เป็นภาษาไทย และไอคอน
  static final Map<String, Map<String, dynamic>> _categoryInfo = {
    'system': {'name': global.language('system_setting'), 'icon': Icons.settings, 'color': Colors.blueGrey},
    'user': {'name': global.language('users_and_employees'), 'icon': Icons.people, 'color': Colors.indigo},
    'partner': {'name': global.language('customers_and_sellers'), 'icon': Icons.handshake, 'color': Colors.teal},
    'sale_setting': {'name': global.language('sale_settings'), 'icon': Icons.point_of_sale, 'color': Colors.amber},
    'product': {'name': global.language('product'), 'icon': Icons.inventory_2, 'color': Colors.deepPurple},
    'purchase': {'name': global.language('purchase_transactions'), 'icon': Icons.shopping_cart, 'color': Colors.orange},
    'sale': {'name': global.language('sale_transactions'), 'icon': Icons.storefront, 'color': Colors.green},
    'stock': {'name': global.language('stock'), 'icon': Icons.warehouse, 'color': Colors.brown},
    'finance': {'name': global.language('finance'), 'icon': Icons.account_balance, 'color': Colors.blue},
    'report': {'name': global.language('report'), 'icon': Icons.bar_chart, 'color': Colors.cyan},
    'restaurant': {'name': global.language('restaurant'), 'icon': Icons.restaurant, 'color': Colors.red},
    'check': {'name': global.language('audit'), 'icon': Icons.fact_check, 'color': Colors.pink},
    'master': {'name': global.language('menu_master'), 'icon': Icons.dataset, 'color': Colors.lime},
    'import_export': {'name': global.language('import_export'), 'icon': Icons.import_export, 'color': Colors.deepOrange},
    'ai': {'name': global.language('ai_smart_system'), 'icon': Icons.smart_toy, 'color': Colors.purple},
  };

  Widget _buildScreenPermissionsTable(BranchPermissionModel branchPerm) {
    // จัดกลุ่มหน้าจอตามหมวดหมู่
    final Map<String, List<ScreenDefinition>> screensByCategory = {};
    for (final screenDef in allScreenDefinitions) {
      screensByCategory.putIfAbsent(screenDef.category, () => []).add(screenDef);
    }

    return Column(
      children: [
        // ปุ่มลัดสำหรับกำหนดสิทธิ์
        Padding(
          padding: EdgeInsets.only(bottom: 8),
          child: Wrap(
            spacing: 8,
            runSpacing: 4,
            children: [
              // เลือกหน้าจอทั้งหมด
              _buildQuickButton(
                icon: Icons.select_all,
                label: global.language('select_all'),
                color: Colors.blue,
                onPressed: () => _selectAllScreens(branchPerm),
              ),
              // ยกเลิกทั้งหมด
              _buildQuickButton(
                icon: Icons.deselect,
                label: global.language('cancel_all'),
                color: Colors.grey,
                onPressed: () => _deselectAllScreens(branchPerm),
              ),
              // เปิดสิทธิ์ดูทั้งหมด
              _buildQuickButton(
                icon: Icons.visibility,
                label: global.language('view_all'),
                color: Colors.green,
                onPressed: () => _setAllPermission(branchPerm, 'view', true),
              ),
              // เปิดสิทธิ์เพิ่มทั้งหมด
              _buildQuickButton(
                icon: Icons.add_circle,
                label: global.language('add_all'),
                color: Colors.teal,
                onPressed: () => _setAllPermission(branchPerm, 'add', true),
              ),
              // เปิดสิทธิ์แก้ไขทั้งหมด (แก้ได้ทั้งหมด)
              _buildQuickButton(
                icon: Icons.edit,
                label: global.language('edit_all'),
                color: Colors.orange,
                onPressed: () => _setAllPermission(branchPerm, 'editAll', true),
              ),
              // เปิดสิทธิ์ลบทั้งหมด (ลบได้ทั้งหมด)
              _buildQuickButton(
                icon: Icons.delete,
                label: global.language('clear_all'),
                color: Colors.red,
                onPressed: () => _setAllPermission(branchPerm, 'deleteAll', true),
              ),
              // เปิดสิทธิ์พิมพ์ทั้งหมด
              _buildQuickButton(
                icon: Icons.print,
                label: global.language('print_all'),
                color: Colors.purple,
                onPressed: () => _setAllPermission(branchPerm, 'print', true),
              ),
              // เปิดทุกสิทธิ์ (Full Access)
              _buildQuickButton(
                icon: Icons.check_circle,
                label: global.language('all_permissions'),
                color: Colors.indigo,
                onPressed: () => _setFullAccess(branchPerm),
              ),
            ],
          ),
        ),

        // Header (คงที่)
        Container(
          padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 4),
          decoration: BoxDecoration(
            color: Colors.blue[50],
            borderRadius: BorderRadius.circular(4),
          ),
          child: Row(
            children: [
              Expanded(flex: 3, child: Text(global.language('screen'), style: TextStyle(fontWeight: FontWeight.bold))),
              Expanded(child: Center(child: Text(global.language('view'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11)))),
              Expanded(child: Center(child: Text(global.language('add'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11)))),
              Expanded(child: Center(child: Text(global.language('edit_own'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10)))),
              Expanded(child: Center(child: Text(global.language('edit_all'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10)))),
              Expanded(child: Center(child: Text(global.language('delete_own'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10)))),
              Expanded(child: Center(child: Text(global.language('clear_all'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10)))),
              Expanded(child: Center(child: Text(global.language('print'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11)))),
            ],
          ),
        ),

        // แสดงหน้าจอแยกตามหมวดหมู่
        ...screensByCategory.entries.map((entry) {
          final categoryCode = entry.key;
          final screens = entry.value;
          final categoryInfo = _categoryInfo[categoryCode] ?? {
            'name': categoryCode,
            'icon': Icons.folder,
            'color': Colors.grey,
          };

          return _buildCategorySection(
            branchPerm: branchPerm,
            categoryCode: categoryCode,
            categoryName: categoryInfo['name'] as String,
            categoryIcon: categoryInfo['icon'] as IconData,
            categoryColor: categoryInfo['color'] as Color,
            screens: screens,
          );
        }),
      ],
    );
  }

  /// สร้าง section สำหรับแต่ละหมวดหมู่
  Widget _buildCategorySection({
    required BranchPermissionModel branchPerm,
    required String categoryCode,
    required String categoryName,
    required IconData categoryIcon,
    required Color categoryColor,
    required List<ScreenDefinition> screens,
  }) {
    // ตรวจสอบว่ามีหน้าจอในหมวดหมู่นี้ที่ถูกเลือกหรือไม่
    final selectedCount = screens.where((s) =>
      branchPerm.screens.any((sp) => sp.screenCode == s.code)
    ).length;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // หัวข้อหมวดหมู่
        Container(
          margin: const EdgeInsets.only(top: 12, bottom: 4),
          padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 12),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [
                categoryColor.withValues(alpha: 0.2),
                categoryColor.withValues(alpha: 0.05),
              ],
            ),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: categoryColor.withValues(alpha: 0.3)),
          ),
          child: Row(
            children: [
              Icon(categoryIcon, color: categoryColor, size: 20),
              const SizedBox(width: 8),
              Text(
                categoryName,
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: categoryColor.withValues(alpha: 0.9),
                  fontSize: 14,
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: categoryColor.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  '$selectedCount/${screens.length}',
                  style: TextStyle(
                    fontSize: 11,
                    color: categoryColor,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
              const Spacer(),
              // ปุ่มเลือกทั้งหมดในหมวดหมู่
              InkWell(
                onTap: () => _selectCategoryScreens(branchPerm, screens),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: Colors.green.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: Colors.green.withValues(alpha: 0.3)),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.check_box, size: 14, color: Colors.green),
                      SizedBox(width: 4),
                      Text('เลือกหมวด', style: TextStyle(fontSize: 10, color: Colors.green)),
                    ],
                  ),
                ),
              ),
              const SizedBox(width: 4),
              // ปุ่มยกเลิกทั้งหมดในหมวดหมู่
              InkWell(
                onTap: () => _deselectCategoryScreens(branchPerm, screens),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: Colors.red.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: Colors.red.withValues(alpha: 0.3)),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.close, size: 14, color: Colors.red),
                      SizedBox(width: 4),
                      Text('ยกเลิกหมวด', style: TextStyle(fontSize: 10, color: Colors.red)),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
        // รายการหน้าจอในหมวดหมู่
        ...screens.map((screenDef) {
          final screenPerm = branchPerm.screens.firstWhere(
            (s) => s.screenCode == screenDef.code,
            orElse: () => ScreenPermissionModel(
              screenCode: screenDef.code,
              screenName: screenDef.name,
            ),
          );
          final isEnabled = branchPerm.screens.any((s) => s.screenCode == screenDef.code);

          return _buildScreenRow(branchPerm, screenDef, screenPerm, isEnabled);
        }),
      ],
    );
  }

  /// เลือกหน้าจอทั้งหมดในหมวดหมู่
  void _selectCategoryScreens(BranchPermissionModel branchPerm, List<ScreenDefinition> screens) {
    setState(() {
      for (var screenDef in screens) {
        if (!branchPerm.screens.any((s) => s.screenCode == screenDef.code)) {
          branchPerm.screens.add(ScreenPermissionModel(
            screenCode: screenDef.code,
            screenName: screenDef.name,
            canView: true,
          ));
        }
      }
    });
  }

  /// ยกเลิกเลือกหน้าจอทั้งหมดในหมวดหมู่
  void _deselectCategoryScreens(BranchPermissionModel branchPerm, List<ScreenDefinition> screens) {
    setState(() {
      final screenCodes = screens.map((s) => s.code).toSet();
      branchPerm.screens.removeWhere((s) => screenCodes.contains(s.screenCode));
    });
  }

  Widget _buildScreenRow(
    BranchPermissionModel branchPerm,
    ScreenDefinition screenDef,
    ScreenPermissionModel screenPerm,
    bool isEnabled,
  ) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 4),
      decoration: BoxDecoration(
        border: Border(bottom: BorderSide(color: Colors.grey[200]!)),
      ),
      child: Row(
        children: [
          // ชื่อหน้าจอ + checkbox
          Expanded(
            flex: 3,
            child: Row(
              children: [
                Checkbox(
                  value: isEnabled,
                  onChanged: (value) {
                    setState(() {
                      if (value == true) {
                        branchPerm.screens.add(ScreenPermissionModel(
                          screenCode: screenDef.code,
                          screenName: screenDef.name,
                          canView: true,
                        ));
                      } else {
                        branchPerm.screens.removeWhere((s) => s.screenCode == screenDef.code);
                      }
                    });
                  },
                ),
                Expanded(child: Text(screenDef.name, style: const TextStyle(fontSize: 13))),
              ],
            ),
          ),
          // ดู
          Expanded(
            child: Center(
              child: Checkbox(
                value: isEnabled && screenPerm.canView,
                onChanged: isEnabled
                    ? (value) {
                        setState(() {
                          final index = branchPerm.screens.indexWhere((s) => s.screenCode == screenDef.code);
                          if (index >= 0) {
                            branchPerm.screens[index] = branchPerm.screens[index].copyWith(canView: value);
                          }
                        });
                      }
                    : null,
              ),
            ),
          ),
          // เพิ่ม
          Expanded(
            child: Center(
              child: Checkbox(
                value: isEnabled && screenPerm.canAdd,
                onChanged: isEnabled
                    ? (value) {
                        setState(() {
                          final index = branchPerm.screens.indexWhere((s) => s.screenCode == screenDef.code);
                          if (index >= 0) {
                            branchPerm.screens[index] = branchPerm.screens[index].copyWith(canAdd: value);
                          }
                        });
                      }
                    : null,
              ),
            ),
          ),
          // แก้ไขตัวเอง
          Expanded(
            child: Center(
              child: Checkbox(
                value: isEnabled && screenPerm.canEditOwn,
                onChanged: isEnabled
                    ? (value) {
                        setState(() {
                          final index = branchPerm.screens.indexWhere((s) => s.screenCode == screenDef.code);
                          if (index >= 0) {
                            branchPerm.screens[index] = branchPerm.screens[index].copyWith(canEditOwn: value);
                          }
                        });
                      }
                    : null,
              ),
            ),
          ),
          // แก้ไขทั้งหมด
          Expanded(
            child: Center(
              child: Checkbox(
                value: isEnabled && screenPerm.canEditAll,
                onChanged: isEnabled
                    ? (value) {
                        setState(() {
                          final index = branchPerm.screens.indexWhere((s) => s.screenCode == screenDef.code);
                          if (index >= 0) {
                            branchPerm.screens[index] = branchPerm.screens[index].copyWith(canEditAll: value);
                          }
                        });
                      }
                    : null,
              ),
            ),
          ),
          // ลบตัวเอง
          Expanded(
            child: Center(
              child: Checkbox(
                value: isEnabled && screenPerm.canDeleteOwn,
                onChanged: isEnabled
                    ? (value) {
                        setState(() {
                          final index = branchPerm.screens.indexWhere((s) => s.screenCode == screenDef.code);
                          if (index >= 0) {
                            branchPerm.screens[index] = branchPerm.screens[index].copyWith(canDeleteOwn: value);
                          }
                        });
                      }
                    : null,
              ),
            ),
          ),
          // ลบทั้งหมด
          Expanded(
            child: Center(
              child: Checkbox(
                value: isEnabled && screenPerm.canDeleteAll,
                onChanged: isEnabled
                    ? (value) {
                        setState(() {
                          final index = branchPerm.screens.indexWhere((s) => s.screenCode == screenDef.code);
                          if (index >= 0) {
                            branchPerm.screens[index] = branchPerm.screens[index].copyWith(canDeleteAll: value);
                          }
                        });
                      }
                    : null,
              ),
            ),
          ),
          // พิมพ์
          Expanded(
            child: Center(
              child: Checkbox(
                value: isEnabled && screenPerm.canPrint,
                onChanged: isEnabled
                    ? (value) {
                        setState(() {
                          final index = branchPerm.screens.indexWhere((s) => s.screenCode == screenDef.code);
                          if (index >= 0) {
                            branchPerm.screens[index] = branchPerm.screens[index].copyWith(canPrint: value);
                          }
                        });
                      }
                    : null,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
