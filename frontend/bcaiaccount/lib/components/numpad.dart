import 'package:flutter/material.dart';
import '../global.dart' as global;

enum ButtonType { number, operator, function }

// NumPad Dialog for Mobile/Tablet and Desktop devices
class NumPadDialog extends StatefulWidget {
  final String title;
  final String? initialValue;
  final int maxLength;
  final int decimalPlaces;
  final bool allowNegative;
  final bool forceNumericKeyboard;
  final VoidCallback? onCancel;
  final Function(String)? onConfirm;

  const NumPadDialog({
    super.key,
    required this.title,
    this.initialValue,
    this.maxLength = 10,
    this.decimalPlaces = 2,
    this.allowNegative = false,
    this.forceNumericKeyboard = false,
    this.onCancel,
    this.onConfirm,
  });

  @override
  State<NumPadDialog> createState() => _NumPadDialogState();
}

class _NumPadDialogState extends State<NumPadDialog> {
  String _displayValue = '';
  late TextEditingController _textController;
  late FocusNode _focusNode;
  bool _isEditing = false;

  @override
  void initState() {
    super.initState();
    _displayValue = _cleanInitialValue(widget.initialValue ?? '');
    _textController = TextEditingController(text: _displayValue);
    _focusNode = FocusNode();
  }

  String _cleanInitialValue(String value) {
    if (value.isEmpty) return value;

    // ลบ .0 หรือ .00 ออกจากตัวเลข
    if (value.endsWith('.0')) {
      return value.substring(0, value.length - 2);
    } else if (value.endsWith('.00')) {
      return value.substring(0, value.length - 3);
    }

    return value;
  }

