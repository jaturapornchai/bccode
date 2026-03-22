import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/profile/profile_bloc.dart';
import 'package:smlaicloud/bloc/shop/shop_bloc.dart';
import 'package:smlaicloud/bloc/user/user_bloc.dart';
import 'package:smlaicloud/components/singin_button.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/tabbed_menu_screen.dart';
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/create_shop_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/shop_list_model.dart';
import 'package:smlaicloud/model/shop_model.dart';
import 'package:smlaicloud/select_language_screen.dart';
import 'package:smlaicloud/usersystem/otp/telephone_screen.dart';
import 'package:smlaicloud/usersystem/utils.dart';
import 'package:smlaicloud/usersystem/business_type_selector.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:smlaicloud/bloc/list_shop/list_shop_bloc.dart';
import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';
import 'package:smlaicloud/bloc/shop_select/shop_select_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/flavors.dart';
import 'package:smlaicloud/screens/currency/currency_setup_dialog.dart';
import 'package:smlaicloud/repositories/company_branch_repository.dart';
import 'package:url_launcher/url_launcher.dart';

class LoginShopScreen extends StatefulWidget {
  const LoginShopScreen({super.key});

  @override
  State<LoginShopScreen> createState() => LoginShopScreenState();
}

class LoginShopScreenState extends State<LoginShopScreen> {
  final _formKeyCreateShop = GlobalKey<FormState>();

  /// 0 = login , 1 = list shop , 2 = create shop , 3 = select branch
  int stateScreen = 0;
  late FirebaseAuth _auth;

  global.LoginEnum loginType = global.LoginEnum.none;
  late CreateShopModel createShopData;

  int _indexStep = 0;
  bool _isCreatingShop = false;

  List<CompanyBranchModel> companyBranchListData = [];

  // UI/UX enhancements for branch list
  final TextEditingController _searchBranchController = TextEditingController();
  String _searchBranchQuery = '';
  List<CompanyBranchModel> filteredBranchList = [];

  List<LanguageModel> defaultlanguageList = [
    LanguageModel(code: "th", codeTranslator: "th", name: "Thai", isuse: false),
    LanguageModel(code: "en", codeTranslator: "en", name: "English", isuse: false),
    LanguageModel(code: "cn", codeTranslator: "cn", name: "Chinese", isuse: false),
    LanguageModel(code: "ja", codeTranslator: "ja", name: "Japanese", isuse: false),
    LanguageModel(code: "ko", codeTranslator: "ko", name: "Korean", isuse: false),
    LanguageModel(code: "lo", codeTranslator: "lo", name: "Lao", isuse: false),
    LanguageModel(code: "my", codeTranslator: "my", name: "Burmese", isuse: false),
    LanguageModel(code: "ms", codeTranslator: "ms", name: "Malaysian", isuse: false),
    LanguageModel(code: "vi", codeTranslator: "vi", name: "Vietnamese", isuse: false),
    LanguageModel(code: "km", codeTranslator: "km", name: "Khmer", isuse: false),
  ];

  List<BusinessTypeModel> businessTypeList = [];

  // UI/UX enhancements for shop list
  final TextEditingController _searchController = TextEditingController();
  String _currentFilter = 'all'; // all, frequent, recent
  String _searchQuery = '';
  List<ShopListModel> filteredShopList = [];

  // Accent colors for shop cards — muted, professional tones
  final List<Color> _shopAccentColors = [
    const Color(0xFF3949AB), // Indigo
    const Color(0xFF00897B), // Teal
    const Color(0xFF5C6BC0), // Light indigo
    const Color(0xFF43A047), // Green
    const Color(0xFFEF6C00), // Orange
    const Color(0xFF7B1FA2), // Purple
    const Color(0xFF1565C0), // Blue
    const Color(0xFFC62828), // Red
    const Color(0xFF00838F), // Cyan
    const Color(0xFF4E342E), // Brown
  ];

  Future<UserCredential?> googleSignIn() async {
    try {
      // Initialize and authenticate with Google Sign-In 7.x API
      final GoogleSignIn googleSignIn = GoogleSignIn.instance;
      // ใช้ clientId เฉพาะสำหรับ Windows Desktop
      final String? clientId = Platform.isWindows ? global.googleSignInClientIdWindows : null;
      await googleSignIn.initialize(clientId: clientId);

      // Obtain the auth details from the request
      final GoogleSignInAccount googleUser = await googleSignIn.authenticate();

      // authentication is synchronous in v7.x, not a Future
      final GoogleSignInAuthentication googleAuth = googleUser.authentication;

      // Create a new credential
      final OAuthCredential credential = GoogleAuthProvider.credential(
        // accessToken: googleAuth.accessToken, // Not available in v7.x
        idToken: googleAuth.idToken,
      );

      // print(googleUser);

      // Sign in to Firebase with the Google [UserCredential]

      return await _auth.signInWithCredential(credential);
    } catch (e) {
      // print(e);
      return null;
    }
  }

  Future<String?> getCurrentUserIdToken() async {
    User? currentUser = _auth.currentUser;
    if (currentUser != null) {
      String? idToken = await currentUser.getIdToken();
      return idToken;
    } else {
      // No user is signed in.
      // print("No user is signed in");
      return null;
    }
  }

  @override
  void initState() {
    if (kDebugMode) {
      AppLogger.debug("[LoginShopScreen] initState");
    }
    if (kIsWeb) {
      _auth = FirebaseAuth.instance;
    } else if (Platform.isWindows) {
      _auth = FirebaseAuth.instance;
    } else if (Platform.isMacOS) {
      _auth = FirebaseAuth.instance;
    }

    clearDataCreateShop();

    context.read<ProfileBloc>().add(const GetProfile());
    context.read<ListShopBloc>().add(ListShopLoad());
    super.initState();
  }

  /// ดึงข้อความแสดงการเลือกภาษาตามภาษาปัจจุบัน
  String _getLanguageText() {
    switch (global.userLanguage) {
      case 'en':
        return 'Select language English';
      case 'th':
        return global.language("select_language_thai");
      case 'lo':
        return 'ເລືອກພາສາ ລາວ';
      case 'cn':
        return '选择语言 中文';
      case 'ja':
        return '言語を選択 日本語';
      case 'ko':
        return '언어 선택 한국어';
      default:
        return 'Select language English';
    }
  }

