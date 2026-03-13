import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/wht_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// WHT Section Widget - สำหรับจัดการภาษีหัก ณ ที่จ่าย (รองรับหลายรายการ)
class WHTSectionWidget extends StatelessWidget {
  final TransactionModel screenData;
  final Function(void Function()) setState;
  final BuildContext context;

  const WHTSectionWidget({
    super.key,
    required this.screenData,
    required this.setState,
    required this.context,
  });

  static Color get _primaryColor => global.theme.primaryColor;

  @override
  Widget build(BuildContext context) {
    // Initialize whtEntries if null
    screenData.whtEntries ??= [];

    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header with Add button
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Icon(Icons.receipt_long_outlined, color: _primaryColor, size: 20),
                  SizedBox(width: 8),
                  Text(
                    global.language("wht_withholding_tax"),
                    style: TextStyle(
                      color: global.theme.textColor,
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                ],
              ),
              // Add button
              ElevatedButton.icon(
                onPressed: () => _showAddWHTDialog(context),
                icon: Icon(Icons.add, size: 18),
                label: Text(global.language("add")),
                style: ElevatedButton.styleFrom(
                  backgroundColor: _primaryColor,
                  foregroundColor: global.theme.onPrimaryColor,
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // List of WHT entries
          if (screenData.whtEntries!.isEmpty)
            Center(
              child: Padding(
                padding: EdgeInsets.all(16),
                child: Text(
                  global.language("no_wht_entries"),
                  style: TextStyle(
                    color: global.theme.textSecondaryColor,
                    fontSize: 14,
                  ),
                ),
              ),
            )
          else
            ...screenData.whtEntries!.asMap().entries.map((entry) {
              final index = entry.key;
              final whtEntry = entry.value;
              return _buildWHTEntryCard(context, index, whtEntry);
            }),

          // Total WHT Amount
          if (screenData.whtEntries!.isNotEmpty) ...[
            Divider(height: 24),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  global.language("total_wht"),
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    // แสดงยอดในสกุลเงินเอกสาร (Document Currency)
                    if (screenData.currency != null &&
                        screenData.currency!.isNotEmpty &&
                        screenData.currency!.toUpperCase() != global.getBaseCurrency() &&
                        screenData.exchangerate != null &&
                        screenData.exchangerate! > 0) ...[
                      Text(
                        global.formatNumberWithCurrency(
                          _calculateTotalWHT() / screenData.exchangerate!,
                          currency: screenData.currency,
                          currencySymbol: screenData.currencysymbol,
                        ),
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                          color: _primaryColor,
                        ),
                      ),
                      // แสดงยอดในสกุลเงินหลัก (Base Currency)
                      Text(
                        global.formatNumber(_calculateTotalWHT()),
                        style: TextStyle(
                          fontSize: 12,
                          color: global.theme.textSecondaryColor,
                          fontWeight: FontWeight.w400,
                        ),
                      ),
                    ] else ...[
                      // แสดงแค่สกุลเงินเดียว
                      Text(
                        global.formatNumber(_calculateTotalWHT()),
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                          color: _primaryColor,
                        ),
                      ),
                    ],
                  ],
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  /// สร้าง card สำหรับแสดง WHT entry แต่ละรายการ
  Widget _buildWHTEntryCard(BuildContext context, int index, WHTEntryModel whtEntry) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
        side: BorderSide(color: global.theme.dividerBorderColor),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            // Income type icon
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: _primaryColor.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                Icons.receipt_outlined,
                color: _primaryColor,
                size: 20,
              ),
            ),
            const SizedBox(width: 12),

            // Details
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    whtEntry.description,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 14,
                    ),
                  ),
                  SizedBox(height: 4),
                  Row(
                    children: [
                      Text(
                        "${global.language("rate")}: ${whtEntry.rate.toStringAsFixed(1)}%",
                        style: TextStyle(
                          fontSize: 12,
                          color: global.theme.textColor,
                        ),
                      ),
                      const SizedBox(width: 12),
                      // แสดง tax base พร้อม dual currency
                      Builder(
                        builder: (context) {
                          final bool hasMultiCurrency = screenData.currency != null &&
                              screenData.currency!.isNotEmpty &&
                              screenData.currency!.toUpperCase() != global.getBaseCurrency() &&
                              screenData.exchangerate != null &&
                              screenData.exchangerate! > 0;
                          final String taxBaseText = hasMultiCurrency
                              ? '${global.formatNumberWithCurrency(whtEntry.taxbase / screenData.exchangerate!, currency: screenData.currency, currencySymbol: screenData.currencysymbol)} (${global.formatNumber(whtEntry.taxbase)})'
                              : global.formatNumber(whtEntry.taxbase);
                          return Text(
                            "${global.language("tax_base_short")}: $taxBaseText",
                            style: TextStyle(
                              fontSize: 12,
                              color: global.theme.textColor,
                            ),
                          );
                        },
                      ),
                      const SizedBox(width: 12),
                      // แสดง amount พร้อม dual currency
                      Builder(
                        builder: (context) {
                          final bool hasMultiCurrency = screenData.currency != null &&
                              screenData.currency!.isNotEmpty &&
                              screenData.currency!.toUpperCase() != global.getBaseCurrency() &&
                              screenData.exchangerate != null &&
                              screenData.exchangerate! > 0;
                          final String amountText = hasMultiCurrency
                              ? '${global.formatNumberWithCurrency(whtEntry.amount / screenData.exchangerate!, currency: screenData.currency, currencySymbol: screenData.currencysymbol)} (${global.formatNumber(whtEntry.amount)})'
                              : global.formatNumber(whtEntry.amount);
                          return Text(
                            "${global.language("amount")}: $amountText",
                            style: TextStyle(
                              fontSize: 12,
                              color: global.theme.textColor,
                              fontWeight: FontWeight.w600,
                            ),
                          );
                        },
                      ),
                    ],
                  ),
                  // Show note if exists
                  if (whtEntry.note != null && whtEntry.note!.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(
                      whtEntry.note!,
                      style: TextStyle(
                        fontSize: 12,
                        color: global.theme.textSecondaryColor,
                        fontStyle: FontStyle.italic,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),

            // Edit and Delete buttons
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  onPressed: () => _showEditWHTDialog(context, index, whtEntry),
                  icon: Icon(Icons.edit_outlined, size: 20),
                  color: global.theme.infoHighlightTextColor,
                  tooltip: global.language("edit"),
                ),
                IconButton(
                  onPressed: () => _confirmDeleteWHT(context, index),
                  icon: Icon(Icons.delete_outline, size: 20),
                  color: global.theme.negativeHighlightTextColor,
                  tooltip: global.language("delete"),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  /// คำนวณยอดรวม WHT
  double _calculateTotalWHT() {
    double total = 0.0;
    for (var entry in screenData.whtEntries!) {
      total += entry.amount;
    }
    return total;
  }

  /// แสดง dialog สำหรับเพิ่ม WHT entry ใหม่
  void _showAddWHTDialog(BuildContext context) {
    _showWHTDialog(context, null, null);
  }

  /// แสดง dialog สำหรับแก้ไข WHT entry
  void _showEditWHTDialog(BuildContext context, int index, WHTEntryModel whtEntry) {
    _showWHTDialog(context, index, whtEntry);
  }

  /// แสดง dialog สำหรับเพิ่ม/แก้ไข WHT entry
  void _showWHTDialog(BuildContext context, int? editIndex, WHTEntryModel? existingEntry) {
    String? selectedIncomeTypeCode = existingEntry?.description != null
        ? WHTIncomeType.values
            .firstWhere(
              (type) => type.nameTH == existingEntry!.description || type.nameEN == existingEntry.description,
              orElse: () => WHTIncomeType.values.first,
            )
            .code
        : null;

    // ตรวจสอบว่าเป็น multi-currency หรือไม่
    final bool hasMultiCurrency = screenData.currency != null &&
        screenData.currency!.isNotEmpty &&
        screenData.currency!.toUpperCase() != global.getBaseCurrency() &&
        screenData.exchangerate != null &&
        screenData.exchangerate! > 0;

    final rateController = TextEditingController(
      text: existingEntry?.rate.toStringAsFixed(1) ?? '0.0',
    );
    // สำหรับ multi-currency ให้แปลงเป็น document currency ก่อนแสดง
    final taxBaseController = TextEditingController(
      text: existingEntry != null
          ? (hasMultiCurrency
              ? (existingEntry.taxbase / screenData.exchangerate!).toStringAsFixed(2)
              : existingEntry.taxbase.toStringAsFixed(2))
          : '',
    );
    final amountController = TextEditingController(
      text: existingEntry != null
          ? (hasMultiCurrency
              ? (existingEntry.amount / screenData.exchangerate!).toStringAsFixed(2)
              : existingEntry.amount.toStringAsFixed(2))
          : '',
    );
    final noteController = TextEditingController(
      text: existingEntry?.note ?? '',
    );

    // Function to auto-calculate amount from taxBase and rate
    void autoCalculateAmount() {
      final rate = double.tryParse(rateController.text) ?? 0.0;
      final taxBase = double.tryParse(taxBaseController.text) ?? 0.0;
      if (rate > 0 && taxBase > 0) {
        final calculatedAmount = taxBase * (rate / 100);
        amountController.text = calculatedAmount.toStringAsFixed(2);
      }
    }

    // Function to auto-calculate taxBase from amount and rate
    void autoCalculateTaxBase() {
      final rate = double.tryParse(rateController.text) ?? 0.0;
      final amount = double.tryParse(amountController.text) ?? 0.0;
      if (rate > 0 && amount > 0) {
        final calculatedTaxBase = amount / (rate / 100);
        taxBaseController.text = calculatedTaxBase.toStringAsFixed(2);
      }
    }

    showDialog(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) {
          // Get selected income type
          WHTIncomeType? selectedType = selectedIncomeTypeCode != null
              ? WHTIncomeType.findByCode(selectedIncomeTypeCode!)
              : null;

          return AlertDialog(
            title: Text(
              editIndex == null ? global.language("add_wht") : global.language("edit_wht"),
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            content: SingleChildScrollView(
              child: SizedBox(
                width: MediaQuery.of(context).size.width * 0.9,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Header: ประเภทเงินได้
                    Text(
                      global.language("income_type"),
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 14,
                      ),
                    ),
                    const SizedBox(height: 12),

                    // Radio buttons for income types
                    ...WHTIncomeType.values.map((type) {
                      final isSelected = selectedIncomeTypeCode == type.code;
                      return InkWell(
                        onTap: () {
                          setDialogState(() {
                            selectedIncomeTypeCode = type.code;
                            rateController.text = type.defaultRate.toStringAsFixed(1);
                            // Set taxBase เป็น document currency ถ้า hasMultiCurrency, ไม่งั้นเป็น base currency
                            final double taxBaseDefault = hasMultiCurrency && screenData.totalamountDoc != null
                                ? screenData.totalamountDoc!
                                : screenData.totalamount;
                            taxBaseController.text = taxBaseDefault.toStringAsFixed(2);
                            autoCalculateAmount();
                          });
                        },
                        child: Container(
                          margin: const EdgeInsets.only(bottom: 8),
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: isSelected
                                ? _primaryColor.withValues(alpha: 0.1)
                                : global.theme.surfaceColor,
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(
                              color: isSelected ? _primaryColor : global.theme.dividerBorderColor,
                              width: isSelected ? 2 : 1,
                            ),
                          ),
                          child: Row(
                            children: [
                              Icon(
                                isSelected
                                    ? Icons.radio_button_checked
                                    : Icons.radio_button_unchecked,
                                color: isSelected ? _primaryColor : global.theme.iconSecondaryColor,
                                size: 20,
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      type.nameTH,
                                      style: TextStyle(
                                        fontSize: 14,
                                        fontWeight: isSelected ? FontWeight.w600 : FontWeight.w500,
                                        color: isSelected ? _primaryColor : global.theme.textColor,
                                      ),
                                    ),
                                    Text(
                                      "${type.nameEN} (${type.defaultRate}%)",
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: global.theme.textSecondaryColor,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    }),

                    const SizedBox(height: 16),

                    // Rate input
                    TextFormField(
                      controller: rateController,
                      decoration: InputDecoration(
                        labelText: "${global.language("wht_rate")} (%)",
                        prefixIcon: Icon(Icons.percent_outlined),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
                      ),
                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                      inputFormatters: [
                        FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
                      ],
                      onChanged: (value) {
                        // Auto-calculate amount when rate changes
                        autoCalculateAmount();
                      },
                    ),

                    const SizedBox(height: 12),

                    // Tax Base input (editable)
                    TextFormField(
                      controller: taxBaseController,
                      decoration: InputDecoration(
                        labelText: hasMultiCurrency
                            ? '${global.language("tax_base")} (${screenData.currency})'
                            : global.language("tax_base"),
                        prefixIcon: Icon(Icons.calculate_outlined),
                        hintText: global.language("tax_base_hint"),
                        // แสดง helper text สำหรับ dual currency
                        // เมื่อเป็น multi-currency: user พิมพ์ document currency, helper แสดง base currency
                        helperText: () {
                          final bool hasMultiCurrency = screenData.currency != null &&
                              screenData.currency!.isNotEmpty &&
                              screenData.currency!.toUpperCase() != global.getBaseCurrency() &&
                              screenData.exchangerate != null &&
                              screenData.exchangerate! > 0;
                          if (hasMultiCurrency && taxBaseController.text.isNotEmpty) {
                            final double docTaxBase = double.tryParse(taxBaseController.text) ?? 0.0;
                            if (docTaxBase > 0) {
                              final double baseTaxBase = docTaxBase * screenData.exchangerate!;
                              return '≈ ${global.formatNumber(baseTaxBase)}';
                            }
                          }
                          return null;
                        }(),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
                      ),
                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                      inputFormatters: [
                        FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
                      ],
                      onChanged: (value) {
                        // Auto-calculate amount when tax base changes
                        autoCalculateAmount();
                        // Update helper text
                        setDialogState(() {});
                      },
                    ),

                    const SizedBox(height: 12),

                    // Amount input (auto-calculated, can be edited)
                    TextFormField(
                      controller: amountController,
                      decoration: InputDecoration(
                        labelText: hasMultiCurrency
                            ? '${global.language("wht_amount")} (${screenData.currency})'
                            : global.language("wht_amount"),
                        prefixIcon: Icon(Icons.money_off_outlined),
                        suffixIcon: Icon(Icons.info_outline, size: 18, color: global.theme.iconSecondaryColor),
                        // แสดง helper text สำหรับ dual currency
                        // เมื่อเป็น multi-currency: user พิมพ์ document currency, helper แสดง base currency
                        helperText: () {
                          final bool hasMultiCurrency = screenData.currency != null &&
                              screenData.currency!.isNotEmpty &&
                              screenData.currency!.toUpperCase() != global.getBaseCurrency() &&
                              screenData.exchangerate != null &&
                              screenData.exchangerate! > 0;
                          if (hasMultiCurrency && amountController.text.isNotEmpty) {
                            final double docAmount = double.tryParse(amountController.text) ?? 0.0;
                            if (docAmount > 0) {
                              final double baseAmount = docAmount * screenData.exchangerate!;
                              return '≈ ${global.formatNumber(baseAmount)} | ${global.language("auto_calculated_editable")}';
                            }
                          }
                          return global.language("auto_calculated_editable");
                        }(),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
                      ),
                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                      inputFormatters: [
                        FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
                      ],
                      onChanged: (value) {
                        // Auto-calculate tax base when amount changes (reverse calculation)
                        autoCalculateTaxBase();
                        // Update helper text
                        setDialogState(() {});
                      },
                    ),

                    const SizedBox(height: 12),

                    // Note input (optional)
                    TextFormField(
                      controller: noteController,
                      decoration: InputDecoration(
                        labelText: global.language("note_optional"),
                        prefixIcon: Icon(Icons.note_outlined),
                        hintText: global.language("note_hint"),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
                      ),
                      maxLines: 2,
                      maxLength: 200,
                    ),
                  ],
                ),
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(dialogContext).pop(),
                child: Text(global.language("cancel")),
              ),
              ElevatedButton(
                onPressed: () {
                  if (selectedIncomeTypeCode == null) {
                    // Show error
                    global.showInfoSnackBar(context, global.language("please_select_income_type"));
                    return;
                  }

                  final rate = double.tryParse(rateController.text) ?? 0.0;
                  // อ่านค่าจาก controller (เป็น document currency ถ้า hasMultiCurrency)
                  final taxBaseInput = double.tryParse(taxBaseController.text) ?? 0.0;
                  final amountInput = double.tryParse(amountController.text) ?? 0.0;

                  if (taxBaseInput <= 0) {
                    global.showInfoSnackBar(context, global.language("please_enter_tax_base"));
                    return;
                  }

                  if (amountInput <= 0) {
                    global.showInfoSnackBar(context, global.language("please_enter_amount"));
                    return;
                  }

                  // แปลงเป็น base currency ก่อนเก็บใน model
                  final taxBase = hasMultiCurrency ? taxBaseInput * screenData.exchangerate! : taxBaseInput;
                  final amount = hasMultiCurrency ? amountInput * screenData.exchangerate! : amountInput;

                  final selectedType = WHTIncomeType.findByCode(selectedIncomeTypeCode!);
                  final note = noteController.text.trim();

                  final newEntry = WHTEntryModel(
                    description: selectedType!.nameTH,
                    rate: rate,
                    taxbase: taxBase,
                    amount: amount,
                    note: note.isEmpty ? null : note,
                  );

                  setState(() {
                    if (editIndex == null) {
                      // Add new
                      screenData.whtEntries!.add(newEntry);
                    } else {
                      // Edit existing
                      screenData.whtEntries![editIndex] = newEntry;
                    }
                  });

                  Navigator.of(dialogContext).pop();
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: _primaryColor,
                  foregroundColor: global.theme.onPrimaryColor,
                ),
                child: Text(global.language("save")),
              ),
            ],
          );
        },
      ),
    );
  }

  /// ยืนยันการลบ WHT entry
  void _confirmDeleteWHT(BuildContext context, int index) {
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(global.language("confirm_delete")),
        content: Text(global.language("confirm_delete_wht_entry")),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: Text(global.language("cancel")),
          ),
          ElevatedButton(
            onPressed: () {
              setState(() {
                screenData.whtEntries!.removeAt(index);
              });
              Navigator.of(dialogContext).pop();
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.negativeHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            child: Text(global.language("delete")),
          ),
        ],
      ),
    );
  }
}
