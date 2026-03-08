class ShopModel {
  final String shopId;
  final String shopName;
  final int role;

  ShopModel({
    required this.shopId,
    required this.shopName,
    required this.role,
  });

  factory ShopModel.fromJson(Map<String, dynamic> json) {
    return ShopModel(
      shopId: json['shopid'] ?? '',
      shopName: json['shopname'] ?? '',
      role: json['role'] ?? 0,
    );
  }

  String get roleName {
    switch (role) {
      case 2:
        return 'Super Admin';
      case 1:
        return 'Admin';
      default:
        return 'User';
    }
  }
}
