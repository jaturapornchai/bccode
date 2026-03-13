import 'dart:io';

import 'package:smlaicloud/bloc/qr/qr_bloc.dart';
import 'package:smlaicloud/model/book_bank_model.dart';
import 'package:smlaicloud/model/qr_model.dart';
import 'package:smlaicloud/screen_search/bookbank_select_screen.dart';
import 'package:smlaicloud/utils/dialog_template.dart';
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

import 'dart:convert';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class QrScreen extends StatefulWidget {
  const QrScreen({super.key});

  @override
  State<QrScreen> createState() => QrscreenState();
}

class QrscreenState extends State<QrScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  late DropzoneViewController dropZoneController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<TextEditingController> fieldTextController = [];
  List<FocusNode> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<QrModel> qrListDatas = [];
  List<QrModel> qrTempListDatas = [];
  List<String> qrGuidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late QrState blocqrstate;
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
  Uint8List? imageWeb;
  final ImagePicker imagePicker = ImagePicker();

  late QrModel screenData;
  final _debouncer = global.Debouncer(1000);
  final _formKey = GlobalKey<FormState>();

  List<bool> selectedQrTemp = [];
  bool selectQrTempAll = false;
  TextEditingController bookBankSerchController = TextEditingController();

  final List<QrTypeModel> qrPromptPay = [
    QrTypeModel(code: 100, name: 'QR PromptPay'),
    QrTypeModel(code: 101, name: 'K Plus Shop'),
  ];

  final List<QrTypeModel> qrLugentList = [
    QrTypeModel(code: 110, name: 'Lugent PromptPat'),
    QrTypeModel(code: 111, name: 'Lugent AliPay'),
    QrTypeModel(code: 112, name: 'Lugent True Money'),
    QrTypeModel(code: 113, name: 'Lugent Line Pay'),
  ];

  final List<QrTypeModel> qrGBList = [
    QrTypeModel(code: 131, name: 'GB Thai QR'),
    QrTypeModel(code: 132, name: 'GB True Money'),
    QrTypeModel(code: 133, name: 'GB ShopeePay'),
    QrTypeModel(code: 134, name: 'GB AliPay'),
    QrTypeModel(code: 135, name: 'GB WeChat Pay'),
  ];

  final List<QrTypeModel> qrXenditList = [
    QrTypeModel(code: 201, name: 'XENDIT PromptPay'),
    // QrTypeModel(code: 202, name: 'XENDIT True Money'),
    // QrTypeModel(code: 203, name: 'XENDIT Line Pay'),
    // QrTypeModel(code: 204, name: 'XENDIT AliPay'),
    // QrTypeModel(code: 205, name: 'XENDIT WeChat Pay'),
  ];

  final List<QrTypeModel> qrBCQRAPI = [
    QrTypeModel(code: 301, name: 'BC PromptPay'),
    QrTypeModel(code: 302, name: 'BC Credit'),
  ];

  final List<QrTypeModel> qrTigerBoard = [
    QrTypeModel(code: 401, name: 'Tiger Board'),
  ];

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

    screenData = QrModel(guidfixed: "", code: "", logo: "");
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
        'https://raw.githubusercontent.com/smlsoft/dedepos_template/main/qrcode.json';

    try {
      final fileContent = await global.readFileFromGithub(githubRawUrl);
      final qrs = (json.decode(fileContent) as List)
          .map((qr) => QrModel.fromJson(qr))
          .toList();
      qrTempListDatas = [];

      for (int i = 0; i < qrs.length; i++) {
        qrTempListDatas.add(qrs[i]);
      }

      if (qrListDatas.isNotEmpty) {
        /// qrTempListDatas remove where qrListDatas
        for (int i = 0; i < qrListDatas.length; i++) {
          qrTempListDatas.removeWhere(
            (element) => element.code == qrListDatas[i].code,
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
    context.read<QrBloc>().add(
      QrLoadList(
        offset: (qrListDatas.isEmpty) ? 0 : qrListDatas.length,
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

    bookBankSerchController.dispose();

    super.dispose();
  }

  void clearEditData() {
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].clear();
    }
    isChange = false;
    focusNodeIndex = 0;
    screenData = QrModel(
      guidfixed: "",
      code: "",
      logo: "",
      qrnames: [],
      qrtype: 100,
      bankcode: "",
      banknames: [],
      bookbankcode: "",
      bookbanknames: [],
      bookbankimages: [],
      closeqr: 0,
    );

    fieldFocusNodes[focusNodeIndex].requestFocus();
    selectedQrTemp = [];
    selectQrTempAll = false;

    setState(() {
      imageFile = File('');
      imageWeb = null;
    });
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
    imageWeb = null;
    imageFile = File('');
    context.read<QrBloc>().add(QrGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('qr_code')),
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
                      qrTempListDatas,
                      "qrcode",
                    );

                if (selectedData.isNotEmpty) {
                  List<QrModel> qrsTempSeleted = [];
                  for (int i = 0; i < selectedData.length; i++) {
                    if (selectedData[i] == true) {
                      qrsTempSeleted.add(qrTempListDatas[i]);
                    }
                  }

                  // ignore: use_build_context_synchronously
                  context.read<QrBloc>().add(QrBulkSave(qrs: qrsTempSeleted));
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
                        qrGuidListChecked.clear();
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
          if (qrGuidListChecked.isNotEmpty)
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
                            context.read<QrBloc>().add(
                              QrDeleteMany(guid: qrGuidListChecked),
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
                int index = qrListDatas.indexOf(
                  qrListDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = qrListDatas[index - 1].guidfixed!;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  getData(selectGuid);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = qrListDatas.indexOf(
                  qrListDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = qrListDatas[index + 1].guidfixed!;
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
                            qrListDatas = [];
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
                        hintText: (kIsWeb)
                            ? "${global.language('search')} (F2)"
                            : global.language('search'),
                        prefixIcon: Icon(Icons.search, size: 20),
                        prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                      ),
                    ),
                  ),
                  ListFontSizeControl(onChanged: () => setState(() {})),
                  const SizedBox(width: 4),
                  if (qrListDatas.isNotEmpty)
                    Text(
                      '(${qrListDatas.length})',
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
                      global.language("qr_code"),
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
                      global.language("qr_name"),
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
                  children: qrListDatas
                      .map(
                        (value) => listObject(
                          qrListDatas.indexOf(value),
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

  void switchToEdit(QrModel value) {
    setState(() {
      selectGuid = value.guidfixed!;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listObject(int index, QrModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < qrGuidListChecked.length; i++) {
      if (qrGuidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    final isSelected = selectGuid == value.guidfixed!;
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
            selectGuid = value.guidfixed!;
            if (isCheck == true) {
              qrGuidListChecked.remove(value.guidfixed);
            } else {
              qrGuidListChecked.add(value.guidfixed!);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              "${global.language("chosen")} ${qrGuidListChecked.length} ${global.language("list")}",
              Colors.blue,
            );
          });
        } else {
          setState(() {
            discardData(
              callBack: () {
                isSaveAllow = false;
                isEditMode = false;
                selectGuid = value.guidfixed!;
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
        color: _getContainerColor(value.guidfixed!, index),
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
                value.code!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.qrnames!),
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
      QrModel qrModel = QrModel(
        guidfixed: "",
        code: fieldTextController[0].text,
        qrnames: packLanguage(),
        qrtype: screenData.qrtype,
        bankcode: screenData.bankcode,
        banknames: screenData.banknames,
        bookbankcode: screenData.bookbankcode,
        bookbanknames: screenData.bookbanknames,
        bookbankimages: screenData.bookbankimages,
        closeqr: screenData.closeqr,
      );

      if (imageFile.path.isNotEmpty) {
        context.read<QrBloc>().add(
          QrSaveWithImage(
            qr: qrModel,
            imageFile: imageFile,
            imageWeb: imageWeb!,
          ),
        );
      } else {
        context.read<QrBloc>().add(QrSave(qr: qrModel));
      }
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    var names = packLanguage();
    showCheckBox = false;
    QrModel qrModel = QrModel(
      guidfixed: guid,
      code: fieldTextController[0].text,
      logo: screenData.logo,
      qrnames: names,
      qrtype: screenData.qrtype,
      bankcode: screenData.bankcode,
      banknames: screenData.banknames,
      bookbankcode: screenData.bookbankcode,
      bookbanknames: screenData.bookbanknames,
      bookbankimages: screenData.bookbankimages,
      closeqr: screenData.closeqr,
    );

    if (imageWeb != null) {
      context.read<QrBloc>().add(
        QrWithImageUpdate(
          guid: guid,
          qr: qrModel,
          imageFile: imageFile,
          imageWeb: imageWeb!,
        ),
      );
    } else {
      context.read<QrBloc>().add(QrUpdate(guid: guid, qrModel: qrModel));
    }
  }

  void getDataToEditScreen(QrModel qr) {
    isChange = false;
    selectGuid = qr.guidfixed!;
    fieldTextController[0].text = qr.code!;
    screenData = qr;

    if (screenData.bankcode!.isNotEmpty) {
      bookBankSerchController.text =
          "${screenData.bankcode} ~ ${screenData.bookbankcode!}";
    } else {
      bookBankSerchController.text = "";
    }

    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i + 1].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < qr.qrnames!.length; j++) {
        if (languageList[i].code! == qr.qrnames![j].code) {
          fieldTextController[i + 1].text = qr.qrnames![j].name;
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
        title: Text(headerEdit + global.language("qr_code")),
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
                            context.read<QrBloc>().add(
                              QrDelete(guid: selectGuid),
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
                    qrListDatas[qrListDatas.indexOf(
                      qrListDatas.firstWhere(
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
                      labelText: global.language("qr_code"),
                      labelStyle: TextStyle(
                        color: global.theme.inputTextBoxForceColor,
                      ),
                    ),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'This field is required';
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
                              "${global.language("qr_name")} (${languageList[i].name})",
                        ),
                        validator: (value) {
                          if (i == 0) {
                            if (value == null || value.isEmpty) {
                              return 'This field is required';
                            }
                          }

                          return null;
                        },
                      ),
                    ),
                  InputDecorator(
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: global.language("qr_type"),
                    ),
                    child: Column(
                      children: [
                        Row(
                          children: [
                            for (int i = 0; i < qrPromptPay.length; i++)
                              Row(
                                children: [
                                  Radio(
                                    focusNode: FocusNode(skipTraversal: true),
                                    value: qrPromptPay[i].code,
                                    groupValue: screenData.qrtype,
                                    onChanged: (value) {
                                      setState(() {
                                        screenData.qrtype = value!;
                                      });
                                    },
                                  ),
                                  Text(qrPromptPay[i].name),
                                ],
                              ),
                          ],
                        ),
                        Row(
                          children: [
                            for (int i = 0; i < qrLugentList.length; i++)
                              Row(
                                children: [
                                  Radio(
                                    focusNode: FocusNode(skipTraversal: true),
                                    value: qrLugentList[i].code,
                                    groupValue: screenData.qrtype,
                                    onChanged: (value) {
                                      setState(() {
                                        screenData.qrtype = value!;
                                      });
                                    },
                                  ),
                                  Text(qrLugentList[i].name),
                                ],
                              ),
                          ],
                        ),
                        Row(
                          children: [
                            for (int i = 0; i < qrGBList.length; i++)
                              Row(
                                children: [
                                  Radio(
                                    focusNode: FocusNode(skipTraversal: true),
                                    value: qrGBList[i].code,
                                    groupValue: screenData.qrtype,
                                    onChanged: (value) {
                                      setState(() {
                                        screenData.qrtype = value!;
                                      });
                                    },
                                  ),
                                  Text(qrGBList[i].name),
                                ],
                              ),
                          ],
                        ),

                        /// xendit
                        Row(
                          children: [
                            for (int i = 0; i < qrXenditList.length; i++)
                              Row(
                                children: [
                                  Radio(
                                    focusNode: FocusNode(skipTraversal: true),
                                    value: qrXenditList[i].code,
                                    groupValue: screenData.qrtype,
                                    onChanged: (value) {
                                      setState(() {
                                        screenData.qrtype = value!;
                                      });
                                    },
                                  ),
                                  Text(qrXenditList[i].name),
                                ],
                              ),
                          ],
                        ),

                        /// SMLQRAPI
                        Row(
                          children: [
                            for (int i = 0; i < qrBCQRAPI.length; i++)
                              Row(
                                children: [
                                  Radio(
                                    focusNode: FocusNode(skipTraversal: true),
                                    value: qrBCQRAPI[i].code,
                                    groupValue: screenData.qrtype,
                                    onChanged: (value) {
                                      setState(() {
                                        screenData.qrtype = value!;
                                      });
                                    },
                                  ),
                                  Text(qrBCQRAPI[i].name),
                                ],
                              ),
                          ],
                        ),

                        /// TigerBoard
                        Row(
                          children: [
                            for (int i = 0; i < qrTigerBoard.length; i++)
                              Row(
                                children: [
                                  Radio(
                                    focusNode: FocusNode(skipTraversal: true),
                                    value: qrTigerBoard[i].code,
                                    groupValue: screenData.qrtype,
                                    onChanged: (value) {
                                      setState(() {
                                        screenData.qrtype = value!;
                                      });
                                    },
                                  ),
                                  Text(qrTigerBoard[i].name),
                                ],
                              ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  SizedBox(height: 15),
                  TextFormField(
                    readOnly: true,
                    textInputAction: TextInputAction.next,
                    textAlign: TextAlign.left,
                    controller: bookBankSerchController,
                    textCapitalization: TextCapitalization.characters,
                    inputFormatters: <TextInputFormatter>[
                      FilteringTextInputFormatter.allow(
                        RegExp('[a-z A-Z 0-9 -]'),
                      ),
                    ],
                    onChanged: (value) {},
                    decoration: InputDecoration(
                      border: const OutlineInputBorder(),
                      contentPadding: EdgeInsets.only(
                        left: 10,
                        top: 0,
                        bottom: 0,
                        right: 10,
                      ),
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      labelText: global.language("bookBank_code"),
                      labelStyle: TextStyle(
                        color: global.theme.inputTextBoxForceColor,
                      ),
                      suffixIcon: IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        icon: const Icon(Icons.search),
                        onPressed: () {
                          (isEditMode)
                              ? Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (context) =>
                                        const BookBankSelectScreen(),
                                  ),
                                ).then((value) {
                                  if (value != null) {
                                    setState(() {
                                      BookBankModel returnValue =
                                          value as BookBankModel;
                                      bookBankSerchController.text =
                                          "${returnValue.bankcode} ~ ${returnValue.passbook}";
                                      screenData.bankcode =
                                          returnValue.bankcode;
                                      screenData.banknames =
                                          returnValue.banknames;
                                      screenData.bookbankcode =
                                          returnValue.passbook;
                                      screenData.bookbanknames =
                                          returnValue.names;
                                      screenData.bookbankimages =
                                          returnValue.images;
                                    });
                                  }
                                })
                              : null;
                        },
                      ),
                    ),
                  ),
                  SizedBox(height: 15),
                  Row(
                    children: [
                      Radio(
                        focusNode: FocusNode(skipTraversal: true),
                        value: 0,
                        groupValue: screenData.closeqr,
                        onChanged: (value) {
                          setState(() {
                            screenData.closeqr = value!;
                          });
                        },
                      ),
                      Text(global.language('money_now')),
                      const SizedBox(width: 10),
                      Radio(
                        focusNode: FocusNode(skipTraversal: true),
                        value: 1,
                        groupValue: screenData.closeqr,
                        onChanged: (value) {
                          setState(() {
                            screenData.closeqr = value!;
                          });
                        },
                      ),
                      Text(global.language('money_end_day')),
                      const SizedBox(width: 10),
                      Radio(
                        focusNode: FocusNode(skipTraversal: true),
                        value: 2,
                        groupValue: screenData.closeqr,
                        onChanged: (value) {
                          setState(() {
                            screenData.closeqr = value!;
                          });
                        },
                      ),
                      Text(global.language('money_next_day')),
                    ],
                  ),
                  const SizedBox(height: 15),
                  (isEditMode)
                      ? Padding(
                          padding: EdgeInsets.all(8.0),
                          child: Row(
                            children: [
                              Expanded(
                                child: ElevatedButton.icon(
                                  focusNode: FocusNode(skipTraversal: true),
                                  onPressed: () async {
                                    FocusScope.of(context).unfocus();
                                    setState(() {
                                      imageWeb = null;
                                      imageFile = File('');
                                      screenData.logo = "";
                                    });
                                  },
                                  icon: Icon(Icons.delete),
                                  label: Text(
                                    global.language('delete_picture'),
                                  ),
                                ),
                              ),
                              const SizedBox(width: 5),
                              Expanded(
                                child: ElevatedButton.icon(
                                  focusNode: FocusNode(skipTraversal: true),
                                  onPressed: (kIsWeb)
                                      ? () async {
                                          FocusScope.of(context).unfocus();
                                          XFile? image = await imagePicker
                                              .pickImage(
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
                                            final XFile? photo =
                                                await imagePicker.pickImage(
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
                                  label: Text(
                                    global.language("select_picture"),
                                  ),
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
                        )
                      : Container(),
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
                              color: global.theme.cardColor,
                              border: Border.all(color: global.theme.textColor),
                              borderRadius: BorderRadius.circular(5),
                              image: (imageWeb != null)
                                  ? DecorationImage(
                                      image: MemoryImage(imageWeb!),
                                      fit: BoxFit.fill,
                                    )
                                  : (screenData.logo != '')
                                  ? DecorationImage(
                                      image: NetworkImage(screenData.logo!),
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
      qrGuidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<QrBloc, QrState>(
            listener: (context, state) {
              blocqrstate = state;
              // Load
              if (state is QrLoadSuccess) {
                setState(() {
                  if (state.qrs.isNotEmpty) {
                    qrListDatas.addAll(state.qrs);

                    for (var qr in qrListDatas) {
                      qrTempListDatas.removeWhere((ele) => ele.code == qr.code);
                    }
                    // สร้าง keys ใหม่ให้ตรงกับจำนวน items
                    listKeys.clear();
                    for (int i = 0; i < qrListDatas.length; i++) {
                      listKeys.add(GlobalKey());
                    }
                    // print(qrTempListDatas);
                  }
                });
              }
              // Save
              if (state is QrSaveSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    global.language("save_success"),
                    global.theme.primaryColor,
                  );
                  clearEditData();
                  qrListDatas.clear();
                  loadDataList(searchText);
                });
              }
              if (state is QrSaveFailed) {
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
              if (state is QrUpdateSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: Colors.white),
                    global.language("edit_success"),
                    global.theme.primaryColor,
                  );
                  clearEditData();
                  qrListDatas.clear();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                  isSaveAllow = false;
                  getData(selectGuid);
                });
              }
              if (state is QrUpdateFailed) {
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
              if (state is QrDeleteSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    global.theme.primaryColor,
                  );
                  qrListDatas.clear();
                  clearEditData();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                });
              }
              // Delete Many
              if (state is QrDeleteManySuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    global.theme.primaryColor,
                  );
                  qrListDatas.clear();
                  clearEditData();
                  loadDataList(searchText);
                  showCheckBox = false;
                });
              }
              // Get
              if (state is QrGetSuccess) {
                setState(() {
                  getDataToEditScreen(state.qrs);

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
