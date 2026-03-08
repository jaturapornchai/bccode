import 'package:json_annotation/json_annotation.dart';

part 'user_login_model.g.dart';

@JsonSerializable()
class UserLoginModel {
  String email;
  String code;
  String name;
  String token;
  String refreshtoken;
  String photourl;

  UserLoginModel({
    required this.email,
    required this.code,
    required this.name,
    required this.token,
    required this.refreshtoken,
    required this.photourl,
  });

  factory UserLoginModel.fromJson(Map<String, dynamic> json) =>
      _$UserLoginModelFromJson(json);

  Map<String, dynamic> toJson() => _$UserLoginModelToJson(this);
}
