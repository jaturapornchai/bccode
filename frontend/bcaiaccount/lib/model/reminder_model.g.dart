// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'reminder_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReminderModel _$ReminderModelFromJson(
  Map<String, dynamic> json,
) => ReminderModel(
  reminderId: json['reminder_id'] as String,
  shopId: json['shopid'] as String,
  name: json['name'] as String,
  emails: (json['emails'] as List<dynamic>).map((e) => e as String).toList(),
  lineGroupTokens: (json['line_group_tokens'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
  dailyReminder: json['daily_reminder'] as bool,
  dailyReminderTimes: (json['daily_reminder_times'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
  specificDayReminders: (json['specific_day_reminders'] as Map<String, dynamic>)
      .map(
        (k, e) =>
            MapEntry(k, (e as List<dynamic>).map((e) => e as String).toList()),
      ),
  monthlyReminders: (json['monthly_reminders'] as Map<String, dynamic>).map(
    (k, e) =>
        MapEntry(k, (e as List<dynamic>).map((e) => e as String).toList()),
  ),
  reorderReminder: json['reorder_reminder'] as bool,
  reminderTypes: (json['reminder_types'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
);

Map<String, dynamic> _$ReminderModelToJson(ReminderModel instance) =>
    <String, dynamic>{
      'reminder_id': instance.reminderId,
      'shopid': instance.shopId,
      'name': instance.name,
      'emails': instance.emails,
      'line_group_tokens': instance.lineGroupTokens,
      'daily_reminder': instance.dailyReminder,
      'daily_reminder_times': instance.dailyReminderTimes,
      'specific_day_reminders': instance.specificDayReminders,
      'monthly_reminders': instance.monthlyReminders,
      'reorder_reminder': instance.reorderReminder,
      'reminder_types': instance.reminderTypes,
    };
