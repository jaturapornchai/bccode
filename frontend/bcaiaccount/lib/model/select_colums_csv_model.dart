import 'package:json_annotation/json_annotation.dart';

part 'select_colums_csv_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ListColumsCsvModel {
  String code;
  String name;

  ListColumsCsvModel({
    this.code = '',
    this.name = '',
  });

  factory ListColumsCsvModel.fromJson(Map<String, dynamic> json) =>
      _$ListColumsCsvModelFromJson(json);

  Map<String, dynamic> toJson() => _$ListColumsCsvModelToJson(this);
}
