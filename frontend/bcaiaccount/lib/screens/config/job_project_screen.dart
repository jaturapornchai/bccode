import 'dart:io';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/bloc/job_project/job_project_bloc.dart';
import 'package:smlaicloud/model/job_project_model.dart';
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

class JobProjectScreen extends StatefulWidget {
  /// isProjectMode = true → หน้า "โครงการ" (parentcode ว่าง)
  /// isProjectMode = false → หน้า "งาน" (parentcode ต้องมีค่า)
  final bool isProjectMode;

  const JobProjectScreen({super.key, this.isProjectMode = true});

  @override
  State<JobProjectScreen> createState() => JobProjectScreenState();
}

class JobProjectScreenState extends State<JobProjectScreen>
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
  List<JobProjectModel> listData = [];
  // filter โครงการ (ใช้เฉพาะหน้างาน)
  String _filterProjectCode = "";
  // filtered ตาม mode + project filter
  List<JobProjectModel> get _displayList {
    if (widget.isProjectMode) {
      return listData.where((e) => e.parentcode.isEmpty).toList();
    }
    final jobs = listData.where((e) => e.parentcode.isNotEmpty);
    if (_filterProjectCode.isEmpty) return jobs.toList();
    return jobs.where((e) => e.parentcode == _filterProjectCode).toList();
  }

  // รายการโครงการทั้งหมด (สำหรับ filter bar)
  List<JobProjectModel> get _projectList =>
      listData.where((e) => e.parentcode.isEmpty).toList();
  List<String> jobProjectGuidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectCode = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late JobProjectState blocJobProjectState;
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

  // parentcode state: code ที่เลือก + display text ใน field
  String _selectedParentCode = "";
  final TextEditingController _parentDisplayController = TextEditingController();

  void changeScreenEvent(global.ScreenEventEnum event) {
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
    super.initState();

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        context.read<JobProjectBloc>().add(
              const JobProjectLoadList(offset: 0, limit: 999999, search: ""),
            );
      }
    });
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    _parentDisplayController.dispose();
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
    _selectedParentCode = "";
    _parentDisplayController.clear();
    isChange = false;
    focusNodeIndex = 0;
  }

  void discardData({required Function callBack}) {
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

  void getData(String guid) {
    headerEdit = global.language("show");
    changeScreenEvent(global.ScreenEventEnum.list);
    context.read<JobProjectBloc>().add(JobProjectGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(widget.isProjectMode
            ? global.language('project')
            : global.language('job')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
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
                      jobProjectGuidListChecked.clear();
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
          if (jobProjectGuidListChecked.isNotEmpty)
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
                          context.read<JobProjectBloc>().add(
                                JobProjectDeleteMany(
                                    guid: jobProjectGuidListChecked),
                              );
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
            onPressed: () {
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
            },
            icon: Icon(Icons.add),
          ),
          const ManualButton(path: 'settings-job-project'),
        ],
      ),
      body: Column(
        children: [
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
                        context.read<JobProjectBloc>().add(
                              JobProjectLoadList(
                                  offset: 0, limit: 999999, search: value),
                            );
                      });
                    },
                    autofocus: false,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.symmetric(
                          vertical: 8, horizontal: 4),
                      border: InputBorder.none,
                      hintText: global.language('search'),
                      prefixIcon: Icon(Icons.search, size: 20),
                      prefixIconConstraints:
                          const BoxConstraints(minWidth: 32, minHeight: 32),
                    ),
                  ),
                ),
                ListFontSizeControl(onChanged: () => setState(() {})),
                const SizedBox(width: 4),
                if (_displayList.isNotEmpty)
                  Text('(${_displayList.length})',
                      style: TextStyle(
                          fontSize: 11,
                          color: global.theme.textSecondaryColor)),
              ],
            ),
          ),
          // filter bar เลือกโครงการ (เฉพาะหน้างาน)
          if (!widget.isProjectMode) _buildProjectFilterBar(),
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
                    global.language("job_project_code"),
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
                    global.language("job_project_name"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (!widget.isProjectMode)
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("project"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
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
              focusNode:
                  FocusNode(skipTraversal: true, canRequestFocus: true),
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
                        if (event.logicalKey ==
                            LogicalKeyboardKey.arrowUp) {
                          isKeyDown = false;
                          int index = _displayList.indexOf(
                            _displayList.firstWhere(
                              (element) =>
                                  element.guidfixed == selectCode,
                            ),
                          );
                          if (index > 0) {
                            selectCode = _displayList[index - 1].guidfixed;
                            currentListIndex = index - 1;
                            isKeyUp = true;
                            getData(selectCode);
                          }
                        }
                        if (event.logicalKey ==
                            LogicalKeyboardKey.arrowDown) {
                          isKeyUp = false;
                          int index = _displayList.indexOf(
                            _displayList.firstWhere(
                              (element) =>
                                  element.guidfixed == selectCode,
                            ),
                          );
                          selectCode = _displayList[index + 1].guidfixed;
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
                children: _displayList
                    .map(
                      (value) => listObject(
                        _displayList.indexOf(value),
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

  void switchToEdit(JobProjectModel value) {
    setState(() {
      selectCode = value.guidfixed;
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

  Color _getContainerColor(String itemGuid, int index) {
    if (selectCode.isNotEmpty && selectCode == itemGuid) {
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

  Widget listObject(int index, JobProjectModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < jobProjectGuidListChecked.length; i++) {
      if (jobProjectGuidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    final isSelected = selectCode == value.guidfixed;
    TextStyle textStyle = isSelected
        ? TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.w700,
            color: global.theme.textColor)
        : TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.w400,
            color: global.theme.textSecondaryColor);
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
        onTap: () {
          if (showCheckBox == true) {
            setState(() {
              selectCode = value.guidfixed;
              if (isCheck == true) {
                jobProjectGuidListChecked.remove(value.guidfixed);
              } else {
                jobProjectGuidListChecked.add(value.guidfixed);
              }
              global.showSnackBar(
                context,
                Icon(Icons.check, color: global.theme.onPrimaryColor),
                "${global.language("chosen")} ${jobProjectGuidListChecked.length} ${global.language("list")}",
                global.theme.infoHighlightTextColor,
              );
            });
          } else {
            setState(() {
              discardData(
                callBack: () {
                  isSaveAllow = false;
                  changeScreenEvent(global.ScreenEventEnum.list);
                  selectCode = value.guidfixed;
                  getData(selectCode);
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
            color: _getContainerColor(value.guidfixed, index),
          ),
          padding: EdgeInsets.only(
            left: 10,
            right: 10,
            top: global.deviceConfig.listDataLineSpace,
            bottom: global.deviceConfig.listDataLineSpace,
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              // icon + indent: โครงการ vs งาน
              if (value.parentcode.isEmpty)
                Padding(
                  padding: const EdgeInsets.only(right: 6),
                  child: Icon(
                    Icons.account_tree,
                    size: global.deviceConfig.listDataFontSize + 2,
                    color: global.theme.infoHighlightTextColor,
                  ),
                )
              else
                Padding(
                  padding: const EdgeInsets.only(left: 14, right: 6),
                  child: Icon(
                    Icons.subdirectory_arrow_right,
                    size: global.deviceConfig.listDataFontSize + 2,
                    color: global.theme.textSecondaryColor,
                  ),
                ),
              Expanded(
                flex: 5,
                child: Text(
                  value.code,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: value.parentcode.isEmpty
                      ? textStyle.copyWith(fontWeight: FontWeight.w700)
                      : textStyle,
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
              if (!widget.isProjectMode)
                Expanded(
                  flex: 5,
                  child: Text(
                    value.parentcode,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      fontSize: global.deviceConfig.listDataFontSize - 1,
                      color: global.theme.textSecondaryColor,
                    ),
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

  bool verifyData(JobProjectModel value) {
    List<String> errorList = [];
    if (value.code.isEmpty) {
      errorList.add(global.language("code"));
    }
    if (value.names.isEmpty || value.names[0].name.isEmpty) {
      errorList.add(global.language("name"));
    }
    // หน้างาน: ต้องเลือกโครงการหลัก
    if (!widget.isProjectMode && value.parentcode.isEmpty) {
      errorList.add(global.language("parent_project"));
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
      JobProjectModel jobProjectModel = JobProjectModel(
        guidfixed: selectCode,
        code: fieldTextController[0].text,
        parentcode: _selectedParentCode,
        names: packLanguage(),
      );

      if (verifyData(jobProjectModel)) {
        showCheckBox = false;
        if (screenEvent == global.ScreenEventEnum.add) {
          context
              .read<JobProjectBloc>()
              .add(JobProjectSave(jobProject: jobProjectModel));
        } else {
          context.read<JobProjectBloc>().add(
                JobProjectUpdate(
                    guid: selectCode, jobProject: jobProjectModel),
              );
        }
      }
    }
  }

  void getDataToEditScreen(JobProjectModel jobProjectModel) {
    isChange = false;
    selectCode = jobProjectModel.guidfixed;
    fieldTextController[0].text = jobProjectModel.code;
    _selectedParentCode = jobProjectModel.parentcode;
    // หา display text ของ parent จาก listData
    if (jobProjectModel.parentcode.isNotEmpty) {
      final parent = listData.where((e) => e.code == jobProjectModel.parentcode).firstOrNull;
      _parentDisplayController.text = parent != null
          ? "${parent.code}  ${global.packName(parent.names)}"
          : jobProjectModel.parentcode;
    } else {
      _parentDisplayController.clear();
    }
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i + 1].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < jobProjectModel.names.length; j++) {
        if (languageList[i].code! == jobProjectModel.names[j].code) {
          fieldTextController[i + 1].text = jobProjectModel.names[j].name;
        }
      }
    }
  }

  void findFocusNext(int index) {
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

  /// badge แสดงประเภท: โครงการ (ไม่มี parent) หรือ งาน (มี parent)
  /// filter bar แสดง chip โครงการให้คลิกเลือก (เฉพาะหน้างาน)
  Widget _buildProjectFilterBar() {
    final projects = listData.where((e) => e.parentcode.isEmpty).toList();
    if (projects.isEmpty) return const SizedBox.shrink();
    return Container(
      color: global.theme.surfaceColor,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: [
            // chip "ทั้งหมด"
            Padding(
              padding: const EdgeInsets.only(right: 6),
              child: ChoiceChip(
                label: Text(
                  global.language("all"),
                  style: TextStyle(
                    fontSize: 12,
                    color: _filterProjectCode.isEmpty
                        ? global.theme.onPrimaryColor
                        : global.theme.textColor,
                  ),
                ),
                selected: _filterProjectCode.isEmpty,
                selectedColor: global.theme.buttonColor,
                backgroundColor: global.theme.cardColor,
                side: BorderSide(color: global.theme.dividerBorderColor),
                onSelected: (_) => setState(() => _filterProjectCode = ""),
              ),
            ),
            // chip ต่อ project
            ...projects.map((p) {
              final isSelected = _filterProjectCode == p.code;
              return Padding(
                padding: const EdgeInsets.only(right: 6),
                child: ChoiceChip(
                  avatar: Icon(
                    Icons.account_tree,
                    size: 14,
                    color: isSelected
                        ? global.theme.onPrimaryColor
                        : global.theme.infoHighlightTextColor,
                  ),
                  label: Text(
                    "${p.code}  ${global.packName(p.names)}",
                    style: TextStyle(
                      fontSize: 12,
                      color: isSelected
                          ? global.theme.onPrimaryColor
                          : global.theme.textColor,
                    ),
                  ),
                  selected: isSelected,
                  selectedColor: global.theme.buttonColor,
                  backgroundColor: global.theme.cardColor,
                  side: BorderSide(
                    color: isSelected
                        ? global.theme.buttonColor
                        : global.theme.dividerBorderColor,
                  ),
                  onSelected: (_) =>
                      setState(() => _filterProjectCode = p.code),
                ),
              );
            }),
          ],
        ),
      ),
    );
  }

  Widget _buildTypeIndicator() {
    final bool isProject = _selectedParentCode.isEmpty;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: isProject
            ? global.theme.infoHighlightColor
            : global.theme.positiveHighlightColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: isProject
              ? global.theme.infoHighlightTextColor.withValues(alpha: 0.3)
              : global.theme.positiveHighlightTextColor.withValues(alpha: 0.3),
        ),
      ),
      child: Row(
        children: [
          Icon(
            isProject ? Icons.account_tree : Icons.work_outline,
            size: 18,
            color: isProject
                ? global.theme.infoHighlightTextColor
                : global.theme.positiveHighlightTextColor,
          ),
          const SizedBox(width: 8),
          Text(
            isProject
                ? global.language("project")
                : "${global.language("job")} — ${global.language("parent_project")}: $_selectedParentCode",
            style: TextStyle(
              fontWeight: FontWeight.w600,
              fontSize: 13,
              color: isProject
                  ? global.theme.infoHighlightTextColor
                  : global.theme.positiveHighlightTextColor,
            ),
          ),
        ],
      ),
    );
  }

  /// field เลือกโครงการหลัก — tap เพื่อเปิด dialog search
  Widget _buildParentDropdown() {
    final bool isReadOnly = screenEvent == global.ScreenEventEnum.list;
    return TextFormField(
      controller: _parentDisplayController,
      readOnly: true,
      onTap: isReadOnly ? null : _showProjectSelectDialog,
      decoration: InputDecoration(
        border: const OutlineInputBorder(),
        enabledBorder: OutlineInputBorder(
          borderSide: BorderSide(
            color: global.theme.dividerBorderColor,
            width: 0.0,
          ),
        ),
        floatingLabelBehavior: FloatingLabelBehavior.always,
        labelText: global.language("parent_project"),
        labelStyle: TextStyle(color: global.theme.inputTextBoxColor),
        hintText: isReadOnly ? null : global.language("search"),
        hintStyle: TextStyle(color: global.theme.formHintColor),
        prefixIcon: Icon(
          _selectedParentCode.isEmpty ? Icons.account_tree : Icons.folder,
          size: 18,
          color: global.theme.infoHighlightTextColor,
        ),
        suffixIcon: isReadOnly
            ? null
            : Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (_selectedParentCode.isNotEmpty)
                    IconButton(
                      focusNode: FocusNode(skipTraversal: true),
                      icon: Icon(Icons.clear,
                          size: 18, color: global.theme.iconSecondaryColor),
                      tooltip: global.language("clear"),
                      onPressed: () {
                        setState(() {
                          _selectedParentCode = "";
                          _parentDisplayController.clear();
                          isChange = true;
                        });
                      },
                    ),
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: Icon(Icons.search,
                        size: 18, color: global.theme.iconColor),
                    onPressed: _showProjectSelectDialog,
                  ),
                ],
              ),
      ),
    );
  }

  /// dialog ค้นหาและเลือกโครงการหลัก
  void _showProjectSelectDialog() {
    // รายการโครงการ (parentcode ว่าง และไม่ใช่ตัวเอง)
    final allProjects = listData
        .where((e) => e.parentcode.isEmpty && e.guidfixed != selectCode)
        .toList();

    final searchCtrl = TextEditingController();
    List<JobProjectModel> filtered = List.from(allProjects);

    showDialog<JobProjectModel>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          return AlertDialog(
            backgroundColor: global.theme.cardColor,
            contentPadding: EdgeInsets.zero,
            titlePadding: EdgeInsets.zero,
            title: Container(
              padding: const EdgeInsets.fromLTRB(16, 12, 8, 8),
              decoration: BoxDecoration(
                color: global.theme.appBarColor,
                borderRadius: const BorderRadius.vertical(
                    top: Radius.circular(4)),
              ),
              child: Row(
                children: [
                  Icon(Icons.account_tree,
                      color: global.theme.onPrimaryColor, size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      global.language("parent_project"),
                      style: TextStyle(
                          color: global.theme.onPrimaryColor,
                          fontSize: 15),
                    ),
                  ),
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: Icon(Icons.close,
                        color: global.theme.onPrimaryColor, size: 18),
                    onPressed: () => Navigator.pop(ctx),
                  ),
                ],
              ),
            ),
            content: SizedBox(
              width: 420,
              height: 420,
              child: Column(
                children: [
                  // search bar
                  Container(
                    padding: const EdgeInsets.all(8),
                    color: global.theme.searchBarColor,
                    child: TextField(
                      controller: searchCtrl,
                      autofocus: true,
                      style: TextStyle(color: global.theme.textColor),
                      decoration: InputDecoration(
                        isDense: true,
                        contentPadding: const EdgeInsets.symmetric(
                            vertical: 8, horizontal: 4),
                        border: InputBorder.none,
                        hintText: global.language('search'),
                        hintStyle:
                            TextStyle(color: global.theme.formHintColor),
                        prefixIcon: Icon(Icons.search,
                            size: 20, color: global.theme.iconColor),
                        prefixIconConstraints: const BoxConstraints(
                            minWidth: 32, minHeight: 32),
                      ),
                      onChanged: (q) {
                        setDialogState(() {
                          filtered = allProjects
                              .where((e) =>
                                  e.code.toLowerCase().contains(
                                      q.toLowerCase()) ||
                                  global
                                      .packName(e.names)
                                      .toLowerCase()
                                      .contains(q.toLowerCase()))
                              .toList();
                        });
                      },
                    ),
                  ),
                  // header
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 10, vertical: 5),
                    color: global.theme.columnHeaderColor,
                    child: Row(
                      children: [
                        SizedBox(
                          width: 100,
                          child: Text(
                            global.language("job_project_code"),
                            style: TextStyle(
                              color: global.theme.columnHeaderTextColor,
                              fontWeight: FontWeight.bold,
                              fontSize: 12,
                            ),
                          ),
                        ),
                        Expanded(
                          child: Text(
                            global.language("job_project_name"),
                            style: TextStyle(
                              color: global.theme.columnHeaderTextColor,
                              fontWeight: FontWeight.bold,
                              fontSize: 12,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                  // list
                  Expanded(
                    child: filtered.isEmpty
                        ? Center(
                            child: Text(global.language("no_data"),
                                style: TextStyle(
                                    color: global.theme.textSecondaryColor)),
                          )
                        : ListView.builder(
                            itemCount: filtered.length,
                            itemBuilder: (ctx, i) {
                              final item = filtered[i];
                              final isSelected =
                                  item.code == _selectedParentCode;
                              return InkWell(
                                onTap: () => Navigator.pop(ctx, item),
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 10, vertical: 8),
                                  color: isSelected
                                      ? global.theme.rowSelectedColor
                                      : (i % 2 == 0
                                          ? global.theme
                                              .columnAlternateEvenColor
                                          : global.theme
                                              .columnAlternateOddColor),
                                  child: Row(
                                    children: [
                                      SizedBox(
                                        width: 100,
                                        child: Text(
                                          item.code,
                                          style: TextStyle(
                                            fontWeight: isSelected
                                                ? FontWeight.w700
                                                : FontWeight.w400,
                                            color: global.theme.textColor,
                                            fontSize: 13,
                                          ),
                                        ),
                                      ),
                                      Expanded(
                                        child: Text(
                                          global.packName(item.names),
                                          overflow: TextOverflow.ellipsis,
                                          style: TextStyle(
                                            color: global.theme
                                                .textSecondaryColor,
                                            fontSize: 13,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              );
                            },
                          ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    ).then((selected) {
      if (selected != null) {
        setState(() {
          _selectedParentCode = selected.code;
          _parentDisplayController.text =
              "${selected.code}  ${global.packName(selected.names)}";
          isChange = true;
        });
      }
    });
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
                      WidgetsBinding.instance
                          .addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                    },
                  );
                },
              )
            : null,
        title: Text(headerEdit +
            (widget.isProjectMode
                ? global.language("project")
                : global.language("job"))),
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
                          context.read<JobProjectBloc>().add(
                                JobProjectDelete(guid: selectCode),
                              );
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
                  _displayList[_displayList.indexOf(
                    _displayList.firstWhere(
                      (element) => element.guidfixed == selectCode,
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
                    inputDecorationTheme:
                        Theme.of(context).inputDecorationTheme.copyWith(
                              contentPadding: EdgeInsets.fromLTRB(
                                12 * scale,
                                20 * scale,
                                12 * scale,
                                12 * scale,
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
                            // badge แสดงประเภทเฉพาะหน้างาน
                            if (!widget.isProjectMode) ...[
                              _buildTypeIndicator(),
                              const SizedBox(height: 15),
                            ],
                            TextFormField(
                              readOnly: fieldFocusNodes[0].isReadOnly,
                              onFieldSubmitted: (value) {
                                findFocusNext(0);
                              },
                              textInputAction: TextInputAction.next,
                              focusNode: fieldFocusNodes[0].focusNode,
                              textAlign: TextAlign.left,
                              controller: fieldTextController[0],
                              textCapitalization:
                                  TextCapitalization.characters,
                              inputFormatters: <TextInputFormatter>[
                                FilteringTextInputFormatter.allow(
                                  RegExp('[a-z A-Z 0-9 -]'),
                                ),
                              ],
                              onChanged: (value) {
                                isChange = true;
                                fieldTextController[0].value =
                                    TextEditingValue(
                                  text: value.toUpperCase(),
                                  selection:
                                      fieldTextController[0].selection,
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
                                  borderSide: BorderSide(
                                      color:
                                          global.theme.dividerBorderColor,
                                      width: 0.0),
                                ),
                                floatingLabelBehavior:
                                    FloatingLabelBehavior.always,
                                labelText:
                                    global.language("job_project_code"),
                                labelStyle: TextStyle(
                                  color:
                                      global.theme.inputTextBoxForceColor,
                                ),
                              ),
                            ),
                            const SizedBox(height: 15),
                            // dropdown เลือกโครงการหลัก (แสดงเฉพาะหน้างาน)
                            if (!widget.isProjectMode) ...[
                              _buildParentDropdown(),
                              const SizedBox(height: 15),
                            ],
                            for (int i = 0;
                                i < languageList.length;
                                i++)
                              Padding(
                                padding:
                                    const EdgeInsets.only(bottom: 15),
                                child: TextFormField(
                                  readOnly:
                                      fieldFocusNodes[i + 1].isReadOnly,
                                  onChanged: (value) {
                                    isChange = true;
                                  },
                                  onFieldSubmitted: (value) {
                                    findFocusNext(i + 1);
                                  },
                                  textInputAction: TextInputAction.next,
                                  focusNode:
                                      fieldFocusNodes[i + 1].focusNode,
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
                                        color: global
                                            .theme.dividerBorderColor,
                                        width: 0.0,
                                      ),
                                    ),
                                    floatingLabelBehavior:
                                        FloatingLabelBehavior.always,
                                    labelStyle: TextStyle(
                                      color:
                                          global.theme.inputTextBoxColor,
                                    ),
                                    border: const OutlineInputBorder(),
                                    labelText:
                                        "${global.language("job_project_name")} (${languageList[i].name})",
                                  ),
                                ),
                              ),
                            if (isSaveAllow)
                              SizedBox(
                                width: double.infinity,
                                child: ElevatedButton.icon(
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor:
                                        global.theme.buttonColor,
                                  ),
                                  focusNode:
                                      FocusNode(skipTraversal: true),
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
      jobProjectGuidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<JobProjectBloc, JobProjectState>(
                listener: (context, state) {
                  if (state is JobProjectLoadSuccess) {
                    setState(() {
                      listData = state.jobProject;
                    });
                  }
                  if (state is JobProjectGetSuccess) {
                    setState(() {
                      getDataToEditScreen(state.jobProject);
                    });
                  }
                  if (state is JobProjectSaveSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save,
                            color: global.theme.onPrimaryColor),
                        global.language("save_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      clearEditData();
                      isSaveAllow = false;
                      selectCode = state.responsesID;
                      context.read<JobProjectBloc>().add(
                            const JobProjectLoadList(
                                offset: 0, limit: 999999, search: ""),
                          );
                    });
                  }
                  if (state is JobProjectUpdateSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save,
                            color: global.theme.onPrimaryColor),
                        global.language("save_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      clearEditData();
                      isSaveAllow = false;
                      context.read<JobProjectBloc>().add(
                            const JobProjectLoadList(
                                offset: 0, limit: 999999, search: ""),
                          );
                    });
                  }
                  if (state is JobProjectDeleteSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete,
                            color: global.theme.onPrimaryColor),
                        global.language("delete_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      selectCode = "";
                      clearEditData();
                      isSaveAllow = false;
                      changeScreenEvent(global.ScreenEventEnum.list);
                      context.read<JobProjectBloc>().add(
                            const JobProjectLoadList(
                                offset: 0, limit: 999999, search: ""),
                          );
                    });
                  }
                  if (state is JobProjectDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete,
                            color: global.theme.onPrimaryColor),
                        global.language("delete_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      jobProjectGuidListChecked.clear();
                      showCheckBox = false;
                      selectCode = "";
                      clearEditData();
                      isSaveAllow = false;
                      changeScreenEvent(global.ScreenEventEnum.list);
                      context.read<JobProjectBloc>().add(
                            const JobProjectLoadList(
                                offset: 0, limit: 999999, search: ""),
                          );
                    });
                  }
                  if (state is JobProjectSaveFailed) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.error,
                          color: global.theme.onPrimaryColor),
                      "${global.language("not_success_save")} ${state.message}",
                      global.theme.negativeHighlightTextColor,
                    );
                  }
                  if (state is JobProjectUpdateFailed) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.error,
                          color: global.theme.onPrimaryColor),
                      "${global.language("not_success_save")} ${state.message}",
                      global.theme.negativeHighlightTextColor,
                    );
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
