import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_inappwebview/flutter_inappwebview.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:uuid/uuid.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Chatbot Overlay Widget - แสดงเป็น floating window หรือ embedded panel
class ChatbotOverlay extends StatefulWidget {
  final VoidCallback onClose;
  final bool isEmbedded; // true = แสดงเป็น side panel แทน popup

  const ChatbotOverlay({super.key, required this.onClose, this.isEmbedded = false});

  @override
  State<ChatbotOverlay> createState() => _ChatbotOverlayState();
}

class _ChatbotOverlayState extends State<ChatbotOverlay>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final TextEditingController _messageController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final List<ChatMessage> _messages = [];
  final List<Uint8List> _attachedImages = [];
  final FocusNode _inputFocusNode = FocusNode();
  bool _isLoading = false;
  bool _isMinimized = false;
  bool _searchDatabase = true;
  bool _searchKnowledgeBase = true;
  bool _searchInternet = false;
  String _statusText = ''; // สถานะ SSE (กำลังดึงข้อมูล...)
  late AnimationController _animationController;
  late Animation<double> _scaleAnimation;
  late String _sessionId; // Session ID สำหรับแยก conversation

  // Question history navigation
  List<String> _questionHistory = []; // history of user questions
  int _historyIndex = -1; // -1 means no history selected
  String _savedInput = ''; // save current input when navigating

  // ขนาดและตำแหน่งหน้าต่างที่ปรับได้ (left-based)
  double _width = 400;
  double _height = 600;
  double _posX = -1; // ตำแหน่ง left (-1 = ยังไม่ init, จะคำนวณจาก screen size)
  double _posY = 16; // ตำแหน่ง top
  bool _posInitialized = false;
  final double _minWidth = 300;
  final double _maxWidth = 800;
  final double _minHeight = 400;
  final double _maxHeight = 900;

  // สีธีม
  Color get primaryColor => global.theme.primaryColor;
  Color get primaryDarkColor => global.theme.primaryColor;
  Color get aiMessageColor => global.theme.surfaceColor;
  Color get userMessageColor => global.theme.primaryColor;

  @override
  void initState() {
    super.initState();

    // สร้าง Session ID ใหม่
    _sessionId = const Uuid().v4();
    if (kDebugMode) {
      AppLogger.debug('💬 New Chatbot Session: $_sessionId');
    }

    // Load question history from backend
    _loadQuestionHistory();

    // Handle arrow up/down for history navigation
    _inputFocusNode.onKeyEvent = (node, event) {
      if (event is KeyDownEvent) {
        if (event.logicalKey == LogicalKeyboardKey.arrowUp &&
            _messageController.text.split('\n').length <= 1) {
          _navigateHistory(1); // older
          return KeyEventResult.handled;
        }
        if (event.logicalKey == LogicalKeyboardKey.arrowDown &&
            _messageController.text.split('\n').length <= 1) {
          _navigateHistory(-1); // newer
          return KeyEventResult.handled;
        }
      }
      return KeyEventResult.ignored;
    };

    // Animation setup
    _animationController = AnimationController(
      duration: const Duration(milliseconds: 300),
      vsync: this,
    );
    _scaleAnimation = CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeOutBack,
    );
    _animationController.forward();

    // ข้อความต้อนรับ + ตัวอย่างคำถาม
    _messages.add(
      ChatMessage(
        text: 'สวัสดีค่ะ! น้องกุ้งพร้อมช่วยแล้วค่ะ\n'
            'ถามเรื่องยอดขาย สต็อก ลูกค้า การเงิน ได้เลยนะคะ\n'
            'น้องกุ้งจะวิเคราะห์และดึงข้อมูลให้อัตโนมัติค่ะ',
        isUser: false,
        timestamp: DateTime.now(),
      ),
    );
  }

  @override
  void dispose() {
    _messageController.dispose();
    _scrollController.dispose();
    _animationController.dispose();
    _inputFocusNode.dispose();
    super.dispose();
  }

  /// แปลง Exception เป็น string สะอาด (ลบ "Exception:" prefix)
  String _cleanError(Object e) {
    final s = e.toString();
    // Exception('message') → "Exception: message" → strip prefix
    if (s.startsWith('Exception: ')) return s.substring(11);
    return s;
  }

  Future<void> _sendMessage() async {
    final rawText = _messageController.text.trim();
    final hasImages = _attachedImages.isNotEmpty;
    if (rawText.isEmpty && !hasImages) return;

    final message = rawText.isEmpty ? 'ดูรูปนี้' : rawText;
    _questionHistory.insert(0, message);
    _historyIndex = -1;

    // Save to question history
    _questionHistory.insert(0, message);
    _historyIndex = -1;

    // Snapshot images before clearing
    final imageBytes = List<Uint8List>.from(_attachedImages);
    final images = imageBytes.map((b) => base64Encode(b)).toList();

    setState(() {
      _messages.add(
        ChatMessage(
          text: message,
          isUser: true,
          timestamp: DateTime.now(),
          images: imageBytes.isNotEmpty ? imageBytes : null,
        ),
      );
      _attachedImages.clear();
      _isLoading = true;
    });

    _messageController.clear();
    _scrollToBottom();

    try {
      final result = await _callChatAPI(message, images: images.isNotEmpty ? images : null);

      setState(() {
        _messages.add(
          ChatMessage(
            text: result['text'],
            isUser: false,
            timestamp: DateTime.now(),
            isHtml: _containsHtml(result['text'] as String),
            tokenUsage: result['token_usage'],
            suggestedQuestions: result['suggested_questions'],
            cached: result['cached'],
            toolsUsed: result['tools_used'] as List<Map<String, dynamic>>?,
            citations: result['citations'] as List<Map<String, dynamic>>?,
            thinking: result['thinking'] as String?,
            iterations: result['iterations'] as int?,
            model: (result['token_usage'] as Map<String, dynamic>?)?['model']?.toString(),
            totalElapsedMs: result['total_elapsed_ms'] as int?,
            rawResponse: result['raw_response'] as Map<String, dynamic>?,
          ),
        );
        _isLoading = false;
      });
    } catch (e) {
      final errorText = _cleanError(e);
      final isRateLimit = errorText.contains(global.language('too_frequent')) || errorText.contains('rate limit');

      setState(() {
        _messages.add(
          ChatMessage(
            text: isRateLimit
                ? '$errorText\n\n⏳ กำลังลองใหม่อัตโนมัติใน 30 วินาที... (ครั้งที่ 1/3)'
                : errorText,
            isUser: false,
            timestamp: DateTime.now(),
            isError: true,
          ),
        );
        // ถ้า rate limit → auto-retry หลัง 30 วินาที (backend cooldown = 2 นาที)
        if (isRateLimit) {
          _autoRetry(message, attempt: 1);
        } else {
          _isLoading = false;
        }
      });
    }

    _scrollToBottom();
  }

  /// Auto-retry สำหรับ rate limit — รอ 30 วินาทีแล้วลองใหม่ สูงสุด 3 ครั้ง
  /// Backend cooldown สำหรับ 429 = 2 นาที, retry 30s x 3 = 90s ครอบคลุม
  Future<void> _autoRetry(String message, {int attempt = 1}) async {
    const maxAttempts = 3;
    const retryDelay = 30; // วินาที

    await Future.delayed(const Duration(seconds: retryDelay));
    if (!mounted) return;

    try {
      final result = await _callChatAPI(message);

      setState(() {
        _messages.add(
          ChatMessage(
            text: result['text'],
            isUser: false,
            timestamp: DateTime.now(),
            isHtml: _containsHtml(result['text'] as String),
            tokenUsage: result['token_usage'],
            suggestedQuestions: result['suggested_questions'],
            cached: result['cached'],
            toolsUsed: result['tools_used'] as List<Map<String, dynamic>>?,
            citations: result['citations'] as List<Map<String, dynamic>>?,
            thinking: result['thinking'] as String?,
            iterations: result['iterations'] as int?,
            model: (result['token_usage'] as Map<String, dynamic>?)?['model']?.toString(),
            totalElapsedMs: result['total_elapsed_ms'] as int?,
            rawResponse: result['raw_response'] as Map<String, dynamic>?,
          ),
        );
        _isLoading = false;
      });
    } catch (e) {
      final errorText = _cleanError(e);
      final isRateLimit = errorText.contains(global.language('too_frequent')) || errorText.contains('rate limit');

      if (isRateLimit && attempt < maxAttempts) {
        // ยังมี retry เหลือ → แจ้งแล้วลองอีกครั้ง
        setState(() {
          _messages.add(
            ChatMessage(
              text: '$errorText\n\n⏳ ${global.language("chatbot_retry_message")} $retryDelay ${global.language("seconds")}... (${global.language("attempt")} ${attempt + 1}/$maxAttempts)',
              isUser: false,
              timestamp: DateTime.now(),
              isError: true,
            ),
          );
        });
        _scrollToBottom();
        _autoRetry(message, attempt: attempt + 1);
      } else {
        // หมด retry หรือไม่ใช่ rate limit → จบ
        setState(() {
          _messages.add(
            ChatMessage(
              text: isRateLimit
                  ? '$errorText\n\nลองใหม่ครบ $maxAttempts ครั้งแล้ว กรุณารอสักครู่แล้วลองอีกครั้ง หรือเปลี่ยน AI Model ในหน้า Setup → Integrations'
                  : errorText,
              isUser: false,
              timestamp: DateTime.now(),
              isError: true,
            ),
          );
          _isLoading = false;
        });
      }
    }
    _scrollToBottom();
  }

  // ส่งข้อความจาก HTML (เช่น กดปุ่มวิเคราะห์สินค้า)
  Future<void> _sendMessageFromHTML(String message) async {
    if (message.isEmpty) return;

    setState(() {
      _messages.add(
        ChatMessage(text: message, isUser: true, timestamp: DateTime.now()),
      );
      _isLoading = true;
    });

    _scrollToBottom();

    try {
      final result = await _callInfoAPI(message);

      setState(() {
        _messages.add(
          ChatMessage(
            text: result['text'],
            isUser: false,
            timestamp: DateTime.now(),
            isHtml: _containsHtml(result['text'] as String),
            tokenUsage: result['token_usage'],
            suggestedQuestions: result['suggested_questions'],
            cached: result['cached'],
            toolsUsed: result['tools_used'] as List<Map<String, dynamic>>?,
            citations: result['citations'] as List<Map<String, dynamic>>?,
            thinking: result['thinking'] as String?,
            iterations: result['iterations'] as int?,
            model: (result['token_usage'] as Map<String, dynamic>?)?['model']?.toString(),
            totalElapsedMs: result['total_elapsed_ms'] as int?,
            rawResponse: result['raw_response'] as Map<String, dynamic>?,
          ),
        );
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _messages.add(
          ChatMessage(
            text: _cleanError(e),
            isUser: false,
            timestamp: DateTime.now(),
            isError: true,
          ),
        );
        _isLoading = false;
      });
    }

    _scrollToBottom();
  }

  /// แปลง backend error ให้ user อ่านง่าย
  String _friendlyError(String error) {
    final lower = error.toLowerCase();
    // กรณีไม่มี API key ของ provider ไหนเลย
    if (lower.contains('ไม่มี ai provider ที่พร้อมใช้') || lower.contains('no ai provider')) {
      return 'ไม่มี AI Provider ที่พร้อมใช้งาน\n\n'
          'กรุณาไปที่ ตั้งค่าระบบ → Integrations\n'
          'ใส่ API Key อย่างน้อย 1 ตัว (แนะนำ Groq — ฟรี)';
    }
    // กรณี provider ทุกตัวที่มี key ใช้ไม่ได้
    if (lower.contains('ai provider ทุกตัวใช้ไม่ได้') || lower.contains('all providers failed')) {
      // ดึงชื่อ provider ที่ fail ตัวสุดท้ายจาก error
      final match = RegExp(r'ใช้ไม่ได้: (.+)', caseSensitive: false).firstMatch(error);
      final detail = match?.group(1) ?? '';
      return 'AI Provider ที่มี API Key ลองแล้วใช้ไม่ได้ทุกตัว\n'
          '(ตัวที่ fail จะถูกหยุด 1 ชม. แล้วลองใหม่อัตโนมัติ)\n\n'
          'แก้ไข: เพิ่ม API Key ของ provider อื่น\n'
          'ไปที่ ตั้งค่าระบบ → Integrations → ใส่เพิ่ม'
          '${detail.isNotEmpty ? '\n\nสาเหตุ: $detail' : ''}';
    }
    // กรณี API key ไม่ได้ตั้งค่า
    if (lower.contains('api_key not configured') || lower.contains('api key not configured')) {
      final match = RegExp(r'(\w+)_API_KEY', caseSensitive: false).firstMatch(error);
      final provider = match?.group(1) ?? 'AI';
      return 'ยังไม่ได้ตั้งค่า $provider API Key\n\n'
          'กรุณาไปที่ ตั้งค่าระบบ → Integrations → ใส่ API Key\n'
          'แนะนำ: ใส่หลายตัว ระบบจะ auto-fallback อัตโนมัติ';
    }
    if (lower.contains('rate limit') || lower.contains('too many requests')) {
      return 'เรียก AI บ่อยเกินไป — กรุณารอสักครู่แล้วลองใหม่';
    }
    if (lower.contains('quota exceeded') || lower.contains('insufficient_quota')) {
      return 'AI API โควต้าหมด — กรุณาตรวจสอบบัญชี AI Provider';
    }
    return error;
  }

  Future<void> _loadQuestionHistory() async {
    try {
      final url = Uri.parse(global.goApiUrlPath('api/v1/ai-provider/question-history'));
      final response = await http.post(
        url,
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'shop_id': global.getShopId(),
          'user_id': '',
          'limit': 50,
        }),
      ).timeout(const Duration(seconds: 10));
      final body = jsonDecode(response.body);
      final questions = body['questions'] as List<dynamic>?;
      if (questions != null) {
        setState(() {
          _questionHistory = questions.map((e) => e.toString()).toList();
        });
      }
    } catch (e) {
      if (kDebugMode) AppLogger.debug('Load history error: $e');
    }
  }

  void _navigateHistory(int direction) {
    if (_questionHistory.isEmpty) return;

    if (_historyIndex == -1 && direction > 0) {
      // Save current input before navigating
      _savedInput = _messageController.text;
      _historyIndex = 0;
    } else {
      _historyIndex += direction;
    }

    if (_historyIndex < 0) {
      // Back to current input
      _historyIndex = -1;
      _messageController.text = _savedInput;
      _messageController.selection = TextSelection.fromPosition(
        TextPosition(offset: _messageController.text.length),
      );
      return;
    }

    if (_historyIndex >= _questionHistory.length) {
      _historyIndex = _questionHistory.length - 1;
      return;
    }

    _messageController.text = _questionHistory[_historyIndex];
    _messageController.selection = TextSelection.fromPosition(
      TextPosition(offset: _messageController.text.length),
    );
  }

  Future<Map<String, dynamic>> _callChatAPI(String message, {List<String>? images}) async {
    final shopId = global.getShopId();

    // น้องกุ้ง v2 — SSE streaming endpoint
    final apiUrl = global.goApiUrlPath('api/v1/chatbot/chat-agent-v2');
    final url = Uri.parse(apiUrl);

    if (kDebugMode) {
      AppLogger.debug('🦐 น้องกุ้ง API: ${url.toString()}');
      AppLogger.debug('Shop: $shopId, Session: $_sessionId');
    }

    try {
      final request = http.Request('POST', url);
      request.headers['Content-Type'] = 'application/json; charset=utf-8';
      request.body = jsonEncode({
        'shop_id': shopId,
        'session_id': _sessionId,
        'question': message,
        if (images != null && images.isNotEmpty) 'images': images,
        'search_database': _searchDatabase,
        'search_kb': _searchKnowledgeBase,
        'search_internet': _searchInternet,
      });

      final streamedResponse = await http.Client().send(request).timeout(
        const Duration(seconds: 200),
        onTimeout: () {
          throw Exception('ขออภัย ระบบใช้เวลานานเกินไป กรุณาลองใหม่');
        },
      );

      if (streamedResponse.statusCode != 200) {
        final body = await streamedResponse.stream.bytesToString();
        try {
          final errData = jsonDecode(body);
          if (errData is Map) {
            final raw = errData['error']?.toString() ?? errData['message']?.toString() ?? '';
            if (raw.isNotEmpty) throw Exception(_friendlyError(raw));
          }
        } catch (_) {}
        throw Exception('HTTP ${streamedResponse.statusCode}');
      }

      // Parse response — รองรับทั้ง SSE stream และ JSON sync
      // เริ่มจับเวลา total elapsed
      final requestStart = DateTime.now();
      String answerText = '';
      String thinkingText = '';
      Map<String, dynamic>? tokenUsage;
      List<String>? suggestedQuestions;
      List<Map<String, dynamic>>? toolsUsed;
      List<Map<String, dynamic>>? citations;
      int? iterationsCount;
      Map<String, dynamic>? rawResponse;
      // เก็บ tool_done events real-time — ใช้ดัชนีตามลำดับการเข้ามา
      final List<Map<String, dynamic>> liveTools = [];

      final rawBody = await streamedResponse.stream
          .transform(utf8.decoder)
          .join();

      // ตรวจว่าเป็น SSE (มี "data: ") หรือ JSON ธรรมดา
      final isSSE = rawBody.trimLeft().startsWith('data: ');

      if (!isSSE) {
        // Sync JSON response — backend fallback mode
        try {
          final json = jsonDecode(rawBody) as Map<String, dynamic>;
          rawResponse = json;
          if (json['data'] is Map) {
            final respData = json['data'] as Map<String, dynamic>;
            answerText = respData['answer']?.toString() ?? '';
            thinkingText = respData['thinking']?.toString() ?? '';
            iterationsCount = (respData['iterations'] as num?)?.toInt();
            if (respData.containsKey('tools_used') && respData['tools_used'] is List) {
              toolsUsed = (respData['tools_used'] as List)
                  .map((e) => e is Map ? Map<String, dynamic>.from(e) : <String, dynamic>{})
                  .where((e) => e.isNotEmpty)
                  .toList();
            }
          }
          if (json.containsKey('token_usage') && json['token_usage'] is Map) {
            tokenUsage = json['token_usage'] as Map<String, dynamic>;
          }
          if (json.containsKey('suggested_questions') && json['suggested_questions'] is List) {
            suggestedQuestions = (json['suggested_questions'] as List).map((e) => e.toString()).toList();
          }
          if (json['error'] != null && json['error'].toString().isNotEmpty) {
            throw Exception(_friendlyError(json['error'].toString()));
          }
        } catch (e) {
          if (e is Exception && e.toString().contains('friendlyError')) rethrow;
          if (kDebugMode) AppLogger.debug('JSON parse error: $e');
        }
      } else {
        // SSE stream parsing
        final lines = const LineSplitter().convert(rawBody);

        for (final line in lines) {
          if (!line.startsWith('data: ')) continue;
          final jsonStr = line.substring(6).trim();
          if (jsonStr.isEmpty) continue;

          try {
            final event = jsonDecode(jsonStr) as Map<String, dynamic>;
            final type = event['type'] as String? ?? '';
            final data = event['data'];

            switch (type) {
              case 'status':
                if (mounted && data is String) {
                  setState(() => _statusText = data);
                }
                break;

              case 'thinking':
                // thinking สั้น (status update) → แสดงใน statusText
                // thinking ยาว (extended thinking content) → เก็บไว้ใน thinkingText
                if (data is String) {
                  if (data.length > 200) {
                    thinkingText = data;
                  } else if (mounted) {
                    setState(() => _statusText = data);
                  }
                }
                break;

              case 'tool_start':
                if (mounted && data is Map) {
                  final toolName = data['tool']?.toString() ?? '';
                  setState(() => _statusText = '🔧 กำลังดึงข้อมูล: $toolName');
                }
                break;

              case 'tool_done':
                // เก็บข้อมูล tool ที่จบแล้วเข้า liveTools — รวม params/result_preview/error/duration
                // ถ้า backend ส่ง tools_used ใน done event อยู่แล้ว → ใช้ตัวนั้นแทน (final source of truth)
                if (data is Map) {
                  liveTools.add(Map<String, dynamic>.from(data));
                  final toolName = data['tool']?.toString() ?? '';
                  final dur = data['duration_ms'];
                  final ok = data['success'] == true;
                  if (mounted) {
                    setState(() => _statusText = ok
                        ? '✓ $toolName เสร็จ (${dur}ms)'
                        : '✗ $toolName ผิดพลาด');
                  }
                }
                break;

              case 'answer':
                if (data is String) answerText = data;
                break;

              case 'error':
                throw Exception(_friendlyError(data?.toString() ?? 'Unknown error'));

              case 'done':
                if (data is Map<String, dynamic>) {
                  rawResponse = data;
                  if (data.containsKey('data') && data['data'] is Map) {
                    final respData = data['data'] as Map<String, dynamic>;
                    if (answerText.isEmpty) {
                      answerText = respData['answer']?.toString() ?? '';
                    }
                    if (thinkingText.isEmpty) {
                      thinkingText = respData['thinking']?.toString() ?? '';
                    }
                    iterationsCount = (respData['iterations'] as num?)?.toInt();
                    if (respData.containsKey('tools_used') && respData['tools_used'] is List) {
                      toolsUsed = (respData['tools_used'] as List)
                          .map((e) => e is Map ? Map<String, dynamic>.from(e) : <String, dynamic>{})
                          .where((e) => e.isNotEmpty)
                          .toList();
                    }
                    if (respData.containsKey('citations') && respData['citations'] is List) {
                      citations = (respData['citations'] as List)
                          .map((e) => e is Map ? Map<String, dynamic>.from(e) : <String, dynamic>{})
                          .where((e) => e.isNotEmpty)
                          .toList();
                    }
                  }
                  if (data.containsKey('token_usage') && data['token_usage'] is Map) {
                    tokenUsage = data['token_usage'] as Map<String, dynamic>;
                  }
                  if (data.containsKey('suggested_questions') && data['suggested_questions'] is List) {
                    suggestedQuestions = (data['suggested_questions'] as List).map((e) => e.toString()).toList();
                  }
                }
                break;
            }
          } catch (e) {
            if (e is Exception && e.toString().contains('friendlyError')) rethrow;
            if (kDebugMode) AppLogger.debug('SSE parse skip: $jsonStr');
          }
        }
      }

      // Reset status
      if (mounted) setState(() => _statusText = '');

      if (answerText.isEmpty) {
        throw Exception(global.language('chatbot_response_empty'));
      }

      // Fallback: ถ้า done event ไม่มี tools_used แต่มี liveTools จาก SSE → ใช้ liveTools
      // (เกิดได้ถ้า backend ไม่ส่ง final tools_used หรือ stream หลุดก่อน done)
      if ((toolsUsed == null || toolsUsed.isEmpty) && liveTools.isNotEmpty) {
        toolsUsed = liveTools;
      }

      // คำนวณเวลาที่ใช้ทั้ง request → response
      final totalElapsedMs = DateTime.now().difference(requestStart).inMilliseconds;

      if (kDebugMode) {
        AppLogger.debug('🦐 น้องกุ้ง answer: ${answerText.length} chars (${totalElapsedMs}ms)');
        if (tokenUsage != null) AppLogger.debug('💰 Tokens: ${tokenUsage['total_tokens']}');
        if (toolsUsed != null) AppLogger.debug('🔧 Tools: ${toolsUsed.length}');
      }

      return {
        'text': answerText,
        'cached': false,
        'token_usage': tokenUsage,
        'suggested_questions': suggestedQuestions,
        'tools_used': toolsUsed,
        'citations': citations,
        'thinking': thinkingText.isNotEmpty ? thinkingText : null,
        'iterations': iterationsCount,
        'total_elapsed_ms': totalElapsedMs,
        'raw_response': rawResponse,
      };
    } on http.ClientException catch (e) {
      if (mounted) setState(() => _statusText = '');
      if (kDebugMode) AppLogger.error('ClientException: $e');
      throw Exception('ไม่สามารถเชื่อมต่อกับน้องกุ้ง\nกรุณาตรวจสอบว่า API Server ทำงานอยู่');
    } catch (e) {
      if (mounted) setState(() => _statusText = '');
      if (kDebugMode) AppLogger.error('Error in _callChatAPI: $e');
      rethrow;
    }
  }

  // เรียก API สำหรับค้นหาข้อมูล
  Future<Map<String, dynamic>> _callInfoAPI(String message) async {
    return _callChatAPI(message);
  }

  void _scrollToBottom() {
    Future.delayed(const Duration(milliseconds: 100), () {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }

  void _toggleMinimize() {
    setState(() {
      _isMinimized = !_isMinimized;
    });
  }



  Future<void> _resetChat() async {
    // แสดง confirmation dialog
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Row(
            children: [
              Icon(Icons.refresh, color: Colors.orange),
              SizedBox(width: 10),
              Text(global.language('chatbot_reset_conversation')),
            ],
          ),
          content: Text(
            global.language('chatbot_reset_confirmation'),
            style: TextStyle(fontSize: 14),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: Text(global.language('chatbot_cancel')),
            ),
            ElevatedButton(
              onPressed: () => Navigator.of(context).pop(true),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.orange,
                foregroundColor: Colors.white,
              ),
              child: Text(global.language('chatbot_start_new')),
            ),
          ],
        );
      },
    );

    if (confirmed == true) {
      // ล้าง session บน server
      try {
        final clearUrl = Uri.parse(global.goApiUrlPath('api/v1/chatbot/clear-session'));
        http.post(clearUrl,
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({'shop_id': global.getShopId(), 'session_id': _sessionId}),
        );
      } catch (_) {}

      setState(() {
        // สร้าง Session ID ใหม่
        _sessionId = const Uuid().v4();
        _statusText = '';
        if (kDebugMode) {
          AppLogger.debug('🔄 Reset Chat - New Session: $_sessionId');
        }

        _messages.clear();
        _messages.add(
          ChatMessage(
            text: 'สวัสดีค่ะ! น้องกุ้งพร้อมช่วยแล้วค่ะ 🦐\nถามเรื่องยอดขาย สต็อก ลูกค้า การเงิน ได้เลยนะคะ',
            isUser: false,
            timestamp: DateTime.now(),
          ),
        );
      });
    }
  }

  // แปลง HTML เป็น Markdown
  /// ตรวจว่า text มี HTML tags หรือไม่
  bool _containsHtml(String text) {
    return RegExp(r'<(table|div|style|h[1-6]|ul|ol|p|br|strong|em|a |span|td|th)\b', caseSensitive: false).hasMatch(text);
  }

  String _htmlToMarkdown(String html) {
    // ลบ HTML tags พื้นฐาน
    String markdown = html
        .replaceAll(RegExp(r'<!DOCTYPE[^>]*>', caseSensitive: false), '')
        .replaceAll(RegExp(r'<html[^>]*>', caseSensitive: false), '')
        .replaceAll(RegExp(r'</html>', caseSensitive: false), '')
        .replaceAll(
          RegExp(r'<head[^>]*>.*?</head>', caseSensitive: false, dotAll: true),
          '',
        )
        .replaceAll(
          RegExp(
            r'<style[^>]*>.*?</style>',
            caseSensitive: false,
            dotAll: true,
          ),
          '',
        )
        .replaceAll(
          RegExp(
            r'<script[^>]*>.*?</script>',
            caseSensitive: false,
            dotAll: true,
          ),
          '',
        )
        .replaceAll(RegExp(r'<body[^>]*>', caseSensitive: false), '')
        .replaceAll(RegExp(r'</body>', caseSensitive: false), '')
        .replaceAll(RegExp(r'<div[^>]*>', caseSensitive: false), '')
        .replaceAll(RegExp(r'</div>', caseSensitive: false), '');

    // แปลง HTML tags เป็น Markdown
    markdown = markdown
        // Headers
        .replaceAllMapped(
          RegExp(r'<h1[^>]*>(.*?)</h1>', caseSensitive: false, dotAll: true),
          (match) => '# ${match.group(1)?.trim()}\n\n',
        )
        .replaceAllMapped(
          RegExp(r'<h2[^>]*>(.*?)</h2>', caseSensitive: false, dotAll: true),
          (match) => '## ${match.group(1)?.trim()}\n\n',
        )
        .replaceAllMapped(
          RegExp(r'<h3[^>]*>(.*?)</h3>', caseSensitive: false, dotAll: true),
          (match) => '### ${match.group(1)?.trim()}\n\n',
        )
        // Bold
        .replaceAllMapped(
          RegExp(
            r'<strong[^>]*>(.*?)</strong>',
            caseSensitive: false,
            dotAll: true,
          ),
          (match) => '**${match.group(1)?.trim()}**',
        )
        .replaceAllMapped(
          RegExp(r'<b[^>]*>(.*?)</b>', caseSensitive: false, dotAll: true),
          (match) => '**${match.group(1)?.trim()}**',
        )
        // Italic
        .replaceAllMapped(
          RegExp(r'<em[^>]*>(.*?)</em>', caseSensitive: false, dotAll: true),
          (match) => '*${match.group(1)?.trim()}*',
        )
        .replaceAllMapped(
          RegExp(r'<i[^>]*>(.*?)</i>', caseSensitive: false, dotAll: true),
          (match) => '*${match.group(1)?.trim()}*',
        )
        // Paragraphs
        .replaceAll(RegExp(r'<p[^>]*>', caseSensitive: false), '')
        .replaceAll(RegExp(r'</p>', caseSensitive: false), '\n\n')
        // Line breaks
        .replaceAll(RegExp(r'<br\s*/?>', caseSensitive: false), '\n')
        // Lists
        .replaceAllMapped(
          RegExp(r'<li[^>]*>(.*?)</li>', caseSensitive: false, dotAll: true),
          (match) => '- ${match.group(1)?.trim()}\n',
        )
        .replaceAll(RegExp(r'<ul[^>]*>', caseSensitive: false), '\n')
        .replaceAll(RegExp(r'</ul>', caseSensitive: false), '\n')
        .replaceAll(RegExp(r'<ol[^>]*>', caseSensitive: false), '\n')
        .replaceAll(RegExp(r'</ol>', caseSensitive: false), '\n')
        // Links - try double quotes first, then single quotes
        .replaceAllMapped(
          RegExp(
            r'<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>',
            caseSensitive: false,
            dotAll: true,
          ),
          (match) => '[${match.group(2)?.trim()}](${match.group(1)?.trim()})',
        )
        .replaceAllMapped(
          RegExp(
            r"<a[^>]*href='([^']*)'[^>]*>(.*?)</a>",
            caseSensitive: false,
            dotAll: true,
          ),
          (match) => '[${match.group(2)?.trim()}](${match.group(1)?.trim()})',
        )
        // Code
        .replaceAllMapped(
          RegExp(
            r'<code[^>]*>(.*?)</code>',
            caseSensitive: false,
            dotAll: true,
          ),
          (match) => '`${match.group(1)?.trim()}`',
        )
        .replaceAllMapped(
          RegExp(r'<pre[^>]*>(.*?)</pre>', caseSensitive: false, dotAll: true),
          (match) => '```\n${match.group(1)?.trim()}\n```\n',
        );

    // ลบ HTML entities
    markdown = markdown
        .replaceAll('&nbsp;', ' ')
        .replaceAll('&lt;', '<')
        .replaceAll('&gt;', '>')
        .replaceAll('&amp;', '&')
        .replaceAll('&quot;', '"')
        .replaceAll('&#39;', "'");

    // ลบ whitespace และ newlines ที่เกิน
    markdown = markdown
        .replaceAll(RegExp(r'\n{3,}'), '\n\n')
        .replaceAll(RegExp(r'[ \t]+'), ' ')
        .trim();

    return markdown;
  }

  // Copy ข้อความเป็น Markdown
  Future<void> _copyToClipboard(String text, bool isHtml) async {
    final String contentToCopy = isHtml ? _htmlToMarkdown(text) : text;

    await Clipboard.setData(ClipboardData(text: contentToCopy));

    // แสดง SnackBar
    if (mounted) {
      global.showSuccessSnackBar(context, global.language('chatbot_copied_message'));
    }
  }

  @override
  Widget build(BuildContext context) {
    // Embedded mode: แสดงเป็น side panel เต็มพื้นที่
    if (widget.isEmbedded) {
      return Container(
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          border: Border(
            left: BorderSide(color: global.theme.dividerBorderColor, width: 1),
          ),
        ),
        child: Column(
          children: [
            _buildHeader(),
            if (!_isMinimized) ...[
              Expanded(child: _buildChatContent()),
              _buildInputArea(),
            ],
          ],
        ),
      );
    }

    // Popup mode: แสดงเป็น floating window
    // Init position ครั้งแรก — ชิดขวา
    final screenSize = MediaQuery.of(context).size;
    if (!_posInitialized) {
      _posX = screenSize.width - _width - 16;
      _posInitialized = true;
    }

    return ScaleTransition(
      scale: _scaleAnimation,
      child: Stack(
        children: [
          // Main container with dynamic position (left-based)
          Positioned(
            left: _posX,
            top: _posY,
            child: Stack(
              clipBehavior: Clip.none, // อนุญาตให้ resize handle ยื่นนอกขอบได้
              children: [
                // Main container
                Container(
                  width: _isMinimized ? 300 : _width,
                  height: _isMinimized ? 60 : _height,
                  decoration: BoxDecoration(
                    color: global.theme.cardColor,
                    borderRadius: BorderRadius.circular(16),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.2),
                        blurRadius: 20,
                        offset: const Offset(0, 10),
                      ),
                    ],
                  ),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(16),
                    child: Column(
                      children: [
                        // Header - draggable
                        _buildHeader(),

                        // Chat content (ซ่อนถ้า minimize)
                        if (!_isMinimized) ...[
                          Expanded(child: _buildChatContent()),
                          _buildInputArea(),
                        ],
                      ],
                    ),
                  ),
                ),
                // Resize handles (แสดงเฉพาะเมื่อไม่ minimize)
                if (!_isMinimized) ...[
                  // Top edge resize
                  Positioned(
                    top: 0,
                    left: 8,
                    right: 8,
                    child: GestureDetector(
                      onPanUpdate: (details) {
                        setState(() {
                          final newHeight = _height - details.delta.dy;
                          if (newHeight >= _minHeight &&
                              newHeight <= _maxHeight) {
                            _height = newHeight;
                            _posY += details.delta.dy;
                          }
                        });
                      },
                      child: MouseRegion(
                        cursor: SystemMouseCursors.resizeUpDown,
                        child: Container(height: 8, color: Colors.transparent),
                      ),
                    ),
                  ),
                  // Bottom edge resize
                  Positioned(
                    bottom: 0,
                    left: 8,
                    right: 8,
                    child: GestureDetector(
                      onPanUpdate: (details) {
                        setState(() {
                          _height = (_height + details.delta.dy).clamp(
                            _minHeight,
                            _maxHeight,
                          );
                        });
                      },
                      child: MouseRegion(
                        cursor: SystemMouseCursors.resizeUpDown,
                        child: Container(height: 8, color: Colors.transparent),
                      ),
                    ),
                  ),
                  // Left edge resize — visible drag handle (กว้างสำหรับ touch/iPad)
                  Positioned(
                    top: 48, // เว้น header
                    bottom: 60, // เว้น input area
                    left: -12, // ยื่นออกนอกกรอบ
                    child: GestureDetector(
                      behavior: HitTestBehavior.opaque, // รับ touch ทั้งพื้นที่
                      onPanUpdate: (details) {
                        setState(() {
                          final newWidth = _width - details.delta.dx;
                          if (newWidth >= _minWidth && newWidth <= _maxWidth) {
                            _width = newWidth;
                            _posX += details.delta.dx;
                          }
                        });
                      },
                      child: MouseRegion(
                        cursor: SystemMouseCursors.resizeLeftRight,
                        child: SizedBox(
                          width: 24, // พื้นที่กดกว้าง 24px
                          child: Center(
                            child: Container(
                              width: 6,
                              height: 48,
                              decoration: BoxDecoration(
                                color: global.theme.iconSecondaryColor,
                                borderRadius: BorderRadius.circular(3),
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.black12,
                                    blurRadius: 2,
                                    offset: const Offset(-1, 0),
                                  ),
                                ],
                              ),
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  for (int i = 0; i < 3; i++) ...[
                                    Container(
                                      width: 3,
                                      height: 3,
                                      decoration: BoxDecoration(
                                        color: global.theme.textSecondaryColor,
                                        shape: BoxShape.circle,
                                      ),
                                    ),
                                    if (i < 2) const SizedBox(height: 3),
                                  ],
                                ],
                              ),
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                  // Right edge resize
                  Positioned(
                    top: 8,
                    bottom: 8,
                    right: 0,
                    child: GestureDetector(
                      onPanUpdate: (details) {
                        setState(() {
                          _width = (_width + details.delta.dx).clamp(
                            _minWidth,
                            _maxWidth,
                          );
                        });
                      },
                      child: MouseRegion(
                        cursor: SystemMouseCursors.resizeLeftRight,
                        child: Container(width: 8, color: Colors.transparent),
                      ),
                    ),
                  ),
                  // Top-left corner resize
                  Positioned(
                    top: 0,
                    left: 0,
                    child: GestureDetector(
                      onPanUpdate: (details) {
                        setState(() {
                          final newWidth = _width - details.delta.dx;
                          final newHeight = _height - details.delta.dy;
                          if (newWidth >= _minWidth && newWidth <= _maxWidth) {
                            _width = newWidth;
                            _posX += details.delta.dx;
                          }
                          if (newHeight >= _minHeight &&
                              newHeight <= _maxHeight) {
                            _height = newHeight;
                            _posY += details.delta.dy;
                          }
                        });
                      },
                      child: MouseRegion(
                        cursor: SystemMouseCursors.resizeUpLeftDownRight,
                        child: Container(
                          width: 16,
                          height: 16,
                          color: Colors.transparent,
                        ),
                      ),
                    ),
                  ),
                  // Top-right corner resize
                  Positioned(
                    top: 0,
                    right: 0,
                    child: GestureDetector(
                      onPanUpdate: (details) {
                        setState(() {
                          final newWidth = _width + details.delta.dx;
                          final newHeight = _height - details.delta.dy;
                          if (newWidth >= _minWidth && newWidth <= _maxWidth) {
                            _width = newWidth;
                          }
                          if (newHeight >= _minHeight &&
                              newHeight <= _maxHeight) {
                            _height = newHeight;
                            _posY += details.delta.dy;
                          }
                        });
                      },
                      child: MouseRegion(
                        cursor: SystemMouseCursors.resizeUpRightDownLeft,
                        child: Container(
                          width: 16,
                          height: 16,
                          color: Colors.transparent,
                        ),
                      ),
                    ),
                  ),
                  // Bottom-left corner resize
                  Positioned(
                    bottom: 0,
                    left: 0,
                    child: GestureDetector(
                      onPanUpdate: (details) {
                        setState(() {
                          final newWidth = _width - details.delta.dx;
                          if (newWidth >= _minWidth && newWidth <= _maxWidth) {
                            _width = newWidth;
                            _posX += details.delta.dx;
                          }
                          _height = (_height + details.delta.dy).clamp(
                            _minHeight,
                            _maxHeight,
                          );
                        });
                      },
                      child: MouseRegion(
                        cursor: SystemMouseCursors.resizeUpRightDownLeft,
                        child: Container(
                          width: 16,
                          height: 16,
                          color: Colors.transparent,
                        ),
                      ),
                    ),
                  ),
                  // Bottom-right corner resize
                  Positioned(
                    bottom: 0,
                    right: 0,
                    child: GestureDetector(
                      onPanUpdate: (details) {
                        setState(() {
                          _width = (_width + details.delta.dx).clamp(
                            _minWidth,
                            _maxWidth,
                          );
                          _height = (_height + details.delta.dy).clamp(
                            _minHeight,
                            _maxHeight,
                          );
                        });
                      },
                      child: MouseRegion(
                        cursor: SystemMouseCursors.resizeUpLeftDownRight,
                        child: Container(
                          width: 16,
                          height: 16,
                          color: Colors.transparent,
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader() {
    return GestureDetector(
      onPanUpdate: (details) {
        setState(() {
          _posX += details.delta.dx;
          _posY += details.delta.dy;
        });
      },
      child: MouseRegion(
        cursor: SystemMouseCursors.move,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [primaryDarkColor, primaryColor],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.smart_toy,
                  size: 20,
                  color: Colors.white,
                ),
              ),
              SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'AI Assistant',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    Text(
                      global.language('chatbot_ready_to_help'),
                      style: TextStyle(fontSize: 11, color: Colors.white70),
                    ),
                  ],
                ),
              ),
              // ปุ่ม reset
              IconButton(
                icon: const Icon(Icons.refresh, color: Colors.white, size: 20),
                onPressed: _resetChat,
                tooltip: global.language('chatbot_reset_conversation'),
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
              ),
              const SizedBox(width: 8),
              // ปุ่ม minimize
              IconButton(
                icon: Icon(
                  _isMinimized ? Icons.expand_less : Icons.expand_more,
                  color: Colors.white,
                  size: 20,
                ),
                onPressed: _toggleMinimize,
                tooltip: _isMinimized
                    ? global.language('chatbot_expand')
                    : global.language('chatbot_minimize'),
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
              ),
              const SizedBox(width: 8),
              // ปุ่มปิด
              IconButton(
                icon: const Icon(Icons.close, color: Colors.white, size: 20),
                onPressed: widget.onClose,
                tooltip: global.language('chatbot_close'),
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildChatContent() {
    return Container(
      color: global.theme.surfaceColor,
      child: ListView.builder(
        controller: _scrollController,
        padding: const EdgeInsets.all(12),
        itemCount: _messages.length + (_isLoading ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _messages.length && _isLoading) {
            return _buildTypingIndicator();
          }
          return _buildMessageBubble(_messages[index]);
        },
      ),
    );
  }

  Future<void> _pickImages() async {
    try {
      final picker = ImagePicker();
      final images = await picker.pickMultiImage(imageQuality: 80, maxWidth: 1024);
      if (images.isEmpty) return;
      for (final img in images) {
        final bytes = await img.readAsBytes();
        setState(() => _attachedImages.add(bytes));
      }
    } catch (e) {
      if (kDebugMode) AppLogger.error('Image pick error: $e');
    }
  }

  Widget _buildSearchToggles() {
    final items = [
      (global.language('chatbot_search_database'), _searchDatabase, (bool v) => setState(() => _searchDatabase = v)),
      (global.language('chatbot_search_kb'), _searchKnowledgeBase, (bool v) => setState(() => _searchKnowledgeBase = v)),
      (global.language('chatbot_search_internet'), _searchInternet, (bool v) => setState(() => _searchInternet = v)),
    ];
    return Row(
      children: items.map((item) {
        final label = item.$1;
        final value = item.$2;
        final onChanged = item.$3;
        return Padding(
          padding: const EdgeInsets.only(right: 8),
          child: InkWell(
            onTap: () => onChanged(!value),
            borderRadius: BorderRadius.circular(12),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                SizedBox(
                  width: 18,
                  height: 18,
                  child: Checkbox(
                    value: value,
                    onChanged: (v) => onChanged(v ?? false),
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    activeColor: global.theme.primaryColor,
                  ),
                ),
                const SizedBox(width: 3),
                Text(label, style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor)),
              ],
            ),
          ),
        );
      }).toList(),
    );
  }

  Widget _buildInputArea() {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border(top: BorderSide(color: global.theme.dividerBorderColor)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Image preview strip
          if (_attachedImages.isNotEmpty)
            SizedBox(
              height: 70,
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                itemCount: _attachedImages.length,
                itemBuilder: (ctx, i) => Padding(
                  padding: const EdgeInsets.only(right: 6, bottom: 4),
                  child: Stack(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.memory(_attachedImages[i], height: 64, width: 64, fit: BoxFit.cover),
                      ),
                      Positioned(
                        top: -4,
                        right: -4,
                        child: GestureDetector(
                          onTap: () => setState(() => _attachedImages.removeAt(i)),
                          child: Container(
                            decoration: const BoxDecoration(color: Colors.red, shape: BoxShape.circle),
                            padding: const EdgeInsets.all(2),
                            child: const Icon(Icons.close, size: 12, color: Colors.white),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          // Search source toggles
          _buildSearchToggles(),
          const SizedBox(height: 4),
          // Input row
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              // Upload button
              IconButton(
                icon: Icon(Icons.attach_file, color: global.theme.textSecondaryColor, size: 20),
                onPressed: _pickImages,
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
              ),
              // Text field (multi-line)
              Expanded(
                child: Focus(
                  onKeyEvent: (node, event) {
                    if (event is! KeyDownEvent) return KeyEventResult.ignored;

                    // Enter = ส่ง, Shift+Enter = ขึ้นบรรทัดใหม่
                    if (event.logicalKey == LogicalKeyboardKey.enter &&
                        !HardwareKeyboard.instance.isShiftPressed &&
                        !_isLoading) {
                      WidgetsBinding.instance.addPostFrameCallback((_) {
                        final text = _messageController.text;
                        if (text.endsWith('\n')) {
                          _messageController.text = text.substring(0, text.length - 1);
                        }
                        _sendMessage();
                      });
                      return KeyEventResult.ignored;
                    }

                    // Thai backspace fix: Flutter default ลบ grapheme cluster ทั้งก้อน
                    // (พยัญชนะ + สระ + วรรณยุกต์) ทำให้กด backspace ครั้งเดียวสระหายไปด้วย
                    // แก้โดยลบทีละ rune (code point) — ผู้ใช้ลบวรรณยุกต์ก่อน แล้วค่อยสระ แล้วค่อยพยัญชนะ
                    if (event.logicalKey == LogicalKeyboardKey.backspace &&
                        !HardwareKeyboard.instance.isControlPressed &&
                        !HardwareKeyboard.instance.isAltPressed) {
                      final value = _messageController.value;
                      final sel = value.selection;
                      // ข้ามถ้ากำลัง compose IME, มี selection range, หรือ cursor อยู่หัว
                      if (!sel.isValid || !sel.isCollapsed) return KeyEventResult.ignored;
                      if (value.composing.isValid && !value.composing.isCollapsed) {
                        return KeyEventResult.ignored;
                      }
                      final offset = sel.baseOffset;
                      if (offset <= 0) return KeyEventResult.ignored;
                      final before = value.text.substring(0, offset);
                      final after = value.text.substring(offset);
                      final runes = before.runes.toList();
                      if (runes.isEmpty) return KeyEventResult.ignored;
                      runes.removeLast();
                      final newBefore = String.fromCharCodes(runes);
                      _messageController.value = TextEditingValue(
                        text: newBefore + after,
                        selection: TextSelection.collapsed(offset: newBefore.length),
                      );
                      return KeyEventResult.handled;
                    }

                    return KeyEventResult.ignored;
                  },
                  child: TextField(
                    controller: _messageController,
                    focusNode: _inputFocusNode,
                    decoration: InputDecoration(
                      hintText: global.language('chatbot_type_message'),
                      hintStyle: TextStyle(fontSize: 13, color: global.theme.formHintColor),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(16),
                        borderSide: BorderSide(color: global.theme.formBorderColor),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(16),
                        borderSide: BorderSide(color: global.theme.formBorderColor),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(16),
                        borderSide: BorderSide(color: primaryColor, width: 1.5),
                      ),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                      filled: true,
                      fillColor: global.theme.formFillColor,
                      isDense: true,
                    ),
                    style: const TextStyle(fontSize: 13),
                    maxLines: 5,
                    minLines: 1,
                    keyboardType: TextInputType.multiline,
                    textInputAction: TextInputAction.newline,
                  ),
                ),
              ),
              const SizedBox(width: 4),
              // Send button
              Container(
                decoration: BoxDecoration(color: primaryColor, shape: BoxShape.circle),
                child: IconButton(
                  icon: const Icon(Icons.send, color: Colors.white, size: 18),
                  onPressed: _isLoading ? null : _sendMessage,
                  padding: const EdgeInsets.all(8),
                  constraints: const BoxConstraints(),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildMessageBubble(ChatMessage message) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        mainAxisAlignment: message.isUser
            ? MainAxisAlignment
                  .start // User ชิดซ้าย
            : MainAxisAlignment.end, // AI ชิดขวา
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (message.isUser) ...[
            // ไอคอน User ทางซ้าย
            Container(
              margin: const EdgeInsets.only(right: 6),
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                color: userMessageColor.withValues(alpha: 0.2),
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.person, size: 14, color: Colors.white),
            ),
          ],
          Flexible(
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: message.isUser
                    ? userMessageColor
                    : (message.isError ? Colors.red[50] : aiMessageColor),
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(12),
                  topRight: const Radius.circular(12),
                  bottomLeft: Radius.circular(message.isUser ? 2 : 12),
                  bottomRight: Radius.circular(message.isUser ? 12 : 2),
                ),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // แสดง Thinking (collapsible) — แยก widget จากคำตอบ
                  if (!message.isUser &&
                      message.thinking != null &&
                      message.thinking!.isNotEmpty)
                    _ThinkingWidget(thinking: message.thinking!),

                  // แสดง Iterations badge
                  if (!message.isUser && message.iterations != null && message.iterations! > 1)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 6),
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                        decoration: BoxDecoration(
                          color: Colors.blue.shade50,
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(color: Colors.blue.shade200),
                        ),
                        child: Text(
                          'วิเคราะห์ ${message.iterations} รอบ',
                          style: TextStyle(fontSize: 10, color: Colors.blue.shade700, fontWeight: FontWeight.bold),
                        ),
                      ),
                    ),

                  // แสดงรูปที่แนบมา (user message เท่านั้น)
                  if (message.isUser && message.images != null && message.images!.isNotEmpty) ...[
                    Wrap(
                      spacing: 4,
                      runSpacing: 4,
                      children: message.images!.map((img) => ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.memory(img, height: 80, width: 80, fit: BoxFit.cover),
                      )).toList(),
                    ),
                    const SizedBox(height: 6),
                  ],

                  // แสดง HTML สำหรับ AI response หรือ Text สำหรับ User
                  if (message.isHtml && !message.isUser)
                    _HtmlContentView(
                      htmlContent: message.text,
                      onSendMessage: _sendMessageFromHTML,
                    )
                  else
                    SelectableText(
                      message.text,
                      style: TextStyle(
                        fontSize: 13,
                        color: message.isUser
                            ? Colors.white
                            : (message.isError
                                  ? Colors.red[900]
                                  : global.theme.textColor),
                        height: 1.3,
                      ),
                      toolbarOptions: const ToolbarOptions(
                        copy: true,
                        selectAll: true,
                      ),
                    ),
                  const SizedBox(height: 4),

                  // แสดง Citations (แหล่งอ้างอิง kb/web) เป็นปุ่มกดได้
                  if (!message.isUser &&
                      message.citations != null &&
                      message.citations!.isNotEmpty) ...[
                    const SizedBox(height: 6),
                    _buildCitationsSection(message.citations!),
                  ],

                  // แสดง Tools Used (เฉพาะ AI response ที่ใช้ tools)
                  if (!message.isUser &&
                      message.toolsUsed != null &&
                      message.toolsUsed!.isNotEmpty) ...[
                    const SizedBox(height: 6),
                    _buildToolsUsedSection(message.toolsUsed!),
                  ],

                  // แสดง Token Usage (เฉพาะ AI response)
                  if (!message.isUser && message.tokenUsage != null) ...[
                    const SizedBox(height: 8),
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: global.theme.surfaceColor,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: global.theme.dividerBorderColor),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Icon(
                                Icons.token,
                                size: 14,
                                color: Colors.blue.shade700,
                              ),
                              const SizedBox(width: 4),
                              Text(
                                'Token Usage',
                                style: TextStyle(
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.blue.shade700,
                                ),
                              ),
                              if (message.cached == true) ...[
                                const SizedBox(width: 8),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 6,
                                    vertical: 2,
                                  ),
                                  decoration: BoxDecoration(
                                    color: Colors.green.shade100,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: Text(
                                    'CACHED',
                                    style: TextStyle(
                                      fontSize: 10,
                                      fontWeight: FontWeight.bold,
                                      color: Colors.green.shade700,
                                    ),
                                  ),
                                ),
                              ],
                            ],
                          ),
                          const SizedBox(height: 4),
                          // Model + iterations + elapsed (1 row)
                          Wrap(
                            spacing: 8,
                            runSpacing: 2,
                            children: [
                              _miniStat(
                                Icons.smart_toy_outlined,
                                'Model',
                                '${message.tokenUsage!['model'] ?? 'N/A'}',
                              ),
                              if (message.iterations != null && message.iterations! > 0)
                                _miniStat(
                                  Icons.repeat,
                                  'Iterations',
                                  '${message.iterations}',
                                ),
                              if (message.totalElapsedMs != null)
                                _miniStat(
                                  Icons.timer_outlined,
                                  'Elapsed',
                                  '${(message.totalElapsedMs! / 1000).toStringAsFixed(2)}s',
                                ),
                            ],
                          ),
                          const SizedBox(height: 4),
                          // Tokens — แยก prompt/completion/total
                          Wrap(
                            spacing: 8,
                            runSpacing: 2,
                            children: [
                              _miniStat(
                                Icons.arrow_upward,
                                'Prompt',
                                '${message.tokenUsage!['prompt_tokens'] ?? 0}',
                                color: Colors.blue.shade700,
                              ),
                              _miniStat(
                                Icons.arrow_downward,
                                'Completion',
                                '${message.tokenUsage!['completion_tokens'] ?? 0}',
                                color: Colors.green.shade700,
                              ),
                              _miniStat(
                                Icons.layers,
                                'Total',
                                '${message.tokenUsage!['total_tokens'] ?? 0}',
                                color: Colors.indigo.shade700,
                              ),
                              _miniStat(
                                Icons.attach_money,
                                'Cost',
                                '${message.tokenUsage!['cost_thb'] ?? 0} ฿',
                                color: Colors.amber.shade800,
                              ),
                            ],
                          ),
                          if (message.tokenUsage!['has_thinking'] == false)
                            Padding(
                              padding: const EdgeInsets.only(top: 4),
                              child: Text(
                                '⚠️ Model นี้ไม่มีการวางแผนคำถาม',
                                style: TextStyle(fontSize: 11, color: Colors.orange.shade700),
                              ),
                            ),
                          // Debug raw JSON (collapsible) — เฉพาะเมื่อมี rawResponse
                          if (message.rawResponse != null) ...[
                            const SizedBox(height: 6),
                            _RawJsonDebugPanel(
                              rawResponse: message.rawResponse!,
                              onCopy: () => _copyToClipboard(_prettyJson(message.rawResponse), false),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ],

                  // แสดง Suggested Questions เป็นปุ่มกด
                  if (!message.isUser &&
                      message.suggestedQuestions != null &&
                      message.suggestedQuestions!.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 6,
                      runSpacing: 6,
                      children: message.suggestedQuestions!.map((q) {
                        return Material(
                          color: Colors.transparent,
                          child: InkWell(
                            onTap: () {
                              _messageController.text = q;
                              _sendMessage();
                            },
                            borderRadius: BorderRadius.circular(16),
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                              decoration: BoxDecoration(
                                color: Colors.purple.shade50,
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(color: Colors.purple.shade300),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(Icons.lightbulb_outline, size: 12, color: Colors.purple.shade600),
                                  const SizedBox(width: 4),
                                  Flexible(
                                    child: Text(
                                      q,
                                      style: TextStyle(fontSize: 11, color: Colors.purple.shade800),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        );
                      }).toList(),
                    ),
                  ],

                  // แสดง timestamp และปุ่ม copy
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        _formatTime(message.timestamp),
                        style: TextStyle(
                          fontSize: 10,
                          color: message.isUser
                              ? Colors.white70
                              : global.theme.textSecondaryColor,
                        ),
                      ),
                      // ปุ่ม copy สำหรับข้อความ AI เท่านั้น
                      if (!message.isUser)
                        InkWell(
                          onTap: () =>
                              _copyToClipboard(message.text, message.isHtml),
                          child: Padding(
                            padding: const EdgeInsets.all(4),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(Icons.copy, size: 12, color: global.theme.textSecondaryColor),
                                const SizedBox(width: 4),
                                Text(
                                  global.language('chatbot_copy'),
                                  style: TextStyle(fontSize: 10, color: global.theme.textSecondaryColor),
                                ),
                              ],
                            ),
                          ),
                        ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          if (!message.isUser) ...[
            // ไอคอน AI ทางขวา
            Container(
              margin: const EdgeInsets.only(left: 6),
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                color: primaryColor.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(Icons.smart_toy, size: 14, color: primaryColor),
            ),
          ],
        ],
      ),
    );
  }

  /// Tool name → label ภาษาไทย
  static final Map<String, String> _toolLabels = {
    'search_products': global.language('filter_product'),
    'get_daily_sales': global.language('report_daily_sales'),
    'get_sales_by_date_range': global.language('sales_by_date_range'),
    'get_top_selling_products': global.language('best_selling_products'),
    'get_sales_by_seller': global.language('sales_by_seller'),
    'get_monthly_summary': global.language('monthly_summary'),
    'get_dashboard_kpis': 'KPI Dashboard',
    'get_business_health': global.language('business_health'),
    'get_profit_analysis': global.language('profit_analysis'),
    'get_accounts_receivable': global.language('accounts_receivable'),
    'get_accounts_payable': global.language('accounts_payable'),
    'get_cash_flow': global.language('cash_flow'),
    'get_inventory_value': global.language('inventory_value'),
    'get_low_stock_alerts': global.language('low_stock_alerts'),
    'get_dead_stock': global.language('dead_stock'),
    'get_inventory_turnover': global.language('inventory_turnover'),
    'get_top_customers': global.language('top_customers'),
    'get_customer_growth': global.language('customer_growth'),
    'get_customer_segments': global.language('customer_group'),
    'get_yoy_comparison': global.language('yoy_comparison'),
    'get_mom_comparison': global.language('mom_comparison'),
    'list_units': global.language('product_unit'),
    'query_mongodb': 'Query MongoDB',
    'aggregate_mongodb': 'Aggregate MongoDB',
    'list_mongodb_collections': 'รายชื่อ Collections',
    'query_clickhouse': 'Query ClickHouse',
    'list_clickhouse_tables': 'รายชื่อ Tables',
    'query_postgresql': 'Query PostgreSQL',
    'web_search': 'ค้นหาอินเทอร์เน็ต',
  };

  /// mini stat chip (icon + label + value)
  Widget _miniStat(IconData icon, String label, String value, {Color? color}) {
    final c = color ?? global.theme.iconSecondaryColor;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 11, color: c),
        const SizedBox(width: 3),
        Text(
          '$label:',
          style: TextStyle(fontSize: 10, color: c),
        ),
        const SizedBox(width: 2),
        Text(
          value,
          style: TextStyle(fontSize: 10, color: c, fontWeight: FontWeight.w600),
        ),
      ],
    );
  }

  /// Pretty-print JSON สำหรับโชว์ใน params/result panel
  static String _prettyJson(dynamic value) {
    if (value == null) return '(null)';
    try {
      const encoder = JsonEncoder.withIndent('  ');
      return encoder.convert(value);
    } catch (_) {
      return value.toString();
    }
  }

  /// Source badge config — สี + emoji + label
  static ({String emoji, String text, Color bg, Color border, Color textColor}) _sourceBadgeConfig(String source) {
    switch (source) {
      case 'web_search':
        return (
          emoji: '🌐',
          text: 'ค้นหาจากอินเทอร์เน็ต',
          bg: Colors.blue.shade50,
          border: Colors.blue.shade200,
          textColor: Colors.blue.shade700,
        );
      case 'custom_query':
        return (
          emoji: '💻',
          text: 'เขียนโปรแกรมเอง',
          bg: Colors.purple.shade50,
          border: Colors.purple.shade200,
          textColor: Colors.purple.shade700,
        );
      default:
        return (
          emoji: '📊',
          text: 'ข้อมูลภายในระบบ',
          bg: Colors.teal.shade50,
          border: Colors.teal.shade200,
          textColor: Colors.teal.shade700,
        );
    }
  }

  Widget _buildCitationsSection(List<Map<String, dynamic>> citations) {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.link, size: 14, color: global.theme.primaryColor),
              const SizedBox(width: 4),
              Text(
                'แหล่งอ้างอิง (${citations.length})',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: global.theme.primaryColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Wrap(
            spacing: 6,
            runSpacing: 6,
            children: citations.map((c) {
              final type = (c['type'] ?? '').toString();
              final label = (c['label'] ?? '').toString().isEmpty
                  ? (type == 'web' ? 'Web' : 'เอกสาร')
                  : c['label'].toString();
              final isKB = type == 'kb';
              return InkWell(
                onTap: () => _handleCitationTap(c),
                borderRadius: BorderRadius.circular(14),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                  decoration: BoxDecoration(
                    color: isKB
                        ? global.theme.primaryColor.withValues(alpha: 0.1)
                        : global.theme.infoHighlightColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(
                      color: isKB
                          ? global.theme.primaryColor.withValues(alpha: 0.4)
                          : global.theme.infoHighlightColor.withValues(alpha: 0.5),
                    ),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        isKB ? Icons.description : Icons.public,
                        size: 13,
                        color: isKB ? global.theme.primaryColor : global.theme.infoHighlightColor,
                      ),
                      const SizedBox(width: 4),
                      ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 220),
                        child: Text(
                          label,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            fontSize: 11,
                            color: global.theme.textColor,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  void _handleCitationTap(Map<String, dynamic> citation) {
    final type = (citation['type'] ?? '').toString();
    if (type == 'web') {
      final url = (citation['url'] ?? '').toString();
      if (url.isNotEmpty) {
        _showWebViewDialog(url);
      }
      return;
    }
    // kb → show text dialog
    final label = (citation['label'] ?? 'เอกสาร').toString();
    final text = (citation['text'] ?? '').toString();
    _showKBTextDialog(label, text);
  }

  void _showWebViewDialog(String url) {
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (BuildContext context) => _WebViewDialog(url: url),
    );
  }

  void _showKBTextDialog(String label, String text) {
    showDialog(
      context: context,
      builder: (ctx) => Dialog(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 700, maxHeight: 600),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: global.theme.primaryColor,
                  borderRadius: const BorderRadius.vertical(top: Radius.circular(4)),
                ),
                child: Row(
                  children: [
                    Icon(Icons.description, color: global.theme.onPrimaryColor, size: 18),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        label,
                        style: TextStyle(
                          color: global.theme.onPrimaryColor,
                          fontWeight: FontWeight.bold,
                          fontSize: 14,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    IconButton(
                      icon: Icon(Icons.close, color: global.theme.onPrimaryColor, size: 20),
                      onPressed: () => Navigator.of(ctx).pop(),
                    ),
                  ],
                ),
              ),
              Flexible(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: SelectableText(
                    text.isEmpty ? '(ไม่มีเนื้อหา)' : text,
                    style: TextStyle(
                      fontSize: 13,
                      color: global.theme.textColor,
                      height: 1.6,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildToolsUsedSection(List<Map<String, dynamic>> tools) {
    // นับสถิติ: success/fail + total duration
    final successCount = tools.where((t) => t['error'] == null || t['error'].toString().isEmpty).length;
    final failCount = tools.length - successCount;
    final totalDuration = tools.fold<int>(0, (sum, t) {
      final d = t['duration_ms'];
      return sum + (d is num ? d.toInt() : 0);
    });

    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.teal.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.teal.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header — สรุปจำนวน + เวลา + สถิติ
          Row(
            children: [
              Icon(Icons.build_circle, size: 14, color: Colors.teal.shade700),
              const SizedBox(width: 4),
              Text(
                '${global.language('ai_data_source')} (${tools.length})',
                style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.teal.shade700),
              ),
              const SizedBox(width: 8),
              if (successCount > 0)
                Text(
                  '✓ $successCount',
                  style: TextStyle(fontSize: 10, color: Colors.green.shade700, fontWeight: FontWeight.w600),
                ),
              if (failCount > 0) ...[
                const SizedBox(width: 6),
                Text(
                  '✗ $failCount',
                  style: TextStyle(fontSize: 10, color: Colors.red.shade700, fontWeight: FontWeight.w600),
                ),
              ],
              const Spacer(),
              Text(
                '⏱ ${totalDuration}ms',
                style: TextStyle(fontSize: 10, color: Colors.teal.shade600, fontWeight: FontWeight.w500),
              ),
            ],
          ),
          const SizedBox(height: 6),
          // แต่ละ tool เป็น expandable tile
          ...tools.asMap().entries.map((entry) {
            final idx = entry.key;
            final tool = entry.value;
            return _ToolExpansionTile(
              index: idx + 1,
              tool: tool,
              labelMap: _toolLabels,
              sourceBadgeBuilder: _sourceBadgeConfig,
              prettyJson: _prettyJson,
              onCopyJson: (json) => _copyToClipboard(json, false),
            );
          }),
        ],
      ),
    );
  }

  Widget _buildTypingIndicator() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.end, // AI ชิดขวา
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: aiMessageColor,
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(12),
                topRight: Radius.circular(12),
                bottomLeft: Radius.circular(12),
                bottomRight: Radius.circular(2),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    SizedBox(
                      width: 16, height: 16,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: primaryColor,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      _statusText.isNotEmpty ? 'น้องกุ้ง' : global.language('ai_thinking'),
                      style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
                    ),
                  ],
                ),
                const SizedBox(height: 2),
                Text(
                  _statusText.isNotEmpty ? _statusText : global.language('fetching_data_from_system'),
                  style: TextStyle(fontSize: 10, color: global.theme.iconSecondaryColor),
                ),
              ],
            ),
          ),
          // ไอคอน AI ทางขวา
          Container(
            margin: const EdgeInsets.only(left: 6),
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: primaryColor.withValues(alpha: 0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.smart_toy, size: 14, color: primaryColor),
          ),
        ],
      ),
    );
  }

  String _formatTime(DateTime time) {
    final hour = time.hour.toString().padLeft(2, '0');
    final minute = time.minute.toString().padLeft(2, '0');
    return '$hour:$minute';
  }

}

