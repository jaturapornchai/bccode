import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/lineoa_model.dart';
import 'package:smlaicloud/services/lineoa_api_service.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:smlaicloud/widgets/line_link_qr_dialog.dart';

/// Line OA Configuration Screen
/// หน้าจอตั้งค่าเชื่อมต่อ Line OA
class LineOAConfigScreen extends StatefulWidget {
  const LineOAConfigScreen({super.key});

  @override
  State<LineOAConfigScreen> createState() => _LineOAConfigScreenState();
}

class _LineOAConfigScreenState extends State<LineOAConfigScreen>
    with global.ThemeRefreshMixin {
  bool _isLoading = true;
  Map<LineOAType, LineOAConfigModel> _configs = {};
  LineOAType? _selectedType;

  @override
  void initState() {
    super.initState();
    _loadConfigs();
  }

  Future<void> _loadConfigs() async {
    setState(() => _isLoading = true);

    final result = await LineOAApiService.getConfigs();

    if (result.isSuccess) {
      final configMap = <LineOAType, LineOAConfigModel>{};

      // Initialize all types with empty configs
      for (var type in LineOAType.values) {
        configMap[type] =
            LineOAConfigModel.empty(global.getShopId(), type);
      }

      // Update with existing configs
      for (var config in result.configs) {
        final type = config.type;
        if (type != null) {
          configMap[type] = config;
        }
      }

      setState(() {
        _configs = configMap;
        _isLoading = false;
      });
    } else {
      setState(() => _isLoading = false);
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("load_data_failed")}: ${result.errorMessage}');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(global.language('lineoa_config_title')),
        actions: [
          EditFontSizeControl(onChanged: () => setState(() {})),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadConfigs,
            tooltip: global.language('database_master_info.refresh'),
          ),
        ],
      ),
      body: Builder(
        builder: (context) {
          final scaleFactor = global.editFontScaleFactor;
          return MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: TextScaler.linear(scaleFactor),
            ),
            child: Theme(
              data: Theme.of(context).copyWith(
                inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 12 * scaleFactor,
                    vertical: 10 * scaleFactor,
                  ),
                ),
                iconTheme: IconThemeData(size: 24 * scaleFactor),
              ),
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _buildBody(),
            ),
          );
        },
      ),
    );
  }

  Widget _buildBody() {
    return Row(
      children: [
        // Left panel - Line OA types list
        SizedBox(
          width: 280,
          child: Card(
            margin: const EdgeInsets.all(8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.green.shade50,
                    borderRadius: const BorderRadius.vertical(
                      top: Radius.circular(12),
                    ),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.link, color: Colors.green.shade700),
                      const SizedBox(width: 8),
                      Text(
                        global.language('connect_lineoa'),
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: Colors.green.shade700,
                        ),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: ListView.builder(
                    itemCount: LineOAType.values.length,
                    itemBuilder: (context, index) {
                      final type = LineOAType.values[index];
                      final config = _configs[type];
                      final isConfigured = config?.isConfigured ?? false;
                      final isActive = config?.isActive ?? false;

                      return ListTile(
                        selected: _selectedType == type,
                        selectedTileColor: Colors.green.shade100,
                        leading: Stack(
                          children: [
                            CircleAvatar(
                              backgroundColor: isConfigured
                                  ? (isActive
                                      ? Colors.green.shade100
                                      : global.theme.rowEditColor)
                                  : global.theme.dividerBorderColor,
                              child: Icon(
                                _getIconForType(type),
                                color: isConfigured
                                    ? (isActive
                                        ? Colors.green.shade700
                                        : Colors.orange.shade700)
                                    : global.theme.textSecondaryColor,
                              ),
                            ),
                            if (isConfigured && isActive)
                              Positioned(
                                right: 0,
                                bottom: 0,
                                child: Container(
                                  width: 12,
                                  height: 12,
                                  decoration: BoxDecoration(
                                    color: Colors.green,
                                    shape: BoxShape.circle,
                                    border: Border.all(
                                      color: global.theme.cardColor,
                                      width: 2,
                                    ),
                                  ),
                                ),
                              ),
                          ],
                        ),
                        title: Text(
                          type.displayName,
                          style: const TextStyle(fontWeight: FontWeight.w500),
                        ),
                        subtitle: Text(
                          isConfigured
                              ? (isActive ? global.language('connected') : global.language('not_activated'))
                              : global.language('not_configured'),
                          style: TextStyle(
                            fontSize: 12,
                            color: isConfigured
                                ? (isActive ? Colors.green : Colors.orange)
                                : global.theme.textSecondaryColor,
                          ),
                        ),
                        trailing: const Icon(Icons.chevron_right),
                        onTap: () {
                          setState(() => _selectedType = type);
                        },
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        ),

        // Right panel - Configuration form
        Expanded(
          child: _selectedType == null
              ? _buildEmptyState()
              : _buildConfigForm(_selectedType!),
        ),
      ],
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.touch_app,
            size: 64,
            color: global.theme.iconSecondaryColor,
          ),
          const SizedBox(height: 16),
          Text(
            global.language('select_lineoa_type'),
            style: TextStyle(
              fontSize: 16,
              color: global.theme.textSecondaryColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildConfigForm(LineOAType type) {
    final config = _configs[type] ??
        LineOAConfigModel.empty(global.getShopId(), type);

    return LineOAConfigFormWidget(
      key: ValueKey(type.code),
      config: config,
      type: type,
      onSaved: (updatedConfig) {
        setState(() {
          _configs[type] = updatedConfig;
        });
      },
    );
  }

  IconData _getIconForType(LineOAType type) {
    switch (type) {
      case LineOAType.sales:
        return Icons.point_of_sale;
      case LineOAType.purchasing:
        return Icons.shopping_cart;
      case LineOAType.stock:
        return Icons.inventory;
      case LineOAType.businessOwner:
        return Icons.business;
      case LineOAType.manager:
        return Icons.manage_accounts;
    }
  }
}

/// Line OA Configuration Form Widget
class LineOAConfigFormWidget extends StatefulWidget {
  final LineOAConfigModel config;
  final LineOAType type;
  final Function(LineOAConfigModel) onSaved;

  const LineOAConfigFormWidget({
    super.key,
    required this.config,
    required this.type,
    required this.onSaved,
  });

  @override
  State<LineOAConfigFormWidget> createState() => _LineOAConfigFormWidgetState();
}

class _LineOAConfigFormWidgetState extends State<LineOAConfigFormWidget> with global.ThemeRefreshMixin {
  final _formKey = GlobalKey<FormState>();
  late TextEditingController _channelIdController;
  late TextEditingController _channelSecretController;
  late TextEditingController _accessTokenController;
  late TextEditingController _liffIdController;
  late bool _isActive;

  bool _isSaving = false;
  bool _isTesting = false;
  LineOATestResult? _testResult;

  @override
  void initState() {
    super.initState();
    _channelIdController = TextEditingController(text: widget.config.channelId);
    _channelSecretController =
        TextEditingController(text: widget.config.channelSecret);
    _accessTokenController =
        TextEditingController(text: widget.config.accessToken);
    _liffIdController = TextEditingController(text: widget.config.liffId);
    _isActive = widget.config.isActive;
  }

  @override
  void dispose() {
    _channelIdController.dispose();
    _channelSecretController.dispose();
    _accessTokenController.dispose();
    _liffIdController.dispose();
    super.dispose();
  }

  Future<void> _testConnection() async {
    if (_channelIdController.text.isEmpty ||
        _channelSecretController.text.isEmpty ||
        _accessTokenController.text.isEmpty) {
      global.showWarningSnackBar(context, global.language('please_fill_all_fields'));
      return;
    }

    setState(() {
      _isTesting = true;
      _testResult = null;
    });

    final result = await LineOAApiService.testConnection(
      channelId: _channelIdController.text,
      channelSecret: _channelSecretController.text,
      accessToken: _accessTokenController.text,
    );

    setState(() {
      _isTesting = false;
      _testResult = result;
    });

    if (mounted) {
      if (result.success) {
        global.showSuccessSnackBar(context, result.message);
      } else {
        global.showErrorSnackBar(context, result.message);
      }
    }
  }

  Future<void> _saveConfig() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isSaving = true);

    final updatedConfig = widget.config.copyWith(
      channelId: _channelIdController.text,
      channelSecret: _channelSecretController.text,
      accessToken: _accessTokenController.text,
      liffId: _liffIdController.text,
      isActive: _isActive,
    );

    final result = await LineOAApiService.saveConfig(updatedConfig);

    setState(() => _isSaving = false);

    if (result.isSuccess) {
      widget.onSaved(updatedConfig.copyWith(guid: result.guid));
      if (mounted) {
        global.showSuccessSnackBar(context, global.language('save_success'));
      }
    } else {
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("save_failed")}: ${result.errorMessage}');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.all(8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.green.shade50,
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(12),
              ),
            ),
            child: Row(
              children: [
                Icon(Icons.settings, color: Colors.green.shade700),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${global.language("settings")} ${widget.type.displayName}',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Colors.green.shade700,
                        ),
                      ),
                      Text(
                        widget.type.description,
                        style: TextStyle(
                          fontSize: 13,
                          color: Colors.green.shade600,
                        ),
                      ),
                    ],
                  ),
                ),
                // Active switch
                Row(
                  children: [
                    Text(global.language('activate')),
                    Switch(
                      value: _isActive,
                      activeTrackColor: Colors.green.shade200,
                      activeThumbColor: Colors.green,
                      onChanged: (value) {
                        setState(() => _isActive = value);
                      },
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Form
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Channel ID
                    _buildTextField(
                      controller: _channelIdController,
                      label: global.language('channel_id'),
                      hint: 'ใส่ Channel ID จาก LINE Developers',
                      required: true,
                    ),
                    const SizedBox(height: 16),

                    // Channel Secret
                    _buildTextField(
                      controller: _channelSecretController,
                      label: global.language('channel_secret'),
                      hint: 'ใส่ Channel Secret จาก LINE Developers',
                      required: true,
                      obscure: true,
                    ),
                    const SizedBox(height: 16),

                    // Access Token
                    _buildTextField(
                      controller: _accessTokenController,
                      label: global.language('channel_access_token'),
                      hint: 'ใส่ Long-lived Access Token',
                      required: true,
                      maxLines: 3,
                    ),
                    const SizedBox(height: 16),

                    // LIFF ID
                    _buildTextField(
                      controller: _liffIdController,
                      label: 'LIFF ID (ถ้ามี)',
                      hint: 'ใส่ LIFF ID สำหรับ LIFF App',
                      required: false,
                    ),
                    const SizedBox(height: 24),

                    // Test result
                    if (_testResult != null) _buildTestResult(),

                    const SizedBox(height: 16),

                    // Action buttons
                    Row(
                      children: [
                        // Test connection button
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: _isTesting ? null : _testConnection,
                            icon: _isTesting
                                ? const SizedBox(
                                    width: 20,
                                    height: 20,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                    ),
                                  )
                                : Icon(Icons.wifi_tethering),
                            label: Text(
                                _isTesting ? global.language('testing') : global.language('test_connection')),
                            style: OutlinedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 16),
                              foregroundColor: global.theme.primaryColor,
                            ),
                          ),
                        ),
                        const SizedBox(width: 16),

                        // Save button
                        Expanded(
                          child: ElevatedButton.icon(
                            onPressed: _isSaving ? null : _saveConfig,
                            icon: _isSaving
                                ? SizedBox(
                                    width: 20,
                                    height: 20,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                      color: global.theme.onPrimaryColor,
                                    ),
                                  )
                                : Icon(Icons.save),
                            label: Text(_isSaving ? global.language('saving') : global.language('form_design_save')),
                            style: ElevatedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 16),
                              backgroundColor: Colors.green,
                              foregroundColor: global.theme.onPrimaryColor,
                            ),
                          ),
                        ),
                      ],
                    ),

                    const SizedBox(height: 32),

                    // Employee section
                    if (widget.config.guid.isNotEmpty) ...[
                      const Divider(),
                      const SizedBox(height: 16),
                      _buildEmployeeSection(),
                    ],
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hint,
    bool required = false,
    bool obscure = false,
    int maxLines = 1,
  }) {
    return TextFormField(
      controller: controller,
      obscureText: obscure,
      maxLines: maxLines,
      decoration: InputDecoration(
        labelText: required ? '$label *' : label,
        hintText: hint,
        border: const OutlineInputBorder(),
        suffixIcon: controller.text.isNotEmpty
            ? IconButton(
                icon: const Icon(Icons.copy),
                onPressed: () {
                  Clipboard.setData(ClipboardData(text: controller.text));
                  global.showInfoSnackBar(context, global.language('copied'));
                },
                tooltip: global.language('chatbot_copy'),
              )
            : null,
      ),
      validator: required
          ? (value) {
              if (value == null || value.isEmpty) {
                return '${global.language("please_fill")} $label';
              }
              return null;
            }
          : null,
    );
  }

  Widget _buildTestResult() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _testResult!.success
            ? Colors.green.shade50
            : Colors.red.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: _testResult!.success ? Colors.green : Colors.red,
        ),
      ),
      child: Row(
        children: [
          Icon(
            _testResult!.success ? Icons.check_circle : Icons.error,
            color: _testResult!.success ? Colors.green : Colors.red,
          ),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _testResult!.success
                      ? global.language('link_success')
                      : global.language('connection_error'),
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color:
                        _testResult!.success ? Colors.green : Colors.red,
                  ),
                ),
                Text(
                  _testResult!.message,
                  style: TextStyle(
                    fontSize: 13,
                    color: _testResult!.success
                        ? Colors.green.shade700
                        : Colors.red.shade700,
                  ),
                ),
                if (_testResult!.botInfo != null) ...[
                  const SizedBox(height: 8),
                  Text(
                    'Bot: ${_testResult!.botInfo!['displayName'] ?? 'N/A'}',
                    style: const TextStyle(fontSize: 13),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmployeeSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.people, color: global.theme.primaryColor),
            const SizedBox(width: 8),
            Text(
              global.language('connected_employees'),
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
            Spacer(),
            ElevatedButton.icon(
              onPressed: () => _showAddEmployeeDialog(),
              icon: Icon(Icons.person_add, size: 18),
              label: Text(global.language('add_employee')),
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.primaryColor,
                foregroundColor: global.theme.onPrimaryColor,
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        LineOAEmployeeListWidget(
          configGuid: widget.config.guid,
        ),
      ],
    );
  }

  void _showAddEmployeeDialog() {
    showDialog(
      context: context,
      builder: (context) => AddEmployeeDialog(
        configGuid: widget.config.guid,
        onAdded: () {
          // Refresh employee list
          setState(() {});
        },
      ),
    );
  }
}

