import 'package:json_annotation/json_annotation.dart';

part 'reminder_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ReminderModel {
  @JsonKey(name: 'reminder_id')
  late final String reminderId;

  @JsonKey(name: 'shopid')
  late final String shopId;

  late final String name;
  late final List<String> emails;

  @JsonKey(name: 'line_group_tokens')
  late final List<String> lineGroupTokens;

  @JsonKey(name: 'daily_reminder')
  late final bool dailyReminder;

  @JsonKey(name: 'daily_reminder_times')
  late final List<String> dailyReminderTimes;

  @JsonKey(name: 'specific_day_reminders')
  late final Map<String, List<String>> specificDayReminders;

  @JsonKey(name: 'monthly_reminders')
  late final Map<String, List<String>> monthlyReminders;

  @JsonKey(name: 'reorder_reminder')
  late final bool reorderReminder;

  @JsonKey(name: 'reminder_types')
  late final List<String> reminderTypes;

  ReminderModel({
    required this.reminderId,
    required this.shopId,
    required this.name,
    required this.emails,
    required this.lineGroupTokens,
    required this.dailyReminder,
    required this.dailyReminderTimes,
    required this.specificDayReminders,
    required this.monthlyReminders,
    required this.reorderReminder,
    required this.reminderTypes,
  });

  factory ReminderModel.fromJson(Map<String, dynamic> json) =>
      _$ReminderModelFromJson(json);

  Map<String, dynamic> toJson() => _$ReminderModelToJson(this);
}
