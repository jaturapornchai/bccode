import 'dart:async';
import 'dart:math';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';

// วิดเจ็ตพื้นหลังเมฆเคลื่อนไหว
class CloudBackground extends StatefulWidget {
  final Color backgroundColor1; // สีเริ่มต้นของท้องฟ้าด้านบน
  final Color backgroundColor2; // สีเริ่มต้นของท้องฟ้าด้านล่าง
  final int initialCloudCount; // จำนวนเมฆเริ่มต้น
  final int cloudCreateInterval; // ระยะเวลาในการสร้างเมฆใหม่ (วินาที)
  final double cloudOpacity; // ความโปร่งใสของเมฆ
  final double maxCloudSpeed; // ความเร็วสูงสุดของเมฆ
  final int layerCount; // จำนวนชั้นของเมฆ (เพื่อความลึก)

  const CloudBackground({
    super.key,
    this.backgroundColor1 = const Color(0xFF7EB4E2), // สีฟ้าอ่อนลง
    this.backgroundColor2 = const Color(0xFFE6F3FE), // สีฟ้าขาวมากขึ้น
    this.initialCloudCount = 4, // เพิ่มจำนวนเมฆเริ่มต้น
    this.cloudCreateInterval = 10,
    this.cloudOpacity = 0.9, // ลดความทึบลงเล็กน้อย
    this.maxCloudSpeed = 0.6, // ปรับความเร็วให้เหมาะสม
    this.layerCount = 3, // จำนวนชั้นเพื่อสร้างความลึก
  });

  @override
  State<CloudBackground> createState() => _CloudBackgroundState();
}

class _CloudBackgroundState extends State<CloudBackground> {
  final List<Cloud> clouds = [];
  final Random random = Random();
  late Timer cloudTimer;
  int nextFlockId = 0; // รหัสฝูงนกถัดไปที่จะสร้าง

  @override
  void initState() {
    super.initState();
    // สร้างเมฆและนกเริ่มต้น
    WidgetsBinding.instance.addPostFrameCallback((_) {
      // สร้างเมฆในหลายชั้นเพื่อเพิ่มความลึก
      for (int layer = 0; layer < widget.layerCount; layer++) {
        // คำนวณค่าการกระจายตามชั้น
        double layerScale = 1.0 - (0.2 * layer); // ขนาดลดลงตามชั้นที่ไกลออกไป
        double layerOpacity = widget.cloudOpacity *
            (0.7 + (0.3 * layer / widget.layerCount)); // ความโปร่งใสแตกต่างกัน
        double layerSpeed = widget.maxCloudSpeed *
            (0.3 + 0.7 * layer / widget.layerCount); // ชั้นไกลเคลื่อนที่ช้ากว่า

        // สร้างเมฆในแต่ละชั้น
        for (int i = 0; i < widget.initialCloudCount; i++) {
          _addCloudWithLayer(layer, layerScale, layerOpacity, layerSpeed);
        }
      }

      // ตั้งเวลาเพื่อเพิ่มเมฆใหม่
      cloudTimer = Timer.periodic(Duration(seconds: widget.cloudCreateInterval),
          (timer) {
        if (mounted) {
          setState(() {
            // สร้างเมฆใหม่ในแต่ละชั้น
            int randomLayer = random.nextInt(widget.layerCount);
            double layerScale = 1.0 - (0.2 * randomLayer);
            double layerOpacity = widget.cloudOpacity *
                (0.7 + (0.3 * randomLayer / widget.layerCount));
            double layerSpeed = widget.maxCloudSpeed *
                (0.3 + 0.7 * randomLayer / widget.layerCount);

            _addCloudWithLayer(
                randomLayer, layerScale, layerOpacity, layerSpeed);

            // ลบเมฆที่ออกไปนอกจอแล้ว
            clouds.removeWhere((cloud) =>
                cloud.position.dx > MediaQuery.of(context).size.width + 100);
          });
        }
      });
    });
  }

  @override
  void dispose() {
    cloudTimer.cancel();
    super.dispose();
  }