  @override
  void dispose() {
    _textController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _onNumericKeyTap(String value) {
    setState(() {
      if (value == '.') {
        if (!_displayValue.contains('.') && widget.decimalPlaces > 0) {
          _displayValue = _displayValue.isEmpty ? '0.' : '$_displayValue.';
        }
      } else if (value == '%') {
        if (_displayValue.isNotEmpty) {
          final doubleValue = double.tryParse(_displayValue);
          if (doubleValue != null) {
            _displayValue = (doubleValue / 100).toString();
          }
        }
      } else {
        if (_displayValue.length < widget.maxLength) {
          if (_displayValue == '0' && value != '.') {
            _displayValue = value;
          } else {
            final newValue = _displayValue + value;
            final parts = newValue.split('.');
            if (parts.length <= 2) {
              if (parts.length == 2 && parts[1].length > widget.decimalPlaces) {
                return;
              }
              _displayValue = newValue;
            }
          }
        } else {
          final newValue = _displayValue + value;
          final parts = newValue.split('.');
          if (parts.length <= 2) {
            if (parts.length == 2 && parts[1].length <= widget.decimalPlaces) {
              _displayValue = newValue;
            }
          } else {
            _displayValue = newValue;
          }
        }
      }
    });
  }

  void _onTextFieldChanged(String value) {
    setState(() {
      if (value.isNotEmpty) {
        final parts = value.split('.');
        if (parts.length <= 2) {
          if (parts.length == 2 && parts[1].length > widget.decimalPlaces) {
            return;
          } else {
            _displayValue = value;
          }
        } else {
          _displayValue = value;
        }
      } else {
        _displayValue = '';
      }
    });
  }

  void _onClear() {
    setState(() {
      _displayValue = '';
    });
  }

  void _onBackspace() {
    setState(() {
      if (_displayValue.isNotEmpty) {
        _displayValue = _displayValue.substring(0, _displayValue.length - 1);
      }
    });
  }

  void _toggleEditMode() {
    setState(() {
      _isEditing = !_isEditing;
      if (_isEditing) {
        _textController.text = _displayValue;
        _focusNode.requestFocus();
      } else {
        _focusNode.unfocus();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
      elevation: 0,
      backgroundColor: Colors.transparent,
      child: Container(
        width: 320,
        decoration: BoxDecoration(
          color: Colors.black,
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
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w500,
                  color: Colors.white70,
                ),
              ),

              const SizedBox(height: 10),

              // Display - Calculator style
              Container(
                width: double.infinity,
                height: 80,
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 16,
                ),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: GestureDetector(
                  onTap: _toggleEditMode,
                  child: Container(
                    alignment: Alignment.centerRight,
                    child: _isEditing
                        ? TextField(
                            controller: _textController,
                            focusNode: _focusNode,
                            keyboardType: const TextInputType.numberWithOptions(
                              decimal: true,
                            ),
                            inputFormatters: [
                              global.NumberInputFormatter(
                                maximumFractionDigits: widget.decimalPlaces,
                              ),
                            ],
                            textAlign: TextAlign.right,
                            style: const TextStyle(
                              fontSize: 36,
                              fontWeight: FontWeight.w300,
                              color: Colors.black,
                            ),
                            decoration: const InputDecoration(
                              border: InputBorder.none,
                              contentPadding: EdgeInsets.zero,
                            ),
                            onChanged: _onTextFieldChanged,
                            onSubmitted: (value) => _toggleEditMode(),
                          )
                        : Text(
                            _displayValue.isEmpty
                                ? '0'
                                : _formatDisplayValue(_displayValue),
                            style: const TextStyle(
                              fontSize: 36,
                              fontWeight: FontWeight.w300,
                              color: Colors.black,
                            ),
                            textAlign: TextAlign.right,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                  ),
                ),
              ),

              const SizedBox(height: 10),

              // Calculator keyboard layout - iOS style using Grid
              SizedBox(
                height: 400, // เพิ่มความสูงเพื่อรองรับแถวเพิ่มเติม
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 2.0),
                  child: Column(
                    children: [
                      // Main numpad grid (4 rows x 3 columns)
                      Expanded(
                        flex: 4,
                        child: GridView.count(
                          crossAxisCount: 3,
                          crossAxisSpacing: 4,
                          mainAxisSpacing: 4,
                          childAspectRatio: 1.2,
                          physics: const NeverScrollableScrollPhysics(),
                          children: [
                            // Row 1: 7, 8, 9
                            _buildCalculatorButton(
                              '7',
                              ButtonType.number,
                              () => _onNumericKeyTap('7'),
                            ),
                            _buildCalculatorButton(
                              '8',
                              ButtonType.number,
                              () => _onNumericKeyTap('8'),
                            ),
                            _buildCalculatorButton(
                              '9',
                              ButtonType.number,
                              () => _onNumericKeyTap('9'),
                            ),

                            // Row 2: 4, 5, 6
                            _buildCalculatorButton(
                              '4',
                              ButtonType.number,
                              () => _onNumericKeyTap('4'),
                            ),
                            _buildCalculatorButton(
                              '5',
                              ButtonType.number,
                              () => _onNumericKeyTap('5'),
                            ),
                            _buildCalculatorButton(
                              '6',
                              ButtonType.number,
                              () => _onNumericKeyTap('6'),
                            ),

                            // Row 3: 1, 2, 3
                            _buildCalculatorButton(
                              '1',
                              ButtonType.number,
                              () => _onNumericKeyTap('1'),
                            ),
                            _buildCalculatorButton(
                              '2',
                              ButtonType.number,
                              () => _onNumericKeyTap('2'),
                            ),
                            _buildCalculatorButton(
                              '3',
                              ButtonType.number,
                              () => _onNumericKeyTap('3'),
                            ),

                            // Row 4: 0, ., ⌫
                            _buildCalculatorButton(
                              '0',
                              ButtonType.number,
                              () => _onNumericKeyTap('0'),
                            ),
                            widget.decimalPlaces > 0
                                ? _buildCalculatorButton(
                                    '.',
                                    ButtonType.number,
                                    () => _onNumericKeyTap('.'),
                                  )
                                : Container(),
                            _buildCalculatorButton(
                              '⌫',
                              ButtonType.function,
                              _onBackspace,
                            ),
                          ],
                        ),
                      ),

                      const SizedBox(height: 8),

                      // Action buttons row (1 row x 3 columns)
                      SizedBox(
                        height: 60,
                        child: Row(
                          children: [
                            // Clear button
                            Expanded(
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 2,
                                ),
                                child: _buildActionButton(
                                  'C',
                                  const Color(0xFFA6A6A6),
                                  Colors.black,
                                  _onClear,
                                ),
                              ),
                            ),
                            // Cancel button
                            Expanded(
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 2,
                                ),
                                child: _buildActionButton(
                                  global.language('cancel'),
                                  const Color(0xFF333333),
                                  Colors.white70,
                                  () {
                                    if (widget.onCancel != null) {
                                      widget.onCancel!();
                                    }
                                    Navigator.of(context).pop();
                                  },
                                ),
                              ),
                            ),
                            // OK button
                            Expanded(
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 2,
                                ),
                                child: _buildActionButton(
                                  global.language('ok'),
                                  const Color(0xFFFF9500),
                                  Colors.white,
                                  () => _confirmInput(),
                                ),
                              ),
                            ),
                          ],
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
    );
  }

  void _confirmInput() {
    if (_isEditing) {
      _displayValue = _textController.text;
    }
    if (widget.onConfirm != null) {
      widget.onConfirm!(_displayValue);
    }
    Navigator.of(context).pop(_displayValue);
  }

  String _formatDisplayValue(String value) {
    // Format large numbers with commas for better readability
    if (value.isEmpty) return '0';

    try {
      final parts = value.split('.');
      final integerPart = parts[0];
      final decimalPart = parts.length > 1 ? '.${parts[1]}' : '';

      // Add commas to integer part
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

  Widget _buildCalculatorButton(
    String text,
    ButtonType type,
    VoidCallback? onTap,
  ) {
    Color buttonColor;
    Color textColor;

    switch (type) {
      case ButtonType.number:
        buttonColor = const Color(0xFF333333);
        textColor = Colors.white;
        break;
      case ButtonType.operator:
        buttonColor = const Color(0xFFFF9500);
        textColor = Colors.white;
        break;
      case ButtonType.function:
        buttonColor = const Color(0xFFA6A6A6);
        textColor = Colors.black;
        break;
    }

    return SizedBox(
      height: 80,
      width: 80,
      child: ElevatedButton(
        onPressed: onTap,
        style: ElevatedButton.styleFrom(
          backgroundColor: buttonColor,
          foregroundColor: textColor,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          padding: EdgeInsets.zero,
        ),
        child: FittedBox(
          fit: BoxFit.scaleDown,
          child: Text(
            text,
            style: TextStyle(
              fontSize: type == ButtonType.function ? 20 : 24,
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
    VoidCallback? onTap,
  ) {
    return SizedBox(
      height: 60,
      child: ElevatedButton(
        onPressed: onTap,
        style: ElevatedButton.styleFrom(
          backgroundColor: backgroundColor,
          foregroundColor: textColor,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
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

// Helper function to show numpad dialog
Future<String?> showNumPadDialog(
  BuildContext context, {
  required String title,
  String? initialValue,
  int maxLength = 10,
  int decimalPlaces = 2,
  bool allowNegative = false,
  bool forceNumericKeyboard = false,
}) async {
  return await showDialog<String>(
    context: context,
    barrierDismissible: false,
    builder: (BuildContext context) {
      return NumPadDialog(
        title: title,
        initialValue: initialValue,
        maxLength: maxLength,
        decimalPlaces: decimalPlaces,
        allowNegative: allowNegative,
        forceNumericKeyboard: forceNumericKeyboard,
      );
    },
  );
}

// Alias for backward compatibility
Future<String?> showNumericInputDialog(
  BuildContext context, {
  required String title,
  String? initialValue,
  int maxLength = 10,
  int decimalPlaces = 2,
  bool allowNegative = false,
  bool forceNumericKeyboard = false,
}) async {
  return await showNumPadDialog(
    context,
    title: title,
    initialValue: initialValue,
    maxLength: maxLength,
    decimalPlaces: decimalPlaces,
    allowNegative: allowNegative,
    forceNumericKeyboard: forceNumericKeyboard,
  );
}
