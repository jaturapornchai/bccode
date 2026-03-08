class ProfileModel {
  final String username;
  final String name;
  final String avatar;

  ProfileModel({
    required this.username,
    required this.name,
    required this.avatar,
  });

  factory ProfileModel.fromJson(Map<String, dynamic> json) {
    return ProfileModel(
      username: json['username'] ?? '',
      name: json['name'] ?? '',
      avatar: json['avatar'] ?? '',
    );
  }
}
