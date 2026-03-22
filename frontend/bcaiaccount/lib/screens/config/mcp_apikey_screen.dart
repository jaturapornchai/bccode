import 'dart:convert';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import '../../global.dart' as global;
import '../../modules/mcp/models/mcp_apikey_model.dart';
import '../../modules/mcp/services/mcp_apikey_service.dart';

/// MCP API Key Management Screen
/// Allows users to create, view, edit and delete API keys for MCP access
class MCPAPIKeyScreen extends StatefulWidget {
  const MCPAPIKeyScreen({super.key});

  @override
  State<MCPAPIKeyScreen> createState() => _MCPAPIKeyScreenState();
}

class _MCPAPIKeyScreenState extends State<MCPAPIKeyScreen>
    with global.ThemeRefreshMixin {
  final MCPAPIKeyService _service = MCPAPIKeyService();
  List<MCPAPIKeyModel> _apiKeys = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadAPIKeys();
  }

  Future<void> _loadAPIKeys() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final keys = await _service.getAPIKeys(global.prefs.getString("shopid") ?? "");
      setState(() {
        _apiKeys = keys;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<void> _createAPIKey() async {
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) => const _CreateAPIKeyDialog(),
    );

    if (result != null) {
      if (mounted) {
        await _showNewAPIKeyWithConfigDialog(result);
        _loadAPIKeys();
      }
    }
  }

  /// แสดง dialog หลังสร้าง API key สำเร็จ — แสดงทั้ง key และ Claude Desktop config
  Future<void> _showNewAPIKeyWithConfigDialog(Map<String, dynamic> data) async {
    final apiKey = data['api_key'] as String? ?? '';
    final exportData = data['export'] as Map<String, dynamic>?;
    final claudeConfig = exportData?['claude_desktop_config'];
    final configJson = claudeConfig != null
        ? const JsonEncoder.withIndent('  ').convert(claudeConfig)
        : '';

    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.check_circle, color: global.theme.positiveHighlightTextColor),
            SizedBox(width: 8),
            Expanded(child: Text(global.language('api_key_created'))),
          ],
        ),
        content: SizedBox(
          width: 500,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language('api_key_warning'),
                  style: TextStyle(color: global.theme.warningHighlightTextColor, fontWeight: FontWeight.bold),
                ),
                SizedBox(height: 16),
                // API Key
                Text('API Key:', style: TextStyle(fontWeight: FontWeight.bold)),
                SizedBox(height: 4),
                _buildCopyableBox(context, apiKey),
                if (configJson.isNotEmpty) ...[
                  SizedBox(height: 20),
                  Text('Claude Desktop Config:', style: TextStyle(fontWeight: FontWeight.bold)),
                  SizedBox(height: 4),
                  Text(
                    global.language('mcp_config_copy_hint'),
                    style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
                  ),
                  SizedBox(height: 8),
                  _buildCopyableBox(context, configJson, maxLines: 12),
                ],
              ],
            ),
          ),
        ),
        actions: [
          ElevatedButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('ok')),
          ),
        ],
      ),
    );
  }

  Widget _buildCopyableBox(BuildContext context, String text, {int maxLines = 2}) {
    return Container(
      padding: EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: SelectableText(
              text,
              style: TextStyle(fontFamily: 'monospace', fontSize: 12),
              maxLines: maxLines,
            ),
          ),
          IconButton(
            icon: Icon(Icons.copy),
            onPressed: () {
              Clipboard.setData(ClipboardData(text: text));
              global.showInfoSnackBar(context, global.language('copied_to_clipboard'));
            },
          ),
        ],
      ),
    );
  }

  /// Export config สำหรับ key ที่มีอยู่
  Future<void> _exportConfig(MCPAPIKeyModel key) async {
    try {
      final data = await _service.exportAPIKey(key.id);
      final claudeConfig = data['claude_desktop_config'];
      final configJson = claudeConfig != null
          ? const JsonEncoder.withIndent('  ').convert(claudeConfig)
          : '';

      if (!mounted) return;

      await showDialog(
        context: context,
        builder: (context) => AlertDialog(
          title: Row(
            children: [
              Icon(Icons.settings, color: global.theme.infoHighlightTextColor),
              SizedBox(width: 8),
              Text('Claude Desktop Config'),
            ],
          ),
          content: SizedBox(
            width: 500,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    global.language('mcp_config_copy_hint'),
                    style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
                  ),
                  SizedBox(height: 12),
                  _buildCopyableBox(context, configJson, maxLines: 15),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('close')),
            ),
          ],
        ),
      );
    } catch (e) {
      if (mounted) {
        global.showErrorSnackBar(context, 'Error: $e');
      }
    }
  }

  Future<void> _toggleKeyStatus(MCPAPIKeyModel key) async {
    try {
      await _service.updateAPIKey(
        key.id,
        UpdateAPIKeyRequest(isActive: !key.isActive),
      );
      _loadAPIKeys();
      if (mounted) {
        global.showInfoSnackBar(context, key.isActive
            ? global.language('api_key_deactivated')
            : global.language('api_key_activated'));
      }
    } catch (e) {
      if (mounted) {
        global.showErrorSnackBar(context, 'Error: $e');
      }
    }
  }

  Future<void> _deleteKey(MCPAPIKeyModel key) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('confirm_delete')),
        content: Text(global.language('confirm_delete_api_key').replaceAll('{name}', key.name)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
            onPressed: () => Navigator.pop(context, true),
            child: Text(global.language('delete')),
          ),
        ],
      ),
    );

    if (confirm == true) {
      try {
        await _service.deleteAPIKey(key.id);
        _loadAPIKeys();
        if (mounted) {
          global.showInfoSnackBar(context, global.language('api_key_deleted'));
        }
      } catch (e) {
        if (mounted) {
          global.showErrorSnackBar(context, 'Error: $e');
        }
      }
    }
  }

  void _showKeyDetails(MCPAPIKeyModel key) {
    showDialog(
      context: context,
      builder: (context) => _APIKeyDetailsDialog(apiKey: key),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Text(global.language('mcp_api_keys')),
        actions: [
          IconButton(
            icon: Icon(Icons.refresh),
            onPressed: _loadAPIKeys,
            tooltip: global.language('refresh'),
          ),
          const ManualButton(path: 'settings-mcp-token'),
        ],
      ),
      body: _buildBody(),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _createAPIKey,
        icon: Icon(Icons.add),
        label: Text(global.language('create_api_key')),
        backgroundColor: global.theme.appBarColor,
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.error_outline, size: 64, color: global.theme.negativeHighlightTextColor),
            SizedBox(height: 16),
            Text(_errorMessage!, textAlign: TextAlign.center),
            SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadAPIKeys,
              child: Text(global.language('retry')),
            ),
          ],
        ),
      );
    }

    if (_apiKeys.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.vpn_key_off, size: 64, color: global.theme.iconSecondaryColor),
            SizedBox(height: 16),
            Text(
              global.language('no_api_keys'),
              style: TextStyle(fontSize: 18, color: global.theme.iconSecondaryColor),
            ),
            SizedBox(height: 8),
            Text(
              global.language('create_first_api_key'),
              style: TextStyle(color: global.theme.iconSecondaryColor),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadAPIKeys,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: _apiKeys.length + 1,
        itemBuilder: (context, index) {
          if (index == 0) return _buildExampleConfigSection();
          final key = _apiKeys[index - 1];
          return _buildAPIKeyCard(key);
        },
      ),
    );
  }

  Widget _buildExampleConfigSection() {
    // ใช้ URL จากตั้งค่าระบบ (BackendUrlManager) แทน hardcode
    final backendUrl = global.goApiUrlPath('');
    final cleanUrl = backendUrl.endsWith('/') ? backendUrl.substring(0, backendUrl.length - 1) : backendUrl;
    final exampleConfig = JsonEncoder.withIndent('  ').convert({
      "mcpServers": {
        "bc-erp": {
          "command": "node",
          "args": ["/path/to/backend/mcp-server/index.js"],
          "env": {
            "BC_API_URL": cleanUrl,
            "BC_API_KEY": "bc_live_xxxxx..."
          }
        }
      }
    });

    return Card(
      margin: EdgeInsets.only(bottom: 16),
      color: global.theme.infoHighlightColor,
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.info_outline, color: global.theme.infoHighlightTextColor, size: 20),
                SizedBox(width: 8),
                Text(
                  global.language('mcp_usage_guide'),
                  style: TextStyle(fontSize: 13, color: global.theme.infoHighlightTextColor),
                ),
              ],
            ),
            SizedBox(height: 12),
            Text(
              global.language('mcp_example_config'),
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
            ),
            SizedBox(height: 8),
            _buildCopyableBox(context, exampleConfig, maxLines: 15),
          ],
        ),
      ),
    );
  }

  Widget _buildAPIKeyCard(MCPAPIKeyModel key) {
    final dateFormat = DateFormat('dd/MM/yyyy HH:mm');
    final statusColor = key.isActive && !key.isExpired ? global.theme.positiveHighlightTextColor : global.theme.negativeHighlightTextColor;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: () => _showKeyDetails(key),
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header row
              Row(
                children: [
                  Icon(Icons.vpn_key, size: 24),
                  SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          key.name,
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        if (key.description != null && key.description!.isNotEmpty)
                          Text(
                            key.description!,
                            style: TextStyle(color: global.theme.textSecondaryColor, fontSize: 13),
                          ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: statusColor.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: statusColor),
                    ),
                    child: Text(
                      key.statusText,
                      style: TextStyle(
                        color: statusColor,
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ],
              ),
              SizedBox(height: 12),
              // Info row
              Row(
                children: [
                  _buildPermissionBadge(key.allowedTools),
                  SizedBox(width: 8),
                  _buildInfoChip(Icons.speed, '${key.rateLimitPerMinute}/min'),
                  const Spacer(),
                  Text(
                    dateFormat.format(key.createdAt.toLocal()),
                    style: TextStyle(color: global.theme.textSecondaryColor, fontSize: 11),
                  ),
                ],
              ),
              SizedBox(height: 12),
              // Actions
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton.icon(
                    onPressed: () => _exportConfig(key),
                    icon: Icon(Icons.settings, size: 18, color: global.theme.infoHighlightTextColor),
                    label: Text(
                      global.language('export_config'),
                      style: TextStyle(color: global.theme.infoHighlightTextColor),
                    ),
                  ),
                  TextButton.icon(
                    onPressed: () => _toggleKeyStatus(key),
                    icon: Icon(
                      key.isActive ? Icons.pause : Icons.play_arrow,
                      size: 18,
                    ),
                    label: Text(key.isActive
                      ? global.language('deactivate')
                      : global.language('activate')),
                  ),
                  TextButton.icon(
                    onPressed: () => _deleteKey(key),
                    icon: Icon(Icons.delete, size: 18, color: global.theme.negativeHighlightTextColor),
                    label: Text(
                      global.language('delete'),
                      style: TextStyle(color: global.theme.negativeHighlightTextColor),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildPermissionBadge(List<String> allowedTools) {
    final preset = MCPTools.getPresetType(allowedTools);
    final label = MCPTools.getPermissionLabel(allowedTools);
    Color color;
    IconData icon;
    switch (preset) {
      case 'readonly':
        color = global.theme.infoHighlightTextColor;
        icon = Icons.visibility;
        break;
      case 'developer':
        color = global.theme.warningHighlightTextColor;
        icon = Icons.build;
        break;
      default:
        color = Colors.purple;
        icon = Icons.tune;
    }
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 13, color: color),
          SizedBox(width: 4),
          Text(label, style: TextStyle(fontSize: 11, color: color, fontWeight: FontWeight.bold)),
        ],
      ),
    );
  }

  Widget _buildInfoChip(IconData icon, String text) {
    return Container(
      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.blue.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: global.theme.infoHighlightTextColor),
          SizedBox(width: 4),
          Text(text, style: TextStyle(fontSize: 12, color: global.theme.infoHighlightTextColor)),
        ],
      ),
    );
  }
}