  // เพิ่มเมฆโดยระบุชั้น สเกล ความโปร่งใส และความเร็ว
  void _addCloudWithLayer(
      int layer, double scale, double opacity, double maxSpeed) {
    // ขนาดเมฆตามระยะชั้น (เมฆไกลจะเล็กกว่า)
    final baseSize =
        120.0 + random.nextDouble() * 240; // เพิ่มขนาดพื้นฐานให้ใหญ่ขึ้น 100%
    final size = baseSize * scale;

    // ความเร็วตามระยะชั้น (เมฆไกลจะช้ากว่า)
    final speed = 0.3 + random.nextDouble() * maxSpeed;

    // สร้างตำแหน่ง Y ที่กระจายตามชั้น
    double maxHeight = MediaQuery.of(context).size.height;
    double sectionHeight = maxHeight / widget.layerCount;
    double baseY = layer * sectionHeight;
    final yPosition = baseY + random.nextDouble() * sectionHeight;

    // ตำแหน่ง X เริ่มต้นสุ่มบ้างเพื่อไม่ให้เมฆเข้ามาพร้อมกันทั้งหมด
    final xPosition = -size - random.nextDouble() * 200;

    // สุ่มรูปแบบเมฆ (เพิ่มความหลากหลาย)
    final cloudStyle = random.nextInt(3); // 0, 1, หรือ 2

    clouds.add(Cloud(
      size: size,
      speed: speed,
      position: Offset(xPosition, yPosition),
      opacity: opacity,
      layer: layer,
      style: cloudStyle, // ใช้รูปแบบเมฆที่หลากหลาย
    ));
  }

  @override
  Widget build(BuildContext context) {
    return Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [widget.backgroundColor1, widget.backgroundColor2],
          ),
        ),
        child: AnimatedClouds(clouds: clouds));
  }
}

// คลาสเก็บข้อมูลของเมฆแต่ละก้อน
class Cloud {
  double size;
  double speed;
  Offset position;
  double opacity;
  int layer; // ชั้นของเมฆ (เพื่อความลึก)
  int style; // เพิ่มประเภทของรูปแบบเมฆ

  Cloud({
    required this.size,
    required this.speed,
    required this.position,
    this.opacity = 1.0,
    this.layer = 0,
    required this.style, // เก็บรูปแบบของเมฆ
  });
}

// วิดเจ็ตสำหรับแสดงเมฆเคลื่อนไหว
class AnimatedClouds extends StatefulWidget {
  final List<Cloud> clouds;

  const AnimatedClouds({super.key, required this.clouds});

  @override
  State<AnimatedClouds> createState() => _AnimatedCloudsState();
}

class _AnimatedCloudsState extends State<AnimatedClouds>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    // สร้าง animation controller ที่ทำงานต่อเนื่อง
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 1),
    )..repeat();

    _controller.addListener(() {
      if (!mounted) return;

      // อัพเดตตำแหน่งของเมฆทุกๆ เฟรม
      for (var cloud in widget.clouds) {
        cloud.position = Offset(
          cloud.position.dx + cloud.speed,
          cloud.position.dy,
        );
      }
      if (mounted) {
        setState(() {});
      }
    });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: widget.clouds.map((cloud) {
        return Positioned(
          left: cloud.position.dx,
          top: cloud.position.dy,
          child: CloudWidget(
            size: cloud.size,
            opacity: cloud.opacity,
            style: cloud.style,
          ),
        );
      }).toList(),
    );
  }
}

// วิดเจ็ตสำหรับวาดรูปเมฆ
class CloudWidget extends StatelessWidget {
  final double size;
  final double opacity;
  final int style;

  const CloudWidget({
    super.key,
    required this.size,
    this.opacity = 1.0,
    required this.style,
  });

  @override
  Widget build(BuildContext context) {
    return Opacity(
      opacity: opacity,
      child: SizedBox(
        width: size,
        // ปรับสัดส่วนให้กว้างขึ้น
        height: size * 0.55,
        child: CustomPaint(
          painter: CloudPainter(style: style),
          foregroundPainter: CloudHighlightPainter(),
        ),
      ),
    );
  }
}

