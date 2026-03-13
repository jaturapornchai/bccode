import 'package:smlaicloud/bloc/promotion/promotion_bloc.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/promotion_model.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:uuid/uuid.dart';

class PromotionScreen extends StatefulWidget {
  const PromotionScreen({super.key});

  @override
  State<PromotionScreen> createState() => PromotionScreenState();
}

class PromotionScreenState extends State<PromotionScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  List<LanguageModel> languageList = <LanguageModel>[];
  List<PromotionModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late PromotionState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  PromotionModel screenData = PromotionModel();
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  bool loadingData = false;
  late DateTime dateNow = DateTime.now();
  final _formKey = GlobalKey<FormState>();
  TextEditingController codeController = TextEditingController();
  TextEditingController indexController = TextEditingController();
  TextEditingController nameController = TextEditingController();
  TextEditingController discountTextController = TextEditingController();
  TextEditingController fromDateController = TextEditingController();
  TextEditingController toDateController = TextEditingController();
  TextEditingController qtyLimitController = TextEditingController();
  TextEditingController promotionQtyController = TextEditingController();
  TextEditingController limitAmountController = TextEditingController();
  List<TextEditingController> qtyMinmunController = [];
  List<TextEditingController> discountController = [];

  FocusNode codeControllerFocus = FocusNode();

  FocusNode indexControllerFocus = FocusNode();
  FocusNode nameControllerFocus = FocusNode();
  FocusNode discountTextControllerFocus = FocusNode();
  FocusNode qtyLimitControllerFocus = FocusNode();
  FocusNode promotionQtyControllerFocus = FocusNode();
  FocusNode limitAmountControllerFocus = FocusNode();
  FocusNode fromDateControllerFocus = FocusNode();
  FocusNode toDateControllerFocus = FocusNode();
  List<FocusNode> qtyMinmunControllerFocus = [];
  List<FocusNode> discountControllerFocus = [];

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
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

    listScrollController.addListener(onScrollList);

    global.promotionList = [
      PromotionListModel(code: 1, name: global.language("promotion_type_1")),
      PromotionListModel(code: 2, name: global.language("promotion_type_2")),
      PromotionListModel(code: 3, name: global.language("promotion_type_3")),
      PromotionListModel(code: 4, name: global.language("promotion_type_4")),
      PromotionListModel(code: 5, name: global.language("promotion_type_5")),
      PromotionListModel(code: 6, name: global.language("promotion_type_6")),
      PromotionListModel(code: 7, name: global.language("promotion_type_7")),
      PromotionListModel(code: 8, name: global.language("promotion_type_8")),
    ];

    screenData.promotionbarcodeinclude = [];
    screenData.promotionbarcodeinclude.add(
      PromotionBarcodeIncludeModel(
        promotionproduct: [
          // PromotionBarcodeModel(
          //     barcode: "A00001",
          //     name: [LanguageDataModel(code: "th", name: "สินค้า A"), LanguageDataModel(code: "EN", name: "Product A")],
          //     qty: 2,
          //     unitCode: "pcs",
          //     unitName: [LanguageDataModel(code: "th", name: "ชิ้น"), LanguageDataModel(code: "EN", name: "Piece")],
          //     price: 100),
          // PromotionBarcodeModel(
          //     barcode: "A00002",
          //     name: [LanguageDataModel(code: "th", name: "สินค้า A"), LanguageDataModel(code: "EN", name: "Product A")],
          //     qty: 2,
          //     unitCode: "pcs",
          //     unitName: [LanguageDataModel(code: "th", name: "ชิ้น"), LanguageDataModel(code: "EN", name: "Piece")],
          //     price: 100),
        ],
        includeproduct: [],
      ),
    );

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();

    codeController.dispose();
    nameController.dispose();
    fromDateController.dispose();
    toDateController.dispose();
    for (int i = 0; i < qtyMinmunController.length; i++) {
      qtyMinmunController[i].dispose();
    }
    for (int i = 0; i < discountController.length; i++) {
      discountController[i].dispose();
    }

    codeControllerFocus.dispose();
    nameControllerFocus.dispose();
    fromDateControllerFocus.dispose();
    toDateControllerFocus.dispose();
    for (int i = 0; i < qtyMinmunControllerFocus.length; i++) {
      qtyMinmunControllerFocus[i].dispose();
    }
    for (int i = 0; i < discountControllerFocus.length; i++) {
      discountControllerFocus[i].dispose();
    }

    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<PromotionBloc>().add(
      PromotionLoadList(
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

  void clearEditData() {
    qtyMinmunController = [];
    discountController = [];
    codeController.text = "";
    indexController.text = "";
    nameController.text = "";
    discountTextController.text = "";
    fromDateController.text = "";
    toDateController.text = "";
    qtyLimitController.text = "";
    promotionQtyController.text = "";
    limitAmountController.text = "";

    screenData = PromotionModel();
    screenData.code = "P-${Uuid().v4().split("-")[1].toUpperCase()}";

    isDataChange = false;
    setState(() {
      loadDataToScreen();
    });
  }

  void loadDataToScreen() {
    DateTime fromDateFormat = DateTime.parse(
      screenData.fromDate.toIso8601String(),
    );
    DateTime toDateFormat = DateTime.parse(screenData.toDate.toIso8601String());
    if (screenData.datebegin.isNotEmpty) {
      fromDateFormat = DateTime.parse(screenData.datebegin);
      toDateFormat = DateTime.parse(screenData.dateend);
    }

    codeController.text = screenData.code;

    indexController.text = screenData.index.toString();
    discountTextController.text = screenData.discounttext.toString();
    qtyLimitController.text = screenData.limitqty.toString();
    promotionQtyController.text = screenData.promotionqty.toString();
    limitAmountController.text = screenData.limitamount.toString();
    fromDateController.text = DateFormat(
      'dd/MM/yyyy',
    ).format(fromDateFormat.toLocal());
    toDateController.text = DateFormat(
      'dd/MM/yyyy',
    ).format(toDateFormat.toLocal());
  }

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
  }

  void formDateSelect(BuildContext context) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: DateTime.parse(screenData.fromDate.toIso8601String()),
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );

    if (pickedDate != null && pickedDate != dateNow) {
      setState(() {
        screenData.fromDate = pickedDate.toLocal();

        fromDateController.text = DateFormat(
          'dd/MM/yyyy',
        ).format(screenData.fromDate);
      });
    }
  }

  void toDateSelect(BuildContext context) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: DateTime.parse(screenData.toDate.toIso8601String()),
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );

    if (pickedDate != null && pickedDate != dateNow) {
      setState(() {
        screenData.toDate = pickedDate.toLocal();

        toDateController.text = DateFormat(
          'dd/MM/yyyy',
        ).format(DateTime.parse(screenData.toDate.toIso8601String()));
      });
    }
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
    context.read<PromotionBloc>().add(PromotionGet(guid: guid));
  }

  void switchToEdit(PromotionModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('promotion')),
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
          const ManualButton(path: 'promotion'),
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
                            backgroundColor: global.theme.primaryColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<PromotionBloc>().add(
                              PromotionDeleteMany(guid: guidListChecked),
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
                  selectGuid = listData[index - 1].guidfixed;
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
                selectGuid = listData[index + 1].guidfixed;
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
                      onSubmitted: (value) {},
                      onChanged: (value) {
                        debouncer.run(() {
                          setState(() {
                            listData = [];
                          });
                          loadDataList(value);
                        });
                      },
                      autofocus: false,
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
              key: headerKey,
              padding: EdgeInsets.only(
                left: 10,
                right: 10,
                top: 5,
                bottom: 5,
              ),
              color: global.theme.columnHeaderColor,
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("promotion_code"),
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
                      global.language("promotion_name"),
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
                  color: global.theme.primaryColor,
                  size: 50,
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget listObject(int index, PromotionModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
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
              guidListChecked.remove(value.guidfixed);
            } else {
              guidListChecked.add(value.guidfixed);
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
                selectGuid = value.guidfixed;
                getData(selectGuid);
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
              flex: 5,
              child: Text(
                value.code,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                global.activeLangName(value.name),
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

  void saveOrUpdateData() {
    showCheckBox = false;

    PromotionModel promotionModel = PromotionModel(
      guidfixed: screenData.guidfixed,
      code: codeController.text,
      name: screenData.name,
      datebegin: screenData.fromDate.toUtc().toIso8601String(),
      dateend: screenData.toDate.toUtc().toIso8601String(),
      promotiontype: screenData.promotiontype,
      index: int.tryParse(indexController.text),
      promotionbarcodeinclude: screenData.promotionbarcodeinclude,
      customeronly: screenData.customeronly,
      discounttext: discountTextController.text,
      limitqty: double.tryParse(qtyLimitController.text) ?? 0,
      promotionqty: double.tryParse(promotionQtyController.text) ?? 0,
      limitamount: double.tryParse(limitAmountController.text) ?? 0,
    );

    if (selectGuid.trim().isEmpty) {
      // print(promotionModel.toJson());
      context.read<PromotionBloc>().add(
        PromotionSave(promotionModel: promotionModel),
      );
    } else {
      updateData(selectGuid, promotionModel);
    }
  }

  void updateData(String guid, PromotionModel promotionModel) {
    showCheckBox = false;

    context.read<PromotionBloc>().add(
      PromotionUpdate(guid: guid, promotionModel: promotionModel),
    );
  }

  Future<ProductBarcodeModel> barcodeSearch() async {
    ProductBarcodeModel res = ProductBarcodeModel(guidfixed: '', itemcode: '');
    res = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const BarcodeSearchScreen(word: '', screen: ''),
      ),
    );
    return res;
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: InputDecorator(
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            labelText: global.language("select_promotion"),
          ),
          child: DropdownButtonHideUnderline(
            child: DropdownButton<int>(
              value: screenData.promotiontype,
              icon: const Icon(Icons.arrow_drop_down),
              style: const TextStyle(color: Colors.deepPurple),
              underline: Container(color: Colors.deepPurpleAccent),
              onChanged: (int? value) {
                setState(() {
                  screenData.promotiontype = value!;
                  indexControllerFocus.requestFocus();
                });
              },
              isDense: true,
              isExpanded: true,
              items: global.promotionList.map<DropdownMenuItem<int>>((
                PromotionListModel value,
              ) {
                return DropdownMenuItem<int>(
                  value: value.code,
                  child: Text(value.name),
                );
              }).toList(),
            ),
          ),
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          focusNode: indexControllerFocus,
          textAlign: TextAlign.left,
          controller: indexController,
          keyboardType: TextInputType.number,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.digitsOnly,
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.index = int.parse(value);
            indexController.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: indexController.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("index"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return 'This field is required';
            }
            return null;
          },
          onEditingComplete: () {
            codeControllerFocus.requestFocus();
          },
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          focusNode: codeControllerFocus,
          textAlign: TextAlign.left,
          controller: codeController,
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.code = value.toUpperCase();
            codeController.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: codeController.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("promotion_code"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return 'This field is required';
            }
            return null;
          },
          onEditingComplete: () {
            nameControllerFocus.requestFocus();
          },
        ),
      ),
    );

    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel nameObj = screenData.name.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (nameObj.code == '') {
        screenData.name.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
    }
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel nameObj = screenData.name.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (nameObj.code != '') {
        formWidgets.add(
          Padding(
            padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
            child: TextFormField(
              onChanged: (value) {
                nameObj.name = value;
              },
              textAlign: TextAlign.left,
              controller: TextEditingController(text: nameObj.name),
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: const OutlineInputBorder(),
                labelText:
                    "${global.language("promotion_name")} (${getLangName(nameObj.code)})",
              ),
              validator: (value) {
                if (languageIndex == 0) {
                  if (value == null || value.isEmpty) {
                    return 'This field is required';
                  }
                }

                return null;
              },
            ),
          ),
        );
      }
    }

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language("from_date"),
                  suffixIcon: Row(
                    mainAxisAlignment:
                        MainAxisAlignment.spaceBetween, // added line
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        icon: const Icon(Icons.calendar_today),
                        onPressed: () {
                          formDateSelect(context);
                        },
                      ),
                    ],
                  ),
                ),
                controller: fromDateController,
                onChanged: (value) {
                  setState(() {
                    try {
                      List<String> valueSplit = value
                          .replaceAll(".", "/")
                          .split("/");
                      if (valueSplit.length == 3) {
                        if (valueSplit[2].length == 2) {
                          valueSplit[2] = '25${valueSplit[2]}';
                        }
                        int year = int.tryParse(valueSplit[2]) ?? 0;
                        year = year - 543;
                        int month = int.tryParse(valueSplit[1]) ?? 0;
                        int day = int.tryParse(valueSplit[0]) ?? 0;
                        value =
                            "$year-${month.toString().padLeft(2, '0')}-${day.toString().padLeft(2, '0')}";
                      }

                      if (global.isValidDate(value)) {
                        screenData.fromDate = DateTime.parse(value).toLocal();
                      }
                    } catch (e) {
                      // print(e);
                    }
                  });
                },
                onSubmitted: (value) => {
                  fromDateController.text = DateFormat(
                    'dd/MM/yyyy',
                  ).format(DateTime.parse(screenData.fromDate.toString())),
                },
              ),
            ),
            Text(" - "),
            Expanded(
              child: TextField(
                focusNode: toDateControllerFocus,
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language("to_date"),
                  suffixIcon: Row(
                    mainAxisAlignment:
                        MainAxisAlignment.spaceBetween, // added line
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        icon: const Icon(Icons.calendar_today),
                        onPressed: () {
                          toDateSelect(context);
                        },
                      ),
                    ],
                  ),
                ),
                controller: toDateController,
                onChanged: (value) {
                  setState(() {
                    try {
                      List<String> valueSplit = value
                          .replaceAll(".", "/")
                          .split("/");
                      if (valueSplit.length == 3) {
                        if (valueSplit[2].length == 2) {
                          valueSplit[2] = '25${valueSplit[2]}';
                        }
                        int year = int.tryParse(valueSplit[2]) ?? 0;
                        year = year - 543;
                        int month = int.tryParse(valueSplit[1]) ?? 0;
                        int day = int.tryParse(valueSplit[0]) ?? 0;
                        value =
                            "$year-${month.toString().padLeft(2, '0')}-${day.toString().padLeft(2, '0')}";
                      }

                      if (global.isValidDate(value)) {
                        screenData.toDate = DateTime.parse(value).toLocal();
                      }
                    } catch (e) {
                      // print(e);
                    }
                  });
                },
                onSubmitted: (value) => {
                  toDateController.text = DateFormat(
                    'dd/MM/yyyy',
                  ).format(DateTime.parse(screenData.toDate.toString())),
                },
              ),
            ),
          ],
        ),
      ),
    );
    if (screenData.promotiontype != 1) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextFormField(
            focusNode: discountTextControllerFocus,
            textAlign: TextAlign.left,
            controller: discountTextController,
            textCapitalization: TextCapitalization.characters,
            onChanged: (value) {
              isDataChange = true;
            },
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("discount_text"),
            ),
            validator: (value) {
              return null;
            },
            onEditingComplete: () {
              qtyLimitControllerFocus.requestFocus();
            },
          ),
        ),
      );
    }

    if (screenData.promotiontype != 1 &&
        screenData.promotiontype != 2 &&
        screenData.promotiontype != 6 &&
        screenData.promotiontype != 7 &&
        screenData.promotiontype != 8) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextFormField(
            focusNode: qtyLimitControllerFocus,
            textAlign: TextAlign.left,
            controller: qtyLimitController,
            keyboardType: TextInputType.number,
            onChanged: (value) {
              isDataChange = true;
            },
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("qty_limit"),
            ),
            validator: (value) {
              return null;
            },
            onEditingComplete: () {
              promotionQtyControllerFocus.requestFocus();
            },
          ),
        ),
      );
    }

    if (screenData.promotiontype != 1 &&
        screenData.promotiontype != 2 &&
        screenData.promotiontype != 3 &&
        screenData.promotiontype != 4 &&
        screenData.promotiontype != 5 &&
        screenData.promotiontype != 6 &&
        screenData.promotiontype != 7 &&
        screenData.promotiontype != 8) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextFormField(
            focusNode: promotionQtyControllerFocus,
            textAlign: TextAlign.left,
            controller: promotionQtyController,
            keyboardType: TextInputType.number,
            onChanged: (value) {
              isDataChange = true;
            },
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("promotion_qty"),
            ),
            validator: (value) {
              return null;
            },
            onEditingComplete: () {
              limitAmountControllerFocus.requestFocus();
            },
          ),
        ),
      );
    }
    if (screenData.promotiontype != 1 &&
        screenData.promotiontype != 2 &&
        screenData.promotiontype != 3 &&
        screenData.promotiontype != 4 &&
        screenData.promotiontype != 5) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextFormField(
            focusNode: limitAmountControllerFocus,
            textAlign: TextAlign.left,
            controller: limitAmountController,
            keyboardType: TextInputType.number,
            onChanged: (value) {
              isDataChange = true;
            },
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("limit_amount"),
            ),
            validator: (value) {
              return null;
            },
            onEditingComplete: () {},
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: EdgeInsets.symmetric(horizontal: 10, vertical: 15),
        child: Row(
          children: [
            Checkbox(
              value: screenData.customeronly == 0 ? false : true,
              onChanged: (bool? value) {
                setState(() {
                  screenData.customeronly = value == false ? 0 : 1;
                });
              },
            ),
            SizedBox(width: 8), // เว้นระยะห่างระหว่าง Checkbox และข้อความ
            Text(
              global.language("customer_only"), // แสดงสถานะปัจจุบัน
              style: TextStyle(fontSize: 16),
            ),
          ],
        ),
      ),
    );
    if (screenData.promotiontype != 7 && screenData.promotiontype != 8) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Column(
            children: [
              Row(
                children: [
                  Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      global.language("promotion_include"),
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                  ),
                  SizedBox(width: 10),
                  ElevatedButton.icon(
                    icon: Icon(Icons.add),
                    label: Text(global.language("add_product")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.blue, // สีปุ่ม
                      foregroundColor: Colors.white, // สีข้อความ
                    ),
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) =>
                              const BarcodeSearchScreen(word: "", screen: ""),
                        ),
                      ).then((value) {
                        if (value != null) {
                          setState(() {
                            ProductBarcodeModel result = value;

                            // ตรวจสอบว่าบาร์โค้ดซ้ำหรือไม่
                            bool isDuplicate = screenData
                                .promotionbarcodeinclude[0]
                                .promotionproduct
                                .any(
                                  (product) =>
                                      product.barcode == result.barcode,
                                );

                            if (!isDuplicate && result.guidfixed.isNotEmpty) {
                              screenData
                                  .promotionbarcodeinclude[0]
                                  .promotionproduct
                                  .add(
                                    PromotionBarcodeModel(
                                      barcode: result.barcode,
                                      unitcode: result.itemunitcode,
                                      unitname: result.itemunitnames,
                                      name: result.names,
                                      qty: 1,
                                    ),
                                  );
                            } else {
                              // แสดงข้อความแจ้งเตือน (ถ้าจำเป็น)
                              global.showInfoSnackBar(context, global.language("barcode_already_exists"));
                            }
                          });
                        }
                      });
                    },
                  ),
                ],
              ),
              SizedBox(height: 5),
              Container(
                decoration: BoxDecoration(
                  border: Border.all(
                    color: global.theme.textSecondaryColor,
                    width: 1,
                  ), // Border สีเทา
                  borderRadius: BorderRadius.circular(10), // ขอบมน
                ),
                child: SizedBox(
                  height: 250, // เพิ่มความสูงเพื่อให้มีพื้นที่มากขึ้น
                  child: ListView.builder(
                    itemCount: screenData
                        .promotionbarcodeinclude[0]
                        .promotionproduct
                        .length,
                    itemBuilder: (context, index) {
                      final product = screenData
                          .promotionbarcodeinclude[0]
                          .promotionproduct[index];
                      return Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8.0,
                          vertical: 4.0,
                        ),
                        child: Card(
                          elevation: 2,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(
                              8,
                            ), // ขอบมนของ Card
                          ),
                          child: ListTile(
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 8,
                            ),
                            title: Text(
                              "${global.activeLangName(product.name)} (${product.barcode})",
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ), // ปรับตัวอักษรให้เด่น
                            ),
                            subtitle: Row(
                              children: [
                                Text(global.language("qty")),
                                SizedBox(width: 8),
                                SizedBox(
                                  width: 60,
                                  height: 30,
                                  child: TextField(
                                    controller: TextEditingController(
                                      text: product.qty
                                          .toString(), // Set initial value
                                    ),
                                    keyboardType: TextInputType.number,
                                    decoration: InputDecoration(
                                      border: OutlineInputBorder(),
                                      contentPadding: EdgeInsets.symmetric(
                                        horizontal: 8,
                                        vertical: 5,
                                      ),
                                    ),
                                    onChanged: (value) {
                                      product.qty =
                                          double.tryParse(value) ?? product.qty;
                                    },
                                    onSubmitted: (value) {
                                      setState(() {
                                        product.qty =
                                            double.tryParse(value) ??
                                            product.qty;
                                      });
                                    },
                                  ),
                                ),
                                SizedBox(width: 8),
                                Text(global.activeLangName(product.unitname)),
                                if (screenData.promotiontype == 5)
                                  SizedBox(width: 8),
                                if (screenData.promotiontype == 5)
                                  Text(global.language("discount")),
                                if (screenData.promotiontype == 5)
                                  SizedBox(width: 8),
                                if (screenData.promotiontype == 5)
                                  SizedBox(
                                    width: 60,
                                    height: 30,
                                    child: TextField(
                                      controller: TextEditingController(
                                        text: product
                                            .discounttext, // Set initial value
                                      ),
                                      decoration: InputDecoration(
                                        border: OutlineInputBorder(),
                                        contentPadding: EdgeInsets.symmetric(
                                          horizontal: 8,
                                          vertical: 5,
                                        ),
                                      ),
                                      onChanged: (value) {
                                        product.discounttext = value;
                                      },
                                      onSubmitted: (value) {
                                        setState(() {
                                          product.discounttext = value;
                                        });
                                      },
                                    ),
                                  ),
                              ],
                            ),
                            trailing: IconButton(
                              icon: Icon(Icons.delete, color: Colors.red),
                              onPressed: () => _removeItem(index),
                            ),
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ),
            ],
          ),
        ),
      );
    }
    if (screenData.promotiontype != 7 &&
        screenData.promotiontype != 8 &&
        screenData.promotiontype != 4 &&
        screenData.promotiontype != 5 &&
        screenData.promotiontype != 2 &&
        screenData.promotiontype != 3) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Column(
            children: [
              Row(
                children: [
                  Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      global.language("include_product"),
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                  ),
                  SizedBox(width: 10),
                  ElevatedButton.icon(
                    icon: Icon(Icons.add),
                    label: Text(global.language("add_product")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.blue, // สีปุ่ม
                      foregroundColor: Colors.white, // สีข้อความ
                    ),
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) =>
                              const BarcodeSearchScreen(word: "", screen: ""),
                        ),
                      ).then((value) {
                        if (value != null) {
                          setState(() {
                            ProductBarcodeModel result = value;

                            // ตรวจสอบว่าบาร์โค้ดซ้ำหรือไม่
                            bool isDuplicate = screenData
                                .promotionbarcodeinclude[0]
                                .includeproduct
                                .any(
                                  (product) =>
                                      product.barcode == result.barcode,
                                );

                            if (!isDuplicate && result.guidfixed.isNotEmpty) {
                              screenData
                                  .promotionbarcodeinclude[0]
                                  .includeproduct
                                  .add(
                                    PromotionBarcodeModel(
                                      barcode: result.barcode,
                                      unitcode: result.itemunitcode,
                                      unitname: result.itemunitnames,
                                      name: result.names,
                                      qty: 1,
                                    ),
                                  );
                            } else {
                              // แสดงข้อความแจ้งเตือน (ถ้าจำเป็น)
                              global.showInfoSnackBar(context, global.language("barcode_already_exists"));
                            }
                          });
                        }
                      });
                    },
                  ),
                ],
              ),
              SizedBox(height: 5),
              Container(
                decoration: BoxDecoration(
                  border: Border.all(
                    color: global.theme.textSecondaryColor,
                    width: 1,
                  ), // Border สีเทา
                  borderRadius: BorderRadius.circular(10), // ขอบมน
                ),
                child: SizedBox(
                  height: 250, // เพิ่มความสูงเพื่อให้มีพื้นที่มากขึ้น
                  child: ListView.builder(
                    itemCount: screenData
                        .promotionbarcodeinclude[0]
                        .includeproduct
                        .length,
                    itemBuilder: (context, index) {
                      final product = screenData
                          .promotionbarcodeinclude[0]
                          .includeproduct[index];
                      return Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8.0,
                          vertical: 4.0,
                        ),
                        child: Card(
                          elevation: 2,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(
                              8,
                            ), // ขอบมนของ Card
                          ),
                          child: ListTile(
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 8,
                            ),
                            title: Text(
                              "${global.activeLangName(product.name)} (${product.barcode})",
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ), // ปรับตัวอักษรให้เด่น
                            ),
                            subtitle: Row(
                              children: [
                                Text(global.language("qty")),
                                SizedBox(width: 8),
                                SizedBox(
                                  width: 60,
                                  height: 30,
                                  child: TextField(
                                    controller: TextEditingController(
                                      text: product.qty
                                          .toString(), // Set initial value
                                    ),
                                    keyboardType: TextInputType.number,
                                    decoration: InputDecoration(
                                      border: OutlineInputBorder(),
                                      contentPadding: EdgeInsets.symmetric(
                                        horizontal: 8,
                                        vertical: 5,
                                      ),
                                    ),
                                    onChanged: (value) {
                                      product.qty =
                                          double.tryParse(value) ?? product.qty;
                                    },
                                    onSubmitted: (value) {
                                      setState(() {
                                        product.qty =
                                            double.tryParse(value) ??
                                            product.qty;
                                      });
                                    },
                                  ),
                                ),
                                SizedBox(width: 8),
                                Text(global.activeLangName(product.unitname)),
                                if (screenData.promotiontype == 5)
                                  SizedBox(width: 8),
                                if (screenData.promotiontype == 5)
                                  Text(global.language("discount")),
                                if (screenData.promotiontype == 5)
                                  SizedBox(width: 8),
                                if (screenData.promotiontype == 5)
                                  SizedBox(
                                    width: 60,
                                    height: 30,
                                    child: TextField(
                                      controller: TextEditingController(
                                        text: product
                                            .discounttext, // Set initial value
                                      ),
                                      decoration: InputDecoration(
                                        border: OutlineInputBorder(),
                                        contentPadding: EdgeInsets.symmetric(
                                          horizontal: 8,
                                          vertical: 5,
                                        ),
                                      ),
                                      onChanged: (value) {
                                        product.discounttext = value;
                                      },
                                      onSubmitted: (value) {
                                        setState(() {
                                          product.discounttext = value;
                                        });
                                      },
                                    ),
                                  ),
                              ],
                            ),
                            trailing: IconButton(
                              icon: Icon(Icons.delete, color: Colors.red),
                              onPressed: () => _removeIncludeItem(index),
                            ),
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ),
            ],
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
        title: Text(headerEdit + global.language("promotion")),
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
                            backgroundColor: Colors.red,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.primaryColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<PromotionBloc>().add(
                              PromotionDelete(guid: selectGuid),
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
          }
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
            padding: const EdgeInsets.only(top: 10, bottom: 15),
            child: Form(
              key: _formKey,
              child: Column(children: formWidgets),
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

  void _removeItem(int index) {
    setState(() {
      screenData.promotionbarcodeinclude[0].promotionproduct.removeAt(index);
    });
  }

  void _removeIncludeItem(int index) {
    setState(() {
      screenData.promotionbarcodeinclude[0].includeproduct.removeAt(index);
    });
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    if (showCheckBox == false) {
      guidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<PromotionBloc, PromotionState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is PromotionLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.promotions.isNotEmpty) {
                        listData.addAll(state.promotions);
                        // สร้าง keys ใหม่ให้ตรงกับจำนวน items
                        listKeys.clear();
                        for (int i = 0; i < listData.length; i++) {
                          listKeys.add(GlobalKey());
                        }
                      }
                    });
                  }
                  if (state is PromotionLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is PromotionSaveSuccess) {
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
                  if (state is PromotionSaveFailed) {
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
                  if (state is PromotionUpdateSuccess) {
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
                      getData(selectGuid);
                    });
                  }
                  if (state is PromotionUpdateFailed) {
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
                  if (state is PromotionDeleteSuccess) {
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
                  if (state is PromotionDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("delete_success"),
                        Colors.blue,
                      );
                      listData.clear();
                      clearEditData();
                      loadDataList(searchText);
                      showCheckBox = false;
                    });
                  }
                  // Get
                  if (state is PromotionGetSuccess) {
                    setState(() {
                      isDataChange = false;
                      screenData = state.promotions;

                      // for (int i = 0; i < screenData.details!.length; i++) {
                      //   qtyMinmunController.add(TextEditingController());
                      //   qtyMinmunControllerFocus.add(FocusNode());
                      //   discountController.add(TextEditingController());
                      //   discountControllerFocus.add(FocusNode());
                      //   qtyMinmunController[i].text = screenData.details![i].minimum.toString();
                      //   discountController[i].text = screenData.details![i].discount.toString();
                      // }

                      loadDataToScreen();

                      if (isEditMode) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                        setState(() {});
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
