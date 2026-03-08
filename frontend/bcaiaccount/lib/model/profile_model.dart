import 'package:json_annotation/json_annotation.dart';

part 'profile_model.g.dart';

@JsonSerializable()
class ProfileModel {
  String? username;
  String? name;
  String? avatar;
  String? timezonelabel;
  String? timezoneoffset;
  String? yeartype;

  // === ข้อมูล LINE (ระดับ user — ใช้ร่วมทุก shop) ===
  @JsonKey(name: 'line_user_id')
  String? lineUserId;
  @JsonKey(name: 'line_display_name')
  String? lineDisplayName;
  @JsonKey(name: 'line_picture_url')
  String? linePictureUrl;

  ProfileModel({
    String? username,
    String? name,
    String? avatar,
    String? timezonelabel,
    String? timezoneoffset,
    String? yeartype,
    this.lineUserId,
    this.lineDisplayName,
    this.linePictureUrl,
  })  : username = username ?? "",
        name = name ?? "",
        avatar = avatar ?? "",
        timezonelabel = timezonelabel ?? "",
        timezoneoffset = timezoneoffset ?? "",
        yeartype = yeartype ?? "";

  factory ProfileModel.fromJson(Map<String, dynamic> json) =>
      _$ProfileModelFromJson(json);
  Map<String, dynamic> toJson() => _$ProfileModelToJson(this);
}
