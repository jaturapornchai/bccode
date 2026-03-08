import 'package:smlaicloud/repositories/client.dart';
import 'package:dio/dio.dart';
import 'dart:typed_data';
import 'package:smlaicloud/screens/report/file_download.dart';
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ExportCsvRepository {
  /// export product barcode

  Future<bool> exportProductBarcode(String languageCode) async {
    Dio client = Client().init();

    try {
      final response = await client.get(
        '/product/barcode/export?lang=$languageCode',
        options: Options(responseType: ResponseType.stream),
      );

      String contentDisposition =
          response.headers.value('content-disposition') ?? '';
      RegExp regex = RegExp(r'filename="([^"]*)"');
      String fileName =
          regex.firstMatch(contentDisposition)?.group(1) ??
          'product-barcode.csv';

      final bytesBuilder = await response.data.stream.fold<BytesBuilder>(
        BytesBuilder(),
        (BytesBuilder previous, List<int> element) => previous..add(element),
      );

      bool result = await downloadFileBytes(bytesBuilder, fileName);

      if (result) {
        return true;
      } else {
        return false;
      }
    } on DioException catch (e) {
      if (kDebugMode) {
        AppLogger.error('Dio exception: ${e.message}');
      }
      return false; // Error in download
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('General exception: $e');
      }
      return false; // Error in download
    }
  }

  /// export sale invoice
  Future<bool> exportSaleInvoice(String languageCode) async {
    Dio client = Client().init();

    try {
      final response = await client.get(
        '/transaction/sale-invoice/export?lang=$languageCode',
        options: Options(responseType: ResponseType.stream),
      );

      String contentDisposition =
          response.headers.value('content-disposition') ?? '';
      RegExp regex = RegExp(r'filename="([^"]*)"');
      String fileName =
          regex.firstMatch(contentDisposition)?.group(1) ?? 'sale-invoice.csv';

      final bytesBuilder = await response.data.stream.fold<BytesBuilder>(
        BytesBuilder(),
        (BytesBuilder previous, List<int> element) => previous..add(element),
      );

      bool result = await downloadFileBytes(bytesBuilder, fileName);

      if (result) {
        return true;
      } else {
        return false;
      }
    } on DioException catch (e) {
      if (kDebugMode) {
        AppLogger.error('Dio exception: ${e.message}');
      }
      return false; // Error in download
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('General exception: $e');
      }
      return false; // Error in download
    }
  }
}