/// Employee List Widget
class LineOAEmployeeListWidget extends StatefulWidget {
  final String configGuid;

  const LineOAEmployeeListWidget({
    super.key,
    required this.configGuid,
  });

  @override
  State<LineOAEmployeeListWidget> createState() =>
      _LineOAEmployeeListWidgetState();
}

class _LineOAEmployeeListWidgetState extends State<LineOAEmployeeListWidget> with global.ThemeRefreshMixin {
  bool _isLoading = true;
  List<LineOAEmployeeModel> _employees = [];

  @override
  void initState() {
    super.initState();
    _loadEmployees();
  }

  Future<void> _loadEmployees() async {
    setState(() => _isLoading = true);

    final result = await LineOAApiService.getEmployees(widget.configGuid);

    if (result.isSuccess) {
      setState(() {
        _employees = result.employees;
        _isLoading = false;
      });
    } else {
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(32),
          child: CircularProgressIndicator(),
        ),
      );
    }

    if (_employees.isEmpty) {
      return Container(
        padding: const EdgeInsets.all(32),
        decoration: BoxDecoration(
          color: global.theme.dividerBorderColor,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Center(
          child: Column(
            children: [
              Icon(Icons.people_outline, size: 48, color: global.theme.iconSecondaryColor),
              const SizedBox(height: 8),
              Text(
                global.language('no_connected_employees'),
                style: TextStyle(color: global.theme.textSecondaryColor),
              ),
            ],
          ),
        ),
      );
    }

    return ListView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: _employees.length,
      itemBuilder: (context, index) {
        final employee = _employees[index];
        return Card(
          child: ListTile(
            leading: CircleAvatar(
              backgroundImage: employee.pictureUrl != null
                  ? NetworkImage(employee.pictureUrl!)
                  : null,
              child: employee.pictureUrl == null
                  ? const Icon(Icons.person)
                  : null,
            ),
            title: Text(employee.employeeName),
            subtitle: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('${global.language("code_label")}: ${employee.employeeCode}'),
                if (employee.displayName != null)
                  Text('LINE: ${employee.displayName}'),
              ],
            ),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (employee.lineUserId.isEmpty)
                  IconButton(
                    icon: Icon(Icons.link, color: global.theme.primaryColor),
                    onPressed: () => _generateLink(employee),
                    tooltip: global.language('create_connect_link'),
                  ),
                IconButton(
                  icon: const Icon(Icons.delete, color: Colors.red),
                  onPressed: () => _removeEmployee(employee),
                  tooltip: global.language('delete'),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Future<void> _generateLink(LineOAEmployeeModel employee) async {
    if (!mounted) return;

    // Use new QR code dialog with LIFF integration
    // LIFF ID for lineoa-liff app - can be configured via environment or settings
    const liffId = '2008792333-tziDpgnJ';
    const liffBaseUrl = 'https://dev-api.bcaicloud.com/liff';

    await showLineLinkQRDialog(
      context: context,
      employeeCode: employee.employeeCode,
      employeeName: employee.employeeName,
      shopId: global.getShopId(),
      liffId: liffId,
      liffBaseUrl: liffBaseUrl,
      onLinked: (result) {
        // Refresh employee list when linked
        _loadEmployees();
        if (mounted) {
          global.showSuccessSnackBar(context, '${global.language("link_success")}: ${result.displayName}');
        }
      },
    );
  }

  Future<void> _removeEmployee(LineOAEmployeeModel employee) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('alert_confirm_delete')),
        content: Text('${global.language("confirm_delete")} ${employee.employeeName}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: Text(global.language('delete')),
          ),
        ],
      ),
    );

    if (confirm == true) {
      final success = await LineOAApiService.removeEmployee(employee.guid);
      if (success) {
        _loadEmployees();
        if (mounted) {
          global.showSuccessSnackBar(context, global.language('delete_success'));
        }
      }
    }
  }
}

