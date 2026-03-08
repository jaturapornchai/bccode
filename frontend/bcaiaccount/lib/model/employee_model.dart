import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/contact_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'employee_model.g.dart';

@JsonSerializable()
class EmployeeModel {
  String guidfixed;
  String code;
  String name;
  String email;
  String? profilepicture;
  bool? isenabled;
  ContactEmployeeModel? contact;
  String? pincode;
  bool? isusepos;
  List<CompanyBranchModel>? branches;

  EmployeeModel({
    required this.guidfixed,
    required this.code,
    String? name,
    String? email,
    String? profilepicture,
    bool? isenabled,
    ContactEmployeeModel? contact,
    String? pincode,
    bool? isusepos,
    List<CompanyBranchModel>? branches,
  })  : name = name ?? "",
        email = email ?? "",
        profilepicture = profilepicture ?? "",
        isenabled = isenabled ?? false,
        pincode = pincode ?? "",
        contact = contact ?? ContactEmployeeModel(),
        isusepos = isusepos ?? false,
        branches = branches ?? [];

  factory EmployeeModel.fromJson(Map<String, dynamic> json) =>
      _$EmployeeModelFromJson(json);
  Map<String, dynamic> toJson() => _$EmployeeModelToJson(this);
}
