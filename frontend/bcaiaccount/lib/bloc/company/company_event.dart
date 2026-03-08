part of 'company_bloc.dart';

abstract class CompanyEvent extends Equatable {
  const CompanyEvent();

  @override
  List<Object> get props => [];
}

class CompanyGet extends CompanyEvent {
  final String guid;

  const CompanyGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class CompanyLoad extends CompanyEvent {
  const CompanyLoad();

  @override
  List<Object> get props => [];
}

class CompanyDelete extends CompanyEvent {
  final String guid;

  const CompanyDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CompanyDeleteMany extends CompanyEvent {
  final List<String> guid;

  const CompanyDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CompanySave extends CompanyEvent {
  final CompanyModel company;

  const CompanySave({
    required this.company,
  });

  @override
  List<Object> get props => [CompanyModel];
}

class CompanyWithImageSave extends CompanyEvent {
  final List<File> imageFiles;
  final CompanyModel company;
  final List<Uint8List> imageWeb;
  const CompanyWithImageSave({
    required this.imageWeb,
    required this.imageFiles,
    required this.company,
  });

  @override
  List<Object> get props => [CompanyModel, imageFiles];
}

class CompanyUpdate extends CompanyEvent {
  final String guid;
  final CompanyModel company;

  const CompanyUpdate({
    required this.guid,
    required this.company,
  });

  @override
  List<Object> get props => [CompanyModel];
}

class CompanyWithImageUpdate extends CompanyEvent {
  final String guid;
  final CompanyModel company;
  final List<File> imageFiles;
  final List<Uint8List> imageWeb;
  final List<ImagesModel> imagesUris;

  const CompanyWithImageUpdate({
    required this.guid,
    required this.company,
    required this.imageFiles,
    required this.imageWeb,
    required this.imagesUris,
  });

  @override
  List<Object> get props => [CompanyModel];
}
