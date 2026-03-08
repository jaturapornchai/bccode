// Copyright 2013 The Flutter Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import 'dart:io';
import 'dart:convert';
import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

final FirebaseAuth _auth = FirebaseAuth.instance;

/// GoogleAuthHelper สำหรับ Windows Desktop OAuth
/// ใช้ custom OAuth2 flow กับ localhost redirect
class GoogleAuthHelper {
  // สร้างตัวแปร singleton เพื่อป้องกันการสร้าง instance ซ้ำซ้อน
  static GoogleAuthHelper? _instance;

  // OAuth2 configuration
  // ใช้ 127.0.0.1 แทน localhost ตามที่ Google แนะนำสำหรับ Desktop app
  // Port 8085 เพราะ 8080 อาจถูกใช้โดย service อื่น
  static const String _authorizationEndpoint = 'https://accounts.google.com/o/oauth2/v2/auth';
  static const String _tokenEndpoint = 'https://oauth2.googleapis.com/token';
  static const int _redirectPort = 8085;
  static const String _redirectUri = 'http://127.0.0.1:8085';

  HttpServer? _server;

  // Factory constructor สำหรับ singleton pattern
  factory GoogleAuthHelper() {
    _instance ??= GoogleAuthHelper._internal();
    return _instance!;
  }

  // Constructor แบบ private
  GoogleAuthHelper._internal();

