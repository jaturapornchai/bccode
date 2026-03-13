import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:smlaicloud/global.dart' as global;

/// Public help page for account deletion instructions
/// Route: /account-deletion
/// Access: Public (no authentication required)
/// Purpose: Google Play Console data deletion policy compliance
///
/// หน้าช่วยเหลือสาธารณะสำหรับคำแนะนำการลบบัญชี
/// เข้าถึงได้โดยไม่ต้องเข้าสู่ระบบ
/// สำหรับนโยบายการลบข้อมูลใน Google Play Console
class AccountDeletionHelpScreen extends StatelessWidget {
  const AccountDeletionHelpScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      appBar: _buildAppBar(context),
      body: _buildBody(context),
    );
  }

  /// AppBar: Modern gradient design
  /// แบบเว็บไซต์สมัยใหม่ พร้อม gradient
  PreferredSizeWidget _buildAppBar(BuildContext context) {
    return AppBar(
      flexibleSpace: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [
              global.theme.primaryColor,
              global.theme.primaryLightColor,
            ],
            begin: Alignment.centerLeft,
            end: Alignment.centerRight,
          ),
        ),
      ),
      elevation: 0,
      centerTitle: true,
      automaticallyImplyLeading: false,
      title: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(
              Icons.store,
              color: global.theme.primaryColor,
              size: 24,
            ),
          ),
          SizedBox(width: 12),
          Text(
            global.language('account_deletion_app_name'),
            style: TextStyle(
              color: global.theme.onPrimaryColor,
              fontSize: 22,
              fontWeight: FontWeight.bold,
              letterSpacing: 0.5,
            ),
          ),
        ],
      ),
    );
  }

  /// Body: Scrollable centered content with modern design
  /// เนื้อหาแบบ scroll ได้ พร้อมดีไซน์สมัยใหม่
  Widget _buildBody(BuildContext context) {
    return Center(
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: _getMaxContentWidth(context),
        ),
        child: SingleChildScrollView(
          padding: EdgeInsets.symmetric(
            horizontal: _getHorizontalPadding(context),
            vertical: 48,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _buildHeroSection(context),
              const SizedBox(height: 48),
              _buildStepsSection(context),
              const SizedBox(height: 48),
              _buildScreenshotGallery(context),
              const SizedBox(height: 48),
              _buildPrivacyNote(context),
              const SizedBox(height: 48),
              _buildContactSection(context),
              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }

  /// Hero section with title and intro
  /// ส่วนหัวพร้อมหัวข้อและคำอธิบาย
  Widget _buildHeroSection(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            global.theme.cardColor,
            global.theme.infoHighlightColor,
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.08),
            blurRadius: 20,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      padding: const EdgeInsets.all(40),
      child: Column(
        children: [
          // Icon
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  global.theme.primaryColor,
                  global.theme.primaryLightColor,
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: global.theme.primaryColor.withValues(alpha: 0.3),
                  blurRadius: 15,
                  offset: const Offset(0, 5),
                ),
              ],
            ),
            child: Icon(
              Icons.delete_forever_rounded,
              color: global.theme.onPrimaryColor,
              size: 40,
            ),
          ),
          const SizedBox(height: 24),
          // Title
          Text(
            global.language('account_deletion_title'),
            style: TextStyle(
              fontSize: _getTitleFontSize(context),
              fontWeight: FontWeight.bold,
              color: global.theme.primaryColor,
              height: 1.3,
              letterSpacing: 0.5,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          // Divider
          Container(
            width: 60,
            height: 4,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  global.theme.primaryColor,
                  global.theme.primaryLightColor,
                ],
              ),
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(height: 24),
          // Intro text
          Text(
            global.language('account_deletion_intro'),
            style: TextStyle(
              fontSize: _getBodyFontSize(context) + 2,
              height: 1.8,
              color: global.theme.textColor,
              letterSpacing: 0.3,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  /// Steps section with modern card design
  /// ส่วนขั้นตอนพร้อมดีไซน์การ์ดสมัยใหม่
  Widget _buildStepsSection(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Section title
        Padding(
          padding: const EdgeInsets.only(left: 8, bottom: 24),
          child: Row(
            children: [
              Container(
                width: 6,
                height: 32,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      global.theme.primaryColor,
                      global.theme.primaryLightColor,
                    ],
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                  ),
                  borderRadius: BorderRadius.circular(3),
                ),
              ),
              const SizedBox(width: 16),
              Text(
                global.language('account_deletion_steps_title'),
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: global.theme.primaryColor,
                  letterSpacing: 0.5,
                ),
              ),
            ],
          ),
        ),
        // Steps
        _buildModernStepCard(
          context,
          1,
          global.language('account_deletion_step_1'),
          Icons.login_rounded,
          global.theme.infoHighlightTextColor,
        ),
        SizedBox(height: 20),
        _buildModernStepCard(
          context,
          2,
          global.language('account_deletion_step_2'),
          Icons.store_rounded,
          global.theme.warningHighlightTextColor,
        ),
        SizedBox(height: 20),
        _buildModernStepCard(
          context,
          3,
          global.language('account_deletion_step_3'),
          Icons.delete_rounded,
          global.theme.negativeHighlightTextColor,
        ),
      ],
    );
  }

  /// Modern step card with icon and gradient
  /// การ์ดขั้นตอนสมัยใหม่พร้อม icon และ gradient
  Widget _buildModernStepCard(
    BuildContext context,
    int stepNumber,
    String description,
    IconData icon,
    Color accentColor,
  ) {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: accentColor.withValues(alpha: 0.15),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Container(
          decoration: BoxDecoration(
            border: Border(
              left: BorderSide(
                color: accentColor,
                width: 5,
              ),
            ),
          ),
          padding: const EdgeInsets.all(24),
          child: Row(
            children: [
              // Icon with gradient background
              Container(
                width: 60,
                height: 60,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      accentColor,
                      accentColor.withValues(alpha: 0.7),
                    ],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: accentColor.withValues(alpha: 0.3),
                      blurRadius: 8,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Icon(
                  icon,
                  color: global.theme.onPrimaryColor,
                  size: 32,
                ),
              ),
              const SizedBox(width: 20),
              // Content
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${global.language("account_deletion_step")} $stepNumber',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: accentColor,
                        letterSpacing: 1,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      description,
                      style: TextStyle(
                        fontSize: _getBodyFontSize(context),
                        height: 1.6,
                        color: global.theme.textColor,
                        fontWeight: FontWeight.w500,
                      ),
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

  /// Screenshot gallery with modern card design
  /// แกลเลอรี่ภาพหน้าจอแบบการ์ดสมัยใหม่
  Widget _buildScreenshotGallery(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final columns = _getImageColumns(context);
    final horizontalPadding = _getHorizontalPadding(context);

    final availableWidth = screenWidth - (horizontalPadding * 2);
    final spacing = 24.0;
    final imageWidth = (availableWidth - (spacing * (columns - 1))) / columns;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Section title
        Padding(
          padding: const EdgeInsets.only(left: 8, bottom: 24),
          child: Row(
            children: [
              Container(
                width: 6,
                height: 32,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      global.theme.primaryColor,
                      global.theme.primaryLightColor,
                    ],
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                  ),
                  borderRadius: BorderRadius.circular(3),
                ),
              ),
              const SizedBox(width: 16),
              Text(
                'ภาพตัวอย่างขั้นตอน',
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: global.theme.primaryColor,
                  letterSpacing: 0.5,
                ),
              ),
            ],
          ),
        ),
        Wrap(
          spacing: spacing,
          runSpacing: spacing,
          alignment: WrapAlignment.center,
          children: [
            _buildModernScreenshotCard(
              context,
              'assets/img/step1_login.png',
              global.language('account_deletion_step_1'),
              imageWidth.clamp(200.0, 400.0),
              1,
            ),
            _buildModernScreenshotCard(
              context,
              'assets/img/step2_select_shop.png',
              global.language('account_deletion_step_2'),
              imageWidth.clamp(200.0, 400.0),
              2,
            ),
            _buildModernScreenshotCard(
              context,
              'assets/img/step3_delete_account.png',
              global.language('account_deletion_step_3'),
              imageWidth.clamp(200.0, 400.0),
              3,
            ),
          ],
        ),
      ],
    );
  }

  /// Modern screenshot card with hover effect
  /// การ์ดรูปภาพสมัยใหม่พร้อม hover effect
  Widget _buildModernScreenshotCard(
    BuildContext context,
    String imagePath,
    String caption,
    double width,
    int number,
  ) {
    return Container(
      width: width,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.1),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Step number badge
          Padding(
            padding: const EdgeInsets.all(16),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    global.theme.primaryColor,
                    global.theme.primaryLightColor,
                  ],
                ),
                borderRadius: BorderRadius.circular(20),
                boxShadow: [
                  BoxShadow(
                    color: global.theme.primaryColor.withValues(alpha: 0.3),
                    blurRadius: 8,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: Text(
                '${global.language("account_deletion_step")} $number',
                style: TextStyle(
                  color: global.theme.onPrimaryColor,
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  letterSpacing: 0.5,
                ),
              ),
            ),
          ),
          // Image with rounded corners
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: Container(
                decoration: BoxDecoration(
                  border: Border.all(
                    color: global.theme.dividerBorderColor,
                    width: 1,
                  ),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Image.asset(
                  imagePath,
                  width: width - 32,
                  fit: BoxFit.contain,
                ),
              ),
            ),
          ),
          const SizedBox(height: 16),
          // Caption
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 20),
            child: Text(
              caption,
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
                color: global.theme.textColor,
                height: 1.4,
              ),
              textAlign: TextAlign.center,
              maxLines: 3,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }

  /// Privacy note with modern design
  /// หมายเหตุนโยบายความเป็นส่วนตัวแบบสมัยใหม่
  Widget _buildPrivacyNote(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            global.theme.infoHighlightColor,
            global.theme.infoHighlightColor,
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: global.theme.infoHighlightTextColor.withValues(alpha: 0.3),
          width: 2,
        ),
      ),
      padding: const EdgeInsets.all(24),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: global.theme.infoHighlightTextColor,
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: global.theme.infoHighlightTextColor.withValues(alpha: 0.3),
                  blurRadius: 10,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: Icon(
              Icons.shield_rounded,
              color: global.theme.onPrimaryColor,
              size: 28,
            ),
          ),
          SizedBox(width: 20),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language('privacy'),
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: global.theme.infoHighlightTextColor,
                  ),
                ),
                SizedBox(height: 8),
                Text(
                  global.language('account_deletion_privacy_note'),
                  style: TextStyle(
                    fontSize: _getBodyFontSize(context),
                    height: 1.6,
                    color: global.theme.infoHighlightTextColor,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// Contact section with modern LINE button
  /// ส่วนติดต่อพร้อมปุ่ม LINE สมัยใหม่
  Widget _buildContactSection(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            global.theme.cardColor,
            global.theme.positiveHighlightColor,
          ],
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
        ),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.1),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      padding: const EdgeInsets.all(40),
      child: Column(
        children: [
          // Icon
          Container(
            width: 70,
            height: 70,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [
                  Color(0xFF00B900),
                  Color(0xFF00E600),
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: const Color(0xFF00B900).withValues(alpha: 0.4),
                  blurRadius: 15,
                  offset: const Offset(0, 5),
                ),
              ],
            ),
            child: Icon(
              Icons.chat_bubble_rounded,
              color: global.theme.onPrimaryColor,
              size: 36,
            ),
          ),
          const SizedBox(height: 24),
          // Title
          Text(
            global.language('account_deletion_contact_title'),
            style: TextStyle(
              fontSize: 26,
              fontWeight: FontWeight.bold,
              color: global.theme.primaryColor,
              letterSpacing: 0.5,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          // Description
          Text(
            global.language('account_deletion_contact_desc'),
            style: TextStyle(
              fontSize: _getBodyFontSize(context) + 1,
              height: 1.7,
              color: global.theme.textColor,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 20),
          // LINE ID with background
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
            decoration: BoxDecoration(
              color: const Color(0xFF00B900).withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(30),
              border: Border.all(
                color: const Color(0xFF00B900).withValues(alpha: 0.3),
                width: 2,
              ),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(
                  Icons.alternate_email,
                  color: Color(0xFF00B900),
                  size: 20,
                ),
                SizedBox(width: 8),
                Text(
                  global.language('account_deletion_line_id'),
                  style: const TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF00B900),
                    letterSpacing: 0.5,
                  ),
                ),
              ],
            ),
          ),
          SizedBox(height: 12),
          Text(
            global.language('account_deletion_line_topic'),
            style: TextStyle(
              fontSize: _getBodyFontSize(context),
              color: global.theme.textSecondaryColor,
              fontStyle: FontStyle.italic,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 32),
          // LINE buttons with modern design
          Wrap(
            spacing: 16,
            runSpacing: 16,
            alignment: WrapAlignment.center,
            children: [
              // Primary button: Open LINE
              Container(
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [
                      Color(0xFF00B900),
                      Color(0xFF00E600),
                    ],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(30),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF00B900).withValues(alpha: 0.4),
                      blurRadius: 15,
                      offset: const Offset(0, 6),
                    ),
                  ],
                ),
                child: ElevatedButton.icon(
                  onPressed: () => _openLineOrCopyId(context),
                  icon: Icon(Icons.chat_bubble_rounded, size: 22),
                  label: Text(
                    global.language('account_deletion_open_line'),
                    style: const TextStyle(fontSize: 16, letterSpacing: 0.5),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.transparent,
                    foregroundColor: global.theme.onPrimaryColor,
                    shadowColor: Colors.transparent,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 40,
                      vertical: 20,
                    ),
                    textStyle: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(30),
                    ),
                  ),
                ),
              ),
              // Secondary button: Copy LINE ID
              Container(
                decoration: BoxDecoration(
                  color: global.theme.cardColor,
                  borderRadius: BorderRadius.circular(30),
                  border: Border.all(
                    color: const Color(0xFF00B900),
                    width: 2,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.08),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: ElevatedButton.icon(
                  onPressed: () => _copyLineId(context),
                  icon: Icon(Icons.content_copy_rounded, size: 20),
                  label: Text(
                    global.language('account_deletion_copy_line_id'),
                    style: const TextStyle(fontSize: 16, letterSpacing: 0.5),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: global.theme.cardColor,
                    foregroundColor: const Color(0xFF00B900),
                    shadowColor: Colors.transparent,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 40,
                      vertical: 20,
                    ),
                    textStyle: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(30),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // ========== RESPONSIVE HELPER FUNCTIONS ==========
  // ========== ฟังก์ชันช่วยเหลือสำหรับ Responsive Design ==========

  /// Get max content width based on screen size
  /// คำนวณความกว้างสูงสุดของเนื้อหาตามขนาดหน้าจอ
  double _getMaxContentWidth(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    if (width < 600) return width * 0.95; // Mobile: 95% width
    if (width < 900) return 700; // Tablet: fixed 700px
    return 1100; // Desktop: fixed 1100px
  }

  /// Get horizontal padding based on screen size
  /// คำนวณ padding ซ้าย-ขวาตามขนาดหน้าจอ
  double _getHorizontalPadding(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    if (width < 600) return 16.0;
    if (width < 900) return 40.0;
    return 60.0;
  }

  /// Get number of image columns based on screen size
  /// คำนวณจำนวนคอลัมน์ของรูปภาพตามขนาดหน้าจอ
  int _getImageColumns(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    if (width < 600) return 1; // Mobile: 1 column
    if (width < 900) return 2; // Tablet: 2 columns
    return 3; // Desktop: 3 columns
  }

  /// Get title font size based on screen size
  /// คำนวณขนาดฟอนต์หัวข้อตามขนาดหน้าจอ
  double _getTitleFontSize(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    if (width < 600) return 26.0;
    if (width < 900) return 34.0;
    return 42.0;
  }

  /// Get body font size based on screen size
  /// คำนวณขนาดฟอนต์เนื้อหาตามขนาดหน้าจอ
  double _getBodyFontSize(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    if (width < 600) return 14.0;
    return 16.0;
  }

  // ========== INTERACTIVE FUNCTIONS ==========
  // ========== ฟังก์ชันสำหรับการโต้ตอบ ==========

  /// Try to open LINE app/web, fallback to copy ID
  /// พยายามเปิด LINE app หรือเว็บ ถ้าไม่ได้ก็ copy ID
  Future<void> _openLineOrCopyId(BuildContext context) async {
    const lineId = '@bcaccount';

    try {
      // For web, try to open LINE web URL
      if (kIsWeb) {
        final lineUrl = Uri.parse('https://line.me/R/ti/p/$lineId');
        if (await canLaunchUrl(lineUrl)) {
          await launchUrl(lineUrl, mode: LaunchMode.externalApplication);
          return;
        }
      } else {
        // For mobile platforms, try LINE app deep link
        final lineUrl = Uri.parse('line://ti/p/$lineId');
        if (await canLaunchUrl(lineUrl)) {
          await launchUrl(lineUrl);
          return;
        }
      }
    } catch (e) {
      // If any error occurs, fall through to copy
    }

    // Fallback: copy to clipboard
    if (context.mounted) {
      await _copyLineId(context);
    }
  }

  /// Copy LINE ID to clipboard and show confirmation
  /// คัดลอก LINE ID ไปยัง clipboard และแสดงข้อความยืนยัน
  Future<void> _copyLineId(BuildContext context) async {
    await Clipboard.setData(const ClipboardData(text: '@bcaccount'));

    if (context.mounted) {
      global.showSuccessSnackBar(context, global.language('account_deletion_copied'));
    }
  }
}
