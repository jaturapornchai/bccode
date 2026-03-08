/// MCP API Key Model
/// Represents an API key for MCP (Model Context Protocol) access
library;
import '../../../global.dart' as global;
class MCPAPIKeyModel {
  final String id;
  final String apiKey;
  final String shopId;
  final String name;
  final String? description;
  final bool isActive;
  final List<String> allowedTools;
  final int rateLimitPerMinute;
  final DateTime createdAt;
  final DateTime? expiresAt;
  final DateTime? lastUsedAt;
  final String? createdBy;

  MCPAPIKeyModel({
    required this.id,
    required this.apiKey,
    required this.shopId,
    required this.name,
    this.description,
    required this.isActive,
    required this.allowedTools,
    required this.rateLimitPerMinute,
    required this.createdAt,
    this.expiresAt,
    this.lastUsedAt,
    this.createdBy,
  });

  factory MCPAPIKeyModel.fromJson(Map<String, dynamic> json) {
    return MCPAPIKeyModel(
      id: json['id'] ?? '',
      apiKey: json['api_key'] ?? '',
      shopId: json['shop_id'] ?? '',
      name: json['name'] ?? '',
      description: json['description'],
      isActive: json['is_active'] ?? true,
      allowedTools: List<String>.from(json['allowed_tools'] ?? []),
      rateLimitPerMinute: json['rate_limit_per_minute'] ?? 60,
      createdAt: DateTime.parse(json['created_at'] ?? DateTime.now().toIso8601String()),
      expiresAt: json['expires_at'] != null ? DateTime.parse(json['expires_at']) : null,
      lastUsedAt: json['last_used_at'] != null ? DateTime.parse(json['last_used_at']) : null,
      createdBy: json['created_by'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'api_key': apiKey,
      'shop_id': shopId,
      'name': name,
      'description': description,
      'is_active': isActive,
      'allowed_tools': allowedTools,
      'rate_limit_per_minute': rateLimitPerMinute,
      'created_at': createdAt.toIso8601String(),
      'expires_at': expiresAt?.toIso8601String(),
      'last_used_at': lastUsedAt?.toIso8601String(),
      'created_by': createdBy,
    };
  }

  /// Returns a masked version of the API key for display
  String get maskedApiKey {
    if (apiKey.length > 12) {
      return '${apiKey.substring(0, 8)}****${apiKey.substring(apiKey.length - 4)}';
    }
    return '****';
  }

  /// Returns true if the API key has expired
  bool get isExpired {
    if (expiresAt == null) return false;
    return DateTime.now().isAfter(expiresAt!);
  }

  /// Returns the status text
  String get statusText {
    if (!isActive) return 'Inactive';
    if (isExpired) return 'Expired';
    return 'Active';
  }

  MCPAPIKeyModel copyWith({
    String? id,
    String? apiKey,
    String? shopId,
    String? name,
    String? description,
    bool? isActive,
    List<String>? allowedTools,
    int? rateLimitPerMinute,
    DateTime? createdAt,
    DateTime? expiresAt,
    DateTime? lastUsedAt,
    String? createdBy,
  }) {
    return MCPAPIKeyModel(
      id: id ?? this.id,
      apiKey: apiKey ?? this.apiKey,
      shopId: shopId ?? this.shopId,
      name: name ?? this.name,
      description: description ?? this.description,
      isActive: isActive ?? this.isActive,
      allowedTools: allowedTools ?? this.allowedTools,
      rateLimitPerMinute: rateLimitPerMinute ?? this.rateLimitPerMinute,
      createdAt: createdAt ?? this.createdAt,
      expiresAt: expiresAt ?? this.expiresAt,
      lastUsedAt: lastUsedAt ?? this.lastUsedAt,
      createdBy: createdBy ?? this.createdBy,
    );
  }
}

/// Request model for creating a new API key
class CreateAPIKeyRequest {
  final String shopId;
  final String name;
  final String? description;
  final List<String>? allowedTools;
  final int? rateLimitPerMinute;
  final String? expiresAt;
  final String? createdBy;

  CreateAPIKeyRequest({
    required this.shopId,
    required this.name,
    this.description,
    this.allowedTools,
    this.rateLimitPerMinute,
    this.expiresAt,
    this.createdBy,
  });

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{
      'shop_id': shopId,
      'name': name,
    };
    if (description != null) json['description'] = description;
    if (allowedTools != null) json['allowed_tools'] = allowedTools;
    if (rateLimitPerMinute != null) json['rate_limit_per_minute'] = rateLimitPerMinute;
    if (expiresAt != null) json['expires_at'] = expiresAt;
    if (createdBy != null) json['created_by'] = createdBy;
    return json;
  }
}

