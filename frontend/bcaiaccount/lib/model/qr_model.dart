import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'qr_model.g.dart';

@JsonSerializable(explicitToJson: true)
class QrModel {
  String? guidfixed;
  String? code;
  List<LanguageDataModel>? qrnames;

  /// PromptPay Company
  /// 100 = QR PromptPay , 101 = K Plus

  /// Lugent
  /// 110 =  Lugent PromptPat , 111 = Lugent AliPay , 112 = Lugent True Money , 113 = Lugent Line Pay

  /// GB
  /// 131 = GB PromptPay , 132 = GB True Money , 133 = GB Line Pay , 134 = GB AliPay , 135 = GB WeChat Pay,

  /// XENDIT
  /// 201 = XENDIT PromptPay , 202 = XENDIT True Money , 203 = XENDIT Line Pay , 204 = XENDIT AliPay , 205 = XENDIT WeChat Pay,

  /// SMLQRAPI
  /// 301 = SML PromptPay , 302 = SML Credit

  /// Tigerboard
  /// 401 = Tiger Board

  int? qrtype;
  bool? isactive;
  String? logo;

  String? bankcode;
  List<LanguageDataModel>? banknames;
  String? bookbankcode;
  List<LanguageDataModel>? bookbanknames;
  List<ImagesModel>? bookbankimages;
  String? qrcode;
  String? apikey;
  String? billerCode;
  String? billerID;
  String? storeID;
  String? terminalID;
  String? merchantName;
  String? accessCode;
  String? bankcharge;
  String? customercharge;
  String? secert;
  String? token;

  /// 0 = เงินเข้าทันที , 1 = สิ้นวัน , 2 = วันถัดไป
  int? closeqr;

  String? appid;
  String? host;

  QrModel({
    String? guidfixed,
    String? code,
    List<LanguageDataModel>? qrnames,
    int? qrtype,
    bool? isactive,
    String? logo,
    String? bankcode,
    List<LanguageDataModel>? banknames,
    String? bookbankcode,
    List<LanguageDataModel>? bookbanknames,
    List<ImagesModel>? bookbankimages,
    String? qrcode,
    String? apikey,
    String? billerCode,
    String? billerID,
    String? storeID,
    String? terminalID,
    String? merchantName,
    String? accessCode,
    String? bankcharge,
    String? customercharge,
    int? closeqr,
    String? secert,
    String? token,
    String? appid,
    String? host,
  }) : guidfixed = guidfixed ?? '',
       code = code ?? '',
       qrnames = qrnames ?? [],
       qrtype = qrtype ?? 0,
       isactive = isactive ?? true,
       logo = logo ?? '',
       bankcode = bankcode ?? '',
       banknames = banknames ?? [],
       bookbankcode = bookbankcode ?? '',
       bookbanknames = bookbanknames ?? [],
       bookbankimages = bookbankimages ?? [],
       qrcode = qrcode ?? '',
       apikey = apikey ?? '',
       billerCode = billerCode ?? '',
       billerID = billerID ?? '',
       storeID = storeID ?? '',
       terminalID = terminalID ?? '',
       merchantName = merchantName ?? '',
       accessCode = accessCode ?? '',
       bankcharge = bankcharge ?? '',
       customercharge = customercharge ?? '',
       closeqr = closeqr ?? 0,
       secert = secert ?? '',
       token = token ?? '',
       appid = appid ?? '',
       host = host ?? '';

  factory QrModel.fromJson(Map<String, dynamic> json) =>
      _$QrModelFromJson(json);

  Map<String, dynamic> toJson() => _$QrModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class QrModelList {
  String code;
  List<QrModel> qrlist;
  QrModelList({required this.code, required this.qrlist});

  factory QrModelList.fromJson(Map<String, dynamic> json) =>
      _$QrModelListFromJson(json);

  Map<String, dynamic> toJson() => _$QrModelListToJson(this);
}
