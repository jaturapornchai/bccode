import 'package:json_annotation/json_annotation.dart';

part 'doc_payload_model.g.dart';

@JsonSerializable(explicitToJson: true, includeIfNull: false)
class DocPayLoadModel {
  final String shopid;
  final String system;
  final int offset;
  final int limit;
  final String search;
  final String custcode;
  final int dateorder;
  final String? fromdate;    // วันที่เริ่มต้น format: "2026-02-01"
  final String? todate;      // วันที่สิ้นสุด format: "2026-02-28"
  final double? minamount;   // ยอดเงินต่ำสุด
  final double? maxamount;   // ยอดเงินสูงสุด
  final List<String>? custcodes; // รายการเจ้าหนี้ (multi-select)

  DocPayLoadModel({
    required this.shopid,
    required this.system,
    required this.offset,
    required this.limit,
    required this.search,
    required this.custcode,
    required this.dateorder,
    this.fromdate,
    this.todate,
    this.minamount,
    this.maxamount,
    this.custcodes,
  });

  factory DocPayLoadModel.fromJson(Map<String, dynamic> json) =>
      _$DocPayLoadModelFromJson(json);
  Map<String, dynamic> toJson() => _$DocPayLoadModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class MongoGetDataPayloadModel {
  final String shopid;
  final String collection;
  final String guidfixed;

  MongoGetDataPayloadModel({
    required this.shopid,
    required this.collection,
    required this.guidfixed,
  });

  factory MongoGetDataPayloadModel.fromJson(Map<String, dynamic> json) =>
      _$MongoGetDataPayloadModelFromJson(json);
  Map<String, dynamic> toJson() => _$MongoGetDataPayloadModelToJson(this);
}
