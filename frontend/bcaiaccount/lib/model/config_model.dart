import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/price_model.dart';

part 'config_model.g.dart';

class ConfigModel {
  late List<LanguageModel> languages;
  late List<PriceModel> prices;
  double vatrate = 7.0;
  int vattypesale = 0;
  int vattypepurchase = 0;
  int inquirytypesale = 0;
  int inquirytypepurchase = 0;

  int findLanguageIndex(String code) {
    int index = 0;
    for (int i = 0; i < languages.length; i++) {
      if (languages[i].code == code) {
        index = i;
        break;
      }
    }
    return index;
  }

  String findLanguageName(String code) {
    String name = "";
    for (int i = 0; i < languages.length; i++) {
      if (languages[i].code == code) {
        name = languages[i].name!;
        break;
      }
    }
    return name;
  }
}

@JsonSerializable(explicitToJson: true)
class DayWorkTimeModel {
  /// 0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday
  int dayNumber = 0;
  List<String> openTime = [];
  List<String> closeTime = [];
  bool isClose = false;

  DayWorkTimeModel({
    required this.dayNumber,
    required this.openTime,
    required this.closeTime,
    required this.isClose,
  });

  factory DayWorkTimeModel.fromJson(Map<String, dynamic> json) =>
      _$DayWorkTimeModelFromJson(json);

  Map<String, dynamic> toJson() => _$DayWorkTimeModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class DayCloseModel {
  List<String> dates = [];
  List<String> openTime = [];
  List<String> closeTime = [];
  bool isClose = false;

  DayCloseModel({
    required this.dates,
    required this.openTime,
    required this.closeTime,
    required this.isClose,
  });

  factory DayCloseModel.fromJson(Map<String, dynamic> json) =>
      _$DayCloseModelFromJson(json);

  Map<String, dynamic> toJson() => _$DayCloseModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ConfigSystemModel {
  List<String> languageList = [];
  double? vatrate;
  int? vattypesale;
  int? vattypepurchase;
  int? inquirytypesale;
  int? inquirytypepurchase;
  String? headerreceiptpos;
  String? footerreceiptpos;

  ConfigSystemModel({
    required this.languageList,
    double? vatrate,
    int? vattypesale,
    int? vattypepurchase,
    int? inquirytypesale,
    int? inquirytypepurchase,
    String? headerreceiptpos,
    String? footerreceiptpos,
  })  : vattypepurchase = vattypepurchase ?? 0,
        inquirytypesale = inquirytypesale ?? 0,
        inquirytypepurchase = inquirytypepurchase ?? 0,
        vattypesale = vattypesale ?? 0,
        vatrate = vatrate ?? 7.0,
        headerreceiptpos = headerreceiptpos ?? "",
        footerreceiptpos = footerreceiptpos ?? "";

  factory ConfigSystemModel.fromJson(Map<String, dynamic> json) =>
      _$ConfigSystemModelFromJson(json);

  Map<String, dynamic> toJson() => _$ConfigSystemModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class CompanyModel {
  List<LanguageDataModel> names;
  String taxID;
  List<LanguageDataModel> branchNames;
  List<LanguageDataModel> addresses;
  List<String> phones;
  List<String> emailOwners;
  List<String> emailStaffs;
  String latitude;
  String longitude;
  bool? usebranch;
  bool? usedepartment;
  List<ImagesModel> images;
  String? logo;
  CompanyModel({
    required this.names,
    required this.taxID,
    required this.branchNames,
    required this.addresses,
    required this.phones,
    required this.emailOwners,
    required this.emailStaffs,
    required this.latitude,
    required this.longitude,
    bool? usebranch,
    bool? usedepartment,
    required this.images,
    String? logo,
  })  : usebranch = usebranch ?? false,
        usedepartment = usedepartment ?? false,
        logo = logo ?? "";

  factory CompanyModel.fromJson(Map<String, dynamic> json) =>
      _$CompanyModelFromJson(json);

  Map<String, dynamic> toJson() => _$CompanyModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class WorkingTimeModel {
  final List<DayWorkTimeModel> dayWorkTimes;
  final List<DayCloseModel> dayCloses;

  WorkingTimeModel({
    required this.dayWorkTimes,
    required this.dayCloses,
  });

  factory WorkingTimeModel.fromJson(Map<String, dynamic> json) =>
      _$WorkingTimeModelFromJson(json);

  Map<String, dynamic> toJson() => _$WorkingTimeModelToJson(this);
}

@JsonSerializable()
class DeviceConfigModel {
  /// ขนาด Font List ข้อมูล
  double listDataFontSize;

  /// ระยะห่างระหว่างบรรทัด List ข้อมูล
  double listDataLineSpace;

  /// หน้าจอสินค้า แสดง SKU
  bool itemDisplaySku;

  /// หน้าจอสินค้า แสดง ราคา
  bool itemDisplayPrice;

  DeviceConfigModel({
    required this.listDataFontSize,
    required this.listDataLineSpace,
    required this.itemDisplaySku,
    required this.itemDisplayPrice,
  });

  factory DeviceConfigModel.fromJson(Map<String, dynamic> json) =>
      _$DeviceConfigModelFromJson(json);

  Map<String, dynamic> toJson() => _$DeviceConfigModelToJson(this);
}
