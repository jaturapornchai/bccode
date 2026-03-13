import 'package:smlaicloud/bloc/holiday/holiday_bloc.dart';
import 'package:smlaicloud/model/holiday_model.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:image_picker/image_picker.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class HolidayScreen extends StatefulWidget {
  const HolidayScreen({super.key});

  @override
  State<HolidayScreen> createState() => HolidayScreenState();
}

class HolidayScreenState extends State<HolidayScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  late DropzoneViewController dropZoneController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = [];
  List<TextEditingController> fieldTextController = [];
  List<FocusNode> fieldFocusNodes = [];
  final ImagePicker imagePicker = ImagePicker();
  int focusNodeIndex = 0;
  List<HolidayModel> closeDayListData = [];
  List<String> closeDayGuidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];

  int loaDataPerPage = 1000;
  String searchText = "";
  String selectGuid = "";
  String screenGuid = "";
  bool isChange = false;
  bool isSaveAllow = false;
  DateTime selectedDate = DateTime.now();
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late SplitViewController splitViewController;
  late HolidayModel screenData;
  final _debouncer = global.Debouncer(1000);

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);
    languageList = [];
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
      fieldFocusNodes.add(focusNode);
    }

    FocusNode focusNode1 = FocusNode();
    focusNode1.addListener(() {
      focusNodeIndex = languageList.length + 1;
    });
    fieldFocusNodes.add(focusNode1);
    fieldTextController.add(TextEditingController());
    setState(() {});
    loadDataList("");
  }

  @override
  void initState() {
    fieldTextController.add(TextEditingController());
    tabController = TabController(vsync: this, length: 2);
    screenData = HolidayModel(guidfixed: "", date: "", desc: []);

    tabController.addListener(() {
      setState(() {});
    });

    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    // เรียงลำดับ Focus
    // Focus รหัส

    setSystemLanguageList();

    listScrollController.addListener(onScrollList);

    super.initState();
  }

  void loadDataList(String search) {
    context.read<HolidayBloc>().add(
      HolidayLoadList(
        offset: (closeDayListData.isEmpty) ? 0 : closeDayListData.length,
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
      fieldFocusNodes[i].dispose();
    }

    super.dispose();
  }

  void clearEditData() {
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].clear();
    }
    isChange = false;
    focusNodeIndex = 0;

    screenData = HolidayModel(guidfixed: "", date: "", desc: []);
    selectGuid = "";
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

  void getData(HolidayModel offdayData) {
    headerEdit = global.language("show");
    isEditMode = false;
    selectGuid = offdayData.guidfixed;
    screenData = offdayData;
    fieldTextController[languageList.length].text = offdayData.date;
    for (int i = 0; i < languageList.length; i++) {
      for (int z = 0; z < offdayData.desc.length; z++) {
        if (languageList[i].code! == offdayData.desc[z].code) {
          fieldTextController[i].text = offdayData.desc[z].name;
        }
      }
    }
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('holidays')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
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
                        closeDayGuidListChecked.clear();
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
              icon: const Icon(Icons.check_box),
            ),
          ),
          if (closeDayGuidListChecked.isNotEmpty)
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
                            context.read<HolidayBloc>().add(
                              HolidayDeleteMany(guid: closeDayGuidListChecked),
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
              icon: const Icon(Icons.add),
            ),
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
                int index = closeDayListData.indexOf(
                  closeDayListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );

                if (index > 0) {
                  HolidayModel offdayData = closeDayListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  );
                  selectGuid = closeDayListData[index - 1].guidfixed;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  getData(offdayData);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = closeDayListData.indexOf(
                  closeDayListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                HolidayModel offdayData = closeDayListData.firstWhere(
                  (element) => element.guidfixed == selectGuid,
                );
                selectGuid = closeDayListData[index + 1].guidfixed;
                currentListIndex = index + 1;
                isKeyDown = true;
                getData(offdayData);
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
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
                        _debouncer.run(() {
                          setState(() {
                            closeDayListData = [];
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
                  if (closeDayListData.isNotEmpty)
                    Text(
                      '(${closeDayListData.length})',
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
                    flex: 1,
                    child: Text(
                      global.language("desc"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 1,
                    child: Text(
                      global.language("doc_date"),
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
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: closeDayListData
                      .map(
                        (value) => listObject(
                          closeDayListData.indexOf(value),
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
      ),
    );
  }

  void switchToEdit(HolidayModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(value);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listObject(int index, HolidayModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < closeDayGuidListChecked.length; i++) {
      if (closeDayGuidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
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
              closeDayGuidListChecked.remove(value.guidfixed);
            } else {
              closeDayGuidListChecked.add(value.guidfixed);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              "${global.language("chosen")} ${closeDayGuidListChecked.length} ${global.language("list")}",
              Colors.blue,
            );
          });
        } else {
          setState(() {
            discardData(
              callBack: () {
                isSaveAllow = false;
                isEditMode = false;
                selectGuid = value.guidfixed;
                getData(value);
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
        color: _getContainerColor(value.guidfixed, index),
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
              flex: 1,
              child: Builder(builder: (context) {
                final isSelected = selectGuid.isNotEmpty && selectGuid == value.guidfixed;
                return Text(
                  global.packName(value.desc),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontSize: global.deviceConfig.listDataFontSize,
                    fontWeight: isSelected ? FontWeight.w700 : FontWeight.w400,
                    color: isSelected ? global.theme.textColor : global.theme.textSecondaryColor,
                  ),
                );
              }),
            ),
            Expanded(
              flex: 1,
              child: Builder(builder: (context) {
                final isSelected = selectGuid.isNotEmpty && selectGuid == value.guidfixed;
                return Text(
                  value.date,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontSize: global.deviceConfig.listDataFontSize,
                    fontWeight: isSelected ? FontWeight.w700 : FontWeight.w400,
                    color: isSelected ? global.theme.textColor : global.theme.textSecondaryColor,
                  ),
                );
              }),
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
      return isEditMode
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
    if (selectGuid.trim().isEmpty) {
      HolidayModel offdayModel = HolidayModel(
        guidfixed: "",
        date: fieldTextController[languageList.length].text,
        desc: packLanguage(),
      );

      if (offdayModel.desc.isNotEmpty && offdayModel.date != "") {
        context.read<HolidayBloc>().add(HolidaySave(holidayModel: offdayModel));
      } else {
        global.showSnackBar(
          context,
          Icon(Icons.edit, color: Colors.white),
          global.language("empty_data"),
          Colors.red,
        );
      }
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    showCheckBox = false;

    HolidayModel offdayModel = HolidayModel(
      guidfixed: screenData.guidfixed,
      date: fieldTextController[languageList.length].text,
      desc: packLanguage(),
    );
    if (offdayModel.desc.isNotEmpty && offdayModel.date != "") {
      context.read<HolidayBloc>().add(
        HolidayUpdate(guid: guid, holidayModel: offdayModel),
      );
    } else {
      global.showSnackBar(
        context,
        Icon(Icons.edit, color: Colors.white),
        global.language("empty_data"),
        Colors.red,
      );
    }
  }

  void getDataToEditScreen(HolidayModel offday) {
    isChange = false;
    selectGuid = offday.guidfixed;

    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < offday.desc.length; j++) {
        if (languageList[i].code! == offday.desc[j].code) {
          fieldTextController[i].text = offday.desc[j].name;
        }
      }
    }

    fieldTextController[languageList.length].text = offday.date;
    DateTime datetime = DateTime.parse(offday.date);
    selectedDate = datetime;
    screenData = offday;
  }

  void findFocusNext(int index) {
    focusNodeIndex = index + 1;
    if (focusNodeIndex > fieldFocusNodes.length - 1) {
      focusNodeIndex = 0;
    }
    fieldFocusNodes[focusNodeIndex].requestFocus();
    fieldTextController[focusNodeIndex].selection = TextSelection.fromPosition(
      TextPosition(offset: fieldTextController[focusNodeIndex].text.length),
    );
  }

  Future<void> _selectDate(BuildContext context) async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: selectedDate,
      firstDate: DateTime(2022, 8),
      lastDate: DateTime(2101),
    );
    if (picked != null && picked != selectedDate) {
      setState(() {
        selectedDate = picked;
      });
      fieldTextController[languageList.length].text = picked.toString().split(
        " ",
      )[0];
    }
  }

  Widget editScreen({mobileScreen}) {
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
                icon: const Icon(Icons.arrow_back),
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
        title: Text(headerEdit + global.language("holidays")),
        actions: <Widget>[
          EditFontSizeControl(onChanged: () => setState(() {})),
          const SizedBox(width: 8),
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
                            context.read<HolidayBloc>().add(
                              HolidayDelete(guid: selectGuid),
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
          if (isEditMode && global.systemLanguage.length > 1)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  for (int i = 2; i <= languageList.length; i++) {
                    try {
                      var translation = await translator.translate(
                        fieldTextController[1].text,
                        to: languageList[i - 1].codeTranslator!,
                      );
                      fieldTextController[i].text = translation.text;
                    } catch (_) {}
                  }
                  setState(() {});
                },
                icon: const Icon(Icons.translate),
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
                    closeDayListData[closeDayListData.indexOf(
                      closeDayListData.firstWhere(
                        (element) => element.guidfixed == selectGuid,
                      ),
                    )],
                  );
                },
                icon: const Icon(Icons.edit),
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
        ],
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true),
        onKeyEvent: (node, event) {
          if (kIsWeb) {
            if (event is KeyDownEvent) {
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
                      color: global.theme.cardColor,
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            child: Column(
              children: [
                const SizedBox(height: 10),
                for (int i = 0; i < languageList.length; i++)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 10),
                    child: TextFormField(
                      readOnly: !isEditMode,
                      onChanged: (value) {
                        isChange = true;
                      },
                      onFieldSubmitted: (value) {
                        findFocusNext(i);
                      },
                      textInputAction: TextInputAction.next,
                      textAlign: TextAlign.left,
                      controller: fieldTextController[i],
                      decoration: InputDecoration(
                        contentPadding: EdgeInsets.only(
                          left: 10,
                          top: 0,
                          bottom: 0,
                          right: 10,
                        ),
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        labelStyle: (i == 0)
                            ? TextStyle(
                                color: global.theme.inputTextBoxForceColor,
                              )
                            : TextStyle(color: global.theme.inputTextBoxColor),
                        border: const OutlineInputBorder(),
                        labelText:
                            "${global.language("desc")} (${languageList[i].name})",
                      ),
                    ),
                  ),
                TextFormField(
                  readOnly: !isEditMode,
                  controller: fieldTextController[languageList.length],
                  decoration: InputDecoration(
                    border: const OutlineInputBorder(),
                    contentPadding: EdgeInsets.only(
                      left: 10,
                      top: 0,
                      bottom: 0,
                      right: 10,
                    ),
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    labelText: global.language("doc_date"),
                    labelStyle: TextStyle(
                      color: global.theme.inputTextBoxForceColor,
                    ),
                  ),
                  onTap: () => (isEditMode) ? _selectDate(context) : null,
                ),
                const SizedBox(height: 10),
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
                        global.language("save") + ((kIsWeb) ? " (F10)" : ""),
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
    );
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    if (showCheckBox == false) {
      closeDayGuidListChecked.clear();
    }
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topRight,
          end: Alignment.bottomLeft,
          colors: [Colors.blue.shade200, Colors.blue],
        ),
      ),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        resizeToAvoidBottomInset: true,
        body: LayoutBuilder(
          builder: (context, constraints) {
            return BlocListener<HolidayBloc, HolidayState>(
              listener: (context, state) {
                if (state is HolidayLoadSuccess) {
                  setState(() {
                    if (state.holidays.isNotEmpty) {
                      closeDayListData = state.holidays;
                      closeDayListData.sort((a, b) {
                        var adate = a.date;
                        var bdate = b.date;
                        return adate.compareTo(bdate);
                      });
                    }
                  });
                }
                // Save
                if (state is HolidaySaveSuccess) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: Colors.white),
                      global.language("savesuc"),
                      Colors.blue,
                    );
                    clearEditData();
                    closeDayListData.clear();
                    loadDataList(searchText);
                  });
                }
                if (state is HolidaySaveFailed) {
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
                if (state is HolidayUpdateSuccess) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: Colors.white),
                      global.language("edit_success"),
                      Colors.blue,
                    );
                    clearEditData();
                    closeDayListData.clear();
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(0);
                    });
                    loadDataList(searchText);
                    isSaveAllow = false;
                  });
                }
                if (state is HolidayUpdateFailed) {
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
                if (state is HolidayDeleteSuccess) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.delete, color: Colors.white),
                      global.language("delete_success"),
                      Colors.blue,
                    );
                    closeDayListData.clear();
                    clearEditData();
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(0);
                    });
                    loadDataList(searchText);
                  });
                }
                // Delete Many
                if (state is HolidayDeleteManySuccess) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.delete, color: Colors.white),
                      global.language("not_delete_success"),
                      Colors.blue,
                    );
                    closeDayListData.clear();
                    clearEditData();
                    loadDataList(searchText);
                    showCheckBox = false;
                  });
                }
                // Get
                if (state is HolidayGetSuccess) {
                  setState(() {
                    getDataToEditScreen(state.holiday);
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
      ),
    );
  }
}
