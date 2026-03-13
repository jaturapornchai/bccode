import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/create_shop_model.dart';
import 'package:smlaicloud/model/shop_model.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/timezones_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/company_branch_repository.dart';
import 'package:smlaicloud/repositories/shop_repository.dart';
import 'package:smlaicloud/repositories/business_type_repository.dart';
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

// ─── Preset Data ───

class PresetCurrency {
  final String code;
  final String name;
  final String symbol;

  const PresetCurrency({
    required this.code,
    required this.name,
    required this.symbol,
  });
}

const List<PresetCurrency> presetCurrencies = [
  PresetCurrency(code: 'THB', name: 'Thai Baht', symbol: '฿'),
  PresetCurrency(code: 'USD', name: 'US Dollar', symbol: '\$'),
  PresetCurrency(code: 'EUR', name: 'Euro', symbol: '€'),
  PresetCurrency(code: 'JPY', name: 'Japanese Yen', symbol: '¥'),
  PresetCurrency(code: 'CNY', name: 'Chinese Yuan', symbol: '¥'),
  PresetCurrency(code: 'GBP', name: 'British Pound', symbol: '£'),
  PresetCurrency(code: 'KRW', name: 'Korean Won', symbol: '₩'),
  PresetCurrency(code: 'VND', name: 'Vietnamese Dong', symbol: '₫'),
  PresetCurrency(code: 'LAK', name: 'Lao Kip', symbol: '₭'),
  PresetCurrency(code: 'KHR', name: 'Cambodian Riel', symbol: '៛'),
  PresetCurrency(code: 'MMK', name: 'Myanmar Kyat', symbol: 'K'),
  PresetCurrency(code: 'MYR', name: 'Malaysian Ringgit', symbol: 'RM'),
  PresetCurrency(code: 'SGD', name: 'Singapore Dollar', symbol: 'S\$'),
  PresetCurrency(code: 'IDR', name: 'Indonesian Rupiah', symbol: 'Rp'),
  PresetCurrency(code: 'PHP', name: 'Philippine Peso', symbol: '₱'),
];

/// Mapping สกุลเงิน → ค่าแนะนำ (yeartype, timezone, language)
class CurrencyCountryDefaults {
  final String yeartype;
  final String timezone; // IANA timezone
  final String language;

  const CurrencyCountryDefaults({
    required this.yeartype,
    required this.timezone,
    required this.language,
  });
}

const Map<String, CurrencyCountryDefaults> currencyDefaults = {
  'THB': CurrencyCountryDefaults(yeartype: 'buddhist', timezone: 'Asia/Bangkok', language: 'th'),
  'USD': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'America/New_York', language: 'en'),
  'EUR': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Europe/London', language: 'en'),
  'JPY': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Asia/Tokyo', language: 'ja'),
  'CNY': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Asia/Shanghai', language: 'cn'),
  'GBP': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Europe/London', language: 'en'),
  'KRW': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Asia/Seoul', language: 'ko'),
  'VND': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Asia/Ho_Chi_Minh', language: 'vi'),
  'LAK': CurrencyCountryDefaults(yeartype: 'buddhist', timezone: 'Asia/Vientiane', language: 'lo'),
  'KHR': CurrencyCountryDefaults(yeartype: 'buddhist', timezone: 'Asia/Phnom_Penh', language: 'km'),
  'MMK': CurrencyCountryDefaults(yeartype: 'buddhist', timezone: 'Asia/Yangon', language: 'my'),
  'MYR': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Asia/Kuala_Lumpur', language: 'ms'),
  'SGD': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Asia/Singapore', language: 'en'),
  'IDR': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Asia/Jakarta', language: 'en'),
  'PHP': CurrencyCountryDefaults(yeartype: 'christian', timezone: 'Asia/Manila', language: 'en'),
};

/// รายการภาษาที่รองรับ
class LanguageOption {
  final String code;
  final String name;
  final String flag;

  const LanguageOption({required this.code, required this.name, required this.flag});
}

const List<LanguageOption> languageOptions = [
  LanguageOption(code: 'th', name: 'ไทย', flag: '🇹🇭'),
  LanguageOption(code: 'en', name: 'English', flag: '🇺🇸'),
  LanguageOption(code: 'cn', name: '中文', flag: '🇨🇳'),
  LanguageOption(code: 'ja', name: '日本語', flag: '🇯🇵'),
  LanguageOption(code: 'ko', name: '한국어', flag: '🇰🇷'),
  LanguageOption(code: 'lo', name: 'ລາວ', flag: '🇱🇦'),
  LanguageOption(code: 'my', name: 'မြန်မာ', flag: '🇲🇲'),
  LanguageOption(code: 'ms', name: 'Melayu', flag: '🇲🇾'),
  LanguageOption(code: 'vi', name: 'Tiếng Việt', flag: '🇻🇳'),
  LanguageOption(code: 'km', name: 'ខ្មែរ', flag: '🇰🇭'),
];

