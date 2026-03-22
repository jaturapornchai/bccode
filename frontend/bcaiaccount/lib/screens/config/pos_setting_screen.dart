import 'dart:async';
import 'dart:io';

// import 'package:smlaicloud/flavors.dart'; // Not used
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/book_bank_model.dart';
import 'package:smlaicloud/model/pos_media_model.dart';
import 'package:smlaicloud/model/pos_setting_model.dart';
import 'package:smlaicloud/model/qr_model.dart';
import 'package:smlaicloud/model/sale_channel_model.dart';
import 'package:smlaicloud/model/timezones_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/report_repository.dart';
import 'package:smlaicloud/screen_search/bookbank_select_screen.dart';
import 'package:smlaicloud/screen_search/company_branch_search_screen.dart';
import 'package:smlaicloud/screen_search/employee_search_screen.dart';
import 'package:smlaicloud/screen_search/pos_media_search_screen.dart';
import 'package:smlaicloud/screen_search/product_location_search_screen.dart';
import 'package:smlaicloud/screen_search/product_warehouse_search_screen.dart';
import 'package:smlaicloud/screen_search/qr_search_screen.dart';
import 'package:smlaicloud/screen_search/sale_chanels_search_screen.dart';
// import 'package:dropdown_search/dropdown_search.dart'; // Not used
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/utils/focus_utils.dart';
import 'package:image_picker/image_picker.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class PosSettingScreen extends StatefulWidget {
  const PosSettingScreen({super.key});

  @override
  State<PosSettingScreen> createState() => PosSettingScreenState();
}

