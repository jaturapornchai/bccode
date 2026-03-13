import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:intl/intl.dart';

class ExchangeRateHistoryScreen extends StatefulWidget {
  final CurrencyModel currency;

  const ExchangeRateHistoryScreen({super.key, required this.currency});

  @override
  State<ExchangeRateHistoryScreen> createState() =>
      _ExchangeRateHistoryScreenState();
}

class _ExchangeRateHistoryScreenState extends State<ExchangeRateHistoryScreen> with global.ThemeRefreshMixin {
  final CurrencyApiService _apiService = CurrencyApiService();
  List<ExchangeRateHistoryModel> _exchangeRates = [];
  bool _isLoading = false;
  final _formKey = GlobalKey<FormState>();
  final TextEditingController _rateController = TextEditingController();
  DateTime _selectedDate = DateTime.now();
  bool _isAddingRate = false;

  @override
  void initState() {
    super.initState();
    _loadExchangeRates();
  }

  @override
  void dispose() {
    _rateController.dispose();
    super.dispose();
  }

  Future<void> _loadExchangeRates() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final String shopId = global.prefs.getString("shopid") ?? "";
      final response = await _apiService.getExchangeRateHistory(
        shopId: shopId,
        currency: widget.currency.code,
        page: 1,
        limit: 1000,
      );

      if (response.success && response.data != null) {
        setState(() {
          _exchangeRates = (response.data as List)
              .map((e) => ExchangeRateHistoryModel.fromJson(e))
              .toList();
          // Sort by date descending (newest first)
          _exchangeRates.sort((a, b) => b.date.compareTo(a.date));
        });
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

  Future<void> _saveExchangeRate() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    try {
      AppLogger.info('Creating exchange rate for currency: ${widget.currency.code}');
      AppLogger.info('Selected date: ${DateFormat('yyyy-MM-dd').format(_selectedDate)}');
      AppLogger.info('Rate text: ${_rateController.text.trim()}');

      final exchangeRate = ExchangeRateHistoryModel(
        currency: widget.currency.code,
        date: DateFormat('yyyy-MM-dd').format(_selectedDate),
        rate: double.parse(_rateController.text.trim()),
      );

      AppLogger.info('ExchangeRateHistoryModel created - currency: ${exchangeRate.currency}, date: ${exchangeRate.date}, rate: ${exchangeRate.rate}');

      await _apiService.createExchangeRate(exchangeRate);

      if (mounted) {
        global.showSuccessSnackBar(context, global.language("add_exchange_rate_success"));
        setState(() {
          _isAddingRate = false;
          _rateController.clear();
          _selectedDate = DateTime.now();
        });
        _loadExchangeRates();
      }
    } catch (e) {
      AppLogger.error('Failed to save exchange rate: $e');
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    }
  }

  Future<void> _deleteExchangeRate(ExchangeRateHistoryModel rate) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language("confirm_delete")),
        content: Text(
            '${global.language("confirm_delete_exchange_rate")} ${rate.date} = ${rate.rate} THB ?'),
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
        final response = await _apiService.deleteExchangeRate(rate.guidfixed);
        if (response.success) {
          if (mounted) {
            global.showSuccessSnackBar(context, global.language("delete_exchange_rate_success"));
          }
          _loadExchangeRates();
        }
      } catch (e) {
        AppLogger.error('Failed to delete exchange rate: $e');
        if (mounted) {
          global.showErrorSnackBar(context, '${global.language("error_deleting_data")}: $e');
        }
      }
    }
  }


  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
            '${global.language("exchange_rate_history")} - ${widget.currency.code} (${widget.currency.symbol})'),
        backgroundColor: global.theme.appBarColor,
        actions: [
          IconButton(
            icon: Icon(_isAddingRate ? Icons.close : Icons.add),
            onPressed: () {
              setState(() {
                _isAddingRate = !_isAddingRate;
                if (!_isAddingRate) {
                  _rateController.clear();
                  _selectedDate = DateTime.now();
                }
              });
            },
            tooltip: _isAddingRate ? global.language("cancel") : global.language("add_exchange_rate"),
          ),
        ],
      ),
      body: Column(
        children: [
          // Add exchange rate form
          if (_isAddingRate)
            Card(
              margin: EdgeInsets.all(16),
              child: Padding(
                padding: EdgeInsets.all(16),
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Text(
                        global.language("add_new_exchange_rate"),
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 16),

                      // Date picker (Thai format)
                      CustomDatePicker(
                        labelText: global.language("date"),
                        initialDate: _selectedDate,
                        firstDate: DateTime(2000),
                        lastDate: DateTime(2100),
                        decoration: const InputDecoration(
                          border: OutlineInputBorder(),
                          prefixIcon: Icon(Icons.calendar_today),
                        ),
                        onDateSelected: (date) {
                          if (date != null) {
                            setState(() {
                              _selectedDate = date;
                            });
                          }
                        },
                      ),
                      const SizedBox(height: 16),

                      // Exchange rate input
                      TextFormField(
                        controller: _rateController,
                        decoration: InputDecoration(
                          labelText: '${global.language("exchange_rate")} (1 ${widget.currency.code} = ? THB)',
                          hintText: global.language("exchange_rate_hint"),
                          border: const OutlineInputBorder(),
                          prefixIcon: Icon(Icons.monetization_on),
                          suffixText: 'THB',
                        ),
                        keyboardType: const TextInputType.numberWithOptions(
                            decimal: true),
                        validator: (value) {
                          if (value == null || value.trim().isEmpty) {
                            return global.language("please_enter_exchange_rate");
                          }
                          final rate = double.tryParse(value.trim());
                          if (rate == null || rate <= 0) {
                            return global.language("please_enter_valid_number");
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 16),

                      // Save button
                      ElevatedButton.icon(
                        onPressed: _saveExchangeRate,
                        icon: Icon(Icons.save),
                        label: Text(global.language("save")),
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
            ),

          // Exchange rate history list
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _exchangeRates.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.history,
                              size: 64,
                              color: global.theme.textSecondaryColor,
                            ),
                            SizedBox(height: 16),
                            Text(
                              global.language("no_exchange_rate_history"),
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
                        onRefresh: _loadExchangeRates,
                        child: ListView.builder(
                          itemCount: _exchangeRates.length,
                          itemBuilder: (context, index) {
                            final rate = _exchangeRates[index];
                            final displayDate = global.formatDateTime(
                              dateTime: DateTime.parse(rate.date),
                              showTime: false,
                            );

                            return Card(
                              margin: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 4,
                              ),
                              child: ListTile(
                                leading: CircleAvatar(
                                  backgroundColor: global.theme.appBarColor
                                      .withValues(alpha: 0.2),
                                  child: Text(
                                    widget.currency.symbol,
                                    style: TextStyle(
                                      fontWeight: FontWeight.bold,
                                      color: global.theme.appBarColor,
                                    ),
                                  ),
                                ),
                                title: Text(
                                  '1 ${widget.currency.code} = ${rate.rate.toStringAsFixed(4)} THB',
                                  style: TextStyle(
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                subtitle: Text('${global.language("date")}: $displayDate'),
                                trailing: IconButton(
                                  icon: Icon(Icons.delete),
                                  color: global.theme.negativeHighlightTextColor,
                                  onPressed: () => _deleteExchangeRate(rate),
                                  tooltip: global.language("delete"),
                                ),
                              ),
                            );
                          },
                        ),
                      ),
          ),
        ],
      ),
    );
  }
}
