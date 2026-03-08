import 'dart:html' as html;
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

void savePdf(Uint8List pdfData, String fileName, BuildContext context) {
  final blob = html.Blob([pdfData]);
  final url = html.Url.createObjectUrlFromBlob(blob);
  final anchor = html.AnchorElement(href: url)
    ..setAttribute('download', fileName)
    ..style.display = 'none';
  html.document.body?.children.add(anchor);
  anchor.click();
  html.document.body?.children.remove(anchor);
  html.Url.revokeObjectUrl(url);

  // เพิ่ม SnackBar ถ้าต้องการ
  global.showInfoSnackBar(context, global.language('file_saved_successfully'));
}
