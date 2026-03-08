import 'package:smlaicloud/services/result_table_api_service.dart';
import 'package:smlaicloud/utils/report_builder/report_query_log_entry.dart';

/// Utility for constructing linked report queries plus their logging metadata.
class ReportResultBuilder {
  ReportResultBuilder({required this.headerAlias, this.headerSummaryConfig});

  final ResultTableQueryAlias headerAlias;
  final Map<String, dynamic>? headerSummaryConfig;
  LinkedQueryDefinition? _headerDefinition;
  final List<LinkedQueryDefinition> _childDefinitions = [];
  final List<ReportQueryLogEntry> _queryLogs = [];

  void setHeaderQuery(String query) {
    _headerDefinition = LinkedQueryDefinition(
      alias: headerAlias,
      query: query,
      summaryConfig: headerSummaryConfig,
    );
    _queryLogs.add(ReportQueryLogEntry(alias: headerAlias, query: query));
  }

  void addChildQuery({
    required ResultTableQueryAlias alias,
    required ResultTableQueryAlias parentAlias,
    required String query,
    required List<String> parentKeys,
    required List<String> childKeys,
  }) {
    final definition = LinkedQueryDefinition(
      alias: alias,
      query: query,
      linkConfig: LinkedQueryLinkConfig(
        parentAlias: parentAlias,
        parentKeys: parentKeys,
        childKeys: childKeys,
      ),
    );
    _childDefinitions.add(definition);
    _queryLogs.add(ReportQueryLogEntry(alias: alias, query: query));
  }

  List<ReportQueryLogEntry> get queryLogs => List.unmodifiable(_queryLogs);

  List<LinkedQueryDefinition> buildDefinitions() {
    final header = _headerDefinition;
    if (header == null) {
      throw StateError('Header query must be set before building definitions');
    }
    return [header, ..._childDefinitions];
  }

  /// Builds an aggregate SQL statement that wraps the header query and
  /// applies the provided aggregate fields. Used for document-level totals.
  String buildAggregateQuery({required List<String> aggregateFields}) {
    final header = _headerDefinition;
    if (header == null) {
      throw StateError(
        'Header query must be set before building aggregate query',
      );
    }
    if (aggregateFields.isEmpty) {
      throw ArgumentError('aggregateFields must not be empty');
    }

    final selectClause = aggregateFields.join(',\n  ');
    final cteName = '${headerAlias.value}_aggregate_source';

    return '''
WITH $cteName AS (
${header.query}
)
SELECT
  $selectClause
FROM $cteName;
''';
  }
}
