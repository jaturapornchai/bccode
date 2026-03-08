import 'package:smlaicloud/bloc/chart_account/chart_account_bloc.dart';
import 'package:smlaicloud/bloc/doc_format/doc_format_bloc.dart';
import 'package:smlaicloud/model/accountbook_model.dart';
import 'package:smlaicloud/model/accountgroup_model.dart';
import 'package:smlaicloud/model/doc_format_model.dart';
import 'package:smlaicloud/screen_search/accountchart_select_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';

class DocFormatScreen extends StatefulWidget {
  const DocFormatScreen({super.key});

  @override
  State<DocFormatScreen> createState() => DocFormatScreenState();
}

class DocFormatScreenState extends State<DocFormatScreen>
    with SingleTickerProviderStateMixin {
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  List<DocFormatModel> listData = [];
  List<String> guidListChecked = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  String headerEdit = "";
  late MediaQueryData queryData;
  GlobalKey headerKey = GlobalKey();
  bool isEditMode = false;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  bool loadingData = false;
  bool showCheckBox = false;

  final _formKey = GlobalKey<FormState>();

  late DocFormatModel screenData;
  List<DefaultDocFormatModel> defaultData = [];
  List<String> docFormatTypeList = [];

  TextEditingController docFormatCodeController = TextEditingController();
  TextEditingController desciptionController = TextEditingController();
  TextEditingController qtyRunningController = TextEditingController();
  TextEditingController exampleController = TextEditingController();
  TextEditingController docFormatController = TextEditingController();
  FocusNode docFormatCodeFocusNode = FocusNode();

  List<TextEditingController> actionCodeController = [];
  List<TextEditingController> detailController = [];
  List<TextEditingController> debitController = [];
  List<TextEditingController> creditController = [];
  List<TextEditingController> debitShowController = [];
  List<TextEditingController> creditShowController = [];
  List<bool> isEntrySelfAccountController = [];

  List<AccountBookModel> accountBookList = [];
  List<AccountGroupModel> accountGroupList = [];
  AccountGroupModel accountGroupSeleted = AccountGroupModel();
  AccountBookModel accountBookSeleted = AccountBookModel();

  @override
  void initState() {
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.3, max: 0.8)],
    );
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });
    listScrollController.addListener(onScrollList);
    clearEditData();
    context.read<DocFormatBloc>().add(const DocFormatLoadDefault());
    context.read<ChartAccountBloc>().add(const AccountGroupLoad(search: ''));
    context.read<ChartAccountBloc>().add(const AccountBookLoad(search: ''));

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();

    docFormatCodeController.dispose();
    qtyRunningController.dispose();
    exampleController.dispose();
    docFormatController.dispose();

    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;

    context.read<DocFormatBloc>().add(
      DocFormatLoadList(
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
    docFormatTypeList = ['YYMMDD', 'YYMM', 'YYYYMMDD', 'YYYYMM'];

    screenData = DocFormatModel(
      dateformate: docFormatTypeList.first,
      details: [],
      doccode: (defaultData.isNotEmpty) ? defaultData.first.doccode : '',
      docnumber: 4,
      module: (defaultData.isNotEmpty) ? defaultData.first.doccode : '',
      docformat: '',
      isautoformat: false,
      yeartype: 0,
    );

    isDataChange = false;
    loadDataToScreen();
  }

  // load data to edit screen
  void loadDataToScreen() {
    docFormatCodeController.text = screenData.doccode;
    desciptionController.text = screenData.description!;
    qtyRunningController.text = screenData.docnumber.toString();
    docFormatController.text = screenData.docformat;

    if (!screenData.isautoformat) {
      exampleValue();
    }

    actionCodeController = [];
    detailController = [];
    debitController = [];
    creditController = [];
    debitShowController = [];
    creditShowController = [];
    isEntrySelfAccountController = [];

    for (int i = 0; i < screenData.details.length; i++) {
      actionCodeController.add(
        TextEditingController(text: screenData.details[i].actioncode),
      );
      detailController.add(
        TextEditingController(text: screenData.details[i].detail),
      );
      debitController.add(
        TextEditingController(text: screenData.details[i].debit),
      );
      creditController.add(
        TextEditingController(text: screenData.details[i].credit),
      );
      debitShowController.add(
        TextEditingController(
          text:
              '${screenData.details[i].accountdebit!.accountcode} ~ ${screenData.details[i].accountdebit!.accountname}',
        ),
      );
      creditShowController.add(
        TextEditingController(
          text:
              '${screenData.details[i].accountcredit!.accountcode} ~ ${screenData.details[i].accountcredit!.accountname}',
        ),
      );
      isEntrySelfAccountController.add(
        screenData.details[i].isentryselfaccount,
      );
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
    context.read<DocFormatBloc>().add(DocFormatGet(guid: guid));
  }

  void switchToEdit(DocFormatModel value) {
    setState(() {
      selectGuid = value.guidfixed!;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Future<List<AccountGroupModel>> getAccountGroupList(filter) async {
    return accountGroupList;
  }

  Future<List<AccountBookModel>> getAccountBookList(filter) async {
    return accountBookList;
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('doc_format')),
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
                            context.read<DocFormatBloc>().add(
                              DocFormatDeleteMany(guid: guidListChecked),
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
              key: headerKey,
              padding: const EdgeInsets.only(
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
                      global.language("menu"),
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
                      global.language("doc_format_code"),
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
                      global.language("desciption"),
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

  Widget listObject(int index, DocFormatModel value, bool showCheckBox) {
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
                global.language(
                  defaultData
                      .firstWhere(
                        (item) => item.doccode == value.module,
                        orElse: () => DefaultDocFormatModel(
                          name: '',
                          dateformate: '',
                          doccode: '',
                          docnumber: 0,
                        ),
                      )
                      .name,
                ),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                value.doccode,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                value.description!,
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

    screenData.doccode = docFormatCodeController.text;
    screenData.description = desciptionController.text;
    screenData.docnumber = int.parse(qtyRunningController.text);
    screenData.docformat = docFormatController.text;

    for (int i = 0; i < screenData.details.length; i++) {
      screenData.details[i].actioncode = actionCodeController[i].text;
      screenData.details[i].detail = detailController[i].text;
      screenData.details[i].debit = debitController[i].text;
      screenData.details[i].credit = creditController[i].text;
      screenData.details[i].isentryselfaccount =
          isEntrySelfAccountController[i];
    }
    if (selectGuid.trim().isEmpty) {
      context.read<DocFormatBloc>().add(
        DocFormatSave(docFormatModel: screenData),
      );
    } else {
      context.read<DocFormatBloc>().add(
        DocFormatUpdate(guid: selectGuid, docFormatModel: screenData),
      );
    }
  }

  void exampleValue() {
    String formattedDate = '';
    String formattedNumber = '';
    DateTime now = DateTime.now();
    int buddhistYear = now.year + 543;
    if (screenData.yeartype == 0) {
      if (screenData.dateformate == 'YYMMDD') {
        formattedDate = DateFormat('yyMMdd').format(now);
      } else if (screenData.dateformate == 'YYMM') {
        formattedDate = DateFormat('yyMM').format(now);
      } else if (screenData.dateformate == 'YYYYMMDD') {
        formattedDate = DateFormat('yyyyMMdd').format(now);
      } else if (screenData.dateformate == 'YYYYMM') {
        formattedDate = DateFormat('yyyyMM').format(now);
      }
    } else {
      if (screenData.dateformate == 'YYMMDD') {
        formattedDate = DateFormat('yyMMdd').format(now);
        formattedDate = formattedDate.replaceRange(
          0,
          2,
          buddhistYear.toString().substring(2, 4),
        );
      } else if (screenData.dateformate == 'YYMM') {
        formattedDate = DateFormat('yyMM').format(now);
        formattedDate = formattedDate.replaceRange(
          0,
          2,
          buddhistYear.toString().substring(2, 4),
        );
      } else if (screenData.dateformate == 'YYYYMMDD') {
        formattedDate = DateFormat('yyyyMMdd').format(now);
        formattedDate = formattedDate.replaceRange(
          0,
          2,
          buddhistYear.toString().substring(2, 4),
        );
      } else if (screenData.dateformate == 'YYYYMM') {
        formattedDate = DateFormat('yyyyMM').format(now);
        formattedDate = formattedDate.replaceRange(
          0,
          2,
          buddhistYear.toString().substring(2, 4),
        );
      }
    }

    /// qtyRunningController
    int digitCount = int.parse(
      (qtyRunningController.text.isEmpty || qtyRunningController.text == '0')
          ? '4'
          : qtyRunningController.text,
    );
    int runningStart = 1;
    formattedNumber = runningStart.toString().padLeft(digitCount, '0');

    int hashCount = int.parse(qtyRunningController.text);
    String hashedValue = '#'.padLeft(hashCount, '#');

    setState(() {
      exampleController.text =
          '${docFormatCodeController.text}$formattedDate$formattedNumber';
      docFormatController.text =
          '${docFormatCodeController.text}${screenData.dateformate}$hashedValue';
    });
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    formWidgets.add(
      Padding(
        padding: EdgeInsets.all(10.0),
        child: Row(
          children: [
            Expanded(
              flex: 1,
              child: SizedBox(
                height: 50,
                child: InputDecorator(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("select_menu"),
                  ),
                  child: SizedBox(
                    height: double.infinity,
                    child: DropdownButtonHideUnderline(
                      child: DropdownButton<String>(
                        value: screenData.module,
                        icon: const Icon(Icons.arrow_drop_down),
                        style: const TextStyle(color: Colors.deepPurple),
                        underline: Container(color: Colors.deepPurpleAccent),
                        onChanged: (String? value) {
                          setState(() {
                            screenData.module = value!;

                            docFormatCodeController.text = value;

                            exampleValue();
                          });
                        },
                        isDense: true,
                        isExpanded: true,
                        items: defaultData.map<DropdownMenuItem<String>>((
                          DefaultDocFormatModel value,
                        ) {
                          return DropdownMenuItem<String>(
                            value: value.doccode,
                            child: Text(global.language(value.name)),
                          );
                        }).toList(),
                      ),
                    ),
                  ),
                ),
              ),
            ),
            SizedBox(width: 10),
            Expanded(
              flex: 2,
              child: TextFormField(
                controller: desciptionController,
                textAlign: TextAlign.left,
                textCapitalization: TextCapitalization.characters,
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("desciption"),
                ),
              ),
            ),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.all(10.0),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 0,
              groupValue: screenData.yeartype,
              onChanged: (value) {
                setState(() {
                  screenData.yeartype = value as int;
                  exampleValue();
                });
              },
            ),
            Text(global.language("anno_domini")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 1,
              groupValue: screenData.yeartype,
              onChanged: (value) {
                setState(() {
                  screenData.yeartype = value as int;
                  exampleValue();
                });
              },
            ),
            Text(global.language("buddhist_calendar ")),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.all(10.0),
        child: Row(
          children: [
            Expanded(
              flex: 1,
              child: TextFormField(
                enabled: (screenData.isautoformat) ? false : true,
                textCapitalization: TextCapitalization.characters,
                inputFormatters: <TextInputFormatter>[
                  FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
                ],
                onChanged: (value) {
                  screenData.doccode = value.toUpperCase();
                  docFormatCodeController.value = TextEditingValue(
                    text: value.toUpperCase(),
                    selection: docFormatCodeController.selection,
                  );
                  exampleValue();
                },
                controller: docFormatCodeController,
                textAlign: TextAlign.left,
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("doc_format_code"),
                ),
              ),
            ),
            SizedBox(width: 10),
            Expanded(
              flex: 1,
              child: SizedBox(
                height: 50,
                child: InputDecorator(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("doc_format_type"),
                  ),
                  child: SizedBox(
                    height: double.infinity,
                    child: DropdownButtonHideUnderline(
                      child: DropdownButton<String>(
                        value: screenData.dateformate,
                        icon: const Icon(Icons.arrow_drop_down),
                        style: const TextStyle(color: Colors.deepPurple),
                        underline: Container(color: Colors.deepPurpleAccent),
                        onChanged: (!screenData.isautoformat)
                            ? (String? value) {
                                setState(() {
                                  screenData.dateformate = value!;
                                  exampleValue();
                                });
                              }
                            : null,
                        isDense: true,
                        isExpanded: true,
                        items: docFormatTypeList.map<DropdownMenuItem<String>>((
                          value,
                        ) {
                          return DropdownMenuItem<String>(
                            value: value.toString(),
                            child: Text(value.toString()),
                          );
                        }).toList(),
                      ),
                    ),
                  ),
                ),
              ),
            ),
            SizedBox(width: 10),
            Expanded(
              flex: 1,
              child: TextFormField(
                enabled: (screenData.isautoformat) ? false : true,
                controller: qtyRunningController,
                onChanged: (value) {
                  exampleValue();
                },
                textAlign: TextAlign.left,
                inputFormatters: [
                  FilteringTextInputFormatter.allow(
                    RegExp(r'^[0-9]*$'),
                  ), // Allow only numbers
                ],
                keyboardType: TextInputType.number, // Show numerical keyboard
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("qty_running"),
                ),
              ),
            ),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.all(10.0),
        child: Row(
          children: [
            Expanded(
              flex: 1,
              child: TextFormField(
                enabled: (screenData.isautoformat) ? false : true,
                controller: exampleController,
                readOnly: true,
                textAlign: TextAlign.left,
                textCapitalization: TextCapitalization.characters,
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("example"),
                ),
              ),
            ),
            SizedBox(width: 10),
            Expanded(
              flex: 1,
              child: TextFormField(
                focusNode: docFormatCodeFocusNode,
                controller: docFormatController,
                readOnly: (screenData.isautoformat) ? false : true,
                textAlign: TextAlign.left,
                textCapitalization: TextCapitalization.characters,
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("document_format"),
                ),
              ),
            ),
            SizedBox(width: 10),
            Expanded(
              child: CheckboxListTile(
                value: screenData.isautoformat,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.isautoformat = value!;
                    if (value == true) {
                      docFormatCodeFocusNode.requestFocus();
                    } else {
                      exampleValue();
                    }
                  });
                },
                title: Text(global.language("format_manual")),
                subtitle: Text(global.language("format_manual_description")),
              ),
            ),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.all(10.0),
        child: Row(
          children: [
            /* XXXX
            Expanded(
              child: SizedBox(
                height: 50,
                child: DropdownSearch<AccountGroupModel>(
                  enabled: isEditMode,
                  items: (String filter, infiniteScrollProps) => getAccountGroupList(filter),
                  compareFn: (item, selectedItem) => item.code == selectedItem.code!,
                  itemAsString: (AccountGroupModel? accountgrouplist) {
                    if (accountgrouplist!.code!.isEmpty) return '';
                    return "${accountgrouplist.code!}~${accountgrouplist.name1!}";
                  },
                  decoratorProps: DropDownDecoratorProps(
                    decoration: InputDecoration(
                      border: const OutlineInputBorder(),
                      contentPadding: const EdgeInsets.only(left: 10, top: 0, bottom: 0, right: 10),
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      labelText: global.language("account_group"),
                    ),
                  ),
                  onChanged: (isEditMode)
                      ? (AccountGroupModel? value) {
                          setState(() {
                            screenData.accountgroup = value!.code;
                            accountGroupSeleted = value;
                          });
                        }
                      : null,
                  popupProps: const PopupPropsMultiSelection.dialog(
                    showSearchBox: true,
                    showSelectedItems: true,
                  ),
                  selectedItem: accountGroupSeleted,
                ),
              ),
            ), */
            const SizedBox(width: 10),
            /* XXXX
            Expanded(
              child: SizedBox(
                height: 50,
                child: DropdownSearch<AccountBookModel>(
                  enabled: isEditMode,
                  asyncItems: (String filter) => getAccountBookList(filter),
                  compareFn: (item, selectedItem) => item.code == selectedItem.code!,
                  itemAsString: (AccountBookModel? accountbooklist) {
                    if (accountbooklist!.code!.isEmpty) return '';
                    return "${accountbooklist.code!}~${accountbooklist.name1!}";
                  },
                  dropdownDecoratorProps: DropDownDecoratorProps(
                    dropdownSearchDecoration: InputDecoration(
                      border: const OutlineInputBorder(),
                      contentPadding: const EdgeInsets.only(left: 10, top: 0, bottom: 0, right: 10),
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      labelText: global.language("account_book"),
                    ),
                  ),
                  onChanged: (isEditMode)
                      ? (AccountBookModel? value) {
                          setState(() {
                            screenData.bookcode = value!.code;
                            accountBookSeleted = value;
                          });
                        }
                      : null,
                  popupProps: const PopupPropsMultiSelection.dialog(
                    showSearchBox: true,
                    showSelectedItems: true,
                  ),
                  selectedItem: accountBookSeleted,
                ),
              ),
            ),*/
          ],
        ),
      ),
    );

    formWidgets.add(
      const Padding(padding: EdgeInsets.all(10.0), child: Divider()),
    );

    TextStyle boldTextStyle = const TextStyle(
      fontSize: 16,
      fontWeight: FontWeight.bold,
    );
    formWidgets.add(
      Padding(
        padding: EdgeInsets.all(8.0),
        child: Row(
          children: [
            Expanded(
              flex: 2,
              child: Center(
                child: Text(
                  global.language("action_code"),
                  style: boldTextStyle,
                ),
              ),
            ),
            Expanded(
              flex: 3,
              child: Center(
                child: Text(
                  global.language("description"),
                  style: boldTextStyle,
                ),
              ),
            ),
            Expanded(
              flex: 2,
              child: Center(
                child: Text(global.language("debit"), style: boldTextStyle),
              ),
            ),
            Expanded(
              flex: 2,
              child: Center(
                child: Text(global.language("credit"), style: boldTextStyle),
              ),
            ),
            Expanded(
              flex: 1,
              child: Center(
                child: Text(
                  global.language("isentryselfaccount"),
                  style: boldTextStyle,
                ),
              ),
            ),
            Expanded(
              flex: 1,
              child: Center(child: Text('', style: boldTextStyle)),
            ),
          ],
        ),
      ),
    );

    for (int i = 0; i < screenData.details.length; i++) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: Row(
            children: [
              Expanded(
                flex: 2,
                child: Padding(
                  padding: const EdgeInsets.only(right: 10.0),
                  child: TextField(
                    controller: actionCodeController[i],
                    textAlign: TextAlign.left,
                    textCapitalization: TextCapitalization.characters,
                    decoration: const InputDecoration(
                      border: OutlineInputBorder(),
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 3,
                child: Padding(
                  padding: const EdgeInsets.only(right: 10.0),
                  child: TextField(
                    controller: detailController[i],
                    textAlign: TextAlign.left,
                    textCapitalization: TextCapitalization.characters,
                    decoration: const InputDecoration(
                      border: OutlineInputBorder(),
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 2,
                child: Padding(
                  padding: const EdgeInsets.only(right: 10.0),
                  child: TextField(
                    readOnly: true,
                    controller: debitShowController[i],
                    textAlign: TextAlign.left,
                    textCapitalization: TextCapitalization.characters,
                    decoration: InputDecoration(
                      border: const OutlineInputBorder(),
                      suffixIcon: IconButton(
                        icon: const Icon(Icons.search),
                        onPressed: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) =>
                                  const AccountChartSelectScreen(word: ''),
                            ),
                          ).then((value) {
                            setState(() {
                              if (value.accountcode != "") {
                                screenData.details[i].accountdebit = value;
                                debitController[i].text = value.accountcode;
                                debitShowController[i].text =
                                    value.accountcode +
                                    ' ~ ' +
                                    value.accountname;
                              }
                            });
                          });
                        },
                      ),
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 2,
                child: TextField(
                  readOnly: true,
                  controller: creditShowController[i],
                  textAlign: TextAlign.left,
                  textCapitalization: TextCapitalization.characters,
                  decoration: InputDecoration(
                    border: const OutlineInputBorder(),
                    suffixIcon: IconButton(
                      icon: const Icon(Icons.search),
                      onPressed: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (context) =>
                                const AccountChartSelectScreen(word: ''),
                          ),
                        ).then((value) {
                          setState(() {
                            if (value.accountcode != "") {
                              screenData.details[i].accountcredit = value;
                              creditController[i].text = value.accountcode;
                              creditShowController[i].text =
                                  value.accountcode + ' ~ ' + value.accountname;
                            }
                          });
                        });
                      },
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Align(
                  alignment: Alignment.center,
                  child: Checkbox(
                    value: isEntrySelfAccountController[i],
                    onChanged: (bool? value) {
                      setState(() {
                        isEntrySelfAccountController[i] = value!;
                      });
                    },
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: IconButton(
                  onPressed: () {
                    setState(() {
                      screenData.details.removeAt(i);

                      actionCodeController.removeAt(i);
                      detailController.removeAt(i);
                      debitController.removeAt(i);
                      creditController.removeAt(i);
                      debitShowController.removeAt(i);
                      creditShowController.removeAt(i);
                      isEntrySelfAccountController.removeAt(i);
                    });
                  },
                  icon: const Icon(Icons.delete),
                ),
              ),
            ],
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: const EdgeInsets.all(10.0),
        child: SizedBox(
          width: double.infinity,
          height: 40,
          child: ElevatedButton(
            style: ButtonStyle(
              backgroundColor: WidgetStateProperty.all<Color>(
                Colors.green,
              ), // Set the background color
            ),
            onPressed: (isEditMode)
                ? () {
                    setState(() {
                      actionCodeController.add(TextEditingController());
                      detailController.add(TextEditingController());
                      debitController.add(TextEditingController());
                      creditController.add(TextEditingController());
                      debitShowController.add(TextEditingController());
                      creditShowController.add(TextEditingController());
                      isEntrySelfAccountController.add(false);

                      screenData.details.add(
                        DetailModel(
                          detail: '',
                          debit: '',
                          credit: '',
                          isentryselfaccount: false,
                          actioncode: '',
                        ),
                      );
                    });
                  }
                : null,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.add),
                Text(global.language("add_description")),
              ],
            ),
          ),
        ),
      ),
    );

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
        title: Text(headerEdit + global.language("doc_format")),
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
                            context.read<DocFormatBloc>().add(
                              DocFormatDelete(guid: selectGuid),
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
              BlocListener<DocFormatBloc, DocFormatState>(
                listener: (context, state) {
                  if (state is DocFormatLoadDefaultSuccess) {
                    setState(() {
                      /// remove data where doccode == GL
                      state.docFormats.removeWhere(
                        (element) => element.doccode == 'GL',
                      );
                      defaultData = state.docFormats;

                      screenData.module = defaultData[0].doccode;
                      screenData.doccode = defaultData[0].doccode;
                      docFormatCodeController = TextEditingController(
                        text: defaultData[0].doccode,
                      );
                      exampleValue();
                      loadDataList("");
                    });
                  }
                  // Load
                  if (state is DocFormatLoadListSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.docFormat.isNotEmpty) {
                        /// remove data state.docFormats where module == GL
                        state.docFormat.removeWhere(
                          (element) => element.module == 'GL',
                        );
                        listData.addAll(state.docFormat);
                      } else {
                        /// dialog confrim create default data
                        List<DocFormatModel> addDefaultData = [];

                        for (int i = 0; i < defaultData.length; i++) {
                          /// add data where defaultData.doccode != GL
                          if (defaultData[i].doccode != 'GL') {
                            addDefaultData.add(
                              DocFormatModel(
                                module: defaultData[i].doccode,
                                doccode: defaultData[i].doccode,
                                description: '',
                                dateformate: defaultData[i].dateformate,
                                docnumber: 4,
                                details: [],
                                docformat: '',
                                isautoformat: false,
                                yeartype: 0,
                              ),
                            );
                          }
                        }

                        saveDefaultData(context, addDefaultData);
                      }
                    });
                  }
                  if (state is DocFormatLoadListFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is DocFormatSaveSuccess) {
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

                  // Save Bulk
                  if (state is DocFormatSaveBulkSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        global.language("save_success"),
                        Colors.blue,
                      );

                      loadDataList(searchText);
                    });
                  }
                  if (state is DocFormatSaveFailed) {
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
                  if (state is DocFormatUpdateSuccess) {
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
                  if (state is DocFormatUpdateFailed) {
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
                  if (state is DocFormatDeleteSuccess) {
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
                  if (state is DocFormatDeleteManySuccess) {
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
                  if (state is DocFormatGetSuccess) {
                    setState(() {
                      isDataChange = false;

                      screenData = state.docFormat;
                      loadDataToScreen();

                      if (isEditMode) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                      }
                    });
                  }
                },
              ),
              BlocListener<ChartAccountBloc, ChartAccountState>(
                listener: (context, state) {
                  if (state is GroupAccountLoadSuccess) {
                    setState(() {
                      accountGroupList.addAll(state.groupAccounts);
                    });
                  }
                  if (state is BookAccountLoadSuccess) {
                    setState(() {
                      accountBookList.addAll(state.bookAccounts);
                    });
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

  Future<String?> saveDefaultData(
    BuildContext context,
    List<DocFormatModel> addDefaultData,
  ) {
    return showDialog<String>(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: Text(global.language('create_default_data')),
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
              context.read<DocFormatBloc>().add(
                DocFormatBulkSave(docFormatModel: addDefaultData),
              );
            },
            child: Text(global.language('yes')),
          ),
        ],
      ),
    );
  }
}