/// Dialog for creating a new API key
class _CreateAPIKeyDialog extends StatefulWidget {
  const _CreateAPIKeyDialog();

  @override
  State<_CreateAPIKeyDialog> createState() => _CreateAPIKeyDialogState();
}

class _CreateAPIKeyDialogState extends State<_CreateAPIKeyDialog>
    with global.ThemeRefreshMixin {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _rateLimitController = TextEditingController(text: '600');
  final Set<String> _selectedTools = {};
  DateTime? _expiresAt;
  bool _isLoading = false;
  String _preset = 'readonly'; // readonly, developer, custom

  IconData _getCategoryIcon(String category) {
    switch (category) {
      case 'Sales & Products': return Icons.point_of_sale;
      case 'Dashboard & KPIs': return Icons.dashboard;
      case 'Financial': return Icons.account_balance;
      case 'Inventory': return Icons.inventory;
      case 'Customers': return Icons.people;
      case 'Comparison': return Icons.compare_arrows;
      case 'Unit of Measure': return Icons.straighten;
      case 'API Development': return Icons.api;
      case 'PostgreSQL': return Icons.storage;
      case 'MongoDB': return Icons.dns;
      case 'ClickHouse': return Icons.analytics;
      default: return Icons.build;
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _descriptionController.dispose();
    _rateLimitController.dispose();
    super.dispose();
  }

  /// สร้าง allowed_tools ตาม preset ที่เลือก
  List<String> _buildAllowedTools() {
    switch (_preset) {
      case 'readonly':
        return ['readonly'];
      case 'developer':
        return ['*'];
      case 'custom':
        return _selectedTools.toList();
      default:
        return ['readonly'];
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    try {
      final service = MCPAPIKeyService();
      final data = await service.createAPIKeyWithExport(CreateAPIKeyRequest(
        shopId: global.prefs.getString("shopid") ?? "",
        name: _nameController.text,
        description: _descriptionController.text.isEmpty ? null : _descriptionController.text,
        allowedTools: _buildAllowedTools(),
        rateLimitPerMinute: int.tryParse(_rateLimitController.text) ?? 600,
        expiresAt: _expiresAt != null
          ? DateFormat('yyyy-MM-dd').format(_expiresAt!)
          : null,
        createdBy: global.prefs.getString("username"),
      ));

      if (mounted) {
        Navigator.pop(context, data);
      }
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        global.showErrorSnackBar(context, 'Error: $e');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(global.language('create_api_key')),
      content: SizedBox(
        width: 450,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Name
                TextFormField(
                  controller: _nameController,
                  decoration: InputDecoration(
                    labelText: global.language('name'),
                    hintText: global.language('example_claude_desktop_key'),
                    border: const OutlineInputBorder(),
                  ),
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return global.language('required');
                    }
                    return null;
                  },
                ),
                SizedBox(height: 16),

                // Description
                TextFormField(
                  controller: _descriptionController,
                  decoration: InputDecoration(
                    labelText: global.language('description'),
                    hintText: global.language('optional'),
                    border: const OutlineInputBorder(),
                  ),
                  maxLines: 2,
                ),
                SizedBox(height: 16),

                // Rate Limit
                TextFormField(
                  controller: _rateLimitController,
                  decoration: InputDecoration(
                    labelText: global.language('rate_limit'),
                    suffixText: 'req/min',
                    border: const OutlineInputBorder(),
                  ),
                  keyboardType: TextInputType.number,
                  inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                ),
                SizedBox(height: 16),

                // Expiration
                Row(
                  children: [
                    Expanded(
                      child: CustomDatePicker(
                        labelText: global.language('expires_at'),
                        useIconSelectDate: true,
                        initialDate: _expiresAt ?? DateTime.now().add(const Duration(days: 365)),
                        firstDate: DateTime.now(),
                        lastDate: DateTime.now().add(const Duration(days: 3650)),
                        onDateSelected: (date) {
                          if (date != null) {
                            setState(() => _expiresAt = date);
                          }
                        },
                        decoration: const InputDecoration(),
                      ),
                    ),
                    if (_expiresAt != null)
                      IconButton(
                        icon: Icon(Icons.clear, color: global.theme.iconSecondaryColor),
                        onPressed: () => setState(() => _expiresAt = null),
                      ),
                  ],
                ),
                const Divider(),

                // ===== Permission Preset =====
                Text(global.language('permission_level'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                SizedBox(height: 8),
                _buildPresetOption(
                  value: 'readonly',
                  icon: Icons.visibility,
                  color: global.theme.infoHighlightTextColor,
                  title: global.language('role_readonly'),
                  subtitle: '${global.language("mcp_readonly_desc")} (${MCPTools.readonlyTools.length} tools)',
                ),
                _buildPresetOption(
                  value: 'developer',
                  icon: Icons.build,
                  color: global.theme.warningHighlightTextColor,
                  title: global.language('role_developer'),
                  subtitle: '${global.language("mcp_readwrite_desc")} (${MCPTools.allTools.length} tools)',
                ),
                _buildPresetOption(
                  value: 'custom',
                  icon: Icons.tune,
                  color: Colors.purple,
                  title: global.language('role_custom'),
                  subtitle: global.language('mcp_custom_desc'),
                ),

                // ===== Custom Tool Selector (เฉพาะ preset = custom) =====
                if (_preset == 'custom') ...[
                  SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          '${global.language("select_tools")} (${_selectedTools.length}/${MCPTools.allTools.length})',
                          style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
                        ),
                      ),
                      TextButton(
                        onPressed: () => setState(() => _selectedTools.addAll(MCPTools.allTools)),
                        child: Text(global.language('all'), style: TextStyle(fontSize: 11)),
                      ),
                      TextButton(
                        onPressed: () => setState(() {
                          _selectedTools.clear();
                          _selectedTools.addAll(MCPTools.readonlyTools);
                        }),
                        child: Text(global.language('role_readonly'), style: TextStyle(fontSize: 11)),
                      ),
                      TextButton(
                        onPressed: () => setState(() => _selectedTools.clear()),
                        child: Text(global.language('clear'), style: TextStyle(fontSize: 11, color: global.theme.negativeHighlightTextColor)),
                      ),
                    ],
                  ),
                  SizedBox(height: 4),
                  ...MCPTools.toolCategories.entries.expand((category) => [
                    Padding(
                      padding: EdgeInsets.only(top: 8, bottom: 2),
                      child: Row(
                        children: [
                          Icon(_getCategoryIcon(category.key), size: 16, color: global.theme.iconSecondaryColor),
                          SizedBox(width: 6),
                          Text(
                            category.key,
                            style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: global.theme.textSecondaryColor),
                          ),
                          SizedBox(width: 8),
                          Expanded(child: Divider(color: global.theme.dividerBorderColor)),
                        ],
                      ),
                    ),
                    ...category.value.map((tool) {
                      final isWrite = MCPTools.isWriteTool(tool);
                      return CheckboxListTile(
                        title: Row(
                          children: [
                            Expanded(child: Text(MCPTools.getToolName(tool), style: TextStyle(fontSize: 13))),
                            if (isWrite)
                              Container(
                                padding: EdgeInsets.symmetric(horizontal: 5, vertical: 1),
                                decoration: BoxDecoration(
                                  color: Colors.red.withValues(alpha: 0.1),
                                  borderRadius: BorderRadius.circular(4),
                                  border: Border.all(color: Colors.red.withValues(alpha: 0.3)),
                                ),
                                child: Text('write', style: TextStyle(fontSize: 9, color: global.theme.negativeHighlightTextColor, fontWeight: FontWeight.bold)),
                              ),
                          ],
                        ),
                        subtitle: Text(MCPTools.getToolDescription(tool), style: TextStyle(fontSize: 11)),
                        value: _selectedTools.contains(tool),
                        dense: true,
                        visualDensity: VisualDensity.compact,
                        onChanged: (value) {
                          setState(() {
                            if (value == true) {
                              _selectedTools.add(tool);
                            } else {
                              _selectedTools.remove(tool);
                            }
                          });
                        },
                      );
                    }),
                  ]),
                ],
              ],
            ),
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isLoading ? null : () => Navigator.pop(context),
          child: Text(global.language('cancel')),
        ),
        ElevatedButton(
          onPressed: _isLoading ? null : _submit,
          child: _isLoading
            ? SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
            : Text(global.language('create')),
        ),
      ],
    );
  }

  Widget _buildPresetOption({
    required String value,
    required IconData icon,
    required Color color,
    required String title,
    required String subtitle,
  }) {
    final selected = _preset == value;
    return GestureDetector(
      onTap: () => setState(() => _preset = value),
      child: Container(
        margin: EdgeInsets.only(bottom: 8),
        padding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(10),
          border: Border.all(
            color: selected ? color : global.theme.dividerBorderColor,
            width: selected ? 2 : 1,
          ),
          color: selected ? color.withValues(alpha: 0.05) : null,
        ),
        child: Row(
          children: [
            Icon(icon, color: selected ? color : global.theme.iconSecondaryColor, size: 22),
            SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: selected ? color : null)),
                  Text(subtitle, style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor)),
                ],
              ),
            ),
            Radio<String>(
              value: value,
              groupValue: _preset,
              onChanged: (v) => setState(() => _preset = v!),
              activeColor: color,
            ),
          ],
        ),
      ),
    );
  }
}

