import 'dart:io';

import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

class MultiBranchSearchScreen extends StatefulWidget {
  const MultiBranchSearchScreen({
    super.key,
    required this.word,
    this.preSelectedBranches = const [],
  });

  final String word;
  final List<SearchGuidCodeNameModel> preSelectedBranches;

  @override
  State<MultiBranchSearchScreen> createState() =>
      _MultiBranchSearchScreenState();
}

class _MultiBranchSearchScreenState extends State<MultiBranchSearchScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<CompanyBranchModel> companyBranchListData = [];
  Set<String> selectedBranchGuids = {};
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
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    searchText = widget.word;
    searchController.text = searchText;

    // Initialize pre-selected branches
    selectedBranchGuids = widget.preSelectedBranches
        .map((branch) => branch.guid)
        .toSet();

    super.initState();
  }

  void loadDataList(String search) {
    if (_isLoading || !_hasMoreData) return;

    setState(() {
      _isLoading = true;
    });

    context.read<CompanyBranchBloc>().add(
      CompanyBranchLoadList(
        offset: (companyBranchListData.isEmpty)
            ? 0
            : companyBranchListData.length,
        limit: global.loadDataPerPage,
        search: search,
      ),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText);
    }
  }

  void toggleBranchSelection(CompanyBranchModel branch) {
    setState(() {
      if (selectedBranchGuids.contains(branch.guidfixed)) {
        selectedBranchGuids.remove(branch.guidfixed);
      } else {
        selectedBranchGuids.add(branch.guidfixed);
      }
    });
  }

  void selectAllVisible() {
    setState(() {
      for (var branch in companyBranchListData) {
        selectedBranchGuids.add(branch.guidfixed);
      }
    });
  }

  void clearAllSelection() {
    setState(() {
      selectedBranchGuids.clear();
    });
  }

  List<SearchGuidCodeNameModel> getSelectedBranches() {
    // Get branches from current data
    final currentDataBranches = companyBranchListData
        .where((branch) => selectedBranchGuids.contains(branch.guidfixed))
        .map(
          (branch) => SearchGuidCodeNameModel(
            guid: branch.guidfixed,
            code: branch.code,
            names: branch.names,
            isCancel: false,
          ),
        )
        .toList();

    // Add pre-selected branches that might not be in current data
    final preSelectedNotInData = widget.preSelectedBranches
        .where(
          (preBranch) =>
              selectedBranchGuids.contains(preBranch.guid) &&
              !currentDataBranches.any(
                (current) => current.guid == preBranch.guid,
              ),
        )
        .toList();

    return [...currentDataBranches, ...preSelectedNotInData];
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.cardColor,
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.escape) {
                Navigator.pop(context, BranchSelectionModel.cancelled());
                return KeyEventResult.handled;
              }
              if (event.logicalKey == LogicalKeyboardKey.enter &&
                  selectedBranchGuids.isNotEmpty) {
                Navigator.pop(
                  context,
                  BranchSelectionModel.fromBranches(getSelectedBranches()),
                );
                return KeyEventResult.handled;
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: BlocListener<CompanyBranchBloc, CompanyBranchState>(
          listener: (context, state) {
            if (state is CompanyBranchLoadSuccess) {
              setState(() {
                _isLoading = false;
                if (state.companyBranch.isNotEmpty) {
                  companyBranchListData.addAll(state.companyBranch);
                  _hasMoreData =
                      state.companyBranch.length >= global.loadDataPerPage;
                } else {
                  _hasMoreData = false;
                }
              });
            }
            if (state is CompanyBranchLoadFailed) {
              setState(() {
                _isLoading = false;
              });
            }
          },
          child: Column(
            children: [
              // Simple AppBar
              AppBar(
                title: Text('เลือกสาขา (${selectedBranchGuids.length})'),
                backgroundColor: global.theme.primaryColor,
                foregroundColor: global.theme.onPrimaryColor,
                elevation: 1,
                leading: IconButton(
                  icon: Icon(Icons.arrow_back),
                  onPressed: () {
                    Navigator.pop(context, BranchSelectionModel.cancelled());
                  },
                ),
                actions: [
                  // Select All
                  IconButton(
                    icon: Icon(Icons.select_all),
                    onPressed: selectAllVisible,
                    tooltip: global.language('select_all'),
                  ),
                  // Clear All
                  IconButton(
                    icon: Icon(Icons.clear_all),
                    onPressed: clearAllSelection,
                    tooltip: global.language('clear_all'),
                  ),
                  IconButton(
                    icon: Icon(Icons.check),
                    onPressed: () {
                      Navigator.pop(
                        context,
                        BranchSelectionModel.fromBranches(
                          getSelectedBranches(),
                        ),
                      );
                    },
                    tooltip: 'ยืนยัน (${selectedBranchGuids.length})',
                  ),
                ],
              ),

              // Simple Search Box
              Container(
                margin: const EdgeInsets.all(8),
                padding: const EdgeInsets.symmetric(horizontal: 12),
                decoration: BoxDecoration(
                  color: global.theme.dividerBorderColor,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    Icon(Icons.search, color: global.theme.textSecondaryColor),
                    SizedBox(width: 8),
                    Expanded(
                      child: TextFormField(
                        onFieldSubmitted: (value) {
                          searchFocusNode.requestFocus();
                        },
                        onChanged: (value) {
                          try {
                            _debouncer.run(() {
                              setState(() {
                                companyBranchListData.clear();
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
                        style: TextStyle(color: global.theme.textColor),
                        decoration: InputDecoration(
                          filled: false,
                          border: InputBorder.none,
                          hintText: global.language('search_branch_placeholder'),
                          hintStyle: TextStyle(color: global.theme.formHintColor),
                          contentPadding: EdgeInsets.symmetric(vertical: 12),
                        ),
                      ),
                    ),
                    if (searchController.text.isNotEmpty)
                      IconButton(
                        icon: Icon(Icons.clear, color: global.theme.textSecondaryColor),
                        onPressed: () {
                          searchController.clear();
                          setState(() {
                            companyBranchListData.clear();
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

              // List Content
              Expanded(
                child: companyBranchListData.isEmpty && !_isLoading
                    ? _buildEmptyState()
                    : Column(
                        children: [
                          Expanded(
                            child: ListView.builder(
                              controller: listScrollController,
                              itemCount:
                                  companyBranchListData.length +
                                  (_isLoading ? 1 : 0),
                              itemBuilder: (context, index) {
                                if (index == companyBranchListData.length) {
                                  // Loading indicator at bottom
                                  return const Padding(
                                    padding: EdgeInsets.all(16),
                                    child: Center(
                                      child: CircularProgressIndicator(),
                                    ),
                                  );
                                }
                                return _buildBranchItem(
                                  companyBranchListData[index],
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
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.search_off, size: 64, color: global.theme.iconSecondaryColor),
          const SizedBox(height: 16),
          Text(
            searchText.isEmpty ? 'ไม่พบข้อมูลสาขา' : 'ไม่พบสาขาที่ค้นหา',
            style: TextStyle(fontSize: 18, color: global.theme.textSecondaryColor),
          ),
          if (searchText.isNotEmpty) ...[
            SizedBox(height: 16),
            TextButton(
              onPressed: () {
                searchController.clear();
                setState(() {
                  companyBranchListData.clear();
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

  Widget _buildBranchItem(CompanyBranchModel value) {
    final isSelected = selectedBranchGuids.contains(value.guidfixed);
    final branchName = global.packName(value.names);

    return ListTile(
      leading: Checkbox(
        value: isSelected,
        onChanged: (_) => toggleBranchSelection(value),
        activeColor: global.theme.primaryColor,
      ),
      title: Text(
        branchName,
        style: TextStyle(
          fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
        ),
      ),
      subtitle: Text('รหัส: ${value.code}'),
      onTap: () => toggleBranchSelection(value),
      tileColor: isSelected ? global.theme.primaryColor.withValues(alpha: 0.1) : null,
      trailing: isSelected
          ? Icon(Icons.check_circle, color: global.theme.primaryColor)
          : null,
    );
  }
}
