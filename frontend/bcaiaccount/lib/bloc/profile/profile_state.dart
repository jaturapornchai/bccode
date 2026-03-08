part of 'profile_bloc.dart';

abstract class ProfileState extends Equatable {
  const ProfileState();

  @override
  List<Object> get props => [];
}

class ProfileInitial extends ProfileState {}

/// get profile
class GetProfileInProgress extends ProfileState {}

class GetProfileSuccess extends ProfileState {
  final ProfileModel profile;

  const GetProfileSuccess({
    required this.profile,
  });

  @override
  List<Object> get props => [profile];
}

class GetProfileFailed extends ProfileState {
  final String message;

  const GetProfileFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

/// update profile
class UpdateProfileInProgress extends ProfileState {}

class UpdateProfileSuccess extends ProfileState {}

class UpdateProfileFailed extends ProfileState {
  final String message;

  const UpdateProfileFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