// Painter สำหรับวาดไฮไลต์บนเมฆ
class CloudHighlightPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.white.withOpacity(0.4)
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 5)
      ..blendMode = BlendMode.screen;

    // วาดไฮไลต์บริเวณด้านบนของเมฆ
    canvas.drawOval(
      Rect.fromLTWH(
        size.width * 0.25,
        size.height * 0.1,
        size.width * 0.5,
        size.height * 0.3,
      ),
      paint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) {
    return false;
  }
}

// Painter สำหรับวาดรูปเมฆที่สวยงามขึ้น
class CloudPainter extends CustomPainter {
  final int style;

  CloudPainter({required this.style});

  @override
  void paint(Canvas canvas, Size size) {
    // เลือกฟังก์ชันวาดตามสไตล์
    switch (style % 3) {
      case 0:
        _drawCloudStyle1(canvas, size);
        break;
      case 1:
        _drawCloudStyle2(canvas, size);
        break;
      case 2:
        _drawCloudStyle3(canvas, size);
        break;
    }
  }

  // สไตล์เมฆแบบที่ 1 - เมฆนุ่มนวลและมีขอบโค้งมน
  void _drawCloudStyle1(Canvas canvas, Size size) {
    // สร้าง gradient สำหรับเมฆเพื่อให้ดูมีมิติ
    final gradientPaint = Paint()
      ..shader = ui.Gradient.linear(
        Offset(size.width / 2, 0),
        Offset(size.width / 2, size.height),
        [
          Colors.white,
          Colors.white.withOpacity(0.85),
        ],
      )
      ..style = PaintingStyle.fill;

    // เพิ่ม blur effect เพื่อให้ขอบเมฆนุ่มขึ้น
    final shadowPaint = Paint()
      ..color = Colors.white.withOpacity(0.6)
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 8);

    // ใช้ Path แทนการวาดวงกลมเพื่อให้มีรูปร่างที่ซับซ้อนและสวยงามขึ้น
    final path = Path();

    final width = size.width;
    final height = size.height;

    // สร้างเส้นโค้งที่ซับซ้อนขึ้นเพื่อให้เมฆดูเป็นธรรมชาติ
    path.moveTo(width * 0.2, height * 0.7);

    // โค้งด้านล่างซ้าย
    path.quadraticBezierTo(
        width * 0.05, height * 0.7, width * 0.1, height * 0.5);

    // โค้งซ้าย
    path.quadraticBezierTo(
        width * 0.0, height * 0.25, width * 0.25, height * 0.25);

    // โค้งบนซ้าย
    path.quadraticBezierTo(
        width * 0.25, height * 0.05, width * 0.4, height * 0.15);

    // โค้งบนกลาง
    path.quadraticBezierTo(
        width * 0.5, height * 0.0, width * 0.6, height * 0.15);

    // โค้งบนขวา
    path.quadraticBezierTo(
        width * 0.75, height * 0.05, width * 0.75, height * 0.25);

    // โค้งขวา
    path.quadraticBezierTo(
        width * 1.0, height * 0.25, width * 0.9, height * 0.5);

    // โค้งล่างขวา
    path.quadraticBezierTo(
        width * 0.95, height * 0.7, width * 0.8, height * 0.7);

    // ปิดเส้น
    path.lineTo(width * 0.2, height * 0.7);

    // วาดเงาก่อนเพื่อให้มี blur effect
    canvas.drawPath(path, shadowPaint);