/// Request model for updating an API key
class UpdateAPIKeyRequest {
  final String? name;
  final String? description;
  final bool? isActive;
  final List<String>? allowedTools;
  final int? rateLimitPerMinute;

  UpdateAPIKeyRequest({
    this.name,
    this.description,
    this.isActive,
    this.allowedTools,
    this.rateLimitPerMinute,
  });

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
    if (name != null) json['name'] = name;
    if (description != null) json['description'] = description;
    if (isActive != null) json['is_active'] = isActive;
    if (allowedTools != null) json['allowed_tools'] = allowedTools;
    if (rateLimitPerMinute != null) json['rate_limit_per_minute'] = rateLimitPerMinute;
    return json;
  }
}

/// Available MCP Tools — ต้องตรงกับ backend AvailableTools ใน server.go
class MCPTools {
  // ===== Sales & Products =====
  static const String searchProducts = 'search_products';
  static const String getDailySales = 'get_daily_sales';
  static const String getSalesByDateRange = 'get_sales_by_date_range';
  static const String getTopSellingProducts = 'get_top_selling_products';
  static const String getSalesBySeller = 'get_sales_by_seller';
  static const String getMonthlySummary = 'get_monthly_summary';

  // ===== Dashboard & KPIs =====
  static const String getDashboardKpis = 'get_dashboard_kpis';
  static const String getBusinessHealth = 'get_business_health';

  // ===== Financial =====
  static const String getProfitAnalysis = 'get_profit_analysis';
  static const String getAccountsReceivable = 'get_accounts_receivable';
  static const String getAccountsPayable = 'get_accounts_payable';
  static const String getCashFlow = 'get_cash_flow';

  // ===== Inventory =====
  static const String getInventoryValue = 'get_inventory_value';
  static const String getLowStockAlerts = 'get_low_stock_alerts';
  static const String getDeadStock = 'get_dead_stock';
  static const String getInventoryTurnover = 'get_inventory_turnover';

  // ===== Customers =====
  static const String getTopCustomers = 'get_top_customers';
  static const String getCustomerGrowth = 'get_customer_growth';
  static const String getCustomerSegments = 'get_customer_segments';

  // ===== Comparison =====
  static const String getYoyComparison = 'get_yoy_comparison';
  static const String getMomComparison = 'get_mom_comparison';

  // ===== Unit of Measure (read) =====
  static const String listUnits = 'list_units';
  static const String searchUnits = 'search_units';

  // ===== Unit of Measure (write) =====
  static const String createUnit = 'create_unit';
  static const String createUnits = 'create_units';
  static const String updateUnit = 'update_unit';
  static const String deleteUnit = 'delete_unit';
  static const String deleteUnits = 'delete_units';

  // ===== API Development =====
  static const String listApiEndpoints = 'list_api_endpoints';
  static const String getApiSpec = 'get_api_spec';
  static const String getApiExample = 'get_api_example';
  static const String listEnums = 'list_enums';
  static const String getModelSchema = 'get_model_schema';

  // ===== Database — PostgreSQL =====
  static const String getDatabaseSchema = 'get_database_schema';
  static const String executeQuery = 'execute_query';
  static const String getTableSample = 'get_table_sample';

  // ===== Database — MongoDB =====
  static const String queryMongodb = 'query_mongodb';
  static const String listMongodbCollections = 'list_mongodb_collections';
  static const String aggregateMongodb = 'aggregate_mongodb';

  // ===== Database — ClickHouse =====
  static const String queryClickhouse = 'query_clickhouse';
  static const String listClickhouseTables = 'list_clickhouse_tables';

  /// Write tools — ต้องมี Developer permission ถึงจะใช้ได้
  static const Set<String> writeTools = {
    createUnit, createUnits, updateUnit, deleteUnit, deleteUnits,
  };

  /// ตรวจว่า tool เป็น write tool หรือไม่
  static bool isWriteTool(String tool) => writeTools.contains(tool);

  /// Readonly tools ทั้งหมด (ไม่รวม write)
  static List<String> get readonlyTools =>
      allTools.where((t) => !writeTools.contains(t)).toList();

  static const List<String> allTools = [
    // Sales & Products
    searchProducts, getDailySales, getSalesByDateRange, getTopSellingProducts, getSalesBySeller, getMonthlySummary,
    // Dashboard & KPIs
    getDashboardKpis, getBusinessHealth,
    // Financial
    getProfitAnalysis, getAccountsReceivable, getAccountsPayable, getCashFlow,
    // Inventory
    getInventoryValue, getLowStockAlerts, getDeadStock, getInventoryTurnover,
    // Customers
    getTopCustomers, getCustomerGrowth, getCustomerSegments,
    // Comparison
    getYoyComparison, getMomComparison,
    // Unit of Measure
    listUnits, searchUnits, createUnit, createUnits, updateUnit, deleteUnit, deleteUnits,
    // API Development
    listApiEndpoints, getApiSpec, getApiExample, listEnums, getModelSchema,
    // Database — PostgreSQL
    getDatabaseSchema, executeQuery, getTableSample,
    // Database — MongoDB
    queryMongodb, listMongodbCollections, aggregateMongodb,
    // Database — ClickHouse
    queryClickhouse, listClickhouseTables,
  ];

