import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/creditor/creditor_bloc.dart';
import 'package:smlaicloud/bloc/debtor/debtor_bloc.dart';
import 'package:smlaicloud/bloc/employee/employee_bloc.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/model/creditor_model.dart';
import 'package:smlaicloud/model/debtor_model.dart';
import 'package:smlaicloud/model/employee_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/bi_report/entity_selection_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/product_model.dart';

enum EntityType { creditor, employee, debtor, barcode }

class MultiEntitySearchScreen extends StatefulWidget {
  const MultiEntitySearchScreen({
    super.key,
    required this.word,
    required this.entityType,
    this.preSelectedEntities = const [],
  });

  final String word;
  final EntityType entityType;
  final List<SearchGuidCodeNameModel> preSelectedEntities;

  @override
  State<MultiEntitySearchScreen> createState() =>
      _MultiEntitySearchScreenState();
}

class _MultiEntitySearchScreenState extends State<MultiEntitySearchScreen> {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<dynamic> entityListData = []; // Can hold CreditorModel or EmployeeModel
  Set<String> selectedEntityGuids = {};
  final _debouncer = global.Debouncer(1000);

  // Performance optimization - cache loaded data
  bool _isLoading = false;
  bool _hasMoreData = true;

  void setSystemLanguageList() async {
    try {
      await global.setSystemLanguage(context);
      loadDataList(searchText);
    } catch (e) {
      // Handle language setting error gracefully
      loadDataList(searchText);
    }
  }

  @override
  void initState() {
    super.initState();
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    searchText = widget.word;
    searchController.text = searchText;

    // Initialize pre-selected entities
    selectedEntityGuids = widget.preSelectedEntities
        .map((entity) => entity.guid)
        .toSet();
  }

  void onScrollList() {
    if (listScrollController.position.pixels ==
        listScrollController.position.maxScrollExtent) {
      // เมื่อ scroll ถึงจุดสุดท้าย ให้โหลดข้อมูลเพิ่ม
      if (!_isLoading && _hasMoreData) {
        loadDataList(searchText);
      }
    }
  }

  @override
  void dispose() {
    listScrollController.removeListener(onScrollList); // เพิ่มบรรทัดนี้
    listScrollController.dispose();
    searchController.dispose();
    super.dispose();
  }

