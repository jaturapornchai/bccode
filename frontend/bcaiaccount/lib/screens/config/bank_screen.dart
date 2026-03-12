import 'dart:io';

import 'package:smlaicloud/utils/dialog_template.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/bank/bank_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/bank_model.dart';
import 'package:image_picker/image_picker.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

import 'dart:convert';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class BankScreen extends StatefulWidget {
  const BankScreen({super.key});

  @override
  State<BankScreen> createState() => BankScreenState();
}

class BankScreenState extends State<BankScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<TextEditingController> fieldTextController = [];
  List<FocusNode> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<BankModel> bankListDatas = [];
  List<BankModel> bankTempListDatas = [];
  List<String> bankGuidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late BankState blocBankState;
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
  File imageFile = File('');
  Uint8List imageWeb = Uint8List(0);
  late BankModel screenData;
  final _debouncer = global.Debouncer(1000);
  final _formKey = GlobalKey<FormState>();

  List<bool> selectedBankTemp = [];
  bool selectBankTempAll = false;

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
      fieldFocusNodes.add(focusNode);
    }
  }

  @override
  void initState() {
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    screenData = BankModel(guidfixed: "", code: "", logo: "");
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    // เรียงลำดับ Focus
    // Focus รหัส
    FocusNode focusNode = FocusNode();
    focusNode.addListener(() {
      focusNodeIndex = 0;
    });
    fieldFocusNodes.add(focusNode);
    fieldTextController.add(TextEditingController());
    setSystemLanguageList();

    listScrollController.addListener(onScrollList);
    loadDataList("");

    super.initState();
  }

  Future<void> getTemplate() async {
    const githubRawUrl =
        'https://raw.githubusercontent.com/smlsoft/dedepos_template/main/bank.json';

    try {
      final fileContent = await global.readFileFromGithub(githubRawUrl);
      final banks = (json.decode(fileContent) as List)
          .map((bank) => BankModel.fromJson(bank))
          .toList();
      bankTempListDatas = [];

      for (int i = 0; i < banks.length; i++) {
        bankTempListDatas.add(banks[i]);
      }

      if (bankListDatas.isNotEmpty) {
        /// bankTempListDatas remove where bankListDatas
        for (int i = 0; i < bankListDatas.length; i++) {
          bankTempListDatas.removeWhere(
            (element) => element.code == bankListDatas[i].code,
          );
        }
      }

      loadDataList("");
    } catch (error) {
      // Handle error
      // ignore: avoid_print
      if (kDebugMode) {
        AppLogger.error('Error reading file: $error');
      }
    }
  }

  void loadDataList(String search) {
    context.read<BankBloc>().add(
      BankLoadList(
        offset: (bankListDatas.isEmpty) ? 0 : bankListDatas.length,
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
    screenData.logo = "";
    imageFile = File('');
    imageWeb = Uint8List(0);
    fieldFocusNodes[focusNodeIndex].requestFocus();
    selectedBankTemp = [];
    selectBankTempAll = false;
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

  void getData(String guid) {
    headerEdit = global.language("show");
    isEditMode = false;
    context.read<BankBloc>().add(BankGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('bank')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                // Navigator.pop(context);
                global.gotoMainMenu(context);
                isEditMode = false;
              },
            );
          },
        ),
        actions: <Widget>[
          Padding(
            padding: const EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () async {
                await getTemplate();

                // ignore: use_build_context_synchronously
                List<dynamic> selectedData =
                    await DialogTemplate.showDataListTemplateDialog(
                      context,
                      bankTempListDatas,
                      "bank",
                    );

                if (selectedData.isNotEmpty) {
                  List<BankModel> banksTempSeleted = [];
                  for (int i = 0; i < selectedData.length; i++) {
                    if (selectedData[i] == true) {
                      banksTempSeleted.add(bankTempListDatas[i]);
                    }
                  }

                  // ignore: use_build_context_synchronously
                  context.read<BankBloc>().add(
                    BankBulkSave(banks: banksTempSeleted),
                  );
                }
              },
              icon: const Icon(Icons.cloud_download),
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
                      if (showCheckBox) {
                        showCheckBox = false;
                        bankGuidListChecked.clear();
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
          if (bankGuidListChecked.isNotEmpty)
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
                            context.read<BankBloc>().add(
                              BankDeleteMany(guid: bankGuidListChecked),
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
                int index = bankListDatas.indexOf(
                  bankListDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = bankListDatas[index - 1].guidfixed;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  getData(selectGuid);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = bankListDatas.indexOf(
                  bankListDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = bankListDatas[index + 1].guidfixed;
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
                        _debouncer.run(() {
                          setState(() {
                            bankListDatas = [];
                          });
                          loadDataList(value);
                        });
                      },
                      autofocus: false,
                      focusNode: searchFocusNode,
                      controller: searchController,
                      decoration: InputDecoration(
                        isDense: true,
                        contentPadding: EdgeInsets.only(
                          top: 0,
                          bottom: 0,
                          left: 0,
                          right: 0,
                        ),
                        border: InputBorder.none,
                        prefixIcon: Icon(Icons.search, size: 20, color: global.theme.iconColor),
                        prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                        hintText: (kIsWeb)
                            ? "${global.language('search')} (F2)"
                            : global.language('search'),
                      ),
                    ),
                  ),
                  ListFontSizeControl(onChanged: () => setState(() {})),
                const SizedBox(width: 4),
                if (bankListDatas.isNotEmpty)
                  Text(
                    '(${bankListDatas.length})',
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
                      global.language("bank_code"),
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
                      global.language("bank_name"),
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
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: bankListDatas
                      .map(
                        (value) => listObject(
                          bankListDatas.indexOf(value),
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

  void switchToEdit(BankModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listObject(int index, BankModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < bankGuidListChecked.length; i++) {
      if (bankGuidListChecked[i] == value.guidfixed) {
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
              bankGuidListChecked.remove(value.guidfixed);
            } else {
              bankGuidListChecked.add(value.guidfixed);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              "${global.language("chosen")} ${bankGuidListChecked.length} ${global.language("list")}",
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
                style: TextStyle(
                  fontSize: global.deviceConfig.listDataFontSize,
                ),
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.names),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: global.deviceConfig.listDataFontSize,
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

  void saveOrUpdateData() {
    showCheckBox = false;
    if (selectGuid.trim().isEmpty) {
      BankModel bankModel = BankModel(
        guidfixed: "",
        logo: "",
        code: fieldTextController[0].text,
        names: packLanguage(),
      );

      if (imageWeb.isNotEmpty) {
        context.read<BankBloc>().add(
          BankSaveWithImage(
            bank: bankModel,
            imageFile: imageFile,
            imageWeb: imageWeb,
          ),
        );
      } else {
        context.read<BankBloc>().add(BankSave(bank: bankModel));
      }
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    var names = packLanguage();
    showCheckBox = false;
    BankModel bankModel = BankModel(
      guidfixed: guid,
      logo: screenData.logo,
      code: fieldTextController[0].text,
      names: names,
    );
    Uint8List imageWebUpdate = Uint8List(0);
    ImagesModel imageUris = ImagesModel(uri: "", xorder: 0);

    if (imageWeb.isNotEmpty || screenData.logo != '') {
      imageWebUpdate = imageWeb;
      imageUris = ImagesModel(uri: screenData.logo, xorder: 0);
    }

    if (imageWebUpdate.isNotEmpty) {
      context.read<BankBloc>().add(
        BankWithImageUpdate(
          guid: guid,
          bank: bankModel,
          imageFile: imageFile,
          imagesUri: imageUris,
          imageWeb: imageWeb,
        ),
      );
    } else {
      context.read<BankBloc>().add(
        BankUpdate(guid: guid, bankModel: bankModel),
      );
    }
  }

  void getDataToEditScreen(BankModel bank) {
    isChange = false;
    selectGuid = bank.guidfixed;
    fieldTextController[0].text = bank.code;
    screenData.logo = bank.logo;
    imageFile = File('');
    imageWeb = Uint8List(0);
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i + 1].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < bank.names.length; j++) {
        if (languageList[i].code! == bank.names[j].code) {
          fieldTextController[i + 1].text = bank.names[j].name;
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
    fieldTextController[focusNodeIndex].selection = TextSelection.fromPosition(
      TextPosition(offset: fieldTextController[focusNodeIndex].text.length),
    );
  }

  Widget editScreen({mobileScreen}) {
    return Scaffold(
      backgroundColor: isEditMode ? global.theme.toolBarEditModeColor.withValues(alpha: 0.05) : global.theme.backgroundColor,
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
        title: Text(headerEdit + global.language("bank")),
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
                            context.read<BankBloc>().add(
                              BankDelete(guid: selectGuid),
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
                    bankListDatas[bankListDatas.indexOf(
                      bankListDatas.firstWhere(
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
                onPressed: () {
                  if (_formKey.currentState!.validate()) {
                    saveOrUpdateData();
                  }
                },
                icon: const Icon(Icons.save),
              ),
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
                if (_formKey.currentState!.validate()) {
                  saveOrUpdateData();
                }
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: SingleChildScrollView(
          controller: editScrollController,
          child: Container(
            color: Colors.white,
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            child: Form(
              key: _formKey,
              child: Column(
                children: [
                  const SizedBox(height: 10),
                  TextFormField(
                    enabled: selectGuid.isEmpty,
                    readOnly: !isEditMode,
                    onFieldSubmitted: (value) {
                      findFocusNext(0);
                    },
                    textInputAction: TextInputAction.next,
                    focusNode: fieldFocusNodes[0],
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
                      border: const OutlineInputBorder(),
                      contentPadding: EdgeInsets.only(
                        left: 10,
                        top: 0,
                        bottom: 0,
                        right: 10,
                      ),
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      labelText: global.language("bank_code"),
                      labelStyle: TextStyle(
                        color: global.theme.inputTextBoxForceColor,
                      ),
                    ),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return global.language('this_field_is_required');
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 15),
                  for (int i = 0; i < languageList.length; i++)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 15),
                      child: TextFormField(
                        readOnly: !isEditMode,
                        onChanged: (value) {
                          isChange = true;
                        },
                        onFieldSubmitted: (value) {
                          findFocusNext(focusNodeIndex + 1);
                        },
                        textInputAction: TextInputAction.next,
                        focusNode: fieldFocusNodes[i + 1],
                        textAlign: TextAlign.left,
                        controller: fieldTextController[i + 1],
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
                              : TextStyle(
                                  color: global.theme.inputTextBoxColor,
                                ),
                          border: const OutlineInputBorder(),
                          labelText:
                              "${global.language("bank_name")} (${languageList[i].name})",
                        ),
                        validator: (value) {
                          if (i == 0) {
                            if (value == null || value.isEmpty) {
                              return global.language('this_field_is_required');
                            }
                          }

                          return null;
                        },
                      ),
                    ),
                  const SizedBox(height: 15),

                  // Container(
                  //   width: 300,
                  //   padding: EdgeInsets.only(left: 5, right: 5, bottom: 5, top: 5),
                  //   decoration: BoxDecoration(border: Border.all(color: global.theme.textSecondaryColor), borderRadius: BorderRadius.all(Radius.circular(5.0))),
                  //   child: Column(
                  //     children: [
                  //       Row(
                  //         children: [
                  //           Expanded(
                  //               child: IconButton(
                  //             focusNode: FocusNode(skipTraversal: true),
                  //             onPressed: () async {
                  //               imageWeb = Uint8List(0);
                  //               imageFile = File('');
                  //               screenData.logo = '';

                  //               setState(() {});
                  //             },
                  //             icon: const Icon(
                  //               Icons.delete,
                  //             ),
                  //           )),
                  //           const SizedBox(width: 5),
                  //           Expanded(
                  //               child: IconButton(
                  //             focusNode: FocusNode(skipTraversal: true),
                  //             onPressed: () async {
                  //               // final XFile? photo = await imagePicker.pickImage(
                  //               //     source: ImageSource.gallery,
                  //               //     maxHeight: 480,
                  //               //     maxWidth: 640,
                  //               //     imageQuality: 60);
                  //               // if (photo != null) {
                  //               //   var f = await photo.readAsBytes();
                  //               //   imageWeb[imageIndex] = f;
                  //               //   imageFile.add(File(photo.path));
                  //               //   setState(() {});
                  //               // }
                  //               final XFile? image = await imagePicker.pickImage(source: ImageSource.gallery, maxHeight: 480, maxWidth: 640);
                  //               if (image != null) {
                  //                 var f = await image.readAsBytes();
                  //                 imageWeb = f;
                  //                 imageFile = File(image.path);
                  //                 setState(() {});
                  //               }
                  //             },
                  //             icon: const Icon(
                  //               Icons.folder,
                  //             ),
                  //           )),
                  //           const SizedBox(width: 5),
                  //           if (kIsWeb == false)
                  //             Expanded(
                  //                 child: IconButton(
                  //               focusNode: FocusNode(skipTraversal: true),
                  //               onPressed: () async {
                  //                 final XFile? photo = await imagePicker.pickImage(source: ImageSource.camera, maxHeight: 480, maxWidth: 640, imageQuality: 60);
                  //                 if (photo != null) {
                  //                   var f = await photo.readAsBytes();
                  //                   imageWeb = f;
                  //                   imageFile = File(photo.path);
                  //                   setState(() {});
                  //                 }
                  //               },
                  //               icon: const Icon(
                  //                 Icons.camera_alt,
                  //               ),
                  //             )),
                  //         ],
                  //       ),
                  //       SizedBox(
                  //           width: 300,
                  //           height: 300,
                  //           child: Stack(children: [
                  //             DropzoneView(
                  //               operation: DragOperation.copy,
                  //               cursor: CursorType.grab,
                  //               onCreated: (ctrl) => dropZoneController = ctrl,
                  //               onLoaded: () {},
                  //               onError: (ev) {},
                  //               onHover: () {},
                  //               onLeave: () {},
                  //               onDrop: (ev) async {
                  //                 final bytes = await dropZoneController.getFileData(ev);
                  //                 setState(() {
                  //                   imageWeb = bytes;
                  //                 });
                  //               },
                  //               onDropMultiple: (ev) async {},
                  //             ),
                  //             Center(
                  //                 child: DecoratedBox(
                  //               decoration: BoxDecoration(
                  //                 color: Colors.white,
                  //                 border: Border.all(color: global.theme.textColor),
                  //                 borderRadius: BorderRadius.circular(5),
                  //                 boxShadow: const [
                  //                   BoxShadow(
                  //                       offset: Offset(0, 4),
                  //                       color: Colors.cyan, //edited
                  //                       spreadRadius: 4,
                  //                       blurRadius: 10 //edited
                  //                       )
                  //                 ],
                  //                 image: (imageWeb.isNotEmpty)
                  //                     ? DecorationImage(image: MemoryImage(imageWeb), fit: BoxFit.fill)
                  //                     : (screenData.logo != '')
                  //                         ? DecorationImage(image: NetworkImage(screenData.logo))
                  //                         : const DecorationImage(image: AssetImage('assets/img/noimage.png')),
                  //               ),
                  //               child: const SizedBox(
                  //                 width: 500,
                  //                 height: 500,
                  //               ),
                  //             )),
                  //           ])),
                  //     ],
                  //   ),
                  // ),
                  Padding(
                    padding: EdgeInsets.all(8.0),
                    child: Row(
                      children: [
                        Expanded(
                          child: ElevatedButton.icon(
                            focusNode: FocusNode(skipTraversal: true),
                            onPressed: () async {
                              FocusScope.of(context).unfocus();
                              setState(() {
                                imageWeb = Uint8List(0);
                                imageFile = File('');
                              });
                            },
                            icon: Icon(Icons.delete),
                            label: Text(global.language('delete_picture')),
                          ),
                        ),
                        const SizedBox(width: 5),
                        Expanded(
                          child: ElevatedButton.icon(
                            focusNode: FocusNode(skipTraversal: true),
                            onPressed: (kIsWeb)
                                ? () async {
                                    FocusScope.of(context).unfocus();
                                    XFile? image = await imagePicker.pickImage(
                                      source: ImageSource.gallery,
                                      maxHeight: 480,
                                      maxWidth: 640,
                                      imageQuality: 60,
                                    );
                                    if (image != null) {
                                      var f = await image.readAsBytes();
                                      setState(() {
                                        imageWeb = f;
                                        imageFile = File(image.path);
                                      });
                                    }
                                  }
                                : () {
                                    FocusScope.of(context).unfocus();
                                    setState(() async {
                                      final XFile? photo = await imagePicker
                                          .pickImage(
                                            source: ImageSource.gallery,
                                            maxHeight: 480,
                                            maxWidth: 640,
                                            imageQuality: 60,
                                          );
                                      if (photo != null) {
                                        var f = await photo.readAsBytes();
                                        setState(() {
                                          imageWeb = f;
                                          imageFile = File(photo.path);
                                        });
                                      }
                                    });
                                  },
                            icon: Icon(Icons.folder),
                            label: Text(global.language("select_picture")),
                          ),
                        ),
                        const SizedBox(width: 5),
                        if (kIsWeb == false)
                          Expanded(
                            child: ElevatedButton.icon(
                              focusNode: FocusNode(skipTraversal: true),
                              onPressed: () async {
                                FocusScope.of(context).unfocus();
                                final XFile? photo = await imagePicker
                                    .pickImage(
                                      source: ImageSource.camera,
                                      maxHeight: 480,
                                      maxWidth: 640,
                                      imageQuality: 60,
                                    );
                                if (photo != null) {
                                  var f = await photo.readAsBytes();
                                  setState(() {
                                    imageWeb = f;
                                    imageFile = File(photo.path);
                                  });
                                }
                              },
                              icon: Icon(Icons.camera_alt),
                              label: Text(global.language('take_photo')),
                            ),
                          ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.only(
                      left: 10,
                      right: 10,
                      bottom: 10,
                    ),
                    width: double.infinity,
                    height: 300,
                    child: Stack(
                      children: [
                        if (kIsWeb)
                          DropzoneView(
                            operation: DragOperation.copy,
                            cursor: CursorType.grab,
                            onCreated: (ctrl) => dropZoneController = ctrl,
                            onLoaded: () {},
                            onError: (ev) {},
                            onHover: () {},
                            onLeave: () {},
                            onDrop: (ev) async {
                              final bytes = await dropZoneController
                                  .getFileData(ev);
                              setState(() {
                                imageWeb = bytes;
                              });
                            },
                            onDropMultiple: (ev) async {},
                          ),
                        Center(
                          child: DecoratedBox(
                            decoration: BoxDecoration(
                              color: Colors.white,
                              border: Border.all(color: global.theme.textColor),
                              borderRadius: BorderRadius.circular(5),
                              image: (imageWeb.isNotEmpty)
                                  ? DecorationImage(
                                      image: MemoryImage(imageWeb),
                                    )
                                  : (screenData.logo != '')
                                  ? DecorationImage(
                                      image: NetworkImage(screenData.logo),
                                    )
                                  : const DecorationImage(
                                      image: AssetImage(
                                        'assets/img/noimage.png',
                                      ),
                                    ),
                            ),
                            child: const SizedBox(
                              width: double.infinity,
                              height: 400,
                            ),
                          ),
                        ),
                      ],
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
                          if (_formKey.currentState!.validate()) {
                            saveOrUpdateData();
                          }
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
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    if (showCheckBox == false) {
      bankGuidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<BankBloc, BankState>(
            listener: (context, state) {
              blocBankState = state;
              // Load
              if (state is BankLoadSuccess) {
                setState(() {
                  if (state.banks.isNotEmpty) {
                    bankListDatas.addAll(state.banks);

                    for (var bank in bankListDatas) {
                      bankTempListDatas.removeWhere(
                        (ele) => ele.code == bank.code,
                      );
                    }
                    // print(bankTempListDatas);
                  }
                });
              }
              // Save
              if (state is BankSaveSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    global.language("save_success"),
                    global.theme.primaryColor,
                  );
                  clearEditData();
                  bankListDatas.clear();
                  loadDataList(searchText);
                });
              }
              if (state is BankSaveFailed) {
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
              if (state is BankUpdateSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: Colors.white),
                    global.language("edit_success"),
                    global.theme.primaryColor,
                  );
                  clearEditData();
                  bankListDatas.clear();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                  isSaveAllow = false;
                  getData(selectGuid);
                });
              }
              if (state is BankUpdateFailed) {
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
              if (state is BankDeleteSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    global.theme.primaryColor,
                  );
                  bankListDatas.clear();
                  clearEditData();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                });
              }
              // Delete Many
              if (state is BankDeleteManySuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    global.theme.primaryColor,
                  );
                  bankListDatas.clear();
                  clearEditData();
                  loadDataList(searchText);
                  showCheckBox = false;
                });
              }
              // Get
              if (state is BankGetSuccess) {
                setState(() {
                  getDataToEditScreen(state.bank);

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
    );
  }
}
