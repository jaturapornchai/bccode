import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import '../../global.dart' as global;

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

class _AIProvider {
  final String providerName;
  String apiKey;
  String baseUrl;
  String model;
  List<String> capabilities;
  bool isActive;
  int priority;
  final String? cooldownUntil; // ISO8601 string
  final String? lastError;

  _AIProvider({
    required this.providerName,
    required this.apiKey,
    this.baseUrl = '',
    required this.model,
    this.capabilities = const [],
    required this.isActive,
    required this.priority,
    this.cooldownUntil,
    this.lastError,
  });

  factory _AIProvider.fromJson(Map<String, dynamic> json) {
    return _AIProvider(
      providerName: json['provider_name'] as String? ?? '',
      apiKey: json['api_key_masked'] as String? ?? json['api_key'] as String? ?? '',
      baseUrl: json['base_url'] as String? ?? '',
      model: json['model'] as String? ?? '',
      capabilities: (json['capabilities'] as List?)?.map((e) => e.toString()).toList() ?? [],
      isActive: json['is_active'] as bool? ?? false,
      priority: (json['priority'] as num?)?.toInt() ?? 0,
      cooldownUntil: json['cooldown_until']?.toString(),
      lastError: json['last_error'] as String?,
    );
  }

  bool get isCustom => providerName.startsWith('custom');
}

// ---------------------------------------------------------------------------
// Provider Metadata (hardcoded)
// ---------------------------------------------------------------------------

class _ProviderMeta {
  final String name;
  final IconData icon;
  final Color color;
  final String defaultModel;
  final String description;

  const _ProviderMeta({
    required this.name,
    required this.icon,
    required this.color,
    required this.defaultModel,
    required this.description,
  });
}

const Map<String, _ProviderMeta> _kProviderMeta = {
  'groq': _ProviderMeta(
    name: 'Groq',
    icon: Icons.flash_on,
    color: Colors.orange,
    defaultModel: 'llama-3.3-70b-versatile',
    description: 'เร็วมาก มี free tier',
  ),
  'openrouter': _ProviderMeta(
    name: 'OpenRouter',
    icon: Icons.router,
    color: Colors.purple,
    defaultModel: 'google/gemini-2.0-flash-exp:free',
    description: 'หลาย model มี free tier',
  ),
  'deepseek': _ProviderMeta(
    name: 'DeepSeek',
    icon: Icons.psychology,
    color: Colors.blue,
    defaultModel: 'deepseek-chat',
    description: 'ราคาถูก คุณภาพดี',
  ),
  'gemini': _ProviderMeta(
    name: 'Google Gemini',
    icon: Icons.auto_awesome,
    color: Colors.teal,
    defaultModel: 'gemini-2.0-flash',
    description: 'Google AI ฟรี',
  ),
  'ollama': _ProviderMeta(
    name: 'Ollama',
    icon: Icons.computer,
    color: Colors.grey,
    defaultModel: 'qwen2.5:7b',
    description: 'รัน AI บนเครื่องตัวเอง (ฟรี)',
  ),
  'custom': _ProviderMeta(
    name: 'กำหนดเอง',
    icon: Icons.settings_ethernet,
    color: Colors.indigo,
    defaultModel: 'auto',
    description: 'ตั้ง Base URL เอง (OpenAI-compatible)',
  ),
};

// Capability display info: (label, icon, color)
(String, IconData, Color) _capabilityInfo(String cap) {
  return switch (cap) {
    'tools'    => ('Tools', Icons.build, Colors.blue),
    'vision'   => ('Vision', Icons.visibility, Colors.purple),
    'thinking' => ('Thinking', Icons.psychology, Colors.orange),
    _          => (cap, Icons.extension, Colors.grey),
  };
}

// ---------------------------------------------------------------------------
// Screen
// ---------------------------------------------------------------------------

class AIProviderScreen extends StatefulWidget {
  const AIProviderScreen({super.key});

  @override
  State<AIProviderScreen> createState() => _AIProviderScreenState();
}

