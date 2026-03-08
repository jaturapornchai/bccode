import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/repositories/docref_repository.dart';

/// Widget แสดง Document Flow - การไหลของเอกสาร (จากไหน → ไปไหน)
class DocFlowWidget extends StatefulWidget {
  final String docno;
  final int transflag;
  final Function(String docno, int transflag)? onDocumentTap;

  const DocFlowWidget({super.key, required this.docno, required this.transflag, this.onDocumentTap});

  @override
  State<DocFlowWidget> createState() => _DocFlowWidgetState();
}

class _DocFlowWidgetState extends State<DocFlowWidget> {
  bool _isLoading = true;
  String? _errorMessage;
  DocFlowResponse? _docFlowData;

  // Colors
  static const Color _primaryColor = Color(0xFF667eea);
  static const Color _accentColor = Color(0xFF764ba2);

  @override
  void initState() {
    super.initState();
    _loadDocFlow();
  }

  @override
  void didUpdateWidget(covariant DocFlowWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.docno != widget.docno || oldWidget.transflag != widget.transflag) {
      _loadDocFlow();
    }
  }

  Future<void> _loadDocFlow() async {
    if (widget.docno.isEmpty) {
      setState(() {
        _isLoading = false;
        _docFlowData = null;
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final docRefRepo = DocRefRepository();
      final result = await docRefRepo.getDocFlow(docno: widget.docno, transflag: widget.transflag);

      if (mounted) {
        setState(() {
          _docFlowData = result;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    // ถ้าไม่มี docno ไม่ต้องแสดงอะไร
    if (widget.docno.isEmpty) {
      return const SizedBox.shrink();
    }

    final hasReferencedBy = _docFlowData?.referencedBy.isNotEmpty ?? false;
    final hasReferenceTo = _docFlowData?.referenceTo.isNotEmpty ?? false;
    final hasAnyDocs = hasReferencedBy || hasReferenceTo;

    if (_isLoading) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 24),
        child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
      );
    }

    if (_errorMessage != null) {
      return _buildErrorState();
    }

    if (_docFlowData == null || !hasAnyDocs) {
      return _buildEmptyState();
    }

    return Card(
      elevation: 0,
      color: Colors.grey.shade50,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
        side: BorderSide(color: Colors.grey.shade200),
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            _buildFlowHeader(hasReferencedBy: hasReferencedBy, hasReferenceTo: hasReferenceTo),
            const SizedBox(height: 6),
            _buildFlowDiagram(),
          ],
        ),
      ),
    );
  }

  Widget _buildFlowHeader({required bool hasReferencedBy, required bool hasReferenceTo}) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(color: _primaryColor.withOpacity(0.15), borderRadius: BorderRadius.circular(14)),
          child: const Icon(Icons.alt_route, color: _primaryColor, size: 20),
        ),
        SizedBox(width: 12),
        Text(global.language('document_flow'), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
        Spacer(),
        if (hasReferencedBy) _buildCountPill(label: global.language('incoming'), count: _docFlowData!.referencedBy.length, color: Colors.orange),
        if (hasReferenceTo) ...[ SizedBox(width: 8), _buildCountPill(label: global.language('outgoing'), count: _docFlowData!.referenceTo.length, color: Colors.green)],
      ],
    );
  }

  Widget _buildCountPill({required String label, required int count, required Color color}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(color: color.withOpacity(0.12), borderRadius: BorderRadius.circular(20)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            label,
            style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: color),
          ),
          const SizedBox(width: 6),
          CircleAvatar(
            radius: 10,
            backgroundColor: color,
            child: Text('$count', style: const TextStyle(fontSize: 10, color: Colors.white)),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.shade300),
      ),
      child: Row(
        children: [
          Icon(Icons.info_outline, color: Colors.grey.shade600),
          SizedBox(width: 8),
          Expanded(
            child: Text(global.language('no_reference_for_document'), style: TextStyle(fontSize: 12, color: Colors.grey.shade700)),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorState() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.red.shade50,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.red.shade200),
      ),
      child: Row(
        children: [
          Icon(Icons.error_outline, color: Colors.red.shade400),
          SizedBox(width: 8),
          Expanded(
            child: Text(_errorMessage ?? global.language('error_loading_data'), style: TextStyle(fontSize: 12, color: Colors.red.shade600)),
          ),
          TextButton(onPressed: _loadDocFlow, child: Text(global.language('retry'))),
        ],
      ),
    );
  }

  Widget _buildFlowDiagram() {
    final incomingDocs = _docFlowData?.referencedBy ?? [];
    final outgoingDocs = _docFlowData?.referenceTo ?? [];

    if (incomingDocs.isEmpty && outgoingDocs.isEmpty) {
      return Padding(
        padding: EdgeInsets.all(16),
        child: Text(
          global.language('no_reference_document'),
          style: TextStyle(fontSize: 13, color: Colors.grey.shade500, fontStyle: FontStyle.italic),
        ),
      );
    }

    // ลำดับ: เอกสารปัจจุบัน (ซ้าย) → ถูกอ้างอิงจาก (ขวา)
    // เพราะตามลำดับเวลา ใบสั่งซื้อมาก่อน แล้วค่อยรับสินค้า
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: EdgeInsets.symmetric(vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // อ้างอิงไปหา (เอกสารต้นทาง) อยู่ทางซ้าย
          if (outgoingDocs.isNotEmpty) _buildNodeLane(title: global.language('reference_to'), color: Colors.green, docs: outgoingDocs, isOnLeftSide: true, isReferencedByDocs: false),
          _buildCurrentNodeColumn(),
          // ถูกอ้างอิงจาก (เอกสารที่สร้างจากเอกสารปัจจุบัน) อยู่ทางขวา
          if (incomingDocs.isNotEmpty) _buildNodeLane(title: global.language('referenced_by'), color: Colors.orange, docs: incomingDocs, isOnLeftSide: false, isReferencedByDocs: true),
        ],
      ),
    );
  }

  /// สร้าง Lane แสดงเอกสารอ้างอิง
  /// [isOnLeftSide] = true หมายถึง lane อยู่ทางซ้ายของเอกสารปัจจุบัน
  /// [isReferencedByDocs] = true หมายถึงเป็นเอกสารที่อ้างอิงมายังเอกสารปัจจุบัน (ใช้ doc.docno)
  Widget _buildNodeLane({required String title, required Color color, required List<DocRefModel> docs, required bool isOnLeftSide, required bool isReferencedByDocs}) {
    return Column(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          _LaneHeader(title: title, color: color, count: docs.length, isIncoming: !isOnLeftSide),
          const SizedBox(height: 12),
          ...docs.asMap().entries.map((entry) {
            final index = entry.key;
            final doc = entry.value;
            // เลือก field ตามประเภทของเอกสาร
            final docno = isReferencedByDocs ? doc.docno : doc.docnoref;
            final transflag = isReferencedByDocs ? doc.docnotransflag : doc.docnoreftransflag;
            final connectorBendUp = index.isEven;

            final node = _FlowNodeCard(docno: docno, transflag: transflag, color: color, onTap: widget.onDocumentTap);

            // ลูกศรชี้ออกจากเอกสารปัจจุบันเสมอ (ชี้ไปทางที่ lane อยู่)
            final connector = _FlowConnector(color: color, bendUp: connectorBendUp, reverse: isOnLeftSide);

            return Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              // ถ้าอยู่ซ้าย: [node, connector] | ถ้าอยู่ขวา: [connector, node]
              child: Row(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.center, children: isOnLeftSide ? [node, connector] : [connector, node]),
            );
          }),
        ],
    );
  }

  Widget _buildCurrentNodeColumn() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          global.language('current_document'),
          style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.grey.shade700),
        ),
        const SizedBox(height: 12),
        _buildCurrentDocument(),
      ],
    );
  }

  Widget _buildCurrentDocument() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _primaryColor.withOpacity(0.3), width: 2),
        boxShadow: [BoxShadow(color: _primaryColor.withOpacity(0.15), blurRadius: 12, offset: const Offset(0, 6))],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              gradient: LinearGradient(colors: [_primaryColor, _accentColor], begin: Alignment.topLeft, end: Alignment.bottomRight),
              borderRadius: BorderRadius.circular(16),
            ),
            child: const Icon(Icons.description_rounded, color: Colors.white, size: 36),
          ),
          const SizedBox(height: 12),
          Text(
            widget.docno,
            style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.black87),
          ),
        ],
      ),
    );
  }
}

