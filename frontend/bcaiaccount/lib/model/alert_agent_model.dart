// Model สำหรับ BC Alert Agent
// เก็บข้อมูลการตั้งค่าการแจ้งเตือนอัตโนมัติ
import '../global.dart' as global;

class AlertAgentModel {
  final String id;
  final String shopId;
  final String taskName;
  final bool isActive;
  final ReportConfigModel reportConfig;
  final ScheduleModel schedule;
  final ChannelsModel channels;
  final MetadataModel metadata;

  AlertAgentModel({
    required this.id,
    required this.shopId,
    required this.taskName,
    required this.isActive,
    required this.reportConfig,
    required this.schedule,
    required this.channels,
    required this.metadata,
  });

  factory AlertAgentModel.fromJson(Map<String, dynamic> json) {
    return AlertAgentModel(
      id: json['_id'] is Map ? json['_id']['\$oid'] ?? '' : json['_id'] ?? '',
      shopId: json['shopid'] ?? '',
      taskName: json['task_name'] ?? '',
      isActive: json['is_active'] ?? false,
      reportConfig: ReportConfigModel.fromJson(json['report_config'] ?? {}),
      schedule: ScheduleModel.fromJson(json['schedule'] ?? {}),
      channels: ChannelsModel.fromJson(json['channels'] ?? {}),
      metadata: MetadataModel.fromJson(json['metadata'] ?? {}),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'shopid': shopId,
      'task_name': taskName,
      'is_active': isActive,
      'report_config': reportConfig.toJson(),
      'schedule': schedule.toJson(),
      'channels': channels.toJson(),
      'metadata': metadata.toJson(),
    };
  }

  AlertAgentModel copyWith({String? id, String? shopId, String? taskName, bool? isActive, ReportConfigModel? reportConfig, ScheduleModel? schedule, ChannelsModel? channels, MetadataModel? metadata}) {
    return AlertAgentModel(
      id: id ?? this.id,
      shopId: shopId ?? this.shopId,
      taskName: taskName ?? this.taskName,
      isActive: isActive ?? this.isActive,
      reportConfig: reportConfig ?? this.reportConfig,
      schedule: schedule ?? this.schedule,
      channels: channels ?? this.channels,
      metadata: metadata ?? this.metadata,
    );
  }
}

/// ประเภทรายงาน
class ReportConfigModel {
  final String type; // sales, stock, debtor, creditor, etc.
  final String period; // today, yesterday, this_week, this_month, etc.

  ReportConfigModel({required this.type, required this.period});

  factory ReportConfigModel.fromJson(Map<String, dynamic> json) {
    return ReportConfigModel(type: json['type'] ?? 'sales', period: json['period'] ?? 'today');
  }

  Map<String, dynamic> toJson() {
    return {'type': type, 'period': period};
  }

  ReportConfigModel copyWith({String? type, String? period}) {
    return ReportConfigModel(type: type ?? this.type, period: period ?? this.period);
  }
}

/// ตารางเวลาการทำงาน
/// scheduleMode:
/// - 'interval' = ทำงานทุก X นาที
/// - 'weekly' = กำหนดวันในสัปดาห์+เวลา (อา.-ส.)
/// - 'monthly' = กำหนดวันที่ในเดือน+เวลา (1-31, วันสิ้นเดือน)
/// หมายเหตุ: สามารถป้อนข้อมูลไว้ได้ทุกโหมด แต่ระบบจะทำงานตาม scheduleMode ที่เลือกเท่านั้น
class ScheduleModel {
  final String scheduleMode; // 'interval', 'weekly', 'monthly'

  // === Mode 1: Interval ===
  // ทำงานทุก X นาที (1, 5, 15, 30, 60)
  final int intervalMinutes;

  // === Mode 2: Weekly ===
  // กำหนดวันในสัปดาห์ + เวลา
  // key: 0=อาทิตย์, 1=จันทร์, 2=อังคาร, 3=พุธ, 4=พฤหัส, 5=ศุกร์, 6=เสาร์
  // value: list of times เช่น ["08:00", "12:00", "18:00"]
  final Map<int, List<String>> weeklySchedule;

  // === Mode 3: Monthly ===
  // กำหนดวันที่ในเดือน + เวลา
  // key: 1-31 = วันที่, -1 = วันสิ้นเดือน (28/29/30/31 แล้วแต่เดือน)
  // value: list of times เช่น ["08:00", "20:00"]
  final Map<int, List<String>> monthlySchedule;

  ScheduleModel({this.scheduleMode = 'interval', this.intervalMinutes = 60, Map<int, List<String>>? weeklySchedule, Map<int, List<String>>? monthlySchedule})
    : weeklySchedule = weeklySchedule ?? {},
      monthlySchedule = monthlySchedule ?? {};

  factory ScheduleModel.fromJson(Map<String, dynamic> json) {
    // แปลง weeklySchedule จาก JSON (key เป็น String)
    Map<int, List<String>> weekly = {};
    if (json['weekly_schedule'] != null) {
      final Map<String, dynamic> weeklyJson = json['weekly_schedule'];
      weeklyJson.forEach((key, value) {
        weekly[int.parse(key)] = List<String>.from(value ?? []);
      });
    }

    // แปลง monthlySchedule จาก JSON (key เป็น String)
    Map<int, List<String>> monthly = {};
    if (json['monthly_schedule'] != null) {
      final Map<String, dynamic> monthlyJson = json['monthly_schedule'];
      monthlyJson.forEach((key, value) {
        monthly[int.parse(key)] = List<String>.from(value ?? []);
      });
    }

    return ScheduleModel(scheduleMode: json['schedule_mode'] ?? 'interval', intervalMinutes: json['interval_minutes'] ?? 60, weeklySchedule: weekly, monthlySchedule: monthly);
  }

