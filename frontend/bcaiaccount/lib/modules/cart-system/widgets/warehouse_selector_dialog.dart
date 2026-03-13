import 'package:flutter/material.dart';
import '../../../global.dart' as global;
import '../models/clickhouse_warehouse_model.dart';
import '../models/clickhouse_location_model.dart';
import '../services/clickhouse_warehouse_service.dart';
import '../../../utils/logger/app_logger.dart';

/// Dialog สำหรับเลือกคลังและที่เก็บ (ใช้ ClickHouse)
class WarehouseSelectorDialog extends StatefulWidget {
  final String shopId;
  final String? currentWarehouseId;
  final String? currentLocationId;

  const WarehouseSelectorDialog({
    super.key,
    required this.shopId,
    this.currentWarehouseId,
    this.currentLocationId,
  });

  @override
  State<WarehouseSelectorDialog> createState() =>
      _WarehouseSelectorDialogState();
}

class _WarehouseSelectorDialogState extends State<WarehouseSelectorDialog> with global.ThemeRefreshMixin {
  final ClickHouseWarehouseService _service = ClickHouseWarehouseService();
  final TextEditingController _searchController = TextEditingController();

  List<ClickHouseWarehouseModel> _warehouses = [];
  List<ClickHouseLocationModel> _locations = [];
  ClickHouseWarehouseModel? _selectedWarehouse;
  ClickHouseLocationModel? _selectedLocation;
  bool _isLoadingWarehouses = false;
  bool _isLoadingLocations = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadWarehouses();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadWarehouses() async {
    setState(() {
      _isLoadingWarehouses = true;
      _errorMessage = null;
    });

    try {
      final warehouses = await _service.searchWarehouses(
        shopId: widget.shopId,
        keyword: _searchController.text,
        limit: 100,
      );

      if (mounted) {
        setState(() {
          _warehouses = warehouses;
          _isLoadingWarehouses = false;

          // เลือกคลังปัจจุบันถ้ามี
          if (widget.currentWarehouseId != null && _warehouses.isNotEmpty) {
            try {
              _selectedWarehouse = _warehouses.firstWhere(
                (w) => w.guidFixed == widget.currentWarehouseId,
              );
              // โหลด locations ของคลังที่เลือก
              _loadLocationsForWarehouse(_selectedWarehouse!);
            } catch (e) {
              // ไม่พบคลังที่ต้องการ
              AppLogger.warning('Warehouse not found: ${widget.currentWarehouseId}');
            }
          }
        });
      }
    } catch (e) {
      AppLogger.error('Error loading warehouses', error: e);
      if (mounted) {
        setState(() {
          _errorMessage = '${global.language("error_loading_warehouse")}: $e';
          _isLoadingWarehouses = false;
        });
      }
    }
  }

  Future<void> _loadLocationsForWarehouse(ClickHouseWarehouseModel warehouse) async {
    setState(() {
      _isLoadingLocations = true;
    });

    try {
      final locations = await _service.getLocationsByWarehouse(
        shopId: widget.shopId,
        warehouseCode: warehouse.code,
        limit: 100,
      );

      if (mounted) {
        setState(() {
          _locations = locations;
          _isLoadingLocations = false;

          // เลือกที่เก็บปัจจุบันถ้ามี
          if (widget.currentLocationId != null && _locations.isNotEmpty) {
            try {
              _selectedLocation = _locations.firstWhere(
                (l) => l.guidFixed == widget.currentLocationId,
              );
            } catch (e) {
              // ไม่พบที่เก็บที่ต้องการ
              AppLogger.warning('Location not found: ${widget.currentLocationId}');
            }
          }
        });
      }
    } catch (e) {
      AppLogger.error('Error loading locations', error: e);
      if (mounted) {
        setState(() {
          _isLoadingLocations = false;
        });
      }
    }
  }