  /// จัดกลุ่ม tools เป็น category สำหรับ UI
  static const Map<String, List<String>> toolCategories = {
    'Sales & Products': [searchProducts, getDailySales, getSalesByDateRange, getTopSellingProducts, getSalesBySeller, getMonthlySummary],
    'Dashboard & KPIs': [getDashboardKpis, getBusinessHealth],
    'Financial': [getProfitAnalysis, getAccountsReceivable, getAccountsPayable, getCashFlow],
    'Inventory': [getInventoryValue, getLowStockAlerts, getDeadStock, getInventoryTurnover],
    'Customers': [getTopCustomers, getCustomerGrowth, getCustomerSegments],
    'Comparison': [getYoyComparison, getMomComparison],
    'Unit of Measure': [listUnits, searchUnits, createUnit, createUnits, updateUnit, deleteUnit, deleteUnits],
    'API Development': [listApiEndpoints, getApiSpec, getApiExample, listEnums, getModelSchema],
    'PostgreSQL': [getDatabaseSchema, executeQuery, getTableSample],
    'MongoDB': [queryMongodb, listMongodbCollections, aggregateMongodb],
    'ClickHouse': [queryClickhouse, listClickhouseTables],
  };

  static String getToolName(String tool) {
    return _toolInfo[tool]?['name'] ?? tool;
  }

  static String getToolDescription(String tool) {
    return _toolInfo[tool]?['desc'] ?? '';
  }

  /// แปลง allowed_tools → label สำหรับแสดงผล
  static String getPermissionLabel(List<String> allowedTools) {
    if (allowedTools.contains('*') || allowedTools.contains('readwrite')) {
      return 'All tools (${allTools.length} tools)';
    }
    if (allowedTools.length == 1 && allowedTools.first == 'readonly') {
      return 'Readonly (${readonlyTools.length} tools)';
    }
    if (allowedTools.contains('readonly')) {
      final extraWriteCount = allowedTools.where((t) => t != 'readonly').length;
      return 'Readonly + $extraWriteCount write';
    }
    return '${allowedTools.length} tools';
  }

  /// แปลง allowed_tools → preset type
  static String getPresetType(List<String> allowedTools) {
    if (allowedTools.contains('*') || allowedTools.contains('readwrite')) return 'developer';
    if (allowedTools.length == 1 && allowedTools.first == 'readonly') return 'readonly';
    return 'custom';
  }

