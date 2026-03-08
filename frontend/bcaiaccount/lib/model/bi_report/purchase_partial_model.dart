import 'package:json_annotation/json_annotation.dart';

part 'purchase_partial_model.g.dart';

@JsonSerializable()
class PurchasePartialCreditorName {
  final String code;
  final String name;
  final bool isauto;
  final bool isdelete;

  const PurchasePartialCreditorName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory PurchasePartialCreditorName.fromJson(Map<String, dynamic> json) =>
      _$PurchasePartialCreditorNameFromJson(json);

  Map<String, dynamic> toJson() => _$PurchasePartialCreditorNameToJson(this);
}

@JsonSerializable()
class PurchasePartialBranchName {
  final String code;
  final String name;
  final bool isauto;
  final bool isdelete;

  const PurchasePartialBranchName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory PurchasePartialBranchName.fromJson(Map<String, dynamic> json) =>
      _$PurchasePartialBranchNameFromJson(json);

  Map<String, dynamic> toJson() => _$PurchasePartialBranchNameToJson(this);
}

@JsonSerializable()
class PurchasePartialModel {
  final String guidfixed;
  final String docdate;
  final String docno;
  final String docrefno;
  final String creditorcode;
  final List<PurchasePartialCreditorName> creditornames;
  final double totalvalue;
  final double totalexceptvat;
  final double totalbeforevat;
  final double totalvatvalue;
  final double totalamount;
  final String branchcode;
  final List<PurchasePartialBranchName> branchnames;
  final List<dynamic> transactions;

  const PurchasePartialModel({
    required this.guidfixed,
    required this.docdate,
    required this.docno,
    required this.docrefno,
    required this.creditorcode,
    required this.creditornames,
    required this.totalvalue,
    required this.totalexceptvat,
    required this.totalbeforevat,
    required this.totalvatvalue,
    required this.totalamount,
    required this.branchcode,
    required this.branchnames,
    required this.transactions,
  });

  factory PurchasePartialModel.fromJson(Map<String, dynamic> json) =>
      _$PurchasePartialModelFromJson(json);

  Map<String, dynamic> toJson() => _$PurchasePartialModelToJson(this);

  // Helper methods to get names in Thai
  String get creditorNameTh {
    final thName = creditornames.firstWhere(
      (name) => name.code == 'th',
      orElse: () => creditornames.isNotEmpty
          ? creditornames.first
          : const PurchasePartialCreditorName(
              code: '',
              name: '',
              isauto: false,
              isdelete: false,
            ),
    );
    return thName.name;
  }

  String get branchNameTh {
    final thName = branchnames.firstWhere(
      (name) => name.code == 'th',
      orElse: () => branchnames.isNotEmpty
          ? branchnames.first
          : const PurchasePartialBranchName(
              code: '',
              name: '',
              isauto: false,
              isdelete: false,
            ),
    );
    return thName.name;
  }
}