  /// ล็อกอินด้วย Google Account สำหรับ Windows Desktop
  Future<UserCredential?> signIn() async {
    try {
      // เลือก credentials ตาม platform
      final clientId = Platform.isWindows ? global.googleSignInClientIdWindows : global.googleSignInClientId;
      final clientSecret = Platform.isWindows ? global.googleSignInClientSecretWindows : global.googleSignInClientSecret;

      if (kDebugMode) {
        AppLogger.debug("Starting Google Sign-In for Windows...");
        AppLogger.debug("Using clientId: ${clientId.substring(0, 20)}...");
        AppLogger.debug("Full clientId: $clientId");

        // ตรวจสอบ Firebase project configuration
        if (Firebase.apps.isNotEmpty) {
          final firebaseOptions = Firebase.apps.first.options;
          AppLogger.debug("🔍 Firebase App ID: ${firebaseOptions.appId}");
          AppLogger.debug("🔍 Firebase Project ID: ${firebaseOptions.projectId}");
          AppLogger.debug("🔍 Firebase Auth Domain: ${firebaseOptions.authDomain}");
          AppLogger.debug("🔍 Firebase API Key: ${firebaseOptions.apiKey}");

          // ตรวจสอบว่า Client ID ตรงกับ Firebase App ID หรือไม่
          // หมาย: Firebase App ID อาจมีหลาย App ID สำหรับแต่ละ platform
          // เช่น: 1:183871505957:android:xxx สำหรับ Android
          //      212036599086-cqd0bv7sq55u37j36oktqm88p5h3ir6b สำหรับ Windows
          if (firebaseOptions.appId.isNotEmpty) {
            AppLogger.debug("🔍 Checking if Client ID matches Firebase App ID...");
            AppLogger.debug("🔍 Client ID: $clientId");
            AppLogger.debug("🔍 Firebase App ID: ${firebaseOptions.appId}");

            // ตรวจสอบว่า Client ID ตรงกับ Firebase App ID หรือไม่
            // ถ้า Firebase App ID มี prefix เช่น "1:" หมายถึง Android App ID
            // ให้ตรวจสอบว่า Client ID ตรงกับส่วนหลังของ Firebase App ID
            final appIdParts = firebaseOptions.appId.split(':');
            final appIdWithoutPrefix = appIdParts.length > 1 ? appIdParts[1] : firebaseOptions.appId;

            if (clientId.contains(appIdWithoutPrefix)) {
              AppLogger.debug("✅ Client ID matches Firebase App ID");
            } else {
              AppLogger.warning("⚠️ Client ID does NOT match Firebase App ID!");
              AppLogger.warning("⚠️ Firebase App ID: $appIdWithoutPrefix");
              AppLogger.warning("⚠️ Client ID: $clientId");
              AppLogger.warning("⚠️ This may cause 'invalid-credential' error");
              AppLogger.warning("⚠️ Please check Firebase Console → Authentication → Sign-in method → Google");
            }
          }
        } else {
          AppLogger.warning("⚠️ Firebase is not initialized!");
        }
      }

      // สร้าง authorization URL
      final authUrl = Uri.parse(_authorizationEndpoint).replace(
        queryParameters: {
          'client_id': clientId,
          'redirect_uri': _redirectUri,
          'response_type': 'code',
          'scope': 'email profile openid',
          'access_type': 'offline',
          'prompt': 'select_account', // บังคับให้เลือก account ใหม่ทุกครั้ง
        },
      );

      if (kDebugMode) {
        AppLogger.debug("Authorization URL: $authUrl");
      }

      // เริ่ม local server เพื่อรับ callback
      final authCode = await _startServerAndGetAuthCode(authUrl);

      if (authCode == null) {
        if (kDebugMode) {
          AppLogger.error("Failed to get authorization code");
        }
        return null;
      }

      if (kDebugMode) {
        AppLogger.debug("Got authorization code: ${authCode.substring(0, 10)}...");
      }

      // แลก authorization code เป็น tokens
      final tokens = await _exchangeCodeForTokens(authCode, clientId, clientSecret);

      if (tokens == null) {
        if (kDebugMode) {
          AppLogger.error("Failed to exchange code for tokens");
        }
        return null;
      }

      final accessToken = tokens['access_token'] as String?;
      final idToken = tokens['id_token'] as String?;

      if (kDebugMode) {
        AppLogger.debug("Access Token received: ${accessToken != null ? 'Yes (${accessToken.length} chars)' : 'No'}");
        AppLogger.debug("ID Token received: ${idToken != null ? 'Yes (${idToken.length} chars)' : 'No'}");
      }

      if (accessToken == null) {
        if (kDebugMode) {
          AppLogger.error("No access token in response");
        }
        return null;
      }

      if (idToken == null) {
        if (kDebugMode) {
          AppLogger.error("No ID token in response - Firebase Auth requires ID token");
        }
        return null;
      }

      if (kDebugMode) {
        AppLogger.info("Google OAuth tokens obtained successfully!");
      }

      // ตรวจสอบ ID Token expiration ก่อนสร้าง credential
      if (kDebugMode) {
        AppLogger.debug("Checking ID Token expiration...");
      }

      final idTokenPayload = decodeIdToken(idToken);
      if (idTokenPayload.isNotEmpty) {
        final exp = idTokenPayload['exp'] as int?;
        final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

        if (kDebugMode) {
          AppLogger.debug("ID Token expiration time: ${exp != null ? DateTime.fromMillisecondsSinceEpoch(exp * 1000) : 'N/A'}");
          AppLogger.debug("Current time: ${DateTime.fromMillisecondsSinceEpoch(now * 1000)}");
        }

        if (exp != null && exp < now) {
          if (kDebugMode) {
            AppLogger.error("❌ ID Token has expired! exp: $exp, now: $now");
            AppLogger.error("❌ Token expired at: ${DateTime.fromMillisecondsSinceEpoch(exp * 1000)}");
          }
          return null;
        }

        if (kDebugMode) {
          AppLogger.debug("✅ ID Token is valid and not expired");
          AppLogger.debug("📋 ID Token payload: $idTokenPayload");
        }
      } else {
        if (kDebugMode) {
          AppLogger.warning("⚠️ Could not decode ID Token payload, proceeding anyway...");
        }
      }

      // สร้าง credential สำหรับ Firebase Auth
      if (kDebugMode) {
        AppLogger.debug("Creating GoogleAuthProvider credential...");
        AppLogger.debug("📝 Access Token (first 50 chars): ${accessToken.substring(0, accessToken.length > 50 ? 50 : accessToken.length)}");
        AppLogger.debug("📝 ID Token (first 50 chars): ${idToken.substring(0, idToken.length > 50 ? 50 : idToken.length)}");
      }

      final AuthCredential authCredential = GoogleAuthProvider.credential(accessToken: accessToken, idToken: idToken);

      // ล็อกอินเข้าสู่ Firebase
      if (kDebugMode) {
        AppLogger.debug("Signing in to Firebase with credential...");
      }

      final UserCredential authResult = await _auth.signInWithCredential(authCredential);

      if (kDebugMode) {
        AppLogger.info("Firebase Auth Success!");
        AppLogger.info("  User: ${authResult.user?.displayName}");
        AppLogger.info("  Email: ${authResult.user?.email}");
        AppLogger.info("  UID: ${authResult.user?.uid}");
      }

      return authResult;
    } catch (e, stackTrace) {
      if (kDebugMode) {
        AppLogger.error("Error during Google Sign In: $e");
        AppLogger.error("Error type: ${e.runtimeType}");
        AppLogger.error("Stack trace: $stackTrace");

        // ตรวจสอบว่าเป็น Firebase Auth error หรือไม่
        if (e is FirebaseAuthException) {
          AppLogger.error("Firebase Auth Error Code: ${e.code}");
          AppLogger.error("Firebase Auth Error Message: ${e.message}");
          AppLogger.error("Firebase Auth Error Details: ${e.code} - ${e.message}");

          // แสดงคำแนะนำเฉพาะสำหรับ error ที่พบบ่อย
          switch (e.code) {
            case 'invalid-credential':
              AppLogger.error("🔧 Fix suggestions for 'invalid-credential':");
              AppLogger.error("   1. Check if ID token has expired");
              AppLogger.error("   2. Verify Firebase project configuration");
              AppLogger.error("   3. Ensure Google Sign-In is enabled in Firebase Console");
              AppLogger.error("   4. Check Client ID matches Firebase configuration");
              break;
            case 'invalid-id-token':
              AppLogger.error("🔧 Fix suggestions for 'invalid-id-token':");
              AppLogger.error("   1. ID token is malformed or tampered");
              AppLogger.error("   2. Token may have been corrupted during transmission");
              break;
            case 'tenant-id-mismatch':
              AppLogger.error("🔧 Fix suggestions for 'tenant-id-mismatch':");
              AppLogger.error("   1. Check if you're using Firebase Auth with multi-tenancy");
              AppLogger.error("   2. Ensure tenant ID is correct");
              break;
            default:
              AppLogger.error("🔧 General Firebase Auth error occurred");
          }
        }
      }
      return null;
    }
  }

