import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:path_provider/path_provider.dart';
import 'dart:io';
import 'package:file_picker/file_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

Future<void> downloadFile(String url, String fileName) async {
  if (Platform.isMacOS) {
    if (kDebugMode) {
      AppLogger.debug('This is macOS');
    }
    String? folderPath = await selectDownloadDirectory();
    if (folderPath == null) return; // User didn't select a folder

    var response = await http.get(Uri.parse(url));
    if (response.statusCode == 200) {
      File file = File('$folderPath/$fileName');
      await file.writeAsBytes(response.bodyBytes);
      if (kDebugMode) {
        AppLogger.info('File downloaded to $folderPath/$fileName');
      }
    } else {
      if (kDebugMode) {
        AppLogger.debug(
            'Failed to download file: Server responded with status code ${response.statusCode}');
      }
    }
  } else {
    if (kDebugMode) {
      AppLogger.debug('This is Android');
    }
    var response = await http.get(Uri.parse(url));
    var documentDirectory = await getApplicationDocumentsDirectory();
    var filePathAndName = '${documentDirectory.path}/$fileName';
    File file = File(filePathAndName);
    file.writeAsBytesSync(response.bodyBytes);
  }
}

Future<String?> selectDownloadDirectory() async {
  String? selectedDirectory = await FilePicker.platform.getDirectoryPath();

  if (selectedDirectory == null) {
    // User canceled the picker
    if (kDebugMode) {
      AppLogger.debug("No directory selected");
    }
    return null;
  }
  if (kDebugMode) {
    AppLogger.debug("Selected directory: $selectedDirectory");
  }
  return selectedDirectory;
}

Future<bool> downloadFileBytes(dynamic fileData, String fileName) async {
  if (Platform.isMacOS) {
    if (kDebugMode) {
      AppLogger.debug('This is macOS');
    }
    String? selectedDirectory = await FilePicker.platform.getDirectoryPath();
    if (selectedDirectory == null) {
      if (kDebugMode) {
        AppLogger.debug("No directory selected.");
      }
      return false;
    }

    try {
      final file = File('$selectedDirectory/$fileName');
      // Convert BytesBuilder to List<int> using takeBytes()
      await file.writeAsBytes(fileData.takeBytes());
      if (kDebugMode) {
        AppLogger.debug('File saved to: $selectedDirectory/$fileName');
      }
      return true;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Failed to save file: $e');
      }
      return false;
    }
  }
  return true;
}

Future<bool> downloadAssetFileBytes(Uint8List fileData, String fileName) async {
  if (Platform.isMacOS) {
    // macOS logic
    String? selectedDirectory = await FilePicker.platform.getDirectoryPath();
    if (selectedDirectory == null) {
      if (kDebugMode) {
        AppLogger.debug("No directory selected.");
      }
      return false;
    }

    try {
      final file = File('$selectedDirectory/$fileName');
      await file.writeAsBytes(fileData);
      if (kDebugMode) {
        AppLogger.debug('File saved to: $selectedDirectory/$fileName');
      }
      return true;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Failed to save file: $e');
      }
      return false;
    }
  }
  // Add else if for other platforms if necessary
  return true;
}
