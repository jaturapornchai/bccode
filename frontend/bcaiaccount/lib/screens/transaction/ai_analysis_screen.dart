import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/transaction_model.dart';

/// หน้าจอ AI วิเคราะห์เอกสาร
/// เปิดเป็น tab ใหญ่ใหม่ สำหรับวิเคราะห์เอกสารด้วย AI
/// - สามารถเปิดจากหน้าเมนูหลัก (ไม่มีข้อมูลเอกสาร)
/// - หรือเปิดจากหน้า Transaction Edit (มีข้อมูลเอกสาร)
class AiAnalysisScreen extends StatefulWidget {
  /// ข้อมูลเอกสาร - optional เพราะอาจเปิดจากเมนูหลัก
  final TransactionModel? documentData;

  /// ประเภทเอกสาร - optional
  final global.TransactionTypeEnum? transactionType;

  const AiAnalysisScreen({
    super.key,
    this.documentData,
    this.transactionType,
  });

  @override
  State<AiAnalysisScreen> createState() => _AiAnalysisScreenState();
}

class _AiAnalysisScreenState extends State<AiAnalysisScreen>
    with global.ThemeRefreshMixin {
  @override
  Widget build(BuildContext context) {
    final hasDocument = widget.documentData != null &&
        widget.documentData!.docno.isNotEmpty;

    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false, // ไม่แสดงปุ่ม back เพราะปิดจาก tab แทน
        title: Row(
          children: [
            Icon(Icons.auto_awesome, color: Colors.amber),
            SizedBox(width: 8),
            Text('AI ${global.language("analyze_document")}'),
          ],
        ),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.auto_awesome, size: 80, color: Colors.amber.shade300),
            const SizedBox(height: 24),
            Text(
              'AI Analysis',
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
                color: Colors.amber.shade700,
              ),
            ),
            const SizedBox(height: 16),
            if (hasDocument) ...[
              // มีข้อมูลเอกสาร - แสดงรายละเอียด
              Text(
                '${global.language('document')}: ${widget.documentData!.docno}',
                style: TextStyle(fontSize: 16),
              ),
              Text(
                '${global.language('items_count')}: ${widget.documentData!.details?.length ?? 0}',
                style: TextStyle(fontSize: 16),
              ),
            ] else ...[
              // ไม่มีข้อมูลเอกสาร - แสดงข้อความแนะนำ
              Text(
                global.language('welcome_ai_analysis'),
                style: TextStyle(fontSize: 16),
              ),
              SizedBox(height: 8),
              Text(
                global.language('ai_analysis_open_from_document'),
                style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
              ),
            ],
            SizedBox(height: 32),
            Text(
              '${global.language('developing')}...',
              style: TextStyle(fontSize: 14, color: global.theme.iconSecondaryColor),
            ),
          ],
        ),
      ),
    );
  }
}
