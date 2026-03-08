import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/services/result_table_api_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// Model สำหรับ DocRef record
class DocRefModel {
  final int id;
  final String docno;
  final int docnotransflag;
  final String docnoref;
  final int docnoreftransflag;

  DocRefModel({required this.id, required this.docno, required this.docnotransflag, required this.docnoref, required this.docnoreftransflag});

  factory DocRefModel.fromJson(Map<String, dynamic> json) {
    return DocRefModel(
      id: json['id'] ?? 0,
      docno: json['docno'] ?? '',
      docnotransflag: json['docnotransflag'] ?? 0,
      docnoref: json['docnoref'] ?? '',
      docnoreftransflag: json['docnoreftransflag'] ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {'id': id, 'docno': docno, 'docnotransflag': docnotransflag, 'docnoref': docnoref, 'docnoreftransflag': docnoreftransflag};
  }
}

/// Response model สำหรับ DocFlow
class DocFlowResponse {
  final List<DocRefModel> referencedBy; // เอกสารที่อ้างมาหาเอกสารนี้ (ใครอ้างเรา)
  final List<DocRefModel> referenceTo; // เอกสารที่เอกสารนี้อ้างไป (เราอ้างใคร)

  DocFlowResponse({List<DocRefModel>? referencedBy, List<DocRefModel>? referenceTo}) : referencedBy = referencedBy ?? [], referenceTo = referenceTo ?? [];
}

class DocRefRepository {
  /// ดึงข้อมูล DocFlow แบบ recursive - ไล่หา chain ทั้งหมดจนสุด
  /// [docno] - เลขที่เอกสาร
  /// [transflag] - ประเภทเอกสาร
  Future<DocFlowResponse> getDocFlow({required String docno, required int transflag}) async {
    AppLogger.info('📄 [DocRef] getDocFlow: docno=$docno, transflag=$transflag');
    try {
      // Set เก็บ docno ที่เคย query แล้ว เพื่อป้องกัน infinite loop
      final Set<String> visitedReferencedBy = {};
      final Set<String> visitedReferenceTo = {};

      // ไล่หา chain ทั้งสองทาง
      final referencedByDocs = await _getReferencedByRecursive(docno, visitedReferencedBy);
      final referenceToDocs = await _getReferenceToRecursive(docno, visitedReferenceTo);

      AppLogger.info('📄 [DocRef] Result: referencedBy=${referencedByDocs.length}, referenceTo=${referenceToDocs.length}');

      return DocFlowResponse(referencedBy: referencedByDocs, referenceTo: referenceToDocs);
    } catch (ex) {
      AppLogger.error('❌ [DocRef] getDocFlow Error: $ex');
      rethrow;
    }
  }

  /// Recursive: หาเอกสารที่อ้างมาหาเอกสารนี้ (ใครอ้างเรา) และไล่ต่อไปจนสุด
  /// ไล่ขึ้น: docnoref → docno → docnoref → docno ...
  Future<List<DocRefModel>> _getReferencedByRecursive(String docno, Set<String> visited) async {
    if (visited.contains(docno)) return []; // ป้องกัน infinite loop
    visited.add(docno);

    List<DocRefModel> results = [];

    try {
      // หาเอกสารที่ docnoref = docno ปัจจุบัน (ใครอ้างมาหาเรา)
      final query =
          '''
        SELECT id, docno, docnotransflag, docnoref, docnoreftransflag 
        FROM docref 
        WHERE docnoref = '$docno'
        ORDER BY id
      ''';

      AppLogger.debug('📄 [DocRef] ReferencedBy Query: WHERE docnoref = $docno');

      final response = await ResultTableApiService.executeSelectQueries(database: global.getShopId(), queries: [query]);

      AppLogger.debug('📄 [DocRef] ReferencedBy Full Response: $response');

      final data = response['results']?[0] ?? {};
      final rows = data['rows'] as List? ?? [];

      AppLogger.debug('📄 [DocRef] ReferencedBy Response: success - count: ${rows.length}');
      if (rows.isNotEmpty) {
        AppLogger.debug('📄 [DocRef] ReferencedBy Data: $rows');
      } else {
        AppLogger.debug('📄 [DocRef] ReferencedBy: No data found or data is null');
      }

      if (rows.isNotEmpty) {
        final docs = rows.map((item) => DocRefModel.fromJson(item as Map<String, dynamic>)).toList();
        results.addAll(docs);

        // Log รายละเอียดเอกสารที่ได้
        for (final doc in docs) {
          AppLogger.debug('📄 [DocRef] → Found ReferencedBy: docno=${doc.docno} (flag:${doc.docnotransflag}) refs to docnoref=${doc.docnoref} (flag:${doc.docnoreftransflag})');
        }

        // Recursive: ไล่ต่อจาก docno ที่ได้ (เอกสารที่อ้างมาหาเรา อาจมีคนอ้างต่ออีก)
        for (final doc in docs) {
          if (doc.docno.isNotEmpty && !visited.contains(doc.docno)) {
            AppLogger.debug('📄 [DocRef] → Recursive check: ${doc.docno}');
            final childDocs = await _getReferencedByRecursive(doc.docno, visited);
            results.addAll(childDocs);
          }
        }
      }
    } catch (e, stackTrace) {
      // ถ้า query ล้มเหลว ให้ skip
      AppLogger.error('❌ [DocRef] ReferencedBy Error: $e');
      AppLogger.debug('📄 [DocRef] ReferencedBy StackTrace: $stackTrace');
    }

    return results;
  }

