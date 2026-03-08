import 'package:flutter/foundation.dart';
import 'package:smlaicloud/app_const.dart';

class Environment {
  factory Environment() {
    return _singleton;
  }

  Environment._internal();

  static final Environment _singleton = Environment._internal();

  static const String DEV = 'DEV';
  static const String PROD = 'PROD';
  static const String UAT = 'UAT';

  late BaseConfig config;
  late bool isDev;

  void initConfig(String environment) {
    config = _getConfig(environment);
    isDev = environment == DEV || environment == UAT;
  }

  BaseConfig _getConfig(String environment) {
    switch (environment) {
      case DEV:
        return DevConfig();
      case PROD:
        return ProdConfig();
      case UAT:
        return UATConfig();
      default:
        throw Exception('Invalid environment: $environment');
    }
  }

  // API URLs by environment
  static String get devServiceClickhouse => AppConfig.devServiceClickhouse;
  static String get devServiceApi => AppConfig.devServiceApi;

  static String get uatServiceClickhouse => AppConfig.uatServiceClickhouse;
  static String get uatServiceApi => AppConfig.uatServiceApi;

  static String get prodServiceClickhouse => AppConfig.prodServiceClickhouse;
  static String get prodServiceApi => AppConfig.prodServiceApi;

  // BC AI Cloud API URLs
  static String get serviceDevApi => AppConfig.serviceDevApi;
  static String get reportDevApi => AppConfig.reportDevApi;

  // Report BI APIs
  static String get reportDevBiApi => AppConfig.reportDevBiApi;
  static String get reportProdBiApi => AppConfig.reportPrdBiApi;
  static String get reportUATBiApi => AppConfig.reportUATBiApi;
}

abstract class BaseConfig {
  String get serviceApi;
  String get reportApi;
  String get webSocketCartService;
  String get serviceClickhouse;
  String get reportApiPath;
  String get reportApiPort;
  String get reportBiApi;
}

class DevConfig extends BaseConfig {
  @override
  // debug = localhost:9090, release/web = dev-api.bcaicloud.com (mainapi)
  String get serviceApi => kDebugMode
      ? AppConfig.serviceDevApi
      : 'https://dev-api.bcaicloud.com';

  @override
  // debug = external report API, release/web = ผ่าน Caddy proxy
  String get reportApi => kDebugMode
      ? AppConfig.reportDevApi
      : 'https://dev-api.bcaicloud.com/apireport';

  @override
  String get webSocketCartService => AppConfig.webSocketCartServiceDev;

  @override
  // debug = external clickhouse, release/web = ผ่าน Caddy → goapi
  String get serviceClickhouse => kDebugMode
      ? Environment.devServiceClickhouse
      : 'https://dev-api.bcaicloud.com/goapi/clickhouse';

  @override
  // debug = external goapi, release/web = ผ่าน Caddy → goapi
  String get reportApiPath => kDebugMode
      ? Environment.devServiceApi
      : 'https://dev-api.bcaicloud.com/goapi';

  @override
  String get reportApiPort => "";

  @override
  // debug = external BI API, release/web = ผ่าน Caddy proxy
  String get reportBiApi => kDebugMode
      ? AppConfig.reportDevBiApi
      : 'https://dev-api.bcaicloud.com/dedebiapi/report';
}

class ProdConfig extends BaseConfig {
  @override
  String get serviceApi => AppConfig.servicePrdApi;

  @override
  String get reportApi => AppConfig.reportPrdApi;

  @override
  String get webSocketCartService => AppConfig.webSocketCartServicePrd;

  @override
  String get serviceClickhouse => Environment.prodServiceClickhouse;

  @override
  String get reportApiPath => Environment.prodServiceApi;

  @override
  String get reportApiPort => "";

  @override
  String get reportBiApi => AppConfig.reportPrdBiApi;
}

class UATConfig extends BaseConfig {
  @override
  String get serviceApi => AppConfig.serviceUATApi;

  @override
  String get reportApi => AppConfig.reportUATApi;

  @override
  String get webSocketCartService => AppConfig.webSocketCartServiceUAT;

  @override
  String get serviceClickhouse => Environment.uatServiceClickhouse;

  @override
  String get reportApiPath => Environment.uatServiceApi;

  @override
  String get reportApiPort => "";

  @override
  String get reportBiApi => AppConfig.reportUATBiApi;
}
