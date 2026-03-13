import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/screens/currency/currency_form_screen.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class CurrencyListScreen extends StatefulWidget {
  const CurrencyListScreen({super.key});

  @override
  State<CurrencyListScreen> createState() => _CurrencyListScreenState();
}

class _CurrencyListScreenState extends State<CurrencyListScreen> with global.ThemeRefreshMixin {
  final CurrencyApiService _apiService = CurrencyApiService();
  List<CurrencyModel> _currencies = [];
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadCurrencies();
  }

  Future<void> _loadCurrencies() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final String shopId = global.prefs.getString("shopid") ?? "";
      final response = await _apiService.getCurrencies(
        shopId: shopId,
        page: 1,
        limit: 1000,
      );

      if (response.success && response.data != null) {
        setState(() {
          _currencies = (response.data as List)
              .map((e) => CurrencyModel.fromJson(e))
              .toList();
        });
      }
    } catch (e) {
      AppLogger.error('Failed to load currencies: $e');
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_loading")}: $e');
      }
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future<void> _deleteCurrency(CurrencyModel currency) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language("confirm_delete")),
        content: Text('${global.language("confirm_delete_currency")} ${currency.code} - ${currency.name} ?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language("cancel")),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
            child: Text(global.language("delete")),
          ),
        ],
      ),
    );

    if (confirm == true) {
      try {
        final response = await _apiService.deleteCurrency(currency.guidfixed);
        if (response.success) {
          if (mounted) {
            global.showSuccessSnackBar(context, global.language("delete_currency_success"));
          }
          _loadCurrencies();
        }
      } catch (e) {
        AppLogger.error('Failed to delete currency: $e');
        if (mounted) {
          global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
        }
      }
    }
  }

  void _navigateToForm({CurrencyModel? currency}) async {
    final result = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => CurrencyFormScreen(currency: currency),
      ),
    );

    if (result == true) {
      _loadCurrencies();
    }
  }

  @override
  Widget build(BuildContext context) {
    final baseCurrency = global.getBaseCurrency();

    return Scaffold(
      appBar: AppBar(
        title: Text(global.language("currency")),
        backgroundColor: global.theme.appBarColor,
        actions: const [ManualButton(path: 'settings-currency')],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                // Currency cards (natural height)
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Wrap(
                    spacing: 12,
                    runSpacing: 12,
                    children: [
                      ..._currencies.map((currency) {
                        final isBase = currency.code.toUpperCase() == baseCurrency.toUpperCase();
                        return _buildCurrencyCard(currency, isBase);
                      }),
                      _buildAddCard(),
                    ],
                  ),
                ),
              ],
            ),
    );
  }

  Widget _buildCurrencyCard(CurrencyModel currency, bool isBase) {
    final isDisabled = currency.isdisabled;
    final isDark = global.isDarkMode();
    final accentColor = isDark ? global.theme.primaryLightColor : global.theme.appBarColor;
    final color = isDisabled
        ? global.theme.iconSecondaryColor
        : isBase
            ? accentColor
            : isDark ? Colors.blueGrey.shade300 : Colors.blueGrey;

    return SizedBox(
      width: 160,
      child: Card(
        elevation: isBase ? 4 : 2,
        color: isDisabled
            ? global.theme.surfaceColor
            : null, // ใช้ ThemeData.cardColor อัตโนมัติ
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: isBase
              ? BorderSide(color: accentColor, width: 2)
              : BorderSide(color: global.theme.dividerBorderColor, width: 0.5),
        ),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(12, 16, 12, 8),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Symbol
              Text(
                currency.symbol.isNotEmpty
                    ? currency.symbol
                    : currency.code.substring(0, 1),
                style: TextStyle(
                  fontSize: 36,
                  fontWeight: FontWeight.bold,
                  color: color,
                ),
              ),
              const SizedBox(height: 4),

              // Code
              Text(
                currency.code,
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: isDisabled ? global.theme.textSecondaryColor : global.theme.textColor,
                ),
              ),
              const SizedBox(height: 2),

              // Name
              Text(
                currency.name,
                textAlign: TextAlign.center,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: 12,
                  color: global.theme.textSecondaryColor,
                ),
              ),
              const SizedBox(height: 6),

              // Badge
              if (isBase)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: accentColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    global.language("base_currency"),
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.w600,
                      color: accentColor,
                    ),
                  ),
                )
              else if (isDisabled)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: global.theme.dividerBorderColor,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    global.language("disabled"),
                    style: TextStyle(fontSize: 10, color: global.theme.textSecondaryColor),
                  ),
                )
              else
                const SizedBox(height: 16),

              // Actions
              Divider(height: 16, color: global.theme.dividerBorderColor),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  InkWell(
                    onTap: () => _navigateToForm(currency: currency),
                    borderRadius: BorderRadius.circular(8),
                    child: Padding(
                      padding: const EdgeInsets.all(6),
                      child: Icon(Icons.edit_outlined, size: 20, color: global.theme.warningHighlightTextColor),
                    ),
                  ),
                  InkWell(
                    onTap: () => _deleteCurrency(currency),
                    borderRadius: BorderRadius.circular(8),
                    child: Padding(
                      padding: const EdgeInsets.all(6),
                      child: Icon(Icons.delete_outline, size: 20, color: global.theme.negativeHighlightTextColor),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildAddCard() {
    return SizedBox(
      width: 160,
      child: InkWell(
        onTap: () => _navigateToForm(),
        borderRadius: BorderRadius.circular(12),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 40),
          decoration: BoxDecoration(
            color: Theme.of(context).cardColor,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: global.theme.dividerBorderColor,
              width: 1.5,
              strokeAlign: BorderSide.strokeAlignInside,
            ),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.add_circle_outline, size: 40, color: global.theme.iconSecondaryColor),
              SizedBox(height: 8),
              Text(
                global.language("add_currency"),
                style: TextStyle(
                  fontSize: 14,
                  color: global.theme.textSecondaryColor,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

}
