import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// หน้าจอตั้งค่าการอนุมัติใบเสนอราคา (QT Approval Setting Screen)
/// กำหนดกฎการอนุมัติตามวงเงินและประเภทใบเสนอราคา
class QTApprovalSettingScreen extends StatefulWidget {
  const QTApprovalSettingScreen({super.key});

  @override
  State<QTApprovalSettingScreen> createState() =>
      _QTApprovalSettingScreenState();
}

class _QTApprovalSettingScreenState extends State<QTApprovalSettingScreen> {
  bool _isLoading = true;
  bool _isLoadingUsers = false;
  List<PurchaseTypeModel> _purchaseTypes = [];
  List<POApprovalSettingModel> _settings = [];
  PurchaseTypeModel? _selectedPurchaseType;
  POApprovalSettingModel? _currentSetting;
  List<ApprovalRuleModel> _editingRules = [];
  List<ApproverInfoModel> _allUsers = [];

  // ระดับผู้อนุมัติ (สอดคล้องกับ user_screen.dart)
  final List<Map<String, dynamic>> _approvalLevels = [
    {'level': 0, 'name': global.language('no_approval_required'), 'description': global.language('approval_auto_approved')},
    {'level': 1, 'name': global.language('level_1'), 'description': global.language('dept_head')},
    {'level': 2, 'name': global.language('level_2'), 'description': global.language('manager')},
    {'level': 3, 'name': global.language('level_3'), 'description': global.language('executive')},
    {'level': 4, 'name': global.language('max'), 'description': global.language('managing_director')},
  ];

  @override
  void initState() {
    super.initState();
    _loadData();
    _loadAllUsers();
  }

  /// โหลดรายชื่อผู้ใช้ทั้งหมดสำหรับเลือกผู้อนุมัติ
  Future<void> _loadAllUsers() async {
    setState(() => _isLoadingUsers = true);

    debugPrint('[QTApprovalSetting] Loading users for approver selection...');
    final result = await ApprovalApiServiceUsers.getApproversUserList();

    if (result.isSuccess) {
      debugPrint('[QTApprovalSetting] Loaded ${result.items.length} users');
      setState(() {
        _allUsers = result.items;
        _isLoadingUsers = false;
      });
    } else {
      debugPrint('[QTApprovalSetting] Failed to load users: ${result.errorMessage}');
      setState(() => _isLoadingUsers = false);
    }
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    // โหลดประเภทใบเสนอราคา
    final typesResult = await ApprovalApiService.getPurchaseTypes();
    // โหลดการตั้งค่าอนุมัติ
    final settingsResult = await ApprovalApiService.getPOApprovalSettings();

    if (typesResult.isSuccess && settingsResult.isSuccess) {
      // ถ้าไม่มีประเภทใบเสนอราคา ให้สร้าง "ทั่วไป" อัตโนมัติ
      if (typesResult.items.isEmpty) {
        await _createDefaultPurchaseType();
        return;
      }

      setState(() {
        _purchaseTypes = typesResult.items;
        _settings = settingsResult.items;
        _isLoading = false;
      });
    } else {
      setState(() => _isLoading = false);
      if (mounted) {
        _showError(global.language('load_data_failed'));
      }
    }
  }

  /// สร้างประเภทใบเสนอราคา "ทั่วไป" เป็นค่าเริ่มต้น
  Future<void> _createDefaultPurchaseType() async {
    final defaultType = PurchaseTypeModel(
      code: 'GENERAL',
      names: [LanguageDataModel(code: 'th', name: 'ทั่วไป')],
      descriptions: [LanguageDataModel(code: 'th', name: 'ใบเสนอราคาทั่วไป')],
      isActive: true,
    );

    final result = await ApprovalApiService.savePurchaseType(defaultType);
    if (result.isSuccess) {
      await _loadData();
    } else {
      setState(() => _isLoading = false);
      if (mounted) {
        _showError(global.language('cannot_create_default_quotation_type'));
      }
    }
  }

