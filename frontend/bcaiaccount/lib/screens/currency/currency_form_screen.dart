import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class CurrencyFormScreen extends StatefulWidget {
  final CurrencyModel? currency;

  const CurrencyFormScreen({super.key, this.currency});

  @override
  State<CurrencyFormScreen> createState() => _CurrencyFormScreenState();
}

class _CurrencyFormScreenState extends State<CurrencyFormScreen> with global.ThemeRefreshMixin {
  final _formKey = GlobalKey<FormState>();
  final CurrencyApiService _apiService = CurrencyApiService();

  late TextEditingController _codeController;
  late TextEditingController _nameController;
  late TextEditingController _symbolController;
  bool _isDisabled = false;
  bool _isSaving = false;

  bool get _isEditMode => widget.currency != null;

  // รายการสัญลักษณ์สกุลเงินทั่วโลก (เรียงตาม ASEAN ก่อน)
  static const List<Map<String, String>> _currencySymbols = [
    // ===== ASEAN Currencies =====
    {'symbol': '฿', 'name': 'Thai Baht (THB)'},
    {'symbol': '₫', 'name': 'Vietnamese Dong (VND)'},
    {'symbol': '₭', 'name': 'Lao Kip (LAK)'},
    {'symbol': '៛', 'name': 'Cambodian Riel (KHR)'},
    {'symbol': 'K', 'name': 'Myanmar Kyat (MMK)'},
    {'symbol': 'RM', 'name': 'Malaysian Ringgit (MYR)'},
    {'symbol': 'S\$', 'name': 'Singapore Dollar (SGD)'},
    {'symbol': 'Rp', 'name': 'Indonesian Rupiah (IDR)'},
    {'symbol': '₱', 'name': 'Philippine Peso (PHP)'},
    {'symbol': 'B\$', 'name': 'Brunei Dollar (BND)'},
    // ===== Major Currencies =====
    {'symbol': '\$', 'name': 'US Dollar (USD)'},
    {'symbol': '€', 'name': 'Euro (EUR)'},
    {'symbol': '¥', 'name': 'Yen/Yuan (JPY/CNY)'},
    {'symbol': '£', 'name': 'Pound Sterling (GBP)'},
    {'symbol': '₹', 'name': 'Indian Rupee (INR)'},
    {'symbol': '₽', 'name': 'Russian Ruble (RUB)'},
    {'symbol': '₩', 'name': 'Korean Won (KRW)'},
    // ===== Asia Pacific =====
    {'symbol': 'HK\$', 'name': 'Hong Kong Dollar (HKD)'},
    {'symbol': 'NT\$', 'name': 'Taiwan Dollar (TWD)'},
    {'symbol': 'A\$', 'name': 'Australian Dollar (AUD)'},
    {'symbol': 'NZ\$', 'name': 'New Zealand Dollar (NZD)'},
    // ===== Other Currencies =====
    {'symbol': '₪', 'name': 'Israeli Shekel (ILS)'},
    {'symbol': '₺', 'name': 'Turkish Lira (TRY)'},
    {'symbol': 'zł', 'name': 'Polish Zloty (PLN)'},
    {'symbol': 'kr', 'name': 'Scandinavian Krona'},
    {'symbol': 'CHF', 'name': 'Swiss Franc (CHF)'},
    {'symbol': 'R', 'name': 'South African Rand (ZAR)'},
    {'symbol': 'R\$', 'name': 'Brazilian Real (BRL)'},
    {'symbol': 'C\$', 'name': 'Canadian Dollar (CAD)'},
    {'symbol': '₡', 'name': 'Costa Rican Colon (CRC)'},
    {'symbol': '؋', 'name': 'Afghan Afghani (AFN)'},
    {'symbol': 'դdelays', 'name': 'Armenian Dram (AMD)'},
    {'symbol': '₼', 'name': 'Azerbaijani Manat (AZN)'},
  ];

  @override
  void initState() {
    super.initState();
    _codeController = TextEditingController(text: widget.currency?.code ?? "");
    _nameController = TextEditingController(text: widget.currency?.name ?? "");
    // ถ้าเป็นการเพิ่มใหม่ (ไม่ใช่แก้ไข) ให้ตั้งค่าเริ่มต้นเป็น Thai Baht
    _symbolController = TextEditingController(text: widget.currency?.symbol ?? "฿");
    _isDisabled = widget.currency?.isdisabled ?? false;
  }

  @override
  void dispose() {
    _codeController.dispose();
    _nameController.dispose();
    _symbolController.dispose();
    super.dispose();
  }

  Future<void> _saveCurrency() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    // ตรวจสอบว่าเลือกสัญลักษณ์หรือยัง
    if (_symbolController.text.trim().isEmpty) {
      global.showWarningSnackBar(context, global.language("please_select_currency_symbol"));
      return;
    }

    setState(() {
      _isSaving = true;
    });