// ─── Local Cache for Branch Settings ───
// เก็บ settings ไว้ใน SharedPreferences ตาม branch guidfixed
// เป็น fallback เมื่อ backend ยังไม่ persist fields เหล่านี้

String _cacheKey(String branchGuid, String field) =>
    'branch_settings_${branchGuid}_$field';

/// บันทึก branch settings ลง local cache
Future<void> _saveBranchSettingsToCache({
  required String branchGuid,
  required String currencyCode,
  required String language,
  required String timezone,
  required String timezonelabel,
  required String timezoneoffset,
  required String yeartype,
}) async {
  if (branchGuid.isEmpty) return;
  await global.appConfig.setString(_cacheKey(branchGuid, 'base_currency'), currencyCode);
  await global.appConfig.setString(_cacheKey(branchGuid, 'language'), language);
  await global.appConfig.setString(_cacheKey(branchGuid, 'timezone'), timezone);
  await global.appConfig.setString(_cacheKey(branchGuid, 'timezonelabel'), timezonelabel);
  await global.appConfig.setString(_cacheKey(branchGuid, 'timezoneoffset'), timezoneoffset);
  await global.appConfig.setString(_cacheKey(branchGuid, 'yeartype'), yeartype);
  if (kDebugMode) {
    AppLogger.debug("[Branch Cache] saved settings for branch=$branchGuid");
  }
}

/// Helper: ดึงค่าจาก cache เฉพาะที่ไม่ว่าง
String? _getCachedNonEmpty(String branchGuid, String field) {
  final value = global.appConfig.getString(_cacheKey(branchGuid, field));
  return (value != null && value.isNotEmpty) ? value : null;
}

/// Normalize empty string → null
String? _nonEmpty(String? value) =>
    (value != null && value.isNotEmpty) ? value : null;

/// โหลด branch settings จาก local cache → fill เข้า branch object
/// สำคัญ: ต้อง normalize "" → null ก่อน แล้วค่อย fill จาก cache
void _loadBranchSettingsFromCache(String branchGuid) {
  if (branchGuid.isEmpty) return;
  final branch = global.companyBranchSelectData;

  // Step 1: Normalize empty strings → null (backend อาจส่ง "" กลับมา)
  branch.baseCurrency = _nonEmpty(branch.baseCurrency);
  branch.language = _nonEmpty(branch.language);
  branch.timezone = _nonEmpty(branch.timezone);
  branch.timezonelabel = _nonEmpty(branch.timezonelabel);
  branch.timezoneoffset = _nonEmpty(branch.timezoneoffset);
  branch.yeartype = _nonEmpty(branch.yeartype);

  // Step 2: Fill จาก cache (เฉพาะ field ที่ยังเป็น null)
  branch.baseCurrency ??= _getCachedNonEmpty(branchGuid, 'base_currency');
  branch.language ??= _getCachedNonEmpty(branchGuid, 'language');
  branch.timezone ??= _getCachedNonEmpty(branchGuid, 'timezone');
  branch.timezonelabel ??= _getCachedNonEmpty(branchGuid, 'timezonelabel');
  branch.timezoneoffset ??= _getCachedNonEmpty(branchGuid, 'timezoneoffset');
  branch.yeartype ??= _getCachedNonEmpty(branchGuid, 'yeartype');

  if (kDebugMode) {
    AppLogger.debug("[Branch Cache] loaded: currency=${branch.baseCurrency}, "
        "lang=${branch.language}, tz=${branch.timezone}, year=${branch.yeartype}");
  }
}

// ─── Main Check Function ───

