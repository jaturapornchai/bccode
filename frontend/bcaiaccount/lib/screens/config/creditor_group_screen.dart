import 'dart:io';
import 'package:smlaicloud/bloc/creditor_group/creditor_group_bloc.dart';
import 'package:smlaicloud/model/creditor_group_model.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/widgets/manual_button.dart';

class CreditorGroupScreen extends StatefulWidget {
  const CreditorGroupScreen({super.key});

  @override
  State<CreditorGroupScreen> createState() => CreditorGroupScreenState();
}

class CreditorGroupScreenState extends State<CreditorGroupScreen>
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
  List<CreditorGroupModel> listData = [];
  CreditorGroupModel dataTemp = CreditorGroupModel(
    guidfixed: '',
    groupcode: '',
  );
  List<String> creditorGroupGuidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late CreditorGroupState blocCreditorGroupState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);

  bool isLoadTranslation = false;

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
    loadDataList("");
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
    listScrollController.addListener(onScrollList);

    super.initState();
  }

  void loadDataList(String search) {
    context.read<CreditorGroupBloc>().add(
      CreditorGroupLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
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
    fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
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

  void getData(String guid) {
    headerEdit = global.language("show");
    changeScreenEvent(global.ScreenEventEnum.list);
    context.read<CreditorGroupBloc>().add(CreditorGroupGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('creditor_group')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                Navigator.pop(context);
                changeScreenEvent(global.ScreenEventEnum.list);
              },
            );
          },
        ),
        actions: <Widget>[
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                discardData(
                  callBack: () {
                    setState(() {
                      if (showCheckBox) {
                        showCheckBox = false;
                        creditorGroupGuidListChecked.clear();
                      } else {
                        showCheckBox = true;
                        global.showSnackBar(
                          context,
                          Icon(Icons.delete, color: Colors.white),
                          global.language("choose_item_delete"),
                          Colors.blue,
                        );
                      }
                    });
                  },
                );
              },
              icon: (showCheckBox)
                  ? const Icon(Icons.close)
                  : const Icon(Icons.check_box),
            ),
          ),
          if (creditorGroupGuidListChecked.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
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
                            context.read<CreditorGroupBloc>().add(
                              CreditorGroupDeleteMany(
                                guid: creditorGroupGuidListChecked,
                              ),
                            );
                          },
                          child: Text(global.language('confirm')),
                        ),
                      ],
                    ),
                  );
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
              ),
            ),
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                discardData(
                  callBack: () {
                    setState(() {
                      changeScreenEvent(global.ScreenEventEnum.add);
                      selectGuid = "";
                      showCheckBox = false;
                      isChange = false;
                      clearEditData();
                      headerEdit = global.language("append");
                      isSaveAllow = true;
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(1);
                        fieldFocusNodes[0].focusNode.requestFocus();
                      });
                    });
                  },
                );
              },
              icon: const Icon(Icons.add),
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(5),
            decoration: BoxDecoration(
              color: global.theme.searchBarColor,
              borderRadius: BorderRadius.circular(2),
            ),
            child: Row(
              children: [
                Expanded(
                  child: TextFormField(
                    onFieldSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      debouncer.run(() {
                        setState(() {
                          listData = [];
                        });
                        loadDataList(value);
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
                if (listData.isNotEmpty)
                  Text(
                    '(${listData.length})',
                    style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor),
                  ),
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
                    global.language("creditor_code"),
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
                    global.language("creditor_name"),
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
                              (element) => element.guidfixed == selectGuid,
                            ),
                          );
                          if (index > 0) {
                            selectGuid = listData[index - 1].guidfixed;
                            currentListIndex = index - 1;
                            isKeyUp = true;
                            getData(selectGuid);
                          }
                        }
                        if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                          isKeyUp = false;
                          int index = listData.indexOf(
                            listData.firstWhere(
                              (element) => element.guidfixed == selectGuid,
                            ),
                          );
                          selectGuid = listData[index + 1].guidfixed;
                          currentListIndex = index + 1;
                          isKeyDown = true;
                          getData(selectGuid);
                        }
                      } catch (_) {}
                    }
                  }
                }
              },
              child: ListView(
                controller: listScrollController,
                children: listData
                    .map(
                      (value) => listObject(
                        listData.indexOf(value),
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

  void switchToEdit(CreditorGroupModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      changeScreenEvent(global.ScreenEventEnum.edit);
    });
  }

  Widget listObject(int index, CreditorGroupModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < creditorGroupGuidListChecked.length; i++) {
      if (creditorGroupGuidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
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
        if (showCheckBox == true) {
          setState(() {
            selectGuid = value.guidfixed;
            if (isCheck == true) {
              creditorGroupGuidListChecked.remove(value.guidfixed);
            } else {
              creditorGroupGuidListChecked.add(value.guidfixed);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              "${global.language("chosen")} ${creditorGroupGuidListChecked.length} ${global.language("list")}",
              Colors.blue,
            );
          });
        } else {
          setState(() {
            discardData(
              callBack: () {
                isSaveAllow = false;
                changeScreenEvent(global.ScreenEventEnum.list);
                selectGuid = value.guidfixed;
                getData(selectGuid);
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
          color: _getContainerColor(value.guidfixed, index),
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
                value.groupcode,
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


  Color _getContainerColor(String itemGuid, int index) {
    if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
      return (screenEvent == global.ScreenEventEnum.edit)
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

  List<LanguageDataModel> packLanguage() {
    List<LanguageDataModel> names = [];
    for (int i = 0; i < languageList.length; i++) {
      if (languageList[i].code!.trim().isNotEmpty) {
        names.add(
          LanguageDataModel(
            code: languageList[i].code!,
            name: fieldTextController[i + 1].text,
          ),
        );
      }
    }

    for (var defaultValueLang in dataTemp.names) {
      LanguageDataModel result = names.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        names.add(defaultValueLang);
      }
    }

    return names;
  }

  bool verifyData(CreditorGroupModel value) {
    List<String> errorList = [];
    if (value.groupcode.isEmpty) {
      errorList.add(global.language("code"));
    }
    if (value.names.isEmpty || value.names[0].name.isEmpty) {
      errorList.add(global.language("name"));
    }
    if (errorList.isNotEmpty) {
      global.showSnackBar(
        context,
        Icon(Icons.error, color: Colors.white),
        "${global.language("not_success_save")} ${errorList.join(", ")}",
        Colors.red,
      );
      return false;
    }
    return true;
  }

  void saveOrUpdateData() {
    if (screenEvent == global.ScreenEventEnum.edit ||
        screenEvent == global.ScreenEventEnum.add) {
      CreditorGroupModel creditorGroupModel = CreditorGroupModel(
        guidfixed: "",
        groupcode: fieldTextController[0].text,
        names: packLanguage(),
      );
      if (verifyData(creditorGroupModel)) {
        showCheckBox = false;
        if (selectGuid.trim().isEmpty) {
          context.read<CreditorGroupBloc>().add(
            CreditorGroupSave(creditorGroup: creditorGroupModel),
          );
        } else {
          showCheckBox = false;
          context.read<CreditorGroupBloc>().add(
            CreditorGroupUpdate(
              guid: selectGuid,
              creditorGroup: creditorGroupModel,
            ),
          );
        }
      }
    }
  }

  void getDataToEditScreen(CreditorGroupModel creditorGroupModel) {
    isChange = false;
    dataTemp = creditorGroupModel;
    selectGuid = creditorGroupModel.guidfixed;
    fieldTextController[0].text = creditorGroupModel.groupcode;
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i + 1].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < creditorGroupModel.names.length; j++) {
        if (languageList[i].code! == creditorGroupModel.names[j].code) {
          fieldTextController[i + 1].text = creditorGroupModel.names[j].name;
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
        fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        fieldTextController[focusNodeIndex]
            .selection = TextSelection.fromPosition(
          TextPosition(offset: fieldTextController[focusNodeIndex].text.length),
        );
        break;
      }
    }
  }

  Widget editScreen({mobileScreen}) {
    return Scaffold(
      backgroundColor: (screenEvent == global.ScreenEventEnum.edit || screenEvent == global.ScreenEventEnum.add) ? global.theme.toolBarEditModeColor.withValues(alpha: 0.05) : global.theme.backgroundColor,
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
                icon: const Icon(Icons.arrow_back),
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
        title: Text(headerEdit + global.language("creditor_group")),
        actions: <Widget>[
          if (selectGuid.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
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
                            context.read<CreditorGroupBloc>().add(
                              CreditorGroupDelete(guid: selectGuid),
                            );
                          },
                          child: Text(global.language('confirm')),
                        ),
                      ],
                    ),
                  );
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
              ),
            ),
          if (isSaveAllow == false && selectGuid.trim().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showCheckBox = false;
                  switchToEdit(
                    listData[listData.indexOf(
                      listData.firstWhere(
                        (element) => element.guidfixed == selectGuid,
                      ),
                    )],
                  );
                },
                icon: const Icon(Icons.edit),
              ),
            ),
          if ((screenEvent == global.ScreenEventEnum.add ||
                  screenEvent == global.ScreenEventEnum.edit) &&
              global.systemLanguage.length > 1)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  setState(() {
                    isLoadTranslation = true;
                  });
                  for (int i = 1; i <= languageList.length; i++) {
                    try {
                      if (fieldTextController[i].text.isEmpty) {
                        var translation = await translator.translate(
                          fieldTextController[1].text,
                          to: languageList[i - 1].codeTranslator!,
                        );
                        fieldTextController[i].text = translation.text;
                      }
                    } catch (e) {
                      if (kDebugMode) {
                        AppLogger.debug(e);
                      }
                    }
                  }
                  setState(() {
                    isLoadTranslation = false;
                  });
                },
                icon: const Icon(Icons.translate),
              ),
            ),
          if (isSaveAllow == true)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () => saveOrUpdateData(),
                icon: const Icon(Icons.save),
              ),
            ),
          const ManualButton(path: 'creditor-group'),
        ],
      ),
      body: RawKeyboardListener(
        focusNode: FocusNode(skipTraversal: true),
        onKey: (event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyUpEvent) {
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                // print("Tab");
              }
              if (event.logicalKey == LogicalKeyboardKey.f10) {
                saveOrUpdateData();
              }
            }
          }
        },
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
                        labelText: global.language("creditor_code"),
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
                                color: global.theme.textSecondaryColor,
                                width: 0.0,
                              ),
                            ),
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                            labelStyle: TextStyle(
                              color: global.theme.inputTextBoxColor,
                            ),
                            border: const OutlineInputBorder(),
                            labelText:
                                "${global.language("creditor_name")} (${languageList[i].name})",
                            suffixIcon: isLoadTranslation
                                ? const Padding(
                                    padding: EdgeInsets.all(8.0),
                                    child: SizedBox(
                                      width: 20,
                                      height: 20,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                      ),
                                    ),
                                  )
                                : null,
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
    );
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    if (showCheckBox == false) {
      creditorGroupGuidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<CreditorGroupBloc, CreditorGroupState>(
            listener: (context, state) {
              blocCreditorGroupState = state;
              // Load
              if (state is CreditorGroupLoadSuccess) {
                setState(() {
                  if (state.creditorGroups.isNotEmpty) {
                    listData.addAll(state.creditorGroups);
                  }
                });
              }
              // Save
              if (state is CreditorGroupSaveSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    global.language("save_success"),
                    global.theme.primaryColor,
                  );
                  clearEditData();
                  listData.clear();
                  loadDataList(searchText);
                });
              }
              if (state is CreditorGroupSaveFailed) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    "${global.language("not_success_save")} : ${state.message}",
                    Colors.red,
                  );
                });
              }
              // Update
              if (state is CreditorGroupUpdateSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: Colors.white),
                    global.language("edit_success"),
                    global.theme.primaryColor,
                  );
                  clearEditData();
                  listData.clear();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                  isSaveAllow = false;
                  getData(selectGuid);
                });
              }
              if (state is CreditorGroupUpdateFailed) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: Colors.white),
                    "${global.language("not_edit_success")} : ${state.message}",
                    Colors.red,
                  );
                });
              }
              // Delete
              if (state is CreditorGroupDeleteSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    global.theme.primaryColor,
                  );
                  listData.clear();
                  clearEditData();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                });
              }
              // Delete Many
              if (state is CreditorGroupDeleteManySuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    global.theme.primaryColor,
                  );
                  listData.clear();
                  clearEditData();
                  loadDataList(searchText);
                  showCheckBox = false;
                });
              }
              // Get
              if (state is CreditorGroupGetSuccess) {
                setState(() {
                  getDataToEditScreen(state.creditorGroup);
                  if (screenEvent == global.ScreenEventEnum.edit) {
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(1);
                    });
                    setState(() {
                      findFocusNext(0);
                    });
                  }
                });
                if (currentListIndex >= 0) {
                  RenderBox? boxHeader =
                      headerKey.currentContext?.findRenderObject()
                          as RenderBox?;
                  Offset? positionHeader = boxHeader?.localToGlobal(
                    Offset.zero,
                  );
                  RenderBox? box =
                      listKeys[currentListIndex].currentContext
                              ?.findRenderObject()
                          as RenderBox?;
                  Offset? position = box?.localToGlobal(Offset.zero);
                  if (position != null &&
                      positionHeader != null &&
                      boxHeader != null &&
                      box != null) {
                    // Scroll Up
                    if (isKeyUp &&
                        position.dy <=
                            (positionHeader.dy +
                                (boxHeader.size.height +
                                    (box.size.height * 2)))) {
                      setState(() {
                        listScrollController.animateTo(
                          listScrollController.offset -
                              (boxHeader.size.height + box.size.height),
                          duration: const Duration(milliseconds: 100),
                          curve: Curves.ease,
                        );
                        isKeyUp = false;
                      });
                    }
                    // Scroll Down
                    if (isKeyDown &&
                        position.dy > (queryData.size.height - 100)) {
                      setState(() {
                        listScrollController.animateTo(
                          listScrollController.offset +
                              (position.dy - (queryData.size.height - 100)),
                          duration: const Duration(milliseconds: 100),
                          curve: Curves.easeOut,
                        );
                        isKeyDown = false;
                      });
                    }
                  }
                }
              }
            },
            child: (constraints.maxWidth > 800)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 14,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: global.theme.primaryColor,
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