  void _showError(String message) {
    global.showErrorSnackBar(context, message);
  }

  void _showSuccess(String message) {
    global.showSuccessSnackBar(context, message);
  }

  void _selectPurchaseType(PurchaseTypeModel type) {
    // ค้นหาการตั้งค่าที่มีอยู่
    final existingSetting = _settings.firstWhere(
      (s) => s.purchaseTypeCode == type.code,
      orElse: () => POApprovalSettingModel(
        purchaseTypeCode: type.code,
        purchaseTypeName: type.getName(global.systemLanguage),
        rules: [],
      ),
    );

    setState(() {
      _selectedPurchaseType = type;
      _currentSetting = existingSetting;
      _editingRules = List.from(existingSetting.rules);
    });
  }

  void _addRule() {
    // กำหนดค่า minAmount จากกฎสุดท้าย
    double minAmount = 0;
    if (_editingRules.isNotEmpty) {
      final lastRule = _editingRules.last;
      minAmount = lastRule.maxAmount;
    }

    setState(() {
      _editingRules.add(ApprovalRuleModel(
        minAmount: minAmount,
        maxAmount: 0,
        approvalLevel: 0,
        approvalLevelName: global.language('no_approval_required'),
      ));
    });
  }

  void _removeRule(int index) {
    setState(() {
      _editingRules.removeAt(index);
    });
  }

  void _updateRule(int index, ApprovalRuleModel updatedRule) {
    setState(() {
      _editingRules[index] = updatedRule;
    });
  }

  Future<void> _saveSetting() async {
    if (_selectedPurchaseType == null) return;

    // Validate rules
    for (int i = 0; i < _editingRules.length; i++) {
      final rule = _editingRules[i];
      if (i > 0 && rule.minAmount < _editingRules[i - 1].maxAmount) {
        _showError('${global.language("rule_min_amount_error")} ${i + 1}');
        return;
      }
    }

    setState(() => _isLoading = true);

    final model = POApprovalSettingModel(
      guid: _currentSetting?.guid,
      purchaseTypeCode: _selectedPurchaseType!.code,
      purchaseTypeName: _selectedPurchaseType!.getName(global.systemLanguage),
      rules: _editingRules,
    );

    final result = await ApprovalApiService.savePOApprovalSetting(model);

    if (result.isSuccess) {
      _showSuccess(global.language('save_success'));
      await _loadData();
      // Re-select the purchase type to refresh
      if (_selectedPurchaseType != null) {
        _selectPurchaseType(_selectedPurchaseType!);
      }
    } else {
      setState(() => _isLoading = false);
      _showError('${global.language("save_failed")}: ${result.errorMessage}');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Text(global.language('approve_quotation')),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => global.gotoMainMenu(context),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadData,
            tooltip: global.language('database_master_info.refresh'),
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _buildBody(),
    );
  }

  Widget _buildBody() {
    return Row(
      children: [
        // Left panel - Purchase types
        SizedBox(
          width: 280,
          child: _buildPurchaseTypeList(),
        ),
        const VerticalDivider(width: 1),
        // Right panel - Approval rules
        Expanded(child: _buildRulesPanel()),
      ],
    );
  }