  void loadDataList(String search) {
    if (_isLoading || !_hasMoreData) return;

    setState(() {
      _isLoading = true;
    });

    if (widget.entityType == EntityType.creditor) {
      context.read<CreditorBloc>().add(
        CreditorLoadList(
          offset: (entityListData.isEmpty) ? 0 : entityListData.length,
          limit: global.loadDataPerPage,
          search: search,
          groups: [],
        ),
      );
    } else if (widget.entityType == EntityType.debtor) {
      // เพิ่มการจัดการ debtor
      context.read<DebtorBloc>().add(
        DebtorLoadList(
          offset: (entityListData.isEmpty) ? 0 : entityListData.length,
          limit: global.loadDataPerPage,
          search: search,
          groups: [],
        ),
      );
    } else if (widget.entityType == EntityType.employee) {
      context.read<EmployeeBloc>().add(
        EmployeeLoadList(
          offset: (entityListData.isEmpty) ? 0 : entityListData.length,
          limit: global.loadDataPerPage,
          search: search,
        ),
      );
    } else if (widget.entityType == EntityType.barcode) {
      context.read<ProductBarcodeBloc>().add(
        ProductBarcodeLoadListSearch(
          offset: (entityListData.isEmpty) ? 0 : entityListData.length,
          limit: global.loadDataPerPage,
          search: search,
          itemtype: "0,1,2,3,4,5",
          branchcode: global.companyBranchSelectData.code,
          businesstypecode: global.companyBranchSelectData.businesstype!.code!,
          isbom: "all",
          isusesubbarcodes: "notshowsubbarcodes",
          shopsid: global.getShopId(),
        ),
      );
    } else {
      // dilog ไม่รองรับ EntityType
      showDialog(
        context: context,
        builder: (context) {
          return AlertDialog(
            title: Text(global.language("unsupported_type")),
            content: Text('กรุณาเลือกประเภทอื่น'),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(context).pop(),
                child: Text(global.language('close')),
              ),
            ],
          );
        },
      );
    }
  }

  void toggleEntitySelection(dynamic entity) {
    setState(() {
      String guid = '';
      if (entity is CreditorModel) {
        guid = entity.guidfixed;
      } else if (entity is DebtorModel) {
        guid = entity.guidfixed;
      } else if (entity is EmployeeModel) {
        guid = entity.guidfixed;
      } else if (entity is ProductBarcodeModel) {
        guid = entity.guidfixed;
      }

      if (guid.isEmpty) return;

      if (widget.entityType == EntityType.barcode) {
        // สำหรับ barcode: เลือกได้เพียงรายการเดียว
        if (selectedEntityGuids.contains(guid)) {
          selectedEntityGuids.clear(); // ยกเลิกการเลือก
        } else {
          selectedEntityGuids.clear(); // ลบการเลือกเก่า
          selectedEntityGuids.add(guid); // เลือกรายการใหม่
        }
      } else {
        // สำหรับประเภทอื่น: เลือกได้หลายรายการ
        if (selectedEntityGuids.contains(guid)) {
          selectedEntityGuids.remove(guid);
        } else {
          selectedEntityGuids.add(guid);
        }
      }
    });
  }

  void selectAllVisible() {
    setState(() {
      for (var entity in entityListData) {
        String guid = '';
        if (entity is CreditorModel) {
          guid = entity.guidfixed;
        } else if (entity is DebtorModel) {
          guid = entity.guidfixed;
        } else if (entity is EmployeeModel) {
          guid = entity.guidfixed;
        } else if (entity is ProductBarcodeModel) {
          guid = entity.guidfixed;
        }
        selectedEntityGuids.add(guid);
      }
    });
  }

  void clearAllSelection() {
    setState(() {
      selectedEntityGuids.clear();
    });
  }

  List<SearchGuidCodeNameModel> getSelectedEntities() {
    // Get entities from current data
    final currentDataEntities = entityListData
        .where((entity) {
          String guid = '';
          if (entity is CreditorModel) {
            guid = entity.guidfixed;
          } else if (entity is DebtorModel) {
            guid = entity.guidfixed;
          } else if (entity is EmployeeModel) {
            guid = entity.guidfixed;
          } else if (entity is ProductBarcodeModel) {
            guid = entity.guidfixed;
          }
          return selectedEntityGuids.contains(guid);
        })
        .map((entity) {
          if (entity is CreditorModel) {
            return SearchGuidCodeNameModel(
              guid: entity.guidfixed,
              code: entity.code,
              names: entity.names,
              isCancel: false,
            );
          } else if (entity is DebtorModel) {
            return SearchGuidCodeNameModel(
              guid: entity.guidfixed,
              code: entity.code,
              names: entity.names,
              isCancel: false,
            );
          } else if (entity is EmployeeModel) {
            return SearchGuidCodeNameModel(
              guid: entity.guidfixed,
              code: entity.code,
              names: [LanguageDataModel(code: 'th', name: entity.name)],
              isCancel: false,
            );
          } else if (entity is ProductBarcodeModel) {
            return SearchGuidCodeNameModel(
              guid: entity.guidfixed,
              code: entity.barcode!,
              names: entity.names!,
              isCancel: false,
            );
          }
          return null;
        })
        .whereType<SearchGuidCodeNameModel>()
        .toList();

    // Add pre-selected entities that might not be in current data
    final preSelectedNotInData = widget.preSelectedEntities
        .where(
          (preEntity) =>
              selectedEntityGuids.contains(preEntity.guid) &&
              !currentDataEntities.any(
                (current) => current.guid == preEntity.guid,
              ),
        )
        .toList();

    return [...currentDataEntities, ...preSelectedNotInData];
  }

  String _getTitle() {
    switch (widget.entityType) {
      case EntityType.creditor:
        return 'เลือกคู่ค้า (เจ้าหนี้)';
      case EntityType.debtor:
        return 'เลือกคู่ค้า (ลูกหนี้)';
      case EntityType.employee:
        return global.language('select_salesperson');
      case EntityType.barcode:
        return 'เลือกบาร์โค้ด (เลือกได้ 1 รายการ)';
    }
  }

  String _getSearchHint() {
    switch (widget.entityType) {
      case EntityType.creditor:
        return 'ค้นหาคู่ค้า (เจ้าหนี้)...';
      case EntityType.debtor:
        return 'ค้นหาคู่ค้า (ลูกหนี้)...';
      case EntityType.employee:
        return 'ค้นหาพนักงานขาย...';
      case EntityType.barcode:
        return 'ค้นหาบาร์โค้ด...';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.escape) {
                Navigator.pop(context, EntitySelectionModel.cancelled());
                return KeyEventResult.handled;
              }
              if (event.logicalKey == LogicalKeyboardKey.enter &&
                  selectedEntityGuids.isNotEmpty) {
                Navigator.pop(
                  context,
                  EntitySelectionModel.fromEntities(getSelectedEntities()),
                );
                return KeyEventResult.handled;
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: MultiBlocListener(
          listeners: [
            BlocListener<CreditorBloc, CreditorState>(
              listener: (context, state) {
                if (widget.entityType == EntityType.creditor) {
                  if (state is CreditorLoadSuccess) {
                    setState(() {
                      _isLoading = false;
                      if (state.creditors.isNotEmpty) {
                        entityListData.addAll(state.creditors);
                        _hasMoreData =
                            state.creditors.length >= global.loadDataPerPage;
                      } else {
                        _hasMoreData = false;
                      }
                    });
                  }
                  if (state is CreditorLoadFailed) {
                    setState(() {
                      _isLoading = false;
                    });
                  }
                }
              },
            ),
            // เพิ่ม BlocListener สำหรับ ProductBarcode
            BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
              listener: (context, state) {
                if (widget.entityType == EntityType.barcode) {
                  if (state is ProductBarcodeLoadSearchSuccess) {
                    setState(() {
                      _isLoading = false;
                      if (state.productBarcodes.isNotEmpty) {
                        entityListData.addAll(state.productBarcodes);
                        _hasMoreData =
                            state.productBarcodes.length >=
                            global.loadDataPerPage;
                      } else {
                        _hasMoreData = false;
                      }
                    });
                  }
                  if (state is ProductBarcodeLoadSearchFailed) {
                    setState(() {
                      _isLoading = false;
                    });
                  }
                }
              },
            ),

            // เพิ่ม BlocListener สำหรับ DebtorBloc
            BlocListener<DebtorBloc, DebtorState>(
              listener: (context, state) {
                if (widget.entityType == EntityType.debtor) {
                  if (state is DebtorLoadSuccess) {
                    setState(() {
                      _isLoading = false;
                      if (state.debtors.isNotEmpty) {
                        entityListData.addAll(state.debtors);
                        _hasMoreData =
                            state.debtors.length >= global.loadDataPerPage;
                      } else {
                        _hasMoreData = false;
                      }
                    });
                  }
                  if (state is DebtorLoadFailed) {
                    setState(() {
                      _isLoading = false;
                    });
                  }
                }
              },
            ),
            BlocListener<EmployeeBloc, EmployeeState>(
              listener: (context, state) {
                if (widget.entityType == EntityType.employee) {
                  if (state is EmployeeLoadSuccess) {
                    setState(() {
                      _isLoading = false;
                      if (state.employees.isNotEmpty) {
                        entityListData.addAll(state.employees);
                        _hasMoreData =
                            state.employees.length >= global.loadDataPerPage;
                      } else {
                        _hasMoreData = false;
                      }
                    });
                  }
                  if (state is EmployeeLoadFailed) {
                    setState(() {
                      _isLoading = false;
                    });
                  }
                }
              },
            ),
          ],
          child: Column(
            children: [
              // Simple AppBar
              AppBar(
                title: Text('${_getTitle()} (${selectedEntityGuids.length})'),
                backgroundColor: Colors.indigo.shade600,
                foregroundColor: Colors.white,
                elevation: 1,
                leading: IconButton(
                  icon: const Icon(Icons.arrow_back),
                  onPressed: () {
                    Navigator.pop(context, EntitySelectionModel.cancelled());
                  },
                ),
                actions: [
                  // ซ่อน Select All สำหรับ barcode
                  if (widget.entityType != EntityType.barcode)
                    IconButton(
                      icon: const Icon(Icons.select_all),
                      onPressed: selectAllVisible,
                      tooltip: global.language('select_all'),
                    ),

                  // Clear All
                  IconButton(
                    icon: const Icon(Icons.clear_all),
                    onPressed: clearAllSelection,
                    tooltip: global.language('clear_all'),
                  ),

                  // Confirm Button
                  if (selectedEntityGuids.isNotEmpty)
                    IconButton(
                      icon: const Icon(Icons.check),
                      onPressed: () {
                        Navigator.pop(
                          context,
                          EntitySelectionModel.fromEntities(
                            getSelectedEntities(),
                          ),
                        );
                      },
                      tooltip: widget.entityType == EntityType.barcode
                          ? 'ยืนยัน (1 รายการ)'
                          : 'ยืนยัน (${selectedEntityGuids.length})',
                    ),
                ],
              ),

              // Simple Search Box
              Container(
                margin: const EdgeInsets.all(8),
                padding: const EdgeInsets.symmetric(horizontal: 12),
                decoration: BoxDecoration(
                  color: Colors.grey.shade100,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    Icon(Icons.search, color: Colors.grey.shade600),
                    const SizedBox(width: 8),
                    Expanded(
                      child: TextFormField(
                        onFieldSubmitted: (value) {
                          searchFocusNode.requestFocus();
                        },
                        onChanged: (value) {
                          try {
                            _debouncer.run(() {
                              setState(() {
                                entityListData.clear();
                                searchText = value;
                                _hasMoreData = true;
                                _isLoading = false;
                              });
                              loadDataList(value);
                            });
                          } catch (_) {}
                        },
                        autofocus: true,
                        focusNode: searchFocusNode,
                        controller: searchController,
                        decoration: InputDecoration(
                          border: InputBorder.none,
                          hintText: _getSearchHint(),
                          contentPadding: const EdgeInsets.symmetric(
                            vertical: 12,
                          ),
                        ),
                      ),
                    ),
                    if (searchController.text.isNotEmpty)
                      IconButton(
                        icon: Icon(Icons.clear, color: Colors.grey.shade600),
                        onPressed: () {
                          searchController.clear();
                          setState(() {
                            entityListData.clear();
                            searchText = '';
                            _hasMoreData = true;
                            _isLoading = false;
                          });
                          loadDataList('');
                        },
                      ),
                  ],
                ),
              ),

              // แสดงส่วนที่เลือกแล้วแบบ Wrap Widget (เหมือนสินค้า)
              if (selectedEntityGuids.isNotEmpty)
                _buildSelectedEntitiesWidget(),

              // List Content
              Expanded(
                child: entityListData.isEmpty && !_isLoading
                    ? _buildEmptyState()
                    : Column(
                        children: [
                          Expanded(
                            child:
                                _buildEntityGrid(), // เปลี่ยนเป็น grid แบบ wrap
                          ),
                          // Loading indicator at bottom
                          if (_isLoading)
                            const Padding(
                              padding: EdgeInsets.all(16),
                              child: Center(child: CircularProgressIndicator()),
                            ),
                        ],
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.search_off, size: 64, color: Colors.grey.shade400),
          SizedBox(height: 16),
          Text(
            searchText.isEmpty ? global.language('no_data') : global.language('import_product_detail.no_data_found'),
            style: TextStyle(fontSize: 18, color: Colors.grey.shade600),
          ),
          if (searchText.isNotEmpty) ...[
            SizedBox(height: 16),
            TextButton(
              onPressed: () {
                searchController.clear();
                setState(() {
                  entityListData.clear();
                  searchText = '';
                  _hasMoreData = true;
                  _isLoading = false;
                });
                loadDataList('');
              },
              child: Text(global.language('clear_search')),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildEntityItem(dynamic entity) {
    String guid, code, name;

    if (entity is CreditorModel) {
      guid = entity.guidfixed;
      code = entity.code;
      name = global.packName(entity.names);
    } else if (entity is DebtorModel) {
      guid = entity.guidfixed;
      code = entity.code;
      name = global.packName(entity.names);
    } else if (entity is EmployeeModel) {
      guid = entity.guidfixed;
      code = entity.code;
      name = entity.name;
    } else if (entity is ProductBarcodeModel) {
      // เพิ่มการจัดการ ProductBarcodeModel
      guid = entity.guidfixed;
      code = entity.barcode!;
      name = global.packName(entity.names!);
    } else {
      // Fallback case
      guid = '';
      code = '';
      name = 'Unknown Entity';
    }

    final isSelected = selectedEntityGuids.contains(guid);

    return ListTile(
      leading: widget.entityType == EntityType.barcode
          ? Icon(Icons.qr_code, color: Colors.purple.shade600)
          : Checkbox(
              value: isSelected,
              onChanged: (_) => toggleEntitySelection(entity),
              activeColor: Colors.indigo.shade600,
            ),
      title: Text(
        name,
        style: TextStyle(
          fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
        ),
      ),
      subtitle: Text(
        widget.entityType == EntityType.barcode
            ? 'รหัสบาร์โค้ด: $code'
            : 'รหัส: $code',
      ),
      onTap: () {
        if (widget.entityType == EntityType.barcode) {
          // สำหรับ barcode: เลือกแล้วกลับทันที
          final selectedEntity = SearchGuidCodeNameModel(
            guid: guid,
            code: code,
            names: entity is ProductBarcodeModel
                ? entity.names!
                : [LanguageDataModel(code: 'th', name: name)],
            isCancel: false,
          );
          Navigator.pop(
            context,
            EntitySelectionModel.fromEntities([selectedEntity]),
          );
        } else {
          // สำหรับประเภทอื่น: ใช้ toggle selection ปกติ
          toggleEntitySelection(entity);
        }
      },
      tileColor: isSelected ? Colors.indigo.shade50 : null,
      trailing: isSelected
          ? Icon(Icons.check_circle, color: Colors.indigo.shade600)
          : null,
    );
  }

  /// สร้าง widget แสดงรายการแบบ wrap grid (แทน ListView)
  Widget _buildEntityGrid() {
    return SingleChildScrollView(
      controller: listScrollController,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: Wrap(
        spacing: 8,
        runSpacing: 8,
        children: [...entityListData.map((entity) => _buildEntityCard(entity))],
      ),
    );
  }

  /// สร้าง card สำหรับแต่ละ entity
  Widget _buildEntityCard(dynamic entity) {
    String guid, code, name;

    if (entity is CreditorModel) {
      guid = entity.guidfixed;
      code = entity.code;
      name = global.packName(entity.names);
    } else if (entity is DebtorModel) {
      guid = entity.guidfixed;
      code = entity.code;
      name = global.packName(entity.names);
    } else if (entity is EmployeeModel) {
      guid = entity.guidfixed;
      code = entity.code;
      name = entity.name;
    } else if (entity is ProductBarcodeModel) {
      guid = entity.guidfixed;
      code = entity.barcode!;
      name = global.packName(entity.names!);
    } else {
      guid = '';
      code = '';
      name = 'Unknown Entity';
    }

    final isSelected = selectedEntityGuids.contains(guid);

    return GestureDetector(
      onTap: () {
        if (widget.entityType == EntityType.barcode) {
          // สำหรับ barcode: เลือกแล้วกลับทันที
          final selectedEntity = SearchGuidCodeNameModel(
            guid: guid,
            code: code,
            names: entity is ProductBarcodeModel
                ? entity.names!
                : [LanguageDataModel(code: 'th', name: name)],
            isCancel: false,
          );
          Navigator.pop(
            context,
            EntitySelectionModel.fromEntities([selectedEntity]),
          );
        } else {
          toggleEntitySelection(entity);
        }
      },
      child: Container(
        width: 160,
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isSelected ? Colors.indigo.shade50 : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected ? Colors.indigo.shade400 : Colors.grey.shade300,
            width: 1.5,
          ),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.05),
              blurRadius: 4,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            // Icon และ checkbox
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: widget.entityType == EntityType.barcode
                        ? Colors.purple.shade100
                        : Colors.indigo.shade100,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(
                    widget.entityType == EntityType.barcode
                        ? Icons.qr_code
                        : widget.entityType == EntityType.creditor
                        ? Icons.person_outline
                        : widget.entityType == EntityType.debtor
                        ? Icons.account_balance_wallet
                        : Icons.people,
                    size: 16,
                    color: widget.entityType == EntityType.barcode
                        ? Colors.purple.shade600
                        : Colors.indigo.shade600,
                  ),
                ),
                const Spacer(),
                widget.entityType == EntityType.barcode
                    ? const SizedBox.shrink()
                    : Checkbox(
                        value: isSelected,
                        onChanged: (_) => toggleEntitySelection(entity),
                        activeColor: Colors.indigo.shade600,
                        visualDensity: VisualDensity.compact,
                      ),
                if (isSelected)
                  Icon(
                    Icons.check_circle,
                    size: 18,
                    color: Colors.indigo.shade600,
                  ),
              ],
            ),
            const SizedBox(height: 8),
            // รหัสและชื่อ
            Text(
              code,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: Colors.grey.shade700,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            const SizedBox(height: 4),
            Flexible(
              child: Text(
                name,
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: isSelected ? FontWeight.w600 : FontWeight.w500,
                  color: isSelected
                      ? Colors.indigo.shade700
                      : Colors.grey.shade800,
                ),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(height: 4),
            // ประเภท
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: widget.entityType == EntityType.barcode
                    ? Colors.purple.shade50
                    : Colors.indigo.shade50,
                borderRadius: BorderRadius.circular(6),
              ),
              child: Text(
                widget.entityType == EntityType.barcode
                    ? global.language('barcode')
                    : widget.entityType == EntityType.creditor
                    ? global.language('creditor')
                    : widget.entityType == EntityType.debtor
                    ? global.language('debtor')
                    : global.language('employee'),
                style: TextStyle(
                  fontSize: 10,
                  color: widget.entityType == EntityType.barcode
                      ? Colors.purple.shade600
                      : Colors.indigo.shade600,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// สร้าง widget สำหรับแสดงรายการที่เลือกแล้วแบบ Wrap (เหมือนสินค้า)
  Widget _buildSelectedEntitiesWidget() {
    final selectedEntities = getSelectedEntities();

    if (selectedEntities.isEmpty) return const SizedBox.shrink();

    return Container(
      margin: const EdgeInsets.all(8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.orange.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.orange.shade300),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.handshake, color: Colors.orange.shade600, size: 20),
              const SizedBox(width: 8),
              Text(
                'คู่ค้าที่เลือก (${selectedEntities.length})',
                style: TextStyle(
                  color: Colors.orange.shade700,
                  fontWeight: FontWeight.w600,
                  fontSize: 14,
                ),
              ),
              Spacer(),
              TextButton(
                onPressed: clearAllSelection,
                style: TextButton.styleFrom(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  minimumSize: Size.zero,
                  tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                ),
                child: Text(
                  global.language('clear_all'),
                  style: TextStyle(color: Colors.red.shade600, fontSize: 12),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 6,
            runSpacing: 6,
            children: [
              ...selectedEntities.map(
                (entity) => _buildSelectedEntityChip(entity),
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// สร้าง chip สำหรับรายการที่เลือก
  Widget _buildSelectedEntityChip(SearchGuidCodeNameModel entity) {
    final displayName = entity.names.first.name;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.orange.shade400),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 2,
            offset: const Offset(0, 1),
          ),
        ],
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.all(2),
            decoration: BoxDecoration(
              color: Colors.orange.shade100,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(
              widget.entityType == EntityType.creditor
                  ? Icons.person_outline
                  : widget.entityType == EntityType.debtor
                  ? Icons.account_balance_wallet
                  : widget.entityType == EntityType.employee
                  ? Icons.people
                  : Icons.qr_code,
              size: 12,
              color: Colors.orange.shade600,
            ),
          ),
          const SizedBox(width: 6),
          Flexible(
            child: Text(
              entity.code,
              style: TextStyle(
                color: Colors.orange.shade700,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 4),
          Flexible(
            child: Text(
              displayName,
              style: TextStyle(
                color: Colors.orange.shade700,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 4),
          GestureDetector(
            onTap: () {
              // ลบออกจากการเลือก
              setState(() {
                selectedEntityGuids.remove(entity.guid);
              });
            },
            child: Container(
              padding: const EdgeInsets.all(2),
              decoration: BoxDecoration(
                color: Colors.red.shade100,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(Icons.close, size: 12, color: Colors.red.shade600),
            ),
          ),
        ],
      ),
    );
  }
}
