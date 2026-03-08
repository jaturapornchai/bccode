import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;

// Date Picker Widget ที่รองรับทั้งปี พ.ศ. และ ค.ศ. และหลายภาษา
// ✨ Modern Design with Gradient & Glassmorphism
class CustomDatePicker extends StatefulWidget {
  final ValueChanged<DateTime?>? onDateSelected;
  final String? labelText;
  final bool? useBuddhistCalendar; // null = อ่านจาก global.getYearType()
  final String? languageCode; // null = อ่านจาก global.getBranchLanguage()
  final bool? useIconSelectDate;
  final DateTime? initialDate;
  final DateTime? firstDate;
  final DateTime? lastDate;

  const CustomDatePicker({
    super.key,
    this.onDateSelected,
    this.labelText,
    this.useBuddhistCalendar, // null = อ่านจาก global.getYearType() อัตโนมัติ
    this.languageCode, // null = อ่านจาก global.getBranchLanguage() อัตโนมัติ
    this.useIconSelectDate = true,
    this.initialDate,
    this.firstDate,
    this.lastDate,
    required InputDecoration decoration,
  });

  @override
  _CustomDatePickerState createState() => _CustomDatePickerState();
}

class _CustomDatePickerState extends State<CustomDatePicker> {
  final TextEditingController _controller = TextEditingController();
  DateTime? _selectedDate;
  final GlobalKey _calendarIconKey = GlobalKey();
  bool isDateValid = true;
  final FocusNode _focusNode = FocusNode();

  // 🎨 Modern Color Palette
  static const Color _primaryColor = Color(0xFF667eea);
  static const Color _primaryDarkColor = Color(0xFF5a67d8);
  static const Color _accentColor = Color(0xFF764ba2);

  // ดึงภาษาที่ใช้ในการแสดงผล (default จาก global.getBranchLanguage())
  String get _languageCode => widget.languageCode ?? global.getBranchLanguage();

  // ดึงประเภทปี (default จาก global.getYearType())
  bool get _useBuddhistCalendar => widget.useBuddhistCalendar ?? (global.getYearType() == 'buddhist');

  @override
  void initState() {
    super.initState();
    _selectedDate = widget.initialDate ?? DateTime.now();
    _updateDisplayDate();
    _focusNode.addListener(() {
      setState(() {
        _updateDisplayDate();
      });
      if (!_focusNode.hasFocus) {
        _completeDate(_controller.text);
      }
    });
  }