  /// แสดง dialog เพิ่มประเภทใบเสนอราคาใหม่
  Future<void> _showAddPurchaseTypeDialog() async {
    final codeController = TextEditingController();
    final nameController = TextEditingController();
    final descController = TextEditingController();

    final result = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('add_quotation_type')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: codeController,
              decoration: InputDecoration(
                labelText: global.language('code'),
                hintText: global.language('eg_general_special_vip'),
                border: OutlineInputBorder(),
              ),
              textCapitalization: TextCapitalization.characters,
            ),
            SizedBox(height: 12),
            TextField(
              controller: nameController,
              decoration: InputDecoration(
                labelText: global.language('type_name'),
                hintText: global.language('eg_general_special_quotation'),
                border: OutlineInputBorder(),
              ),
            ),
            SizedBox(height: 12),
            TextField(
              controller: descController,
              decoration: InputDecoration(
                labelText: global.language('description_optional'),
                border: OutlineInputBorder(),
              ),
              maxLines: 2,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(global.language('form_design_save')),
          ),
        ],
      ),
    );

    if (result == true) {
      if (codeController.text.isEmpty || nameController.text.isEmpty) {
        _showError(global.language('please_enter_code_and_type_name'));
        return;
      }

      setState(() => _isLoading = true);
      final newType = PurchaseTypeModel(
        code: codeController.text.toUpperCase(),
        names: [LanguageDataModel(code: global.systemLanguage, name: nameController.text)],
        descriptions: descController.text.isNotEmpty
            ? [LanguageDataModel(code: global.systemLanguage, name: descController.text)]
            : [],
        isActive: true,
      );

      final saveResult = await ApprovalApiService.savePurchaseType(newType);
      if (saveResult.isSuccess) {
        _showSuccess('เพิ่มประเภทใบเสนอราคาสำเร็จ');
        await _loadData();
      } else {
        setState(() => _isLoading = false);
        _showError('บันทึกไม่สำเร็จ: ${saveResult.errorMessage}');
      }
    }
  }

  /// ลบประเภทใบเสนอราคา
  Future<void> _deletePurchaseType(PurchaseTypeModel type) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('alert_confirm_delete')),
        content: Text('ต้องการลบประเภท "${type.getName(global.systemLanguage)}" หรือไม่?\n\nการลบจะลบกฎการอนุมัติที่เกี่ยวข้องด้วย'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () => Navigator.pop(context, true),
            child: Text(global.language('delete')),
          ),
        ],
      ),
    );

    if (confirm == true) {
      setState(() => _isLoading = true);

      // ลบการตั้งค่าอนุมัติก่อน (ถ้ามี)
      await ApprovalApiService.deletePOApprovalSetting(type.code);
      // ลบประเภทใบเสนอราคา (ใช้ guid สำหรับ mainapi)
      final success = await ApprovalApiService.deletePurchaseType(type.guid ?? '');

      if (success) {
        _showSuccess(global.language('delete_success'));
        setState(() {
          if (_selectedPurchaseType?.code == type.code) {
            _selectedPurchaseType = null;
            _currentSetting = null;
            _editingRules = [];
          }
        });
        await _loadData();
      } else {
        setState(() => _isLoading = false);
        _showError(global.language('delete_failed'));
      }
    }
  }

  Widget _buildPurchaseTypeList() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.teal.shade50,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(0)),
          ),
          child: Row(
            children: [
              Icon(Icons.request_quote, color: Colors.teal.shade700),
              SizedBox(width: 8),
              Expanded(
                child: Text(
                  global.language('quotation_type'),
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.teal.shade700,
                  ),
                ),
              ),
              // ปุ่มเพิ่มประเภทใบเสนอราคา
              IconButton(
                icon: Icon(Icons.add_circle, color: Colors.teal.shade700),
                onPressed: _showAddPurchaseTypeDialog,
                tooltip: global.language('add_quotation_type'),
              ),
            ],
          ),
        ),
        Expanded(
          child: _purchaseTypes.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.info_outline, size: 48, color: Colors.grey.shade400),
                      const SizedBox(height: 16),
                      const Text('ยังไม่มีประเภทใบเสนอราคา'),
                      const SizedBox(height: 8),
                      TextButton(
                        onPressed: () {
                          Navigator.pushNamed(context, '/quotation_type_screen');
                        },
                        child: const Text('ไปเพิ่มประเภทใบเสนอราคา'),
                      ),
                    ],
                  ),
                )
              : ListView.builder(
                  itemCount: _purchaseTypes.length,
                  itemBuilder: (context, index) {
                    final type = _purchaseTypes[index];
                    final isSelected =
                        _selectedPurchaseType?.code == type.code;
                    // ตรวจสอบว่ามีการตั้งค่าแล้วหรือไม่
                    final hasSetting =
                        _settings.any((s) => s.purchaseTypeCode == type.code);

                    return ListTile(
                      selected: isSelected,
                      selectedTileColor: Colors.teal.shade100,
                      leading: Stack(
                        children: [
                          CircleAvatar(
                            backgroundColor: isSelected
                                ? Colors.teal.shade700
                                : Colors.grey.shade200,
                            child: Icon(
                              Icons.request_quote,
                              color: isSelected ? Colors.white : Colors.grey,
                              size: 20,
                            ),
                          ),
                          if (hasSetting)
                            Positioned(
                              right: 0,
                              bottom: 0,
                              child: Container(
                                width: 12,
                                height: 12,
                                decoration: BoxDecoration(
                                  color: Colors.green,
                                  shape: BoxShape.circle,
                                  border: Border.all(color: Colors.white, width: 1),
                                ),
                                child: const Icon(
                                  Icons.check,
                                  size: 8,
                                  color: Colors.white,
                                ),
                              ),
                            ),
                        ],
                      ),
                      title: Text(
                        type.getName(global.systemLanguage),
                        style: TextStyle(
                          fontWeight:
                              isSelected ? FontWeight.bold : FontWeight.normal,
                        ),
                      ),
                      subtitle: Text(
                        type.code,
                        style: TextStyle(
                          color: Colors.grey.shade600,
                          fontSize: 12,
                        ),
                      ),
                      trailing: IconButton(
                        icon: Icon(Icons.delete_outline,
                          color: Colors.red.shade300, size: 20),
                        onPressed: () => _deletePurchaseType(type),
                        tooltip: global.language('delete_this_type'),
                      ),
                      onTap: () => _selectPurchaseType(type),
                    );
                  },
                ),
        ),
      ],
    );
  }

  Widget _buildRulesPanel() {
    if (_selectedPurchaseType == null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.arrow_back, size: 48, color: Colors.grey.shade400),
            const SizedBox(height: 16),
            Text(
              'เลือกประเภทใบเสนอราคาจากด้านซ้าย',
              style: TextStyle(color: Colors.grey.shade600),
            ),
          ],
        ),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // Header
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.green.shade50,
          ),
          child: Row(
            children: [
              Icon(
                Icons.fact_check,
                color: Colors.green.shade700,
              ),
              SizedBox(width: 8),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'กฎการอนุมัติ: ${_selectedPurchaseType!.getName(global.systemLanguage)}',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: Colors.green.shade700,
                      ),
                    ),
                    SizedBox(height: 4),
                    Text(
                      global.language('set_approval_conditions_by_amount'),
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey.shade600,
                      ),
                    ),
                  ],
                ),
              ),
              // ปุ่มเพิ่มกฎ
              OutlinedButton.icon(
                onPressed: _addRule,
                icon: Icon(Icons.add, size: 18),
                label: Text(global.language('add_rule')),
              ),
              const SizedBox(width: 8),
              // ปุ่มบันทึก
              ElevatedButton.icon(
                onPressed: _saveSetting,
                icon: Icon(Icons.save, size: 18),
                label: Text(global.language('form_design_save')),
              ),
            ],
          ),
        ),

        // คำอธิบายระดับ
        Container(
          margin: const EdgeInsets.all(16),
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.grey.shade100,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language('approval_level_description'),
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: Colors.grey.shade700,
                ),
              ),
              SizedBox(height: 4),
              Text(
                global.language('approval_level_1_desc'),
                style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
              ),
              Text(
                global.language('approval_level_2_desc'),
                style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
              ),
              Text(
                global.language('approval_level_3_desc'),
                style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
              ),
              Text(
                global.language('approval_level_max_desc'),
                style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
              ),
            ],
          ),
        ),

        // Rules list
        Expanded(
          child: _editingRules.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.rule, size: 48, color: Colors.grey.shade400),
                      SizedBox(height: 16),
                      Text(global.language('no_approval_rules_yet')),
                      SizedBox(height: 8),
                      ElevatedButton.icon(
                        onPressed: _addRule,
                        icon: Icon(Icons.add),
                        label: Text(global.language('add_rule')),
                      ),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  itemCount: _editingRules.length,
                  itemBuilder: (context, index) {
                    return _buildRuleCard(index, _editingRules[index]);
                  },
                ),
        ),
      ],
    );
  }

  Widget _buildRuleCard(int index, ApprovalRuleModel rule) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Rule header
            Row(
              children: [
                Container(
                  width: 28,
                  height: 28,
                  decoration: BoxDecoration(
                    color: Colors.teal.shade100,
                    shape: BoxShape.circle,
                  ),
                  child: Center(
                    child: Text(
                      '${index + 1}',
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: Colors.teal.shade700,
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  'กฎที่ ${index + 1}',
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
                Spacer(),
                IconButton(
                  icon: const Icon(Icons.delete_outline, color: Colors.red),
                  onPressed: () => _removeRule(index),
                  tooltip: global.language('delete_this_rule'),
                ),
              ],
            ),
            const Divider(),

            // Amount range
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    initialValue: rule.minAmount.toStringAsFixed(0),
                    keyboardType: TextInputType.number,
                    inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                    decoration: InputDecoration(
                      labelText: global.language('min_amount_baht'),
                      border: OutlineInputBorder(),
                      isDense: true,
                    ),
                    onChanged: (value) {
                      _updateRule(
                        index,
                        rule.copyWith(minAmount: double.tryParse(value) ?? 0),
                      );
                    },
                  ),
                ),
                const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 12),
                  child: Text('-', style: TextStyle(fontSize: 20)),
                ),
                Expanded(
                  child: TextFormField(
                    initialValue: rule.maxAmount == 0
                        ? ''
                        : rule.maxAmount.toStringAsFixed(0),
                    keyboardType: TextInputType.number,
                    inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                    decoration: InputDecoration(
                      labelText: global.language('max_amount_unlimited'),
                      border: OutlineInputBorder(),
                      isDense: true,
                    ),
                    onChanged: (value) {
                      _updateRule(
                        index,
                        rule.copyWith(maxAmount: double.tryParse(value) ?? 0),
                      );
                    },
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),

            // Approval level dropdown
            DropdownButtonFormField<int>(
              // ignore: deprecated_member_use
              value: rule.approvalLevel,
              decoration: InputDecoration(
                labelText: global.language('approval_level'),
                border: OutlineInputBorder(),
                isDense: true,
              ),
              items: _approvalLevels.map((level) {
                return DropdownMenuItem<int>(
                  value: level['level'],
                  child: Text('${level['name']} - ${level['description']}'),
                );
              }).toList(),
              onChanged: (value) {
                final levelInfo = _approvalLevels.firstWhere(
                  (l) => l['level'] == value,
                );
                _updateRule(
                  index,
                  rule.copyWith(
                    approvalLevel: value ?? 0,
                    approvalLevelName: levelInfo['name'],
                  ),
                );
              },
            ),

            // =====================================================
            // Approvers Selection Section - เลือกผู้มีสิทธิ์อนุมัติ
            // =====================================================
            if (rule.approvalLevel > 0) ...[
              SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.purple.shade50,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.purple.shade200),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.people, color: Colors.purple.shade700, size: 20),
                        SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            global.language('approvers_anyone_in_list'),
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 13,
                              color: Colors.purple.shade700,
                            ),
                          ),
                        ),
                        OutlinedButton.icon(
                          onPressed: () => _showApproverSelectionDialog(index, rule),
                          icon: Icon(Icons.add, size: 16),
                          label: Text(global.language('select')),
                          style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.purple.shade700,
                            side: BorderSide(color: Colors.purple.shade300),
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    // แสดงรายชื่อผู้อนุมัติที่เลือกไว้
                    if (rule.approvers.isEmpty)
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.grey.shade100,
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Row(
                          children: [
                            Icon(Icons.warning_amber, size: 16, color: Colors.orange.shade700),
                            SizedBox(width: 8),
                            Text(
                              global.language('no_approver_selected_use_permission'),
                              style: TextStyle(
                                fontSize: 12,
                                color: Colors.grey.shade600,
                                fontStyle: FontStyle.italic,
                              ),
                            ),
                          ],
                        ),
                      )
                    else
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: rule.approvers.map((approver) {
                          return _buildApproverChip(index, rule, approver);
                        }).toList(),
                      ),
                  ],
                ),
              ),
            ],

            // Summary text
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Colors.teal.shade50,
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                children: [
                  Icon(Icons.info_outline,
                      size: 16, color: Colors.teal.shade700),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _getRuleSummary(rule),
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.teal.shade700,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// สร้าง Chip แสดงผู้อนุมัติที่เลือก
  Widget _buildApproverChip(int ruleIndex, ApprovalRuleModel rule, ApproverInfoModel approver) {
    return Chip(
      avatar: CircleAvatar(
        backgroundColor: Colors.purple.shade100,
        child: Text(
          approver.userName.isNotEmpty ? approver.userName[0].toUpperCase() : '?',
          style: TextStyle(fontSize: 12, color: Colors.purple.shade700),
        ),
      ),
      label: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            approver.userName.isNotEmpty ? approver.userName : approver.userCode,
            style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w500),
          ),
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (approver.hasEmail)
                Padding(
                  padding: const EdgeInsets.only(right: 4),
                  child: Icon(Icons.email, size: 12, color: Colors.blue.shade600),
                ),
              if (approver.hasLine)
                Icon(Icons.chat, size: 12, color: Colors.green.shade600),
              if (!approver.hasEmail && !approver.hasLine)
                Text(
                  global.language('no_email_line'),
                  style: TextStyle(fontSize: 10, color: Colors.grey.shade500),
                ),
            ],
          ),
        ],
      ),
      deleteIcon: const Icon(Icons.close, size: 16),
      onDeleted: () {
        // ลบผู้อนุมัติออกจากกฎ
        final newApprovers = rule.approvers.where((a) => a.userCode != approver.userCode).toList();
        _updateRule(ruleIndex, rule.copyWith(approvers: newApprovers));
      },
      backgroundColor: Colors.purple.shade50,
      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
    );
  }

  /// แสดง Dialog เลือกผู้อนุมัติ
  Future<void> _showApproverSelectionDialog(int ruleIndex, ApprovalRuleModel rule) async {
    // รอให้โหลด users เสร็จก่อนเปิด dialog
    if (_allUsers.isEmpty && !_isLoadingUsers) {
      await _loadAllUsers();
    }

    // รอถ้ายังโหลดอยู่ (max 30 วินาที)
    int waitCount = 0;
    while (_isLoadingUsers && waitCount < 60) {
      await Future.delayed(const Duration(milliseconds: 500));
      waitCount++;
    }

    // ตรวจสอบว่า widget ยัง mounted อยู่
    if (!mounted) return;

    // สร้างรายการที่เลือกไว้
    final selectedUserCodes = Set<String>.from(rule.approvers.map((a) => a.userCode));

    final result = await showDialog<List<ApproverInfoModel>>(
      context: context,
      builder: (dialogContext) => _QTApproverSelectionDialog(
        allUsers: _allUsers,
        selectedUserCodes: selectedUserCodes,
        isLoading: false,
      ),
    );

    if (result != null && mounted) {
      _updateRule(ruleIndex, rule.copyWith(approvers: result));
    }
  }

  String _getRuleSummary(ApprovalRuleModel rule) {
    final minText = global.formatNumber(rule.minAmount);
    final maxText = rule.maxAmount == 0
        ? global.language('coupon_hint_unlimited')
        : global.formatNumber(rule.maxAmount);

    if (rule.approvalLevel == 0) {
      return 'วงเงิน $minText - $maxText บาท: ไม่ต้องอนุมัติ (อนุมัติอัตโนมัติ)';
    } else {
      String approverText = rule.approvalLevelName;
      if (rule.approvers.isNotEmpty) {
        approverText += ' (${rule.approvers.length} คน)';
      }
      return 'วงเงิน $minText - $maxText บาท: ต้องอนุมัติโดย $approverText';
    }
  }
}