  Map<String, dynamic> toJson() {
    // แปลง weeklySchedule เป็น JSON (key เป็น String)
    Map<String, dynamic> weeklyJson = {};
    weeklySchedule.forEach((key, value) {
      weeklyJson[key.toString()] = value;
    });

    // แปลง monthlySchedule เป็น JSON (key เป็น String)
    Map<String, dynamic> monthlyJson = {};
    monthlySchedule.forEach((key, value) {
      monthlyJson[key.toString()] = value;
    });

    return {'schedule_mode': scheduleMode, 'interval_minutes': intervalMinutes, 'weekly_schedule': weeklyJson, 'monthly_schedule': monthlyJson};
  }

  ScheduleModel copyWith({String? scheduleMode, int? intervalMinutes, Map<int, List<String>>? weeklySchedule, Map<int, List<String>>? monthlySchedule}) {
    return ScheduleModel(
      scheduleMode: scheduleMode ?? this.scheduleMode,
      intervalMinutes: intervalMinutes ?? this.intervalMinutes,
      weeklySchedule: weeklySchedule ?? this.weeklySchedule,
      monthlySchedule: monthlySchedule ?? this.monthlySchedule,
    );
  }

  /// ชื่อวันในสัปดาห์ (ภาษาไทย)
  static const List<String> dayNames = ['อาทิตย์', 'จันทร์', 'อังคาร', 'พุธ', 'พฤหัสบดี', 'ศุกร์', 'เสาร์'];

  /// ชื่อวันย่อ
  static const List<String> dayNamesShort = ['อา.', 'จ.', 'อ.', 'พ.', 'พฤ.', 'ศ.', 'ส.'];

  /// Key พิเศษสำหรับวันสิ้นเดือน
  static const int lastDayOfMonth = -1;
}

/// ช่องทางการแจ้งเตือน
class ChannelsModel {
  final EmailChannelModel email;
  final LineChannelModel line;

  ChannelsModel({required this.email, required this.line});

  factory ChannelsModel.fromJson(Map<String, dynamic> json) {
    return ChannelsModel(email: EmailChannelModel.fromJson(json['email'] ?? {}), line: LineChannelModel.fromJson(json['line'] ?? {}));
  }

  Map<String, dynamic> toJson() {
    return {'email': email.toJson(), 'line': line.toJson()};
  }

  ChannelsModel copyWith({EmailChannelModel? email, LineChannelModel? line}) {
    return ChannelsModel(email: email ?? this.email, line: line ?? this.line);
  }
}

/// ช่องทาง Email
class EmailChannelModel {
  final bool isEnabled;
  final List<String> addresses;

  EmailChannelModel({required this.isEnabled, required this.addresses});

  factory EmailChannelModel.fromJson(Map<String, dynamic> json) {
    return EmailChannelModel(isEnabled: json['is_enabled'] ?? false, addresses: List<String>.from(json['addresses'] ?? []));
  }

  Map<String, dynamic> toJson() {
    return {'is_enabled': isEnabled, 'addresses': addresses};
  }

  EmailChannelModel copyWith({bool? isEnabled, List<String>? addresses}) {
    return EmailChannelModel(isEnabled: isEnabled ?? this.isEnabled, addresses: addresses ?? this.addresses);
  }
}

/// ช่องทาง LINE
class LineChannelModel {
  final bool isEnabled;
  final List<String> tokens;

  LineChannelModel({required this.isEnabled, required this.tokens});

  factory LineChannelModel.fromJson(Map<String, dynamic> json) {
    return LineChannelModel(isEnabled: json['is_enabled'] ?? false, tokens: List<String>.from(json['tokens'] ?? []));
  }

  Map<String, dynamic> toJson() {
    return {'is_enabled': isEnabled, 'tokens': tokens};
  }

  LineChannelModel copyWith({bool? isEnabled, List<String>? tokens}) {
    return LineChannelModel(isEnabled: isEnabled ?? this.isEnabled, tokens: tokens ?? this.tokens);
  }
}

/// Metadata
class MetadataModel {
  final String? lastRunAt;

  MetadataModel({this.lastRunAt});

  factory MetadataModel.fromJson(Map<String, dynamic> json) {
    return MetadataModel(lastRunAt: json['last_run_at']?.toString());
  }

  Map<String, dynamic> toJson() {
    return {'last_run_at': lastRunAt};
  }
}

/// ประเภทรายงานที่รองรับ
enum ReportType {
  sales('sales', 'alert_sales_summary'),
  grossProfit('gross_profit', 'alert_gross_profit_report');

  final String code;
  final String _labelKey;

  const ReportType(this.code, this._labelKey);

  String get label => global.language(_labelKey);
}

/// ช่วงเวลารายงาน
enum ReportPeriod {
  today('today', 'alert_today'),
  yesterday('yesterday', 'alert_yesterday'),
  thisWeek('this_week', 'alert_this_week'),
  lastWeek('last_week', 'alert_last_week'),
  thisMonth('this_month', 'alert_this_month'),
  lastMonth('last_month', 'alert_last_month');

  final String code;
  final String _labelKey;

  const ReportPeriod(this.code, this._labelKey);

  String get label => global.language(_labelKey);
}
