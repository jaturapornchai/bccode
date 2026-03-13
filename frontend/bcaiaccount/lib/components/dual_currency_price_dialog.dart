import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../global.dart' as global;

/// ผลลัพธ์จาก DualCurrencyPriceDialog
/// เก็บราคาทั้ง 2 สกุลเงิน
class DualCurrencyPriceResult {
  final double docPrice;
  final double basePrice;

  DualCurrencyPriceResult({
    required this.docPrice,
    required this.basePrice,
  });
}

/// Dialog สำหรับแก้ไขราคาสินค้าแบบ 2 สกุลเงินพร้อมกัน
///
/// แสดงช่องป้อนราคา 2 ช่อง (doc currency + base currency)
/// พร้อมปุ่มกดแปลงค่าระหว่างกัน (ไม่ auto-calculate)
/// ผู้ใช้สามารถป้อนค่าได้อิสระทั้ง 2 ช่อง แล้วกดปุ่มแปลงเมื่อต้องการ
class DualCurrencyPriceDialog extends StatefulWidget {
  final String title;
  final double initialDocPrice;
  final double initialBasePrice;
  final double exchangeRate;
  final String docCurrencyCode;
  final String docCurrencySymbol;
  final String baseCurrencyCode;
  final String baseCurrencySymbol;

  const DualCurrencyPriceDialog({
    super.key,
    required this.title,
    required this.initialDocPrice,
    required this.initialBasePrice,
    required this.exchangeRate,
    required this.docCurrencyCode,
    required this.docCurrencySymbol,
    required this.baseCurrencyCode,
    required this.baseCurrencySymbol,
  });

  @override
  State<DualCurrencyPriceDialog> createState() =>
      _DualCurrencyPriceDialogState();
}

enum _ActiveField { doc, base }

class _DualCurrencyPriceDialogState extends State<DualCurrencyPriceDialog> with global.ThemeRefreshMixin {
  String _docValue = '';
  String _baseValue = '';
  _ActiveField _activeField = _ActiveField.doc;

  late TextEditingController _docTextController;
  late TextEditingController _baseTextController;
  late FocusNode _docFocusNode;
  late FocusNode _baseFocusNode;
  bool _isDocEditing = false;
  bool _isBaseEditing = false;

  @override
  void initState() {
    super.initState();
    _docValue = _cleanInitialValue(
      widget.initialDocPrice > 0 ? widget.initialDocPrice.toString() : '',
    );
    _baseValue = _cleanInitialValue(
      widget.initialBasePrice > 0 ? widget.initialBasePrice.toString() : '',
    );
    _docTextController = TextEditingController(text: _docValue);
    _baseTextController = TextEditingController(text: _baseValue);
    _docFocusNode = FocusNode();
    _baseFocusNode = FocusNode();
  }

  /// ปัดเศษค่าเริ่มต้นตามทศนิยมสาขา
  String _cleanInitialValue(String value) {
    if (value.isEmpty) return value;
    final parsed = double.tryParse(value);
    if (parsed == null) return value;
    // จำนวนเต็ม → ไม่ต้องมีทศนิยม
    if (parsed == parsed.roundToDouble() && !value.contains('.')) {
      return parsed.round().toString();
    }
    // ปัดเศษทศนิยมตามค่าสาขา
    final dp = global.getDecimalPrice();
    final result = parsed.toStringAsFixed(dp);
    if (result.endsWith('0')) {
      final trimmed = result.replaceAll(RegExp(r'0+$'), '');
      return trimmed.endsWith('.')
          ? trimmed.substring(0, trimmed.length - 1)
          : trimmed;
    }
    return result;
  }

  @override
  void dispose() {
    _docTextController.dispose();
    _baseTextController.dispose();
    _docFocusNode.dispose();
    _baseFocusNode.dispose();
    super.dispose();
  }

  /// ค่า active field ปัจจุบัน
  String get _activeValue =>
      _activeField == _ActiveField.doc ? _docValue : _baseValue;