class _AIProviderScreenState extends State<AIProviderScreen>
    with global.ThemeRefreshMixin {
  List<_AIProvider> _providers = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadProviders();
  }

  // -------------------------------------------------------------------------
  // API calls
  // -------------------------------------------------------------------------

  Future<void> _loadProviders() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final url = Uri.parse(global.goApiUrlPath('api/v1/ai-provider/list'));
      final response = await http.post(
        url,
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'shop_id': global.getShopId()}),
      ).timeout(const Duration(seconds: 15));

      final body = jsonDecode(response.body);
      final List<dynamic> items = body['providers'] ?? [];
      setState(() {
        _providers = items
            .map((e) => _AIProvider.fromJson(e as Map<String, dynamic>))
            .toList()
          ..sort((a, b) => a.priority.compareTo(b.priority));
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<void> _saveProvider(Map<String, dynamic> data) async {
    final url = Uri.parse(global.goApiUrlPath('api/v1/ai-provider/save'));
    await http.post(
      url,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({...data, 'shop_id': global.getShopId()}),
    ).timeout(const Duration(seconds: 15));
  }

  Future<Map<String, dynamic>> _testProvider(
      String providerName, String apiKey, String model) async {
    final url = Uri.parse(global.goApiUrlPath('api/v1/ai-provider/test'));
    final response = await http.post(
      url,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'shop_id': global.getShopId(),
        'provider_name': providerName,
        'api_key': apiKey,
        'model': model,
      }),
    ).timeout(const Duration(seconds: 30));
    return jsonDecode(response.body) as Map<String, dynamic>;
  }

  Future<void> _deleteProvider(String providerName) async {
    final url = Uri.parse(global.goApiUrlPath('api/v1/ai-provider/delete'));
    await http.post(
      url,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'shop_id': global.getShopId(),
        'provider_name': providerName,
      }),
    ).timeout(const Duration(seconds: 15));
  }

  // -------------------------------------------------------------------------
  // Actions
  // -------------------------------------------------------------------------

  Future<void> _onToggleActive(_AIProvider p) async {
    try {
      await _saveProvider({
        'provider_name': p.providerName,
        'api_key': p.apiKey,
        'model': p.model,
        'is_active': !p.isActive,
        'priority': p.priority,
      });
      await _loadProviders();
      if (mounted) {
        global.showInfoSnackBar(
            context, p.isActive ? 'ปิดใช้งาน ${_metaOf(p.providerName).name} แล้ว' : 'เปิดใช้งาน ${_metaOf(p.providerName).name} แล้ว');
      }
    } catch (e) {
      if (mounted) global.showErrorSnackBar(context, 'Error: $e');
    }
  }

  Future<void> _onMovePriority(_AIProvider p, int direction) async {
    final sorted = List<_AIProvider>.from(_providers)
      ..sort((a, b) => a.priority.compareTo(b.priority));
    final idx = sorted.indexOf(p);
    final swapIdx = idx + direction;
    if (swapIdx < 0 || swapIdx >= sorted.length) return;

    final swapTarget = sorted[swapIdx];
    final tmpPriority = p.priority;

    try {
      await _saveProvider({
        'provider_name': p.providerName,
        'api_key': p.apiKey,
        'model': p.model,
        'is_active': p.isActive,
        'priority': swapTarget.priority,
      });
      await _saveProvider({
        'provider_name': swapTarget.providerName,
        'api_key': swapTarget.apiKey,
        'model': swapTarget.model,
        'is_active': swapTarget.isActive,
        'priority': tmpPriority,
      });
      await _loadProviders();
    } catch (e) {
      if (mounted) global.showErrorSnackBar(context, 'Error: $e');
    }
  }

  Future<void> _onSaveEdit(_AIProvider p, String newApiKey, String newModel, {String? newBaseUrl}) async {
    try {
      await _saveProvider({
        'provider_name': p.providerName,
        'api_key': newApiKey,
        'base_url': newBaseUrl ?? p.baseUrl,
        'model': newModel,
        'is_active': p.isActive,
        'priority': p.priority,
      });
      await _loadProviders();
      if (mounted) {
        global.showInfoSnackBar(context, 'บันทึก ${_metaOf(p.providerName).name} สำเร็จ');
      }
    } catch (e) {
      if (mounted) global.showErrorSnackBar(context, 'Error: $e');
    }
  }

  Future<void> _onDelete(_AIProvider p) async {
    final meta = _metaOf(p.providerName);
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('ยืนยันการลบ'),
        content: Text('ต้องการลบ ${meta.name} ออกจากระบบ?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('ยกเลิก'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.negativeHighlightTextColor),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('ลบ'),
          ),
        ],
      ),
    );
    if (confirm != true) return;
    try {
      await _deleteProvider(p.providerName);
      await _loadProviders();
      if (mounted) {
        global.showInfoSnackBar(context, 'ลบ ${meta.name} สำเร็จ');
      }
    } catch (e) {
      if (mounted) global.showErrorSnackBar(context, 'Error: $e');
    }
  }

  Future<void> _onAddProvider() async {
    final existingNames = _providers.map((p) => p.providerName).toSet();
    // Custom สามารถเพิ่มได้หลายตัว (ไม่ filter ออก)
    final available = _kProviderMeta.entries
        .where((e) => e.key == 'custom' || !existingNames.contains(e.key))
        .toList();

    if (available.isEmpty) {
      global.showInfoSnackBar(context, 'เพิ่ม Provider ครบทุกตัวแล้ว');
      return;
    }

    final nextPriority =
        _providers.isEmpty ? 1 : (_providers.map((p) => p.priority).reduce((a, b) => a > b ? a : b) + 1);

    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (ctx) => _AddProviderDialog(
        availableProviders: available,
        nextPriority: nextPriority,
      ),
    );

    if (result != null) {
      try {
        await _saveProvider(result);
        await _loadProviders();
        if (mounted) {
          final name = result['provider_name'] as String;
          final displayName = _metaOf(name).name;
          global.showInfoSnackBar(context, 'เพิ่ม $displayName สำเร็จ');
        }
      } catch (e) {
        if (mounted) global.showErrorSnackBar(context, 'Error: $e');
      }
    }
  }

  // -------------------------------------------------------------------------
  // Helpers
  // -------------------------------------------------------------------------

  _ProviderMeta _metaOf(String name) {
    if (_kProviderMeta.containsKey(name)) return _kProviderMeta[name]!;
    // custom-xxx → ใช้ custom metadata
    if (name.startsWith('custom')) return _kProviderMeta['custom']!;
    return const _ProviderMeta(
      name: 'Unknown',
      icon: Icons.cloud,
      color: Colors.grey,
      defaultModel: '',
      description: '',
    );
  }

  // -------------------------------------------------------------------------
  // Build
  // -------------------------------------------------------------------------

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: const Text('AI Provider'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadProviders,
            tooltip: 'รีเฟรช',
          ),
        ],
      ),
      body: _buildBody(),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _onAddProvider,
        icon: const Icon(Icons.add),
        label: const Text('เพิ่ม Provider'),
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
            Icon(Icons.error_outline,
                size: 64, color: global.theme.negativeHighlightTextColor),
            const SizedBox(height: 16),
            Text(_errorMessage!, textAlign: TextAlign.center),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadProviders,
              child: const Text('ลองใหม่'),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadProviders,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 100),
        children: [
          _buildInfoCard(),
          const SizedBox(height: 16),
          if (_providers.isEmpty) _buildEmptyState(),
          ..._providers
              .asMap()
              .entries
              .map((entry) => _ProviderCard(
                    key: ValueKey(entry.value.providerName),
                    provider: entry.value,
                    index: entry.key,
                    total: _providers.length,
                    meta: _metaOf(entry.value.providerName),
                    onToggleActive: () => _onToggleActive(entry.value),
                    onMoveUp: entry.key > 0
                        ? () => _onMovePriority(entry.value, -1)
                        : null,
                    onMoveDown: entry.key < _providers.length - 1
                        ? () => _onMovePriority(entry.value, 1)
                        : null,
                    onDelete: () => _onDelete(entry.value),
                    onSave: (apiKey, model, {String? newBaseUrl}) =>
                        _onSaveEdit(entry.value, apiKey, model, newBaseUrl: newBaseUrl),
                    onTest: _testProvider,
                  )),
        ],
      ),
    );
  }

  Widget _buildInfoCard() {
    return Card(
      color: global.theme.infoHighlightColor,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(Icons.info_outline,
                color: global.theme.infoHighlightTextColor, size: 22),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                'ตั้งค่า AI Provider สำหรับ AI Assistant (Chatbot)\n'
                'ระบบจะใช้ตามลำดับความสำคัญ ถ้าตัวแรกใช้ไม่ได้จะไปตัวถัดไปอัตโนมัติ (Fallback)',
                style: TextStyle(
                    fontSize: 13, color: global.theme.infoHighlightTextColor),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 48),
      child: Column(
        children: [
          Icon(Icons.cloud_off, size: 64, color: global.theme.iconSecondaryColor),
          const SizedBox(height: 16),
          Text(
            'ยังไม่มี AI Provider',
            style: TextStyle(
                fontSize: 18, color: global.theme.iconSecondaryColor),
          ),
          const SizedBox(height: 8),
          Text(
            'กดปุ่ม + เพื่อเพิ่ม Provider',
            style: TextStyle(color: global.theme.textSecondaryColor),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Provider Card
// ---------------------------------------------------------------------------

class _ProviderCard extends StatefulWidget {
  final _AIProvider provider;
  final int index;
  final int total;
  final _ProviderMeta meta;
  final VoidCallback onToggleActive;
  final VoidCallback? onMoveUp;
  final VoidCallback? onMoveDown;
  final VoidCallback onDelete;
  final Future<void> Function(String apiKey, String model, {String? newBaseUrl}) onSave;
  final Future<Map<String, dynamic>> Function(
      String providerName, String apiKey, String model) onTest;

  const _ProviderCard({
    super.key,
    required this.provider,
    required this.index,
    required this.total,
    required this.meta,
    required this.onToggleActive,
    required this.onMoveUp,
    required this.onMoveDown,
    required this.onDelete,
    required this.onSave,
    required this.onTest,
  });

  @override
  State<_ProviderCard> createState() => _ProviderCardState();
}

class _ProviderCardState extends State<_ProviderCard>
    with global.ThemeRefreshMixin {
  Timer? _cooldownTimer;
  Duration _cooldownRemaining = Duration.zero;

  @override
  void initState() {
    super.initState();
    _startCooldownTimer();
  }

  @override
  void didUpdateWidget(_ProviderCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.provider.cooldownUntil != widget.provider.cooldownUntil) {
      _cooldownTimer?.cancel();
      _startCooldownTimer();
    }
  }

  void _startCooldownTimer() {
    if (widget.provider.cooldownUntil == null) {
      _cooldownRemaining = Duration.zero;
      return;
    }
    final until = DateTime.tryParse(widget.provider.cooldownUntil!);
    if (until == null) {
      _cooldownRemaining = Duration.zero;
      return;
    }
    final remaining = until.difference(DateTime.now());
    if (remaining.isNegative) {
      _cooldownRemaining = Duration.zero;
      return;
    }
    _cooldownRemaining = remaining;
    _cooldownTimer = Timer.periodic(const Duration(seconds: 1), (t) {
      final r = until.difference(DateTime.now());
      if (!mounted) {
        t.cancel();
        return;
      }
      setState(() {
        _cooldownRemaining = r.isNegative ? Duration.zero : r;
        if (_cooldownRemaining == Duration.zero) t.cancel();
      });
    });
  }

  @override
  void dispose() {
    _cooldownTimer?.cancel();
    super.dispose();
  }

  bool get _isInCooldown => _cooldownRemaining > Duration.zero;

  String get _cooldownText {
    final s = _cooldownRemaining.inSeconds;
    if (s >= 3600) {
      final h = s ~/ 3600;
      final m = (s % 3600) ~/ 60;
      return '${h}h ${m}m';
    } else if (s >= 60) {
      final m = s ~/ 60;
      final sec = s % 60;
      return '${m}m ${sec}s';
    }
    return '${s}s';
  }

  Widget _buildStatusBadge() {
    if (!widget.provider.isActive) {
      return _badge(Icons.circle, 'ปิดใช้งาน', Colors.grey);
    }
    if (_isInCooldown) {
      return _badge(Icons.timer, 'Cooldown $_cooldownText', Colors.red);
    }
    return _badge(Icons.circle, 'พร้อมใช้', Colors.green);
  }

  Widget _badge(IconData icon, String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.4)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 10, color: color),
          const SizedBox(width: 4),
          Text(label,
              style: TextStyle(
                  fontSize: 11, color: color, fontWeight: FontWeight.bold)),
        ],
      ),
    );
  }

  void _openEditDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => _EditProviderDialog(
        provider: widget.provider,
        meta: widget.meta,
        onSave: widget.onSave,
        onTest: widget.onTest,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final meta = widget.meta;
    final p = widget.provider;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      color: global.theme.cardColor,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // ---- Header ----
            Row(
              children: [
                // Priority badge
                Container(
                  width: 28,
                  height: 28,
                  decoration: BoxDecoration(
                    color: meta.color.withValues(alpha: 0.15),
                    shape: BoxShape.circle,
                  ),
                  alignment: Alignment.center,
                  child: Text(
                    '${widget.index + 1}',
                    style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: meta.color,
                        fontSize: 13),
                  ),
                ),
                const SizedBox(width: 10),
                // Provider icon + name
                Icon(meta.icon, color: meta.color, size: 22),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Text(meta.name,
                              style: const TextStyle(
                                  fontSize: 15, fontWeight: FontWeight.bold)),
                          if (p.isCustom) ...[
                            const SizedBox(width: 6),
                            Text('(${p.providerName})',
                                style: TextStyle(
                                    fontSize: 12,
                                    color: global.theme.textSecondaryColor,
                                    fontFamily: 'monospace')),
                          ],
                        ],
                      ),
                      Text(meta.description,
                          style: TextStyle(
                              fontSize: 12,
                              color: global.theme.textSecondaryColor)),
                      if (p.capabilities.isNotEmpty) ...[
                        const SizedBox(height: 4),
                        Wrap(
                          spacing: 4,
                          children: p.capabilities.map((cap) {
                            final info = _capabilityInfo(cap);
                            return Container(
                              padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
                              decoration: BoxDecoration(
                                color: info.$3.withValues(alpha: 0.12),
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(info.$2, size: 10, color: info.$3),
                                  const SizedBox(width: 3),
                                  Text(info.$1, style: TextStyle(fontSize: 10, color: info.$3)),
                                ],
                              ),
                            );
                          }).toList(),
                        ),
                      ],
                    ],
                  ),
                ),
                // Status badge
                _buildStatusBadge(),
                const SizedBox(width: 8),
                // Active toggle
                Switch(
                  value: p.isActive,
                  activeThumbColor: meta.color,
                  onChanged: (_) => widget.onToggleActive(),
                ),
              ],
            ),

            const SizedBox(height: 8),

            // ---- Read-only info ----
            Row(
              children: [
                Text('Model: ', style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
                Expanded(
                  child: Text(p.model, style: const TextStyle(fontSize: 13, fontFamily: 'monospace', fontWeight: FontWeight.w500)),
                ),
              ],
            ),

            // ---- Last error ----
            if (p.lastError != null && p.lastError!.isNotEmpty) ...[
              const SizedBox(height: 6),
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(Icons.error_outline,
                      size: 14,
                      color: global.theme.negativeHighlightTextColor),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(
                      p.lastError!,
                      style: TextStyle(
                          fontSize: 11,
                          color: global.theme.negativeHighlightTextColor),
                    ),
                  ),
                ],
              ),
            ],

            const SizedBox(height: 8),

            // ---- Actions (read-only mode) ----
            Row(
              children: [
                // Up / Down priority
                IconButton(
                  icon: const Icon(Icons.arrow_upward, size: 18),
                  tooltip: 'เพิ่มลำดับความสำคัญ',
                  onPressed: widget.onMoveUp,
                  color: widget.onMoveUp != null
                      ? global.theme.infoHighlightTextColor
                      : global.theme.iconSecondaryColor,
                  visualDensity: VisualDensity.compact,
                ),
                IconButton(
                  icon: const Icon(Icons.arrow_downward, size: 18),
                  tooltip: 'ลดลำดับความสำคัญ',
                  onPressed: widget.onMoveDown,
                  color: widget.onMoveDown != null
                      ? global.theme.infoHighlightTextColor
                      : global.theme.iconSecondaryColor,
                  visualDensity: VisualDensity.compact,
                ),
                const Spacer(),
                // Edit button
                TextButton.icon(
                  onPressed: () => _openEditDialog(context),
                  icon: Icon(Icons.edit_outlined,
                      size: 16,
                      color: global.theme.infoHighlightTextColor),
                  label: Text(
                    'แก้ไข',
                    style:
                        TextStyle(color: global.theme.infoHighlightTextColor),
                  ),
                ),
                const SizedBox(width: 4),
                // Delete button
                TextButton.icon(
                  onPressed: widget.onDelete,
                  icon: Icon(Icons.delete_outline,
                      size: 16,
                      color: global.theme.negativeHighlightTextColor),
                  label: Text(
                    'ลบ',
                    style: TextStyle(
                        color: global.theme.negativeHighlightTextColor),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Edit Provider Dialog
// ---------------------------------------------------------------------------

class _EditProviderDialog extends StatefulWidget {
  final _AIProvider provider;
  final _ProviderMeta meta;
  final Future<void> Function(String apiKey, String model, {String? newBaseUrl}) onSave;
  final Future<Map<String, dynamic>> Function(String providerName, String apiKey, String model) onTest;

  const _EditProviderDialog({
    required this.provider,
    required this.meta,
    required this.onSave,
    required this.onTest,
  });

  @override
  State<_EditProviderDialog> createState() => _EditProviderDialogState();
}

class _EditProviderDialogState extends State<_EditProviderDialog> {
  late TextEditingController _apiKeyCtrl;
  late TextEditingController _modelCtrl;
  late TextEditingController _baseUrlCtrl;
  bool _obscureApiKey = true;
  bool _isTesting = false;
  bool _isSaving = false;
  String? _testMessage;
  bool _testPassed = false;

  @override
  void initState() {
    super.initState();
    _apiKeyCtrl = TextEditingController(text: widget.provider.apiKey);
    _modelCtrl = TextEditingController(text: widget.provider.model);
    _baseUrlCtrl = TextEditingController(text: widget.provider.baseUrl);
  }

  @override
  void dispose() {
    _apiKeyCtrl.dispose();
    _modelCtrl.dispose();
    _baseUrlCtrl.dispose();
    super.dispose();
  }

  Future<void> _doTest() async {
    setState(() { _isTesting = true; _testMessage = null; });
    try {
      final result = await widget.onTest(
        widget.provider.providerName,
        _apiKeyCtrl.text.trim(),
        _modelCtrl.text.trim(),
      );
      if (!mounted) return;
      final success = result['success'] as bool? ?? false;
      final message = result['message'] as String? ?? '';
      final ms = result['response_time_ms'] as int?;
      setState(() {
        _testPassed = success;
        _testMessage = success
            ? 'ผ่าน! ${ms != null ? "(${ms}ms)" : ""}'
            : 'ไม่ผ่าน: $message';
      });
    } catch (e) {
      if (mounted) setState(() { _testPassed = false; _testMessage = 'Error: $e'; });
    } finally {
      if (mounted) setState(() => _isTesting = false);
    }
  }

  Future<void> _doSave() async {
    setState(() => _isSaving = true);
    try {
      final isCustomOrOllama = widget.provider.isCustom || widget.provider.providerName == 'ollama';
      await widget.onSave(
        _apiKeyCtrl.text.trim(),
        _modelCtrl.text.trim(),
        newBaseUrl: isCustomOrOllama ? _baseUrlCtrl.text.trim() : null,
      );
      if (mounted) Navigator.pop(context);
    } finally {
      if (mounted) setState(() => _isSaving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final p = widget.provider;
    final meta = widget.meta;
    final isCustomOrOllama = p.isCustom || p.providerName == 'ollama';

    return AlertDialog(
      title: Row(
        children: [
          Icon(meta.icon, color: meta.color, size: 22),
          const SizedBox(width: 8),
          Text('แก้ไข ${meta.name}'),
        ],
      ),
      content: SizedBox(
        width: 400,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Base URL (custom/ollama)
              if (isCustomOrOllama) ...[
                TextFormField(
                  controller: _baseUrlCtrl,
                  style: const TextStyle(fontSize: 13, fontFamily: 'monospace'),
                  decoration: InputDecoration(
                    labelText: p.providerName == 'ollama' ? 'Ollama URL (ว่าง = เครื่องนี้)' : 'Base URL',
                    border: const OutlineInputBorder(),
                    isDense: true,
                  ),
                ),
                const SizedBox(height: 12),
              ],

              // API Key
              TextFormField(
                controller: _apiKeyCtrl,
                obscureText: _obscureApiKey,
                style: const TextStyle(fontSize: 13, fontFamily: 'monospace'),
                decoration: InputDecoration(
                  labelText: p.providerName == 'ollama' ? 'API Key (ไม่ต้องกรอก)' : 'API Key',
                  border: const OutlineInputBorder(),
                  isDense: true,
                  suffixIcon: IconButton(
                    icon: Icon(_obscureApiKey ? Icons.visibility : Icons.visibility_off, size: 18),
                    onPressed: () => setState(() => _obscureApiKey = !_obscureApiKey),
                  ),
                ),
              ),
              const SizedBox(height: 12),

              // Model
              TextFormField(
                controller: _modelCtrl,
                style: const TextStyle(fontSize: 13),
                decoration: InputDecoration(
                  labelText: 'Model',
                  hintText: meta.defaultModel,
                  border: const OutlineInputBorder(),
                  isDense: true,
                ),
              ),
              const SizedBox(height: 12),

              // Test button
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: _isTesting ? null : _doTest,
                  icon: _isTesting
                      ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                      : Icon(_testPassed ? Icons.check_circle : Icons.play_circle_outline, size: 18,
                          color: _testPassed ? Colors.green : null),
                  label: Text(_isTesting ? 'กำลังทดสอบ...' : 'ทดสอบการเชื่อมต่อ'),
                ),
              ),

              // Test result
              if (_testMessage != null) ...[
                const SizedBox(height: 6),
                Row(
                  children: [
                    Icon(_testPassed ? Icons.check_circle : Icons.error, size: 14,
                        color: _testPassed ? Colors.green : Colors.red),
                    const SizedBox(width: 6),
                    Expanded(
                      child: Text(_testMessage!, style: TextStyle(fontSize: 12,
                          color: _testPassed ? Colors.green : Colors.red)),
                    ),
                  ],
                ),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context), child: const Text('ยกเลิก')),
        FilledButton.icon(
          onPressed: _isSaving ? null : _doSave,
          icon: _isSaving
              ? const SizedBox(width: 14, height: 14, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
              : const Icon(Icons.save, size: 18),
          label: const Text('บันทึก'),
        ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Add Provider Dialog
// ---------------------------------------------------------------------------

class _AddProviderDialog extends StatefulWidget {
  final List<MapEntry<String, _ProviderMeta>> availableProviders;
  final int nextPriority;

  const _AddProviderDialog({
    required this.availableProviders,
    required this.nextPriority,
  });

  @override
  State<_AddProviderDialog> createState() => _AddProviderDialogState();
}

class _AddProviderDialogState extends State<_AddProviderDialog>
    with global.ThemeRefreshMixin {
  late String _selectedKey;
  late TextEditingController _apiKeyCtrl;
  late TextEditingController _modelCtrl;
  late TextEditingController _baseUrlCtrl;
  late TextEditingController _customNameCtrl;
  bool _obscureApiKey = true;
  bool _isTesting = false;
  bool _testPassed = false;
  String? _testMessage;
  bool _isLoadingModels = false;
  List<Map<String, String>> _availableModels = [];
  Set<String> _capabilities = {};

  bool get _isCustom => _selectedKey == 'custom';
  bool get _isOllama => _selectedKey == 'ollama';

  @override
  void initState() {
    super.initState();
    _selectedKey = widget.availableProviders.first.key;
    _apiKeyCtrl = TextEditingController();
    _modelCtrl = TextEditingController(
        text: widget.availableProviders.first.value.defaultModel);
    _baseUrlCtrl = TextEditingController();
    _customNameCtrl = TextEditingController();
  }

  @override
  void dispose() {
    _apiKeyCtrl.dispose();
    _modelCtrl.dispose();
    _baseUrlCtrl.dispose();
    _customNameCtrl.dispose();
    super.dispose();
  }

  void _onProviderChanged(String? key) {
    if (key == null) return;
    setState(() {
      _selectedKey = key;
      _modelCtrl.text = _kProviderMeta[key]?.defaultModel ?? '';
      _testPassed = false;
      _testMessage = null;
      _availableModels = [];
      _capabilities = {};
    });
  }

  String get _providerName {
    if (!_isCustom) return _selectedKey;
    final suffix = _customNameCtrl.text.trim();
    if (suffix.isEmpty) return 'custom';
    return 'custom-$suffix';
  }

  Future<void> _doTest() async {
    final apiKey = _apiKeyCtrl.text.trim();
    // Custom provider ไม่บังคับ API Key (เช่น Ollama, local proxy)
    if (!_isCustom && !_isOllama && apiKey.isEmpty) {
      global.showErrorSnackBar(context, 'กรุณากรอก API Key');
      return;
    }
    final model = _modelCtrl.text.trim();
    if (model.isEmpty) {
      global.showErrorSnackBar(context, 'กรุณากรอก Model');
      return;
    }
    if (_isCustom && _baseUrlCtrl.text.trim().isEmpty) {
      global.showErrorSnackBar(context, 'กรุณากรอก Base URL');
      return;
    }

    setState(() {
      _isTesting = true;
      _testMessage = null;
    });

    try {
      final url = Uri.parse(global.goApiUrlPath('api/v1/ai-provider/test'));
      final body = <String, dynamic>{
        'shop_id': global.getShopId(),
        'provider_name': _providerName,
        'api_key': apiKey,
        'model': model,
      };
      if (_isCustom || _isOllama) {
        final bUrl = _baseUrlCtrl.text.trim();
        if (bUrl.isNotEmpty) body['base_url'] = bUrl;
      }

      final response = await http.post(
        url,
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode(body),
      ).timeout(const Duration(seconds: 120));

      final result = jsonDecode(response.body) as Map<String, dynamic>;
      final success = result['success'] as bool? ?? false;
      final message = result['message'] as String? ?? '';
      final ms = result['response_time_ms'] as int?;

      if (!mounted) return;

      // Auto-detect capabilities from test response
      if (success && result['capabilities'] is List) {
        _capabilities = (result['capabilities'] as List).map((e) => e.toString()).toSet();
      }

      setState(() {
        _testPassed = success;
        _testMessage = success
            ? 'ผ่าน! ${ms != null ? "(${ms}ms)" : ""}${_capabilities.isNotEmpty ? " [${_capabilities.join(", ")}]" : ""}'
            : 'ไม่ผ่าน: $message';
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _testPassed = false;
        _testMessage = 'Error: $e';
      });
    } finally {
      if (mounted) setState(() => _isTesting = false);
    }
  }

  Future<void> _doLoadModels() async {
    final apiKey = _apiKeyCtrl.text.trim();
    if (!_isCustom && !_isOllama && apiKey.isEmpty) {
      global.showErrorSnackBar(context, 'กรุณากรอก API Key ก่อน');
      return;
    }
    if (_isCustom && _baseUrlCtrl.text.trim().isEmpty) {
      global.showErrorSnackBar(context, 'กรุณากรอก Base URL ก่อน');
      return;
    }

    setState(() => _isLoadingModels = true);

    try {
      final url = Uri.parse(global.goApiUrlPath('api/v1/ai-provider/models'));
      final body = <String, dynamic>{
        'provider_name': _providerName,
        'api_key': apiKey,
      };
      if (_isCustom || _isOllama) {
        final bUrl = _baseUrlCtrl.text.trim();
        if (bUrl.isNotEmpty) body['base_url'] = bUrl;
      }

      final response = await http.post(
        url,
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode(body),
      ).timeout(const Duration(seconds: 15));

      final result = jsonDecode(response.body) as Map<String, dynamic>;
      final success = result['success'] as bool? ?? false;

      if (!mounted) return;

      if (!success) {
        global.showErrorSnackBar(
            context, result['message'] as String? ?? 'ดึง models ไม่ได้');
        setState(() => _isLoadingModels = false);
        return;
      }

      final models = (result['models'] as List<dynamic>? ?? [])
          .map((m) => Map<String, String>.from(m as Map))
          .toList();

      setState(() {
        _availableModels = models;
        _isLoadingModels = false;
      });

      if (models.isEmpty) {
        global.showInfoSnackBar(context, 'ไม่พบ models');
      }
    } catch (e) {
      if (mounted) {
        global.showErrorSnackBar(context, 'Error: $e');
        setState(() => _isLoadingModels = false);
      }
    }
  }

  void _onSubmit() {
    if (!_testPassed) {
      global.showErrorSnackBar(context, 'กรุณาทดสอบให้ผ่านก่อนบันทึก');
      return;
    }
    final apiKey = _apiKeyCtrl.text.trim();
    final model = _modelCtrl.text.trim();

    final data = <String, dynamic>{
      'provider_name': _providerName,
      'api_key': apiKey,
      'model': model,
      'capabilities': _capabilities.toList(),
      'is_active': true,
      'priority': widget.nextPriority,
    };
    if (_isCustom || _isOllama) {
      final url = _baseUrlCtrl.text.trim();
      if (url.isNotEmpty) data['base_url'] = url;
    }

    Navigator.pop(context, data);
  }

  @override
  Widget build(BuildContext context) {
    final meta = _kProviderMeta[_selectedKey]!;

    return AlertDialog(
      title: const Text('เพิ่ม AI Provider'),
      content: SizedBox(
        width: 460,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Provider selector
              DropdownButtonFormField<String>(
                initialValue: _selectedKey,
                decoration: const InputDecoration(
                  labelText: 'Provider',
                  border: OutlineInputBorder(),
                ),
                isExpanded: true,
                items: widget.availableProviders.map((e) {
                  final m = e.value;
                  return DropdownMenuItem<String>(
                    value: e.key,
                    child: Row(
                      children: [
                        Icon(m.icon, color: m.color, size: 18),
                        const SizedBox(width: 8),
                        Text(m.name),
                        const SizedBox(width: 8),
                        Flexible(
                          child: Text(m.description,
                              overflow: TextOverflow.ellipsis,
                              style: TextStyle(
                                  fontSize: 11,
                                  color: global.theme.textSecondaryColor)),
                        ),
                      ],
                    ),
                  );
                }).toList(),
                onChanged: _onProviderChanged,
              ),
              const SizedBox(height: 16),

              // Custom name (for custom providers)
              if (_isCustom) ...[
                TextFormField(
                  controller: _customNameCtrl,
                  decoration: const InputDecoration(
                    labelText: 'ชื่อ Provider (ภาษาอังกฤษ)',
                    hintText: 'เช่น localai, proxy, ollama',
                    border: OutlineInputBorder(),
                    isDense: true,
                  ),
                ),
                const SizedBox(height: 12),
              ],

              // Base URL (for custom + ollama)
              if (_isCustom || _isOllama) ...[
                TextFormField(
                  controller: _baseUrlCtrl,
                  style:
                      const TextStyle(fontFamily: 'monospace', fontSize: 13),
                  decoration: InputDecoration(
                    labelText: _isOllama ? 'Ollama URL (ว่าง = เครื่องนี้)' : 'Base URL',
                    hintText: _isOllama
                        ? 'http://192.168.1.100:11434/v1'
                        : 'http://host.docker.internal:3333/v1',
                    border: const OutlineInputBorder(),
                    isDense: true,
                  ),
                ),
                const SizedBox(height: 12),
              ],

              // API Key
              TextFormField(
                controller: _apiKeyCtrl,
                obscureText: _obscureApiKey,
                style: const TextStyle(fontFamily: 'monospace', fontSize: 13),
                decoration: InputDecoration(
                  labelText: _isOllama ? 'API Key (ไม่ต้องกรอก)' : 'API Key',
                  border: const OutlineInputBorder(),
                  isDense: true,
                  suffixIcon: IconButton(
                    icon: Icon(_obscureApiKey
                        ? Icons.visibility
                        : Icons.visibility_off,
                        size: 18),
                    onPressed: () =>
                        setState(() => _obscureApiKey = !_obscureApiKey),
                  ),
                ),
              ),
              const SizedBox(height: 12),

              // Model + search button
              Row(
                children: [
                  Expanded(
                    child: TextFormField(
                      controller: _modelCtrl,
                      decoration: InputDecoration(
                        labelText: 'Model',
                        hintText: meta.defaultModel,
                        border: const OutlineInputBorder(),
                        isDense: true,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  SizedBox(
                    height: 40,
                    child: OutlinedButton.icon(
                      onPressed: _isLoadingModels ? null : _doLoadModels,
                      icon: _isLoadingModels
                          ? const SizedBox(
                              width: 14,
                              height: 14,
                              child:
                                  CircularProgressIndicator(strokeWidth: 2))
                          : const Icon(Icons.search, size: 16),
                      label: const Text('ค้นหา', style: TextStyle(fontSize: 12)),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                'Default: ${meta.defaultModel}',
                style: TextStyle(
                    fontSize: 11, color: global.theme.textSecondaryColor),
              ),

              // Model list (from search)
              if (_availableModels.isNotEmpty) ...[
                const SizedBox(height: 8),
                Container(
                  constraints: const BoxConstraints(maxHeight: 200),
                  decoration: BoxDecoration(
                    border: Border.all(color: Colors.grey.shade300),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: ListView.builder(
                    shrinkWrap: true,
                    itemCount: _availableModels.length,
                    itemBuilder: (ctx, i) {
                      final m = _availableModels[i];
                      final id = m['id'] ?? '';
                      final owner = m['owned_by'] ?? '';
                      final isSelected = _modelCtrl.text == id;
                      return ListTile(
                        dense: true,
                        selected: isSelected,
                        selectedTileColor:
                            meta.color.withValues(alpha: 0.08),
                        leading: Icon(
                          isSelected
                              ? Icons.check_circle
                              : Icons.circle_outlined,
                          size: 16,
                          color: isSelected ? meta.color : Colors.grey,
                        ),
                        title: Text(id,
                            style: const TextStyle(
                                fontSize: 12, fontFamily: 'monospace')),
                        subtitle: owner.isNotEmpty
                            ? Text(owner,
                                style: const TextStyle(fontSize: 10))
                            : null,
                        onTap: () {
                          setState(() {
                            _modelCtrl.text = id;
                            _testPassed = false;
                          });
                        },
                      );
                    },
                  ),
                ),
              ],

              const SizedBox(height: 16),

              // Test button
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: _isTesting ? null : _doTest,
                  icon: _isTesting
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2))
                      : Icon(
                          _testPassed
                              ? Icons.check_circle
                              : Icons.play_circle_outline,
                          size: 18,
                          color: _testPassed ? Colors.green : null),
                  label: Text(_isTesting ? 'กำลังทดสอบ...' : 'ทดสอบการเชื่อมต่อ'),
                  style: OutlinedButton.styleFrom(
                    side: BorderSide(
                        color: _testPassed ? Colors.green : Colors.grey),
                  ),
                ),
              ),

              // Test result message
              if (_testMessage != null) ...[
                const SizedBox(height: 6),
                Row(
                  children: [
                    Icon(
                      _testPassed ? Icons.check_circle : Icons.error,
                      size: 14,
                      color: _testPassed ? Colors.green : Colors.red,
                    ),
                    const SizedBox(width: 6),
                    Expanded(
                      child: Text(
                        _testMessage!,
                        style: TextStyle(
                          fontSize: 12,
                          color: _testPassed ? Colors.green : Colors.red,
                        ),
                      ),
                    ),
                  ],
                ),
              ],

              // Capability checkboxes — แสดงเสมอ, auto-detect เมื่อทดสอบผ่าน
              const SizedBox(height: 10),
              Text('ความสามารถ:${_testPassed ? " (ตรวจอัตโนมัติแล้ว)" : ""}',
                  style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
              Wrap(
                spacing: 0,
                children: ['tools', 'vision', 'thinking'].map((cap) {
                  final info = _capabilityInfo(cap);
                  return SizedBox(
                    width: 130,
                    child: CheckboxListTile(
                      dense: true,
                      contentPadding: EdgeInsets.zero,
                      visualDensity: VisualDensity.compact,
                      controlAffinity: ListTileControlAffinity.leading,
                      value: _capabilities.contains(cap),
                      onChanged: (v) => setState(() {
                        if (v == true) { _capabilities.add(cap); } else { _capabilities.remove(cap); }
                      }),
                      title: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(info.$2, size: 14, color: info.$3),
                          const SizedBox(width: 4),
                          Text(info.$1, style: const TextStyle(fontSize: 12)),
                        ],
                      ),
                    ),
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
          child: const Text('ยกเลิก'),
        ),
        ElevatedButton.icon(
          onPressed: _testPassed ? _onSubmit : null,
          icon: const Icon(Icons.add, size: 18),
          label: const Text('เพิ่ม'),
        ),
      ],
    );
  }
}
