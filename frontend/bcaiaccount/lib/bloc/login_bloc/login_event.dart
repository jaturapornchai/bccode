part of 'login_bloc.dart';

abstract class LoginEvent extends Equatable {
  const LoginEvent();

  @override
  List<Object> get props => [];
}

// ignore: must_be_immutable
class LoginOnLoad extends LoginEvent {
  String userName;
  String passWord;
  LoginOnLoad({required this.userName, required this.passWord});

  @override
  List<Object> get props => [];
}

// ignore: must_be_immutable
class RegisterUser extends LoginEvent {
  String userName;
  String passWord;
  String timezonelabel;
  String timezoneoffset;
  String yeartype;
  RegisterUser({
    required this.userName,
    required this.passWord,
    required this.timezonelabel,
    required this.timezoneoffset,
    required this.yeartype,
  });

  @override
  List<Object> get props => [];
}

// ignore: must_be_immutable
class TokenLogin extends LoginEvent {
  String token;
  TokenLogin({required this.token});

  @override
  List<Object> get props => [];
}

class CreateShop extends LoginEvent {
  final CreateShopModel createShop;

  const CreateShop({
    required this.createShop,
  });

  @override
  List<Object> get props => [createShop];
}

/// Logout
class Logout extends LoginEvent {
  const Logout();

  @override
  List<Object> get props => [];
}

/// Token Invalid
class TokenInvalid extends LoginEvent {
  const TokenInvalid();

  @override
  List<Object> get props => [];
}

/// LINE Login Event
/// ใช้ข้อมูล LINE user ในการเข้าสู่ระบบผ่าน Main API
class LineLogin extends LoginEvent {
  final String lineUserId;
  final String? accessToken; // JWT token (optional - ไม่จำเป็นสำหรับ QR code flow)
  final String? displayName;
  final String? pictureUrl;
  final String? email;

  const LineLogin({
    required this.lineUserId,
    this.accessToken,
    this.displayName,
    this.pictureUrl,
    this.email,
  });

  @override
  List<Object> get props => [lineUserId];
}

/// Google Login Event
/// ใช้ข้อมูล Google user ในการเข้าสู่ระบบผ่าน Main API
class GoogleLogin extends LoginEvent {
  final String googleUserId;
  final String accessToken; // JWT token จาก Google Backend
  final String? displayName;
  final String? pictureUrl;
  final String? email;

  const GoogleLogin({
    required this.googleUserId,
    required this.accessToken,
    this.displayName,
    this.pictureUrl,
    this.email,
  });

  @override
  List<Object> get props => [googleUserId, accessToken];
}
