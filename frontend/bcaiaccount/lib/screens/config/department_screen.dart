import 'dart:io';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/department/department_bloc.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/department_model.dart';
import 'package:smlaicloud/screen_search/company_branch_search_screen.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/focus_utils.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class DepartmentScreen extends StatefulWidget {
  const DepartmentScreen({super.key});

  @override
  State<DepartmentScreen> createState() => DepartmentScreenState();
}

class DepartmentScreenState extends State<DepartmentScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<TextEditingController> fieldTextController = [];
  List<global.FieldFocusModel> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<DepartmentModel> listData = [];
  List<String> departmentGuidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectCode = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late DepartmentState blocDepartmentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  int _hoverIndex = -1;

  String selectCompanyBranchGuid = "";
  String selectCompanyBranchCode = "";
  late List<LanguageDataModel> selectCompanyBranchName;
  List<DepartmentModel> departmentList = [];
  List<CompanyBranchModel>? selectedDepartment;
  late CompanyBranchModel screenData;
  TextEditingController departmentCode = TextEditingController();

  void changeScreenEvent(global.ScreenEventEnum event) {
    // print(event);
    screenEvent = event;
    for (int index = 0; index < fieldFocusNodes.length; index++) {
      fieldFocusNodes[index].isReadOnly =
          (screenEvent == global.ScreenEventEnum.list) ? true : false;
    }
    if (screenEvent == global.ScreenEventEnum.edit) {
      fieldFocusNodes[0].isReadOnly = true;
    }
    setState(() {});
  }

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }

    for (int i = 0; i < languageList.length; i++) {
      fieldTextController.add(TextEditingController());
      FocusNode focusNode = FocusNode();
      focusNode.addListener(() {
        focusNodeIndex = i;
      });
      fieldFocusNodes.add(global.FieldFocusModel(focusNode: focusNode));
    }
    setState(() {});
  }

  @override
  void initState() {
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    // เรียงลำดับ Focus
    // Focus รหัส
    FocusNode focusNode = FocusNode();
    focusNode.addListener(() {
      focusNodeIndex = 0;
    });
    fieldFocusNodes.add(global.FieldFocusModel(focusNode: focusNode));
    fieldTextController.add(TextEditingController());

    setSystemLanguageList();
    // Auto-load branch list เพื่อเลือกสาขาแรกอัตโนมัติ
    _autoLoadFirstBranch();
    super.initState();
  }

  void _autoLoadFirstBranch() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        context.read<CompanyBranchBloc>().add(
          const CompanyBranchLoadList(offset: 0, limit: 1, search: ""),
        );
      }
    });
  }

  void searchDataList(String searchTerm) {
    departmentList = screenData.departments
        .where(
          (department) =>
              department.names.any(
                (department) => department.name.contains(searchTerm),
              ) ||
              department.code.toString().contains(searchTerm),
        )
        .toList();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].dispose();
    }
    for (int i = 0; i < fieldFocusNodes.length; i++) {
      fieldFocusNodes[i].focusNode.dispose();
    }

    super.dispose();
  }

  void clearEditData() {
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].clear();
    }
    isChange = false;
    focusNodeIndex = 0;
  }

  void discardData({required Function callBack}) {
    // กรณีมีการแก้ไขข้อมูล ให้เตือน
    if ((screenEvent == global.ScreenEventEnum.add ||
            screenEvent == global.ScreenEventEnum.edit) &&
        isChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.buttonNoColor,
              ),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.buttonYesColor,
              ),
              onPressed: () {
                Navigator.pop(context);
                callBack();
              },
              child: Text(global.language('yes')),
            ),
          ],
        ),
      );
    } else {
      callBack();
    }
  }

  void getData(String code) {
    headerEdit = global.language("show");
    changeScreenEvent(global.ScreenEventEnum.list);
    // context.read<DepartmentBloc>().add(DepartmentGet(guid: guid));

    DepartmentModel findData = screenData.departments.firstWhere(
      (element) => element.code == code,
      orElse: () => DepartmentModel(code: ""),
    );

    getDataToEditScreen(findData);
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('department')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                // Navigator.pop(context);
                global.gotoMainMenu(context);
                changeScreenEvent(global.ScreenEventEnum.list);
              },
            );
          },
        ),
        actions: <Widget>[
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              discardData(
                callBack: () {
                  setState(() {
                    if (showCheckBox) {
                      showCheckBox = false;
                      departmentGuidListChecked.clear();
                    } else {
                      showCheckBox = true;
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: global.theme.onPrimaryColor),
                        global.language("choose_item_delete"),
                        global.theme.infoHighlightTextColor,
                      );
                    }
                  });
                },
              );
            },
            icon: (showCheckBox)
                ? Icon(Icons.close)
                : Icon(Icons.check_box),
          ),
          if (departmentGuidListChecked.isNotEmpty)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                showDialog<String>(
                  context: context,
                  builder: (BuildContext context) => AlertDialog(
                    title: Text(global.language('confirm_delete')),
                    actions: <Widget>[
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.buttonNoColor,
                        ),
                        onPressed: () => Navigator.pop(context),
                        child: Text(global.language('no')),
                      ),
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.buttonYesColor,
                        ),
                        onPressed: () {
                          Navigator.pop(context);
                          deleteDataMany();
                        },
                        child: Text(global.language('confirm')),
                      ),
                    ],
                  ),
                );
                setState(() {});
              },
              icon: Icon(Icons.delete),
            ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: selectCompanyBranchGuid.isNotEmpty
                ? () {
                    discardData(
                      callBack: () {
                        setState(() {
                          changeScreenEvent(global.ScreenEventEnum.add);
                          selectCode = "";
                          showCheckBox = false;
                          isChange = false;
                          clearEditData();
                          headerEdit = global.language("append");
                          isSaveAllow = true;
                          WidgetsBinding.instance.addPostFrameCallback((
                            timeStamp,
                          ) {
                            tabController.animateTo(1);
                            focusFirstTextField(context);
                          });
                        });
                      },
                    );
                  }
                : null,
            icon: Icon(Icons.add),
          ),
          const ManualButton(path: 'settings-department'),
        ],
      ),
      body: Column(
        children: [
          Container(
            padding: const EdgeInsets.only(bottom: 5),
            width: double.infinity,
            height: 50,
            child: ElevatedButton(
              onPressed: () {
                discardData(
                  callBack: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) =>
                            const CompanyBranchSearchScreen(word: ""),
                      ),
                    ).then((value) {
                      setState(() {
                        SearchGuidCodeNameModel result = value;
                        if (result.isCancel == false) {
                          selectCompanyBranchGuid = result.guid;
                          selectCompanyBranchCode = result.code;
                          selectCompanyBranchName = result.names;
                          context.read<CompanyBranchBloc>().add(
                            CompanyBranchGet(guid: selectCompanyBranchGuid),
                          );
                        }
                      });
                    });
                  },
                );
              },
              child: Text(
                (selectCompanyBranchGuid.isEmpty)
                    ? global.language("select_company_branch")
                    : selectCompanyBranchName[0].name,
              ),
            ),
          ),
          if (selectCompanyBranchGuid.isEmpty)
            Container(
              margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: Colors.amber[50],
                border: Border.all(color: Colors.amber),
                borderRadius: BorderRadius.circular(6),
              ),
              child: Row(
                children: [
                  Icon(Icons.info_outline, color: Colors.amber, size: 20),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      global.language("please_select_company_branch").replaceAll("*", "").trim(),
                      style: TextStyle(color: Colors.brown, fontSize: 13),
                    ),
                  ),
                ],
              ),
            ),
          Container(
            padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            color: global.theme.searchBarColor,
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    onSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      debouncer.run(() {
                        setState(() {
                          departmentList = [];
                        });
                        searchDataList(value);
                      });
                    },
                    autofocus: false,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.symmetric(vertical: 8, horizontal: 4),
                      border: InputBorder.none,
                      hintText: global.language('search'),
                      prefixIcon: Icon(Icons.search, size: 20),
                      prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                    ),
                  ),
                ),
                ListFontSizeControl(onChanged: () => setState(() {})),
                const SizedBox(width: 4),
                if (departmentList.isNotEmpty)
                  Text('(${departmentList.length})', style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor)),
              ],
            ),
          ),
          Container(color: global.theme.appBarColor, height: 6),
          Container(
            color: global.theme.columnHeaderColor,
            key: headerKey,
            padding: const EdgeInsets.only(
              left: 10,
              right: 10,
              top: 5,
              bottom: 5,
            ),
            child: Row(
              children: [
                Expanded(
                  flex: 5,
                  child: Text(
                    global.language("department_code"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                  ),
                ),
                Expanded(
                  flex: 10,
                  child: Text(
                    global.language("department_name"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (showCheckBox)
                  Expanded(
                    flex: 1,
                    child: Icon(
                      Icons.check,
                      color: global.theme.columnHeaderTextColor,
                      size: 12,
                    ),
                  ),
              ],
            ),
          ),
          Expanded(
            child: RawKeyboardListener(
              autofocus: true,
              focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
              onKey: (RawKeyEvent event) {
                if (screenEvent == global.ScreenEventEnum.list) {
                  if (kIsWeb ||
                      Platform.isWindows ||
                      Platform.isLinux ||
                      Platform.isMacOS) {
                    if (event is RawKeyUpEvent) {
                      try {
                        if (event.logicalKey == LogicalKeyboardKey.f2) {
                          isKeyDown = false;
                          searchFocusNode.requestFocus();
                        }
                        if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                          isKeyDown = false;
                          int index = listData.indexOf(
                            listData.firstWhere(
                              (element) => element.code == selectCode,
                            ),
                          );
                          if (index > 0) {
                            selectCode = listData[index - 1].code;
                            currentListIndex = index - 1;
                            isKeyUp = true;
                            getData(selectCode);
                          }
                        }
                        if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                          isKeyUp = false;
                          int index = listData.indexOf(
                            listData.firstWhere(
                              (element) => element.code == selectCode,
                            ),
                          );
                          selectCode = listData[index + 1].code;
                          currentListIndex = index + 1;
                          isKeyDown = true;
                          getData(selectCode);
                        }
                      } catch (_) {}
                    }
                  }
                }
              },
              child: ListView(
                controller: listScrollController,
                children: departmentList
                    .map(
                      (value) => listObject(
                        departmentList.indexOf(value),
                        value,
                        showCheckBox,
                      ),
                    )
                    .toList(),
              ),
            ),
          ),
        ],
      ),
    );
  }

  void switchToEdit(DepartmentModel value) {
    setState(() {
      selectCode = value.code;
      getData(selectCode);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      changeScreenEvent(global.ScreenEventEnum.edit);

      if (screenEvent == global.ScreenEventEnum.edit) {
        WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
          tabController.animateTo(1);
          focusFirstTextField(context);
        });
      }
    });
  }

  Color _getContainerColor(String itemCode, int index) {
    if (selectCode.isNotEmpty && selectCode == itemCode) {
      return isSaveAllow
          ? global.theme.rowEditColor
          : global.theme.rowSelectedColor;
    }
    if (_hoverIndex == index) {
      return global.theme.rowHoverColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  Widget listObject(int index, DepartmentModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < departmentGuidListChecked.length; i++) {
      if (departmentGuidListChecked[i] == value.code) {
        isCheck = true;
        break;
      }
    }
    final isSelected = selectCode == value.code;
    TextStyle textStyle = isSelected
        ? TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w700, color: global.theme.textColor)
        : TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w400, color: global.theme.textSecondaryColor);
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
        onTap: () {
          if (showCheckBox == true) {
            setState(() {
              selectCode = value.code;
              if (isCheck == true) {
                departmentGuidListChecked.remove(value.code);
              } else {
                departmentGuidListChecked.add(value.code);
              }
              global.showSnackBar(
                context,
                Icon(Icons.check, color: global.theme.onPrimaryColor),
                "${global.language("chosen")} ${departmentGuidListChecked.length} ${global.language("list")}",
                global.theme.infoHighlightTextColor,
              );
            });
          } else {
            setState(() {
              discardData(
                callBack: () {
                  isSaveAllow = false;
                  changeScreenEvent(global.ScreenEventEnum.list);
                  selectCode = value.code;
                  getData(selectCode);
                  //searchFocusNode.requestFocus();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(1);
                  });
                },
              );
            });
          }
        },
        onDoubleTap: () {
          if (showCheckBox == false) {
            switchToEdit(value);
          }
        },
        child: Container(
          key: index < listKeys.length ? listKeys[index] : null,
          decoration: BoxDecoration(
            color: _getContainerColor(value.code, index),
          ),
          padding: EdgeInsets.only(
            left: 10,
            right: 10,
            top: global.deviceConfig.listDataLineSpace,
            bottom: global.deviceConfig.listDataLineSpace,
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                flex: 5,
                child: Text(
                  value.code,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: textStyle,
                ),
              ),
              Expanded(
                flex: 10,
                child: Text(
                  global.packName(value.names),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: textStyle,
                ),
              ),
              if (showCheckBox)
                Expanded(
                  flex: 1,
                  child: (isCheck)
                      ? Icon(
                          Icons.check,
                          size: global.deviceConfig.listDataFontSize,
                        )
                      : Container(),
                ),
            ],
          ),
        ),
      ),
    );
  }

  List<LanguageDataModel> packLanguage() {
    List<LanguageDataModel> names = [];
    for (int i = 0; i < languageList.length; i++) {
      if (languageList[i].code!.trim().isNotEmpty &&
          fieldTextController[i + 1].text.trim().isNotEmpty) {
        names.add(
          LanguageDataModel(
            code: languageList[i].code!,
            name: fieldTextController[i + 1].text,
          ),
        );
      }
    }
    return names;
  }

  bool verifyData(DepartmentModel value) {
    List<String> errorList = [];
    if (value.code.isEmpty) {
      errorList.add(global.language("code"));
    }
    if (value.names.isEmpty || value.names[0].name.isEmpty) {
      errorList.add(global.language("name"));
    }
    if (errorList.isNotEmpty) {
      global.showSnackBar(
        context,
        Icon(Icons.error, color: global.theme.onPrimaryColor),
        "${global.language("not_success_save")} ${errorList.join(", ")}",
        global.theme.negativeHighlightTextColor,
      );
      return false;
    }
    return true;
  }

  void saveOrUpdateData() {
    if (screenEvent == global.ScreenEventEnum.edit ||
        screenEvent == global.ScreenEventEnum.add) {
      DepartmentModel departmentModel = DepartmentModel(
        guidfixed: "",
        code: fieldTextController[0].text,
        names: packLanguage(),
      );

      DepartmentModel findData = screenData.departments.firstWhere(
        (element) => element.code == departmentModel.code,
        orElse: () => DepartmentModel(code: ""),
      );
      if (findData.code.isNotEmpty) {
        screenData.departments.remove(findData);
        screenData.departments.add(departmentModel);
      } else {
        screenData.departments.add(departmentModel);
      }

      if (verifyData(departmentModel)) {
        showCheckBox = false;
        context.read<CompanyBranchBloc>().add(
          CompanyBranchUpdate(
            guid: selectCompanyBranchGuid,
            companyBranch: screenData,
          ),
        );
      }
    }
  }

  void deleteData(String code) {
    DepartmentModel findData = screenData.departments.firstWhere(
      (element) => element.code == code,
      orElse: () => DepartmentModel(code: ""),
    );
    if (findData.code.isNotEmpty) {
      screenData.departments.remove(findData);
      showCheckBox = false;
      context.read<CompanyBranchBloc>().add(
        CompanyBranchUpdate(
          guid: selectCompanyBranchGuid,
          companyBranch: screenData,
        ),
      );
    }
  }

  void deleteDataMany() {
    for (int i = 0; i < departmentGuidListChecked.length; i++) {
      DepartmentModel findData = screenData.departments.firstWhere(
        (element) => element.code == departmentGuidListChecked[i],
        orElse: () => DepartmentModel(code: ""),
      );
      if (findData.code.isNotEmpty) {
        screenData.departments.remove(findData);
      }
    }
    departmentGuidListChecked = [];
    showCheckBox = false;
    context.read<CompanyBranchBloc>().add(
      CompanyBranchUpdate(
        guid: selectCompanyBranchGuid,
        companyBranch: screenData,
      ),
    );
  }

  void getDataToEditScreen(DepartmentModel departmentModel) {
    isChange = false;
    selectCode = departmentModel.code;
    fieldTextController[0].text = departmentModel.code;
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i + 1].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < departmentModel.names.length; j++) {
        if (languageList[i].code! == departmentModel.names[j].code) {
          fieldTextController[i + 1].text = departmentModel.names[j].name;
        }
      }
    }
  }

  void findFocusNext(int index) {
    // print("findFocusNext($index)");
    focusNodeIndex = index + 1;
    if (focusNodeIndex > fieldFocusNodes.length - 1) {
      focusNodeIndex = 0;
    }
    while (true) {
      if (fieldFocusNodes[focusNodeIndex].isReadOnly == true) {
        focusNodeIndex++;
        if (focusNodeIndex > fieldFocusNodes.length - 1) {
          break;
        }
      } else {
        focusAndCursorToEnd(fieldFocusNodes[focusNodeIndex].focusNode);
        break;
      }
    }
  }

  Widget editScreen({mobileScreen}) {
    return Scaffold(
      backgroundColor: global.theme.cardColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor:
            (screenEvent == global.ScreenEventEnum.edit ||
                screenEvent == global.ScreenEventEnum.add)
            ? global.theme.toolBarEditModeColor
            : global.theme.appBarColor,
        automaticallyImplyLeading: false,
        leading: mobileScreen
            ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(Icons.arrow_back),
                onPressed: () async {
                  showCheckBox = false;
                  discardData(
                    callBack: () {
                      changeScreenEvent(global.ScreenEventEnum.list);
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                    },
                  );
                },
              )
            : null,
        title: Text(headerEdit + global.language("department")),
        actions: <Widget>[
          EditFontSizeControl(onChanged: () => setState(() {})),
          const SizedBox(width: 8),
          if (selectCode.isNotEmpty)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                showCheckBox = false;
                showDialog<String>(
                  context: context,
                  builder: (BuildContext context) => AlertDialog(
                    title: Text(global.language('delete_confirm')),
                    actions: <Widget>[
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.buttonNoColor,
                        ),
                        onPressed: () => Navigator.pop(context),
                        child: Text(global.language('no')),
                      ),
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.buttonYesColor,
                        ),
                        onPressed: () {
                          Navigator.pop(context);
                          deleteData(selectCode);
                          // context.read<DepartmentBloc>().add(
                          //     DepartmentDelete(guid: selectCode));
                        },
                        child: Text(global.language('confirm')),
                      ),
                    ],
                  ),
                );
                setState(() {});
              },
              icon: Icon(Icons.delete),
            ),
          if (isSaveAllow == false && selectCode.trim().isNotEmpty)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                showCheckBox = false;
                switchToEdit(
                  listData[listData.indexOf(
                    listData.firstWhere(
                      (element) => element.code == selectCode,
                    ),
                  )],
                );
              },
              icon: Icon(Icons.edit),
            ),
          if (isSaveAllow == true)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () => saveOrUpdateData(),
              icon: Icon(Icons.save),
            ),
        ],
      ),
      body: Focus(
            skipTraversal: true,
            onKeyEvent: (node, event) {
              if (event is KeyUpEvent) return KeyEventResult.ignored;
              if (event.logicalKey == LogicalKeyboardKey.f10) {
                if (event is KeyDownEvent) saveOrUpdateData();
                return KeyEventResult.handled;
              }
              if (event.logicalKey == LogicalKeyboardKey.enter) {
                if (event is KeyDownEvent) {
                  FocusManager.instance.primaryFocus?.nextFocus();
                }
                return KeyEventResult.handled;
              }
              return KeyEventResult.ignored;
            },
            child: FocusTraversalGroup(
              policy: TextFieldTraversalPolicy(),
              child: Builder(
              builder: (context) {
                final scale = global.editFontScaleFactor;
                return MediaQuery(
                  data: MediaQuery.of(context).copyWith(
                    textScaler: TextScaler.linear(scale),
                  ),
                  child: Theme(
                    data: Theme.of(context).copyWith(
                      inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
                        contentPadding: EdgeInsets.fromLTRB(
                          12 * scale, 20 * scale, 12 * scale, 12 * scale,
                        ),
                      ),
                    ),
                    child: IconTheme(
                      data: IconTheme.of(context).copyWith(
                        size: 24.0 * scale,
                      ),
                      child: SingleChildScrollView(
                        controller: editScrollController,
                        child: Container(
                          color: global.theme.cardColor,
                width: double.infinity,
                padding: const EdgeInsets.all(10),
                child: Column(
                  children: [
                      const SizedBox(height: 10),
                      TextFormField(
                        readOnly: fieldFocusNodes[0].isReadOnly,
                        onFieldSubmitted: (value) {
                          findFocusNext(0);
                        },
                        textInputAction: TextInputAction.next,
                        focusNode: fieldFocusNodes[0].focusNode,
                        textAlign: TextAlign.left,
                        controller: fieldTextController[0],
                        textCapitalization: TextCapitalization.characters,
                        inputFormatters: <TextInputFormatter>[
                          FilteringTextInputFormatter.allow(
                            RegExp('[a-z A-Z 0-9 -]'),
                          ),
                        ],
                        onChanged: (value) {
                          isChange = true;
                          fieldTextController[0].value = TextEditingValue(
                            text: value.toUpperCase(),
                            selection: fieldTextController[0].selection,
                          );
                        },
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                      contentPadding: EdgeInsets.only(
                        left: 10,
                        top: 0,
                        bottom: 0,
                        right: 10,
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
                      ),
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      labelText: global.language("department_code"),
                      labelStyle: TextStyle(
                        color: global.theme.inputTextBoxForceColor,
                      ),
                    ),
                  ),
                  const SizedBox(height: 15),
                  for (int i = 0; i < languageList.length; i++)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 15),
                      child: TextFormField(
                        readOnly: fieldFocusNodes[i + 1].isReadOnly,
                        onChanged: (value) {
                          isChange = true;
                        },
                        onFieldSubmitted: (value) {
                          findFocusNext(i + 1);
                        },
                        textInputAction: TextInputAction.next,
                        focusNode: fieldFocusNodes[i + 1].focusNode,
                        textAlign: TextAlign.left,
                        controller: fieldTextController[i + 1],
                        decoration: InputDecoration(
                          contentPadding: EdgeInsets.only(
                            left: 10,
                            top: 0,
                            bottom: 0,
                            right: 10,
                          ),
                          enabledBorder: OutlineInputBorder(
                            borderSide: BorderSide(
                              color: global.theme.dividerBorderColor,
                              width: 0.0,
                            ),
                          ),
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                          labelStyle: TextStyle(
                            color: global.theme.inputTextBoxColor,
                          ),
                          border: const OutlineInputBorder(),
                          labelText:
                              "${global.language("department_name")} (${languageList[i].name})",
                        ),
                      ),
                    ),
                  if (isSaveAllow)
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.buttonColor,
                        ),
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () {
                          saveOrUpdateData();
                        },
                        icon: Icon(Icons.save),
                        label: Text(
                          global.language("save") +
                              ((kIsWeb ||
                                      Platform.isWindows ||
                                      Platform.isLinux ||
                                      Platform.isMacOS)
                                  ? " (F10)"
                                  : ""),
                        ),
                      ),
                    ),
                    ],
                  ),
                ),
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
          ),
    );
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    if (showCheckBox == false) {
      departmentGuidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<CompanyBranchBloc, CompanyBranchState>(
                listener: (context, state) {
                  // Auto-select first branch
                  if (state is CompanyBranchLoadSuccess) {
                    if (state.companyBranch.isNotEmpty &&
                        selectCompanyBranchGuid.isEmpty) {
                      final firstBranch = state.companyBranch.first;
                      setState(() {
                        selectCompanyBranchGuid = firstBranch.guidfixed;
                        selectCompanyBranchCode = firstBranch.code;
                        selectCompanyBranchName = firstBranch.names;
                      });
                      context.read<CompanyBranchBloc>().add(
                        CompanyBranchGet(guid: selectCompanyBranchGuid),
                      );
                    }
                  }
                  // Load
                  if (state is CompanyBranchGetSuccess) {
                    setState(() {
                      screenData = state.companyBranch;
                      if (screenData.departments.isNotEmpty) {
                        departmentList = [];
                        for (
                          int i = 0;
                          i < screenData.departments.length;
                          i++
                        ) {
                          departmentList.add(
                            DepartmentModel(
                              guidfixed: screenData.departments[i].guidfixed,
                              code: screenData.departments[i].code,
                              names: screenData.departments[i].names,
                            ),
                          );
                        }
                      }

                      // print(screenData.toJson());
                    });
                  }
                  if (state is CompanyBranchUpdateSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: global.theme.onPrimaryColor),
                        global.language("save_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      clearEditData();
                      isSaveAllow = false;
                      getData(selectCode);
                      departmentList = [];

                      context.read<CompanyBranchBloc>().add(
                        CompanyBranchGet(guid: selectCompanyBranchGuid),
                      );
                    });
                  }
                },
              ),
            ],
            child: (constraints.maxWidth > 800)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 14,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: global.theme.infoHighlightTextColor,
                    viewMode: SplitViewMode.Horizontal,
                    indicator: const SplitIndicator(
                      viewMode: SplitViewMode.Horizontal,
                    ),
                    activeIndicator: const SplitIndicator(
                      viewMode: SplitViewMode.Horizontal,
                      isActive: true,
                    ),
                    children: [
                      listScreen(mobileScreen: false),
                      editScreen(mobileScreen: false),
                    ],
                  )
                : TabBarView(
                    physics: const NeverScrollableScrollPhysics(),
                    controller: tabController,
                    children: [
                      listScreen(mobileScreen: true),
                      editScreen(mobileScreen: true),
                    ],
                  ),
          );
        },
      ),
    );
  }
}