  void clearDataCreateShop() {
    createShopData = CreateShopModel(
      address: [],
      branchcode: '',
      images: [],
      logo: '',
      name1: '',
      names: [],
      profilepicture: '',
      settings: Settings(
        emailowners: [],
        emailstaffs: [],
        isusebranch: false,
        isusedepartment: false,
        languageconfigs: [],
        latitude: 0,
        longitude: 0,
        taxid: '',
        vatrate: 7,
        vattypesale: 0,
        vattypepurchase: 0,
        inquirytypesale: 0,
        inquirytypepurchase: 0,
      ),
      telephone: '',
      businesstype: BusinessTypeModel(),
    );
    createShopData.settings!.languageconfigs!.add(LanguageModel(code: "th", codeTranslator: "th", name: "Thai", isuse: false));
    addTextControllerName();
  }

  void logoutWithGoogle() async {
    // await _auth.signOut();
    await getCurrentUserIdToken();
  }

  void addTextControllerName() {
    createShopData.names = [];
    for (int i = 0; i < createShopData.settings!.languageconfigs!.length; i++) {
      createShopData.names!.add(LanguageDataModel(code: createShopData.settings!.languageconfigs![i].code!, name: ""));
    }
    setState(() {});
  }

  Future<void> getTemplateBusinessType() async {
    const githubRawUrl = 'https://raw.githubusercontent.com/smlsoft/dedepos_template/main/business_type.json';

    try {
      final fileContent = await global.readFileFromGithub(githubRawUrl);
      final businessType = (json.decode(fileContent) as List).map((bank) => BusinessTypeModel.fromJson(bank)).toList();
      businessTypeList = [];

      for (int i = 0; i < businessType.length; i++) {
        businessTypeList.add(businessType[i]);
      }

      setState(() {
        if (createShopData.businesstype!.code!.isEmpty) {
          createShopData.businesstype = BusinessTypeModel(code: businessType[0].code, names: businessType[0].names);
        }
      });
    } catch (error) {
      // Handle error
      // ignore: avoid_print
      if (kDebugMode) {
        AppLogger.error('Error reading file: $error');
      }
    }
  }

