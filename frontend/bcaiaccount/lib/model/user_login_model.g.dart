// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user_login_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

UserLoginModel _$UserLoginModelFromJson(Map<String, dynamic> json) =>
    UserLoginModel(
      email: json['email'] as String,
      code: json['code'] as String,
      name: json['name'] as String,
      token: json['token'] as String,
      refreshtoken: json['refreshtoken'] as String,
      photourl: json['photourl'] as String,
    );

Map<String, dynamic> _$UserLoginModelToJson(UserLoginModel instance) =>
    <String, dynamic>{
      'email': instance.email,
      'code': instance.code,
      'name': instance.name,
      'token': instance.token,
      'refreshtoken': instance.refreshtoken,
      'photourl': instance.photourl,
    };