  set _activeValue(String val) {
    if (_activeField == _ActiveField.doc) {
      _docValue = val;
    } else {
      _baseValue = val;
    }
  }

  /// กดปุ่มแปลงค่า → คำนวณอีกฝั่งจาก active field
  void _convertCurrency() {
    final rate = widget.exchangeRate;
    if (rate <= 0) return;

    setState(() {
      if (_activeField == _ActiveField.doc) {
        // doc → base: base = doc × rate
        final docPrice = double.tryParse(_docValue) ?? 0;
        _baseValue = docPrice > 0 ? _formatForStorage(docPrice * rate) : '';
      } else {
        // base → doc: doc = base / rate
        final basePrice = double.tryParse(_baseValue) ?? 0;
        _docValue = basePrice > 0 ? _formatForStorage(basePrice / rate) : '';
      }
    });
  }

  /// Format ตัวเลขเก็บค่าภายใน (ไม่มี comma)
  String _formatForStorage(double value) {
    if (value == value.roundToDouble()) {
      return value.round().toString();
    }
    // ตัดทศนิยมตามค่าสาขา
    final dp = global.getDecimalPrice();
    final result = value.toStringAsFixed(dp);
    // ลบ trailing zeros
    if (result.endsWith('0')) {
      final trimmed = result.replaceAll(RegExp(r'0+$'), '');
      return trimmed.endsWith('.') ? trimmed.substring(0, trimmed.length - 1) : trimmed;
    }
    return result;
  }

  void _onNumericKeyTap(String value) {
    setState(() {
      String current = _activeValue;
      if (value == '.') {
        if (!current.contains('.')) {
          current = current.isEmpty ? '0.' : '$current.';
        }
      } else {
        if (current.length < 12) {
          if (current == '0' && value != '.') {
            current = value;
          } else {
            final newValue = current + value;
            final parts = newValue.split('.');
            if (parts.length <= 2) {
              final dp = global.getDecimalPrice();
              if (parts.length == 2 && parts[1].length > dp) return;
              current = newValue;
            }
          }
        }
      }
      _activeValue = current;
    });
  }

  void _onClear() {
    setState(() {
      _activeValue = '';
    });
  }

  void _onBackspace() {
    setState(() {
      String current = _activeValue;
      if (current.isNotEmpty) {
        _activeValue = current.substring(0, current.length - 1);
      }
    });
  }

  void _switchActiveField(_ActiveField field) {
    if (_activeField == field) return;
    setState(() {
      // ออกจากโหมดแก้ไข text ก่อน
      _isDocEditing = false;
      _isBaseEditing = false;
      _docFocusNode.unfocus();
      _baseFocusNode.unfocus();
      _activeField = field;
    });
  }

  void _toggleDocTextEdit() {
    setState(() {
      _activeField = _ActiveField.doc;
      _isDocEditing = !_isDocEditing;
      _isBaseEditing = false;
      _baseFocusNode.unfocus();
      if (_isDocEditing) {
        _docTextController.text = _docValue;
        _docFocusNode.requestFocus();
      } else {
        _docFocusNode.unfocus();
      }
    });
  }

  void _toggleBaseTextEdit() {
    setState(() {
      _activeField = _ActiveField.base;
      _isBaseEditing = !_isBaseEditing;
      _isDocEditing = false;
      _docFocusNode.unfocus();
      if (_isBaseEditing) {
        _baseTextController.text = _baseValue;
        _baseFocusNode.requestFocus();
      } else {
        _baseFocusNode.unfocus();
      }
    });
  }

  void _onDocTextChanged(String value) {
    setState(() {
      _docValue = value;
      _activeField = _ActiveField.doc;
    });
  }

  void _onBaseTextChanged(String value) {
    setState(() {
      _baseValue = value;
      _activeField = _ActiveField.base;
    });
  }