class ChatMessage {
  final String text;
  final bool isUser;
  final DateTime timestamp;
  final bool isError;
  final bool isHtml; // เพื่อระบุว่าเป็น HTML หรือไม่
  final Map<String, dynamic>? tokenUsage; // Token usage จาก Go API
  final List<String>? suggestedQuestions; // คำถามแนะนำ
  final bool? cached; // ใช้ cache หรือไม่
  final List<Map<String, dynamic>>? toolsUsed; // Tools ที่ AI เรียกใช้ (params/result/error/source)
  final String? thinking; // Extended thinking จาก AI
  final int? iterations; // จำนวนรอบที่ AI วิเคราะห์
  final List<Uint8List>? images; // รูปที่แนบมาพร้อมข้อความ
  final String? model; // ชื่อ model ที่ใช้ตอบ
  final int? totalElapsedMs; // เวลารวม request → response (ms)
  final Map<String, dynamic>? rawResponse; // raw JSON สำหรับ debug panel
  final List<Map<String, dynamic>>? citations; // แหล่งอ้างอิง (kb/web)

  ChatMessage({
    required this.text,
    required this.isUser,
    required this.timestamp,
    this.isError = false,
    this.isHtml = false,
    this.tokenUsage,
    this.suggestedQuestions,
    this.cached,
    this.toolsUsed,
    this.thinking,
    this.iterations,
    this.images,
    this.model,
    this.totalElapsedMs,
    this.rawResponse,
    this.citations,
  });
}

