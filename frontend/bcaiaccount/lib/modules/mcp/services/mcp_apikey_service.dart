import 'dart:convert';
import 'package:http/http.dart' as http;
import '../../../global.dart' as global;
import '../../../utils/logger/app_logger.dart';
import '../models/mcp_apikey_model.dart';

/// MCP API Key Service
/// Handles all API interactions for MCP API key management
class MCPAPIKeyService {
  static final MCPAPIKeyService _instance = MCPAPIKeyService._internal();
  factory MCPAPIKeyService() => _instance;
  MCPAPIKeyService._internal();

  /// สร้าง URL สำหรับ GoAPI endpoint
  String _url(String endpoint) => global.goApiUrlPath(endpoint);

  /// Headers พร้อม JWT token สำหรับทุก request
  Map<String, String> get _authHeaders {
    final token = global.appConfig.getString('token') ?? '';
    return {
      'Content-Type': 'application/json',
      if (token.isNotEmpty) 'Authorization': 'Bearer $token',
    };
  }

  /// Get all API keys for a shop
  Future<List<MCPAPIKeyModel>> getAPIKeys(String shopId) async {
    try {
      final url = _url('api/mcp/keys?shop_id=$shopId');
      AppLogger.debug('[MCP API Keys] GET $url');
      final response = await http.get(
        Uri.parse(url),
        headers: _authHeaders,
      );
      AppLogger.debug('[MCP API Keys] Status: ${response.statusCode}');
      if (response.statusCode != 200) {
        AppLogger.error('[MCP API Keys] Body: ${response.body}');
      }

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        if (data['success'] == true && data['data'] != null) {
          final List<dynamic> keys = data['data'];
          return keys.map((json) => MCPAPIKeyModel.fromJson(json)).toList();
        }
        return [];
      } else {
        throw Exception('Failed to load API keys: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error loading API keys: $e');
    }
  }

  /// Get a single API key by ID
  Future<MCPAPIKeyModel> getAPIKey(String keyId) async {
    try {
      final response = await http.get(
        Uri.parse(_url('api/mcp/keys/$keyId')),
        headers: _authHeaders,
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        if (data['success'] == true && data['data'] != null) {
          return MCPAPIKeyModel.fromJson(data['data']);
        }
        throw Exception('API key not found');
      } else {
        throw Exception('Failed to load API key: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error loading API key: $e');
    }
  }

  /// Create a new API key
  Future<MCPAPIKeyModel> createAPIKey(CreateAPIKeyRequest request) async {
    try {
      final response = await http.post(
        Uri.parse(_url('api/mcp/keys')),
        headers: _authHeaders,
        body: json.encode(request.toJson()),
      );

      if (response.statusCode == 201) {
        final data = json.decode(response.body);
        return MCPAPIKeyModel.fromJson(data);
      } else {
        final error = json.decode(response.body);
        throw Exception(error['error'] ?? 'Failed to create API key');
      }
    } catch (e) {
      throw Exception('Error creating API key: $e');
    }
  }

  /// Update an existing API key
  Future<void> updateAPIKey(String keyId, UpdateAPIKeyRequest request) async {
    try {
      final response = await http.put(
        Uri.parse(_url('api/mcp/keys/$keyId')),
        headers: _authHeaders,
        body: json.encode(request.toJson()),
      );

      if (response.statusCode != 200) {
        final error = json.decode(response.body);
        throw Exception(error['error'] ?? 'Failed to update API key');
      }
    } catch (e) {
      throw Exception('Error updating API key: $e');
    }
  }

  /// Delete an API key (soft delete)
  Future<void> deleteAPIKey(String keyId) async {
    try {
      final response = await http.delete(
        Uri.parse(_url('api/mcp/keys/$keyId')),
        headers: _authHeaders,
      );

      if (response.statusCode != 200) {
        final error = json.decode(response.body);
        throw Exception(error['error'] ?? 'Failed to delete API key');
      }
    } catch (e) {
      throw Exception('Error deleting API key: $e');
    }
  }

  /// Get audit logs for a shop
  Future<List<MCPAuditLogModel>> getAuditLogs(
    String shopId, {
    int limit = 50,
    int skip = 0,
  }) async {
    try {
      final response = await http.get(
        Uri.parse(_url('api/mcp/audit-logs?shop_id=$shopId&limit=$limit&skip=$skip')),
        headers: _authHeaders,
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        if (data['success'] == true && data['data'] != null) {
          final List<dynamic> logs = data['data'];
          return logs.map((json) => MCPAuditLogModel.fromJson(json)).toList();
        }
        return [];
      } else {
        throw Exception('Failed to load audit logs: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error loading audit logs: $e');
    }
  }

  /// Export API key config สำหรับ Claude Desktop/Code
  Future<Map<String, dynamic>> exportAPIKey(String keyId) async {
    try {
      final response = await http.get(
        Uri.parse(_url('api/mcp/keys/$keyId/export')),
        headers: _authHeaders,
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        if (data['success'] == true && data['data'] != null) {
          return data['data'] as Map<String, dynamic>;
        }
        throw Exception('Export failed: no data');
      } else {
        final error = json.decode(response.body);
        throw Exception(error['error'] ?? 'Failed to export API key');
      }
    } catch (e) {
      throw Exception('Error exporting API key: $e');
    }
  }

  /// สร้าง API key พร้อม export config (คืน full key + config กลับมา)
  Future<Map<String, dynamic>> createAPIKeyWithExport(CreateAPIKeyRequest request) async {
    try {
      final response = await http.post(
        Uri.parse(_url('api/mcp/keys/create-with-export')),
        headers: _authHeaders,
        body: json.encode(request.toJson()),
      );

      if (response.statusCode == 201) {
        final data = json.decode(response.body);
        if (data['success'] == true && data['data'] != null) {
          return data['data'] as Map<String, dynamic>;
        }
        throw Exception('Create failed: no data');
      } else {
        final error = json.decode(response.body);
        throw Exception(error['error'] ?? 'Failed to create API key with export');
      }
    } catch (e) {
      throw Exception('Error creating API key: $e');
    }
  }

  /// Test an API key by calling the MCP health endpoint
  Future<bool> testAPIKey(String apiKey) async {
    try {
      final response = await http.get(
        Uri.parse(_url('mcp/health')),
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': apiKey,
        },
      );

      return response.statusCode == 200;
    } catch (e) {
      return false;
    }
  }

  /// Copy API key to clipboard helper
  String getFullAPIKey(String maskedKey) {
    // This is just a helper - the full key is only shown once during creation
    return maskedKey;
  }
}
