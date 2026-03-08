import 'package:json_annotation/json_annotation.dart';

part 'vat_sale_model.g.dart';

/// Model for VAT Sale Report detail data
/// Represents individual tax document records with creditor information
@JsonSerializable()
class VatSaleModel {
  /// Tax document date
  final String taxdocdate;

  /// Tax document number (invoice number)
  final String taxdocno;

  /// Document date
  final String docdate;

  /// Document number
  final String docno;

  /// Creditor code
  final String creditorcode;

  /// Creditor name (multilingual)
  final List<CreditorName> creditorname;

  /// Branch code
  final String branchcode;

  /// Tax ID number
  final String taxid;

  /// Total amount before VAT
  final double totalbeforevat;

  /// VAT value
  final double totalvatvalue;

  /// Amount exempt from VAT
  final double totalexceptvat;

  /// Total amount after VAT
  final double totalaftervat;

  /// Detail total discount
  final double detailtotaldiscount;

  /// Total amount (final)
  final double totalamount;

  /// Document status
  final String status;

  const VatSaleModel({
    required this.taxdocdate,
    required this.taxdocno,
    required this.docdate,
    required this.docno,
    required this.creditorcode,
    required this.creditorname,
    required this.branchcode,
    required this.taxid,
    required this.totalbeforevat,
    required this.totalvatvalue,
    required this.totalexceptvat,
    required this.totalaftervat,
    required this.detailtotaldiscount,
    required this.totalamount,
    required this.status,
  });

  factory VatSaleModel.fromJson(Map<String, dynamic> json) =>
      _$VatSaleModelFromJson(json);

  Map<String, dynamic> toJson() => _$VatSaleModelToJson(this);
}

/// Nested class for multilingual creditor names
@JsonSerializable()
class CreditorName {
  /// Language code (e.g., 'th', 'en')
  final String code;

  /// Creditor name in specified language
  final String name;

  /// Auto-generated flag
  final bool isauto;

  /// Deletion flag
  final bool isdelete;

  const CreditorName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory CreditorName.fromJson(Map<String, dynamic> json) =>
      _$CreditorNameFromJson(json);

  Map<String, dynamic> toJson() => _$CreditorNameToJson(this);
}
