import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class BackgroundMain extends StatelessWidget {
  final Widget child;
  const BackgroundMain({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    Size size = MediaQuery.of(context).size;
    return SizedBox(
      height: size.height,
      width: double.infinity,
      child: Stack(
        children: <Widget>[
          Container(color: global.theme.backgroundColor),
          child,
        ],
      ),
    );
  }
}
