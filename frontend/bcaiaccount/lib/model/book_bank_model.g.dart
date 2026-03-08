// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'book_bank_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

BookBankModel _$BookBankModelFromJson(Map<String, dynamic> json) =>
    BookBankModel(
      passbook: json['passbook'] as String?,
      bookcode: json['bookcode'] as String?,
      bankcode: json['bankcode'] as String?,
      guidfixed: json['guidfixed'] as String?,
      images: (json['images'] as List<dynamic>?)
          ?.map((e) => ImagesModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      banknames: (json['banknames'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      accountcode: json['accountcode'] as String?,
      accountname: json['accountname'] as String?,
      bankbranch: json['bankbranch'] as String?,
    );

Map<String, dynamic> _$BookBankModelToJson(BookBankModel instance) =>
    <String, dynamic>{
      'passbook': instance.passbook,
      'bookcode': instance.bookcode,
      'bankcode': instance.bankcode,
      'guidfixed': instance.guidfixed,
      'images': instance.images?.map((e) => e.toJson()).toList(),
      'names': instance.names?.map((e) => e.toJson()).toList(),
      'banknames': instance.banknames?.map((e) => e.toJson()).toList(),
      'accountcode': instance.accountcode,
      'accountname': instance.accountname,
      'bankbranch': instance.bankbranch,
    };
