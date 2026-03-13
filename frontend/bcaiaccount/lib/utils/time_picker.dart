import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;

/// Custom Time Picker Widget ที่มี popup ให้เลือกชั่วโมงและนาที
/// ✨ Modern Design with Gradient & Glassmorphism
class CustomTimePicker extends StatefulWidget {
  final ValueChanged<TimeOfDay?>? onTimeSelected;
  final String? labelText;
  final bool useIconSelectTime;
  final TimeOfDay? initialTime;
  final TextEditingController? controller;
  final String? languageCode; // "th" หรือ "en", ถ้าเป็น null จะอ่านจาก global.getBranchLanguage()

  const CustomTimePicker({
    super.key,
    this.onTimeSelected,
    this.labelText,
    this.useIconSelectTime = true,
    this.initialTime,
    this.controller,
    this.languageCode, // null = อ่านจาก global.getBranchLanguage()
  });

  @override
  State<CustomTimePicker> createState() => _CustomTimePickerState();
}

class _CustomTimePickerState extends State<CustomTimePicker> with global.ThemeRefreshMixin {
  late TextEditingController _controller;
  TimeOfDay? _selectedTime;
  bool isTimeValid = true;
  final FocusNode _focusNode = FocusNode();
  bool _isExternalController = false;
  final GlobalKey _timeIconKey = GlobalKey();

  // 🎨 Modern Color Palette (same as date_picker)
  static Color get _primaryColor => global.theme.primaryColor;

  // ดึงภาษาที่ใช้ในการแสดงผล (default จาก global.getBranchLanguage())
  String get _languageCode => widget.languageCode ?? global.getBranchLanguage();

  @override
  void initState() {
    super.initState();
    _isExternalController = widget.controller != null;
    _controller = widget.controller ?? TextEditingController();
    _selectedTime = widget.initialTime ?? TimeOfDay.now();

    // ถ้าใช้ external controller และมีค่าอยู่แล้ว ให้ parse เวลาจาก text
    if (_isExternalController && _controller.text.isNotEmpty) {
      _parseTimeFromText(_controller.text);
    } else {
      _updateDisplayTime();
    }

    _focusNode.addListener(() {
      if (!_focusNode.hasFocus) {
        _completeTime(_controller.text);
      }
    });
  }