/// ตรวจสอบ settings ของ branch หลังเลือก shop/branch
/// ถ้ามี field ว่าง → ลอง load จาก local cache → ถ้ายังว่าง แสดง dialog บังคับตั้งค่า
/// return true = ผ่าน (settings ครบ), false = ยังไม่ผ่าน
Future<bool> checkAndSetupBranchSettings(BuildContext context) async {
  try {
    final branch = global.companyBranchSelectData;
    final branchGuid = branch.guidfixed;

    // Step 1: Re-fetch branch จาก backend (GET by GUID → ข้อมูลครบกว่า list)
    if (branchGuid.isNotEmpty) {
      try {
        final repo = CompanyBranchRepository();
        final result = await repo.getBranch(branchGuid);
        if (result.success && result.data != null) {
          final fresh = CompanyBranchModel.fromJson(result.data);
          // Merge settings จาก backend (เฉพาะที่มีค่า)
          if (_nonEmpty(fresh.baseCurrency) != null) branch.baseCurrency = fresh.baseCurrency;
          if (_nonEmpty(fresh.language) != null) branch.language = fresh.language;
          if (_nonEmpty(fresh.timezone) != null) branch.timezone = fresh.timezone;
          if (_nonEmpty(fresh.timezonelabel) != null) branch.timezonelabel = fresh.timezonelabel;
          if (_nonEmpty(fresh.timezoneoffset) != null) branch.timezoneoffset = fresh.timezoneoffset;
          if (_nonEmpty(fresh.yeartype) != null) branch.yeartype = fresh.yeartype;
          if (kDebugMode) {
            AppLogger.debug("[Branch Settings] re-fetched from backend: "
                "currency=${fresh.baseCurrency}, lang=${fresh.language}, "
                "tz=${fresh.timezone}, year=${fresh.yeartype}");
          }
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.warning("[Branch Settings] re-fetch failed: $e");
        }
      }
    }

    // Step 2: ลองโหลดจาก local cache (fallback เมื่อ backend ยังไม่มีค่า)
    _loadBranchSettingsFromCache(branchGuid);

    final hasBaseCurrency = branch.baseCurrency != null && branch.baseCurrency!.isNotEmpty;
    final hasLanguage = branch.language != null && branch.language!.isNotEmpty;
    final hasTimezone = branch.timezone != null && branch.timezone!.isNotEmpty;
    final hasYeartype = branch.yeartype != null && branch.yeartype!.isNotEmpty;

    if (kDebugMode) {
      AppLogger.debug(
        "[Branch Settings Check] baseCurrency=${branch.baseCurrency}, "
        "language=${branch.language}, timezone=${branch.timezone}, "
        "yeartype=${branch.yeartype}",
      );
    }

    // ถ้าครบทุก field → ผ่านเลย
    if (hasBaseCurrency && hasLanguage && hasTimezone && hasYeartype) {
      // ยังต้องเช็คว่ามี currency ใน DB ด้วย
      return await _ensureCurrencyExists(branch.baseCurrency!);
    }

    if (!context.mounted) return false;

    // แสดง dialog ให้ตั้งค่าครบ
    final result = await showDialog<_BranchSettingsResult>(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => _BranchSetupDialogContent(
        initialCurrency: branch.baseCurrency,
        initialLanguage: branch.language,
        initialTimezone: branch.timezone,
        initialYeartype: branch.yeartype,
      ),
    );

    if (result == null) return false;

    // สร้าง currency ถ้ายังไม่มี
    await _ensureCurrencyCreated(result.currencyCode);

    // อัพเดต branch settings ทั้งหมด
    try {
      await _updateBranchSettings(
        currencyCode: result.currencyCode,
        language: result.language,
        timezone: result.timezone,
        timezonelabel: result.timezonelabel,
        timezoneoffset: result.timezoneoffset,
        yeartype: result.yeartype,
      );
    } catch (e) {
      AppLogger.error("[Branch Setup] บันทึกการตั้งค่าสาขาไม่สำเร็จ: $e");
      // ยังไม่ต้อง show error — เพราะ local cache ช่วย fallback อยู่
    }

    // บันทึกลง local cache เสมอ (ไม่ว่า backend จะ save ได้หรือไม่)
    await _saveBranchSettingsToCache(
      branchGuid: branchGuid,
      currencyCode: result.currencyCode,
      language: result.language,
      timezone: result.timezone,
      timezonelabel: result.timezonelabel,
      timezoneoffset: result.timezoneoffset,
      yeartype: result.yeartype,
    );

    return true;
  } catch (e) {
    AppLogger.error("[Branch Settings Check] Error: $e");
    // error อื่นๆ (เช่น context, dialog) → ให้ผ่านไปก่อน
    return true;
  }
}

/// Backward compatibility — เรียกจากที่เดิมได้
Future<bool> checkAndSetupCurrency(BuildContext context) async {
  return checkAndSetupBranchSettings(context);
}

// ─── Auto Setup Shop Defaults ───

