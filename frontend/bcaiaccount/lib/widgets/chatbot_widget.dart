import 'dart:async';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Reusable Chatbot Widget
/// ใช้งานได้ทั้งแบบ full screen และแบบฝังในหน้าอื่น
///
/// 🤖 AI Chatbot API (Go Backend)
/// - Endpoint: POST /api/v1/chatbot/chat-html
/// - Features: Product & Customer queries, HTML responses, 15min cache
class ChatbotWidget extends StatefulWidget {
  final String? title;
  final bool showAppBar;
  final String? welcomeMessage;
  final String? shopId; // Shop ID สำหรับ query
  final String functionName; // "product" หรือ "customer"

  const ChatbotWidget({
    super.key,
    this.title,
    this.showAppBar = true,
    this.welcomeMessage,
    this.shopId,
    this.functionName = 'product', // default: product
  });

  @override
  State<ChatbotWidget> createState() => _ChatbotWidgetState();
}

class _ChatbotWidgetState extends State<ChatbotWidget> {
  final TextEditingController _messageController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final List<ChatMessage> _messages = [];
  bool _isLoading = false;

  // สีธีม
  final Color primaryColor = const Color(0xFF1A73E8);
  final Color primaryDarkColor = const Color(0xFF0D47A1);
  final Color aiMessageColor = const Color(0xFFE3F2FD);
  final Color userMessageColor = const Color(0xFF1A73E8);

  // API URL - ใช้ Go API endpoint (URL มาจากตั้งค่าระบบ)
  String get chatApiUrl {
    final path = global.goApiUrlPath('api/v1/chatbot/chat-html');
    AppLogger.info('Chatbot API URL: $path');
    return path;
  }

  @override
  void initState() {
    super.initState();
    // ข้อความต้อนรับ
    _messages.add(
      ChatMessage(
        text:
            widget.welcomeMessage ??
            global.language("chatbot_welcome_message"),
        isUser: false,
        timestamp: DateTime.now(),
      ),
    );
  }

