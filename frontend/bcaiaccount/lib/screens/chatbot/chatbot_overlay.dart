import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_inappwebview/flutter_inappwebview.dart';
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
  bool _isLoading = false;
  bool _isMinimized = false;
  late AnimationController _animationController;
  late Animation<double> _scaleAnimation;
  late String _sessionId; // Session ID สำหรับแยก conversation

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
        text: global.language('chatbot_welcome_message'),
        isUser: false,
        timestamp: DateTime.now(),
        suggestedQuestions: [
          global.language('chatbot_q_today_sales'),
          global.language('chatbot_q_top_selling'),
          global.language('chatbot_q_low_stock'),
          global.language('chatbot_q_compare_sales'),
          global.language('chatbot_q_top_customers'),
          global.language('chatbot_q_profit'),
          global.language('chatbot_q_business_health'),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _messageController.dispose();
    _scrollController.dispose();
    _animationController.dispose();
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
    final message = _messageController.text.trim();
    if (message.isEmpty) return;

    setState(() {
      _messages.add(
        ChatMessage(text: message, isUser: true, timestamp: DateTime.now()),
      );
      _isLoading = true;
    });

    _messageController.clear();
    _scrollToBottom();

    try {
      final result = await _callChatAPI(message);

      setState(() {
        _messages.add(
          ChatMessage(
            text: result['text'],
            isUser: false,
            timestamp: DateTime.now(),
            isHtml: false,
            tokenUsage: result['token_usage'],
            suggestedQuestions: result['suggested_questions'],
            cached: result['cached'],
            toolsUsed: result['tools_used'] as List<Map<String, dynamic>>?,
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
            isHtml: false,
            tokenUsage: result['token_usage'],
            suggestedQuestions: result['suggested_questions'],
            cached: result['cached'],
            toolsUsed: result['tools_used'] as List<Map<String, dynamic>>?,
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
            isHtml: false,
            tokenUsage: result['token_usage'],
            suggestedQuestions: result['suggested_questions'],
            cached: result['cached'],
            toolsUsed: result['tools_used'] as List<Map<String, dynamic>>?,
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

  Future<Map<String, dynamic>> _callChatAPI(String message) async {
    final shopId = global.getShopId();

    // ใช้ Agent AI endpoint (chat-agent — ReAct pattern เรียก MCP tools อัตโนมัติ)
    final apiUrl = global.goApiUrlPath('api/v1/chatbot/chat-agent');
    final url = Uri.parse(apiUrl);

    if (kDebugMode) {
      AppLogger.debug('Chatbot API URL: ${url.toString()}');
      AppLogger.debug('Shop ID: $shopId, Session ID: $_sessionId');
    }

    try {
      // สร้าง request body ตาม chat-agent spec
      final requestBody = {
        'shop_id': shopId,
        'question': message,
      };

      final response = await http
          .post(
            url,
            headers: {'Content-Type': 'application/json; charset=utf-8'},
            body: utf8.encode(jsonEncode(requestBody)),
          )
          .timeout(
            const Duration(seconds: 130), // Agent อาจใช้เวลา 5-30 วินาที (backend timeout 120s)
            onTimeout: () {
              throw Exception('ขออภัย ระบบใช้เวลานานเกินไป กรุณาลองใหม่');
            },
          );

      if (kDebugMode) {
        AppLogger.debug('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
        AppLogger.debug('📡 CHAT-AGENT API RESPONSE:');
        AppLogger.debug('Status: ${response.statusCode}');
        AppLogger.debug('Raw body length: ${response.body.length} chars');
      }

      if (response.statusCode == 200) {
        final data = jsonDecode(utf8.decode(response.bodyBytes));

        if (kDebugMode) {
          AppLogger.debug('📥 Parsed JSON:');
          AppLogger.debug(const JsonEncoder.withIndent('  ').convert(data));
          AppLogger.debug('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
        }

        String responseText = '';

        if (data is Map) {
          // ตรวจสอบ success status
          if (data['success'] == false) {
            final errorMsg = data['error']?.toString() ?? data['message']?.toString() ?? 'Unknown error';
            throw Exception(_friendlyError(errorMsg));
          }

          // chat-agent response: {success: true, data: {answer: "...", tools_used: [...]}}
          if (data.containsKey('data') && data['data'] is Map) {
            final responseData = data['data'] as Map<String, dynamic>;
            responseText = responseData['answer']?.toString() ?? '';
          }

          // fallback: ลองดึงจาก response (compat กับ chat-gemini format เก่า)
          if (responseText.isEmpty) {
            responseText = data['response']?.toString() ?? '';
          }

          // fallback: ลองดึงจาก message
          if (responseText.isEmpty && data.containsKey('message')) {
            responseText = data['message'].toString();
          }

          // เก็บข้อมูลเสริม
          final cached = data['cached'] ?? false;
          Map<String, dynamic>? tokenUsage;
          List<String>? suggestedQuestions;
          List<Map<String, dynamic>>? toolsUsed;

          if (data.containsKey('token_usage') && data['token_usage'] is Map) {
            tokenUsage = data['token_usage'] as Map<String, dynamic>;
          }
          if (data.containsKey('suggested_questions') && data['suggested_questions'] is List) {
            suggestedQuestions = (data['suggested_questions'] as List).map((e) => e.toString()).toList();
          }

          // ดึง tools_used จาก data
          if (data.containsKey('data') && data['data'] is Map) {
            final responseData = data['data'] as Map<String, dynamic>;
            if (responseData.containsKey('tools_used') && responseData['tools_used'] is List) {
              toolsUsed = (responseData['tools_used'] as List)
                  .map((e) => e is Map ? Map<String, dynamic>.from(e) : <String, dynamic>{})
                  .where((e) => e.isNotEmpty)
                  .toList();
            }
          }

          if (kDebugMode) {
            AppLogger.debug('📦 Response: ${cached ? "From cache" : "Fresh AI response"}');
            if (tokenUsage != null) {
              AppLogger.debug('💰 Token Usage: ${tokenUsage['total_tokens']} tokens');
            }
            if (toolsUsed != null) {
              AppLogger.debug('🔧 Tools used: ${toolsUsed.length}');
            }
          }

          if (responseText.isEmpty) {
            throw Exception(global.language('chatbot_response_empty'));
          }

          return {
            'text': responseText,
            'cached': cached,
            'token_usage': tokenUsage,
            'suggested_questions': suggestedQuestions,
            'tools_used': toolsUsed,
          };
        } else {
          responseText = response.body;
        }

        if (responseText.isEmpty) {
          throw Exception(global.language('chatbot_response_empty'));
        }

        return {'text': responseText};
      } else if (response.statusCode == 402) {
        throw Exception('Token limit exceeded - กรุณาลองใหม่ภายหลัง');
      } else if (response.statusCode == 504) {
        throw Exception('Gateway timeout - กรุณาลองใหม่อีกครั้ง');
      } else {
        // พยายาม parse JSON error จาก backend
        String errorMsg = 'HTTP ${response.statusCode}: ${response.reasonPhrase}';
        try {
          final errData = jsonDecode(utf8.decode(response.bodyBytes));
          if (errData is Map) {
            final raw = errData['error']?.toString() ?? errData['message']?.toString() ?? '';
            if (raw.isNotEmpty) {
              errorMsg = _friendlyError(raw);
            }
          }
        } catch (_) {
          // ถ้า parse JSON ไม่ได้ ใช้ default error message
        }
        throw Exception(errorMsg);
      }
    } on http.ClientException catch (e) {
      if (kDebugMode) {
        AppLogger.error('ClientException in _callChatAPI: $e');
      }
      throw Exception(
        'ไม่สามารถเชื่อมต่อกับ Chatbot API\n'
        'URL: $url\n'
        'กรุณาตรวจสอบว่า API Server ทำงานอยู่\n'
        'Error: $e',
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error in _callChatAPI: $e');
      }
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
      setState(() {
        // สร้าง Session ID ใหม่
        _sessionId = const Uuid().v4();
        if (kDebugMode) {
          AppLogger.debug('🔄 Reset Chatbot - New Session: $_sessionId');
        }

        _messages.clear();
        _messages.add(
          ChatMessage(
            text: global.language('chatbot_welcome_emoji'),
            isUser: false,
            timestamp: DateTime.now(),
          ),
        );
      });
    }
  }

  // แปลง HTML เป็น Markdown
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

  Widget _buildInputArea() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border(top: BorderSide(color: global.theme.dividerBorderColor)),
      ),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _messageController,
              decoration: InputDecoration(
                hintText: global.language('chatbot_type_message'),
                hintStyle: TextStyle(fontSize: 14, color: global.theme.formHintColor),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(20),
                  borderSide: BorderSide(color: global.theme.formBorderColor),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(20),
                  borderSide: BorderSide(color: global.theme.formBorderColor),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(20),
                  borderSide: BorderSide(color: primaryColor, width: 1.5),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 10,
                ),
                filled: true,
                fillColor: global.theme.formFillColor,
                isDense: true,
              ),
              style: const TextStyle(fontSize: 14),
              maxLines: null,
              textInputAction: TextInputAction.send,
              onSubmitted: (_) => _sendMessage(),
            ),
          ),
          const SizedBox(width: 8),
          Container(
            decoration: BoxDecoration(
              color: primaryColor,
              shape: BoxShape.circle,
            ),
            child: IconButton(
              icon: const Icon(Icons.send, color: Colors.white, size: 18),
              onPressed: _isLoading ? null : _sendMessage,
              padding: const EdgeInsets.all(8),
              constraints: const BoxConstraints(),
            ),
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
                  // แสดง HTML สำหรับ AI response หรือ Text สำหรับ User
                  if (message.isHtml && !message.isUser)
                    // ใช้ InAppWebView เพื่อแสดง HTML เหมือน browser 100%
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
                          Text(
                            '• Model: ${message.tokenUsage!['model'] ?? 'N/A'}',
                            style: const TextStyle(fontSize: 11),
                          ),
                          Text(
                            '• Tokens: ${message.tokenUsage!['total_tokens'] ?? 0}',
                            style: const TextStyle(fontSize: 11),
                          ),
                          Text(
                            '• Cost: ${message.tokenUsage!['cost_thb'] ?? 0} THB',
                            style: const TextStyle(fontSize: 11),
                          ),
                        ],
                      ),
                    ),
                  ],

                  // แสดง Suggested Questions (เฉพาะ AI response)
                  if (!message.isUser &&
                      message.suggestedQuestions != null &&
                      message.suggestedQuestions!.isNotEmpty) ...[
                    SizedBox(height: 8),
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: Colors.purple.shade50,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.purple.shade200),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Icon(
                                Icons.lightbulb_outline,
                                size: 14,
                                color: Colors.purple.shade700,
                              ),
                              SizedBox(width: 4),
                              Text(
                                global.language('chatbot_suggested_questions'),
                                style: TextStyle(
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.purple.shade700,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 6),
                          ...message.suggestedQuestions!.asMap().entries.map((
                            entry,
                          ) {
                            return Padding(
                              padding: const EdgeInsets.only(bottom: 4),
                              child: InkWell(
                                onTap: () {
                                  // กดแล้วส่งข้อความเลย
                                  _messageController.text = entry.value;
                                  _sendMessage();
                                },
                                child: Row(
                                  children: [
                                    Text(
                                      '${entry.key + 1}.',
                                      style: TextStyle(
                                        fontSize: 11,
                                        color: Colors.purple.shade700,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    const SizedBox(width: 4),
                                    Expanded(
                                      child: Text(
                                        entry.value,
                                        style: TextStyle(
                                          fontSize: 11,
                                          color: Colors.purple.shade900,
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            );
                          }),
                        ],
                      ),
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
                            padding: EdgeInsets.all(4),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(
                                  Icons.copy,
                                  size: 12,
                                  color: global.theme.textSecondaryColor,
                                ),
                                SizedBox(width: 4),
                                Text(
                                  global.language('chatbot_copy'),
                                  style: TextStyle(
                                    fontSize: 10,
                                    color: global.theme.textSecondaryColor,
                                  ),
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
  };

  Widget _buildToolsUsedSection(List<Map<String, dynamic>> tools) {
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
          Row(
            children: [
              Icon(Icons.build_circle, size: 14, color: Colors.teal.shade700),
              const SizedBox(width: 4),
              Text(
                global.language('ai_data_source'),
                style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.teal.shade700),
              ),
            ],
          ),
          const SizedBox(height: 4),
          ...tools.map((tool) {
            final toolName = tool['tool']?.toString() ?? '';
            final label = _toolLabels[toolName] ?? toolName;
            final durationMs = tool['duration_ms'] ?? 0;
            final hasError = tool['error'] != null && tool['error'].toString().isNotEmpty;

            return Padding(
              padding: const EdgeInsets.only(top: 2),
              child: Row(
                children: [
                  Icon(
                    hasError ? Icons.error_outline : Icons.check_circle_outline,
                    size: 12,
                    color: hasError ? Colors.red.shade400 : Colors.teal.shade500,
                  ),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(
                      label,
                      style: TextStyle(fontSize: 11, color: Colors.teal.shade800),
                    ),
                  ),
                  Text(
                    '${durationMs}ms',
                    style: TextStyle(fontSize: 10, color: global.theme.iconSecondaryColor),
                  ),
                ],
              ),
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
                      global.language('ai_thinking'),
                      style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
                    ),
                  ],
                ),
                const SizedBox(height: 2),
                Text(
                  global.language('fetching_data_from_system'),
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
  final List<Map<String, dynamic>>? toolsUsed; // Tools ที่ AI เรียกใช้

  ChatMessage({
    required this.text,
    required this.isUser,
    required this.timestamp,
    this.isError = false,
    this.isHtml = false, // default เป็น plain text
    this.tokenUsage, // optional
    this.suggestedQuestions, // optional
    this.cached, // optional
    this.toolsUsed, // optional
  });
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
          baseUrl: WebUri(
            kDebugMode
                ? 'http://localhost:9999'
                : 'https://bcaicallcenter.dedetouch.com',
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