/// Dialog for viewing API key details
class _APIKeyDetailsDialog extends StatelessWidget {
  final MCPAPIKeyModel apiKey;

  const _APIKeyDetailsDialog({required this.apiKey});

  @override
  Widget build(BuildContext context) {
    final dateFormat = DateFormat('dd/MM/yyyy HH:mm');

    return AlertDialog(
      title: Text(apiKey.name),
      content: SizedBox(
        width: 400,
        child: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              if (apiKey.description != null && apiKey.description!.isNotEmpty) ...[
                Text(apiKey.description!, style: TextStyle(color: global.theme.textSecondaryColor)),
                SizedBox(height: 16),
              ],

              _buildDetailRow(global.language('status'), apiKey.statusText),
              _buildDetailRow('Permission', MCPTools.getPermissionLabel(apiKey.allowedTools)),
              _buildDetailRow(global.language('rate_limit'), '${apiKey.rateLimitPerMinute} req/min'),
              _buildDetailRow(global.language('created_at'), dateFormat.format(apiKey.createdAt.toLocal())),
              if (apiKey.expiresAt != null)
                _buildDetailRow(global.language('expires_at'), dateFormat.format(apiKey.expiresAt!.toLocal())),
              if (apiKey.lastUsedAt != null)
                _buildDetailRow(global.language('last_used'), dateFormat.format(apiKey.lastUsedAt!.toLocal())),
              if (apiKey.createdBy != null)
                _buildDetailRow(global.language('created_by'), apiKey.createdBy!),

              SizedBox(height: 16),
              Text(
                global.language('allowed_tools'),
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              SizedBox(height: 8),
              Wrap(
                spacing: 6,
                runSpacing: 6,
                children: apiKey.allowedTools.map((tool) {
                  // keyword presets
                  if (tool == 'readonly' || tool == '*' || tool == 'readwrite') {
                    return Chip(
                      avatar: Icon(
                        tool == 'readonly' ? Icons.visibility : Icons.build,
                        size: 14,
                        color: tool == 'readonly' ? global.theme.infoHighlightTextColor : global.theme.warningHighlightTextColor,
                      ),
                      label: Text(tool, style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
                      backgroundColor: (tool == 'readonly' ? global.theme.infoHighlightTextColor : global.theme.warningHighlightTextColor).withValues(alpha: 0.1),
                    );
                  }
                  final isWrite = MCPTools.isWriteTool(tool);
                  return Chip(
                    label: Text(MCPTools.getToolName(tool), style: TextStyle(fontSize: 11)),
                    backgroundColor: isWrite ? Colors.red.withValues(alpha: 0.1) : Colors.blue.withValues(alpha: 0.1),
                    side: isWrite ? BorderSide(color: Colors.red.withValues(alpha: 0.3)) : BorderSide.none,
                  );
                }).toList(),
              ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(global.language('close')),
        ),
      ],
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child: Text(label, style: TextStyle(color: global.theme.textSecondaryColor)),
          ),
          Expanded(child: Text(value, style: TextStyle(fontWeight: FontWeight.w500))),
        ],
      ),
    );
  }
}
