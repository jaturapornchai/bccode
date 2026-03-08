import 'dart:convert';
import 'package:http/http.dart' as http;
import '../../../global.dart' as global;
import '../../../utils/logger/app_logger.dart';

class ThaiWordTokenizer {
  static Future<List<String>> tokenize(String text) async {
    if (text.trim().isEmpty) return [];

    try {
      final baseUrl = global.goApiUrlPath('tokenize');
      final url = Uri.parse(baseUrl);

      AppLogger.debug('🔍 Tokenize API Request URL: $url');
      AppLogger.debug('📤 Request text: "$text"');

      final response = await http
          .post(
            url,
            headers: {'Content-Type': 'application/json; charset=utf-8'},
            body: jsonEncode({'text': text.trim()}),
          )
          .timeout(const Duration(seconds: 5));

      AppLogger.debug('📥 Tokenize Response Status: ${response.statusCode}');
      AppLogger.debug(
        '📥 Tokenize Response Body (raw bytes): ${response.bodyBytes.length} bytes',
      );

      if (response.statusCode == 200) {
        // ใช้ utf8.decode เพื่อรองรับภาษาไทย
        final jsonData = jsonDecode(utf8.decode(response.bodyBytes));
        AppLogger.debug('📦 Parsed data: $jsonData');

        if (jsonData['status'] == 'success' && jsonData['tokens'] != null) {
          final tokens = List<String>.from(jsonData['tokens']);
          AppLogger.debug('🎯 Raw tokens from API: $tokens');

          final filtered = tokens
              .where((t) => t.length >= 2 || RegExp(r'^\d+$').hasMatch(t))
              .toList();
          AppLogger.debug('✅ Filtered tokens: $filtered');
          return filtered.toSet().toList();
        }
      }

      AppLogger.warning('⚠️ API tokenize failed, using fallback');
      return _fallbackTokenize(text);
    } catch (e) {
      AppLogger.error('❌ Tokenize error: $e');
      return _fallbackTokenize(text);
    }
  }

  static List<String> _fallbackTokenize(String text) {
    return text
        .trim()
        .toLowerCase()
        .split(RegExp(r'[\s,;.:()\[\]{}]+'))
        .where((t) => t.isNotEmpty)
        .toSet()
        .toList();
  }

  static String buildFullTextSearchQuery({
    required List<String> keywords,
    required List<String> searchFields,
    String operator = 'OR',
  }) {
    if (keywords.isEmpty) return '1=1';
    final conditions = keywords.map((kw) {
      final fields = searchFields
          .map((f) => "lower($f) LIKE lower('%$kw%')")
          .join(' OR ');
      return '($fields)';
    }).toList();
    return conditions.join(' $operator ');
  }
}
