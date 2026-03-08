import 'package:json_annotation/json_annotation.dart';

part 'pagination.g.dart';

@JsonSerializable()
class Pagination {
  int total = 0;
  int page = 0;
  int perPage = 0;
  int prev = 0;
  int next = 0;
  int totalPage = 0;

  Pagination(
      {required this.total,
      required this.page,
      required this.perPage,
      required this.prev,
      required this.next,
      required this.totalPage});

  factory Pagination.fromJson(Map<String, dynamic> json) =>
      _$PaginationFromJson(json);

  Map<String, dynamic> toJson() => _$PaginationToJson(this);
}
