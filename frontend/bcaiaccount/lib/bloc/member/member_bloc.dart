import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/member_repository.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/model/member_model.dart';

part 'member_event.dart';
part 'member_state.dart';

class MemberBloc extends Bloc<MemberEvent, MemberState> {
  final MemberRepository _memberRepository;

  MemberBloc({required MemberRepository memberRepository})
    : _memberRepository = memberRepository,
      super(MemberInitial()) {
    on<ListMemberLoad>(_onMemberLoad);
    on<ListMemberLoadById>(_onGetMemberId);
    on<MemberSaved>(_onMemberSaved);
    on<MemberUpdate>(_onMemberUpdate);
    on<MemberDelete>(_onMemberDelete);
  }
  void _onMemberLoad(ListMemberLoad event, Emitter<MemberState> emit) async {
    MemberLoadSuccess memberLoadSuccess;
    List<MemberModel> previousMember = [];
    if (state is MemberLoadSuccess) {
      memberLoadSuccess = (state as MemberLoadSuccess).copyWith();
      previousMember = memberLoadSuccess.member;
    }
    emit(MemberInProgress());

    try {
      final result = await _memberRepository.getMemberList(
        perPage: event.perPage,
        page: event.page,
        search: event.search,
      );

      if (result.success) {
        if (event.nextPage) {
          List<MemberModel> member = (result.data as List)
              .map((member) => MemberModel.fromJson(member))
              .toList();
          // print(_member);
          emit(MemberLoadSuccess(member: member, page: result.page));
        } else {
          List<MemberModel> member = (result.data as List)
              .map((member) => MemberModel.fromJson(member))
              .toList();
          // print(_member);
          previousMember.addAll(member);
          emit(MemberLoadSuccess(member: previousMember, page: result.page));
        }
      } else {
        emit(MemberLoadFailed(message: 'Member Not Found'));
      }
    } catch (e) {
      emit(MemberLoadFailed(message: e.toString()));
    }
  }

  void _onGetMemberId(
    ListMemberLoadById event,
    Emitter<MemberState> emit,
  ) async {
    emit(MemberLoadByIdInProgress());
    try {
      final result = await _memberRepository.getMemberId(event.id);

      if (result.success) {
        MemberModel member = MemberModel.fromJson(result.data);
        // print(_member);
        emit(MemberLoadByIdLoadSuccess(member: member));
      } else {
        emit(MemberLoadByIdLoadFailed(message: 'Member Not Found'));
      }
    } catch (e) {
      emit(MemberLoadByIdLoadFailed(message: e.toString()));
    }
  }

  void _onMemberSaved(MemberSaved event, Emitter<MemberState> emit) async {
    emit(MemberFormSaveInProgress());
    try {
      // print(event.member.toString());

      await _memberRepository.saveMember(event.member);
      // print('Success');
      emit(MemberFormSaveSuccess());
    } catch (e) {
      emit(MemberFormSaveFailure(message: e.toString()));
    }
  }

  void _onMemberUpdate(MemberUpdate event, Emitter<MemberState> emit) async {
    emit(MemberUpdateInProgress());
    try {
      // // print(event.inventory.toString());

      await _memberRepository.updateMember(event.member);

      emit(MemberUpdateSuccess());
    } catch (e) {
      emit(MemberUpdateFailure(message: e.toString()));
    }
  }

  void _onMemberDelete(MemberDelete event, Emitter<MemberState> emit) async {
    emit(MemberUpdateInProgress());
    try {
      // // print(event.inventory.toString());

      await _memberRepository.deleteMember(event.id);

      emit(MemberDeleteSuccess());
    } catch (e) {
      emit(MemberDeleteFailure(message: e.toString()));
    }
  }
}
