import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/bi_report/bi_report_bloc.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/export_models.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:url_launcher/url_launcher.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ExportReportDialog extends StatefulWidget {
  final BiReportType reportType;
  final String jobId;

  const ExportReportDialog({
    super.key,
    required this.reportType,
    required this.jobId,
  });

  @override
  State<ExportReportDialog> createState() => _ExportReportDialogState();
}

class _ExportReportDialogState extends State<ExportReportDialog> with global.ThemeRefreshMixin {
  ExportFormat? _selectedFormat;

  @override
  Widget build(BuildContext context) {
    return BlocListener<BiReportBloc, BiReportState>(
      listener: (context, state) {
        if (state is BiReportExportReady) {
          _showDownloadDialog(context, state);
        } else if (state is BiReportExportSubmitFailure ||
            state is BiReportExportStatusFailure) {
          _showErrorMessage(context, state);
        }
      },
      child: Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        child: Container(
          width: 400,
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              _buildHeader(),
              const SizedBox(height: 24),
              _buildFormatSelection(),
              const SizedBox(height: 24),
              _buildExportStatus(),
              const SizedBox(height: 24),
              _buildActionButtons(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHeader() {
    return Row(
      children: [
        Icon(Icons.download, color: Colors.indigo.shade600, size: 28),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Export รายงาน',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: global.theme.textColor,
                ),
              ),
              Text(
                widget.reportType.displayName,
                style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
              ),
            ],
          ),
        ),
        IconButton(
          onPressed: () {
            context.read<BiReportBloc>().add(const ResetBiReportExportState());
            Navigator.of(context).pop();
          },
          icon: Icon(Icons.close),
          tooltip: global.language('export_report_close'),
        ),
      ],
    );
  }

  Widget _buildFormatSelection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          global.language('export_report_select_format'),
          style: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            color: global.theme.textColor,
          ),
        ),
        SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: _buildFormatOption(
                format: ExportFormat.pdf,
                icon: Icons.picture_as_pdf,
                title: 'PDF',
                description: global.language('export_report_pdf_desc'),
              ),
            ),
            SizedBox(width: 12),
            Expanded(
              child: _buildFormatOption(
                format: ExportFormat.excel,
                icon: Icons.table_chart,
                title: 'Excel',
                description: global.language('export_report_excel_desc'),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildFormatOption({
    required ExportFormat format,
    required IconData icon,
    required String title,
    required String description,
  }) {
    final isSelected = _selectedFormat == format;

    return GestureDetector(
      onTap: () {
        setState(() {
          _selectedFormat = format;
        });
      },
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          border: Border.all(
            color: isSelected ? Colors.indigo.shade600 : global.theme.dividerBorderColor,
            width: 2,
          ),
          borderRadius: BorderRadius.circular(12),
          color: isSelected ? Colors.indigo.shade50 : global.theme.cardColor,
        ),
        child: Column(
          children: [
            Icon(
              icon,
              size: 32,
              color: isSelected ? Colors.indigo.shade600 : global.theme.textSecondaryColor,
            ),
            const SizedBox(height: 8),
            Text(
              title,
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: isSelected
                    ? Colors.indigo.shade600
                    : global.theme.textColor,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              description,
              style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildExportStatus() {
    return BlocBuilder<BiReportBloc, BiReportState>(
      buildWhen: (previous, current) {
        return current is BiReportExportSubmitting ||
            current is BiReportExportStatusProgress ||
            current is BiReportExportStatusChecking ||
            current is BiReportExportReady ||
            current is BiReportExportSubmitFailure ||
            current is BiReportExportStatusFailure ||
            current is BiReportExportInitial;
      },
      builder: (context, state) {
        return switch (state) {
          BiReportExportSubmitting() => _buildProgressWidget(
            global.language('export_report_submitting'),
            0,
          ),
          BiReportExportStatusChecking() => _buildProgressWidget(
            global.language('export_report_checking_status'),
            10,
          ),
          BiReportExportStatusProgress() => _buildProgressWidget(
            state.statusMessage,
            state.progress,
          ),
          BiReportExportReady() => _buildReadyWidget(state),
          BiReportExportSubmitFailure() => _buildErrorWidget(state.message),
          BiReportExportStatusFailure() => _buildErrorWidget(state.message),
          _ => const SizedBox.shrink(),
        };
      },
    );
  }

  Widget _buildProgressWidget(String message, int progress) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Column(
        children: [
          Row(
            children: [
              SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  valueColor: AlwaysStoppedAnimation<Color>(
                    global.theme.infoHighlightTextColor,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  message,
                  style: TextStyle(
                    fontSize: 14,
                    color: global.theme.infoHighlightTextColor,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
          if (progress > 0) ...[
            const SizedBox(height: 12),
            LinearProgressIndicator(
              value: progress / 100,
              backgroundColor: global.theme.surfaceColor,
              valueColor: AlwaysStoppedAnimation<Color>(global.theme.infoHighlightTextColor),
            ),
            const SizedBox(height: 4),
            Text(
              '$progress%',
              style: TextStyle(
                fontSize: 12,
                color: global.theme.infoHighlightTextColor,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildReadyWidget(BiReportExportReady state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.positiveHighlightColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: global.theme.positiveHighlightColor),
      ),
      child: Column(
        children: [
          Row(
            children: [
              Icon(Icons.check_circle, color: global.theme.positiveHighlightTextColor, size: 24),
              SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      global.language('export_report_file_ready'),
                      style: TextStyle(
                        fontSize: 14,
                        color: global.theme.positiveHighlightTextColor,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    Text(
                      state.fileName,
                      style: TextStyle(
                        fontSize: 12,
                        color: global.theme.positiveHighlightTextColor,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: () => _downloadFile(state.downloadUrl, state.fileName),
              icon: Icon(Icons.download),
              label: Text(global.language('export_report_download')),
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.positiveHighlightTextColor,
                foregroundColor: global.theme.onPrimaryColor,
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorWidget(String message) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.negativeHighlightColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: global.theme.negativeHighlightColor),
      ),
      child: Row(
        children: [
          Icon(Icons.error, color: global.theme.negativeHighlightTextColor, size: 24),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              message,
              style: TextStyle(fontSize: 14, color: global.theme.negativeHighlightTextColor),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildActionButtons() {
    return BlocBuilder<BiReportBloc, BiReportState>(
      builder: (context, state) {
        final isExporting =
            state is BiReportExportSubmitting ||
            state is BiReportExportStatusChecking ||
            state is BiReportExportStatusProgress;

        return Row(
          children: [
            Expanded(
              child: OutlinedButton(
                onPressed: isExporting
                    ? () {
                        context.read<BiReportBloc>().add(
                          const CancelBiReportExportRequested(),
                        );
                      }
                    : () {
                        context.read<BiReportBloc>().add(
                          const ResetBiReportExportState(),
                        );
                        Navigator.of(context).pop();
                      },
                style: OutlinedButton.styleFrom(
                  foregroundColor: global.theme.iconColor,
                  side: BorderSide(color: global.theme.dividerBorderColor),
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                child: Text(isExporting ? global.language('export_report_cancel') : global.language('export_report_close')),
              ),
            ),
            SizedBox(width: 12),
            Expanded(
              child: ElevatedButton(
                onPressed: isExporting || _selectedFormat == null
                    ? null
                    : _startExport,
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.indigo.shade600,
                  foregroundColor: global.theme.onPrimaryColor,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                child: Text(global.language('export_report_start')),
              ),
            ),
          ],
        );
      },
    );
  }

  void _startExport() {
    if (_selectedFormat == null) return;

    final token = global.appConfig.getString("token");
    if (token == null || token.isEmpty) {
      global.showErrorSnackBar(context, global.language('export_report_no_token'));
      return;
    }

    context.read<BiReportBloc>().add(
      SubmitBiReportExportRequested(
        reportType: widget.reportType,
        jobId: widget.jobId,
        format: _selectedFormat!,
        token: token,
      ),
    );
  }

  void _downloadFile(String url, String fileName) async {
    try {
      if (kDebugMode) {
        AppLogger.debug('🌐 Downloading on platform: ${defaultTargetPlatform.name}');
        AppLogger.debug('📥 Download URL: $url');
      }

      if (kIsWeb) {
        // For web platform, use direct download through BLoC
        final state = context.read<BiReportBloc>().state;
        if (state is BiReportExportReady) {
          context.read<BiReportBloc>().add(
            DownloadBiReportExportRequested(
              reportType: state.reportType,
              exportJobId: state.exportJobId,
              fileName: state.fileName,
              token: global.userLoginData.token,
            ),
          );
        }
      } else {
        // For mobile platforms, use external application
        final uri = Uri.parse(url);
        if (await canLaunchUrl(uri)) {
          await launchUrl(uri, mode: LaunchMode.externalApplication);
        } else {
          throw 'Could not launch $url';
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Download error: $e');
      }
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language('export_report_cannot_open_link')}: $e');
      }
    }
  }

  void _showDownloadDialog(BuildContext context, BiReportExportReady state) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language('export_report_file_ready')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(global.language('export_report_complete')),
            SizedBox(height: 8),
            Text(
              '${global.language('export_report_filename')}: ${state.fileName}',
              style: TextStyle(fontWeight: FontWeight.w500),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(global.language('export_report_close')),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.of(context).pop();
              _downloadFile(state.downloadUrl, state.fileName);
            },
            child: Text(global.language('export_report_download')),
          ),
        ],
      ),
    );
  }

  void _showErrorMessage(BuildContext context, BiReportState state) {
    String message = global.language('export_report_error');

    if (state is BiReportExportSubmitFailure) {
      message = state.message;
    } else if (state is BiReportExportStatusFailure) {
      message = state.message;
    }

    global.showErrorSnackBar(context, message);
  }
}
