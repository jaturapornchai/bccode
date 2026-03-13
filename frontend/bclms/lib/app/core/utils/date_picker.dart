import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/core/utils/global.dart' as global;

/// Custom Date Picker Widget สำหรับ bclms — ใช้ปี พ.ศ. + ภาษาไทย
/// Modern Design ตาม BC AI Cloud standard
class BclmsDatePicker extends StatefulWidget {
  final ValueChanged<DateTime?>? onDateSelected;
  final String? labelText;
  final bool useIconSelectDate;
  final DateTime? initialDate;
  final DateTime? firstDate;
  final DateTime? lastDate;

  const BclmsDatePicker({
    super.key,
    this.onDateSelected,
    this.labelText,
    this.useIconSelectDate = true,
    this.initialDate,
    this.firstDate,
    this.lastDate,
  });

  @override
  State<BclmsDatePicker> createState() => _BclmsDatePickerState();
}

class _BclmsDatePickerState extends State<BclmsDatePicker> {
  final TextEditingController _controller = TextEditingController();
  DateTime? _selectedDate;
  final GlobalKey _calendarIconKey = GlobalKey();
  bool isDateValid = true;
  final FocusNode _focusNode = FocusNode();
  OverlayEntry? _overlayEntry;

  @override
  void initState() {
    super.initState();
    _selectedDate = widget.initialDate ?? DateTime.now();
    _updateDisplayDate();
    _focusNode.addListener(() {
      setState(() => _updateDisplayDate());
      if (!_focusNode.hasFocus) {
        _completeDate(_controller.text);
      }
    });
  }