/// ตรวจสอบและตั้งค่าเริ่มต้นของร้านค้า (ชื่อบริษัท, ภาษา, สกุลเงิน)
/// ถ้ายังไม่มี → ตั้งค่า default แล้ว PUT ไปยัง backend
/// เรียกหลังจากเลือก shop แล้ว ก่อนเข้า main menu
Future<void> autoSetupShopDefaults() async {
  try {
    final shopId = global.getShopId();

    if (shopId.isEmpty) {
      if (kDebugMode) {
        AppLogger.warning("[Shop Setup] shopId is empty — skip auto setup");
      }
      return;
    }

    // โหลด shop data จาก backend ก่อน (เพราะ global.shopSelectData อาจยังว่าง)
    final repo = ShopRepository();
    ShopModel shop;
    try {
      final loadResult = await repo.loadShopInfo(shopId);
      if (loadResult.success && loadResult.data != null) {
        shop = ShopModel.fromJson(loadResult.data);
        // ถ้า guidfixed ว่าง แสดงว่า shop ไม่มีใน DB (soft-deleted หรือยังไม่สร้าง)
        // Backend คืน empty ShopInfo กรณีไม่เจอ แต่ยัง success: true
        if (shop.guidfixed == null || shop.guidfixed!.isEmpty) {
          AppLogger.warning("[Shop Setup] Shop $shopId not found in database (guidfixed empty) — skip auto setup");
          return;
        }
        global.shopSelectData = shop;
        if (kDebugMode) {
          AppLogger.debug("[Shop Setup] โหลด shop data สำเร็จ: guidfixed=${shop.guidfixed}, names=${shop.names?.length ?? 0}, langs=${shop.settings?.languageconfigs?.length ?? 0}");
        }
      } else {
        AppLogger.warning("[Shop Setup] GET /shop/$shopId failed — skip auto setup");
        return;
      }
    } catch (e) {
      AppLogger.error("[Shop Setup] GET /shop/$shopId error: $e");
      return;
    }

    bool needsUpdate = false;

    // 1. ชื่อบริษัท — ถ้าว่าง ตั้งเป็น "TEMP"
    shop.names ??= [];
    if (shop.names!.isEmpty || shop.names!.every((n) => n.name.isEmpty)) {
      shop.names = [LanguageDataModel(code: 'th', name: 'TEMP')];
      needsUpdate = true;
      if (kDebugMode) {
        AppLogger.info("[Shop Setup] ตั้งชื่อบริษัทเริ่มต้น: TEMP (th)");
      }
    }

    // 2. ภาษา — ถ้าว่าง เพิ่มภาษาไทย
    shop.settings ??= Settings();
    shop.settings!.languageconfigs ??= [];
    if (shop.settings!.languageconfigs!.isEmpty) {
      shop.settings!.languageconfigs = [
        LanguageModel(
          code: 'th',
          codeTranslator: 'th',
          name: 'Thai',
          isuse: true,
          isdefault: true,
        ),
      ];
      needsUpdate = true;
      if (kDebugMode) {
        AppLogger.info("[Shop Setup] เพิ่มภาษาเริ่มต้น: ไทย (th)");
      }
    }

    // 3. สกุลเงินหลัก — ถ้าว่าง ตั้งเป็น THB
    if (_nonEmpty(shop.settings!.baseCurrency) == null) {
      shop.settings!.baseCurrency = 'THB';
      needsUpdate = true;
      if (kDebugMode) {
        AppLogger.info("[Shop Setup] ตั้งสกุลเงินเริ่มต้น: THB");
      }
    }

    // 4. ถ้ามีอะไรเปลี่ยน → PUT /shop/{shopid}
    if (needsUpdate) {
      try {
        final client = Client().init();
        final data = shop.toJson();
        final response = await client.put('/shop/$shopId', data: data);
        if (kDebugMode) {
          AppLogger.info("[Shop Setup] PUT /shop/$shopId → status=${response.statusCode}");
        }
      } on DioException catch (e) {
        AppLogger.error("[Shop Setup] PUT failed: status=${e.response?.statusCode}, body=${e.response?.data}");
      } catch (e) {
        AppLogger.error("[Shop Setup] PUT failed: $e");
      }
      global.shopSelectData = shop;
    }

    // 5. ประเภทธุรกิจ — ถ้าว่าง ให้สร้าง default
    await _ensureBusinessTypeExists();

    // 6. สาขาหลัก — ถ้า companynames ว่าง ให้ copy จาก shop names
    await _ensureBranchCompanyNames(shop);

  } catch (e) {
    AppLogger.error("[Shop Setup] autoSetupShopDefaults error: $e");
  }
}

