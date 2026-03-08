part of 'user_bloc.dart';

abstract class UserState extends Equatable {
  const UserState();

  @override
  List<Object> get props => [];
}

class UserInitial extends UserState {}

class UserInProgress extends UserState {}

class UserLoadSuccess extends UserState {
  final List<UserModel> users;

  const UserLoadSuccess({required this.users});

  UserLoadSuccess copyWith({
    List<UserModel>? users,
  }) =>
      UserLoadSuccess(users: users ?? this.users);

  @override
  List<Object> get props => [users];
}

class UserLoadFailed extends UserState {
  final String message;

  const UserLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UserDeleteInProgress extends UserState {}

class UserDeleteSuccess extends UserState {}

class UserDeleteFailed extends UserState {
  final String message;

  const UserDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UserGetInProgress extends UserState {}

class UserGetSuccess extends UserState {
  final UserModel user;

  const UserGetSuccess({required this.user});

  UserGetSuccess copyWith({
    UserModel? user,
  }) =>
      UserGetSuccess(user: user ?? this.user);

  @override
  List<Object> get props => [user];
}

class UserGetFailed extends UserState {
  final String message;

  const UserGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UserSaveAndUpdateInitial extends UserState {}

class UserSaveAndUpdateInProgress extends UserState {}

class UserSaveAndUpdateSuccess extends UserState {}

class UserSaveAndUpdateFailed extends UserState {
  final String message;

  const UserSaveAndUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
