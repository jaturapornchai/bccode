import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'contact_model.g.dart';

@JsonSerializable()
class ContactModel {
  List<LanguageDataModel>? address;
  String? countrycode;
  String? districtcode;
  double? latitude;
  double? longitude;
  String? phonenumber;
  String? provincecode;
  String? subdistrictcode;
  String? zipcode;

  ContactModel({
    List<LanguageDataModel>? address,
    String? countrycode,
    String? districtcode,
    double? latitude,
    double? longitude,
    String? phonenumber,
    String? provincecode,
    String? subdistrictcode,
    String? zipcode,
  })  : address = address ?? <LanguageDataModel>[],
        countrycode = countrycode ?? "",
        districtcode = districtcode ?? "",
        latitude = latitude ?? 0,
        longitude = longitude ?? 0,
        phonenumber = phonenumber ?? "",
        provincecode = provincecode ?? "",
        subdistrictcode = subdistrictcode ?? "",
        zipcode = zipcode ?? "";

  factory ContactModel.fromJson(Map<String, dynamic> json) =>
      _$ContactModelFromJson(json);
  Map<String, dynamic> toJson() => _$ContactModelToJson(this);
}

@JsonSerializable()
class ContactEmployeeModel {
  String? address;
  String? countrycode;
  String? districtcode;
  double? latitude;
  double? longitude;
  String? phonenumber;
  String? provincecode;
  String? subdistrictcode;
  String? zipcode;

  ContactEmployeeModel({
    String? address,
    String? countrycode,
    String? districtcode,
    double? latitude,
    double? longitude,
    String? phonenumber,
    String? provincecode,
    String? subdistrictcode,
    String? zipcode,
  })  : address = address ?? "",
        countrycode = countrycode ?? "",
        districtcode = districtcode ?? "",
        latitude = latitude ?? 0,
        longitude = longitude ?? 0,
        phonenumber = phonenumber ?? "",
        provincecode = provincecode ?? "",
        subdistrictcode = subdistrictcode ?? "",
        zipcode = zipcode ?? "";

  factory ContactEmployeeModel.fromJson(Map<String, dynamic> json) =>
      _$ContactEmployeeModelFromJson(json);
  Map<String, dynamic> toJson() => _$ContactEmployeeModelToJson(this);
}