/// ตรวจสอบว่ามีประเภทธุรกิจแล้วหรือยัง ถ้าไม่มี → สร้าง default "ธุรกิจหลัก"
Future<void> _ensureBusinessTypeExists() async {
  try {
    final btRepo = BusinessTypeRepository();
    final listResult = await btRepo.getBusinessTypeList(limit: 1);
    if (listResult.success) {
      final dataList = listResult.data;
      if (dataList is List && dataList.isNotEmpty) {
        return; // มีอยู่แล้ว
      }
    }
    // ไม่มี → สร้าง default
    if (kDebugMode) {
      AppLogger.info("[Shop Setup] ไม่มีประเภทธุรกิจ → สร้าง default 'ธุรกิจหลัก'");
    }
    final defaultBT = BusinessTypeModel(
      code: '00000',
      names: [
        LanguageDataModel(code: 'th', name: 'ธุรกิจหลัก'),
        LanguageDataModel(code: 'en', name: 'Main Business'),
      ],
      isdefault: true,
    );
    await btRepo.saveBusinessType(defaultBT);
    if (kDebugMode) {
      AppLogger.info("[Shop Setup] สร้างประเภทธุรกิจ default สำเร็จ");
    }
  } catch (e) {
    AppLogger.error("[Shop Setup] _ensureBusinessTypeExists error: $e");
  }
}

/// ตรวจสอบว่าสาขาหลัก (code 00000) มี companynames หรือยัง
/// ถ้าว่าง → copy จาก shop names แล้ว PUT กลับไป
Future<void> _ensureBranchCompanyNames(ShopModel shop) async {
  try {
    final branchRepo = CompanyBranchRepository();
    final result = await branchRepo.getBranchBycode('00000');
    if (!result.success || result.data == null) return;

    final branch = CompanyBranchModel.fromJson(result.data);
    final hasCompanyNames = branch.companynames.isNotEmpty &&
        branch.companynames.any((n) => n.name.isNotEmpty);
    if (hasCompanyNames) return;

    // companynames ว่าง → copy จาก shop names
    final shopNames = shop.names ?? [];
    if (shopNames.isEmpty || shopNames.every((n) => n.name.isEmpty)) return;

    branch.companynames = shopNames
        .map((n) => LanguageDataModel(code: n.code, name: n.name))
        .toList();

    if (branch.guidfixed.isNotEmpty) {
      await branchRepo.updateBranch(branch.guidfixed, branch);
      if (kDebugMode) {
        AppLogger.info("[Shop Setup] อัปเดต companynames ของสาขาหลัก สำเร็จ");
      }
    }
  } catch (e) {
    AppLogger.error("[Shop Setup] _ensureBranchCompanyNames error: $e");
  }
}

/// ตรวจสอบว่ามี currency ใน DB แล้วหรือยัง
Future<bool> _ensureCurrencyExists(String currencyCode) async {
  try {
    final apiService = CurrencyApiService();
    final response = await apiService.getCurrencyList(
      shopId: global.shopSelectData.guidfixed ?? '',
      limit: 100,
    );

    if (response.success && response.data != null) {
      final dataList = response.data as List<dynamic>? ?? [];
      final currencies = dataList
          .map((e) => CurrencyModel.fromJson(e as Map<String, dynamic>))
          .where((c) => !c.isdisabled)
          .toList();
      if (currencies.isNotEmpty) return true;
    }

    // ไม่มี currency → สร้างจาก baseCurrency ที่ตั้งไว้
    final preset = presetCurrencies.where((c) => c.code == currencyCode).firstOrNull;
    if (preset != null) {
      await apiService.createCurrency(CurrencyModel(
        code: preset.code,
        name: preset.name,
        symbol: preset.symbol,
      ));
    }
    return true;
  } catch (e) {
    if (kDebugMode) {
      AppLogger.error("[Currency Check] Error: $e");
    }
    return true;
  }
}

