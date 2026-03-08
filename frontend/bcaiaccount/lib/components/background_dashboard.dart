import 'package:flutter/material.dart';
import 'package:smlaicloud/utils/util.dart';

class BackgroundDashboard extends StatelessWidget {
  final Widget child;

  const BackgroundDashboard({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    Size size = MediaQuery.of(context).size;
    double heightContainer = 0.0;
    double heightClipPath = 0.0;

    if (Util.isLandscape(context)) {
      heightContainer = 0.4;
      heightClipPath = 0.6;
    } else {
      heightContainer = 0.3;
      heightClipPath = 0.4;
    }

    return SizedBox(
      height: size.height,
      width: double.infinity,
      child: Container(
        color: Colors.white,
        child: Stack(
          children: <Widget>[
            Container(
              height: size.height * heightContainer,
              decoration: const BoxDecoration(
                borderRadius: BorderRadius.only(
                  bottomLeft: Radius.circular(120),
                  bottomRight: Radius.circular(120),
                ),
                color: Color(0xFFF88975),
              ),
            ),
            ClipPath(
              clipper: MyClipper(),
              child: Container(
                height: size.height * heightClipPath,
                decoration: const BoxDecoration(color: Color(0xFFF56045)),
              ),
            ),
            child,
          ],
        ),
      ),
    );
  }
}

class MyClipper extends CustomClipper<Path> {
  @override
  Path getClip(Size size) {
    var path = Path();
    path.lineTo(0, size.height - 280);
    path.quadraticBezierTo(
      size.width / 5,
      size.height,
      size.width,
      size.height - 280,
    );
    path.lineTo(size.width, 0);
    path.close();
    return path;
  }

  @override
  bool shouldReclip(CustomClipper<Path> oldClipper) {
    return false;
  }
}