  @override
  void didUpdateWidget(covariant BclmsDatePicker oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.initialDate != null &&
        widget.initialDate != oldWidget.initialDate &&
        widget.initialDate != _selectedDate) {
      setState(() {
        _selectedDate = widget.initialDate;
        _updateDisplayDate();
      });
    }
  }

  void _removeOverlay() {
    _overlayEntry?.remove();
    _overlayEntry = null;
  }

  @override
  void dispose() {
    _removeOverlay();
    _controller.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _updateDisplayDate() {
    if (_selectedDate != null) {
      int year = _selectedDate!.year + 543;
      String twoDigitYear = (year % 100).toString().padLeft(2, '0');
      String basicDateFormat =
          '${_selectedDate!.day.toString().padLeft(2, '0')}/${_selectedDate!.month.toString().padLeft(2, '0')}/$twoDigitYear';
      if (_focusNode.hasFocus) {
        _controller.text = basicDateFormat;
      } else {
        String dayName = global.thaiDayName(_selectedDate!);
        String monthName = global.thaiMonthName(_selectedDate!.month);
        _controller.text = '$basicDateFormat ($dayName, $monthName)';
      }
    }
  }

  DateTime? _parseDate(String text) {
    String dateOnly = text.split('(').first.trim();
    List<String> parts = dateOnly.replaceAll(".", "/").replaceAll("-", "/").split('/');
    if (parts.isEmpty || parts.length > 3) return null;

    DateTime now = DateTime.now();
    int currentBuddhistYear = now.year + 543;
    String twoDigitYear = (currentBuddhistYear % 100).toString().padLeft(2, '0');

    String dayStr = parts[0].padLeft(2, '0');
    String monthStr = parts.length > 1 ? parts[1].padLeft(2, '0') : now.month.toString().padLeft(2, '0');
    String yearStr = parts.length > 2 ? parts[2].padLeft(2, '0') : twoDigitYear;

    int day = int.tryParse(dayStr) ?? 0;
    int month = int.tryParse(monthStr) ?? 0;
    int yy = int.tryParse(yearStr) ?? 0;

    if (month < 1 || month > 12 || day < 1 || day > 31) return null;

    int century = (currentBuddhistYear ~/ 100) * 100;
    int fullYear = century + yy;
    if (fullYear > currentBuddhistYear + 50) fullYear -= 100;
    int christianYear = fullYear - 543;

    try {
      DateTime date = DateTime(christianYear, month, day);
      if (date.month == month && date.day == day) return date;
      return null;
    } catch (e) {
      return null;
    }
  }

  void _completeDate(String text) {
    if (text.isEmpty) return;
    DateTime? parsedDate = _parseDate(text);
    setState(() {
      if (parsedDate != null) {
        _selectedDate = parsedDate;
        isDateValid = true;
        _updateDisplayDate();
        widget.onDateSelected?.call(_selectedDate);
      } else {
        isDateValid = false;
      }
    });
  }

  void _handleInput(String text) {
    if (text.length > 8) {
      _controller.value = _controller.value.copyWith(
        text: text.substring(0, 8),
        selection: const TextSelection.collapsed(offset: 8),
      );
      return;
    }
    _controller.value = _controller.value.copyWith(
      text: text,
      selection: TextSelection.collapsed(offset: text.length),
    );
    setState(() => isDateValid = true);
  }

  Future<void> _showCustomDatePicker() async {
    if (_overlayEntry != null) {
      _removeOverlay();
      return;
    }

    DateTime initialDate = _selectedDate ?? DateTime.now();
    DateTime now = DateTime.now();
    DateTime firstDate = widget.firstDate ?? DateTime(now.year - 10, 1, 1);
    DateTime lastDate = widget.lastDate ?? DateTime(now.year + 10, 12, 31);

    if (initialDate.isBefore(firstDate)) initialDate = firstDate;
    if (initialDate.isAfter(lastDate)) initialDate = lastDate;

    final RenderBox? iconBox = _calendarIconKey.currentContext?.findRenderObject() as RenderBox?;
    if (iconBox == null) return;

    final iconPos = iconBox.localToGlobal(Offset.zero);
    final iconSize = iconBox.size;
    final screenSize = MediaQuery.of(context).size;

    const double popupWidth = 360;
    const double popupHeight = 460;

    double left = iconPos.dx + iconSize.width - popupWidth;
    if (left < 8) left = 8;
    if (left + popupWidth > screenSize.width - 8) left = screenSize.width - popupWidth - 8;

    final double iconBottom = iconPos.dy + iconSize.height;
    final bool showAbove = (iconBottom + popupHeight + 8) > screenSize.height;
    double top = showAbove ? (iconPos.dy - popupHeight - 8).clamp(8, double.infinity) : iconBottom + 8;

    final overlay = Overlay.of(context);

    _overlayEntry = OverlayEntry(
      builder: (overlayContext) {
        return Stack(
          children: [
            Positioned.fill(
              child: GestureDetector(
                onTap: _removeOverlay,
                behavior: HitTestBehavior.opaque,
                child: Container(color: Colors.black.withValues(alpha: 0.3)),
              ),
            ),
            Positioned(
              left: left,
              top: top,
              child: Material(
                color: Colors.transparent,
                child: _BclmsDatePickerDialog(
                  initialDate: initialDate,
                  firstDate: firstDate,
                  lastDate: lastDate,
                  onDateSelected: (DateTime? pickedDate) {
                    _removeOverlay();
                    if (pickedDate != null && pickedDate != _selectedDate) {
                      setState(() {
                        _selectedDate = pickedDate;
                        isDateValid = true;
                        _updateDisplayDate();
                      });
                      widget.onDateSelected?.call(_selectedDate);
                    }
                  },
                  onCancel: _removeOverlay,
                ),
              ),
            ),
          ],
        );
      },
    );
    overlay.insert(_overlayEntry!);
  }

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: _controller,
      focusNode: _focusNode,
      keyboardType: TextInputType.datetime,
      inputFormatters: [
        FilteringTextInputFormatter.allow(RegExp(r'[0-9/.\-\(\),\s\u0E00-\u0E7F]')),
        LengthLimitingTextInputFormatter(40),
      ],
      onChanged: _handleInput,
      onSubmitted: (_) => _completeDate(_controller.text),
      style: const TextStyle(fontSize: 14, color: Colors.black87),
      decoration: InputDecoration(
        labelText: widget.labelText,
        floatingLabelBehavior: FloatingLabelBehavior.always,
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        suffixIcon: !widget.useIconSelectDate
            ? null
            : IconButton(
                key: _calendarIconKey,
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(Icons.calendar_today, color: Colors.grey.shade600, size: 20),
                onPressed: _showCustomDatePicker,
              ),
        filled: true,
        fillColor: isDateValid ? Colors.white : Colors.red.shade50,
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppTheme.terracotta, width: 2.0),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.grey.shade300),
        ),
      ),
    );
  }
}

