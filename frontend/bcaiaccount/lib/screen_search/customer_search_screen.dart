import 'dart:io';

import 'package:smlaicloud/bloc/debtor/debtor_bloc.dart';
import 'package:smlaicloud/bloc/debtor_group/debtor_group_bloc.dart';
import 'package:smlaicloud/model/debtor_group_model.dart';
import 'package:smlaicloud/model/debtor_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/components/loading_overlay.dart';

class CustomerSearchScreen extends StatefulWidget {
  const CustomerSearchScreen({super.key, required this.word});
  final String word;

  @override
  State<CustomerSearchScreen> createState() => CustomerSearchScreenState();
}

class CustomerSearchScreenState extends State<CustomerSearchScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<DebtorModel> custListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int _hoverIndex = -1;
  int currentListIndex = 0;
  final _debouncer = global.Debouncer(1000);
  bool loadingData = false;
  bool isLoading = false; // สำหรับ LoadingOverlay
  List<DebtorGroupModel> listDataGroup = [];
  List<DebtorGroupModel> selectedFilters = [];
  List<String> selectedFilterCodes = [];

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    loadDataGroupList();
    loadDataList(searchText, []);
  }

  @override
  void initState() {
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    searchText = widget.word;
    searchController.text = searchText;

    super.initState();
  }

  void loadDataGroupList() {
    context.read<DebtorGroupBloc>().add(
      const DebtorGroupLoadList(offset: 0, limit: 100, search: ""),
    );
  }

  void loadDataList(String search, List<String>? filter) {
    setState(() {
      loadingData = true;
    });
    searchText = search;

    context.read<DebtorBloc>().add(
      DebtorLoadList(
        offset: (custListData.isEmpty) ? 0 : custListData.length,
        limit: 100,
        search: search,
        groups: filter,
      ),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText, selectedFilterCodes);
    }
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    super.dispose();
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('debtor')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(
              context,
              global.SearchDebtorModel(
                guid: '',
                code: '',
                names: [],
                ismember: false,
                pricelevel: '',
                pointscode: '',
              ),
            );
          },
        ),
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.escape) {
                Navigator.pop(
                  context,
                  global.SearchDebtorModel(
                    guid: custListData[currentListIndex].guidfixed,
                    code: custListData[currentListIndex].code,
                    names: custListData[currentListIndex].names,
                    ismember: false,
                    pricelevel: custListData[currentListIndex].pricelevel!,
                    pointscode: custListData[currentListIndex].pointscode ?? '',
                  ),
                );
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  Navigator.pop(
                    context,
                    global.SearchDebtorModel(
                      guid: custListData[currentListIndex].guidfixed,
                      code: custListData[currentListIndex].code,
                      names: custListData[currentListIndex].names,
                      ismember: custListData[currentListIndex].ismember!,
                      pricelevel: custListData[currentListIndex].pricelevel!,
                      pointscode:
                          custListData[currentListIndex].pointscode ?? '',
                    ),
                  );
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = custListData.indexOf(
                  custListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = custListData[index - 1].guidfixed;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = custListData.indexOf(
                  custListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index < custListData.length - 1) {
                  selectGuid = custListData[index + 1].guidfixed;
                }
                isKeyDown = true;
                setState(() {});
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(5),
              color: global.theme.appBarColor,
              child: Container(
                padding: const EdgeInsets.all(5),
                decoration: BoxDecoration(
                  color: global.theme.surfaceColor,
                  borderRadius: BorderRadius.circular(2),
                  boxShadow: [
                    BoxShadow(
                      color: global.theme.dividerBorderColor.withValues(alpha: 0.5),
                      spreadRadius: 5,
                      blurRadius: 7,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: TextFormField(
                        onFieldSubmitted: (value) {
                          searchFocusNode.requestFocus();
                        },
                        onChanged: (value) {
                          _debouncer.run(() {
                            try {
                              setState(() {
                                custListData = [];
                              });
                              loadDataList(value, selectedFilterCodes);
                            } catch (_) {}
                          });
                        },
                        autofocus: true,
                        focusNode: searchFocusNode,
                        controller: searchController,
                        style: TextStyle(color: global.theme.textColor),
                        decoration: InputDecoration(
                          isDense: true,
                          filled: false,
                          contentPadding: const EdgeInsets.only(
                            top: 10,
                            bottom: 10,
                          ),
                          border: InputBorder.none,
                          hintText: global.language('search'),
                          hintStyle: TextStyle(color: global.theme.formHintColor),
                        ),
                      ),
                    ),
                    IconButton(
                      onPressed: () async {
                        selectedFilters = await filterDebtorGroup(
                          selectedFilters,
                        );
                        if (selectedFilters.isNotEmpty) {
                          selectedFilterCodes.clear();
                          for (var element in selectedFilters) {
                            selectedFilterCodes.add(element.guidfixed);
                          }
                        } else {
                          selectedFilterCodes.clear();
                        }
                        custListData = [];
                        loadDataList(searchText, selectedFilterCodes);
                        setState(() {});
                      },
                      icon: Icon(
                        (selectedFilters.isEmpty)
                            ? Icons.filter_alt_off
                            : Icons.filter_alt,
                        color: (selectedFilters.isEmpty)
                            ? global.theme.textColor
                            : global.theme.primaryColor,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Container(
              padding: const EdgeInsets.only(
                left: 10,
                right: 10,
                top: 5,
                bottom: 5,
              ),
              decoration: BoxDecoration(
                color: global.theme.columnHeaderColor,
                border: Border(
                bottom: BorderSide(width: 1.0, color: global.theme.dividerBorderColor),
                ),
              ),
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("debtor_code"),
                      style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("debtor_name"),
                      style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("address"),
                      style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("telephone"),
                      style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: custListData
                      .asMap().entries.map((e) => listObject(e.value, e.key))
                      .toList(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }


  Color _getContainerColor(String itemGuid, int index) {
    if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
      return global.theme.rowSelectedColor;
    }
    if (_hoverIndex == index) {
      return global.theme.rowHoverColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  Widget listObject(DebtorModel value, int index) {
    final isSelected = selectGuid == value.guidfixed;
    TextStyle textStyle = isSelected
        ? TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w700, color: global.theme.textColor)
        : TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w400, color: global.theme.textSecondaryColor);
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
      onTap: () {
        Navigator.pop(
          context,
          global.SearchDebtorModel(
            guid: value.guidfixed,
            code: value.code,
            names: value.names,
            ismember: value.ismember!,
            pricelevel: value.pricelevel!,
            pointscode: value.pointscode ?? '',
          ),
        );
      },
      child: Container(
        decoration: BoxDecoration(
          color: _getContainerColor(value.guidfixed ?? "", index),
          border: Border(
                bottom: BorderSide(width: 1.0, color: global.theme.dividerBorderColor),
          ),
        ),
        padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              flex: 5,
              child: Text(
                value.code,
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                global.packName(value.names),
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                (value.addressforbilling.address!.isNotEmpty)
                    ? value.addressforbilling.address![0]
                    : '',
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                (value.addressforbilling.phoneprimary!.isNotEmpty)
                    ? value.addressforbilling.phoneprimary!
                    : '',
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    ),
    );
  }

  @override
  Widget build(BuildContext context) {
    for (int i = 0; i < custListData.length; i++) {
      if (custListData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<DebtorBloc, DebtorState>(
                listener: (context, state) {
                  // เริ่ม Loading
                  if (state is DebtorInProgress) {
                    setState(() {
                      isLoading = true;
                    });
                  }

                  // Load Success
                  if (state is DebtorLoadSuccess) {
                    setState(() {
                      isLoading = false; // ปิด Loading
                      if (state.debtors.isNotEmpty) {
                        loadingData = false;
                        custListData.addAll(state.debtors);
                        if (custListData.isNotEmpty) {
                          selectGuid = custListData[0].guidfixed;
                        } else {
                          selectGuid = "";
                        }
                      }
                    });
                  }

                  // Load Failed
                  if (state is DebtorLoadFailed) {
                    setState(() {
                      isLoading = false; // ปิด Loading แม้เกิด error
                    });
                  }
                },
              ),
              BlocListener<DebtorGroupBloc, DebtorGroupState>(
                listener: (context, state) {
                  // Load
                  if (state is DebtorGroupLoadSuccess) {
                    setState(() {
                      if (state.debtorGroups.isNotEmpty) {
                        listDataGroup = state.debtorGroups;
                      }
                    });
                  }
                },
              ),
            ],
            child: Stack(
              children: [
                // Main content
                (constraints.maxWidth > 800)
                    ? listScreen(mobileScreen: false)
                    : listScreen(mobileScreen: true),

                // Loading Overlay
                if (isLoading)
                  LoadingOverlay(
                    message: global.language('loading_customer_data'),
                    subtitle: global.language('please_wait'),
                  ),
              ],
            ),
          );
        },
      ),
    );
  }

  Future<List<DebtorGroupModel>> filterDebtorGroup(
    List<DebtorGroupModel> selectedFilters,
  ) async {
    List<DebtorGroupModel> selectedValues = selectedFilters;

    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setState) {
            return AlertDialog(
              title: Column(
                children: [
                  Text(global.language("filter_debtor_group")),
                  const Divider(),
                  Wrap(
                    spacing: 8.0,
                    children: selectedValues.map((filter) {
                      return InputChip(
                        label: Text(filter.groupcode),
                        deleteIcon: Icon(Icons.close),
                        onDeleted: () {
                          setState(() {
                            selectedValues.remove(filter);
                          });
                        },
                      );
                    }).toList(),
                  ),
                  const Divider(),
                ],
              ),
              content: SizedBox(
                width: double.maxFinite,
                child: ListView.builder(
                  itemCount: listDataGroup.length,
                  itemBuilder: (BuildContext context, int index) {
                    final filter = listDataGroup[index];
                    final isSelected = selectedValues.contains(filter);
                    return CheckboxListTile(
                      title: Text(
                        "${filter.groupcode} ~ ${(global.packName(filter.names))} ",
                      ),
                      value: isSelected,
                      onChanged: (bool? value) {
                        setState(() {
                          if (value == true) {
                            selectedValues.add(filter);
                          } else {
                            selectedValues.remove(filter);
                          }
                        });
                      },
                    );
                  },
                ),
              ),
              actions: [
                ElevatedButton(
                  onPressed: () {
                    Navigator.pop(context);
                  },
                  child: Text(global.language("filter")),
                ),
                ElevatedButton(
                  onPressed: () {
                    setState(() {
                      Navigator.pop(context);
                      selectedValues.clear();
                    });
                  },
                  child: Text(global.language("cancel")),
                ),
              ],
            );
          },
        );
      },
    );

    return selectedValues;
  }
}