// =====================================================
// Dialog สำหรับเลือกผู้อนุมัติ
// =====================================================

class _QTApproverSelectionDialog extends StatefulWidget {
  final List<ApproverInfoModel> allUsers;
  final Set<String> selectedUserCodes;
  final bool isLoading;

  const _QTApproverSelectionDialog({
    required this.allUsers,
    required this.selectedUserCodes,
    required this.isLoading,
  });

  @override
  State<_QTApproverSelectionDialog> createState() => _QTApproverSelectionDialogState();
}

class _QTApproverSelectionDialogState extends State<_QTApproverSelectionDialog> {
  late Set<String> _selectedUserCodes;
  String _searchText = '';

  @override
  void initState() {
    super.initState();
    _selectedUserCodes = Set<String>.from(widget.selectedUserCodes);
  }

  List<ApproverInfoModel> get _filteredUsers {
    if (_searchText.isEmpty) {
      return widget.allUsers;
    }
    final search = _searchText.toLowerCase();
    return widget.allUsers.where((user) {
      return user.userCode.toLowerCase().contains(search) ||
          user.userName.toLowerCase().contains(search) ||
          (user.email?.toLowerCase().contains(search) ?? false) ||
          (user.lineDisplayName?.toLowerCase().contains(search) ?? false) ||
          (user.position?.toLowerCase().contains(search) ?? false) ||
          (user.department?.toLowerCase().contains(search) ?? false);
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          Icon(Icons.people, color: Colors.purple.shade700),
          SizedBox(width: 8),
          Text(global.language('select_approvers')),
        ],
      ),
      content: SizedBox(
        width: 500,
        height: 400,
        child: Column(
          children: [
            // Search box
            TextField(
              decoration: InputDecoration(
                hintText: global.language('search_name_email_line'),
                prefixIcon: const Icon(Icons.search),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
                isDense: true,
              ),
              onChanged: (value) {
                setState(() => _searchText = value);
              },
            ),
            const SizedBox(height: 12),
            // จำนวนที่เลือก
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: Colors.purple.shade50,
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                children: [
                  Icon(Icons.check_circle, size: 16, color: Colors.purple.shade700),
                  const SizedBox(width: 8),
                  Text(
                    'เลือกแล้ว ${_selectedUserCodes.length} คน',
                    style: TextStyle(
                      fontWeight: FontWeight.w500,
                      color: Colors.purple.shade700,
                    ),
                  ),
                  const Spacer(),
                  if (_selectedUserCodes.isNotEmpty)
                    TextButton(
                      onPressed: () {
                        setState(() => _selectedUserCodes.clear());
                      },
                      child: Text(global.language('clear_all')),
                    ),
                ],
              ),
            ),
            const SizedBox(height: 8),
            // User list
            Expanded(
              child: widget.isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _filteredUsers.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.person_off, size: 48, color: Colors.grey.shade400),
                              SizedBox(height: 8),
                              Text(
                                _searchText.isEmpty ? global.language('user_not_found') : global.language('user_search_not_found'),
                                style: TextStyle(color: Colors.grey.shade600),
                              ),
                            ],
                          ),
                        )
                      : ListView.builder(
                          itemCount: _filteredUsers.length,
                          itemBuilder: (context, index) {
                            final user = _filteredUsers[index];
                            final isSelected = _selectedUserCodes.contains(user.userCode);
                            return _buildUserTile(user, isSelected);
                          },
                        ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(global.language('cancel')),
        ),
        ElevatedButton.icon(
          onPressed: () {
            // สร้างรายการผู้อนุมัติที่เลือก
            final selectedApprovers = widget.allUsers
                .where((u) => _selectedUserCodes.contains(u.userCode))
                .toList();
            Navigator.pop(context, selectedApprovers);
          },
          icon: const Icon(Icons.check, size: 18),
          label: Text('ยืนยัน (${_selectedUserCodes.length})'),
        ),
      ],
    );
  }

  Widget _buildUserTile(ApproverInfoModel user, bool isSelected) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      color: isSelected ? Colors.purple.shade50 : null,
      child: ListTile(
        dense: true,
        leading: CircleAvatar(
          backgroundColor: isSelected ? Colors.purple.shade200 : Colors.grey.shade200,
          child: Text(
            user.userName.isNotEmpty ? user.userName[0].toUpperCase() : '?',
            style: TextStyle(
              color: isSelected ? Colors.purple.shade700 : Colors.grey.shade600,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
        title: Row(
          children: [
            Expanded(
              child: Text(
                user.userName.isNotEmpty ? user.userName : user.userCode,
                style: TextStyle(
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                ),
              ),
            ),
            // แสดง badge email/LINE
            if (user.hasEmail)
              Container(
                margin: const EdgeInsets.only(left: 4),
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.blue.shade100,
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.email, size: 12, color: Colors.blue.shade700),
                    const SizedBox(width: 2),
                    Text(
                      'Email',
                      style: TextStyle(fontSize: 10, color: Colors.blue.shade700),
                    ),
                  ],
                ),
              ),
            if (user.hasLine)
              Container(
                margin: const EdgeInsets.only(left: 4),
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.green.shade100,
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.chat, size: 12, color: Colors.green.shade700),
                    const SizedBox(width: 2),
                    Text(
                      'LINE',
                      style: TextStyle(fontSize: 10, color: Colors.green.shade700),
                    ),
                  ],
                ),
              ),
          ],
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              user.userCode,
              style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
            ),
            if (user.position != null || user.department != null)
              Text(
                [user.position, user.department].where((s) => s != null && s.isNotEmpty).join(' - '),
                style: TextStyle(fontSize: 11, color: Colors.grey.shade500),
              ),
            if (user.email != null && user.email!.isNotEmpty)
              Row(
                children: [
                  Icon(Icons.email_outlined, size: 11, color: Colors.grey.shade500),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(
                      user.email!,
                      style: TextStyle(fontSize: 11, color: Colors.grey.shade500),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            if (user.lineDisplayName != null && user.lineDisplayName!.isNotEmpty)
              Row(
                children: [
                  Icon(Icons.chat_outlined, size: 11, color: Colors.green.shade400),
                  const SizedBox(width: 4),
                  Text(
                    user.lineDisplayName!,
                    style: TextStyle(fontSize: 11, color: Colors.green.shade600),
                  ),
                ],
              ),
          ],
        ),
        trailing: Checkbox(
          value: isSelected,
          activeColor: Colors.purple.shade700,
          onChanged: (value) {
            setState(() {
              if (value == true) {
                _selectedUserCodes.add(user.userCode);
              } else {
                _selectedUserCodes.remove(user.userCode);
              }
            });
          },
        ),
        onTap: () {
          setState(() {
            if (isSelected) {
              _selectedUserCodes.remove(user.userCode);
            } else {
              _selectedUserCodes.add(user.userCode);
            }
          });
        },
      ),
    );
  }
}
