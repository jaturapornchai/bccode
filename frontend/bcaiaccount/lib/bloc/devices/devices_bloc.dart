import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/devices_model.dart';
import 'package:smlaicloud/repositories/devices_repository.dart';

part 'devices_event.dart';
part 'devices_state.dart';

class DevicesBloc extends Bloc<DevicesEvent, DevicesState> {
  final DevicesRepository _devicesRepository;

  DevicesBloc({required DevicesRepository devicesRepository})
      : _devicesRepository = devicesRepository,
        super(DevicesInitial()) {
    on<DevicesLoadList>(onDevicesLoad);
    on<DevicesSave>(onDevicesSave);
    on<DevicesUpdate>(onDevicesUpdate);
    on<DevicesDelete>(onDevicesDelete);
    on<DevicesDeleteMany>(onDevicesDeleteMany);
    on<DevicesGet>(onDevicesGet);
  }

  void onDevicesLoad(DevicesLoadList event, Emitter<DevicesState> emit) async {
    emit(DevicesInProgress());

    try {
      final results = await _devicesRepository.getDeviceList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<DevicesModel> Devicess = (results.data as List)
            .map((Devicess) => DevicesModel.fromJson(Devicess))
            .toList();
        emit(DevicesLoadSuccess(devices: Devicess));
      } else {
        emit(const DevicesLoadFailed(message: 'Devices Group Not Found'));
      }
    } catch (e) {
      emit(DevicesLoadFailed(message: e.toString()));
    }
  }

  void onDevicesDelete(DevicesDelete event, Emitter<DevicesState> emit) async {
    emit(DevicesDeleteInProgress());
    try {
      await _devicesRepository.deleteDevice(event.guid);

      emit(DevicesDeleteSuccess());
    } catch (e) {
      // emit(DevicesDeleteFailure(message: e.toString()));
    }
  }

  void onDevicesDeleteMany(
      DevicesDeleteMany event, Emitter<DevicesState> emit) async {
    emit(DevicesDeleteManyInProgress());
    try {
      await _devicesRepository.deleteDeviceMany(event.guid);

      emit(DevicesDeleteManySuccess());
    } catch (e) {
      // emit(DevicesDeleteFailure(message: e.toString()));
    }
  }

  void onDevicesSave(DevicesSave event, Emitter<DevicesState> emit) async {
    emit(DevicesSaveInProgress());
    try {
      await _devicesRepository.saveDevice(event.devices);
      emit(DevicesSaveSuccess());
    } catch (e) {
      emit(DevicesSaveFailed(message: e.toString()));
    }
  }

  void onDevicesUpdate(DevicesUpdate event, Emitter<DevicesState> emit) async {
    emit(DevicesUpdateInProgress());
    try {
      await _devicesRepository.updateDevice(event.guid, event.devices);
      emit(DevicesUpdateSuccess());
    } catch (e) {
      emit(DevicesUpdateFailed(message: e.toString()));
    }
  }

  void onDevicesGet(DevicesGet event, Emitter<DevicesState> emit) async {
    emit(DevicesGetInProgress());
    try {
      final result = await _devicesRepository.getDevice(event.guid);
      if (result.success) {
        DevicesModel Devicess = DevicesModel.fromJson(result.data);
        emit(DevicesGetSuccess(devices: Devicess));
      } else {
        emit(const DevicesGetFailed(message: 'Devices Not Found'));
      }
    } catch (e) {
      // emit(DevicesDeleteFailure(message: e.toString()));
    }
  }
}