/// สร้าง currency ใน DB ถ้ายังไม่มี
Future<void> _ensureCurrencyCreated(String currencyCode) async {
  try {
    final apiService = CurrencyApiService();
    final response = await apiService.getCurrencyList(
      shopId: global.shopSelectData.guidfixed ?? '',
      limit: 100,
    );

    bool hasCurrency = false;
    if (response.success && response.data != null) {
      final dataList = response.data as List<dynamic>? ?? [];
      hasCurrency = dataList
          .map((e) => CurrencyModel.fromJson(e as Map<String, dynamic>))
          .where((c) => !c.isdisabled)
          .isNotEmpty;
    }

    if (!hasCurrency) {
      final preset = presetCurrencies.where((c) => c.code == currencyCode).firstOrNull;
      if (preset != null) {
        await apiService.createCurrency(CurrencyModel(
          code: preset.code,
          name: preset.name,
          symbol: preset.symbol,
        ));
        if (kDebugMode) {
          AppLogger.info("[Branch Setup] สร้าง currency: $currencyCode");
        }
      }
    }
  } catch (e) {
    if (kDebugMode) {
      AppLogger.error("[Branch Setup] สร้าง currency ไม่ได้: $e");
    }
  }
}

/// อัพเดต branch settings ทั้งหมด
/// save ไปยัง backend (PUT /organization/branch) + อัพเดต in-memory global
Future<void> _updateBranchSettings({
  required String currencyCode,
  required String language,
  required String timezone,
  required String timezonelabel,
  required String timezoneoffset,
  required String yeartype,
}) async {
  final branch = global.companyBranchSelectData;
  branch.baseCurrency = currencyCode;
  branch.language = language;
  branch.timezone = timezone;
  branch.timezonelabel = timezonelabel;
  branch.timezoneoffset = timezoneoffset;
  branch.yeartype = yeartype;

  // อัพเดตไปยัง backend เมื่อมี branch จริง (guidfixed ไม่ว่าง)
  if (branch.guidfixed.isNotEmpty) {
    try {
      final repo = CompanyBranchRepository();
      final result = await repo.updateBranch(branch.guidfixed, branch);
      if (kDebugMode) {
        AppLogger.info("[Branch Setup] PUT /organization/branch/${branch.guidfixed} "
            "→ success=${result.success}");
      }
      // Verify: re-fetch เพื่อยืนยันว่า backend save จริง
      try {
        final verify = await repo.getBranch(branch.guidfixed);
        if (verify.success && verify.data != null) {
          final saved = CompanyBranchModel.fromJson(verify.data);
          if (kDebugMode) {
            AppLogger.debug("[Branch Setup] Verify GET: "
                "currency=${saved.baseCurrency}, lang=${saved.language}, "
                "tz=${saved.timezone}, year=${saved.yeartype}");
          }
          // ถ้า backend ไม่ได้ save → log warning
          if (_nonEmpty(saved.baseCurrency) == null && currencyCode.isNotEmpty) {
            AppLogger.warning("[Branch Setup] ⚠️ Backend ไม่ได้ save base_currency! "
                "ส่ง '$currencyCode' แต่ GET กลับมา '${saved.baseCurrency}'");
          }
        }
      } catch (_) {}
    } catch (e) {
      AppLogger.error("[Branch Setup] PUT failed: $e");
      // ไม่ throw — local cache จะช่วย fallback
    }
  } else {
    if (kDebugMode) {
      AppLogger.warning("[Branch Setup] ⚠️ branch.guidfixed is empty — cannot PUT");
    }
  }

  global.companyBranchSelectData = branch;

  if (kDebugMode) {
    AppLogger.info(
      "[Branch Setup] อัพเดตสำเร็จ — currency=$currencyCode, "
      "language=$language, timezone=$timezone, yeartype=$yeartype",
    );
  }
}

// ─── Result Model ───

class _BranchSettingsResult {
  final String currencyCode;
  final String language;
  final String timezone;
  final String timezonelabel;
  final String timezoneoffset;
  final String yeartype;

  _BranchSettingsResult({
    required this.currencyCode,
    required this.language,
    required this.timezone,
    required this.timezonelabel,
    required this.timezoneoffset,
    required this.yeartype,
  });
}

// ─── Dialog Widget ───

class _BranchSetupDialogContent extends StatefulWidget {
  final String? initialCurrency;
  final String? initialLanguage;
  final String? initialTimezone;
  final String? initialYeartype;

  const _BranchSetupDialogContent({
    this.initialCurrency,
    this.initialLanguage,
    this.initialTimezone,
    this.initialYeartype,
  });

  @override
  State<_BranchSetupDialogContent> createState() => _BranchSetupDialogContentState();
}

class _BranchSetupDialogContentState extends State<_BranchSetupDialogContent> with global.ThemeRefreshMixin {
  String? _selectedCurrency;
  String? _selectedLanguage;
  TimezonesModel? _selectedTimezone;
  String? _selectedYeartype;
  bool _saving = false;

  List<TimezonesModel> _timezoneList = [];

