import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/staff/staff_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/staff_model.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class StaffScreen extends StatefulWidget {
  const StaffScreen({super.key});

  @override
  State<StaffScreen> createState() => StaffScreenState();
}

class StaffScreenState extends State<StaffScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<FocusNode> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<File> imageFile = [];
  List<StaffModel> listDatas = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<TextEditingController> fieldTextController = [];
  List<GlobalKey> listKeys = [];
  late SplitViewController splitViewController;
  String searchText = "";
  String selectGuid = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late StaffState blocStaffState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  bool isEditMode = false;
  late StaffModel screenData;
  final _debouncer = global.Debouncer(1000);

  void setSystemLanguageList() async {
    clearEditData();
    await global.setSystemLanguage(context);
    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }

    for (int i = 0; i < languageList.length; i++) {
      fieldTextController.add(TextEditingController());
    }

    loadDataList("");
  }

  @override
  void initState() {
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    clearEditData();
    setSystemLanguageList();
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    // เรียงลำดับ Focus
    // Focus รหัส
    for (var i = 0; i < 100; i++) {
      fieldFocusNodes.add(FocusNode());
    }

    listScrollController.addListener(onScrollList);

    super.initState();
  }

  void loadDataList(String search) {
    context.read<StaffBloc>().add(
      StaffLoadList(
        offset: (listDatas.isEmpty) ? 0 : listDatas.length,
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

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    for (var i = 0; i < 100; i++) {
      fieldFocusNodes[i].dispose();
    }

    super.dispose();
  }

  void clearEditData() {
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].clear();
    }
    screenData = StaffModel(
      guidfixed: "",
      code: "",
      email: "",
      names: [],
      cashier: false,
      order: false,
    );
    isChange = false;
    focusNodeIndex = 0;
  }

  void discardData({required Function callBack}) {
    if (isEditMode && isChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.infoHighlightTextColor),
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
    isEditMode = false;
    context.read<StaffBloc>().add(StaffGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('Staff')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                Navigator.pop(context);
                isEditMode = false;
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
                      guidListChecked.clear();
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
            icon: Icon(Icons.check_box),
          ),
          if (guidListChecked.isNotEmpty)
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
                          backgroundColor: global.theme.negativeHighlightTextColor,
                        ),
                        onPressed: () => Navigator.pop(context),
                        child: Text(global.language('no')),
                      ),
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.infoHighlightTextColor,
                        ),
                        onPressed: () {
                          Navigator.pop(context);
                          context.read<StaffBloc>().add(
                            StaffDeleteMany(guid: guidListChecked),
                          );
                        },
                        child: Text(global.language('delete')),
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
                    isEditMode = true;
                    selectGuid = "";
                    showCheckBox = false;
                    isChange = false;
                    clearEditData();
                    headerEdit = global.language("append");
                    isSaveAllow = true;
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(1);
                      fieldFocusNodes[0].requestFocus();
                    });
                  });
                },
              );
            },
            icon: Icon(Icons.add),
          ),
        ],
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = listDatas.indexOf(
                  listDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = listDatas[index - 1].guidfixed;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  getData(selectGuid);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = listDatas.indexOf(
                  listDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = listDatas[index + 1].guidfixed;
                currentListIndex = index + 1;
                isKeyDown = true;
                getData(selectGuid);
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            Container(
              padding: EdgeInsets.all(5),
              color: global.theme.appBarColor,
              child: Container(
                padding: EdgeInsets.all(5),
                color: global.theme.searchBarColor,
                child: Padding(
                  padding: EdgeInsets.only(left: 10, right: 10),
                  child: TextFormField(
                    onFieldSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      _debouncer.run(() {
                        setState(() {
                          listDatas = [];
                        });
                        loadDataList(value);
                      });
                    },
                    autofocus: false,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.only(
                        top: 10,
                        bottom: 10,
                      ),
                      border: InputBorder.none,
                      hintText: (kIsWeb)
                          ? "${global.language('search')} (F2)"
                          : global.language('search'),
                    ),
                  ),
                ),
              ),
            ),
            Container(color: global.theme.appBarColor, height: 2),
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
                      global.language("Staff_code"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("staff_name"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  if (showCheckBox)
                    Expanded(
                      flex: 1,
                      child: Icon(Icons.check, color: global.theme.textColor, size: 12),
                    ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: listDatas
                      .asMap()
                      .entries
                      .map(
                        (entry) =>
                            listObject(entry.key, entry.value, showCheckBox),
                      )
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
      return isSaveAllow
          ? global.theme.rowEditColor
          : global.theme.rowSelectedColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  void switchToEdit(StaffModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listObject(int index, StaffModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    return GestureDetector(
      onTap: () {
        if (showCheckBox == true) {
          setState(() {
            selectGuid = value.guidfixed;
            if (isCheck == true) {
              guidListChecked.remove(value.guidfixed);
            } else {
              guidListChecked.add(value.guidfixed);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: global.theme.onPrimaryColor),
              "${global.language("chosen")} ${guidListChecked.length} ${global.language("list")}",
              global.theme.infoHighlightTextColor,
            );
          });
        } else {
          setState(() {
            discardData(
              callBack: () {
                isSaveAllow = false;
                isEditMode = false;
                selectGuid = value.guidfixed;
                getData(selectGuid);
                searchFocusNode.requestFocus();
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
              flex: 10,
              child: Text(
                global.packName(value.names),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            if (showCheckBox)
              Expanded(
                flex: 1,
                child: (isCheck)
                    ? Icon(Icons.check, size: 12)
                    : Container(),
              ),
          ],
        ),
      ),
    );
  }

  List<LanguageDataModel> packLanguage() {
    List<LanguageDataModel> names = [];
    for (int i = 0; i < languageList.length; i++) {
      if (languageList[i].code!.trim().isNotEmpty &&
          fieldTextController[i].text.trim().isNotEmpty) {
        names.add(
          LanguageDataModel(
            code: languageList[i].code!,
            name: fieldTextController[i].text,
          ),
        );
      }
    }
    return names;
  }

  void saveOrUpdateData() {
    showCheckBox = false;
    screenData.names = packLanguage();
    if (selectGuid.trim().isEmpty) {
      context.read<StaffBloc>().add(StaffSave(staffModel: screenData));
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    showCheckBox = false;
    context.read<StaffBloc>().add(
      StaffUpdate(guid: guid, staffModel: screenData),
    );
  }

  void getDataToEditScreen(StaffModel staff) {
    isChange = false;
    selectGuid = staff.guidfixed;
    screenData = staff;

    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < staff.names.length; j++) {
        if (languageList[i].code! == staff.names[j].code) {
          fieldTextController[i].text = staff.names[j].name;
        }
      }
    }
  }

  void findFocusNext(int index) {
    focusNodeIndex = index + 1;
    if (focusNodeIndex > fieldFocusNodes.length - 1) {
      focusNodeIndex = 0;
    }
    fieldFocusNodes[focusNodeIndex].requestFocus();
  }

  Widget editScreen({mobileScreen}) {
    int nodeIndex = 0;
    List<Widget> formWidgets = [];
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            findFocusNext(0);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[nodeIndex],
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.code),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isChange = true;
            screenData.code = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("Staff_code"),
          ),
        ),
      ),
    );
    nodeIndex++;
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[nodeIndex],
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.email),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isChange = true;
            screenData.email = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("email"),
          ),
        ),
      ),
    );
    nodeIndex++;
    for (int i = 0; i < languageList.length; i++) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(bottom: 10),
          child: TextFormField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isChange = true;
            },
            onFieldSubmitted: (value) {
              findFocusNext(focusNodeIndex);
            },
            textInputAction: TextInputAction.next,
            focusNode: fieldFocusNodes[nodeIndex],
            textAlign: TextAlign.left,
            controller: fieldTextController[i],
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("staff_name")} (${languageList[i].name})",
            ),
          ),
        ),
      );
      nodeIndex++;
    }
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(bottom: 10),
        child: Row(
          children: <Widget>[
            Checkbox(
              value: screenData.cashier,
              onChanged: (bool? newValue) {
                setState(() {
                  screenData.cashier = newValue!;
                });
              },
            ),
            Text(global.language("cashier")),
          ],
        ),
      ),
    );
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(bottom: 10),
        child: Row(
          children: <Widget>[
            Checkbox(
              value: screenData.order,
              onChanged: (bool? newValue) {
                setState(() {
                  screenData.order = newValue!;
                });
              },
            ),
            Text(global.language("order")),
          ],
        ),
      ),
    );
    if (isSaveAllow) {
      formWidgets.add(
        SizedBox(
          width: double.infinity,
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              saveOrUpdateData();
            },
            icon: Icon(Icons.save),
            label: Text(global.language("save") + ((kIsWeb) ? " (F10)" : "")),
          ),
        ),
      );
    }

    return Scaffold(
      backgroundColor: global.theme.cardColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: (isEditMode)
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
                      isEditMode = false;
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                    },
                  );
                },
              )
            : null,
        title: Text(headerEdit + global.language("Staff")),
        actions: <Widget>[
          EditFontSizeControl(onChanged: () => setState(() {})),
          const SizedBox(width: 8),
          if (selectGuid.isNotEmpty)
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
                          backgroundColor: global.theme.negativeHighlightTextColor,
                        ),
                        onPressed: () => Navigator.pop(context),
                        child: Text(global.language('no')),
                      ),
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.infoHighlightTextColor,
                        ),
                        onPressed: () {
                          Navigator.pop(context);
                          context.read<StaffBloc>().add(
                            StaffDelete(guid: selectGuid),
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
          if (isEditMode && global.systemLanguage.length > 1)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () async {
                setState(() {});
              },
              icon: Icon(Icons.translate),
            ),
          if (isSaveAllow == false && selectGuid.trim().isNotEmpty)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                showCheckBox = false;
                switchToEdit(
                  listDatas[listDatas.indexOf(
                    listDatas.firstWhere(
                      (element) => element.guidfixed == selectGuid,
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
        focusNode: FocusNode(skipTraversal: true),
        onKey: (node, event) {
          if (kIsWeb) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.f2) {
                searchFocusNode.requestFocus();
              }
              if (event.logicalKey == LogicalKeyboardKey.f10) {
                saveOrUpdateData();
              }
            }
          }
          return KeyEventResult.ignored;
        },
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
                      width: double.infinity,
                      padding: const EdgeInsets.all(10),
                      child: Column(children: formWidgets),
                    ),
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    if (showCheckBox == false) {
      guidListChecked.clear();
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<StaffBloc, StaffState>(
            listener: (context, state) {
              blocStaffState = state;
              // Load
              if (state is StaffLoadSuccess) {
                setState(() {
                  if (state.staffs.isNotEmpty) {
                    listDatas.addAll(state.staffs);
                    // สร้าง keys ใหม่ให้ตรงกับจำนวน items
                    listKeys.clear();
                    for (int i = 0; i < listDatas.length; i++) {
                      listKeys.add(GlobalKey());
                    }
                  }
                });
              }
              // Save
              if (state is StaffSaveSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: global.theme.onPrimaryColor),
                    global.language("save_success"),
                    global.theme.infoHighlightTextColor,
                  );
                  clearEditData();
                  listDatas.clear();
                  loadDataList(searchText);
                });
              }
              if (state is StaffSaveFailed) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: global.theme.onPrimaryColor),
                    "${global.language("not_success_save")} : ${state.message}",
                    global.theme.negativeHighlightTextColor,
                  );
                });
              }
              // Update
              if (state is StaffUpdateSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: global.theme.onPrimaryColor),
                    global.language("edit_success"),
                    global.theme.infoHighlightTextColor,
                  );
                  clearEditData();
                  listDatas.clear();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                  isSaveAllow = false;
                  getData(selectGuid);
                });
              }
              if (state is StaffUpdateFailed) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: global.theme.onPrimaryColor),
                    "${global.language("not_edit_success")} : ${state.message}",
                    global.theme.negativeHighlightTextColor,
                  );
                });
              }
              // Delete
              if (state is StaffDeleteSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: global.theme.onPrimaryColor),
                    global.language("delete_success"),
                    global.theme.infoHighlightTextColor,
                  );
                  listDatas.clear();
                  clearEditData();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                });
              }
              // Delete Many
              if (state is StaffDeleteManySuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: global.theme.onPrimaryColor),
                    global.language("not_delete_success"),
                    global.theme.infoHighlightTextColor,
                  );
                  listDatas.clear();
                  clearEditData();
                  loadDataList(searchText);
                  showCheckBox = false;
                });
              }
              // Get
              if (state is StaffGetSuccess) {
                setState(() {
                  getDataToEditScreen(state.staff);
                  if (isEditMode) {
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
                  Offset? positionheader = boxHeader?.localToGlobal(
                    Offset.zero,
                  );
                  RenderBox? box =
                      listKeys[currentListIndex].currentContext
                              ?.findRenderObject()
                          as RenderBox?;
                  Offset? position = box?.localToGlobal(Offset.zero);
                  if (position != null &&
                      positionheader != null &&
                      boxHeader != null &&
                      box != null) {
                    // Scroll Up
                    if (isKeyUp &&
                        position.dy <=
                            (positionheader.dy +
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
                    gripSize: 8,
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