    // วาดเมฆด้วย gradient
    canvas.drawPath(path, gradientPaint);
  }

  // สไตล์เมฆแบบที่ 2 - กลุ่มก้อนวงกลมซ้อนกัน
  void _drawCloudStyle2(Canvas canvas, Size size) {
    final paint = Paint()
      ..shader = ui.Gradient.linear(
        Offset(size.width / 2, 0),
        Offset(size.width / 2, size.height),
        [
          Colors.white,
          Colors.white.withOpacity(0.8),
        ],
      )
      ..style = PaintingStyle.fill;

    final shadowPaint = Paint()
      ..color = Colors.white.withOpacity(0.5)
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 6);

    final width = size.width;
    final height = size.height;

    // วาดกลุ่มวงกลมซ้อนกันให้ดูเป็นเมฆกลุ่มก้อน
    final circles = [
      Offset(width * 0.3, height * 0.4),
      Offset(width * 0.5, height * 0.3),
      Offset(width * 0.7, height * 0.4),
      Offset(width * 0.2, height * 0.5),
      Offset(width * 0.4, height * 0.6),
      Offset(width * 0.6, height * 0.6),
      Offset(width * 0.8, height * 0.5),
    ];

    final radiuses = [
      width * 0.15,
      width * 0.18,
      width * 0.15,
      width * 0.12,
      width * 0.14,
      width * 0.14,
      width * 0.12,
    ];

    // วาดเงาก่อน
    for (int i = 0; i < circles.length; i++) {
      canvas.drawCircle(circles[i], radiuses[i] + 2, shadowPaint);
    }

    // วาดวงกลมหลัก
    for (int i = 0; i < circles.length; i++) {
      canvas.drawCircle(circles[i], radiuses[i], paint);
    }
  }

  // สไตล์เมฆแบบที่ 3 - มีขอบที่ชัดเจนกว่าและรูปร่างเป็นเมฆที่สมบูรณ์
  void _drawCloudStyle3(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.fill;

    final highlightPaint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.fill;

    final shadowPaint = Paint()
      ..color = Colors.white.withOpacity(0.7)
      ..style = PaintingStyle.fill
      ..maskFilter = const MaskFilter.blur(BlurStyle.outer, 4);

    final width = size.width;
    final height = size.height;

    // สร้าง path สำหรับเมฆที่มีขอบชัดเจนมากขึ้น
    final path = Path();

    path.moveTo(width * 0.2, height * 0.65);

    // ด้านล่าง
    path.lineTo(width * 0.8, height * 0.65);

    // ด้านขวา
    path.quadraticBezierTo(
        width * 0.9, height * 0.65, width * 0.9, height * 0.55);
    path.quadraticBezierTo(
        width * 0.9, height * 0.45, width * 0.8, height * 0.45);

    // ด้านบนขวา
    path.quadraticBezierTo(
        width * 0.75, height * 0.3, width * 0.65, height * 0.3);

    // ด้านบนกลาง
    path.quadraticBezierTo(
        width * 0.6, height * 0.2, width * 0.5, height * 0.25);
    path.quadraticBezierTo(
        width * 0.4, height * 0.2, width * 0.35, height * 0.3);

    // ด้านบนซ้าย
    path.quadraticBezierTo(
        width * 0.25, height * 0.3, width * 0.2, height * 0.45);

    // ด้านซ้าย
    path.quadraticBezierTo(
        width * 0.1, height * 0.45, width * 0.1, height * 0.55);
    path.quadraticBezierTo(
        width * 0.1, height * 0.65, width * 0.2, height * 0.65);

    // วาดเงา
    canvas.drawPath(path, shadowPaint);

    // วาดเมฆหลัก
    canvas.drawPath(path, paint);

    // เพิ่มไฮไลท์ด้านบนเมฆเพื่อสร้างมิติ
    final highlightPath = Path();
    highlightPath.moveTo(width * 0.3, height * 0.35);
    highlightPath.quadraticBezierTo(
        width * 0.4, height * 0.25, width * 0.5, height * 0.3);
    highlightPath.quadraticBezierTo(
        width * 0.6, height * 0.25, width * 0.7, height * 0.35);
    highlightPath.quadraticBezierTo(
        width * 0.6, height * 0.4, width * 0.5, height * 0.35);
    highlightPath.quadraticBezierTo(
        width * 0.4, height * 0.4, width * 0.3, height * 0.35);

    canvas.drawPath(highlightPath, highlightPaint);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) {
    if (identical(this, oldDelegate)) return false;

    return oldDelegate is! CloudPainter || oldDelegate.style != style;
  }
}
