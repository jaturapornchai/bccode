import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'pos_media_model.g.dart';

@JsonSerializable()
class PosMediaModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> description;
  List<ResourceModel> resources;

  PosMediaModel({
    required this.guidfixed,
    required this.code,
    required this.description,
    required this.resources,
  });

  factory PosMediaModel.fromJson(Map<String, dynamic> json) =>
      _$PosMediaModelFromJson(json);
  Map<String, dynamic> toJson() => _$PosMediaModelToJson(this);
}

@JsonSerializable()
class ResourceModel {
  List<int> daysofweek;
  List<LanguageDataModel> description;
  int displaytime;
  String fromDate;
  String fromTime;
  int mediaType;
  String toDate;
  String toTime;
  String uri;
  String? uriVideo;
  String? urilink;

  ResourceModel({
    required this.daysofweek,
    required this.description,
    required this.displaytime,
    required this.fromDate,
    required this.fromTime,
    required this.mediaType,
    required this.toDate,
    required this.toTime,
    required this.uri,
    String? urilink,
    String? uriVideo,
  })  : urilink = urilink ?? '',
        uriVideo = uriVideo ?? '';

  factory ResourceModel.fromJson(Map<String, dynamic> json) =>
      _$ResourceModelFromJson(json);
  Map<String, dynamic> toJson() => _$ResourceModelToJson(this);
}
