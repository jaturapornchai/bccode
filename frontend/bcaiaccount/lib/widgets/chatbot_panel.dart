import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_inappwebview/flutter_inappwebview.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:uuid/uuid.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Chatbot Panel Widget - สำหรับฝังในหน้าจอต่างๆ (รองรับ HTML response)
class ChatbotPanel extends StatefulWidget {
  final String? title;
  final String? welcomeMessage;

  const ChatbotPanel({super.key, this.title, this.welcomeMessage});

  @override
  State<ChatbotPanel> createState() => _ChatbotPanelState();
}

class _ChatbotPanelState extends State<ChatbotPanel>
    with global.ThemeRefreshMixin {
  final TextEditingController _messageController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final List<ChatMessage> _messages = [];
  bool _isLoading = false;
  late String _sessionId;

  // สีธีม
  Color get primaryColor => global.theme.primaryColor;
  Color get primaryDarkColor => global.theme.primaryColor;
  Color get aiMessageColor => global.theme.infoHighlightColor;
  Color get userMessageColor => global.theme.primaryColor;

  @override
  void initState() {
    super.initState();
    _sessionId = const Uuid().v4();

    // ข้อความต้อนรับ
    _messages.add(
      ChatMessage(
        text:
            widget.welcomeMessage ??
            global.language("chatbot_welcome_message"),
        isUser: false,
        timestamp: DateTime.now(),
        isHtml: false,
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

    try {
      final result = await _callChatAPI(message);

      setState(() {
        _messages.add(
          ChatMessage(
            text: result['text'],
            isUser: false,
            timestamp: DateTime.now(),
            isHtml: true, // HTML response from /api/chat-html
            aiUsage: result['ai_usage'],
          ),
        );
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _messages.add(
          ChatMessage(
            text: '${global.language("chatbot_error_message")} $e',
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
    final shopId = global.getShopId();

    final baseUrl = kDebugMode
        ? 'http://localhost:9999'
        : 'https://bcaicallcenter.dedetouch.com';
    final url = Uri.parse('$baseUrl/api/chat-html');

    if (kDebugMode) {
      AppLogger.debug('Sending chat request to: ${url.toString()}');
      AppLogger.debug('Shop ID: $shopId');
      AppLogger.debug('Session ID: $_sessionId');
      AppLogger.debug('Message: $message');
    }

    try {
      final response = await http
          .post(
            url,
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({
              'shop_id': shopId,
              'session_id': _sessionId,
              'message': message,
            }),
          )
          .timeout(
            const Duration(seconds: 90),
            onTimeout: () {
              throw Exception('Connection timeout');
            },
          );

      if (kDebugMode) {
        AppLogger.debug('Response status: ${response.statusCode}');
        AppLogger.debug('Response body length: ${response.body.length} chars');
      }

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);

        String responseText = '';
        Map<String, dynamic>? aiUsage;

        if (data is Map) {
          // ดึง ai_usage ถ้ามี
          if (data.containsKey('ai_usage')) {
            aiUsage = data['ai_usage'] as Map<String, dynamic>?;
          }

          // ลำดับความสำคัญ: html > response_html > response > response_text
          if (data.containsKey('html')) {
            responseText = data['html'].toString();
          } else if (data.containsKey('response_html')) {
            responseText = data['response_html'].toString();
          } else if (data.containsKey('response')) {
            responseText = data['response'].toString();
          } else if (data.containsKey('response_text')) {
            responseText = data['response_text'].toString();
          } else {
            responseText = response.body;
          }
        } else {
          responseText = response.body;
        }

        return {'text': responseText, 'ai_usage': aiUsage};
      } else {
        throw Exception('HTTP ${response.statusCode}');
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Chat API Error: $e');
      }
      rethrow;
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
    return Container(
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        border: Border(left: BorderSide(color: global.theme.dividerBorderColor, width: 1)),
      ),
      child: Column(
        children: [
          _buildHeader(),
          Expanded(child: _buildChatMessages()),
          _buildChatInput(),
        ],
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: primaryDarkColor,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.1),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: global.theme.cardColor.withValues(alpha: 0.2),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(Icons.smart_toy, size: 24, color: global.theme.onPrimaryColor),
          ),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  widget.title ?? 'AI Assistant',
                  style: TextStyle(
                    color: global.theme.onPrimaryColor,
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  global.language("chatbot_ready_to_help"),
                  style: TextStyle(color: global.theme.onPrimaryColor.withValues(alpha: 0.7), fontSize: 12),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChatMessages() {
    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.all(16),
      itemCount: _messages.length + (_isLoading ? 1 : 0),
      itemBuilder: (context, index) {
        if (index == _messages.length && _isLoading) {
          return _buildTypingIndicator();
        }
        return _buildMessageBubble(_messages[index]);
      },
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
            Container(
              margin: const EdgeInsets.only(right: 8),
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: primaryColor.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(Icons.smart_toy, size: 20, color: primaryColor),
            ),
          ],
          Flexible(
            child: Container(
              padding: message.isHtml
                  ? EdgeInsets.zero
                  : const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: message.isUser
                    ? userMessageColor
                    : (message.isError
                          ? global.theme.negativeHighlightColor
                          : (message.isHtml ? global.theme.cardColor : aiMessageColor)),
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(16),
                  topRight: const Radius.circular(16),
                  bottomLeft: Radius.circular(message.isUser ? 16 : 4),
                  bottomRight: Radius.circular(message.isUser ? 4 : 16),
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.05),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: message.isHtml
                  ? _HtmlContentView(
                      htmlContent: message.text,
                      onSendMessage: (msg) {
                        // ส่งข้อความไปหา chatbot เมื่อกดปุ่มใน HTML
                        _messageController.text = msg;
                        _sendMessage();
                      },
                    )
                  : Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          message.text,
                          style: TextStyle(
                            fontSize: 15,
                            color: message.isUser
                                ? global.theme.cardColor
                                : (message.isError
                                      ? global.theme.negativeHighlightTextColor
                                      : global.theme.textColor),
                            height: 1.4,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          _formatTime(message.timestamp),
                          style: TextStyle(
                            fontSize: 11,
                            color: message.isUser
                                ? Colors.white70
                                : global.theme.textSecondaryColor,
                          ),
                        ),
                      ],
                    ),
            ),
          ),
          if (message.isUser) ...[
            Container(
              margin: const EdgeInsets.only(left: 8),
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: userMessageColor.withValues(alpha: 0.2),
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
      padding: EdgeInsets.only(bottom: 12),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          Container(
            margin: const EdgeInsets.only(right: 8),
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: primaryColor.withValues(alpha: 0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.smart_toy, size: 20, color: primaryColor),
          ),
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
            child: Text(global.language("chatbot_typing")),
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

  Widget _buildChatInput() {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.1),
            blurRadius: 4,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _messageController,
              decoration: InputDecoration(
                hintText: global.language("chatbot_type_message"),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: BorderSide(color: global.theme.formBorderColor),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: BorderSide(color: global.theme.formBorderColor),
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
                fillColor: global.theme.formFillColor,
              ),
              maxLines: null,
              textInputAction: TextInputAction.send,
              onSubmitted: (_) => _sendMessage(),
            ),
          ),
          const SizedBox(width: 12),
          Container(
            decoration: BoxDecoration(
              color: primaryColor,
              shape: BoxShape.circle,
            ),
            child: IconButton(
              icon: Icon(Icons.send, color: global.theme.onPrimaryColor),
              onPressed: _isLoading ? null : _sendMessage,
            ),
          ),
        ],
      ),
    );
  }
}

