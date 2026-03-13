import 'dart:io';

import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/model/department_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

class DepartmentInBranchSearchScreen extends StatefulWidget {
  const DepartmentInBranchSearchScreen({super.key, required this.guid});
  final String guid;

  @override
  State<DepartmentInBranchSearchScreen> createState() =>
      DepartmentInBranchSearchScreenState();
}

class DepartmentInBranchSearchScreenState
    extends State<DepartmentInBranchSearchScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String guid = "";
  List<DepartmentModel> companyDepartmentListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int _hoverIndex = -1;
  int currentListIndex = 0;

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    loadDataList(guid);
  }

  @override
  void initState() {
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    guid = widget.guid;

    super.initState();
  }

  void loadDataList(String guid) {
    context.read<CompanyBranchBloc>().add(CompanyBranchGet(guid: guid));
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(guid);
    }
  }

  @override
  void dispose() {
    listScrollController.dispose();
    super.dispose();
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('company_branch')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(
              context,
              SearchGuidCodeNameModel(
                guid: "",
                code: "",
                names: [],
                isCancel: true,
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
              // print(event.logicalKey.keyLabel);
              if (event.logicalKey == LogicalKeyboardKey.escape) {
                Navigator.pop(
                  context,
                  SearchGuidCodeNameModel(
                    guid: "",
                    code: "",
                    names: [],
                    isCancel: true,
                  ),
                );
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  Navigator.pop(
                    context,
                    SearchGuidCodeNameModel(
                      guid:
                          companyDepartmentListData[currentListIndex].guidfixed,
                      code: companyDepartmentListData[currentListIndex].code,
                      names: companyDepartmentListData[currentListIndex].names,
                      isCancel: false,
                    ),
                  );
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = companyDepartmentListData.indexOf(
                  companyDepartmentListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = companyDepartmentListData[index - 1].guidfixed;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = companyDepartmentListData.indexOf(
                  companyDepartmentListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index < companyDepartmentListData.length - 1) {
                  selectGuid = companyDepartmentListData[index + 1].guidfixed;
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
                      global.language("company_branch_code"),
                      style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("company_branch_name"),
                      style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: companyDepartmentListData
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

  Widget listObject(DepartmentModel value, int index) {
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
          SearchGuidCodeNameModel(
            guid: value.guidfixed,
            code: value.code,
            names: value.names,
            isCancel: false,
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
              flex: 10,
              child: Text(
                global.packName(value.names),
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
    for (int i = 0; i < companyDepartmentListData.length; i++) {
      if (companyDepartmentListData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<CompanyBranchBloc, CompanyBranchState>(
            listener: (context, state) {
              // Load
              if (state is CompanyBranchGetSuccess) {
                setState(() {
                  // // print(state.companyBranch.toJson());
                  if (state.companyBranch.departments.isNotEmpty) {
                    companyDepartmentListData.addAll(
                      state.companyBranch.departments,
                    );
                    if (companyDepartmentListData.isNotEmpty) {
                      selectGuid = companyDepartmentListData[0].guidfixed;
                    } else {
                      selectGuid = "";
                    }
                  }
                });
              }
            },
            child: (constraints.maxWidth > 800)
                ? listScreen(mobileScreen: false)
                : listScreen(mobileScreen: true),
          );
        },
      ),
    );
  }
}