  @override
  void dispose() {
    _messageController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _sendMessage() async {
    final message = _messageController.text.trim();
    if (message.isEmpty) return;

    // เพิ่มข้อความของ user
    setState(() {
      _messages.add(
        ChatMessage(text: message, isUser: true, timestamp: DateTime.now()),
      );
      _isLoading = true;
    });

    _messageController.clear();
    _scrollToBottom();

    // เรียก API
    try {
      final response = await _callChatAPI(message);

      setState(() {
        _messages.add(
          ChatMessage(
            text: response['html'] ?? response['text'] ?? '',
            isUser: false,
            timestamp: DateTime.now(),
            tokenUsage: response['token_usage'],
            suggestedQuestions: response['suggested_questions'],
            cached: response['cached'],
          ),
        );
        _isLoading = false;
      });
    } on TimeoutException catch (e) {
      // Timeout errors
      AppLogger.error('⏱️ Timeout error: $e');
      setState(() {
        _messages.add(
          ChatMessage(
            text:
                '⏱️ ${global.language("chatbot_timeout_retry")}',
            isUser: false,
            timestamp: DateTime.now(),
            isError: true,
          ),
        );
        _isLoading = false;
      });
    } on http.ClientException catch (e) {
      // Network/Connection errors
      AppLogger.error('❌ Connection error: $e');
      setState(() {
        _messages.add(
          ChatMessage(
            text:
                '❌ ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้\n\n'
                '🔗 API URL: $chatApiUrl\n\n'
                '💡 กรุณาตรวจสอบ:\n'
                '1. API Server ทำงานอยู่หรือไม่\n'
                '2. Port และ URL ถูกต้องหรือไม่\n'
                '3. Firewall/Network settings\n\n'
                'Error: ${e.message}',
            isUser: false,
            timestamp: DateTime.now(),
            isError: true,
          ),
        );
        _isLoading = false;
      });
    } catch (e) {
      // Other errors
      AppLogger.error('❌ Error: $e');
      setState(() {
        _messages.add(
          ChatMessage(
            text: '${global.language("chatbot_error_message")}\n$e',
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

  Future<Map<String, dynamic>> _callChatAPI(String message) async {
    // ใช้ shop_id จาก widget หรือ fallback
    final shopId = widget.shopId ?? '33UYr4vEDECjXsql4x3Lb4FdWfX';

    AppLogger.debug('🤖 Sending chat request to Go API');
    AppLogger.debug('   Question: $message');
    AppLogger.debug('   Shop ID: $shopId');
    AppLogger.debug('   Function: ${widget.functionName}');
    AppLogger.debug('   API URL: $chatApiUrl');

    // สร้าง request body ตาม Go API spec
    final requestBody = {
      'shop_id': shopId,
      'question': message,
      'function_name': widget.functionName,
    };

    final response = await http
        .post(
          Uri.parse(chatApiUrl),
          headers: {'Content-Type': 'application/json; charset=utf-8'},
          body: utf8.encode(jsonEncode(requestBody)),
        )
        .timeout(
          const Duration(seconds: 90), // AI response อาจใช้เวลา 14-30 วินาที
          onTimeout: () {
            throw Exception(
              'Request timeout - AI กำลังคิด กรุณาลองใหม่อีกครั้ง',
            );
          },
        );

    AppLogger.debug('✅ Response status: ${response.statusCode}');

    if (response.statusCode == 200) {
      final data = jsonDecode(utf8.decode(response.bodyBytes));

      // ตรวจสอบ response format จาก Go API
      if (data is Map) {
        // ตรวจสอบ success status
        if (data['success'] == false) {
          final errorMsg = data['error'] ?? data['message'] ?? 'Unknown error';
          throw Exception(errorMsg);
        }

        // ดึง HTML จาก data.html (ตามโครงสร้าง API ใหม่)
        if (data.containsKey('data') && data['data'] is Map) {
          final responseData = data['data'] as Map<String, dynamic>;
          final html =
              responseData['html']?.toString() ??
              responseData['answer']?.toString() ??
              '';
          final cached = data['cached'] ?? false;

          AppLogger.debug(
            '📦 Response: ${cached ? "From cache (15min)" : "Fresh AI response"}',
          );

          // ดึงข้อมูล token usage และ suggested questions
          Map<String, dynamic>? tokenUsage;
          List<String>? suggestedQuestions;

          if (data.containsKey('token_usage')) {
            tokenUsage = data['token_usage'] as Map<String, dynamic>;
            AppLogger.debug(
              '💰 Token: ${tokenUsage['total_tokens']}, Cost: ${tokenUsage['cost_thb']} THB',
            );
          }

          if (data.containsKey('suggested_questions')) {
            final suggestions = data['suggested_questions'] as List;
            suggestedQuestions = suggestions.map((e) => e.toString()).toList();
            AppLogger.debug('💡 Suggested: ${suggestions.length} questions');
          }

          return {
            'html': html,
            'cached': cached,
            'token_usage': tokenUsage,
            'suggested_questions': suggestedQuestions,
          };
        } else if (data.containsKey('message')) {
          return {'html': data['message'].toString()};
        }
      }

      // ถ้าไม่มี format ที่รู้จัก
      return {'html': response.body};
    } else if (response.statusCode == 402) {
      throw Exception('Token limit exceeded - กรุณาลองใหม่ภายหลัง');
    } else if (response.statusCode == 504) {
      throw Exception('Gateway timeout - AI ใช้เวลานานเกินไป กรุณาลองใหม่');
    } else {
      throw Exception('HTTP ${response.statusCode}: ${response.reasonPhrase}');
    }
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[100],
      appBar: widget.showAppBar ? _buildAppBar() : null,
      body: Column(
        children: [
          // พื้นที่แสดงข้อความ
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.all(16),
              itemCount: _messages.length + (_isLoading ? 1 : 0),
              itemBuilder: (context, index) {
                if (index == _messages.length && _isLoading) {
                  return _buildTypingIndicator();
                }
                return _buildMessageBubble(_messages[index]);
              },
            ),
          ),

          // พื้นที่พิมพ์ข้อความ
          _buildChatInput(),
        ],
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      elevation: 2,
      backgroundColor: primaryDarkColor,
      foregroundColor: Colors.white,
      title: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.2),
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(Icons.smart_toy, size: 24),
          ),
          SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                widget.title ?? 'AI Assistant',
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(
                global.language("chatbot_ready_to_help"),
                style: const TextStyle(fontSize: 12, fontWeight: FontWeight.normal),
              ),
            ],
          ),
        ],
      ),
      actions: [
        // ปุ่มล้างข้อความ
        IconButton(
          icon: Icon(Icons.delete_outline),
          tooltip: global.language("chatbot_clear_messages"),
          onPressed: () {
            showDialog(
              context: context,
              builder: (context) => AlertDialog(
                title: Text(global.language("chatbot_confirm_clear_title")),
                content: Text(global.language("chatbot_confirm_clear_message")),
                actions: [
                  TextButton(
                    onPressed: () => Navigator.pop(context),
                    child: Text(global.language("cancel")),
                  ),
                  ElevatedButton(
                    onPressed: () {
                      setState(() {
                        _messages.clear();
                        _messages.add(
                          ChatMessage(
                            text:
                                widget.welcomeMessage ??
                                global.language("chatbot_welcome_short"),
                            isUser: false,
                            timestamp: DateTime.now(),
                          ),
                        );
                      });
                      Navigator.pop(context);
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.red,
                      foregroundColor: Colors.white,
                    ),
                    child: Text(global.language("chatbot_clear_messages")),
                  ),
                ],
              ),
            );
          },
        ),
      ],
    );
  }

  Widget _buildMessageBubble(ChatMessage message) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        mainAxisAlignment: message.isUser
            ? MainAxisAlignment.end
            : MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (!message.isUser) ...[
            // AI Avatar
            Container(
              margin: const EdgeInsets.only(right: 8),
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: primaryColor.withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(Icons.smart_toy, size: 20, color: primaryColor),
            ),
          ],
          // Message bubble
          Flexible(
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: message.isUser
                    ? userMessageColor
                    : (message.isError ? Colors.red[50] : aiMessageColor),
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(16),
                  topRight: const Radius.circular(16),
                  bottomLeft: Radius.circular(message.isUser ? 16 : 4),
                  bottomRight: Radius.circular(message.isUser ? 4 : 16),
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.05),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    message.text,
                    style: TextStyle(
                      fontSize: 15,
                      color: message.isUser
                          ? Colors.white
                          : (message.isError
                                ? Colors.red[900]
                                : Colors.black87),
                      height: 1.4,
                    ),
                  ),
                  const SizedBox(height: 4),

                  // แสดง Token Usage (เฉพาะ AI response)
                  if (!message.isUser && message.tokenUsage != null) ...[
                    const SizedBox(height: 8),
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: Colors.blue.shade50,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.blue.shade200),
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
                                global.language("chatbot_suggested_questions"),
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
                                  _messageController.text = entry.value;
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

                  const SizedBox(height: 4),
                  Text(
                    _formatTime(message.timestamp),
                    style: TextStyle(
                      fontSize: 11,
                      color: message.isUser
                          ? Colors.white70
                          : Colors.grey.shade600,
                    ),
                  ),
                ],
              ),
            ),
          ),
          if (message.isUser) ...[
            // User Avatar
            Container(
              margin: const EdgeInsets.only(left: 8),
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: userMessageColor.withOpacity(0.2),
                shape: BoxShape.circle,
              ),
              child: Icon(Icons.person, size: 20, color: userMessageColor),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildTypingIndicator() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // AI Avatar
          Container(
            margin: const EdgeInsets.only(right: 8),
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: primaryColor.withOpacity(0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.smart_toy, size: 20, color: primaryColor),
          ),
          // Typing animation
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: aiMessageColor,
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(16),
                topRight: Radius.circular(16),
                bottomLeft: Radius.circular(4),
                bottomRight: Radius.circular(16),
              ),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                _buildDot(0),
                const SizedBox(width: 4),
                _buildDot(1),
                const SizedBox(width: 4),
                _buildDot(2),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDot(int index) {
    return TweenAnimationBuilder(
      tween: Tween<double>(begin: 0.0, end: 1.0),
      duration: const Duration(milliseconds: 600),
      builder: (context, double value, child) {
        return Transform.translate(
          offset: Offset(
            0,
            -8 * (0.5 - (0.5 - ((value + index / 3) % 1.0 - 0.5).abs()).abs()),
          ),
          child: Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              color: primaryColor,
              shape: BoxShape.circle,
            ),
          ),
        );
      },
      onEnd: () {
        if (mounted) {
          setState(() {}); // Restart animation
        }
      },
    );
  }

  String _formatTime(DateTime time) {
    final hour = time.hour.toString().padLeft(2, '0');
    final minute = time.minute.toString().padLeft(2, '0');
    return '$hour:$minute';
  }

  Widget _buildChatInput() {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 4,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          // ช่องพิมพ์ข้อความ
          Expanded(
            child: TextField(
              controller: _messageController,
              decoration: InputDecoration(
                hintText: global.language("chatbot_type_message"),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: BorderSide(color: primaryColor, width: 2),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 12,
                ),
                filled: true,
                fillColor: Colors.grey[50],
              ),
              maxLines: null,
              textInputAction: TextInputAction.send,
              onSubmitted: (_) => _sendMessage(),
            ),
          ),
          const SizedBox(width: 12),
          // ปุ่มส่งข้อความ
          Container(
            decoration: BoxDecoration(
              color: primaryColor,
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: primaryColor.withOpacity(0.3),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: IconButton(
              icon: const Icon(Icons.send, color: Colors.white),
              onPressed: _isLoading ? null : _sendMessage,
            ),
          ),
        ],
      ),
    );
  }
}

class ChatMessage {
  final String text;
  final bool isUser;
  final DateTime timestamp;
  final bool isError;
  final Map<String, dynamic>? tokenUsage;
  final List<String>? suggestedQuestions;
  final bool? cached;

  ChatMessage({
    required this.text,
    required this.isUser,
    required this.timestamp,
    this.isError = false,
    this.tokenUsage,
    this.suggestedQuestions,
    this.cached,
  });
}
