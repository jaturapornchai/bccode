/// ฟังก์ชันกลางสำหรับใช้ทั้ง Project
/// - จัดรูปแบบวันที่ภาษาไทย (พ.ศ.)
/// - Helper functions ต่างๆ
library;

/// รูปแบบวันที่
enum DateTimeFormatEnum {
  /// วันจันทร์ ที่ 22 กุมภาพันธ์ 2569
  fullDate,

  /// 22/02/2569
  date,

  /// 14:30
  time,

  /// 22/02/2569 14:30
  dateTime,

  /// 22/02/2569 14:30 (วันจันทร์)
  dateTimeDay,

  /// 22/02/2569 (วันจันทร์)
  dateDay,
}

/// ชื่อวันภาษาไทย
const _thaiDayNames = [
  'วันจันทร์',
  'วันอังคาร',
  'วันพุธ',
  'วันพฤหัสบดี',
  'วันศุกร์',
  'วันเสาร์',
  'วันอาทิตย์',
];

/// ชื่อวันภาษาไทยแบบสั้น
const _thaiDayShortNames = [
  'จ.',
  'อ.',
  'พ.',
  'พฤ.',
  'ศ.',
  'ส.',
  'อา.',
];

/// ชื่อเดือนภาษาไทย
const _thaiMonthNames = [
  'มกราคม',
  'กุมภาพันธ์',
  'มีนาคม',
  'เมษายน',
  'พฤษภาคม',
  'มิถุนายน',
  'กรกฎาคม',
  'สิงหาคม',
  'กันยายน',
  'ตุลาคม',
  'พฤศจิกายน',
  'ธันวาคม',
];

/// ชื่อเดือนภาษาไทยแบบสั้น
const _thaiMonthShortNames = [
  'ม.ค.',
  'ก.พ.',
  'มี.ค.',
  'เม.ย.',
  'พ.ค.',
  'มิ.ย.',
  'ก.ค.',
  'ส.ค.',
  'ก.ย.',
  'ต.ค.',
  'พ.ย.',
  'ธ.ค.',
];

/// แปลงวันที่เป็นข้อความภาษาไทย พ.ศ.
///
/// ```dart
/// dateTimeBuddhist(DateTime.now()); // "วันเสาร์ ที่ 22 กุมภาพันธ์ 2569"
/// dateTimeBuddhist(DateTime.now(), format: DateTimeFormatEnum.date); // "22/02/2569"
/// ```
String dateTimeBuddhist(
  DateTime dateTime, {
  DateTimeFormatEnum format = DateTimeFormatEnum.fullDate,
}) {
  final day = dateTime.day;
  final month = dateTime.month;
  final buddhistYear = dateTime.year + 543;
  final dayOfWeek = _thaiDayNames[dateTime.weekday - 1]; // weekday: 1=จันทร์
  final monthName = _thaiMonthNames[month - 1];
  final yearStr = buddhistYear.toString();

  final dd = day.toString().padLeft(2, '0');
  final mm = month.toString().padLeft(2, '0');
  final hh = dateTime.hour.toString().padLeft(2, '0');
  final min = dateTime.minute.toString().padLeft(2, '0');

  switch (format) {
    case DateTimeFormatEnum.fullDate:
      return '$dayOfWeek ที่ $day $monthName $yearStr';
    case DateTimeFormatEnum.date:
      return '$dd/$mm/$yearStr';
    case DateTimeFormatEnum.time:
      return '$hh:$min';
    case DateTimeFormatEnum.dateTime:
      if (dateTime.hour == 0 && dateTime.minute == 0) {
        return '$dd/$mm/$yearStr';
      }
      return '$dd/$mm/$yearStr $hh:$min';
    case DateTimeFormatEnum.dateTimeDay:
      if (dateTime.hour == 0 && dateTime.minute == 0) {
        return '$dd/$mm/$yearStr ($dayOfWeek)';
      }
      return '$dd/$mm/$yearStr $hh:$min ($dayOfWeek)';
    case DateTimeFormatEnum.dateDay:
      return '$dd/$mm/$yearStr ($dayOfWeek)';
  }
}

/// แปลงวันที่เป็นข้อความแบบสั้น เช่น "22 ก.พ. 2569"
String dateShortBuddhist(DateTime dateTime) {
  final buddhistYear = dateTime.year + 543;
  final monthShort = _thaiMonthShortNames[dateTime.month - 1];
  return '${dateTime.day} $monthShort $buddhistYear';
}

/// ชื่อวันภาษาไทย (จันทร์-อาทิตย์)
String thaiDayName(DateTime dateTime) {
  return _thaiDayNames[dateTime.weekday - 1];
}

/// ชื่อวันภาษาไทยแบบสั้น (จ.-อา.)
String thaiDayShortName(DateTime dateTime) {
  return _thaiDayShortNames[dateTime.weekday - 1];
}

/// ชื่อเดือนภาษาไทย
String thaiMonthName(int month) {
  return _thaiMonthNames[month - 1];
}

/// ชื่อเดือนภาษาไทยแบบสั้น
String thaiMonthShortName(int month) {
  return _thaiMonthShortNames[month - 1];
}

/// แปลงปี ค.ศ. เป็น พ.ศ.
int toBuddhistYear(int year) {
  return year + 543;
}