  static final Map<String, Map<String, String>> _toolInfo = {
    searchProducts: {'name': 'Search Products', 'desc': 'ค้นหาสินค้าตามชื่อหรือบาร์โค้ด'},
    getDailySales: {'name': 'Daily Sales', 'desc': 'สรุปยอดขายรายวัน'},
    getSalesByDateRange: {'name': 'Sales by Date Range', 'desc': 'ยอดขายตามช่วงวันที่ แบ่งตามวัน/สัปดาห์/เดือน'},
    getTopSellingProducts: {'name': 'Top Selling Products', 'desc': 'สินค้าขายดีตามช่วงเวลา'},
    getSalesBySeller: {'name': 'Sales by Seller', 'desc': 'ยอดขายแยกตามพนักงาน'},
    getMonthlySummary: {'name': 'Monthly Summary', 'desc': 'สรุปยอดขายรายเดือน เทียบเดือนก่อน'},
    getDashboardKpis: {'name': 'Dashboard KPIs', 'desc': 'ตัวชี้วัดหลัก: ยอดขาย กำไร จำนวนบิล'},
    getBusinessHealth: {'name': 'Business Health', 'desc': 'ภาพรวมสุขภาพธุรกิจ'},
    getProfitAnalysis: {'name': 'Profit Analysis', 'desc': 'วิเคราะห์กำไรตามช่วงเวลา'},
    getAccountsReceivable: {'name': 'Accounts Receivable', 'desc': global.language('accounts_receivable')},
    getAccountsPayable: {'name': 'Accounts Payable', 'desc': global.language('accounts_payable')},
    getCashFlow: {'name': 'Cash Flow', 'desc': global.language('cash_flow')},
    getInventoryValue: {'name': 'Inventory Value', 'desc': global.language('inventory_value')},
    getLowStockAlerts: {'name': 'Low Stock Alerts', 'desc': 'สินค้าที่ stock ต่ำ'},
    getDeadStock: {'name': 'Dead Stock', 'desc': 'สินค้าค้างสต็อก ไม่เคลื่อนไหว'},
    getInventoryTurnover: {'name': 'Inventory Turnover', 'desc': global.language('inventory_turnover')},
    getTopCustomers: {'name': 'Top Customers', 'desc': 'ลูกค้ายอดซื้อสูงสุด'},
    getCustomerGrowth: {'name': 'Customer Growth', 'desc': 'อัตราการเติบโตของฐานลูกค้า'},
    getCustomerSegments: {'name': 'Customer Segments', 'desc': 'แบ่งกลุ่มลูกค้าตามพฤติกรรม'},
    getYoyComparison: {'name': 'Year-over-Year', 'desc': 'เปรียบเทียบ ปีต่อปี'},
    getMomComparison: {'name': 'Month-over-Month', 'desc': 'เปรียบเทียบ เดือนต่อเดือน'},
    // Unit of Measure
    listUnits: {'name': 'List Units', 'desc': 'แสดงรายการหน่วยนับทั้งหมด'},
    searchUnits: {'name': 'Search Units', 'desc': 'ค้นหาหน่วยนับ'},
    createUnit: {'name': 'Create Unit', 'desc': 'สร้างหน่วยนับใหม่ (write)'},
    createUnits: {'name': 'Create Units (Bulk)', 'desc': 'สร้างหน่วยนับหลายรายการ (write)'},
    updateUnit: {'name': 'Update Unit', 'desc': 'แก้ไขหน่วยนับ (write)'},
    deleteUnit: {'name': 'Delete Unit', 'desc': 'ลบหน่วยนับ (write)'},
    deleteUnits: {'name': 'Delete Units (Bulk)', 'desc': 'ลบหน่วยนับหลายรายการ (write)'},
    // API Development
    listApiEndpoints: {'name': 'API Catalog', 'desc': 'ดูรายการ API endpoints ทั้งหมด'},
    getApiSpec: {'name': 'API Spec', 'desc': 'ดู request/response spec ของ API'},
    getApiExample: {'name': 'API Example', 'desc': 'ดูตัวอย่าง curl + code snippets'},
    listEnums: {'name': 'Enum Catalog', 'desc': 'ดู enum values ทั้งหมดของ backend'},
    getModelSchema: {'name': 'Model Schema', 'desc': 'ดู Go struct + Dart/TS model'},
    // Database
    getDatabaseSchema: {'name': 'Database Schema (PG)', 'desc': 'ดูโครงสร้าง tables ใน PostgreSQL'},
    executeQuery: {'name': 'Execute Query (PG)', 'desc': 'รัน SELECT SQL query บน PostgreSQL'},
    getTableSample: {'name': 'Table Sample (PG)', 'desc': 'ดูข้อมูลตัวอย่างจาก table'},
    queryMongodb: {'name': 'Query MongoDB', 'desc': 'ค้นหาข้อมูลใน MongoDB ด้วย JSON filter'},
    listMongodbCollections: {'name': 'List Collections (Mongo)', 'desc': 'แสดงรายการ collections ใน MongoDB'},
    aggregateMongodb: {'name': 'Aggregate MongoDB', 'desc': 'รัน aggregation pipeline บน MongoDB'},
    queryClickhouse: {'name': 'Query ClickHouse', 'desc': 'รัน SELECT query บน ClickHouse (OLAP)'},
    listClickhouseTables: {'name': 'List Tables (CH)', 'desc': 'แสดงรายการ tables ใน ClickHouse'},
  };
}

/// Audit Log Model
class MCPAuditLogModel {
  final String id;
  final String apiKeyId;
  final String shopId;
  final String toolName;
  final Map<String, dynamic>? requestParams;
  final String responseStatus;
  final String? errorMessage;
  final int executionTimeMs;
  final DateTime createdAt;

  MCPAuditLogModel({
    required this.id,
    required this.apiKeyId,
    required this.shopId,
    required this.toolName,
    this.requestParams,
    required this.responseStatus,
    this.errorMessage,
    required this.executionTimeMs,
    required this.createdAt,
  });

  factory MCPAuditLogModel.fromJson(Map<String, dynamic> json) {
    return MCPAuditLogModel(
      id: json['id'] ?? '',
      apiKeyId: json['api_key_id'] ?? '',
      shopId: json['shop_id'] ?? '',
      toolName: json['tool_name'] ?? '',
      requestParams: json['request_params'],
      responseStatus: json['response_status'] ?? '',
      errorMessage: json['error_message'],
      executionTimeMs: json['execution_time_ms'] ?? 0,
      createdAt: DateTime.parse(json['created_at'] ?? DateTime.now().toIso8601String()),
    );
  }

  bool get isSuccess => responseStatus == 'success';
}
