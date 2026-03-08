import 'package:flutter/material.dart';

/// Horizontal ruler widget
class HorizontalRuler extends StatelessWidget {
  final double width;
  final double zoom;

  const HorizontalRuler({
    super.key,
    required this.width,
    this.zoom = 1.0,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 25,
      width: width * zoom,
      color: Colors.grey[300],
      child: CustomPaint(
        painter: _RulerPainter(
          width: width,
          zoom: zoom,
          isHorizontal: true,
        ),
      ),
    );
  }
}

/// Vertical ruler widget
class VerticalRuler extends StatelessWidget {
  final double height;
  final double zoom;

  const VerticalRuler({
    super.key,
    required this.height,
    this.zoom = 1.0,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 25,
      height: height * zoom,
      color: Colors.grey[300],
      child: CustomPaint(
        painter: _RulerPainter(
          width: height,
          zoom: zoom,
          isHorizontal: false,
        ),
      ),
    );
  }
}

class _RulerPainter extends CustomPainter {
  final double width;
  final double zoom;
  final bool isHorizontal;

  _RulerPainter({
    required this.width,
    required this.zoom,
    required this.isHorizontal,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.black87
      ..strokeWidth = 1;

    final textPainter = TextPainter(
      textDirection: TextDirection.ltr,
    );

    // Draw ruler marks
    const int interval = 10; // Every 10 pixels
    final int totalMarks = (width / interval).ceil();

    for (int i = 0; i <= totalMarks; i++) {
      final double position = i * interval * zoom;

      if (position > (isHorizontal ? size.width : size.height)) break;

      double lineHeight;
      if (i % 10 == 0) {
        // Major mark every 100 pixels
        lineHeight = 15;

        // Draw number
        textPainter.text = TextSpan(
          text: '${i * interval}',
          style: const TextStyle(
            color: Colors.black87,
            fontSize: 9,
          ),
        );
        textPainter.layout();

        if (isHorizontal) {
          canvas.drawLine(
            Offset(position, size.height),
            Offset(position, size.height - lineHeight),
            paint,
          );
          textPainter.paint(
            canvas,
            Offset(position + 2, 2),
          );
        } else {
          canvas.drawLine(
            Offset(size.width, position),
            Offset(size.width - lineHeight, position),
            paint,
          );
          canvas.save();
          canvas.translate(2, position + 2);
          canvas.rotate(-1.5708); // -90 degrees
          textPainter.paint(canvas, Offset.zero);
          canvas.restore();
        }
      } else if (i % 5 == 0) {
        // Medium mark every 50 pixels
        lineHeight = 10;
        if (isHorizontal) {
          canvas.drawLine(
            Offset(position, size.height),
            Offset(position, size.height - lineHeight),
            paint,
          );
        } else {
          canvas.drawLine(
            Offset(size.width, position),
            Offset(size.width - lineHeight, position),
            paint,
          );
        }
      } else {
        // Small mark every 10 pixels
        lineHeight = 5;
        if (isHorizontal) {
          canvas.drawLine(
            Offset(position, size.height),
            Offset(position, size.height - lineHeight),
            paint,
          );
        } else {
          canvas.drawLine(
            Offset(size.width, position),
            Offset(size.width - lineHeight, position),
            paint,
          );
        }
      }
    }
  }

  @override
  bool shouldRepaint(_RulerPainter oldDelegate) {
    return oldDelegate.width != width ||
        oldDelegate.zoom != zoom ||
        oldDelegate.isHorizontal != isHorizontal;
  }
}
