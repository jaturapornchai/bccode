import 'package:flutter/material.dart';
import '../models/debtor_model.dart';
import '../services/clickhouse_debtor_creditor_service.dart';
import '../../../utils/logger/app_logger.dart';
import '../../../global.dart' as global;

/// Dialog สำหรับเลือกลูกหนี้
class DebtorSelectorDialog extends StatefulWidget {
  final String shopId;

  const DebtorSelectorDialog({
    super.key,
    required this.shopId,
  });

  @override
  State<DebtorSelectorDialog> createState() => _DebtorSelectorDialogState();
}

class _DebtorSelectorDialogState extends State<DebtorSelectorDialog> with global.ThemeRefreshMixin {
  final TextEditingController _searchController = TextEditingController();
  final ClickHouseDebtorCreditorService _service =
      ClickHouseDebtorCreditorService();

  List<DebtorModel> _debtors = [];
  bool _isLoading = false;
  bool _hasSearched = false;
  String? _errorMessage;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _searchDebtors() async {
    final keyword = _searchController.text.trim();
    if (keyword.isEmpty) {
      setState(() {
        _errorMessage = global.language("please_enter_search_keyword");
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
      _hasSearched = true;
    });

    try {
      final debtors = await _service.searchDebtors(
        keyword: keyword,
        shopId: widget.shopId,
        limit: 50,
      );

      if (mounted) {
        setState(() {
          _debtors = debtors;
          _isLoading = false;
        });
      }
    } catch (e) {
      AppLogger.error('Error searching debtors: $e');
      if (mounted) {
        setState(() {
          _errorMessage = '${global.language("search_error")}: $e';
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Container(
        constraints: BoxConstraints(maxWidth: 600, maxHeight: 700),
        child: Column(
          children: [
            // Header
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: global.theme.infoHighlightColor,
                border: Border(
                  bottom: BorderSide(color: global.theme.dividerBorderColor),
                ),
              ),
              child: Row(
                children: [
                  Icon(Icons.person, color: global.theme.infoHighlightTextColor),
                  SizedBox(width: 8),
                  Text(
                    global.language("select_debtor"),
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  const Spacer(),
                  IconButton(
                    icon: Icon(Icons.close),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
            ),

            // Search Bar
            Padding(
              padding: EdgeInsets.all(16),
              child: TextField(
                controller: _searchController,
                decoration: InputDecoration(
                  hintText: global.language("search_debtor_hint"),
                  prefixIcon: Icon(Icons.search),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  suffixIcon: IconButton(
                    icon: Icon(Icons.search),
                    onPressed: _searchDebtors,
                  ),
                ),
                onSubmitted: (_) => _searchDebtors(),
              ),
            ),

            // Error Message
            if (_errorMessage != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: global.theme.negativeHighlightColor,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: global.theme.negativeHighlightTextColor.withValues(alpha: 0.3)),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.error_outline, color: global.theme.negativeHighlightTextColor),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _errorMessage!,
                          style: TextStyle(color: global.theme.negativeHighlightTextColor),
                        ),
                      ),
                    ],
                  ),
                ),
              ),

            // Loading Indicator
            if (_isLoading)
              const Expanded(
                child: Center(
                  child: CircularProgressIndicator(),
                ),
              ),

            // Results List
            if (!_isLoading && _hasSearched)
              Expanded(
                child: _debtors.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.search_off,
                              size: 64,
                              color: global.theme.iconSecondaryColor,
                            ),
                            SizedBox(height: 16),
                            Text(
                              global.language("debtor_not_found"),
                              style: TextStyle(
                                fontSize: 16,
                                color: global.theme.iconSecondaryColor,
                              ),
                            ),
                          ],
                        ),
                      )
                    : ListView.builder(
                        itemCount: _debtors.length,
                        itemBuilder: (context, index) {
                          final debtor = _debtors[index];
                          return ListTile(
                            leading: CircleAvatar(
                              backgroundColor: global.theme.infoHighlightColor,
                              child: Text(
                                debtor.code.isNotEmpty
                                    ? debtor.code[0].toUpperCase()
                                    : '?',
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  color: global.theme.infoHighlightTextColor,
                                ),
                              ),
                            ),
                            title: Text(debtor.displayName),
                            subtitle: Text('${global.language("code")}: ${debtor.code}'),
                            trailing: Icon(Icons.arrow_forward_ios, size: 16),
                            onTap: () {
                              Navigator.pop(context, debtor);
                            },
                          );
                        },
                      ),
              ),

            // Initial State
            if (!_isLoading && !_hasSearched)
              Expanded(
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.person_search,
                        size: 64,
                        color: global.theme.iconSecondaryColor,
                      ),
                      SizedBox(height: 16),
                      Text(
                        global.language("enter_keyword_to_search_debtor"),
                        style: TextStyle(
                          fontSize: 16,
                          color: global.theme.iconSecondaryColor,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}
