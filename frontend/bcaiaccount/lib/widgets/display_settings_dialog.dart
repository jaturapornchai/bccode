import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:smlaicloud/global.dart' as global;

/// Helper function เพื่อ get TextStyle จาก Google Fonts (10 Popular Thai Fonts)
TextStyle _getGoogleFontStyle(String fontFamily, {double? fontSize, Color? color, FontWeight? fontWeight}) {
  switch (fontFamily) {
    case 'Sarabun':
      return GoogleFonts.sarabun(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'Kanit':
      return GoogleFonts.kanit(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'Prompt':
      return GoogleFonts.prompt(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'Noto Sans Thai':
      return GoogleFonts.notoSansThai(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'IBM Plex Sans Thai':
      return GoogleFonts.ibmPlexSansThai(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'Mitr':
      return GoogleFonts.mitr(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'K2D':
      return GoogleFonts.k2d(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'Bai Jamjuree':
      return GoogleFonts.baiJamjuree(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'Pridi':
      return GoogleFonts.pridi(fontSize: fontSize, color: color, fontWeight: fontWeight);
    case 'Sriracha':
      return GoogleFonts.sriracha(fontSize: fontSize, color: color, fontWeight: fontWeight);
    default:
      return GoogleFonts.sarabun(fontSize: fontSize, color: color, fontWeight: fontWeight);
  }
}

/// Dialog สำหรับตั้งค่าการแสดงผล
/// - Font Family
/// - Zoom Level (50% - 300%)
class DisplaySettingsDialog extends StatefulWidget {
  final VoidCallback? onSettingsChanged;

  const DisplaySettingsDialog({
    super.key,
    this.onSettingsChanged,
  });

  @override
  State<DisplaySettingsDialog> createState() => _DisplaySettingsDialogState();
}

class _DisplaySettingsDialogState extends State<DisplaySettingsDialog> {
  late String _selectedFontFamily;
  late double _zoomLevel;

  @override
  void initState() {
    super.initState();
    _selectedFontFamily = global.displayFontFamily;
    _zoomLevel = global.displayZoomLevel;
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          Icon(Icons.settings, color: global.theme.primaryColor),
          SizedBox(width: 8),
          Text(global.language("display_settings")),
        ],
      ),
      content: SizedBox(
        width: 400,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
            // Font Family Section
            const Text(
              'Font Family',
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              decoration: BoxDecoration(
                border: Border.all(color: Colors.grey.shade300),
                borderRadius: BorderRadius.circular(8),
              ),
              child: DropdownButton<String>(
                value: _selectedFontFamily,
                isExpanded: true,
                underline: const SizedBox(),
                items: global.availableFontFamilies.map((font) {
                  return DropdownMenuItem<String>(
                    value: font,
                    child: Text(
                      font,
                      style: _getGoogleFontStyle(font, fontSize: 14),
                    ),
                  );
                }).toList(),
                onChanged: (value) {
                  if (value != null) {
                    setState(() {
                      _selectedFontFamily = value;
                    });
                  }
                },
              ),
            ),

            const SizedBox(height: 24),

            // Zoom Level Section
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Zoom Level',
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  decoration: BoxDecoration(
                    color: global.theme.primaryColor.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Text(
                    '${_zoomLevel.toInt()}%',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: global.theme.primaryColor,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                const Text('50%', style: TextStyle(fontSize: 12, color: Colors.grey)),
                Expanded(
                  child: Slider(
                    value: _zoomLevel,
                    min: 50,
                    max: 300,
                    divisions: 25, // ทุก 10%
                    activeColor: global.theme.primaryColor,
                    onChanged: (value) {
                      setState(() {
                        _zoomLevel = value;
                      });
                    },
                  ),
                ),
                const Text('300%', style: TextStyle(fontSize: 12, color: Colors.grey)),
              ],
            ),

            // Quick Zoom Buttons - ใช้ Wrap แทน Row เพื่อป้องกัน overflow
            const SizedBox(height: 8),
            Wrap(
              alignment: WrapAlignment.center,
              spacing: 4,
              runSpacing: 4,
              children: [
                _buildQuickZoomButton(50),
                _buildQuickZoomButton(75),
                _buildQuickZoomButton(100),
                _buildQuickZoomButton(150),
                _buildQuickZoomButton(200),
              ],
            ),

            const SizedBox(height: 16),

            // Preview Section
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.grey.shade100,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.grey.shade300),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${global.language("preview")}:',
                    style: const TextStyle(fontSize: 12, color: Colors.grey),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    global.language("font_preview_text"),
                    style: _getGoogleFontStyle(
                      _selectedFontFamily,
                      fontSize: 16 * (_zoomLevel / 100),
                    ),
                  ),
                  Text(
                    'ระบบบัญชี Cloud ERP',
                    style: _getGoogleFontStyle(
                      _selectedFontFamily,
                      fontSize: 14 * (_zoomLevel / 100),
                      color: Colors.grey.shade600,
                    ),
                  ),
                ],
              ),
            ),
          ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () {
            // Reset to defaults
            setState(() {
              _selectedFontFamily = 'Sarabun';
              _zoomLevel = 100.0;
            });
          },
          child: Text(global.language("reset")),
        ),
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(global.language("cancel")),
        ),
        ElevatedButton(
          style: ElevatedButton.styleFrom(
            backgroundColor: global.theme.primaryColor,
          ),
          onPressed: () async {
            // Save settings
            global.displayFontFamily = _selectedFontFamily;
            global.displayZoomLevel = _zoomLevel;
            await global.saveDisplaySettings();

            // Notify parent
            widget.onSettingsChanged?.call();

            if (context.mounted) {
              Navigator.of(context).pop(true);
            }
          },
          child: Text(global.language("save"), style: TextStyle(color: Colors.white)),
        ),
      ],
    );
  }

  Widget _buildQuickZoomButton(int zoom) {
    final isSelected = _zoomLevel.toInt() == zoom;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: InkWell(
        onTap: () {
          setState(() {
            _zoomLevel = zoom.toDouble();
          });
        },
        borderRadius: BorderRadius.circular(16),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: isSelected ? global.theme.primaryColor : Colors.grey.shade200,
            borderRadius: BorderRadius.circular(16),
          ),
          child: Text(
            '$zoom%',
            style: TextStyle(
              fontSize: 12,
              color: isSelected ? Colors.white : Colors.grey.shade700,
              fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
            ),
          ),
        ),
      ),
    );
  }
}

/// แสดง dialog ตั้งค่าการแสดงผล
Future<bool?> showDisplaySettingsDialog(
  BuildContext context, {
  VoidCallback? onSettingsChanged,
}) {
  return showDialog<bool>(
    context: context,
    builder: (context) => DisplaySettingsDialog(
      onSettingsChanged: onSettingsChanged,
    ),
  );
}