class _LaneHeader extends StatelessWidget {
  final String title;
  final Color color;
  final int count;
  final bool isIncoming;

  const _LaneHeader({required this.title, required this.color, required this.count, required this.isIncoming});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
      decoration: BoxDecoration(
        color: color.withOpacity(0.08),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(isIncoming ? Icons.call_received : Icons.call_made, size: 14, color: color),
          const SizedBox(width: 6),
          Text(
            title,
            style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: color),
          ),
          const SizedBox(width: 6),
          CircleAvatar(
            radius: 10,
            backgroundColor: color,
            child: Text('$count', style: const TextStyle(fontSize: 10, color: Colors.white)),
          ),
        ],
      ),
    );
  }
}

/// การ์ดแสดงเอกสารอ้างอิง พร้อม hover effect
class _FlowNodeCard extends StatefulWidget {
  final String docno;
  final int transflag;
  final Color color;
  final Function(String docno, int transflag)? onTap;

  const _FlowNodeCard({required this.docno, required this.transflag, required this.color, required this.onTap});

  @override
  State<_FlowNodeCard> createState() => _FlowNodeCardState();
}

class _FlowNodeCardState extends State<_FlowNodeCard> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final typeName = DocRefRepository.getTransflagName(widget.transflag);

    return MouseRegion(
      cursor: widget.onTap != null ? SystemMouseCursors.click : SystemMouseCursors.basic,
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: GestureDetector(
        onTap: widget.onTap != null ? () => widget.onTap!(widget.docno, widget.transflag) : null,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
          width: 180,
          padding: const EdgeInsets.all(12),
          transform: _isHovered ? Matrix4.diagonal3Values(1.03, 1.03, 1.0) : Matrix4.identity(),
          decoration: BoxDecoration(
            color: _isHovered ? widget.color.withValues(alpha: 0.05) : Colors.white,
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: _isHovered ? widget.color : widget.color.withValues(alpha: 0.5), width: _isHovered ? 2 : 1.5),
            boxShadow: [
              BoxShadow(
                color: widget.color.withValues(alpha: _isHovered ? 0.25 : 0.12),
                blurRadius: _isHovered ? 16 : 8,
                offset: Offset(0, _isHovered ? 6 : 4),
              ),
            ],
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: widget.color.withValues(alpha: _isHovered ? 0.25 : 0.15),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(Icons.article_outlined, size: 18, color: widget.color),
              ),
              SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      widget.docno,
                      style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: _isHovered ? widget.color : Colors.black87),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    SizedBox(height: 2),
                    Text(
                      typeName.isEmpty ? global.language('unknown_type') : typeName,
                      style: TextStyle(fontSize: 10, color: Colors.grey.shade600),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// เส้นเชื่อมระหว่างเอกสาร - แบบเรียบง่ายสวยงาม
class _FlowConnector extends StatelessWidget {
  final Color color;
  final bool bendUp;
  final bool reverse;

  const _FlowConnector({required this.color, required this.bendUp, required this.reverse});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 60,
      height: 50,
      child: CustomPaint(
        painter: _FlowConnectorPainter(color: color, reverse: reverse),
      ),
    );
  }
}

