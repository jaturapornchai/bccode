import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class ImageTooltip extends StatefulWidget {
  final Widget child;
  final Image image;

  const ImageTooltip({key, required this.child, required this.image})
    : super(key: key);

  @override
  ImageTooltipState createState() => ImageTooltipState();
}

class ImageTooltipState extends State<ImageTooltip> {
  OverlayEntry? overlayEntry;
  bool isVisible = false;

  void removeTooltip() {
    if (isVisible) {
      isVisible = false;
      overlayEntry?.remove();
      overlayEntry = null;
    }
  }

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onHover: (PointerHoverEvent event) {
        if (!isVisible) {
          isVisible = true;
          _showImageTooltip(context, event.position);
        }
      },
      onExit: (PointerExitEvent event) {
        removeTooltip();
      },
      child: widget.child,
    );
  }

  void _showImageTooltip(BuildContext context, Offset position) {
    final screenSize = MediaQuery.of(context).size;
    const imageWidth = 200.0; // Set your desired image width
    const imageHeight = 200.0; // Set your desired image height
    const tooltipPadding = 8.0;

    double left = position.dx;
    double top = position.dy;

    if (position.dx + imageWidth + tooltipPadding > screenSize.width) {
      left = screenSize.width - imageWidth - tooltipPadding;
    }

    if (position.dy + imageHeight + tooltipPadding > screenSize.height) {
      top = screenSize.height - imageHeight - tooltipPadding;
    }

    overlayEntry = OverlayEntry(
      builder: (BuildContext context) {
        return Positioned(
          left: left + tooltipPadding,
          top: top + tooltipPadding,
          child: Material(
            color: Colors.transparent,
            child: Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(5),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.1),
                    spreadRadius: 1,
                    blurRadius: 5,
                    offset: const Offset(0, 3),
                  ),
                ],
              ),
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: SizedBox(
                  width: imageWidth,
                  height: imageHeight,
                  child: widget.image,
                ),
              ),
            ),
          ),
        );
      },
    );
    Overlay.of(context).insert(overlayEntry!);
  }
}
