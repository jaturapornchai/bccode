import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/screens/currency/exchange_rate_history_screen.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:intl/intl.dart';

/// หน้ากำหนดอัตราแลกเปลี่ยน
/// แสดงตารางอัตราแลกเปลี่ยนของทุกสกุลเงิน พร้อมประวัติ
class ExchangeRateScreen extends StatefulWidget {
  const ExchangeRateScreen({super.key});

  @override
  State<ExchangeRateScreen> createState() => _ExchangeRateScreenState();
}

class _ExchangeRateScreenState extends State<ExchangeRateScreen> with global.ThemeRefreshMixin {
  final CurrencyApiService _apiService = CurrencyApiService();
  List<CurrencyWithLatestRate> _currencies = [];
  List<ExchangeRateHistoryModel> _allExchangeRates = [];
  bool _isLoading = false;
  String _searchText = "";
  final TextEditingController _searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  /// โหลดข้อมูลสกุลเงินและอัตราแลกเปลี่ยน
  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final String shopId = global.prefs.getString("shopid") ?? "";
      
      // โหลดรายการสกุลเงิน
      final currencyResponse = await _apiService.getCurrencies(
        shopId: shopId,
        page: 1,
        limit: 1000,
      );

      // โหลดประวัติอัตราแลกเปลี่ยนทั้งหมด
      final rateResponse = await _apiService.getExchangeRateHistory(
        shopId: shopId,
        page: 1,
        limit: 1000,
      );

