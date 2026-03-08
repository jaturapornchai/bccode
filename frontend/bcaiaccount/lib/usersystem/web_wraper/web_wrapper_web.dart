import 'package:flutter/foundation.dart';
// Copyright 2013 The Flutter Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import 'package:flutter/material.dart';
import 'package:google_sign_in_web/web_only.dart' as gsi;
import 'package:google_sign_in/google_sign_in.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Renders the Google Sign-In button for web
Widget renderButton() {
  // Ensure GoogleSignIn is initialized before rendering button
  try {
    // Attempt to initialize if not already done
    GoogleSignIn.instance.authenticationEvents;

    return gsi.renderButton(
      configuration: gsi.GSIButtonConfiguration(
        type: gsi.GSIButtonType.standard,
        theme: gsi.GSIButtonTheme.outline,
        size: gsi.GSIButtonSize.large,
        text: gsi.GSIButtonText.signin,
        shape: gsi.GSIButtonShape.rectangular,
        logoAlignment: gsi.GSIButtonLogoAlignment.left,
      ),
    );
  } catch (e) {
    // If there's an error, return a custom button
    if (kDebugMode) {
      AppLogger.error('Error initializing Google Sign-In: $e');
    }
    return SizedBox(
      width: double.infinity,
      height: 50,
      child: ElevatedButton(
        style: ElevatedButton.styleFrom(
          foregroundColor: Colors.black,
          backgroundColor: Colors.white,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(10),
          ),
        ),
        onPressed: () async {
          try {
            await GoogleSignIn.instance.authenticate();
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error('Google Sign-In Error: $e');
            }
          }
        },
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.login, color: Colors.blue),
            SizedBox(width: 10),
            Text('Sign in with Google', style: TextStyle(fontSize: 16)),
          ],
        ),
      ),
    );
  }
}
