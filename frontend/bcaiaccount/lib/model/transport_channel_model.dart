import 'package:json_annotation/json_annotation.dart';

part 'transport_channel_model.g.dart';

@JsonSerializable()
class TransportChannelModel {
  String guidfixed;
  String code;
  String name;
  String imageuri;

  TransportChannelModel({
    String? guidfixed,
    String? code,
    String? name,
    String? imageuri,
  })  : guidfixed = guidfixed ?? "",
        code = code ?? "",
        name = name ?? "",
        imageuri = imageuri ?? "";

  factory TransportChannelModel.fromJson(Map<String, dynamic> json) =>
      _$TransportChannelModelFromJson(json);

  Map<String, dynamic> toJson() => _$TransportChannelModelToJson(this);
}