      if (currencyResponse.success && currencyResponse.data != null) {
        final currencies = (currencyResponse.data as List)
            .map((e) => CurrencyModel.fromJson(e))
            .toList();

        // กรองเฉพาะสกุลเงินที่ไม่ได้ปิดการใช้งาน
        currencies.removeWhere((c) => c.isdisabled);

        if (rateResponse.success && rateResponse.data != null) {
          _allExchangeRates = (rateResponse.data as List)
              .map((e) => ExchangeRateHistoryModel.fromJson(e))
              .toList();

          // จัดเรียงตามวันที่ล่าสุด
          _allExchangeRates.sort((a, b) => b.date.compareTo(a.date));
        }

        // รวมข้อมูลสกุลเงินกับอัตราแลกเปลี่ยนล่าสุด
        _currencies = currencies.map((currency) {
          // หาอัตราแลกเปลี่ยนล่าสุดของสกุลเงินนี้
          final latestRate = _allExchangeRates
              .where((r) => r.currency == currency.code)
              .firstOrNull;

          // หาประวัติทั้งหมดของสกุลเงินนี้
          final history = _allExchangeRates
              .where((r) => r.currency == currency.code)
              .toList();

          return CurrencyWithLatestRate(
            currency: currency,
            latestRate: latestRate,
            history: history,
          );
        }).toList();

        // เรียงลำดับตามรหัสสกุลเงิน
        _currencies.sort((a, b) => a.currency.code.compareTo(b.currency.code));
      }
    } catch (e) {
      AppLogger.error('Failed to load exchange rates: $e');
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_loading_data")}: $e');
      }
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  /// กรองรายการตามข้อความค้นหา
  List<CurrencyWithLatestRate> get _filteredCurrencies {
    if (_searchText.isEmpty) return _currencies;
    return _currencies.where((c) {
      final searchLower = _searchText.toLowerCase();
      return c.currency.code.toLowerCase().contains(searchLower) ||
          c.currency.name.toLowerCase().contains(searchLower) ||
          c.currency.symbol.toLowerCase().contains(searchLower);
    }).toList();
  }

  /// นำทางไปหน้าประวัติอัตราแลกเปลี่ยนของสกุลเงิน
  void _navigateToHistory(CurrencyModel currency) async {
    await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ExchangeRateHistoryScreen(currency: currency),
      ),
    );
    // รีโหลดข้อมูลเมื่อกลับมา
    _loadData();
  }

  /// แสดง Dialog ประวัติอัตราแลกเปลี่ยน
  void _showHistoryDialog(CurrencyWithLatestRate data) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('${global.language("exchange_rate_history")} - ${data.currency.code}'),
        content: SizedBox(
          width: double.maxFinite,
          child: data.history.isEmpty
              ? Center(
                  child: Text(
                    global.language("no_exchange_rate_history_short"),
                    style: TextStyle(color: global.theme.textSecondaryColor),
                  ),
                )
              : ListView.builder(
                  shrinkWrap: true,
                  itemCount: data.history.length > 10 ? 10 : data.history.length,
                  itemBuilder: (context, index) {
                    final rate = data.history[index];
                    final displayDate = global.formatDateTime(
                      dateTime: DateTime.parse(rate.date),
                      showTime: false,
                    );
                    return ListTile(
                      dense: true,
                      leading: CircleAvatar(
                        radius: 16,
                        backgroundColor: index == 0
                            ? global.theme.positiveHighlightColor
                            : global.theme.dividerBorderColor,
                        child: Text(
                          '${index + 1}',
                          style: TextStyle(
                            fontSize: 12,
                            color: index == 0 ? global.theme.positiveHighlightTextColor : global.theme.textSecondaryColor,
                          ),
                        ),
                      ),
                      title: Text(
                        '1 ${data.currency.code} = ${rate.rate.toStringAsFixed(4)} THB',
                        style: TextStyle(
                          fontWeight: index == 0 ? FontWeight.bold : FontWeight.normal,
                          color: index == 0 ? global.theme.positiveHighlightTextColor : null,
                        ),
                      ),
                      subtitle: Text(displayDate),
                      trailing: index == 0
                          ? Chip(
                              label: Text(global.language("latest")),
                              backgroundColor: global.theme.positiveHighlightColor,
                              labelStyle: TextStyle(
                                fontSize: 10,
                                color: global.theme.positiveHighlightTextColor,
                              ),
                              padding: EdgeInsets.zero,
                              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                            )
                          : null,
                    );
                  },
                ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language("close")),
          ),
          if (data.history.isNotEmpty)
            ElevatedButton(
              onPressed: () {
                Navigator.pop(context);
                _navigateToHistory(data.currency);
              },
              child: Text(global.language("view_all")),
            ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(global.language("set_exchange_rate")),
        backgroundColor: global.theme.appBarColor,
        actions: [
          IconButton(
            icon: Icon(Icons.refresh),
            onPressed: _loadData,
            tooltip: global.language("refresh"),
          ),
        ],
      ),
      body: Column(
        children: [
          // Search bar
          Padding(
            padding: EdgeInsets.all(16.0),
            child: TextField(
              controller: _searchController,
              decoration: InputDecoration(
                hintText: global.language("search_currency"),
                prefixIcon: Icon(Icons.search),
                suffixIcon: _searchText.isNotEmpty
                    ? IconButton(
                        icon: Icon(Icons.clear),
                        onPressed: () {
                          _searchController.clear();
                          setState(() {
                            _searchText = "";
                          });
                        },
                      )
                    : null,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
              onChanged: (value) {
                setState(() {
                  _searchText = value;
                });
              },
            ),
          ),

          // Header row
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: global.theme.appBarColor.withValues(alpha: 0.1),
              border: Border(
                bottom: BorderSide(color: global.theme.dividerBorderColor),
              ),
            ),
            child: Row(
              children: [
                Expanded(
                  flex: 2,
                  child: Text(
                    global.language("currency"),
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Text(
                    '${global.language("exchange_rate")} (THB)',
                    style: TextStyle(fontWeight: FontWeight.bold),
                    textAlign: TextAlign.center,
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Text(
                    global.language("updated_date"),
                    style: TextStyle(fontWeight: FontWeight.bold),
                    textAlign: TextAlign.center,
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Text(
                    global.language("manage"),
                    style: TextStyle(fontWeight: FontWeight.bold),
                    textAlign: TextAlign.center,
                  ),
                ),
              ],
            ),
          ),

          // Data table
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _filteredCurrencies.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.currency_exchange,
                              size: 64,
                              color: global.theme.iconSecondaryColor,
                            ),
                            SizedBox(height: 16),
                            Text(
                              _searchText.isEmpty
                                  ? global.language("no_currency_data")
                                  : global.language("no_search_results"),
                              textAlign: TextAlign.center,
                              style: TextStyle(
                                fontSize: 16,
                                color: global.theme.textSecondaryColor,
                              ),
                            ),
                          ],
                        ),
                      )
                    : RefreshIndicator(
                        onRefresh: _loadData,
                        child: ListView.builder(
                          itemCount: _filteredCurrencies.length,
                          itemBuilder: (context, index) {
                            final data = _filteredCurrencies[index];
                            final latestRate = data.latestRate;
                            final hasRate = latestRate != null;

                            // Format date
                            String dateText = '-';
                            if (hasRate) {
                              try {
                                final date = DateTime.parse(latestRate.date);
                                dateText = DateFormat('dd/MM/yyyy').format(date);
                              } catch (e) {
                                dateText = latestRate.date;
                              }
                            }

                            return Container(
                              decoration: BoxDecoration(
                                border: Border(
                                  bottom: BorderSide(color: global.theme.dividerBorderColor),
                                ),
                                color: index % 2 == 0 ? global.theme.columnAlternateEvenColor : global.theme.columnAlternateOddColor,
                              ),
                              child: InkWell(
                                onTap: () => _navigateToHistory(data.currency),
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 16,
                                    vertical: 12,
                                  ),
                                  child: Row(
                                    children: [
                                      // Currency info
                                      Expanded(
                                        flex: 2,
                                        child: Row(
                                          children: [
                                            CircleAvatar(
                                              radius: 16,
                                              backgroundColor: global.theme.appBarColor
                                                  .withValues(alpha: 0.2),
                                              child: Text(
                                                data.currency.symbol.isNotEmpty
                                                    ? data.currency.symbol
                                                    : data.currency.code.substring(0, 1),
                                                style: TextStyle(
                                                  fontWeight: FontWeight.bold,
                                                  fontSize: 12,
                                                  color: global.theme.appBarColor,
                                                ),
                                              ),
                                            ),
                                            const SizedBox(width: 12),
                                            Expanded(
                                              child: Column(
                                                crossAxisAlignment: CrossAxisAlignment.start,
                                                children: [
                                                  Text(
                                                    data.currency.code,
                                                    style: TextStyle(
                                                      fontWeight: FontWeight.bold,
                                                    ),
                                                  ),
                                                  Text(
                                                    data.currency.name,
                                                    style: TextStyle(
                                                      fontSize: 12,
                                                      color: global.theme.textSecondaryColor,
                                                    ),
                                                    overflow: TextOverflow.ellipsis,
                                                  ),
                                                ],
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                      // Exchange rate
                                      Expanded(
                                        flex: 2,
                                        child: Text(
                                          hasRate
                                              ? latestRate.rate.toStringAsFixed(4)
                                              : global.language("not_set"),
                                          textAlign: TextAlign.center,
                                          style: TextStyle(
                                            fontWeight: hasRate ? FontWeight.bold : FontWeight.normal,
                                            color: hasRate ? global.theme.positiveHighlightTextColor : global.theme.negativeHighlightTextColor,
                                          ),
                                        ),
                                      ),
                                      // Date
                                      Expanded(
                                        flex: 2,
                                        child: Text(
                                          dateText,
                                          textAlign: TextAlign.center,
                                          style: TextStyle(fontSize: 13),
                                        ),
                                      ),
                                      // Actions
                                      Expanded(
                                        flex: 1,
                                        child: Row(
                                          mainAxisAlignment: MainAxisAlignment.center,
                                          children: [
                                            // History button
                                            IconButton(
                                              icon: Icon(Icons.history),
                                              color: global.theme.infoHighlightTextColor,
                                              onPressed: () => _showHistoryDialog(data),
                                              tooltip: global.language("history"),
                                              padding: EdgeInsets.zero,
                                              constraints: const BoxConstraints(),
                                              iconSize: 20,
                                            ),
                                            const SizedBox(width: 8),
                                            // Edit button
                                            IconButton(
                                              icon: Icon(Icons.edit),
                                              color: global.theme.warningHighlightTextColor,
                                              onPressed: () => _navigateToHistory(data.currency),
                                              tooltip: global.language("edit"),
                                              padding: EdgeInsets.zero,
                                              constraints: const BoxConstraints(),
                                              iconSize: 20,
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            );
                          },
                        ),
                      ),
          ),

          // Summary footer
          if (_currencies.isNotEmpty)
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: global.theme.cardColor,
                border: Border(
                  top: BorderSide(color: global.theme.dividerBorderColor),
                ),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    '${global.language("currency_count")}: ${_currencies.length}',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  Text(
                    '${global.language("exchange_rate_set_count")}: ${_currencies.where((c) => c.latestRate != null).length}',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

/// Helper class สำหรับรวมข้อมูลสกุลเงินกับอัตราแลกเปลี่ยนล่าสุด
class CurrencyWithLatestRate {
  final CurrencyModel currency;
  final ExchangeRateHistoryModel? latestRate;
  final List<ExchangeRateHistoryModel> history;

  CurrencyWithLatestRate({
    required this.currency,
    this.latestRate,
    required this.history,
  });
}