  void _confirmInput() async {
    // อ่านค่าจาก text controller ถ้ากำลังแก้ไขอยู่
    if (_isDocEditing) _docValue = _docTextController.text;
    if (_isBaseEditing) _baseValue = _baseTextController.text;

    final docPrice = double.tryParse(_docValue) ?? 0;
    final basePrice = double.tryParse(_baseValue) ?? 0;

    // ตรวจสอบว่าราคา 2 สกุลเงินตรงกับอัตราแลกเปลี่ยนหรือไม่ (เกิน 1% → ไม่ยอม)
    if (docPrice > 0 && basePrice > 0 && widget.exchangeRate > 0) {
      final expectedBase = docPrice * widget.exchangeRate;
      final diff = (basePrice - expectedBase).abs();
      final diffPercent = diff / expectedBase * 100;

      if (diffPercent > 1.0) {
        await showDialog<void>(
          context: context,
          builder: (ctx) => AlertDialog(
            title: Row(
              children: [
                Icon(Icons.warning_amber_rounded, color: global.theme.warningHighlightTextColor),
                SizedBox(width: 8),
                Text(global.language('warning')),
              ],
            ),
            content: Text(
              '${widget.docCurrencyCode} ${_formatDisplayValue(docPrice.toString())}'
              ' × ${widget.exchangeRate}'
              ' = ${_formatDisplayValue(expectedBase.toStringAsFixed(global.getDecimalPrice()))}'
              ' ${widget.baseCurrencyCode}\n\n'
              '${widget.baseCurrencyCode}'
              ' ${_formatDisplayValue(basePrice.toString())}'
              ' (diff ${diffPercent.toStringAsFixed(1)}%)',
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: Text(global.language('ok')),
              ),
            ],
          ),
        );
        return; // กลับไปแก้ไข ไม่ยอมให้ผ่าน
      }
    }

