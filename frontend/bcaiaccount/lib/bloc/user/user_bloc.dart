import 'package:smlaicloud/model/user_model.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/user_repository.dart';

part 'user_event.dart';
part 'user_state.dart';

class UserBloc extends Bloc<UserEvent, UserState> {
  final UserRepository _userRepository;

  UserBloc({required UserRepository userRepository})
      : _userRepository = userRepository,
        super(UserInitial()) {
    on<UserLoadList>(onUserLoad);
    on<UserSaveAndUpdate>(onUserSaveOrUpdate);
    on<UserDelete>(onUserDelete);
    on<UserGet>(onUserGet);
    on<UserSaveMyLineData>(onUserSaveMyLineData);
  }

  void onUserLoad(UserLoadList event, Emitter<UserState> emit) async {
    emit(UserInProgress());
    try {
      final results = await _userRepository.getUserList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<UserModel> users = (results.data as List)
            .map((user) => UserModel.fromJson(user))
            .toList();
        // print(users.length);
        emit(UserLoadSuccess(users: users));
      } else {
        emit(UserLoadFailed(message: 'User Not Found ${results.message}'));
      }
    } catch (e) {
      emit(UserLoadFailed(message: e.toString()));
    }
  }

  void onUserDelete(UserDelete event, Emitter<UserState> emit) async {
    emit(UserDeleteInProgress());
    try {
      await _userRepository.deleteUser(event.username);

      emit(UserDeleteSuccess());
    } catch (e) {
      emit(UserDeleteFailed(message: e.toString()));
    }
  }

  void onUserSaveOrUpdate(
      UserSaveAndUpdate event, Emitter<UserState> emit) async {
    emit(UserSaveAndUpdateInProgress());
    try {
      await _userRepository.saveAndUpdateUser(event.userModel);
      emit(UserSaveAndUpdateSuccess());
    } catch (e) {
      emit(UserSaveAndUpdateFailed(message: e.toString()));
    }
  }

  void onUserGet(UserGet event, Emitter<UserState> emit) async {
    emit(UserGetInProgress());
    try {
      final result = await _userRepository.getUser(event.username);
      if (result.success) {
        UserModel user = UserModel.fromJson(result.data);
        emit(UserGetSuccess(user: user));
      } else {
        emit(const UserGetFailed(message: 'User Not Found'));
      }
    } catch (e) {
      emit(UserGetFailed(message: e.toString()));
    }
  }

  /// บันทึก LINE data ของตัวเอง (ใช้ /profile/my-line endpoint)
  void onUserSaveMyLineData(
      UserSaveMyLineData event, Emitter<UserState> emit) async {
    emit(UserSaveAndUpdateInProgress());
    try {
      await _userRepository.saveMyLineData(
        lineUserId: event.lineUserId,
        lineDisplayName: event.lineDisplayName,
        linePictureUrl: event.linePictureUrl,
      );
      emit(UserSaveAndUpdateSuccess());
    } catch (e) {
      emit(UserSaveAndUpdateFailed(message: e.toString()));
    }
  }
}
