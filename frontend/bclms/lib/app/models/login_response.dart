class LoginResponse {
  final bool success;
  final String? token;
  final String? refresh;
  final String? message;

  LoginResponse({
    required this.success,
    this.token,
    this.refresh,
    this.message,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) {
    return LoginResponse(
      success: json['success'] ?? false,
      token: json['token'],
      refresh: json['refresh'],
      message: json['message'],
    );
  }
}
