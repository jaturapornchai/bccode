# Flutter Usage Guide - Transliteration API

## Installation

Add HTTP package to `pubspec.yaml`:

```yaml
dependencies:
  http: ^1.1.0
```

Run:
```bash
flutter pub get
```

## API Service Class

Create `lib/services/transliteration_service.dart`:

```dart
import 'dart:convert';
import 'package:http/http.dart' as http;

class TransliterationService {
  final String baseUrl;
  
  TransliterationService({
    this.baseUrl = 'http://localhost:9090/goapi',
  });

  /// แปลงภาษาไทยเป็นอังกฤษ
  Future<TransliterationResult> transliterateThaiToEnglish(String text) async {
    return _transliterate('/transliterate/th-en', text);
  }

  /// แปลงภาษาอังกฤษเป็นไทย
  Future<TransliterationResult> transliterateEnglishToThai(String text) async {
    return _transliterate('/transliterate/en-th', text);
  }

  Future<TransliterationResult> _transliterate(String endpoint, String text) async {
    try {
      final url = Uri.parse('$baseUrl$endpoint');
      
      final response = await http.post(
        url,
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
        },
        body: utf8.encode(json.encode({'text': text})),
      ).timeout(
        const Duration(seconds: 10),
        onTimeout: () {
          throw Exception('Request timeout');
        },
      );

      if (response.statusCode == 200) {
        // Decode with UTF-8 support
        final jsonResponse = json.decode(utf8.decode(response.bodyBytes));
        return TransliterationResult.fromJson(jsonResponse);
      } else {
        throw Exception('Failed: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error: $e');
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
```

## Usage Example - Simple Widget

```dart
import 'package:flutter/material.dart';
import 'services/transliteration_service.dart';

class TransliterationDemo extends StatefulWidget {
  const TransliterationDemo({Key? key}) : super(key: key);

  @override
  State<TransliterationDemo> createState() => _TransliterationDemoState();
}

class _TransliterationDemoState extends State<TransliterationDemo> {
  final _service = TransliterationService(
    baseUrl: 'http://localhost:9090/goapi', // Change to your API URL
  );
  
  final _thaiController = TextEditingController();
  final _englishController = TextEditingController();
  
  String _result = '';
  bool _isLoading = false;

  @override
  void dispose() {
    _thaiController.dispose();
    _englishController.dispose();
    super.dispose();
  }

  Future<void> _translateThaiToEnglish() async {
    if (_thaiController.text.isEmpty) return;
    
    setState(() {
      _isLoading = true;
      _result = '';
    });

    try {
      final result = await _service.transliterateThaiToEnglish(
        _thaiController.text,
      );
      
      setState(() {
        _result = result.output;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _result = 'Error: $e';
        _isLoading = false;
      });
    }
  }

  Future<void> _translateEnglishToThai() async {
    if (_englishController.text.isEmpty) return;
    
    setState(() {
      _isLoading = true;
      _result = '';
    });

    try {
      final result = await _service.transliterateEnglishToThai(
        _englishController.text,
      );
      
      setState(() {
        _result = result.output;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _result = 'Error: $e';
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Transliteration Demo'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Thai to English Section
            const Text(
              'Thai → English',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _thaiController,
              decoration: const InputDecoration(
                hintText: 'ใส่ข้อความภาษาไทย',
                border: OutlineInputBorder(),
              ),
              maxLines: 3,
            ),
            const SizedBox(height: 8),
            ElevatedButton(
              onPressed: _isLoading ? null : _translateThaiToEnglish,
              child: const Text('แปลเป็นอังกฤษ'),
            ),
            
            const SizedBox(height: 24),
            const Divider(),
            const SizedBox(height: 24),
            
            // English to Thai Section
            const Text(
              'English → Thai',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _englishController,
              decoration: const InputDecoration(
                hintText: 'Enter English text',
                border: OutlineInputBorder(),
              ),
              maxLines: 3,
            ),
            const SizedBox(height: 8),
            ElevatedButton(
              onPressed: _isLoading ? null : _translateEnglishToThai,
              child: const Text('แปลเป็นไทย'),
            ),
            
            const SizedBox(height: 24),
            const Divider(),
            const SizedBox(height: 24),
            
            // Result Section
            if (_isLoading)
              const Center(child: CircularProgressIndicator())
            else if (_result.isNotEmpty) ...[
              const Text(
                'Result:',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.grey[100],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.grey[300]!),
                ),
                child: Text(
                  _result,
                  style: const TextStyle(fontSize: 16),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
```