    if (!mounted) return;
    Navigator.of(context).pop(
      DualCurrencyPriceResult(docPrice: docPrice, basePrice: basePrice),
    );
  }

  String _formatDisplayValue(String value) {
    if (value.isEmpty) return '0';
    try {
      final parts = value.split('.');
      final integerPart = parts[0];
      final decimalPart = parts.length > 1 ? '.${parts[1]}' : '';
      final formatter = RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))');
      final formattedInteger = integerPart.replaceAllMapped(
        formatter,
        (Match m) => '${m[1]},',
      );
      return formattedInteger + decimalPart;
    } catch (e) {
      return value;
    }
  }

  @override
  Widget build(BuildContext context) {
    // ข้อความปุ่มแปลงตามทิศทาง active field
    final convertLabel = _activeField == _ActiveField.doc
        ? '${widget.docCurrencyCode} → ${widget.baseCurrencyCode}'
        : '${widget.baseCurrencyCode} → ${widget.docCurrencyCode}';

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
      elevation: 0,
      backgroundColor: Colors.transparent,
      child: Container(
        width: 340,
        decoration: BoxDecoration(
          color: global.theme.textColor,
          borderRadius: BorderRadius.circular(28),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.3),
              spreadRadius: 0,
              blurRadius: 20,
              offset: const Offset(0, 10),
            ),
          ],
        ),
        child: Padding(
          padding: const EdgeInsets.all(10),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Title
              Text(
                widget.title,
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w500,
                  color: global.theme.onPrimaryColor.withValues(alpha: 0.7),
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 8),

              // === Doc Currency Display ===
              _buildCurrencyDisplay(
                field: _ActiveField.doc,
                currencyCode: widget.docCurrencyCode,
                currencySymbol: widget.docCurrencySymbol,
                value: _docValue,
                isEditing: _isDocEditing,
                controller: _docTextController,
                focusNode: _docFocusNode,
                onTap: () => _switchActiveField(_ActiveField.doc),
                onDoubleTap: _toggleDocTextEdit,
                onTextChanged: _onDocTextChanged,
              ),

              // === ปุ่มแปลงสกุลเงิน ===
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 4),
                child: SizedBox(
                  height: 32,
                  child: ElevatedButton.icon(
                    onPressed: _convertCurrency,
                    icon: Icon(
                      Icons.swap_vert,
                      color: global.theme.warningHighlightTextColor,
                      size: 18,
                    ),
                    label: Text(
                      '$convertLabel  ×${_formatDisplayValue(widget.exchangeRate.toString())}',
                      style: TextStyle(
                        fontSize: 12,
                        color: global.theme.warningHighlightTextColor,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.appBarColor,
                      elevation: 0,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                        side: BorderSide(color: global.theme.warningHighlightTextColor, width: 1),
                      ),
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                    ),
                  ),
                ),
              ),

              // === Base Currency Display ===
              _buildCurrencyDisplay(
                field: _ActiveField.base,
                currencyCode: widget.baseCurrencyCode,
                currencySymbol: widget.baseCurrencySymbol,
                value: _baseValue,
                isEditing: _isBaseEditing,
                controller: _baseTextController,
                focusNode: _baseFocusNode,
                onTap: () => _switchActiveField(_ActiveField.base),
                onDoubleTap: _toggleBaseTextEdit,
                onTextChanged: _onBaseTextChanged,
              ),

              const SizedBox(height: 8),

              // === Calculator Keyboard ===
              for (final row in [
                ['7', '8', '9'],
                ['4', '5', '6'],
                ['1', '2', '3'],
                ['0', '.', '⌫'],
              ])
                Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: Row(
                    children: row.map((key) {
                      return Expanded(
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 2),
                          child: _buildCalcButton(
                            key,
                            key == '⌫'
                                ? _onBackspace
                                : () => _onNumericKeyTap(key),
                            isFunction: key == '⌫',
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                ),

              const SizedBox(height: 4),

              // Action buttons row
              Row(
                children: [
                  Expanded(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 2),
                      child: _buildActionButton(
                        'C',
                        global.theme.dividerBorderColor,
                        global.theme.textColor,
                        _onClear,
                      ),
                    ),
                  ),
                  Expanded(
                    child: Padding(
                      padding: EdgeInsets.symmetric(horizontal: 2),
                      child: _buildActionButton(
                        global.language('cancel'),
                        global.theme.appBarColor,
                        global.theme.onPrimaryColor.withValues(alpha: 0.7),
                        () => Navigator.of(context).pop(),
                      ),
                    ),
                  ),
                  Expanded(
                    child: Padding(
                      padding: EdgeInsets.symmetric(horizontal: 2),
                      child: _buildActionButton(
                        global.language('ok'),
                        global.theme.primaryColor,
                        global.theme.onPrimaryColor,
                        _confirmInput,
                      ),
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

  /// สร้างช่องแสดงราคาสกุลเงิน
  Widget _buildCurrencyDisplay({
    required _ActiveField field,
    required String currencyCode,
    required String currencySymbol,
    required String value,
    required bool isEditing,
    required TextEditingController controller,
    required FocusNode focusNode,
    required VoidCallback onTap,
    required VoidCallback onDoubleTap,
    required ValueChanged<String> onTextChanged,
  }) {
    final isActive = _activeField == field;
    final borderColor = isActive ? global.theme.warningHighlightTextColor : Colors.transparent;
    final bgColor = isActive ? global.theme.cardColor : global.theme.dividerBorderColor;
    final labelColor = isActive
        ? global.theme.warningHighlightTextColor
        : global.theme.textSecondaryColor;

    return GestureDetector(
      onTap: onTap,
      onDoubleTap: onDoubleTap,
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: bgColor,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: borderColor, width: 2),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            // Currency label
            Row(
              children: [
                Text(
                  '$currencySymbol $currencyCode',
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: labelColor,
                  ),
                ),
                if (isActive)
                  Padding(
                    padding: const EdgeInsets.only(left: 6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 1),
                      decoration: BoxDecoration(
                        color: global.theme.infoHighlightColor,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Text(
                        '✎',
                        style: TextStyle(
                          fontSize: 9,
                          color: global.theme.warningHighlightTextColor,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 2),
            // Price value
            SizedBox(
              height: 40,
              child: Align(
                alignment: Alignment.centerRight,
                child: isEditing
                    ? TextField(
                        controller: controller,
                        focusNode: focusNode,
                        keyboardType: const TextInputType.numberWithOptions(
                            decimal: true),
                        inputFormatters: [
                          FilteringTextInputFormatter.allow(
                              RegExp(r'[\d.]')),
                        ],
                        textAlign: TextAlign.right,
                        style: TextStyle(
                          fontSize: 28,
                          fontWeight: FontWeight.w300,
                          color: isActive ? global.theme.textColor : global.theme.textSecondaryColor,
                        ),
                        decoration: const InputDecoration(
                          border: InputBorder.none,
                          contentPadding: EdgeInsets.zero,
                        ),
                        onChanged: onTextChanged,
                      )
                    : Text(
                        value.isEmpty
                            ? '0'
                            : _formatDisplayValue(value),
                        style: TextStyle(
                          fontSize: 28,
                          fontWeight: FontWeight.w300,
                          color: isActive ? global.theme.textColor : global.theme.textSecondaryColor,
                        ),
                        textAlign: TextAlign.right,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCalcButton(String text, VoidCallback onTap,
      {bool isFunction = false}) {
    final buttonColor =
        isFunction ? global.theme.dividerBorderColor : global.theme.appBarColor;
    final textColor = isFunction ? global.theme.textColor : global.theme.onPrimaryColor;

    return SizedBox(
      height: 80,
      width: 80,
      child: ElevatedButton(
        onPressed: onTap,
        style: ElevatedButton.styleFrom(
          backgroundColor: buttonColor,
          foregroundColor: textColor,
          elevation: 0,
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          padding: EdgeInsets.zero,
        ),
        child: FittedBox(
          fit: BoxFit.scaleDown,
          child: Text(
            text,
            style: TextStyle(
              fontSize: isFunction ? 20 : 24,
              fontWeight: FontWeight.w400,
              color: textColor,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildActionButton(
    String text,
    Color backgroundColor,
    Color textColor,
    VoidCallback onTap,
  ) {
    return SizedBox(
      height: 60,
      child: ElevatedButton(
        onPressed: onTap,
        style: ElevatedButton.styleFrom(
          backgroundColor: backgroundColor,
          foregroundColor: textColor,
          elevation: 0,
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          padding: EdgeInsets.zero,
        ),
        child: FittedBox(
          fit: BoxFit.scaleDown,
          child: Text(
            text,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w500,
              color: textColor,
            ),
          ),
        ),
      ),
    );
  }
}

/// แสดง Dialog แก้ไขราคาแบบ 2 สกุลเงิน
/// คืนค่า [DualCurrencyPriceResult] หรือ null ถ้ายกเลิก
Future<DualCurrencyPriceResult?> showDualCurrencyPriceDialog(
  BuildContext context, {
  required String title,
  required double initialDocPrice,
  required double initialBasePrice,
  required double exchangeRate,
  required String docCurrencyCode,
  required String docCurrencySymbol,
  required String baseCurrencyCode,
  required String baseCurrencySymbol,
}) async {
  return await showDialog<DualCurrencyPriceResult>(
    context: context,
    barrierDismissible: false,
    builder: (BuildContext context) {
      return DualCurrencyPriceDialog(
        title: title,
        initialDocPrice: initialDocPrice,
        initialBasePrice: initialBasePrice,
        exchangeRate: exchangeRate,
        docCurrencyCode: docCurrencyCode,
        docCurrencySymbol: docCurrencySymbol,
        baseCurrencyCode: baseCurrencyCode,
        baseCurrencySymbol: baseCurrencySymbol,
      );
    },
  );
}
