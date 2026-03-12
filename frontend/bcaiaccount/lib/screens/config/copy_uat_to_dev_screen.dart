import 'dart:convert';
import 'package:smlaicloud/widgets/manual_button.dart';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// หน้าจอโอนข้อมูล MongoDB จาก UAT (Production) → DEV
/// ดึง list shops จาก Production MongoDB แล้วเลือก shop ต้นทาง
/// แสดง preview จำนวน documents ทุก collection ก่อนยืนยัน copy
class CopyUatToDevScreen extends StatefulWidget {
  const CopyUatToDevScreen({super.key});

  @override
  State<CopyUatToDevScreen> createState() => _CopyUatToDevScreenState();
}

class _CopyUatToDevScreenState extends State<CopyUatToDevScreen>
    with global.ThemeRefreshMixin {
  bool _isLoadingShops = false;
  bool _isLoadingPreview = false;
  bool _isCopying = false;
  String _statusMessage = '';
  bool _isError = false;

  List<_ShopItem> _shops = [];
  List<_ShopItem> _filteredShops = [];
  _ShopItem? _selectedShop;
  final _searchController = TextEditingController();

  // Preview data
  List<_CollectionPreview> _previewCollections = [];
  int _previewTotalDocs = 0;

  @override
  void initState() {
    super.initState();
    _loadSourceShops();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  /// ดึงรายการ shops จาก Production MongoDB
  Future<void> _loadSourceShops() async {
    setState(() {
      _isLoadingShops = true;
      _statusMessage = '';
      _isError = false;
    });

    try {
      final url = global.goApiUrlPath('listsourceshops');
      AppLogger.info('[CopyUatToDev] เรียก API: $url');

      final response = await http.get(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          final List<dynamic> shopList = data['data'] ?? [];
          final shops = shopList.map((s) {
            final guidfixed = s['guidfixed'] ?? '';
            final name1 = s['name1'] ?? '';
            String displayName = name1;
            if (s['names'] != null && s['names'] is List && (s['names'] as List).isNotEmpty) {
              final names = s['names'] as List;
              displayName = names.first['name'] ?? name1;
            }
            final branchCode = s['branchcode'] ?? '';
            return _ShopItem(
              guidfixed: guidfixed,
              name: displayName,
              branchCode: branchCode,
            );
          }).toList();

          setState(() {
            _shops = shops;
            _filteredShops = shops;
            _statusMessage = '${global.language("found")} ${shops.length} ${global.language("shop_from_production")}';
          });
        } else {
          setState(() {
            _statusMessage = data['error'] ?? global.language('cannot_load_shop_data');
            _isError = true;
          });
        }
      } else {
        setState(() {
          _statusMessage = 'HTTP Error: ${response.statusCode} - ${response.reasonPhrase}';
          _isError = true;
        });
      }
    } catch (e) {
      AppLogger.error('[CopyUatToDev] โหลดรายการร้านค้า error: $e');
      setState(() {
        _statusMessage = '${global.language("error_occurred")}: $e';
        _isError = true;
      });
    } finally {
      if (mounted) setState(() => _isLoadingShops = false);
    }
  }

  /// ดึง preview — จำนวน documents แต่ละ collection ที่จะโอน
  Future<void> _loadPreview() async {
    if (_selectedShop == null) return;

    setState(() {
      _isLoadingPreview = true;
      _previewCollections = [];
      _previewTotalDocs = 0;
      _statusMessage = global.language('counting_data_to_transfer');
      _isError = false;
    });

    try {
      final url = global.goApiUrlPath('previewcopymongo');
      AppLogger.info('[CopyUatToDev] Preview: $url, shop=${_selectedShop!.guidfixed}');

      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'source_shop_id': _selectedShop!.guidfixed}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          final List<dynamic> collections = data['collections'] ?? [];
          final previews = collections.map((c) {
            return _CollectionPreview(
              name: c['name'] ?? '',
              count: (c['count'] ?? 0) as int,
            );
          }).toList();

          // เรียงจากมากไปน้อย
          previews.sort((a, b) => b.count.compareTo(a.count));

          setState(() {
            _previewCollections = previews;
            _previewTotalDocs = (data['total_documents'] ?? 0) as int;
            _statusMessage = '${global.language("found")} ${previews.length} collections, ${global.language("total")} $_previewTotalDocs ${global.language("items")}';
          });
        } else {
          setState(() {
            _statusMessage = data['error'] ?? global.language('cannot_load_preview');
            _isError = true;
          });
        }
      } else {
        setState(() {
          _statusMessage = 'HTTP Error: ${response.statusCode}';
          _isError = true;
        });
      }
    } catch (e) {
      AppLogger.error('[CopyUatToDev] Preview error: $e');
      setState(() {
        _statusMessage = '${global.language("error_occurred")}: $e';
        _isError = true;
      });
    } finally {
      if (mounted) setState(() => _isLoadingPreview = false);
    }
  }

  /// เริ่ม copy ข้อมูลจาก source shop → target shop (shop ปัจจุบัน)
  Future<void> _startCopy() async {
    if (_selectedShop == null) return;

    final targetShopId = global.appConfig.getString('shopid') ?? '';
    if (targetShopId.isEmpty) {
      setState(() {
        _statusMessage = global.language('target_shop_id_not_found');
        _isError = true;
      });
      return;
    }

    final confirmed = await _showConfirmDialog(targetShopId);
    if (confirmed != true) return;

    setState(() {
      _isCopying = true;
      _statusMessage = global.language('starting_data_transfer');
      _isError = false;
    });

    try {
      final url = global.goApiUrlPath('copymongouattodev');
      AppLogger.info('[CopyUatToDev] เริ่ม copy: source=${_selectedShop!.guidfixed}, target=$targetShopId');

      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'source_shop_id': _selectedShop!.guidfixed,
          'target_shop_id': targetShopId,
        }),
      );

      if (response.statusCode == 202 || response.statusCode == 200) {
        setState(() {
          _statusMessage = '${global.language("data_transfer_started")}\n'
              '${global.language("transfer_mongodb_only")}';
          _isError = false;
        });
      } else {
        final data = jsonDecode(response.body);
        setState(() {
          _statusMessage = '${global.language("error_occurred")}: ${data['error'] ?? response.reasonPhrase}';
          _isError = true;
        });
      }
    } catch (e) {
      AppLogger.error('[CopyUatToDev] Copy error: $e');
      setState(() {
        _statusMessage = '${global.language("error_occurred")}: $e';
        _isError = true;
      });
    } finally {
      if (mounted) setState(() => _isCopying = false);
    }
  }

  Future<bool?> _showConfirmDialog(String targetShopId) {
    return showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            Icon(Icons.warning_amber, color: global.theme.warningHighlightTextColor, size: 28),
            SizedBox(width: 8),
            Text(global.language('confirm_data_transfer')),
          ],
        ),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language('transfer_warning_message'),
                style: TextStyle(color: global.theme.negativeHighlightTextColor, fontWeight: FontWeight.bold),
              ),
              SizedBox(height: 16),
              _infoRow(global.language('source_production'), _selectedShop!.name),
              _infoRow('${global.language("shop")} ID ${global.language("source")}', _selectedShop!.guidfixed),
              Divider(),
              _infoRow(global.language('destination_dev'), global.language('current_shop')),
              _infoRow('${global.language("shop")} ID ${global.language("destination")}', targetShopId),
              Divider(),
              _infoRow(global.language('number_of_collections'), '${_previewCollections.length}'),
              _infoRow(global.language('total_documents'), _formatNumber(_previewTotalDocs)),
              SizedBox(height: 8),
              Text(
                global.language('transfer_mongodb_only_note'),
                style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.warningHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            child: Text(global.language('confirm_transfer')),
          ),
        ],
      ),
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 160,
            child: Text(label, style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
          ),
          Expanded(
            child: Text(value, style: TextStyle(fontSize: 13)),
          ),
        ],
      ),
    );
  }

  String _formatNumber(int number) {
    if (number >= 1000000) {
      return '${(number / 1000000).toStringAsFixed(1)}M';
    } else if (number >= 1000) {
      return '${(number / 1000).toStringAsFixed(1)}K';
    }
    return number.toString();
  }

  void _filterShops(String query) {
    setState(() {
      if (query.isEmpty) {
        _filteredShops = _shops;
      } else {
        final lower = query.toLowerCase();
        _filteredShops = _shops
            .where((s) =>
                s.name.toLowerCase().contains(lower) ||
                s.guidfixed.toLowerCase().contains(lower) ||
                s.branchCode.toLowerCase().contains(lower))
            .toList();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(global.language('transfer_mongodb_uat_to_dev')),
        backgroundColor: global.theme.appBarColor,
        foregroundColor: global.theme.onPrimaryColor,
        actions: [
          IconButton(
            icon: Icon(Icons.refresh),
            onPressed: _isLoadingShops ? null : _loadSourceShops,
            tooltip: global.language('reload_shop_list'),
          ),
          const ManualButton(path: 'settings-data-transfer'),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 900),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Status message
              if (_statusMessage.isNotEmpty)
                Container(
                  padding: EdgeInsets.all(12),
                  margin: EdgeInsets.only(bottom: 16),
                  decoration: BoxDecoration(
                    color: _isError ? global.theme.negativeHighlightColor : global.theme.positiveHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(
                      color: _isError ? global.theme.negativeHighlightColor : global.theme.positiveHighlightColor,
                    ),
                  ),
                  child: Row(
                    children: [
                      Icon(
                        _isError ? Icons.error_outline : Icons.check_circle_outline,
                        color: _isError ? global.theme.negativeHighlightTextColor : global.theme.positiveHighlightTextColor,
                        size: 20,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _statusMessage,
                          style: TextStyle(
                            color: _isError ? global.theme.negativeHighlightTextColor : global.theme.positiveHighlightTextColor,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),

              // คำอธิบาย
              Card(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Icon(Icons.info_outline, color: global.theme.primaryColor),
                          SizedBox(width: 8),
                          Text(
                            global.language('transfer_mongodb_production_to_dev'),
                            style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                          ),
                        ],
                      ),
                      SizedBox(height: 8),
                      Text(
                        '${global.language("transfer_step_1")}\n'
                        '${global.language("transfer_step_2")}\n'
                        '${global.language("transfer_step_3")}',
                        style: TextStyle(fontSize: 13, color: global.theme.textSecondaryColor),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),

              // เลือกร้านค้าต้นทาง
              Card(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Text(
                        global.language('select_source_shop'),
                        style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: global.theme.textColor),
                      ),
                      const SizedBox(height: 12),

                      if (_isLoadingShops)
                        Center(
                          child: Padding(
                            padding: EdgeInsets.all(24),
                            child: Column(
                              children: [
                                const CircularProgressIndicator(),
                                SizedBox(height: 12),
                                Text(global.language('loading_shop_list')),
                              ],
                            ),
                          ),
                        )
                      else if (_shops.isEmpty)
                        Center(
                          child: Padding(
                            padding: EdgeInsets.all(24),
                            child: Column(
                              children: [
                                Icon(Icons.store_mall_directory, size: 48, color: global.theme.iconSecondaryColor),
                                SizedBox(height: 8),
                                Text(global.language('shop_not_found')),
                                SizedBox(height: 8),
                                ElevatedButton.icon(
                                  onPressed: _loadSourceShops,
                                  icon: Icon(Icons.refresh, size: 18),
                                  label: Text(global.language('retry')),
                                ),
                              ],
                            ),
                          ),
                        )
                      else
                        _buildShopList(),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),

              // Preview + ปุ่มโอน
              if (_selectedShop != null) ...[
                // ปุ่ม preview
                if (_previewCollections.isEmpty && !_isLoadingPreview)
                  ElevatedButton.icon(
                    onPressed: _loadPreview,
                    icon: Icon(Icons.preview),
                    label: Text('${global.language("view_data_details")} "${_selectedShop!.name}"'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.primaryColor,
                      foregroundColor: global.theme.onPrimaryColor,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                    ),
                  ),

                if (_isLoadingPreview)
                  Center(
                    child: Padding(
                      padding: EdgeInsets.all(24),
                      child: Column(
                        children: [
                          const CircularProgressIndicator(),
                          SizedBox(height: 12),
                          Text(global.language('counting_data')),
                        ],
                      ),
                    ),
                  ),

                // แสดง preview collections
                if (_previewCollections.isNotEmpty) ...[
                  _buildPreviewCard(),
                  const SizedBox(height: 16),

                  // ปุ่มโอนข้อมูล
                  ElevatedButton.icon(
                    onPressed: _isCopying ? null : _startCopy,
                    icon: _isCopying
                        ? SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2, color: global.theme.cardColor),
                          )
                        : Icon(Icons.cloud_download),
                    label: Text(_isCopying
                        ? global.language('transferring_data')
                        : '${global.language("start_transfer")} (${_previewCollections.length} collections, ${_formatNumber(_previewTotalDocs)} docs)'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.warningHighlightTextColor,
                      foregroundColor: global.theme.onPrimaryColor,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      textStyle: TextStyle(fontSize: 15),
                    ),
                  ),
                ],
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildShopList() {
    return Column(
      children: [
        TextField(
          controller: _searchController,
          decoration: InputDecoration(
            hintText: global.language('search_shop'),
            prefixIcon: Icon(Icons.search),
            border: OutlineInputBorder(),
            isDense: true,
          ),
          onChanged: _filterShops,
        ),
        SizedBox(height: 8),
        Text(
          '${global.language("showing")} ${_filteredShops.length} ${global.language("from")} ${_shops.length} ${global.language("shop")}',
          style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
        ),
        const SizedBox(height: 8),
        ConstrainedBox(
          constraints: const BoxConstraints(maxHeight: 350),
          child: ListView.builder(
            shrinkWrap: true,
            itemCount: _filteredShops.length,
            itemBuilder: (ctx, index) {
              final shop = _filteredShops[index];
              final isSelected = _selectedShop?.guidfixed == shop.guidfixed;
              return Card(
                color: isSelected ? global.theme.infoHighlightColor : null,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                  side: isSelected
                      ? BorderSide(color: global.theme.primaryColor, width: 2)
                      : BorderSide.none,
                ),
                child: ListTile(
                  leading: CircleAvatar(
                    backgroundColor: isSelected ? global.theme.primaryColor : global.theme.dividerBorderColor,
                    child: Icon(
                      Icons.store,
                      color: isSelected ? global.theme.onPrimaryColor : global.theme.textSecondaryColor,
                      size: 20,
                    ),
                  ),
                  title: Text(
                    shop.name,
                    style: TextStyle(fontWeight: isSelected ? FontWeight.bold : FontWeight.normal),
                  ),
                  subtitle: Text(
                    'ID: ${shop.guidfixed}${shop.branchCode.isNotEmpty ? ' | ${global.language("branch")}: ${shop.branchCode}' : ''}',
                    style: TextStyle(fontSize: 11),
                  ),
                  trailing: isSelected
                      ? Icon(Icons.check_circle, color: global.theme.primaryColor)
                      : null,
                  onTap: () {
                    setState(() {
                      _selectedShop = shop;
                      // reset preview เมื่อเปลี่ยน shop
                      _previewCollections = [];
                      _previewTotalDocs = 0;
                    });
                  },
                ),
              );
            },
          ),
        ),
      ],
    );
  }

  Widget _buildPreviewCard() {
    return Card(
      color: global.theme.warningHighlightColor,
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Icon(Icons.list_alt, color: global.theme.warningHighlightTextColor),
                SizedBox(width: 8),
                Expanded(
                  child: Text(
                    '${global.language("transfer_data_details")} — ${_selectedShop!.name}',
                    style: TextStyle(fontSize: 15, fontWeight: FontWeight.bold),
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.refresh, size: 20),
                  onPressed: _isLoadingPreview ? null : _loadPreview,
                  tooltip: global.language('refresh'),
                ),
              ],
            ),
            const SizedBox(height: 8),

            // สรุป
            Container(
              padding: EdgeInsets.all(12),
              color: global.theme.searchBarColor,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [
                  _summaryItem('Collections', '${_previewCollections.length}', Icons.folder),
                  _summaryItem('${global.language("total")} Documents', _formatNumber(_previewTotalDocs), Icons.description),
                ],
              ),
            ),
            const SizedBox(height: 12),

            // ตาราง collections
            Container(
              color: global.theme.searchBarColor,
              child: ConstrainedBox(
                constraints: BoxConstraints(maxHeight: 400),
                child: SingleChildScrollView(
                  child: DataTable(
                    headingRowHeight: 40,
                    dataRowMinHeight: 32,
                    dataRowMaxHeight: 40,
                    columnSpacing: 24,
                    columns: [
                      const DataColumn(label: Text('#', style: TextStyle(fontWeight: FontWeight.bold))),
                      DataColumn(label: Text(global.language('collection_name'), style: TextStyle(fontWeight: FontWeight.bold))),
                      DataColumn(
                        label: Text(global.language('amount'), style: TextStyle(fontWeight: FontWeight.bold)),
                        numeric: true,
                      ),
                    ],
                    rows: _previewCollections.asMap().entries.map((entry) {
                      final idx = entry.key;
                      final c = entry.value;
                      return DataRow(cells: [
                        DataCell(Text('${idx + 1}')),
                        DataCell(Text(c.name, style: TextStyle(fontSize: 13))),
                        DataCell(Text(
                          _formatNumber(c.count),
                          style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                        )),
                      ]);
                    }).toList(),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _summaryItem(String label, String value, IconData icon) {
    return Column(
      children: [
        Icon(icon, color: global.theme.warningHighlightTextColor, size: 24),
        const SizedBox(height: 4),
        Text(value, style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
        Text(label, style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor)),
      ],
    );
  }
}

class _ShopItem {
  final String guidfixed;
  final String name;
  final String branchCode;

  _ShopItem({required this.guidfixed, required this.name, required this.branchCode});
}

class _CollectionPreview {
  final String name;
  final int count;

  _CollectionPreview({required this.name, required this.count});
}