    try {
      final currency = CurrencyModel(
        guidfixed: widget.currency?.guidfixed ?? "",
        code: _codeController.text.trim().toUpperCase(),
        name: _nameController.text.trim(),
        symbol: _symbolController.text.trim(),
        isdisabled: _isDisabled,
      );

      if (_isEditMode) {
        await _apiService.updateCurrency(currency.guidfixed, currency);
        if (mounted) {
          global.showSuccessSnackBar(context, global.language("edit_currency_success"));
        }
      } else {
        await _apiService.createCurrency(currency);
        if (mounted) {
          global.showSuccessSnackBar(context, global.language("add_currency_success"));
        }
      }

      if (mounted) {
        Navigator.pop(context, true);
      }
    } catch (e) {
      AppLogger.error('Failed to save currency: $e');
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    } finally {
      setState(() {
        _isSaving = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditMode ? global.language("edit_currency") : global.language("add_currency")),
        backgroundColor: global.theme.appBarColor,
        actions: [
          EditFontSizeControl(onChanged: () => setState(() {})),
          if (_isSaving)
            Center(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: CircularProgressIndicator(
                  color: global.theme.onPrimaryColor,
                  strokeWidth: 2,
                ),
              ),
            )
          else
            IconButton(
              icon: Icon(Icons.save),
              onPressed: _saveCurrency,
              tooltip: global.language("save"),
            ),
        ],
      ),
      body: Builder(
        builder: (context) {
          final scaleFactor = global.editFontScaleFactor;
          return MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: TextScaler.linear(scaleFactor),
            ),
            child: Theme(
              data: Theme.of(context).copyWith(
                inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 12 * scaleFactor,
                    vertical: 10 * scaleFactor,
                  ),
                ),
                iconTheme: IconThemeData(size: 24 * scaleFactor),
              ),
              child: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16.0),
          children: [
            // Currency Code
            TextFormField(
              controller: _codeController,
              decoration: InputDecoration(
                labelText: '${global.language("currency_code")} *',
                hintText: global.language('currency_code_example'),
                border: OutlineInputBorder(),
                prefixIcon: Icon(Icons.code),
              ),
              textCapitalization: TextCapitalization.characters,
              enabled: !_isEditMode, // ไม่ให้แก้ไข code ถ้าเป็นโหมดแก้ไข
              validator: (value) {
                if (value == null || value.trim().isEmpty) {
                  return global.language("please_enter_currency_code");
                }
                if (value.trim().length < 2 || value.trim().length > 5) {
                  return global.language("currency_code_length_error");
                }
                return null;
              },
            ),
            const SizedBox(height: 16),

            // Currency Name
            TextFormField(
              controller: _nameController,
              decoration: InputDecoration(
                labelText: '${global.language("currency_name")} *',
                hintText: global.language('currency_name_example'),
                border: OutlineInputBorder(),
                prefixIcon: Icon(Icons.text_fields),
              ),
              validator: (value) {
                if (value == null || value.trim().isEmpty) {
                  return global.language("please_enter_currency_name");
                }
                return null;
              },
            ),
            const SizedBox(height: 16),

            // Currency Symbol Selector
            Text(
              '${global.language("symbol")} *',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w500,
              ),
            ),
            const SizedBox(height: 12),
            Container(
              decoration: BoxDecoration(
                border: Border.all(color: global.theme.dividerBorderColor),
                borderRadius: BorderRadius.circular(8),
              ),
              padding: const EdgeInsets.all(12),
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                children: _currencySymbols.map((item) {
                  final isSelected = _symbolController.text == item['symbol'];

                  return InkWell(
                    onTap: () {
                      setState(() {
                        _symbolController.text = item['symbol']!;
                      });
                    },
                    child: Container(
                      width: 100,
                      height: 80,
                      decoration: BoxDecoration(
                        border: Border.all(
                          color: isSelected ? global.theme.appBarColor : global.theme.dividerBorderColor,
                          width: isSelected ? 2 : 1,
                        ),
                        borderRadius: BorderRadius.circular(8),
                        color: isSelected
                            ? global.theme.appBarColor.withValues(alpha: 0.1)
                            : global.theme.cardColor,
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            item['symbol']!,
                            style: TextStyle(
                              fontSize: 28,
                              fontWeight: FontWeight.bold,
                              color: isSelected ? global.theme.appBarColor : global.theme.textColor,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            item['name']!,
                            textAlign: TextAlign.center,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              fontSize: 9,
                              color: isSelected ? global.theme.appBarColor : global.theme.textSecondaryColor,
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                }).toList(),
              ),
            ),
            const SizedBox(height: 16),

            // Disabled checkbox
            SwitchListTile(
              title: Text(global.language("disabled")),
              subtitle: Text(global.language("disabled_currency_not_shown")),
              value: _isDisabled,
              onChanged: (value) {
                setState(() {
                  _isDisabled = value;
                });
              },
              activeThumbColor: global.theme.appBarColor,
            ),

            const SizedBox(height: 24),

            // Save button
            ElevatedButton.icon(
              onPressed: _isSaving ? null : _saveCurrency,
              icon: _isSaving
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Icon(Icons.save),
              label: Text(_isSaving ? global.language("saving") : global.language("save")),
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.appBarColor,
                foregroundColor: global.theme.onPrimaryColor,
                padding: const EdgeInsets.all(16),
              ),
            ),
          ],
        ),
      ),
            ),
          );
        },
      ),
    );
  }
}