## Quick Examples

### 1. Simple One-Line Call

```dart
// Thai to English
final service = TransliterationService();
final result = await service.transliterateThaiToEnglish(
  'เครื่องมือ มากิต้า และ บ๊อช'
);
print(result.output); // "เครื่องมือ makita และ bosch"
print(result.matched); // 2
```

### 2. With Error Handling

```dart
try {
  final result = await service.transliterateEnglishToThai(
    'I use Makita and Bosch tools'
  );
  
  if (result.isSuccess && result.hasTranslation) {
    print('Translated: ${result.output}');
    print('Words matched: ${result.matched}');
  }
} on Exception catch (e) {
  print('Error: $e');
}
```

### 3. Batch Processing

```dart
Future<List<String>> translateMultiple(List<String> texts) async {
  final service = TransliterationService();
  final results = <String>[];
  
  for (final text in texts) {
    try {
      final result = await service.transliterateThaiToEnglish(text);
      results.add(result.output);
    } catch (e) {
      results.add(text); // Keep original if failed
    }
  }
  
  return results;
}

// Usage
final thaiTexts = [
  'สว่าน มากิต้า',
  'เครื่องตัด บ๊อช',
  'เครื่องเจียร ดีวอลท์',
];

final translated = await translateMultiple(thaiTexts);
print(translated);
// ["drill bit makita", "เครื่องตัด bosch", "เครื่องเจียร dewalt"]
```

## Configuration for Production

For production use, configure the base URL:

```dart
// lib/config/api_config.dart
class ApiConfig {
  static const String baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:9090/goapi',
  );
}

// Usage
final service = TransliterationService(
  baseUrl: ApiConfig.baseUrl,
);
```

Run with custom URL:
```bash
flutter run --dart-define=API_BASE_URL=https://your-api-domain.com
```

## Common Issues

### Issue 1: Thai characters not displaying correctly

**Solution**: Make sure to use `utf8.decode(response.bodyBytes)` instead of `response.body`

```dart
final jsonResponse = json.decode(utf8.decode(response.bodyBytes));
```

### Issue 2: Connection timeout in emulator

**Solution**: Use `10.0.2.2` instead of `localhost` for Android emulator:

```dart
final service = TransliterationService(
  baseUrl: Platform.isAndroid 
    ? 'http://10.0.2.2:9090/goapi'  // Android emulator
    : 'http://localhost:9090/goapi', // iOS simulator / real device
);
```

### Issue 3: CORS error in web

**Solution**: Add CORS headers on the backend (already configured in the API)

## Example API Responses

### Success Response
```json
{
  "status": "success",
  "input": "เครื่องมือ มากิต้า และ บ๊อช ดีมาก",
  "output": "เครื่องมือ makita และ bosch ดีมาก",
  "matched": 2
}
```

### No Translation Response
```json
{
  "status": "success",
  "input": "สว่าน เจาะกำแพง",
  "output": "สว่าน เจาะกำแพง",
  "matched": 0
}
```

### Error Response
```json
{
  "status": "error",
  "message": "Text is required and cannot be empty"
}
```

## API Endpoints

- `POST /transliterate/th-en` - Thai → English
- `POST /transliterate/en-th` - English → Thai

## Integration with Product Search

```dart
// Example: Enhance product search with transliteration
Future<List<Product>> searchProducts(String keyword) async {
  final service = TransliterationService();
  
  // Search with original keyword
  final results1 = await searchInDatabase(keyword);
  
  // Try Thai → English translation
  final thToEn = await service.transliterateThaiToEnglish(keyword);
  final results2 = thToEn.hasTranslation 
    ? await searchInDatabase(thToEn.output)
    : <Product>[];
  
  // Try English → Thai translation
  final enToTh = await service.transliterateEnglishToThai(keyword);
  final results3 = enToTh.hasTranslation
    ? await searchInDatabase(enToTh.output)
    : <Product>[];
  
  // Merge unique results
  final merged = <String, Product>{};
  for (final product in [...results1, ...results2, ...results3]) {
    merged[product.id] = product;
  }
  
  return merged.values.toList();
}
```

## Support

For issues or questions, check the main API documentation.
