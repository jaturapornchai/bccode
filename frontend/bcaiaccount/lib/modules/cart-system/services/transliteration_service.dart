import 'dart:convert';
import 'package:http/http.dart' as http;
import '../../../utils/logger/app_logger.dart';

class TransliterationService {
  final String baseUrl;

  TransliterationService({required this.baseUrl});

  /// แปลงภาษาไทยเป็นอังกฤษ
  Future<TransliterationResult> transliterateThaiToEnglish(String text) async {
    return _transliterate('/transliterate/th-en', text);
  }

  /// แปลงภาษาอังกฤษเป็นไทย
  Future<TransliterationResult> transliterateEnglishToThai(String text) async {
    return _transliterate('/transliterate/en-th', text);
  }

  Future<TransliterationResult> _transliterate(
    String endpoint,
    String text,
  ) async {
    try {
      final url = Uri.parse('$baseUrl$endpoint');

      AppLogger.debug('🌐 [Transliteration] URL: $url');
      AppLogger.debug('📤 [Transliteration] Input: $text');

      final response = await http
          .post(
            url,
            headers: {'Content-Type': 'application/json; charset=utf-8'},
            body: utf8.encode(json.encode({'text': text})),
          )
          .timeout(
            const Duration(seconds: 10),
            onTimeout: () {
              throw Exception('Request timeout');
            },
          );

      AppLogger.debug('📥 [Transliteration] Status: ${response.statusCode}');

      if (response.statusCode == 200) {
        // Decode with UTF-8 support
        final jsonResponse = json.decode(utf8.decode(response.bodyBytes));
        AppLogger.debug('📥 [Transliteration] Response: $jsonResponse');

        final result = TransliterationResult.fromJson(jsonResponse);
        AppLogger.info(
          'Translation: "${result.input}" → "${result.output}" (${result.matched} matches)',
        );

        return result;
      } else {
        throw Exception('Failed: ${response.statusCode}');
      }
    } catch (e) {
      AppLogger.error('Exception in transliteration: $e');
      // Return original text if translation fails
      return TransliterationResult(
        status: 'error',
        input: text,
        output: text,
        matched: 0,
      );
    }
  }
}

/// Model for API response
class TransliterationResult {
  final String status;
  final String input;
  final String output;
  final int matched;

  TransliterationResult({
    required this.status,
    required this.input,
    required this.output,
    required this.matched,
  });

  factory TransliterationResult.fromJson(Map<String, dynamic> json) {
    return TransliterationResult(
      status: json['status'] as String,
      input: json['input'] as String,
      output: json['output'] as String,
      matched: json['matched'] as int,
    );
  }

  bool get isSuccess => status == 'success';
  bool get hasTranslation => matched > 0;
}