  @override
  void didUpdateWidget(covariant CustomTimePicker oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.initialTime != null &&
        widget.initialTime != oldWidget.initialTime &&
        widget.initialTime != _selectedTime) {
      setState(() {
        _selectedTime = widget.initialTime;
        _updateDisplayTime();
      });
    }
  }

  @override
  void dispose() {
    if (!_isExternalController) {
      _controller.dispose();
    }
    _focusNode.dispose();
    super.dispose();
  }

  void _parseTimeFromText(String text) {
    final time = _parseTime(text);
    if (time != null) {
      _selectedTime = time;
    }
  }

  void _updateDisplayTime() {
    if (_selectedTime != null) {
      _controller.text =
          '${_selectedTime!.hour.toString().padLeft(2, '0')}:${_selectedTime!.minute.toString().padLeft(2, '0')}';
    }
  }

  TimeOfDay? _parseTime(String text) {
    // รองรับ format: HH:mm, H:mm, HH.mm, H.mm
    String timeOnly = text.trim();

    List<String> parts = timeOnly
        .replaceAll(".", ":")
        .replaceAll("-", ":")
        .split(':');

    if (parts.isEmpty || parts.length > 2) return null;

    int hour = int.tryParse(parts[0]) ?? -1;
    int minute = parts.length > 1 ? (int.tryParse(parts[1]) ?? 0) : 0;

    if (hour < 0 || hour > 23 || minute < 0 || minute > 59) return null;

    return TimeOfDay(hour: hour, minute: minute);
  }

  void _handleInput(String value) {
    // อนุญาตให้พิมพ์เฉพาะตัวเลขและ :
    setState(() {
      isTimeValid = true;
    });
  }

  void _completeTime(String text) {
    final time = _parseTime(text);
    if (time != null) {
      setState(() {
        _selectedTime = time;
        isTimeValid = true;
        _updateDisplayTime();
      });
      if (widget.onTimeSelected != null) {
        widget.onTimeSelected!(_selectedTime);
      }
    } else if (text.isNotEmpty) {
      setState(() {
        isTimeValid = false;
      });
    }
  }

  Future<void> _showTimePicker() async {
    final RenderBox? iconBox =
        _timeIconKey.currentContext?.findRenderObject() as RenderBox?;
    if (iconBox == null) return;

    final iconPosition = iconBox.localToGlobal(Offset.zero);
    final iconSize = iconBox.size;

    // Use same calculation as date_picker.dart
    final double left = iconPosition.dx - 320 + iconSize.width;
    final double top = iconPosition.dy + iconSize.height + 8;
    final screenWidth = MediaQuery.of(context).size.width;
    final screenHeight = MediaQuery.of(context).size.height;
    final double adjustedLeft = left + 360 > screenWidth
        ? screenWidth - 380
        : (left < 20 ? 20 : left);
    final double adjustedTop = top + 520 > screenHeight
        ? iconPosition.dy - 520
        : top;

    final TimeOfDay? picked = await showDialog<TimeOfDay>(
      context: context,
      barrierColor: Colors.black.withValues(alpha: 0.4),
      builder: (BuildContext context) {
        return Stack(
          children: [
            Positioned(
              left: adjustedLeft,
              top: adjustedTop,
              child: Material(
                color: Colors.transparent,
                child: CustomTimePickerDialog(
                  initialTime: _selectedTime ?? TimeOfDay.now(),
                  languageCode: _languageCode,
                ),
              ),
            ),
          ],
        );
      },
    );

    if (picked != null) {
      setState(() {
        _selectedTime = picked;
        isTimeValid = true;
        _updateDisplayTime();
      });
      if (widget.onTimeSelected != null) {
        widget.onTimeSelected!(_selectedTime);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: _controller,
      focusNode: _focusNode,
      keyboardType: TextInputType.datetime,
      inputFormatters: [
        FilteringTextInputFormatter.allow(RegExp(r'[0-9:.\-]')),
        LengthLimitingTextInputFormatter(5),
      ],
      onChanged: _handleInput,
      onSubmitted: (_) => _completeTime(_controller.text),
      style: TextStyle(fontSize: 14, color: global.theme.formTextColor),
      decoration: InputDecoration(
        labelText: widget.labelText,
        floatingLabelBehavior: FloatingLabelBehavior.always,
        floatingLabelStyle: TextStyle(
          backgroundColor: global.theme.cardColor,
          color: global.theme.formLabelColor,
          fontWeight: FontWeight.bold,
          fontSize: 14,
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 12,
          vertical: 14,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: BorderSide(color: global.theme.formBorderColor),
        ),
        suffixIcon: widget.useIconSelectTime
            ? IconButton(
                key: _timeIconKey,
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(
                  Icons.access_time,
                  color: global.theme.iconSecondaryColor,
                  size: 20,
                ),
                onPressed: _showTimePicker,
              )
            : null,
        filled: true,
        fillColor: isTimeValid ? global.theme.formFillColor : global.theme.negativeHighlightColor,
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: _primaryColor, width: 2.0),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: global.theme.formBorderColor),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: global.theme.negativeHighlightTextColor, width: 1.5),
        ),
      ),
    );
  }
}

// ✨ Modern Time Picker Dialog with Gradient Header (เหมือน DatePicker)
class CustomTimePickerDialog extends StatefulWidget {
  final TimeOfDay initialTime;
  final String languageCode; // "th" หรือ "en"

  const CustomTimePickerDialog({
    super.key,
    required this.initialTime,
    required this.languageCode,
  });

