class AppConfig {
  // BC AI Cloud API URLs

  // DEV - ใช้ localhost สำหรับ development
  static const String serviceDevApi = 'http://localhost:9090';
  static const String reportDevApi = 'https://api.dev.dedepos.com/apireport';
  static const String webSocketCartServiceDev = 'ws://143.198.192.64:13001/cart-service';
  static const String reportDevBiApi = 'https://api.dev.dedepos.com/dedebiapi/report';

  // PROD
  static const String servicePrdApi = 'https://smlaicloudapi.dedepos.com';
  static const String reportPrdApi = 'https://api.dedepos.com/apireport';
  static const String webSocketCartServicePrd = 'wss://api.dedepos.com/cart-service';
  static const String reportPrdBiApi = 'https://api.dedepos.com/dedebiapi/report';

  /// UAT
  static const String serviceUATApi = 'https://smlaicloudapi.dedepos.com';
  static const String reportUATApi = 'https://api.dedepos.com/apireport';
  static const String webSocketCartServiceUAT = 'wss://api.dedepos.com/cart-service';
  static const String reportUATBiApi = 'https://api.dedepos.com/dedebiapi/report';

  // Clickhouse & GoAPI URLs (GoAPI รวมเข้า MainAPI แล้ว — ใช้ path /goapi)
  // DEV
  static const String devServiceClickhouse = 'https://api.dev.dedepos.com/apireport/clickhouse';
  static const String devServiceApi = 'http://localhost:9090/goapi';

  // UAT
  static const String uatServiceClickhouse = 'https://api.dev.dedepos.com/apireport/clickhouse';
  static const String uatServiceApi = 'https://smlaicloudapi.dedepos.com/goapi';

  // PROD
  static const String prodServiceClickhouse = 'https://api.dev.dedepos.com/apireport/clickhouse';
  static const String prodServiceApi = 'https://smlaicloudapi.dedepos.com/goapi';
}
