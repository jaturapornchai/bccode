class BranchModel {
  final String guidFixed;
  final List<BranchName> names;

  BranchModel({
    required this.guidFixed,
    required this.names,
  });

  factory BranchModel.fromJson(Map<String, dynamic> json) {
    final nameList = (json['names'] as List?)
            ?.map((e) => BranchName.fromJson(e))
            .toList() ??
        [];
    return BranchModel(
      guidFixed: json['guidfixed'] ?? '',
      names: nameList,
    );
  }

  /// ดึงชื่อสาขาตามภาษา (default: th)
  String getDisplayName([String langCode = 'th']) {
    final match = names.where((n) => n.code == langCode);
    if (match.isNotEmpty) return match.first.name;
    if (names.isNotEmpty) return names.first.name;
    return guidFixed;
  }
}

class BranchName {
  final String code;
  final String name;

  BranchName({required this.code, required this.name});

  factory BranchName.fromJson(Map<String, dynamic> json) {
    return BranchName(
      code: json['code'] ?? '',
      name: json['name'] ?? '',
    );
  }
}