  @override
  void initState() {
    super.initState();
    // ต้องตรวจสอบว่าค่า initial อยู่ใน preset list หรือไม่ — ถ้าไม่มีให้ set null
    final initCurrency = widget.initialCurrency;
    _selectedCurrency = (initCurrency != null && initCurrency.isNotEmpty &&
        presetCurrencies.any((c) => c.code == initCurrency))
        ? initCurrency
        : null;
    _selectedLanguage = (widget.initialLanguage != null && widget.initialLanguage!.isNotEmpty)
        ? widget.initialLanguage
        : null;
    _selectedYeartype = (widget.initialYeartype != null && widget.initialYeartype!.isNotEmpty)
        ? widget.initialYeartype
        : null;

    // โหลด timezone list
    _timezoneList = List.from(global.timezonesListData);

    // หา timezone ที่ตรงกับค่าปัจจุบัน
    if (widget.initialTimezone != null && widget.initialTimezone!.isNotEmpty) {
      _selectedTimezone = _findTimezoneByIana(widget.initialTimezone!);
    }
  }

  /// หา TimezonesModel จาก IANA timezone name (เช่น "Asia/Bangkok")
  TimezonesModel? _findTimezoneByIana(String iana) {
    for (final tz in _timezoneList) {
      if (tz.utc.contains(iana)) return tz;
    }
    return null;
  }

  /// เมื่อเลือกสกุลเงิน → auto-fill ทุก field ตาม country mapping
  void _onCurrencyChangedForceAll(String currencyCode) {
    setState(() {
      _selectedCurrency = currencyCode;

      final defaults = currencyDefaults[currencyCode];
      if (defaults != null) {
        _selectedYeartype = defaults.yeartype;
        _selectedLanguage = defaults.language;
        _selectedTimezone = _findTimezoneByIana(defaults.timezone);
      }
    });
  }

  bool get _isValid =>
      _selectedCurrency != null &&
      _selectedCurrency!.isNotEmpty &&
      _selectedLanguage != null &&
      _selectedLanguage!.isNotEmpty &&
      _selectedTimezone != null &&
      _selectedYeartype != null &&
      _selectedYeartype!.isNotEmpty;