  /// Recursive: หาเอกสารที่เอกสารนี้อ้างไป (เราอ้างใคร) และไล่ต่อไปจนสุด
  /// ไล่ลง: docno → docnoref → docno → docnoref ...
  Future<List<DocRefModel>> _getReferenceToRecursive(String docno, Set<String> visited) async {
    if (visited.contains(docno)) return []; // ป้องกัน infinite loop
    visited.add(docno);

    List<DocRefModel> results = [];

    try {
      // หาเอกสารที่ docno ปัจจุบัน อ้างไป (เราอ้างใคร)
      final query =
          '''
        SELECT id, docno, docnotransflag, docnoref, docnoreftransflag 
        FROM docref 
        WHERE docno = '$docno'
        ORDER BY id
      ''';

      AppLogger.debug('📄 [DocRef] ReferenceTo Query: WHERE docno = $docno');

      final response = await ResultTableApiService.executeSelectQueries(database: global.getShopId(), queries: [query]);

      AppLogger.debug('📄 [DocRef] ReferenceTo Full Response: $response');

      final data = response['results']?[0] ?? {};
      final rows = data['rows'] as List? ?? [];

      AppLogger.debug('📄 [DocRef] ReferenceTo Response: success - count: ${rows.length}');
      if (rows.isNotEmpty) {
        AppLogger.debug('📄 [DocRef] ReferenceTo Data: $rows');
      } else {
        AppLogger.debug('📄 [DocRef] ReferenceTo: No data found or data is null');
      }

      if (rows.isNotEmpty) {
        final docs = rows.map((item) => DocRefModel.fromJson(item as Map<String, dynamic>)).toList();
        results.addAll(docs);

        // Log รายละเอียดเอกสารที่ได้
        for (final doc in docs) {
          AppLogger.debug('📄 [DocRef] → Found ReferenceTo: docno=${doc.docno} (flag:${doc.docnotransflag}) refs to docnoref=${doc.docnoref} (flag:${doc.docnoreftransflag})');
        }

        // Recursive: ไล่ต่อจาก docnoref ที่ได้ (เอกสารที่เราอ้างไป อาจอ้างต่ออีก)
        for (final doc in docs) {
          if (doc.docnoref.isNotEmpty && !visited.contains(doc.docnoref)) {
            AppLogger.debug('📄 [DocRef] → Recursive check: ${doc.docnoref}');
            final childDocs = await _getReferenceToRecursive(doc.docnoref, visited);
            results.addAll(childDocs);
          }
        }
      }
    } catch (e, stackTrace) {
      // ถ้า query ล้มเหลว ให้ skip
      AppLogger.error('❌ [DocRef] ReferenceTo Error: $e');
      AppLogger.debug('📄 [DocRef] ReferenceTo StackTrace: $stackTrace');
    }

    return results;
  }

  /// ดึงชื่อประเภทเอกสารจาก transflag
  static String getTransflagName(int transflag) {
    switch (transflag) {
      // Purchase
      case 1:
        return global.language('transaction_purchase');
      case 2:
        return 'ซื้อเชื่อ';
      case 3:
        return global.language('transaction_purchase_order');
      case 4:
        return global.language('purchase_request');
      case 5:
        return global.language('transaction_stock_receive_product');
      case 6:
        return global.language('partial_receive');
      case 7:
        return 'ส่งคืนซื้อ';
      // Sale
      case 11:
        return global.language('transaction_sale');
      case 12:
        return 'ขายเชื่อ';
      case 13:
        return global.language('transaction_sale_order');
      case 14:
        return global.language('transaction_quotation');
      case 17:
        return 'รับคืนขาย';
      // Stock
      case 41:
        return global.language('transaction_stock_receive_product');
      case 42:
        return global.language('transaction_stock_pick_up_product');
      case 43:
        return global.language('transaction_stock_transfer');
      case 44:
        return global.language('delivery_note');
      case 48:
        return global.language('transaction_stock_transfer');
      // Adjust
      case 66:
        return global.language('stock_movement_adjust_increase');
      case 68:
        return global.language('stock_movement_adjust_decrease');
      case 866:
        return 'ปรับมูลค่าเพิ่ม';
      case 868:
        return 'ปรับมูลค่าลด';
      // Payment
      case 71:
        return global.language('pay_advance');
      case 72:
        return 'คืนจ่ายล่วงหน้า';
      case 73:
        return 'มัดจำ';
      case 74:
        return 'คืนมัดจำ';
      case 81:
        return global.language('receive_advance');
      case 82:
        return 'คืนรับล่วงหน้า';
      case 83:
        return global.language('receive_deposit');
      case 84:
        return 'คืนรับมัดจำ';
      // Purchase Partial (PP)
      case 310:
        return global.language('partial_receive');
      // No type
      case 0:
        return '';
      default:
        return 'เอกสาร ($transflag)';
    }
  }
}
