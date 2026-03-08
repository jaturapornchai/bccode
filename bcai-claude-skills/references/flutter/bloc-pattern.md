# BLoC Pattern — Code Templates เต็มรูปแบบ

## ไฟล์ที่ต้องสร้าง (3 ไฟล์)

### 1. `{feature}_bloc.dart` — ไฟล์หลัก

```dart
import 'dart:convert';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/{feature}_repository.dart';
import 'package:smlaicloud/model/{feature}_model.dart';

part '{feature}_event.dart';
part '{feature}_state.dart';

class FeatureBloc extends Bloc<FeatureEvent, FeatureState> {
  final FeatureRepository _repository;

  FeatureBloc({required FeatureRepository repository})
      : _repository = repository,
        super(FeatureInitial()) {
    on<FeatureLoadList>(onFeatureLoad);
    on<FeatureSave>(onFeatureSave);
    on<FeatureUpdate>(onFeatureUpdate);
    on<FeatureDelete>(onFeatureDelete);
    on<FeatureGet>(onFeatureGet);
  }

  void onFeatureLoad(FeatureLoadList event, Emitter<FeatureState> emit) async {
    emit(FeatureInProgress());
    try {
      final results = await _repository.getList(
        shopId: event.shopId,
        offset: event.offset,
        limit: event.limit,
      );
      if (results.success) {
        List<FeatureModel> items = (results.data as List)
            .map((item) => FeatureModel.fromJson(item))
            .toList();
        emit(FeatureLoadSuccess(items: items));
      } else {
        emit(FeatureLoadFailed(message: results.message));
      }
    } catch (e) {
      emit(FeatureLoadFailed(message: e.toString()));
    }
  }

  void onFeatureSave(FeatureSave event, Emitter<FeatureState> emit) async {
    emit(FeatureSaveInProgress());
    try {
      await _repository.save(event.item);
      emit(FeatureSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(FeatureSaveFailed(message: error['message']));
    }
  }

  void onFeatureUpdate(FeatureUpdate event, Emitter<FeatureState> emit) async {
    emit(FeatureUpdateInProgress());
    try {
      await _repository.update(event.guid, event.item);
      emit(FeatureUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(FeatureUpdateFailed(message: error['message']));
    }
  }

  void onFeatureDelete(FeatureDelete event, Emitter<FeatureState> emit) async {
    emit(FeatureDeleteInProgress());
    try {
      await _repository.delete(event.guid);
      emit(FeatureDeleteSuccess());
    } catch (e) {
      emit(FeatureDeleteFailed(message: e.toString()));
    }
  }

  void onFeatureGet(FeatureGet event, Emitter<FeatureState> emit) async {
    emit(FeatureGetInProgress());
    try {
      final result = await _repository.get(event.guid);
      if (result.success) {
        emit(FeatureGetSuccess(item: FeatureModel.fromJson(result.data)));
      } else {
        emit(FeatureGetFailed(message: result.message));
      }
    } catch (e) {
      emit(FeatureGetFailed(message: e.toString()));
    }
  }
}
```

### 2. `{feature}_state.dart`

```dart
part of '{feature}_bloc.dart';

abstract class FeatureState extends Equatable {
  const FeatureState();
  @override
  List<Object> get props => [];
}

// Initial
class FeatureInitial extends FeatureState {}

// Load
class FeatureInProgress extends FeatureState {}
class FeatureLoadSuccess extends FeatureState {
  final List<FeatureModel> items;
  const FeatureLoadSuccess({required this.items});
  @override
  List<Object> get props => [items];
}
class FeatureLoadFailed extends FeatureState {
  final String message;
  const FeatureLoadFailed({required this.message});
  @override
  List<Object> get props => [message];
}

// Get
class FeatureGetInProgress extends FeatureState {}
class FeatureGetSuccess extends FeatureState {
  final FeatureModel item;
  const FeatureGetSuccess({required this.item});
  @override
  List<Object> get props => [item];
}
class FeatureGetFailed extends FeatureState {
  final String message;
  const FeatureGetFailed({required this.message});
  @override
  List<Object> get props => [message];
}

// Save
class FeatureSaveInProgress extends FeatureState {}
class FeatureSaveSuccess extends FeatureState {}
class FeatureSaveFailed extends FeatureState {
  final String message;
  const FeatureSaveFailed({required this.message});
  @override
  List<Object> get props => [message];
}

// Update
class FeatureUpdateInProgress extends FeatureState {}
class FeatureUpdateSuccess extends FeatureState {}
class FeatureUpdateFailed extends FeatureState {
  final String message;
  const FeatureUpdateFailed({required this.message});
  @override
  List<Object> get props => [message];
}

// Delete
class FeatureDeleteInProgress extends FeatureState {}
class FeatureDeleteSuccess extends FeatureState {}
class FeatureDeleteFailed extends FeatureState {
  final String message;
  const FeatureDeleteFailed({required this.message});
  @override
  List<Object> get props => [message];
}
```

### 3. `{feature}_event.dart`

```dart
part of '{feature}_bloc.dart';

abstract class FeatureEvent extends Equatable {
  const FeatureEvent();
  @override
  List<Object> get props => [];
}

class FeatureLoadList extends FeatureEvent {
  final String shopId;
  final int offset;
  final int limit;
  const FeatureLoadList({
    required this.shopId,
    this.offset = 0,
    this.limit = 20,
  });
  @override
  List<Object> get props => [shopId, offset, limit];
}

class FeatureGet extends FeatureEvent {
  final String guid;
  const FeatureGet({required this.guid});
  @override
  List<Object> get props => [guid];
}

class FeatureSave extends FeatureEvent {
  final FeatureModel item;
  const FeatureSave({required this.item});
  @override
  List<Object> get props => [item];
}

class FeatureUpdate extends FeatureEvent {
  final String guid;
  final FeatureModel item;
  const FeatureUpdate({required this.guid, required this.item});
  @override
  List<Object> get props => [guid, item];
}

class FeatureDelete extends FeatureEvent {
  final String guid;
  const FeatureDelete({required this.guid});
  @override
  List<Object> get props => [guid];
}
```

## UI — BlocBuilder Pattern

```dart
BlocBuilder<FeatureBloc, FeatureState>(
  builder: (context, state) {
    if (state is FeatureInProgress) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state is FeatureLoadSuccess) {
      return ListView.builder(
        itemCount: state.items.length,
        itemBuilder: (_, i) => FeatureListTile(item: state.items[i]),
      );
    }
    if (state is FeatureLoadFailed) {
      return Center(child: Text(state.message));
    }
    return const SizedBox.shrink();
  },
)
```

## BlocListener สำหรับ Side Effects

```dart
BlocListener<FeatureBloc, FeatureState>(
  listener: (context, state) {
    if (state is FeatureSaveSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('บันทึกสำเร็จ')),
      );
      Navigator.pop(context);
    }
    if (state is FeatureSaveFailed) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(state.message)),
      );
    }
  },
  child: ...,
)
```

## Register BlocProvider

```dart
BlocProvider(
  create: (context) => FeatureBloc(
    repository: FeatureRepository(),
  )..add(FeatureLoadList(shopId: shopId)),
  child: FeatureScreen(),
)
```