  /// เริ่ม local HTTP server และรอรับ authorization code
  Future<String?> _startServerAndGetAuthCode(Uri authUrl) async {
    try {
      // ปิด server เก่าถ้ามี
      await _server?.close(force: true);

      // เริ่ม server ใหม่
      _server = await HttpServer.bind(InternetAddress.loopbackIPv4, _redirectPort);

      if (kDebugMode) {
        AppLogger.debug("Local server started on port $_redirectPort");
      }

      // เปิด browser ไปหน้า Google login
      if (await canLaunchUrl(authUrl)) {
        await launchUrl(authUrl, mode: LaunchMode.externalApplication);
      } else {
        if (kDebugMode) {
          AppLogger.error("Cannot launch URL: $authUrl");
        }
        await _server?.close(force: true);
        return null;
      }

      // รอรับ request จาก Google callback (timeout 2 นาที)
      final request = await _server!.first.timeout(
        const Duration(minutes: 2),
        onTimeout: () {
          throw TimeoutException('OAuth callback timeout');
        },
      );

      // ดึง authorization code จาก query parameters
      final code = request.uri.queryParameters['code'];
      final error = request.uri.queryParameters['error'];

      // ส่ง response กลับไปที่ browser
      if (code != null) {
        request.response
          ..statusCode = HttpStatus.ok
          ..headers.contentType = ContentType.html
          ..write(_getSuccessHtml());
      } else {
        request.response
          ..statusCode = HttpStatus.badRequest
          ..headers.contentType = ContentType.html
          ..write(_getErrorHtml(error ?? 'Unknown error'));
      }
      await request.response.close();

      // ปิด server
      await _server?.close(force: true);
      _server = null;

      if (error != null) {
        if (kDebugMode) {
          AppLogger.error("OAuth error: $error");
        }
        return null;
      }

      return code;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error("Server error: $e");
      }
      await _server?.close(force: true);
      _server = null;
      return null;
    }
  }

  /// แลก authorization code เป็น access token และ id token
  Future<Map<String, dynamic>?> _exchangeCodeForTokens(String code, String clientId, String clientSecret) async {
    try {
      if (kDebugMode) {
        AppLogger.debug("Starting token exchange...");
        AppLogger.debug("Token endpoint: $_tokenEndpoint");
        AppLogger.debug("Redirect URI: $_redirectUri");
      }

      final client = HttpClient();
      final request = await client.postUrl(Uri.parse(_tokenEndpoint));
      request.headers.contentType = ContentType('application', 'x-www-form-urlencoded');

      final body = Uri(queryParameters: {'code': code, 'client_id': clientId, 'client_secret': clientSecret, 'redirect_uri': _redirectUri, 'grant_type': 'authorization_code'}).query;

      if (kDebugMode) {
        AppLogger.debug("Sending token exchange request...");
      }

      request.write(body);
      final response = await request.close();
      final responseBody = await response.transform(utf8.decoder).join();

      if (kDebugMode) {
        AppLogger.debug("Token exchange response status: ${response.statusCode}");
      }

      if (response.statusCode == 200) {
        if (kDebugMode) {
          AppLogger.debug("Token exchange successful!");
        }
        final tokens = json.decode(responseBody) as Map<String, dynamic>;
        if (kDebugMode) {
          AppLogger.debug("Tokens received: access_token=${tokens.containsKey('access_token')}, id_token=${tokens.containsKey('id_token')}");
        }
        return tokens;
      } else {
        if (kDebugMode) {
          AppLogger.error("Token exchange failed with status ${response.statusCode}");
          AppLogger.error("Response body: $responseBody");
        }
        return null;
      }
    } catch (e, stackTrace) {
      if (kDebugMode) {
        AppLogger.error("Token exchange error: $e");
        AppLogger.error("Stack trace: $stackTrace");
      }
      return null;
    }
  }

  /// HTML แสดงเมื่อ login สำเร็จ
  String _getSuccessHtml() {
    return '''
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Login Success</title>
  <style>
    body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #f0f8ff; }
    h1 { color: #4CAF50; }
    p { color: #666; }
  </style>
</head>
<body>
  <h1>✓ เข้าสู่ระบบสำเร็จ!</h1>
  <p>คุณสามารถปิดหน้าต่างนี้และกลับไปที่แอพได้</p>
  <script>setTimeout(() => window.close(), 3000);</script>
</body>
</html>
''';
  }

  /// HTML แสดงเมื่อ login ผิดพลาด
  String _getErrorHtml(String error) {
    return '''
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Login Failed</title>
  <style>
    body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #fff0f0; }
    h1 { color: #f44336; }
    p { color: #666; }
  </style>
</head>
<body>
  <h1>✗ เข้าสู่ระบบไม่สำเร็จ</h1>
  <p>Error: $error</p>
  <p>กรุณาปิดหน้าต่างนี้และลองใหม่อีกครั้ง</p>
</body>
</html>
''';
  }

  /// ดึงข้อมูลผู้ใช้จาก JWT token
  Map<String, dynamic> decodeIdToken(String? idToken) {
    if (idToken == null) {
      if (kDebugMode) {
        AppLogger.error("decodeIdToken: idToken is null");
      }
      return {};
    }

    try {
      // แยกส่วนข้อมูลจาก JWT token
      final parts = idToken.split('.');
      if (kDebugMode) {
        AppLogger.debug("JWT token parts count: ${parts.length}");
      }

      if (parts.length > 1) {
        final payload = parts[1];
        final normalized = base64Url.normalize(payload);
        final decoded = utf8.decode(base64Url.decode(normalized));
        final data = json.decode(decoded) as Map<String, dynamic>;

        if (kDebugMode) {
          AppLogger.debug("JWT token payload decoded successfully");
          AppLogger.debug("Issuer (iss): ${data['iss']}");
          AppLogger.debug("Audience (aud): ${data['aud']}");
          AppLogger.debug("Subject (sub): ${data['sub']}");
          AppLogger.debug("Email: ${data['email']}");
          AppLogger.debug("Email verified: ${data['email_verified']}");
        }

        return data;
      } else {
        if (kDebugMode) {
          AppLogger.error("Invalid JWT token format - expected at least 2 parts, got ${parts.length}");
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error("Error decoding ID token: $e");
        AppLogger.error("ID token length: ${idToken.length}");
        AppLogger.error("ID token first 100 chars: ${idToken.substring(0, idToken.length > 100 ? 100 : idToken.length)}");
      }
    }

    return {};
  }
}
