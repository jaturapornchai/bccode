import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

class Util {
  static bool isLandscape(BuildContext context) {
    return MediaQuery.of(context).orientation == Orientation.landscape;
  }
}

class ReconnectingOverlay extends StatelessWidget {
  const ReconnectingOverlay({super.key});

  @override
  Widget build(BuildContext context) => Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: const [CircularProgressIndicator()],
    ),
  );
}

Future<Map<String, dynamic>> clickhouseSelectGroup(List<String> querys) async {
  var httpClient = http.Client();
  // ใช้ goApiUrlPath() เพื่อสร้าง URL ที่ถูกต้อง (รองรับทุก platform)
  var urlString = global.goApiUrlPath("clickhouse/selectgroup");
  var url = Uri.parse(urlString);

  if (kDebugMode) {
    AppLogger.debug("querys : $querys");
  }

  try {
    var response = await httpClient.post(
      url,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({"querys": querys}),
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      return {
        "status": "error",
        "code": response.statusCode,
        "message":
            "${global.language('error.api_call_failed')}: ${response.reasonPhrase}",
      };
    }
  } catch (e) {
    return {
      "status": "error",
      "code": 500,
      "message": "${global.language('error.connection_failed')}: $e",
    };
  } finally {
    httpClient.close();
  }
}

Future<Map<String, dynamic>> pgSqlSelectGroup(List<String> querys) async {
  var httpClient = http.Client();
  // ใช้ goApiUrlPath() เพื่อสร้าง URL ที่ถูกต้อง (รองรับทุก platform)
  var urlString = global.goApiUrlPath("get");
  var url = Uri.parse(urlString);

  if (kDebugMode) {
    AppLogger.debug("querys : $querys");
  }

  try {
    var response = await httpClient.post(
      url,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({"database": global.getShopId(), "queries": querys}),
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      return {
        "status": "error",
        "code": response.statusCode,
        "message":
            "${global.language('error.api_call_failed')}: ${response.reasonPhrase}",
      };
    }
  } catch (e) {
    return {
      "status": "error",
      "code": 500,
      "message": "${global.language('error.connection_failed')}: $e",
    };
  } finally {
    httpClient.close();
  }
}