class PosSettingScreenState extends State<PosSettingScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  List<LanguageModel> languageList = <LanguageModel>[];
  List<PosSettingModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late PosSettingState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late PosSettingModel screenData;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  bool loadingData = false;
  late DateTime dateNow = DateTime.now();
  final _formKey = GlobalKey<FormState>();
  File imageFile = File('');
  Uint8List? imageWeb;
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;

  TextEditingController codeController = TextEditingController();
  TextEditingController devicenumberController = TextEditingController();
  TextEditingController doccodeController = TextEditingController();
  TextEditingController docformatinv = TextEditingController();
  TextEditingController docformattaxinv = TextEditingController();
  TextEditingController docformatesalereturn = TextEditingController();
  TextEditingController vatRateController = TextEditingController();
  TextEditingController servicechargeController = TextEditingController();

  String pinActive = "";
  TextEditingController field1 = TextEditingController();
  TextEditingController field2 = TextEditingController();
  TextEditingController field3 = TextEditingController();
  TextEditingController field4 = TextEditingController();
  TextEditingController field5 = TextEditingController();
  TextEditingController field6 = TextEditingController();
  TextEditingController field7 = TextEditingController();
  TextEditingController field8 = TextEditingController();

  List<String> permistionList = [];
  bool isActivePin = false;
  bool validationBranch = false;
  bool validationWarehouse = false;

  List<SlipListModel> slipSeleted = [];

  List<TextEditingController> bookBankQrSerchController = [];
  List<TextEditingController> bookBankCreditcardSerchController = [];
  List<TextEditingController> bookBankTransfercardSerchController = [];

  late PosMediaModel posMediaData;

  late TimezonesModel timezoneSelected;

  final List<TimeOfDay> _selectedTime = [];
  List<TextEditingController> mediaFromTimeController = [];
  List<TextEditingController> mediaToTimeController = [];

  String activeAPIKEY = "";
  bool isLoadingActive = false;

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

    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    permistionList = [global.language("sell"), global.language("return_goods"), global.language("cancel_bill"), global.language("discount")];

    clearEditData();
    setSystemLanguageList();

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();

    codeController.dispose();
    devicenumberController.dispose();
    doccodeController.dispose();
    docformatinv.dispose();
    docformattaxinv.dispose();
    docformatesalereturn.dispose();
    vatRateController.dispose();
    servicechargeController.dispose();

    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<PosSettingBloc>().add(
      PosSettingLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
      ),
    );
  }

  void loadDataWarehouse0000() {
    context.read<WarehouseBloc>().add(const WarehouseGetByCode(code: "00000"));
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText);
    }
  }

  void clearEditData() {
    listScrollController.addListener(onScrollList);

    screenData = PosSettingModel(
      branch: BranchModel(
        code: global.companyBranchSelectData.code,
        guidfixed: global.companyBranchSelectData.guidfixed,
        names: global.companyBranchSelectData.names,
      ),
      code: '',
      devicenumber: '',
      docformatinv: 'YYMMDD####',
      docformattaxinv: 'IVYYMMDD####',
      docformatesalereturn: 'CNYYMMDD####',
      location: LocationModel(code: '', names: []),
      receiptform: '1',
      warehouse: WarehouseModel(code: '', names: [], guidfixed: ''),
      guidfixed: '',
      activepin: '',
      employees: [],
      salechannels: [],
      doccode: '',
      vattype: 0,
      vatrate: 7.0,
      servicecharge: 0.0,
      isejournal: false,
      isusecreadit: false,
      billheader: [],
      billfooter: [],
      isvatregister: true,
      slips: [
        SlipModel(
          code: "S-01",
          name: "ใบสรุปรายการ",
          isrequire: false,
          formcode: global.summaryOrderBill.code,
          formnames: global.summaryOrderBill.names,
          headernames: [LanguageDataModel(code: "th", name: "ใบสรุป")],
        ),
        SlipModel(
          code: "S-02",
          name: global.language("pdf_short_tax_invoice"),
          isrequire: false,
          formcode: global.taxInvoice.code,
          formnames: global.taxInvoice.names,
          headernames: [
            LanguageDataModel(
              code: "th",
              name: "ใบเสร็จรับเงิน/ใบกำกับภาษีอย่างย่อ",
            ),
          ],
        ),
        SlipModel(
          code: "S-03",
          name: "ใบกำกับภาษีอย่างเต็ม",
          isrequire: false,
          formcode: global.taxInvoiceFull.code,
          formnames: global.taxInvoiceFull.names,
          headernames: [
            LanguageDataModel(code: "th", name: "ใบเสร็จรับเงิน/ใบกำกับภาษี"),
          ],
        ),
        SlipModel(
          code: "S-04",
          name: global.language("pdf_receipt"),
          isrequire: false,
          formcode: global.slipReceipt.code,
          formnames: global.slipReceipt.names,
          headernames: [LanguageDataModel(code: "th", name: "ใบเสร็จรับเงิน")],
        ),
        SlipModel(
          code: "R-01",
          name: "ใบรับคืนอย่างย่อ",
          isrequire: false,
          formcode: global.slipReturn.code,
          formnames: global.slipReturn.names,
          headernames: [LanguageDataModel(code: "th", name: "ใบลดหนี้")],
        ),
        SlipModel(
          code: "R-02",
          name: "ใบรับคืนอย่างเต็ม",
          isrequire: false,
          formcode: global.slipReturnFull.code,
          formnames: global.slipReturnFull.names,
          headernames: [
            LanguageDataModel(code: "th", name: "ใบลดหนี้/ใบกำกับภาษี"),
          ],
        ),
      ],
      mediaguid: '',
      qrcodes: [],
      creditcards: [],
      transfers: [],
      timezoneoffset: '',
      timezonelabel: '',
      timeforsales: [],
      categorygroupnumber: 1,
      kitchengroupnumber: 1,
      tablegroupnumber: 1,
      zonegroupnumber: 1,
    );

    bookBankQrSerchController = [];
    bookBankCreditcardSerchController = [];
    bookBankTransfercardSerchController = [];

    posMediaData = PosMediaModel(
      guidfixed: '',
      code: '',
      description: [],
      resources: [],
    );

    slipSeleted = [
      SlipListModel(
        code: screenData.slips![0].formcode,
        names: screenData.slips![0].formnames,
        headernames: [],
      ),
      SlipListModel(
        code: screenData.slips![1].formcode,
        names: screenData.slips![1].formnames,
        headernames: [],
      ),
      SlipListModel(
        code: screenData.slips![2].formcode,
        names: screenData.slips![2].formnames,
        headernames: [],
      ),
      SlipListModel(
        code: screenData.slips![3].formcode,
        names: screenData.slips![3].formnames,
        headernames: [],
      ),
      SlipListModel(
        code: screenData.slips![4].formcode,
        names: screenData.slips![4].formnames,
        headernames: [],
      ),
      SlipListModel(
        code: screenData.slips![5].formcode,
        names: screenData.slips![5].formnames,
        headernames: [],
      ),
    ];

    timezoneSelected = TimezonesModel(
      abbr: '',
      isDst: false,
      offset: '',
      text: '',
      utc: [],
      value: '',
    );

    mediaFromTimeController = [];
    mediaToTimeController = [];

    isActivePin = false;
    isDataChange = false;
    validationBranch = false;
    validationWarehouse = false;
    setState(() {
      loadDataWarehouse0000();
      loadDataToScreen();
      imageFile = File('');
      imageWeb = null;
    });
  }

  Future<void> getNativeTimezone() async {
    DateTime now = DateTime.now();
    int timeZoneOffset = now.timeZoneOffset.inHours;

    String currentTimeZone = "";
    if (kIsWeb == false) {
      // เอา timezone จาก DateTime.now()
      currentTimeZone = DateTime.now().timeZoneName;
    } else {
      currentTimeZone = "Asia/Bangkok";
    }

    // print(timeZoneOffset.abs());
    // print(currentTimeZone);

    // Find the first timezone that matches the given conditions
    TimezonesModel timezone = global.timezonesListData.firstWhere(
      (element) => element.offset == timeZoneOffset.abs().toString(),
    );
    for (var element in timezone.utc) {
      if (element == currentTimeZone) {
        timezoneSelected = timezone;
        break;
      }
    }
  }

  void loadDataToScreen() {
    codeController.text = screenData.code;
    devicenumberController.text = screenData.devicenumber;
    docformatinv.text = screenData.docformatinv;
    docformattaxinv.text = screenData.docformattaxinv;
    docformatesalereturn.text = screenData.docformatesalereturn!;
    doccodeController.text = screenData.doccode!;
    vatRateController.text = screenData.vatrate.toString();
    servicechargeController.text = screenData.servicecharge.toString();

    if (screenData.timezonelabel!.isNotEmpty) {
      TimezonesModel timezonesModel = global.timezonesListData.firstWhere(
        (element) => element.text == screenData.timezonelabel,
      );
      timezoneSelected = timezonesModel;
    } else {
      getNativeTimezone();
    }

    if (screenData.qrcodes!.isNotEmpty) {
      for (int i = 0; i < screenData.qrcodes!.length; i++) {
        bookBankQrSerchController.add(TextEditingController());
        bookBankQrSerchController[i].text =
            "${screenData.qrcodes![i].bankcode} ~ ${screenData.qrcodes![i].bookbankcode!}";
      }
    }

    if (screenData.creditcards!.isNotEmpty) {
      for (int i = 0; i < screenData.creditcards!.length; i++) {
        bookBankCreditcardSerchController.add(TextEditingController());
        bookBankCreditcardSerchController[i].text =
            "${screenData.creditcards![i].bookbank!.bankcode} ~ ${screenData.creditcards![i].bookbank!.passbook}";
      }
    }

    if (screenData.transfers!.isNotEmpty) {
      for (int i = 0; i < screenData.transfers!.length; i++) {
        bookBankTransfercardSerchController.add(TextEditingController());
        bookBankTransfercardSerchController[i].text =
            "${screenData.transfers![i].bookbank!.bankcode} ~ ${screenData.transfers![i].bookbank!.passbook}";
      }
    }

    if (screenData.slips!.isNotEmpty) {
      slipSeleted = [];
      for (int i = 0; i < screenData.slips!.length; i++) {
        slipSeleted.add(
          SlipListModel(
            code: screenData.slips![i].formcode,
            names: screenData.slips![i].formnames,
            headernames: screenData.slips![i].headernames,
          ),
        );
      }
    }

    if (screenData.timeforsales.isNotEmpty) {
      for (int i = 0; i < screenData.timeforsales.length; i++) {
        _selectedTime.add(TimeOfDay.now());
        mediaFromTimeController.add(TextEditingController());
        mediaToTimeController.add(TextEditingController());

        mediaFromTimeController[i].text = screenData.timeforsales[i].from;
        mediaToTimeController[i].text = screenData.timeforsales[i].to;
      }
    }

    if (screenData.categorygroupnumber == 0 ||
        screenData.categorygroupnumber == null) {
      screenData.categorygroupnumber = 1;
    }

    if (screenData.kitchengroupnumber == 0 ||
        screenData.kitchengroupnumber == null) {
      screenData.kitchengroupnumber = 1;
    }

    if (screenData.tablegroupnumber == 0 ||
        screenData.tablegroupnumber == null) {
      screenData.tablegroupnumber = 1;
    }

    if (screenData.zonegroupnumber == 0 || screenData.zonegroupnumber == null) {
      screenData.zonegroupnumber = 1;
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
    context.read<PosSettingBloc>().add(PosSettingGet(guid: guid));
  }

  void switchToEdit(PosSettingModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      focusFirstTextField(context);
    });
  }

  Future<void> activatePin() async {
    /// 1 = dev , 0 = prod , 2 = uat
    int isDeveloper = global.isdevPin;

    ReportRepository reportRepository = ReportRepository();
    try {
      ApiResponse result = await reportRepository.activePos(
        pinActive,
        screenData.code,
        isDeveloper,
        activeAPIKEY,
      );
      if (result.success) {
        isActivePin = true;
        if (isActivePin) {
          updateData(selectGuid);
          setState(() {
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              global.language("active_pin_success"),
              Colors.blue,
            );
            Navigator.pop(context);
            field1.text = '';
            field2.text = '';
            field3.text = '';
            field4.text = '';
            field5.text = '';
            field6.text = '';
            field7.text = '';
            field8.text = '';
            isLoadingActive = false;
            activeAPIKEY = "";
          });
        }
      }
    } catch (e) {
      isActivePin = false;
      if (mounted) {
        global.showSnackBar(
          context,
          const Icon(Icons.error, color: Colors.white),
          e.toString(),
          Colors.red,
        );
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
        title: Text(global.language("pos_setting")),
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
                            backgroundColor: global.theme.primaryColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<PosSettingBloc>().add(
                              PosSettingDeleteMany(guid: guidListChecked),
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
              padding: const EdgeInsets.all(5),
              decoration: BoxDecoration(
                color: global.theme.searchBarColor,
                borderRadius: BorderRadius.circular(2),
              ),
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
                        contentPadding: EdgeInsets.only(
                          top: 0,
                          bottom: 0,
                          left: 0,
                          right: 0,
                        ),
                        border: InputBorder.none,
                        prefixIcon: Icon(Icons.search, size: 20, color: global.theme.iconColor),
                        prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                        hintText: global.language('search'),
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
                      global.language("pos_setting_code"),
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
                      global.language("pos_setting_devicenumber"),
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
                    child: Align(
                      alignment: Alignment.center,
                      child: Text(
                        global.language("active_pin"),
                        style: TextStyle(
                          color: global.theme.columnHeaderTextColor,
                          fontWeight: FontWeight.bold,
                          fontSize: global.deviceConfig.listDataFontSize + 2,
                        ),
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
                  color: global.theme.primaryColor,
                  size: 50,
                ),
              ),
          ],
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

  Widget listObject(int index, PosSettingModel value, bool showCheckBox) {
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
                value.code,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                value.devicenumber,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Padding(
                padding: EdgeInsets.all(8.0),
                child: ElevatedButton.icon(
                  onPressed: () {
                    getData(value.guidfixed);
                    if ((value.activepin!.isEmpty)) {
                      // เช็คค่าใน Local storage ก่อน
                      String? user = global.appConfig.getString("user");
                      if (user == null || user.isEmpty) {
                        // ถ้าไม่มี user ใน storage ให้ logout และไปหน้า login ใหม่
                        context.read<LoginBloc>().add(const Logout());
                        return;
                      }

                      widgetActivePin(value);
                    } else {
                      /// show dialog confirm cancel pin
                      showDialog<String>(
                        context: context,
                        builder: (BuildContext context) => AlertDialog(
                          title: Text(global.language('confirm_cancel_pin')),
                          content: Text(global.language('how_to_cancel_pin')),
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
                              onPressed: () async {
                                // screenData.activepin = "";
                                // updateData(value.guidfixed);
                                ReportRepository reportRepository =
                                    ReportRepository();
                                try {
                                  /// get apikey in clickhouse
                                  ApiResponse result = await reportRepository
                                      .getApiKey(value.activepin!);
                                  if (result.success) {
                                    List<dynamic> data = result.data;

                                    /// เก็บ apikey ไว้ใช้ในการลบ ใน apikeyservice
                                    activeAPIKEY = data[0]["apikey"];

                                    try {
                                      /// delete pin in clickhouse pinlist
                                      ApiResponse result =
                                          await reportRepository.deletePos(
                                            value.activepin!,
                                          );
                                      if (result.success) {
                                        /// call bloc DeleteApikey
                                        if (mounted) {
                                          context.read<PosSettingBloc>().add(
                                            DeleteApikey(apikey: activeAPIKEY),
                                          );
                                        }
                                      }
                                    } catch (e) {
                                      if (mounted) {
                                        global.showSnackBar(
                                          context,
                                          const Icon(
                                            Icons.error,
                                            color: Colors.white,
                                          ),
                                          e.toString(),
                                          Colors.red,
                                        );
                                      }
                                    }
                                  }
                                } catch (e) {
                                  context.read<PosSettingBloc>().add(
                                    DeleteApikey(apikey: activeAPIKEY),
                                  );

                                  // if (mounted) {
                                  //   global.showSnackBar(
                                  //     context,
                                  //     const Icon(Icons.error, color: Colors.white),
                                  //     e.toString(),
                                  //     Colors.red,
                                  //   );
                                  // }
                                }
                              },
                              child: Text(global.language('yes')),
                            ),
                          ],
                        ),
                      );
                    }
                  },
                  icon: (value.activepin!.isEmpty)
                      ? const Icon(Icons.close)
                      : Icon(Icons.check),
                  label: Text(
                    global.language(
                      (value.activepin!.isEmpty) ? "not_active_pin" : "cancel",
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    foregroundColor: Colors.white,
                    backgroundColor: (value.activepin!.isEmpty)
                        ? Colors.red
                        : Colors.green,
                  ),
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

  Future<String?> widgetActivePin(PosSettingModel value) {
    bool isPinEmty = false;

    return showDialog<String?>(
      context: context,
      barrierDismissible:
          false, // Allows tapping outside the dialog to close it
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setState) {
            return AlertDialog(
              scrollable: true,
              title: Text(' ${global.language("verification_pin")}'),
              content: Column(
                children: [
                  Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                        children: [
                          OtpInput(field1, true), // auto focus
                          OtpInput(field2, false),
                          OtpInput(field3, false),
                          OtpInput(field4, false),
                          OtpInput(field5, false),
                          OtpInput(field6, false),
                          OtpInput(field7, false),
                          OtpInput(field8, false),
                        ],
                      ),
                    ],
                  ),
                  (isPinEmty)
                      ? Text(
                          global.language("please_enter_pin"),
                          style: const TextStyle(
                            color: Colors.red,
                            fontSize: 12,
                          ),
                        )
                      : Container(),
                ],
              ),
              actions: <Widget>[
                Padding(
                  padding: EdgeInsets.all(8.0),
                  child: ElevatedButton.icon(
                    icon: isLoadingActive
                        ? Container(
                            width: 24,
                            height: 24,
                            padding: const EdgeInsets.all(2.0),
                            child: const CircularProgressIndicator(
                              color: Colors.white,
                              strokeWidth: 3,
                            ),
                          )
                        : Icon(Icons.pin),
                    label: Text(global.language("update")),
                    onPressed:
                        isLoadingActive // Disable the button when loading
                        ? null
                        : () async {
                            if (!validateOtpFields()) {
                              setState(() {
                                isPinEmty = true;
                              });

                              return;
                            }

                            selectGuid = value.guidfixed;

                            pinActive =
                                field1.text +
                                field2.text +
                                field3.text +
                                field4.text +
                                field5.text +
                                field6.text +
                                field7.text +
                                field8.text;
                            screenData.activepin = pinActive;

                            setState(() {
                              isPinEmty = false;
                              isLoadingActive = true;
                            });

                            /// delay 1 second next get APIKEY
                            await Future.delayed(const Duration(seconds: 1));

                            /// get APIKEY
                            if (mounted) {
                              context.read<PosSettingBloc>().add(
                                const GetApiKey(),
                              );
                            }
                          },
                  ),
                ),
                Padding(
                  padding: EdgeInsets.all(8.0),
                  child: TextButton(
                    child: Text(global.language("cancel")),
                    onPressed: () {
                      field1.text = '';
                      field2.text = '';
                      field3.text = '';
                      field4.text = '';
                      field5.text = '';
                      field6.text = '';
                      field7.text = '';
                      field8.text = '';
                      isLoadingActive = false;
                      activeAPIKEY = "";
                      isPinEmty = false;

                      Navigator.pop(context);
                    },
                  ),
                ),
              ],
            );
          },
        );
      },
    );
  }

  bool validateOtpFields() {
    List<TextEditingController> controllers = [
      field1,
      field2,
      field3,
      field4,
      field5,
      field6,
      field7,
      field8,
    ];

    // Example validation: ensure all fields are not empty
    for (TextEditingController controller in controllers) {
      if (controller.text.isEmpty || controller.text.length != 1) {
        // You can adjust the condition based on your validation rules
        return false; // Indicates validation failed
      }
    }

    // All fields pass the validation
    return true;
  }

  void saveOrUpdateData() {
    validationBranch = false;
    validationWarehouse = false;
    if (verifyData(screenData)) {
      showCheckBox = false;
      // print(screenData.toJson());
      if (selectGuid.trim().isEmpty) {
        if (imageFile.path.isNotEmpty) {
          context.read<PosSettingBloc>().add(
            PosSettingWithImageSave(
              posSetting: screenData,
              imageFile: imageFile,
              imageWeb: imageWeb,
            ),
          );
        } else {
          context.read<PosSettingBloc>().add(
            PosSettingSave(posSetting: screenData),
          );
        }
      } else {
        updateData(selectGuid);
      }
    }
  }

  bool verifyData(PosSettingModel value) {
    List<String> errorList = [];
    if (value.branch.guidfixed.isEmpty) {
      validationBranch = true;
      errorList.add(global.language("please_select_company_branch"));
      setState(() {});
    }
    if (value.warehouse.guidfixed.isEmpty) {
      validationWarehouse = true;
      errorList.add(global.language("please_select_warehouse"));
      setState(() {});
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

  void updateData(String guid) {
    showCheckBox = false;
    if (imageWeb != null) {
      context.read<PosSettingBloc>().add(
        PosSettingWithImageUpdate(
          guid: guid,
          posSetting: screenData,
          imageFile: imageFile,
          imageWeb: imageWeb!,
        ),
      );
    } else {
      context.read<PosSettingBloc>().add(
        PosSettingUpdate(guid: guid, posSetting: screenData),
      );
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

  Future<List<SlipListModel>> getDataSlip(filter) async {
    return global.slipPosSaleList;
  }

  void _selecttimeforsalesFrom(BuildContext context, int index) async {
    final TimeOfDay? pickedTime = await showTimePicker(
      context: context,
      initialTime: _selectedTime[index],
      builder: (BuildContext context, Widget? child) {
        return MediaQuery(
          data: MediaQuery.of(context).copyWith(alwaysUse24HourFormat: true),
          child: child!,
        );
      },
    );

    // ignore: unrelated_type_equality_checks
    if (pickedTime != null) {
      setState(() {
        // Format the hour and minute to ensure they have two digits (e.g., 09, 12, 23, etc.)
        String formattedHour = pickedTime.hour.toString().padLeft(2, '0');
        String formattedMinute = pickedTime.minute.toString().padLeft(2, '0');

        // Combine the formatted hour and minute with a ":" separator
        String formattedTime = '$formattedHour:$formattedMinute';

        screenData.timeforsales[index].from = formattedTime;
        mediaFromTimeController[index].text = formattedTime;
      });
    }
  }

  void _selectMediaToTime(BuildContext context, int index) async {
    final TimeOfDay? pickedTime = await showTimePicker(
      context: context,
      initialTime: _selectedTime[index],
      builder: (BuildContext context, Widget? child) {
        return MediaQuery(
          data: MediaQuery.of(context).copyWith(alwaysUse24HourFormat: true),
          child: child!,
        );
      },
    );

    // ignore: unrelated_type_equality_checks
    if (pickedTime != null) {
      setState(() {
        // Format the hour and minute to ensure they have two digits (e.g., 09, 12, 23, etc.)
        String formattedHour = pickedTime.hour.toString().padLeft(2, '0');
        String formattedMinute = pickedTime.minute.toString().padLeft(2, '0');

        // Combine the formatted hour and minute with a ":" separator
        String formattedTime = '$formattedHour:$formattedMinute';

        // Convert the time string to a TimeOfDay object
        TimeOfDay fromTime = global.getTimeOfDayFromString(
          screenData.timeforsales[index].from,
        );
        TimeOfDay toTime = global.getTimeOfDayFromString(formattedTime);

        // Check if toTime is not earlier than fromTime
        if (toTime.hour < fromTime.hour ||
            (toTime.hour == fromTime.hour &&
                toTime.minute <= fromTime.minute)) {
          global.showSnackBar(
            context,
            Icon(Icons.edit, color: Colors.white),
            global.language("to_time_must_be_later_than_from_time"),
            Colors.red,
          );
          screenData.timeforsales[index].to = '';
          mediaToTimeController[index].text = '';
          return;
        } else {
          // Update the toTime and the text in the text field
          screenData.timeforsales[index].to = formattedTime;
          mediaToTimeController[index].text = formattedTime;
        }
      });
    }
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          enabled: isEditMode,
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
            labelText: global.language("pos_setting_code"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return 'This field is required';
            }
            return null;
          },
          onEditingComplete: () {},
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          enabled: isEditMode,
          textAlign: TextAlign.left,
          controller: devicenumberController,
          textCapitalization: TextCapitalization.characters,
          // inputFormatters: <TextInputFormatter>[FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]'))],
          onChanged: (value) {
            isDataChange = true;
            screenData.devicenumber = value;
            devicenumberController.value = TextEditingValue(
              text: value,
              selection: devicenumberController.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("pos_setting_devicenumber"),
          ),
          onEditingComplete: () {},
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: true,
              groupValue: screenData.isvatregister,
              onChanged: (isEditMode)
                  ? (value) {
                      setState(() {
                        screenData.isvatregister = value as bool?;
                      });
                    }
                  : null,
            ),
            Text(global.language("vat_register")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: false,
              groupValue: screenData.isvatregister,
              onChanged: (isEditMode)
                  ? (value) {
                      setState(() {
                        screenData.isvatregister = value as bool?;
                      });
                    }
                  : null,
            ),
            Text(global.language("vat_no_register")),
          ],
        ),
      ),
    );
    if (screenData.isvatregister == true) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 0,
                groupValue: screenData.vattype,
                onChanged: (isEditMode)
                    ? (value) {
                        setState(() {
                          screenData.vattype = value as int?;
                        });
                      }
                    : null,
              ),
              Text(global.language("vat_include")),
              const SizedBox(width: 10),
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 1,
                groupValue: screenData.vattype,
                onChanged: (isEditMode)
                    ? (value) {
                        setState(() {
                          screenData.vattype = value as int?;
                        });
                      }
                    : null,
              ),
              Text(global.language("vat_exclude")),
            ],
          ),
        ),
      );

      /// vatrate
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextField(
            readOnly: !isEditMode,
            enabled: isEditMode,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            inputFormatters: [global.NumberInputFormatter()],
            controller: vatRateController,
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("var_rate"),
            ),
            onChanged: (isEditMode)
                ? (value) {
                    setState(() {
                      screenData.vatrate = double.parse(value);
                    });
                  }
                : null,
          ),
        ),
      );

      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Checkbox(
                value: screenData.isejournal,
                onChanged: (isEditMode)
                    ? (bool? value) {
                        setState(() {
                          screenData.isejournal = value ?? false;
                        });
                      }
                    : null,
              ),
              Text(global.language("ejournal")),
            ],
          ),
        ),
      );
    }
    if (global.posVersion == global.PosVersionEnum.restaurant) {
      /// bessiness type
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 0,
                groupValue: screenData.businesstype,
                onChanged: (isEditMode)
                    ? (value) {
                        setState(() {
                          screenData.businesstype = value as int?;
                        });
                      }
                    : null,
              ),
              Text(global.language("retail_store")),
              const SizedBox(width: 10),
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 1,
                groupValue: screenData.businesstype,
                onChanged: (isEditMode)
                    ? (value) {
                        setState(() {
                          screenData.businesstype = value as int?;
                        });
                      }
                    : null,
              ),
              Text(global.language("restaurant")),
            ],
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextField(
          readOnly: !isEditMode,
          enabled: isEditMode,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          inputFormatters: [global.NumberInputFormatter()],
          controller: servicechargeController,
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("servicecharge"),
          ),
          onChanged: (isEditMode)
              ? (value) {
                  setState(() {
                    screenData.servicecharge = double.parse(value);
                  });
                }
              : null,
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Checkbox(
              value: screenData.isusecreadit,
              onChanged: (isEditMode)
                  ? (bool? value) {
                      setState(() {
                        screenData.isusecreadit = value ?? false;
                      });
                    }
                  : null,
            ),
            Text(global.language("isusecreadit")),
          ],
        ),
      ),
    );
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          enabled: isEditMode,
          textAlign: TextAlign.left,
          controller: doccodeController,
          textCapitalization: TextCapitalization.characters,
          // keyboardType: TextInputType.number,
          // inputFormatters: [FilteringTextInputFormatter.digitsOnly],
          maxLength: 3,
          onChanged: (value) {
            isDataChange = true;
            screenData.doccode = value;
            doccodeController.value = TextEditingValue(
              text: value,
              selection: doccodeController.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("pos_setting_doccode"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return 'This field is required';
            }
            return null;
          },
          onEditingComplete: () {},
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          enabled: isEditMode,
          textAlign: TextAlign.left,
          controller: docformatinv,
          textCapitalization: TextCapitalization.characters,
          // inputFormatters: <TextInputFormatter>[FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]'))],
          onChanged: (value) {
            isDataChange = true;
            screenData.docformatinv = value;
            docformatinv.value = TextEditingValue(
              text: value,
              selection: docformatinv.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("pos_setting_inv"),
          ),
          onEditingComplete: () {},
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          enabled: isEditMode,
          textAlign: TextAlign.left,
          controller: docformattaxinv,
          textCapitalization: TextCapitalization.characters,
          // inputFormatters: <TextInputFormatter>[FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]'))],
          onChanged: (value) {
            isDataChange = true;
            screenData.docformattaxinv = value;
            docformattaxinv.value = TextEditingValue(
              text: value,
              selection: docformattaxinv.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("pos_setting_taxinv"),
          ),
          onEditingComplete: () {},
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          enabled: isEditMode,
          textAlign: TextAlign.left,
          controller: docformatesalereturn,
          textCapitalization: TextCapitalization.characters,
          // inputFormatters: <TextInputFormatter>[FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]'))],
          onChanged: (value) {
            isDataChange = true;
            screenData.docformatesalereturn = value;
            docformatesalereturn.value = TextEditingValue(
              text: value,
              selection: docformatesalereturn.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("pos_setting_sale_return"),
          ),
          onEditingComplete: () {},
        ),
      ),
    );

    if (global.posVersion == global.PosVersionEnum.restaurant) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Expanded(
                child: InputDecorator(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("category_group_number"),
                  ),
                  child: DropdownButtonHideUnderline(
                    child: DropdownButton<int>(
                      value: screenData.categorygroupnumber,
                      icon: const Icon(Icons.arrow_drop_down),
                      style: const TextStyle(color: Colors.deepPurple),
                      underline: Container(color: Colors.deepPurpleAccent),
                      onChanged: (int? value) {
                        setState(() {
                          screenData.categorygroupnumber = value!;
                        });
                      },
                      isDense: true,
                      isExpanded: true,
                      items: global.groupNumber.map<DropdownMenuItem<int>>((
                        int value,
                      ) {
                        return DropdownMenuItem<int>(
                          value: value.toInt(),
                          child: Text(
                            "${global.language("category_group_number")} ${value.toInt()}",
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ),
              ),
              SizedBox(width: 10),
              Expanded(
                child: InputDecorator(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("kitchen_group_number"),
                  ),
                  child: DropdownButtonHideUnderline(
                    child: DropdownButton<int>(
                      value: screenData.kitchengroupnumber,
                      icon: const Icon(Icons.arrow_drop_down),
                      style: const TextStyle(color: Colors.deepPurple),
                      underline: Container(color: Colors.deepPurpleAccent),
                      onChanged: (int? value) {
                        setState(() {
                          screenData.kitchengroupnumber = value!;
                        });
                      },
                      isDense: true,
                      isExpanded: true,
                      items: global.groupNumber.map<DropdownMenuItem<int>>((
                        int value,
                      ) {
                        return DropdownMenuItem<int>(
                          value: value.toInt(),
                          child: Text(
                            "${global.language("kitchen_group_number")} ${value.toInt()}",
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      );

      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Expanded(
                child: InputDecorator(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("table_group_number"),
                  ),
                  child: DropdownButtonHideUnderline(
                    child: DropdownButton<int>(
                      value: screenData.tablegroupnumber,
                      icon: const Icon(Icons.arrow_drop_down),
                      style: const TextStyle(color: Colors.deepPurple),
                      underline: Container(color: Colors.deepPurpleAccent),
                      onChanged: (int? value) {
                        setState(() {
                          screenData.tablegroupnumber = value!;
                        });
                      },
                      isDense: true,
                      isExpanded: true,
                      items: global.groupNumber.map<DropdownMenuItem<int>>((
                        int value,
                      ) {
                        return DropdownMenuItem<int>(
                          value: value.toInt(),
                          child: Text(
                            "${global.language("table_group_number")} ${value.toInt()}",
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ),
              ),
              SizedBox(width: 10),
              Expanded(
                child: InputDecorator(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("zone_group_number"),
                  ),
                  child: DropdownButtonHideUnderline(
                    child: DropdownButton<int>(
                      value: screenData.zonegroupnumber,
                      icon: const Icon(Icons.arrow_drop_down),
                      style: const TextStyle(color: Colors.deepPurple),
                      underline: Container(color: Colors.deepPurpleAccent),
                      onChanged: (int? value) {
                        setState(() {
                          screenData.zonegroupnumber = value!;
                        });
                      },
                      isDense: true,
                      isExpanded: true,
                      items: global.groupNumber.map<DropdownMenuItem<int>>((
                        int value,
                      ) {
                        return DropdownMenuItem<int>(
                          value: value.toInt(),
                          child: Text(
                            "${global.language("zone_group_number")} ${value.toInt()}",
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      );
    } else {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Expanded(
                child: InputDecorator(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("category_group_number"),
                  ),
                  child: DropdownButtonHideUnderline(
                    child: DropdownButton<int>(
                      value: screenData.categorygroupnumber,
                      icon: const Icon(Icons.arrow_drop_down),
                      style: const TextStyle(color: Colors.deepPurple),
                      underline: Container(color: Colors.deepPurpleAccent),
                      onChanged: (int? value) {
                        setState(() {
                          screenData.categorygroupnumber = value!;
                        });
                      },
                      isDense: true,
                      isExpanded: true,
                      items: global.groupNumber.map<DropdownMenuItem<int>>((
                        int value,
                      ) {
                        return DropdownMenuItem<int>(
                          value: value.toInt(),
                          child: Text(
                            "${global.language("category_group_number")} ${value.toInt()}",
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.only(bottom: 5),
              width: double.infinity,
              height: 50,
              child: ElevatedButton(
                style: (validationBranch)
                    ? ButtonStyle(
                        backgroundColor: WidgetStateProperty.all<Color>(
                          global.theme.buttonDangerColor,
                        ),
                        shape: WidgetStateProperty.all<OutlinedBorder>(
                          RoundedRectangleBorder(
                            side: const BorderSide(
                              color: Colors.red,
                              width: 3.0,
                            ), // Adjust the width as needed
                            borderRadius: BorderRadius.circular(8),
                          ),
                        ),
                      )
                    : ButtonStyle(
                        backgroundColor: WidgetStateProperty.all<Color>(
                          global.theme.buttonDangerColor,
                        ),
                      ),
                onPressed: (isEditMode)
                    ? () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (context) =>
                                const CompanyBranchSearchScreen(word: ""),
                          ),
                        ).then((value) {
                          setState(() {
                            SearchGuidCodeNameModel result = value;
                            if (result.isCancel == false) {
                              screenData.branch.guidfixed = result.guid;
                              screenData.branch.code = result.code;
                              screenData.branch.names = result.names;
                            }
                          });
                        });
                      }
                    : null,
                child: Text(
                  (screenData.branch.guidfixed.isEmpty)
                      ? global.language("select_company_branch")
                      : global.packName(screenData.branch.names),
                ),
              ),
            ),

            /// Text validation สาขา
            (validationBranch)
                ? Container(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      global.language("please_select_company_branch"),
                      style: const TextStyle(color: Colors.red),
                    ),
                  )
                : Container(),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.only(bottom: 5),
              width: double.infinity,
              height: 50,
              child: ElevatedButton(
                style: (validationWarehouse)
                    ? ButtonStyle(
                        backgroundColor: WidgetStateProperty.all<Color>(
                          global.theme.positiveHighlightTextColor,
                        ),
                        shape: WidgetStateProperty.all<OutlinedBorder>(
                          RoundedRectangleBorder(
                            side: const BorderSide(
                              color: Colors.red,
                              width: 3.0,
                            ), // Adjust the width as needed
                            borderRadius: BorderRadius.circular(8),
                          ),
                        ),
                      )
                    : ButtonStyle(
                        backgroundColor: WidgetStateProperty.all<Color>(
                          global.theme.positiveHighlightTextColor,
                        ),
                      ),
                onPressed: (isEditMode)
                    ? () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (context) =>
                                const ProductWarehouseSearchScreen(word: ""),
                          ),
                        ).then((value) {
                          setState(() {
                            SearchGuidCodeNameModel result = value;
                            if (result.isCancel == false) {
                              screenData.warehouse.guidfixed = result.guid;
                              screenData.warehouse.code = result.code;
                              screenData.warehouse.names = result.names;
                            }
                          });
                        });
                      }
                    : null,
                child: Text(
                  (screenData.warehouse.guidfixed.isEmpty)
                      ? global.language("select_warehouse")
                      : global.packName(screenData.warehouse.names),
                ),
              ),
            ),

            /// Text validation สาขา
            (validationWarehouse)
                ? Container(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      global.language("please_select_warehouse"),
                      style: const TextStyle(color: Colors.red),
                    ),
                  )
                : Container(),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Container(
          padding: const EdgeInsets.only(bottom: 5),
          width: double.infinity,
          height: 50,
          child: ElevatedButton(
            style: ButtonStyle(
              backgroundColor: WidgetStateProperty.all<Color>(
                global.theme.primaryColor,
              ),
            ),
            onPressed: (screenData.warehouse.guidfixed.isNotEmpty && isEditMode)
                ? () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => ProductLocationSearchScreen(
                          whcode: screenData.warehouse.code,
                        ),
                      ),
                    ).then((value) {
                      setState(() {
                        SearchGuidCodeNameModel result = value;
                        if (result.isCancel == false) {
                          screenData.location.code = result.code;
                          screenData.location.names = result.names;
                        }
                      });
                    });
                  }
                : null,
            child: Text(
              (screenData.location.code.isEmpty)
                  ? global.language("select_location")
                  : global.packName(screenData.location.names),
            ),
          ),
        ),
      ),
    );
    if (screenData.qrcodes!.isNotEmpty) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Column(
            children: [
              Divider(),
              Align(
                alignment: Alignment.centerLeft,
                child: Text(global.language("qrcode_list")),
              ),
            ],
          ),
        ),
      );
    }

    for (int i = 0; i < screenData.qrcodes!.length; i++) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Container(
            decoration: BoxDecoration(
              border: Border.all(width: 1, color: global.theme.textSecondaryColor),
              borderRadius: BorderRadius.circular(8),
              color: global.theme.dividerBorderColor,
            ),
            width: double.infinity,
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Column(
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: SizedBox(
                          height: 40,
                          child: ElevatedButton(
                            style: ButtonStyle(
                              backgroundColor: WidgetStateProperty.all<Color>(
                                global.theme.negativeHighlightTextColor,
                              ),
                            ),
                            onPressed: (isEditMode)
                                ? () {
                                    Navigator.push(
                                      context,
                                      MaterialPageRoute(
                                        builder: (context) =>
                                            const QrSearchScreen(word: ""),
                                      ),
                                    ).then((value) {
                                      setState(() {
                                        if (value.guidfixed.isNotEmpty) {
                                          screenData.qrcodes![i] = value;
                                          bookBankQrSerchController[i].text =
                                              "${value.bankcode} ~ ${global.packName(value.bookbanknames!)}";
                                        }
                                      });
                                    });
                                  }
                                : null,
                            child: Row(
                              children: [
                                Icon(
                                  (screenData.qrcodes![i].guidfixed!.isEmpty)
                                      ? Icons.search
                                      : Icons.qr_code,
                                ),
                                SizedBox(width: 10),
                                Text(
                                  (screenData.qrcodes![i].guidfixed!.isEmpty)
                                      ? global.language("serch_qrcode")
                                      : "${screenData.qrcodes![i].code} ~ ${global.packName(screenData.qrcodes![i].qrnames!)}",
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      IconButton(
                        onPressed: (isEditMode)
                            ? () {
                                setState(() {
                                  screenData.qrcodes!.removeAt(i);
                                });
                              }
                            : null,
                        icon: const Icon(Icons.delete),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
                  (screenData.qrcodes![i].code!.isNotEmpty)
                      ? Column(
                          children: [
                            (screenData.qrcodes![i].qrtype! != 401)
                                ? TextFormField(
                                    readOnly: true,
                                    textInputAction: TextInputAction.next,
                                    textAlign: TextAlign.left,
                                    controller: bookBankQrSerchController[i],
                                    textCapitalization:
                                        TextCapitalization.characters,
                                    inputFormatters: <TextInputFormatter>[
                                      FilteringTextInputFormatter.allow(
                                        RegExp('[a-z A-Z 0-9 -]'),
                                      ),
                                    ],
                                    onChanged: (value) {},
                                    decoration: InputDecoration(
                                      border: const OutlineInputBorder(),
                                      contentPadding: const EdgeInsets.only(
                                        left: 10,
                                        top: 0,
                                        bottom: 0,
                                        right: 10,
                                      ),
                                      floatingLabelBehavior:
                                          FloatingLabelBehavior.always,
                                      labelText: global.language(
                                        "bookBank_code",
                                      ),
                                      labelStyle: TextStyle(
                                        color:
                                            global.theme.inputTextBoxForceColor,
                                      ),
                                      suffixIcon: IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
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
                                                      BookBankModel
                                                      returnValue =
                                                          value
                                                              as BookBankModel;
                                                      bookBankQrSerchController[i]
                                                              .text =
                                                          "${returnValue.bankcode} ~ ${returnValue.passbook}";
                                                      screenData
                                                              .qrcodes![i]
                                                              .bankcode =
                                                          returnValue.bankcode;
                                                      screenData
                                                              .qrcodes![i]
                                                              .banknames =
                                                          returnValue.banknames;
                                                      screenData
                                                              .qrcodes![i]
                                                              .bookbankcode =
                                                          returnValue.passbook;
                                                      screenData
                                                              .qrcodes![i]
                                                              .bookbanknames =
                                                          returnValue.names;
                                                      screenData
                                                              .qrcodes![i]
                                                              .bookbankimages =
                                                          returnValue.images;
                                                    });
                                                  }
                                                })
                                              : null;
                                        },
                                      ),
                                    ),
                                  )
                                : Container(),
                            const SizedBox(height: 15),
                            (screenData.qrcodes![i].qrtype == 100)
                                ? fromPromptPay(i)
                                : (screenData.qrcodes![i].qrtype == 101)
                                ? fromKPlusShop(i)
                                : (screenData.qrcodes![i].qrtype! <= 130)
                                ? fromLugent(i)
                                : (screenData.qrcodes![i].qrtype! <= 135)
                                ? fromGB(i)
                                : (screenData.qrcodes![i].qrtype! == 201)
                                ? fromXendit(i)
                                : (screenData.qrcodes![i].qrtype! == 301 ||
                                      screenData.qrcodes![i].qrtype! == 302)
                                ? fromSMLAPI(i)
                                : (screenData.qrcodes![i].qrtype! == 401)
                                ? fromTigerBoard(i)
                                : Container(),
                          ],
                        )
                      : Container(),
                ],
              ),
            ),
          ),
        ),
      );
    }

    if (isSaveAllow) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              style: ButtonStyle(
                backgroundColor: WidgetStateProperty.all<Color>(
                  global.theme.surfaceColor,
                ),
                foregroundColor: WidgetStateProperty.all<Color>(
                  global.theme.textColor,
                ),
              ),
              focusNode: FocusNode(skipTraversal: true),
              onPressed: (isEditMode)
                  ? () {
                      setState(() {
                        bookBankQrSerchController.add(TextEditingController());
                        screenData.qrcodes?.add(
                          QrModel(
                            guidfixed: "",
                            code: "",
                            qrnames: [],
                            qrtype: 100,
                            isactive: true,
                            logo: "",
                            bankcode: "",
                            banknames: [],
                            bookbankcode: "",
                            bookbanknames: [],
                            bookbankimages: [],
                            qrcode: "",
                            apikey: "",
                            billerCode: "",
                            billerID: "",
                            storeID: "",
                            terminalID: "",
                            merchantName: "",
                            accessCode: "",
                            bankcharge: "",
                            customercharge: "",
                          ),
                        );
                      });
                    }
                  : null,
              icon: Icon(Icons.qr_code),
              label: Text(global.language("add_qrcode")),
            ),
          ),
        ),
      );
    }

    if (screenData.creditcards!.isNotEmpty) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Column(
            children: [
              Divider(),
              Align(
                alignment: Alignment.centerLeft,
                child: Text(global.language("creditcard_list")),
              ),
            ],
          ),
        ),
      );
    }

    for (int i = 0; i < screenData.creditcards!.length; i++) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Container(
            decoration: BoxDecoration(
              border: Border.all(width: 1, color: global.theme.textSecondaryColor),
              borderRadius: BorderRadius.circular(8),
              color: global.theme.dividerBorderColor,
            ),
            width: double.infinity,
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      IconButton(
                        onPressed: (isEditMode)
                            ? () {
                                setState(() {
                                  screenData.creditcards!.removeAt(i);
                                });
                              }
                            : null,
                        icon: const Icon(Icons.delete),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
                  Column(
                    children: [
                      Column(
                        children: listNamesFields(
                          screenData.creditcards![i].names!,
                          "name",
                        ),
                      ),
                      SizedBox(height: 10),
                      TextFormField(
                        readOnly: true,
                        textInputAction: TextInputAction.next,
                        textAlign: TextAlign.left,
                        controller: bookBankCreditcardSerchController[i],
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
                          labelText: global.language("bookbank_code"),
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
                                          bookBankCreditcardSerchController[i]
                                                  .text =
                                              "${returnValue.bankcode} ~ ${returnValue.passbook}";
                                          screenData.creditcards![i].bookbank =
                                              returnValue;
                                        });
                                      }
                                    })
                                  : null;
                            },
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      );
    }

    if (isSaveAllow) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              style: ButtonStyle(
                backgroundColor: WidgetStateProperty.all<Color>(
                  global.theme.positiveHighlightColor,
                ),
                foregroundColor: WidgetStateProperty.all<Color>(
                  global.theme.textColor,
                ),
              ),
              focusNode: FocusNode(skipTraversal: true),
              onPressed: (isEditMode)
                  ? () {
                      setState(() {
                        bookBankCreditcardSerchController.add(
                          TextEditingController(),
                        );
                        screenData.creditcards?.add(
                          CreditcardsModel(
                            names: [],
                            bookbank: BookBankModel(),
                          ),
                        );
                      });
                    }
                  : null,
              icon: Icon(Icons.credit_card),
              label: Text(global.language("add_creditcard")),
            ),
          ),
        ),
      );
    }

    if (screenData.transfers!.isNotEmpty) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Column(
            children: [
              Divider(),
              Align(
                alignment: Alignment.centerLeft,
                child: Text(global.language("transfers_list")),
              ),
            ],
          ),
        ),
      );
    }

    for (int i = 0; i < screenData.transfers!.length; i++) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Container(
            decoration: BoxDecoration(
              border: Border.all(width: 1, color: global.theme.textSecondaryColor),
              borderRadius: BorderRadius.circular(8),
              color: global.theme.dividerBorderColor,
            ),
            width: double.infinity,
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      IconButton(
                        onPressed: (isEditMode)
                            ? () {
                                setState(() {
                                  screenData.transfers!.removeAt(i);
                                });
                              }
                            : null,
                        icon: const Icon(Icons.delete),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
                  Column(
                    children: [
                      Column(
                        children: listNamesFields(
                          screenData.transfers![i].names!,
                          "name",
                        ),
                      ),
                      SizedBox(height: 10),
                      TextFormField(
                        readOnly: true,
                        textInputAction: TextInputAction.next,
                        textAlign: TextAlign.left,
                        controller: bookBankTransfercardSerchController[i],
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
                          labelText: global.language("bookbank_code"),
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
                                          bookBankTransfercardSerchController[i]
                                                  .text =
                                              "${returnValue.bankcode} ~ ${returnValue.passbook}";
                                          screenData.transfers![i].bookbank =
                                              returnValue;
                                        });
                                      }
                                    })
                                  : null;
                            },
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      );
    }

    if (isSaveAllow) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              style: ButtonStyle(
                backgroundColor: WidgetStateProperty.all<Color>(
                  global.theme.warningHighlightColor,
                ),
                foregroundColor: WidgetStateProperty.all<Color>(
                  global.theme.textColor,
                ),
              ),
              focusNode: FocusNode(skipTraversal: true),
              onPressed: (isEditMode)
                  ? () {
                      setState(() {
                        bookBankTransfercardSerchController.add(
                          TextEditingController(),
                        );
                        screenData.transfers?.add(
                          TransfersModel(names: [], bookbank: BookBankModel()),
                        );
                      });
                    }
                  : null,
              icon: Icon(Icons.transform_outlined),
              label: Text(global.language("add_transfer")),
            ),
          ),
        ),
      );
    }

    if (screenData.employees!.isNotEmpty) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Column(
            children: [
              Divider(),
              Align(
                alignment: Alignment.centerLeft,
                child: Text(global.language("employee_list")),
              ),
            ],
          ),
        ),
      );
    }

    for (int i = 0; i < screenData.employees!.length; i++) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: Container(
            padding: const EdgeInsets.only(bottom: 5),
            width: double.infinity,
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: ElevatedButton(
                        style: ButtonStyle(
                          backgroundColor: WidgetStateProperty.all<Color>(
                            global.theme.textSecondaryColor,
                          ),
                        ),
                        onPressed: (isEditMode)
                            ? () {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (context) =>
                                        const EmployeeSearchScreen(word: ""),
                                  ),
                                ).then((value) {
                                  setState(() {
                                    if (value.guidfixed.isNotEmpty) {
                                      screenData.employees![i].guidfixed =
                                          value.guidfixed;
                                      screenData.employees![i].code =
                                          value.code;
                                      screenData.employees![i].name =
                                          value.name;
                                    }
                                  });
                                });
                              }
                            : null,
                        child: Row(
                          children: [
                            Icon(
                              (screenData.employees![i].guidfixed!.isEmpty)
                                  ? Icons.person_search
                                  : Icons.person,
                            ),
                            SizedBox(width: 10),
                            Text(
                              (screenData.employees![i].guidfixed!.isEmpty)
                                  ? global.language("serch_employee")
                                  : "${screenData.employees![i].code}  ~  ${screenData.employees![i].name}",
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    IconButton(
                      onPressed: (isEditMode)
                          ? () {
                              setState(() {
                                screenData.employees!.removeAt(i);
                              });
                            }
                          : null,
                      icon: const Icon(Icons.delete),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                /* XXXX
                  DropdownSearch<String>.multiSelection(
                    enabled: (screenData.employees![i].guidfixed!.isNotEmpty && isEditMode),
                    items: permistionList,
                    dropdownDecoratorProps: DropDownDecoratorProps(
                      dropdownSearchDecoration: InputDecoration(
                        border: const OutlineInputBorder(),
                        contentPadding: const EdgeInsets.only(left: 10, top: 0, bottom: 0, right: 10),
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        labelText: global.language("permistion"),
                      ),
                    ),
                    onChanged: (List<String> value) {
                      setState(() {
                        screenData.employees![i].permissions = value;
                      });
                    },
                    popupProps: const PopupPropsMultiSelection.dialog(
                      showSearchBox: true,
                      showSelectedItems: true,
                    ),
                    selectedItems: screenData.employees![i].permissions!,
                  ),*/
              ],
            ),
          ),
        ),
      );
    }
    if (isSaveAllow) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              style: ButtonStyle(
                backgroundColor: WidgetStateProperty.all<Color>(
                  global.theme.infoHighlightColor,
                ),
                foregroundColor: WidgetStateProperty.all<Color>(
                  global.theme.textColor,
                ),
              ),
              focusNode: FocusNode(skipTraversal: true),
              onPressed: (isEditMode)
                  ? () {
                      setState(() {
                        screenData.employees?.add(
                          EmployeePosModel(
                            code: '',
                            guidfixed: '',
                            name: '',
                            permissions: [
                              permistionList.first,
                              permistionList.last,
                            ],
                          ),
                        );
                      });
                    }
                  : null,
              icon: Icon(Icons.person_add_alt_1),
              label: Text(global.language("add_employee")),
            ),
          ),
        ),
      );
    }

    /// sale channel
    if (screenData.salechannels!.isNotEmpty) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Column(
            children: [
              Divider(),
              Align(
                alignment: Alignment.centerLeft,
                child: Text(global.language("sale_channel_list")),
              ),
            ],
          ),
        ),
      );
    }

    for (int i = 0; i < screenData.salechannels!.length; i++) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: Container(
            padding: const EdgeInsets.only(bottom: 5),
            width: double.infinity,
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: ElevatedButton(
                        style: ButtonStyle(
                          backgroundColor: WidgetStateProperty.all<Color>(
                            global.theme.primaryColor,
                          ),
                          foregroundColor: WidgetStateProperty.all<Color>(
                            global.theme.onPrimaryColor,
                          ),
                        ),
                        onPressed: (isEditMode)
                            ? () {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (context) =>
                                        const SalechannelsSearchScreen(
                                          word: "",
                                        ),
                                  ),
                                ).then((value) {
                                  setState(() {
                                    if (value.guidfixed.isNotEmpty) {
                                      screenData.salechannels![i] = value;
                                    }
                                  });
                                });
                              }
                            : null,
                        child: Row(
                          children: [
                            Text(
                              (screenData.salechannels![i].code!.isEmpty)
                                  ? global.language("serch_sale_channel")
                                  : "${screenData.salechannels![i].code}  ~  ${screenData.salechannels![i].name}",
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    IconButton(
                      onPressed: (isEditMode)
                          ? () {
                              setState(() {
                                screenData.salechannels!.removeAt(i);
                              });
                            }
                          : null,
                      icon: const Icon(Icons.delete),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
              ],
            ),
          ),
        ),
      );
    }
    if (isSaveAllow) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              style: ButtonStyle(
                backgroundColor: WidgetStateProperty.all<Color>(
                  global.theme.primaryColor,
                ),
                foregroundColor: WidgetStateProperty.all<Color>(
                  global.theme.onPrimaryColor,
                ),
              ),
              focusNode: FocusNode(skipTraversal: true),
              onPressed: (isEditMode)
                  ? () {
                      setState(() {
                        screenData.salechannels?.add(SaleChannelModel());
                      });
                    }
                  : null,
              icon: Icon(Icons.add_shopping_cart_rounded),
              label: Text(global.language("add_sale_channel")),
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: SizedBox(
          child: Column(
            children: [
              ListView.builder(
                shrinkWrap: true,
                itemCount: screenData.slips!.length,
                itemBuilder: (context, index) {
                  SlipModel slip = screenData.slips![index];
                  return ListTile(
                    title: Row(
                      children: [
                        Text(
                          "${slip.code}.${slip.name}",
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: (index <= 3)
                                ? Colors.blue[900]
                                : Colors.red[900],
                          ),
                        ),
                      ],
                    ),
                    subtitle: Padding(
                      padding: EdgeInsets.all(8.0),
                      child: Column(
                        children: [
                          /* XXXX
                          DropdownSearch<SlipListModel>(
                            enabled: isEditMode,
                            asyncItems: (String filter) => getDataSlip(filter),
                            compareFn: (item, selectedItem) => item.code == selectedItem.code,
                            itemAsString: (SlipListModel? slip) {
                              if (slip!.code.isEmpty) return '';
                              return '${slip.code} - ${global.packName(slip.names)}';
                            },
                            dropdownDecoratorProps: DropDownDecoratorProps(
                              dropdownSearchDecoration: InputDecoration(
                                border: const OutlineInputBorder(),
                                contentPadding: const EdgeInsets.only(left: 10, top: 0, bottom: 0, right: 10),
                                floatingLabelBehavior: FloatingLabelBehavior.always,
                                // labelText: global.language("slip_type"),
                                filled: true, // Set filled property to true
                                fillColor: (index <= 3) ? Colors.blue[50] : Colors.red[50], // Set the desired background color
                              ),
                            ),
                            onChanged: (SlipListModel? value) {
                              setState(() {
                                screenData.slips![index].formcode = value!.code;
                                screenData.slips![index].formnames = value.names;
                              });
                            },
                            popupProps: const PopupPropsMultiSelection.dialog(
                              showSearchBox: true,
                              showSelectedItems: true,
                            ),
                            selectedItem: slipSeleted[index],
                          ),*/
                          const SizedBox(height: 10),
                          Column(
                            children: listNamesFields(
                              screenData.slips![index].headernames,
                              "header_names",
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );

    /// สื่อโฆษณา
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Container(
          padding: const EdgeInsets.only(bottom: 5),
          width: double.infinity,
          height: 50,
          child: ElevatedButton(
            style: ButtonStyle(
              backgroundColor: WidgetStateProperty.all<Color>(
                global.theme.secondaryColor,
              ),
            ),
            onPressed: (isEditMode)
                ? () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) =>
                            const PosMediaSearchScreen(word: ''),
                      ),
                    ).then((value) {
                      setState(() {
                        SearchGuidCodeNameModel result = value;
                        if (result.isCancel == false) {
                          screenData.mediaguid = result.guid;

                          posMediaData.code = result.code;
                          posMediaData.description = result.names;
                        }
                      });
                    });
                  }
                : null,
            child: Text(
              (screenData.mediaguid!.isEmpty)
                  ? global.language("select_media")
                  : '${posMediaData.code} ~ ${global.packName(posMediaData.description)}',
            ),
          ),
        ),
      ),
    );

    /* XXXX
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: DropdownSearch<TimezonesModel>(
          enabled: isEditMode,
          asyncItems: (String filter) => global.getTimezonesList(filter),
          compareFn: (item, selectedItem) => item.text == selectedItem.text,
          itemAsString: (TimezonesModel? timezone) {
            if (timezone!.text.isEmpty) return '';
            return timezone.text;
          },
          dropdownDecoratorProps: DropDownDecoratorProps(
            dropdownSearchDecoration: InputDecoration(
              border: const OutlineInputBorder(),
              contentPadding: const EdgeInsets.only(left: 10, top: 0, bottom: 0, right: 10),
              floatingLabelBehavior: FloatingLabelBehavior.always,
              labelText: global.language("timezone"),
            ),
          ),
          onChanged: (isEditMode)
              ? (TimezonesModel? value) {
                  setState(() {
                    screenData.timezoneoffset = value!.offset;
                    screenData.timezonelabel = value.text;
                    timezoneSelected = value;
                  });
                }
              : null,
          popupProps: const PopupPropsMultiSelection.dialog(
            showSearchBox: true,
            showSelectedItems: true,
          ),
          selectedItem: timezoneSelected,
        ),
      ),
    ); */

    if (screenData.timeforsales.isNotEmpty) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Column(
            children: [
              Divider(),
              Align(
                alignment: Alignment.centerLeft,
                child: Text(global.language("time_for_sale")),
              ),
            ],
          ),
        ),
      );
    }

    for (int i = 0; i < screenData.timeforsales.length; i++) {
      formWidgets.add(
        Container(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          width: double.infinity,
          child: Column(
            children: [
              const SizedBox(height: 10),
              Column(
                children: listNamesFields(
                  screenData.timeforsales[i].names,
                  "name",
                ),
              ),
              Row(
                children: [
                  Expanded(
                    child: TextFormField(
                      enabled: isEditMode,
                      controller: mediaFromTimeController[i],
                      onTap: () async {
                        _selecttimeforsalesFrom(context, i);
                      },
                      readOnly: true,
                      decoration: InputDecoration(
                        labelText: global.language("from_time"),
                        border: const OutlineInputBorder(),
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        suffixIcon: Row(
                          mainAxisAlignment:
                              MainAxisAlignment.spaceBetween, // added line
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            IconButton(
                              focusNode: FocusNode(skipTraversal: true),
                              icon: const Icon(Icons.timer_sharp),
                              onPressed: () {
                                _selecttimeforsalesFrom(context, i);
                              },
                            ),
                          ],
                        ),
                      ),
                      inputFormatters: [global.TimeInputFormatter()],
                    ),
                  ),
                  SizedBox(width: 10),
                  Expanded(
                    child: TextFormField(
                      enabled: isEditMode,
                      controller: mediaToTimeController[i],
                      onTap: () async {
                        _selectMediaToTime(context, i);
                      },
                      readOnly: true,
                      decoration: InputDecoration(
                        labelText: global.language("from_to"),
                        border: const OutlineInputBorder(),
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        suffixIcon: Row(
                          mainAxisAlignment:
                              MainAxisAlignment.spaceBetween, // added line
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            IconButton(
                              focusNode: FocusNode(skipTraversal: true),
                              icon: const Icon(Icons.timer_sharp),
                              onPressed: () {
                                _selectMediaToTime(context, i);
                              },
                            ),
                          ],
                        ),
                      ),
                      inputFormatters: [global.TimeInputFormatter()],
                    ),
                  ),
                  const SizedBox(width: 10),
                  IconButton(
                    onPressed: (isEditMode)
                        ? () {
                            setState(() {
                              mediaFromTimeController.removeAt(i);
                              mediaToTimeController.removeAt(i);
                              screenData.timeforsales.removeAt(i);
                            });
                          }
                        : null,
                    icon: const Icon(Icons.delete),
                  ),
                ],
              ),
            ],
          ),
        ),
      );
    }

    if (isSaveAllow) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              style: ButtonStyle(
                backgroundColor: WidgetStateProperty.all<Color>(
                  global.theme.positiveHighlightColor,
                ),
                foregroundColor: WidgetStateProperty.all<Color>(
                  global.theme.textColor,
                ),
              ),
              focusNode: FocusNode(skipTraversal: true),
              onPressed: (isEditMode)
                  ? () {
                      setState(() {
                        _selectedTime.add(TimeOfDay.now());
                        mediaFromTimeController.add(TextEditingController());
                        mediaToTimeController.add(TextEditingController());
                        screenData.timeforsales.add(
                          TimeForsaleModel(from: '', names: [], to: ''),
                        );
                      });
                    }
                  : null,
              icon: Icon(Icons.access_time_sharp),
              label: Text(global.language("add_time_for_sale")),
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Column(
          children: listNamesFields(
            screenData.billheader!,
            "header_receipt_pos",
          ),
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Column(
          children: listNamesFields(
            screenData.billfooter!,
            "footer_receipt_pos",
          ),
        ),
      ),
    );

    if (isEditMode) {
      formWidgets.add(
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
                      imageWeb = null;
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
                            final XFile? photo = await imagePicker.pickImage(
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
                      final XFile? photo = await imagePicker.pickImage(
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
      );
    }

    formWidgets.add(
      Container(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
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
                  final bytes = await dropZoneController.getFileData(ev);
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
                      : (screenData.logourl != '')
                      ? DecorationImage(
                          image: NetworkImage(screenData.logourl!),
                          fit: BoxFit.fill,
                        )
                      : const DecorationImage(
                          image: AssetImage('assets/img/noimage.png'),
                        ),
                ),
                child: const SizedBox(width: double.infinity, height: 400),
              ),
            ),
          ],
        ),
      ),
    );

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: const EdgeInsets.only(
            left: 10,
            right: 10,
            bottom: 15,
            top: 10,
          ),
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
        title: Text(headerEdit + global.language("pos_setting")),
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
                            context.read<PosSettingBloc>().add(
                              PosSettingDelete(guid: selectGuid),
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
      body: Focus(
        skipTraversal: true,
        onKeyEvent: (node, event) {
          if (event is KeyUpEvent) return KeyEventResult.ignored;
          if (event.logicalKey == LogicalKeyboardKey.f10) {
            if (event is KeyDownEvent) {
              if (_formKey.currentState!.validate()) {
                saveOrUpdateData();
              }
            }
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
      ),
    );
  }

  Column fromPromptPay(int i) {
    return Column(
      children: [
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          maxLength: 14,
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].qrcode,
          ),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: [FilteringTextInputFormatter.allow(RegExp('[0-9]'))],
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].qrcode = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("telephone_or_citizen_id"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
      ],
    );
  }

  Column fromXendit(int i) {
    return Column(
      children: [
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].apikey,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].apikey = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("api_key"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
      ],
    );
  }

  Column fromSMLAPI(int i) {
    return Column(
      children: [
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].apikey,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].apikey = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("api_key"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
      ],
    );
  }

  Column fromKPlusShop(int i) {
    return Column(
      children: [
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].billerID,
          ),
          textCapitalization: TextCapitalization.characters,
          // inputFormatters: [FilteringTextInputFormatter.allow(RegExp('[0-9]'))],
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].billerID = value.toUpperCase();
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("merchant_id"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
      ],
    );
  }

  Column fromLugent(int i) {
    return Column(
      children: [
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].apikey,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].apikey = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("api_key"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].billerCode,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].billerCode = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("biller_code"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].billerID,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].billerID = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("biller_id"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].storeID,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].storeID = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("store_id"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].terminalID,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].terminalID = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("terminal_id"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].merchantName,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].merchantName = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("merchant_name"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].accessCode,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].accessCode = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("access_code"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].bankcharge,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].bankcharge = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("bank_charge"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].customercharge,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].customercharge = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("customer_charge"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 0),
          child: Row(
            children: [
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 0,
                groupValue: screenData.qrcodes![i].closeqr,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].closeqr = value!;
                  });
                },
              ),
              Text(global.language('money_now')),
              const SizedBox(width: 10),
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 1,
                groupValue: screenData.qrcodes![i].closeqr,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].closeqr = value!;
                  });
                },
              ),
              Text(global.language('money_end_day')),
              const SizedBox(width: 10),
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 2,
                groupValue: screenData.qrcodes![i].closeqr,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].closeqr = value!;
                  });
                },
              ),
              Text(global.language('money_next_day')),
            ],
          ),
        ),
        SizedBox(height: 15),
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 0),
          child: Row(
            children: [
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: true,
                groupValue: screenData.qrcodes![i].isactive,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].isactive = value!;
                  });
                },
              ),
              Text(global.language('Active')),
              const SizedBox(width: 10),
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: false,
                groupValue: screenData.qrcodes![i].isactive,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].isactive = value!;
                  });
                },
              ),
              Text(global.language('inactive')),
            ],
          ),
        ),
      ],
    );
  }

  Column fromGB(int i) {
    return Column(
      children: [
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].apikey,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].apikey = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("api_key"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].accessCode,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].accessCode = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("access_code"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.qrcodes![i].token),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].token = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("token"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 0),
          child: Row(
            children: [
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 0,
                groupValue: screenData.qrcodes![i].closeqr,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].closeqr = value!;
                  });
                },
              ),
              Text(global.language('money_now')),
              const SizedBox(width: 10),
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 1,
                groupValue: screenData.qrcodes![i].closeqr,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].closeqr = value!;
                  });
                },
              ),
              Text(global.language('money_end_day')),
              const SizedBox(width: 10),
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: 2,
                groupValue: screenData.qrcodes![i].closeqr,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].closeqr = value!;
                  });
                },
              ),
              Text(global.language('money_next_day')),
            ],
          ),
        ),
        SizedBox(height: 15),
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 0),
          child: Row(
            children: [
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: true,
                groupValue: screenData.qrcodes![i].isactive,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].isactive = value!;
                  });
                },
              ),
              Text(global.language('Active')),
              const SizedBox(width: 10),
              Radio(
                focusNode: FocusNode(skipTraversal: true),
                value: false,
                groupValue: screenData.qrcodes![i].isactive,
                onChanged: (value) {
                  setState(() {
                    screenData.qrcodes![i].isactive = value!;
                  });
                },
              ),
              Text(global.language('inactive')),
            ],
          ),
        ),
      ],
    );
  }

  Column fromTigerBoard(int i) {
    return Column(
      children: [
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.qrcodes![i].host),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].host = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language('host'),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        const SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.qrcodes![i].appid),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].appid = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language('app_id_label'),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
        SizedBox(height: 15),
        TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {},
          textInputAction: TextInputAction.next,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.qrcodes![i].apikey,
          ),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.qrcodes![i].apikey = value;
          },
          decoration: InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("api_key"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
      ],
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
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<WarehouseBloc, WarehouseState>(
                listener: (context, state) {
                  if (state is WarehouseGetSuccess) {
                    setState(() {
                      if (state.warehouse.guidfixed.isNotEmpty) {
                        screenData.warehouse.guidfixed =
                            state.warehouse.guidfixed;
                        screenData.warehouse.code = state.warehouse.code;
                        screenData.warehouse.names = state.warehouse.names;
                      }
                    });
                  }
                },
              ),
              BlocListener<LoginBloc, LoginState>(
                listener: (context, state) {
                  /// Logout
                  if (state is LogoutSuccess) {
                    Navigator.pushNamedAndRemoveUntil(
                      context,
                      '/login_screen',
                      (route) => false,
                    );
                  }
                },
              ),
              BlocListener<PosSettingBloc, PosSettingState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is PosSettingLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.posSettings.isNotEmpty) {
                        listData.addAll(state.posSettings);
                      }
                    });
                  }
                  if (state is PosSettingLoadFailed) {
                    setState(() {
                      loadingData = false;
                      global.showSnackBar(
                        context,
                        const Icon(Icons.error_outline, color: Colors.white),
                        state.message,
                        Colors.red,
                      );
                    });
                  }
                  // Save
                  if (state is PosSettingSaveSuccess) {
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
                  if (state is PosSettingSaveFailed) {
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
                  if (state is PosSettingUpdateSuccess) {
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
                  if (state is PosSettingUpdateFailed) {
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
                  if (state is PosSettingDeleteSuccess) {
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
                  if (state is PosSettingDeleteManySuccess) {
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
                  if (state is PosSettingGetSuccess) {
                    setState(() {
                      isDataChange = false;
                      screenData = state.posSetting;

                      if (screenData.mediaguid!.isNotEmpty) {
                        context.read<PosMediaBloc>().add(
                          PosMediaGet(guid: screenData.mediaguid!),
                        );
                      }

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
                  if (state is GetApiKeySuccess) {
                    setState(() {
                      activeAPIKEY = state.token;

                      activatePin();
                    });
                  } else if (state is GetApiKeyFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        const Icon(Icons.error_outline, color: Colors.white),
                        state.message,
                        Colors.red,
                      );
                    });
                  }

                  if (state is DeleteApikeySuccess) {
                    setState(() {
                      activeAPIKEY = '';
                      screenData.activepin = "";
                      updateData(screenData.guidfixed);

                      /// close dialog
                      Navigator.pop(context);
                    });
                  } else if (state is DeleteApikeyFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        const Icon(Icons.error_outline, color: Colors.white),
                        state.message,
                        Colors.red,
                      );
                    });
                  }
                },
              ),
              BlocListener<PosMediaBloc, PosMediaState>(
                listener: (context, state) {
                  if (state is PosMediaGetSuccess) {
                    setState(() {
                      posMediaData = state.posMedia;
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

  List<Widget> listNamesFields(
    List<LanguageDataModel> names,
    String fieldname,
  ) {
    List<Widget> forms = [];
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel nameObj = names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (nameObj.code == '') {
        names.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
    }
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel nameObj = names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (nameObj.code != '') {
        forms.add(
          Padding(
            padding: EdgeInsets.only(bottom: 10),
            child: TextFormField(
              readOnly: !isEditMode,
              enabled: isEditMode,
              onChanged: (value) {
                isDataChange = true;
                nameObj.name = value;
              },
              textAlign: TextAlign.left,
              controller: TextEditingController(text: nameObj.name),
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: const OutlineInputBorder(),
                labelText:
                    "${global.language(fieldname)} (${getLangName(nameObj.code)})",
              ),
              validator: (value) {
                // if (languageIndex == 0) {
                //   if (value == null || value.isEmpty) {
                //     return 'This field is required';
                //   }
                // }

                return null;
              },
            ),
          ),
        );
      }
    }

    return forms;
  }
}

// Create an input widget that takes only one digit
class OtpInput extends StatelessWidget {
  final TextEditingController controller;
  final bool autoFocus;
  const OtpInput(this.controller, this.autoFocus, {super.key});

  @override
  Widget build(BuildContext context) {
    return Focus(
      focusNode: FocusNode(skipTraversal: true),
      onKey: (node, event) {
        if (event is RawKeyDownEvent) {
          if (event.logicalKey == LogicalKeyboardKey.backspace) {
            controller.clear();
            FocusScope.of(context).previousFocus();
          }
        }
        return KeyEventResult.ignored;
      },
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: SizedBox(
          height: 60,
          width: 50,
          child: TextField(
            autofocus: autoFocus,
            textAlign: TextAlign.center,
            keyboardType: TextInputType.number,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            controller: controller,
            maxLength: 1,
            cursorColor: Theme.of(context).primaryColor,
            decoration: InputDecoration(
              enabledBorder: OutlineInputBorder(
                borderSide: BorderSide(color: Colors.green, width: 1.0),
              ),
              focusedBorder: OutlineInputBorder(
                borderSide: BorderSide(
                  color: global.theme.primaryColor, // Border color when TextField is focused
                  width: 2.0,
                ),
              ),
              counterText: '',
              hintStyle: TextStyle(color: global.theme.textColor, fontSize: 20.0),
            ),
            onChanged: (value) {
              controller.text = value;
              FocusScope.of(context).nextFocus();
            },
          ),
        ),
      ),
    );
  }
}
