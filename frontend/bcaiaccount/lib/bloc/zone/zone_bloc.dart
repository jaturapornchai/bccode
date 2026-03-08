import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/zone_data_model.dart';
import 'package:smlaicloud/repositories/zone_repository.dart';

part 'zone_event.dart';
part 'zone_state.dart';

class ZoneBloc extends Bloc<ZoneEvent, ZoneState> {
  final ZoneRepository _zoneRepository;

  ZoneBloc({
    required ZoneRepository zoneRepository,
  })  : _zoneRepository = zoneRepository,
        super(ZoneInitial()) {
    on<ZoneLoadList>(onZoneLoad);
    on<ZoneSave>(onZoneSave);
    on<ZoneUpdate>(onZoneUpdate);
    on<ZoneDelete>(onZoneDelete);
    on<ZoneDeleteMany>(onZoneDeleteMany);
    on<ZoneGet>(onZoneGet);
  }

  void onZoneLoad(ZoneLoadList event, Emitter<ZoneState> emit) async {
    emit(ZoneInProgress());

    try {
      final results = await _zoneRepository.getZoneList(
          offset: event.offset,
          limit: event.limit,
          search: event.search,
          groupNumber: event.groupNumber);

      if (results.success) {
        List<ZoneDataModel> zones = (results.data as List)
            .map((zone) => ZoneDataModel.fromJson(zone))
            .toList();
        emit(ZoneLoadSuccess(zones: zones));
      } else {
        emit(const ZoneLoadFailed(message: 'Zone Group Not Found'));
      }
    } catch (e) {
      emit(ZoneLoadFailed(message: e.toString()));
    }
  }

  void onZoneDelete(ZoneDelete event, Emitter<ZoneState> emit) async {
    emit(ZoneDeleteInProgress());
    try {
      await _zoneRepository.deleteZone(event.guid);

      emit(ZoneDeleteSuccess());
    } catch (e) {
      // emit(ZoneDeleteFailure(message: e.toString()));
    }
  }

  void onZoneDeleteMany(ZoneDeleteMany event, Emitter<ZoneState> emit) async {
    emit(ZoneDeleteManyInProgress());
    try {
      await _zoneRepository.deleteZoneMany(event.guid);

      emit(ZoneDeleteManySuccess());
    } catch (e) {
      // emit(ZoneDeleteFailure(message: e.toString()));
    }
  }

  void onZoneSave(ZoneSave event, Emitter<ZoneState> emit) async {
    emit(ZoneSaveInProgress());
    try {
      await _zoneRepository.saveZone(event.zoneDataModel);
      emit(ZoneSaveSuccess());
    } catch (e) {
      emit(ZoneSaveFailed(message: e.toString()));
    }
  }

  void onZoneUpdate(ZoneUpdate event, Emitter<ZoneState> emit) async {
    emit(ZoneUpdateInProgress());
    try {
      await _zoneRepository.updateZone(event.guid, event.zoneDataModel);
      emit(ZoneUpdateSuccess());
    } catch (e) {
      emit(ZoneUpdateFailed(message: e.toString()));
    }
  }

  void onZoneGet(ZoneGet event, Emitter<ZoneState> emit) async {
    emit(ZoneGetInProgress());
    try {
      final result = await _zoneRepository.getZone(event.guid);
      if (result.success) {
        ZoneDataModel zone = ZoneDataModel.fromJson(result.data);

        emit(ZoneGetSuccess(zone: zone));
      } else {
        emit(const ZoneGetFailed(message: 'Zone Not Found'));
      }
    } catch (e) {
      emit(ZoneDeleteFailed(message: e.toString()));
    }
  }
}
