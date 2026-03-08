import 'package:flutter_test/flutter_test.dart';
import 'package:smlaicloud/global.dart' as global;

void main() {
  group('safeParseDatetime', () {
    test('parses valid ISO 8601 datetime', () {
      final result = global.safeParseDatetime('2026-01-07T06:00:00.000Z');
      expect(result.year, 2026);
      expect(result.month, 1);
      expect(result.day, 7);
      expect(result.isUtc, true);
    });

    test('parses date-only string', () {
      final result = global.safeParseDatetime('2026-01-07');
      expect(result.year, 2026);
      expect(result.month, 1);
      expect(result.day, 7);
    });

    test('returns DateTime.now() for null input', () {
      final before = DateTime.now();
      final result = global.safeParseDatetime(null);
      final after = DateTime.now();
      expect(result.isAfter(before.subtract(const Duration(seconds: 1))), true);
      expect(result.isBefore(after.add(const Duration(seconds: 1))), true);
    });

    test('returns DateTime.now() for empty string', () {
      final before = DateTime.now();
      final result = global.safeParseDatetime('');
      final after = DateTime.now();
      expect(result.isAfter(before.subtract(const Duration(seconds: 1))), true);
      expect(result.isBefore(after.add(const Duration(seconds: 1))), true);
    });

    test('returns DateTime.now() for invalid string', () {
      final before = DateTime.now();
      final result = global.safeParseDatetime('not-a-date');
      final after = DateTime.now();
      expect(result.isAfter(before.subtract(const Duration(seconds: 1))), true);
      expect(result.isBefore(after.add(const Duration(seconds: 1))), true);
    });
  });

  group('getLocalTimezoneInfo', () {
    test('returns map with required keys', () {
      final info = global.getLocalTimezoneInfo();
      expect(info.containsKey('docdatelocal'), true);
      expect(info.containsKey('doctimelocal'), true);
      expect(info.containsKey('timezone'), true);
    });

    test('docdatelocal has YYYYMMDD format (8 digits)', () {
      final info = global.getLocalTimezoneInfo();
      expect(info['docdatelocal']!.length, 8);
      expect(int.tryParse(info['docdatelocal']!), isNotNull);
    });

    test('doctimelocal has HH:mm:ss format', () {
      final info = global.getLocalTimezoneInfo();
      final parts = info['doctimelocal']!.split(':');
      expect(parts.length, 3);
      expect(int.tryParse(parts[0]), isNotNull);
      expect(int.tryParse(parts[1]), isNotNull);
      expect(int.tryParse(parts[2]), isNotNull);
    });

    test('timezone starts with UTC', () {
      final info = global.getLocalTimezoneInfo();
      expect(info['timezone']!.startsWith('UTC'), true);
    });

    test('accepts custom DateTime parameter', () {
      final customDate = DateTime(2026, 6, 15, 14, 30, 45);
      final info = global.getLocalTimezoneInfo(customDate);
      expect(info['docdatelocal'], '20260615');
      expect(info['doctimelocal'], '14:30:45');
    });
  });

  group('safeToDouble', () {
    test('converts int to double', () {
      expect(global.safeToDouble(42), 42.0);
    });

    test('returns double as-is', () {
      expect(global.safeToDouble(3.14), 3.14);
    });

    test('converts string to double', () {
      expect(global.safeToDouble('99.5'), 99.5);
    });

    test('returns default for null', () {
      expect(global.safeToDouble(null), 0.0);
    });

    test('returns custom default for invalid string', () {
      expect(global.safeToDouble('abc', -1.0), -1.0);
    });
  });

  group('safeToInt', () {
    test('returns int as-is', () {
      expect(global.safeToInt(42), 42);
    });

    test('converts double to int', () {
      expect(global.safeToInt(3.7), 3);
    });

    test('converts string to int', () {
      expect(global.safeToInt('99'), 99);
    });

    test('returns default for null', () {
      expect(global.safeToInt(null), 0);
    });

    test('returns custom default for invalid string', () {
      expect(global.safeToInt('abc', -1), -1);
    });
  });
}