/// Widget สำหรับแสดง Thinking (collapsible)
class _ThinkingWidget extends StatefulWidget {
  final String thinking;
  const _ThinkingWidget({required this.thinking});

  @override
  State<_ThinkingWidget> createState() => _ThinkingWidgetState();
}

class _ThinkingWidgetState extends State<_ThinkingWidget> {
  bool _isExpanded = false;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        gradient: LinearGradient(
          colors: [Colors.blue.shade50, Colors.indigo.shade50],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        border: Border.all(color: Colors.blue.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          InkWell(
            onTap: () => setState(() => _isExpanded = !_isExpanded),
            borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              child: Row(
                children: [
                  Icon(
                    _isExpanded ? Icons.expand_less : Icons.expand_more,
                    size: 16,
                    color: Colors.blue.shade700,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    '💭 วิธีคิดคำตอบ',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                      color: Colors.blue.shade700,
                    ),
                  ),
                ],
              ),
            ),
          ),
          if (_isExpanded) ...[
            Divider(height: 1, color: Colors.blue.shade200),
            Container(
              constraints: const BoxConstraints(maxHeight: 300),
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(10),
                child: SelectableText(
                  widget.thinking,
                  style: TextStyle(
                    fontSize: 11,
                    color: Colors.blueGrey.shade600,
                    height: 1.5,
                  ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

/// Widget สำหรับแสดง Tool call แบบ ExpansionTile
/// เปิดดู params, result preview, error ได้
class _ToolExpansionTile extends StatefulWidget {
  final int index;
  final Map<String, dynamic> tool;
  final Map<String, String> labelMap;
  final ({String emoji, String text, Color bg, Color border, Color textColor}) Function(String) sourceBadgeBuilder;
  final String Function(dynamic) prettyJson;
  final void Function(String) onCopyJson;

  const _ToolExpansionTile({
    required this.index,
    required this.tool,
    required this.labelMap,
    required this.sourceBadgeBuilder,
    required this.prettyJson,
    required this.onCopyJson,
  });

  @override
  State<_ToolExpansionTile> createState() => _ToolExpansionTileState();
}

class _ToolExpansionTileState extends State<_ToolExpansionTile> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final tool = widget.tool;
    final toolName = tool['tool']?.toString() ?? '';
    final label = widget.labelMap[toolName] ?? toolName;
    final durationMs = tool['duration_ms'] ?? 0;
    final iteration = tool['iteration'];
    final hasError = tool['error'] != null && tool['error'].toString().isNotEmpty;
    final source = tool['source']?.toString() ?? 'mcp';
    final params = tool['params'];
    final result = tool['result'] ?? tool['result_preview'];
    final errorText = tool['error']?.toString() ?? '';

    final badge = widget.sourceBadgeBuilder(source);

    return Container(
      margin: const EdgeInsets.only(top: 4),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(
          color: hasError ? Colors.red.shade200 : Colors.teal.shade100,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header — กดเปิด/ปิดได้
          InkWell(
            onTap: () => setState(() => _expanded = !_expanded),
            borderRadius: BorderRadius.circular(6),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
              child: Row(
                children: [
                  // Icon expand/collapse
                  Icon(
                    _expanded ? Icons.expand_less : Icons.expand_more,
                    size: 14,
                    color: Colors.teal.shade600,
                  ),
                  const SizedBox(width: 2),
                  // Index
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                    decoration: BoxDecoration(
                      color: Colors.grey.shade200,
                      borderRadius: BorderRadius.circular(3),
                    ),
                    child: Text(
                      '#${widget.index}',
                      style: TextStyle(fontSize: 9, color: Colors.grey.shade700, fontWeight: FontWeight.w600),
                    ),
                  ),
                  const SizedBox(width: 4),
                  // Status icon
                  Icon(
                    hasError ? Icons.error_outline : Icons.check_circle_outline,
                    size: 12,
                    color: hasError ? Colors.red.shade400 : Colors.teal.shade500,
                  ),
                  const SizedBox(width: 4),
                  // Source badge
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                    decoration: BoxDecoration(
                      color: badge.bg,
                      borderRadius: BorderRadius.circular(4),
                      border: Border.all(color: badge.border),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text(badge.emoji, style: const TextStyle(fontSize: 9)),
                        const SizedBox(width: 2),
                        Text(badge.text, style: TextStyle(fontSize: 9, color: badge.textColor)),
                      ],
                    ),
                  ),
                  const SizedBox(width: 6),
                  // Tool label
                  Expanded(
                    child: Text(
                      label,
                      style: TextStyle(fontSize: 11, color: Colors.teal.shade800, fontWeight: FontWeight.w500),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  // Iteration badge
                  if (iteration != null) ...[
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                      decoration: BoxDecoration(
                        color: Colors.indigo.shade50,
                        borderRadius: BorderRadius.circular(3),
                        border: Border.all(color: Colors.indigo.shade200),
                      ),
                      child: Text(
                        'iter $iteration',
                        style: TextStyle(fontSize: 9, color: Colors.indigo.shade700),
                      ),
                    ),
                    const SizedBox(width: 4),
                  ],
                  // Duration
                  Text(
                    '${durationMs}ms',
                    style: TextStyle(fontSize: 10, color: Colors.grey.shade600, fontWeight: FontWeight.w500),
                  ),
                ],
              ),
            ),
          ),
          // Expanded content — params + result + error
          if (_expanded) ...[
            const Divider(height: 1),
            Padding(
              padding: const EdgeInsets.all(6),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Tool name (raw)
                  Row(
                    children: [
                      Icon(Icons.tag, size: 11, color: Colors.grey.shade500),
                      const SizedBox(width: 3),
                      Text(
                        toolName,
                        style: TextStyle(
                          fontSize: 10,
                          color: Colors.grey.shade600,
                          fontFamily: 'monospace',
                        ),
                      ),
                    ],
                  ),
                  // Params
                  if (params != null) ...[
                    const SizedBox(height: 6),
                    _buildJsonBlock(
                      'Params',
                      Icons.input,
                      Colors.blue,
                      widget.prettyJson(params),
                      onCopy: () => widget.onCopyJson(widget.prettyJson(params)),
                    ),
                  ],
                  // Result
                  if (result != null && !hasError) ...[
                    const SizedBox(height: 6),
                    _buildJsonBlock(
                      'Result Preview',
                      Icons.output,
                      Colors.green,
                      widget.prettyJson(result),
                      onCopy: () => widget.onCopyJson(widget.prettyJson(result)),
                    ),
                  ],
                  // Error
                  if (hasError) ...[
                    const SizedBox(height: 6),
                    _buildJsonBlock(
                      'Error',
                      Icons.error,
                      Colors.red,
                      errorText,
                      onCopy: () => widget.onCopyJson(errorText),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildJsonBlock(String title, IconData icon, MaterialColor color, String content, {VoidCallback? onCopy}) {
    return Container(
      decoration: BoxDecoration(
        color: color.shade50,
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: color.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
            child: Row(
              children: [
                Icon(icon, size: 11, color: color.shade700),
                const SizedBox(width: 4),
                Text(
                  title,
                  style: TextStyle(
                    fontSize: 10,
                    color: color.shade700,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const Spacer(),
                if (onCopy != null)
                  InkWell(
                    onTap: onCopy,
                    child: Padding(
                      padding: const EdgeInsets.all(2),
                      child: Icon(Icons.copy, size: 11, color: color.shade600),
                    ),
                  ),
              ],
            ),
          ),
          // Content (scrollable, max-height 180)
          Container(
            width: double.infinity,
            constraints: const BoxConstraints(maxHeight: 180),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: const BorderRadius.only(
                bottomLeft: Radius.circular(3),
                bottomRight: Radius.circular(3),
              ),
            ),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(6),
              child: SelectableText(
                content,
                style: TextStyle(
                  fontSize: 10,
                  fontFamily: 'monospace',
                  color: Colors.grey.shade800,
                  height: 1.4,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Widget แสดง raw JSON debug panel (collapsible)
/// ใช้ดู response เต็มจาก backend สำหรับ debug + transparency
class _RawJsonDebugPanel extends StatefulWidget {
  final Map<String, dynamic> rawResponse;
  final VoidCallback onCopy;

  const _RawJsonDebugPanel({required this.rawResponse, required this.onCopy});

  @override
  State<_RawJsonDebugPanel> createState() => _RawJsonDebugPanelState();
}

class _RawJsonDebugPanelState extends State<_RawJsonDebugPanel> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final pretty = const JsonEncoder.withIndent('  ').convert(widget.rawResponse);
    final lineCount = pretty.split('\n').length;
    final byteSize = pretty.length;

    return Container(
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.grey.shade300),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          InkWell(
            onTap: () => setState(() => _expanded = !_expanded),
            borderRadius: BorderRadius.circular(6),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
              child: Row(
                children: [
                  Icon(
                    _expanded ? Icons.expand_less : Icons.expand_more,
                    size: 14,
                    color: Colors.grey.shade700,
                  ),
                  const SizedBox(width: 2),
                  Icon(Icons.code, size: 12, color: Colors.grey.shade700),
                  const SizedBox(width: 4),
                  Text(
                    'Raw JSON Debug',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      color: Colors.grey.shade700,
                    ),
                  ),
                  const SizedBox(width: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                    decoration: BoxDecoration(
                      color: Colors.grey.shade200,
                      borderRadius: BorderRadius.circular(3),
                    ),
                    child: Text(
                      '$lineCount lines · ${(byteSize / 1024).toStringAsFixed(1)}KB',
                      style: TextStyle(fontSize: 9, color: Colors.grey.shade600),
                    ),
                  ),
                  const Spacer(),
                  if (_expanded)
                    InkWell(
                      onTap: widget.onCopy,
                      child: Padding(
                        padding: const EdgeInsets.all(2),
                        child: Icon(Icons.copy, size: 12, color: Colors.grey.shade600),
                      ),
                    ),
                ],
              ),
            ),
          ),
          // JSON content
          if (_expanded) ...[
            const Divider(height: 1),
            Container(
              width: double.infinity,
              constraints: const BoxConstraints(maxHeight: 250),
              padding: const EdgeInsets.all(6),
              child: SingleChildScrollView(
                child: SelectableText(
                  pretty,
                  style: TextStyle(
                    fontSize: 9,
                    fontFamily: 'monospace',
                    color: Colors.grey.shade800,
                    height: 1.3,
                  ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

/// Widget สำหรับแสดง HTML Content โดยปรับความสูงอัตโนมัติ
class _HtmlContentView extends StatefulWidget {
  final String htmlContent;
  final Function(String)?
  onSendMessage; // Callback เพื่อส่งข้อความไปยัง chatbot

  const _HtmlContentView({required this.htmlContent, this.onSendMessage});

  @override
  State<_HtmlContentView> createState() => _HtmlContentViewState();
}

class _HtmlContentViewState extends State<_HtmlContentView> {
  double _contentHeight = 100; // ความสูงเริ่มต้น
  // ignore: unused_field
  InAppWebViewController? _controller;
  bool _isDisposed = false;

  @override
  void dispose() {
    _isDisposed = true;
    // ไม่ต้อง clearCache เพราะเป็น async method และจะถูก cleanup อัตโนมัติ
    _controller = null;
    super.dispose();
  }

  // ฟังก์ชันสำหรับ inject JavaScript เพื่อดัก event
  void _setupEventListeners(InAppWebViewController controller) {
    // Handler สำหรับ resize height เมื่อ details toggle
    controller.addJavaScriptHandler(
      handlerName: 'resizeHeight',
      callback: (args) {
        if (_isDisposed || !mounted || args.isEmpty) return;
        try {
          final h = double.tryParse(args[0].toString());
          if (h != null && h > 0) {
            setState(() => _contentHeight = h);
          }
        } catch (_) {}
      },
    );

    // Handler หลักสำหรับ event ทั่วไป
    controller.addJavaScriptHandler(
      handlerName: 'flutterHandler',
      callback: (args) {
        try {
          // ตรวจสอบว่า widget ยัง mount อยู่หรือไม่
          if (_isDisposed || !mounted) {
            if (kDebugMode) {
              AppLogger.debug('Widget disposed, ignoring event');
            }
            return;
          }

          // รับข้อมูลจาก HTML
          if (args.isNotEmpty) {
            final data = args[0];
            if (kDebugMode) {
              AppLogger.debug('Event from HTML: $data');
            }

            // จัดการ event ตามประเภท
            if (data is Map) {
              final eventType = data['type'];
              final eventData = data['data'];

              switch (eventType) {
                case 'button_click':
                  _handleButtonClick(eventData);
                  break;
                case 'link_click':
                  _handleLinkClick(eventData);
                  break;
                case 'form_submit':
                  _handleFormSubmit(eventData);
                  break;
                default:
                  if (kDebugMode) {
                    AppLogger.debug('Unknown event type: $eventType');
                  }
              }
            }
          }
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error in flutterHandler: $e');
          }
        }
      },
    );

    // Handler สำหรับ Product Properties (จาก HTML)
    controller.addJavaScriptHandler(
      handlerName: 'onProductProperties',
      callback: (args) {
        try {
          if (_isDisposed || !mounted) return;

          if (kDebugMode) {
            AppLogger.debug('Product properties received: $args');
          }

          // จัดการข้อมูล product properties
          if (args.isNotEmpty) {
            _handleProductProperties(args[0]);
          }
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error in onProductProperties: $e');
          }
        }
      },
    );

    // Handler สำหรับ Console Messages (ป้องกัน crash)
    controller.addJavaScriptHandler(
      handlerName: 'onConsoleMessage',
      callback: (args) {
        try {
          if (kDebugMode && args.isNotEmpty) {
            AppLogger.debug('Console message: $args');
          }
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error in onConsoleMessage: $e');
          }
        }
      },
    );
  }

  void _handleProductProperties(dynamic data) {
    if (_isDisposed || !mounted) return;

    try {
      if (kDebugMode) {
        AppLogger.debug('Handling product properties: $data');
      }

      // ส่งข้อมูลสินค้ากลับไปถาม AI ผ่าน info-html API
      if (mounted && data is Map) {
        final name = data['name'] ?? '';
        final code = data['code'] ?? '';
        final barcode = data['barcode'] ?? '';
        final unit = data['unit'] ?? '';

        // สร้างข้อความถาม AI
        final message =
            '${global.language('chatbot_analyze_product')}: '
            '${global.language('chatbot_product_name')}: $name, '
            '${global.language('chatbot_product_code')}: $code, '
            '${global.language('chatbot_product_barcode')}: $barcode, '
            '${global.language('chatbot_product_unit')}: $unit';

        if (kDebugMode) {
          AppLogger.debug('Sending to info-html API: $message');
        }

        // ส่งข้อความไปยัง info-html API ผ่าน callback
        widget.onSendMessage?.call(message);
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error handling product properties: $e');
      }
    }
  }

  void _handleButtonClick(dynamic data) {
    if (_isDisposed || !mounted) return;

    if (kDebugMode) {
      AppLogger.debug('Button clicked: $data');
    }

    // ไม่ต้องทำอะไร - Product properties จะถูกจัดการใน _handleProductProperties
  }

  void _handleLinkClick(dynamic data) {
    if (_isDisposed || !mounted) return;

    if (kDebugMode) {
      AppLogger.debug('Link clicked: $data');
    }

    // ดึง URL จาก link
    if (data is Map && data.containsKey('href')) {
      final url = data['href'].toString();
      if (url.isNotEmpty &&
          (url.startsWith('http://') || url.startsWith('https://'))) {
        // เปิด dialog แสดง website
        _showWebViewDialog(url);
      }
    }
  }

  void _showWebViewDialog(String url) {
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (BuildContext context) {
        return _WebViewDialog(url: url);
      },
    );
  }

  void _handleFormSubmit(dynamic data) {
    if (_isDisposed || !mounted) return;

    if (kDebugMode) {
      AppLogger.debug('Form submitted: $data');
    }
    // จัดการฟอร์มที่ submit
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _contentHeight,
      child: InAppWebView(
        initialData: InAppWebViewInitialData(
          data: widget.htmlContent,
          // baseUrl ใช้สำหรับ resolve relative paths ใน HTML —
          // ตอนนี้ระบบ HTML ของน้องกุ้งไม่มี image/external resource (ตาม Output Format rule)
          // ใช้ goApiUrl ของ goapi เป็น base — fallback เป็น about:blank ถ้าไม่ตั้ง
          baseUrl: WebUri(
            global.myAppConfig.goApiUrl.isNotEmpty
                ? global.myAppConfig.goApiUrl
                : 'about:blank',
          ),
        ),
        initialSettings: InAppWebViewSettings(
          transparentBackground: true,
          supportZoom: false,
          disableContextMenu: true,
          javaScriptEnabled: true,
          // ปิด scrollbar และ scroll
          verticalScrollBarEnabled: false,
          horizontalScrollBarEnabled: false,
          disableVerticalScroll: true,
          disableHorizontalScroll: true,
        ),
        onWebViewCreated: (controller) {
          if (_isDisposed || !mounted) return;
          _controller = controller;
          // ตั้งค่า event listeners
          _setupEventListeners(controller);
        },
        onLoadStop: (controller, url) async {
          if (_isDisposed || !mounted) return;

          // Inject JavaScript เพื่อดัก event จากปุ่มใน HTML
          try {
            await controller.evaluateJavascript(
              source: """
            // ดัก event จากปุ่มทั้งหมด
            document.addEventListener('click', function(e) {
              // ถ้าเป็นปุ่ม button หรือ element ที่มี data-action
              if (e.target.tagName === 'BUTTON' || e.target.hasAttribute('data-action')) {
                e.preventDefault(); // ป้องกัน default action
                
                // ส่งข้อมูลไปยัง Flutter
                window.flutter_inappwebview.callHandler('flutterHandler', {
                  type: 'button_click',
                  data: {
                    text: e.target.textContent || e.target.innerText,
                    id: e.target.id,
                    action: e.target.getAttribute('data-action'),
                    value: e.target.getAttribute('data-value'),
                    className: e.target.className
                  }
                });
              }
              
              // ถ้าเป็นลิงก์
              if (e.target.tagName === 'A') {
                e.preventDefault();
                
                window.flutter_inappwebview.callHandler('flutterHandler', {
                  type: 'link_click',
                  data: {
                    text: e.target.textContent,
                    href: e.target.href,
                    target: e.target.target
                  }
                });
              }
            });
            
            // ดัก form submit
            document.addEventListener('submit', function(e) {
              e.preventDefault();
              
              const formData = new FormData(e.target);
              const data = {};
              formData.forEach((value, key) => {
                data[key] = value;
              });
              
              window.flutter_inappwebview.callHandler('flutterHandler', {
                type: 'form_submit',
                data: data
              });
            });
          """,
            );
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error('Error injecting JavaScript: $e');
            }
          }

          // คำนวณความสูงจริงของเนื้อหา
          if (_isDisposed || !mounted) return;

          try {
            final height = await controller.evaluateJavascript(
              source: """
            (function() {
              // ลบ margin/padding
              document.body.style.margin = '0';
              document.body.style.padding = '0';
              document.body.style.overflow = 'hidden';
              
              // คืนค่าความสูงจริงของเนื้อหา
              return Math.max(
                document.body.scrollHeight,
                document.body.offsetHeight,
                document.documentElement.clientHeight,
                document.documentElement.scrollHeight,
                document.documentElement.offsetHeight
              );
            })();
          """,
            );

            if (height != null && mounted && !_isDisposed) {
              setState(() {
                _contentHeight = double.parse(height.toString());
              });
            }
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error('Error calculating content height: $e');
            }
          }
        },
      ),
    );
  }
}

// WebView Dialog Widget with proper disposal
class _WebViewDialog extends StatefulWidget {
  final String url;

  const _WebViewDialog({required this.url});

  @override
  State<_WebViewDialog> createState() => _WebViewDialogState();
}

class _WebViewDialogState extends State<_WebViewDialog> {
  // ignore: unused_field
  InAppWebViewController? _controller;
  bool _isDisposed = false;

  @override
  void dispose() {
    _isDisposed = true;
    _controller = null;
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.all(20),
      child: Container(
        width: MediaQuery.of(context).size.width * 0.9,
        height: MediaQuery.of(context).size.height * 0.9,
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.3),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Column(
          children: [
            // Header bar
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: Colors.blue.shade700,
                borderRadius: const BorderRadius.only(
                  topLeft: Radius.circular(12),
                  topRight: Radius.circular(12),
                ),
              ),
              child: Row(
                children: [
                  const Icon(Icons.language, color: Colors.white, size: 20),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      widget.url,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 14,
                        fontWeight: FontWeight.w500,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, color: Colors.white),
                    onPressed: () {
                      if (!_isDisposed && mounted) {
                        Navigator.of(context).pop();
                      }
                    },
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                  ),
                ],
              ),
            ),
            // WebView content
            Expanded(
              child: ClipRRect(
                borderRadius: const BorderRadius.only(
                  bottomLeft: Radius.circular(12),
                  bottomRight: Radius.circular(12),
                ),
                child: InAppWebView(
                  initialUrlRequest: URLRequest(url: WebUri(widget.url)),
                  initialSettings: InAppWebViewSettings(
                    javaScriptEnabled: true,
                    supportZoom: true,
                    useWideViewPort: true,
                    loadWithOverviewMode: true,
                    builtInZoomControls: true,
                    displayZoomControls: false,
                  ),
                  onWebViewCreated: (controller) {
                    if (!_isDisposed && mounted) {
                      _controller = controller;
                    }
                  },
                  onLoadStart: (controller, url) {
                    if (!_isDisposed && mounted && kDebugMode) {
                      AppLogger.debug('Dialog WebView Loading: $url');
                    }
                  },
                  onLoadStop: (controller, url) {
                    if (!_isDisposed && mounted && kDebugMode) {
                      AppLogger.info('Dialog WebView Loaded: $url');
                    }
                  },
                  onLoadError: (controller, url, code, message) {
                    if (!_isDisposed && mounted && kDebugMode) {
                      AppLogger.error('Dialog WebView Load error: $message');
                    }
                  },
                  onProgressChanged: (controller, progress) {
                    if (!_isDisposed && mounted && kDebugMode) {
                      AppLogger.debug(
                        'Dialog WebView Loading progress: $progress%',
                      );
                    }
                  },
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
