// ignore_for_file: avoid_print

import 'dart:convert';

import 'package:smlaicloud/model/profile_model.dart';
import 'package:smlaicloud/repositories/profile_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'profile_event.dart';
part 'profile_state.dart';

class ProfileBloc extends Bloc<ProfileEvent, ProfileState> {
  final ProfileRepository _profileRepository;

  ProfileBloc({required ProfileRepository profileRepository})
      : _profileRepository = profileRepository,
        super(ProfileInitial()) {
    on<GetProfile>(_onGetProfile);
    on<UpdateProfile>(_onUpdateProfile);
  }

  /// get profile
  void _onGetProfile(GetProfile event, Emitter<ProfileState> emit) async {
    emit(GetProfileInProgress());
    try {
      final result = await _profileRepository.getProfile();
      if (result.success) {
        ProfileModel profile = ProfileModel.fromJson(result.data);
        emit(GetProfileSuccess(profile: profile));
      } else {
        emit(const GetProfileFailed(message: 'Not Found'));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(GetProfileFailed(message: error['message']));
    }
  }

  /// update profile
  void _onUpdateProfile(UpdateProfile event, Emitter<ProfileState> emit) async {
    emit(UpdateProfileInProgress());
    try {
      final result = await _profileRepository.updateProfile(event.profile);
      if (result.success) {
        emit(UpdateProfileSuccess());
      } else {
        emit(const UpdateProfileFailed(message: 'Not Found'));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(UpdateProfileFailed(message: error['message']));
    }
  }
}
