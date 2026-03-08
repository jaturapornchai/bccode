import 'dart:io';

import 'package:smlaicloud/bloc/creditor/creditor_bloc.dart';
import 'package:smlaicloud/bloc/creditor_group/creditor_group_bloc.dart';
import 'package:smlaicloud/model/creditor_group_model.dart';
import 'package:smlaicloud/model/creditor_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:loading_animation_widget/loading_animation_widget.dart';

class SupplierSearchScreen extends StatefulWidget {
  const SupplierSearchScreen({super.key, required this.word});
  final String word;

  @override
  State<SupplierSearchScreen> createState() => SupplierSearchScreenState();
}

class SupplierSearchScreenState extends State<SupplierSearchScreen>
    with SingleTickerProviderStateMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<CreditorModel> custListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int currentListIndex = 0;
  final _debouncer = global.Debouncer(1000);
  bool loadingData = false;
  List<CreditorGroupModel> listDataGroup = [];
  List<CreditorGroupModel> selectedFilters = [];
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
    context.read<CreditorGroupBloc>().add(
      const CreditorGroupLoadList(offset: 0, limit: 1000, search: ""),
    );
  }

  void loadDataList(String search, List<String>? filter) {
    setState(() {
      loadingData = true;
    });
    searchText = search;

    context.read<CreditorBloc>().add(
      CreditorLoadList(
        offset: (custListData.isEmpty) ? 0 : custListData.length,
        limit: global.loadDataPerPage,
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
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('supplier')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(
              context,
              SearchGuidCodeNameModel(
                code: '',
                names: [],
                isCancel: false,
                guid: '',
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
                  SearchGuidCodeNameModel(
                    code: custListData[currentListIndex].code,
                    names: custListData[currentListIndex].names,
                    isCancel: false,
                    guid: custListData[currentListIndex].guidfixed,
                  ),
                );
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  Navigator.pop(
                    context,
                    SearchGuidCodeNameModel(
                      code: custListData[currentListIndex].code,
                      names: custListData[currentListIndex].names,
                      isCancel: false,
                      guid: custListData[currentListIndex].guidfixed,
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
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(2),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.grey.withValues(alpha: 0.5),
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
                        decoration: InputDecoration(
                          isDense: true,
                          contentPadding: const EdgeInsets.only(
                            top: 10,
                            bottom: 10,
                          ),
                          border: InputBorder.none,
                          hintText: global.language('search'),
                        ),
                      ),
                    ),
                    IconButton(
                      onPressed: () async {
                        selectedFilters = await filterCrediterGroup(
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
                            ? Colors.black
                            : Colors.blue,
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
                border: const Border(
                  bottom: BorderSide(width: 1.0, color: Colors.grey),
                ),
              ),
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("creditor_code"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("creditor_name"),
                      style: const TextStyle(
                        color: Colors.black,
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
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("telephone"),
                      style: const TextStyle(
                        color: Colors.black,
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
                      .map((value) => listObject(value))
                      .toList(),
                ),
              ),
            ),
            if (loadingData)
              Center(
                child: LoadingAnimationWidget.staggeredDotsWave(
                  color: Colors.blue,
                  size: 50,
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget listObject(CreditorModel value) {
    return GestureDetector(
      onTap: () {
        Navigator.pop(
          context,
          SearchGuidCodeNameModel(
            code: value.code,
            names: value.names,
            isCancel: false,
            guid: value.guidfixed,
          ),
        );
      },
      child: Container(
        decoration: BoxDecoration(
          color: (selectGuid == value.guidfixed)
              ? Colors.cyan[100]
              : Colors.white,
          border: const Border(
            bottom: BorderSide(width: 1.0, color: Colors.grey),
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
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                global.packName(value.names),
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
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
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
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<CreditorBloc, CreditorState>(
                listener: (context, state) {
                  // Load
                  if (state is CreditorLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.creditors.isNotEmpty) {
                        custListData.addAll(state.creditors);
                        if (custListData.isNotEmpty) {
                          selectGuid = custListData[0].guidfixed;
                        } else {
                          selectGuid = "";
                        }
                      }
                    });
                  }
                },
              ),
              BlocListener<CreditorGroupBloc, CreditorGroupState>(
                listener: (context, state) {
                  // Load
                  if (state is CreditorGroupLoadSuccess) {
                    setState(() {
                      if (state.creditorGroups.isNotEmpty) {
                        listDataGroup = state.creditorGroups;
                      }
                    });
                  }
                },
              ),
            ],
            child: (constraints.maxWidth > 800)
                ? listScreen(mobileScreen: false)
                : listScreen(mobileScreen: true),
          );
        },
      ),
    );
  }

  Future<List<CreditorGroupModel>> filterCrediterGroup(
    List<CreditorGroupModel> selectedFilters,
  ) async {
    List<CreditorGroupModel> selectedValues = selectedFilters;

    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setState) {
            return AlertDialog(
              title: Column(
                children: [
                  Text(global.language("filter_credit_group")),
                  const Divider(),
                  Wrap(
                    spacing: 8.0,
                    children: selectedValues.map((filter) {
                      return InputChip(
                        label: Text(filter.groupcode),
                        deleteIcon: const Icon(Icons.close),
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
