part of 'profile_bloc.dart';

abstract class ProfileEvent extends Equatable {
  const ProfileEvent();

  @override
  List<Object> get props => [];
}

/// get profile
class GetProfile extends ProfileEvent {
  const GetProfile();

  @override
  List<Object> get props => [];
}

/// update profile
class UpdateProfile extends ProfileEvent {
  final ProfileModel profile;

  const UpdateProfile({
    required this.profile,
  });

  @override
  List<Object> get props => [profile];
}
