// ignore_for_file: unused_local_variable, avoid_web_libraries_in_flutter

import 'dart:convert';
import 'dart:html' as html;

import 'package:flutter/foundation.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

void downloadFile(String url, String fileName) {
  final html.AnchorElement anchor = html.AnchorElement(href: url)
    ..setAttribute("download", fileName)
    ..setAttribute("target", "_blank")
    ..click();
}

Future<bool> downloadFileBytes(dynamic fileData, String fileName) async {
  final blob = html.Blob([fileData.takeBytes()]);
  final url = html.Url.createObjectUrlFromBlob(blob);

  final anchor = html.AnchorElement(href: url)
    ..setAttribute("download", fileName)
    ..click();

  html.Url.revokeObjectUrl(url);
  return true;
}

Future<bool> downloadAssetFileBytes(
    Uint8List fileData, String saveFileName) async {
  try {
    // Load the asset file
    final String base64Data = base64Encode(fileData);
    final String url = 'data:application/octet-stream;base64,$base64Data';

    // Create an anchor element and trigger download
    final html.AnchorElement anchor = html.AnchorElement(href: url)
      ..setAttribute("download", saveFileName)
      ..click();

    return true; // Indicate success
  } catch (e) {
    if (kDebugMode) {
      AppLogger.error("An error occurred while downloading the file: $e");
    }
    return false; // Indicate failure
  }
}