/// วาดเส้นเชื่อมแบบ smooth curve พร้อม gradient
class _FlowConnectorPainter extends CustomPainter {
  final Color color;
  final bool reverse;

  _FlowConnectorPainter({required this.color, required this.reverse});

  @override
  void paint(Canvas canvas, Size size) {
    // เส้นเริ่มจากขอบซ้ายและสิ้นสุดที่ขอบขวา
    final startX = reverse ? size.width : 0.0;
    final endX = reverse ? 0.0 : size.width;
    final centerY = size.height / 2;

    // วาดเส้นโค้ง smooth ด้วย cubic bezier
    final path = Path();
    path.moveTo(startX, centerY);
    path.cubicTo(startX + (endX - startX) * 0.3, centerY - 12, startX + (endX - startX) * 0.7, centerY + 12, endX, centerY);

    // วาดเส้นหลักด้วย gradient
    final gradientPaint = Paint()
      ..shader = LinearGradient(
        colors: [color.withValues(alpha: 0.4), color, color.withValues(alpha: 0.4)],
        begin: Alignment.centerLeft,
        end: Alignment.centerRight,
      ).createShader(Rect.fromLTWH(0, 0, size.width, size.height))
      ..strokeWidth = 2
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round;

    canvas.drawPath(path, gradientPaint);

    // วาดจุดที่จุดเริ่มต้นและจุดสิ้นสุดของเส้น
    final dotPaint = Paint()
      ..color = color
      ..style = PaintingStyle.fill;

    const dotRadius = 4.0;
    canvas.drawCircle(Offset(startX, centerY), dotRadius, dotPaint);
    canvas.drawCircle(Offset(endX, centerY), dotRadius, dotPaint);
  }

  @override
  bool shouldRepaint(covariant _FlowConnectorPainter oldDelegate) {
    return oldDelegate.reverse != reverse || oldDelegate.color != color;
  }
}
