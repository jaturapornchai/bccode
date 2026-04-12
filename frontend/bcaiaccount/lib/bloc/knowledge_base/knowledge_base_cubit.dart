import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'knowledge_base_state.dart';

class KnowledgeBaseCubit extends Cubit<KnowledgeBaseState> {
  KnowledgeBaseCubit() : super(const KnowledgeBaseState());

  // KB endpoints live on goapi (RAGFlow-backed) at /api/v1/kb/*
  String get apiBaseUrl => global.goApiUrlPath('api/v1/kb');

  // Load documents
  Future<void> loadDocuments({String? branchId}) async {
    emit(state.copyWith(isLoading: true));

    try {
      final shopId = global.getShopId();
      final timestamp = DateTime.now().millisecondsSinceEpoch;

      final requestBody = {
        'shop_id': shopId,
        if (branchId != null) 'branch_id': branchId,
        'limit': 100,
        'skip': 0,
        'timestamp': timestamp,
      };

      if (kDebugMode) {
        AppLogger.debug('═══════════════════════════════════════');
        AppLogger.debug('📄 Loading Documents');
        AppLogger.debug('Branch ID: $branchId');
        AppLogger.debug('Request: ${jsonEncode(requestBody)}');
        AppLogger.debug('═══════════════════════════════════════');
      }

      final response = await http.post(
        Uri.parse('$apiBaseUrl/list?t=$timestamp'),
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 'no-cache, no-store, must-revalidate',
          'Pragma': 'no-cache',
          'Expires': '0',
        },
        body: jsonEncode(requestBody),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);

        if (kDebugMode) {
          AppLogger.debug('Response status: ${response.statusCode}');
          AppLogger.debug(
            'Documents count: ${(data['documents'] as List?)?.length ?? 0}',
          );
        }

        if (data['success'] == true && data['documents'] != null) {
          final documents = (data['documents'] as List)
              .map((doc) => DocumentModel.fromJson(doc))
              .toList();

          if (kDebugMode) {
            AppLogger.info('✅ Loaded ${documents.length} documents successfully');
            AppLogger.debug('═══════════════════════════════════════');
          }

          emit(state.copyWith(documents: documents, isLoading: false));
        } else {
          if (kDebugMode) {
            AppLogger.info('⚠️ No documents found or success=false');
            AppLogger.debug('═══════════════════════════════════════');
          }
          emit(state.copyWith(isLoading: false));
        }
      } else {
        emit(
          state.copyWith(
            isLoading: false,
            errorMessage: 'Failed to load documents',
          ),
        );
      }
    } catch (e) {
      emit(state.copyWith(isLoading: false, errorMessage: e.toString()));
    }
  }

  // Update document status in local state immediately
  void updateDocumentStatus(String filename, bool newStatus) {
    final updatedDocuments = state.documents.map((doc) {
      if (doc.filename == filename) {
        return DocumentModel(
          shopId: doc.shopId,
          branchId: doc.branchId,
          filename: doc.filename,
          contentType: doc.contentType,
          size: doc.size,
          uploadedAt: doc.uploadedAt,
          uploadedBy: doc.uploadedBy,
          description: doc.description,
          tags: doc.tags,
          version: doc.version,
          status: newStatus,
          allDay: doc.allDay,
          startDateTime: doc.startDateTime,
          endDateTime: doc.endDateTime,
        );
      }
      return doc;
    }).toList();

    emit(state.copyWith(documents: updatedDocuments));
  }

  // Update document all_day in local state immediately
  void updateDocumentAllDay(String filename, bool newAllDay) {
    final updatedDocuments = state.documents.map((doc) {
      if (doc.filename == filename) {
        return DocumentModel(
          shopId: doc.shopId,
          branchId: doc.branchId,
          filename: doc.filename,
          contentType: doc.contentType,
          size: doc.size,
          uploadedAt: doc.uploadedAt,
          uploadedBy: doc.uploadedBy,
          description: doc.description,
          tags: doc.tags,
          version: doc.version,
          status: doc.status,
          allDay: newAllDay,
          startDateTime: doc.startDateTime,
          endDateTime: doc.endDateTime,
        );
      }
      return doc;
    }).toList();

    emit(state.copyWith(documents: updatedDocuments));
  }

  // Update document date range in local state immediately
  void updateDocumentDateRange(
    String filename,
    String? startDateTime,
    String? endDateTime,
  ) {
    final updatedDocuments = state.documents.map((doc) {
      if (doc.filename == filename) {
        return DocumentModel(
          shopId: doc.shopId,
          branchId: doc.branchId,
          filename: doc.filename,
          contentType: doc.contentType,
          size: doc.size,
          uploadedAt: doc.uploadedAt,
          uploadedBy: doc.uploadedBy,
          description: doc.description,
          tags: doc.tags,
          version: doc.version,
          status: doc.status,
          allDay: doc.allDay,
          startDateTime: startDateTime,
          endDateTime: endDateTime,
        );
      }
      return doc;
    }).toList();

    emit(state.copyWith(documents: updatedDocuments));
  }

  // Remove document from local state immediately
  void removeDocument(String filename) {
    final updatedDocuments = state.documents
        .where((doc) => doc.filename != filename)
        .toList();

    emit(state.copyWith(documents: updatedDocuments));
  }
}

// DocumentModel class
class DocumentModel {
  final String shopId;
  final String branchId;
  final String filename;
  final String contentType;
  final int size;
  final String uploadedAt;
  final String uploadedBy;
  final String description;
  final List<String> tags;
  final int version;
  final bool status;
  final bool allDay;
  final String? startDateTime;
  final String? endDateTime;

  DocumentModel({
    required this.shopId,
    required this.branchId,
    required this.filename,
    required this.contentType,
    required this.size,
    required this.uploadedAt,
    required this.uploadedBy,
    required this.description,
    required this.tags,
    required this.version,
    required this.status,
    required this.allDay,
    this.startDateTime,
    this.endDateTime,
  });

  factory DocumentModel.fromJson(Map<String, dynamic> json) {
    return DocumentModel(
      shopId: json['shop_id'] ?? '',
      branchId: json['branch_id'] ?? '',
      filename: json['filename'] ?? '',
      contentType: json['content_type'] ?? '',
      size: json['size'] ?? 0,
      uploadedAt: json['uploaded_at'] ?? '',
      uploadedBy: json['uploaded_by'] ?? '',
      description: json['description'] ?? '',
      tags: (json['tags'] as List?)?.map((e) => e.toString()).toList() ?? [],
      version: json['version'] ?? 1,
      status: json['status'] ?? true,
      allDay: json['all_day'] ?? true,
      startDateTime: json['start_date_time'],
      endDateTime: json['end_date_time'],
    );
  }
}
