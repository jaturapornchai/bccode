import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/printer_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'zone_data_model.g.dart';

@JsonSerializable()
class ZoneDataModel {
  late String guidfixed;
  late String code;
  late List<LanguageDataModel> names;
  late PrinterModel? printers;
  int? groupnumber;

  ZoneDataModel({
    required this.guidfixed,
    required this.code,
    required this.names,
    PrinterModel? printers,
    int? groupnumber,
  })  : printers = printers ??
            PrinterModel(
                address: '', code: '', guidfixed: '', names: [], type: 0),
        groupnumber = groupnumber ?? 1;

  factory ZoneDataModel.fromJson(Map<String, dynamic> json) =>
      _$ZoneDataModelFromJson(json);
  Map<String, dynamic> toJson() => _$ZoneDataModelToJson(this);
}
