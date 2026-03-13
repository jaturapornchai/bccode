import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class SingInButton extends StatelessWidget {
  final String labelText;
  final Function() press;
  final AssetImage img;
  const SingInButton({
    super.key,
    required this.labelText,
    required this.press,
    required this.img,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: press,
      child: Container(
        width: 400,
        height: 60,
        padding: const EdgeInsets.fromLTRB(0, 10, 0, 10),
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          borderRadius: const BorderRadius.only(
            topLeft: Radius.circular(40),
            topRight: Radius.circular(40),
            bottomLeft: Radius.circular(40),
            bottomRight: Radius.circular(40),
          ),

          /// border color
          border: Border.all(color: const Color(0xFFF56045), width: 5),

          /// drop shadow
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.5),
              spreadRadius: 0,
              blurRadius: 5,
              offset: const Offset(0, 3), // changes position of shadow
            ),
          ],
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Image(image: img, height: 50.0),
            Padding(
              padding: const EdgeInsets.only(left: 10),
              child: Text(
                labelText,
                style: TextStyle(fontSize: 25, color: global.theme.textColor),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
