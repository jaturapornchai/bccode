import 'dart:convert';

import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/order_setting_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/report_repository.dart';
import 'package:smlaicloud/screen_search/order_template_setting_search_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class OrderSettingScreen extends StatefulWidget {
  const OrderSettingScreen({super.key});

  @override
  State<OrderSettingScreen> createState() => OrderSettingScreenState();
}

class OrderSettingScreenState extends State<OrderSettingScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  List<LanguageModel> languageList = <LanguageModel>[];
  List<OrderSettingModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late OrderSettingState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late OrderSettingModel screenData;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  bool loadingData = false;
  late DateTime dateNow = DateTime.now();
  final _formKey = GlobalKey<FormState>();

  TextEditingController codeController = TextEditingController();
  TextEditingController docformat = TextEditingController();

  String pinActive = "";
  TextEditingController field1 = TextEditingController();
  TextEditingController field2 = TextEditingController();
  TextEditingController field3 = TextEditingController();
  TextEditingController field4 = TextEditingController();
  TextEditingController field5 = TextEditingController();
  TextEditingController field6 = TextEditingController();
  TextEditingController field7 = TextEditingController();
  TextEditingController field8 = TextEditingController();

  bool isActivePin = false;
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
    docformat.dispose();

    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<OrderSettingBloc>().add(
      OrderSettingLoadList(
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
    listScrollController.addListener(onScrollList);

    screenData = OrderSettingModel(
      code: '',
      devicenumber: '',
      guidfixed: '',
      activepin: '',
      devicetype: 0,
      docformat: 'YYMMDD####',
      isposactive: false,
      settingcode: '',
      settingname: '',
    );

    isActivePin = false;
    isDataChange = false;
    setState(() {
      loadDataToScreen();
    });
  }

  void loadDataToScreen() {
    codeController.text = screenData.code!;
    docformat.text = screenData.docformat!;
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
    context.read<OrderSettingBloc>().add(OrderSettingGet(guid: guid));
  }

  void switchToEdit(OrderSettingModel value) {
    setState(() {
      selectGuid = value.guidfixed!;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Future<void> activatePin() async {
    /// 1 = dev , 0 = prod , 2 = uat
    int isDeveloper = global.isdevPin;

    ReportRepository reportRepository = ReportRepository();
    try {
      ApiResponse result = await reportRepository.activePos(
        pinActive,
        screenData.code!,
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

  void resetActivePin() {
    screenData.activepin = "";
  }

  void safelyPopContext() {
    if (mounted) {
      Navigator.pop(context);
    }
  }

  void handleException(dynamic e) {
    String errorMessage = getErrorMessage(e);
    String? jsonError = extractJsonErrorMessage(errorMessage);

    if (jsonError != null && jsonError == "Pin Number not found") {
      resetActivePin();
      updateData(screenData.guidfixed!);
      safelyPopContext();
    } else {
      showGlobalErrorSnackBar(errorMessage);
    }
  }

  String getErrorMessage(dynamic e) {
    return e is FormatException ? e.message : e.toString();
  }

  String? extractJsonErrorMessage(String message) {
    RegExp jsonRegex = RegExp(r'\{.*\}');
    Match? jsonMatch = jsonRegex.firstMatch(message);
    if (jsonMatch != null) {
      try {
        Map<String, dynamic> parsedMessage = jsonDecode(
          jsonMatch.group(0) ?? '',
        );
        return parsedMessage['msg'];
      } catch (_) {
        return null;
      }
    }
    return null;
  }

  void showGlobalErrorSnackBar(String message) {
    if (mounted) {
      global.showSnackBar(
        context,
        Icon(Icons.error, color: Colors.white),
        message,
        Colors.red,
      );
    }
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language("order_setting")),
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
                            context.read<OrderSettingBloc>().add(
                              OrderSettingDeleteMany(guid: guidListChecked),
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
                      global.language("order_setting_code"),
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
                      global.language("order_setting_doc_format"),
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

  Widget listObject(int index, OrderSettingModel value, bool showCheckBox) {
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
              flex: 5,
              child: Text(
                value.docformat!,
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
                    getData(value.guidfixed!);

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

  Future<String?> widgetActivePin(OrderSettingModel value) {
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

                            selectGuid = value.guidfixed!;

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
    if (verifyData(screenData)) {
      showCheckBox = false;
      // print(screenData.toJson());

      if (selectGuid.trim().isEmpty) {
        context.read<OrderSettingBloc>().add(
          OrderSettingSave(orderSetting: screenData),
        );
      } else {
        updateData(selectGuid);
      }
    }
  }

  bool verifyData(OrderSettingModel value) {
    List<String> errorList = [];
    if (screenData.settingcode!.isEmpty) {
      errorList.add(global.language("please_select_template_order_setting"));
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

    context.read<OrderSettingBloc>().add(
      OrderSettingUpdate(guid: guid, orderSetting: screenData),
    );
  }

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
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
            labelText: global.language("order_setting_code"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return global.language('this_field_is_required');
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
          controller: docformat,
          textCapitalization: TextCapitalization.characters,
          // inputFormatters: <TextInputFormatter>[FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]'))],
          onChanged: (value) {
            isDataChange = true;
            screenData.docformat = value;
            docformat.value = TextEditingValue(
              text: value,
              selection: docformat.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("order_setting_doc_format"),
          ),
          onEditingComplete: () {},
        ),
      ),
    );

    // formWidgets.add(
    //   Padding(
    // padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
    //     child: Row(
    //       children: [
    //         Radio(
    //           focusNode: FocusNode(skipTraversal: true),
    //           value: 0,
    //           groupValue: screenData.devicetype,
    //           onChanged: (isEditMode)
    //               ? (value) {
    //                   setState(() {
    //                     screenData.devicetype = value as int?;
    //                   });
    //                 }
    //               : null,
    //         ),
    //         Text(global.language("machine_employee")),
    //         const SizedBox(width: 10),
    //         Radio(
    //           focusNode: FocusNode(skipTraversal: true),
    //           value: 1,
    //           groupValue: screenData.devicetype,
    //           onChanged: (isEditMode)
    //               ? (value) {
    //                   setState(() {
    //                     screenData.devicetype = value as int?;
    //                   });
    //                 }
    //               : null,
    //         ),
    //         Text(global.language("machine_customer")),
    //       ],
    //     ),
    //   ),
    // );

    /// setting order
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
            onPressed: (isEditMode)
                ? () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) =>
                            const OrderTemplateSettingSearchScreen(word: ''),
                      ),
                    ).then((value) {
                      setState(() {
                        SearchGuidCodeNameModel result = value;
                        if (result.isCancel == false) {
                          screenData.settingcode = result.guid;
                          screenData.settingname = result.code;
                        }
                      });
                    });
                  }
                : null,
            child: Text(
              (screenData.settingcode!.isEmpty)
                  ? global.language("select_template_order_setting")
                  : '${global.language("type_order_setting")} : ${screenData.settingname}',
            ),
          ),
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
        title: Text(headerEdit + global.language("order_setting")),
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
                            context.read<OrderSettingBloc>().add(
                              OrderSettingDelete(guid: selectGuid),
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
      body: KeyboardListener(
        focusNode: FocusNode(),
        onKeyEvent: (KeyEvent event) {
          if (event is KeyDownEvent) {
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
              BlocListener<OrderSettingBloc, OrderSettingState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is OrderSettingLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.orderSettings.isNotEmpty) {
                        listData.addAll(state.orderSettings);
                      }
                    });
                  }
                  if (state is OrderSettingLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is OrderSettingSaveSuccess) {
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
                  if (state is OrderSettingSaveFailed) {
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
                  if (state is OrderSettingUpdateSuccess) {
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
                  if (state is OrderSettingUpdateFailed) {
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
                  if (state is OrderSettingDeleteSuccess) {
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
                  if (state is OrderSettingDeleteManySuccess) {
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
                  if (state is OrderSettingGetSuccess) {
                    setState(() {
                      isDataChange = false;
                      screenData = state.orderSettings;

                      if (screenData.settingcode!.isNotEmpty) {
                        context.read<OrderTemplateSettingBloc>().add(
                          OrderTemplateSettingGet(
                            guid: screenData.settingcode!,
                          ),
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
                },
              ),
              BlocListener<PosSettingBloc, PosSettingState>(
                listener: (context, state) {
                  if (state is GetApiKeySuccess) {
                    setState(() {
                      activeAPIKEY = state.token;
                      if (activeAPIKEY.isNotEmpty) {
                        activatePin();
                      } else {
                        setState(() {
                          global.showSnackBar(
                            context,
                            const Icon(
                              Icons.error_outline,
                              color: Colors.white,
                            ),
                            global.language("not_success_get_apikey"),
                            Colors.red,
                          );
                        });
                      }
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
                      updateData(screenData.guidfixed!);

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
              BlocListener<OrderTemplateSettingBloc, OrderTemplateSettingState>(
                listener: (context, state) {
                  if (state is OrderTemplateSettingGetSuccess) {
                    setState(() {
                      screenData.settingname = state.orderTemplateSettings.code;
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
                //     return global.language('this_field_is_required');
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
            if (controller.text.isEmpty) {
              FocusScope.of(context).previousFocus();
            }
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
              border: OutlineInputBorder(),
              counterText: '',
              hintStyle: TextStyle(color: global.theme.textColor, fontSize: 20.0),
            ),
            onChanged: (value) {
              if (value.length == 1) {
                FocusScope.of(context).nextFocus();
              }
            },
          ),
        ),
      ),
    );
  }
}