// Calendar Dialog
class _BclmsDatePickerDialog extends StatefulWidget {
  final DateTime initialDate;
  final DateTime firstDate;
  final DateTime lastDate;
  final ValueChanged<DateTime?>? onDateSelected;
  final VoidCallback? onCancel;

  const _BclmsDatePickerDialog({
    required this.initialDate,
    required this.firstDate,
    required this.lastDate,
    this.onDateSelected,
    this.onCancel,
  });

  @override
  State<_BclmsDatePickerDialog> createState() => _BclmsDatePickerDialogState();
}

class _BclmsDatePickerDialogState extends State<_BclmsDatePickerDialog>
    with SingleTickerProviderStateMixin {
  late DateTime _currentDate;
  late DateTime _displayedMonth;
  int _displayedYear = 0;
  bool _isYearMode = false;
  bool _isMonthMode = false;

  static const _weekDays = ['อา', 'จ', 'อ', 'พ', 'พฤ', 'ศ', 'ส'];

  late AnimationController _animController;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _currentDate = widget.initialDate;
    _displayedMonth = DateTime(_currentDate.year, _currentDate.month, 1);
    _displayedYear = _currentDate.year + 543;
    _animController = AnimationController(duration: const Duration(milliseconds: 200), vsync: this);
    _scaleAnimation = CurvedAnimation(parent: _animController, curve: Curves.easeOutBack);
    _animController.forward();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  Widget _buildHeader() {
    int yearToDisplay = _displayedMonth.year + 543;
    String monthName = global.thaiMonthName(_displayedMonth.month);

    return Container(
      padding: const EdgeInsets.symmetric(vertical: 20, horizontal: 16),
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          colors: [AppTheme.terracotta, AppTheme.earthBrown],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      child: Column(
        children: [
          Text(
            global.thaiDayName(_currentDate),
            style: TextStyle(fontSize: 14, color: Colors.white.withValues(alpha: 0.8), fontWeight: FontWeight.w500),
          ),
          const SizedBox(height: 4),
          Text(
            '${_currentDate.day} ${global.thaiMonthName(_currentDate.month)} ${_currentDate.year + 543}',
            style: const TextStyle(fontSize: 22, color: Colors.white, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              if (!_isYearMode && !_isMonthMode)
                _buildNavButton(Icons.chevron_left_rounded, () {
                  setState(() => _displayedMonth = DateTime(_displayedMonth.year, _displayedMonth.month - 1, 1));
                }),
              Expanded(
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    _buildHeaderButton(monthName, _isMonthMode, () {
                      setState(() { _isMonthMode = !_isMonthMode; _isYearMode = false; });
                    }),
                    const SizedBox(width: 8),
                    _buildHeaderButton('พ.ศ. $yearToDisplay', _isYearMode, () {
                      setState(() { _isYearMode = !_isYearMode; _isMonthMode = false; });
                    }),
                  ],
                ),
              ),
              if (!_isYearMode && !_isMonthMode)
                _buildNavButton(Icons.chevron_right_rounded, () {
                  setState(() => _displayedMonth = DateTime(_displayedMonth.year, _displayedMonth.month + 1, 1));
                }),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildNavButton(IconData icon, VoidCallback onPressed) {
    return Material(
      color: Colors.white.withValues(alpha: 0.2),
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(8),
        child: Container(padding: const EdgeInsets.all(8), child: Icon(icon, color: Colors.white, size: 24)),
      ),
    );
  }

  Widget _buildHeaderButton(String text, bool isActive, VoidCallback onTap) {
    return Material(
      color: isActive ? Colors.white.withValues(alpha: 0.3) : Colors.transparent,
      borderRadius: BorderRadius.circular(20),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(20),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Text(text, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: Colors.white)),
        ),
      ),
    );
  }

  Widget _buildWeekdayNames() {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: _weekDays.map((day) {
          bool isWeekend = day == 'อา' || day == 'ส';
          return Expanded(
            child: Text(
              day,
              textAlign: TextAlign.center,
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13, color: isWeekend ? AppTheme.terracotta : Colors.grey.shade600),
            ),
          );
        }).toList(),
      ),
    );
  }

  Widget _buildCalendarGrid() {
    final firstDayOfMonth = _displayedMonth;
    final lastDayOfMonth = DateTime(firstDayOfMonth.year, firstDayOfMonth.month + 1, 0);
    final firstWeekday = firstDayOfMonth.weekday % 7;
    final daysInMonth = lastDayOfMonth.day;
    final totalDays = firstWeekday + daysInMonth;
    final rows = (totalDays / 7).ceil();

    List<Widget> dayWidgets = [];
    for (int i = 0; i < firstWeekday; i++) {
      dayWidgets.add(const SizedBox());
    }

    for (int day = 1; day <= daysInMonth; day++) {
      final currentDate = DateTime(_displayedMonth.year, _displayedMonth.month, day);
      bool isSelected = currentDate.year == _currentDate.year && currentDate.month == _currentDate.month && currentDate.day == _currentDate.day;
      bool isToday = currentDate.year == DateTime.now().year && currentDate.month == DateTime.now().month && currentDate.day == DateTime.now().day;
      int weekday = (firstWeekday + day - 1) % 7;
      bool isWeekend = weekday == 0 || weekday == 6;

      dayWidgets.add(
        GestureDetector(
          onTap: () {
            setState(() => _currentDate = currentDate);
            widget.onDateSelected?.call(_currentDate);
          },
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            margin: const EdgeInsets.all(2),
            decoration: BoxDecoration(
              gradient: isSelected
                  ? const LinearGradient(colors: [AppTheme.terracotta, AppTheme.earthBrown], begin: Alignment.topLeft, end: Alignment.bottomRight)
                  : null,
              color: isSelected ? null : (isToday ? AppTheme.terracotta.withValues(alpha: 0.1) : null),
              borderRadius: BorderRadius.circular(10),
              border: isToday && !isSelected ? Border.all(color: AppTheme.terracotta, width: 2) : null,
              boxShadow: isSelected
                  ? [BoxShadow(color: AppTheme.terracotta.withValues(alpha: 0.4), blurRadius: 8, offset: const Offset(0, 2))]
                  : null,
            ),
            alignment: Alignment.center,
            child: Text(
              '$day',
              style: TextStyle(
                color: isSelected ? Colors.white : (isToday ? AppTheme.terracottaDark : (isWeekend ? AppTheme.terracotta : Colors.grey.shade700)),
                fontWeight: isToday || isSelected ? FontWeight.bold : FontWeight.w500,
                fontSize: 14,
              ),
            ),
          ),
        ),
      );
    }

    while (dayWidgets.length < rows * 7) {
      dayWidgets.add(const SizedBox());
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: GridView.count(crossAxisCount: 7, shrinkWrap: true, physics: const NeverScrollableScrollPhysics(), childAspectRatio: 1.1, children: dayWidgets),
    );
  }

  Widget _buildMonthGrid() {
    return Container(
      padding: const EdgeInsets.all(12),
      child: GridView.count(
        crossAxisCount: 3,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        childAspectRatio: 2.0,
        children: List.generate(12, (i) {
          int month = i + 1;
          bool isSelected = month == _displayedMonth.month;
          bool isCurrent = month == DateTime.now().month && _displayedMonth.year == DateTime.now().year;
          return GestureDetector(
            onTap: () => setState(() { _displayedMonth = DateTime(_displayedMonth.year, month, 1); _isMonthMode = false; }),
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              margin: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                gradient: isSelected ? const LinearGradient(colors: [AppTheme.terracotta, AppTheme.earthBrown]) : null,
                color: isSelected ? null : (isCurrent ? AppTheme.terracotta.withValues(alpha: 0.1) : const Color(0xFFF8FAFC)),
                borderRadius: BorderRadius.circular(12),
                border: isCurrent && !isSelected ? Border.all(color: AppTheme.terracotta, width: 2) : null,
              ),
              alignment: Alignment.center,
              child: Text(
                global.thaiMonthShortName(month),
                style: TextStyle(color: isSelected ? Colors.white : Colors.grey.shade700, fontWeight: isSelected || isCurrent ? FontWeight.bold : FontWeight.w500, fontSize: 14),
              ),
            ),
          );
        }),
      ),
    );
  }

  Widget _buildYearGrid() {
    int startYear = _displayedYear - 5;
    return Container(
      padding: const EdgeInsets.all(12),
      child: GridView.count(
        crossAxisCount: 3,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        childAspectRatio: 2.0,
        children: List.generate(12, (i) {
          int year = startYear + i;
          bool isSelected = year == _displayedYear;
          int currentYear = DateTime.now().year + 543;
          bool isCurrent = year == currentYear;
          return GestureDetector(
            onTap: () => setState(() {
              _displayedYear = year;
              _displayedMonth = DateTime(year - 543, _displayedMonth.month, 1);
              _isYearMode = false;
            }),
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              margin: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                gradient: isSelected ? const LinearGradient(colors: [AppTheme.terracotta, AppTheme.earthBrown]) : null,
                color: isSelected ? null : (isCurrent ? AppTheme.terracotta.withValues(alpha: 0.1) : const Color(0xFFF8FAFC)),
                borderRadius: BorderRadius.circular(12),
                border: isCurrent && !isSelected ? Border.all(color: AppTheme.terracotta, width: 2) : null,
              ),
              alignment: Alignment.center,
              child: Text(
                '$year',
                style: TextStyle(color: isSelected ? Colors.white : Colors.grey.shade700, fontWeight: isSelected || isCurrent ? FontWeight.bold : FontWeight.w500, fontSize: 14),
              ),
            ),
          );
        }),
      ),
    );
  }

  Widget _buildNavigationButtons() {
    if (!_isYearMode) return const SizedBox.shrink();
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          _buildYearNavButton(Icons.keyboard_double_arrow_left_rounded, () => setState(() => _displayedYear -= 12)),
          Text('${_displayedYear - 5} - ${_displayedYear + 6}', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Colors.grey.shade600)),
          _buildYearNavButton(Icons.keyboard_double_arrow_right_rounded, () => setState(() => _displayedYear += 12)),
        ],
      ),
    );
  }

  Widget _buildYearNavButton(IconData icon, VoidCallback onPressed) {
    return Material(
      color: const Color(0xFFF8FAFC),
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(8),
        child: Container(padding: const EdgeInsets.all(8), child: Icon(icon, color: AppTheme.terracotta, size: 24)),
      ),
    );
  }

  Widget _buildActionButtons() {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          Expanded(
            child: TextButton(
              onPressed: widget.onCancel ?? () => Navigator.of(context).pop(),
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12), side: BorderSide(color: Colors.grey.shade300)),
              ),
              child: Text('ยกเลิก', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: Colors.grey.shade600)),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                gradient: const LinearGradient(colors: [AppTheme.terracotta, AppTheme.earthBrown]),
                borderRadius: BorderRadius.circular(12),
                boxShadow: [BoxShadow(color: AppTheme.terracotta.withValues(alpha: 0.4), blurRadius: 8, offset: const Offset(0, 2))],
              ),
              child: TextButton(
                onPressed: () {
                  setState(() {
                    _currentDate = DateTime.now();
                    _displayedMonth = DateTime(_currentDate.year, _currentDate.month, 1);
                    _displayedYear = _currentDate.year + 543;
                  });
                  widget.onDateSelected?.call(_currentDate);
                },
                style: TextButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
                child: const Text('วันนี้', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: Colors.white)),
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
        width: 360,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          boxShadow: [
            BoxShadow(color: Colors.black.withValues(alpha: 0.15), blurRadius: 20, offset: const Offset(0, 10)),
            BoxShadow(color: AppTheme.terracotta.withValues(alpha: 0.1), blurRadius: 40, offset: const Offset(0, 20)),
          ],
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _buildHeader(),
              _buildNavigationButtons(),
              if (_isYearMode) _buildYearGrid() else if (_isMonthMode) _buildMonthGrid() else ...[_buildWeekdayNames(), _buildCalendarGrid()],
              _buildActionButtons(),
            ],
          ),
        ),
      ),
    );
  }
}
