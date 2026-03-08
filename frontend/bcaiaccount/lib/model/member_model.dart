import 'package:json_annotation/json_annotation.dart';

part 'member_model.g.dart';

@JsonSerializable()
class MemberModel {
  String guidfixed;
  String address;
  String branchcode;
  int branchtype;
  int contacttype;
  String name;
  int personaltype;
  String surname;
  String taxid;
  String telephone;
  String zipcode;

  MemberModel({
    String? guidfixed,
    String? address,
    String? branchcode,
    required this.branchtype,
    required this.contacttype,
    required this.name,
    required this.personaltype,
    String? surname,
    String? taxid,
    String? telephone,
    String? zipcode,
  })  : guidfixed = guidfixed ?? '',
        address = address ?? '',
        branchcode = branchcode ?? '',
        surname = surname ?? '',
        taxid = taxid ?? '',
        telephone = telephone ?? '',
        zipcode = zipcode ?? '';

  factory MemberModel.fromJson(Map<String, dynamic> json) =>
      _$MemberModelFromJson(json);

  Map<String, dynamic> toJson() => _$MemberModelToJson(this);
}
