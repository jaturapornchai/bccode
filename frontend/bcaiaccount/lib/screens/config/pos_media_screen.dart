import 'dart:io';

import 'package:smlaicloud/bloc/image/image_upload_bloc.dart';
import 'package:smlaicloud/bloc/pos_media/pos_media_bloc.dart';
import 'package:smlaicloud/components/custom_segmented_button.dart';
import 'package:smlaicloud/components/video_player.dart';
import 'package:smlaicloud/model/pos_media_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart' hide SegmentedButton, ButtonSegment;
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:image_picker/image_picker.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';
import 'package:intl/intl.dart';

class PosMediaScreen extends StatefulWidget {
  const PosMediaScreen({super.key});

  @override
  State<PosMediaScreen> createState() => PosMediaScreenState();
}

class PosMediaScreenState extends State<PosMediaScreen>
    with SingleTickerProviderStateMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  List<LanguageModel> languageList = <LanguageModel>[];
  List<PosMediaModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late PosMediaState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  bool isEditMode = false;
  late PosMediaModel screenData;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  bool loadingData = false;
  late DateTime dateNow = DateTime.now();
  final _formKey = GlobalKey<FormState>();
  TextEditingController codeController = TextEditingController();
  List<Uint8List> imageWeb = [];
  List<File> imageFile = [];
  final ImagePicker imagePicker = ImagePicker();
  List<TextEditingController> mediaFromDateController = [];
  List<TextEditingController> mediaToDateController = [];
  List<TextEditingController> mediaFromTimeController = [];
  List<TextEditingController> mediaToTimeController = [];
  List<DayOfWeekModel> dayOfWeekList = [];
  List<List<DayOfWeekModel>> dayOfWeekSeleted = [];
  final List<TimeOfDay> _selectedTime = [];
  List<TextEditingController> uriController = [];
  List<TextEditingController> uriLinkController = [];
  List<TextEditingController> uriVideoController = [];
  List<bool> _isLoadingUploadVideo = [];

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
    dayOfWeekList.add(
      DayOfWeekModel(code: '1', name: global.language('monday')),
    );
    dayOfWeekList.add(
      DayOfWeekModel(code: '2', name: global.language('tuesday')),
    );
    dayOfWeekList.add(
      DayOfWeekModel(code: '3', name: global.language('wendesday')),
    );
    dayOfWeekList.add(
      DayOfWeekModel(code: '4', name: global.language('thursday')),
    );
    dayOfWeekList.add(
      DayOfWeekModel(code: '5', name: global.language('friday')),
    );
    dayOfWeekList.add(
      DayOfWeekModel(code: '6', name: global.language('saturday')),
    );
    dayOfWeekList.add(
      DayOfWeekModel(code: '7', name: global.language('sunday')),
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

    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<PosMediaBloc>().add(
      PosMediaLoadList(
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

    screenData = PosMediaModel(
      guidfixed: '',
      code: '',
      description: [],
      resources: [],
    );
    imageFile = [];
    imageWeb = [];
    isDataChange = false;

    dayOfWeekSeleted = [];
    mediaFromDateController = [];
    mediaToDateController = [];
    mediaFromTimeController = [];
    mediaToTimeController = [];

    uriController = [];
    uriVideoController = [];
    uriLinkController = [];
    _isLoadingUploadVideo = [];

    setState(() {
      loadDataToScreen();
    });
  }

  void loadDataToScreen() {
    codeController.text = screenData.code;

    imageWeb = [];
    imageFile = [];

    for (int i = 0; i < screenData.resources.length; i++) {
      _isLoadingUploadVideo.add(false);
      uriController.add(TextEditingController());
      uriVideoController.add(TextEditingController());
      uriLinkController.add(TextEditingController());
      _selectedTime.add(TimeOfDay.now());
      imageWeb.add(Uint8List(0));
      imageFile.add(File(''));
      mediaFromDateController.add(TextEditingController());
      mediaToDateController.add(TextEditingController());
      mediaFromTimeController.add(TextEditingController());
      mediaToTimeController.add(TextEditingController());

      mediaFromDateController[i].text =
          (screenData.resources[i].fromDate.isNotEmpty)
          ? DateFormat(
              'dd/MM/yyyy',
            ).format(DateTime.parse(screenData.resources[i].fromDate))
          : '';
      mediaToDateController[i].text =
          (screenData.resources[i].toDate.isNotEmpty)
          ? DateFormat(
              'dd/MM/yyyy',
            ).format(DateTime.parse(screenData.resources[i].toDate))
          : '';
      mediaFromTimeController[i].text = screenData.resources[i].fromTime;
      mediaToTimeController[i].text = screenData.resources[i].toTime;
      dayOfWeekSeleted.add(
        screenData.resources[i].daysofweek
            .map(
              (e) => dayOfWeekList.firstWhere(
                (element) => element.code == e.toString(),
              ),
            )
            .toList(),
      );

      if (screenData.resources[i].mediaType == 0) {
        uriController[i].text = screenData.resources[i].uri;
      } else if (screenData.resources[i].mediaType == 1) {
        uriVideoController[i].text = screenData.resources[i].uri;
      } else if (screenData.resources[i].mediaType == 2) {
        uriLinkController[i].text = screenData.resources[i].uri;
      }
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
    context.read<PosMediaBloc>().add(PosMediaGet(guid: guid));
  }

  void switchToEdit(PosMediaModel value) {
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
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language("pos_media")),
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
                            context.read<PosMediaBloc>().add(
                              PosMediaDeleteMany(guid: guidListChecked),
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
                color: Colors.white,
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
                      global.language("pos_media_code"),
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
                      global.language("name"),
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

  Widget listObject(int index, PosMediaModel value, bool showCheckBox) {
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
                value.code,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                global.packName(value.description),
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

    for (int i = 0; i < screenData.resources.length; i++) {
      if (mediaFromDateController[i].text.isNotEmpty) {
        DateTime fromDateUtc = DateTime.parse(screenData.resources[i].fromDate);
        screenData.resources[i].fromDate = fromDateUtc
            .toUtc()
            .toIso8601String();
      }

      if (mediaToDateController[i].text.isNotEmpty) {
        DateTime toDateUtc = DateTime.parse(screenData.resources[i].toDate);
        screenData.resources[i].toDate = toDateUtc.toUtc().toIso8601String();
      }

      screenData.resources[i].daysofweek = [];
      for (var element in dayOfWeekSeleted[i]) {
        screenData.resources[i].daysofweek.add(int.parse(element.code));
      }

      if (screenData.resources[i].mediaType == 0) {
        /// uri upload
        screenData.resources[i].uri = uriController[i].text;
      } else if (screenData.resources[i].mediaType == 1) {
        /// uri video
        screenData.resources[i].uri = uriVideoController[i].text;
      } else if (screenData.resources[i].mediaType == 2) {
        /// uri link video
        screenData.resources[i].uri = uriLinkController[i].text;
      }
    }

    if (selectGuid.trim().isEmpty) {
      context.read<PosMediaBloc>().add(PosMediaSave(posMedia: screenData));
    } else {
      context.read<PosMediaBloc>().add(
        PosMediaUpdate(guid: selectGuid, posMedia: screenData),
      );
    }
  }

  void _selectMediaFromDate(BuildContext context, int mediaIndex) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: DateTime.parse(
        (screenData.resources[mediaIndex].fromDate.isNotEmpty)
            ? screenData.resources[mediaIndex].fromDate.toString()
            : dateNow.toIso8601String(),
      ),
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );

    if (pickedDate != null) {
      setState(() {
        screenData.resources[mediaIndex].fromDate = pickedDate
            .toLocal()
            .toIso8601String();

        mediaFromDateController[mediaIndex].text = DateFormat(
          'dd/MM/yyyy',
        ).format(DateTime.parse(screenData.resources[mediaIndex].fromDate));
      });
    }
  }

  void _selectMediaToDate(BuildContext context, int mediaIndex) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: DateTime.parse(
        (screenData.resources[mediaIndex].toDate.isNotEmpty)
            ? screenData.resources[mediaIndex].toDate.toString()
            : dateNow.toIso8601String(),
      ),
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );

    if (pickedDate != null) {
      setState(() {
        screenData.resources[mediaIndex].toDate = pickedDate
            .toLocal()
            .toIso8601String();

        mediaToDateController[mediaIndex].text = DateFormat(
          'dd/MM/yyyy',
        ).format(DateTime.parse(screenData.resources[mediaIndex].toDate));
      });
    }
  }

  void _selectMediaFromTime(BuildContext context, int mediaIndex) async {
    final TimeOfDay? pickedTime = await showTimePicker(
      context: context,
      initialTime: _selectedTime[mediaIndex],
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

        screenData.resources[mediaIndex].fromTime = formattedTime;
        mediaFromTimeController[mediaIndex].text = formattedTime;
      });
    }
  }

  void _selectMediaToTime(BuildContext context, int mediaIndex) async {
    final TimeOfDay? pickedTime = await showTimePicker(
      context: context,
      initialTime: _selectedTime[mediaIndex],
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
          screenData.resources[mediaIndex].fromTime,
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
          screenData.resources[mediaIndex].toTime = '';
          mediaToTimeController[mediaIndex].text = '';
          return;
        } else {
          // Update the toTime and the text in the text field
          screenData.resources[mediaIndex].toTime = formattedTime;
          mediaToTimeController[mediaIndex].text = formattedTime;
        }
      });
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

  Future<List<DayOfWeekModel>> getDataDayOfWeek(filter) async {
    return dayOfWeekList;
  }

  Future<void> pickImages(int mediaIndex) async {
    String modifiedName = "";
    final XFile? image = await imagePicker.pickImage(
      source: ImageSource.gallery,
    );
    if (image != null) {
      if (kIsWeb) {
        // Code for web platform
        Uint8List bytes = await image.readAsBytes();

        modifiedName = image.name;
        if (modifiedName.startsWith('scaled_')) {
          modifiedName = modifiedName.replaceFirst('scaled_', '');
        }

        imageWeb[mediaIndex] = bytes;
        imageFile[mediaIndex] = File('');
      } else if (Platform.isMacOS) {
        // Code for macOS platform
        Uint8List bytes = await image.readAsBytes();
        modifiedName = image.name.replaceFirst('scaled_', '');

        imageWeb[mediaIndex] = bytes;
        imageFile[mediaIndex] = File(image.path);
      }

      if (mounted) {
        context.read<ImageUploadBloc>().add(
          ImageUploadResposneUri(
            imageFiles: imageFile[mediaIndex],
            imageWeb: imageWeb[mediaIndex],
            imageName: modifiedName,
            index: mediaIndex,
          ),
        );
      }
    }
  }

  /// pickVideo
  Future<void> pickVideo(int mediaIndex) async {
    String modifiedName = "";
    final XFile? video = await imagePicker.pickVideo(
      source: ImageSource.gallery,
    );
    if (video != null) {
      Uint8List bytes = await video.readAsBytes();
      modifiedName = video.name.replaceFirst('scaled_', '');

      imageWeb[mediaIndex] = bytes;
      imageFile[mediaIndex] = File(video.path);

      if (mounted) {
        context.read<ImageUploadBloc>().add(
          VideoUploadResposneUri(
            videoFiles: imageFile[mediaIndex],
            videoWeb: imageWeb[mediaIndex],
            videoName: modifiedName,
            index: mediaIndex,
          ),
        );
      }
    }
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
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
            labelText: global.language("pos_media_code"),
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

    formWidgets.addAll(listNamesFields(screenData.description, "name"));

    formWidgets.add(
      (screenData.resources.isNotEmpty)
          ? Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
              child: Text(global.language("pos_media")),
            )
          : Container(),
    );

    List<Widget> mediaList = [];
    for (
      int mediaIndex = 0;
      mediaIndex < screenData.resources.length;
      mediaIndex++
    ) {
      mediaList.add(
        Padding(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: SizedBox(
            width: double.infinity,
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Column(
                    children: [
                      SizedBox(
                        height: 50,
                        width: double.infinity,
                        child: ElevatedButton.icon(
                          /// style button
                          style: ElevatedButton.styleFrom(
                            foregroundColor: Colors.white,
                            backgroundColor: Colors.red,
                          ),
                          onPressed: () async {
                            screenData.resources.removeAt(mediaIndex);
                            imageWeb.removeAt(mediaIndex);
                            imageFile.removeAt(mediaIndex);

                            mediaFromDateController.removeAt(mediaIndex);
                            mediaToDateController.removeAt(mediaIndex);
                            mediaFromTimeController.removeAt(mediaIndex);
                            mediaToTimeController.removeAt(mediaIndex);
                            dayOfWeekSeleted.removeAt(mediaIndex);
                            uriController.removeAt(mediaIndex);
                            uriVideoController.removeAt(mediaIndex);
                            uriLinkController.removeAt(mediaIndex);
                            _isLoadingUploadVideo.removeAt(mediaIndex);

                            setState(() {});
                          },
                          icon: Icon(Icons.delete),
                          label: Text(global.language("delete")),
                        ),
                      ),
                      const SizedBox(height: 10),
                      Container(
                        padding: const EdgeInsets.only(
                          left: 5,
                          right: 5,
                          bottom: 5,
                          top: 5,
                        ),
                        decoration: BoxDecoration(
                          border: Border.all(color: Colors.grey),
                          borderRadius: const BorderRadius.all(
                            Radius.circular(5.0),
                          ),
                        ),
                        child: Column(
                          children: [
                            SizedBox(
                              width: double.infinity,
                              child: SegmentedButton<int>(
                                style: SegmentedButton.styleFrom(
                                  selectedBackgroundColor: Colors.blue,
                                  selectedForegroundColor: Colors.white,
                                  foregroundColor: Colors.white,
                                  backgroundColor: Colors.grey,
                                ),
                                segments: <ButtonSegment<int>>[
                                  ButtonSegment<int>(
                                    value: 0,
                                    label: Text(
                                      global.language("upload_image"),
                                    ),
                                    icon: const Icon(Icons.image),
                                  ),
                                  ButtonSegment<int>(
                                    value: 1,
                                    label: Text(
                                      global.language("upload_video"),
                                    ),
                                    icon: const Icon(Icons.video_camera_back),
                                  ),
                                  ButtonSegment<int>(
                                    value: 2,
                                    label: Text(global.language("url_video")),
                                    icon: const Icon(Icons.link_rounded),
                                  ),
                                ],
                                selected: {
                                  screenData.resources[mediaIndex].mediaType,
                                },
                                onSelectionChanged: (Set<int> newSelection) {
                                  screenData.resources[mediaIndex].mediaType =
                                      newSelection.first;
                                  _isLoadingUploadVideo[mediaIndex] = false;
                                  setState(() {});
                                },
                              ),
                            ),
                            const SizedBox(height: 10),
                            (screenData.resources[mediaIndex].mediaType == 0)
                                ? Column(
                                    children: [
                                      SizedBox(
                                        width: 300,
                                        height: 344,
                                        child: Stack(
                                          children: [
                                            Center(
                                              child: DecoratedBox(
                                                decoration: BoxDecoration(
                                                  color: Colors.white,
                                                  border: Border.all(
                                                    color: Colors.black,
                                                  ),
                                                  borderRadius:
                                                      BorderRadius.circular(5),
                                                  boxShadow: const [
                                                    BoxShadow(
                                                      offset: Offset(0, 4),
                                                      color:
                                                          Colors.cyan, //edited
                                                      spreadRadius: 4,
                                                      blurRadius: 10, //edited
                                                    ),
                                                  ],
                                                  image:
                                                      (uriController[mediaIndex]
                                                                  .text !=
                                                              '' &&
                                                          screenData
                                                                  .resources[mediaIndex]
                                                                  .mediaType ==
                                                              0)
                                                      ? DecorationImage(
                                                          image: NetworkImage(
                                                            uriController[mediaIndex]
                                                                .text,
                                                          ),
                                                          fit: BoxFit.fill,
                                                        )
                                                      : const DecorationImage(
                                                          image: AssetImage(
                                                            'assets/img/noimage.png',
                                                          ),
                                                        ),
                                                ),
                                                child: const SizedBox(
                                                  width: 300,
                                                  height: 333,
                                                ),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),

                                      /// sized box height 10
                                      SizedBox(height: 10),
                                      SizedBox(
                                        height: 50,
                                        width: double.infinity,
                                        child: ElevatedButton.icon(
                                          /// style button
                                          style: ElevatedButton.styleFrom(
                                            foregroundColor: Colors.white,
                                            backgroundColor: Colors.blue,
                                          ),
                                          onPressed: () async {
                                            pickImages(mediaIndex);
                                          },
                                          icon: Icon(Icons.folder),
                                          label: Text(
                                            global.language("select_image"),
                                          ),
                                        ),
                                      ),
                                    ],
                                  )
                                : (screenData.resources[mediaIndex].mediaType ==
                                      1)
                                ? Column(
                                    children: [
                                      /// loading CircularProgressIndicator upload video
                                      (_isLoadingUploadVideo[mediaIndex])
                                          ? const Center(
                                              child:
                                                  CircularProgressIndicator(),
                                            )
                                          : (uriVideoController[mediaIndex]
                                                .text
                                                .isNotEmpty)
                                          ? Column(
                                              children: [
                                                Row(
                                                  children: [
                                                    Expanded(
                                                      child: Text(
                                                        uriVideoController[mediaIndex]
                                                            .text,
                                                      ),
                                                    ),

                                                    /// show alert dialog see video player width 500 height 500
                                                    IconButton(
                                                      icon: const Icon(
                                                        Icons
                                                            .play_circle_outline,
                                                      ),
                                                      onPressed: () {
                                                        showDialog(
                                                          context: context,
                                                          builder: (BuildContext context) {
                                                            return AlertDialog(
                                                              content: SizedBox(
                                                                width: 900,
                                                                height: 500,
                                                                child: ButterFlyAssetVideo(
                                                                  uri:
                                                                      uriVideoController[mediaIndex]
                                                                          .text,
                                                                ),
                                                              ),
                                                            );
                                                          },
                                                        );
                                                      },
                                                    ),

                                                    /// icon copy uri
                                                    IconButton(
                                                      icon: const Icon(
                                                        Icons.copy,
                                                      ),
                                                      onPressed: () {
                                                        Clipboard.setData(
                                                          ClipboardData(
                                                            text:
                                                                uriVideoController[mediaIndex]
                                                                    .text,
                                                          ),
                                                        );
                                                        global.showSnackBar(
                                                          context,
                                                          const Icon(
                                                            Icons.copy,
                                                            color: Colors.white,
                                                          ),
                                                          global.language(
                                                            "copy_uri_success",
                                                          ),
                                                          Colors.blue,
                                                        );
                                                      },
                                                    ),
                                                  ],
                                                ),
                                              ],
                                            )
                                          : Container(),

                                      SizedBox(height: 10),
                                      SizedBox(
                                        height: 50,
                                        width: double.infinity,
                                        child: ElevatedButton.icon(
                                          /// style button
                                          style: ElevatedButton.styleFrom(
                                            foregroundColor: Colors.white,
                                            backgroundColor: Colors.blue,
                                          ),
                                          onPressed: () async {
                                            pickVideo(mediaIndex);
                                          },
                                          icon: const Icon(
                                            Icons.video_camera_back_rounded,
                                          ),
                                          label: Text(
                                            global.language("select_video"),
                                          ),
                                        ),
                                      ),
                                    ],
                                  )
                                : Container(
                                    /// text fild uri
                                    padding: const EdgeInsets.only(
                                      top: 10,
                                      left: 10,
                                      right: 10,
                                      bottom: 10,
                                    ),
                                    child: TextFormField(
                                      readOnly: !isEditMode,
                                      textAlign: TextAlign.left,
                                      controller: uriLinkController[mediaIndex],
                                      onChanged: (value) {
                                        isDataChange = true;
                                      },
                                      keyboardType: TextInputType.multiline,
                                      maxLines: 10,
                                      decoration: InputDecoration(
                                        floatingLabelBehavior:
                                            FloatingLabelBehavior.always,
                                        border: OutlineInputBorder(),
                                        labelText: global.language("uri"),
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
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 10),

                /// widget form
                Expanded(
                  child: Column(
                    children: [
                      /// description
                      Column(
                        children: listNamesFields(
                          screenData.resources[mediaIndex].description,
                          "description",
                        ),
                      ),

                      /// fromdate todate
                      Padding(
                        padding: const EdgeInsets.only(
                          left: 10,
                          right: 10,
                          bottom: 10,
                        ),
                        child: Row(
                          children: [
                            Expanded(
                              child: TextField(
                                readOnly: true,
                                decoration: InputDecoration(
                                  floatingLabelBehavior:
                                      FloatingLabelBehavior.always,
                                  border: OutlineInputBorder(),
                                  labelText: global.language("from_date"),
                                  suffixIcon: Row(
                                    mainAxisAlignment: MainAxisAlignment
                                        .spaceBetween, // added line
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
                                        icon: const Icon(Icons.calendar_today),
                                        onPressed: () {
                                          _selectMediaFromDate(
                                            context,
                                            mediaIndex,
                                          );
                                        },
                                      ),
                                    ],
                                  ),
                                ),
                                controller: mediaFromDateController[mediaIndex],
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
                                        int year =
                                            int.tryParse(valueSplit[2]) ?? 0;
                                        year = year - 543;
                                        int month =
                                            int.tryParse(valueSplit[1]) ?? 0;
                                        int day =
                                            int.tryParse(valueSplit[0]) ?? 0;
                                        value =
                                            "$year-${month.toString().padLeft(2, '0')}-${day.toString().padLeft(2, '0')}";
                                      }

                                      if (global.isValidDate(value)) {
                                        screenData
                                            .resources[mediaIndex]
                                            .fromDate = DateTime.parse(
                                          value,
                                        ).toLocal().toIso8601String();
                                      }
                                    } catch (e) {
                                      // print(e);
                                    }
                                  });
                                },
                                onSubmitted: (value) => {
                                  mediaFromDateController[mediaIndex].text =
                                      DateFormat('dd/MM/yyyy').format(
                                        DateTime.parse(
                                          screenData
                                              .resources[mediaIndex]
                                              .fromDate
                                              .toString(),
                                        ),
                                      ),
                                },
                              ),
                            ),
                            SizedBox(width: 10),
                            Expanded(
                              child: TextField(
                                readOnly: true,
                                decoration: InputDecoration(
                                  floatingLabelBehavior:
                                      FloatingLabelBehavior.always,
                                  border: OutlineInputBorder(),
                                  labelText: global.language("to_date"),
                                  suffixIcon: Row(
                                    mainAxisAlignment: MainAxisAlignment
                                        .spaceBetween, // added line
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
                                        icon: const Icon(Icons.calendar_today),
                                        onPressed: () {
                                          _selectMediaToDate(
                                            context,
                                            mediaIndex,
                                          );
                                        },
                                      ),
                                    ],
                                  ),
                                ),
                                controller: mediaToDateController[mediaIndex],
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
                                        int year =
                                            int.tryParse(valueSplit[2]) ?? 0;
                                        year = year - 543;
                                        int month =
                                            int.tryParse(valueSplit[1]) ?? 0;
                                        int day =
                                            int.tryParse(valueSplit[0]) ?? 0;
                                        value =
                                            "$year-${month.toString().padLeft(2, '0')}-${day.toString().padLeft(2, '0')}";
                                      }

                                      if (global.isValidDate(value)) {
                                        screenData
                                            .resources[mediaIndex]
                                            .toDate = DateTime.parse(
                                          value,
                                        ).toLocal().toIso8601String();
                                      }
                                    } catch (e) {
                                      // print(e);
                                    }
                                  });
                                },
                                onSubmitted: (value) => {
                                  mediaToDateController[mediaIndex].text =
                                      DateFormat('dd/MM/yyyy').format(
                                        DateTime.parse(
                                          screenData
                                              .resources[mediaIndex]
                                              .toDate
                                              .toString(),
                                        ),
                                      ),
                                },
                              ),
                            ),
                          ],
                        ),
                      ),

                      /// fromtime totime
                      Padding(
                        padding: const EdgeInsets.only(
                          left: 10,
                          right: 10,
                          bottom: 10,
                        ),
                        child: Row(
                          children: [
                            Expanded(
                              child: TextFormField(
                                controller: mediaFromTimeController[mediaIndex],
                                onTap: () async {
                                  _selectMediaFromTime(context, mediaIndex);
                                },
                                readOnly: true,
                                decoration: InputDecoration(
                                  labelText: global.language("from_time"),
                                  border: const OutlineInputBorder(),
                                  floatingLabelBehavior:
                                      FloatingLabelBehavior.always,
                                  suffixIcon: Row(
                                    mainAxisAlignment: MainAxisAlignment
                                        .spaceBetween, // added line
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
                                        icon: const Icon(Icons.timer_sharp),
                                        onPressed: () {
                                          _selectMediaFromTime(
                                            context,
                                            mediaIndex,
                                          );
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
                                controller: mediaToTimeController[mediaIndex],
                                onTap: () async {
                                  _selectMediaToTime(context, mediaIndex);
                                },
                                readOnly: true,
                                decoration: InputDecoration(
                                  labelText: global.language("from_to"),
                                  border: const OutlineInputBorder(),
                                  floatingLabelBehavior:
                                      FloatingLabelBehavior.always,
                                  suffixIcon: Row(
                                    mainAxisAlignment: MainAxisAlignment
                                        .spaceBetween, // added line
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
                                        icon: const Icon(Icons.timer_sharp),
                                        onPressed: () {
                                          _selectMediaToTime(
                                            context,
                                            mediaIndex,
                                          );
                                        },
                                      ),
                                    ],
                                  ),
                                ),
                                inputFormatters: [global.TimeInputFormatter()],
                              ),
                            ),
                          ],
                        ),
                      ),

                      /// displaytime
                      Padding(
                        padding: const EdgeInsets.only(
                          left: 10,
                          right: 10,
                          bottom: 10,
                        ),
                        child: TextField(
                          keyboardType: TextInputType.number,
                          inputFormatters: <TextInputFormatter>[
                            FilteringTextInputFormatter.digitsOnly,
                          ],
                          controller: TextEditingController(
                            text: screenData.resources[mediaIndex].displaytime
                                .toString(),
                          ),
                          onChanged: (value) {
                            isDataChange = true;
                            screenData.resources[mediaIndex].displaytime =
                                int.parse(value);
                          },
                          decoration: InputDecoration(
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                            border: OutlineInputBorder(),
                            labelText: global.language("display_time"),
                          ),
                        ),
                      ),

                      /// dayofweek checkbox
                      /* XXXX
                    Padding(
                      padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
                      child: DropdownSearch<DayOfWeekModel>.multiSelection(
                        enabled: isEditMode,
                        items: (String filter, infiniteScrollProps) => getDataDayOfWeek(filter),
                        compareFn: (item, selectedItem) => item.code == selectedItem.code,
                        itemAsString: (DayOfWeekModel? group) {
                          if (group == null) return '';
                          return '${group.code} - ${group.name}';
                        },
                        decoratorProps: DropDownDecoratorProps(
                          decoration: InputDecoration(
                            border: const OutlineInputBorder(),
                            contentPadding: const EdgeInsets.only(left: 10, top: 15, bottom: 10, right: 10),
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                            labelText: global.language("day_of_week"),
                          ),
                        ),
                        onChanged: (List<DayOfWeekModel> value) {
                          setState(() {
                            dayOfWeekSeleted[mediaIndex] = value;
                          });
                        },
                        popupProps: const PopupPropsMultiSelection.dialog(
                          showSearchBox: false,
                          showSelectedItems: true,
                        ),
                        selectedItems: dayOfWeekSeleted[mediaIndex],
                      ),
                    ),*/
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    formWidgets.add(Column(children: mediaList));

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              screenData.resources.add(
                ResourceModel(
                  daysofweek: [],
                  description: [],
                  displaytime: 5,
                  fromDate: '',
                  fromTime: '',
                  mediaType: 0,
                  toDate: '',
                  toTime: '',
                  uri: '',
                ),
              );

              mediaFromDateController.add(TextEditingController());
              mediaToDateController.add(TextEditingController());
              mediaFromTimeController.add(TextEditingController());
              mediaToTimeController.add(TextEditingController());
              _selectedTime.add(TimeOfDay.now());
              uriController.add(TextEditingController());
              uriVideoController.add(TextEditingController());
              uriLinkController.add(TextEditingController());
              _isLoadingUploadVideo.add(false);

              dayOfWeekSeleted.add([
                dayOfWeekList[0],
                dayOfWeekList[1],
                dayOfWeekList[2],
                dayOfWeekList[3],
                dayOfWeekList[4],
                dayOfWeekList[5],
                dayOfWeekList[6],
              ]);

              setState(() {
                imageWeb.add(Uint8List(0));
                imageFile.add(File(''));
                FocusScope.of(context).unfocus();
              });
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_media")),
          ),
        ),
      );

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
        title: Text(headerEdit + global.language("pos_media")),
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
                            context.read<PosMediaBloc>().add(
                              PosMediaDelete(guid: selectGuid),
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
              BlocListener<PosMediaBloc, PosMediaState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is PosMediaLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.posMedias.isNotEmpty) {
                        listData.addAll(state.posMedias);
                      }
                    });
                  }
                  if (state is PosMediaLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is PosMediaSaveSuccess) {
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
                  if (state is PosMediaSaveFailed) {
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
                  if (state is PosMediaUpdateSuccess) {
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
                  if (state is PosMediaUpdateFailed) {
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
                  if (state is PosMediaDeleteSuccess) {
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
                  if (state is PosMediaDeleteManySuccess) {
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
                  if (state is PosMediaGetSuccess) {
                    setState(() {
                      isDataChange = false;
                      screenData = state.posMedia;

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
              BlocListener<ImageUploadBloc, ImageUploadState>(
                listener: (context, state) {
                  if (state is ImageUploadResposneUriSaveSuccess) {
                    setState(() {
                      if (screenData.resources[state.index].mediaType == 0) {
                        uriController[state.index].text = state.uri;
                      }
                    });
                  } else if (state is ImageUploadResposneUriSaveFailure) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        "${global.language("not_success_save")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                  if (state is VideoUploadResposneUriInProgress) {
                    setState(() {
                      _isLoadingUploadVideo[state.index] = true;
                    });
                  } else if (state is VideoUploadResposneUriSaveSuccess) {
                    setState(() {
                      if (screenData.resources[state.index].mediaType == 1) {
                        uriVideoController[state.index].text = state.uri;
                        _isLoadingUploadVideo[state.index] = false;
                      }
                    });
                  } else if (state is VideoUploadResposneUriSaveFailure) {
                    _isLoadingUploadVideo[state.index] = false;
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        "${global.language("not_success_save")} : ${state.message}",
                        Colors.red,
                      );
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
            padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
            child: TextFormField(
              readOnly: !isEditMode,
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
                if (languageIndex == 0 && fieldname == "name") {
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
            decoration: const InputDecoration(
              border: OutlineInputBorder(),
              counterText: '',
              hintStyle: TextStyle(color: Colors.black, fontSize: 20.0),
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
