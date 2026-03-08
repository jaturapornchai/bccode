import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'wallet_model.g.dart';

@JsonSerializable(explicitToJson: true)
class WalletModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];
  String apikey;

  WalletModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
    required this.apikey,
  }) : names = names ?? <LanguageDataModel>[];

  factory WalletModel.fromJson(Map<String, dynamic> json) =>
      _$WalletModelFromJson(json);

  Map<String, dynamic> toJson() => _$WalletModelToJson(this);
}
