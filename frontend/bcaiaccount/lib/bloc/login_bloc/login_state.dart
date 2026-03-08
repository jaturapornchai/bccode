part of 'login_bloc.dart';

abstract class LoginState extends Equatable {
  const LoginState();

  @override
  List<Object> get props => [];
}

class LoginInitial extends LoginState {}

class LoginInProgress extends LoginState {}

// ignore: must_be_immutable
class LoginSuccess extends LoginState {
  UserLoginModel userLogin;

  LoginSuccess({required this.userLogin});

  @override
  List<Object> get props => [userLogin];
}

class LoginFailed extends LoginState {
  final String message;
  const LoginFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class TokenLoginInitial extends LoginState {}

class TokenLoginInProgress extends LoginState {}

// ignore: must_be_immutable
class TokenLoginSuccess extends LoginState {
  UserLoginModel userLogin;

  TokenLoginSuccess({required this.userLogin});

  @override
  List<Object> get props => [userLogin];
}

class TokenLoginFailed extends LoginState {
  final String message;
  const TokenLoginFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class CreateShopInProgress extends LoginState {}

class CreateShopSuccess extends LoginState {}

class CreateShopFailed extends LoginState {
  final String message;

  const CreateShopFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class RegisterInProgress extends LoginState {}

class RegisterSuccess extends LoginState {}

class RegisterFailed extends LoginState {
  final String message;

  const RegisterFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class LogoutInProgress extends LoginState {}

class LogoutSuccess extends LoginState {}

class LogoutFailed extends LoginState {
  final String message;

  const LogoutFailed({required this.message});

  @override
  List<Object> get props => [message];
}

/// LINE Login States
class LineLoginInProgress extends LoginState {}

// ignore: must_be_immutable
class LineLoginSuccess extends LoginState {
  UserLoginModel userLogin;

  LineLoginSuccess({required this.userLogin});

  @override
  List<Object> get props => [userLogin];
}

class LineLoginFailed extends LoginState {
  final String message;
  const LineLoginFailed({required this.message});

  @override
  List<Object> get props => [message];
}

/// Google Login States
class GoogleLoginInProgress extends LoginState {}

// ignore: must_be_immutable
class GoogleLoginSuccess extends LoginState {
  UserLoginModel userLogin;

  GoogleLoginSuccess({required this.userLogin});

  @override
  List<Object> get props => [userLogin];
}

class GoogleLoginFailed extends LoginState {
  final String message;
  const GoogleLoginFailed({required this.message});

  @override
  List<Object> get props => [message];
}
