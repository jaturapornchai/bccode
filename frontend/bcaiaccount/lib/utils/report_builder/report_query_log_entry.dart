import 'package:smlaicloud/services/result_table_api_service.dart';

/// Represents a SQL statement tied to a specific query alias for logging.
class ReportQueryLogEntry {
  const ReportQueryLogEntry({required this.alias, required this.query});

  final ResultTableQueryAlias alias;
  final String query;

  String get aliasLabel => alias.value;
}