  @override
  void didUpdateWidget(covariant CustomDatePicker oldWidget) {
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

  OverlayEntry? _overlayEntry;

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
      int year = _useBuddhistCalendar
          ? _selectedDate!.year + 543
          : _selectedDate!.year;
      String twoDigitYear = (year % 100).toString().padLeft(2, '0');

      String basicDateFormat =
          '${_selectedDate!.day.toString().padLeft(2, '0')}/${_selectedDate!.month.toString().padLeft(2, '0')}/$twoDigitYear';

      if (_focusNode.hasFocus) {
        _controller.text = basicDateFormat;
      } else {
        // ใช้ภาษาตาม languageCode (th หรือ en)
        String dayName = _languageCode == 'th'
            ? global.getThaiDayNameFull(_selectedDate!.weekday)
            : global.getEnglishDayNameFull(_selectedDate!.weekday);
        String monthName = _languageCode == 'th'
            ? global.getThaiMonthNameFull(_selectedDate!.month)
            : global.getEnglishMonthNameFull(_selectedDate!.month);
        _controller.text = '$basicDateFormat ($dayName, $monthName)';
      }
    }
  }

  DateTime? _parseDate(String text) {
    String dateOnly = text.split('(').first.trim();

    List<String> parts = dateOnly
        .replaceAll(".", "/")
        .replaceAll("-", "/")
        .split('/');
    if (parts.isEmpty || parts.length > 3) return null;
    DateTime now = DateTime.now();
    int currentYear = _useBuddhistCalendar ? now.year + 543 : now.year;
    String twoDigitYear = (currentYear % 100).toString().padLeft(2, '0');
    String dayStr = parts[0].padLeft(2, '0');
    String monthStr = parts.length > 1
        ? parts[1].padLeft(2, '0')
        : now.month.toString().padLeft(2, '0');
    String yearStr = parts.length > 2 ? parts[2].padLeft(2, '0') : twoDigitYear;

    int day = int.tryParse(dayStr) ?? 0;
    int month = int.tryParse(monthStr) ?? 0;
    int yy = int.tryParse(yearStr) ?? 0;

    if (month < 1 || month > 12 || day < 1 || day > 31) return null;

    int century = (currentYear ~/ 100) * 100;
    int fullYear = century + yy;
    if (fullYear > currentYear + 50) fullYear -= 100;

    int christianYear = _useBuddhistCalendar ? fullYear - 543 : fullYear;

    DateTime minDate = DateTime(now.year - 10, 1, 1);
    DateTime maxDate = DateTime(now.year + 10, 12, 31);

    try {
      DateTime date = DateTime(christianYear, month, day);
      if (date.month == month && date.day == day) {
        if (date.isBefore(minDate) || date.isAfter(maxDate)) {
          return null;
        }
        return date;
      }
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
        if (widget.onDateSelected != null) {
          widget.onDateSelected!(_selectedDate);
        }
      } else {
        isDateValid = false;
        if (text.isNotEmpty) {
          DateTime now = DateTime.now();
          int currentBuddhistYear = now.year + 543;
          String yearRange = _useBuddhistCalendar
              ? 'พ.ศ. ${currentBuddhistYear - 10} - ${currentBuddhistYear + 10}'
              : 'ค.ศ. ${now.year - 10} - ${now.year + 10}';

          global.showErrorSnackBar(context, '${global.language("date_out_of_range")} $yearRange');
        }
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

    setState(() {
      isDateValid = true;
    });
  }

  Future<void> _showCustomDatePicker() async {
    // ถ้า popup เปิดอยู่แล้ว ให้ปิดก่อน
    if (_overlayEntry != null) {
      _removeOverlay();
      return;
    }

    DateTime initialDate = _selectedDate ?? DateTime.now();
    DateTime now = DateTime.now();
    DateTime firstDate = widget.firstDate ?? DateTime(now.year - 10, 1, 1);
    DateTime lastDate = widget.lastDate ?? DateTime(now.year + 10, 12, 31);

    if (initialDate.isBefore(firstDate)) {
      initialDate = firstDate;
    } else if (initialDate.isAfter(lastDate)) {
      initialDate = lastDate;
    }

    // คำนวณตำแหน่งของ icon ปฏิทินบนหน้าจอ
    final RenderBox? iconBox =
        _calendarIconKey.currentContext?.findRenderObject() as RenderBox?;
    if (iconBox == null) return;

    final iconPos = iconBox.localToGlobal(Offset.zero);
    final iconSize = iconBox.size;
    final screenSize = MediaQuery.of(context).size;

    const double popupWidth = 360;
    const double popupHeight = 460;

    // คำนวณตำแหน่ง X — ให้ขอบขวาของ popup ชิดกับขอบขวาของ icon
    double left = iconPos.dx + iconSize.width - popupWidth;
    // ถ้าเลยขอบซ้ายจอ ให้ขยับมาทางขวา
    if (left < 8) left = 8;
    // ถ้าเลยขอบขวาจอ ให้ขยับมาทางซ้าย
    if (left + popupWidth > screenSize.width - 8) {
      left = screenSize.width - popupWidth - 8;
    }

    // คำนวณตำแหน่ง Y — แสดงด้านล่างหรือด้านบนของ icon
    final double iconBottom = iconPos.dy + iconSize.height;
    final bool showAbove = (iconBottom + popupHeight + 8) > screenSize.height;
    double top;
    if (showAbove) {
      top = iconPos.dy - popupHeight - 8;
      if (top < 8) top = 8;
    } else {
      top = iconBottom + 8;
    }

    final overlay = Overlay.of(context);

    _overlayEntry = OverlayEntry(
      builder: (overlayContext) {
        return Stack(
          children: [
            // Barrier — กดที่ไหนก็ได้เพื่อปิด popup
            Positioned.fill(
              child: GestureDetector(
                onTap: _removeOverlay,
                behavior: HitTestBehavior.opaque,
                child: Container(color: Colors.black.withValues(alpha: 0.3)),
              ),
            ),
            // Popup วางตำแหน่งใกล้ icon ปฏิทิน
            Positioned(
              left: left,
              top: top,
              child: Material(
                color: Colors.transparent,
                child: CustomDatePickerDialog(
                  initialDate: initialDate,
                  firstDate: firstDate,
                  lastDate: lastDate,
                  useBuddhistCalendar: _useBuddhistCalendar,
                  languageCode: _languageCode,
                  onDateSelected: (DateTime? pickedDate) {
                    _removeOverlay();
                    if (pickedDate != null && pickedDate != _selectedDate) {
                      setState(() {
                        _selectedDate = pickedDate;
                        isDateValid = true;
                        _updateDisplayDate();
                      });
                      if (widget.onDateSelected != null) {
                        widget.onDateSelected!(_selectedDate);
                      }
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
        FilteringTextInputFormatter.allow(
          RegExp(r'[0-9/.\-\(\),\s\u0E00-\u0E7F]'),
        ),
        LengthLimitingTextInputFormatter(40),
      ],
      onChanged: _handleInput,
      onSubmitted: (_) => _completeDate(_controller.text),
      style: const TextStyle(fontSize: 14, color: Colors.black87),
      decoration: InputDecoration(
        labelText: widget.labelText,
        floatingLabelBehavior: FloatingLabelBehavior.always,
        floatingLabelStyle: const TextStyle(
          backgroundColor: Colors.white,
          color: Colors.black87,
          fontWeight: FontWeight.bold,
          fontSize: 14,
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 12,
          vertical: 14,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: BorderSide(color: Colors.grey.shade300),
        ),
        suffixIcon: (widget.useIconSelectDate == false)
            ? null
            : IconButton(
                key: _calendarIconKey,
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(
                  Icons.calendar_today,
                  color: Colors.grey.shade600,
                  size: 20,
                ),
                onPressed: _showCustomDatePicker,
              ),
        filled: true,
        fillColor: isDateValid ? Colors.white : Colors.red.shade50,
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: _primaryColor, width: 2.0),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.grey.shade300),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.red.shade400, width: 1.5),
        ),
      ),
    );
  }
}

// ✨ Modern Date Picker Dialog with Gradient Header
class CustomDatePickerDialog extends StatefulWidget {
  final DateTime initialDate;
  final DateTime firstDate;
  final DateTime lastDate;
  final bool? useBuddhistCalendar; // null = อ่านจาก global.getYearType()
  final String? languageCode; // null = อ่านจาก global.getBranchLanguage()
  final ValueChanged<DateTime?>? onDateSelected; // callback เมื่อเลือกวันที่
  final VoidCallback? onCancel; // callback เมื่อกดยกเลิก

  const CustomDatePickerDialog({
    super.key,
    required this.initialDate,
    required this.firstDate,
    required this.lastDate,
    this.useBuddhistCalendar, // null = อ่านจาก global.getYearType() อัตโนมัติ
    this.languageCode, // null = อ่านจาก global.getBranchLanguage() อัตโนมัติ
    this.onDateSelected,
    this.onCancel,
  });

  @override
  _CustomDatePickerDialogState createState() => _CustomDatePickerDialogState();
}

class _CustomDatePickerDialogState extends State<CustomDatePickerDialog>
    with SingleTickerProviderStateMixin {
  late DateTime _currentDate;
  late DateTime _displayedMonth;
  int _displayedYear = 0;
  bool _isYearMode = false;
  bool _isMonthMode = false;

  // 🎨 Modern Color Palette
  static const Color _primaryColor = Color(0xFF667eea);
  static const Color _primaryDarkColor = Color(0xFF5a67d8);
  static const Color _accentColor = Color(0xFF764ba2);
  static const Color _surfaceColor = Color(0xFFF8FAFC);

  // ดึงประเภทปี (default จาก global.getYearType())
  bool get _useBuddhistCalendar => widget.useBuddhistCalendar ?? (global.getYearType() == 'buddhist');

  // ดึงภาษาที่ใช้ในการแสดงผล (default จาก global.getBranchLanguage())
  String get _languageCode => widget.languageCode ?? global.getBranchLanguage();

  late AnimationController _animController;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _currentDate = widget.initialDate;
    _displayedMonth = DateTime(_currentDate.year, _currentDate.month, 1);
    _displayedYear = _useBuddhistCalendar
        ? _currentDate.year + 543
        : _currentDate.year;

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
    super.dispose();
  }

  Widget _buildHeader() {
    int yearToDisplay = _useBuddhistCalendar
        ? _displayedMonth.year + 543
        : _displayedMonth.year;
    String monthName = _languageCode == 'th'
        ? global.getThaiMonthNameFull(_displayedMonth.month)
        : global.getEnglishMonthNameFull(_displayedMonth.month);
    String yearLabel = _languageCode == 'th' ? global.language('buddhist_era_short') : '';

    return Container(
      padding: const EdgeInsets.symmetric(vertical: 20, horizontal: 16),
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          colors: [_primaryColor, _accentColor],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      child: Column(
        children: [
          // Selected Date Display
          Text(
            _languageCode == 'th'
                ? global.getThaiDayNameFull(_currentDate.weekday)
                : global.getEnglishDayNameFull(_currentDate.weekday),
            style: TextStyle(
              fontSize: 14,
              color: Colors.white.withValues(alpha: 0.8),
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            '${_currentDate.day} ${_languageCode == 'th' ? global.getThaiMonthNameFull(_currentDate.month) : global.getEnglishMonthNameFull(_currentDate.month)} ${_useBuddhistCalendar ? _currentDate.year + 543 : _currentDate.year}',
            style: const TextStyle(
              fontSize: 22,
              color: Colors.white,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 16),
          // Month/Year Navigation
          Row(
            children: [
              if (!_isYearMode && !_isMonthMode)
                _buildNavButton(
                  icon: Icons.chevron_left_rounded,
                  onPressed: () {
                    setState(() {
                      _displayedMonth = DateTime(
                        _displayedMonth.year,
                        _displayedMonth.month - 1,
                        1,
                      );
                    });
                  },
                ),
              Expanded(
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    _buildHeaderButton(
                      text: monthName,
                      isActive: _isMonthMode,
                      onTap: () {
                        setState(() {
                          _isMonthMode = !_isMonthMode;
                          _isYearMode = false;
                        });
                      },
                    ),
                    const SizedBox(width: 8),
                    _buildHeaderButton(
                      text: '$yearLabel$yearToDisplay',
                      isActive: _isYearMode,
                      onTap: () {
                        setState(() {
                          _isYearMode = !_isYearMode;
                          _isMonthMode = false;
                        });
                      },
                    ),
                  ],
                ),
              ),
              if (!_isYearMode && !_isMonthMode)
                _buildNavButton(
                  icon: Icons.chevron_right_rounded,
                  onPressed: () {
                    setState(() {
                      _displayedMonth = DateTime(
                        _displayedMonth.year,
                        _displayedMonth.month + 1,
                        1,
                      );
                    });
                  },
                ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildNavButton({
    required IconData icon,
    required VoidCallback onPressed,
  }) {
    return Material(
      color: Colors.white.withValues(alpha: 0.2),
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.all(8),
          child: Icon(icon, color: Colors.white, size: 24),
        ),
      ),
    );
  }

  Widget _buildHeaderButton({
    required String text,
    required bool isActive,
    required VoidCallback onTap,
  }) {
    return Material(
      color: isActive
          ? Colors.white.withValues(alpha: 0.3)
          : Colors.transparent,
      borderRadius: BorderRadius.circular(20),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(20),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Text(
            text,
            style: const TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: Colors.white,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildWeekdayNames() {
    final List<String> weekDays = _languageCode == 'th'
        ? ['อา', 'จ', 'อ', 'พ', 'พฤ', 'ศ', 'ส']
        : ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];

    return Container(
      padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: weekDays.map((day) {
          bool isWeekend =
              day == 'อา' || day == 'ส' || day == 'Su' || day == 'Sa';
          return Expanded(
            child: Text(
              day,
              textAlign: TextAlign.center,
              style: TextStyle(
                fontWeight: FontWeight.w600,
                fontSize: 13,
                color: isWeekend ? _accentColor : Colors.grey.shade600,
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  Widget _buildCalendarGrid() {
    final DateTime firstDayOfMonth = _displayedMonth;
    final DateTime lastDayOfMonth = DateTime(
      firstDayOfMonth.year,
      firstDayOfMonth.month + 1,
      0,
    );
    final int firstWeekday = firstDayOfMonth.weekday % 7;
    final int daysInMonth = lastDayOfMonth.day;

    // Calculate rows needed
    final int totalDays = firstWeekday + daysInMonth;
    final int rows = (totalDays / 7).ceil();
    final int totalCells = rows * 7;

    List<Widget> dayWidgets = [];

    // Empty cells before first day
    for (int i = 0; i < firstWeekday; i++) {
      dayWidgets.add(const SizedBox());
    }

    // Day cells
    for (int day = 1; day <= daysInMonth; day++) {
      final DateTime currentDate = DateTime(
        _displayedMonth.year,
        _displayedMonth.month,
        day,
      );
      bool isSelected =
          currentDate.year == _currentDate.year &&
          currentDate.month == _currentDate.month &&
          currentDate.day == _currentDate.day;
      bool isToday =
          currentDate.year == DateTime.now().year &&
          currentDate.month == DateTime.now().month &&
          currentDate.day == DateTime.now().day;
      int weekday = (firstWeekday + day - 1) % 7;
      bool isWeekend = weekday == 0 || weekday == 6;

      dayWidgets.add(
        GestureDetector(
          onTap: () {
            setState(() {
              _currentDate = currentDate;
            });
            if (widget.onDateSelected != null) {
              widget.onDateSelected!(_currentDate);
            } else {
              Navigator.of(context).pop(_currentDate);
            }
          },
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            margin: const EdgeInsets.all(2),
            decoration: BoxDecoration(
              gradient: isSelected
                  ? const LinearGradient(
                      colors: [_primaryColor, _accentColor],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    )
                  : null,
              color: isSelected
                  ? null
                  : (isToday ? _primaryColor.withValues(alpha: 0.1) : null),
              borderRadius: BorderRadius.circular(10),
              border: isToday && !isSelected
                  ? Border.all(color: _primaryColor, width: 2)
                  : null,
              boxShadow: isSelected
                  ? [
                      BoxShadow(
                        color: _primaryColor.withValues(alpha: 0.4),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ]
                  : null,
            ),
            alignment: Alignment.center,
            child: Text(
              '$day',
              style: TextStyle(
                color: isSelected
                    ? Colors.white
                    : (isToday
                          ? _primaryDarkColor
                          : (isWeekend ? _accentColor : Colors.grey.shade700)),
                fontWeight: isToday || isSelected
                    ? FontWeight.bold
                    : FontWeight.w500,
                fontSize: 14,
              ),
            ),
          ),
        ),
      );
    }

    // Empty cells after last day
    while (dayWidgets.length < totalCells) {
      dayWidgets.add(const SizedBox());
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: GridView.count(
        crossAxisCount: 7,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        childAspectRatio: 1.1,
        children: dayWidgets,
      ),
    );
  }

  Widget _buildMonthGrid() {
    List<Widget> monthWidgets = [];
    for (int month = 1; month <= 12; month++) {
      String monthName = _languageCode == 'th'
          ? global.getThaiMonthNameShort(month)
          : global.getEnglishMonthNameShort(month);
      bool isSelected = month == _displayedMonth.month;
      bool isCurrent =
          month == DateTime.now().month &&
          _displayedMonth.year == DateTime.now().year;

      monthWidgets.add(
        GestureDetector(
          onTap: () {
            setState(() {
              _displayedMonth = DateTime(_displayedMonth.year, month, 1);
              _isMonthMode = false;
            });
          },
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            margin: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              gradient: isSelected
                  ? const LinearGradient(
                      colors: [_primaryColor, _accentColor],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    )
                  : null,
              color: isSelected
                  ? null
                  : (isCurrent
                        ? _primaryColor.withValues(alpha: 0.1)
                        : _surfaceColor),
              borderRadius: BorderRadius.circular(12),
              border: isCurrent && !isSelected
                  ? Border.all(color: _primaryColor, width: 2)
                  : null,
              boxShadow: isSelected
                  ? [
                      BoxShadow(
                        color: _primaryColor.withValues(alpha: 0.3),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ]
                  : null,
            ),
            alignment: Alignment.center,
            child: Text(
              monthName,
              style: TextStyle(
                color: isSelected ? Colors.white : Colors.grey.shade700,
                fontWeight: isSelected || isCurrent
                    ? FontWeight.bold
                    : FontWeight.w500,
                fontSize: 14,
              ),
            ),
          ),
        ),
      );
    }

    return Container(
      padding: const EdgeInsets.all(12),
      child: GridView.count(
        crossAxisCount: 3,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        childAspectRatio: 2.0,
        children: monthWidgets,
      ),
    );
  }

  Widget _buildYearGrid() {
    int middleYear = _displayedYear;
    int startYear = middleYear - 5;
    List<Widget> yearWidgets = [];

    for (int year = startYear; year < startYear + 12; year++) {
      bool isSelected = year == _displayedYear;
      int currentYear = _useBuddhistCalendar
          ? DateTime.now().year + 543
          : DateTime.now().year;
      bool isCurrent = year == currentYear;

      yearWidgets.add(
        GestureDetector(
          onTap: () {
            setState(() {
              _displayedYear = year;
              int christianYear = _useBuddhistCalendar
                  ? year - 543
                  : year;
              _displayedMonth = DateTime(
                christianYear,
                _displayedMonth.month,
                1,
              );
              _isYearMode = false;
            });
          },
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            margin: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              gradient: isSelected
                  ? const LinearGradient(
                      colors: [_primaryColor, _accentColor],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    )
                  : null,
              color: isSelected
                  ? null
                  : (isCurrent
                        ? _primaryColor.withValues(alpha: 0.1)
                        : _surfaceColor),
              borderRadius: BorderRadius.circular(12),
              border: isCurrent && !isSelected
                  ? Border.all(color: _primaryColor, width: 2)
                  : null,
              boxShadow: isSelected
                  ? [
                      BoxShadow(
                        color: _primaryColor.withValues(alpha: 0.3),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ]
                  : null,
            ),
            alignment: Alignment.center,
            child: Text(
              '$year',
              style: TextStyle(
                color: isSelected ? Colors.white : Colors.grey.shade700,
                fontWeight: isSelected || isCurrent
                    ? FontWeight.bold
                    : FontWeight.w500,
                fontSize: 14,
              ),
            ),
          ),
        ),
      );
    }

    return Container(
      padding: const EdgeInsets.all(12),
      child: GridView.count(
        crossAxisCount: 3,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        childAspectRatio: 2.0,
        children: yearWidgets,
      ),
    );
  }

  Widget _buildNavigationButtons() {
    if (_isYearMode) {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            _buildYearNavButton(
              icon: Icons.keyboard_double_arrow_left_rounded,
              onPressed: () {
                setState(() {
                  _displayedYear -= 12;
                });
              },
            ),
            Text(
              '${_displayedYear - 5} - ${_displayedYear + 6}',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: Colors.grey.shade600,
              ),
            ),
            _buildYearNavButton(
              icon: Icons.keyboard_double_arrow_right_rounded,
              onPressed: () {
                setState(() {
                  _displayedYear += 12;
                });
              },
            ),
          ],
        ),
      );
    }
    return const SizedBox.shrink();
  }

  Widget _buildYearNavButton({
    required IconData icon,
    required VoidCallback onPressed,
  }) {
    return Material(
      color: _surfaceColor,
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.all(8),
          child: Icon(icon, color: _primaryColor, size: 24),
        ),
      ),
    );
  }

  Widget _buildActionButtons() {
    return Container(
      padding: EdgeInsets.all(16),
      child: Row(
        children: [
          Expanded(
            child: TextButton(
              onPressed: () {
                if (widget.onCancel != null) {
                  widget.onCancel!();
                } else {
                  Navigator.of(context).pop();
                }
              },
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: BorderSide(color: Colors.grey.shade300),
                ),
              ),
              child: Text(
                _languageCode == 'th' ? global.language('cancel') : 'Cancel',
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Colors.grey.shade600,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                gradient: const LinearGradient(
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
              child: TextButton(
                onPressed: () {
                  // Go to today
                  setState(() {
                    _currentDate = DateTime.now();
                    _displayedMonth = DateTime(
                      _currentDate.year,
                      _currentDate.month,
                      1,
                    );
                    _displayedYear = _useBuddhistCalendar
                        ? _currentDate.year + 543
                        : _currentDate.year;
                  });
                  if (widget.onDateSelected != null) {
                    widget.onDateSelected!(_currentDate);
                  } else {
                    Navigator.of(context).pop(_currentDate);
                  }
                },
                style: TextButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                child: Text(
                  _languageCode == 'th' ? global.language('alert_today') : 'Today',
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Colors.white,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildContent() {
    if (_isYearMode) {
      return _buildYearGrid();
    } else if (_isMonthMode) {
      return _buildMonthGrid();
    } else {
      return Column(children: [_buildWeekdayNames(), _buildCalendarGrid()]);
    }
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
              _buildNavigationButtons(),
              _buildContent(),
              _buildActionButtons(),
            ],
          ),
        ),
      ),
    );
  }
}
