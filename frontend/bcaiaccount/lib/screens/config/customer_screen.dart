import 'dart:async';
import 'dart:io';

import 'package:smlaicloud/bloc/customer/customer_bloc.dart';
import 'package:smlaicloud/bloc/customer_group/customer_group_bloc.dart';
import 'package:smlaicloud/model/customer_address_model.dart';
import 'package:smlaicloud/model/customer_group_model.dart';
import 'package:smlaicloud/model/customer_model.dart';
import 'package:smlaicloud/model/customer_request_model.dart';
import 'package:dropdown_search/dropdown_search.dart';
import 'package:email_validator/email_validator.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class CustomerScreen extends StatefulWidget {
  const CustomerScreen({super.key});

  @override
  State<CustomerScreen> createState() => CustomerScreenState();
}

class CustomerScreenState extends State<CustomerScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
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
  List<CustomerModel> listData = [];
  List<CustomerGroupModel> listDataGroup = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late CustomerState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late CustomerModel screenData;
  late CustomerRequestModel screenDataRequest;
  List<Uint8List> imageWeb = [];
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;
  Color colorSelected = Colors.white;
  final _debouncer = global.Debouncer(1000);
  late Timer screenTimer;
  bool loadingData = false;
  List<CustomerGroupModel> groupSelected = [];
  List<File> imageFile = [];
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final _popupBuilderKey = GlobalKey<DropdownSearchState<String>>();
  final _popupBuilderKey2 = GlobalKey<DropdownSearchState<String>>();

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
    loadDataGroupList();
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

    // แก้ไข: ใช้ milliseconds แทน microseconds (เปลี่ยนจาก 2000 ครั้ง/วินาที → 2 ครั้ง/วินาที)
    screenTimer = Timer.periodic(const Duration(milliseconds: 500), (timer) {
      if (refreshFocus) {
        fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        refreshFocus = false;
      }
    });
    super.initState();
  }

  @override
  void dispose() {
    screenTimer.cancel(); // ✅ เพิ่ม: cancel timer ก่อน dispose
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

  void loadDataGroupList() {
    context.read<CustomerGroupBloc>().add(
      const CustomerGroupLoadList(offset: 0, limit: 1000, search: ""),
    );
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<CustomerBloc>().add(
      CustomerLoadList(
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
    List<LanguageDataModel> addressNames = [];
    for (int k = 0; k < languageList.length; k++) {
      names.add(LanguageDataModel(code: languageList[k].code!, name: ""));
      addressNames.add(
        LanguageDataModel(code: languageList[k].code!, name: ""),
      );
    }
    List<LanguageDataModel> nameBill = [];
    for (int k = 0; k < languageList.length; k++) {
      nameBill.add(LanguageDataModel(code: languageList[k].code!, name: ""));
    }
    screenData = CustomerModel(
      guidfixed: "",
      code: "",
      names: names,
      customertype: 1,
      branchnumber: '00000',
      personaltype: 1,
      addressforbilling: CustomerAddressModel(
        guid: "",
        address: ["", "", ""],
        countrycode: "",
        provincecode: "",
        districtcode: "",
        subdistrictcode: "",
        zipcode: "",
        latitude: 0,
        longitude: 0,
        contactnames: nameBill,
        phoneprimary: "",
        phonesecondary: "",
      ),
      addressforshipping: [],
      images: [],
      groups: [],
      taxid: "",
      email: "",
      iscreditor: false,
      isdebtor: false,
    );

    screenDataRequest = CustomerRequestModel(
      guidfixed: "",
      code: "",
      names: names,
      customertype: 1,
      branchnumber: '00000',
      personaltype: 1,
      addressforbilling: CustomerAddressModel(
        guid: "",
        address: ["", "", ""],
        countrycode: "",
        provincecode: "",
        districtcode: "",
        subdistrictcode: "",
        zipcode: "",
        latitude: 0,
        longitude: 0,
        contactnames: nameBill,
        phoneprimary: "",
        phonesecondary: "",
      ),
      addressforshipping: [],
      images: [],
      groups: [],
      taxid: "",
      email: "",
      iscreditor: false,
      isdebtor: false,
    );
    groupSelected = [];
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
    context.read<CustomerBloc>().add(CustomerGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('customer')),
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
                            backgroundColor: global.theme.primaryColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<CustomerBloc>().add(
                              CustomerDeleteMany(guid: guidListChecked),
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
                      global.language("customer_code"),
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
                      global.language("customer_name"),
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

  void switchToEdit(CustomerModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
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

  Widget listObject(int index, CustomerModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน CustomerLoadSuccess แล้ว
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
                //searchFocusNode.requestFocus();
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
              flex: 10,
              child: Text(
                global.packName(value.names),
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

  void saveOrUpdateData() {
    showCheckBox = false;

    for (var ele in groupSelected) {
      screenDataRequest.groups.add(ele.guidfixed);
    }
    screenData.groups = groupSelected;

    if (selectGuid.trim().isEmpty) {
      if (imageFile.isNotEmpty) {
        context.read<CustomerBloc>().add(
          CustomerWithImageSave(
            customer: screenData,
            imageFile: imageFile,
            imageWeb: imageWeb,
          ),
        );
      } else {
        context.read<CustomerBloc>().add(CustomerSave(customer: screenData));
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
      // print(screenData.images.length);
      if (imageWeb[i].isNotEmpty || screenData.images[i].uri != '') {
        imageFileUpdate.add(imageFile[i]);
        imageWebUpdate.add(imageWeb[i]);
        imageUris.add(ImagesModel(uri: screenData.images[i].uri, xorder: i));
      }
    }
    // print("imageWebUpdate.isNotEmpty " + imageWebUpdate.isNotEmpty.toString());
    if (imageWebUpdate.isNotEmpty) {
      context.read<CustomerBloc>().add(
        CustomerWithImageUpdate(
          guid: guid,
          customer: screenData,
          imageFiles: imageFile,
          imagesUris: imageUris,
          imageWeb: imageWeb,
        ),
      );
    } else {
      screenData.images = [];
      context.read<CustomerBloc>().add(
        CustomerUpdate(guid: guid, customerModel: screenData),
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

  void getDataToEditScreen(CustomerModel customer) {
    isDataChange = false;
    selectGuid = customer.guidfixed;

    screenData.code = customer.code;
    screenData.email = customer.email;
    screenData.fundcode = customer.fundcode;
    screenData.creditday = customer.creditday;
    screenData.guidfixed = customer.guidfixed;
    screenData.images = customer.images;

    screenData.personaltype = customer.personaltype;
    screenData.customertype = customer.customertype;
    screenData.branchnumber = customer.branchnumber;
    screenData.taxid = customer.taxid;
    screenData.groups = customer.groups;
    screenData.iscreditor = customer.iscreditor;
    screenData.isdebtor = customer.isdebtor;
    groupSelected = customer.groups;

    screenData.names = [];
    screenData.addressforbilling.contactnames = [];
    screenData.addressforshipping = [];
    //เพิ่มภาษาตาม config
    for (var lang in languageList) {
      screenData.names.add(LanguageDataModel(code: lang.code!, name: ""));
      screenData.addressforbilling.contactnames!.add(
        LanguageDataModel(code: lang.code!, name: ""),
      );
    }

    for (var addressforshipping in customer.addressforshipping) {
      CustomerAddressModel newData = CustomerAddressModel(
        guid: addressforshipping.guid,
        address: addressforshipping.address,
        countrycode: addressforshipping.countrycode,
        provincecode: addressforshipping.provincecode,
        districtcode: addressforshipping.districtcode,
        subdistrictcode: addressforshipping.subdistrictcode,
        zipcode: addressforshipping.zipcode,
        latitude: addressforshipping.latitude,
        longitude: addressforshipping.longitude,
        contactnames: [],
        phoneprimary: addressforshipping.phoneprimary,
        phonesecondary: addressforshipping.phonesecondary,
      );

      for (var lang in languageList) {
        newData.contactnames!.add(
          LanguageDataModel(code: lang.code!, name: ""),
        );
      }
      screenData.addressforshipping.add(newData);
    }

    //ใส่ value ตามภาษา
    for (var data in customer.addressforbilling.contactnames!) {
      for (var ele in screenData.addressforbilling.contactnames!) {
        if (data.code == ele.code) {
          ele.name = data.name;
        }
      }
    }

    for (var data in customer.names) {
      for (var ele in screenData.names) {
        if (data.code == ele.code) {
          ele.name = data.name;
        }
      }
    }

    // for (var addressforshipping in customer.addressforshipping) {
    //   for (var data in addressforshipping.contactnames) {
    //     for (var arr in screenData.addressforshipping) {
    //       for (var ele in arr.contactnames) {
    //         if (data.code == ele.code) {
    //           ele.name = data.name;
    //         }
    //       }
    //     }
    //   }
    // }

    for (int i = 0; i < customer.addressforshipping.length; i++) {
      for (var data in customer.addressforshipping[i].contactnames!) {
        for (var ele in screenData.addressforshipping[i].contactnames!) {
          if (data.code == ele.code) {
            ele.name = data.name;
          }
        }
      }
    }

    //เก็บค่าที่ไม่ได้เปิดใช้งานภาษาเข้าทาง array
    for (var defaultValueLang in customer.names) {
      LanguageDataModel result = screenData.names.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        screenData.names.add(defaultValueLang);
      }
    }

    for (var defaultValueLang in customer.addressforbilling.contactnames!) {
      LanguageDataModel result = screenData.addressforbilling.contactnames!
          .firstWhere(
            (data) => data.code == defaultValueLang.code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
      if (result.code == '') {
        screenData.addressforbilling.contactnames!.add(defaultValueLang);
      }
    }

    for (var addressforshipping in customer.addressforshipping) {
      for (var defaultValueLang in addressforshipping.contactnames!) {
        for (var arr in screenData.addressforshipping) {
          LanguageDataModel result = arr.contactnames!.firstWhere(
            (data) => data.code == defaultValueLang.code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
          if (result.code == '') {
            arr.contactnames!.add(defaultValueLang);
          }
        }
      }
    }

    imageWeb = [];
    imageFile = [];

    for (int i = 0; i < customer.images.length; i++) {
      imageWeb.add(Uint8List(0));
      imageFile.add(File(''));
    }
  }

  Future<List<CustomerGroupModel>> getDataGroup(filter) async {
    return listDataGroup;
  }

  CustomerGroupModel getGroup(String guid) {
    CustomerGroupModel data = CustomerGroupModel(guidfixed: "", groupcode: "");
    for (var ele in listDataGroup) {
      if (ele.guidfixed == guid) {
        data = ele;
      }
    }
    return data;
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    focusNodeMax = 0;
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextField(
          readOnly: !isEditMode,
          onSubmitted: (value) {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.code),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.code = value.toUpperCase();
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("customer_code"),
          ),
        ),
      ),
    );
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel customerName = screenData.names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (customerName.code == '') {
        screenData.names.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: TextField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.names[languageIndex].name = value;
            },
            onSubmitted: (value) {
              if (kIsWeb) {
                findFocusNext(focusNodeIndex);
              }
            },
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData.names[languageIndex].name,
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("customer_name")} (${getLangName(screenData.names[languageIndex].code)})",
            ),
          ),
        ),
      );
    }
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 1,
              groupValue: screenData.personaltype,
              onChanged: (value) {
                setState(() {
                  screenData.personaltype = value!;
                  focusNodeIndex = 3;
                  refreshFocus = true;
                });
              },
            ),
            Text(global.language("customer_individual")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 2,
              groupValue: screenData.personaltype,
              onChanged: (value) {
                setState(() {
                  screenData.personaltype = value!;
                  focusNodeIndex = 3;
                  refreshFocus = true;
                });
              },
            ),
            Text(global.language("customer_company")),
          ],
        ),
      ),
    );
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.taxid),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.taxid = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language(
              (screenData.personaltype == 1)
                  ? "customer_tax_id_card_number"
                  : "customer_tax_id_bussiness",
            ),
          ),
        ),
      ),
    );
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 1,
              groupValue: screenData.customertype,
              onChanged: (value) {
                setState(() {
                  screenData.customertype = value!;
                  screenData.branchnumber = '0000';
                  focusNodeIndex = 5;
                  refreshFocus = true;
                });
              },
            ),
            Text(global.language("head_office")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 2,
              groupValue: screenData.customertype,
              onChanged: (value) {
                setState(() {
                  screenData.customertype = value!;
                  screenData.branchnumber = '';
                  focusNodeIndex = 4;
                  refreshFocus = true;
                });
              },
            ),
            Text(global.language("branch")),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode || screenData.customertype == 1,
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          key: _popupBuilderKey2,
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.branchnumber),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.branchnumber = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("branch_number"),
          ),
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            bool isValid = (screenData.email!.isEmpty)
                ? true
                : EmailValidator.validate(screenData.email!);
            if (!isValid) {
              global.showSnackBar(
                context,
                Icon(Icons.delete, color: Colors.white),
                global.language("email_error"),
                Colors.red,
              );
            }
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.email),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.email = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("customer_email"),
          ),
        ),
      ),
    );

    // formWidgets.add(Padding(
    // padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
    //     child: TextFormField(
    //         readOnly: !isEditMode,
    //         onFieldSubmitted: (value) {
    //           findFocusNext(focusNodeIndex);
    //         },
    //         focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
    //         controller: TextEditingController(text: screenData.fundcode),
    //         onChanged: (value) {
    //           isDataChange = true;
    //           screenData.fundcode = value;
    //         },
    //         decoration: InputDecoration(
    //           floatingLabelBehavior: FloatingLabelBehavior.always,
    // border: OutlineInputBorder(),
    //           labelText: global.language("fund_code"),
    //         ))));

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          keyboardType: TextInputType.number,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.digitsOnly,
          ],
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          controller: TextEditingController(
            text: screenData.creditday.toString(),
          ),
          onChanged: (value) {
            isDataChange = true;
            screenData.creditday = int.parse(value);
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("credit_day"),
          ),
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: Row(
          // mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            Row(
              children: [
                Checkbox(
                  value: screenData.iscreditor,
                  onChanged: (bool? newValue) {
                    setState(() {
                      screenData.iscreditor = newValue!;
                    });
                  },
                ),
                Text(global.language("supplier")),
              ],
            ),
            SizedBox(width: 10),
            Row(
              children: [
                Checkbox(
                  value: screenData.isdebtor,
                  onChanged: (bool? newValue) {
                    setState(() {
                      screenData.isdebtor = newValue!;
                    });
                  },
                ),
                Text(global.language("customer")),
              ],
            ),
          ],
        ),
      ),
    );

    /* XXXX
    formWidgets.add(
      (!kIsWeb)
          ? Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
              child: DropdownSearch<CustomerGroupModel>.multiSelection(
                enabled: isEditMode,
                key: _popupBuilderKey,
                items: (String filter, infiniteScrollProps) => getDataGroup(filter),
                compareFn: (item, selectedItem) => item.guidfixed == selectedItem.guidfixed,
                itemAsString: (CustomerGroupModel p) => p.groupcode,
                decoratorProps: DropDownDecoratorProps(
                  decoration: InputDecoration(
                    border: const OutlineInputBorder(),
                    contentPadding: const EdgeInsets.only(left: 10, top: 0, bottom: 0, right: 10),
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    labelText: global.language("group"),
                  ),
                ),
                onChanged: (List<CustomerGroupModel> value) {
                  setState(() {
                    groupSelected = value;
                  });
                },
                popupProps: const PopupPropsMultiSelection.modalBottomSheet(
                  showSearchBox: true,
                  showSelectedItems: true,
                ),
                selectedItems: groupSelected,
              ),
            )
          : DropdownSearch<CustomerGroupModel>.multiSelection(
              enabled: isEditMode,
              compareFn: (item, selectedItem) => item.guidfixed == selectedItem.guidfixed,
              items: (String filter, infiniteScrollProps) => getDataGroup(filter),
              itemAsString: (CustomerGroupModel p) => p.groupcode,
              decoratorProps: DropDownDecoratorProps(
                decoration: InputDecoration(
                  border: const OutlineInputBorder(),
                  contentPadding: const EdgeInsets.only(left: 10, top: 0, bottom: 0, right: 10),
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  labelText: global.language("group"),
                ),
              ),
              onChanged: (List<CustomerGroupModel> value) {
                setState(() {
                  groupSelected = value;
                });
              },
              popupProps: const PopupPropsMultiSelection.dialog(
                showSearchBox: true,
                showSelectedItems: true,
              ),
              selectedItems: groupSelected,
            ),
    );*/

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: Text(global.language("address_for_billing")),
      ),
    );

    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel contactnames = screenData
          .addressforbilling
          .contactnames!
          .firstWhere(
            (element) => element.code == languageList[languageIndex].code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
      if (contactnames.code == '') {
        screenData.addressforbilling.contactnames!.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: TextField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.addressforbilling.contactnames![languageIndex].name =
                  value;
            },
            onSubmitted: (value) {
              if (kIsWeb) {
                findFocusNext(focusNodeIndex);
              }
            },
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData
                  .addressforbilling
                  .contactnames![languageIndex]
                  .name,
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("customer_name")} (${getLangName(screenData.addressforbilling.contactnames![languageIndex].code)})",
            ),
          ),
        ),
      );
    }

    for (int addressIndex = 0; addressIndex < 3; addressIndex++) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: TextFormField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.addressforbilling.address![addressIndex] = value;
            },
            onFieldSubmitted: (value) {
              findFocusNext(focusNodeIndex);
            },
            textInputAction: TextInputAction.next,
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData.addressforbilling.address![addressIndex],
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("tax_address_${addressIndex + 1}"),
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onChanged: (value) {
            isDataChange = true;
            screenData.addressforbilling.phoneprimary = value;
          },
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.addressforbilling.phoneprimary,
          ),
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("telephone_primary"),
          ),
        ),
      ),
    );
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onChanged: (value) {
            isDataChange = true;
            screenData.addressforbilling.phonesecondary = value;
          },
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.addressforbilling.phonesecondary,
          ),
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("telephone_secondary"),
          ),
        ),
      ),
    );

    for (
      int addressShipIndex = 0;
      addressShipIndex < screenData.addressforshipping.length;
      addressShipIndex++
    ) {
      formWidgets.add(
        (screenData.addressforshipping.isNotEmpty)
            ? Padding(
                padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
                child: Text(
                  global.language(
                    "address_for_shipping ${addressShipIndex + 1}",
                  ),
                ),
              )
            : Container(),
      );
      for (
        int languageIndex = 0;
        languageIndex < languageList.length;
        languageIndex++
      ) {
        LanguageDataModel contactnames = screenData
            .addressforshipping[addressShipIndex]
            .contactnames!
            .firstWhere(
              (element) => element.code == languageList[languageIndex].code,
              orElse: () => LanguageDataModel(code: '', name: ''),
            );
        if (contactnames.code == '') {
          screenData.addressforshipping[addressShipIndex].contactnames!.add(
            LanguageDataModel(
              code: languageList[languageIndex].code!,
              name: '',
            ),
          );
        }
        formWidgets.add(
          Padding(
            padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
            child: TextField(
              readOnly: !isEditMode,
              onChanged: (value) {
                isDataChange = true;
                screenData
                        .addressforshipping[addressShipIndex]
                        .contactnames![languageIndex]
                        .name =
                    value;
              },
              onSubmitted: (value) {
                if (kIsWeb) {
                  findFocusNext(focusNodeIndex);
                }
              },
              focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
              textAlign: TextAlign.left,
              controller: TextEditingController(
                text: screenData
                    .addressforshipping[addressShipIndex]
                    .contactnames![languageIndex]
                    .name,
              ),
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: const OutlineInputBorder(),
                labelText:
                    "${global.language("customer_name")} (${getLangName(screenData.addressforshipping[addressShipIndex].contactnames![languageIndex].code)})",
              ),
            ),
          ),
        );
      }
      for (
        int addressIndex = 0;
        addressIndex <
            screenData.addressforshipping[addressShipIndex].address!.length;
        addressIndex++
      ) {
        formWidgets.add(
          Padding(
            padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
            child: TextFormField(
              readOnly: !isEditMode,
              onChanged: (value) {
                isDataChange = true;
                screenData
                        .addressforshipping[addressShipIndex]
                        .address![addressIndex] =
                    value;
              },
              onFieldSubmitted: (value) {
                findFocusNext(focusNodeIndex);
              },
              textInputAction: TextInputAction.next,
              focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
              textAlign: TextAlign.left,
              controller: TextEditingController(
                text: screenData
                    .addressforshipping[addressShipIndex]
                    .address![addressIndex],
              ),
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: OutlineInputBorder(),
                labelText: global.language(
                  "shipping_address_${addressIndex + 1}",
                ),
              ),
            ),
          ),
        );
      }
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: TextFormField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.addressforshipping[addressShipIndex].phoneprimary =
                  value;
            },
            onFieldSubmitted: (value) {
              findFocusNext(focusNodeIndex);
            },
            textInputAction: TextInputAction.next,
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text:
                  screenData.addressforshipping[addressShipIndex].phoneprimary,
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("telephone_primary"),
            ),
          ),
        ),
      );
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: TextFormField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.addressforshipping[addressShipIndex].phonesecondary =
                  value;
            },
            onFieldSubmitted: (value) {
              findFocusNext(focusNodeIndex);
            },
            textInputAction: TextInputAction.next,
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData
                  .addressforshipping[addressShipIndex]
                  .phonesecondary,
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("telephone_secondary"),
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      (screenData.images.isNotEmpty)
          ? Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
              child: Text(global.language("customer_image")),
            )
          : Container(),
    );

    List<Widget> imageList = [];
    for (
      int imageIndex = 0;
      imageIndex < screenData.images.length;
      imageIndex++
    ) {
      imageList.add(
        Container(
          width: 300,
          padding: EdgeInsets.only(left: 5, right: 5, bottom: 5, top: 5),
          decoration: BoxDecoration(
            border: Border.all(color: global.theme.textSecondaryColor),
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
                              screenData.images.removeAt(imageIndex);
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
                          color: global.theme.cardColor,
                          border: Border.all(color: global.theme.textColor),
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
                              : (screenData.images[imageIndex].uri != '')
                              ? DecorationImage(
                                  image: NetworkImage(
                                    screenData.images[imageIndex].uri,
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
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              setState(() {
                screenData.images.add(ImagesModel(uri: '', xorder: 0));
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
    formWidgets.add(const SizedBox(height: 5));
    if (isEditMode) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              setState(() {
                List<LanguageDataModel> names = [];
                for (int k = 0; k < languageList.length; k++) {
                  names.add(
                    LanguageDataModel(code: languageList[k].code!, name: ""),
                  );
                }
                screenData.addressforshipping.add(
                  CustomerAddressModel(
                    guid: "",
                    address: ["", "", ""],
                    countrycode: "",
                    provincecode: "",
                    districtcode: "",
                    subdistrictcode: "",
                    zipcode: "",
                    latitude: 0,
                    longitude: 0,
                    phoneprimary: "",
                    phonesecondary: "",
                    contactnames: names,
                  ),
                );
              });
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_address_shipping")),
          ),
        ),
      );
    }

    formWidgets.add(const SizedBox(height: 5));

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              saveOrUpdateData();
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
        title: Text(headerEdit + global.language("customer")),
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
                            context.read<CustomerBloc>().add(
                              CustomerDelete(guid: selectGuid),
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
                onPressed: () => saveOrUpdateData(),
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
              saveOrUpdateData();
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
                      padding: const EdgeInsets.only(top: 10, bottom: 10),
                      child: Form(child: Column(children: formWidgets)),
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

    // ❌ ห้าม clear listKeys ใน build() - จะทำให้เกิด infinite loop!
    // listKeys จะถูก clear ใน BlocListener เมื่อมีการเปลี่ยนแปลง data

    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<CustomerBloc, CustomerState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is CustomerLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.customers.isNotEmpty) {
                        listData.addAll(state.customers);
                        // Clear และ recreate keys เมื่อมีการเพิ่ม data
                        listKeys.clear();
                        // สร้าง keys ใหม่ให้ตรงกับจำนวน items
                        for (int i = 0; i < listData.length; i++) {
                          listKeys.add(GlobalKey());
                        }
                      }
                      // Clear checked list เมื่อไม่ได้อยู่ในโหมด checkbox
                      if (showCheckBox == false) {
                        guidListChecked.clear();
                      }
                    });
                  }
                  if (state is CustomerLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is CustomerSaveSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: Colors.white),
                      global.language("save_success"),
                      Colors.blue,
                    );
                    setState(() {
                      clearEditData();
                      listData.clear();
                      listKeys.clear();
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    loadDataList(searchText);
                  }
                  if (state is CustomerSaveFailed) {
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
                  if (state is CustomerUpdateSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: Colors.white),
                      global.language("edit_success"),
                      Colors.blue,
                    );
                    setState(() {
                      clearEditData();
                      listData.clear();
                      listKeys.clear();
                      isSaveAllow = false;
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(0);
                    });
                    loadDataList(searchText);
                    getData(selectGuid);
                  }
                  if (state is CustomerUpdateFailed) {
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
                  if (state is CustomerDeleteSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.delete, color: Colors.white),
                      global.language("delete_success"),
                      Colors.blue,
                    );
                    setState(() {
                      listData.clear();
                      listKeys.clear();
                      clearEditData();
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(0);
                    });
                    loadDataList(searchText);
                  }
                  // Delete Many
                  if (state is CustomerDeleteManySuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.delete, color: Colors.white),
                      global.language("not_delete_success"),
                      Colors.blue,
                    );
                    setState(() {
                      listData.clear();
                      listKeys.clear();
                      clearEditData();
                      showCheckBox = false;
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    loadDataList(searchText);
                  }
                  // Get
                  if (state is CustomerGetSuccess) {
                    setState(() {
                      getDataToEditScreen(state.customer);
                    });
                    // ย้าย side effects ออกมาจาก setState
                    if (isEditMode) {
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(1);
                        // เรียก findFocusNext หลังจาก animate เสร็จ
                        findFocusNext(0);
                      });
                    }
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
                          // ย้าย animateTo ออกมาจาก setState
                          listScrollController.animateTo(
                            listScrollController.offset -
                                (boxHeader.size.height + box.size.height),
                            duration: const Duration(milliseconds: 100),
                            curve: Curves.ease,
                          );
                          setState(() {
                            isKeyUp = false;
                          });
                        }
                        // Scroll Down
                        if (isKeyDown &&
                            position.dy > (queryData.size.height - 100)) {
                          // ย้าย animateTo ออกมาจาก setState
                          listScrollController.animateTo(
                            listScrollController.offset +
                                (position.dy - (queryData.size.height - 100)),
                            duration: const Duration(milliseconds: 100),
                            curve: Curves.easeOut,
                          );
                          setState(() {
                            isKeyDown = false;
                          });
                        }
                      }
                    }
                  }
                },
              ),
              BlocListener<CustomerGroupBloc, CustomerGroupState>(
                listener: (context, state) {
                  // Load
                  if (state is CustomerGroupLoadSuccess) {
                    setState(() {
                      if (state.customerGroups.isNotEmpty) {
                        listDataGroup = state.customerGroups;
                      }
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
}