// Models
class ChatMessage {
  final String text;
  final bool isUser;
  final DateTime timestamp;
  final bool isError;
  final bool isHtml;
  final Map<String, dynamic>? aiUsage;

  ChatMessage({
    required this.text,
    required this.isUser,
    required this.timestamp,
    this.isError = false,
    this.isHtml = false,
    this.aiUsage,
  });
}

/// Widget สำหรับแสดง HTML Content
class _HtmlContentView extends StatefulWidget {
  final String htmlContent;
  final Function(String)? onSendMessage;

  const _HtmlContentView({required this.htmlContent, this.onSendMessage});

  @override
  State<_HtmlContentView> createState() => _HtmlContentViewState();
}

class _HtmlContentViewState extends State<_HtmlContentView>
    with global.ThemeRefreshMixin {
  double _contentHeight = 400; // เริ่มต้นที่ 400 แทน 100
  static const double _maxHeight = 2000; // จำกัดความสูงสูงสุด

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minHeight: 100, maxHeight: _maxHeight),
      height: _contentHeight,
      child: InAppWebView(
        initialData: InAppWebViewInitialData(
          data:
              '''
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no">
    <style>
        body { 
            margin: 0; 
            padding: 8px; 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            font-size: 14px;
            line-height: 1.5;
            overflow-x: hidden;
        }
        * { 
            box-sizing: border-box;
            max-width: 100%;
        }
        img {
            max-width: 100% !important;
            height: auto !important;
        }
    </style>
</head>
<body>
${widget.htmlContent}
</body>
</html>
          ''',
        ),
        initialSettings: InAppWebViewSettings(
          transparentBackground: true,
          javaScriptEnabled: true,
          supportZoom: false,
          useShouldOverrideUrlLoading: true,
          cacheEnabled: false,
          javaScriptCanOpenWindowsAutomatically: false,
          mediaPlaybackRequiresUserGesture: true,
          allowsInlineMediaPlayback: true,
          isFraudulentWebsiteWarningEnabled: false,
        ),
        onWebViewCreated: (controller) async {
          if (kDebugMode) {
            AppLogger.debug('WebView created for HTML content');
          }

          // เพิ่ม JavaScript handlers เหมือน chatbot_overlay.dart
          controller.addJavaScriptHandler(
            handlerName: 'flutterHandler',
            callback: (args) {
              try {
                if (kDebugMode) {
                  AppLogger.debug('Flutter handler called with: $args');
                }
              } catch (e) {
                if (kDebugMode) {
                  AppLogger.error('Error in flutterHandler: $e');
                }
              }
            },
          );

          controller.addJavaScriptHandler(
            handlerName: 'onProductProperties',
            callback: (args) {
              try {
                if (mounted && args.isNotEmpty) {
                  final data = args[0];
                  if (kDebugMode) {
                    AppLogger.debug('Product properties received: $data');
                  }

                  // ส่งข้อมูลสินค้ากลับไปถาม AI
                  if (data is Map) {
                    final name = data['name'] ?? '';
                    final code = data['code'] ?? '';
                    final barcode = data['barcode'] ?? '';
                    final unit = data['unit'] ?? '';

                    // สร้างข้อความถาม AI
                    final message =
                        'วิเคราะห์สินค้านี้: '
                        'ชื่อ: $name, '
                        'รหัส: $code, '
                        'บาร์โค้ด: $barcode, '
                        'หน่วย: $unit';

                    if (kDebugMode) {
                      AppLogger.debug('Sending message: $message');
                    }

                    // ส่งข้อความไปยัง chatbot ผ่าน callback
                    widget.onSendMessage?.call(message);
                  }
                }
              } catch (e) {
                if (kDebugMode) {
                  AppLogger.error('Error in onProductProperties: $e');
                }
              }
            },
          );

          controller.addJavaScriptHandler(
            handlerName: 'onConsoleMessage',
            callback: (args) {
              try {
                if (kDebugMode && args.isNotEmpty) {
                  AppLogger.debug('Console message: $args');
                }
              } catch (e) {
                // Ignore
              }
            },
          );

          // เพิ่ม JavaScript error handler
          try {
            await controller.evaluateJavascript(
              source: '''
              window.onerror = function(msg, url, lineNo, columnNo, error) {
                console.log('JS Error: ' + msg);
                return true;
              };
              
              // ป้องกัน navigation
              document.addEventListener('click', function(e) {
                var target = e.target;
                while(target && target.tagName !== 'A') {
                  target = target.parentElement;
                }
                if(target && target.tagName === 'A') {
                  e.preventDefault();
                  e.stopPropagation();
                  console.log('Link click prevented: ' + target.href);
                  return false;
                }
              }, true);
            ''',
            );
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error('Error setting up WebView handlers: $e');
            }
          }
        },
        shouldOverrideUrlLoading: (controller, navigationAction) async {
          final uri = navigationAction.request.url;

          if (kDebugMode) {
            AppLogger.debug('Navigation attempted to: $uri');
          }

          // ป้องกันการ navigate ไปหน้าอื่น
          return NavigationActionPolicy.CANCEL;
        },
        onLoadStop: (controller, url) async {
          try {
            // รอให้ DOM โหลดเสร็จสมบูรณ์
            await Future.delayed(const Duration(milliseconds: 100));

            if (!mounted) return;

            // คำนวณความสูงของ content
            final heightResult = await controller.evaluateJavascript(
              source: 'document.body.scrollHeight',
            );

            if (heightResult != null && mounted) {
              final calculatedHeight = (heightResult as num).toDouble();
              // จำกัดความสูงไม่ให้เกิน max และไม่ต่ำกว่า min
              final clampedHeight =
                  calculatedHeight.clamp(100.0, _maxHeight) + 20;

              if (kDebugMode) {
                AppLogger.debug(
                  'HTML content height: $calculatedHeight -> clamped to: $clampedHeight',
                );
              }

              setState(() {
                _contentHeight = clampedHeight;
              });
            }
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error('Error calculating HTML height: $e');
            }
            // ใช้ความสูงเริ่มต้นถ้าเกิด error
            if (mounted) {
              setState(() {
                _contentHeight = 400;
              });
            }
          }
        },
        onConsoleMessage: (controller, consoleMessage) {
          // กรอง console log
          if (kDebugMode) {
            final msg = consoleMessage.message;
            // ไม่แสดง Object และ error ที่ไม่สำคัญ
            if (!msg.contains('Object') &&
                !msg.contains('Failed to load') &&
                !msg.contains('net::ERR')) {
              AppLogger.debug('HTML Console: $msg');
            }
          }
        },
        onReceivedError: (controller, request, error) {
          if (kDebugMode) {
            AppLogger.error('WebView Error: ${error.description}');
          }
          // ไม่ให้ crash แค่แสดง error
        },
        onReceivedHttpError: (controller, request, errorResponse) {
          if (kDebugMode) {
            AppLogger.error('HTTP Error: ${errorResponse.statusCode}');
          }
        },
        onRenderProcessGone: (controller, detail) {
          if (kDebugMode) {
            AppLogger.debug('WebView Render Process Gone: ${detail.didCrash}');
          }
          // ป้องกัน crash โดยไม่ทำอะไร
          return;
        },
      ),
    );
  }
}
