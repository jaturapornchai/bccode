import 'dart:io';

import 'package:smlaicloud/bloc/book_bank/book_bank_bloc.dart';
import 'package:smlaicloud/model/book_bank_model.dart';
import 'package:smlaicloud/screen_search/accountchart_select_screen.dart';
import 'package:smlaicloud/screen_search/bank_select_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class BookBankScreen extends StatefulWidget {
  const BookBankScreen({super.key});

  @override
  State<BookBankScreen> createState() => BookBankScreenState();
}

class BookBankScreenState extends State<BookBankScreen>
    with SingleTickerProviderStateMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  bool refreshFocus = false;
  TextEditingController searchController = TextEditingController();
  TextEditingController groupController = TextEditingController();
  int focusNodeMax = 0;
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<global.FieldFocusModel> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<BookBankModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late BookBankState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  bool isEditMode = false;
  late BookBankModel screenData;
  List<Uint8List> imageWeb = [];
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;
  Color colorSelected = Colors.white;
  final _debouncer = global.Debouncer(1000);

  bool loadingData = false;
  List<File> imageFile = [];
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  TextEditingController bookBankCode = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  void setSystemLanguageList() async {
    clearEditData();
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }
    clearEditData();
    loadDataList("");
  }

  @override
  void initState() {
    setSystemLanguageList();
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    // เรียงลำดับ Focus
    for (int i = 0; i < 100; i++) {
      fieldFocusNodes.add(global.FieldFocusModel(focusNode: FocusNode()));
      fieldFocusNodes[i].focusNode.addListener(() {
        if (fieldFocusNodes[i].focusNode.hasFocus) {
          focusNodeIndex = i;
          fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        }
      });
    }
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    listScrollController.addListener(onScrollList);

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    groupController.dispose();
    for (int i = 0; i < fieldFocusNodes.length; i++) {
      fieldFocusNodes[i].focusNode.dispose();
    }
    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<BookBankBloc>().add(
      BookBankLoadList(
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

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
  }

  void clearEditData() {
    List<LanguageDataModel> names = [];
    for (int k = 0; k < languageList.length; k++) {
      names.add(LanguageDataModel(code: languageList[k].code!, name: ""));
    }
    bookBankCode.text = "";
    screenData = BookBankModel(
      bankcode: '',
      bookcode: '',
      guidfixed: '',
      images: [],
      passbook: '',
      bankbranch: '',
    );

    isDataChange = false;
    focusNodeIndex = 0;
    refreshFocus = true;
  }

  void discardData({required Function callBack}) {
    if (isEditMode && isDataChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: Colors.blue),
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
    context.read<BookBankBloc>().add(BookBankGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('book_bank')),
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
                        guidListChecked.clear();
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
          if (guidListChecked.isNotEmpty)
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
                            backgroundColor: Colors.red,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.blue,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<BookBankBloc>().add(
                              BookBankDeleteMany(guid: guidListChecked),
                            );
                          },
                          child: Text(global.language('delete')),
                        ),
                      ],
                    ),
                  );
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
              ),
            ),

          /// เพิ่มข้อมูลใหม่
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
                      isDataChange = false;
                      clearEditData();
                      headerEdit = global.language("append");
                      isSaveAllow = true;
                      if (mobileScreen) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                      }
                      fieldFocusNodes[0].focusNode.requestFocus();
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
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = listData[index - 1].guidfixed!;
                  currentListIndex = index + 1;
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
                selectGuid = listData[index + 1].guidfixed!;
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
                color: Colors.white,
                borderRadius: BorderRadius.circular(2),
              ),
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
                        contentPadding: const EdgeInsets.only(
                          top: 0,
                          bottom: 0,
                          left: 0,
                          right: 0,
                        ),
                        border: InputBorder.none,
                        hintText: global.language('search'),
                      ),
                    ),
                  ),
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: const FaIcon(FontAwesomeIcons.font),
                    onPressed: () async {
                      setState(() {
                        global.listDataFontSizeChange();
                      });
                    },
                  ),
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: const Icon(Icons.line_weight),
                    onPressed: () async {
                      setState(() {
                        global.listDataLineSpaceChange();
                      });
                    },
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
                    flex: 5,
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
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("book_bank_code"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("pass_book"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("book_bank_name"),
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

  void switchToEdit(BookBankModel value) {
    setState(() {
      selectGuid = value.guidfixed!;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listObject(int index, BookBankModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    bool selected = selectGuid == value.guidfixed;
    TextStyle textStyle = TextStyle(
      fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
      fontSize: (selected)
          ? global.deviceConfig.listDataFontSize + 2.0
          : global.deviceConfig.listDataFontSize,
    );
    return GestureDetector(
      onTap: () {
        if (showCheckBox == true) {
          setState(() {
            selectGuid = value.guidfixed!;
            if (isCheck == true) {
              guidListChecked.remove(value.guidfixed);
            } else {
              guidListChecked.add(value.guidfixed!);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              "${global.language("chosen")} ${guidListChecked.length} ${global.language("list")}",
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
        decoration: BoxDecoration(
          color: (selectGuid == value.guidfixed)
              ? Colors.cyan[100]
              : (index % 2 == 0)
              ? global.theme.columnAlternateEvenColor
              : global.theme.columnAlternateOddColor,
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
                value.bankcode!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                global.packName(value.banknames!),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                value.bookcode!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                value.passbook!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                global.packName(value.names!),
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
    );
  }

  void saveOrUpdateData() {
    showCheckBox = false;

    if (selectGuid.trim().isEmpty) {
      if (imageFile.isNotEmpty) {
        context.read<BookBankBloc>().add(
          BookBankWithImageSave(
            bookBank: screenData,
            imageFile: imageFile,
            imageWeb: imageWeb,
          ),
        );
      } else {
        context.read<BookBankBloc>().add(BookBankSave(bookBank: screenData));
      }
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    showCheckBox = false;
    List<File> imageFileUpdate = [];
    List<Uint8List> imageWebUpdate = [];
    List<ImagesModel> imageUris = [];
    for (int i = 0; i < imageWeb.length; i++) {
      // print(imageWeb.length);
      // print(imageFile.length);
      // print(screenData.images!.length);
      if (imageWeb[i].isNotEmpty || screenData.images![i].uri != '') {
        imageFileUpdate.add(imageFile[i]);
        imageWebUpdate.add(imageWeb[i]);
        imageUris.add(ImagesModel(uri: screenData.images![i].uri, xorder: i));
      }
    }
    // print("imageWebUpdate.isNotEmpty " + imageWebUpdate.isNotEmpty.toString());
    if (imageWebUpdate.isNotEmpty) {
      context.read<BookBankBloc>().add(
        BookBankWithImageUpdate(
          guid: guid,
          bookBank: screenData,
          imageFiles: imageFile,
          imagesUris: imageUris,
          imageWeb: imageWeb,
        ),
      );
    } else {
      screenData.images = [];
      context.read<BookBankBloc>().add(
        BookBankUpdate(guid: guid, bookBank: screenData),
      );
    }
  }

  void findFocusNext(int index) {
    focusNodeIndex = index;
    do {
      focusNodeIndex++;
      if (focusNodeIndex > focusNodeMax) {
        focusNodeIndex = 0;
      }
    } while (fieldFocusNodes[focusNodeIndex].isReadOnly);
    // print("findFocusNext=$focusNodeIndex");
    refreshFocus = true;
  }

  void getDataToEditScreen(BookBankModel bookBank) {
    isDataChange = false;
    selectGuid = bookBank.guidfixed!;
    screenData.bankcode = bookBank.bankcode;
    screenData.banknames = bookBank.banknames;
    screenData.bookcode = bookBank.bookcode;
    screenData.passbook = bookBank.passbook;
    screenData.images = bookBank.images;
    screenData.accountcode = bookBank.accountcode;
    screenData.accountname = bookBank.accountname;
    screenData.bankbranch = bookBank.bankbranch;

    bookBankCode.text = screenData.bookcode!;

    screenData.names = [];
    //เพิ่มภาษาตาม config
    for (var lang in languageList) {
      screenData.names!.add(LanguageDataModel(code: lang.code!, name: ""));
    }

    //ใส่ value ตามภาษา
    for (var data in bookBank.names!) {
      for (var ele in screenData.names!) {
        if (data.code == ele.code) {
          ele.name = data.name;
        }
      }
    }

    //เก็บค่าที่ไม่ได้เปิดใช้งานภาษาเข้าทาง array
    for (var defaultValueLang in bookBank.names!) {
      LanguageDataModel result = screenData.names!.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        screenData.names!.add(defaultValueLang);
      }
    }

    imageWeb = [];
    imageFile = [];

    for (int i = 0; i < bookBank.images!.length; i++) {
      imageWeb.add(Uint8List(0));
      imageFile.add(File(''));
    }
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];
    focusNodeMax = 0;

    formWidgets.add(
      Container(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        width: double.infinity,
        height: 50,
        child: ElevatedButton(
          onPressed: () {
            Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => const BankSelectScreen(word: ''),
              ),
            ).then((value) {
              setState(() {
                if (value.code != "") {
                  screenData.bankcode = value.code;
                  screenData.banknames = value.names;
                }
              });
            });
          },
          child: Text(
            (screenData.bankcode!.isEmpty)
                ? global.language("select_bank")
                : "${screenData.bankcode} : ${global.packName(screenData.banknames!)}",
          ),
        ),
      ),
    );

    /// รหัสสมุดเงินฝาก
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          enabled: screenData.bookcode!.isEmpty,
          onEditingComplete: () {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: bookBankCode,
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.bookcode = value.toUpperCase();
            bookBankCode.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: bookBankCode.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("book_bank_code"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return global.language('this_field_is_required');
            }

            return null;
          },
        ),
      ),
    );

    /// เลขที่สมุดเงินฝาก
    focusNodeMax++;
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          onEditingComplete: () {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.passbook),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[0-9]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.passbook = value.toUpperCase();
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("pass_book"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return global.language('this_field_is_required');
            }

            return null;
          },
        ),
      ),
    );
    focusNodeMax++;
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          onEditingComplete: () {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.bankbranch),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.bankbranch = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("pass_book_branch"),
          ),
        ),
      ),
    );

    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel names = screenData.names!.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (names.code == '') {
        screenData.names!.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.names![languageIndex].name = value;
            },
            onSubmitted: (value) {
              if (kIsWeb) {
                findFocusNext(focusNodeIndex);
              }
            },
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData.names![languageIndex].name,
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("book_bank_name")} (${getLangName(screenData.names![languageIndex].code)})",
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      Container(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        width: double.infinity,
        height: 50,
        child: ElevatedButton(
          style: ElevatedButton.styleFrom(
            backgroundColor: Colors.green, // Set the background color here
          ),
          onPressed: () {
            Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => const AccountChartSelectScreen(word: ''),
              ),
            ).then((value) {
              setState(() {
                if (value.accountcode != "") {
                  screenData.accountcode = value.accountcode;
                  screenData.accountname = value.accountname;
                }
              });
            });
          },
          child: Row(
            children: [
              const Icon(
                Icons.search, // Replace with the desired icon
                color: Colors.white, // Set the icon color
              ),
              SizedBox(width: 8), // Add spacing between the icon and text
              Text(
                (screenData.accountcode == "")
                    ? global.language("select_accountchart")
                    : "${screenData.accountcode} : ${screenData.accountname}",
              ),
            ],
          ),
        ),
      ),
    );

    formWidgets.add(
      (screenData.images!.isNotEmpty)
          ? Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
              child: Text(global.language("book_bank_image")),
            )
          : Container(),
    );

    List<Widget> imageList = [];
    for (
      int imageIndex = 0;
      imageIndex < screenData.images!.length;
      imageIndex++
    ) {
      imageList.add(
        Container(
          width: 300,
          padding: const EdgeInsets.only(left: 5, right: 5, bottom: 5, top: 5),
          decoration: BoxDecoration(
            border: Border.all(color: Colors.grey),
            borderRadius: const BorderRadius.all(Radius.circular(5.0)),
          ),
          child: Column(
            children: [
              Row(
                children: [
                  (isEditMode)
                      ? Expanded(
                          child: IconButton(
                            focusNode: FocusNode(skipTraversal: true),
                            onPressed: () async {
                              screenData.images!.removeAt(imageIndex);
                              imageWeb.removeAt(imageIndex);
                              imageFile.removeAt(imageIndex);
                              setState(() {});
                            },
                            icon: const Icon(Icons.delete),
                          ),
                        )
                      : Container(),
                  const SizedBox(width: 5),
                  (isEditMode)
                      ? Expanded(
                          child: IconButton(
                            focusNode: FocusNode(skipTraversal: true),
                            onPressed: (kIsWeb)
                                ? () async {
                                    XFile? image = await imagePicker.pickImage(
                                      source: ImageSource.gallery,
                                      maxHeight: 480,
                                      maxWidth: 640,
                                      imageQuality: 60,
                                    );
                                    if (image != null) {
                                      var f = await image.readAsBytes();
                                      setState(() {
                                        imageWeb[imageIndex] = f;
                                        imageFile[imageIndex] = File(
                                          image.path,
                                        );
                                        FocusScope.of(context).unfocus();
                                      });
                                    }
                                  }
                                : () async {
                                    final XFile? photo = await imagePicker
                                        .pickImage(
                                          source: ImageSource.camera,
                                          maxHeight: 480,
                                          maxWidth: 640,
                                          imageQuality: 60,
                                        );
                                    if (photo != null) {
                                      var f = await photo.readAsBytes();
                                      imageWeb[imageIndex] = f;
                                      imageFile.add(File(photo.path));
                                      setState(() {
                                        FocusScope.of(context).unfocus();
                                      });
                                    }
                                  },
                            icon: const Icon(Icons.folder),
                          ),
                        )
                      : Container(),
                  const SizedBox(width: 5),
                  if (kIsWeb == false)
                    Expanded(
                      child: IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          FocusScope.of(context).unfocus();
                          final XFile? photo = await imagePicker.pickImage(
                            source: ImageSource.camera,
                            maxHeight: 480,
                            maxWidth: 640,
                            imageQuality: 60,
                          );
                          if (photo != null) {
                            var f = await photo.readAsBytes();
                            imageWeb[imageIndex] = f;
                            imageFile.add(File(photo.path));
                            setState(() {});
                          }
                        },
                        icon: const Icon(Icons.camera_alt),
                      ),
                    ),
                ],
              ),
              SizedBox(
                width: 300,
                height: 300,
                child: Stack(
                  children: [
                    DropzoneView(
                      operation: DragOperation.copy,
                      cursor: CursorType.grab,
                      onCreated: (ctrl) => dropZoneController = ctrl,
                      onLoaded: () {},
                      onError: (ev) {},
                      onHover: () {},
                      onLeave: () {},
                      onDrop: (ev) async {
                        final bytes = await dropZoneController.getFileData(ev);
                        setState(() {
                          imageWeb[imageIndex] = bytes;
                        });
                      },
                      onDropMultiple: (ev) async {},
                    ),
                    Center(
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          color: Colors.white,
                          border: Border.all(color: Colors.black),
                          borderRadius: BorderRadius.circular(5),
                          boxShadow: const [
                            BoxShadow(
                              offset: Offset(0, 4),
                              color: Colors.cyan, //edited
                              spreadRadius: 4,
                              blurRadius: 10, //edited
                            ),
                          ],
                          image: (imageWeb[imageIndex].isNotEmpty)
                              ? DecorationImage(
                                  image: MemoryImage(imageWeb[imageIndex]),
                                  fit: BoxFit.fill,
                                )
                              : (screenData.images![imageIndex].uri != '')
                              ? DecorationImage(
                                  image: NetworkImage(
                                    screenData.images![imageIndex].uri,
                                  ),
                                  fit: BoxFit.fill,
                                )
                              : const DecorationImage(
                                  image: AssetImage('assets/img/noimage.png'),
                                ),
                        ),
                        child: const SizedBox(width: 500, height: 500),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      );
    }

    formWidgets.add(Wrap(children: imageList));
    formWidgets.add(const SizedBox(height: 5));
    if (isEditMode) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              setState(() {
                screenData.images!.add(ImagesModel(uri: '', xorder: 0));
                imageWeb.add(Uint8List(0));
                imageFile.add(File(''));
                FocusScope.of(context).unfocus();
              });
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_picture")),
          ),
        ),
      );
    }

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              if (_formKey.currentState!.validate()) {
                saveOrUpdateData();
              }
            },
            icon: Icon(Icons.save),
            label: Text(global.language("save") + ((kIsWeb) ? " (F10)" : "")),
          ),
        ),
      );
    }

    return Scaffold(
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
        title: Text(headerEdit + global.language("book_bank")),
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
                            backgroundColor: Colors.red,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.blue,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<BookBankBloc>().add(
                              BookBankDelete(guid: selectGuid),
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
                  for (
                    int indexLanguage = 2;
                    indexLanguage <= languageList.length;
                    indexLanguage++
                  ) {
                    try {} catch (_) {}
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
      body: RawKeyboardListener(
        focusNode: FocusNode(),
        onKey: (RawKeyEvent event) {
          if (event is RawKeyDownEvent) {
            // print(event.logicalKey);
            if (event.logicalKey == LogicalKeyboardKey.f10) {
              if (_formKey.currentState!.validate()) {
                saveOrUpdateData();
              }
            }
            if (event.logicalKey == LogicalKeyboardKey.tab ||
                event.logicalKey == LogicalKeyboardKey.enter) {
              if (event.isShiftPressed) {
                //findFocusPrev(focusNodeIndex);
              } else {
                findFocusNext(focusNodeIndex);
              }
            }
          }
        },
        child: SingleChildScrollView(
          controller: editScrollController,
          child: Container(
            width: double.infinity,
            padding: const EdgeInsets.only(top: 10, bottom: 15),
            child: Form(
              child: Form(
                key: _formKey,
                child: Column(children: formWidgets),
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
      guidListChecked.clear();
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<BookBankBloc, BookBankState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is BookBankLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.bookBanks.isNotEmpty) {
                        listData.addAll(state.bookBanks);
                      }
                    });
                  }
                  if (state is BookBankLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is BookBankSaveSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        global.language("save_success"),
                        Colors.blue,
                      );
                      clearEditData();
                      listData.clear();
                      loadDataList(searchText);
                    });
                  }
                  if (state is BookBankSaveFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        "//${global.language("not_success_save")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                  // Update
                  if (state is BookBankUpdateSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        global.language("edit_success"),
                        Colors.blue,
                      );
                      clearEditData();
                      listData.clear();
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                      loadDataList(searchText);
                      isSaveAllow = false;
                      // getData(selectGuid);
                    });
                  }
                  if (state is BookBankUpdateFailed) {
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
                  if (state is BookBankDeleteSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("delete_success"),
                        Colors.blue,
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
                  if (state is BookBankDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("not_delete_success"),
                        Colors.blue,
                      );
                      listData.clear();
                      clearEditData();
                      loadDataList(searchText);
                      showCheckBox = false;
                    });
                  }
                  // Get
                  if (state is BookBankGetSuccess) {
                    setState(() {
                      getDataToEditScreen(state.bookBank);
                      if (isEditMode) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
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
              ),
            ],
            child: (constraints.maxWidth > 800)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 8,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: Colors.blue,
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
