import 'dart:io';
import 'dart:typed_data';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import '../../global.dart' as global;

Future<void> savePdf(
    Uint8List pdfData, String fileName, BuildContext context) async {
  final result = await FilePicker.platform.saveFile(
    dialogTitle: global.language('pdf.save_dialog_title'),
    fileName: fileName,
    type: FileType.custom,
    allowedExtensions: ['pdf'],
  );
  if (result != null) {
    final file = File(result);
    await file.writeAsBytes(pdfData);
    global.showInfoSnackBar(context, global.language('pdf.save_success'));
  }
}