  Widget createShop() {
    return Container(
      padding: const EdgeInsets.all(20),
      child: Center(
        child: ConstrainedBox(
          constraints: BoxConstraints(maxWidth: 600),
          child: Container(
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(24),
              boxShadow: [BoxShadow(color: F.primaryGradientColor.withValues(alpha: 0.1), blurRadius: 30, offset: const Offset(0, 10))],
            ),
            child: Column(
              children: [
                // Header
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(colors: [F.primaryGradientColor, F.secondaryGradientColor]),
                    borderRadius: const BorderRadius.only(topLeft: Radius.circular(24), topRight: Radius.circular(24)),
                  ),
                  child: Row(
                    children: [
                      _buildCreateShopHeaderButton(
                        icon: Icons.arrow_back_rounded,
                        onPressed: () {
                          context.read<LoginBloc>().add(const Logout());
                        },
                        tooltip: global.language('logout'),
                      ),
                      Spacer(),
                      Column(
                        children: [
                          const Icon(Icons.add_business_rounded, color: Colors.white, size: 28),
                          SizedBox(height: 4),
                          Text(
                            global.language("create_shop"),
                            style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.white),
                          ),
                        ],
                      ),
                      Spacer(),
                      _buildCreateShopHeaderButton(
                        icon: Icons.list_rounded,
                        onPressed: () {
                          setState(() {
                            stateScreen = 1;
                          });
                        },
                        tooltip: global.language('select_shop'),
                      ),
                    ],
                  ),
                ),
                // Stepper content
                Expanded(
                  child: Theme(
                    data: Theme.of(context).copyWith(colorScheme: Theme.of(context).colorScheme.copyWith(primary: F.primaryGradientColor)),
                    child: Stepper(
                      currentStep: _indexStep,
                      onStepCancel: () {
                        if (_indexStep > 0) {
                          setState(() {
                            _indexStep -= 1;
                          });
                        }
                      },
                      onStepContinue: () {
                        if (kDebugMode) {
                          AppLogger.debug('🏪 onStepContinue: _indexStep=$_indexStep, _isCreatingShop=$_isCreatingShop');
                        }
                        if (_indexStep < 1) {
                          setState(() {
                            addTextControllerName();
                            _indexStep += 1;
                          });
                        } else if (_indexStep == 1) {
                          if (_isCreatingShop) {
                            if (kDebugMode) {
                              AppLogger.debug('🏪 onStepContinue: already creating, skip');
                            }
                            return;
                          }
                          final isValid = _formKeyCreateShop.currentState!.validate();
                          if (kDebugMode) {
                            AppLogger.debug('🏪 form validate: $isValid');
                            AppLogger.debug('🏪 shopData names: ${createShopData.names?.map((n) => "${n.code}=${n.name}").toList()}');
                          }
                          if (isValid) {
                            setState(() => _isCreatingShop = true);
                            createShopData.settings!.languageconfigs![0].isdefault = true;
                            context.read<LoginBloc>().add(CreateShop(createShop: createShopData));
                          }
                        }
                      },
                      onStepTapped: (int index) {
                        setState(() {
                          addTextControllerName();
                          _indexStep = index;
                        });
                      },
                      controlsBuilder: (BuildContext context, ControlsDetails details) {
                        return Padding(
                          padding: EdgeInsets.only(top: 8.0),
                          child: Row(
                            children: <Widget>[
                              if (_indexStep == 0) ElevatedButton(onPressed: details.onStepContinue, child: Text(global.language('next'))),
                              if (_indexStep == 1)
                                ElevatedButton(
                                  onPressed: _isCreatingShop ? null : details.onStepContinue,
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: const Color(0xFF0A6ED1), // SAP Action Blue
                                  ),
                                  child: _isCreatingShop
                                      ? SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                                      : Text(global.language('save')),
                                ),
                              if (_indexStep == 1 && details.stepIndex > 0) TextButton(onPressed: details.onStepCancel, child: Text(global.language('back'))),
                            ],
                          ),
                        );
                      },
                      steps: <Step>[
                        Step(
                          title: Text(global.language('step_1_select_language')),
                          content: Container(
                            alignment: Alignment.centerLeft,
                            child: Column(
                              children: [
                                for (var i = 0; i < createShopData.settings!.languageconfigs!.length; i++)
                                  Container(
                                    margin: EdgeInsets.only(bottom: 10),
                                    child: Row(
                                      children: [
                                        Expanded(
                                          child: Row(
                                            children: [
                                              SizedBox(width: 30, child: Text((i + 1).toString())),
                                              Expanded(
                                                child: ElevatedButton(
                                                  onPressed: () {
                                                    List<LanguageModel> languagesSelectList = [];
                                                    languagesSelectList.addAll(defaultlanguageList);

                                                    for (var selected in createShopData.settings!.languageconfigs!) {
                                                      languagesSelectList.removeWhere((element) => element.code == selected.code);
                                                    }

                                                    showDialog(
                                                      context: context,
                                                      builder: (BuildContext context) {
                                                        return AlertDialog(
                                                          title: Text(global.language('select_language')),
                                                          content: SizedBox(
                                                            width: 300,
                                                            height: 400,
                                                            child: ListView.builder(
                                                              itemCount: languagesSelectList.length,
                                                              itemBuilder: (context, index) {
                                                                return ListTile(
                                                                  title: Row(
                                                                    children: [
                                                                      Text(languagesSelectList[index].name!),
                                                                      const Spacer(),
                                                                      Image.asset('assets/flags/${languagesSelectList[index].code}.png', width: 30, height: 30),
                                                                    ],
                                                                  ),
                                                                  onTap: () {
                                                                    setState(() {
                                                                      Navigator.of(context).pop(languagesSelectList[index]);
                                                                    });
                                                                  },
                                                                );
                                                              },
                                                            ),
                                                          ),
                                                        );
                                                      },
                                                    ).then((value) {
                                                      setState(() {
                                                        if (value != null) {
                                                          createShopData.settings!.languageconfigs![i] = value!;
                                                          createShopData.settings!.languageconfigs![i].isuse = true;
                                                          addTextControllerName();
                                                        }
                                                      });
                                                    });
                                                  },
                                                  child: Row(
                                                    children: [
                                                      (createShopData.settings!.languageconfigs![i].code!.isNotEmpty)
                                                          ? Image.asset('assets/flags/${createShopData.settings!.languageconfigs![i].code!}.png', width: 30, height: 30)
                                                          : Container(),
                                                      const SizedBox(width: 10),
                                                      (createShopData.settings!.languageconfigs![i].code!.isNotEmpty)
                                                          ? Text(defaultlanguageList.firstWhere((element) => element.code == createShopData.settings!.languageconfigs![i].code!).name!)
                                                          : Text(global.language('select_language')),
                                                    ],
                                                  ),
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                        (i == 0)
                                            ? const SizedBox(width: 50)
                                            : SizedBox(
                                                width: 50,
                                                child: IconButton(
                                                  onPressed: () {
                                                    setState(() {
                                                      createShopData.settings!.languageconfigs!.removeAt(i);
                                                    });
                                                  },
                                                  color: Colors.red,
                                                  icon: const Icon(Icons.delete),
                                                ),
                                              ),
                                      ],
                                    ),
                                  ),
                                SizedBox(height: 10),
                                SizedBox(
                                  width: double.infinity,
                                  child: ElevatedButton(
                                    onPressed: () {
                                      setState(() {
                                        createShopData.settings!.languageconfigs!.add(LanguageModel(code: "", codeTranslator: "", name: "", isuse: false));
                                        addTextControllerName();
                                      });
                                    },
                                    child: Text(global.language('add_language')),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                        Step(
                          title: Text(global.language('step_2_enter_shop_name')),
                          content: Form(
                            key: _formKeyCreateShop,
                            child: Column(
                              children: [
                                Padding(
                                  padding: EdgeInsets.all(8.0),
                                  child: SizedBox(
                                    width: double.infinity,
                                    child: ElevatedButton(
                                      /// set color button
                                      style: ElevatedButton.styleFrom(
                                        foregroundColor: Colors.black,
                                        backgroundColor: const Color(0xFFD5E9F7), // SAP Selection
                                      ),
                                      onPressed: () async {
                                        await getTemplateBusinessType();
                                        if (!mounted) return;

                                        final result = await showBusinessTypeSelector(
                                          context: context,
                                          businessTypes: businessTypeList,
                                          selected: createShopData.businesstype,
                                        );
                                        if (result != null) {
                                          setState(() {
                                            createShopData.businesstype!.code = result.code;
                                            createShopData.businesstype!.names = result.names;
                                          });
                                        }
                                      },
                                      child: Text("${global.language("company_type")} ~ ${global.activeLangName(createShopData.businesstype!.names!)}"),
                                    ),
                                  ),
                                ),
                                const SizedBox(height: 10),
                                for (int i = 0; i < createShopData.settings!.languageconfigs!.length; i++)
                                  Padding(
                                    padding: EdgeInsets.all(8.0),
                                    child: TextFormField(
                                      controller: TextEditingController(text: createShopData.names![i].name),
                                      autofocus: true,
                                      onChanged: (value) {
                                        createShopData.names![i].name = value;
                                      },
                                      decoration: InputDecoration(
                                        contentPadding: const EdgeInsets.only(left: 10, top: 0, bottom: 0, right: 10),
                                        floatingLabelBehavior: FloatingLabelBehavior.always,
                                        border: OutlineInputBorder(),
                                        labelText: "${global.language("company_name")} (${createShopData.names![i].code?.toUpperCase() ?? ''})",
                                        labelStyle: const TextStyle(fontSize: 16.0),
                                      ),
                                      onEditingComplete: () {
                                        setState(() {
                                          createShopData.names![i].name = createShopData.names![i].name;
                                        });
                                      },
                                      validator: (value) {
                                        if (i == 0) {
                                          if (value == null || value.isEmpty) {
                                            return global.language('please_enter_value');
                                          }
                                          return null;
                                        } else {
                                          return null;
                                        }
                                      },
                                    ),
                                  ),
                              ],
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildCreateShopHeaderButton({required IconData icon, required VoidCallback onPressed, required String tooltip}) {
    return Tooltip(
      message: tooltip,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onPressed,
          borderRadius: BorderRadius.circular(12),
          child: Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.2), borderRadius: BorderRadius.circular(12)),
            child: Icon(icon, color: Colors.white, size: 22),
          ),
        ),
      ),
    );
  }

  // Helper function to get icon based on shop name
  IconData _getShopIcon(String shopName) {
    String lower = shopName.toLowerCase();
    if (lower.contains('restaurant') || lower.contains(global.language('restaurant')) || lower.contains(global.language('food'))) {
      return Icons.restaurant;
    }
    if (lower.contains('cafe') || lower.contains('คาเฟ') || lower.contains('กาแฟ')) {
      return Icons.local_cafe;
    }
    if (lower.contains('demo') || lower.contains('test') || lower.contains(global.language('alert_testing'))) {
      return Icons.science;
    }
    if (lower.contains('hotel') || lower.contains('โรงแรม')) {
      return Icons.hotel;
    }
    if (lower.contains('retail') || lower.contains('ขายปลีก')) {
      return Icons.shopping_cart;
    }
    return Icons.store;
  }

  // Filter shops based on search query and filter
  void _filterShops(List<ShopListModel> allShops) {
    List<ShopListModel> filtered = allShops;

    // Apply filter first
    if (_currentFilter == 'frequent') {
      // Filter for frequently used shops (shops created by current user)
      String currentUser = appConfig.getString("user") ?? "";
      filtered = filtered.where((shop) => shop.createdby == currentUser).toList();
    } else if (_currentFilter == 'recent') {
      // Filter for recently accessed shops (sort by last access time)
      // For now, show first 10 shops as recent
      filtered = filtered.take(10).toList();
    }

    // Apply search query
    if (_searchQuery.isNotEmpty) {
      filtered = filtered.where((shop) {
        String name = (shop.name.isEmpty) ? global.packName(shop.names!) : shop.name;
        return name.toLowerCase().contains(_searchQuery.toLowerCase());
      }).toList();
    }

    setState(() {
      filteredShopList = filtered;
    });
  }

  // Modern App Bar Widget
  Widget _buildModernAppBar() {
    final userName = global.userLoginData.name.isNotEmpty ? global.userLoginData.name : global.userLoginData.email;
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
      child: Row(
        children: [
          // User Avatar
          Container(
            padding: const EdgeInsets.all(2),
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              border: Border.all(color: Colors.white.withValues(alpha: 0.5), width: 1.5),
            ),
            child: CircleAvatar(
              radius: 19,
              backgroundColor: Colors.white.withValues(alpha: 0.2),
              backgroundImage: global.userLoginData.photourl.isNotEmpty ? NetworkImage(global.userLoginData.photourl) : null,
              child: global.userLoginData.photourl.isEmpty
                  ? Text(
                      userName.isNotEmpty ? userName[0].toUpperCase() : 'U',
                      style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 15),
                    )
                  : null,
            ),
          ),
          const SizedBox(width: 10),
          // Title + user name (single line each)
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  global.language('select_shop'),
                  style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
                ),
                Text(
                  userName,
                  style: TextStyle(fontSize: 12, color: Colors.white.withValues(alpha: 0.8)),
                  overflow: TextOverflow.ellipsis,
                  maxLines: 1,
                ),
              ],
            ),
          ),
          // Action buttons
          if (global.currentLoginMethod != global.LoginEnum.password) ...[
            _buildAppBarButton(
              icon: Icons.add_business_rounded,
              onPressed: () => setState(() { stateScreen = 2; clearDataCreateShop(); }),
              tooltip: global.language('create_shop'),
            ),
            const SizedBox(width: 6),
          ],
          _buildAppBarButton(
            icon: null,
            customWidget: ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: Image.asset('assets/flags/${global.userLanguage}.png', width: 20, height: 14, fit: BoxFit.cover),
            ),
            onPressed: () async {
              final value = await showLanguageSelectionDialog(context);
              if (value != null) {
                global.userLanguage = value;
                global.appConfig.setString('language', global.userLanguage);
                await global.languageSelect(global.userLanguage);
                setState(() {});
              }
            },
            tooltip: _getLanguageText(),
          ),
          const SizedBox(width: 6),
          _buildAppBarButton(
            icon: Icons.menu_book_rounded,
            onPressed: () => launchUrl(Uri.parse('https://bcaicloud.com/manual/loginscreen'), mode: LaunchMode.externalApplication),
            tooltip: global.language('manual'),
          ),
          const SizedBox(width: 6),
          _buildAppBarButton(
            icon: Icons.logout_rounded,
            onPressed: () => context.read<LoginBloc>().add(const Logout()),
            tooltip: global.language('logout'),
            isDestructive: true,
          ),
        ],
      ),
    );
  }

  Widget _buildAppBarButton({IconData? icon, Widget? customWidget, required VoidCallback onPressed, String? tooltip, bool isDestructive = false}) {
    return Tooltip(
      message: tooltip ?? '',
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onPressed,
          borderRadius: BorderRadius.circular(12),
          child: Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: isDestructive ? Colors.red.withValues(alpha: 0.2) : Colors.white.withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: isDestructive ? Colors.red.withValues(alpha: 0.3) : Colors.white.withValues(alpha: 0.2)),
            ),
            child: customWidget ?? Icon(icon, color: isDestructive ? Colors.red[100] : Colors.white, size: 22),
          ),
        ),
      ),
    );
  }

  // Build search bar widget
  Widget _buildSearchBar() {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 20, vertical: 10),
      child: Container(
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [BoxShadow(color: F.primaryGradientColor.withValues(alpha: 0.1), blurRadius: 20, offset: const Offset(0, 4))],
        ),
        child: TextField(
          controller: _searchController,
          decoration: InputDecoration(
            hintText: global.language("search_shop_placeholder"),
            prefixIcon: Container(
              margin: const EdgeInsets.all(12),
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(color: F.primaryGradientColor.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(10)),
              child: Icon(Icons.search_rounded, color: F.primaryGradientColor, size: 20),
            ),
            suffixIcon: _searchQuery.isNotEmpty
                ? IconButton(
                    icon: const Icon(Icons.clear_rounded, color: Colors.grey),
                    onPressed: () {
                      setState(() {
                        _searchController.clear();
                        _searchQuery = '';
                      });
                    },
                  )
                : null,
            hintStyle: TextStyle(color: global.theme.formHintColor, fontSize: 15),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(16), borderSide: BorderSide.none),
            enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(16), borderSide: BorderSide.none),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(16),
              borderSide: BorderSide(color: F.primaryGradientColor, width: 2),
            ),
            contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          ),
          style: TextStyle(color: global.theme.textColor, fontSize: 15),
          onChanged: (value) {
            setState(() {
              _searchQuery = value;
            });
          },
        ),
      ),
    );
  }

  // Build filter chips widget
  Widget _buildFilterChips() {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 20),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: [
            _buildModernFilterChip(
              label: global.language("filter_all"),
              icon: Icons.apps_rounded,
              isSelected: _currentFilter == 'all',
              onSelected: () {
                setState(() {
                  _currentFilter = 'all';
                });
              },
            ),
            SizedBox(width: 10),
            _buildModernFilterChip(
              label: global.language("filter_frequent"),
              icon: Icons.star_rounded,
              isSelected: _currentFilter == 'frequent',
              onSelected: () {
                setState(() {
                  _currentFilter = 'frequent';
                });
              },
            ),
            SizedBox(width: 10),
            _buildModernFilterChip(
              label: global.language("filter_recent"),
              icon: Icons.access_time_rounded,
              isSelected: _currentFilter == 'recent',
              onSelected: () {
                setState(() {
                  _currentFilter = 'recent';
                });
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildModernFilterChip({required String label, required IconData icon, required bool isSelected, required VoidCallback onSelected}) {
    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onSelected,
          borderRadius: BorderRadius.circular(25),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            decoration: BoxDecoration(
              gradient: isSelected ? LinearGradient(colors: [F.primaryGradientColor, F.secondaryGradientColor]) : null,
              color: isSelected ? null : global.theme.cardColor,
              borderRadius: BorderRadius.circular(25),
              border: Border.all(color: isSelected ? Colors.transparent : global.theme.dividerBorderColor),
              boxShadow: isSelected ? [BoxShadow(color: F.primaryGradientColor.withValues(alpha: 0.3), blurRadius: 8, offset: const Offset(0, 4))] : null,
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(icon, size: 18, color: isSelected ? Colors.white : global.theme.iconSecondaryColor),
                const SizedBox(width: 6),
                Text(
                  label,
                  style: TextStyle(color: isSelected ? Colors.white : global.theme.textSecondaryColor, fontWeight: isSelected ? FontWeight.w600 : FontWeight.w500, fontSize: 13),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // Build empty state widget
  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(32),
            decoration: BoxDecoration(color: F.primaryGradientColor.withValues(alpha: 0.1), shape: BoxShape.circle),
            child: Icon(Icons.store_mall_directory_outlined, size: 80, color: F.primaryGradientColor.withValues(alpha: 0.6)),
          ),
          SizedBox(height: 24),
          Text(
            global.language("no_shop_found"),
            style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Color(0xFF1a1a2e)),
          ),
          SizedBox(height: 8),
          Text(
            _searchQuery.isNotEmpty
                ? global.language("try_different_search")
                : (global.currentLoginMethod == global.LoginEnum.password
                    ? global.language("contact_owner_to_add")
                    : global.language("create_new_shop")),
            style: TextStyle(fontSize: 15, color: Colors.grey[500]),
          ),
          const SizedBox(height: 24),
          // ปุ่มสร้างร้านค้า — เฉพาะ Google login เท่านั้น
          if (_searchQuery.isEmpty && global.currentLoginMethod != global.LoginEnum.password)
            ElevatedButton.icon(
              onPressed: () {
                setState(() {
                  stateScreen = 2;
                  clearDataCreateShop();
                });
              },
              icon: Icon(Icons.add_rounded),
              label: Text(global.language('create_shop')),
              style: ElevatedButton.styleFrom(
                backgroundColor: F.primaryGradientColor,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
              ),
            ),
        ],
      ),
    );
  }

  Widget listShopSelect() {
    return BlocBuilder<ListShopBloc, ListShopState>(
      builder: (context, state) {
        if (state is ListShopLoadSuccess) {
          // Filter shops whenever state changes or search/filter changes
          // ใช้ addPostFrameCallback เพื่อไม่ให้เรียก setState ระหว่าง build
          WidgetsBinding.instance.addPostFrameCallback((_) {
            _filterShops(state.shop);
          });

          return Column(
            children: [
              // Search Bar
              _buildSearchBar(),
              const SizedBox(height: 10),
              // Filter Chips
              _buildFilterChips(),
              const SizedBox(height: 16),
              // Shop List or Empty State
              Expanded(
                child: filteredShopList.isEmpty
                    ? _buildEmptyState()
                    : SingleChildScrollView(
                        padding: const EdgeInsets.fromLTRB(24, 8, 24, 24),
                        child: Center(
                          child: Wrap(
                            spacing: 16,
                            runSpacing: 16,
                            alignment: WrapAlignment.center,
                            children: List.generate(filteredShopList.length, (index) => cardItem(filteredShopList[index], index)),
                          ),
                        ),
                      ),
              ),
            ],
          );
        }
        return Container();
      },
    );
  }

  Widget cardItem(ShopListModel data, int index) {
    final accent = _shopAccentColors[index % _shopAccentColors.length];
    final isOwner = appConfig.getString("user") == (data.createdby ?? '');
    final isSelected = data.shopid == global.getShopId();
    final shopName = (data.name.isEmpty) ? global.packName(data.names!) : data.name;
    final shopIcon = _getShopIcon(shopName);
    final roleText = data.role == 0 ? global.language("owner") : global.language("staff");

    return SizedBox(
      width: 220,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(18),
          onTap: () => context.read<ShopSelectBloc>().add(ShopSelect(shop: data)),
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: isSelected ? accent.withValues(alpha: 0.06) : global.theme.cardColor,
              borderRadius: BorderRadius.circular(18),
              border: Border.all(color: isSelected ? accent : global.theme.dividerBorderColor, width: isSelected ? 2 : 1),
              boxShadow: [
                BoxShadow(color: accent.withValues(alpha: isSelected ? 0.15 : 0.05), blurRadius: isSelected ? 16 : 8, offset: const Offset(0, 4)),
              ],
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Icon
                Container(
                  width: 56,
                  height: 56,
                  decoration: BoxDecoration(
                    gradient: LinearGradient(colors: [accent, accent.withValues(alpha: 0.7)], begin: Alignment.topLeft, end: Alignment.bottomRight),
                    borderRadius: BorderRadius.circular(16),
                    boxShadow: [BoxShadow(color: accent.withValues(alpha: 0.3), blurRadius: 8, offset: const Offset(0, 3))],
                  ),
                  child: Icon(shopIcon, size: 28, color: Colors.white),
                ),
                const SizedBox(height: 14),
                // Shop name
                Text(
                  shopName,
                  style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: global.theme.textColor, height: 1.3),
                  textAlign: TextAlign.center,
                  overflow: TextOverflow.ellipsis,
                  maxLines: 2,
                ),
                const SizedBox(height: 6),
                // Shop ID
                Text(
                  'ID: ${data.shopid.length > 8 ? '${data.shopid.substring(0, 8)}...' : data.shopid}',
                  style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor),
                ),
                const SizedBox(height: 10),
                // Badges row
                Wrap(
                  spacing: 6,
                  runSpacing: 4,
                  alignment: WrapAlignment.center,
                  children: [
                    // Role badge
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                      decoration: BoxDecoration(
                        color: isOwner ? Colors.amber.withValues(alpha: 0.15) : global.theme.surfaceColor,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(isOwner ? Icons.star_rounded : Icons.person_rounded, size: 11, color: isOwner ? Colors.amber.shade700 : global.theme.iconSecondaryColor),
                          const SizedBox(width: 3),
                          Text(roleText, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: isOwner ? Colors.amber.shade800 : global.theme.textSecondaryColor)),
                        ],
                      ),
                    ),
                    // Active badge
                    if (isSelected)
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                        decoration: BoxDecoration(color: accent.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(8)),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(Icons.check_circle_rounded, size: 11, color: accent),
                            const SizedBox(width: 3),
                            Text(global.language("in_use"), style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: accent)),
                          ],
                        ),
                      ),
                  ],
                ),
                const SizedBox(height: 12),
                // Enter button
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  decoration: BoxDecoration(
                    color: isSelected ? accent : accent.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(isSelected ? Icons.login_rounded : Icons.arrow_forward_rounded, size: 14, color: isSelected ? Colors.white : accent),
                      const SizedBox(width: 4),
                      Text(
                        isSelected ? global.language("enter") : global.language("select"),
                        style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: isSelected ? Colors.white : accent),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // Helper function to get icon based on branch name
  IconData _getBranchIcon(String branchName) {
    String lower = branchName.toLowerCase();
    if (lower.contains('head office') || lower.contains(global.language('head_office')) || lower.contains('ใหญ่')) {
      return Icons.business;
    }
    if (lower.contains('branch') || lower.contains(global.language('branch'))) {
      return Icons.home_work;
    }
    if (lower.contains('demo') || lower.contains('test') || lower.contains(global.language('alert_testing'))) {
      return Icons.science;
    }
    return Icons.store;
  }

  // Filter branches based on search query only
  void _filterBranches(List<CompanyBranchModel> allBranches) {
    List<CompanyBranchModel> filtered = allBranches;

    // Apply search query
    if (_searchBranchQuery.isNotEmpty) {
      filtered = filtered.where((branch) {
        String name = global.packName(branch.names);
        return name.toLowerCase().contains(_searchBranchQuery.toLowerCase());
      }).toList();
    }

    setState(() {
      filteredBranchList = filtered;
    });
  }

  // Build search bar widget for branches
  Widget _buildBranchSearchBar() {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 20, vertical: 10),
      child: Container(
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [BoxShadow(color: F.primaryGradientColor.withValues(alpha: 0.1), blurRadius: 20, offset: const Offset(0, 4))],
        ),
        child: TextField(
          controller: _searchBranchController,
          decoration: InputDecoration(
            hintText: global.language("search_branch_placeholder"),
            prefixIcon: Container(
              margin: const EdgeInsets.all(12),
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(color: F.primaryGradientColor.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(10)),
              child: Icon(Icons.search_rounded, color: F.primaryGradientColor, size: 20),
            ),
            suffixIcon: _searchBranchQuery.isNotEmpty
                ? IconButton(
                    icon: const Icon(Icons.clear_rounded, color: Colors.grey),
                    onPressed: () {
                      setState(() {
                        _searchBranchController.clear();
                        _searchBranchQuery = '';
                      });
                    },
                  )
                : null,
            hintStyle: TextStyle(color: global.theme.formHintColor, fontSize: 15),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(16), borderSide: BorderSide.none),
            enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(16), borderSide: BorderSide.none),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(16),
              borderSide: BorderSide(color: F.primaryGradientColor, width: 2),
            ),
            contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          ),
          style: TextStyle(color: global.theme.textColor, fontSize: 15),
          onChanged: (value) {
            setState(() {
              _searchBranchQuery = value;
            });
          },
        ),
      ),
    );
  }

  // Build empty state widget for branches
  Widget _buildBranchEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(32),
            decoration: BoxDecoration(color: F.primaryGradientColor.withValues(alpha: 0.1), shape: BoxShape.circle),
            child: Icon(Icons.home_work_outlined, size: 80, color: F.primaryGradientColor.withValues(alpha: 0.6)),
          ),
          SizedBox(height: 24),
          Text(
            global.language("no_branch_found"),
            style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Color(0xFF1a1a2e)),
          ),
          SizedBox(height: 8),
          Text(_searchBranchQuery.isNotEmpty ? global.language("try_different_search") : global.language("create_new_branch"), style: TextStyle(fontSize: 15, color: Colors.grey[500])),
        ],
      ),
    );
  }

  Widget cardItemBranch(CompanyBranchModel data, int index) {
    final accent = _shopAccentColors[index % _shopAccentColors.length];
    final isSelected = appConfig.getString("branch_guidfixed") == data.guidfixed;
    final branchName = global.packName(data.names);
    final branchIcon = _getBranchIcon(branchName);

    return Material(
      color: Colors.transparent,
      child: InkWell(
        borderRadius: BorderRadius.circular(14),
        onTap: () async {
          global.companyBranchSelectData = data;
          appConfig.setString("branch_guidfixed", data.guidfixed);
          appConfig.setInt("branch_total", companyBranchListData.length);

          if (!mounted) return;
          await autoSetupShopDefaults();

          if (!mounted) return;
          await checkAndSetupCurrency(context);

          if (!mounted) return;
          Navigator.of(context).pushAndRemoveUntil(MaterialPageRoute(builder: (_) => TabbedMenuScreen(key: TabbedMenuScreen.globalKey)), (route) => false);
        },
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            color: isSelected ? accent.withValues(alpha: 0.08) : global.theme.cardColor,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: isSelected ? accent.withValues(alpha: 0.3) : Colors.transparent, width: 1.5),
            boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 6, offset: const Offset(0, 2))],
          ),
          child: Row(
            children: [
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(color: accent.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(14)),
                child: Icon(branchIcon, size: 24, color: accent),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      branchName,
                      style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: global.theme.textColor, height: 1.3),
                      overflow: TextOverflow.ellipsis,
                      maxLines: 2,
                    ),
                    if (isSelected) ...[
                      const SizedBox(height: 4),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                        decoration: BoxDecoration(color: accent.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(6)),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(Icons.check_circle_rounded, size: 10, color: accent),
                            const SizedBox(width: 3),
                            Text(global.language("in_use"), style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: accent)),
                          ],
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              Icon(Icons.chevron_right_rounded, color: global.theme.iconSecondaryColor, size: 22),
            ],
          ),
        ),
      ),
    );
  }

  Widget loginWithGoogle() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 25),
      child: Container(
        margin: const EdgeInsets.only(left: 20, right: 20),
        child: SingInButton(
          labelText: 'Sign in with google',
          press: () {
            loginType = global.LoginEnum.google;
            if (kIsWeb) {
              googleSignIn().then((value) async {
                if (value != null) {
                  String? userIdToken = await getCurrentUserIdToken();
                  if (userIdToken != null) {
                    appConfig.setString("user", value.user!.email ?? "");
                    if (mounted) {
                      context.read<LoginBloc>().add(TokenLogin(token: userIdToken));
                    }
                  }
                }
              });
            } else if (Platform.isWindows || Platform.isMacOS) {
              // ใช้ GoogleAuthHelper สำหรับ Windows/MacOS
              // รองรับ OAuth2 flow กับ localhost redirect
              GoogleAuthHelper googleAuthHelper = GoogleAuthHelper();
              googleAuthHelper.signIn().then((value) async {
                if (value != null) {
                  String? userIdToken = await getCurrentUserIdToken();
                  if (userIdToken != null) {
                    appConfig.setString("user", value.user!.email ?? "");
                    if (mounted) {
                      context.read<LoginBloc>().add(TokenLogin(token: userIdToken));
                    }
                  }
                }
              });
            }
          },
          img: const AssetImage("assets/img/google_logo.png"),
        ),
      ),
    );
  }

  Widget registerPhone() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 25),
      child: Container(
        margin: const EdgeInsets.only(left: 20, right: 20),
        child: SingInButton(
          labelText: 'Register With Phone Number',
          press: () {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const TelephoneScreen()));
          },
          img: const AssetImage("assets/img/logo_phone.png"),
        ),
      ),
    );
  }

  Widget imageLogo() {
    // return Image.asset("assets/img/sml-merchant-icon.png", height: 240);
    return const Text(
      "BC",
      style: TextStyle(fontSize: 60, fontWeight: FontWeight.bold, color: Color(0xFF354A5F)),
    );
  }

  void showSncakBar(BuildContext context, String message) {
    global.showInfoSnackBar(context, message);
  }

  Widget listBranchSelect() {
    // Filter branches whenever data changes or search/filter changes
    _filterBranches(companyBranchListData);

    return Column(
      children: [
        // Header with title and back button
        Padding(
          padding: EdgeInsets.fromLTRB(20, 24, 20, 8),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(color: F.primaryGradientColor.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(12)),
                child: Icon(Icons.home_work_rounded, color: F.primaryGradientColor, size: 24),
              ),
              SizedBox(width: 12),
              Expanded(
                child: Text(
                  global.language('select_branch'),
                  style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Color(0xFF1a1a2e)),
                ),
              ),
              Material(
                color: Colors.transparent,
                child: InkWell(
                  onTap: () {
                    setState(() {
                      stateScreen = 1;
                    });
                  },
                  borderRadius: BorderRadius.circular(12),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    decoration: BoxDecoration(color: F.primaryGradientColor.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(12)),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.swap_horiz_rounded, size: 18, color: F.primaryGradientColor),
                        SizedBox(width: 6),
                        Text(
                          global.language('select_shop'),
                          style: TextStyle(color: F.primaryGradientColor, fontWeight: FontWeight.w600, fontSize: 13),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
        // Search Bar
        _buildBranchSearchBar(),
        const SizedBox(height: 16),
        // Branch List or Empty State
        Expanded(
          child: filteredBranchList.isEmpty
              ? _buildBranchEmptyState()
              : ListView.separated(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
                  itemCount: filteredBranchList.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 8),
                  itemBuilder: (context, index) => cardItemBranch(filteredBranchList[index], index),
                ),
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: [
        BlocListener<LoginBloc, LoginState>(
          listener: (context, state) {
            /// Logout
            if (state is LogoutSuccess) {
              Navigator.pushNamedAndRemoveUntil(context, '/login_screen', (route) => false);
            }

            if (state is CreateShopSuccess) {
              setState(() => _isCreatingShop = false);
              context.read<ListShopBloc>().add(ListShopLoad());
            }

            if (state is CreateShopFailed) {
              setState(() => _isCreatingShop = false);
            }
          },
        ),
        BlocListener<ProfileBloc, ProfileState>(
          listener: (context, state) {
            if (state is GetProfileSuccess) {
              global.profileData = state.profile;
            }
          },
        ),
        BlocListener<ListShopBloc, ListShopState>(
          listener: (context, state) {
            if (state is ListShopLoadSuccess) {
              setState(() {
                stateScreen = 1;
              });
            }
          },
        ),
        BlocListener<ShopSelectBloc, ShopSelectState>(
          listener: (context, state) {
            if (state is ShopSelectLoadSuccess) {
              /// load ข้อมูล ร้านค้า

              context.read<ShopBloc>().add(GetShopInfo(shopid: global.getShopId()));

              /// load ข้อมูลสาขา
              context.read<CompanyBranchBloc>().add(const CompanyBranchLoadList(offset: 0, limit: 100, search: ""));

              try {
                String userJsonString = appConfig.getString("user")!;
                String userEmail = "";

                // Check if userJsonString is a simple email string or JSON object
                if (userJsonString.startsWith("{") && userJsonString.endsWith("}")) {
                  // It's a JSON object
                  Map<String, dynamic> userJson = jsonDecode(userJsonString);
                  userEmail = userJson['email'] ?? "";
                } else {
                  // It's a simple email string
                  userEmail = userJsonString;
                }

                if (userEmail.isNotEmpty) {
                  context.read<UserBloc>().add(UserGet(username: userEmail));
                }
              } catch (e) {
                if (kDebugMode) {
                  AppLogger.error('Error parsing user data: $e');
                }
                // Handle error - maybe show a snackbar or log the error
                global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), 'Error parsing user data', Colors.red);
              }

              /// ข้ามการเลือกสาขา

              // Navigator.of(context).pushAndRemoveUntil(
              //     MaterialPageRoute(builder: (_) => const MenuScreen()),
              //     (route) => false);
            } else if (state is ShopSelectLoadFailed) {
              global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), state.message, Colors.red);
            }
          },
        ),
        BlocListener<ShopBloc, ShopState>(
          listener: (context, state) {
            if (state is GetShopInfoSuccess) {
              appConfig.setString("shop_info", jsonEncode(state.shop.toJson()));
              if (state.shop.ismainshop == false && state.shop.mainshopid != null && state.shop.mainshopid!.isNotEmpty) {
                context.read<ShopBloc>().add(GetMainShopCenterTypes(mainShopId: state.shop.mainshopid!));
              } else {
                if (kDebugMode) {
                  AppLogger.debug("Conditions not met for GetMainShopCenterTypes");
                }
              }
            }

            if (state is GetMainShopCenterTypesInProgress) {
              if (kDebugMode) {
                AppLogger.debug("GetMainShopCenterTypesInProgress");
              }
            }

            if (state is GetMainShopCenterTypesSuccess) {
              if (kDebugMode) {
                AppLogger.debug(
                  "GetMainShopCenterTypesSuccess - productCenterType: ${state.productCenterType}, debtorCenterType: ${state.debtorCenterType}, posProductCenterType: ${state.posProductCenterType}",
                );
              }
              // Update shop_info with center types from main shop using ShopModel
              String? shopInfoJson = appConfig.getString("shop_info");
              if (shopInfoJson != null && shopInfoJson.isNotEmpty) {
                try {
                  ShopModel shopModel = ShopModel.fromJson(jsonDecode(shopInfoJson));
                  shopModel = ShopModel(
                    guidfixed: shopModel.guidfixed,
                    address: shopModel.address,
                    branchcode: shopModel.branchcode,
                    images: shopModel.images,
                    logo: shopModel.logo,
                    name1: shopModel.name1,
                    names: shopModel.names,
                    profilepicture: shopModel.profilepicture,
                    settings: shopModel.settings,
                    telephone: shopModel.telephone,
                    ismainshop: shopModel.ismainshop,
                    productcentertype: state.productCenterType,
                    debtorcentertype: state.debtorCenterType,
                    posproductcentertype: state.posProductCenterType,
                    mainshopid: shopModel.mainshopid,
                  );

                  appConfig.remove("shop_info");
                  appConfig.setString("shop_info", jsonEncode(shopModel.toJson()));
                  if (kDebugMode) {
                    AppLogger.debug("Updated shop_info with center types");
                  }
                } catch (e) {
                  if (kDebugMode) {
                    AppLogger.error("Error updating shop_info with center types: $e");
                  }
                }
              }
            }

            if (state is GetMainShopCenterTypesFailed) {
              if (kDebugMode) {
                AppLogger.error("GetMainShopCenterTypesFailed: ${state.message}");
              }
            }
          },
        ),
        BlocListener<CompanyBranchBloc, CompanyBranchState>(
          listener: (context, state) async {
            if (state is CompanyBranchLoadSuccess) {
              var branches = state.companyBranch;

              // ถ้าไม่มีสาขาเลย → สร้างสาขาเริ่มต้น "00000 สำนักงานใหญ่"
              if (branches.isEmpty) {
                try {
                  final defaultBranch = CompanyBranchModel(
                    guidfixed: '',
                    code: '00000',
                    names: [LanguageDataModel(code: 'th', name: 'สำนักงานใหญ่')],
                  );
                  final repo = CompanyBranchRepository();
                  final result = await repo.saveBranch(defaultBranch);
                  if (result.success && result.data != null) {
                    final created = CompanyBranchModel.fromJson(result.data);
                    branches = [created];
                    if (kDebugMode) {
                      AppLogger.info('[Branch] สร้างสาขาเริ่มต้น 00000 สำนักงานใหญ่ สำเร็จ');
                    }
                  }
                } catch (e) {
                  if (kDebugMode) {
                    AppLogger.error('[Branch] สร้างสาขาเริ่มต้นไม่สำเร็จ: $e');
                  }
                }
              }

              if (branches.length <= 1) {
                // ไม่มีสาขา หรือมีสาขาเดียว → ข้ามไปเมนูเลย
                if (branches.isNotEmpty) {
                  appConfig.setString("branch_guidfixed", branches[0].guidfixed);
                  global.companyBranchSelectData = branches[0];
                }
                appConfig.setInt("branch_total", branches.length);

                // ตั้งค่าเริ่มต้นร้านค้า (ชื่อ, ภาษา, สกุลเงิน) ถ้ายังไม่มี
                if (!mounted) return;
                await autoSetupShopDefaults();

                // ตรวจสอบสกุลเงิน — ถ้าไม่มีให้บังคับเลือกก่อนเข้า main menu
                if (!mounted) return;
                await checkAndSetupCurrency(context);

                if (!mounted) return;
                Navigator.of(context).pushAndRemoveUntil(MaterialPageRoute(builder: (_) => TabbedMenuScreen(key: TabbedMenuScreen.globalKey)), (route) => false);
              } else {
                // มีหลายสาขา → แสดงหน้าเลือกสาขา
                setState(() {
                  companyBranchListData.clear();
                  companyBranchListData.addAll(state.companyBranch);
                  stateScreen = 3;
                });
              }
            }

            if (state is CompanyBranchLoadFailed) {
              global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), state.message, Colors.red);
            }
          },
        ),
        BlocListener<UserBloc, UserState>(
          listener: (context, state) {
            if (state is UserGetSuccess) {
              // save to local
              appConfig.setString("role", state.user.role.toString());
            } else if (state is UserGetFailed) {
              // show error message
              global.showErrorSnackBar(context, "Error: ${state.message}");
            }
          },
        ),
      ],
      child: Scaffold(
        body: Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: global.isDarkMode()
                  ? [const Color(0xFF0D1B2A), const Color(0xFF16213E)]
                  : [F.primaryGradientColor, F.secondaryGradientColor],
            ),
          ),
          child: SafeArea(
            child: Column(
              children: [
                // Custom App Bar
                _buildModernAppBar(),
                // Main Content
                Expanded(
                  child: Container(
                    margin: const EdgeInsets.only(top: 16),
                    decoration: BoxDecoration(
                      color: global.isDarkMode() ? global.theme.backgroundColor : const Color(0xFFF8FAFC),
                      borderRadius: const BorderRadius.only(topLeft: Radius.circular(32), topRight: Radius.circular(32)),
                    ),
                    child: ClipRRect(
                      borderRadius: const BorderRadius.only(topLeft: Radius.circular(32), topRight: Radius.circular(32)),
                      child: (stateScreen == 1)
                          ? listShopSelect()
                          : (stateScreen == 2)
                          ? createShop()
                          : listBranchSelect(),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