  /// สร้าง IANA timezone จาก TimezonesModel
  String _getIanaTimezone() {
    if (_selectedTimezone == null) return '';
    // ใช้ utc ตัวแรกเป็น IANA timezone
    if (_selectedTimezone!.utc.isNotEmpty) {
      return _selectedTimezone!.utc.first;
    }
    return _selectedTimezone!.value;
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          Icon(Icons.settings, color: global.theme.infoHighlightTextColor),
          SizedBox(width: 8),
          Expanded(
            child: Text(
              global.language("branch_initial_setup"),
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
            ),
          ),
        ],
      ),
      content: SizedBox(
        width: 500,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language("please_setup_branch_settings"),
                style: TextStyle(color: global.theme.iconSecondaryColor, fontSize: 14),
              ),
              const SizedBox(height: 20),

              // ── 1. สกุลเงินหลัก ──
              _buildSectionLabel(
                icon: Icons.currency_exchange,
                label: global.language("base_currency"),
                color: Colors.amber.shade700,
              ),
              const SizedBox(height: 8),
              _buildCurrencyDropdown(),
              const SizedBox(height: 20),

              // ── 2. ภาษาที่ใช้ในสาขา ──
              _buildSectionLabel(
                icon: Icons.language,
                label: global.language("branch_language"),
                color: global.theme.positiveHighlightTextColor,
              ),
              const SizedBox(height: 8),
              _buildLanguageDropdown(),
              const SizedBox(height: 20),

              // ── 3. เขตเวลา ──
              _buildSectionLabel(
                icon: Icons.access_time,
                label: global.language("timezone"),
                color: Colors.teal.shade700,
              ),
              const SizedBox(height: 8),
              _buildTimezoneDropdown(),
              const SizedBox(height: 20),

              // ── 4. ประเภทปี ──
              _buildSectionLabel(
                icon: Icons.calendar_today,
                label: global.language("yeartype"),
                color: Colors.purple.shade700,
              ),
              const SizedBox(height: 8),
              _buildYeartypeSelector(),
            ],
          ),
        ),
      ),
      actions: [
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: (_isValid && !_saving)
                ? () {
                    setState(() => _saving = true);
                    Navigator.of(context).pop(_BranchSettingsResult(
                      currencyCode: _selectedCurrency!,
                      language: _selectedLanguage!,
                      timezone: _getIanaTimezone(),
                      timezonelabel: _selectedTimezone?.text ?? '',
                      timezoneoffset: _selectedTimezone?.offset ?? '',
                      yeartype: _selectedYeartype!,
                    ));
                  }
                : null,
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.infoHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
              padding: const EdgeInsets.symmetric(vertical: 14),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: _saving
                ? SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: global.theme.onPrimaryColor,
                    ),
                  )
                : Text(
                    global.language("confirm"),
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
          ),
        ),
      ],
    );
  }

  // ── Section Label ──

  Widget _buildSectionLabel({
    required IconData icon,
    required String label,
    required Color color,
  }) {
    return Row(
      children: [
        Icon(icon, size: 20, color: color),
        const SizedBox(width: 8),
        Text(
          label,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: color,
          ),
        ),
      ],
    );
  }

  // ── Currency Dropdown ──

  Widget _buildCurrencyDropdown() {
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(8),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<String>(
          value: _selectedCurrency,
          isExpanded: true,
          padding: EdgeInsets.symmetric(horizontal: 12),
          hint: Text(global.language("select_base_currency")),
          items: presetCurrencies.map((c) {
            return DropdownMenuItem<String>(
              value: c.code,
              child: Row(
                children: [
                  CircleAvatar(
                    radius: 14,
                    backgroundColor: Colors.amber.shade50,
                    child: Text(
                      c.symbol,
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: Colors.amber.shade800,
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Text('${c.code} - ${c.name}'),
                ],
              ),
            );
          }).toList(),
          onChanged: (value) {
            if (value != null) {
              // ถ้ายังไม่มีค่าตั้งไว้ → auto-fill ตาม country
              // ถ้ามีค่าอยู่แล้ว → force เปลี่ยนทุก field ตาม country ใหม่
              final isFirstTime = widget.initialCurrency == null || widget.initialCurrency!.isEmpty;
              if (isFirstTime) {
                _onCurrencyChangedForceAll(value);
              } else {
                _onCurrencyChangedForceAll(value);
              }
            }
          },
        ),
      ),
    );
  }

  // ── Language Dropdown ──

  Widget _buildLanguageDropdown() {
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(8),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<String>(
          value: _selectedLanguage,
          isExpanded: true,
          padding: EdgeInsets.symmetric(horizontal: 12),
          hint: Text(global.language("select_language")),
          items: languageOptions.map((lang) {
            return DropdownMenuItem<String>(
              value: lang.code,
              child: Text('${lang.code.toUpperCase()} ${lang.name}'),
            );
          }).toList(),
          onChanged: (value) {
            setState(() => _selectedLanguage = value);
          },
        ),
      ),
    );
  }

  // ── Timezone Dropdown ──

  Widget _buildTimezoneDropdown() {
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(8),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<TimezonesModel>(
          value: _selectedTimezone,
          isExpanded: true,
          padding: EdgeInsets.symmetric(horizontal: 12),
          hint: Text(global.language("select_timezone")),
          items: _timezoneList.map((tz) {
            return DropdownMenuItem<TimezonesModel>(
              value: tz,
              child: Text(
                tz.text,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(fontSize: 13),
              ),
            );
          }).toList(),
          onChanged: (value) {
            setState(() => _selectedTimezone = value);
          },
        ),
      ),
    );
  }

  // ── Year Type Selector ──

  Widget _buildYeartypeSelector() {
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(8),
      ),
      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: Row(
        children: [
          Expanded(
            child: RadioListTile<String>(
              value: 'christian',
              groupValue: _selectedYeartype,
              title: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    global.language("christian_era"),
                    style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                  ),
                  Text(
                    '${global.language("christian_era_short")} (${DateTime.now().year})',
                    style: TextStyle(fontSize: 11, color: global.theme.iconSecondaryColor),
                  ),
                ],
              ),
              dense: true,
              contentPadding: EdgeInsets.zero,
              onChanged: (value) => setState(() => _selectedYeartype = value),
            ),
          ),
          Expanded(
            child: RadioListTile<String>(
              value: 'buddhist',
              groupValue: _selectedYeartype,
              title: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    global.language("buddhist_era"),
                    style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                  ),
                  Text(
                    '${global.language("buddhist_era_short")} (${DateTime.now().year + 543})',
                    style: TextStyle(fontSize: 11, color: global.theme.iconSecondaryColor),
                  ),
                ],
              ),
              dense: true,
              contentPadding: EdgeInsets.zero,
              onChanged: (value) => setState(() => _selectedYeartype = value),
            ),
          ),
        ],
      ),
    );
  }
}
