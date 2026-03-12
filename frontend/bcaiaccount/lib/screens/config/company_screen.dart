import 'package:flutter/material.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/shop/shop_bloc.dart';
import 'package:smlaicloud/model/create_shop_model.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/shop_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/global.dart' as global;

class CompanyScreen extends StatefulWidget {
  const CompanyScreen({super.key});

  @override
  State<CompanyScreen> createState() => _CompanyScreenState();
}

class _CompanyScreenState extends State<CompanyScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  late ShopModel screenData;
  final List<LanguageModel> defaultLanguageList = [
    LanguageModel(code: "th", codeTranslator: "th", name: "Thai", isuse: false),
    LanguageModel(
      code: "en",
      codeTranslator: "en",
      name: "English",
      isuse: false,
    ),
    LanguageModel(
      code: "cn",
      codeTranslator: "cn",
      name: "Chinese",
      isuse: false,
    ),
    LanguageModel(
      code: "ja",
      codeTranslator: "ja",
      name: "Japanese",
      isuse: false,
    ),
    LanguageModel(
      code: "ko",
      codeTranslator: "ko",
      name: "Korean",
      isuse: false,
    ),
    LanguageModel(code: "lo", codeTranslator: "lo", name: "Lao", isuse: false),
    LanguageModel(
      code: "my",
      codeTranslator: "my",
      name: "Burmese",
      isuse: false,
    ),
    LanguageModel(
      code: "ms",
      codeTranslator: "ms",
      name: "Malaysian",
      isuse: false,
    ),
    LanguageModel(
      code: "vi",
      codeTranslator: "vi",
      name: "Vietnamese",
      isuse: false,
    ),
    LanguageModel(
      code: "km",
      codeTranslator: "km",
      name: "Khmer",
      isuse: false,
    ),
  ];

  bool isSaving = false;
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();

  // Tab controller for switching between Company Data and Languages
  late TabController _tabController;
  bool isExpanded = false;

  // สกุลเงินหลัก (Base Currency)
  final CurrencyApiService _currencyApiService = CurrencyApiService();
  List<CurrencyModel> _currencies = [];
  bool _isLoadingCurrencies = false;

  @override
  void initState() {
    super.initState();
    screenData = ShopModel();
    loadData();
    _loadCurrencies();
    _tabController = TabController(length: 2, vsync: this);
  }

  /// โหลดรายการสกุลเงินจาก API
  Future<void> _loadCurrencies() async {
    if (_isLoadingCurrencies) return;
    setState(() => _isLoadingCurrencies = true);
    try {
      final String shopId = global.prefs.getString("shopid") ?? "";
      final response = await _currencyApiService.getCurrencies(
        shopId: shopId,
        disabled: false,
        limit: 100,
      );
      if (response.success && response.data != null) {
        setState(() {
          _currencies = (response.data as List)
              .map((e) => CurrencyModel.fromJson(e))
              .toList();
        });
      }
    } catch (e) {
      debugPrint('Error loading currencies: $e');
      // ถ้าโหลดไม่ได้ ใช้ค่า default THB
      if (_currencies.isEmpty) {
        setState(() {
          _currencies = [CurrencyModel(code: 'THB', name: 'บาท', symbol: '฿')];
        });
      }
    } finally {
      setState(() => _isLoadingCurrencies = false);
    }
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void loadData() async {
    await global.setSystemLanguage(context);
    setState(() {
      screenData = global.shopSelectData;
      // Initialize language configurations if needed
      screenData.settings ??= Settings();
      screenData.settings!.languageconfigs ??= [];
      screenData.names ??= [];

      // ถ้า languageconfigs ว่าง ให้เพิ่ม Thai เป็นภาษาเริ่มต้น
      // เพื่อให้มี TextField สำหรับกรอกชื่อบริษัทได้
      if (screenData.settings!.languageconfigs!.isEmpty) {
        screenData.settings!.languageconfigs = [
          LanguageModel(
            code: 'th',
            codeTranslator: 'th',
            name: 'Thai',
            isuse: true,
            isdefault: true,
          ),
        ];
      }

      // ถ้า names ว่าง ให้เพิ่ม entry สำหรับภาษาเริ่มต้น
      if (screenData.names!.isEmpty) {
        for (var lang in screenData.settings!.languageconfigs!) {
          if (lang.code != null && lang.code!.isNotEmpty) {
            screenData.names!.add(
              LanguageDataModel(code: lang.code!, name: ''),
            );
          }
        }
      }
    });
  }

  void saveOrUpdateData() {
    if (!_formKey.currentState!.validate()) {
      showSnackBar(
        global.language("please_fill_required_fields"),
        isError: true,
      );
      return;
    }

    setState(() => isSaving = true);

    // Set the first language as default
    if (screenData.settings!.languageconfigs!.isNotEmpty) {
      screenData.settings!.languageconfigs![0].isdefault = true;
    }

    // ใช้ global.getShopId() เป็น fallback กรณี guidfixed เป็นค่าว่าง
    final shopId =
        (screenData.guidfixed != null && screenData.guidfixed!.isNotEmpty)
        ? screenData.guidfixed!
        : global.getShopId();

    if (shopId.isEmpty) {
      showSnackBar('Shop ID not found', isError: true);
      return;
    }

    context.read<ShopBloc>().add(
      ShopUpdate(shopid: shopId, shopdata: screenData),
    );
  }

  void showSnackBar(String message, {bool isError = false}) {
    if (!mounted) return;
    final cleanMessage = message.replaceAll('//', '');
    if (isError) {
      global.showErrorSnackBar(context, cleanMessage);
    } else {
      global.showSuccessSnackBar(context, cleanMessage);
    }
  }

  @override
  Widget build(BuildContext context) {
    // Get screen width for responsive design
    final screenWidth = MediaQuery.of(context).size.width;
    final isSmallScreen = screenWidth < 600;

    return Scaffold(
      appBar: AppBar(
        title: Text(
          global.language('company'),
          style: TextStyle(fontWeight: FontWeight.w600),
        ),
        elevation: 0,
        leading: IconButton(
          icon: Icon(Icons.arrow_back),
          onPressed: () => Navigator.pop(context),
          tooltip: global.language('back'),
        ),
        actions: [
          isSaving
              ? Padding(
                  padding: EdgeInsets.symmetric(horizontal: 16),
                  child: SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      color: global.theme.cardColor,
                      strokeWidth: 2,
                    ),
                  ),
                )
              : IconButton(
                  tooltip: global.language('save'),
                  onPressed: saveOrUpdateData,
                  icon: Icon(Icons.save_outlined),
                ),
          const ManualButton(path: 'settings-company'),
        ],
        bottom: TabBar(
          controller: _tabController,
          indicatorWeight: 3,
          labelStyle: TextStyle(fontWeight: FontWeight.w600),
          tabs: [
            Tab(
              icon: Icon(Icons.business_outlined),
              text: global.language("company_data"),
            ),
            Tab(
              icon: Icon(Icons.translate_outlined),
              text: global.language("languages"),
            ),
          ],
        ),
      ),
      body: MultiBlocListener(
        listeners: [
          BlocListener<ShopBloc, ShopState>(
            listener: (context, state) {
              setState(() => isSaving = false);

              if (state is ShopUpdateSuccess) {
                showSnackBar(global.language("update_success"));
                Navigator.pop(context);
              } else if (state is ShopUpdateFailed) {
                showSnackBar(state.message, isError: true);
              }
            },
          ),
        ],
        child: Form(
          key: _formKey,
          child: TabBarView(
            controller: _tabController,
            children: [
              _buildCompanyDataTab(isSmallScreen),
              _buildLanguagesTab(isSmallScreen),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCompanyDataTab(bool isSmallScreen) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Center(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Company Names Card
            Card(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              elevation: 2,
              surfaceTintColor: global.theme.cardColor,
              child: Padding(
                padding: EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Theme.of(
                              context,
                            ).primaryColor.withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Icon(
                            Icons.business_outlined,
                            size: 22,
                            color: Theme.of(context).primaryColor,
                          ),
                        ),
                        SizedBox(width: 12),
                        Text(
                          global.language("company_names"),
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 18,
                          ),
                        ),
                      ],
                    ),
                    const Divider(height: 32),
                    _buildCompanyNameFields(),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Base Currency Card
            _buildBaseCurrencySelector(),
          ],
        ),
      ),
    );
  }

  Widget _buildLanguagesTab(bool isSmallScreen) {
    return SingleChildScrollView(
      padding: EdgeInsets.all(16),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 800),
          child: Card(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            elevation: 2,
            surfaceTintColor: global.theme.cardColor,
            child: Padding(
              padding: EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: global.theme.infoHighlightColor,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Icon(
                              Icons.translate_outlined,
                              size: 22,
                              color: global.theme.infoHighlightTextColor,
                            ),
                          ),
                          SizedBox(width: 12),
                          Text(
                            global.language("language_settings"),
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 18,
                            ),
                          ),
                        ],
                      ),
                      ElevatedButton.icon(
                        onPressed: _addLanguage,
                        icon: Icon(Icons.add, size: 18),
                        label: Text(global.language('add_language')),
                        style: ElevatedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 10,
                          ),
                          elevation: 0,
                        ),
                      ),
                    ],
                  ),
                  const Divider(height: 32),
                  if (screenData.settings?.languageconfigs == null ||
                      screenData.settings!.languageconfigs!.isEmpty)
                    Center(
                      child: Padding(
                        padding: EdgeInsets.symmetric(vertical: 40),
                        child: Column(
                          children: [
                            Container(
                              padding: EdgeInsets.all(16),
                              decoration: BoxDecoration(
                                color: global.theme.infoHighlightColor,
                                shape: BoxShape.circle,
                              ),
                              child: Icon(
                                Icons.language_outlined,
                                size: 64,
                                color: global.theme.infoHighlightTextColor,
                              ),
                            ),
                            SizedBox(height: 24),
                            Text(
                              global.language('no_languages_configured'),
                              style: TextStyle(
                                color: global.theme.textColor,
                                fontSize: 16,
                              ),
                            ),
                            SizedBox(height: 24),
                            ElevatedButton.icon(
                              onPressed: _addLanguage,
                              icon: Icon(Icons.add, size: 18),
                              label: Text(global.language('add_language')),
                              style: ElevatedButton.styleFrom(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                  vertical: 12,
                                ),
                                minimumSize: const Size(180, 48),
                              ),
                            ),
                          ],
                        ),
                      ),
                    )
                  else
                    _buildLanguageList(),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildCompanyNameFields() {
    List<Widget> defaultLanguageField = [];
    List<Widget> additionalLanguageFields = [];

    if (screenData.settings?.languageconfigs != null) {
      // First ensure all languages have corresponding entries in names
      for (var language in screenData.settings!.languageconfigs!) {
        if (language.code != null && language.code!.isNotEmpty) {
          screenData.names ??= [];

          final nameIndex = screenData.names!.indexWhere(
            (element) => element.code == language.code,
          );

          if (nameIndex < 0) {
            screenData.names!.add(
              LanguageDataModel(code: language.code!, name: ''),
            );
          }
        }
      }

      // Then create form fields for each language
      for (var i = 0; i < screenData.settings!.languageconfigs!.length; i++) {
        var language = screenData.settings!.languageconfigs![i];
        if (language.code != null && language.code!.isNotEmpty) {
          final nameIndex = screenData.names!.indexWhere(
            (element) => element.code == language.code,
          );

          if (nameIndex >= 0) {
            final isDefaultLanguage = i == 0;

            Widget field = Padding(
              padding: const EdgeInsets.only(bottom: 20),
              child: _buildTextField(
                language.code!,
                isDefaultLanguage,
                nameIndex,
              ),
            );

            if (isDefaultLanguage) {
              defaultLanguageField.add(field);
            } else {
              additionalLanguageFields.add(field);
            }
          }
        }
      }
    }

    // Return only main language if there are no additional languages
    if (additionalLanguageFields.isEmpty) {
      return Column(children: defaultLanguageField);
    }

    // Return main language and collapsible additional languages
    return Column(
      children: [
        ...defaultLanguageField,
        const SizedBox(height: 8),
        if (additionalLanguageFields.isNotEmpty)
          Card(
            elevation: 0,
            color: global.theme.surfaceColor,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
              side: BorderSide(color: global.theme.dividerBorderColor),
            ),
            child: ExpansionTile(
              shape: Border(),
              title: Text(
                global.language('additional_language'),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
              ),
              leading: Icon(
                Icons.translate,
                color: global.theme.textColor,
                size: 20,
              ),
              tilePadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 8,
              ),
              childrenPadding: const EdgeInsets.only(
                left: 16,
                right: 16,
                bottom: 16,
                top: 0,
              ),
              expandedCrossAxisAlignment: CrossAxisAlignment.start,
              children: additionalLanguageFields,
            ),
          ),
      ],
    );
  }

  Widget _buildTextField(String languageCode, bool isRequired, int nameIndex) {
    final languageName = defaultLanguageList
        .firstWhere(
          (element) => element.code == languageCode,
          orElse: () => LanguageModel(
            code: '',
            name: '',
            codeTranslator: '',
            isuse: false,
          ),
        )
        .name;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Label outside the TextField
        Row(
          children: [
            Container(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(4),
                boxShadow: [
                  BoxShadow(
                    color: global.theme.dividerBorderColor.withValues(alpha: 0.2),
                    blurRadius: 3,
                    offset: const Offset(0, 1),
                  ),
                ],
              ),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(4),
                child: Image.asset(
                  'assets/flags/$languageCode.png',
                  width: 24,
                  height: 18,
                ),
              ),
            ),
            const SizedBox(width: 10),
            Text(
              "$languageName${isRequired ? ' *' : ''}",
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
                color: isRequired
                    ? Theme.of(context).primaryColor
                    : global.theme.textColor,
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        // TextField
        TextFormField(
          initialValue: screenData.names![nameIndex].name,
          decoration: InputDecoration(
            hintText: isRequired
                ? global.language('required')
                : global.language('optional'),
            contentPadding: EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 16,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(color: global.theme.dividerBorderColor),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(color: global.theme.dividerBorderColor),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: Theme.of(context).primaryColor,
                width: 2,
              ),
            ),
            fillColor: isRequired
                ? Theme.of(context).primaryColor.withValues(alpha: 0.05)
                : global.theme.onPrimaryColor,
            filled: true,
          ),
          onChanged: (value) {
            screenData.names![nameIndex].name = value;
          },
          validator: isRequired
              ? (value) {
                  if (value == null || value.isEmpty) {
                    return global.language('required_field');
                  }
                  return null;
                }
              : null,
        ),
      ],
    );
  }

  Widget _buildLanguageList() {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      clipBehavior: Clip.antiAlias,
      child: ListView.separated(
        physics: const NeverScrollableScrollPhysics(),
        shrinkWrap: true,
        itemCount: screenData.settings!.languageconfigs!.length,
        separatorBuilder: (_, _) =>
            Divider(height: 1, color: global.theme.dividerBorderColor),
        itemBuilder: (context, index) {
          final language = screenData.settings!.languageconfigs![index];
          final isDefault = index == 0;

          return Container(
            color: isDefault ? global.theme.infoHighlightColor : null,
            child: ListTile(
              leading: Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: isDefault ? global.theme.infoHighlightTextColor : global.theme.dividerBorderColor,
                  borderRadius: BorderRadius.circular(20),
                ),
                alignment: Alignment.center,
                child: Text(
                  (index + 1).toString(),
                  style: TextStyle(
                    color: isDefault ? global.theme.onPrimaryColor : global.theme.textColor,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              title: Row(
                children: [
                  if (language.code != null && language.code!.isNotEmpty) ...[
                    Container(
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(4),
                        boxShadow: [
                          BoxShadow(
                            color: global.theme.dividerBorderColor.withValues(alpha: 0.2),
                            blurRadius: 3,
                            offset: const Offset(0, 1),
                          ),
                        ],
                      ),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(4),
                        child: Image.asset(
                          'assets/flags/${language.code}.png',
                          width: 28,
                          height: 20,
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                  ],
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        (language.code != null && language.code!.isNotEmpty)
                            ? _getLanguageName(language.code!)
                            : global.language('select_language'),
                        style: TextStyle(
                          fontWeight: isDefault
                              ? FontWeight.bold
                              : FontWeight.normal,
                          fontSize: 15,
                        ),
                      ),
                      if (isDefault)
                        Padding(
                          padding: EdgeInsets.only(top: 4),
                          child: Container(
                            padding: EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: global.theme.infoHighlightColor,
                              borderRadius: BorderRadius.circular(4),
                              border: Border.all(color: global.theme.infoHighlightTextColor.withValues(alpha: 0.3)),
                            ),
                            child: Text(
                              global.language('default_language'),
                              style: TextStyle(
                                fontSize: 12,
                                color: global.theme.infoHighlightTextColor,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ),
                        ),
                    ],
                  ),
                ],
              ),
              trailing: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IconButton(
                    icon: Icon(Icons.edit_outlined, size: 22),
                    onPressed: () => _selectLanguage(index),
                    tooltip: global.language('change'),
                    style: IconButton.styleFrom(
                      backgroundColor: global.theme.infoHighlightColor,
                      foregroundColor: global.theme.infoHighlightTextColor,
                    ),
                  ),
                  const SizedBox(width: 8),
                  if (!isDefault)
                    IconButton(
                      icon: Icon(Icons.delete_outline, size: 22),
                      onPressed: () => _deleteLanguage(index),
                      tooltip: global.language('delete'),
                      style: IconButton.styleFrom(
                        backgroundColor: global.theme.negativeHighlightColor,
                        foregroundColor: global.theme.negativeHighlightTextColor,
                      ),
                    ),
                ],
              ),
              onTap: () => _selectLanguage(index),
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 8,
              ),
            ),
          );
        },
      ),
    );
  }

  void _selectLanguage(int index) async {
    // Create a list of available languages that haven't been selected yet
    List<LanguageModel> availableLanguages = [];
    availableLanguages.addAll(defaultLanguageList);

    // Remove languages that are already in use
    for (var lang in screenData.settings!.languageconfigs!) {
      if (lang.code != null && lang.code!.isNotEmpty) {
        availableLanguages.removeWhere((element) => element.code == lang.code);
      }
    }

    // Add back the current language if it's being changed
    final currentLang = screenData.settings!.languageconfigs![index];
    if (currentLang.code != null && currentLang.code!.isNotEmpty) {
      final originalLang = defaultLanguageList.firstWhere(
        (element) => element.code == currentLang.code,
        orElse: () =>
            LanguageModel(code: "", codeTranslator: "", name: "", isuse: false),
      );

      if (originalLang.code!.isNotEmpty) {
        availableLanguages.add(originalLang);
        // Sort by name
        availableLanguages.sort((a, b) => a.name!.compareTo(b.name!));
      }
    }

    final langCode = await _showLanguageSelectionDialog(availableLanguages);

    if (langCode != null) {
      setState(() {
        screenData.settings!.languageconfigs![index].code = langCode;
        screenData.settings!.languageconfigs![index].codeTranslator = langCode;
        screenData.settings!.languageconfigs![index].name = defaultLanguageList
            .firstWhere((l) => l.code == langCode)
            .name;
        screenData.settings!.languageconfigs![index].isuse = true;

        // Ensure names list has an entry for this language
        if (screenData.names != null) {
          bool hasLanguage = screenData.names!.any(
            (element) => element.code == langCode,
          );

          if (!hasLanguage) {
            screenData.names!.add(LanguageDataModel(code: langCode, name: ''));
          }
        }
      });
    }
  }

  Future<String?> _showLanguageSelectionDialog(
    List<LanguageModel> languages,
  ) async {
    return showDialog<String>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(global.language('select_language')),
          contentPadding: EdgeInsets.only(top: 20),
          content: SizedBox(
            width: 320,
            height: 400,
            child: Column(
              children: [
                Padding(
                  padding: EdgeInsets.symmetric(horizontal: 24),
                  child: TextField(
                    decoration: InputDecoration(
                      hintText: global.language('search'),
                      prefixIcon: Icon(Icons.search_outlined),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(30),
                        borderSide: BorderSide.none,
                      ),
                      filled: true,
                      fillColor: global.theme.surfaceColor,
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 12,
                      ),
                    ),
                    // Optional: Add search functionality
                    onChanged: (value) {
                      // Filter languages based on search
                    },
                  ),
                ),
                const SizedBox(height: 16),
                Expanded(
                  child: ListView.builder(
                    itemCount: languages.length,
                    itemBuilder: (context, index) {
                      return ListTile(
                        leading: Container(
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(4),
                            boxShadow: [
                              BoxShadow(
                                color: global.theme.dividerBorderColor.withValues(alpha: 0.2),
                                blurRadius: 3,
                                offset: const Offset(0, 1),
                              ),
                            ],
                          ),
                          child: ClipRRect(
                            borderRadius: BorderRadius.circular(4),
                            child: Image.asset(
                              'assets/flags/${languages[index].code}.png',
                              width: 32,
                              height: 24,
                            ),
                          ),
                        ),
                        title: Text(
                          languages[index].name!,
                          style: TextStyle(fontSize: 15),
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(4),
                        ),
                        onTap: () {
                          Navigator.pop(context, languages[index].code);
                        },
                        hoverColor: global.theme.surfaceColor,
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('cancel')),
            ),
          ],
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        );
      },
    );
  }

  void _deleteLanguage(int index) async {
    final confirmed = await _showConfirmationDialog(
      global.language('confirm_delete'),
      global.language('delete_language_confirmation'),
    );

    if (confirmed) {
      setState(() {
        screenData.settings!.languageconfigs!.removeAt(index);
      });
    }
  }

  Future<bool> _showConfirmationDialog(String title, String content) async {
    final result = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(title),
          content: Text(content),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: Text(global.language('cancel')),
            ),
            ElevatedButton.icon(
              onPressed: () => Navigator.of(context).pop(true),
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.negativeHighlightColor,
                foregroundColor: global.theme.negativeHighlightTextColor,
              ),
              icon: Icon(Icons.delete_outline, size: 18),
              label: Text(global.language('delete')),
            ),
          ],
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          actionsPadding: const EdgeInsets.all(16),
        );
      },
    );
    return result ?? false;
  }

  void _addLanguage() {
    setState(() {
      screenData.settings!.languageconfigs!.add(
        LanguageModel(code: "", codeTranslator: "", name: "", isuse: false),
      );
    });

    // If we're on the company tab, switch to language tab
    if (_tabController.index == 0) {
      _tabController.animateTo(1);
    }

    // Add a slight delay before showing language selection dialog
    Future.delayed(const Duration(milliseconds: 300), () {
      _selectLanguage(screenData.settings!.languageconfigs!.length - 1);
    });
  }

  String _getLanguageName(String code) {
    LanguageModel language = defaultLanguageList.firstWhere(
      (element) => element.code == code,
      orElse: () =>
          LanguageModel(code: '', name: '', codeTranslator: '', isuse: false),
    );
    return language.name ?? '';
  }

  /// Widget สำหรับเลือกสกุลเงินหลัก (Base Currency)
  /// Base Currency มีค่าอัตราแลกเปลี่ยน 1:1 เสมอ
  Widget _buildBaseCurrencySelector() {
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      elevation: 2,
      surfaceTintColor: global.theme.cardColor,
      child: Padding(
        padding: EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: global.theme.positiveHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(
                    Icons.currency_exchange,
                    size: 22,
                    color: global.theme.positiveHighlightTextColor,
                  ),
                ),
                SizedBox(width: 12),
                Text(
                  global.language("base_currency"),
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 18,
                  ),
                ),
              ],
            ),
            const Divider(height: 32),
            // คำอธิบาย
            Container(
              padding: EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: global.theme.infoHighlightColor,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: global.theme.infoHighlightColor),
              ),
              child: Row(
                children: [
                  Icon(
                    Icons.info_outline,
                    color: global.theme.infoHighlightTextColor,
                    size: 20,
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      'สกุลเงินหลักของบริษัท มีอัตราแลกเปลี่ยน 1:1',
                      style: TextStyle(
                        color: global.theme.infoHighlightTextColor,
                        fontSize: 13,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            // Dropdown เลือกสกุลเงินหลัก
            if (_isLoadingCurrencies)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: CircularProgressIndicator(),
                ),
              )
            else if (_currencies.isEmpty)
              Container(
                padding: EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: global.theme.warningHighlightColor,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: global.theme.warningHighlightColor),
                ),
                child: Row(
                  children: [
                    Icon(Icons.warning_amber, color: global.theme.warningHighlightTextColor),
                    SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        'ยังไม่มีสกุลเงินในระบบ กรุณาเพิ่มสกุลเงินในเมนู global.language("currency") ก่อน',
                        style: TextStyle(color: global.theme.warningHighlightTextColor),
                      ),
                    ),
                  ],
                ),
              )
            else
              DropdownButtonFormField<String>(
                initialValue:
                    _currencies.any(
                      (c) => c.code == screenData.settings?.baseCurrency,
                    )
                    ? screenData.settings?.baseCurrency
                    : null,
                decoration: InputDecoration(
                  labelText: global.language("base_currency"),
                  hintText: global.language('select_base_currency'),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  prefixIcon: Icon(Icons.monetization_on_outlined),
                  filled: true,
                  fillColor: global.theme.formFillColor,
                ),
                items: _currencies.map((currency) {
                  return DropdownMenuItem<String>(
                    value: currency.code,
                    child: Text(
                      '${currency.symbol.isNotEmpty ? currency.symbol : currency.code} ${currency.code} - ${currency.name}',
                      style: TextStyle(fontSize: 14),
                    ),
                  );
                }).toList(),
                onChanged: (value) {
                  setState(() {
                    screenData.settings ??= Settings();
                    screenData.settings!.baseCurrency = value;
                  });
                },
              ),
          ],
        ),
      ),
    );
  }
}
