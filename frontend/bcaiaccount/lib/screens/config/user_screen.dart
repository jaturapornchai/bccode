import 'dart:convert';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:dio/dio.dart';
import 'package:smlaicloud/api/client.dart';
import 'package:smlaicloud/bloc/user/user_bloc.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/user_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';
import 'package:smlaicloud/widgets/line_link_qr_dialog.dart';

class UserScreen extends StatefulWidget {
  const UserScreen({super.key});

  @override
  State<UserScreen> createState() => UserScreenState();
}

class UserScreenState extends State<UserScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<global.FieldFocusModel> fieldFocusNodes = [];
  List<UserModel> listData = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String username = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late UserState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  int _hoverIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool isEditMode = false;
  late UserModel screenData;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  bool loadingData = false;
  TextEditingController userCode = TextEditingController();

  final _formKey = GlobalKey<FormState>();
  TextEditingController usernameController = TextEditingController();
  int userRole = 0;

  // LINE OA data (จาก mainapi)
  String? lineUserId;
  String? lineDisplayName;
  String? linePictureUrl;

  // === ฟิลด์สำหรับระบบอนุมัติ ===
  TextEditingController positionController = TextEditingController();
  TextEditingController departmentController = TextEditingController();

  // ข้อมูลการอนุมัติ - ใบสั่งซื้อ (Purchase Order)
  int poApprovalRole = 0; // 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ
  TextEditingController poMaxApprovalAmountController = TextEditingController();

  // ข้อมูลการอนุมัติ - ใบเสนอราคา (Quotation)
  int quotationApprovalRole = 0;
  TextEditingController quotationMaxApprovalAmountController = TextEditingController();

  // สำหรับ backward compatibility
  int approvalRole = 0;
  TextEditingController maxApprovalAmountController = TextEditingController();

  // รายการตำแหน่งงานและแผนกที่ใช้บ่อย
  final List<String> commonPositions = [
    global.language('employee'),
    global.language('dept_head'),
    global.language('department_manager'),
    global.language('director'),
    global.language('vice_managing_director'),
    global.language('managing_director'),
  ];

  final List<String> commonDepartments = [
    global.language('general'),
    global.language('accounting'),
    global.language('finance'),
    global.language('purchasing'),
    'IT',
    global.language('marketing'),
    global.language('transaction_sale'),
    global.language('warehouse'),
    global.language('production'),
    global.language('human_resources'),
  ];

  @override
  void initState() {
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    clearEditData();
    loadDataList("");
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });
    listScrollController.addListener(onScrollList);

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    usernameController.dispose();
    positionController.dispose();
    departmentController.dispose();
    maxApprovalAmountController.dispose();
    poMaxApprovalAmountController.dispose();
    quotationMaxApprovalAmountController.dispose();
    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<UserBloc>().add(
      UserLoadList(
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
    screenData = UserModel(role: 0, shopid: '', username: '', editusername: '');
    usernameController.text = "";
    userRole = 0;
    // Clear LINE data
    lineUserId = null;
    lineDisplayName = null;
    linePictureUrl = null;
    // Clear approval data
    positionController.text = "";
    departmentController.text = "";
    approvalRole = 0;
    maxApprovalAmountController.text = "";
    // Clear document-type approval data
    poApprovalRole = 0;
    poMaxApprovalAmountController.text = "";
    quotationApprovalRole = 0;
    quotationMaxApprovalAmountController.text = "";

    isDataChange = false;
    setState(() {});
  }

  /// โหลดข้อมูลการอนุมัติจาก screenData (ข้อมูลมาจาก mainapi แล้ว)
  void loadApprovalDataFromScreenData() {
    setState(() {
      positionController.text = screenData.position ?? '';
      departmentController.text = screenData.department ?? '';
      approvalRole = screenData.approvalRole ?? 0;
      final maxAmount = screenData.maxApprovalAmount ?? 0;
      maxApprovalAmountController.text = maxAmount > 0 ? maxAmount.toString() : '';

      // โหลดข้อมูลแยกตามประเภทเอกสาร
      if (screenData.poApproval != null) {
        poApprovalRole = screenData.poApproval!.approvalRole;
        final poMaxAmount = screenData.poApproval!.maxApprovalAmount;
        poMaxApprovalAmountController.text = poMaxAmount > 0 ? poMaxAmount.toString() : '';
      } else {
        poApprovalRole = approvalRole;
        poMaxApprovalAmountController.text = maxApprovalAmountController.text;
      }

      if (screenData.quotationApproval != null) {
        quotationApprovalRole = screenData.quotationApproval!.approvalRole;
        final quotationMaxAmount = screenData.quotationApproval!.maxApprovalAmount;
        quotationMaxApprovalAmountController.text = quotationMaxAmount > 0 ? quotationMaxAmount.toString() : '';
      } else {
        quotationApprovalRole = approvalRole;
        quotationMaxApprovalAmountController.text = maxApprovalAmountController.text;
      }

      // โหลดข้อมูล LINE จาก screenData
      lineUserId = screenData.lineUserId;
      lineDisplayName = screenData.lineDisplayName;
      linePictureUrl = screenData.linePictureUrl;
    });
    debugPrint('=== loadApprovalDataFromScreenData ===');
    debugPrint('LINE data from screenData:');
    debugPrint('  lineUserId: ${screenData.lineUserId}');
    debugPrint('  lineDisplayName: ${screenData.lineDisplayName}');
    debugPrint('  linePictureUrl: ${screenData.linePictureUrl}');
    debugPrint('Local state after load:');
    debugPrint('  lineUserId: $lineUserId');
    debugPrint('  lineDisplayName: $lineDisplayName');
    debugPrint('Approval data: position=${positionController.text}, poRole=$poApprovalRole');
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

  void getData(String username) {
    headerEdit = global.language("show");
    isEditMode = false;
    context.read<UserBloc>().add(UserGet(username: username));
  }

  void switchToEdit(UserModel value) {
    setState(() {
      username = value.username;
      getData(username);
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
        title: Text(global.language('user')),
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
          const ManualButton(path: 'settings-user'),
          /// ปุ่ม Cleanup - ลบ users ที่ username ว่าง
          Padding(
            padding: const EdgeInsets.only(right: 10.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () => _cleanupEmptyUsers(),
              icon: const Icon(Icons.cleaning_services),
              tooltip: global.language('cleanup_invalid_users'),
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
                      username = "";
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
                    (element) => element.username == username,
                  ),
                );
                if (index > 0) {
                  username = listData[index - 1].username;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  getData(username);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.username == username,
                  ),
                );
                username = listData[index + 1].username;
                currentListIndex = index + 1;
                isKeyDown = true;
                getData(username);
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
                  ListFontSizeControl(onChanged: () => setState(() {})),
                const SizedBox(width: 4),
                if (listData.isNotEmpty)
                  Text(
                    '(${listData.length})',
                    style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor),
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
                    flex: 4,
                    child: Text(
                      global.language("user_name"),
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
                      global.language("user_role"),
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
                    flex: 4,
                    child: Text(
                      'LINE',
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                      textAlign: TextAlign.center,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView(
                controller: listScrollController,
                children: listData
                    .map((value) => listObject(listData.indexOf(value), value))
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

  Widget listObject(int index, UserModel value) {
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    bool selected = username == value.username;
    TextStyle textStyle = TextStyle(
      fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
      fontSize: (selected)
          ? global.deviceConfig.listDataFontSize + 2.0
          : global.deviceConfig.listDataFontSize,
    );
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
      onTap: () {
        setState(() {
          discardData(
            callBack: () {
              isSaveAllow = false;
              isEditMode = false;
              username = value.username;
              getData(username);
              WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                tabController.animateTo(1);
              });
            },
          );
        });
      },
      onDoubleTap: () {
        switchToEdit(value);
      },
      child: Container(
        key: index < listKeys.length ? listKeys[index] : null,
        decoration: BoxDecoration(
          color: (username == value.username)
              ? global.theme.rowSelectedColor
              : (index % 2 == 0)
              ? global.theme.columnAlternateEvenColor
              : global.theme.columnAlternateOddColor,
        ),
        padding: EdgeInsets.symmetric(
          horizontal: 10,
          vertical: global.deviceConfig.listDataLineSpace + 10,
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Expanded(
              flex: 4,
              child: Text(
                value.username,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 2,
              child: Text(
                (value.role == 0
                    ? global.language("role_user")
                    : (value.role == 1)
                    ? global.language("role_admin")
                    : global.language("role_owner")),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 4,
              child: Center(
                child: _buildLineButtonForList(value),
              ),
            ),
          ],
        ),
      ),
    ),
    );
  }

  /// สร้างปุ่มเชื่อมต่อ LINE สำหรับแสดงใน list
  Widget _buildLineButtonForList(UserModel user) {
    final hasLine = user.lineUserId != null && user.lineUserId!.isNotEmpty;

    // ถ้าเชื่อมต่อ LINE แล้ว แสดงรูป ชื่อ และปุ่มจัดการ
    if (hasLine) {
      return Column(
        mainAxisAlignment: MainAxisAlignment.center,
        mainAxisSize: MainAxisSize.min,
        children: [
          // แถวแรก: รูปโปรไฟล์และชื่อ LINE
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            mainAxisSize: MainAxisSize.min,
            children: [
              // รูปโปรไฟล์ LINE
              if (user.linePictureUrl != null && user.linePictureUrl!.isNotEmpty)
                ClipOval(
                  child: Image.network(
                    user.linePictureUrl!,
                    width: 28,
                    height: 28,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) => Container(
                      width: 28,
                      height: 28,
                      decoration: BoxDecoration(
                        color: Colors.green.shade100,
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(Icons.person, size: 18, color: Colors.green),
                    ),
                  ),
                )
              else
                Container(
                  width: 28,
                  height: 28,
                  decoration: BoxDecoration(
                    color: Colors.green.shade100,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.person, size: 18, color: Colors.green),
                ),
              const SizedBox(width: 6),
              // ชื่อ LINE
              Flexible(
                child: Text(
                  user.lineDisplayName ?? '',
                  style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w500),
                  overflow: TextOverflow.ellipsis,
                  maxLines: 1,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          // แถวที่สอง: ปุ่มจัดการ (ส่งข้อความ, เชื่อมต่อใหม่, ยกเลิก)
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            mainAxisSize: MainAxisSize.min,
            children: [
              // ปุ่มทดสอบส่งข้อความ LINE
              _buildIconButton(
                icon: Icons.send,
                color: global.theme.primaryColor,
                tooltip: global.language('send_test_message'),
                onPressed: () => _testLinePush(user),
              ),
              const SizedBox(width: 4),
              // ปุ่มเชื่อมต่อใหม่
              _buildIconButton(
                icon: Icons.refresh,
                color: Colors.orange,
                tooltip: global.language('reconnect_line'),
                onPressed: () => _linkLineOA(user),
              ),
              const SizedBox(width: 4),
              // ปุ่มยกเลิกการเชื่อมต่อ
              _buildIconButton(
                icon: Icons.link_off,
                color: Colors.red,
                tooltip: global.language('unlink_line'),
                onPressed: () => _unlinkLine(user),
              ),
            ],
          ),
        ],
      );
    }

    // ยังไม่เชื่อมต่อ LINE - แสดงปุ่มเชื่อมต่อ
    return SizedBox(
      height: 32,
      child: ElevatedButton.icon(
        onPressed: () => _linkLineOA(user),
        style: ElevatedButton.styleFrom(
          backgroundColor: Colors.green,
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
          minimumSize: const Size(0, 32),
          textStyle: const TextStyle(fontSize: 11),
        ),
        icon: Icon(Icons.link, size: 14),
        label: Text(global.language('connect_line')),
      ),
    );
  }

  /// สร้างปุ่ม icon สำหรับจัดการ LINE
  Widget _buildIconButton({
    required IconData icon,
    required Color color,
    required String tooltip,
    required VoidCallback onPressed,
  }) {
    return Tooltip(
      message: tooltip,
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(4),
        child: Container(
          width: 28,
          height: 24,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(4),
          ),
          child: Icon(icon, size: 14, color: Colors.white),
        ),
      ),
    );
  }

  /// ยกเลิกการเชื่อมต่อ LINE
  void _unlinkLine(UserModel user) async {
    if (!mounted) return;

    // แสดง dialog ยืนยัน
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('unlink_line')),
        content: Text(
          'ต้องการยกเลิกการเชื่อมต่อ LINE ของ "${user.lineDisplayName ?? user.username}" หรือไม่?\n\n'
          'ผู้ใช้จะไม่ได้รับการแจ้งเตือนผ่าน LINE อีกต่อไป',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: Text(global.language('confirm')),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    final shopId = global.getShopId();

    // 1. ลบข้อมูล LINE จาก MongoDB Atlas (lineoa-liff)
    // ใช้ POST แทน DELETE เพราะ Vercel block DELETE requests
    try {
      final dio = Dio();
      const lineoaBaseUrl = 'https://dev-api.bcaicloud.com/liff';
      await dio.post(
        '$lineoaBaseUrl/api/link',
        data: {
          'shopId': shopId,
          'employeeCode': user.username,
          'action': 'unlink',
        },
      );
      debugPrint('LINE unlink from MongoDB Atlas success');
    } catch (lineError) {
      debugPrint('LINE unlink from MongoDB Atlas error: $lineError');
      // ไม่ต้อง throw เพราะยังต้องลบจาก mainapi ด้วย
    }

    // 2. บันทึกข้อมูลโดยลบ LINE data ใน mainapi
    setState(() {
      // อัปเดต local state ถ้าเป็น user ที่กำลัง edit
      if (user.username == screenData.username) {
        lineUserId = null;
        lineDisplayName = null;
        linePictureUrl = null;
        screenData.lineUserId = null;
        screenData.lineDisplayName = null;
        screenData.linePictureUrl = null;
      }
    });

    // สร้าง UserModel สำหรับ save โดยลบ LINE data
    final updatedUser = UserModel(
      username: user.username,
      editusername: '',
      shopid: shopId,
      role: user.role,
      position: user.position,
      department: user.department,
      lineUserId: '', // ล้างค่า LINE
      lineDisplayName: '',
      linePictureUrl: '',
      poApproval: user.poApproval,
      quotationApproval: user.quotationApproval,
    );

    if (!mounted) return;

    // บันทึกไปยัง mainapi
    context.read<UserBloc>().add(UserSaveAndUpdate(userModel: updatedUser));

    if (!mounted) return;

    global.showWarningSnackBar(context, 'ยกเลิกการเชื่อมต่อ LINE ของ ${user.lineDisplayName ?? user.username} แล้ว');
  }

  /// ลบ users ที่ username ว่างออกจากระบบ และ cleanup LINE linking ที่ไม่มีผู้ใช้แล้ว
  void _cleanupEmptyUsers() async {
    if (!mounted) return;

    // แสดง dialog ยืนยัน
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('cleanup_invalid_users')),
        content: Text(
          global.language('cleanup_users_confirm_message'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: Text(global.language('confirm')),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    int mainApiDeleted = 0;
    int lineDeleted = 0;

    try {
      final client = Client().init();

      // 1. ลบ users ที่ username ว่างจาก mainapi
      final response = await client.delete('/shop/users/cleanup');

      if (!mounted) return;

      final data = response.data;
      if (data['success'] == true) {
        mainApiDeleted = data['data']?['deleted_count'] ?? 0;
      }

      // 2. เรียก lineoa-liff API เพื่อ cleanup orphaned LINE records
      // รวบรวมรายชื่อ valid usernames จาก listData
      final validUsernames = listData
          .where((u) => u.username.trim().isNotEmpty)
          .map((u) => u.username)
          .toList();

      try {
        final dio = Dio();
        const lineoaBaseUrl = 'https://dev-api.bcaicloud.com/liff';
        final shopId = global.getShopId();

        // ใช้ POST แทน DELETE เพราะ Vercel block DELETE requests
        final lineResponse = await dio.post(
          '$lineoaBaseUrl/api/link',
          data: {
            'shopId': shopId,
            'action': 'cleanup',
            'validUsers': validUsernames,
          },
        );

        if (lineResponse.data['success'] == true) {
          lineDeleted = lineResponse.data['data']?['deleted_count'] ?? 0;
        }
      } catch (lineError) {
        debugPrint('LINE cleanup error: $lineError');
        // ไม่ต้อง throw เพราะอาจเป็นแค่ไม่มีข้อมูล LINE ที่ต้องลบ
      }

      if (!mounted) return;

      // แสดงผลลัพธ์รวม
      global.showSuccessSnackBar(context, 'ลบข้อมูลแล้ว:\n- ผู้ใช้ที่ไม่ถูกต้อง: $mainApiDeleted รายการ\n- ข้อมูล LINE: $lineDeleted รายการ');

      // รีโหลดข้อมูล
      setState(() {
        listData.clear();
      });
      loadDataList(searchText);

    } catch (e) {
      if (!mounted) return;
      global.showErrorSnackBar(context, 'เกิดข้อผิดพลาด: $e');
    }
  }

  void _linkLineOA(UserModel user) async {
    if (!mounted) return;

    // Use QR code dialog with LIFF integration
    const liffId = '2008792333-tziDpgnJ';
    const liffBaseUrl = 'https://dev-api.bcaicloud.com/liff';

    await showLineLinkQRDialog(
      context: context,
      employeeCode: user.username,
      employeeName: user.username,
      shopId: global.getShopId(),
      liffId: liffId,
      liffBaseUrl: liffBaseUrl,
      onLinked: (result) {
        // Update LINE data state variables และ screenData
        if (mounted) {
          setState(() {
            lineUserId = result.userId;
            lineDisplayName = result.displayName;
            linePictureUrl = result.pictureUrl;
            // อัปเดต screenData ด้วย
            screenData.lineUserId = result.userId;
            screenData.lineDisplayName = result.displayName;
            screenData.linePictureUrl = result.pictureUrl;
          });

          global.showSuccessSnackBar(context, 'เชื่อมต่อ LINE สำเร็จ: ${result.displayName} - กำลังบันทึก...');

          // ตรวจสอบว่าเป็นการ link LINE ของตัวเองหรือไม่
          // getString("user") เก็บต่างรูปแบบตาม login method:
          //   Password login → plain username เช่น "maxkorn"
          //   Google login → JSON string เช่น {"email":"maxkorn@gmail.com",...}
          //   Google login (web) → email string เช่น "maxkorn@gmail.com"
          final storedUser = appConfig.getString("user") ?? '';
          String currentUser = storedUser;
          if (storedUser.startsWith("{") && storedUser.endsWith("}")) {
            try {
              final userJson = jsonDecode(storedUser);
              currentUser = userJson['email'] ?? userJson['name'] ?? storedUser;
            } catch (_) {
              // parse ไม่ได้ ใช้ค่าเดิม
            }
          }
          final isLinkingSelf = user.username == currentUser;

          if (isLinkingSelf) {
            // ใช้ endpoint /profile/my-line สำหรับ link LINE ของตัวเอง
            debugPrint('Linking own LINE account - using /profile/my-line');
            context.read<UserBloc>().add(UserSaveMyLineData(
              lineUserId: result.userId,
              lineDisplayName: result.displayName,
              linePictureUrl: result.pictureUrl ?? '',
            ));
          } else {
            // ใช้ endpoint /shop/permission สำหรับ owner แก้ไข LINE ของ user อื่น
            debugPrint('Linking LINE for other user - using /shop/permission');
            // ต้อง set screenData และ controller จาก user object ก่อน
            // เพราะอาจเรียกจาก list view ที่ยังไม่ได้เปิด edit screen
            screenData = user;
            usernameController.text = user.username;
            userRole = user.role;
            // อัปเดต LINE data ใน screenData อีกครั้งหลัง copy จาก user
            screenData.lineUserId = result.userId;
            screenData.lineDisplayName = result.displayName;
            screenData.linePictureUrl = result.pictureUrl;
            saveOrUpdateData();
          }
        }
      },
    );
  }

  /// ทดสอบส่งข้อความ LINE Push
  void _testLinePush(UserModel user) async {
    if (!mounted) return;

    if (user.lineUserId == null || user.lineUserId!.isEmpty) {
      global.showWarningSnackBar(context, 'ผู้ใช้ยังไม่ได้เชื่อมต่อ LINE');
      return;
    }

    // แสดง loading
    global.showInfoSnackBar(context, 'กำลังส่งข้อความทดสอบ...');

    try {
      final dio = Dio();
      // ใช้ URL จากตั้งค่าระบบ
      final apiUrl = global.goApiUrlPath('api/approval/test-line-push');

      final response = await dio.post(
        apiUrl,
        data: {
          'line_user_id': user.lineUserId,
          'user_name': user.lineDisplayName ?? user.username,
          'shop_id': global.appConfig.getString('shopId') ?? 'default',
        },
      );

      if (!mounted) return;

      if (response.data['success'] == true) {
        global.showSuccessSnackBar(context, response.data['message'] ?? 'ส่งข้อความสำเร็จ');
      } else {
        global.showErrorSnackBar(context, 'เกิดข้อผิดพลาด: ${response.data['message']}');
      }
    } catch (e) {
      if (!mounted) return;
      global.showErrorSnackBar(context, 'เกิดข้อผิดพลาด: $e');
    }
  }

  /// สร้าง Chip สำหรับเลือกบทบาทการอนุมัติ (แบบ flexible รองรับหลายประเภทเอกสาร)
  Widget _buildApprovalRoleChipFor({
    required int currentRole,
    required int roleValue,
    required String label,
    required IconData icon,
    required Color themeColor,
    required Function(int) onRoleChanged,
  }) {
    final isSelected = currentRole == roleValue;
    return FilterChip(
      selected: isSelected,
      label: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            icon,
            size: 16,
            color: isSelected ? Colors.white : themeColor,
          ),
          const SizedBox(width: 4),
          Text(label),
        ],
      ),
      selectedColor: themeColor,
      checkmarkColor: Colors.white,
      labelStyle: TextStyle(
        color: isSelected ? Colors.white : themeColor,
        fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
      ),
      onSelected: isEditMode
          ? (selected) {
              setState(() {
                onRoleChanged(selected ? roleValue : 0);
                isDataChange = true;
              });
            }
          : null,
    );
  }

  /// สร้าง Section สำหรับการอนุมัติแต่ละประเภทเอกสาร
  Widget _buildDocumentApprovalSection({
    required String title,
    required IconData icon,
    required Color themeColor,
    required int currentRole,
    required TextEditingController maxAmountController,
    required Function(int) onRoleChanged,
  }) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: themeColor.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: themeColor.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, color: themeColor, size: 20),
              const SizedBox(width: 8),
              Text(
                title,
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 14,
                  color: themeColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // บทบาทการอนุมัติ
          const Text(
            'บทบาทการอนุมัติ',
            style: TextStyle(fontWeight: FontWeight.w500, fontSize: 13),
          ),
          SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              _buildApprovalRoleChipFor(
                currentRole: currentRole,
                roleValue: 0,
                label: global.language('no_permission_title'),
                icon: Icons.block,
                themeColor: themeColor,
                onRoleChanged: onRoleChanged,
              ),
              _buildApprovalRoleChipFor(
                currentRole: currentRole,
                roleValue: 1,
                label: global.language('level_1'),
                icon: Icons.looks_one,
                themeColor: themeColor,
                onRoleChanged: onRoleChanged,
              ),
              _buildApprovalRoleChipFor(
                currentRole: currentRole,
                roleValue: 2,
                label: global.language('level_2'),
                icon: Icons.looks_two,
                themeColor: themeColor,
                onRoleChanged: onRoleChanged,
              ),
              _buildApprovalRoleChipFor(
                currentRole: currentRole,
                roleValue: 3,
                label: global.language('level_3'),
                icon: Icons.looks_3,
                themeColor: themeColor,
                onRoleChanged: onRoleChanged,
              ),
              _buildApprovalRoleChipFor(
                currentRole: currentRole,
                roleValue: 4,
                label: global.language('max'),
                icon: Icons.star,
                themeColor: themeColor,
                onRoleChanged: onRoleChanged,
              ),
            ],
          ),
          const SizedBox(height: 8),
          // คำอธิบายระดับการอนุมัติ
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: Colors.grey.shade100,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language('level_description'),
                  style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.grey.shade700),
                ),
                SizedBox(height: 4),
                Text(
                  global.language('approval_level_1_desc'),
                  style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
                ),
                Text(
                  global.language('approval_level_2_desc'),
                  style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
                ),
                Text(
                  global.language('approval_level_3_desc'),
                  style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
                ),
                Text(
                  global.language('approval_level_max_desc'),
                  style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),

          // วงเงินอนุมัติสูงสุด
          TextFormField(
            controller: maxAmountController,
            readOnly: !isEditMode,
            keyboardType: TextInputType.number,
            onChanged: (value) {
              isDataChange = true;
            },
            decoration: InputDecoration(
              labelText: global.language('max_approval_amount_label'),
              hintText: global.language('zero_means_unlimited'),
              border: const OutlineInputBorder(),
              prefixIcon: const Icon(Icons.attach_money),
              suffixText: global.language('baht'),
              helperText: currentRole == 0 ? global.language('must_select_approval_role_first') : null,
            ),
            enabled: isEditMode && currentRole > 0,
          ),
        ],
      ),
    );
  }


  void saveOrUpdateData() async {
    screenData.shopid = global.getShopId();

    screenData.editusername = screenData.username;

    if (screenData.editusername == usernameController.text) {
      screenData.editusername = "";
    }

    screenData.username = usernameController.text;
    screenData.role = userRole;

    // === ข้อมูลพนักงาน ===
    screenData.position = positionController.text.trim();
    screenData.department = departmentController.text.trim();

    // === ข้อมูล LINE OA ===
    screenData.lineUserId = lineUserId;
    screenData.lineDisplayName = lineDisplayName;
    screenData.linePictureUrl = linePictureUrl;

    debugPrint('=== saveOrUpdateData ===');
    debugPrint('LINE data being saved:');
    debugPrint('  lineUserId (local): $lineUserId');
    debugPrint('  lineDisplayName (local): $lineDisplayName');
    debugPrint('  screenData.lineUserId: ${screenData.lineUserId}');
    debugPrint('  screenData.lineDisplayName: ${screenData.lineDisplayName}');

    // === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
    screenData.poApproval = DocumentApprovalModel(
      approvalRole: poApprovalRole,
      maxApprovalAmount: double.tryParse(poMaxApprovalAmountController.text.replaceAll(',', '')) ?? 0,
    );
    screenData.quotationApproval = DocumentApprovalModel(
      approvalRole: quotationApprovalRole,
      maxApprovalAmount: double.tryParse(quotationMaxApprovalAmountController.text.replaceAll(',', '')) ?? 0,
    );

    if (!mounted) return;
    // บันทึกทั้งหมดไปยัง mainapi (รวมข้อมูลพนักงาน, LINE, และการอนุมัติ)
    context.read<UserBloc>().add(UserSaveAndUpdate(userModel: screenData));
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, top: 10),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Username TextField
            Expanded(
              child: TextFormField(
                enabled: (appConfig.getString("createdby") == screenData.username)
                    ? false
                    : true,
                readOnly: !isEditMode,
                onChanged: (value) {
                  isDataChange = true;
                },
                textAlign: TextAlign.left,
                controller: usernameController,
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("user_name"),
                ),
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'This field is required';
                  }
                  return null;
                },
              ),
            ),
          ],
        ),
      ),
    );

    formWidgets.add(
      Container(
        width: double.infinity,
        padding: const EdgeInsets.only(
          left: 10,
          right: 10,
          top: 10,
          bottom: 15,
        ),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 0,
              groupValue: userRole,
              onChanged:
                  ((appConfig.getString("createdby") != screenData.username))
                  ? (value) {
                      setState(() {
                        userRole = value as int;
                      });
                    }
                  : null,
            ),
            Text(global.language("role_user")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 1,
              groupValue: userRole,
              onChanged:
                  ((appConfig.getString("createdby") != screenData.username))
                  ? (value) {
                      setState(() {
                        userRole = value as int;
                      });
                    }
                  : null,
            ),
            Text(global.language("role_admin")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 2,
              groupValue: userRole,
              onChanged:
                  ((appConfig.getString("createdby") != screenData.username))
                  ? (value) {
                      setState(() {
                        userRole = value as int;
                      });
                    }
                  : null,
            ),
            Text(global.language("role_owner")),
          ],
        ),
      ),
    );

    // Line OA Section - แสดงเฉพาะเมื่อเชื่อมต่อ LINE แล้ว
    if (lineUserId != null && lineUserId!.isNotEmpty) {
      formWidgets.add(
        Container(
          width: double.infinity,
          margin: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.green.shade50,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.green.shade200),
          ),
          child: Row(
            children: [
              if (linePictureUrl != null && linePictureUrl!.isNotEmpty)
                ClipRRect(
                  borderRadius: BorderRadius.circular(20),
                  child: Image.network(
                    linePictureUrl!,
                    width: 40,
                    height: 40,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) => Container(
                      width: 40,
                      height: 40,
                      decoration: BoxDecoration(
                        color: Colors.green.shade100,
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: const Icon(Icons.person, color: Colors.green),
                    ),
                  ),
                )
              else
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: Colors.green.shade100,
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: const Icon(Icons.person, color: Colors.green),
                ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      lineDisplayName ?? 'LINE User',
                      style: const TextStyle(
                        fontWeight: FontWeight.w500,
                        fontSize: 14,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Row(
                      children: [
                        const Icon(Icons.check_circle, color: Colors.green, size: 14),
                        const SizedBox(width: 4),
                        Text(
                          'เชื่อมต่อ LINE แล้ว',
                          style: TextStyle(
                            color: Colors.green.shade600,
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      );
    }

    // === Section: ข้อมูลพนักงาน (ตำแหน่งงาน/แผนก) ===
    formWidgets.add(
      Container(
        width: double.infinity,
        margin: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: Colors.grey.shade50,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.grey.shade300),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.badge, color: Colors.grey.shade700, size: 20),
                const SizedBox(width: 8),
                Text(
                  'ข้อมูลพนักงาน',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                    color: Colors.grey.shade700,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),

            // ตำแหน่งงาน (Autocomplete)
            Autocomplete<String>(
              initialValue: TextEditingValue(text: positionController.text),
              optionsBuilder: (TextEditingValue textEditingValue) {
                if (textEditingValue.text.isEmpty) {
                  return commonPositions;
                }
                return commonPositions.where((String option) {
                  return option.toLowerCase().contains(textEditingValue.text.toLowerCase());
                });
              },
              onSelected: (String selection) {
                positionController.text = selection;
                isDataChange = true;
              },
              fieldViewBuilder: (context, controller, focusNode, onSubmitted) {
                // ซิงค์ค่าจาก positionController — defer เพื่อไม่ให้ setState ตอน build
                if (controller.text != positionController.text && positionController.text.isNotEmpty) {
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    if (mounted && controller.text != positionController.text) {
                      controller.text = positionController.text;
                    }
                  });
                }
                return TextFormField(
                  controller: controller,
                  focusNode: focusNode,
                  readOnly: !isEditMode,
                  onChanged: (value) {
                    positionController.text = value;
                    isDataChange = true;
                  },
                  decoration: InputDecoration(
                    labelText: global.language('job_position'),
                    hintText: global.language('job_position_hint'),
                    border: OutlineInputBorder(),
                    prefixIcon: Icon(Icons.work),
                  ),
                );
              },
            ),
            const SizedBox(height: 12),

            // แผนก (Autocomplete)
            Autocomplete<String>(
              initialValue: TextEditingValue(text: departmentController.text),
              optionsBuilder: (TextEditingValue textEditingValue) {
                if (textEditingValue.text.isEmpty) {
                  return commonDepartments;
                }
                return commonDepartments.where((String option) {
                  return option.toLowerCase().contains(textEditingValue.text.toLowerCase());
                });
              },
              onSelected: (String selection) {
                departmentController.text = selection;
                isDataChange = true;
              },
              fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
                // ซิงค์ค่าจาก departmentController — defer เพื่อไม่ให้ setState ตอน build
                if (controller.text != departmentController.text && departmentController.text.isNotEmpty) {
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    if (mounted && controller.text != departmentController.text) {
                      controller.text = departmentController.text;
                    }
                  });
                }
                return TextFormField(
                  controller: controller,
                  focusNode: focusNode,
                  readOnly: !isEditMode,
                  onChanged: (value) {
                    departmentController.text = value;
                    isDataChange = true;
                  },
                  decoration: InputDecoration(
                    labelText: global.language('department'),
                    hintText: global.language('department_hint'),
                    border: OutlineInputBorder(),
                    prefixIcon: Icon(Icons.business),
                  ),
                );
              },
            ),
          ],
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
      backgroundColor: global.theme.backgroundColor,
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
        title: Text(headerEdit + global.language("user")),
        centerTitle: true,
        actions: <Widget>[
          // ปุ่มลบ - แสดงเฉพาะเมื่อเลือก user แล้ว
          if (username.trim().isNotEmpty)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  debugPrint('Delete button pressed - username: $username');
                  showDialog<String>(
                    context: context,
                    builder: (BuildContext context) => AlertDialog(
                      title: Text(global.language('delete_confirm')),
                      content: Text('${global.language("confirm_delete")} "$username"?'),
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
                            debugPrint('Confirming delete for: $username');
                            context.read<UserBloc>().add(
                              UserDelete(username: username),
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
          if (isSaveAllow == false && username.trim().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  switchToEdit(
                    listData[listData.indexOf(
                      listData.firstWhere(
                        (element) => element.username == username,
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

    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<UserBloc, UserState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is UserLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.users.isNotEmpty) {
                        // กรองออก users ที่ username ว่าง
                        final validUsers = state.users.where((user) => user.username.trim().isNotEmpty).toList();
                        // เรียงตามชื่อผู้ใช้งาน (A-Z)
                        validUsers.sort((a, b) => a.username.toLowerCase().compareTo(b.username.toLowerCase()));
                        listData.addAll(validUsers);
                        // สร้าง keys ใหม่ให้ตรงกับจำนวน items
                        listKeys.clear();
                        for (int i = 0; i < listData.length; i++) {
                          listKeys.add(GlobalKey());
                        }
                      }
                    });
                  }
                  if (state is UserLoadFailed) {
                    setState(() {
                      loadingData = false;

                      global.showSnackBar(
                        context,
                        const Icon(Icons.error, color: Colors.white),
                        state.message,
                        Colors.red,
                      );
                    });
                  }
                  // Save
                  if (state is UserSaveAndUpdateSuccess) {
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
                  if (state is UserSaveAndUpdateFailed) {
                    setState(() {
                      clearEditData();
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        "//${global.language("not_success_save")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }

                  // Delete
                  if (state is UserDeleteSuccess) {
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

                  // Get
                  if (state is UserGetSuccess) {
                    setState(() {
                      isDataChange = false;

                      screenData = state.user;
                      usernameController.text = screenData.username;
                      userRole = screenData.role;

                      // Clear approval data ก่อน (จะโหลดจาก MongoDB)
                      positionController.text = '';
                      departmentController.text = '';
                      approvalRole = 0;
                      maxApprovalAmountController.text = '';
                      // Clear document-type approval data
                      poApprovalRole = 0;
                      poMaxApprovalAmountController.text = '';
                      quotationApprovalRole = 0;
                      quotationMaxApprovalAmountController.text = '';

                      if (isEditMode) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                      }
                    });

                    // โหลดข้อมูลจาก mainapi (ข้อมูลมาพร้อมกับ screenData แล้ว รวม LINE data)
                    WidgetsBinding.instance.addPostFrameCallback((_) {
                      if (mounted) loadApprovalDataFromScreenData();
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