  void _onWarehouseSelected(ClickHouseWarehouseModel warehouse) {
    setState(() {
      _selectedWarehouse = warehouse;
      _selectedLocation = null; // Reset location เมื่อเปลี่ยนคลัง
      _locations = []; // Clear locations
    });

    // โหลด locations ของคลังใหม่
    _loadLocationsForWarehouse(warehouse);
  }

  void _confirmSelection() {
    if (_selectedWarehouse == null) {
      global.showWarningSnackBar(context, global.language("please_select_warehouse"));
      return;
    }

    // คืนค่าเป็น Map ที่มีข้อมูลคลังและที่เก็บ
    Navigator.pop(context, {
      'warehouse': _selectedWarehouse!,
      'location': _selectedLocation,
    });
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Container(
        constraints: BoxConstraints(maxWidth: 700, maxHeight: 800),
        child: Column(
          children: [
            // Header
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: global.theme.warningHighlightColor,
                border: Border(
                  bottom: BorderSide(color: global.theme.dividerBorderColor),
                ),
              ),
              child: Row(
                children: [
                  Icon(Icons.warehouse, color: global.theme.warningHighlightTextColor),
                  SizedBox(width: 8),
                  Text(
                    global.language("select_warehouse_and_location"),
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  const Spacer(),
                  IconButton(
                    icon: Icon(Icons.close),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
            ),

            // Search Bar
            Padding(
              padding: EdgeInsets.all(16),
              child: TextField(
                controller: _searchController,
                decoration: InputDecoration(
                  hintText: global.language("search_warehouse"),
                  prefixIcon: Icon(Icons.search),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  suffixIcon: IconButton(
                    icon: Icon(Icons.refresh),
                    onPressed: _loadWarehouses,
                  ),
                ),
                onSubmitted: (_) => _loadWarehouses(),
              ),
            ),

            // Error Message
            if (_errorMessage != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: global.theme.negativeHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: global.theme.negativeHighlightTextColor.withValues(alpha: 0.3)),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.error_outline, color: global.theme.negativeHighlightTextColor),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _errorMessage!,
                          style: TextStyle(color: global.theme.negativeHighlightTextColor),
                        ),
                      ),
                    ],
                  ),
                ),
              ),

            // Content
            Expanded(
              child: _isLoadingWarehouses
                  ? const Center(child: CircularProgressIndicator())
                  : _warehouses.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(
                                Icons.warehouse_outlined,
                                size: 64,
                                color: global.theme.iconSecondaryColor,
                              ),
                              SizedBox(height: 16),
                              Text(
                                global.language("warehouse_not_found"),
                                style: TextStyle(
                                  fontSize: 16,
                                  color: global.theme.iconSecondaryColor,
                                ),
                              ),
                            ],
                          ),
                        )
                      : Row(
                          children: [
                            // คลังสินค้า (ซ้าย)
                            Expanded(
                              flex: 1,
                              child: Container(
                                decoration: BoxDecoration(
                                  border: Border(
                                    right: BorderSide(color: global.theme.dividerBorderColor),
                                  ),
                                ),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Padding(
                                      padding: EdgeInsets.all(16),
                                      child: Text(
                                        '${global.language("warehouse")} (${_warehouses.length})',
                                        style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          fontSize: 16,
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      child: ListView.builder(
                                        itemCount: _warehouses.length,
                                        itemBuilder: (context, index) {
                                          final warehouse = _warehouses[index];
                                          final isSelected =
                                              _selectedWarehouse?.guidFixed ==
                                                  warehouse.guidFixed;
                                          return ListTile(
                                            selected: isSelected,
                                            selectedTileColor: global.theme.infoHighlightColor,
                                            leading: Icon(
                                              Icons.warehouse,
                                              color: isSelected
                                                  ? global.theme.infoHighlightTextColor
                                                  : global.theme.iconSecondaryColor,
                                            ),
                                            title: Text(warehouse.displayName),
                                            subtitle: Text('${global.language("code")}: ${warehouse.code}'),
                                            trailing: isSelected
                                                ? Icon(
                                                    Icons.check_circle,
                                                    color: global.theme.infoHighlightTextColor,
                                                  )
                                                : null,
                                            onTap: () =>
                                                _onWarehouseSelected(warehouse),
                                          );
                                        },
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),

                            // ที่เก็บ (ขวา)
                            Expanded(
                              flex: 1,
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Padding(
                                    padding: EdgeInsets.all(16),
                                    child: Text(
                                      '${global.language("location")} ${_selectedWarehouse != null ? '(${_locations.length})' : ''}',
                                      style: TextStyle(
                                        fontWeight: FontWeight.bold,
                                        fontSize: 16,
                                      ),
                                    ),
                                  ),
                                  Expanded(
                                    child: _selectedWarehouse == null
                                        ? Center(
                                            child: Text(
                                              global.language("please_select_warehouse"),
                                              style: TextStyle(
                                                color: global.theme.iconSecondaryColor,
                                              ),
                                            ),
                                          )
                                        : _isLoadingLocations
                                            ? const Center(
                                                child:
                                                    CircularProgressIndicator(),
                                              )
                                            : _locations.isEmpty
                                                ? Center(
                                                    child: Text(
                                                      global.language("no_location_in_warehouse"),
                                                      style: TextStyle(
                                                        color: global.theme.iconSecondaryColor,
                                                      ),
                                                    ),
                                                  )
                                                : ListView.builder(
                                                    itemCount: _locations.length,
                                                    itemBuilder:
                                                        (context, index) {
                                                      final location =
                                                          _locations[index];
                                                      final isSelected =
                                                          _selectedLocation
                                                                  ?.guidFixed ==
                                                              location.guidFixed;
                                                      return ListTile(
                                                        selected: isSelected,
                                                        selectedTileColor:
                                                            global.theme.positiveHighlightColor,
                                                        leading: Icon(
                                                          Icons.place,
                                                          color: isSelected
                                                              ? global.theme.positiveHighlightTextColor
                                                              : global.theme.iconSecondaryColor,
                                                        ),
                                                        title: Text(
                                                            location.displayName),
                                                        subtitle: Text(
                                                            '${global.language("code")}: ${location.locationCode}'),
                                                        trailing: isSelected
                                                            ? Icon(
                                                                Icons
                                                                    .check_circle,
                                                                color: global
                                                                    .theme.positiveHighlightTextColor,
                                                              )
                                                            : null,
                                                        onTap: () {
                                                          setState(() {
                                                            _selectedLocation =
                                                                location;
                                                          });
                                                        },
                                                      );
                                                    },
                                                  ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
            ),

            // Footer with action buttons
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                border: Border(
                  top: BorderSide(color: global.theme.dividerBorderColor),
                ),
              ),
              child: Row(
                children: [
                  // แสดงข้อมูลที่เลือก
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (_selectedWarehouse != null) ...[
                          Text(
                            '${global.language("warehouse")}: ${_selectedWarehouse!.displayName}',
                            style: TextStyle(fontWeight: FontWeight.bold),
                          ),
                          if (_selectedLocation != null)
                            Text(
                              '${global.language("location")}: ${_selectedLocation!.displayName}',
                              style: TextStyle(color: global.theme.textColor),
                            ),
                        ] else
                          Text(
                            global.language("warehouse_not_selected"),
                            style: TextStyle(color: global.theme.iconSecondaryColor),
                          ),
                      ],
                    ),
                  ),
                  SizedBox(width: 16),
                  TextButton(
                    onPressed: () => Navigator.pop(context),
                    child: Text(global.language("cancel")),
                  ),
                  SizedBox(width: 8),
                  ElevatedButton(
                    onPressed: _confirmSelection,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.infoHighlightTextColor,
                      foregroundColor: global.theme.onPrimaryColor,
                    ),
                    child: Text(global.language("confirm")),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