  @override
  State<CustomTimePickerDialog> createState() => _CustomTimePickerDialogState();
}

class _CustomTimePickerDialogState extends State<CustomTimePickerDialog>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  late int _selectedHour;
  late int _selectedMinute;
  late FixedExtentScrollController _hourController;
  late FixedExtentScrollController _minuteController;

  // 🎨 Modern Color Palette
  static Color get _primaryColor => global.theme.primaryColor;
  static Color get _accentColor => global.theme.primaryLightColor;
  Color get _surfaceColor => global.theme.surfaceColor;

  // ดึงภาษาที่ใช้ในการแสดงผล
  String get _languageCode => widget.languageCode;

  late AnimationController _animController;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _selectedHour = widget.initialTime.hour;
    _selectedMinute = widget.initialTime.minute;
    _hourController = FixedExtentScrollController(initialItem: _selectedHour);
    _minuteController = FixedExtentScrollController(
      initialItem: _selectedMinute,
    );

    _animController = AnimationController(
      duration: const Duration(milliseconds: 200),
      vsync: this,
    );
    _scaleAnimation = CurvedAnimation(
      parent: _animController,
      curve: Curves.easeOutBack,
    );
    _animController.forward();
  }

  @override
  void dispose() {
    _animController.dispose();
    _hourController.dispose();
    _minuteController.dispose();
    super.dispose();
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 24, horizontal: 20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [_primaryColor, _accentColor],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.access_time_rounded,
                color: global.theme.onPrimaryColor.withValues(alpha: 0.9),
                size: 24,
              ),
              const SizedBox(width: 8),
              Text(
                _languageCode == 'th' ? 'เลือกเวลา' : 'Select Time',
                style: TextStyle(
                  fontSize: 16,
                  color: global.theme.onPrimaryColor.withValues(alpha: 0.9),
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
            decoration: BoxDecoration(
              color: global.theme.onPrimaryColor.withValues(alpha: 0.2),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: global.theme.onPrimaryColor.withValues(alpha: 0.3),
                width: 1,
              ),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  _selectedHour.toString().padLeft(2, '0'),
                  style: TextStyle(
                    fontSize: 48,
                    fontWeight: FontWeight.bold,
                    color: global.theme.onPrimaryColor,
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  child: Text(
                    ':',
                    style: TextStyle(
                      fontSize: 48,
                      fontWeight: FontWeight.bold,
                      color: global.theme.onPrimaryColor,
                    ),
                  ),
                ),
                Text(
                  _selectedMinute.toString().padLeft(2, '0'),
                  style: TextStyle(
                    fontSize: 48,
                    fontWeight: FontWeight.bold,
                    color: global.theme.onPrimaryColor,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 8),
          Text(
            _languageCode == 'th' ? 'นาฬิกา (24 ชั่วโมง)' : 'Clock (24 hours)',
            style: TextStyle(
              fontSize: 12,
              color: global.theme.onPrimaryColor.withValues(alpha: 0.7),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQuickSelectButtons() {
    final quickTimes = [
            {'label': global.language('midnight'), 'hour': 0, 'minute': 0},
            {'label': global.language('morning'), 'hour': 8, 'minute': 0},
            {'label': global.language('noon'), 'hour': 12, 'minute': 0},
            {'label': global.language('evening'), 'hour': 18, 'minute': 0},
          ];

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.only(left: 4, bottom: 8),
            child: Text(
              global.language('quick_select'),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: global.theme.textSecondaryColor,
              ),
            ),
          ),
          Row(
            children: quickTimes.map((time) {
              final isSelected =
                  _selectedHour == time['hour'] &&
                  _selectedMinute == time['minute'];
              return Expanded(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 3),
                  child: Material(
                    color: Colors.transparent,
                    child: InkWell(
                      onTap: () {
                        setState(() {
                          _selectedHour = time['hour'] as int;
                          _selectedMinute = time['minute'] as int;
                          _hourController.jumpToItem(_selectedHour);
                          _minuteController.jumpToItem(_selectedMinute);
                        });
                      },
                      borderRadius: BorderRadius.circular(10),
                      child: Container(
                        padding: const EdgeInsets.symmetric(vertical: 10),
                        decoration: BoxDecoration(
                          gradient: isSelected
                              ? LinearGradient(
                                  colors: [_primaryColor, _accentColor],
                                  begin: Alignment.topLeft,
                                  end: Alignment.bottomRight,
                                )
                              : null,
                          color: isSelected ? null : _surfaceColor,
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(
                            color: isSelected
                                ? Colors.transparent
                                : global.theme.dividerBorderColor,
                          ),
                        ),
                        child: Column(
                          children: [
                            Text(
                              time['label'] as String,
                              style: TextStyle(
                                fontSize: 11,
                                fontWeight: FontWeight.w600,
                                color: isSelected
                                    ? global.theme.onPrimaryColor
                                    : global.theme.textColor,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              '${(time['hour'] as int).toString().padLeft(2, '0')}:${(time['minute'] as int).toString().padLeft(2, '0')}',
                              style: TextStyle(
                                fontSize: 10,
                                color: isSelected
                                    ? global.theme.onPrimaryColor.withValues(alpha: 0.8)
                                    : global.theme.formHintColor,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildTimeWheel() {
    return Container(
      padding: EdgeInsets.symmetric(horizontal: 20, vertical: 8),
      child: Row(
        children: [
          // Hour Selector
          Expanded(
            child: _buildTimeColumn(
              label: _languageCode == 'th' ? global.language('hours') : 'Hours',
              value: _selectedHour,
              maxValue: 23,
              onIncrement: () {
                setState(() {
                  _selectedHour = (_selectedHour + 1) % 24;
                  _hourController.jumpToItem(_selectedHour);
                });
              },
              onDecrement: () {
                setState(() {
                  _selectedHour = (_selectedHour - 1 + 24) % 24;
                  _hourController.jumpToItem(_selectedHour);
                });
              },
            ),
          ),
          // Separator
          Container(
            padding: const EdgeInsets.only(top: 24),
            child: Column(
              children: [
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: _primaryColor,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(height: 12),
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: _primaryColor,
                    shape: BoxShape.circle,
                  ),
                ),
              ],
            ),
          ),
          // Minute Selector
          Expanded(
            child: _buildTimeColumn(
              label: global.language('minutes'),
              value: _selectedMinute,
              maxValue: 59,
              onIncrement: () {
                setState(() {
                  _selectedMinute = (_selectedMinute + 1) % 60;
                  _minuteController.jumpToItem(_selectedMinute);
                });
              },
              onDecrement: () {
                setState(() {
                  _selectedMinute = (_selectedMinute - 1 + 60) % 60;
                  _minuteController.jumpToItem(_selectedMinute);
                });
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTimeColumn({
    required String label,
    required int value,
    required int maxValue,
    required VoidCallback onIncrement,
    required VoidCallback onDecrement,
  }) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          label,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: global.theme.textSecondaryColor,
          ),
        ),
        const SizedBox(height: 10),
        // Up Button
        Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onIncrement,
            onLongPress: () {
              // Long press to increment by 5
              for (int i = 0; i < 5; i++) {
                onIncrement();
              }
            },
            borderRadius: BorderRadius.circular(14),
            child: Container(
              width: 100,
              height: 48,
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    _primaryColor.withValues(alpha: 0.1),
                    _accentColor.withValues(alpha: 0.1),
                  ],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(14),
                border: Border.all(color: _primaryColor.withValues(alpha: 0.2)),
              ),
              child: Icon(
                Icons.keyboard_arrow_up_rounded,
                size: 36,
                color: _primaryColor,
              ),
            ),
          ),
        ),
        const SizedBox(height: 10),
        // Value Display with touch
        GestureDetector(
          onVerticalDragUpdate: (details) {
            if (details.delta.dy < -3) {
              onIncrement();
            } else if (details.delta.dy > 3) {
              onDecrement();
            }
          },
          child: Container(
            width: 100,
            height: 70,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [_primaryColor, _accentColor],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(18),
              boxShadow: [
                BoxShadow(
                  color: _primaryColor.withValues(alpha: 0.4),
                  blurRadius: 12,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            alignment: Alignment.center,
            child: Text(
              value.toString().padLeft(2, '0'),
              style: TextStyle(
                fontSize: 36,
                fontWeight: FontWeight.bold,
                color: global.theme.onPrimaryColor,
              ),
            ),
          ),
        ),
        const SizedBox(height: 10),
        // Down Button
        Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onDecrement,
            onLongPress: () {
              // Long press to decrement by 5
              for (int i = 0; i < 5; i++) {
                onDecrement();
              }
            },
            borderRadius: BorderRadius.circular(14),
            child: Container(
              width: 100,
              height: 48,
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    _primaryColor.withValues(alpha: 0.1),
                    _accentColor.withValues(alpha: 0.1),
                  ],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(14),
                border: Border.all(color: _primaryColor.withValues(alpha: 0.2)),
              ),
              child: Icon(
                Icons.keyboard_arrow_down_rounded,
                size: 36,
                color: _primaryColor,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildActionButtons() {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
      child: Row(
        children: [
          // Now Button
          Expanded(
            child: TextButton.icon(
              onPressed: () {
                final now = TimeOfDay.now();
                setState(() {
                  _selectedHour = now.hour;
                  _selectedMinute = now.minute;
                  _hourController.jumpToItem(_selectedHour);
                  _minuteController.jumpToItem(_selectedMinute);
                });
              },
              icon: Icon(Icons.schedule, size: 18, color: _primaryColor),
              label: Text(
                _languageCode == 'th' ? 'ตอนนี้' : 'Now',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: _primaryColor,
                ),
              ),
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: BorderSide(color: _primaryColor.withValues(alpha: 0.3)),
                ),
              ),
            ),
          ),
          const SizedBox(width: 10),
          // Cancel Button
          Expanded(
            child: TextButton(
              onPressed: () => Navigator.of(context).pop(),
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: BorderSide(color: global.theme.dividerBorderColor),
                ),
              ),
              child: Text(
                _languageCode == 'th' ? global.language('cancel') : 'Cancel',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textSecondaryColor,
                ),
              ),
            ),
          ),
          const SizedBox(width: 10),
          // Confirm Button
          Expanded(
            flex: 2,
            child: Container(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [_primaryColor, _accentColor],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(12),
                boxShadow: [
                  BoxShadow(
                    color: _primaryColor.withValues(alpha: 0.4),
                    blurRadius: 8,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: TextButton.icon(
                onPressed: () {
                  Navigator.of(context).pop(
                    TimeOfDay(hour: _selectedHour, minute: _selectedMinute),
                  );
                },
                icon: Icon(
                  Icons.check_rounded,
                  size: 20,
                  color: global.theme.onPrimaryColor,
                ),
                label: Text(
                  _languageCode == 'th' ? global.language('ok') : 'OK',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: global.theme.onPrimaryColor,
                  ),
                ),
                style: TextButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return ScaleTransition(
      scale: _scaleAnimation,
      child: Container(
        width: 340,
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          borderRadius: BorderRadius.circular(20),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.15),
              blurRadius: 20,
              offset: const Offset(0, 10),
            ),
            BoxShadow(
              color: _primaryColor.withValues(alpha: 0.1),
              blurRadius: 40,
              offset: const Offset(0, 20),
            ),
          ],
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _buildHeader(),
              _buildQuickSelectButtons(),
              _buildTimeWheel(),
              _buildActionButtons(),
            ],
          ),
        ),
      ),
    );
  }
}
