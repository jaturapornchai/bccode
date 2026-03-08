// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'profile_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProfileModel _$ProfileModelFromJson(Map<String, dynamic> json) => ProfileModel(
  username: json['username'] as String?,
  name: json['name'] as String?,
  avatar: json['avatar'] as String?,
  timezonelabel: json['timezonelabel'] as String?,
  timezoneoffset: json['timezoneoffset'] as String?,
  yeartype: json['yeartype'] as String?,
  lineUserId: json['line_user_id'] as String?,
  lineDisplayName: json['line_display_name'] as String?,
  linePictureUrl: json['line_picture_url'] as String?,
);

Map<String, dynamic> _$ProfileModelToJson(ProfileModel instance) =>
    <String, dynamic>{
      'username': instance.username,
      'name': instance.name,
      'avatar': instance.avatar,
      'timezonelabel': instance.timezonelabel,
      'timezoneoffset': instance.timezoneoffset,
      'yeartype': instance.yeartype,
      'line_user_id': instance.lineUserId,
      'line_display_name': instance.lineDisplayName,
      'line_picture_url': instance.linePictureUrl,
    };
