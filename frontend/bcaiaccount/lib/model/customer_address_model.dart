import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'customer_address_model.g.dart';

@JsonSerializable(explicitToJson: true)
class CustomerAddressModel {
  late String? guid;

  /// ที่อยู่
  late List<String>? address;

  /// รหัสประเทศ
  late String? countrycode;

  /// รหัสจังหวัด
  late String? provincecode;

  /// รหัสอำเภอ
  late String? districtcode;

  /// รหัสตำบล
  late String? subdistrictcode;

  /// รหัสไปรษณีย์
  late String? zipcode;

  /// ผู้ติดต่อ
  late List<LanguageDataModel>? contactnames;

  /// เบอร์โทร
  late String? phoneprimary;

  /// เบอร์โทรสำรอง
  late String? phonesecondary;

  /// ตำแหน่งดาวเทียม
  late double? latitude;
  late double? longitude;

  CustomerAddressModel({
    String? guid,
    List<String>? address,
    String? countrycode,
    String? provincecode,
    String? districtcode,
    String? subdistrictcode,
    String? zipcode,
    List<LanguageDataModel>? contactnames,
    String? phoneprimary,
    String? phonesecondary,
    double? latitude,
    double? longitude,
  })  : guid = guid ?? '',
        address = address ?? [],
        countrycode = countrycode ?? '',
        provincecode = provincecode ?? '',
        districtcode = districtcode ?? '',
        subdistrictcode = subdistrictcode ?? '',
        zipcode = zipcode ?? '',
        contactnames = contactnames ?? [],
        phoneprimary = phoneprimary ?? '',
        phonesecondary = phonesecondary ?? '',
        latitude = latitude ?? 0.0,
        longitude = longitude ?? 0.0;

  factory CustomerAddressModel.fromJson(Map<String, dynamic> json) =>
      _$CustomerAddressModelFromJson(json);

  Map<String, dynamic> toJson() => _$CustomerAddressModelToJson(this);
}
