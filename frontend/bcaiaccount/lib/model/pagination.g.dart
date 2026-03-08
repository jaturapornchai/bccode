// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'pagination.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Pagination _$PaginationFromJson(Map<String, dynamic> json) => Pagination(
  total: (json['total'] as num).toInt(),
  page: (json['page'] as num).toInt(),
  perPage: (json['perPage'] as num).toInt(),
  prev: (json['prev'] as num).toInt(),
  next: (json['next'] as num).toInt(),
  totalPage: (json['totalPage'] as num).toInt(),
);

Map<String, dynamic> _$PaginationToJson(Pagination instance) =>
    <String, dynamic>{
      'total': instance.total,
      'page': instance.page,
      'perPage': instance.perPage,
      'prev': instance.prev,
      'next': instance.next,
      'totalPage': instance.totalPage,
    };