/// Add Employee Dialog
class AddEmployeeDialog extends StatefulWidget {
  final String configGuid;
  final VoidCallback onAdded;

  const AddEmployeeDialog({
    super.key,
    required this.configGuid,
    required this.onAdded,
  });

  @override
  State<AddEmployeeDialog> createState() => _AddEmployeeDialogState();
}

class _AddEmployeeDialogState extends State<AddEmployeeDialog> with global.ThemeRefreshMixin {
  final _employeeCodeController = TextEditingController();
  final _employeeNameController = TextEditingController();
  bool _isLoading = false;

  @override
  void dispose() {
    _employeeCodeController.dispose();
    _employeeNameController.dispose();
    super.dispose();
  }

  Future<void> _addEmployee() async {
    if (_employeeCodeController.text.isEmpty ||
        _employeeNameController.text.isEmpty) {
      global.showWarningSnackBar(context, global.language('please_fill_all_fields'));
      return;
    }

    setState(() => _isLoading = true);

    final result = await LineOAApiService.addEmployee(
      lineOaConfigGuid: widget.configGuid,
      employeeCode: _employeeCodeController.text,
      employeeName: _employeeNameController.text,
    );

    setState(() => _isLoading = false);

    if (result.isSuccess) {
      widget.onAdded();
      if (mounted) Navigator.pop(context);
    } else {
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("add_failed")}: ${result.errorMessage}');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(global.language('add_employee')),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextField(
            controller: _employeeCodeController,
            decoration: InputDecoration(
              labelText: global.language('emp_code'),
              border: OutlineInputBorder(),
            ),
          ),
          SizedBox(height: 16),
          TextField(
            controller: _employeeNameController,
            decoration: InputDecoration(
              labelText: global.language('employee_name'),
              border: OutlineInputBorder(),
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(global.language('cancel')),
        ),
        ElevatedButton(
          onPressed: _isLoading ? null : _addEmployee,
          child: _isLoading
              ? const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : Text(global.language('add')),
        ),
      ],
    );
  }
}
