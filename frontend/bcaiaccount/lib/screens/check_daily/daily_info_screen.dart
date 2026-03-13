import 'dart:convert';
import 'dart:io';
import 'package:smlaicloud/bloc/company/company_bloc.dart';
import 'package:smlaicloud/bloc/trans/trans_bloc.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:json_view/json_view.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';

class DailyInfoScreen extends StatefulWidget {
  const DailyInfoScreen({super.key});

  @override
  State<DailyInfoScreen> createState() => DailyInfoScreenState();
}

class DailyInfoScreenState extends State<DailyInfoScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<TransactionModel> listData = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isChange = false;
  late TransState blocTransState;
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  late TransactionModel screenData;
  String jsonString = "";
  Map<String, dynamic> jsonData = {};

  void changeScreenEvent(global.ScreenEventEnum event) {
    screenEvent = event;

    setState(() {});
  }

  @override
  void initState() {
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    super.initState();
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );

    listScrollController.addListener(onScrollList);
    loadDataList("");

    screenData = TransactionModel(
      shopid: global.apiShopCode,
      guidref: '',
      docno: '',
      docdatetime: '',
      docrefno: '',
      docrefdate: '',
      docreftype: 0,
      doctype: 0,
      vattype: 0,
      custcode: '',
      custnames: [],
      salecode: '',
      salename: '',
      discountword: '',
      totalcost: 0,
      totalvalue: 0,
      totaldiscount: 0,
      totalvatvalue: 0,
      totalaftervat: 0,
      totalexceptvat: 0,
      totalamount: 0,
      cashiercode: '',
      posid: '',
      membercode: '',
      vatrate: 0,
      status: 0,
      inquirytype: 0,
      taxdocdate: '',
      taxdocno: '',
      totalbeforevat: 0,
      transflag: 0,
      iscancel: false,
      ismanualamount: false,
      description: '',
      details: <TransactionDetailModel>[],
      paymentdetailraw: "",
      paymentStructs: [],
      ispos: false,
      isbom: false,
      getpoint: 0,
      usepoint: 0,
      pointdiscountamount: 0,
      paypointamount: 0,
      pointscode: '',
      isclosed: false,
      islocked: false,
      isref: false,
      detailcount: 0,
    );

    super.initState();
  }

  void loadDataList(String search) {
    searchText = search;
    context.read<TransBloc>().add(
      TransLoad(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
        type: global.TransactionTypeEnum.sale,
        custcode: "",
        ispos: "null",
        dateorder: 0,
      ),
    );
    context.read<CompanyBloc>().add(const CompanyLoad());
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

    super.dispose();
  }

  void clearEditData() {
    isChange = false;
  }

  void getData(String guid) {
    screenData = listData.firstWhere((element) => element.guidfixed == guid);
    jsonString = jsonEncode(screenData.toJson());
    // Assuming jsonString contains your JSON data
    jsonData = json.decode(jsonString);
    changeScreenEvent(global.ScreenEventEnum.list);
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('daily_info')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(context);
            changeScreenEvent(global.ScreenEventEnum.list);
          },
        ),
      ),
      body: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(5),
            decoration: BoxDecoration(
              color: global.theme.surfaceColor,
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
                    style: TextStyle(color: global.theme.textColor),
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.symmetric(vertical: 8, horizontal: 4),
                      border: InputBorder.none,
                      filled: false,
                      hintText: global.language('search'),
                      hintStyle: TextStyle(color: global.theme.formHintColor),
                          prefixIcon: Icon(Icons.search, size: 20, color: global.theme.formHintColor),
                          prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                    ),
                  ),
                ),
                ListFontSizeControl(onChanged: () => setState(() {})),
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
                    global.language("docdate"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Text(
                    global.language("docno"),
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
                    global.language("amount"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
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
                              (element) => element.guidfixed! == selectGuid,
                            ),
                          );
                          if (index > 0) {
                            selectGuid = listData[index - 1].guidfixed!;
                            currentListIndex = index - 1;
                            isKeyUp = true;
                            getData(selectGuid);
                          }
                        }
                        if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                          isKeyUp = false;
                          int index = listData.indexOf(
                            listData.firstWhere(
                              (element) => element.guidfixed! == selectGuid,
                            ),
                          );
                          selectGuid = listData[index + 1].guidfixed!;
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
                    .map((value) => listObject(listData.indexOf(value), value))
                    .toList(),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget listObject(int index, TransactionModel value) {
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    bool selected = selectGuid == value.guidfixed!;
    TextStyle textStyle = TextStyle(
      fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
      fontSize: (selected)
          ? global.deviceConfig.listDataFontSize + 2.0
          : global.deviceConfig.listDataFontSize,
    );
    return MouseRegion(

      cursor: SystemMouseCursors.click,

      child: GestureDetector(
      onTap: () {
        setState(() {
          changeScreenEvent(global.ScreenEventEnum.list);
          selectGuid = value.guidfixed!;
          getData(selectGuid);
          WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
            tabController.animateTo(1);
          });
        });
      },
      child: Container(
        key: index < listKeys.length ? listKeys[index] : null,
        decoration: BoxDecoration(
          color: (selectGuid == value.guidfixed!)
              ? global.theme.rowSelectedColor
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
              flex: 1,
              child: Text(
                value.docdatetime,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 2,
              child: Text(
                value.docno,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 1,
              child: Text(
                value.totalamount.toString(),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
          ],
        ),
      ),
    ),
    );
  }

  Widget editScreen({mobileScreen}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        leading: mobileScreen
            ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: const Icon(Icons.arrow_back),
                onPressed: () async {
                  changeScreenEvent(global.ScreenEventEnum.list);
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                },
              )
            : null,
        title: const Text("DATA JSON"),
      ),
      body: Scaffold(
        backgroundColor:
            global.theme.backgroundColor, // ใช้สีพื้นหลังจาก theme
        body: LayoutBuilder(
          builder: (BuildContext context, BoxConstraints constraints) {
            return
            // DetailPage(
            //   screenData: screenData,
            //   companydata: companydata,
            // );
            // DetailPage(screenData: screenData);
            SingleChildScrollView(
              controller: editScrollController,
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: (screenData.guidfixed!.isNotEmpty)
                    ? SizedBox(
                        // Wrap the Column in a Container
                        height: constraints
                            .maxHeight, // Use maxHeight to limit the height
                        child: Column(
                          children: [
                            Expanded(
                              child: JsonConfig(
                                data: JsonConfigData(
                                  animation: true,
                                  animationDuration: const Duration(
                                    milliseconds: 300,
                                  ),
                                  animationCurve: Curves.ease,
                                  itemPadding: const EdgeInsets.only(left: 8),
                                  color: JsonColorScheme(
                                    stringColor: global.theme.textSecondaryColor,
                                  ),
                                  style: const JsonStyleScheme(
                                    arrow: Icon(Icons.arrow_right),
                                  ),
                                ),
                                child: JsonView(
                                  json: jsonData, // Specify your JSON data here
                                ),
                              ),
                            ),
                          ],
                        ),
                      )
                    : const Text(""),
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

    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<TransBloc, TransState>(
            listener: (context, state) {
              blocTransState = state;
              // Load
              if (state is TransLoadSuccess) {
                setState(() {
                  if (state.trans.isNotEmpty) {
                    listData.addAll(state.trans);
                  }
                });
              }
            },
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
