import 'package:flutter/foundation.dart';
import 'dart:convert';
import 'dart:io';
import 'dart:async';
import 'package:url_launcher/url_launcher.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// OAuth2 helper for Windows Desktop Google Sign-In
/// Simply opens browser to Google login page - no token handling
class WindowsDesktopOAuth2Helper {
  /// Open Google authentication in browser (Simple approach)
  /// Just opens Google sign-in page without OAuth flow
  static Future<String?> signInWithGoogle(BuildContext context) async {
    try {
      // Simple Google sign-in URL (no OAuth parameters)
      final googleSignInUrl = Uri.parse('https://accounts.google.com/signin');

      if (kDebugMode) {        AppLogger.debug('🔐 Opening Google Sign-In page in browser...');
              AppLogger.debug('🌐 URL: $googleSignInUrl');
      }

      try {
        // Try external application mode first
        if (await canLaunchUrl(googleSignInUrl)) {
          await launchUrl(googleSignInUrl,
              mode: LaunchMode.externalApplication);
          if (kDebugMode) {
            AppLogger.info('✅ Google Sign-In page opened successfully');
          }
        } else {
          // Fallback to platform default
          await launchUrl(googleSignInUrl, mode: LaunchMode.platformDefault);
          if (kDebugMode) {
            AppLogger.debug('✅ Browser opened with default mode');
          }
        }

        // Return success
        return 'google_signin_opened';
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('❌ Failed to open browser: $e');
        }
        return null;
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error opening Google Sign-In: $e');
      }
      return null;
    }
  }

  /// Get user profile using access token
  static Future<Map<String, dynamic>?> getUserProfile(
      String accessToken) async {
    try {
      final client = HttpClient();
      final request = await client.getUrl(
        Uri.parse('https://www.googleapis.com/oauth2/v2/userinfo'),
      );
      request.headers.add('Authorization', 'Bearer $accessToken');

      final response = await request.close();

      if (response.statusCode == 200) {
        final responseBody = await response.transform(utf8.decoder).join();
        final userInfo = json.decode(responseBody);
        if (kDebugMode) {
          AppLogger.debug('✅ User profile retrieved: ${userInfo['email']}');
        }
        return userInfo;
      } else {
        if (kDebugMode) {
          AppLogger.error('❌ Failed to get user profile: ${response.statusCode}');
        }
        return null;
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Profile error: $e');
      }
      return null;
    }
  }
}
